// Author: Connor Sanders

package Logiceme

import (
	"bytes"
	"fmt"
	"io/fs"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// segmentDefinitions owns alphabet names, defaults and field selection. The
// ordered logiceme prefix also preserves template-load/error order.
type segmentDefinition struct {
	symbol, family, name, initial, related string
	slot                                   func(*SegmentalSystem) *Segment
}

var segmentDefinitions = []segmentDefinition{
	{"C", "Consonant", "Execute", "", "Execution", func(s *SegmentalSystem) *Segment { return &s.Execute }},
	{"B", "Consonant", "Declare", "", "Declaration", func(s *SegmentalSystem) *Segment { return &s.Declare }},
	{"D", "Consonant", "Calculate", "", "Calculation", func(s *SegmentalSystem) *Segment { return &s.Calculate }},
	{"F", "Consonant", "Define", "", "Definition", func(s *SegmentalSystem) *Segment { return &s.Define }},
	{"G", "Consonant", "Evaluate", "", "Evaluation", func(s *SegmentalSystem) *Segment { return &s.Evaluate }},
	{"H", "Consonant", "Describe", "", "Description", func(s *SegmentalSystem) *Segment { return &s.Describe }},
	{"K", "Consonant", "Operate", "", "Operation", func(s *SegmentalSystem) *Segment { return &s.Operate }},
	{"M", "Consonant", "Instantiate", "", "Instantiation", func(s *SegmentalSystem) *Segment { return &s.Instantiate }},
	{"P", "Consonant", "Iterate", "", "Iteration", func(s *SegmentalSystem) *Segment { return &s.Iterate }},
	{"V", "Vowel", "Execution", "", "", func(s *SegmentalSystem) *Segment { return &s.Execution }},
	{"A", "Vowel", "Declaration", "", "", func(s *SegmentalSystem) *Segment { return &s.Declaration }},
	{"E", "Vowel", "Calculation", "", "", func(s *SegmentalSystem) *Segment { return &s.Calculation }},
	{"I", "Vowel", "Evaluation", "", "", func(s *SegmentalSystem) *Segment { return &s.Evaluation }},
	{"O", "Vowel", "Definition", "", "", func(s *SegmentalSystem) *Segment { return &s.Definition }},
	{"U", "Vowel", "Description", "", "", func(s *SegmentalSystem) *Segment { return &s.Description }},
	{"W", "Vowel", "Operation", "", "", func(s *SegmentalSystem) *Segment { return &s.Operation }},
	{"X", "Vowel", "Iteration", "", "", func(s *SegmentalSystem) *Segment { return &s.Iteration }},
	{"Y", "Vowel", "Instantiation", "", "", func(s *SegmentalSystem) *Segment { return &s.Instantiation }},
	{".", "Featural", "Byte", "", "", func(s *SegmentalSystem) *Segment { return &s.Byte }},
	{"?", "Featural", "Boolean", "True", "", func(s *SegmentalSystem) *Segment { return &s.Boolean }},
	{",", "Featural", "Integer", "0", "", func(s *SegmentalSystem) *Segment { return &s.Integer }},
	{"%", "Featural", "Float", "0.0", "", func(s *SegmentalSystem) *Segment { return &s.Float }},
	{"~", "Featural", "String", "``", "", func(s *SegmentalSystem) *Segment { return &s.String }},
	{"!", "Featural", "Null", "Null", "", func(s *SegmentalSystem) *Segment { return &s.Null }},
	{"&", "Featural", "Array", "[]", "", func(s *SegmentalSystem) *Segment { return &s.Array }},
	{"#", "Featural", "Hashmap", "{}", "", func(s *SegmentalSystem) *Segment { return &s.Hashmap }},
	{"^", "Conjugate", "Square", "", "", nil},
	{"*", "Conjugate", "Multiply", "", "", nil},
	{"+", "Conjugate", "Add", "", "", nil},
}

func segmentDefinitionFor(symbol string) (segmentDefinition, bool) {
	for _, definition := range segmentDefinitions {
		if definition.symbol == symbol {
			return definition, true
		}
	}
	return segmentDefinition{}, false
}

type Parameter struct {
	Key     string
	Value   string
	Type    string
	SubType string
	Plural  bool
}

func ScanParams(params []string) []Parameter {
	var pars []Parameter
	for _, param := range params {
		if strings.Contains(param, " = ") {
			plur := false
			var par Parameter
			sParam := strings.Split(param, " = ")
			par.Key = sParam[0]
			par.Value = sParam[1]
			sVal := strings.Split(sParam[1], ".")
			par.Type = sVal[0]
			par.SubType = ""
			if len(sVal) > 1 {
				par.SubType = sVal[1]
			} else if par.Value == "Void" {
				par.Type = "Meta"
				par.SubType = sParam[0]
			}
			if len(par.SubType) > 1 {
				if par.SubType[len(par.SubType)-1:] == "s" {
					plur = true
				}
			}
			par.Plural = plur
			pars = append(pars, par)
		}
	}
	return pars
}

func GetParams(expressionStr string) []string {
	return strings.Split(strings.Split(strings.Split(expressionStr, "(")[1], ")")[0], ", ")
}

func GetFunctionalType(params []string) string {
	eType := params[0]
	if strings.Contains(params[0], " = ") {
		eType = "Default"
	} else {
		eType = strings.Replace(eType, `"`, ``, -1)
	}
	return eType
}

type Segment struct {
	Name        string      `json:"name,omitempty"`
	Type        string      `json:"type,omitempty"`
	SubType     string      `json:"subtype,omitempty"`
	Symbol      string      `json:"symbol,omitempty"`
	Content     []byte      `json:"content,omitempty"`
	ContentType string      `json:"content_type,omitempty"`
	Tags        []*LucidTag `json:"tags,omitempty"`
	Meta        MetaDict    `json:"meta,omitempty"`
	Inputs      []string    `json:"inputs,omitempty"`
}

func (s *Segment) getSegmentName() string {
	if d, ok := segmentDefinitionFor(s.Symbol); ok && (d.family == "Vowel" || d.family == "Consonant") {
		return d.name
	}
	return "Error"
}

func (s *Segment) loadTemplate(resources fs.FS) error {
	fileName := "builtins/segments/" + s.SubType + "s/" + s.Name + ".lu"
	b, err := readAsset(resources, fileName)
	if err != nil {
		return fmt.Errorf("load segment %s: %w", s.Name, err)
	}
	if err := validateLucidTemplate(b, true); err != nil {
		return fmt.Errorf("load segment template %s: %w", fileName, err)
	}
	s.Content = b
	s.ContentType = "Logiceme"
	s.Tags = []*LucidTag{}
	return nil
}

func (s *Segment) LoadFeatural() {
	if d, ok := segmentDefinitionFor(s.Symbol); ok && d.family == "Featural" {
		s.Content, s.ContentType = []byte(d.initial), d.name
	}
}

func DetermineSegmentType(symbol string) []string {
	if d, ok := segmentDefinitionFor(symbol); ok {
		return []string{d.family, symbol, d.name}
	}
	return nil
}

type Featural struct {
	Name    string  `json:"name,omitempty"`
	Type    string  `json:"type,omitempty"`
	Value   string  `json:"value,omitempty"`
	Segment Segment `json:"segment,omitempty"`
}

type SegmentalSystem struct {
	templateCases    map[string]*segmentCases
	fileMeta         MetaDict
	itemVisibility   string
	initializeConfig bool
	Name             string  `json:"name,omitempty"`
	Type             string  `json:"type,omitempty"`
	Language         string  `json:"language,omitempty"`
	Declaration      Segment `json:"declaration,omitempty"`
	Calculation      Segment `json:"calculation,omitempty"`
	Evaluation       Segment `json:"evaluation,omitempty"`
	Definition       Segment `json:"definition,omitempty"`
	Description      Segment `json:"description,omitempty"`
	Operation        Segment `json:"operation,omitempty"`
	Iteration        Segment `json:"iteration,omitempty"`
	Instantiation    Segment `json:"instantiation,omitempty"`
	Execution        Segment `json:"execution,omitempty"`
	Declare          Segment `json:"declare,omitempty"`
	Calculate        Segment `json:"calculate,omitempty"`
	Define           Segment `json:"define,omitempty"`
	Evaluate         Segment `json:"evaluate,omitempty"`
	Describe         Segment `json:"describe,omitempty"`
	Operate          Segment `json:"operate,omitempty"`
	Instantiate      Segment `json:"instantiate,omitempty"`
	Iterate          Segment `json:"iterate,omitempty"`
	Execute          Segment `json:"execute,omitempty"`
	Byte             Segment `json:"byte,omitempty"`
	Boolean          Segment `json:"boolean,omitempty"`
	Integer          Segment `json:"integer,omitempty"`
	Float            Segment `json:"float,omitempty"`
	String           Segment `json:"string,omitempty"`
	Null             Segment `json:"null,omitempty"`
	Array            Segment `json:"array,omitempty"`
	Hashmap          Segment `json:"hashmap,omitempty"`
}

func (ss *SegmentalSystem) SegmentalDataSwitch(name string) string {
	return vowelName(name)
}

func vowelName(name string) string {
	name = strings.TrimSuffix(name, "s")
	for _, d := range segmentDefinitions {
		if d.family == "Vowel" && d.name == name {
			return name
		}
	}
	return "Void"
}

func (ss *SegmentalSystem) SegmentSwitch(symbol string) Segment {
	if d, ok := segmentDefinitionFor(symbol); ok && d.slot != nil {
		return *d.slot(ss)
	}
	return Segment{}
}

type SegmentInput struct {
	Key   string `json:"key,omitempty"`
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
}

type DataCompilation struct {
	Conjoinments    []CompiledConjoinment
	Meta            MetaDict
	SegmentalSystem *SegmentalSystem
}

type Glyph struct {
	Name              string
	Type              string
	SegmentalSystem   *SegmentalSystem
	Determinant       Segment
	Vowels            []Segment
	Consonant         Segment
	Featurals         []Featural
	Conjoinments      []Glyph
	Meta              MetaDict
	Content           []byte
	InputString       string
	Count             int
	Exponent          int
	Polarity          bool
	err               error
	initializationErr error
	inlineHeaders     []byte
	plan              *glyphPlan
	origin            *glyphOrigin
}

// DetermineSegTags Glyph bytes
func (g *Glyph) segmentSelection() string {
	if g.Determinant.Symbol == "E" {
		return "E"
	}
	if g.Determinant.Symbol != "" {
		return g.Determinant.getSegmentName()
	}
	return "Default"
}

func (g *Glyph) DetermineSegTags(seg Segment) Segment {
	seg.Tags = ParseSegmentTags(seg.Tags, g.Meta, g.segmentSelection(), true)
	return seg
}

func (g *Glyph) CompileConjoinmentsBytes(conjoinmentType string, plural bool) [][]byte {
	var cBytes [][]byte
	for _, conjoinment := range g.Conjoinments {
		checkStr := conjoinment.Type
		if plural {
			checkStr = checkStr + "s"
		}
		if checkStr == conjoinmentType {
			content := conjoinment.Content
			if g.inlineConjoinment(&conjoinment) {
				var headers []byte
				var err error
				content, headers, err = prepareInlineConjoinment(&conjoinment)
				if err != nil {
					g.err = err
					return nil
				}
				g.inlineHeaders = append(g.inlineHeaders, headers...)
			} else {
				// Resolve each child's available metadata before parent composition.
				// Unknown fields remain available for enclosing-scope substitution.
				content = scopedMetaInjection(content, conjoinment.Meta)
			}
			cBytes = append(cBytes, content)
		}
	}
	return cBytes
}

type CompiledConjoinment struct {
	Key             string
	Type            string
	DeterminantType string
	Contents        [][]byte
	Plural          bool
	Index           int
}

func (cc *CompiledConjoinment) InsertMeta(meta MetaDict) {
	intContents := [][]byte{}
	baseStr := "Meta."
	for _, c := range cc.Contents {
		for _, m := range meta.GetMetaData() {
			keyStr := baseStr + m.Key
			if bytes.Contains(c, []byte(keyStr)) {
				rStr := "[|" + keyStr + "|]"
				c = bytes.Replace(c, []byte(rStr), []byte(m.Value), -1)
			}
		}
		intContents = append(intContents, c)
	}
	cc.Contents = intContents
}

func (g *Glyph) CompileConjoinments(params []Parameter) []CompiledConjoinment {
	var cConjs []CompiledConjoinment
	for _, param := range params {
		if param.Type == "Conjoinments" {
			cConj := CompiledConjoinment{Key: param.Key, Type: param.SubType, DeterminantType: g.Determinant.Name, Plural: param.Plural, Index: -1}
			cConj.Contents = g.CompileConjoinmentsBytes(param.SubType, param.Plural)
			if len(cConj.Contents) > 0 {
				cConjs = append(cConjs, cConj)
			}
		}
	}
	return cConjs
}

func segmentConsonantSwitch(name string) string {
	for _, d := range segmentDefinitions {
		if d.family == "Consonant" && (name == d.name || name == strings.ToLower(d.name)) {
			return d.related
		}
	}
	return "Void"
}

func (g *Glyph) checkParamsConsonant(expressionStr string) string {
	consonantStr := `[|"Consonant"|]`
	if strings.Contains(expressionStr, consonantStr) {
		consKey := segmentConsonantSwitch(g.Consonant.Name)
		if consKey != "Void" && consKey != "" {
			consKey = `"` + consKey + `",`
			expressionStr = strings.Replace(expressionStr, consonantStr, consKey, -1)
			return expressionStr
		}
	}
	expressionStr = strings.Replace(expressionStr, consonantStr, "", -1)
	return expressionStr
}

// StartFunctional begins the logiceme function process
func (g *Glyph) StartFunctional(expressionStr string) ([]byte, []Parameter, string) {
	reBytes := []byte{}
	expressionStr = g.checkParamsConsonant(expressionStr)
	rParams := GetParams(expressionStr)
	params := ScanParams(rParams)
	eType := GetFunctionalType(rParams)
	return reBytes, params, eType
}

// FunctionalTags compile functional lucid tags
func (g *Glyph) FunctionalTags(rawContent []byte, eType string) []*LucidTag {
	tags := CompileMainTags(rawContent)
	return ParseSegmentTags(tags, g.Meta, eType, false)
}

// Compile an initialized Glyph
func (g *Glyph) Compile() error {
	if g.initializationErr != nil {
		return g.initializationErr
	}
	if err := attachGlyphOrigins(g); err != nil {
		return err
	}
	names := &generatedNames{used: glyphInputSource(g), counts: map[string]int{}}
	if err := prepareGlyph(g, glyphContext{}, names); err != nil {
		return locateGlyphError(g, err)
	}
	return g.compilePreparedGlyph()
}

// Headers/Inputs alone cannot satisfy a vowel's selected body. Distinguish a
// damaged supported execution case from an unsupported determinant/vowel pair.
func (g *Glyph) validateSelectedCase(vowel Segment) error {
	if g.SegmentalSystem.Language == "Rust" {
		if err := validateGlyphForm(g); err != nil {
			return err
		}
	}
	executionShape := glyphHasVowel(g, "V") ||
		(g.Determinant.Symbol == "C" &&
			(glyphHasVowel(g, "E") || glyphHasVowel(g, "I")))
	if executionShape && (len(g.Vowels) != 1 || g.Consonant.Symbol != "") {
		return lerrors.Unsupported.Errorf("unsupported Execution glyph %s: execution forms require one vowel and no terminal consonant", glyphForm(g))
	}
	selected := g.segmentSelection()
	if vowel.hasTemplateCase(g.SegmentalSystem, selected) {
		return nil
	}
	switch glyphForm(g) {
	case "V", "MV", "CE", "CI", "BI":
		return lerrors.InvalidConfiguration.Errorf("segment template %s has no case %q for glyph %s", vowel.Name, selected, g.Name)
	}
	return lerrors.Unsupported.Errorf("unsupported glyph combination %s: segment %s has no case %q", glyphForm(g), vowel.Name, selected)
}

// ParseMetaInserts parses out the meta vals located between the {||} and replaces with a uuid placeholder string
func ParseMetaInserts(definition []byte) ([]byte, [][]string) {
	var varManager [][]string
	if bytes.Contains(definition, []byte("{|")) && bytes.Contains(definition, []byte("|}")) {
		parsedBytes := metaInsertFrames(definition)
		for _, parsedByte := range parsedBytes {
			varUuid := GenerateUuid()
			varUuid = strings.Replace(varUuid, "-", "", -1)
			varUuidRP := "{" + varUuid + "}"
			valEnt := bytes.Trim(parsedByte, "{|")
			valEnt = bytes.Trim(valEnt, "|}")
			varEntry := []string{varUuid, string(valEnt)}
			varManager = append(varManager, varEntry)
			definition = bytes.Replace(definition, parsedByte, []byte(varUuidRP), -1)
		}
	}
	return definition, varManager
}

// Match the existing non-greedy MetaDefs framing without regexp scratch space
// proportional to the complete composition. Replacement order remains separate.
func metaInsertFrames(definition []byte) [][]byte {
	var frames [][]byte
	for {
		start := bytes.Index(definition, []byte("{|"))
		if start < 0 {
			return frames
		}
		end := bytes.Index(definition[start+2:], []byte("|}"))
		if end < 0 {
			return frames
		}
		end += start + 4
		frames = append(frames, definition[start:end])
		definition = definition[end:]
	}
}

// ParseGlyphExpressions parses the glyph expressions wrapped in () and replaces with a uuid placeholder string
func ParseGlyphExpressions(definition []byte) ([]byte, [][]string) {
	var groups [][]string
	var result bytes.Buffer
	depth, start, copied := 0, 0, 0
	for index, symbol := range definition {
		switch symbol {
		case '(':
			if depth == 0 {
				start = index
			}
			depth++
		case ')':
			if depth == 0 {
				continue
			}
			depth--
			if depth != 0 {
				continue
			}
			body, children := ParseGlyphExpressions(definition[start+1 : index])
			id := strings.ReplaceAll(GenerateUuid(), "-", "")
			groups = append(groups, children...)
			groups = append(groups, []string{id, string(body)})
			// Replace this occurrence only; an identical group may occur inside
			// a later parent whose source has not been scanned yet.
			result.Write(definition[copied:start])
			result.WriteString("|" + id + "|")
			copied = index + 1
		}
	}
	result.Write(definition[copied:])
	return result.Bytes(), groups
}

// InsertGlyphVals replaces the variable uuid placeholder with the original string value
func InsertGlyphVals(content []byte, vars [][]string) []byte {
	for _, v := range vars {
		if bytes.Contains(content, []byte(v[0])) {
			content = bytes.Replace(content, []byte(v[0]), []byte(v[1]), -1)
		}
	}
	return content
}

// InitializeGlyph initializes the parsing process of a glyph sequence
func (ss *SegmentalSystem) InitializeGlyph(definition []byte, valVars [][]string, expVars [][]string) Glyph {
	glyph, err := ss.initializeGlyphChecked(definition, valVars, expVars)
	// Retain partial diagnostic state; Compile returns this error before rendering.
	//noinspection GoDfaErrorMayBeNotNil
	glyph.initializationErr = err
	return glyph
}

func (ss *SegmentalSystem) initializeGlyphChecked(definition []byte, valVars [][]string, expVars [][]string) (Glyph, error) {
	name := string(definition)
	bound, values := ParseMetaInserts(definition)
	bound, groups := ParseGlyphExpressions(bound)
	bindings := glyphBindings{values: make(map[string]string, len(values)), groups: make(map[string]string, len(expVars)+len(groups))}
	for _, value := range values {
		bindings.values[value[0]] = value[1]
	}
	bindings.orderedValues = append(valVars, values...)
	if len(valVars) != 0 || len(expVars) != 0 {
		// Explicit public substitution lists retain their ordered replacement
		// semantics; ordinary parsed metadata uses exact placeholder lookup.
		bindings.orderedGroups = append(expVars, groups...)
		bindings.external = true
	}
	for _, group := range append(expVars, groups...) {
		key := "|" + group[0] + "|"
		if _, present := bindings.groups[key]; !present {
			bindings.groups[key] = group[1]
		}
	}
	return ss.initializeBoundGlyph(bound, name, &bindings)
}

// Bind placeholders once per source tree. Descendants contain references to
// these bindings, so rescanning them and searching every sibling's metadata is
// both redundant and quadratic for large composed Functions.
type glyphBindings struct {
	values, groups map[string]string
	orderedValues  [][]string
	orderedGroups  [][]string
	external       bool
}

func (ss *SegmentalSystem) initializeBoundGlyph(definition []byte, name string, bindings *glyphBindings) (Glyph, error) {
	var seg Segment
	reGlyph := Glyph{Name: name, Type: "Simple", SegmentalSystem: ss, Meta: MetaDict{}}
	inputCount := 0
	input := false
	inputStr := ""
	curDef, nextDef, hasChildren := bytes.Cut(definition, []byte("^"))
	for ind, b := range curDef {
		if string(b) == "{" {
			if input {
				inputStr = inputStr + string(b)
			} else {
				input = true
			}
			inputCount = inputCount + 1
		} else if string(b) == "}" {
			inputCount = inputCount - 1
			if input {
				if inputCount == 0 {
					input = false
					reGlyph.InputString = inputStr
					inputStr = ""
				} else {
					inputStr = inputStr + string(b)
				}
			}
		} else if input {
			inputStr = inputStr + string(b)
		} else {
			sType := DetermineSegmentType(string(b))
			if len(sType) != 3 {
				if ss.Language == "Rust" {
					if b == '(' || b == ')' {
						return reGlyph, lerrors.Syntax.Errorf("unbalanced composition parenthesis delimiter in %s", reGlyph.Name)
					}
					return reGlyph, lerrors.Unsupported.Errorf("unknown glyph symbol %q in %s", string(b), reGlyph.Name)
				}
				return reGlyph, lerrors.InvalidInput.Errorf("unknown glyph symbol %q in %s", string(b), reGlyph.Name)
			}
			if sType[0] == "Consonant" || sType[0] == "Vowel" {
				configured := ss.SegmentSwitch(sType[1])
				if configured.Symbol != sType[1] || len(configured.Content) == 0 {
					return reGlyph, lerrors.InvalidConfiguration.Errorf("segment %s (%s) is not configured", sType[1], sType[2])
				}
			}
			seg = ss.SegmentSwitch(sType[1])
			switch sType[0] {
			case "Consonant":
				if ind == 0 {
					reGlyph.Determinant = seg
				} else {
					if ss.Language == "Rust" && reGlyph.Consonant.Symbol != "" {
						return reGlyph, lerrors.Unsupported.Errorf("unsupported stacked Rust codas in %s", reGlyph.Name)
					}
					reGlyph.Consonant = seg
				}
			case "Vowel":
				if ind == 0 && b == 'E' && len(curDef) > 1 && curDef[1] == 'A' {
					reGlyph.Determinant = seg
				} else {
					reGlyph.Vowels = append(reGlyph.Vowels, seg)
				}
			}
		}
	}
	if hasChildren {
		children := [][]byte{nextDef}
		if bytes.HasPrefix(nextDef, []byte("|")) {
			// Expand only this group's placeholder; nested groups own their delimiters.
			exp := nextDef
			if group, found := bindings.groups[string(nextDef)]; found {
				exp = []byte(group)
			}
			children = bytes.Split(exp, []byte("*"))
		}
		for i, child := range children {
			var glyph Glyph
			var err error
			if bindings.external {
				glyph, err = ss.initializeGlyphChecked(child, bindings.orderedValues, bindings.orderedGroups)
			} else {
				glyph, err = ss.initializeBoundGlyph(child, string(child), bindings)
			}
			if err != nil {
				return reGlyph, err
			}
			if i == 0 {
				reGlyph.Conjoinments = make([]Glyph, 0, len(children))
			}
			reGlyph.Type = "Complex"
			reGlyph.Conjoinments = append(reGlyph.Conjoinments, glyph)
		}
	}
	if bindings.external {
		reGlyph.InputString = string(InsertGlyphVals([]byte(reGlyph.InputString), bindings.orderedValues))
	} else if value, found := bindings.values[reGlyph.InputString]; found {
		reGlyph.InputString = value
	} else if reGlyph.InputString != "" {
		reGlyph.InputString = string(InsertGlyphVals([]byte(reGlyph.InputString), bindings.orderedValues))
	}
	for _, vowel := range reGlyph.Vowels {
		if err := reGlyph.validateSelectedCase(vowel); err != nil {
			return reGlyph, err
		}
	}
	return reGlyph, nil
}

// AddLogiceme adds a loaded logiceme into it's correct place in a logiceme system
func (ss *SegmentalSystem) AddLogiceme(segment Segment, load bool, resources fs.FS) error {
	if load {
		if err := segment.loadTemplate(resources); err != nil {
			return err
		}
	}
	if d, ok := segmentDefinitionFor(segment.Symbol); ok && (d.family == "Vowel" || d.family == "Consonant") {
		*d.slot(ss) = segment
		if load {
			if ss.templateCases == nil {
				ss.templateCases = make(map[string]*segmentCases)
			}
			ss.templateCases[segment.Symbol] = parsedSegmentCases.load(segment.Content)
		}
	}
	return nil
}

// AddFeatural maps featurals into the system
func (ss *SegmentalSystem) AddFeatural(segment Segment) {
	if d, ok := segmentDefinitionFor(segment.Symbol); ok && d.family == "Featural" {
		segment.LoadFeatural()
		*d.slot(ss) = segment
	}
}

// loadSegmentalSystem builds a fresh Rust alphabet from its fixed identities.
func loadSegmentalSystem(resources fs.FS) (SegmentalSystem, error) {
	system := SegmentalSystem{Name: "rust", Language: "Rust"}
	for _, definition := range segmentDefinitions {
		if definition.family != "Consonant" && definition.family != "Vowel" {
			continue
		}
		segment := Segment{Symbol: definition.symbol, Type: "logiceme", SubType: strings.ToLower(definition.family), Name: strings.ToLower(definition.name), Inputs: []string{"Type"}}
		// Preserve the effective ReadString record, including its final backtick.
		segment.Meta.Add("", "string", "Default`")
		if err := system.AddLogiceme(segment, true, resources); err != nil {
			return system, err
		}
	}
	system.Name, system.Type = "Default", "Full"
	return system, nil
}
