package Lyger

import (
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

func runtimeObjectKind(object *objectGram) (valueType, bool, error) {
	if object == nil || object.noun == nil {
		return valueInvalid, false, nil
	}
	initialized := runtimeInitialization(object)
	// A typed read names a type (including a bound Function signature), not
	// a constructor. Ordinary App and native Function construction stays separate.
	if initialized && runtimeReadPhrase(object) != nil {
		typ, err := sourceTypeFromObject(object)
		return typ.Kind, true, err
	}
	if object.noun.binding != nil {
		return valueInvalid, false, nil
	}
	if role, native := object.noun.nativeRole(); native && role == roleError && object.phrase != nil {
		return valueError, true, nil
	}
	name, _ := checkType(object.noun.rawBytes)
	kind, ok := parseValueType(name)
	if !ok {
		return valueInvalid, false, nil
	}
	explicit := initialized || kind == valueNull || kind == valueList || kind == valueMap
	_, runtime := runtimeTypeName(kind)
	return kind, explicit && runtime && kind != valueApp && kind != valueFunction, nil
}

func runtimeObjectEntry(object *objectGram) (*entry, string, bool, error) {
	kind, ok, err := runtimeObjectKind(object)
	if !ok || err != nil {
		return nil, "", ok, err
	}
	category, _ := kind.StorageKind()
	value := initEntry(kind.String(), category)
	value.Runtime = &runtimeReference{Expression: object}
	return value, strings.ToLower(category.String()) + "." + strings.ToLower(kind.String()), true, nil
}

func phraseObjects(phrase *phraseGram) []*objectGram {
	if phrase == nil {
		return nil
	}
	if len(phrase.objects) > 0 {
		return phrase.objects
	}
	if phrase.obj != nil {
		return []*objectGram{phrase.obj}
	}
	return nil
}

func (b *instructionBuilder) constructedValue(object *objectGram, ctx instructionContext) (*instructionValue, error) {
	kind, _, err := runtimeObjectKind(object)
	if err != nil {
		return nil, err
	}
	result := &instructionValue{Kind: operandConstruct, Type: kind, Runtime: true, Token: object.noun, RuntimeType: runtimeTypeRef{Kind: kind}}
	if read := runtimeReadPhrase(object); read != nil {
		result.Kind = operandRead
		var err error
		result.Owner, err = b.phraseValue(read, ctx)
		if err != nil {
			return nil, err
		}
		if read.phrase != nil {
			result.Key, err = b.phraseValue(read.phrase, ctx)
			if err != nil {
				return nil, err
			}
		}
		result.RuntimeType, err = sourceTypeFromObject(object)
		if err != nil {
			return nil, err
		}
		if read == object.phrase && (kind == valueList || kind == valueMap || kind == valuePointer) {
			if element := result.Owner.RuntimeType.Element; element != nil && element.Kind == kind {
				result.RuntimeType = *element
			}
		}
		result.RuntimeType, err = canonicalRuntimeType(result.RuntimeType, b.semantics.Aliases)
		return result, err
	}
	objects := phraseObjects(object.phrase)
	if len(objects) == 1 && (kind == valueMap || kind == valueList && sourceTypeObject(objects[0])) {
		element, err := sourceTypeFromObject(objects[0])
		if err != nil {
			return nil, err
		}
		element, err = canonicalRuntimeType(element, b.semantics.Aliases)
		if err != nil {
			return nil, err
		}
		result.RuntimeType.Element = &element
	} else {
		for _, child := range objects {
			value, err := b.objectValue(child, ctx)
			if err != nil {
				return nil, err
			}
			result.Elements = append(result.Elements, value)
		}
	}
	if kind == valueList || kind == valueMap || kind == valuePointer {
		if result.RuntimeType.Element == nil {
			result.RuntimeType.Element = &runtimeTypeRef{Kind: valueThing}
			if len(result.Elements) > 0 {
				element := result.Elements[0].RuntimeType
				if element.Kind == valueInvalid {
					element.Kind = result.Elements[0].Type
				}
				for _, child := range result.Elements[1:] {
					childType := child.RuntimeType
					if childType.Kind == valueInvalid {
						childType.Kind = child.Type
					}
					if !childType.equal(element) {
						element = runtimeTypeRef{Kind: valueThing}
						break
					}
				}
				result.RuntimeType.Element = &element
			}
		}
	}
	switch kind {
	case valueNull, valueMap:
		if len(result.Elements) != 0 {
			return nil, lerrors.TypeMismatch.Errorf("%s construction does not accept values", kind)
		}
	case valueList:
	default:
		if len(result.Elements) != 1 {
			return nil, lerrors.TypeMismatch.Errorf("%s construction requires one value", kind)
		}
	}
	return result, nil
}

func sourceTypeToken(tok *token) bool {
	if tok == nil {
		return false
	}
	if strings.HasPrefix(tok.tType, "type.") {
		return true
	}
	for _, name := range []string{"Error", "Function"} {
		if strings.EqualFold(string(tok.rawBytes), name) && (tok.binding == nil || tok.binding.Record.ID == name) {
			return true
		}
	}
	return false
}

// A typed read may put its source after nested type arguments, as in
// New List of Number from boxed. Initializing an element is a value expression.
func runtimeReadPhrase(object *objectGram) *phraseGram {
	if object == nil || object.phrase == nil {
		return nil
	}
	for _, prep := range object.phrase.prepositions {
		if prep.hasRole("preposition", "conjoinment.objective") {
			return object.phrase
		}
	}
	objects := phraseObjects(object.phrase)
	if len(objects) == 1 && !runtimeInitialization(objects[0]) {
		// A bound Function name denotes its signature in this type prefix.
		if _, err := sourceTypeFromObject(objects[0]); err == nil {
			return runtimeReadPhrase(objects[0])
		}
	}
	return nil
}

func sourceTypeObject(object *objectGram) bool {
	return object != nil && sourceTypeToken(object.noun) && !runtimeInitialization(object)
}

func runtimeInitialization(object *objectGram) bool {
	if object != nil {
		for _, modifier := range object.adjectives {
			if modifier.hasRole("adjective", "initialization") {
				return true
			}
		}
	}
	return false
}
