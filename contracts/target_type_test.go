package contracts

import "testing"

func TestAdjectiveAuthoritativeTargetTypeBASE048(t *testing.T) {
	transformer := supportedTransformer()
	certificate := requireCertificate(t, transformer)
	for _, test := range []struct {
		name     string
		typeName string
		facts    []Predicate
		want     Status
	}{
		{"changed-actual-target-stale-facts", "Service", []Predicate{lexicalType("App")}, Contradicted},
		{"missing-actual-type-stale-facts", "", []Predicate{lexicalType("App")}, Unresolved},
		{"actual-type-establishes-domain", "App", nil, Established},
		{"contradictory-caller-type", "App", []Predicate{lexicalType("Service")}, Contradicted},
		{"contradictory-conjunctive-type", "App", []Predicate{{Kind: "and", Args: []Predicate{lexicalType("Service")}}}, Contradicted},
		{"contradictory-negative-type", "App", []Predicate{{Kind: "not", Args: []Predicate{lexicalType("App")}}}, Contradicted},
	} {
		t.Run(test.name, func(t *testing.T) {
			instance := transformerInstance()
			instance.Target.Type, instance.Facts = test.typeName, test.facts
			result, err := InstantiateAdjective(transformer, certificate, instance, AnalysisOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if result.Status != test.want {
				t.Fatalf("actual target type %q was replaced by caller facts: got %s, want %s: %+v", test.typeName, result.Status, test.want, result.Obligations)
			}
		})
	}
}

func TestAdjectiveGenericTargetType(t *testing.T) {
	transformer := ContractTransformer{Meaning: "generic", Version: "1", Target: TargetRef{Root: "$target", Phase: Elaboration}, Body: Plan{Entry: Node{Kind: Event}}, Profiles: []Profile{{Name: "NoBlocking"}}}
	instance := AdjectiveInstance{Target: TargetRef{Root: "actual", Phase: Elaboration}}
	certificate := requireCertificate(t, transformer)
	result, err := InstantiateAdjective(transformer, certificate, instance, AnalysisOptions{})
	if err != nil || result.Status != Established {
		t.Fatalf("unconstrained generic target was rejected: %+v, %v", result, err)
	}
	transformer.Requires = []Predicate{lexicalType("App")}
	certificate = requireCertificate(t, transformer)
	result, err = InstantiateAdjective(transformer, certificate, instance, AnalysisOptions{})
	if err != nil || result.Status != Unresolved {
		t.Fatalf("missing required type was inferred: %+v, %v", result, err)
	}
	instance.Target.Type = "App"
	result, err = InstantiateAdjective(transformer, certificate, instance, AnalysisOptions{})
	if err != nil || result.Status != Established {
		t.Fatalf("authoritative concrete type did not discharge generic requirement: %+v, %v", result, err)
	}
}

func TestReferenceAuthoritativeTargetType(t *testing.T) {
	candidates := []Candidate{lexicalCandidate("theme", lexicalType("Theme")), lexicalCandidate("backend", lexicalType("Backend"))}
	context := SelectionContext{Target: lexicalTarget("primary"), Facts: []Predicate{lexicalType("Backend")}}
	selection := SelectReference(candidates, context)
	if selection.Status != Contradicted || selection.Meaning != "" || len(selection.Candidates) != 2 {
		t.Fatalf("caller fact selected the wrong typed meaning: %+v", selection)
	}
	context.Target.Type = ""
	selection = SelectReference(candidates, context)
	if selection.Status != Unresolved || selection.Meaning != "" {
		t.Fatalf("caller assertion supplied a missing target type: %+v", selection)
	}
	context.Target.Type, context.Facts = "Theme", nil
	selection = SelectReference(candidates, context)
	if selection.Status != Established || selection.Meaning != "theme" {
		t.Fatalf("actual typed candidate not selected: %+v", selection)
	}
}

func TestProviderAuthoritativeTargetType(t *testing.T) {
	role := RoleInterface{ID: "app-role", Version: "1", Requires: []Predicate{lexicalType("App")}, Operations: []RoleOperation{{Name: "perform", Profiles: []Profile{{Name: "NoBlocking"}}}}}
	provider := ProviderImplementation{ID: "provider", Target: TargetRef{Root: "actual", Type: "Service"}, Operations: map[string]Plan{"perform": {Entry: Node{Kind: Event}}}}
	for _, test := range []struct {
		typeName string
		want     Status
	}{{"Service", Contradicted}, {"", Unresolved}, {"App", Established}} {
		provider.Target.Type = test.typeName
		result, err := CheckProvider(role, provider, []Predicate{lexicalType("App")}, AnalysisOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if got := obligationStatus(result.Obligations); got != test.want {
			t.Fatalf("provider type %q: got %s, want %s: %+v", test.typeName, got, test.want, result.Obligations)
		}
	}
}
