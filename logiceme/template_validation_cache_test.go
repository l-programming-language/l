package Logiceme

import (
	"errors"
	"fmt"
	"io/fs"
	"reflect"
	"strings"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/l-programming-language/l/lerrors"
)

func TestLucidTemplateValidationCacheRereadsResources(t *testing.T) {
	const asset = "builtins/segments/consonants/probe.lu"
	content := lucidFramingTemplate("selected resource cache probe")
	resources := fstest.MapFS{asset: &fstest.MapFile{Data: content}}
	load := func() error {
		segment := Segment{Name: "probe", SubType: "consonant"}
		return segment.loadTemplate(resources)
	}
	if err := load(); err != nil {
		t.Fatal(err)
	}
	resources[asset].Data = []byte("damaged selected resource")
	if err := load(); !errors.Is(err, lerrors.InvalidConfiguration) || !strings.Contains(err.Error(), asset) {
		t.Fatalf("changed selected resource = %v; want contextual validation error", err)
	}
	resources[asset].Data = content
	if err := load(); err != nil {
		t.Fatalf("restored selected resource: %v", err)
	}
	delete(resources, asset)
	if err := load(); !errors.Is(err, fs.ErrNotExist) || !strings.Contains(err.Error(), asset) {
		t.Fatalf("missing selected resource = %v; want original filesystem cause and path", err)
	}
}

func TestLucidTemplateValidationCachePreservesInputs(t *testing.T) {
	content := lucidFramingTemplate("cache ownership probe")
	original := string(content)
	if err := validateLucidTemplate(content, true); err != nil {
		t.Fatal(err)
	}
	// Mutating the caller's original buffer must not reuse its old validation.
	content[len(content)-1] = '!'
	if err := validateLucidTemplate(content, true); !errors.Is(err, lerrors.InvalidConfiguration) || !strings.Contains(err.Error(), "no Lucid tags") {
		t.Fatalf("changed content validation = %v; want original missing-tags error", err)
	}
	if err := validateLucidTemplate([]byte(original), true); err != nil {
		t.Fatalf("original content was affected by caller mutation: %v", err)
	}

	literal := []byte("tag-free cache ownership probe")
	if err := validateLucidTemplate(literal, false); err != nil {
		t.Fatal(err)
	}
	if err := validateLucidTemplate(literal, true); !errors.Is(err, lerrors.InvalidConfiguration) || !strings.Contains(err.Error(), "no Lucid tags") {
		t.Fatalf("required tags validation = %v; want missing-tags error", err)
	}
}

func TestLucidTemplateValidationCacheConcurrent(t *testing.T) {
	const workers = 8
	var group sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		group.Add(1)
		go func() {
			defer group.Done()
			for index := 0; index < 20; index++ {
				content := lucidFramingTemplate(fmt.Sprintf("concurrent cache probe %d", index))
				if err := validateLucidTemplate(content, true); err != nil {
					t.Error(err)
				}
				content[len(content)-1] = '!'
				if err := validateLucidTemplate(content, true); !errors.Is(err, lerrors.InvalidConfiguration) {
					t.Errorf("changed concurrent content validation = %v; want invalid configuration", err)
				}
			}
		}()
	}
	group.Wait()
}

func TestLucidTemplateValidationCacheBounded(t *testing.T) {
	var cache lucidTemplateValidationCache
	oversized := lucidTemplateValidationKey{source: strings.Repeat("x", maxValidatedLucidBytes+1)}
	cache.remember(oversized)
	if cache.contains(oversized) {
		t.Fatal("cache retained an oversized template")
	}
	for index := 0; index <= maxValidatedLucidTemplates; index++ {
		cache.remember(lucidTemplateValidationKey{source: fmt.Sprintf("template %d", index)})
	}
	if len(cache.valid) != maxValidatedLucidTemplates {
		t.Fatalf("retained templates = %d; want bounded capacity %d", len(cache.valid), maxValidatedLucidTemplates)
	}
	if !cache.contains(lucidTemplateValidationKey{source: "template 0"}) {
		t.Fatal("cache forgot a previously retained template")
	}
	if cache.contains(lucidTemplateValidationKey{source: "template 0", requireTags: true}) {
		t.Fatal("cache reused validation with a different tag requirement")
	}
}

func TestLucidParserConcurrentResultsAreIndependent(t *testing.T) {
	const tag = "{|/|-{Meta}--{Condition}--{é\r\n\xffbody}-|/|}"
	cases := []struct {
		selector, source string
		want             []string
	}{
		{"MainTag", "prefix " + tag + " suffix", []string{tag}},
		{"unknown selector", tag, []string{tag}},
		{"Expression", "-{a}- -{b}-", []string{"-{a}-", "-{b}-"}},
		{"SubExpression", "={meta}=={condition}=={body}=", []string{"={meta}=", "={condition}=", "={body}="}},
		{"MainTag", "no complete tags {|/|", nil},
	}
	for worker := 0; worker < 8; worker++ {
		t.Run(fmt.Sprintf("worker %d", worker), func(t *testing.T) {
			t.Parallel()
			for iteration := 0; iteration < 20; iteration++ {
				for _, test := range cases {
					got := lucidParser(test.source, test.selector)
					if !reflect.DeepEqual(got, test.want) {
						t.Fatalf("%s parsed = %q; want unchanged source slices %q", test.selector, got, test.want)
					}
					if len(got) != 0 {
						got[0] = "caller changed returned slice"
					}
					if next := lucidParser(test.source, test.selector); !reflect.DeepEqual(next, test.want) {
						t.Fatalf("%s reused caller-owned results: %q", test.selector, next)
					}
				}
			}
		})
	}
}
