package Lyger

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// definitionSite identifies one authored declaration for both literal and source
// constructions. Aliases are resolved by callers before selecting that site.
func (c *Compilation) definitionSite(rec *record) (*lStatement, *registrationSite, error) {
	source := c.Semantics.Declarations[rec]
	if source == nil {
		return nil, nil, nil
	}
	sites, err := source.sGram.registrationSites()
	if err != nil {
		return nil, nil, err
	}
	for _, site := range sites {
		target, err := site.target()
		if err != nil {
			return nil, nil, err
		}
		if c.Semantics.Subjects[target] == rec {
			return source, &site, nil
		}
	}
	return source, nil, nil
}

func (c *Compilation) recordTemplate(rec *record) (*DefinitionTemplate, error) {
	template, _, err := c.recordDefinition(rec)
	return template, err
}

// Literal leaves and source constructions select the same authored site once.
func (c *Compilation) recordDefinition(rec *record) (*DefinitionTemplate, *objectGram, error) {
	rec, err := canonicalRecord(rec, c.Semantics.Aliases)
	if err != nil {
		return nil, nil, err
	}
	if rec == nil {
		return nil, nil, lerrors.Internal.Errorf("definition reference has no record")
	}
	source, site, err := c.definitionSite(rec)
	if err != nil {
		return nil, nil, err
	}
	if source == nil {
		return nil, nil, lerrors.Internal.Errorf("definition %s has no source declaration", rec.ID)
	}
	if site == nil {
		return nil, nil, lerrors.Internal.Errorf("definition %s has no source declaration site", rec.ID)
	}
	literal, err := site.literal()
	if err != nil || literal == nil {
		object := site.valueObject()
		if object != nil && object.noun != nil && object.noun.binding != nil && c.Semantics.Declarations[object.noun.binding.Record] != nil {
			return nil, object, err
		}
		return nil, nil, err
	}
	template, err := ParseDefinitionTemplate(literal.rawBytes, source.source.Span)
	if err == nil {
		template.dependencies, err = c.declarationDependencies(rec)
	}
	return template, nil, err
}

func definitionArguments(template *DefinitionTemplate, args []DefinitionArgument) (map[string]DefinitionArgument, error) {
	bindings := make(map[string]DefinitionArgument)
	next := 0
	for _, slot := range template.Slots {
		if slot.Type == "Atoms" {
			bindings[slot.Name] = definitionAtomListArgument(args[next:]...)
			next = len(args)
			continue
		}
		slotKind, _ := parseDefinitionArgKindExact(slot.Type)
		if slotKind == definitionArgSyllables {
			var expressions []string
			var children []DefinitionInstance
			for ; next < len(args); next++ {
				if args[next].Definition != nil {
					for _, step := range definitionSteps(*args[next].Definition) {
						expressions = append(expressions, step.Expression)
						children = append(children, step)
					}
				} else if kind, _ := parseDefinitionArgKindExact(args[next].Kind); kind == definitionArgSyllable {
					expressions = append(expressions, args[next].Text)
				} else {
					return nil, lerrors.TypeMismatch.Errorf("%s requires definition syllables, got %s", slot.Name, args[next].Kind)
				}
			}
			bindings[slot.Name] = DefinitionArgument{Kind: definitionArgSyllables.String(), Syllables: expressions, sequence: children}
			continue
		}
		if next >= len(args) {
			return nil, lerrors.InvalidInput.Errorf("missing argument for %s", slot.Name)
		}
		argument := args[next]
		if slot.Type == "" {
			kind, _ := parseDefinitionArgKindExact(argument.Kind)
			switch kind {
			case definitionArgText, definitionArgNumber, definitionArgTruth, definitionArgNull:
				argument = definitionAtomListArgument(argument)
			}
		}
		bindings[slot.Name] = argument
		next++
	}
	if next != len(args) {
		return nil, lerrors.InvalidInput.Errorf("definition received %d unused arguments", len(args)-next)
	}
	return bindings, nil
}

