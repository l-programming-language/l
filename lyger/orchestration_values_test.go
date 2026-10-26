package Lyger

import (
	"math"
	"reflect"
	"strings"
	"testing"
)

func TestLoweredValueImportOrderAndOwnership(t *testing.T) {
	backing := make([]string, 1, 8)
	backing[0] = "std::fmt"
	paths := []string{"std::string::ToString", "std::fmt", "crate::runtime", "std::string::ToString"}
	base := orchestrationEvaluation("message", backing...)
	combined := orchestrationEvaluation("message", paths...)
	left := orchestrationEvaluation("message", "std::fmt", "crate::runtime")
	right := orchestrationEvaluation("message", "std::fmt", "std::borrow::Cow")
	// Production emits the supplied metadata in order, including repetitions.
	if combined != `E{|Imports: ~&["std::string::ToString", "std::fmt", "crate::runtime", "std::string::ToString"]&~, Statement: ~message~|}` ||
		base != `E{|Imports: ~&["std::fmt"]&~, Statement: ~message~|}` ||
		left != `E{|Imports: ~&["std::fmt", "crate::runtime"]&~, Statement: ~message~|}` ||
		right != `E{|Imports: ~&["std::fmt", "std::borrow::Cow"]&~, Statement: ~message~|}` {
		t.Fatalf("ordered import metadata = %q / %q / %q / %q", base, combined, left, right)
	}
	if !reflect.DeepEqual(backing, []string{"std::fmt"}) || !reflect.DeepEqual(paths, []string{"std::string::ToString", "std::fmt", "crate::runtime", "std::string::ToString"}) {
		t.Fatal("metadata emission changed caller import arrays")
	}
	copy := orchestrationEvaluation("message", backing...)
	paths[0], backing[0] = "changed", "changed"
	if copy != base || base != `E{|Imports: ~&["std::fmt"]&~, Statement: ~message~|}` || combined != `E{|Imports: ~&["std::string::ToString", "std::fmt", "crate::runtime", "std::string::ToString"]&~, Statement: ~message~|}` {
		t.Fatal("emitted metadata retained mutable caller storage")
	}

	original := &primitive{Type: valueText, Text: "`message`", Bytes: []byte("retained")}
	operand := &instructionValue{Kind: operandLiteral, Type: valueText, Literal: original}
	semantics := newSemanticProgram()
	semantics.Instructions = []*instruction{{Kind: instructionShow, Value: operand}, {Kind: instructionShow, Value: operand}}
	plan := requireOrchestrationPlan(t, semantics)
	first, second := plan.Root.Body[0].Value.Literal, plan.Root.Body[1].Value.Literal
	if first == original || second == original || first == second || first.Text != original.Text || second.Text != original.Text {
		t.Fatal("planned expression siblings share semantic literal storage")
	}
	first.Text, first.Bytes[0] = "`changed`", 'x'
	if original.Text != "`message`" || second.Text != "`message`" || string(original.Bytes) != "retained" || string(second.Bytes) != "retained" {
		t.Fatal("changing one planned expression changed its source or sibling")
	}
	original.Text, original.Bytes[0] = "`source changed`", 'y'
	actual, err := (&orchestrationEmitter{plan: plan}).instruction(plan.Root.Body[1])
	if err != nil || !reflect.DeepEqual(actual, []string{`BOD^E{|Statement: ~runtime.show(&"message".as_bytes().to_vec()).await?;~|}`}) || string(second.Bytes) != "retained" {
		t.Fatalf("emission lost the owned literal snapshot: %q/%v", actual, err)
	}
}

