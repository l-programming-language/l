package Lyger

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func renderTestGraph() (*Program, []*App) {
	obj := &object{ID: "endpoint", Meta: map[string]string{"Type": "root"}}
	app := &App{ID: "Atlas", Meta: map[string]string{"AppName": "Atlas", "AppLang": "rust", "AppVersion": "2.1", "AppGenDir": "/out"}, Settings: map[string]string{"Host": "https://atlas.test"}, Services: []*service{{ID: "api", Meta: map[string]string{"Type": "root", "Package": "atlas", "Location": "/api", "Parent": "None"}, Objects: []*object{obj}}}, Objects: []*object{obj}}
	program := &Program{Books: []*book{{Name: "Atlas", Chapters: []*chapter{{Name: "api", sentences: []*sentence{{Name: "endpoint", Words: []*w{{Syllables: []*s{{CompiledBytes: []byte("A")}, {CompiledBytes: []byte("W^BOD^E{|Statement: ~consume(42)~|}")}}}}}}}}}}}
	return program, []*App{app}
}

func TestRenderProgramViewsOwnTheirBytes(t *testing.T) {
	program, apps := renderTestGraph()
	syllable := program.Books[0].Chapters[0].sentences[0].Words[0].Syllables[1]
	original := append([]byte(nil), syllable.CompiledBytes...)
	observable, backend, err := renderProgram(program, apps)
	if err != nil {
		t.Fatal(err)
	}
	originalBackend := append([]byte(nil), backend...)
	index := bytes.Index(observable, original)
	if index < 0 {
		t.Fatal("source expression was not rendered")
	}
	observable[index] = 'X'
	if !bytes.Equal(backend, originalBackend) || !bytes.Equal(syllable.CompiledBytes, original) {
		t.Fatal("observable output aliases the backend or graph")
	}
	backend[bytes.Index(backend, original)] = 'Y'
	if !bytes.Equal(syllable.CompiledBytes, original) || program.AMLBytes != nil || program.BackendBytes != nil {
		t.Fatal("rendering mutated graph state")
	}
	syllable.CompiledBytes = []byte("BOD^E{|Statement: ~consume(73)~|}")
	changed, _, err := renderProgram(program, apps)
	if err != nil || !bytes.Contains(changed, []byte("consume(73)")) || bytes.Contains(changed, []byte("consume(42)")) {
		t.Fatalf("changed source expression was not rendered: %v", err)
	}
}

func TestRenderProgramMetadataIsValidJSONAndDeterministic(t *testing.T) {
	program, apps := renderTestGraph()
	apps[0].Meta["AppVersion"] = "雪 \"quoted\" <version>"
	apps[0].Settings = map[string]string{"Zone": "west", "Host": "https://atlas.test/?q=\"雪\"&r=1"}
	first, _, err := renderProgram(program, apps)
	if err != nil {
		t.Fatal(err)
	}
	for run := 0; run < 10; run++ {
		actual, _, err := renderProgram(program, apps)
		if err != nil || !bytes.Equal(actual, first) {
			t.Fatalf("metadata rendering is not deterministic: %v", err)
		}
	}
	remaining := string(first)
	count := 0
	for {
		start := strings.Index(remaining, "<JSON>")
		if start < 0 {
			break
		}
		remaining = remaining[start+len("<JSON>"):]
		end := strings.Index(remaining, "</JSON>")
		if end < 0 || !json.Valid([]byte(remaining[:end])) {
			t.Fatal("outline/settings contains invalid JSON")
		}
		if count == 0 {
			var meta struct{ AppMeta amlAppMetadata }
			if err := json.Unmarshal([]byte(remaining[:end]), &meta); err != nil || meta.AppMeta.AppVersion != apps[0].Meta["AppVersion"] {
				t.Fatal("JSON escaping altered a metadata value")
			}
		}
		remaining = remaining[end+len("</JSON>"):]
		count++
	}
	if count != 9 || bytes.Index(first, []byte(`"key":"Host"`)) > bytes.Index(first, []byte(`"key":"Zone"`)) {
		t.Fatal("settings were omitted or emitted outside deterministic key order")
	}
}

