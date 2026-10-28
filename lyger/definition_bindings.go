package Lyger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/l-programming-language/l/lerrors"
)

const definitionMarker = "<DEF>"

// Preserve native string data through logeme's markup and content processing.
var logemeTextEscaper = strings.NewReplacer(
	"|", `\x7c`,
	"~", `\x7e`,
	"  ", `\x20\x20`,
	"(/", `(\x2f`,
)

// definitionFrame separates source framing from the still-encoded body.
// Empty or nested bodies remain the semantic parser's responsibility.
type definitionFrame struct {
	Body   string
	Quoted bool
}

func splitDefinitionFrame(raw string) (definitionFrame, bool) {
	frame := definitionFrame{}
	body := raw
	if strings.HasPrefix(body, "`") {
		if len(body) < 2 || !strings.HasSuffix(body, "`") {
			return frame, false
		}
		frame.Quoted = true
		body = body[1 : len(body)-1]
	}
	width := len(definitionMarker)
	if len(body) < 2*width || !strings.HasPrefix(body, definitionMarker) || !strings.HasSuffix(body, definitionMarker) {
		return frame, false
	}
	frame.Body = body[width : len(body)-width]
	return frame, true
}

// DefinitionSlot describes a named argument in source order. An empty Type is
// a reference to an already instantiated definition, not an implicit Text.
type DefinitionSlot struct {
	Name string
	Type string
}

type DefinitionProperty struct {
	Key   string
	Type  string
	Value *DefinitionArgument
}

// DefinitionArgument separates source text from generated expressions and
// preserves collection cardinality. Text contains the unquoted scalar value.
// Syllables and Definition contain expressions that were already instantiated.
type DefinitionArgument struct {
	Kind       string
	Text       string
	Syllables  []string
	Props      []DefinitionProperty
	Texts      []string
	Definition *DefinitionInstance
	reference  string
	atoms      *[]DefinitionArgument
	typeAtom   string
	index      *uint64
	sequence   []DefinitionInstance
}

type DefinitionBinding struct {
	Slot    DefinitionSlot
	Kind    string
	Emitted string
	Sources []SourceSpan
}

// DefinitionInstance owns fresh output and an argument ledger. Expression is
// a glyph expression; Words preserves an explicit word/sentence grouping.
type DefinitionInstance struct {
	Symbol             string  // source-defined spelling, retained for diagnostics
	definition         *record // canonical nominal identity, set by semantic instantiation
	Original           []byte
	Span               SourceSpan
	OriginalExpression string
	Expression         string
	Words              [][]string
	Bindings           []DefinitionBinding
}

type definitionPart struct {
	text string
	slot *DefinitionSlot
}

// DefinitionTemplate retains the raw source prototype. Instantiation uses its
// parsed private parts and never substitutes into or updates the prototype.
type DefinitionTemplate struct {
	Original       []byte
	Span           SourceSpan
	Slots          []DefinitionSlot
	parts          []definitionPart
	words          [][][]definitionPart
	function       bool
	signatureSlots map[string]string
}

