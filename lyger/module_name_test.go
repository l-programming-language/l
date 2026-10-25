package Lyger

import (
	"bytes"
	"os"
	"reflect"
	"strings"
	"testing"

	Logiceme "github.com/l-programming-language/l/logiceme"
)

func TestSourceModuleNamesPreserveIdentity(t *testing.T) {
	// These independent spellings cover the shared Object/Class module and
	// Service dependency-alias mapper without interpreting native source.
	cases := []struct{ source, native string }{
		{"ordinary", "ordinary"},
		{"name_1", "name_1"},
		{"move", "__l_name_6d6f7665"},
		{"__l_name_6d6f7665", "__l_name_5f5f6c5f6e616d655f3664366637363635"},
		{"a-b", "__l_name_612d62"},
		{"__l_name_612d62", "__l_name_5f5f6c5f6e616d655f363132643632"},
	}
	seen := make(map[string]string)
	for _, tc := range cases {
		got := sourceModuleName(tc.source)
		if got != tc.native {
			t.Errorf("source %q: native %q, want %q", tc.source, got, tc.native)
		}
		if prior, exists := seen[got]; exists {
			t.Errorf("source identities %q and %q share native name %q", prior, tc.source, got)
		}
		seen[got] = tc.source
	}
}

func TestFrontendEncodedSiblingClassIdentity(t *testing.T) {
	const filename = "test/classes/name-encoding.l"
	source, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	before := bytes.Clone(source)
	cfg := canonicalTestOptions()
	result, err := NewLygerService(cfg).Compile(source, filename)
	if err != nil {
		t.Fatal(err)
	}
	units, err := result.ApplicationUnits()
	if err != nil || len(units) != 1 {
		t.Fatalf("source application units: %d, %v", len(units), err)
	}
	beforeAML := bytes.Clone(units[0].BackendAML)
	service := result.Compilation.Program.Apps[0].Services[0]
	if len(service.Classes) != 2 || service.Classes[0].ID != "move" || service.Classes[1].ID != "__l_name_6d6f7665" || service.Classes[0] == service.Classes[1] {
		t.Fatal("source Class identities were changed or merged")
	}
	bundle, err := Logiceme.NewLogicemeService(cfg.Resources).Evaluate(units[0].BackendAML, Logiceme.GenerationOptions{})
	if err != nil {
		t.Fatalf("distinct permitted Class identities must generate: %v", err)
	}
	wantPaths := []string{"/src/first.rs", "/src/__l_class___l_name_6d6f7665.rs", "/src/last.rs", "/src/__l_class___l_name_5f5f6c5f6e616d655f3664366637363635.rs", "/src/main.rs", "/Cargo.toml", "/src/config.rs", "/start.sh", "/stop.sh"}
	if len(bundle.MapArrays) != 1 || !reflect.DeepEqual(bundle.MapArrays[0], wantPaths) {
		t.Fatalf("source-owned files/order changed: %v", bundle.MapArrays)
	}
	assertClassModules(t, bundle, map[string][]string{
		"/src/main.rs":                        {"mod __l_name_6d6f7665;", "mod __l_name_5f5f6c5f6e616d655f3664366637363635;"},
		"/src/__l_class___l_name_6d6f7665.rs": {"mod first;"},
		"/src/__l_class___l_name_5f5f6c5f6e616d655f3664366637363635.rs": {"mod last;"},
		"/src/first.rs": {"fn first("},
		"/src/last.rs":  {"fn last("},
	})
	if !bytes.Equal(source, before) || !bytes.Equal(units[0].BackendAML, beforeAML) {
		t.Fatal("native naming changed source or original AML")
	}
	again, err := result.ApplicationUnits()
	if err != nil || !bytes.Equal(again[0].BackendAML, beforeAML) {
		t.Fatalf("repeated projection changed source identity: %v", err)
	}
	buildActualBundle(t, bundle, units[0].ID, false, "")
}

func TestWorkspaceEncodedServiceAliasIdentity(t *testing.T) {
	source, err := os.ReadFile("test/workspace/services.l")
	if err != nil {
		t.Fatal(err)
	}
	// Only source-level Service names change. Functions, Objects, Processes,
	// constructed identities and app-specific values keep their existing roles.
	source = bytes.Replace(source, []byte("Generic `left` Service"), []byte("Generic `move` Service"), 1)
	source = bytes.Replace(source, []byte("Generic `right` Service"), []byte("Generic `__l_name_6d6f7665` Service"), 1)
	before := bytes.Clone(source)
	cfg := canonicalTestOptions()
	result, err := NewLygerService(cfg).Compile(source, "service-name-encoding.l")
	if err != nil {
		t.Fatal(err)
	}
	plan, err := result.WorkspacePlan()
	if err != nil {
		t.Fatalf("distinct constructed Services require distinct native aliases: %v", err)
	}
	units, err := plan.Units(workspaceDependencyLocations(plan))
	if err != nil {
		t.Fatal(err)
	}
	if len(units) != 6 || len(result.Compilation.Program.Apps) != 2 {
		t.Fatal("source Service/App package identities were lost")
	}
	for _, index := range []int{0, 3} {
		rows := workspaceDependencyTestRows(t, units[index])
		if len(rows) != 2 || !strings.HasPrefix(rows[0], "__l_name_6d6f7665 = ") || !strings.HasPrefix(rows[1], "__l_name_5f5f6c5f6e616d655f3664366637363635 = ") {
			t.Fatalf("ordered native Service aliases changed: %#v", rows)
		}
	}
	for _, app := range result.Compilation.Program.Apps {
		if app.Services[1].ID != "move" || app.Services[2].ID != "__l_name_6d6f7665" {
			t.Fatal("native aliases replaced original Service names")
		}
	}
	if !bytes.Equal(source, before) {
		t.Fatal("Service alias preparation changed source")
	}
}
