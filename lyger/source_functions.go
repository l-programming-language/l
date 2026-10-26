package Lyger

import (
	"github.com/l-programming-language/l/lerrors"
	"strings"
)

type functionParameter struct {
	Record *record
	Type   runtimeTypeRef
}
type sourceFunction struct {
	aliases        map[*record]*record
	Definition     *record
	Template       *DefinitionTemplate
	Return         *record
	ReturnTemplate *DefinitionTemplate
	Members        map[string]*DefinitionTemplate
	Word           *lexeme
	Record         *record
	Header         *lStatement
	Scope          *lexicalScope
	Inputs         []functionParameter
	Outputs        []runtimeTypeRef
	Body           []*instruction
	Captures       []*instructionValue
	CaptureCalls   []*record
}

func (f *sourceFunction) signature() runtimeTypeRef {
	result := runtimeTypeRef{Kind: valueFunction, Definition: f.Definition, Outputs: f.Outputs}
	for _, parameter := range f.Inputs {
		result.Inputs = append(result.Inputs, parameter.Type)
	}
	return result
}

func (c *Compilation) reserveFunctionParameter(statement *lStatement, tok *token) {
	if c.Scope.declarations[tok] != nil {
		return
	}
	position := 0
	for index, current := range statement.tokens {
		if current == tok {
			position = index
			break
		}
	}
	declaration := &scopeDeclaration{SubjectID: tok.declarationRoot(), Token: tok, TokenIndex: position, Scope: tok.scope}
	tok.scope.declarations = append(tok.scope.declarations, declaration)
	c.Scope.declarations[tok] = declaration
}

func runtimeRecord(tok *token, typ runtimeTypeRef, statement *lStatement) (*record, error) {
	kind, category := checkType([]byte(typ.Kind.String()))
	if typ.Kind == valueError {
		kind, category = "Error", storagePrimitive
	}
	if typ.Kind == valueFunction {
		kind, category = "Function", storageApplicative
	}
	if kind == "Unknown" {
		return nil, lerrors.TypeMismatch.Errorf("%s has no source runtime representation", typ.Kind)
	}
	value := initEntry(typ.Kind.String(), category)
	value.Runtime = &runtimeReference{Source: statement}
	return &record{ID: strings.Trim(string(tok.rawBytes), `"`), Type: strings.ToLower(category.String()) + "." + strings.ToLower(typ.Kind.String()), Value: value, runtimeType: &typ}, nil
}

