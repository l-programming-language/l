package Lyger

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const paragraphCaptureDefinition = "Publicly create \"Report|-s|-ed|-ing\" for Showing privateText and then Showing any Text."

// Inspect canonical operands as well as generated Rust: an equal literal alone
// would not prove that an action kept its defining binding and source span.
func paragraphCaptureCall(t *testing.T, compilation *Compilation, definitionFile, invocation string) (*record, *instruction) {
	t.Helper()
	definition := compilation.Memory["Report"]
	if definition == nil {
		t.Fatal("published source action is missing")
	}
	declaration := compilation.Semantics.Declarations[definition]
	if declaration == nil || filepath.Base(declaration.source.Span.Path) != definitionFile || strings.TrimSpace(string(declaration.source.Raw)) != paragraphCaptureDefinition {
		t.Fatal("published action lost its source declaration")
	}
	var captured *record
	for _, tok := range declaration.tokens {
		if string(tok.rawBytes) == "privateText" && tok.binding != nil {
			captured = tok.binding.Record
		}
	}
	if captured == nil || compilation.Memory["privateText"] != nil {
		t.Fatal("private captured declaration is missing or leaked into global memory")
	}
	var calls []*instruction
	var visit func([]*instruction)
	visit = func(body []*instruction) {
		for _, current := range body {
			if current.Kind == instructionCall && current.Target != nil && current.Target.Record == definition {
				calls = append(calls, current)
			}
			visit(current.Actions)
			visit(current.Then)
			visit(current.Else)
			if current.Block != nil {
				visit(current.Block.Body)
			}
		}
	}
	visit(compilation.Semantics.Instructions)
	for _, body := range compilation.Semantics.Continuations {
		visit(body.Body)
	}
	if len(calls) != 1 {
		t.Fatalf("canonical Report expansions = %d; want one", len(calls))
	}
	call := calls[0]
	if call.Definition != declaration || call.Statement == nil || filepath.Base(call.Statement.source.Span.Path) != "main.l" || strings.TrimSpace(string(call.Statement.source.Raw)) != invocation {
		t.Fatal("action expansion conflated definition and invocation provenance")
	}
	if len(call.Actions) != 2 || call.Actions[0].Kind != instructionShow || call.Actions[1].Kind != instructionShow {
		t.Fatal("source action lost its ordered captured-value and argument effects")
	}
	for _, action := range call.Actions {
		if action.Statement != declaration || action.Source != declaration.source.Span || action.CallSite != call.Statement {
			t.Fatal("expanded action lost its defining source span or outer invocation")
		}
	}
	value := call.Actions[0].Value
	if value == nil || value.Kind != operandSymbol || value.Record != captured || value.Token == nil || value.Token.binding == nil || value.Token.binding.Record != captured {
		t.Fatal("expanded Show rebound the captured symbol outside its definition")
	}
	if definitionFile == "local-words.l" {
		found := false
		for _, path := range compilation.SourcePaths {
			found = found || path == declaration.source.Span.Path
		}
		if !found {
			t.Fatal("imported action lost its source inventory entry")
		}
	}
	return captured, call
}

func TestParagraphCaptureStaticDefinitionSurvivesCallerShadow(t *testing.T) {
	for _, imported := range []bool{false, true} {
		for _, aliased := range []bool{false, true} {
			t.Run(fmt.Sprintf("imported=%t/alias=%t", imported, aliased), func(t *testing.T) {
				source := paragraphSource(t)
				definition := "    Define \"privateText\" as `definition text`.\n" + paragraphCaptureDefinition + "\n"
				source.main = definition
				definitionFile := "main.l"
				if imported {
					source.library = definition
					source.main = "Use `local-words.l`.\n"
					definitionFile = "local-words.l"
				}
				source.main += "\n    Define \"privateText\" as `caller shadow`.\n"
				verb := "Report"
				if aliased {
					source.main += "Use Report as \"Relay\".\n"
					verb = "Relay"
				}
				invocation := verb + " `argument text`."
				source.main += invocation + "\nShow privateText.\n"
				result := executableTestResult(t, source)
				captured, call := paragraphCaptureCall(t, result.Compilation, definitionFile, invocation)
				if captured.Value.Primitive == nil || captured.Value.Primitive.Text != "`definition text`" {
					t.Fatal("definition-site Text was replaced by the caller's shadow")
				}
				var caller *record
				for _, statement := range result.Compilation.Statements {
					if strings.TrimSpace(string(statement.source.Raw)) == "Show privateText." {
						for _, tok := range statement.tokens {
							if string(tok.rawBytes) == "privateText" && tok.binding != nil {
								caller = tok.binding.Record
							}
						}
					}
				}
				if caller == nil || caller == captured || caller.Value.Primitive == nil || caller.Value.Primitive.Text != "`caller shadow`" {
					t.Fatal("caller shadow control did not retain independent canonical storage")
				}
				if aliased {
					canonical, err := canonicalRecord(call.Target.Token.binding.Record, result.Compilation.Semantics.Aliases)
					if err != nil || canonical != result.Compilation.Memory["Report"] || call.Target.Token.binding.Record == canonical {
						t.Fatalf("aliased invocation did not preserve its distinct alias and canonical action: %v", err)
					}
				}
				want := []string{"show:definition text", "show:argument text", "show:caller shadow"}
				if actual := paragraphActivationTrace(t, result); !reflect.DeepEqual(actual, want) {
					t.Fatalf("captured/argument/caller output = %v; want %v", actual, want)
				}
			})
		}
	}
}

func TestParagraphCaptureImportedRuntimeRequiresActiveStorage(t *testing.T) {
	for _, tc := range []struct {
		name, alias, verb string
		escape            bool
	}{
		{"direct active", "", "Report", false},
		{"alias active", "Use Report as \"Relay\".\n", "Relay", false},
		{"alias escapes", "Use Report as \"Relay\".\n", "Relay", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := paragraphSource(t)
			source.main = "Use IO's Print from `io` as \"Print\".\n" +
				"\"message\" is `worker text`.\nCreate \"Worker\" as a New Basic App.\nThe App Prints message.\n" +
				"\n    Run Worker.\nUse `local-words.l`.\n"
			source.library = "Define \"privateText\" as Worker's Output.\n" + paragraphCaptureDefinition + "\n"
			if tc.escape {
				source.main += "\n"
			}
			invocation := tc.verb + " `argument text`."
			source.main += tc.alias + invocation + "\n"
			result := executableTestResult(t, source)
			captured, call := paragraphCaptureCall(t, result.Compilation, "local-words.l", invocation)
			if captured.Value.Runtime == nil || captured.Value.Runtime.Owner != result.Compilation.Memory["Worker"] || !call.Actions[0].Value.Runtime {
				t.Fatal("imported runtime capture lost its canonical Output storage")
			}
			if !tc.escape {
				want := []string{"run:Worker", "output:Worker", "show:saved:Worker", "show:argument text"}
				if actual := paragraphActivationTrace(t, result); !reflect.DeepEqual(actual, want) {
					t.Fatalf("active definition-site storage = %v; want %v", actual, want)
				}
				return
			}
			_, err := result.OrchestratorUnit("entry", executableTestCompanions(result))
			if err == nil {
				t.Fatal("aliased imported action allowed private runtime storage to escape")
			}
			message := err.Error()
			invocationAt := fmt.Sprintf("main.l:%d:", call.Statement.source.Span.Start.Line)

			if !strings.Contains(message, "private runtime capture privateText") || !strings.Contains(message, invocationAt) {
				t.Fatalf("capture rejection lost its cause or invocation context: %v", err)
			}
		})
	}
}
