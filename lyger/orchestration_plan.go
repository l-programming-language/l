package Lyger

import (
	"fmt"

	"github.com/l-programming-language/l/lerrors"
)

// orchestrationPlan is a typed, source-preserving boundary between L's
// semantic instructions and Rust source composition. It contains no rendered
// source and never inspects caller-supplied native fragments.
type orchestrationPlan struct {
	FunctionProgramState bool
	Identities           map[*record]uint64
	Root                 *orchestrationBlock
	Storage              []*orchestrationStorage
	Apps                 []*App
	Callables            []*orchestrationCallable
	RuntimeValues        bool
}

// A block identifies an activation, not merely a lexical scope. Two App
// continuations may share Scope but must never share their private Storage.
type orchestrationBlock struct {
	ID        int
	Scope     *lexicalScope
	Storage   []*orchestrationStorage
	Body      []*orchestrationInstruction
	Callables []*orchestrationCallable
}

type orchestrationStorage struct {
	Parameter   bool
	RuntimeType runtimeTypeRef
	ID          int
	Record      *record
	Type        valueType
	Owner       *orchestrationBlock
	Initial     *primitive
	Mutable     bool
	Published   bool
	Capture     *orchestrationCaptureAccess
}

// Access describes the semantic boundary, not a blanket clone policy. A stored
// Text is borrowed for inspection; assignment makes an independent snapshot.
// Runtime output is already an owned snapshot. Scalar numbers/truths copy.
type orchestrationAccess uint8

const (
	orchestrationAccessCopy orchestrationAccess = iota
	orchestrationAccessBorrow
	orchestrationAccessOwned
	orchestrationAccessSnapshot
)

type orchestrationValue struct {
	RuntimeType runtimeTypeRef
	Elements    []*orchestrationValue
	Key         *orchestrationValue
	Source      *instructionValue
	Type        valueType
	Record      *record
	Storage     *orchestrationStorage
	Literal     *primitive
	App         *App
	Owner       *orchestrationValue
	Payload     *orchestrationValue
	Access      orchestrationAccess
}

type orchestrationCondition struct {
	Source *instructionCondition
	Left   *orchestrationValue
	Right  *orchestrationValue
}

type orchestrationEffects uint16

const (
	orchestrationEffectAssign orchestrationEffects = 1 << iota
	orchestrationEffectBuild
	orchestrationEffectRun
	orchestrationEffectStop
	orchestrationEffectAwait
	orchestrationEffectReadOutput
	orchestrationEffectReadError
	orchestrationEffectShow
	orchestrationEffectExit
)

type orchestrationInstruction struct {
	AppCases           []*orchestrationInstruction
	Source             *instruction
	Target             *orchestrationValue
	Value              *orchestrationValue
	Condition          *orchestrationCondition
	Then               []*orchestrationInstruction
	Else               []*orchestrationInstruction
	Actions            []*orchestrationInstruction
	Block              *orchestrationBlock
	Continuation       *orchestrationBlock
	App                *App
	Effects            orchestrationEffects
	ReachesNext        bool
	Callable           *orchestrationCallable
	CaptureSource      *orchestrationCallableReference
	CaptureValues      []*orchestrationValue
	CaptureHandles     []*orchestrationCallableReference
	CaptureName        string
	Key                *orchestrationValue
	Function           *orchestrationFunction
	Arguments, Results []*orchestrationValue
}

type orchestrationFrame struct {
	block     *orchestrationBlock
	slots     map[*record]*orchestrationStorage
	callables map[*record]*orchestrationCallable
}

type orchestrationPlanner struct {
	functionDepth   int
	semantics       *semanticProgram
	plan            *orchestrationPlan
	dynamic         map[*record]bool
	shared          map[*record]*orchestrationStorage
	frames          []*orchestrationFrame
	active          map[*App]bool
	apps            map[*App]bool
	nextBlock       int
	captureLayouts  map[*record]*orchestrationCaptureLayout
	callables       map[*record]*orchestrationCallable
	captureContexts []orchestrationCaptureContext
	nextCaptureCall int
}

func planOrchestration(semantics *semanticProgram) (*orchestrationPlan, error) {
	if semantics == nil {
		return nil, lerrors.Internal.Errorf("orchestration requires semantic instructions")
	}
	dynamic, err := semantics.runtimeRecords()
	if err != nil {
		return nil, err
	}
	p := &orchestrationPlanner{semantics: semantics, plan: &orchestrationPlan{Identities: make(map[*record]uint64)}, dynamic: dynamic, shared: make(map[*record]*orchestrationStorage), active: make(map[*App]bool), apps: make(map[*App]bool)}
	p.plan.Root = p.newBlock(nil)
	if err := p.prepareCaptureLayouts(); err != nil {
		return nil, err
	}
	body, _, err := p.instructions(semantics.Instructions, newOrchestrationFlow())
	if err != nil {
		return nil, err
	}
	p.plan.Root.Body = body
	return p.plan, nil
}

