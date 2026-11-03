package contracts

import (
	"reflect"
	"testing"
)

func interactionAlternatives(nodes []Node) Node {
	result := nodes[len(nodes)-1]
	for i := len(nodes) - 2; i >= 0; i-- {
		result = Node{Kind: Branch, Then: []Node{nodes[i]}, Else: []Node{result}}
	}
	return result
}

func TestInferenceDemandBudgetAndBranchOrder(t *testing.T) {
	zero := int64(0)
	cost := ZeroResourceCost()
	for _, reverse := range []bool{false, true} {
		for _, budget := range []int{1, 32, 64} {
			nodes := make([]Node, 33)
			for i := range nodes {
				nodes[i] = Node{Kind: Event}
			}
			index := 0
			if reverse {
				index = len(nodes) - 1
			}
			nodes[index] = Node{Kind: Callback, Callback: "action"}
			transformer := ContractTransformer{Meaning: "interaction-demand", Version: "1", Target: TargetRef{Root: "target"}, Callbacks: map[string]SymbolicActionContract{"action": {EffectsClosed: true, Outcomes: []Outcome{Normal}, Costs: &cost}}, Demand: map[string]Multiplicity{"action": {Upper: &zero}}, Body: Plan{Entry: interactionAlternatives(nodes)}}
			certificate, err := VerifyAdjective(transformer, AnalysisOptions{MaxAlternatives: budget})
			if err != nil {
				t.Fatal(err)
			}
			if certificate.Status == Established {
				t.Fatalf("reverse=%t budget=%d certified zero calls", reverse, budget)
			}
		}
	}
}

func TestInferenceOwnershipBudgetFinalOwnerAndGuard(t *testing.T) {
	resource := ResourceRef{ID: "guard"}
	blocking := []Node{{Kind: Event, Phase: Destruction, Effects: []Effect{{Kind: "block", Definite: true}}}}
	for _, reverse := range []bool{false, true} {
		for _, budget := range []int{1, 32, 64} {
			nodes := make([]Node, 33)
			for i := range nodes {
				nodes[i] = Node{Kind: Acquire, Resource: resource, Ownership: "shared", FinalOwner: Contradicted, Cleanup: blocking}
			}
			index := 0
			if reverse {
				index = len(nodes) - 1
			}
			nodes[index].FinalOwner = Established
			summary := reviewInfer(t, Plan{Entry: Node{Kind: Scope, Children: []Node{interactionAlternatives(nodes)}}}, AnalysisOptions{MaxAlternatives: budget})
			if AllEstablished(CheckProfile(summary, Profile{Name: "NoBlocking"})) {
				t.Fatalf("reverse=%t budget=%d lost final-owner cleanup", reverse, budget)
			}
			if !AllEstablished(CheckProfile(summary, Profile{Name: "NoBlocking", Scope: "body_only"})) {
				t.Fatalf("cleanup uncertainty leaked into body-only scope: %+v", summary)
			}
		}
	}
	for _, budget := range []int{1, 2} {
		guard := Node{Kind: Acquire, Resource: resource, Ownership: "guard", SuspensionSafe: true}
		unsafe := guard
		unsafe.SuspensionSafe = false
		plan := Plan{Entry: Node{Kind: Sequence, Children: []Node{interactionAlternatives([]Node{guard, unsafe}), {Kind: Event, Effects: []Effect{{Kind: "suspend", Definite: true}}}}}}
		summary := reviewInfer(t, plan, AnalysisOptions{MaxAlternatives: budget})
		if len(summary.Obligations) == 0 || AllEstablished(summary.Obligations) {
			t.Fatalf("budget %d lost possible unsafe guard", budget)
		}
	}
}

func TestInferenceOwnershipSnapshotsPreserveOrderAndCosts(t *testing.T) {
	cost := ZeroResourceCost()
	cost.Work = ConstantCost(3)
	acquire := func(id string) Node {
		return Node{Kind: Acquire, Resource: ResourceRef{ID: id}, FinalOwner: Established, Cleanup: []Node{{Kind: Event, Costs: &cost}}}
	}
	first := Node{Kind: Sequence, Children: []Node{acquire("a"), acquire("b")}}
	second := Node{Kind: Sequence, Children: []Node{acquire("b"), acquire("a")}}
	summary := reviewInfer(t, Plan{Entry: Node{Kind: Scope, Children: []Node{interactionAlternatives([]Node{first, first, second})}}}, AnalysisOptions{MaxAlternatives: 2})
	orders := make(map[string]bool)
	for _, path := range summary.Paths {
		order := ""
		for _, effect := range path.Effects {
			if effect.Kind == "release" {
				order += effect.Resource.ID
			}
		}
		orders[order] = true
	}
	if !orders["ab"] || !orders["ba"] || len(orders) != 2 {
		t.Fatalf("lost correlated release orders: %v", orders)
	}
	if work, ok := summary.Resources.Work.Constant(); !ok || work != 6 {
		t.Fatalf("alternative cleanups were double charged: %+v", summary.Resources)
	}
	if !AllEstablished(CheckProfile(summary, Profile{Name: "NoBlocking"})) {
		t.Fatal("known effect-free destructor alternatives lost their positive guarantee")
	}
}

