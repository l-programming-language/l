package Lyger

import (
	"encoding/json"
	"strings"
	"testing"
	"testing/fstest"
)

func TestBuiltinPrivateDictionaryImport(t *testing.T) {
	files := dictionaryIndexFilesystem(t)
	var index dictionaryIndex
	if err := json.Unmarshal(files["stdlib/index.json"].Data, &index); err != nil {
		t.Fatal(err)
	}
	index.Languages["eng"]["internal"] = dictionaryEntry{Entry: "stdlib/eng/internal/dictionary.l", Dictionary: "Internal"}
	raw, err := json.Marshal(index)
	if err != nil {
		t.Fatal(err)
	}
	files["stdlib/index.json"] = &fstest.MapFile{Data: raw}
	files["stdlib/eng/internal/dictionary.l"] = &fstest.MapFile{Data: []byte("Create \"Internal\" as a Dictionary.\nDefine Internal's \"Place\" as `<DEF>OB{|Key: ~<@Name@>~, Value: ~<@Value@>~|}<DEF>`.\n")}
	files["builtins/eng/core.l"].Data = append(files["builtins/eng/core.l"].Data, []byte("\n    Use Internal from `internal`.\n\n")...)
	service := NewLygerService(Options{Resources: files})
	c := service.NewCompilation([]byte("Create \"Worker\" as a New Basic App.\n"), "private.l")
	c.Make()
	if c.Error != nil {
		t.Fatal(c.Error)
	}
	bound, err := c.instantiateBuiltin("Internal.Place", map[string]DefinitionArgument{"Name": definitionReferenceArgument("target"), "Value": definitionReferenceArgument("source")})
	if err != nil || !strings.Contains(bound.Expression, "Value: ~source~") {
		t.Fatalf("selected private composition: %s, %v", bound.Expression, err)
	}
	for _, unit := range c.SourceUnits {
		if unit.Kind != SourceBuiltin {
			continue
		}
		binding, _, err := unit.root.lookupLocal([]byte("Internal"))
		if err != nil || binding != nil {
			t.Fatalf("private dictionary leaked into builtin globals: %v, %v", binding, err)
		}
	}
	invalid := service.NewCompilation([]byte("Use Internal as \"Exposed\".\n"), "private.l")
	invalid.Make()
	if invalid.Error == nil {
		t.Fatal("private dictionary resolved without a user import")
	}
}