func (p *orchestrationPlanner) newBlock(scope *lexicalScope) *orchestrationBlock {
	block := &orchestrationBlock{ID: p.nextBlock, Scope: scope}
	p.nextBlock++
	return block
}

func (p *orchestrationPlanner) record(value *instructionValue) (*record, valueType, error) {
	return p.semantics.instructionRecord(value)
}

func planPrimitive(original *primitive) (*primitive, error) {
	if original == nil {
		return nil, lerrors.Internal.Errorf("literal has no primitive value")
	}
	value, err := original.value()
	if err != nil {
		return nil, err
	}
	if value.Type != valueText && value.Type != valueNumber && value.Type != valueTruth && value.Type != valueBytes && value.Type != valueNull {
		return nil, lerrors.Unsupported.Errorf("unsupported primitive %s", value.Type)
	}
	cloned := *original
	cloned.Bytes = append([]byte(nil), original.Bytes...)
	return &cloned, nil
}

func (p *orchestrationPlanner) storage(rec *record, kind valueType) (*orchestrationStorage, *primitive, error) {
	for index := len(p.captureContexts) - 1; index >= 0; index-- {
		if slot := p.captureContexts[index].values[rec]; slot != nil {
			return slot, nil, nil
		}
	}
	if _, ok := runtimeTypeName(kind); !ok {
		return nil, nil, lerrors.Unsupported.Errorf("unsupported runtime symbol %s of type %s", rec.ID, kind)
	}
	owner, slots := p.plan.Root, p.shared
	if rec.storageScope.isParagraph() {
		var frame *orchestrationFrame
		for index := len(p.frames) - 1; index >= 0; index-- {
			if p.frames[index].block.Scope == rec.storageScope {
				frame = p.frames[index]
				break
			}
		}
		if frame == nil {
			if p.dynamic[rec] || rec.Value.Runtime != nil {
				return nil, nil, lerrors.Unsupported.Errorf("private runtime capture %s is outside its paragraph activation", rec.ID)
			}
			literal, err := planPrimitive(rec.Value.Primitive)
			return nil, literal, err
		}
		owner, slots = frame.block, frame.slots
	}
	if owner == p.plan.Root && p.functionDepth > 0 {
		p.plan.FunctionProgramState = true
	}
	if slot := slots[rec]; slot != nil {
		return slot, nil, nil
	}
	typ, err := recordRuntimeType(rec)
	if err != nil {
		return nil, nil, err
	}
	slot := &orchestrationStorage{ID: len(p.plan.Storage), Record: rec, Type: kind, RuntimeType: typ, Owner: owner, Published: owner == p.plan.Root && p.dynamic[rec]}
	if !p.dynamic[rec] && rec.Value.Runtime == nil {
		var err error
		slot.Initial, err = planPrimitive(rec.Value.Primitive)
		if err != nil {
			return nil, nil, err
		}
	}
	slots[rec] = slot
	owner.Storage = append(owner.Storage, slot)
	p.plan.Storage = append(p.plan.Storage, slot)
	return slot, nil, nil
}

