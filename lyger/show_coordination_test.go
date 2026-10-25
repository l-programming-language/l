package Lyger

import (
	"reflect"
	"testing"
)

func TestShowCoordinatedValuesRetainOrder(t *testing.T) {
	result := requireDistributionFrontend(t, "Define \"First\" as `first|`.\nDefine \"Second\" as `second|`.\nDefine \"Third\" as `third|`.\nShow First and Second and Third.\n", nil)
	if got := phase1Kinds(result.Compilation.Semantics.Instructions); !reflect.DeepEqual(got, []string{"show", "show", "show"}) {
		t.Fatalf("coordinated effects=%v", got)
	}
	instructions := result.Compilation.Semantics.Instructions
	for i, name := range []string{"First", "Second", "Third"} {
		if instructions[i].Value.Record != result.Compilation.Memory[name] {
			t.Errorf("operand %d does not use %s", i, name)
		}
	}
}
