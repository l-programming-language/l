package Logiceme

import (
	"bytes"
	"testing"
)

func TestWorkspaceUsesExplicitRuntimeLibrary(t *testing.T) {
	files := compositionResources(t, "7")
	library := bytes.Replace(files["runtime/generated/support.aml"].Data, []byte(`<BOOK "runtime">`), []byte(`<BOOK "chosen_runtime">`), 1)
	library = bytes.Replace(library, []byte(`"Name":"runtime"`), []byte(`"Name":"chosen_runtime"`), 1)
	delete(files, "runtime/generated/support.aml")
	delete(files, "runtime/generated/support.json")
	_, support, err := NewLogicemeService(files).EvaluateWorkspaceEntry(workspaceRuntimeInput(), WorkspaceRuntime{Package: "chosen_runtime", RelativeDirectory: "../runtime"}, library)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(workspaceFiles(t, support)["/src/runtime.rs"], []byte("selected cancellation: 7")) {
		t.Fatal("workspace did not consume its explicitly supplied library composition")
	}
	if _, _, err := NewLogicemeService(files).EvaluateWorkspaceEntry(workspaceRuntimeInput(), WorkspaceRuntime{Package: "chosen_runtime", RelativeDirectory: "../runtime"}); err == nil {
		t.Fatal("entry without a supplied library bypassed canonical runtime loading")
	}
}
