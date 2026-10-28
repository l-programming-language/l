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

func TestMemberSourceGeneratesBuildableSource(t *testing.T) {
	t.Parallel()
	root := distributiontest.RepoRoot(t)
	config := Lyger.Options{Language: Lyger.DefaultLanguage, AssetRoot: root, Resources: os.DirFS(root)}
	for _, tc := range []struct {
		name, app, text, owner, aliases, runtimeMembers string
		encodedText                                     string
	}{
		{
			name: "adjacent literal spans", app: "FragmentApp", text: "first``second",
			owner: "Archive's", aliases: "Use Archive's Render as \"Display\".\n",
		},
		{
			name: "mixed literal spans", app: "FragmentUnicodeApp", text: "first 雪`middle`second <DEF> tag",
			owner: "Archive's", aliases: "Use Archive's Render as \"Display\".\n",
		},
		{
			name: "literal marker", app: "MarkerApp", text: "literal <DEF> tag 雪",
			owner: "Archive's", aliases: "Use Archive's Render as \"Display\".\n",
		},
		{
			name: "aliased scoped members", app: "ForecastApp", text: "through copied member 雪",
			owner: "it's", aliases: "Use Archive as \"CopyArchive\".\nUse CopyArchive's Render as \"Display\".\n",
			runtimeMembers: "Execute ForecastApp.\n" +
				"If it has an Error, exit with ForecastApp's Error, else define \"response\" as ForecastApp's Output.\n" +
				"If message isnt response, exit with an error of `unexpected output`.\n" +
				"Otherwise, show response and then exit.\n",
		},
		{
			name: "encoded backend delimiters", app: "EncodingApp",
			text:        `雪 <DEF> |} [|Meta.Key|] [|Meta.Statement|] </S|> </W|> a~, b c: ~d  e   f (/={tags: tags}=={for Header in Headers}=={header}=/) /*|demo|*\ end`,
			encodedText: `"雪 <DEF> \x7c} [\x7cMeta.Key\x7c] [\x7cMeta.Statement\x7c] </S\x7c> </W\x7c> a\x7e, b c: \x7ed\x20\x20e\x20\x20 f (\x2f={tags: tags}=={for Header in Headers}=={header}=/) /*\x7cdemo\x7c*\\ end"`,
			owner:       "Archive's", aliases: "Use Archive's Render as \"Display\".\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			directory := t.TempDir()
			library := "Use IO's Print from `io` as \"Emit\".\nCreate \"Archive\" as a Dictionary.\nUse Emit as " + tc.owner + " \"Render\".\n"
			body := tc.aliases + "\"message\" is `" + tc.text + "`.\n" +
				"Create \"" + tc.app + "\" as a New Basic App.\nThe App Displays message.\n" + tc.runtimeMembers
			source := []byte("Use `library.l`.\n" + body)
			path := filepath.Join(directory, "story.l")
			for name, content := range map[string][]byte{"library.l": []byte(library), "story.l": source} {
				if err := os.WriteFile(filepath.Join(directory, name), content, 0600); err != nil {
					t.Fatal(err)
				}
			}
			compiler := newCompiler(&cliData{jobType: "build", mainPath: path, osType: "ubuntu"})
			compiler.options = config
			compiler.lyger = Lyger.NewLygerService(config)
			compiler.log = Logiceme.NewLogicemeService(config.Resources)
			compiler.compilation = &Lyger.Compilation{}
			compiler.mainBytes = source
			var generated *generatedProgram
			if output := sourceStdout(t, func() { generated = compiler.process() }); output != "" {
				t.Errorf("source generation wrote compiler stdout: %q", output)
			}
			if generated != nil {
				nativetest.RecordBundle(t, "compiler-process", generated.Sources)
			}
			if compiler.err != nil || generated == nil || generated.Compilation == nil {
				t.Fatalf("source did not generate a bundle: %v", compiler.err)
			}
			result := generated.Compilation
			assertFrontendResources(t, result.Compilation)
			wantExpanded := []byte(library + strings.TrimSuffix(body, "\n"))
			var local []string
			for _, statement := range result.Compilation.SourceStatements {
				if statement.Span.Path == path || statement.Span.Path == filepath.Join(directory, "library.l") {
					local = append(local, strings.TrimSpace(string(statement.Raw)))
				}
			}
			if actual := strings.Join(local, "\n"); actual != string(wantExpanded) {
				t.Errorf("originating local source lost literal, member owner, or declaration order:\ngot: %s\nwant: %s", actual, wantExpanded)
			}
			if !bytes.Equal(compiler.mainBytes, source) || !bytes.Equal(compiler.compiledBytes, result.ExpandedSource) || !bytes.Equal(compiler.asmBytes, result.AML) {
				t.Error("compiler changed the source or lost its returned output views")
			}
			if result.Compilation == nil || result.Compilation.Program == nil || len(result.Compilation.Program.Books) != 1 || result.Compilation.Program.Books[0].Name != tc.app {
				t.Fatalf("source application identity was not retained")
			}
			message := result.Compilation.Memory["message"]
			if message == nil || message.Value == nil || message.Value.Primitive == nil {
				t.Fatal("source Text is missing from semantic memory")
			}
			if got, want := message.Value.Primitive.Text, "`"+tc.text+"`"; got != want {
				t.Errorf("semantic Text = %q; want original source value %q", got, want)
			}
			literal := []byte(strconv.Quote(tc.text))
			if tc.encodedText != "" {
				literal = []byte(tc.encodedText)
			}
			for name, markup := range map[string][]byte{"observable": result.AML, "backend": result.BackendAML} {
				if !bytes.Contains(markup, literal) {
					t.Errorf("%s AML lost source Text %q", name, tc.text)
				}
			}
			assertMemberSourceBuild(t, generated.Sources, tc.text)
		})
	}
}
