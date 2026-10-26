package Lyger

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestIREnumNamesAndInvalidValues(t *testing.T) {
	for _, tc := range []struct {
		name  string
		count int
		want  []string
		get   func(uint8) fmt.Stringer
	}{
		{"instructionKind", int(instructionKindCount), []string{"", "assign", "call", "build", "run", "show", "exit", "if", "block", "await", "stop", "bind-action", "alias-action", "function", "invoke", "return", "update"}, func(n uint8) fmt.Stringer { return instructionKind(n) }},
		{"operandKind", int(operandKindCount), []string{"", "symbol", "literal", "member", "error", "construct", "read"}, func(n uint8) fmt.Stringer { return operandKind(n) }},
		{"nativeMember", int(nativeMemberCount), []string{"", "output", "error"}, func(n uint8) fmt.Stringer { return nativeMember(n) }},
		{"pronounScopeKind", int(pronounScopeKindCount), []string{"", "paragraph", "action-argument"}, func(n uint8) fmt.Stringer { return pronounScopeKind(n) }},
		{"nativeRole", int(nativeRoleCount), []string{"", "action.generative", "action.executive", "action.observative", "action.terminative", "action.definitive", "functional.deflective", "functional.error", "action.suspensive", "action.interruptive", "action.declarative", "action.expansive", "action.locative", "action.invocative", "action.returning"}, func(n uint8) fmt.Stringer { return nativeRole(n) }},
		{"valueType", int(valueTypeCount), []string{"", "Thing", "Text", "Number", "Truth", "Bytes", "Pointer", "Lexeme", "App", "Service", "Class", "Object", "Attribute", "Definition", "Function", "Property", "Operation", "List", "Dictionary", "Error", "ID", "None", "Unknown", "Null", "Map", "Instance"}, func(n uint8) fmt.Stringer { return valueType(n) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.count != len(tc.want) {
				t.Fatalf("variant count = %d, want %d", tc.count, len(tc.want))
			}
			for n, want := range tc.want {
				if got := tc.get(uint8(n)).String(); got != want {
					t.Errorf("variant %d = %q, want %q", n, got, want)
				}
			}
			for _, n := range []uint8{uint8(tc.count), 255} {
				want := fmt.Sprintf("%s(%d)", tc.name, n)
				if got := tc.get(n).String(); got != want {
					t.Errorf("invalid %d = %q, want %q", n, got, want)
				}
			}
		})
	}
	for _, invalid := range []any{instructionKind(255), operandKind(255), nativeMember(255), pronounScopeKind(255), valueType(255)} {
		if _, err := json.Marshal(invalid); err == nil {
			t.Errorf("invalid %T encoded without error", invalid)
		}
	}
}

func TestNativeRoleRelationships(t *testing.T) {
	for _, tc := range []struct{ raw, instruction, member, result string }{
		{"action.generative", "build", "", ""},
		{"action.executive", "run", "", ""},
		{"action.suspensive", "await", "", ""},
		{"action.interruptive", "stop", "", ""},
		{"action.observative", "show", "", ""},
		{"action.terminative", "exit", "", ""},
		{"action.definitive", "assign", "", ""},
		{"action.declarative", "assign", "", ""},
		{"action.expansive", "update", "", ""},
		{"action.locative", "update", "", ""},
		{"action.invocative", "invoke", "", ""},
		{"action.returning", "return", "", ""},
		{"functional.deflective", "", "output", "Text"},
		{"functional.error", "", "error", "Error"},
	} {
		t.Run(tc.raw, func(t *testing.T) {
			role, ok := parseNativeRole(tc.raw)
			if !ok || role.String() != tc.raw {
				t.Fatalf("role = %s, %v", role, ok)
			}
			instruction, instructionOK := role.InstructionKind()
			member, memberOK := role.Member()
			kind, kindOK := member.ValueType()
			if instruction.String() != tc.instruction || instructionOK != (tc.instruction != "") || member.String() != tc.member || memberOK != (tc.member != "") || kind.String() != tc.result || kindOK != memberOK {
				t.Fatalf("mappings = %s/%v, %s/%v, %s/%v", instruction, instructionOK, member, memberOK, kind, kindOK)
			}
		})
	}
	for _, raw := range []string{"", "ACTION.EXECUTIVE", " action.executive", "action.custom"} {
		if role, ok := parseNativeRole(raw); ok || role != roleInvalid {
			t.Errorf("unknown role %q = %s, %v", raw, role, ok)
		}
	}
	for _, role := range []nativeRole{roleInvalid, nativeRoleCount, 255} {
		if kind, ok := role.InstructionKind(); ok || kind != instructionInvalid {
			t.Errorf("invalid role produced instruction %s", kind)
		}
		if member, ok := role.Member(); ok || member != memberInvalid {
			t.Errorf("invalid role produced member %s", member)
		}
	}
}

