// Author: Connor Sanders

package Logiceme

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// MetaObj - Protected and Protections are defined as metaObj
type MetaObj struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
	Type  string `json:"type,omitempty"`
}

// MetaDictIndex ...
type MetaDictIndex struct {
	Keys    []string `json:"keys,omitempty"`
	Indexes []int    `json:"indexes,omitempty"`
}

// MetaDict ...
type MetaDict struct {
	Index MetaDictIndex `json:"index,omitempty"`
	Metas []*MetaObj    `json:"metas,omitempty"`
}

type metaGetReturn struct {
	metaVal   string
	metaType  string
	metaIndex int
}

// ParseMetaJSON decodes string metadata without losing its source order. Array
// values retain the segment language's &[...]& representation.
func ParseMetaJSON(keyStr string, jsonStr string) (MetaDict, error) {
	var reDict MetaDict
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal([]byte(jsonStr), &envelope); err != nil {
		return reDict, lerrors.Syntax.Errorf("metadata JSON: %w", err)
	}
	if envelope == nil {
		return reDict, lerrors.Syntax.Errorf("metadata JSON must be an object")
	}
	if wrapped, ok := envelope[keyStr]; keyStr != "" && ok {
		jsonStr = string(wrapped)
	}
	decoder := json.NewDecoder(strings.NewReader(jsonStr))
	start, err := decoder.Token()
	if err != nil {
		return reDict, lerrors.Syntax.Wrap(err, fmt.Sprintf("metadata %q must be an object", keyStr))
	}
	if start != json.Delim('{') {
		return reDict, lerrors.Syntax.Errorf("metadata %q must be an object", keyStr)
	}
	for decoder.More() {
		keyToken, err := decoder.Token()
		if err != nil {
			return MetaDict{}, lerrors.Syntax.Errorf("metadata key: %w", err)
		}
		key, ok := keyToken.(string)
		if !ok {
			return MetaDict{}, lerrors.Syntax.Errorf("metadata key must be a string")
		}
		var raw json.RawMessage
		if err := decoder.Decode(&raw); err != nil {
			return MetaDict{}, lerrors.Syntax.Errorf("metadata %q: %w", key, err)
		}
		var value string
		if len(raw) == 0 || raw[0] != '"' {
			return MetaDict{}, lerrors.Syntax.Errorf("metadata %q must be a string", key)
		}
		if err := json.Unmarshal(raw, &value); err != nil {
			return MetaDict{}, lerrors.Syntax.Errorf("metadata %q: %w", key, err)
		}
		valueType := "String"
		if strings.Contains(value, "&[") && strings.Contains(value, "]&") {
			valueType = "Array"
		}
		reDict.Add(key, value, valueType)
	}
	if _, err := decoder.Token(); err != nil {
		return MetaDict{}, lerrors.Syntax.Errorf("metadata object: %w", err)
	}
	return reDict, nil
}

// LoadMetaJSON preserves the original convenience API. Call ParseMetaJSON when
// invalid input must be reported; this form returns an empty dictionary on error.
func LoadMetaJSON(keyStr string, jsonStr string) MetaDict {
	reDict, err := ParseMetaJSON(keyStr, jsonStr)
	if err != nil {
		return MetaDict{}
	}
	return reDict
}

// LoadJSON into the MetaDict
func (r *MetaDict) LoadJSON(keyStr string, jsonStr string) {
	parsed, err := ParseMetaJSON(keyStr, jsonStr)
	if err != nil {
		return
	}
	for _, meta := range parsed.GetMetaData() {
		r.Add(meta.Key, meta.Value, meta.Type)
	}
}

func regParse(regStr string, contentStr string) [][]byte {
	reg := regexp.MustCompile(regStr)
	res := reg.FindAll([]byte(contentStr), -1)
	return res
}

