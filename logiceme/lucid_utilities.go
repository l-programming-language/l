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
	"fmt"
	"reflect"
	"strings"
)

/*
==========================
LUCID UTILITY FUNCTIONS
==========================
*/

// removeDuplicateHeaders
func removeDuplicateHeaders(xs *[]string) {
	found := make(map[string]bool)
	j := 0
	for i, x := range *xs {
		if !found[x] {
			found[x] = true
			(*xs)[j] = (*xs)[i]
			j++
		}
	}
	*xs = (*xs)[:j]
}

// checkVarReturn
func checkVarReturn(reVarType string) []string {
	switch reVarType {
	case "Error":
		return []string{"Error", "Error"}
	case "Declarations":
		return []string{"Array", "Conjoinment.Declaration"}
	case "Declaration":
		return []string{"Item", "Conjoinment.Declaration"}
	case "Calculations":
		return []string{"Array", "Conjoinment.Calculation"}
	case "Calculation":
		return []string{"Item", "Conjoinment.Calculation"}
	case "Evaluations":
		return []string{"Array", "Conjoinment.Evaluation"}
	case "Evaluation":
		return []string{"Item", "Conjoinment.Evaluation"}
	case "Definitions":
		return []string{"Array", "Conjoinment.Definition"}
	case "Definition":
		return []string{"Item", "Conjoinment.Definition"}
	case "Descriptions":
		return []string{"Array", "Conjoinment.Description"}
	case "Description":
		return []string{"Item", "Conjoinment.Description"}
	case "Operations":
		return []string{"Array", "Conjoinment.Operation"}
	case "Operation":
		return []string{"Item", "Conjoinment.Operation"}
	case "Iteration":
		return []string{"Array", "Conjoinment.Iteration"}
	case "Iterations":
		return []string{"Item", "Conjoinment.Iteration"}
	case "Instantiations":
		return []string{"Array", "Conjoinment.Instantiation"}
	case "Instantiation":
		return []string{"Item", "Conjoinment.Instantiation"}
	case "Executions":
		return []string{"Array", "Conjoinment.Execution"}
	case "Execution":
		return []string{"Item", "Conjoinment.Execution"}
	case "Ops":
		return []string{"Array", "Meta.Op"}
	case "Op":
		return []string{"Item", "Meta.Op"}
	case "Props":
		return []string{"Array", "Meta.Prop"}
	case "Prop":
		return []string{"Item", "Meta.Prop"}
	case "Funcs":
		return []string{"Array", "Meta.Func"}
	case "Func":
		return []string{"Item", "Meta.Func"}
	case "Recurs":
		return []string{"Array", "Meta.Recur"}
	case "Recur":
		return []string{"Item", "Meta.Recur"}
	case "Array-Number":
		return []string{"Array", "Number"}
	case "Array-String":
		return []string{"Array", "String"}
	default:
		// TODO Check for variable in Working Variables if var is not root Meta.
		return []string{"Void", "Void"}
	}
}

// determineExpressionType
func determineTagType(tagStr string) string {
	tagType := "Variable"
	if strings.Contains(tagStr, "-{") {
		tagType = "Expression"
		if strings.Contains(tagStr, "-{Headers}-") {
			tagType = "HeaderList"
		} else if strings.Contains(tagStr, "-{for Header in Headers}-") {
			tagType = "Headers"
		}
	}
	return tagType
}

// determineExpressionType
func determineExpressionType(rawExpressions []string) string {
	mainExpressionBlock := rawExpressions[1]
	if strings.Contains(mainExpressionBlock, "-{for ") || strings.Contains(mainExpressionBlock, "={for ") {
		return "forEach"
	} else if strings.Contains(mainExpressionBlock, "-{append-for ") || strings.Contains(mainExpressionBlock, "={append-for ") {
		return "Append-ForEach"
	} else if strings.Contains(mainExpressionBlock, "-{if ") || strings.Contains(mainExpressionBlock, "={if ") {
		return "Conditional"
	} else if strings.Contains(mainExpressionBlock, "-{append}-") || strings.Contains(mainExpressionBlock, "={append") {
		return "Appendage"
	} else if strings.Contains(mainExpressionBlock, "-{replace ") || strings.Contains(mainExpressionBlock, "={replace ") {
		return "Replacement"
	} else if strings.Contains(mainExpressionBlock, "-{Attribute:") || strings.Contains(mainExpressionBlock, "={Attribute:") {
		return "Attribute"
	} else if strings.Contains(mainExpressionBlock, "-{Service:") || strings.Contains(mainExpressionBlock, "={Service:") {
		return "Service"
	} else if strings.Contains(mainExpressionBlock, "-{Template:") || strings.Contains(mainExpressionBlock, "={Template:") {
		return "Template"
	}
	return "Unknown"
}

