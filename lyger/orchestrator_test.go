package Lyger

import (
	"bytes"
	"context"
	"crypto/sha256"
	Logiceme "github.com/l-programming-language/l/logiceme"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestOrchestratorFlowRequiresAssignmentAndDefersOutputReadiness(t *testing.T) {
	result := executableTestResult(t, memberResolutionFixture(t))
	app := result.Compilation.Memory["HelloWorldApp"]
	response := result.Compilation.Memory["y"]
	appValue := &instructionValue{Kind: operandSymbol, Type: valueApp, Record: app}
	output := &instructionValue{Kind: operandMember, Type: valueText, Owner: appValue, Member: memberOutput, Runtime: true}
	y := &instructionValue{Kind: operandSymbol, Type: valueText, Record: response, Runtime: true}
	literal := &instructionValue{Kind: operandLiteral, Type: valueText, Literal: &primitive{Type: valueText, Text: "`response`"}}
	condition := &instructionCondition{Operator: "truth", Left: &instructionValue{Kind: operandLiteral, Type: valueTruth, Literal: &primitive{Type: valueTruth, Truth: true}}}
	run := &instruction{Kind: instructionRun, Target: appValue, Synchronous: true}
	assign := &instruction{Kind: instructionAssign, Target: y, Value: literal}
	showOutput := &instruction{Kind: instructionShow, Value: output}
	showY := &instruction{Kind: instructionShow, Value: y}
	exit := &instruction{Kind: instructionExit}
	for _, tc := range []struct {
		name         string
		instructions []*instruction
		wantError    string
	}{
		{"output before run", []*instruction{showOutput}, ""},
		{"build is not run", []*instruction{{Kind: instructionBuild, Target: appValue, Synchronous: true}, showOutput}, ""},
		{"run then output", []*instruction{run, showOutput}, ""},
		{"conditional run only", []*instruction{{Kind: instructionIf, Condition: condition, Then: []*instruction{run}}, showOutput}, ""},
		{"both paths run", []*instruction{{Kind: instructionIf, Condition: condition, Then: []*instruction{run}, Else: []*instruction{run}}, showOutput}, ""},
		{"exit or run", []*instruction{{Kind: instructionIf, Condition: condition, Then: []*instruction{exit}, Else: []*instruction{run}}, showOutput}, ""},
		{"unassigned runtime symbol", []*instruction{showY}, "assigned"},
		{"one branch assigns", []*instruction{{Kind: instructionIf, Condition: condition, Then: []*instruction{assign}}, showY}, "assigned"},
		{"exit or assign", []*instruction{{Kind: instructionIf, Condition: condition, Then: []*instruction{exit}, Else: []*instruction{assign}}, showY}, ""},
		{"both paths assign", []*instruction{{Kind: instructionIf, Condition: condition, Then: []*instruction{assign}, Else: []*instruction{assign}}, showY}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result.Compilation.Semantics.Instructions = tc.instructions
			_, err := result.OrchestratorUnit("main", executableTestCompanions(result))
			if tc.wantError == "" && err != nil {
				t.Fatal(err)
			}
			if tc.wantError != "" && (err == nil || !strings.Contains(err.Error(), tc.wantError)) {
				t.Fatalf("flow validation = %v, want %s", err, tc.wantError)
			}
		})
	}
}

