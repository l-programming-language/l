package contracts

import (
	"reflect"
	"testing"
)

func TestLogicalCacheAllSemanticDependencies(t *testing.T) {
	t.Run("BASE-047", func(t *testing.T) {
		for _, component := range []string{"fact", "parameter", "facet", "role", "semantic-target", "rule", "contract"} {
			t.Run(component, func(t *testing.T) {
				session := NewSession()
				key := logicalKeyFixture()
				conditional := EvidenceResult{Status: Established, Assumptions: []string{"declared-interface"}}
				if err := session.storeLogical(key, conditional); err != nil {
					t.Fatal(err)
				}
				var got EvidenceResult
				if found, err := session.loadLogical(key, &got); err != nil || !found || !reflect.DeepEqual(got, conditional) {
					t.Fatalf("initial logical cache: %v %v %+v", found, err, got)
				}
				switch component {
				case "fact":
					key.Substitutions = StableID("facts", []Predicate{{Kind: "eq", Subject: "available", Value: Atom{Kind: "bool", Bool: false}}})
				case "parameter":
					key.Substitutions = StableID("parameters", map[string]string{"T": "DifferentType"})
				case "facet":
					key.FacetSchemas = "facets-v2"
				case "role":
					key.RoleInterfaces = "role-v2"
				case "semantic-target":
					key.TargetModel = "semantic-model-v2"
				case "rule":
					key.Rules = "checker-v2"
				case "contract":
					key.Specification = "spec-v2"
				}
				if found, err := session.loadLogical(key, &got); err != nil || found {
					t.Fatalf("changed %s reused theorem: found=%v error=%v", component, found, err)
				}
			})
		}
	})
}

func TestReferenceCacheCompetingCandidateBecomesApplicable(t *testing.T) {
	t.Run("LEX-036", func(t *testing.T) {
		context := SelectionContext{Target: TargetRef{Root: "app", Path: []string{"primary"}, Type: "Theme"}, Environment: "same-imports", Policy: ReferencePolicy, LexicalRole: "adjective"}
		candidates := []Candidate{
			{Meaning: "theme", LexicalRole: "adjective", Applicability: Predicate{Kind: "type", Subject: "target", Value: Atom{Kind: "type", Text: "Theme"}}},
			{Meaning: "competing", LexicalRole: "adjective", Applicability: Predicate{Kind: "type", Subject: "target", Value: Atom{Kind: "type", Text: "Backend"}}},
		}
		session := NewSession()
		before, err := session.SelectReference(candidates, context)
		if err != nil || before.Status != Established || before.Meaning != "theme" {
			t.Fatalf("initial candidate set: %+v %v", before, err)
		}
		candidates[1].Applicability.Value.Text = "Theme"
		warm, err := session.SelectReference(candidates, context)
		cold := SelectReference(candidates, context)
		if err != nil || !reflect.DeepEqual(warm, cold) || warm.Status != Contradicted || before.DependencyFingerprint == warm.DependencyFingerprint {
			t.Fatalf("existing inapplicable candidate retained stale selection: warm=%+v cold=%+v error=%v", warm, cold, err)
		}
	})
}

func TestReferenceCacheRolePhraseAndPolicy(t *testing.T) {
	t.Run("LEX-037", func(t *testing.T) {
		for _, mutation := range []string{"role", "type", "phrase", "policy", "fact"} {
			t.Run(mutation, func(t *testing.T) {
				context := SelectionContext{Target: TargetRef{Root: "app", Path: []string{"primary"}, Type: "Theme"}, Environment: "imports", Policy: ReferencePolicy, LexicalRole: "adjective"}
				candidates := []Candidate{{Meaning: "theme", Phrase: "Bright Theme", LexicalRole: "adjective", Applicability: Predicate{Kind: "type", Subject: "target", Value: Atom{Kind: "type", Text: "Theme"}}}}
				session := NewSession()
				before, err := session.SelectReference(candidates, context)
				if err != nil || before.Status != Established {
					t.Fatalf("initial selection %+v %v", before, err)
				}
				switch mutation {
				case "role":
					context.Target.Path = []string{"secondary"}
				case "type":
					context.Target.Type = "Backend"
				case "phrase":
					candidates[0].Phrase = "Different Phrase"
				case "policy":
					context.Policy = "unapproved-selector/2"
				case "fact":
					context.Facts = []Predicate{{Kind: "eq", Subject: "feature", Value: Atom{Kind: "bool", Bool: true}}}
				}
				warm, err := session.SelectReference(candidates, context)
				cold := SelectReference(candidates, context)
				if err != nil || !reflect.DeepEqual(warm, cold) || warm.DependencyFingerprint == before.DependencyFingerprint || session.CacheStats().Binding.Misses != 2 {
					t.Fatalf("changed %s did not revalidate: warm=%+v cold=%+v error=%v stats=%+v", mutation, warm, cold, err, session.CacheStats())
				}
			})
		}
	})
}
