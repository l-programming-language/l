package Lyger

import (
	"bytes"
	"reflect"
	"strconv"
	"testing"
)

func TestParagraphIdentityArtifactCollisionReservation(t *testing.T) {
	var firstIDs []string
	for attempt := 0; attempt < 2; attempt++ {
		source := paragraphSource(t)
		source.main = paragraphGlobals + "Create \"Worker_1\" as a New Basic App.\nThe App Prints x1.\n"
		for _, text := range []string{"first private", "second private"} {
			source.main += "\n    \"message\" is `" + text + "`.\nCreate \"Worker\" as a New Basic App.\nThe App Prints message.\n"
		}
		compilation := executableTestResult(t, source).Compilation
		workers := paragraphBindingDeclarations(t, compilation, "Worker", 2)
		for _, worker := range workers {
			if worker.record.ID != "Worker" || worker.record.Value.Applicative.App.ID != "Worker" {
				t.Fatal("artifact naming changed a source declaration ID")
			}
		}
		units, err := (&CompilationResult{Compilation: compilation}).ApplicationUnits()
		if err != nil {
			t.Fatal(err)
		}
		if len(units) != 3 || units[0].ID != "Worker_1" {
			t.Fatalf("units lost unique source name: %+v", units)
		}
		var ids []string
		seen := make(map[string]bool)
		for index, unit := range units {
			if seen[unit.ID] || unit.ID == "" {
				t.Fatalf("duplicate artifact %q", unit.ID)
			}
			seen[unit.ID] = true
			ids = append(ids, unit.ID)
			if !bytes.Contains(unit.BackendAML, []byte("<BOOK "+strconv.Quote(unit.ID)+">")) || !bytes.Contains(unit.BackendAML, []byte(`"AppName":`+strconv.Quote(unit.ID))) {
				t.Fatalf("unit %q disagrees with its book or application metadata", unit.ID)
			}
			if index > 0 {
				text := []string{"first private", "second private"}[index-1]
				if !bytes.Contains(unit.BackendAML, []byte(text)) {
					t.Fatalf("unit %s lost source-ordered body %q", unit.ID, text)
				}
			}
		}
		if attempt == 0 {
			firstIDs = ids
		} else if !reflect.DeepEqual(ids, firstIDs) {
			t.Fatalf("artifact names changed across compilations: %v vs %v", ids, firstIDs)
		}
	}
}

func TestParagraphIdentityApplicationOrderIgnoresResolutionOrder(t *testing.T) {
	source := paragraphSource(t)
	source.main = paragraphGlobals + "Create \"Earlier\" as a New Future App.\nThe App Prints x1.\nCreate \"Later\" as a New Basic App.\nThe App Prints x2.\n\"Future\" describes a Generic App with a New Main Service using any Process.\n"
	compilation := executableTestResult(t, source).Compilation
	units, err := (&CompilationResult{Compilation: compilation}).ApplicationUnits()
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, unit := range units {
		ids = append(ids, unit.ID)
	}
	if !reflect.DeepEqual(ids, []string{"Earlier", "Later"}) {
		t.Fatalf("application order follows resolution: %v", ids)
	}
}

func TestParagraphIdentityCoordinatedAppConstructionSource(t *testing.T) {
	source := paragraphSource(t)
	source.main = paragraphGlobals + "\"First\" is a New Basic App and \"Second\" is a New Basic App.\nFirst Prints x1.\nSecond Prints x2.\n"
	compilation := executableTestResult(t, source).Compilation
	declarations := paragraphBindingDeclarations(t, compilation, "Second", 1)
	statement := declarations[0].statement.sGram
	if len(statement.clauses) != 1 {
		t.Fatal("test requires a second passive declaration")
	}
	want := statement.clauses[0].predicate.obj
	for _, construction := range compilation.Semantics.Constructions {
		if construction.App == declarations[0].record.Value.Applicative.App {
			if construction.Root.Source != want {
				t.Fatal("second App reused the first declaration's construction AST")
			}
			return
		}
	}
	t.Fatal("second App has no construction")
}

func TestParagraphIdentityRootAppKeepsBuildName(t *testing.T) {
	source := paragraphSource(t)
	source.main = paragraphGlobals + "Create \"Worker\" as a New Basic App.\nThe App Prints x1.\n\n    Create \"Worker\" as a New Basic App.\nThe App Prints x2.\n"
	compilation := executableTestResult(t, source).Compilation
	units, err := (&CompilationResult{Compilation: compilation}).ApplicationUnits()
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, unit := range units {
		ids = append(ids, unit.ID)
	}
	if len(units) != 2 || units[0].ID != "Worker" || units[1].ID == "Worker" {
		t.Fatalf("private collision renamed the root App: %v", ids)
	}
	if compilation.Memory["Worker"].Value.Applicative.App.buildName() != "Worker" {
		t.Fatal("published App no longer keeps its build name")
	}
}
