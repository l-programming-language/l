//go:build darwin || linux

package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	distributiontest "github.com/l-programming-language/l/internal/testsupport/distribution"
	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
)

// LAY-041–043: installed CLI, away-from-checkout native positive control, then
// missing new resources with usable old-path copies deliberately left behind.
func TestLibraryLayoutInstalledProviderIsolation(t *testing.T) {
	nativeCases.Parallel(t)
	compiler := standalone(t)
	source := "Create \"Worker\" as a New Basic App.\nThe App Prints `layout|`.\n"
	assets := distributiontest.CompleteOverride(t)
	control := project(t, source, nil)
	binary := buildProgram(t, compiler, control, "main.l", assets)
	outcome(t, binary, t.TempDir(), nil, 0, "layout|\n", "")
	for _, tc := range []struct{ current, former string }{
		{"builtins/eng/core.l", "builtins/eng/prelude.l"},
		{"lexicon/eng/grammar.json", "grammar.json"},
	} {
		t.Run(tc.current, func(t *testing.T) {
			selected := filepath.Join(assets, filepath.FromSlash(tc.current))
			original, err := os.ReadFile(selected)
			if err != nil {
				t.Fatal(err)
			}
			write(t, filepath.Join(assets, filepath.FromSlash(tc.former)), string(original))
			if err := os.Remove(selected); err != nil {
				t.Fatal(err)
			}
			dir := project(t, source, nil)
			out, diag, buildErr := nativetest.BuildPrerequisite.Run(t, compiler, dir, nil, "build", "--assets", assets, "main.l")
			if err := os.WriteFile(selected, original, 0600); err != nil {
				t.Fatal(err)
			}
			if buildErr == nil || len(out) != 0 || !strings.Contains(string(diag), tc.current) {
				t.Fatalf("missing selected input accepted/fell back: %v stdout=%q stderr=%s", buildErr, out, diag)
			}
		})
	}
}