func (p *orchestrationPlanner) value(value *instructionValue, flow orchestrationFlow) (*orchestrationValue, error) {
	if value == nil {
		return nil, lerrors.Internal.Errorf("missing instruction operand")
	}
	result := &orchestrationValue{Source: value, Type: value.Type, RuntimeType: value.RuntimeType, Access: orchestrationAccessCopy}
	if result.RuntimeType.Kind == valueInvalid {
		result.RuntimeType.Kind = value.Type
	}
	switch value.Kind {
	case operandConstruct, operandRead:
		return p.runtimeOperand(value, flow)
	case operandLiteral:
		literal, err := planPrimitive(value.Literal)
		if err != nil {
			return nil, err
		}
		if value.Type != valueInvalid && value.Type != literal.Type {
			return nil, lerrors.Internal.Errorf("literal has inconsistent type %s", value.Type)
		}
		result.Literal, result.Type, result.RuntimeType = literal, literal.Type, runtimeTypeRef{Kind: literal.Type}
		if p.plan != nil && literal.Type == valueNull {
			p.plan.RuntimeValues = true
		}
		if result.Type == valueText {
			result.Access = orchestrationAccessOwned
		}
	case operandSymbol:
		rec, kind, err := p.record(value)
		if err != nil {
			return nil, err
		}
		result.RuntimeType, err = recordRuntimeType(rec)
		if err != nil {
			return nil, err
		}
		result.Record, result.Type = rec, kind
		if kind == valueApp && rec.Value.Runtime == nil && !p.dynamic[rec] {
			stored, err := rec.Value.value()
			if err != nil {
				return nil, err
			}
			result.App = stored.Data.(*App) //nolint:errcheck // value() validates the tagged payload before this type-specific projection.
			if !p.apps[result.App] {
				p.apps[result.App] = true
				p.plan.Apps = append(p.plan.Apps, result.App)
			}
			result.Access = orchestrationAccessBorrow
			return result, nil
		}
		result.Storage, result.Literal, err = p.storage(rec, kind)
		if err != nil {
			return nil, err
		}
		if flow.live && (p.dynamic[rec] || value.Runtime || rec.Value.Runtime != nil) && result.Storage.Capture == nil && !flow.assigned[result.Storage] {
			return nil, lerrors.InvalidInput.Errorf("symbol %s is not definitely assigned", rec.ID)
		}
		if kind != valueNumber && kind != valueTruth && kind != valueNull {
			result.Access = orchestrationAccessBorrow
			if result.Literal != nil {
				result.Access = orchestrationAccessOwned
			}
		}
	case operandMember:
		owner, err := p.value(value.Owner, flow)
		if err != nil {
			return nil, err
		}
		if owner.Type != valueApp {
			return nil, lerrors.TypeMismatch.Errorf("Build/Run or member access requires an application, got %s", owner.Type)
		}
		kind, ok := value.Member.ValueType()
		if !ok {
			return nil, lerrors.Internal.Errorf("unsupported application member %s", value.Member)
		}
		if value.Type != valueInvalid && value.Type != kind {
			return nil, lerrors.Internal.Errorf("member has inconsistent value type %s", value.Type)
		}
		result.Owner, result.Type, result.RuntimeType, result.Access = owner, kind, runtimeTypeRef{Kind: kind}, orchestrationAccessOwned
	case operandError:
		payload, err := p.value(value.Payload, flow)
		if err != nil {
			return nil, err
		}
		if payload.Type != valueText {
			return nil, lerrors.TypeMismatch.Errorf("Error payload requires Text, got %s", payload.Type)
		}
		result.Payload, result.Type, result.RuntimeType, result.Access = payload, valueError, runtimeTypeRef{Kind: valueError}, orchestrationAccessOwned
	default:
		return nil, lerrors.Internal.Errorf("unsupported instruction value %s", value.Kind)
	}
	return result, nil
}

func (p *orchestrationPlanner) condition(condition *instructionCondition, flow orchestrationFlow) (*orchestrationCondition, error) {
	if condition == nil {
		return nil, lerrors.Internal.Errorf("condition requires one supported operator")
	}
	if len(condition.Terms) != 0 {
		return nil, lerrors.Unsupported.Errorf("condition requires one supported operator")
	}
	left, err := p.value(condition.Left, flow)
	if err != nil {
		return nil, err
	}
	result := &orchestrationCondition{Source: condition, Left: left}
	op, _ := parseConditionOperator(condition.Operator)
	if op == conditionTruth {
		if left.Type != valueTruth || condition.Right != nil {
			return nil, lerrors.TypeMismatch.Errorf("truth condition requires one Truth operand")
		}
		return result, nil
	}
	right, err := p.value(condition.Right, flow)
	if err != nil {
		return nil, err
	}
	result.Right = right
	switch op {
	case conditionHas:
		if left.Type != valueApp || right.Source.Kind != operandMember || right.Source.Member != memberError || right.Type != valueError {
			return nil, lerrors.TypeMismatch.Errorf("has condition supports an application's Error")
		}
		if !samePlannedApp(right.Owner, left) {
			return nil, lerrors.TypeMismatch.Errorf("has condition Error belongs to a different application")
		}
	case conditionIs, conditionEquals, conditionIsnt:
		if left.Type != right.Type || (left.Type != valueText && left.Type != valueNumber && left.Type != valueTruth) {
			return nil, lerrors.TypeMismatch.Errorf("%s requires matching scalar types, got %s and %s", condition.Operator, left.Type, right.Type)
		}
	case conditionLess, conditionGreater, conditionLessEqual, conditionGreaterEqual:
		if left.Type != valueNumber || right.Type != valueNumber {
			return nil, lerrors.TypeMismatch.Errorf("%s requires Number operands", condition.Operator)
		}
	default:
		return nil, lerrors.Unsupported.Errorf("unsupported condition operator %s", condition.Operator)
	}
	return result, nil
}

func orchestrationValueEffects(value *orchestrationValue) orchestrationEffects {
	if value == nil {
		return 0
	}
	effects := orchestrationValueEffects(value.Owner) | orchestrationValueEffects(value.Payload) | orchestrationValueEffects(value.Key)
	for _, child := range value.Elements {
		effects |= orchestrationValueEffects(child)
	}
	if value.Source.Kind == operandMember {
		if value.Source.Member == memberOutput {
			effects |= orchestrationEffectReadOutput
		} else if value.Source.Member == memberError {
			effects |= orchestrationEffectReadError
		}
	}
	return effects
}

func (slot *orchestrationStorage) name() string { return fmt.Sprintf("lValue%d", slot.ID) }
