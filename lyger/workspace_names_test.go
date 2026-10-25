package Lyger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	Logiceme "github.com/l-programming-language/l/logiceme"
)

func compileWorkspaceNames(t *testing.T, source []byte) *CompilationResult {
	t.Helper()
	before := bytes.Clone(source)
	cfg := canonicalTestOptions()
	result, err := NewLygerService(cfg).Compile(source, "workspace-names.l")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(source, before) {
		t.Fatal("compilation changed source bytes")
	}
	return result
}

func workspaceNameSource(names ...string) []byte {
	var source strings.Builder
	for _, name := range names {
		fmt.Fprintf(&source, "Create %q as a New Basic App.\n\n", name)
	}
	return []byte(source.String())
}

func assertWorkspaceNativeName(t *testing.T, result *CompilationResult, unit WorkspaceUnit, name string) {
	t.Helper()
	assertWorkspaceNativeIdentities(t, result, unit, name, name)
}

func assertWorkspaceNativeIdentities(t *testing.T, result *CompilationResult, unit WorkspaceUnit, name, targetName string) {
	t.Helper()
	if !bytes.Contains(unit.BackendAML, []byte("<BOOK "+strconv.Quote(name)+">")) ||
		!bytes.Contains(unit.BackendAML, []byte(`"AppName":`+strconv.Quote(name))) {
		t.Errorf("canonical unit %q requires native Book/AppMeta name %q", unit.ID, name)
	}
	for _, line := range strings.Split(string(unit.BackendAML), "\n") {
		if !strings.HasPrefix(line, `<JSON>{"Services":`) {
			continue
		}
		var record struct{ Services []amlServiceMetadata }
		if err := json.Unmarshal([]byte(strings.TrimSuffix(strings.TrimPrefix(line, "<JSON>"), "</JSON>")), &record); err != nil {
			t.Fatal(err)
		}
		for _, service := range record.Services {
			if service.Target != nil && (service.Target.Primary || unit.Library) && service.Target.Name != targetName {
				t.Errorf("canonical unit %q target = %q; want %q", unit.ID, service.Target.Name, targetName)
			}
		}
	}
	bundle, err := Logiceme.NewLogicemeService(result.Compilation.Gram.resources).Evaluate(unit.BackendAML)
	if err != nil {
		t.Fatalf("generate canonical unit %q: %v", unit.ID, err)
	}
	manifest := ""
	for i, path := range bundle.MapArrays[0] {
		if path == "/Cargo.toml" {
			manifest = string(bundle.ByteArrays[0][i])
		}
	}
	if !strings.Contains(manifest, "[package]\nname = "+strconv.Quote(name)+"\n") {
		t.Errorf("canonical unit %q Cargo package must be %q:\n%s", unit.ID, name, manifest)
	}
	target := "[[bin]]"
	if unit.Library {
		target = "[lib]"
	}
	if !strings.Contains(manifest, target+"\nname = "+strconv.Quote(targetName)+"\n") {
		t.Errorf("canonical unit %q Cargo target must be %q:\n%s", unit.ID, targetName, manifest)
	}
}

func TestWorkspaceExplicitBinaryNames(t *testing.T) {
	for _, tc := range []struct{ name, id, target, nativePackage, nativeTarget string }{
		{"reserved package with safe target", "build", "safe_target", "build", "safe_target"},
		{"ordinary package with reserved target", "Worker", "build", "l_app", "l_app"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := compileWorkspaceNames(t, workspaceNameSource(tc.id))
			app := result.Compilation.Program.Apps[0]
			app.Services[0].Target = &sourceTarget{Kind: "bin", Name: tc.target, Root: "main", Primary: true}
			canonical, err := result.ApplicationUnits()
			if err != nil {
				t.Fatal(err)
			}
			assertWorkspaceNativeIdentities(t, result, WorkspaceUnit{BuildUnit: canonical[0]}, tc.id, tc.target)
			plan, err := result.WorkspacePlan()
			if err != nil {
				t.Fatal(err)
			}
			units, err := plan.Units(workspaceDependencyLocations(plan))
			if err != nil || len(units) != 1 || units[0].ID != tc.id {
				t.Fatalf("explicit-target unit lost canonical identity: %+v, %v", units, err)
			}
			assertWorkspaceNativeIdentities(t, result, units[0], tc.nativePackage, tc.nativeTarget)
			after, err := result.ApplicationUnits()
			if err != nil || !reflect.DeepEqual(canonical, after) {
				t.Fatalf("explicit-target projection changed canonical AML: %v", err)
			}
			if tc.id == tc.nativePackage && !bytes.Equal(units[0].BackendAML, canonical[0].BackendAML) {
				t.Fatal("safe effective binary acquired an unnecessary native alias")
			}
		})
	}
}

