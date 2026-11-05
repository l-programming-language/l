package contracts

import (
	"reflect"
	"strings"
	"testing"
)

func facetRuleFixture() (FacetSchema, FacetSlot, ConfigurationContribution, Predicate) {
	schema := paletteSchema("exactly-one-at-close")
	slot := FacetSlot{Facet: schema.ID, Target: lexicalTarget("primary"), Closed: true}
	selection := paletteContribution(slot.Target, "dark")
	return schema, slot, selection, FacetFact(slot, selection.Value)
}

func TestFacetDerivedClosureAndProvenance(t *testing.T) {
	schema, slot, _, fact := facetRuleFixture()
	rules := []FacetRule{{Consequences: []Predicate{fact}, Origin: Origin{Source: "first"}}, {When: []Predicate{fact}, Consequences: []Predicate{fact}, Requires: []Predicate{fact}, Origin: Origin{Source: "cycle"}}}
	result := requireConfiguration(t, []FacetSchema{schema}, []FacetSlot{slot}, nil, rules)
	if !AllEstablished(result.Obligations) || len(result.Selections) != 1 || len(result.Selections[0].Values) != 1 || len(result.Selections[0].Origins) != 2 || len(result.Facts) != 1 {
		t.Fatalf("validated derivation failed closure or lost repeated rule origins: %+v", result)
	}
	reversed := requireConfiguration(t, []FacetSchema{schema}, []FacetSlot{slot}, nil, []FacetRule{rules[1], rules[0]})
	if !AllEstablished(reversed.Obligations) || !reflect.DeepEqual(result.Facts, reversed.Facts) {
		t.Fatalf("rule order changed finite closure: %+v", reversed)
	}
}

func TestFacetDirectDerivedValidationParity(t *testing.T) {
	for _, name := range []string{"domain", "authorization", "unknown-target", "inapplicable", "unknown-applicability"} {
		t.Run(name, func(t *testing.T) {
			schema, slot, selection, _ := facetRuleFixture()
			switch name {
			case "domain":
				selection.Value.Text = "not-in-domain"
			case "authorization":
				schema.AllowedContributions = []string{"require"}
			case "unknown-target":
				selection.Target = TargetRef{Unknown: true}
			case "inapplicable":
				schema.Applicability = Predicate{Kind: "false"}
			case "unknown-applicability":
				schema.Applicability = Predicate{Kind: "capability", Subject: "unproved"}
			}
			fact := FacetFact(FacetSlot{Facet: selection.Facet, Target: selection.Target, Scope: selection.Scope}, selection.Value)
			direct, directErr := CheckConfiguration([]FacetSchema{schema}, []FacetSlot{slot}, []ConfigurationContribution{selection}, nil, nil)
			derived, derivedErr := CheckConfiguration([]FacetSchema{schema}, []FacetSlot{slot}, nil, []FacetRule{{Consequences: []Predicate{fact}}}, nil)
			if (directErr == nil) != (derivedErr == nil) {
				t.Fatalf("direct/rule error mismatch: %v / %v", directErr, derivedErr)
			}
			if directErr == nil && (AllEstablished(direct.Obligations) || AllEstablished(derived.Obligations) || ObligationStatus(direct.Obligations) != ObligationStatus(derived.Obligations)) {
				t.Fatalf("direct/rule validation mismatch: %+v / %+v", direct, derived)
			}
		})
	}
}

func TestFacetConditionsAreNotIgnored(t *testing.T) {
	for _, mode := range []string{"direct", "rule"} {
		for _, condition := range []string{"true", "false", "capability"} {
			t.Run(mode+"/"+condition, func(t *testing.T) {
				schema, slot, dark, _ := facetRuleFixture()
				light := paletteContribution(slot.Target, "light")
				premise := Predicate{Kind: condition, Subject: "condition"}
				contributions := []ConfigurationContribution{dark}
				var rules []FacetRule
				if mode == "direct" {
					light.Condition = []Predicate{premise}
					contributions = append(contributions, light)
				} else {
					rules = []FacetRule{{When: []Predicate{premise}, Consequences: []Predicate{FacetFact(slot, light.Value)}}}
				}
				result := requireConfiguration(t, []FacetSchema{schema}, []FacetSlot{slot}, contributions, rules)
				want := map[string]Status{"true": Contradicted, "false": Established, "capability": Unresolved}[condition]
				if got := ObligationStatus(result.Obligations); got != want {
					t.Fatalf("condition %s produced %s, want %s: %+v", condition, got, want, result)
				}
				if condition != "true" && len(result.Selections[0].Values) != 1 {
					t.Fatal("unestablished selection was applied")
				}
			})
		}
	}
}

