//go:build darwin || linux

package main

import (
	"bytes"
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"

	distributiontest "github.com/l-programming-language/l/internal/testsupport/distribution"
	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	"github.com/l-programming-language/l/internal/testsupport/phase1"
	Logiceme "github.com/l-programming-language/l/logiceme"
	Lyger "github.com/l-programming-language/l/lyger"
)

func TestParagraphDemoSourceGeneratesSource(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	sources := paragraphDemoSources(t, root)
	project := writeParagraphProject(t, sources, sources["main.l"])
	mainPath := filepath.Join(project, "main.l")
	compiler := newCompiler(&cliData{jobType: "build", mainPath: mainPath, osType: "ubuntu"})
	compiler.initialize(compilerInitOptions{AssetRoot: root})
	if compiler.err != nil {
		t.Fatalf("initialize paragraph source compiler: %v", compiler.err)
	}
	var generated *generatedProgram
	if stdout := sourceStdout(t, func() { generated = compiler.process() }); stdout != "" {
		t.Fatalf("paragraph source compilation wrote stdout: %q", stdout)
	}
	if generated != nil {
		nativetest.RecordBundle(t, "compiler-process", generated.Sources)
	}
	if compiler.err != nil || generated == nil || generated.Compilation == nil || generated.Compilation.Compilation == nil {
		t.Fatalf("paragraph demo did not generate source: %v", compiler.err)
	}
	wantIDs := []string{"HelloWorldApp1", "HelloWorldApp2"}
	if !reflect.DeepEqual(generated.ApplicationIDs, wantIDs) {
		t.Fatalf("paragraph applications = %v; want %v", generated.ApplicationIDs, wantIDs)
	}
	result := generated.Compilation
	if !bytes.Equal(compiler.mainBytes, sources["main.l"]) || !bytes.Equal(compiler.compiledBytes, result.ExpandedSource) || !bytes.Equal(compiler.asmBytes, result.AML) {
		t.Fatal("paragraph compilation changed source or lost its output views")
	}
	statements := 0
	for _, statement := range result.Compilation.SourceStatements {
		if statement.Span.Path == mainPath {
			statements++
		}
	}
	if statements != 13 {
		t.Fatalf("paragraph demo has %d originating main statements; want 13", statements)
	}
	assertFrontendResources(t, generated.Compilation.Compilation)
	bundle := generated.Sources
	if len(bundle.ByteArrays) != 2 || len(bundle.MapArrays) != 2 || len(bundle.DependencyArrays) != 2 || len(bundle.EnvDepArrays) != 2 {
		t.Fatalf("paragraph source bundle dimensions = %d/%d/%d/%d; want two complete applications", len(bundle.ByteArrays), len(bundle.MapArrays), len(bundle.DependencyArrays), len(bundle.EnvDepArrays))
	}
	for index, message := range []string{"Hello World!", "Hello Again World!"} {
		t.Run(wantIDs[index], func(t *testing.T) {
			literal := []byte(strconv.Quote(message))
			for name, markup := range map[string][]byte{"observable": result.AML, "backend": result.BackendAML} {
				if !bytes.Contains(markup, literal) {
					t.Fatalf("%s AML lost paragraph App Text %q", name, message)
				}
			}
			assertMemberSourceBuild(t, Logiceme.Executable{
				ByteArrays:       bundle.ByteArrays[index : index+1],
				MapArrays:        bundle.MapArrays[index : index+1],
				DependencyArrays: bundle.DependencyArrays[index : index+1],
				EnvDepArrays:     bundle.EnvDepArrays[index : index+1],
			}, message)
		})
	}
}

