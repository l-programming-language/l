package Lyger

import (
	"bytes"
	"strings"
	"testing"
)

func TestQualifiedMemberSourceShadowing(t *testing.T) {
	const outer = "Create \"Tools\" as a Dictionary.\nCreate Tools's \"Value\" as a Dictionary.\n\n"
	for _, tc := range []struct {
		name, local, spelling string
		memberLine            int
	}{
		{"declared member", "    Create \"Tools\" as a Dictionary.\nCreate Tools's \"Value\" as a Dictionary.\nUse Tools.Value as \"Copy\".\n", "Tools.Value", 5},
		{"forward member", "    Create \"Tools\" as a Dictionary.\nUse Tools.Value as \"Copy\".\nCreate Tools's \"Value\" as a Dictionary.\n", "Tools.Value", 6},
		{"forward owner", "    Use Tools.Value as \"Copy\".\nCreate \"Tools\" as a Dictionary.\nCreate Tools's \"Value\" as a Dictionary.\n", "Tools.Value", 6},
		{"possessive forward control", "    Create \"Tools\" as a Dictionary.\nUse Tools's Value as \"Copy\".\nCreate Tools's \"Value\" as a Dictionary.\n", "Value", 6},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := outer + tc.local
			c := paragraphPublicationSource(t, source)
			if c.Error != nil {
				t.Fatal(c.Error)
			}
			members := paragraphBindingDeclarations(t, c, "Value", 2)
			owners := paragraphBindingDeclarations(t, c, "Tools", 2)
			alias := paragraphBindingDeclarations(t, c, "Copy", 1)[0].record
			local := members[1].record
			canonical, err := canonicalRecord(alias, c.Semantics.Aliases)
			if err != nil || canonical != local {
				t.Fatalf("qualified alias selected outer=%t, want local canonical member: %v", canonical == members[0].record, err)
			}
			if owners[1].record.members["Tools.Value"] != local || alias.Value.Iterative.Dictionary != local.Value.Iterative.Dictionary {
				t.Fatal("local alias detached from the canonical owner's member")
			}
			if c.Memory["Tools"] != owners[0].record || c.Memory["Tools.Value"] != members[0].record || c.Memory["Copy"] != nil {
				t.Fatal("qualified local lookup changed root publication")
			}
			paragraphBindingUses(t, c, tc.spelling, []*record{local})
			declaration := c.Semantics.Declarations[canonical]
			if declaration != members[1].statement || declaration.source.Span.Start.Line != tc.memberLine || declaration.source.Span.Paragraph != 2 {
				t.Fatal("qualified resolution changed declaration provenance")
			}
			span := declaration.source.Span
			if strings.TrimSpace(source[span.Start.Offset:span.End.Offset]) != strings.TrimSpace(string(declaration.source.Raw)) || !bytes.Equal(c.MainBytes, []byte(source)) {
				t.Fatal("qualified resolution changed source bytes or span offsets")
			}
		})
	}
}

func TestQualifiedMemberSourceMissingLocalMember(t *testing.T) {
	for _, spelling := range []string{"Tools.Value", "Tools's Value"} {
		t.Run(spelling, func(t *testing.T) {
			source := "Create \"Tools\" as a Dictionary.\nCreate Tools's \"Value\" as a Dictionary.\n\n" +
				"    Create \"Tools\" as a Dictionary.\nUse " + spelling + " as \"Copy\".\n"
			c := paragraphPublicationSource(t, source)
			if c.Error == nil || !strings.Contains(c.Error.Error(), "main.l:5:1") || !strings.Contains(c.Error.Error(), "unresolved") {
				t.Fatalf("missing private member error = %v; want unresolved source line 5", c.Error)
			}
			for _, declaration := range c.Semantics.Declarations {
				if declaration.source != nil && declaration.source.Unit.Kind == SourceProgram && declaration.source.Span.Start.Line == 5 {
					t.Fatal("unresolved qualified alias was published")
				}
			}
		})
	}
}

func TestQualifiedMemberSourceAliasOwners(t *testing.T) {
	for _, tc := range []struct{ name, aliases, spelling string }{
		{"owner alias", "Use Tools as \"Alias\".\n", "Alias.Value"},
		{"owner alias chain", "Use Tools as \"Alias\".\nUse Alias as \"Second\".\n", "Second.Value"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := "    Create \"Tools\" as a Dictionary.\nCreate Tools's \"Value\" as a Dictionary.\n" + tc.aliases +
				"Use " + tc.spelling + " as \"Copy\".\n"
			c := paragraphPublicationSource(t, source)
			if c.Error != nil {
				t.Fatal(c.Error)
			}
			member := paragraphBindingDeclarations(t, c, "Value", 1)[0]
			alias := paragraphBindingDeclarations(t, c, "Copy", 1)[0].record
			canonical, err := canonicalRecord(alias, c.Semantics.Aliases)
			if err != nil || canonical != member.record || alias.Value.Iterative.Dictionary != member.record.Value.Iterative.Dictionary {
				t.Fatalf("aliased owner lost canonical member identity: %v", err)
			}
			paragraphBindingUses(t, c, tc.spelling, []*record{member.record})
			if c.Semantics.Declarations[canonical] != member.statement || member.statement.source.Span.Start.Line != 2 || len(c.Memory) != 0 {
				t.Fatal("owner alias changed provenance or published private bindings")
			}
			if alias.storageScope.records[tc.spelling] != nil {
				t.Fatal("owner alias synthesized a copied qualified member binding")
			}
		})
	}
}

