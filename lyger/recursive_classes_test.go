package Lyger

import (
	"bytes"
	"encoding/json"
	Logiceme "github.com/l-programming-language/l/logiceme"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestFrontendFeatureTypeRemoved(t *testing.T) {
	cfg := canonicalTestOptions()
	vocab, err := loadLexicon(cfg.Resources, cfg.Language)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := vocab.Nouns["feature"]; ok {
		t.Error("removed Feature remains a built-in noun")
	}
	if _, ok := parseValueType("Feature"); ok {
		t.Error("removed Feature remains a value type")
	}
	class := vocab.Nouns["class"]
	encoded, err := json.Marshal(class)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(encoded, []byte("features")) || bytes.Contains(encoded, []byte("inputs[2]")) {
		t.Errorf("unexpected Class constructor argument emitted: %s", encoded)
	}
	result, err := NewLygerService(cfg).Compile([]byte("\"Feature\" is `ordinary text`.\n"), "feature-name.l")
	if err != nil || result == nil {
		t.Fatalf("Feature spelling cannot become an ordinary user declaration: %v", err)
	}
}

func TestFrontendSurvivingValueTypeNames(t *testing.T) {
	names := []string{"", "Thing", "Text", "Number", "Truth", "Bytes", "Pointer", "Lexeme", "App", "Service", "Class", "Object", "Attribute", "Definition", "Function", "Property", "Operation", "List", "Dictionary", "Error", "ID", "None", "Unknown", "Null"}
	if int(valueTypeCount) != len(names) {
		t.Fatalf("value type inventory has %d, want %d", valueTypeCount, len(names))
	}
	for _, name := range names {
		value, ok := parseValueType(name)
		if !ok {
			t.Fatalf("surviving type %q disappeared", name)
		}
		encoded, err := value.MarshalText()
		if err != nil || string(encoded) != name || value.String() != name {
			t.Fatalf("surviving type %q changed: %s, %v", name, encoded, err)
		}
	}
}

func TestFrontendClassGraphContentOrderAndMetadata(t *testing.T) {
	program, apps := renderTestGraph()
	originalObject := apps[0].Services[0].Objects[0]
	originalSentence := program.Books[0].Chapters[0].sentences[0]
	makeFile := func(name string) (*object, *sentence) {
		return &object{ID: name, Meta: map[string]string{"Type": "root"}}, &sentence{Name: name, Words: []*w{{Syllables: []*s{{CompiledBytes: []byte("A")}}}}}
	}
	first, firstSentence := makeFile("first")
	deep, deepSentence := makeFile("deep")
	last, lastSentence := makeFile("last")
	child := &class{ID: "child", Meta: map[string]string{"Parent": "group", "Package": "atlas", "Location": "/api/", "Label": "child"}, Objects: []*object{deep}}
	childText := &text{Name: "child", sentences: []*sentence{deepSentence}}
	group := &class{ID: "group", Meta: map[string]string{"Package": "atlas", "Location": "/api/", "Label": "parent, \"雪\"", "Lookup": "<JSON>literal</JSON>"}, Objects: []*object{first, last}, Children: []*class{child}, members: []sourceMember{{Object: first}, {Class: child}, {Object: last}}}
	groupText := &text{Name: "group", sentences: []*sentence{firstSentence, lastSentence}, children: []*text{childText}, members: []amlSourceMember{{Sentence: firstSentence}, {Text: childText}, {Sentence: lastSentence}}}
	apps[0].Services[0].Classes = []*class{group}
	apps[0].Services[0].members = []sourceMember{{Class: group}, {Object: originalObject}}
	program.Books[0].Chapters[0].Texts = []*text{groupText}
	program.Books[0].Chapters[0].members = []amlSourceMember{{Text: groupText}, {Sentence: originalSentence}}
	before, err := json.Marshal(apps)
	if err != nil {
		t.Fatal(err)
	}
	observable, backend, err := renderProgram(program, apps)
	if err != nil {
		t.Fatal(err)
	}
	for _, view := range [][]byte{observable, backend} {
		cursor := bytes.Index(view, []byte("<CONTENT>"))
		for _, item := range []string{`<TEXT "group">`, `<SENTENCE "first">`, `<TEXT "child">`, `<SENTENCE "deep">`, `<SENTENCE "last">`, `<SENTENCE "endpoint">`} {
			pos := bytes.Index(view[cursor:], []byte(item))
			if pos < 0 {
				t.Fatalf("content order lost %s\n%s", item, view)
			}
			cursor += pos + len(item)
		}
		if !bytes.Contains(view, []byte(`"Lookup":"\u003cJSON\u003eliteral\u003c/JSON\u003e"`)) || !bytes.Contains(view, []byte(`"Label":"parent, \"雪\""`)) {
			t.Fatalf("scoped metadata lost exact values: %s", view)
		}
	}
	after, err := json.Marshal(apps)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("rendering changed source graph")
	}
	again, _, err := renderProgram(program, apps)
	if err != nil || !bytes.Equal(again, observable) {
		t.Fatal("repeated recursive rendering is nondeterministic")
	}
}

