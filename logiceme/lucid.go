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
	"regexp"
	"strings"
	"sync"

	"github.com/l-programming-language/l/lerrors"
)

/*
==========================
LUCID TAG STRUCTURES
==========================
*/

// TagMeta ...
type TagMeta struct {
	AppMeta MetaDict
}

// SubTag
// (/={SubTagMeta}=={SubTagExpression}=={SubTagTemplate}=/)
type SubTag struct {
	ReplaceString string
	TagString     string
	TagType       string
	Expressions   []*TagExpression
	TagMeta       *TagMeta
}

// NewSubTag
func NewSubTag(replaceString string, tagString string, tagType string, expressions []*TagExpression, tagMeta *TagMeta) *SubTag {
	return &(SubTag{replaceString, tagString, tagType, expressions, tagMeta})
}

// InterpretedSubTag
type InterpretedSubTag struct {
	ReplaceString     string
	InterpretedString string
}

// NewInterpretedSubTag
func NewInterpretedSubTag(replaceString string, interpretedString string) *InterpretedSubTag {
	return &(InterpretedSubTag{replaceString, interpretedString})
}

// TagExpression
type TagExpression struct {
	ReplaceString     string
	ExpressionString  string
	ExpressionType    string
	ExpressionOutput  string
	ExpressionSubTags []*SubTag
	TagMeta           *TagMeta
}

// NewTagExpression
func NewTagExpression(replaceString string, expressionString string, expressionType string, expressionOutput string, subTags []*SubTag, tagMeta *TagMeta) *TagExpression {
	return &(TagExpression{replaceString, expressionString, expressionType, expressionOutput, subTags, tagMeta})
}

// LucidTag
// {|/|-{TagMeta}--{TagExpression}--{TagTemplate}-|/|}
type LucidTag struct {
	ReplaceString string
	TagString     string
	TagType       string
	Expressions   []*TagExpression
	TagMeta       *TagMeta
}

// GetTemplate
func (lt *LucidTag) GetTemplate() []byte {
	exp := *lt.Expressions[2]
	return []byte(exp.ExpressionString)
}

// GetSubTags
func (lt *LucidTag) GetSubTags() []*SubTag {
	var reTags []*SubTag
	for _, te := range lt.Expressions {
		reTags = append(reTags, te.ExpressionSubTags...)
	}
	return reTags
}

// InterpretSubTags
func (lt *LucidTag) InterpretSubTags(compilation DataCompilation) []byte {
	li := LucidInterpreter{
		lt.GetTemplate(),
		[]byte{},
		lt.GetSubTags(),
		[]*InterpretedSubTag{},
		compilation.SegmentalSystem,
		compilation.Meta,
		compilation.Conjoinments,
		[]*WorkingVariable{},
	}
	li.Interpret()
	return li.InterpretedContent
}

// scopedMetaInjection
func scopedMetaInjection(content []byte, meta MetaDict) []byte {
	intStr := string(content)
	condVars := lucidParser(intStr, "ExpressionVar")
	for _, variable := range condVars {
		if strings.Contains(variable, "Meta.") {
			var cleanedVariable string
			if strings.Contains(variable, "[|%") && strings.Contains(variable, "%|]") {
				// Remove all new lines from tag val
				cleanedVariable = strings.Trim(variable, "[|%")
				cleanedVariable = strings.Trim(cleanedVariable, "%|]")
			} else if strings.Contains(variable, "[|&") && strings.Contains(variable, "&|]") {
				// Remove only leading new lines from tag val
				cleanedVariable = strings.Trim(variable, "[|&")
				cleanedVariable = strings.Trim(cleanedVariable, "&|]")
			} else {
				// Default tag handler
				cleanedVariable = strings.Trim(variable, "[|")
				cleanedVariable = strings.Trim(cleanedVariable, "|]")
			}
			sVar := strings.Split(cleanedVariable, ".")[1]
			var featural string
			sVar, featural = scanTokenVarFeaturals(sVar)
			metaVar := meta.GetVal(sVar)
			if metaVar != "NotFound" {
				if strings.Contains(metaVar, "&[") && strings.Contains(metaVar, "]&") {
					if featural != "~" {
						content = bytes.Replace(content, []byte(variable), []byte(metaVar), -1)
					} else {
						reStr := StringifyStringArray(metaVar)
						content = bytes.Replace(content, []byte(variable), []byte(reStr), -1)
					}
				} else {
					content = bytes.Replace(content, []byte(variable), []byte(metaVar), -1)
				}
			}
		}
	}
	return content
}

