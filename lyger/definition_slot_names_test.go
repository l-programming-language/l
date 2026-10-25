package Lyger

import (
	"reflect"
	"testing"
)

func TestDefinitionSlotNamesUseLIdentifierSyntax(t *testing.T) {
	for _, name := range []string{"go", "func", "range", "type", "interface", "map", "package", "defer", "go.变量_2", "δ.return", "fn", "let", "match", "_", "变量.δ_2"} {
		t.Run(name, func(t *testing.T) {
			raw := []byte("<DEF>E{|Statement: ~<@" + name + "|Text@>~|}<DEF>")
			template, err := ParseDefinitionTemplate(raw, SourceSpan{})
			if err != nil {
				t.Fatalf("L slot label %q was rejected: %v", name, err)
			}
			want := DefinitionSlot{Name: name, Type: "Text"}
			if !reflect.DeepEqual(template.Slots, []DefinitionSlot{want}) {
				t.Fatalf("slot identity changed: %#v", template.Slots)
			}
			instance, err := template.Instantiate(map[string]DefinitionArgument{name: {Kind: "Text", Text: "label"}})
			if err != nil || instance.Expression != `E{|Statement: ~"label"~|}` || len(instance.Bindings) != 1 || instance.Bindings[0].Slot != want {
				t.Fatalf("L slot did not retain its binding: %#v, %v", instance, err)
			}
		})
	}
	for _, name := range []string{"go", "range.类型_2", "δ.return"} {
		template, err := ParseDefinitionTemplate([]byte("<DEF>E{|Statement: ~<@Value|"+name+"@>~|}<DEF>"), SourceSpan{})
		if err != nil {
			t.Errorf("source type label %q was rejected before argument interpretation: %v", name, err)
			continue
		}
		if !reflect.DeepEqual(template.Slots, []DefinitionSlot{{Name: "Value", Type: name}}) {
			t.Errorf("source type label changed: %#v", template.Slots)
		}
		if _, err := template.Instantiate(map[string]DefinitionArgument{"Value": {Kind: name, Text: "label"}}); err == nil {
			t.Errorf("an unknown source type %q became renderable", name)
		}
	}
	for _, name := range []string{"", ".value", "value.", "a..b", "2a", "a.2b", "r#type", "a-b", "a b", "a\u0301"} {
		for _, slot := range []string{name + "|Text", "Value|" + name} {
			if _, err := ParseDefinitionTemplate([]byte("<DEF>E{|Statement: ~<@"+slot+"@>~|}<DEF>"), SourceSpan{}); err == nil {
				t.Errorf("invalid dotted identifier accepted in slot %q", slot)
			}
		}
	}
}
