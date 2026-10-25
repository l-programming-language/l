//go:build darwin || linux

package cli

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	"github.com/l-programming-language/l/internal/testsupport/phase1"
)

func TestMain(m *testing.M) { phase1.Main(m) }

func write(t testing.TB, path, source string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
}
func read(t testing.TB, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
func project(t testing.TB, source string, files map[string]string) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(dir, "main.l"), source)
	for path, source := range files {
		write(t, filepath.Join(dir, path), source)
	}
	return dir
}
func outcome(t testing.TB, binary, cwd string, env []string, status int, stdout, stderr string) {
	t.Helper()
	out, diag, err := nativetest.RuntimeExecution.Run(t, binary, cwd, env)
	code := 0
	if err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) {
			t.Fatal(err)
		}
		code = exit.ExitCode()
	}
	if code != status || string(out) != stdout || string(diag) != stderr {
		t.Fatalf("execution status=%d stdout=%q stderr=%q; want %d %q %q", code, out, diag, status, stdout, stderr)
	}
}
func standalone(t testing.TB) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "compiler")
	phase1.CopyFile(t, phase1.CLI(t), path, 0700)
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil || len(entries) != 1 {
		t.Fatalf("standalone setup=%v: %v", entries, err)
	}
	return path
}
