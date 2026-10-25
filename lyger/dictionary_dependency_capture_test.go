package Lyger

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/l-programming-language/l/lerrors"
)

func TestDictionaryImportsCannotCaptureCallerAddedMembers(t *testing.T) {
	files := dictionaryIndexFilesystem(t)
	var index dictionaryIndex
	if err := json.Unmarshal(files["stdlib/index.json"].Data, &index); err != nil {
		t.Fatal(err)
	}
	const entry = "stdlib/eng/tools/dictionary.l"
	index.Languages["eng"]["tools"] = dictionaryEntry{Entry: entry, Dictionary: "Tools"}
	encoded, err := json.Marshal(index)
	if err != nil {
		t.Fatal(err)
	}
	files["stdlib/index.json"] = &fstest.MapFile{Data: encoded}
	const prefix = "Create \"Tools\" as a Dictionary.\n"
	files[entry] = &fstest.MapFile{Data: []byte(prefix + "Define Tools's \"Gift\" as `module|`.\n")}
	cfg := Options{Resources: files}
	main := filepath.Join(t.TempDir(), "main.l")
	const imports = "Use IO from `io`.\nUse Tools's Gift from `tools` as \"Gift\".\n"
	if _, err := NewLygerService(cfg).Compile([]byte(imports+"Show Gift.\n"), main); err != nil {
		t.Fatalf("valid independent export prerequisite: %v", err)
	}
	files[entry] = &fstest.MapFile{Data: []byte(prefix + "Use IO's Gift from `io` as \"ImportedGift\".\nUse ImportedGift as Tools's \"Gift\".\n")}
	source := imports + "Define IO's \"Gift\" as `caller|`.\nShow Gift.\n"
	result, err := NewLygerService(cfg).Compile([]byte(source), main)
	if result != nil || !errors.Is(err, lerrors.Unsupported) || !strings.Contains(err.Error(), "cannot modify dictionary stdlib/eng/io/dictionary.l") || !strings.Contains(err.Error(), "main.l:3:1") {
		t.Fatalf("caller-added namespace member must not satisfy dictionary export dependency: result-present=%t error=%v", result != nil, err)
	}
}
