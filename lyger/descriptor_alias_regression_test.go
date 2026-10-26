package Lyger

import "testing"

func TestDescriptorAliasesRemainReferences(t *testing.T) {
	for _, tc := range []struct{ name, declaration string }{
		{"Compact", `"Compact" describes a Generic App with a New Main Service using any Process.`},
		{"Marked", `"Marked" describes any Function as ` + "`<DEF>[[\"A\", \"<@Function@>\"]]<DEF>`."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			name, declaration := tc.name, tc.declaration
			r := requireDistributionFrontend(t, declaration+"\nUse "+name+" as \"Selected\".\n", nil)
			if got := distributionCanonical(t, r.Compilation, r.Compilation.Memory["Selected"]); got != r.Compilation.Memory[name] {
				t.Fatalf("descriptor alias changed canonical identity: %+v", got)
			}
		})
	}
}
