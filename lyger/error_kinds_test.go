package Lyger

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/l-programming-language/l/lerrors"
)

func requireFrontendErrorKind(t *testing.T, err error, kind lerrors.Kind, message string) {
	t.Helper()
	var diagnostic *lerrors.Error
	if err == nil || !errors.Is(err, kind) || !errors.As(err, &diagnostic) || diagnostic.Kind() != kind {
		t.Fatalf("error = %v; want originating category %s", err, kind)
	}
	if message != "" && err.Error() != message {
		t.Fatalf("diagnostic = %q; want %q", err.Error(), message)
	}
}

func TestFrontendErrorKindsAtOrigins(t *testing.T) {
	t.Run("invalid input", func(t *testing.T) {
		_, err := NewLygerService(Options{}).Compile(nil, "empty.l")
		requireFrontendErrorKind(t, err, lerrors.InvalidInput, "error Compilation.init(): MainBytes cannot be empty or nil")
	})
	t.Run("invalid configuration", func(t *testing.T) {
		_, err := loadGrammar(Options{Language: "../eng"})
		requireFrontendErrorKind(t, err, lerrors.InvalidConfiguration, `load grammar: valid language is required: "../eng"`)
	})
	t.Run("source syntax", func(t *testing.T) {
		_, err := PrepareSource("broken.l", []byte("`unterminated"))
		requireFrontendErrorKind(t, err, lerrors.Syntax, "broken.l: unterminated ` literal at byte 0")
	})
	t.Run("ambiguous reference", func(t *testing.T) {
		lexeme := newLexeme("verb")
		if err := lexeme.reLabel(`"Write|-s"`); err != nil {
			t.Fatal(err)
		}
		one := &record{ID: "One", Value: &entry{Type: storageLexeme, Lexeme: lexeme}}
		two := &record{ID: "Two", Value: &entry{Type: storageLexeme, Lexeme: lexeme}}
		binding, err := resolveTokenBinding([]byte("Writes"), map[string]*record{"z": two, "a": one})
		requireFrontendErrorKind(t, err, lerrors.AmbiguousReference, `ambiguous reference "Writes": One and Two`)
		if binding != nil {
			t.Fatal("ambiguous lookup returned a binding")
		}
	})
	t.Run("type mismatch", func(t *testing.T) {
		_, operand := planTestApp("Worker")
		planner := orchestrationPlanner{semantics: newSemanticProgram(), plan: &orchestrationPlan{}, apps: make(map[*App]bool)}
		_, _, err := planner.instruction(&instruction{Kind: instructionShow, Value: operand}, newOrchestrationFlow())
		requireFrontendErrorKind(t, err, lerrors.TypeMismatch, "Show requires a scalar value, got App")
	})
	t.Run("alias cycle", func(t *testing.T) {
		first, second := &record{ID: "First"}, &record{ID: "Second"}
		_, err := canonicalRecord(first, map[*record]*record{first: second, second: first})
		requireFrontendErrorKind(t, err, lerrors.Cycle, "alias cycle at First")
	})
	t.Run("unsupported operator", func(t *testing.T) {
		literal := &instructionValue{Kind: operandLiteral, Type: valueText, Literal: &primitive{Type: valueText, Text: "`text`"}}
		_, err := (&orchestrationPlanner{}).condition(&instructionCondition{Operator: "contains", Left: literal, Right: literal}, newOrchestrationFlow())
		requireFrontendErrorKind(t, err, lerrors.Unsupported, "unsupported condition operator contains")
	})
	t.Run("internal invariant", func(t *testing.T) {
		_, err := (&orchestrationEmitter{}).literalValue(nil)
		requireFrontendErrorKind(t, err, lerrors.Internal, "literal has no primitive value")
	})
}

func TestFrontendErrorKindsThroughPublicCompilation(t *testing.T) {
	for _, tc := range []struct {
		name, source, phase, detail string
		kind                        lerrors.Kind
	}{
		{"terminal unresolved", "Show Missing.\n", ":1:1: bind reference:", "unresolved references after dependency resolution", lerrors.UnresolvedReference},
		{"duplicate public declaration", "Public \"dup\" is 1.\nPublic \"dup\" is 2.\n", ":2:1: register statement:", "duplicate declaration dup", lerrors.DuplicateDeclaration},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "main.l")
			result, err := NewLygerService(canonicalTestOptions()).Compile([]byte(tc.source), path)
			requireFrontendErrorKind(t, err, tc.kind, "")
			if result != nil || !strings.Contains(err.Error(), path+tc.phase) || !strings.Contains(err.Error(), tc.detail) {
				t.Fatalf("public failure lost source, phase, or detail: result=%v err=%v", result, err)
			}
		})
	}
}

