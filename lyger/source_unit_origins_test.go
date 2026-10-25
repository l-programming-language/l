package Lyger

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	distributiontest "github.com/l-programming-language/l/internal/testsupport/distribution"
)

func TestSourceUnitOriginsAndPhysicalProjection(t *testing.T) {
	for _, override := range []bool{false, true} {
		t.Run(fmt.Sprint(override), func(t *testing.T) {
			cfg := Options{}
			if override {
				root := distributiontest.CompleteOverride(t)
				cfg = Options{Resources: os.DirFS(root), AssetRoot: root}
			}
			project := t.TempDir()
			main := filepath.Join(project, "main.l")
			included := writeSourceFile(t, project, "nested/local.l", "\"value\" is `snow 雪`.\r\nShow value.\r\n")
			source := "Use IO's Print from `io` as \"FirstPrint\".\r\nUse IO's Print from `io` as \"SecondPrint\".\r\n\r\n    Use `nested/local.l`.\r\n\r\n    Use `nested/local.l`.\r\n"
			if err := os.WriteFile(main, []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			result, err := NewLygerService(cfg).Compile([]byte(source), main)
			if err != nil {
				t.Fatal(err)
			}
			c := result.Compilation
			assertFrontendResources(t, c)
			if !bytes.Equal(c.MainBytes, []byte(source)) || !bytes.Equal(c.Prepared.Original, []byte(source)) {
				t.Fatal("unit preparation changed original main bytes")
			}
			if len(c.SourceUnits) != 3 || len(c.SourceOrigins) != 5 {
				t.Fatalf("units/origins=%d/%d; want3/5", len(c.SourceUnits), len(c.SourceOrigins))
			}
			units := map[SourceKind]*SourceUnit{}
			for _, unit := range c.SourceUnits {
				if units[unit.Kind] != nil {
					t.Fatalf("duplicate %s unit", unit.Kind)
				}
				units[unit.Kind] = unit
				checkSourceOrigin(t, c, unit.Origin, override)
			}
			if units[SourceProgram].root != c.Scope.Root || c.Scope.Root.records == nil || c.Scope.Root.unit != units[SourceProgram] || c.Scope.Root.parent != units[SourceBuiltin].root || units[SourceDictionary].root.parent != units[SourceBuiltin].root {
				t.Fatal("lexical unit roots lost program Memory or builtin inheritance")
			}
			if !reflect.DeepEqual(c.Memory, c.Scope.Root.records) || c.Scope.Root.isParagraph() || units[SourceBuiltin].root.isParagraph() || units[SourceDictionary].root.isParagraph() {
				t.Fatal("unit roots became activation frames")
			}
			if c.SourceUnits[0].Kind != SourceBuiltin || c.SourceUnits[1].Kind != SourceDictionary || c.SourceUnits[2].Kind != SourceProgram {
				t.Fatal("source unit inventory lost dependency order")
			}
			var paths []string
			for _, origin := range c.SourceOrigins {
				checkSourceOrigin(t, c, origin, override)
				if origin.PhysicalPath != "" {
					paths = append(paths, origin.PhysicalPath)
				}
			}
			if !reflect.DeepEqual(c.SourcePaths, paths) {
				t.Fatalf("physical projection=%v; origins=%v", c.SourcePaths, paths)
			}
			if !override && !reflect.DeepEqual(paths, []string{main, included, included}) {
				t.Fatalf("embedded identities became physical dependencies: %v", paths)
			}
			var locals []sourceOccurrence
			for _, occurrence := range c.SourceOccurrences {
				if occurrence.Path == included {
					locals = append(locals, occurrence)
				}
			}
			if len(locals) != 2 || locals[0].ID == locals[1].ID || locals[0].Parent != locals[1].Parent || locals[0].ImportSite == locals[1].ImportSite || locals[0].Unit != units[SourceProgram] || locals[1].Unit != units[SourceProgram] || locals[0].Origin != locals[1].Origin {
				t.Fatalf("local inclusion lost shared provider or separate occurrence ancestry: %+v", locals)
			}
			for _, statement := range c.SourceStatements {
				if statement.Unit == nil || statement.Origin.Name != statement.Span.Path {
					t.Fatal("statement lost typed unit/origin")
				}
			}
		})
	}
}

func checkSourceOrigin(t *testing.T, c *Compilation, origin SourceOrigin, override bool) {
	t.Helper()
	var data []byte
	var err error
	if strings.HasPrefix(origin.Identity, "resource:") {
		data, err = fs.ReadFile(c.Gram.resources, origin.Name)
		if override {
			want := filepath.Join(c.Gram.assetRoot, filepath.FromSlash(origin.Name))
			if origin.PhysicalPath != want {
				t.Fatalf("override physical path=%q; want%q", origin.PhysicalPath, want)
			}
		} else if origin.PhysicalPath != "" || filepath.IsAbs(origin.Name) {
			t.Fatalf("embedded origin is physical: %+v", origin)
		}
	} else {
		data, err = os.ReadFile(origin.PhysicalPath)
		if !filepath.IsAbs(origin.PhysicalPath) {
			t.Fatalf("program origin is not physical: %+v", origin)
		}
	}
	if err != nil {
		t.Fatal(err)
	}
	if want := fmt.Sprintf("%x", sha256.Sum256(data)); origin.SHA256 != want {
		t.Fatalf("origin content identity=%s; want%s for %s", origin.SHA256, want, origin.Name)
	}
}

func TestSourceUnitDiamondIdentity(t *testing.T) {
	files := dictionaryIndexFilesystem(t)
	var index dictionaryIndex
	if err := json.Unmarshal(files["stdlib/index.json"].Data, &index); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"Left", "Right"} {
		key := strings.ToLower(name)
		entry := "stdlib/eng/" + key + "/dictionary.l"
		index.Languages["eng"][key] = dictionaryEntry{Entry: entry, Dictionary: name}
		files[entry] = &fstest.MapFile{Data: []byte("Use IO's Print from `io` as \"Imported\".\nCreate \"" + name + "\" as a Dictionary.\nUse Imported as " + name + "'s \"Print\".\n")}
	}
	raw, err := json.Marshal(index)
	if err != nil {
		t.Fatal(err)
	}
	files["stdlib/index.json"].Data = raw
	cfg := Options{Resources: files}
	source := "Use Left's Print from `left` as \"FirstPrint\".\nUse Right's Print from `right` as \"SecondPrint\".\n"
	result, err := NewLygerService(cfg).Compile([]byte(source), filepath.Join(t.TempDir(), "main.l"))
	if err != nil {
		t.Fatal(err)
	}
	c := result.Compilation
	assertFrontendResources(t, c)
	if len(c.SourceUnits) != 5 {
		t.Fatalf("diamond units=%d; want5", len(c.SourceUnits))
	}
	var ioUnit *SourceUnit
	for _, unit := range c.SourceUnits {
		if unit.Dictionary == "IO" {
			if ioUnit != nil {
				t.Fatal("diamond loaded IO twice")
			}
			ioUnit = unit
		}
	}
	if ioUnit == nil || distributionCanonical(t, c, c.Memory["FirstPrint"]) != distributionCanonical(t, c, c.Memory["SecondPrint"]) {
		t.Fatal("diamond exports lost canonical IO definition")
	}
	declarations, edges := 0, 0
	for _, statement := range c.SourceStatements {
		if statement.Unit == ioUnit && strings.Contains(string(statement.Raw), "Create IO's") {
			declarations++
		}
		if statement.imported != nil && statement.imported.unit == ioUnit {
			edges++
		}
	}
	if declarations != 1 || edges != 2 {
		t.Fatalf("IO definition/import edges=%d/%d; want1/2", declarations, edges)
	}
}
