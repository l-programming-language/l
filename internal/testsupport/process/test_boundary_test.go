package process

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	distributiontest "github.com/l-programming-language/l/internal/testsupport/distribution"
)

func TestProcessSupportIsTestOnly(t *testing.T) {
	r := distributiontest.RepoRoot(t)
	err := filepath.WalkDir(r, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path != r && (strings.HasPrefix(entry.Name(), ".") || entry.Name() == "testdata" || entry.Name() == "docs") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imp := range parsed.Imports {
			if strings.Contains(imp.Path.Value, "internal/testsupport/process") {
				t.Errorf("production source imports process test support: %s", path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
