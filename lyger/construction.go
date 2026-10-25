package Lyger

import (
	"fmt"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// constructionNode is a bound instance of a source-described application
// component. Nominal remains the source-defined type; function structure is
// recognized from the bound definition when lowered.
type constructionNode struct {
	Nominal      string
	Name         string
	Definition   *record
	Source       *objectGram
	Children     []*constructionNode
	Wrappers     []*descriptorRule
	Parameters   []*objectGram
	EmptyInputs  bool
	EmptyOutputs bool
}

type applicationConstruction struct {
	App    *App
	Root   *constructionNode
	Source *lStatement
	Body   []DefinitionInstance
}

func (c *Compilation) descriptorFor(tok *token) (*descriptorDefinition, error) {
	if tok == nil || tok.binding == nil {
		return nil, nil
	}
	rec, err := canonicalRecord(tok.binding.Record, c.Semantics.Aliases)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, lerrors.Internal.Errorf("descriptor reference has no record")
	}
	return c.Semantics.Descriptors[rec], nil
}

func (dst *constructionNode) mergeDescriptor(src *constructionNode) {
	if src.Name != "" {
		dst.Name = src.Name
	}
	if src.Definition != nil {
		dst.Definition = src.Definition
	}
	dst.Children = append(dst.Children, src.Children...)
	dst.Wrappers = append(dst.Wrappers, src.Wrappers...)
	dst.Parameters = append(dst.Parameters, src.Parameters...)
	dst.EmptyInputs = dst.EmptyInputs || src.EmptyInputs
	dst.EmptyOutputs = dst.EmptyOutputs || src.EmptyOutputs
}

func (c *Compilation) constructObject(obj *objectGram, active map[*descriptorRule]bool) (*constructionNode, error) {
	if obj == nil || obj.noun == nil {
		return nil, lerrors.Internal.Errorf("cannot construct an object without a nominal type")
	}
	nominal, err := nominalType(obj.noun)
	if err != nil {
		return nil, err
	}
	node := &constructionNode{Nominal: nominal, Source: obj}
	if obj.noun.binding != nil {
		node.Definition, err = canonicalRecord(obj.noun.binding.Record, c.Semantics.Aliases)
		if err != nil {
			return nil, err
		}
	}
	if obj.name != nil && obj.name.noun != nil {
		node.Name = unquoteImportWord(obj.name.noun.rawBytes)
	}
	for _, adjective := range valueModifiers(obj.adjectives) {
		descriptor, err := c.descriptorFor(adjective)
		if err != nil {
			return nil, err
		}
		if descriptor == nil {
			continue
		}
		target, err := c.resolver().nominalIdentity(obj.noun)
		if err != nil {
			return nil, err
		}
		matched := false
		for _, rule := range descriptor.Rules {
			if rule.Target != target {
				continue
			}
			matched = true
			if rule.Template != nil {
				node.Wrappers = append(node.Wrappers, rule)
				continue
			}
			if active[rule] {
				return nil, lerrors.Cycle.Errorf("recursive descriptor %s for %s", descriptor.Record.ID, node.Nominal)
			}
			active[rule] = true
			instance, err := c.constructObject(rule.Object, active)
			delete(active, rule)
			if err != nil {
				return nil, err
			}
			node.mergeDescriptor(instance)
		}
		if !matched {
			return nil, lerrors.Unsupported.Errorf("descriptor %s has no constructor for %s", descriptor.Record.ID, node.Nominal)
		}
	}
	if err := c.constructRelations(node, obj.phrase, active); err != nil {
		return nil, err
	}
	if obj.clause != nil {
		// A source-defined `using` relation supplies the enclosing parameter;
		// it does not create another application component.
		node.Parameters = append(node.Parameters, obj.clause.orderedObjects()...)
	}
	return node, nil
}

