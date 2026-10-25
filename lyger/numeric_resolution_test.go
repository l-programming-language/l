package Lyger

import (
	"bytes"
	"errors"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/l-programming-language/l/lerrors"
)

func TestNumericResolutionConfiguredVocabulary(t *testing.T) {
	for _, inflected := range []bool{false, true} {
		name := "root"
		if inflected {
			name = "inflection"
		}
		t.Run(name, func(t *testing.T) {
			config := canonicalTestOptions()
			gram, err := loadGrammar(config)
			if err != nil {
				t.Fatal(err)
			}
			lex := gram.lex.Verbs["observative"]
			lex.Root = "1e999"
			wantType := "verbs.observative.root"
			if inflected {
				lex.Root = "1e"
				lex.Postfixes = map[string]string{"-999": "active"}
				wantType = "verbs.observative.active"
			}
			gram.lex.Verbs["observative"] = lex
			tok := newToken([]byte("1e999,"), gram)
			if err := tok.scan(); err != nil {
				t.Fatalf("configured word became a Number: %v", err)
			}
			if tok.tType != wantType || tok.mem.Type != storageLexeme || tok.mem.Lexeme.Root != lex.Root || !tok.conjoined || string(tok.rawBytes) != "1e999" || tok.bindingError != nil {
				t.Fatalf("configured word lost lexical identity or comma handling: %+v", tok)
			}
		})
	}
	t.Run("complete compilation", func(t *testing.T) {
		dir := t.TempDir()
		source := []byte("1e999 `payload`.\n")
		c := NewLygerService(canonicalTestOptions()).NewCompilation(source, filepath.Join(dir, "main.l"))
		if c.Error != nil {
			t.Fatal(c.Error)
		}
		lex := c.Gram.lex.Verbs["observative"]
		lex.Root = "1e999"
		c.Gram.lex.Verbs["observative"] = lex
		c.Make()
		if c.Error != nil || c.Program == nil || !bytes.Equal(c.MainBytes, source) {
			t.Fatalf("configured action did not compile unchanged: %v", c.Error)
		}
		if len(c.Semantics.Instructions) != 1 {
			t.Fatalf("configured Show emitted %d instructions", len(c.Semantics.Instructions))
		}
		instruction := c.Semantics.Instructions[0]
		if instruction.Kind != instructionShow || instruction.Value == nil || instruction.Value.Kind != operandLiteral || instruction.Value.Type != valueText || instruction.Value.Literal == nil || instruction.Value.Literal.Text != "`payload`" {
			t.Fatalf("configured verb lost its Show operation or exact operand: %+v", instruction)
		}
	})
}

func TestNumericResolutionSourceNames(t *testing.T) {
	for _, tc := range []struct {
		name, main, library, canonical, origin string
		pronoun                                bool
	}{
		{"declared", "\"1e999\" is `payload`.\nShow 1e999.\n", "", "1e999", "main.l", false},
		{"forward declaration", "Show 1e999.\n\"1e999\" is `payload`.\n", "", "1e999", "main.l", false},
		{"alias", "\"origin\" is `payload`.\nUse origin as \"1e999\".\nShow 1e999.\n", "", "origin", "main.l", false},
		{"import", "Use `local-words.l`.\nShow 1e999.\n", "\"1e999\" is `payload`.\n", "1e999", "local-words.l", false},
		{"imported alias", "Use Tools's Item from `local-words.l` as \"1e999\".\nShow 1e999.\n", "Create \"Tools\" as a Dictionary.\nTools's \"Item\" is `payload`.\n", "Tools.Item", "local-words.l", false},
		{"pronoun", "\"1e999\" is `payload`.\nShow it.\n", "", "1e999", "main.l", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := compileMemberResolutionSource(t, memberResolutionSource{main: tc.main, library: tc.library})
			if c.Error != nil || c.Program == nil || len(c.ProcStatements) != 0 {
				t.Fatalf("declared numeric-looking name failed: %v", c.Error)
			}
			rec := c.Memory[tc.canonical]
			if rec == nil || rec.Value.Primitive == nil || rec.Value.Primitive.Text != "`payload`" {
				t.Fatalf("name lost its original Text value: %+v", rec)
			}
			declaration := c.Semantics.Declarations[rec]
			if declaration == nil || declaration.source == nil || filepath.Base(declaration.source.Span.Path) != tc.origin {
				t.Fatal("canonical declaration lost source attribution")
			}
			if !tc.pronoun {
				paragraphBindingUses(t, c, "1e999", []*record{rec})
			} else {
				last := c.Statements[len(c.Statements)-1]
				found, err := c.resolver().reference(last.tokens[1])
				if err != nil || found != rec || !last.tokens[1].isPronoun() {
					t.Fatalf("pronoun lost its declaring subject: %v", err)
				}
			}
			if !bytes.Equal(c.Prepared.Original, []byte(tc.main)) || !bytes.Equal(c.ExpandedBytes, renderExpandedSource(c.SourceStatements, c.TextCache)) {
				t.Fatal("numeric resolution changed source or expanded rendering")
			}
			foundPayload := false
			for _, value := range c.TextCache {
				if string(value) == "`payload`" {
					foundPayload = true
				}
			}
			if !foundPayload {
				t.Fatal("literal pool lost the original payload bytes")
			}
		})
	}
}

