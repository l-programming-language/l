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
func (c *Compilation) instantiateDefinition(rec *record, arguments map[string]DefinitionArgument, descriptors ...string) (result DefinitionInstance, resultErr error) {
	if c.contractRecorder != nil {
		defer func() {
			if resultErr == nil {
				c.contractRecorder.recordLowering(rec, result, arguments)
			}
		}()
	}
	if len(c.definitionTemporaries) == 0 {
		defer c.definitionScope()()
	}
	node, err := c.definitionConstruction(rec, descriptors...)
	if err != nil {
		return DefinitionInstance{}, err
	}
	used := make(map[string]bool)
	bound, err := c.instantiateNamedConstruction(node, arguments, used, make(map[*record]bool), nil)
	if err != nil {
		return DefinitionInstance{}, err
	}
	return bound, validateDefinitionArguments(rec, arguments, used)
}

// Result naming and construction use the same selected descriptor chain.
func (c *Compilation) definitionConstruction(rec *record, descriptors ...string) (*constructionNode, error) {
	node := &constructionNode{Definition: rec, Nominal: rec.ID}
	for _, name := range descriptors {
		modifier, err := c.builtinRecord(name)
		if err != nil {
			return nil, err
		}
		descriptor := c.Semantics.Descriptors[modifier]
		if descriptor == nil {
			return nil, lerrors.TypeMismatch.Errorf("%s is not a descriptor", name)
		}
		matched, err := c.applyDescriptor(node, descriptor, nominalIdentity{definition: rec}, make(map[*descriptorRule]bool), constructionConcrete, nil)
		if err != nil {
			return nil, err
		}
		if !matched {
			return nil, lerrors.TypeMismatch.Errorf("%s does not describe %s", name, rec.ID)
		}
	}
	return node, nil
}

func validateDefinitionArguments(rec *record, arguments map[string]DefinitionArgument, used map[string]bool) error {
	for name := range arguments {
		if !used[name] {
			return lerrors.InvalidInput.Errorf("definition %s received unused argument %s", rec.ID, name)
		}
	}
	return nil
}

func (c *Compilation) instantiateNamedObject(object *objectGram, arguments map[string]DefinitionArgument, used map[string]bool, active, dynamic map[*record]bool, values ...DefinitionArgument) (DefinitionInstance, error) {
	node, err := c.constructObject(object, make(map[*descriptorRule]bool))
	if err != nil {
		return DefinitionInstance{}, err
	}
	if len(node.Arguments) > 0 {
		if len(values) > 0 {
			return DefinitionInstance{}, lerrors.InvalidInput.Errorf("%s received %d unused arguments", node.Nominal, len(values))
		}
		for _, parameter := range node.Arguments {
			value, _, err := c.namedObjectArgument(parameter, arguments, used, active, dynamic)
			if err != nil {
				return DefinitionInstance{}, err
			}
			values = append(values, value)
		}
	}
	return c.instantiateNamedConstruction(node, arguments, used, active, dynamic, values...)
}

func (c *Compilation) namedObjectArgument(object *objectGram, arguments map[string]DefinitionArgument, used map[string]bool, active, dynamic map[*record]bool) (DefinitionArgument, bool, error) {
	return c.definitionOperand(object, active, dynamic, func(child *objectGram) (DefinitionInstance, bool, error) {
		instance, err := c.instantiateNamedObject(child, arguments, used, active, dynamic)
		return instance, false, err
	})
}

func (c *Compilation) instantiateNamedConstruction(node *constructionNode, local map[string]DefinitionArgument, localUsed map[string]bool, active, dynamic map[*record]bool, values ...DefinitionArgument) (DefinitionInstance, error) {
	if err := node.validateRole(dynamic); err != nil {
		return DefinitionInstance{}, err
	}
	var signatures map[*record]bool
	if len(node.Wrappers) > 0 || len(values) > 0 || node.Name != "" {
		signatures = make(map[*record]bool)
	}
	header, wrappers, selected, err := c.constructionHeader(node, node.Source != nil && node.Source.name != nil, signatures)
	if err != nil {
		return DefinitionInstance{}, err
	}
	return c.instantiateConstruction(node, header, wrappers, selected, values, local, localUsed, active, dynamic)
}

