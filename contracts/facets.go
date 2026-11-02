package contracts

import (
	"fmt"
	"reflect"
	"sort"
)

type FacetSchema struct {
	ID                   string    `json:"id"`
	Version              string    `json:"version"`
	Applicability        Predicate `json:"applicability"`
	Domain               string    `json:"domain"`
	Values               []Atom    `json:"values,omitempty"`
	Lower                *int64    `json:"lower,omitempty"`
	Upper                *int64    `json:"upper,omitempty"`
	Cardinality          string    `json:"cardinality"`
	AllowedContributions []string  `json:"allowed_contributions"`
	ClosurePoint         string    `json:"closure_point"`
	Origin               Origin    `json:"origin"`
}

type PropertyRequest struct {
	Predicate         Predicate `json:"predicate"`
	Target            TargetRef `json:"target"`
	Scope             string    `json:"scope"`
	RequiredAssurance string    `json:"required_assurance"`
	Origin            Origin    `json:"origin"`
}

type ConfigurationContribution struct {
	Occurrence BindingOccurrenceID `json:"occurrence"`
	Facet      string              `json:"facet"`
	Target     TargetRef           `json:"target"`
	Scope      string              `json:"scope"`
	Kind       string              `json:"kind"`
	Value      Atom                `json:"value"`
	Condition  []Predicate         `json:"condition,omitempty"`
	Origin     Origin              `json:"origin"`
}

type FacetRule struct {
	When         []Predicate `json:"when"`
	Requires     []Predicate `json:"requires"`
	Consequences []Predicate `json:"consequences,omitempty"`
	Origin       Origin      `json:"origin"`
}

type FacetSlot struct {
	Facet  string    `json:"facet"`
	Target TargetRef `json:"target"`
	Scope  string    `json:"scope"`
	Closed bool      `json:"closed"`
}

type FacetSelection struct {
	Slot    FacetSlot `json:"slot"`
	Values  []Atom    `json:"values"`
	Origins []Origin  `json:"origins"`
}

type FacetReplacement struct {
	Previous          FacetSelection            `json:"previous"`
	Contribution      ConfigurationContribution `json:"contribution"`
	ReleaseObligation Obligation                `json:"release_obligation"`
}

type ConfigurationResult struct {
	Selections    []FacetSelection            `json:"selections"`
	Requests      []PropertyRequest           `json:"requests"`
	Contributions []ConfigurationContribution `json:"contributions"`
	Conditional   []ConfigurationContribution `json:"conditional,omitempty"`
	Facts         []Predicate                 `json:"facts"`
	Obligations   []Obligation                `json:"obligations"`
	Replacements  []FacetReplacement          `json:"replacements,omitempty"`
}

func containsString(values []string, value string) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}

func (schema FacetSchema) Validate() error {
	if schema.ID == "" || schema.Version == "" || schema.ClosurePoint == "" {
		return fmt.Errorf("facet requires identity, version and closure point")
	}
	if !containsString([]string{"optional-single", "exactly-one-at-close", "set"}, schema.Cardinality) {
		return fmt.Errorf("facet %s has unsupported cardinality %q", schema.ID, schema.Cardinality)
	}
	if !containsString([]string{"enum", "bool", "reference", "quantity"}, schema.Domain) {
		return fmt.Errorf("facet %s has unsupported domain %q", schema.ID, schema.Domain)
	}
	if schema.Domain == "enum" && len(schema.Values) == 0 {
		return fmt.Errorf("facet %s has an empty enum domain", schema.ID)
	}
	if schema.Lower != nil && schema.Upper != nil && *schema.Lower > *schema.Upper {
		return fmt.Errorf("facet %s has an inverted quantity interval", schema.ID)
	}
	for _, kind := range schema.AllowedContributions {
		if !containsString([]string{"require", "select", "replace", "assert"}, kind) {
			return fmt.Errorf("facet %s has unsupported contribution %q", schema.ID, kind)
		}
	}
	return nil
}

func (schema FacetSchema) accepts(value Atom) bool {
	switch schema.Domain {
	case "enum":
		for _, allowed := range schema.Values {
			if allowed == value {
				return true
			}
		}
	case "bool":
		return value.Kind == "bool"
	case "reference":
		return value.Kind == "reference" && value.Text != ""
	case "quantity":
		return value.Kind == "number" && (schema.Lower == nil || value.Number >= *schema.Lower) && (schema.Upper == nil || value.Number <= *schema.Upper)
	}
	return false
}

