package Lyger

import (
	"fmt"
	"strings"
	"testing"
)

var functionReuseSource = identityFunction + `    Publicly create "Reuse" as a New Function with Inputs of "Callback" as Identity and Outputs of Identity.
Callback ` + "`first`" + ` as "First".
Define "Local" as Callback.
Local ` + "`second`" + ` as "Second".
Local ` + "`third`" + ` as "Third".
Callback ` + "`fourth`" + ` as "Fourth".
Show First and Second and Third and Fourth.
Return Callback.

Reuse Identity as "Saved".
Saved ` + "`fifth`" + ` as "Fifth".
Saved ` + "`sixth`" + ` as "Sixth".
Show Fifth and Sixth.

    Publicly create "Again" as a New Function with Inputs of "Ready" as Truth and Outputs of Text.
If Ready, Return ` + "`leaf`" + `.
Again true as "First".
Again true as "Second".
Return First.

Again false as "Result".
Show Result.
`

func functionBoxingSource() (string, string) {
	var source, output strings.Builder
	source.WriteString(identityFunction + "Define \"Items\" as a New List of `item`.\n\n" + functionSource("Keep|-s|-ed|-ing", "\"Value\" as Thing", "Thing", "Return Value."))
	for _, tc := range []struct{ name, value, observe, output string }{
		{"Text", "`text`", "Show %s.", "text"},
		{"Number", "7", "Show %s.", "7"},
		{"Truth", "true", "Show %s.", "true"},
		{"Null", "Null", "Show %s.", "null"},
		{"List", "Items", "Show %s.", "[item]"},
		{"Function", "Identity", "Define \"Read%s\" as a New Identity from %[1]s.\nRead%[1]s `callback` as \"Value%[1]s\".\nShow Value%[1]s.", "callback"},
	} {
		fmt.Fprintf(&source, functionSource("Box%s", "", "Thing", "Return %s.")+"Keep %s as \"Input%s\".\nBox%s as \"Output%s\".\n", tc.name, tc.value, tc.value, tc.name, tc.name, tc.name)
		for _, prefix := range []string{"Input", "Output"} {
			fmt.Fprintf(&source, tc.observe+"\n", prefix+tc.name)
			output.WriteString(tc.output)
		}
		source.WriteString("\n")
	}
	return source.String(), output.String()
}

func functionTransferCases() []struct{ name, source, output, diagnostic string } {
	boxing, output := functionBoxingSource()
	const setup = "Define \"boxed\" as a New Thing of 7.\n\n"
	return []struct{ name, source, output, diagnostic string }{
		{"reuse", functionReuseSource, "firstsecondthirdfourthfifthsixthleaf", ""},
		{"boxing", boxing, output, ""},
		{"argument check", setup + functionSource("Keep|-s|-ed|-ing", "\"Value\" as Thing", "", "Show `unreachable`.") + "Create \"Forward|-s|-ed|-ing\" for Keeping any Null.\nForward a New Null from boxed.\n", "", "runtime value type mismatch"},
		{"return check", setup + functionSource("Reject", "", "Thing", "Define \"Nothing\" as a New Null from boxed.\nReturn Nothing.") + "Reject as \"Result\".\nShow Result.\n", "", "runtime value type mismatch"},
	}
}

func TestFunctionTransfersNative(t *testing.T) {
	for _, tc := range functionTransferCases() {
		t.Run(tc.name, func(t *testing.T) {
			runtimeBundleResult(t, runtimeValuesBundle(t, tc.source), tc.output, tc.diagnostic)
		})
	}
}
