package Lyger

import (
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

func (v declarationVisibility) destination(local, root *lexicalScope) (*lexicalScope, error) {
	switch v {
	case visibilityDefault:
		return local, nil
	case visibilityPublic:
		return root, nil
	default:
		return nil, lerrors.Internal.Errorf("invalid declaration visibility %d", v)
	}
}

type declarationPublication struct {
	declaration extractedDeclaration
	staged      record
	value       val
	reservation *scopeDeclaration
	destination *lexicalScope
	owner       *record
	namespace   *dictionary
}

// publishRegistration connects grammar declarations to their reserved source
// bindings. Public adds an index to the same record, never a second value.
func (c *Compilation) publishRegistration(stmt *lStatement, prepared statementRegistration) error {
	positions := make(map[*token]int, len(stmt.tokens))
	for index, tok := range stmt.tokens {
		positions[tok] = index
	}
	var pending []declarationPublication
	for _, declaration := range prepared.declarations {
		tok, rec := declaration.Token, declaration.Record
		index, exists := positions[tok]
		if !exists || tok == nil || !tok.isDeclare() || rec == nil {
			return lerrors.Internal.Errorf("declaration token does not belong to its source statement")
		}
		if c.Scope == nil {
			return lerrors.Internal.Errorf("publication requires source scope")
		}
		reservation := c.Scope.declarations[tok]
		if reservation == nil {
			return lerrors.Internal.Errorf("declaration %q has no source binding", tok.rawBytes)
		}
		staged := *rec
		rec = &staged
		if rec.Value != nil && rec.Value.Lexeme != nil && rec.Value.Lexeme.Type == roleOutput.String() {
			ref := rec.Value.Runtime
			if ref == nil || ref.Member != memberOutput {
				return lerrors.Internal.Errorf("output assignment has no member owner")
			}
			rec.Value = &entry{Type: storagePrimitive, Primitive: newPrimitive(valueText), Runtime: ref}
			rec.Type = "primitive.text"
		}
		value, err := rec.Value.value()
		if err != nil {
			return err
		}
		id, err := c.resolver().declarationID(stmt, index)
		if err != nil {
			return err
		}
		destination, err := declaration.Visibility.destination(reservation.Scope, reservation.Scope.unitRoot())
		if err != nil {
			return err
		}
		var owner *record
		var namespace *dictionary
		if reservation.Owner != nil {
			resolved, ready, err := c.resolver().owner(reservation.Owner)
			if err != nil {
				return err
			}
			if !ready {
				return lerrors.UnresolvedReference.Errorf("member declaration owner is unresolved")
			}
			owner = resolved.Record
			if owner != nil && owner.storageScope != nil && owner.storageScope.unit != nil && owner.storageScope.unit.Kind == SourceDictionary && owner.storageScope.unit != reservation.Scope.unit {
				return lerrors.Unsupported.Errorf("cannot modify dictionary %s member %s from another source unit", owner.storageScope.unit.Origin.Name, id)
			}
			if owner != nil {
				ownerValue, err := owner.Value.value()
				if err != nil {
					return err
				}
				if ownerValue.Type == valueDictionary {
					namespace = ownerValue.Data.(*dictionary) //nolint:errcheck // value() validates the tagged payload before this type-specific projection.
				}
				if declaration.Visibility != visibilityPublic {
					destination = owner.storageScope
				}
			}
		}
		if destination == nil {
			destination = c.Scope.Root
		}
		var prior *record
		for name, existing := range destination.records {
			if strings.EqualFold(name, id) {
				prior = existing
				break
			}
		}
		for _, candidate := range pending {
			if candidate.destination == destination && strings.EqualFold(candidate.staged.ID, id) {
				prior = candidate.declaration.Record
				break
			}
		}
		for _, existing := range destination.records {
			previous := c.Semantics.Declarations[existing]
			priorImport := previous != nil && previous.source != nil && previous.source.imported != nil
			if stmt.source.imported == nil && !priorImport {
				continue
			}
			collision := strings.EqualFold(existing.ID, id)
			if !collision {
				_, collision = existing.binding(newBindingQuery([]byte(id)))
				if !collision {
					_, collision = staged.binding(newBindingQuery([]byte(existing.ID)))
				}
			}
			if !collision {
				continue
			}
			old, err := canonicalRecord(existing, c.Semantics.Aliases)
			if err != nil {
				return err
			}
			target, err := canonicalRecord(declaration.Alias, c.Semantics.Aliases)
			if err != nil {
				return err
			}
			if old != target {
				if previous != nil && previous.source != nil {
					return lerrors.DuplicateDeclaration.Errorf("conflicting import %s (previous declaration at %s:%d:%d)", id, previous.source.Span.Path, previous.source.Span.Start.Line, previous.source.Span.Start.Column)
				}
				return lerrors.DuplicateDeclaration.Errorf("conflicting import %s", id)
			}
		}

		if prior != nil && stmt.source.imported == nil && (destination.isParagraph() || destination.kind == lexicalDictionary || declaration.Visibility == visibilityPublic) {
			if source := c.Semantics.Declarations[prior]; source != nil && source.source != nil {
				return lerrors.DuplicateDeclaration.Errorf("duplicate declaration %s (previous declaration at %s:%d:%d)", id, source.source.Span.Path, source.source.Span.Start.Line, source.source.Span.Start.Column)
			}
			return lerrors.DuplicateDeclaration.Errorf("duplicate declaration %s", id)
		}
		staged.ID, staged.storageScope = id, destination
		pending = append(pending, declarationPublication{declaration, staged, value, reservation, destination, owner, namespace})
	}
	for _, descriptor := range prepared.descriptors {
		if descriptor == nil || descriptor.Record == nil || descriptor.Targets == nil {
			return lerrors.Internal.Errorf("invalid prepared descriptor")
		}
		known := c.Semantics.Descriptors[descriptor.Record] != nil
		for _, publication := range pending {
			if publication.declaration.Record == descriptor.Record {
				known = true
				break
			}
		}
		if !known {
			return lerrors.Internal.Errorf("descriptor %s has no prepared or published declaration", descriptor.Record.ID)
		}
	}
	for _, publication := range pending {
		publication.commit(c, stmt)
	}
	for _, descriptor := range prepared.descriptors {
		c.Semantics.Descriptors[descriptor.Record] = descriptor
	}
	for _, function := range prepared.functions {
		if function.Native && function.Constructor.Name == "" {
			function.Constructor.Name = sourceModuleName(strings.ToLower(function.Record.ID))
		}
		function.Record.function = function
		if !function.Native {
			c.Semantics.Functions[function.Record] = function
		}
		function.Scope.function = function
	}
	return nil
}

func (p declarationPublication) commit(c *Compilation, stmt *lStatement) {
	rec := p.declaration.Record
	*rec = p.staged
	if p.declaration.Alias != nil {
		c.Semantics.Aliases[rec] = p.declaration.Alias
	}
	id := rec.ID
	p.destination.records[id] = rec
	if p.declaration.Visibility == visibilityPublic && p.destination != p.reservation.Scope {
		p.reservation.Scope.records[id] = rec
	}
	if p.owner != nil {
		rec.declarationOwner = p.owner
		if p.owner.members == nil {
			p.owner.members = make(map[string]*record)
		}
		p.owner.members[id] = rec
	}
	if p.value.Type == valueDictionary && c.Semantics.Aliases[rec] == nil {
		p.value.Data.(*dictionary).ID = id //nolint:errcheck // Declaration preparation validates the tagged payload before this atomic commit.
	}
	if p.namespace != nil {
		name := id[strings.LastIndexByte(id, '.')+1:]
		p.namespace.Values[name], p.namespace.Types[name] = p.value.Data, p.value.Type.String()
	}
	p.reservation.Record = rec
	c.Semantics.Subjects[p.declaration.Token], c.Semantics.Declarations[rec] = rec, stmt
	stmt.records = append(stmt.records, rec)
	if p.value.Type == valueApp && c.Semantics.Aliases[rec] == nil && rec.Value.Runtime == nil {
		app := p.value.Data.(*App) //nolint:errcheck // Declaration preparation validates the tagged payload before this atomic commit.
		app.ID = id
		c.Semantics.Apps = append(c.Semantics.Apps, app)
	}
}
