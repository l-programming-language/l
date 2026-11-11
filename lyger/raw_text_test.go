package Lyger

import (
	"reflect"
	"testing"
)

func TestDefinitionRawTextLiteral(t *testing.T) {
	for _, tc := range []struct{ name, value, literal string }{
		{"invalid octets", string([]byte{'A', 0xff, 'B', 0x80, 'C'}), `b"\x41\xff\x42\x80\x43"`},
		{"mixed Unicode and markup", "雪" + string([]byte{0xff, '|', '~', 0, '"', '\\'}), `b"\xe9\x9b\xaa\xff\x7c\x7e\x00\x22\x5c"`},
		{"incomplete sequence", string([]byte{0xf0, 0x9f}), `b"\xf0\x9f"`},
		{"valid Unicode unchanged", "雪|~", `"雪\x7c\x7e"`},
	} {
		for _, slot := range []string{"Text", "Thing"} {
			t.Run(tc.name+"/"+slot, func(t *testing.T) {
				template := parseTestDefinition(t, `<DEF>E{|Statement: ~consume(<@Value|`+slot+`@>)~|}<DEF>`)
				bindings := map[string]DefinitionArgument{"Value": {Kind: "Text", Text: tc.value}}
				first := bindTestDefinition(t, template, bindings)
				want := `E{|Statement: ~consume(` + tc.literal + `)~|}`
				if first.Expression != want || first.OriginalExpression != want || len(first.Bindings) != 1 || first.Bindings[0].Emitted != tc.literal {
					t.Fatalf("native Text bytes changed: %#v; expected %s", first, want)
				}
				second := bindTestDefinition(t, template, bindings)
				if !reflect.DeepEqual(first, second) || bindings["Value"].Text != tc.value {
					t.Fatal("native Text rendering changed caller data or repeated output")
				}
			})
		}
	}
}
