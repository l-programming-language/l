package Logiceme

import (
	"bytes"
	"encoding/json"
	"github.com/l-programming-language/l/lerrors"
	"strings"
)

// sourceMember records the single CONTENT order across heterogeneous members.
// Indices refer to the owner's Class/Object collections, never catalog order.
type sourceMember struct {
	Kind  string
	Index int
}
type structuralSpan struct{ start, openEnd, end int }

func framedStructuralTags(content []byte, wanted string) ([]structuralSpan, error) {
	var spans []structuralSpan
	depth, start, opening := 0, 0, 0
	for offset := 0; offset < len(content); {
		relative := bytes.IndexByte(content[offset:], '<')
		if relative < 0 {
			break
		}
		i := offset + relative
		if bytes.HasPrefix(content[i:], []byte("<|W|>")) {
			end, err := framedWordEnd(content, i)
			if err != nil {
				return nil, err
			}
			offset = end
			continue
		}
		if bytes.HasPrefix(content[i:], []byte("<JSON>")) {
			end, err := framedJSONEnd(content, i)
			if err != nil {
				return nil, err
			}
			offset = end
			continue
		}
		end, err := framedTagEnd(content, i)
		if err != nil {
			return nil, err
		}
		fields := strings.Fields(string(content[i+1 : end-1]))
		offset = end
		if len(fields) == 0 {
			continue
		}
		if fields[0] == "PARAGRAPH" || fields[0] == "/PARAGRAPH" {
			return nil, lerrors.InvalidInput.Errorf("PARAGRAPH was removed from AML; migrate its Objects and shared Attributes to Classes and explicit compositions")
		}
		if fields[0] == wanted {
			if depth == 0 {
				start, opening = i, end
			}
			depth++
		}
		if fields[0] == "/"+wanted {
			depth--
			if depth < 0 {
				return nil, lerrors.Syntax.Errorf("%s closing tag has no opening tag", wanted)
			}
			if depth == 0 {
				spans = append(spans, structuralSpan{start, opening, end})
			}
		}
	}
	if depth != 0 {
		return nil, lerrors.Syntax.Errorf("%s has no closing tag", wanted)
	}
	return spans, nil
}

type scannedSourceMember struct {
	text     *Text
	sentence *Sentence
}

func (sp *ScannedProgram) scanSourceMembers(tags []Tag, raw []byte, owner string) ([]scannedSourceMember, MetaDict, error) {
	outer, err := framedStructuralTags(raw, owner)
	if err != nil {
		return nil, MetaDict{}, err
	}
	if len(outer) != 1 || outer[0].start != 0 {
		return nil, MetaDict{}, lerrors.Syntax.Errorf("expected one %s", owner)
	}
	body := raw[outer[0].openEnd : len(raw)-len("</"+owner+">")]
	var members []scannedSourceMember
	var context bytes.Buffer
	cursor := 0
	for offset := 0; offset < len(body); {
		relative := bytes.IndexByte(body[offset:], '<')
		if relative < 0 {
			break
		}
		i := offset + relative
		if bytes.HasPrefix(body[i:], []byte("<JSON>")) {
			end, e := framedJSONEnd(body, i)
			if e != nil {
				return nil, MetaDict{}, e
			}
			offset = end
			continue
		}
		end, e := framedTagEnd(body, i)
		if e != nil {
			return nil, MetaDict{}, e
		}
		offset = end
		fields := strings.Fields(string(body[i+1 : end-1]))
		if len(fields) == 0 {
			continue
		}
		kind := fields[0]
		if kind == "CHAPTER" {
			return nil, MetaDict{}, lerrors.InvalidInput.Errorf("nested CHAPTER is not a Class; use TEXT")
		}
		if kind != "TEXT" && kind != "SENTENCE" {
			continue
		}
		spans, e := framedStructuralTags(body[i:], kind)
		if e != nil {
			return nil, MetaDict{}, e
		}
		if len(spans) == 0 {
			return nil, MetaDict{}, lerrors.Syntax.Errorf("incomplete %s", kind)
		}
		span := spans[0]
		child := body[i : i+span.end]
		vars, e := ParseTagVariables(DetermineTag(tags, kind), [][]byte{child[:span.openEnd]})
		if e != nil {
			return nil, MetaDict{}, e
		}
		if len(vars) != 1 || vars[0].GetVal("NAME") == "NotFound" || vars[0].GetVal("NAME") == "" {
			return nil, MetaDict{}, lerrors.InvalidInput.Errorf("%s requires a nonempty identity", kind)
		}
		name := vars[0].GetVal("NAME")
		if kind == "TEXT" {
			item, e := sp.ScanText(tags, child)
			if e != nil {
				return nil, MetaDict{}, e
			}
			item.Name = name
			members = append(members, scannedSourceMember{text: &item})
		} else {
			item, e := sp.ScanSentence(tags, child)
			if e != nil {
				return nil, MetaDict{}, e
			}
			item.Name = name
			members = append(members, scannedSourceMember{sentence: &item})
		}
		context.Write(body[cursor:i])
		offset = i + span.end
		cursor = offset
	}
	context.Write(body[cursor:])
	meta, e := ScanMetaArray(tags, context.Bytes(), nil)
	if e != nil {
		return nil, MetaDict{}, e
	}
	return members, meta, nil
}