// ParseDefinitionTemplate accepts a raw source literal or definitionMarker-framed content.
// Markerless content is accepted for the existing lexeme.Value representation.
func ParseDefinitionTemplate(raw []byte, span SourceSpan) (*DefinitionTemplate, error) {
	template := &DefinitionTemplate{Original: bytes.Clone(raw), Span: span}
	body := strings.TrimSpace(string(raw))
	frame, framed := splitDefinitionFrame(body)
	if strings.HasPrefix(body, "`") {
		if len(body) < 2 || !strings.HasSuffix(body, "`") {
			return nil, template.failure(lerrors.Syntax.Errorf("unterminated definition literal"))
		}
		body = body[1 : len(body)-1]
	}
	if strings.Contains(body, definitionMarker) {
		if !framed {
			return nil, template.failure(lerrors.Syntax.Errorf("definition requires paired %s markers", definitionMarker))
		}
		body = frame.Body
		if strings.Contains(body, definitionMarker) {
			return nil, template.failure(lerrors.Syntax.Errorf("nested %s markers", definitionMarker))
		}
	}
	body = decodeDefinitionEscapes(body)
	if strings.TrimSpace(body) == "" {
		return nil, template.failure(lerrors.Syntax.Errorf("empty definition"))
	}
	var err error
	if strings.HasPrefix(strings.TrimSpace(body), "[") {
		var elements []json.RawMessage
		if err := json.Unmarshal([]byte(body), &elements); err != nil {
			return nil, template.failure(lerrors.Syntax.Errorf("definition grouping: %w", err))
		}
		if len(elements) == 0 {
			return nil, template.failure(lerrors.Syntax.Errorf("empty definition grouping"))
		}
		var words [][]string
		if bytes.HasPrefix(bytes.TrimSpace(elements[0]), []byte("[")) {
			err = json.Unmarshal([]byte(body), &words)
		} else {
			var word []string
			err = json.Unmarshal([]byte(body), &word)
			words = [][]string{word}
		}
		if err != nil {
			return nil, template.failure(lerrors.Syntax.Errorf("definition grouping must contain strings or words: %w", err))
		}
		for _, word := range words {
			if len(word) == 0 {
				return nil, template.failure(lerrors.Syntax.Errorf("empty word in definition grouping"))
			}
			var parsedWord [][]definitionPart
			for _, syllable := range word {
				parts, err := template.parseParts(syllable)
				if err != nil {
					return nil, err
				}
				parsedWord = append(parsedWord, parts)
			}
			template.words = append(template.words, parsedWord)
		}
	} else {
		if start := strings.Index(body, "{|"); start >= 0 && !strings.ContainsAny(body[:start], "^*(") {
			if end, err := definitionMetadataEnd(body, start); err == nil {
				meta := strings.TrimSpace(body[start+2 : end])
				template.function = strings.HasPrefix(meta, "Func{") && strings.HasSuffix(meta, "}Func")
				if template.function {
					template.signatureSlots = make(map[string]string)
					fields, err := splitDefinitionFields(meta[5 : len(meta)-5])
					if err != nil {
						return nil, template.failure(err)
					}
					for _, field := range fields {
						key, value, ok := strings.Cut(field, ":")
						value = strings.TrimSpace(value)
						if ok && strings.HasPrefix(value, "<@") && strings.HasSuffix(value, "@>") {
							name, _, _ := strings.Cut(value[2:len(value)-2], "|")
							template.signatureSlots[name] = strings.TrimSpace(key)
						}
					}
				}
			}
		}
		template.parts, err = template.parseParts(body)
		if err != nil {
			return nil, err
		}
	}
	return template, nil
}

// IsFunction identifies a root Func metadata structure, independent of the
// source-defined word's name or grammatical part of speech.
func (t *DefinitionTemplate) IsFunction() bool { return t.function }

// Signature values bind the source slots occupying Func fields. A second Text
// slot does not silently acquire the callable's name merely because its type fits.
func (t *DefinitionTemplate) signatureArguments(name string, inputs []DefinitionProperty, outputs []string) (map[string]DefinitionArgument, error) {
	values := map[string]DefinitionArgument{
		"Key": {Kind: "Text", Text: name}, "Inputs": {Kind: "Props", Props: inputs}, "Outputs": {Kind: "Texts", Texts: outputs},
	}
	bindings := make(map[string]DefinitionArgument)
	for _, slot := range t.Slots {
		value, exists := values[t.signatureSlots[slot.Name]]
		if !exists {
			continue
		}
		if slot.Type != value.Kind {
			return nil, lerrors.Unsupported.Errorf("unsupported function slot %s: %s", slot.Name, slot.Type)
		}
		bindings[slot.Name] = value
	}
	return bindings, nil
}

func (t *DefinitionTemplate) failure(err error) error {
	if err == nil {
		return nil
	}
	err = fmt.Errorf("definition: %w", err)
	if t.Span.Path != "" {
		return t.Span.wrapError(err)
	}
	return err
}

// Only the source's quote/backslash escapes are decoded here. Native expression
// escapes such as \n remain text in the template, and arguments are not decoded
// again when instantiated.
func decodeDefinitionEscapes(source string) string {
	var result strings.Builder
	for i := 0; i < len(source); i++ {
		if source[i] == '\\' && i+1 < len(source) && (source[i+1] == '"' || source[i+1] == '\\' || source[i+1] == '`') {
			i++
		}
		result.WriteByte(source[i])
	}
	return result.String()
}

