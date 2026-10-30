package Lyger

import (
	"strconv"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// Grammar terms describe the compiler's closed grammar operations. Qualifiers,
// configured categories and case names remain text owned by the language.
type termKind uint8

const (
	termInvalid termKind = iota
	termUnknown
	termGrammar
	termLexical
	termType
	termIdentifier
	termCase
	termKindCount
)

var termKindNames = [termKindCount]string{
	termUnknown: "unknown", termGrammar: "grammar", termLexical: "lexical",
	termType: "type", termIdentifier: "identifier", termCase: "case",
}

func (k termKind) String() string {
	if k >= termKindCount {
		return "termKind(" + strconv.Itoa(int(k)) + ")"
	}
	return termKindNames[k]
}

type grammarFamily uint8

const (
	familyInvalid grammarFamily = iota
	familyCondition
	familyConditional
	familyValue
	familySubject
	familyObject
	familyPhrase
	familyClause
	familyPredicate
	familyStatement
	grammarFamilyCount
)

var grammarFamilies = [grammarFamilyCount]struct {
	name  string
	arity int
}{
	familyCondition: {"condition", 1}, familyConditional: {"conditional", 2},
	familyValue: {"value", 1}, familySubject: {"subject", 1}, familyObject: {"object", 1},
	familyPhrase: {"phrase", 2}, familyClause: {"clause", 3},
	familyPredicate: {"predicate", 2}, familyStatement: {"statement", 3},
}

func (f grammarFamily) String() string {
	if f >= grammarFamilyCount {
		return "grammarFamily(" + strconv.Itoa(int(f)) + ")"
	}
	return grammarFamilies[f].name
}

func parseGrammarFamily(raw string) (grammarFamily, bool) {
	for family := familyInvalid + 1; family < grammarFamilyCount; family++ {
		if grammarFamilies[family].name == raw {
			return family, true
		}
	}
	return familyInvalid, false
}

func (f grammarFamily) NameArity() (int, bool) {
	if f == familyInvalid || f >= grammarFamilyCount {
		return 0, false
	}
	return grammarFamilies[f].arity, true
}

type grammarCaseMode uint8

const (
	// UNKNOWN rather than NONE: a clause's case is usually not known until build.
	grammarCaseInferred grammarCaseMode = iota
	grammarCaseNone
)

func (mode grammarCaseMode) name() (string, bool) {
	switch mode {
	case grammarCaseInferred:
		return "UNKNOWN", true
	case grammarCaseNone:
		return "NONE", true
	}
	return "", false
}

// grammarTyping borrows a node's fields; subtype inference remains on the node.
type grammarTyping struct{ sType, gType, cType *string }

func (typing grammarTyping) decode(raw string) error {
	term, err := decodeGrammarTerm(raw)
	if err != nil {
		return err
	}
	*typing.sType = term.Name
	return nil
}

func (typing grammarTyping) decompose(family grammarFamily, mode grammarCaseMode) bool {
	arity, valid := family.NameArity()
	caseName, known := mode.name()
	parts := strings.Split(*typing.sType, ".")
	if !valid || !known || len(parts) < arity {
		return false
	}
	*typing.gType, *typing.cType = strings.Join(parts[:arity], "."), caseName
	if len(parts) == arity+1 {
		*typing.cType = parts[arity]
	}
	return true
}

func (typing grammarTyping) load(raw string, family grammarFamily, mode grammarCaseMode) error {
	if err := typing.decode(raw); err != nil {
		return err
	}
	typing.decompose(family, mode) // A short predicate name retains its prior rule and case.
	return nil
}

type partOfSpeech uint8

const (
	posInvalid partOfSpeech = iota
	posNoun
	posVerb
	posAdjective
	posAdverb
	posPreposition
	posArticle
	posConjunction
	posEvaluation
	posPronoun
	partOfSpeechCount
)

var partOfSpeechNames = [partOfSpeechCount]string{
	posNoun: "noun", posVerb: "verb", posAdjective: "adjective", posAdverb: "adverb",
	posPreposition: "preposition", posArticle: "article", posConjunction: "conjunction",
	posEvaluation: "evaluation", posPronoun: "pronoun",
}

func (p partOfSpeech) String() string {
	if p >= partOfSpeechCount {
		return "partOfSpeech(" + strconv.Itoa(int(p)) + ")"
	}
	return partOfSpeechNames[p]
}

func parsePartOfSpeech(raw string) (partOfSpeech, bool) {
	for pos := posInvalid + 1; pos < partOfSpeechCount; pos++ {
		if partOfSpeechNames[pos] == raw {
			return pos, true
		}
	}
	return posInvalid, false
}

// GrammarKind is partial: lexical pronouns participate in source binding, but
// the current grammar configuration does not admit PRONOUN rule elements.
func (p partOfSpeech) GrammarKind() (termKind, bool) {
	switch p {
	case posNoun, posVerb, posAdjective, posAdverb, posPreposition, posArticle, posConjunction, posEvaluation:
		return termLexical, true
	}
	return termInvalid, false
}

// AcceptsPossessive applies only to repeated adjective/adverb association.
func (p partOfSpeech) AcceptsPossessive() bool { return p == posAdjective || p == posAdverb }

type typePattern uint8

const (
	patternInvalid typePattern = iota
	patternText
	patternNumber
	patternTruth
	patternNull
	patternType
	patternList
	patternDictionary
	typePatternCount
)

var typePatterns = [typePatternCount]struct{ name, tokenType string }{
	patternText: {"`~`", "primitive.text"}, patternNumber: {"%~%", "primitive.number"},
	patternTruth: {"?~?", "primitive.truth"}, patternNull: {"!~!", "primitive.null"},
	patternType: {"#~#", "type"}, patternList: {"&~&", "type.iterative.list"},
	patternDictionary: {"{~}", "type.iterative.dictionary"},
}

func (p typePattern) String() string {
	if p >= typePatternCount {
		return "typePattern(" + strconv.Itoa(int(p)) + ")"
	}
	return typePatterns[p].name
}

func (p typePattern) TokenType() string {
	if p >= typePatternCount {
		return ""
	}
	return typePatterns[p].tokenType
}

type grammarTerm struct {
	Raw, Name, Qualifier         string
	Kind                         termKind
	Family                       grammarFamily
	POS                          partOfSpeech
	Pattern                      typePattern
	Multiple, Optional, Declared bool
}

type lexicalMatchMode uint8

const (
	lexicalMatchClass lexicalMatchMode = iota
	lexicalMatchQualified
)

// typeToken associates a grammar structure's default typed token field.
func (term grammarTerm) typeToken(available []*token) (*token, error) {
	if len(available) > 0 {
		actual := available[0].tType
		if term.Pattern == patternType {
			actual, _, _ = strings.Cut(actual, ".")
		}
		if actual == term.Pattern.TokenType() {
			return available[0], nil
		}
	}
	if !term.Optional {
		return nil, lerrors.Syntax.Errorf("Error: %s is not an optional type reference!", term.Pattern.TokenType())
	}
	return nil, nil
}

// Sequence identifiers omit declaration equality; lexicalToken adds that scalar gate.
func (term grammarTerm) matchesLexical(tok *token, mode lexicalMatchMode) bool {
	if mode > lexicalMatchQualified {
		return false
	}
	if term.Kind == termIdentifier {
		// TODO: Consider possessive handling for identifier associations.
		return term.Declared || tok.defined || tok.getPOS(true) == "noun"
	}
	return tok.getPOS(true) == term.className() && (mode == lexicalMatchClass || term.Qualifier == "" || term.Qualifier == tok.getCase())
}

// lexicalToken associates the first available token with a scalar field.
func (term grammarTerm) lexicalToken(available []*token, mode lexicalMatchMode) (*token, error) {
	if mode > lexicalMatchQualified {
		return nil, lerrors.InvalidConfiguration.Errorf("unknown lexical match mode %d", mode)
	}
	if len(available) > 0 && (term.Kind != termIdentifier || term.Declared == available[0].isDeclare()) && term.matchesLexical(available[0], mode) {
		return available[0], nil
	}
	if !term.Optional {
		return nil, lerrors.Syntax.Errorf("Error: %s is not an optional type!", term.className())
	}
	return nil, nil
}

// associateLexicalToken retains the field when an optional term does not match.
func (term grammarTerm) associateLexicalToken(available []*token, dst **token, mode lexicalMatchMode) error {
	matched, err := term.lexicalToken(available, mode)
	if err == nil && matched != nil {
		*dst = matched
	}
	return err
}

// associateLexicalTokens extends existing associations and commits only on success.
func (term grammarTerm) associateLexicalTokens(available []*token, dst *[]*token, mode lexicalMatchMode) error {
	if mode > lexicalMatchQualified {
		return lerrors.InvalidConfiguration.Errorf("unknown lexical match mode %d", mode)
	}
	associations := *dst
	if term.Multiple {
		// Repeated adjectives/adverbs accept possessives; qualifiers do not restrict this prefix.
		for _, tok := range available {
			if tok.getPOS(true) != term.className() && (!term.POS.AcceptsPossessive() || !tok.isPossessive()) {
				break
			}
			associations = append(associations, tok)
		}
	} else if len(available) > 0 && term.matchesLexical(available[0], mode) {
		associations = append(associations, available[0])
	}
	if !term.Optional && len(associations) == 0 {
		return lerrors.Syntax.Errorf("Error: %s is not an optional type!", term.className())
	}
	*dst = associations
	return nil
}

// className projects the selected grammar category for token-path matching and
// existing diagnostics. ID and CASE have no part of speech or grammar family.
func (t grammarTerm) className() string {
	switch t.Kind {
	case termGrammar:
		return t.Family.String()
	case termLexical:
		return t.POS.String()
	case termIdentifier:
		return "id"
	case termCase:
		return "case"
	case termType:
		return t.Pattern.TokenType()
	}
	return ""
}

func parseGrammarTerm(raw string) (grammarTerm, error) {
	term := grammarTerm{Raw: raw, Kind: termUnknown}
	name := raw
	if strings.HasSuffix(name, "*") {
		term.Optional = true
		name = strings.TrimSuffix(name, "*")
	}
	if strings.HasPrefix(name, "[") && strings.HasSuffix(name, "]") {
		term.Multiple = true
		name = name[1 : len(name)-1]
	}
	if len(name) >= 2 && strings.HasPrefix(name, `"`) && strings.HasSuffix(name, `"`) {
		term.Declared = true
		name = name[1 : len(name)-1]
		if !strings.EqualFold(name, "ID") {
			return grammarTerm{}, lerrors.InvalidConfiguration.Errorf("reference %q: only ID can be declared", raw)
		}
	}
	for pattern := patternInvalid + 1; pattern < typePatternCount; pattern++ {
		if name == typePatterns[pattern].name {
			term.Name, term.Kind, term.Pattern = name, termType, pattern
			return term, nil
		}
	}
	parts := strings.Split(name, ".")
	for _, part := range parts {
		if !grammarName(part) {
			return grammarTerm{}, lerrors.InvalidConfiguration.Errorf("malformed grammar reference %q", raw)
		}
	}
	term.Name = strings.ToLower(name)
	base, qualifier, _ := strings.Cut(term.Name, ".")
	term.Qualifier = qualifier
	switch {
	case base == "id":
		term.Kind = termIdentifier
	case term.Name == "case":
		term.Kind = termCase
	default:
		if family, ok := parseGrammarFamily(base); ok {
			term.Kind, term.Family = termGrammar, family
		} else if pos, ok := parsePartOfSpeech(base); ok {
			if kind, valid := pos.GrammarKind(); valid {
				term.Kind, term.POS = kind, pos
			}
		}
	}
	return term, nil
}