func (c *Compilation) constructRelations(node *constructionNode, phrase *phraseGram, active map[*descriptorRule]bool) error {
	if phrase == nil {
		return nil
	}
	association := false
	for _, prep := range phrase.prepositions {
		if prep.hasRole("preposition", "consignment.associative") {
			association = true
		}
	}
	for _, obj := range phrase.orderedObjects() {
		if obj.noun == nil {
			return lerrors.Internal.Errorf("constructor relation has no object")
		}
		if obj.noun.hasRole("noun", "functional.inflective") || obj.noun.hasRole("noun", "functional.deflective") {
			// Signature absence is explicit in the source (No Inputs/Outputs).
			absent := false
			for _, tok := range obj.tokens {
				if tok.hasRole("article", "none") {
					absent = true
				}
			}
			if !absent {
				kind := "output"
				if obj.noun.hasRole("noun", "functional.inflective") {
					kind = "input"
				}
				return lerrors.Unsupported.Errorf("nonempty function %s signature requires explicit property values", kind)
			}
			if obj.noun.hasRole("noun", "functional.inflective") {
				node.EmptyInputs = true
			} else {
				node.EmptyOutputs = true
			}
			continue
		}
		if association {
			child, err := c.constructObject(obj, active)
			if err != nil {
				return err
			}
			node.Children = append(node.Children, child)
		} else {
			node.Parameters = append(node.Parameters, obj)
			if err := c.constructRelations(node, obj.phrase, active); err != nil {
				return err
			}
		}
	}
	return c.constructRelations(node, phrase.phrase, active)
}

// declarationObject selects the value belonging to one source declaration,
// including coordinated passive declarations and conditional declaration clauses.
func declarationObject(statement *statementGram, target *token) *objectGram {
	var clauseObject func(*clauseGram) *objectGram
	predicateObject := func(predicate *predicateGram) *objectGram {
		if predicate == nil {
			return nil
		}
		for _, clause := range predicate.clauses {
			if obj := clauseObject(clause); obj != nil {
				return obj
			}
		}
		for _, conditional := range predicate.conditionals {
			for _, clause := range conditional.clauses {
				if obj := clauseObject(clause); obj != nil {
					return obj
				}
			}
		}
		return nil
	}
	clauseObject = func(clause *clauseGram) *objectGram {
		if clause.subject != nil && clause.subject.noun == target {
			if clause.predicate != nil {
				return clause.predicate.obj
			}
			if clause.subject.phrase != nil {
				return clause.subject.phrase.obj
			}
		}
		if obj := predicateObject(clause.predicate); obj != nil {
			return obj
		}
		for _, child := range clause.clauses {
			if obj := clauseObject(child); obj != nil {
				return obj
			}
		}
		return nil
	}
	if statement.subject != nil && statement.subject.noun == target && statement.predicate != nil {
		return statement.predicate.obj
	}
	if obj := predicateObject(statement.predicate); obj != nil {
		return obj
	}
	for _, clause := range statement.clauses {
		if obj := clauseObject(clause); obj != nil {
			return obj
		}
	}
	return nil
}

func (c *Compilation) constructApplications() error {
	// Registration can retry forward references; emission follows declarations.
	var constructions []*applicationConstruction
	var apps []*App
	seen := make(map[*App]bool)
	rootApps := make(map[*App]bool)
	for _, scoped := range c.Scope.Ordered {
		source := scoped.Statement
		for _, tok := range source.tokens {
			rec := c.Semantics.Subjects[tok]
			if rec == nil || c.Semantics.Aliases[rec] != nil {
				continue
			}
			value, err := rec.Value.value()
			if err != nil {
				return fmt.Errorf("application declaration %s: %w", rec.ID, err)
			}
			if value.Type != valueApp {
				continue
			}
			app := value.Data.(*App) //nolint:errcheck // value() validates the tagged payload before this type-specific projection.
			if seen[app] {
				continue
			}
			node, err := c.constructObject(declarationObject(source.sGram, tok), make(map[*descriptorRule]bool))
			if err != nil {
				return fmt.Errorf("construct %s: %w", app.ID, err)
			}
			node.Name = app.ID
			seen[app] = true
			rootApps[app] = rec.storageScope == c.Scope.Root
			apps = append(apps, app)
			constructions = append(constructions, &applicationConstruction{App: app, Root: node, Source: source})
		}
	}
	for _, app := range c.Semantics.Apps {
		if !seen[app] {
			return lerrors.Internal.Errorf("application %s has no source declaration", app.ID)
		}
	}
	// Reserve source names before allocating suffixes, including names that
	// already look generated. Case folding also avoids native filesystem clashes.
	counts := make(map[string]int)
	used := make(map[string]bool)
	for _, app := range apps {
		key := strings.ToLower(app.ID)
		counts[key]++
		used[key] = true
	}
	for _, app := range apps {
		app.buildID = ""
		if rootApps[app] || counts[strings.ToLower(app.ID)] == 1 {
			continue
		}
		for suffix := 1; ; suffix++ {
			candidate := fmt.Sprintf("%s_%d", app.ID, suffix)
			key := strings.ToLower(candidate)
			if used[key] {
				continue
			}
			app.buildID = candidate
			used[key] = true
			break
		}
	}
	c.Semantics.Apps, c.Semantics.Constructions = apps, constructions
	return nil
}

