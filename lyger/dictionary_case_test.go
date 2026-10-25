package Lyger

import "testing"

func TestDictionaryImportsPreserveCaseInsensitiveBindings(t *testing.T) {
	for _, prefix := range []string{
		"Use io's print from `io` as \"Display\".\n",
		"Use iO's pRiNt from `io` as \"Display\".\n",
	} {
		t.Run(prefix, func(t *testing.T) {
			r := requireDistributionFrontend(t, prefix+"Create \"Worker\" as a New Basic App.\nThe App Displays `mixed|`.\nExecute Worker.\n", nil)
			target := distributionCanonical(t, r.Compilation, r.Compilation.Memory["Display"])
			if target.ID != "IO.Print" {
				t.Fatalf("canonical export=%s", target.ID)
			}
		})
	}
}