func TestFacetRulesUseFinalOrderedSelections(t *testing.T) {
	schema, slot, dark, darkFact := facetRuleFixture()
	light := paletteContribution(slot.Target, "light")
	light.Kind = "replace"
	feature := FacetSchema{ID: "feature", Version: "1", Domain: "bool", Cardinality: "optional-single", AllowedContributions: []string{"select"}, ClosurePoint: "constructed"}
	featureFact := FacetFact(FacetSlot{Facet: feature.ID, Target: slot.Target}, Atom{Kind: "bool", Bool: true})
	result := requireConfiguration(t, []FacetSchema{schema, feature}, []FacetSlot{slot}, []ConfigurationContribution{dark, light}, []FacetRule{{When: []Predicate{darkFact}, Consequences: []Predicate{featureFact}}})
	if len(result.Replacements) != 1 || len(result.Facts) != 1 || result.Facts[0].Value != light.Value || hasStatus(result.Obligations, Contradicted) {
		t.Fatalf("rule revived an obsolete ordered selection: %+v", result)
	}
	conflict := requireConfiguration(t, []FacetSchema{schema}, []FacetSlot{slot}, []ConfigurationContribution{dark, light}, []FacetRule{{Consequences: []Predicate{darkFact}}})
	if !hasStatus(conflict.Obligations, Contradicted) || len(conflict.Replacements) != 1 || len(conflict.Selections[0].Values) != 2 {
		t.Fatalf("rule performed an implicit replacement: %+v", conflict)
	}
}

func TestFacetOrderedConditionsSeePriorSelections(t *testing.T) {
	schema, slot, dark, fact := facetRuleFixture()
	light := paletteContribution(slot.Target, "light")
	light.Kind, light.Condition = "replace", []Predicate{fact}
	result := requireConfiguration(t, []FacetSchema{schema}, []FacetSlot{slot}, []ConfigurationContribution{dark, light}, nil)
	if len(result.Conditional) != 0 || len(result.Replacements) != 1 || len(result.Selections[0].Values) != 1 || result.Selections[0].Values[0] != light.Value {
		t.Fatalf("ordered condition ignored its predecessor: %+v", result)
	}
	featureFact := FacetFact(FacetSlot{Facet: schema.ID, Target: lexicalTarget("secondary")}, dark.Value)
	light.Condition = []Predicate{featureFact}
	pending := requireConfiguration(t, []FacetSchema{schema}, []FacetSlot{slot}, []ConfigurationContribution{dark, light}, []FacetRule{{Consequences: []Predicate{featureFact}}})
	if len(pending.Conditional) != 1 || len(pending.Replacements) != 0 || !hasStatus(pending.Obligations, Unresolved) {
		t.Fatalf("rule replayed an earlier uncertain replacement: %+v", pending)
	}
}

func TestFacetRuleApplicabilityReachesFixedPoint(t *testing.T) {
	schema, slot, _, fact := facetRuleFixture()
	feature := FacetSchema{ID: "feature", Version: "1", Domain: "bool", Cardinality: "optional-single", AllowedContributions: []string{"select"}, ClosurePoint: "constructed"}
	featureFact := FacetFact(FacetSlot{Facet: feature.ID, Target: slot.Target}, Atom{Kind: "bool", Bool: true})
	schema.Applicability = featureFact
	result := requireConfiguration(t, []FacetSchema{schema, feature}, []FacetSlot{slot}, nil, []FacetRule{{Consequences: []Predicate{fact}}, {Consequences: []Predicate{featureFact}}})
	if !AllEstablished(result.Obligations) || len(result.Facts) != 2 {
		t.Fatalf("early unresolved applicability survived the fixed point: %+v", result)
	}
}

