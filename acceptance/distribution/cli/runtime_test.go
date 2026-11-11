//go:build darwin || linux

package cli

import (
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	distributiontest "github.com/l-programming-language/l/internal/testsupport/distribution"
	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	"github.com/l-programming-language/l/internal/testsupport/phase1"
)

// The installed compiler and resulting binaries run outside the checkout. The
// alternate provider has no default entry and supplies an authored L policy.
func TestReplaceableRuntimeInstalledCLI(t *testing.T) {
	nativeCases.Parallel(t)
	compiler := standalone(t)
	for _, alternate := range []bool{false, true} {
		name, assets := "default", ""
		if alternate {
			name, assets = "alternate", t.TempDir()
			for name, file := range distributiontest.AlternateRuntime(t) {
				write(t, filepath.Join(assets, filepath.FromSlash(name)), string(file.Data))
			}
			regenerateProvider(t, assets)
		}
		t.Run(name, func(t *testing.T) {
			for _, tc := range []struct {
				name, source, stdout, stderr string
				status                       int
			}{
				{"raw output", "Use IO's Print from `io` as \"Print\".\nCreate \"Worker\" as a New Basic App.\nThe App Prints `raw|\n\n雪\t`.\nExecute Worker.\nShow Worker's Output.\n", "raw|\n\n雪\t\n", "", 0},
				{"explicit error", "Exit with an error of `authored|\n\n`.\n", "", "authored|\n\n\n", 1},
			} {
				t.Run(tc.name, func(t *testing.T) {
					project := project(t, tc.source, nil)
					binary := buildProgram(t, compiler, project, "main.l", assets)
					outcome(t, binary, t.TempDir(), nil, tc.status, tc.stdout, tc.stderr)
				})
			}
			t.Run("misuse", func(t *testing.T) {
				project := project(t, "Create \"Worker\" as a New Basic App.\nAwait Worker.\n", nil)
				binary := buildProgram(t, compiler, project, "main.l", assets)
				stdout, stderr, err := nativetest.RuntimeExecution.Run(t, binary, t.TempDir(), nil)
				var exit *exec.ExitError
				if !errors.As(err, &exit) || exit.ExitCode() != 1 || len(stdout) != 0 {
					t.Fatalf("misuse status/output: %v stdout=%q stderr=%q", err, stdout, stderr)
				}
				message := string(stderr)
				if !strings.Contains(message, filepath.Join(project, "main.l")+":2:1:") {
					t.Fatalf("selected runtime lost source attribution: %q", message)
				}
				if alternate {
					if !strings.HasPrefix(message, "Await requires an invocation (source: ") || !strings.HasSuffix(message, ")\n") {
						t.Fatalf("default policy replaced selected L policy: %q", message)
					}
				} else if !strings.HasSuffix(message, ": Await requires an invocation\n") {
					t.Fatalf("bundled Runtime diagnostic changed: %q", message)
				}
			})
			if alternate {
				// Reuse the public process observers, including causal gates and
				// leak checks before fallback cleanup. Default coverage remains
				// in the original lifecycle/continuation/shutdown packages.
				for _, id := range []string{"F05", "F10", "F11", "F12", "F17", "F18"} {
					t.Run(id, func(t *testing.T) { phase1.RunFixtureAssets(t, id, assets) })
				}
			}
		})
	}
}
