package Lyger

import (
	"errors"
	"strconv"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// validateAuthoredName applies only to natural-language source words. Internal
// identities and template slot labels keep their separate dotted-name contract.
func validateAuthoredName(raw []byte) error {
	word := strings.TrimSuffix(string(raw), ",")
	if len(word) >= 2 && word[0] == '`' && word[len(word)-1] == '`' {
		return nil
	}
	if len(word) >= 2 && word[0] == '"' && word[len(word)-1] == '"' {
		label, err := strconv.Unquote(word)
		if err != nil {
			return nil // Declaration parsing owns malformed quoted labels.
		}
		word = label
	} else if _, err := strconv.ParseFloat(word, 64); err == nil || errors.Is(err, strconv.ErrRange) {
		// Overflow remains a provisional numeric error until bindings resolve.
		return nil
	}
	if strings.ContainsRune(word, '.') {
		return lerrors.Syntax.Errorf("dot-qualified authored names are invalid: %q; use grammatical possessives or an explicitly scoped alias", word)
	}
	return nil
}