func (t *DefinitionTemplate) parseParts(source string) ([]definitionPart, error) {
	var parts []definitionPart
	for {
		start := strings.Index(source, "<@")
		if start < 0 {
			if strings.Contains(source, "@>") {
				return nil, t.failure(lerrors.Syntax.Errorf("slot closing marker has no opening marker"))
			}
			parts = append(parts, definitionPart{text: source})
			return parts, nil
		}
		parts = append(parts, definitionPart{text: source[:start]})
		end := strings.Index(source[start+2:], "@>")
		if end < 0 {
			return nil, t.failure(lerrors.Syntax.Errorf("unterminated parameter slot"))
		}
		end += start + 2
		fields := strings.Split(source[start+2:end], "|")
		if len(fields) > 2 || !definitionName(fields[0]) {
			return nil, t.failure(lerrors.Syntax.Errorf("invalid parameter slot %q", source[start:end+2]))
		}
		slot := DefinitionSlot{Name: fields[0]}
		if len(fields) == 2 {
			if !definitionName(fields[1]) {
				return nil, t.failure(lerrors.Syntax.Errorf("invalid type for parameter %q", slot.Name))
			}
			slot.Type = fields[1]
		}
		found := false
		for _, prior := range t.Slots {
			if prior.Name == slot.Name {
				if prior.Type != slot.Type {
					return nil, t.failure(lerrors.TypeMismatch.Errorf("parameter %q has conflicting types", slot.Name))
				}
				found = true
			}
		}
		if !found {
			t.Slots = append(t.Slots, slot)
		}
		parts = append(parts, definitionPart{slot: &slot})
		source = source[end+2:]
	}
}

func definitionName(name string) bool {
	for _, part := range strings.Split(name, ".") {
		if part == "" {
			return false
		}
		for index, char := range part {
			if char != '_' && !unicode.IsLetter(char) && (index == 0 || !unicode.IsDigit(char)) {
				return false
			}
		}
	}
	return true
}

// Instantiate binds every slot by name. Values are emitted only at parsed slot
// locations, so a replacement never reinterprets argument text as another slot.
func (t *DefinitionTemplate) Instantiate(bindings map[string]DefinitionArgument) (DefinitionInstance, error) {
	instance := DefinitionInstance{Original: bytes.Clone(t.Original), Span: t.Span}
	values := make(map[string]string, len(t.Slots))
	for _, slot := range t.Slots {
		argument, ok := bindings[slot.Name]
		if !ok {
			return DefinitionInstance{}, t.failure(lerrors.InvalidInput.Errorf("missing parameter %q", slot.Name))
		}
		value, err := argument.renderFor(slot)
		if err != nil {
			return DefinitionInstance{}, t.failure(fmt.Errorf("parameter %q: %w", slot.Name, err))
		}
		values[slot.Name] = value
		binding := DefinitionBinding{Slot: slot, Kind: argument.Kind, Emitted: value}
		for _, child := range argument.sequence {
			binding.Sources = append(binding.Sources, child.Span)
		}
		instance.Bindings = append(instance.Bindings, binding)
	}
	render := func(parts []definitionPart) (string, string, error) {
		var result strings.Builder
		for _, part := range parts {
			if part.slot == nil {
				result.WriteString(part.text)
			} else {
				result.WriteString(values[part.slot.Name])
			}
		}
		original := result.String()
		normalized, err := normalizeDefinitionFunctions(original)
		if err == nil {
			normalized, err = normalizeDefinitionGroups(normalized)
		}
		if err == nil {
			err = validateDefinitionExpression(normalized)
		}
		return original, normalized, err
	}
	if t.words != nil {
		for _, word := range t.words {
			var rendered []string
			for _, parts := range word {
				_, expression, err := render(parts)
				if err != nil {
					return DefinitionInstance{}, t.failure(err)
				}
				rendered = append(rendered, expression)
			}
			instance.Words = append(instance.Words, rendered)
		}
	} else {
		original, expression, err := render(t.parts)
		if err != nil {
			return DefinitionInstance{}, t.failure(err)
		}
		instance.OriginalExpression, instance.Expression = original, expression
	}
	return instance, nil
}

