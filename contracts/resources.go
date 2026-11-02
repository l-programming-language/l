package contracts

import (
	"math"
	"sort"
)

// CostExpr is a bounded symbolic cost expression. Unknown is not infinity:
// it says that this analyzer did not establish a finite bound.
type CostExpr struct {
	Kind   string     `json:"kind"`
	Value  int64      `json:"value,omitempty"`
	Name   string     `json:"name,omitempty"`
	Args   []CostExpr `json:"args,omitempty"`
	Reason string     `json:"reason,omitempty"`
}

func ConstantCost(value int64) CostExpr    { return CostExpr{Kind: "constant", Value: value} }
func SymbolicCost(name string) CostExpr    { return CostExpr{Kind: "parameter", Name: name} }
func UnknownCost(reason string) CostExpr   { return CostExpr{Kind: "unknown", Reason: reason} }
func (c CostExpr) Constant() (int64, bool) { return c.Value, c.Kind == "constant" }

func normalizeCost(c CostExpr) CostExpr {
	if c.Kind == "" {
		return UnknownCost("cost was not supplied")
	}
	return c
}

func addInteger(a, b int64) (int64, bool) {
	if b > 0 && a > math.MaxInt64-b || b < 0 && a < math.MinInt64-b {
		return 0, false
	}
	return a + b, true
}

func AddCost(a, b CostExpr) CostExpr {
	a, b = normalizeCost(a), normalizeCost(b)
	if a.Kind == "unknown" {
		return a
	}
	if b.Kind == "unknown" {
		return b
	}
	if a.Kind == "constant" && a.Value == 0 {
		return b
	}
	if b.Kind == "constant" && b.Value == 0 {
		return a
	}
	if a.Kind == "constant" && b.Kind == "constant" {
		if value, ok := addInteger(a.Value, b.Value); ok {
			return ConstantCost(value)
		}
		return UnknownCost("cost arithmetic overflow")
	}
	return CostExpr{Kind: "add", Args: []CostExpr{a, b}}
}

func MultiplyCost(a, b CostExpr) CostExpr {
	a, b = normalizeCost(a), normalizeCost(b)
	// Zero repetitions perform no callback work, even when its latent cost is unknown.
	if a.Kind == "constant" && a.Value == 0 || b.Kind == "constant" && b.Value == 0 {
		return ConstantCost(0)
	}
	if a.Kind == "unknown" {
		return a
	}
	if b.Kind == "unknown" {
		return b
	}
	if a.Kind == "constant" && a.Value == 1 {
		return b
	}
	if b.Kind == "constant" && b.Value == 1 {
		return a
	}
	if a.Kind == "constant" && b.Kind == "constant" {
		if a.Value == math.MinInt64 && b.Value == -1 || b.Value == math.MinInt64 && a.Value == -1 {
			return UnknownCost("cost arithmetic overflow")
		}
		value := a.Value * b.Value
		if value/b.Value != a.Value {
			return UnknownCost("cost arithmetic overflow")
		}
		return ConstantCost(value)
	}
	return CostExpr{Kind: "multiply", Args: []CostExpr{a, b}}
}

func MaxCost(a, b CostExpr) CostExpr {
	a, b = normalizeCost(a), normalizeCost(b)
	if a.Kind == "unknown" {
		return a
	}
	if b.Kind == "unknown" {
		return b
	}
	if a.Kind == "constant" && b.Kind == "constant" {
		if a.Value > b.Value {
			return a
		}
		return b
	}
	if StableID("cost", a) == StableID("cost", b) {
		return a
	}
	return CostExpr{Kind: "max", Args: []CostExpr{a, b}}
}

// ResourceCost separates cumulative work from exact additional live-resource
// accounting. NetLive is signed; the other quantitative metrics are nonnegative.
// Retained is additional owned live resource at the selected return boundary,
// not allocator-retained or process-resident memory.
type ResourceCost struct {
	Model          string   `json:"model"`
	Work           CostExpr `json:"work"`
	Allocated      CostExpr `json:"allocated"`
	PayloadCopied  CostExpr `json:"payload_copied"`
	MetadataCopied CostExpr `json:"metadata_copied"`
	IO             CostExpr `json:"io"`
	NetLive        CostExpr `json:"net_live"`
	PeakLive       CostExpr `json:"peak_live"`
	Retained       CostExpr `json:"retained"`
}

func ZeroResourceCost() ResourceCost {
	z := ConstantCost(0)
	return ResourceCost{Model: "additional-owned-live-units/1", Work: z, Allocated: z, PayloadCopied: z, MetadataCopied: z, IO: z, NetLive: z, PeakLive: z, Retained: z}
}

func UnknownResourceCost(reason string) ResourceCost {
	u := UnknownCost(reason)
	return ResourceCost{Model: "additional-owned-live-units/1", Work: u, Allocated: u, PayloadCopied: u, MetadataCopied: u, IO: u, NetLive: u, PeakLive: u, Retained: u}
}

