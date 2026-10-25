package Lyger

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"strconv"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// token contains the parsed contents information for a lexical word of LStatement,
type token struct {
	rawBytes     []byte // []byte(key)
	tType        string // prepositions.consignment.associative.root / type / primitive,
	category     string // POS type (noun, verb, etc.),
	subCategory  string // vocab type, i.e. primitive, structure, applicative, etc...
	class        string // vocab sub-type, i.e. () (text, app, number, etc.), or structure type (list, etc...)
	abstract     bool   // concrete if false
	conjoined    bool   // if rawBytes ended with a comma
	defined      bool   // if true token is user defined word aka id, if false token is a built in word
	mem          *entry // raw data value
	gram         *grammar
	binding      *tokenBinding
	bindingError error
	scope        *lexicalScope
	rangeError   error // Provisional primitive failure until names resolve.
}

// isPossessive returns whether the token is possessive or now
func (t *token) isPossessive() bool {
	return strings.Contains(t.tType, ".possessive")
}

// isRawType verifies whether the token is a raw type or not
func (t *token) isRawType() bool {
	tMap := getTypeMap()
	if _, ok := tMap[string(t.rawBytes)]; ok {
		return true
	}
	return false
}

// restoreLiteral restores cached spans in the original token, preserving
// intervening bytes without rescanning replacements or decoding escapes.
func (t *token) restoreLiteral(cache map[string][]byte) error {
	if len(cache) == 0 {
		return nil
	}
	raw := t.rawBytes
	var output []byte
	copied := 0
	for i := 0; i < len(raw); {
		if raw[i] != '`' && raw[i] != '"' {
			i++
			continue
		}
		end, err := quotedEnd(raw, i)
		if err != nil {
			return fmt.Errorf("restore literal %q: %w", raw, err)
		}
		if raw[i] == '"' {
			i = end
			continue
		}
		text, found := cache[string(raw[i:end])]
		if !found {
			i = end
			continue
		}
		if copied == 0 {
			current, err := t.mem.value()
			if err != nil {
				return fmt.Errorf("restore literal %q: %w", raw, err)
			}
			if current.Type != valueText {
				return lerrors.Internal.Errorf("restore literal %q: cached token has type %s, want Text", raw, current.Type)
			}
		}
		part, err := loadPrimitive(text)
		if err != nil {
			return fmt.Errorf("restore literal %q: %w", raw, err)
		}
		if part.Type != valueText {
			return lerrors.Internal.Errorf("restore literal %q: cached value has type %s, want Text", raw, part.Type)
		}
		output = append(output, raw[copied:i]...)
		output = append(output, text...)
		copied = end
		i = end
	}
	if copied == 0 {
		return nil
	}
	output = append(output, raw[copied:]...)
	primitive, err := loadPrimitive(output)
	if err != nil {
		return fmt.Errorf("restore literal %q: %w", raw, err)
	}
	if primitive.Type != valueText {
		return lerrors.Internal.Errorf("restore literal %q: restored value has type %s, want Text", raw, primitive.Type)
	}
	t.rawBytes, t.mem.Primitive = output, primitive
	return nil
}

// getCase returns the case of the token, i.e. appropriative.active
func (t *token) getCase() string {
	splitType := strings.Split(t.tType, ".")
	subType := strings.Join(splitType[1:], ".")
	subType = strings.Replace(subType, ".root", "", 1)
	return subType
}

// getPOS parses part of speech from tType
func (t *token) getPOS(trim bool) string {
	if t.tType == "primitive.unknown" {
		return "unknown"
	}
	if t.tType == "nouns.NotFound.possessive" {
		return "unknown"
	}
	rawPOS := strings.Split(t.tType, ".")[0]
	if trim {
		return trimSuffix(rawPOS, "s")
	}
	return rawPOS
}

// isDeclare checks if token is conditional, i.e. if/else/otherwise
func (t *token) isConditional() (bool, string) {
	if t.getPOS(false) == "evaluations" {
		sStr := strings.Split(t.tType, ".")
		if len(sStr) > 2 {
			//  conditional (if) | exceptional (else) | alternative (otherwise)
			switch sStr[1] {
			case "conditional", "exceptional", "alternative":
				return true, sStr[1]
			default:
				return false, ""
			}
		}
	}
	return false, ""
}

// isDeclare checks if token is declarative, i.e. wrapped in ""
func (t *token) isDeclare() bool {
	rawStr := string(t.rawBytes)
	if len(rawStr) > 0 && rawStr[0:1] == `"` && rawStr[len(rawStr)-1:] == `"` {
		return true
	}
	return false
}

// getImplicitRoot returns the implicit root of a token if it's declarative
func (t *token) getImplicitRoot() string {
	if t.isDeclare() && bytes.Contains(t.rawBytes, []byte("|")) {
		sRoot := bytes.Split(t.rawBytes, []byte("|")) // TECH DEBT: WHEN PREFIXES INTRODUCED, ENSURE ROOT IS IN MIDDLE
		if len(sRoot) > 0 {
			return string(sRoot[0]) + `"`
		}
	}
	return string(t.rawBytes)
}

