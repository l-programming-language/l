package Logiceme

import (
	"reflect"
	"testing"
)

func TestCommunicationOperandParsing(t *testing.T) {
	const record = `Op{Operator: <-, Operands: &[ch, "a, b"]&}Op`
	expected := Op{Operator: "<-", Operands: []string{"ch", `"a`, `b"`}}
	for _, system := range []*SegmentalSystem{nil, {}, {Language: "Rust"}} {
		name := "nil"
		if system != nil {
			name = system.Language
		}
		t.Run(name, func(t *testing.T) {
			interpreter := LucidInterpreter{SegmentalSystem: system}
			interpreter.Meta.SetAdd("Op", record, "Struct")
			interpreter.Meta.SetAdd("Ops", "&["+record+"]&", "Array")
			for _, key := range []string{"Op", "Ops"} {
				loaded, _ := interpreter.metaSwitch(key)
				if key == "Op" && !reflect.DeepEqual(loaded, expected) {
					t.Errorf("Op fields differ: %#v", loaded)
				}
				if key == "Ops" && !reflect.DeepEqual(loaded, []Op{expected}) {
					t.Errorf("Ops fields differ: %#v", loaded)
				}
			}
		})
	}
	t.Run("public_loaders", func(t *testing.T) {
		if got := LoadOp(&MetaObj{Value: record}); !reflect.DeepEqual(got, expected) {
			t.Errorf("public LoadOp changed: %#v", got)
		}
		if got := LoadOps(&MetaObj{Value: "&[" + record + "]&"}); !reflect.DeepEqual(got, []Op{expected}) {
			t.Errorf("public LoadOps changed: %#v", got)
		}
	})
}
