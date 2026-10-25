package Lyger

import (
	"fmt"
	"sort"
	"strings"
)

// statementDiagnostic keeps source provenance and the original error together.
// Literal cache IDs are an internal representation, so only the diagnostic
// message restores them; parser state and observable output bytes stay intact.
type statementDiagnostic struct {
	message string
	cause   error
}

func (d *statementDiagnostic) Error() string { return d.message }
func (d *statementDiagnostic) Unwrap() error { return d.cause }

func (s SourceSpan) wrapError(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s:%d:%d: %w", s.Path, s.Start.Line, s.Start.Column, err)
}

func (c *Compilation) checkStatementError(statement *lStatement, phase string, err error) bool {
	if err == nil {
		return false
	}
	if c.Error != nil {
		return true
	}
	message := err.Error()
	keys := make([]string, 0, len(c.TextCache))
	for key := range c.TextCache {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var replacements []string
	for _, key := range keys {
		replacements = append(replacements, key, string(c.TextCache[key]))
	}
	if len(replacements) > 0 {
		message = strings.NewReplacer(replacements...).Replace(message)
	}
	location := c.MainPath
	if statement != nil {
		if statement.source != nil {
			span := statement.source.Span
			location = fmt.Sprintf("%s:%d:%d", span.Path, span.Start.Line, span.Start.Column)
		} else if len(statement.headBytes) > 0 {
			location = string(statement.headBytes)
		}
	}
	c.Error = &statementDiagnostic{message: fmt.Sprintf("%s: %s: %s", location, phase, message), cause: err}
	return true
}
