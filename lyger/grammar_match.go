package Lyger

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// tokenSelection identifies source occurrences, never roles or spellings.
type tokenSelection map[*token]bool

func selectTokens(groups ...[]*token) tokenSelection {
	selected := make(tokenSelection)
	for _, group := range groups {
		selected.add(group...)
	}
	return selected
}

func (selected tokenSelection) add(tokens ...*token) {
	for _, tok := range tokens {
		selected[tok] = true
	}
}

func (selected tokenSelection) remaining(source []*token) []*token {
	result := make([]*token, 0, len(source))
	for _, tok := range source {
		if !selected[tok] {
			result = append(result, tok)
		}
	}
	return result
}

// A nested grammar consumes a prefix. Its parent owns the remaining suffix.
func grammarPrefix(tokens, remaining []*token) ([]*token, error) {
	count := len(tokens) - len(remaining)
	if count <= 0 {
		return nil, lerrors.Syntax.Errorf("grammar matched no tokens")
	}
	for i, tok := range remaining {
		if tokens[count+i] != tok {
			return nil, lerrors.Internal.Errorf("grammar associations are not a source prefix")
		}
	}
	return tokens[:count], nil
}

// buildGrammarNode keeps initialization and build failures on the same path.
// Parents commit the returned child only after both stages succeed.
func buildGrammarNode[T interface{ build() error }](node T, err error) (T, error) {
	if err != nil {
		return node, err
	}
	return node, node.build()
}

// buildGrammarChild installs one child only after initialization and parsing.
func buildGrammarChild[T interface{ build() error }](destination *T, initialize func(string, []*token, *grammar) (T, error), name string, tokens []*token, gram *grammar) error {
	node, err := buildGrammarNode(initialize(name, tokens, gram))
	if err == nil {
		*destination = node
	}
	return err
}

// Repeated children and their separators share one successful parse boundary.
func buildGrammarChildren[T any](destination *[]T, separators *[]*token, parse func(grammarTerm, []*token, *grammar) ([]T, []*token, error), term grammarTerm, tokens []*token, gram *grammar) error {
	children, conjunctions, err := parse(term, tokens, gram)
	if err == nil {
		*destination = append(*destination, children...)
		*separators = append(*separators, conjunctions...)
	}
	return err
}

// Conditional groups retain their source splitter and initialize every child
// before building any, so a rejected group cannot publish earlier children.
func buildGrammarGroups[T interface{ build() error }](term grammarTerm, tokens []*token, gram *grammar, initialize func(string, []*token, *grammar) (T, error), split func([]*token) ([][]*token, []*token, error)) ([]T, []*token, error) {
	groups := [][]*token{tokens}
	var separators []*token
	if term.Multiple {
		if len(tokens) == 0 {
			return nil, nil, lerrors.Syntax.Errorf("error: %ss initialized with empty slice of tokens", term.Family)
		}
		var err error
		groups, separators, err = split(tokens)
		if err != nil {
			return nil, nil, err
		}
	}
	var children []T
	for _, group := range groups {
		child, err := initialize(term.Name, group, gram)
		if err != nil {
			return nil, nil, err
		}
		children = append(children, child)
	}
	for _, child := range children {
		if err := child.build(); err != nil {
			return nil, nil, err
		}
	}
	return children, separators, nil
}

func buildGrammarConditions(term grammarTerm, tokens []*token, gram *grammar) ([]*conditionGram, []*token, error) {
	return buildGrammarGroups(term, tokens, gram, initConditionGram, identifyConditionSequences)
}

func buildGrammarConditionals(term grammarTerm, tokens []*token, gram *grammar) ([]*conditionalGram, []*token, error) {
	return buildGrammarGroups(term, tokens, gram, initConditionalGram, func(tokens []*token) ([][]*token, []*token, error) {
		groups, err := identifyConditionalSequences(tokens)
		return groups, nil, err
	})
}

// buildGrammarPrefix applies a node's selected rule, then commits its consumed
// prefix. Failed rules leave prefix handling to the enclosing rollback boundary.
func buildGrammarPrefix(rule *grammeme, caseType string, tokens *[]*token, build func(string) error, remaining func() []*token) error {
	structure, err := rule.getCasedStructure(caseType)
	if err != nil {
		return err
	}
	if err := build(structure); err != nil {
		return err
	}
	*tokens, err = grammarPrefix(*tokens, remaining())
	return err
}

// A speculative failure retains its cause without rendering a discarded branch.
type grammarRuleError struct {
	part  string
	cause error
}

func (e grammarRuleError) Error() string { return e.part + ": " + e.cause.Error() }
func (e grammarRuleError) Unwrap() error { return e.cause }

