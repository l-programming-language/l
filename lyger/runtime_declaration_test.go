package Lyger

import (
	"strings"
	"testing"
)

func TestRuntimeDeclarationsUseGrammaticalOperands(t *testing.T) {
	type assignmentExpectation struct {
		target, kind, reference string
	}
	member := []assignmentExpectation{{"y", "member", "HelloWorldApp"}}
	secondApp := "Create \"SecondApp\" as a New Basic App.\nThe App Prints x.\n"
	for _, tc := range []struct {
		name, source, shown string
		assignments         []assignmentExpectation
	}{
		{"named passive", "\"y\" is HelloWorldApp's Output.\nShow y.\n", "y", member},
		{"active control", "Define \"y\" as HelloWorldApp's Output.\nShow y.\n", "y", member},
		{"paragraph pronoun", "\"y\" is it's Output.\nShow y.\n", "y", member},
		{"named owner differs from paragraph subject", secondApp + "\"y\" is HelloWorldApp's Output.\nShow y.\n", "y", member},
		{"pronoun selects actual paragraph subject", secondApp + "\"y\" is it's Output.\nShow y.\n", "y", []assignmentExpectation{{"y", "member", "SecondApp"}}},
		{"alias owner", "Use HelloWorldApp as \"Runner\".\n\"y\" is Runner's Output.\nShow y.\n", "y", member},
		{"alias chain owner", "Use HelloWorldApp as \"Worker\".\nUse Worker as \"WorkerAlias\".\n\"y\" is WorkerAlias's Output.\nShow y.\n", "y", member},
		{"result alias", "\"y\" is HelloWorldApp's Output.\nUse y as \"Response\".\nShow Response.\n", "y", member},
		{"stored symbol copy", "Define \"y\" as HelloWorldApp's Output.\n\"z\" is y.\nShow z.\n", "z", []assignmentExpectation{{"y", "member", "HelloWorldApp"}, {"z", "symbol", "y"}}},
		{"trailing effect", "\"y\" is HelloWorldApp's Output and Show x.\n", "x", member},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := memberResolutionFixture(t)
			cut := strings.Index(source.main, "If it has an Error")
			if cut < 0 {
				t.Fatal("source fixture has no outer conditional")
			}
			source.main = source.main[:cut] + tc.source
			result := executableTestResult(t, source)
			compilation := result.Compilation
			instructions := compilation.Semantics.Instructions
			if len(instructions) != len(tc.assignments)+2 {
				t.Fatalf("instruction count = %d, want %d; aliases must not assign and declarations must not duplicate", len(instructions), len(tc.assignments)+2)
			}
			if instructions[0].Kind.String() != "call" || instructions[len(instructions)-1].Kind.String() != "show" {
				t.Fatal("runtime declarations must follow Execute and precede Show")
			}
			for index, current := range instructions {
				if index > 0 && current.Order < instructions[index-1].Order {
					t.Fatal("instructions no longer follow source order")
				}
				found := false
				for _, scoped := range compilation.Scope.Ordered {
					if scoped.Statement == current.Statement {
						found = true
						if current.Source != scoped.Source || current.Order != scoped.Order || current.Source != current.Statement.source.Span {
							t.Fatal("instruction lost its original source span or expanded-source order")
						}
					}
				}
				if !found {
					t.Fatal("instruction has no original scoped statement")
				}
			}
			for index, want := range tc.assignments {
				current := instructions[index+1]
				target := compilation.Memory[want.target]
				if current.Kind.String() != "assign" || current.Target == nil || current.Target.Record != target || current.Target.Record != compilation.Semantics.Subjects[current.Target.Token] || !current.Target.Runtime {
					t.Fatalf("assignment %d lost its declaration's storage identity: %+v", index, current)
				}
				if target.Value.Runtime == nil || target.Value.Type.String() != "Primitive" || target.Value.Primitive == nil || target.Value.Primitive.Type.String() != "Text" || target.Value.Primitive.Text != "" {
					t.Fatal("runtime assignment captured Output in semantic memory")
				}
				if current.Value == nil || current.Value.Kind.String() != want.kind || current.Value.Type.String() != "Text" || !current.Value.Runtime {
					t.Fatalf("assignment %d lost its runtime RHS type: %+v", index, current.Value)
				}
				if want.kind == "member" {
					memberResolutionInstructionValue(t, current.Value, compilation.Memory[want.reference], "output", "Text")
				} else if current.Value.Record != compilation.Memory[want.reference] {
					t.Fatal("stored symbol copy became a fresh member read")
				}
			}
			shown := instructions[len(instructions)-1].Value
			if shown == nil || shown.Kind.String() != "symbol" || shown.Record != compilation.Memory[tc.shown] {
				t.Fatal("Show or its alias lost the original storage identity")
			}
			if _, err := result.OrchestratorUnit("main", executableTestCompanions(result)); err != nil {
				t.Fatal(err)
			}
		})
	}
}
