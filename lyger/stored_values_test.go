package Lyger

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestStoredValueRejectsMalformedStorage(t *testing.T) {
	for _, tc := range []struct {
		name  string
		entry *entry
	}{
		{"nil entry", nil},
		{"empty entry", &entry{}},
		{"missing primitive", &entry{Type: storagePrimitive}},
		{"wrong wrapper", &entry{Type: storagePrimitive, Lexeme: &lexeme{}}},
		{"multiple wrappers", &entry{Type: storagePrimitive, Primitive: &primitive{Type: valueText}, Lexeme: &lexeme{}}},
		{"unknown storage", &entry{Type: storageUnknown, Primitive: &primitive{Type: valueText}}},
		{"unreadable primitive", &entry{Type: storagePrimitive, Primitive: &primitive{Type: valueUnknown}}},
		{"nil App", &entry{Type: storageApplicative, Applicative: &applicative{Type: valueApp}}},
		{"wrong applicative", &entry{Type: storageApplicative, Applicative: &applicative{Type: valueApp, Service: &service{}}}},
		{"nil Dictionary", &entry{Type: storageIterative, Iterative: &iterative{Type: valueDictionary}}},
		{"wrong iterative", &entry{Type: storageIterative, Iterative: &iterative{Type: valueDictionary, List: newList()}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if value, err := tc.entry.value(); err == nil || value.Data != nil {
				t.Fatalf("malformed storage returned %+v, %v", value, err)
			}
		})
	}
	var primitiveValue *primitive
	var applicativeValue *applicative
	var iterativeValue *iterative
	if _, err := primitiveValue.value(); err == nil {
		t.Error("nil primitive accepted")
	}
	if _, err := applicativeValue.value(); err == nil {
		t.Error("nil applicative accepted")
	}
	if _, err := iterativeValue.value(); err == nil {
		t.Error("nil iterative accepted")
	}
}

func TestStorageEnumNamesAndValueMappings(t *testing.T) {
	wantNames := []string{"", "Primitive", "Lexeme", "Applicative", "Iterative", "Unknown"}
	if int(storageKindCount) != len(wantNames) {
		t.Fatal("unexpected storage domain")
	}
	for i, want := range wantNames {
		if got := storageKind(i).String(); got != want {
			t.Errorf("storage %d = %q, want %q", i, got, want)
		}
	}
	for _, invalid := range []storageKind{storageKindCount, 255} {
		if !strings.HasPrefix(invalid.String(), "storageKind(") {
			t.Errorf("invalid storage displayed %q", invalid)
		}
		if _, err := json.Marshal(invalid); err == nil {
			t.Error("invalid storage serialized")
		}
	}
	want := map[valueType]string{
		valueThing: "Primitive", valueText: "Primitive", valueNumber: "Primitive", valueTruth: "Primitive", valueBytes: "Primitive", valuePointer: "Primitive",
		valueLexeme: "Lexeme", valueApp: "Applicative", valueService: "Applicative", valueClass: "Applicative", valueObject: "Applicative", valueAttribute: "Applicative", valueDefinition: "Applicative", valueFunction: "Applicative", valueProperty: "Applicative", valueOperation: "Applicative",
		valueList: "Iterative", valueDictionary: "Iterative",
	}
	for kind := valueInvalid; kind <= valueTypeCount; kind++ {
		storage, ok := kind.StorageKind()
		if ok != (want[kind] != "") || storage.String() != want[kind] {
			t.Errorf("storage for %s = %s/%v, want %q", kind, storage, ok, want[kind])
		}
	}
	if storage, ok := valueType(255).StorageKind(); ok || storage != storageInvalid {
		t.Error("out-of-range value mapped to storage")
	}
}

func TestStoredValueTextualSnapshots(t *testing.T) {
	encoded, err := json.Marshal(&entry{})
	if err != nil || !strings.Contains(string(encoded), `"Type":""`) {
		t.Fatalf("empty storage snapshot = %s, %v", encoded, err)
	}
	encoded, err = json.Marshal(&primitive{Type: valueText, Text: "`unchanged`"})
	if err != nil || string(encoded) != "{\"type\":\"Text\",\"text\":\"`unchanged`\"}" {
		t.Fatalf("primitive snapshot = %s, %v", encoded, err)
	}
}
