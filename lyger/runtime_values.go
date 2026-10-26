package Lyger

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// runtimeTypeRef refines the existing representation tag; source identity is
// canonical declaration identity, never a displayed name or Rust source text.
type runtimeField struct {
	Name string
	Type runtimeTypeRef
}

type runtimeTypeRef struct {
	Kind            valueType
	Element         *runtimeTypeRef
	Identity        *record
	Definition      *record
	Inputs, Outputs []runtimeTypeRef
}

func (a runtimeTypeRef) equal(b runtimeTypeRef) bool {
	if a.Kind != b.Kind || a.Identity != b.Identity || a.Definition != b.Definition || (a.Element == nil) != (b.Element == nil) || len(a.Inputs) != len(b.Inputs) || len(a.Outputs) != len(b.Outputs) {
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
	if ref.Identity != nil {
		ref.Identity, err = canonicalRecord(ref.Identity, aliases)
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

var runtimeRepresentations = map[valueType][2]string{
	valueText: {"Vec<u8>", "Text"}, valueNumber: {"f64", "Number"}, valueTruth: {"bool", "Truth"},
	valueBytes: {"Vec<u8>", "Bytes"}, valueNull: {"()", "Null"}, valueThing: {"LValue", ""},
	valueList: {"LList", "List"}, valueMap: {"LMap", "Map"}, valueInstance: {"LInstance", "Instance"},
	valuePointer: {"LPointer", "Pointer"}, valueError: {"runtime::Error", "Error"},
	valueApp: {"std::sync::Arc<runtime::App>", "App"}, valueFunction: {"", "Function"},
}

func runtimeTypeName(kind valueType) (string, bool) {
	entry, ok := runtimeRepresentations[kind]
	return entry[0], ok
}

func runtimeCopyExpression(value emittedExpression, access orchestrationAccess) (string, error) {
	if _, ok := runtimeTypeName(value.kind); !ok {
		return "", lerrors.Unsupported.Errorf("no runtime representation for %s", value.kind)
	}
	if value.kind == valueError && value.optional {
		return "(" + value.text + ").ok_or_else(|| runtime::Error::new(\"App has no Error\"))?", nil
	}
	if value.kind == valueText && (value.arc || access == orchestrationAccessSnapshot) {
		return "(" + value.text + ").to_vec()", nil
	}
	if access == orchestrationAccessBorrow || access == orchestrationAccessSnapshot || value.borrow {
		switch value.kind {
		case valueNumber, valueTruth, valueNull:
			if value.borrow {
				return "*(" + value.text + ")", nil
			}
		default:
			return "(" + value.text + ").clone()", nil
		}
	}
	return value.text, nil
}

func boxRuntimeValue(value emittedExpression) (string, error) {
	text, err := runtimeCopyExpression(value, orchestrationAccessSnapshot)
	if err != nil {
		return "", err
	}
	if value.kind == valueThing {
		return text, nil
	}
	if value.kind == valueFunction {
		if value.typeExpression == "" {
			return "", lerrors.Internal.Errorf("Function storage requires its planned type")
		}
		return "LValue::function(" + text + ", " + value.typeExpression + ")", nil
	}
	if value.kind == valueNull {
		if value.literal {
			return "LValue::Null", nil
		}
		return "{ let () = " + text + "; LValue::Null }", nil
	}
	return "LValue::" + runtimeRepresentations[value.kind][1] + "(" + text + ")", nil
}

func unboxRuntimeValue(expression string, kind valueType) (string, error) {
	if _, ok := runtimeTypeName(kind); !ok {
		return "", lerrors.Unsupported.Errorf("no runtime representation for %s", kind)
	}
	if kind == valueThing {
		return expression, nil
	}
	return "(" + expression + ").into_" + strings.ToLower(runtimeRepresentations[kind][1]) + "()?", nil
}

func unboxTypedRuntimeValue(expression string, typ runtimeTypeRef) (string, error) {
	if typ.Kind != valueFunction {
		return unboxRuntimeValue(expression, typ.Kind)
	}
	native, err := nativeRuntimeType(typ)
	if err != nil {
		return "", err
	}
	return "(" + expression + ").into_function::<" + native + ">()?", nil
}

func runtimeTypeExpression(t runtimeTypeRef, identities map[*record]uint64) (string, error) {
	variant := runtimeRepresentations[t.Kind][1]
	if t.Kind == valueThing {
		variant = "Thing"
	}
	if variant == "" {
		return "", lerrors.Unsupported.Errorf("no runtime type for %s", t.Kind)
	}
	switch t.Kind {
	case valueList, valueMap, valuePointer:
		if t.Element == nil {
			return "", lerrors.TypeMismatch.Errorf("%s requires an element type", t.Kind)
		}
		elem, err := runtimeTypeExpression(*t.Element, identities)
		if err != nil {
			return "", err
		}
		return "LType::" + variant + "(Box::new(" + elem + "))", nil
	case valueInstance:
		id, ok := identities[t.Identity]
		if !ok {
			return "", lerrors.Internal.Errorf("Instance shape has no canonical type identity")
		}
		return fmt.Sprintf("LType::Instance(%d)", id), nil
	case valueFunction:
		id, ok := identities[t.Definition]
		if !ok {
			return "", lerrors.Internal.Errorf("Function has no canonical definition identity")
		}
		parts := make([]string, 2)
		for i, types := range [][]runtimeTypeRef{t.Inputs, t.Outputs} {
			values := make([]string, len(types))
			for n, t := range types {
				var err error
				values[n], err = runtimeTypeExpression(t, identities)
				if err != nil {
					return "", err
				}
			}
			parts[i] = "vec![" + strings.Join(values, ",") + "]"
		}
		return fmt.Sprintf("LType::Function(%d,%s)", id, strings.Join(parts, ",")), nil
	}
	return "LType::" + variant, nil
}

//go:embed runtime_values.rs
var runtimeValueSupport string

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
	if _, ok = runtimeTypeName(kind); !ok {
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
	if kind == valueInstance {
		for _, modifier := range valueModifiers(object.adjectives) {
			if modifier.binding != nil {
				if result.Identity != nil {
					return result, lerrors.TypeMismatch.Errorf("Instance needs exactly one shape descriptor")
				}
				result.Identity = modifier.binding.Record
			}
		}
		if result.Identity == nil {
			return result, lerrors.TypeMismatch.Errorf("Instance requires a shape descriptor")
		}
	}
	return result, nil
}
