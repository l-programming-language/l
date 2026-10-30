package Lyger

import (
	"reflect"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// ComposeDefinition expands a selected source word after Make. It exposes the
// existing serialized composition to resource generation tools, without making
// the backend depend on a frontend object, a parser callback, or a new AML form.
func (c *Compilation) ComposeDefinition(name string) (DefinitionInstance, error) {
	rec, err := c.exportedRecord(name)
	if err != nil {
		return DefinitionInstance{}, err
	}
	return c.composeExport(rec)
}

func (c *Compilation) exportedRecord(name string) (*record, error) {
	if c == nil || c.Scope == nil || c.Semantics == nil {
		return nil, lerrors.InvalidInput.Errorf("definition export requires an analyzed compilation")
	}
	if c.Error != nil {
		return nil, c.Error
	}
	path := strings.Split(name, ".")
	binding, _, err := c.Scope.Root.lookupLocal([]byte(path[0]))
	if err != nil {
		return nil, err
	}
	if binding == nil {
		return nil, lerrors.UnresolvedReference.Errorf("definition %s is not exported", name)
	}
	rec, err := canonicalRecord(binding.Record, c.Semantics.Aliases)
	if err != nil {
		return nil, err
	}
	for _, member := range path[1:] {
		binding, _, err = c.resolver().lookupMember(rec, []byte(member))
		if err != nil {
			return nil, err
		}
		if binding == nil {
			return nil, lerrors.UnresolvedReference.Errorf("definition %s is not exported", name)
		}
		rec, err = canonicalRecord(binding.Record, c.Semantics.Aliases)
		if err != nil {
			return nil, err
		}
	}
	return rec, nil
}

func (c *Compilation) composeExport(rec *record) (DefinitionInstance, error) {
	rec, err := canonicalRecord(rec, c.Semantics.Aliases)
	if err != nil {
		return DefinitionInstance{}, err
	}
	if function := rec.function; function != nil && function.Native {
		return c.functionDefinition(function, make(map[*record]bool), nil)
	}
	bound, err := c.instantiateCall(rec, nil, make(map[*record]bool))
	// This copy identifies an exported declaration, not the nominal type of
	// an intermediate constructor used while expanding its body.
	bound.definition = rec
	return bound, err
}

// ComposeLibrary serializes a selected source composition and its declaration
// dependencies through the same AML renderer used by application compilation.
func (c *Compilation) ComposeLibrary(name, filename string) ([]byte, error) {
	if err := validateAMLName(filename); err != nil {
		return nil, err
	}
	rec, err := c.exportedRecord(name)
	if err != nil {
		return nil, err
	}
	roots := []*record{rec}
	if kind, err := instructionType(rec); err != nil {
		return nil, err
	} else if kind == valueDictionary {
		roots, err = c.publicExports(rec, make(map[*record]bool))
		if err != nil {
			return nil, err
		}
		if len(roots) == 0 {
			return nil, lerrors.InvalidInput.Errorf("Dictionary %s has no public source declarations", name)
		}
	}

	var declarations []DefinitionInstance
	for _, root := range roots {
		definition, err := c.composeExport(root)
		if err != nil {
			return nil, err
		}
		declarations = append(declarations, definition)
	}
	definitions, err := definitionClosure(declarations...)
	if err != nil {
		return nil, err
	}
	return renderLibrary(filename, filename, definitions)
}

func renderLibrary(name, filename string, definitions []DefinitionInstance) ([]byte, error) {
	file := &sentence{Name: filename, Words: []*w{{Syllables: []*s{{CompiledBytes: []byte("A")}}}}}
	for _, definition := range definitions {
		appendDefinitionWords(file, definition)
	}
	library := newRootService("library")
	library.Target = &sourceTarget{Kind: "lib", Name: name, Root: filename}
	library.Objects = []*object{newRootObject(filename)}
	app := &App{ID: name, Meta: appMetadata(name), Services: []*service{library}}
	program := &Program{Apps: []*App{app}, Books: []*book{{Name: name, Chapters: []*chapter{{Name: "library", sentences: []*sentence{file}}}}}}
	if err := program.Make(); err != nil {
		return nil, err
	}
	return program.BackendBytes, nil
}

func (c *Compilation) publicExports(owner *record, active map[*record]bool) ([]*record, error) {
	if active[owner] {
		return nil, lerrors.Cycle.Errorf("recursive public namespace %s", owner.ID)
	}
	active[owner] = true
	defer delete(active, owner)
	var result []*record
	for _, ordered := range c.Scope.Ordered {
		for _, token := range ordered.Statement.tokens {
			member := c.Semantics.Subjects[token]
			if member == nil || owner.members[member.ID] != member {
				continue
			}
			visibility, err := ordered.Statement.sGram.declarationVisibility(token)
			if err != nil {
				return nil, err
			}
			if visibility != visibilityPublic {
				continue
			}
			kind, err := instructionType(member)
			if err != nil {
				return nil, err
			}
			if kind != valueDictionary {
				result = append(result, member)
			}
			children, err := c.publicExports(member, active)
			if err != nil {
				return nil, err
			}
			result = append(result, children...)
		}
	}
	return result, nil
}

func definitionClosure(selected ...DefinitionInstance) ([]DefinitionInstance, error) {
	var ordered []DefinitionInstance
	seen := make(map[*record]DefinitionInstance)
	active := make(map[*record]bool)
	var visit func(DefinitionInstance) error
	visit = func(definition DefinitionInstance) error {
		if definition.definition == nil || definition.Expression == "" && len(definition.Words) == 0 {
			return lerrors.InvalidInput.Errorf("exported composition has no canonical declaration or expression")
		}
		if prior, exists := seen[definition.definition]; exists {
			if prior.Expression != definition.Expression || !reflect.DeepEqual(prior.Words, definition.Words) {
				return lerrors.InvalidInput.Errorf("definition %s has conflicting declaration expansions", definition.Symbol)
			}
			return nil
		}
		if active[definition.definition] {
			return lerrors.Cycle.Errorf("recursive declaration dependency at %s", definition.Symbol)
		}
		active[definition.definition] = true
		for _, dependency := range definition.Dependencies {
			if err := visit(dependency); err != nil {
				return err
			}
		}
		delete(active, definition.definition)
		seen[definition.definition] = definition
		ordered = append(ordered, definition)
		return nil
	}
	for _, root := range selected {
		if err := visit(root); err != nil {
			return nil, err
		}
	}
	return ordered, nil
}