func (argument DefinitionArgument) renderFor(slot DefinitionSlot) (string, error) {
	expected, actual := strings.ToLower(slot.Type), strings.ToLower(argument.Kind)
	kind, _ := parseDefinitionArgKindFolded(argument.Kind)
	if expected == "" || expected == "syllable" {
		if argument.atoms != nil {
			if argument.Text != "" || argument.reference != "" || argument.typeAtom != "" || argument.index != nil || argument.Definition != nil || len(argument.sequence) != 0 || len(argument.Syllables) != 0 || len(argument.Props) != 0 || len(argument.Texts) != 0 {
				return "", lerrors.TypeMismatch.Errorf("atom list cannot contain a definition or another value")
			}
			values := make([]string, len(*argument.atoms))
			for index, atom := range *argument.atoms {
				var err error
				values[index], err = atom.metadataAtom()
				if err != nil {
					return "", err
				}
			}
			return strings.Join(values, ", "), nil
		}
		if argument.typeAtom != "" {
			return argument.metadataAtom()
		}
		if argument.index != nil {
			return argument.propertyAtom()
		}
		if argument.reference != "" {
			return argument.referenceAtom()
		}
		if len(argument.sequence) != 0 {
			var expressions []string
			for _, child := range argument.sequence {
				if child.Expression == "" {
					return "", lerrors.TypeMismatch.Errorf("composition sequence requires instantiated syllables")
				}
				expressions = append(expressions, child.Expression)
			}
			return strings.Join(expressions, "*"), nil
		}
		if argument.Definition != nil && argument.Definition.Expression != "" {
			return argument.Definition.Expression, nil
		}
		if kind == definitionArgSyllable && argument.Text != "" {
			return argument.Text, nil
		}
		return "", lerrors.TypeMismatch.Errorf("requires an instantiated syllable definition")
	}
	if expected == "thing" {
		switch kind {
		case definitionArgText, definitionArgNumber, definitionArgTruth, definitionArgNull:
			expected = actual
		default:
			return "", lerrors.TypeMismatch.Errorf("Thing requires a primitive value, got %q", argument.Kind)
		}
	}
	// Preserve distinct unknown-name errors before enum dispatch: two unrelated
	// source types must not compare equal merely because both are unrecognized.
	if expected != actual {
		return "", lerrors.TypeMismatch.Errorf("requires %s, got %s", slot.Type, argument.Kind)
	}
	if argument.reference != "" {
		return argument.referenceAtom()
	}
	switch kind {
	case definitionArgText:
		return definitionTextLiteral(argument.Text), nil
	case definitionArgNumber:
		number, err := strconv.ParseFloat(argument.Text, 64)
		if err != nil {
			return "", lerrors.Syntax.Wrap(err, fmt.Sprintf("invalid number %q", argument.Text))
		}
		literal, finite := nativeNumberLiteral(number)
		if !finite {
			return "", lerrors.Syntax.Errorf("invalid number %q", argument.Text)
		}
		return literal, nil
	case definitionArgTruth:
		if argument.Text != "true" && argument.Text != "false" {
			return "", lerrors.Syntax.Errorf("invalid truth value %q", argument.Text)
		}
		return argument.Text, nil
	case definitionArgNull:
		return "()", nil
	case definitionArgSyllables:
		if len(argument.Syllables) == 0 {
			return "", lerrors.InvalidInput.Errorf("requires at least one syllable")
		}
		for _, expression := range argument.Syllables {
			if err := validateDefinitionExpression(expression); err != nil {
				return "", err
			}
		}
		if len(argument.Syllables) == 1 {
			return argument.Syllables[0], nil
		}
		return "(" + strings.Join(argument.Syllables, "*") + ")", nil
	case definitionArgProps:
		var props []string
		for _, property := range argument.Props {
			if !validDefinitionNativeName(property.Key) || property.Type != "" && !validDefinitionNativeType(property.Type) || property.Type == "" && property.Value == nil {
				return "", lerrors.Syntax.Errorf("invalid property %q of type %q", property.Key, property.Type)
			}
			fields := []string{"Key: " + property.Key}
			if property.Type != "" {
				fields = append(fields, "Type: "+property.Type)
			}
			if property.Value != nil {
				value, err := property.Value.propertyAtom()
				if err != nil {
					return "", fmt.Errorf("property %s: %w", property.Key, err)
				}
				fields = append(fields, "Value: "+value)
			}
			props = append(props, "Prop{"+strings.Join(fields, ", ")+"}Prop")
		}
		return "&[" + strings.Join(props, ", ") + "]&", nil
	case definitionArgTexts:
		// Text lists are retained as JSON strings until a consuming structure
		// supplies their role (for example result types in Func metadata).
		values := append([]string{}, argument.Texts...)
		encoded, err := json.Marshal(values)
		return string(encoded), lerrors.Internal.Wrap(err, "")
	default:
		return "", lerrors.Unsupported.Errorf("unsupported parameter type %q", slot.Type)
	}
}