func TestLoweredValueCheckedConversions(t *testing.T) {
	for _, tc := range []struct {
		primitive *primitive
		want      string
	}{
		{&primitive{Type: valueText, Text: "`雪, <DEF> and \\n`"}, `BOD^E{|Statement: ~runtime.show(&"雪, <DEF> and \\n".as_bytes().to_vec()).await?;~|}`},
		{&primitive{Type: valueNumber, Number: 3.5}, `BOD^E{|Statement: ~runtime.show(&runtime::format_number(3.5_f64)).await?;~|}`},
		{&primitive{Type: valueTruth, Truth: true}, `BOD^E{|Statement: ~runtime.show(&true.to_string().into_bytes()).await?;~|}`},
		{&primitive{Type: valueBytes, Bytes: []byte{65, 0, 255}}, `BOD^E{|Statement: ~runtime.show(&vec![0x41u8, 0x00u8, 0xffu8]).await?;~|}`},
		{&primitive{Type: valueNull}, `BOD^E{|Statement: ~runtime.show(&(LValue::Null).display_bytes()?).await?;~|}`},
	} {
		t.Run(tc.primitive.Type.String(), func(t *testing.T) {
			before := *tc.primitive
			operand := &instructionValue{Kind: operandLiteral, Type: tc.primitive.Type, Literal: tc.primitive}
			semantics := newSemanticProgram()
			semantics.Instructions = []*instruction{{Kind: instructionShow, Value: operand}}
			plan := requireOrchestrationPlan(t, semantics)
			emitter := orchestrationEmitter{plan: plan}
			actual, err := emitter.instruction(plan.Root.Body[0])
			if err != nil || !reflect.DeepEqual(actual, []string{tc.want}) || plan.Root.Body[0].Value.Type != tc.primitive.Type {
				t.Fatalf("Show emission = %q, %v", actual, err)
			}
			semantics.Instructions = []*instruction{{Kind: instructionExit, Value: &instructionValue{Kind: operandError, Type: valueError, Payload: operand}}}
			failure, err := planOrchestration(semantics)
			if tc.primitive.Type == valueText {
				if err != nil {
					t.Fatal(err)
				}
				failureValue, err := emitter.value(failure.Root.Body[0].Value)
				if err != nil || failureValue.text != `runtime::Error::from_bytes("雪, <DEF> and \\n".as_bytes().to_vec())` || failureValue.kind != valueError {
					t.Fatalf("Text Error emission = %#v, %v", failureValue, err)
				}
				failure.Root.Body[0].Value.Payload.Literal.Text = "`changed`"
			} else if err == nil || failure != nil || !strings.Contains(err.Error(), "Error payload requires Text") {
				t.Fatalf("non-Text Error payload accepted: %#v/%v", failure, err)
			}
			again, err := emitter.instruction(plan.Root.Body[0])
			if !reflect.DeepEqual(*tc.primitive, before) || err != nil || !reflect.DeepEqual(again, []string{tc.want}) {
				t.Fatal("Error planning or emission changed an earlier expression")
			}
		})
	}
	for kind := valueInvalid; kind <= valueTypeCount; kind++ {
		operand := &instructionValue{Kind: operandLiteral, Type: kind, Literal: &primitive{Type: kind}}
		if kind != valueText && kind != valueNumber && kind != valueTruth && kind != valueBytes && kind != valueNull {
			actual, err := (&orchestrationPlanner{}).value(operand, newOrchestrationFlow())
			if err == nil || actual != nil {
				t.Errorf("unsupported literal type %s = %#v/%v", kind, actual, err)
			}
		}
		if kind != valueText {
			actual, err := (&orchestrationPlanner{}).value(&instructionValue{Kind: operandError, Type: valueError, Payload: operand}, newOrchestrationFlow())
			if err == nil || actual != nil {
				t.Errorf("unsupported Error type %s = %#v/%v", kind, actual, err)
			}
		}
	}
	_, app := planTestApp("Worker")
	failure := &instructionValue{Kind: operandError, Type: valueError, Payload: &instructionValue{Kind: operandLiteral, Literal: &primitive{Type: valueText, Text: "`reason`"}}}
	for _, operand := range []*instructionValue{app, failure} {
		semantics := newSemanticProgram()
		semantics.Instructions = []*instruction{{Kind: instructionShow, Value: operand}}
		if plan, err := planOrchestration(semantics); err == nil || plan != nil || !strings.Contains(err.Error(), "Show requires a scalar value") {
			t.Fatalf("nonscalar Show accepted: %#v/%v", plan, err)
		}
	}

	stored := planTestScalar("retained", valueText, nil)
	read := planTestRef(stored)
	semantics := newSemanticProgram()
	semantics.Instructions = []*instruction{
		{Kind: instructionShow, Value: read},
		{Kind: instructionExit, Value: &instructionValue{Kind: operandError, Type: valueError, Payload: read}},
	}
	plan := requireOrchestrationPlan(t, semantics)
	display, payload := plan.Root.Body[0].Value, plan.Root.Body[1].Value.Payload
	if len(plan.Storage) != 1 || display.Storage != payload.Storage || display.Access != orchestrationAccessBorrow || payload.Access != orchestrationAccessBorrow || plan.Storage[0].Mutable {
		t.Fatal("display and Error payload must borrow the same unchanged stored Text")
	}
	actual, err := (&orchestrationEmitter{plan: plan}).instructions(plan.Root.Body)
	want := []string{
		`BOD^E{|Statement: ~runtime.show(&lValue0).await?;~|}`,
		`BOD^E{|Statement: ~let lExit = runtime.fail(runtime::Error::from_bytes((lValue0).to_vec()));~|}`,
		`BOD^E{|Statement: ~return Err(lExit);~|}`,
	}
	if err != nil || !reflect.DeepEqual(actual, want) || stored.Value.Primitive.Text != "`value`" || plan.Storage[0].Initial == stored.Value.Primitive {
		t.Fatalf("Error did not take an independent Text snapshot: %q/%v", actual, err)
	}
}

