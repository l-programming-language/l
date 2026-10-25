//go:build darwin || linux

package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	"github.com/l-programming-language/l/lerrors"
	Logiceme "github.com/l-programming-language/l/logiceme"
)

// The retained scenario paths now exercise Apps directly. Their independent
// expected Rust and native observers preserve the original process contracts.
type runnableServiceCase struct {
	name, source, stdout string
	apps                 []string
}

var runnableServiceCases = []runnableServiceCase{
	{"basic", "basic.l", "work completed\n", []string{"Worker"}},
	{"entry-and-target-alias", "alias.l", "work completed\n", []string{"Worker"}},
	{"mixed-apps", "mixed.l", "app\nwork completed\n", []string{"Host", "Worker"}},
	{"independent-apps", "multiple.l", "first\nsecond\n", []string{"First", "Second"}},
	{"canonical-io", "canonical-io.l", "work completed\n", []string{"Worker"}},
	{"raw-unicode-output", "unicode.l", "Snow 雪, \"quoted\"\n\n  ", []string{"Worker"}},
	{"snapshot-and-rerun", "snapshot.l", "work completed\nwork completed\n", []string{"Worker"}},
	{"stop-and-repeated-await", "stop.l", "settled", []string{"Worker"}},
	{"build-does-not-launch", "build-only.l", "built", []string{"Worker"}},
	{"empty-entrypoint", "empty.l", "", []string{"Worker"}},
	{"passive-lifecycle", "passive.l", "work completed\n", []string{"Worker"}},
	{"imported-app-declaration", "imported.l", "work completed\n", []string{"Worker"}},
	{"private-paragraph-app", "private.l", "work completed\ndone", []string{"Worker"}},
	{"await-conditional-exit", "conditional-exit.l", "", []string{"Worker"}},
}

func runnableRead(t testing.TB, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("test", "runnable_service", filepath.FromSlash(name)))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func runnableProject(t *testing.T, source string) (string, string) {
	t.Helper()
	root := t.TempDir()
	for destination, input := range map[string]string{"main.l": source, "probe.l": "probe.l", "worker.l": "worker.l", "worker-app-control.l": "worker-app-control.l"} {
		if err := os.WriteFile(filepath.Join(root, destination), runnableRead(t, input), 0600); err != nil {
			t.Fatal(err)
		}
	}
	physical, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	return physical, filepath.Join(physical, "main.l")
}

func runnableProcess(t *testing.T, source string) (*compiler, *generatedProgram, string) {
	t.Helper()
	project, filename := runnableProject(t, source)
	c := newCompiler(&cliData{jobType: "build", mainPath: filename, osType: "ubuntu"})
	c.initialize()
	if c.err != nil {
		t.Fatalf("initialize source compiler: %v", c.err)
	}
	var generated *generatedProgram
	if output := sourceStdout(t, func() { generated = c.process() }); output != "" {
		t.Fatalf("compiler diagnostics contaminated stdout: %q", output)
	}
	if c.err != nil || generated == nil {
		t.Fatalf("App scenario must generate source through compiler.process(); downstream Rust/publication/runtime not reached: %v", c.err)
	}
	if !bytes.Equal(c.mainBytes, runnableRead(t, source)) || !bytes.Equal(c.compiledBytes, generated.Compilation.ExpandedSource) || !bytes.Equal(c.asmBytes, generated.Compilation.AML) {
		t.Fatal("source generation changed source bytes or lost compilation views")
	}
	for _, name := range []string{"main", "main.apps", "Cargo.toml", "target"} {
		if _, err := os.Stat(filepath.Join(project, name)); !os.IsNotExist(err) {
			t.Fatalf("source-only process created native output %s: %v", name, err)
		}
	}
	nativetest.RecordBundle(t, "process-apps", generated.Sources)
	nativetest.RecordBundle(t, "process-libraries", generated.LibrarySources)
	return c, generated, project
}

func runnableBundleFiles(t testing.TB, bundle Logiceme.Executable, index int) map[string][]byte {
	t.Helper()
	if index >= len(bundle.ByteArrays) || index >= len(bundle.MapArrays) || index >= len(bundle.DependencyArrays) || index >= len(bundle.EnvDepArrays) || len(bundle.ByteArrays[index]) != len(bundle.MapArrays[index]) {
		t.Fatalf("incomplete source bundle at %d", index)
	}
	files := make(map[string][]byte)
	for i, name := range bundle.MapArrays[index] {
		if _, duplicate := files[name]; duplicate {
			t.Fatalf("duplicate raw artifact %s", name)
		}
		files[name] = bundle.ByteArrays[index][i]
	}
	return files
}

func TestRunnableServiceProcessSource(t *testing.T) {
	for _, tc := range runnableServiceCases {
		t.Run(tc.name, func(t *testing.T) {
			_, generated, _ := runnableProcess(t, tc.source)
			if !reflect.DeepEqual(generated.ApplicationIDs, tc.apps) || len(generated.Compilation.Compilation.Program.Apps) != len(tc.apps) {
				t.Fatalf("canonical App inventory=%v; want %v", generated.ApplicationIDs, tc.apps)
			}
			if len(generated.Sources.ByteArrays) != len(tc.apps) || len(generated.Sources.MapArrays) != len(tc.apps) || len(generated.Sources.DependencyArrays) != len(tc.apps) || len(generated.Sources.EnvDepArrays) != len(tc.apps) {
				t.Fatal("App source bundle dimensions differ from the canonical inventory")
			}
			if len(generated.LibraryIDs) != 0 || len(generated.LibrarySources.ByteArrays) != 0 || len(generated.LibrarySources.MapArrays) != 0 || len(generated.LibrarySources.DependencyArrays) != 0 || len(generated.LibrarySources.EnvDepArrays) != 0 {
				t.Fatal("simple App scenarios acquired standalone Service libraries")
			}
		})
	}
}

