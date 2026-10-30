package Lyger

import (
	"strconv"
	"strings"

	"github.com/l-programming-language/l/lerrors"
	"github.com/l-programming-language/l/utilities"
)

// determineGlyphType inputs a glyph definition and checks for the type
func determineGlyphType(defString string) string {
	if strings.HasPrefix(defString, "`[[") && strings.HasSuffix(defString, "]]`") {
		return "Sentence"
	} else if strings.HasPrefix(defString, "`[") && strings.HasSuffix(defString, "]`") {
		return "Word"
	} else {
		return "Syllable"
	}
}

func trimSuffix(s, suffix string) string {
	return strings.TrimSuffix(s, suffix)
}

// newCacheUuid
func newCacheUuid() string {
	cUuid := utilities.GenerateUuid()
	return "`" + cUuid + "`"
}

// inStringSlice checks if a string slice contains a string
func inStringSlice(chkSlice []string, chk string) bool {
	for _, c := range chkSlice {
		if c == chk {
			return true
		}
	}
	return false
}

// removeFileStamp inputs l bytes and removes file stamps
func removeFileStamp(lContent []byte) ([]byte, error) {
	par := newParser(lContent)
	return par.removeLineStamps()
}

// newLineStamp creates a new line stamp in bytes based on line number and file name $$HEAD<file.l:1>HEAD$$
func newLineStamp(lineNum int, fileName string) []byte {
	return []byte("$$HEAD<" + fileName + ":" + strconv.Itoa(lineNum) + ">HEAD$$")
}

// bytesToBool
func bytesToBool(b []byte) bool {
	if string(b) == "true" || string(b) == "?true?" {
		return true
	}
	return false
}

// checkType inputs a token byte slice and determines which (if any) explicit data type/category it is
func checkType(tBytes []byte) (string, storageKind) {
	primitives := []string{"null", "text", "number", "truth", "bytes", "thing", "pointer"}
	applicatives := []string{"app", "service", "class", "object", "attribute", "definition"}
	iteratives := []string{"list", "map", "dictionary"}
	lexemes := []string{"noun", "verb", "adverb", "adjective"}
	tStr := strings.ToLower(string(tBytes))
	if inStringSlice(primitives, tStr) {
		return strings.ToUpper(tStr[:1]) + tStr[1:], storagePrimitive
	}
	if inStringSlice(applicatives, tStr) {
		return strings.ToUpper(tStr[:1]) + tStr[1:], storageApplicative
	}
	if inStringSlice(iteratives, tStr) {
		return strings.ToUpper(tStr[:1]) + tStr[1:], storageIterative
	}
	if inStringSlice(lexemes, tStr) {
		return strings.ToUpper(tStr[:1]) + tStr[1:], storageLexeme
	}
	return "Unknown", storageUnknown
}

// isText checks if input bytes is text
func isText(tBytes []byte) bool {
	tStr := string(tBytes)
	if tStr != "" && tStr != "`" {
		if tStr[0:1] == "`" && tStr[len(tStr)-1:] == "`" {
			return true
		}
	}
	return false
}

// isTruth checks if input bytes is a bool
func isTruth(tBytes []byte) bool {
	if string(tBytes) == "false" || string(tBytes) == "true" {
		return true
	}
	return false
}

// isNum checks if input bytes is a number
func isNum(tBytes []byte) bool {
	_, err := strconv.ParseFloat(string(tBytes), 64)
	return err == nil
}

// isID checks if input bytes is an ID
func isID(tBytes []byte) bool {
	tStr := string(tBytes)
	if tStr != "" && tStr != `"` && !strings.Contains(tStr, " ") {
		if tStr[0:1] == `"` && tStr[len(tStr)-1:] == `"` {
			return true
		}
	}
	return false
}

// cleanTokenBytes inputs the byte slice of a token and cleans it, i.e. lower case, no spaces, etc.
func cleanTokenBytes(tBytes []byte) []byte {
	return []byte(strings.ToLower(string(tBytes)))
}

// processLexFix inputs a root string, postfix/prefix (fix) string, and type then returns a processed word string
func processLexFix(root string, fix string, fType string) string {
	dCount := strings.Count(fix, "-") - 1
	if len(root) >= dCount {
		if fType == "postfix" {
			root = root[:len(root)-dCount] + strings.Replace(fix, "-", "", -1)
		} else if fType == "prefix" {
			root = strings.Replace(fix, "-", "", -1) + root[dCount:]
		}
	}
	return root
}

// checkLexemeMap inputs token bytes and a lexeme POS Map and checks for a lexical match
func checkLexemeMap(tBytes []byte, lMap map[string]lexeme) (lexeme, string) {
	for key, ent := range lMap {
		isMatch, lType := ent.match(tBytes)
		if isMatch {
			lType = key + "." + lType
			return ent, lType
		}
	}
	return lexeme{}, "NotFound"
}

