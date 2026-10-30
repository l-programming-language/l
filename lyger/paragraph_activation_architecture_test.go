package Lyger

import (
	"strings"
	"testing"
)

func TestParagraphActivationFirstRunExitRetainsNestedInvocation(t *testing.T) {
	source := paragraphSource(t)
	source.main = paragraphGlobals +
		"Create \"Launch|-es|-ed|-ing\" for Building any App and Running it and then Awaiting it.\n" +
		"Create \"Dispatch|-es|-ed|-ing\" for Launching any App and then Showing `pending action`.\n" +
		"Create \"Second\" as a New Basic App.\nThe App Prints x2.\n" +
		"\n    Create \"First\" as a New Public Basic App.\nThe App Prints x1.\nExit.\n" +
		"\nDispatch First.\nRun Second.\nShow `pending outer`.\n"
	result := executableTestResult(t, source)
	var invocation *lStatement
	for _, statement := range result.Compilation.Statements {
		if strings.TrimSpace(string(statement.source.Raw)) == "Dispatch First." {
			invocation = statement
		}
	}
	if invocation == nil {
		t.Fatal("outer source invocation disappeared")
	}
	var operations int
	var inspect func([]*instruction)
	inspect = func(instructions []*instruction) {
		for _, current := range instructions {
			if current.Kind == instructionRun && current.CallSite != nil {
				operations++
				if current.CallSite != invocation {
					t.Errorf("nested Run diagnostics point to %s:%d, want outer invocation at line %d", current.CallSite.source.Span.Path, current.CallSite.source.Span.Start.Line, invocation.source.Span.Start.Line)
				}
				if current.Statement == nil || !strings.HasPrefix(strings.TrimSpace(string(current.Statement.source.Raw)), "Create \"Launch") {
					t.Error("nested Run lost its own definition-site attribution")
				}
			}
			inspect(current.Actions)
			inspect(current.Then)
			inspect(current.Else)
			if current.Block != nil {
				inspect(current.Block.Body)
			}
		}
	}
	inspect(result.Compilation.Semantics.Instructions)
	if operations != 1 {
		t.Fatalf("nested Run operations = %d; want one source-expanded Run", operations)
	}
}

func TestParagraphActivationRetainsEscapingPublicRuntimeCapture(t *testing.T) {
	source := paragraphSource(t)
	const declaration = "Publicly create \"Report|-s|-ed|-ing\" for Showing privateText and then Showing any Text."
	const invocation = "Report `caller`."
	makeSource := func(value string) string {
		return paragraphGlobals + "\n    Create \"Witness\" as a New Public Basic App.\nThe App Prints x1.\n" +
			"Define \"privateText\" as " + value + ".\n" + declaration + "\n\nRun Witness.\nAwait Witness.\n" + invocation + "\n"
	}
	source.main = makeSource("Witness's Output")
	runtimeResult := executableTestResult(t, source)
	plan := requireOrchestrationPlan(t, runtimeResult.Compilation.Semantics)
	var boundInContinuation bool
	var invocationPlan *orchestrationInstruction
	visitCapturePlan(plan.Root.Body, func(current *orchestrationInstruction) {
		if current.Continuation != nil {
			for _, effect := range current.Continuation.Body {
				boundInContinuation = boundInContinuation || effect.Source.Kind == instructionBindAction
			}
		}
		if current.Source.Kind == instructionCall && current.Source.Callable == runtimeResult.Compilation.Memory["Report"] {
			invocationPlan = current
		}
	})
	if !boundInContinuation || invocationPlan == nil || invocationPlan.Actions[0].Value.Storage.Capture == nil {
		t.Fatal("Await did not create and preserve its public action environment")
	}
}
