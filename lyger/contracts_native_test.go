package Lyger

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/l-programming-language/l/contracts"
)

const contractNativeUpdateSource = "Create \"items\" as a New List of 1 and 2.\nPlace 4 in items at 0.\nPlace 5 in items at 1.\n"

func TestContractsNativeUsesPreserveSelectedUpdateAndSnapshots(t *testing.T) {
	service := NewLygerService(Options{})
	analysis := analyzeContractSource(t, service, []byte(contractNativeUpdateSource), "/project/native-uses.l", ContractAnalysisOptions{})
	var updates []contracts.NativeUse
	for _, use := range analysis.NativeUses {
		if use.Name == "Storage.List.Update" {
			updates = append(updates, use)
		}
	}
	if len(updates) != 2 {
		t.Fatalf("expected actual repeated update invocations, got %+v", analysis.NativeUses)
	}
	first, second := updates[0], updates[1]
	if first.Meaning != second.Meaning || first.ID == second.ID || first.Occurrence == second.Occurrence {
		t.Fatal("canonical selection and distinct applications collapsed")
	}
	if first.Target.Unknown || first.Signature == nil || !reflect.DeepEqual(first.Signature.Inputs, []string{"&mut LList", "f64", "LValue"}) || !reflect.DeepEqual(first.Signature.Outputs, []string{"runtime::Result<()>"}) {
		t.Fatalf("missing compiler-derived signature/target: %+v", first)
	}
	if first.Resources["List"] == first.Resources["destination"] || first.Resources["List"] == second.Resources["List"] || first.Resources["destination"] != second.Resources["destination"] {
		t.Fatal("temporary snapshots or final destination lost their runtime identities")
	}
	if first.Values["Index"].Number != 0 || second.Values["Index"].Number != 1 || first.Values["Value"].Number != 4 || second.Values["Value"].Number != 5 {
		t.Fatalf("typed literal substitutions missing: %+v / %+v", first.Values, second.Values)
	}
	if first.Implementation == "" || len(first.Dependencies) == 0 || first.OutcomeDisposition != "value" {
		t.Fatal("selected implementation or checked-return boundary missing")
	}
	for _, use := range updates {
		found := false
		checked := false
		contractWalk(analysis.Plan.Entry, func(node contracts.Node) {
			if node.Native == use.ID {
				found = node.Meaning == use.Meaning && node.ID == use.ID && reflect.DeepEqual(node.Target, use.Target)
			}
			if node.Kind == contracts.CheckedResult && len(node.Children) == 1 && node.Children[0].Native == use.ID && len(node.Cleanup) == 1 && node.Cleanup[0].Kind == contracts.Native {
				checked = true
			}
		})
		if !found || !checked {
			t.Fatal("native application is disconnected from actual typed plan")
		}
	}
	ordinary, err := service.Compile([]byte(contractNativeUpdateSource), "/project/native-uses.l")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(ordinary.AML, analysis.Compilation.AML) || !bytes.Equal(ordinary.BackendAML, analysis.Compilation.BackendAML) || !bytes.Equal(ordinary.ExpandedSource, analysis.Compilation.ExpandedSource) {
		t.Fatal("native analysis changed ordinary source compilation")
	}
}

func TestContractsNativeScopeSelectorsAndDeterminism(t *testing.T) {
	service := NewLygerService(Options{})
	var previous []byte
	for i := 0; i < 3; i++ {
		analysis := analyzeContractSource(t, service, []byte(contractNativeUpdateSource), "/project/native-uses.l", ContractAnalysisOptions{})
		for _, scope := range analysis.Scopes {
			selected, err := analysis.FindContractScope(scope.ID)
			if err != nil || selected.NodeID != scope.NodeID || !scope.Complete {
				t.Fatalf("exact selector lost scope: %+v %v", scope, err)
			}
			if scope.Boundary == "native_body" && len(scope.Exclusions) == 0 {
				t.Fatal("native body scope concealed caller exclusions")
			}
		}
		if _, err := analysis.FindContractScope("absent"); err == nil {
			t.Fatal("missing scope silently selected another operation")
		}
		encoded := contractJSON(t, struct {
			Uses   []contracts.NativeUse
			Scopes []contracts.AnalysisScope
		}{analysis.NativeUses, analysis.Scopes})
		if i > 0 && !bytes.Equal(previous, encoded) {
			t.Fatal("native records depend on map order or incidental identities")
		}
		previous = encoded
	}
}

