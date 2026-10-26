package Logiceme

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime/debug"
	"strings"
	"sync"
	"testing"
	"time"
)

// Complete composed application scenarios, shared by generation and compilation.
// Native-negative cases still require exact generation; only Cargo is expected
// to reject them, and only with the specified source diagnostic.
type lucidCase struct {
	name, manifest, lock, dependency, diagnostic string
}

var lucidCases = []lucidCase{
	{name: "future-composed"},
	{name: "future-move"},
	{name: "future-infer"},
	{name: "future-borrow-escape", diagnostic: "E0373"},
	{name: "future-move-reuse", diagnostic: "E0382"},
	{name: "generics-functions"},
	{name: "generics-records"},
	{name: "generics-traits"},
	{name: "generics-calls"},
	{name: "generics-late-bound-lifetime", diagnostic: "E0794"},
	{name: "generics-unused-parameters", diagnostic: "E0392"},
	{name: "generics-argument-arity", diagnostic: "E0107"},
	{name: "keyed-conditionals"},
	{name: "keyed-nonexhaustive", diagnostic: "E0004"},
	{name: "keyed-call-pattern", diagnostic: "E0164"},
	{name: "empty-instantiations"},
	{name: "trait-signature-outputs"},
	{name: "callable-output-void"},
	{name: "callable-typed-flat"},
	{name: "callable-typed-record"},
	{name: "callable-outputs"},
	{name: "closure-inline"},
	{name: "closure-stored-capture"},
	{name: "callable-unit-mismatch", diagnostic: "E0308"},
	{name: "function-void-flat"},
	{name: "function-void-record"},
	{name: "invocation-named", manifest: "common/Cargo-tokio.toml.txt", lock: "common/Cargo-tokio.lock", dependency: lucidTokio},
	{name: "invocation-void"},
	{name: "declaration-operation"},
	{name: "declaration-description"},
	{name: "declaration-evaluation"},
	{name: "declaration-instantiation"},
	{name: "declaration-iteration"},
	{name: "grouping"}, {name: "statements"}, {name: "fields"}, {name: "calculations"},
	{name: "communication", manifest: "common/Cargo-tokio.toml.txt", lock: "common/Cargo-tokio.lock", dependency: lucidTokio},
	{name: "conditionals"},
	{name: "select-communication", dependency: lucidSelectTokio},
	{name: "select-bindings", dependency: lucidSelectTokio},
	{name: "select-default-order", dependency: lucidSelectTokio},
	{name: "select-degenerate", dependency: lucidSelectTokio},
	{name: "select-nesting-hygiene", dependency: lucidSelectTokio},
	{name: "select-closed-receive", dependency: lucidSelectTokio},
	{name: "select-sync-context", dependency: lucidSelectTokio, diagnostic: "E0728"},
	{name: "select-borrow-conflict", dependency: lucidSelectTokio, diagnostic: "E0499"},
	{name: "select-moved-send", dependency: lucidSelectTokio, diagnostic: "E0382"},
	{name: "iteration-conditions"},
	{name: "iteration-bindings"},
	{name: "iteration-async", manifest: "common/Cargo-tokio.toml.txt", lock: "common/Cargo-tokio.lock", dependency: lucidTokio},
	{name: "iteration-traversal"},
	{name: "iteration-labels"},
	{name: "iteration-nesting-scope"},
	{name: "iteration-nonboolean", diagnostic: "E0308"},
	{name: "iteration-binding-scope", diagnostic: "E0425"},
	{name: "iteration-sync-await", manifest: "common/Cargo-tokio.toml.txt", lock: "common/Cargo-tokio.lock", dependency: lucidTokio, diagnostic: "E0728"},
	{name: "schedule-direct", manifest: "common/Cargo-tokio.toml.txt", lock: "common/Cargo-tokio.lock", dependency: lucidTokio},
	{name: "schedule-composed", manifest: "common/Cargo-tokio.toml.txt", lock: "common/Cargo-tokio.lock", dependency: lucidTokio},
}

