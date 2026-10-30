package Lyger

import (
	"bytes"
	"testing"
)

func TestImportProjectionKeepsValidatedWords(t *testing.T) {
	gram := scopeTestGrammar(t)
	raw := []byte("  uSe Tools’s Render from `dir/file.l` as \"Pretty \\\"Print\\\"\". *** note ***\r\n")
	original := bytes.Clone(raw)
	matched, imported, words, err := gram.scanImportStatement(raw)
	if err != nil || !matched || len(words) != 7 {
		t.Fatalf("scan import = %t, %q, %v", matched, words, err)
	}
	before := bytes.Join(words, []byte(" "))
	source := SourceStatement{Prepared: raw, Raw: raw, Span: SourceSpan{Path: "main.l", Start: SourcePosition{Line: 3, Column: 1}}}
	alias, err := imported.aliasStatement(source, words)
	if err != nil {
		t.Fatal(err)
	}
	want := "uSe Tools’s Render as \"Pretty \\\"Print\\\"\""
	if got := string(statementBody(alias.Prepared)); got != want {
		t.Fatalf("alias = %q; want %q", got, want)
	}
	if !bytes.Equal(raw, original) || !bytes.Equal(bytes.Join(words, []byte(" ")), before) || !bytes.Equal(imported.statementBytes, original) || alias.Span != source.Span || !bytes.Equal(alias.Raw, original) {
		t.Fatal("projection changed source bytes, validated words or provenance")
	}
}
