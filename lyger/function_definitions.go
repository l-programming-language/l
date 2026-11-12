package Lyger

import (
	"strconv"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// Native inputs and local results carry source-owned reference labels.
func nativeReferenceRecord(statement *lStatement, tok *token, name string) (*record, error) {
	word := newLexeme("noun")
	if err := word.reLabel(string(tok.rawBytes)); err != nil {
		return nil, err
	}
	argument := definitionReferenceArgument(name)
	argument.source = statement.source.Span
	return &record{ID: word.Root, Type: "lexeme.noun.generic", Value: &entry{Type: storageLexeme, Lexeme: word}, definitionArgument: &argument}, nil
}

// A local constructor declaration names its result through the selected
// definition's existing name slot. The binding is lexical and source-owned.
func (c *Compilation) prepareDefinitionBindingSite(statement *lStatement, site registrationSite, prepared *statementRegistration) (bool, error) {
	if site.parameter || site.isDescriptor() {
		return false, nil
	}
	target, err := site.target()
	if err != nil {
		return false, err
	}
	if target == nil || target.scope == nil || target.scope.kind != lexicalProcess && (target.scope.function == nil || !target.scope.function.Native || target.scope.function.Header == statement) {
		return false, nil
	}
	object := site.valueObject()
	if object == nil || object.noun == nil || object.noun.binding == nil {
		return false, nil
	}
	node, err := c.constructObject(object, make(map[*descriptorRule]bool))
	if err != nil {
		return true, err
	}
	rec, err := nativeReferenceRecord(statement, target, node.Name)
	if err != nil {
		return true, err
	}
	if node.Name == "" {
		node.Name = sourceModuleName(strings.ToLower(rec.ID))
	}
	rec.definitionArgument.reference, rec.staticConstructor = node.Name, node
	prepared.declarations = append(prepared.declarations, extractedDeclaration{Token: target, Record: rec})
	return true, nil
}

type callableReference struct {
	binding  *record
	function *sourceFunction
}

func (argument DefinitionArgument) validateCallable() error {
	proof := argument.callable
	if proof == nil || proof.binding == nil || proof.function == nil || !proof.function.Native || argument.reference == "" || argument.atoms != nil || argument.typeAtom != "" || argument.Text != "" || argument.Definition != nil || len(argument.sequence) != 0 || len(argument.Syllables) != 0 || len(argument.Props) != 0 || len(argument.Texts) != 0 {
		return lerrors.TypeMismatch.Errorf("Callable requires a resolved native-backed Function reference")
	}
	if !proof.binding.hasInvocationFamily() {
		return lerrors.Syntax.Errorf("Function invocation requires this binding's complete verb family")
	}
	return nil
}

func (function *sourceFunction) checkArgumentCount(count int) error {
	if count != len(function.Constructor.Inputs) {
		return lerrors.InvalidInput.Errorf("Function %s requires %d arguments, got %d", function.Record.ID, len(function.Constructor.Inputs), count)
	}
	return nil
}

func callableArgument(binding *record, function *sourceFunction) DefinitionArgument {
	argument := definitionReferenceArgument(function.Constructor.Name)
	argument.callable = &callableReference{binding: binding, function: function}
	return argument
}

func (c *Compilation) invokeFunctionDefinition(function *sourceFunction, args []DefinitionArgument, active, dynamic map[*record]bool) (DefinitionInstance, error) {
	if err := function.checkArgumentCount(len(args)); err != nil {
		return DefinitionInstance{}, err
	}
	header, err := c.sourceHeader(function.Record, false, nil)
	if err != nil {
		return DefinitionInstance{}, err
	}
	args, _, before, err := c.bindWrapperOperands(header, args, nil)
	if err != nil {
		return DefinitionInstance{}, err
	}
	invoke, err := c.builtinRecord("Invoke")
	if err != nil {
		return DefinitionInstance{}, err
	}
	bindings := map[string]DefinitionArgument{"Target": callableArgument(function.Record, function), "Arguments": definitionAtomListArgument(args...)}
	used := make(map[string]bool)
	invocation, err := c.instantiateSourceCall(invoke, nil, bindings, used, active, dynamic)
	if err == nil {
		err = validateDefinitionArguments(invoke, bindings, used)
	}
	if err != nil {
		return DefinitionInstance{}, err
	}
	if !active[function.Record] {
		declaration, err := c.functionDefinition(function, active, dynamic)
		if err != nil {
			return DefinitionInstance{}, err
		}
		invocation.Dependencies = append(invocation.Dependencies, declaration)
	}
	invocation.before = append(before, invocation.before...)
	invocation.definition = function.Record
	return invocation, nil
}

// A bare Function word supplies its canonical callable reference. Construction
// remains explicit through the initialization modifier or supplied operands.
func (c *Compilation) functionReferenceArgument(object *objectGram, active, dynamic map[*record]bool) (DefinitionArgument, bool, error) {
	binding, err := c.transparentReference(object)
	if err != nil {
		return DefinitionArgument{}, false, err
	}
	function, err := c.nativeFunctionReference(binding)
	if err != nil || function == nil {
		return DefinitionArgument{}, false, err
	}
	argument := callableArgument(binding, function)
	argument.source = c.operandSource(object.noun)
	if !active[function.Record] {
		definition, err := c.functionDefinition(function, active, dynamic)
		if err != nil {
			return DefinitionArgument{}, false, err
		}
		argument.dependencies = append(argument.dependencies, definition)
	}
	return argument, true, nil
}

// These are reference-only views. Construction and descriptor transformations
// retain their own semantics instead of acquiring another word's call authority.
func (c *Compilation) transparentReference(object *objectGram) (*record, error) {
	if object == nil || object.noun == nil || object.noun.binding == nil || object.name != nil || object.directivePhrase() != nil || nominalContext(nil, object.adjectives) {
		return nil, nil
	}
	for _, modifier := range valueModifiers(object.adjectives) {
		descriptor, err := c.descriptorFor(modifier)
		if err != nil || descriptor != nil {
			return nil, err
		}
	}
	node, err := c.constructObject(object, make(map[*descriptorRule]bool))
	if err != nil || len(node.Arguments) != 0 {
		return nil, err
	}
	return object.noun.binding.Record, nil
}

func (c *Compilation) nativeFunctionReference(rec *record) (*sourceFunction, error) {
	seen := make(map[*record]bool)
	for rec != nil {
		var err error
		rec, err = canonicalRecord(rec, c.Semantics.Aliases)
		if err != nil {
			return nil, err
		}
		if rec.function != nil && rec.function.Native {
			return rec.function, nil
		}
		if seen[rec] {
			return nil, lerrors.Cycle.Errorf("recursive definition invocation: %s", rec.ID)
		}
		seen[rec] = true
		if c.Semantics.Declarations[rec] == nil {
			return nil, nil
		}
		selected, err := c.Semantics.selectedDefinition(rec)
		if err != nil {
			return nil, err
		}
		rec, err = c.transparentReference(selected.object)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (c *Compilation) functionDefinition(function *sourceFunction, active, dynamic map[*record]bool) (result DefinitionInstance, resultErr error) {
	defer c.withDefinitionReturn(functionReturnContext(function))()
	defer func() {
		if resultErr == nil && c.contractRecorder != nil {
			c.contractRecorder.recordLowering(function.Record, result, nil)
		}
	}()
	defer c.definitionScope()()
	active[function.Record] = true
	defer delete(active, function.Record)
	if len(function.Body) == 0 {
		return DefinitionInstance{}, lerrors.InvalidInput.Errorf("Function %s has an empty body", function.Record.ID)
	}
	body, err := c.composeInstructions(function.Body, active, dynamic)
	if err != nil {
		return DefinitionInstance{}, err
	}
	process, err := c.instantiateBuiltin("Process", map[string]DefinitionArgument{"Definitions": {Kind: "Syllables", Syllables: []string{body.Expression}, sequence: []DefinitionInstance{body}}})
	if err != nil {
		return DefinitionInstance{}, err
	}
	declaration, err := c.instantiateFunctionConstructor(function.Constructor, DefinitionArgument{Kind: definitionArgDefinition.String(), Definition: &process}, false, active)
	if err != nil {
		return DefinitionInstance{}, err
	}
	owned, err := c.declarationDependencies(function.Record)
	if err != nil {
		return DefinitionInstance{}, err
	}
	declaration.Dependencies = append(declaration.Dependencies, owned...)
	declaration.Bindings = append(declaration.Bindings, body.Bindings...)
	declaration.definition, declaration.Symbol, declaration.Span = function.Record, function.Record.ID, function.Header.source.Span
	return declaration, nil
}

func (c *Compilation) composeInstructions(instructions []*instruction, active, dynamic map[*record]bool) (DefinitionInstance, error) {
	planner := &orchestrationPlanner{semantics: c.Semantics}
	body, _, err := planner.instructions(instructions, newOrchestrationFlow())
	if err != nil {
		return DefinitionInstance{}, err
	}
	composer := &nativeInstructionComposer{compilation: c, active: active, dynamic: dynamic}
	return composer.instructions(body)
}

func (c *Compilation) definitionScope() func() {
	if c.definitionReserved == nil || c.definitionStatements != len(c.Statements) {
		c.definitionReserved = make(map[string]bool)
		c.definitionStatements = len(c.Statements)
		// Labels may be supplied by signature keys, local declarations, or
		// explicit constructor names. Conservatively retain each source token's
		// exact and ordinary-local spelling; no authored name is forbidden.
		for _, statement := range c.Statements {
			for _, token := range statement.tokens {
				name := unquoteImportWord(token.rawBytes)
				if strings.HasPrefix(strings.ToLower(name), "__l_nested_") {
					c.definitionReserved[name] = true
					c.definitionReserved[strings.ToLower(name)] = true
				}
			}
		}
	}
	c.definitionTemporaries = append(c.definitionTemporaries, 0)
	return func() { c.definitionTemporaries = c.definitionTemporaries[:len(c.definitionTemporaries)-1] }
}
func (c *Compilation) definitionLabel() string {
	index := len(c.definitionTemporaries) - 1
	if index < 0 {
		panic("definition label has no expansion scope")
	}
	for {
		name := "__l_nested_" + strconv.Itoa(c.definitionTemporaries[index])
		c.definitionTemporaries[index]++
		if !c.definitionReserved[name] {
			return name
		}
	}
}