func TestFrontendClassGraphRejectsPartialPublication(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*class, *text)
	}{
		{"cycle", func(group *class, _ *text) { group.Children = []*class{group} }},
		{"duplicate object", func(group *class, _ *text) { group.Objects = append(group.Objects, group.Objects[0]) }},
		{"missing content", func(_ *class, content *text) { content.sentences = nil }},
		{"inconsistent parent", func(group *class, _ *text) { group.Meta["Parent"] = "absent" }},
		{"reordered content", func(_ *class, content *text) { content.sentences[0].Name = "other" }},
		{"wrong member category", func(group *class, _ *text) { group.members = []sourceMember{{Class: group}} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			program, apps := renderTestGraph()
			svc, ch := apps[0].Services[0], program.Books[0].Chapters[0]
			group := &class{ID: "group", Meta: map[string]string{}, Objects: svc.Objects}
			content := &text{Name: "group", sentences: ch.sentences}
			svc.Classes, svc.Objects = []*class{group}, nil
			ch.Texts, ch.sentences = []*text{content}, nil
			tc.mutate(group, content)
			observable, backend, err := renderProgram(program, apps)
			if err == nil || observable != nil || backend != nil {
				t.Fatalf("invalid source graph published partial output: %v", err)
			}
		})
	}
	program, apps := renderTestGraph()
	program.Apps = apps
	program.AMLBytes, program.BackendBytes = []byte("prior observable"), []byte("prior backend")
	program.Books[0].Chapters[0].sentences = nil
	if err := program.Make(); err == nil {
		t.Fatal("invalid content succeeded")
	}
	if string(program.AMLBytes) != "prior observable" || string(program.BackendBytes) != "prior backend" {
		t.Fatal("failed rendering erased earlier published views")
	}
}

func TestFrontendRecursiveClassSource(t *testing.T) {
	source, err := os.ReadFile("test/classes/recursive.l")
	if err != nil {
		t.Fatal(err)
	}
	original := append([]byte(nil), source...)
	cfg := canonicalTestOptions()
	result, err := NewLygerService(cfg).Compile(source, "test/classes/recursive.l")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(source, original) {
		t.Fatal("frontend changed source bytes")
	}
	units, err := result.ApplicationUnits()
	if err != nil || len(units) != 1 {
		t.Fatalf("units: %d, %v", len(units), err)
	}
	for name, view := range map[string][]byte{"observable": result.Compilation.Program.AMLBytes, "backend": units[0].BackendAML} {
		t.Run(name, func(t *testing.T) {
			if bytes.Contains(view, []byte(`"Features"`)) || bytes.Contains(view, []byte("<PARAGRAPH")) {
				t.Fatal("removed Feature layer survives")
			}
			if !bytes.Contains(view, []byte(`"StructureVersion":"2"`)) {
				t.Fatal("recursive graph has no version")
			}
			cursor := 0
			for _, fragment := range []string{`<TEXT "outer">`, `<TEXT "middle">`, `<TEXT "leaf">`, `<SENTENCE "main">`, "</SENTENCE>", "</TEXT>", "</TEXT>", "</TEXT>"} {
				at := bytes.Index(view[cursor:], []byte(fragment))
				if at < 0 {
					t.Fatalf("ordered recursive content missing %s", fragment)
				}
				cursor += at + len(fragment)
			}
			var classes []struct {
				Name, Parent, Root         string
				Dimensions, GenericObjects []string
				Module                     *sourceModule
			}
			for _, line := range strings.Split(string(view), "\n") {
				if strings.HasPrefix(line, `<JSON>{"Classes":`) {
					var record struct{ Classes json.RawMessage }
					if err := json.Unmarshal([]byte(strings.TrimSuffix(strings.TrimPrefix(line, "<JSON>"), "</JSON>")), &record); err != nil {
						t.Fatal(err)
					}
					if err := json.Unmarshal(record.Classes, &classes); err != nil {
						t.Fatal(err)
					}
				}
			}
			if len(classes) != 3 || classes[0].Name != "outer" || classes[0].Parent != "" || classes[1].Parent != "outer" || classes[2].Parent != "middle" || !reflect.DeepEqual(classes[0].Dimensions, []string{"middle"}) || !reflect.DeepEqual(classes[1].Dimensions, []string{"leaf"}) || !reflect.DeepEqual(classes[2].GenericObjects, []string{"main", "__l_class_leaf"}) {
				t.Fatalf("recursive membership changed: %#v", classes)
			}
			for _, group := range classes {
				if group.Root != "__l_class_"+group.Name || group.Module == nil || group.Module.Name != group.Name || group.Module.Visibility != "crate" {
					t.Fatalf("Class has no explicit owned module root: %#v", group)
				}
			}
		})
	}
	bundle, err := Logiceme.NewLogicemeService(cfg.Resources).Evaluate(units[0].BackendAML, Logiceme.GenerationOptions{IncludeRuntime: false})
	recordBundle(t, bundle)
	if err != nil {
		t.Fatal(err)
	}
	wantPaths := [][]string{{"/src/main.rs", "/src/__l_class_leaf.rs", "/src/__l_class_middle.rs", "/src/__l_class_outer.rs", "/src/__l_service_main.rs", "/Cargo.toml", "/src/config.rs", "/start.sh", "/stop.sh"}}
	if len(bundle.ByteArrays) != 1 || len(bundle.ByteArrays[0]) != len(wantPaths[0]) || !reflect.DeepEqual(bundle.MapArrays, wantPaths) {
		t.Fatalf("expected one complete generated app with ordinary Rust support: %v", bundle.MapArrays)
	}
	file := observeBundleMain(t, bundle)
	file.function(t, "main")
	assertClassModules(t, bundle, map[string][]string{
		"/src/__l_service_main.rs": {"mod outer;", "outer::middle::leaf::main::main();"},
		"/src/__l_class_outer.rs":  {"mod middle;"},
		"/src/__l_class_middle.rs": {"mod leaf;"},
		"/src/__l_class_leaf.rs":   {"mod main;"},
	})
	buildActualBundle(t, bundle, units[0].ID, false, "")

}

