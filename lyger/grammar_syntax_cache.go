package Lyger

import (
	"strings"
	"sync"
)

const grammarSyntaxEntries = 256
const grammarSyntaxBytes = 4096

// Only immutable syntax is shared: keys contain the complete selected text,
// and values contain no grammar configuration, bindings, tokens or parse state.
// FIFO eviction and a per-key size limit bound retention for custom grammars.
type grammarSyntaxCache[T any] struct {
	mu      sync.RWMutex
	entries map[string]T
	keys    [grammarSyntaxEntries]string
	next    int
}

func (c *grammarSyntaxCache[T]) lookup(raw string) (T, bool) {
	c.mu.RLock()
	value, exists := c.entries[raw]
	c.mu.RUnlock()
	return value, exists
}

func (c *grammarSyntaxCache[T]) store(raw string, value T) {
	if len(raw) > grammarSyntaxBytes {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.entries[raw]; exists {
		return
	}
	if c.entries == nil {
		c.entries = make(map[string]T)
	}
	if len(c.entries) == grammarSyntaxEntries {
		delete(c.entries, c.keys[c.next])
	}
	c.entries[raw] = value
	c.keys[c.next] = raw
	c.next = (c.next + 1) % grammarSyntaxEntries
}

var grammarTerms grammarSyntaxCache[grammarTerm]
var grammarRules grammarSyntaxCache[[]string]

// Rule parts are immutable and borrowed only by the ordered validation and
// matching loops. Each decoded term is returned by value at its original point
// of use, preserving the first failing element and its diagnostic cause.
func splitGramRule(rule string) []string {
	if parts, exists := grammarRules.lookup(rule); exists {
		return parts
	}
	if len(rule) <= grammarSyntaxBytes {
		rule = strings.Clone(rule)
	}
	parts := strings.Split(rule, " | ")
	grammarRules.store(rule, parts)
	return parts
}

func decodeGrammarTerm(raw string) (grammarTerm, error) {
	if term, exists := grammarTerms.lookup(raw); exists {
		return term, nil
	}
	if len(raw) <= grammarSyntaxBytes {
		raw = strings.Clone(raw)
	}
	term, err := parseGrammarTerm(raw)
	if err == nil {
		grammarTerms.store(raw, term)
	}
	return term, err
}
