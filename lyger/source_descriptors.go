package Lyger

import (
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// An explicit variable reference to the canonical target distinguishes a source
// wrapper from a constructor replacement. The surrounding source constructor
// retains ordinary lexical selection, argument checking and dependencies.
func (c *Compilation) sourceDescriptorObject(rule *descriptorRule) (*objectGram, error) {
	if rule.Template != nil || rule.Object == nil || rule.Object.phrase == nil || rule.Object.phrase.gType != "phrase.declarative" {
		return nil, nil
	}
	object := rule.Object.phrase.obj
	if object == nil {
		return nil, nil
	}
	count := 0
	for phrase := object.phrase; phrase != nil; phrase = phrase.phrase {
		for _, argument := range phrase.orderedObjects() {
			base, err := c.descriptorBaseArgument(rule, argument)
			if err != nil {
				return nil, err
			}
			if base {
				count++
			}
		}
	}
	if count > 1 {
		return nil, lerrors.InvalidInput.Errorf("descriptor for %s has multiple base parameters", rule.Target)
	}
	if count == 1 {
		return object, nil
	}
	return nil, nil
}

func (c *Compilation) descriptorBaseArgument(rule *descriptorRule, object *objectGram) (bool, error) {
	if object == nil || object.article == nil || !object.article.hasRole("article", "variable") {
		return false, nil
	}
	identity, err := c.resolver().nominalIdentity(object.noun)
	return identity == rule.Target, err
}

func (c *Compilation) instantiateSourceDescriptor(rule *descriptorRule, base DefinitionInstance, original []DefinitionArgument, name string, active, dynamic map[*record]bool) (DefinitionInstance, error) {
	object, err := c.sourceDescriptorObject(rule)
	if err != nil {
		return DefinitionInstance{}, err
	}
	if object == nil {
		return DefinitionInstance{}, lerrors.Internal.Errorf("source wrapper has no explicit base parameter")
	}
	node, err := c.constructObject(object, make(map[*descriptorRule]bool))
	if err != nil {
		return DefinitionInstance{}, err
	}
	var arguments []DefinitionArgument
	forward := 0
	for _, parameter := range node.Arguments {
		isBase, err := c.descriptorBaseArgument(rule, parameter)
		if err != nil {
			return DefinitionInstance{}, err
		}
		switch {
		case isBase:
			arguments = append(arguments, DefinitionArgument{Kind: definitionArgDefinition.String(), Definition: &base})
		case parameter.article != nil && parameter.article.hasRole("article", "variable"):
			if forward == len(original) {
				return DefinitionInstance{}, lerrors.InvalidInput.Errorf("descriptor for %s forwards more arguments than its target accepts", rule.Target)
			}
			arguments = append(arguments, original[forward])
			forward++
		default:
			argument, _, err := c.staticObjectArgument(parameter, nil, active, dynamic)
			if err != nil {
				return DefinitionInstance{}, err
			}
			arguments = append(arguments, argument)
		}
	}
	node.Arguments, node.Parameters, node.Name = nil, nil, name
	instance, _, err := c.instantiateStaticConstruction(node, arguments, active, dynamic)
	return instance, err
}

func (c *Compilation) namedDescriptorArguments(record *record, template *DefinitionTemplate, local map[string]DefinitionArgument) ([]DefinitionArgument, error) {
	parameters, err := c.definitionParameters(record, template)
	if err != nil {
		return nil, err
	}
	arguments := make([]DefinitionArgument, len(parameters))
	for index, parameter := range parameters {
		name := parameter.Name
		found := false
		for key, argument := range local {
			if !strings.EqualFold(key, name) {
				continue
			}
			if found {
				return nil, lerrors.AmbiguousReference.Errorf("descriptor parameter %s matches multiple input bindings", name)
			}
			arguments[index], found = argument, true
		}
		if !found {
			return nil, lerrors.InvalidInput.Errorf("descriptor has no original argument %s", name)
		}
	}
	return arguments, nil
}

// A selected signature determines which operands are values. Structured bodies,
// labels and type atoms retain their existing roles and are never evaluated as
// values merely because a descriptor also refers to the base construction.
type definitionParameter struct {
	Name  string
	Value bool
}

func (c *Compilation) definitionParameters(rec *record, template *DefinitionTemplate) ([]definitionParameter, error) {
	var parameters []definitionParameter
	if template == nil {
		if function := rec.function; function != nil && function.Native {
			for _, input := range function.Constructor.Inputs {
				parameters = append(parameters, definitionParameter{Name: input.Key, Value: true})
			}
			return parameters, nil
		}
		var err error
		template, err = c.recordTemplate(rec)
		if err != nil {
			return nil, err
		}
	}
	if template != nil {
		for index, slot := range template.Slots {
			if template.scopedLabels[slot.Name] || index == 0 && slot.Type == "Label" {
				continue
			}
			parameters = append(parameters, definitionParameter{Name: slot.Name, Value: template.metadataSlots[slot.Name] && (slot.Type == "" || slot.Type == "Atoms")})
		}
		return parameters, nil
	}
	process, err := c.builtinRecord("Process")
	if err != nil {
		return nil, err
	}
	for _, source := range c.Scope.Ordered {
		if source.Statement != c.Semantics.Declarations[rec] {
			continue
		}
		for _, formal := range source.Arguments {
			typ, err := c.resolver().reference(formal.Token)
			if err != nil {
				return nil, err
			}
			parameters = append(parameters, definitionParameter{Value: typ != process})
		}
		return parameters, nil
	}
	return nil, lerrors.InvalidInput.Errorf("source descriptor requires an ordered target signature")
}

// Materialize original value operands once, in signature order, before either
// the base or a wrapper uses them. Positional and named construction share this
// operation; each wrapper receives references to the very same results.
func (c *Compilation) bindWrapperOperands(rec *record, template *DefinitionTemplate, values []DefinitionArgument, local map[string]DefinitionArgument) ([]DefinitionArgument, map[string]DefinitionArgument, []DefinitionInstance, error) {
	parameters, err := c.definitionParameters(rec, template)
	if err != nil {
		return nil, nil, nil, err
	}
	values = append([]DefinitionArgument(nil), values...)
	bound := make(map[string]DefinitionArgument, len(local))
	for key, argument := range local {
		bound[key] = argument
	}
	var before []DefinitionInstance
	for index, parameter := range parameters {
		if !parameter.Value {
			continue
		}
		var argument DefinitionArgument
		key := ""
		if len(values) > 0 {
			if index >= len(values) {
				continue // The ordinary signature binder reports missing operands.
			}
			argument = values[index]
		} else {
			for candidate := range bound {
				if strings.EqualFold(candidate, parameter.Name) {
					if key != "" {
						return nil, nil, nil, lerrors.AmbiguousReference.Errorf("descriptor parameter %s matches multiple input bindings", parameter.Name)
					}
					key = candidate
				}
			}
			if key == "" {
				continue
			}
			argument = bound[key]
		}
		argument, err = c.atomArgument(argument)
		if err != nil {
			return nil, nil, nil, err
		}
		argument, steps := argument.boundResult()
		before = append(before, steps...)
		if len(values) > 0 {
			values[index] = argument
		} else {
			bound[key] = argument
		}
	}
	return values, bound, before, nil
}