func (c *Compilation) instantiateCall(rec *record, args []DefinitionArgument, active map[*record]bool) (DefinitionInstance, error) {
	if len(c.definitionTemporaries) == 0 {
		defer c.definitionScope()()
	}
	dynamic, err := c.Semantics.runtimeRecords()
	if err != nil {
		return DefinitionInstance{}, err
	}
	return c.instantiateStaticCall(rec, args, active, dynamic)
}

func (c *Compilation) instantiateStaticCall(rec *record, args []DefinitionArgument, active, dynamic map[*record]bool) (DefinitionInstance, error) {
	rec, err := canonicalRecord(rec, c.Semantics.Aliases)
	if err != nil {
		return DefinitionInstance{}, err
	}
	if rec == nil {
		return DefinitionInstance{}, lerrors.Internal.Errorf("definition invocation has no record")
	}
	if dynamic[rec] || (rec.Value != nil && rec.Value.Runtime != nil) {
		return DefinitionInstance{}, lerrors.Unsupported.Errorf("runtime value %s cannot be captured by an App body", rec.ID)
	}
	if function := rec.function; function != nil && function.Native {
		return c.invokeFunctionDefinition(function, args, active, dynamic)
	}
	if active[rec] {
		return DefinitionInstance{}, lerrors.Cycle.Errorf("recursive definition invocation: %s", rec.ID)
	}
	active[rec] = true
	defer delete(active, rec)
	template, object, err := c.recordDefinition(rec)
	if err != nil {
		return DefinitionInstance{}, err
	}
	if template != nil {
		bindings, err := definitionArguments(template, args)
		if err != nil {
			return DefinitionInstance{}, fmt.Errorf("%s: %w", rec.ID, err)
		}
		instance, err := c.instantiateTemplate(template, bindings)
		instance.Symbol = rec.ID
		instance.definition = rec
		return instance, err
	}
	if object != nil {
		instance, forwarded, err := c.instantiateStaticObject(object, args, active, dynamic)
		if err == nil && len(args) > 0 && !forwarded {
			err = lerrors.InvalidInput.Errorf("%s received %d unused arguments", rec.ID, len(args))
		}
		instance.Symbol, instance.definition = rec.ID, rec
		return instance, err
	}
	var before []DefinitionInstance
	for _, ordered := range c.Scope.Ordered {
		if ordered.Statement != c.Semantics.Declarations[rec] || len(ordered.Arguments) == 0 {
			continue
		}
		if len(args) != len(ordered.Arguments) {
			return DefinitionInstance{}, lerrors.InvalidInput.Errorf("source action requires %d arguments, got %d", len(ordered.Arguments), len(args))
		}
		process, err := c.builtinRecord("Process")
		if err != nil {
			return DefinitionInstance{}, err
		}
		args = append([]DefinitionArgument(nil), args...)
		for index, formal := range ordered.Arguments {
			typ, err := c.resolver().reference(formal.Token)
			if err != nil {
				return DefinitionInstance{}, err
			}
			if typ == process {
				continue
			}
			argument, err := c.atomArgument(args[index])
			if err != nil {
				return DefinitionInstance{}, err
			}
			var steps []DefinitionInstance
			args[index], steps = argument.boundResult()
			before = append(before, steps...)
		}
		break
	}
	forwarded := false
	resultSymbol := ""
	var resultDefinition *record
	bound, err := c.composeSourceActions(rec, func(action *clauseGram) (DefinitionInstance, error) {
		instance, used, err := c.instantiateStaticAction(action, args, active, dynamic)
		if err != nil {
			return DefinitionInstance{}, err
		}
		forwarded = forwarded || used
		if resultDefinition != nil && (resultDefinition != instance.definition || resultSymbol != instance.Symbol) {
			return DefinitionInstance{}, lerrors.TypeMismatch.Errorf("action definition %s returns incompatible nominal types", rec.ID)
		}
		resultSymbol, resultDefinition = instance.Symbol, instance.definition
		return instance, nil
	})
	if err != nil {
		return DefinitionInstance{}, err
	}
	if len(args) > 0 && !forwarded {
		return DefinitionInstance{}, lerrors.InvalidInput.Errorf("%s received %d unused arguments", rec.ID, len(args))
	}
	bound.before = append(before, bound.before...)
	bound.Symbol, bound.definition = resultSymbol, resultDefinition
	return bound, nil
}

