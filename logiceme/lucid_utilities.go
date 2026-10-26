// Author: Connor Sanders

package Logiceme

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
)

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

func checkVarReturn(kind string) []string {
	switch kind {
	case "Error":
		return []string{"Error", "Error"}
	case "Array-Number":
		return []string{"Array", "Number"}
	case "Array-String":
		return []string{"Array", "String"}
	}
	name, namespace := vowelName(kind), "Conjoinment."
	plural := name != kind
	if name == "Void" {
		name, namespace = strings.TrimSuffix(kind, "s"), "Meta."
		plural = name != kind
		switch name {
		case "Op", "Prop", "Func", "Recur":
		default:
			return []string{"Void", "Void"}
		}
	}
	// Preserve the interpreter's established Iteration item/array labels.
	if namespace == "Conjoinment." && name == "Iteration" {
		plural = !plural
	}
	cardinality := "Item"
	if plural {
		cardinality = "Array"
	}
	return []string{cardinality, namespace + name}
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

func determineExpressionOutput(expressions []string) string {
	body := expressions[2]
	if strings.Contains(body, "(/={") && strings.Contains(body, "}=/)") {
		return "SubTags"
	}
	// Keep the recognizer's precedence when more than one name is present.
	for _, name := range []string{"Declare", "Calculate", "Define", "Evaluate", "Describe", "Operate", "Instantiate", "Execute", "Iterate"} {
		if strings.Contains(body, "[|"+name+"(") && strings.Contains(body, ")|]") {
			return name
		}
	}
	if strings.Contains(body, "[|") && strings.Contains(body, "|]") {
		return "Variable"
	}
	if strings.Contains(body, "Meta.Imports") {
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

func checkForStatementString(forStr string) (string, string) {
	reType := "standard"
	if strings.Contains(forStr, "Count(") {
		forStr = reformatForStr(forStr, "count")
		return forStr, "count"
	}
	return forStr, reType
}

func countArray(inInt interface{}) (string, interface{}) {
	reCount := []int{}
	iterVars := reflect.ValueOf(inInt)
	for in := 0; in < iterVars.Len(); in++ {
		reCount = append(reCount, in+1)
	}
	return "Array-Number", reCount
}

func varLength(inInt interface{}) int {
	iterVars := reflect.ValueOf(inInt)
	return iterVars.Len()
}

type WorkingVariable struct {
	VarName  string
	VarType  string
	VarValue interface{}
}

func newWorkingVariable(varName string, varType string, varValue interface{}) *WorkingVariable {
	return &(WorkingVariable{varName, varType, varValue})
}

type WorkingVariables struct {
	WorkingVariables []*WorkingVariable
}
