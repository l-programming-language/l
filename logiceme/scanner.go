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
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// Regex
type Regex struct {
	String string
	Type   string
	Reg    *regexp.Regexp
}

// LoadRegex
func LoadRegex(reType string, reSub string) (Regex, error) {
	reString := ""
	// ||` and `||
	if reType == "Comment" {
		// Handles /*| and |*\
		reString = `\/\*\|((?:.*?\r?\n?)*)\|\*\\`
	} else if reType == "Array" {
		// Update to handle |[ and ]|
		reString = `\|\[((?:.*?\r?\n?)*)\]\|`
	} else if reType == "Hashmap" {
		// Update to handle |{ and }|
		reString = `\|\{((?:.*?\r?\n?)*)\}\|`
	} else if reType == "Tag" {
		// <TAG </TAG>
		reString = `\<` + reSub + `((?:.*?\r?\n?)*)\<\/` + reSub + `\>`
	} else if reType == "Variable" {
		// " ">
		reString = `\"((?:.*?\r?\n?)*)\"\>`
	} else if reType == "Word" {
		// <|W|> </W|>
		reString = `\<\|W\|\>((?:.*?\r?\n?)*)\<\/W\|\>`
	} else if reType == "Syllable" {
		// <|S|> </S|>
		reString = `\<\|S\|\>((?:.*?\r?\n?)*)\<\/S\|\>`
	} else if reType == "Count" {
		// <|S|>X(
		reString = `^(\<\|S\|\>[-]?\d\()`
	} else if reType == "Exponent" {
		// X)</S|>
		reString = `(\)[-]?\d\<\/S\|\>)$`
	} else if reType == "MetaDefs" {
		// {| |}
		reString = `\{\|((?:.*?\r?\n?)*)\|\}`
	} else if reType == "GlyphExps" {
		// ( )
		reString = `\(((?:.*?\r?\n?)*)\)`
	} else if reType == "Structs" {
		// Op{ }Op
		reString = reSub + `\{((?:.*?\r?\n?)*)\}` + reSub
	} else if reType == "StructArray" {
		// &[ and ]&
		reString = `\&\[((?:.*?\r?\n?)*)\]\&`
	} else if reType == "StructHashmap" {
		// #[ and ]#
		reString = `\#\[((?:.*?\r?\n?)*)\]\#`
	}
	regex := Regex{String: reString, Type: reType}
	regex.Reg = regexp.MustCompile(reString)
	return regex, nil
}

// Regex Parse
func (r *Regex) Parse(content []byte) [][]byte {
	return r.Reg.FindAll(content, -1)
}

// Regex Replace
func (r *Regex) Replace(content []byte, replacement []byte) []byte {
	return r.Reg.ReplaceAll(content, replacement)
}

// Tag
type Tag struct {
	Type      string
	Variables []string
	Recursive bool
}

// LoadTag
func LoadTag(tagString string) (Tag, error) {
	recur := false
	splitString := strings.Split(tagString, " | ")
	variables := strings.Split(splitString[1], ",")
	if splitString[2] == "TRUE" {
		recur = true
	}
	return Tag{Type: splitString[0], Variables: variables, Recursive: recur}, nil
}