// This checks the public projection and canonical snapshots together: native
// names must never become the IDs accepted by orchestration or publication.
func checkWorkspaceNameProjection(t *testing.T, result *CompilationResult, want []WorkspacePackage, native []string) (*WorkspacePlan, []WorkspaceUnit) {
	t.Helper()
	canonical, err := result.ApplicationUnits()
	if err != nil {
		t.Fatal(err)
	}
	aml, backend := bytes.Clone(result.AML), bytes.Clone(result.BackendAML)
	programAML, programBackend := bytes.Clone(result.Compilation.Program.AMLBytes), bytes.Clone(result.Compilation.Program.BackendBytes)
	plan, err := result.WorkspacePlan()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(plan.Packages(), want) {
		t.Fatalf("canonical packages = %#v; want %#v", plan.Packages(), want)
	}
	locations := workspaceDependencyLocations(plan)
	units, err := plan.Units(locations)
	if err != nil || len(units) != len(want) {
		t.Fatalf("workspace units = %d, %v; want %d", len(units), err, len(want))
	}
	for i, unit := range units {
		if unit.ID != want[i].ID || unit.Library != want[i].Library {
			t.Errorf("unit %d changed canonical identity: %+v", i, unit.BuildUnit.ID)
		}
		assertWorkspaceNativeName(t, result, unit, native[i])
	}
	again, err := plan.Units(locations)
	if err != nil || !reflect.DeepEqual(units, again) {
		t.Fatalf("repeated workspace rendering changed: %v", err)
	}
	secondPlan, err := result.WorkspacePlan()
	if err != nil {
		t.Fatal(err)
	}
	secondUnits, err := secondPlan.Units(locations)
	if err != nil || !reflect.DeepEqual(units, secondUnits) {
		t.Fatalf("repeated workspace planning changed: %v", err)
	}
	units[0].BackendAML[0] = '!'
	units, err = plan.Units(locations)
	if err != nil || !reflect.DeepEqual(units, again) {
		t.Fatalf("workspace result retained caller-owned bytes: %v", err)
	}
	after, err := result.ApplicationUnits()
	if err != nil || !reflect.DeepEqual(canonical, after) || !bytes.Equal(result.AML, aml) || !bytes.Equal(result.BackendAML, backend) ||
		!bytes.Equal(result.Compilation.Program.AMLBytes, programAML) || !bytes.Equal(result.Compilation.Program.BackendBytes, programBackend) {
		t.Fatalf("workspace projection changed canonical graph or snapshots: %v", err)
	}
	companions := make([]AppArtifact, len(canonical))
	for i, unit := range canonical {
		companions[i] = AppArtifact{ID: unit.ID, RelativeExecutable: fmt.Sprintf("main.apps/app-%03d", i)}
	}
	if _, err := result.OrchestratorUnit("entry", companions); err != nil {
		t.Fatalf("canonical companion IDs no longer match orchestration: %v", err)
	}
	return plan, units
}

