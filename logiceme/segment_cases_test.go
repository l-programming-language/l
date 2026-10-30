package Logiceme

import (
	"bytes"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func TestSelectedSegmentCasesMatchFreshParsing(t *testing.T) {
	system, err := loadSegmentalSystem(runtimeGenerationResources(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, definition := range segmentDefinitions {
		if definition.family != "Consonant" && definition.family != "Vowel" {
			continue
		}
		segment := system.SegmentSwitch(definition.symbol)
		for _, candidate := range system.templateCases[definition.symbol].cases {
			if !strings.HasPrefix(candidate.condition, `if Type == "`) {
				continue
			}
			selected := strings.TrimSuffix(strings.TrimPrefix(candidate.condition, `if Type == "`), `"`)
			for _, headers := range []bool{false, true} {
				var meta MetaDict
				meta.Add("Process", "selected child", "Conjoinment")
				want := ParseSegmentTags(CompileMainTags(segment.Content), meta, selected, headers)
				got := parseSegmentTags(segment.selectedTemplateTags(&system, selected, headers), meta, selected, headers, true)
				if !reflect.DeepEqual(got, want) || !segment.hasTemplateCase(&system, selected) {
					t.Fatalf("%s/%s/headers=%v changed selected tags", definition.symbol, selected, headers)
				}
				// A consumer may mutate its parsed result. The next glyph must
				// still obtain the original selected resource and metadata.
				owned := got[len(got)-1]
				owned.Expressions[2].ExpressionString = "changed by consumer"
				owned.TagMeta.AppMeta.Add("Owned", "glyph", "String")
				if owned.Expressions[2].TagMeta != owned.TagMeta {
					t.Fatal("selected case lost its shared local metadata identity")
				}
				if children := owned.Expressions[2].ExpressionSubTags; len(children) != 0 {
					children[0].ReplaceString = "changed child"
				}
				again := parseSegmentTags(segment.selectedTemplateTags(&system, selected, headers), meta, selected, headers, true)
				if !reflect.DeepEqual(again, want) {
					t.Fatalf("%s/%s reused a mutable selected case", definition.symbol, selected)
				}
			}
		}
	}
}

func TestSelectedSegmentCasesRespectContentAndTagReplacement(t *testing.T) {
	system, err := loadSegmentalSystem(runtimeGenerationResources(t))
	if err != nil {
		t.Fatal(err)
	}
	segment := system.Operation
	segment.Content = bytes.ReplaceAll(segment.Content, []byte(`if Type == "Default"`), []byte(`if Type == "Changed"`))
	if segment.hasTemplateCase(&system, "Default") || !segment.hasTemplateCase(&system, "Changed") {
		t.Fatal("direct Content replacement reused previously loaded cases")
	}
	want := ParseSegmentTags(CompileMainTags(segment.Content), MetaDict{}, "Changed", false)
	got := parseSegmentTags(segment.selectedTemplateTags(&system, "Changed", false), MetaDict{}, "Changed", false, true)
	if !reflect.DeepEqual(got, want) {
		t.Fatal("direct Content replacement changed fresh parsing")
	}
	segment.Tags = CompileMainTags(system.Operation.Content)
	if !segment.hasTemplateCase(&system, "Default") || segment.hasTemplateCase(&system, "Changed") {
		t.Fatal("explicit Tags no longer control selected-case validation")
	}
}

func TestSelectedSegmentCasesConcurrentOwnership(t *testing.T) {
	system, err := loadSegmentalSystem(runtimeGenerationResources(t))
	if err != nil {
		t.Fatal(err)
	}
	segment := system.Operation
	want := ParseSegmentTags(CompileMainTags(segment.Content), MetaDict{}, "Default", false)
	var workers sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for iteration := 0; iteration < 10; iteration++ {
				got := parseSegmentTags(segment.selectedTemplateTags(&system, "Default", false), MetaDict{}, "Default", false, true)
				if !reflect.DeepEqual(got, want) {
					t.Error("concurrent selection changed parsed resource")
					return
				}
				got[0].Expressions[2].ExpressionString = "private glyph mutation"
			}
		}()
	}
	workers.Wait()
}

func TestSegmentTagCopyPreservesMetadataIndex(t *testing.T) {
	system, err := loadSegmentalSystem(runtimeGenerationResources(t))
	if err != nil {
		t.Fatal(err)
	}
	original := CompileMainTags(system.Operation.Content)[0]
	original.TagMeta.AppMeta.Add("Key", "first", "String")
	original.TagMeta.AppMeta.Add("Key", "second", "SourceString")
	copied := cloneSegmentTag(original)
	if !reflect.DeepEqual(copied, original) {
		t.Fatal("copy normalized duplicate metadata records or their indexes")
	}
	copied.TagMeta.AppMeta.Index.Keys[0] = "Changed"
	copied.TagMeta.AppMeta.Index.Indexes[0] = 9
	copied.TagMeta.AppMeta.Metas[0].Value = "changed value"
	if original.TagMeta.AppMeta.Index.Keys[0] != "Key" || original.TagMeta.AppMeta.Index.Indexes[0] != 0 || original.TagMeta.AppMeta.Metas[0].Value != "first" {
		t.Fatal("copied metadata index or records still alias original data")
	}
}
