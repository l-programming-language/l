package Lyger

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// semanticResolver borrows the compilation's bindings. Source spellings and
// lexical forms remain on tokens; ownership follows declaration/alias identity.
type semanticResolver struct {
	scopes    *scopePlan
	semantics *semanticProgram
}

type semanticOwner struct {
	Record    *record
	Namespace string // A native type namespace has no instance Record.
}

type memberResolution struct {
	Binding *tokenBinding
	Value   *entry
	Type    string
}

// nominalIdentity compares source definitions by canonical declaration, and
// native targets by their existing value/POS tags. Names remain diagnostic text.
type nominalIdentity struct {
	definition *record
	native     valueType
	lexical    partOfSpeech
}

func (n nominalIdentity) String() string {
	if n.definition != nil {
		return strings.ToLower(n.definition.ID)
	}
	if n.lexical != posInvalid {
		return n.lexical.String()
	}
	return strings.ToLower(n.native.String())
}

func (r semanticResolver) nominalIdentity(tok *token) (nominalIdentity, error) {
	if tok == nil || tok.mem == nil {
		return nominalIdentity{}, lerrors.Internal.Errorf("nominal target requires a typed token")
	}
	if _, err := tok.mem.value(); err != nil {
		return nominalIdentity{}, err
	}
	rec, err := r.reference(tok)
	if err != nil {
		return nominalIdentity{}, err
	}
	if rec != nil {
		return nominalIdentity{definition: rec}, nil
	}
	name, category := checkType(tok.rawBytes)
	if category == storageLexeme {
		if pos, ok := parsePartOfSpeech(strings.ToLower(name)); ok {
			return nominalIdentity{native: valueLexeme, lexical: pos}, nil
		}
	} else if category != storageUnknown {
		if kind, ok := parseValueType(name); ok && kind != valueInvalid && kind != valueUnknown {
			return nominalIdentity{native: kind}, nil
		}
	}
	return nominalIdentity{}, lerrors.UnresolvedReference.Errorf("unresolved nominal target %q", tok.rawBytes)
}

func (c *Compilation) resolver() semanticResolver {
	return semanticResolver{scopes: c.Scope, semantics: c.Semantics}
}

func canonicalRecord(rec *record, aliases map[*record]*record) (*record, error) {
	seen := make(map[*record]bool)
	for rec != nil {
		if seen[rec] {
			return nil, lerrors.Cycle.Errorf("alias cycle at %s", rec.ID)
		}
		seen[rec] = true
		next, aliased := aliases[rec]
		if !aliased {
			return rec, nil
		}
		if next == nil {
			return nil, lerrors.Internal.Errorf("alias %s has no target", rec.ID)
		}
		rec = next
	}
	return nil, nil
}

func (r semanticResolver) lookup(tok *token) (*tokenBinding, lookupState, error) {
	scope := tok.scope
	if scope == nil {
		return nil, lookupAbsent, lerrors.Internal.Errorf("reference %q has no lexical scope", tok.rawBytes)
	}
	for ; scope != nil; scope = scope.parent {
		binding, state, err := scope.lookupLocal(tok.rawBytes)
		if err != nil || state != lookupAbsent {
			return binding, state, err
		}
		if function := scope.function; function != nil && function.Return != nil {
			prefix := function.Return.ID[:strings.LastIndex(function.Return.ID, ".")+1]
			if binding, ok := function.Return.binding(newBindingQuery([]byte(prefix + string(tok.rawBytes)))); ok {
				return binding, lookupFound, nil
			}
		}
		// A local owner commits the remaining path to that namespace, even
		// while the owner or one of its members is awaiting publication.
		for dot := bytes.LastIndexByte(tok.rawBytes, '.'); dot > 0; dot = bytes.LastIndexByte(tok.rawBytes[:dot], '.') {
			owner, state, err := scope.lookupLocal(tok.rawBytes[:dot])
			if err != nil || state == lookupPending {
				return nil, state, err
			}
			if state == lookupFound {
				return r.lookupMember(owner.Record, tok.rawBytes[dot+1:])
			}
		}
	}
	return nil, lookupAbsent, nil
}

// lookupMember never reopens lexical lookup after an owner has been selected.
func (r semanticResolver) lookupMember(owner *record, tail []byte) (*tokenBinding, lookupState, error) {
	owner, err := canonicalRecord(owner, r.semantics.Aliases)
	if err != nil {
		return nil, lookupPending, err
	}
	if owner == nil {
		return nil, lookupPending, nil
	}
	qualified := owner.ID + "." + string(tail)
	binding, err := r.scopes.resolveMemberBinding(owner, []byte(qualified))
	if err != nil || binding != nil {
		return binding, lookupFound, err
	}
	for dot := bytes.LastIndexByte(tail, '.'); dot > 0; dot = bytes.LastIndexByte(tail[:dot], '.') {
		prefix := []byte(owner.ID + "." + string(tail[:dot]))
		member, err := r.scopes.resolveMemberBinding(owner, prefix)
		if err != nil {
			return nil, lookupPending, err
		}
		if member != nil {
			return r.lookupMember(member.Record, tail[dot+1:])
		}
	}
	return nil, lookupPending, nil
}

