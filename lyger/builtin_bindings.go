package Lyger

import (
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// Builtin lookup uses the selected source unit's lexical records. The program
// cannot replace a compiler composition by shadowing its word, and a missing
// resource definition has no compiler-owned fallback.
func (c *Compilation) builtinRecord(name string) (*record, error) {
	if c == nil || c.Semantics == nil {
		return nil, lerrors.Internal.Errorf("builtin binding requires an analyzed compilation")
	}
	path := strings.Split(name, ".")
	for _, unit := range c.SourceUnits {
		if unit == nil || unit.Kind != SourceBuiltin && (unit.Kind != SourceDictionary || !strings.EqualFold(path[0], unit.Dictionary)) {
			continue
		}
		if unit.root == nil {
			return nil, lerrors.Internal.Errorf("builtin source has no lexical environment")
		}
		binding, state, err := unit.root.lookupLocal([]byte(path[0]))
		if err != nil {
			return nil, err
		}
		if state != lookupFound || binding == nil {
			continue
		}
		rec, err := canonicalRecord(binding.Record, c.Semantics.Aliases)
		if err != nil {
			return nil, err
		}
		for _, member := range path[1:] {
			var declared *record
			for identity, candidate := range rec.members {
				if strings.EqualFold(identity, rec.ID+"."+member) {
					declared = candidate
					break
				}
			}
			if declared != nil {
				rec, err = canonicalRecord(declared, c.Semantics.Aliases)
				if err != nil {
					return nil, err
				}
				continue
			}
			binding, _, err := c.resolver().lookupMember(rec, []byte(member))
			if err != nil {
				return nil, err
			}
			if binding == nil {
				return nil, lerrors.InvalidConfiguration.Errorf("builtin definition %s is absent", name)
			}
			rec, err = canonicalRecord(binding.Record, c.Semantics.Aliases)
			if err != nil {
				return nil, err
			}
		}
		declaration := c.Semantics.Declarations[rec]
		if declaration == nil || declaration.source.Unit != unit {
			return nil, lerrors.InvalidConfiguration.Errorf("builtin definition %s has no selected source declaration", name)
		}
		return rec, nil
	}
	return nil, lerrors.InvalidConfiguration.Errorf("builtin definition %s is absent", name)
}

func (c *Compilation) instantiateBuiltin(name string, arguments map[string]DefinitionArgument) (DefinitionInstance, error) {
	return c.instantiateDescribedBuiltin(name, nil, arguments)
}

func (c *Compilation) instantiateDescribedBuiltin(name string, descriptors []string, arguments map[string]DefinitionArgument) (DefinitionInstance, error) {
	if len(c.definitionTemporaries) == 0 {
		defer c.definitionScope()()
	}
	rec, err := c.builtinRecord(name)
	if err != nil {
		return DefinitionInstance{}, err
	}
	used := make(map[string]bool)
	node := &constructionNode{Definition: rec, Nominal: rec.ID}
	for _, name := range descriptors {
		modifier, err := c.builtinRecord(name)
		if err != nil {
			return DefinitionInstance{}, err
		}
		descriptor := c.Semantics.Descriptors[modifier]
		if descriptor == nil {
			return DefinitionInstance{}, lerrors.TypeMismatch.Errorf("%s is not a descriptor", name)
		}
		matched := false
		for _, rule := range descriptor.Rules {
			if rule.Target != (nominalIdentity{definition: rec}) {
				continue
			}
			matched = true
			if rule.Template != nil {
				node.Wrappers = append(node.Wrappers, rule)
				continue
			}
			selected, err := c.constructObject(rule.Object, make(map[*descriptorRule]bool))
			if err != nil {
				return DefinitionInstance{}, err
			}
			node.mergeDescriptor(selected)
		}
		if !matched {
			return DefinitionInstance{}, lerrors.TypeMismatch.Errorf("%s does not describe %s", name, rec.ID)
		}
	}
	bound, err := c.instantiateNamedConstruction(node, arguments, used, make(map[*record]bool))
	if err != nil {
		return DefinitionInstance{}, err
	}
	for name := range arguments {
		if !used[name] {
			return DefinitionInstance{}, lerrors.InvalidInput.Errorf("definition %s received unused argument %s", rec.ID, name)
		}
	}
	return bound, nil
}

// Named arguments remain typed data through nested source constructions. Each
// selected leaf consumes only its declared slots; the enclosing composition
// validates that no caller argument was silently discarded.
func namedDefinitionBindings(template *DefinitionTemplate, arguments map[string]DefinitionArgument, used map[string]bool) map[string]DefinitionArgument {
	bindings := make(map[string]DefinitionArgument, len(template.Slots))
	for _, slot := range template.Slots {
		if argument, exists := arguments[slot.Name]; exists {
			bindings[slot.Name] = argument
			used[slot.Name] = true
		}
	}
	return bindings
}

func (c *Compilation) instantiateNamedDefinition(rec *record, arguments map[string]DefinitionArgument, used map[string]bool, active map[*record]bool) (DefinitionInstance, error) {
	rec, err := canonicalRecord(rec, c.Semantics.Aliases)
	if err != nil {
		return DefinitionInstance{}, err
	}
	if rec == nil {
		return DefinitionInstance{}, lerrors.Internal.Errorf("definition has no canonical record")
	}
	if function := rec.function; function != nil && function.Native {
		values := make([]DefinitionArgument, len(function.Constructor.Inputs))
		for i, input := range function.Constructor.Inputs {
			found := false
			for name, argument := range arguments {
				if !strings.EqualFold(name, input.Key) {
					continue
				}
				if found {
					return DefinitionInstance{}, lerrors.InvalidInput.Errorf("ambiguous argument for %s", input.Key)
				}
				found = true
				values[i] = argument
				used[name] = true
			}
			if !found {
				return DefinitionInstance{}, lerrors.InvalidInput.Errorf("missing argument for %s", input.Key)
			}
		}
		return c.invokeFunctionDefinition(function, values, active, nil)
	}
	if active[rec] {
		return DefinitionInstance{}, lerrors.Cycle.Errorf("recursive definition invocation: %s", rec.ID)
	}
	active[rec] = true
	defer delete(active, rec)
	template, err := c.recordTemplate(rec)
	if err != nil {
		return DefinitionInstance{}, err
	}
	if template != nil {
		bound, err := template.Instantiate(namedDefinitionBindings(template, arguments, used))
		bound.Symbol, bound.definition = rec.ID, rec
		return bound, err
	}
	return c.composeSourceActions(rec, func(action *clauseGram) (DefinitionInstance, error) {
		var values []DefinitionArgument
		var ledger []DefinitionBinding
		var dependencies []DefinitionInstance
		for _, object := range action.orderedObjects() {
			child, err := c.instantiateNamedObject(object, arguments, used, active)
			if err != nil {
				return DefinitionInstance{}, err
			}
			values = append(values, DefinitionArgument{Kind: definitionArgDefinition.String(), Definition: &child})
			ledger = append(ledger, child.Bindings...)
			dependencies = append(dependencies, child.Dependencies...)
		}
		instance, err := c.instantiateStaticCall(action.verb.binding.Record, values, active, nil)
		instance.Bindings = append(ledger, instance.Bindings...)
		instance.Dependencies = append(dependencies, instance.Dependencies...)
		return instance, err
	})
}

func (c *Compilation) instantiateNamedObject(object *objectGram, arguments map[string]DefinitionArgument, used map[string]bool, active map[*record]bool) (DefinitionInstance, error) {
	node, err := c.constructObject(object, make(map[*descriptorRule]bool))
	if err != nil {
		return DefinitionInstance{}, err
	}
	local := arguments
	localUsed := used
	if len(node.Arguments) > 0 {
		var values []DefinitionArgument
		for _, parameter := range node.Arguments {
			value, _, err := c.definitionOperand(parameter, active, nil, func(child *objectGram) (DefinitionInstance, bool, error) {
				instance, err := c.instantiateNamedObject(child, arguments, used, active)
				return instance, false, err
			})
			if err != nil {
				return DefinitionInstance{}, err
			}
			values = append(values, value)
		}
		template, err := c.recordTemplate(node.Definition)
		if err != nil {
			return DefinitionInstance{}, err
		}
		if function := node.Definition.function; function != nil && function.Native {
			return c.invokeFunctionDefinition(function, values, active, nil)
		}
		if template == nil {
			return DefinitionInstance{}, lerrors.TypeMismatch.Errorf("constructor %s requires a declared parameter signature", node.Nominal)
		}
		local, err = definitionArguments(template, values)
		if err != nil {
			return DefinitionInstance{}, err
		}
		localUsed = make(map[string]bool)
	}
	return c.instantiateNamedConstruction(node, local, localUsed, active)
}

func (c *Compilation) instantiateNamedConstruction(node *constructionNode, local map[string]DefinitionArgument, localUsed map[string]bool, active map[*record]bool) (DefinitionInstance, error) {
	wrappers, selected, err := constructionTemplates(node)
	if err != nil {
		return DefinitionInstance{}, err
	}
	var instance DefinitionInstance
	if selected < 0 {
		instance, err = c.instantiateNamedDefinition(node.Definition, local, localUsed, active)
		selected = len(wrappers)
	} else {
		wrapper := wrappers[selected]
		instance, err = wrapper.Instantiate(namedDefinitionBindings(wrapper, local, localUsed))
	}
	if err != nil {
		return DefinitionInstance{}, err
	}
	for i := selected - 1; i >= 0; i-- {
		wrapper := wrappers[i]
		bindings := map[string]DefinitionArgument{wrapper.Slots[0].Name: {Kind: definitionArgDefinition.String(), Definition: &instance}}
		instance, err = wrapper.Instantiate(bindings)
		if err != nil {
			return DefinitionInstance{}, err
		}
	}
	instance.Symbol = node.Definition.ID
	instance.definition = node.Definition
	return instance, nil
}

func constructionTemplates(node *constructionNode) ([]*DefinitionTemplate, int, error) {
	var err error
	wrappers := make([]*DefinitionTemplate, len(node.Wrappers))
	selected := -1
	for i, rule := range node.Wrappers {
		wrappers[i], err = ParseDefinitionTemplate(rule.Template.rawBytes, rule.Source.source.Span)
		if err != nil {
			return nil, -1, err
		}
		wrapper := wrappers[i]
		nominal := rule.Target.String()
		nominal = nominal[strings.LastIndex(nominal, ".")+1:]
		if len(wrapper.Slots) != 1 || wrapper.Slots[0].Type != "" || !strings.EqualFold(wrapper.Slots[0].Name, nominal) {
			if selected >= 0 {
				return nil, -1, lerrors.InvalidInput.Errorf("constructor %s has competing descriptor definitions", node.Nominal)
			}
			selected = i
		}
	}
	return wrappers, selected, nil
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