func TestFacetSeededSelectionsUseValidatedLedger(t *testing.T) {
	schema, slot, dark, fact := facetRuleFixture()
	for _, wrap := range []bool{false, true} {
		seed := fact
		if wrap {
			seed = Predicate{Kind: "and", Args: []Predicate{fact}}
		}
		result, err := CheckConfiguration([]FacetSchema{schema}, []FacetSlot{slot}, nil, nil, []Predicate{seed})
		if err != nil || !AllEstablished(result.Obligations) || len(result.Selections) != 1 || len(result.Selections[0].Values) != 1 || len(result.Facts) != 1 {
			t.Fatalf("seed bypassed ledger or closure: %+v, %v", result, err)
		}
		light := paletteContribution(slot.Target, "light")
		conflict, err := CheckConfiguration([]FacetSchema{schema}, nil, []ConfigurationContribution{light}, nil, []Predicate{seed})
		if err != nil || !hasStatus(conflict.Obligations, Contradicted) {
			t.Fatalf("seed conflict was ignored: %+v, %v", conflict, err)
		}
	}
	negative := Predicate{Kind: "not", Args: []Predicate{fact}}
	conflict, err := CheckConfiguration([]FacetSchema{schema}, nil, []ConfigurationContribution{dark}, nil, []Predicate{negative})
	if err != nil || !hasStatus(conflict.Obligations, Contradicted) {
		t.Fatalf("negative configuration premise was ignored: %+v, %v", conflict, err)
	}
	unknown := fact
	unknown.Target = TargetRef{Unknown: true}
	result, err := CheckConfiguration([]FacetSchema{schema}, nil, nil, nil, []Predicate{{Kind: "not", Args: []Predicate{unknown}}})
	if err != nil || !hasStatus(result.Obligations, Unresolved) || len(result.Facts) != 0 {
		t.Fatalf("unknown nested facet premise became established: %+v, %v", result, err)
	}
	for _, malformed := range []Predicate{
		FacetFact(FacetSlot{Facet: "missing", Target: slot.Target}, dark.Value),
		FacetFact(slot, Atom{Kind: "enum", Text: "invalid"}),
	} {
		for _, wrapper := range []string{"and", "not", "or", "implies"} {
			_, err := CheckConfiguration([]FacetSchema{schema}, nil, nil, nil, []Predicate{{Kind: wrapper, Args: []Predicate{malformed}}})
			if err == nil {
				t.Fatalf("%s hid invalid seeded facet: %+v", wrapper, malformed)
			}
		}
	}
}

func TestFacetLegacyIdentityAndMalformedRules(t *testing.T) {
	schema, slot, _, fact := facetRuleFixture()
	schema.ID, slot.Facet, slot.Scope = "family:palette", "family:palette", "normal:exit"
	fact = FacetFact(slot, fact.Value)
	result := requireConfiguration(t, []FacetSchema{schema}, []FacetSlot{slot}, nil, []FacetRule{{Consequences: []Predicate{fact}}})
	if !AllEstablished(result.Obligations) || result.Selections[0].Slot.Scope != slot.Scope || result.Selections[0].Slot.Facet != schema.ID {
		t.Fatalf("unambiguous legacy identity did not round trip: %+v", result)
	}
	prefix := schema
	prefix.ID = "family"
	if _, err := CheckConfiguration([]FacetSchema{schema, prefix}, nil, nil, []FacetRule{{Consequences: []Predicate{fact}}}, nil); err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("ambiguous legacy subject accepted: %v", err)
	}
	direct := ConfigurationContribution{Facet: schema.ID, Scope: slot.Scope, Target: slot.Target, Kind: "select", Value: fact.Value}
	if _, err := CheckConfiguration([]FacetSchema{schema, prefix}, nil, []ConfigurationContribution{direct}, nil, nil); err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("direct selection emitted an ambiguous legacy fact: %v", err)
	}
	for _, field := range []string{"kind", "resource", "args", "lower", "unknown-facet"} {
		bad := fact
		switch field {
		case "kind":
			bad.Kind = "effects_closed"
		case "resource":
			bad.Resource = ResourceRef{ID: "not-configuration"}
		case "args":
			bad.Args = []Predicate{{Kind: "true"}}
		case "lower":
			zero := int64(0)
			bad.Lower = &zero
		case "unknown-facet":
			bad.Subject = "facet:missing:"
		}
		if _, err := CheckConfiguration([]FacetSchema{schema}, nil, nil, []FacetRule{{When: []Predicate{{Kind: "false"}}, Consequences: []Predicate{bad}}}, nil); err == nil {
			t.Fatalf("false rule concealed malformed %s consequence", field)
		}
	}
}