// Interpret LucidTag
func (lt *LucidTag) Interpret(compilation DataCompilation) []byte {
	if compilation.SegmentalSystem != nil && compilation.SegmentalSystem.Language == "Rust" {
		return lt.InterpretSubTags(compilation)
	}
	template := lt.GetTemplate()
	tType := CheckTemplateType(template)
	if tType == "Variable" {
		template = scopedMetaInjection(template, compilation.Meta)
		template = InterpretConjoinments(template, compilation.Conjoinments, compilation.Meta)
	} else if tType == "SubTag" {
		template = lt.InterpretSubTags(compilation)
	}
	return template
}

// NewLucidTag
func NewLucidTag(replaceString string, tagString string, tagType string, expressions []*TagExpression, tagMeta *TagMeta) *LucidTag {
	return &(LucidTag{replaceString, tagString, tagType, expressions, tagMeta})
}

/*
==========================
LUCID PARSER FUNCTIONS
==========================
*/

// lucidSourceSpan identifies original bytes, including the enclosing delimiters.
type lucidSourceSpan struct {
	start, end int
}

type lucidSubTagSpan struct {
	lucidSourceSpan
	expressions []lucidSourceSpan
}

// scanLucidSubTags recognizes balanced frames without rebuilding their bodies.
// Child frames are consumed while the parent expression remains open, so each
// root's three expression spans include their complete literal and child text.
func scanLucidSubTags(source string) ([]lucidSubTagSpan, error) {
	type frame struct {
		start, expressionStart int
		expressions            []lucidSourceSpan
		open                   bool
	}
	var frames []frame
	var tags []lucidSubTagSpan
	space := func(b byte) bool { return b == ' ' || b == '\t' || b == '\r' || b == '\n' }
	for i := 0; i < len(source); {
		if strings.HasPrefix(source[i:], "(/") {
			expression := i + 2
			for expression < len(source) && space(source[expression]) {
				expression++
			}
			if strings.HasPrefix(source[expression:], "={") {
				if len(frames) > 0 && !frames[len(frames)-1].open {
					return nil, lerrors.InvalidConfiguration.Errorf("nested Lucid subtag outside an expression at byte %d", i)
				}
				frames = append(frames, frame{start: i, expressionStart: expression, open: true})
				i = expression + 2
				continue
			}
		}
		if len(frames) == 0 {
			i++
			continue
		}
		top := &frames[len(frames)-1]
		if top.open {
			if strings.HasPrefix(source[i:], "}=") {
				top.expressions = append(top.expressions, lucidSourceSpan{top.expressionStart, i + 2})
				top.open = false
				i += 2
				continue
			}
			i++
			continue
		}
		if space(source[i]) {
			i++
			continue
		}
		if strings.HasPrefix(source[i:], "/)") {
			if len(top.expressions) != 3 {
				return nil, lerrors.InvalidConfiguration.Errorf("Lucid subtag at byte %d requires three expressions, got %d", top.start, len(top.expressions))
			}
			tag := lucidSubTagSpan{lucidSourceSpan{top.start, i + 2}, top.expressions}
			frames = frames[:len(frames)-1]
			i += 2
			if len(frames) == 0 {
				tags = append(tags, tag)
			}
			continue
		}
		if strings.HasPrefix(source[i:], "={") {
			if len(top.expressions) >= 3 {
				return nil, lerrors.InvalidConfiguration.Errorf("Lucid subtag at byte %d has more than three expressions", top.start)
			}
			top.expressionStart = i
			top.open = true
			i += 2
			continue
		}
		return nil, lerrors.InvalidConfiguration.Errorf("malformed Lucid subtag expression framing at byte %d", i)
	}
	if len(frames) > 0 {
		return nil, lerrors.InvalidConfiguration.Errorf("incomplete Lucid subtag at byte %d", frames[len(frames)-1].start)
	}
	return tags, nil
}

// recursiveSubTagParser returns only the outermost framed tags, in source order.
func recursiveSubTagParser(source string) ([]string, error) {
	spans, err := scanLucidSubTags(source)
	if err != nil {
		return nil, err
	}
	var tags []string
	for _, span := range spans {
		tags = append(tags, source[span.start:span.end])
	}
	return tags, nil
}

var lucidParserPatterns struct {
	once   sync.Once
	byType map[string]*regexp.Regexp
}