func qualifiedDictionaryRecord(name string) *record {
	return &record{ID: name, Type: "iterative.dictionary", members: make(map[string]*record), Value: &entry{
		Type: storageIterative, Iterative: &iterative{Type: valueDictionary, Dictionary: newDict()},
	}}
}

func TestQualifiedMemberNestedOwnerBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name                 string
		inner, leaf, aliased bool
	}{
		{"complete nested path", true, true, false},
		{"missing intermediate", false, false, false},
		{"missing leaf", true, false, false},
		{"aliased intermediate", true, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			outerLeaf := qualifiedDictionaryRecord("Tools.Inner.Value")
			root := &lexicalScope{records: map[string]*record{outerLeaf.ID: outerLeaf}}
			owner := qualifiedDictionaryRecord("Tools")
			local := &lexicalScope{parent: root, records: map[string]*record{"Tools": owner}}
			semantics := newSemanticProgram()
			inner := qualifiedDictionaryRecord("Tools.Inner")
			if tc.inner {
				owner.members[inner.ID] = inner
			}
			if tc.aliased {
				canonical := qualifiedDictionaryRecord("CanonicalInner")
				semantics.Aliases[inner] = canonical
				inner = canonical
			}
			leaf := qualifiedDictionaryRecord(inner.ID + ".Value")
			if tc.leaf {
				inner.members[leaf.ID] = leaf
			}
			tok := &token{rawBytes: []byte("Tools.Inner.Value"), scope: local}
			binding, state, err := (semanticResolver{semantics: semantics}).lookup(tok)
			if err != nil {
				t.Fatal(err)
			}
			if tc.leaf {
				if state != lookupFound || binding == nil || binding.Record != leaf {
					t.Fatalf("nested member = %v, %s; want canonical local leaf", binding, state)
				}
			} else if state != lookupPending || binding != nil {
				t.Fatalf("missing nested member = %v, %s; must block outer fallback", binding, state)
			}
			if string(tok.rawBytes) != "Tools.Inner.Value" || root.records[outerLeaf.ID] != outerLeaf || len(local.records) != 1 {
				t.Fatal("nested lookup mutated spelling or lexical indexes")
			}
		})
	}
}

func TestQualifiedMemberCompleteBindingPriority(t *testing.T) {
	for _, tc := range []struct{ name, spelling string }{
		{"public member with invisible owner", "Hidden.Exported"},
		{"native namespace", "App.Label"},
		{"nearer complete binding", "Tools.Value"},
		{"qualified owner prefix", "Hidden.Exported.Value"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rootOwner := qualifiedDictionaryRecord(strings.Split(tc.spelling, ".")[0])
			root := &lexicalScope{records: make(map[string]*record)}
			if tc.name == "nearer complete binding" || tc.name == "qualified owner prefix" {
				root.records[rootOwner.ID] = rootOwner
			}
			local := &lexicalScope{parent: root, records: make(map[string]*record)}
			wanted := qualifiedDictionaryRecord(tc.spelling)
			if tc.name == "qualified owner prefix" {
				owner := qualifiedDictionaryRecord("Hidden.Exported")
				owner.members[wanted.ID] = wanted
				local.records[owner.ID] = owner
			} else {
				local.records[wanted.ID] = wanted
			}
			binding, state, err := (semanticResolver{semantics: newSemanticProgram()}).lookup(&token{rawBytes: []byte(tc.spelling), scope: local})
			if err != nil || state != lookupFound || binding == nil || binding.Record != wanted {
				t.Fatalf("nearer qualified lookup = %v, %s, %v; want local canonical binding", binding, state, err)
			}
		})
	}
}

func TestQualifiedMemberDoesNotSynthesizeRuntimeRead(t *testing.T) {
	app := &record{ID: "Worker", Value: &entry{Type: storageApplicative, Applicative: &applicative{Type: valueApp, App: &App{ID: "Worker"}}}}
	scope := &lexicalScope{records: map[string]*record{app.ID: app}}
	for _, spelling := range []string{"Worker.Output", "Worker.Error"} {
		t.Run(spelling, func(t *testing.T) {
			tok := &token{rawBytes: []byte(spelling), scope: scope}
			binding, state, err := (semanticResolver{semantics: newSemanticProgram()}).lookup(tok)
			if err != nil || binding != nil || state != lookupPending || tok.mem != nil || tok.binding != nil {
				t.Fatalf("static qualified lookup synthesized a runtime read: binding=%v state=%s error=%v", binding, state, err)
			}
		})
	}
}

