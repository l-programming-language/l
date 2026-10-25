package Lyger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// amlRenderer writes two views of the same graph. The observable form retains
// the CLI's unquoted syllables; the backend form supplies logeme's delimiters.
// Neither view is used as an input to the other, or aliases the source graph.
type amlRenderer struct {
	observable bytes.Buffer
	backend    bytes.Buffer
}

func (r *amlRenderer) write(text string) {
	r.observable.WriteString(text)
	r.backend.WriteString(text)
}

func (r *amlRenderer) json(value interface{}) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return lerrors.Internal.Wrap(err, "")
	}
	r.write("<JSON>" + string(encoded) + "</JSON>\n")
	return nil
}

type amlAppMetadata struct {
	AppName          string
	AppLang          string
	AppVersion       string
	AppGenDir        string
	StructureVersion string `json:",omitempty"`
	Deps             string `json:",omitempty"`
}

type amlServiceMetadata struct {
	Name           string
	Type           string
	Package        string
	Location       string
	Parent         string
	GenericObjects []string
	Dimensions     []string
	Target         *sourceTarget `json:",omitempty"`
}

type amlObjectMetadata struct {
	Name   string
	Type   string
	Module *sourceModule `json:",omitempty"`
}

// renderProgram serializes the validated Service / recursive Class / Object graph.
// It publishes neither view unless the complete graph renders successfully.
func renderProgram(program *Program, apps []*App) (observable []byte, backend []byte, err error) {
	if program == nil {
		return nil, nil, lerrors.Internal.Errorf("AML rendering requires an application book")
	}
	applications := make(map[string]*App, len(apps))
	for _, app := range apps {
		if app == nil {
			return nil, nil, lerrors.Internal.Errorf("AML rendering received a nil application")
		}
		name := app.buildName()
		if name == "" {
			name = app.Meta["AppName"]
		}
		if name == "" || applications[name] != nil {
			return nil, nil, lerrors.Internal.Errorf("AML application identity is empty or duplicated: %q", name)
		}
		applications[name] = app
	}
	if len(program.Books) != len(applications) {
		return nil, nil, lerrors.Internal.Errorf("AML has %d books for %d applications", len(program.Books), len(applications))
	}
	var renderer amlRenderer
	renderer.write("<PROGRAM>\n<STRUCTURE>\n")
	seen := make(map[string]bool)
	for _, book := range program.Books {
		if book == nil {
			return nil, nil, lerrors.Internal.Errorf("AML rendering received a nil book")
		}
		app := applications[book.Name]
		if app == nil || seen[book.Name] {
			return nil, nil, lerrors.Internal.Errorf("AML book %q has no unique application", book.Name)
		}
		seen[book.Name] = true
		if err := renderer.book(book, app); err != nil {
			return nil, nil, fmt.Errorf("AML book %q: %w", book.Name, err)
		}
	}
	renderer.write("</STRUCTURE>\n</PROGRAM>")
	return renderer.observable.Bytes(), renderer.backend.Bytes(), nil
}

