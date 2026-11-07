package Lyger

import (
	"encoding/json"
	"testing"
	"testing/fstest"
)

// LAY-007: a regular file is insufficient authority for an indexed source.
func TestLibraryLayoutRegistrySourceRoots(t *testing.T) {
	for _, tc := range []struct {
		path  string
		valid bool
	}{
		{"builtins/eng/storage/extra.l", true},
		{"runtime/eng/extra.l", true},
		{"stdlib/eng/extra.l", true},
		{"fixture/extra.l", false},
		{"runtime/rust/extra.l", false},
		{"runtime/generated/extra.l", false},
		{"builtins/templates/extra.l", false},
		{"builtins/segments/extra.l", false},
	} {
		t.Run(tc.path, func(t *testing.T) {
			files := dictionaryIndexFilesystem(t)
			var index dictionaryIndex
			if err := json.Unmarshal(files["dictionaries.json"].Data, &index); err != nil {
				t.Fatal(err)
			}
			index.Languages["eng"]["extra"] = dictionaryEntry{Entry: tc.path, Dictionary: "Extra"}
			raw, err := json.Marshal(index)
			if err != nil {
				t.Fatal(err)
			}
			files["dictionaries.json"] = &fstest.MapFile{Data: raw}
			files[tc.path] = &fstest.MapFile{Data: []byte("Create \"Extra\" as a Dictionary.\n")}
			_, err = readDictionaryIndex(files)
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%t, error=%v", tc.valid, err)
			}
		})
	}
}
