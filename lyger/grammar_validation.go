package Lyger

import (
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"unicode"

	"github.com/l-programming-language/l/lerrors"
)

// readGrammarRules checks configuration references before parser alternatives
// can mistake an invalid optional rule for absent source syntax. This validates
// the grammar's reference graph, not the semantics of every language construct.
func readGrammarRules(resources fs.FS, path string) (map[string]grammeme, error) {
	var rules map[string]grammeme
	if err := readLanguageJSON(resources, path, &rules); err != nil {
		return nil, err
	}
	if err := validateGrammarReferences(rules); err != nil {
		return nil, fmt.Errorf("validate grammar %s: %w", path, err)
	}
	return rules, nil
}

func grammarName(name string) bool {
	if name == "" {
		return false
	}
	for index, char := range name {
		if char != '_' && !unicode.IsLetter(char) && (index == 0 || !unicode.IsDigit(char)) {
			return false
		}
	}
	return true
}

func validateGrammarReferences(rules map[string]grammeme) error {
	if len(rules) == 0 {
		return lerrors.InvalidConfiguration.Errorf("grammar contains no rules")
	}
	names := make([]string, 0, len(rules))
	for name := range rules {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		parts := strings.Split(name, ".")
		family, _ := parseGrammarFamily(parts[0])
		if size, valid := family.NameArity(); !valid || len(parts) != size || name != strings.ToLower(name) {
			return lerrors.InvalidConfiguration.Errorf("rule %q has an invalid grammar family or category", name)
		}
		for _, part := range parts {
			if !grammarName(part) {
				return lerrors.InvalidConfiguration.Errorf("rule %q has an invalid name", name)
			}
		}
		rule := rules[name]
		caseCount, err := validateGrammarSequence(rules, rule.Structure, true)
		if err != nil {
			return fmt.Errorf("rule %q structure: %w", name, err)
		}
		if caseCount > 1 || (caseCount == 0 && len(rule.Cases) > 0) || (caseCount > 0 && len(rule.Cases) == 0) {
			return lerrors.InvalidConfiguration.Errorf("rule %q has inconsistent CASE structure and alternatives", name)
		}
		cases := make([]string, 0, len(rule.Cases))
		for name := range rule.Cases {
			cases = append(cases, name)
		}
		sort.Strings(cases)
		for _, caseName := range cases {
			if !grammarName(caseName) || caseName != strings.ToLower(caseName) {
				return lerrors.InvalidConfiguration.Errorf("rule %q has an invalid case name %q", name, caseName)
			}
			if _, err := validateGrammarSequence(rules, rule.Cases[caseName], false); err != nil {
				return fmt.Errorf("rule %q case %q: %w", name, caseName, err)
			}
		}
	}
	return nil
}

func validateGrammarSequence(rules map[string]grammeme, sequence string, allowCase bool) (int, error) {
	if strings.TrimSpace(sequence) == "" {
		return 0, lerrors.InvalidConfiguration.Errorf("empty grammar sequence")
	}
	cases := 0
	for _, raw := range splitGramRule(sequence) {
		term, err := decodeGrammarTerm(raw)
		if err != nil {
			return 0, err
		}
		if term.Kind == termCase {
			if !allowCase || term.Multiple {
				return 0, lerrors.InvalidConfiguration.Errorf("CASE cannot repeat or occur inside an alternative")
			}
			cases++
			continue
		}
		if term.Kind == termType || term.Kind == termLexical || term.Kind == termIdentifier {
			continue
		}
		if err := validateGrammarReference(rules, term); err != nil {
			return 0, fmt.Errorf("reference %q: %w", raw, err)
		}
	}
	return cases, nil
}

func validateGrammarReference(rules map[string]grammeme, term grammarTerm) error {
	parts := strings.Split(term.Name, ".")
	family := term.Family
	size, valid := family.NameArity()
	if !valid || family == familyStatement {
		return lerrors.InvalidConfiguration.Errorf("unknown grammatical element")
	}
	if len(parts) < size {
		// Phrases/conditionals infer their category from source tokens. Clauses
		// also accept a category/form filter, such as CLAUSE.APPLICATIVE.
		if family == familyPredicate {
			return lerrors.InvalidConfiguration.Errorf("predicate requires an explicit category")
		}
		for candidate := range rules {
			candidateParts := strings.Split(candidate, ".")
			if candidateParts[0] != family.String() || len(candidateParts) != size {
				continue
			}
			if len(parts) == 1 {
				return nil
			}
			for _, category := range candidateParts[1:] {
				if category == parts[1] || (parts[1] == "root" && category == "imperative") {
					return nil
				}
			}
		}
		return lerrors.InvalidConfiguration.Errorf("dynamic grammar family has no configured target")
	}
	if len(parts) > size+1 {
		return lerrors.InvalidConfiguration.Errorf("too many grammar categories")
	}
	target := gramTypePreProc(strings.Join(parts[:size], "."))
	rule, found := rules[target]
	if !found {
		return lerrors.InvalidConfiguration.Errorf("rule %q is not configured", target)
	}
	if len(parts) == size+1 {
		if _, found := rule.Cases[parts[size]]; !found {
			return lerrors.InvalidConfiguration.Errorf("case %q is not configured on rule %q", parts[size], target)
		}
	}
	return nil
}
