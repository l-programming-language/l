package Lyger

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func refactorBenchmarkSource(count int) []byte {
	source := []byte{}
	for i := 0; i < count; i++ {
		source = fmt.Appendf(source, "\"value%d\" is `snow 雪\nline %d`.\n*** comment ***\n", i, i)
	}
	return source
}

func BenchmarkRefactorPrepareSource(b *testing.B) {
	for _, count := range []int{64, 512, 2048} {
		b.Run(fmt.Sprintf("statements=%d", count), func(b *testing.B) {
			source := refactorBenchmarkSource(count)
			b.SetBytes(int64(len(source)))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				prepared, err := PrepareSource("bench.l", source)
				if err != nil || len(prepared.Statements) != count {
					b.Fatalf("prepare: %v", err)
				}
			}
		})
	}
}

func BenchmarkRefactorRenderExpandedSource(b *testing.B) {
	for _, count := range []int{32, 256} {
		b.Run(fmt.Sprintf("statements=%d", count), func(b *testing.B) {
			source := refactorBenchmarkSource(count)
			source = append(source, []byte("Define \"Sequence\" as `<DEF>W^<@Parts|Syllables@><DEF>`.\n")...)
			prepared, err := PrepareSource("bench.l", source)
			if err != nil {
				b.Fatal(err)
			}
			b.SetBytes(int64(len(source)))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if len(renderExpandedSource(prepared.Statements, prepared.TextCache)) == 0 {
					b.Fatal("empty source view")
				}
			}
		})
	}
}

func refactorBenchmarkWrite(b *testing.B, root, name, source string) string {
	b.Helper()
	path := filepath.Join(root, name)
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		b.Fatal(err)
	}
	return path
}

func BenchmarkRefactorImportExpansion(b *testing.B) {
	for _, shape := range []string{"chain", "repeated"} {
		for _, size := range []int{8, 32} {
			b.Run(fmt.Sprintf("%s/files=%d", shape, size), func(b *testing.B) {
				root := b.TempDir()

				var main strings.Builder
				if shape == "chain" {
					main.WriteString("Use `part0.l`.\n")
					for i := 0; i < size; i++ {
						source := fmt.Sprintf("\"value%d\" is `raw 雪`.\n", i)
						if i+1 < size {
							source = fmt.Sprintf("Use `part%d.l`.\n", i+1) + source
						}
						refactorBenchmarkWrite(b, root, fmt.Sprintf("part%d.l", i), source)
					}
				} else {
					refactorBenchmarkWrite(b, root, "leaf.l", "\"message\" is `raw 雪`.\n")
					for i := 0; i < size; i++ {
						main.WriteString("Use `leaf.l`.\n")
					}
				}
				path := refactorBenchmarkWrite(b, root, "main.l", main.String())
				raw := []byte(main.String())
				gram := scopeTestGrammar(b)
				if gram == nil {
					b.Fatal("fixture grammar failed to load")
				}
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					prepared, err := PrepareSource(path, raw)
					if err != nil {
						b.Fatal(err)
					}
					built := newPreparedBuilder(prepared, gram)
					if err = built.process(); err != nil || len(testPhysicalStatements(root, built.SourceStatements)) != size {
						b.Fatalf("expand: %v", err)
					}
				}
			})
		}
	}
}

func BenchmarkRefactorImportRecognition(b *testing.B) {
	for _, source := range []struct{ name, text string }{
		{"path", "Use `folder/file.l`.\n"},
		{"member", "Use Library's Render from `folder/file.l` as \"Pretty Print\".\n"},
		{"not-import", "Show `ordinary Text`.\n"},
	} {
		b.Run(source.name, func(b *testing.B) {
			gram, raw := scopeTestGrammar(b), []byte(source.text)
			if gram == nil {
				b.Fatal("fixture grammar failed to load")
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				matched, _, err := gram.isImportStatement(raw)
				if err != nil {
					b.Fatal(err)
				}
				if matched != (source.name != "not-import") {
					b.Fatal("import classification changed")
				}
			}
		})
	}
}

