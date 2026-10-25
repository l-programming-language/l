package Logiceme

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/l-programming-language/l/lerrors"
)

func workspaceRuntimeInput() []byte {
	return []byte(`<PROGRAM><STRUCTURE><BOOK "entry"><OUTLINE>
<JSON>{"AppMeta":{"AppGenDir":"/src"}}</JSON>
<JSON>{"Services":[{"Name":"main","Location":"/src/","GenericObjects":["main"]}]}</JSON>
<JSON>{"Objects":[{"Name":"main","Type":"root"}]}</JSON>
</OUTLINE><SETTINGS><JSON>{"Key":"TOKEN","Value":"workspace","Type":"String"}</JSON></SETTINGS><CONTENT><CHAPTER "main"><SENTENCE "main"><|W|><|S|>` + "`A`" + `</S|><|S|>` + "`FW{|Key: ~l_main~, Inputs: ~&[Prop{Key: runtime, Type: &runtime::Runtime}Prop]&~, Outputs: ~&[impl core::future::Future<Output = runtime::Result<()>> + '_]&~|}^W^BOD^E{|Statement: ~async { runtime.show(b\"source-owned\\xff\").await?; Ok(()) }~|}`" + `</S|></W|></SENTENCE></CHAPTER></CONTENT></BOOK></STRUCTURE></PROGRAM>`)
}

func workspaceFiles(t *testing.T, bundle Executable) map[string][]byte {
	t.Helper()
	if len(bundle.MapArrays) != 1 || len(bundle.ByteArrays) != 1 || len(bundle.MapArrays[0]) != len(bundle.ByteArrays[0]) {
		t.Fatalf("invalid single-package bundle: %#v", bundle.MapArrays)
	}
	files := map[string][]byte{}
	for i, name := range bundle.MapArrays[0] {
		if _, found := files[name]; found {
			t.Fatalf("duplicate artifact %q", name)
		}
		files[name] = bundle.ByteArrays[0][i]
	}
	return files
}

