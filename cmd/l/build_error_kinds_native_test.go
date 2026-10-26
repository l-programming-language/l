//go:build darwin || linux

package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/l-programming-language/l/lerrors"
)

func TestBuildBoundaryErrorKinds(t *testing.T) {
	t.Run("bundle arguments", func(t *testing.T) {
		_, err := planSourceBundle([][]byte{[]byte("fn main() {}")}, nil)
		if !errors.Is(err, lerrors.InvalidInput) || err.Error() != "source bundle dimensions do not match" {
			t.Fatalf("bundle error = %v", err)
		}
	})
	t.Run("generated syntax", func(t *testing.T) {
		bundle, err := plainSourceBundle("fn main( {\n")
		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
			defer cancel()
			err = compileCargoWorkspace(ctx, nil, bundle, nil, nil, filepath.Join(t.TempDir(), "src"), nil, filepath.Join(t.TempDir(), "out"))
		}
		var native *exec.ExitError
		if !errors.Is(err, lerrors.Internal) || !errors.As(err, &native) || native.ExitCode() == 0 {
			t.Fatalf("generated Rust syntax lost classified native compiler cause: %v", err)
		}
	})
	t.Run("unsupported dependency", func(t *testing.T) {
		_, err := planSourceBundle([][]byte{[]byte(cargoFixtureManifest + "\n[dependencies]\nexternal=\"999.0.0\"\n"), []byte("fn main() {}")}, []string{"/Cargo.toml", "/src/main.rs"})
		if !errors.Is(err, lerrors.Unsupported) {
			t.Fatalf("dependency error = %v", err)
		}
	})
	t.Run("materialization", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "main.rs")
		if err := os.WriteFile(path, []byte("preserve"), 0600); err != nil {
			t.Fatal(err)
		}
		err := createSourceFile(path, []byte("replacement"))
		var native *os.PathError
		if !errors.Is(err, lerrors.IO) || !errors.Is(err, os.ErrExist) || !errors.As(err, &native) || native.Path != path {
			t.Fatalf("materialization error = %v", err)
		}
		if got, err := os.ReadFile(path); err != nil || string(got) != "preserve" {
			t.Fatalf("overwrote source: %q %v", got, err)
		}
	})
	t.Run("inspected ownership", func(t *testing.T) {
		output := filepath.Join(t.TempDir(), "main")
		companion := output + ".apps"
		if err := os.Mkdir(companion, 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(companion, ".l-owner"), []byte("unrelated"), 0600); err != nil {
			t.Fatal(err)
		}
		err := validateOutput(output, companion, "main")
		if !errors.Is(err, lerrors.InvalidInput) || errors.Is(err, lerrors.IO) || err.Error() != "refusing unrelated companion directory "+companion {
			t.Fatalf("ownership error = %v", err)
		}
	})
	t.Run("compiler input", func(t *testing.T) {
		compiler := newCompiler(nil)
		compiler.initialize()
		if !errors.Is(compiler.err, lerrors.InvalidInput) || compiler.err.Error() != "compiler input is required" {
			t.Fatalf("compiler input = %v", compiler.err)
		}
	})
}
