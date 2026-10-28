package Lyger

import (
	"encoding/json"
	"testing"
)

func TestRefactorIRTextualDiscriminators(t *testing.T) {
	for _, tc := range []struct {
		name   string
		value  any
		fields map[string]string
	}{
		{"empty instruction", instruction{}, map[string]string{"Kind": ""}},
		{"instruction", instruction{Kind: instructionShow}, map[string]string{"Kind": "show"}},
		{"empty operand", instructionValue{}, map[string]string{"Kind": "", "Type": "", "Member": ""}},
		{"native operand", instructionValue{Kind: operandMember, Type: valueText, Member: memberOutput}, map[string]string{"Kind": "member", "Type": "Text", "Member": "output"}},
		{"scope", pronounBinding{Kind: scopeActionArgument}, map[string]string{"Kind": "action-argument"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := json.Marshal(tc.value)
			if err != nil {
				t.Fatal(err)
			}
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(encoded, &fields); err != nil {
				t.Fatal(err)
			}
			for name, want := range tc.fields {
				var got string
				if err := json.Unmarshal(fields[name], &got); err != nil || got != want {
					t.Fatalf("%s = %s (%v), want string %q", name, fields[name], err, want)
				}
			}
		})
	}
}

func TestRefactorNativeMembersFollowRoleNotSpelling(t *testing.T) {
	for _, tc := range []struct{ owner, role, name, kind string }{
		{"App", "functional.deflective", "output", "Text"},
		{"App", "functional.error", "error", "Error"},
		{"App", "functional.custom", "", ""},
		{"App", "action.executive", "", ""},
		{"Dictionary", "functional.deflective", "", ""},
		{"", "functional.error", "", ""},
	} {
		t.Run(tc.owner+"/"+tc.role, func(t *testing.T) {
			owner, _ := parseValueType(tc.owner)
			spec, ok := nativeMemberFor(owner, &lexeme{Root: "renamed", Type: tc.role})
			kind, _ := spec.ValueType()
			if ok != (tc.name != "") || spec.String() != tc.name || kind.String() != tc.kind {
				t.Fatalf("member = %+v, %v; want %q/%q", spec, ok, tc.name, tc.kind)
			}
		})
	}
}