func TestOrchestratorScalarConditionsAndTypeErrors(t *testing.T) {
	result := executableTestResult(t, memberResolutionFixture(t))
	text := &instructionValue{Kind: operandLiteral, Type: valueText, Literal: &primitive{Type: valueText, Text: "`text`"}}
	number := &instructionValue{Kind: operandLiteral, Type: valueNumber, Literal: &primitive{Type: valueNumber, Number: 3.5}}
	truth := &instructionValue{Kind: operandLiteral, Type: valueTruth, Literal: &primitive{Type: valueTruth, Truth: true}}
	for _, tc := range []struct {
		name, op    string
		left, right *instructionValue
		valid       bool
	}{
		{"truth", "truth", truth, nil, true},
		{"text equals", "equals", text, text, true},
		{"number is", "is", number, number, true},
		{"truth isnt", "isnt", truth, truth, true},
		{"less", "less", number, number, true},
		{"greater", "greater", number, number, true},
		{"less equal", "less_equal", number, number, true},
		{"greater equal", "greater_equal", number, number, true},
		{"mixed equality", "is", text, number, false},
		{"text ordering", "less", text, text, false},
		{"truth requires Truth", "truth", text, nil, false},
		{"unknown operator", "contains", text, text, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result.Compilation.Semantics.Instructions = []*instruction{{Kind: instructionIf, Source: SourceSpan{Path: "condition.l", Start: SourcePosition{Line: 7, Column: 2}}, Condition: &instructionCondition{Operator: tc.op, Left: tc.left, Right: tc.right}, Then: []*instruction{{Kind: instructionExit}}}}
			_, err := result.OrchestratorUnit("main", executableTestCompanions(result))
			if tc.valid && err != nil {
				t.Fatal(err)
			}
			if !tc.valid && (err == nil || !strings.Contains(err.Error(), "condition.l:7:2")) {
				t.Fatalf("type error lacks source attribution: %v", err)
			}
		})
	}
}

func TestTruthLiteralConversionPreservesBothSpellings(t *testing.T) {
	for _, tc := range []struct {
		source string
		want   bool
	}{{"true", true}, {"false", false}, {"?true?", true}, {"?false?", false}} {
		if got := bytesToBool([]byte(tc.source)); got != tc.want {
			t.Errorf("%s = %t, want %t", tc.source, got, tc.want)
		}
		if tc.source == "true" || tc.source == "false" {
			value, err := loadPrimitive([]byte(tc.source))
			if err != nil || value.Type.String() != "Truth" || value.Truth != tc.want {
				t.Errorf("load source Truth %s: %+v, %v", tc.source, value, err)
			}
		}
	}
}

