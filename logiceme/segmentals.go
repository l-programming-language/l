// Author: Connor Sanders

package Logiceme

import (
	"reflect"
	"strings"
)

type Op struct {
	Prefix   string   `json:"prefix,omitempty"`
	Type     string   `json:"type,omitempty"`
	Operands []string `json:"operands,omitempty"`
	Operator string   `json:"operator,omitempty"`
	Postfix  string   `json:"postfix,omitempty"`
}

// LoadOp decodes a native operation record without interpreting operand syntax.
func LoadOp(meta *MetaObj) Op {
	return opFromFields(parseStructFields(meta.Value))
}

func opFromFields(fields [][]string) Op {
	reOp := Op{}
	for _, pOp := range fields {
		value := pOp[1]
		if pOp[0] == "Operands" {
			if strings.HasPrefix(value, "&[") && strings.HasSuffix(value, "]&") {
				value = cleanVal(value)
			} else if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
				value = value[1 : len(value)-1]
			}
			reOp.Operands = strings.Split(value, ", ")
			continue
		}
		value = cleanVal(value)
		if pOp[0] == "Prefix" {
			reOp.Prefix = value
		} else if pOp[0] == "Type" {
			reOp.Type = value
		} else if pOp[0] == "Operator" {
			reOp.Operator = value
		} else if pOp[0] == "Postfix" {
			reOp.Postfix = value
		}
	}
	return reOp
}

// LoadOps decodes the ordered native operation records.
func LoadOps(meta *MetaObj) []Op { return loadMetaRecords(meta, LoadOp) }

func loadMetaRecords[T any](meta *MetaObj, load func(*MetaObj) T) []T {
	result := []T{}
	for _, item := range ParseStructArray(meta.Value) {
		result = append(result, load(item))
	}
	return result
}

func (o *Op) FieldSwitch(fieldStr string) (interface{}, string) {
	switch fieldStr {
	case "Prefix":
		return o.Prefix, "String"
	case "Type":
		return o.Type, "String"
	case "Operands":
		return o.Operands, "Array-String"
	case "Operator":
		return o.Operator, "String"
	case "Postfix":
		return o.Postfix, "String"
	default:
		return "Invalid", "String"
	}
}

type Prop struct {
	Key        string `json:"key,omitempty"`
	Type       string `json:"type,omitempty"`
	Value      string `json:"value,omitempty"`
	StructTags string `json:"struct_tags,omitempty"`
}

func (pr *Prop) IsVoid() bool { return *pr == (Prop{}) }

func StringifyProp(prop Prop) string {
	return prop.Key + " " + prop.Type
}

func StringifyProps(props []Prop) string {
	values := make([]string, len(props))
	for i, prop := range props {
		values[i] = StringifyProp(prop)
	}
	return strings.Join(values, ", ")
}

func LoadProp(meta *MetaObj) Prop {
	reOp := Prop{}
	parsedOp := ParseStruct(meta.Value)
	for _, pOp := range parsedOp {
		if pOp[0] == "Key" {
			reOp.Key = pOp[1]
		} else if pOp[0] == "Type" {
			reOp.Type = pOp[1]
		} else if pOp[0] == "Value" {
			reOp.Value = pOp[1]
		} else if pOp[0] == "StructTags" {
			reOp.StructTags = pOp[1]
		}
	}
	return reOp
}

func LoadProps(meta *MetaObj) []Prop { return loadMetaRecords(meta, LoadProp) }

func (pr *Prop) FieldSwitch(fieldStr string) (interface{}, string) {
	switch fieldStr {
	case "Key":
		return pr.Key, "String"
	case "Type":
		return pr.Type, "String"
	case "Value":
		return pr.Value, "String"
	case "StructTags":
		return pr.StructTags, "String"
	default:
		return "Invalid", "String"
	}
}

type Func struct {
	Key      string   `json:"key,omitempty"`
	Inputs   []Prop   `json:"inputs,omitempty"`
	Outputs  []string `json:"outputs,omitempty"`
	Receiver Prop     `json:"receiver,omitempty"`
}

