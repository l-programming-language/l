package Lyger

import (
	"sort"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// tokenBinding keeps lexical identity separate from the category used to store
// the value. In particular a noun's active form may serve as a verb, while the
// underlying definition remains a noun.
type tokenBinding struct {
	Record     *record
	Root       string
	POS        string
	Form       string
	Possessive bool
}

func copyStrings(in map[string]string) map[string]string {
	if in == nil {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// cloneUse isolates mutable invocation arguments from the stored definition.
func (l *lexeme) cloneUse() *lexeme {
	if l == nil {
		return nil
	}
	out := *l
	out.Prefixes, out.Postfixes = copyStrings(l.Prefixes), copyStrings(l.Postfixes)
	out.Inputs, out.Outputs = append([]string(nil), l.Inputs...), append([]string(nil), l.Outputs...)
	if l.Arguments != nil {
		out.Arguments = append([]string{}, l.Arguments...)
	}
	out.Adjectives, out.Adverbs = append([]*token(nil), l.Adjectives...), append([]*token(nil), l.Adverbs...)
	out.Actions = make([]*lexeme, len(l.Actions))
	for i, action := range l.Actions {
		out.Actions[i] = action.cloneUse()
	}
	return &out
}

// bindFrom copies storage wrappers. Applicative/iterative payload pointers keep
// their identity so that a member reference addresses the original instance.
func (e *entry) bindFrom(source *entry) {
	*e = *source
	if source.Primitive != nil {
		p := *source.Primitive
		p.Bytes = append([]byte(nil), p.Bytes...)
		e.Primitive = &p
	}
	e.Lexeme = source.Lexeme.cloneUse()
	if source.Applicative != nil {
		a := *source.Applicative
		e.Applicative = &a
	}
	if source.Iterative != nil {
		i := *source.Iterative
		e.Iterative = &i
	}
}

type bindingQuery struct {
	word       string
	possessive bool
}

func newBindingQuery(contents []byte) bindingQuery {
	word, possessive := possessiveOwner(string(contents))
	return bindingQuery{word: strings.ToLower(word), possessive: possessive}
}

// A Function's invocation word and value share one record. Lexical morphology
// does not change the value's Function storage or its complete signature.
func (r *record) bindingWord() *lexeme {
	if r == nil || r.Value == nil {
		return nil
	}
	if r.Value.Type == storageLexeme {
		return r.Value.Lexeme
	}
	if !r.isFunction() {
		return nil
	}
	word := newLexeme("verb")
	if source := r.Value.Applicative.Function.Source; source != nil && source.Word != nil {
		word = source.Word.cloneUse()
	}
	name := r.ID[strings.LastIndex(r.ID, ".")+1:]
	word.Root = strings.ToLower(name)
	return word
}

func (r *record) isFunction() bool {
	return r != nil && r.Value != nil && r.Value.Type == storageApplicative && r.Value.Applicative != nil && r.Value.Applicative.Function != nil
}

func (r *record) binding(query bindingQuery) (*tokenBinding, bool) {
	if r == nil || r.Value == nil {
		return nil, false
	}
	word := query.word
	b := &tokenBinding{Record: r, Root: strings.ToLower(r.ID), Form: "root", Possessive: query.possessive}
	if lexical := r.bindingWord(); lexical != nil {
		// A member is visible through its qualified name until an import creates
		// an unqualified alias. Suffix matching must not leak dictionary members.
		prefix := ""
		if dot := strings.LastIndex(r.ID, "."); dot >= 0 {
			prefix = strings.ToLower(r.ID[:dot+1])
		}
		if !strings.HasPrefix(word, prefix) {
			return nil, false
		}
		matched, form := lexical.match([]byte(strings.TrimPrefix(word, prefix)))
		if !matched {
			return nil, false
		}
		b.Root, b.POS, b.Form = lexical.Root, lexical.POS, form
		return b, true
	}
	return b, word == strings.ToLower(r.ID)
}

func (b *tokenBinding) lexicalType() string {
	if b.POS != "" {
		pos, form := b.POS, b.Form
		if pos == "noun" && form == "active" {
			pos, form = "verb", "applicative"
		}
		if pos == "noun" && form == "descriptive" {
			pos, form = "adjective", "root"
		}
		if pos == "verb" && form == "objective" {
			pos, form = "noun", "root"
		}
		t := pos + "s.generic." + form
		if b.Possessive {
			t += ".possessive"
		}
		return t
	}
	if b.Possessive {
		return "nouns.generic.root.possessive"
	}
	r := b.Record
	if r.Value.Type == storagePrimitive && r.Value.Primitive != nil {
		return "primitive." + strings.ToLower(r.Value.Primitive.Type.String())
	}
	return "nouns.generic.root"
}

func resolveTokenBinding(contents []byte, memory map[string]*record) (*tokenBinding, error) {
	query := newBindingQuery(contents)
	keys := make([]string, 0, len(memory))
	for key := range memory {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var result *tokenBinding
	for _, key := range keys {
		if binding, ok := memory[key].binding(query); ok {
			if result != nil && result.Record != binding.Record {
				return nil, lerrors.AmbiguousReference.Errorf("ambiguous reference %q: %s and %s", contents, result.Record.ID, binding.Record.ID)
			}
			result = binding
		}
	}
	return result, nil
}

// bindStatementMembers resolves the right side of possessive member access in
// the owner's namespace. The tokens retain their separate grammatical roles.
func (r semanticResolver) bindStatementMembers(statement *lStatement) (bool, error) {
	ready := true
	for i, member := range statement.tokens {
		if r.scopes != nil {
			if binding, ok := r.scopes.Pronouns[member]; ok && binding.Kind == scopeParagraph {
				rec, err := r.reference(member)
				if err != nil {
					return false, err
				}
				if rec == nil {
					ready = false
				}
			}
		}
		var owner *token
		possessed := false
		if i > 0 {
			owner = statement.tokens[i-1]
			_, possessed = possessiveOwner(string(owner.rawBytes))
		}
		if !possessed {
			// Any inherited environment must reconsider local reservations, including
			// program/dictionary roots; this does not classify an activation frame.
			if !member.isDeclare() && !member.isPronoun() && (member.getPOS(false) == "unknown" || (member.scope != nil && member.scope.parent != nil)) {
				state, err := r.bindToken(member)
				if err != nil {
					return false, err
				}
				if state == lookupPending {
					ready = false
				}
			}
			continue
		}
		if r.scopes != nil && r.scopes.Pronouns[owner].Kind == scopeActionArgument {
			// The call's argument supplies this owner during instruction lowering.
			continue
		}
		resolvedOwner, ownerReady, err := r.owner(owner)
		if err != nil {
			return false, err
		}
		if !ownerReady {
			ready = false
			continue
		}
		if resolvedOwner.Record == nil && owner.getPOS(false) == "unknown" {
			name, _ := possessiveOwner(string(owner.rawBytes))
			kind, category := checkType([]byte(name))
			owner.mem = initEntry(kind, category)
			owner.tType = "nouns.generic.root.possessive"
		}
		if member.isDeclare() {
			continue
		}
		resolved, memberReady, err := r.member(resolvedOwner, member, statement)
		if err != nil {
			return false, err
		}
		if !memberReady {
			member.binding, member.bindingError, member.defined = nil, nil, false
			member.tType = "primitive.unknown"
			member.mem = &entry{Type: storagePrimitive, Primitive: newPrimitive(valueUnknown)}
			ready = false
			continue
		}
		member.mem = resolved.Value
		member.binding, member.tType, member.defined = resolved.Binding, resolved.Type, resolved.Binding != nil
		member.bindingError = nil
	}
	// A grammatical subject selects its action namespace just as a possessive
	// selects a member. Keep the subject; it is not an invented argument.
	for index, subject := range statement.tokens {
		if subject.getPOS(true) == "article" || subject.getPOS(true) == "adjective" {
			continue
		}
		if subject.isDeclare() || subject.binding == nil || subject.getPOS(true) == "verb" || index+1 >= len(statement.tokens) {
			break
		}
		action := statement.tokens[index+1]
		if action.getPOS(true) == "preposition" {
			break
		}
		member, _, err := r.lookupMember(subject.binding.Record, action.rawBytes)
		if err != nil {
			return false, err
		}
		if member != nil {
			action.binding, action.tType, action.defined = member, member.lexicalType(), true
			action.mem = &entry{}
			action.mem.bindFrom(member.Record.Value)
		}
		break
	}
	return ready, nil
}
