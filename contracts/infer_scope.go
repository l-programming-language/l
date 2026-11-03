package contracts

import "fmt"

// InferScope analyzes a selected typed node under every supported actual entry
// context. Earlier effects/costs are excluded; entry facts, ownership, approved
// assumptions and unresolved preconditions remain. This executes no source App.
// The coordinator separately enforces the selected region's coverage inventory.
func InferScope(plan Plan, nodeID string, options AnalysisOptions) (DerivedSummary, error) {
	if nodeID == "" {
		return DerivedSummary{}, fmt.Errorf("scoped inference requires a selected typed node identity")
	}
	return inferPlan(plan, options, nodeID)
}

func (a *inference) measureScope(node Node, entry PathSummary) error {
	path := clonePath(entry)
	path.EntryFacts = append([]Predicate(nil), entry.Facts...)
	path.Effects, path.Unknowns, path.NativeBoundaries, path.StateRelations = nil, nil, nil, nil
	path.ObservedOutcome = ""
	path.Demand = make(map[string]Multiplicity)
	path.Resources = ZeroResourceCost()
	path.Obligations = nil
	for _, prerequisite := range entry.Obligations {
		if prerequisite.Status != Established {
			path.Obligations = append(path.Obligations, prerequisite)
		}
	}
	a.measuringScope = true
	paths, err := a.step(node, path)
	a.measuringScope = false
	if err != nil {
		return err
	}
	a.scopedPaths = append(a.scopedPaths, paths...)
	a.scopedPaths = a.bound(a.scopedPaths)
	return nil
}