// A reference names a previously bound value. Calls and value construction must
// be supplied by the selected definition, never manufactured in an argument.
// The selected definition controls child grouping; this argument retains the
// ordered, already-instantiated children and their source spans in the ledger.
func definitionSequenceArgument(children []DefinitionInstance) DefinitionArgument {
	return DefinitionArgument{Kind: "Definition", sequence: append([]DefinitionInstance(nil), children...)}
}

func definitionReferenceArgument(name string) DefinitionArgument {
	return DefinitionArgument{Kind: definitionArgSyllable.String(), reference: name}
}

func definitionAtomListArgument(values ...DefinitionArgument) DefinitionArgument {
	atoms := append([]DefinitionArgument{}, values...)
	return DefinitionArgument{Kind: definitionArgSyllable.String(), atoms: &atoms}
}

// This constructor accepts a resolved type spelling, never an expression. Type
// syntax remains native metadata and retains the existing framing validation.
func definitionTypeArgument(name string) DefinitionArgument {
	return DefinitionArgument{Kind: definitionArgSyllable.String(), typeAtom: name}
}

func definitionIndexArgument(index uint64) DefinitionArgument {
	return DefinitionArgument{Kind: definitionArgSyllable.String(), index: &index}
}

func (argument DefinitionArgument) metadataAtom() (string, error) {
	if argument.typeAtom != "" {
		if !validDefinitionNativeType(argument.typeAtom) || argument.reference != "" || argument.Text != "" || argument.Definition != nil || len(argument.sequence) != 0 || argument.atoms != nil || argument.index != nil || len(argument.Syllables) != 0 || len(argument.Props) != 0 || len(argument.Texts) != 0 {
			return "", lerrors.TypeMismatch.Errorf("type atom requires one resolved type")
		}
		return argument.typeAtom, nil
	}
	return argument.propertyAtom()
}

func (argument DefinitionArgument) referenceAtom() (string, error) {
	kind, _ := parseDefinitionArgKindFolded(argument.Kind)
	if kind != definitionArgSyllable && kind != definitionArgText && kind != definitionArgNumber && kind != definitionArgTruth && kind != definitionArgNull || argument.reference == "" || argument.Text != "" || argument.Definition != nil || len(argument.sequence) != 0 || argument.atoms != nil || argument.typeAtom != "" || argument.index != nil || len(argument.Syllables) != 0 || len(argument.Props) != 0 || len(argument.Texts) != 0 {
		return "", lerrors.TypeMismatch.Errorf("reference cannot contain a value or definition")
	}
	for _, path := range strings.Split(argument.reference, "::") {
		for index, member := range strings.Split(path, ".") {
			if validDefinitionNativeName(member) {
				continue
			}
			if index > 0 && member != "" {
				if _, err := strconv.ParseUint(member, 10, 32); err == nil {
					continue
				}
			}
			return "", lerrors.Syntax.Errorf("invalid value reference %q", argument.reference)
		}
	}
	return argument.reference, nil
}

func (argument DefinitionArgument) propertyAtom() (string, error) {
	if argument.reference != "" {
		return argument.referenceAtom()
	}
	if argument.Definition != nil || len(argument.sequence) != 0 || argument.atoms != nil || argument.typeAtom != "" || len(argument.Syllables) != 0 || len(argument.Props) != 0 || len(argument.Texts) != 0 {
		return "", lerrors.TypeMismatch.Errorf("property value cannot contain a definition or collection")
	}
	if argument.index != nil {
		if argument.Kind != definitionArgSyllable.String() || argument.Text != "" {
			return "", lerrors.TypeMismatch.Errorf("index atom requires one unsigned value")
		}
		return strconv.FormatUint(*argument.index, 10), nil
	}
	kind, _ := parseDefinitionArgKindFolded(argument.Kind)
	switch kind {
	case definitionArgText, definitionArgNumber, definitionArgTruth:
		return argument.renderFor(DefinitionSlot{Type: argument.Kind})
	case definitionArgNull:
		return "()", nil
	default:
		return "", lerrors.TypeMismatch.Errorf("property value requires a literal or bound reference")
	}
}

