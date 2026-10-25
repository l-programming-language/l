package Lyger

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func refactorGenerationCompilation(t *testing.T) *Compilation {
	t.Helper()
	source := memberResolutionFixture(t)
	source.main += "\n\nCreate \"SecondWorker\" as a New Basic App.\nThe App Prints x.\n"
	return executableTestResult(t, source).Compilation
}

func TestRefactorGenerationIdentityPolicies(t *testing.T) {
	program, apps := renderTestGraph()
	apps[0].ID = ""
	program.Apps = apps
	observable, backend, err := renderProgram(program, apps)
	if err != nil || !bytes.Contains(observable, []byte(`<BOOK "Atlas">`)) || len(backend) == 0 {
		t.Fatalf("renderer metadata identity fallback changed: %v", err)
	}
	result := &CompilationResult{Compilation: &Compilation{Program: program}}
	if units, err := result.ApplicationUnits(); err == nil || err.Error() != "build graph has a nil, unnamed, or duplicate application" || units != nil {
		t.Fatalf("build accepted noncanonical metadata identity: units=%v, error=%v", units, err)
	}
}

func TestRefactorGenerationValidationOrder(t *testing.T) {
	program, apps := renderTestGraph()
	other, moreApps := renderTestGraph()
	moreApps[0].ID, moreApps[0].Meta["AppName"] = "Beacon", "Beacon"
	other.Books[0].Name = "Missing"
	program.Books = append(program.Books, other.Books[0])
	program.Apps = append(apps, moreApps...)
	program.Books[0].Chapters[0].sentences[0].Words[0].Syllables[0] = nil
	observable, backend, err := renderProgram(program, program.Apps)
	want := `AML book "Atlas": sentence "endpoint" contains an empty or nil syllable`
	if err == nil || err.Error() != want || observable != nil || backend != nil {
		t.Fatalf("renderer must encounter earlier content failure first: %v", err)
	}
	result := &CompilationResult{Compilation: &Compilation{Program: program}}
	if units, err := result.ApplicationUnits(); err == nil || err.Error() != "application book has no unique canonical application" || units != nil {
		t.Fatalf("build must validate all identities before content: units=%v, error=%v", units, err)
	}
}

type refactorGenerationState struct {
	program                  *Program
	apps                     []*App
	appBytes, bodyBytes      []byte
	asm, programAML, backend []byte
}

func refactorGenerationJSON(t *testing.T, value interface{}) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func refactorGenerationBodies(compilation *Compilation) [][]DefinitionInstance {
	var bodies [][]DefinitionInstance
	for _, construction := range compilation.Semantics.Constructions {
		bodies = append(bodies, construction.Body)
	}
	return bodies
}

func captureRefactorGenerationState(t *testing.T, compilation *Compilation) refactorGenerationState {
	t.Helper()
	state := refactorGenerationState{
		program: compilation.Program, apps: append([]*App(nil), compilation.Semantics.Apps...),
		appBytes:  refactorGenerationJSON(t, compilation.Semantics.Apps),
		bodyBytes: refactorGenerationJSON(t, refactorGenerationBodies(compilation)),
		asm:       bytes.Clone(compilation.ASMBytes),
	}
	if compilation.Program != nil {
		state.programAML = bytes.Clone(compilation.Program.AMLBytes)
		state.backend = bytes.Clone(compilation.Program.BackendBytes)
	}
	return state
}

