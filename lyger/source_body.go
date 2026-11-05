package Lyger

import (
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

func (c *Compilation) directiveComposition(object *objectGram, inherited []DefinitionArgument, active, dynamic map[*record]bool) (DefinitionInstance, bool, error) {
	if object == nil || object.directivePhrase() == nil {
		return DefinitionInstance{}, false, nil
	}
	if object.noun.inNativeFunction() {
		instructions, _, err := c.directiveInstructions(object)
		if err != nil {
			return DefinitionInstance{}, false, err
		}
		body, err := c.composeInstructions(instructions, active, dynamic)
		return body, false, err
	}
	var body DefinitionInstance
	var expressions []string
	forwarded := false
	var compose func([]*clauseGram) error
	compose = func(clauses []*clauseGram) error {
		for _, clause := range clauses {
			var instance DefinitionInstance
			var err error
			var used bool
			if clause.subject != nil && clause.subject.noun.isDeclare() {
				rec := c.Semantics.Subjects[clause.subject.noun]
				if rec == nil || rec.staticConstructor == nil {
					return lerrors.UnresolvedReference.Errorf("Process declaration has no constructed binding")
				}
				instance, used, err = c.instantiateStaticConstruction(rec.staticConstructor, inherited, active, dynamic)
			} else {
				instance, used, err = c.instantiateStaticAction(clause, inherited, active, dynamic)
			}
			if err != nil {
				return err
			}
			if instance.result != nil && instance.Expression == "" && instance.Words == nil {
				return lerrors.TypeMismatch.Errorf("source action requires an executable definition")
			}
			forwarded = forwarded || used
			for _, step := range definitionSteps(instance) {
				expressions = append(expressions, step.Expression)
				body.sequence = append(body.sequence, step)
			}
			body.Bindings = append(body.Bindings, instance.Bindings...)
			body.Dependencies = append(body.Dependencies, instance.Dependencies...)
			if err := compose(clause.clauses); err != nil {
				return err
			}
		}
		return nil
	}
	err := compose(object.directivePhrase().clauses)
	body.Expression = strings.Join(expressions, "*")
	return body, forwarded, err
}

// A directive supplies an ordered composition to its selected constructor.
// Its clauses retain the lexical Function, source order, and ordinary instruction
// lowering; the constructor owns how that composition is used.
func (c *Compilation) directiveInstructions(object *objectGram) ([]*instruction, bool, error) {
	if object == nil || object.directivePhrase() == nil {
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
	body, err := builder.clauses(object.directivePhrase().clauses, ctx)
	return body, true, err
}
