package Lyger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Syntax-only tests use the canonical checked-in language definitions. Tests of
// embedded defaults deliberately leave AssetRoot empty instead.
func canonicalTestOptions() Options {
	return Options{Language: "eng", AssetRoot: "..", Resources: os.DirFS("..")}
}

// Tests of user-file inclusion compare that projection explicitly; the complete
// compilation also contains implicit builtin and imported dictionary units.
func testPhysicalStatements(root string, statements []SourceStatement) []SourceStatement {
	var result []SourceStatement
	for _, statement := range statements {
		if testPathWithin(root, statement.Span.Path) {
			result = append(result, statement)
		}
	}
	return result
}

func testPhysicalOccurrences(root string, occurrences []sourceOccurrence) []sourceOccurrence {
	var result []sourceOccurrence
	for _, occurrence := range occurrences {
		if testPathWithin(root, occurrence.Path) {
			result = append(result, occurrence)
		}
	}
	return result
}

func testPathWithin(root, name string) bool {
	if !filepath.IsAbs(name) {
		return false
	}
	relative, err := filepath.Rel(root, name)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func testProgramPaths(paths []string, root string) []string {
	var result []string
	for _, name := range paths {
		if testPathWithin(root, name) {
			result = append(result, name)
		}
	}
	return result
}
func testProgramStatements(statements []*lStatement) []*lStatement {
	var result []*lStatement
	for _, statement := range statements {
		if statement.source.Unit == nil || statement.source.Unit.Kind == SourceProgram {
			result = append(result, statement)
		}
	}
	return result
}
func testProgramSemanticCounts(c *Compilation) (subjects, declarations, descriptors int) {
	for _, source := range c.Semantics.Declarations {
		if source.source.Unit == nil || source.source.Unit.Kind == SourceProgram {
			declarations++
		}
	}
	for token := range c.Semantics.Subjects {
		if token.scope.unit == nil || token.scope.unit.Kind == SourceProgram {
			subjects++
		}
	}
	for record := range c.Semantics.Descriptors {
		if record.storageScope.unit == nil || record.storageScope.unit.Kind == SourceProgram {
			descriptors++
		}
	}
	return
}

func assertFrontendResources(t *testing.T, c *Compilation) {
	t.Helper()
	if c == nil {
		return
	}
	for _, unit := range c.SourceUnits {
		if unit.Kind == SourceBuiltin && unit.Origin.Name != "builtins/"+c.Gram.language+"/core.l" {
			t.Errorf("Rust source selection: builtin origin=%q; expected configured-language/core.l", unit.Origin.Name)
		}
		if unit.Kind == SourceDictionary && unit.Dictionary == "IO" && unit.Origin.Name != "stdlib/"+c.Gram.language+"/io/dictionary.l" {
			t.Errorf("Rust source selection: IO origin=%q; expected stdlib/<language>/io/dictionary.l", unit.Origin.Name)
		}
	}
	if c.Program == nil {
		return
	}
	for _, app := range c.Program.Apps {
		if app.Meta["AppLang"] != "rust" {
			t.Errorf("Rust application metadata: %s AppLang=%q", app.ID, app.Meta["AppLang"])
		}
	}
}
