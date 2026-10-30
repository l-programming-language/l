package Lyger

import (
	"path/filepath"
	"testing"
)

func TestFormatterImportsRetainCanonicalDefinition(t *testing.T) {
	imports := []string{
		"Use Formatting's Format from `format` as \"Direct\".\n",
		"Use Runtime from `runtime`.\nUse Runtime's Format as \"Shared\".\n",
	}
	for _, order := range [][2]int{{0, 1}, {1, 0}} {
		t.Run([]string{"format-first", "runtime-first"}[order[0]], func(t *testing.T) {
			result, err := NewLygerService(Options{}).Compile([]byte(imports[order[0]]+imports[order[1]]), filepath.Join(t.TempDir(), "main.l"))
			if err != nil {
				t.Fatal(err)
			}
			c := result.Compilation
			direct := distributionCanonical(t, c, c.Memory["Direct"])
			shared := distributionCanonical(t, c, c.Memory["Shared"])
			if direct == nil || direct != shared || direct.ID != "Formatting.Format" {
				t.Fatal("Runtime reexport replaced the canonical formatter")
			}
			function := direct.function
			if function == nil || function.Header.source.Unit.Dictionary != "Formatting" || function.Constructor.Name != "format_number" {
				t.Fatal("formatter lost its source owner or native declaration")
			}
			origins := 0
			for _, origin := range c.SourceOrigins {
				if origin.Name == "builtins/runtime/eng/number.l" {
					origins++
				}
			}
			if origins != 1 {
				t.Fatalf("formatter source expanded %d times, want one", origins)
			}
			owner, err := c.exportedRecord("Runtime")
			if err != nil {
				t.Fatal(err)
			}
			exports, err := c.publicExports(owner, make(map[*record]bool))
			if err != nil {
				t.Fatal(err)
			}
			matches := 0
			for _, exported := range exports {
				if distributionCanonical(t, c, exported) == direct {
					matches++
				}
			}
			if matches != 1 {
				t.Fatalf("runtime library exports the canonical formatter %d times, want one", matches)
			}
			if _, err := c.ComposeLibrary("Runtime", "runtime"); err != nil {
				t.Fatalf("complete runtime reexport composition: %v", err)
			}
		})
	}
}

func TestIOImportsOnlyFormatterPolicy(t *testing.T) {
	result, err := NewLygerService(Options{}).Compile([]byte("Use IO's Print from `io` as \"Print\".\n"), filepath.Join(t.TempDir(), "main.l"))
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, unit := range result.Compilation.SourceUnits {
		if unit.Dictionary == "Runtime" {
			t.Fatal("IO loaded process supervision policy")
		}
		found = found || unit.Dictionary == "Formatting"
	}
	if !found {
		t.Fatal("IO bypassed its selected formatter dictionary")
	}
}
