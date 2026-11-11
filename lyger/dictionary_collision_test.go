package Lyger

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/l-programming-language/l/lerrors"
)

func TestDictionaryImportLocalCollisions(t *testing.T) {
	const imported = "Use Catalog's Emit from `catalog` as \"Paint\".\n"
	files := dictionaryIndexFilesystem(t)
	installTestDictionary(t, files, "catalog", "Catalog", "Create \"Catalog\" as a Dictionary.\nCreate Catalog's \"Emit|-s|-ted|-ting|-ter\" for Printing any Text.\n")
	service := NewLygerService(Options{Resources: files})
	main := filepath.Join(t.TempDir(), "main.l")
	if _, err := service.Compile([]byte(imported+"Show `control|`.\n"), main); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"Paint", "Paints"} {
		local := "Create \"" + name + "|-s|-ed|-ing|-er\" for Showing any Text.\n"
		for _, after := range []bool{false, true} {
			order := "before import"
			source := local + imported
			if after {
				order, source = "after import", imported+local
			}
			t.Run(name+"/"+order, func(t *testing.T) {
				result, err := service.Compile([]byte(source+"Show `unused|`.\n"), main)
				if result != nil || !errors.Is(err, lerrors.DuplicateDeclaration) || !strings.Contains(strings.ToLower(err.Error()), "paint") || !strings.Contains(err.Error(), "main.l:1:") || !strings.Contains(err.Error(), "main.l:2:") {
					t.Fatalf("import/local collision must preserve both sites in either source order: result-present=%t error=%v", result != nil, err)
				}
			})
		}
	}
}