// Each optional element commits only after success. Grammar matching cannot mutate bindings.
// Callbacks borrow the candidate during their call and must not retain its address.
func applyGrammarRule[T any](target *T, rule string, clone func(T) T,
	grammatical func(*T, grammarTerm) error,
	lexical func(*T, grammarTerm) error,
	typed func(*T, grammarTerm) error) error {
	var candidate T
	for _, part := range splitGramRule(rule) {
		term, err := decodeGrammarTerm(part)
		if err != nil {
			return err
		}
		candidate = clone(*target)
		var apply func(*T, grammarTerm) error
		var kind string
		switch term.Kind {
		case termGrammar, termCase:
			if term.Family == familyStatement {
				return lerrors.InvalidConfiguration.Errorf("%s: unknown grammar rule element %s", part, part)
			}
			apply, kind = grammatical, "grammatical"
		case termLexical, termIdentifier:
			apply, kind = lexical, "lexical"
		case termType:
			apply, kind = typed, "typed"
		default:
			return lerrors.InvalidConfiguration.Errorf("%s: unknown grammar rule element %s", part, part)
		}
		if apply == nil {
			err = lerrors.InvalidConfiguration.Errorf("unsupported %s rule %s", kind, part)
		} else {
			err = apply(&candidate, term)
		}
		if err != nil {
			if term.Optional {
				continue
			}
			return grammarRuleError{part, err}
		}
		*target = candidate
	}
	return nil
}

// Evaluate alternatives from the same state, preferring the longest valid prefix.
// Sorted names break ties; a complete prefix cannot be beaten by any later case.
// As in applyGrammarRule, callbacks borrow the candidate without retaining it.
func selectGrammarCase[T any](target *T, cases map[string]string, clone func(T) T,
	build func(*T, string) error, remaining func(*T) []*token,
	setCase func(*T, string), tokens []*token) error {
	names := make([]string, 0, len(cases))
	for name := range cases {
		names = append(names, name)
	}
	sort.Strings(names)
	base := clone(*target)
	bestCount := -1
	var best T
	// Successful alternatives need no diagnostics from rejected branches.
	// Retain their causes and source position, and render only if all fail.
	type failure struct {
		name     string
		err      error
		next     *token
		position bool
	}
	var rejected []failure
	var candidate T
	for _, name := range names {
		candidate = clone(base)
		if err := build(&candidate, cases[name]); err != nil {
			detail := failure{name: name, err: err, position: true}
			if suffix := remaining(&candidate); len(suffix) > 0 {
				detail.next = suffix[0]
			}
			rejected = append(rejected, detail)
			continue
		}
		prefix, err := grammarPrefix(tokens, remaining(&candidate))
		if err != nil {
			rejected = append(rejected, failure{name: name, err: err})
			continue
		}
		if len(prefix) > bestCount {
			setCase(&candidate, name)
			if len(prefix) == len(tokens) {
				*target = candidate
				return nil
			}
			best, bestCount = candidate, len(prefix)
		}
	}
	if bestCount >= 0 {
		*target = best
		return nil
	}
	failures := make([]string, 0, len(rejected))
	causes := make([]error, 0, len(rejected))
	for _, detail := range rejected {
		message := detail.name + ": " + detail.err.Error()
		if detail.position {
			if detail.next != nil {
				message += fmt.Sprintf("; remaining token %q", detail.next.rawBytes)
			} else {
				message += "; at end of statement"
			}
		}
		failures = append(failures, message)
		causes = append(causes, detail.err)
	}
	message := "no grammar case matches (" + strings.Join(failures, "; ") + ")"
	if len(causes) == 0 {
		return lerrors.Syntax.Errorf("%s", message)
	}
	return lerrors.Syntax.Wrap(errors.Join(causes...), message)
}

func (s conditionGram) clone() conditionGram {
	s.values = append([]*valueGram(nil), s.values...)
	s.prepositions = append([]*token(nil), s.prepositions...)
	return s
}

func (s conditionalGram) clone() conditionalGram {
	s.evaluations = append([]*token(nil), s.evaluations...)
	s.conditions = append([]*conditionGram(nil), s.conditions...)
	s.conjunctions = append([]*token(nil), s.conjunctions...)
	s.clauses = append([]*clauseGram(nil), s.clauses...)
	return s
}

func (s nominalGram) clone() nominalGram {
	s.adjectives = append([]*token(nil), s.adjectives...)
	s.adjectiveForms = append([]selectedTokenForm(nil), s.adjectiveForms...)
	return s
}

func (s valueGram) clone() valueGram {
	s.nominalGram = s.nominalGram.clone()
	return s
}

func (s subjectGram) clone() subjectGram {
	s.nominalGram = s.nominalGram.clone()
	return s
}

