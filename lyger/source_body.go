package Lyger

import "github.com/l-programming-language/l/lerrors"

// A directive supplies an ordered composition to its selected constructor.
// Its clauses retain the lexical Function, source order, and ordinary instruction
// lowering; the constructor owns how that composition is used.
func (c *Compilation) directiveInstructions(object *objectGram) ([]*instruction, bool, error) {
	if object == nil || object.phrase == nil || object.phrase.gType != "phrase.directive" {
		return nil, false, nil
	}
	builder := &instructionBuilder{
		semanticResolver: semanticResolver{scopes: c.Scope, semantics: c.Semantics},
		ordered:          make(map[*lStatement]scopedStatement),
		activeCalls:      make(map[*record]bool),
	}
	var owner *scopedStatement
	for index := range c.Scope.Ordered {
		scoped := &c.Scope.Ordered[index]
		builder.ordered[scoped.Statement] = *scoped
		for _, tok := range scoped.Statement.tokens {
			if tok == object.noun {
				owner = scoped
				break
			}
		}
	}
	if owner == nil {
		return nil, true, lerrors.Internal.Errorf("directive has no source statement")
	}
	ctx := instructionContext{statement: owner.Statement, order: owner.Order}
	for scope := owner.Scope; scope != nil; scope = scope.parent {
		if scope.function != nil {
			ctx.function = scope.function
			break
		}
	}
	if ctx.function == nil || !ctx.function.Native {
		return nil, true, lerrors.Unsupported.Errorf("directive composition requires a Function with native metadata operands")
	}
	body, err := builder.clauses(object.phrase.clauses, ctx)
	return body, true, err
}
