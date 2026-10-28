package Lyger

import (
	"fmt"
	"sort"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// instruction is an ordered outer-program effect. Calls retain their source
// declaration and synchronous expansion; no instruction executes during compile.
type instruction struct {
	Application        *clauseGram
	Constructor        *constructionNode
	Key                *instructionValue
	Kind               instructionKind
	Source             SourceSpan
	SourceToken        *token
	Statement          *lStatement
	Order              int
	Definition         *lStatement
	CallSite           *lStatement
	Target             *instructionValue
	Value              *instructionValue
	Condition          *instructionCondition
	Then               []*instruction
	Else               []*instruction
	ElseStatement      *lStatement
	Actions            []*instruction
	Block              *scopedInstructions
	Synchronous        bool
	Callable           *record
	Captures           []*instructionValue
	CaptureCalls       []*record
	Function           *sourceFunction
	Arguments, Results []*instructionValue
}

// A symbol reference remains a reference even when its record contains a
// compile-time value. Runtime distinguishes conditional assignments and member
// reads from literal values; Output's bytes are never trimmed or normalized.
type instructionValue struct {
	Atom        *DefinitionArgument
	RuntimeType runtimeTypeRef
	Elements    []*instructionValue
	Key         *instructionValue
	Kind        operandKind
	Type        valueType
	ID          string
	Record      *record
	Literal     *primitive
	Owner       *instructionValue
	Member      nativeMember
	Payload     *instructionValue
	Token       *token
	Runtime     bool
	Argument    *instructionValue
}

type instructionCondition struct {
	Operator string
	Left     *instructionValue
	Right    *instructionValue
	Negated  bool
	Terms    []*instructionCondition
}

type instructionBuilder struct {
	semanticResolver
	ordered     map[*lStatement]scopedStatement
	activeCalls map[*record]bool
	captures    map[*record]*instruction
}

type instructionContext struct {
	function  *sourceFunction
	statement *lStatement
	order     int
	arguments map[*token]*instructionValue
	owner     *instructionValue
	callSite  *lStatement
}

func instructionType(record *record) (valueType, error) {
	if record == nil {
		return valueInvalid, lerrors.Internal.Errorf("instruction requires a resolved record")
	}
	value, err := record.Value.value()
	if err != nil {
		return valueInvalid, fmt.Errorf("read instruction value %s: %w", record.ID, err)
	}
	return value.Type, nil
}

func (b *instructionBuilder) symbolValue(tok *token) (*instructionValue, error) {
	found, err := b.reference(tok)
	if err != nil {
		return nil, err
	}
	if found == nil {
		return nil, lerrors.UnresolvedReference.Errorf("unresolved instruction reference %q", tok.rawBytes)
	}
	var original *record
	if binding, ok := b.scopes.Pronouns[tok]; ok {
		original = b.semantics.Subjects[binding.TargetToken]
	} else if tok.isDeclare() {
		original = b.semantics.Subjects[tok]
	} else if tok.binding != nil {
		original = tok.binding.Record
	}
	if original != nil && original.runtimeFunctionAlias {
		found = original
	}
	kind, err := instructionType(found)
	if err != nil {
		return nil, err
	}
	runtime := found.Value.Runtime != nil
	if declaration := b.semantics.Declarations[found]; declaration != nil && declaration.sGram != nil && declaration.sGram.predicate != nil {
		runtime = runtime || len(declaration.sGram.predicate.conditionals) > 0
	}
	typ, err := recordRuntimeType(found)
	if err != nil {
		return nil, err
	}
	return &instructionValue{Kind: operandSymbol, Type: kind, RuntimeType: typ, ID: found.ID, Record: found, Token: tok, Runtime: runtime}, nil
}

func (b *instructionBuilder) argumentValue(tok *token, ctx instructionContext) (*instructionValue, error) {
	if actual := ctx.arguments[tok]; actual != nil {
		return actual, nil
	}
	if binding, ok := b.scopes.Pronouns[tok]; ok && binding.Kind == scopeActionArgument {
		if actual := ctx.arguments[binding.TargetToken]; actual != nil {
			return actual, nil
		}
		return nil, lerrors.UnresolvedReference.Errorf("unbound action argument %s", binding.SubjectID)
	}
	return nil, nil
}

func (b *instructionBuilder) tokenValue(tok *token, ctx instructionContext) (*instructionValue, error) {
	if tok == nil {
		return nil, lerrors.Internal.Errorf("instruction operand has no noun token")
	}
	if actual, err := b.argumentValue(tok, ctx); actual != nil || err != nil {
		return actual, err
	}
	if _, ok := b.scopes.Pronouns[tok]; ok {
		return b.symbolValue(tok)
	}
	if tok.binding != nil {
		return b.symbolValue(tok)
	}
	if tok.isDeclare() {
		return b.symbolValue(tok)
	}
	if tok.mem != nil && tok.mem.Type == storagePrimitive && tok.mem.Primitive != nil && tok.mem.Primitive.Type != valueUnknown && tok.mem.Primitive.Type != valueNone {
		literal := *tok.mem.Primitive
		literal.Bytes = append([]byte(nil), literal.Bytes...)
		value, err := tok.mem.value()
		if err != nil {
			return nil, err
		}
		kind := value.Type
		return &instructionValue{Kind: operandLiteral, Type: kind, RuntimeType: runtimeTypeRef{Kind: kind}, Literal: &literal, Token: tok}, nil
	}
	if lex := tok.lexeme(); lex != nil && strings.HasPrefix(lex.Type, "functional.") {
		if ctx.owner == nil {
			return nil, lerrors.InvalidInput.Errorf("member %s has no instruction owner", lex.Root)
		}
		owner := ctx.owner
		if owner.Type != valueApp && owner.Record != nil && owner.Record.Value.Runtime != nil {
			if rec := owner.Record.Value.Runtime.Owner; rec != nil {
				kind, err := instructionType(rec)
				if err != nil {
					return nil, err
				}
				owner = &instructionValue{Kind: operandSymbol, Type: kind, ID: rec.ID, Record: rec}
			}
		}
		return instructionMember(owner, tok)
	}
	return b.symbolValue(tok)
}

func instructionMember(owner *instructionValue, member *token) (*instructionValue, error) {
	spec, ok := nativeMemberFor(owner.Type, member.lexeme())
	if !ok {
		return nil, lerrors.TypeMismatch.Errorf("invalid runtime member %q on %s", member.rawBytes, owner.Type)
	}
	kind, _ := spec.ValueType()
	return &instructionValue{Kind: operandMember, Type: kind, Owner: owner, Member: spec, Token: member, Runtime: true}, nil
}

func (b *instructionBuilder) nominalValue(noun *token, tokens []*token, ctx instructionContext) (*instructionValue, error) {
	if ctx.function != nil && ctx.function.Native {
		argument, err := argumentFromToken(noun)
		if err != nil {
			return nil, err
		}
		argument.source = ctx.statement.source.Span
		return &instructionValue{Kind: operandLiteral, Atom: &argument, Token: noun}, nil
	}
	if noun != nil && (noun.binding != nil || noun.isDeclare()) {
		return b.tokenValue(noun, ctx)
	}
	for _, current := range tokens {
		if current == noun {
			break
		}
		if _, possessive := possessiveOwner(string(current.rawBytes)); possessive {
			owner, err := b.tokenValue(current, ctx)
			if err != nil {
				return nil, err
			}
			return instructionMember(owner, noun)
		}
	}
	return b.tokenValue(noun, ctx)
}

func (b *instructionBuilder) objectValue(object *objectGram, ctx instructionContext) (*instructionValue, error) {
	if object == nil {
		return nil, lerrors.Internal.Errorf("instruction operand has no object grammar")
	}
	if actual, err := b.argumentValue(object.noun, ctx); actual != nil || err != nil {
		return actual, err
	}
	if role, native := object.noun.nativeRole(); native && role == roleError && object.phrase != nil && runtimeReadPhrase(object) == nil {
		payload, err := b.phraseValue(object.phrase, ctx)
		if err != nil {
			return nil, err
		}
		return &instructionValue{Kind: operandError, Type: valueError, RuntimeType: runtimeTypeRef{Kind: valueError}, Payload: payload, Token: object.noun, Runtime: true}, nil
	}
	if _, ok, err := runtimeObjectKind(object); err != nil {
		return nil, err
	} else if ok {
		return b.constructedValue(object, ctx)
	}
	return b.nominalValue(object.noun, object.tokens, ctx)
}

func (b *instructionBuilder) phraseValue(phrase *phraseGram, ctx instructionContext) (*instructionValue, error) {
	if phrase == nil {
		return nil, lerrors.Internal.Errorf("instruction has no value phrase")
	}
	if phrase.obj != nil {
		return b.objectValue(phrase.obj, ctx)
	}
	if phrase.subject != nil {
		return b.nominalValue(phrase.subject.noun, phrase.subject.tokens, ctx)
	}
	return nil, lerrors.Internal.Errorf("instruction phrase %s has no operand", phrase.sType)
}

func orderedInstructionClauses(clauses []*clauseGram, tokens []*token) []*clauseGram {
	var ordered []*clauseGram
	seen := make(map[*clauseGram]bool)
	for _, clause := range clauses {
		if clause != nil && !seen[clause] {
			seen[clause] = true
			ordered = append(ordered, clause)
		}
	}
	positions := make(map[*token]int, len(tokens))
	for index, token := range tokens {
		if _, exists := positions[token]; !exists {
			positions[token] = index
		}
	}
	position := func(clause *clauseGram) int {
		if index, exists := positions[clause.verb]; exists {
			return index
		}
		return len(tokens)
	}
	sort.SliceStable(ordered, func(i, j int) bool {
		return position(ordered[i]) < position(ordered[j])
	})
	return ordered
}

func (ctx instructionContext) newInstruction(kind instructionKind) *instruction {
	return &instruction{Kind: kind, Source: ctx.statement.source.Span, Statement: ctx.statement, Order: ctx.order, CallSite: ctx.callSite}
}

// newAssignment resolves the declaration target for either grammatical form.
func (b *instructionBuilder) newAssignment(target *token, ctx instructionContext) (*instruction, error) {
	if target == nil || !target.isDeclare() {
		return nil, lerrors.InvalidInput.Errorf("runtime assignment requires a named subject")
	}
	current := ctx.newInstruction(instructionAssign)
	current.SourceToken = target
	var err error
	current.Target, err = b.tokenValue(target, ctx)
	if err != nil {
		return nil, err
	}
	current.Target.Runtime = true
	return current, nil
}

func (b *instructionBuilder) clauses(clauses []*clauseGram, ctx instructionContext) ([]*instruction, error) {
	var instructions []*instruction
	for _, clause := range orderedInstructionClauses(clauses, ctx.statement.tokens) {
		current, err := b.clause(clause, ctx)
		if err != nil {
			return nil, err
		}
		instructions = append(instructions, current...)
	}
	return instructions, nil
}

func (b *instructionBuilder) clause(clause *clauseGram, ctx instructionContext) ([]*instruction, error) {
	site := registrationSite{subject: clause.subject, predicate: clause.predicate, clause: clause}
	if ctx.function != nil && ctx.function.Native {
		var current *instruction
		if clause.subject != nil && clause.subject.noun != nil && clause.subject.noun.isDeclare() {
			rec := b.semantics.Subjects[clause.subject.noun]
			if rec == nil {
				return nil, lerrors.UnresolvedReference.Errorf("source declaration has no canonical binding")
			}
			if rec.staticConstructor != nil {
				current = ctx.newInstruction(instructionAssign)
				current.Constructor = rec.staticConstructor
			}
		} else if !site.isAlias() && !site.isDescriptor() {
			current = ctx.newInstruction(instructionCall)
			current.Application, current.Synchronous = clause, true
			if clause.verb == nil || clause.verb.binding == nil {
				return nil, lerrors.UnresolvedReference.Errorf("source operation has no definition")
			}
			if ctx.function.bindsReturn(clause.verb) {
				current.Kind = instructionReturn
			}
		}
		following, err := b.clauses(clause.clauses, ctx)
		if current != nil {
			return append([]*instruction{current}, following...), err
		}
		return following, err
	}
	if clause.predicate != nil && clause.subject != nil {
		return b.passiveDeclaration(clause.subject, clause.predicate, ctx)
	}
	if event, err := b.declarationInstruction(site, ctx); err != nil || event != nil {
		if err != nil {
			return nil, err
		}
		following, err := b.clauses(clause.clauses, ctx)
		return append([]*instruction{event}, following...), err
	}
	if site.isAlias() {
		aliases, err := b.functionAlias(site, ctx)
		if err != nil {
			return nil, err
		}
		following, err := b.clauses(clause.clauses, ctx)
		return append(aliases, following...), err
	}
	if site.isDescriptor() {
		return b.clauses(clause.clauses, ctx)
	}
	if clause.verb == nil {
		return nil, lerrors.Internal.Errorf("instruction clause %s has no verb", clause.sType)
	}
	if clause.verb.isFunctionReturn() {
		return b.functionOperation(clause, ctx, instructionReturn)
	}
	if clause.verb.binding != nil && clause.verb.binding.Record.isFunction() {
		return b.functionOperation(clause, ctx, instructionInvoke)
	}
	lex := clause.verb.lexeme()
	if lex == nil {
		return nil, lerrors.Internal.Errorf("instruction verb %q has no lexical value", clause.verb.rawBytes)
	}
	for _, declaration := range b.ordered[ctx.statement].declarationsBefore(clause.verb) {
		rec, err := b.reference(declaration.Token)
		if err != nil {
			return nil, err
		}
		if rec != nil {
			kind, err := instructionType(rec)
			if err != nil {
				return nil, err
			}
			if kind == valueApp {
				ctx.owner = &instructionValue{Kind: operandSymbol, Type: kind, ID: rec.ID, Record: rec, Token: declaration.Token}
			}
		}
	}
	role, _ := clause.verb.nativeRole()
	kind, native := role.InstructionKind()
	if !native {
		if clause.verb.binding != nil {
			return b.customCall(clause, ctx)
		}
		return nil, lerrors.Unsupported.Errorf("unsupported outer instruction action %s", lex.Type)
	}
	current := ctx.newInstruction(kind)
	current.SourceToken = clause.verb
	var err error
	switch kind {
	case instructionUpdate:
		if clause.phrase == nil || clause.phrase.subject == nil {
			return nil, lerrors.InvalidInput.Errorf("update requires a destination")
		}
		current.Target, err = b.nominalValue(clause.phrase.subject.noun, clause.phrase.subject.tokens, ctx)
		if err == nil {
			current.Value, err = b.objectValue(clause.obj, ctx)
		}
		if err == nil && clause.phrase.phrase != nil {
			current.Key, err = b.phraseValue(clause.phrase.phrase, ctx)
		}
	case instructionAssign:
		if clause.subject == nil {
			return nil, lerrors.InvalidInput.Errorf("runtime assignment requires a named subject")
		}
		current, err = b.newAssignment(clause.subject.noun, ctx)
		if err == nil {
			current.Value, err = b.phraseValue(clause.subject.phrase, ctx)
			if err == nil && current.Value.RuntimeType.Kind == current.Target.Type {
				current.Target.Record.runtimeType = &current.Value.RuntimeType
				current.Target.RuntimeType = current.Value.RuntimeType
			}
		}
	case instructionBuild, instructionRun, instructionAwait, instructionStop, instructionShow:
		// Native coordinated unary clauses perform one operation per operand
		// in source order. Source actions instead expand their entire body for
		// each operand in customCall.
		var operations []*instruction
		for _, object := range clause.orderedObjects() {
			operation := ctx.newInstruction(kind)
			operation.SourceToken = clause.verb
			value, err := b.objectValue(object, ctx)
			if err != nil {
				return nil, err
			}
			if kind == instructionShow {
				operation.Value = value
			} else {
				operation.Target = value
				if value.Type != valueApp {
					return nil, lerrors.TypeMismatch.Errorf("%s requires an App operand, got %s", lex.Root, value.Type)
				}
			}
			// Source execution waits for each operation's acknowledgment. Run's
			// child lifetime continues independently after that acknowledgment.
			operation.Synchronous = true
			operations = append(operations, operation)
		}
		if len(operations) == 0 {
			return nil, lerrors.InvalidInput.Errorf("%s requires an operand", lex.Root)
		}
		following, err := b.clauses(clause.clauses, ctx)
		return append(operations, following...), err
	case instructionExit:
		if clause.obj != nil {
			current.Value, err = b.objectValue(clause.obj, ctx)
		} else if clause.phrase != nil {
			current.Value, err = b.phraseValue(clause.phrase, ctx)
		}
	}
	if err != nil {
		return nil, err
	}
	following, err := b.clauses(clause.clauses, ctx)
	if err != nil {
		return nil, err
	}
	return append([]*instruction{current}, following...), nil
}

// directiveClauses follows grammar ownership to the source-defined action
// body. The grammar retains argument tokens and source spans needed for scoped
// references and attribution in addition to the lexical action metadata.
func directiveClauses(statement *lStatement) []*clauseGram {
	var result []*clauseGram
	seen := make(map[*phraseGram]bool)
	var phrase func(*phraseGram)
	var clause func(*clauseGram)
	object := func(current *objectGram) {
		if current != nil {
			phrase(current.phrase)
			clause(current.clause)
		}
	}
	clause = func(current *clauseGram) {
		if current == nil {
			return
		}
		if current.subject != nil {
			phrase(current.subject.phrase)
		}
		phrase(current.phrase)
		object(current.obj)
		for _, child := range current.clauses {
			clause(child)
		}
	}
	phrase = func(current *phraseGram) {
		if current == nil || seen[current] {
			return
		}
		seen[current] = true
		if current.gType == "phrase.directive" {
			result = append(result, current.clauses...)
			return
		}
		if current.subject != nil {
			phrase(current.subject.phrase)
		}
		phrase(current.phrase)
		object(current.obj)
		for _, child := range current.clauses {
			clause(child)
		}
	}
	if statement.sGram != nil {
		if statement.sGram.subject != nil {
			phrase(statement.sGram.subject.phrase)
		}
		if predicate := statement.sGram.predicate; predicate != nil {
			for _, child := range predicate.clauses {
				clause(child)
			}
		}
	}
	return result
}

func (b *instructionBuilder) customCall(clause *clauseGram, ctx instructionContext) ([]*instruction, error) {
	callee, err := canonicalRecord(clause.verb.binding.Record, b.semantics.Aliases)
	if err != nil {
		return nil, err
	}
	definition := b.semantics.Declarations[callee]
	if definition == nil || definition.sGram == nil {
		return nil, lerrors.Internal.Errorf("action %s has no resolved source declaration", callee.ID)
	}
	if b.activeCalls[callee] {
		return nil, lerrors.Cycle.Errorf("recursive source action %s", callee.ID)
	}
	body := directiveClauses(definition)
	if len(body) == 0 {
		return nil, lerrors.InvalidInput.Errorf("action %s has no directive clauses", callee.ID)
	}
	objects := clause.orderedObjects()
	parameters := b.ordered[definition].Arguments
	groups := [][]*objectGram{objects}
	if len(parameters) == 1 && len(objects) > 1 {
		groups = nil
		for _, object := range objects {
			groups = append(groups, []*objectGram{object})
		}
	}
	b.activeCalls[callee] = true
	defer delete(b.activeCalls, callee)
	var calls []*instruction
	for _, actuals := range groups {
		if len(actuals) != len(parameters) {
			return nil, lerrors.InvalidInput.Errorf("action %s expects %d arguments, got %d", callee.ID, len(parameters), len(objects))
		}
		arguments := make(map[*token]*instructionValue)
		var argumentValues []*instructionValue
		for index, object := range actuals {
			actual, err := b.objectValue(object, ctx)
			if err != nil {
				return nil, err
			}
			wanted, err := nominalType(parameters[index].Token)
			if err != nil {
				return nil, err
			}
			if actual.Type != valueInvalid && !strings.EqualFold(actual.Type.String(), wanted) {
				return nil, lerrors.TypeMismatch.Errorf("action %s argument %d requires %s, got %s", callee.ID, index+1, wanted, actual.Type)
			}
			bound := *actual
			bound.Argument = actual
			arguments[parameters[index].Token] = &bound
			argumentValues = append(argumentValues, &bound)
		}
		call := ctx.newInstruction(instructionCall)
		call.SourceToken = clause.verb
		call.Definition, call.Synchronous = definition, true
		call.Callable = clause.verb.binding.Record
		call.Arguments = argumentValues
		call.Target = &instructionValue{Kind: operandSymbol, Type: valueLexeme, ID: callee.ID, Record: callee, Token: clause.verb}
		expanded := ctx
		expanded.statement, expanded.arguments, expanded.callSite = definition, arguments, ctx.invocationSite()
		call.Actions, err = b.clauses(body, expanded)
		if err != nil {
			return nil, err
		}
		calls = append(calls, call)
	}
	following, err := b.clauses(clause.clauses, ctx)
	if err != nil {
		return nil, err
	}
	return append(calls, following...), nil
}

func (b *instructionBuilder) condition(condition *conditionGram, ctx instructionContext) (*instructionCondition, error) {
	if condition == nil || len(condition.values) == 0 {
		return nil, lerrors.Internal.Errorf("conditional has no value grammar")
	}
	result := &instructionCondition{Operator: condition.cType}
	left := condition.values[0]
	var err error
	result.Left, err = b.nominalValue(left.noun, left.tokens, ctx)
	if err != nil {
		return nil, err
	}
	if len(condition.values) > 1 {
		right := condition.values[1]
		rightContext := ctx
		if condition.cType == "has" {
			rightContext.owner = result.Left
		}
		result.Right, err = b.nominalValue(right.noun, right.tokens, rightContext)
		if err != nil {
			return nil, err
		}
	}
	if len(condition.values) > 2 {
		return nil, lerrors.Unsupported.Errorf("condition %s has unsupported operand count", condition.cType)
	}
	result.Negated = condition.article.hasRole("article", "none") || (condition.adverb != nil && strings.Contains(condition.adverb.tType, ".nullification"))
	return result, nil
}

func (b *instructionBuilder) conditional(conditional *conditionalGram, ctx instructionContext) (*instruction, error) {
	current := ctx.newInstruction(instructionIf)
	if len(conditional.tokens) > 0 {
		current.SourceToken = conditional.tokens[0]
	}
	if len(conditional.conditions) == 0 {
		return nil, lerrors.Internal.Errorf("conditional has no parsed condition")
	}
	if len(conditional.conditions) > 1 {
		return nil, lerrors.Unsupported.Errorf("multiple condition joins need an explicit instruction operator")
	}
	var err error
	current.Condition, err = b.condition(conditional.conditions[0], ctx)
	if err != nil {
		return nil, err
	}
	branchContext := ctx
	if current.Condition.Operator == "has" {
		branchContext.owner = current.Condition.Left
	}
	current.Then, err = b.clauses(conditional.orderedClauses(), branchContext)
	return current, err
}

func (b *instructionBuilder) statement(statement *lStatement, ctx instructionContext) ([]*instruction, bool, error) {
	predicate := statement.sGram.predicate
	if len(predicate.conditionals) > 0 {
		var branch *instruction
		for _, conditional := range predicate.conditionals {
			term, err := decodeGrammarTerm(conditional.gType)
			if err != nil {
				return nil, false, err
			}
			switch term.Name {
			case "conditional.other":
				instructions, err := b.clauses(conditional.orderedClauses(), ctx)
				return instructions, true, err
			case "conditional.if":
				if branch != nil {
					return nil, false, lerrors.Syntax.Errorf("unexpected second primary conditional")
				}
				var err error
				branch, err = b.conditional(conditional, ctx)
				if err != nil {
					return nil, false, err
				}
			case "conditional.else":
				if branch == nil || branch.ElseStatement != nil {
					return nil, false, lerrors.Syntax.Errorf("else has no unmatched conditional")
				}
				var err error
				branch.Else, err = b.clauses(conditional.orderedClauses(), ctx)
				if err != nil {
					return nil, false, err
				}
				branch.ElseStatement = statement
			default:
				return nil, false, lerrors.Unsupported.Errorf("unsupported instruction conditional %s", conditional.gType)
			}
		}
		return []*instruction{branch}, false, nil
	}
	if subject := statement.sGram.subject; subject != nil && subject.noun != nil && subject.noun.isDeclare() {
		instructions, err := b.passiveDeclaration(subject, predicate, ctx)
		if err != nil {
			return nil, false, err
		}
		following, err := b.clauses(statement.sGram.clauses, ctx)
		return append(instructions, following...), false, err
	}
	if strings.HasPrefix(statement.sType, "verbs.declarative.") || strings.HasPrefix(statement.sType, "verbs.definitive.") || strings.HasPrefix(statement.sType, "verbs.descriptive.") || strings.HasPrefix(statement.sType, "verbs.associative.") {
		var instructions []*instruction
		for _, clause := range predicate.clauses {
			if clause.subject != nil && clause.subject.noun != nil && clause.subject.noun.isDeclare() {
				rec := b.semantics.Subjects[clause.subject.noun]
				if rec == nil || (b.captures[rec] == nil && (rec.Value.Runtime == nil && rec.staticConstructor == nil || b.semantics.Aliases[rec] != nil)) {
					following, err := b.clauses(clause.clauses, ctx)
					if err != nil {
						return nil, false, err
					}
					instructions = append(instructions, following...)
					continue
				}
			}
			following, err := b.clause(clause, ctx)
			if err != nil {
				return nil, false, err
			}
			instructions = append(instructions, following...)
		}
		return instructions, false, nil
	}
	// The integrating pass owns declarations and application-body effects.
	// Other subjective effects must not silently disappear from outer IR.
	if subject := statement.sGram.subject; subject != nil {
		if statement.passiveLifecycle() {
			var instructions []*instruction
			target, err := b.nominalValue(subject.noun, subject.tokens, ctx)
			if err != nil {
				return nil, false, err
			}
			for _, clause := range predicate.clauses {
				kind, passive := clause.passiveLifecycleKind()
				if !passive {
					following, err := b.clause(clause, ctx)
					if err != nil {
						return nil, false, err
					}
					instructions = append(instructions, following...)
					continue
				}
				if target.Type != valueApp {
					return nil, false, lerrors.TypeMismatch.Errorf("%s requires an App operand, got %s", clause.verb.lexeme().Root, target.Type)
				}
				current := ctx.newInstruction(kind)
				current.SourceToken, current.Target = clause.verb, target
				current.Synchronous = true
				instructions = append(instructions, current)
			}
			following, err := b.clauses(statement.sGram.clauses, ctx)
			return append(instructions, following...), false, err
		}
		if len(predicate.clauses) > 0 && predicate.clauses[0].verb != nil && predicate.clauses[0].verb.binding != nil && predicate.clauses[0].verb.binding.Record.isFunction() {
			instructions, err := b.clauses(predicate.clauses, ctx)
			return instructions, false, err
		}
		if subject.noun != nil && subject.noun.isDeclare() {
			return nil, false, nil
		}
		// A bound subject's canonical value wins over its spelling: an App
		// named Service remains an App, and a Service named App has no body.
		if subject.noun != nil && (subject.noun.binding != nil || subject.noun.isPronoun()) {
			value, err := b.nominalValue(subject.noun, subject.tokens, ctx)
			if err != nil {
				return nil, false, err
			}
			if value.Type == valueApp {
				return nil, false, nil
			}
			if value.Type == valueService {
				return nil, false, lerrors.Unsupported.Errorf("standalone Service %s body is unsupported; compose its behavior within an App", value.ID)
			}
			return nil, false, lerrors.Unsupported.Errorf("unsupported subjective instruction for %s", value.Type)
		}
		nominal, err := nominalType(subject.noun)
		if err != nil {
			return nil, false, err
		}
		if nominal == "app" {
			return nil, false, nil
		}
		if nominal == "service" {
			// Definite type references retain their original paragraph subject;
			// they must not search for an earlier or same-named Service.
			rec, err := b.reference(b.ordered[statement].IncomingSubjectToken)
			if err != nil {
				return nil, false, err
			}
			if rec != nil {
				kind, err := instructionType(rec)
				if err != nil {
					return nil, false, err
				}
				if kind == valueService {
					return nil, false, lerrors.Unsupported.Errorf("standalone Service %s body is unsupported; compose its behavior within an App", rec.ID)
				}
			}
			return nil, false, lerrors.UnresolvedReference.Errorf("definite Service reference has no Service subject in its paragraph")
		}
		value, err := b.nominalValue(subject.noun, subject.tokens, ctx)
		if err != nil {
			return nil, false, err
		}
		if value.Type == valueApp {
			return nil, false, nil
		}
		return nil, false, lerrors.Unsupported.Errorf("unsupported subjective instruction for %s", value.Type)
	}
	instructions, err := b.clauses(predicate.clauses, ctx)
	return instructions, false, err
}

// A passive lifecycle operation targets its grammatical subject in the
// orchestrator. It is not a native action inside that App's child program.
func (statement *lStatement) passiveLifecycle() bool {
	if statement == nil || statement.sGram == nil || statement.sGram.subject == nil || statement.sGram.predicate == nil || !strings.HasPrefix(statement.sType, "verbs.passive.") {
		return false
	}
	clauses := statement.sGram.predicate.clauses
	if len(clauses) == 0 {
		return false
	}
	// The first clause establishes the outer passive operation. Following
	// coordinated clauses may be imperative operations or source actions.
	_, passive := clauses[0].passiveLifecycleKind()
	return passive
}

func (clause *clauseGram) passiveLifecycleKind() (instructionKind, bool) {
	if clause == nil || clause.verb == nil || !strings.HasSuffix(clause.gType, ".passive") {
		return instructionInvalid, false
	}
	role, _ := clause.verb.nativeRole()
	kind, _ := role.InstructionKind()
	switch kind {
	case instructionBuild, instructionRun, instructionAwait, instructionStop:
		return kind, true
	}
	return instructionInvalid, false
}

// buildInstructions consumes successful grammar and source scope, preserving the
// scope plan's original ordering even when the caller's retry queue is reordered.
func buildInstructions(statements []*lStatement, scopes *scopePlan, semantics *semanticProgram) ([]*instruction, error) {
	if scopes == nil {
		return nil, lerrors.Internal.Errorf("instruction lowering requires source scope")
	}
	b := &instructionBuilder{semanticResolver: semanticResolver{scopes: scopes, semantics: semantics}, ordered: make(map[*lStatement]scopedStatement), activeCalls: make(map[*record]bool)}
	selected := make(map[*lStatement]bool)
	for _, statement := range statements {
		if statement == nil || selected[statement] {
			return nil, lerrors.Internal.Errorf("instruction source contains a nil or repeated statement")
		}
		selected[statement] = true
	}
	for _, ordered := range scopes.Ordered {
		if !selected[ordered.Statement] {
			return nil, lerrors.Internal.Errorf("instruction source does not contain every scoped statement")
		}
		b.ordered[ordered.Statement] = ordered
	}
	if len(b.ordered) != len(selected) {
		return nil, lerrors.Internal.Errorf("instruction source contains a statement without scope")
	}
	if err := b.prepareCaptures(); err != nil {
		return nil, err
	}
	for _, ordered := range scopes.Ordered {
		if function := ordered.Scope.function; function != nil {
			function.Body = nil
		}
	}
	var result []*instruction
	var previous *instruction
	var previousStatement *lStatement
	for _, ordered := range scopes.Ordered {
		statement := ordered.Statement
		if statement.source == nil || statement.sGram == nil || statement.sGram.predicate == nil {
			return nil, ordered.Source.wrapError(lerrors.Internal.Errorf("instruction lowering requires successful statement grammar"))
		}
		// Headers execute in the enclosing activation; child scopes inherit their Function.
		var function *sourceFunction
		for scope := ordered.Scope; scope != nil; scope = scope.parent {
			if scope.function != nil && statement != scope.function.Header {
				function = scope.function
				break
			}
		}
		ctx := instructionContext{statement: statement, order: ordered.Order, function: function}
		header := ordered.Scope.function != nil && statement == ordered.Scope.function.Header
		if ordered.IncomingSubjectToken != nil {
			var err error
			ctx.owner, err = b.symbolValue(ordered.IncomingSubjectToken)
			if err != nil {
				return nil, ordered.Source.wrapError(err)
			}
		}
		instructions, otherwise, err := b.statement(statement, ctx)
		if err != nil {
			return nil, ordered.Source.wrapError(err)
		}
		if unit := statement.source.Unit; ctx.function == nil && unit != nil && unit.Kind != SourceProgram {
			forbidden := len(instructions) > 0 || otherwise
			for _, rec := range statement.records {
				kind, err := instructionType(rec)
				if err != nil {
					return nil, err
				}
				forbidden = forbidden || kind == valueApp
			}
			if forbidden {
				return nil, ordered.Source.wrapError(lerrors.Unsupported.Errorf("%s %s cannot execute initialization effects or create App instances", unit.Kind, unit.Origin.Name))
			}
			continue
		}
		if otherwise {
			if previous == nil || previous.Kind != instructionIf || previousStatement == nil || previousStatement.source.Span.Path != statement.source.Span.Path || previousStatement.source.Span.Paragraph != statement.source.Span.Paragraph || previous.ElseStatement != nil {
				return nil, ordered.Source.wrapError(lerrors.Syntax.Errorf("otherwise has no preceding unmatched conditional in this source paragraph"))
			}
			previous.Else, previous.ElseStatement = instructions, statement
			previous, previousStatement = nil, nil
			continue
		}
		if function != nil {
			function.Body = append(function.Body, instructions...)
		} else {
			result = append(result, instructions...)
		}
		previous, previousStatement = nil, statement
		if header {
			previousStatement = nil
		} else if len(instructions) > 0 {
			previous = instructions[len(instructions)-1]
		}
	}
	result, err := b.activateCaptures(result)
	if err != nil {
		return nil, err
	}
	for _, function := range semantics.Functions {
		function.Captures, function.CaptureCalls = b.functionCaptures(function)
	}
	return b.scopeInstructions(result)
}

func (ctx instructionContext) invocationSite() *lStatement {
	if ctx.callSite != nil {
		return ctx.callSite
	}
	return ctx.statement
}

func (b *instructionBuilder) passiveDeclaration(subject *subjectGram, predicate *predicateGram, ctx instructionContext) ([]*instruction, error) {
	if ctx.function != nil && ctx.function.Native {
		clause := &clauseGram{subject: subject, predicate: predicate, clauses: predicate.clauses}
		return b.clause(clause, ctx)
	}
	var instructions []*instruction
	if event, err := b.declarationInstruction(registrationSite{subject: subject, predicate: predicate}, ctx); err != nil || event != nil {
		if err != nil {
			return nil, err
		}
		following, err := b.clauses(predicate.clauses, ctx)
		return append([]*instruction{event}, following...), err
	}
	rec := b.semantics.Subjects[subject.noun]
	if rec != nil && rec.Value.Runtime != nil && b.semantics.Aliases[rec] == nil {
		current, err := b.newAssignment(subject.noun, ctx)
		if err != nil {
			return nil, err
		}
		current.Value, err = b.objectValue(predicate.obj, ctx)
		if err == nil && current.Value.RuntimeType.Kind == current.Target.Type {
			current.Target.Record.runtimeType = &current.Value.RuntimeType
			current.Target.RuntimeType = current.Value.RuntimeType
		}
		if err != nil {
			return nil, err
		}
		instructions = append(instructions, current)
	}
	following, err := b.clauses(predicate.clauses, ctx)
	return append(instructions, following...), err
}

func (s *semanticProgram) instructionRecord(value *instructionValue) (*record, valueType, error) {
	if value == nil || value.Kind != operandSymbol || value.Record == nil {
		return nil, valueInvalid, lerrors.Internal.Errorf("expected a resolved symbol")
	}
	rec, err := canonicalRecord(value.Record, s.Aliases)
	if value.Record.runtimeFunctionAlias {
		rec = value.Record
	}
	if err != nil {
		return nil, valueInvalid, err
	}
	if rec == nil || rec.Value == nil {
		return nil, valueInvalid, lerrors.Internal.Errorf("symbol has no semantic value")
	}
	kind, err := instructionType(rec)
	if err != nil {
		return nil, valueInvalid, err
	}
	if value.Type != valueInvalid && value.Type != kind {
		return nil, valueInvalid, lerrors.Internal.Errorf("symbol %s has inconsistent value type %s", rec.ID, value.Type)
	}
	return rec, kind, nil
}

// runtimeRecords classifies canonical assignment targets after source actions
// expand. Both orchestrator storage and child-App lowering use these facts.
func (s *semanticProgram) walkInstructions(visit func(*instruction) error) error {
	var walk func([]*instruction) error
	walk = func(body []*instruction) error {
		for _, current := range body {
			if current == nil {
				return lerrors.Internal.Errorf("orchestrator contains a nil instruction")
			}
			if err := visit(current); err != nil {
				return current.wrapError(err)
			}
			if current.Kind == instructionFunction {
				if err := walk(current.Function.Body); err != nil {
					return err
				}
			}
			if current.Block != nil {
				if err := walk(current.Block.Body); err != nil {
					return err
				}
			}
			for _, nested := range [][]*instruction{current.Actions, current.Then, current.Else} {
				if err := walk(nested); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := walk(s.Instructions); err != nil {
		return err
	}
	for _, continuation := range s.Continuations {
		if continuation == nil {
			return lerrors.Internal.Errorf("application continuation has no block")
		}
		if err := walk(continuation.Body); err != nil {
			return err
		}
	}
	return nil
}

func (s *semanticProgram) runtimeRecords() (map[*record]bool, error) {
	dynamic := make(map[*record]bool)
	err := s.walkInstructions(func(current *instruction) error {
		for _, value := range current.Results {
			rec, _, err := s.instructionRecord(value)
			if err != nil {
				return err
			}
			dynamic[rec] = true
		}
		if current.Kind == instructionFunction {
			for _, parameter := range current.Function.Inputs {
				rec, err := canonicalRecord(parameter.Record, s.Aliases)
				if err != nil {
					return err
				}
				dynamic[rec] = true
			}
		}
		if current.Kind == instructionAssign || current.Kind == instructionFunction {
			rec, _, err := s.instructionRecord(current.Target)
			if err != nil {
				return err
			}
			dynamic[rec] = true
		}
		return nil
	})
	return dynamic, err
}
