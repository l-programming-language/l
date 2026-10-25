//go:build darwin || linux

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	distributiontest "github.com/l-programming-language/l/internal/testsupport/distribution"
	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	Logiceme "github.com/l-programming-language/l/logiceme"
	Lyger "github.com/l-programming-language/l/lyger"
)

func TestDemoGeneratesBuildableSource(t *testing.T) {
	t.Parallel()
	root := distributiontest.RepoRoot(t)
	config := Lyger.Options{Language: Lyger.DefaultLanguage, AssetRoot: root, Resources: os.DirFS(root)}
	original, err := os.ReadFile(filepath.Join(root, "demo/basic/main.l"))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, app, text string }{
		{"demo", "Hello", "Hello World!"},
		{"different source", "WeatherReport", `Rain, "quoted" snow 雪!`},
		{"renamed vocabulary", "RenamedApplication", "source defined vocabulary"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := bytes.Clone(original)
			sourcePath := filepath.Join(root, "demo/basic/main.l")
			if tc.name != "demo" {
				source = bytes.ReplaceAll(source, []byte(`"Hello"`), []byte(strconv.Quote(tc.app)))
				source = bytes.ReplaceAll(source, []byte("Execute Hello."), []byte("Execute "+tc.app+"."))
				source = bytes.ReplaceAll(source, []byte("Show Hello's"), []byte("Show "+tc.app+"'s"))
				source = bytes.Replace(source, []byte("`Hello World!`"), []byte("`"+tc.text+"`"), 1)
			}
			if tc.name == "renamed vocabulary" {
				directory := t.TempDir()
				source = []byte("Use IO's Print from `io` as \"Display\".\nUse Execute as \"Launch\".\nUse Basic as \"Plain\".\n\"message\" is `" + tc.text + "`.\nCreate \"" + tc.app + "\" as a New Plain App.\nThe App Displays message.\nLaunch it.\nShow it's Output.\n")
				sourcePath = filepath.Join(directory, "story.l")
			}
			compiler := newCompiler(&cliData{jobType: "build", mainPath: sourcePath, osType: "ubuntu"})
			compiler.options = config
			compiler.lyger = Lyger.NewLygerService(config)
			compiler.log = Logiceme.NewLogicemeService(config.Resources)
			compiler.compilation = &Lyger.Compilation{}
			compiler.mainBytes = source
			var generated *generatedProgram
			if stdout := sourceStdout(t, func() { generated = compiler.process() }); stdout != "" {
				t.Fatalf("source generation wrote compiler diagnostics to stdout: %q", stdout)
			}
			if generated != nil {
				nativetest.RecordBundle(t, "compiler-process", generated.Sources)
			}
			if compiler.err != nil || generated == nil {
				t.Fatalf("actual source generation: %v", compiler.err)
			}
			if len(generated.Compilation.Compilation.Program.Books) != 1 || generated.Compilation.Compilation.Program.Books[0].Name != tc.app {
				t.Fatalf("source application name did not reach application graph")
			}
			if len(generated.Sources.ByteArrays) != 1 || len(generated.Sources.MapArrays) != 1 {
				t.Fatalf("unexpected source bundle shape")
			}
			assertFrontendResources(t, generated.Compilation.Compilation)
			assertMemberSourceBuild(t, generated.Sources, tc.text)
			nameInBuildSource := false
			for i, content := range generated.Sources.ByteArrays[0] {
				if !strings.HasSuffix(generated.Sources.MapArrays[0][i], ".rs") && bytes.Contains(content, []byte(tc.app)) {
					nameInBuildSource = true
				}
			}
			if !nameInBuildSource {
				t.Fatal("source application name did not reach backend build source")
			}
			if evidence := os.Getenv("L_DEMO_SOURCE_EVIDENCE_DIR"); evidence != "" && tc.name == "demo" {
				sourceEvidenceMu.Lock()
				defer sourceEvidenceMu.Unlock()
				if err := os.MkdirAll(evidence, 0700); err != nil {
					t.Fatal(err)
				}
				for i, p := range generated.Sources.MapArrays[0] {
					if strings.HasSuffix(p, ".rs") {
						data := generated.Sources.ByteArrays[0][i]
						dest := filepath.Join(evidence, strings.TrimPrefix(p, "/"))
						if err := os.MkdirAll(filepath.Dir(dest), 0700); err != nil {
							t.Fatal(err)
						}
						if err := os.WriteFile(dest, data, 0600); err != nil {
							t.Fatal(err)
						}
					}
				}
				for name, data := range map[string][]byte{"frontend.aml": generated.Compilation.AML, "backend.aml": generated.Compilation.BackendAML, "expanded.l": generated.Compilation.ExpandedSource, "build.txt": []byte("PASS: actual Cargo build --frozen, native artifact verified; binary not executed.\n")} {
					if err := os.WriteFile(filepath.Join(evidence, name), data, 0600); err != nil {
						t.Fatal(err)
					}
				}
			}
		})
	}
}
