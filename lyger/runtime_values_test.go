package Lyger

import (
	"testing"
)

func TestRuntimeValuesRejectIncompatibleUpdates(t *testing.T) {
	for _, source := range []string{
		"Define \"items\" as a New List of Number.\nPut `wrong` in items.\n",
		"Define \"bag\" as a New Map of Text.\nPut 1 in bag as `key`.\n",
		"Define \"handle\" as a New Pointer of 1.\nPut true in handle.\n",
		"Define \"items\" as a New List of Number.\nDefine \"selected\" as a New Text from items at 0.\n",
	} {
		result, err := distributionFrontend(t, source+"Create \"Witness\" as a New Basic App.\n", nil)
		if err == nil {
			_, err = planOrchestration(result.Compilation.Semantics)
		}
		if err == nil {
			t.Fatalf("invalid typed use accepted: %s", source)
		}
	}
}

func TestRuntimeTypeRefCanonicalIdentity(t *testing.T) {
	shape, alias, other := &record{}, &record{}, &record{}
	ref := runtimeTypeRef{Kind: valueFunction, Inputs: []runtimeTypeRef{{Kind: valueList, Element: &runtimeTypeRef{Kind: valueFunction, Definition: alias}}}}
	canonical, err := canonicalRuntimeType(ref, map[*record]*record{alias: shape})
	if err != nil || canonical.Inputs[0].Element.Definition != shape || ref.Inputs[0].Element.Definition != alias {
		t.Fatalf("canonical type must resolve nested aliases without mutating its source: %#v, %v", canonical, err)
	}
	otherType, err := canonicalRuntimeType(ref, map[*record]*record{alias: other})
	if err != nil || canonical.equal(otherType) {
		t.Fatal("different Function identities must remain distinct through Function/List types")
	}
}