func (c *Compilation) prepareFunctionSite(statement *lStatement, site registrationSite, prepared *statementRegistration) (bool, error) {
	if site.parameter {
		return true, nil
	}
	if site.resultOf != nil {
		call := site.resultOf
		rec, err := c.resolver().reference(call.verb)
		if err != nil {
			return true, err
		}
		if rec == nil {
			return true, lerrors.TypeMismatch.Errorf("invocation requires a source Function")
		}
		signature, err := recordRuntimeType(rec)
		if err != nil {
			return true, err
		}
		if signature.Kind != valueFunction {
			return true, lerrors.TypeMismatch.Errorf("invocation requires a source Function")
		}
		var labels []*subjectGram
		for _, phrase := range call.phrases {
			if phrase.gType == "phrase.results" {
				labels = append(labels, phrase.subjects...)
			}
		}
		if len(labels) != len(signature.Outputs) {
			return true, lerrors.TypeMismatch.Errorf("Function %s returns %d values, got %d result names", rec.ID, len(signature.Outputs), len(labels))
		}
		for i, label := range labels {
			if label != site.subject {
				continue
			}
			result, err := runtimeRecord(label.noun, signature.Outputs[i], statement)
			if err != nil {
				return true, err
			}
			prepared.declarations = append(prepared.declarations, extractedDeclaration{Token: label.noun, Record: result})
			return true, nil
		}
		return true, lerrors.Internal.Errorf("Function result has no ordered position")
	}
	object := site.valueObject()
	if object == nil || object.cType != "callable" {
		return false, nil
	}
	target, err := site.target()
	if err != nil {
		return true, err
	}
	if target == nil || !target.isDeclare() {
		return true, lerrors.Syntax.Errorf("ordinary Function requires a named declaration")
	}
	scope := target.scope
	if !scope.isParagraph() || statement.source.Span.Start != scope.source.Start {
		return true, lerrors.Syntax.Errorf("Function declaration must open an encapsulated paragraph")
	}
	if len(prepared.functions) != 0 {
		return true, lerrors.Syntax.Errorf("a paragraph has only one Function body")
	}
	if object.noun.binding == nil {
		return true, lerrors.TypeMismatch.Errorf("Function constructor requires a bound callable definition")
	}
	template, err := c.recordTemplate(object.noun.binding.Record)
	if err != nil {
		return true, err
	}
	if template == nil || !template.IsFunction() {
		return true, lerrors.TypeMismatch.Errorf("signature requires a Function constructor")
	}
	word := newLexeme("verb")
	if err := word.reLabel(string(target.rawBytes)); err != nil {
		return true, err
	}
	definition, err := canonicalRecord(object.noun.binding.Record, c.Semantics.Aliases)
	if err != nil {
		return true, err
	}
	function := &sourceFunction{aliases: c.Semantics.Aliases, Header: statement, Scope: scope, Definition: definition, Template: template, Word: word, Members: make(map[string]*DefinitionTemplate)}
	for _, name := range []string{"Return", "Suspend", "Invoke", "Yield", "Bind", "Exit"} {
		binding, _, err := c.resolver().lookupMember(function.Definition, []byte(name))
		if err != nil {
			return true, err
		}
		if binding == nil {
			continue
		}
		function.Members[name], err = c.recordTemplate(binding.Record)
		if err != nil {
			return true, err
		}
		if name == "Return" {
			function.Return, function.ReturnTemplate = binding.Record, function.Members[name]
		}
	}
	value := initEntry("Function", storageApplicative)
	value.Applicative.Function.Source = function
	value.Runtime = &runtimeReference{Source: statement}
	rec := &record{ID: strings.SplitN(strings.Trim(string(target.rawBytes), `"`), "|", 2)[0], Type: "applicative.function", Value: value}
	function.Record = rec
	prepared.declarations = append(prepared.declarations, extractedDeclaration{Token: target, Record: rec})
	seen := make(map[string]bool)
	for _, parameter := range object.phrase.subjects {
		name := parameter.noun.declarationRoot()
		if name == "" || seen[name] {
			return true, lerrors.DuplicateDeclaration.Errorf("duplicate or unnamed Function parameter %s", name)
		}
		seen[name] = true
		if parameter.phrase == nil || parameter.phrase.obj == nil {
			return true, lerrors.TypeMismatch.Errorf("Function parameter requires a source type")
		}
		typ, err := sourceTypeFromObject(parameter.phrase.obj)
		if err == nil {
			typ, err = canonicalRuntimeType(typ, c.Semantics.Aliases)
		}
		if err != nil {
			return true, err
		}
		record, err := runtimeRecord(parameter.noun, typ, statement)
		if err != nil {
			return true, err
		}
		c.reserveFunctionParameter(statement, parameter.noun)
		function.Inputs = append(function.Inputs, functionParameter{record, typ})
		prepared.declarations = append(prepared.declarations, extractedDeclaration{Token: parameter.noun, Record: record})
	}
	for _, output := range object.phrase.orderedObjects() {
		typ, err := sourceTypeFromObject(output)
		if err == nil {
			typ, err = canonicalRuntimeType(typ, c.Semantics.Aliases)
		}
		if err != nil {
			return true, err
		}
		function.Outputs = append(function.Outputs, typ)
	}
	signature := function.signature()
	rec.runtimeType = &signature
	prepared.functions = append(prepared.functions, function)
	return true, nil
}

func (b *instructionBuilder) functionOperation(clause *clauseGram, ctx instructionContext, kind instructionKind) ([]*instruction, error) {
	current := ctx.newInstruction(kind)
	current.SourceToken = clause.verb
	if kind == instructionReturn {
		if ctx.function == nil {
			return nil, lerrors.InvalidInput.Errorf("Return requires a Function activation")
		}
		current.Function = ctx.function
		for _, value := range clause.values {
			operand, err := b.nominalValue(value.noun, value.tokens, ctx)
			if err != nil {
				return nil, err
			}
			current.Arguments = append(current.Arguments, operand)
		}
		if err := checkFunctionValues("Return", ctx.function.Outputs, current.Arguments); err != nil {
			return nil, err
		}
	} else {
		target, err := b.tokenValue(clause.verb, ctx)
		if err != nil {
			return nil, err
		}
		current.Target = target
		signature := target.RuntimeType
		if signature.Kind == valueInvalid && target.Record != nil {
			signature, err = recordRuntimeType(target.Record)
			if err != nil {
				return nil, err
			}
		}
		if signature.Kind != valueFunction {
			return nil, lerrors.TypeMismatch.Errorf("invocation requires Function, got %s", signature.Kind)
		}
		for _, value := range clause.values {
			operand, err := b.nominalValue(value.noun, value.tokens, ctx)
			if err != nil {
				return nil, err
			}
			current.Arguments = append(current.Arguments, operand)
		}
		for _, phrase := range clause.phrases {
			for _, value := range phrase.values {
				operand, err := b.nominalValue(value.noun, value.tokens, ctx)
				if err != nil {
					return nil, err
				}
				current.Arguments = append(current.Arguments, operand)
			}
			if phrase.gType == "phrase.results" {
				for _, label := range phrase.subjects {
					result, err := b.tokenValue(label.noun, ctx)
					if err != nil {
						return nil, err
					}
					current.Results = append(current.Results, result)
				}
			}
		}
		if err := checkFunctionValues("invocation", signature.Inputs, current.Arguments); err != nil {
			return nil, err
		}
		if len(current.Results) != len(signature.Outputs) {
			return nil, lerrors.TypeMismatch.Errorf("invocation requires %d result names, got %d", len(signature.Outputs), len(current.Results))
		}
	}
	return []*instruction{current}, nil
}