func TestLoweredValueEvaluationEmission(t *testing.T) {
	for _, tc := range []struct {
		expression string
		imports    []string
		want       string
	}{
		{"ready", nil, "E{|Statement: ~ready~|}"},
		{`runtime::Error::new("reason")`, []string{"crate::runtime"}, `E{|Imports: ~&["crate::runtime"]&~, Statement: ~runtime::Error::new("reason")~|}`},
		{"left == right", []string{"std::fmt", "std::path::Path"}, `E{|Imports: ~&["std::fmt", "std::path::Path"]&~, Statement: ~left == right~|}`},
	} {
		if actual := orchestrationEvaluation(tc.expression, tc.imports...); actual != tc.want {
			t.Errorf("evaluation = %q, want %q", actual, tc.want)
		}
		if actual := orchestrationStatement(tc.expression, tc.imports...); actual != "BOD^"+tc.want {
			t.Errorf("statement = %q, want %q", actual, "BOD^"+tc.want)
		}
	}
}

func TestLoweredValuePrimitiveAndConditionResults(t *testing.T) {
	for _, tc := range []struct {
		primitive *primitive
		want      string
		kind      valueType
	}{
		{&primitive{Type: valueText, Text: "`line\n雪`"}, `"line\n雪".as_bytes().to_vec()`, valueText},
		{&primitive{Type: valueNumber, Number: -12.5}, "-12.5_f64", valueNumber},
		{&primitive{Type: valueTruth, Truth: false}, "false", valueTruth},
	} {
		actual, kind, err := nativePrimitiveExpression(tc.primitive)
		if err != nil || actual != tc.want || kind != tc.kind {
			t.Fatalf("primitive lowering = %#v/%v", actual, err)
		}
	}
	for _, primitive := range []*primitive{nil, {Type: valueNumber, Number: math.NaN()}, {Type: valueNumber, Number: math.Inf(1)}, {Type: valuePointer}} {
		actual, kind, err := nativePrimitiveExpression(primitive)
		if err == nil || actual != "" || kind != valueInvalid {
			t.Fatalf("unsupported primitive returned a value: %#v/%v", actual, err)
		}
	}
	left := &instructionValue{Kind: operandLiteral, Type: valueNumber, Literal: &primitive{Type: valueNumber, Number: 1}}
	right := &instructionValue{Kind: operandLiteral, Type: valueNumber, Literal: &primitive{Type: valueNumber, Number: 2}}
	condition, err := (&orchestrationPlanner{}).condition(&instructionCondition{Operator: "less", Left: left, Right: right, Negated: true}, newOrchestrationFlow())
	if err != nil {
		t.Fatal(err)
	}
	actual, err := (&orchestrationEmitter{}).condition(condition)
	if err != nil || actual != "!(1_f64 < 2_f64)" || condition.Left.Type != valueNumber || condition.Right.Type != valueNumber {
		t.Fatalf("condition lowering = %#v/%v", actual, err)
	}
}

func TestLoweredValueInstructionDependenciesAndState(t *testing.T) {
	for _, tc := range []struct {
		name    string
		current *instruction
		want    []string
		live    bool
	}{
		{
			name:    "Show Number",
			current: &instruction{Kind: instructionShow, Value: &instructionValue{Kind: operandLiteral, Type: valueNumber, Literal: &primitive{Type: valueNumber, Number: 3.5}}},
			want:    []string{`BOD^E{|Statement: ~runtime.show(&runtime::format_number(3.5_f64)).await?;~|}`},
			live:    true,
		},
		{
			name:    "Exit Error",
			current: &instruction{Kind: instructionExit, Value: &instructionValue{Kind: operandError, Type: valueError, Payload: &instructionValue{Kind: operandLiteral, Type: valueText, Literal: &primitive{Type: valueText, Text: "`reason`"}}}},
			want:    []string{`BOD^E{|Statement: ~let lExit = runtime.fail(runtime::Error::from_bytes("reason".as_bytes().to_vec()));~|}`, "BOD^E{|Statement: ~return Err(lExit);~|}"},
		},
		{name: "Exit without value", current: &instruction{Kind: instructionExit}, want: []string{"BOD^E{|Statement: ~return Ok(());~|}"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			semantics := newSemanticProgram()
			semantics.Instructions = []*instruction{tc.current}
			plan := requireOrchestrationPlan(t, semantics)
			planned := plan.Root.Body[0]
			actual, err := (&orchestrationEmitter{plan: plan}).instruction(planned)
			if err != nil || !reflect.DeepEqual(actual, tc.want) || planned.ReachesNext != tc.live {
				t.Fatalf("instruction lowering = %#v, live=%v, error=%v", actual, planned.ReachesNext, err)
			}
		})
	}
}
