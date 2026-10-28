package Lyger

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestLiteralRestorationFragments(t *testing.T) {
	for _, tc := range []struct {
		name, raw, suffix string
	}{
		{"adjacent", "`first``second`", ""},
		{"intervening", "`first`middle`second`", ""},
		{"several", "`one``two``three`", ""},
		{"empty", "````", ""},
		{"unicode", "`first 雪`middle`second <DEF> tag`", ""},
		{"escapes", "`first \\` \\\\ \"quoted\"``second`", ""},
		{"crlf", "`first\r\nline``second`", ""},
		{"external comma", "`first``second`", ","},
		{"internal comma", "`first,``second,`", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			prepared, statement := refactorPreparedStatement(t, "\"message\" is "+tc.raw+tc.suffix+".\n")
			if len(prepared.Literals) < 2 {
				t.Fatal("reproduction did not prepare multiple literal spans")
			}
			var literal *token
			for _, candidate := range statement.tokens {
				if candidate.mem != nil && candidate.mem.Primitive != nil && candidate.mem.Primitive.Type == valueText {
					if literal != nil {
						t.Fatal("expected one composite Text token")
					}
					literal = candidate
				}
			}
			if literal == nil || statement.sGram.predicate.obj.noun != literal {
				t.Fatal("grammar did not retain the original composite Text token")
			}
			if literal.conjoined != (tc.suffix == ",") {
				t.Fatal("token lost its comma boundary")
			}
			literal.binding = &tokenBinding{Record: &record{ID: "retained owner"}, Root: "retained root"}
			literal.bindingError = errors.New("retained diagnostic")
			literal.mem.Runtime = &runtimeReference{OwnerID: "retained app", Member: memberOutput, Source: statement}
			beforeToken, beforeEntry := *literal, *literal.mem
			beforeBinding, beforeRuntime := *literal.binding, *literal.mem.Runtime
			beforeSource, err := json.Marshal(prepared)
			if err != nil {
				t.Fatal(err)
			}
			if err := statement.uncacheLiterals(prepared.TextCache); err != nil {
				t.Fatal(err)
			}
			if string(literal.rawBytes) != tc.raw || literal.mem.Primitive.Text != tc.raw {
				t.Errorf("restored bytes=%q, Text=%q; want %q", literal.rawBytes, literal.mem.Primitive.Text, tc.raw)
			}
			beforeToken.rawBytes, beforeEntry.Primitive = literal.rawBytes, literal.mem.Primitive
			if literal.mem != beforeToken.mem || literal.binding != beforeToken.binding || literal.mem.Runtime != beforeEntry.Runtime ||
				!reflect.DeepEqual(*literal, beforeToken) || !reflect.DeepEqual(*literal.mem, beforeEntry) ||
				!reflect.DeepEqual(*literal.binding, beforeBinding) || !reflect.DeepEqual(*literal.mem.Runtime, beforeRuntime) {
				t.Fatal("restoration changed token identity or metadata")
			}
			// Neither the output buffer nor the primitive string may alias source storage.
			literal.rawBytes[1] = 'X'
			if literal.mem.Primitive.Text != tc.raw {
				t.Error("primitive Text aliases the mutable output buffer or was not restored")
			}
			afterSource, err := json.Marshal(prepared)
			if err != nil || !bytes.Equal(beforeSource, afterSource) {
				t.Fatalf("restoration/output mutation changed source, pool, or provenance: %v", err)
			}
		})
	}
}

func TestLiteralRestorationOriginalOccurrences(t *testing.T) {
	const first = "`11111111-1111-1111-1111-111111111111`"
	const second = "`22222222-2222-2222-2222-222222222222`"
	for _, tc := range []struct {
		name, raw, want string
		pool            map[string][]byte
	}{
		{"repeated", first + first, "`one``one`", map[string][]byte{first: []byte("`one`")}},
		{"inserted key", first + second, second + "`payload`", map[string][]byte{first: []byte(second), second: []byte("`payload`")}},
		{"unknown spans", "`unknown`" + first + "middle`missing`", "`unknown``one`middle`missing`", map[string][]byte{first: []byte("`one`")}},
		{"identifier", "\"name" + first + "\"", "\"name" + first + "\"", map[string][]byte{first: []byte("`one`")}},
		{"escaped identifier", "\"name\\\"" + first + "\"", "\"name\\\"" + first + "\"", map[string][]byte{first: []byte("`one`")}},
		{"quoted segment", first + "\"name\\\"" + second + "\"" + second, "`one`\"name\\\"" + second + "\"`two`", map[string][]byte{first: []byte("`one`"), second: []byte("`two`")}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tok := literalRestorationToken(t, tc.raw)
			beforePool, err := json.Marshal(tc.pool)
			if err != nil {
				t.Fatal(err)
			}
			if err := tok.restoreLiteral(tc.pool); err != nil {
				t.Fatal(err)
			}
			if string(tok.rawBytes) != tc.want {
				t.Fatalf("restored %q; want %q", tok.rawBytes, tc.want)
			}
			if tc.raw[0] == '`' && (tok.mem.Primitive.Type != valueText || tok.mem.Primitive.Text != tc.want) {
				t.Fatalf("restored primitive = %+v; want Text %q", tok.mem.Primitive, tc.want)
			}
			afterPool, err := json.Marshal(tc.pool)
			if err != nil || !bytes.Equal(beforePool, afterPool) {
				t.Fatalf("restoration mutated pool: %v", err)
			}
		})
	}
}

