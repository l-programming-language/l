package contracts

import "testing"

func supportedTransformer() ContractTransformer {
	one := int64(1)
	target := TargetRef{Root: "$target", Type: "App", Phase: Elaboration}
	child := target
	child.Path = []string{"role", "child"}
	other := target
	other.Path = []string{"role", "other"}
	ready := Predicate{Kind: "eq", Subject: "ready", Value: Atom{Kind: "bool", Bool: true}}
	return ContractTransformer{Meaning: "fixture::ConstructAndAct", Version: "1", Target: target, Requires: []Predicate{lexicalType("App")}, Provides: []Predicate{ready}, Callbacks: map[string]SymbolicActionContract{"action": {EffectsClosed: true, Ensures: []Predicate{ready}, Outcomes: []Outcome{Normal}}}, Demand: map[string]Multiplicity{"action": {Lower: 1, Upper: &one}}, Profiles: []Profile{{Name: "NoBlocking", Scope: "whole_region"}}, Frame: &FrameContract{Target: target, Allowed: []TargetRef{child}, Preserves: []TargetRef{other}, Scope: "elaboration"}, Origin: Origin{Source: "wrapper.l", Start: 1, End: 30}, Body: Plan{Entry: Node{Kind: Sequence, Children: []Node{{Kind: Acquire, ID: "construction:one", Resource: ResourceRef{ID: "fresh:child"}, Target: child, Effects: []Effect{{Kind: "write", Target: child, Definite: true, Origin: Origin{Source: "wrapper.l"}}}}, {Kind: Callback, Callback: "action"}}}}}
}

func transformerInstance() AdjectiveInstance {
	return AdjectiveInstance{Target: TargetRef{Root: "app:concrete", Type: "App", Phase: Elaboration}, Facts: []Predicate{lexicalType("App")}, Callbacks: map[string]Node{"action": {Kind: Event, Facts: []Predicate{{Kind: "eq", Subject: "ready", Value: Atom{Kind: "bool", Bool: true}}}}}}
}

