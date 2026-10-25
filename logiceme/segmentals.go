/*
Author: Connor Sanders
Version: 0.0.1
Released: 11/01/2020

-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
		Logiceme Evaluator V0.0.1
-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
*/

package Logiceme

import (
	"reflect"
	"strings"
)

/*
===============================================================================================
GENERAL STRUCTURES
===============================================================================================
*/

// Op
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
func LoadOps(meta *MetaObj) []Op {
	reOps := []Op{}
	for _, opMeta := range ParseStructArray(meta.Value) {
		reOps = append(reOps, LoadOp(opMeta))
	}
	return reOps
}

// FieldSwitch
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

// Prop
type Prop struct {
	Key        string `json:"key,omitempty"`
	Type       string `json:"type,omitempty"`
	Value      string `json:"value,omitempty"`
	StructTags string `json:"struct_tags,omitempty"`
}

// IsVoid
func (pr *Prop) IsVoid() bool {
	reBool := false
	eCt := 0
	if pr.Key == "" {
		eCt = eCt + 1
	}
	if pr.Type == "" {
		eCt = eCt + 1
	}
	if pr.Value == "" {
		eCt = eCt + 1
	}
	if pr.StructTags == "" {
		eCt = eCt + 1
	}
	if eCt == 4 {
		reBool = true
	}
	return reBool
}

// StringifyProp
func StringifyProp(prop Prop) string {
	return prop.Key + " " + prop.Type
}

// StringifyProps
func StringifyProps(props []Prop) string {
	reStr := ""
	for ind, prop := range props {
		sProp := StringifyProp(prop)
		if ind == 0 {
			reStr = reStr + sProp
		} else {
			reStr = reStr + ", " + sProp
		}
	}
	return reStr
}

// LoadProp
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

// LoadProps
func LoadProps(meta *MetaObj) []Prop {
	reOps := []Prop{}
	opMetas := ParseStructArray(meta.Value)
	for _, opMeta := range opMetas {
		op := LoadProp(opMeta)
		reOps = append(reOps, op)
	}
	return reOps
}

// FieldSwitch
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

// Func
type Func struct {
	Key      string   `json:"key,omitempty"`
	Inputs   []Prop   `json:"inputs,omitempty"`
	Outputs  []string `json:"outputs,omitempty"`
	Receiver Prop     `json:"receiver,omitempty"`
}

// LoadFuncInput
func LoadFuncInput(key string, segInputs []SegmentInput) SegmentInput {
	reSeg := SegmentInput{Key: key, Type: "Hashmap"}
	inStr := "#[Func{"
	for ind, segIn := range segInputs {
		if ind > 0 {
			inStr = inStr + ", "
		}
		if segIn.Key == "Key" {
			inStr = inStr + "Key: " + segIn.Value
		} else if segIn.Key == "Inputs" {
			inStr = inStr + "Inputs: " + segIn.Value
		} else if segIn.Key == "Outputs" {
			inStr = inStr + "Outputs: " + segIn.Value
		} else if segIn.Key == "Receiver" {
			inStr = inStr + "Receiver: " + segIn.Value
		}
	}
	inStr = inStr + "}Func]#"
	reSeg.Value = inStr
	return reSeg
}

// LoadFunc
func LoadFunc(meta *MetaObj) Func {
	reOp := Func{}
	parsedOp := ParseStruct(meta.Value)
	for _, pOp := range parsedOp {
		if pOp[0] == "Key" {
			reOp.Key = pOp[1]
		} else if pOp[0] == "Inputs" {
			reProps := []Prop{}
			inArray := ParseStructArray(pOp[1])
			for _, opMeta := range inArray {
				prop := LoadProp(opMeta)
				reProps = append(reProps, prop)
			}
			reOp.Inputs = reProps
		} else if pOp[0] == "Outputs" {
			cStr := strings.Trim(strings.Trim(pOp[1], "["), "]")
			sStrs := strings.Split(cStr, ", ")
			reOp.Outputs = sStrs
		} else if pOp[0] == "Receiver" {
			prop := Prop{}
			p := ParseStruct(pOp[1])
			for _, pp := range p {
				if pp[0] == "Key" {
					prop.Key = pp[1]
				} else if pp[0] == "Type" {
					prop.Type = pp[1]
				} else if pp[0] == "Value" {
					prop.Value = pp[1]
				} else if pp[0] == "StructTags" {
					prop.StructTags = pp[1]
				}
			}
			reOp.Receiver = prop
		}
	}
	return reOp
}

