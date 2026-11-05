package contracts

// ResourceValueFact denotes an exact typed scalar observation at one runtime
// resource path. It is a premise supplied by typed transfer, never parsed from
// generated native text or inferred from an arbitrary predicate's display name.
func ResourceValueFact(location FailureLocation, value Atom) Predicate {
	return Predicate{Kind: "eq", Subject: "resource_value:" + StableID("resource-path", location.Path), Resource: location.Resource, Value: value}
}

// TypedFinalValueRelation compares the selected region's immutable entry
// observation with its surviving completion fact. Unknown writes/aliases and
// cleanup invalidate completion facts through the ordinary effect transfer.
func TypedFinalValueRelation(path PathSummary, location FailureLocation) (Status, string) {
	if location.Resource.ID == "" || location.Resource.Unknown {
		return Unresolved, "runtime resource identity is unresolved"
	}
	before, beforeKnown := exactResourceValue(path.EntryFacts, location)
	after, afterKnown := exactResourceValue(path.Facts, location)
	if !beforeKnown || !afterKnown {
		return Unresolved, "exact typed entry and completion values are not both available"
	}
	if before == after {
		return Established, "exact typed entry and completion values are equal"
	}
	if before.Kind != after.Kind {
		return Unresolved, "typed observations have incompatible scalar domains"
	}
	if !path.Reachable {
		return Unresolved, "different typed completion value is only conditionally reachable"
	}
	return Contradicted, "exact typed completion value differs from the selected region entry"
}

func exactResourceValue(facts []Predicate, location FailureLocation) (Atom, bool) {
	subject := ResourceValueFact(location, Atom{}).Subject
	var result Atom
	found, consistent := false, true
	var visit func(Predicate)
	visit = func(fact Predicate) {
		if fact.Kind == "and" {
			for _, child := range fact.Args {
				visit(child)
			}
			return
		}
		if fact.Kind != "eq" || fact.Subject != subject || fact.Resource != location.Resource || fact.Target.Root != "" || fact.Target.Unknown || len(fact.Target.Path) != 0 || len(fact.Args) != 0 || fact.Lower != nil || fact.Upper != nil {
			return
		}
		if fact.Value.Kind != "number" && fact.Value.Kind != "bool" && fact.Value.Kind != "text" {
			return
		}
		if found && result != fact.Value {
			consistent = false
		}
		result, found = fact.Value, true
	}
	for _, fact := range facts {
		visit(fact)
	}
	return result, found && consistent
}
