package contracts

import "testing"

func TestResourcesBASE055CumulativeVersusPeak(t *testing.T) {
	allocate := ZeroResourceCost()
	allocate.Allocated = ConstantCost(8)
	allocate.NetLive = ConstantCost(8)
	allocate.PeakLive = ConstantCost(8)
	release := ZeroResourceCost()
	release.NetLive = ConstantCost(-8)
	body := SequenceResources(allocate, release)
	result := RepeatResources(body, 100)
	if allocated, _ := result.Allocated.Constant(); allocated != 800 {
		t.Fatal("wrong cumulative allocation", result)
	}
	if peak, _ := result.PeakLive.Constant(); peak != 8 {
		t.Fatal("wrong peak", result)
	}
	if retained, _ := result.Retained.Constant(); retained != 0 {
		t.Fatal("released temporary retained", result)
	}
}

func TestResourcesBASE056RetainedResults(t *testing.T) {
	allocate := ZeroResourceCost()
	allocate.Allocated = ConstantCost(8)
	allocate.NetLive = ConstantCost(8)
	allocate.PeakLive = ConstantCost(8)
	result := RepeatResources(allocate, 100)
	if peak, _ := result.PeakLive.Constant(); peak != 800 {
		t.Fatal("retained peak did not grow", result)
	}
	if retained, _ := result.Retained.Constant(); retained != 800 {
		t.Fatal("retained count lost", result)
	}
}

func TestResourcesBASE057EvolvingCallbackState(t *testing.T) {
	expression := EvolvingCallbackWork("callback", SymbolicCost("n"))
	if expression.Kind != "indexed_sum" || expression.Name != "callback(element_i,state_i)" || expression.Args[0].Name != "n" {
		t.Fatalf("state-indexed relation lost: %+v", expression)
	}
	summary := inferTest(t, Node{Kind: Repeat, CountSymbol: "n", Children: []Node{{Kind: Callback, Callback: "action"}}}, nil)
	if summary.Resources.Work.Kind != "indexed_sum" || len(summary.Unknowns) == 0 {
		t.Fatal("unknown callback assumed zero or fixed initial state")
	}
}

func TestResourcesOverflowAndOmittedCosts(t *testing.T) {
	if AddCost(ConstantCost(9223372036854775807), ConstantCost(1)).Kind != "unknown" {
		t.Fatal("overflow silently wrapped")
	}
	if MultiplyCost(ConstantCost(9223372036854775807), ConstantCost(2)).Kind != "unknown" {
		t.Fatal("multiplication overflow silently wrapped")
	}
	if SequenceResources(ZeroResourceCost(), ResourceCost{}).Work.Kind != "unknown" {
		t.Fatal("omitted costs normalized to zero")
	}
}
