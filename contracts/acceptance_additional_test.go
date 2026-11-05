package contracts

import "testing"

func TestAcceptanceInstancesBASE003(t *testing.T) {
	formal := TargetRef{Root: "receiver", Path: []string{"value"}, Phase: Invocation}
	definition := Node{Kind: Event, Effects: []Effect{{Kind: "write", Target: formal, Resource: ResourceRef{ID: "storage"}, Definite: true}}}
	first, second := TargetRef{Root: "app", Path: []string{"first"}, Phase: Invocation}, TargetRef{Root: "app", Path: []string{"second"}, Phase: Invocation}
	plan := Plan{Entry: sequenceTest(
		Node{Kind: Call, Meaning: "setter", Substitutions: Substitutions{Targets: map[string]TargetRef{"receiver": first}, Resources: map[string]ResourceRef{"storage": {ID: "left"}}}},
		Node{Kind: Call, Meaning: "setter", Substitutions: Substitutions{Targets: map[string]TargetRef{"receiver": second}, Resources: map[string]ResourceRef{"storage": {ID: "right"}}}},
	), Definitions: map[MeaningKey]Node{"setter": definition}}
	summary := reviewInfer(t, plan, AnalysisOptions{})
	if len(summary.Effects) != 2 || summary.Effects[0].Target.String() == summary.Effects[1].Target.String() || summary.Effects[0].Resource.ID == summary.Effects[1].Resource.ID {
		t.Fatalf("parameterized summary collapsed distinct instances: %+v", summary.Effects)
	}
	for _, effect := range summary.Effects {
		if effect.Target.Root != "app" || len(effect.Target.Path) != 2 || effect.Target.Path[1] != "value" {
			t.Fatalf("member projection was not rebound: %+v", effect)
		}
	}
}

func TestAcceptanceCapturePhasesBASE014(t *testing.T) {
	construction := eventTest("allocate", Construction)
	destruction := eventTest("allocate", Destruction)
	closure := Node{Kind: Scope, Children: []Node{construction, {Kind: Acquire, Resource: ResourceRef{ID: "capture"}, Cleanup: []Node{destruction}}, {Kind: Callback, Callback: "closure"}}}
	summary := inferTest(t, closure, map[string]Node{"closure": eventTest("allocate", Invocation)})
	phases := make(map[Phase]int)
	for _, effect := range summary.Effects {
		phases[effect.Phase]++
	}
	if phases[Construction] != 1 || phases[Invocation] != 1 || phases[Destruction] != 1 {
		t.Fatalf("capture lifecycle events lost their phase: %+v", summary.Effects)
	}
	statusTest(t, summary, "NoHeapAllocation", Contradicted)
}

func TestAcceptanceDemandDependenciesBASE015(t *testing.T) {
	cost := ZeroResourceCost()
	cost.Work = ConstantCost(9)
	drop := eventTest("io", Destruction)
	dropCost := ZeroResourceCost()
	dropCost.IO = ConstantCost(1)
	drop.Costs = &dropCost
	callback := Node{Kind: Scope, CleanupOutcomes: []Outcome{Normal, ReturnedFailure}, Children: []Node{
		{Kind: Acquire, Resource: ResourceRef{ID: "capture"}, Cleanup: []Node{drop}},
		{Kind: Event, Costs: &cost, Effects: []Effect{{Kind: "allocate", Phase: Invocation, Definite: true}}},
		{Kind: Event, Outcome: ReturnedFailure},
	}}
	without := inferTest(t, Node{Kind: Event}, nil)
	with := inferTest(t, Node{Kind: Callback, Callback: "action"}, map[string]Node{"action": callback})
	if without.Fingerprint == with.Fingerprint || with.Paths[0].Outcome != ReturnedFailure || with.Paths[0].Demand["action"].Lower != 1 || with.Paths[0].Owned["capture"].State != "released" {
		t.Fatalf("changed demand retained stale failure/ownership: %+v", with)
	}
	if work, ok := with.Resources.Work.Constant(); !ok || work != 9 {
		t.Fatal("callback cost was not substituted")
	}
	statusTest(t, without, "NoHeapAllocation", Established)
	statusTest(t, with, "NoHeapAllocation", Contradicted)
	statusTest(t, with, "NoExternalIO", Contradicted)
	session := NewSession()
	key := implementationKeyFixture(logicalKeyFixture())
	key.Dependencies["demand"] = without.Fingerprint
	if err := session.store(key, EvidenceResult{Status: Established}); err != nil {
		t.Fatal(err)
	}
	key.Dependencies["demand"] = with.Fingerprint
	var stale EvidenceResult
	if found, err := session.load(key, &stale); err != nil || found {
		t.Fatalf("demand-dependent assurance survived changed callback: %v, %v", found, err)
	}
}

