package Lyger

import (
	"fmt"
	"strconv"
	"unicode/utf8"

	"github.com/l-programming-language/l/lerrors"
)

// word selects the actual builtin declaration from this compilation's provider.
// Its arguments are data atoms or previously composed definitions, never bodies
// synthesized in the frontend's target language.
func (e *orchestrationEmitter) word(name string, arguments map[string]DefinitionArgument) ([]string, error) {
	return e.describedWord(name, nil, arguments)
}

func (e *orchestrationEmitter) describedWord(name string, descriptors []string, arguments map[string]DefinitionArgument) ([]string, error) {
	if e.compilation == nil {
		return nil, lerrors.Internal.Errorf("composition requires selected builtin definitions")
	}
	instance, err := e.compilation.instantiateDescribedBuiltin(name, descriptors, arguments)
	if err != nil {
		return nil, err
	}
	e.dependencies = append(e.dependencies, instance.Dependencies...)
	return builtinProcessExpressions(&DefinitionTemplate{Span: instance.Span}, instance)
}

func (e *orchestrationEmitter) bindValue(name, value string, mutable bool) ([]string, error) {
	arguments := map[string]DefinitionArgument{"Name": definitionReferenceArgument(name), "Value": definitionReferenceArgument(value)}
	if mutable {
		return e.describedWord("Storage.Binding", []string{"Storage.Mutable"}, arguments)
	}
	return e.word("Storage.Binding", arguments)
}

func (e *orchestrationEmitter) namedBinding(name, word string, arguments map[string]DefinitionArgument) ([]string, error) {
	value, err := e.compilation.instantiateBuiltin(word, arguments)
	if err != nil {
		return nil, err
	}
	return e.describedWord("Storage.Binding", []string{"Storage.Produced"}, map[string]DefinitionArgument{"Name": definitionReferenceArgument(name), "Value": {Kind: "Definition", Definition: &value}})
}

func (e *orchestrationEmitter) valueBinding(word string, arguments map[string]DefinitionArgument) (string, error) {
	return e.describedValueBinding(word, nil, arguments)
}

func (e *orchestrationEmitter) describedValueBinding(word string, descriptors []string, arguments map[string]DefinitionArgument) (string, error) {
	if arguments == nil {
		arguments = make(map[string]DefinitionArgument)
	}
	name := fmt.Sprintf("lBound%d", e.valueCount)
	e.valueCount++
	rec, err := e.compilation.builtinRecord(word)
	if err != nil {
		return "", err
	}
	var body []string
	if function := rec.function; function != nil && function.Native {
		invocation, err := e.compilation.instantiateDescribedBuiltin(word, descriptors, arguments)
		if err != nil {
			return "", err
		}
		body, err = e.describedWord("Storage.Binding", []string{"Storage.Produced"}, map[string]DefinitionArgument{
			"Name": definitionReferenceArgument(name), "Value": {Kind: "Definition", Definition: &invocation},
		})
		if err != nil {
			return "", err
		}
	} else {
		arguments["Name"] = definitionReferenceArgument(name)
		body, err = e.describedWord(word, descriptors, arguments)
	}
	if err != nil {
		return "", err
	}
	e.pending = append(e.pending, body...)
	return name, nil
}

func (e *orchestrationEmitter) resultBinding(word string, arguments map[string]DefinitionArgument) (string, error) {
	value, err := e.valueBinding(word, arguments)
	if err != nil {
		return "", err
	}
	return e.checkedBinding(value)
}

func (e *orchestrationEmitter) borrowArgument(arguments map[string]DefinitionArgument, key string, mutable bool) error {
	var descriptors []string
	if mutable {
		descriptors = []string{"Storage.Editable"}
	}
	value, err := e.describedValueBinding("Storage.Borrow", descriptors, map[string]DefinitionArgument{"Value": arguments[key]})
	if err != nil {
		return err
	}
	arguments[key] = definitionReferenceArgument(value)
	return nil
}

func (e *orchestrationEmitter) borrowedResultBinding(word string, arguments map[string]DefinitionArgument, key string) (string, error) {
	if err := e.borrowArgument(arguments, key, false); err != nil {
		return "", err
	}
	return e.resultBinding(word, arguments)
}

func (e *orchestrationEmitter) checkedBinding(value string) (string, error) {
	return e.describedValueBinding("Storage.Binding", []string{"Storage.Checked"}, map[string]DefinitionArgument{"Value": definitionReferenceArgument(value)})
}

func (e *orchestrationEmitter) vector(values []DefinitionArgument, elementType string) (string, error) {
	if len(values) == 0 {
		return e.describedValueBinding("Storage.Vector", []string{"Storage.Empty"}, map[string]DefinitionArgument{"Type": definitionTypeArgument(elementType)})
	}
	return e.valueBinding("Storage.Vector", map[string]DefinitionArgument{"Values": definitionAtomListArgument(values...)})
}

