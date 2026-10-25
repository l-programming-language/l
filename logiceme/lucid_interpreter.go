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
	"bytes"
	"encoding/json"
	//"encoding/gob"
	"fmt"
	"reflect"
	// "regexp"
	"strconv"
	"strings"
)

// LucidInterpreter
type LucidInterpreter struct {
	RawContent         []byte
	InterpretedContent []byte
	SubTags            []*SubTag
	InterpretedSubTags []*InterpretedSubTag
	SegmentalSystem    *SegmentalSystem
	Meta               MetaDict
	Conjoinments       []CompiledConjoinment
	WorkingVariables   []*WorkingVariable
}

/*
=================================
LUCID VARIABLE SWITCHES
=================================
*/

// varInitSwitch
func varInitSwitch(curToken string) (string, string) {
	switch curToken {
	case "Meta":
		return "Meta", "Root"
	case "Conjoinments":
		return "Conjoinments", "Root"
	default:
		return "Other", "Void"
	}
}

// parseMetaTokenHashStr
func parseMetaTokenHashStr(curToken string) string {
	if strings.Contains(curToken, "[") && strings.Contains(curToken, "]") {
		curToken = strings.Split(curToken, "[")[0]
	}
	return curToken
}

// metaSwitch
func (li *LucidInterpreter) metaSwitch(curToken string) (interface{}, string) {
	var featural string
	curToken = parseMetaTokenHashStr(curToken)
	curToken, featural = scanTokenVarFeaturals(curToken)
	metaVar := li.Meta.Get(curToken)
	// Primitive interpretation can bypass AML preparation. Only these optional
	// Rust inputs acquire canonical absence without changing the caller's Meta.
	if li.SegmentalSystem != nil && li.SegmentalSystem.Language == "Rust" &&
		metaVar.Type == "NotFound" && (curToken == "Lifetimes" || curToken == "GenericTypes") {
		return "Void", "String"
	}
	if metaVar.Type == "SourceString" {
		return metaVar.Value, "String"
	}
	if metaVar.Type == "PreparedOps" {
		var values []Op
		if json.Unmarshal([]byte(metaVar.Value), &values) == nil {
			return values, "Ops"
		}
	}
	if metaVar.Type == "PreparedStrings" {
		var values []string
		if json.Unmarshal([]byte(metaVar.Value), &values) == nil {
			return values, "Array-String"
		}
	}
	if metaVar.Value == "Void" {
		return "Void", "String"
	}
	switch curToken {
	case "Ops":
		return LoadOps(metaVar), "Ops"
	case "Op":
		return LoadOp(metaVar), "Op"
	case "Props":
		if li.SegmentalSystem != nil && li.SegmentalSystem.Language == "Rust" {
			props, err := readNativeProps(metaVar.Value)
			if err == nil {
				return props, "Props"
			}
		}
		return LoadProps(metaVar), "Props"
	case "Prop":
		return LoadProp(metaVar), "Prop"
	case "Funcs":
		if li.SegmentalSystem != nil && li.SegmentalSystem.Language == "Rust" {
			functions, err := readNativeFuncs(metaVar.Value)
			if err != nil {
				// AML preparation reports malformed records before rendering.
				// Direct interpretation must not retry the generic decoder.
				return []Func(nil), "Funcs"
			}
			return functions, "Funcs"
		}
		return LoadFuncs(metaVar), "Funcs"
	case "Func":
		if li.SegmentalSystem != nil && li.SegmentalSystem.Language == "Rust" {
			function, err := readNativeFunc(map[string]string{"Func": metaVar.Value})
			if err == nil {
				return function, "Func"
			}
		}
		return LoadFunc(metaVar), "Func"
	case "Recurs":
		return LoadRecurs(metaVar), "Recurs"
	case "Recur":
		return LoadRecur(metaVar), "Recur"
	default:
		if metaVar.Type == "Array" && featural == "~" {
			reStr := StringifyStringArray(metaVar.Value)
			return reStr, "String"
		}
		return metaVar.Value, metaVar.Type
	}
}

// convertBytesToStr
func convertBytesToStr(contents [][]byte) []string {
	var reStrs []string
	for _, reByte := range contents {
		reStr := string(reByte)
		reStrs = append(reStrs, reStr)
	}
	return reStrs
}

// segmentalInterfaceSwitch
func segmentalInterfaceSwitch(segmentalType string, conjoinment CompiledConjoinment) ([]string, string) {
	switch segName := segmentalType; segName {
	case "Declaration":
		return convertBytesToStr(conjoinment.Contents), "Declaration"
	case "Calculation":
		return convertBytesToStr(conjoinment.Contents), "Calculation"
	case "Evaluation":
		return convertBytesToStr(conjoinment.Contents), "Evaluation"
	case "Definition":
		return convertBytesToStr(conjoinment.Contents), "Definition"
	case "Description":
		return convertBytesToStr(conjoinment.Contents), "Description"
	case "Operation":
		return convertBytesToStr(conjoinment.Contents), "Operation"
	case "Iteration":
		return convertBytesToStr(conjoinment.Contents), "Iteration"
	case "Instantiation":
		return convertBytesToStr(conjoinment.Contents), "Instantiation"
	case "Execution":
		return convertBytesToStr(conjoinment.Contents), "Execution"
	}
	return []string{"Void"}, "Void"
}

