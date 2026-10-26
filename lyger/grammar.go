package Lyger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"strconv"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// importStatement holds the data used by the compilation builder to manage import statements
type importStatement struct {
	statementBytes []byte // i.e. Use CoreLib's Print from `core.l` as "Print".\n
	module         string // i.e. CoreLib
	keyWord        string // i.e. Print
	filePath       string // i.e. core.l
	alias          string // i.e. Print
}

// load data from an inputted a pattern, import type, and stmt bytes
func (i *importStatement) load(iType string, stmt []byte, words [][]byte) {
	*i = importStatement{statementBytes: bytes.Clone(stmt)}
	if iType == "file" {
		i.filePath = unquoteImportWord(words[1])
		return
	}
	object := 1
	if owner, ok := possessiveOwner(string(words[object])); ok {
		i.module = owner
		object++
	}
	i.keyWord = unquoteImportWord(words[object])
	i.filePath = unquoteImportWord(words[object+2])
	if iType == "object.alias" {
		i.alias = unquoteImportWord(words[object+4])
	}
}

func unquoteImportWord(word []byte) string {
	if len(word) >= 2 && word[0] == '`' && word[len(word)-1] == '`' {
		return strings.ReplaceAll(string(word[1:len(word)-1]), "\\`", "`")
	}
	if value, err := strconv.Unquote(string(word)); err == nil {
		return value
	}
	return string(word)
}

// statementTypeMap inputs a verb token sType and returns a grammeme readable gType
func statementTypeMap(sType string) (string, error) {
	switch sType {
	case "verbs.passive.root":
		return "statement.subjective.passive", nil
	case "evaluations.conditional.root", "evaluations.alternative.root", "evaluations.iterative.root":
		return "statement.objective.conditional", nil
	}
	splitType := strings.Split(sType, ".")
	baseType := splitType[len(splitType)-1]
	switch baseType {
	case "root":
		return "statement.objective.imperative", nil
	case "active":
		return "statement.subjective.active", nil
	case "passive":
		return "statement.subjective.passive", nil
	default:
		return "unknown", lerrors.Syntax.Errorf("unknown")
	}
}

// grammar structures the vocabulary and statement methods required for compilation
type grammar struct {
	resources  fs.FS
	assetRoot  string
	language   string
	lex        *lexicon
	grammarMap map[string]grammeme
}

// grammeme lookup preprocess to ensure "root" clause types are transformed into imperative
func gramTypePreProc(gType string) string {
	sStr := strings.Split(gType, ".")
	if len(sStr) == 3 {
		if sStr[0] == "clause" && sStr[2] == "root" {
			gType = strings.Replace(gType, ".root", ".imperative", 1)
		}
	}
	return gType
}

// getGrammeme returns a grammeme or an error for an inputted gType
func (g *grammar) getGrammeme(gType string) (*grammeme, error) {
	term, err := decodeGrammarTerm(gType)
	if err != nil {
		return nil, err
	}
	cleanedType := gramTypePreProc(term.Name)
	if gram, ok := g.grammarMap[cleanedType]; ok {
		return &gram, nil
	}
	return &grammeme{}, lerrors.InvalidConfiguration.Errorf("%s", "grammeme type not found: "+gType)
}

// Import forms use one tokenized source statement for matching and extraction.
var importRules = []struct {
	name    string
	pattern []string
}{
	{"file", []string{"USE", "~"}},
	{"object", []string{"USE", "OBJECT", "FROM", "~"}},
	{"object.alias", []string{"USE", "OBJECT", "FROM", "~", "AS", "ID"}},
}

// checkImportPatterns compares an import grammar pattern with a check stmt pattern
func checkImportPatterns(importPattern []string, checkedPattern []string) bool {
	if len(importPattern) == len(checkedPattern) {
		for ind, imp := range importPattern {
			if imp == "OBJECT" {
				if imp != checkedPattern[ind] && "POSSESSIVE."+imp != checkedPattern[ind] {
					return false
				}
			} else {
				if imp != checkedPattern[ind] {
					return false
				}
			}
		}
		return true
	}
	return false
}

