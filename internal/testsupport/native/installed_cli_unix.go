//go:build darwin || linux

package native

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"

	distributiontest "github.com/l-programming-language/l/internal/testsupport/distribution"
	"github.com/l-programming-language/l/internal/timing"
)

var installed struct {
	once                  sync.Once
	directory, executable string
	err                   error
}

func InstalledCLI(t testing.TB) string {
	t.Helper()
	installed.once.Do(func() {
		finish := timing.Start("test.cli.setup", t.Name())
		defer func() { finish(installed.err) }()
		installed.directory, installed.err = os.MkdirTemp("", "l-installed-cli-")
		if installed.err != nil {
			return
		}
		installed.executable = filepath.Join(installed.directory, "l")
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cmd := exec.CommandContext(ctx, "go", "build", "-mod=readonly", "-o", installed.executable, "./cmd/l")
		cmd.Dir = distributiontest.RepoRoot(t)
		output, err := cmd.CombinedOutput()
		if err != nil {
			installed.err = fmt.Errorf("installed CLI build: %w\n%s", err, output)
			return
		}
	})
	if installed.err != nil {
		t.Fatalf("installed-cli prerequisite; runtime not exercised: %v", installed.err)
	}
	return installed.executable
}

// Main removes immutable package-owned CLI build products after the package.
func Main(m *testing.M) {
	code := m.Run()
	if installed.directory != "" {
		_ = os.RemoveAll(installed.directory) //nolint:errcheck // Post-suite temporary executable cleanup must preserve the test process exit status.
	}
	os.Exit(code)
}

// CopyCLI installs the package-owned test executable at a caller-selected path.
func CopyCLI(t testing.TB, destination string) {
	t.Helper()
	data, err := os.ReadFile(InstalledCLI(t))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(destination, data, 0700); err != nil {
		t.Fatal(err)
	}
}
