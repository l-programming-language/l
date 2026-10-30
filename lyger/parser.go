package Lyger

import (
	"bytes"
	"regexp"
	"strconv"

	"github.com/l-programming-language/l/lerrors"
)

type parseKind uint8

const (
	parseInvalid parseKind = iota
	parseLineStamp
	parseInputs
	parseKindCount
)

var parseKindNames = [parseKindCount]string{
	parseLineStamp: "lineStamp", parseInputs: "inputs",
}

var parserPatterns = [parseKindCount]*parserPattern{
	parseLineStamp: newParserPattern("$$HEAD<", ">HEAD$$"),
	parseInputs:    newParserPattern("<@", "@>"),
}

// Single-line frames end at their first closing delimiter. Multiline frames
// retain the original expression's repetition semantics. Literal scanning
// avoids a regex execution bitmap for ordinary source stamps and inputs.
type parserPattern struct {
	open, close []byte
	expansion   *regexp.Regexp
}

func newParserPattern(open, close string) *parserPattern {
	return &parserPattern{
		open: []byte(open), close: []byte(close),
		expansion: regexp.MustCompile(regexp.QuoteMeta(open) + `((?:.*?\r?\n?)*)` + regexp.QuoteMeta(close)),
	}
}

func (p *parserPattern) next(source []byte, offset int) ([4]int, bool) {
	start := bytes.Index(source[offset:], p.open)
	if start < 0 {
		return [4]int{}, false
	}
	start += offset
	body := start + len(p.open)
	end := bytes.Index(source[body:], p.close)
	if end < 0 {
		return [4]int{}, false
	}
	end += body
	if bytes.ContainsAny(source[body:end], "\r\n") {
		match := p.expansion.FindSubmatchIndex(source[start:])
		if match == nil {
			return [4]int{}, false
		}
		return [4]int{start + match[0], start + match[1], start + match[2], start + match[3]}, true
	}
	return [4]int{start, end + len(p.close), body, end}, true
}

func (kind parseKind) String() string {
	if kind >= parseKindCount {
		return "parseKind(" + strconv.Itoa(int(kind)) + ")"
	}
	return parseKindNames[kind]
}

func (kind parseKind) pattern() (*parserPattern, error) {
	if kind == parseInvalid || kind >= parseKindCount {
		return nil, lerrors.Internal.Errorf("invalid parser mode %d", kind)
	}
	return parserPatterns[kind], nil
}

// parser borrows content for the fixed lexical operations that remain in use.
type parser struct {
	contentBytes []byte
}

func newParser(content []byte) *parser {
	return &parser{contentBytes: content}
}

func (p *parser) parse(kind parseKind) ([][]byte, error) {
	pattern, err := kind.pattern()
	if err != nil {
		return [][]byte{}, err
	}
	var matches [][]byte
	for offset := 0; ; {
		match, ok := pattern.next(p.contentBytes, offset)
		if !ok {
			return matches, nil
		}
		matches = append(matches, p.contentBytes[match[0]:match[1]])
		offset = match[1]
	}
}

func (p *parser) replaceAll(kind parseKind, replacement []byte) ([]byte, error) {
	pattern, err := kind.pattern()
	if err != nil {
		return []byte{}, err
	}
	var result []byte
	offset := 0
	for {
		match, ok := pattern.next(p.contentBytes, offset)
		if !ok {
			return append(result, p.contentBytes[offset:]...), nil
		}
		result = append(result, p.contentBytes[offset:match[0]]...)
		result = pattern.expansion.Expand(result, replacement, p.contentBytes, match[:])
		offset = match[1]
	}
}

func (p *parser) removeLineStamps() ([]byte, error) {
	return p.replaceAll(parseLineStamp, nil)
}

// statementScanBytes makes internal stamps opaque to source sentence splitting.
// Positions stay identical, and parseStatements returns slices of the original
// input. Marker-like bytes inside source quotes or comments are ordinary text.
func (p *parser) statementScanBytes() ([]byte, error) {
	if !bytes.Contains(p.contentBytes, []byte("$$HEAD<")) {
		return p.contentBytes, nil
	}
	var masked []byte
	for i := 0; i < len(p.contentBytes); {
		if p.contentBytes[i] == '`' || p.contentBytes[i] == '"' {
			end, err := quotedEnd(p.contentBytes, i)
			if err != nil {
				return nil, err
			}
			i = end
			continue
		}
		if bytes.HasPrefix(p.contentBytes[i:], []byte("***")) {
			end, err := commentEnd(p.contentBytes, i)
			if err != nil {
				return nil, err
			}
			i = end
			continue
		}
		if bytes.HasPrefix(p.contentBytes[i:], []byte("$$HEAD<")) {
			if match, ok := parserPatterns[parseLineStamp].next(p.contentBytes, i); ok {
				if masked == nil {
					masked = bytes.Clone(p.contentBytes)
				}
				end := match[1]
				for j := i; j < end; j++ {
					masked[j] = '_'
				}
				i = end
				continue
			}
		}
		i++
	}
	if masked == nil {
		return p.contentBytes, nil
	}
	return masked, nil
}

// parseStatements parses and returns all L statements defined in contentBytes
func (p *parser) parseStatements() [][]byte {
	var reBytes [][]byte
	scanBytes, err := p.statementScanBytes()
	if err != nil {
		return reBytes
	}
	parts, err := sourceParts(scanBytes)
	if err != nil {
		return reBytes
	}
	for _, part := range parts {
		end := part.end
		if part.terminator >= 0 {
			end = part.terminator
		}
		reBytes = append(reBytes, bytes.TrimSpace(p.contentBytes[part.start:end]))
	}
	return reBytes
}

// parseTokens parses and returns a token byte slice for each token in contentBytes
func (p *parser) parseTokens() [][]byte {
	words, err := sourceWords(p.contentBytes)
	if err != nil {
		return nil
	}
	return words
}
