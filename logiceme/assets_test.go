package Logiceme

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	distribution "github.com/l-programming-language/l"
	"github.com/l-programming-language/l/lerrors"
)

// Isolate the original public APIs so an accidental fatal exit is an observed
// failure, rather than a successful early end to the parent package test run.
func TestBackendAssetErrorsReturn(t *testing.T) {
	const environment = "L_BACKEND_ASSET_ERROR_CHILD"
	const complete = "asset error returned to caller"
	if action := os.Getenv(environment); action != "" {
		var err error
		switch action {
		case "read":
			_, err = ReadSysFile("absent.lu")
		case "segment":
			segment := Segment{Name: "absent", SubType: "vowel"}
			err = segment.loadTemplate(os.DirFS("."))
		case "evaluate segment", "evaluate templates":
			missing := "builtins/segments/consonants/execute.lu"
			if action == "evaluate templates" {
				missing = "builtins/templates/config/config.lu"
			}
			_, err = NewLogicemeService(missingAssetFS{distribution.Files(), missing}).Evaluate(nil)
		default:
			t.Fatalf("unknown test action %q", action)
		}
		if err == nil || !errors.Is(err, lerrors.IO) || !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("asset error lost its filesystem cause: %v", err)
		}
		fmt.Println(complete)
		return
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, action := range []string{"read", "segment", "evaluate segment", "evaluate templates"} {
		t.Run(action, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, executable, "-test.run=^TestBackendAssetErrorsReturn$")
			command.Dir = t.TempDir()
			command.Env = append(os.Environ(), environment+"="+action)
			var stdout, stderr bytes.Buffer
			command.Stdout, command.Stderr = &stdout, &stderr
			err := command.Run()
			if err != nil || !strings.HasPrefix(stdout.String(), complete+"\n") {
				t.Fatalf("asset API terminated instead of returning: %v\nstdout: %s\nstderr: %s", err, stdout.Bytes(), stderr.Bytes())
			}
			if stderr.Len() != 0 {
				t.Fatalf("asset loader wrote diagnostics instead of returning its error: %s", stderr.Bytes())
			}
		})
	}
}

func TestBackendAssetReadPreservesBytes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "template.lu")
	want := []byte("raw \\n 雪\n")
	if err := os.WriteFile(path, want, 0600); err != nil {
		t.Fatal(err)
	}
	got, err := ReadSysFile(path)
	if err != nil || !bytes.Equal(got, want) {
		t.Fatalf("asset content = %q, %v; want %q", got, err, want)
	}
}

// Hide one selected resource without allowing an embedded fallback.
type missingAssetFS struct {
	fs.FS
	missing string
}

func (m missingAssetFS) Open(name string) (fs.File, error) {
	if name == m.missing {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	return m.FS.Open(name)
}
