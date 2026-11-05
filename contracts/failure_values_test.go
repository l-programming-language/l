package contracts

import "testing"

func TestTypedCompletionValuesAfterCleanup(t *testing.T) {
	location := FailureLocation{Resource: ResourceRef{ID: "state"}, Path: []string{"value"}}
	initial := ResourceValueFact(location, Atom{Kind: "number", Number: 1})
	changed := ResourceValueFact(location, Atom{Kind: "number", Number: 2})
	for _, test := range []struct {
		name          string
		body, cleanup []Node
		want          Status
	}{
		{"unchanged", nil, nil, Established},
		{"changed", []Node{{Kind: Event, Effects: []Effect{{Kind: "write", Resource: location.Resource, Definite: true}}, Facts: []Predicate{changed}}}, nil, Contradicted},
		{"restored", []Node{{Kind: Event, Effects: []Effect{{Kind: "write", Resource: location.Resource, Definite: true}}, Facts: []Predicate{changed}}}, []Node{{Kind: Event, Effects: []Effect{{Kind: "write", Resource: location.Resource, Definite: true}}, Facts: []Predicate{initial}}}, Established},
		{"unknown-cleanup", nil, []Node{{Kind: Native, Native: "opaque-cleanup"}}, Unresolved},
		{"may-alias-write", []Node{{Kind: Event, Effects: []Effect{{Kind: "write", Resource: ResourceRef{ID: "possible-alias"}, Definite: true}}}}, nil, Unresolved},
	} {
		t.Run(test.name, func(t *testing.T) {
			children := append(append([]Node(nil), test.body...), Node{Kind: Event, Outcome: ReturnedFailure})
			region := Node{Kind: Scope, ID: "region", Children: children, Cleanup: test.cleanup, CleanupOutcomes: []Outcome{Normal, ReturnedFailure}}
			plan := Plan{Entry: Node{Kind: Sequence, Children: []Node{{Kind: Event, Facts: []Predicate{initial}}, region}}}
			summary, err := InferScope(plan, "region", AnalysisOptions{})
			if err != nil {
				t.Fatal(err)
			}
			status, reason := TypedFinalValueRelation(summary.Paths[0], location)
			if status != test.want {
				t.Fatalf("got %s (%s), want %s: %+v", status, reason, test.want, summary.Paths[0])
			}
		})
	}
}

func TestTypedCompletionValuesRequireExactLocationAndEntry(t *testing.T) {
	location := FailureLocation{Resource: ResourceRef{ID: "state"}, Path: []string{"a"}}
	fact := ResourceValueFact(location, Atom{Kind: "text", Text: "value"})
	path := initialPath()
	path.EntryFacts, path.Facts = []Predicate{fact}, []Predicate{fact}
	other := location
	other.Path = []string{"b"}
	if status, _ := TypedFinalValueRelation(path, other); status != Unresolved {
		t.Fatal("same resource conflated distinct runtime paths")
	}
	path.EntryFacts = nil
	if status, _ := TypedFinalValueRelation(path, location); status != Unresolved {
		t.Fatal("completion fact manufactured entry value")
	}
	path.EntryFacts = []Predicate{{Kind: "eq", Subject: "arbitrary-label", Resource: location.Resource, Value: fact.Value}}
	if status, _ := TypedFinalValueRelation(path, location); status != Unresolved {
		t.Fatal("arbitrary predicate label became a typed value observation")
	}
}

func TestBudgetRetainsOnlyCompletedFailureWitnesses(t *testing.T) {
	location := FailureLocation{Resource: ResourceRef{ID: "slot"}}
	entry := ResourceValueFact(location, Atom{Kind: "number", Number: 1})
	changed := ResourceValueFact(location, Atom{Kind: "number", Number: 2})
	count := int64(32)
	failed := []Node{eventTest("block", Invocation), {Kind: Event, Facts: []Predicate{changed}, Effects: []Effect{{Kind: "write", Resource: location.Resource, Definite: true}}, Outcome: ReturnedFailure}}
	branch := Node{Kind: Branch, Then: failed, Else: []Node{{Kind: Repeat, Count: &count, Children: []Node{{Kind: Event}}}}}
	region := Node{Kind: Scope, ID: "operation", Children: []Node{branch}, CleanupOutcomes: []Outcome{ReturnedFailure}}
	footprint := FailureFootprint{Locations: []FailureLocation{location}, Outcomes: region.CleanupOutcomes, Observation: "state_at_completion"}
	for _, want := range []Status{Contradicted, Unresolved} {
		plan := Plan{Entry: sequenceTest(Node{Kind: Event, Facts: []Predicate{entry}}, region)}
		s, err := inferPlan(plan, AnalysisOptions{}, region.ID, 14)
		if err != nil || !hasWorkCutoff(s.Unknowns) {
			t.Fatalf("missing cutoff: %+v, %v", s, err)
		}
		if got := CheckFailureSummaryProfile(s, Profile{Name: "PreservesOnFailure", Scope: "body_only"}, footprint, nil); ObligationStatus(got) != want {
			t.Fatalf("preservation: %+v, want %s", got, want)
		}
		if got := CheckProfile(s, Profile{Name: "NoBlocking", Outcomes: region.CleanupOutcomes}); ObligationStatus(got) != want {
			t.Fatalf("returned failure: %+v, want %s", got, want)
		}
		region.Cleanup = []Node{{Kind: Event, Facts: []Predicate{entry}}}
	}
}
