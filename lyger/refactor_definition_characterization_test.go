package Lyger

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestRefactorDefinitionReferenceSlotPrecedence(t *testing.T) {
	for _, slot := range []string{"", "Syllable", "sYlLaBlE"} {
		for _, tc := range []struct {
			name string
			arg  DefinitionArgument
			want string
		}{
			{"unclassified definition", DefinitionArgument{Kind: "SourceSpecificKind", Definition: &DefinitionInstance{Expression: "A"}, Text: "W"}, "A"},
			{"empty kind definition", DefinitionArgument{Definition: &DefinitionInstance{Expression: "A"}}, "A"},
			{"folded syllable", DefinitionArgument{Kind: "sYlLaBlE", Text: "W"}, "W"},
			{"empty definition falls back", DefinitionArgument{Kind: "Syllable", Definition: &DefinitionInstance{}, Text: "W"}, "W"},
		} {
			t.Run(slot+"/"+tc.name, func(t *testing.T) {
				parameter := "Body"
				if slot != "" {
					parameter += "|" + slot
				}
				template := parseTestDefinition(t, `<DEF><@`+parameter+`@><DEF>`)
				instance := bindTestDefinition(t, template, map[string]DefinitionArgument{"Body": tc.arg})
				if instance.Expression != tc.want || instance.Bindings[0].Kind != tc.arg.Kind {
					t.Fatalf("reference precedence = %#v, want %s and raw kind %q", instance, tc.want, tc.arg.Kind)
				}
			})
		}
	}
}

