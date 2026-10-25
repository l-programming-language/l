package Lyger

import (
	"sort"
	"strings"
)

func displayReplacer(cache map[string][]byte) *strings.Replacer {
	keys := make([]string, 0, len(cache))
	for key := range cache {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	replacements := make([]string, 0, len(keys)*2)
	for _, key := range keys {
		value := string(cache[key])
		if frame, framed := splitDefinitionFrame(value); framed && frame.Quoted {
			value = "`" + decodeDefinitionEscapes(frame.Body) + "`"
		}
		replacements = append(replacements, key, value)
	}
	return strings.NewReplacer(replacements...)
}

// renderExpandedSource emits the source view from ordered originating
// statements. It removes internal stamps/cache identities and definition
// markers, without altering definition targets, parameters, or source order.
func renderExpandedSource(statements []SourceStatement, cache map[string][]byte) []byte {
	lines := make([]string, 0, len(statements))
	replacer := displayReplacer(cache)
	for _, statement := range statements {
		body := statementBody(statement.Prepared)
		lines = append(lines, replacer.Replace(string(body))+".")
	}
	return []byte(strings.Join(lines, "\n"))
}
