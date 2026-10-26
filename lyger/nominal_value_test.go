package Lyger

import "testing"

func TestNominalInspectionChecksUnresolvedStorage(t *testing.T) {
	for _, kind := range []valueType{valueID, valueNone, valueUnknown} {
		t.Run(kind.String(), func(t *testing.T) {
			noun := &token{rawBytes: []byte("Future"), mem: &entry{Type: storagePrimitive, Primitive: &primitive{Type: kind}}}
			if name, err := nominalType(noun); err != nil || name != "future" {
				t.Fatalf("unresolved nominal = %q, %v", name, err)
			}
			if template, err := definitionLiteral([]*token{noun}); err != nil || template != nil {
				t.Fatalf("unresolved template = %v, %v", template, err)
			}
			noun.mem.Lexeme = &lexeme{}
			if name, err := nominalType(noun); err == nil {
				t.Errorf("malformed unresolved nominal = %q", name)
			}
			if template, err := definitionLiteral([]*token{noun}); err == nil {
				t.Errorf("malformed unresolved template = %v", template)
			}
		})
	}
}
