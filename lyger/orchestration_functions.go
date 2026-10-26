package Lyger

import (
	"fmt"
	"github.com/l-programming-language/l/lerrors"
	"strings"
)

type orchestrationFunction struct {
	Source         *sourceFunction
	Body           *orchestrationBlock
	Inputs         []*orchestrationStorage
	Self           *orchestrationStorage
	Captures       []*orchestrationValue
	CaptureHandles []*orchestrationCallableReference
}

func (p *orchestrationPlanner) functionInstruction(current *instruction, flow orchestrationFlow) (*orchestrationInstruction, orchestrationFlow, error) {
	p.plan.RuntimeValues = true
	result := &orchestrationInstruction{Source: current, Effects: orchestrationEffectAssign, ReachesNext: flow.live}
	if current.Kind == instructionReturn {
		for _, source := range current.Arguments {
			value, err := p.value(source, flow)
			if err != nil {
				return nil, flow, err
			}
			result.Arguments = append(result.Arguments, value)
		}
		flow.live = false
		result.ReachesNext = false
		return result, flow, nil
	}
	if current.Kind == instructionInvoke {
		var err error
		result.Target, err = p.value(current.Target, flow)
		if err != nil {
			return nil, flow, err
		}
		for _, source := range current.Arguments {
			value, err := p.value(source, flow)
			if err != nil {
				return nil, flow, err
			}
			result.Arguments = append(result.Arguments, value)
		}
		for _, source := range current.Results {
			rec, kind, err := p.record(source)
			if err != nil {
				return nil, flow, err
			}
			slot, _, err := p.storage(rec, kind)
			if err != nil {
				return nil, flow, err
			}
			result.Results = append(result.Results, &orchestrationValue{Source: source, Record: rec, Type: kind, RuntimeType: slot.RuntimeType, Storage: slot})
			flow.assigned[slot], flow.written[slot] = true, true
		}
		return result, flow, nil
	}
	source := current.Function
	function := &orchestrationFunction{Source: source, Body: p.newBlock(source.Scope)}
	result.Function = function
	for _, capture := range source.Captures {
		value, err := p.value(capture, flow)
		if err != nil {
			return nil, flow, err
		}
		value.Access = orchestrationAccessSnapshot
		function.Captures = append(function.Captures, value)
	}
	for _, binding := range source.CaptureCalls {
		handle, err := p.captureReference(binding, flow)
		if err != nil {
			return nil, flow, err
		}
		function.CaptureHandles = append(function.CaptureHandles, handle)
	}
	slot, _, err := p.storage(source.Record, valueFunction)
	if err != nil {
		return nil, flow, err
	}
	result.Target = &orchestrationValue{Source: current.Target, Record: source.Record, Type: valueFunction, RuntimeType: source.signature(), Storage: slot}
	flow.assigned[slot], flow.written[slot] = true, true
	p.registerRuntimeType(source.signature())
	frame := &orchestrationFrame{block: function.Body, slots: make(map[*record]*orchestrationStorage)}
	p.frames = append(p.frames, frame)
	context := orchestrationCaptureContext{values: make(map[*record]*orchestrationStorage), calls: make(map[*record]*orchestrationCallableReference)}
	for index, value := range function.Captures {
		context.values[value.Record] = &orchestrationStorage{Record: value.Record, Type: value.Type, RuntimeType: value.RuntimeType, Mutable: true, Capture: &orchestrationCaptureAccess{Name: "lSelf.captures", Index: index}}
	}
	for index, binding := range source.CaptureCalls {
		context.calls[binding] = &orchestrationCallableReference{Layout: function.CaptureHandles[index].Layout, Capture: &orchestrationCaptureAccess{Name: "lSelf.captures", Index: len(function.Captures) + index}}
	}
	p.captureContexts = append(p.captureContexts, context)
	defer func() {
		p.frames = p.frames[:len(p.frames)-1]
		p.captureContexts = p.captureContexts[:len(p.captureContexts)-1]
	}()
	inner := flow.clone()
	p.functionDepth++
	defer func() { p.functionDepth-- }()
	allocate := func(rec *record, typ runtimeTypeRef) *orchestrationStorage {
		s := &orchestrationStorage{ID: len(p.plan.Storage), Record: rec, Type: typ.Kind, RuntimeType: typ, Owner: function.Body, Parameter: true}
		p.plan.Storage = append(p.plan.Storage, s)
		function.Body.Storage = append(function.Body.Storage, s)
		frame.slots[rec] = s
		inner.assigned[s] = true
		return s
	}
	function.Self = allocate(source.Record, source.signature())
	context.values[source.Record] = function.Self
	for _, param := range source.Inputs {
		function.Inputs = append(function.Inputs, allocate(param.Record, param.Type))
	}
	function.Body.Body, inner, err = p.instructions(source.Body, inner)
	if err != nil {
		return nil, flow, err
	}
	if inner.live && len(source.Outputs) > 0 {
		return nil, flow, lerrors.InvalidInput.Errorf("Function %s does not return on every continuing path", source.Record.ID)
	}
	return result, flow, nil
}

