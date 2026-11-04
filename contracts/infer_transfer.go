package contracts

// observeEffect applies caller-context consequences only. Resource accounting
// stays with the event or already-composed callee summary, never both.
func (a *inference) observeEffect(p *PathSummary, effect Effect) {
	switch effect.Kind {
	case "read", "write", "replace", "release", "allocate", "block", "io", "panic", "spawn", "signal", "wait", "suspend", "retain", "escape", "close":
	default:
		a.unknown(p, Node{Origin: effect.Origin, Phase: effect.Phase}, "unsupported_effect", "typed event has an unsupported effect classification: "+string(effect.Kind))
	}
	if !p.Reachable {
		effect.Definite = false
	}
	p.Effects = append(p.Effects, effect)
	if effectMayModify(effect.Kind) {
		a.invalidate(p, effect.Resource)
		a.invalidateStructural(p, effect.Target, string(effect.Kind))
	}
	if effect.Kind != "suspend" {
		return
	}
	for _, owned := range p.Owned {
		if (owned.State == "owned" || owned.State == "conditional") && owned.Mode == "guard" && !owned.SuspensionSafe {
			status := Unresolved
			if owned.State == "owned" && effect.Definite && p.Reachable {
				status = Contradicted
			}
			a.guardObligation(p, owned.Resource, effect.Origin, status)
		}
	}
	if p.OwnershipWidening != nil {
		for _, owned := range p.OwnershipWidening.Resources {
			if owned.MayLive && owned.MayUnsafeGuard {
				a.guardObligation(p, owned.Resource, effect.Origin, Unresolved)
			}
		}
	}
}

func effectMayModify(kind EffectKind) bool {
	switch kind {
	case "write", "replace", "release", "io", "close", "deallocate", "spawn", "transfer":
		return true
	}
	return false
}

func (a *inference) guardObligation(p *PathSummary, resource ResourceRef, origin Origin, status Status) {
	p.Obligations = append(p.Obligations, Obligation{ID: StableID("suspension", resource), Claim: Claim{Predicate: Predicate{Kind: "suspension_safe", Resource: resource}, Scope: "whole_region"}, Status: status, Reason: "non-suspension-safe guard may be live across suspension", Origin: origin})
}

// Imported unknown behavior has the same interference consequences as inline
// unknown behavior, without adding its already-accounted resource costs again.
func (a *inference) observeUnknown(p *PathSummary, unknown UnknownReason) {
	relevant := len(unknown.Components) == 0
	for _, component := range unknown.Components {
		relevant = relevant || component == "effects" || component == "ownership"
	}
	if !relevant {
		return
	}
	a.invalidate(p, ResourceRef{Unknown: true})
	a.invalidateStructural(p, TargetRef{Unknown: true}, "unknown")
	resources := make(map[string]ResourceRef)
	for id, owned := range p.Owned {
		if (owned.State == "owned" || owned.State == "conditional") && (owned.Mode == "exclusive" || owned.Mode == "guard") {
			resources[id] = owned.Resource
		}
	}
	if p.OwnershipWidening != nil {
		for id, owned := range p.OwnershipWidening.Resources {
			if owned.MayLive && (owned.MayGuard || owned.MayExclusive) {
				resources[id] = owned.Resource
			}
		}
	}
	for _, id := range sortedMapKeys(resources) {
		resource := resources[id]
		p.Obligations = append(p.Obligations, Obligation{ID: StableID("unknown-interference", struct {
			Resource ResourceRef
			Origin   Origin
		}{resource, unknown.Origin}), Claim: Claim{Predicate: Predicate{Kind: "ownership", Resource: resource}, Scope: "whole_region"}, Status: Unresolved, Reason: "unknown behavior may interfere with a live exclusive resource or guard", Origin: unknown.Origin})
	}
}

// Every predicate wrapper retains the dependencies of every child. Removing a
// whole invalidated wrapper is conservative even when another child is stable.
func (a *inference) resourceFactAffected(fact Predicate, resource ResourceRef) bool {
	if (fact.Resource.ID != "" || fact.Resource.Unknown) && AliasRelation(fact.Resource, resource, a.options.Aliases, a.options.Disjoint) != Contradicted {
		return true
	}
	for _, child := range fact.Args {
		if a.resourceFactAffected(child, resource) {
			return true
		}
	}
	return false
}

func commonStateRelations(left, right []StateRelation) []StateRelation {
	available := make(map[string]bool, len(right))
	for _, relation := range right {
		available[StableID("state-relation", relation)] = true
	}
	var common []StateRelation
	for _, relation := range left {
		if available[StableID("state-relation", relation)] {
			common = append(common, relation)
		}
	}
	return common
}

// A callee's entry is not automatically the enclosing region's entry. A
// relative state relation may compose only while those observations agree.
func (a *inference) relationHasRegionEntry(prefix PathSummary, relation StateRelation) bool {
	for _, previous := range prefix.StateRelations {
		if previous.Relation == "unchanged" && previous.Observation == relation.Observation && StableID("location", previous.Location) == StableID("location", relation.Location) {
			return true
		}
	}
	for _, unknown := range prefix.Unknowns {
		if len(unknown.Components) == 0 {
			return false
		}
		for _, component := range unknown.Components {
			if component == "effects" || component == "ownership" {
				return false
			}
		}
	}
	for _, effect := range prefix.Effects {
		if !effectMayModify(effect.Kind) {
			continue
		}
		if AliasRelation(effect.Resource, relation.Location.Resource, a.options.Aliases, a.options.Disjoint) != Contradicted {
			return false
		}
	}
	return true
}

func joinDemands(left, right map[string]Multiplicity) map[string]Multiplicity {
	result := make(map[string]Multiplicity, len(left)+len(right))
	for key := range left {
		result[key] = joinDemand(demandAt(left, key), demandAt(right, key))
	}
	for key := range right {
		if _, exists := result[key]; !exists {
			result[key] = joinDemand(demandAt(left, key), demandAt(right, key))
		}
	}
	return result
}

func joinDemand(left, right Multiplicity) Multiplicity {
	result := Multiplicity{Lower: left.Lower}
	if right.Lower < result.Lower {
		result.Lower = right.Lower
	}
	if left.Upper != nil && right.Upper != nil {
		upper := *left.Upper
		if *right.Upper > upper {
			upper = *right.Upper
		}
		result.Upper = &upper
	}
	if left.Symbol == right.Symbol {
		result.Symbol = left.Symbol
	}
	return result
}