// The generic host-AML BOOK regex consumes sibling books as one match. Source
// multi-application boundary frames books before using the ordinary ScanBook
// implementation. Word bodies, JSON, and quoted tag attributes are skipped
// intact; definition storage is then performed separately for each book.
func scanSourceProgram(scanner *Scanner) error {
	masked, restore, err := maskDocumentNativeValues(scanner.ProgramBytes)
	if err != nil {
		return err
	}
	scanner.ProgramBytes = masked
	if err := scanner.StripCommentBytes(); err != nil {
		return err
	}
	for _, kind := range []string{"PROGRAM", "STRUCTURE"} {
		spans, err := framedStructuralTags(scanner.ProgramBytes, kind)
		if err != nil {
			return err
		}
		if len(spans) != 1 {
			return lerrors.Syntax.Errorf("Rust document requires exactly one complete %s", kind)
		}
	}
	structures, _, err := ParseTags(scanner.Tags, scanner.ProgramBytes, "Tag", "STRUCTURE")
	if err != nil {
		return err
	}
	if len(structures) != 1 {
		return lerrors.Syntax.Errorf("Rust program requires one STRUCTURE tag")
	}
	program := ScannedProgram{Name: scanner.Name, Tags: scanner.Tags, Structure: structures, nativeSource: true}
	content := structures[0]
	bookStart := -1
	var opening []byte
	var words [][2]int
	for offset := 0; offset < len(content); {
		relative := bytes.IndexByte(content[offset:], '<')
		if relative < 0 {
			break
		}
		start := offset + relative
		if bytes.HasPrefix(content[start:], []byte("<|W|>")) {
			end, err := framedWordEnd(content, start)
			if err != nil {
				return err
			}
			if bookStart >= 0 {
				words = append(words, [2]int{start, end})
			}
			offset = end
			continue
		}
		if bytes.HasPrefix(content[start:], []byte("<JSON>")) {
			end, err := framedJSONEnd(content, start)
			if err != nil {
				return err
			}
			offset = end
			continue
		}
		end, err := framedTagEnd(content, start)
		if err != nil {
			return err
		}
		end-- // Existing book spans include the closing '>'.
		fields := strings.Fields(string(content[start+1 : end]))
		if len(fields) != 0 {
			switch fields[0] {
			case "BOOK":
				if bookStart >= 0 {
					return lerrors.Syntax.Errorf("Rust application books cannot be nested")
				}
				bookStart, opening = start, content[start:end+1]
				words = nil
			case "/BOOK":
				if bookStart < 0 {
					return lerrors.Syntax.Errorf("Rust application book closing tag has no opening tag")
				}
				// Reuse the exact framed word spans. Searching their native
				// bytes again for closing tags would expose literal markers.
				local := Scanner{Name: scanner.Name, Tags: scanner.Tags}
				var framed bytes.Buffer
				cursor := bookStart
				for _, word := range words {
					standIn := []byte("<|W|>" + GenerateUuid() + "</W|>")
					local.DefinitionStore = append(local.DefinitionStore, DefinitionStore{
						StandIn: standIn, Raw: bytes.Clone(content[word[0]:word[1]]),
					})
					framed.Write(content[cursor:word[0]])
					framed.Write(standIn)
					cursor = word[1]
				}
				framed.Write(content[cursor : end+1])
				local.ProgramBytes = framed.Bytes()
				bookProgram := ScannedProgram{Name: scanner.Name, Tags: scanner.Tags, DefinitionStore: local.DefinitionStore, nativeSource: true}
				book, err := bookProgram.ScanBook(local.ProgramBytes)
				if err != nil {
					return err
				}
				variables, err := ParseTagVariables(DetermineTag(program.Tags, "BOOK"), [][]byte{opening})
				if err != nil {
					return err
				}
				if len(variables) != 1 {
					return lerrors.Syntax.Errorf("Rust application book requires a name")
				}
				book.Name = variables[0].Get("NAME").Value
				restoreBookNativeValues(&book, restore)
				program.Books = append(program.Books, book)
				bookStart = -1
			}
		}
		offset = end + 1
	}
	if bookStart >= 0 || len(program.Books) == 0 {
		return lerrors.Syntax.Errorf("Rust program requires complete application BOOK tags")
	}
	scanner.Program = program
	return nil
}

// Read the original ordered records. ScanBook's ordinary MetaDict deliberately
// retains its legacy duplicate behavior for generic host-AML callers.
func bookSettings(tags []Tag, book Book) ([]MetaObj, error) {
	groups, _, err := ParseTags(tags, book.Program, "Tag", "SETTINGS")
	if err != nil {
		return nil, err
	}
	var settings []MetaObj
	for _, group := range groups {
		records, err := ScanJSON(tags, group)
		if err != nil {
			return nil, err
		}
		for _, record := range records {
			var setting MetaObj
			if err := json.Unmarshal(record, &setting); err != nil {
				return nil, lerrors.Syntax.Errorf("Rust configuration setting: %w", err)
			}
			settings = append(settings, setting)
		}
	}
	return settings, nil
}