func validDefinitionNativeName(value string) bool {
	raw := strings.HasPrefix(value, "r#")
	name := strings.TrimPrefix(value, "r#")
	if name == "" || raw && name == "_" {
		return false
	}
	for index, r := range name {
		start := r == '_' || unicode.IsLetter(r) || unicode.In(r, unicode.Nl, unicode.Other_ID_Start)
		if !start && (index == 0 || !(unicode.IsDigit(r) || unicode.IsMark(r) || unicode.In(r, unicode.Pc, unicode.Other_ID_Continue))) {
			return false
		}
	}
	if raw {
		return name != "crate" && name != "self" && name != "Self" && name != "super"
	}
	// Rust 2024 strict/reserved keywords are not native function/property names.
	// Weak keywords remain contextual; rustc owns final native legality.
	switch name {
	case "as", "break", "const", "continue", "crate", "else", "enum", "extern", "false", "fn", "for", "if", "impl", "in", "let", "loop", "match", "mod", "move", "mut", "pub", "ref", "return", "self", "Self", "static", "struct", "super", "trait", "true", "type", "unsafe", "use", "where", "while", "async", "await", "dyn", "abstract", "become", "box", "do", "final", "gen", "macro", "override", "priv", "typeof", "unsized", "virtual", "yield", "try":
		return false
	}
	return true
}

// Native type interiors remain opaque. The frontend validates the typed slot
// and framing; rustc owns native syntax, type resolution and ownership legality.
func validDefinitionNativeType(value string) bool {
	return strings.TrimSpace(value) != "" && !strings.ContainsAny(value, "\x00\r\n") && !strings.Contains(value, "|}")
}

// definitionMetadataEnd preserves Rust lifetimes while skipping native literals.
func definitionMetadataEnd(source string, start int) (int, error) {
	for i := start + 2; i < len(source); i++ {
		end, literal, err := definitionNativeLiteralEnd(source, i)
		if err != nil {
			return 0, err
		}
		if literal {
			i = end - 1
			continue
		}
		if strings.HasPrefix(source[i:], "|}") {
			return i, nil
		}
	}
	return 0, lerrors.Syntax.Errorf("unterminated glyph metadata")
}

// Normalize the intrinsic Func{...}Func metadata form, regardless of the name
// of the source-defined word containing it. Native metadata is left intact.
func normalizeDefinitionFunctions(source string) (string, error) {
	var result strings.Builder
	for {
		start := strings.Index(source, "{|")
		if start < 0 {
			result.WriteString(source)
			return result.String(), nil
		}
		end, err := definitionMetadataEnd(source, start)
		if err != nil {
			return "", err
		}
		result.WriteString(source[:start+2])
		meta := strings.TrimSpace(source[start+2 : end])
		if strings.HasPrefix(meta, "Func{") {
			if !strings.HasSuffix(meta, "}Func") {
				return "", lerrors.Syntax.Errorf("unterminated Func metadata")
			}
			meta, err = normalizeDefinitionSignature(meta[5 : len(meta)-5])
			if err != nil {
				return "", err
			}
		} else {
			meta = source[start+2 : end]
		}
		result.WriteString(meta)
		result.WriteString("|}")
		source = source[end+2:]
	}
}

