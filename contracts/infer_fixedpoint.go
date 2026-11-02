package contracts

import "sort"

// Definition SCCs are visited after their callees. Finite effect domains reach
// a fixed point independently of quantitative widening.
func (a *inference) inferDefinitions() error {
	a.summarizing = true
	defer func() { a.summarizing = false }()
	components := definitionSCCs(a.plan)
	for _, component := range components {
		cyclic := len(component) > 1
		if len(component) == 1 {
			for _, callee := range definitionCalls(a.plan.Definitions[component[0]], a.plan.Callbacks, nil) {
				cyclic = cyclic || callee == component[0]
			}
		}
		for _, meaning := range component {
			p := initialPath()
			p.Reachable = !cyclic
			a.summaries[meaning] = summarize([]PathSummary{p})
		}
		rounds := 1
		if cyclic {
			rounds = a.options.MaxSCCRounds
		}
		for round := 0; round < rounds; round++ {
			changed := false
			next := make(map[MeaningKey]DerivedSummary)
			for _, meaning := range component {
				paths, err := a.run(a.plan.Definitions[meaning], []PathSummary{initialPath()})
				if err != nil {
					return err
				}
				if cyclic {
					for i := range paths {
						paths[i].Effects = uniqueEffects(paths[i].Effects)
						paths[i].Unknowns = uniqueUnknowns(paths[i].Unknowns)
						if round >= a.options.WidenAfter {
							paths[i].Resources = UnknownResourceCost("recursive quantitative inference widened; no finite bound established")
							for callback, demand := range paths[i].Demand {
								demand.Lower = 0
								demand.Upper = nil
								demand.Symbol = "recursive"
								paths[i].Demand[callback] = demand
							}
							paths[i].Unknowns = append(paths[i].Unknowns, UnknownReason{Code: "recursive_cost", Message: "recursive quantitative inference widened; no finite bound established", Boundary: string(meaning), Components: []string{"costs", "demand"}})
						}
					}
				}
				value := summarize(a.bound(paths))
				next[meaning] = value
				changed = changed || summaryKey(value) != summaryKey(a.summaries[meaning])
			}
			for meaning, value := range next {
				a.summaries[meaning] = value
			}
			if !changed {
				if cyclic {
					// A finite effect fixed point is not a termination or resource
					// bound. In particular, an empty recursive destructor cannot
					// obtain a zero cost from the initial approximation.
					for _, meaning := range component {
						value := a.summaries[meaning]
						for i := range value.Paths {
							value.Paths[i].Resources = UnknownResourceCost("recursive quantitative inference has no supported decreasing bound")
							value.Paths[i].Unknowns = append(value.Paths[i].Unknowns, UnknownReason{Code: "recursive_cost", Message: "recursive effect fixed point does not establish a finite resource or completion bound", Boundary: string(meaning), Components: []string{"costs", "outcomes", "demand"}})
						}
						a.summaries[meaning] = summarize(value.Paths)
					}
				}
				break
			}
			if cyclic && round == rounds-1 {
				for _, meaning := range component {
					value := a.summaries[meaning]
					for i := range value.Paths {
						value.Paths[i].Unknowns = append(value.Paths[i].Unknowns, UnknownReason{Code: "inference_budget", Message: "recursive summary iteration budget exhausted", Boundary: string(meaning), Components: []string{"effects", "outcomes", "ownership", "costs", "demand"}})
						value.Paths[i].Resources = UnknownResourceCost("recursive inference budget exhausted")
					}
					a.summaries[meaning] = summarize(value.Paths)
				}
			}
		}
	}
	return nil
}

func summaryKey(summary DerivedSummary) string {
	// Path order and redundant alternative copies have no logical force. Keep
	// original node order in Plan and use canonical alternatives for convergence.
	keys := make([]string, 0, len(summary.Paths))
	for _, path := range summary.Paths {
		keys = append(keys, StableID("path", path))
	}
	return StableID("paths", sortedStrings(keys))
}

func definitionCalls(node Node, callbacks map[string]Node, seen map[string]bool) []MeaningKey {
	var result []MeaningKey
	if node.Kind == Call {
		result = append(result, node.Meaning)
	}
	if node.Kind == Callback {
		if seen == nil {
			seen = make(map[string]bool)
		}
		if body, ok := callbacks[node.Callback]; ok && !seen[node.Callback] {
			seen[node.Callback] = true
			result = append(result, definitionCalls(body, callbacks, seen)...)
			delete(seen, node.Callback)
		}
	}
	for _, children := range [][]Node{node.Children, node.Then, node.Else, node.Cleanup} {
		for _, child := range children {
			result = append(result, definitionCalls(child, callbacks, seen)...)
		}
	}
	return result
}

func definitionSCCs(plan Plan) [][]MeaningKey {
	keys := make([]string, 0, len(plan.Definitions))
	for meaning := range plan.Definitions {
		keys = append(keys, string(meaning))
	}
	sort.Strings(keys)
	indices, low := make(map[MeaningKey]int), make(map[MeaningKey]int)
	onStack := make(map[MeaningKey]bool)
	var stack []MeaningKey
	var result [][]MeaningKey
	index := 0
	var visit func(MeaningKey)
	visit = func(meaning MeaningKey) {
		index++
		indices[meaning] = index
		low[meaning] = index
		stack = append(stack, meaning)
		onStack[meaning] = true
		for _, next := range definitionCalls(plan.Definitions[meaning], plan.Callbacks, nil) {
			if _, ok := plan.Definitions[next]; !ok {
				continue
			}
			if indices[next] == 0 {
				visit(next)
				if low[next] < low[meaning] {
					low[meaning] = low[next]
				}
			} else if onStack[next] && indices[next] < low[meaning] {
				low[meaning] = indices[next]
			}
		}
		if low[meaning] != indices[meaning] {
			return
		}
		var component []MeaningKey
		for {
			last := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			onStack[last] = false
			component = append(component, last)
			if last == meaning {
				break
			}
		}
		sort.Slice(component, func(i, j int) bool { return component[i] < component[j] })
		result = append(result, component)
	}
	for _, key := range keys {
		if indices[MeaningKey(key)] == 0 {
			visit(MeaningKey(key))
		}
	}
	return result
}
