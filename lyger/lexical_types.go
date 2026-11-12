package Lyger

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/l-programming-language/l/lerrors"
)

// lexeme is the base structure of any Keyword in L
type lexeme struct {
	ID          string       `json:"id,omitempty"`
	Root        string       `json:"root,omitempty"`
	POS         string       `json:"pos,omitempty"`
	Type        string       `json:"type,omitempty"`
	Adverbs     []*token     `json:"adverbs,omitempty"`
	Adjectives  []*token     `json:"adjectives,omitempty"`
	Actions     []*lexeme    `json:"actions,omitempty"` // Used in complex Lexemes such as "Execute"
	Synchronous bool         `json:"synchronous,omitempty"`
	Forms       []lexemeForm `json:"forms,omitempty"`
	Value       string       `json:"value,omitempty"`
	Inputs      []string     `json:"inputs,omitempty"`
	Arguments   []string     `json:"arguments,omitempty"` // Per-use references; Inputs remains the source signature.
	Outputs     []string     `json:"outputs,omitempty"`
}

// A declared role owns its affix, even when another role uses the same affix.
// Keeping declaration order makes context-free same-spelling lookup stable.
type lexemeForm struct {
	Role   string `json:"role"`
	Fix    string `json:"fix"`
	Prefix bool   `json:"prefix,omitempty"`
}

type expandedLexemeForm struct{ Role, Spelling string }

func (form lexemeForm) spelling(root string) string {
	position := "postfix"
	if form.Prefix {
		position = "prefix"
	}
	return processLexFix(root, form.Fix, position)
}

func (l *lexeme) forms() []expandedLexemeForm {
	forms := make([]expandedLexemeForm, 1, len(l.Forms)+1)
	forms[0] = expandedLexemeForm{Role: "root", Spelling: l.Root}
	for _, form := range l.Forms {
		forms = append(forms, expandedLexemeForm{Role: form.Role, Spelling: form.spelling(l.Root)})
	}
	return forms
}

type declarationLabel struct {
	Root  string
	Slots []string // Includes the root at position zero, preserving authored case.
}

func parseDeclarationLabel(raw string) (declarationLabel, error) {
	if len(raw) < 2 || raw[0] != '"' || raw[len(raw)-1] != '"' {
		return declarationLabel{}, lerrors.Syntax.Errorf("lexeme label not declarative: %s", raw)
	}
	decoded, err := strconv.Unquote(raw)
	if err != nil {
		return declarationLabel{}, lerrors.Syntax.Errorf("invalid declaration label %s: %v", raw, err)
	}
	slots := strings.Split(decoded, "|")
	if slots[0] == "" || strings.IndexFunc(decoded, unicode.IsSpace) >= 0 || strings.ContainsAny(decoded, "\"\x00") {
		return declarationLabel{}, lerrors.Syntax.Errorf("declaration label requires a nonempty root and no spaces: %s", raw)
	}
	if err := validateAuthoredName([]byte(raw)); err != nil {
		return declarationLabel{}, err
	}
	return declarationLabel{Root: slots[0], Slots: slots}, nil
}

func lexicalCases(pos string) []string {
	switch pos {
	case "noun":
		return []string{"root", "plural", "descriptive", "active"}
	case "verb":
		return []string{"root", "active", "passive", "applicative", "objective"}
	default:
		return nil
	}
}

// validateInputs checks a lexeme use without changing its formal signature or
// positional arguments. Reference resolution and value types belong to the
// semantic instruction and definition-instantiation layers.
func (l *lexeme) validateInputs() error {
	if l == nil {
		return lerrors.Internal.Errorf("cannot validate a nil lexeme")
	}
	for i, input := range l.Inputs {
		if strings.TrimSpace(input) == "" {
			return lerrors.InvalidInput.Errorf("lexeme %q has an invalid formal input at position %d", l.Root, i+1)
		}
	}
	for i, argument := range l.Arguments {
		if strings.TrimSpace(argument) == "" {
			return lerrors.InvalidInput.Errorf("lexeme %q has an empty argument at position %d", l.Root, i+1)
		}
	}
	// A nil Arguments slice is an unbound prototype. Native lexemes without a
	// source signature are checked by instruction planning, which knows types.
	if l.Arguments != nil && (len(l.Inputs) > 0 || l.Value != "") && len(l.Arguments) != len(l.Inputs) {
		return lerrors.InvalidInput.Errorf("lexeme %q requires %d arguments, got %d", l.Root, len(l.Inputs), len(l.Arguments))
	}
	return nil
}

// construct validates the action graph. It does not fabricate a
// template from action names or resolve scoped references without their source.
func (l *lexeme) construct() error {
	active := make(map[*lexeme]bool)
	var validate func(*lexeme) error
	validate = func(current *lexeme) error {
		if current == nil {
			return lerrors.Internal.Errorf("action sequence contains a nil lexeme")
		}
		if active[current] {
			return lerrors.Cycle.Errorf("action cycle at lexeme %q", current.Root)
		}
		if err := current.validateInputs(); err != nil {
			return err
		}
		active[current] = true
		defer delete(active, current)
		for i, action := range current.Actions {
			if err := validate(action); err != nil {
				return fmt.Errorf("lexeme %q action %d: %w", current.Root, i+1, err)
			}
		}
		return nil
	}
	return validate(l)
}