func TestWorkspaceReservedApplicationNames(t *testing.T) {
	var cases []struct{ name, native string }
	for _, name := range []string{"build", "deps", "examples", "incremental"} {
		for _, spelling := range []string{name, strings.ToUpper(name), strings.ToUpper(name[:1]) + name[1:]} {
			cases = append(cases, struct{ name, native string }{spelling, "l_app"})
		}
	}
	for _, name := range []string{"Worker", "build_app", "deps2"} {
		cases = append(cases, struct{ name, native string }{name, name})
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := compileWorkspaceNames(t, workspaceNameSource(tc.name))
			app := result.Compilation.Program.Apps[0]
			if app.ID != tc.name || app.buildName() != tc.name || app.Services[0].Target != nil {
				t.Fatal("test requires a canonical ordinary one-file App")
			}
			canonical, err := result.ApplicationUnits()
			if err != nil {
				t.Fatal(err)
			}
			// Direct backend callers retain their supplied package and target name.
			assertWorkspaceNativeName(t, result, WorkspaceUnit{BuildUnit: canonical[0]}, tc.name)
			_, units := checkWorkspaceNameProjection(t, result, []WorkspacePackage{{tc.name, false}}, []string{tc.native})
			if bytes.Contains(units[0].BackendAML, []byte("StructureVersion")) {
				t.Fatal("native naming promoted a flat App to structural topology")
			}
			if app.ID != tc.name || app.buildName() != tc.name || result.Compilation.Program.Books[0].Name != tc.name {
				t.Fatal("native naming replaced the canonical App identity")
			}
		})
	}
}

func TestWorkspaceReservedApplicationTopology(t *testing.T) {
	for _, tc := range []struct{ name, fixture, original string }{
		{"single Service", "test/classes/recursive.l", "Recursive"},
		{"multiple Services", "test/workspace/services.l", "First"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source, err := os.ReadFile(tc.fixture)
			if err != nil {
				t.Fatal(err)
			}
			// A new test-owned variant; the retained fixture remains unchanged.
			source = bytes.Replace(source, []byte(strconv.Quote(tc.original)), []byte(`"build"`), 1)
			result := compileWorkspaceNames(t, source)
			app := result.Compilation.Program.Apps[0]
			service := app.Services[0]
			if service.Target == nil || !service.Target.Primary || service.Target.Name != "build" {
				t.Fatal("test requires an explicit canonical primary bin")
			}
			target := *service.Target
			var want []WorkspacePackage
			var native []string
			if tc.name == "single Service" {
				want, native = []WorkspacePackage{{"build", false}}, []string{"l_app"}
			} else {
				want = []WorkspacePackage{{"build", false}, {"l_service_6275696c64_1", true}, {"l_service_6275696c64_2", true}, {"Second", false}, {"l_service_5365636f6e64_1", true}, {"l_service_5365636f6e64_2", true}}
				native = []string{"l_app", "l_service_6275696c64_1", "l_service_6275696c64_2", "Second", "l_service_5365636f6e64_1", "l_service_5365636f6e64_2"}
			}
			plan, units := checkWorkspaceNameProjection(t, result, want, native)
			if app.Services[0] != service || *service.Target != target {
				t.Fatal("projection changed the caller's Service or primary target")
			}
			if tc.name == "multiple Services" {
				rows := []string{`left = { package = "l_service_6275696c64_1", path = "../package-1" }`, `right = { package = "l_service_6275696c64_2", path = "../package-2" }`}
				if got := workspaceDependencyTestRows(t, units[0]); !reflect.DeepEqual(got, rows) {
					t.Fatalf("App alias changed Service identities/dependency rows: %#v", got)
				}
			}
			// A plan must not retain the target that its caller may subsequently edit.
			service.Target.Name = "caller_changed"
			after, err := plan.Units(workspaceDependencyLocations(plan))
			if err != nil || !reflect.DeepEqual(units, after) {
				t.Fatalf("workspace plan retained caller-owned target: %v", err)
			}
		})
	}
}

