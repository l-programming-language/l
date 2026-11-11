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
	const invocation = "Key: ~(<@Target@>)~"
	if strings.Count(string(core), invocation) != 1 {
		t.Fatal("constructor contract fixture requires the selected Function invocation")
	}
	var words []string
	for _, line := range strings.Split(string(core), "\n") {
		if strings.HasPrefix(line, `Define "Function`) || strings.HasPrefix(line, `Define any Function's`) {
			line = strings.Replace(line, "Function", "Routine", 1)
			words = append(words, strings.ReplaceAll(line, invocation, "Key: ~(/* selected Routine */ <@Target@>)~"))
		}
	}
	return strings.Join(words, "\n") + "\n\n" + identityFunction +
		"    Publicly create \"Respond|-s|-ed|-ing|-er\" as a New Routine with Inputs of \"Value\" as Text and Outputs of Text.\nReturn Value.\n\n"
}

func TestFunctionConstructorContractRejectsStaticSubstitution(t *testing.T) {
	prefix := functionContractSource(t)
	for name, source := range map[string]string{
		"callback": functionSource("Apply|--ies|--ied|-ing|--ier", "\"Replay|-s|-ed|-ing|-er\" as Echo", "Text", "Replay `value` as \"Result\".\nReturn Result.") + "Apply Respond as \"Selected\".\n",
		"return":   "    Publicly create \"Select|-s|-ed|-ing|-or\" as a New Function with No Inputs and Outputs of Echo.\nReturn Respond.\n",
		"List":     "Define \"callbacks\" as a New List of Echo.\nPut Respond in callbacks.\n",
		"Map":      "Define \"callbacks\" as a New Map of Echo.\nPut Respond in callbacks as `key`.\n",
		"Pointer":  "Define \"saved\" as a New Pointer of Echo.\nPut Respond in saved.\n",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := compileFunctionPlan(t, prefix+source); !errors.Is(err, lerrors.TypeMismatch) {
				t.Fatalf("different Function constructor contract accepted: %v", err)
			}
		})
	}
}

var functionContractReads = map[string]string{
	"Thing":   "Define \"saved\" as a New Thing of Respond.\nDefine \"Selected\" as a New Echo from saved.\n",
	"List":    "Define \"saved\" as a New List of a New Thing of Respond.\nDefine \"Selected\" as a New Echo from saved at 0.\n",
	"Map":     "Define \"saved\" as a New Map of Thing.\nPut Respond in saved as `key`.\nDefine \"Selected\" as a New Echo from saved at `key`.\n",
	"Pointer": "Define \"saved\" as a New Pointer of a New Thing of Respond.\nDefine \"Selected\" as a New Echo from saved.\n",
	"result":  functionSource("Select|-s|-ed|-ing|-or", "", "Thing", "Define \"boxed\" as a New Thing of Respond.\nReturn boxed.") + "Select as \"saved\".\nDefine \"Selected\" as a New Echo from saved.\n",
}

func TestFunctionReturnRejectsDifferentConstructorContract(t *testing.T) {
	prefix := functionContractSource(t)
	index := strings.LastIndex(prefix, "Return Value.")
	source := prefix[:index] + "Publicly create \"Finish|-es|-ed|-ing|-er\" for Returning any Text.\n" + prefix[index:] +
		"    Publicly create \"Relay|-s|-ed|-ing|-er\" as a New Function with Inputs of \"Value\" as Text and Outputs of Text.\nFinish Value.\n"
	if _, err := compileFunctionPlan(t, source); !errors.Is(err, lerrors.TypeMismatch) || !strings.Contains(err.Error(), "Return belongs to a different Function constructor contract") {
		t.Fatalf("word expansion substituted its caller's Return definition: %v", err)
	}
}
