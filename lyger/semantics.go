package Lyger

import (
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// semanticProgram supplements lexical memory with source-defined constructor
// rules and ordered effects. Grammar nodes remain attached for attribution.
type semanticProgram struct {
	Descriptors   map[*record]*descriptorDefinition
	Declarations  map[*record]*lStatement
	Aliases       map[*record]*record
	Subjects      map[*token]*record
	Apps          []*App
	Constructions []*applicationConstruction
	Instructions  []*instruction
	Continuations map[*App]*scopedInstructions
	Functions     map[*record]*sourceFunction
	Returns       map[*record]bool
}

type descriptorDefinition struct {
	Record  *record
	Targets map[nominalIdentity]bool
	Rules   []*descriptorRule
}

type descriptorRule struct {
	Source   *lStatement
	Target   nominalIdentity
	Object   *objectGram
	Template *token
}

// runtimeReference gives a typed value a deferred source, without pretending
// that a compilation has already obtained application stdout.
type runtimeReference struct {
	Expression *objectGram
	Owner      *record
	OwnerID    string
	Member     nativeMember
	Source     *lStatement
}

func newSemanticProgram() *semanticProgram {
	return &semanticProgram{Functions: make(map[*record]*sourceFunction), Returns: make(map[*record]bool), Descriptors: make(map[*record]*descriptorDefinition), Declarations: make(map[*record]*lStatement), Aliases: make(map[*record]*record), Subjects: make(map[*token]*record)}
}

func symbolRoot(tok *token) string {
	if tok == nil {
		return ""
	}
	if tok.binding != nil {
		return strings.ToLower(tok.binding.Record.ID)
	}
	return strings.ToLower(strings.Trim(strings.Split(string(tok.rawBytes), "|")[0], `"`))
}

func nominalType(tok *token) (string, error) {
	if tok == nil {
		return "", nil
	}
	if tok.mem != nil {
		// Unbound identifier tokens carry transitional primitive storage until
		// their source-defined nominal name is registered.
		if tok.binding == nil && tok.mem.unresolved() {
			return symbolRoot(tok), nil
		}
		value, err := tok.mem.value()
		if err != nil {
			return "", err
		}
		if tok.binding != nil {
			return strings.ToLower(tok.binding.Root), nil
		}
		if value.Type == valueLexeme {
			return strings.ToLower(value.Data.(*lexeme).Root), nil //nolint:errcheck // value() validates the tagged payload before this type-specific projection.
		}
		return strings.ToLower(value.Type.String()), nil
	}
	if tok.binding != nil {
		return strings.ToLower(tok.binding.Root), nil
	}
	return symbolRoot(tok), nil
}

func definitionLiteral(tokens []*token) (*token, error) {
	for _, tok := range tokens {
		if tok == nil || tok.mem == nil || tok.mem.Type != storagePrimitive {
			continue
		}
		if tok.binding == nil && tok.mem.unresolved() {
			continue
		}
		value, err := tok.mem.value()
		if err != nil {
			return nil, err
		}
		if value.isDef() {
			return tok, nil
		}
	}
	return nil, nil
}

// A statement is prepared in full before any semantic relationships are published.
type statementRegistration struct {
	declarations    []extractedDeclaration
	descriptors     []*descriptorDefinition
	functions       []*sourceFunction
	processTokens   map[*token]token
	processBindings map[*token]*scopeDeclaration
}

func (d *descriptorDefinition) clone() *descriptorDefinition {
	copy := &descriptorDefinition{Record: d.Record, Targets: make(map[nominalIdentity]bool), Rules: append([]*descriptorRule(nil), d.Rules...)}
	for target, enabled := range d.Targets {
		copy.Targets[target] = enabled
	}
	return copy
}

func (c *Compilation) prepareAlias(site registrationSite) (extractedDeclaration, error) {
	source, alias, err := site.aliasOperands()
	if err != nil {
		return extractedDeclaration{}, err
	}
	// Keep the source edge rather than flattening it to the canonical value.
	binding := source.binding
	if binding == nil {
		binding, _, err = c.resolver().lookup(source)
	}
	if err != nil {
		return extractedDeclaration{}, err
	}
	if binding == nil {
		return extractedDeclaration{}, lerrors.UnresolvedReference.Errorf("alias source %q is unresolved", source.rawBytes)
	}
	target := binding.Record
	canonical, err := canonicalRecord(target, c.Semantics.Aliases)
	if err != nil {
		return extractedDeclaration{}, err
	}
	if _, err := target.Value.value(); err != nil {
		return extractedDeclaration{}, err
	}
	value := &entry{}
	value.bindFrom(target.Value)
	label, err := parseDeclarationLabel(string(alias.rawBytes))
	if err != nil {
		return extractedDeclaration{}, err
	}
	rec := &record{ID: label.Root, Type: target.Type, Value: value, function: canonical.function}
	if original := target.bindingWord(); original != nil {
		word := original.cloneUse()
		if binding.POS == "verb" && binding.Form == "objective" {
			word.POS = "noun"
		}
		// Only an actual possessive reference to the unchanged canonical root
		// removes qualification. Renamed or actor aliases cannot launder a
		// different root or role through an intermediate complete family.
		sourceQualified, aliasQualified := false, false
		for i, tok := range site.clause.tokens {
			if i > 0 {
				_, qualified := possessiveOwner(string(site.clause.tokens[i-1].rawBytes))
				if tok == source {
					sourceQualified = qualified
				} else if tok == alias {
					aliasQualified = qualified
				}
			}
		}
		canonicalWord := canonical.bindingWord()
		inherit := len(label.Slots) == 1 && sourceQualified && !aliasQualified && binding.Form == "root" &&
			canonicalWord != nil && canonicalWord.POS == original.POS && canonicalWord.validateFamily() == nil &&
			strings.EqualFold(label.Root, canonicalWord.Root) && strings.EqualFold(label.Root, original.Root) && original.validateFamily() == nil
		if canonical.definitionArgument != nil {
			// Typed inputs and local results use nominal records to carry native
			// references. Renaming their data does not declare vocabulary.
			if err := word.reLabel(string(alias.rawBytes)); err != nil {
				return extractedDeclaration{}, err
			}
		} else if !inherit {
			if err := word.declareFamily(string(alias.rawBytes)); err != nil {
				return extractedDeclaration{}, err
			}
		}
		if value.Type == storageLexeme {
			value.Lexeme = word
			rec.Type = word.getValueType()
		} else {
			rec.word = word
		}
	} else if err := rec.declareInvocation(string(alias.rawBytes)); err != nil {
		return extractedDeclaration{}, err
	}
	return extractedDeclaration{Token: alias, Record: rec, Alias: target}, nil
}

func (c *Compilation) prepareDescriptor(stmt *lStatement, site registrationSite, prepared *statementRegistration) error {
	subject, err := site.target()
	if err != nil {
		return err
	}
	if subject == nil {
		return lerrors.Syntax.Errorf("descriptor requires a subject")
	}
	obj := site.valueObject()
	isDescriptorType := obj != nil && obj.noun != nil && obj.noun.tType == "type.lexeme.adjective"
	rec, err := c.resolver().reference(subject)
	if err != nil {
		return err
	}
	id := symbolRoot(subject)
	if rec != nil {
		id = rec.ID
	}
	var descriptor *descriptorDefinition
	if subject.isDeclare() {
		lex := newLexeme("adjective")
		lit, err := site.literal()
		if err != nil {
			return err
		}
		if lit != nil {
			if err := lex.define(lit.rawBytes); err != nil {
				return err
			}
		}
		if err := lex.reLabel(string(subject.rawBytes)); err != nil {
			return err
		}
		rec = &record{ID: strings.Trim(string(subject.rawBytes), `"`), Type: "lexeme.adjective.generic", Value: &entry{Type: storageLexeme, Lexeme: lex}}
		descriptor = &descriptorDefinition{Record: rec, Targets: make(map[nominalIdentity]bool)}
		prepared.declarations = append(prepared.declarations, extractedDeclaration{Token: subject, Record: rec})
	} else {
		for _, staged := range prepared.descriptors {
			if staged.Record == rec {
				descriptor = staged
				break
			}
		}
		if descriptor == nil && c.Semantics.Descriptors[rec] != nil {
			descriptor = c.Semantics.Descriptors[rec].clone()
		}
	}
	if descriptor == nil {
		return lerrors.TypeMismatch.Errorf("%s is not a descriptor", id)
	}
	if isDescriptorType {
		if obj.clause == nil {
			return lerrors.Syntax.Errorf("descriptor %s has no target declaration", id)
		}
		targets := obj.clause.objects
		if len(targets) == 0 && obj.clause.obj != nil {
			targets = []*objectGram{obj.clause.obj}
		}
		for _, target := range targets {
			name, err := c.resolver().nominalIdentity(target.noun)
			if err != nil {
				return err
			}
			descriptor.Targets[name] = true
		}
	} else {
		if obj == nil || obj.noun == nil {
			return lerrors.Syntax.Errorf("descriptor %s has no described object", id)
		}
		target, err := c.resolver().nominalIdentity(obj.noun)
		if err != nil {
			return err
		}
		if len(descriptor.Targets) > 0 && !descriptor.Targets[target] {
			return lerrors.TypeMismatch.Errorf("descriptor %s cannot describe %s", id, target)
		}
		template, err := site.literal()
		if err != nil {
			return err
		}
		descriptor.Targets[target] = true
		descriptor.Rules = append(descriptor.Rules, &descriptorRule{Source: stmt, Target: target, Object: obj, Template: template})
	}
	for _, staged := range prepared.descriptors {
		if staged == descriptor {
			return nil
		}
	}
	prepared.descriptors = append(prepared.descriptors, descriptor)
	return nil
}

func (c *Compilation) prepareRegistration(stmt *lStatement, sites []registrationSite) (statementRegistration, error) {
	var prepared statementRegistration
	visibilities, err := stmt.sGram.declarationVisibilities()
	if err != nil {
		return prepared, err
	}
	for _, site := range sites {
		if err := c.constructionValueRoles(site.valueObject()); err != nil {
			return statementRegistration{}, err
		}
		if handled, err := c.prepareDefinitionBindingSite(stmt, site, &prepared); handled || err != nil {
			if err != nil {
				return statementRegistration{}, err
			}
			continue
		}
		if handled, err := c.prepareFunctionSite(stmt, site, &prepared); handled || err != nil {
			if err != nil {
				return statementRegistration{}, err
			}
			continue
		}
		if site.isDescriptor() {
			if err := c.prepareDescriptor(stmt, site, &prepared); err != nil {
				return statementRegistration{}, err
			}
			continue
		}
		var declaration extractedDeclaration
		if site.isAlias() {
			declaration, err = c.prepareAlias(site)
		} else {
			declaration, err = site.declaration()
		}
		if err != nil {
			return statementRegistration{}, err
		}
		value := declaration.Record.Value
		if declaration.Alias == nil && value.Lexeme != nil && value.Lexeme.Value == "" && len(value.Lexeme.Actions) > 0 {
			formals, err := sourceParametersIn(stmt, site.parameterTokens())
			if err != nil {
				return statementRegistration{}, err
			}
			value.Lexeme.Inputs = make([]string, len(formals))
			for index, formal := range formals {
				value.Lexeme.Inputs[index] = formal.Name
			}
		}
		var typed *objectGram
		if runtime := value.Runtime; runtime != nil && runtimeReadPhrase(runtime.Expression) != nil {
			typed = runtime.Expression
		} else if value.Applicative != nil && value.Applicative.Function != nil {
			typed = site.valueObject()
		}
		if typed != nil {
			typ, typeErr := sourceTypeFromObject(typed)
			if typeErr == nil {
				typ, typeErr = canonicalRuntimeType(typ, c.Semantics.Aliases)
			}
			if typeErr != nil {
				return statementRegistration{}, typeErr
			}
			declaration.Record.runtimeType = &typ
		}
		prepared.declarations = append(prepared.declarations, declaration)
	}
	for i := range prepared.declarations {
		prepared.declarations[i].Visibility = visibilities[prepared.declarations[i].Token]
	}
	return prepared, nil
}

func (c *Compilation) registerStatement(stmt *lStatement) (err error) {
	sites, err := stmt.sGram.registrationSites()
	if err != nil {
		return err
	}
	prepared, err := c.prepareRegistration(stmt, sites)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			for tok, original := range prepared.processTokens {
				*tok = original
			}
		}
	}()
	if err = c.prepareProcessInputs(stmt, &prepared); err != nil {
		return err
	}
	return c.publishRegistration(stmt, prepared)
}
