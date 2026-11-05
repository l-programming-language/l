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
	record := tok.binding.Record
	if function := record.function; function != nil && function.Native {
		return grammarArity{Count: len(function.Constructor.Inputs), Fixed: true}, nil
	}
	word := record.bindingWord()
	if word == nil {
		return grammarArity{}, nil
	}
	if word.Value == "" {
		// Source actions without explicit formals inherit caller operands through
		// their composition. They do not declare a fixed zero-operand signature.
		return grammarArity{Count: len(word.Inputs), Fixed: len(word.Inputs) > 0}, nil
	}
	template, err := ParseDefinitionTemplate([]byte(word.Value), SourceSpan{})
	if err != nil {
		return grammarArity{}, err
	}
	return templateGrammarSignature(template, constructor, named), nil
}

func templateGrammarSignature(template *DefinitionTemplate, constructor, named bool) grammarArity {
	arity := grammarArity{Fixed: true}
	for index, slot := range template.Slots {
		if slot.Type == "Syllables" || slot.Type == "Atoms" {
			return grammarArity{}
		}
		if template.scopedLabels[slot.Name] || constructor && index == 0 && (slot.Type == "Label" || named) {
			continue
		}
		if constructor && template.IsFunction() && template.signatureSlots[slot.Name] != "" {
			continue
		}
		arity.Count++
	}
	return arity
}

// Select the same canonical descriptor rules as construction, without emitting
// declarations, binding values, or allocating definition labels during grammar.
func (c *Compilation) constructorGrammarSignature(noun *token, modifiers []*token, named bool) (grammarArity, error) {
	if len(valueModifiers(modifiers)) == 0 {
		return definitionGrammarSignature(noun, true, named)
	}
	node, err := c.grammarConstruction(noun, modifiers, make(map[*descriptorRule]bool))
	if err != nil {
		return grammarArity{}, err
	}
	selectedNoun := *noun
	if node.Definition != nil {
		selectedNoun.binding = &tokenBinding{Record: node.Definition}
	}
	arity, err := definitionGrammarSignature(&selectedNoun, true, named || node.Name != "")
	if err != nil {
		return arity, err
	}
	templates, selected, err := c.constructionTemplates(node)
	if err != nil {
		return arity, err
	}
	if selected >= 0 {
		arity = templateGrammarSignature(templates[selected], true, named || node.Name != "")
	}
	for _, argument := range node.Arguments {
		if arity.Fixed && arity.Count > 0 && (argument.article == nil || !argument.article.hasRole("article", "variable")) {
			arity.Count--
		}
	}
	return arity, nil
}

// Select only the construction header. Value binding, native type resolution,
// dependencies and body construction remain in the ordinary construction pass.
func (c *Compilation) grammarConstruction(noun *token, modifiers []*token, active map[*descriptorRule]bool) (*constructionNode, error) {
	nominal, err := nominalType(noun)
	if err != nil {
		return nil, err
	}
	node := &constructionNode{Nominal: nominal}
	if noun.binding != nil {
		node.Definition, err = canonicalRecord(noun.binding.Record, c.Semantics.Aliases)
		if err != nil {
			return nil, err
		}
	}
	for _, modifier := range valueModifiers(modifiers) {
		descriptor, err := c.descriptorFor(modifier)
		if err != nil {
			return nil, err
		}
		if descriptor == nil {
			continue
		}
		target, err := c.resolver().nominalIdentity(noun)
		if err != nil {
			return nil, err
		}
		if err := c.descriptorGrammarReady(descriptor.Record, target); err != nil {
			return nil, err
		}
		matched := false
		for _, rule := range descriptor.Rules {
			if rule.Target != target {
				continue
			}
			matched = true
			wrapper, err := c.sourceDescriptorObject(rule)
			if err != nil {
				return nil, err
			}
			if rule.Template != nil || wrapper != nil {
				node.Wrappers = append(node.Wrappers, rule)
				continue
			}
			if active[rule] {
				return nil, lerrors.Cycle.Errorf("recursive descriptor %s for %s", descriptor.Record.ID, target)
			}
			active[rule] = true
			selected, err := c.grammarConstruction(rule.Object.noun, rule.Object.adjectives, active)
			delete(active, rule)
			if err != nil {
				return nil, err
			}
			if rule.Object.name != nil && rule.Object.name.noun != nil {
				selected.Name = unquoteImportWord(rule.Object.name.noun.rawBytes)
			}
			for phrase := rule.Object.phrase; phrase != nil; phrase = phrase.phrase {
				if phrase.gType != "phrase.signature" && phrase.gType != "phrase.directive" {
					selected.Arguments = append(selected.Arguments, phrase.orderedObjects()...)
				}
			}
			if rule.Object.clause != nil {
				selected.Arguments = append(selected.Arguments, rule.Object.clause.orderedObjects()...)
			}
			node.mergeDescriptor(selected)
		}
		if !matched {
			return nil, lerrors.Unsupported.Errorf("descriptor %s has no constructor for %s", descriptor.Record.ID, target)
		}
	}
	return node, nil
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

func (c *Compilation) sourceActionInputs(statement *lStatement, target *token) ([]string, error) {
	for _, source := range c.Scope.Ordered {
		if source.Statement != statement || len(source.Declarations) == 0 || source.Declarations[0].Token != target {
			continue
		}
		inputs := make([]string, len(source.Arguments))
		for index, argument := range source.Arguments {
			name, err := nominalType(argument.Token)
			if err != nil {
				return nil, err
			}
			inputs[index] = name
		}
		return inputs, nil
	}
	return nil, nil
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