// Both named and positional invocations expand the same source action sequence.
// Operand binding stays with the caller; ordering, provenance and dependency
// ownership remain one generic definition operation.
func (c *Compilation) composeSourceActions(rec *record, invoke func(*clauseGram) (DefinitionInstance, error)) (DefinitionInstance, error) {
	source := c.Semantics.Declarations[rec]
	if source == nil || source.sGram.predicate == nil || len(source.sGram.predicate.clauses) == 0 {
		return DefinitionInstance{}, lerrors.InvalidInput.Errorf("%s has no action definition", rec.ID)
	}
	clause := source.sGram.predicate.clauses[0]
	if clause.subject == nil || clause.subject.phrase == nil {
		return DefinitionInstance{}, lerrors.InvalidInput.Errorf("%s has no directive body", rec.ID)
	}
	dependencies, err := c.declarationDependencies(rec)
	if err != nil {
		return DefinitionInstance{}, err
	}
	bound := DefinitionInstance{Symbol: rec.ID, definition: rec, Original: append([]byte(nil), source.source.Raw...), Span: source.source.Span, Dependencies: dependencies}
	var actions []DefinitionInstance
	var expressions []string
	for _, action := range clause.subject.phrase.clauses {
		if action.verb == nil || action.verb.binding == nil {
			return DefinitionInstance{}, lerrors.TypeMismatch.Errorf("source action in %s is not a definition", rec.ID)
		}
		instance, err := invoke(action)
		if err != nil {
			return DefinitionInstance{}, err
		}
		if instance.Expression == "" {
			return DefinitionInstance{}, lerrors.TypeMismatch.Errorf("source action %s requires a composed syllable", rec.ID)
		}
		actions = append(actions, instance)
		for _, step := range definitionSteps(instance) {
			expressions = append(expressions, step.Expression)
		}
		bound.Bindings = append(bound.Bindings, instance.Bindings...)
		bound.Dependencies = append(bound.Dependencies, instance.Dependencies...)
	}
	if len(actions) == 0 {
		return DefinitionInstance{}, lerrors.InvalidInput.Errorf("%s has an empty action definition", rec.ID)
	}
	bound.Expression = strings.Join(expressions, "*")
	bound.sequence = pureDefinitionSequence(actions)
	bound.OriginalExpression = bound.Expression
	return bound, nil
}

// instantiateStaticAction is shared by directive expansion and typed native
// Function bodies. It resolves operands at their authored lexical scope.
func (c *Compilation) instantiateStaticAction(action *clauseGram, inherited []DefinitionArgument, active, dynamic map[*record]bool) (DefinitionInstance, bool, error) {
	if action.verb == nil || action.verb.binding == nil {
		return DefinitionInstance{}, false, lerrors.UnresolvedReference.Errorf("source operation has no definition")
	}
	var arguments []DefinitionArgument
	var ledger []DefinitionBinding
	used := false
	for _, object := range action.orderedObjects() {
		argument, forwarded, err := c.staticObjectArgument(object, inherited, active, dynamic)
		if err != nil {
			return DefinitionInstance{}, false, err
		}
		used = used || forwarded
		if argument.Definition != nil {
			ledger = append(ledger, argument.Definition.Bindings...)
		}
		arguments = append(arguments, argument)
	}
	for _, value := range action.values {
		argument, _, err := c.staticObjectArgument(&objectGram{noun: value.noun}, nil, active, dynamic)
		if err != nil {
			return DefinitionInstance{}, false, err
		}
		arguments = append(arguments, argument)
	}
	instance, err := c.instantiateStaticCall(action.verb.binding.Record, arguments, active, dynamic)
	if err != nil {
		return DefinitionInstance{}, false, err
	}
	instance.Bindings = append(ledger, instance.Bindings...)
	return instance, used, nil
}

