package Logiceme

import (
	"reflect"
	"testing"
)

func TestForwardedChildProjectionPreservesGroups(t *testing.T) {
	glyph := Glyph{Conjoinments: []Glyph{
		{Type: "Calculation", Content: []byte("first")},
		{Type: "Operation", Content: []byte("middle")},
		{Type: "Calculation", Content: []byte("last")},
	}}
	for _, want := range glyph.compiledChildGroups() {
		got, found := glyph.compiledChildGroup(want.Key)
		if !found || !reflect.DeepEqual(got, want) {
			t.Fatalf("%s = %+v, found %v; want %+v", want.Key, got, found, want)
		}
	}
	if _, found := glyph.compiledChildGroup("Unknowns"); found {
		t.Fatal("unknown child category was accepted")
	}
	data, err := glyph.functionalArgumentData([]Parameter{
		{Key: "Values", Type: "Conjoinments", SubType: "Calculations"},
		{Key: "Absent", Type: "Conjoinments", SubType: "Declaration"},
		{Key: "Empty", Type: "Conjoinments", SubType: "Declarations"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(data.Conjoinments) != 3 || data.Conjoinments[0].Key != "Values" ||
		data.Conjoinments[0].Type != "Calculations" || !data.Conjoinments[0].Plural ||
		data.Conjoinments[1].Plural || !data.Conjoinments[2].Plural {
		t.Fatalf("forwarded categories changed: %+v", data.Conjoinments)
	}
	data.Conjoinments[0].Contents[0] = []byte("replacement")
	if string(glyph.Conjoinments[0].Content) != "first" {
		t.Fatal("projection retained mutable child slice storage")
	}
}