// tokenMatch inputs two token slices and returns whether there's a match and between where
func tokenMatch(primeSlice []*token, subSlice []*token) (bool, []int) {
	// Tokens identify source occurrences, so identical spelling is not identity.
	if len(subSlice) == 0 || len(subSlice) > len(primeSlice) {
		return false, []int{0, 0}
	}
	for start := 0; start+len(subSlice) <= len(primeSlice); start++ {
		match := true
		for offset, tok := range subSlice {
			if primeSlice[start+offset] != tok {
				match = false
				break
			}
		}
		if match {
			return true, []int{start, start + len(subSlice) - 1}
		}
	}
	return false, []int{0, 0}
}

// tokenReducer inputs two token slices and reduces the primeSlice by the subSlice if there's a match
func tokenReducer(primeSlice []*token, subSlice []*token) []*token {
	match, positions := tokenMatch(primeSlice, subSlice)
	if !match {
		return primeSlice
	}
	// Never overwrite another grammar node's shared token backing array.
	result := make([]*token, 0, len(primeSlice)-len(subSlice))
	result = append(result, primeSlice[:positions[0]]...)
	return append(result, primeSlice[positions[1]+1:]...)
}

// generalizePrepTokenType inputs a token's tType and removes postfix lexical type, i.e. '.root' or '.passive'
func generalizePrepTokenType(tType string) (string, string) {
	lexicalTypes := []string{"root", "plural", "active", "passive", "applicative", "object", "null", "alt"}
	for _, lt := range lexicalTypes {
		chkStr := "." + lt
		if strings.Contains(tType, chkStr) {
			genType := strings.Replace(tType, chkStr, "", -1)
			return genType, lt
		}
	}
	return tType, ""
}

// identifyConditionalSequences inputs a slice of tokens and uses 'else' and 'if' to identify conditional sequences
//
//	i.e. "if x is y and z is a, show x, else if x is a, show z, else exit with an error."
//		[[if, x, is, y, and, z, is, a, show, x], [else, if, x, is, a, show, z], [else, exit, with, an, error]]
func identifyConditionalSequences(tokens []*token) ([][]*token, error) {
	var reSequences [][]*token
	var curSequence []*token
	for ind, t := range tokens {
		if ind == 0 && t.tType != "evaluations.conditional.root" {
			return reSequences, lerrors.Syntax.Errorf("must begin conditional sequences with if")
		}
		if t.tType == "evaluations.exceptional.root" {
			reSequences = append(reSequences, curSequence)
			curSequence = []*token{t}
		} else {
			curSequence = append(curSequence, t)
			if ind == len(tokens)-1 {
				reSequences = append(reSequences, curSequence)
				curSequence = []*token{}
			}
		}
	}
	return reSequences, nil
}

// identifyConditionSequences inputs a slice of tokens and uses 'and' and 'or' to identify condition sequences
//
//	 i.e. "x is y and z is a or z is b, show x"
//			[[x, is, y], [z, is, a], [z, is, b]]	[and, or]	nil
func identifyConditionSequences(tokens []*token) ([][]*token, []*token, error) {
	var reSequences [][]*token
	var conjunctions []*token
	var curSequence []*token
	for ind, t := range tokens {
		if t.tType == "evaluations.conditional.root" || t.tType == "evaluations.exceptional.root" {
			return reSequences, conjunctions, lerrors.Syntax.Errorf("condition sequence cannot contain an if/else")
		}
		if t.getPOS(false) == "conjunctions" {
			if len(curSequence) == 0 {
				return reSequences, conjunctions, lerrors.Syntax.Errorf("cannot begin condition sequence with a conjunction")
			}
			conjunctions = append(conjunctions, t)
			reSequences = append(reSequences, curSequence)
			curSequence = []*token{}
		} else {
			if t.conjoined {
				curSequence = append(curSequence, t)
				reSequences = append(reSequences, curSequence)
				break
			} else {
				curSequence = append(curSequence, t)
				if ind == len(tokens)-1 {
					reSequences = append(reSequences, curSequence)
					curSequence = []*token{}
				}
			}
		}
	}
	return reSequences, conjunctions, nil
}

// isDeclarativeGrammar inputs a grammeme sType and determines if declarative for use in declaration memory
func isDeclarativeGrammar(sType string) bool {
	if strings.Contains(sType, ".declarative") || strings.Contains(sType, ".definitive") ||
		strings.Contains(sType, ".descriptive") {
		return true
	}
	return false
}

// checkActions inputs a slice of lexeme action sequence and determines sequence type and whether there's an error
func checkActions(actionSequence []*lexeme) (string, error) {
	seqType := "async"
	for ind, actionLex := range actionSequence {
		if actionLex.checkDescriptors("adverb", "progression.root") {
			if ind == len(actionSequence)-1 {
				seqType = "sync"
			} else {
				return "", lerrors.Syntax.Errorf("progressive adverb cannot be used in middle of an action sequence")
			}
		}
	}
	return seqType, nil
}
