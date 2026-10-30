package Logiceme

import (
	"strings"
	"testing"
)

func TestStructureSpanRetainsLegacyFirstMatch(t *testing.T) {
	base := string(compositionResources(t, "7")["builtins/runtime/support.aml"].Data)
	for _, test := range []struct {
		name, input, message string
	}{
		{"ordinary", base, ""},
		{"closing whitespace", strings.Replace(base, "</STRUCTURE>", "</STRUCTURE >", 1), "Rust program requires one STRUCTURE tag"},
		{"nested before book", strings.Replace(base, "<STRUCTURE>", "<STRUCTURE><STRUCTURE></STRUCTURE>", 1), "Rust program requires complete application BOOK tags"},
		{"nested around book", strings.Replace(strings.Replace(base, "<STRUCTURE>", "<STRUCTURE><STRUCTURE>", 1), "</STRUCTURE>", "</STRUCTURE></STRUCTURE>", 1), ""},
		{"JSON outside structure", strings.Replace(base, "<PROGRAM>", `<PROGRAM><JSON>{"text":"<STRUCTURE>fake</STRUCTURE>"}</JSON>`, 1), "Rust program requires one STRUCTURE tag"},
	} {
		t.Run(test.name, func(t *testing.T) {
			scanner, err := loadScannerBytes([]byte(test.input))
			if err == nil {
				err = scanSourceProgram(&scanner)
			}
			if test.message != "" {
				if err == nil || err.Error() != test.message {
					t.Fatalf("scan error = %v; want %q", err, test.message)
				}
			} else if err != nil || len(scanner.Program.Books) != 1 {
				t.Fatalf("scan returned %d books, %v; want one complete book", len(scanner.Program.Books), err)
			}
		})
	}
}
