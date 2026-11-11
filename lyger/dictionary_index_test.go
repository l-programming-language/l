package Lyger

import (
	"encoding/json"
	"errors"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	distributiontest "github.com/l-programming-language/l/internal/testsupport/distribution"
	"github.com/l-programming-language/l/lerrors"
)

func dictionaryIndexFilesystem(t *testing.T) fstest.MapFS {
	t.Helper()
	return distributiontest.Files(t)
}

func loadIndexDictionary(t *testing.T, files fstest.MapFS) error {
	t.Helper()
	config := Options{Resources: files}
	grammar, err := loadGrammar(config)
	if err != nil {
		t.Fatal(err)
	}
	_, err = newPreparedBuilder(nil, grammar).loadDictionary("probe", map[string]bool{}, nil)
	return err
}

func TestDictionaryIndexValidation(t *testing.T) {
	const entry = `{"entry":"stdlib/eng/probe.l","dictionary":"Probe"}`
	const normal = `{"probe":` + entry + `}`
	for _, tc := range []struct {
		name, index, context string
		kind                 error
	}{
		{"duplicate logical name", `{"format":1,"languages":{"eng":{"probe":` + entry + `,"probe":` + entry + `}}}`, "probe", lerrors.InvalidConfiguration},
		{"escaped duplicate name", `{"format":1,"languages":{"eng":{"probe":` + entry + `,"pr\u006fbe":` + entry + `}}}`, "probe", lerrors.InvalidConfiguration},
		{"duplicate language", `{"format":1,"languages":{"eng":` + normal + `,"eng":` + normal + `}}`, "eng", lerrors.InvalidConfiguration},
		{"duplicate entry field", `{"format":1,"languages":{"eng":{"probe":{"entry":"stdlib/eng/probe.l","entry":"stdlib/eng/probe.l","dictionary":"Probe"}}}}`, "entry", lerrors.InvalidConfiguration},
		{"backslash logical name", `{"format":1,"languages":{"eng":{"probe":` + entry + `,"bad\\name":` + entry + `}}}`, "bad", lerrors.InvalidConfiguration},
		{"backslash resource entry", `{"format":1,"languages":{"eng":{"probe":` + entry + `,"bad":{"entry":"stdlib/eng/bad\\name.l","dictionary":"Bad"}}}}`, "bad", lerrors.InvalidConfiguration},
		{"backslash language", `{"format":1,"languages":{"eng":` + normal + `,"bad\\name":` + normal + `}}`, "bad", lerrors.InvalidConfiguration},
		{"missing unused entry", `{"format":1,"languages":{"eng":{"probe":` + entry + `,"missing":{"entry":"stdlib/eng/missing/dictionary.l","dictionary":"Missing"}}}}`, "stdlib/eng/missing/dictionary.l", fs.ErrNotExist},
		{"directory unused entry", `{"format":1,"languages":{"eng":{"probe":` + entry + `,"directory":{"entry":"stdlib/eng/directory.l","dictionary":"Directory"}}}}`, "stdlib/eng/directory.l", lerrors.InvalidConfiguration},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files := dictionaryIndexFilesystem(t)
			files["dictionaries.json"] = &fstest.MapFile{Data: []byte(`{"format":1,"languages":{"eng":{"probe":{"entry":"stdlib/eng/probe.l","dictionary":"Probe"}}}}`)}
			files["stdlib/eng/probe.l"] = &fstest.MapFile{Data: []byte("Create \"Probe\" as a Dictionary.\n")}
			if err := loadIndexDictionary(t, files); err != nil {
				t.Fatalf("valid dictionary prerequisite: %v", err)
			}
			files["dictionaries.json"] = &fstest.MapFile{Data: []byte(tc.index)}
			files["stdlib/eng/directory.l"] = &fstest.MapFile{Mode: fs.ModeDir}
			err := loadIndexDictionary(t, files)
			if err == nil || !errors.Is(err, tc.kind) || !strings.Contains(err.Error(), tc.context) || !strings.Contains(err.Error(), "dictionaries.json") {
				t.Fatalf("invalid registry must fail before selected dictionary expansion, with index/entry provenance: %v", err)
			}
		})
	}
}

// installTestDictionary adds an explicitly owned fixture to the selected provider.
func installTestDictionary(t *testing.T, files fstest.MapFS, key, name, source string) string {
	t.Helper()
	var index dictionaryIndex
	if err := json.Unmarshal(files["dictionaries.json"].Data, &index); err != nil {
		t.Fatal(err)
	}
	entry := "stdlib/eng/" + key + "/dictionary.l"
	index.Languages["eng"][key] = dictionaryEntry{Entry: entry, Dictionary: name}
	data, err := json.Marshal(index)
	if err != nil {
		t.Fatal(err)
	}
	files["dictionaries.json"] = &fstest.MapFile{Data: data}
	files[entry] = &fstest.MapFile{Data: []byte(source)}
	return entry
}