func TestFrontendFeatureFreeAML(t *testing.T) {
	program, apps := renderTestGraph()
	observable, backend, err := renderProgram(program, apps)
	if err != nil {
		t.Fatal(err)
	}
	for _, view := range [][]byte{observable, backend} {
		if bytes.Contains(view, []byte(`"Features"`)) || bytes.Contains(view, []byte("<PARAGRAPH")) {
			t.Fatal("canonical AML contains removed Feature structure")
		}
		if bytes.Contains(view, []byte("StructureVersion")) {
			t.Fatal("flat output changed versions")
		}
	}
}

func TestFrontendInterleavedClassSource(t *testing.T) {
	source, err := os.ReadFile("test/classes/interleaved.l")
	if err != nil {
		t.Fatal(err)
	}
	cfg := canonicalTestOptions()
	result, err := NewLygerService(cfg).Compile(source, "test/classes/interleaved.l")
	if err != nil {
		t.Fatal(err)
	}
	units, err := result.ApplicationUnits()
	if err != nil || len(units) != 1 {
		t.Fatalf("source units: %d, %v", len(units), err)
	}
	originalAML := append([]byte(nil), units[0].BackendAML...)
	svc := result.Compilation.Program.Apps[0].Services[0]
	if len(svc.members) != 2 || svc.members[0].Class == nil || svc.members[1].Object == nil {
		t.Fatal("source interleaving lost at Service lowering")
	}
	group := svc.members[0].Class
	if len(group.members) != 4 || group.members[0].Object == nil || group.members[1].Class == nil || group.members[2].Object == nil || group.members[3].Object == nil || group.members[3].Object.ID != group.Root {
		t.Fatal("source interleaving lost at Class lowering")
	}
	bundle, err := Logiceme.NewLogicemeService(cfg.Resources).Evaluate(units[0].BackendAML, Logiceme.GenerationOptions{IncludeRuntime: false})
	recordBundle(t, bundle)
	if err != nil {
		t.Fatalf("%v\n%s", err, units[0].BackendAML)
	}
	want := []string{"/src/first.rs", "/src/deep.rs", "/src/__l_class_leaf.rs", "/src/last.rs", "/src/__l_class_branch.rs", "/src/main.rs", "/Cargo.toml", "/src/config.rs", "/start.sh", "/stop.sh"}
	if len(bundle.MapArrays) != 1 || !reflect.DeepEqual(bundle.MapArrays[0], want) {
		t.Fatalf("depth-first file order: %v", bundle.MapArrays)
	}
	if !bytes.Equal(originalAML, units[0].BackendAML) {
		t.Fatal("backend mutated frontend AML")
	}

	assertClassModules(t, bundle, map[string][]string{
		"/src/main.rs":             {"mod branch;"},
		"/src/__l_class_branch.rs": {"mod first;", "mod leaf;", "mod last;"},
		"/src/__l_class_leaf.rs":   {"mod deep;"},
	})
	buildActualBundle(t, bundle, units[0].ID, false, "")
}

func assertClassModules(t *testing.T, bundle Logiceme.Executable, expected map[string][]string) {
	t.Helper()
	for name, fragments := range expected {
		var source []byte
		for index, actual := range bundle.MapArrays[0] {
			if actual == name {
				source = bundle.ByteArrays[0][index]
			}
		}
		cursor := 0
		for _, fragment := range fragments {
			position := bytes.Index(source[cursor:], []byte(fragment))
			if position < 0 || bytes.Count(source, []byte(fragment)) != 1 {
				t.Fatalf("%s requires ordered, exactly-once module item %q:\n%s", name, fragment, source)
			}
			cursor += position + len(fragment)
		}
	}
}
