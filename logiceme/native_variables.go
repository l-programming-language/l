package Logiceme

import (
	"strings"
	"sync"
)

const (
	maxNativeVariableSources = 128
	maxNativeVariableBytes   = 8192
)

// Only original template occurrence positions are shared. Their interpretation
// still uses each glyph's metadata, children and lexical loop bindings. Keeping
// the original parser also preserves its multiline and malformed-frame behavior.
type nativeVariableCache struct {
	mu      sync.Mutex
	entries map[string][]lucidSourceSpan
	order   []string
	next    int
}

var nativeVariables nativeVariableCache

// Returned spans are borrowed immutable syntax, never rendered values.
func (cache *nativeVariableCache) load(source string) []lucidSourceSpan {
	if len(source) > maxNativeVariableBytes {
		return parseNativeVariables(source)
	}
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if spans, found := cache.entries[source]; found {
		return spans
	}
	spans := parseNativeVariables(source)
	if cache.entries == nil {
		cache.entries = make(map[string][]lucidSourceSpan)
	}
	source = strings.Clone(source)
	if len(cache.order) == maxNativeVariableSources {
		delete(cache.entries, cache.order[cache.next])
		cache.order[cache.next] = source
		cache.next = (cache.next + 1) % maxNativeVariableSources
	} else {
		cache.order = append(cache.order, source)
	}
	cache.entries[source] = spans
	return spans
}

func parseNativeVariables(source string) []lucidSourceSpan {
	var spans []lucidSourceSpan
	cursor := 0
	for _, variable := range lucidParser(source, "ExpressionVar") {
		start := cursor + strings.Index(source[cursor:], variable)
		cursor = start + len(variable)
		spans = append(spans, lucidSourceSpan{start, cursor})
	}
	return spans
}
