package Logiceme

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	distribution "github.com/l-programming-language/l"
	"github.com/l-programming-language/l/lerrors"
)

func compositionResources(t *testing.T, value string) fstest.MapFS {
	t.Helper()
	resources := runtimeGenerationResources(t)
	files := fstest.MapFS{}
	for _, root := range []string{"builtins/segments", "builtins/templates"} {
		if err := fs.WalkDir(resources, root, func(name string, entry fs.DirEntry, err error) error {
			if err != nil || entry.IsDir() {
				return err
			}
			data, err := fs.ReadFile(resources, name)
			if err == nil {
				files[name] = &fstest.MapFile{Data: bytes.Clone(data), Mode: 0600}
			}
			return err
		}); err != nil {
			t.Fatal(err)
		}
	}
	const sourcePath = "builtins/runtime/eng/number.l"
	source := []byte("source-owned composition " + value)
	aml := []byte(`<PROGRAM><STRUCTURE><BOOK "support"><OUTLINE><JSON>{"AppMeta":{"AppGenDir":"/src","StructureVersion":"2"}}</JSON><JSON>{"Services":[{"Name":"library","Location":"/src/","GenericObjects":["runtime"],"Target":{"Kind":"lib","Name":"support","Root":"runtime","Primary":false}}]}</JSON><JSON>{"Objects":[{"Name":"runtime","Type":"root"}]}</JSON></OUTLINE><SETTINGS></SETTINGS><CONTENT><CHAPTER "library"><SENTENCE "runtime"><|W|><|S|>` + "`A`" + `</S|><|S|>` + "`FW{|Key: ~identity~, Outputs: ~&[u8]&~|}^W^BOD^E{|Statement: ~return " + value + ";~|}`" + `</S|></W|></SENTENCE></CHAPTER></CONTENT></BOOK></STRUCTURE></PROGRAM>`)
	manifest, err := json.Marshal(distribution.RuntimeComposition{Version: 1, SHA256: compositionHash(aml), Sources: map[string]string{sourcePath: compositionHash(source)}})
	if err != nil {
		t.Fatal(err)
	}
	files[sourcePath] = &fstest.MapFile{Data: source, Mode: 0600}
	files["builtins/runtime/support.aml"] = &fstest.MapFile{Data: aml, Mode: 0600}
	files["builtins/runtime/support.json"] = &fstest.MapFile{Data: manifest, Mode: 0600}
	// There must be no dependency on the former static runtime source.
	return files
}

func compositionHash(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func TestRuntimeCompositionUsesSelectedResources(t *testing.T) {
	first, err := loadSupportTemplates(compositionResources(t, "7"), true)
	if err != nil {
		t.Fatal(err)
	}
	second, err := loadSupportTemplates(compositionResources(t, "9"), true)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(first["runtime/runtime.rs"], []byte("return 7;")) || !bytes.Contains(second["runtime/runtime.rs"], []byte("return 9;")) {
		t.Fatal("selected AML did not control composed support")
	}
	first["runtime/runtime.rs"][0] = '!'
	if second["runtime/runtime.rs"][0] == '!' {
		t.Fatal("generated runtime support aliases another provider")
	}
}

func TestRuntimeCompositionRejectsStaleInputs(t *testing.T) {
	for _, name := range []string{"builtins/runtime/eng/number.l", "builtins/runtime/support.aml"} {
		t.Run(name, func(t *testing.T) {
			files := compositionResources(t, "7")
			files[name].Data = append(files[name].Data, '\n')
			if _, err := loadSupportTemplates(files, true); !errors.Is(err, lerrors.InvalidConfiguration) || !strings.Contains(err.Error(), name) || !strings.Contains(err.Error(), "stale") {
				t.Fatalf("stale selected input = %v", err)
			}
		})
	}
	files := compositionResources(t, "7")
	delete(files, "builtins/runtime/support.aml")
	if _, err := loadSupportTemplates(files, true); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("missing selected composition = %v", err)
	}
	if _, err := loadSupportTemplates(files, false); err != nil {
		t.Fatalf("ordinary generation unexpectedly requires runtime: %v", err)
	}
}

func TestRuntimeCompositionRequiresLibraryTarget(t *testing.T) {
	files := compositionResources(t, "7")
	files["builtins/runtime/support.aml"].Data = bytes.Replace(files["builtins/runtime/support.aml"].Data, []byte(`"Kind":"lib"`), []byte(`"Kind":"bin"`), 1)
	var manifest distribution.RuntimeComposition
	if err := json.Unmarshal(files["builtins/runtime/support.json"].Data, &manifest); err != nil {
		t.Fatal(err)
	}
	manifest.SHA256 = compositionHash(files["builtins/runtime/support.aml"].Data)
	updated, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	files["builtins/runtime/support.json"].Data = updated
	if _, err := loadSupportTemplates(files, true); !errors.Is(err, lerrors.InvalidConfiguration) || !strings.Contains(err.Error(), "library target") {
		t.Fatalf("binary composition = %v", err)
	}
}