func TestQualifiedMemberPendingClearsStaleBinding(t *testing.T) {
	for _, pendingOwner := range []bool{false, true} {
		name := "published owner without member"
		if pendingOwner {
			name = "reserved owner"
		}
		t.Run(name, func(t *testing.T) {
			outer := qualifiedDictionaryRecord("Tools.Value")
			root := &lexicalScope{records: map[string]*record{outer.ID: outer}}
			local := &lexicalScope{parent: root, records: make(map[string]*record)}
			r := semanticResolver{semantics: newSemanticProgram()}
			tok := &token{rawBytes: []byte("Tools.Value"), scope: local}
			if state, err := r.bindToken(tok); err != nil || state != lookupFound || tok.binding.Record != outer {
				t.Fatalf("initial root control failed: %s, %v", state, err)
			}
			owner := qualifiedDictionaryRecord("Tools")
			reservation := &scopeDeclaration{Token: &token{rawBytes: []byte(`"Tools"`)}, Scope: local}
			if pendingOwner {
				local.declarations = []*scopeDeclaration{reservation}
			} else {
				local.records[owner.ID] = owner
			}
			state, err := r.bindToken(tok)
			if err != nil || state != lookupPending || tok.binding != nil || tok.defined || tok.bindingError != nil || tok.tType != "primitive.unknown" || tok.mem == nil || tok.mem.Type != storagePrimitive || tok.mem.Primitive.Type != valueUnknown {
				t.Fatalf("pending local owner retained a stale root binding: state=%s token=%+v error=%v", state, tok, err)
			}
			leaf := qualifiedDictionaryRecord("Tools.Value")
			owner.members[leaf.ID], local.records[owner.ID], reservation.Record = leaf, owner, owner
			if state, err = r.bindToken(tok); err != nil || state != lookupFound || tok.binding == nil || tok.binding.Record != leaf || tok.mem.Iterative.Dictionary != leaf.Value.Iterative.Dictionary || string(tok.rawBytes) != "Tools.Value" {
				t.Fatalf("retry lost canonical local member or source spelling: %s, %v", state, err)
			}
			if outer.Value.Iterative.Dictionary == leaf.Value.Iterative.Dictionary || root.records[outer.ID] != outer {
				t.Fatal("retry mutated the root member")
			}
		})
	}
}

func TestQualifiedMemberMorphologyAndAmbiguity(t *testing.T) {
	for _, ambiguous := range []bool{false, true} {
		name := "case and possessive morphology"
		if ambiguous {
			name = "stable member ambiguity"
		}
		t.Run(name, func(t *testing.T) {
			lex := newLexeme("verb")
			if err := lex.reLabel(`"Write|-s"`); err != nil {
				t.Fatal(err)
			}
			lex.Inputs = []string{"original"}
			member := &record{ID: "Library.Write", Value: &entry{Type: storageLexeme, Lexeme: lex}}
			owner := qualifiedDictionaryRecord("Library")
			owner.members["a"] = member
			if ambiguous {
				owner.members["z"] = &record{ID: "Library.Other", Value: &entry{Type: storageLexeme, Lexeme: lex}}
			}
			scope := &lexicalScope{records: map[string]*record{owner.ID: owner}}
			tok := &token{rawBytes: []byte("LiBrArY.Writes’"), scope: scope}
			state, err := (semanticResolver{semantics: newSemanticProgram()}).bindToken(tok)
			if ambiguous {
				if err == nil || !strings.Contains(err.Error(), "Library.Write and Library.Other") || tok.binding != nil || tok.bindingError != err {
					t.Fatalf("qualified ambiguity lost stable alternatives: %v", err)
				}
				return
			}
			if err != nil || state != lookupFound || tok.binding == nil || tok.binding.Record != member || tok.binding.Form != "active" || !tok.binding.Possessive || tok.tType != "verbs.generic.active.possessive" {
				t.Fatalf("qualified member lost morphology: %s, %v, %+v", state, err, tok.binding)
			}
			tok.mem.Lexeme.Inputs[0] = "changed"
			if lex.Inputs[0] != "original" || string(tok.rawBytes) != "LiBrArY.Writes’" {
				t.Fatal("qualified binding changed its stored definition or source spelling")
			}
		})
	}
}

func TestQualifiedMemberAliasOwnerErrors(t *testing.T) {
	for _, cycle := range []bool{false, true} {
		name, want := "missing target", "alias Alias has no target"
		if cycle {
			name, want = "cycle", "alias cycle at Alias"
		}
		t.Run(name, func(t *testing.T) {
			alias := qualifiedDictionaryRecord("Alias")
			semantics := newSemanticProgram()
			semantics.Aliases[alias] = nil
			if cycle {
				other := qualifiedDictionaryRecord("Other")
				semantics.Aliases[alias], semantics.Aliases[other] = other, alias
			}
			tok := &token{rawBytes: []byte("Alias.Value"), scope: &lexicalScope{records: map[string]*record{alias.ID: alias}}}
			binding, _, err := (semanticResolver{semantics: semantics}).lookup(tok)
			if err == nil || err.Error() != want || binding != nil {
				t.Fatalf("qualified alias error = %v; want %q", err, want)
			}
		})
	}
}