func (r semanticResolver) bindToken(tok *token) (lookupState, error) {
	binding, state, err := r.lookup(tok)
	tok.bindingError = err
	if err != nil {
		return state, err
	}
	if state == lookupPending {
		tok.binding, tok.bindingError, tok.defined = nil, nil, false
		tok.tType = "primitive.unknown"
		tok.mem = &entry{Type: storagePrimitive, Primitive: newPrimitive(valueUnknown)}
	} else if state == lookupFound {
		tok.mem = &entry{}
		tok.mem.bindFrom(binding.Record.Value)
		tok.binding, tok.tType, tok.defined = binding, binding.lexicalType(), true
		tok.bindingError = nil
	}
	return state, nil
}

// reference returns nil for a dependency that has not been published yet.
// Action arguments require the call's instruction context, never global lookup.
func (r semanticResolver) reference(tok *token) (*record, error) {
	if tok == nil {
		return nil, nil
	}
	if r.scopes != nil {
		if binding, ok := r.scopes.Pronouns[tok]; ok {
			if binding.Kind == scopeActionArgument {
				return nil, lerrors.UnresolvedReference.Errorf("unbound action argument %s", binding.SubjectID)
			}
			return canonicalRecord(r.semantics.Subjects[binding.TargetToken], r.semantics.Aliases)
		}
	}
	if tok.isDeclare() {
		return canonicalRecord(r.semantics.Subjects[tok], r.semantics.Aliases)
	}
	if tok.binding != nil {
		return canonicalRecord(tok.binding.Record, r.semantics.Aliases)
	}
	binding, _, err := r.lookup(tok)
	if err != nil || binding == nil {
		return nil, err
	}
	return canonicalRecord(binding.Record, r.semantics.Aliases)
}

func (r semanticResolver) owner(tok *token) (semanticOwner, bool, error) {
	rec, err := r.reference(tok)
	if err != nil {
		return semanticOwner{}, false, err
	}
	if rec != nil {
		return semanticOwner{Record: rec, Namespace: rec.ID}, true, nil
	}
	if tok != nil && !tok.isPronoun() {
		_, state, err := r.lookup(tok)
		if err != nil || state == lookupPending {
			return semanticOwner{}, false, err
		}
		name, _ := possessiveOwner(string(tok.rawBytes))
		if kind, _ := checkType([]byte(name)); kind != "Unknown" {
			return semanticOwner{Namespace: kind}, true, nil
		}
	}
	return semanticOwner{}, false, nil
}

func nativeMemberFor(ownerType valueType, lex *lexeme) (nativeMember, bool) {
	if ownerType != valueApp || lex == nil {
		return memberInvalid, false
	}
	role, _ := parseNativeRole(lex.Type)
	return role.Member()
}

func (r semanticResolver) member(owner semanticOwner, tok *token, source *lStatement) (memberResolution, bool, error) {
	qualified := []byte(owner.Namespace + "." + string(tok.rawBytes))
	var binding *tokenBinding
	var err error
	if owner.Record != nil {
		binding, _, err = r.lookupMember(owner.Record, tok.rawBytes)
	} else {
		query := *tok
		query.rawBytes = qualified
		binding, _, err = r.lookup(&query)
	}
	if err != nil {
		return memberResolution{}, false, err
	}
	if binding != nil {
		value := &entry{}
		value.bindFrom(binding.Record.Value)
		return memberResolution{Binding: binding, Value: value, Type: binding.lexicalType()}, true, nil
	}
	// Native members require a concrete application, not a type namespace or
	// a same-named global definition left on the token by an earlier pass.
	if owner.Record != nil && owner.Record.Value != nil {
		app := owner.Record.Value.Applicative
		if app != nil && app.Type == valueApp && app.App != nil {
			lex, category, err := tok.gram.lex.lookup(tok.rawBytes)
			if err == nil {
				if spec, ok := nativeMemberFor(valueApp, lex); ok {
					value := &entry{Type: storageLexeme, Lexeme: lex.cloneUse(), Runtime: &runtimeReference{
						Owner: owner.Record, OwnerID: strings.ToLower(owner.Record.ID), Member: spec, Source: source,
					}}
					return memberResolution{Value: value, Type: category}, true, nil
				}
			}
		}
	}
	return memberResolution{}, false, nil
}

func (r semanticResolver) declarationID(statement *lStatement, index int) (string, error) {
	label, err := strconv.Unquote(string(statement.tokens[index].rawBytes))
	if err != nil {
		return "", lerrors.Syntax.Wrap(err, "")
	}
	root := strings.SplitN(label, "|", 2)[0]
	if index > 0 {
		previous := statement.tokens[index-1]
		if _, possessed := possessiveOwner(string(previous.rawBytes)); possessed {
			owner, ready, err := r.owner(previous)
			if err != nil {
				return "", err
			}
			if !ready {
				return "", lerrors.UnresolvedReference.Errorf("member declaration owner %q is unresolved", previous.rawBytes)
			}
			return owner.Namespace + "." + root, nil
		}
	}
	return root, nil
}
