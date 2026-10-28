package Lyger

import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestNamedCompositionUsesConstructorMetadata(t *testing.T) {
	c := NewLygerService(canonicalTestOptions()).NewCompilation([]byte("Create \"Worker\" as a New Basic App.\n"), "binding.l")
	c.Make()
	if c.Error != nil {
		t.Fatal(c.Error)
	}
	for _, test := range []struct{ label, fragment string }{{"plain", "Key: ~target~"}, {"mutable", "Key: ~mut target~"}} {
		t.Run(test.label, func(t *testing.T) {
			var descriptors []string
			if test.label == "mutable" {
				descriptors = []string{"Storage.Mutable"}
			}
			bound, err := c.instantiateDescribedBuiltin("Storage.Binding", descriptors, map[string]DefinitionArgument{"Name": definitionReferenceArgument("target"), "Value": definitionReferenceArgument("source")})
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(bound.Expression, test.fragment) || !strings.Contains(bound.Expression, "Value: ~source~") {
				t.Fatalf("bound composition lost atoms: %s", bound.Expression)
			}
			if _, err = c.instantiateDescribedBuiltin("Storage.Binding", descriptors, map[string]DefinitionArgument{"Name": definitionReferenceArgument("target"), "Value": definitionReferenceArgument("source"), "Unused": definitionReferenceArgument("extra")}); err == nil {
				t.Fatal("unused named atom accepted")
			}
		})
	}
}

func TestNamedCompositionSelectsScopedWords(t *testing.T) {
	files := dictionaryIndexFilesystem(t)
	const core = "stdlib/eng/storage/dictionary.l"
	original := string(files[core].Data)
	original += "\nDefine Storage's \"Place\" as `<DEF>OB{|Key: ~<@Name@>~, Value: ~<@Value@>~|}<DEF>`.\nCreate Storage's \"Keep|-s|-t|-ing|-er\" for Processing a New Storage's Place.\nDefine Storage's \"Unique\" as a New Adjective describing any Storage's Place.\nIt describes any Storage's Place as `<DEF>OB{|Key: ~mut <@Name@>~, Value: ~<@Value@>~|}<DEF>`.\n"
	files[core] = &fstest.MapFile{Data: []byte(original)}
	c := NewLygerService(Options{Resources: files}).NewCompilation([]byte("Create \"Worker\" as a New Basic App.\n"), "binding.l")
	c.Make()
	if c.Error != nil {
		t.Fatal(c.Error)
	}
	bound, err := c.instantiateBuiltin("Storage.Keep", map[string]DefinitionArgument{"Name": definitionReferenceArgument("target"), "Value": definitionReferenceArgument("source")})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(bound.Expression, "Value: ~source~") {
		t.Fatalf("scoped composition lost value: %s", bound.Expression)
	}
	bound, err = c.instantiateDescribedBuiltin("Storage.Place", []string{"Storage.Unique"}, map[string]DefinitionArgument{"Name": definitionReferenceArgument("target"), "Value": definitionReferenceArgument("source")})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(bound.Expression, "Key: ~mut target~") {
		t.Fatalf("scoped descriptor lost name: %s", bound.Expression)
	}
}
