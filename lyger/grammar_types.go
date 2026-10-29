package Lyger

import (
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// Lexical fields are optional. Build their association set directly instead of
// sorting them into a temporary token slice and then constructing the same set.
func selectLexicalTokens(groups ...[]*token) tokenSelection {
	selected := selectTokens(groups...)
	delete(selected, nil)
	return selected
}

// grammeme is the base structure of any Grammar in L
type grammeme struct {
	ID        string            `json:"id,omitempty"`
	Structure string            `json:"structure,omitempty"`
	Cases     map[string]string `json:"cases,omitempty"`
	Category  string            `json:"category,omitempty"`
	Class     string            `json:"class,omitempty"`
}

// getCase returns either an error or a caseStr
func (g *grammeme) getCase(caseType string) (string, error) {
	if caseStr, ok := g.Cases[caseType]; ok {
		return caseStr, nil
	}
	return "", lerrors.InvalidConfiguration.Errorf("%s", "case missing for: "+caseType)
}

// getCasedStructure inputs a case type, then returns the structure string with case string sub
func (g *grammeme) getCasedStructure(caseType string) (string, error) {
	if caseType == "NONE" || caseType == "UNKNOWN" {
		return g.Structure, nil
	}
	caseStr, err := g.getCase(caseType)
	if err != nil {
		return g.Structure, err
	}
	return strings.Replace(g.Structure, "CASE", caseStr, -1), nil
}

// conditionGram holds data for the noun being acted in a statement
type conditionGram struct {
	sType        string
	gType        string
	cType        string
	rule         *grammeme
	tokens       []*token
	values       []*valueGram
	article      *token
	adverb       *token
	verb         *token
	prepositions []*token
	gram         *grammar
}

// getUnassociatedTokens recursively returns a slice of ordered token currently unassigned to sub-grammar
func (sb *conditionGram) getUnassociatedTokens() []*token {
	selected := selectLexicalTokens(sb.prepositions, []*token{sb.article, sb.adverb, sb.verb})
	for _, c := range sb.values {
		selected.add(c.tokens...)
	}
	return selected.remaining(sb.tokens)
}

// loadTyping parses sType into gType (grammar type) and cType (case type)
func (sb *conditionGram) loadTyping(sType string) error {
	return (grammarTyping{&sb.sType, &sb.gType, &sb.cType}).load(sType, familyCondition, grammarCaseInferred)
}

func (sb *conditionGram) gramSwitch(term grammarTerm) error {
	switch {
	case term.Kind == termGrammar && term.Family == familyValue:
		vGram, sErr := buildGrammarNode(initValueGram(term.Name, sb.getUnassociatedTokens(), sb.gram))
		if sErr != nil {
			return sErr
		}
		sb.values = append(sb.values, vGram)
	case term.Kind == termCase:
		if sb.cType == "UNKNOWN" {
			return sb.caseBuild()
		}
	default:
		return lerrors.InvalidConfiguration.Errorf("%s", "Unhandled CONDITION Grammar Type: "+term.className())
	}
	return nil
}

// lexSwitch inits associates tokens to conditionGram fields based on the decoded term and token order
func (sb *conditionGram) lexSwitch(term grammarTerm) error {
	// VERB.COMPARATIVE.ACTIVE and VERB.APPROPRIATIVE.ACTIVE retain their cases.
	var destination **token
	switch {
	case term.Kind == termLexical && term.POS == posAdverb:
		destination = &sb.adverb
	case term.Kind == termLexical && term.POS == posVerb:
		destination = &sb.verb
	case term.Kind == termLexical && term.POS == posArticle:
		destination = &sb.article
	case term.Kind == termLexical && term.POS == posPreposition:
		prep, err := term.lexicalToken(sb.getUnassociatedTokens(), lexicalMatchQualified)
		if err != nil {
			return err
		}
		if prep != nil {
			sb.prepositions = append(sb.prepositions, prep)
		}
		return nil
	default:
		return lerrors.InvalidConfiguration.Errorf("%s", "Unhandled Condition Lex Type: "+term.className())
	}
	matched, err := term.lexicalToken(sb.getUnassociatedTokens(), lexicalMatchQualified)
	if err != nil {
		return err
	}
	*destination = matched
	return nil
}

func (sb *conditionGram) buildProcess(casedStr string) error {
	return applyGrammarRule(sb, casedStr, (conditionGram).clone, (*conditionGram).gramSwitch, (*conditionGram).lexSwitch, nil)
}

func (sb *conditionGram) caseBuild() error {
	return selectGrammarCase(sb, sb.rule.Cases, (conditionGram).clone, (*conditionGram).buildProcess, (*conditionGram).getUnassociatedTokens, func(s *conditionGram, name string) { s.cType = name }, sb.tokens)
}

func (sb *conditionGram) build() error {
	return buildGrammarPrefix(sb.rule, sb.cType, &sb.tokens, sb.buildProcess, sb.getUnassociatedTokens)
}

func initConditionGram(sType string, tokens []*token, gram *grammar) (*conditionGram, error) {
	var r conditionGram
	if len(tokens) == 0 {
		return &r, lerrors.Syntax.Errorf("error: condition initialized with empty slice of tokens")
	}
	r.tokens = tokens
	if err := r.loadTyping(sType); err != nil {
		return &r, err
	}
	r.gram = gram
	g, err := r.gram.getGrammeme(r.gType)
	if err != nil {
		r.tokens = []*token{} // Unlike Predicate|Statement, clause tokens unused after build will be subtracted from slice
		return &r, err
	}
	r.rule = g
	return &r, nil
}

func initConditionGrams(sType string, tokens []*token, gram *grammar) ([]*conditionGram, []*token, error) {
	var r []*conditionGram
	if len(tokens) == 0 {
		return r, tokens, lerrors.Syntax.Errorf("error: conditions initialized with empty slice of tokens")
	}
	compoundSequences, conjuncts, err := identifyConditionSequences(tokens)
	if err != nil {
		return r, conjuncts, err
	}
	formatType := strings.Replace(strings.Replace(sType, "[", "", -1), "]", "", -1)
	for _, cs := range compoundSequences {
		initGram, initErr := initConditionGram(formatType, cs, gram)
		if initErr != nil {
			return r, conjuncts, initErr
		}
		r = append(r, initGram)
	}
	return r, conjuncts, nil
}

// conditionalGram holds data for the noun being acted in a statement
type conditionalGram struct {
	sType        string
	gType        string
	cType        string
	rule         *grammeme
	tokens       []*token
	evaluations  []*token // i.e. if / else / else, if / while / otherwise
	conditions   []*conditionGram
	conjunctions []*token // and / or
	adverb       *token   // i.e. "then"
	clause       *clauseGram
	gram         *grammar
	clauses      []*clauseGram
}

func (conditional *conditionalGram) orderedClauses() []*clauseGram {
	if len(conditional.clauses) == 0 && conditional.clause != nil {
		return []*clauseGram{conditional.clause}
	}
	return conditional.clauses
}

// getUnassociatedTokens recursively returns a slice of ordered token currently unassigned to sub-grammar
func (sb *conditionalGram) getUnassociatedTokens() []*token {
	selected := selectLexicalTokens(sb.evaluations, sb.conjunctions, []*token{sb.adverb})
	if sb.clause != nil {
		selected.add(sb.clause.tokens...)
	}
	for _, c := range sb.conditions {
		selected.add(c.tokens...)
	}
	for _, clause := range sb.clauses {
		selected.add(clause.tokens...)
	}
	return selected.remaining(sb.tokens)
}

// subTypeSwitch checks for a conditionalGram subType if one isn't provided in the sType
func (sb *conditionalGram) subTypeSwitch() string {
	if len(sb.tokens) > 0 && sb.tokens[0].getPOS(true) == "evaluations" {
		switch sb.tokens[0].tType {
		case "evaluations.conditional.root":
			return "if"
		case "evaluations.exceptional.root":
			if len(sb.tokens) > 1 && sb.tokens[0].tType == "evaluations.conditional.root" {
				return "else.if"
			}
			return "else.else"
		case "evaluations.alternative.root":
			return "other"
		case "evaluations.iterative.root":
			return "while"
		}
	}
	return ""
}

// loadTyping parses sType into gType (grammar type) and cType (case type)
func (sb *conditionalGram) loadTyping(sType string) error {
	typing := grammarTyping{&sb.sType, &sb.gType, &sb.cType}
	if err := typing.decode(sType); err != nil {
		return err
	}
	if !strings.Contains(sb.sType, ".") {
		subType := sb.subTypeSwitch()
		if subType == "" {
			return lerrors.Syntax.Errorf("conditional typing error: first token of a conditional must be a conjunctive")
		}
		sb.sType += "." + subType
	}
	typing.decompose(familyConditional, grammarCaseInferred)
	return nil
}

func (sb *conditionalGram) gramSwitch(term grammarTerm) error {
	switch {
	case term.Kind == termGrammar && term.Family == familyClause:
		clauses, separators, err := buildGrammarClauses(term, sb.getUnassociatedTokens(), sb.gram)
		if err != nil {
			return err
		}
		sb.clauses = append(sb.clauses, clauses...)
		if len(clauses) > 0 && sb.clause == nil {
			sb.clause = clauses[0]
		}
		sb.conjunctions = append(sb.conjunctions, separators...)

	case term.Kind == termGrammar && term.Family == familyCondition:
		var conditions []*conditionGram
		var conjunctions []*token
		if term.Multiple {
			var err error
			conditions, conjunctions, err = initConditionGrams(term.Name, sb.getUnassociatedTokens(), sb.gram)
			if err != nil {
				return err
			}
		} else {
			condition, err := initConditionGram(term.Name, sb.getUnassociatedTokens(), sb.gram)
			if err != nil {
				return err
			}
			conditions = append(conditions, condition)
		}
		for _, condition := range conditions {
			if err := condition.build(); err != nil {
				return err
			}
		}
		sb.conditions = append(sb.conditions, conditions...)
		sb.conjunctions = append(sb.conjunctions, conjunctions...)

	case term.Kind == termCase:
		if sb.cType == "UNKNOWN" {
			return sb.caseBuild()
		}
	default:
		return lerrors.InvalidConfiguration.Errorf("%s", "Unhandled CONDITIONAL Grammar Type: "+term.className())
	}
	return nil
}

// lexSwitch inits associates tokens to conditionalGram fields based on the decoded term and token order
func (sb *conditionalGram) lexSwitch(term grammarTerm) error {
	switch {
	case term.Kind == termLexical && term.POS == posEvaluation:
		return term.associateLexicalTokens(sb.getUnassociatedTokens(), &sb.evaluations, lexicalMatchClass)
	case term.Kind == termLexical && term.POS == posConjunction:
		return term.associateLexicalTokens(sb.getUnassociatedTokens(), &sb.conjunctions, lexicalMatchClass)
	case term.Kind == termLexical && term.POS == posAdverb:
		return term.associateLexicalToken(sb.getUnassociatedTokens(), &sb.adverb, lexicalMatchClass)
	default:
		return lerrors.InvalidConfiguration.Errorf("%s", "Unhandled CONDITIONAL Lex Type: "+term.className())
	}
}

func (sb *conditionalGram) buildProcess(casedStr string) error {
	return applyGrammarRule(sb, casedStr, (conditionalGram).clone, (*conditionalGram).gramSwitch, (*conditionalGram).lexSwitch, nil)
}

func (sb *conditionalGram) caseBuild() error {
	return selectGrammarCase(sb, sb.rule.Cases, (conditionalGram).clone, (*conditionalGram).buildProcess, (*conditionalGram).getUnassociatedTokens, func(s *conditionalGram, name string) { s.cType = name }, sb.tokens)
}

func (sb *conditionalGram) build() error {
	return buildGrammarPrefix(sb.rule, sb.cType, &sb.tokens, sb.buildProcess, sb.getUnassociatedTokens)
}

// trimTokens down to if/else & while/then
func (sb *conditionalGram) trimTokens() error {
	if len(sb.tokens) > 0 {
		var condType string
		if sb.tokens[0].tType == "evaluations.conditional.root" {
			condType = "if"
		} else if sb.tokens[0].tType == "evaluations.exceptional.root" {
			condType = "else"
			if len(sb.tokens) > 1 && sb.tokens[0].tType == "evaluations.conditional.root" {
				condType += ".if"
			} else {
				condType += ".else"
			}
		} else if sb.tokens[0].tType == "evaluations.iterative.root" {
			condType = "while"
		} else {
			return nil
		}
		var condTokens []*token
		for ind, inToken := range sb.tokens {
			if condType == "if" {
				if inToken.tType != "evaluations.exceptional.root" {
					condTokens = append(condTokens, inToken)
				} else {
					break
				}
			} else if condType == "else.if" {
				if ind == 0 || inToken.tType != "evaluations.exceptional.root" {
					condTokens = append(condTokens, inToken)
				} else {
					break
				}
			} // TODO LATER - ADD else if BLOCK FOR trimming input tokens for "WHILE / THEN", UPDATE GRAMMAR CLAUSE
		}
		if len(condTokens) > 0 {
			sb.tokens = condTokens
		}
		return nil
	}
	return lerrors.Syntax.Errorf("conditional must be initialized with a set of tokens")
}

func initConditionalGram(sType string, tokens []*token, gram *grammar) (*conditionalGram, error) {
	var r conditionalGram
	if len(tokens) == 0 {
		return &r, lerrors.Syntax.Errorf("error: conditional initialized with empty slice of tokens")
	}
	r.tokens = tokens
	tErr := r.trimTokens()
	if tErr != nil {
		r.tokens = []*token{}
		return &r, tErr
	}
	tErr = r.loadTyping(sType)
	if tErr != nil {
		r.tokens = []*token{} // Unlike Predicate|Statement, clause tokens unused after build will be subtracted from slice
		return &r, tErr
	}
	r.gram = gram
	g, err := r.gram.getGrammeme(r.gType)
	if err != nil {
		r.tokens = []*token{} // Unlike Predicate|Statement, clause tokens unused after build will be subtracted from slice
		return &r, err
	}
	r.rule = g
	return &r, nil
}

func initConditionalGrams(sType string, tokens []*token, gram *grammar) ([]*conditionalGram, error) {
	var r []*conditionalGram
	if len(tokens) == 0 {
		return r, lerrors.Syntax.Errorf("error: conditionals initialized with empty slice of tokens")
	}
	compoundSequences, err := identifyConditionalSequences(tokens)
	if err != nil {
		return r, err
	}
	formatType := strings.Replace(strings.Replace(sType, "[", "", -1), "]", "", -1)
	for _, cs := range compoundSequences {
		initGram, initErr := initConditionalGram(formatType, cs, gram)
		if initErr != nil {
			return r, initErr
		}
		r = append(r, initGram)
	}
	return r, nil
}

// valueGram holds data for the primitive noun being acted in a statement
type valueGram struct {
	sType      string
	gType      string
	cType      string // case type if case
	rule       *grammeme
	tokens     []*token
	article    *token
	adjectives []*token
	noun       *token
	gram       *grammar
}

// loadTyping parses sType into gType (grammar type) and cType (case type)
func (sb *valueGram) loadTyping(sType string) error {
	return (grammarTyping{&sb.sType, &sb.gType, &sb.cType}).load(sType, familyValue, grammarCaseInferred)
}

// getUnassociatedTokens recursively returns a slice of ordered token currently unassigned to sub-grammar
func (sb *valueGram) getUnassociatedTokens() []*token {
	return selectLexicalTokens(sb.adjectives, []*token{sb.article, sb.noun}).remaining(sb.tokens)
}

// typeSwitch inits associates tokens to valueGram fields based on the decoded type pattern and token order
func (sb *valueGram) typeSwitch(term grammarTerm) error {
	available := sb.getUnassociatedTokens()
	if len(available) > 0 && strings.HasPrefix(term.Pattern.TokenType(), "primitive.") {
		raw := available[0].rawBytes
		literal := isText(raw) || isTruth(raw) || isNum(raw) || string(raw) == "null"
		if !literal {
			return lerrors.Syntax.Errorf("%q is a reference, not a literal", raw)
		}
	}
	switch term.Pattern {
	case patternType, patternList, patternDictionary, patternText, patternNumber, patternTruth, patternNull:
		noun, err := term.typeToken(sb.getUnassociatedTokens())
		if err != nil {
			return err
		}
		if noun != nil {
			sb.noun = noun
		}
	default:
		return lerrors.InvalidConfiguration.Errorf("%s", "Unhandled VALUE Type Error: "+term.Pattern.TokenType())
	}
	return nil
}

func (sb *valueGram) gramSwitch(term grammarTerm) error {
	switch {
	case term.Kind == termCase:
		if sb.cType == "UNKNOWN" {
			return sb.caseBuild()
		}
	default:
		return lerrors.InvalidConfiguration.Errorf("%s", "Unhandled VALUE Grammar Type: "+term.className())
	}
	return nil
}

// lexSwitch inits associates tokens to subjectGram fields based on the decoded term and token order
func (sb *valueGram) lexSwitch(term grammarTerm) error {
	switch {
	case term.Kind == termLexical && term.POS == posAdjective:
		return term.associateLexicalTokens(sb.getUnassociatedTokens(), &sb.adjectives, lexicalMatchClass)
	case term.Kind == termLexical && term.POS == posArticle:
		return term.associateLexicalToken(sb.getUnassociatedTokens(), &sb.article, lexicalMatchClass)
	case term.Kind == termIdentifier:
		return term.associateLexicalToken(sb.getUnassociatedTokens(), &sb.noun, lexicalMatchClass)
	default:
		return lerrors.InvalidConfiguration.Errorf("%s", "Unhandled VALUE Lex Type: "+term.className())
	}
}

func (sb *valueGram) buildProcess(casedStr string) error {
	return applyGrammarRule(sb, casedStr, (valueGram).clone, (*valueGram).gramSwitch, (*valueGram).lexSwitch, (*valueGram).typeSwitch)
}

func (sb *valueGram) caseBuild() error {
	return selectGrammarCase(sb, sb.rule.Cases, (valueGram).clone, (*valueGram).buildProcess, (*valueGram).getUnassociatedTokens, func(s *valueGram, name string) { s.cType = name }, sb.tokens)
}

func (sb *valueGram) build() error {
	return buildGrammarPrefix(sb.rule, sb.cType, &sb.tokens, sb.buildProcess, sb.getUnassociatedTokens)
}

func initValueGram(gType string, tokens []*token, gram *grammar) (*valueGram, error) {
	var r valueGram
	if len(tokens) == 0 {
		return &r, lerrors.Syntax.Errorf("error: value initialized with empty slice of tokens")
	}
	if err := r.loadTyping(gType); err != nil {
		return &r, err
	}
	r.gram = gram
	g, err := r.gram.getGrammeme(r.gType)
	if err != nil {
		return &r, err
	}
	r.rule = g
	r.tokens = tokens
	return &r, nil
}

// subjectGram holds data for the noun being acted in a statement
type subjectGram struct {
	sType      string
	gType      string
	cType      string // case type if case
	rule       *grammeme
	tokens     []*token
	article    *token
	adjectives []*token
	noun       *token
	phrase     *phraseGram
	gram       *grammar
}

// loadTyping parses sType into gType (grammar type) and cType (case type)
func (sb *subjectGram) loadTyping(sType string) error {
	return (grammarTyping{&sb.sType, &sb.gType, &sb.cType}).load(sType, familySubject, grammarCaseInferred)
}

// getLexTokens returns a slice of tokens used by the lex fields of subjectGram
func (sb *subjectGram) getLexTokens() []*token {
	return selectTokens(sb.adjectives, []*token{sb.article, sb.noun}).ordered(sb.tokens)
}

// getUnassociatedTokens recursively returns a slice of ordered token currently unassigned to sub-grammar
func (sb *subjectGram) getUnassociatedTokens() []*token {
	selected := selectLexicalTokens(sb.adjectives, []*token{sb.article, sb.noun})
	if sb.phrase != nil {
		selected.add(sb.phrase.tokens...)
	}
	return selected.remaining(sb.tokens)
}

// getAssociatedTokens recursively returns a slice of ordered token currently unassigned to sub-grammar
func (sb *subjectGram) getAssociatedTokens(update bool) []*token {
	selected := selectLexicalTokens(sb.adjectives, []*token{sb.article, sb.noun})
	if sb.phrase != nil {
		selected.add(sb.phrase.tokens...)
	}
	associatedTokens := selected.firstRun(sb.tokens)
	if update {
		sb.tokens = associatedTokens
	}
	return associatedTokens
}

func (sb *subjectGram) gramSwitch(term grammarTerm) error {
	switch {
	case term.Kind == termGrammar && term.Family == familyPhrase:
		return buildGrammarChild(&sb.phrase, initPhraseGram, term.Name, sb.getUnassociatedTokens(), sb.gram)
	case term.Kind == termCase:
		if sb.cType == "UNKNOWN" {
			return sb.caseBuild()
		}
	default:
		return lerrors.InvalidConfiguration.Errorf("%s", "Unhandled Subject Grammar Type: "+term.className())
	}
	return nil
}

// lexSwitch inits associates tokens to subjectGram fields based on the decoded term and token order
func (sb *subjectGram) lexSwitch(term grammarTerm) error {
	switch {
	case term.Kind == termLexical && term.POS == posAdjective:
		available := sb.getUnassociatedTokens()
		// A definitive subject names a value. A bound descriptor immediately
		// before the following phrase (or the end) is that reference, not a
		// modifier with a missing noun. Keep its lexical identity on the token.
		if sb.cType == "definitive" {
			for i, tok := range available {
				if tok.getPOS(true) != "adjective" || tok.isPossessive() {
					break
				}
				if tok.binding != nil && (i+1 == len(available) || available[i+1].getPOS(true) == "preposition" || available[i+1].getPOS(true) == "conjunction") {
					available = available[:i]
					break
				}
			}
		}
		return term.associateLexicalTokens(available, &sb.adjectives, lexicalMatchClass)
	case term.Kind == termLexical && term.POS == posArticle:
		return term.associateLexicalToken(sb.getUnassociatedTokens(), &sb.article, lexicalMatchClass)
	case term.Kind == termIdentifier:
		return term.associateLexicalToken(sb.getUnassociatedTokens(), &sb.noun, lexicalMatchClass)
	default:
		return lerrors.InvalidConfiguration.Errorf("%s", "Unhandled Subject Lex Type: "+term.className())
	}
}

// typeSwitch associates a native type used as a definite subject reference.
func (sb *subjectGram) typeSwitch(term grammarTerm) error {
	if term.Pattern != patternType {
		return lerrors.InvalidConfiguration.Errorf("unsupported subject type %s", term.Pattern.TokenType())
	}
	if sb.article == nil || sb.article.tType != "articles.definite.root" {
		return lerrors.Syntax.Errorf("a native type subject reference requires a definite article")
	}
	noun, err := term.typeToken(sb.getUnassociatedTokens())
	if err != nil {
		return err
	}
	sb.noun = noun
	return nil
}

func (sb *subjectGram) buildProcess(casedStr string) error {
	return applyGrammarRule(sb, casedStr, (subjectGram).clone, (*subjectGram).gramSwitch, (*subjectGram).lexSwitch, (*subjectGram).typeSwitch)
}

func (sb *subjectGram) caseBuild() error {
	return selectGrammarCase(sb, inferredSignatureCases(sb.rule.Cases), (subjectGram).clone, (*subjectGram).buildProcess, (*subjectGram).getUnassociatedTokens, func(s *subjectGram, name string) { s.cType = name }, sb.tokens)
}

func (sb *subjectGram) build() error {
	return buildGrammarPrefix(sb.rule, sb.cType, &sb.tokens, sb.buildProcess, sb.getUnassociatedTokens)
}

func initSubjectGram(gType string, tokens []*token, gram *grammar) (*subjectGram, error) {
	var r subjectGram
	if len(tokens) == 0 {
		return &r, lerrors.Syntax.Errorf("error: subject initialized with empty slice of tokens")
	}
	if err := r.loadTyping(gType); err != nil {
		return &r, err
	}
	r.gram = gram
	g, err := r.gram.getGrammeme(r.gType)
	if err != nil {
		return &r, err
	}
	r.rule = g
	r.tokens = tokens
	return &r, nil
}

// objectGram holds data for a non-subject noun in a statement
type objectGram struct {
	sType      string
	gType      string
	cType      string
	rule       *grammeme
	tokens     []*token
	article    *token
	adjectives []*token
	noun       *token
	clause     *clauseGram
	phrase     *phraseGram
	gram       *grammar
	name       *valueGram
}

// loadTyping parses sType into gType (grammar type) and cType (case type)
func (sb *objectGram) loadTyping(sType string) error {
	return (grammarTyping{&sb.sType, &sb.gType, &sb.cType}).load(sType, familyObject, grammarCaseInferred)
}

// getUnassociatedTokens recursively returns a slice of ordered token currently unassigned to sub-grammar
func (sb *objectGram) getUnassociatedTokens() []*token {
	selected := selectLexicalTokens(sb.adjectives, []*token{sb.article, sb.noun})
	if sb.clause != nil {
		selected.add(sb.clause.tokens...)
	}
	if sb.phrase != nil {
		selected.add(sb.phrase.tokens...)
	}
	if sb.name != nil {
		selected.add(sb.name.tokens...)
	}
	return selected.remaining(sb.tokens)
}

func (sb *objectGram) gramSwitch(term grammarTerm) error {
	switch {
	case term.Kind == termGrammar && term.Family == familyValue:
		return buildGrammarChild(&sb.name, initValueGram, term.Name, sb.getUnassociatedTokens(), sb.gram)
	case term.Kind == termGrammar && term.Family == familyClause:
		return buildGrammarChild(&sb.clause, initClauseGram, term.Name, sb.getUnassociatedTokens(), sb.gram)
	case term.Kind == termGrammar && term.Family == familyPhrase:
		return buildGrammarChild(&sb.phrase, initPhraseGram, term.Name, sb.getUnassociatedTokens(), sb.gram)
	case term.Kind == termCase:
		if sb.cType == "UNKNOWN" {
			return sb.caseBuild()
		}
	default:
		return lerrors.InvalidConfiguration.Errorf("%s", "Unhandled Object Grammar Type: "+term.className())
	}
	return nil
}

// lexSwitch inits associates tokens to objectGram fields based on the decoded term and token order
func (sb *objectGram) lexSwitch(term grammarTerm) error {
	switch {
	case term.Kind == termLexical && term.POS == posAdjective:
		return term.associateLexicalTokens(sb.getUnassociatedTokens(), &sb.adjectives, lexicalMatchClass)
	case term.Kind == termLexical && term.POS == posArticle:
		return term.associateLexicalToken(sb.getUnassociatedTokens(), &sb.article, lexicalMatchClass)
	case term.Kind == termIdentifier:
		// A callable keeps the same optional pre-nominal name as other named
		// objects, while its enclosing rule still owns the explicit signature.
		available := sb.getUnassociatedTokens()
		if sb.cType == "callable" && len(available) > 0 && isText(available[0].rawBytes) {
			if err := buildGrammarChild(&sb.name, initValueGram, "value.text", available, sb.gram); err != nil {
				return err
			}
		}
		return term.associateLexicalToken(sb.getUnassociatedTokens(), &sb.noun, lexicalMatchClass)
	default:
		return lerrors.InvalidConfiguration.Errorf("%s", "Unhandled Object Lex Type: "+term.className())
	}
}

// typeSwitch inits associates tokens to objectGram fields based on the decoded type pattern and token order
func (sb *objectGram) typeSwitch(term grammarTerm) error {
	available := sb.getUnassociatedTokens()
	if len(available) > 0 && strings.HasPrefix(term.Pattern.TokenType(), "primitive.") {
		raw := available[0].rawBytes
		literal := isText(raw) || isTruth(raw) || isNum(raw) || string(raw) == "null"
		if !literal {
			return lerrors.Syntax.Errorf("%q is a reference, not a literal", raw)
		}
	}
	switch term.Pattern {
	case patternType, patternList, patternDictionary, patternText, patternNumber, patternTruth, patternNull:
	default:
		return lerrors.InvalidConfiguration.Errorf("%s", "Unhandled Object Type Error: "+term.Pattern.TokenType())
	}
	noun, err := term.typeToken(sb.getUnassociatedTokens())
	if err != nil || noun == nil {
		return err
	}
	if (term.Pattern == patternList || term.Pattern == patternDictionary) && noun.isRawType() {
		if !term.Optional {
			return lerrors.Syntax.Errorf("%s", "invalid instantiated iterative name: "+string(noun.rawBytes))
		}
		return nil
	}
	sb.noun = noun
	return nil
}

func (sb *objectGram) buildProcess(casedStr string) error {
	return applyGrammarRule(sb, casedStr, (objectGram).clone, (*objectGram).gramSwitch, (*objectGram).lexSwitch, (*objectGram).typeSwitch)
}

func (sb *objectGram) caseBuild() error {
	return selectGrammarCase(sb, withoutGrammarCase(sb.rule.Cases, "callable", "argument", "type_argument"), (objectGram).clone, (*objectGram).buildProcess, (*objectGram).getUnassociatedTokens, func(s *objectGram, name string) { s.cType = name }, sb.tokens)
}

func (sb *objectGram) build() error {
	return buildGrammarPrefix(sb.rule, sb.cType, &sb.tokens, sb.buildProcess, sb.getUnassociatedTokens)
}

func initObjectGram(gType string, tokens []*token, gram *grammar) (*objectGram, error) {
	var r objectGram
	if len(tokens) == 0 {
		return &r, lerrors.Syntax.Errorf("error: object initialized with empty slice of tokens")
	}
	if err := r.loadTyping(gType); err != nil {
		return &r, err
	}
	r.gram = gram
	g, err := r.gram.getGrammeme(r.gType)
	if err != nil {
		return &r, err
	}
	r.rule = g
	r.tokens = tokens
	return &r, nil
}

// phraseGram holds data for the noun being acted in a statement
type phraseGram struct {
	sType        string
	gType        string
	cType        string
	rule         *grammeme
	tokens       []*token
	prepositions []*token
	subject      *subjectGram
	obj          *objectGram
	phrase       *phraseGram
	clauses      []*clauseGram
	gram         *grammar
	conjunctions []*token

	objects  []*objectGram
	values   []*valueGram
	subjects []*subjectGram
	markers  []*token
	phrases  []*phraseGram
}

func (phrase *phraseGram) orderedObjects() []*objectGram {
	if phrase == nil {
		return nil
	}
	if len(phrase.objects) > 0 {
		return phrase.objects
	}
	if phrase.obj != nil {
		return []*objectGram{phrase.obj}
	}
	return nil
}

// subTypeSwitch checks for a phraseGram subType if one isn't provided in the sType
func (sb *phraseGram) subTypeSwitch() string {
	if len(sb.tokens) > 0 && sb.tokens[0].getPOS(true) == "preposition" {
		chkVal, _ := generalizePrepTokenType(sb.tokens[0].tType)
		switch chkVal {
		case "prepositions.conjoinment.objective":
			return "objective"
		case "prepositions.definement.declarative":
			return "declarative"
		case "prepositions.conjoinment.directive":
			return "directive"
		default:
			return "generic"
		}
	}
	return ""
}

// loadTyping parses sType into gType (grammar type) and cType (case type)
func (sb *phraseGram) loadTyping(sType string) error {
	typing := grammarTyping{&sb.sType, &sb.gType, &sb.cType}
	if err := typing.decode(sType); err != nil {
		return err
	}
	if !strings.Contains(sb.sType, ".") {
		subType := sb.subTypeSwitch()
		if subType == "" {
			return lerrors.Syntax.Errorf("phrase typing error: first token of a phrase must be a preposition")
		}
		sb.sType += "." + subType
	}
	typing.decompose(familyPhrase, grammarCaseInferred)
	return nil
}

// getUnassociatedTokens recursively returns a slice of ordered token currently unassigned to sub-grammar
func (sb *phraseGram) getUnassociatedTokens() []*token {
	selected := selectLexicalTokens(sb.prepositions, sb.conjunctions)
	if sb.subject != nil {
		selected.add(sb.subject.tokens...)
	}
	if sb.obj != nil {
		selected.add(sb.obj.tokens...)
	}
	if sb.phrase != nil {
		selected.add(sb.phrase.tokens...)
	}
	for _, c := range sb.clauses {
		selected.add(c.tokens...)
	}
	for _, object := range sb.objects {
		selected.add(object.tokens...)
	}
	for _, subject := range sb.subjects {
		selected.add(subject.tokens...)
	}
	for _, value := range sb.values {
		selected.add(value.tokens...)
	}
	for _, phrase := range sb.phrases {
		selected.add(phrase.tokens...)
	}
	selected.add(sb.markers...)
	return selected.remaining(sb.tokens)
}

func (sb *phraseGram) gramSwitch(term grammarTerm) error {
	switch {
	case term.Kind == termGrammar && term.Family == familyValue:
		values, separators, err := buildGrammarValues(term, sb.getUnassociatedTokens(), sb.gram)
		if err != nil {
			return err
		}
		sb.values = append(sb.values, values...)
		sb.conjunctions = append(sb.conjunctions, separators...)
	case term.Kind == termGrammar && term.Family == familySubject:
		subjects, separators, err := grammarSequence(sb.getUnassociatedTokens(), term.Multiple, func(input []*token) (*subjectGram, []*token, error) {
			if sb.gType == "phrase.signature" {
				input = signatureTypePrefix(input)
			}
			subject, err := buildGrammarNode(initSubjectGram(term.Name, input, sb.gram))
			if err != nil {
				return nil, nil, err
			}
			return subject, subject.tokens, nil
		})
		if err != nil {
			return err
		}
		sb.subjects = append(sb.subjects, subjects...)
		if len(subjects) > 0 && sb.subject == nil {
			sb.subject = subjects[0]
		}
		sb.conjunctions = append(sb.conjunctions, separators...)

	case term.Kind == termGrammar && term.Family == familyObject:
		var objects []*objectGram
		var conjunctions []*token
		var err error
		if sb.gType == "phrase.signature" {
			objects, conjunctions, err = grammarSequence(sb.getUnassociatedTokens(), term.Multiple, func(input []*token) (*objectGram, []*token, error) {
				object, err := buildGrammarNode(initObjectGram(term.Name, signatureTypePrefix(input), sb.gram))
				if err != nil {
					return nil, nil, err
				}
				return object, object.tokens, nil
			})
		} else {
			objects, conjunctions, err = buildGrammarObjects(term, sb.getUnassociatedTokens(), sb.gram)
		}
		if err != nil {
			return err
		}
		sb.objects = append(sb.objects, objects...)
		if len(objects) > 0 && sb.obj == nil {
			sb.obj = objects[0]
		}
		sb.conjunctions = append(sb.conjunctions, conjunctions...)

	case term.Kind == termGrammar && term.Family == familyPhrase:
		return buildGrammarChild(&sb.phrase, initPhraseGram, term.Name, sb.getUnassociatedTokens(), sb.gram)
	case term.Kind == termGrammar && term.Family == familyClause:
		clauses, conjunctions, err := buildGrammarClauses(term, sb.getUnassociatedTokens(), sb.gram)
		if err != nil {
			return err
		}
		sb.clauses = append(sb.clauses, clauses...)
		sb.conjunctions = append(sb.conjunctions, conjunctions...)

	case term.Kind == termCase:
		if sb.cType == "UNKNOWN" {
			return sb.caseBuild()
		}
	default:
		return lerrors.InvalidConfiguration.Errorf("%s", "Unhandled Type: "+term.className())
	}
	return nil
}

// lexSwitch inits associates tokens to phraseGram fields based on the decoded term and token order
func (sb *phraseGram) lexSwitch(term grammarTerm) error {
	switch {
	case term.Kind == termLexical && (term.POS == posNoun || term.POS == posArticle || term.POS == posConjunction):
		available := sb.getUnassociatedTokens()
		if len(available) == 0 || !available[0].hasRole(term.POS.String(), term.Qualifier) {
			return lerrors.Syntax.Errorf("expected %s", term.Name)
		}
		sb.markers = append(sb.markers, available[0])
		return nil
	case term.Kind == termLexical && term.POS == posPreposition:
		return term.associateLexicalTokens(sb.getUnassociatedTokens(), &sb.prepositions, lexicalMatchQualified)
	default:
		return lerrors.InvalidConfiguration.Errorf("%s", "Unhandled Type: "+term.className())
	}
}

func (sb *phraseGram) buildProcess(casedStr string) error {
	return applyGrammarRule(sb, casedStr, (phraseGram).clone, (*phraseGram).gramSwitch, (*phraseGram).lexSwitch, nil)
}

func (sb *phraseGram) caseBuild() error {
	return selectGrammarCase(sb, inferredSignatureCases(sb.rule.Cases), (phraseGram).clone, (*phraseGram).buildProcess, (*phraseGram).getUnassociatedTokens, func(s *phraseGram, name string) { s.cType = name }, sb.tokens)
}

func (sb *phraseGram) build() error {
	return buildGrammarPrefix(sb.rule, sb.cType, &sb.tokens, sb.buildProcess, sb.getUnassociatedTokens)
}

func initPhraseGram(sType string, tokens []*token, gram *grammar) (*phraseGram, error) {
	var r phraseGram
	if len(tokens) == 0 {
		return &r, lerrors.Syntax.Errorf("error: phrase initialized with empty slice of tokens")
	}
	r.tokens = tokens
	err := r.loadTyping(sType)
	if err != nil {
		return &r, err
	}
	r.gram = gram
	g, err := r.gram.getGrammeme(r.gType)
	if err != nil {
		return &r, err
	}
	r.rule = g
	return &r, nil
}

// clauseGram holds data for the noun being acted in a statement
// HINT: ALL CLAUSES START WITH [ADVERB]* | VERB ...
type clauseGram struct {
	sType        string
	gType        string
	cType        string
	rule         *grammeme
	tokens       []*token
	adverbs      []*token
	prepositions []*token // i.e. in clause.receptive.active
	verb         *token
	obj          *objectGram  // i.e. in clause.declarative.active
	subject      *subjectGram // i.e. in clause.definitive.applicative
	predicate    *predicateGram
	phrase       *phraseGram
	clauses      []*clauseGram
	gram         *grammar
	conjunctions []*token

	objects  []*objectGram
	values   []*valueGram
	subjects []*subjectGram
	markers  []*token
	phrases  []*phraseGram
}

func (clause *clauseGram) orderedObjects() []*objectGram {
	if clause == nil {
		return nil
	}
	if len(clause.objects) > 0 {
		return clause.objects
	}
	if clause.obj != nil {
		return []*objectGram{clause.obj}
	}
	return nil
}

// subTypeSwitch checks for a clauseGram subType if one isn't provided in the sType
func (sb *clauseGram) subTypeSwitch() string {
	var primeVerb *token
	for _, t := range sb.tokens {
		curType := t.getPOS(true)
		if curType != "adverb" && curType != "verb" {
			return ""
		}
		if curType == "verb" {
			primeVerb = t
			break
		}
	}
	if primeVerb != nil {
		sb.verb = primeVerb
		return strings.Replace(primeVerb.tType, "verbs.", "", 1)
	}
	return ""
}

// loadTyping parses sType into gType (grammar type) and cType (case type)
func (sb *clauseGram) loadTyping(sType string) error {
	typing := grammarTyping{&sb.sType, &sb.gType, &sb.cType}
	if err := typing.decode(sType); err != nil {
		return err
	}
	splitTypes := strings.Split(sb.sType, ".")
	if len(splitTypes) < 3 {
		subType := sb.subTypeSwitch()
		if subType == "" {
			return lerrors.Syntax.Errorf("clause typing error: first tokens of a clause must be an adverb or verb")
		}
		if len(splitTypes) == 2 {
			caseMatch := false
			// eVerb; i.e. applicative, imperative, etc.
			eVerb := splitTypes[1]
			for _, sStr := range strings.Split(subType, ".") {
				if strings.ToLower(eVerb) == strings.ToLower(sStr) { //nolint:staticcheck // SA6005: preserve lowercase identity for Unicode grammar names; EqualFold equates additional names.
					caseMatch = true
					break
				} else if strings.ToLower(eVerb) == "imperative" && strings.ToLower(sStr) == "root" {
					caseMatch = true
					subType = strings.Replace(subType, ".root", ".imperative", 1)
					break
				}
			}
			if !caseMatch {
				return lerrors.Syntax.Errorf("%s", "clause typing error: expected "+subType+" to be of type "+eVerb)
			}
			sb.sType = strings.Replace(sb.sType, "."+eVerb, "", 1) // [clause.imperative] to [clause]
		}
		sb.sType += "." + subType
	}
	if typing.decompose(familyClause, grammarCaseInferred) {
		return nil
	}
	return lerrors.Syntax.Errorf("clause typing error: verb type incorrect")
}

// getUnassociatedTokens recursively returns a slice of ordered token currently unassigned to sub-grammar
func (sb *clauseGram) getUnassociatedTokens() []*token {
	selected := selectLexicalTokens(sb.adverbs, sb.prepositions, []*token{sb.verb}, sb.conjunctions)
	if sb.subject != nil {
		selected.add(sb.subject.tokens...)
	}
	if sb.predicate != nil {
		selected.add(sb.predicate.tokens...)
	}
	if sb.obj != nil {
		selected.add(sb.obj.tokens...)
	}
	if sb.phrase != nil {
		selected.add(sb.phrase.tokens...)
	}
	for _, c := range sb.clauses {
		selected.add(c.tokens...)
	}
	for _, object := range sb.objects {
		selected.add(object.tokens...)
	}
	for _, subject := range sb.subjects {
		selected.add(subject.tokens...)
	}
	for _, value := range sb.values {
		selected.add(value.tokens...)
	}
	for _, phrase := range sb.phrases {
		selected.add(phrase.tokens...)
	}
	selected.add(sb.markers...)
	return selected.remaining(sb.tokens)
}

func (sb *clauseGram) gramSwitch(term grammarTerm) error {
	switch {
	case term.Kind == termGrammar && term.Family == familyValue:
		values, separators, err := buildGrammarValues(term, sb.getUnassociatedTokens(), sb.gram)
		if err != nil {
			return err
		}
		sb.values = append(sb.values, values...)
		sb.conjunctions = append(sb.conjunctions, separators...)
	case term.Kind == termGrammar && term.Family == familyPredicate:
		return buildGrammarChild(&sb.predicate, initPredicateGram, term.Name, sb.getUnassociatedTokens(), sb.gram)
	case term.Kind == termGrammar && term.Family == familySubject:
		return buildGrammarChild(&sb.subject, initSubjectGram, term.Name, sb.getUnassociatedTokens(), sb.gram)
	case term.Kind == termGrammar && term.Family == familyObject:
		// Unary native roots accept coordinated targets. The configured OBJECT
		// still names the operand grammar; each coordinated member is parsed as
		// that same object and kept in source order for instruction lowering.
		if sb.verb != nil {
			role, _ := sb.verb.nativeRole()
			kind, _ := role.InstructionKind()
			switch kind {
			case instructionBuild, instructionRun, instructionAwait, instructionStop, instructionShow:
				term.Multiple = true
			}
		}
		objects, conjunctions, err := buildGrammarObjects(term, sb.getUnassociatedTokens(), sb.gram)
		if err != nil {
			return err
		}
		sb.objects = append(sb.objects, objects...)
		if len(objects) > 0 && sb.obj == nil {
			sb.obj = objects[0]
		}
		sb.conjunctions = append(sb.conjunctions, conjunctions...)

	case term.Kind == termGrammar && term.Family == familyPhrase:
		phrase, err := buildGrammarNode(initPhraseGram(term.Name, sb.getUnassociatedTokens(), sb.gram))
		if err != nil {
			return err
		}
		if sb.phrase == nil {
			sb.phrase = phrase
		}
		sb.phrases = append(sb.phrases, phrase)
	case term.Kind == termGrammar && term.Family == familyClause:
		clauses, conjunctions, err := buildGrammarClauses(term, sb.getUnassociatedTokens(), sb.gram)
		if err != nil {
			return err
		}
		sb.clauses = append(sb.clauses, clauses...)
		sb.conjunctions = append(sb.conjunctions, conjunctions...)

	case term.Kind == termCase:
		if sb.cType == "UNKNOWN" {
			return sb.caseBuild()
		}
	default:
		return lerrors.InvalidConfiguration.Errorf("%s", "Unhandled CLAUSE Type: "+term.className())
	}
	return nil
}

// lexSwitch inits associates tokens to clauseGram fields based on the decoded term and token order
func (sb *clauseGram) lexSwitch(term grammarTerm) error {
	switch {
	case term.Kind == termLexical && term.POS == posAdverb:
		// A discovered verb may already be associated by loadTyping. Only
		// tokens preceding it are clause adverbs; post-verb possessives
		// belong to the operand ("Showing it's Output").
		available := sb.getUnassociatedTokens()
		if sb.verb != nil {
			preceding := make(map[*token]bool)
			for _, tok := range sb.tokens {
				if tok == sb.verb {
					break
				}
				preceding[tok] = true
			}
			end := 0
			for end < len(available) && preceding[available[end]] {
				end++
			}
			available = available[:end]
		}
		return term.associateLexicalTokens(available, &sb.adverbs, lexicalMatchClass)
	case term.Kind == termLexical && term.POS == posPreposition:
		return term.associateLexicalTokens(sb.getUnassociatedTokens(), &sb.prepositions, lexicalMatchClass)
	case term.Kind == termLexical && term.POS == posVerb:
		if sb.verb != nil {
			return nil
		}
		return term.associateLexicalToken(sb.getUnassociatedTokens(), &sb.verb, lexicalMatchClass)
	default:
		return lerrors.InvalidConfiguration.Errorf("%s", "Unhandled Type: "+term.className())
	}
}

func (sb *clauseGram) buildProcess(casedStr string) error {
	return applyGrammarRule(sb, casedStr, (clauseGram).clone, (*clauseGram).gramSwitch, (*clauseGram).lexSwitch, nil)
}

func (sb *clauseGram) caseBuild() error {
	cases := sb.rule.Cases
	if strings.HasPrefix(sb.gType, "clause.generic.") && sb.gType != "clause.generic.passive" {
		zeroWord := sb.verb != nil && sb.verb.binding != nil && sb.verb.binding.Record.Value != nil && sb.verb.binding.Record.Value.Lexeme != nil && len(sb.verb.binding.Record.Value.Lexeme.Inputs) == 0 && len(sb.getUnassociatedTokens()) == 0
		if sb.verb != nil && ((sb.verb.binding != nil && sb.verb.binding.Record.isFunction()) || sb.verb.isFunctionReturn() && !sb.verb.inNativeFunction() || zeroWord) {
			cases = map[string]string{"callable": cases["callable"]}
		} else {
			cases = withoutGrammarCase(cases, "callable")
		}
	}
	return selectGrammarCase(sb, cases, (clauseGram).clone, (*clauseGram).buildProcess, (*clauseGram).getUnassociatedTokens, func(s *clauseGram, name string) { s.cType = name }, sb.tokens)
}

func (sb *clauseGram) build() error {
	return buildGrammarPrefix(sb.rule, sb.cType, &sb.tokens, sb.buildProcess, sb.getUnassociatedTokens)
}

func initClauseGram(sType string, tokens []*token, gram *grammar) (*clauseGram, error) {
	var r clauseGram
	if len(tokens) == 0 {
		return &r, lerrors.Syntax.Errorf("error: clause initialized with empty slice of tokens")
	}
	r.tokens = tokens
	tErr := r.loadTyping(sType)
	if tErr != nil {
		r.tokens = []*token{} // Unlike Predicate|Statement, clause tokens unused after build will be subtracted from slice
		return &r, tErr
	}
	r.gram = gram
	g, err := r.gram.getGrammeme(r.gType)
	if err != nil {
		r.tokens = []*token{} // Unlike Predicate|Statement, clause tokens unused after build will be subtracted from slice
		return &r, err
	}
	r.rule = g
	return &r, nil
}

// predicateGram holds data for the verb and objects acting on the subject
type predicateGram struct {
	sType        string // raw type in lower case strings
	gType        string // grammeme type
	cType        string // case type
	rule         *grammeme
	tokens       []*token
	verb         *token        // i.e. in predicate.passive, "is"
	adverb       *token        // i.e. in predicate.conditional.iteration, "then"
	obj          *objectGram   // i.e. in passive.assign
	clauses      []*clauseGram // subsequent clauses
	conditionals []*conditionalGram
	conjunction  *token
	gram         *grammar
	conjunctions []*token
}

// getUnassociatedTokens recursively returns a slice of ordered token currently unassigned to sub-grammar
func (sb *predicateGram) getUnassociatedTokens() []*token {
	selected := selectLexicalTokens([]*token{sb.conjunction, sb.verb, sb.adverb}, sb.conjunctions)
	if sb.obj != nil {
		selected.add(sb.obj.tokens...)
	}
	for _, c := range sb.clauses {
		selected.add(c.tokens...)
	}
	for _, c := range sb.conditionals {
		selected.add(c.tokens...)
	}
	return selected.remaining(sb.tokens)
}

func (sb *predicateGram) gramSwitch(term grammarTerm) error {
	switch {
	case term.Kind == termGrammar && term.Family == familyObject:
		return buildGrammarChild(&sb.obj, initObjectGram, term.Name, sb.getUnassociatedTokens(), sb.gram)
	case term.Kind == termGrammar && term.Family == familyClause:
		clauses, conjunctions, err := buildGrammarClauses(term, sb.getUnassociatedTokens(), sb.gram)
		if err != nil {
			return err
		}
		sb.clauses = append(sb.clauses, clauses...)
		sb.conjunctions = append(sb.conjunctions, conjunctions...)

	case term.Kind == termGrammar && term.Family == familyConditional:
		var conditionals []*conditionalGram
		if term.Multiple {
			condGrams, conditionErr := initConditionalGrams(term.Name, sb.getUnassociatedTokens(), sb.gram)
			if conditionErr != nil {
				return conditionErr
			}
			conditionals = append(conditionals, condGrams...)
		} else {
			condGram, conditionErr := initConditionalGram(term.Name, sb.getUnassociatedTokens(), sb.gram)
			if conditionErr != nil {
				return conditionErr
			}
			conditionals = append(conditionals, condGram)
		}
		for _, conditional := range conditionals {
			if err := conditional.build(); err != nil {
				return err
			}
			sb.conditionals = append(sb.conditionals, conditional)
		}
	default:
		return lerrors.InvalidConfiguration.Errorf("%s", "Unhandled Type: "+term.className())
	}
	return nil
}

// lexSwitch inits associates tokens to clauseGram fields based on the decoded term and token order
func (sb *predicateGram) lexSwitch(term grammarTerm) error {
	switch {
	case term.Kind == termLexical && term.POS == posConjunction:
		conjunction, err := term.lexicalToken(sb.getUnassociatedTokens(), lexicalMatchClass)
		if err != nil {
			return err
		}
		sb.conjunction = conjunction
		return nil
	case term.Kind == termLexical && term.POS == posAdverb:
		return term.associateLexicalToken(sb.getUnassociatedTokens(), &sb.adverb, lexicalMatchClass)
	case term.Kind == termLexical && term.POS == posVerb:
		return term.associateLexicalToken(sb.getUnassociatedTokens(), &sb.verb, lexicalMatchClass)
	default:
		return lerrors.InvalidConfiguration.Errorf("%s", "Unhandled PREDICATE LEX Type: "+term.className())
	}
}

func (sb *predicateGram) buildProcess(casedStr string) error {
	return applyGrammarRule(sb, casedStr, (predicateGram).clone, (*predicateGram).gramSwitch, (*predicateGram).lexSwitch, nil)
}

func (sb *predicateGram) build() error {
	return buildGrammarPrefix(sb.rule, sb.cType, &sb.tokens, sb.buildProcess, sb.getUnassociatedTokens)
}

// loadTyping parses sType into gType (grammar type) and cType (case type)
func (sb *predicateGram) loadTyping(sType string) error {
	return (grammarTyping{&sb.sType, &sb.gType, &sb.cType}).load(sType, familyPredicate, grammarCaseNone)
}

func initPredicateGram(sType string, tokens []*token, gram *grammar) (*predicateGram, error) {
	var r predicateGram
	if err := r.loadTyping(sType); err != nil {
		return &r, err
	}
	r.gram = gram
	g, err := r.gram.getGrammeme(r.gType)
	if err != nil {
		return &r, err
	}
	r.rule = g
	if r.cType != "NONE" {
		_, cErr := r.rule.getCase(r.cType)
		if cErr != nil {
			return &r, cErr
		}
	}
	r.tokens = tokens
	return &r, nil
}

// statementGram holds data for the noun being acted in a statement
type statementGram struct {
	sType        string
	gType        string
	rule         *grammeme
	tokens       []*token
	subject      *subjectGram
	predicate    *predicateGram // because statement case is always a predicateGram
	clauses      []*clauseGram
	conjunctions []*token
	gram         *grammar
}

// getPrimaryType returns the primary statement type from sType, statement. i.e. verbs.declarative.root -> declarative
func (sg *statementGram) getPrimaryType() string {
	parts := strings.Split(sg.sType, ".")
	if len(parts) < 2 {
		return ""
	}
	sStr := parts[1]
	if sStr == "passive" {
		pTok := false
		for _, sToken := range sg.tokens {
			if pTok {
				// The same bound Function is a value after the copula and an
				// invocation word at the head of an ordinary active clause.
				if sToken.binding != nil && sToken.binding.Record.isFunction() && (sToken.binding.Form == "root" || sToken.binding.Form == "objective") {
					return "assign"
				}
				if sToken.getPOS(false) == "verbs" {
					return strings.Split(sToken.tType, ".")[1]
				} else {
					return "assign"
				}
			} else if sToken.getPOS(false) == "verbs" && strings.Contains(sToken.tType, ".passive.") {
				pTok = true
			}
		}
	}
	return sStr
}

// getUnassociatedTokens recursively returns a slice of ordered token currently unassigned to sub-grammar
func (sg *statementGram) getUnassociatedTokens() []*token {
	selected := selectTokens()
	if sg.subject != nil {
		selected.add(sg.subject.tokens...)
	}
	if sg.predicate != nil {
		selected.add(sg.predicate.tokens...)
	}
	for _, clause := range sg.clauses {
		selected.add(clause.tokens...)
	}
	selected.add(sg.conjunctions...)
	return selected.remaining(sg.tokens)
}

func (sg *statementGram) gramSwitch(term grammarTerm) error {
	switch {
	case term.Kind == termGrammar && term.Family == familyClause:
		clauses, conjunctions, err := buildGrammarClauses(term, sg.getUnassociatedTokens(), sg.gram)
		if err != nil {
			return err
		}
		sg.clauses = append(sg.clauses, clauses...)
		sg.conjunctions = append(sg.conjunctions, conjunctions...)
	case term.Kind == termGrammar && term.Family == familyPredicate:
		predGram, predErr := initPredicateGram(term.Name, sg.getUnassociatedTokens(), sg.gram)
		if predErr != nil {
			return predErr
		}
		sg.predicate = predGram
		predErr = sg.predicate.build()
		if predErr != nil {
			return predErr
		}
	case term.Kind == termGrammar && term.Family == familySubject:
		subGram, subErr := initSubjectGram(term.Name, sg.getUnassociatedTokens(), sg.gram)
		if subErr != nil {
			return subErr
		}
		sg.subject = subGram
		subErr = sg.subject.build()
		if subErr != nil {
			return subErr
		}
		_ = sg.subject.getAssociatedTokens(true)
	default:
		return lerrors.InvalidConfiguration.Errorf("%s", "Unhandled Type: "+term.className())
	}
	return nil
}

// build identifies and inits & builds pointers to subjectGram and predicateGram for subject and predicate
func (sg *statementGram) build() error {
	casedStr, err := sg.rule.getCasedStructure(sg.getPrimaryType())
	if err != nil {
		return err
	}
	if err = applyGrammarRule(sg, casedStr, (statementGram).clone, (*statementGram).gramSwitch, nil, nil); err != nil {
		return err
	}
	if remaining := sg.getUnassociatedTokens(); len(remaining) != 0 {
		return lerrors.Syntax.Errorf("statement has %d unassociated tokens, beginning with %q", len(remaining), remaining[0].rawBytes)
	}
	_, err = sg.declarationVisibilities()
	return err
}

func initStatementGram(lsType string, tokens []*token, gram *grammar) (*statementGram, error) {
	var rGram statementGram
	rGram.sType = lsType
	rGram.gram = gram
	rGram.tokens = tokens
	gType, err := statementTypeMap(rGram.sType)
	if err != nil {
		return &rGram, err
	}
	rGram.gType = gType
	g, err := rGram.gram.getGrammeme(rGram.gType)
	if err != nil {
		return &rGram, err
	}
	rGram.rule = g
	return &rGram, nil
}
