//go:build darwin || linux

package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	distributiontest "github.com/l-programming-language/l/internal/testsupport/distribution"
	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	Lyger "github.com/l-programming-language/l/lyger"
)

// The test deliberately edits a selected provider. Regeneration is an explicit
// fixture operation through the real frontend, never a CLI loading side effect.
func regenerateProvider(t *testing.T, root string) {
	t.Helper()
	aml, manifest, err := Lyger.ComposeRuntimeResources(os.DirFS(root))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(root, "runtime/generated/support.aml"), string(aml))
	write(t, filepath.Join(root, "runtime/generated/support.json"), string(raw))
}

// LAY-041–043: installed CLI, away-from-checkout native positive control, then
// missing new resources with usable old-path copies deliberately left behind.
func TestLibraryLayoutInstalledProviderIsolation(t *testing.T) {
	nativeCases.Parallel(t)
	compiler := standalone(t)
	source := "Use IO's Print from `io` as \"Print\".\nCreate \"Worker\" as a New Basic App.\nThe App Prints `layout|`.\nExecute Worker.\nShow Worker's Output.\n"
	assets := distributiontest.CompleteOverride(t)
	control := project(t, source, nil)
	binary := buildProgram(t, compiler, control, "main.l", assets)
	outcome(t, binary, t.TempDir(), nil, 0, "layout|\n", "")
	for _, tc := range []struct{ current, former string }{
		{"dictionaries.json", "stdlib/index.json"},
		{"builtins/eng/prelude.l", "builtin.l"},
		{"builtins/eng/storage/dictionary.l", "stdlib/eng/storage/dictionary.l"},
		{"runtime/generated/support.aml", "builtins/runtime/support.aml"},
		{"runtime/rust/host.rs", "builtins/templates/runtime/host.lu"},
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
