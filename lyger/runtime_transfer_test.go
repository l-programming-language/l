package Lyger

import (
	"fmt"
	"strings"
	"testing"
)

var runtimeTransferCases = []struct{ name, signature, setup, body, observe, output string }{
	{"Bytes", "Bytes", "Define \"InputBytes\" as a New Bytes of `bytes`.", "Return Value.", "Show ResultBytes.", "bytes"},
	{"Thing", "Thing", "Define \"InputThing\" as a New Thing of `thing`.", "Return Value.", "Show ResultThing.", "thing"},
	{"Null", "Null", "Define \"InputNull\" as Null.", "Return Value.", "Show ResultNull.", "null"},
	{"Map", "Map of Text", "Define \"InputMap\" as a New Map of Text.\nPut `original` in InputMap as `key`.", "Put `copy` in Value as `key`.\nReturn Value.", "Show InputMap and ResultMap.", "{key: original}{key: copy}"},
	{"Pointer", "Pointer of Text", "Define \"InputPointer\" as a New Pointer of `original`.", "Put `shared` in Value.\nReturn Value.", "Show InputPointer and ResultPointer.", "sharedshared"},
	{"Instance", "Person Instance", "Define \"InputInstance\" as a New Person Instance.\nPut `original` in InputInstance as `name`.", "Put `copy` in Value as `name`.\nReturn Value.", "Show InputInstance and ResultInstance.", "{name: original}{name: copy}"},
	{"Error", "Error", "Define \"InputError\" as an Error of `failure`.", "Return Value.", "Define \"Errors\" as a New List of Error.\nPut ResultError in Errors.\nShow Errors.", "[failure]"},
}

const transferShape = "\"Person\" is a New Adjective describing any Instance.\nIt describes a Generic Instance with a `name` Text.\n\n"

func runtimeTransferSource(name, signature, setup, body, observe string) string {
	return fmt.Sprintf(functionSource("Keep%s", "\"Value\" as %s", "%s", "%s")+"%s\nKeep%s Input%s as \"Result%s\".\n%s\n\n", name, signature, signature, body, setup, name, name, name, observe)
}

func TestRuntimeTransfersThroughFunctions(t *testing.T) {
	for _, tc := range runtimeTransferCases {
		t.Run(tc.name, func(t *testing.T) {
			requireFunctionGenerated(t, transferShape+runtimeTransferSource(tc.name, tc.signature, tc.setup, tc.body, tc.observe))
		})
	}
}

func TestRuntimeTransfersNative(t *testing.T) {
	var source, output strings.Builder
	source.WriteString(transferShape)
	for _, tc := range runtimeTransferCases {
		source.WriteString(runtimeTransferSource(tc.name, tc.signature, tc.setup, tc.body, tc.observe))
		output.WriteString(tc.output)
	}
	runtimeBundleResult(t, runtimeValuesBundle(t, source.String()), output.String(), "")
}