func (r *amlRenderer) book(book *book, app *App) error {
	if err := validateAMLName(book.Name); err != nil {
		return err
	}
	graph, err := prepareAMLGraph(app)
	if err != nil {
		return err
	}
	r.write("<BOOK " + strconv.Quote(book.Name) + ">\n<OUTLINE>\n")
	meta := amlAppMetadata{AppName: app.Meta["AppName"], AppLang: app.Meta["AppLang"], AppVersion: app.Meta["AppVersion"], AppGenDir: app.Meta["AppGenDir"], StructureVersion: app.Meta["StructureVersion"], Deps: app.Meta["Deps"]}
	if meta.StructureVersion != "" && meta.StructureVersion != "2" {
		return lerrors.InvalidInput.Errorf("unsupported StructureVersion %q", meta.StructureVersion)
	}
	if len(graph.classes) != 0 {
		meta.StructureVersion = "2"
	}
	for _, service := range graph.services {
		if service.Target != nil {
			meta.StructureVersion = "2"
		}
	}
	if meta.StructureVersion == "2" {
		for i := range graph.services {
			if graph.services[i].Parent == noParent {
				graph.services[i].Parent = ""
			}
		}
	}
	if err := r.json(struct{ AppMeta amlAppMetadata }{meta}); err != nil {
		return err
	}
	r.write("<JSON>{\"Integrations\":[]}</JSON>\n<JSON>{\"Resources\":[]}</JSON>\n<JSON>{\"Datasources\":[]}</JSON>\n")
	if err := r.json(struct{ Services []amlServiceMetadata }{graph.services}); err != nil {
		return err
	}
	if err := r.json(struct{ Classes []amlClassMetadata }{graph.classes}); err != nil {
		return err
	}
	encodedObjects, err := json.Marshal(graph.objects)
	if err != nil {
		return lerrors.Internal.Wrap(err, "")
	}
	// The space after Objects is part of the existing observable document format.
	r.write("<JSON>{\"Objects\": " + string(encodedObjects) + "}</JSON>\n</OUTLINE>\n<SETTINGS>\n")
	keys := make([]string, 0, len(app.Settings))
	for key := range app.Settings {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		setting := struct {
			Key   string `json:"key"`
			Value string `json:"value"`
			Type  string `json:"type"`
		}{key, app.Settings[key], "String"}
		if err := r.json(setting); err != nil {
			return err
		}
	}
	r.write("</SETTINGS>\n<CONTENT>\n")
	seenChapters := make(map[string]bool)
	for _, chapter := range book.Chapters {
		if chapter == nil {
			return lerrors.InvalidInput.Errorf("content contains a nil Chapter")
		}
		expected, exists := graph.serviceMembers[chapter.Name]
		if !exists || seenChapters[chapter.Name] {
			return lerrors.InvalidInput.Errorf("content Chapter %q has no unique outlined Service", chapter.Name)
		}
		seenChapters[chapter.Name] = true
		if len(chapter.children) != 0 {
			return lerrors.Unsupported.Errorf("Chapter %q contains unsupported child Chapters", chapter.Name)
		}
		r.write("<CHAPTER " + strconv.Quote(chapter.Name) + ">\n")
		members, err := orderedAMLSourceMembers(chapter.sentences, chapter.Texts, chapter.members)
		if err != nil {
			return err
		}
		if err := r.sourceMembers(members, expected, graph); err != nil {
			if meta.StructureVersion == "2" {
				return fmt.Errorf("Chapter %q: %w", chapter.Name, err)
			}
			return err
		}
		r.write("</CHAPTER>\n")
	}
	if len(seenChapters) != len(graph.serviceMembers) {
		return lerrors.InvalidInput.Errorf("outlined Service has no content Chapter")
	}
	if meta.StructureVersion == "2" && len(book.Sentences) != len(graph.rootObjects) {
		return lerrors.InvalidInput.Errorf("root Object catalog and Sentence content disagree")
	}
	for i, sentence := range book.Sentences {
		if meta.StructureVersion == "2" && (sentence == nil || sentence.Name != graph.rootObjects[i].ID) {
			return lerrors.InvalidInput.Errorf("root Object content order disagrees with its catalog")
		}
		if err := r.sentence(sentence, graph.objectTypes); err != nil {
			return err
		}
	}
	r.write("</CONTENT>\n</BOOK>\n")
	return nil
}

// sourceMembers checks the two projections against the same ordered graph.
func (r *amlRenderer) sourceMembers(members []amlSourceMember, expected []sourceMember, graph *amlGraph) error {
	if len(members) != len(expected) {
		return lerrors.InvalidInput.Errorf("source membership and content cardinality disagree")
	}
	for i, member := range members {
		if expected[i].Object != nil {
			if member.Sentence == nil || member.Sentence.Name != expected[i].Object.ID {
				return lerrors.InvalidInput.Errorf("source Object order disagrees with content")
			}
			if err := r.sentence(member.Sentence, graph.objectTypes); err != nil {
				return err
			}
			continue
		}
		if member.Text == nil || member.Text.Name != expected[i].Class.ID {
			return lerrors.InvalidInput.Errorf("source Class order disagrees with content")
		}
		group := member.Text
		r.write("<TEXT " + strconv.Quote(group.Name) + ">\n")
		children, err := orderedAMLSourceMembers(group.sentences, group.children, group.members)
		if err != nil {
			return err
		}
		if err := r.sourceMembers(children, graph.classMembers[group.Name], graph); err != nil {
			return fmt.Errorf("Class %q: %w", group.Name, err)
		}
		r.write("</TEXT>\n")
	}
	return nil
}

func (r *amlRenderer) sentence(sentence *sentence, objects map[string]string) error {
	if sentence == nil {
		return lerrors.Internal.Errorf("content contains a nil sentence")
	}
	if _, found := objects[sentence.Name]; !found {
		return lerrors.Internal.Errorf("sentence %q has no outlined object", sentence.Name)
	}
	if err := validateAMLName(sentence.Name); err != nil {
		return err
	}
	r.write("<SENTENCE " + strconv.Quote(sentence.Name) + ">\n")
	for _, word := range sentence.Words {
		if word == nil || len(word.Syllables) == 0 {
			return lerrors.Internal.Errorf("sentence %q contains an empty or nil word", sentence.Name)
		}
		r.write("<|W|>\n")
		for _, syllable := range word.Syllables {
			if syllable == nil || len(bytes.TrimSpace(syllable.CompiledBytes)) == 0 {
				return lerrors.Internal.Errorf("sentence %q contains an empty or nil syllable", sentence.Name)
			}
			r.write("<|S|>")
			r.observable.Write(syllable.CompiledBytes)
			r.backend.WriteByte('`')
			r.backend.Write(syllable.CompiledBytes)
			r.backend.WriteByte('`')
			r.write("</S|>\n")
		}
		r.write("</W|>\n")
	}
	r.write("</SENTENCE>\n")
	return nil
}

func validateAMLName(name string) error {
	if name == "" || strings.ContainsAny(name, "\"\r\n<>") {
		return lerrors.InvalidInput.Errorf("invalid AML element name %q", name)
	}
	return nil
}
