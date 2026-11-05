package contracts

import (
	"fmt"
	"sort"
	"strings"
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

// configurationChecker keeps all selected values in one validated ledger. Facts
// are a projection of that ledger, never an independent rule-output store.
type configurationChecker struct {
	result      ConfigurationResult
	definitions map[string]FacetSchema
	selection   map[string]*FacetSelection
	premises    []Predicate
}

func (checker *configurationChecker) validate(contribution ConfigurationContribution) error {
	schema, exists := checker.definitions[contribution.Facet]
	if !exists {
		return fmt.Errorf("unknown facet %s", contribution.Facet)
	}
	if !containsString(schema.AllowedContributions, contribution.Kind) {
		return fmt.Errorf("facet %s does not authorize %s", schema.ID, contribution.Kind)
	}
	if !schema.accepts(contribution.Value) {
		return fmt.Errorf("facet %s rejects value outside its domain", schema.ID)
	}
	_, _, err := checker.resolveSubject(FacetFact(FacetSlot{Facet: contribution.Facet, Scope: contribution.Scope}, contribution.Value).Subject)
	return err
}

// Legacy subjects remain accepted only when the declared schema identities
// determine one facet/scope split. Both fields may themselves contain colons.
func (checker *configurationChecker) normalize(fact Predicate, origin Origin) (ConfigurationContribution, error) {
	value := ConfigurationContribution{Target: fact.Target, Kind: "select", Value: fact.Value, Origin: origin}
	if fact.Kind != "present" || !strings.HasPrefix(fact.Subject, "facet:") || len(fact.Args) != 0 || fact.Resource != (ResourceRef{}) || fact.Lower != nil || fact.Upper != nil {
		return value, fmt.Errorf("cross-facet selections require an exact FacetFact predicate")
	}
	identity, scope, err := checker.resolveSubject(fact.Subject)
	if err != nil {
		return value, err
	}
	value.Facet, value.Scope = identity, scope
	return value, checker.validate(value)
}

func (checker *configurationChecker) resolveSubject(subject string) (string, string, error) {
	var matches []string
	for identity := range checker.definitions {
		if strings.HasPrefix(subject, "facet:"+identity+":") {
			matches = append(matches, identity)
		}
	}
	sort.Strings(matches)
	if len(matches) == 0 {
		return "", "", fmt.Errorf("unknown facet in configuration predicate %q", subject)
	}
	if len(matches) != 1 {
		return "", "", fmt.Errorf("ambiguous legacy facet/scope predicate %q matches %v; use unambiguous facet identities and scopes", subject, matches)
	}
	return matches[0], strings.TrimPrefix(subject, "facet:"+matches[0]+":"), nil
}

func (checker *configurationChecker) facts() []Predicate {
	facts := append([]Predicate(nil), checker.premises...)
	for _, key := range sortedMapKeys(checker.selection) {
		current := checker.selection[key]
		for _, value := range current.Values {
			facts = append(facts, FacetFact(current.Slot, value))
		}
	}
	return facts
}

func (checker *configurationChecker) applicability(value ConfigurationContribution) (Status, string) {
	if value.Target.Unknown || value.Target.Root == "" {
		return Unresolved, "facet attachment identity is unresolved"
	}
	facts, status, reason := authoritativeTargetFacts(value.Target, checker.facts())
	if status != Established {
		return status, reason
	}
	applicability := checker.definitions[value.Facet].Applicability
	if applicability.Kind == "" {
		applicability = Predicate{Kind: "true"}
	}
	status, reason = Entails(facts, applicability)
	return status, "facet applicability: " + reason
}

func (checker *configurationChecker) obligation(value ConfigurationContribution, status Status, reason string) {
	slot := FacetSlot{Facet: value.Facet, Target: value.Target, Scope: value.Scope}
	checker.result.Obligations = append(checker.result.Obligations, facetObligation(value.Target, value.Origin, status, reason, FacetFact(slot, value.Value)))
}

// selectValue is called only after static validation and applicability. It
// preserves every origin while deduplicating pure selected values only.
func (checker *configurationChecker) selectValue(value ConfigurationContribution) bool {
	slot := FacetSlot{Facet: value.Facet, Target: value.Target, Scope: value.Scope}
	key := facetSlotKey(slot)
	current := checker.selection[key]
	if current == nil {
		current = &FacetSelection{Slot: slot}
		checker.selection[key] = current
	}
	if value.Kind == "replace" {
		previous := *current
		previous.Values, previous.Origins = append([]Atom(nil), current.Values...), append([]Origin(nil), current.Origins...)
		obligation := facetObligation(value.Target, value.Origin, Unresolved, "replacement requires concrete ownership, release and implementation rechecking", Predicate{Kind: "replacement", Subject: value.Facet})
		checker.result.Replacements = append(checker.result.Replacements, FacetReplacement{previous, value, obligation})
		checker.result.Obligations = append(checker.result.Obligations, obligation)
		current.Values, current.Origins = nil, nil
		var retained []Predicate
		for _, premise := range checker.premises {
			if !checker.premiseDependsOn(premise, slot) {
				retained = append(retained, premise)
			}
		}
		checker.premises = retained
	}
	predicate := FacetFact(slot, value.Value)
	if status, _ := Entails(checker.premises, predicate); status == Contradicted {
		checker.obligation(value, Contradicted, "selected facet value contradicts an explicit configuration premise")
	}
	duplicate := false
	for _, selected := range current.Values {
		duplicate = duplicate || selected == value.Value
	}
	if !duplicate && len(current.Values) > 0 && checker.definitions[value.Facet].Cardinality != "set" {
		conflict := facetObligation(value.Target, value.Origin, Contradicted, "exclusive facet has simultaneous distinct selections on the same exact target and scope", predicate)
		for _, origin := range current.Origins {
			conflict.Dependencies = append(conflict.Dependencies, StableID("origin", origin))
		}
		checker.result.Obligations = append(checker.result.Obligations, conflict)
	}
	if !duplicate {
		current.Values = append(current.Values, value.Value)
	}
	current.Origins = append(current.Origins, value.Origin)
	return !duplicate
}

func (checker *configurationChecker) premiseDependsOn(fact Predicate, slot FacetSlot) bool {
	if strings.HasPrefix(fact.Subject, "facet:") {
		facet, scope, err := checker.resolveSubject(fact.Subject)
		if err != nil || facet == slot.Facet && scope == slot.Scope && TargetRelation(fact.Target, slot.Target) != Contradicted {
			return true
		}
	}
	for _, child := range fact.Args {
		if checker.premiseDependsOn(child, slot) {
			return true
		}
	}
	return false
}

func (checker *configurationChecker) contribute(value ConfigurationContribution) {
	facts := checker.facts()
	condition, _ := Entails(facts, Predicate{Kind: "and", Args: value.Condition})
	if condition == Contradicted {
		return
	}
	if condition != Established {
		checker.result.Conditional = append(checker.result.Conditional, value)
		checker.obligation(value, Unresolved, "ordered facet contribution condition is unresolved; contribution was not applied")
		return
	}
	status, reason := checker.applicability(value)
	if status != Established {
		checker.obligation(value, status, reason)
		return
	}
	if value.Kind == "require" || value.Kind == "assert" {
		slot := FacetSlot{Facet: value.Facet, Target: value.Target, Scope: value.Scope}
		checker.result.Requests = append(checker.result.Requests, PropertyRequest{Predicate: FacetFact(slot, value.Value), Target: value.Target, Scope: value.Scope, RequiredAssurance: "established", Origin: value.Origin})
		return
	}
	checker.selectValue(value)
}

// A conjunction of input premises exposes its selections to the same ledger as
// direct and derived contributions. Other logical wrappers retain their shape,
// but every nested facet reference must still be a valid configuration fact.
func (checker *configurationChecker) seed(facts []Predicate) error {
	var selections []ConfigurationContribution
	var collect func(Predicate) error
	collect = func(fact Predicate) error {
		if fact.Kind == "and" {
			for _, child := range fact.Args {
				if err := collect(child); err != nil {
					return err
				}
			}
			return nil
		}
		if strings.HasPrefix(fact.Subject, "facet:") {
			selection, err := checker.normalize(fact, Origin{})
			if err != nil {
				return err
			}
			selections = append(selections, selection)
			return nil
		}
		checker.premises = append(checker.premises, fact)
		return nil
	}
	for _, fact := range facts {
		if err := collect(fact); err != nil {
			return err
		}
	}
	// Do not expose unchecked nested facts while checking applicability.
	premises := checker.premises
	checker.premises = nil
	for _, fact := range premises {
		if !hasFacetPredicate(fact) {
			checker.premises = append(checker.premises, fact)
		}
	}
	for _, fact := range premises {
		if !hasFacetPredicate(fact) {
			continue
		}
		valid, err := checker.validateNestedPremise(fact)
		if err != nil {
			return err
		}
		if valid {
			checker.premises = append(checker.premises, fact)
		}
	}
	for _, value := range selections {
		checker.contribute(value)
	}
	return nil
}

func hasFacetPredicate(fact Predicate) bool {
	if strings.HasPrefix(fact.Subject, "facet:") {
		return true
	}
	for _, child := range fact.Args {
		if hasFacetPredicate(child) {
			return true
		}
	}
	return false
}

func (checker *configurationChecker) validateNestedPremise(fact Predicate) (bool, error) {
	valid := true
	if strings.HasPrefix(fact.Subject, "facet:") {
		value, err := checker.normalize(fact, Origin{})
		if err != nil {
			return false, err
		}
		status, reason := checker.applicability(value)
		if status != Established {
			checker.obligation(value, status, reason)
			valid = false
		}
	}
	for _, child := range fact.Args {
		ok, err := checker.validateNestedPremise(child)
		if err != nil {
			return false, err
		}
		valid = valid && ok
	}
	return valid, nil
}

func (checker *configurationChecker) rules(rules []FacetRule, consequences [][]ConfigurationContribution) {
	applied := make([][]bool, len(rules))
	for i := range rules {
		applied[i] = make([]bool, len(consequences[i]))
	}
	for changed := true; changed; {
		changed = false
		for i, rule := range rules {
			condition, _ := Entails(checker.facts(), Predicate{Kind: "and", Args: rule.When})
			if condition != Established {
				continue
			}
			for j, value := range consequences[i] {
				if applied[i][j] {
					continue
				}
				if status, _ := checker.applicability(value); status != Established {
					continue
				}
				applied[i][j] = true
				changed = checker.selectValue(value) || changed
			}
		}
	}
	for i, rule := range rules {
		condition, _ := Entails(checker.facts(), Predicate{Kind: "and", Args: rule.When})
		if condition == Contradicted {
			continue
		}
		for j, value := range consequences[i] {
			if applied[i][j] {
				continue
			}
			status, reason := checker.applicability(value)
			if condition != Established {
				status, reason = Unresolved, "cross-facet rule applicability is unresolved; consequence was not selected"
			}
			checker.obligation(value, status, reason)
		}
		for _, requirement := range rule.Requires {
			status, reason := Entails(checker.facts(), requirement)
			if condition != Established {
				status, reason = Unresolved, "cross-facet rule applicability is unresolved"
			}
			checker.result.Obligations = append(checker.result.Obligations, facetObligation(requirement.Target, rule.Origin, status, reason, requirement))
		}
	}
}

// CheckConfiguration preserves the direct contribution sequence. Rule-derived
// selections use the same schema/cardinality state, and closure is checked only
// after their finite fixed point. No rule performs an ordered replacement.
func CheckConfiguration(schemas []FacetSchema, slots []FacetSlot, contributions []ConfigurationContribution, rules []FacetRule, facts []Predicate) (ConfigurationResult, error) {
	checker := configurationChecker{result: ConfigurationResult{Contributions: append([]ConfigurationContribution(nil), contributions...)}, definitions: make(map[string]FacetSchema), selection: make(map[string]*FacetSelection)}
	for _, schema := range schemas {
		if err := schema.Validate(); err != nil {
			return checker.result, err
		}
		if _, duplicate := checker.definitions[schema.ID]; duplicate {
			return checker.result, fmt.Errorf("duplicate facet identity %s", schema.ID)
		}
		checker.definitions[schema.ID] = schema
	}
	for _, slot := range slots {
		if _, exists := checker.definitions[slot.Facet]; !exists {
			return checker.result, fmt.Errorf("unknown facet %s", slot.Facet)
		}
		key := facetSlotKey(slot)
		if prior := checker.selection[key]; prior != nil {
			prior.Slot.Closed = prior.Slot.Closed || slot.Closed
		} else {
			checker.selection[key] = &FacetSelection{Slot: slot}
		}
	}
	for _, contribution := range contributions {
		if err := checker.validate(contribution); err != nil {
			return checker.result, err
		}
	}
	consequences := make([][]ConfigurationContribution, len(rules))
	for i, rule := range rules {
		for _, fact := range rule.Consequences {
			value, err := checker.normalize(fact, rule.Origin)
			if err != nil {
				return checker.result, err
			}
			consequences[i] = append(consequences[i], value)
		}
	}
	if err := checker.seed(facts); err != nil {
		return checker.result, err
	}
	for _, contribution := range contributions {
		checker.contribute(contribution)
	}
	checker.rules(rules, consequences)
	checker.result.Facts = checker.facts()
	for _, key := range sortedMapKeys(checker.selection) {
		current := checker.selection[key]
		schema := checker.definitions[current.Slot.Facet]
		if current.Slot.Closed && schema.Cardinality == "exactly-one-at-close" && len(current.Values) == 0 {
			checker.result.Obligations = append(checker.result.Obligations, facetObligation(current.Slot.Target, schema.Origin, Unresolved, "exactly-one facet is unset at configuration closure", Predicate{Kind: "present", Subject: schema.ID}))
		}
		checker.result.Selections = append(checker.result.Selections, *current)
	}
	for _, request := range checker.result.Requests {
		status, reason := Entails(checker.result.Facts, request.Predicate)
		checker.result.Obligations = append(checker.result.Obligations, facetObligation(request.Target, request.Origin, status, reason, request.Predicate))
	}
	return checker.result, nil
}
