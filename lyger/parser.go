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

var parserPatterns = [parseKindCount]*regexp.Regexp{
	parseLineStamp: regexp.MustCompile(`\$\$HEAD\<((?:.*?\r?\n?)*)\>HEAD\$\$`),
	parseInputs:    regexp.MustCompile(`\<\@((?:.*?\r?\n?)*)\@\>`),
}

func (kind parseKind) String() string {
	if kind >= parseKindCount {
		return "parseKind(" + strconv.Itoa(int(kind)) + ")"
	}
	return parseKindNames[kind]
}

func (kind parseKind) pattern() (*regexp.Regexp, error) {
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
	return pattern.FindAll(p.contentBytes, -1), nil
}

func (p *parser) replaceAll(kind parseKind, replacement []byte) ([]byte, error) {
	pattern, err := kind.pattern()
	if err != nil {
		return []byte{}, err
	}
	return pattern.ReplaceAll(p.contentBytes, replacement), nil
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
			if match := parserPatterns[parseLineStamp].FindIndex(p.contentBytes[i:]); match != nil && match[0] == 0 {
				if masked == nil {
					masked = bytes.Clone(p.contentBytes)
				}
				end := i + match[1]
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