// The same operand rules serve source directives, constructors and App bodies.
// Only selected source definitions compose operations; other operands stay atoms.
func (c *Compilation) staticObjectArgument(object *objectGram, inherited []DefinitionArgument, active, dynamic map[*record]bool) (DefinitionArgument, bool, error) {
	if object != nil && object.noun != nil {
		target := object.noun
		if pronoun, exists := c.Scope.Pronouns[target]; exists && pronoun.Kind == scopeActionArgument {
			target = pronoun.TargetToken
		}
		for _, ordered := range c.Scope.Ordered {
			for index, formal := range ordered.Arguments {
				if formal.Token != target {
					continue
				}
				if len(inherited) != len(ordered.Arguments) {
					return DefinitionArgument{}, false, lerrors.InvalidInput.Errorf("source action requires %d arguments, got %d", len(ordered.Arguments), len(inherited))
				}
				argument := inherited[index]
				if kind, _ := parseDefinitionArgKindExact(argument.Kind); kind >= definitionArgText && kind <= definitionArgNull && argument.reference == "" {
					name, err := nominalType(formal.Token)
					if err != nil {
						return DefinitionArgument{}, false, err
					}
					if _, err := argument.renderFor(DefinitionSlot{Type: name}); err != nil {
						return DefinitionArgument{}, false, err
					}
				}
				return argument, true, nil
			}
		}
	}
	return c.definitionOperand(object, active, dynamic, func(child *objectGram) (DefinitionInstance, bool, error) {
		return c.instantiateStaticObject(child, inherited, active, dynamic)
	})
}

func (c *Compilation) definitionOperand(object *objectGram, active, dynamic map[*record]bool, compose func(*objectGram) (DefinitionInstance, bool, error)) (DefinitionArgument, bool, error) {
	if object == nil || object.noun == nil {
		return DefinitionArgument{}, false, lerrors.InvalidInput.Errorf("source operand has no value")
	}
	if reference, ok, err := c.functionReferenceArgument(object, active, dynamic); err != nil || ok {
		return reference, false, err
	}
	if object.noun.binding != nil || object.noun.isPronoun() {
		rec, err := c.resolver().reference(object.noun)
		if err != nil {
			return DefinitionArgument{}, false, err
		}
		if rec != nil {
			if dynamic[rec] || rec.Value != nil && rec.Value.Runtime != nil {
				return DefinitionArgument{}, false, lerrors.Unsupported.Errorf("runtime value %s cannot be captured by an App body", rec.ID)
			}
			if rec.definitionArgument != nil {
				return *rec.definitionArgument, false, nil
			}
			if rec.Value != nil && rec.Value.Type == storageLexeme {
				child, used, err := compose(object)
				return DefinitionArgument{Kind: definitionArgDefinition.String(), Definition: &child}, used, err
			}
		}
	}
	argument, err := c.staticOperandArgument(object.noun)
	return argument, false, err
}

// instantiateStaticObject binds a constructed operand's own parameters before
// applying its definition. Inherited call arguments are only used when the
// constructor leaves its parameters open for the surrounding source directive.
func (c *Compilation) instantiateStaticObject(object *objectGram, inherited []DefinitionArgument, active, dynamic map[*record]bool) (DefinitionInstance, bool, error) {
	selected, arguments, err := c.bindFormalMember(object, inherited)
	if err != nil {
		return DefinitionInstance{}, false, err
	}
	node, err := c.constructObject(selected, make(map[*descriptorRule]bool))
	if err != nil {
		return DefinitionInstance{}, false, err
	}
	node.Source = object
	if node.Name == "" {
		template, err := c.recordTemplate(node.Definition)
		if err != nil {
			return DefinitionInstance{}, false, err
		}
		if template != nil && len(template.Slots) > 0 && template.Slots[0].Type == "Label" {
			node.Name = c.definitionLabel()
		}
	}
	return c.instantiateStaticConstruction(node, arguments, active, dynamic)
}

