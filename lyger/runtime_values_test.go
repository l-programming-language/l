package Lyger

import (
	"errors"
	"os/exec"
	"strings"
	"testing"

	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	Logiceme "github.com/l-programming-language/l/logiceme"
)

var runtimeValueSources = []struct{ name, source, want string }{
	{"list snapshot", "Create \"items\" as a New List of 1 and 2.\n\"copy\" is items.\nPut 3 in items.\nPlace 4 in items at 0.\nDefine \"selected\" as a New Number from copy at 0.\nShow selected.\n", "LList::new"},
	{"map", "Define \"bag\" as a New Map of Text.\nPut `value` in bag as `key`.\nDefine \"selected\" as a New Text from bag at `key`.\nShow selected.\n", "LMap::new"},
	{"pointer sharing", "Define \"message\" as `first`.\nDefine \"handle\" as a New Pointer of message.\nDefine \"shared\" as handle.\nPut `second` in shared.\nDefine \"selected\" as a New Text from handle.\nShow selected.\n", "LPointer::typed"},
	{"bytes and null", "Define \"buffer\" as a New Bytes of `raw`.\nDefine \"selected\" as a New Number from buffer at 1.\nDefine \"nothing\" as Null.\nShow buffer and selected and nothing.\n", "l_index"},
	{"instance", "\"Person\" is a New Adjective describing any Instance.\nIt describes a Generic Instance with a `name` Text and a `age` Number.\nDefine \"ada\" as a New Person Instance.\nPut `Ada` in ada as `name`.\nPut 37 in ada as `age`.\nDefine \"selected\" as a New Text from ada at `name`.\nShow selected.\n", "LInstance::new"},
	{"Thing extraction", "Define \"boxed\" as a New Thing of `first`.\nDefine \"selected\" as a New Text from boxed.\nShow selected.\n", "into_text"},
	{"nested mixed list", "Define \"numbers\" as a New List of 1.\nDefine \"words\" as a New List of `x`.\nDefine \"mixed\" as a New List of numbers and words.\nShow mixed.\n", "LType::Thing"},
	{"typed extraction", "Define \"items\" as a New List of 1.\nDefine \"boxed\" as a New Thing of items.\nDefine \"copied\" as a New List of Number from boxed.\nShow copied.\n", ".expect_type(LType::List"},
	{"error and app containers", "Define \"problem\" as an Error of `failure`.\nDefine \"faultLog\" as a New List of Error.\nPut problem in faultLog.\nCreate \"Worker\" as a New Basic App.\nDefine \"apps\" as a New List of App.\nPut Worker in apps.\n", "LValue::App"},
}

func TestRuntimeValuesSourceGeneration(t *testing.T) {
	for _, tc := range runtimeValueSources {
		t.Run(tc.name, func(t *testing.T) {
			bundle := runtimeValuesBundle(t, tc.source)
			var source strings.Builder
			for _, book := range bundle.ByteArrays {
				for _, file := range book {
					source.Write(file)
				}
			}
			if !strings.Contains(source.String(), tc.want) || !strings.Contains(source.String(), "struct LEnvironment") {
				t.Fatal("generated Rust lost runtime construction or support")
			}
		})
	}
}

func TestRuntimeValuesRejectIncompatibleUpdates(t *testing.T) {
	for _, source := range []string{
		"Define \"items\" as a New List of Number.\nPut `wrong` in items.\n",
		"Define \"bag\" as a New Map of Text.\nPut 1 in bag as `key`.\n",
		"Define \"handle\" as a New Pointer of 1.\nPut true in handle.\n",
		"Define \"items\" as a New List of Number.\nDefine \"selected\" as a New Text from items at 0.\n",
	} {
		result, err := distributionFrontend(t, source+"Create \"Witness\" as a New Basic App.\n", nil)
		if err == nil {
			_, err = result.OrchestratorUnit("entry", executableTestCompanions(result))
		}
		if err == nil {
			t.Fatalf("invalid typed use accepted: %s", source)
		}
	}
}

func runtimeValuesBundle(t *testing.T, source string) Logiceme.Executable {
	t.Helper()
	result := requireDistributionFrontend(t, source+"\nCreate \"Witness\" as a New Basic App.\n", nil)
	unit, err := result.OrchestratorUnit("entry", executableTestCompanions(result))
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := Logiceme.NewLogicemeService(phase1Options(t).Resources).Evaluate(unit.BackendAML, Logiceme.GenerationOptions{IncludeRuntime: true})
	if err != nil {
		t.Fatal(err)
	}
	return bundle
}

