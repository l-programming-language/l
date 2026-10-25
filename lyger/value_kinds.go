package Lyger

import (
	"strconv"

	"github.com/l-programming-language/l/lerrors"
)

// A value tag identifies a representation, not the set of accepted source type
// names. ID/None/Unknown/Null can describe states without a readable payload.
type valueType uint8

const (
	valueInvalid valueType = iota
	valueThing
	valueText
	valueNumber
	valueTruth
	valueBytes
	valuePointer
	valueLexeme
	valueApp
	valueService
	valueClass
	valueObject
	valueAttribute
	valueDefinition
	valueFunction
	valueProperty
	valueOperation
	valueList
	valueDictionary
	valueError
	valueID
	valueNone
	valueUnknown
	valueNull
	valueTypeCount
)

var valueTypeNames = [valueTypeCount]string{
	valueThing: "Thing", valueText: "Text", valueNumber: "Number", valueTruth: "Truth",
	valueBytes: "Bytes", valuePointer: "Pointer", valueLexeme: "Lexeme", valueApp: "App",
	valueService: "Service", valueClass: "Class", valueObject: "Object",
	valueAttribute: "Attribute", valueDefinition: "Definition", valueFunction: "Function",
	valueProperty: "Property", valueOperation: "Operation", valueList: "List", valueDictionary: "Dictionary",
	valueError: "Error", valueID: "ID", valueNone: "None", valueUnknown: "Unknown", valueNull: "Null",
}

func (k valueType) String() string {
	if k >= valueTypeCount {
		return "valueType(" + strconv.Itoa(int(k)) + ")"
	}
	return valueTypeNames[k]
}

func (k valueType) MarshalText() ([]byte, error) {
	if k >= valueTypeCount {
		return nil, lerrors.Internal.Errorf("invalid value type %d", uint8(k))
	}
	return []byte(valueTypeNames[k]), nil
}

func parseValueType(raw string) (valueType, bool) {
	for v := valueInvalid; v < valueTypeCount; v++ {
		if valueTypeNames[v] == raw {
			return v, true
		}
	}
	return valueInvalid, false
}

// Storage is independent of lexical role and deferred runtime provenance.
type storageKind uint8

const (
	storageInvalid storageKind = iota
	storagePrimitive
	storageLexeme
	storageApplicative
	storageIterative
	storageUnknown
	storageKindCount
)

var storageKindNames = [storageKindCount]string{
	storagePrimitive: "Primitive", storageLexeme: "Lexeme",
	storageApplicative: "Applicative", storageIterative: "Iterative", storageUnknown: "Unknown",
}

func (k storageKind) String() string {
	if k >= storageKindCount {
		return "storageKind(" + strconv.Itoa(int(k)) + ")"
	}
	return storageKindNames[k]
}

func (k storageKind) MarshalText() ([]byte, error) {
	if k >= storageKindCount {
		return nil, lerrors.Internal.Errorf("invalid storage kind %d", uint8(k))
	}
	return []byte(storageKindNames[k]), nil
}

func (k valueType) StorageKind() (storageKind, bool) {
	switch k {
	case valueThing, valueText, valueNumber, valueTruth, valueBytes, valuePointer:
		return storagePrimitive, true
	case valueLexeme:
		return storageLexeme, true
	case valueApp, valueService, valueClass, valueObject, valueAttribute,
		valueDefinition, valueFunction, valueProperty, valueOperation:
		return storageApplicative, true
	case valueList, valueDictionary:
		return storageIterative, true
	default:
		return storageInvalid, false
	}
}
