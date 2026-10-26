package Logiceme

import (
	"fmt"
	"strings"
	"sync"

	"github.com/l-programming-language/l/lerrors"
)

const (
	maxValidatedLucidTemplates = 256
	maxValidatedLucidBytes     = 64 * 1024
)

type lucidTemplateValidationKey struct {
	source      string
	requireTags bool
}

// Cache only successful, content-only validation. Resource reads and mutable
// parsed/rendered results remain independent for every selected filesystem.
// Both limits bound retained caller content; oversized or additional templates
// still receive exactly the same validation without entering the cache.
type lucidTemplateValidationCache struct {
	mu    sync.RWMutex
	valid map[lucidTemplateValidationKey]struct{}
}

func (cache *lucidTemplateValidationCache) contains(key lucidTemplateValidationKey) bool {
	cache.mu.RLock()
	_, found := cache.valid[key]
	cache.mu.RUnlock()
	return found
}

func (cache *lucidTemplateValidationCache) remember(key lucidTemplateValidationKey) {
	if len(key.source) > maxValidatedLucidBytes {
		return
	}
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if len(cache.valid) >= maxValidatedLucidTemplates {
		return
	}
	if cache.valid == nil {
		cache.valid = make(map[lucidTemplateValidationKey]struct{})
	}
	cache.valid[key] = struct{}{}
}

var validatedLucidTemplates lucidTemplateValidationCache

// Validate the framing that the Lucid parser indexes directly. This
// checks template structure, without changing source text or interpreting it.
func validateLucidTemplate(content []byte, requireTags bool) error {
	source := string(content)
	key := lucidTemplateValidationKey{source: source, requireTags: requireTags}
	if validatedLucidTemplates.contains(key) {
		return nil
	}
	tags := lucidParser(source, "MainTag")
	if requireTags && len(tags) == 0 {
		return lerrors.InvalidConfiguration.Errorf("template contains no Lucid tags")
	}
	if strings.Count(source, "{|/|") != len(tags) || strings.Count(source, "|/|}") != len(tags) {
		return lerrors.InvalidConfiguration.Errorf("template has an incomplete Lucid tag")
	}
	for i, tag := range tags {
		expressions := lucidParser(tag, "Expression")
		if len(expressions) != 3 {
			return lerrors.InvalidConfiguration.Errorf("template tag %d requires three expressions, got %d", i+1, len(expressions))
		}
		if expressions[1] == "-{Inputs}-" {
			inputs := strings.TrimSuffix(strings.TrimPrefix(expressions[2], "-{["), "]}-")
			if !strings.HasPrefix(expressions[2], "-{[") || !strings.HasSuffix(expressions[2], "]}-") {
				return lerrors.InvalidConfiguration.Errorf("template tag %d has malformed Inputs", i+1)
			}
			for _, input := range strings.Split(inputs, ", ") {
				if len(strings.Split(input, ": ")) != 2 {
					return lerrors.InvalidConfiguration.Errorf("template tag %d has malformed input %q", i+1, input)
				}
			}
		}
		if _, err := recursiveSubTagParser(expressions[2]); err != nil {
			return fmt.Errorf("template tag %d: %w", i+1, err)
		}
	}
	validatedLucidTemplates.remember(key)
	return nil
}
