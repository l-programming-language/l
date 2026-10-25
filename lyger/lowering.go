package Lyger

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

func (c *Compilation) recordTemplate(rec *record) (*DefinitionTemplate, error) {
	rec, err := canonicalRecord(rec, c.Semantics.Aliases)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, lerrors.Internal.Errorf("definition reference has no record")
	}
	source := c.Semantics.Declarations[rec]
	if source == nil {
		return nil, lerrors.Internal.Errorf("definition %s has no source declaration", rec.ID)
	}
	sites, err := source.sGram.registrationSites()
	if err != nil {
		return nil, err
	}
	for _, site := range sites {
		target, err := site.target()
		if err != nil {
			return nil, err
		}
		if c.Semantics.Subjects[target] != rec {
			continue
		}
		literal, err := site.literal()
		if err != nil {
			return nil, err
		}
		if literal == nil {
			return nil, nil
		}
		return ParseDefinitionTemplate(literal.rawBytes, source.source.Span)
	}
	return nil, lerrors.Internal.Errorf("definition %s has no source declaration site", rec.ID)
}

func definitionArguments(template *DefinitionTemplate, args []DefinitionArgument) (map[string]DefinitionArgument, error) {
	bindings := make(map[string]DefinitionArgument)
	next := 0
	for _, slot := range template.Slots {
		slotKind, _ := parseDefinitionArgKindExact(slot.Type)
		if slotKind == definitionArgSyllables {
			var expressions []string
			for ; next < len(args); next++ {
				if args[next].Definition != nil {
					expressions = append(expressions, args[next].Definition.Expression)
				} else if kind, _ := parseDefinitionArgKindExact(args[next].Kind); kind == definitionArgSyllable {
					expressions = append(expressions, args[next].Text)
				} else {
					return nil, lerrors.TypeMismatch.Errorf("%s requires definition syllables, got %s", slot.Name, args[next].Kind)
				}
			}
			bindings[slot.Name] = DefinitionArgument{Kind: definitionArgSyllables.String(), Syllables: expressions}
			continue
		}
		if next >= len(args) {
			return nil, lerrors.InvalidInput.Errorf("missing argument for %s", slot.Name)
		}
		bindings[slot.Name] = args[next]
		next++
	}
	if next != len(args) {
		return nil, lerrors.InvalidInput.Errorf("definition received %d unused arguments", len(args)-next)
	}
	return bindings, nil
}

func (c *Compilation) instantiateCall(rec *record, args []DefinitionArgument, active map[*record]bool) (DefinitionInstance, error) {
	rec, err := canonicalRecord(rec, c.Semantics.Aliases)
	if err != nil {
		return DefinitionInstance{}, err
	}
	if rec == nil {
		return DefinitionInstance{}, lerrors.Internal.Errorf("definition invocation has no record")
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
		bindings, err := definitionArguments(template, args)
		if err != nil {
			return DefinitionInstance{}, fmt.Errorf("%s: %w", rec.ID, err)
		}
		instance, err := template.Instantiate(bindings)
		instance.Symbol = rec.ID
		instance.definition = rec
		return instance, err
	}
	source := c.Semantics.Declarations[rec]
	if source.sGram.predicate == nil || len(source.sGram.predicate.clauses) == 0 {
		return DefinitionInstance{}, lerrors.InvalidInput.Errorf("%s has no action definition", rec.ID)
	}
	clause := source.sGram.predicate.clauses[0]
	if clause.subject == nil || clause.subject.phrase == nil {
		return DefinitionInstance{}, lerrors.InvalidInput.Errorf("%s has no directive body", rec.ID)
	}
	var expressions []string
	resultSymbol := ""
	var resultDefinition *record
	for _, action := range clause.subject.phrase.clauses {
		if action.verb == nil || action.verb.binding == nil {
			return DefinitionInstance{}, lerrors.TypeMismatch.Errorf("source action in %s is not a definition", rec.ID)
		}
		var actionArgs []DefinitionArgument
		for _, obj := range action.orderedObjects() {
			if obj.noun == nil || obj.noun.binding == nil {
				return DefinitionInstance{}, lerrors.UnresolvedReference.Errorf("action operand in %s has no bound definition", rec.ID)
			}
			child, err := c.instantiateCall(obj.noun.binding.Record, args, active)
			if err != nil {
				return DefinitionInstance{}, err
			}
			actionArgs = append(actionArgs, DefinitionArgument{Kind: definitionArgDefinition.String(), Definition: &child})
		}
		instance, err := c.instantiateCall(action.verb.binding.Record, actionArgs, active)
		if err != nil {
			return DefinitionInstance{}, err
		}
		expressions = append(expressions, instance.Expression)
		if resultDefinition != nil && (resultDefinition != instance.definition || resultSymbol != instance.Symbol) {
			return DefinitionInstance{}, lerrors.TypeMismatch.Errorf("action definition %s returns incompatible nominal types", rec.ID)
		}
		resultSymbol = instance.Symbol
		resultDefinition = instance.definition
	}
	if len(expressions) == 0 {
		return DefinitionInstance{}, lerrors.InvalidInput.Errorf("%s has an empty action definition", rec.ID)
	}
	expression := strings.Join(expressions, "*")
	return DefinitionInstance{Symbol: resultSymbol, definition: resultDefinition, Original: append([]byte(nil), source.source.Raw...), Span: source.source.Span, Expression: expression, OriginalExpression: expression}, nil
}