func (c *Compilation) instantiateConstruction(node *constructionNode, header definitionHeader, wrappers []*DefinitionTemplate, selected int, values []DefinitionArgument, local map[string]DefinitionArgument, used map[string]bool, active, dynamic map[*record]bool) (DefinitionInstance, error) {
	var before []DefinitionInstance
	var err error
	positional := len(values) > 0
	if len(wrappers) > 0 {
		if err := c.wrapperValueRoles(node, &header, selected); err != nil {
			return DefinitionInstance{}, err
		}
		values, local, before, err = c.bindWrapperOperands(header, values, local)
		if err != nil {
			return DefinitionInstance{}, err
		}
	}
	arguments := values
	if !positional && used != nil {
		arguments = nil
	}
	var instance DefinitionInstance
	if selected == len(wrappers) {
		instance, err = c.instantiateSourceCall(node.Definition, arguments, local, used, active, dynamic)
	} else if wrappers[selected] == nil {
		instance, err = c.instantiateSourceDescriptor(node.Wrappers[selected], DefinitionInstance{}, arguments, node.Name, local, used, active, dynamic)
	} else {
		bindings, bindErr := bindDefinitionArguments(header.template, arguments, local, used)
		if bindErr != nil {
			return DefinitionInstance{}, bindErr
		}
		instance, err = c.instantiateTemplate(header.template, bindings)
	}
	if err != nil {
		return DefinitionInstance{}, err
	}
	if node.Name != "" && (selected == len(wrappers) || wrappers[selected] != nil) && (header.implicitName == nil || header.implicitName.Type == "Label" || header.template.hasDeclarationName()) {
		result := definitionReferenceArgument(node.Name)
		if node.Source != nil {
			result.source = c.operandSource(node.Source.noun)
		}
		if header.implicitName == nil {
			instance, err = bindDefinitionResult(result, DefinitionArgument{Kind: definitionArgDefinition.String(), Definition: &instance})
			if err != nil {
				return DefinitionInstance{}, err
			}
		}
		instance.result = &result
		if local == nil {
			local = make(map[string]DefinitionArgument)
		}
		local["Name"] = result
	}
	if function := node.Definition.function; function != nil && function.Native && len(values) > 0 {
		if local == nil {
			local = make(map[string]DefinitionArgument)
		}
		for index, input := range function.Constructor.Inputs {
			local[input.Key] = values[index]
		}
	}
	for index := selected - 1; index >= 0; index-- {
		if wrappers[index] == nil {
			instance, err = c.instantiateSourceDescriptor(node.Wrappers[index], instance, values, node.Name, local, used, active, dynamic)
		} else {
			instance, err = c.instantiateDescriptorWrapper(node.Wrappers[index], wrappers[index], instance, local, used)
		}
		if err != nil {
			return DefinitionInstance{}, err
		}
	}
	instance.before = append(before, instance.before...)
	instance.Symbol, instance.definition = node.Definition.ID, node.Definition
	return instance, nil
}

func (c *Compilation) constructionTemplates(node *constructionNode) ([]*DefinitionTemplate, int, error) {
	var err error
	wrappers := make([]*DefinitionTemplate, len(node.Wrappers))
	selected := -1
	for i, rule := range node.Wrappers {
		wraps := false
		if rule.Template == nil {
			_, wraps, err = c.sourceDescriptorObject(rule)
		} else {
			wrappers[i], err = ParseDefinitionTemplate(rule.Template.rawBytes, rule.Source.source.Span)
			if err == nil {
				var base string
				base, err = c.descriptorBaseSlot(rule, wrappers[i])
				wraps = base != ""
			}
		}
		if err != nil {
			return nil, -1, err
		}
		if !wraps {
			if selected >= 0 {
				return nil, -1, lerrors.InvalidInput.Errorf("constructor %s has competing descriptor definitions", node.Nominal)
			}
			selected = i
		}
	}
	return wrappers, selected, nil
}

