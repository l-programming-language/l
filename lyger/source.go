package Lyger

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// SourcePosition refers to the original file. Offsets and columns count bytes;
// lines and columns are one-based, and offsets are zero-based.
type SourcePosition struct {
	Offset int
	Line   int
	Column int
}

// SourceSpan has an exclusive end. Paragraph and Order are local to Path.
type SourceSpan struct {
	Path      string
	Start     SourcePosition
	End       SourcePosition
	Paragraph int
	Order     int
}

type sourceOccurrenceID uint64

// paragraphLayout describes original physical layout, before import expansion
// or whitespace trimming. It does not assign a lexical or runtime scope.
type paragraphLayout struct {
	Number       int
	FirstCode    SourcePosition
	Encapsulated bool
}

type SourceLiteral struct {
	ID   string
	Raw  []byte
	Span SourceSpan
}

// SourceStatement keeps exact source separate from the stamped/cached parser
// input. A synthesized import alias retains the span of its original import.
type SourceStatement struct {
	Raw        []byte
	Prepared   []byte
	Span       SourceSpan
	Terminator []byte
	occurrence sourceOccurrenceID
	Unit       *SourceUnit
	Origin     SourceOrigin
	imported   *dictionaryImport
}

type PreparedSource struct {
	Path          string
	Original      []byte
	PreparedBytes []byte
	TextCache     map[string][]byte
	Statements    []SourceStatement
	Literals      []SourceLiteral
	paragraphs    []paragraphLayout
}

type sourcePart struct {
	start, end, codeStart, terminator int
}

// quotedEnd treats escaped quote bytes as literal data. It is shared by source
// splitting and import tokenization, so paths containing spaces remain one word.
func quotedEnd(data []byte, start int) (int, error) {
	quote := data[start]
	for i := start + 1; i < len(data); i++ {
		if data[i] == '\\' && i+1 < len(data) {
			i++
			continue
		}
		if data[i] == quote {
			return i + 1, nil
		}
	}
	return len(data), lerrors.Syntax.Errorf("unterminated %c literal at byte %d", quote, start)
}

func commentEnd(data []byte, start int) (int, error) {
	if end := bytes.Index(data[start+3:], []byte("***")); end >= 0 {
		return start + 3 + end + 3, nil
	}
	return len(data), lerrors.Syntax.Errorf("unterminated comment at byte %d", start)
}

func sourceSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\r' || b == '\n'
}

// statementEnd accepts a period followed by whitespace, a comment, or EOF.
// Separators belong to the preceding sentence; an uninterrupted token's dot
// remains part of that token. The caller skips literals before checking dots.
func statementEnd(data []byte, period int) (int, bool) {
	if next := period + 1; next < len(data) && !sourceSpace(data[next]) && !bytes.HasPrefix(data[next:], []byte("***")) {
		return 0, false
	}
	for i := period + 1; i < len(data); {
		if data[i] == '\n' {
			return i + 1, true
		}
		if sourceSpace(data[i]) {
			i++
			continue
		}
		if bytes.HasPrefix(data[i:], []byte("***")) {
			end, err := commentEnd(data, i)
			if err != nil {
				return 0, false
			}
			i = end
			continue
		}
		return i, true
	}
	return len(data), true
}

func sourceParts(data []byte) ([]sourcePart, error) {
	var parts []sourcePart
	start, codeStart := 0, -1
	for i := 0; i < len(data); {
		if bytes.HasPrefix(data[i:], []byte("***")) {
			end, err := commentEnd(data, i)
			if err != nil {
				return nil, err
			}
			i = end
			continue
		}
		if codeStart < 0 && !sourceSpace(data[i]) {
			codeStart = i
		}
		if data[i] == '`' || data[i] == '"' {
			end, err := quotedEnd(data, i)
			if err != nil {
				return nil, err
			}
			i = end
			continue
		}
		if data[i] == '.' {
			if end, ok := statementEnd(data, i); ok {
				parts = append(parts, sourcePart{start, end, codeStart, i})
				start, codeStart, i = end, -1, end
				continue
			}
		}
		i++
	}
	if codeStart >= 0 {
		parts = append(parts, sourcePart{start, len(data), codeStart, -1})
	}
	return parts, nil
}

type sourceLine struct {
	start     int
	paragraph int
}

// sourceIndex belongs to one preparation. It records byte offsets and paragraph
// identity without keeping the source buffer or adding state to PreparedSource.
type sourceIndex []sourceLine

func (index sourceIndex) location(offset int) (SourcePosition, int) {
	line := sort.Search(len(index), func(i int) bool { return index[i].start > offset }) - 1
	return SourcePosition{Offset: offset, Line: line + 1, Column: offset - index[line].start + 1}, index[line].paragraph
}

// Paragraph boundaries are physical whitespace-only lines outside comments and
// literals. Every newline advances the byte-position index, while a comment or
// quoted span keeps its lines occupied for paragraph purposes.
func indexSource(data []byte) (sourceIndex, error) {
	index := sourceIndex{{paragraph: 1}}
	paragraph, lineStart, lineOccupied := 1, 0, false
	for i := 0; i < len(data); {
		if bytes.HasPrefix(data[i:], []byte("***")) || data[i] == '`' || data[i] == '"' {
			var end int
			var err error
			if bytes.HasPrefix(data[i:], []byte("***")) {
				end, err = commentEnd(data, i)
			} else {
				end, err = quotedEnd(data, i)
			}
			if err != nil {
				return nil, err
			}
			for j := i; j < end; j++ {
				if data[j] == '\n' {
					lineStart = j + 1
					index = append(index, sourceLine{start: lineStart, paragraph: paragraph})
				}
			}
			lineOccupied, i = true, end
			continue
		}
		if data[i] == '\n' {
			if !lineOccupied && len(bytes.TrimSpace(data[lineStart:i])) == 0 {
				paragraph++
			}
			lineStart, lineOccupied = i+1, false
			index = append(index, sourceLine{start: lineStart, paragraph: paragraph})
		} else if !sourceSpace(data[i]) {
			lineOccupied = true
		}
		i++
	}
	return index, nil
}

