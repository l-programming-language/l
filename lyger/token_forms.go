package Lyger

import "strings"

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

// clauseHead observes the same adverb/possessive prefix accepted by grammar.
func clauseHead(tokens []*token) *token {
	for _, tok := range tokens {
		if tok.getPOS(true) == "verb" {
			return tok
		}
		if tok.getPOS(true) != "adverb" && !tok.isPossessive() {
			return nil
		}
	}
	return nil
}

func (sg *statementGram) commitForms() error {
	var clauses []*clauseGram
	if err := sg.walkRegistration(func(site registrationSite, _ []*token, _ bool) error {
		if site.clause != nil && site.clause.verbForm != nil {
			clauses = append(clauses, site.clause)
		}
		return nil
	}); err != nil {
		return err
	}
	for _, clause := range clauses {
		clause.verbForm.commit(clause.verb)
	}
	return nil
}
