package Lyger

import (
	"errors"
	"strings"
	"testing"

	"github.com/l-programming-language/l/lerrors"
)

func TestDictionaryImportLocalCollisions(t *testing.T) {
	const imported = "Use IO's Print from `io` as \"Paint\".\n"
	requireDistributionFrontend(t, imported+"Show `control|`.\n", nil)
	for _, name := range []string{"Paint", "Paints"} {
		local := "\"" + name + "\" is `local|`.\n"
		for _, after := range []bool{false, true} {
			order := "before import"
			source := local + imported
			if after {
				order, source = "after import", imported+local
			}
			t.Run(name+"/"+order, func(t *testing.T) {
				result, err := distributionFrontend(t, source+"Show `unused|`.\n", nil)
				if result != nil || !errors.Is(err, lerrors.DuplicateDeclaration) || !strings.Contains(strings.ToLower(err.Error()), "paint") || !strings.Contains(err.Error(), "main.l:1:") || !strings.Contains(err.Error(), "main.l:2:") {
					t.Fatalf("import/local collision must preserve both sites in either source order: result-present=%t error=%v", result != nil, err)
				}
			})
		}
	}
}
