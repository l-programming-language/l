package Lyger

import "github.com/l-programming-language/l/lerrors"

// Builtin lookup uses the selected source unit's lexical records. The program
// cannot replace a compiler composition by shadowing its word, and a missing
// resource definition has no compiler-owned fallback.
func (c *Compilation) builtinRecord(name string) (*record, error) {
	if c == nil || c.Semantics == nil {
		return nil, lerrors.Internal.Errorf("builtin binding requires an analyzed compilation")
	}
	for _, unit := range c.SourceUnits {
		if unit == nil || unit.Kind != SourceBuiltin {
			continue
		}
		if unit.root == nil {
			return nil, lerrors.Internal.Errorf("builtin source has no lexical environment")
		}
		binding, state, err := unit.root.lookupLocal([]byte(name))
		if err != nil {
			return nil, err
		}
		if state != lookupFound || binding == nil {
			return nil, lerrors.InvalidConfiguration.Errorf("builtin definition %s is absent", name)
		}
		rec, err := canonicalRecord(binding.Record, c.Semantics.Aliases)
		if err != nil {
			return nil, err
		}
		declaration := c.Semantics.Declarations[rec]
		if declaration == nil || declaration.source.Unit != unit {
			return nil, lerrors.InvalidConfiguration.Errorf("builtin definition %s has no selected source declaration", name)
		}
		return rec, nil
	}
	return nil, lerrors.InvalidConfiguration.Errorf("compilation has no builtin source unit")
}

func (c *Compilation) builtinDefinition(name string) (*DefinitionTemplate, error) {
	rec, err := c.builtinRecord(name)
	if err != nil {
		return nil, err
	}
	template, err := c.recordTemplate(rec)
	if err == nil && template == nil {
		err = lerrors.InvalidConfiguration.Errorf("builtin %s has no logeme definition", name)
	}
	return template, err
}

func (c *Compilation) instantiateBuiltin(name string, arguments map[string]DefinitionArgument) (DefinitionInstance, error) {
	rec, err := c.builtinRecord(name)
	if err != nil {
		return DefinitionInstance{}, err
	}
	template, err := c.recordTemplate(rec)
	if err != nil {
		return DefinitionInstance{}, err
	}
	if template == nil {
		return DefinitionInstance{}, lerrors.InvalidConfiguration.Errorf("builtin %s has no logeme definition", name)
	}
	bound, err := template.Instantiate(arguments)
	if err != nil {
		return DefinitionInstance{}, err
	}
	bound.Symbol, bound.definition = rec.ID, rec
	return bound, nil
}

func (c *Compilation) builtinMemberDefinition(owner, member string) (*DefinitionTemplate, error) {
	rec, err := c.builtinRecord(owner)
	if err != nil {
		return nil, err
	}
	binding, _, err := c.resolver().lookupMember(rec, []byte(member))
	if err != nil {
		return nil, err
	}
	if binding == nil {
		return nil, lerrors.InvalidConfiguration.Errorf("builtin definition %s has no %s member", owner, member)
	}
	template, err := c.recordTemplate(binding.Record)
	if err == nil && template == nil {
		err = lerrors.InvalidConfiguration.Errorf("builtin member %s.%s has no logeme definition", owner, member)
	}
	return template, err
}