// lucidParser - Function that read tags
func lucidParser(textString string, parseType string) []string {
	// Regexp matching is concurrent-safe. Publish these unchanged patterns once;
	// parsed slices and structural tags still belong to each individual call.
	lucidParserPatterns.once.Do(func() {
		lucidParserPatterns.byType = map[string]*regexp.Regexp{
			"MainTag":         regexp.MustCompile(`\{\|/\|((?:.*?\r?\n?)*)\|/\|}`),
			"Expression":      regexp.MustCompile(`-\{((?:.*?\r?\n?)*)}-`),
			"SubTag":          regexp.MustCompile(`\(/((?:.*?\r?\n?)*)/\)`),
			"SubExpression":   regexp.MustCompile(`=\{((?:.*?\r?\n?)*)}=`),
			"ExpressionVar":   regexp.MustCompile(`\[\|((?:.*?\r?\n?)*)\|]`),
			"VariableKey":     regexp.MustCompile(`\[((?:.*?\r?\n?)*)]`),
			"Imports":         regexp.MustCompile(`import \(((?:.*?\r?\n?)*)\n\)`),
			"Dependencies":    regexp.MustCompile(`@@"((?:.*?\r?\n?)*)"@@`),
			"Links":           regexp.MustCompile(`!!"((?:.*?\r?\n?)*)"!!`),
			"Alias":           regexp.MustCompile(`<A\{((?:.*?\r?\n?)*)}A>`),
			"EnvDependencies": regexp.MustCompile(`<E\{((?:.*?\r?\n?)*)}E>`),
		}
	})
	re := lucidParserPatterns.byType[parseType]
	if re == nil {
		re = lucidParserPatterns.byType["MainTag"]
	}
	rawParsedTags := []string{}
	if parseType == "SubTag" {
		tags, err := recursiveSubTagParser(textString)
		if err != nil {
			return nil
		}
		return tags
	} else if parseType == "SubExpression" {
		// The caller supplies a tag's interior. Reuse the frame scanner so
		// nested expressions cannot terminate or replace their parent's body.
		framed := "(/" + textString + "/)"
		tags, err := scanLucidSubTags(framed)
		if err != nil || len(tags) != 1 {
			return nil
		}
		for _, expression := range tags[0].expressions {
			rawParsedTags = append(rawParsedTags, framed[expression.start:expression.end])
		}
	} else {
		rawParsedTags = re.FindAllString(textString, -1)
	}
	return rawParsedTags
}

// parseMainTags
func parseMainTags(content []byte) [][]byte {
	reBytes := [][]byte{}
	segmentStrs := lucidParser(string(content), "MainTag")
	for _, str := range segmentStrs {
		reBytes = append(reBytes, []byte(str))
	}
	return reBytes
}

// RecursiveSubTagParser recursively parses lucid subtags and returns an array of SubTag pointers
func RecursiveSubTagParser(tagExpressionStr string, tagMeta *TagMeta) []*SubTag {
	var subTags []*SubTag
	rawSubTags := lucidParser(tagExpressionStr, "SubTag")
	for _, rawSubTag := range rawSubTags {
		subTagStr := strings.TrimPrefix(rawSubTag, "(/")
		subTagStr = strings.TrimSuffix(subTagStr, "/)")
		var subExpressions []*TagExpression
		rawSubExpressions := lucidParser(subTagStr, "SubExpression")
		for _, rawSubExpression := range rawSubExpressions {
			var subTs []*SubTag
			parseSubs := lucidParser(rawSubExpression, "SubTag")
			if len(parseSubs) > 0 {
				subTs = RecursiveSubTagParser(rawSubExpression, tagMeta)
			}
			subExpressionStr := strings.TrimPrefix(rawSubExpression, "={")
			subExpressionStr = strings.TrimSuffix(subExpressionStr, "}=")
			subExpressionType := determineExpressionType(rawSubExpressions)
			subExpressionOutput := ""
			if len(subTs) > 0 {
				subExpressionOutput = "SubTags"
			} else {
				subExpressionOutput = determineExpressionOutput(rawSubExpressions)
			}
			subTagExp := NewTagExpression(rawSubExpression, subExpressionStr, subExpressionType, subExpressionOutput, subTs, tagMeta)
			subExpressions = append(subExpressions, subTagExp)
		}
		tagType := "sub"
		if strings.Contains(rawSubTag, "={for Header in Headers}=") {
			tagType = "headers"
		}
		subTag := NewSubTag(rawSubTag, subTagStr, tagType, subExpressions, tagMeta)
		subTags = append(subTags, subTag)
	}
	return subTags
}

