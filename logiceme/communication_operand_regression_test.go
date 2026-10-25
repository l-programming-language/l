package Logiceme

import (
	"reflect"
	"testing"
)

func TestCommunicationOperandCompatibility(t *testing.T) {
	const record = `Op{Operator: <-, Operands: &[ch, "a, b"]&}Op`
	legacy := Op{Operator: "<-", Operands: []string{"ch", `"a`, `b"`}}
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
				if key == "Op" && !reflect.DeepEqual(loaded, legacy) {
					t.Errorf("legacy Op changed: %#v", loaded)
				}
				if key == "Ops" && !reflect.DeepEqual(loaded, []Op{legacy}) {
					t.Errorf("legacy Ops changed: %#v", loaded)
				}
			}
		})
	}
	t.Run("public_loaders", func(t *testing.T) {
		if got := LoadOp(&MetaObj{Value: record}); !reflect.DeepEqual(got, legacy) {
			t.Errorf("public LoadOp changed: %#v", got)
		}
		if got := LoadOps(&MetaObj{Value: "&[" + record + "]&"}); !reflect.DeepEqual(got, []Op{legacy}) {
			t.Errorf("public LoadOps changed: %#v", got)
		}
	})
}