// PrepareSource preserves provenance before caching text or removing comments.
// PreparedBytes is the internal parser view; it is not a compatibility output.
func PrepareSource(path string, content []byte) (*PreparedSource, error) {
	p := &PreparedSource{Path: path, Original: bytes.Clone(content), TextCache: make(map[string][]byte)}
	parts, err := sourceParts(content)
	if err != nil {
		return p, fmt.Errorf("%s: %w", path, err)
	}
	index, err := indexSource(content)
	if err != nil {
		return p, fmt.Errorf("%s: %w", path, err)
	}
	for order, part := range parts {
		start, paragraph := index.location(part.codeStart)
		end, _ := index.location(part.end)
		if len(p.paragraphs) == 0 || p.paragraphs[len(p.paragraphs)-1].Number != paragraph {
			lineStart := start.Offset - start.Column + 1
			// Four leading ASCII spaces open the single supported level. Further
			// indentation does not add levels, and continuation lines cannot open one.
			encapsulated := bytes.HasPrefix(content[lineStart:part.codeStart], []byte("    "))
			p.paragraphs = append(p.paragraphs, paragraphLayout{Number: paragraph, FirstCode: start, Encapsulated: encapsulated})
		}
		span := SourceSpan{Path: path, Start: start, End: end, Paragraph: paragraph, Order: order}
		statement := SourceStatement{Raw: bytes.Clone(content[part.start:part.end]), Span: span}
		bodyEnd := part.end
		if part.terminator >= 0 {
			bodyEnd = part.terminator
			statement.Terminator = bytes.Clone(content[part.terminator:part.end])
		}
		var body []byte
		for i := part.start; i < bodyEnd; {
			if bytes.HasPrefix(content[i:], []byte("***")) {
				end, err := commentEnd(content, i)
				if err != nil {
					return p, fmt.Errorf("%s: %w", path, err)
				}
				body = append(body, ' ')
				i = end
				continue
			}
			if content[i] == '`' {
				end, err := quotedEnd(content, i)
				if err != nil {
					return p, fmt.Errorf("%s: %w", path, err)
				}
				id, raw := newCacheUuid(), bytes.Clone(content[i:end])
				p.TextCache[id] = raw
				literalStart, _ := index.location(i)
				literalEnd, _ := index.location(end)
				literalSpan := SourceSpan{Path: path, Start: literalStart, End: literalEnd, Paragraph: span.Paragraph, Order: order}
				p.Literals = append(p.Literals, SourceLiteral{ID: id, Raw: bytes.Clone(raw), Span: literalSpan})
				body = append(body, id...)
				i = end
				continue
			}
			if content[i] == '"' {
				end, err := quotedEnd(content, i)
				if err != nil {
					return p, fmt.Errorf("%s: %w", path, err)
				}
				body = append(body, content[i:end]...)
				i = end
				continue
			}
			body = append(body, content[i])
			i++
		}
		body = bytes.TrimSpace(body)
		statement.Prepared = append(newLineStamp(span.Start.Line, path), body...)
		statement.Prepared = append(statement.Prepared, '.', '\n')
		p.Statements = append(p.Statements, statement)
		p.PreparedBytes = append(p.PreparedBytes, statement.Prepared...)
	}
	return p, nil
}

// sourceWords splits whitespace outside literal/identifier quotes, retaining
// exact token bytes. Error-aware callers validate sourceParts first.
func sourceWords(content []byte) ([][]byte, error) {
	var words [][]byte
	for i := 0; i < len(content); {
		for i < len(content) && sourceSpace(content[i]) {
			i++
		}
		start := i
		for i < len(content) && !sourceSpace(content[i]) {
			if content[i] == '`' || content[i] == '"' {
				end, err := quotedEnd(content, i)
				if err != nil {
					return nil, err
				}
				i = end
			} else {
				i++
			}
		}
		if start < i {
			words = append(words, content[start:i])
		}
	}
	return words, nil
}

func statementBody(raw []byte) []byte {
	body, err := removeFileStamp(raw)
	if err != nil {
		return nil
	}
	parts, err := sourceParts(body)
	if err != nil || len(parts) != 1 {
		return nil
	}
	if parts[0].terminator >= 0 {
		body = body[:parts[0].terminator]
	}
	var stripped []byte
	for i := 0; i < len(body); {
		if bytes.HasPrefix(body[i:], []byte("***")) {
			var err error
			i, err = commentEnd(body, i)
			if err != nil {
				return nil
			}
			stripped = append(stripped, ' ')
		} else if body[i] == '`' || body[i] == '"' {
			end, err := quotedEnd(body, i)
			if err != nil {
				return nil
			}
			stripped = append(stripped, body[i:end]...)
			i = end
		} else {
			stripped = append(stripped, body[i])
			i++
		}
	}
	return bytes.TrimSpace(stripped)
}

func possessiveOwner(word string) (string, bool) {
	for _, suffix := range []string{"'s", "’s", "s'", "s’"} {
		if len(word) < len(suffix) || !strings.EqualFold(word[len(word)-len(suffix):], suffix) {
			continue
		}
		width := len(suffix)
		if strings.HasPrefix(suffix, "s") {
			width-- // A plural owner's final s/S is part of its name.
		}
		return word[:len(word)-width], true
	}
	return word, false
}