func TestStandaloneServiceProcessSource(t *testing.T) {
	project, filename := runnableProject(t, "basic.l")
	const source = "Create \"Worker\" as a New Main Service.\n"
	if err := os.WriteFile(filename, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	c := newCompiler(&cliData{jobType: "build", mainPath: filename, osType: "ubuntu"})
	c.initialize()
	var generated *generatedProgram
	if output := sourceStdout(t, func() { generated = c.process() }); c.err != nil || generated == nil || output != "" {
		t.Fatalf("structural standalone Service: %v/%q", c.err, output)
	}
	record := generated.Compilation.Compilation.Memory["Worker"]
	if record == nil || record.Value == nil || record.Value.Applicative == nil || record.Value.Applicative.Service == nil {
		t.Fatal("standalone Service lost its semantic record")
	}
	if len(generated.ApplicationIDs) != 0 || len(generated.LibraryIDs) != 0 || len(generated.Compilation.Compilation.Program.Apps) != 0 {
		t.Fatal("structural Service acquired an App or library target")
	}
	for _, bundle := range []Logiceme.Executable{generated.Sources, generated.LibrarySources} {
		if len(bundle.ByteArrays) != 0 || len(bundle.MapArrays) != 0 || len(bundle.DependencyArrays) != 0 || len(bundle.EnvDepArrays) != 0 {
			t.Fatal("structural Service emitted a source bundle")
		}
	}
	for _, name := range []string{"main", "main.apps", "Cargo.toml", "target"} {
		if _, err := os.Stat(filepath.Join(project, name)); !os.IsNotExist(err) {
			t.Fatalf("source-only structural declaration created native output %s: %v", name, err)
		}
	}
}

func TestStandaloneServiceBodyUnsupported(t *testing.T) {
	_, filename := runnableProject(t, "library-only.l")
	c := newCompiler(&cliData{jobType: "build", mainPath: filename, osType: "ubuntu"})
	c.initialize()
	var generated *generatedProgram
	output := sourceStdout(t, func() { generated = c.process() })
	if generated != nil || output != "" || !errors.Is(c.err, lerrors.Unsupported) || !strings.Contains(c.err.Error(), "standalone Service") || !strings.Contains(c.err.Error(), filename+":3:") {
		t.Fatalf("standalone body requires a source-attributed Unsupported diagnostic: %v/%q", c.err, output)
	}
}

func TestRunnableServicePackageNameSourceControls(t *testing.T) {
	for _, app := range []string{"l_service_worker", "L_SERVICE_WORKER", "build", "l_runtime_support"} {
		t.Run(app, func(t *testing.T) {
			project, filename := runnableProject(t, "controls/mixed.l")
			source := bytes.ReplaceAll(runnableRead(t, "controls/mixed.l"), []byte("Host"), []byte(app))
			if err := os.WriteFile(filename, source, 0600); err != nil {
				t.Fatal(err)
			}
			c := newCompiler(&cliData{jobType: "build", mainPath: filepath.Join(project, "main.l"), osType: "ubuntu"})
			c.initialize()
			if c.err != nil {
				t.Fatal(c.err)
			}
			var generated *generatedProgram
			if output := sourceStdout(t, func() { generated = c.process() }); c.err != nil || generated == nil || output != "" {
				t.Fatalf("collision source control: %v/%q", c.err, output)
			}
			if !reflect.DeepEqual(generated.ApplicationIDs, []string{app, "Worker"}) {
				t.Fatalf("collision fixture lost canonical IDs: %v", generated.ApplicationIDs)
			}
		})
	}
}

// Observe the App-only manifest through JSON independently of the production type.
type runnableManifest struct {
	Version    int                 `json:"version"`
	Entrypoint string              `json:"entrypoint"`
	GOOS       string              `json:"goos"`
	GOARCH     string              `json:"goarch"`
	Apps       []runnableCompanion `json:"apps"`
}

type runnableCompanion struct {
	ID         string `json:"id"`
	Executable string `json:"executable"`
	SHA256     string `json:"sha256"`
}

func runnableSubstituteExpected(t testing.TB, data []byte, source, generation string) []byte {
	t.Helper()
	// Only the expectation is instantiated. Actual native sources are always
	// compared and compiled byte-for-byte, including diagnostic source paths.
	for marker, value := range map[string]string{"@@SOURCE@@": source, "@@GENERATION@@": generation} {
		if strings.Contains(value, "@@") || strings.ContainsAny(value, "\r\n\"\\") {
			t.Fatalf("unsupported test-owned expected substitution %q", value)
		}
		data = bytes.ReplaceAll(data, []byte(marker), []byte(value))
	}
	if bytes.Contains(data, []byte("@@")) {
		t.Fatal("unbound independent expectation slot")
	}
	return data
}

func runnableError(t testing.TB, line int, operation, message string, filename string) string {
	t.Helper()
	return fmt.Sprintf("%s:%d:1: %s Worker: %s\n", filename, line, operation, message)
}
