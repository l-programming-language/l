package contracts

import "testing"

func TestRestrictedPredicateFragment(t *testing.T) {
	low, high := int64(2), int64(5)
	outside := int64(6)
	if status, _ := Entails([]Predicate{{Kind: "eq", Subject: "n", Value: Atom{Kind: "number", Number: 3}}}, Predicate{Kind: "interval", Subject: "n", Lower: &low, Upper: &high}); status != Established {
		t.Fatal("integer interval containment failed")
	}
	if status, _ := Entails([]Predicate{{Kind: "interval", Subject: "n", Lower: &low, Upper: &high}}, Predicate{Kind: "interval", Subject: "n", Lower: &outside}); status != Contradicted {
		t.Fatal("disjoint intervals were not refuted")
	}
	alias := Predicate{Kind: "resource_equal", Resource: ResourceRef{ID: "a"}, Value: Atom{Kind: "reference", Text: "b"}}
	if status, _ := Entails([]Predicate{alias}, Predicate{Kind: "resource_equal", Resource: ResourceRef{ID: "b"}, Value: Atom{Kind: "reference", Text: "a"}}); status != Established {
		t.Fatal("symmetric canonical resource alias failed")
	}
	if status, _ := Entails(nil, Predicate{Kind: "resource_disjoint", Resource: ResourceRef{ID: "a"}, Value: Atom{Kind: "reference", Text: "b"}}); status != Unresolved {
		t.Fatal("different names implied disjointness")
	}
	state := func(name string) Predicate { return Predicate{Value: Atom{Kind: "enum", Text: name}} }
	fact := Predicate{Kind: "protocol", Subject: "app", Args: []Predicate{state("running"), state("completed")}}
	if status, _ := Entails([]Predicate{fact}, Predicate{Kind: "protocol", Subject: "app", Args: []Predicate{state("running"), state("completed"), state("stopped")}}); status != Established {
		t.Fatal("protocol state-set containment failed")
	}
	if status, _ := Entails([]Predicate{fact}, Predicate{Kind: "protocol", Subject: "app", Value: Atom{Kind: "enum", Text: "running"}}); status != Unresolved {
		t.Fatal("alternative protocol state collapsed")
	}
	row := Predicate{Kind: "effect_contains", Subject: "callback", Args: []Predicate{{Kind: "eq", Subject: "effect", Value: Atom{Kind: "enum", Text: "read"}}}}
	if status, _ := Entails([]Predicate{row}, row); status != Unresolved {
		t.Fatal("open effect row certified containment")
	}
	if status, _ := Entails([]Predicate{row, {Kind: "effects_closed", Subject: "callback"}}, row); status != Established {
		t.Fatal("closed matching effect row failed")
	}
	if status, _ := Entails([]Predicate{{Kind: "unsupported"}}, Predicate{Kind: "unsupported"}); status != Unresolved {
		t.Fatal("unsupported assertion certified itself")
	}
}

func TestReferenceMissingTypeIsUndetermined(t *testing.T) {
	result := SelectReference([]Candidate{lexicalCandidate("typed", lexicalType("Theme"))}, SelectionContext{Target: TargetRef{Root: "instance", Unknown: false}})
	if result.Status != Unresolved {
		t.Fatalf("missing type was treated as a different exact type: %+v", result)
	}
}

func TestReferenceUnresolvedAttachmentRetainsCandidates(t *testing.T) {
	candidates := []Candidate{lexicalCandidate("first", lexicalType("Theme")), lexicalCandidate("second", lexicalType("Backend"))}
	for _, context := range []SelectionContext{{Target: TargetRef{Unknown: true}}, {Target: lexicalTarget("primary"), Policy: "future-unapproved"}} {
		selection := SelectReference(candidates, context)
		if selection.Status != Unresolved || len(selection.Candidates) != len(candidates) || selection.Meaning != "" {
			t.Fatalf("unresolved attachment/policy hid competing meanings: %+v", selection)
		}
	}
}

func TestRuntimeFrameUsesAliasEvidence(t *testing.T) {
	target := lexicalTarget("protected")
	other := lexicalTarget("other")
	a, b := ResourceRef{ID: "a"}, ResourceRef{ID: "b"}
	frame := FrameContract{Target: target, Resource: a, Scope: "runtime", Allowed: []TargetRef{other}, Preserves: []TargetRef{target}}
	writes := []FrameFootprint{{Kind: "write", Target: other, Resource: b, Definite: true}}
	result := CheckFrame(frame, writes, FrameContext{Aliases: []FrameResourcePair{{Left: a, Right: b}}})
	if !hasStatus(result, Contradicted) {
		t.Fatal("known aliases did not refute runtime preservation")
	}
}

func TestAdjectiveInstantiationChecksCallbackResources(t *testing.T) {
	transformer := supportedTransformer()
	contract := transformer.Callbacks["action"]
	budget := ZeroResourceCost()
	budget.Work = ConstantCost(1)
	contract.Costs = &budget
	transformer.Callbacks["action"] = contract
	certificate := requireCertificate(t, transformer)
	if certificate.Status != Established {
		t.Fatalf("conditional bounded callback failed: %+v", certificate.Obligations)
	}
	instance := transformerInstance()
	callback := instance.Callbacks["action"]
	actual := ZeroResourceCost()
	actual.Work = ConstantCost(2)
	callback.Costs = &actual
	instance.Callbacks["action"] = callback
	requireInstantiation(t, transformer, certificate, instance, Contradicted)
}
