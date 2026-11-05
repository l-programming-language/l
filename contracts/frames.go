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

type FrameContext struct {
	Disjoint []ResourcePair `json:"disjoint,omitempty"`
	Aliases  []ResourcePair `json:"aliases,omitempty"`
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
			var alias Status
			if frame.Scope == "runtime" {
				alias = AliasRelation(frame.Resource, write.Resource, context.Aliases, context.Disjoint)
			}
			if alias == Unresolved {
				if status != Contradicted {
					status, reason = Unresolved, "runtime resources may alias despite different structural paths"
				}
				continue
			}
			violation := false
			if preserved != nil {
				violation = alias == Established || pathContains(*preserved, write.Target) || pathContains(write.Target, *preserved)
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

func checkSummaryFrame(frame FrameContract, summary DerivedSummary, origin Origin) []Obligation {
	writes := FootprintsFromEffects(summary.Effects)
	if len(summary.Unknowns) > 0 {
		writes = append(writes, FrameFootprint{Kind: "unknown", Origin: origin})
	}
	return CheckFrame(frame, writes, FrameContext{})
}