func TestRefactorDefinitionUnknownNamesAndReferenceErrors(t *testing.T) {
	for _, tc := range []struct {
		name, slot string
		arg        DefinitionArgument
		want       string
	}{
		{"same unknown", "Foreign", DefinitionArgument{Kind: "fOrEiGn"}, `unsupported parameter type "Foreign"`},
		{"different unknown", "Foreign", DefinitionArgument{Kind: "Other"}, "requires Foreign, got Other"},
		{"argument whitespace", "Text", DefinitionArgument{Kind: " Text", Text: "value"}, "requires Text, got  Text"},
		{"reference missing", "", DefinitionArgument{Kind: "Text", Text: "A"}, "requires an instantiated syllable definition"},
		{"reference empty", "Syllable", DefinitionArgument{Kind: "Definition", Definition: &DefinitionInstance{}}, "requires an instantiated syllable definition"},
		{"explicit Definition is not reference", "Definition", DefinitionArgument{Kind: "Definition", Definition: &DefinitionInstance{Expression: "A"}}, `unsupported parameter type "Definition"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			parameter := "Value"
			if tc.slot != "" {
				parameter += "|" + tc.slot
			}
			template := parseTestDefinition(t, `<DEF><@`+parameter+`@><DEF>`)
			_, err := template.Instantiate(map[string]DefinitionArgument{"Value": tc.arg})
			if err == nil || !strings.Contains(err.Error(), tc.want) || !strings.Contains(err.Error(), `library.l:4:2: definition: parameter "Value":`) {
				t.Fatalf("error = %v, want source/slot context and %q", err, tc.want)
			}
		})
	}
}

func TestRefactorDefinitionVariadicBindingRemainsExactCase(t *testing.T) {
	for _, tc := range []struct {
		slot, kind string
		wantError  string
	}{
		{"Syllables", "Syllable", ""},
		{"syllables", "Syllable", "definition received 1 unused arguments"},
		{"sYlLaBlEs", "Syllable", "definition received 1 unused arguments"},
		{"Syllables", "syllable", "Steps requires definition syllables, got syllable"},
	} {
		t.Run(tc.slot+"/"+tc.kind, func(t *testing.T) {
			template := parseTestDefinition(t, `<DEF>W^<@Steps|`+tc.slot+`@><DEF>`)
			bindings, err := definitionArguments(template, []DefinitionArgument{{Kind: tc.kind, Text: "A"}, {Kind: tc.kind, Text: "W"}})
			if tc.wantError != "" {
				if err == nil || err.Error() != tc.wantError {
					t.Fatalf("binding error = %v, want %q", err, tc.wantError)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			want := DefinitionArgument{Kind: "Syllables", Syllables: []string{"A", "W"}}
			if !reflect.DeepEqual(bindings["Steps"], want) {
				t.Fatalf("variadic binding = %#v, want %#v", bindings["Steps"], want)
			}
		})
	}
}

func TestRefactorDefinitionFunctionBindingRemainsExactCase(t *testing.T) {
	for _, tc := range []struct{ before, after, want string }{
		{"<@Name|Text@>", "<@Name|text@>", "unsupported function slot Name: text"},
		{"<@Inputs|Props@>", "<@Inputs|props@>", "unsupported function slot Inputs: props"},
		{"<@Outputs|Texts@>", "<@Outputs|texts@>", "unsupported function slot Outputs: texts"},
	} {
		t.Run(tc.after, func(t *testing.T) {
			compilation := constructionTestCompilation(t)
			var findFunction func(*constructionNode) *constructionNode
			findFunction = func(node *constructionNode) *constructionNode {
				if node.Definition != nil && node.Definition.ID == "Function" {
					return node
				}
				for _, child := range node.Children {
					if found := findFunction(child); found != nil {
						return found
					}
				}
				return nil
			}
			node := findFunction(compilation.Semantics.Constructions[0].Root)
			if node == nil {
				t.Fatal("Basic construction lost Function")
			}
			sites, err := compilation.Semantics.Declarations[node.Definition].sGram.registrationSites()
			if err != nil || len(sites) != 1 {
				t.Fatalf("function declaration sites: %v", err)
			}
			literal, err := sites[0].literal()
			if err != nil || literal == nil || bytes.Count(literal.rawBytes, []byte(tc.before)) != 1 {
				t.Fatalf("expected one function slot %q: %v", tc.before, err)
			}
			// Mutate the isolated compilation's bound definition AST, not a
			// duplicate bootstrap or a production resource on disk.
			literal.rawBytes = bytes.Replace(literal.rawBytes, []byte(tc.before), []byte(tc.after), 1)
			_, err = compilation.lowerFunction(node, compilation.Semantics.Constructions[0].Body)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("function slot failure = %v; want lowering error %q", err, tc.want)
			}
		})
	}
}

func TestRefactorDefinitionRenderingCaseAndRawLedger(t *testing.T) {
	for _, tc := range []struct{ slot, kind string }{
		{"Text", "Text"}, {"text", "TEXT"}, {"tExT", "TeXt"}, {"Thing", "tExT"},
	} {
		t.Run(tc.slot+"/"+tc.kind, func(t *testing.T) {
			raw := `<DEF>BOD^E{|Statement: ~consume(<@Value|` + tc.slot + `@>)~|}<DEF>`
			template := parseTestDefinition(t, raw)
			text := "雪, <@Other@> and \"quoted\"\\text\n"
			argument := DefinitionArgument{Kind: tc.kind, Text: text}
			instance := bindTestDefinition(t, template, map[string]DefinitionArgument{"Value": argument})
			want := `BOD^E{|Statement: ~consume(` + strconv.Quote(text) + `)~|}`
			if instance.Expression != want {
				t.Fatalf("expression = %q, want %q", instance.Expression, want)
			}
			ledger := []DefinitionBinding{{Slot: DefinitionSlot{Name: "Value", Type: tc.slot}, Kind: tc.kind, Emitted: strconv.Quote(text)}}
			if !reflect.DeepEqual(instance.Bindings, ledger) || !bytes.Equal(template.Original, []byte(raw)) {
				t.Fatalf("raw prototype or ledger changed: %#v", instance.Bindings)
			}
			for _, value := range []interface{}{argument, instance.Bindings[0]} {
				encoded, err := json.Marshal(value)
				if err != nil {
					t.Fatal(err)
				}
				var decoded map[string]interface{}
				if err := json.Unmarshal(encoded, &decoded); err != nil {
					t.Fatal(err)
				}
				if decoded["Kind"] != tc.kind {
					t.Fatalf("public Kind lost its raw textual representation: %s", encoded)
				}
			}
		})
	}
}