// LoadSegmentalConjoinments
func LoadSegmentalConjoinments(segmentalType string, conjoinments []CompiledConjoinment) ([]string, string) {
	var reStrs []string
	switch segName := segmentalType; segName {
	case "Declaration":
		for _, conj := range conjoinments {
			conjBytes, _ := segmentalInterfaceSwitch(segmentalType, conj)
			reStrs = append(reStrs, conjBytes...)
		}
		return reStrs, "Declarations"
	case "Calculation":
		for _, conj := range conjoinments {
			conjBytes, _ := segmentalInterfaceSwitch(segmentalType, conj)
			reStrs = append(reStrs, conjBytes...)
		}
		return reStrs, "Calculations"
	case "Evaluation":
		for _, conj := range conjoinments {
			conjBytes, _ := segmentalInterfaceSwitch(segmentalType, conj)
			reStrs = append(reStrs, conjBytes...)
		}
		return reStrs, "Evaluations"
	case "Definition":
		for _, conj := range conjoinments {
			conjBytes, _ := segmentalInterfaceSwitch(segmentalType, conj)
			reStrs = append(reStrs, conjBytes...)
		}
		return reStrs, "Definitions"
	case "Description":
		for _, conj := range conjoinments {
			conjBytes, _ := segmentalInterfaceSwitch(segmentalType, conj)
			reStrs = append(reStrs, conjBytes...)
		}
		return reStrs, "Descriptions"
	case "Operation":
		for _, conj := range conjoinments {
			conjBytes, _ := segmentalInterfaceSwitch(segmentalType, conj)
			reStrs = append(reStrs, conjBytes...)
		}
		return reStrs, "Operations"
	case "Iteration":
		for _, conj := range conjoinments {
			conjBytes, _ := segmentalInterfaceSwitch(segmentalType, conj)
			reStrs = append(reStrs, conjBytes...)
		}
		return reStrs, "Iterations"
	case "Instantiation":
		for _, conj := range conjoinments {
			conjBytes, _ := segmentalInterfaceSwitch(segmentalType, conj)
			reStrs = append(reStrs, conjBytes...)
		}
		return reStrs, "Instantiations"
	case "Execution":
		for _, conj := range conjoinments {
			conjBytes, _ := segmentalInterfaceSwitch(segmentalType, conj)
			reStrs = append(reStrs, conjBytes...)
		}
		return reStrs, "Executions"
	}
	return []string{"Void"}, "Void"
}

// processConjoinmentsSwitch
func (li *LucidInterpreter) processConjoinmentsSwitch(conjType string, plural bool, pos bool, posInt int) ([]string, string) {
	segmentalType := li.SegmentalSystem.SegmentalDataSwitch(conjType)
	if plural {
		var reConjs []CompiledConjoinment
		for _, conjoinment := range li.Conjoinments {
			if conjoinment.Type == conjType {
				reConjs = append(reConjs, conjoinment)
			}
		}
		if pos {
			if len(reConjs) > 0 {
				if len(reConjs[0].Contents) > posInt {
					reConj := reConjs[0]
					reConj.Contents = [][]byte{reConj.Contents[posInt]}
					return segmentalInterfaceSwitch(segmentalType, reConj)
				}
			}
			return []string{"Void"}, "Void"
		}
		return LoadSegmentalConjoinments(segmentalType, reConjs)
	}
	for _, conjoinment := range li.Conjoinments {
		if conjoinment.Type == conjType {
			return segmentalInterfaceSwitch(segmentalType, conjoinment)
		}
	}
	return []string{"Void"}, "Void"
}

// conjoinmentsSwitch
func (li *LucidInterpreter) conjoinmentsSwitch(curToken string) ([]string, string) {
	pos := false
	posInt := 0
	arrayPOS := ""
	if strings.Contains(curToken, "[") {
		if !strings.Contains(curToken, "]") {
			curToken = curToken + "]"
		}
	}
	curToken, arrayPOS = parseMetaStructArrayPOS(curToken)
	if arrayPOS != "" {
		pos = true
		posType, posVal := checkMetaArrayPOS(arrayPOS)
		posInt = 0
		if posType == "Function" {
			intPost, _ := li.handleLucidMetaStructFunction(posVal, arrayPOS, "calc")
			posInt = intPost.(int) //nolint:errcheck // The calc result selector returns an int for every supported function and its default.
		} else if posType == "Position" {
			_, posInt = checkIfStrInt(arrayPOS)
		}
	}
	if li.SegmentalSystem != nil && li.SegmentalSystem.Language == "Rust" {
		for _, c := range li.Conjoinments {
			// Direct interpreter callers may omit Key. Explicitly forwarded
			// aliases always have one and must not expose their source name.
			if c.Key == curToken || (c.Key == "" && c.Type == curToken) {
				values := convertBytesToStr(c.Contents)
				// Named forwarding of an absent optional singular child retains
				// the absence sentinel. Keyless direct records preserve presence;
				// empty plural collections remain arrays for loops and Length.
				if c.Key != "" && !c.Plural && len(values) == 0 {
					return []string{"Void"}, "Void"
				}
				category := strings.TrimSuffix(c.Type, "s")
				if checkVarReturn(category)[0] == "Void" {
					// Direct interpreter clients historically use arbitrary
					// collection names with the generic text carrier. Real
					// forwarded categories retain their logical identity.
					category = "Calculation"
				}
				if pos {
					if posInt >= 0 && posInt < len(values) {
						return []string{values[posInt]}, category
					}
					return []string{"Void"}, "Void"
				}
				if c.Plural {
					return values, category + "s"
				}
				return values, category
			}
		}
		// An alias exposes only its forwarded name, not its source category.
		return []string{"Void"}, "Void"
	}
	switch curToken {
	case "Declarations":
		return li.processConjoinmentsSwitch(curToken, true, pos, posInt)
	case "Declaration":
		return li.processConjoinmentsSwitch(curToken, false, pos, posInt)
	case "Calculations":
		return li.processConjoinmentsSwitch(curToken, true, pos, posInt)
	case "Calculation":
		return li.processConjoinmentsSwitch(curToken, false, pos, posInt)
	case "Evaluations":
		return li.processConjoinmentsSwitch(curToken, true, pos, posInt)
	case "Evaluation":
		return li.processConjoinmentsSwitch(curToken, false, pos, posInt)
	case "Definitions":
		return li.processConjoinmentsSwitch(curToken, true, pos, posInt)
	case "Definition":
		return li.processConjoinmentsSwitch(curToken, false, pos, posInt)
	case "Descriptions":
		return li.processConjoinmentsSwitch(curToken, true, pos, posInt)
	case "Description":
		return li.processConjoinmentsSwitch(curToken, false, pos, posInt)
	case "Operations":
		return li.processConjoinmentsSwitch(curToken, true, pos, posInt)
	case "Operation":
		return li.processConjoinmentsSwitch(curToken, false, pos, posInt)
	case "Iterations":
		return li.processConjoinmentsSwitch(curToken, true, pos, posInt)
	case "Iteration":
		return li.processConjoinmentsSwitch(curToken, false, pos, posInt)
	case "Instantiations":
		return li.processConjoinmentsSwitch(curToken, true, pos, posInt)
	case "Instantiation":
		return li.processConjoinmentsSwitch(curToken, false, pos, posInt)
	case "Executions":
		return li.processConjoinmentsSwitch(curToken, true, pos, posInt)
	case "Execution":
		return li.processConjoinmentsSwitch(curToken, false, pos, posInt)
	default:
		// TODO Check for variable in Working Variables if var is not root Meta.
		return []string{"Void"}, "Void"
	}
}