func TestInferenceSummaryEffectsPreserveCallerContextAndCosts(t *testing.T) {
	cost := ZeroResourceCost()
	cost.Work = ConstantCost(7)
	leaf := reviewInfer(t, Plan{Entry: Node{Kind: Event, Costs: &cost, Effects: []Effect{{Kind: "suspend", Definite: true}}}}, AnalysisOptions{})
	options := AnalysisOptions{NativeSummaries: map[string]DerivedSummary{"leaf": leaf}, NativeWitnesses: map[string]string{"leaf": "test-only:interaction"}, ApprovedAssumptions: map[string]bool{"test-only:interaction": true}}
	for _, safe := range []bool{false, true} {
		for _, released := range []bool{false, true} {
			guard := ResourceRef{ID: "guard"}
			nodes := []Node{{Kind: Acquire, Resource: guard, Ownership: "guard", SuspensionSafe: safe, Cleanup: []Node{{Kind: Event}}, FinalOwner: Established}}
			if released {
				nodes = append(nodes, Node{Kind: Release, Resource: guard})
			}
			nodes = append(nodes, Node{Kind: Native, Native: "leaf"})
			summary := reviewInfer(t, Plan{Entry: Node{Kind: Sequence, Children: nodes}}, options)
			if (!safe && !released) == AllEstablished(summary.Obligations) {
				t.Fatalf("safe=%t released=%t obligations=%+v", safe, released, summary.Obligations)
			}
			if work, ok := summary.Resources.Work.Constant(); !ok || work != 7 {
				t.Fatalf("summary costs changed: %+v", summary.Resources)
			}
		}
	}
}

func TestInferenceNestedPredicateInvalidationAndDisjointness(t *testing.T) {
	resource, other := ResourceRef{ID: "mutable"}, ResourceRef{ID: "other"}
	for _, kind := range []string{"and", "or", "not", "implies"} {
		fact := Predicate{Kind: kind, Args: []Predicate{{Kind: "and", Args: []Predicate{{Kind: "eq", Subject: "state", Resource: resource, Value: Atom{Kind: "number", Number: 1}}}}}}
		for _, write := range []ResourceRef{resource, {Unknown: true}, other} {
			analysis := inference{options: AnalysisOptions{Disjoint: []ResourcePair{{Left: resource, Right: other}}}}
			path := initialPath()
			path.Facts = []Predicate{fact}
			analysis.invalidate(&path, write)
			if (write.ID == other.ID) != (len(path.Facts) == 1) {
				t.Fatalf("kind=%s write=%+v facts=%+v", kind, write, path.Facts)
			}
		}
	}
}

func FuzzDemandJoinSoundness(f *testing.F) {
	f.Add(uint8(0), uint8(1), false, false)
	f.Add(uint8(4), uint8(9), true, false)
	f.Add(uint8(2), uint8(0), false, true)
	f.Fuzz(func(t *testing.T, first, second uint8, unknownFirst, unknownSecond bool) {
		a, b := int64(first), int64(second)
		left, right := Multiplicity{Lower: a, Upper: &a}, Multiplicity{Lower: b, Upper: &b}
		if unknownFirst {
			left.Upper = nil
		}
		if unknownSecond {
			right.Upper = nil
		}
		joined := joinDemand(left, right)
		if !reflect.DeepEqual(joined, joinDemand(right, left)) || joined.Lower > a || joined.Lower > b {
			t.Fatalf("unsound/asymmetric join: %+v", joined)
		}
		if unknownFirst || unknownSecond {
			if joined.Upper != nil {
				t.Fatal("unknown upper became finite")
			}
		} else if joined.Upper == nil || *joined.Upper < a || *joined.Upper < b {
			t.Fatal("join excludes a branch")
		}
	})
}