func TestWorkspaceRuntimeBridge(t *testing.T) {
	config := runtimeGenerationResources(t)
	service := NewLogicemeService(config)
	input := workspaceRuntimeInput()
	before := bytes.Clone(input)
	legacy, err := service.Evaluate(input, GenerationOptions{IncludeRuntime: true})
	if err != nil {
		t.Fatal(err)
	}
	legacyBefore := fmt.Sprintf("%#v", legacy)
	descriptor := WorkspaceRuntime{Package: "l_runtime_support_2", RelativeDirectory: "../support/runtime"}
	entry, support, err := service.EvaluateWorkspaceEntry(input, descriptor)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(input, before) || fmt.Sprintf("%#v", legacy) != legacyBefore {
		t.Fatal("workspace bridge mutated input or previous direct bundle")
	}
	entryFiles, supportFiles := workspaceFiles(t, entry), workspaceFiles(t, support)
	if !reflect.DeepEqual(entry.MapArrays[0], []string{"/src/main.rs", "/Cargo.toml", "/src/config.rs", "/start.sh", "/stop.sh"}) {
		t.Fatalf("entry has inline runtime artifacts: %v", entry.MapArrays)
	}
	main := entryFiles["/src/main.rs"]
	if bytes.Count(main, []byte("use ::l_runtime_support_2::runtime;")) != 1 || bytes.Contains(main, []byte("mod runtime;")) || bytes.Contains(main, []byte("mod process;")) {
		t.Fatalf("entry did not use allocated runtime library:\n%s", main)
	}
	if !bytes.Contains(main, []byte(`runtime.show(b"source-owned\xff").await?; Ok(())`)) || bytes.Count(main, []byte("\nfn l_main(")) != 1 {
		t.Fatal("bridge changed or duplicated source-owned body")
	}
	wrapper, err := fs.ReadFile(config, "builtins/templates/runtime/main.lu")
	if err != nil {
		t.Fatal(err)
	}
	boundary := bytes.Index(wrapper, []byte("// The source assembler supplies a same-module generated item with this contract:"))
	if boundary < 0 || bytes.Count(main, wrapper[boundary:]) != 1 {
		t.Fatal("runtime entry suffix changed")
	}
	if !bytes.Contains(entryFiles["/Cargo.toml"], []byte(`l_runtime_support_2 = { path = "../support/runtime" }`)) {
		t.Fatal("allocated runtime dependency missing")
	}
	legacyFiles := workspaceFiles(t, legacy)
	for _, name := range []string{"/src/config.rs", "/start.sh", "/stop.sh"} {
		if !bytes.Equal(entryFiles[name], legacyFiles[name]) {
			t.Fatalf("workspace changed ordinary entry support %s", name)
		}
	}
	const libraryRoot = "#![forbid(unsafe_code)]\n\npub mod runtime;\n#[cfg(any(target_os = \"linux\", target_os = \"macos\"))]\n#[path = \"process_unix.rs\"]\nmod process;\n#[cfg(not(any(target_os = \"linux\", target_os = \"macos\")))]\n#[path = \"process_other.rs\"]\nmod process;\n"
	if string(supportFiles["/src/lib.rs"]) != libraryRoot {
		t.Fatalf("library root = %q", supportFiles["/src/lib.rs"])
	}
	if !reflect.DeepEqual(support.MapArrays[0], []string{"/src/lib.rs", "/src/runtime.rs", "/src/process_unix.rs", "/src/process_other.rs", "/Cargo.toml"}) {
		t.Fatalf("runtime library inventory: %v", support.MapArrays)
	}
	for _, name := range []string{"runtime", "process_unix", "process_other"} {
		canonical, err := fs.ReadFile(config, "builtins/templates/runtime/"+name+".lu")
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(canonical, supportFiles["/src/"+name+".rs"]) {
			t.Fatalf("changed canonical %s", name)
		}
	}
	cargo := string(supportFiles["/Cargo.toml"])
	for _, required := range []string{`name = "l_runtime_support_2"`, "autobins = false", "[lib]\nname = \"l_runtime_support_2\"\npath = \"src/lib.rs\"", `tokio = { version = "=1.53.1"`, `nix = { version = "=0.31.3"`} {
		if !strings.Contains(cargo, required) {
			t.Fatalf("runtime library lacks %q:\n%s", required, cargo)
		}
	}
	if strings.Contains(cargo, "[[bin]]") {
		t.Fatal("runtime support acquired binary target")
	}
	entryBefore, supportBefore := fmt.Sprintf("%#v", entry), fmt.Sprintf("%#v", support)
	again, err := service.Evaluate(input, GenerationOptions{IncludeRuntime: true})
	if err != nil || !reflect.DeepEqual(again, legacy) {
		t.Fatalf("workspace mode leaked into direct API: %v", err)
	}
	entryAgain, supportAgain, err := service.EvaluateWorkspaceEntry(input, descriptor)
	if err != nil || !reflect.DeepEqual(entryAgain, entry) || !reflect.DeepEqual(supportAgain, support) || fmt.Sprintf("%#v", entry) != entryBefore || fmt.Sprintf("%#v", support) != supportBefore {
		t.Fatalf("workspace repeat/isolation: %v", err)
	}
}

func TestWorkspaceRuntimeValidation(t *testing.T) {
	for _, test := range []struct {
		name       string
		descriptor WorkspaceRuntime
	}{
		{"missing package", WorkspaceRuntime{RelativeDirectory: "../support/runtime"}},
		{"keyword package", WorkspaceRuntime{Package: "self", RelativeDirectory: "../support/runtime"}},
		{"invalid package", WorkspaceRuntime{Package: "native-runtime", RelativeDirectory: "../support/runtime"}},
		{"missing path", WorkspaceRuntime{Package: "l_runtime_support"}},
		{"absolute path", WorkspaceRuntime{Package: "l_runtime_support", RelativeDirectory: "/support/runtime"}},
		{"noncanonical path", WorkspaceRuntime{Package: "l_runtime_support", RelativeDirectory: "../support/../runtime"}},
		{"unsafe path", WorkspaceRuntime{Package: "l_runtime_support", RelativeDirectory: "../support/\nruntime"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			input := workspaceRuntimeInput()
			before := bytes.Clone(input)
			entry, support, err := NewLogicemeService(runtimeGenerationResources(t)).EvaluateWorkspaceEntry(input, test.descriptor)
			if !errors.Is(err, lerrors.InvalidInput) || !reflect.DeepEqual(entry, Executable{}) || !reflect.DeepEqual(support, Executable{}) || !bytes.Equal(input, before) {
				t.Fatalf("invalid bridge leaked output/input: %v", err)
			}
		})
	}
}

