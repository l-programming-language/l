package Lyger

import (
	"github.com/l-programming-language/l/lerrors"
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
		context.values[value.Record] = &orchestrationStorage{Record: value.Record, Type: value.Type, RuntimeType: value.RuntimeType, Mutable: true, Capture: &orchestrationCaptureAccess{Name: "lEnvironment", Index: index}}
	}
	for index, binding := range source.CaptureCalls {
		context.calls[binding] = &orchestrationCallableReference{Layout: function.CaptureHandles[index].Layout, Capture: &orchestrationCaptureAccess{Name: "lEnvironment", Index: len(function.Captures) + index}}
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