func TestNumericResolutionLocalShadowing(t *testing.T) {
	source := "\"1e999\" is `outer`.\n\n" +
		"    Show 1e999.\n\"1e999\" is `inner`.\nShow 1e999.\n\nShow 1e999.\n"
	c := compileMemberResolutionSource(t, memberResolutionSource{main: source})
	if c.Error != nil {
		t.Fatal(c.Error)
	}
	declarations := paragraphBindingDeclarations(t, c, "1e999", 2)
	outer, local := declarations[0].record, declarations[1].record
	if outer == local || outer.storageScope != c.Scope.Root || local.storageScope == c.Scope.Root || c.Memory["1e999"] != outer {
		t.Fatal("local numeric-looking name replaced or reused its outer record")
	}
	if outer.Value.Primitive.Text != "`outer`" || local.Value.Primitive.Text != "`inner`" || declarations[1].statement.source.Span.Start.Line != 4 {
		t.Fatal("local declaration lost its value or original location")
	}
	paragraphBindingUses(t, c, "1e999", []*record{local, local, outer})
}

func TestNumericResolutionPendingNames(t *testing.T) {
	for _, tc := range []struct{ name, source, location string }{
		{"root reservation", "Show 1e999.\n\"1e999\" is Missing.\n", ":1:1: bind reference:"},
		{"local reservation shadows outer", "\"1e999\" is `outer`.\n\n    Show 1e999.\n\"1e999\" is Missing.\n", ":3:5: bind reference:"},
		{"ordinary missing name", "Show Missing.\n", ":1:1: bind reference:"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := compileMemberResolutionSource(t, memberResolutionSource{main: tc.source})
			if c.Error == nil || !errors.Is(c.Error, lerrors.UnresolvedReference) || errors.Is(c.Error, strconv.ErrRange) || !strings.Contains(c.Error.Error(), tc.location) {
				t.Fatalf("pending name became a numeric failure or lost provenance: %v", c.Error)
			}
		})
	}
}

