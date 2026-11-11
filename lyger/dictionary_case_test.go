package Lyger

import "testing"

func TestDictionaryImportsPreserveCaseInsensitiveBindings(t *testing.T) {
	for _, prefix := range []string{
		"Use catalog's emit from `words.l` as \"Display\".\n",
		"Use cAtAlOg's eMiT from `words.l` as \"Display\".\n",
	} {
		t.Run(prefix, func(t *testing.T) {
			r := requireDistributionFrontend(t, prefix+"Create \"Worker\" as a New Basic App.\nThe App Displays `mixed|`.\n", map[string]string{"words.l": "Use Print as \"Imported\".\nCreate \"Catalog\" as a Dictionary.\nUse Imported as Catalog's \"Emit\".\n"})
			target := distributionCanonical(t, r.Compilation, r.Compilation.Memory["Display"])
			if target.ID != "Print" {
				t.Fatalf("canonical export=%s", target.ID)
			}
		})
	}
}