func LoadFuncInput(key string, inputs []SegmentInput) SegmentInput {
	fields := make([]string, len(inputs))
	for i, input := range inputs {
		switch input.Key {
		case "Key", "Inputs", "Outputs", "Receiver":
			fields[i] = input.Key + ": " + input.Value
		}
	}
	return SegmentInput{Key: key, Type: "Hashmap", Value: "#[Func{" + strings.Join(fields, ", ") + "}Func]#"}
}

func LoadFunc(meta *MetaObj) Func {
	result := Func{}
	for _, field := range ParseStruct(meta.Value) {
		switch field[0] {
		case "Key":
			result.Key = field[1]
		case "Inputs":
			result.Inputs = LoadProps(&MetaObj{Value: field[1]})
		case "Outputs":
			result.Outputs = strings.Split(strings.Trim(strings.Trim(field[1], "["), "]"), ", ")
		case "Receiver":
			result.Receiver = LoadProp(&MetaObj{Value: field[1]})
		}
	}
	return result
}

func LoadFuncs(meta *MetaObj) []Func { return loadMetaRecords(meta, LoadFunc) }

func (fu *Func) FieldSwitch(fieldStr string) (interface{}, string) {
	switch fieldStr {
	case "Key":
		return fu.Key, "String"
	case "Inputs":
		return fu.Inputs, "Array-Prop"
	case "Outputs":
		return fu.Outputs, "Array-String"
	case "Receiver":
		return fu.Receiver, "Prop"
	default:
		return "Invalid", "String"
	}
}

type Recur struct {
	Type string `json:"type,omitempty"`
	Key  string `json:"key,omitempty"`
}

func LoadRecur(_ *MetaObj) Recur {
	return Recur{}
}

func LoadRecurs(_ *MetaObj) []Recur {
	return []Recur{}
}

func (r *Recur) FieldSwitch(fieldStr string) (interface{}, string) {
	switch fieldStr {
	case "Type":
		return r.Type, "String"
	case "Key":
		return r.Key, "String"
	default:
		return "Invalid", "String"
	}
}

func MetaStructFieldSwitch(value interface{}, kind, field string) (interface{}, string) {
	strict := false
	if reflected, ok := value.(reflect.Value); ok {
		value, strict = reflected.Interface(), true
	}
	switch kind {
	case "Op":
		return metaRecordField(value, strict, field, (*Op).FieldSwitch)
	case "Prop":
		return metaRecordField(value, strict, field, (*Prop).FieldSwitch)
	case "Func":
		return metaRecordField(value, strict, field, (*Func).FieldSwitch)
	case "Recur":
		return metaRecordField(value, strict, field, (*Recur).FieldSwitch)
	default:
		return "Invalid", "String"
	}
}

// Reflected mismatches reject; direct mismatches retain the zero-record lookup.
func metaRecordField[T any](value interface{}, strict bool, field string, lookup func(*T, string) (interface{}, string)) (interface{}, string) {
	record, matches := value.(T)
	if strict && !matches {
		return "Error", "String"
	}
	return lookup(&record, field)
}

func MetaStructCheckSwitch(kind string) (bool, bool) {
	kind = parseMetaTokenHashStr(kind)
	singular := strings.TrimSuffix(kind, "s")
	switch singular {
	case "Op", "Prop", "Func", "Recur":
		return true, singular != kind
	default:
		return false, false
	}
}

func MetaStructHashSwitch(value interface{}, kind string, position int) (interface{}, string) {
	switch kind {
	case "Ops":
		return metaRecordElement[Op](value, position, "Op")
	case "Props":
		return metaRecordElement[Prop](value, position, "Prop")
	case "Funcs":
		return metaRecordElement[Func](value, position, "Func")
	case "Recurs":
		return metaRecordElement[Recur](value, position, "Recur")
	default:
		return "Error", "String"
	}
}

func metaRecordElement[T any](value interface{}, position int, kind string) (interface{}, string) {
	records, ok := value.([]T)
	if !ok {
		return "Error", "String"
	}
	return records[position], kind
}