func TestLiteralRestorationFragmentFailureIsAtomic(t *testing.T) {
	const first = "`11111111-1111-1111-1111-111111111111`"
	const second = "`22222222-2222-2222-2222-222222222222`"
	for _, tc := range []struct {
		name, raw, invalid, diagnostic string
		storage                        string
	}{
		{"empty second", first + second, "", "newPrimitive unable to process", ""},
		{"unquoted second", first + second, "not quoted", "newPrimitive unable to process", ""},
		{"number middle", first + second + first, "42", "cached value has type Number, want Text", ""},
		{"truth middle", first + second + first, "true", "cached value has type Truth, want Text", ""},
		{"identifier middle", first + second + first, `"name"`, "cached value has type ID, want Text", ""},
		{"unclosed backtick", first + "`unclosed", "`two`", "unterminated ` literal", ""},
		{"unclosed identifier", first + `"unclosed`, "`two`", "unterminated \" literal", ""},
		{"wrong storage", first + second, "`two`", "cached token has type Number, want Text", "number"},
		{"missing entry", first + second, "`two`", "restore literal", "nil"},
		{"invalid assembled text", "prefix" + first, "`two`", "newPrimitive unable to process", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Malformed internal tokens cannot go through the source scanner; start
			// with valid storage and alter only the boundary under test.
			tok := literalRestorationToken(t, first)
			tok.rawBytes = []byte(tc.raw)
			tok.mem.Primitive.Text = tc.raw
			tok.binding = &tokenBinding{Record: &record{ID: "retained"}, Root: "retained"}
			tok.bindingError = errors.New("retained diagnostic")
			tok.mem.Runtime = &runtimeReference{OwnerID: "retained", Member: memberOutput}
			if tc.storage == "number" {
				tok.mem.Primitive = &primitive{Type: valueNumber, Number: 42}
			}
			entryPointer := tok.mem
			if tc.storage == "nil" {
				tok.mem = nil
			}
			beforeToken, beforeEntry, beforePrimitive := *tok, *entryPointer, *entryPointer.Primitive
			beforeBinding, beforeRuntime := *tok.binding, *entryPointer.Runtime
			beforeRaw := bytes.Clone(tok.rawBytes)
			pool := map[string][]byte{first: []byte("`one`"), second: []byte(tc.invalid)}
			beforePool, err := json.Marshal(pool)
			if err != nil {
				t.Fatal(err)
			}
			err = tok.restoreLiteral(pool)
			if err == nil || !strings.Contains(err.Error(), "restore literal") || !strings.Contains(err.Error(), tc.diagnostic) {
				t.Errorf("restoration error = %v; want %q", err, tc.diagnostic)
			}
			if tok.mem != beforeToken.mem || entryPointer.Primitive != beforeEntry.Primitive || tok.binding != beforeToken.binding || entryPointer.Runtime != beforeEntry.Runtime ||
				!bytes.Equal(tok.rawBytes, beforeRaw) || !reflect.DeepEqual(*tok, beforeToken) ||
				!reflect.DeepEqual(*entryPointer, beforeEntry) || !reflect.DeepEqual(*entryPointer.Primitive, beforePrimitive) ||
				!reflect.DeepEqual(*tok.binding, beforeBinding) || !reflect.DeepEqual(*entryPointer.Runtime, beforeRuntime) {
				t.Error("failed restoration partially mutated token or storage")
			}
			afterPool, err := json.Marshal(pool)
			if err != nil || !bytes.Equal(beforePool, afterPool) {
				t.Fatalf("failed restoration changed cache: %v", err)
			}
		})
	}
	// A complete cache miss must not demand Text storage from other lexical tokens.
	tok := literalRestorationToken(t, `"identifier"`)
	tok.mem = nil
	if err := tok.restoreLiteral(map[string][]byte{first: []byte("`one`")}); err != nil || tok.mem != nil || string(tok.rawBytes) != `"identifier"` {
		t.Fatalf("cache miss changed an unrelated token: %+v, %v", tok, err)
	}
}