// A wrapper refers to its base by a word from the target's canonical family.
// The rule's lexical environment owns aliases and inflections; a same-spelled
// foreign declaration is not the original target.
func (c *Compilation) descriptorBaseSlot(rule *descriptorRule, wrapper *DefinitionTemplate) (string, error) {
	base := ""
	for _, slot := range wrapper.Slots {
		if slot.Type != "" {
			continue
		}
		matches := false
		if rule.Target.definition == nil {
			matches = strings.EqualFold(slot.Name, rule.Target.String())
		} else if rule.Object != nil && rule.Object.noun != nil {
			query := &token{rawBytes: []byte(slot.Name), scope: rule.Object.noun.scope}
			binding, _, err := (semanticResolver{scopes: c.Scope, semantics: c.Semantics}).lookup(query)
			if err != nil {
				return "", err
			}
			if binding != nil {
				record, err := canonicalRecord(binding.Record, c.Semantics.Aliases)
				if err != nil {
					return "", err
				}
				matches = record == rule.Target.definition
			}
		}
		if matches {
			if base != "" {
				return "", lerrors.InvalidInput.Errorf("descriptor for %s has multiple base parameters", rule.Target)
			}
			base = slot.Name
		}
	}
	return base, nil
}

func namedDefinitionArgument(name string, arguments map[string]DefinitionArgument, used map[string]bool) (DefinitionArgument, bool, error) {
	key := ""
	for candidate := range arguments {
		if strings.EqualFold(candidate, name) {
			if key != "" {
				return DefinitionArgument{}, false, lerrors.AmbiguousReference.Errorf("parameter %s matches multiple input bindings", name)
			}
			key = candidate
		}
	}
	if key != "" && used != nil {
		used[key] = true
	}
	return arguments[key], key != "", nil
}

func (c *Compilation) instantiateDescriptorWrapper(rule *descriptorRule, wrapper *DefinitionTemplate, base DefinitionInstance, local map[string]DefinitionArgument, used map[string]bool) (DefinitionInstance, error) {
	baseSlot, err := c.descriptorBaseSlot(rule, wrapper)
	if err != nil {
		return DefinitionInstance{}, err
	}
	bindings := make(map[string]DefinitionArgument, len(wrapper.Slots))
	for _, slot := range wrapper.Slots {
		if slot.Name == baseSlot {
			bindings[slot.Name] = DefinitionArgument{Kind: definitionArgDefinition.String(), Definition: &base}
			continue
		}
		argument, found, err := namedDefinitionArgument(slot.Name, local, used)
		if err != nil {
			return DefinitionInstance{}, err
		}
		if found {
			bindings[slot.Name] = argument
		}
	}
	instance, err := c.instantiateTemplate(wrapper, bindings)
	if err == nil && instance.result == nil {
		instance.result = base.result
	}
	return instance, err
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
	wrappers, selected, err := c.constructionTemplates(node)
	if err != nil {
		return nil, err
	}
	if selected >= 0 {
		template = wrappers[selected]
	}
	if template == nil {
		return nil, nil
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
// composition. Compiler AML binding preserves its once-only evaluation.
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
	bound, err := bindDefinitionResult(definitionReferenceArgument(name), argument)
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
		if template.metadataSlots[slot.Name] && (slot.Type == "" || slot.Type == "Atoms" || slot.Type == "Callable") {
			argument, err := c.atomArgument(bindings[slot.Name])
			if err != nil {
				return DefinitionInstance{}, err
			}
			bindings[slot.Name] = argument
		}
	}
	instance, err := template.Instantiate(bindings)
	if err == nil && instance.result == nil && template.hasDeclarationName() && len(template.Slots) > 0 {
		// A selected declaration's first Key slot names its result even when
		// forwarded through a named source action. Keep the bound identity;
		// never recover a name from rendered AML.
		if name := bindings[template.Slots[0].Name]; name.reference != "" {
			instance.result = &name
		}
	}
	return instance, err
}