func TestInferenceScopeRetainsEntryFactsAndOwnership(t *testing.T) {
	resource := ResourceRef{ID: "guard"}
	fact := Predicate{Kind: "eq", Subject: "ready", Value: Atom{Kind: "bool", Bool: true}}
	cost := ZeroResourceCost()
	cost.Work = ConstantCost(5)
	plan := Plan{Entry: Node{Kind: Sequence, Children: []Node{
		{Kind: Event, Facts: []Predicate{fact}, Costs: &cost, Effects: []Effect{{Kind: "block", Definite: true}}},
		{Kind: Acquire, Resource: resource, Ownership: "guard", FinalOwner: Established, Cleanup: []Node{{Kind: Event}}},
		{Kind: Sequence, ID: "selected", Children: []Node{{Kind: Require, Requires: []Predicate{fact}}, {Kind: Event, Costs: &cost, Effects: []Effect{{Kind: "suspend", Definite: true}}}}},
	}}}
	summary, err := InferScope(plan, "selected", AnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(summary.Obligations) != 2 || summary.Obligations[0].Status != Established || summary.Obligations[1].Status != Contradicted {
		t.Fatalf("selected entry context lost: %+v", summary.Obligations)
	}
	if len(summary.Effects) != 1 || summary.Effects[0].Kind != "suspend" {
		t.Fatalf("prefix effects leaked into scope: %+v", summary.Effects)
	}
	if work, ok := summary.Resources.Work.Constant(); !ok || work != 5 {
		t.Fatalf("prefix costs leaked into scope: %+v", summary.Resources)
	}
	if _, err := InferScope(plan, "missing", AnalysisOptions{}); err == nil {
		t.Fatal("missing selection accepted")
	}
}

func TestInferenceReturnPreservesCleanupAndCallerContinuation(t *testing.T) {
	body := Node{Kind: Scope, Children: []Node{
		{Kind: Acquire, Resource: ResourceRef{ID: "local"}, FinalOwner: Established, Cleanup: []Node{{Kind: Event, Effects: []Effect{{Kind: "io", Definite: true, Phase: Destruction}}}}},
		{Kind: Return, Children: []Node{{Kind: Event, Facts: []Predicate{{Kind: "eq", Subject: "returned-error-value", Value: Atom{Kind: "bool", Bool: true}}}}}},
		{Kind: Event, Effects: []Effect{{Kind: "block", Definite: true}}},
	}}
	plan := Plan{Definitions: map[MeaningKey]Node{"function": body}, Entry: Node{Kind: Sequence, Children: []Node{{Kind: Call, Meaning: "function"}, {Kind: Event, Effects: []Effect{{Kind: "allocate", Definite: true}}}}}}
	summary := reviewInfer(t, plan, AnalysisOptions{})
	if summary.Paths[0].Outcome != Normal || summary.Paths[0].ControlReturn {
		t.Fatal("return value changed failure channel or escaped callee")
	}
	statusTest(t, summary, "NoBlocking", Established)
	statusTest(t, summary, "NoExternalIO", Contradicted)
	statusTest(t, summary, "NoHeapAllocation", Contradicted)
}

func TestInferenceCalleeStateRelationUsesEnclosingEntry(t *testing.T) {
	resource := ResourceRef{ID: "state"}
	relation := StateRelation{Location: FailureLocation{Resource: resource}, Observation: "after_cleanup", Relation: "unchanged"}
	leaf := initialPath()
	leaf.StateRelations = []StateRelation{relation}
	analysis := inference{options: AnalysisOptions{}.normalized()}
	for _, mutated := range []bool{false, true} {
		prefix := initialPath()
		if mutated {
			analysis.observeEffect(&prefix, Effect{Kind: "write", Resource: resource, Definite: true})
		}
		paths := analysis.applySummary(prefix, summarize([]PathSummary{leaf}), Substitutions{})
		if (len(paths[0].StateRelations) == 1) == mutated {
			t.Fatalf("mutated=%t miscomposed entry relation: %+v", mutated, paths[0].StateRelations)
		}
	}
}

func TestInferenceInstantiatedCallBodiesRemainDistinct(t *testing.T) {
	first := Node{Kind: Event, ID: "first-instance", Effects: []Effect{{Kind: "allocate", Definite: true}}}
	second := Node{Kind: Event, ID: "second-instance", Effects: []Effect{{Kind: "block", Definite: true}}}
	plan := Plan{Definitions: map[MeaningKey]Node{"canonical": {Kind: Event, Effects: []Effect{{Kind: "io", Definite: true}}}}, Entry: Node{Kind: Sequence, Children: []Node{{Kind: Call, Meaning: "canonical", Body: &first}, {Kind: Call, Meaning: "canonical", Body: &second}}}}
	summary := reviewInfer(t, plan, AnalysisOptions{})
	statusTest(t, summary, "NoHeapAllocation", Contradicted)
	statusTest(t, summary, "NoBlocking", Contradicted)
	statusTest(t, summary, "NoExternalIO", Established)
	selected, err := InferScope(plan, first.ID, AnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	statusTest(t, selected, "NoBlocking", Established)
	statusTest(t, selected, "NoHeapAllocation", Contradicted)
}
