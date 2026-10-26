package Lyger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"sort"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

func readDictionaryIndex(resources fs.FS) (*dictionaryIndex, error) {
	const name = "stdlib/index.json"
	data, err := fs.ReadFile(resources, name)
	if err != nil {
		return nil, lerrors.IO.Errorf("read dictionary index %s: %w", name, err)
	}
	// JSON maps would silently discard repeated keys, including escaped spellings
	// of the same dictionary name. Reject ambiguity before decoding the index.
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := uniqueIndexKeys(decoder); err != nil {
		return nil, lerrors.InvalidConfiguration.Errorf("decode dictionary index %s: %w", name, err)
	}
	if _, err := decoder.Token(); err != io.EOF {
		if err == nil {
			err = fmt.Errorf("multiple JSON values")
		}
		return nil, lerrors.InvalidConfiguration.Errorf("decode dictionary index %s: trailing content: %w", name, err)
	}
	var index dictionaryIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, lerrors.InvalidConfiguration.Errorf("decode dictionary index %s: %w", name, err)
	}
	if index.Format != 1 {
		return nil, lerrors.InvalidConfiguration.Errorf("dictionary index %s has unsupported format %d", name, index.Format)
	}
	validPath := func(value string) bool {
		return fs.ValidPath(value) && value != "." && !strings.Contains(value, "\\")
	}
	// Sorting makes malformed bundle diagnostics independent of map iteration.
	languages := make([]string, 0, len(index.Languages))
	for language := range index.Languages {
		languages = append(languages, language)
	}
	sort.Strings(languages)
	entries := make(map[string]string)
	for _, language := range languages {
		if !validPath(language) {
			return nil, lerrors.InvalidConfiguration.Errorf("dictionary index %s has invalid language %q", name, language)
		}
		dictionaries := index.Languages[language]
		keys := make([]string, 0, len(dictionaries))
		for key := range dictionaries {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			entry := dictionaries[key]
			if !validPath(key) || !validPath(entry.Entry) || !strings.HasSuffix(entry.Entry, ".l") || strings.TrimSpace(entry.Dictionary) == "" {
				return nil, lerrors.InvalidConfiguration.Errorf("dictionary index %s has invalid entry %s/%s", name, language, key)
			}
			if prior := entries[entry.Entry]; prior != "" {
				if prior != entry.Dictionary {
					return nil, lerrors.InvalidConfiguration.Errorf("dictionary index %s entry %s has conflicting identities", name, entry.Entry)
				}
				continue
			}
			info, err := fs.Stat(resources, entry.Entry)
			if err != nil {
				return nil, lerrors.IO.Errorf("dictionary index %s entry %s: %w", name, entry.Entry, err)
			}
			if !info.Mode().IsRegular() {
				return nil, lerrors.InvalidConfiguration.Errorf("dictionary index %s entry %s must be a source file", name, entry.Entry)
			}
			entries[entry.Entry] = entry.Dictionary
		}
	}
	return &index, nil
}

// uniqueIndexKeys consumes exactly one JSON value, rejecting duplicate keys in
// every object without imposing a different JSON syntax or escaping convention.
func uniqueIndexKeys(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, composite := token.(json.Delim)
	if !composite {
		return nil
	}
	keys := make(map[string]bool)
	for decoder.More() {
		if delimiter == '{' {
			key, err := decoder.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			if !ok {
				return fmt.Errorf("object key must be a string")
			}
			if keys[name] {
				return fmt.Errorf("duplicate object key %q", name)
			}
			keys[name] = true
		}
		if err := uniqueIndexKeys(decoder); err != nil {
			return err
		}
	}
	_, err = decoder.Token()
	return err
}