// rootSwitch
func (li *LucidInterpreter) rootSwitch(curToken string, switchType string) (interface{}, string) {
	switch switchType {
	case "Meta":
		// TODO Return only data type variable
		return li.metaSwitch(curToken)
	case "Conjoinments":
		// TODO Return either a byte array OR array of byte arrays
		return li.conjoinmentsSwitch(curToken)
	default:
		// TODO Check for variable in Working Variables if var is not root Meta.
		return "Void", "String"
	}
}

/*
=================================
LUCID VARIABLE MANAGEMENT
=================================
*/

// cleanFieldName
func cleanFieldName(fieldStr string) (string, bool, int) {
	if strings.Contains(fieldStr, "[") {
		sStr := strings.Split(fieldStr, "[")
		sIndex := sStr[1]
		cleanField := sStr[0]
		isInt, indInt := checkIfStrInt(sIndex)
		if isInt {
			return cleanField, true, indInt
		}
		return cleanField, false, 0
	}
	return fieldStr, false, 0
}

// cleanStructKey
func cleanStructKey(keyStr string) string {
	if strings.Contains(keyStr, ".") {
		keyStr = strings.Split(keyStr, ".")[1]
	}
	return keyStr
}

// hashFieldArray
func hashFieldArray(strStruct interface{}, fieldType string, fieldIndex int) (interface{}, string) {
	if fieldType == "Array-String" {
		array := strStruct.([]string) //nolint:errcheck // Array-String is paired with a []string by MetaStructFieldSwitch.
		if len(array) > fieldIndex {
			return array[fieldIndex], "String"
		}
	}
	return strStruct, fieldType
}

// findStructData
func (li *LucidInterpreter) findStructData(strStruct interface{}, structType string, fieldName string) (interface{}, string) {
	fieldInd := 0
	isHash := false
	structKey := cleanStructKey(structType)
	fieldName, isHash, fieldInd = cleanFieldName(fieldName)
	reData, dataType := MetaStructFieldSwitch(strStruct, structKey, fieldName)
	if isHash {
		reData, dataType = hashFieldArray(reData, dataType, fieldInd)
	}
	return reData, dataType
}

// createVariable
func (li *LucidInterpreter) createVariable(varName string, varType string, varValue interface{}) *WorkingVariable {
	newVar := newWorkingVariable(varName, varType, varValue)
	li.WorkingVariables = append(li.WorkingVariables, newVar)
	return newVar
}

// getVariable
func (li *LucidInterpreter) getVariable(varName string) *WorkingVariable {
	for _, variable := range li.WorkingVariables {
		if variable.VarName == varName {
			return variable
		}
	}
	return newWorkingVariable(varName, "NotFound", "NotFound")
}

// getVariableValue
func (li *LucidInterpreter) getVariableValue(varName string) (interface{}, string) {
	if strings.Contains(varName, ".") {
		cleanedVariable := strings.Split(varName, ".")[0]
		fieldName := strings.Split(varName, ".")[1]
		wVar := li.getVariable(cleanedVariable)
		if wVar.VarValue == "NotFound" {
			varType, varVal := li.determineVarValues(varName)
			return varVal, varType
		}
		return li.findStructData(wVar.VarValue, wVar.VarType, fieldName)
	}
	return li.getVariable(varName).VarValue, li.getVariable(varName).VarType
}

// parseMetaStructArrayPOS
func parseMetaStructArrayPOS(tokenVar string) (string, string) {
	if strings.Contains(tokenVar, "[") && strings.Contains(tokenVar, "]") {
		arrayPOS := strings.Split(tokenVar, "[")[1]
		arrayPOS = strings.Split(arrayPOS, "]")[0]
		curToken := strings.Split(tokenVar, "[")[0]
		return curToken, arrayPOS
	}
	return tokenVar, ""
}

// LucidMetaFunctionSwitch
func LucidMetaFunctionSwitch(checkStr string) (bool, string) {
	switch checkStr {
	case "Count":
		return true, "Count"
	case "Length":
		return true, "Length"
	case "Add":
		return true, "Add"
	default:
		return false, "Void"
	}
}

// checkMetaArrayPOS
func checkMetaArrayPOS(arrayPOS string) (string, string) {
	var function bool
	var funcType string
	if strings.Contains(arrayPOS, "(") {
		posKey := strings.Split(arrayPOS, "(")[0]
		function, funcType = LucidMetaFunctionSwitch(posKey)
		if function {
			return "Function", funcType
		}
		return "Error", funcType
	}
	isInt, _ := checkIfStrInt(arrayPOS)
	if isInt {
		return "Position", arrayPOS
	}
	return "Error", "Void"
}

// scanFuncInputs
func scanFuncInputs(funcStr string) []string {
	input := strings.Split(funcStr, "(")[1]
	input = strings.Split(input, ")")[0]
	if strings.Contains(input, ", ") {
		sInputs := strings.Split(input, ", ")
		return sInputs
	}
	return []string{input}
}

// checkAddInputs
func checkAddInputs(inputs []string) []string {
	var re []string
	for _, input := range inputs {
		isInt, _ := checkIfStrInt(input)
		if isInt {
			re = append(re, input)
		} else {
			re = append(re, "VARIABLE:"+input)
		}
	}
	return re
}

// lengthHandler
func (li *LucidInterpreter) lengthHandler() int {
	res := 0
	// Length(Meta.Evals)
	return res
}

