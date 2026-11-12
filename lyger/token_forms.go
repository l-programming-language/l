package Lyger

import (
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// tokenForm is a read-only grammatical view of an already resolved token.
// Speculative grammar nodes retain the original token and commit this view only
// after the complete statement succeeds; no lookup or alias selection occurs.
type tokenForm struct {
	form        string
	lexicalType string
}

func (t *token) grammaticalForm(form string) *tokenForm {
	word := t.lexeme()
	if word == nil {
		return nil
	}
	if form == "imperative" {
		form = "root"
	}
	if word.POS == "noun" && form == "applicative" {
		form = "active"
	}
	spelling := newBindingQuery(t.rawBytes).word
	spelling = spelling[strings.LastIndex(spelling, ".")+1:]
	if !word.matchesForm(spelling, form) {
		return nil
	}
	view := &tokenForm{form: form}
	if t.binding != nil {
		binding := *t.binding
		binding.Form = form
		view.lexicalType = binding.lexicalType()
	} else {
		// Keep the configured category and native role; a contextual form is
		// not a new generic source binding.
		view.lexicalType = t.tType[:strings.LastIndex(t.tType, ".")+1] + form
	}
	return view
}

func (view *tokenForm) commit(t *token) {
	if t.binding != nil && t.binding.Form != view.form {
		binding := *t.binding
		binding.Form = view.form
		t.binding = &binding
	}
	t.tType = view.lexicalType
}

// projected keeps speculative constructor-signature queries on the selected
// identity while leaving the source occurrence unchanged.
func (view *tokenForm) projected(t *token) *token {
	if view == nil || t == nil {
		return t
	}
	copy := *t
	view.commit(&copy)
	return &copy
}

type selectedTokenForm struct {
	token *token
	view  *tokenForm
}

func nominalContext(article *token, adjectives []*token) bool {
	if article != nil && article.hasRole("article", "variable") {
		return true
	}
	for _, adjective := range adjectives {
		if adjective.hasRole("adjective", "initialization") || adjective.hasRole("adjective", "abstraction") {
			return true
		}
	}
	return false
}

func (t *token) nominalForm() *tokenForm {
	word := t.lexeme()
	if word == nil || t.isDeclare() {
		return nil
	}
	var roles []string
	switch word.POS {
	case "noun":
		roles = []string{"root", "plural"}
	case "verb":
		roles = []string{"objective"}
	}
	for _, role := range roles {
		if view := t.grammaticalForm(role); view != nil {
			return view
		}
	}
	return nil
}

func followingNominal(tokens []*token) bool {
	for _, tok := range tokens {
		if tok.isPossessive() {
			continue
		}
		if tok.nominalForm() != nil {
			return true
		}
		if tok.getPOS(true) == "adjective" {
			continue
		}
		return tok.defined || tok.isRawType() || tok.isDeclare() || tok.getPOS(true) == "noun"
	}
	return false
}

// A descriptive view consumes a modifier only while a following nominal is
// available. Otherwise an overlapping noun spelling stays the nominal itself.
func associateNominalAdjectives(term grammarTerm, available []*token, adjectives *[]*token, views *[]selectedTokenForm) error {
	for index, tok := range available {
		if tok.isPossessive() {
			*adjectives = append(*adjectives, tok)
			continue
		}
		var view *tokenForm
		if word := tok.lexeme(); word != nil && word.POS == "noun" && followingNominal(available[index+1:]) {
			view = tok.grammaticalForm("descriptive")
		}
		if tok.getPOS(true) == "adjective" {
			if tok.nominalForm() != nil && !followingNominal(available[index+1:]) {
				break
			}
		} else if view == nil {
			break
		}
		*adjectives = append(*adjectives, tok)
		if view != nil {
			*views = append(*views, selectedTokenForm{tok, view})
		}
		if !term.Multiple {
			break
		}
	}
	if !term.Optional && len(*adjectives) == 0 {
		return lerrors.Syntax.Errorf("Error: adjective is not an optional type!")
	}
	return nil
}

// clauseHead observes the same adverb/possessive prefix accepted by grammar.
func clauseHead(tokens []*token) *token {
	return clauseHeadForForm(tokens, "")
}

// Only an explicit applicative context can promote a selected noun's active
// form; ordinary lookup and the possessive owner retain their original roles.
func clauseHeadForForm(tokens []*token, form string) *token {
	for _, tok := range tokens {
		if tok.getPOS(true) == "verb" {
			return tok
		}
		if form == "applicative" && !tok.isPossessive() && tok.binding != nil && tok.binding.POS == "noun" && tok.grammaticalForm(form) != nil {
			return tok
		}
		if tok.getPOS(true) != "adverb" && !tok.isPossessive() {
			return nil
		}
	}
	return nil
}

func (sg *statementGram) commitForms() error {
	collector := grammarFormCollector{seen: make(map[any]bool), forms: make(map[*token]*tokenForm)}
	collector.visit(sg)
	if collector.err != nil {
		return collector.err
	}
	for _, tok := range collector.order {
		collector.forms[tok].commit(tok)
	}
	return nil
}

type grammarFormCollector struct {
	seen  map[any]bool
	forms map[*token]*tokenForm
	order []*token
	err   error
}

func (c *grammarFormCollector) add(tok *token, view *tokenForm, modifiers []selectedTokenForm) {
	if tok != nil && view != nil {
		if prior := c.forms[tok]; prior != nil && prior.form != view.form {
			c.err = lerrors.Internal.Errorf("one source token acquired conflicting grammatical forms")
		} else if prior == nil {
			c.forms[tok] = view
			c.order = append(c.order, tok)
		}
	}
	for _, modifier := range modifiers {
		c.add(modifier.token, modifier.view, nil)
	}
}

// Walk only the accepted grammar graph. Aliased object/phrase fields are
// visited once, and all views are collected before any token is committed.
func (c *grammarFormCollector) visit(node any) {
	if c.seen[node] {
		return
	}
	c.seen[node] = true
	switch node := node.(type) {
	case *statementGram:
		if node == nil {
			return
		}
		c.visit(node.subject)
		c.visit(node.predicate)
		for _, child := range node.clauses {
			c.visit(child)
		}
	case *subjectGram:
		if node == nil {
			return
		}
		c.add(node.noun, node.nounForm, node.adjectiveForms)
		c.visit(node.phrase)
	case *objectGram:
		if node == nil {
			return
		}
		c.add(node.noun, node.nounForm, node.adjectiveForms)
		c.visit(node.name)
		c.visit(node.phrase)
		c.visit(node.clause)
	case *valueGram:
		if node == nil {
			return
		}
		c.add(node.noun, node.nounForm, node.adjectiveForms)
	case *phraseGram:
		if node == nil {
			return
		}
		c.visit(node.subject)
		c.visit(node.obj)
		c.visit(node.phrase)
		for _, child := range node.subjects {
			c.visit(child)
		}
		for _, child := range node.objects {
			c.visit(child)
		}
		for _, child := range node.values {
			c.visit(child)
		}
		for _, child := range node.clauses {
			c.visit(child)
		}
	case *clauseGram:
		if node == nil {
			return
		}
		c.add(node.verb, node.verbForm, nil)
		c.visit(node.subject)
		c.visit(node.obj)
		c.visit(node.predicate)
		c.visit(node.phrase)
		for _, child := range node.objects {
			c.visit(child)
		}
		for _, child := range node.values {
			c.visit(child)
		}
		for _, child := range node.phrases {
			c.visit(child)
		}
		for _, child := range node.clauses {
			c.visit(child)
		}
	case *predicateGram:
		if node == nil {
			return
		}
		c.visit(node.obj)
		for _, child := range node.clauses {
			c.visit(child)
		}
		for _, child := range node.conditionals {
			c.visit(child)
		}
	case *conditionalGram:
		if node == nil {
			return
		}
		for _, child := range node.clauses {
			c.visit(child)
		}
		for _, child := range node.conditions {
			c.visit(child)
		}
	case *conditionGram:
		if node == nil {
			return
		}
		for _, child := range node.values {
			c.visit(child)
		}
	}
}