func TestLiteralFragmentsImportedOccurrences(t *testing.T) {
	root := t.TempDir()
	const source = "\"message\" is `first`middle`second`.\n"

	leaf := writeSourceFile(t, root, "leaf.l", source)
	main := writeSourceFile(t, root, "main.l", "Use `leaf.l`.\nUse `leaf.l`.\n")
	built, err := buildTemporarySource(t, main)
	if err != nil {
		t.Fatal(err)
	}
	statements := testPhysicalStatements(root, built.SourceStatements)
	if len(statements) != 2 {
		t.Fatalf("imported statements = %d; want 2", len(built.SourceStatements))
	}
	before, err := json.Marshal(built)
	if err != nil {
		t.Fatal(err)
	}
	var restored []*token
	for _, statement := range statements {
		if statement.Span.Path != leaf || statement.Span.Start.Offset != 0 || statement.Span.End.Offset != len(source) || string(statement.Raw) != source {
			t.Fatalf("import lost source provenance: %+v", statement)
		}
		words := newParser(statementBody(statement.Prepared)).parseTokens()
		if len(words) != 3 {
			t.Fatalf("imported declaration tokens = %q", words)
		}
		tok := literalRestorationToken(t, string(words[2]))
		if err := tok.restoreLiteral(built.TextCache); err != nil {
			t.Fatal(err)
		}
		if string(tok.rawBytes) != "`first`middle`second`" || tok.mem.Primitive.Text != "`first`middle`second`" {
			t.Errorf("imported Text = %q / %q", tok.rawBytes, tok.mem.Primitive.Text)
		}
		restored = append(restored, tok)
	}
	fragments := 0
	for id, raw := range built.TextCache {
		if string(raw) != "`first`" && string(raw) != "`second`" {
			continue
		}
		fragments++
		occurrences := 0
		for _, statement := range statements {
			occurrences += bytes.Count(statement.Prepared, []byte(id))
		}
		if occurrences != 1 {
			t.Errorf("literal ID %q appears in %d imported occurrences", id, occurrences)
		}
	}
	if fragments != 4 {
		t.Errorf("literal IDs = %d; want 4 independent occurrences", fragments)
	}
	second := bytes.Clone(restored[1].rawBytes)
	restored[0].rawBytes[1] = 'X'
	if !bytes.Equal(restored[1].rawBytes, second) {
		t.Error("imported tokens share mutable output storage")
	}
	after, err := json.Marshal(built)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("restoration changed imported source or literal pool: %v", err)
	}
	if disk, err := os.ReadFile(leaf); err != nil || string(disk) != source {
		t.Fatalf("imported file changed: %q, %v", disk, err)
	}
}

func TestLiteralFragmentsCompileToExactAML(t *testing.T) {
	fixture := memberResolutionFixture(t)
	const anchor = "`Hello World!`"
	if strings.Count(fixture.main, anchor) != 1 {
		t.Fatal("expected one source literal replacement point")
	}
	baseline := compileMemberResolutionSource(t, fixture)
	requireMemberResolutionProgram(t, baseline)
	if bytes.Count(baseline.ExpandedBytes, []byte(anchor)) != 1 {
		t.Fatal("expected one expanded-source replacement point")
	}
	for _, tc := range []struct{ name, raw, scalar string }{
		{"adjacent", "`first``second`", "first``second"},
		{"intervening", "`first`middle`second`", "first`middle`second"},
		{"unicode", "`first 雪`middle`second <DEF> tag`", "first 雪`middle`second <DEF> tag"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := fixture
			source.main = strings.Replace(source.main, anchor, tc.raw, 1)
			comp := compileMemberResolutionSource(t, source)
			requireMemberResolutionProgram(t, comp)
			x := comp.Memory["x"]
			if x == nil || x.Value == nil || x.Value.Type != storagePrimitive || x.Value.Primitive == nil || x.Value.Primitive.Type != valueText {
				t.Fatal("compilation did not retain a primitive Text record")
			}
			if x.Value.Primitive.Text != tc.raw {
				t.Errorf("semantic Text = %q; want %q", x.Value.Primitive.Text, tc.raw)
			}
			wantExpanded := bytes.Replace(baseline.ExpandedBytes, []byte(anchor), []byte(tc.raw), 1)
			if !bytes.Equal(comp.ExpandedBytes, wantExpanded) {
				t.Errorf("expanded source differs, including separators/EOF:\n%s", comp.ExpandedBytes)
			}
			wantPrint := []byte("Inputs: ~&[" + strconv.Quote(tc.scalar) + "]&~")
			for name, markup := range map[string][]byte{"observable": comp.ASMBytes, "backend": comp.Program.BackendBytes} {
				if !bytes.Contains(markup, wantPrint) {
					t.Errorf("%s AML missing exact %s:\n%s", name, wantPrint, markup)
				}
			}
			before, err := json.Marshal(struct {
				Prepared *PreparedSource
				Pool     map[string][]byte
			}{comp.Prepared, comp.TextCache})
			if err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 2; i++ {
				if rendered := renderExpandedSource(comp.SourceStatements, comp.TextCache); !bytes.Equal(rendered, wantExpanded) {
					t.Errorf("repeated rendering %d changed source bytes", i)
				}
			}
			after, err := json.Marshal(struct {
				Prepared *PreparedSource
				Pool     map[string][]byte
			}{comp.Prepared, comp.TextCache})
			if err != nil || !bytes.Equal(before, after) {
				t.Fatalf("rendering changed prepared source/cache: %v", err)
			}
			t.Logf("expanded=%d bytes; required Rust expression %q", len(comp.ExpandedBytes), wantPrint)
		})
	}
}