func argumentFromToken(tok *token) (DefinitionArgument, error) {
	if tok == nil || tok.mem == nil {
		return DefinitionArgument{}, lerrors.Internal.Errorf("argument has no value")
	}
	if tok.mem.Runtime != nil {
		return DefinitionArgument{}, lerrors.Unsupported.Errorf("runtime %s.%s is unavailable during source generation", tok.mem.Runtime.OwnerID, tok.mem.Runtime.Member)
	}
	value, err := tok.mem.value()
	if err != nil {
		return DefinitionArgument{}, err
	}
	kind, supported := value.Type.DefinitionArgKind()
	if !supported {
		return DefinitionArgument{}, lerrors.Unsupported.Errorf("unsupported argument value %s", value.Type)
	}
	argument := DefinitionArgument{Kind: kind.String()}
	switch kind {
	case definitionArgText:
		argument.Text = unquoteImportWord([]byte(value.Data.(string))) //nolint:errcheck // value() validates the tagged payload before this type-specific projection.
	case definitionArgNumber:
		argument.Text = strconv.FormatFloat(value.Data.(float64), 'g', -1, 64) //nolint:errcheck // value() validates the tagged payload before this type-specific projection.
	case definitionArgTruth:
		argument.Text = strconv.FormatBool(value.Data.(bool)) //nolint:errcheck // value() validates the tagged payload before this type-specific projection.
	}
	return argument, nil
}

func (c *Compilation) statementApp(stmt *lStatement) (*App, error) {
	if stmt.sGram.subject == nil || stmt.sGram.subject.noun == nil {
		return nil, nil
	}
	noun := stmt.sGram.subject.noun
	if noun.isDeclare() {
		return nil, nil
	}
	if noun.binding != nil {
		rec, err := c.resolver().reference(noun)
		if err != nil {
			return nil, err
		}
		if rec != nil {
			value, err := rec.Value.value()
			if err != nil {
				return nil, err
			}
			if value.Type == valueApp {
				return value.Data.(*App), nil //nolint:errcheck // value() validates the tagged payload before this type-specific projection.
			}
			if rec.Value.Type == storageApplicative {
				return nil, nil
			}
		}
	}
	nominal, err := nominalType(noun)
	if err != nil {
		return nil, err
	}
	if nominal != "app" {
		return nil, nil
	}
	for _, scoped := range c.Scope.Ordered {
		if scoped.Statement != stmt {
			continue
		}
		rec, err := c.resolver().reference(scoped.IncomingSubjectToken)
		if err != nil {
			return nil, err
		}
		if rec != nil {
			value, err := rec.Value.value()
			if err != nil {
				return nil, err
			}
			if value.Type == valueApp {
				return value.Data.(*App), nil //nolint:errcheck // value() validates the tagged payload before this type-specific projection.
			}
		}
	}
	return nil, lerrors.UnresolvedReference.Errorf("definite App reference has no application subject in its paragraph")
}