func TestWorkspaceApplicationNameAllocation(t *testing.T) {
	result := compileWorkspaceNames(t, workspaceNameSource("build", "deps", "Worker", "l_app", "L_APP_1"))
	// Exercise an existing explicit-target contract independently of package ID.
	worker := result.Compilation.Program.Apps[2]
	worker.Services[0].Target = &sourceTarget{Kind: "bin", Name: "L_APP_2", Root: "main", Primary: true}
	want := []WorkspacePackage{{"build", false}, {"deps", false}, {"Worker", false}, {"l_app", false}, {"L_APP_1", false}}
	native := []string{"l_app_3", "l_app_4", "Worker", "l_app", "L_APP_1"}
	// Worker intentionally has a distinct, safe target; check its package/target
	// separately rather than imposing package equality on supplied target names.
	canonical, err := result.ApplicationUnits()
	if err != nil {
		t.Fatal(err)
	}
	plan, err := result.WorkspacePlan()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(plan.Packages(), want) {
		t.Fatalf("canonical package IDs changed: %#v", plan.Packages())
	}
	units, err := plan.Units(workspaceDependencyLocations(plan))
	if err != nil {
		t.Fatal(err)
	}
	for i, unit := range units {
		if unit.ID != want[i].ID {
			t.Errorf("unit %d lost canonical ID %q", i, want[i].ID)
		}
		if i == 2 {
			if !bytes.Equal(unit.BackendAML, canonical[i].BackendAML) || !bytes.Contains(unit.BackendAML, []byte(`"Name":"L_APP_2"`)) {
				t.Fatal("safe explicit target or canonical package name changed")
			}
			continue
		}
		assertWorkspaceNativeName(t, result, unit, native[i])
	}
	second, err := result.WorkspacePlan()
	if err != nil {
		t.Fatal(err)
	}
	again, err := second.Units(workspaceDependencyLocations(second))
	if err != nil || !reflect.DeepEqual(units, again) {
		t.Fatalf("occupied-name allocation was not deterministic: %v", err)
	}
	after, err := result.ApplicationUnits()
	if err != nil || !reflect.DeepEqual(canonical, after) {
		t.Fatalf("occupied-name allocation changed canonical AML: %v", err)
	}
}

func TestWorkspaceCaseEquivalentBinaryNames(t *testing.T) {
	for _, tc := range []struct {
		name   string
		names  []string
		native []string
	}{
		{"two variants", []string{"Worker", "worker"}, []string{"Worker", "l_app"}},
		{"three variants", []string{"Worker", "worker", "WORKER"}, []string{"Worker", "l_app", "l_app_1"}},
		{"first source spelling wins", []string{"worker", "Worker"}, []string{"worker", "l_app"}},
		{"later occupied aliases", []string{"Worker", "worker", "l_app", "L_APP_1"}, []string{"Worker", "l_app_2", "l_app", "L_APP_1"}},
		{"reserved and duplicate names", []string{"build", "Worker", "worker", "BUILD", "l_app_1"}, []string{"l_app", "Worker", "l_app_2", "l_app_3", "l_app_1"}},
		{"hyphen and underscore stay distinct", []string{"worker-one", "worker_one"}, []string{"worker-one", "worker_one"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := compileWorkspaceNames(t, workspaceNameSource(tc.names...))
			want := make([]WorkspacePackage, len(tc.names))
			for i, name := range tc.names {
				want[i] = WorkspacePackage{ID: name}
			}
			_, units := checkWorkspaceNameProjection(t, result, want, tc.native)
			for _, unit := range units {
				if bytes.Contains(unit.BackendAML, []byte("StructureVersion")) {
					t.Fatal("binary disambiguation changed flat source topology")
				}
			}
		})
	}
}

