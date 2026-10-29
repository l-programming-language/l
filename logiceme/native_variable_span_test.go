package Logiceme

import (
	"fmt"
	"math/rand"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"testing"
)

func TestNativeVariableSpansMatchOriginalRegex(t *testing.T) {
	original := regexp.MustCompile(`\[\|((?:.*?\r?\n?)*)\|]`)
	var cache nativeVariableCache
	inputs := []string{
		"", "[|", "|]", "[||]", "[|[||]", "[|x|][|y|]", "[|unclosed [|x|] trailing",
		"before [|α\r\nβ\rx\ny\xff|] after", "[|Meta.Value|] [|Meta.Value|]",
		"[|[|Meta.Value|]|]", "[|first|] [|unfinished", "[|||]",
	}
	random := rand.New(rand.NewSource(1)) //nolint:gosec // Deterministic parser differential corpus, not security randomness.
	alphabet := []byte("[|]x\r\n\xff")
	for sample := 0; sample < 500; sample++ {
		input := make([]byte, random.Intn(96))
		for index := range input {
			input[index] = alphabet[random.Intn(len(alphabet))]
		}
		inputs = append(inputs, string(input))
	}
	for _, input := range inputs {
		var spans [][]int
		for _, span := range cache.load(input) {
			spans = append(spans, []int{span.start, span.end})
		}
		if want := original.FindAllStringIndex(input, -1); !reflect.DeepEqual(spans, want) {
			t.Fatalf("%q: spans %v, want %v", input, spans, want)
		}
	}
}

func TestNativeVariableCacheBoundedAndConcurrent(t *testing.T) {
	var cache nativeVariableCache
	for index := 0; index <= maxNativeVariableSources; index++ {
		cache.load(fmt.Sprintf("[|Meta.Value%d|]", index))
	}
	if len(cache.entries) != maxNativeVariableSources {
		t.Fatalf("cache retained %d sources", len(cache.entries))
	}
	if _, found := cache.entries["[|Meta.Value0|]"]; found {
		t.Fatal("oldest source was not evicted")
	}
	oversized := strings.Repeat(" ", maxNativeVariableBytes) + "[|Meta.Value|]"
	cache.load(oversized)
	if _, found := cache.entries[oversized]; found {
		t.Fatal("oversized source was retained")
	}
	var workers sync.WaitGroup
	for index := 0; index < 8; index++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for repeat := 0; repeat < 20; repeat++ {
				if got := cache.load("[|Meta.Value|]"); len(got) != 1 || got[0] != (lucidSourceSpan{0, 14}) {
					t.Errorf("shared syntax changed: %v", got)
				}
			}
		}()
	}
	workers.Wait()
}

func TestNativeVariableExpansionDoesNotRescanValues(t *testing.T) {
	var meta MetaDict
	meta.Add("Value", "[|Meta.Other|]", "SourceString")
	meta.Add("Other", "replaced", "SourceString")
	interpreter := LucidInterpreter{Meta: meta, SegmentalSystem: &SegmentalSystem{Language: "Rust"}}
	if got := interpreter.interpretNativeLiteral("[|Meta.Value|] [|Meta.Other|]"); got != "[|Meta.Other|] replaced" {
		t.Fatalf("native value was interpreted again: %q", got)
	}
}
