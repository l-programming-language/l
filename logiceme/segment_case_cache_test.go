package Logiceme

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/l-programming-language/l/lerrors"
)

func TestParsedSegmentCacheRereadsSelectedResources(t *testing.T) {
	const asset = "builtins/segments/vowels/operation.lu"
	content := lucidFramingTemplate("first selected body")
	files := fstest.MapFS{asset: &fstest.MapFile{Data: content}}
	load := func() (SegmentalSystem, error) {
		system := SegmentalSystem{Language: "Rust"}
		err := system.AddLogiceme(Segment{Symbol: "W", Name: "operation", SubType: "vowel"}, true, files)
		return system, err
	}
	first, err := load()
	if err != nil {
		t.Fatal(err)
	}
	second, err := load()
	if err != nil || first.templateCases["W"] != second.templateCases["W"] {
		t.Fatalf("identical bytes did not reuse immutable parsed cases: %v", err)
	}
	owned := first.Operation.selectedTemplateTags(&first, "Default", false)
	owned[0].Expressions[2].ExpressionString = "caller mutation"
	owned[0].TagMeta.AppMeta.Add("Owned", "changed", "String")
	if got := second.Operation.selectedTemplateTags(&second, "Default", false); got[0].Expressions[2].ExpressionString != "first selected body" || got[0].TagMeta.AppMeta.GetVal("Owned") != "NotFound" {
		t.Fatal("another evaluation received mutated selected tags")
	}
	// Mutating the original provider buffer must neither change the cached
	// immutable tree nor reuse it for different selected bytes.
	copy(content, bytes.Replace(content, []byte("first"), []byte("other"), 1))
	changed, err := load()
	if err != nil || changed.Operation.selectedTemplateTags(&changed, "Default", false)[0].Expressions[2].ExpressionString != "other selected body" {
		t.Fatalf("provider mutation did not control parsing: %v", err)
	}
	if first.Operation.selectedTemplateTags(&first, "Default", false)[0].Expressions[2].ExpressionString != "first selected body" {
		t.Fatal("provider mutation changed an earlier evaluation")
	}
	files[asset].Data = []byte("damaged selected template")
	if _, err := load(); !errors.Is(err, lerrors.InvalidConfiguration) || !strings.Contains(err.Error(), asset) {
		t.Fatalf("malformed selected resource reused cache: %v", err)
	}
	delete(files, asset)
	if _, err := load(); !errors.Is(err, fs.ErrNotExist) || !strings.Contains(err.Error(), asset) {
		t.Fatalf("missing selected resource reused cache: %v", err)
	}
	files[asset] = &fstest.MapFile{Data: lucidFramingTemplate("first selected body")}
	if _, err := load(); err != nil {
		t.Fatalf("restored resource retained an error: %v", err)
	}
}

func TestParsedSegmentCacheBounded(t *testing.T) {
	var cache segmentCaseCache
	first := lucidFramingTemplate("first entry")
	cache.load(first)
	for i := 0; i < maxParsedSegments; i++ {
		cache.load(lucidFramingTemplate(fmt.Sprintf("entry %d", i)))
	}
	if len(cache.entries) != maxParsedSegments || len(cache.order) != maxParsedSegments || cache.entries[string(first)] != nil {
		t.Fatal("parsed case cache did not evict its oldest entry at capacity")
	}
	oversized := lucidFramingTemplate(strings.Repeat("x", maxParsedSegmentBytes))
	parsed := cache.load(oversized)
	if len(parsed.cases) != 1 || cache.entries[string(oversized)] != nil || len(cache.entries) != maxParsedSegments {
		t.Fatal("oversized template was skipped or retained in the cache")
	}
}

func TestParsedSegmentCacheConcurrentLoads(t *testing.T) {
	var cache segmentCaseCache
	content := lucidFramingTemplate("concurrent body")
	var workers sync.WaitGroup
	for i := 0; i < 8; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for j := 0; j < 10; j++ {
				parsed := cache.load(content)
				system := SegmentalSystem{templateCases: map[string]*segmentCases{"W": parsed}}
				segment := Segment{Symbol: "W", Content: bytes.Clone(content)}
				owned := segment.selectedTemplateTags(&system, "Default", false)
				if len(owned) != 1 || owned[0].Expressions[2].ExpressionString != "concurrent body" {
					t.Error("concurrent load received another caller's mutations")
					return
				}
				owned[0].Expressions[2].ExpressionString = "mutated"
			}
		}()
	}
	workers.Wait()
	if len(cache.entries) != 1 {
		t.Fatal("concurrent loads did not share one immutable entry")
	}
}