func (e *orchestrationEmitter) results(values []DefinitionArgument) (string, error) {
	if len(values) == 0 {
		return e.describedValueBinding("Storage.Tuple", []string{"Storage.Empty"}, nil)
	}
	return e.valueBinding("Storage.Tuple", map[string]DefinitionArgument{"Values": definitionAtomListArgument(values...)})
}

func references(values []string) []DefinitionArgument {
	result := make([]DefinitionArgument, len(values))
	for i, value := range values {
		result[i] = definitionReferenceArgument(value)
	}
	return result
}

func (e *orchestrationEmitter) typeValue(typ runtimeTypeRef) (string, error) {
	var atoms []DefinitionArgument
	var encode func(runtimeTypeRef) error
	encode = func(t runtimeTypeRef) error {
		if _, supported := runtimeTypeName(t.Kind); !supported {
			return lerrors.Unsupported.Errorf("no runtime type for %s", t.Kind)
		}
		atoms = append(atoms, definitionIndexArgument(uint64(t.Kind)))
		switch t.Kind {
		case valueList, valueMap, valuePointer:
			if t.Element == nil {
				return lerrors.TypeMismatch.Errorf("%s requires an element type", t.Kind)
			}
			return encode(*t.Element)
		case valueFunction:
			id, ok := e.plan.Identities[t.Definition]
			if !ok {
				return lerrors.Internal.Errorf("Function has no canonical definition identity")
			}
			atoms = append(atoms, definitionIndexArgument(id), definitionIndexArgument(uint64(len(t.Inputs))))
			for _, t := range t.Inputs {
				if err := encode(t); err != nil {
					return err
				}
			}
			atoms = append(atoms, definitionIndexArgument(uint64(len(t.Outputs))))
			for _, t := range t.Outputs {
				if err := encode(t); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := encode(typ); err != nil {
		return "", err
	}
	return e.vector(atoms, "u64")
}

func (e *orchestrationEmitter) literalValue(value *primitive) (emittedExpression, error) {
	if value == nil {
		return emittedExpression{}, lerrors.Internal.Errorf("literal has no primitive value")
	}
	stored, err := value.value()
	if err != nil {
		return emittedExpression{}, err
	}
	result := emittedExpression{kind: stored.Type}
	var atom DefinitionArgument
	switch stored.Type {
	case valueText, valueBytes:
		var raw string
		if stored.Type == valueText {
			raw = unquoteImportWord([]byte(value.Text))
		} else {
			raw = string(value.Bytes)
		}
		if utf8.ValidString(raw) {
			result.text, err = e.valueBinding("Storage.Text.Own", map[string]DefinitionArgument{"Value": {Kind: "Text", Text: raw}})
		} else {
			values := make([]DefinitionArgument, len(raw))
			for i := 0; i < len(raw); i++ {
				values[i] = definitionIndexArgument(uint64(raw[i]))
			}
			result.text, err = e.vector(values, "u8")
		}
		return result, err
	case valueNull:
		atom = DefinitionArgument{Kind: "Null"}
	case valueTruth:
		atom = DefinitionArgument{Kind: "Truth", Text: strconv.FormatBool(value.Truth)}
	case valueNumber:
		n := value.Number
		if _, finite := nativeNumberLiteral(n); !finite {
			return emittedExpression{}, lerrors.Unsupported.Errorf("non-finite Number is not supported")
		}
		atom = DefinitionArgument{Kind: "Number", Text: strconv.FormatFloat(n, 'g', -1, 64)}
	default:
		return emittedExpression{}, lerrors.Unsupported.Errorf("unsupported primitive %s", stored.Type)
	}
	result.text, err = e.valueBinding("Storage.Literal", map[string]DefinitionArgument{"Value": atom})
	return result, err
}

func (e *orchestrationEmitter) copyValue(value emittedExpression) (string, error) {
	if value.optional {
		return e.resultBinding("Storage.Require", map[string]DefinitionArgument{"Value": definitionReferenceArgument(value.text), "Message": {Kind: "Text", Text: "App has no Error"}})
	}
	if value.kind == valueText && value.arc {
		return e.valueBinding("Storage.Own", map[string]DefinitionArgument{"Value": definitionReferenceArgument(value.text)})
	}
	switch value.kind {
	case valuePointer:
		return e.copySelected("Storage.Pointer.Copy", value.text)
	case valueThing:
		return e.copySelected("Storage.Value.Copy", value.text)
	case valueList:
		return e.copySelected("Storage.List.Copy", value.text)
	case valueMap:
		return e.copySelected("Storage.Map.Copy", value.text)
	case valueFunction:
		return e.copySelected("Storage.Function.Copy", value.text)
	default:
		return e.copyAtom(value.text, value.borrow)
	}
}

func (e *orchestrationEmitter) copySelected(word, value string) (string, error) {
	borrowed, err := e.valueBinding("Storage.Borrow", map[string]DefinitionArgument{"Value": definitionReferenceArgument(value)})
	if err != nil {
		return "", err
	}
	return e.valueBinding(word, map[string]DefinitionArgument{"Value": definitionReferenceArgument(borrowed)})
}

func (e *orchestrationEmitter) copyAtom(value string, borrowed bool) (string, error) {
	if !borrowed {
		var err error
		value, err = e.valueBinding("Storage.Borrow", map[string]DefinitionArgument{"Value": definitionReferenceArgument(value)})
		if err != nil {
			return "", err
		}
	}
	return e.valueBinding("Storage.Copy", map[string]DefinitionArgument{"Value": definitionReferenceArgument(value)})
}

func (e *orchestrationEmitter) boxValue(value emittedExpression, typ runtimeTypeRef) (string, error) {
	copied, err := e.copyValue(value)
	if err != nil {
		return "", err
	}
	if value.kind == valueThing {
		return copied, nil
	}
	signature, err := e.typeValue(typ)
	if err != nil {
		return "", err
	}
	visitor := "Storage.Ignore"
	switch value.kind {
	case valueList, valueMap, valueFunction:
		visitor = "Storage." + value.kind.String() + ".Reach"
	case valuePointer:
		visitor = "Storage.Pointer.Visit"
	}
	var specialization string
	if value.kind == valueFunction {
		specialization, err = nativeFunctionPointer(typ)
		if err != nil {
			return "", err
		}
	}
	return e.wrapValue(copied, signature, "Storage."+value.kind.String()+".Render", visitor, specialization)
}

func (e *orchestrationEmitter) wrapValue(value, signature, rendererName, visitorName, specialization string) (string, error) {
	renderer, err := e.functionReference(rendererName)
	if err != nil {
		return "", err
	}
	visitor, err := e.functionReference(visitorName)
	if err != nil {
		return "", err
	}
	if specialization != "" {
		specialized, err := e.valueBinding("Storage.Specialize", map[string]DefinitionArgument{"Function": visitor, "Type": definitionTypeArgument(specialization)})
		if err != nil {
			return "", err
		}
		visitor = definitionReferenceArgument(specialized)
	}
	return e.valueBinding("Storage.Value.Wrap", map[string]DefinitionArgument{
		"Value": definitionReferenceArgument(value), "Signature": definitionReferenceArgument(signature), "Renderer": renderer, "Visitor": visitor,
	})
}

func (e *orchestrationEmitter) functionReference(name string) (DefinitionArgument, error) {
	rec, err := e.compilation.builtinRecord(name)
	if err != nil {
		return DefinitionArgument{}, err
	}
	function := rec.function
	if function == nil {
		return DefinitionArgument{}, lerrors.TypeMismatch.Errorf("%s is not a source Function", name)
	}
	declaration, err := e.compilation.functionDefinition(function, make(map[*record]bool), nil)
	if err != nil {
		return DefinitionArgument{}, err
	}
	e.dependencies = append(e.dependencies, declaration)
	return definitionReferenceArgument(function.Constructor.Name), nil
}

func (e *orchestrationEmitter) extractValue(value string, typ runtimeTypeRef) (string, error) {
	if typ.Kind == valueThing {
		return value, nil
	}
	snapshotName := "Storage.Clone"
	switch typ.Kind {
	case valueList, valueMap, valuePointer, valueFunction:
		snapshotName = "Storage." + typ.Kind.String() + ".Copy"
	}
	message := "expected " + typ.Kind.String()
	if typ.Kind == valueFunction {
		message = "Function type mismatch"
	}
	nativeType, err := nativeRuntimeType(typ)
	if err != nil {
		return "", err
	}
	return e.restoreValue(value, nativeType, message, snapshotName)
}

func (e *orchestrationEmitter) restoreValue(value, nativeType, message, snapshotName string) (string, error) {
	snapshot, err := e.functionReference(snapshotName)
	if err != nil {
		return "", err
	}
	outcome, err := e.valueBinding("Storage.Value.Restore", map[string]DefinitionArgument{
		"Value": definitionReferenceArgument(value), "Message": {Kind: "Text", Text: message}, "Snapshot": snapshot,
	})
	if err != nil {
		return "", err
	}
	typed, err := e.valueBinding("Storage.Annotation", map[string]DefinitionArgument{"Type": definitionTypeArgument("runtime::Result<" + nativeType + ">"), "Value": definitionReferenceArgument(outcome)})
	if err != nil {
		return "", err
	}
	return e.checkedBinding(typed)
}

func (e *orchestrationEmitter) processDefinition(words []string) (DefinitionInstance, error) {
	return e.compilation.processDefinition(words)
}

func (e *orchestrationEmitter) selectWords(condition string, then, otherwise []string) ([]string, error) {
	body, err := e.processDefinition(then)
	if err != nil {
		return nil, err
	}
	fallback, err := e.processDefinition(otherwise)
	if err != nil {
		return nil, err
	}
	return e.word("Storage.Select", map[string]DefinitionArgument{"Condition": definitionReferenceArgument(condition), "Then": {Kind: "Syllable", Definition: &body}, "Else": {Kind: "Syllable", Definition: &fallback}})
}
