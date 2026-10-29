package Lyger

import "strings"

// A binding index narrows lookup to records with a matching surface form. The
// normal matcher still decides lexical form, possession and stable ambiguity.
// It exists only while semantic publication owns all writes to these tables.
type recordBindingIndex struct {
	words map[string]map[string]*record
	keys  map[string][]string
}

func newRecordBindingIndex(records map[string]*record) *recordBindingIndex {
	index := &recordBindingIndex{words: make(map[string]map[string]*record), keys: make(map[string][]string)}
	for key, rec := range records {
		index.store(key, rec)
	}
	return index
}

func (index *recordBindingIndex) store(key string, rec *record) {
	for _, word := range index.keys[key] {
		delete(index.words[word], key)
		if len(index.words[word]) == 0 {
			delete(index.words, word)
		}
	}
	delete(index.keys, key)
	if rec == nil || rec.Value == nil {
		return
	}
	forms := []string{strings.ToLower(rec.ID)}
	if word := rec.bindingWord(); word != nil {
		prefix := ""
		if dot := strings.LastIndexByte(rec.ID, '.'); dot >= 0 {
			prefix = strings.ToLower(rec.ID[:dot+1])
		}
		forms = []string{prefix + word.Root}
		for fix := range word.Postfixes {
			forms = append(forms, prefix+processLexFix(word.Root, fix, "postfix"))
		}
		for fix := range word.Prefixes {
			forms = append(forms, prefix+processLexFix(word.Root, fix, "prefix"))
		}
	}
	for _, word := range forms {
		if index.words[word] == nil {
			index.words[word] = make(map[string]*record)
		}
		index.words[word][key] = rec
	}
	index.keys[key] = forms
}

func (s *lexicalScope) storeBinding(key string, rec *record) {
	s.records[key] = rec
	if s.bindings != nil {
		s.bindings.store(key, rec)
	}
}

// Semantic analysis has a single publication path. Drop its indexes when it
// returns so later clients can inspect or modify the exposed Memory records
// without retaining a stale derived view.
func (p *scopePlan) indexBindings() func() {
	scopes := make(map[*lexicalScope]bool)
	add := func(scope *lexicalScope) {
		for ; scope != nil && !scopes[scope]; scope = scope.parent {
			scopes[scope] = true
			scope.bindings = newRecordBindingIndex(scope.records)
		}
	}
	add(p.Root)
	for _, statement := range p.Ordered {
		add(statement.Scope)
		for _, tok := range statement.Statement.tokens {
			add(tok.scope)
		}
	}
	return func() {
		for scope := range scopes {
			scope.bindings = nil
		}
	}
}
