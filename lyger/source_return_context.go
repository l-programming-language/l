package Lyger

import "github.com/l-programming-language/l/lerrors"

type sourceReturnContext struct {
	returned       *record
	constructor    *record
	implementation *record
}

func functionReturnContext(function *sourceFunction) sourceReturnContext {
	if function == nil {
		return sourceReturnContext{}
	}
	return sourceReturnContext{returned: function.Return, constructor: function.Definition}
}

// Source actions retain their lexical Return contract when authored inside a
// Function. Root-authored actions receive the caller's selected Return; its
// implementation preserves authored delegates. Context never changes tokens.
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
	if selected != nil && context.implementation == nil {
		return selected, true, nil
	}
	return bound, true, nil
}

// Follow the selected source body, never a real Function body or an opaque
// constructor's supplied directive. Only transparent W sequences execute those
// operands unconditionally; source actors share the action's canonical record.
func (b *instructionBuilder) sourceActionReturns(action *clauseGram, context sourceReturnContext, active map[*record]bool) (bool, error) {
	if action.subject != nil && action.subject.noun.isDeclare() {
		site := registrationSite{subject: action.subject, predicate: action.predicate, clause: action}
		return b.sourceObjectReturns(site.valueObject(), context, active)
	}
	selected, _, err := sourceReturnBinding(b.semantics, action.verb, context)
	if err != nil {
		return false, err
	}
	return b.sourceDefinitionReturns(selected, action.orderedObjects(), nil, context, active)
}

func (b *instructionBuilder) sourceDefinitionReturns(selected *record, objects []*objectGram, body []*clauseGram, context sourceReturnContext, active map[*record]bool) (bool, error) {
	if selected.function != nil || active[selected] {
		return false, nil
	}
	if b.semantics.Returns[selected] {
		return true, nil
	}
	shape, err := b.semantics.selectedDefinition(selected)
	if err != nil || shape.site == nil {
		return false, err
	}
	if shape.template != nil {
		if !isProcessDefinitionTemplate(shape.template) {
			return false, nil
		}
		for _, object := range objects {
			returns, err := b.sourceObjectReturns(object, context, active)
			if err != nil || returns {
				return returns, err
			}
		}
	} else {
		active[selected] = true
		defer delete(active, selected)
		body = shape.site.actionClauses()
		if len(body) == 0 {
			return b.sourceObjectReturns(shape.object, context, active)
		}
	}
	for _, child := range body {
		returns, err := b.sourceActionReturns(child, context, active)
		if err != nil || returns {
			return returns, err
		}
	}
	return false, nil
}

func (b *instructionBuilder) sourceObjectReturns(object *objectGram, context sourceReturnContext, active map[*record]bool) (bool, error) {
	if object == nil || object.noun == nil || object.noun.binding == nil {
		return false, nil
	}
	for _, modifier := range valueModifiers(object.adjectives) {
		if !modifier.hasRole("adjective", "initialization") && !modifier.isPossessive() {
			return false, nil
		}
	}
	selected, err := canonicalRecord(object.noun.binding.Record, b.semantics.Aliases)
	if err != nil {
		return false, err
	}
	var body []*clauseGram
	if phrase := object.directivePhrase(); phrase != nil {
		body = phrase.clauses
	}
	return b.sourceDefinitionReturns(selected, constructorOperands(object), body, context, active)
}
