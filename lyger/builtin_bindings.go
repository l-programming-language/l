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
	rec, err := c.builtinRecord(name)
	if err != nil {
		return DefinitionInstance{}, err
	}
	return c.instantiateDefinition(rec, arguments, descriptors...)
}

// Selected definitions share named expansion, including Function members and
// private dictionary operations with source-owned construction descriptors.
func (c *Compilation) instantiateDefinition(rec *record, arguments map[string]DefinitionArgument, descriptors ...string) (DefinitionInstance, error) {
	if len(c.definitionTemporaries) == 0 {
		defer c.definitionScope()()
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
	return bound, validateDefinitionArguments(rec, arguments, used)
}

func validateDefinitionArguments(rec *record, arguments map[string]DefinitionArgument, used map[string]bool) error {
	for name := range arguments {
		if !used[name] {
			return lerrors.InvalidInput.Errorf("definition %s received unused argument %s", rec.ID, name)
		}
	}
	return nil
}

func (c *Compilation) instantiateNamedDefinition(rec *record, arguments map[string]DefinitionArgument, used map[string]bool, active map[*record]bool, values ...DefinitionArgument) (DefinitionInstance, error) {
	return c.instantiateSourceCall(rec, values, arguments, used, active, nil)
}

func (c *Compilation) instantiateNamedObject(object *objectGram, arguments map[string]DefinitionArgument, used map[string]bool, active map[*record]bool, values ...DefinitionArgument) (DefinitionInstance, error) {
	node, err := c.constructObject(object, make(map[*descriptorRule]bool))
	if err != nil {
		return DefinitionInstance{}, err
	}
	if len(node.Arguments) > 0 {
		if len(values) > 0 {
			return DefinitionInstance{}, lerrors.InvalidInput.Errorf("%s received %d unused arguments", node.Nominal, len(values))
		}
		for _, parameter := range node.Arguments {
			value, _, err := c.namedObjectArgument(parameter, arguments, used, active)
			if err != nil {
				return DefinitionInstance{}, err
			}
			values = append(values, value)
		}
	}
	return c.instantiateNamedConstruction(node, arguments, used, active, values...)
}

func (c *Compilation) namedObjectArgument(object *objectGram, arguments map[string]DefinitionArgument, used map[string]bool, active map[*record]bool) (DefinitionArgument, bool, error) {
	return c.definitionOperand(object, active, nil, func(child *objectGram) (DefinitionInstance, bool, error) {
		instance, err := c.instantiateNamedObject(child, arguments, used, active)
		return instance, false, err
	})
}

func (c *Compilation) instantiateNamedConstruction(node *constructionNode, local map[string]DefinitionArgument, localUsed map[string]bool, active map[*record]bool, values ...DefinitionArgument) (DefinitionInstance, error) {
	wrappers, selected, err := constructionTemplates(node)
	if err != nil {
		return DefinitionInstance{}, err
	}
	var instance DefinitionInstance
	if selected < 0 {
		instance, err = c.instantiateNamedDefinition(node.Definition, local, localUsed, active, values...)
		selected = len(wrappers)
	} else {
		wrapper := wrappers[selected]
		bindings, bindErr := bindDefinitionArguments(wrapper, values, local, localUsed)
		if bindErr != nil {
			return DefinitionInstance{}, bindErr
		}
		instance, err = c.instantiateTemplate(wrapper, bindings)
	}
	if err != nil {
		return DefinitionInstance{}, err
	}
	for i := selected - 1; i >= 0; i-- {
		wrapper := wrappers[i]
		bindings := map[string]DefinitionArgument{wrapper.Slots[0].Name: {Kind: definitionArgDefinition.String(), Definition: &instance}}
		instance, err = c.instantiateTemplate(wrapper, bindings)
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

// Callable structure follows authored constructors. This borrows the selected
// leaf's signature; executable expansion still traverses the original words.
func (c *Compilation) callableTemplate(rec *record, active map[*record]bool) (*DefinitionTemplate, error) {
	rec, err := canonicalRecord(rec, c.Semantics.Aliases)
	if err != nil {
		return nil, err
	}
	if active[rec] {
		return nil, lerrors.Cycle.Errorf("recursive constructor %s", rec.ID)
	}
	active[rec] = true
	defer delete(active, rec)
	template, object, err := c.recordDefinition(rec)
	if err != nil || template != nil || object == nil {
		return template, err
	}
	node, err := c.constructObject(object, make(map[*descriptorRule]bool))
	if err != nil {
		return nil, err
	}
	template, err = c.callableTemplate(node.Definition, active)
	if err != nil || template == nil || !template.IsFunction() {
		return nil, err
	}
	wrappers, selected, err := constructionTemplates(node)
	if err != nil {
		return nil, err
	}
	if selected >= 0 {
		template = wrappers[selected]
	}
	bound := *template
	bound.Slots = nil
	supplied := 0
	for _, slot := range template.Slots {
		if slot.Type == "Type" && supplied < len(node.Arguments) {
			supplied++
			continue
		}
		bound.Slots = append(bound.Slots, slot)
	}
	return &bound, nil
}

// Value-taking metadata receives a named result, never a nested executable
// composition. The selected Binding definition owns the operation and storage.
func (c *Compilation) atomArgument(argument DefinitionArgument) (DefinitionArgument, error) {
	if argument.atoms != nil {
		atoms := append([]DefinitionArgument(nil), (*argument.atoms)...)
		for index, value := range atoms {
			var err error
			atoms[index], err = c.atomArgument(value)
			if err != nil {
				return DefinitionArgument{}, err
			}
		}
		return definitionAtomListArgument(atoms...), nil
	}
	if argument.Definition == nil || argument.Definition.result != nil {
		return argument, nil
	}
	if c.definitionValues[argument.Definition] {
		return DefinitionArgument{}, lerrors.Cycle.Errorf("value binding must consume its operand composition")
	}
	if c.definitionValues == nil {
		c.definitionValues = make(map[*DefinitionInstance]bool)
	}
	c.definitionValues[argument.Definition] = true
	defer delete(c.definitionValues, argument.Definition)
	name := c.definitionLabel()
	bound, err := c.instantiateDescribedBuiltin("Storage.Binding", []string{"Storage.Produced"}, map[string]DefinitionArgument{"Name": definitionReferenceArgument(name), "Value": argument})
	if err != nil {
		return DefinitionArgument{}, err
	}
	result := definitionReferenceArgument(name)
	result.source = argument.source
	bound.result = &result
	argument.Definition = &bound
	return argument, nil
}

func (c *Compilation) instantiateTemplate(template *DefinitionTemplate, arguments map[string]DefinitionArgument) (DefinitionInstance, error) {
	if len(c.definitionTemporaries) == 0 {
		defer c.definitionScope()()
	}
	bindings := make(map[string]DefinitionArgument, len(arguments))
	for name, argument := range arguments {
		bindings[name] = argument
	}
	for _, slot := range template.Slots {
		if template.metadataSlots[slot.Name] && (slot.Type == "" || slot.Type == "Atoms") {
			argument, err := c.atomArgument(bindings[slot.Name])
			if err != nil {
				return DefinitionInstance{}, err
			}
			bindings[slot.Name] = argument
		}
	}
	return template.Instantiate(bindings)
}
