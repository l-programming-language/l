package contracts

type FrameFootprint struct {
	Target    TargetRef   `json:"target"`
	Resource  ResourceRef `json:"resource"`
	Kind      string      `json:"kind"`
	Origin    Origin      `json:"origin"`
	Definite  bool        `json:"definite"`
	Condition string      `json:"condition,omitempty"`
}

type FrameContract struct {
	Target    TargetRef   `json:"target"`
	Allowed   []TargetRef `json:"allowed"`
	Preserves []TargetRef `json:"preserves"`
	Resource  ResourceRef `json:"resource"`
	Scope     string      `json:"scope"`
	Phase     Phase       `json:"phase"`
	Origin    Origin      `json:"origin"`
}

type FrameResourcePair struct {
	Left  ResourceRef `json:"left"`
	Right ResourceRef `json:"right"`
}

type FrameContext struct {
	Disjoint []FrameResourcePair `json:"disjoint,omitempty"`
	Aliases  []FrameResourcePair `json:"aliases,omitempty"`
}

func frameResourcesDisjoint(left, right ResourceRef, context FrameContext) bool {
	return frameResourceRelation(left, right, context) == Contradicted
}

func frameResourceRelation(left, right ResourceRef, context FrameContext) Status {
	aliases := make([]ResourcePair, 0, len(context.Aliases))
	disjoint := make([]ResourcePair, 0, len(context.Disjoint))
	for _, pair := range context.Aliases {
		aliases = append(aliases, ResourcePair(pair))
	}
	for _, pair := range context.Disjoint {
		disjoint = append(disjoint, ResourcePair(pair))
	}
	return AliasRelation(left, right, aliases, disjoint)
}

// CheckFrame checks actual typed writes and replacements. Structural identity
// is sufficient only for elaboration frames; runtime frames need resource facts.
func CheckFrame(frame FrameContract, writes []FrameFootprint, context FrameContext) []Obligation {
	var obligations []Obligation
	check := func(preserved *TargetRef) Obligation {
		status, reason := Established, "actual typed footprint preserves the declared frame"
		for _, write := range writes {
			if frame.Phase != "" && write.Target.Phase != "" && frame.Phase != write.Target.Phase {
				continue
			}
			if write.Kind == "unknown" || write.Target.Unknown || write.Target.Root == "" {
				if status != Contradicted {
					status, reason = Unresolved, "unknown write footprint prevents preservation proof"
				}
				continue
			}
			if frame.Scope == "runtime" && !frameResourcesDisjoint(frame.Resource, write.Resource, context) && frameResourceRelation(frame.Resource, write.Resource, context) != Established {
				if status != Contradicted {
					status, reason = Unresolved, "runtime resources may alias despite different structural paths"
				}
				continue
			}
			violation := false
			if preserved != nil {
				violation = pathContains(*preserved, write.Target) || pathContains(write.Target, *preserved)
				if frame.Scope == "runtime" && frameResourceRelation(frame.Resource, write.Resource, context) == Established {
					violation = true
				}
			} else {
				inside := false
				for _, allowed := range frame.Allowed {
					if pathContains(allowed, write.Target) {
						inside = true
					}
				}
				violation = !inside
			}
			if violation {
				if write.Definite && write.Condition == "" {
					status, reason = Contradicted, "typed "+write.Kind+" at "+write.Target.String()+" violates the declared frame"
				} else if status != Contradicted {
					status, reason = Unresolved, "possible typed write overlaps the protected frame"
				}
			}
		}
		target := frame.Target
		if preserved != nil {
			target = *preserved
		}
		obligation := facetObligation(target, frame.Origin, status, reason, Predicate{Kind: "frame", Target: target})
		obligation.Claim.Scope, obligation.Claim.Phase = frame.Scope, frame.Phase
		for _, write := range writes {
			obligation.Dependencies = append(obligation.Dependencies, StableID("typed-write", write))
		}
		return obligation
	}
	obligations = append(obligations, check(nil))
	for _, preserved := range frame.Preserves {
		preserved := preserved
		obligations = append(obligations, check(&preserved))
	}
	return obligations
}

func FootprintsFromEffects(effects []Effect) []FrameFootprint {
	var footprints []FrameFootprint
	for _, effect := range effects {
		switch effect.Kind {
		case "write", "replace":
			footprints = append(footprints, FrameFootprint{Target: effect.Target, Resource: effect.Resource, Kind: string(effect.Kind), Origin: effect.Origin, Definite: effect.Definite, Condition: effect.Condition})
		case "opaque", "unknown":
			footprints = append(footprints, FrameFootprint{Target: effect.Target, Resource: effect.Resource, Kind: "unknown", Origin: effect.Origin})
		}
	}
	return footprints
}
