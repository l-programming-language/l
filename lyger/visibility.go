package Lyger

import (
	"fmt"

	"github.com/l-programming-language/l/lerrors"
)

type declarationVisibility uint8

const (
	visibilityDefault declarationVisibility = iota
	visibilityPublic
	declarationVisibilityCount
)

const publicVisibilityRole = "visibility.public"

var declarationVisibilityNames = [declarationVisibilityCount]string{
	visibilityDefault: "default",
	visibilityPublic:  "public",
}

func (v declarationVisibility) String() string {
	if v >= declarationVisibilityCount {
		return fmt.Sprintf("declarationVisibility(%d)", v)
	}
	return declarationVisibilityNames[v]
}

func (l *lexeme) declarationVisibility() (declarationVisibility, bool) {
	if l == nil || (l.POS != posAdjective.String() && l.POS != posAdverb.String()) || l.Type != publicVisibilityRole {
		return visibilityDefault, false
	}
	return visibilityPublic, true
}

func (t *token) declarationVisibility() (declarationVisibility, bool) {
	if t == nil || t.mem == nil || t.mem.Type != storageLexeme {
		return visibilityDefault, false
	}
	return t.mem.Lexeme.declarationVisibility()
}

func valueModifiers(tokens []*token) []*token {
	var values []*token
	for _, tok := range tokens {
		if _, visibility := tok.declarationVisibility(); !visibility {
			values = append(values, tok)
		}
	}
	return values
}

// Each clause owns its modifiers. Descendant clauses are removed before matching
// declaration tokens, so publishing a declaration never publishes its siblings.
// This also works before descriptor/alias values have been constructed.
func (sg *statementGram) declarationVisibilities() (map[*token]declarationVisibility, error) {
	result := make(map[*token]declarationVisibility)
	err := sg.walkRegistration(func(_ registrationSite, own []*token, _ bool) error {
		var target *token
		var modifiers []*token
		count := 0
		for _, tok := range own {
			if tok.isDeclare() {
				target, count = tok, count+1
				result[tok] = visibilityDefault
			}
			if _, ok := tok.declarationVisibility(); ok {
				modifiers = append(modifiers, tok)
			}
		}
		if len(modifiers) > 0 {
			if count != 1 {
				return lerrors.Syntax.Errorf("visibility modifier %q requires exactly one declaration in its clause", modifiers[0].rawBytes)
			}
			result[target] = visibilityPublic
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (sg *statementGram) declarationVisibility(target *token) (declarationVisibility, error) {
	visibilities, err := sg.declarationVisibilities()
	if err != nil {
		return visibilityDefault, err
	}
	visibility, exists := visibilities[target]
	if !exists {
		return visibilityDefault, lerrors.Internal.Errorf("visibility target is not a declaration in this statement")
	}
	return visibility, nil
}