func TestWorkspaceRuntimeEntryShapes(t *testing.T) {
	base := string(workspaceRuntimeInput())
	book := base[strings.Index(base, "<BOOK ") : strings.Index(base, "</BOOK>")+len("</BOOK>")]
	wordStart, wordEnd := strings.Index(base, "<|W|>"), strings.Index(base, "</W|>")+len("</W|>")
	literal := base[:wordStart] + "<|W|><|S|>`EA{|Contents: ~fn l_main() {}~|}`</S|></W|>" + base[wordEnd:]
	rows := marshalTestJSON(t, []string{`l_runtime_support = { path = "../other" }`})
	encoded := marshalTestJSON(t, string(rows))
	conflict := strings.Replace(base, `"AppGenDir":"/src"`, `"AppGenDir":"/src","Deps":`+string(encoded), 1)
	normalizedConflict := strings.Replace(conflict, "l_runtime_support =", "l-runtime-support =", 1)
	for _, test := range []struct{ name, input, cause string }{
		{"multiple books", strings.Replace(base, "</STRUCTURE>", book+"</STRUCTURE>", 1), "exactly one application book"},
		{"opaque entry", literal, "composed primary Object"},
		{"dependency collision", conflict, "collides with supplied entry dependency"},
		{"normalized dependency collision", normalizedConflict, "collides with supplied entry dependency"},
		{"package collision", strings.Replace(base, `<BOOK "entry">`, `<BOOK "l_runtime_support">`, 1), "runtime package collides with entry package"},
		{"no primary", strings.Replace(base, `"AppGenDir":"/src"`, `"AppGenDir":"/another"`, 1), "no primary entry Object"},
	} {
		t.Run(test.name, func(t *testing.T) {
			input := []byte(test.input)
			before := bytes.Clone(input)
			entry, support, err := NewLogicemeService(runtimeGenerationResources(t)).EvaluateWorkspaceEntry(input, WorkspaceRuntime{Package: "l_runtime_support", RelativeDirectory: "../support/runtime"})
			if !errors.Is(err, lerrors.InvalidInput) || !strings.Contains(err.Error(), test.cause) {
				t.Fatalf("entry rejection = %v, want %s", err, test.cause)
			}
			if !reflect.DeepEqual(entry, Executable{}) || !reflect.DeepEqual(support, Executable{}) || !bytes.Equal(input, before) {
				t.Fatal("invalid entry leaked output or mutated input")
			}
		})
	}
}

func TestWorkspaceRuntimeResourceOwnership(t *testing.T) {
	config := runtimeGenerationResources(t)
	provider := config
	files := fstest.MapFS{}
	for _, root := range []string{"builtins/segments", "builtins/templates"} {
		if err := fs.WalkDir(provider, root, func(name string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				return nil
			}
			data, err := fs.ReadFile(provider, name)
			if err != nil {
				return err
			}
			files[name] = &fstest.MapFile{Data: bytes.Clone(data), Mode: 0600}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	const asset = "builtins/templates/runtime/runtime.lu"
	selected := append([]byte("// explicit selected provider\n"), files[asset].Data...)
	files[asset] = &fstest.MapFile{Data: bytes.Clone(selected), Mode: 0600}
	config = files
	service := NewLogicemeService(config)
	input := workspaceRuntimeInput()
	_, support, err := service.EvaluateWorkspaceEntry(input, WorkspaceRuntime{Package: "l_runtime_support", RelativeDirectory: "../support/runtime"})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(workspaceFiles(t, support)["/src/runtime.rs"], selected) {
		t.Fatal("workspace runtime ignored selected ResourceFS")
	}
	before := fmt.Sprintf("%#v", support)
	ordinary, err := service.Evaluate(input, GenerationOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if _, present := workspaceFiles(t, ordinary)["/src/runtime.rs"]; present {
		t.Fatal("workspace mode leaked into ordinary evaluation")
	}
	_, repeated, err := service.EvaluateWorkspaceEntry(input, WorkspaceRuntime{Package: "l_runtime_support", RelativeDirectory: "../support/runtime"})
	if err != nil || !reflect.DeepEqual(support, repeated) || fmt.Sprintf("%#v", support) != before {
		t.Fatalf("selected resources/repeated bridge changed prior output: %v", err)
	}
}
