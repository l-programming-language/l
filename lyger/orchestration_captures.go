package Lyger

import (
	"fmt"
	"github.com/l-programming-language/l/lerrors"
)

type orchestrationCaptureLayout struct {
	Definition *record
	Values     []*instructionValue
	Calls      []*record
}
type orchestrationCallable struct {
	ID     int
	Record *record
	Layout *orchestrationCaptureLayout
	Owner  *orchestrationBlock
}

func (slot *orchestrationCallable) name() string { return fmt.Sprintf("lCallable%d", slot.ID) }

type orchestrationCaptureAccess struct {
	Name  string
	Index int
}
type orchestrationCallableReference struct {
	Slot    *orchestrationCallable
	Capture *orchestrationCaptureAccess
	Layout  *orchestrationCaptureLayout
}
type orchestrationCaptureContext struct {
	values map[*record]*orchestrationStorage
	calls  map[*record]*orchestrationCallableReference
}

func (p *orchestrationPlanner) prepareCaptureLayouts() error {
	p.captureLayouts = make(map[*record]*orchestrationCaptureLayout)
	p.callables = make(map[*record]*orchestrationCallable)
	return p.semantics.walkInstructions(func(current *instruction) error {
		if current.Kind == instructionBindAction {
			if current.Callable == nil {
				return lerrors.Internal.Errorf("capture declaration has no binding")
			}
			p.captureLayouts[current.Callable] = &orchestrationCaptureLayout{Definition: current.Callable, Values: current.Captures, Calls: current.CaptureCalls}
		}
		return nil
	})
}

func (p *orchestrationPlanner) callableSlot(binding *record) (*orchestrationCallable, error) {
	canonical, err := canonicalRecord(binding, p.semantics.Aliases)
	if err != nil {
		return nil, err
	}
	layout := p.captureLayouts[canonical]
	if layout == nil {
		return nil, lerrors.Internal.Errorf("action %s has no runtime capture layout", binding.ID)
	}
	owner, slots := p.plan.Root, p.callables
	if binding.storageScope.isParagraph() {
		var frame *orchestrationFrame
		for index := len(p.frames) - 1; index >= 0; index-- {
			if p.frames[index].block.Scope == binding.storageScope {
				frame = p.frames[index]
				break
			}
		}
		if frame == nil {
			return nil, lerrors.InvalidInput.Errorf("action %s is outside its binding activation", binding.ID)
		}
		if frame.callables == nil {
			frame.callables = make(map[*record]*orchestrationCallable)
		}
		owner, slots = frame.block, frame.callables
	}
	if owner == p.plan.Root && p.functionDepth > 0 {
		p.plan.FunctionProgramState = true
	}
	if slot := slots[binding]; slot != nil {
		return slot, nil
	}
	slot := &orchestrationCallable{ID: len(p.plan.Callables), Record: binding, Layout: layout, Owner: owner}
	slots[binding] = slot
	p.plan.Callables = append(p.plan.Callables, slot)
	owner.Callables = append(owner.Callables, slot)
	return slot, nil
}
func (p *orchestrationPlanner) captureReference(binding *record, flow orchestrationFlow) (*orchestrationCallableReference, error) {
	for index := len(p.captureContexts) - 1; index >= 0; index-- {
		if found := p.captureContexts[index].calls[binding]; found != nil {
			return found, nil
		}
	}
	slot, err := p.callableSlot(binding)
	if err != nil {
		return nil, err
	}
	if flow.live && !flow.bound[slot] {
		return nil, lerrors.InvalidInput.Errorf("action %s is not definitely bound", binding.ID)
	}
	return &orchestrationCallableReference{Slot: slot, Layout: slot.Layout}, nil
}
func (p *orchestrationPlanner) bindCapture(current *instruction, flow orchestrationFlow) (*orchestrationInstruction, orchestrationFlow, error) {
	result := &orchestrationInstruction{Source: current, Effects: orchestrationEffectAssign, ReachesNext: flow.live}
	slot, err := p.callableSlot(current.Callable)
	if err != nil {
		return nil, flow, err
	}
	result.Callable = slot
	if current.Kind == instructionAliasAction {
		result.CaptureSource, err = p.captureReference(current.Target.Record, flow)
		if err != nil {
			return nil, flow, err
		}
	} else {
		for _, source := range current.Captures {
			value, err := p.value(source, flow)
			if err != nil {
				return nil, flow, err
			}
			value.Access = orchestrationAccessSnapshot
			result.CaptureValues = append(result.CaptureValues, value)
			result.Effects |= orchestrationValueEffects(value)
		}
		for _, binding := range current.CaptureCalls {
			handle, err := p.captureReference(binding, flow)
			if err != nil {
				return nil, flow, err
			}
			result.CaptureHandles = append(result.CaptureHandles, handle)
		}
	}
	flow.bound[slot] = true
	return result, flow, nil
}
func (p *orchestrationPlanner) enterCaptureCall(current *instruction, result *orchestrationInstruction, flow orchestrationFlow) (func(), error) {
	if current.Callable == nil {
		return func() {}, nil
	}
	source, err := p.captureReference(current.Callable, flow)
	if err != nil {
		return nil, err
	}
	result.CaptureSource = source
	result.CaptureName = fmt.Sprintf("lCaptureCall%d", p.nextCaptureCall)
	p.nextCaptureCall++
	context := orchestrationCaptureContext{values: make(map[*record]*orchestrationStorage), calls: make(map[*record]*orchestrationCallableReference)}
	for index, value := range source.Layout.Values {
		typ, err := recordRuntimeType(value.Record)
		if err != nil {
			return nil, err
		}
		context.values[value.Record] = &orchestrationStorage{Record: value.Record, Type: value.Type, RuntimeType: typ, Mutable: true, Capture: &orchestrationCaptureAccess{Name: result.CaptureName, Index: index}}
	}
	for index, binding := range source.Layout.Calls {
		canonical, err := canonicalRecord(binding, p.semantics.Aliases)
		if err != nil {
			return nil, err
		}
		context.calls[binding] = &orchestrationCallableReference{Layout: p.captureLayouts[canonical], Capture: &orchestrationCaptureAccess{Name: result.CaptureName, Index: len(source.Layout.Values) + index}}
	}
	p.captureContexts = append(p.captureContexts, context)
	return func() { p.captureContexts = p.captureContexts[:len(p.captureContexts)-1] }, nil
}