// loadEntry uses t.rawBytes and loads the data entry
func (t *token) loadEntry() error {
	t.rangeError = nil
	// 1) Check if token is a literal primitive data value (text, number, truth, ID)
	if bytes.HasSuffix(t.rawBytes, []byte(",")) {
		t.conjoined = true
		t.rawBytes = bytes.TrimSuffix(t.rawBytes, []byte(","))
	}
	prim, primErr := loadPrimitive(t.rawBytes)
	if primErr == nil {
		t.mem = &entry{Type: storagePrimitive, Primitive: prim}
		t.tType = "primitive." + strings.ToLower(prim.Type.String())
		return nil
	}
	// 2) Check if token is an explicit data type (Primitive, Applicative, Iterative, Lexical)
	tType, tCat := checkType(t.rawBytes)
	if tType != "Unknown" && tCat != storageUnknown {
		t.mem = initEntry(tType, tCat)
		t.tType = "type." + strings.ToLower(tCat.String()) + "." + strings.ToLower(tType)
		return nil
	}
	// 3) Lookup if token is a primitive lexeme
	l, lCat, lErr := t.gram.lex.lookup(t.rawBytes)
	if lErr == nil {
		t.mem = &entry{Type: storageLexeme, Lexeme: l}
		t.tType = lCat
		return nil
	}
	// 4) Else load an unknown primitive for now
	if errors.Is(primErr, strconv.ErrRange) {
		t.rangeError = primErr
	}
	t.mem = &entry{Type: storagePrimitive, Primitive: newPrimitive(valueUnknown)}
	t.tType = "primitive.unknown"
	return nil
}

// scan token and determine key, category/subCategory, class, load new lexeme using gram.lex.lookup()
func (t *token) scan() error {
	err := t.loadEntry()
	if err != nil {
		return err
	}
	return nil
}

// newToken takes token rawBytes and grammar and returns a pointer to a new token struct
func newToken(rawBytes []byte, gram *grammar) *token {
	newT := token{
		rawBytes: rawBytes,
		gram:     gram,
		defined:  false, // set to true if user defined downstream
	}
	return &newT
}

// lexicon structures the Lyger keyword lexicon
type lexicon struct {
	Nouns        map[string]lexeme `json:"nouns,omitempty"`
	Verbs        map[string]lexeme `json:"verbs,omitempty"`
	Adjectives   map[string]lexeme `json:"adjectives,omitempty"`
	Adverbs      map[string]lexeme `json:"adverbs,omitempty"`
	Prepositions map[string]lexeme `json:"prepositions,omitempty"`
	Articles     map[string]lexeme `json:"articles,omitempty"`
	Conjunctions map[string]lexeme `json:"conjunctions,omitempty"`
	Evaluations  map[string]lexeme `json:"evaluations,omitempty"`
}

// scanLex for a matching lexical entry for an inputted POS Type
func (l *lexicon) scanLex(tBytes []byte, lType string) (lexeme, string) {
	switch lType {
	case "Nouns":
		var isPossessive bool
		owner, possessive := possessiveOwner(string(tBytes))
		isPossessive, tBytes = possessive, []byte(owner)
		reLexeme, lCat := checkLexemeMap(tBytes, l.Nouns)
		if isPossessive && lCat != "NotFound" {
			lCat = lCat + ".possessive"
		}
		return reLexeme, lCat
	case "Verbs":
		return checkLexemeMap(tBytes, l.Verbs)
	case "Adjectives":
		return checkLexemeMap(tBytes, l.Adjectives)
	case "Adverbs":
		return checkLexemeMap(tBytes, l.Adverbs)
	case "Prepositions":
		return checkLexemeMap(tBytes, l.Prepositions)
	case "Articles":
		return checkLexemeMap(tBytes, l.Articles)
	case "Conjunctions":
		return checkLexemeMap(tBytes, l.Conjunctions)
	case "Evaluations":
		return checkLexemeMap(tBytes, l.Evaluations)
	default:
		return lexeme{}, "Error"
	}
}

// lookup checks an input token to determine if it represents a defined lexeme of the lexicon
func (l *lexicon) lookup(tBytes []byte) (*lexeme, string, error) {
	var reLexeme lexeme
	var lCat string
	lTypes := []string{
		"Nouns",
		"Verbs",
		"Adjectives",
		"Adverbs",
		"Prepositions",
		"Articles",
		"Conjunctions",
		"Evaluations",
	}
	tBytes = cleanTokenBytes(tBytes)
	for _, lType := range lTypes {
		reLexeme, lCat = l.scanLex(tBytes, lType)
		if lCat != "NotFound" {
			return &reLexeme, strings.ToLower(lType) + "." + lCat, nil
		} else if lCat == "Error" {
			fmt.Println("\nMATCH ERROR!!! THIS SHOULDN'T HAPPEN BTW" + string(tBytes))
			return &reLexeme, strings.ToLower(lType) + "." + lCat, lerrors.Internal.Errorf("%s", "unknown POS Type: "+lType)
		}
	}
	return &reLexeme, lCat, lerrors.UnresolvedReference.Errorf("NotFound")
}

func loadLexicon(resources fs.FS, lang string) (*lexicon, error) {
	path, err := languageDefinitionPath(lang, "vocabulary.json")
	if err != nil {
		return nil, fmt.Errorf("load vocabulary: %w", err)
	}
	var lex lexicon
	if err := readLanguageJSON(resources, path, &lex); err != nil {
		return nil, fmt.Errorf("load vocabulary: %w", err)
	}
	return &lex, nil
}
