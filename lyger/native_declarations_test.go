package Lyger

import "testing"

func TestNativeDeclarationNameIsMetadata(t *testing.T) {
	for _, test := range []struct{ source, name string }{
		{`HY{|Key: ~Value~, Props: ~&[Prop{Key: text, Type: String}Prop]&~|}`, "Value"},
		{`HY{|GenericTypes: ~T, E~, Key: ~Pair~, Props: ~&[]&~|}`, "Pair"},
		{`KY{|Key: ~Value~|}`, ""},
		{`E{|Statement: ~struct Value;~|}`, ""},
	} {
		name, err := nativeDeclarationName(test.source)
		if err != nil || name != test.name {
			t.Errorf("declaration %s: %q, %v", test.source, name, err)
		}
	}
	for _, invalid := range []string{
		`HY{|Props: ~&[]&~|}`,
		`HY{|Key: ~Value~, Key: ~Other~|}`,
		`HY{|Key: ~Value; fn injected() {}~|}`,
		`HY{|Key: ~Value~|}*E{|Statement: ~fn injected() {}~|}`,
	} {
		if _, err := nativeDeclarationName(invalid); err == nil {
			t.Errorf("invalid declaration accepted: %s", invalid)
		}
	}
}
