package Lyger

import "testing"

func TestDictionaryImportsPreserveCaseInsensitiveBindings(t *testing.T) {
	for _, prefix := range []string{
		"Use catalog's emit from `words.l` as \"Display|-s|-ed|-ing|-er\".\n",
		"Use cAtAlOg's eMiT from `words.l` as \"Display|-s|-ed|-ing|-er\".\n",
	} {
		t.Run(prefix, func(t *testing.T) {
			r := requireDistributionFrontend(t, prefix+"Create \"Worker\" as a New Basic App.\nThe App Displays `mixed|`.\n", map[string]string{"words.l": "Use Print as \"Relay|-s|-ed|-ing|-er\".\nCreate \"Catalog\" as a Dictionary.\nUse Relay as Catalog's \"Emit|-s|-ted|-ting|-ter\".\n"})
			target := distributionCanonical(t, r.Compilation, r.Compilation.Memory["Display"])
			if target.ID != "Print" {
				t.Fatalf("canonical export=%s", target.ID)
			}
		})
	}
}
