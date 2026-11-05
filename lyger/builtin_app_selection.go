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
	borrowed, err := e.word("Storage.Borrow", map[string]DefinitionArgument{
		"Name": definitionReferenceArgument(name + "Borrowed"), "Value": definitionReferenceArgument(target),
	})
	if err != nil {
		return nil, err
	}
	retained, err := e.word("Retain", map[string]DefinitionArgument{
		"Name": definitionReferenceArgument(name), "Target": definitionReferenceArgument(name + "Borrowed"),
	})
	if err != nil {
		return nil, err
	}
	reject, err := e.compilation.builtinRecord("Storage.Reject")
	if err != nil {
		return nil, err
	}
	defer e.compilation.withDefinitionReturn(sourceReturnContext{returned: returned})()
	rejected, err := e.compilation.instantiateCall(reject, []DefinitionArgument{{Kind: "Text", Text: "App handle is not owned by this program"}}, make(map[*record]bool))
	if err != nil {
		return nil, err
	}
	fallback, err := builtinProcessExpressions(nil, rejected)
	if err != nil {
		return nil, err
	}
	e.dependencies = append(e.dependencies, rejected.Dependencies...)
	e.definitionBindings = append(e.definitionBindings, rejected.Bindings...)
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
	retained = append(borrowed, retained...)
	return append(retained, fallback...), nil
}
