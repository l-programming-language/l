package Lyger

import (
	"errors"
	"strings"
	"testing"

	"github.com/l-programming-language/l/lerrors"
)

func TestRecordRuntimeTypeRejectsMalformedStorage(t *testing.T) {
	for _, rec := range []*record{nil, {}, {Value: &entry{Type: storagePrimitive}}, {Value: initEntry("Text", storagePrimitive), runtimeType: &runtimeTypeRef{Kind: valueNumber}}} {
		if _, err := recordRuntimeType(rec); !errors.Is(err, lerrors.Internal) {
			t.Fatalf("invalid runtime record accepted or cause lost: %#v: %v", rec, err)
		}
	}
	rec := &record{Value: initEntry("Pointer", storagePrimitive), runtimeType: &runtimeTypeRef{Kind: valuePointer, Element: &runtimeTypeRef{Kind: valueText}}}
	typ, err := recordRuntimeType(rec)
	if err != nil || !typ.equal(*rec.runtimeType) {
		t.Fatalf("complete checked runtime type lost: %#v, %v", typ, err)
	}
}

const runtimeAppReadSource = "Create \"Worker\" as a New Basic App.\nThe App Prints `child`.\nDefine \"apps\" as a New List of App.\nPut Worker in apps.\nDefine \"selected\" as a New App from apps at 0.\nExecute selected.\nShow selected's Output.\n"
const runtimeFunctionReadSource = "Define \"callbacks\" as a New List of Identity.\nDefine \"callback\" as a New Identity from callbacks at 0.\ncallback `value` as \"Result\".\nShow Result.\n"

func TestRuntimeReadPreservesRequestedIdentity(t *testing.T) {
	otherFunction := strings.ReplaceAll(strings.ReplaceAll(identityFunction, "Identity", "Other"), "Text", "Number")
	for _, tc := range []struct{ name, source string }{
		{"Function", identityFunction + otherFunction + "Define \"callbacks\" as a New List of Identity.\nDefine \"selected\" as a New Other from callbacks at 0.\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := functionGenerated(t, tc.source)
			if !errors.Is(err, lerrors.TypeMismatch) || !strings.Contains(err.Error(), "read type does not match") {
				t.Fatalf("requested type must be checked against stored identity/signature: %v", err)
			}
		})
	}
}

func TestRuntimeHandleReadsNative(t *testing.T) {
	c := compileMemberResolutionSource(t, memberResolutionSource{main: distributionIOImport + "\n" + identityFunction + runtimeAppReadSource + runtimeFunctionReadSource})
	captureNativeResult(t, c, "child\nvalue")
}

func TestUndeclaredRuntimeTypeHasNoImplicitImplementation(t *testing.T) {
	for name, source := range map[string]string{
		"construction": "Define \"value\" as a New Instance.\n",
		"descriptor":   "\"Person\" is a New Adjective describing any Instance.\nIt describes a Generic Instance with a `name` Text.\nDefine \"value\" as a New Person Instance.\n",
		"input":        functionSource("Keep", "\"Value\" as Instance", "", "Show Value."),
		"output":       functionSource("Keep", "", "Instance", "Return Null."),
		"List":         "Define \"values\" as a New List of Instance.\n",
		"Map":          "Define \"values\" as a New Map of Instance.\n",
		"Pointer":      "Define \"value\" as a New Pointer of Instance.\n",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := functionGenerated(t, source); err == nil {
				t.Fatal("undeclared runtime type acquired an implicit implementation")
			}
		})
	}
	if _, err := functionGenerated(t, "Define \"Instance\" as 7.\nShow Instance.\n"); err != nil {
		t.Fatalf("ordinary declaration must not be blocked by a reserved-name blacklist: %v", err)
	}
}
