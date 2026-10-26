package utilities

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/l-programming-language/l/lerrors"
)

func TestSourceValidationPreservesFilesystemCauses(t *testing.T) {
	for _, mode := range []string{"missing", "non-directory ancestor"} {
		t.Run(mode, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "missing.l")
			if mode == "non-directory ancestor" {
				ancestor := filepath.Join(dir, "file")
				if err := os.WriteFile(ancestor, []byte("preserved"), 0600); err != nil {
					t.Fatal(err)
				}
				path = filepath.Join(ancestor, "source.l")
			}
			ok, err := CheckIfValidFile(path)
			var filesystem *os.PathError
			if ok || !errors.Is(err, lerrors.IO) || !errors.As(err, &filesystem) || filesystem.Path != path {
				t.Fatalf("lost filesystem failure: valid=%v err=%v", ok, err)
			}
			if mode == "missing" && (!errors.Is(err, os.ErrNotExist) || err.Error() != "file "+path+" does not exist") {
				t.Fatalf("changed missing-file diagnostic: %v", err)
			}
		})
	}
}
