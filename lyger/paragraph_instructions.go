package Lyger

import (
	"github.com/l-programming-language/l/lerrors"
)

// scopedInstructions describes lexical lifetime, not an executable function.
// Every App associated with a paragraph shares its Scope but owns its Body.
type scopedInstructions struct {
	Scope *lexicalScope
	Body  []*instruction
}

type paragraphPlanner struct {
	builder   *instructionBuilder
	producers map[*record]*App
	led       map[*lexicalScope]bool
}

func (b *instructionBuilder) scopeInstructions(instructions []*instruction) ([]*instruction, error) {
	p := paragraphPlanner{builder: b, producers: make(map[*record]*App), led: make(map[*lexicalScope]bool)}
	b.semantics.Continuations = make(map[*App]*scopedInstructions)
	for _, ordered := range b.scopes.Ordered {
		if !ordered.Scope.isParagraph() {
			continue
		}
		if ordered.Source.Path == ordered.Scope.source.Path && ordered.Source.Start == ordered.Scope.source.Start && len(ordered.Declarations) > 0 {
			rec := b.semantics.Subjects[ordered.Declarations[0].Token]
			if rec != nil && b.semantics.Aliases[rec] == nil {
				kind, err := instructionType(rec)
				if err != nil {
					return nil, err
				}
				p.led[ordered.Scope] = kind == valueApp
			}
		}
	}
	root := &scopedInstructions{Scope: b.scopes.Root}
	blocks := map[*lexicalScope]*scopedInstructions{b.scopes.Root: root}
	var blockFor func(*lexicalScope, *instruction) *scopedInstructions
	blockFor = func(scope *lexicalScope, origin *instruction) *scopedInstructions {
		if block := blocks[scope]; block != nil {
			return block
		}
		parent := blockFor(scope.parent, origin)
		block := &scopedInstructions{Scope: scope}
		blocks[scope] = block
		parent.Body = append(parent.Body, &instruction{Kind: instructionBlock, Block: block, Source: origin.Source, Statement: origin.Statement, Order: origin.Order})
		return block
	}
	for _, current := range instructions {
		scope := b.ordered[current.Statement].Scope
		if scope == nil {
			return nil, lerrors.Internal.Errorf("instruction has no lexical scope")
		}
		block := blockFor(scope, current)
		block.Body = append(block.Body, current)
	}
	if err := p.partition(root); err != nil {
		return nil, err
	}
	return root.Body, nil
}

func (p *paragraphPlanner) valueOwner(value *instructionValue) (*App, error) {
	if value == nil {
		return nil, nil
	}
	if value.Owner != nil {
		return p.valueOwner(value.Owner)
	}
	if value.Payload != nil {
		return p.valueOwner(value.Payload)
	}
	if value.Record == nil {
		return nil, nil
	}
	rec, err := canonicalRecord(value.Record, p.builder.semantics.Aliases)
	if err != nil {
		return nil, err
	}
	if app := p.producers[rec]; app != nil {
		return app, nil
	}
	kind, err := instructionType(rec)
	if err != nil {
		return nil, err
	}
	if kind == valueApp {
		return rec.Value.Applicative.App, nil
	}
	return nil, nil
}

func (p *paragraphPlanner) contextOwner(current *instruction, scope *lexicalScope) (*App, error) {
	ordered := p.builder.ordered[current.Statement]
	if tok := ordered.subjectBefore(current.SourceToken); tok != nil && ordered.Scope == scope {
		rec, err := p.builder.reference(tok)
		if err != nil {
			return nil, err
		}
		if rec != nil {
			app, err := p.valueOwner(&instructionValue{Record: rec})
			if err != nil || app != nil {
				return app, err
			}
		}
	}
	// A scalar subject with no producer does not erase its enclosing App's
	// continuation. Pronoun resolution remains independently source-bound.
	var latest *App
	var subject *record
	for _, candidate := range p.builder.scopes.Ordered {
		if candidate.Order > ordered.Order {
			break
		}
		if candidate.Scope != scope {
			continue
		}
		declarations := candidate.Declarations
		if candidate.Statement == current.Statement {
			declarations = candidate.declarationsBefore(current.SourceToken)
		}
		for _, declaration := range declarations {
			rec := p.builder.semantics.Subjects[declaration.Token]
			if rec == nil {
				continue
			}
			subject = rec
			if p.builder.semantics.Aliases[rec] != nil {
				continue
			}
			kind, err := instructionType(rec)
			if err != nil {
				return nil, err
			}
			if kind == valueApp {
				latest = rec.Value.Applicative.App
			}
		}
	}
	if subject != nil {
		app, err := p.valueOwner(&instructionValue{Record: subject})
		if err != nil || app != nil {
			return app, err
		}
	}
	return latest, nil
}

func (p *paragraphPlanner) owner(current *instruction) (*App, error) {
	if current.Block != nil {
		// The child block's statements retain their own lexical environment;
		// its activation belongs to the importing parent at this source point.
		return p.contextOwner(current, current.Block.Scope.parent)
	}
	var candidates []*instructionValue
	switch current.Kind {
	case instructionIf:
		if current.Condition != nil {
			candidates = []*instructionValue{current.Condition.Left, current.Condition.Right}
		}
	case instructionAssign, instructionShow, instructionExit:
		candidates = []*instructionValue{current.Value}
	}
	for _, value := range candidates {
		app, err := p.valueOwner(value)
		if err != nil || app != nil {
			return app, err
		}
	}
	return p.contextOwner(current, p.builder.ordered[current.Statement].Scope)
}

func (p *paragraphPlanner) markAssignments(instructions []*instruction, owner *App) {
	for _, current := range instructions {
		if current.Kind == instructionAssign && current.Target != nil && current.Target.Record != nil {
			p.producers[current.Target.Record] = owner
		}
		p.markAssignments(current.Actions, owner)
		p.markAssignments(current.Then, owner)
		p.markAssignments(current.Else, owner)
		if current.Block != nil {
			p.markAssignments(current.Block.Body, owner)
		}
	}
}

func (p *paragraphPlanner) partition(block *scopedInstructions) error {
	var retained []*instruction
	for _, current := range block.Body {
		if current.Block != nil {
			if err := p.partition(current.Block); err != nil {
				return err
			}
			if p.led[current.Block.Scope] {
				continue
			}
		}
		if !p.led[block.Scope] {
			retained = append(retained, current)
			continue
		}
		owner, err := p.owner(current)
		if err != nil {
			return current.wrapError(err)
		}
		if owner == nil {
			return current.wrapError(lerrors.InvalidInput.Errorf("paragraph effect has no associated application"))
		}
		continuation := p.builder.semantics.Continuations[owner]
		if continuation == nil {
			continuation = &scopedInstructions{Scope: block.Scope}
			p.builder.semantics.Continuations[owner] = continuation
		} else if continuation.Scope != block.Scope {
			return current.wrapError(lerrors.DuplicateDeclaration.Errorf("application has behavior in multiple lexical paragraphs"))
		}
		continuation.Body = append(continuation.Body, current)
		p.markAssignments([]*instruction{current}, owner)
	}
	block.Body = retained
	return nil
}