func normalizeDefinitionSignature(source string) (string, error) {
	fields, err := splitDefinitionFields(source)
	if err != nil {
		return "", err
	}
	values := make(map[string]string)
	for _, field := range fields {
		key, value, ok := strings.Cut(field, ":")
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		if !ok || (key != "Key" && key != "Inputs" && key != "Outputs" && key != "Receiver") {
			return "", lerrors.Syntax.Errorf("invalid Func field %q", field)
		}
		if _, exists := values[key]; exists {
			return "", lerrors.Syntax.Errorf("duplicate Func field %q", key)
		}
		values[key] = value
	}
	name := values["Key"]
	if strings.HasPrefix(name, "\"") {
		name, err = strconv.Unquote(name)
		if err != nil {
			return "", lerrors.Syntax.Errorf("invalid Func name: %w", err)
		}
	}
	if !validDefinitionNativeName(name) || name == "_" {
		return "", lerrors.Syntax.Errorf("Func requires a valid name, got %q", name)
	}
	parts := []string{"Key: ~" + name + "~"}
	if inputs := values["Inputs"]; inputs != "" && inputs != "[]" && inputs != "&[]&" {
		if !strings.HasPrefix(inputs, "&[") || !strings.HasSuffix(inputs, "]&") {
			return "", lerrors.TypeMismatch.Errorf("Func Inputs require a Props value")
		}
		props, err := splitDefinitionFields(inputs[2 : len(inputs)-2])
		if err != nil {
			return "", err
		}
		for _, prop := range props {
			prop = strings.TrimSpace(prop)
			if !strings.HasPrefix(prop, "Prop{") || !strings.HasSuffix(prop, "}Prop") {
				return "", lerrors.TypeMismatch.Errorf("Func Inputs require Prop records")
			}
			fields, err := splitDefinitionFields(prop[5 : len(prop)-5])
			if err != nil {
				return "", err
			}
			for _, field := range fields {
				key, _, _ := strings.Cut(field, ":")
				if strings.TrimSpace(key) == "Value" {
					return "", lerrors.TypeMismatch.Errorf("Func input properties cannot have values")
				}
			}
		}
		parts = append(parts, "Inputs: ~"+inputs+"~")
	}
	if outputs := values["Outputs"]; outputs != "" {
		var types []string
		if err := json.Unmarshal([]byte(outputs), &types); err != nil {
			return "", lerrors.TypeMismatch.Errorf("Func Outputs require a Texts value: %w", err)
		}
		for _, output := range types {
			if !validDefinitionNativeType(output) {
				return "", lerrors.Syntax.Errorf("invalid Func output type %q", output)
			}
		}
		if len(types) > 0 {
			// Protect this list from the backend's comma-separated Func fields.
			// Parenthesized types also keep leading/trailing type brackets
			// distinct from the backend's array delimiters.
			protected := make([]string, len(types))
			for i, output := range types {
				protected[i] = "(" + output + ")"
			}
			parts = append(parts, "Outputs: ~&["+strings.Join(protected, ", ")+"]&~")
		}
	}
	if receiver := values["Receiver"]; receiver != "" {
		if !strings.HasPrefix(receiver, "#[Prop{") || !strings.HasSuffix(receiver, "}Prop]#") {
			return "", lerrors.TypeMismatch.Errorf("Func Receiver requires a Prop value")
		}
		parts = append(parts, "Receiver: ~"+receiver+"~")
	}
	return strings.Join(parts, ", "), nil
}

func splitDefinitionFields(source string) ([]string, error) {
	var parts []string
	var stack []byte
	start := 0
	for i := 0; i < len(source); i++ {
		b := source[i]
		end, literal, err := definitionNativeLiteralEnd(source, i)
		if err != nil {
			return nil, err
		}
		if literal {
			i = end - 1
			continue
		}
		switch b {
		case '{', '[', '(':
			stack = append(stack, b)
		case '}', ']', ')':
			if len(stack) == 0 || !definitionBracketsMatch(stack[len(stack)-1], b) {
				return nil, lerrors.Syntax.Errorf("unbalanced Func metadata")
			}
			stack = stack[:len(stack)-1]
		case ',':
			if len(stack) == 0 {
				parts = append(parts, source[start:i])
				start = i + 1
			}
		}
	}
	if len(stack) != 0 {
		return nil, lerrors.Syntax.Errorf("unbalanced Func metadata")
	}
	return append(parts, source[start:]), nil
}

