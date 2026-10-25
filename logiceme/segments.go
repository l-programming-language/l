/*
Author: Connor Sanders
Version: 0.0.1
Released: 11/01/2020

-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
		Logiceme Evaluator V0.0.1
-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
*/

package Logiceme

import (
	"bytes"
	"fmt"
	"io/fs"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

/*
==========================
UTILITY FUNCTIONS
==========================
*/

// Parameter
type Parameter struct {
	Key     string
	Value   string
	Type    string
	SubType string
	Plural  bool
}

// ScanParams
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

// GetParams
func GetParams(expressionStr string) []string {
	return strings.Split(strings.Split(strings.Split(expressionStr, "(")[1], ")")[0], ", ")
}

// GetFunctionalType
func GetFunctionalType(params []string) string {
	eType := params[0]
	if strings.Contains(params[0], " = ") {
		eType = "Default"
	} else {
		eType = strings.Replace(eType, `"`, ``, -1)
	}
	return eType
}

/*
==========================
SEGMENT STRUCTS
==========================
*/

// Segment
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

// getSegmentName
func (s *Segment) getSegmentName() string {
	switch segSymbol := s.Symbol; segSymbol {
	case "A":
		return "Declaration"
	case "E":
		return "Calculation"
	case "I":
		return "Evaluation"
	case "O":
		return "Definition"
	case "U":
		return "Description"
	case "W":
		return "Operation"
	case "X":
		return "Iteration"
	case "Y":
		return "Instantiation"
	case "V":
		return "Execution"
	case "B":
		return "Declare"
	case "C":
		return "Execute"
	case "D":
		return "Calculate"
	case "F":
		return "Define"
	case "G":
		return "Evaluate"
	case "H":
		return "Describe"
	case "K":
		return "Operate"
	case "M":
		return "Instantiate"
	case "P":
		return "Iterate"
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

// LoadFeatural
func (s *Segment) LoadFeatural() {
	// TODO Add Functions to convert between byte slice and data types
	// Add default values
	// Test Module
	switch segSymbol := s.Symbol; segSymbol {
	case ".":
		s.Content = []byte{}
		s.ContentType = "Byte"
	case "?":
		s.Content = []byte("True")
		s.ContentType = "Boolean"
	case ",":
		s.Content = []byte("0")
		s.ContentType = "Integer"
	case "%":
		s.Content = []byte("0.0")
		s.ContentType = "Float"
	case "~":
		s.Content = []byte("``")
		s.ContentType = "String"
	case "!":
		s.Content = []byte("Null")
		s.ContentType = "Null"
	case "&":
		s.Content = []byte("[]")
		s.ContentType = "Array"
	case "#":
		s.Content = []byte("{}")
		s.ContentType = "Hashmap"
	}
}

// featuralCheck
func featuralCheck(segmentStr string) []string {
	var reStrs []string
	featurals := []string{".", "?", ",", "%", "~", "!", "&", "#"}
	featuralTypes := []string{"Byte", "Boolean", "Integer", "Float", "String", "Null", "Array", "Hashmap"}
	for ind, featural := range featurals {
		if segmentStr == featural {
			fType := featuralTypes[ind]
			reStrs = append(reStrs, featural)
			reStrs = append(reStrs, fType)
			return reStrs
		}
	}
	return reStrs
}

// vowelCheck
func vowelCheck(segmentStr string) []string {
	var reStrs []string
	vowels := []string{"A", "E", "I", "O", "U", "W", "X", "Y", "V"}
	vowelTypes := []string{"Declaration", "Calculation", "Evaluation", "Definition", "Description", "Operation", "Iteration", "Instantiation", "Execution"}
	for ind, vowel := range vowels {
		if segmentStr == vowel {
			fType := vowelTypes[ind]
			reStrs = append(reStrs, vowel)
			reStrs = append(reStrs, fType)
			return reStrs
		}
	}
	return reStrs
}

// consonantCheck
func consonantCheck(segmentStr string) []string {
	var reStrs []string
	consonants := []string{"B", "D", "F", "G", "H", "K", "M", "P", "C"}
	consonantTypes := []string{"Declare", "Calculate", "Define", "Evaluate", "Describe", "Operate", "Instantiate", "Iterate", "Execute"}
	for ind, consonant := range consonants {
		if segmentStr == consonant {
			fType := consonantTypes[ind]
			reStrs = append(reStrs, consonant)
			reStrs = append(reStrs, fType)
			return reStrs
		}
	}
	return reStrs
}

// conjugateCheck
func conjugateCheck(segmentStr string) []string {
	var reStrs []string
	conjugates := []string{"^", "*", "+"}
	conjugateTypes := []string{"Square", "Multiply", "Add"}
	for ind, conjugate := range conjugates {
		if segmentStr == conjugate {
			fType := conjugateTypes[ind]
			reStrs = append(reStrs, conjugate)
			reStrs = append(reStrs, fType)
			return reStrs
		}
	}
	return reStrs
}

// DetermineSegmentType
func DetermineSegmentType(segmentStr string) []string {
	var reStrs []string
	fStrs := featuralCheck(segmentStr)
	if len(fStrs) == 2 {
		reStrs = append(reStrs, "Featural")
		reStrs = append(reStrs, fStrs...)
		return reStrs
	}
	cStrs := consonantCheck(segmentStr)
	if len(cStrs) == 2 {
		reStrs = append(reStrs, "Consonant")
		reStrs = append(reStrs, cStrs...)
		return reStrs
	}
	vStrs := vowelCheck(segmentStr)
	if len(vStrs) == 2 {
		reStrs = append(reStrs, "Vowel")
		reStrs = append(reStrs, vStrs...)
		return reStrs
	}
	jStrs := conjugateCheck(segmentStr)
	if len(jStrs) == 2 {
		reStrs = append(reStrs, "Conjugate")
		reStrs = append(reStrs, jStrs...)
		return reStrs
	}
	return reStrs
}

// Featural
type Featural struct {
	Name    string  `json:"name,omitempty"`
	Type    string  `json:"type,omitempty"`
	Value   string  `json:"value,omitempty"`
	Segment Segment `json:"segment,omitempty"`
}

// SegmentalSystem
type SegmentalSystem struct {
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

// SegmentalDataSwitch
func (ss *SegmentalSystem) SegmentalDataSwitch(segmentName string) string {
	switch segName := segmentName; segName {
	case "Declaration", "Declarations":
		return "Declaration"
	case "Calculation", "Calculations":
		return "Calculation"
	case "Evaluation", "Evaluations":
		return "Evaluation"
	case "Definition", "Definitions":
		return "Definition"
	case "Description", "Descriptions":
		return "Description"
	case "Operation", "Operations":
		return "Operation"
	case "Iteration", "Iterations":
		return "Iteration"
	case "Instantiation", "Instantiations":
		return "Instantiation"
	case "Execution", "Executions":
		return "Execution"
	}
	return "Void"
}

// SegmentSwitch
func (ss *SegmentalSystem) SegmentSwitch(segmentSymbol string) Segment {
	switch segSymbol := segmentSymbol; segSymbol {
	case "A":
		return ss.Declaration
	case "E":
		return ss.Calculation
	case "I":
		return ss.Evaluation
	case "O":
		return ss.Definition
	case "U":
		return ss.Description
	case "W":
		return ss.Operation
	case "X":
		return ss.Iteration
	case "Y":
		return ss.Instantiation
	case "V":
		return ss.Execution
	case "B":
		return ss.Declare
	case "C":
		return ss.Execute
	case "D":
		return ss.Calculate
	case "F":
		return ss.Define
	case "G":
		return ss.Evaluate
	case "H":
		return ss.Describe
	case "K":
		return ss.Operate
	case "M":
		return ss.Instantiate
	case "P":
		return ss.Iterate
	case ".":
		return ss.Byte
	case "?":
		return ss.Boolean
	case ",":
		return ss.Integer
	case "%":
		return ss.Float
	case "~":
		return ss.String
	case "!":
		return ss.Null
	case "&":
		return ss.Array
	case "#":
		return ss.Hashmap
	}
	return Segment{}
}

// SegmentInput
type SegmentInput struct {
	Key   string `json:"key,omitempty"`
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
}

/*
==========================
GLYPH STRUCTS
==========================
*/

// DataCompilation
type DataCompilation struct {
	Conjoinments    []CompiledConjoinment
	Meta            MetaDict
	SegmentalSystem *SegmentalSystem
}

// Glyph
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
func (g *Glyph) DetermineSegTags(seg Segment) Segment {
	if g.Determinant.Symbol == "" {
		// Get Default Tag if No Determinant
		tags := ParseSegmentTags(seg.Tags, g.Meta, "Default", true)
		seg.Tags = tags
	} else {
		// Get Non Default if Determinant
		if g.Determinant.Symbol == "E" {
			// E Vowel Determinant
			//fmt.Println("CHECK SEGMENT, ", string(seg.Content))
			//fmt.Println("CHECK FOR E DETERMINANT TAG!", string(g.Determinant.Content))
			tags := ParseSegmentTags(seg.Tags, g.Meta, "E", true)
			seg.Tags = tags
		} else {
			// Regular Consonant Determinant
			segName := g.Determinant.getSegmentName()
			tags := ParseSegmentTags(seg.Tags, g.Meta, segName, true)
			seg.Tags = tags
		}
	}
	return seg
}

// CompileConjoinmentsBytes
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

// CompiledConjoinment
type CompiledConjoinment struct {
	Key             string
	Type            string
	DeterminantType string
	Contents        [][]byte
	Plural          bool
	Index           int
}

// InsertMeta
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

// CompileConjoinments
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

// segmentConsonantSwitch
func segmentConsonantSwitch(consonantType string) string {
	switch segType := consonantType; segType {
	case "calculate", "Calculate":
		return "Calculation"
	case "declare", "Declare":
		return "Declaration"
	case "define", "Define":
		return "Definition"
	case "evaluate", "Evaluate":
		return "Evaluation"
	case "describe", "Describe":
		return "Description"
	case "operate", "Operate":
		return "Operation"
	case "iterate", "Iterate":
		return "Iteration"
	case "instantiate", "Instantiate":
		return "Instantiation"
	case "execute", "Execute":
		return "Execution"
	default:
		return "Void"
	}
}

// checkParamsConsonant
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
	selected := "Default"
	if g.Determinant.Symbol == "E" {
		selected = "E"
	} else if g.Determinant.Symbol != "" {
		selected = g.Determinant.getSegmentName()
	}
	tags := vowel.Tags
	if len(tags) == 0 {
		tags = CompileMainTags(vowel.Content)
	}
	if len(ParseSegmentTags(tags, g.Meta, selected, false)) > 0 {
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
		varsParser, _ := LoadRegex("MetaDefs", "") //nolint:errcheck // LoadRegex returns a compiled built-in expression and a nil error.
		parsedBytes := varsParser.Parse(definition)
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

// ParseGlyphExpressions parses the glyph expressions wrapped in () and replaces with a uuid placeholder string
func ParseGlyphExpressions(definition []byte) ([]byte, [][]string) {
	var varManager [][]string
	curExp := ""
	curCount := 0
	inExp := false
	if bytes.Contains(definition, []byte("(")) && bytes.Contains(definition, []byte(")")) {
		for _, d := range definition {
			if string(d) == "(" {
				if curCount == 0 {
					inExp = true
				}
				curCount = curCount + 1
				curExp = curExp + string(d)
			} else if string(d) == ")" {
				curCount = curCount - 1
				curExp = curExp + string(d)
				if curCount == 0 {
					varUuid := GenerateUuid()
					varUuid = strings.Replace(varUuid, "-", "", -1)
					varUuidRP := "|" + varUuid + "|"
					valEnt := curExp[1:]
					valEnt = valEnt[:len(valEnt)-1]
					curEnt := []string{varUuid, valEnt}
					if bytes.Contains([]byte(valEnt), []byte("(")) && bytes.Contains([]byte(valEnt), []byte(")")) {
						subDef, subVarManager := ParseGlyphExpressions([]byte(valEnt))
						varManager = append(varManager, subVarManager...)
						curEnt = []string{varUuid, string(subDef)}
					}
					varManager = append(varManager, curEnt)
					definition = bytes.Replace(definition, []byte(curExp), []byte(varUuidRP), -1)
					inExp = false
					curExp = ""
				}
			} else {
				if inExp {
					curExp = curExp + string(d)
				}
			}
		}
	}
	return definition, varManager
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

// I.E. `BA~`, `O`, `BA~^O`
// [Determinate, Vowel, Featural-META-Featural, Conjugate]
// TODO Ensure the following runs:
// [`A`, `W{Key: ~main~}^U^O^E{Imports: [~“fmt”~], Statement: ~fmt.Println(“Hello World!”)~}`]
//  DECLARE, Operation^Description^DEFINITION^CALCULATION

// InitializeGlyph initializes the parsing process of a glyph sequence
func (ss *SegmentalSystem) InitializeGlyph(definition []byte, valVars [][]string, expVars [][]string) Glyph {
	glyph, err := ss.initializeGlyphChecked(definition, valVars, expVars)
	// Retain partial diagnostic state; Compile returns this error before rendering.
	//noinspection GoDfaErrorMayBeNotNil
	glyph.initializationErr = err
	return glyph
}

func (ss *SegmentalSystem) initializeGlyphChecked(definition []byte, valVars [][]string, expVars [][]string) (Glyph, error) {
	var seg Segment
	//var feat Featural
	var varValManager, varExpManager [][]string
	reGlyph := Glyph{Name: string(definition), Type: "Simple", SegmentalSystem: ss, Meta: MetaDict{}}
	definition, varValManager = ParseMetaInserts(definition)
	valVars = append(valVars, varValManager...)
	inputCount := 0
	input := false
	inputStr := ""
	definition, varExpManager = ParseGlyphExpressions(definition)
	expVars = append(expVars, varExpManager...)
	splitDefinition := bytes.Split(definition, []byte("^"))
	curDef := splitDefinition[0]
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
			if sType[0] == "Consonant" {
				if ind == 0 {
					// Determinant
					seg = ss.SegmentSwitch(sType[1])
					reGlyph.Determinant = seg
				} else {
					// Consonant
					if ss.Language == "Rust" && reGlyph.Consonant.Symbol != "" {
						return reGlyph, lerrors.Unsupported.Errorf("unsupported stacked Rust codas in %s", reGlyph.Name)
					}
					seg = ss.SegmentSwitch(sType[1])
					reGlyph.Consonant = seg
				}
			} else if sType[0] == "Vowel" {
				if ind == 0 {
					det := false
					if sType[1] == "E" && len(curDef) > ind+1 {
						nextLogiceme := string(curDef[ind+1])
						if nextLogiceme == "A" {
							// EA Vowel
							seg = ss.SegmentSwitch(sType[1])
							reGlyph.Determinant = seg
							det = true
						}
					}
					if !det {
						// Default Vowel
						seg = ss.SegmentSwitch(sType[1])
						reGlyph.Vowels = append(reGlyph.Vowels, seg)
					}
				} else {
					// Determined Vowel
					seg = ss.SegmentSwitch(sType[1])
					reGlyph.Vowels = append(reGlyph.Vowels, seg)
				}
			}
		}
	}
	if len(splitDefinition) > 1 {
		var nextDefContent string
		for ind, sD := range splitDefinition {
			if ind > 0 {
				if len(splitDefinition)-1 == ind {
					nextDefContent = nextDefContent + string(sD)
				} else {
					nextDefContent = nextDefContent + string(sD) + "^"
				}
			}
		}
		if bytes.HasPrefix([]byte(nextDefContent), []byte("|")) {
			// Expand this group's placeholder only. Nested groups must remain
			// opaque until their owning child is parsed; expanding them here
			// exposes their ^ and * delimiters to the wrong tree level.
			exp := []byte(nextDefContent)
			for _, variable := range expVars {
				if nextDefContent == "|"+variable[0]+"|" {
					exp = []byte(variable[1])
					break
				}
			}
			conjs := bytes.Split(exp, []byte("*"))
			for _, conj := range conjs {
				cGlyph, err := ss.initializeGlyphChecked(conj, valVars, expVars)
				if err != nil {
					return reGlyph, err
				}
				reGlyph.Type = "Complex"
				reGlyph.Conjoinments = append(reGlyph.Conjoinments, cGlyph)
			}
		} else {
			cGlyph, err := ss.initializeGlyphChecked([]byte(nextDefContent), valVars, expVars)
			if err != nil {
				return reGlyph, err
			}
			reGlyph.Type = "Complex"
			reGlyph.Conjoinments = append(reGlyph.Conjoinments, cGlyph)
		}
	}
	valInput := InsertGlyphVals([]byte(reGlyph.InputString), valVars)
	reGlyph.InputString = string(valInput)
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
	switch segSymbol := segment.Symbol; segSymbol {
	case "A":
		ss.Declaration = segment
	case "E":
		ss.Calculation = segment
	case "I":
		ss.Evaluation = segment
	case "O":
		ss.Definition = segment
	case "U":
		ss.Description = segment
	case "W":
		ss.Operation = segment
	case "X":
		ss.Iteration = segment
	case "Y":
		ss.Instantiation = segment
	case "V":
		ss.Execution = segment
	case "B":
		ss.Declare = segment
	case "C":
		ss.Execute = segment
	case "D":
		ss.Calculate = segment
	case "F":
		ss.Define = segment
	case "G":
		ss.Evaluate = segment
	case "H":
		ss.Describe = segment
	case "K":
		ss.Operate = segment
	case "M":
		ss.Instantiate = segment
	case "P":
		ss.Iterate = segment
	}
	return nil
}

// AddFeatural maps featurals into the system
func (ss *SegmentalSystem) AddFeatural(segment Segment) {
	segment.LoadFeatural()
	switch segSymbol := segment.Symbol; segSymbol {
	case ".":
		ss.Byte = segment
	case "?":
		ss.Boolean = segment
	case ",":
		ss.Integer = segment
	case "%":
		ss.Float = segment
	case "~":
		ss.String = segment
	case "!":
		ss.Null = segment
	case "&":
		ss.Array = segment
	case "#":
		ss.Hashmap = segment
	}
}

// loadSegmentalSystem builds a fresh Rust alphabet from its fixed identities.
func loadSegmentalSystem(resources fs.FS) (SegmentalSystem, error) {
	segments := [...]struct{ symbol, subtype, name string }{
		{"C", "consonant", "execute"},
		{"B", "consonant", "declare"},
		{"D", "consonant", "calculate"},
		{"F", "consonant", "define"},
		{"G", "consonant", "evaluate"},
		{"H", "consonant", "describe"},
		{"K", "consonant", "operate"},
		{"M", "consonant", "instantiate"},
		{"P", "consonant", "iterate"},
		{"V", "vowel", "execution"},
		{"A", "vowel", "declaration"},
		{"E", "vowel", "calculation"},
		{"I", "vowel", "evaluation"},
		{"O", "vowel", "definition"},
		{"U", "vowel", "description"},
		{"W", "vowel", "operation"},
		{"X", "vowel", "iteration"},
		{"Y", "vowel", "instantiation"},
	}
	system := SegmentalSystem{Name: "rust", Language: "Rust"}
	for _, definition := range segments {
		segment := Segment{
			Symbol: definition.symbol, Type: "logiceme", SubType: definition.subtype,
			Name: definition.name, Inputs: []string{"Type"},
		}
		// Preserve the effective ReadString record, including its final
		// backtick. This storage refactor does not change the glyph protocol.
		segment.Meta.Add("", "string", "Default`")
		if err := system.AddLogiceme(segment, true, resources); err != nil {
			return system, err
		}
	}
	system.Name = "Default"
	system.Type = "Full"
	return system, nil
}
