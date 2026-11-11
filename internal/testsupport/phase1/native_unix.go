//go:build darwin || linux

package phase1

import (
	"context"

	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"sync"

	"testing"

	"github.com/l-programming-language/l/internal/timing"
)

var installed struct {
	once                  sync.Once
	directory, executable string
	err                   error
}

func CLI(t testing.TB) string {
	t.Helper()
	installed.once.Do(func() {
		finish := timing.Start("test.cli.setup", t.Name())
		defer func() { finish(installed.err) }()
		installed.directory, installed.err = os.MkdirTemp("", "l-phase1-installed-")
		if installed.err != nil {
			return
		}
		installed.executable = filepath.Join(installed.directory, "l")
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cmd := exec.CommandContext(ctx, "go", "build", "-mod=readonly", "-o", installed.executable, "./cmd/l")
		cmd.Dir = RepoRoot(t)
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