func TestFacetDerivedDistinctRolesAndDeterminism(t *testing.T) {
	schema, slot, _, fact := facetRuleFixture()
	other := slot
	other.Target = lexicalTarget("secondary")
	rules := []FacetRule{{Consequences: []Predicate{fact, FacetFact(other, Atom{Kind: "enum", Text: "light"})}}}
	first := requireConfiguration(t, []FacetSchema{schema}, []FacetSlot{slot, other}, nil, rules)
	second := requireConfiguration(t, []FacetSchema{schema}, []FacetSlot{other, slot}, nil, rules)
	if !AllEstablished(first.Obligations) || len(first.Selections) != 2 || !reflect.DeepEqual(first, second) {
		t.Fatalf("same-type roles conflated or result nondeterministic: %+v / %+v", first, second)
	}
}

func TestFacetRulesMetadataAndInstantiation(t *testing.T) {
	schema, slot, _, fact := facetRuleFixture()
	transformer := ContractTransformer{Meaning: "fixture:facet-rule", Version: "1", Target: lexicalTarget(), Body: Plan{Entry: Node{Kind: Event}}, Facets: []FacetSchema{schema}, Slots: []FacetSlot{slot}, Rules: []FacetRule{{Consequences: []Predicate{fact}, Requires: []Predicate{fact}}}}
	manifest := FacetRoleManifest{SchemaVersion: SchemaVersion, Kind: "facet_and_role_manifest", Producer: "test-only", Transformers: []ContractTransformer{transformer}}
	data, err := EncodeDocument(manifest)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeDocument(data)
	if err != nil {
		t.Fatal(err)
	}
	loaded, ok := decoded.(FacetRoleManifest)
	if !ok || len(loaded.Transformers) != 1 {
		t.Fatalf("legacy facet-rule manifest changed shape: %+v", decoded)
	}
	transformer = loaded.Transformers[0]
	certificate, err := VerifyAdjective(transformer, AnalysisOptions{})
	if err != nil || certificate.Status != Established {
		t.Fatalf("decoded rule could not certify configuration: %+v, %v", certificate, err)
	}
	instance := AdjectiveInstance{Target: TargetRef{Root: "concrete", Path: []string{"role"}, Type: slot.Target.Type, Phase: slot.Target.Phase}}
	requireInstantiation(t, transformer, certificate, instance, Established)
	concrete := rebaseTransformer(transformer, instance.Target)
	configuration := requireConfiguration(t, concrete.Facets, concrete.Slots, concrete.Contributions, concrete.Rules)
	expected := instance.Target
	expected.Path = append(append([]string(nil), instance.Target.Path...), slot.Target.Path...)
	if len(configuration.Facts) != 1 || TargetRelation(configuration.Facts[0].Target, expected) != Established {
		t.Fatalf("rule retained symbolic target: %+v", configuration)
	}
	transformer.Rules = append(transformer.Rules, FacetRule{Consequences: []Predicate{FacetFact(slot, Atom{Kind: "enum", Text: "light"})}})
	stale := requireInstantiation(t, transformer, certificate, instance, Contradicted)
	cold, err := VerifyAdjective(transformer, AnalysisOptions{})
	if err != nil || cold.Status != stale.Status {
		t.Fatalf("cold and stale-certificate verification disagreed: %+v / %+v, %v", cold, stale, err)
	}
}

func TestFacetReplacementInvalidatesNestedConfigurationPremises(t *testing.T) {
	schema, slot, dark, darkFact := facetRuleFixture()
	light := paletteContribution(slot.Target, "light")
	dark.Kind = "replace"
	negative := Predicate{Kind: "not", Args: []Predicate{darkFact}}
	other := slot
	other.Target = lexicalTarget("secondary")
	unrelated := Predicate{Kind: "not", Args: []Predicate{FacetFact(other, dark.Value)}}
	result, err := CheckConfiguration([]FacetSchema{schema}, []FacetSlot{slot}, []ConfigurationContribution{light, dark}, []FacetRule{{When: []Predicate{negative}, Consequences: []Predicate{FacetFact(other, light.Value)}}}, []Predicate{negative, unrelated})
	if err != nil || hasStatus(result.Obligations, Contradicted) || len(result.Replacements) != 1 {
		t.Fatalf("old negative premise contradicted authorized replacement: %+v, %v", result, err)
	}
	if len(result.Facts) != 2 || result.Facts[0].Kind != "not" || TargetRelation(result.Facts[0].Args[0].Target, other.Target) != Established || len(result.Selections) != 1 {
		t.Fatalf("stale negative premise fired rule or unrelated fact was lost: %+v", result)
	}
}