// getValueType returns the full lexeme type, i.e. lexeme.noun
func (l *lexeme) getValueType() string {
	if l.Type != "" {
		return "lexeme." + l.POS + "." + l.Type
	}
	return "lexeme.generic"
}

// reLabel accepts ordinary binding names. Supplying cases explicitly always
// declares a complete family; declareFamily additionally rejects bare roots.
func (l *lexeme) reLabel(labelString string) error {
	return l.applyLabel(labelString, false)
}

func (l *lexeme) declareFamily(label string) error {
	return l.applyLabel(label, l.POS == "noun" || l.POS == "verb")
}

func (l *lexeme) applyLabel(labelString string, requireFamily bool) error {
	label, err := parseDeclarationLabel(labelString)
	if err != nil {
		return err
	}
	var forms []lexemeForm
	if len(label.Slots) > 1 {
		cases := lexicalCases(l.POS)
		if len(cases) == 0 {
			return lerrors.Syntax.Errorf("only nouns and verbs can declare forms, got %s", l.POS)
		}
		if len(label.Slots) > len(cases) {
			return lerrors.Syntax.Errorf("%s family %q requires %d forms (%s), got %d", l.POS, label.Root, len(cases), strings.Join(cases, "|"), len(label.Slots))
		}
		for index, fix := range label.Slots[1:] {
			if strings.HasPrefix(fix, "-") {
				forms = append(forms, lexemeForm{Role: cases[index+1], Fix: strings.ToLower(fix)})
			}
		}
	}
	candidate := *l
	candidate.Root, candidate.Forms = strings.ToLower(label.Root), forms
	if len(label.Slots) > 1 || requireFamily {
		if err := candidate.validateFamily(); err != nil {
			return err
		}
	}
	l.Root, l.Forms = candidate.Root, candidate.Forms
	return nil
}

// validateFamily also checks inherited canonical families without inventing a
// new label. Form spellings may coincide, but the entire family cannot be root.
func (l *lexeme) validateFamily() error {
	cases := lexicalCases(l.POS)
	if len(cases) == 0 || l.Root == "" {
		return lerrors.Syntax.Errorf("complete noun or verb family requires a nonempty root, got %s %q", l.POS, l.Root)
	}
	if !validFamilySpelling(l.Root) {
		return lerrors.Syntax.Errorf("%s family has invalid root %q", l.POS, l.Root)
	}
	roles := make(map[string]int, len(l.Forms))
	distinct := false
	for _, form := range l.Forms {
		if (!form.Prefix && !strings.HasPrefix(form.Fix, "-")) || (form.Prefix && !strings.HasSuffix(form.Fix, "-")) || strings.Count(form.Fix, "-")-1 > len(l.Root) {
			return lerrors.Syntax.Errorf("%s family %q has invalid %s affix %q", l.POS, l.Root, form.Role, form.Fix)
		}
		spelling := form.spelling(l.Root)
		if !validFamilySpelling(spelling) {
			return lerrors.Syntax.Errorf("%s family %q has invalid expanded %s form %q", l.POS, l.Root, form.Role, spelling)
		}
		roles[form.Role]++
		distinct = distinct || !strings.EqualFold(l.Root, spelling)
	}
	var missing []string
	for _, role := range cases[1:] {
		if roles[role] != 1 {
			missing = append(missing, role)
		}
	}
	if len(l.Forms) != len(cases)-1 || len(missing) != 0 {
		return lerrors.Syntax.Errorf("%s family %q requires %d forms (%s); missing or repeated forms: %s", l.POS, l.Root, len(cases), strings.Join(cases, "|"), strings.Join(missing, ", "))
	}
	if !distinct {
		return lerrors.Syntax.Errorf("%s family %q requires at least one nonroot form differing from its root", l.POS, l.Root)
	}
	return nil
}

func validFamilySpelling(spelling string) bool {
	// A declared case must survive token loading and binding as its own word.
	// Comment delimiters, conjunction commas and possessive suffixes cannot name cases.
	if strings.Contains(spelling, "***") {
		return false
	}
	_, possessive := possessiveOwner(spelling)
	return !possessive && !strings.HasSuffix(spelling, ",") && spelling != "" && utf8.ValidString(spelling) && !strings.ContainsAny(spelling, ".\"`|") && strings.IndexFunc(spelling, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }) < 0
}

