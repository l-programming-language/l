package Lyger

import (
	"strconv"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

type pronounScopeKind uint8

const (
	scopeInvalid pronounScopeKind = iota
	scopeParagraph
	scopeActionArgument
	pronounScopeKindCount
)

var pronounScopeNames = [pronounScopeKindCount]string{
	scopeParagraph: "paragraph", scopeActionArgument: "action-argument",
}

func (k pronounScopeKind) String() string {
	if k >= pronounScopeKindCount {
		return "pronounScopeKind(" + strconv.Itoa(int(k)) + ")"
	}
	return pronounScopeNames[k]
}

func (k pronounScopeKind) MarshalText() ([]byte, error) {
	if k >= pronounScopeKindCount {
		return nil, lerrors.Internal.Errorf("invalid pronoun scope %d", uint8(k))
	}
	return []byte(pronounScopeNames[k]), nil
}

// pronounBinding identifies a semantic subject without changing the token's
// lexical pronoun role. Records and action arguments can be resolved later.
type pronounBinding struct {
	SubjectID        string
	Kind             pronounScopeKind
	Declaration      *lStatement
	TargetToken      *token
	TargetTokenIndex int
	TokenIndex       int
	Possessive       bool
}

type scopeDeclaration struct {
	SubjectID  string
	Token      *token
	TokenIndex int
	Scope      *lexicalScope
	Owner      *token
	Record     *record
}

type scopedStatement struct {
	Statement            *lStatement
	Scope                *lexicalScope
	Order                int // expanded source order, independent of retry order
	Source               SourceSpan
	IncomingSubjectID    string
	IncomingSubjectToken *token
	DeclaredSubjectID    string
	Declarations         []scopeDeclaration
	Arguments            []scopeDeclaration
}

// Effects can follow declarations within one sentence. Their implicit context
// sees only the preceding declaration prefix; RHS pronouns keep their original
// incoming bindings in scopePlan.Pronouns.
func (s scopedStatement) declarationsBefore(at *token) []scopeDeclaration {
	position := -1
	for index, tok := range s.Statement.tokens {
		if tok == at {
			position = index
			break
		}
	}
	count := 0
	for _, declaration := range s.Declarations {
		if declaration.TokenIndex >= position {
			break
		}
		count++
	}
	return s.Declarations[:count]
}

func (s scopedStatement) subjectBefore(at *token) *token {
	declarations := s.declarationsBefore(at)
	if len(declarations) > 0 {
		return declarations[len(declarations)-1].Token
	}
	return s.IncomingSubjectToken
}

type scopePlan struct {
	Root         *lexicalScope
	environments map[sourceParagraphKey]*lexicalScope
	declarations map[*token]*scopeDeclaration
	Ordered      []scopedStatement
	Pronouns     map[*token]pronounBinding
}

type scopedSubject struct {
	id         string
	statement  *lStatement
	token      *token
	tokenIndex int
}

func (t *token) lexeme() *lexeme {
	if t != nil && t.binding != nil && t.binding.Record.isFunction() {
		return t.binding.Record.bindingWord()
	}
	if t != nil && t.mem != nil {
		return t.mem.Lexeme
	}
	return nil
}

func (t *token) isFunctionReturn() bool {
	if t == nil || t.binding == nil {
		return false
	}
	for scope := t.scope; scope != nil; scope = scope.parent {
		if scope.function != nil {
			selected, err := canonicalRecord(scope.function.Return, scope.function.aliases)
			bound, boundErr := canonicalRecord(t.binding.Record, scope.function.aliases)
			return err == nil && boundErr == nil && selected != nil && selected == bound
		}
	}
	return false
}

func (t *token) hasRole(pos, role string) bool {
	lex := t.lexeme()
	return lex != nil && lex.POS == pos && lex.Type == role
}

func (t *token) isPronoun() bool {
	lex := t.lexeme()
	return lex != nil && lex.POS == "noun" && strings.HasPrefix(lex.Type, "pronoun.")
}

func (t *token) declarationRoot() string {
	if t == nil || !t.isDeclare() {
		return ""
	}
	label, err := strconv.Unquote(string(t.rawBytes))
	if err != nil {
		return ""
	}
	return strings.ToLower(strings.SplitN(label, "|", 2)[0])
}

// scopeDeclarations recognizes labels in the existing declarative/definitive
// syntax, including an associative alias and an inline conditional declaration.
// Quoted names nested inside a directive body do not replace its outer subject.
func scopeDeclarations(statement *lStatement, incoming scopedSubject) ([]scopeDeclaration, int) {
	var declarations []scopeDeclaration
	pending, directive := false, -1
	invocation := false
	for index, current := range statement.tokens {
		invocation = invocation || (current.hasRole("preposition", "definement.declarative") && len(declarations) == 0)
		if current.hasRole("preposition", "conjoinment.directive") && len(declarations) > 0 {
			directive = index
			break
		}
		if current.hasRole("verb", "action.declarative") || current.hasRole("verb", "action.definitive") || current.hasRole("verb", "action.associative") {
			pending = true
			continue
		}
		if root := current.declarationRoot(); root != "" && (index == 0 || pending || invocation || (index+1 < len(statement.tokens) && statement.tokens[index+1].getPOS(false) == "verbs" && strings.Contains(statement.tokens[index+1].tType, ".passive."))) {
			if index > 0 {
				previous := statement.tokens[index-1]
				if owner, possessed := possessiveOwner(string(previous.rawBytes)); possessed {
					if previous.isPronoun() {
						owner = incoming.id
					}
					root = strings.ToLower(owner) + "." + root
				}
			}
			declarations = append(declarations, scopeDeclaration{SubjectID: root, Token: current, TokenIndex: index})
			pending = false
		}
	}
	return declarations, directive
}

// actionArgument recognizes the nominal after the variable article in a
// directive body. Its identity is an argument slot, not the name of a type or
// a source-paragraph declaration. This covers source-defined action composition
// such as Building any App and then Running it without knowing the action name.
func actionArgument(tokens []*token, article int) (int, bool) {
	for index := article + 1; index < len(tokens); index++ {
		current := tokens[index]
		pos := current.getPOS(false)
		if current.isPronoun() {
			return 0, false
		}
		if pos == "nouns" || pos == "type" || pos == "unknown" {
			return index, true
		}
		if pos != "adjectives" && pos != "articles" {
			return 0, false
		}
	}
	return 0, false
}

// planSourceScopes runs once on the immutable scan order, before a dependency
// worklist retries statements. A file/paragraph pair has its own antecedent;
// imported source and descriptor refinements cannot leak a different subject
// into the caller. All references in a statement see its incoming paragraph
// subject, then its declarations become available to subsequent statements.
func planSourceScopes(statements []*lStatement, memory map[string]*record, occurrences []sourceOccurrence) (*scopePlan, error) {
	if memory == nil {
		memory = make(map[string]*record)
	}
	plan := &scopePlan{Root: &lexicalScope{records: memory}, Pronouns: make(map[*token]pronounBinding), declarations: make(map[*token]*scopeDeclaration)}
	if err := plan.prepareEnvironments(occurrences); err != nil {
		return plan, err
	}
	paragraphs := make(map[sourceParagraphKey]scopedSubject)
	for order, statement := range statements {
		if statement == nil || statement.source == nil {
			return plan, lerrors.Internal.Errorf("statement %d has no original source provenance", order)
		}
		source := statement.source.Span
		if source.Path == "" || source.Paragraph < 1 {
			return plan, source.wrapError(lerrors.Internal.Errorf("statement has no source paragraph identity"))
		}
		key := sourceParagraphKey{statement.source.occurrence, source.Path, source.Paragraph}
		environment := plan.environment(statement.source)
		for _, tok := range statement.tokens {
			tok.scope = environment
		}
		if imported := statement.source.imported; imported != nil {
			// The selected export resolves in its own lexical environment; the
			// quoted binding declaration remains owned by the importer.
			for i := 1; i < len(statement.tokens)-2; i++ {
				statement.tokens[i].scope = imported.unit.root
			}
		}
		incoming := paragraphs[key]
		declarations, directive := scopeDeclarations(statement, incoming)
		ordered := scopedStatement{Statement: statement, Scope: environment, Order: order, Source: source, IncomingSubjectID: incoming.id, IncomingSubjectToken: incoming.token, Declarations: declarations}
		var arguments []scopedSubject
		for index, current := range statement.tokens {
			if directive >= 0 && index > directive && current.hasRole("article", "variable") {
				if target, ok := actionArgument(statement.tokens, index); ok {
					id := declarations[0].SubjectID + "::argument:" + strconv.Itoa(target)
					argument := scopedSubject{id: id, statement: statement, token: statement.tokens[target], tokenIndex: target}
					arguments = append(arguments, argument)
					ordered.Arguments = append(ordered.Arguments, scopeDeclaration{SubjectID: id, Token: argument.token, TokenIndex: target})
				}
			}
			if !current.isPronoun() {
				continue
			}
			subject, kind := incoming, scopeParagraph
			if directive >= 0 && index > directive && len(arguments) > 0 {
				if len(arguments) > 1 {
					return plan, source.wrapError(lerrors.AmbiguousReference.Errorf("pronoun %q has more than one preceding generic action argument", current.rawBytes))
				}
				subject, kind = arguments[0], scopeActionArgument
			}
			if subject.id == "" {
				return plan, source.wrapError(lerrors.UnresolvedReference.Errorf("pronoun %q has no preceding subject in source paragraph %d", current.rawBytes, source.Paragraph))
			}
			_, possessive := possessiveOwner(string(current.rawBytes))
			plan.Pronouns[current] = pronounBinding{SubjectID: subject.id, Kind: kind, Declaration: subject.statement, TargetToken: subject.token, TargetTokenIndex: subject.tokenIndex, TokenIndex: index, Possessive: possessive || current.isPossessive()}
		}
		if len(declarations) > 0 {
			last := declarations[len(declarations)-1]
			ordered.DeclaredSubjectID = last.SubjectID
			paragraphs[key] = scopedSubject{id: last.SubjectID, statement: statement, token: last.Token, tokenIndex: last.TokenIndex}
		}
		ordered.Scope = environment
		plan.Ordered = append(plan.Ordered, ordered)
		for index := range declarations {
			declaration := &declarations[index]
			declaration.Scope = environment
			if declaration.TokenIndex > 0 {
				prior := statement.tokens[declaration.TokenIndex-1]
				if _, possessed := possessiveOwner(string(prior.rawBytes)); possessed {
					declaration.Owner = prior
				}
			}
			environment.declarations = append(environment.declarations, declaration)
			plan.declarations[declaration.Token] = declaration
		}
	}
	return plan, nil
}

func (t *token) nativeRole() (nativeRole, bool) {
	lex := t.lexeme()
	if lex == nil {
		return roleInvalid, false
	}
	return parseNativeRole(lex.Type)
}