func (state refactorGenerationState) requireUnchanged(t *testing.T, compilation *Compilation) {
	t.Helper()
	if compilation.Program != state.program || !bytes.Equal(compilation.ASMBytes, state.asm) {
		t.Error("failed lowering replaced the published Program or ASM bytes")
	}
	if !bytes.Equal(refactorGenerationJSON(t, compilation.Semantics.Apps), state.appBytes) {
		t.Error("failed lowering mutated application metadata, services, or retained fields")
	}
	if !bytes.Equal(refactorGenerationJSON(t, refactorGenerationBodies(compilation)), state.bodyBytes) {
		t.Error("failed lowering published construction bodies")
	}
	for index, app := range state.apps {
		if compilation.Semantics.Apps[index] != app || compilation.Memory[app.ID].Value.Applicative.App != app {
			t.Error("failed lowering changed canonical application identity")
		}
	}
	if state.program != nil && (!bytes.Equal(state.program.AMLBytes, state.programAML) || !bytes.Equal(state.program.BackendBytes, state.backend)) {
		t.Error("failed lowering mutated a previously completed Program's output")
	}
}

func TestRefactorGenerationAtomicFailure(t *testing.T) {
	for _, tc := range []struct {
		name, want     string
		fresh          bool
		invalidService bool
		renderFailure  bool
	}{
		{name: "later function preserves existing result", want: "function main has no definition"},
		{name: "render failure publishes nothing", want: "invalid AML element name", fresh: true, renderFailure: true},
		{name: "render failure preserves existing result", want: "invalid AML element name", renderFailure: true},
		{name: "later service preserves existing result", want: "application child must be a named service", invalidService: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			compilation := refactorGenerationCompilation(t)
			for _, construction := range compilation.Semantics.Constructions {
				app := construction.App
				app.Meta = map[string]string{"retained": app.ID}
				app.Settings = map[string]string{"retained": app.ID}
				app.Status, app.Output, app.Error = "retained status", "retained output\n", "retained error"
				app.AML, app.Params = map[string]string{"retained": "aml"}, map[string]string{"retained": "parameter"}
				app.Compilation = []byte("retained compilation")
				construction.Body = []DefinitionInstance{{Symbol: "retained", Expression: "A"}}
				if tc.fresh {
					app.Services = nil
				}
			}
			second := compilation.Semantics.Constructions[1]
			if tc.fresh {
				compilation.Program, compilation.ASMBytes = nil, nil
			}
			if tc.renderFailure {
				second.Root.Children[0].Name = "invalid\"service"
			} else if tc.invalidService {
				second.Root.Children[0].Name = ""
			} else {
				second.Root.Children[0].Children[0].Children[0].Definition = nil
			}
			before := captureRefactorGenerationState(t, compilation)
			err := compilation.lowerProgram()
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("lowering error = %v, want %q", err, tc.want)
			}
			before.requireUnchanged(t, compilation)
		})
	}
}

func TestRefactorGenerationReentryReplacesOwnedServices(t *testing.T) {
	compilation := refactorGenerationCompilation(t)
	canonical := append([]*App(nil), compilation.Semantics.Apps...)
	for _, app := range canonical {
		app.Status, app.Output, app.Error = "retained status", "retained output\n", "retained error"
		app.AML, app.Params = map[string]string{"retained": "aml"}, map[string]string{"retained": "parameter"}
		app.Compilation = []byte("retained compilation")
		app.Objects = []*object{{ID: "retainedObject", Meta: map[string]string{"Type": "root"}}}
	}
	for run := 0; run < 2; run++ {
		priorBodies := refactorGenerationJSON(t, refactorGenerationBodies(compilation))
		if err := compilation.lowerProgram(); err != nil {
			t.Fatalf("lowering pass %d: %v", run+1, err)
		}
		if !bytes.Equal(priorBodies, refactorGenerationJSON(t, refactorGenerationBodies(compilation))) {
			t.Fatal("repeated lowering changed source-derived application bodies")
		}
		for index, app := range canonical {
			if compilation.Program.Apps[index] != app || compilation.Semantics.Apps[index] != app || compilation.Memory[app.ID].Value.Applicative.App != app {
				t.Fatal("successful generation published temporary App identities")
			}
			if len(app.Services) != 1 || len(app.Services[0].Objects) != 1 || len(app.Services[0].Objects[0].Attributes) != 1 {
				t.Fatal("repeated generation duplicated services, objects, or attributes")
			}
			if app.Status != "retained status" || app.Output != "retained output\n" || app.Error != "retained error" || app.AML["retained"] != "aml" || app.Params["retained"] != "parameter" || string(app.Compilation) != "retained compilation" || len(app.Objects) != 1 || app.Objects[0].ID != "retainedObject" {
				t.Fatal("generation overwrote fields outside its ownership")
			}
		}
		if !bytes.Contains(compilation.ASMBytes, []byte(`"Name":"retainedObject"`)) {
			t.Fatal("temporary generation omitted retained application objects")
		}
		if run == 0 {
			before := bytes.Clone(compilation.ASMBytes)
			backend := bytes.Clone(compilation.Program.BackendBytes)
			defer func() {
				if !bytes.Equal(compilation.ASMBytes, before) {
					t.Error("repeated successful lowering changed observable AML")
				}
				if !bytes.Equal(compilation.Program.BackendBytes, backend) {
					t.Error("repeated successful lowering changed backend AML")
				}
			}()
		}
	}
}

