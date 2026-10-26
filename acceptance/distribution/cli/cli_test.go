//go:build darwin || linux

package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	distributiontest "github.com/l-programming-language/l/internal/testsupport/distribution"
	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	"github.com/l-programming-language/l/internal/testsupport/phase1"
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
	t.Parallel()
	root := distributiontest.RepoRoot(t)
	example := read(t, filepath.Join(root, "demo/basic/main.l"))
	t.Run("FS01_IO01", func(t *testing.T) {
		nativeCases.Parallel(t)
		dir := project(t, example, nil)
		binary := buildProgram(t, standalone(t), dir, "main.l", "")
		outcome(t, binary, t.TempDir(), nil, 0, "Hello World!\n", "")
	})
	t.Run("FS02", func(t *testing.T) {
		nativeCases.Parallel(t)
		cli := standalone(t)
		link := filepath.Join(t.TempDir(), "symlink-l")
		if err := os.Symlink(cli, link); err != nil {
			t.Fatal(err)
		}
		dir := project(t, "Use `nested/entry.l`.\n", map[string]string{"nested/entry.l": "Use `../helpers.l`.\nDisplay Greeting.\n", "helpers.l": "\"Greeting\" is `source relative|`.\nCreate \"Display|-s|-ed|-ing|-er\" for Showing any Text.\n"})
		binary := buildProgram(t, link, t.TempDir(), filepath.Join(dir, "main.l"), "")
		outcome(t, binary, t.TempDir(), nil, 0, "source relative|", "")
	})
	t.Run("FS03", func(t *testing.T) {
		nativeCases.Parallel(t)
		dir := project(t, example, nil)
		recorder := nativetest.CLIProbe(t)
		out, diag, err := nativetest.BuildPrerequisite.Run(t, "go", root, recorder.Env(nil), "run", "-mod=readonly", "./cmd/l", "build", filepath.Join(dir, "main.l"))
		if err != nil {
			t.Fatalf("go-run embedded-assets prerequisite; runtime not exercised: %v\nstdout=%q\nstderr=%s", err, out, diag)
		}
		if string(out) != filepath.Join(dir, "main")+"\n" || len(diag) != 0 {
			t.Fatalf("go run streams=%q/%q", out, diag)
		}
		recorder.AssertPublished(t, filepath.Join(dir, "main"))
		outcome(t, filepath.Join(dir, "main"), t.TempDir(), nil, 0, "Hello World!\n", "")
	})
	t.Run("FS04", func(t *testing.T) {
		t.Parallel()
		cli := standalone(t)
		for _, invalid := range []bool{false, true} {
			invalid := invalid
			t.Run(fmt.Sprint(invalid), func(t *testing.T) {
				nativeCases.Parallel(t)
				files := map[string]string{}
				if invalid {
					files["builtin.l"] = "This unrelated file is intentionally not valid L.\n"
				}
				dir := project(t, example, files)
				binary := buildProgram(t, cli, dir, "main.l", "")
				outcome(t, binary, t.TempDir(), nil, 0, "Hello World!\n", "")
				if invalid && read(t, filepath.Join(dir, "builtin.l")) != files["builtin.l"] {
					t.Fatal("compiler changed unrelated builtin.l")
				}
			})
		}
	})
	t.Run("FS10", func(t *testing.T) {
		t.Parallel()
		// Each missing-resource case has an independent valid-bundle control. No
		// fallback assertion can pass on an unrelated builtin or syntax failure.
		for _, resource := range []string{"lexicon/eng/grammar.json", "lexicon/eng/vocabulary.json", "builtins/eng/core.l", "stdlib/eng/io/dictionary.l", "stdlib/index.json", "builtins/segments/vowels/execution.lu", "builtins/templates/config/config.lu", "builtins/templates/runtime/runtime.lu"} {
			resource := resource
			t.Run(resource, func(t *testing.T) {
				nativeCases.Parallel(t)
				assets := distributiontest.CompleteOverride(t)
				cli := phase1.CLI(t)
				control := project(t, example, nil)
				buildProgram(t, cli, control, "main.l", assets)
				if err := os.Remove(filepath.Join(assets, filepath.FromSlash(resource))); err != nil {
					t.Fatal(err)
				}
				dir := project(t, example, nil)
				out, diag, err := nativetest.BuildPrerequisite.Run(t, cli, dir, nil, "build", "--assets", assets, "main.l")
				if err == nil || len(out) != 0 || !strings.Contains(string(diag), filepath.Base(resource)) {
					t.Fatalf("missing selected resource was hidden or wrong failure: %v stdout=%q stderr=%s", err, out, diag)
				}
				if _, err := os.Stat(filepath.Join(dir, "main")); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("failed resource build published output: %v", err)
				}
			})
		}
	})
	t.Run("FS11", func(t *testing.T) {
		nativeCases.Parallel(t)
		cli := standalone(t)
		first := project(t, example, nil)
		binary := buildProgram(t, cli, first, "main.l", "")
		outcome(t, binary, t.TempDir(), nil, 0, "Hello World!\n", "")
		assets := distributiontest.CompleteOverride(t)
		second := project(t, example, nil)
		binary = buildProgram(t, cli, second, "main.l", assets)
		outcome(t, binary, t.TempDir(), nil, 0, "Hello World!\n", "")
	})
	t.Run("FS16", func(t *testing.T) {
		nativeCases.Parallel(t)
		dir := project(t, example, nil)
		binary := buildProgram(t, standalone(t), dir, "main.l", "")
		destination := filepath.Join(t.TempDir(), "main")
		if err := os.Rename(binary, destination); err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(binary+".apps", destination+".apps"); err != nil {
			t.Fatal(err)
		}
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
		outcome(t, destination, t.TempDir(), []string{"PATH="}, 0, "Hello World!\n", "")
	})
	t.Run("IO03", func(t *testing.T) {
		nativeCases.Parallel(t)
		source := "Use IO's Print from `io` as \"Print\".\nCreate \"Worker\" as a New Basic App.\nThe App Prints `a, b <DEF> 世界`.\nThe App Prints 42.\nExecute Worker.\nShow `raw|`.\nShow Worker's Output.\nShow `|end`.\n"
		dir := project(t, source, nil)
		binary := buildProgram(t, phase1.CLI(t), dir, "main.l", "")
		outcome(t, binary, t.TempDir(), nil, 0, "raw|a, b <DEF> 世界\n42\n|end", "")
	})
	t.Run("DI01", func(t *testing.T) {
		t.Parallel()
		// Deliberately malformed candidates must not shadow the selected io bundle.
		for _, neighbor := range []string{"io", "io.l"} {
			neighbor := neighbor
			t.Run(neighbor, func(t *testing.T) {
				nativeCases.Parallel(t)
				dir := project(t, example, map[string]string{neighbor: "not a valid dictionary!"})
				binary := buildProgram(t, phase1.CLI(t), dir, "main.l", "")
				outcome(t, binary, t.TempDir(), nil, 0, "Hello World!\n", "")
			})
		}
		t.Run("explicit file", func(t *testing.T) {
			nativeCases.Parallel(t)
			dir := project(t, "Use LocalIO's Greeting from `./io.l` as \"Greeting\".\nShow Greeting.\n", map[string]string{"io.l": "Create \"LocalIO\" as a Dictionary.\nDefine LocalIO's \"Greeting\" as `local io|`.\n"})
			binary := buildProgram(t, phase1.CLI(t), dir, "main.l", "")
			outcome(t, binary, t.TempDir(), nil, 0, "local io|", "")
		})
	})
	t.Run("DI10_source_inclusion", func(t *testing.T) {
		nativeCases.Parallel(t)
		dir := project(t, "Use `effects.l`.\nShow `after`.\n", map[string]string{"effects.l": "Show `before`.\n"})
		binary := buildProgram(t, phase1.CLI(t), dir, "main.l", "")
		outcome(t, binary, t.TempDir(), nil, 0, "beforeafter", "")
	})

	t.Run("DI12", func(t *testing.T) {
		t.Parallel()
		for _, alias := range []string{"Join", "Collect"} {
			alias := alias
			t.Run(alias, func(t *testing.T) {
				nativeCases.Parallel(t)
				assets := distributiontest.CompleteOverride(t)
				indexPath := filepath.Join(assets, "stdlib/index.json")
				var index struct {
					Format    int `json:"format"`
					Languages map[string]map[string]struct {
						Entry      string `json:"entry"`
						Dictionary string `json:"dictionary"`
					} `json:"languages"`
				}
				if err := json.Unmarshal([]byte(read(t, indexPath)), &index); err != nil {
					t.Fatal(err)
				}
				if index.Format != 1 || index.Languages["eng"] == nil {
					t.Fatal("unsupported dictionary index")
				}
				index.Languages["eng"]["joins"] = struct {
					Entry      string `json:"entry"`
					Dictionary string `json:"dictionary"`
				}{"stdlib/eng/joins/dictionary.l", "Joins"}
				data, err := json.Marshal(index)
				if err != nil {
					t.Fatal(err)
				}
				write(t, indexPath, string(data))
				write(t, filepath.Join(assets, "stdlib/eng/joins/dictionary.l"), "Create \"Joins\" as a Dictionary.\nCreate Joins's \"Join|-s|-ed|-ing|-er\" for Awaiting any App and then Showing it's Output.\n")
				source := "Use IO's Print from `io` as \"Print\".\nUse Joins's Join from `joins` as \"" + alias + "\".\nCreate \"Worker\" as a New Basic App.\nThe App Prints `worker`.\nStart Worker.\nCreate \"Other\" as a New Basic App.\nThe App Prints `other`.\n" + alias + " Worker.\n"
				dir := project(t, source, nil)
				binary := buildProgram(t, phase1.CLI(t), dir, "main.l", assets)
				outcome(t, binary, t.TempDir(), nil, 0, "worker\n", "")
			})
		}
	})

	t.Run("LF01", func(t *testing.T) {
		nativeCases.Parallel(t)
		main := read(t, filepath.Join(root, "demo/dictionaries/file-import/main.l"))
		helper := read(t, filepath.Join(root, "demo/dictionaries/file-import/helpers.l"))
		dir := project(t, main, map[string]string{"helpers.l": helper})
		binary := buildProgram(t, phase1.CLI(t), dir, "main.l", "")
		outcome(t, binary, t.TempDir(), nil, 0, "local hello|", "")
		if read(t, filepath.Join(dir, "main.l")) != main || read(t, filepath.Join(dir, "helpers.l")) != helper {
			t.Fatal("compiler changed authored local-file sources")
		}
	})
	t.Run("LF02", func(t *testing.T) {
		nativeCases.Parallel(t)
		helper := read(t, filepath.Join(root, "demo/dictionaries/file-import/helpers.l"))
		source := "Use LocalTools from `helpers.l` as \"Utilities\".\nUse Utilities's Display as \"Paint\".\nPaint Utilities's Greeting.\n"
		dir := project(t, source, map[string]string{"helpers.l": helper})
		binary := buildProgram(t, phase1.CLI(t), dir, "main.l", "")
		outcome(t, binary, t.TempDir(), nil, 0, "local hello|", "")
	})
	t.Run("LF03", func(t *testing.T) {
		nativeCases.Parallel(t)
		helper := read(t, filepath.Join(root, "demo/dictionaries/file-import/helpers.l"))
		dir := project(t, "Use `nested/entry.l`.\n", map[string]string{"helpers.l": helper, "nested/entry.l": "Use LocalTools's Greeting from `../helpers.l` as \"Greeting\".\nUse LocalTools's Display as \"Display\".\nDisplay Greeting.\n"})
		cwd := t.TempDir()
		write(t, filepath.Join(cwd, "helpers.l"), "wrong cwd source!")
		binary := buildProgram(t, phase1.CLI(t), cwd, filepath.Join(dir, "main.l"), "")
		outcome(t, binary, cwd, nil, 0, "local hello|", "")
	})
	t.Run("MG04_phase1_demo", func(t *testing.T) {
		nativeCases.Parallel(t)
		source := read(t, filepath.Join(root, "demo/phase1/main.l"))
		dir := project(t, source, nil)
		binary := buildProgram(t, phase1.CLI(t), dir, "main.l", "")
		outcome(t, binary, t.TempDir(), nil, 0, "First\nSecond\n", "")
		if read(t, filepath.Join(dir, "main.l")) != source {
			t.Fatal("compiler changed phase1 demo")
		}
	})
	t.Run("MG09_newline_mismatch", func(t *testing.T) {
		nativeCases.Parallel(t)
		source := read(t, filepath.Join(root, "cmd/l/test/main.l"))
		dir := project(t, source, nil)
		binary := buildProgram(t, phase1.CLI(t), dir, "main.l", "")
		outcome(t, binary, t.TempDir(), nil, 1, "", "text values of both x & y should be Hello World\n")
	})
}
