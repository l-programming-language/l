package Lyger

import "github.com/l-programming-language/l/lerrors"

func (e *orchestrationEmitter) captureReference(source *orchestrationCallableReference) (string, error) {
	if source == nil {
		return "", lerrors.Internal.Errorf("missing captured callable")
	}
	if capture := source.Capture; capture != nil {
		value, err := e.borrowedResultBinding("Storage.Environment.Read", map[string]DefinitionArgument{
			"Environment": definitionReferenceArgument(capture.Name), "Index": definitionIndexArgument(uint64(capture.Index)),
		}, "Environment")
		if err != nil {
			return "", err
		}
		return e.restoreValue(value, "LEnvironment", "expected Environment", "Storage.Environment.Copy")
	}
	if source.Slot == nil {
		return "", lerrors.Internal.Errorf("captured callable has no slot")
	}
	if index, exists := e.programActions[source.Slot.Record]; exists {
		return e.borrowedResultBinding("Storage.Program.Actions.Read", map[string]DefinitionArgument{
			"Program": definitionReferenceArgument("lProgram"), "Index": definitionIndexArgument(uint64(index)),
		}, "Program")
	}
	return e.borrowedResultBinding("Storage.Environment.Require", map[string]DefinitionArgument{
		"Value":   definitionReferenceArgument(source.Slot.name()),
		"Message": {Kind: "Text", Text: "action " + source.Slot.Record.ID + " is not initialized"},
	}, "Value")
}

func (e *orchestrationEmitter) bindCapture(current *orchestrationInstruction) ([]string, error) {
	var value string
	var err error
	if current.Source.Kind == instructionAliasAction {
		value, err = e.captureReference(current.CaptureSource)
	} else {
		var types, fields string
		types, fields, err = e.captureEnvironment(current.CaptureValues, current.CaptureHandles)
		if err == nil {
			value, err = e.resultBinding("Storage.Environment.Create", map[string]DefinitionArgument{
				"Types": definitionReferenceArgument(types), "Values": definitionReferenceArgument(fields),
			})
		}
	}
	if err != nil {
		return nil, err
	}
	if index, exists := e.programActions[current.Callable.Record]; exists {
		_, err = e.borrowedResultBinding("Storage.Program.Actions.Write", map[string]DefinitionArgument{
			"Program": definitionReferenceArgument("lProgram"), "Index": definitionIndexArgument(uint64(index)),
			"Value": definitionReferenceArgument(value),
		}, "Program")
		return nil, err
	}
	return e.word("Storage.Publish", map[string]DefinitionArgument{
		"Name": definitionReferenceArgument(current.Callable.name()), "Value": definitionReferenceArgument(value),
	})
}

// Capture order is shared by ordinary actions and native Function values. Each
// binding remains a value reference; the source words own copies and boxing.
func (e *orchestrationEmitter) captureEnvironment(values []*orchestrationValue, handles []*orchestrationCallableReference) (string, string, error) {
	var fields, types []DefinitionArgument
	for _, capture := range values {
		value, err := e.value(capture)
		if err != nil {
			return "", "", err
		}
		field, err := e.boxValue(value, capture.RuntimeType)
		if err != nil {
			return "", "", err
		}
		typeName, err := e.typeValue(capture.RuntimeType)
		if err != nil {
			return "", "", err
		}
		fields = append(fields, definitionReferenceArgument(field))
		types = append(types, definitionReferenceArgument(typeName))
	}
	for _, source := range handles {
		handle, err := e.captureReference(source)
		if err != nil {
			return "", "", err
		}
		typeName, err := e.typeValue(runtimeTypeRef{Kind: valueThing})
		if err != nil {
			return "", "", err
		}
		// Boxing owns its signature; the schema needs its own vector value.
		signature, err := e.copyAtom(typeName, false)
		if err != nil {
			return "", "", err
		}
		field, err := e.wrapValue(handle, signature, "Storage.Environment.Render", "Storage.Environment.Visit", "")
		if err != nil {
			return "", "", err
		}
		fields = append(fields, definitionReferenceArgument(field))
		types = append(types, definitionReferenceArgument(typeName))
	}
	typeVector, err := e.vector(types, "Vec<u64>")
	if err != nil {
		return "", "", err
	}
	fieldVector, err := e.vector(fields, "LValue")
	return typeVector, fieldVector, err
}

// Functions borrow root state for published declarations and continuations.
// Explicit Function captures remain independent snapshots of these values.
func (e *orchestrationEmitter) programContext(actions []*record) ([]string, error) {
	start := len(e.pending)
	e.programStorage = make(map[*orchestrationStorage]int)
	e.programActions = make(map[*record]int)
	var types, apps []DefinitionArgument
	if e.plan.FunctionProgramState {
		for index, slot := range e.plan.Root.Storage {
			e.programStorage[slot] = index
			typ, err := e.typeValue(slot.RuntimeType)
			if err != nil {
				return nil, err
			}
			types = append(types, definitionReferenceArgument(typ))
		}
		for index, rec := range actions {
			e.programActions[rec] = index
		}
	}
	for _, app := range e.plan.Apps {
		retained, err := e.valueBinding("Retain", map[string]DefinitionArgument{"Target": definitionReferenceArgument(e.apps[app])})
		if err != nil {
			return nil, err
		}
		apps = append(apps, definitionReferenceArgument(retained))
	}
	appVector, err := e.vector(apps, "std::sync::Arc<runtime::App>")
	if err != nil {
		return nil, err
	}
	typeVector, err := e.vector(types, "Vec<u64>")
	if err != nil {
		return nil, err
	}
	program, err := e.valueBinding("Storage.Program.Create", map[string]DefinitionArgument{
		"Apps": definitionReferenceArgument(appVector), "Types": definitionReferenceArgument(typeVector),
		"Action_count": definitionIndexArgument(uint64(len(e.programActions))),
	})
	if err != nil {
		return nil, err
	}
	context, err := e.word("Storage.Context.Create", map[string]DefinitionArgument{
		"Name": definitionReferenceArgument("lContext"), "Runtime": definitionReferenceArgument("runtime"),
		"Program": definitionReferenceArgument(program), "Binding": definitionReferenceArgument("lProgram"),
	})
	if err != nil {
		return nil, err
	}
	result := append([]string(nil), e.pending[start:]...)
	e.pending = e.pending[:start]
	return append(result, context...), nil
}
