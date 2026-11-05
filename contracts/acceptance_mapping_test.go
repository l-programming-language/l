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
			ID string `json:"id"`
		} `json:"cases"`
	}
	var mapping struct {
		Version string `json:"catalog_version"`
		Cases   []struct {
			ID     string `json:"id"`
			Status string `json:"status"`
			Tests  []struct {
				File string `json:"file"`
				Name string `json:"name"`
			} `json:"tests"`
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
	if len(catalog.Cases) == 0 || len(mapping.Cases) != len(catalog.Cases) || catalog.Version == "" || catalog.Version != mapping.Version {
		t.Fatal("mapping must cover the entire versioned catalog")
	}
	want := make(map[string]bool)
	for _, test := range catalog.Cases {
		if test.ID == "" || want[test.ID] {
			t.Fatalf("empty or duplicate catalog ID %q", test.ID)
		}
		want[test.ID] = true
	}
	seen := make(map[string]bool)
	functions := make(map[string]map[string]bool)
	for _, test := range mapping.Cases {
		t.Run(test.ID, func(t *testing.T) {
			if !want[test.ID] || seen[test.ID] {
				t.Fatal("unknown or duplicate catalog case")
			}
			seen[test.ID] = true
			if !containsString([]string{"covered", "partial", "not_implemented"}, test.Status) {
				t.Fatal("unsupported coverage status")
			}
			if test.Status == "partial" && len(test.Limitations) == 0 {
				t.Fatal("partial coverage must state its remaining requirement")
			}
			if test.Status == "covered" && len(test.Tests) == 0 {
				t.Fatal("covered compiler behavior requires actual local tests")
			}
			for _, reference := range test.Tests {
				if filepath.IsAbs(reference.File) || strings.HasPrefix(filepath.Clean(reference.File), "..") || !strings.HasSuffix(reference.File, "_test.go") {
					t.Fatalf("invalid repository test path %q", reference.File)
				}
				if functions[reference.File] == nil {
					file, err := parser.ParseFile(token.NewFileSet(), filepath.Join("..", reference.File), nil, 0)
					if err != nil {
						t.Fatal(err)
					}
					functions[reference.File] = make(map[string]bool)
					for _, declaration := range file.Decls {
						if declared, ok := declaration.(*ast.FuncDecl); ok && declared.Recv == nil && strings.HasPrefix(declared.Name.Name, "Test") {
							functions[reference.File][declared.Name.Name] = true
						}
					}
				}
				function := strings.SplitN(reference.Name, "/", 2)[0]
				if !functions[reference.File][function] {
					t.Fatalf("mapped test %s does not exist in %s", reference.Name, reference.File)
				}
			}
		})
	}
	for id := range want {
		if !seen[id] {
			t.Errorf("unmapped catalog case %s", id)
		}
	}
}
