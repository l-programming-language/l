package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/l-programming-language/l/tools/internal/testgate"
)

func TestHarnessJSONFixture(t *testing.T) {
	if os.Getenv("L_HARNESS_JSON_FIXTURE") != "1" {
		return
	}
	if _, err := io.WriteString(os.Stdout, "{\"Action\":\"output\",\"Output\":\"raw \\u0000雪\\r\\n\"}\n{\"Action\":\"run\",\"Package\":\"fixture\",\"Test\":\"TestDone\"}\n{\"Action\":\"pass\",\"Package\":\"fixture\",\"Test\":\"TestDone\"}\n"); err != nil {
		t.Fatal(err)
	}
	os.Exit(0)
}

func TestRunCommandUsesOneJSONArtifact(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	r := runCommand(context.Background(), dir, commandSpec{
		Name: "json", Executable: executable, Args: []string{"-test.run=^TestHarnessJSONFixture$"},
		Env: append(os.Environ(), "L_HARNESS_JSON_FIXTURE=1"), TestJSON: true, Timeout: 5 * time.Second,
	})
	if !r.succeeded() {
		t.Fatalf("JSON command failed: %+v", r)
	}
	if r.StdoutPath != r.TestEventsPath || filepath.Base(r.StdoutPath) != "json.test-events.jsonl" {
		t.Errorf("JSON stdout must share its event artifact: %+v", r)
	}
	for path, want := range map[string]string{
		r.StdoutPath:     "{\"Action\":\"output\",\"Output\":\"raw \\u0000雪\\r\\n\"}\n{\"Action\":\"run\",\"Package\":\"fixture\",\"Test\":\"TestDone\"}\n{\"Action\":\"pass\",\"Package\":\"fixture\",\"Test\":\"TestDone\"}\n",
		r.TestOutputPath: "raw \x00雪\r\n",
		r.StderrPath:     "",
	} {
		got, err := os.ReadFile(path)
		if err != nil || string(got) != want {
			t.Errorf("artifact %s: %q, %v; want %q", path, got, err, want)
		}
	}
	if r.Events == nil || len(r.Events.Tests) != 1 || r.Events.Tests[0] != (testgate.Outcome{Package: "fixture", Name: "TestDone", Status: "pass"}) {
		t.Errorf("lost completed outcome: %+v", r.Events)
	}
	files, err := os.ReadDir(dir)
	if err != nil || len(files) != 3 {
		t.Errorf("expected exactly raw JSON, stderr and readable output: %v, %v", files, err)
	}
	if _, err := os.Stat(filepath.Join(dir, "json.stdout.log")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("redundant stdout artifact exists: %v", err)
	}
}

func TestFrontendProbeBackingFilesAreInert(t *testing.T) {
	root := integrationFixture(t)
	root, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, probe := range []struct{ pkg, template, name string }{
		{"cmd/l", "compiler.go.txt", "TestHarnessCompilerProbe"},
		{"lyger", "imports.go.txt", "TestHarnessImportProbe"},
	} {
		for _, dir := range []string{probe.pkg, "tools/harness/probes"} {
			if err := os.MkdirAll(filepath.Join(root, dir), 0700); err != nil {
				t.Fatal(err)
			}
		}
		if err := os.WriteFile(filepath.Join(root, probe.pkg, "fixture.go"), []byte("package fixture\n"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, probe.pkg, "fixture_test.go"), []byte("package fixture\n"), 0600); err != nil {
			t.Fatal(err)
		}
		content := "package fixture\nimport (\"testing\"; \"os\"; \"path/filepath\")\nfunc " + probe.name + "(t *testing.T) { if err := os.WriteFile(filepath.Join(os.Getenv(\"L_HARNESS_ARTIFACT_DIR\"), \"report.json\"), []byte(\"{}\"), 0600); err != nil { t.Fatal(err) } }\n"
		if err := os.WriteFile(filepath.Join(root, "tools/harness/probes", probe.template), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	dir := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	results, err := runFrontendDiagnostics(ctx, root, dir)
	if err != nil || len(results) != 2 {
		t.Fatalf("frontend probes: %+v, %v", results, err)
	}
	for index, name := range []string{"compiler", "imports"} {
		if !results[index].succeeded() {
			t.Fatalf("%s probe failed: %+v", name, results[index])
		}
		data, err := os.ReadFile(filepath.Join(dir, "frontend", name, "overlay.json"))
		if err != nil {
			t.Fatal(err)
		}
		var overlay struct{ Replace map[string]string }
		if err := json.Unmarshal(data, &overlay); err != nil || len(overlay.Replace) != 1 {
			t.Fatalf("invalid overlay: %s, %v", data, err)
		}
		for virtual, backing := range overlay.Replace {
			if filepath.Base(backing) != "probe_test.go.txt" || filepath.Base(virtual) != "zz_harness_probe_test.go" {
				t.Errorf("incorrect source boundary: %s -> %s", virtual, backing)
			}
			if _, err := os.Stat(virtual); !errors.Is(err, os.ErrNotExist) {
				t.Errorf("virtual source materialized: %v", err)
			}
			if _, err := os.ReadFile(backing); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestHarnessRejectsUnsupportedBackendOption(t *testing.T) {
	root := integrationFixture(t)
	t.Setenv("L_HARNESS_ROOT", root)
	t.Setenv("L_HARNESS_ORIGINAL_GOFLAGS", "")
	var stdout, stderr bytes.Buffer
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	code := runCLI(ctx, []string{"diagnose", "--backend"}, &stdout, &stderr)
	if code != 2 || !strings.Contains(stderr.String(), "flag provided but not defined: -backend") || stdout.Len() != 0 {
		t.Errorf("expected immediate unknown option: code=%d stdout=%s stderr=%s", code, &stdout, &stderr)
	}
	if _, err := os.Stat(filepath.Join(root, ".harness/runs")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("unsupported option allocated a run: %v", err)
	}
}

func TestFileSHA256ValuesAndReadErrors(t *testing.T) {
	for _, tc := range []struct{ name, input, want string }{
		{"empty", "", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{"text", "abc", "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "input")
			if err := os.WriteFile(path, []byte(tc.input), 0600); err != nil {
				t.Fatal(err)
			}
			if got, err := testgate.FileHash(path); err != nil || got != tc.want {
				t.Fatalf("digest=%q, error=%v", got, err)
			}
		})
	}
	for _, name := range []string{"missing", "directory"} {
		t.Run(name, func(t *testing.T) {
			path := t.TempDir()
			if name == "missing" {
				path = filepath.Join(path, "absent")
			}
			got, err := testgate.FileHash(path)
			var native *os.PathError
			if got != "" || !errors.As(err, &native) || native.Path != path || name == "missing" && !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("lost native read failure: digest=%q, error=%v", got, err)
			}
		})
	}
}
