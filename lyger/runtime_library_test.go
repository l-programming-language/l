package Lyger

import (
	"bytes"
	"testing"

	Logiceme "github.com/l-programming-language/l/logiceme"
)

func TestRuntimeLibrarySelectsDeclaredDependencies(t *testing.T) {
	result, err := NewLygerService(Options{}).Compile([]byte("Show `selected`.\n"), "selected.l")
	if err != nil {
		t.Fatal(err)
	}
	unit, err := result.OrchestratorUnit("entry", nil)
	if err != nil {
		t.Fatal(err)
	}
	backend := Logiceme.NewLogicemeService(nil)
	imports, err := backend.ReadImports(unit.BackendAML)
	if err != nil {
		t.Fatal(err)
	}
	library, err := result.RuntimeLibrary("support", imports)
	if err != nil || library == nil {
		t.Fatalf("canonical runtime selection: %v", err)
	}
	_, support, err := backend.EvaluateWorkspaceEntry(unit.BackendAML, Logiceme.WorkspaceRuntime{Package: "support", RelativeDirectory: "../support"}, library)
	if err != nil {
		t.Fatal(err)
	}
	for i, paths := range support.MapArrays {
		for j, path := range paths {
			if path != "/src/runtime.rs" {
				continue
			}
			code := support.ByteArrays[i][j]
			for _, required := range []string{"fn runtime_new(", "fn runtime_run(", "fn runtime_show(", "fn error_copy("} {
				if !bytes.Contains(code, []byte(required)) {
					t.Fatalf("missing selected or host dependency %s", required)
				}
			}
			if bytes.Contains(code, []byte("fn app_build(")) {
				t.Fatal("unused App build root retained")
			}
		}
	}
	opaque, err := NewLygerService(Options{}).Compile([]byte("Define \"Opaque\" as `<DEF>E{|Statement: ~runtime::app_build~|}<DEF>`.\nShow `opaque`.\n"), "opaque.l")
	if err != nil {
		t.Fatal(err)
	}
	if selected, err := opaque.RuntimeLibrary("support", imports); err != nil || selected != nil {
		t.Fatalf("opaque authored native leaf must retain full support: %v", err)
	}
	for _, name := range []string{"stdlib/eng/storage/dictionary.l", "builtins/operations/rust/process-run.def", "builtins/templates/runtime/host.lu"} {
		files := dictionaryIndexFilesystem(t)
		files[name].Data = append(bytes.Clone(files[name].Data), '\n')
		changed, err := NewLygerService(Options{Resources: files}).Compile([]byte("Show `changed`.\n"), "changed.l")
		if err != nil {
			t.Fatal(err)
		}
		if selected, err := changed.RuntimeLibrary("support", imports); err != nil || selected != nil {
			t.Fatalf("modified %s must retain full support: %v", name, err)
		}
	}
}
