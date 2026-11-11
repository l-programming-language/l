package Lyger

import (
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// Native instruction composition uses the checked semantic plan and the frozen
// AML control/operand interface. Source definitions still own calls and returns.
type nativeInstructionComposer struct {
	compilation     *Compilation
	active, dynamic map[*record]bool
}

func (e *nativeInstructionComposer) instructions(body []*orchestrationInstruction) (DefinitionInstance, error) {
	var steps []DefinitionInstance
	for _, current := range body {
		instance, err := e.instruction(current)
		if err != nil {
			return DefinitionInstance{}, current.Source.wrapError(err)
		}
		if instance.result != nil && instance.Expression == "" && instance.Words == nil {
			return DefinitionInstance{}, current.Source.wrapError(lerrors.TypeMismatch.Errorf("a value reference is not an executable definition"))
		}
		for _, step := range definitionSteps(instance) {
			if step.Words != nil {
				if len(step.Words) != 1 {
					return DefinitionInstance{}, current.Source.wrapError(lerrors.InvalidConfiguration.Errorf("definition composition requires one ordered word"))
				}
				step.Expression, step.Words = strings.Join(step.Words[0], "*"), nil
			}
			steps = append(steps, step)
		}
	}
	return composeDefinitionSequence(steps), nil
}

func (e *nativeInstructionComposer) instruction(current *orchestrationInstruction) (DefinitionInstance, error) {
	source := current.Source
	if source.Constructor != nil {
		instance, _, err := e.compilation.instantiateStaticConstruction(source.Constructor, nil, e.active, e.dynamic)
		return instance, err
	}
	if source.Application != nil {
		instance, _, err := e.compilation.instantiateStaticAction(source.Application, nil, e.active, e.dynamic)
		return instance, err
	}
	if source.Kind != instructionIf {
		return DefinitionInstance{}, lerrors.Unsupported.Errorf("%s requires unavailable runtime execution", source.Kind)
	}
	condition, err := e.condition(current.Condition)
	if err != nil {
		return DefinitionInstance{}, err
	}
	then, err := e.instructions(current.Then)
	if err != nil {
		return DefinitionInstance{}, err
	}
	otherwise, err := e.instructions(current.Else)
	if err != nil {
		return DefinitionInstance{}, err
	}
	parts := []DefinitionInstance{condition, nativeProcess(then)}
	if len(current.Else) != 0 {
		parts = append(parts, nativeProcess(otherwise))
	}
	expression := "BOG^GI^("
	for i, part := range parts {
		if i > 0 {
			expression += "*"
		}
		expression += part.Expression
	}
	result := composeDefinitionSequence(parts)
	result.Expression, result.sequence = expression+")", nil
	result.Span = source.Source
	return result, nil
}

func (e *nativeInstructionComposer) condition(condition *orchestrationCondition) (DefinitionInstance, error) {
	left, err := nativeConditionArgument(condition.Left)
	if err != nil {
		return DefinitionInstance{}, err
	}
	op, ok := parseConditionOperator(condition.Source.Operator)
	if !ok {
		return DefinitionInstance{}, lerrors.Internal.Errorf("invalid planned condition operator")
	}
	operands := []DefinitionArgument{left}
	operator := "=="
	operands = append(operands, DefinitionArgument{Kind: "Truth", Text: "true"})
	if op != conditionTruth {
		operator, ok = op.NativeOperator()
		if !ok {
			return DefinitionInstance{}, lerrors.Unsupported.Errorf("%s requires unavailable runtime execution", condition.Source.Operator)
		}
		right, err := nativeConditionArgument(condition.Right)
		if err != nil {
			return DefinitionInstance{}, err
		}
		operands[1] = right
	}
	prefix := ""
	if condition.Source.Negated {
		prefix = "!"
	}
	return composeNativeAML("DE{|Ops: ~&[Op{Type: Calculate, Prefix: "+prefix+", Operator: "+operator+", Operands: &[<@Operands|Atoms@>]&}Op]&~|}", SourceSpan{}, map[string]DefinitionArgument{"Operands": definitionAtomListArgument(operands...)})
}

func nativeConditionArgument(value *orchestrationValue) (DefinitionArgument, error) {
	if value == nil {
		return DefinitionArgument{}, lerrors.Internal.Errorf("planned value is missing")
	}
	if value.Source != nil && value.Source.Atom != nil {
		return *value.Source.Atom, nil
	}
	if value.Literal != nil {
		return nativeLiteralArgument(value.Literal)
	}
	return DefinitionArgument{}, lerrors.Unsupported.Errorf("%s condition operand requires unavailable runtime execution", value.Type)
}

func nativeLiteralArgument(value *primitive) (DefinitionArgument, error) {
	if value == nil {
		return DefinitionArgument{}, lerrors.Internal.Errorf("literal has no primitive value")
	}
	if value.Type == valueNumber {
		if _, finite := nativeNumberLiteral(value.Number); !finite {
			return DefinitionArgument{}, lerrors.Unsupported.Errorf("non-finite Number is not supported")
		}
	}
	argument, err := argumentFromToken(&token{mem: &entry{Type: storagePrimitive, Primitive: value}})
	if err != nil {
		return DefinitionArgument{}, err
	}
	if _, err := argument.metadataAtom(); err != nil {
		return DefinitionArgument{}, err
	}
	return argument, nil
}

func nativeProcess(body DefinitionInstance) DefinitionInstance {
	result := body
	result.Expression, result.sequence, result.result = "W", nil, nil
	if body.Expression != "" {
		result.Expression += "^(" + body.Expression + ")"
	}
	return result
}

func composeDefinitionSequence(steps []DefinitionInstance) DefinitionInstance {
	result := DefinitionInstance{sequence: []DefinitionInstance{}}
	var expressions []string
	for _, step := range steps {
		result.Dependencies = append(result.Dependencies, step.Dependencies...)
		result.Bindings = append(result.Bindings, step.Bindings...)
		for _, child := range appendDefinitionSequence(nil, step) {
			if child.Expression == "" {
				continue
			}
			expressions = append(expressions, child.Expression)
			result.sequence = append(result.sequence, child)
		}
	}
	result.Expression = strings.Join(expressions, "*")
	return result
}