func (c *Compilation) lowerAppBody(app *App) ([]DefinitionInstance, error) {
	var body []DefinitionInstance
	for _, stmt := range c.Statements {
		if stmt.passiveLifecycle() {
			continue
		}
		target, err := c.statementApp(stmt)
		if err != nil {
			return nil, err
		}
		if target != app {
			continue
		}
		for _, action := range stmt.sGram.predicate.clauses {
			if action.verb == nil || action.verb.binding == nil {
				if action.verb != nil {
					if role, native := action.verb.nativeRole(); native {
						if kind, ok := role.InstructionKind(); ok {
							return nil, stmt.source.Span.wrapError(lerrors.Unsupported.Errorf("%s is unsupported in an App body; lifecycle operations belong to the outer orchestrator", kind))
						}
					}
				}
				return nil, lerrors.Unsupported.Errorf("application action has no source-defined behavior")
			}
			var args []DefinitionArgument
			for _, obj := range action.orderedObjects() {
				argument, err := argumentFromToken(obj.noun)
				if err != nil {
					return nil, err
				}
				args = append(args, argument)
			}
			instance, err := c.instantiateCall(action.verb.binding.Record, args, make(map[*record]bool))
			if err != nil {
				return nil, err
			}
			body = append(body, instance)
		}
	}
	return body, nil
}

func (c *Compilation) lowerFunction(node *constructionNode, body []DefinitionInstance) (DefinitionInstance, error) {
	if node.Definition == nil {
		return DefinitionInstance{}, lerrors.InvalidInput.Errorf("function %s has no definition", node.Name)
	}
	template, err := c.recordTemplate(node.Definition)
	if err != nil {
		return DefinitionInstance{}, err
	}
	if template == nil || !template.IsFunction() {
		return DefinitionInstance{}, lerrors.TypeMismatch.Errorf("%s does not define a function structure", node.Nominal)
	}
	if node.Name == "" || !node.EmptyInputs || !node.EmptyOutputs {
		return DefinitionInstance{}, lerrors.InvalidInput.Errorf("function constructor requires name and explicit signature")
	}
	bindings := make(map[string]DefinitionArgument)
	var expressions []string
	for _, instance := range body {
		expressions = append(expressions, instance.Expression)
	}
	process := DefinitionInstance{Expression: strings.Join(expressions, "*")}
	for _, slot := range template.Slots {
		if slot.Type == "" {
			matched := false
			for _, parameter := range node.Parameters {
				nominal, err := nominalType(parameter.noun)
				if err != nil {
					return DefinitionInstance{}, err
				}
				if strings.EqualFold(nominal, slot.Name) {
					matched = true
				}
			}
			if !matched {
				return DefinitionInstance{}, lerrors.InvalidInput.Errorf("function slot %s has no declared constructor parameter", slot.Name)
			}
			bindings[slot.Name] = DefinitionArgument{Kind: definitionArgDefinition.String(), Definition: &process}
			continue
		}
		kind, _ := parseDefinitionArgKindExact(slot.Type)
		switch kind {
		case definitionArgText:
			bindings[slot.Name] = DefinitionArgument{Kind: definitionArgText.String(), Text: node.Name}
		case definitionArgProps:
			bindings[slot.Name] = DefinitionArgument{Kind: definitionArgProps.String(), Props: []DefinitionProperty{}}
		case definitionArgTexts:
			bindings[slot.Name] = DefinitionArgument{Kind: definitionArgTexts.String(), Texts: []string{}}
		default:
			return DefinitionInstance{}, lerrors.Unsupported.Errorf("unsupported function slot %s: %s", slot.Name, slot.Type)
		}
	}
	instance, err := template.Instantiate(bindings)
	if err != nil {
		return DefinitionInstance{}, err
	}
	for i := len(node.Wrappers) - 1; i >= 0; i-- {
		rule := node.Wrappers[i]
		wrapper, err := ParseDefinitionTemplate(rule.Template.rawBytes, rule.Source.source.Span)
		if err != nil {
			return DefinitionInstance{}, err
		}
		args, err := definitionArguments(wrapper, []DefinitionArgument{{Kind: definitionArgDefinition.String(), Definition: &instance}})
		if err != nil {
			return DefinitionInstance{}, err
		}
		instance, err = wrapper.Instantiate(args)
		if err != nil {
			return DefinitionInstance{}, err
		}
	}
	instance.definition, err = canonicalRecord(node.Definition, c.Semantics.Aliases)
	if err != nil {
		return DefinitionInstance{}, err
	}
	instance.Symbol = instance.definition.ID
	return instance, nil
}
