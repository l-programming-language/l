package Lyger

import (
	"strconv"

	"github.com/l-programming-language/l/lerrors"
)

type builtinAppCase struct {
	canonical string
	body      []string
}

// Selection binds native names only. Retention, identity comparison, conditional
// execution and failure are owned by the selected source definitions.
func (e *orchestrationEmitter) composeAppSelection(name string, value *orchestrationValue, target string, cases []builtinAppCase, returned *record) ([]string, error) {
	if value == nil || value.Type != valueApp {
		return nil, lerrors.TypeMismatch.Errorf("App selection requires an App")
	}
	if returned == nil {
		return nil, lerrors.InvalidConfiguration.Errorf("App selection requires its enclosing Function's Return definition")
	}
	retained, err := e.word("Retain", map[string]DefinitionArgument{
		"Name": definitionReferenceArgument(name), "Target": definitionReferenceArgument(target),
	})
	if err != nil {
		return nil, err
	}
	result, err := e.compilation.instantiateDefinition(returned, map[string]DefinitionArgument{"Statement": definitionReferenceArgument(name + "Result")})
	if err != nil {
		return nil, err
	}
	rejected, err := e.word("Storage.Reject", map[string]DefinitionArgument{
		"Name": definitionReferenceArgument(name), "Message": {Kind: "Text", Text: "App handle is not owned by this program"},
		"Return": {Kind: "Syllable", Definition: &result},
	})
	if err != nil {
		return nil, err
	}
	fallback := rejected
	for index := len(cases) - 1; index >= 0; index-- {
		candidate := cases[index]
		matchName := name + "Match" + strconv.Itoa(index)
		matches, err := e.describedWord("Storage.Comparison", []string{"Storage.Identical"}, map[string]DefinitionArgument{
			"Name": definitionReferenceArgument(matchName),
			"Left": definitionReferenceArgument(name), "Right": definitionReferenceArgument(candidate.canonical),
		})
		if err != nil {
			return nil, err
		}
		selected, err := e.selectWords(matchName, candidate.body, fallback)
		if err != nil {
			return nil, err
		}
		fallback = append(matches, selected...)
	}
	return append(retained, fallback...), nil
}