// A possessive formal owns member lookup at its actual argument's type. The
// source's formal token, member spelling and lexical environment supply the
// selection; no operation/type helper registry or caller namespace is consulted.
func (c *Compilation) bindFormalMember(object *objectGram, arguments []DefinitionArgument) (*objectGram, []DefinitionArgument, error) {
	var owner *token
	for i, tok := range object.tokens {
		if tok == object.noun && i > 0 {
			if _, possessive := possessiveOwner(string(object.tokens[i-1].rawBytes)); possessive {
				owner = object.tokens[i-1]
			}
			break
		}
	}
	if owner == nil {
		return object, arguments, nil
	}
	if pronoun, exists := c.Scope.Pronouns[owner]; exists && pronoun.Kind == scopeActionArgument {
		owner = pronoun.TargetToken
	}
	for _, ordered := range c.Scope.Ordered {
		for index, formal := range ordered.Arguments {
			if formal.Token != owner {
				continue
			}
			if len(arguments) != len(ordered.Arguments) {
				return nil, nil, lerrors.InvalidInput.Errorf("source action requires %d arguments, got %d", len(ordered.Arguments), len(arguments))
			}
			argument := arguments[index]
			name, _ := possessiveOwner(string(formal.Token.rawBytes))
			wanted, _ := checkType([]byte(name))
			if _, err := argument.renderFor(DefinitionSlot{Type: wanted}); err != nil {
				return nil, nil, ordered.Source.wrapError(err)
			}
			resolved, ready, err := c.resolver().member(semanticOwner{Namespace: argument.Kind}, object.noun, ordered.Statement)
			if err != nil {
				return nil, nil, err
			}
			if !ready || resolved.Binding == nil {
				return nil, nil, lerrors.UnresolvedReference.Errorf("%s has no selected member %s", argument.Kind, object.noun.rawBytes)
			}
			selected := object.clone()
			noun := *object.noun
			noun.binding, noun.mem, noun.tType = resolved.Binding, resolved.Value, resolved.Type
			selected.noun = &noun
			return &selected, []DefinitionArgument{argument}, nil
		}
	}
	return object, arguments, nil
}