// injectMetaConjoinments
func injectMetaConjoinments(tag *LucidTag, m MetaDict) *LucidTag {
	curExpStr := tag.Expressions[2].ExpressionString
	for _, mo := range m.GetMetaData() {
		if mo.Type == "Conjoinment" {
			var key string
			var val string
			if strings.Contains(curExpStr, mo.Key+"~") {
				// stringify array
				if strings.Contains(curExpStr, mo.Key+"~%") {
					// remove all new lines
					key = "[|%Meta." + mo.Key
					key = key + "~%"
					val = "[|%Conjoinments." + mo.Key + "%|]"
				} else if strings.Contains(curExpStr, mo.Key+"~&") {
					// remove leading new lines
					key = "[|&Meta." + mo.Key
					key = key + "~&"
					val = "[|&Conjoinments." + mo.Key + "&|]"
				} else {
					// default
					key = "[|Meta." + mo.Key
					key = key + "~"
					val = "[|Conjoinments." + mo.Key + "|]"
				}
			} else {
				// default
				if strings.Contains(curExpStr, mo.Key+"%") {
					// remove all new lines
					key = "[|%Meta." + mo.Key
					key = key + "%"
					val = "[|%Conjoinments." + mo.Key + "%|]"
				} else if strings.Contains(curExpStr, mo.Key+"&") {
					// remove leading new lines
					key = "[|&Meta." + mo.Key
					key = key + "&"
					val = "[|&Conjoinments." + mo.Key + "&|]"
				} else {
					// default
					key = "[|Meta." + mo.Key
					val = "[|Conjoinments." + mo.Key + "|]"
				}
			}
			key = key + "|]"
			curExpStr = strings.Replace(curExpStr, key, val, -1)
		}
	}
	tag.Expressions[2].ExpressionString = curExpStr
	// Body rewrites change child replacement strings too. Keep the parsed
	// child tree aligned with the exact source the interpreter will render.
	tag.Expressions[2].ExpressionSubTags = RecursiveSubTagParser(curExpStr, tag.Expressions[2].TagMeta)
	return tag
}

// ParseSegmentTags - parses out cases and sub-cases of LucidTags
func ParseSegmentTags(tags []*LucidTag, m MetaDict, segType string, headers bool) []*LucidTag {
	var reTags []*LucidTag
	checkStr := `if Type == "` + segType + `"`
	for _, tag := range tags {
		t := *tag
		cond := *t.Expressions[1]
		if cond.ExpressionString == "Headers" {
			if headers {
				reTags = append(reTags, tag)
			}
		} else if cond.ExpressionString == "Inputs" {
			if headers {
				reTags = append(reTags, tag)
			}
		} else if cond.ExpressionString == checkStr {
			tag = injectMetaConjoinments(tag, m)
			reTags = append(reTags, tag)
		}
	}
	return reTags
}

/*
==========================
LUCID COMPILER FUNCTIONS
==========================
*/

// CompileMainTags - compiles the root level tags for a given logiceme
func CompileMainTags(content []byte) []*LucidTag {
	var reTags []*LucidTag
	tagsBytes := parseMainTags(content)
	for _, tagBytes := range tagsBytes {
		lucidTagStr := strings.TrimPrefix(string(tagBytes), "{|/|")
		lucidTagStr = strings.TrimSuffix(lucidTagStr, "|/|}")
		tagType := determineTagType(lucidTagStr)
		var tagExpressions []*TagExpression
		tMeta := TagMeta{MetaDict{}}
		rStr := string(tagBytes)
		if tagType != "Variable" {
			parsedTagExpressions := lucidParser(lucidTagStr, "Expression")
			expressionType := determineExpressionType(parsedTagExpressions)
			for _, rawTagExpression := range parsedTagExpressions {
				tagExpressionStr := strings.TrimPrefix(rawTagExpression, "-{")
				tagExpressionStr = strings.TrimSuffix(tagExpressionStr, "}-")
				expressionSubTags := RecursiveSubTagParser(tagExpressionStr, &(tMeta))
				subExpressionOutput := ""
				if len(expressionSubTags) > 0 {
					subExpressionOutput = "SubTags"
				} else {
					subExpressionOutput = determineExpressionOutput(parsedTagExpressions)
				}
				tagExpression := NewTagExpression(rawTagExpression, tagExpressionStr, expressionType, subExpressionOutput, expressionSubTags, &(tMeta))
				tagExpressions = append(tagExpressions, tagExpression)
			}
		}
		lTag := NewLucidTag(rStr, lucidTagStr, tagType, tagExpressions, &(tMeta))
		reTags = append(reTags, lTag)
	}
	return reTags
}
