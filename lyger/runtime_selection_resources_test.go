package Lyger

import (
	"bytes"
	"testing"
)

func TestRuntimeSelectionRequiresImmutableAuditedResources(t *testing.T) {
	const source = "Show `selected`.\n"
	embedded, err := NewLygerService(Options{}).Compile([]byte(source), "selected.l")
	if err != nil {
		t.Fatal(err)
	}
	first, err := embedded.RuntimeLibrary("support", []string{"runtime::runtime_show as _"})
	if err != nil || len(first) == 0 {
		t.Fatalf("embedded selection: %v", err)
	}
	first[0] = '!'
	again, err := embedded.RuntimeLibrary("support", []string{"runtime::runtime_show as _"})
	if err != nil || len(again) == 0 || again[0] == '!' {
		t.Fatalf("returned AML aliases a later selection: %v", err)
	}
	if selected, err := embedded.RuntimeLibrary("support", []string{"runtime::unknown"}); err != nil || selected != nil {
		t.Fatalf("unknown native dependency must retain full support: %v", err)
	}
	custom, err := NewLygerService(Options{Resources: dictionaryIndexFilesystem(t)}).Compile([]byte(source), "selected.l")
	if err != nil {
		t.Fatal(err)
	}
	if selected, err := custom.RuntimeLibrary("support", nil); err != nil || selected != nil {
		t.Fatalf("even byte-identical mutable resources must retain full support: %v", err)
	}
}

func TestRuntimeSelectionIncludesPrivateHostRoots(t *testing.T) {
	files := dictionaryIndexFilesystem(t)
	const path = "builtins/runtime/eng/dictionary.l"
	files[path].Data = append(files[path].Data, []byte("\n    Create Runtime's \"Hidden\" as a Generic `runtime_host_only` Function with No Inputs and Outputs of `u8`.\nReturn a New Runtime's Byte of 1.\n\nPublicly use Runtime's Hidden as Runtime's Host's \"Hidden\".\n")...)
	result, err := NewLygerService(Options{Resources: files}).Compile([]byte("Use Runtime from `runtime`.\n"), "host.l")
	if err != nil {
		t.Fatal(err)
	}
	aml, selection, err := result.Compilation.ComposeRuntimeLibrary()
	if err != nil {
		t.Fatal(err)
	}
	for _, span := range selection[""] {
		if bytes.Contains(aml[span[0]:span[1]], []byte("runtime_host_only")) {
			return
		}
	}
	t.Fatal("private host root was omitted from the required runtime composition")
}
