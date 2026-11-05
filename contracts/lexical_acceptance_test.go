package contracts

import (
	"reflect"
	"testing"
)

func lexicalTarget(path ...string) TargetRef {
	return TargetRef{Root: "app:A", Path: path, Type: "Theme", Phase: Elaboration}
}
func lexicalType(name string) Predicate {
	return Predicate{Kind: "type", Subject: "target", Value: Atom{Kind: "type", Text: name}}
}
func lexicalCandidate(meaning MeaningKey, predicate Predicate) Candidate {
	return Candidate{Meaning: meaning, Applicability: predicate, LexicalRole: "adjective", Origins: []Origin{{Source: string(meaning)}}}
}
func lexicalSelection(candidates ...Candidate) ReferenceSelection {
	return SelectReference(candidates, SelectionContext{Target: lexicalTarget("primary"), LexicalRole: "adjective", Environment: "environment:complete"})
}

func TestLexicalAcceptance(t *testing.T) {
	t.Run("LEX-001", func(t *testing.T) {
		candidates := []Candidate{lexicalCandidate("theme::Light", lexicalType("Theme")), lexicalCandidate("client::Light", lexicalType("Client"))}
		for _, kind := range []string{"Theme", "Client"} {
			target := lexicalTarget("primary")
			target.Type = kind
			selected := SelectReference(candidates, SelectionContext{Target: target, Environment: "both"})
			want := MeaningKey("theme::Light")
			if kind == "Client" {
				want = "client::Light"
			}
			if selected.Status != Established || selected.Meaning != want {
				t.Fatalf("wrong typed meaning: %+v", selected)
			}
		}
	})
	t.Run("LEX-002", func(t *testing.T) {
		a, b := lexicalCandidate("a::Light", lexicalType("Theme")), lexicalCandidate("b::Light", lexicalType("Theme"))
		first, second := lexicalSelection(a, b), lexicalSelection(b, a)
		if first.Status != Contradicted || first.Meaning != "" || !reflect.DeepEqual(first, second) {
			t.Fatalf("ambiguity depends on enumeration order: %+v %+v", first, second)
		}
	})
	t.Run("LEX-003", func(t *testing.T) {
		result := lexicalSelection(lexicalCandidate("generic", Predicate{Kind: "true"}), lexicalCandidate("special", lexicalType("Theme")))
		if result.Status != Contradicted || result.Meaning != "" {
			t.Fatalf("specialization silently won: %+v", result)
		}
	})
	t.Run("LEX-004", func(t *testing.T) {
		result := lexicalSelection(lexicalCandidate("known", lexicalType("Theme")), lexicalCandidate("unknown", Predicate{Kind: "reflection"}))
		if result.Status != Unresolved || len(result.Candidates) != 2 || result.Meaning != "" {
			t.Fatalf("unknown candidate disappeared: %+v", result)
		}
	})
	t.Run("LEX-005", func(t *testing.T) {
		one := lexicalCandidate("canonical", lexicalType("Theme"))
		two := one
		two.Origins = []Origin{{Source: "alias"}}
		result := lexicalSelection(one, two)
		if result.Status != Established || result.Meaning != "canonical" || len(result.Candidates) != 2 {
			t.Fatalf("aliases competed as meanings: %+v", result)
		}
	})
	t.Run("LEX-006", func(t *testing.T) {
		candidates := []Candidate{lexicalCandidate("selected", lexicalType("Theme")), lexicalCandidate("quiet", lexicalType("Client"))}
		before := lexicalSelection(candidates...)
		summary := reviewInfer(t, Plan{Entry: Node{Kind: Event, Effects: []Effect{{Kind: "block", Definite: true}}}}, AnalysisOptions{})
		if AllEstablished(CheckProfile(summary, Profile{Name: "NoBlocking"})) {
			t.Fatal("blocking selected implementation passed")
		}
		after := lexicalSelection(candidates...)
		if !reflect.DeepEqual(before, after) || after.Meaning != "selected" {
			t.Fatal("profile changed binding")
		}
	})
	t.Run("LEX-008", func(t *testing.T) {
		primary, secondary := lexicalTarget("theme", "primary"), lexicalTarget("theme", "secondary")
		if TargetRelation(primary, secondary) != Contradicted {
			t.Fatal("same type collapsed roles")
		}
		candidate := lexicalCandidate("modifier", lexicalType("Theme"))
		result := SelectReference([]Candidate{candidate}, SelectionContext{Target: TargetRef{Root: "app:A", Type: "Theme", Unknown: true}})
		if result.Status != Unresolved {
			t.Fatal("ambiguous attachment guessed a role")
		}
	})
	t.Run("LEX-009", func(t *testing.T) {
		left, right := lexicalTarget("outer", "primary"), lexicalTarget("outer", "secondary")
		if targetKey(left) == targetKey(right) {
			t.Fatal("nested projection absent from identity")
		}
		right.Root = "app:B"
		if targetKey(left) == targetKey(right) {
			t.Fatal("instance absent from target identity")
		}
	})
	t.Run("LEX-010", func(t *testing.T) {
		body := Node{Kind: Sequence, Children: []Node{{Kind: Acquire, ID: "occurrence:1", Meaning: "constructor", Resource: ResourceRef{ID: "resource:1"}}, {Kind: Acquire, ID: "occurrence:2", Meaning: "constructor", Resource: ResourceRef{ID: "resource:2"}}}}
		summary := reviewInfer(t, Plan{Entry: body}, AnalysisOptions{})
		if len(summary.Paths) != 1 || len(summary.Paths[0].Owned) != 2 {
			t.Fatalf("same meaning deduplicated construction: %+v", summary)
		}
	})
	t.Run("LEX-033", func(t *testing.T) {
		candidate := lexicalCandidate("domain::Real Time", lexicalType("Theme"))
		candidate.Phrase = "Real Time"
		result := lexicalSelection(candidate)
		if result.Status != Established || result.Candidates[0].Candidate.Phrase != "Real Time" {
			t.Fatalf("multiword semantic entry was split: %+v", result)
		}
	})
	t.Run("LEX-034", func(t *testing.T) {
		entry := lexicalCandidate("idiom", Predicate{Kind: "true"})
		entry.Phrase = "Real Time"
		composition := lexicalCandidate("composition", Predicate{Kind: "true"})
		composition.Phrase = "Real Time"
		if result := lexicalSelection(entry, composition); result.Status != Contradicted {
			t.Fatalf("phrase ambiguity silently resolved: %+v", result)
		}
	})
	t.Run("LEX-038", func(t *testing.T) {
		for _, spelling := range []string{"A", "LongCompoundIdentifier", "Multi Word"} {
			candidate := lexicalCandidate("meaning", Predicate{Kind: "true"})
			candidate.Phrase = spelling
			if result := lexicalSelection(candidate); result.Status != Established || result.Meaning != "meaning" {
				t.Fatalf("spelling changed semantics: %+v", result)
			}
		}
	})
}

