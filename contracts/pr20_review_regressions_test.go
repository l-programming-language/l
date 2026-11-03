// Proposed regression tests for PR #20 at
// ada795faae59bfe417f85927391debc161c9c473.
// These tests express the required sound behavior. They were syntax-checked,
// but NOT compiled or executed against the repository during this review.
// Copy into contracts/ and run: go test ./contracts -run '^TestPR20Review' -count=1
package contracts

import "testing"

func reviewInfer(t *testing.T, plan Plan, options AnalysisOptions) DerivedSummary {
	t.Helper()
	s, err := Infer(plan, options)
	if err != nil {
		t.Fatalf("fixture did not analyze: %v", err)
	}
	return s
}

// F01: Explicit unknown demand is not the same as an absent map entry.
func TestPR20ReviewUnknownDemandIsNotZero(t *testing.T) {
	zero := int64(0)
	transformer := ContractTransformer{
		Meaning: "review:demand", Version: "1", Target: TargetRef{Root: "target"},
		Demand: map[string]Multiplicity{"action": {Lower: 0, Upper: &zero}},
	}
	summary := DerivedSummary{Paths: []PathSummary{{
		Outcome: Normal, Reachable: true,
		Demand: map[string]Multiplicity{"action": {Lower: 0}},
	}}}
	obligations := checkTransformerSummary(transformer, summary)
	if len(obligations) == 0 || AllEstablished(obligations) {
		t.Fatalf("unknown callback count certified as zero: %+v", obligations)
	}
}

// F01 integration: path widening must not certify a never-called callback
// when a reachable alternative calls it once.
func TestPR20ReviewBudgetJoinCannotCertifyZeroDemand(t *testing.T) {
	zero := int64(0)
	cost := ZeroResourceCost()
	transformer := ContractTransformer{
		Meaning: "review:branch-demand", Version: "1", Target: TargetRef{Root: "target"},
		Callbacks: map[string]SymbolicActionContract{
			"action": {EffectsClosed: true, Outcomes: []Outcome{Normal}, Costs: &cost},
		},
		Demand: map[string]Multiplicity{"action": {Lower: 0, Upper: &zero}},
		Body: Plan{Entry: Node{
			Kind: Branch,
			Then: []Node{{Kind: Callback, Callback: "action"}},
			Else: []Node{{Kind: Event}},
		}},
	}
	certificate, err := VerifyAdjective(transformer, AnalysisOptions{MaxAlternatives: 1})
	if err != nil {
		t.Fatalf("fixture did not verify: %v", err)
	}
	if certificate.Status == Established {
		t.Fatalf("budgeted analysis certified false zero-demand promise: %+v", certificate)
	}
}

// F02: two paths own the same slot but can carry different destruction work.
// Lower analysis precision must not erase the blocking alternative.
func TestPR20ReviewBudgetJoinRetainsDestructorAlternatives(t *testing.T) {
	resource := ResourceRef{ID: "resource"}
	clean := Node{Kind: Acquire, Resource: resource, Ownership: "owned", FinalOwner: Established,
		Cleanup: []Node{{Kind: Event, Phase: Destruction}}}
	blocking := Node{Kind: Acquire, Resource: resource, Ownership: "owned", FinalOwner: Established,
		Cleanup: []Node{{Kind: Event, Phase: Destruction,
			Effects: []Effect{{Kind: "block", Phase: Destruction, Definite: true}}}}}
	plan := Plan{Entry: Node{Kind: Scope, Children: []Node{{
		Kind: Branch, Then: []Node{clean}, Else: []Node{blocking},
	}}}}
	for _, budget := range []int{32, 1} {
		summary := reviewInfer(t, plan, AnalysisOptions{MaxAlternatives: budget})
		obligations := CheckProfile(summary, Profile{Name: "NoBlocking"})
		if AllEstablished(obligations) {
			t.Fatalf("budget %d erased possible blocking destructor: %+v", budget, summary)
		}
	}
}

