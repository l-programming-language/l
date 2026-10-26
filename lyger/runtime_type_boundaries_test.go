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
const runtimeFunctionReadSource = "Define \"callbacks\" as a New List of Identity.\nDefine \"callback\" as a New Identity from callbacks at 0.\nCall callback with `value` as \"Result\".\nShow Result.\n"

func TestRuntimeHandleReads(t *testing.T) {
	for _, tc := range []struct{ name, source, contains string }{
		{"App", distributionIOImport + runtimeAppReadSource, ".expect_type(LType::App)"},
		{"Function", identityFunction + runtimeFunctionReadSource, ".expect_type(LType::Function("},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source, err := functionGenerated(t, tc.source)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(source, tc.contains) {
				t.Fatalf("typed handle read missing %q", tc.contains)
			}
		})
	}
}

func TestRuntimeReadPreservesRequestedIdentity(t *testing.T) {
	otherShape := strings.ReplaceAll(transferShape, "Person", "Other")
	otherFunction := strings.ReplaceAll(strings.ReplaceAll(identityFunction, "Identity", "Other"), "Text", "Number")
	for _, tc := range []struct{ name, source string }{
		{"Instance", transferShape + otherShape + "Define \"ada\" as a New Person Instance.\nDefine \"people\" as a New List of ada.\nDefine \"selected\" as a New Other Instance from people at 0.\n"},
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