func paletteSchema(cardinality string) FacetSchema {
	return FacetSchema{ID: "palette", Version: "1", Domain: "enum", Values: []Atom{{Kind: "enum", Text: "dark"}, {Kind: "enum", Text: "light"}}, Cardinality: cardinality, AllowedContributions: []string{"select", "require", "assert", "replace"}, ClosurePoint: "constructed"}
}
func paletteContribution(target TargetRef, value string) ConfigurationContribution {
	return ConfigurationContribution{Facet: "palette", Target: target, Kind: "select", Value: Atom{Kind: "enum", Text: value}, Occurrence: BindingOccurrenceID(value), Origin: Origin{Source: value}}
}
func requireConfiguration(t *testing.T, schemas []FacetSchema, slots []FacetSlot, contributions []ConfigurationContribution, rules []FacetRule) ConfigurationResult {
	t.Helper()
	result, err := CheckConfiguration(schemas, slots, contributions, rules, nil)
	if err != nil {
		t.Fatal(err)
	}
	return result
}
func hasStatus(obligations []Obligation, status Status) bool {
	for _, obligation := range obligations {
		if obligation.Status == status {
			return true
		}
	}
	return false
}

func TestFacetAcceptance(t *testing.T) {
	schema := paletteSchema("optional-single")
	target := lexicalTarget("primary")
	dark, light := paletteContribution(target, "dark"), paletteContribution(target, "light")
	t.Run("LEX-011", func(t *testing.T) {
		result := requireConfiguration(t, []FacetSchema{schema}, nil, []ConfigurationContribution{dark, light}, nil)
		if !hasStatus(result.Obligations, Contradicted) || len(result.Selections[0].Origins) != 2 {
			t.Fatalf("exact conflict missing origins: %+v", result)
		}
	})
	t.Run("LEX-012", func(t *testing.T) {
		other := light
		other.Target = lexicalTarget("secondary")
		result := requireConfiguration(t, []FacetSchema{schema}, nil, []ConfigurationContribution{dark, other}, nil)
		if !AllEstablished(result.Obligations) || len(result.Selections) != 2 {
			t.Fatalf("different role selections conflicted: %+v", result)
		}
	})
	t.Run("LEX-013", func(t *testing.T) {
		other := light
		other.Target = lexicalTarget("secondary")
		rule := FacetRule{When: []Predicate{{Kind: "true"}}, Requires: []Predicate{{Kind: "resource_disjoint", Subject: "dynamic:A/B"}}}
		result := requireConfiguration(t, []FacetSchema{schema}, nil, []ConfigurationContribution{dark, other}, []FacetRule{rule})
		if !hasStatus(result.Obligations, Unresolved) {
			t.Fatal("different target paths proved resource disjointness")
		}
	})
	t.Run("LEX-014", func(t *testing.T) {
		during, after := dark, light
		during.Scope = "crash-during-session"
		after.Scope = "normal-session-end"
		result := requireConfiguration(t, []FacetSchema{schema}, nil, []ConfigurationContribution{during, after}, nil)
		if !AllEstablished(result.Obligations) || len(result.Selections) != 2 {
			t.Fatal("distinct lifecycle scopes conflated")
		}
	})
	t.Run("LEX-015", func(t *testing.T) {
		request := dark
		request.Kind = "require"
		result := requireConfiguration(t, []FacetSchema{schema}, nil, []ConfigurationContribution{request}, nil)
		if len(result.Requests) != 1 || !hasStatus(result.Obligations, Unresolved) || len(result.Selections) != 0 {
			t.Fatalf("request manufactured guarantee: %+v", result)
		}
	})
	t.Run("LEX-016", func(t *testing.T) {
		result := requireConfiguration(t, []FacetSchema{schema}, nil, []ConfigurationContribution{dark}, nil)
		selected := FacetFact(FacetSlot{Facet: "palette", Target: target}, dark.Value)
		if status, _ := Entails(result.Facts, selected); status != Established {
			t.Fatal("selection fact missing")
		}
		if status, _ := Entails(result.Facts, Predicate{Kind: "effects_closed", Subject: "NoBlocking"}); status != Unresolved {
			t.Fatal("selection label certified behavior")
		}
	})
	t.Run("LEX-017", func(t *testing.T) {
		required := paletteSchema("exactly-one-at-close")
		slot := FacetSlot{Facet: "palette", Target: target}
		open := requireConfiguration(t, []FacetSchema{required}, []FacetSlot{slot}, nil, nil)
		slot.Closed = true
		closed := requireConfiguration(t, []FacetSchema{required}, []FacetSlot{slot}, nil, nil)
		if !AllEstablished(open.Obligations) || !hasStatus(closed.Obligations, Unresolved) {
			t.Fatal("closure cardinality incorrectly applied")
		}
	})
	t.Run("LEX-018", func(t *testing.T) {
		second := dark
		second.Origin = Origin{Source: "second"}
		second.Occurrence = "second"
		result := requireConfiguration(t, []FacetSchema{paletteSchema("set")}, nil, []ConfigurationContribution{dark, second, light}, nil)
		if len(result.Selections[0].Values) != 2 || len(result.Selections[0].Origins) != 3 || len(result.Contributions) != 3 {
			t.Fatal("logical set normalization erased occurrence work")
		}
	})
	t.Run("LEX-019", func(t *testing.T) {
		replacement := light
		replacement.Kind = "replace"
		result := requireConfiguration(t, []FacetSchema{schema}, nil, []ConfigurationContribution{dark, replacement}, nil)
		if hasStatus(result.Obligations, Contradicted) || len(result.Replacements) != 1 || len(result.Selections[0].Values) != 1 || result.Selections[0].Values[0] != light.Value || len(result.Replacements[0].Previous.Values) != 1 || result.Replacements[0].ReleaseObligation.Status != Unresolved {
			t.Fatalf("ordered replacement lost history or release obligation: %+v", result)
		}
	})
	t.Run("LEX-020", func(t *testing.T) {
		for _, contributions := range [][]ConfigurationContribution{{dark, light}, {light, dark}} {
			result := requireConfiguration(t, []FacetSchema{schema}, nil, contributions, nil)
			if !hasStatus(result.Obligations, Contradicted) {
				t.Fatal("implicit last-word-wins")
			}
		}
	})
	t.Run("LEX-021", func(t *testing.T) {
		result := requireConfiguration(t, []FacetSchema{schema}, nil, []ConfigurationContribution{dark}, nil)
		if !AllEstablished(result.Obligations) {
			t.Fatal("fixture configuration inconsistent")
		}
		summary := reviewInfer(t, Plan{Entry: Node{Kind: Event, Effects: []Effect{{Kind: "io", Definite: true}}}}, AnalysisOptions{})
		if AllEstablished(CheckProfile(summary, Profile{Name: "NoExternalIO"})) {
			t.Fatal("facet consistency certified external IO")
		}
	})
	t.Run("LEX-022", func(t *testing.T) {
		a := FacetFact(FacetSlot{Facet: "palette", Target: target}, dark.Value)
		feature := FacetSchema{ID: "feature", Version: "1", Domain: "bool", Cardinality: "optional-single", AllowedContributions: []string{"select"}, ClosurePoint: "constructed"}
		b := FacetFact(FacetSlot{Facet: "feature", Target: target, Scope: "enabled"}, Atom{Kind: "bool", Bool: true})
		rules := []FacetRule{{When: []Predicate{a}, Consequences: []Predicate{b}, Requires: []Predicate{b}}, {When: []Predicate{b}, Consequences: []Predicate{a}, Requires: []Predicate{a}}}
		result := requireConfiguration(t, []FacetSchema{schema, feature}, nil, []ConfigurationContribution{dark}, rules)
		if !AllEstablished(result.Obligations) || len(result.Facts) != 2 {
			t.Fatalf("finite rule cycle failed to converge: %+v", result)
		}
	})
	t.Run("LEX-023", func(t *testing.T) {
		rules := []FacetRule{{When: []Predicate{{Kind: "true"}}, Requires: []Predicate{{Kind: "arbitrary_global_search"}}}}
		result := requireConfiguration(t, []FacetSchema{schema}, nil, []ConfigurationContribution{dark}, rules)
		if !hasStatus(result.Obligations, Unresolved) || len(result.Selections) != 1 {
			t.Fatal("unsupported search manufactured configuration")
		}
	})
}