func TestRenderProgramUsesBookOrderAndApplicationIdentity(t *testing.T) {
	program, apps := renderTestGraph()
	second, others := renderTestGraph()
	second.Books[0].Name = "Beacon"
	others[0].ID, others[0].Meta["AppName"] = "Beacon", "Beacon"
	program.Books = append(program.Books, second.Books[0])
	// Lookup follows book identity even when application storage has a different order.
	apps = append(others, apps...)
	observable, _, err := renderProgram(program, apps)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Count(observable, []byte("<BOOK ")) != 2 || bytes.Index(observable, []byte(`<BOOK "Atlas">`)) > bytes.Index(observable, []byte(`<BOOK "Beacon">`)) {
		t.Fatal("renderer did not preserve the program's book order")
	}
}

func TestRenderProgramRejectsIncompleteGraph(t *testing.T) {
	for _, example := range []struct {
		name string
		edit func(*Program, []*App) (*Program, []*App)
	}{
		{"nil program", func(_ *Program, apps []*App) (*Program, []*App) { return nil, apps }},
		{"missing application", func(program *Program, _ []*App) (*Program, []*App) { return program, nil }},
		{"mismatched identity", func(program *Program, apps []*App) (*Program, []*App) {
			program.Books[0].Name = "Missing"
			return program, apps
		}},
		{"nil syllable", func(program *Program, apps []*App) (*Program, []*App) {
			program.Books[0].Chapters[0].sentences[0].Words[0].Syllables[0] = nil
			return program, apps
		}},
		{"unsupported content", func(program *Program, apps []*App) (*Program, []*App) {
			program.Books[0].Chapters[0].Texts = []*text{{Name: "class"}}
			return program, apps
		}},
	} {
		t.Run(example.name, func(t *testing.T) {
			program, apps := renderTestGraph()
			program, apps = example.edit(program, apps)
			observable, backend, err := renderProgram(program, apps)
			if err == nil || observable != nil || backend != nil {
				t.Fatal("invalid graph returned successful or partial output")
			}
		})
	}
}

func TestRenderProgramObservableFormatAndBackendDelimiters(t *testing.T) {
	program, apps := renderTestGraph()
	observable, backend, err := renderProgram(program, apps)
	if err != nil {
		t.Fatal(err)
	}
	want := `<PROGRAM>
<STRUCTURE>
<BOOK "Atlas">
<OUTLINE>
<JSON>{"AppMeta":{"AppName":"Atlas","AppLang":"rust","AppVersion":"2.1","AppGenDir":"/out"}}</JSON>
<JSON>{"Integrations":[]}</JSON>
<JSON>{"Resources":[]}</JSON>
<JSON>{"Datasources":[]}</JSON>
<JSON>{"Services":[{"Name":"api","Type":"root","Package":"atlas","Location":"/api","Parent":"None","GenericObjects":["endpoint"],"Dimensions":[]}]}</JSON>
<JSON>{"Classes":[]}</JSON>
<JSON>{"Objects": [{"Name":"endpoint","Type":"root"}]}</JSON>
</OUTLINE>
<SETTINGS>
<JSON>{"key":"Host","value":"https://atlas.test","type":"String"}</JSON>
</SETTINGS>
<CONTENT>
<CHAPTER "api">
<SENTENCE "endpoint">
<|W|>
<|S|>A</S|>
<|S|>W^BOD^E{|Statement: ~consume(42)~|}</S|>
</W|>
</SENTENCE>
</CHAPTER>
</CONTENT>
</BOOK>
</STRUCTURE>
</PROGRAM>`
	if string(observable) != want {
		t.Fatalf("observable AML differs:\n%s", observable)
	}
	wantBackend := strings.ReplaceAll(strings.ReplaceAll(want, "<|S|>", "<|S|>`"), "</S|>", "`</S|>")
	if string(backend) != wantBackend {
		t.Fatalf("backend AML differs:\n%s", backend)
	}
	if bytes.HasSuffix(observable, []byte("\n")) || bytes.HasSuffix(backend, []byte("\n")) {
		t.Fatal("renderer added a newline after the final tag")
	}
}
