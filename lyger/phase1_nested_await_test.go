package Lyger

import (
	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	"strings"
	"testing"
)

const phase1SharedAwaitSource = "Use IO's Print from `io` as \"Print\".\n" + `"firstMessage" is ` + "`first`" + `.
"secondMessage" is ` + "`second`" + `.

    Create "First" as a New Public Basic App.
The App Prints firstMessage.
Define "firstResult" as First's Output.
Await Second.
Show firstResult.
Create "Second" as a New Public Basic App.
The App Prints secondMessage.
Define "secondResult" as Second's Output.
Publicly define "published" as Second's Output.
Show secondResult.

Run First and Second.
Await First.
Show published.
`

func TestPhase1NestedAwaitAssignments(t *testing.T) {
	for _, name := range []string{
		"caller-local-and-publication", "cached-join", "three-level-storage", "partial-caller-assignment",
		"fresh-callee-isolation", "callee-local-does-not-escape", "exit-or-assignment",
		"callee-exiting-branch",
	} {
		t.Run(name, func(t *testing.T) {
			source := phase1SharedAwaitSource
			if name == "cached-join" {
				source = strings.Replace(source, "\nAwait First.\n", "\nAwait Second.\nAwait First.\n", 1)
			}
			if name == "three-level-storage" {
				source = strings.Replace(source, "Show secondResult.\n", "Await Third.\nShow secondResult.\nCreate \"Third\" as a New Public Basic App.\nThe App Prints firstMessage.\nDefine \"thirdResult\" as Third's Output.\nShow thirdResult.\n", 1)
				source = strings.Replace(source, "Run First and Second.", "Run First and Second and Third.", 1)
			}
			if name == "exit-or-assignment" {
				source = strings.Replace(source, "Show published.\n", "", 1)
			}
			r := phase1Require(t, source, nil)
			first, second := phase1Continuation(t, r, "First"), phase1Continuation(t, r, "Second")
			secondInstructions := 3
			if name == "three-level-storage" {
				secondInstructions = 4
				if phase1Continuation(t, r, "Third").Scope != first.Scope {
					t.Fatal("third continuation escaped the shared lexical paragraph")
				}
			}
			if first.Scope != second.Scope || len(first.Body) != 3 || len(second.Body) != secondInstructions {
				t.Fatal("regression requires distinct continuations in the same lexical paragraph")
			}
			assign, await, show := first.Body[0], first.Body[1], first.Body[2]
			if assign.Kind != instructionAssign || await.Kind != instructionAwait || show.Kind != instructionShow || assign.Target.Record != show.Value.Record {
				t.Fatal("caller snapshot lost its ordered assignment, join, or stored read")
			}
			condition := &instructionCondition{Operator: "truth", Left: &instructionValue{Kind: operandLiteral, Type: valueTruth, Literal: &primitive{Type: valueTruth, Truth: true}}}
			var wantError []string

			switch name {
			case "partial-caller-assignment":
				first.Body[0] = &instruction{Kind: instructionIf, Condition: condition, Then: []*instruction{assign}, Source: assign.Source}
				wantError = []string{"symbol firstResult is not definitely assigned"}
			case "fresh-callee-isolation":
				calleeRead := *show
				calleeRead.Source = SourceSpan{Path: "callee-read.l", Start: SourcePosition{Line: 1, Column: 1}}
				second.Body = append([]*instruction{&calleeRead}, second.Body...)
				wantError = []string{"callee-read.l:1:1", "symbol firstResult is not definitely assigned"}
			case "callee-local-does-not-escape":
				first.Body = []*instruction{await, show}
				second.Body = append([]*instruction{assign}, second.Body...)
				wantError = []string{"symbol firstResult is not definitely assigned"}
			case "exit-or-assignment":
				second.Body = append(second.Body, &instruction{Kind: instructionExit})
				first.Body = []*instruction{{Kind: instructionIf, Condition: condition, Then: []*instruction{await}, Else: []*instruction{assign}}, show}
			case "callee-exiting-branch":
				second.Body = []*instruction{{Kind: instructionIf, Condition: condition, Then: []*instruction{{Kind: instructionExit}}, Else: second.Body}}
			}
			if len(wantError) > 0 {
				_, err := r.OrchestratorUnit("entry", executableTestCompanions(r))
				if err == nil {
					t.Fatal("uninitialized private activation was accepted")
				}
				for _, text := range wantError {
					if !strings.Contains(err.Error(), text) {
						t.Fatalf("wrong assignment diagnostic: %v; want %q", err, text)
					}
				}
				return
			}
			generated, _ := phase1Generated(t, r)
			phase1ReceiverShape(t, generated, true)
			locals := make(map[string]bool)
			body := generated.function(t, "l_main").Body
			for i, token := range body {
				if token.Text != "let" || i+2 >= len(body) {
					continue
				}
				at := i + 1
				if body[at].Text == "mut" {
					at++
				}
				name := body[at].Text
				// Stored scalar declarations have explicit scalar types; App and invocation handles are separate bindings.
				if at+1 >= len(body) || body[at+1].Text != ":" {
					continue
				}
				end := at + 2
				for end < len(body) && body[end].Text != "=" && body[end].Text != ";" {
					end++
				}
				kind := nativetest.RustTokensText(body[at+2 : end])
				if kind != "Vec<u8>" && kind != "String" && kind != "f64" && kind != "bool" {
					continue
				}
				if locals[name] {
					t.Errorf("distinct continuation activations reuse private storage %s", name)
				}
				locals[name] = true
			}
			wantLocals := 2
			if name == "cached-join" || name == "three-level-storage" {
				wantLocals = 3
			}
			if len(locals) != wantLocals {
				t.Fatalf("private activation storage count = %d, want %d", len(locals), wantLocals)
			}
		})
	}
}
