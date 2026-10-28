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
	return e.copyValue(resolved)
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
			field := "Index"
			if value.Owner.Type == valueMap {
				field = "Key"
			}
			args[field] = definitionReferenceArgument(key.text)
		}
		word := ""
		switch value.Owner.Type {
		case valueList:
			word = "Storage.List.Read"
			args["List"] = definitionReferenceArgument(owner.text)
		case valueMap:
			word = "Storage.Map.Read"
			args["Map"] = definitionReferenceArgument(owner.text)
		case valuePointer:
			word = "Storage.Pointer.Read"
			args["Pointer"] = definitionReferenceArgument(owner.text)
		case valueBytes:
			args["Bytes"] = definitionReferenceArgument(owner.text)
			if err = e.borrowArgument(args, "Bytes", false); err != nil {
				return result, err
			}
			result.text, err = e.resultBinding("Storage.Bytes.Read", args)
			return result, err
		case valueThing:
			word = "Storage.Value.Copy"
			borrowed, err := e.valueBinding("Storage.Borrow", map[string]DefinitionArgument{"Value": definitionReferenceArgument(owner.text)})
			if err != nil {
				return result, err
			}
			args["Value"] = definitionReferenceArgument(borrowed)
		default:
			return result, lerrors.Internal.Errorf("unsupported planned read")
		}
		var boxed string
		if value.Owner.Type == valueThing {
			boxed, err = e.valueBinding(word, args)
		} else {
			if err = e.borrowArgument(args, value.Owner.Type.String(), false); err != nil {
				return result, err
			}
			if value.Owner.Type == valueMap {
				if err = e.borrowArgument(args, "Key", false); err != nil {
					return result, err
				}
			}
			boxed, err = e.resultBinding(word, args)
		}
		if err != nil {
			return result, err
		}
		expected, err := e.typeValue(value.RuntimeType)
		if err != nil {
			return result, err
		}
		outcome, err := e.valueBinding("Storage.Value.Expect", map[string]DefinitionArgument{"Value": definitionReferenceArgument(boxed), "Expected": definitionReferenceArgument(expected)})
		if err != nil {
			return emittedExpression{}, err
		}
		checked, err := e.checkedBinding(outcome)
		if err != nil {
			return result, err
		}
		result.text, err = e.extractValue(checked, value.RuntimeType)
		return result, err
	}
	switch value.Type {
	case valueNull:
		var err error
		result.text, err = e.results(nil)
		return result, err
	case valueText, valueBytes, valueError:
		child := value.Elements[0]
		var err error
		result.text, err = e.transferRuntimeValue(child, child.RuntimeType)
		if err == nil && value.Type == valueError {
			result.text, err = e.valueBinding("Storage.Error.Create", map[string]DefinitionArgument{"Value": definitionReferenceArgument(result.text)})
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
		word = "Storage.Pointer.Create"
		args["Value"] = definitionReferenceArgument(values[0])
	case valueList:
		word = "Storage.List.Create"
		items, err := e.vector(references(values), "LValue")
		if err != nil {
			return result, err
		}
		args["Values"] = definitionReferenceArgument(items)
	case valueMap:
		word = "Storage.Map.Create"
	default:
		return result, lerrors.Internal.Errorf("unsupported planned construction %s", value.Type)
	}
	if value.Type == valueMap {
		result.text, err = e.valueBinding(word, args)
	} else {
		result.text, err = e.resultBinding(word, args)
	}
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
		word := "Storage.Program.Values.Write"
		if shared {
			args["Program"] = definitionReferenceArgument("lProgram")
			args["Index"] = definitionIndexArgument(uint64(index))
		} else {
			word = "Storage.Environment.Update"
			args["Environment"] = definitionReferenceArgument(slot.Capture.Name)
			args["Index"] = definitionIndexArgument(uint64(slot.Capture.Index))
		}
		key := "Environment"
		if shared {
			key = "Program"
		}
		_, err = e.borrowedResultBinding(word, args, key)
		return nil, err
	}
	word := "Storage.Assign"
	if slot.optionalStorage() {
		word = "Storage.Publish"
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
		if err = e.borrowArgument(args, "Pointer", false); err != nil {
			return nil, err
		}
		_, err = e.resultBinding("Storage.Pointer.Update", args)
		return nil, err
	}
	snapshot, err := e.copyValue(target)
	if err != nil {
		return nil, err
	}
	mutable, err := e.describedValueBinding("Storage.Binding", []string{"Storage.Mutable"}, map[string]DefinitionArgument{"Value": definitionReferenceArgument(snapshot)})
	if err != nil {
		return nil, err
	}
	word := "Storage.List.Append"
	args["List"] = definitionReferenceArgument(mutable)
	if current.Key != nil {
		key, err := e.transferRuntimeValue(current.Key, current.Key.RuntimeType)
		if err != nil {
			return nil, err
		}
		if target.kind == valueMap {
			word = "Storage.Map.Update"
			delete(args, "List")
			args["Map"] = definitionReferenceArgument(mutable)
			args["Key"] = definitionReferenceArgument(key)
		} else {
			word = "Storage.List.Update"
			args["Index"] = definitionReferenceArgument(key)
		}
	}
	if err = e.borrowArgument(args, target.kind.String(), true); err != nil {
		return nil, err
	}
	if _, err = e.resultBinding(word, args); err != nil {
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
