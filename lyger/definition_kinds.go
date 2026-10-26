package Lyger

import (
	"strconv"
	"strings"
)

// definitionArgKind classifies the built-in argument forms at interpretation
// boundaries. Public argument kinds and source slot types retain their strings.
type definitionArgKind uint8

const (
	definitionArgInvalid definitionArgKind = iota
	definitionArgText
	definitionArgNumber
	definitionArgTruth
	definitionArgNull
	definitionArgSyllable
	definitionArgSyllables
	definitionArgProps
	definitionArgTexts
	definitionArgDefinition
	definitionArgKindCount
)

var definitionArgKinds = [definitionArgKindCount]struct {
	name   string
	folded string
}{
	definitionArgText:       {"Text", "text"},
	definitionArgNumber:     {"Number", "number"},
	definitionArgTruth:      {"Truth", "truth"},
	definitionArgNull:       {"Null", "null"},
	definitionArgSyllable:   {"Syllable", "syllable"},
	definitionArgSyllables:  {"Syllables", "syllables"},
	definitionArgProps:      {"Props", "props"},
	definitionArgTexts:      {"Texts", "texts"},
	definitionArgDefinition: {"Definition", "definition"},
}

func (kind definitionArgKind) Valid() bool {
	return kind > definitionArgInvalid && kind < definitionArgKindCount
}

func (kind definitionArgKind) String() string {
	if kind >= definitionArgKindCount {
		return "definitionArgKind(" + strconv.Itoa(int(kind)) + ")"
	}
	return definitionArgKinds[kind].name
}

// Constructor binding recognizes canonical spelling without broadening source
// function and variadic slot acceptance.
func parseDefinitionArgKindExact(raw string) (definitionArgKind, bool) {
	for kind := definitionArgInvalid + 1; kind < definitionArgKindCount; kind++ {
		if raw == definitionArgKinds[kind].name {
			return kind, true
		}
	}
	return definitionArgInvalid, false
}

// Rendering historically compares strings.ToLower results. EqualFold accepts
// additional Unicode equivalents and would change reference-slot behavior.
func parseDefinitionArgKindFolded(raw string) (definitionArgKind, bool) {
	folded := strings.ToLower(raw)
	for kind := definitionArgInvalid + 1; kind < definitionArgKindCount; kind++ {
		if folded == definitionArgKinds[kind].folded {
			return kind, true
		}
	}
	return definitionArgInvalid, false
}

// DefinitionArgKind covers stored scalars currently accepted as source-call
// arguments. Other forms accepted by public definition binding do not imply
// that a stored value can be lowered into those forms.
func (k valueType) DefinitionArgKind() (definitionArgKind, bool) {
	switch k {
	case valueText:
		return definitionArgText, true
	case valueNumber:
		return definitionArgNumber, true
	case valueTruth:
		return definitionArgTruth, true
	default:
		return definitionArgInvalid, false
	}
}
