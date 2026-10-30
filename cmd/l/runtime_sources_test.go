//go:build darwin || linux

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	distributiontest "github.com/l-programming-language/l/internal/testsupport/distribution"
	Logiceme "github.com/l-programming-language/l/logiceme"
)

// Only shared runtime support comes from production. Independent scenario
// sources, manifests, locks and expected behavior remain authored fixtures.
func canonicalRuntimeSources(t testing.TB) map[string][]byte {
	t.Helper()
	root := distributiontest.RepoRoot(t)
	input := []byte(`<PROGRAM><STRUCTURE><BOOK "support"><OUTLINE>
<JSON>{"AppMeta":{"AppGenDir":"/src"}}</JSON>
<JSON>{"Services":[{"Name":"main","Location":"/src/","GenericObjects":["main"]}]}</JSON>
<JSON>{"Objects":[{"Name":"main","Type":"root"}]}</JSON>
</OUTLINE><CONTENT><CHAPTER "main"><SENTENCE "main"><|W|><|S|>` + "`A`" + `</S|>
<|S|>` + "`FW{|Key: ~l_main~|}^W`" + `</S|></W|></SENTENCE></CHAPTER></CONTENT></BOOK></STRUCTURE></PROGRAM>`)
	bundle, err := Logiceme.NewLogicemeService(os.DirFS(root)).Evaluate(input, Logiceme.GenerationOptions{IncludeRuntime: true})
	if err != nil {
		t.Fatal(err)
	}
	files := make(map[string][]byte)
	for _, wanted := range []string{"runtime.rs", "runtime/host.rs", "process_unix.rs", "process_other.rs"} {
		for i, name := range bundle.MapArrays[0] {
			if name == "/src/"+wanted {
				files[wanted] = bytes.Clone(bundle.ByteArrays[0][i])
			}
		}
		if len(files[wanted]) == 0 {
			t.Fatalf("selected support omitted %s", wanted)
		}
	}
	for name, asset := range map[string]string{"runtime/host.rs": "host", "process_unix.rs": "process_unix", "process_other.rs": "process_other"} {
		want, err := os.ReadFile(filepath.Join(root, "builtins/templates/runtime", asset+".lu"))
		if err != nil || !bytes.Equal(files[name], want) {
			t.Fatalf("selected support changed canonical %s bytes: %v", name, err)
		}
	}
	return files
}