const lucidSelectTokio = `tokio = { version = "=1.53.1", features = ["rt", "sync", "macros"] }`

func lucidCaseFor(name string) lucidCase {
	for _, c := range lucidCases {
		if c.name != name {
			continue
		}
		if c.manifest == "" {
			c.manifest, c.lock = "common/Cargo.toml.txt", "common/Cargo.lock"
			if c.dependency == lucidSelectTokio {
				c.manifest, c.lock = "common/Cargo-select.toml.txt", "common/Cargo-select.lock"
			}
		}
		return c
	}
	panic("unregistered Rust fixture: " + name)
}

const lucidTokio = `tokio = { version = "=1.53.1", features = ["rt", "sync"] }`

var lucidPaths = []string{"/src/main.rs", "/Cargo.toml", "/src/config.rs", "/start.sh", "/stop.sh"}

type lucidResult struct {
	Scenario        string            `json:"scenario"`
	Stage           string            `json:"stage"`
	Status          string            `json:"status"`
	Issues          []string          `json:"issues"`
	GenerationError string            `json:"generation_error,omitempty"`
	Paths           [][]string        `json:"paths,omitempty"`
	Dependencies    [][]string        `json:"dependencies,omitempty"`
	EnvDependencies [][]string        `json:"env_dependencies,omitempty"`
	Expected        map[string]string `json:"expected_files"`
	ExpectedHashes  map[string]string `json:"expected_hashes"`
	ActualHashes    map[string]string `json:"actual_hashes"`
	Command         []string          `json:"command,omitempty"`
	ExitCode        *int              `json:"exit_code,omitempty"`
	Duration        string            `json:"duration,omitempty"`
}

var lucidArtifacts struct {
	sync.Once
	root string
	err  error
}

func lucidRead(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "lucid", filepath.FromSlash(name)))
	if err != nil {
		t.Fatalf("fixture/harness: read %s: %v", name, err)
	}
	return b
}

func lucidHash(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }

func lucidExpectedFiles(name string) []string {
	return []string{name + ".rs.txt", lucidCaseFor(name).manifest, "common/config.rs.txt", "common/start.sh.txt", "common/stop.sh.txt"}
}

func lucidExpectedBundle(t *testing.T, name string) Executable {
	t.Helper()
	var contents [][]byte
	for _, f := range lucidExpectedFiles(name) {
		contents = append(contents, lucidRead(t, f))
	}
	var deps []string
	if dep := lucidCaseFor(name).dependency; dep != "" {
		deps = []string{dep}
	}
	return Executable{MapArrays: [][]string{append([]string(nil), lucidPaths...)}, ByteArrays: [][][]byte{contents}, DependencyArrays: [][]string{deps}, EnvDepArrays: [][]string{nil}}
}

func lucidWrite(t *testing.T, name string, b []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		t.Fatalf("infrastructure: create artifact directory: %v", err)
	}
	if err := os.WriteFile(name, b, 0644); err != nil {
		t.Fatalf("infrastructure: write artifact: %v", err)
	}
}