func TestContractsNativeBuiltinSelectionResistsCallerShadowing(t *testing.T) {
	source := "Define \"Update\" to `<DEF>E{|Statement: ~caller_only~|}<DEF>`.\n" + contractNativeUpdateSource
	analysis := analyzeContractSource(t, NewLygerService(Options{}), []byte(source), "/project/shadow.l", ContractAnalysisOptions{})
	rec, err := analysis.Compilation.Compilation.builtinRecord("Storage.List.Update")
	if err != nil {
		t.Fatal(err)
	}
	want := analysis.Compilation.Compilation.contractRecorder.meaning(rec, nil)
	for _, use := range analysis.NativeUses {
		if use.Name == "Storage.List.Update" && (use.Meaning != want || !strings.HasPrefix(analysis.Compilation.Compilation.contractRecorder.definitionOrigin(rec).Source, "resource:")) {
			t.Fatal("caller shadow captured selected dictionary helper")
		}
	}
}

func TestContractsNativeUpdateRecordsEverySelectedHelper(t *testing.T) {
	analysis := analyzeContractSource(t, NewLygerService(Options{}), []byte(contractNativeUpdateSource), "/project/native-uses.l", ContractAnalysisOptions{})
	byID := make(map[string]contracts.NativeUse)
	for _, use := range analysis.NativeUses {
		byID[use.ID] = use
	}
	var operations int
	for _, scope := range analysis.Scopes {
		if scope.Boundary != "source_operation" {
			continue
		}
		operations++
		var names []string
		copyAt, updateAt, publicationAt := -1, -1, -1
		for i, id := range scope.NativeUses {
			use := byID[id]
			names = append(names, use.Name)
			if use.Signature == nil || use.Implementation == "" || len(use.Dependencies) == 0 {
				t.Fatalf("selected helper lacks an explicit interface or implementation: %+v", use)
			}
			switch use.Name {
			case "Storage.List.Copy":
				copyAt = i
			case "Storage.List.Update":
				updateAt = i
			case "Storage.Publish", "Storage.Assign":
				publicationAt = i
			}
		}
		if copyAt < 0 || copyAt >= updateAt || updateAt >= publicationAt {
			t.Fatalf("copy, checked mutation and publication lost authored order: %v", names)
		}
	}
	if operations != 2 {
		t.Fatalf("expected two exact statement scopes, got %d", operations)
	}
	before := contractJSON(t, analysis.NativeUses)
	if _, err := analysis.Compilation.OrchestratorUnit("late_entry", nil); err != nil {
		t.Fatal(err)
	}
	if err := analysis.RefreshLoweringInventory(); err != nil {
		t.Fatal(err)
	}
	after := contractJSON(t, analysis.NativeUses)
	if !bytes.Equal(before, after) {
		t.Fatal("late emission changed or duplicated native occurrences")
	}
}

