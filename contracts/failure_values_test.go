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
