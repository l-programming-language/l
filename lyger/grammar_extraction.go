package Lyger

import (
	"bytes"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// recordID builds an ID string for the subject if subject is declarative
func (sb *subjectGram) recordID() string {
	reID := ""
	if sb.article != nil {
		if sb.article.tType == "articles.variable.root" {
			reID += "*"
		}
	}
	for _, adj := range valueModifiers(sb.adjectives) {
		if reID == "" {
			reID += string(adj.rawBytes)
		} else {
			reID += "." + string(adj.rawBytes)
		}
	}
	if sb.noun != nil {
		implicitRoot := sb.noun.getImplicitRoot()
		if reID == "" {
			reID += strings.Replace(implicitRoot, `"`, ``, -1)
		} else {
			reID += "." + strings.Replace(implicitRoot, `"`, ``, -1)
		}
	}
	return reID
}

// getDeclareLabel returns a declaration label for a new variable, i.e. "Execute|-s|-d|--ing|--or" or "Objective"
func (sb *subjectGram) getDeclareLabel() (string, error) {
	if sb.noun != nil {
		if bytes.Contains(sb.noun.rawBytes, []byte(`"`)) {
			return string(sb.noun.rawBytes), nil
		}
	}
	return "", lerrors.Syntax.Errorf("%s", "subject noun value not declarative: "+string(sb.noun.rawBytes))
}

// isLexical determines whether the subject declaration or definition type is both lexical and/or determinate
func (sb *subjectGram) isLexical() (bool, bool) {
	if sb.cType == "type" || sb.cType == "list" || sb.cType == "dict" {
		return false, true
	} else if sb.cType == "definitional" {
		return true, true
	} else if sb.cType == "declarative" || sb.cType == "directive" || sb.cType == "utilization" {
		if sb.phrase != nil {
			return sb.phrase.isLexical()
		}
		return true, false
	}
	return false, false
}

// extractedDeclaration keeps the defining source token with its semantic record.
// Record IDs can acquire member/type metadata; token identity remains unchanged.
type extractedDeclaration struct {
	Token      *token
	Record     *record
	Visibility declarationVisibility
	Alias      *record // Set only by associative registration; committed with the binding.
}

// getDeclaration of a built subjectGram returns declarations with their source tokens or an error
//
//	i.e. "y" as it's Output
//	HINT: Output special keyword is always of primitive text
func (sb *subjectGram) getDeclaration() (extractedDeclaration, error) {
	if sb == nil || sb.noun == nil || sb.phrase == nil {
		return extractedDeclaration{}, lerrors.Syntax.Errorf("subject declaration requires a noun and value phrase")
	}
	lexical, _ := sb.isLexical()
	recEntry, valType, err := sb.phrase.getValue(lexical)
	if err != nil {
		return extractedDeclaration{}, err
	}
	return extractedDeclaration{Token: sb.noun, Record: &record{ID: sb.recordID(), Type: valType, Value: recEntry}}, nil
}

// referenceID builds an ID string for the object
func (sb *objectGram) referenceID() string {
	reID := ""
	if sb.article != nil {
		if sb.article.tType == "articles.variable.root" {
			reID += "*"
		}
	}
	for _, adj := range valueModifiers(sb.adjectives) {
		if reID == "" {
			reID += string(adj.rawBytes)
		} else {
			reID += "." + string(adj.rawBytes)
		}
	}
	if sb.noun != nil {
		implicitRoot := sb.noun.getImplicitRoot()
		if reID == "" {
			reID += strings.Replace(implicitRoot, `"`, ``, -1)
		} else {
			reID += "." + strings.Replace(implicitRoot, `"`, ``, -1)
		}
	}
	return reID
}

// getValue returns a record entry and valueType for the assignable value of the objectGram
func (sb *objectGram) getValue(lexical bool) (*entry, string, error) {
	if sb == nil {
		return nil, "", lerrors.Internal.Errorf("cannot extract value from a nil object")
	}
	if sb.cType == "definitive" && sb.phrase != nil && sb.phrase.cType == "definitional" {
		return sb.phrase.getValue(lexical)
	}
	if sb.noun == nil || sb.noun.mem == nil {
		return nil, "", lerrors.Internal.Errorf("object value requires a bound noun")
	}
	stored, err := sb.noun.mem.value()
	if err != nil {
		return nil, "", err
	}
	switch sb.cType {
	case "definitive":
		value := &entry{}
		value.bindFrom(sb.noun.mem)
		switch value.Type {
		case storageLexeme:
			value.Lexeme.Adjectives = append(value.Lexeme.Adjectives, valueModifiers(sb.adjectives)...)
			return value, value.Lexeme.getValueType(), nil
		case storagePrimitive, storageApplicative, storageIterative:
			return value, strings.ToLower(value.Type.String()) + "." + strings.ToLower(stored.Type.String()), nil
		}
	case "text":
		if stored.isDef() {
			lexEnt, err := initLexicalEntry([]byte(stored.Data.(string))) //nolint:errcheck // value() validates the tagged payload before this type-specific projection.
			if err != nil {
				return nil, "", err
			}
			return lexEnt, "lexeme.generic", nil
		}
		return sb.noun.mem, "primitive.text", nil
	case "number":
		return sb.noun.mem, "primitive.number", nil
	case "truth":
		return sb.noun.mem, "primitive.truth", nil
	case "null":
		return sb.noun.mem, "primitive.null", nil
	case "type":
		if sb.noun.mem.Type == storageLexeme {
			value := &entry{}
			value.bindFrom(sb.noun.mem)
			value.ID = ""
			value.Lexeme.POS = strings.ToLower(value.Lexeme.POS)
			value.Lexeme.Type = "generic"
			return value, value.Lexeme.getValueType(), nil
		}
		id := sb.noun.mem.ID
		if sb.article != nil && sb.article.tType == "articles.variable.root" {
			id = "*"
		}
		if id != "" {
			id += "."
		}
		id += strings.ToLower(stored.Type.String())
		if sb.phrase != nil {
			ent, entType, err := sb.phrase.getValue(lexical)
			if err != nil {
				return nil, "", err
			}
			ent.ID = id
			return ent, entType, nil
		}
		if !lexical {
			if typeStr, ok := getTypeMap()[stored.Type.String()]; ok {
				sb.noun.mem.ID = ""
				return sb.noun.mem, typeStr, nil
			}
		}
	case "list":
		return sb.noun.mem, "iterative.list", nil
	case "dict":
		return sb.noun.mem, "iterative.dict", nil
	}
	return nil, "", lerrors.Unsupported.Errorf("cannot extract value from object %q", sb.cType)
}

// isLexical determines whether the phraseGram declaration or definition type is both lexical and/or determinate
func (sb *phraseGram) isLexical() (bool, bool) {
	if sb.cType == "definitional" {
		return true, true
	}
	if sb.sType == "phrase.directive" {
		return true, true
	}
	// TODO LATER MAY NEED TO SUB GRAMMAR LEXICALITY CHECK
	return false, false
}

// getValue returns a record entry and valueType for the assignable value of the phraseGram
// Directive definitions retain ordered action uses; instruction planning later
// resolves their source-scoped references and definition instantiation renders
// literal templates without rewriting the stored prototypes.
func (sb *phraseGram) getValue(lexical bool) (*entry, string, error) {
	if sb == nil {
		return nil, "", lerrors.Internal.Errorf("cannot extract value from a nil phrase")
	}
	switch sb.cType {
	case "definitional":
		if lexical && sb.subject == nil && sb.obj != nil {
			return sb.obj.getValue(lexical)
		}
	case "value", "definitive":
		if sb.obj != nil {
			return sb.obj.getValue(lexical)
		}
		if len(sb.clauses) > 0 && sb.sType == "phrase.directive" {
			newEntry := initEntry("verb.generic", storageLexeme)
			var actions []*lexeme
			for _, c := range sb.clauses {
				act, err := c.getAction()
				if err != nil {
					return nil, "", err
				}
				actions = append(actions, act.Lexeme)
			}
			if err := newEntry.Lexeme.loadActions(actions); err != nil {
				return nil, "", err
			}
			return newEntry, "lexeme.verb.generic", nil
		}
	case "type":
		if sb.obj != nil {
			return sb.obj.getValue(lexical)
		}
	}
	return nil, "", lerrors.Unsupported.Errorf("cannot extract value from phrase %q", sb.cType)
}

// getAction returns an entry containing active Lexeme for use in Directive Phrases initializing complex Lexemes
func (sb *clauseGram) getAction() (*entry, error) {
	if sb == nil || sb.verb == nil || sb.verb.mem == nil || sb.verb.mem.Type != storageLexeme || sb.verb.mem.Lexeme == nil {
		return nil, lerrors.Internal.Errorf("directive action requires a bound lexeme")
	}
	if _, err := sb.verb.mem.value(); err != nil {
		return nil, err
	}
	actEnt := &entry{}
	actEnt.bindFrom(sb.verb.mem)
	actEnt.Lexeme.Adverbs = append(actEnt.Lexeme.Adverbs, valueModifiers(sb.adverbs)...)
	if sb.obj != nil {
		if err := actEnt.Lexeme.bindArguments([]string{sb.obj.referenceID()}); err != nil {
			return nil, err
		}
	}
	return actEnt, nil
}

// getValue returns a record entry and valueType for the assignable value of the clauseGram
func (sb *clauseGram) getValue() (*entry, string, error) {
	if sb == nil {
		return nil, "", lerrors.Internal.Errorf("cannot extract value from a nil clause")
	}
	if isDeclarativeGrammar(sb.sType) && sb.cType == "adjective" && sb.sType == "clause.descriptive.active" {
		if sb.obj == nil {
			return nil, "", lerrors.Internal.Errorf("descriptive clause requires an object value")
		}
		lexEntry, _, err := sb.obj.getValue(true)
		if err != nil {
			return nil, "", err
		}
		if lexEntry.Type != storageLexeme {
			return nil, "", lerrors.Internal.Errorf("descriptive clause requires a lexical value")
		}
		lexEntry.Lexeme.POS = "adjective"
		return lexEntry, lexEntry.Lexeme.getValueType(), nil
	}
	return nil, "", lerrors.Unsupported.Errorf("cannot extract value from clause %q (%s)", sb.cType, sb.sType)
}

func (sb *subjectGram) finishDeclaration(rec *record) error {
	label, err := sb.getDeclareLabel()
	if err != nil {
		return err
	}
	switch rec.Value.Type {
	case storageLexeme:
		if err := rec.Value.Lexeme.reLabel(label); err != nil {
			return err
		}
	case storagePrimitive, storageApplicative, storageIterative:
	default:
		return lerrors.Internal.Errorf("invalid type to update: %s", rec.Value.Type)
	}
	return rec.compileMeta()
}

func (sb *subjectGram) passiveDeclaration(predicate *predicateGram) (extractedDeclaration, error) {
	value, kind, err := predicate.getValue()
	if err != nil {
		return extractedDeclaration{}, err
	}
	if value == nil || kind == "" {
		return extractedDeclaration{}, lerrors.Internal.Errorf("cannot extract declaration %q from %s", sb.recordID(), predicate.sType)
	}
	rec := &record{ID: sb.recordID(), Type: kind, Value: value}
	if err := sb.finishDeclaration(rec); err != nil {
		return extractedDeclaration{}, err
	}
	return extractedDeclaration{Token: sb.noun, Record: rec}, nil
}

// getValue returns a record entry and valueType for the assignable value of the predicate
func (sb *predicateGram) getValue() (*entry, string, error) {
	if sb == nil {
		return nil, "", lerrors.Internal.Errorf("cannot extract value from a nil predicate")
	}
	switch sb.cType {
	case "assign":
		if sb.obj != nil {
			return sb.obj.getValue(false)
		}
	case "descriptive":
		if len(sb.clauses) > 0 {
			return sb.clauses[0].getValue()
		}
	}
	return nil, "", lerrors.Unsupported.Errorf("cannot extract value from predicate %q", sb.cType)
}

// getDeclarations retains each declaring node and its own visibility.
func (sg *statementGram) getDeclarations() ([]extractedDeclaration, error) {
	sites, err := sg.registrationSites()
	if err != nil {
		return nil, err
	}
	visibilities, err := sg.declarationVisibilities()
	if err != nil {
		return nil, err
	}
	var declarations []extractedDeclaration
	for _, site := range sites {
		extracted, err := site.declaration()
		if err != nil {
			return nil, err
		}
		extracted.Visibility = visibilities[extracted.Token]
		declarations = append(declarations, extracted)
	}
	return declarations, nil
}