func TestContractsNativeInvocationIncludesEvaluationAndCleanup(t *testing.T) {
	source := functionSource("Identity", "\"Value\" as Number", "Number", "Return Value.") + "Identity 7 as \"Result\".\n"
	analysis := analyzeContractSource(t, NewLygerService(Options{}), []byte(source), "/project/invocation.l", ContractAnalysisOptions{})
	uses := make(map[string]contracts.NativeUse)
	for _, use := range analysis.NativeUses {
		uses[use.ID] = use
	}
	var selected *contracts.AnalysisScope
	for _, scope := range analysis.Scopes {
		if scope.Boundary == "source_invocation" {
			copy := scope
			selected = &copy
		}
	}
	if selected == nil {
		t.Fatal("actual source invocation has no scope")
	}
	var literal, invoke, destruction, overhead bool
	for _, id := range selected.NativeUses {
		use := uses[id]
		if use.Signature == nil || use.Implementation == "" {
			t.Fatalf("unlinked invocation use: %+v", use)
		}
		literal = literal || use.Name == "Storage.Literal"
		invoke = invoke || use.Name == "Storage.Invoke"
		overhead = overhead || use.Boundary == "source_body_lowering_overhead"
		if use.Name == "Storage.Invoke" && use.Boundary != "invocation_composition_overhead" {
			t.Fatal("invocation wrapper could duplicate independently modeled callee costs")
		}
		if use.Name == "Storage.Function.Exit" {
			t.Fatal("an alternative composed inside Storage.Invoke was treated as eager execution")
		}
		destruction = destruction || use.Phase == contracts.Destruction && use.Name == "representation-destruction:Function"
	}
	if !literal || !invoke || !destruction || !overhead {
		t.Fatalf("scope omitted evaluation, selected invocation or callee final-owner cleanup: %+v", selected)
	}
	withoutTrust, err := contracts.InferScope(analysis.Plan, selected.NodeID, contracts.AnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if contracts.AllEstablished(contracts.CheckProfile(withoutTrust, contracts.Profile{Name: "NoBlocking"})) {
		t.Fatal("authored source incorrectly certified opaque invocation or destruction")
	}
}

func TestContractsNativeInvocationsHaveDistinctCalleeResources(t *testing.T) {
	source := functionSource("Identity", "\"Value\" as Number", "Number", "Return Value.") + "Identity 7 as \"First\".\nIdentity 9 as \"Second\".\n"
	analysis := analyzeContractSource(t, NewLygerService(Options{}), []byte(source), "/project/instances.l", ContractAnalysisOptions{})
	uses := make(map[string]contracts.NativeUse)
	for _, use := range analysis.NativeUses {
		uses[use.ID] = use
	}
	var releases []contracts.NativeUse
	for _, scope := range analysis.Scopes {
		if scope.Boundary != "source_invocation" {
			continue
		}
		for _, id := range scope.NativeUses {
			if use := uses[id]; use.Phase == contracts.Destruction {
				releases = append(releases, use)
			}
		}
	}
	if len(releases) != 2 {
		t.Fatalf("expected actual destruction for both invocations, got %d", len(releases))
	}
	first, second := releases[0], releases[1]
	if first.Meaning != second.Meaning || first.ID == second.ID || first.Occurrence == second.Occurrence || reflect.DeepEqual(first.Target, second.Target) || reflect.DeepEqual(first.Resources, second.Resources) {
		t.Fatal("same Function's invocations reused occurrence, target or owned resource identities")
	}
	if first.Dependencies["source-invocation"] == second.Dependencies["source-invocation"] {
		t.Fatal("implementation evidence is not bound to the current invocation")
	}
}

func TestContractsNativeReturnErrorIsValueAndStopsBody(t *testing.T) {
	source := functionSource("EchoFailure", "\"Value\" as Error", "Error", "Return Value.") + "Create \"problem\" as a New Error of `ordinary value`.\nEchoFailure problem as \"Result\".\n"
	analysis := analyzeContractSource(t, NewLygerService(Options{}), []byte(source), "/project/error-value.l", ContractAnalysisOptions{})
	foundReturn := false
	for _, definition := range analysis.Plan.Definitions {
		contractWalk(definition, func(node contracts.Node) {
			if node.Kind == contracts.Return {
				foundReturn = true
				if node.Outcome != "" {
					t.Fatal("ordinary Error return acquired an abrupt failure channel")
				}
			}
		})
	}
	if !foundReturn {
		t.Fatal("source Return was reduced to value evaluation without control")
	}
	for _, path := range analysis.Summary.Paths {
		if path.Outcome == contracts.ReturnedFailure || path.ControlReturn {
			t.Fatal("returned Error escaped its source Function boundary as failure/control")
		}
	}
}