func TestRefactorGenerationCurrentDefaultsAndCanonicalIdentity(t *testing.T) {
	compilation := refactorGenerationCompilation(t)
	if len(compilation.Semantics.Apps) != 2 || len(compilation.Semantics.Constructions) != 2 || len(compilation.Program.Apps) != 2 {
		t.Fatal("source did not retain both applications")
	}
	for index, app := range compilation.Semantics.Apps {
		if compilation.Program.Apps[index] != app || compilation.Semantics.Constructions[index].App != app || compilation.Memory[app.ID].Value.Applicative.App != app {
			t.Fatalf("application %s lost its canonical pointer", app.ID)
		}
		wantMeta := map[string]string{"AppName": app.ID, "AppLang": "rust", "AppVersion": "0.0.0", "AppGenDir": "/src"}
		if !reflect.DeepEqual(app.Meta, wantMeta) || !reflect.DeepEqual(app.Settings, map[string]string{"Host": "http://127.0.0.1"}) {
			t.Fatalf("application defaults = %#v / %#v", app.Meta, app.Settings)
		}
		if len(app.Services) != 1 || len(app.Services[0].Objects) != 1 {
			t.Fatalf("application %s has unexpected generated component counts", app.ID)
		}
		service := app.Services[0]
		if !reflect.DeepEqual(service.Meta, map[string]string{"Type": "root", "Package": service.ID, "Location": "/src/", "Parent": "None"}) || !reflect.DeepEqual(service.Objects[0].Meta, map[string]string{"Type": "root"}) {
			t.Fatalf("root component defaults = %#v / %#v", service.Meta, service.Objects[0].Meta)
		}
	}
	result := &CompilationResult{Compilation: compilation}
	unit, err := result.OrchestratorUnit("controller", executableTestCompanions(result))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(unit.BackendAML, []byte(`<JSON>{"AppMeta":{"AppName":"controller","AppLang":"rust","AppVersion":"0.0.0","AppGenDir":"/src"}}</JSON>`)) || !bytes.Contains(unit.BackendAML, []byte("<SETTINGS>\n</SETTINGS>")) {
		t.Fatal("synthetic orchestrator defaults or empty settings changed")
	}
	first, second := compilation.Semantics.Apps[0], compilation.Semantics.Apps[1]
	first.Meta["AppVersion"] = "changed"
	first.Settings["Host"] = "changed"
	first.Services[0].Meta["Parent"] = "changed"
	first.Services[0].Objects[0].Meta["Type"] = "changed"
	if second.Meta["AppVersion"] != "0.0.0" || second.Settings["Host"] != "http://127.0.0.1" || second.Services[0].Meta["Parent"] != "None" || second.Services[0].Objects[0].Meta["Type"] != "root" {
		t.Fatal("independent applications share generation metadata maps")
	}
}