func SequenceResources(a, b ResourceCost) ResourceCost {
	if a.Model == "" || b.Model == "" || a.Model != b.Model {
		return UnknownResourceCost("incompatible or omitted resource accounting model")
	}
	net := AddCost(a.NetLive, b.NetLive)
	return ResourceCost{Model: a.Model, Work: AddCost(a.Work, b.Work), Allocated: AddCost(a.Allocated, b.Allocated), PayloadCopied: AddCost(a.PayloadCopied, b.PayloadCopied), MetadataCopied: AddCost(a.MetadataCopied, b.MetadataCopied), IO: AddCost(a.IO, b.IO), NetLive: net, PeakLive: MaxCost(a.PeakLive, AddCost(a.NetLive, b.PeakLive)), Retained: MaxCost(ConstantCost(0), net)}
}

func BranchResources(a, b ResourceCost) ResourceCost {
	if a.Model == "" || a.Model != b.Model {
		return UnknownResourceCost("incompatible resource accounting models at branch")
	}
	return ResourceCost{Model: a.Model, Work: MaxCost(a.Work, b.Work), Allocated: MaxCost(a.Allocated, b.Allocated), PayloadCopied: MaxCost(a.PayloadCopied, b.PayloadCopied), MetadataCopied: MaxCost(a.MetadataCopied, b.MetadataCopied), IO: MaxCost(a.IO, b.IO), NetLive: MaxCost(a.NetLive, b.NetLive), PeakLive: MaxCost(a.PeakLive, b.PeakLive), Retained: MaxCost(a.Retained, b.Retained)}
}

// RepeatResources is valid only for an established invariant body summary.
// Evolving state must instead retain an indexed sum or an unknown bound.
func RepeatResources(body ResourceCost, count int64) ResourceCost {
	if count < 0 {
		return UnknownResourceCost("negative repetition count")
	}
	if count == 0 {
		return ZeroResourceCost()
	}
	n := ConstantCost(count)
	net := MultiplyCost(body.NetLive, n)
	peak := UnknownCost("repeated live change is not an exact constant")
	if delta, ok := body.NetLive.Constant(); ok {
		peak = body.PeakLive
		if delta > 0 {
			peak = AddCost(MultiplyCost(ConstantCost(count-1), body.NetLive), body.PeakLive)
		}
	}
	return ResourceCost{Model: body.Model, Work: MultiplyCost(n, body.Work), Allocated: MultiplyCost(n, body.Allocated), PayloadCopied: MultiplyCost(n, body.PayloadCopied), MetadataCopied: MultiplyCost(n, body.MetadataCopied), IO: MultiplyCost(n, body.IO), NetLive: net, PeakLive: peak, Retained: MaxCost(ConstantCost(0), net)}
}

// EvolvingCallbackWork preserves the state index; it never multiplies the
// initial state's callback cost by a later iteration count without evidence.
func EvolvingCallbackWork(callback string, count CostExpr) CostExpr {
	return CostExpr{Kind: "indexed_sum", Name: callback + "(element_i,state_i)", Args: []CostExpr{normalizeCost(count)}}
}

type ResourcePair struct {
	Left  ResourceRef `json:"left"`
	Right ResourceRef `json:"right"`
}

// AliasRelation returns established equality, contradicted equality (proven
// disjointness), or unresolved possible aliasing. Different names are no proof.
func AliasRelation(a, b ResourceRef, aliases, disjoint []ResourcePair) Status {
	if a.Unknown || b.Unknown || a.ID == "" || b.ID == "" {
		return Unresolved
	}
	reachable := func(start string) map[string]bool {
		seen := map[string]bool{start: true}
		queue := []string{start}
		for len(queue) > 0 {
			item := queue[0]
			queue = queue[1:]
			for _, edge := range aliases {
				for _, pair := range [][2]string{{edge.Left.ID, edge.Right.ID}, {edge.Right.ID, edge.Left.ID}} {
					if pair[0] == item && pair[1] != "" && !seen[pair[1]] {
						seen[pair[1]] = true
						queue = append(queue, pair[1])
					}
				}
			}
		}
		return seen
	}
	left, right := reachable(a.ID), reachable(b.ID)
	if left[b.ID] {
		return Established
	}
	for _, edge := range disjoint {
		if left[edge.Left.ID] && right[edge.Right.ID] || left[edge.Right.ID] && right[edge.Left.ID] {
			return Contradicted
		}
	}
	return Unresolved
}

func sortedStrings(values []string) []string {
	seen := make(map[string]bool)
	for _, value := range values {
		if value != "" {
			seen[value] = true
		}
	}
	result := make([]string, 0, len(seen))
	for value := range seen {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
