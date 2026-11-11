package Lyger

import (
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// runtimeTypeRef refines the existing representation tag; source identity is
// canonical declaration identity, never a displayed name or Rust source text.
type runtimeTypeRef struct {
	Kind            valueType
	Element         *runtimeTypeRef
	Definition      *record
	Inputs, Outputs []runtimeTypeRef
}

func (a runtimeTypeRef) equal(b runtimeTypeRef) bool {
	if a.Kind != b.Kind || a.Definition != b.Definition || (a.Element == nil) != (b.Element == nil) || len(a.Inputs) != len(b.Inputs) || len(a.Outputs) != len(b.Outputs) {
		return false
	}
	if a.Element != nil && !a.Element.equal(*b.Element) {
		return false
	}
	for i := range a.Inputs {
		if !a.Inputs[i].equal(b.Inputs[i]) {
			return false
		}
	}
	for i := range a.Outputs {
		if !a.Outputs[i].equal(b.Outputs[i]) {
			return false
		}
	}
	return true
}

func canonicalRuntimeType(ref runtimeTypeRef, aliases map[*record]*record) (runtimeTypeRef, error) {
	var err error
	if ref.Definition != nil {
		ref.Definition, err = canonicalRecord(ref.Definition, aliases)
		if err != nil {
			return ref, err
		}
	}
	if ref.Element != nil {
		element, err := canonicalRuntimeType(*ref.Element, aliases)
		if err != nil {
			return ref, err
		}
		ref.Element = &element
	}
	for _, types := range []*[]runtimeTypeRef{&ref.Inputs, &ref.Outputs} {
		copy := append([]runtimeTypeRef(nil), (*types)...)
		for i := range copy {
			copy[i], err = canonicalRuntimeType(copy[i], aliases)
			if err != nil {
				return ref, err
			}
		}
		*types = copy
	}
	return ref, nil
}

func (a runtimeTypeRef) accepts(b runtimeTypeRef) bool {
	return a.Kind == valueThing || a.equal(b)
}

// These are semantic value categories. Their analysis does not prescribe a
// native representation or require a runtime implementation.
func runtimeValueKind(kind valueType) bool {
	switch kind {
	case valueText, valueNumber, valueTruth, valueBytes, valueNull, valueThing,
		valueList, valueMap, valuePointer, valueError, valueApp, valueFunction:
		return true
	default:
		return false
	}
}

func recordRuntimeType(rec *record) (runtimeTypeRef, error) {
	kind, err := instructionType(rec)
	if err != nil {
		return runtimeTypeRef{}, err
	}
	if rec.runtimeType != nil {
		if rec.runtimeType.Kind != kind {
			return runtimeTypeRef{}, lerrors.Internal.Errorf("record %s has inconsistent runtime type", rec.ID)
		}
		return *rec.runtimeType, nil
	}
	result := runtimeTypeRef{Kind: kind}
	if kind == valueList || kind == valueMap || kind == valuePointer {
		result.Element = &runtimeTypeRef{Kind: valueThing}
	}
	return result, nil
}

func sourceTypeFromObject(object *objectGram) (runtimeTypeRef, error) {
	if object == nil || object.noun == nil {
		return runtimeTypeRef{}, lerrors.TypeMismatch.Errorf("missing runtime type")
	}
	if object.noun.binding != nil {
		rec := object.noun.binding.Record
		if rec != nil && rec.Value != nil && rec.Value.Applicative != nil && rec.Value.Applicative.Function != nil && rec.Value.Applicative.Function.Source != nil {
			return rec.Value.Applicative.Function.Source.signature(), nil
		}
		if rec != nil && rec.runtimeType != nil && rec.runtimeType.Kind == valueFunction {
			return *rec.runtimeType, nil
		}
	}
	name, _ := checkType(object.noun.rawBytes)
	kind, ok := parseValueType(name)
	if !ok || kind == valueUnknown {
		ok = false
		for k, spelling := range valueTypeNames {
			if strings.EqualFold(spelling, string(object.noun.rawBytes)) {
				kind, ok = valueType(k), true
				break
			}
		}
	}
	if !ok {
		return runtimeTypeRef{}, lerrors.TypeMismatch.Errorf("unknown runtime type %s", object.noun.rawBytes)
	}
	if !runtimeValueKind(kind) {
		return runtimeTypeRef{}, lerrors.TypeMismatch.Errorf("%s is a source structure, not a runtime type", name)
	}
	result := runtimeTypeRef{Kind: kind}
	if kind == valueFunction && object.noun.binding != nil {
		result.Definition = object.noun.binding.Record
	}
	if kind == valueList || kind == valueMap || kind == valuePointer {
		result.Element = &runtimeTypeRef{Kind: valueThing}
		if object.phrase != nil && object.phrase.obj != nil && runtimeReadPhrase(object) != object.phrase {
			element, err := sourceTypeFromObject(object.phrase.obj)
			if err != nil {
				return result, err
			}
			result.Element = &element
		}
	}
	return result, nil
}
