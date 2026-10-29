package Lyger

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// lexicalScope owns bindings, independently of physical paragraph pronouns and
// runtime activation. The root records map is Compilation.Memory itself.
type lexicalScope struct {
	kind                scopeKind
	unit                *SourceUnit
	parent              *lexicalScope
	function            *sourceFunction
	callableHeader      *lStatement
	records             map[string]*record
	bindings            *recordBindingIndex
	declarations        []*scopeDeclaration
	reservations        map[string][]*scopeDeclaration
	indexedDeclarations int
	source              SourceSpan
}

type scopeKind uint8

const (
	lexicalProgram scopeKind = iota
	lexicalBuiltin
	lexicalDictionary
	lexicalParagraph
)

// An explicit callable signature determines how its paragraph declarations bind.
// Constant-only locals must not commit while that header awaits a forward type.
func (s *lexicalScope) waitsForHeader(statement *lStatement) bool {
	for ; s != nil; s = s.parent {
		if s.callableHeader != nil && s.callableHeader != statement && s.callableHeader.sGram == nil {
			return true
		}
	}
	return false
}

func (s *lexicalScope) isParagraph() bool { return s != nil && s.kind == lexicalParagraph }
func (s *lexicalScope) unitRoot() *lexicalScope {
	for s != nil && s.isParagraph() {
		s = s.parent
	}
	return s
}

type lookupState uint8

const (
	lookupAbsent lookupState = iota
	lookupPending
	lookupFound
	lookupStateCount
)

func (s lookupState) String() string {
	if s >= lookupStateCount {
		return fmt.Sprintf("lookupState(%d)", s)
	}
	return [...]string{"absent", "pending", "found"}[s]
}

type sourceParagraphKey struct {
	Occurrence sourceOccurrenceID
	Path       string
	Number     int
}

func (s *lexicalScope) lookupLocal(raw []byte) (*tokenBinding, lookupState, error) {
	// An unresolved local reserves every declared surface form before retries.
	// This is inheritance, not activation: program/dictionary unit roots
	// also reserve their local words before looking in the builtin parent.
	if s.parent != nil {
		for _, declaration := range s.reserved(raw) {
			if declaration.Record == nil {
				return nil, lookupPending, nil
			}
		}
	}
	records := s.records
	if s.bindings != nil {
		records = s.bindings.words[newBindingQuery(raw).word]
	}
	binding, err := resolveTokenBinding(raw, records)
	if err != nil || binding == nil {
		return nil, lookupAbsent, err
	}
	return binding, lookupFound, nil
}

// Source labels and owners stay fixed while their records are published. Index
// their surface forms once, but inspect the live reservation on every lookup;
// a pending local must still shadow an inherited word until publication.
func (s *lexicalScope) reserved(raw []byte) []*scopeDeclaration {
	if s.reservations == nil {
		s.reservations = make(map[string][]*scopeDeclaration)
	}
	for _, declaration := range s.declarations[s.indexedDeclarations:] {
		for _, word := range declaration.surfaceForms() {
			s.reservations[word] = append(s.reservations[word], declaration)
		}
	}
	s.indexedDeclarations = len(s.declarations)
	return s.reservations[newBindingQuery(raw).word]
}

func (d *scopeDeclaration) surfaceForms() []string {
	prefix := ""
	if d.Owner != nil {
		// Owned members use their owner's table. Native type namespaces
		// instead reserve the qualified form in the lexical environment.
		name, _ := possessiveOwner(string(d.Owner.rawBytes))
		kind, _ := checkType([]byte(name))
		if kind == "Unknown" {
			return nil
		}
		prefix = strings.ToLower(kind) + "."
	}
	label, err := strconv.Unquote(string(d.Token.rawBytes))
	if err != nil {
		return nil
	}
	parts := strings.Split(strings.ToLower(label), "|")
	forms := []string{prefix + parts[0]}
	for _, fix := range parts[1:] {
		if strings.HasPrefix(fix, "-") {
			forms = append(forms, prefix+processLexFix(parts[0], fix, "postfix"))
		}
	}
	return forms
}

func (d *scopeDeclaration) matches(raw []byte) bool {
	word := newBindingQuery(raw).word
	for _, form := range d.surfaceForms() {
		if form == word {
			return true
		}
	}
	return false
}

// prepareEnvironments runs before any binding. Import sites can introduce an
// environment even when their original statement disappears during expansion.
func (p *scopePlan) prepareEnvironments(occurrences []sourceOccurrence) error {
	p.environments = make(map[sourceParagraphKey]*lexicalScope)
	var builtin *lexicalScope
	for _, occurrence := range occurrences {
		unit := occurrence.Unit
		if unit != nil && unit.Kind == SourceBuiltin && unit.root == nil {
			unit.root = &lexicalScope{kind: lexicalBuiltin, unit: unit, records: make(map[string]*record)}
			builtin = unit.root
		}
	}
	for _, occurrence := range occurrences {
		unit := occurrence.Unit
		if unit == nil {
			continue
		}
		if unit.Kind == SourceProgram {
			unit.root = p.Root
			p.Root.unit = unit
			p.Root.parent = builtin
		}
		if unit.Kind == SourceDictionary && unit.root == nil {
			unit.root = &lexicalScope{kind: lexicalDictionary, unit: unit, parent: builtin, records: make(map[string]*record)}
		}
	}
	for _, occurrence := range occurrences {
		inherited := p.Root
		if occurrence.Unit != nil {
			inherited = occurrence.Unit.root
		}
		if occurrence.Parent != 0 {
			key := sourceParagraphKey{occurrence.Parent, occurrence.ImportSite.Path, occurrence.ImportSite.Paragraph}
			inherited = p.environments[key]
			if inherited == nil {
				return occurrence.ImportSite.wrapError(lerrors.Internal.Errorf("import has no enclosing lexical scope"))
			}
		}
		for _, paragraph := range occurrence.Paragraphs {
			scope := inherited
			if paragraph.Encapsulated {
				scope = &lexicalScope{kind: lexicalParagraph, unit: inherited.unit, parent: inherited, records: make(map[string]*record), source: SourceSpan{
					Path: occurrence.Path, Paragraph: paragraph.Number, Start: paragraph.FirstCode,
				}}
			}
			p.environments[sourceParagraphKey{occurrence.ID, occurrence.Path, paragraph.Number}] = scope
		}
	}
	return nil
}

func (p *scopePlan) environment(source *SourceStatement) *lexicalScope {
	if scope := p.environments[sourceParagraphKey{source.occurrence, source.Span.Path, source.Span.Paragraph}]; scope != nil {
		return scope
	}
	return p.Root // Unexpanded, unencapsulated source has the root environment.
}
