package Lyger

import (
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// composeNativeAML fills compiler-owned grammatical structure with checked
// operands. These templates contain only frozen AML forms, never native bodies.
func composeNativeAML(expression string, span SourceSpan, arguments map[string]DefinitionArgument) (DefinitionInstance, error) {
	template, err := ParseDefinitionTemplate([]byte(expression), span)
	if err != nil {
		return DefinitionInstance{}, err
	}
	return template.Instantiate(arguments)
}

// bindDefinitionResult stages an expression exactly once. Explicit source
// results retain their ordered prerequisites and are rebound by reference.
func bindDefinitionResult(name, value DefinitionArgument) (DefinitionInstance, error) {
	span := value.source
	original := value.Definition
	if original != nil {
		span = original.Span
	}
	value, before := value.boundResult()
	expression := "OB{|Key: ~<@Name|Label@>~, Value: ~<@Value@>~|}"
	if value.Definition != nil {
		form, err := definitionValueBindingForm(*value.Definition)
		if err != nil {
			return DefinitionInstance{}, err
		}
		expression = form + "{|Key: ~<@Name|Label@>~|}^(<@Value@>)"
	}
	result, err := composeNativeAML(expression, span, map[string]DefinitionArgument{"Name": name, "Value": value})
	if err != nil {
		return DefinitionInstance{}, err
	}
	result.before = append(before, result.before...)
	result.result = &name
	if original != nil {
		result.Bindings = append(append([]DefinitionBinding(nil), original.Bindings...), result.Bindings...)
	}
	return result, nil
}

// Inspect the AML root, not rendered source. Sibling statements and declarations
// do not acquire an implicit last value. The backend validates child roles.
func definitionValueBindingForm(instance DefinitionInstance) (string, error) {
	expression, err := normalizeDefinitionGroups(instance.Expression)
	if err != nil {
		return "", err
	}
	expression = strings.TrimSpace(expression)
	depth := 0
	for i := 0; i < len(expression); i++ {
		if strings.HasPrefix(expression[i:], "{|") {
			end, err := definitionMetadataEnd(expression, i)
			if err != nil {
				return "", err
			}
			i = end + 1
			continue
		}
		switch expression[i] {
		case '(':
			depth++
		case ')':
			depth--
		case '*':
			if depth == 0 {
				return "", lerrors.TypeMismatch.Errorf("composition requires an explicit result binding")
			}
		}
	}
	end := 0
	for end < len(expression) && expression[end] >= 'A' && expression[end] <= 'Z' {
		end++
	}
	var form string
	switch expression[:end] {
	case "E", "DE", "GE", "CE", "ME":
		form = "DOB"
	case "I", "GI", "CI", "MI":
		form = "GOB"
	case "KY", "KYC", "MY", "HY", "FY", "DY":
		form = "MOB"
	case "BW":
		form = "KOB"
	default:
		return "", lerrors.TypeMismatch.Errorf("composition requires an explicit result binding")
	}
	return form, nil
}
