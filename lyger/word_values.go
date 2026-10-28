package Lyger

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/l-programming-language/l/lerrors"
)

// word selects the actual builtin declaration from this compilation's provider.
// Its arguments are data atoms or previously composed definitions, never bodies
// synthesized in the frontend's target language.
func (e *orchestrationEmitter) word(name string, arguments map[string]DefinitionArgument) ([]string, error) {
	if e.compilation == nil {
		return nil, lerrors.Internal.Errorf("composition requires selected builtin definitions")
	}
	instance, err := e.compilation.instantiateBuiltin(name, arguments)
	if err != nil {
		return nil, err
	}
	return builtinProcessExpressions(&DefinitionTemplate{Span: instance.Span}, instance)
}

func (e *orchestrationEmitter) valueBinding(word string, arguments map[string]DefinitionArgument) (string, error) {
	if arguments == nil {
		arguments = make(map[string]DefinitionArgument)
	}
	name := fmt.Sprintf("lBound%d", e.valueCount)
	e.valueCount++
	arguments["Name"] = definitionReferenceArgument(name)
	body, err := e.word(word, arguments)
	if err != nil {
		return "", err
	}
	e.pending = append(e.pending, body...)
	return name, nil
}

func (e *orchestrationEmitter) vector(values []DefinitionArgument, elementType string) (string, error) {
	if len(values) == 0 {
		return e.valueBinding("Emptyvector", map[string]DefinitionArgument{"Type": definitionTypeArgument(elementType)})
	}
	return e.valueBinding("Vectorvalue", map[string]DefinitionArgument{"Values": definitionAtomListArgument(values...)})
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
			result.text, err = e.valueBinding("Ownedtext", map[string]DefinitionArgument{"Value": {Kind: "Text", Text: raw}})
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
	result.text, err = e.valueBinding("Literalvalue", map[string]DefinitionArgument{"Value": atom})
	return result, err
}

func (e *orchestrationEmitter) copyValue(value emittedExpression, typ runtimeTypeRef) (string, error) {
	if value.optional {
		return e.valueBinding("RequiredError", map[string]DefinitionArgument{"Value": definitionReferenceArgument(value.text)})
	}
	if value.kind == valueText && value.arc {
		return e.valueBinding("Ownedbytes", map[string]DefinitionArgument{"Value": definitionReferenceArgument(value.text)})
	}
	word := "Copiedvalue"
	if value.borrow {
		word = "Copiedborrow"
	}
	switch value.kind {
	case valueThing:
		word = "UseSnapshotvalue"
	case valueList:
		word = "UseSnapshotlist"
	case valueMap:
		word = "UseSnapshotmap"
	case valuePointer:
		word = "UseSnapshotpointer"
	case valueFunction:
		word = "UseSnapshotfunction"
	}
	args := map[string]DefinitionArgument{"Value": definitionReferenceArgument(value.text)}
	if value.kind == valueFunction {
		typ, err := nativeFunctionPointer(typ)
		if err != nil {
			return "", err
		}
		args["Types"] = definitionTypeArgument(typ)
	}
	return e.valueBinding(word, args)
}

func (e *orchestrationEmitter) boxValue(value emittedExpression, typ runtimeTypeRef) (string, error) {
	copied, err := e.copyValue(value, typ)
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
	args := map[string]DefinitionArgument{"Value": definitionReferenceArgument(copied), "Signature": definitionReferenceArgument(signature)}
	if value.kind == valueFunction {
		pointer, err := nativeFunctionPointer(typ)
		if err != nil {
			return "", err
		}
		args["Types"] = definitionTypeArgument(pointer)
	}
	return e.valueBinding("UseBox"+strings.ToLower(value.kind.String()), args)
}

func (e *orchestrationEmitter) extractValue(value string, typ runtimeTypeRef) (string, error) {
	if typ.Kind == valueThing {
		return value, nil
	}
	args := map[string]DefinitionArgument{"Value": definitionReferenceArgument(value)}
	if typ.Kind == valueFunction {
		pointer, err := nativeFunctionPointer(typ)
		if err != nil {
			return "", err
		}
		args["Types"] = definitionTypeArgument(pointer)
	}
	return e.valueBinding("UseExtract"+strings.ToLower(typ.Kind.String()), args)
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
	return e.word("Select", map[string]DefinitionArgument{"Condition": definitionReferenceArgument(condition), "Then": {Kind: "Syllable", Definition: &body}, "Else": {Kind: "Syllable", Definition: &fallback}})
}