// addHandler
func (li *LucidInterpreter) addHandler(funcStr string) int {
	res := 0
	// Add(Count, -1)
	var intInputs []int
	inputs := scanFuncInputs(funcStr)
	checkedInputs := checkAddInputs(inputs)
	for _, checkedInput := range checkedInputs {
		if !strings.Contains(checkedInput, "VARIABLE:") {
			_, intVal := checkIfStrInt(checkedInput)
			intInputs = append(intInputs, intVal)
		} else {
			varS := strings.Split(checkedInput, ":")[1]
			inputType, inputVal := li.determineVarValues(varS)
			if inputType == "Number" {
				intVal := inputVal.([]interface{})[0].(reflect.Value).Int() //nolint:errcheck // The Number working-variable representation contains reflected numeric values.
				s := fmt.Sprintf("%v", intVal)
				iVal, _ := strconv.Atoi(s) //nolint:errcheck // Retain the numeric conversion result, including saturation on host-int overflow.
				intInputs = append(intInputs, iVal)
			} else {
				fmt.Println("===========================TYPE ERROR IN ADD HANDLER========================")
				fmt.Println(inputType)
				return res
			}
		}
	}
	for _, addInt := range intInputs {
		res = res + addInt
	}
	return res
}

// countHandler
func (li *LucidInterpreter) countHandler() []int {
	// Count(Meta.Evals}
	// [0, 1, 2]
	reCount := []int{}
	return reCount
}

// handleLucidMetaStructFunction
func (li *LucidInterpreter) handleLucidMetaStructFunction(funcType string, arrayPOS string, resultType string) (interface{}, string) {
	switch funcType {
	case "Count":
		if resultType == "*" || resultType == "count" {
			return li.countHandler(), "Array-Number"
		}
	case "Length":
		if resultType == "*" || resultType == "calc" {
			return li.lengthHandler(), "Number"
		}
	case "Add":
		if resultType == "*" || resultType == "calc" {
			return li.addHandler(arrayPOS), "Number"
		}
	default:
		return 0, "Number"
	}
	return 0, "Number"
}

// handleLucidMetaStruct
func (li *LucidInterpreter) handleLucidMetaStruct(tokenVar string, structInt interface{}, structType string, plural bool) (interface{}, string) {
	curToken, arrayPOS := parseMetaStructArrayPOS(tokenVar)
	if arrayPOS != "" {
		if !plural {
			return "Error", "Void"
		}
		posType, posVal := checkMetaArrayPOS(arrayPOS)
		posInt := 0
		if posType == "Function" {
			intPost, _ := li.handleLucidMetaStructFunction(posVal, arrayPOS, "calc")
			posInt = intPost.(int) //nolint:errcheck // The calc result selector returns an int for every supported function and its default.
		} else if posType == "Position" {
			_, posInt = checkIfStrInt(arrayPOS)
		}
		hashedMetaStruct, hashedType := MetaStructHashSwitch(structInt, structType, posInt)
		return hashedMetaStruct, hashedType
	}
	return structInt, curToken
}

// scanTokenVarFeaturals
func scanTokenVarFeaturals(tokenVar string) (string, string) {
	lVal := tokenVar[len(tokenVar)-1:]
	if lVal == "~" {
		tokenVar = strings.Trim(tokenVar, "~")
		return tokenVar, lVal
	}
	return tokenVar, ""
}

// handleLucidMetaStructVal
func (li *LucidInterpreter) handleLucidMetaStructVal(tokenVar string, structInt interface{}, structType string) (interface{}, string) {
	var featural string
	tokenVar, featural = scanTokenVarFeaturals(tokenVar)
	curToken, arrayPOS := parseMetaStructArrayPOS(tokenVar)
	fieldVal, fieldType := MetaStructFieldSwitch(structInt, structType, curToken)
	if fieldType == "Array-String" {
		posType, posVal := checkMetaArrayPOS(arrayPOS)
		posInt := 0
		if posType == "Function" {
			intPost, _ := li.handleLucidMetaStructFunction(posVal, arrayPOS, "calc")
			posInt = intPost.(int) //nolint:errcheck // The calc result selector returns an int for every supported function and its default.
		} else if posType == "Position" {
			_, posInt = checkIfStrInt(arrayPOS)
		}
		arrayVal := fieldVal.([]string) //nolint:errcheck // MetaStructFieldSwitch pairs Array-String with a []string payload.
		if arrayPOS != "" {
			return arrayVal[posInt], "String"
		}
		if len(arrayVal) == 0 {
			return "Void", "String"
		} else if featural == "~" {
			reStr := StringifyArray(arrayVal)
			return reStr, "String"
		}
		return fieldVal, fieldType
	} else if fieldType == "Array-Prop" {
		arrayVal := fieldVal.([]Prop) //nolint:errcheck // MetaStructFieldSwitch pairs Array-Prop with a []Prop payload.
		if len(arrayVal) == 0 {
			return "Void", "String"
		} else if featural == "~" {
			// Rust callable parameters retain their existing Prop records.
			// Other Prop interpolation is unchanged.
			if li.SegmentalSystem != nil && li.SegmentalSystem.Language == "Rust" &&
				structType == "Func" && curToken == "Inputs" {
				parameters := make([]string, len(arrayVal))
				for i, prop := range arrayVal {
					parameters[i] = prop.Key + ": " + prop.Type
				}
				return strings.Join(parameters, ", "), "String"
			}
			reStr := StringifyProps(arrayVal)
			return reStr, "String"
		}
	} else if fieldType == "Prop" {
		val := fieldVal.(Prop) //nolint:errcheck // MetaStructFieldSwitch pairs Prop with a Prop payload.
		if val.IsVoid() {
			return "Void", "String"
		} else if featural == "~" {
			reStr := StringifyProp(val)
			return reStr, "String"
		}
	}
	return fieldVal, fieldType
}

