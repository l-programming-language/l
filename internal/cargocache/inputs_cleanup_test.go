package cargocache

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInputTemporaryCleanup(t *testing.T) {
	temporaryPath := filepath.Join(t.TempDir(), ".input-owned")
	primaryCause := errors.New("rename failed")
	primary := &os.LinkError{Op: "rename", Old: temporaryPath, New: "input.rs", Err: primaryCause}
	cleanup := &os.PathError{Op: "remove", Path: temporaryPath, Err: fs.ErrPermission}
	for _, tc := range []struct {
		name    string
		primary error
		cleanup error
		joined  bool
	}{
		{name: "success"},
		{name: "committed-success-ignores-cleanup-failure", cleanup: cleanup},
		{name: "failure-cleanup-succeeds", primary: primary},
		{name: "failure-temporary-absent", primary: primary, cleanup: fs.ErrNotExist},
		{name: "failure-temporary-absence-wrapped", primary: primary, cleanup: fmt.Errorf("remove temporary: %w", fs.ErrNotExist)},
		{name: "failure-retains-cleanup-cause", primary: primary, cleanup: cleanup, joined: true},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			got := cleanupInputTemporary(temporaryPath, tc.primary, func(path string) error {
				calls++
				if path != temporaryPath {
					t.Errorf("cleanup received %q; want owned temporary %q", path, temporaryPath)
				}
				return tc.cleanup
			})
			if calls != 1 {
				t.Fatalf("cleanup calls=%d; want 1", calls)
			}
			if !tc.joined {
				if got != tc.primary {
					t.Fatalf("cleanup changed original result identity: got %v; want %v", got, tc.primary)
				}
				return
			}
			if !errors.Is(got, primary) || !errors.Is(got, primaryCause) || !errors.Is(got, cleanup) || !errors.Is(got, fs.ErrPermission) {
				t.Fatalf("replacement or cleanup cause lost: %v", got)
			}
			var renameError *os.LinkError
			var removeError *os.PathError
			if !errors.As(got, &renameError) || renameError != primary || !errors.As(got, &removeError) || removeError != cleanup {
				t.Fatalf("typed replacement or cleanup cause lost: %v", got)
			}
			if !strings.Contains(got.Error(), "remove Cargo cache temporary input") || !strings.Contains(got.Error(), temporaryPath) {
				t.Fatalf("cleanup operation/path context lost: %v", got)
			}
		})
	}
}

func TestReplaceFilePreservesDestinationOnRenameFailure(t *testing.T) {
	root := t.TempDir()
	destination := filepath.Join(root, "destination")
	if err := os.Mkdir(destination, 0700); err != nil {
		t.Fatal(err)
	}
	retained := filepath.Join(destination, "retained")
	if err := os.WriteFile(retained, []byte("existing destination"), 0600); err != nil {
		t.Fatal(err)
	}
	unrelated := filepath.Join(root, ".input-unrelated")
	if err := os.WriteFile(unrelated, []byte("unrelated input"), 0600); err != nil {
		t.Fatal(err)
	}
	err := replaceFile(destination, []byte("replacement"))
	var renameError *os.LinkError
	if !errors.As(err, &renameError) || renameError.Op != "rename" || renameError.New != destination {
		t.Fatalf("expected destination rename failure, got %v", err)
	}
	if data, err := os.ReadFile(retained); err != nil || string(data) != "existing destination" {
		t.Fatalf("failed replacement changed its destination: %q / %v", data, err)
	}
	if data, err := os.ReadFile(unrelated); err != nil || string(data) != "unrelated input" {
		t.Fatalf("cleanup changed unrelated input: %q / %v", data, err)
	}
	if _, err := os.Lstat(renameError.Old); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("owned failed replacement temporary remains: %v", err)
	}
	temporaries, err := filepath.Glob(filepath.Join(root, ".input-*"))
	if err != nil || len(temporaries) != 1 || temporaries[0] != unrelated {
		t.Fatalf("unexpected temporary inventory after failure: %v / %v", temporaries, err)
	}
}

func TestReplaceFilePreservesExactBytesOnSuccess(t *testing.T) {
	root := t.TempDir()
	destination := filepath.Join(root, "input.rs")
	if err := os.WriteFile(destination, []byte("old bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	want := []byte("first\r\n\x00雪\nlast")
	if err := replaceFile(destination, want); err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(destination); err != nil || !bytes.Equal(got, want) {
		t.Fatalf("replacement bytes changed: %q / %v", got, err)
	}
	temporaries, err := filepath.Glob(filepath.Join(root, ".input-*"))
	if err != nil || len(temporaries) != 0 {
		t.Fatalf("unexpected temporary inventory after success: %v / %v", temporaries, err)
	}
}
