package Lyger

import (
	"strings"
	"testing"

	"github.com/l-programming-language/l/lerrors"
)

func TestAppBodyRejectsRuntimeCaptures(t *testing.T) {
	for _, literal := range []string{"`conditional payload`", "7", "true"} {
		for _, condition := range []string{"false", "true"} {
			for _, alias := range []bool{false, true} {
				name, aliases := "message", ""
				if alias {
					name, aliases = "copy", "Use message as \"first\".\nUse first as \"copy\".\n"
				}
				t.Run(literal+"/"+condition+"/"+name, func(t *testing.T) {
					source := distributionIOImport + "If " + condition + ", define \"message\" as " + literal + ".\n" + aliases +
						"Create \"Worker\" as a New Basic App.\nThe App Prints " + name + ".\nExecute Worker.\n"
					distributionNegative(t, source, "runtime value message cannot be captured by an App body", lerrors.Unsupported)
				})
			}
		}
	}
	// IO.Print itself composes Process and its private typed Print definitions. An
	// alias must retain that composition and the original value's identity.
	distributionNegative(t, distributionIOImport+"Use Print as \"Display\".\nIf false, define \"message\" as `payload`.\nCreate \"Worker\" as a New Basic App.\nThe App Displays message.\n", "runtime value message cannot be captured by an App body", lerrors.Unsupported)
}

func TestAppBodyRetainsStaticCaptures(t *testing.T) {
	for _, declaration := range []string{"\"message\" is `static payload`.\n", "Define \"message\" as `static payload`.\n"} {
		source := distributionIOImport + declaration + "Use message as \"copy\".\nCreate \"Worker\" as a New Basic App.\nThe App Prints copy.\nExecute Worker.\n"
		result := requireDistributionFrontend(t, source, nil)
		if !strings.Contains(string(result.BackendAML), "static payload") {
			t.Fatal("static App capture lost its explicit source value")
		}
		if _, err := result.OrchestratorUnit("entry", executableTestCompanions(result)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRuntimeRecordsFollowExpandedActionsAndAliases(t *testing.T) {
	rec := planTestScalar("value", valueText, nil)
	alias := planTestScalar("alias", valueText, nil)
	assign := &instruction{Kind: instructionAssign, Target: planTestRef(alias)}
	for _, body := range []*instruction{
		{Kind: instructionCall, Actions: []*instruction{assign}},
		{Kind: instructionIf, Then: []*instruction{assign}},
		{Kind: instructionIf, Else: []*instruction{assign}},
		{Block: &scopedInstructions{Body: []*instruction{assign}}},
	} {
		semantics := newSemanticProgram()
		semantics.Aliases[alias] = rec
		semantics.Instructions = []*instruction{body}
		dynamic, err := semantics.runtimeRecords()
		if err != nil || len(dynamic) != 1 || !dynamic[rec] {
			t.Fatalf("expanded assignment lost canonical identity: %v, %v", dynamic, err)
		}
		// A composed App action recursively lowers its bound definitions. The
		// same classification must reject a runtime record before templating.
		c := &Compilation{Semantics: semantics}
		if _, err := c.instantiateCall(alias, nil, make(map[*record]bool)); err == nil || !strings.Contains(err.Error(), "runtime value value cannot be captured") {
			t.Fatalf("runtime definition accepted for source generation: %v", err)
		}
	}
	semantics := newSemanticProgram()
	semantics.Aliases[alias] = rec
	semantics.Continuations = map[*App]*scopedInstructions{{ID: "Worker"}: {Body: []*instruction{assign}}}
	if dynamic, err := semantics.runtimeRecords(); err != nil || !dynamic[rec] {
		t.Fatalf("continuation assignment lost: %v, %v", dynamic, err)
	}
}

func TestAppBodyRejectsConditionalComposedDefinition(t *testing.T) {
	source := "If false, define \"Log|-s|-ged|-ging|-ger\" to `<DEF>BOD^E{|Statement: ~::std::println!(\"{}\", <@Output|Thing@>);~|}<DEF>`.\n" +
		"Use Log as \"Emit\".\nCreate \"Speak|-s|-ed|-ing\" for Processing a New Emit.\nCreate \"Worker\" as a New Basic App.\nThe App Speaks `payload`.\n"
	distributionNegative(t, source, "runtime value Log cannot be captured by an App body", lerrors.Unsupported)
}