// getLucidVarVal
func (li *LucidInterpreter) getLucidVarVal(tokenizedVar []string) (interface{}, string) {
	// Examples of usage
	// for TableField in Meta.
	var switchRe interface{}
	var switchStruct interface{}
	var fieldVal interface{}
	switchStructType := ""
	fieldType := ""
	metaStruct := false
	metaPlural := false
	rootSet := false
	var varInit, initType, reType string
	for ind, tokenVar := range tokenizedVar {
		if ind == 0 {
			varInit, initType = varInitSwitch(tokenVar)
		} else {
			if initType == "Root" && !rootSet {
				switchRe, reType = li.rootSwitch(tokenVar, varInit)
				rootSet = true
				if varInit != "Conjoinments" {
					metaStruct, metaPlural = MetaStructCheckSwitch(reType)
				}
			}
			if varInit != "Conjoinments" {
				if metaStruct {
					if switchStructType != "" {
						// Get struct's field name
						fieldVal, fieldType = li.handleLucidMetaStructVal(tokenVar, switchStruct, switchStructType)
						return fieldVal, fieldType
					} else {
						switchStruct, switchStructType = li.handleLucidMetaStruct(tokenVar, switchRe, reType, metaPlural)
						if ind == len(tokenizedVar)-1 {
							return switchStruct, switchStructType
						}
					}
				}
			}
		}
	}
	return switchRe, reType
}

// preprocessLucidVar
func (li *LucidInterpreter) preprocessLucidVar(cleanedVariable string) []string {
	tokenizedVar := []string{}
	// Sub Variables being referenced by name (dot) or key (square brackets)
	if strings.Contains(cleanedVariable, ".") || strings.Contains(cleanedVariable, "[") {
		varNameParts := strings.Split(cleanedVariable, ".")
		tokenizedVar = append(tokenizedVar, varNameParts...)
	} else {
		tokenizedVar = append(tokenizedVar, cleanedVariable)
	}
	return tokenizedVar
}

// determineVarValues
func (li *LucidInterpreter) determineVarValues(varName string) (string, interface{}) {
	if strings.Contains(varName, ".") {
		if strings.Contains(varName, "Conjoinments.") {
			if strings.Contains(varName, "[") && strings.Contains(varName, "]") {
				varName = strings.Trim(varName, "]")
			}
		}
		tokenizedVar := li.preprocessLucidVar(varName)
		varVal, varType := li.getLucidVarVal(tokenizedVar)
		return varType, varVal
	}
	wVar := li.getVariable(varName)
	if reflect.ValueOf(wVar.VarValue).Kind() != reflect.Slice {
		var reSlice []interface{}
		reSlice = append(reSlice, wVar.VarValue)
		return wVar.VarType, reSlice
	} else {
		var reSlice []interface{}
		s := reflect.ValueOf(wVar.VarValue)
		for i := 0; i < s.Len(); i++ {
			reSlice = append(reSlice, s.Index(i))
		}
		return wVar.VarType, reSlice
	}
}

// cleanLucidVar
func cleanLucidVar(varStr string) string {
	// Delimiters are tokens, not character sets: the final ] may belong to
	// a field index immediately before the closing Lucid token.
	for _, frame := range [][2]string{{"[|(", ")|]"}, {"[|%", "%|]"}, {"[|&", "&|]"}, {"[|", "|]"}} {
		if strings.HasPrefix(varStr, frame[0]) && strings.HasSuffix(varStr, frame[1]) {
			return varStr[len(frame[0]) : len(varStr)-len(frame[1])]
		}
	}
	return varStr
}

// cleanNewLines
func cleanNewLines(varVal string, varTag string) string {
	if strings.Contains(varTag, "%") {
		varVal = string(inlineConjoinmentNewLines([]byte(varVal)))
	} else if strings.Contains(varTag, "&") {
		varVal = TrimPrefixes(varVal, "\n")
	}
	return varVal
}

// interpretExpressionVars
func (li *LucidInterpreter) interpretExpressionVars(expressionVars []string, curExpressionStmt string) string {
	removeDuplicateHeaders(&expressionVars)
	for _, variable := range expressionVars {
		cleanedVariable := cleanLucidVar(variable)
		if strings.Contains(cleanedVariable, "(") && !strings.Contains(cleanedVariable, ".") {
			fmt.Print("<------Pass, will run a interpretExpressionVarFunction soon....-------->\n\n")
		} else {
			rootCleanedVariable := cleanedVariable
			if strings.Contains(cleanedVariable, ".") {
				rootCleanedVariable = strings.Split(cleanedVariable, ".")[0]
			}
			varVal := li.getVariable(rootCleanedVariable)
			if varVal.VarType == "NotFound" {
				var metaVarVal interface{}
				var varType, varStr string
				if strings.Contains(cleanedVariable, "Conjoinments.") {
					metaVarVal, varType = li.handleConditionalMetaVar(cleanedVariable)
					if varType == "Void" {
						curExpressionStmt = strings.ReplaceAll(curExpressionStmt, variable, "")
						continue
					}
				} else {
					tokenizedVar := li.preprocessLucidVar(cleanedVariable)
					metaVarVal, varType = li.getLucidVarVal(tokenizedVar)
				}
				if varType == "String" {
					varStr = fmt.Sprintf("%v", metaVarVal)
					if varStr == "Void" && !li.explicitSourceString(cleanedVariable) {
						varStr = ""
					} else {
						if strings.Contains(variable, "[|(") && strings.Contains(variable, ")|]") {
							varStr = "(" + varStr + ")"
						}
					}
					curExpressionStmt = strings.Replace(curExpressionStmt, variable, varStr, -1)
				} else {
					if strings.Contains(cleanedVariable, "Conjoinments.") {
						lastStr := varType[len(varType)-1:]
						if lastStr != "s" {
							if len(metaVarVal.([]string)) > 0 { //nolint:errcheck // Singular conjoinment projections are supplied as []string by conjoinmentsSwitch.
								varStr = fmt.Sprintf("%v", metaVarVal.([]string)[0]) //nolint:errcheck // The preceding conjoinment projection has already established this []string payload.
							} else {
								varStr = "Void"
							}
							varStr = cleanNewLines(varStr, variable)
							curExpressionStmt = strings.Replace(curExpressionStmt, variable, varStr, -1)
						}
					}
				}
			} else {
				sType := strings.Split(varVal.VarType, ".")
				if sType[0] == "Conjoinment" {
					v := varVal.VarValue
					y := v.(reflect.Value) //nolint:errcheck // Conjoinment working-variable values retain their reflected representation.
					curExpressionStmt = strings.Replace(curExpressionStmt, variable, cleanNewLines(y.String(), variable), -1)
				} else if varVal.VarType != "String" {
					fieldName := strings.Split(variable, ".")[1]
					fieldName = strings.Trim(fieldName, "|]")
					res, _ := li.findStructData(varVal.VarValue, varVal.VarType, fieldName)
					varStr := fmt.Sprintf("%v", res)
					curExpressionStmt = strings.Replace(curExpressionStmt, variable, varStr, -1)
				} else {
					varStr := fmt.Sprintf("%v", varVal.VarValue)
					curExpressionStmt = strings.Replace(curExpressionStmt, variable, varStr, -1)
				}
			}
		}
	}
	return curExpressionStmt
}

