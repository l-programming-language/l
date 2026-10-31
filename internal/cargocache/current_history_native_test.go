//go:build darwin || linux

package cargocache_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/l-programming-language/l/internal/cargocache"
	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
)

type historyNativeStep struct {
	library, value string
	wantFresh      bool
	checkFresh     bool
}

func TestCurrentHistoryReusesRemovedLibrary(t *testing.T) {
	historyNativeSequence(t, []historyNativeStep{
		{"P", "A", false, true},
		{},
		{"P", "A", true, true},
	})
}

func TestCurrentHistoryDoesNotReuseSupersededSource(t *testing.T) {
	historyNativeSequence(t, []historyNativeStep{
		{"P", "A", false, true},
		{},
		{"P", "B", false, true},
		{},
		{"P", "A", false, true},
	})
}

func TestCurrentHistoryDistinguishesEqualPackagePaths(t *testing.T) {
	// All three libraries have the same package, target, and source bytes.
	// Their observed manifest directory must still follow the selected path.
	historyNativeSequence(t, []historyNativeStep{
		{"P", "A", false, true},
		{"Q", "A", false, true},
		{"P", "A", false, false},
	})
}

func historyNativeSequence(t *testing.T, steps []historyNativeStep) {
	t.Helper()
	tools := nativetest.Toolchain(t)
	cache, target := t.TempDir(), t.TempDir()
	for index, step := range steps {
		files := map[string][]byte{
			"Cargo.toml":        []byte("[workspace]\nresolver = \"3\"\nmembers = [\"entry\"]\n"),
			"Cargo.lock":        []byte("version = 4\n\n[[package]]\nname = \"history_entry\"\nversion = \"0.0.0\"\n"),
			"entry/Cargo.toml":  []byte("[package]\nname = \"history_entry\"\nversion = \"0.0.0\"\nedition = \"2024\"\n"),
			"entry/src/main.rs": []byte("fn main() { println!(\"absent\"); }\n"),
		}
		if step.library != "" {
			member := "support/" + step.library
			files["Cargo.toml"] = []byte(fmt.Sprintf("[workspace]\nresolver = \"3\"\nmembers = [\"entry\", %q]\n", member))
			files["Cargo.lock"] = []byte("version = 4\n\n[[package]]\nname = \"history_entry\"\nversion = \"0.0.0\"\ndependencies = [\"history_support\"]\n\n[[package]]\nname = \"history_support\"\nversion = \"0.0.0\"\n")
			files["entry/Cargo.toml"] = append(files["entry/Cargo.toml"], []byte(fmt.Sprintf("\n[dependencies]\nhistory_support = { path = %q }\n", "../"+member))...)
			files["entry/src/main.rs"] = []byte("fn main() { println!(\"{}:{}\", history_support::value(), history_support::origin()); }\n")
			files[member+"/Cargo.toml"] = []byte("[package]\nname = \"history_support\"\nversion = \"0.0.0\"\nedition = \"2024\"\n")
			files[member+"/src/lib.rs"] = []byte(fmt.Sprintf("pub fn value() -> &'static str { %q }\npub fn origin() -> &'static str { env!(\"CARGO_MANIFEST_DIR\") }\n", step.value))
		}
		project := nativetest.Materialize(t, files)
		inputs, err := cargocache.StageCurrent(project, cache, "")
		if err != nil {
			t.Fatal(err)
		}
		if err := cargocache.Activate(inputs, target); err != nil {
			t.Fatal(err)
		}
		stdout, _ := historyNativeCommand(t, nativetest.BuildPrerequisite, tools.Cargo, inputs, nativetest.CleanEnvironment(tools.Rustc),
			"build", "--workspace", "--frozen", "--message-format=json", "--manifest-path", filepath.Join(inputs, "Cargo.toml"), "--target", tools.Host, "--target-dir", target)
		var executable string
		foundLibrary := false
		for _, line := range bytes.Split(stdout, []byte{'\n'}) {
			var event struct {
				Reason, Executable string
				Fresh              bool
				Target             struct{ Name string }
			}
			if json.Unmarshal(line, &event) != nil || event.Reason != "compiler-artifact" {
				continue
			}
			switch event.Target.Name {
			case "history_entry":
				executable = event.Executable
			case "history_support":
				if foundLibrary || step.library == "" || step.checkFresh && event.Fresh != step.wantFresh {
					t.Fatalf("step %d support fresh=%v duplicate=%v; want present=%v fresh=%v", index, event.Fresh, foundLibrary, step.library != "", step.wantFresh)
				}
				foundLibrary = true
			}
		}
		if executable == "" || foundLibrary != (step.library != "") {
			t.Fatalf("step %d incomplete native artifact inventory: executable=%q library=%v\n%s", index, executable, foundLibrary, stdout)
		}
		output, diagnostics := historyNativeCommand(t, nativetest.RuntimeExecution, executable, inputs, nativetest.CleanEnvironment(tools.Rustc))
		want := "absent\n"
		if step.library != "" {
			want = step.value + ":" + filepath.Join(inputs, "support", step.library) + "\n"
		}
		if string(output) != want || len(diagnostics) != 0 {
			t.Fatalf("step %d executed stale library: got %q / %q want %q / empty stderr", index, output, diagnostics, want)
		}
	}
}

func historyNativeCommand(t *testing.T, phase nativetest.CommandPhase, executable, directory string, environment []string, args ...string) ([]byte, []byte) {
	t.Helper()
	ctx, cancel := phase.Context(context.Background(), 5*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, executable, args...)
	command.Dir, command.Env = directory, environment
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error { return syscall.Kill(-command.Process.Pid, syscall.SIGKILL) }
	command.WaitDelay = 2 * time.Second
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("%s command failed: %v; context=%v\nstdout=%s\nstderr=%s", phase, err, ctx.Err(), stdout.Bytes(), stderr.Bytes())
	}
	return stdout.Bytes(), stderr.Bytes()
}