// loadSyncActions keeps sibling actions in source order. Cloning prevents a
// declaration from taking ownership of another definition's mutable use state.
func (l *lexeme) loadSyncActions(actionSequence []*lexeme) error {
	if l == nil {
		return lerrors.Internal.Errorf("cannot load actions into a nil lexeme")
	}
	if len(actionSequence) == 0 {
		return lerrors.InvalidInput.Errorf("synchronous action sequence is empty")
	}
	// Check graph structure before cloneUse, which recursively copies children.
	active := map[*lexeme]bool{l: true}
	var validate func(*lexeme) error
	validate = func(action *lexeme) error {
		if action == nil {
			return lerrors.Internal.Errorf("action sequence contains a nil lexeme")
		}
		if active[action] {
			return lerrors.Cycle.Errorf("action cycle at lexeme %q", action.Root)
		}
		active[action] = true
		defer delete(active, action)
		for _, child := range action.Actions {
			if err := validate(child); err != nil {
				return err
			}
		}
		return nil
	}
	for _, action := range actionSequence {
		if err := validate(action); err != nil {
			return err
		}
	}
	actions := make([]*lexeme, len(actionSequence))
	for i, action := range actionSequence {
		actions[i] = action.cloneUse()
	}
	l.Actions, l.Synchronous = actions, true
	return nil
}

// loadActions records the existing sequence policy separately from action order.
func (l *lexeme) loadActions(actionSequence []*lexeme) error {
	for _, action := range actionSequence {
		if action == nil {
			return lerrors.Internal.Errorf("action sequence contains a nil lexeme")
		}
		for _, adverb := range action.Adverbs {
			if adverb == nil {
				return lerrors.Internal.Errorf("lexeme %q has a nil action adverb", action.Root)
			}
		}
	}
	seqType := "sync"
	var err error
	if len(actionSequence) > 1 {
		seqType, err = checkActions(actionSequence)
		if err != nil {
			return err
		}
	}
	if err = l.loadSyncActions(actionSequence); err != nil {
		return err
	}
	l.Synchronous = seqType == "sync"
	return nil
}

// bindArguments replaces per-use source references. Repeated extraction is
// idempotent and cannot append invocation values to a definition's formals.
func (l *lexeme) bindArguments(inputs []string) error {
	if l == nil {
		return lerrors.Internal.Errorf("cannot load arguments into a nil lexeme")
	}
	for i, input := range inputs {
		if strings.TrimSpace(input) == "" {
			return lerrors.InvalidInput.Errorf("lexeme %q has an empty argument at position %d", l.Root, i+1)
		}
	}
	l.Arguments = append(make([]string, 0, len(inputs)), inputs...)
	return nil
}

// parseInputs
func (l *lexeme) parseInputs() error {
	defParser := newParser([]byte(l.Value))
	inputsBytes, err := defParser.parse(parseInputs)
	if err != nil {
		return err
	}
	// Identify Definition Inputs
	for _, inBytes := range inputsBytes {
		cBytes := bytes.Replace(bytes.Replace(inBytes, []byte("<@"), []byte(""), 1), []byte("@>"), []byte(""), 1)
		l.Inputs = append(l.Inputs, string(cBytes))
	}
	return nil
}

// define a Lexeme from a definitional sequence when creating new user defined lexemes
func (l *lexeme) define(definition []byte) error {
	// Check and clean definition
	frame, framed := splitDefinitionFrame(string(definition))
	if !framed {
		return lerrors.Syntax.Errorf("%s", "invalid lexeme definition: "+string(definition))
	}
	l.Value = frame.Body
	if frame.Quoted {
		l.Value = "`" + l.Value + "`"
	}
	err := l.parseInputs()
	if err != nil {
		return err
	}
	// Identify Definition Outputs, "VALUE.{GLYPH_TYPE}" is the default, may need to add to this part eventually
	defaultOutput := "Value|" + determineGlyphType(l.Value)
	l.Outputs = append(l.Outputs, defaultOutput)
	l.Type = "generic." + strings.ToLower(determineGlyphType(l.Value))
	return nil
}

// matchesForm checks one declared form without applying context-free root
// precedence. Grammar uses it only after lexical identity has been resolved.
func (l *lexeme) matchesForm(word, form string) bool {
	if form == "root" {
		return word == l.Root
	}
	for _, candidate := range l.Forms {
		if candidate.Role == form && word == candidate.spelling(l.Root) {
			return true
		}
	}
	return false
}

// match checks if an inputted token byte slice matches the lexeme, if a match returns root/case
func (l *lexeme) match(tBytes []byte) (bool, string) {
	// 1) check for root match
	if l.Root == string(tBytes) {
		return true, "root"
	}
	for _, form := range l.Forms {
		if string(tBytes) == form.spelling(l.Root) {
			return true, form.Role
		}
	}
	return false, ""
}

// newLexeme returns a new valueless lexeme struct initializes for an inputted type
func newLexeme(lType string) *lexeme {
	return &lexeme{
		POS:        lType,
		Type:       lType,
		Adverbs:    []*token{},
		Adjectives: []*token{},
		Actions:    []*lexeme{},
		Value:      "",
		Inputs:     []string{},
		Outputs:    []string{},
	}
}
