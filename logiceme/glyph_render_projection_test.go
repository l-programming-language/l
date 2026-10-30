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
	glyph.Meta.Add("First", "same", "String")
	glyph.Meta.Add("Second", "same", "SourceString")
	data, err := glyph.functionalArgumentData([]Parameter{
		{Key: "First", Type: "Meta", SubType: "Second"},
		{Key: "Second", Type: "Meta", SubType: "First"},
		{Key: "Values", Type: "Conjoinments", SubType: "Calculations"},
		{Key: "Absent", Type: "Conjoinments", SubType: "Declaration"},
		{Key: "Empty", Type: "Conjoinments", SubType: "Declarations"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if data.Meta.Get("First").Type != "SourceString" || data.Meta.Get("Second").Type != "String" {
		t.Fatalf("forwarded metadata lost its caller snapshot: %+v", data.Meta)
	}
	data.Meta.Metas[0].Value = "changed"
	data.Meta.Index.Keys[0] = "changed"
	if glyph.Meta.GetVal("First") != "same" || glyph.Meta.Get("First").Type != "String" {
		t.Fatal("forwarded metadata aliases its caller")
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