func facetSlotKey(slot FacetSlot) string {
	return StableID("facet", struct{ Target, Facet, Scope string }{targetKey(slot.Target), slot.Facet, slot.Scope})
}

// FacetFact denotes a selected configuration value only. It says nothing about
// implementation behavior or English associations of its label.
func FacetFact(slot FacetSlot, value Atom) Predicate {
	return Predicate{Kind: "present", Subject: "facet:" + slot.Facet + ":" + slot.Scope, Value: value, Target: slot.Target}
}

func facetObligation(target TargetRef, origin Origin, status Status, reason string, predicate Predicate) Obligation {
	claim := Claim{ID: StableID("claim", struct {
		Target    TargetRef
		Predicate Predicate
		Origin    Origin
	}{target, predicate, origin}), Predicate: predicate, Target: target, Scope: "configuration", Phase: target.Phase}
	return Obligation{ID: StableID("obligation", claim), Claim: claim, Status: status, Reason: reason, Origin: origin}
}

// CheckConfiguration preserves the contribution sequence. Only pure selected
// facts are deduplicated; the occurrence ledger is never normalized away.
func CheckConfiguration(schemas []FacetSchema, slots []FacetSlot, contributions []ConfigurationContribution, rules []FacetRule, facts []Predicate) (ConfigurationResult, error) {
	result := ConfigurationResult{Contributions: append([]ConfigurationContribution(nil), contributions...), Facts: append([]Predicate(nil), facts...)}
	definitions := make(map[string]FacetSchema)
	for _, schema := range schemas {
		if err := schema.Validate(); err != nil {
			return result, err
		}
		if _, duplicate := definitions[schema.ID]; duplicate {
			return result, fmt.Errorf("duplicate facet identity %s", schema.ID)
		}
		definitions[schema.ID] = schema
	}
	selection := make(map[string]*FacetSelection)
	for _, slot := range slots {
		if _, exists := definitions[slot.Facet]; !exists {
			return result, fmt.Errorf("unknown facet %s", slot.Facet)
		}
		key := facetSlotKey(slot)
		if prior := selection[key]; prior != nil {
			prior.Slot.Closed = prior.Slot.Closed || slot.Closed
		} else {
			selection[key] = &FacetSelection{Slot: slot}
		}
	}
	for _, contribution := range contributions {
		schema, exists := definitions[contribution.Facet]
		if !exists {
			return result, fmt.Errorf("unknown facet %s", contribution.Facet)
		}
		if !containsString(schema.AllowedContributions, contribution.Kind) {
			return result, fmt.Errorf("facet %s does not authorize %s", schema.ID, contribution.Kind)
		}
		if !schema.accepts(contribution.Value) {
			return result, fmt.Errorf("facet %s rejects value outside its domain", schema.ID)
		}
		slot := FacetSlot{Facet: contribution.Facet, Target: contribution.Target, Scope: contribution.Scope}
		predicate := FacetFact(slot, contribution.Value)
		if contribution.Kind == "require" || contribution.Kind == "assert" {
			result.Requests = append(result.Requests, PropertyRequest{Predicate: predicate, Target: contribution.Target, Scope: contribution.Scope, RequiredAssurance: "established", Origin: contribution.Origin})
			continue
		}
		if contribution.Target.Unknown || contribution.Target.Root == "" {
			result.Obligations = append(result.Obligations, facetObligation(contribution.Target, contribution.Origin, Unresolved, "facet attachment identity is unresolved", predicate))
			continue
		}
		applicability := schema.Applicability
		if applicability.Kind == "" {
			applicability = Predicate{Kind: "true"}
		}
		localFacts := append([]Predicate(nil), result.Facts...)
		if contribution.Target.Type != "" {
			localFacts = append(localFacts, Predicate{Kind: "type", Subject: "target", Value: Atom{Kind: "type", Text: contribution.Target.Type}})
		}
		status, reason := Entails(localFacts, applicability)
		if status != Established {
			result.Obligations = append(result.Obligations, facetObligation(contribution.Target, contribution.Origin, status, "facet applicability: "+reason, applicability))
			continue
		}
		conditionStatus, _ := Entails(localFacts, Predicate{Kind: "and", Args: contribution.Condition})
		if conditionStatus == Contradicted {
			continue
		}
		if conditionStatus != Established {
			result.Conditional = append(result.Conditional, contribution)
			continue
		}
		key := facetSlotKey(slot)
		current := selection[key]
		if current == nil {
			current = &FacetSelection{Slot: slot}
			selection[key] = current
		}
		if contribution.Kind == "replace" {
			previous := *current
			previous.Values, previous.Origins = append([]Atom(nil), current.Values...), append([]Origin(nil), current.Origins...)
			obligation := facetObligation(contribution.Target, contribution.Origin, Unresolved, "replacement requires concrete ownership, release and implementation rechecking", Predicate{Kind: "replacement", Subject: schema.ID})
			result.Replacements = append(result.Replacements, FacetReplacement{previous, contribution, obligation})
			result.Obligations = append(result.Obligations, obligation)
			current.Values, current.Origins = nil, nil
		}
		duplicate := false
		for _, value := range current.Values {
			if value == contribution.Value {
				duplicate = true
			}
		}
		if !duplicate && len(current.Values) > 0 && schema.Cardinality != "set" {
			conflict := facetObligation(contribution.Target, contribution.Origin, Contradicted, "exclusive facet has simultaneous distinct selections on the same exact target and scope", predicate)
			for _, origin := range current.Origins {
				conflict.Dependencies = append(conflict.Dependencies, StableID("origin", origin))
			}
			result.Obligations = append(result.Obligations, conflict)
		}
		if !duplicate {
			current.Values = append(current.Values, contribution.Value)
		}
		current.Origins = append(current.Origins, contribution.Origin)
	}
	keys := make([]string, 0, len(selection))
	for key := range selection {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		current := selection[key]
		schema := definitions[current.Slot.Facet]
		if current.Slot.Closed && schema.Cardinality == "exactly-one-at-close" && len(current.Values) == 0 {
			result.Obligations = append(result.Obligations, facetObligation(current.Slot.Target, schema.Origin, Unresolved, "exactly-one facet is unset at configuration closure", Predicate{Kind: "present", Subject: schema.ID}))
		}
		for _, value := range current.Values {
			result.Facts = append(result.Facts, FacetFact(current.Slot, value))
		}
		result.Selections = append(result.Selections, *current)
	}
	// Consequences are restricted to configuration facts. Behavioral guarantees
	// always require the ordinary evidence/implementation checker.
	seen := make(map[string]bool)
	for _, fact := range result.Facts {
		seen[StableID("fact", fact)] = true
	}
	for changed := true; changed; {
		changed = false
		for _, rule := range rules {
			status, _ := Entails(result.Facts, Predicate{Kind: "and", Args: rule.When})
			if status != Established {
				continue
			}
			for _, consequence := range rule.Consequences {
				if consequence.Kind != "present" || len(consequence.Subject) < 6 || consequence.Subject[:6] != "facet:" {
					return result, fmt.Errorf("cross-facet rules may derive configuration facts only")
				}
				key := StableID("fact", consequence)
				if !seen[key] {
					seen[key] = true
					result.Facts = append(result.Facts, consequence)
					changed = true
				}
			}
		}
	}
	for _, rule := range rules {
		condition, _ := Entails(result.Facts, Predicate{Kind: "and", Args: rule.When})
		if condition == Contradicted {
			continue
		}
		for _, requirement := range rule.Requires {
			status, reason := Entails(result.Facts, requirement)
			if condition != Established {
				status, reason = Unresolved, "cross-facet rule applicability is unresolved"
			}
			result.Obligations = append(result.Obligations, facetObligation(requirement.Target, rule.Origin, status, reason, requirement))
		}
	}
	for _, request := range result.Requests {
		status, reason := Entails(result.Facts, request.Predicate)
		result.Obligations = append(result.Obligations, facetObligation(request.Target, request.Origin, status, reason, request.Predicate))
	}
	return result, nil
}

// SameConfiguration compares pure final facts only; it is not an execution
// equivalence check. Contributions and replacement work remain ordered.
func SameConfiguration(left, right ConfigurationResult) bool {
	a, b := append([]Predicate(nil), left.Facts...), append([]Predicate(nil), right.Facts...)
	sort.Slice(a, func(i, j int) bool { return StableID("fact", a[i]) < StableID("fact", a[j]) })
	sort.Slice(b, func(i, j int) bool { return StableID("fact", b[i]) < StableID("fact", b[j]) })
	return reflect.DeepEqual(a, b)
}