func (c *Compilation) instantiateStaticConstruction(node *constructionNode, inherited []DefinitionArgument, active, dynamic map[*record]bool) (DefinitionInstance, bool, error) {
	template, err := c.recordTemplate(node.Definition)
	if err != nil {
		return DefinitionInstance{}, false, err
	}
	arguments := inherited
	forwarded := len(node.Arguments) == 0
	if template != nil && len(template.Slots) == 0 && len(node.Arguments) == 0 {
		arguments, forwarded = nil, false
	}
	if len(node.Arguments) > 0 {
		arguments = nil
		for _, parameter := range node.Arguments {
			argument, used, err := c.staticObjectArgument(parameter, inherited, active, dynamic)
			if err != nil {
				return DefinitionInstance{}, false, err
			}
			forwarded = forwarded || used
			arguments = append(arguments, argument)
		}
	}
	if instructions, exists, err := c.directiveInstructions(node.Source); err != nil {
		return DefinitionInstance{}, false, err
	} else if exists {
		body, err := c.composeInstructions(instructions, active, dynamic)
		if err != nil {
			return DefinitionInstance{}, false, err
		}
		arguments = append(arguments, DefinitionArgument{Kind: definitionArgDefinition.String(), Definition: &body})
	}
	var instance DefinitionInstance
	if template != nil && template.IsFunction() {
		if len(arguments) != 1 {
			return DefinitionInstance{}, false, lerrors.InvalidInput.Errorf("Function constructor must own one body composition")
		}
		constructor := *node
		constructor.Arguments = nil // The explicit body was composed above.
		instance, err = c.instantiateFunctionConstructor(&constructor, arguments[0], false, active)
		return instance, forwarded, err
	} else if function := node.Definition.function; function != nil && function.Native {
		instance, err = c.invokeFunctionDefinition(function, arguments, active, dynamic)
		if err == nil && node.Name != "" {
			instance, err = c.instantiateDescribedBuiltin("Storage.Binding", []string{"Storage.Produced"}, map[string]DefinitionArgument{"Name": definitionReferenceArgument(node.Name), "Value": {Kind: definitionArgDefinition.String(), Definition: &instance}})
		}
	} else {
		if node.Name != "" {
			argument := definitionReferenceArgument(node.Name)
			if node.Source.name != nil {
				argument.source = c.operandSource(node.Source.name.noun)
			}
			if template != nil && len(template.Slots) > 0 && template.Slots[0].Type == definitionArgText.String() {
				argument = DefinitionArgument{Kind: definitionArgText.String(), Text: node.Name}
			}
			arguments = append([]DefinitionArgument{argument}, arguments...)
		}
		if template != nil {
			if dynamic[node.Definition] || node.Definition.Value != nil && node.Definition.Value.Runtime != nil {
				return DefinitionInstance{}, false, lerrors.Unsupported.Errorf("runtime value %s cannot be captured by an App body", node.Definition.ID)
			}
			bindings, bindErr := definitionArguments(template, arguments)
			if bindErr != nil {
				return DefinitionInstance{}, false, bindErr
			}
			instance, err = c.instantiateNamedConstruction(node, bindings, make(map[string]bool), active)
			if err == nil && node.Name != "" {
				result := definitionReferenceArgument(node.Name)
				result.source = c.operandSource(node.Source.noun)
				instance.result = &result
			}
			return instance, forwarded, err
		}
		instance, err = c.instantiateStaticCall(node.Definition, arguments, active, dynamic)
	}
	if err != nil {
		return DefinitionInstance{}, false, err
	}
	for i := len(node.Wrappers) - 1; i >= 0; i-- {
		rule := node.Wrappers[i]
		wrapper, err := ParseDefinitionTemplate(rule.Template.rawBytes, rule.Source.source.Span)
		if err != nil {
			return DefinitionInstance{}, false, err
		}
		bindings, err := definitionArguments(wrapper, []DefinitionArgument{{Kind: definitionArgDefinition.String(), Definition: &instance}})
		if err != nil {
			return DefinitionInstance{}, false, err
		}
		instance, err = c.instantiateTemplate(wrapper, bindings)
		if err != nil {
			return DefinitionInstance{}, false, err
		}
	}
	instance.definition, err = canonicalRecord(node.Definition, c.Semantics.Aliases)
	if err != nil {
		return DefinitionInstance{}, false, err
	}
	instance.Symbol = instance.definition.ID
	if node.Name != "" {
		result := definitionReferenceArgument(node.Name)
		result.source = c.operandSource(node.Source.noun)
		instance.result = &result
	}
	return instance, forwarded, nil
}

func (c *Compilation) operandSource(tok *token) SourceSpan {
	for _, scoped := range c.Scope.Ordered {
		for _, candidate := range scoped.Statement.tokens {
			if candidate == tok {
				return scoped.Statement.source.Span
			}
		}
	}
	return SourceSpan{}
}

func (c *Compilation) staticOperandArgument(tok *token) (DefinitionArgument, error) {
	argument, err := argumentFromToken(tok)
	if err == nil {
		argument.source = c.operandSource(tok)
	}
	return argument, err
}