func TestFrameAcceptance(t *testing.T) {
	palette := lexicalTarget("theme", "palette")
	backend := lexicalTarget("backend")
	frame := FrameContract{Target: lexicalTarget(), Allowed: []TargetRef{palette}, Preserves: []TargetRef{backend}, Scope: "elaboration"}
	t.Run("LEX-024", func(t *testing.T) {
		if !AllEstablished(CheckFrame(frame, []FrameFootprint{{Kind: "write", Target: palette, Definite: true}}, FrameContext{})) {
			t.Fatal("exact disjoint structural preservation failed")
		}
	})
	t.Run("LEX-025", func(t *testing.T) {
		if !hasStatus(CheckFrame(frame, []FrameFootprint{{Kind: "write", Target: backend, Definite: true, Origin: Origin{Source: "wrapper"}}}, FrameContext{}), Contradicted) {
			t.Fatal("unexpected actual write trusted frame annotation")
		}
	})
	t.Run("LEX-026", func(t *testing.T) {
		local := FrameContract{Target: lexicalTarget(), Allowed: []TargetRef{lexicalTarget("theme")}, Preserves: []TargetRef{palette}}
		if !hasStatus(CheckFrame(local, []FrameFootprint{{Kind: "replace", Target: lexicalTarget("theme"), Definite: true}}, FrameContext{}), Contradicted) {
			t.Fatal("parent replacement preserved descendant")
		}
	})
	t.Run("LEX-027", func(t *testing.T) {
		if !hasStatus(CheckFrame(frame, []FrameFootprint{{Kind: "unknown"}}, FrameContext{}), Unresolved) {
			t.Fatal("unknown write proved preservation")
		}
		runtime := frame
		runtime.Scope = "runtime"
		runtime.Resource = ResourceRef{ID: "handle:A"}
		writes := []FrameFootprint{{Kind: "write", Target: palette, Resource: ResourceRef{ID: "handle:B"}, Definite: true}}
		if !hasStatus(CheckFrame(runtime, writes, FrameContext{}), Unresolved) {
			t.Fatal("path difference implied native resource disjointness")
		}
	})
	t.Run("FALS-007", func(t *testing.T) {
		summary := reviewInfer(t, Plan{Entry: Node{Kind: Event, Effects: []Effect{{Kind: "write", Target: backend, Definite: true}}}}, AnalysisOptions{})
		if !hasStatus(CheckFrame(frame, FootprintsFromEffects(summary.Effects), FrameContext{}), Contradicted) {
			t.Fatal("favorable frame contribution hid actual body write")
		}
	})
}

func TestProviderAcceptance(t *testing.T) {
	guarantee := Predicate{Kind: "eq", Subject: "result", Value: Atom{Kind: "enum", Text: "unchanged"}}
	role := RoleInterface{ID: "store", Version: "1", Operations: []RoleOperation{{Name: "update", Ensures: []Predicate{guarantee}, Profiles: []Profile{{Name: "NoBlocking"}}, Outcomes: []Outcome{Normal}}}}
	provider := ProviderImplementation{ID: "provider", Target: lexicalTarget("store"), Operations: map[string]Plan{"update": {Entry: Node{Kind: Event, Facts: []Predicate{guarantee}}}}}
	good, err := CheckProvider(role, provider, nil, AnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !AllEstablished(good.Obligations) {
		t.Fatalf("supported provider did not conform: %+v", good)
	}
	provider.Operations["update"] = Plan{Entry: Node{Kind: Event, Effects: []Effect{{Kind: "block", Definite: true}}, Facts: []Predicate{{Kind: "eq", Subject: "result", Value: Atom{Kind: "enum", Text: "changed"}}}}}
	bad, err := CheckProvider(role, provider, nil, AnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !hasStatus(bad.Obligations, Contradicted) {
		t.Fatal("operation names proved provider conformance")
	}
}
