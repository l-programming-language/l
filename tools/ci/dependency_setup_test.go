package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRegistryDependencyBuildRejectsLocalAndIncompleteEvidence(t *testing.T) {
	const prefix = "registry+https://github.com/rust-lang/crates.io-index#"
	allowed := map[string]bool{prefix + "tokio@1.53.1": true, prefix + "nix@0.31.3": true}
	valid := `{"reason":"compiler-artifact","package_id":"` + prefix + `tokio@1.53.1","target":{"name":"tokio","kind":["lib"]}}
{"reason":"compiler-artifact","package_id":"` + prefix + `nix@0.31.3","target":{"name":"nix","kind":["lib"]}}
{"reason":"build-finished","success":true}`
	if err := validateRegistryDependencyBuild([]byte(valid), allowed); err != nil {
		t.Fatal(err)
	}
	for name, bad := range map[string]string{
		"local package":    strings.Replace(valid, prefix+"tokio@1.53.1", "path+file:///tmp/generated#app@0.0.0", 1),
		"unlocked package": strings.Replace(valid, "tokio@1.53.1", "tokio@99.0.0", 1),
		"binary":           strings.Replace(valid, `["lib"]`, `["bin"]`, 1),
		"missing finish":   strings.Split(valid, `{"reason":"build-finished"`)[0],
		"failed finish":    strings.Replace(valid, `"success":true`, `"success":false`, 1),
		"missing library":  strings.Replace(valid, `"name":"tokio"`, `"name":"other"`, 1),
		"trailing event":   valid + "\n" + `{"reason":"build-finished","success":true}`,
		"malformed":        valid + "{",
	} {
		t.Run(name, func(t *testing.T) {
			if err := validateRegistryDependencyBuild([]byte(bad), allowed); err == nil {
				t.Fatal("unsafe dependency build evidence accepted")
			}
		})
	}
}

func TestDependencyBuildSelectsOnlyRetainedRegistryLibraries(t *testing.T) {
	args := strings.Join(registryDependencyBuildArgs("/seed", "/target"), " ")
	for _, required := range []string{"build --frozen --lib", "-p tokio@1.53.1", "-p nix@0.31.3", "--target host-tuple", "--manifest-path /seed/Cargo.toml", "--target-dir /target", "--message-format=json"} {
		if !strings.Contains(args, required) {
			t.Errorf("dependency command lost %q: %s", required, args)
		}
	}
	for _, forbidden := range []string{"--workspace", "--tests", "--bins", "--all-targets", "--release"} {
		if strings.Contains(args, forbidden) {
			t.Fatalf("setup selected project targets or another profile: %s", args)
		}
	}
}

func TestDependencySetupRejectsConflictingInheritedOverrides(t *testing.T) {
	if err := rejectDependencyBuildOverrides([]string{"PATH=/bin", "CARGO_HOME=/registry", "RUSTUP_HOME=/toolchain", "RUSTUP_TOOLCHAIN=stable", "RUSTFLAGS=", "RUSTC_WRAPPER="}); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"RUSTFLAGS", "CARGO_ENCODED_RUSTFLAGS", "RUSTC_WRAPPER", "RUSTC_WORKSPACE_WRAPPER", "RUSTC", "CARGO_BUILD_TARGET", "CARGO_TARGET_DIR", "CARGO_PROFILE_DEV_DEBUG", "CARGO_TARGET_AARCH64_APPLE_DARWIN_RUSTFLAGS"} {
		t.Run(key, func(t *testing.T) {
			err := rejectDependencyBuildOverrides([]string{key + "=private-value"})
			if err == nil || !strings.Contains(err.Error(), key) || strings.Contains(err.Error(), "private-value") {
				t.Fatalf("override rejection must identify only the variable name: %v", err)
			}
		})
	}
	if err := rejectDependencyBuildOverrides([]string{"CARGO_PROFILE_DEV_DEBUG="}); err == nil {
		t.Fatal("empty numeric Cargo setting was treated as unset")
	}
}

func TestDependencyInputsGuardNestedManifestsAndSources(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"Cargo.toml", "Cargo.lock", "helper/Cargo.toml", "observer/Cargo.toml", "observer/main.rs"} {
		testWrite(t, filepath.Join(root, name), []byte("original "+name))
	}
	inputs, err := dependencyInputHashes(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := unchangedDependencyInputs(inputs); err != nil {
		t.Fatal(err)
	}
	testWrite(t, filepath.Join(root, "observer", "Cargo.toml"), []byte("changed"))
	if err := unchangedDependencyInputs(inputs); err == nil || !strings.Contains(err.Error(), "observer") {
		t.Fatalf("nested consumed manifest mutation was ignored: %v", err)
	}
}

func TestFailedDependencyBuildStillChecksEveryCopiedInput(t *testing.T) {
	root := t.TempDir()
	seed := filepath.Join(root, "seed")
	testWrite(t, filepath.Join(seed, "Cargo.toml"), []byte("original"))
	manifest := filepath.Join(seed, "observer", "Cargo.toml")
	testWrite(t, manifest, []byte("original nested"))
	lock := "version = 4\n[[package]]\nname = \"tokio\"\nversion = \"1.53.1\"\nsource = \"registry+https://github.com/rust-lang/crates.io-index\"\nchecksum = \"" + strings.Repeat("a", 64) + "\"\n"
	testWrite(t, filepath.Join(seed, "Cargo.lock"), []byte(lock))
	cargo := filepath.Join(root, "cargo")
	testWrite(t, cargo, []byte("#!/bin/sh\nprintf changed > \"$TEST_DEPENDENCY_MANIFEST\"\nexit 9\n"))
	if err := os.Chmod(cargo, 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TEST_DEPENDENCY_MANIFEST", manifest)
	r := runner{root: root, directory: t.TempDir()}
	err := r.buildRegistryDependencies(seed, filepath.Join(root, "target"), cargo, "/unused-rustc")
	if err == nil || !strings.Contains(err.Error(), "exit 9") || !strings.Contains(err.Error(), "changed retained input") {
		t.Fatalf("child failure hid changed nested input: %v", err)
	}
}

func TestDependencyDestinationsHaveFiniteAcceptanceScope(t *testing.T) {
	targets := dependencyCacheTargets("/repo", "/user-cache", "1.98.1", "darwin", "arm64", "aarch64-apple-darwin")
	if len(targets) != 25 {
		t.Fatalf("setup expanded beyond twenty-five acceptance caches: %d", len(targets))
	}
	seen := map[string]bool{}
	for _, target := range targets {
		if seen[target.Target] || target.Lock == "" || strings.Contains(target.Target, "lucid-rust") {
			t.Fatalf("invalid setup destination: %+v", target)
		}
		seen[target.Target] = true
	}
	for _, path := range []string{"/user-cache/l/rust/1.98.1/darwin-arm64/shard-5/target", "/repo/.harness/cache/rust-parity/lyger/1.98.1/aarch64-apple-darwin/shard-3", "/repo/.harness/cache/rust-parity/logiceme/1.98.1/aarch64-apple-darwin/shard-3", "/repo/.harness/cache/rust-parity/cmd/l/1.98.1/aarch64-apple-darwin/shard-3", "/repo/.harness/cache/rust-parity/acceptance/phase1/runtime/1.98.1/aarch64-apple-darwin/shard-3", "/repo/.harness/cache/rust-empty-error-frontend", "/repo/.harness/cache/rust-empty-error-runtime", "/repo/.harness/cache/rust-runtime-parity"} {
		if !seen[path] {
			t.Errorf("missing observed cold route %s", path)
		}
	}
}
