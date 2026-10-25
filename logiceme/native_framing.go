package Logiceme

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// nativeValueMask keeps native metadata opaque while the structural
// scanners examine AML. Tokens are absent from the entire original input and
// restoration is one pass: restored source is never rescanned as metadata.
type nativeValueMask struct {
	prefix string
	pairs  []string
}

func newNativeValueMask(input []byte) *nativeValueMask {
	prefix := "RUSTNATIVE" + strings.ReplaceAll(GenerateUuid(), "-", "") + "VALUE"
	for bytes.Contains(input, []byte(prefix)) {
		prefix += "X"
	}
	return &nativeValueMask{prefix: prefix}
}

// nativeValueEnd uses AML's existing native-value delimiters. Other tildes,
// colons, backticks, and tag-like text are ordinary bytes inside the value.
func nativeValueEnd(input []byte, start int) (int, error) {
	for i := start; i < len(input); i++ {
		if bytes.HasPrefix(input[i:], []byte("~, ")) || bytes.HasPrefix(input[i:], []byte("~|}")) {
			return i, nil
		}
	}
	return 0, lerrors.Syntax.Errorf("Rust native metadata value has no closing delimiter")
}

func (m *nativeValueMask) values(input []byte) ([]byte, error) {
	var masked bytes.Buffer
	cursor := 0
	for offset := 0; offset < len(input); {
		if bytes.HasPrefix(input[offset:], []byte("/*|")) {
			end, err := framedCommentEnd(input, offset)
			if err != nil {
				return nil, err
			}
			offset = end
			continue
		}
		if !bytes.HasPrefix(input[offset:], []byte(": ~")) {
			offset++
			continue
		}
		start := offset + len(": ~")
		end, err := nativeValueEnd(input, start)
		if err != nil {
			return nil, err
		}
		token := fmt.Sprintf("%s%dEND", m.prefix, len(m.pairs)/2)
		m.pairs = append(m.pairs, token, string(input[start:end]))
		masked.Write(input[cursor:start])
		masked.WriteString(token)
		cursor, offset = end, end
	}
	masked.Write(input[cursor:])
	return masked.Bytes(), nil
}

// maskNativeValues masks only values in one glyph definition or framed
// Word. Field names and glyph composition remain visible to the usual parser.
func maskNativeValues(definition []byte) ([]byte, *strings.Replacer, error) {
	mask := newNativeValueMask(definition)
	masked, err := mask.values(definition)
	return masked, strings.NewReplacer(mask.pairs...), err
}

// maskDocumentNativeValues runs before comment stripping and STRUCTURE parsing.
// Only real Words are inspected; a fake Word inside a structural comment, JSON
// string, or quoted tag attribute cannot begin a native metadata value.
func maskDocumentNativeValues(input []byte) ([]byte, *strings.Replacer, error) {
	mask := newNativeValueMask(input)
	var masked bytes.Buffer
	cursor := 0
	for offset := 0; offset < len(input); {
		switch {
		case bytes.HasPrefix(input[offset:], []byte("/*|")):
			end, err := framedCommentEnd(input, offset)
			if err != nil {
				return nil, nil, err
			}
			offset = end
		case bytes.HasPrefix(input[offset:], []byte("<JSON>")):
			end, err := framedJSONEnd(input, offset)
			if err != nil {
				return nil, nil, err
			}
			offset = end
		case bytes.HasPrefix(input[offset:], []byte("<|W|>")):
			end, err := framedWordEnd(input, offset)
			if err != nil {
				return nil, nil, err
			}
			word, err := mask.values(input[offset:end])
			if err != nil {
				return nil, nil, err
			}
			masked.Write(input[cursor:offset])
			masked.Write(word)
			cursor, offset = end, end
		case input[offset] == '<':
			end, err := framedTagEnd(input, offset)
			if err != nil {
				return nil, nil, err
			}
			offset = end
		default:
			offset++
		}
	}
	masked.Write(input[cursor:])
	return masked.Bytes(), strings.NewReplacer(mask.pairs...), nil
}

func framedCommentEnd(input []byte, start int) (int, error) {
	end := bytes.Index(input[start+len("/*|"):], []byte("|*\\"))
	if end < 0 {
		return 0, lerrors.Syntax.Errorf("Rust AML comment has no closing delimiter")
	}
	return start + len("/*|") + end + len("|*\\"), nil
}

func framedJSONEnd(content []byte, start int) (int, error) {
	body := content[start+len("<JSON>"):]
	decoder := json.NewDecoder(bytes.NewReader(body))
	var raw json.RawMessage
	if err := decoder.Decode(&raw); err != nil {
		return 0, lerrors.Syntax.Errorf("Rust book framing JSON: %w", err)
	}
	end := start + len("<JSON>") + int(decoder.InputOffset())
	for end < len(content) && (content[end] == ' ' || content[end] == '\t' || content[end] == '\r' || content[end] == '\n') {
		end++
	}
	if !bytes.HasPrefix(content[end:], []byte("</JSON>")) {
		return 0, lerrors.Syntax.Errorf("Rust book framing JSON has no closing tag")
	}
	return end + len("</JSON>"), nil
}

func framedTagEnd(content []byte, start int) (int, error) {
	quoted := false
	for end := start + 1; end < len(content); end++ {
		switch content[end] {
		case '\\':
			if quoted && end+1 < len(content) {
				end++
			}
		case '"':
			quoted = !quoted
		case '>':
			if !quoted {
				return end + 1, nil
			}
		}
	}
	return 0, lerrors.Syntax.Errorf("Rust book framing has an unterminated tag")
}

func restoreBookNativeValues(book *Book, restore *strings.Replacer) {
	var sentences = func(items []Sentence) {
		for si := range items {
			for wi := range items[si].Words {
				for gi := range items[si].Words[wi].Syllables {
					syllable := &items[si].Words[wi].Syllables[gi]
					syllable.Name = restore.Replace(syllable.Name)
				}
			}
		}
	}
	var text func(*Text)
	text = func(item *Text) {
		sentences(item.Sentences)
		for i := range item.SubTexts {
			text(&item.SubTexts[i])
		}
	}
	var chapter func(*Chapter)
	chapter = func(item *Chapter) {
		sentences(item.Sentences)
		for i := range item.Texts {
			text(&item.Texts[i])
		}
		for i := range item.SubChapters {
			chapter(&item.SubChapters[i])
		}
	}
	for i := range book.Content {
		chapter(&book.Content[i])
	}
}

// Native inputs use the existing : ~value~ metadata framing. Their bytes may
// contain word/book markers or backticks; none are structure while inside the
// input. This only finds the word boundary, leaving its ordinary parsing and
// validation to ScanBook and the existing syllable/glyph pipeline.
func framedWordEnd(content []byte, start int) (int, error) {
	for i := start + len("<|W|>"); i < len(content); i++ {
		if bytes.HasPrefix(content[i:], []byte("/*|")) {
			end, err := framedCommentEnd(content, i)
			if err != nil {
				return 0, err
			}
			i = end - 1
			continue
		}
		if bytes.HasPrefix(content[i:], []byte(": ~")) {
			end, err := nativeValueEnd(content, i+len(": ~"))
			if err != nil {
				return 0, err
			}
			i = end
			continue
		}
		if bytes.HasPrefix(content[i:], []byte("</W|>")) {
			return i + len("</W|>"), nil
		}
	}
	return 0, lerrors.Syntax.Errorf("Rust application word has no closing tag")
}