func TestFrontendErrorContextPreservesCategoryAndCallSite(t *testing.T) {
	cause := lerrors.TypeMismatch.Errorf("operand mismatch")
	definition := SourceSpan{Path: "definition.l", Start: SourcePosition{Line: 4, Column: 5}}
	call := &lStatement{source: &SourceStatement{Span: SourceSpan{Path: "caller.l", Start: SourcePosition{Line: 9, Column: 3}}}}
	for _, tc := range []struct {
		name    string
		current *instruction
		want    string
	}{
		{"nil instruction", nil, "orchestrator: operand mismatch"},
		{"no source", &instruction{Kind: instructionShow}, "orchestrator show: operand mismatch"},
		{"definition source", &instruction{Kind: instructionShow, Source: definition}, "definition.l:4:5: orchestrator show: operand mismatch"},
		{"expanded call site", &instruction{Kind: instructionShow, Source: definition, CallSite: call}, "caller.l:9:3: orchestrator show: operand mismatch"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.current.wrapError(nil); err != nil {
				t.Fatalf("nil failure became an error: %v", err)
			}
			err := tc.current.wrapError(cause)
			requireFrontendErrorKind(t, err, lerrors.TypeMismatch, tc.want)
			if !errors.Is(err, cause) {
				t.Fatal("instruction context replaced the originating error")
			}
		})
	}
	if err := definition.wrapError(nil); err != nil {
		t.Fatalf("source span manufactured an error: %v", err)
	}
	requireFrontendErrorKind(t, definition.wrapError(cause), lerrors.TypeMismatch, "definition.l:4:5: operand mismatch")
}

func TestFrontendGrammarAggregateCategories(t *testing.T) {
	first := lerrors.TypeMismatch.Errorf("wrong operand")
	second := lerrors.UnresolvedReference.Errorf("missing operand")
	for _, empty := range []bool{false, true} {
		name := "classified alternatives"
		cases := map[string]string{"b": "second", "a": "first"}
		want := "no grammar case matches (a: wrong operand; at end of statement; b: missing operand; at end of statement)"
		if empty {
			name, cases, want = "empty alternatives", nil, "no grammar case matches ()"
		}
		t.Run(name, func(t *testing.T) {
			state := 7
			err := selectGrammarCase(&state, cases, func(v int) int { return v },
				func(v *int, rule string) error {
					*v = 99
					if rule == "first" {
						return first
					}
					return second
				}, func(*int) []*token { return nil }, func(*int, string) {}, nil)
			requireFrontendErrorKind(t, err, lerrors.Syntax, want)
			if state != 7 {
				t.Fatal("failed grammar alternatives changed the target")
			}
			if !empty && (!errors.Is(err, first) || !errors.Is(err, second) || !errors.Is(err, lerrors.TypeMismatch) || !errors.Is(err, lerrors.UnresolvedReference)) {
				t.Fatal("aggregate lost its classified alternative causes")
			}
		})
	}
}

func TestFrontendLoweringDistinguishesUnsupportedFromInternal(t *testing.T) {
	for _, tc := range []struct {
		name      string
		condition *instructionCondition
		kind      lerrors.Kind
	}{
		{"missing condition", nil, lerrors.Internal},
		{"compound condition", &instructionCondition{Operator: "and", Terms: []*instructionCondition{{Operator: "truth"}, {Operator: "truth"}}}, lerrors.Unsupported},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := (&orchestrationPlanner{}).condition(tc.condition, newOrchestrationFlow())
			requireFrontendErrorKind(t, err, tc.kind, "condition requires one supported operator")
		})
	}
	for _, tc := range []struct {
		name    string
		current *instruction
		kind    lerrors.Kind
	}{
		{"missing action expansion", &instruction{Kind: instructionCall, Synchronous: true}, lerrors.Internal},
		{"asynchronous action", &instruction{Kind: instructionCall, Actions: []*instruction{{Kind: instructionExit}}}, lerrors.Unsupported},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := (&orchestrationPlanner{}).instruction(tc.current, newOrchestrationFlow())
			requireFrontendErrorKind(t, err, tc.kind, "source action requires synchronous expanded instructions")
		})
	}
}

func TestFrontendRequiredGrammarTokenAbsenceIsSyntax(t *testing.T) {
	for _, tc := range []struct {
		name, raw, want string
		associate       func(grammarTerm) error
	}{
		{"typed token", "`~`", "Error: primitive.text is not an optional type reference!", func(term grammarTerm) error {
			_, err := term.typeToken(nil)
			return err
		}},
		{"single lexical token", "ARTICLE", "Error: article is not an optional type!", func(term grammarTerm) error {
			_, err := term.lexicalToken(nil, lexicalMatchClass)
			return err
		}},
		{"lexical token sequence", "[ADJECTIVE]", "Error: adjective is not an optional type!", func(term grammarTerm) error {
			var associations []*token
			return term.associateLexicalTokens(nil, &associations, lexicalMatchClass)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			term, err := decodeGrammarTerm(tc.raw)
			if err != nil || term.Optional {
				t.Fatalf("required grammar term %q invalid: %+v, %v", tc.raw, term, err)
			}
			err = tc.associate(term)
			requireFrontendErrorKind(t, err, lerrors.Syntax, tc.want)
			if errors.Is(err, lerrors.InvalidConfiguration) {
				t.Fatal("valid grammar with missing source operand was classified as invalid configuration")
			}
			term.Optional = true
			if err := tc.associate(term); err != nil {
				t.Fatalf("optional absence changed: %v", err)
			}
		})
	}
}
