package Lyger

import (
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// registrationSite borrows the grammar's declaration and value nodes. Sibling
// clauses are independent sites; clauses inside values remain definition bodies.
type registrationSite struct {
	subject   *subjectGram
	predicate *predicateGram
	clause    *clauseGram
	parameter bool
	resultOf  *clauseGram
}

func (s registrationSite) isAlias() bool {
	return s.clause != nil && strings.HasPrefix(s.clause.sType, "clause.associative.")
}

func (s registrationSite) aliasOperands() (*token, *token, error) {
	if !s.isAlias() {
		return nil, nil, lerrors.Internal.Errorf("alias operands require an associative clause")
	}
	clause := s.clause
	var source *token
	if clause.subject != nil {
		source = clause.subject.noun
	} else if clause.obj != nil {
		source = clause.obj.noun
	}
	if source == nil || clause.phrase == nil || clause.phrase.cType != "alias" || clause.phrase.subject == nil || clause.phrase.subject.noun == nil || !clause.phrase.subject.noun.isDeclare() {
		return nil, nil, lerrors.Syntax.Errorf("associative declaration requires a source and alias target")
	}
	return source, clause.phrase.subject.noun, nil
}

func (s registrationSite) target() (*token, error) {
	if s.isAlias() {
		_, target, err := s.aliasOperands()
		return target, err
	}
	if s.subject != nil {
		return s.subject.noun, nil
	}
	return nil, nil
}

func (s registrationSite) valueObject() *objectGram {
	if s.predicate != nil {
		if s.predicate.obj != nil {
			return s.predicate.obj
		}
		if s.predicate.cType == "descriptive" && len(s.predicate.clauses) > 0 {
			return s.predicate.clauses[0].obj
		}
	}
	if s.subject != nil && s.subject.phrase != nil {
		return s.subject.phrase.obj
	}
	return nil
}

func (s registrationSite) isDescriptor() bool {
	if (s.predicate != nil && s.predicate.cType == "descriptive") || (s.clause != nil && strings.HasPrefix(s.clause.sType, "clause.descriptive.")) {
		return true
	}
	obj := s.valueObject()
	return obj != nil && obj.noun != nil && obj.noun.tType == "type.lexeme.adjective"
}

func (s registrationSite) literal() (*token, error) {
	if obj := s.valueObject(); obj != nil {
		return definitionLiteral(obj.tokens)
	}
	return nil, nil
}

func (s registrationSite) declaration() (extractedDeclaration, error) {
	if s.subject == nil || s.subject.noun == nil || !s.subject.noun.isDeclare() {
		return extractedDeclaration{}, lerrors.Internal.Errorf("declaration requires its original declaring subject")
	}
	if s.predicate != nil {
		return s.subject.passiveDeclaration(s.predicate)
	}
	extracted, err := s.subject.getDeclaration()
	if err != nil {
		return extractedDeclaration{}, err
	}
	lexical, _ := s.subject.isLexical()
	if lexical && extracted.Record.Value.Type == storageLexeme && s.clause != nil {
		extracted.Record.Value.Lexeme.POS = s.clause.cType
	}
	if err := s.subject.finishDeclaration(extracted.Record); err != nil {
		return extractedDeclaration{}, err
	}
	return extracted, nil
}

// Clauses nested in a value or directive still have their own modifier scope.
// The singular object field aliases objects[0]; visit each clause only once.
func nestedDeclarationClauses(subject *subjectGram, object *objectGram, objects []*objectGram, phrase *phraseGram) []*clauseGram {
	var clauses []*clauseGram
	seen := make(map[*clauseGram]bool)
	add := func(clause *clauseGram) {
		if clause != nil && !seen[clause] {
			seen[clause] = true
			clauses = append(clauses, clause)
		}
	}
	var visitPhrase func(*phraseGram)
	visitObject := func(object *objectGram) {
		if object == nil {
			return
		}
		add(object.clause)
		visitPhrase(object.phrase)
	}
	visitPhrase = func(phrase *phraseGram) {
		if phrase == nil {
			return
		}
		for _, clause := range phrase.clauses {
			add(clause)
		}
		if phrase.subject != nil {
			visitPhrase(phrase.subject.phrase)
		}
		visitObject(phrase.obj)
		for _, object := range phrase.objects {
			visitObject(object)
		}
		visitPhrase(phrase.phrase)
	}
	if subject != nil {
		visitPhrase(subject.phrase)
	}
	visitObject(object)
	for _, object := range objects {
		visitObject(object)
	}
	visitPhrase(phrase)
	return clauses
}

// walkRegistration visits ownership once, including definition bodies for
// visibility validation. Only executable grammar positions may register values.
func (sg *statementGram) walkRegistration(visit func(registrationSite, []*token, bool) error) error {
	if sg == nil {
		return lerrors.Internal.Errorf("registration requires statement grammar")
	}
	var walk func(registrationSite, []*token, []*clauseGram, []*conditionalGram, bool) error
	walk = func(site registrationSite, tokens []*token, clauses []*clauseGram, conditions []*conditionalGram, register bool) error {
		direct := append([]*clauseGram(nil), clauses...)
		var nested []*clauseGram
		if site.clause != nil {
			clause := site.clause
			direct = append(direct, clause.clauses...)
			nested = append(nested, nestedDeclarationClauses(nil, clause.obj, clause.objects, clause.phrase)...)
		}
		nested = append(nested, nestedDeclarationClauses(site.subject, nil, nil, nil)...)
		if site.predicate != nil {
			direct = append(direct, site.predicate.clauses...)
			conditions = append(conditions, site.predicate.conditionals...)
			nested = append(nested, nestedDeclarationClauses(nil, site.predicate.obj, nil, nil)...)
		}
		var supplemental []registrationSite
		if object := site.valueObject(); object != nil && object.cType == "callable" && object.phrase != nil {
			for _, parameter := range object.phrase.subjects {
				supplemental = append(supplemental, registrationSite{subject: parameter, parameter: true})
			}
		}
		if site.clause != nil && site.clause.verb != nil && site.clause.verb.binding != nil && site.clause.verb.binding.Record.isFunction() {
			for _, phrase := range site.clause.phrases {
				if phrase.gType == "phrase.results" {
					for _, result := range phrase.subjects {
						supplemental = append(supplemental, registrationSite{subject: result, resultOf: site.clause})
					}
				}
			}
		}
		own := tokens
		for _, child := range supplemental {
			own = selectTokens(child.subject.tokens).remaining(own)
		}
		for _, child := range append(append([]*clauseGram(nil), direct...), nested...) {
			own = selectTokens(child.tokens).remaining(own)
		}
		for _, conditional := range conditions {
			own = selectTokens(conditional.tokens).remaining(own)
		}
		if err := visit(site, own, register); err != nil {
			return err
		}
		for _, child := range supplemental {
			if err := visit(child, child.subject.tokens, register); err != nil {
				return err
			}
		}
		seen := make(map[*clauseGram]bool)
		for _, group := range []struct {
			clauses  []*clauseGram
			register bool
		}{{direct, register}, {nested, false}} {
			for _, child := range group.clauses {
				if seen[child] {
					continue
				}
				seen[child] = true
				if err := walk(registrationSite{subject: child.subject, predicate: child.predicate, clause: child}, child.tokens, nil, nil, group.register); err != nil {
					return err
				}
			}
		}
		for _, conditional := range conditions {
			if err := walk(registrationSite{}, conditional.tokens, conditional.clauses, nil, register); err != nil {
				return err
			}
		}
		return nil
	}
	return walk(registrationSite{subject: sg.subject, predicate: sg.predicate}, sg.tokens, sg.clauses, nil, true)
}

func (sg *statementGram) registrationSites() ([]registrationSite, error) {
	var sites []registrationSite
	err := sg.walkRegistration(func(site registrationSite, own []*token, register bool) error {
		if !register {
			return nil
		}
		target, err := site.target()
		if err != nil {
			return err
		}
		if target != nil && (target.isDeclare() || site.isDescriptor()) {
			sites = append(sites, site)
		}
		for _, tok := range own {
			if tok.isDeclare() && tok != target {
				return lerrors.Internal.Errorf("declaration token %q has no owning grammar site", tok.rawBytes)
			}
		}
		return nil
	})
	return sites, err
}