/*
=================================
LUCID EXPRESSION INTERPRETERS
=================================
*/

// scanCondVarStr
func scanCondVarStr(varStr string) (string, string) {
	cleanStr := varStr
	funcType := "none"
	if len(varStr) > 4 {
		check := varStr[0:4]
		if check == "Add(" {
			cleanStr = strings.Split(strings.Split(cleanStr, "Add(")[1], ")")[0]
			funcType = "Add"
		} else if len(varStr) > 7 {
			check = varStr[0:7]
			if check == "Length(" {
				cleanStr = strings.Split(strings.Split(cleanStr, "Length(")[1], ")")[0]
				funcType = "Length"
			}
		}
	}
	return cleanStr, funcType
}

// handleConditionalMetaVar
func (li *LucidInterpreter) handleConditionalMetaVar(varStr string) (interface{}, string) {
	var funcType string
	varStr, funcType = scanCondVarStr(varStr)
	vType, vVal := li.determineVarValues(varStr)
	if funcType == "Length" {
		checkedType := checkVarReturn(vType)
		if checkedType[0] != "Array" {
			return 1, "Number"
		}
		return varLength(vVal), "Number"
	}
	return vVal, vType
}

// conditionComparison
func (li *LucidInterpreter) conditionComparison(value interface{}, valueType, condVal, condOperator string) string {
	// Compare the already resolved value. Computed expressions such as Length
	// are not working-variable names and must not read or create a binding.
	if valueType == "Number" {
		varStr := fmt.Sprintf("%v", value)
		if strings.Contains(varStr, ".") {
			checkVar, _ := strconv.ParseFloat(varStr, 64)      //nolint:errcheck // Preserve comparisons using the parser result on invalid or out-of-range numbers.
			checkCondVar, _ := strconv.ParseFloat(condVal, 64) //nolint:errcheck // Preserve comparisons using the parser result on invalid or out-of-range numbers.
			if condOperator == "GreaterOrEqual" {
				if checkVar >= checkCondVar {
					return "true"
				}
			} else if condOperator == "LesserOrEqual" {
				if checkVar <= checkCondVar {
					return "true"
				}
			} else if condOperator == "Greater" {
				if checkVar > checkCondVar {
					return "true"
				}
			} else if condOperator == "Lesser" {
				if checkVar < checkCondVar {
					return "true"
				}
			}
		} else {
			checkVar, _ := strconv.Atoi(varStr)      //nolint:errcheck // Preserve comparisons using the parser result on invalid or out-of-range numbers.
			checkCondVar, _ := strconv.Atoi(condVal) //nolint:errcheck // Preserve comparisons using the parser result on invalid or out-of-range numbers.
			if condOperator == "GreaterOrEqual" {
				if checkVar >= checkCondVar {
					return "true"
				}
			} else if condOperator == "LesserOrEqual" {
				if checkVar <= checkCondVar {
					return "true"
				}
			} else if condOperator == "Greater" {
				if checkVar > checkCondVar {
					return "true"
				}
			} else if condOperator == "Lesser" {
				if checkVar < checkCondVar {
					return "true"
				}
			}
		}
	}
	return "false"
}

// interpretCondition
func (li *LucidInterpreter) interpretCondition(conditionStr string) string {
	var condVar string
	condOperator := "Equals"
	splitStr := strings.Split(conditionStr, " == ")
	if strings.Contains(conditionStr, "!=") {
		condOperator = "NotEquals"
		splitStr = strings.Split(conditionStr, " != ")
	} else if strings.Contains(conditionStr, ">=") {
		condOperator = "GreaterOrEqual"
		splitStr = strings.Split(conditionStr, " >= ")
	} else if strings.Contains(conditionStr, "<=") {
		condOperator = "LesserOrEqual"
		splitStr = strings.Split(conditionStr, " <= ")
	} else if strings.Contains(conditionStr, ">") {
		condOperator = "Greater"
		splitStr = strings.Split(conditionStr, " > ")
	} else if strings.Contains(conditionStr, "<") {
		condOperator = "Lesser"
		splitStr = strings.Split(conditionStr, " < ")
	}
	if strings.Contains(splitStr[0], ", ") {
		sStr := splitStr[0]
		sStr = strings.Replace(sStr, ", ", ",", -1)
		condVar = strings.Split(sStr, " ")[1]
		condVar = strings.Replace(condVar, ",", ", ", -1)
	} else {
		condVar = strings.Split(splitStr[0], " ")[1]
	}
	condVal := splitStr[1]
	bareVoid := strings.TrimSpace(condVal) == "Void"
	condVal = strings.Trim(condVal, `"`)
	condVal = strings.Trim(condVal, `"`)
	if strings.Contains(condVal, "Conjoinments.") || strings.Contains(condVal, "Meta.") {
		reInt, reType := li.handleConditionalMetaVar(condVal)
		if reType == "Void" {
			condVal = "Void"
		} else if reType == "Number" || reType == "String" {
			condVal = fmt.Sprintf("%v", reInt)
		}
	}
	var curVarVal interface{}
	var curType string
	if strings.Contains(condVar, "Conjoinments.") || strings.Contains(condVar, "Meta.") {
		curVarVal, curType = li.handleConditionalMetaVar(condVar)
		if curType == "Void" {
			curVarVal = "Void"
		} else if curType == "Number" || curType == "String" {
			curVarVal = fmt.Sprintf("%v", curVarVal)
		}
	} else {
		curVarVal, curType = li.getVariableValue(condVar)
		curVarVal = fmt.Sprintf("%v", curVarVal)
	}
	// An indexed conjoinment is a scalar source fragment, although the
	// conjoinment API represents it as a one-item slice. Compare its value
	// without the internal import stamp so the terminal Void case is usable.
	if values, ok := curVarVal.([]string); ok && len(values) == 1 && checkVarReturn(curType)[0] == "Item" {
		value := []byte(values[0])
		for _, header := range parseMainTags(value) {
			value = bytes.ReplaceAll(value, header, nil)
		}
		curVarVal = strings.TrimSpace(string(value))
	}
	// Rust native source named Void is present, unlike the bare absence
	// sentinel. Quoted "Void" remains an ordinary textual comparison.
	// explicitSourceString also preserves generic interpolation semantics.
	if bareVoid && li.explicitSourceString(condVar) {
		if condOperator == "Equals" {
			return "false"
		}
		if condOperator == "NotEquals" {
			return "true"
		}
	}
	if condOperator == "Equals" {
		if curVarVal == condVal {
			return "true"
		}
	} else if condOperator == "NotEquals" {
		if curVarVal != condVal {
			return "true"
		}
	} else {
		return li.conditionComparison(curVarVal, curType, condVal, condOperator)
	}
	return "false"
}

