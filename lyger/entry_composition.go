package Lyger

func (c *Compilation) processDefinition(body []string) (DefinitionInstance, error) {
	if len(body) == 0 {
		return c.instantiateDescribedBuiltin("Process", []string{"Storage.Empty"}, nil)
	}
	children := make([]DefinitionInstance, len(body))
	for index, operation := range body {
		children[index] = DefinitionInstance{Expression: operation}
	}
	return c.instantiateBuiltin("Process", map[string]DefinitionArgument{"Definitions": {Kind: "Syllables", Syllables: body, sequence: children}})
}

// A native entry is an ordinary selected Function whose body constructs and
// returns a composed future. Its signature is metadata, not an executable body.
func (c *Compilation) futureEntry(name, output string, body []string) (DefinitionInstance, error) {
	process, err := c.processDefinition(body)
	if err != nil {
		return DefinitionInstance{}, err
	}
	future, err := c.instantiateBuiltin("Storage.Future", map[string]DefinitionArgument{
		"Name": definitionReferenceArgument("lEntryFuture"), "Body": {Kind: "Definition", Definition: &process},
	})
	if err != nil {
		return DefinitionInstance{}, err
	}
	result, err := c.instantiateBuiltin("Function.Return", map[string]DefinitionArgument{"Statement": definitionReferenceArgument("lEntryFuture")})
	if err != nil {
		return DefinitionInstance{}, err
	}
	returned, err := builtinProcessExpressions(&DefinitionTemplate{Span: result.Span}, result)
	if err != nil {
		return DefinitionInstance{}, err
	}
	operation, err := c.processDefinition(append([]string{future.Expression}, returned...))
	if err != nil {
		return DefinitionInstance{}, err
	}
	return c.instantiateBuiltin("Function", map[string]DefinitionArgument{
		"Name": {Kind: "Text", Text: name}, "Inputs": {Kind: "Props", Props: []DefinitionProperty{{Key: "runtime", Type: "&runtime::Runtime"}}},
		"Outputs": {Kind: "Texts", Texts: []string{"impl std::future::Future<Output = " + output + "> + '_"}},
		"Process": {Kind: "Definition", Definition: &operation},
	})
}