// determineExpressionOutput
func determineExpressionOutput(rawExpressions []string) string {
	mainExpressionBlock := rawExpressions[2]
	if strings.Contains(mainExpressionBlock, "(/={") && strings.Contains(mainExpressionBlock, "}=/)") {
		return "SubTags"
	} else if strings.Contains(mainExpressionBlock, "[|Declare(") && strings.Contains(mainExpressionBlock, ")|]") {
		return "Declare"
	} else if strings.Contains(mainExpressionBlock, "[|Calculate(") && strings.Contains(mainExpressionBlock, ")|]") {
		return "Calculate"
	} else if strings.Contains(mainExpressionBlock, "[|Define(") && strings.Contains(mainExpressionBlock, ")|]") {
		return "Define"
	} else if strings.Contains(mainExpressionBlock, "[|Evaluate(") && strings.Contains(mainExpressionBlock, ")|]") {
		return "Evaluate"
	} else if strings.Contains(mainExpressionBlock, "[|Describe(") && strings.Contains(mainExpressionBlock, ")|]") {
		return "Describe"
	} else if strings.Contains(mainExpressionBlock, "[|Operate(") && strings.Contains(mainExpressionBlock, ")|]") {
		return "Operate"
	} else if strings.Contains(mainExpressionBlock, "[|Instantiate(") && strings.Contains(mainExpressionBlock, ")|]") {
		return "Instantiate"
	} else if strings.Contains(mainExpressionBlock, "[|Execute(") && strings.Contains(mainExpressionBlock, ")|]") {
		return "Execute"
	} else if strings.Contains(mainExpressionBlock, "[|Iterate(") && strings.Contains(mainExpressionBlock, ")|]") {
		return "Iterate"
	} else if strings.Contains(mainExpressionBlock, "[|") && strings.Contains(mainExpressionBlock, "|]") {
		return "Variable"
	} else if strings.Contains(mainExpressionBlock, "Meta.Imports") {
		return "Headers"
	}
	return "Unknown"
}

// CheckTemplateType checks the tag type
func CheckTemplateType(template []byte) string {
	reType := "Variable"
	if bytes.Contains(template, []byte("-{")) {
		return "Tag"
	} else if bytes.Contains(template, []byte("={")) {
		return "SubTag"
	}
	return reType
}

// InterpretConjoinments brings in the nested conjoinments an inserts the user defined meta
func InterpretConjoinments(template []byte, conjoinments []CompiledConjoinment, meta MetaDict) []byte {
	keyBase := "Conjoinments."
	for _, conjoinment := range conjoinments {
		conjoinment.InsertMeta(meta)
		// TODO BUILD SUPPORT FOR PLURAL TYPE CONJOINMENTS WITH MORE THAN ONE CONTENT BYTE ARRAY
		if conjoinment.Plural {
			fmt.Println("PLURAL CHECK: ")
		} else {
			if len(conjoinment.Contents) > 0 {
				newContent := conjoinment.Contents[0]
				checkStrA := "%" + keyBase + conjoinment.Key + "%"
				checkStrB := "&" + keyBase + conjoinment.Key + "&"
				replaceStr := "[|" + keyBase + conjoinment.Key + "|]"
				if strings.Contains(string(template), checkStrA) {
					// A: Remove newlines from the interpolated fragment.
					replaceStr = "[|%" + keyBase + conjoinment.Key + "%|]"
					newContent = inlineConjoinmentNewLines(newContent)
				} else if strings.Contains(string(template), checkStrB) {
					// B: Remove leading new lines handler
					replaceStr = "[|&" + keyBase + conjoinment.Key + "&|]"
					// 1: Parse and store Imports Tag
					importTags := parseMainTags(newContent)
					if len(importTags) > 0 {
						newContent = bytes.Replace(newContent, importTags[0], []byte(""), -1)
					}
					newContent = []byte(TrimPrefixes(string(newContent), "\n"))
					// 2: Append stored Imports Tag to beginning of newContent
					if len(importTags) > 0 {
						newContent = append(importTags[0], newContent...)
					}
				}
				template = bytes.Replace(template, []byte(replaceStr), newContent, -1)
			}
		}
	}
	return template
}

// reformatForStr
// for Count in Count(Conjoinments.Procedures)
func reformatForStr(forStr string, strType string) string {
	reStr := forStr
	if strType == "count" {
		reStr = strings.Replace(strings.Replace(reStr, "Count(", "", -1), ")", "", -1)
	}
	return reStr
}

// checkForStatementString
func checkForStatementString(forStr string) (string, string) {
	reType := "standard"
	if strings.Contains(forStr, "Count(") {
		forStr = reformatForStr(forStr, "count")
		return forStr, "count"
	}
	return forStr, reType
}

// countArray
func countArray(inInt interface{}) (string, interface{}) {
	reCount := []int{}
	iterVars := reflect.ValueOf(inInt)
	for in := 0; in < iterVars.Len(); in++ {
		reCount = append(reCount, in+1)
	}
	return "Array-Number", reCount
}

// varLength
func varLength(inInt interface{}) int {
	iterVars := reflect.ValueOf(inInt)
	return iterVars.Len()
}

/*
==========================
LUCID VARIABLE STRUCTURES
==========================
*/

// WorkingVariable
type WorkingVariable struct {
	VarName  string
	VarType  string
	VarValue interface{}
}

// newWorkingVariable
func newWorkingVariable(varName string, varType string, varValue interface{}) *WorkingVariable {
	return &(WorkingVariable{varName, varType, varValue})
}

// WorkingVariables
type WorkingVariables struct {
	WorkingVariables []*WorkingVariable
}
