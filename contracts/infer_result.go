package contracts

// CheckedResult represents the already selected checked-result composition.
// Children evaluate the result as a value. Cleanup contains the actual selected
// check/wrapper work, not inferred exception handling or automatic destruction.
func (a *inference) checkResult(node Node, entry PathSummary) ([]PathSummary, error) {
	values, err := a.sequence(node.Children, []PathSummary{entry})
	if err != nil {
		return nil, err
	}
	var result []PathSummary
	for _, value := range values {
		if value.Outcome != Normal || value.ControlReturn {
			result = append(result, value)
			continue
		}
		observed := value.ObservedOutcome
		checked, err := a.sequence(node.Cleanup, []PathSummary{value})
		if err != nil {
			return nil, err
		}
		for i := range checked {
			if checked[i].Outcome == Normal && !checked[i].ControlReturn {
				if observed == ReturnedFailure || checked[i].ObservedOutcome == ReturnedFailure {
					checked[i].Outcome, checked[i].ObservedOutcome = ReturnedFailure, ReturnedFailure
				} else if checked[i].ObservedOutcome == "" {
					checked[i].ObservedOutcome = observed
				}
			}
		}
		result = append(result, checked...)
	}
	return result, nil
}
