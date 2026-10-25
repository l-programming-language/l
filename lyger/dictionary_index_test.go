package Lyger

import (
	"errors"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	distribution "github.com/l-programming-language/l"
	"github.com/l-programming-language/l/lerrors"
)

func dictionaryIndexFilesystem(t *testing.T) fstest.MapFS {
	t.Helper()
	files := fstest.MapFS{}
	if err := fs.WalkDir(distribution.Files(), ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			data, err := fs.ReadFile(distribution.Files(), name)
			if err != nil {
				return err
			}
			files[name] = &fstest.MapFile{Data: data}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return files
}

func loadIndexDictionary(t *testing.T, files fstest.MapFS) error {
	t.Helper()
	config := Options{Resources: files}
	grammar, err := loadGrammar(config)
	if err != nil {
		t.Fatal(err)
	}
	_, err = newPreparedBuilder(nil, grammar).loadDictionary("io", map[string]bool{}, nil)
	return err
}

func TestDictionaryIndexValidation(t *testing.T) {
	const entry = `{"entry":"stdlib/eng/io/dictionary.l","dictionary":"IO"}`
	const normal = `{"io":` + entry + `}`
	for _, tc := range []struct {
		name, index, context string
		kind                 error
	}{
		{"duplicate logical name", `{"format":1,"languages":{"eng":{"io":` + entry + `,"io":` + entry + `}}}`, "io", lerrors.InvalidConfiguration},
		{"escaped duplicate name", `{"format":1,"languages":{"eng":{"io":` + entry + `,"i\u006f":` + entry + `}}}`, "io", lerrors.InvalidConfiguration},
		{"duplicate language", `{"format":1,"languages":{"eng":` + normal + `,"eng":` + normal + `}}`, "eng", lerrors.InvalidConfiguration},
		{"duplicate entry field", `{"format":1,"languages":{"eng":{"io":{"entry":"stdlib/eng/io/dictionary.l","entry":"stdlib/eng/io/dictionary.l","dictionary":"IO"}}}}`, "entry", lerrors.InvalidConfiguration},
		{"backslash logical name", `{"format":1,"languages":{"eng":{"io":` + entry + `,"bad\\name":` + entry + `}}}`, "bad", lerrors.InvalidConfiguration},
		{"backslash resource entry", `{"format":1,"languages":{"eng":{"io":` + entry + `,"bad":{"entry":"stdlib/eng/bad\\name.l","dictionary":"Bad"}}}}`, "bad", lerrors.InvalidConfiguration},
		{"backslash language", `{"format":1,"languages":{"eng":` + normal + `,"bad\\name":` + normal + `}}`, "bad", lerrors.InvalidConfiguration},
		{"missing unused entry", `{"format":1,"languages":{"eng":{"io":` + entry + `,"missing":{"entry":"stdlib/eng/missing/dictionary.l","dictionary":"Missing"}}}}`, "stdlib/eng/missing/dictionary.l", fs.ErrNotExist},
		{"directory unused entry", `{"format":1,"languages":{"eng":{"io":` + entry + `,"directory":{"entry":"stdlib/eng/directory.l","dictionary":"Directory"}}}}`, "stdlib/eng/directory.l", lerrors.InvalidConfiguration},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files := dictionaryIndexFilesystem(t)
			if err := loadIndexDictionary(t, files); err != nil {
				t.Fatalf("valid dictionary prerequisite: %v", err)
			}
			files["stdlib/index.json"] = &fstest.MapFile{Data: []byte(tc.index)}
			files["stdlib/eng/directory.l"] = &fstest.MapFile{Mode: fs.ModeDir}
			err := loadIndexDictionary(t, files)
			if err == nil || !errors.Is(err, tc.kind) || !strings.Contains(err.Error(), tc.context) || !strings.Contains(err.Error(), "stdlib/index.json") {
				t.Fatalf("invalid registry must fail before selected dictionary expansion, with index/entry provenance: %v", err)
			}
		})
	}
}
