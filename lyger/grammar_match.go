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

// ordered returns nonnil associations in their original source order.
func (selected tokenSelection) ordered(source []*token) []*token {
	result := make([]*token, 0, len(selected))
	for _, tok := range source {
		if tok != nil && selected[tok] {
			result = append(result, tok)
		}
	}
	return result
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

func (selected tokenSelection) firstRun(source []*token) []*token {
	var result []*token
	for _, tok := range source {
		if selected[tok] {
			result = append(result, tok)
		} else if len(result) > 0 {
			break
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
func applyGrammarRule[T any](target *T, rule string, clone func(T) T,
	grammatical func(*T, grammarTerm) error,
	lexical func(*T, grammarTerm) error,
	typed func(*T, grammarTerm) error) error {
	for _, part := range splitGramRule(rule) {
		term, err := decodeGrammarTerm(part)
		if err != nil {
			return err
		}
		candidate := clone(*target)
		switch term.Kind {
		case termGrammar, termCase:
			if term.Family == familyStatement {
				return lerrors.InvalidConfiguration.Errorf("%s: unknown grammar rule element %s", part, part)
			}
			if grammatical == nil {
				err = lerrors.InvalidConfiguration.Errorf("unsupported grammatical rule %s", part)
			} else {
				err = grammatical(&candidate, term)
			}
		case termLexical, termIdentifier:
			if lexical == nil {
				err = lerrors.InvalidConfiguration.Errorf("unsupported lexical rule %s", part)
			} else {
				err = lexical(&candidate, term)
			}
		case termType:
			if typed == nil {
				err = lerrors.InvalidConfiguration.Errorf("unsupported typed rule %s", part)
			} else {
				err = typed(&candidate, term)
			}
		default:
			return lerrors.InvalidConfiguration.Errorf("%s: unknown grammar rule element %s", part, part)
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

// Evaluate every alternative from the same state. Prefer the longest valid prefix;
// equal matches use sorted case names so map iteration cannot change the parse.
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
	for _, name := range names {
		candidate := clone(base)
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

func (s valueGram) clone() valueGram {
	s.adjectives = append([]*token(nil), s.adjectives...)
	return s
}

func (s subjectGram) clone() subjectGram {
	s.adjectives = append([]*token(nil), s.adjectives...)
	return s
}

func (s objectGram) clone() objectGram {
	s.adjectives = append([]*token(nil), s.adjectives...)
	return s
}

func (s phraseGram) clone() phraseGram {
	s.values = append([]*valueGram(nil), s.values...)
	s.subjects = append([]*subjectGram(nil), s.subjects...)
	s.markers = append([]*token(nil), s.markers...)
	s.phrases = append([]*phraseGram(nil), s.phrases...)
	s.prepositions = append([]*token(nil), s.prepositions...)
	s.clauses = append([]*clauseGram(nil), s.clauses...)
	s.conjunctions = append([]*token(nil), s.conjunctions...)
	s.objects = append([]*objectGram(nil), s.objects...)
	return s
}

func (s clauseGram) clone() clauseGram {
	s.values = append([]*valueGram(nil), s.values...)
	s.subjects = append([]*subjectGram(nil), s.subjects...)
	s.markers = append([]*token(nil), s.markers...)
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

func buildGrammarObjects(term grammarTerm, tokens []*token, gram *grammar) ([]*objectGram, []*token, error) {
	return grammarSequence(tokens, term.Multiple, func(input []*token) (*objectGram, []*token, error) {
		object, err := buildGrammarNode(initObjectGram(term.Name, input, gram))
		if err != nil {
			return nil, nil, err
		}
		return object, object.tokens, nil
	})
}

func buildGrammarClauses(term grammarTerm, tokens []*token, gram *grammar) ([]*clauseGram, []*token, error) {
	var leading *token
	if term.Multiple && term.Optional && len(tokens) > 0 && tokens[0].getPOS(false) == "conjunctions" {
		leading, tokens = tokens[0], tokens[1:]
	}
	clauses, separators, err := grammarSequence(tokens, term.Multiple, func(input []*token) (*clauseGram, []*token, error) {
		clause, err := buildGrammarNode(initClauseGram(term.Name, input, gram))
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
