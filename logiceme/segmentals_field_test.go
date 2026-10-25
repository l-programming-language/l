package Logiceme

import (
	"reflect"
	"testing"
)

func TestMetaStructFieldSwitchRetainsFallbacks(t *testing.T) {
	for _, tc := range []struct {
		name, field string
		value       interface{}
	}{
		{"Op", "Prefix", Op{Prefix: "known"}},
		{"Prop", "Key", Prop{Key: "known"}},
		{"Func", "Key", Func{Key: "known"}},
		{"Recur", "Key", Recur{Key: "known"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, representation := range []struct {
				name  string
				value interface{}
				want  string
			}{
				{"direct", tc.value, "known"},
				{"reflected", reflect.ValueOf(tc.value), "known"},
				{"mismatched direct", 42, ""},
				{"mismatched reflected", reflect.ValueOf(42), "Error"},
			} {
				t.Run(representation.name, func(t *testing.T) {
					got, kind := MetaStructFieldSwitch(representation.value, tc.name, tc.field)
					if got != representation.want || kind != "String" {
						t.Fatalf("field projection = (%v, %q), want (%q, String)", got, kind, representation.want)
					}
				})
			}
		})
	}
}