// interpretConditionHandler
func (li *LucidInterpreter) interpretConditionHandler(conditionStr string) string {
	if strings.Contains(conditionStr, " && ") {
		condEvals := []string{}
		splitStr := strings.Split(conditionStr, " && ")
		for _, cond := range splitStr {
			condEval := li.interpretCondition(cond)
			condEvals = append(condEvals, condEval)
		}
		if li.SegmentalSystem != nil && li.SegmentalSystem.Language == "Rust" {
			for _, result := range condEvals {
				if result != "true" {
					return "false"
				}
			}
			return "true"
		}
		if condEvals[0] == "true" && condEvals[1] == "true" {
			return "true"
		} else {
			return "false"
		}
	} else if strings.Contains(conditionStr, " || ") {
		condEvals := []string{}
		splitStr := strings.Split(conditionStr, " || ")
		for _, cond := range splitStr {
			condEval := li.interpretCondition(cond)
			condEvals = append(condEvals, condEval)
		}
		if condEvals[0] == "true" || condEvals[1] == "true" {
			return "true"
		} else {
			return "false"
		}
	}
	tmp := li.interpretCondition(conditionStr)
	return tmp
}

// renderBody evaluates only the original child occurrences. Literal spans stay in
// place; emitted source is never searched for another control tag.
func (li *LucidInterpreter) renderBody(source string, children []*SubTag, resolveLiterals bool) (string, []*InterpretedSubTag) {
	var output strings.Builder
	var interpreted []*InterpretedSubTag
	cursor := 0
	literal := func(text string) {
		if resolveLiterals {
			if li.SegmentalSystem != nil && li.SegmentalSystem.Language == "Rust" {
				text = li.interpretNativeLiteral(text)
			} else {
				text = li.interpretExpressionVars(lucidParser(text, "ExpressionVar"), text)
			}
		}
		output.WriteString(text)
	}
	for _, child := range children {
		offset := strings.Index(source[cursor:], child.ReplaceString)
		if offset < 0 {
			// A body rewritten by metadata preparation must also have its child
			// tree rebuilt. Never replace unrelated or already rendered text.
			continue
		}
		start := cursor + offset
		literal(source[cursor:start])
		result := li.interpretSubTag(child)
		if result != nil {
			output.WriteString(result.InterpretedString)
			interpreted = append(interpreted, result)
		}
		cursor = start + len(child.ReplaceString)
	}
	literal(source[cursor:])
	return output.String(), interpreted
}

// interpretForEach renders the complete body under each iteration's binding.
// A nested loop can shadow a name without changing its enclosing activation.
func (li *LucidInterpreter) interpretForEach(tagExpressions []*TagExpression) string {
	forExpressionStr, strType := checkForStatementString(tagExpressions[1].ExpressionString)
	iterVarsStr := strings.Split(forExpressionStr, " ")[3]
	iterVarStr := strings.Split(forExpressionStr, " ")[1]
	varType, iterVarsInter := li.determineVarValues(iterVarsStr)
	checkedType := checkVarReturn(varType)
	if checkedType[0] == "Error" {
		return "Error with Variable: " + iterVarsStr
	} else if checkedType[0] != "Array" {
		return iterVarsStr + " of type: " + checkedType[1] + " is not Iterable"
	}
	if strType == "count" {
		varType, iterVarsInter = countArray(iterVarsInter)
	}
	if strings.Contains(forExpressionStr, "Count") {
		checkedType = checkVarReturn(varType)
	}

	var binding *WorkingVariable
	for _, variable := range li.WorkingVariables {
		if variable.VarName == iterVarStr {
			binding = variable
			break
		}
	}
	if binding != nil {
		saved := *binding
		defer func() { *binding = saved }()
	} else {
		length := len(li.WorkingVariables)
		binding = li.createVariable(iterVarStr, checkedType[1], nil)
		defer func() { li.WorkingVariables = li.WorkingVariables[:length] }()
	}

	var output strings.Builder
	iterVars := reflect.ValueOf(iterVarsInter)
	body := tagExpressions[2]
	for index := 0; index < iterVars.Len(); index++ {
		binding.VarType, binding.VarValue = checkedType[1], iterVars.Index(index)
		rendered, _ := li.renderBody(body.ExpressionString, body.ExpressionSubTags, true)
		output.WriteString(rendered)
	}
	return output.String()
}

// interpretConditional preserves literal text even when every child is empty.
func (li *LucidInterpreter) interpretConditional(tagExpressions []*TagExpression) string {
	if li.interpretConditionHandler(tagExpressions[1].ExpressionString) != "true" {
		return ""
	}
	body := tagExpressions[2]
	rendered, _ := li.renderBody(body.ExpressionString, body.ExpressionSubTags, true)
	return rendered
}

