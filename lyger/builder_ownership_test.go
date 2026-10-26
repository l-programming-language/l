package Lyger

import (
	"bytes"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestBuilderCacheAdoptionPreservesPreparedOwnership(t *testing.T) {
	root := t.TempDir()
	writeSourceFile(t, root, "definitions.l", "\"local\" is `definition literal`.\n")
	writeSourceFile(t, root, "leaf.l", "\"leaf\" is `imported literal`.\n")
	main := writeSourceFile(t, root, "main.l", "Use `definitions.l`.\n\"main\" is `main literal`.\nUse `leaf.l`.\n")
	raw, err := os.ReadFile(main)
	if err != nil {
		t.Fatal(err)
	}
	prepared, err := PrepareSource(main, raw)
	if err != nil {
		t.Fatal(err)
	}
	wantPool := make(map[string][]byte, len(prepared.TextCache))
	for id, literal := range prepared.TextCache {
		wantPool[id] = bytes.Clone(literal)
	}
	wantOriginal, wantParserView := bytes.Clone(prepared.Original), bytes.Clone(prepared.PreparedBytes)
	gram := scopeTestGrammar(t)
	if gram == nil {
		t.Fatal("fixture grammar failed to load")
	}
	compilation := &Compilation{MainPath: main, Prepared: prepared, Gram: gram, CompiledBytes: prepared.PreparedBytes, TextCache: prepared.TextCache}
	compilation.build()
	if compilation.Error != nil {
		t.Fatal(compilation.Error)
	}
	for _, wanted := range []string{"`definition literal`", "`imported literal`"} {
		count := 0
		for _, literal := range compilation.TextCache {
			if string(literal) == wanted {
				count++
			}
		}
		if count != 1 {
			t.Errorf("imported literal %s occurs %d times; want once", wanted, count)
		}
	}
	if !reflect.DeepEqual(prepared.TextCache, wantPool) {
		t.Error("successful expansion changed the prepared source's local literal pool")
	}
	for id := range wantPool {
		compilation.TextCache[id][1] = 'X'
	}
	if !reflect.DeepEqual(prepared.TextCache, wantPool) || !bytes.Equal(prepared.Original, wantOriginal) || !bytes.Equal(prepared.PreparedBytes, wantParserView) {
		t.Error("published expansion shares mutable buffers with prepared source")
	}
}

func TestBuilderFailureDoesNotPublishPartialExpansion(t *testing.T) {
	root := t.TempDir()
	writeSourceFile(t, root, "definitions.l", "\"local\" is `definition literal`.\n")
	writeSourceFile(t, root, "leaf.l", "\"leaf\" is `imported literal`.\n")
	main := writeSourceFile(t, root, "main.l", "Use `definitions.l`.\nUse `leaf.l`.\nUse `missing.l`.\n")
	raw, err := os.ReadFile(main)
	if err != nil {
		t.Fatal(err)
	}
	prepared, err := PrepareSource(main, raw)
	if err != nil {
		t.Fatal(err)
	}
	wantBytes := []byte("previous expanded bytes")
	wantPool := map[string][]byte{"previous": []byte("previous literal")}
	wantStatements := []SourceStatement{{Raw: []byte("previous raw"), Span: SourceSpan{Path: "previous.l"}}}
	wantPaths := []string{"previous.l"}
	gram := scopeTestGrammar(t)
	if gram == nil {
		t.Fatal("fixture grammar failed to load")
	}
	compilation := &Compilation{MainPath: main, Prepared: prepared, Gram: gram, CompiledBytes: bytes.Clone(wantBytes), TextCache: map[string][]byte{"previous": bytes.Clone(wantPool["previous"])}, SourceStatements: wantStatements, SourcePaths: wantPaths}
	compilation.build()
	if compilation.Error == nil || !strings.Contains(compilation.Error.Error(), "missing.l") {
		t.Fatalf("missing nested import error = %v", compilation.Error)
	}
	if !bytes.Equal(compilation.CompiledBytes, wantBytes) || !reflect.DeepEqual(compilation.TextCache, wantPool) || !reflect.DeepEqual(compilation.SourceStatements, wantStatements) || !reflect.DeepEqual(compilation.SourcePaths, wantPaths) {
		t.Fatal("failed expansion published partial bytes, literals, statements, or paths")
	}
}
