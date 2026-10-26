package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/l-programming-language/l/lerrors"
)

func TestCargoWorkspaceLockIdentities(t *testing.T) {
	const header = "# Generated local package records; external packages are retained by L.\nversion = 4\n"
	const entry = "\n[[package]]\nname = \"entry\"\nversion = \"0.0.0\"\n"
	for _, tc := range []struct {
		name, local, version, localRef, externalRef string
		retainedBefore, retainedAfter               string
	}{
		{name: "unique names"},
		{name: "direct tokio", local: "tokio", version: "0.0.0", localRef: "tokio 0.0.0", externalRef: "tokio 1.53.1"},
		{name: "direct nix", local: "nix", version: "0.0.0", localRef: "nix 0.0.0", externalRef: "nix 0.31.3"},
		{name: "transitive bytes", local: "bytes", version: "0.0.0", localRef: "bytes 0.0.0", retainedBefore: " \"bytes\",", retainedAfter: " \"bytes 1.12.1\","},
		{name: "transitive libc", local: "libc", version: "0.0.0", localRef: "libc 0.0.0", retainedBefore: " \"libc\",", retainedAfter: " \"libc 0.2.190\","},
		{name: "equal version different source", local: "tokio", version: "1.53.1", localRef: "tokio 1.53.1", externalRef: "tokio 1.53.1 (registry+https://github.com/rust-lang/crates.io-index)"},
		{name: "already qualified transitive", local: "windows-sys", version: "0.61.2", localRef: "windows-sys 0.61.2", retainedBefore: " \"windows-sys 0.61.2\",", retainedAfter: " \"windows-sys 0.61.2 (registry+https://github.com/rust-lang/crates.io-index)\","},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := &cargoPackage{name: "entry", version: "0.0.0", dependencies: []cargoDependency{
				{name: "tokio", packageName: "tokio", version: "1.53.1"},
				{name: "nix", packageName: "nix", version: "0.31.3"},
			}}
			packages := []*cargoPackage{root}
			wantDependencies := " \"nix\",\n \"tokio\",\n"
			wantLocal := ""
			if tc.local != "" {
				root.dependencies = append(root.dependencies, cargoDependency{name: "local_alias", packageName: tc.local, local: "../library"})
				packages = append(packages, &cargoPackage{name: tc.local, version: tc.version, library: "src/lib.rs"})
				// These literal edge expectations are independent of the production encoder.
				switch tc.local {
				case "tokio":
					wantDependencies = " \"nix\",\n \"" + tc.localRef + "\",\n \"" + tc.externalRef + "\",\n"
				case "nix":
					wantDependencies = " \"" + tc.localRef + "\",\n \"" + tc.externalRef + "\",\n \"tokio\",\n"
				case "bytes", "libc":
					wantDependencies = " \"" + tc.localRef + "\",\n" + wantDependencies
				default:
					wantDependencies += " \"" + tc.localRef + "\",\n"
				}
				wantLocal = "\n[[package]]\nname = \"" + tc.local + "\"\nversion = \"" + tc.version + "\"\n"
			}
			retained := cargoDependencyLock
			wantRetained := retained
			if tc.retainedBefore != "" {
				if !strings.Contains(retained, tc.retainedBefore) {
					t.Fatalf("retained fixture lacks expected reference %q", tc.retainedBefore)
				}
				wantRetained = strings.ReplaceAll(retained, tc.retainedBefore, tc.retainedAfter)
			}
			want := header + entry + "dependencies = [\n" + wantDependencies + "]\n" + wantLocal + "\n" + wantRetained
			before := lockPlanSnapshot(packages)
			for iteration := 0; iteration < 2; iteration++ {
				got, err := cargoWorkspaceLock(packages)
				if err != nil || string(got) != want {
					t.Errorf("iteration %d lock = %s\nerror = %v\nwant = %s", iteration, got, err, want)
				}
				if !bytes.Equal(before, lockPlanSnapshot(packages)) || cargoDependencyLock != retained {
					t.Fatal("lock serialization mutated its inputs")
				}
			}
		})
	}

	t.Run("local only alias", func(t *testing.T) {
		packages := []*cargoPackage{
			{name: "entry", version: "0.0.0", dependencies: []cargoDependency{{name: "alias", packageName: "service", local: "../service"}}},
			{name: "service", version: "2.0.0", library: "src/lib.rs"},
		}
		want := header + entry + "dependencies = [\n \"service\",\n]\n\n[[package]]\nname = \"service\"\nversion = \"2.0.0\"\n"
		got, err := cargoWorkspaceLock(packages)
		if err != nil || string(got) != want {
			t.Fatalf("local-only lock = %s, %v; want %s", got, err, want)
		}
	})
	t.Run("no dependencies", func(t *testing.T) {
		got, err := cargoWorkspaceLock([]*cargoPackage{{name: "entry", version: "0.0.0"}})
		if err != nil || string(got) != header+entry {
			t.Fatalf("dependency-free lock = %s, %v", got, err)
		}
	})
}