func argumentFromToken(tok *token) (DefinitionArgument, error) {
	if tok != nil && tok.binding != nil && tok.binding.Record.definitionArgument != nil {
		return *tok.binding.Record.definitionArgument, nil
	}
	if tok == nil || tok.mem == nil {
		return DefinitionArgument{}, lerrors.Internal.Errorf("argument has no value")
	}
	if tok.mem.Runtime != nil {
		return DefinitionArgument{}, lerrors.Unsupported.Errorf("runtime %s.%s is unavailable during source generation", tok.mem.Runtime.OwnerID, tok.mem.Runtime.Member)
	}
	value, err := tok.mem.value()
	if err != nil {
		return DefinitionArgument{}, err
	}
	kind, supported := value.Type.DefinitionArgKind()
	if !supported {
		return DefinitionArgument{}, lerrors.Unsupported.Errorf("unsupported argument value %s", value.Type)
	}
	argument := DefinitionArgument{Kind: kind.String()}
	switch kind {
	case definitionArgText:
		argument.Text = unquoteImportWord([]byte(value.Data.(string))) //nolint:errcheck // value() validates the tagged payload before this type-specific projection.
	case definitionArgNumber:
		argument.Text = strconv.FormatFloat(value.Data.(float64), 'g', -1, 64) //nolint:errcheck // value() validates the tagged payload before this type-specific projection.
	case definitionArgTruth:
		argument.Text = strconv.FormatBool(value.Data.(bool)) //nolint:errcheck // value() validates the tagged payload before this type-specific projection.
	}
	return argument, nil
}

func (c *Compilation) statementApp(stmt *lStatement) (*App, error) {
	if stmt.sGram.subject == nil || stmt.sGram.subject.noun == nil {
		return nil, nil
	}
	noun := stmt.sGram.subject.noun
	if noun.isDeclare() {
		return nil, nil
	}
	if noun.binding != nil {
		rec, err := c.resolver().reference(noun)
		if err != nil {
			return nil, err
		}
		if rec != nil {
			value, err := rec.Value.value()
			if err != nil {
				return nil, err
			}
			if value.Type == valueApp {
				return value.Data.(*App), nil //nolint:errcheck // value() validates the tagged payload before this type-specific projection.
			}
			if rec.Value.Type == storageApplicative {
				return nil, nil
			}
		}
	}
	nominal, err := nominalType(noun)
	if err != nil {
		return nil, err
	}
	if nominal != "app" {
		return nil, nil
	}
	for _, scoped := range c.Scope.Ordered {
		if scoped.Statement != stmt {
			continue
		}
		rec, err := c.resolver().reference(scoped.IncomingSubjectToken)
		if err != nil {
			return nil, err
		}
		if rec != nil {
			value, err := rec.Value.value()
			if err != nil {
				return nil, err
			}
			if value.Type == valueApp {
				return value.Data.(*App), nil //nolint:errcheck // value() validates the tagged payload before this type-specific projection.
			}
		}
	}
	return nil, lerrors.UnresolvedReference.Errorf("definite App reference has no application subject in its paragraph")
}

func (c *Compilation) lowerAppBody(app *App) ([]DefinitionInstance, error) {
	defer c.definitionScope()()
	dynamic, err := c.Semantics.runtimeRecords()
	if err != nil {
		return nil, err
	}
	var body []DefinitionInstance
	for _, stmt := range c.Statements {
		if stmt.passiveLifecycle() {
			continue
		}
		target, err := c.statementApp(stmt)
		if err != nil {
			return nil, err
		}
		if target != app {
			continue
		}
		for _, action := range stmt.sGram.predicate.clauses {
			if action.verb == nil || action.verb.binding == nil {
				if action.verb != nil {
					if role, native := action.verb.nativeRole(); native {
						if kind, ok := role.InstructionKind(); ok {
							return nil, stmt.source.Span.wrapError(lerrors.Unsupported.Errorf("%s is unsupported in an App body; lifecycle operations belong to the outer orchestrator", kind))
						}
					}
				}
				return nil, lerrors.Unsupported.Errorf("application action has no source-defined behavior")
			}
			instance, _, err := c.instantiateStaticAction(action, nil, make(map[*record]bool), dynamic)
			if err != nil {
				return nil, err
			}
			body = append(body, instance)
		}
	}
	return body, nil
}

