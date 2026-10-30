package Lyger

import (
	"fmt"
	"strconv"
)

func (current *instruction) wrapError(err error) error {
	if err == nil {
		return nil
	}
	if current == nil {
		return fmt.Errorf("orchestrator: %w", err)
	}
	span := current.Source
	if current.CallSite != nil && current.CallSite.source != nil {
		span = current.CallSite.source.Span
	}
	if span.Path == "" {
		return fmt.Errorf("orchestrator %s: %w", current.Kind, err)
	}
	return span.wrapError(fmt.Errorf("orchestrator %s: %w", current.Kind, err))
}

type conditionOperator uint8

const (
	conditionInvalid conditionOperator = iota
	conditionTruth
	conditionHas
	conditionIs
	conditionEquals
	conditionIsnt
	conditionLess
	conditionGreater
	conditionLessEqual
	conditionGreaterEqual
	conditionOperatorCount
)

var conditionOperators = [conditionOperatorCount]struct{ name, operator string }{
	conditionTruth: {"truth", ""}, conditionHas: {"has", ""},
	conditionIs: {"is", "=="}, conditionEquals: {"equals", "=="}, conditionIsnt: {"isnt", "!="},
	conditionLess: {"less", "<"}, conditionGreater: {"greater", ">"},
	conditionLessEqual: {"less_equal", "<="}, conditionGreaterEqual: {"greater_equal", ">="},
}

func (op conditionOperator) String() string {
	if op >= conditionOperatorCount {
		return "conditionOperator(" + strconv.Itoa(int(op)) + ")"
	}
	return conditionOperators[op].name
}

func parseConditionOperator(raw string) (conditionOperator, bool) {
	for op := conditionInvalid + 1; op < conditionOperatorCount; op++ {
		if conditionOperators[op].name == raw {
			return op, true
		}
	}
	return conditionInvalid, false
}

func (op conditionOperator) NativeOperator() (string, bool) {
	if op >= conditionOperatorCount {
		return "", false
	}
	operator := conditionOperators[op].operator
	return operator, operator != ""
}

func (op conditionOperator) accepts(left, right valueType) bool {
	switch op {
	case conditionIs, conditionEquals, conditionIsnt:
		scalar := left == valueText || left == valueNumber || left == valueTruth
		return scalar && left == right
	case conditionLess, conditionGreater, conditionLessEqual, conditionGreaterEqual:
		return left == valueNumber && right == valueNumber
	default:
		return false
	}
}
