package Lyger

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// lexeme is the base structure of any Keyword in L
type lexeme struct {
	ID          string            `json:"id,omitempty"`
	Root        string            `json:"root,omitempty"`
	POS         string            `json:"pos,omitempty"`
	Type        string            `json:"type,omitempty"`
	Adverbs     []*token          `json:"adverbs,omitempty"`
	Adjectives  []*token          `json:"adjectives,omitempty"`
	Actions     []*lexeme         `json:"actions,omitempty"` // Used in complex Lexemes such as "Execute"
	Synchronous bool              `json:"synchronous,omitempty"`
	Prefixes    map[string]string `json:"prefixes,omitempty"`
	Postfixes   map[string]string `json:"postfixes,omitempty"`
	Value       string            `json:"value,omitempty"`
	Inputs      []string          `json:"inputs,omitempty"`
	Arguments   []string          `json:"arguments,omitempty"` // Per-use references; Inputs remains the source signature.
	Outputs     []string          `json:"outputs,omitempty"`
}

// validateInputs checks a legacy use without changing its formal signature or
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

// construct retains the legacy validation boundary. It does not fabricate a
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

// buildLexCases inputs a split label string and lexical POS and returns the root word and PostFix/Prefix Maps
func (l *lexeme) buildLexCases(splitLabel []string, cases []string) error {
	var root string
	postRoot := false
	post := make(map[string]string)
	pre := make(map[string]string)
	if len(splitLabel) > len(cases) {
		return lerrors.Syntax.Errorf("invalid number of cases on lexeme label for part of speech type")
	}
	for ind, caseLabel := range splitLabel {
		caseStr := cases[ind]
		if caseStr == "root" {
			root = caseLabel
			postRoot = true
		} else {
			if postRoot { // IF PAST THE ROOT CASE WE ARE ADDING POSTFIXES
				if strings.HasPrefix(caseLabel, "-") { // TO BE ADDED POSTFIX MUST START WITH one or more -
					post[caseLabel] = caseStr
				}
			} else { // IF BEFORE THE ROOT CASE WE ARE ADDING PREFIXES
				if strings.HasSuffix(caseLabel, "-") { // TO BE ADDED POSTFIX MUST END WITH one or more -
					pre[caseLabel] = caseStr
				}
			}
		}
	}
	l.Root = strings.ToLower(root)
	l.Postfixes = post
	l.Prefixes = pre
	return nil
}

// loadLabelCases inputs a cased label string and updates lexeme Prefixes, and Postfixes
// NOUN:  "Process|-es|-ive|-ing" = ROOT|PLURAL|DESCRIPTIVE|ACTIVE
// VERB:  "Log|-s|-ged|-ging|-ger = ROOT|ACTIVE|PASSIVE|APPLICATIVE|OBJECTIVE
func (l *lexeme) loadLabelCases(labelString string) error {
	nounCases := []string{"root", "plural", "descriptive", "active"}               //TODO LATER ADD PREFIXES BEFORE ROOT
	verbCases := []string{"root", "active", "passive", "applicative", "objective"} //TODO LATER DITTO ABOVE
	if strings.Contains(labelString, "|") {
		sLabel := strings.Split(labelString, "|")
		if l.POS == "noun" {
			err := l.buildLexCases(sLabel, nounCases)
			if err != nil {
				return err
			}
			return nil
		} else if l.POS == "verb" {
			err := l.buildLexCases(sLabel, verbCases)
			if err != nil {
				return err
			}
			return nil
		}
		return lerrors.Syntax.Errorf("%s", "lexeme label error: unable to case a"+l.POS+", only nouns and verbs can be cased")
	}
	return lerrors.Syntax.Errorf("%s", "lexeme label is not cased: "+labelString)
}

// reLabel inputs a label string and updates lexeme label, which comprises Root, Prefixes, and Postfixes
func (l *lexeme) reLabel(labelString string) error {
	if !strings.HasPrefix(labelString, `"`) && !strings.HasSuffix(labelString, `"`) {
		return lerrors.Syntax.Errorf("%s", "lexeme label not declarative: "+labelString)
	}
	labelString = strings.Replace(labelString, `"`, ``, -1)
	if strings.Contains(labelString, "|") {
		// needs lex label cases processed, i.e. Execute|-s|-d|--ing|--or
		err := l.loadLabelCases(labelString)
		if err != nil {
			return err
		}
	} else {
		// root label only, i.e. Objective
		l.Postfixes = make(map[string]string)
		l.Prefixes = make(map[string]string)
		l.Root = strings.ToLower(labelString)
	}
	return nil
}

// checkDescriptors scans a lexeme for a particular adverb or adjective
//
//	usage example: descType = "adverb", descID = "progression"
func (l *lexeme) checkDescriptors(descType string, descID string) bool {
	switch descType {
	case "adjective":
		descID = "adjectives." + descID
		for _, adj := range l.Adjectives {
			if adj.tType == descID {
				return true
			}
		}
	case "adverb":
		descID = "adverbs." + descID
		for _, adv := range l.Adverbs {
			if adv.tType == descID {
				return true
			}
		}
	}
	return false
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

// match checks if an inputted token byte slice matches the lexeme, if a match returns root/case
func (l *lexeme) match(tBytes []byte) (bool, string) {
	// 1) check for root match
	if l.Root == string(tBytes) {
		return true, "root"
	}
	// 2) check for postfix match
	for pKey, pType := range l.Postfixes {
		if string(tBytes) == processLexFix(l.Root, pKey, "postfix") {
			return true, pType
		}
	}
	// 3) check for prefix match
	for pKey, pType := range l.Prefixes {
		if string(tBytes) == processLexFix(l.Root, pKey, "prefix") {
			return true, pType
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
		Prefixes:   make(map[string]string),
		Postfixes:  make(map[string]string),
		Value:      "",
		Inputs:     []string{},
		Outputs:    []string{},
	}
}
