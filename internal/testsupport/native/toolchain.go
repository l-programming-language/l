package native

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

type toolFileIdentity struct {
	path, resolved string
	info           os.FileInfo
}

type toolchainIdentity struct {
	tools ToolchainInfo
	files []toolFileIdentity
}

var toolchainIdentities = struct {
	sync.Mutex
	entries map[string]toolchainIdentity
}{entries: make(map[string]toolchainIdentity)}

func identifyTool(path string) (toolFileIdentity, error) {
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return toolFileIdentity{}, err
	}
	info, err := os.Stat(resolved)
	return toolFileIdentity{path: path, resolved: resolved, info: info}, err
}

func (identity toolchainIdentity) unchanged() bool {
	for _, before := range identity.files {
		after, err := identifyTool(before.path)
		if err != nil || after.resolved != before.resolved || !os.SameFile(before.info, after.info) ||
			after.info.Size() != before.info.Size() || after.info.Mode() != before.info.Mode() ||
			!after.info.ModTime().Equal(before.info.ModTime()) {
			return false
		}
	}
	return true
}

// Toolchain memoizes only read-only tool identity queries. Every Cargo build
// still executes and supplies its own source and compiler-artifact evidence.
func Toolchain(t testing.TB) ToolchainInfo {
	t.Helper()
	rustup, err := exec.LookPath("rustup")
	if err != nil {
		t.Fatalf("pinned Rust toolchain prerequisite: %v", err)
	}
	rustup, err = filepath.Abs(rustup)
	if err != nil {
		t.Fatal(err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	environment := os.Environ()
	sort.Strings(environment)
	key := strings.Join(append([]string{rustup, cwd, Version}, environment...), "\x00")
	toolchainIdentities.Lock()
	defer toolchainIdentities.Unlock()
	if prior, found := toolchainIdentities.entries[key]; found && prior.unchanged() {
		return prior.tools
	}
	rustupIdentity, err := identifyTool(rustup)
	if err != nil {
		t.Fatalf("read native tool identity: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	locate := func(name string) string {
		b, e := exec.CommandContext(ctx, rustup, "which", "--toolchain", Version, name).CombinedOutput()
		if e != nil {
			t.Fatalf("pinned Rust toolchain prerequisite (%s): %v\n%s", name, e, b)
		}
		return strings.TrimSpace(string(b))
	}
	tc := ToolchainInfo{Cargo: locate("cargo"), Rustc: locate("rustc")}
	identity := toolchainIdentity{files: []toolFileIdentity{rustupIdentity}}
	for _, path := range []string{tc.Cargo, tc.Rustc} {
		file, err := identifyTool(path)
		if err != nil {
			t.Fatalf("retain native tool identity: %v", err)
		}
		identity.files = append(identity.files, file)
	}
	b, e := exec.CommandContext(ctx, tc.Rustc, "-vV").CombinedOutput()
	if e != nil {
		t.Fatalf("rustc identity: %v\n%s", e, b)
	}
	for _, line := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(line, "host: ") {
			tc.Host = strings.TrimPrefix(line, "host: ")
		}
	}
	if tc.Host == "" || !bytes.Contains(b, []byte("release: "+Version)) {
		t.Fatalf("wrong Rust toolchain: %s", b)
	}
	if !identity.unchanged() {
		t.Fatal("native toolchain changed while querying its identity")
	}
	identity.tools = tc
	toolchainIdentities.entries[key] = identity
	return tc
}