func requireCertificate(t *testing.T, transformer ContractTransformer) AdjectiveCertificate {
	t.Helper()
	certificate, err := VerifyAdjective(transformer, AnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	return certificate
}

func TestAdjectiveAcceptance(t *testing.T) {
	t.Run("BASE-050", func(t *testing.T) {
		transformer := supportedTransformer()
		certificate := requireCertificate(t, transformer)
		if certificate.Status != Established || certificate.Mode != "symbolic" || len(certificate.Assumptions) != 1 || len(certificate.Summary.Paths[0].Owned) != 1 {
			t.Fatalf("nontrivial parametric construction/action failed: %+v", certificate)
		}
		result, err := InstantiateAdjective(transformer, certificate, transformerInstance(), AnalysisOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if result.Status != Established {
			t.Fatalf("valid independently supplied target failed: %+v", result.Obligations)
		}
	})
	t.Run("BASE-051", func(t *testing.T) {
		transformer := supportedTransformer()
		transformer.Body.Entry.Children = append(transformer.Body.Entry.Children, Node{Kind: Event, Effects: []Effect{{Kind: "block", Definite: true, Origin: transformer.Origin}}})
		certificate := requireCertificate(t, transformer)
		if certificate.Status != Contradicted || !hasStatus(certificate.Obligations, Contradicted) {
			t.Fatal("violating wrapper received certificate")
		}
		found := false
		for _, obligation := range certificate.Obligations {
			if obligation.Status == Contradicted && obligation.Origin.Source == "wrapper.l" {
				found = true
			}
		}
		if !found {
			t.Fatal("wrapper violation lost attribution")
		}
	})
	t.Run("BASE-052", func(t *testing.T) {
		transformer := supportedTransformer()
		certificate := requireCertificate(t, transformer)
		instance := transformerInstance()
		callback := instance.Callbacks["action"]
		callback.Effects = []Effect{{Kind: "block", Definite: true}}
		instance.Callbacks["action"] = callback
		result, err := InstantiateAdjective(transformer, certificate, instance, AnalysisOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if result.Status != Contradicted {
			t.Fatalf("incompatible callback bypassed certificate premises: %+v", result)
		}
	})
	t.Run("BASE-053", func(t *testing.T) {
		transformer := supportedTransformer()
		certificate := requireCertificate(t, transformer)
		instance := transformerInstance()
		instance.Lowering = []Node{{Kind: Event, Phase: Destruction, Effects: []Effect{{Kind: "block", Phase: Destruction, Definite: true, Origin: Origin{Source: "representation"}}}}}
		result, err := InstantiateAdjective(transformer, certificate, instance, AnalysisOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if result.Status != Contradicted {
			t.Fatal("target-specific destruction escaped final check")
		}
	})
	t.Run("BASE-054", func(t *testing.T) {
		transformer := supportedTransformer()
		transformer.UnsupportedReflection = true
		certificate := requireCertificate(t, transformer)
		if certificate.Mode != "instantiation_only" || certificate.Status != Unresolved {
			t.Fatal("reflection universally certified")
		}
		result, err := InstantiateAdjective(transformer, certificate, transformerInstance(), AnalysisOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if len(result.Summary.Paths) == 0 || result.Certificate.Mode != "instantiation_only" {
			t.Fatal("concrete checks omitted or universal claim fabricated")
		}
	})
	t.Run("BASE-048", func(t *testing.T) {
		transformer := supportedTransformer()
		certificate := requireCertificate(t, transformer)
		instance := transformerInstance()
		instance.Facts = []Predicate{lexicalType("Service")}
		result, err := InstantiateAdjective(transformer, certificate, instance, AnalysisOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if result.Status != Contradicted {
			t.Fatal("changed target assumptions reused success bit")
		}
		certificate.TransformerFingerprint = "stale"
		result, err = InstantiateAdjective(transformer, certificate, transformerInstance(), AnalysisOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if result.Status == Established {
			t.Fatal("stale certificate accepted")
		}
	})
	t.Run("LEX-028", func(t *testing.T) {
		a := Node{Kind: Event, Effects: []Effect{{Kind: "io", Resource: ResourceRef{ID: "log:A"}, Definite: true}}}
		b := Node{Kind: Event, Effects: []Effect{{Kind: "io", Resource: ResourceRef{ID: "log:B"}, Definite: true}}}
		left, right := Plan{Entry: Node{Kind: Sequence, Children: []Node{a, b}}}, Plan{Entry: Node{Kind: Sequence, Children: []Node{b, a}}}
		if equivalentElaboration(left, right) {
			t.Fatal("ordered wrappers treated as set")
		}
		summary, err := Infer(left, AnalysisOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if len(summary.Paths[0].Effects) != 2 || summary.Paths[0].Effects[0].Resource.ID != "log:A" || summary.Paths[0].Effects[1].Resource.ID != "log:B" {
			t.Fatal("effect order changed")
		}
	})
	t.Run("LEX-029", func(t *testing.T) {
		fact := Predicate{Kind: "present", Subject: "constraint"}
		target := lexicalTarget("primary")
		law, err := proveConstraintLaw("left", "right", target, "constraint-commutation", []Predicate{fact})
		if err != nil {
			t.Fatal(err)
		}
		if !law.applies("left", "right", target, []Predicate{fact}) || law.applies("left", "right", lexicalTarget("secondary"), []Predicate{fact}) || law.applies("left", "unrelated", target, []Predicate{fact}) {
			t.Fatal("specific law escaped its target/meaning scope")
		}
		if _, err := proveConstraintLaw("left", "right", target, "constructor-idempotence", nil); err == nil {
			t.Fatal("pure constraint theorem licensed constructive deduplication")
		}
	})
	t.Run("LEX-032", func(t *testing.T) {
		left := Plan{Entry: Node{Kind: Sequence, Origin: Origin{Source: "convenience"}, Children: []Node{{Kind: Event, ID: "convenience:1", Origin: Origin{Source: "convenience"}, Facts: []Predicate{{Kind: "present", Subject: "first"}}}, {Kind: Event, ID: "convenience:2", Origin: Origin{Source: "convenience"}, Facts: []Predicate{{Kind: "present", Subject: "second"}}}}}}
		right := left
		right.Entry = left.Entry
		right.Entry.Origin = Origin{Source: "factored"}
		right.Entry.Children = append([]Node(nil), left.Entry.Children...)
		for i := range right.Entry.Children {
			right.Entry.Children[i].Origin = Origin{Source: "factored"}
			right.Entry.Children[i].ID = "factored"
		}
		if !equivalentElaboration(left, right) {
			t.Fatal("equivalent elaborated behavior failed solely due to origins")
		}
		if left.Entry.Origin.Source != "convenience" || right.Entry.Origin.Source != "factored" {
			t.Fatal("equivalence erased surface provenance")
		}
	})
	t.Run("FALS-009", func(t *testing.T) {
		left := Plan{Entry: Node{Kind: Sequence, Children: []Node{{Kind: Acquire, ID: "one", Meaning: "constructor", Resource: ResourceRef{ID: "one"}}, {Kind: Acquire, ID: "two", Meaning: "constructor", Resource: ResourceRef{ID: "two"}}}}}
		faulty := left
		faulty.Entry.Children = faulty.Entry.Children[:1]
		if equivalentElaboration(left, faulty) {
			t.Fatal("faulty pass collapsed a constructive occurrence")
		}
		summary, err := Infer(left, AnalysisOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if len(summary.Paths[0].Owned) != 2 {
			t.Fatal("analyzer erased repeated constructions")
		}
	})
}
