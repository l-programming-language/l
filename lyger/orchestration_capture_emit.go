package Lyger

import (
	"fmt"
	"github.com/l-programming-language/l/lerrors"
	"strings"
)

func (e *orchestrationEmitter) captureReference(source *orchestrationCallableReference) (string, error) {
	if source == nil {
		return "", lerrors.Internal.Errorf("missing captured callable")
	}
	if source.Capture != nil {
		return fmt.Sprintf("%s.get(%d)?.into_environment()?", source.Capture.Name, source.Capture.Index), nil
	}
	if source.Slot == nil {
		return "", lerrors.Internal.Errorf("captured callable has no slot")
	}
	if index, exists := e.programActions[source.Slot.Record]; exists {
		return fmt.Sprintf("lProgram.action(%d)?", index), nil
	}
	return source.Slot.name() + ".as_ref().ok_or_else(|| runtime::Error::new(" + quoteOrchestrationString("action "+source.Slot.Record.ID+" is not initialized") + "))?.clone()", nil
}
func (e *orchestrationEmitter) bindCapture(current *orchestrationInstruction) ([]string, error) {
	var value string
	if current.Source.Kind == instructionAliasAction {
		var err error
		value, err = e.captureReference(current.CaptureSource)
		if err != nil {
			return nil, err
		}
	} else {
		types, fields, err := e.captureEnvironment(current.CaptureValues, current.CaptureHandles)
		if err != nil {
			return nil, err
		}
		value = "LEnvironment::typed(" + types + ", " + fields + ")?"
	}
	if index, exists := e.programActions[current.Callable.Record]; exists {
		return []string{orchestrationStatement(fmt.Sprintf("lProgram.bind_action(%d, %s)?;", index, value))}, nil
	}
	return []string{orchestrationAssignment(current.Callable.name(), "Some("+value+")")}, nil
}

// Capture order is shared by ordinary actions and native Function values.
func (e *orchestrationEmitter) captureEnvironment(values []*orchestrationValue, handles []*orchestrationCallableReference) (string, string, error) {
	var fields, types []string
	for _, capture := range values {
		value, err := e.value(capture)
		if err != nil {
			return "", "", err
		}
		field, err := boxRuntimeValue(value)
		if err != nil {
			return "", "", err
		}
		fields = append(fields, field)
		typeName, err := runtimeTypeExpression(capture.RuntimeType, e.plan.Identities)
		if err != nil {
			return "", "", err
		}
		types = append(types, typeName)
	}
	for _, source := range handles {
		handle, err := e.captureReference(source)
		if err != nil {
			return "", "", err
		}
		fields = append(fields, "LValue::Environment("+handle+")")
		types = append(types, "LType::Thing")
	}
	return "vec![" + strings.Join(types, ",") + "]", "vec![" + strings.Join(fields, ",") + "]", nil
}
