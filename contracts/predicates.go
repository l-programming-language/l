package contracts

import (
	"reflect"
	"slices"
	"strings"
)

// authoritativeTargetFacts binds the conventional target type predicate to
// the compiler's concrete attachment. Caller premises cannot change that type.
// A missing concrete type cannot be filled in by an asserted caller fact.
func authoritativeTargetFacts(target TargetRef, supplied []Predicate) ([]Predicate, Status, string) {
	status, reason := Established, "caller target type facts agree with the typed attachment"
	var retained []Predicate
	isTargetType := func(predicate Predicate) bool {
		return predicate.Kind == "type" && predicate.Subject == "target" && (predicate.Target.Root == "" && len(predicate.Target.Path) == 0 && !predicate.Target.Unknown || TargetRelation(predicate.Target, target) == Established)
	}
	var collect func(Predicate)
	collect = func(fact Predicate) {
		if fact.Kind == "and" {
			for _, child := range fact.Args {
				collect(child)
			}
			return
		}
		predicate, negative := fact, false
		if fact.Kind == "not" && len(fact.Args) == 1 {
			predicate, negative = fact.Args[0], true
		}
		if !isTargetType(predicate) {
			retained = append(retained, fact)
			return
		}
		if target.Type == "" {
			if status != Contradicted {
				status, reason = Unresolved, "caller type premise has no authoritative concrete target type"
			}
			return
		}
		if (predicate.Value.Text == target.Type) == negative {
			status, reason = Contradicted, "caller type premise contradicts authoritative target type "+target.Type
		}
	}
	for _, fact := range supplied {
		collect(fact)
	}
	if target.Type != "" {
		actual := Predicate{Kind: "type", Subject: "target", Value: Atom{Kind: "type", Text: target.Type}}
		scoped := actual
		scoped.Target = target
		retained = append([]Predicate{actual, scoped}, retained...)
	}
	return retained, status, reason
}

// Entails checks the finite contract fragment. Facts must come from the typed
// analysis or explicit certificate premises; calling this function does not
// turn an authored assertion into implementation evidence.
func Entails(facts []Predicate, want Predicate) (Status, string) {
	return entails(facts, want, 0)
}

