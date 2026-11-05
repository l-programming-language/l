package contracts

import (
	"bytes"
	"reflect"
	"testing"
)

func logicalKeyFixture() logicalCacheKey {
	return logicalCacheKey{Schema: SchemaVersion, Meaning: "worker:meaning", Specification: "spec-v1", Substitutions: "substitutions-v1", FacetSchemas: "facets-v1", RoleInterfaces: "role-v1", Rules: AnalyzerVersion, TargetModel: "semantic-model-v1"}
}

func implementationKeyFixture(logical logicalCacheKey) implementationCacheKey {
	return implementationCacheKey{Schema: SchemaVersion, Logical: StableID("logical", logical), Implementation: "body-v1", Target: "target-panic-unwind", TrustPolicy: "explicit-test-policy-v1", Witnesses: "witness-v1", Dependencies: map[string]string{"provider": "provider-v1", "drop": "drop-v1"}}
}

func TestSeparateCacheBoundaries(t *testing.T) {
	for _, id := range []string{"BASE-046", "BASE-047", "LEX-031"} {
		t.Run(id, func(t *testing.T) {
			session := NewSession()
			logical := logicalKeyFixture()
			implementation := implementationKeyFixture(logical)
			conditional := EvidenceResult{Status: Established, Reason: "conditional logical theorem", Assumptions: []string{"native-spec"}}
			assurance := EvidenceResult{Status: Established, Reason: "current conformance", Assumptions: []string{"test-only-native"}}
			if err := session.store(logical, conditional); err != nil {
				t.Fatal(err)
			}
			if err := session.store(implementation, assurance); err != nil {
				t.Fatal(err)
			}
			var result EvidenceResult
			if found, err := session.load(implementation, &result); err != nil || !found || !reflect.DeepEqual(result, assurance) {
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
			if found, err := session.load(implementation, &result); err != nil || found {
				t.Fatalf("stale assurance reused: %v, %v", found, err)
			}
			if found, err := session.load(logical, &result); err != nil || !found || !reflect.DeepEqual(result, conditional) {
				t.Fatalf("conditional theorem unnecessarily discarded: %v, %v, %+v", found, err, result)
			}
			logical.Specification = "spec-v2"
			if found, err := session.load(logical, &result); err != nil || found {
				t.Fatalf("stale logical theorem reused: %v, %v", found, err)
			}
		})
	}
}

func TestBindingCacheRevalidatesFullCandidateEnvironment(t *testing.T) {
	context := SelectionContext{Target: TargetRef{Root: "app", Path: []string{"primary"}, Type: "Theme"}, Environment: "imports-v1", Policy: ReferencePolicy, LexicalRole: "adjective"}
	candidate := Candidate{Meaning: "theme-dark", LexicalRole: "adjective", Applicability: Predicate{Kind: "type", Subject: "target", Value: Atom{Kind: "type", Text: "Theme"}}}
	for _, scenario := range []string{"overlapping-candidate", "changed-applicability", "changed-target"} {
		t.Run(scenario, func(t *testing.T) {
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
			switch scenario {
			case "overlapping-candidate":
				other := candidate
				other.Meaning = "theme-overlap"
				candidates = append(candidates, other)
				changed.Environment = "imports-v2"
			case "changed-applicability":
				candidates[0].Applicability = Predicate{Kind: "type", Subject: "target", Value: Atom{Kind: "type", Text: "Backend"}}
			case "changed-target":
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

func TestCacheInvalidationAndImmutability(t *testing.T) {
	t.Run("FALS-005", func(t *testing.T) {
		for _, changed := range []string{"schema", "implementation", "target", "binding"} {
			t.Run(changed, func(t *testing.T) {
				session := NewSession()
				key := implementationKeyFixture(logicalKeyFixture())
				if err := session.store(key, EvidenceResult{Status: Established}); err != nil {
					t.Fatal(err)
				}
				switch changed {
				case "schema":
					key.Schema = "1.0"
				case "implementation":
					key.Implementation = "changed-body"
				case "target":
					key.Target = "changed-target"
				case "binding":
					key.Dependencies["binding"] = "changed-selection"
				}
				var got EvidenceResult
				found, err := session.load(key, &got)
				if found || got.Status == Established || (err != nil) != (changed == "schema") {
					t.Fatalf("stale assurance reused: found=%v error=%v result=%+v", found, err, got)
				}
			})
		}
	})
	session := NewSession()
	key := logicalKeyFixture()
	value := EvidenceResult{Status: Established, Assumptions: []string{"fixture-only"}}
	if err := session.store(key, value); err != nil {
		t.Fatal(err)
	}
	value.Assumptions[0] = "mutated"
	var got EvidenceResult
	if found, err := session.load(key, &got); err != nil || !found || got.Assumptions[0] != "fixture-only" {
		t.Fatalf("cache aliased caller memory: %v, %v, %+v", found, err, got)
	}
	got.Assumptions[0] = "mutated-again"
	if found, err := session.load(key, &got); err != nil || !found || got.Assumptions[0] != "fixture-only" {
		t.Fatalf("cache leaked mutable output: %v, %v, %+v", found, err, got)
	}
}

func TestUnsupportedOrIncompleteRecordsRequireFreshAnalysis(t *testing.T) {
	t.Run("BASE-049", func(t *testing.T) {
		data, err := EncodeDocument(nativeManifestFixture())
		if err != nil {
			t.Fatal(err)
		}
		old := bytes.Replace(data, []byte(`"schema_version": "1.1"`), []byte(`"schema_version": "1.0"`), 1)
		if _, err := DecodeDocument(old); err == nil {
			t.Fatal("unsupported predecessor metadata accepted")
		}
		session := NewSession()
		key := logicalKeyFixture()
		key.Substitutions = ""
		if err := session.store(key, EvidenceResult{Status: Established}); err == nil {
			t.Fatal("incomplete logical dependencies admitted")
		}
		var stale EvidenceResult
		if found, err := session.load(key, &stale); found || err == nil || stale.Status == Established {
			t.Fatalf("incomplete record manufactured proof: found=%v error=%v result=%+v", found, err, stale)
		}
		plan := Plan{Entry: Node{Kind: Native, Native: "opaque"}}
		want := reviewInfer(t, plan, AnalysisOptions{})
		got, err := session.Infer(plan, AnalysisOptions{}, nil, LoweringInventory{})
		if err != nil || !reflect.DeepEqual(got, want) || len(got.Unknowns) == 0 {
			t.Fatalf("fresh analysis fabricated missing assurance: result=%+v want=%+v error=%v", got, want, err)
		}
		if stats := session.CacheStats(); stats.Logical.Misses != 1 || stats.Implementation.Misses != 1 {
			t.Fatalf("fresh analysis was not computed: %+v", stats)
		}
	})
}

func TestSessionBindingRevalidation(t *testing.T) {
	key := BindingCacheKey{Schema: SchemaVersion, SourceStructure: "source", Environment: "definition-and-use-environment", Candidates: "candidates", Applicability: "applicability", Target: TargetRef{Root: "exact-target"}, LexicalRole: "adjective", PhrasePolicy: "existing-source-parser/1", ResolutionPolicy: "track-a"}
	binding := ResolvedBinding{Occurrence: "application", Meaning: "selected-meaning", Target: key.Target}
	session := NewSession()
	if err := session.RevalidateBinding(key, binding); err != nil {
		t.Fatal(err)
	}
	if err := session.RevalidateBinding(key, binding); err != nil || session.CacheStats().Binding.Hits != 1 {
		t.Fatalf("consistent warm resolution rejected: %v", err)
	}
	binding.Meaning = "different-meaning"
	if err := session.RevalidateBinding(key, binding); err == nil {
		t.Fatal("conflicting current resolution accepted under unchanged key")
	}
	key.Candidates = "changed-candidates"
	if err := session.RevalidateBinding(key, binding); err != nil {
		t.Fatalf("newly resolved environment was rejected: %v", err)
	}
}