// importWordPattern classifies each already-tokenized word once.
func (g *grammar) importWordPattern(words [][]byte) ([]string, bool, error) {
	if g == nil || g.lex == nil {
		return nil, false, lerrors.InvalidConfiguration.Errorf("import recognition requires a loaded lexicon")
	}
	var pattern []string
	possessive := false
	for _, word := range words {
		text := string(word)
		if _, ok := possessiveOwner(text); ok {
			if possessive {
				return pattern, false, nil
			}
			possessive = true
			continue
		}
		kind := "OBJECT"
		if strings.HasPrefix(text, "`") && strings.HasSuffix(text, "`") {
			kind = "~"
		} else if strings.HasPrefix(text, `"`) && strings.HasSuffix(text, `"`) {
			kind = "ID"
		} else if lex, _, err := g.lex.lookup(word); err == nil {
			if matched, form := lex.match(cleanTokenBytes(word)); matched && form == "root" {
				switch {
				case lex.POS == "verb" && lex.Type == "action.associative":
					kind = "USE"
				case lex.POS == "preposition" && lex.Type == "conjoinment.objective":
					kind = "FROM"
				case lex.POS == "preposition" && lex.Type == "definement.declarative":
					kind = "AS"
				}
			}
		}
		if possessive {
			if kind != "OBJECT" {
				return pattern, false, nil
			}
			kind, possessive = "POSSESSIVE.OBJECT", false
		}
		pattern = append(pattern, kind)
	}
	return pattern, !possessive, nil
}

// isImportStatement checks if statement imports from another file
func (g *grammar) isImportStatement(stmt []byte) (bool, *importStatement, error) {
	var imported importStatement
	words, err := sourceWords(statementBody(stmt))
	if err != nil {
		return false, &imported, err
	}
	pattern, valid, err := g.importWordPattern(words)
	if err != nil || !valid {
		return false, &imported, err
	}
	for _, rule := range importRules {
		if checkImportPatterns(rule.pattern, pattern) {
			imported.load(rule.name, stmt, words)
			return true, &imported, nil
		}
	}
	return false, &imported, nil
}

func loadGrammar(options Options) (*grammar, error) {
	options = options.normalized()
	path, err := languageDefinitionPath(options.Language, "grammar.json")
	if err != nil {
		return nil, fmt.Errorf("load grammar: %w", err)
	}
	lex, err := loadLexicon(options.Resources, options.Language)
	if err != nil {
		return nil, fmt.Errorf("load grammar: %w", err)
	}
	rules, err := readGrammarRules(options.Resources, path)
	if err != nil {
		return nil, fmt.Errorf("load grammar: %w", err)
	}
	return &grammar{lex: lex, grammarMap: rules, resources: options.Resources, assetRoot: options.AssetRoot, language: options.Language}, nil
}

func languageDefinitionPath(language, filename string) (string, error) {
	if strings.TrimSpace(language) == "" || strings.ContainsAny(language, "/\\") || !fs.ValidPath(language) {
		return "", lerrors.InvalidConfiguration.Errorf("valid language is required: %q", language)
	}
	return path.Join("lexicon", language, filename), nil
}

// readLanguageJSON checks one complete object and uses ReadFile so loading a
// language does not leave open descriptors or terminate the calling process.
func readLanguageJSON(resources fs.FS, path string, target interface{}) error {
	content, err := fs.ReadFile(resources, path)
	if err != nil {
		return lerrors.IO.Errorf("read language definition %s: %w", path, err)
	}
	trimmed := bytes.TrimSpace(content)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return lerrors.InvalidConfiguration.Errorf("decode language definition %s: expected a JSON object", path)
	}
	if err := json.Unmarshal(content, target); err != nil {
		return lerrors.InvalidConfiguration.Errorf("decode language definition %s: %w", path, err)
	}
	return nil
}

