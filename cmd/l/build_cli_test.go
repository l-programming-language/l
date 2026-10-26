//go:build darwin || linux

package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	distributiontest "github.com/l-programming-language/l/internal/testsupport/distribution"
	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	"github.com/l-programming-language/l/internal/testsupport/phase1"
	"github.com/l-programming-language/l/internal/timing"
)

func TestBuildCLI(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	cli := filepath.Join(t.TempDir(), "compiler installation", "bin", "l")
	phase1.CopyFile(t, phase1.CLI(t), cli, 0700)
	const printProgram = "Use IO's Print from `io` as \"Print\".\n\"x\" is `literal <DEF> 世界`.\nCreate \"HelloWorldApp\" as a New Basic App.\nThe App Prints x.\n"
	const successfulEnd = "Run it.\nAwait it.\nIf it has an Error, exit with the Error, else define \"y\" as it's Output.\nShow y.\nExit.\n"
	for _, tc := range []struct {
		name, filename             string
		defaultSource              bool
		source, wantOut, wantError string
		apps                       []string
	}{
		{"explicit source", "main.l", false, "", "Hello World!\n", "", []string{"Hello"}},
		{"default source", "main.l", true, "", "Hello World!\n", "", []string{"Hello"}},
		{"renamed source with spaces", "worker project.l", false, printProgram + successfulEnd, "literal <DEF> 世界\n", "", []string{"HelloWorldApp"}},
		{"explicit Build then Run", "main.l", false, printProgram + "Build it.\n" + successfulEnd, "literal <DEF> 世界\n", "", []string{"HelloWorldApp"}},
		{"passive runtime assignment", "main.l", false, printProgram + "Execute it.\n\"y\" is HelloWorldApp's Output.\nShow y.\nExit.\n", "literal <DEF> 世界\n", "", []string{"HelloWorldApp"}},
		{"unavailable Output", "main.l", false, printProgram + "Show HelloWorldApp's Output.\n", "", "HelloWorldApp: Output is unavailable before completed invocation capture\n", []string{"HelloWorldApp"}},
		{"passive unavailable Output", "main.l", false, printProgram + "\"y\" is HelloWorldApp's Output.\nShow y.\nExit.\n", "", "HelloWorldApp: Output is unavailable before completed invocation capture\n", []string{"HelloWorldApp"}},
		{"imported unavailable Output", "main.l", false, printProgram + "Use `misuse.l`.\n", "", "HelloWorldApp: Output is unavailable before completed invocation capture\n", []string{"HelloWorldApp"}},
		{"imported action misuse alias", "main.l", false, printProgram + "Use `misuse.l`.\nUse HelloWorldApp as \"Primary\".\nJoin Primary.\n", "", "HelloWorldApp: Await requires an invocation\n", []string{"HelloWorldApp"}},
		{"scalar Show", "main.l", false, printProgram + "\"count\" is 42.\n\"enabled\" is true.\nShow count.\nShow enabled.\nExit.\n", "42true", "", []string{"HelloWorldApp"}},
		{"explicit paths and native environment", "-worker.l", false, printProgram + successfulEnd, "literal <DEF> 世界\n", "", []string{"HelloWorldApp"}},
		{"Build does not execute source Execute", "main.l", false, "Use BuildObserver's Print from `build-marker.l` as \"Print\".\n\"x\" is `literal <DEF> 世界`.\nCreate \"HelloWorldApp\" as a New Basic App.\nThe App Prints x.\nExecute it.\nExit.\n", "", "", []string{"HelloWorldApp"}},
		{"two Apps and alias", "main.l", false, "Use IO's Print from `io` as \"Print\".\n\"firstText\" is `One`.\n\"secondText\" is `Two`.\nCreate \"First\" as a New Basic App.\nThe App Prints firstText.\nCreate \"Second\" as a New Basic App.\nThe App Prints secondText.\nUse First as \"Primary\".\nRun Primary.\nRun Second.\nAwait Primary.\nAwait Second.\nShow First's Output.\nShow Second's Output.\nExit.\n", "One\nTwo\n", "", []string{"First", "Second"}},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			admitted := timing.Start("admission.wait", t.Name())
			buildCLICases <- struct{}{}
			admitted(nil)
			t.Cleanup(func() { <-buildCLICases })
			var err error
			project := filepath.Join(t.TempDir(), "source project")
			copyBuildProject(t, root, project, tc.filename, tc.source)
			project, err = filepath.EvalSymlinks(project)
			if err != nil {
				t.Fatal(err)
			}
			switch tc.name {
			case "unavailable Output":
				tc.wantError = fmt.Sprintf("%s:5:1: show %s", filepath.Join(project, tc.filename), tc.wantError)
			case "passive unavailable Output":
				tc.wantError = fmt.Sprintf("%s:5:1: assign %s", filepath.Join(project, tc.filename), tc.wantError)
			case "imported unavailable Output", "imported action misuse alias":
				imported := "Show HelloWorldApp's Output.\n"
				errorPath, line, operation := filepath.Join(project, "misuse.l"), 1, "show"
				if tc.name == "imported action misuse alias" {
					imported = "Create \"Join|-s|-ed|-ing|-er\" for Awaiting any App.\n"
					errorPath, line, operation = filepath.Join(project, tc.filename), 7, "await"
				}
				if err := os.WriteFile(filepath.Join(project, "misuse.l"), []byte(imported), 0600); err != nil {
					t.Fatal(err)
				}
				tc.wantError = fmt.Sprintf("%s:%d:1: %s %s", errorPath, line, operation, tc.wantError)
			}
			var sideEffect string
			if tc.name == "Build does not execute source Execute" {
				sideEffect = filepath.Join(project, "runtime-side-effect")
				data, err := os.ReadFile(filepath.Join(root, "testdata/distribution/observers/build-marker.l"))
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(project, "build-marker.l"), data, 0600); err != nil {
					t.Fatal(err)
				}
			}
			args := []string{"build"}
			outputDirectory := project
			var environment []string
			if sideEffect != "" {
				environment = append(environment, "L_BUILD_SIDE_EFFECT="+sideEffect)
			}
			if tc.name == "explicit paths and native environment" {
				outputDirectory = filepath.Join(project, "build output")
				override := distributiontest.CompleteOverride(t)
				assets, err := filepath.Rel(project, override)
				if err != nil {
					t.Fatal(err)
				}
				args = append(args, "--assets", assets, "--output-dir", "build output", "--")
				environment = []string{"RUSTFLAGS=--invalid-rust-flag", "CARGO_BUILD_TARGET=invalid-host", "RUSTC_WRAPPER=/missing/rustc-wrapper", "GOOS=windows", "GOARCH=386", "GO111MODULE=off", "GOFLAGS=-invalid-flag", "GOWORK=" + filepath.Join(project, "missing.work"), "GOTOOLCHAIN=go99.0.0"}
			}
			if !tc.defaultSource {
				args = append(args, tc.filename)
			}
			recorder := nativetest.CLIProbe(t)
			stdout, stderr, err := runBuildCommand(t, cli, project, recorder.Env(environment), args...)
			if err != nil {
				t.Fatalf("l build failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
			}
			stem := strings.TrimSuffix(tc.filename, ".l")
			path := filepath.Join(outputDirectory, stem)
			recorder.AssertPublished(t, path)
			info, err := os.Stat(path)
			if err != nil {
				t.Fatalf("l build returned success but no orchestrator exists at %s: %v", path, err)
			}
			if !info.Mode().IsRegular() || info.Size() == 0 || info.Mode().Perm()&0111 == 0 {
				t.Fatalf("output is not a nonempty executable: %v", info)
			}
			if string(stdout) != path+"\n" || len(stderr) != 0 {
				t.Fatalf("build output = %q, stderr = %q; want only executable path", stdout, stderr)
			}
			if sideEffect != "" {
				if _, err := os.Stat(sideEffect); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("build executed the App: %v", err)
				}
			}
			if info, err := os.Stat(path + ".apps"); err != nil || !info.IsDir() {
				t.Fatalf("missing App companion directory: %v", err)
			}
			nativetest.AssertNativeArtifact(t, path)
			var manifest generationManifest
			manifestPath := nativetest.ReadPublishedGeneration(t, path, &manifest)
			if manifest.GOOS != runtime.GOOS || manifest.GOARCH != runtime.GOARCH || len(manifest.Apps) != len(tc.apps) {
				t.Fatalf("manifest = %+v", manifest)
			}
			for index, app := range manifest.Apps {
				if app.ID != tc.apps[index] {
					t.Fatalf("App %d = %s; want %s", index, app.ID, tc.apps[index])
				}
				appPath := filepath.Join(filepath.Dir(manifestPath), app.Executable)
				if app.SHA256 != nativetest.NativeArtifactHash(t, appPath) {
					t.Fatal("App hash does not match its manifest")
				}
			}
			// Running from elsewhere proves companion lookup is executable-relative.
			stdout, stderr, err = runRuntimeCommand(t, path, t.TempDir(), environment)
			code := 0
			if err != nil {
				if failure, ok := err.(*exec.ExitError); ok {
					code = failure.ExitCode()
				} else {
					t.Fatal(err)
				}
			}
			wantCode := 0
			if tc.wantError != "" {
				wantCode = 1
			}
			if code != wantCode || string(stdout) != tc.wantOut || string(stderr) != tc.wantError {
				t.Fatalf("L execution status=%d stdout=%q stderr=%q; want %d %q %q", code, stdout, stderr, wantCode, tc.wantOut, tc.wantError)
			}
			if sideEffect != "" {
				content, err := os.ReadFile(sideEffect)
				if err != nil || string(content) != "literal <DEF> 世界" {
					t.Fatalf("orchestrator did not execute source action: %q %v", content, err)
				}
			}
			if tc.name == "renamed source with spaces" {
				relocated := filepath.Join(t.TempDir(), "moved program")
				if err := os.Mkdir(relocated, 0700); err != nil {
					t.Fatal(err)
				}
				newPath := filepath.Join(relocated, stem)
				if err := os.Rename(path, newPath); err != nil {
					t.Fatal(err)
				}
				if err := os.Rename(path+".apps", newPath+".apps"); err != nil {
					t.Fatal(err)
				}
				stdout, stderr, err = runRuntimeCommand(t, newPath, t.TempDir(), nil)
				if err != nil || string(stdout) != tc.wantOut || len(stderr) != 0 {
					t.Fatalf("relocated execution: %v %q %q", err, stdout, stderr)
				}
			}
			if tc.name == "explicit Build then Run" {
				original, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				for _, failure := range []struct {
					name, source string
					env          []string
				}{
					{"compiler", "Unrecognized broken program.\n", nil},
					{"orchestrator", printProgram + "Show HelloWorldApp.\n", nil},
					{"Cargo", tc.source, []string{"PATH=" + t.TempDir()}},
				} {
					if err := os.WriteFile(filepath.Join(project, tc.filename), []byte(failure.source), 0600); err != nil {
						t.Fatal(err)
					}
					out, diagnostics, err := runBuildCommand(t, cli, project, failure.env, args...)
					var exit *exec.ExitError
					if !errors.As(err, &exit) || exit.ExitCode() != 1 || len(out) != 0 || len(diagnostics) == 0 {
						t.Fatalf("%s failure status/streams: %v %q %q", failure.name, err, out, diagnostics)
					}
					after, err := os.ReadFile(path)
					if err != nil || !bytes.Equal(original, after) {
						t.Fatalf("%s failure replaced previous executable: %v", failure.name, err)
					}
					remaining, err := filepath.Glob(filepath.Join(path+".apps", "generations", "*", "manifest.json"))
					if err != nil || len(remaining) != 1 {
						t.Fatalf("%s failure changed generations: %v %v", failure.name, remaining, err)
					}
					out, diagnostics, err = runRuntimeCommand(t, path, t.TempDir(), nil)
					if err != nil || string(out) != tc.wantOut || len(diagnostics) != 0 {
						t.Fatalf("prior executable stopped working: %v %q %q", err, out, diagnostics)
					}
				}
				// Missing packaged Apps are runtime Errors; Build does not invoke Go.
				if err := os.Remove(filepath.Join(filepath.Dir(manifestPath), manifest.Apps[0].Executable)); err != nil {
					t.Fatal(err)
				}
				out, diagnostics, err := runRuntimeCommand(t, path, t.TempDir(), []string{"PATH=" + t.TempDir()})
				var exit *exec.ExitError
				if !errors.As(err, &exit) || exit.ExitCode() != 1 || len(out) != 0 || !bytes.Contains(diagnostics, []byte("application companion")) {
					t.Fatalf("missing companion: %v %q %q", err, out, diagnostics)
				}
			}
		})
	}
}