func TestNumericResolutionMemberNames(t *testing.T) {
	for _, form := range []string{"Tools's 1e999", "Tools.1e999"} {
		for _, exists := range []bool{false, true} {
			name := form + "/missing"
			if exists {
				name = form + "/existing"
			}
			t.Run(name, func(t *testing.T) {
				source := "\"1e999\" is `unrelated`.\nCreate \"Tools\" as a Dictionary.\n"
				if exists {
					source += "Tools's \"1e999\" is `member`.\n"
				}
				source += "Show " + form + ".\n"
				c := compileMemberResolutionSource(t, memberResolutionSource{main: source})
				if !exists {
					if c.Error == nil || !errors.Is(c.Error, lerrors.UnresolvedReference) || errors.Is(c.Error, strconv.ErrRange) || !strings.Contains(c.Error.Error(), ":3:1: bind reference:") {
						t.Fatalf("missing member used a global or became a Number: %v", c.Error)
					}
					return
				}
				if c.Error != nil {
					t.Fatal(c.Error)
				}
				member, global := c.Memory["Tools.1e999"], c.Memory["1e999"]
				if member == nil || global == nil || member == global || member.Value.Primitive.Text != "`member`" || global.Value.Primitive.Text != "`unrelated`" {
					t.Fatal("qualified member and unrelated global lost distinct identity")
				}
				last := c.Statements[len(c.Statements)-1]
				tok := last.tokens[len(last.tokens)-1]
				wantRaw := "1e999"
				if strings.Contains(form, ".") {
					wantRaw = form
				}
				if tok.binding == nil || tok.binding.Record != member || tok.bindingError != nil || string(tok.rawBytes) != wantRaw {
					t.Fatal("member binding rewrote its source token or selected the global")
				}
			})
		}
	}
}

func TestNumericResolutionTerminalOverflow(t *testing.T) {
	for _, raw := range []string{"1e999", "-1e999", "0x1p9999", strings.Repeat("9", 400)} {
		name := raw
		if len(name) > 30 {
			name = "long decimal"
		}
		t.Run(name, func(t *testing.T) {
			c := compileMemberResolutionSource(t, memberResolutionSource{main: "\"quantity\" is " + raw + ".\n"})
			var numeric *strconv.NumError
			if c.Error == nil || !errors.Is(c.Error, lerrors.InvalidInput) || !errors.Is(c.Error, strconv.ErrRange) || !errors.As(c.Error, &numeric) || numeric.Num != raw || !strings.Contains(c.Error.Error(), ":1:1: tokenize: invalid Number "+strconv.Quote(raw)) {
				t.Fatalf("overflow lost its source diagnostic or native cause: %v", c.Error)
			}
		})
	}
	for _, tc := range []struct{ name, main, library, filename, location, number string }{
		{"earlier unknown", "Show Missing.\n\"quantity\" is 1e999.\n", "", "main.l", ":2:1:", "1e999"},
		{"earlier reserved name", "Show 1e999.\n\"1e999\" is Missing.\n\"quantity\" is 2e999.\n", "", "main.l", ":3:1:", "2e999"},
		{"imported overflow", "Use `local-words.l`.\n", "*** heading ***\n\n    \"quantity\" is 1e999.\n", "local-words.l", ":3:5:", "1e999"},
		{"comma", "\"quantity\" is 1e999, and \"other\" is 1.\n", "", "main.l", ":1:1:", "1e999"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := compileMemberResolutionSource(t, memberResolutionSource{main: tc.main, library: tc.library})
			var numeric *strconv.NumError
			if c.Error == nil || !errors.Is(c.Error, lerrors.InvalidInput) || !errors.Is(c.Error, strconv.ErrRange) || !errors.As(c.Error, &numeric) || numeric.Num != tc.number || !strings.Contains(c.Error.Error(), tc.filename+tc.location+" tokenize:") {
				t.Fatalf("terminal overflow ordering or provenance changed: %v", c.Error)
			}
		})
	}
}

func TestNumericResolutionPrimitiveControls(t *testing.T) {
	for _, tc := range []struct {
		raw  string
		kind valueType
	}{
		{"42", valueNumber}, {"-0.125", valueNumber}, {"1e3", valueNumber},
		{"\"1e999\"", valueID}, {"`1e999`", valueText}, {"true", valueTruth},
	} {
		t.Run(tc.raw, func(t *testing.T) {
			value, err := loadPrimitive([]byte(tc.raw))
			if err != nil || value.Type != tc.kind {
				t.Fatalf("primitive control changed: %+v, %v", value, err)
			}
			if tc.kind == valueNumber {
				want, err := strconv.ParseFloat(tc.raw, 64)
				if err != nil || value.Number != want {
					t.Fatal("finite Number value changed")
				}
			}
		})
	}
}