// LoadTags
func LoadTags() ([]Tag, error) {
	tags := []Tag{}
	defaultTags := []string{
		"PROGRAM | NAME | FALSE",
		"CONTEXT |  | FALSE",
		"BOOK | NAME | FALSE",
		"SETTINGS |  | ",
		"DICTIONARIES |  | ",
		"DICTIONARY | NAME | TRUE",
		"WORDS |  | FALSE",
		"PHRASES |  | FALSE",
		"META |  | FALSE",
		"CHAPTER | NAME | TRUE",
		"TEXT | NAME | TRUE",
		"SENTENCE | NAME | FALSE",
		"STRUCTURE |  | ",
		"OUTLINE |  | ",
		"CONTENT |  | ",
		"JSON |  | ",
	}
	for _, dTag := range defaultTags {
		tag, err := LoadTag(dTag)
		if err != nil {
			return tags, err
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

// DefinitionStore
type DefinitionStore struct {
	StandIn []byte
	Raw     []byte
}

// RestoreDefinitionBytes
func RestoreDefinitionBytes(ds []DefinitionStore, contentBytes []byte) []byte {
	for _, definition := range ds {
		contentBytes = bytes.Replace(contentBytes, definition.StandIn, definition.Raw, 1)
	}
	return contentBytes
}

// ScannedProgram
type ScannedProgram struct {
	Name            string
	Tags            []Tag
	DefinitionStore []DefinitionStore
	Structure       [][]byte
	Books           []Book
	nativeSource    bool
}

// Scanner
type Scanner struct {
	Name            string
	ProgramBytes    []byte
	DefinitionStore []DefinitionStore
	Tags            []Tag
	Program         ScannedProgram
}

// StoreDefinitionBytes
func (s *Scanner) StoreDefinitionBytes() error {
	var framed bytes.Buffer
	cursor := 0
	for offset := 0; offset < len(s.ProgramBytes); {
		relative := bytes.IndexByte(s.ProgramBytes[offset:], '<')
		if relative < 0 {
			break
		}
		start := offset + relative
		if bytes.HasPrefix(s.ProgramBytes[start:], []byte("<JSON>")) {
			end, err := framedJSONEnd(s.ProgramBytes, start)
			if err != nil {
				return err
			}
			offset = end
			continue
		}
		if !bytes.HasPrefix(s.ProgramBytes[start:], []byte("<|W|>")) {
			end, err := framedTagEnd(s.ProgramBytes, start)
			if err != nil {
				return err
			}
			offset = end
			continue
		}
		end, err := framedWordEnd(s.ProgramBytes, start)
		if err != nil {
			return err
		}
		standIn := []byte("<|W|>" + GenerateUuid() + "</W|>")
		s.DefinitionStore = append(s.DefinitionStore, DefinitionStore{StandIn: standIn, Raw: bytes.Clone(s.ProgramBytes[start:end])})
		framed.Write(s.ProgramBytes[cursor:start])
		framed.Write(standIn)
		cursor = end
		offset = end
	}
	framed.Write(s.ProgramBytes[cursor:])
	s.ProgramBytes = framed.Bytes()
	return nil
}

// ParseDefinitionals
func ParseDefinitionals(content []byte, scanType string) [][]byte {
	regex, regErr := LoadRegex(scanType, "")
	if regErr != nil {
		panic(regErr)
	}
	return regex.Parse(content)
}

// LoadScanner
func LoadScanner(programName string) (Scanner, error) {
	programDir := "programs/" + programName
	tags, err := LoadTags()
	if err != nil {
		return Scanner{}, err
	}
	scanner := Scanner{Name: programDir, Tags: tags}
	programFile, progErr := ReadSysFile(programDir)
	if progErr != nil {
		return scanner, progErr
	}
	scanner.ProgramBytes = programFile
	return scanner, nil
}

// loadScannerBytes
func loadScannerBytes(scannerBytes []byte) (Scanner, error) {
	tags, err := LoadTags()
	if err != nil {
		return Scanner{}, err
	}
	scanner := Scanner{Tags: tags, ProgramBytes: scannerBytes}
	return scanner, nil
}

// StripCommentBytes
func (s *Scanner) StripCommentBytes() error {
	programBytes := s.ProgramBytes
	regex, err := LoadRegex("Comment", "")
	if err != nil {
		return err
	}
	s.ProgramBytes = regex.Replace(programBytes, []byte{})
	return nil
}

// DetermineTag
func DetermineTag(tags []Tag, tagType string) Tag {
	for _, dTag := range tags {
		if dTag.Type == tagType {
			return dTag
		}
	}
	return Tag{}
}

// processTagVar
func processTagVar(tag Tag, parsedBytes []byte) MetaDict {
	reVars := MetaDict{}
	cleanedBytes := bytes.Replace(parsedBytes, []byte(">"), []byte(``), -1)
	splitBytes := bytes.Split(cleanedBytes, []byte(`" "`))
	for ind, tVarKey := range tag.Variables {
		if len(splitBytes) > ind {
			tVarVal := splitBytes[ind]
			tVarVal = bytes.Replace(tVarVal, []byte(`"`), []byte(``), -1)
			reVars.Add(string(tVarKey), string(tVarVal), "~")

		} else {
			reVars.Add(string(tVarKey), "Null", "!")

		}
	}
	return reVars
}

// ParseTagVariables
func ParseTagVariables(tag Tag, contents [][]byte) ([]MetaDict, error) {
	reDicts := []MetaDict{}
	for _, content := range contents {
		if len(tag.Variables) > 0 && tag.Variables[0] != "" {
			varReg, rErr := LoadRegex("Variable", "")
			if rErr != nil {
				return reDicts, rErr
			}
			varBytes := varReg.Parse(content)
			processedVars := MetaDict{}
			if len(varBytes) > 0 {
				processedVars = processTagVar(tag, varBytes[0])
			}
			reDicts = append(reDicts, processedVars)
		}
	}
	return reDicts, nil
}

// ParseTags
func ParseTags(tags []Tag, content []byte, parseType string, tagType string) ([][]byte, []MetaDict, error) {
	reBytes := [][]byte{}
	tag := Tag{}
	reVars := []MetaDict{}
	if tagType != "" {
		tag = DetermineTag(tags, tagType)
	}
	if tag.Recursive {
		var varBytes [][]byte
		spans, err := framedStructuralTags(content, tagType)
		if err != nil {
			return nil, nil, err
		}
		for _, span := range spans {
			reBytes = append(reBytes, content[span.start:span.end])
			varBytes = append(varBytes, content[span.start:span.openEnd])
		}
		reVars, err = ParseTagVariables(tag, varBytes)
		return reBytes, reVars, err
	}
	regex, regErr := LoadRegex(parseType, tagType)
	if regErr != nil {
		fmt.Println("ParseTags - ERROR")
		fmt.Println(regErr)
		return reBytes, reVars, regErr
	}
	// Handle non-Recursive Tag
	reBytes = regex.Parse(content)
	reVars, regErr = ParseTagVariables(tag, reBytes)
	return reBytes, reVars, regErr
}

// ScanHashmapEntry
func ScanHashmapEntry(objBytes []byte) (MetaObj, error) {
	reObj := MetaObj{}
	var bKey []byte
	var bVal []byte
	var bType []byte
	if bytes.Contains(objBytes, []byte("KEY: ")) {
		bKey = bytes.Split(objBytes, []byte("|{KEY: "))[1]
		bKey = bytes.Split(bKey, []byte(", VALUE: "))[0]
		bKey = bytes.Replace(bKey, []byte(`"`), []byte(``), -1)
		bVal = bytes.Split(objBytes, []byte(", VALUE: "))[1]
		bVal = bytes.Split(bVal, []byte(", TYPE: "))[0]
		bType = bytes.Split(objBytes, []byte(", TYPE: "))[1]
		bType = bytes.Split(bType, []byte("}|"))[0]
		reObj.Key = string(bKey)
		reObj.Value = string(bVal)
		reObj.Type = string(bType)
	}
	return reObj, nil
}

// ScanHashmap
func ScanHashmap(metaDict MetaDict, tags []Tag, hashmapBytes []byte) (MetaDict, error) {
	objs, _, err := ParseTags(tags, hashmapBytes, "Hashmap", "")
	if err != nil {
		return metaDict, err
	}
	for _, obj := range objs {
		metaObj, hmErr := ScanHashmapEntry(obj)
		if hmErr != nil {
			return metaDict, hmErr
		}
		metaDict.Add(metaObj.Key, metaObj.Value, metaObj.Type)
	}
	return metaDict, nil
}

// ScanDefinitions
func ScanDefinitions(obj []byte) (MetaDict, error) {
	cleanedObj := bytes.Replace(obj, []byte("|[ "), []byte(""), -1)
	cleanedObj = bytes.Replace(cleanedObj, []byte(" ]|"), []byte(""), -1)
	cleanedObj = bytes.Replace(cleanedObj, []byte(" ], ["), []byte("],["), -1)
	reDict := MetaDict{}
	sObj := string(cleanedObj)
	if strings.Contains(sObj, "],") {
		for ind, s := range strings.Split(sObj, "],") {
			cleanedStr := s + "]"
			reDict.Add(strconv.Itoa(ind), cleanedStr, "Definition")
		}
		return reDict, nil
	}
	reDict.Add("0", sObj, "Definition")
	return reDict, nil
}

// ScanArray
func ScanArray(tags []Tag, arrayBytes []byte, scanType string) (MetaDict, error) {
	reMeta := MetaDict{}
	arrayObjects, _, err := ParseTags(tags, arrayBytes, "Array", "")
	if err != nil {
		return reMeta, err
	}
	for _, arrayObject := range arrayObjects {
		if !strings.Contains(string(arrayObject), "|{") {
			if scanType == "Definitions" {
				return ScanDefinitions([]byte(arrayObject))
			}
		} else {
			reMeta, err = ScanHashmap(reMeta, tags, arrayObject)
			if err != nil {
				return reMeta, err
			}
		}
	}
	return reMeta, nil
}

// ScanJSON
func ScanJSON(_ []Tag, content []byte) ([][]byte, error) {
	reBytes := [][]byte{}
	for {
		start := bytes.Index(content, []byte("<JSON"))
		end := bytes.Index(content, []byte("</JSON>"))
		if end >= 0 && (start < 0 || end < start) {
			return reBytes, lerrors.Syntax.Errorf("JSON record %d has no opening tag", len(reBytes)+1)
		}
		if start < 0 {
			return reBytes, nil
		}
		if !bytes.HasPrefix(content[start:], []byte("<JSON>")) {
			return reBytes, lerrors.Syntax.Errorf("JSON record %d has an invalid opening tag", len(reBytes)+1)
		}
		body := content[start+len("<JSON>"):]
		// Let the JSON decoder find the end of the value, so tag-like text
		// inside a JSON string is not mistaken for a closing markup tag.
		decoder := json.NewDecoder(bytes.NewReader(body))
		var raw json.RawMessage
		if err := decoder.Decode(&raw); err != nil {
			return reBytes, lerrors.Syntax.Errorf("JSON record %d: %w", len(reBytes)+1, err)
		}
		rest := bytes.TrimSpace(body[decoder.InputOffset():])
		if !bytes.HasPrefix(rest, []byte("</JSON>")) {
			return reBytes, lerrors.Syntax.Errorf("JSON record %d has no closing tag", len(reBytes)+1)
		}
		reBytes = append(reBytes, bytes.TrimSpace(body[:decoder.InputOffset()]))
		content = rest[len("</JSON>"):]
	}
}

// determineJSONSliceType
func determineJSONSliceType(content []byte, parentTag string) string {
	if parentTag == "Outline" {
		var record map[string]json.RawMessage
		if err := json.Unmarshal(content, &record); err != nil {
			return "Meta"
		}
		for _, key := range []string{"AppMeta", "Integrations", "Resources", "Datasources", "Services", "Classes", "Objects"} {
			if _, ok := record[key]; ok {
				return key
			}
		}
	}
	return "Meta"
}

// ScanOutline
func ScanOutline(tags []Tag, outlineBytes []byte) (Outline, error) {
	reOutline := Outline{}
	// Outline
	jsonSlice, setErr := ScanJSON(tags, outlineBytes)
	if setErr != nil {
		return reOutline, setErr
	}
	counts := map[string]int{}
	for _, jsonBytes := range jsonSlice {
		var keys map[string]json.RawMessage
		if err := json.Unmarshal(jsonBytes, &keys); err != nil {
			return reOutline, lerrors.Syntax.Errorf("outline: %w", err)
		}
		if _, present := keys["Features"]; present {
			return reOutline, lerrors.InvalidInput.Errorf("Features were removed from AML; migrate Objects and shared Attributes to Classes and explicit compositions")
		}
		jsonType := determineJSONSliceType(jsonBytes, "Outline")
		counts[jsonType]++
		if jsonType == "AppMeta" {
			meta, err := ParseMetaJSON("AppMeta", string(jsonBytes))
			if err != nil {
				return reOutline, fmt.Errorf("outline AppMeta: %w", err)
			}
			reOutline.Meta = meta
			continue
		}
		var record Outline
		if err := json.Unmarshal(jsonBytes, &record); err != nil {
			return reOutline, lerrors.Syntax.Errorf("outline %s: %w", jsonType, err)
		}
		switch jsonType {
		case "Integrations":
			reOutline.Integrations = record.Integrations
		case "Resources":
			reOutline.Resources = record.Resources
		case "Datasources":
			reOutline.Datasources = record.Datasources
		case "Services":
			reOutline.Services = record.Services
		case "Classes":
			reOutline.Classes = record.Classes
		case "Objects":
			reOutline.Objects = record.Objects
		}
	}
	if reOutline.Meta.GetVal("StructureVersion") == "2" {
		for _, kind := range []string{"AppMeta", "Services", "Classes", "Objects"} {
			if counts[kind] > 1 {
				return Outline{}, lerrors.InvalidInput.Errorf("duplicate outline %s records in StructureVersion 2", kind)
			}
		}
	}
	return reOutline, nil
}

// ScanMetaHash
func ScanMetaHash(hashBytes []byte) (MetaObj, error) {
	var result MetaObj
	body := strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(string(hashBytes)), "|{"), "}|")
	for _, field := range strings.Split(body, ", ") {
		key, value, ok := strings.Cut(field, ": ")
		if !ok {
			return result, lerrors.Syntax.Errorf("metadata field requires KEY, VALUE or TYPE: %q", field)
		}
		switch key {
		case "KEY":
			result.Key = value
		case "VALUE":
			result.Value = value
		case "TYPE":
			result.Type = value
		default:
			return result, lerrors.Syntax.Errorf("unknown metadata field %q", key)
		}
	}
	if result.Key == "" || result.Type == "" {
		return result, lerrors.Syntax.Errorf("metadata requires KEY and TYPE")
	}
	return result, nil
}

// ScanMeta
func ScanMeta(tags []Tag, arrayTags [][]byte) (MetaDict, error) {
	var reMeta MetaDict
	for _, arrayTag := range arrayTags {
		hashTags, _, hErr := ParseTags(tags, arrayTag, "Hashmap", "")
		if hErr != nil {
			return reMeta, hErr
		}
		for _, hashTag := range hashTags {
			mObj, mErr := ScanMetaHash(hashTag)
			if mErr != nil {
				return reMeta, mErr
			}
			reMeta.Add(mObj.Key, mObj.Value, mObj.Type)
		}
	}
	return reMeta, nil
}

// ScanMetaArray
func ScanMetaArray(tags []Tag, initBytes []byte, parseTags [][]byte) (MetaDict, error) {
	mBytes := initBytes
	for _, pTag := range parseTags {
		mBytes = bytes.Replace(mBytes, pTag, []byte(""), -1)
	}
	mTags, _, cErr := ParseTags(tags, mBytes, "Tag", "META")
	if cErr != nil {
		return MetaDict{}, cErr
	}
	if len(mTags) > 0 {
		aTags, _, aErr := ParseTags(tags, mTags[0], "Array", "")
		if aErr != nil {
			return MetaDict{}, aErr
		}
		return ScanMeta(tags, aTags)
	}
	return MetaDict{}, nil
}

// ParseSyllable
func ParseSyllable(syllableBytes []byte) ([]string, error) {
	source := strings.TrimSpace(string(syllableBytes))
	if !strings.HasPrefix(source, "<|S|>") || !strings.HasSuffix(source, "</S|>") {
		return nil, lerrors.Syntax.Errorf("syllable requires <|S|> and </S|> tags")
	}
	body := strings.TrimSuffix(strings.TrimPrefix(source, "<|S|>"), "</S|>")
	count, exponent, polarity := "1", "1", "false"
	if !strings.HasPrefix(body, "`") {
		open := strings.Index(body, "(`")
		closingIndex := strings.LastIndex(body, "`)")
		if open < 1 || closingIndex < open+2 {
			return nil, lerrors.Syntax.Errorf("syllable requires backtick-delimited content, optionally count(`content`)exponent")
		}
		count, exponent = body[:open], body[closingIndex+2:]
		if strings.HasPrefix(count, "-") {
			polarity = "true"
			count = strings.TrimPrefix(count, "-")
		}
		for _, modifier := range []string{count, strings.TrimPrefix(exponent, "-")} {
			if modifier == "" || strings.IndexFunc(modifier, func(r rune) bool { return r < '0' || r > '9' }) >= 0 {
				return nil, lerrors.Syntax.Errorf("syllable count and exponent must be integers")
			}
		}
		if _, err := strconv.Atoi(count); err != nil {
			return nil, lerrors.Syntax.Errorf("syllable count: %w", err)
		}
		if _, err := strconv.Atoi(exponent); err != nil {
			return nil, lerrors.Syntax.Errorf("syllable exponent: %w", err)
		}
		body = body[open+1 : closingIndex+1]
	}
	if len(body) < 3 || body[0] != '`' || body[len(body)-1] != '`' {
		return nil, lerrors.Syntax.Errorf("syllable requires nonempty backtick-delimited content")
	}
	return []string{count, body[1 : len(body)-1], exponent, polarity}, nil
}

// ScanSyllable
func (sp *ScannedProgram) ScanSyllable(syllableScan []byte) (Syllable, error) {
	reSyllable := Syllable{Name: string(syllableScan), Count: 1, Exponent: 1, Polarity: false}
	parsedSylStr, pErr := ParseSyllable(syllableScan)
	if pErr != nil {
		return reSyllable, pErr
	}
	if sp.nativeSource {
		body := strings.TrimPrefix(strings.TrimSpace(string(syllableScan)), "<|S|>")
		reSyllable.hasExplicitModifiers = !strings.HasPrefix(body, "`")
	}
	intCount, intErr := strconv.Atoi(parsedSylStr[0])
	if intErr != nil {
		return reSyllable, lerrors.Syntax.Wrap(intErr, "")
	}
	reSyllable.Count = intCount
	intCount, intErr = strconv.Atoi(parsedSylStr[2])
	if intErr != nil {
		return reSyllable, lerrors.Syntax.Wrap(intErr, "")
	}
	reSyllable.Name = parsedSylStr[1]
	reSyllable.Exponent = intCount
	if parsedSylStr[3] == "true" {
		reSyllable.Polarity = true
	}
	return reSyllable, nil
}

// ScanWords
func (sp *ScannedProgram) ScanWords(tags []Tag, wordScan []byte) ([]Word, error) {
	reWords := []Word{}
	scanArray, sErr := ScanArray(tags, wordScan, "Definitions")
	scanWords := ParseDefinitionals(wordScan, "Word")
	if sErr != nil {
		return reWords, sErr
	}
	for _, scanned := range scanArray.Metas {
		reWord := Word{Meta: MetaDict{}}
		reWord.Meta.Add(scanned.Key, scanned.Value, scanned.Type)
		reWords = append(reWords, reWord)
	}
	for _, scanWord := range scanWords {
		reWord := Word{Meta: MetaDict{}}
		syllables := []Syllable{}
		replaceBytes := RestoreDefinitionBytes(sp.DefinitionStore, scanWord)
		scanSyllables := ParseDefinitionals(replaceBytes, "Syllable")
		for _, scanSyllable := range scanSyllables {
			syllable, syErr := sp.ScanSyllable(scanSyllable)
			if syErr != nil {
				return reWords, syErr
			}
			syllables = append(syllables, syllable)
		}
		reWord.Syllables = syllables
		reWords = append(reWords, reWord)
	}
	return reWords, nil
}

// ScanSentence
func (sp *ScannedProgram) ScanSentence(tags []Tag, sentenceScan []byte) (Sentence, error) {
	meta, mErr := ScanMetaArray(tags, sentenceScan, [][]byte{})
	if mErr != nil {
		return Sentence{}, mErr
	}
	reSentence := Sentence{Meta: meta, Words: []Word{}, Tensor: Tensor{}}
	mBytes, _, pErr := ParseTags(tags, sentenceScan, "Tag", "META")
	if pErr != nil {
		return Sentence{}, pErr
	}
	if len(mBytes) > 0 {
		sentenceScan = bytes.Replace(sentenceScan, mBytes[0], []byte(""), -1)
	}
	wordArray, err := sp.ScanWords(tags, sentenceScan)
	if err != nil {
		return reSentence, err
	}
	reSentence.Words = wordArray
	return reSentence, nil
}

// ScanText scans direct members, retaining their interleaved content order.
func (sp *ScannedProgram) ScanText(tags []Tag, textTag []byte) (Text, error) {
	var result Text
	members, context, err := sp.scanSourceMembers(tags, textTag, "TEXT")
	if err != nil {
		return result, err
	}
	result.Context = context
	for _, member := range members {
		if member.text != nil {
			result.Members = append(result.Members, sourceMember{Kind: "Class", Index: len(result.SubTexts)})
			result.SubTexts = append(result.SubTexts, *member.text)
		} else {
			result.Members = append(result.Members, sourceMember{Kind: "Object", Index: len(result.Sentences)})
			result.Sentences = append(result.Sentences, *member.sentence)
		}
	}
	return result, nil
}

func (sp *ScannedProgram) ScanChapter(tags []Tag, chapterTag []byte) (Chapter, error) {
	var result Chapter
	members, context, err := sp.scanSourceMembers(tags, chapterTag, "CHAPTER")
	if err != nil {
		return result, err
	}
	result.Context = context
	for _, member := range members {
		if member.text != nil {
			result.Members = append(result.Members, sourceMember{Kind: "Class", Index: len(result.Texts)})
			result.Texts = append(result.Texts, *member.text)
		} else {
			result.Members = append(result.Members, sourceMember{Kind: "Object", Index: len(result.Sentences)})
			result.Sentences = append(result.Sentences, *member.sentence)
		}
	}
	return result, nil
}

// ScanChapter
func (sp *ScannedProgram) ScanChapters(tags []Tag, chaptersTag []byte) ([]Chapter, error) {
	reChapters := []Chapter{}
	chapTags, vars, cErr := ParseTags(tags, chaptersTag, "Tag", "CHAPTER")
	if cErr != nil {
		return reChapters, cErr
	}
	for ind, chapTag := range chapTags {
		chapter, err := sp.ScanChapter(tags, chapTag)
		if err != nil {
			return reChapters, err
		}
		if len(vars) > ind {
			chapter.Name = vars[ind].Get("NAME").Value
		}
		reChapters = append(reChapters, chapter)
	}
	return reChapters, nil
}

// ScanBook
func (sp *ScannedProgram) ScanBook(bookTag []byte) (Book, error) {
	reBook := Book{Program: bookTag}
	settingsTags, _, sErr := ParseTags(sp.Tags, bookTag, "Tag", "SETTINGS")
	contentTags, _, coErr := ParseTags(sp.Tags, bookTag, "Tag", "CONTENT")
	outlineTags, _, oErr := ParseTags(sp.Tags, bookTag, "Tag", "OUTLINE")
	if oErr != nil {
		return reBook, oErr
	} else if sErr != nil {
		return reBook, sErr
	} else if coErr != nil {
		return reBook, coErr
	}
	if len(contentTags) > 0 {
		reChaps, cErr := sp.ScanChapters(sp.Tags, contentTags[0])
		if cErr != nil {
			return reBook, cErr
		}
		reBook.Content = reChaps
	}
	if len(settingsTags) > 0 {
		sSettings, setErr := ScanJSON(sp.Tags, settingsTags[0])
		if setErr != nil {
			return reBook, setErr
		}
		var bookMeta MetaDict
		for _, sSet := range sSettings {
			var metaObj MetaObj
			if err := json.Unmarshal([]byte(sSet), &metaObj); err != nil {
				return reBook, lerrors.Syntax.Errorf("book settings: %w", err)
			}
			bookMeta.Add(metaObj.Key, metaObj.Value, metaObj.Type)
		}
		reBook.Meta = bookMeta
	}
	if len(outlineTags) > 0 {
		// OUTLINE
		sOutline, outErr := ScanOutline(sp.Tags, outlineTags[0])
		if outErr != nil {
			return reBook, outErr
		}
		reBook.Outline = sOutline
	}
	return reBook, nil
}

// ScanStructure
func (sp *ScannedProgram) ScanStructure() ([]Book, error) {
	reBooks := []Book{}
	if len(sp.Structure) == 0 {
		return reBooks, lerrors.Syntax.Errorf("program requires a STRUCTURE tag")
	}
	bookTags, vars, pErr := ParseTags(sp.Tags, sp.Structure[0], "Tag", "BOOK")
	if pErr != nil {
		return reBooks, pErr
	}
	for ind, bookTag := range bookTags {
		reBook, sErr := sp.ScanBook(bookTag)
		if sErr != nil {
			return reBooks, sErr
		}
		if len(vars) > ind {
			reBook.Name = vars[ind].Get("NAME").Value
		}
		reBooks = append(reBooks, reBook)
	}
	return reBooks, nil
}

// ScanProgramBytes
func (sp *ScannedProgram) ScanProgramBytes() error {
	structureBooks, sErr := sp.ScanStructure()
	if sErr != nil {
		return sErr
	}
	sp.Books = structureBooks
	return nil
}

// Scan an e file
func (s *Scanner) Scan() error {
	scannedProgram := ScannedProgram{Name: s.Name, Tags: s.Tags}
	if err := s.StripCommentBytes(); err != nil {
		return err
	}
	if err := s.StoreDefinitionBytes(); err != nil {
		return err
	}
	scannedProgram.DefinitionStore = s.DefinitionStore
	// GetProgram Code goes here
	cleanedBytes := s.ProgramBytes
	structBytes, _, sErr := ParseTags(s.Tags, cleanedBytes, "Tag", "STRUCTURE")
	scannedProgram.Structure = structBytes
	if sErr != nil {
		return sErr
	}
	// BUILD SCANNED PROGRAM FROM STRUCTURE AND CONTEXT BYTES
	if err := scannedProgram.ScanProgramBytes(); err != nil {
		return err
	}
	s.Program = scannedProgram
	return nil
}