// lStatement encapsulates and processes a statement in L into an instruction from bytes
type lStatement struct {
	rawBytes  []byte   // raw lStatement byte slice from beginning of process
	sType     string   // statement type, i.e. `IMPERATIVE.CREATE.VERB`
	headBytes []byte   // bytes of the statement header bytes
	stmtBytes []byte   // stmt bytes are rawBytes minus headBytes, processed into tokens
	tokens    []*token // output of lStatement tokenization method (i.e. LProcess D)
	sGram     *statementGram
	records   []*record
	gram      *grammar
	source    *SourceStatement // original source span, independent of parser stamps
}

// uncacheLiterals restores each original token once. Grammar nodes share these
// token pointers, so restoration does not traverse their overlapping slices.
func (l *lStatement) uncacheLiterals(textCache map[string][]byte) error {
	for _, token := range l.tokens {
		if err := token.restoreLiteral(textCache); err != nil {
			return err
		}
	}
	return nil
}

// tokenize an lStatements stmtBytes into tokenBytes to load into token structs
func (l *lStatement) tokenize() error {
	p := newParser(l.stmtBytes)
	for _, tokenBytes := range p.parseTokens() {
		newT := newToken(tokenBytes, l.gram)
		err := newT.scan()
		if err != nil {
			return err
		}
		l.tokens = append(l.tokens, newT)
	}
	return nil
}

// determineType of lStatement based on first verb identified.
func (l *lStatement) determineType() {
	for ind, sToken := range l.tokens {
		if ind == 0 && l.sType == "" {
			isCond, _ := sToken.isConditional()
			if isCond {
				l.sType = sToken.tType
			}
		}
		if sToken.getPOS(false) == "verbs" && l.sType == "" {
			if l.sType == "" {
				l.sType = sToken.tType
			}
		}
		if sToken.getPOS(false) == "unknown" {
			l.sType = "unknown"
			break
		}
		/*
			if sToken.getPOS(false) == "unknown" || sToken.tType == "nouns.NotFound.possessive" {
					l.sType = "unknown"
					break
			}
		*/
	}
}

// loadGram initializes a new initStatementGram structure for the lStatement
func (l *lStatement) loadGram() error {
	// Native roots are configuration entries rather than bindable declarations.
	// Diagnose aliasing that unsupported capability before the generic ID rule
	// obscures the cause; source-defined actions remain freely aliasable.
	for index, tok := range l.tokens {
		lex := tok.lexeme()
		if lex != nil && lex.Type == "action.associative" && index+1 < len(l.tokens) {
			if _, native := l.tokens[index+1].nativeRole(); native && l.tokens[index+1].binding == nil {
				return lerrors.Unsupported.Errorf("alias of native root %q is unsupported; define a source action instead", l.tokens[index+1].rawBytes)
			}
		}
	}
	rGram, err := initStatementGram(l.sType, l.tokens, l.gram)
	if err != nil {
		return fmt.Errorf("initialize statement rule for %s: %w", l.sType, err)
	}
	l.sGram = rGram
	err = l.sGram.build()
	if err != nil {
		family, _ := statementTypeMap(l.sType) //nolint:errcheck // The original grammar error stays primary; unknown is the intended fallback diagnostic prefix.
		return fmt.Errorf("%s: %w", family, err)
	}
	return nil
}

// newLStatement inputs rawBytes and grammar, and loads them into headBytes, stmtBytes, and tokens
func newLStatement(rawBytes []byte, gram *grammar) (*lStatement, error) {
	newStmt := lStatement{
		rawBytes:  rawBytes,
		headBytes: []byte(""),
		stmtBytes: rawBytes,
		records:   []*record{},
		gram:      gram,
	}
	p := newParser(newStmt.rawBytes)
	headStamps, err := p.parse(parseLineStamp)
	if err != nil {
		return &newStmt, err
	}
	if len(headStamps) > 1 {
		return &newStmt, lerrors.Internal.Errorf("lStatement should not contain more than a single head stamp")
	} else if len(headStamps) == 1 {
		newStmt.headBytes = headStamps[0]
		newStmt.stmtBytes = bytes.Replace(newStmt.stmtBytes, newStmt.headBytes, []byte(""), -1)
	}
	return &newStmt, nil
}