func (e *orchestrationEmitter) boxedArguments(values []*orchestrationValue) ([]string, error) {
	result := make([]string, 0, len(values))
	for _, value := range values {
		expression, err := e.value(value)
		if err != nil {
			return nil, err
		}
		boxed, err := boxRuntimeValue(expression)
		if err != nil {
			return nil, err
		}
		result = append(result, boxed)
	}
	return result, nil
}
func (e *orchestrationEmitter) functionInstruction(current *orchestrationInstruction) ([]string, error) {
	arguments, err := e.boxedArguments(current.Arguments)
	if err != nil {
		return nil, err
	}
	if current.Source.Kind == instructionReturn {
		return []string{orchestrationStatement("return Ok(LFunctionOutcome::Returned(vec![" + strings.Join(arguments, ", ") + "]));")}, nil
	}
	if current.Source.Kind == instructionInvoke {
		target, err := e.value(current.Target)
		if err != nil {
			return nil, err
		}
		exit := "return Ok(())"
		if e.functionDepth > 0 {
			exit = "return Ok(LFunctionOutcome::Exit)"
		}
		result := []string{orchestrationStatement("let mut lResults = match (" + target.text + ").call(runtime, lProgram, vec![" + strings.Join(arguments, ", ") + "]).await? { LFunctionOutcome::Returned(values) => values, LFunctionOutcome::Exit => " + exit + " }.into_iter();")}
		for _, target := range current.Results {
			value, err := unboxRuntimeValue("lResults.next().ok_or_else(|| runtime::Error::new(\"missing Function result\"))?", target.Type)
			if err != nil {
				return nil, err
			}
			stored, err := e.storeRuntimeValue(target.Storage, value)
			if err != nil {
				return nil, err
			}
			result = append(result, stored...)
		}
		return result, nil
	}
	function := current.Function
	captures, err := e.boxedArguments(function.Captures)
	if err != nil {
		return nil, err
	}
	signature, err := runtimeTypeExpression(function.Source.signature(), e.plan.Identities)
	if err != nil {
		return nil, err
	}
	var captureTypes []string
	for _, capture := range function.Captures {
		typ, err := runtimeTypeExpression(capture.RuntimeType, e.plan.Identities)
		if err != nil {
			return nil, err
		}
		captureTypes = append(captureTypes, typ)
	}
	for _, source := range function.CaptureHandles {
		handle, err := e.captureReference(source)
		if err != nil {
			return nil, err
		}
		captures = append(captures, "LValue::Environment("+handle+")")
		captureTypes = append(captureTypes, "LType::Thing")
	}
	created := fmt.Sprintf("lCreatedFunction%d", function.Body.ID)
	result := []string{orchestrationStatement("let " + created + " = LCallable { signature: " + signature + ", captures: LEnvironment::typed(vec![" + strings.Join(captureTypes, ", ") + "], vec![" + strings.Join(captures, ", ") + "])? , invoke: std::rc::Rc::new(|runtime, lProgram, lSelf, lArguments| Box::pin(async move {")}
	for index, app := range e.plan.Apps {
		result = append(result, orchestrationStatement(fmt.Sprintf("let %s = lProgram.apps[%d].clone();", e.apps[app], index)))
	}
	result = append(result, orchestrationStatement("let "+function.Self.name()+" = lSelf.clone();"))
	if len(function.Inputs) > 0 {
		result = append(result, orchestrationStatement("let mut lArguments = lArguments.into_iter();"))
	}
	for _, parameter := range function.Inputs {
		value, err := unboxRuntimeValue("lArguments.next().ok_or_else(|| runtime::Error::new(\"missing Function argument\"))?", parameter.Type)
		if err != nil {
			return nil, err
		}
		mut := ""
		if parameter.Mutable {
			mut = "mut "
		}
		result = append(result, orchestrationStatement("let "+mut+parameter.name()+" = "+value+";"))
	}
	e.functionDepth++
	body, err := e.block(function.Body, false)
	e.functionDepth--
	if err != nil {
		return nil, err
	}
	result = append(result, body...)
	if len(function.Source.Outputs) == 0 {
		result = append(result, orchestrationStatement("Ok(LFunctionOutcome::Returned(vec![]))"))
	}
	result = append(result, orchestrationStatement("})) };"))
	stored, err := e.storeRuntimeValue(current.Target.Storage, created)
	if err != nil {
		return nil, err
	}
	result = append(result, stored...)
	return result, nil
}