func determineMetaType(varStr string) string {
	//Check if Dict
	// '{| |}'
	regStr := `\'\{\|((?:.*?\r?\n?)*)\|\}\'`
	check := regParse(regStr, varStr)
	if len(check) > 0 {
		return "MetaDict"
	}
	//Check if Array
	// '[| |]'
	regStr = `\'\[((?:.*?\r?\n?)*)\]\'`
	check = regParse(regStr, varStr)
	if len(check) > 0 {
		return "Array"
	}
	//Check if String
	// ~ ~
	regStr = `\~((?:.*?\r?\n?)*)\~`
	check = regParse(regStr, varStr)
	if len(check) > 0 {
		return "String"
	}
	//Check if Integer
	varStr = strings.Trim(varStr, "|")
	if _, err := strconv.Atoi(varStr); err == nil {
		return "Integer"
	}
	//Check if Decimal
	if strings.Contains(varStr, ".") {
		return "Decimal"
	}
	//Check if Boolean
	if varStr == "true" || varStr == "false" {
		return "Boolean"
	}
	//Check if Null
	if varStr == "!" {
		return "Null"
	}
	return "Invalid"
}

func cacheSubMetas(str string) (string, []string, []string) {
	subMetas := []string{}
	subs := []string{}
	regStr := `\'\{\|((?:.*?\r?\n?)*)\|\}\'`
	parsedSubMetas := regParse(regStr, str)
	processBytes := []byte(str)
	for ind, parsedSubMeta := range parsedSubMetas {
		subMetas = append(subMetas, string(parsedSubMeta))
		sInd := strconv.Itoa(ind)
		sub := "'{|" + sInd + "|}'"
		subs = append(subs, sub)
		processBytes = bytes.Replace(processBytes, parsedSubMeta, []byte(sub), -1)
	}
	reStr := string(processBytes)
	return reStr, subMetas, subs
}

func insertSubMetas(str string, subMetas []string, subs []string) string {
	for ind, sub := range subs {
		subMeta := subMetas[ind]
		str = strings.Replace(str, sub, subMeta, -1)
	}
	return str
}

func metaStrPP(metaStr string) string {
	var reStr string
	metaStr = strings.Split(metaStr, "'{|")[1]
	metaStr = strings.Split(metaStr, "|}'")[0]
	bal := 0
	reStr = "~{|"
	for ind, r := range metaStr {
		s := string(r)
		if s == "{" && string(metaStr[ind+1]) == "|" {
			if bal == 0 {
				s = "'{"
			}
			bal = bal + 1
		}
		if s == "}" && string(metaStr[ind-1]) == "|" {
			bal = bal - 1
			if bal == 0 {
				s = "}'"
			}
		}
		reStr = reStr + s
	}
	reStr = reStr + "|}~"
	return reStr
}

func parseLoadString(str string) MetaDict {
	reDict := MetaDict{}
	// ~{| |}~
	regString := `\~\{\|((?:.*?\r?\n?)*)\|\}\~`
	reg := regexp.MustCompile(regString)
	res := reg.FindAll([]byte(str), -1)
	for _, re := range res {
		sRe := strings.Split(string(re), "~{|")[1]
		sRe = strings.Split(sRe, "|}~")[0]
		cachedStr, sMetas, subs := cacheSubMetas(sRe)
		sItems := strings.Split(cachedStr, "|,|")
		for _, sItem := range sItems {
			sItem = insertSubMetas(sItem, sMetas, subs)
			//Check if MetaDict
			sKey := strings.Split(sItem, `" : `)[0]
			sKey = strings.Trim(sKey, `"`)
			if strings.Contains(sItem, `" : {|"`) {
				sVal := `"` + sKey + `" : `
				sVal = strings.Replace(sItem, sVal, "", 1)
				sType := determineMetaType(sVal)
				sVal = metaStrPP(sVal)
				reDict.Add(sKey, sVal, sType)
			} else {
				sVal := strings.Split(sItem, " : ")[1]
				sType := determineMetaType(sVal)
				if sType != "Invalid" {
					reDict.Add(sKey, sVal, sType)
				}
			}
		}
	}
	return reDict
}

// Load a pre-formatted string into the MetaDict
func (r *MetaDict) Load(metaString string) error {
	parsedDict := parseLoadString(metaString)
	r.Index = parsedDict.Index
	r.Metas = parsedDict.Metas
	return nil
}

// get Internal MetaData Method to get a metaVal, metaKey pair
func (r *MetaDict) get(metaKey string) *metaGetReturn {
	inKeys := r.Index.Keys
	metaVal := "NotFound"
	metaType := "NotFound"
	metaIndex := 0
	for ind, inKey := range inKeys {
		if inKey == metaKey {
			metaIndex = r.Index.Indexes[ind]
			metaVal = r.Metas[metaIndex].Value
			metaType = r.Metas[metaIndex].Type
			break
		}
	}
	return &(metaGetReturn{metaVal, metaType, metaIndex})
}

