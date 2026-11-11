//go:build darwin || linux

package main

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	distribution "github.com/l-programming-language/l"
	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	"github.com/l-programming-language/l/lerrors"
	Logiceme "github.com/l-programming-language/l/logiceme"
	Lyger "github.com/l-programming-language/l/lyger"
)

func TestCoreHelloWorld(t *testing.T) {
	for _, test := range []struct{ name, source, output string }{
		{"literal", "Create \"Hello\" as a New Basic App.\nThe App Prints `Hello World!`.\n", "Hello World!\n"},
		{"binding", "\"message\" is `Rain, \"quoted\" snow 雪!`.\nCreate \"Weather\" as a New Basic App.\nThe App Prints message.\n", "Rain, \"quoted\" snow 雪!\n"},
		{"ordered", "Create \"Lines\" as a New Basic App.\nThe App Prints `first` and then Prints `second`.\n", "first\nsecond\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			directory := t.TempDir()
			path := filepath.Join(directory, "main.l")
			if err := os.WriteFile(path, []byte(test.source), 0600); err != nil {
				t.Fatal(err)
			}
			compiler := newCompiler(&cliData{jobType: "build", mainPath: path})
			compiler.initialize()
			generated := compiler.process()
			if compiler.err != nil || generated == nil {
				t.Fatalf("generate standalone App: %v", compiler.err)
			}
			assertCoreBundle(t, generated)
			nativetest.RecordBundle(t, "core-hello-world", generated.Sources)
			var stdout, stderr bytes.Buffer
			if status := run(context.Background(), []string{"build", "--output-dir", directory, path}, &stdout, &stderr); status != 0 {
				t.Fatalf("CLI build status %d: stdout=%q stderr=%q", status, stdout.String(), stderr.String())
			}
			executable := filepath.Join(directory, "main")
			if stdout.String() != executable+"\n" || stderr.Len() != 0 {
				t.Fatalf("CLI output: stdout=%q stderr=%q", stdout.String(), stderr.String())
			}
			nativetest.AssertNativeArtifact(t, executable)
			var manifest generationManifest
			generation := nativetest.ReadPublishedGeneration(t, executable, &manifest)
			if len(manifest.Apps) != 1 {
				t.Fatalf("published App inventory: %+v", manifest.Apps)
			}
			app := manifest.Apps[0]
			if app.SHA256 != nativetest.NativeArtifactHash(t, executable) || app.SHA256 != nativetest.NativeArtifactHash(t, filepath.Join(filepath.Dir(generation), app.Executable)) {
				t.Fatal("published entry is not the compiled App artifact")
			}
			actual, diagnostics, err := nativetest.RuntimeExecution.Run(t, executable, directory, nil)
			if err != nil || string(actual) != test.output || len(diagnostics) != 0 {
				t.Fatalf("App execution: %v stdout=%q stderr=%q; want %q", err, actual, diagnostics, test.output)
			}
		})
	}
}

func assertCoreBundle(t *testing.T, generated *generatedProgram) {
	t.Helper()
	if len(generated.Sources.ByteArrays) != 1 || len(generated.Sources.MapArrays) != 1 {
		t.Fatalf("expected one real App bundle: %v", generated.Sources.MapArrays)
	}
	for index, name := range generated.Sources.MapArrays[0] {
		data := generated.Sources.ByteArrays[0][index]
		for _, retired := range []string{"runtime::", "tokio", "support/runtime", "mod runtime", "Storage"} {
			if strings.Contains(name, retired) || bytes.Contains(data, []byte(retired)) {
				t.Errorf("%s contains retired support %q", name, retired)
			}
		}
	}
	for _, origin := range generated.Compilation.Compilation.SourceOrigins {
		if strings.HasPrefix(origin.Name, "stdlib/") || strings.Contains(origin.Name, "/storage/") || strings.HasSuffix(origin.Name, "/prelude.l") {
			t.Errorf("standalone source loaded retired resource %s", origin.Name)
		}
	}
}

func TestCoreHelloWorldRejectsRuntimeEffects(t *testing.T) {
	for _, action := range []string{
		"Build Hello.", "Run Hello.", "Stop Hello.", "Await Hello.", "Execute Hello.",
		"Show Hello's Output.", "Show `outside`.", "Exit.",
		"Create \"Second\" as a New Basic App.\nThe App Prints `second`.",
	} {
		t.Run(action, func(t *testing.T) {
			compiler := coreSourceCompiler(distribution.Files(), "Create \"Hello\" as a New Basic App.\nThe App Prints `inside`.\n"+action+"\n")
			if generated := compiler.process(); generated != nil || compiler.err == nil {
				t.Fatalf("runtime effect was silently discarded: %s", action)
			}
			if !errors.Is(compiler.err, lerrors.Unsupported) {
				t.Fatalf("expected explicit unsupported diagnostic, got %v", compiler.err)
			}
		})
	}
	compiler := coreSourceCompiler(distribution.Files(), "\"message\" is `no application`.\n")
	if compiler.process() != nil || !errors.Is(compiler.err, lerrors.Unsupported) {
		t.Fatalf("source without an App: %v", compiler.err)
	}
}

type coreResourceOverride struct {
	fs.FS
	core []byte
}

func (resources coreResourceOverride) Open(name string) (fs.File, error) {
	if name == "builtins/eng/core.l" {
		if resources.core == nil {
			return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
		}
		return fstest.MapFS{name: {Data: resources.core}}.Open(name)
	}
	return resources.FS.Open(name)
}

func coreSourceCompiler(resources fs.FS, source string) *compiler {
	options := Lyger.Options{Resources: resources}
	return &compiler{
		cli: &cliData{jobType: "build", mainPath: "standalone.l"}, options: options,
		lyger: Lyger.NewLygerService(options), log: Logiceme.NewLogicemeService(resources), mainBytes: []byte(source),
	}
}

func TestCoreHelloWorldUsesSelectedPrint(t *testing.T) {
	core, err := fs.ReadFile(distribution.Files(), "builtins/eng/core.l")
	if err != nil {
		t.Fatal(err)
	}
	mutated := bytes.Replace(core, []byte("println!"), []byte("eprintln!"), 1)
	if bytes.Equal(core, mutated) {
		t.Fatal("core Print has no selected native println call")
	}
	compiler := coreSourceCompiler(coreResourceOverride{FS: distribution.Files(), core: mutated}, "Create \"Hello\" as a New Basic App.\nThe App Prints `selected`.\n")
	generated := compiler.process()
	if compiler.err != nil || generated == nil {
		t.Fatalf("selected core compilation: %v", compiler.err)
	}
	if !bytes.Contains(generated.Sources.ByteArrays[0][0], []byte("eprintln!")) {
		t.Fatal("Print did not instantiate the selected core definition")
	}
	compiler = coreSourceCompiler(coreResourceOverride{FS: distribution.Files()}, "Create \"Hello\" as a New Basic App.\nThe App Prints `selected`.\n")
	if compiler.process() != nil || compiler.err == nil {
		t.Fatal("missing core definitions unexpectedly compiled")
	}
	if !strings.Contains(compiler.err.Error(), "builtins/eng/core.l") {
		t.Fatalf("missing selected core lost its resource diagnostic: %v", compiler.err)
	}
	compiler = coreSourceCompiler(distribution.Files(), "Create \"Hello\" as a New Basic App.\nThe App Logs `private`.\n")
	if compiler.process() != nil || compiler.err == nil {
		t.Fatal("private Print implementation leaked into the implicit vocabulary")
	}
}
