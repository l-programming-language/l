package contracts

import (
	"reflect"
	"testing"
)

func logicalKeyFixture() LogicalCacheKey {
	return LogicalCacheKey{Schema: SchemaVersion, Meaning: "worker:meaning", Specification: "spec-v1", Substitutions: "substitutions-v1", FacetSchemas: "facets-v1", RoleInterfaces: "role-v1", Rules: AnalyzerVersion, TargetModel: "semantic-model-v1"}
}

func implementationKeyFixture(logical LogicalCacheKey) ImplementationCacheKey {
	return ImplementationCacheKey{Schema: SchemaVersion, Logical: StableID("logical", logical), Implementation: "body-v1", Target: "target-panic-unwind", TrustPolicy: "explicit-test-policy-v1", Witnesses: "witness-v1", Dependencies: map[string]string{"provider": "provider-v1", "drop": "drop-v1"}}
}

func TestSeparateCacheBoundaries(t *testing.T) {
	for _, id := range []string{"BASE-046", "BASE-047", "LEX-031"} {
		t.Run(id, func(t *testing.T) {
			session := NewSession()
			logical := logicalKeyFixture()
			implementation := implementationKeyFixture(logical)
			conditional := EvidenceResult{Status: Established, Reason: "conditional logical theorem", Assumptions: []string{"native-spec"}}
			assurance := EvidenceResult{Status: Established, Reason: "current conformance", Assumptions: []string{"test-only-native"}}
			if err := session.StoreLogical(logical, conditional); err != nil {
				t.Fatal(err)
			}
			if err := session.StoreImplementation(implementation, assurance); err != nil {
				t.Fatal(err)
			}
			var result EvidenceResult
			if found, err := session.LoadImplementation(implementation, &result); err != nil || !found || !reflect.DeepEqual(result, assurance) {
				t.Fatalf("warm lookup = %v, %v, %+v", found, err, result)
			}
			switch id {
			case "BASE-046":
				implementation.Implementation = "body-v2"
			case "BASE-047":
				implementation.Target = "target-panic-abort"
			case "LEX-031":
				implementation.Dependencies = map[string]string{"provider": "provider-v2", "drop": "drop-v2"}
			}
			if found, err := session.LoadImplementation(implementation, &result); err != nil || found {
				t.Fatalf("stale assurance reused: %v, %v", found, err)
			}
			if found, err := session.LoadLogical(logical, &result); err != nil || !found || !reflect.DeepEqual(result, conditional) {
				t.Fatalf("conditional theorem unnecessarily discarded: %v, %v, %+v", found, err, result)
			}
			logical.Specification = "spec-v2"
			if found, err := session.LoadLogical(logical, &result); err != nil || found {
				t.Fatalf("stale logical theorem reused: %v, %v", found, err)
			}
		})
	}
}

func TestBindingCacheRevalidatesFullCandidateEnvironment(t *testing.T) {
	context := SelectionContext{Target: TargetRef{Root: "app", Path: []string{"primary"}, Type: "Theme"}, Environment: "imports-v1", Policy: ReferencePolicy, LexicalRole: "adjective"}
	candidate := Candidate{Meaning: "theme-dark", LexicalRole: "adjective", Applicability: Predicate{Kind: "type", Subject: "target", Value: Atom{Kind: "type", Text: "Theme"}}}
	for _, id := range []string{"LEX-035", "LEX-036", "LEX-037", "FALS-008"} {
		t.Run(id, func(t *testing.T) {
			session := NewSession()
			initial, err := session.SelectReference([]Candidate{candidate}, context)
			if err != nil || initial.Status != Established {
				t.Fatalf("initial selection: %+v, %v", initial, err)
			}
			warm, err := session.SelectReference([]Candidate{candidate}, context)
			if err != nil || !reflect.DeepEqual(initial, warm) || session.CacheStats().Binding.Hits != 1 {
				t.Fatalf("warm selection: %+v, %v", warm, err)
			}
			changed, candidates := context, []Candidate{candidate}
			switch id {
			case "LEX-035", "FALS-008":
				other := candidate
				other.Meaning = "theme-overlap"
				candidates = append(candidates, other)
				changed.Environment = "imports-v2"
			case "LEX-036":
				candidates[0].Applicability = Predicate{Kind: "type", Subject: "target", Value: Atom{Kind: "type", Text: "Backend"}}
			case "LEX-037":
				changed.Target = TargetRef{Root: "app", Path: []string{"secondary"}, Type: "Backend"}
			}
			got, err := session.SelectReference(candidates, changed)
			cold := SelectReference(candidates, changed)
			if err != nil || !reflect.DeepEqual(got, cold) || got.Status == Established {
				t.Fatalf("warm/cold disagreement or stale chosen meaning: %+v / %+v, %v", got, cold, err)
			}
			if initial.DependencyFingerprint == got.DependencyFingerprint {
				t.Fatal("changed candidate dependency fingerprint was reused")
			}
		})
	}
}

