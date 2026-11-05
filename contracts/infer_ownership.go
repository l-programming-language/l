package contracts

import "sort"

// OwnershipSnapshot retains correlations between complete capabilities and
// their acquisition order. Alternatives are choices, never sequential work.
type OwnershipSnapshot struct {
	Owned map[string]OwnedResource `json:"owned"`
	Order []string                 `json:"order,omitempty"`
}

// OwnershipPossibility is the finite fallback after snapshot-budget exhaustion.
// It deliberately contains no representative destructor or guessed order.
type OwnershipPossibility struct {
	Resource       ResourceRef `json:"resource"`
	MayLive        bool        `json:"may_live"`
	MayGuard       bool        `json:"may_guard"`
	MayExclusive   bool        `json:"may_exclusive"`
	MayUnsafeGuard bool        `json:"may_unsafe_guard"`
	MayFinalOwner  bool        `json:"may_final_owner"`
	Origin         Origin      `json:"origin"`
}

type OwnershipWidening struct {
	Resources map[string]OwnershipPossibility `json:"resources"`
}

func sortedMapKeys[K ~string, V any](values map[K]V) []K {
	keys := make([]K, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

func cloneOwnershipSnapshot(snapshot OwnershipSnapshot) OwnershipSnapshot {
	owned := make(map[string]OwnedResource, len(snapshot.Owned))
	for id, value := range snapshot.Owned {
		owned[id] = normalizedOwnedResource(value)
	}
	return OwnershipSnapshot{Owned: owned, Order: append([]string(nil), snapshot.Order...)}
}

func normalizedOwnedResource(owned OwnedResource) OwnedResource {
	if owned.FinalOwner == "" {
		owned.FinalOwner = Unresolved
	}
	return owned
}

func cloneOwnershipSnapshots(snapshots []OwnershipSnapshot) []OwnershipSnapshot {
	if len(snapshots) == 0 {
		return nil
	}
	result := make([]OwnershipSnapshot, len(snapshots))
	for i, snapshot := range snapshots {
		result[i] = cloneOwnershipSnapshot(snapshot)
	}
	return result
}

func cloneOwnershipWidening(widening *OwnershipWidening) *OwnershipWidening {
	if widening == nil {
		return nil
	}
	result := &OwnershipWidening{Resources: make(map[string]OwnershipPossibility, len(widening.Resources))}
	for id, value := range widening.Resources {
		result.Resources[id] = value
	}
	return result
}

func ownershipSnapshots(p PathSummary) []OwnershipSnapshot {
	if len(p.OwnershipAlternatives) > 0 {
		return p.OwnershipAlternatives
	}
	return []OwnershipSnapshot{{Owned: p.Owned, Order: p.OwnershipOrder}}
}

func expandOwnership(p PathSummary) []PathSummary {
	if len(p.OwnershipAlternatives) == 0 {
		return []PathSummary{p}
	}
	alternatives := p.OwnershipAlternatives
	p.OwnershipAlternatives = nil
	paths := make([]PathSummary, 0, len(alternatives))
	for _, snapshot := range alternatives {
		p.Owned, p.OwnershipOrder = snapshot.Owned, snapshot.Order
		paths = append(paths, clonePath(p))
	}
	return paths
}

func (a *inference) joinOwnership(left *PathSummary, right PathSummary) {
	snapshots := make(map[string]OwnershipSnapshot)
	for _, path := range []PathSummary{*left, right} {
		for _, snapshot := range ownershipSnapshots(path) {
			snapshots[StableID("ownership", snapshot)] = snapshot
		}
	}
	if left.OwnershipWidening != nil || right.OwnershipWidening != nil || len(snapshots) > a.options.normalized().MaxAlternatives {
		widening := &OwnershipWidening{Resources: make(map[string]OwnershipPossibility)}
		for _, path := range []PathSummary{*left, right} {
			collectOwnershipPossibilities(widening, path)
		}
		left.Owned = make(map[string]OwnedResource)
		left.OwnershipOrder, left.OwnershipAlternatives = nil, nil
		left.OwnershipWidening = widening
		return
	}
	left.OwnershipAlternatives = nil
	for _, key := range sortedMapKeys(snapshots) {
		left.OwnershipAlternatives = append(left.OwnershipAlternatives, cloneOwnershipSnapshot(snapshots[key]))
	}
	left.Owned = make(map[string]OwnedResource)
	left.OwnershipOrder = nil
	if len(left.OwnershipAlternatives) == 1 {
		snapshot := left.OwnershipAlternatives[0]
		left.Owned, left.OwnershipOrder = snapshot.Owned, snapshot.Order
		left.OwnershipAlternatives = nil
	}
}

func mergeOwnershipPossibility(widening *OwnershipWidening, id string, value OwnershipPossibility) {
	old, exists := widening.Resources[id]
	if !exists {
		widening.Resources[id] = value
		return
	}
	old.MayLive = old.MayLive || value.MayLive
	old.MayGuard = old.MayGuard || value.MayGuard
	old.MayExclusive = old.MayExclusive || value.MayExclusive
	old.MayUnsafeGuard = old.MayUnsafeGuard || value.MayUnsafeGuard
	old.MayFinalOwner = old.MayFinalOwner || value.MayFinalOwner
	if StableID("origin", value.Origin) < StableID("origin", old.Origin) {
		old.Origin = value.Origin
	}
	old.Resource.Unknown = old.Resource.Unknown || value.Resource.Unknown
	widening.Resources[id] = old
}

func collectOwnershipPossibilities(widening *OwnershipWidening, p PathSummary) {
	for _, snapshot := range ownershipSnapshots(p) {
		for id, owned := range snapshot.Owned {
			live := owned.State == "owned" || owned.State == "conditional"
			mergeOwnershipPossibility(widening, id, OwnershipPossibility{Resource: owned.Resource, MayLive: live, MayGuard: live && owned.Mode == "guard", MayExclusive: live && owned.Mode == "exclusive", MayUnsafeGuard: live && owned.Mode == "guard" && !owned.SuspensionSafe, MayFinalOwner: live && owned.FinalOwner != Contradicted, Origin: owned.Origin})
		}
	}
	if p.OwnershipWidening != nil {
		for id, value := range p.OwnershipWidening.Resources {
			mergeOwnershipPossibility(widening, id, value)
		}
	}
}

func widenedResourceIDs(widening *OwnershipWidening) []string {
	if widening == nil {
		return nil
	}
	return sortedMapKeys(widening.Resources)
}

func (p *PathSummary) takeWidened(id string) (OwnershipPossibility, bool) {
	if p.OwnershipWidening == nil {
		return OwnershipPossibility{}, false
	}
	owned, ok := p.OwnershipWidening.Resources[id]
	delete(p.OwnershipWidening.Resources, id)
	return owned, ok
}

func (a *inference) consumeWidenedOwnership(p *PathSummary, node Node, operation string) bool {
	owned, exists := p.takeWidened(node.Resource.ID)
	if !exists {
		return false
	}
	if owned.MayLive || operation == "move" {
		p.Obligations = append(p.Obligations, Obligation{ID: StableID("widened-ownership-use", node), Claim: Claim{Predicate: Predicate{Kind: "ownership", Resource: node.Resource}, Scope: "use"}, Status: Unresolved, Reason: operation + " initialization requirement depends on widened ownership alternatives", Origin: node.Origin})
	}
	return true
}

func (a *inference) releaseWidenedOwnership(p *PathSummary, node Node) bool {
	owned, exists := p.takeWidened(node.Resource.ID)
	if !exists {
		return false
	}
	if !owned.MayLive {
		return true
	}
	phase := node.Phase
	if phase == "" {
		phase = Destruction
	}
	a.observeEffect(p, Effect{Kind: "release", Resource: owned.Resource, Phase: phase, Origin: node.Origin})
	if node.Costs != nil {
		p.Resources = SequenceResources(p.Resources, *node.Costs)
	}
	if owned.MayFinalOwner {
		a.unknown(p, Node{Origin: owned.Origin, Resource: owned.Resource, Phase: phase}, "ownership_budget_cleanup", "destructor alternatives or release order exceeded the ownership budget")
	}
	return true
}

func (a *inference) importOwnership(path, tail PathSummary, sub Substitutions) []PathSummary {
	var result []PathSummary
	for _, alternative := range expandOwnership(tail) {
		current := clonePath(path)
		for _, key := range alternative.OwnershipOrder {
			owned := normalizedOwnedResource(alternative.Owned[key])
			owned.Resource = rebaseResource(owned.Resource, sub)
			owned.Destructor = append([]Node(nil), owned.Destructor...)
			for i := range owned.Destructor {
				owned.Destructor[i] = rebaseNode(owned.Destructor[i], sub)
			}
			if previous, exists := current.Owned[owned.Resource.ID]; exists && previous.State == "owned" {
				current.Obligations = append(current.Obligations, Obligation{ID: StableID("summary-ownership-collision", owned.Resource), Claim: Claim{Predicate: Predicate{Kind: "ownership", Resource: owned.Resource}, Scope: "use"}, Status: Unresolved, Reason: "callee ownership transfer overlaps a live caller resource without an ordered transfer contract", Origin: owned.Origin})
			}
			current.Owned[owned.Resource.ID] = owned
			current.OwnershipOrder = append(current.OwnershipOrder, owned.Resource.ID)
		}
		if alternative.OwnershipWidening != nil {
			if current.OwnershipWidening == nil {
				current.OwnershipWidening = &OwnershipWidening{Resources: make(map[string]OwnershipPossibility)}
			}
			for _, value := range alternative.OwnershipWidening.Resources {
				value.Resource = rebaseResource(value.Resource, sub)
				mergeOwnershipPossibility(current.OwnershipWidening, value.Resource.ID, value)
			}
		}
		result = append(result, current)
	}
	return a.bound(result)
}