func TestNumericResolutionAmbiguityTakesPrecedence(t *testing.T) {
	config := canonicalTestOptions()
	gram, err := loadGrammar(config)
	if err != nil {
		t.Fatal(err)
	}
	tok := newToken([]byte("1e999"), gram)
	if err := tok.scan(); err != nil {
		t.Fatalf("numeric probe blocked semantic ambiguity detection: %v", err)
	}
	lex := &lexeme{Root: "1e999", POS: "verb", Type: "generic"}
	first := &record{ID: "First", Value: &entry{Type: storageLexeme, Lexeme: lex}}
	second := &record{ID: "Second", Value: &entry{Type: storageLexeme, Lexeme: lex.cloneUse()}}
	tok.scope = &lexicalScope{records: map[string]*record{"First": first, "Second": second}}
	_, err = (semanticResolver{}).bindToken(tok)
	if !errors.Is(err, lerrors.AmbiguousReference) || errors.Is(err, strconv.ErrRange) || tok.bindingError != err || tok.binding != nil || string(tok.rawBytes) != "1e999" {
		t.Fatalf("ambiguity was hidden or mutated token identity: %v", err)
	}
}

func TestNumericResolutionDeferredCauseIdentity(t *testing.T) {
	for _, resolved := range []bool{false, true} {
		name, source := "terminal cause", "\"marker\" is `payload 雪`.\n\"quantity\" is 1e999.\n"
		if resolved {
			name, source = "resolved candidate", "\"1e999\" is `payload 雪`.\nShow 1e999.\n"
		}
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			c := NewLygerService(canonicalTestOptions()).NewCompilation([]byte(source), filepath.Join(dir, "identity.l"))
			if c.Error != nil {
				t.Fatal(c.Error)
			}
			prepared := bytes.Clone(c.Prepared.PreparedBytes)
			cache := make(map[string][]byte, len(c.TextCache))
			for key, value := range c.TextCache {
				cache[key] = bytes.Clone(value)
			}
			c.scan()
			if c.Error != nil || len(testProgramStatements(c.Statements)) != 2 {
				t.Fatalf("scan did not defer the numeric candidate: %v", c.Error)
			}
			statement := testProgramStatements(c.Statements)[1]
			position := len(statement.tokens) - 1
			tok := statement.tokens[position]
			cause := tok.rangeError
			var native *strconv.NumError
			if cause == nil || !errors.Is(cause, lerrors.InvalidInput) || !errors.As(cause, &native) || native.Num != "1e999" || tok.bindingError != nil {
				t.Fatalf("candidate lost its original numeric cause or contaminated binding state: %v", cause)
			}
			span := statement.source.Span
			c.analyze()
			if resolved {
				if c.Error != nil || tok.binding == nil || tok.binding.Record != c.Memory["1e999"] || !tok.defined || tok.mem.Primitive.Text != "`payload 雪`" {
					t.Fatalf("candidate did not resolve to the canonical declared value: %v", c.Error)
				}
				if err := statement.unresolvedNumberError(); err != nil {
					t.Fatalf("resolved candidate remained eligible for numeric failure: %v", err)
				}
			} else {
				var terminal *strconv.NumError
				if !errors.Is(c.Error, cause) || !errors.Is(c.Error, strconv.ErrRange) || !errors.As(c.Error, &terminal) || terminal != native || !strings.Contains(c.Error.Error(), "identity.l:2:1: tokenize:") {
					t.Fatalf("terminal failure replaced the original native cause or provenance: %v", c.Error)
				}
			}
			if statement.tokens[position] != tok || tok.rangeError != cause || tok.bindingError != nil || string(tok.rawBytes) != "1e999" || statement.source.Span != span {
				t.Fatal("resolution changed token identity, numeric provenance, source span or binding-error state")
			}
			if !bytes.Equal(c.MainBytes, []byte(source)) || !bytes.Equal(c.Prepared.Original, []byte(source)) || !bytes.Equal(c.Prepared.PreparedBytes, prepared) || len(c.TextCache) != len(cache) {
				t.Fatal("resolution changed source preparation or literal-pool contents")
			}
			for key, value := range cache {
				if !bytes.Equal(c.TextCache[key], value) {
					t.Fatal("resolution changed cached literal bytes")
				}
			}
		})
	}
}