func (s objectGram) clone() objectGram {
	s.nominalGram = s.nominalGram.clone()
	return s
}

func (s phraseGram) clone() phraseGram {
	s.values = append([]*valueGram(nil), s.values...)
	s.subjects = append([]*subjectGram(nil), s.subjects...)
	s.markers = append([]*token(nil), s.markers...)
	s.prepositions = append([]*token(nil), s.prepositions...)
	s.clauses = append([]*clauseGram(nil), s.clauses...)
	s.conjunctions = append([]*token(nil), s.conjunctions...)
	s.objects = append([]*objectGram(nil), s.objects...)
	return s
}

func (s clauseGram) clone() clauseGram {
	s.values = append([]*valueGram(nil), s.values...)
	s.phrases = append([]*phraseGram(nil), s.phrases...)
	s.adverbs = append([]*token(nil), s.adverbs...)
	s.prepositions = append([]*token(nil), s.prepositions...)
	s.clauses = append([]*clauseGram(nil), s.clauses...)
	s.conjunctions = append([]*token(nil), s.conjunctions...)
	s.objects = append([]*objectGram(nil), s.objects...)
	return s
}

func (s predicateGram) clone() predicateGram {
	s.clauses = append([]*clauseGram(nil), s.clauses...)
	s.conditionals = append([]*conditionalGram(nil), s.conditionals...)
	s.conjunctions = append([]*token(nil), s.conjunctions...)
	return s
}

func (s statementGram) clone() statementGram {
	s.clauses = append([]*clauseGram(nil), s.clauses...)
	s.conjunctions = append([]*token(nil), s.conjunctions...)
	return s
}

// A separator belongs to this repeated element only when another element parses.
// Otherwise its parent gets the separator and decides whether it is valid there.
func grammarSequence[T any](tokens []*token, multiple bool, parse func([]*token) (T, []*token, error)) ([]T, []*token, error) {
	var values []T
	var separators []*token
	remaining := tokens
	value, consumed, err := parse(remaining)
	for {
		if err != nil {
			return nil, nil, err
		}
		if len(consumed) == 0 || len(consumed) > len(remaining) {
			return nil, nil, lerrors.Internal.Errorf("grammar sequence made no progress")
		}
		values = append(values, value)
		remaining = remaining[len(consumed):]
		if !multiple || len(remaining) == 0 {
			break
		}
		var separator *token
		next := remaining
		if remaining[0].getPOS(false) == "conjunctions" {
			separator = remaining[0]
			next = remaining[1:]
		} else if !consumed[len(consumed)-1].conjoined {
			break
		}
		if len(next) == 0 {
			break
		}
		// Retain successful lookahead for the next iteration. Its consumption is
		// checked there, before publishing the value, just like the first item.
		value, consumed, err = parse(next)
		if err != nil {
			break
		}
		if separator != nil {
			separators = append(separators, separator)
		}
		remaining = next
	}
	return values, separators, nil
}

func buildGrammarClauses(term grammarTerm, tokens []*token, gram *grammar) ([]*clauseGram, []*token, error) {
	var leading *token
	if term.Multiple && term.Optional && len(tokens) > 0 && tokens[0].getPOS(false) == "conjunctions" {
		leading, tokens = tokens[0], tokens[1:]
	}
	clauses, separators, err := grammarSequence(tokens, term.Multiple, func(input []*token) (*clauseGram, []*token, error) {
		signatureError := gram.signatureError
		clause, err := buildGrammarNode(initClauseGram(term.Name, input, gram))
		if err != nil && term.Name == "clause" {
			// An unqualified clause admits the declared forms of its selected
			// word. Keep a successful root parse; retry its same-spelling
			// passive only when that first grammar candidate cannot consume it.
			head := clauseHead(input, "")
			if head != nil && head.grammaticalForm("root") != nil && head.grammaticalForm("passive") != nil {
				gram.signatureError = signatureError
				clause, err = buildGrammarNode(initClauseGram("clause.passive", input, gram))
			}
		}
		if err != nil {
			return nil, nil, err
		}
		return clause, clause.tokens, nil
	})
	if err != nil {
		return nil, nil, err
	}
	if leading != nil {
		separators = append([]*token{leading}, separators...)
	}
	return clauses, separators, nil
}

func buildGrammarValues(term grammarTerm, tokens []*token, gram *grammar) ([]*valueGram, []*token, error) {
	return grammarSequence(tokens, term.Multiple, func(input []*token) (*valueGram, []*token, error) {
		value, err := buildGrammarNode(initValueGram(term.Name, input, gram))
		if err != nil {
			return nil, nil, err
		}
		return value, value.tokens, nil
	})
}
