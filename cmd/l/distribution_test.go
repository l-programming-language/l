package main

import (
	"testing"

	distributiontest "github.com/l-programming-language/l/internal/testsupport/distribution"
	"github.com/l-programming-language/l/internal/testsupport/phase1"
)

func TestMain(m *testing.M) { phase1.Main(m) }

// Explicit disk-resource controls use today's real initialization API. They do
// not implement embedded defaults or supply a replacement builtin/library.
func canonicalCompilerOptions(t testing.TB) compilerInitOptions {
	t.Helper()
	root := distributiontest.RepoRoot(t)
	return compilerInitOptions{AssetRoot: root}
}