func entails(facts []Predicate, want Predicate, depth int) (Status, string) {
	if depth > 64 {
		return Unresolved, "predicate nesting exceeds the supported bound"
	}
	switch want.Kind {
	case "true":
		return Established, "constant true"
	case "false":
		return Contradicted, "constant false"
	case "and":
		result := Established
		for _, arg := range want.Args {
			s, _ := entails(facts, arg, depth+1)
			if s == Contradicted {
				return Contradicted, "a conjunct is contradicted"
			}
			if s != Established {
				result = Unresolved
			}
		}
		return result, "finite conjunction"
	case "or":
		if len(want.Args) > 64 {
			return Unresolved, "disjunction exceeds the supported bound"
		}
		result := Contradicted
		for _, arg := range want.Args {
			s, _ := entails(facts, arg, depth+1)
			if s == Established {
				return Established, "a disjunct is established"
			}
			if s != Contradicted {
				result = Unresolved
			}
		}
		return result, "finite disjunction"
	case "not":
		if len(want.Args) != 1 {
			return Unresolved, "negation requires one predicate"
		}
		s, reason := entails(facts, want.Args[0], depth+1)
		if s == Established {
			return Contradicted, reason
		}
		if s == Contradicted {
			return Established, reason
		}
		return Unresolved, reason
	case "implies":
		if len(want.Args) != 2 {
			return Unresolved, "conditional requirement requires two predicates"
		}
		s, _ := entails(facts, want.Args[0], depth+1)
		if s == Contradicted {
			return Established, "condition is refuted"
		}
		if s != Established {
			return Unresolved, "condition is undetermined"
		}
		return entails(facts, want.Args[1], depth+1)
	case "eq", "type", "meaning", "present", "finite", "interval", "ownership", "escape", "protocol", "resource_equal", "resource_disjoint", "effect_contains", "effects_closed", "multiplicity", "cost_leq", "facet", "capability", "member":
	default:
		return Unresolved, "unsupported predicate: " + want.Kind
	}
	if want.Kind == "resource_equal" || want.Kind == "resource_disjoint" {
		var aliases, disjoint []ResourcePair
		for _, fact := range facts {
			pair := ResourcePair{Left: fact.Resource, Right: ResourceRef{ID: fact.Value.Text}}
			if fact.Kind == "resource_equal" {
				aliases = append(aliases, pair)
			} else if fact.Kind == "resource_disjoint" {
				disjoint = append(disjoint, pair)
			}
		}
		if want.Resource.ID != "" && want.Value.Text != "" {
			relation := AliasRelation(want.Resource, ResourceRef{ID: want.Value.Text}, aliases, disjoint)
			if want.Kind == "resource_equal" {
				return relation, "canonical resource identity and explicit alias/disjointness premises"
			}
			if relation == Established {
				return Contradicted, "resources have established identical identity"
			}
			if relation == Contradicted {
				return Established, "explicit resource disjointness premise"
			}
			return Unresolved, "resource disjointness is not established"
		}
	}
	if want.Kind == "effect_contains" {
		if closed, _ := entails(facts, Predicate{Kind: "effects_closed", Subject: want.Subject, Target: want.Target, Resource: want.Resource}, depth+1); closed != Established {
			return Unresolved, "effect containment requires a closed effect row"
		}
	}
	for _, fact := range facts {
		if fact.Kind == "and" {
			if s, _ := entails(fact.Args, want, depth+1); s == Established {
				return Established, "conjunctive premise"
			}
		}
		if fact.Kind == "not" && len(fact.Args) == 1 && reflect.DeepEqual(fact.Args[0], want) {
			return Contradicted, "explicit negative premise"
		}
		if reflect.DeepEqual(fact, want) {
			return Established, "matching supported premise"
		}
		if fact.Subject != want.Subject || !reflect.DeepEqual(fact.Target, want.Target) || fact.Resource != want.Resource {
			continue
		}
		if want.Kind == "protocol" && fact.Kind == "protocol" {
			states := fact.Args
			if len(states) == 0 && fact.Value.Text != "" {
				states = []Predicate{{Value: fact.Value}}
			}
			allowed := want.Args
			if len(allowed) == 0 && want.Value.Text != "" {
				allowed = []Predicate{{Value: want.Value}}
			}
			if len(states) > 0 && len(allowed) > 0 {
				inside, outside := false, false
				for _, state := range states {
					matched := false
					for _, value := range allowed {
						matched = matched || state.Value == value.Value
					}
					inside = inside || matched
					outside = outside || !matched
				}
				if !outside {
					return Established, "every possible protocol state satisfies the required state set"
				}
				if !inside {
					return Contradicted, "possible protocol states exclude the required state set"
				}
				return Unresolved, "protocol alternatives include an inadmissible state"
			}
		}
		if want.Kind == "effect_contains" && fact.Kind == "effect_contains" {
			closed, _ := entails(facts, Predicate{Kind: "effects_closed", Subject: want.Subject, Target: want.Target, Resource: want.Resource}, depth+1)
			if closed != Established {
				return Unresolved, "effect containment requires a closed effect row"
			}
			for _, effect := range fact.Args {
				found := false
				for _, allowed := range want.Args {
					found = found || reflect.DeepEqual(effect, allowed)
				}
				if !found {
					return Unresolved, "a conservative possible effect is outside the allowed row"
				}
			}
			return Established, "closed effect row is contained in the allowed row"
		}
		if (want.Kind == "interval" || want.Kind == "multiplicity" || want.Kind == "cost_leq") && (fact.Kind == want.Kind || fact.Kind == "eq") {
			lower, upper := fact.Lower, fact.Upper
			if fact.Kind == "eq" && fact.Value.Kind == "number" {
				lower, upper = &fact.Value.Number, &fact.Value.Number
			}
			if lower != nil && upper != nil && *lower > *upper {
				continue
			}
			if (want.Lower == nil || lower != nil && *lower >= *want.Lower) && (want.Upper == nil || upper != nil && *upper <= *want.Upper) {
				return Established, "supported interval containment"
			}
			if want.Lower != nil && upper != nil && *upper < *want.Lower || want.Upper != nil && lower != nil && *lower > *want.Upper {
				return Contradicted, "disjoint supported intervals"
			}
		}
		if want.Kind == "finite" && fact.Kind == "eq" && fact.Value.Kind == "number" {
			return Established, "integer constant is finite"
		}
		if fact.Kind == want.Kind && fact.Value != want.Value {
			switch want.Kind {
			case "eq", "type", "meaning", "facet", "ownership":
				return Contradicted, "incompatible exact value"
			}
		}
	}
	return Unresolved, "no supported entailment for " + want.Kind + " " + want.Subject
}

// TargetRelation compares structural attachment identities, not their types or
// the resources their implementations may access.
func TargetRelation(left, right TargetRef) Status {
	if left.Unknown || right.Unknown || left.Root == "" || right.Root == "" {
		return Unresolved
	}
	if left.Root == right.Root && slices.Equal(left.Path, right.Path) && left.Phase == right.Phase && left.Generation == right.Generation {
		return Established
	}
	return Contradicted
}

func pathContains(parent, child TargetRef) bool {
	if parent.Unknown || child.Unknown || parent.Root == "" || parent.Root != child.Root || parent.Phase != child.Phase || parent.Generation != child.Generation || len(parent.Path) > len(child.Path) {
		return false
	}
	for i := range parent.Path {
		if parent.Path[i] != child.Path[i] {
			return false
		}
	}
	return true
}

func targetKey(target TargetRef) string {
	return StableID("target", struct {
		Root       string
		Path       string
		Phase      Phase
		Generation string
		Unknown    bool
	}{target.Root, strings.Join(target.Path, "\x00"), target.Phase, target.Generation, target.Unknown})
}
