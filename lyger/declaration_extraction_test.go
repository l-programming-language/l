package Lyger

import "testing"

func TestDeclarationExtractionSourceTokens(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		labels, ids  []string
	}{
		{"passive Text", "\"Payload\" is `unmodified text`", []string{`"Payload"`}, []string{"Payload"}},
		{"imperative App", `Create "Worker" as an App`, []string{`"Worker"`}, []string{"Worker"}},
		{"morphological action", `Create "Stage|-s|-d|-ing" for Building any App and then Running it`, []string{`"Stage|-s|-d|-ing"`}, []string{"Stage"}},
		{"conditional declaration", `If it has an Error, exit with the Error, else define "response" as it's Output`, []string{`"response"`}, []string{"response"}},
		{"ordered branches", "If true is true, define \"first\" as `left`, else define \"second\" as `right`", []string{`"first"`, `"second"`}, []string{"first", "second"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stmt := extractionTestStatement(t, tc.source)
			var declarations []*token
			for _, tok := range stmt.tokens {
				if tok.isDeclare() {
					declarations = append(declarations, tok)
				}
			}
			if len(declarations) != len(tc.labels) {
				t.Fatalf("declaration tokens = %d, want %d", len(declarations), len(tc.labels))
			}
			for index, tok := range declarations {
				if string(tok.rawBytes) != tc.labels[index] {
					t.Fatalf("declaration token %d = %q, want %q", index, tok.rawBytes, tc.labels[index])
				}
			}
			records, err := stmt.sGram.getDeclarations()
			if err != nil {
				t.Fatal(err)
			}
			if len(records) != len(declarations) {
				t.Fatalf("records = %d, want %d", len(records), len(declarations))
			}
			for index, extracted := range records {
				if extracted.Token != declarations[index] {
					t.Fatalf("record %d lost its original declaration token", index)
				}
				record := extracted.Record
				if record == nil {
					t.Fatalf("declaration %d has no record", index)
				}
				if record.ID != tc.ids[index] {
					t.Fatalf("record %d ID = %q, want %q", index, record.ID, tc.ids[index])
				}
			}
		})
	}
}