// interpretExpression
func (li *LucidInterpreter) interpretExpression(tagExpressions []*TagExpression) string {
	expressionType := tagExpressions[0].ExpressionType
	if expressionType == "forEach" {
		return li.interpretForEach(tagExpressions)
	} else if expressionType == "Conditional" {
		return li.interpretConditional(tagExpressions)
	}
	return ""
}

// interpretSubTag
func (li *LucidInterpreter) interpretSubTag(subTag *SubTag) *InterpretedSubTag {
	if li.SegmentalSystem != nil && li.SegmentalSystem.Language == "Rust" && subTag.TagType == "headers" {
		var output strings.Builder
		value, _ := li.metaSwitch("Imports")
		headers, _ := value.([]string) //nolint:errcheck // Absent or non-string import metadata intentionally projects to no headers.
		for _, header := range headers {
			output.WriteString(strings.ReplaceAll(subTag.Expressions[2].ExpressionString, "|Header|", header))
		}
		return NewInterpretedSubTag(subTag.ReplaceString, output.String())
	}
	var intTag *InterpretedSubTag
	if subTag.TagType != "Headers" && subTag.TagType != "HeaderList" {
		tagExpressions := subTag.Expressions
		intSubTagString := li.interpretExpression(tagExpressions)
		intTag = NewInterpretedSubTag(subTag.ReplaceString, intSubTagString)
	}
	return intTag
}

// insertGlobalMeta
func (li *LucidInterpreter) insertGlobalMeta() {
	// 1: Parse Out varialbes from intStr [| |]
	intStr := string(li.InterpretedContent)
	condVars := lucidParser(intStr, "ExpressionVar")
	intStr = li.interpretExpressionVars(condVars, intStr)
	// 2: Check variables (i.e. for things such as [|(meta.x)|], and or meta structs like OPs, Evals, along with meta functions like .Count)
	// 3: Replace variables for meta values
	intContents := []byte(intStr)
	li.InterpretedContent = intContents
}

// interpretSubTags
func (li *LucidInterpreter) interpretSubTags() {
	// Root literals retain the existing single, final global metadata pass.
	rendered, tags := li.renderBody(string(li.RawContent), li.SubTags, false)
	li.InterpretedSubTags = tags
	li.InterpretedContent = []byte(rendered)
}

// Interpret
func (li *LucidInterpreter) Interpret() {
	if li.SegmentalSystem != nil && li.SegmentalSystem.Language == "Rust" {
		rendered, tags := li.renderBody(string(li.RawContent), li.SubTags, true)
		li.InterpretedContent, li.InterpretedSubTags = []byte(rendered), tags
		return
	}
	li.interpretSubTags()
	li.insertGlobalMeta()
}

// Resolve only occurrences in the template, never marker-like bytes supplied
// by a native fragment or an already interpreted child.
func (li *LucidInterpreter) interpretNativeLiteral(source string) string {
	var out strings.Builder
	cursor := 0
	for _, variable := range lucidParser(source, "ExpressionVar") {
		offset := strings.Index(source[cursor:], variable)
		if offset < 0 {
			continue
		}
		start := cursor + offset
		out.WriteString(source[cursor:start])
		out.WriteString(li.interpretExpressionVars([]string{variable}, variable))
		cursor = start + len(variable)
	}
	out.WriteString(source[cursor:])
	return out.String()
}

// SourceString distinguishes an explicitly supplied native identifier from the
// absent-value sentinel. It is internal prepared metadata, never AML.
func (li *LucidInterpreter) explicitSourceString(variable string) bool {
	if li.SegmentalSystem == nil || li.SegmentalSystem.Language != "Rust" || !strings.HasPrefix(variable, "Meta.") {
		return false
	}
	name := strings.TrimPrefix(variable, "Meta.")
	if li.Meta.Get(name).Type == "SourceString" {
		return true
	}
	// A nonempty callable name is native source, even when spelled Void.
	// Use the same Func decoder as interpolation; root Meta.Key is not
	// authoritative for this record field.
	if name == "Func.Key" {
		value, kind := li.metaSwitch("Func")
		function, ok := value.(Func)
		return kind == "Func" && ok && function.Key != ""
	}
	// Supplied output types likewise retain native provenance when joined.
	// Empty outputs and all other Func fields keep their absence behavior.
	if name == "Func.Outputs~" {
		value, kind := li.metaSwitch("Func")
		function, ok := value.(Func)
		return kind == "Func" && ok && len(function.Outputs) != 0
	}
	// The indexed form of these same two fields uses the selected Func, not
	// the singular record or root metadata. Other indexed fields are unchanged.
	parts := strings.Split(name, ".")
	if len(parts) == 2 && strings.HasPrefix(parts[0], "Funcs[") && strings.HasSuffix(parts[0], "]") &&
		(parts[1] == "Key" || parts[1] == "Outputs~") {
		value, kind := li.metaSwitch("Funcs")
		functions, ok := value.([]Func)
		if kind != "Funcs" || !ok {
			return false
		}
		_, position := parseMetaStructArrayPOS(parts[0])
		positionType, positionValue := checkMetaArrayPOS(position)
		var index int
		switch positionType {
		case "Position":
			var err error
			index, err = strconv.Atoi(positionValue)
			if err != nil {
				return false
			}
		case "Function":
			if positionValue != "Add" {
				return false
			}
			result, _ := li.handleLucidMetaStructFunction(positionValue, position, "calc")
			index, ok = result.(int)
			if !ok {
				return false
			}
		default:
			return false
		}
		if index < 0 || index >= len(functions) {
			return false
		}
		if parts[1] == "Key" {
			return functions[index].Key != ""
		}
		return len(functions[index].Outputs) != 0
	}
	// An explicit operand named Void is source text, not an absent slot.
	if len(parts) == 2 && li.Meta.Get(parseMetaTokenHashStr(parts[0])).Type == "PreparedOps" {
		field, _ := parseMetaStructArrayPOS(parts[1])
		switch field {
		case "Prefix", "Postfix", "Operator", "Operands", "Type":
			return true
		}
	}
	return false
}
