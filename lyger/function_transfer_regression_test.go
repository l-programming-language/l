package Lyger

import (
	"fmt"
	"strings"
)

var functionReuseSource = identityFunction + `    Publicly create "Reuse|-s|-d|--ing|-r" as a New Function with Inputs of "Invoke|-s|-d|--ing|--er" as Echo and Outputs of Echo.
Invoke ` + "`first`" + ` as "First".
Define "Repeat|-s|-ed|-ing|-er" as Invoke.
Repeat ` + "`second`" + ` as "Second".
Repeat ` + "`third`" + ` as "Third".
Invoke ` + "`fourth`" + ` as "Fourth".
Show First and Second and Third and Fourth.
Return Invoke.

Reuse Echo as "Replay|-s|-ed|-ing|-er".
Replay ` + "`fifth`" + ` as "Fifth".
Replay ` + "`sixth`" + ` as "Sixth".
Show Fifth and Sixth.

    Publicly create "Recur|-s|-red|-ring|-rer" as a New Function with Inputs of "Ready" as Truth and Outputs of Text.
If Ready, Return ` + "`leaf`" + `.
Recur true as "First".
Recur true as "Second".
Return First.

Recur false as "Result".
Show Result.
`

func functionBoxingSource() (string, string) {
	var source, output strings.Builder
	source.WriteString(identityFunction + "Define \"Items\" as a New List of `item`.\n\n" + functionSource("Keep|-s|---pt|-ing|-er", "\"Value\" as Thing", "Thing", "Return Value."))
	for _, tc := range []struct{ name, value, observe, output string }{
		{"Text", "`text`", "Show %s.", "text"},
		{"Number", "7", "Show %s.", "7"},
		{"Truth", "true", "Show %s.", "true"},
		{"Null", "Null", "Show %s.", "null"},
		{"List", "Items", "Show %s.", "[item]"},
		{"Function", "Echo", "Define \"Read%s|-s|-|-ing|-er\" as a New Echo from %[1]s.\nRead%[1]s `callback` as \"Value%[1]s\".\nShow Value%[1]s.", "callback"},
	} {
		fmt.Fprintf(&source, functionSource("Box%s|-s|-ed|-ing|-er", "", "Thing", "Return %s.")+"Keep %s as \"Input%s\".\nBox%s as \"Output%s\".\n", tc.name, tc.value, tc.value, tc.name, tc.name, tc.name)
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
		{"argument check", setup + functionSource("Keep|-s|---pt|-ing|-er", "\"Value\" as Thing", "", "Show `unreachable`.") + "Create \"Forward|-s|-ed|-ing|-er\" for Keeping any Null.\nForward a New Null from boxed.\n", "", "runtime value type mismatch"},
		{"return check", setup + functionSource("Reject|-s|-ed|-ing|-or", "", "Thing", "Define \"Nothing\" as a New Null from boxed.\nReturn Nothing.") + "Reject as \"Result\".\nShow Result.\n", "", "runtime value type mismatch"},
	}
}
