package Logiceme

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestSegmentDefaultsIsolation(t *testing.T) {
	resources := runtimeGenerationResources(t)
	first, err := loadSegmentalSystem(resources)
	if err != nil {
		t.Fatal(err)
	}
	second, err := loadSegmentalSystem(resources)
	if err != nil {
		t.Fatal(err)
	}
	wantMeta := MetaDict{
		Index: MetaDictIndex{Keys: []string{""}, Indexes: []int{0}},
		Metas: []*MetaObj{{Value: "string", Type: "Default`"}},
	}
	for _, segment := range []Segment{first.Execute, first.Declaration, second.Execute} {
		if !reflect.DeepEqual(segment.Meta, wantMeta) || !reflect.DeepEqual(segment.Inputs, []string{"Type"}) {
			t.Fatalf("default segment changed its effective inputs/metadata: %#v", segment)
		}
	}
	first.Execute.Inputs[0] = "private mutation"
	first.Execute.Meta.Index.Keys[0] = "private mutation"
	first.Execute.Meta.Index.Indexes[0] = 1
	first.Execute.Meta.Metas[0].Value = "private mutation"
	for _, segment := range []Segment{first.Declaration, second.Execute} {
		if !reflect.DeepEqual(segment.Meta, wantMeta) || !reflect.DeepEqual(segment.Inputs, []string{"Type"}) {
			t.Fatal("segment defaults share mutable inputs or metadata")
		}
	}
}

func TestBackendAssetConfigurationErrors(t *testing.T) {
	config := runtimeGenerationResources(t)
	missing := filepath.Join(t.TempDir(), "absent")
	if _, err := loadSupportTemplates(os.DirFS(missing), false); err == nil || !errors.Is(err, os.ErrNotExist) || !strings.Contains(err.Error(), "builtins/templates/config/config.lu") {
		t.Fatalf("template root error lost path or cause: %v", err)
	}
	if _, err := NewLogicemeService(os.DirFS(missing)).Evaluate(nil); err == nil || !errors.Is(err, os.ErrNotExist) || !strings.Contains(err.Error(), "builtins/segments/consonants/execute.lu") {
		t.Fatalf("generation asset root error lost path or cause: %v", err)
	}
	if _, err := NewLogicemeService(config).Evaluate(nil, GenerationOptions{}, GenerationOptions{}); err == nil {
		t.Fatal("ambiguous generation options accepted")
	}
}