func (c *Compilation) validateConstructionParameters(node *constructionNode, body []DefinitionInstance) error {
	for _, parameter := range node.Parameters {
		if parameter.noun == nil || parameter.noun.binding == nil {
			return lerrors.UnresolvedReference.Errorf("constructor %s has an unbound parameter", node.Nominal)
		}
		rec, err := canonicalRecord(parameter.noun.binding.Record, c.Semantics.Aliases)
		if err != nil {
			return err
		}
		if rec == nil {
			return lerrors.UnresolvedReference.Errorf("constructor %s has an unbound parameter", node.Nominal)
		}
		expected := rec.ID
		if len(body) == 0 {
			return lerrors.InvalidInput.Errorf("constructor %s is missing its %s argument", node.Nominal, expected)
		}
		for _, instance := range body {
			if instance.definition != rec || !strings.EqualFold(instance.Symbol, expected) {
				return lerrors.TypeMismatch.Errorf("constructor %s requires %s, received %s", node.Nominal, expected, instance.Symbol)
			}
		}
	}
	for _, child := range node.Children {
		if err := c.validateConstructionParameters(child, body); err != nil {
			return err
		}
	}
	return nil
}

// An App with no actions supplies a zero-element sequence to its constructor,
// rather than omitting the constructor argument. A source-defined variadic
// child container (for example W^<@Steps|Syllables@>) retains its parent glyph
// with no children. Other required parameters still have no implicit value.
func (c *Compilation) emptyConstructionSequence(node *constructionNode) (*DefinitionInstance, error) {
	if len(node.Parameters) > 0 {
		parameter := node.Parameters[0]
		if parameter.noun == nil || parameter.noun.binding == nil {
			return nil, nil // Ordinary parameter validation reports the missing binding.
		}
		rec, err := canonicalRecord(parameter.noun.binding.Record, c.Semantics.Aliases)
		if err != nil || rec == nil {
			return nil, err
		}
		template, err := c.recordTemplate(rec)
		if err != nil || template == nil {
			return nil, err
		}
		if len(template.Slots) != 1 || template.Slots[0].Type != definitionArgSyllables.String() || len(template.parts) != 3 ||
			template.parts[1].slot == nil || strings.TrimSpace(template.parts[2].text) != "" {
			return nil, nil
		}
		parent := strings.TrimSpace(template.parts[0].text)
		if !strings.HasSuffix(parent, "^") {
			return nil, nil
		}
		parent = strings.TrimSpace(strings.TrimSuffix(parent, "^"))
		empty := *template
		empty.Slots = nil
		empty.parts = []definitionPart{{text: parent}}
		instance, err := empty.Instantiate(nil)
		if err != nil {
			return nil, err
		}
		instance.Symbol, instance.definition = rec.ID, rec
		instance.Bindings = []DefinitionBinding{{Slot: template.Slots[0], Kind: definitionArgSyllables.String()}}
		return &instance, nil
	}
	for _, child := range node.Children {
		instance, err := c.emptyConstructionSequence(child)
		if err != nil || instance != nil {
			return instance, err
		}
	}
	return nil, nil
}
