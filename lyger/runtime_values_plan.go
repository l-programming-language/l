package Lyger

import "github.com/l-programming-language/l/lerrors"

func (p *orchestrationPlanner) runtimeOperand(source *instructionValue, flow orchestrationFlow) (*orchestrationValue, error) {
	p.plan.RuntimeValues = true
	result := &orchestrationValue{Source: source, Type: source.Type, RuntimeType: source.RuntimeType, Access: orchestrationAccessOwned}
	p.registerRuntimeType(source.RuntimeType)
	for _, field := range source.Fields {
		p.registerRuntimeType(field.Type)
	}
	for _, element := range source.Elements {
		value, err := p.value(element, flow)
		if err != nil {
			return nil, err
		}
		result.Elements = append(result.Elements, value)
	}
	if source.Kind == operandRead {
		var err error
		result.Owner, err = p.value(source.Owner, flow)
		if err != nil {
			return nil, err
		}
		if source.Key != nil {
			result.Key, err = p.value(source.Key, flow)
			if err != nil {
				return nil, err
			}
		}
		keyType := valueInvalid
		switch result.Owner.Type {
		case valueList, valueBytes:
			keyType = valueNumber
		case valueMap, valueInstance:
			keyType = valueText
		case valuePointer, valueThing:
		default:
			return nil, lerrors.TypeMismatch.Errorf("%s does not support indexed reads", result.Owner.Type)
		}
		if (keyType == valueInvalid) != (result.Key == nil) || result.Key != nil && result.Key.Type != keyType {
			return nil, lerrors.TypeMismatch.Errorf("invalid %s read key", result.Owner.Type)
		}
		if element := result.Owner.RuntimeType.Element; element != nil && element.Kind != valueThing && !result.RuntimeType.accepts(*element) {
			return nil, lerrors.TypeMismatch.Errorf("read type does not match %s element", result.Owner.Type)
		}
		return result, nil
	}
	switch result.Type {
	case valueList:
		for _, element := range result.Elements {
			if !result.RuntimeType.Element.accepts(element.RuntimeType) {
				return nil, lerrors.TypeMismatch.Errorf("List element type mismatch")
			}
		}
	case valueMap, valueInstance, valueNull:
	case valuePointer, valueThing:
	case valueBytes, valueText, valueError:
		if result.Elements[0].Type != valueText && result.Elements[0].Type != valueBytes {
			return nil, lerrors.TypeMismatch.Errorf("%s construction requires Text or Bytes", result.Type)
		}
	default:
		return nil, lerrors.Unsupported.Errorf("unsupported runtime construction %s", result.Type)
	}
	return result, nil
}

func (p *orchestrationPlanner) runtimeUpdate(source *instruction, flow orchestrationFlow) (*orchestrationInstruction, error) {
	result := &orchestrationInstruction{Source: source}
	var err error
	result.Target, err = p.value(source.Target, flow)
	if err != nil {
		return nil, err
	}
	result.Value, err = p.value(source.Value, flow)
	if err != nil {
		return nil, err
	}
	if source.Key != nil {
		result.Key, err = p.value(source.Key, flow)
		if err != nil {
			return nil, err
		}
	}
	target := result.Target
	switch target.Type {
	case valueList:
		if result.Key != nil && result.Key.Type != valueNumber {
			return nil, lerrors.TypeMismatch.Errorf("List index requires Number")
		}
	case valueMap, valueInstance:
		if result.Key == nil || result.Key.Type != valueText {
			return nil, lerrors.TypeMismatch.Errorf("field key requires Text")
		}
	case valuePointer:
		if result.Key != nil {
			return nil, lerrors.TypeMismatch.Errorf("Pointer update has no key")
		}
	default:
		return nil, lerrors.TypeMismatch.Errorf("%s cannot be updated", target.Type)
	}
	if target.RuntimeType.Element != nil && !target.RuntimeType.Element.accepts(result.Value.RuntimeType) {
		return nil, lerrors.TypeMismatch.Errorf("%s update has incompatible value type", target.Type)
	}
	if target.Type != valuePointer {
		if target.Storage == nil {
			return nil, lerrors.TypeMismatch.Errorf("update requires a stored value")
		}
		target.Storage.Mutable = true
	}
	result.Effects = orchestrationEffectAssign | orchestrationValueEffects(result.Target) | orchestrationValueEffects(result.Value) | orchestrationValueEffects(result.Key)
	return result, nil
}

func (p *orchestrationPlanner) registerRuntimeType(typ runtimeTypeRef) {
	if p.plan.Identities == nil {
		p.plan.Identities = make(map[*record]uint64)
	}
	for _, identity := range []*record{typ.Identity, typ.Definition} {
		if identity != nil && p.plan.Identities[identity] == 0 {
			p.plan.Identities[identity] = uint64(len(p.plan.Identities) + 1)
		}
	}
	if typ.Element != nil {
		p.registerRuntimeType(*typ.Element)
	}
	for _, types := range [][]runtimeTypeRef{typ.Inputs, typ.Outputs} {
		for _, child := range types {
			p.registerRuntimeType(child)
		}
	}
}
