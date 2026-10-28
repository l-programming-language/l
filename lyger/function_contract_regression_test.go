package Lyger

import (
	"errors"
	"io/fs"
	"strings"
	"testing"

	"github.com/l-programming-language/l/lerrors"
)

// A distinct source constructor has the same Rust signature but owns different
// executable compositions. Its marker must survive every accepted transfer.
func functionContractSource(t *testing.T) string {
	t.Helper()
	core, err := fs.ReadFile(canonicalTestOptions().Resources, "builtins/eng/core.l")
	if err != nil {
		t.Fatal(err)
	}
	var words []string
	for _, line := range strings.Split(string(core), "\n") {
		if strings.HasPrefix(line, `Define "Function`) || strings.HasPrefix(line, `Define any Function's`) {
			line = strings.Replace(line, "Function", "Routine", 1)
			words = append(words, strings.ReplaceAll(line, "Value: ~<@Target@>~", "Value: ~/* selected Routine */ <@Target@>~"))
		}
	}
	return strings.Join(words, "\n") + "\n\n" + identityFunction +
		"    Publicly create \"Other\" as a New Routine with Inputs of \"Value\" as Text and Outputs of Text.\nReturn Value.\n\n"
}

func TestFunctionConstructorContractRejectsStaticSubstitution(t *testing.T) {
	prefix := functionContractSource(t)
	for name, source := range map[string]string{
		"callback": functionSource("Apply", "\"Callback\" as Identity", "Text", "Callback `value` as \"Result\".\nReturn Result.") + "Apply Other as \"Selected\".\n",
		"return":   "    Publicly create \"Select\" as a New Function with No Inputs and Outputs of Identity.\nReturn Other.\n",
		"List":     "Define \"callbacks\" as a New List of Identity.\nPut Other in callbacks.\n",
		"Map":      "Define \"callbacks\" as a New Map of Identity.\nPut Other in callbacks as `key`.\n",
		"Pointer":  "Define \"saved\" as a New Pointer of Identity.\nPut Other in saved.\n",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := functionGenerated(t, prefix+source); !errors.Is(err, lerrors.TypeMismatch) {
				t.Fatalf("different Function constructor contract accepted: %v", err)
			}
		})
	}
}

var functionContractReads = map[string]string{
	"Thing":   "Define \"saved\" as a New Thing of Other.\nDefine \"Selected\" as a New Identity from saved.\n",
	"List":    "Define \"saved\" as a New List of a New Thing of Other.\nDefine \"Selected\" as a New Identity from saved at 0.\n",
	"Map":     "Define \"saved\" as a New Map of Thing.\nPut Other in saved as `key`.\nDefine \"Selected\" as a New Identity from saved at `key`.\n",
	"Pointer": "Define \"saved\" as a New Pointer of a New Thing of Other.\nDefine \"Selected\" as a New Identity from saved.\n",
	"result":  functionSource("Select", "", "Thing", "Define \"boxed\" as a New Thing of Other.\nReturn boxed.") + "Select as \"saved\".\nDefine \"Selected\" as a New Identity from saved.\n",
}

func TestFunctionConstructorContractRuntimeReads(t *testing.T) {
	prefix := functionContractSource(t)
	for name, source := range functionContractReads {
		t.Run(name, func(t *testing.T) {
			runtimeBundleResult(t, runtimeValuesBundle(t, prefix+source+"Selected `bad` as \"Result\".\n"), "", "runtime value type mismatch")
		})
	}
}

func TestFunctionConstructorAliasesPreserveComposition(t *testing.T) {
	source := functionContractSource(t) + "Use Other as \"Signature\".\n\n" +
		functionSource("Apply", "\"Callback\" as Signature", "Signature", "Callback `first` as \"Result\".\nShow Result.\nReturn Callback.") +
		"Apply Other as \"Saved\".\nSaved `second` as \"Result\".\nShow Result.\n"
	generated, err := functionGenerated(t, source)
	if err != nil || strings.Count(generated, "/* selected Routine */") != 2 {
		t.Fatalf("same canonical contract lost its invocation composition: %v", err)
	}
	runtimeBundleResult(t, runtimeValuesBundle(t, source), "firstsecond", "")
}

func TestFunctionReturnRejectsDifferentConstructorContract(t *testing.T) {
	prefix := functionContractSource(t)
	index := strings.LastIndex(prefix, "Return Value.")
	source := prefix[:index] + "Publicly create \"Finish|-es|-ed|-ing\" for Returning any Text.\n" + prefix[index:] +
		"    Publicly create \"Relay\" as a New Function with Inputs of \"Value\" as Text and Outputs of Text.\nFinish Value.\n"
	if _, err := functionGenerated(t, source); !errors.Is(err, lerrors.TypeMismatch) || !strings.Contains(err.Error(), "Return belongs to a different Function constructor contract") {
		t.Fatalf("word expansion substituted its caller's Return definition: %v", err)
	}
}
