// Package phase1 provides test-owned native acceptance infrastructure. Production
// packages must not import it. It does not implement App lifecycle semantics.
package phase1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type RuntimeExpectation struct {
	Stdout         json.RawMessage `json:"stdout"`
	Stderr         *string         `json:"stderr"`
	StderrContains []string        `json:"stderr_contains"`
	ExitStatus     json.RawMessage `json:"exit_status"`
}
type Fixture struct {
	ID                   string              `json:"id"`
	Title                string              `json:"title"`
	Profile              string              `json:"profile"`
	Directory            string              `json:"-"`
	Files                []string            `json:"files"`
	Apps                 []string            `json:"apps"`
	Assertions           []string            `json:"assertions"`
	ExpectedRuntime      *RuntimeExpectation `json:"expected_runtime"`
	ExpectedCompileError *struct {
		Category    string `json:"category"`
		Span        string `json:"span"`
		Publication string `json:"publication"`
	} `json:"expected_compile_error"`
}

func RepoRoot(t testing.TB) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate Phase 1 test support")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../../.."))
}

func DecodeFixture(data []byte) (Fixture, error) {
	var f Fixture
	d := json.NewDecoder(bytes.NewReader(data))
	if err := d.Decode(&f); err != nil {
		return f, err
	}
	var extra interface{}
	if err := d.Decode(&extra); err != io.EOF {
		return f, fmt.Errorf("trailing manifest content")
	}
	if len(f.ID) != 3 || (f.ID[0] != 'F' && f.ID[0] != 'N') || f.ID[1] < '0' || f.ID[1] > '9' || f.ID[2] < '0' || f.ID[2] > '9' {
		return f, fmt.Errorf("invalid fixture ID %q", f.ID)
	}
	if f.Title == "" || f.Profile == "" || len(f.Assertions) == 0 {
		return f, fmt.Errorf("missing fixture title, profile or assertions")
	}
	seen := make(map[string]bool)
	for _, name := range f.Files {
		if name == "" || filepath.IsAbs(name) || strings.Contains(name, "\\") || filepath.ToSlash(filepath.Clean(name)) != name || strings.HasPrefix(name, "../../") {
			return f, fmt.Errorf("invalid fixture dependency %q", name)
		}
		if seen[name] {
			return f, fmt.Errorf("duplicate fixture dependency %q", name)
		}
		seen[name] = true
		if filepath.Base(name) == "builtin.l" || filepath.Base(name) == "core.l" {
			return f, fmt.Errorf("fixture must not copy the language library: %s", name)
		}
		if !strings.HasSuffix(name, ".l") {
			return f, fmt.Errorf("unsupported fixture dependency %q", name)
		}
	}
	if !seen["main.l"] {
		return f, fmt.Errorf("fixture must declare main.l")
	}
	if (f.ExpectedRuntime == nil) == (f.ExpectedCompileError == nil) {
		return f, fmt.Errorf("exactly one runtime or compile expectation required")
	}
	if f.ExpectedRuntime != nil {
		if _, err := ExpectedBytes(f.ExpectedRuntime.Stdout); err != nil {
			return f, fmt.Errorf("stdout: %w", err)
		}
		if len(f.ExpectedRuntime.ExitStatus) == 0 {
			return f, fmt.Errorf("missing exit status")
		}
	}
	return f, nil
}

// ExpectedBytes expands a declarative byte oracle. It is never applied to actual
// stdout/stderr, so whitespace and NUL bytes cannot be normalized away.
func ExpectedBytes(raw json.RawMessage) ([]byte, error) {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return []byte(s), nil
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		return nil, err
	}
	if r, ok := object["repeat"]; ok {
		var n int
		if len(object) != 2 {
			return nil, fmt.Errorf("invalid repeat fields")
		}
		if err := json.Unmarshal(r, &s); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(object["count"], &n); err != nil || n < 0 {
			return nil, fmt.Errorf("invalid repeat count")
		}
		return bytes.Repeat([]byte(s), n), nil
	}
	if p, ok := object["parts"]; ok {
		if len(object) != 1 {
			return nil, fmt.Errorf("invalid parts fields")
		}
		var parts []json.RawMessage
		if err := json.Unmarshal(p, &parts); err != nil {
			return nil, err
		}
		var out []byte
		for _, part := range parts {
			b, err := ExpectedBytes(part)
			if err != nil {
				return nil, err
			}
			out = append(out, b...)
		}
		return out, nil
	}
	return nil, fmt.Errorf("expected string, repeat or parts")
}

func Materialize(t testing.TB, f Fixture) string {
	t.Helper()
	// Retain the fixture's relative layout. A declared ../companions dependency
	// stays inside this allocation, beside the copied case directory.
	dir := filepath.Join(t.TempDir(), filepath.Base(f.Directory))
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	for _, name := range f.Files {
		CopyFile(t, filepath.Join(f.Directory, name), filepath.Join(dir, name), 0600)
	}
	return dir
}
func CopyFile(t testing.TB, from, to string, mode os.FileMode) {
	t.Helper()
	data, err := os.ReadFile(from)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(to), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(to, data, mode); err != nil {
		t.Fatal(err)
	}
}
