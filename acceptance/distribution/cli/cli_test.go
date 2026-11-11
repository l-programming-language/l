//go:build darwin || linux

package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
)

// Cases have independent projects, resource copies, provenance recorders and
// child process groups. Bound admission before starting their existing watchdogs.
var nativeCases = make(nativetest.Admission, 4)

// Each scenario builds its own authored program using the actual installed CLI.
// Cargo evidence is required before a Go executable could satisfy runtime bytes.
func buildProgram(t testing.TB, cli, cwd, source, assets string) string {
	t.Helper()
	recorder := nativetest.CLIProbe(t)
	args := []string{"build"}
	if assets != "" {
		args = append(args, "--assets", assets)
	}
	args = append(args, source)
	out, diag, err := nativetest.BuildPrerequisite.Run(t, cli, cwd, recorder.Env(nil), args...)
	if err != nil {
		t.Fatalf("Rust CLI generation/build prerequisite; runtime not exercised: %v\nstdout=%q\nstderr=%s", err, out, diag)
	}
	absolute := source
	if !filepath.IsAbs(absolute) {
		absolute = filepath.Join(cwd, absolute)
	}
	absolute, err = filepath.EvalSymlinks(absolute)
	if err != nil {
		t.Fatal(err)
	}
	binary := strings.TrimSuffix(absolute, ".l")
	if string(out) != binary+"\n" || len(diag) != 0 {
		t.Fatalf("build streams=%q/%q; want %q/empty", out, diag, binary+"\n")
	}
	recorder.AssertPublished(t, binary)
	return binary
}

func TestDistributionCLI(t *testing.T) {
	const source = "Create \"Hello\" as a New Basic App.\nThe App Prints `Hello World!`.\n"
	for _, tc := range []struct {
		name, source, output string
		files                map[string]string
	}{
		{"embedded core", source, "Hello World!\n", nil},
		{"source relative include", "Use `nested/entry.l`.\n", "source relative|\n", map[string]string{"nested/entry.l": "Use `../helpers.l`.\nCreate \"Hello\" as a New Basic App.\nThe App Prints Greeting.\n", "helpers.l": "\"Greeting\" is `source relative|`.\n"}},
		{"literal text", "Create \"Worker\" as a New Basic App.\nThe App Prints `a, b <DEF> 世界`.\n", "a, b <DEF> 世界\n", nil},
		{"explicit dictionary", "Use LocalTools's Greeting from `./words.l` as \"Greeting\".\nCreate \"Hello\" as a New Basic App.\nThe App Prints Greeting.\n", "local hello|\n", map[string]string{"words.l": "Create \"LocalTools\" as a Dictionary.\nDefine LocalTools's \"Greeting\" as `local hello|`.\n"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := project(t, tc.source, tc.files)
			binary := buildProgram(t, standalone(t), t.TempDir(), filepath.Join(dir, "main.l"), "")
			outcome(t, binary, t.TempDir(), nil, 0, tc.output, "")
		})
	}
	t.Run("unrelated neighboring builtin", func(t *testing.T) {
		const unrelated = "This unrelated file is intentionally not valid L.\n"
		dir := project(t, source, map[string]string{"builtin.l": unrelated})
		binary := buildProgram(t, standalone(t), dir, "main.l", "")
		outcome(t, binary, t.TempDir(), nil, 0, "Hello World!\n", "")
		if read(t, filepath.Join(dir, "builtin.l")) != unrelated {
			t.Fatal("compiler changed unrelated builtin.l")
		}
	})
	t.Run("relocation", func(t *testing.T) {
		dir := project(t, source, nil)
		binary := buildProgram(t, standalone(t), dir, "main.l", "")
		destination := filepath.Join(t.TempDir(), "main")
		if err := os.Rename(binary, destination); err != nil {
			t.Fatal(err)
		}
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
		outcome(t, destination, t.TempDir(), []string{"PATH="}, 0, "Hello World!\n", "")
	})
}
