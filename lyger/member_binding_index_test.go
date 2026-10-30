package Lyger

import (
	"errors"
	"testing"

	"github.com/l-programming-language/l/lerrors"
)

func TestMemberBindingIndexFollowsPublicationAndLifetime(t *testing.T) {
	owner := qualifiedDictionaryRecord("Tools")
	alias := qualifiedDictionaryRecord("Alias")
	scope := &lexicalScope{records: map[string]*record{owner.ID: owner}}
	plan := &scopePlan{Root: scope}
	semantics := newSemanticProgram()
	semantics.Aliases[alias] = owner
	resolver := semanticResolver{scopes: plan, semantics: semantics}
	release := plan.indexBindings()
	defer release()
	if binding, state, err := resolver.lookupMember(alias, []byte("Writes")); binding != nil || state != lookupPending || err != nil {
		t.Fatalf("unpublished member = %v, %s, %v", binding, state, err)
	}
	publish := func(rec *record) {
		publication := declarationPublication{
			declaration: extractedDeclaration{Record: rec}, staged: *rec,
			value:       val{Type: valueLexeme, Data: rec.Value.Lexeme},
			reservation: &scopeDeclaration{}, destination: scope, owner: owner,
		}
		publication.commit(&Compilation{Scope: plan, Semantics: semantics}, &lStatement{})
	}
	first := indexedTestWord(t, "Tools.Write", `"Write|-s"`, false)
	publish(first)
	if binding, state, err := resolver.lookupMember(alias, []byte("WRITES’")); err != nil || state != lookupFound || binding.Record != first || !binding.Possessive {
		t.Fatalf("published alias member = %v, %s, %v", binding, state, err)
	}
	replacement := indexedTestWord(t, "Tools.Write", `"Write|--ten"`, false)
	publish(replacement)
	if binding, state, err := resolver.lookupMember(owner, []byte("Writes")); err != nil || state != lookupPending || binding != nil {
		t.Fatalf("retired member morphology = %v, %s, %v", binding, state, err)
	}
	if binding, state, err := resolver.lookupMember(owner, []byte("Written")); err != nil || state != lookupFound || binding.Record != replacement {
		t.Fatalf("replacement morphology = %v, %s, %v", binding, state, err)
	}
	release()
	if plan.memberBindings != nil {
		t.Fatal("member index survived semantic analysis")
	}
	owner.members[first.ID] = first
	if binding, state, err := resolver.lookupMember(owner, []byte("Writes")); err != nil || state != lookupFound || binding.Record != first {
		t.Fatalf("post-analysis member mutation = %v, %s, %v", binding, state, err)
	}
}

func TestMemberBindingIndexRetainsNamespacesAndAmbiguity(t *testing.T) {
	left, right := qualifiedDictionaryRecord("Left"), qualifiedDictionaryRecord("Right")
	left.members["a"] = indexedTestWord(t, "Left.Write", `"Write|-s"`, false)
	left.members["z"] = indexedTestWord(t, "Left.Other", `"Write|-s"`, false)
	right.members["a"] = indexedTestWord(t, "Right.Write", `"Write|-s"`, false)
	plan := &scopePlan{Root: &lexicalScope{records: map[string]*record{}}}
	release := plan.indexBindings()
	defer release()
	resolver := semanticResolver{scopes: plan, semantics: newSemanticProgram()}
	_, _, gotErr := resolver.lookupMember(left, []byte("Writes"))
	_, wantErr := resolveTokenBinding([]byte("Left.Writes"), left.members)
	if !errors.Is(gotErr, lerrors.AmbiguousReference) || wantErr == nil || gotErr.Error() != wantErr.Error() {
		t.Fatalf("indexed member ambiguity = %v; direct = %v", gotErr, wantErr)
	}
	if binding, state, err := resolver.lookupMember(right, []byte("Writes")); err != nil || state != lookupFound || binding.Record != right.members["a"] {
		t.Fatalf("other owner inherited ambiguity = %v, %s, %v", binding, state, err)
	}
}