func TestWorkspaceExplicitBinaryNameCollisions(t *testing.T) {
	for _, tc := range []struct {
		name           string
		packages       []string
		targets        []string
		nativePackages []string
		nativeTargets  []string
	}{
		{"identical and case-equivalent explicit targets", []string{"First", "Second", "Third"}, []string{"shared", "shared", "SHARED"}, []string{"First", "l_app", "l_app_1"}, []string{"shared", "l_app", "l_app_1"}},
		{"implicit then explicit", []string{"First", "Second"}, []string{"", "FIRST"}, []string{"First", "l_app"}, []string{"First", "l_app"}},
		{"later occupied target", []string{"First", "Second", "Third"}, []string{"shared", "SHARED", "L_APP"}, []string{"First", "l_app_1", "Third"}, []string{"shared", "l_app_1", "L_APP"}},
		{"distinct targets override case-equivalent packages", []string{"Worker", "worker"}, []string{"one", "two"}, []string{"Worker", "worker"}, []string{"one", "two"}},
		{"hyphen and underscore targets stay distinct", []string{"First", "Second"}, []string{"shared-one", "shared_one"}, []string{"First", "Second"}, []string{"shared-one", "shared_one"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := compileWorkspaceNames(t, workspaceNameSource(tc.packages...))
			for i, target := range tc.targets {
				if target != "" {
					result.Compilation.Program.Apps[i].Services[0].Target = &sourceTarget{Kind: "bin", Name: target, Root: "main", Primary: true}
				}
			}
			canonical, err := result.ApplicationUnits()
			if err != nil {
				t.Fatal(err)
			}
			beforeAML, beforeBackend := bytes.Clone(result.AML), bytes.Clone(result.BackendAML)
			plan, err := result.WorkspacePlan()
			if err != nil {
				t.Fatal(err)
			}
			units, err := plan.Units(workspaceDependencyLocations(plan))
			if err != nil || len(units) != len(tc.packages) {
				t.Fatalf("explicit-target workspace units: %d, %v", len(units), err)
			}
			for i, unit := range units {
				if unit.ID != tc.packages[i] || unit.Library {
					t.Fatalf("unit %d changed canonical identity: %+v", i, unit)
				}
				assertWorkspaceNativeIdentities(t, result, unit, tc.nativePackages[i], tc.nativeTargets[i])
			}
			second, err := result.WorkspacePlan()
			if err != nil {
				t.Fatal(err)
			}
			again, err := second.Units(workspaceDependencyLocations(second))
			if err != nil || !reflect.DeepEqual(units, again) {
				t.Fatalf("explicit-target allocation changed on repeated planning: %v", err)
			}
			after, err := result.ApplicationUnits()
			if err != nil || !reflect.DeepEqual(canonical, after) || !bytes.Equal(result.AML, beforeAML) || !bytes.Equal(result.BackendAML, beforeBackend) {
				t.Fatalf("explicit-target allocation changed canonical AML: %v", err)
			}
		})
	}
}

func TestWorkspaceCaseEquivalentAppsRetainServiceDependencies(t *testing.T) {
	source, err := os.ReadFile("test/workspace/services.l")
	if err != nil {
		t.Fatal(err)
	}
	// The source-owned fixture stays unchanged; only the second App's spelling
	// changes in this variant, including its independently constructed Services.
	source = bytes.Replace(source, []byte(`"Second"`), []byte(`"first"`), 1)
	result := compileWorkspaceNames(t, source)
	want := []WorkspacePackage{
		{"First", false}, {"l_service_4669727374_1", true}, {"l_service_4669727374_2", true},
		{"first", false}, {"l_service_6669727374_1", true}, {"l_service_6669727374_2", true},
	}
	native := []string{"First", "l_service_4669727374_1", "l_service_4669727374_2", "l_app", "l_service_6669727374_1", "l_service_6669727374_2"}
	app := result.Compilation.Program.Apps[1]
	service, target := app.Services[0], *app.Services[0].Target
	plan, units := checkWorkspaceNameProjection(t, result, want, native)
	for _, tc := range []struct {
		index int
		rows  []string
	}{
		{0, []string{`left = { package = "l_service_4669727374_1", path = "../package-1" }`, `right = { package = "l_service_4669727374_2", path = "../package-2" }`}},
		{3, []string{`left = { package = "l_service_6669727374_1", path = "../package-4" }`, `right = { package = "l_service_6669727374_2", path = "../package-5" }`}},
	} {
		if rows := workspaceDependencyTestRows(t, units[tc.index]); !reflect.DeepEqual(rows, tc.rows) {
			t.Fatalf("binary alias changed constructed Service dependency rows: %#v", rows)
		}
	}
	if app.Services[0] != service || *service.Target != target {
		t.Fatal("binary alias changed canonical Service or target identity")
	}
	service.Target.Name = "caller_changed"
	again, err := plan.Units(workspaceDependencyLocations(plan))
	if err != nil || !reflect.DeepEqual(units, again) {
		t.Fatalf("aliased workspace plan retained caller-owned target: %v", err)
	}
}