// JSON sees a test-owned exported projection, not the unexported plan fields.
func lockPlanSnapshot(packages []*cargoPackage) []byte {
	var snapshot []any
	for _, p := range packages {
		var deps []any
		for _, d := range p.dependencies {
			deps = append(deps, []string{d.name, d.version, d.local, d.packageName})
		}
		snapshot = append(snapshot, []any{p.name, p.version, p.binary, p.entry, p.library, p.explicitBinary, p.automaticBinaries, deps})
	}
	data, callErr := json.Marshal(snapshot)
	if callErr != nil {
		panic(callErr)
	}
	return data
}

func TestCargoWorkspaceLockRejectsInvalidLocalIdentities(t *testing.T) {
	for _, tc := range []struct {
		name     string
		packages []*cargoPackage
		cause    string
	}{
		{"missing target", []*cargoPackage{{name: "entry", version: "0.0.0", dependencies: []cargoDependency{{name: "alias", packageName: "missing", local: "../missing"}}}}, "missing"},
		{"duplicate local name", []*cargoPackage{{name: "entry", version: "0.0.0"}, {name: "entry", version: "1.0.0"}}, "entry"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before := lockPlanSnapshot(tc.packages)
			got, err := cargoWorkspaceLock(tc.packages)
			if !errors.Is(err, lerrors.InvalidInput) || !strings.Contains(err.Error(), tc.cause) || got != nil {
				t.Fatalf("lock = %q, error = %v; want InvalidInput mentioning %s and nil output", got, err, tc.cause)
			}
			if !bytes.Equal(before, lockPlanSnapshot(tc.packages)) {
				t.Fatal("invalid input was mutated")
			}
		})
	}
}

func TestCargoWorkspaceLockRejectsMalformedRetainedGraph(t *testing.T) {
	// This serial test replaces the embedded fragment only while its sequential
	// subtests run. Go suspends parallel top-level tests until these finish.
	original := cargoDependencyLock
	t.Cleanup(func() { cargoDependencyLock = original })
	const record = "[[package]]\nname = \"tokio\"\nversion = \"1.53.1\"\nsource = \"registry+https://github.com/rust-lang/crates.io-index\"\nchecksum = \"retained-checksum\"\n"
	for _, tc := range []struct{ name, source, cause string }{
		{"malformed field", strings.Replace(record, "name = \"tokio\"", "name = broken", 1), "name"},
		{"unknown field", record + "unexpected = \"value\"\n", "unexpected"},
		{"missing version", strings.Replace(record, "version = \"1.53.1\"\n", "", 1), "version"},
		{"duplicate field", record + "name = \"again\"\n", "name"},
		{"duplicate identity", record + "\n" + record, "tokio"},
		{"unresolved dependency", record + "dependencies = [\n \"missing\",\n]\n", "missing"},
		{"ambiguous dependency", record + "dependencies = [\n \"tokio\",\n]\n" + strings.Replace(record, "1.53.1", "1.0.0", 1), "tokio"},
		{"unterminated dependencies", record + "dependencies = [\n \"tokio\",\n", "dependencies"},
		{"malformed dependency", record + "dependencies = [\n missing,\n]\n", "dependency"},
		{"missing dependency separator", record + "dependencies = [\n \"tokio\"\n \"tokio\",\n]\n", "comma"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cargoDependencyLock = tc.source
			got, err := cargoWorkspaceLock([]*cargoPackage{{name: "entry", version: "0.0.0", dependencies: []cargoDependency{{name: "tokio", packageName: "tokio", version: "1.53.1"}}}})
			if !errors.Is(err, lerrors.Internal) || !strings.Contains(err.Error(), tc.cause) || got != nil {
				t.Fatalf("lock = %q, error = %v; want Internal mentioning %s and nil output", got, err, tc.cause)
			}
			if cargoDependencyLock != tc.source {
				t.Fatal("retained input was mutated")
			}
		})
	}
}

func TestCargoWorkspaceLockPreservesQualifiedRetainedReferences(t *testing.T) {
	original := cargoDependencyLock
	t.Cleanup(func() { cargoDependencyLock = original })
	const prefix = "# Generated local package records; external packages are retained by L.\nversion = 4\n\n[[package]]\nname = \"entry\"\nversion = \"0.0.0\"\ndependencies = [\n \"tokio\",\n]\n\n"
	for _, tc := range []struct{ name, fragment string }{
		{"qualified reference", strings.Replace(original, " \"bytes\",", " \"bytes 1.12.1 (registry+https://github.com/rust-lang/crates.io-index)\",", 1)},
		{"optional final comma", strings.ReplaceAll(original, ",\n]", "\n]")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cargoDependencyLock = tc.fragment
			got, err := cargoWorkspaceLock([]*cargoPackage{{name: "entry", version: "0.0.0", dependencies: []cargoDependency{{name: "tokio", packageName: "tokio", version: "1.53.1"}}}})
			if err != nil || string(got) != prefix+cargoDependencyLock {
				t.Fatalf("unambiguous retained text was changed: %v\n%s", err, got)
			}
		})
	}
}
