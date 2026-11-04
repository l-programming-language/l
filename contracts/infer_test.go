package contracts

import (
	"strings"
	"testing"
)

func inferTest(t *testing.T, node Node, callbacks map[string]Node) DerivedSummary {
	t.Helper()
	summary, err := Infer(Plan{Entry: node, Callbacks: callbacks}, AnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	return summary
}
func sequenceTest(nodes ...Node) Node { return Node{Kind: Sequence, Children: nodes} }
func eventTest(kind EffectKind, phase Phase) Node {
	return Node{Kind: Event, Phase: phase, Effects: []Effect{{Kind: kind, Phase: phase, Definite: true, Origin: Origin{Source: "fixture.l", Label: string(kind)}}}}
}
func statusTest(t *testing.T, summary DerivedSummary, name string, want Status) {
	t.Helper()
	got := CheckProfile(summary, Profile{Name: name})
	if len(got) != 1 || got[0].Status != want {
		t.Fatalf("%s: got %+v, want %s", name, got, want)
	}
}
func factTest(subject string, value int64) Predicate {
	return Predicate{Kind: "eq", Subject: subject, Value: Atom{Kind: "number", Number: value}}
}

func TestInferenceBASE010SequenceFailure(t *testing.T) {
	fact := factTest("ready", 1)
	plan := Plan{Entry: sequenceTest(Node{Kind: Branch, Then: []Node{{Kind: Event, Facts: []Predicate{fact}}}, Else: []Node{{Kind: Event, Outcome: ReturnedFailure}}}, Node{Kind: Call, Meaning: "consumer"}), Definitions: map[MeaningKey]Node{"consumer": sequenceTest(Node{Kind: Require, Requires: []Predicate{fact}}, eventTest("io", Invocation))}}
	summary, err := Infer(plan, AnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(summary.Paths) != 2 {
		t.Fatalf("lost alternatives: %+v", summary.Paths)
	}
	for _, path := range summary.Paths {
		if path.Outcome == ReturnedFailure {
			if len(path.Effects) != 0 || len(path.Obligations) != 0 {
				t.Fatal("failed predecessor executed consumer")
			}
		} else if len(path.Obligations) != 1 || path.Obligations[0].Status != Established || len(path.Effects) != 1 {
			t.Fatalf("normal call did not use predecessor facts: %+v", path)
		}
	}
}

func TestInferenceBASE011BranchAlternatives(t *testing.T) {
	summary := inferTest(t, sequenceTest(Node{Kind: Branch, Then: []Node{{Kind: Event, Facts: []Predicate{factTest("x", 0)}}}, Else: []Node{{Kind: Event, Facts: []Predicate{factTest("x", 1)}}}}, Node{Kind: Require, Requires: []Predicate{factTest("x", 0)}}), nil)
	if len(summary.Paths) != 2 || summary.Paths[0].Obligations[0].Status != Established || summary.Paths[1].Obligations[0].Status == Established {
		t.Fatalf("branch facts collapsed: %+v", summary.Paths)
	}
}

func TestInferenceBASE012CallbackEffects(t *testing.T) {
	known := inferTest(t, Node{Kind: Callback, Callback: "action"}, map[string]Node{"action": eventTest("io", Invocation)})
	statusTest(t, known, "NoExternalIO", Contradicted)
	unknown := inferTest(t, Node{Kind: Callback, Callback: "opaque"}, nil)
	statusTest(t, unknown, "NoExternalIO", Unresolved)
	if unknown.Resources.Work.Kind != "unknown" {
		t.Fatal("unknown callback charged zero")
	}
}

func TestInferenceBASE013OutcomeDemand(t *testing.T) {
	branch := Node{Kind: Branch, Then: []Node{{Kind: Event}}, Else: []Node{{Kind: Event, Outcome: ReturnedFailure}}}
	eager := inferTest(t, sequenceTest(Node{Kind: Callback, Callback: "error"}, branch), map[string]Node{"error": {Kind: Event}})
	conditional := inferTest(t, Node{Kind: Branch, Then: []Node{{Kind: Event}}, Else: []Node{{Kind: Callback, Callback: "error"}, {Kind: Event, Outcome: ReturnedFailure}}}, map[string]Node{"error": {Kind: Event}})
	for _, path := range eager.Paths {
		if path.Demand["error"].Lower != 1 {
			t.Fatal("eager construction was skipped")
		}
	}
	for _, path := range conditional.Paths {
		want := int64(0)
		if path.Outcome == ReturnedFailure {
			want = 1
		}
		if path.Demand["error"].Lower != want {
			t.Fatal("conditional callback demand changed")
		}
	}
}

func TestInferenceBASE014Phases(t *testing.T) {
	for _, phase := range []Phase{Construction, Invocation, Destruction} {
		t.Run(string(phase), func(t *testing.T) {
			summary := inferTest(t, eventTest("allocate", phase), nil)
			statusTest(t, summary, "NoHeapAllocation", Contradicted)
			if summary.Effects[0].Phase != phase {
				t.Fatal("phase lost")
			}
		})
	}
}

func TestInferenceBASE015DemandChange(t *testing.T) {
	clean := inferTest(t, Node{Kind: Event}, nil)
	changed := inferTest(t, Node{Kind: Callback, Callback: "effect"}, map[string]Node{"effect": eventTest("allocate", Construction)})
	statusTest(t, clean, "NoHeapAllocation", Established)
	statusTest(t, changed, "NoHeapAllocation", Contradicted)
	if clean.Fingerprint == changed.Fingerprint {
		t.Fatal("demand change reused fingerprint")
	}
}

func TestInferenceBASE016Aliases(t *testing.T) {
	a, b, c := ResourceRef{ID: "a"}, ResourceRef{ID: "b"}, ResourceRef{ID: "c"}
	aliases := []ResourcePair{{Left: a, Right: b}}
	disjoint := []ResourcePair{{Left: b, Right: c}}
	if AliasRelation(a, b, aliases, disjoint) != Established || AliasRelation(a, c, aliases, disjoint) != Contradicted || AliasRelation(a, ResourceRef{ID: "external"}, aliases, disjoint) != Unresolved {
		t.Fatal("alias relation conflated names and resources")
	}
}

func TestInferenceBASE017NativeUnknown(t *testing.T) {
	summary := inferTest(t, Node{Kind: Native, Native: "opaque-helper", Phase: Invocation}, nil)
	statusTest(t, summary, "NoHeapAllocation", Unresolved)
	statusTest(t, summary, "NoBlocking", Unresolved)
	if len(summary.NativeBoundaries) != 1 || summary.NativeBoundaries[0].ID != "opaque-helper" {
		t.Fatal("native boundary missing")
	}
}

func TestInferenceBASE018UnsafeAcrossCallbackAndSuspend(t *testing.T) {
	acquire := Node{Kind: Acquire, Resource: ResourceRef{ID: "lock"}, Ownership: "guard"}
	for _, operation := range []Node{eventTest("suspend", Invocation), {Kind: Callback, Callback: "unknown"}} {
		summary := inferTest(t, sequenceTest(acquire, operation), nil)
		if len(summary.Obligations) == 0 || AllEstablished(summary.Obligations) {
			t.Fatal("unsafe guard crossing certified")
		}
	}
}

func TestInferenceBASE019DestructorBlocking(t *testing.T) {
	summary := inferTest(t, Node{Kind: Scope, Children: []Node{{Kind: Acquire, Resource: ResourceRef{ID: "guard"}, Cleanup: []Node{eventTest("block", Destruction)}}}}, nil)
	statusTest(t, summary, "NoBlocking", Contradicted)
	if CheckProfile(summary, Profile{Name: "NoBlocking", Scope: "body_only"})[0].Status != Established {
		t.Fatal("body-only scope unexpectedly includes destruction")
	}
}

func TestInferenceBASE020ConditionalFinalOwner(t *testing.T) {
	summary := inferTest(t, Node{Kind: Scope, Children: []Node{{Kind: Acquire, Resource: ResourceRef{ID: "shared"}, Ownership: "shared", Cleanup: []Node{eventTest("block", Destruction)}}}}, nil)
	statusTest(t, summary, "NoBlocking", Unresolved)
	found := false
	for _, effect := range summary.Effects {
		found = found || effect.Kind == "block" && strings.Contains(effect.Condition, "final owner")
	}
	if !found {
		t.Fatal("conditional destruction omitted")
	}
}

func TestInferenceBASE021ConditionalInitializationMoves(t *testing.T) {
	resource := ResourceRef{ID: "value"}
	acquire := Node{Kind: Acquire, Resource: resource, Cleanup: []Node{eventTest("io", Destruction)}}
	summary := inferTest(t, Node{Kind: Scope, Children: []Node{{Kind: Branch, Then: []Node{acquire}, Else: []Node{acquire, {Kind: Move, Resource: resource}}}}}, nil)
	counts := map[int]int{}
	for _, path := range summary.Paths {
		n := 0
		for _, effect := range path.Effects {
			if effect.Kind == "io" {
				n++
			}
		}
		counts[n]++
	}
	if counts[0] != 1 || counts[1] != 1 || len(counts) != 2 {
		t.Fatalf("release duplicated or lost: %v", counts)
	}
}

func TestInferenceBASE022AbortNoCleanup(t *testing.T) {
	summary := inferTest(t, Node{Kind: Scope, CleanupOutcomes: []Outcome{Normal, Abort}, Children: []Node{{Kind: Acquire, Resource: ResourceRef{ID: "value"}, Cleanup: []Node{eventTest("io", Destruction)}}, {Kind: Event, Outcome: Abort}}}, nil)
	if len(summary.Effects) != 0 || summary.Paths[0].Owned["value"].State != "owned" {
		t.Fatal("abort invented ordinary destruction")
	}
}

func TestInferenceBASE023SharedMutationAndBASE024Growth(t *testing.T) {
	for _, facts := range []MutationFacts{
		{CopyOnWrite: true, Sharing: Unresolved, CapacityEnough: Established, NestedClosed: true},
		{Sharing: Contradicted, CapacityEnough: Unresolved, NestedClosed: true},
		{Sharing: Contradicted, CapacityEnough: Established, NestedClosed: false},
	} {
		summary := inferTest(t, Node{Kind: Mutation, Resource: ResourceRef{ID: "snapshot"}, Phase: Invocation, MutationFacts: &facts}, nil)
		statusTest(t, summary, "NoHeapAllocation", Unresolved)
	}
	valid := MutationFacts{CopyOnWrite: true, Sharing: Contradicted, CapacityEnough: Established, NestedClosed: true}
	statusTest(t, inferTest(t, Node{Kind: Mutation, MutationFacts: &valid}, nil), "NoHeapAllocation", Established)
}

func TestInferenceBASE026DestructorCostsAndFailures(t *testing.T) {
	cost := ZeroResourceCost()
	cost.Work = ConstantCost(7)
	summary := inferTest(t, Node{Kind: Scope, Children: []Node{{Kind: Acquire, Resource: ResourceRef{ID: "nested"}, Cleanup: []Node{{Kind: Event, Phase: Destruction, Costs: &cost, Outcome: Panic}}}}}, nil)
	if value, ok := summary.Resources.Work.Constant(); !ok || value != 7 || summary.Paths[0].Outcome != Panic {
		t.Fatal("destruction cost or failure omitted")
	}
}

func TestInferenceBASE027ApprovedConditionalProfile(t *testing.T) {
	modeled := inferTest(t, Node{Kind: Event}, nil)
	plan := Plan{Entry: Node{Kind: Native, Native: "leaf"}}
	options := AnalysisOptions{NativeSummaries: map[string]DerivedSummary{"leaf": modeled}, NativeWitnesses: map[string]string{"leaf": "test-only:fixture"}}
	unapproved, err := Infer(plan, options)
	if err != nil {
		t.Fatal(err)
	}
	statusTest(t, unapproved, "NoHeapAllocation", Unresolved)
	options.ApprovedAssumptions = map[string]bool{"test-only:fixture": true}
	approved, err := Infer(plan, options)
	if err != nil {
		t.Fatal(err)
	}
	statusTest(t, approved, "NoHeapAllocation", Established)
	statusTest(t, approved, "NoBlocking", Established)
	if len(approved.Assumptions) != 1 || approved.Assumptions[0] != "test-only:fixture" {
		t.Fatal("assumption erased")
	}
}

func TestInferenceBASE030StaleGuard(t *testing.T) {
	resource := ResourceRef{ID: "state"}
	fact := factTest("valid", 1)
	fact.Resource = resource
	guard := Node{Kind: Guard, Facts: []Predicate{fact}}
	for _, mutate := range []bool{false, true} {
		nodes := []Node{guard}
		if mutate {
			nodes = append(nodes, Node{Kind: Event, Effects: []Effect{{Kind: "write", Resource: resource, Definite: true}}})
		}
		nodes = append(nodes, Node{Kind: Require, Requires: []Predicate{fact}})
		summary := inferTest(t, sequenceTest(nodes...), nil)
		for _, path := range summary.Paths {
			if path.Outcome != Normal {
				continue
			}
			status := path.Obligations[0].Status
			if (!mutate && status != Established) || (mutate && status == Established) {
				t.Fatalf("stale=%t status=%s", mutate, status)
			}
		}
	}
}

func TestInferenceBASE031OutcomeScopes(t *testing.T) {
	summary := inferTest(t, Node{Kind: Branch, Then: []Node{{Kind: Event}}, Else: []Node{eventTest("allocate", Invocation), {Kind: Event, Outcome: ReturnedFailure}}}, nil)
	if CheckProfile(summary, Profile{Name: "NoHeapAllocation", Outcomes: []Outcome{Normal}})[0].Status != Established {
		t.Fatal("success-only proof unavailable")
	}
	statusTest(t, summary, "NoHeapAllocation", Contradicted)
}

func TestInferenceBASE032RecursiveWidening(t *testing.T) {
	cost := ZeroResourceCost()
	cost.Work = ConstantCost(1)
	plan := Plan{Entry: Node{Kind: Call, Meaning: "recursive"}, Definitions: map[MeaningKey]Node{"recursive": sequenceTest(Node{Kind: Event, Costs: &cost}, Node{Kind: Call, Meaning: "recursive"})}}
	first, err := Infer(plan, AnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	second, err := Infer(plan, AnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if first.Resources.Work.Kind != "unknown" || first.Fingerprint != second.Fingerprint {
		t.Fatal("recursive analysis failed to widen deterministically")
	}
	if strings.Contains(first.Resources.Work.Reason, "unbounded execution") {
		t.Fatal("unknown bound reported as unbounded execution")
	}
}

func TestInferenceBASE033ReturnedErrorNotIO(t *testing.T) {
	summary := inferTest(t, Node{Kind: Event, Outcome: ReturnedFailure}, nil)
	statusTest(t, summary, "NoExternalIO", Established)
	statusTest(t, summary, "NoPanic", Established)
}

func TestInferenceFALS001AllocationInjection(t *testing.T) {
	clean := inferTest(t, Node{Kind: Event}, nil)
	statusTest(t, clean, "NoHeapAllocation", Established)
	fault := inferTest(t, eventTest("allocate", Invocation), nil)
	statusTest(t, fault, "NoHeapAllocation", Contradicted)
}

func TestInferenceFALS003DestructorIO(t *testing.T) {
	summary := inferTest(t, Node{Kind: Scope, Children: []Node{{Kind: Acquire, Resource: ResourceRef{ID: "object"}, Cleanup: []Node{eventTest("io", Destruction)}}}}, nil)
	statusTest(t, summary, "NoExternalIO", Contradicted)
	if CheckProfile(summary, Profile{Name: "NoExternalIO", Scope: "body_only"})[0].Status != Established {
		t.Fatal("scope boundaries conflated")
	}
}

func TestInferenceSubstitutionPreservesAliases(t *testing.T) {
	formal := ResourceRef{ID: "parameter"}
	actual := ResourceRef{ID: "actual"}
	fact := factTest("stable", 1)
	fact.Resource = actual
	plan := Plan{Entry: sequenceTest(Node{Kind: Event, Facts: []Predicate{fact}}, Node{Kind: Call, Meaning: "write", Substitutions: Substitutions{Resources: map[string]ResourceRef{"parameter": actual}}}, Node{Kind: Require, Requires: []Predicate{fact}}), Definitions: map[MeaningKey]Node{"write": {Kind: Event, Effects: []Effect{{Kind: "write", Resource: formal, Definite: true}}}}}
	summary, err := Infer(plan, AnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if AllEstablished(summary.Obligations) {
		t.Fatal("resource substitution lost invalidation")
	}
}

func TestInferenceRequiresCannotHideOnEvents(t *testing.T) {
	summary := inferTest(t, Node{Kind: Event, Requires: []Predicate{factTest("unsupported_precondition", 1)}}, nil)
	statusTest(t, summary, "NoBlocking", Unresolved)
}

func TestInferencePostconditionsReplaceOldValues(t *testing.T) {
	summary := inferTest(t, sequenceTest(Node{Kind: Event, Facts: []Predicate{factTest("x", 0)}}, Node{Kind: Event, Facts: []Predicate{factTest("x", 1)}}, Node{Kind: Require, Requires: []Predicate{factTest("x", 0)}}), nil)
	if AllEstablished(summary.Obligations) {
		t.Fatal("overwritten scalar fact remained available")
	}
}

func TestInferenceMissingOutcomeAndDestructorRemainUnknown(t *testing.T) {
	summary := inferTest(t, Node{Kind: Native, Native: "opaque"}, nil)
	if CheckProfile(summary, Profile{Name: "NoBlocking", Outcomes: []Outcome{Panic}})[0].Status != Unresolved {
		t.Fatal("unknown outcome was silently excluded")
	}
	summary = inferTest(t, Node{Kind: Scope, Children: []Node{{Kind: Acquire, Resource: ResourceRef{ID: "opaque-object"}}}}, nil)
	statusTest(t, summary, "NoBlocking", Unresolved)
	if CheckProfile(summary, Profile{Name: "NoBlocking", Scope: "body_only"})[0].Status != Established {
		t.Fatal("destructor uncertainty leaked into explicitly narrow body-only claim")
	}
}

func TestInferenceResidualsCoverTerminalOutcomes(t *testing.T) {
	plan := Plan{Entry: Node{Kind: Event, Outcome: Abort}, Residuals: []UnknownReason{{Code: "backend", Message: "runtime finalization is not closed", Components: []string{"effects", "costs"}}}}
	summary, err := Infer(plan, AnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	statusTest(t, summary, "NoBlocking", Unresolved)
	if len(summary.Paths) != 1 || summary.Paths[0].Outcome != Abort || summary.Resources.Work.Kind != "unknown" {
		t.Fatal("terminal source erased backend coverage or resource uncertainty")
	}
}

func TestInferenceNativeClosureNeedsApproval(t *testing.T) {
	plan := Plan{Entry: Node{Kind: Event}, NativeClosure: []NativeBoundary{{ID: "fixture-native", Status: Established, Witness: "fixture-only"}}}
	for _, approved := range []bool{false, true} {
		summary, err := Infer(plan, AnalysisOptions{ApprovedAssumptions: map[string]bool{"fixture-only": approved}})
		if err != nil {
			t.Fatal(err)
		}
		want := Unresolved
		if approved {
			want = Established
			if len(summary.Assumptions) != 1 || summary.Assumptions[0] != "fixture-only" {
				t.Fatal("approved conditional closure lost its assumption")
			}
		}
		statusTest(t, summary, "NoBlocking", want)
	}
}

func TestInferenceUnboundedDemandDoesNotBecomeZero(t *testing.T) {
	count := int64(1)
	summary := DerivedSummary{Paths: []PathSummary{{Outcome: Normal, Reachable: true, Demand: map[string]Multiplicity{"callback": {}}, Resources: ZeroResourceCost()}}}
	a := inference{}
	path := initialPath()
	path.Demand["callback"] = Multiplicity{Lower: 1, Upper: &count}
	result := a.applySummary(path, summary, Substitutions{})
	if result[0].Demand["callback"].Upper != nil {
		t.Fatal("unknown demand was treated as absent demand")
	}
}

func TestInferenceUnknownSummaryInvalidatesGuard(t *testing.T) {
	fact := factTest("still-ready", 1)
	fact.Resource = ResourceRef{ID: "mutable"}
	p := initialPath()
	p.Facts = []Predicate{fact}
	tail := initialPath()
	tail.Unknowns = []UnknownReason{{Code: "native", Components: []string{"effects"}}}
	a := inference{}
	paths := a.applySummary(p, DerivedSummary{Paths: []PathSummary{tail}}, Substitutions{})
	if status, _ := Entails(paths[0].Facts, fact); status == Established {
		t.Fatal("unknown callee summary retained a stale mutable guard")
	}
}

func TestInferenceUnrecognizedClassificationsRemainUnknown(t *testing.T) {
	for _, node := range []Node{{Kind: Event, Effects: []Effect{{Kind: "extension_effect"}}}, {Kind: Event, Outcome: "extension_outcome"}} {
		summary := inferTest(t, node, nil)
		statusTest(t, summary, "NoHeapAllocation", Unresolved)
		statusTest(t, summary, "NoBlocking", Unresolved)
	}
}
