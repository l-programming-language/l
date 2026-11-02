package Lyger

import (
	"bytes"
	"testing"

	"github.com/l-programming-language/l/contracts"
)

func contractWalk(node contracts.Node, visit func(contracts.Node)) {
	visit(node)
	for _, children := range [][]contracts.Node{node.Children, node.Then, node.Else, node.Cleanup} {
		for _, child := range children {
			contractWalk(child, visit)
		}
	}
}

func TestContractsClosureSurvivesTerminalSource(t *testing.T) {
	analysis, err := NewLygerService(Options{}).AnalyzeContracts([]byte("Exit.\n"), "/project/exit.l", ContractAnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, boundary := range analysis.Summary.NativeBoundaries {
		found = found || boundary.ID == "backend-support-pending"
	}
	if !found || len(analysis.Plan.NativeClosure) == 0 {
		t.Fatal("terminal source erased backend closure obligations")
	}
	for _, profile := range []string{"NoHeapAllocation", "NoBlocking", "NoExternalIO", "NoPanic"} {
		if contracts.AllEstablished(contracts.CheckProfile(analysis.Summary, contracts.Profile{Name: profile})) {
			t.Fatal("terminal path bypassed native closure: " + profile)
		}
	}
}

func TestContractsRefreshIncludesFinalCoordinatorHelpers(t *testing.T) {
	analysis, err := NewLygerService(Options{}).AnalyzeContracts([]byte("Exit.\n"), "/project/exit.l", ContractAnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	before := len(analysis.Lowering.Steps)
	if _, err := analysis.Compilation.OrchestratorUnit("selected_final_entry", nil); err != nil {
		t.Fatal(err)
	}
	if err := analysis.RefreshLoweringInventory(); err != nil {
		t.Fatal(err)
	}
	if len(analysis.Lowering.Steps) <= before {
		t.Fatal("final coordinator helper selections were not included")
	}
	if err := analysis.RefreshLoweringInventory(); err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, boundary := range analysis.Plan.NativeClosure {
		if boundary.ID == "backend-support-pending" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("refresh lost or duplicated independent backend closure: %d", count)
	}
	if _, err := contracts.Infer(analysis.Plan, contracts.AnalysisOptions{}); err != nil {
		t.Fatal(err)
	}
}

func TestContractsSourceLifecycleTransitionsRetainOwnership(t *testing.T) {
	analysis, err := NewLygerService(Options{}).AnalyzeContracts([]byte("Create \"Worker\" as a New Basic App.\nRun Worker.\nStop Worker.\nAwait Worker.\n"), "/project/lifecycle.l", ContractAnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	joined, unjoined, startupFailure := false, false, false
	roots := make(map[string]bool)
	for _, state := range analysis.Lifecycle {
		if state.App.Unknown || state.App.Root == "" || state.Owner == "" {
			t.Fatalf("source activity lost canonical supervisor: %+v", state)
		}
		roots[state.App.Root] = true
		joined = joined || state.Joined && state.Claim == contracts.ClaimFinished && !state.Outstanding
		unjoined = unjoined || state.Outstanding && !state.Joined
		startupFailure = startupFailure || state.StartupFailure
	}
	if len(roots) != 1 || !joined || !unjoined || !startupFailure {
		t.Fatalf("lost continuation, cancellation or startup alternative: %+v", analysis.Lifecycle)
	}
	operations := make(map[string]bool)
	for _, obligation := range analysis.LifecycleObligations {
		if obligation.Claim.Predicate.Kind == "protocol" {
			operations[obligation.Claim.Predicate.Subject] = true
			if obligation.Status != contracts.Established {
				t.Fatalf("valid typed operation not admitted: %+v", obligation)
			}
		}
	}
	closure := false
	for _, obligation := range analysis.CheckLifecycleProfile(contracts.Profile{Name: "OwnedLifecycle", Scope: "whole_region"}) {
		closure = closure || obligation.Claim.Predicate.Kind == "lifecycle_closed" && obligation.Status == contracts.Unresolved
	}
	if !operations["Run"] || !operations["Stop"] || !operations["Await"] || !closure {
		t.Fatal("source transitions or independent native closure were omitted")
	}
}

func TestContractsSourceLifecycleDoubleRunKeepsGuard(t *testing.T) {
	analysis, err := NewLygerService(Options{}).AnalyzeContracts([]byte("Create \"Worker\" as a New Basic App.\nRun Worker.\nRun Worker.\n"), "/project/repeat-run.l", ContractAnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	failed := false
	for _, obligation := range analysis.LifecycleObligations {
		failed = failed || obligation.Claim.Predicate.Subject == "Run" && obligation.Status == contracts.Contradicted
	}
	if !failed {
		t.Fatal("external completion incorrectly released an outstanding invocation")
	}
}

func TestContractsBASE058BASE059PointerNativeClosure(t *testing.T) {
	var source string
	for _, fixture := range runtimeValueSources {
		if fixture.name == "pointer sharing" {
			source = fixture.source
		}
	}
	analysis, err := NewLygerService(Options{}).AnalyzeContracts([]byte(source), "/project/pointer.l", ContractAnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assertContractSelectedNative(t, analysis, "Storage.Pointer.Copy")
	assertContractSelectedNative(t, analysis, "Storage.Pointer.Update")
	assertContractSelectedNative(t, analysis, "Storage.Value.Protect")
	// Shared payload identity establishes neither zero metadata copying nor
	// allocation-free cycle checking. These are separate unapproved quantities.
	if analysis.Summary.Resources.MetadataCopied.Kind != "unknown" || analysis.Summary.Resources.PayloadCopied.Kind != "unknown" || analysis.Summary.Resources.Allocated.Kind != "unknown" {
		t.Fatalf("opaque Pointer metadata/cycle cost was silently treated as zero: %+v", analysis.Summary.Resources)
	}
	if contracts.AllEstablished(contracts.CheckProfile(analysis.Summary, contracts.Profile{Name: "NoHeapAllocation"})) {
		t.Fatal("shared Pointer target falsely established allocation-free copying/update")
	}
}

func assertContractSelectedNative(t *testing.T, analysis *ContractAnalysis, name string) contracts.MeaningKey {
	t.Helper()
	c := analysis.Compilation.Compilation
	rec, err := c.builtinRecord(name)
	if err != nil {
		t.Fatal(err)
	}
	meaning := c.contractRecorder.meaning(rec, nil)
	selected, unresolved := false, false
	for _, step := range analysis.Lowering.Steps {
		selected = selected || step.Meaning == meaning
	}
	for _, boundary := range analysis.Summary.NativeBoundaries {
		unresolved = unresolved || boundary.Meaning == meaning && boundary.Status == contracts.Unresolved
	}
	if !selected || !unresolved {
		t.Fatalf("canonical selected %s (%s) omitted from unapproved closure; selected=%t unresolved=%t", name, meaning, selected, unresolved)
	}
	return meaning
}

func TestContractsSevenPilotSelectionsRemainUnapproved(t *testing.T) {
	source := distributionIOImport + "Create \"Worker\" as a New Basic App.\nThe App Prints `pilot`.\nStart Worker.\nAwait Worker.\nExecute Worker.\n"
	for _, fixture := range runtimeValueSources {
		if fixture.name == "list snapshot" || fixture.name == "pointer sharing" {
			source += "\n    " + fixture.source
		}
	}
	analysis, err := NewLygerService(Options{}).AnalyzeContracts([]byte(source), "/project/pilots.l", ContractAnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"Storage.Require", "Storage.Pointer.Copy", "Storage.Pointer.Update", "Storage.Traverse"} {
		assertContractSelectedNative(t, analysis, name)
	}
	for _, name := range []string{"IO.Print", "Start", "Execute", "Main", "Basic"} {
		found := false
		for _, binding := range analysis.Bindings.Bindings {
			found = found || binding.CanonicalName == name
		}
		if !found {
			t.Errorf("pilot canonical selection %s missing from Track A bindings", name)
		}
	}
	for _, profile := range []string{"NoHeapAllocation", "NoBlocking", "NoPanic"} {
		if contracts.AllEstablished(contracts.CheckProfile(analysis.Summary, contracts.Profile{Name: profile})) {
			t.Fatalf("unapproved pilot closure falsely established %s", profile)
		}
	}
	destruction := false
	for _, boundary := range analysis.Summary.NativeBoundaries {
		destruction = destruction || boundary.Implicit && boundary.Status == contracts.Unresolved
	}
	if !destruction {
		t.Fatal("pilot report omitted implicit cleanup/native closure")
	}
}

func TestContractsCanonicalSameNamedAppTargets(t *testing.T) {
	source := "Create \"Worker\" as a New Basic App.\nRun Worker.\nAwait Worker.\n\n    Create \"Worker\" as a New Basic App.\nRun Worker.\nAwait Worker.\n"
	analysis, err := NewLygerService(Options{}).AnalyzeContracts([]byte(source), "/project/targets.l", ContractAnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	roots := make(map[string]bool)
	// The second source paragraph is a retained App continuation, not another
	// outer activation. Exercise the adapter using both actual canonical Apps.
	var operations []*orchestrationInstruction
	for _, construction := range analysis.Compilation.Compilation.Semantics.Constructions {
		operations = append(operations, &orchestrationInstruction{Source: &instruction{Kind: instructionRun}, App: construction.App})
	}
	projected := contracts.Node{Kind: contracts.Sequence, Children: analysis.Compilation.Compilation.contractRecorder.contractInstructions(operations)}
	contractWalk(projected, func(node contracts.Node) {
		for _, effect := range node.Effects {
			if effect.Kind == "spawn" {
				if effect.Target.Unknown || effect.Target.Root == "" {
					t.Fatal("resolved App lost canonical target")
				}
				roots[effect.Target.Root] = true
			}
		}
	})
	if len(roots) != 2 {
		t.Fatalf("same-spelled Apps collapsed: %v", roots)
	}
	for root := range roots {
		found := false
		for _, target := range analysis.Bindings.Targets {
			found = found || target.Root == root
		}
		if !found {
			t.Fatal("lifecycle target differs from construction target")
		}
	}
}

func TestContractsBASE025SnapshotAndPointerProjection(t *testing.T) {
	for _, name := range []string{"list snapshot", "pointer sharing"} {
		t.Run(name, func(t *testing.T) {
			var source string
			for _, fixture := range runtimeValueSources {
				if fixture.name == name {
					source = fixture.source
				}
			}
			service := NewLygerService(Options{})
			ordinary, err := service.Compile([]byte(source), "/project/values.l")
			if err != nil {
				t.Fatal(err)
			}
			analysis, err := service.AnalyzeContracts([]byte(source), "/project/values.l", ContractAnalysisOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(ordinary.BackendAML, analysis.Compilation.BackendAML) || !bytes.Equal(ordinary.ExpandedSource, analysis.Compilation.ExpandedSource) {
				t.Fatal("contract projection changed source semantics")
			}
			want := "logical_snapshot"
			if name == "pointer sharing" {
				want = "shared_pointer_target"
			}
			found := false
			release := false
			contractWalk(analysis.Plan.Entry, func(node contracts.Node) {
				for _, fact := range node.Facts {
					found = found || fact.Kind == "ownership" && fact.Value.Text == want
				}
				release = release || node.Kind == contracts.Acquire && len(node.Cleanup) > 0
			})
			if !found || !release {
				t.Fatalf("missing semantic transfer %s or release closure", want)
			}
			if contracts.AllEstablished(contracts.CheckProfile(analysis.Summary, contracts.Profile{Name: "NoHeapAllocation"})) {
				t.Fatal("representation metadata became native proof")
			}
		})
	}
}

func TestContractsSourceFunctionBodyProjection(t *testing.T) {
	source := functionSource("Echo", "\"Value\" as Text", "Text", "Show Value.\nReturn Value.") + "Echo `hello` as \"Result\".\nShow Result.\n"
	analysis, err := NewLygerService(Options{}).AnalyzeContracts([]byte(source), "/project/function.l", ContractAnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(analysis.Plan.Definitions) == 0 {
		t.Fatal("source Function body omitted")
	}
	called := false
	bodyIO := false
	contractWalk(analysis.Plan.Entry, func(node contracts.Node) { called = called || node.Kind == contracts.Call })
	for _, definition := range analysis.Plan.Definitions {
		contractWalk(definition, func(node contracts.Node) {
			for _, effect := range node.Effects {
				bodyIO = bodyIO || effect.Kind == "io"
			}
		})
	}
	if !called || !bodyIO {
		t.Fatal("known Function invocation and typed body were not connected")
	}
	if contracts.AllEstablished(contracts.CheckProfile(analysis.Summary, contracts.Profile{Name: "NoBlocking"})) {
		t.Fatal("typed function body erased unapproved capture/lowering effects")
	}
}

func TestContractsActivationStorageTargets(t *testing.T) {
	comp, _, err := NewLygerService(Options{}).ResolveContractBindings([]byte("Define \"value\" as `text`.\n"), "/project/storage.l")
	if err != nil {
		t.Fatal(err)
	}
	rec := comp.Memory["value"]
	if rec == nil {
		t.Fatal("missing source value")
	}
	first := &orchestrationStorage{ID: 1, Record: rec, Type: valueText, Owner: &orchestrationBlock{ID: 1}}
	second := &orchestrationStorage{ID: 2, Record: rec, Type: valueText, Owner: &orchestrationBlock{ID: 2}}
	a, b := comp.contractRecorder.contractStorageTarget(first), comp.contractRecorder.contractStorageTarget(second)
	if a.Root != b.Root || a.Generation == b.Generation || contractStorageResource(a).ID == contractStorageResource(b).ID {
		t.Fatal("canonical binding and activation storage identity conflated")
	}
}
