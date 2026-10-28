package Lyger

import "github.com/l-programming-language/l/lerrors"

func (e *orchestrationEmitter) transferRuntimeValue(value *orchestrationValue, target runtimeTypeRef) (string, error) {
	resolved, err := e.value(value)
	if err != nil {
		return "", err
	}
	if target.Kind == valueThing {
		return e.boxValue(resolved, value.RuntimeType)
	}
	return e.copyValue(resolved, value.RuntimeType)
}

func (e *orchestrationEmitter) runtimeOperand(value *orchestrationValue, result emittedExpression) (emittedExpression, error) {
	if value.Source.Kind == operandRead {
		owner, err := e.value(value.Owner)
		if err != nil {
			return result, err
		}
		args := map[string]DefinitionArgument{}
		if value.Key != nil {
			key, err := e.value(value.Key)
			if err != nil {
				return result, err
			}
			args["Index"] = definitionReferenceArgument(key.text)
			args["Key"] = definitionReferenceArgument(key.text)
		}
		word := ""
		switch value.Owner.Type {
		case valueList:
			word = "UseReadlist"
			args["List"] = definitionReferenceArgument(owner.text)
		case valueMap:
			word = "UseReadmap"
			args["Map"] = definitionReferenceArgument(owner.text)
		case valuePointer:
			word = "UseReadpointer"
			args["Pointer"] = definitionReferenceArgument(owner.text)
		case valueBytes:
			args["Bytes"] = definitionReferenceArgument(owner.text)
			result.text, err = e.valueBinding("UseReadbyte", args)
			return result, err
		case valueThing:
			word = "UseSnapshotvalue"
			args["Value"] = definitionReferenceArgument(owner.text)
		default:
			return result, lerrors.Internal.Errorf("unsupported planned read")
		}
		boxed, err := e.valueBinding(word, args)
		if err != nil {
			return result, err
		}
		expected, err := e.typeValue(value.RuntimeType)
		if err != nil {
			return result, err
		}
		checked, err := e.valueBinding("UseExpectvalue", map[string]DefinitionArgument{"Value": definitionReferenceArgument(boxed), "Expected": definitionReferenceArgument(expected)})
		if err != nil {
			return result, err
		}
		result.text, err = e.extractValue(checked, value.RuntimeType)
		return result, err
	}
	switch value.Type {
	case valueNull:
		var err error
		result.text, err = e.valueBinding("Emptyresults", map[string]DefinitionArgument{})
		return result, err
	case valueText, valueBytes, valueError:
		child := value.Elements[0]
		var err error
		result.text, err = e.transferRuntimeValue(child, child.RuntimeType)
		if err == nil && value.Type == valueError {
			result.text, err = e.valueBinding("Constructerror", map[string]DefinitionArgument{"Value": definitionReferenceArgument(result.text)})
		}
		return result, err
	}
	var values []string
	for _, child := range value.Elements {
		v, err := e.transferRuntimeValue(child, runtimeTypeRef{Kind: valueThing})
		if err != nil {
			return result, err
		}
		values = append(values, v)
	}
	if value.Type == valueThing {
		result.text = values[0]
		return result, nil
	}
	if value.RuntimeType.Element == nil {
		return result, lerrors.Internal.Errorf("constructed collection has no element type")
	}
	element, err := e.typeValue(*value.RuntimeType.Element)
	if err != nil {
		return result, err
	}
	args := map[string]DefinitionArgument{"Element": definitionReferenceArgument(element)}
	word := ""
	switch value.Type {
	case valuePointer:
		word = "UseConstructpointer"
		args["Value"] = definitionReferenceArgument(values[0])
	case valueList:
		word = "UseConstructlist"
		items, err := e.vector(references(values), "LValue")
		if err != nil {
			return result, err
		}
		args["Values"] = definitionReferenceArgument(items)
	case valueMap:
		word = "UseConstructmap"
	default:
		return result, lerrors.Internal.Errorf("unsupported planned construction %s", value.Type)
	}
	result.text, err = e.valueBinding(word, args)
	return result, err
}

func (e *orchestrationEmitter) storeRuntimeValue(slot *orchestrationStorage, value string) ([]string, error) {
	index, shared := e.programStorage[slot]
	if shared || slot.Capture != nil {
		boxed, err := e.boxValue(emittedExpression{text: value, kind: slot.Type}, slot.RuntimeType)
		if err != nil {
			return nil, err
		}
		args := map[string]DefinitionArgument{"Value": definitionReferenceArgument(boxed)}
		word := "UseWriteprogramvalues"
		if shared {
			args["Program"] = definitionReferenceArgument("lProgram")
			args["Index"] = definitionIndexArgument(uint64(index))
		} else {
			word = "UseUpdatecapture"
			args["Environment"] = definitionReferenceArgument(slot.Capture.Name)
			args["Index"] = definitionIndexArgument(uint64(slot.Capture.Index))
		}
		_, err = e.valueBinding(word, args)
		return nil, err
	}
	word := "Assignedvalue"
	if slot.optionalStorage() {
		word = "Publishedvalue"
	}
	return e.word(word, map[string]DefinitionArgument{"Name": definitionReferenceArgument(slot.name()), "Value": definitionReferenceArgument(value)})
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
	args := map[string]DefinitionArgument{"Value": definitionReferenceArgument(boxed)}
	if target.kind == valuePointer {
		args["Pointer"] = definitionReferenceArgument(target.text)
		_, err = e.valueBinding("UseUpdatepointer", args)
		return nil, err
	}
	snapshot, err := e.copyValue(target, current.Target.RuntimeType)
	if err != nil {
		return nil, err
	}
	mutable, err := e.valueBinding("Mutablebinding", map[string]DefinitionArgument{"Value": definitionReferenceArgument(snapshot)})
	if err != nil {
		return nil, err
	}
	word := "UseAppendlist"
	args["List"] = definitionReferenceArgument(mutable)
	if current.Key != nil {
		key, err := e.transferRuntimeValue(current.Key, current.Key.RuntimeType)
		if err != nil {
			return nil, err
		}
		word = "UseUpdatelist"
		args["Index"] = definitionReferenceArgument(key)
		if target.kind == valueMap {
			word = "UseUpdatemap"
			args["Map"] = definitionReferenceArgument(mutable)
			args["Key"] = definitionReferenceArgument(key)
		}
	}
	if _, err = e.valueBinding(word, args); err != nil {
		return nil, err
	}
	return e.storeRuntimeValue(current.Target.Storage, mutable)
}

// Optional storage represents a declared-but-unassigned slot. None is never an
// L value; semantic definite assignment still governs every source read.
func (slot *orchestrationStorage) optionalStorage() bool {
	// Anonymous call arguments are initialized by the call, not declared here.
	return slot.Published || slot.Owner != nil && slot.Mutable && slot.Initial == nil && !slot.Parameter
}