func checkFunctionValues(operation string, types []runtimeTypeRef, values []*instructionValue) error {
	if len(types) != len(values) {
		return lerrors.TypeMismatch.Errorf("%s requires %d values, got %d", operation, len(types), len(values))
	}
	for i, wanted := range types {
		actual := values[i].RuntimeType
		if actual.Kind == valueInvalid {
			actual = runtimeTypeRef{Kind: values[i].Type}
		}
		if !wanted.accepts(actual) {
			return lerrors.TypeMismatch.Errorf("%s value %d requires %s, got %s", operation, i+1, wanted.Kind, actual.Kind)
		}
	}
	return nil
}

func (b *instructionBuilder) functionCaptures(function *sourceFunction) ([]*instructionValue, []*record) {
	var captures []*instructionValue
	var calls []*record
	seen, seenCalls := make(map[*record]bool), make(map[*record]bool)
	free := func(rec *record) bool {
		if rec == nil || rec == function.Record {
			return false
		}
		for _, input := range function.Inputs {
			if rec == input.Record {
				return false
			}
		}
		// Header siblings execute outside the body; Public only changes storage.
		scope := rec.storageScope
		if declaration := b.semantics.Declarations[rec]; declaration != nil {
			if declaration == function.Header {
				return true
			}
			scope = b.scopes.environment(declaration.source)
		}
		for ; scope != nil; scope = scope.parent {
			if scope == function.Scope {
				return false
			}
		}
		return true
	}
	call := func(rec *record) {
		if free(rec) && !seenCalls[rec] {
			seenCalls[rec] = true
			calls = append(calls, rec)
		}
	}
	var value func(*instructionValue)
	value = func(current *instructionValue) {
		if current == nil {
			return
		}
		value(current.Owner)
		value(current.Payload)
		value(current.Key)
		for _, element := range current.Elements {
			value(element)
		}
		rec := current.Record
		if free(rec) && current.Type != valueLexeme && !seen[rec] {
			seen[rec] = true
			captures = append(captures, current)
		}
	}
	var body func([]*instruction)
	body = func(instructions []*instruction) {
		for _, current := range instructions {
			value(current.Target)
			value(current.Value)
			value(current.Key)
			for _, argument := range current.Arguments {
				value(argument)
			}
			for _, captured := range current.Captures {
				value(captured)
			}
			for _, captured := range current.CaptureCalls {
				call(captured)
			}
			if current.Condition != nil {
				value(current.Condition.Left)
				value(current.Condition.Right)
			}
			if current.Kind == instructionCall && current.Callable != nil {
				call(current.Callable)
			} else {
				body(current.Actions)
			}
			if current.Kind == instructionAliasAction {
				call(current.Target.Record)
			}
			body(current.Then)
			body(current.Else)
			if current.Block != nil {
				body(current.Block.Body)
			}
			if current.Kind == instructionFunction && current.Function != nil {
				values, handles := b.functionCaptures(current.Function)
				for _, captured := range values {
					value(captured)
				}
				for _, captured := range handles {
					call(captured)
				}
			}
		}
	}
	body(function.Body)
	return captures, calls
}

// A signature owns its ordered items; collection type phrases cannot consume
// the next parameter or result as another collection operand.
func signatureTypePrefix(tokens []*token) []*token {
	for index, tok := range tokens {
		if index > 0 && tok.hasRole("conjunction", "mutual") {
			return tokens[:index]
		}
	}
	return tokens
}

func (b *instructionBuilder) functionAlias(site registrationSite, ctx instructionContext) ([]*instruction, error) {
	sourceToken, targetToken, err := site.aliasOperands()
	if err != nil {
		return nil, err
	}
	source, err := b.tokenValue(sourceToken, ctx)
	if err != nil {
		return nil, err
	}
	if source.Type != valueFunction {
		return nil, nil
	}
	target := b.semantics.Subjects[targetToken]
	if target == nil {
		return nil, lerrors.Internal.Errorf("Function alias has no binding")
	}
	typ := source.RuntimeType
	target.runtimeType = &typ
	target.runtimeFunctionAlias = true
	current := ctx.newInstruction(instructionAssign)
	current.Target = &instructionValue{Kind: operandSymbol, Type: valueFunction, RuntimeType: typ, Record: target, ID: target.ID, Runtime: true}
	current.Value = source
	return []*instruction{current}, nil
}

// Signature item grammar is selected by its enclosing signature, never by a
// general declaration; otherwise a quoted native definition becomes a type.
func inferredSignatureCases(cases map[string]string) map[string]string {
	return withoutGrammarCase(cases, "signature")
}
func withoutGrammarCase(cases map[string]string, excluded ...string) map[string]string {
	result := make(map[string]string, len(cases))
	for name, rule := range cases {
		omit := false
		for _, explicit := range excluded {
			omit = omit || name == explicit
		}
		if !omit {
			result[name] = rule
		}
	}
	return result
}
