package Lyger

import (
	"github.com/l-programming-language/l/lerrors"
)

func buildConstructorPhrase(term grammarTerm, noun *token, modifiers []*token, named bool, tokens []*token, gram *grammar) (*phraseGram, error) {
	phrase, err := initPhraseGram(term.Name, tokens, gram)
	if err != nil {
		return nil, err
	}
	if len(tokens) > 0 && tokens[0].hasRole("preposition", "definement.definitive") {
		var err error
		phrase.objectLimit, err = definitionGrammarSignature(noun, true, named)
		if err == nil && gram.constructorSignature != nil {
			phrase.objectLimit, err = gram.constructorSignature(noun, modifiers, named)
		}
		if err != nil {
			gram.signatureError = err
			return nil, err
		}
	}
	if err := phrase.build(); err != nil {
		return nil, err
	}
	return phrase, nil
}

// Fixed source signatures bound nested operands before a following sibling.
// Native coordinated roots and variadic composition retain their configured
// grammar; only selected source definitions supply this additional boundary.
// A fixed zero-operand constructor is distinct from a variadic source form.
type grammarArity struct {
	Count int
	Fixed bool
}

func definitionGrammarSignature(tok *token, constructor, named bool) (grammarArity, error) {
	if tok == nil || tok.binding == nil || tok.binding.Record == nil {
		return grammarArity{}, nil
	}
	if !constructor {
		for scope := tok.scope; scope != nil; scope = scope.parent {
			if function := scope.function; function != nil {
				if function.Native && function.bindsReturn(tok) {
					return grammarArity{Count: len(function.Constructor.Outputs), Fixed: true}, nil
				}
				break
			}
		}
	}
	header, err := recordHeader(tok.binding.Record, constructor, named)
	return header.arity(), err
}

// Grammar and lowering select the same header without allocating value labels.
func (c *Compilation) constructorGrammarSignature(noun *token, modifiers []*token, named bool) (grammarArity, error) {
	node, err := c.constructObjectMode(&objectGram{noun: noun, adjectives: modifiers}, make(map[*descriptorRule]bool), constructionHeaderOnly)
	if err != nil {
		return grammarArity{}, err
	}
	header, _, _, err := c.constructionHeader(node, named, make(map[*record]bool))
	if err != nil {
		return grammarArity{}, err
	}
	arity := header.arity()
	for _, argument := range node.Arguments {
		if arity.Fixed && arity.Count > 0 && (argument.article == nil || !argument.article.hasRole("article", "variable")) {
			arity.Count--
		}
	}
	return arity, nil
}

// The source-scope plan already fixes pronoun subjects before retries. Retain
// pending refinement operands so grammar never freezes a descriptor's partial
// rule set merely because a consumer appeared earlier in source order.
type descriptorGrammarDependency struct {
	Subject *token
	Target  *token
}

func (c *Compilation) prepareDescriptorGrammar() {
	c.pendingDescriptorRules = make(map[*lStatement]descriptorGrammarDependency)
	for _, statement := range c.Statements {
		for index, tok := range statement.tokens {
			if tok.getPOS(true) != "verb" {
				continue
			}
			if !tok.hasRole("verb", "action.descriptive") || index == 0 {
				break
			}
			for _, target := range statement.tokens[index+1:] {
				if target.isPossessive() {
					continue
				}
				pos := target.getPOS(true)
				if pos == "noun" || pos == "type" || pos == "unknown" {
					c.pendingDescriptorRules[statement] = descriptorGrammarDependency{Subject: statement.tokens[index-1], Target: target}
					break
				}
			}
			break
		}
	}
}

func (c *Compilation) descriptorGrammarReady(descriptor *record, target nominalIdentity) error {
	if len(c.pendingDescriptorRules) == 0 {
		return nil
	}
	for _, statement := range c.Statements {
		pending, found := c.pendingDescriptorRules[statement]
		if !found {
			continue
		}
		record, err := c.resolver().reference(pending.Subject)
		if err != nil {
			return err
		}
		if record != descriptor {
			continue
		}
		identity, err := c.resolver().nominalIdentity(pending.Target)
		if err != nil || identity == target {
			return statement.source.Span.wrapError(lerrors.UnresolvedReference.Errorf("descriptor %s is waiting for its source rule for %s", descriptor.ID, target))
		}
	}
	return nil
}

func buildBoundedGrammarObjects(term grammarTerm, tokens []*token, gram *grammar, limit grammarArity) ([]*objectGram, []*token, error) {
	count := 0
	return grammarSequence(tokens, term.Multiple, func(input []*token) (*objectGram, []*token, error) {
		if limit.Fixed && count == limit.Count {
			return nil, nil, lerrors.Syntax.Errorf("source definition argument boundary")
		}
		object, err := buildGrammarNode(initObjectGram(term.Name, input, gram))
		if err != nil {
			return nil, nil, err
		}
		count++
		return object, object.tokens, nil
	})
}
