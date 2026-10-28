package Lyger

import (
	"fmt"
	"path/filepath"

	"github.com/l-programming-language/l/lerrors"
)

// Compilation organizes the pre-assembled structure of an L program as it is recursively pre-processed
type Compilation struct {
	definitionTemporaries []int
	definitionReserved    map[string]bool
	definitionStatements  int
	Gram                  *grammar
	MainBytes             []byte
	MainPath              string
	RootDir               string
	CompiledBytes         []byte
	ASMBytes              []byte
	Memory                map[string]*record
	TextCache             map[string][]byte
	Statements            []*lStatement
	ProcStatements        []*lStatement
	Program               *Program
	Error                 error
	Prepared              *PreparedSource
	SourceStatements      []SourceStatement
	SourcePaths           []string
	SourceOccurrences     []sourceOccurrence
	SourceUnits           []*SourceUnit
	SourceOrigins         []SourceOrigin
	Semantics             *semanticProgram
	Scope                 *scopePlan
	ExpandedBytes         []byte
	builtinProcesses      *builtinProcessCatalog
	nativeDeclarations    map[*record]bool
}

// newCompilation returns a newly created Compilation struct
func newCompilation(mainBytes []byte, mainPath string) *Compilation {
	var comp Compilation
	comp.MainBytes = mainBytes
	comp.MainPath = mainPath
	return &comp
}

// checkError determines if the Compilation structure has an error or not
func (c *Compilation) checkError(err error) {
	if err != nil && c.Error == nil {
		c.Error = err
	}
}

// checkLError determines if the Compilation structure has an error or not
func (c *Compilation) checkLError(err error, headBytes []byte) bool {
	if err != nil {
		if c.Error != nil {
			return true
		}
		c.Error = err
		if len(headBytes) > 0 {
			c.Error = fmt.Errorf("%s || %w", headBytes, err)
		}
		return true
	}
	return false
}

// init initializes the Compilation structure by ensuring MainBytes is not nil and the RootDir is defined
func (c *Compilation) init(options Options) {
	if len(c.MainBytes) == 0 {
		c.Error = lerrors.InvalidInput.Errorf("error Compilation.init(): MainBytes cannot be empty or nil")
		return
	}
	c.RootDir = filepath.Dir(c.MainPath)
	c.Memory = make(map[string]*record) // initialize working memory
	c.Semantics = newSemanticProgram()
	gram, err := loadGrammar(options)
	if err != nil {
		c.checkError(err)
		return
	}
	c.Gram = gram
	c.builtinProcesses, err = loadBuiltinProcesses(gram.resources)
	c.checkError(err)
}

// preProcess caches text data in MainBytes and strips out comments and cleans MainBytes.
func (c *Compilation) preProcess() {
	prepared, err := PrepareSource(c.MainPath, c.MainBytes)
	c.checkError(err)
	if err != nil {
		return
	}
	c.Prepared = prepared
	c.CompiledBytes = prepared.PreparedBytes
	c.TextCache = prepared.TextCache
}

// build recursively checks for import statements in the L code and builds up CompiledBytes if found using builder
func (c *Compilation) build() {
	b := newPreparedBuilder(c.Prepared, c.Gram)
	if err := b.process(); err != nil {
		c.checkError(fmt.Errorf("expand %s: %w", c.MainPath, err))
		return
	}
	c.CompiledBytes = b.BuiltBytes
	c.TextCache = b.TextCache
	c.SourceStatements = b.SourceStatements
	c.SourcePaths = b.SourcePaths
	c.SourceOccurrences = b.occurrences
	c.SourceUnits = b.units
	c.SourceOrigins = b.origins
}

// scan tokenize & scan the grammar/lexicon of Compilation statements
func (c *Compilation) scan() {
	var isErr bool
	p := newParser(c.CompiledBytes)
	for index, stmtBytes := range p.parseStatements() {
		lStmt, err := newLStatement(stmtBytes, c.Gram)
		isErr = c.checkLError(err, stmtBytes)
		if isErr {
			return
		}
		if index < len(c.SourceStatements) {
			lStmt.source = &c.SourceStatements[index]
		}
		isErr = c.checkStatementError(lStmt, "tokenize", lStmt.tokenize())
		if isErr {
			return
		}
		c.Statements = append(c.Statements, lStmt)
	}
}

// unresolvedNumberError retains a lexical range failure only when exhausted
// semantic resolution has not identified the token as a declared or member name.
func (s *lStatement) unresolvedNumberError() error {
candidates:
	for i, tok := range s.tokens {
		if tok.rangeError == nil || tok.binding != nil || tok.defined || tok.getPOS(false) != "unknown" {
			continue
		}
		if i > 0 {
			if _, possessed := possessiveOwner(string(s.tokens[i-1].rawBytes)); possessed {
				continue
			}
		}
		// Include root reservations here without changing root lookup/replacement.
		for scope := tok.scope; scope != nil; scope = scope.parent {
			for _, declaration := range scope.declarations {
				if declaration.matches(tok.rawBytes) {
					continue candidates
				}
			}
		}
		return tok.rangeError
	}
	return nil
}