// F03: caller-owned unsafe guards must be checked when suspension is imported
// through an approved summary, just as for an inline suspension event.
func TestPR20ReviewNativeSummaryRechecksCallerGuard(t *testing.T) {
	suspension := Node{Kind: Event, Phase: Invocation,
		Effects: []Effect{{Kind: "suspend", Phase: Invocation, Definite: true}}}
	callee := reviewInfer(t, Plan{Entry: suspension}, AnalysisOptions{})
	options := AnalysisOptions{
		NativeSummaries:     map[string]DerivedSummary{"suspending-leaf": callee},
		NativeWitnesses:     map[string]string{"suspending-leaf": "test-only:review-witness"},
		ApprovedAssumptions: map[string]bool{"test-only:review-witness": true},
	}
	for _, operation := range []Node{suspension, {Kind: Native, Native: "suspending-leaf"}} {
		plan := Plan{Entry: Node{Kind: Scope, Children: []Node{
			{Kind: Acquire, Resource: ResourceRef{ID: "guard"}, Ownership: "guard",
				SuspensionSafe: false, FinalOwner: Established,
				Cleanup: []Node{{Kind: Event, Phase: Destruction}}},
			operation,
		}}}
		summary := reviewInfer(t, plan, options)
		if len(summary.Obligations) == 0 || AllEstablished(summary.Obligations) {
			t.Fatalf("%s hid unsafe caller guard across suspension: %+v", operation.Kind, summary)
		}
	}
}

// F04: a resource-dependent predicate nested in a conjunction must not survive
// a write merely because the outer conjunction has no Resource field.
func TestPR20ReviewNestedResourceFactInvalidated(t *testing.T) {
	resource := ResourceRef{ID: "mutable"}
	fact := Predicate{Kind: "eq", Subject: "state", Resource: resource,
		Value: Atom{Kind: "number", Number: 1}}
	plan := Plan{Entry: Node{Kind: Sequence, Children: []Node{
		{Kind: Event, Facts: []Predicate{{Kind: "and", Args: []Predicate{fact}}}},
		{Kind: Event, Effects: []Effect{{Kind: "write", Resource: resource, Definite: true}}},
		{Kind: Require, Requires: []Predicate{fact}},
	}}}
	summary := reviewInfer(t, plan, AnalysisOptions{})
	if len(summary.Obligations) == 0 || AllEstablished(summary.Obligations) {
		t.Fatalf("stale nested resource fact satisfied post-write requirement: %+v", summary)
	}
}

// F05: rule-derived selection facts must obey the same facet schema and
// cardinality rules as explicit contributions.
func TestPR20ReviewDerivedFacetSelectionsRespectSchema(t *testing.T) {
	a := Atom{Kind: "text", Text: "A"}
	b := Atom{Kind: "text", Text: "B"}
	outside := Atom{Kind: "text", Text: "OUTSIDE_DOMAIN"}
	target := TargetRef{Root: "component", Type: "Storage", Phase: Elaboration}
	slot := FacetSlot{Facet: "review.policy", Target: target, Scope: "normal-exit", Closed: true}
	schema := FacetSchema{
		ID: slot.Facet, Version: "1", Applicability: Predicate{Kind: "true"},
		Domain: "enum", Values: []Atom{a, b}, Cardinality: "exactly-one-at-close",
		AllowedContributions: []string{"select", "require"}, ClosurePoint: "configuration",
	}
	for _, tc := range []struct {
		name    string
		derived Predicate
	}{
		{"exclusive-conflict", FacetFact(slot, b)},
		{"out-of-domain", FacetFact(slot, outside)},
		{"undeclared-facet", FacetFact(FacetSlot{Facet: "review.missing", Target: target, Scope: slot.Scope}, b)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, err := CheckConfiguration(
				[]FacetSchema{schema}, []FacetSlot{slot},
				[]ConfigurationContribution{{Facet: schema.ID, Target: target, Scope: slot.Scope, Kind: "select", Value: a}},
				[]FacetRule{{Consequences: []Predicate{tc.derived}, Requires: []Predicate{tc.derived}}}, nil)
			if err != nil {
				return // A precise schema/configuration rejection is also valid.
			}
			if AllEstablished(result.Obligations) {
				t.Fatalf("invalid rule-derived selection accepted: %+v", result)
			}
		})
	}
}