func TestCacheIntegrityAndImmutability(t *testing.T) {
	t.Run("FALS-005", func(t *testing.T) {
		for _, fault := range []string{"schema", "key", "payload", "digest"} {
			t.Run(fault, func(t *testing.T) {
				session := NewSession()
				key := implementationKeyFixture(logicalKeyFixture())
				value := EvidenceResult{Status: Established}
				if err := session.StoreImplementation(key, value); err != nil {
					t.Fatal(err)
				}
				identity, err := Fingerprint(key)
				if err != nil {
					t.Fatal(err)
				}
				entry := session.implementation[identity]
				switch fault {
				case "schema":
					entry.Schema = "unsupported"
				case "key":
					entry.Key = "corrupted"
				case "payload":
					entry.Payload = []byte(`{"status":"unresolved"}`)
				case "digest":
					entry.PayloadFingerprint = "corrupted"
				}
				session.implementation[identity] = entry
				var got EvidenceResult
				if found, err := session.LoadImplementation(key, &got); err != nil || found || session.CacheStats().Implementation.Rejected != 1 {
					t.Fatalf("corrupt cache reused: %v, %v, %+v", found, err, got)
				}
			})
		}
	})
	session := NewSession()
	key := logicalKeyFixture()
	value := EvidenceResult{Status: Established, Assumptions: []string{"fixture-only"}}
	if err := session.StoreLogical(key, value); err != nil {
		t.Fatal(err)
	}
	value.Assumptions[0] = "mutated"
	var got EvidenceResult
	if found, err := session.LoadLogical(key, &got); err != nil || !found || got.Assumptions[0] != "fixture-only" {
		t.Fatalf("cache aliased caller memory: %v, %v, %+v", found, err, got)
	}
	got.Assumptions[0] = "mutated-again"
	if found, err := session.LoadLogical(key, &got); err != nil || !found || got.Assumptions[0] != "fixture-only" {
		t.Fatalf("cache leaked mutable output: %v, %v, %+v", found, err, got)
	}
}

func TestLegacyIncompleteRecordRequiresRevalidation(t *testing.T) {
	t.Run("BASE-049", func(t *testing.T) {
		result, err := MigrateLegacyRecord(LegacyContractRecord{Version: "1.0", Definition: "old-definition", Specification: "old-spec"})
		if err != nil || result.Status != Unresolved || !result.BindingRevalidationRequired || !result.ImplementationRevalidationRequired || len(result.Missing) < 4 {
			t.Fatalf("incomplete legacy record certified: %+v, %v", result, err)
		}
		if result.Record.Meaning != "" || result.Record.Environment != "" || result.Record.Target.Root != "" {
			t.Fatal("migration fabricated missing inputs")
		}
	})
	if _, err := MigrateLegacyRecord(LegacyContractRecord{Version: "future"}); err == nil {
		t.Fatal("unsupported legacy version accepted")
	}
}