func TestAcceptanceUninitializedCleanupBASE021(t *testing.T) {
	resource := ResourceRef{ID: "conditional"}
	acquire := Node{Kind: Acquire, Resource: resource, Cleanup: []Node{eventTest("io", Destruction)}}
	summary := inferTest(t, Node{Kind: Scope, Children: []Node{{Kind: Branch, Then: []Node{acquire}, Else: []Node{{Kind: Branch, Then: []Node{acquire, {Kind: Move, Resource: resource}}, Else: []Node{{Kind: Event}}}}}}}, nil)
	if len(summary.Paths) != 3 {
		t.Fatalf("initialization alternatives lost: %+v", summary.Paths)
	}
	uninitialized, moved, released := 0, 0, 0
	for _, path := range summary.Paths {
		ioEvents := 0
		for _, effect := range path.Effects {
			if effect.Kind == "io" {
				ioEvents++
			}
		}
		switch path.Owned[resource.ID].State {
		case "":
			uninitialized++
		case "moved":
			moved++
		case "released":
			released++
		}
		if path.Owned[resource.ID].State == "released" && ioEvents != 1 || path.Owned[resource.ID].State != "released" && ioEvents != 0 {
			t.Fatalf("cleanup ran on an unowned path or was duplicated: %+v", path)
		}
	}
	if uninitialized != 1 || moved != 1 || released != 1 {
		t.Fatalf("ownership alternatives collapsed: %d/%d/%d", uninitialized, moved, released)
	}
}

func TestAcceptanceRecursiveDestructionBASE026(t *testing.T) {
	plan := Plan{Entry: Node{Kind: Scope, Children: []Node{{Kind: Acquire, Resource: ResourceRef{ID: "recursive"}, Cleanup: []Node{{Kind: Call, Meaning: "drop"}}}}}, Definitions: map[MeaningKey]Node{"drop": {Kind: Call, Meaning: "drop"}}}
	summary := reviewInfer(t, plan, AnalysisOptions{})
	if summary.Resources.Work.Kind != "unknown" || len(summary.Unknowns) == 0 {
		t.Fatalf("recursive destruction became free: %+v", summary)
	}
	statusTest(t, summary, "NoPanic", Unresolved)
}

func TestAcceptancePostEventDetectionBASE029(t *testing.T) {
	claim, _, _, fingerprints := evidenceFixture()
	claim.Predicate = Predicate{Kind: "observed", Subject: "allocation", Value: Atom{Kind: "number", Number: 8}}
	graph := EvidenceGraph{Claims: []Claim{claim}, Nodes: []EvidenceNode{{ID: "detector", Claim: claim.ID, Kind: "empirical", Witness: "post-event-record"}}}
	context := EvidenceContext{Current: map[string]EvidenceFingerprints{"detector": fingerprints}, Observations: map[string]CheckedEvidence{"post-event-record": {ClaimFingerprint: ClaimFingerprint(claim), Checker: "synthetic-post-event-allocation-detector", Fingerprints: fingerprints}}}
	if got := graph.Evaluate("detector", context); got.Status != Established {
		t.Fatalf("actual observation was lost: %+v", got)
	}
	claim.Predicate = Predicate{Kind: "effects_closed", Subject: "allocation", Value: Atom{Kind: "bool", Bool: false}}
	graph.Claims[0] = claim
	context.Observations["post-event-record"] = CheckedEvidence{ClaimFingerprint: ClaimFingerprint(claim), Checker: "synthetic-post-event-allocation-detector", Fingerprints: fingerprints}
	if got := graph.Evaluate("detector", context); got.Status != Unresolved {
		t.Fatalf("post-event observation certified prevention: %+v", got)
	}
	graph.Nodes[0].Kind = "counterexample"
	context.Refutations = context.Observations
	if got := graph.Evaluate("detector", context); got.Status != Contradicted {
		t.Fatalf("observed violation did not refute universal claim: %+v", got)
	}
}

func TestAcceptanceCheckedArithmeticDomainBASE033(t *testing.T) {
	finite := Predicate{Kind: "finite", Subject: "operand", Value: Atom{Kind: "bool", Bool: true}}
	checked := Node{Kind: Branch, Requires: []Predicate{finite}, Then: []Node{{Kind: Event, Facts: []Predicate{factTest("result", 3)}}}, Else: []Node{{Kind: Event, Outcome: ReturnedFailure}}}
	for _, provided := range []bool{true, false} {
		var facts []Predicate
		if provided {
			facts = []Predicate{finite}
		}
		summary := inferTest(t, sequenceTest(Node{Kind: Event, Facts: facts}, checked), nil)
		if len(summary.Paths) != 2 || summary.Paths[0].Outcome != Normal || summary.Paths[1].Outcome != ReturnedFailure {
			t.Fatalf("checked return alternatives collapsed: %+v", summary.Paths)
		}
		if provided {
			statusTest(t, summary, "NoExternalIO", Established)
			if status, _ := Entails(summary.Paths[0].Facts, factTest("result", 3)); status != Established {
				t.Fatal("numeric result semantics lost on normal outcome")
			}
		} else if AllEstablished(summary.Obligations) {
			t.Fatal("pure checked operation silently gained a larger totality domain")
		}
	}
}
