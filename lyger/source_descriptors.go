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
	for _, argument := range constructorOperands(object) {
		base, err := c.descriptorBaseArgument(rule, argument)
		if err != nil {
			return nil, err
		}
		if base {
			count++
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

// One selected header supplies grammar boundaries, result naming and the ordered
// value roles used by wrappers. Structural slots never become values by position.
type definitionParameter struct {
	DefinitionSlot
	Value   bool
	Forward bool // exact transparent slot; its consuming signature supplies the role
	token   *token
	source  *lStatement
}

type definitionHeader struct {
	template     *DefinitionTemplate
	parameters   []definitionParameter
	implicitName *DefinitionSlot
	fixed        bool
}

func templateHeader(template *DefinitionTemplate, constructor, named bool) definitionHeader {
	header := definitionHeader{template: template, fixed: true}
	for index, slot := range template.Slots {
		if template.scopedLabels[slot.Name] {
			continue
		}
		if constructor && index == 0 && (slot.Type == "Label" || named || template.hasDeclarationName()) {
			header.implicitName = &template.Slots[index]
			continue
		}
		if constructor && template.IsFunction() && template.signatureSlots[slot.Name] != "" {
			continue
		}
		header.fixed = header.fixed && slot.Type != "Syllables" && slot.Type != "Atoms"
		header.parameters = append(header.parameters, definitionParameter{DefinitionSlot: slot, Forward: template.identitySlot() != nil, Value: template.metadataSlots[slot.Name] && (slot.Type == "" || slot.Type == "Atoms")})
	}
	return header
}

func recordHeader(rec *record, constructor, named bool) (definitionHeader, error) {
	if rec == nil {
		return definitionHeader{}, nil
	}
	if function := rec.function; function != nil && function.Native {
		header := definitionHeader{fixed: true}
		for _, input := range function.Constructor.Inputs {
			header.parameters = append(header.parameters, definitionParameter{DefinitionSlot: DefinitionSlot{Name: input.Key}, Value: true})
		}
		return header, nil
	}
	if word := rec.bindingWord(); word != nil {
		if word.Value != "" {
			template, err := ParseDefinitionTemplate([]byte(word.Value), SourceSpan{})
			if err != nil {
				return definitionHeader{}, err
			}
			return templateHeader(template, constructor, named), nil
		}
		header := definitionHeader{fixed: len(word.Inputs) > 0}
		for _, name := range word.Inputs {
			header.parameters = append(header.parameters, definitionParameter{DefinitionSlot: DefinitionSlot{Name: name}})
		}
		return header, nil
	}
	return definitionHeader{}, nil
}

func (header definitionHeader) arity() grammarArity {
	return grammarArity{Count: len(header.parameters), Fixed: header.fixed}
}

func constructorOperands(object *objectGram) (arguments []*objectGram) {
	for phrase := object.phrase; phrase != nil; phrase = phrase.phrase {
		if phrase.gType != "phrase.signature" && phrase.gType != "phrase.directive" {
			arguments = append(arguments, phrase.orderedObjects()...)
		}
	}
	if object.clause != nil {
		arguments = append(arguments, object.clause.orderedObjects()...)
	}
	return arguments
}

func (c *Compilation) constructionHeader(node *constructionNode, named bool, active map[*record]bool) (definitionHeader, []*DefinitionTemplate, int, error) {
	wrappers, selected, err := c.constructionTemplates(node)
	if err != nil {
		return definitionHeader{}, nil, -1, err
	}
	if selected >= 0 {
		return templateHeader(wrappers[selected], true, named), wrappers, selected, nil
	}
	header, err := c.sourceHeader(node.Definition, named, active)
	return header, wrappers, len(wrappers), err
}

func (site registrationSite) actionParameters(source *lStatement) ([]definitionParameter, error) {
	owned := make(map[*token]bool)
	for _, action := range site.actionClauses() {
		for _, tok := range action.tokens {
			owned[tok] = true
		}
	}
	// A declaration nested in this body owns its own generic arguments.
	sites, err := source.sGram.registrationSites()
	if err != nil {
		return nil, err
	}
	for _, nested := range sites {
		if nested.subject != nil && owned[nested.subject.noun] {
			for _, action := range nested.actionClauses() {
				for _, tok := range action.tokens {
					delete(owned, tok)
				}
			}
		}
	}
	var parameters []definitionParameter
	for index, tok := range source.tokens {
		if !owned[tok] || !tok.hasRole("article", "variable") {
			continue
		}
		position, ok := actionArgument(source.tokens, index)
		if ok && owned[source.tokens[position]] {
			formal := source.tokens[position]
			name, err := nominalType(formal)
			if err != nil {
				return nil, err
			}
			parameters = append(parameters, definitionParameter{DefinitionSlot: DefinitionSlot{Name: name}, token: formal, source: source})
		}
	}
	return parameters, nil
}

func (c *Compilation) sourceParameters(rec *record) ([]definitionParameter, error) {
	source, site, err := c.Semantics.definitionSite(rec)
	if err != nil || site == nil {
		return nil, err
	}
	parameters, err := site.actionParameters(source)
	if err != nil {
		return nil, err
	}
	for index, formal := range parameters {
		typ, err := c.resolver().reference(formal.token)
		if err != nil {
			return nil, err
		}
		process, err := c.builtinRecord("Process")
		if err != nil {
			return nil, err
		}
		parameters[index].Value = typ != process
	}
	return parameters, nil
}

// Follow only routes that receive inherited operands. Authored arguments and
// directive bodies keep their own scope; this pass never constructs or emits.
func (c *Compilation) sourceHeader(rec *record, named bool, active map[*record]bool) (definitionHeader, error) {
	rec, err := canonicalRecord(rec, c.Semantics.Aliases)
	if err != nil {
		return definitionHeader{}, err
	}
	if rec == nil || rec.function != nil {
		return recordHeader(rec, true, named)
	}
	if active[rec] {
		return definitionHeader{}, lerrors.Cycle.Errorf("recursive constructor signature %s", rec.ID)
	}
	if active != nil {
		active[rec] = true
		defer delete(active, rec)
	}
	shape, err := c.Semantics.selectedDefinition(rec)
	if err != nil {
		return definitionHeader{}, err
	}
	if shape.site == nil {
		return definitionHeader{}, lerrors.UnresolvedReference.Errorf("constructor %s is waiting for its declaration", rec.ID)
	}
	if shape.template != nil {
		return templateHeader(shape.template, true, named), nil
	}
	header := definitionHeader{fixed: true}
	if active == nil {
		return header, nil
	}
	header.parameters, err = c.sourceParameters(rec)
	if err != nil || len(header.parameters) > 0 {
		return header, err
	}
	// Invocation contracts accept the selected target's argument list. Their
	// source implementation is checked against that named contract on expansion.
	for declared := range c.Semantics.Declarations {
		if function := declared.function; function != nil {
			invoke, err := canonicalRecord(function.Invoke, c.Semantics.Aliases)
			if err != nil {
				return header, err
			}
			if invoke == rec {
				return definitionHeader{}, nil
			}
		}
	}
	var objects []*objectGram
	if object := shape.object; object != nil {
		objects = append(objects, object)
	}
	for _, action := range shape.site.actionClauses() {
		objects = append(objects, action.orderedObjects()...)
	}
	inherited := false
	for index := 0; index < len(objects); index++ {
		object := objects[index]
		if object == nil || object.noun == nil || object.noun.binding == nil {
			continue
		}
		node, err := c.constructObjectMode(object, make(map[*descriptorRule]bool), constructionHeaderOnly)
		if err != nil {
			return header, err
		}
		operands := node.Arguments
		if directive := object.directivePhrase(); directive != nil {
			for _, action := range directive.clauses {
				operands = append(operands, action.orderedObjects()...)
			}
		}
		if len(operands) > 0 {
			objects = append(objects, operands...)
			continue
		}
		if rec := node.Definition; rec == nil || rec.definitionArgument != nil || rec.bindingWord() == nil {
			continue
		}
		child, _, _, err := c.constructionHeader(node, object.name != nil, active)
		if err != nil {
			return header, err
		}
		if child.fixed && len(child.parameters) == 0 {
			continue
		}
		if inherited && header.fixed && child.fixed && len(header.parameters) != len(child.parameters) {
			return header, lerrors.InvalidInput.Errorf("constructor %s inherits incompatible operand counts", rec.ID)
		}
		if !inherited || !header.fixed && child.fixed {
			header.parameters, child.parameters = child.parameters, header.parameters
			header.fixed = child.fixed
		}
		for index := range header.parameters {
			other := index
			if other >= len(child.parameters) && len(child.parameters) > 0 && child.parameters[len(child.parameters)-1].Type == "Atoms" {
				other = len(child.parameters) - 1
			}
			if other < len(child.parameters) {
				header.parameters[index].Value = header.parameters[index].Value || child.parameters[other].Value
				header.parameters[index].Forward = header.parameters[index].Forward && child.parameters[other].Forward
			}
		}
		inherited = true
	}
	return header, nil
}

// An exact forwarding template acquires a value role only from a checked
// consuming parameter. Unrelated structural/body slots remain structural.
func (c *Compilation) wrapperValueRoles(node *constructionNode, header *definitionHeader, selected int) error {
	forwarding := false
	for _, parameter := range header.parameters {
		forwarding = forwarding || parameter.Forward
	}
	if !forwarding {
		return nil
	}
	for _, rule := range node.Wrappers[:selected] {
		object, err := c.sourceDescriptorObject(rule)
		if err != nil {
			return err
		}
		if object == nil {
			continue
		}
		target, err := c.constructObjectMode(object, make(map[*descriptorRule]bool), constructionHeaderOnly)
		if err != nil {
			return err
		}
		receiver, _, _, err := c.constructionHeader(target, object.name != nil, make(map[*record]bool))
		if err != nil {
			return err
		}
		forward := 0
		for index, operand := range target.Arguments {
			base, err := c.descriptorBaseArgument(rule, operand)
			if err != nil {
				return err
			}
			if base || operand.article == nil || !operand.article.hasRole("article", "variable") {
				continue
			}
			if forward < len(header.parameters) && header.parameters[forward].Forward && index < len(receiver.parameters) {
				header.parameters[forward].Value = header.parameters[forward].Value || receiver.parameters[index].Value
			}
			forward++
		}
	}
	return nil
}

// Materialize original value operands once, in signature order. The result and
// every wrapper share these checked references and the outer prerequisites.
func (c *Compilation) bindWrapperOperands(header definitionHeader, values []DefinitionArgument, local map[string]DefinitionArgument) ([]DefinitionArgument, map[string]DefinitionArgument, []DefinitionInstance, error) {
	bound := make(map[string]DefinitionArgument, len(local))
	for key, argument := range local {
		bound[key] = argument
	}
	values = append([]DefinitionArgument(nil), values...)
	named := len(values) == 0
	if named {
		values = make([]DefinitionArgument, len(header.parameters))
		for index, parameter := range header.parameters {
			argument, _, err := namedDefinitionArgument(parameter.Name, bound, nil)
			if err != nil {
				return nil, nil, nil, err
			}
			values[index] = argument
		}
	}
	var before []DefinitionInstance
	for index, parameter := range header.parameters {
		if index >= len(values) {
			break
		}
		if parameter.Value {
			argument, err := c.atomArgument(values[index])
			if err != nil {
				return nil, nil, nil, err
			}
			var steps []DefinitionInstance
			values[index], steps = argument.boundResult()
			before = append(before, steps...)
		}
		if named {
			for key := range bound {
				if strings.EqualFold(key, parameter.Name) {
					bound[key] = values[index]
				}
			}
		}
	}
	return values, bound, before, nil
}