// LoadFuncs
func LoadFuncs(meta *MetaObj) []Func {
	reOps := []Func{}
	opMetas := ParseStructArray(meta.Value)
	for _, opMeta := range opMetas {
		op := LoadFunc(opMeta)
		reOps = append(reOps, op)
	}
	return reOps
}

// FieldSwitch
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

// Recur
type Recur struct {
	Type string `json:"type,omitempty"`
	Key  string `json:"key,omitempty"`
}

// LoadRecur
func LoadRecur(_ *MetaObj) Recur {
	return Recur{}
}

// LoadRecurs
func LoadRecurs(_ *MetaObj) []Recur {
	return []Recur{}
}

// FieldSwitch
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

// MetaStructFieldSwitch
func MetaStructFieldSwitch(structEnt interface{}, typeStr string, structField string) (interface{}, string) {
	refVal := false
	var structInt interface{}
	if reflected, ok := structEnt.(reflect.Value); ok {
		structInt = reflected.Interface()
		refVal = true
	}
	switch typeStr {
	case "Op":
		structure := Op{}
		ok := false
		if refVal {
			structure, ok = structInt.(Op)
			if !ok {
				return "Error", "String"
			}
		} else {
			switch value := structEnt.(type) {
			case Op:
				structure = value
			}
		}
		return structure.FieldSwitch(structField)
	case "Prop":
		structure := Prop{}
		ok := false
		if refVal {
			structure, ok = structInt.(Prop)
			if !ok {
				return "Error", "String"
			}
		} else {
			switch value := structEnt.(type) {
			case Prop:
				structure = value
			}
		}
		return structure.FieldSwitch(structField)
	case "Func":
		structure := Func{}
		ok := false
		if refVal {
			structure, ok = structInt.(Func)
			if !ok {
				return "Error", "String"
			}
		} else {
			switch value := structEnt.(type) {
			case Func:
				structure = value
			}
		}
		return structure.FieldSwitch(structField)
	case "Recur":
		structure := Recur{}
		ok := false
		if refVal {
			structure, ok = structInt.(Recur)
			if !ok {
				return "Error", "String"
			}
		} else {
			switch value := structEnt.(type) {
			case Recur:
				structure = value
			}
		}
		return structure.FieldSwitch(structField)
	default:
		return "Invalid", "String"
	}
}

// MetaStructCheckSwitch
func MetaStructCheckSwitch(typeStr string) (bool, bool) {
	typeStr = parseMetaTokenHashStr(typeStr)
	switch typeStr {
	case "Ops":
		return true, true
	case "Op":
		return true, false
	case "Props":
		return true, true
	case "Prop":
		return true, false
	case "Funcs":
		return true, true
	case "Func":
		return true, false
	case "Recurs":
		return true, true
	case "Recur":
		return true, false
	default:
		return false, false
	}
}

// MetaStructHashSwitch
func MetaStructHashSwitch(structInt interface{}, typeStr string, posInt int) (interface{}, string) {
	switch typeStr {
	case "Ops":
		structArray, ok := structInt.([]Op)
		if !ok {
			return "Error", "String"
		}
		return structArray[posInt], "Op"
	case "Props":
		structArray, ok := structInt.([]Prop)
		if !ok {
			return "Error", "String"
		}
		return structArray[posInt], "Prop"
	case "Funcs":
		structArray, ok := structInt.([]Func)
		if !ok {
			return "Error", "String"
		}
		return structArray[posInt], "Func"
	case "Recurs":
		structArray, ok := structInt.([]Recur)
		if !ok {
			return "Error", "String"
		}
		return structArray[posInt], "Recur"
	default:
		return "Error", "String"
	}
}