func (c *Compilation) lowerFunction(node *constructionNode, body []DefinitionInstance) (DefinitionInstance, error) {
	return c.instantiateFunctionConstructor(node, definitionSequenceArgument(body), true, make(map[*record]bool))
}

// Both application constructors and source Function declarations bind the same
// selected signature, body slots and adjectives. Application descriptors also
// require an explicit source parameter identifying each body composition.
func (c *Compilation) instantiateFunctionConstructor(node *constructionNode, body DefinitionArgument, declared bool, active map[*record]bool) (DefinitionInstance, error) {
	if node.Definition == nil {
		return DefinitionInstance{}, lerrors.InvalidInput.Errorf("function %s has no definition", node.Name)
	}
	template, err := c.callableTemplate(node.Definition, make(map[*record]bool))
	if err != nil {
		return DefinitionInstance{}, err
	}
	if template == nil || !template.IsFunction() {
		return DefinitionInstance{}, lerrors.TypeMismatch.Errorf("%s does not define a function structure", node.Nominal)
	}
	wrappers, selected, err := constructionTemplates(node)
	if err != nil {
		return DefinitionInstance{}, err
	}
	if selected >= 0 {
		template = wrappers[selected]
	}
	if node.Name == "" || !node.InputsDefined || !node.OutputsDefined {
		return DefinitionInstance{}, lerrors.InvalidInput.Errorf("function constructor requires name and explicit signature")
	}
	bindings, err := template.signatureArguments(node.Name, node.Inputs, node.Outputs)
	if err != nil {
		return DefinitionInstance{}, err
	}
	remaining, supplied := 0, 0
	for _, slot := range template.Slots {
		if _, signature := bindings[slot.Name]; signature {
			continue
		}
		if slot.Type != "" {
			if slot.Type != "Type" || supplied >= len(node.Arguments) {
				return DefinitionInstance{}, lerrors.Unsupported.Errorf("unsupported or unbound function slot %s: %s", slot.Name, slot.Type)
			}
			argument, _, err := c.staticObjectArgument(node.Arguments[supplied], nil, active, nil)
			if err != nil {
				return DefinitionInstance{}, err
			}
			bindings[slot.Name] = argument
			supplied++
			continue
		}
		if declared {
			matched := false
			for _, object := range node.Parameters {
				nominal, err := nominalType(object.noun)
				if err != nil {
					return DefinitionInstance{}, err
				}
				if strings.EqualFold(nominal, slot.Name) && object.noun.binding != nil {
					matched = true
					if supplied < len(node.Arguments) && node.Arguments[supplied] == object {
						supplied++
					}
				}
			}
			if !matched {
				return DefinitionInstance{}, lerrors.InvalidInput.Errorf("function slot %s has no declared constructor parameter", slot.Name)
			}
		}
		if !declared && supplied < len(node.Arguments) {
			argument, _, err := c.staticObjectArgument(node.Arguments[supplied], []DefinitionArgument{body}, active, nil)
			if err != nil {
				return DefinitionInstance{}, err
			}
			body = argument
			supplied++
		}
		bindings[slot.Name] = body
		remaining++
	}
	if supplied != len(node.Arguments) {
		return DefinitionInstance{}, lerrors.InvalidInput.Errorf("Function constructor received %d unused arguments", len(node.Arguments)-supplied)
	}
	if !declared && remaining != 1 {
		return DefinitionInstance{}, lerrors.InvalidInput.Errorf("Function constructor must own one body composition")
	}
	instance, err := c.instantiateNamedConstruction(node, bindings, make(map[string]bool), active)
	if err != nil {
		return DefinitionInstance{}, err
	}
	instance.Dependencies = append(instance.Dependencies, node.Dependencies...)
	return instance, nil
}