func BenchmarkRefactorLiteralRestoration(b *testing.B) {
	for _, size := range []int{16, 256, 1024} {
		for _, fragments := range []int{1, 3} {
			name := fmt.Sprintf("cache=%d/tokens=8", size)
			if fragments > 1 {
				name += fmt.Sprintf("/fragments=%d", fragments)
			}
			b.Run(name, func(b *testing.B) {
				pool := make(map[string][]byte, size)
				keys := make([]string, size)
				for i := range keys {
					keys[i] = fmt.Sprintf("`literal-%08d`", i)
					pool[keys[i]] = []byte("`exact snow 雪`")
				}
				var raw [8]string
				for i := range raw {
					raw[i] = strings.Join(keys[i:i+fragments], "")
				}
				want := []byte(strings.Repeat("`exact snow 雪`", fragments))
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					// Include fresh mutable token/wrapper construction in every sample.
					// No timed iteration can reuse already-restored tokens.
					tokens := make([]*token, 8)
					for j := range tokens {
						tokens[j] = &token{rawBytes: []byte(raw[j]), tType: "primitive.text", mem: &entry{Type: storagePrimitive, Primitive: &primitive{Type: valueText, Text: raw[j]}}}
					}
					statement := &lStatement{tokens: tokens, sGram: &statementGram{tokens: tokens, predicate: &predicateGram{tokens: tokens, obj: &objectGram{tokens: tokens, noun: tokens[0]}}}}
					if err := statement.uncacheLiterals(pool); err != nil || !bytes.Equal(tokens[0].rawBytes, want) {
						b.Fatalf("restore: %v", err)
					}
				}
			})
		}
	}
}

func BenchmarkRefactorBinding(b *testing.B) {
	for _, size := range []int{32, 512} {
		for _, ambiguous := range []bool{false, true} {
			b.Run(fmt.Sprintf("records=%d/ambiguous=%t", size, ambiguous), func(b *testing.B) {
				memory := make(map[string]*record, size+2)
				for i := 0; i < size; i++ {
					name := fmt.Sprintf("Item%04d", i)
					memory[name] = &record{ID: name, Value: &entry{Type: storagePrimitive, Primitive: &primitive{Type: valueText, Text: "`value`"}}}
				}
				lex := newLexeme("verb")
				if err := lex.reLabel(`"Render|-s|-ed|-ing|-er"`); err != nil {
					b.Fatal(err)
				}
				memory["zz-first"] = &record{ID: "First", Value: &entry{Type: storageLexeme, Lexeme: lex}}
				if ambiguous {
					memory["zz-second"] = &record{ID: "Second", Value: &entry{Type: storageLexeme, Lexeme: lex}}
				}
				query := []byte("ReNdErS")
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					binding, err := resolveTokenBinding(query, memory)
					if ambiguous && err == nil || !ambiguous && (err != nil || binding == nil) {
						b.Fatalf("binding = %v, %v", binding, err)
					}
				}
			})
		}
	}
}

func BenchmarkRefactorGrammarSequence(b *testing.B) {
	config := canonicalTestOptions()
	gram, err := loadGrammar(config)
	if err != nil {
		b.Fatal(err)
	}
	for _, count := range []int{4, 16} {
		b.Run(fmt.Sprintf("objects=%d", count), func(b *testing.B) {
			source := strings.TrimSuffix(strings.Repeat("any Text and ", count), " and ")
			statement, err := newLStatement([]byte(source), gram)
			if err != nil {
				b.Fatal(err)
			}
			if err = statement.tokenize(); err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				term, err := decodeGrammarTerm("[object]")
				if err != nil {
					b.Fatal(err)
				}
				objects, _, err := buildGrammarObjects(term, statement.tokens, gram)
				if err != nil || len(objects) != count {
					b.Fatalf("sequence returned %d objects: %v", len(objects), err)
				}
			}
		})
	}
}

func BenchmarkRefactorInstructionOrdering(b *testing.B) {
	for _, count := range []int{16, 128} {
		b.Run(fmt.Sprintf("clauses=%d", count), func(b *testing.B) {
			tokens := make([]*token, count*4)
			for i := range tokens {
				tokens[i] = &token{}
			}
			clauses := make([]*clauseGram, count)
			for i := range clauses {
				clauses[i] = &clauseGram{verb: tokens[(count-i-1)*4]}
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ordered := orderedInstructionClauses(clauses, tokens)
				if len(ordered) != count || ordered[0].verb != tokens[0] {
					b.Fatal("instruction order changed")
				}
			}
		})
	}
}
