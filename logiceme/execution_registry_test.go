package Logiceme

import (
	"reflect"
	"strings"
	"testing"
)

func TestExecutionRegistry(t *testing.T) {
	for _, tc := range []struct {
		symbol, role, name string
	}{
		{"C", "Consonant", "Execute"},
		{"V", "Vowel", "Execution"},
	} {
		t.Run(tc.symbol+"/identity", func(t *testing.T) {
			defer executionRegistryPanicAsFailure(t)
			want := []string{tc.role, tc.symbol, tc.name}
			if got := DetermineSegmentType(tc.symbol); !reflect.DeepEqual(got, want) {
				t.Errorf("DetermineSegmentType(%q) = %q; want %q", tc.symbol, got, want)
			}
			segment := Segment{Symbol: tc.symbol}
			if got := segment.getSegmentName(); got != tc.name {
				t.Errorf("segment %s name = %q; want %q", tc.symbol, got, tc.name)
			}
		})
		t.Run(tc.symbol+"/storage", func(t *testing.T) {
			defer executionRegistryPanicAsFailure(t)
			var system SegmentalSystem
			want := Segment{
				Symbol: tc.symbol, Name: strings.ToLower(tc.name), Type: "logiceme",
				SubType: strings.ToLower(tc.role), Content: []byte("distinct " + tc.name + " content"),
				Inputs: []string{"Type"},
			}
			if err := system.AddLogiceme(want, false, nil); err != nil {
				t.Fatal(err)
			}
			if got := system.SegmentSwitch(tc.symbol); !reflect.DeepEqual(got, want) {
				t.Fatalf("stored segment %s = %#v; want %#v", tc.symbol, got, want)
			}
		})
	}

	for _, name := range []string{"Execute", "execute"} {
		t.Run("suffix/"+name, func(t *testing.T) {
			defer executionRegistryPanicAsFailure(t)
			if got := segmentConsonantSwitch(name); got != "Execution" {
				t.Fatalf("segmentConsonantSwitch(%q) = %q; want Execution", name, got)
			}
		})
	}
	for _, tc := range []struct {
		name, category string
	}{
		{"Execution", "Item"},
		{"Executions", "Array"},
	} {
		t.Run("return_category/"+tc.name, func(t *testing.T) {
			defer executionRegistryPanicAsFailure(t)
			var system SegmentalSystem
			if got := system.SegmentalDataSwitch(tc.name); got != "Execution" {
				t.Errorf("SegmentalDataSwitch(%q) = %q; want Execution", tc.name, got)
			}
			want := []string{tc.category, "Conjoinment.Execution"}
			if got := checkVarReturn(tc.name); !reflect.DeepEqual(got, want) {
				t.Errorf("checkVarReturn(%q) = %q; want %q", tc.name, got, want)
			}
		})
	}
	for _, expression := range []string{
		`[|Execute(Key = Meta.Key, Deferred = Meta.Deferred)|]`,
		`[|Execute("Instantiation", Instantiation = Conjoinments.Instantiation, Deferred = Meta.Deferred)|]`,
	} {
		t.Run("expression/"+expression, func(t *testing.T) {
			defer executionRegistryPanicAsFailure(t)
			if got := determineExpressionOutput([]string{"", "", expression}); got != "Execute" {
				t.Fatalf("expression output = %q; want Execute for %s", got, expression)
			}
		})
	}
}

func TestExecutionConjoinmentLookup(t *testing.T) {
	// Keep distinct singular and plural entries, with an unrelated collection, so
	// lookup must select the requested category and preserve its source order.
	newInterpreter := func() LucidInterpreter {
		return LucidInterpreter{SegmentalSystem: &SegmentalSystem{}, Conjoinments: []CompiledConjoinment{
			{Type: "Operation", Contents: [][]byte{[]byte("unrelated body")}},
			{Type: "Execution", Contents: [][]byte{[]byte("go single()")}},
			{Type: "Executions", Plural: true, Contents: [][]byte{[]byte("go first()"), []byte("defer second()")}},
		}}
	}
	for _, tc := range []struct {
		input, category string
		want            []string
	}{
		{"Execution", "Execution", []string{"go single()"}},
		{"Executions", "Executions", []string{"go first()", "defer second()"}},
		{"Executions[0]", "Execution", []string{"go first()"}},
		{"Executions[1]", "Execution", []string{"defer second()"}},
		{"Executions[2]", "Void", []string{"Void"}},
	} {
		t.Run(tc.input, func(t *testing.T) {
			defer executionRegistryPanicAsFailure(t)
			interpreter := newInterpreter()
			got, category := interpreter.conjoinmentsSwitch(tc.input)
			if category != tc.category || !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("lookup %s = (%q, %q); want (%q, %q)", tc.input, got, category, tc.want, tc.category)
			}
		})
	}
	t.Run("collection_across_entries", func(t *testing.T) {
		defer executionRegistryPanicAsFailure(t)
		conjoinments := []CompiledConjoinment{
			{Type: "Executions", Plural: true, Contents: [][]byte{[]byte("go first()"), []byte("defer second()")}},
			{Type: "Executions", Plural: true, Contents: [][]byte{[]byte("go third()")}},
		}
		got, category := LoadSegmentalConjoinments("Execution", conjoinments)
		want := []string{"go first()", "defer second()", "go third()"}
		if category != "Executions" || !reflect.DeepEqual(got, want) {
			t.Fatalf("collected executions = (%q, %q); want (%q, Executions)", got, category, want)
		}
	})
	for _, tc := range []struct{ input, want string }{
		{"[|Conjoinments.Execution|]", "go single()"},
		{"[|Conjoinments.Executions[0]|]", "go first()"},
		{"[|Conjoinments.Executions[Add(Count, -1)]|]", "defer second()"},
	} {
		t.Run("interpolation/"+tc.input, func(t *testing.T) {
			defer executionRegistryPanicAsFailure(t)
			interpreter := newInterpreter()
			interpreter.createVariable("Count", "Number", reflect.ValueOf(2))
			if got := interpreter.interpretExpressionVars([]string{tc.input}, tc.input); got != tc.want {
				t.Fatalf("%s = %q; want %q", tc.input, got, tc.want)
			}
		})
	}
}
