package Lyger

import "github.com/l-programming-language/l/lerrors"

type sourceReturnContext struct {
	returned    *record
	constructor *record
}

func functionReturnContext(function *sourceFunction) sourceReturnContext {
	if function == nil {
		return sourceReturnContext{}
	}
	return sourceReturnContext{returned: function.Return, constructor: function.Definition}
}

// Source actions retain their lexical Return contract when authored inside a
// Function. Root-authored actions receive the caller's selected Return instead.
// The context is private to one expansion and never changes the source tokens.
func (c *Compilation) withDefinitionReturn(context sourceReturnContext) func() {
	previous := c.definitionReturn
	c.definitionReturn = context
	return func() { c.definitionReturn = previous }
}

func sourceReturnBinding(semantics *semanticProgram, tok *token, context sourceReturnContext) (*record, bool, error) {
	if tok == nil || tok.binding == nil {
		return nil, false, lerrors.UnresolvedReference.Errorf("source operation has no definition")
	}
	bound, err := canonicalRecord(tok.binding.Record, semantics.Aliases)
	if err != nil {
		return nil, false, err
	}
	var owner *sourceFunction
	for scope := tok.scope; scope != nil; scope = scope.parent {
		if scope.function != nil {
			owner = scope.function
			break
		}
	}
	isReturn := semantics.Returns[bound]
	if !isReturn {
		return bound, false, nil
	}
	selected, err := canonicalRecord(context.returned, semantics.Aliases)
	if err != nil {
		return nil, false, err
	}
	if owner != nil {
		authored, err := canonicalRecord(owner.Return, semantics.Aliases)
		if err != nil {
			return nil, false, err
		}
		if authored != bound || selected != nil && selected != authored || context.constructor != nil && context.constructor != owner.Definition {
			return nil, false, lerrors.TypeMismatch.Errorf("Return belongs to a different Function constructor contract")
		}
		return bound, true, nil
	}
	if selected != nil {
		return selected, true, nil
	}
	return bound, true, nil
}

func (c *Compilation) sourceActionReturn(tok *token) (*record, bool, error) {
	return sourceReturnBinding(c.Semantics, tok, c.definitionReturn)
}

// Classify source-composed unconditional returns before planning. Native
// Function invocations and opaque native leaves do not return from their caller.
// Only actual selected Return operations in a source action sequence terminate it.
func (b *instructionBuilder) sourceActionReturns(action *clauseGram, context sourceReturnContext, active map[*record]bool) (bool, error) {
	selected, isReturn, err := sourceReturnBinding(b.semantics, action.verb, context)
	if err != nil || isReturn {
		return isReturn, err
	}
	if selected.function != nil || active[selected] {
		return false, nil
	}
	if word := selected.bindingWord(); word != nil && word.Value != "" {
		template, err := ParseDefinitionTemplate([]byte(word.Value), SourceSpan{})
		if err != nil {
			return false, err
		}
		if isProcessDefinitionTemplate(template) {
			for _, object := range action.orderedObjects() {
				transparent, err := b.sourceProcessObject(object)
				if err != nil {
					return false, err
				}
				if !transparent {
					continue
				}
				if body := object.directivePhrase(); body != nil {
					for _, child := range body.clauses {
						terminates, err := b.sourceActionReturns(child, context, active)
						if err != nil || terminates {
							return terminates, err
						}
					}
				}
			}
		}
		return false, nil
	}
	definition := b.semantics.Declarations[selected]
	if definition == nil {
		return false, nil
	}
	active[selected] = true
	defer delete(active, selected)
	for _, child := range directiveClauses(definition) {
		terminates, err := b.sourceActionReturns(child, context, active)
		if err != nil || terminates {
			return terminates, err
		}
	}
	return false, nil
}

// A directive body's ownership does not prove that its constructor executes it.
// Only an unmodified, transparent W sequence can carry unconditional exit facts
// across this boundary; descriptors and other native templates remain opaque.
func (b *instructionBuilder) sourceProcessObject(object *objectGram) (bool, error) {
	if object == nil || object.noun == nil || object.noun.binding == nil {
		return false, nil
	}
	for _, modifier := range valueModifiers(object.adjectives) {
		if !modifier.hasRole("adjective", "initialization") {
			return false, nil
		}
	}
	selected, err := canonicalRecord(object.noun.binding.Record, b.semantics.Aliases)
	if err != nil {
		return false, err
	}
	word := selected.bindingWord()
	if selected.function != nil || word == nil || word.Value == "" {
		return false, nil
	}
	template, err := ParseDefinitionTemplate([]byte(word.Value), SourceSpan{})
	return isProcessDefinitionTemplate(template), err
}