// Get MetaDict MetaObj
func (r *MetaDict) Get(metaKey string) *MetaObj {
	mData := r.get(metaKey)
	return &(MetaObj{metaKey, mData.metaVal, mData.metaType})
}

// GetDict returns a nested MetaDict from a MetaDict
func (r *MetaDict) GetDict(metaKey string) MetaDict {
	reDict := MetaDict{}
	mData := r.get(metaKey)
	if mData.metaType == "MetaDict" {
		if mData.metaVal != "Invalid" {
			reDict.Load(mData.metaVal) //nolint:errcheck // MetaDict.Load retains its public error signature but always returns nil.
		}
	}
	return reDict
}

// GetMetas Get All MetaObjs
func (r *MetaDict) GetMetas() []*MetaObj {
	return r.Metas
}

// GetMetaData Get All MetaObjs Data
func (r *MetaDict) GetMetaData() []MetaObj {
	reMeta := make([]MetaObj, len(r.Metas))
	for index, metaPointer := range r.Metas {
		reMeta[index] = *metaPointer
	}
	return reMeta
}

// GetVal Get Value
func (r *MetaDict) GetVal(metaKey string) string {
	return r.get(metaKey).metaVal
}

func determineStructType(structStr string) string {
	reStr := "Void"
	if len(structStr) < 6 {
		return reStr
	}
	fVal3 := structStr[:3]
	lVal3 := structStr[len(structStr)-3:]
	fVal6 := structStr[:6]
	lVal6 := structStr[len(structStr)-6:]
	if strings.Contains(fVal3, "Op{") && strings.Contains(lVal3, "}Op") {
		return "Op"
	} else if strings.Contains(fVal6, "Prop{") && strings.Contains(lVal6, "}Prop") {
		return "Prop"
	} else if strings.Contains(fVal6, "Func{") && strings.Contains(lVal6, "}Func") {
		return "Func"
	} else if strings.Contains(fVal6, "Recur{") && strings.Contains(lVal6, "}Recur") {
		return "Recur"
	}
	return reStr
}

func replaceSubsWithVars(content string, items [][]byte) (string, [][]string) {
	varMgmr := [][]string{}
	for _, it := range items {
		varUuid := GenerateUuid()
		content = strings.Replace(content, string(it), varUuid, -1)
		varEnt := []string{varUuid, string(it)}
		varMgmr = append(varMgmr, varEnt)
	}
	return content, varMgmr
}

func combineVarMgmrs(v1 [][]string, v2 [][]string) [][]string {
	v3 := [][]string{}
	v3 = append(v3, v1...)
	v3 = append(v3, v2...)
	return v3
}

func insertVarIntoVal(val string, varManager [][]string) string {
	for _, varEnt := range varManager {
		key := varEnt[0]
		varVal := varEnt[1]
		val = strings.Replace(val, key, varVal, -1)
	}
	return val
}

func cleanVal(val string) string {
	if strings.HasPrefix(val, "&[") && strings.HasSuffix(val, "]&") {
		return val[2 : len(val)-2]
	}
	return val
}

// ParseStruct returns an Hashmap as a slice of string slices from a preformatted string
func ParseStruct(structStr string) [][]string {
	fields := parseStructFields(structStr)
	for _, field := range fields {
		field[1] = cleanVal(field[1])
	}
	return fields
}

