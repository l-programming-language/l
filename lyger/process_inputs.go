package Lyger

import (
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// Retain signature/body labels before grammar classification without reserving
// them in the enclosing scope. Publication assigns their actual lexical owner.
func reserveSignatureDeclarations(stmt *lStatement, declarations []scopeDeclaration) []scopeDeclaration {
	seen := make(map[*token]bool)
	for _, declaration := range declarations {
		seen[declaration.Token] = true
	}
	for index, tok := range stmt.tokens {
		if tok.isDeclare() && !seen[tok] {
			declarations = append(declarations, scopeDeclaration{SubjectID: tok.declarationRoot(), Token: tok, TokenIndex: index, provisional: true})
			seen[tok] = true
		}
	}
	return declarations
}

// A Process is a source-defined W sequence, not a callable or a native closure.
// Aliases resolve before inspecting this existing structural definition.
func (c *Compilation) processTemplate(object *objectGram) (bool, error) {
	rec, err := c.resolver().reference(object.noun)
	if err != nil || rec == nil {
		return false, err
	}
	template, err := c.recordTemplate(rec)
	return ownsProcessDefinitionBody(template), err
}

// A W sequence owns its supplied body even when the selected definition adds
// other operations. That ownership permits lexical input bindings but proves
// nothing about whether the body runs or returns from its caller.
func ownsProcessDefinitionBody(template *DefinitionTemplate) bool {
	return template != nil && len(template.Slots) == 1 && template.Slots[0].Type == "Syllables" &&
		len(template.parts) >= 2 && strings.HasPrefix(strings.TrimSpace(template.parts[0].text), "W^")
}

// Only a transparent W sequence carries an unconditional exit from its body.
func isProcessDefinitionTemplate(template *DefinitionTemplate) bool {
	return template != nil && len(template.Slots) == 1 && template.Slots[0].Type == "Syllables" &&
		len(template.parts) == 3 && strings.TrimSpace(template.parts[0].text) == "W^" &&
		template.parts[1].slot != nil && strings.TrimSpace(template.parts[2].text) == ""
}

func (c *Compilation) prepareProcessInputs(stmt *lStatement, prepared *statementRegistration) error {
	seen := make(map[*objectGram]bool)
	var object func(*objectGram) error
	phrase := func(p *phraseGram) error {
		for ; p != nil; p = p.phrase {
			for _, child := range p.orderedObjects() {
				if err := object(child); err != nil {
					return err
				}
			}
		}
		return nil
	}
	object = func(o *objectGram) error {
		if o == nil || seen[o] {
			return nil
		}
		seen[o] = true
		if o.directivePhrase() != nil {
			process, err := c.processTemplate(o)
			if err != nil {
				return err
			}
			if process {
				if err := c.bindProcessInputs(stmt, o, prepared); err != nil {
					return err
				}
			}
		}
		return phrase(o.phrase)
	}
	return stmt.sGram.walkRegistration(func(site registrationSite, _ []*token, _ bool) error {
		if site.subject != nil {
			if err := phrase(site.subject.phrase); err != nil {
				return err
			}
		}
		if err := object(site.valueObject()); err != nil {
			return err
		}
		if site.subject != nil && site.subject.noun != nil {
			if binding := prepared.processBindings[site.subject.noun]; binding != nil && binding.Record == nil {
				if handled, err := c.prepareDefinitionBindingSite(stmt, site, prepared); err != nil {
					return err
				} else if handled {
					binding.Record = prepared.declarations[len(prepared.declarations)-1].Record
					binding.Record.storageScope = binding.Scope
					binding.Scope.storeBinding(binding.Record.ID, binding.Record)
				}
			}
		}
		if site.clause != nil {
			for _, o := range site.clause.orderedObjects() {
				if err := object(o); err != nil {
					return err
				}
			}
			return phrase(site.clause.phrase)
		}
		return nil
	})
}

func (c *Compilation) bindProcessInputs(stmt *lStatement, object *objectGram, prepared *statementRegistration) error {
	signature, body := object.signaturePhrase(), object.directivePhrase()
	if signature != nil && len(signature.orderedObjects()) != 0 {
		return lerrors.Unsupported.Errorf("structured Process inputs do not declare callable outputs")
	}
	scope := &lexicalScope{kind: lexicalProcess, unit: object.noun.scope.unit, parent: object.noun.scope, records: make(map[string]*record), source: stmt.source.Span}
	var subjects []*subjectGram
	if signature != nil {
		subjects = signature.subjects
	}
	for _, input := range subjects {
		if input.phrase == nil || input.phrase.obj == nil {
			return lerrors.InvalidInput.Errorf("Process input requires an explicit type")
		}
		nativeType, dependency, err := c.nativeSignatureType(input.phrase.obj.noun)
		if err != nil {
			return err
		}
		if nativeType == "" {
			return lerrors.Unsupported.Errorf("Process input requires an explicit native type or selected native declaration")
		}
		tok := input.noun
		rec, err := nativeReferenceRecord(stmt, tok, "")
		if err != nil {
			return err
		}
		if _, exists := scope.records[rec.ID]; exists {
			return lerrors.DuplicateDeclaration.Errorf("duplicate Process input %s", rec.ID)
		}
		declaredType := definitionTypeArgument(nativeType)
		declaredType.source = stmt.source.Span
		if dependency != nil {
			declaredType.dependencies = append(declaredType.dependencies, *dependency)
		}
		rec.definitionArgument.processType = &declaredType
		rec.storageScope = scope
		reservation, err := prepared.processToken(c, tok, scope)
		if err != nil {
			return err
		}
		reservation.Record = rec
		scope.storeBinding(rec.ID, rec)
		prepared.declarations = append(prepared.declarations, extractedDeclaration{Token: tok, Record: rec})
	}
	for index, tok := range body.tokens {
		if _, err := prepared.processToken(c, tok, scope); err != nil {
			return err
		}
		if !tok.isDeclare() {
			if index > 0 {
				if _, member := possessiveOwner(string(body.tokens[index-1].rawBytes)); member {
					continue // The selected owner, not a new local homonym, owns this member.
				}
			}
			resolver := c.resolver()
			resolver.expectedPOS = bindingContextPOS(body.tokens, index)
			if state, err := resolver.bindToken(tok); err != nil {
				return err
			} else if state == lookupFound {
				tok.pendingDeclaration = nil // Actual lexical ownership supersedes grammar scaffolding.
			}
		}
	}
	return nil
}

// Only tokens are borrowed during preparation. Reservations and records remain
// in fresh Process scopes until the whole statement passes publication checks.
func (prepared *statementRegistration) processToken(c *Compilation, tok *token, scope *lexicalScope) (*scopeDeclaration, error) {
	if prepared.processTokens == nil {
		prepared.processTokens = make(map[*token]token)
		prepared.processBindings = make(map[*token]*scopeDeclaration)
	}
	if _, saved := prepared.processTokens[tok]; !saved {
		prepared.processTokens[tok] = *tok
	}
	tok.scope = scope
	if !tok.isDeclare() {
		return nil, nil
	}
	reservation := prepared.processBindings[tok]
	if reservation == nil {
		original := c.Scope.declarations[tok]
		if original == nil {
			return nil, lerrors.Internal.Errorf("Process declaration %q has no reserved declaration", tok.rawBytes)
		}
		copy := *original
		reservation = &copy
		prepared.processBindings[tok] = reservation
	}
	moveProcessReservation(reservation, scope)
	return reservation, nil
}

func moveProcessReservation(reservation *scopeDeclaration, scope *lexicalScope) {
	previous := reservation.Scope
	for index, entry := range previous.declarations {
		if entry == reservation {
			previous.declarations = append(previous.declarations[:index], previous.declarations[index+1:]...)
			previous.reservations, previous.indexedDeclarations = nil, 0
			break
		}
	}
	reservation.Scope = scope
	reservation.provisional = false
	scope.declarations = append(scope.declarations, reservation)
}

func (c *Compilation) processInputArguments(object *objectGram) []DefinitionArgument {
	if object == nil || object.directivePhrase() == nil {
		return nil
	}
	var inputs []DefinitionArgument
	var subjects []*subjectGram
	if signature := object.signaturePhrase(); signature != nil {
		subjects = signature.subjects
	}
	for _, input := range subjects {
		rec := c.Semantics.Subjects[input.noun]
		if rec == nil || rec.storageScope.kind != lexicalProcess {
			continue
		}
		argument := *rec.definitionArgument
		argument.reference = c.definitionLabel()
		*rec.definitionArgument = argument
		inputs = append(inputs, argument)
	}
	for _, tok := range object.directivePhrase().tokens {
		if rec := c.Semantics.Subjects[tok]; tok.isDeclare() && rec != nil && rec.staticConstructor != nil {
			name := c.definitionLabel()
			rec.staticConstructor.Name, rec.definitionArgument.reference = name, name
		}
	}
	return inputs
}

// The surrounding pattern or iteration introduces the hygienic input name.
// Rebind that same name through frozen AML declaration/assignment, so native
// checking enforces its authored type and ownership without
// evaluating or moving an inactive branch's payload outside that branch.
func (c *Compilation) annotateProcessInputs(body DefinitionInstance, inputs []DefinitionArgument) (DefinitionInstance, error) {
	var annotations []DefinitionInstance
	for _, input := range inputs {
		if input.processType == nil || input.processType.typeAtom == "_" {
			continue
		}
		annotation, err := composeNativeAML("OB{|Key: ~<@Temporary|Label@>~, Value: ~<@Value@>~|}*BO{|Key: ~<@Name|Label@>~, Type: ~<@Type|Type@>~|}*O{|Key: ~<@Name|Label@>~, Value: ~<@Temporary|Label@>~|}", input.source, map[string]DefinitionArgument{
			"Temporary": definitionReferenceArgument(c.definitionLabel()), "Name": input, "Type": *input.processType, "Value": input,
		})
		if err != nil {
			return DefinitionInstance{}, err
		}
		annotations = append(annotations, annotation)
	}
	if len(annotations) == 0 {
		return body, nil
	}
	steps := append([]DefinitionInstance(nil), annotations...)
	for _, step := range definitionSteps(body) {
		steps = appendDefinitionSequence(steps, step)
	}
	expressions := make([]string, len(steps))
	for index, step := range steps {
		expressions[index] = step.Expression
	}
	var bindings []DefinitionBinding
	var dependencies []DefinitionInstance
	for _, annotation := range annotations {
		bindings = append(bindings, annotation.Bindings...)
		dependencies = append(dependencies, annotation.Dependencies...)
	}
	body.Bindings = append(bindings, body.Bindings...)
	body.Dependencies = append(dependencies, body.Dependencies...)
	body.Expression, body.sequence, body.before = strings.Join(expressions, "*"), steps, nil
	return body, nil
}

func scopedDefinitionArguments(template *DefinitionTemplate, args []DefinitionArgument) ([]DefinitionArgument, error) {
	var inputs []DefinitionArgument
	for _, argument := range args {
		if argument.Definition != nil {
			inputs = append(inputs, argument.Definition.inputs...)
		}
	}
	if len(inputs) == 0 {
		return args, nil
	}
	if len(inputs) != len(template.scopedLabels) {
		return nil, lerrors.InvalidInput.Errorf("structured body provides %d inputs for %d pattern/iteration bindings", len(inputs), len(template.scopedLabels))
	}
	bound := make([]DefinitionArgument, 0, len(args)+len(inputs))
	input, next := 0, 0
	for _, slot := range template.Slots {
		if template.scopedLabels[slot.Name] {
			position := input
			if len(inputs) == 2 && template.scopedIndex != "" {
				position = 1
				if slot.Name == template.scopedIndex {
					position = 0
				}
			}
			bound = append(bound, inputs[position])
			input++
		} else if next < len(args) {
			argument := args[next]
			if argument.Definition != nil && len(argument.Definition.inputs) > 0 {
				body := *argument.Definition
				body.inputs = nil
				argument.Definition = &body
			}
			bound = append(bound, argument)
			next++
		}
	}
	return append(bound, args[next:]...), nil
}
