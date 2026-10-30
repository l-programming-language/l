package Lyger

import (
	"github.com/l-programming-language/l/lerrors"
	"strconv"
	"strings"
)

// A local constructor declaration names its result through the selected
// definition's existing name slot. The binding is lexical and source-owned.
func (c *Compilation) prepareDefinitionBindingSite(statement *lStatement, site registrationSite, prepared *statementRegistration) (bool, error) {
	target, err := site.target()
	if err != nil {
		return false, err
	}
	if target == nil || target.scope == nil || target.scope.function == nil || !target.scope.function.Native || target.scope.function.Header == statement {
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
	word := newLexeme("noun")
	if err := word.reLabel(string(target.rawBytes)); err != nil {
		return true, err
	}
	if node.Name == "" {
		node.Name = sourceModuleName(strings.ToLower(word.Root))
	}
	argument := definitionReferenceArgument(node.Name)
	argument.source = statement.source.Span
	rec := &record{ID: word.Root, Type: "lexeme.noun.generic", Value: &entry{Type: storageLexeme, Lexeme: word}, definitionArgument: &argument, staticConstructor: node}
	prepared.declarations = append(prepared.declarations, extractedDeclaration{Token: target, Record: rec})
	return true, nil
}

func (c *Compilation) invokeFunctionDefinition(function *sourceFunction, args []DefinitionArgument, active, dynamic map[*record]bool) (DefinitionInstance, error) {
	if len(args) != len(function.Constructor.Inputs) {
		return DefinitionInstance{}, lerrors.InvalidInput.Errorf("Function %s requires %d arguments, got %d", function.Record.ID, len(function.Constructor.Inputs), len(args))
	}
	var before []DefinitionInstance
	args = append([]DefinitionArgument(nil), args...)
	for index, argument := range args {
		argument, err := c.atomArgument(argument)
		if err != nil {
			return DefinitionInstance{}, err
		}
		var steps []DefinitionInstance
		args[index], steps = argument.boundResult()
		before = append(before, steps...)
	}
	bindings := map[string]DefinitionArgument{"Target": definitionReferenceArgument(function.Constructor.Name), "Arguments": definitionAtomListArgument(args...)}
	used := make(map[string]bool)
	invocation, err := c.instantiateNamedDefinition(function.Invoke, bindings, used, active)
	if err == nil {
		err = validateDefinitionArguments(function.Invoke, bindings, used)
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
	if object == nil || object.noun == nil || object.noun.binding == nil || object.name != nil || object.phrase != nil || object.clause != nil {
		return DefinitionArgument{}, false, nil
	}
	for _, modifier := range object.adjectives {
		if modifier.hasRole("adjective", "initialization") {
			return DefinitionArgument{}, false, nil
		}
	}
	rec, err := canonicalRecord(object.noun.binding.Record, c.Semantics.Aliases)
	if err != nil {
		return DefinitionArgument{}, false, err
	}
	function := rec.function
	if function == nil || !function.Native {
		return DefinitionArgument{}, false, nil
	}
	argument := definitionReferenceArgument(function.Constructor.Name)
	argument.source = c.operandSource(object.noun)
	if !active[rec] {
		definition, err := c.functionDefinition(function, active, dynamic)
		if err != nil {
			return DefinitionArgument{}, false, err
		}
		argument.dependencies = append(argument.dependencies, definition)
	}
	return argument, true, nil
}

func (c *Compilation) functionDefinition(function *sourceFunction, active, dynamic map[*record]bool) (DefinitionInstance, error) {
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
	emitter := &orchestrationEmitter{compilation: c, definitionActive: active, definitionDynamic: dynamic}
	operations, err := emitter.instructions(body)
	if err != nil {
		return DefinitionInstance{}, err
	}
	composition := DefinitionInstance{Expression: strings.Join(operations, "*"), Dependencies: emitter.dependencies, Bindings: emitter.definitionBindings, sequence: make([]DefinitionInstance, len(operations))}
	for index, operation := range operations {
		composition.sequence[index] = DefinitionInstance{Expression: operation}
	}
	return composition, nil
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