func lucidRecord(t *testing.T, name, stage string) (*lucidResult, string) {
	t.Helper()
	lucidArtifacts.Do(func() {
		root := os.Getenv("L_LUCID_ARTIFACT_DIR")
		if root == "" {
			root = filepath.Join("..", ".harness", "runs", "lucid-native-"+time.Now().UTC().Format("20060102T150405.000000000Z"))
		}
		lucidArtifacts.root, lucidArtifacts.err = filepath.Abs(root)
	})
	if lucidArtifacts.err != nil {
		t.Fatal(lucidArtifacts.err)
	}
	dir := filepath.Join(lucidArtifacts.root, stage, name)
	if err := os.MkdirAll(filepath.Dir(dir), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(dir, 0755); os.IsExist(err) {
		// Reusing the artifact root must neither overwrite prior evidence nor
		// compile files left over from an earlier, different returned bundle.
		dir, err = os.MkdirTemp(filepath.Dir(dir), name+"-")
		if err != nil {
			t.Fatal(err)
		}
	} else if err != nil {
		t.Fatal(err)
	}
	t.Logf("Rust artifacts: %s", dir)
	r := &lucidResult{Scenario: name, Stage: stage, Status: "pass", Expected: map[string]string{}, ExpectedHashes: map[string]string{}, ActualHashes: map[string]string{}}
	t.Cleanup(func() {
		if t.Failed() && r.Status == "pass" {
			r.Status = "fail"
		}
		b, err := json.MarshalIndent(r, "", "  ")
		if err != nil {
			t.Error(err)
			return
		}
		lucidWrite(t, filepath.Join(dir, "result.json"), append(b, '\n'))
	})
	if stage == "validation" || stage == "primitive" || stage == "isolation" {
		return r, dir
	}
	for i, f := range lucidExpectedFiles(name) {
		r.Expected[lucidPaths[i]] = filepath.ToSlash(filepath.Join("logiceme/testdata/lucid", f))
		r.ExpectedHashes[lucidPaths[i]] = lucidHash(lucidRead(t, f))
	}
	return r, dir
}

func lucidIssue(t *testing.T, r *lucidResult, format string, args ...any) {
	t.Helper()
	message := fmt.Sprintf(format, args...)
	r.Issues = append(r.Issues, message)
	t.Error(message)
}

func lucidGenerate(t *testing.T, name string, r *lucidResult, dir string) (bundle Executable, ok bool) {
	t.Helper()
	input := lucidRead(t, name+".aml.txt")
	original := bytes.Clone(input)
	lucidWrite(t, filepath.Join(dir, "input.aml.txt"), input)
	service := NewLogicemeService(runtimeGenerationResources(t))
	var generationErr error
	func() {
		defer func() {
			if p := recover(); p != nil {
				generationErr = fmt.Errorf("production panic: %v\n%s", p, debug.Stack())
			}
		}()
		bundle, generationErr = service.Evaluate(input, GenerationOptions{IncludeRuntime: false})
	}()
	if !bytes.Equal(input, original) {
		lucidIssue(t, r, "production generation: Evaluate mutated the complete AML input")
	}
	materialized := lucidSaveBundle(t, r, dir, bundle)
	if generationErr != nil {
		r.GenerationError = generationErr.Error()
		lucidIssue(t, r, "production generation/validation: %v", generationErr)
		if !reflect.DeepEqual(bundle, Executable{}) {
			lucidIssue(t, r, "production generation: error returned a partial bundle")
		}
		return bundle, false
	}
	lucidCompare(t, r, dir, bundle, lucidExpectedBundle(t, name))
	return bundle, materialized
}

func lucidSaveBundle(t *testing.T, r *lucidResult, dir string, b Executable) bool {
	t.Helper()
	r.Paths, r.Dependencies, r.EnvDependencies = b.MapArrays, b.DependencyArrays, b.EnvDepArrays
	raw, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	lucidWrite(t, filepath.Join(dir, "bundle.json"), append(raw, '\n'))
	valid := len(b.MapArrays) == len(b.ByteArrays)
	for ai, paths := range b.MapArrays {
		if ai >= len(b.ByteArrays) || len(paths) != len(b.ByteArrays[ai]) {
			valid = false
			continue
		}
		seen := map[string]bool{}
		for fi, p := range paths {
			rel := strings.TrimPrefix(p, "/")
			if !strings.HasPrefix(p, "/") || p != path.Clean(p) || strings.Contains(p, "\\") || !filepath.IsLocal(filepath.FromSlash(rel)) || seen[p] {
				lucidIssue(t, r, "production bundle: unsafe or duplicate path %q", p)
				valid = false
				continue
			}
			seen[p] = true
			r.ActualHashes[fmt.Sprintf("%d:%s", ai, p)] = lucidHash(b.ByteArrays[ai][fi])
			lucidWrite(t, filepath.Join(dir, "actual", fmt.Sprintf("app-%03d", ai), filepath.FromSlash(rel)), b.ByteArrays[ai][fi])
		}
	}
	if !valid {
		lucidIssue(t, r, "production bundle: cannot materialize the complete path/content inventory")
	}
	return valid
}

func lucidCompare(t *testing.T, r *lucidResult, dir string, got, want Executable) {
	t.Helper()
	if !reflect.DeepEqual(got.MapArrays, want.MapArrays) {
		lucidIssue(t, r, "production bundle: ordered paths = %q; want %q", got.MapArrays, want.MapArrays)
	}
	for _, pair := range []struct {
		label     string
		got, want [][]string
	}{{"dependencies", got.DependencyArrays, want.DependencyArrays}, {"environment dependencies", got.EnvDepArrays, want.EnvDepArrays}} {
		if len(pair.got) != 1 || len(pair.got[0]) != len(pair.want[0]) || strings.Join(pair.got[0], "\x00") != strings.Join(pair.want[0], "\x00") {
			lucidIssue(t, r, "production bundle: %s = %q; want %q", pair.label, pair.got, pair.want)
		}
	}
	if len(got.MapArrays) != 1 || len(got.ByteArrays) != 1 || len(got.MapArrays[0]) != len(got.ByteArrays[0]) {
		lucidIssue(t, r, "production bundle: expected exactly one complete application")
		return
	}
	actual := map[string][]byte{}
	for i, p := range got.MapArrays[0] {
		actual[p] = got.ByteArrays[0][i]
	}
	for i, p := range want.MapArrays[0] {
		g, found := actual[p]
		w := want.ByteArrays[0][i]
		if !found {
			lucidIssue(t, r, "production bundle: missing %s", p)
			continue
		}
		if bytes.Equal(g, w) {
			continue
		}
		offset := 0
		for offset < len(g) && offset < len(w) && g[offset] == w[offset] {
			offset++
		}
		context := func(b []byte) string {
			start := offset - 40
			if start < 0 {
				start = 0
			}
			end := offset + 80
			if end > len(b) {
				end = len(b)
			}
			return fmt.Sprintf("%q", b[start:end])
		}
		detail := fmt.Sprintf("unclassified exact-byte mismatch %s: offset %d, actual length %d, expected length %d; actual %s; expected %s", p, offset, len(g), len(w), context(g), context(w))
		lucidIssue(t, r, "%s", detail)
		diff := detail + "\n--- expected\n+++ actual\n-" + strings.ReplaceAll(string(w), "\n", "\n-") + "\n+" + strings.ReplaceAll(string(g), "\n", "\n+") + "\n"
		lucidWrite(t, filepath.Join(dir, "diffs", filepath.FromSlash(strings.TrimPrefix(p, "/"))+".diff"), []byte(diff))
	}
}

func TestLucidTemplatesFrozen(t *testing.T) {
	var hashes map[string]string
	if err := json.Unmarshal(lucidRead(t, "templates.sha256.json"), &hashes); err != nil {
		t.Fatal(err)
	}
	resources := runtimeGenerationResources(t)
	paths, err := fs.Glob(resources, "builtins/segments/*/*.lu")
	if err != nil {
		t.Fatal(err)
	}
	if len(hashes) != 18 || len(paths) != 18 {
		t.Fatalf("freeze inventory: hashes=%d files=%d; want18 each", len(hashes), len(paths))
	}
	for _, p := range paths {
		b, err := fs.ReadFile(resources, p)
		if err != nil {
			t.Error(err)
			continue
		}
		if got := lucidHash(b); got != hashes[p] {
			t.Errorf("frozen template changed: %s: got %s want %s", p, got, hashes[p])
		}
	}
}

func TestLucidGeneration(t *testing.T) {
	for _, c := range lucidCases {
		name := c.name
		t.Run(name, func(t *testing.T) {
			r, dir := lucidRecord(t, name, "generation")
			lucidGenerate(t, name, r, dir)
		})
	}
}
