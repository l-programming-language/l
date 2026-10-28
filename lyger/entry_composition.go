package Lyger

import "github.com/l-programming-language/l/lerrors"

func (c *Compilation) processDefinition(body []string) (DefinitionInstance, error) {
	if len(body) == 0 {
		return c.instantiateBuiltin("Emptyprocess", nil)
	}
	return c.instantiateBuiltin("Process", map[string]DefinitionArgument{"Definitions": {Kind: "Syllables", Syllables: body}})
}

// A native entry is an ordinary selected Function whose body constructs and
// returns a composed future. Its signature is metadata, not an executable body.
func (c *Compilation) futureEntry(name, output string, body []string) (DefinitionInstance, error) {
	process, err := c.processDefinition(body)
	if err != nil {
		return DefinitionInstance{}, err
	}
	future, err := c.instantiateBuiltin("Futurevalue", map[string]DefinitionArgument{
		"Name": definitionReferenceArgument("lEntryFuture"), "Body": {Kind: "Definition", Definition: &process},
	})
	if err != nil {
		return DefinitionInstance{}, err
	}
	returned, err := c.builtinMemberDefinition("Function", "Return")
	if err != nil {
		return DefinitionInstance{}, err
	}
	result, err := returned.Instantiate(map[string]DefinitionArgument{"Statement": definitionReferenceArgument("lEntryFuture")})
	if err != nil {
		return DefinitionInstance{}, err
	}
	operation, err := c.processDefinition([]string{future.Expression, result.Expression})
	if err != nil {
		return DefinitionInstance{}, err
	}
	return c.instantiateBuiltin("Function", map[string]DefinitionArgument{
		"Name": {Kind: "Text", Text: name}, "Inputs": {Kind: "Props", Props: []DefinitionProperty{{Key: "runtime", Type: "&runtime::Runtime"}}},
		"Outputs": {Kind: "Texts", Texts: []string{"impl std::future::Future<Output = " + output + "> + '_"}},
		"Process": {Kind: "Definition", Definition: &operation},
	})
}

func (c *Compilation) runtimeDefinitions() ([]string, error) {
	aggregate, err := c.builtinDefinition("Storedvalues")
	if err != nil {
		return nil, err
	}
	arguments := make(map[string]DefinitionArgument, len(aggregate.Slots))
	for _, slot := range aggregate.Slots {
		if slot.Type != "" {
			return nil, lerrors.InvalidConfiguration.Errorf("Storedvalues dependency %s must be a definition", slot.Name)
		}
		dependency, err := c.instantiateBuiltin(slot.Name, nil)
		if err != nil {
			return nil, err
		}
		arguments[slot.Name] = DefinitionArgument{Kind: "Definition", Definition: &dependency}
	}
	bound, err := aggregate.Instantiate(arguments)
	if err != nil {
		return nil, err
	}
	if bound.Expression != "" {
		return []string{bound.Expression}, nil
	}
	var definitions []string
	for _, word := range bound.Words {
		definitions = append(definitions, word...)
	}
	return definitions, nil
}