func TestParagraphExamplesGenerateSource(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	sources := paragraphDemoSources(t, root)
	for _, example := range paragraphExamples {
		example := example
		t.Run(example.filename, func(t *testing.T) {
			t.Parallel()
			buildCLICases <- struct{}{}
			t.Cleanup(func() { <-buildCLICases })
			main := paragraphExampleSource(t, root, example.filename)
			project := paragraphExampleProject(t, sources, example.filename, main)
			defer paragraphProjectUnchanged(t, project, example.filename, main, sources)
			compiler := newCompiler(&cliData{jobType: "build", mainPath: filepath.Join(project, example.filename), osType: "ubuntu"})
			compiler.initialize(compilerInitOptions{AssetRoot: root})
			if compiler.err != nil {
				t.Fatal(compiler.err)
			}
			var generated *generatedProgram
			if stdout := sourceStdout(t, func() { generated = compiler.process() }); stdout != "" {
				t.Fatalf("paragraph source generation wrote stdout: %q", stdout)
			}
			if generated != nil {
				nativetest.RecordBundle(t, "compiler-process", generated.Sources)
			}
			if compiler.err != nil || generated == nil || generated.Compilation == nil {
				t.Fatalf("actual paragraph demo did not generate source: %v", compiler.err)
			}
			if len(generated.ApplicationIDs) != len(example.apps) {
				t.Fatalf("application IDs = %v; want %v", generated.ApplicationIDs, example.apps)
			}
			assertFrontendResources(t, generated.Compilation.Compilation)
			bundle := generated.Sources
			if len(bundle.ByteArrays) != len(example.apps) || len(bundle.MapArrays) != len(example.apps) || len(bundle.DependencyArrays) != len(example.apps) || len(bundle.EnvDepArrays) != len(example.apps) {
				t.Fatal("actual paragraph demo lost its parallel source bundle dimensions")
			}
			if !bytes.Equal(compiler.mainBytes, main) || !bytes.Equal(compiler.compiledBytes, generated.Compilation.ExpandedSource) || !bytes.Equal(compiler.asmBytes, generated.Compilation.AML) || (len(example.apps) > 0 && len(compiler.asmBytes) == 0) {
				t.Fatal("actual paragraph demo lost source or observable compilation views")
			}
			if example.filename == "coordinate-existing.l" || example.filename == "local-definitions.l" || example.filename == "local-imports.l" {
				name := "result"
				if example.filename == "local-definitions.l" {
					name = "message"
				}
				outer := generated.Compilation.Compilation.Memory[name]
				if outer == nil || outer.Value == nil || outer.Value.Primitive == nil || outer.Value.Primitive.Text != "`done`" || outer.Value.Runtime != nil {
					t.Fatalf("local storage replaced outer %s required by the generated continuation", name)
				}
			}
			var private []string
			switch example.filename {
			case "local-definitions.l":
				private = []string{"Launch", "Compact", "result", "First", "Second"}
			case "local-imports.l":
				private = []string{"Display", "LocalLibrary", "Paint", "Worker"}
			}
			for name := range generated.Compilation.Compilation.Memory {
				for _, local := range private {
					if strings.EqualFold(name, local) {
						t.Fatalf("scope-private source binding %s leaked into global memory", name)
					}
				}
			}
			var artifacts []Lyger.AppArtifact
			for index, id := range example.apps {
				if generated.ApplicationIDs[index] != id {
					t.Fatalf("application %d = %q; want %q", index, generated.ApplicationIDs[index], id)
				}
				assertParagraphSourceSyntax(t, bundle.ByteArrays[index], bundle.MapArrays[index], &example.printed[index])
				artifacts = append(artifacts, Lyger.AppArtifact{ID: id, RelativeExecutable: fmt.Sprintf("paragraph.apps/app-%03d", index)})
			}
			unit, err := generated.Compilation.OrchestratorUnit("paragraph", artifacts)
			if err != nil {
				t.Fatalf("actual paragraph orchestrator lowering: %v", err)
			}
			orchestrator, err := compiler.log.Evaluate(unit.BackendAML, Logiceme.GenerationOptions{IncludeRuntime: true})
			nativetest.RecordBundle(t, "explicit-orchestrator", orchestrator)
			if err != nil {
				t.Fatalf("actual paragraph orchestrator generation: %v", err)
			}
			if len(orchestrator.ByteArrays) != 1 || len(orchestrator.MapArrays) != 1 {
				t.Fatal("paragraph orchestrator must have one source bundle, including for zero-App programs")
			}
			assertParagraphSourceSyntax(t, orchestrator.ByteArrays[0], orchestrator.MapArrays[0], nil)
		})
	}
}

func TestParagraphBuildCLI(t *testing.T) {
	t.Parallel()
	root := distributiontest.RepoRoot(t)
	sources := paragraphDemoSources(t, root)
	cli := phase1.CLI(t)
	examples := append([]paragraphExample{{filename: "main.l", apps: []string{"HelloWorldApp1", "HelloWorldApp2"}, status: 1, stderr: "text values of both x1 & y should be Hello World\n"}}, paragraphExamples...)
	for _, example := range examples {
		example := example
		t.Run(example.filename, func(t *testing.T) {
			t.Parallel()
			buildCLICases <- struct{}{}
			t.Cleanup(func() { <-buildCLICases })
			main := paragraphExampleSource(t, root, example.filename)
			project := paragraphExampleProject(t, sources, example.filename, main)
			defer paragraphProjectUnchanged(t, project, example.filename, main, sources)
			recorder := nativetest.CLIProbe(t)
			stdout, stderr, err := runBuildCommand(t, cli, project, recorder.Env(nil), "build", example.filename)
			if err != nil {
				t.Fatalf("embedded CLI build prerequisite; runtime not exercised: %v\nstdout=%s\nstderr=%s", err, stdout, stderr)
			}
			binary := filepath.Join(project, strings.TrimSuffix(example.filename, ".l"))
			if string(stdout) != binary+"\n" || len(stderr) != 0 {
				t.Fatalf("build streams=%q/%q", stdout, stderr)
			}
			recorder.AssertPublished(t, binary)
			assertParagraphPublishedApps(t, binary, example.apps)
			stdout, stderr, err = runRuntimeCommand(t, binary, t.TempDir(), nil)
			paragraphExecution(t, stdout, stderr, err, example.status, example.stdout, example.stderr)
		})
	}
}

func assertParagraphPublishedApps(t *testing.T, path string, ids []string) {
	t.Helper()
	nativetest.AssertNativeArtifact(t, path)
	var manifest generationManifest
	manifestPath := nativetest.ReadPublishedGeneration(t, path, &manifest)
	if manifest.GOOS != runtime.GOOS || manifest.GOARCH != runtime.GOARCH || len(manifest.Apps) != len(ids) {
		t.Fatalf("paragraph manifest = %+v; want native applications %v", manifest, ids)
	}
	for index, id := range ids {
		app := manifest.Apps[index]
		if app.ID != id {
			t.Fatalf("paragraph App %d = %q; want %q", index, app.ID, id)
		}
		appPath := filepath.Join(filepath.Dir(manifestPath), app.Executable)
		if app.SHA256 != nativetest.NativeArtifactHash(t, appPath) {
			t.Fatalf("paragraph App %s does not match its manifest hash", id)
		}
	}
}