func TestValueTypeNamesAndScalarBoundary(t *testing.T) {
	for kind := valueInvalid; kind < valueTypeCount; kind++ {
		parsed, ok := parseValueType(kind.String())
		if !ok || parsed != kind {
			t.Errorf("cannot parse known value %q", kind.String())
		}
		want := map[string]string{"Text": "Vec<u8>", "Number": "f64", "Truth": "bool", "Bytes": "Vec<u8>", "Null": "()"}[kind.String()]
		scope := &lexicalScope{kind: lexicalParagraph}
		rec := &record{ID: "value", storageScope: scope, Value: &entry{Type: storagePrimitive, Primitive: &primitive{Type: kind}}}
		semantics := newSemanticProgram()
		semantics.Instructions = []*instruction{{Kind: instructionBlock, Block: &scopedInstructions{Scope: scope, Body: []*instruction{{
			Kind: instructionAssign, Target: &instructionValue{Kind: operandSymbol, Record: rec},
			Value: &instructionValue{Kind: operandLiteral, Literal: rec.Value.Primitive},
		}}}}}
		plan, err := planOrchestration(semantics)
		if want == "" {
			if err == nil || plan != nil {
				t.Errorf("nonscalar %s produced native storage %+v", kind, plan)
			}
			continue
		}
		if err != nil {
			t.Fatalf("Rust storage for %s: %v", kind, err)
		}
		if len(plan.Storage) != 1 || len(plan.Root.Body[0].Block.Storage) != 1 || plan.Storage[0] != plan.Root.Body[0].Block.Storage[0] {
			t.Fatalf("Rust storage for %s did not retain one paragraph slot: %+v", kind, plan)
		}
		slot := plan.Storage[0]
		_, declaration, err := orchestrationStorageDeclaration(slot)
		if err != nil || slot.Type != kind || slot.Initial != nil || slot.Published || !strings.Contains(declaration, ": "+want) {
			t.Errorf("Rust storage for %s = %q/%v; want explicit uninitialized %s", kind, declaration, err, want)
		}
	}
	for _, kind := range []valueType{valueTypeCount, 255} {
		rec := &record{ID: "invalid", Value: &entry{Type: storagePrimitive, Primitive: &primitive{Type: kind}}}
		semantics := newSemanticProgram()
		semantics.Instructions = []*instruction{{Kind: instructionAssign, Target: &instructionValue{Kind: operandSymbol, Record: rec}, Value: &instructionValue{Kind: operandLiteral, Literal: rec.Value.Primitive}}}
		if plan, err := planOrchestration(semantics); err == nil || plan != nil {
			t.Errorf("invalid type %d produced Rust storage %+v/%v", kind, plan, err)
		}
	}
	for _, raw := range []string{"text", " Text", "Error ", "Invented"} {
		if kind, ok := parseValueType(raw); ok || kind != valueInvalid {
			t.Errorf("unknown value %q = %s, %v", raw, kind, ok)
		}
	}
}