var nullRuntimeCases = []struct{ name, source, stdout, diagnostic string }{
	{"null-cast", "Define \"boxed\" as a New Thing of 7.\nShow a New Null from boxed.\n", "", "runtime value type mismatch"},
	{"null-list-bound", "Define \"items\" as a New List of Null.\nShow a New Null from items at 0.\n", "", "List index out of bounds"},
	{"null-map-key", "Define \"items\" as a New Map of Null.\nShow a New Null from items at `missing`.\n", "", "Map key is absent"},
	{"null-container-cast", "Define \"boxed\" as a New Thing of 7.\nDefine \"items\" as a New List of a New Null from boxed.\nShow items.\n", "", "runtime value type mismatch"},
	{"null-shared-results", "    Publicly create \"Publish\" as a New Function with Inputs of \"Value\" as Text and No Outputs.\nPublicly define \"dummy\" as Value.\n\n    Publicly create \"Pair\" as a New Function with No Inputs and Outputs of Null and Text.\nReturn Null and `text`.\n\nPair as \"Nothing\" and \"Message\".\nShow Nothing and Message.\n", "nulltext", ""},
}

func TestNullBoxingPreservesReads(t *testing.T) {
	for _, tc := range nullRuntimeCases {
		t.Run(tc.name, func(t *testing.T) {
			source, err := functionGenerated(t, tc.source)
			if err != nil {
				t.Fatal(err)
			}
			if tc.diagnostic != "" && strings.Count(source, ".expect_type(LType::Null)") != 1 {
				t.Fatal("Null boxing lost or duplicated the checked read")
			}
			if tc.diagnostic == "" {
				for _, field := range []string{"lCalled0Result.0", "lCalled0Result.1"} {
					if strings.Count(source, field) != 1 {
						t.Fatalf("typed Null/Text result field %s must be consumed once", field)
					}
				}
				if strings.Contains(source, "lResults.next()") {
					t.Fatal("typed results must not use an erased value iterator")
				}
			}
		})
	}
}

func TestRuntimeValuesNative(t *testing.T) {
	var source strings.Builder
	for _, tc := range runtimeValueSources {
		source.WriteString("\n    " + tc.source)
	}
	cases := append([]struct{ name, source, stdout, diagnostic string }{
		{"snapshot-and-sharing", source.String(), "1valuesecondraw97nullAdafirst[[1], [x]][1]", ""},
		{"nested-type-mismatch", "Define \"items\" as a New List of 1.\nDefine \"boxed\" as a New Thing of items.\nDefine \"copied\" as a New List of Text from boxed.\n", "", "runtime value type mismatch"},
		{"cycle-through-function", "Define \"boxed\" as a New Thing of Null.\nDefine \"link\" as a New Pointer of boxed.\n\n    Publicly create \"ReadShared\" as a New Function with No Inputs and Outputs of Thing.\nDefine \"snapshot\" as a New Thing from link.\nReturn snapshot.\n\nPut ReadShared in link.\n", "", "Pointer write would create an ownership cycle"},
		{"cycle-through-container", "Define \"boxed\" as a New Thing of Null.\nDefine \"handle\" as a New Pointer of boxed.\nDefine \"links\" as a New List of handle.\nPut links in handle.\n", "", "Pointer write would create an ownership cycle"},
	}, nullRuntimeCases...)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runtimeBundleResult(t, runtimeValuesBundle(t, tc.source), tc.stdout, tc.diagnostic)
		})
	}
}

func runtimeBundleResult(t *testing.T, bundle Logiceme.Executable, expected, diagnostic string) {
	t.Helper()
	recordBundle(t, bundle)
	binary := buildActualBundle(t, bundle, "entry", true, "")
	stdout, stderr, err := nativetest.RuntimeExecution.Run(t, binary, "", nil)
	var exit *exec.ExitError
	if diagnostic != "" && (!errors.As(err, &exit) || exit.ExitCode() != 1) {
		t.Fatalf("runtime diagnostic requires exit 1, got %v", err)
	}
	if string(stdout) != expected || (err == nil) != (diagnostic == "") || diagnostic == "" && len(stderr) > 0 || diagnostic != "" && !strings.Contains(string(stderr), diagnostic) {
		t.Fatalf("runtime values: stdout=%q stderr=%q error=%v; expected %q", stdout, stderr, err, expected)
	}
}

func TestRuntimeTypeRefCanonicalIdentity(t *testing.T) {
	shape, alias, other := &record{}, &record{}, &record{}
	ref := runtimeTypeRef{Kind: valueFunction, Inputs: []runtimeTypeRef{{Kind: valueList, Element: &runtimeTypeRef{Kind: valueInstance, Identity: alias}}}}
	canonical, err := canonicalRuntimeType(ref, map[*record]*record{alias: shape})
	if err != nil || canonical.Inputs[0].Element.Identity != shape || ref.Inputs[0].Element.Identity != alias {
		t.Fatalf("canonical type must resolve nested aliases without mutating its source: %#v, %v", canonical, err)
	}
	otherType, err := canonicalRuntimeType(ref, map[*record]*record{alias: other})
	if err != nil || canonical.equal(otherType) {
		t.Fatal("different Instance identities must remain distinct through Function/List types")
	}
}