// This is a framing scanner, not a native expression/type parser. In
// particular, an apostrophe starts a character literal only when one scalar or
// escape has a closing apostrophe; lifetime/label tokens are otherwise opaque.
func definitionNativeLiteralEnd(source string, position int) (int, bool, error) {
	if source[position] == 'r' {
		quote := position + 1
		for quote < len(source) && source[quote] == '#' {
			quote++
		}
		if quote < len(source) && source[quote] == '"' {
			closingDelimiter := `"` + source[position+1:quote]
			if offset := strings.Index(source[quote+1:], closingDelimiter); offset >= 0 {
				return quote + 1 + offset + len(closingDelimiter), true, nil
			}
			return 0, false, lerrors.Syntax.Errorf("unterminated Rust raw string at byte %d", position)
		}
	}
	if source[position] == '"' || source[position] == '`' {
		end, err := quotedEnd([]byte(source), position)
		return end, true, err
	}
	if source[position] != '\'' || position+1 >= len(source) {
		return position, false, nil
	}
	end := position + 1
	if source[end] == '\\' {
		end++
		if end >= len(source) {
			return position, false, nil
		}
		switch source[end] {
		case 'u':
			if end+1 >= len(source) || source[end+1] != '{' {
				return position, false, nil
			}
			offset := strings.IndexByte(source[end+2:], '}')
			if offset < 0 {
				return position, false, nil
			}
			end += offset + 3
		case 'x':
			end += 3
		default:
			end++
		}
	} else {
		_, size := utf8.DecodeRuneInString(source[end:])
		end += size
	}
	if end < len(source) && source[end] == '\'' {
		return end + 1, true, nil
	}
	return position, false, nil
}

func definitionBracketsMatch(open, closing byte) bool {
	return open == '{' && closing == '}' || open == '[' && closing == ']' || open == '(' && closing == ')'
}

// Parentheses group sibling expressions separated by *. A single chained
// expression needs no extra group; removing it also avoids nested placeholder
// groups in the backend while preserving the source expression separately.
func normalizeDefinitionGroups(source string) (string, error) {
	position := 0
	var group func(bool) (string, bool, error)
	group = func(nested bool) (string, bool, error) {
		var output strings.Builder
		siblings := false
		for position < len(source) {
			if strings.HasPrefix(source[position:], "{|") {
				end, err := definitionMetadataEnd(source, position)
				if err != nil {
					return "", false, err
				}
				output.WriteString(source[position : end+2])
				position = end + 2
				continue
			}
			current := source[position]
			position++
			switch current {
			case '(':
				inner, many, err := group(true)
				if err != nil {
					return "", false, err
				}
				if many {
					output.WriteByte('(')
				}
				output.WriteString(inner)
				if many {
					output.WriteByte(')')
				}
			case ')':
				if !nested {
					return "", false, lerrors.Syntax.Errorf("unbalanced glyph expression")
				}
				if strings.TrimSpace(output.String()) == "" {
					return "", false, lerrors.Syntax.Errorf("empty glyph group")
				}
				return output.String(), siblings, nil
			default:
				if current == '*' {
					siblings = true
				}
				output.WriteByte(current)
			}
		}
		if nested {
			return "", false, lerrors.Syntax.Errorf("unbalanced glyph expression")
		}
		return output.String(), siblings, nil
	}
	result, _, err := group(false)
	return result, err
}

// Validate only expression structure here. Metadata has its own syntax, and
// unused registered templates are validated when their arguments are supplied.
func validateDefinitionExpression(expression string) error {
	if strings.TrimSpace(expression) == "" {
		return lerrors.Syntax.Errorf("empty glyph expression")
	}
	depth, expectsOperand := 0, true
	for i := 0; i < len(expression); i++ {
		if strings.HasPrefix(expression[i:], "{|") {
			end, err := definitionMetadataEnd(expression, i)
			if err != nil {
				return err
			}
			i = end + 1
			continue
		}
		switch expression[i] {
		case ' ', '\t', '\n', '\r':
		case '(':
			depth++
			expectsOperand = true
		case ')':
			if depth == 0 || expectsOperand {
				return lerrors.Syntax.Errorf("unbalanced or empty glyph expression")
			}
			depth--
		case '^', '*':
			if expectsOperand {
				return lerrors.Syntax.Errorf("glyph operator has no preceding expression")
			}
			expectsOperand = true
		default:
			expectsOperand = false
		}
	}
	if depth != 0 || expectsOperand {
		return lerrors.Syntax.Errorf("incomplete glyph expression")
	}
	return nil
}
