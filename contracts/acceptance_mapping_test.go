package contracts

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// This is catalog bookkeeping validation, not evidence that the mapped tests
// ran or that their assertions establish more than their documented scope.
func TestAcceptanceMappingIntegrity(t *testing.T) {
	var catalog struct {
		Version string `json:"catalog_version"`
		Cases   []struct {
			ID        string `json:"id"`
			Milestone string `json:"milestone"`
		} `json:"cases"`
	}
	var mapping struct {
		Version  string `json:"catalog_version"`
		Commands map[string]struct {
			Command  string `json:"command"`
			Outcome  string `json:"outcome"`
			Evidence string `json:"evidence"`
		} `json:"commands"`
		Cases []struct {
			ID        string `json:"id"`
			Milestone string `json:"milestone"`
			Status    string `json:"status"`
			Tests     []struct {
				File string `json:"file"`
				Name string `json:"name"`
			} `json:"tests"`
			Commands    []string `json:"command_ids"`
			Limitations []string `json:"limitations"`
		} `json:"cases"`
	}
	for _, document := range []struct {
		path  string
		value any
	}{{"acceptance-catalog.json", &catalog}, {"acceptance-map.json", &mapping}} {
		data, err := os.ReadFile(document.path)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(data, document.value); err != nil {
			t.Fatal(err)
		}
	}
	if len(catalog.Cases) != 110 || len(mapping.Cases) != len(catalog.Cases) || catalog.Version != mapping.Version {
		t.Fatal("mapping must cover the entire versioned 110-case catalog")
	}
	want := make(map[string]string)
	for _, test := range catalog.Cases {
		want[test.ID] = test.Milestone
	}
	seen := make(map[string]bool)
	for _, test := range mapping.Cases {
		t.Run(test.ID, func(t *testing.T) {
			if milestone, exists := want[test.ID]; !exists || milestone != test.Milestone || seen[test.ID] {
				t.Fatal("unknown, duplicate or moved catalog case")
			}
			seen[test.ID] = true
			if !containsString([]string{"covered", "partial", "not_implemented"}, test.Status) {
				t.Fatal("unsupported coverage status")
			}
			if test.Status == "partial" && len(test.Limitations) == 0 {
				t.Fatal("partial coverage must state its remaining requirement")
			}
			if test.Status == "covered" && test.ID != "BASE-001" && len(test.Tests) == 0 {
				t.Fatal("covered compiler behavior requires actual local tests")
			}
			if test.Status == "covered" && len(test.Commands) == 0 {
				t.Fatal("covered case has no executed command record")
			}
			for _, id := range test.Commands {
				command, exists := mapping.Commands[id]
				if !exists || command.Command == "" || command.Evidence == "" || !containsString([]string{"passed", "failed", "blocked", "not_run", "running"}, command.Outcome) {
					t.Fatalf("missing or malformed command evidence %q", id)
				}
			}
			for _, reference := range test.Tests {
				if filepath.IsAbs(reference.File) || strings.HasPrefix(filepath.Clean(reference.File), "..") || !strings.HasSuffix(reference.File, "_test.go") {
					t.Fatalf("invalid repository test path %q", reference.File)
				}
				path := filepath.Join("..", reference.File)
				file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
				if err != nil {
					t.Fatal(err)
				}
				function := strings.SplitN(reference.Name, "/", 2)[0]
				found := false
				for _, declaration := range file.Decls {
					if declared, ok := declaration.(*ast.FuncDecl); ok && declared.Name.Name == function && strings.HasPrefix(function, "Test") {
						found = true
					}
				}
				if !found {
					t.Fatalf("mapped test %s does not exist in %s", reference.Name, reference.File)
				}
			}
		})
	}
}