func TestOrchestratorStandaloneRuntimeAssignmentRemainsOrdered(t *testing.T) {
	for _, tc := range []struct{ name, following string }{
		{"direct", "Show y.\n"},
		{"alias", "Use y as \"Response\".\nShow Response.\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := memberResolutionFixture(t)
			cut := strings.Index(source.main, "If it has an Error")
			if cut < 0 {
				t.Fatal("source fixture has no outer error conditional")
			}
			source.main = source.main[:cut] + "Define \"y\" as HelloWorldApp's Output.\n" + tc.following
			result := executableTestResult(t, source)
			instructions := result.Compilation.Semantics.Instructions
			if len(instructions) != 3 || instructions[0].Kind.String() != "call" || instructions[1].Kind.String() != "assign" || instructions[2].Kind.String() != "show" {
				t.Fatalf("runtime declaration must follow Execute and precede Show: %+v", instructions)
			}
			if instructions[1].Target.Record != result.Compilation.Memory["y"] || instructions[1].Value.Kind.String() != "member" || instructions[1].Value.Owner.Record != result.Compilation.Memory["HelloWorldApp"] || instructions[2].Value.Record != result.Compilation.Memory["y"] {
				t.Fatal("runtime declaration or alias lost its target or actual member owner")
			}
			if _, err := result.OrchestratorUnit("main", executableTestCompanions(result)); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestOrchestratorUnitUsesTypedIRWithoutChangingApplication(t *testing.T) {
	result := executableTestResult(t, memberResolutionFixture(t))
	var inequality bool
	for _, current := range result.Compilation.Semantics.Instructions {
		if current.Kind.String() == "if" && current.Condition.Operator == "isnt" {
			inequality = true
			if current.Condition.Negated {
				t.Fatal("source isnt was negated twice")
			}
		}
	}
	if !inequality {
		t.Fatal("source isnt condition was omitted")
	}
	before := bytes.Clone(result.BackendAML)
	companions := executableTestCompanions(result)
	unit, err := result.OrchestratorUnit("entry", companions)
	if err != nil {
		t.Fatal(err)
	}
	if unit.ID != "entry" || bytes.Count(unit.BackendAML, []byte("<BOOK ")) != 1 || !bytes.Contains(unit.BackendAML, []byte(`<BOOK "entry">`)) {
		t.Fatalf("orchestrator unit lacks its own application/main: %s", unit.BackendAML)
	}
	for _, want := range []string{"Hello World!", companions[0].RelativeExecutable} {
		if !bytes.Contains(unit.BackendAML, []byte(want)) {
			t.Errorf("orchestrator omitted %q", want)
		}
	}
	if bytes.Contains(unit.BackendAML, []byte("EA{|")) || !bytes.Contains(unit.BackendAML, []byte("GI")) {
		t.Fatal("orchestrator must lower structured conditions through glyphs")
	}
	again, err := result.OrchestratorUnit("entry", companions)
	if err != nil || !bytes.Equal(unit.BackendAML, again.BackendAML) || !bytes.Equal(before, result.BackendAML) || !bytes.Equal(before, result.Compilation.Program.BackendBytes) {
		t.Fatalf("orchestrator rendering changed its input or is nondeterministic: %v", err)
	}
	unit.BackendAML[0] = '!'
	if again.BackendAML[0] != '<' || result.BackendAML[0] != '<' {
		t.Fatal("orchestrator shares AML backing memory")
	}
}

func TestOrchestratorGlyphsGenerateBuildableSource(t *testing.T) {
	result := executableTestResult(t, memberResolutionFixture(t))
	buildOrchestratorForTest(t, result)
}

type orchestratorTestBuild struct {
	instrumentation string
	race            bool
}

func buildOrchestratorForTest(t *testing.T, result *CompilationResult, options ...orchestratorTestBuild) string {
	t.Helper()
	if len(options) > 1 {
		t.Fatal("at most one observer option")
	}
	option := orchestratorTestBuild{}
	if len(options) == 1 {
		option = options[0]
	}
	unit, err := result.OrchestratorUnit("main", executableTestCompanions(result))
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := Logiceme.NewLogicemeService(phase1Options(t).Resources).Evaluate(unit.BackendAML, Logiceme.GenerationOptions{IncludeRuntime: true})
	recordBundle(t, bundle)
	if err != nil {
		t.Fatalf("actual Rust orchestration generation failed; build/runtime blocked: %v", err)
	}
	return buildActualBundle(t, bundle, "main", true, option.instrumentation)
}

func TestOrchestratorScalarConditionsExecuteWithExactText(t *testing.T) {
	result := executableTestResult(t, memberResolutionFixture(t))
	text := func(value string) *instructionValue {
		return &instructionValue{Kind: operandLiteral, Type: valueText, Literal: &primitive{Type: valueText, Text: strconv.Quote(value)}}
	}
	number := func(value float64) *instructionValue {
		return &instructionValue{Kind: operandLiteral, Type: valueNumber, Literal: &primitive{Type: valueNumber, Number: value}}
	}
	truth := func(value bool) *instructionValue {
		return &instructionValue{Kind: operandLiteral, Type: valueTruth, Literal: &primitive{Type: valueTruth, Truth: value}}
	}
	show := func(value *instructionValue) *instruction { return &instruction{Kind: instructionShow, Value: value} }
	var instructions []*instruction
	for index, condition := range []*instructionCondition{
		{Operator: "truth", Left: truth(true)},
		{Operator: "equals", Left: text("same"), Right: text("same")},
		{Operator: "is", Left: number(2), Right: number(2)},
		{Operator: "isnt", Left: truth(true), Right: truth(false)},
		{Operator: "less", Left: number(1), Right: number(2)},
		{Operator: "greater", Left: number(2), Right: number(1)},
		{Operator: "less_equal", Left: number(2), Right: number(2)},
		{Operator: "greater_equal", Left: number(2), Right: number(1)},
	} {
		instructions = append(instructions, &instruction{Kind: instructionIf, Condition: condition, Then: []*instruction{show(text(string(rune('a' + index))))}, Else: []*instruction{show(text("WRONG"))}})
	}
	instructions = append(instructions,
		&instruction{Kind: instructionIf, Condition: &instructionCondition{Operator: "truth", Left: truth(true), Negated: true}, Then: []*instruction{show(text("WRONG"))}, Else: []*instruction{show(text("i"))}},
		&instruction{Kind: instructionIf, Condition: &instructionCondition{Operator: "truth", Left: truth(true)}, Then: []*instruction{{Kind: instructionIf, Condition: &instructionCondition{Operator: "truth", Left: truth(true)}, Then: []*instruction{show(text("j"))}}}},
		show(number(3.5)), show(truth(true)), show(truth(false)),
	)
	wantText := ";\nraw \"quote\" \\ path ~, ^ * | <DEF> `雪`\n"
	instructions = append(instructions, show(text(wantText)))
	result.Compilation.Semantics.Instructions = instructions
	binary := buildOrchestratorForTest(t, result)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("run scalar orchestration: %v\n%s", err, stderr.Bytes())
	}
	if stdout.String() != "abcdefghij3.5truefalse"+wantText || stderr.Len() != 0 {
		generated, readErr := os.ReadFile(filepath.Join(filepath.Dir(binary), "src/main.rs"))
		if readErr != nil {
			t.Fatalf("read generated orchestration: %v", readErr)
		}
		t.Fatalf("scalar execution changed values or control flow:\nstdout=%q\nstderr=%q\n%s", stdout.String(), stderr.String(), generated)
	}
}

func TestOrchestratorAssignmentRetainsValueAfterAnotherRun(t *testing.T) {
	helper := buildNativeHelper(t)
	child, err := os.ReadFile(helper)
	if err != nil {
		t.Fatal(err)
	}
	buildOwner := t.Name()
	t.Cleanup(func() {
		got, err := os.ReadFile(helper)
		if err != nil || !bytes.Equal(got, child) {
			t.Errorf("shared native helper changed after assignment profiles: %v", err)
		}
	})
	for _, tc := range []struct{ name, declaration string }{
		{"active", "Define \"y\" as HelloWorldApp's Output."},
		{"passive", "\"y\" is HelloWorldApp's Output."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("native helper build: %s; SHA256 %x; orchestrator builds remain independent", buildOwner, sha256.Sum256(child))
			assertOrchestratorAssignmentRetainsValueAfterAnotherRun(t, tc.declaration, child)
		})
	}
}

func assertOrchestratorAssignmentRetainsValueAfterAnotherRun(t *testing.T, declaration string, child []byte) {
	t.Helper()
	source := memberResolutionFixture(t)
	cut := strings.Index(source.main, "If it has an Error")
	if cut < 0 {
		t.Fatal("source fixture has no outer conditional")
	}
	source.main = source.main[:cut] + declaration + "\nRun HelloWorldApp.\nAwait HelloWorldApp.\nShow y.\nShow HelloWorldApp's Output.\n"
	result := executableTestResult(t, source)
	binary := buildOrchestratorForTest(t, result)
	dir := filepath.Dir(binary)
	companion := filepath.Join(dir, filepath.FromSlash(executableTestCompanions(result)[0].RelativeExecutable))
	if err := os.MkdirAll(filepath.Dir(companion), 0700); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(companion, child, 0700); err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(companion); err != nil || !bytes.Equal(got, child) {
		t.Fatalf("runtime companion differs from parent Cargo artifact: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var stdout, stderr bytes.Buffer
	run := exec.CommandContext(ctx, binary)
	run.Dir = t.TempDir()
	run.Stdout, run.Stderr = &stdout, &stderr
	if err := run.Run(); err != nil {
		t.Fatalf("run source orchestration: %v\n%s", err, stderr.Bytes())
	}
	if stdout.String() != "first\nsecond\n" || stderr.Len() != 0 {
		t.Fatalf("assignment followed later Output or replayed child stdout: %q / %q", stdout.String(), stderr.String())
	}
}