// buildMemory recursively gets statement types, build statement grammars, builds working memory, and resolves unknowns
func (c *Compilation) buildMemory() {
	for len(c.ProcStatements) > 0 {
		var pending []*lStatement
		progress := false
		for _, statement := range c.ProcStatements {
			if len(statement.tokens) > 0 && statement.tokens[0].scope.waitsForHeader(statement) {
				pending = append(pending, statement)
				continue
			}
			ready, err := c.dictionaryImportReady(statement)
			if err == nil && ready {
				ready, err = c.resolver().bindStatementMembers(statement)
			}
			if c.checkStatementError(statement, "bind members", err) {
				return
			}
			if !ready {
				pending = append(pending, statement)
				continue
			}
			for _, token := range statement.tokens {
				if token.bindingError != nil {
					c.checkStatementError(statement, "bind reference", token.bindingError)
					return
				}
			}
			statement.sType = ""
			statement.determineType()
			if statement.sType == "unknown" {
				pending = append(pending, statement)
				continue
			}
			if c.checkStatementError(statement, "grammar", statement.loadGram()) {
				return
			}
			if c.checkStatementError(statement, "resolve literals", statement.uncacheLiterals(c.TextCache)) {
				return
			}
			if c.checkStatementError(statement, "register statement", c.registerStatement(statement)) {
				return
			}
			for _, tok := range statement.tokens {
				if tok.pendingDeclaration == nil {
					continue
				}
				state, err := c.resolver().bindToken(tok)
				if err == nil && (state != lookupFound || tok.binding.Record != c.Semantics.Subjects[tok.pendingDeclaration]) {
					err = lerrors.UnresolvedReference.Errorf("same-statement declaration %s was not published", tok.rawBytes)
				}
				if c.checkStatementError(statement, "bind declaration", err) {
					return
				}
				tok.pendingDeclaration = nil
			}
			progress = true
		}
		c.ProcStatements = pending
		if !progress {
			if err := c.validateDictionaryExports(); err != nil {
				c.checkStatementError(pending[0], "dictionary export", err)
				return
			}
			if imported := pending[0].source.imported; imported != nil {
				c.checkStatementError(pending[0], "dictionary import", lerrors.UnresolvedReference.Errorf("dictionary %s does not export %s.%s", imported.unit.Origin.Name, imported.module, imported.member))
				return
			}

			for _, statement := range pending {
				if err := statement.unresolvedNumberError(); err != nil {
					c.checkStatementError(statement, "tokenize", err)
					return
				}
			}
			c.checkStatementError(pending[0], "bind reference", lerrors.UnresolvedReference.Errorf("unresolved references after dependency resolution: %s", pending[0].stmtBytes))
			return
		}
	}
}

// analyze the grammar/lexicon of Compilation tokenized statements to classify
func (c *Compilation) analyze() {
	var err error
	c.Scope, err = planSourceScopes(c.Statements, c.Memory, c.SourceOccurrences)
	if err != nil {
		c.checkError(err)
		return
	}
	c.ProcStatements = append([]*lStatement{}, c.Statements...)
	// 1) determine each c.Statement type (i.e. IMPERATIVE.CREATE.VERB) with Subject and Predicate Tokens (LProcess E.1)
	//		1.A) Any statement with an unknown is set aside.
	//		1.B) Create working memory map for declared/defined IDs
	//		1.C) Recursively repeat steps 1.A & 1.B on unknowns
	c.buildMemory()
	if c.Error == nil {
		c.checkError(c.validateDictionaryExports())
	}
	// 2) analyze SUBJECT and PREDICATE Grammar of each c.Statement (LProcess E.2)
	// 3) identify working memory records (LProcess F.1)
	// 4) analyze lexicon of each non-build/run statement and load output to identified memory records (LProcess F.2)
}

// Make resolves source statements, constructs applications, and publishes AML.
func (c *Compilation) Make() {
	if c.Error != nil || c.Program != nil {
		return
	}
	c.scan() // tokenize & scan the grammar/lexicon of Compilation statements
	if c.Error != nil {
		return
	}
	c.analyze() // analyze the grammar/lexicon of Compilation tokenized statements to classify
	if c.Error != nil {
		return
	}
	var err error
	c.Semantics.Instructions, err = buildInstructions(c.Statements, c.Scope, c.Semantics)
	c.checkError(err)
	if c.Error != nil {
		return
	}
	c.ExpandedBytes = renderExpandedSource(c.SourceStatements, c.TextCache)
	c.checkError(c.constructApplications())
	if c.Error != nil {
		return
	}
	c.checkError(c.lowerProgram())
}