// Retain value framing so callers can distinguish metadata from source brackets.
func parseStructFields(structStr string) [][]string {
	structVector := [][]string{}
	structStr = strings.Trim(strings.Trim(structStr, "#["), "]#")
	structType := determineStructType(structStr)
	startTag := structType + "{"
	endTag := "}" + structType
	subArrayReg, _ := LoadRegex("StructArray", "")  //nolint:errcheck // LoadRegex returns a compiled built-in expression and a nil error.
	subHashReg, _ := LoadRegex("StructHashmap", "") //nolint:errcheck // LoadRegex returns a compiled built-in expression and a nil error.
	structs := subArrayReg.Parse([]byte(structStr))
	hashes := subHashReg.Parse([]byte(structStr))
	content, varMgmr := replaceSubsWithVars(structStr, structs)
	vMgmr := varMgmr
	content, varMgmr = replaceSubsWithVars(content, hashes)
	vMgmr = combineVarMgmrs(vMgmr, varMgmr)
	content = strings.Replace(content, startTag, "", -1)
	content = strings.Replace(content, endTag, "", -1)
	sContent := strings.Split(content, ", ")
	for _, sC := range sContent {
		sEnt := strings.Split(sC, ": ")
		if len(sEnt) > 1 {
			sVal := insertVarIntoVal(sEnt[1], vMgmr)
			sEntry := []string{sEnt[0], sVal}
			structVector = append(structVector, sEntry)
		}
	}
	return structVector
}

// ParseStructArray returns parsed and processed *MetaObj Arrays from a pre-formatted string
func ParseStructArray(arrayStr string) []*MetaObj {
	reMeta := []*MetaObj{}
	reArray := []string{}
	mArray := strings.Trim(strings.Trim(arrayStr, "&["), "]&")
	structType := determineStructType(mArray)
	structReg, _ := LoadRegex("Structs", structType) //nolint:errcheck // LoadRegex retains its public error signature but returns nil after compilation.
	structs := structReg.Parse([]byte(mArray))
	for _, sVal := range structs {
		reArray = append(reArray, string(sVal))
	}
	for ind, reArr := range reArray {
		indStr := strconv.Itoa(ind)
		mObj := MetaObj{Key: indStr, Value: reArr, Type: "Struct"}
		reMeta = append(reMeta, &mObj)
	}
	return reMeta
}

// Add MetaDict Value
func (r *MetaDict) Add(metaKey string, metaVal string, metaType string) string {
	if r.get(metaKey).metaVal != "NotFound" {
		return "AlreadyExists"
	}
	newMetaObj := &(MetaObj{metaKey, metaVal, metaType})
	r.Index.Keys = append(r.Index.Keys, metaKey)
	r.Index.Indexes = append(r.Index.Indexes, len(r.Metas))
	r.Metas = append(r.Metas, newMetaObj)
	return "Added"
}

// Combine Two MetaDicts
func (r *MetaDict) Combine(metaDict MetaDict) MetaDict {
	var vMeta MetaDict
	for _, rObj := range r.GetMetaData() {
		vMeta.SetAdd(rObj.Key, rObj.Value, rObj.Type)
	}
	for _, mObj := range metaDict.GetMetaData() {
		vMeta.SetAdd(mObj.Key, mObj.Value, mObj.Type)
	}
	return vMeta
}

// Set MetaDict Value
func (r *MetaDict) Set(metaKey string, metaVal string, metaType string) string {
	curMeta := r.get(metaKey)
	if curMeta.metaVal == "NotFound" {
		return "NotFound"
	} else if curMeta.metaVal == metaVal {
		return "Updated"
	}
	r.Metas[curMeta.metaIndex] = &(MetaObj{metaKey, metaVal, metaType})
	return "Updated"
}

// SetAdd MetaDict Value
func (r *MetaDict) SetAdd(metaKey string, metaVal string, metaType string) string {
	addre := r.Add(metaKey, metaVal, metaType)
	if addre == "AlreadyExists" {
		r.Set(metaKey, metaVal, metaType)
		return "Updated"
	}
	return "Added"
}

// ReadString updates a MetaDict with the contents of a pre-formatted string
func (r *MetaDict) ReadString(metaString string) error {
	metaStr := strings.Split(metaString, "[{")[1]
	metaStr = strings.Split(metaStr, "}]")[0]
	for _, metaString := range strings.Split(metaStr, "}, {") {
		metaObj := MetaObj{}
		for _, metaEntry := range strings.Split(metaString, "`, ") {
			splitEntry := strings.Split(metaEntry, ": `")
			if splitEntry[0] == "Key" {
				metaObj.Key = splitEntry[1]
			} else if splitEntry[0] == "Type" {
				metaObj.Type = splitEntry[1]
			} else if splitEntry[0] == "Value" {
				metaObj.Value = splitEntry[1]
			}
		}
		r.Add(metaObj.Key, metaObj.Type, metaObj.Value)
	}
	return nil
}
