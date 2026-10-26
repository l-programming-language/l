package Lyger

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestImportedGrammarErrorKeepsSourceLocation(t *testing.T) {
	for _, tc := range []struct{ name, source, want string }{
		{"missing operand", "\"value\" is a.", "at end of statement"},
		{"wrong operand", "\"value\" is a with.", `remaining token "with"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			directory := t.TempDir()
			for name, source := range map[string]string{
				"broken.l": "*** introduction ***\n\n  " + tc.source + "\n",
			} {
				if err := os.WriteFile(filepath.Join(directory, name), []byte(source), 0600); err != nil {
					t.Fatal(err)
				}
			}
			service := NewLygerService(canonicalTestOptions())
			result, err := service.Compile([]byte("Use `broken.l`.\n"), filepath.Join(directory, "main.l"))
			if result != nil || err == nil {
				t.Fatalf("invalid imported syntax returned result=%v, err=%v", result, err)
			}
			for _, want := range []string{filepath.Join(directory, "broken.l") + ":3:3", "grammar", "statement.subjective.passive", tc.want} {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("grammar error lost %q: %v", want, err)
				}
			}
		})
	}
}

func TestStatementDiagnosticRestoresLiteralsAndRetainsCause(t *testing.T) {
	prepared, err := PrepareSource("named source.l", []byte("\n\"value\" is `quoted 雪`.\n"))
	if err != nil {
		t.Fatal(err)
	}
	c := &Compilation{MainPath: prepared.Path, TextCache: prepared.TextCache}
	statement := &lStatement{source: &prepared.Statements[0]}
	cause := errors.New("invalid operand")
	key := prepared.Literals[0].ID
	wrapped := fmt.Errorf("operand %s: %w", key, cause)
	if !c.checkStatementError(statement, "grammar", wrapped) {
		t.Fatal("input failure was not recorded")
	}
	if !errors.Is(c.Error, cause) || !strings.Contains(c.Error.Error(), "named source.l:2:1") || !strings.Contains(c.Error.Error(), "`quoted 雪`") || strings.Contains(c.Error.Error(), key) {
		t.Fatalf("diagnostic lost source, literal, or cause: %v", c.Error)
	}
	if string(c.TextCache[key]) != "`quoted 雪`" || !strings.Contains(string(prepared.PreparedBytes), key) {
		t.Fatal("diagnostic rendering mutated compiler input")
	}
	first := c.Error
	c.checkStatementError(statement, "later phase", errors.New("later failure"))
	if c.Error != first {
		t.Fatal("later diagnostic replaced the first failure")
	}
}
