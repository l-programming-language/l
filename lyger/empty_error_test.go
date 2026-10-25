//go:build darwin || linux

package Lyger

import (
	"bytes"
	"context"
	"errors"
	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	"github.com/l-programming-language/l/internal/testsupport/phase1"
	Logiceme "github.com/l-programming-language/l/logiceme"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestEmptyErrorCompleteProgram(t *testing.T) {
	source, err := os.ReadFile("test/empty-error.l")
	if err != nil {
		t.Fatal(err)
	}
	compiled := phase1Require(t, string(source), nil)
	units, err := compiled.ApplicationUnits()
	if err != nil || len(units) != 1 || units[0].ID != "Worker" {
		t.Fatalf("actual application units: %+v, %v", units, err)
	}
	companions := executableTestCompanions(compiled)
	entry, err := compiled.OrchestratorUnit("main", companions)
	if err != nil {
		t.Fatal(err)
	}
	build := func(unit BuildUnit, runtime bool) string {
		bundle, err := Logiceme.NewLogicemeService(phase1Options(t).Resources).Evaluate(unit.BackendAML, Logiceme.GenerationOptions{IncludeRuntime: runtime})
		if err != nil {
			t.Fatal(err)
		}
		recordBundle(t, bundle)
		observeBundleSources(t, bundle)
		files := map[string][]byte{}
		for i, path := range bundle.MapArrays[0] {
			files[strings.TrimPrefix(path, "/")] = bytes.Clone(bundle.ByteArrays[0][i])
		}
		attachRetainedBuildLock(t, files, unit.ID, runtime)
		project := nativetest.Materialize(t, files)
		cache, err := filepath.Abs("../.harness/cache/rust-empty-error-frontend")
		if err != nil {
			t.Fatal(err)
		}
		built := nativetest.Build(t, project, nativetest.BuildOptions{TargetDir: cache})
		if built.Err != nil {
			t.Fatalf("actual %s project: %v\n%s", unit.ID, built.Err, built.Stderr)
		}
		for _, artifact := range built.Artifacts {
			if artifact.Executable != "" {
				nativetest.AssertNativeArtifact(t, artifact.Executable)
				return artifact.Executable
			}
		}
		t.Fatal("Cargo returned no actual executable")
		return ""
	}
	worker := build(units[0], false)
	main := build(entry, true)
	dir := t.TempDir()
	binary := filepath.Join(dir, "main")
	phase1.CopyFile(t, main, binary, 0700)
	phase1.CopyFile(t, worker, filepath.Join(dir, companions[0].RelativeExecutable), 0700)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := exec.Command(binary)
	command.Dir = t.TempDir()
	result := phase1.RunRuntimeDriver(ctx, command)
	var exit *exec.ExitError
	if ctx.Err() != nil || !errors.As(result.Err, &exit) || exit.ExitCode() != 1 || len(result.Stdout) != 0 || !bytes.Equal(result.Stderr, []byte("\n")) {
		t.Fatalf("empty Error: status=%v stdout=%q stderr=%q; want status 1, empty stdout, one newline", result.Err, result.Stdout, result.Stderr)
	}
	if result.CleanupErr != nil || len(result.Leaks) != 0 || len(result.SignaledGroups) != 0 {
		t.Fatalf("cleanup required outer fallback: %+v", result)
	}
}
