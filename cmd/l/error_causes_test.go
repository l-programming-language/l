package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/l-programming-language/l/lerrors"
)

func TestOutputOwnershipPreservesFilesystemCauses(t *testing.T) {
	for _, mode := range []string{"missing marker", "marker directory", "missing generations"} {
		t.Run(mode, func(t *testing.T) {
			output := filepath.Join(t.TempDir(), "main")
			companion := output + ".apps"
			if err := os.Mkdir(companion, 0700); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(companion, ".l-owner")
			want := "refusing unrelated companion directory " + companion
			if mode == "marker directory" {
				if err := os.Mkdir(path, 0700); err != nil {
					t.Fatal(err)
				}
			} else if mode == "missing generations" {
				if err := os.WriteFile(path, []byte(companionOwner+"main\n"), 0600); err != nil {
					t.Fatal(err)
				}
				path = filepath.Join(companion, "generations")
				want = "invalid generation directory in " + companion
			}
			err := validateOutput(output, companion, "main")
			var filesystem *os.PathError
			if err == nil || !errors.Is(err, lerrors.IO) || err.Error() != want || !errors.As(err, &filesystem) || filesystem.Path != path {
				t.Fatalf("lost ownership message or cause: %v", err)
			}
			if mode != "marker directory" && !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("missing not-exist cause: %v", err)
			}
			if _, err := os.Lstat(output); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("validation created output: %v", err)
			}
		})
	}
}
