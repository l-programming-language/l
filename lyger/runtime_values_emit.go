package Lyger

import (
	"fmt"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

func (e *orchestrationEmitter) transferRuntimeValue(value *orchestrationValue, target runtimeTypeRef) (string, error) {
	expression, err := e.value(value)
	if err != nil {
		return "", err
	}
	if target.Kind == valueThing {
		return boxRuntimeValue(expression)
	}
	return runtimeCopyExpression(expression, orchestrationAccessSnapshot)
}

func (e *orchestrationEmitter) runtimeOperand(value *orchestrationValue, result emittedExpression) (emittedExpression, error) {
	if value.Source.Kind == operandRead {
		owner, err := e.value(value.Owner)
		if err != nil {
			return result, err
		}
		key := ""
		if value.Key != nil {
			v, err := e.value(value.Key)
			if err != nil {
				return result, err
			}
			key = v.text
		}
		switch value.Owner.Type {
		case valueList:
			result.text = owner.text + ".get(" + key + ")?"
		case valueBytes:
			result.text = "LValue::Number(*(" + owner.text + ").get(l_index(" + key + ")?).ok_or_else(|| runtime::Error::new(\"byte index out of bounds\"))? as f64)"
		case valueMap, valueInstance:
			result.text = owner.text + ".get(&" + key + ")?"
		case valuePointer:
			result.text = owner.text + ".get()?"
		case valueThing:
			result.text, err = runtimeCopyExpression(owner, orchestrationAccessSnapshot)
			if err != nil {
				return result, err
			}
		default:
			return result, lerrors.Internal.Errorf("unsupported planned read")
		}
		expected, err := runtimeTypeExpression(value.RuntimeType, e.plan.Identities)
		if err != nil {
			return result, err
		}
		result.text = "(" + result.text + ").expect_type(" + expected + ")?"
		result.text, err = unboxTypedRuntimeValue(result.text, value.RuntimeType)
		return result, err
	}
	elements := make([]string, len(value.Elements))
	for i, child := range value.Elements {
		var err error
		elements[i], err = e.transferRuntimeValue(child, runtimeTypeRef{Kind: valueThing})
		if err != nil {
			return result, err
		}
	}
	switch value.Type {
	case valueNull:
		result.text = "()"
	case valueThing:
		result.text = elements[0]
	case valuePointer:
		element, err := runtimeTypeExpression(*value.RuntimeType.Element, e.plan.Identities)
		if err != nil {
			return result, err
		}
		result.text = "LPointer::typed(" + element + "," + elements[0] + ")?"
	case valueBytes, valueText, valueError:
		child := value.Elements[0]
		var err error
		result.text, err = e.transferRuntimeValue(child, child.RuntimeType)
		if value.Type == valueError {
			result.text = "runtime::Error::from_bytes(" + result.text + ")"
		}
		if err != nil {
			return result, err
		}
	case valueList, valueMap:
		element, err := runtimeTypeExpression(*value.RuntimeType.Element, e.plan.Identities)
		if err != nil {
			return result, err
		}
		if value.Type == valueList {
			result.text = "LList::new(" + element + ", vec![" + strings.Join(elements, ",") + "])?"
		} else {
			result.text = "LMap::new(" + element + ")"
		}
	case valueInstance:
		id, ok := e.plan.Identities[value.RuntimeType.Identity]
		if !ok {
			return result, lerrors.Internal.Errorf("Instance identity was not planned")
		}
		fields := make([]string, len(value.Source.Fields))
		for i, field := range value.Source.Fields {
			typ, err := runtimeTypeExpression(field.Type, e.plan.Identities)
			if err != nil {
				return result, err
			}
			fields[i] = "(" + nativeTextBytes(field.Name) + "," + typ + ")"
		}
		result.text = fmt.Sprintf("LInstance::new(%d, vec![%s])", id, strings.Join(fields, ","))
	default:
		return result, lerrors.Internal.Errorf("unsupported planned construction %s", value.Type)
	}
	return result, nil
}

func (e *orchestrationEmitter) storeRuntimeValue(slot *orchestrationStorage, expression string) ([]string, error) {
	index, shared := e.programStorage[slot]
	capture := slot.Capture
	if shared || capture != nil {
		typ, err := runtimeTypeExpression(slot.RuntimeType, e.plan.Identities)
		if err != nil {
			return nil, err
		}
		boxed, err := boxRuntimeValue(emittedExpression{text: expression, kind: slot.Type, typeExpression: typ})
		if err != nil {
			return nil, err
		}
		if shared {
			return []string{orchestrationStatement(fmt.Sprintf("lProgram.set(%d, %s)?;", index, boxed))}, nil
		}
		return []string{orchestrationStatement(fmt.Sprintf("{ let lCaptureValue = %s; %s.set(%d, lCaptureValue)?; }", boxed, capture.Name, capture.Index))}, nil
	}

	if slot.Published {
		expression = "Some(" + expression + ")"
	}
	return []string{orchestrationAssignment(slot.name(), expression)}, nil
}

func (e *orchestrationEmitter) runtimeUpdate(current *orchestrationInstruction) ([]string, error) {
	target, err := e.value(current.Target)
	if err != nil {
		return nil, err
	}
	boxed, err := e.transferRuntimeValue(current.Value, runtimeTypeRef{Kind: valueThing})
	if err != nil {
		return nil, err
	}
	if target.kind == valuePointer {
		return []string{orchestrationStatement("{ let lUpdateValue = " + boxed + "; " + target.text + ".set(lUpdateValue)?; }")}, nil
	}
	snapshot, err := runtimeCopyExpression(target, orchestrationAccessSnapshot)
	if err != nil {
		return nil, err
	}
	key := ""
	if current.Key != nil {
		key, err = e.transferRuntimeValue(current.Key, current.Key.RuntimeType)
		if err != nil {
			return nil, err
		}
	}
	call := "lUpdated.put(lUpdateValue)?;"
	if current.Key != nil {
		call = "lUpdated.set(" + key + ", lUpdateValue)?;"
	}
	result := []string{orchestrationStatement("{"), orchestrationStatement("let lUpdateValue = " + boxed + ";"), orchestrationStatement("let mut lUpdated = " + snapshot + ";"), orchestrationStatement(call)}
	write, err := e.storeRuntimeValue(current.Target.Storage, "lUpdated")
	if err != nil {
		return nil, err
	}
	result = append(result, write...)
	return append(result, orchestrationStatement("}")), nil
}
