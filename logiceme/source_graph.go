package Logiceme

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// TargetDescriptor is source topology, independent of App process identity.
type TargetDescriptor struct {
	Kind    string `json:"Kind"`
	Name    string `json:"Name"`
	Root    string `json:"Root"`
	Primary bool   `json:"Primary,omitempty"`
}
type ModuleDescriptor struct {
	Name       string `json:"Name"`
	Visibility string `json:"Visibility,omitempty"`
}
type preparedSourceTarget struct {
	TargetDescriptor
	Path    string
	Literal bool
}

type sourceHeader struct {
	Attributes, Modules, Imports []string
	root, primary, literal       bool
	library                      bool
	structural                   bool
	initializeConfig             bool
}

func sourceContext(meta MetaDict) MetaDict {
	if meta.Get("Meta").Type == "MetaDict" {
		return meta.GetDict("Meta")
	}
	return meta
}
func sourceMeta(parent, local MetaDict) MetaDict { return parent.Combine(local) }

// buildSourceGraphApp validates catalog edges against the parsed CONTENT tree,
// then builds the one ordered membership representation consumed downstream.
func buildSourceGraphApp(app App, book Book) (App, error) {
	services := map[string]OutlineEntry{}
	classes := map[string]OutlineEntry{}
	objects := map[string]Object{}
	for _, entry := range book.Outline.Services {
		if entry.Name == "" {
			return App{}, lerrors.InvalidInput.Errorf("Service identity is empty")
		}
		if _, ok := services[entry.Name]; ok {
			return App{}, lerrors.InvalidInput.Errorf("duplicate Service %q", entry.Name)
		}
		services[entry.Name] = entry
	}
	for _, entry := range book.Outline.Classes {
		if entry.Name == "" {
			return App{}, lerrors.InvalidInput.Errorf("Class identity is empty")
		}
		if _, ok := classes[entry.Name]; ok {
			return App{}, lerrors.InvalidInput.Errorf("duplicate Class %q", entry.Name)
		}
		classes[entry.Name] = entry
	}
	for _, entry := range book.Outline.Objects {
		if entry.Name == "" {
			return App{}, lerrors.InvalidInput.Errorf("Object identity is empty")
		}
		if _, ok := objects[entry.Name]; ok {
			return App{}, lerrors.InvalidInput.Errorf("duplicate Object %q", entry.Name)
		}
		objects[entry.Name] = entry
	}
	usedClasses := map[string]bool{}
	usedObjects := map[string]bool{}
	usedServices := map[string]bool{}
	makeObjects := func(entry OutlineEntry, sentences []Sentence) ([]Object, error) {
		locations := map[string]string{}
		for _, name := range entry.GenericObjects {
			if _, ok := locations[name]; ok {
				return nil, lerrors.InvalidInput.Errorf("%s has duplicate Object membership %q", entry.Name, name)
			}
			locations[name] = entry.Location
		}
		for _, name := range entry.DynamicObjects {
			if _, ok := locations[name]; ok {
				return nil, lerrors.InvalidInput.Errorf("%s has conflicting Object membership %q", entry.Name, name)
			}
			object, ok := objects[name]
			if !ok {
				return nil, lerrors.InvalidInput.Errorf("%s references missing Object %q", entry.Name, name)
			}
			locations[name] = object.Location
		}
		var result []Object
		for _, sentence := range sentences {
			location, ok := locations[sentence.Name]
			if !ok {
				return nil, lerrors.InvalidInput.Errorf("%s CONTENT Object %q has no outline membership", entry.Name, sentence.Name)
			}
			object, ok := objects[sentence.Name]
			if !ok {
				return nil, lerrors.InvalidInput.Errorf("%s references missing Object %q", entry.Name, sentence.Name)
			}
			if usedObjects[sentence.Name] {
				return nil, lerrors.InvalidInput.Errorf("Object %q has multiple owners or duplicate CONTENT", sentence.Name)
			}
			usedObjects[sentence.Name] = true
			object.Location = location
			object.Sentence = sentence
			object.Sentence.Meta = sourceMeta(object.Meta, sentence.Meta)
			object.structureVersion = "2"
			object.Compile()
			result = append(result, object)
			delete(locations, sentence.Name)
		}
		if len(locations) != 0 {
			for _, name := range append(append([]string{}, entry.GenericObjects...), entry.DynamicObjects...) {
				if _, ok := locations[name]; ok {
					return nil, lerrors.InvalidInput.Errorf("%s references Object %q without CONTENT", entry.Name, name)
				}
			}
		}
		return result, nil
	}
	checkChildren := func(entry OutlineEntry, texts []Text) error {
		wanted := map[string]bool{}
		for _, id := range entry.Dimensions {
			if wanted[id] {
				return lerrors.InvalidInput.Errorf("%s has duplicate Class membership %q", entry.Name, id)
			}
			if _, ok := classes[id]; !ok {
				return lerrors.InvalidInput.Errorf("%s references missing Class %q", entry.Name, id)
			}
			wanted[id] = true
		}
		for _, text := range texts {
			if !wanted[text.Name] {
				return lerrors.InvalidInput.Errorf("%s CONTENT Class %q conflicts with Dimensions", entry.Name, text.Name)
			}
			delete(wanted, text.Name)
		}
		for _, id := range entry.Dimensions {
			if wanted[id] {
				return lerrors.InvalidInput.Errorf("%s references Class %q without CONTENT", entry.Name, id)
			}
		}
		return nil
	}
	var buildClass func(Text, string) (Class, error)
	buildClass = func(text Text, parent string) (Class, error) {
		entry, ok := classes[text.Name]
		if !ok {
			return Class{}, lerrors.InvalidInput.Errorf("missing Class %q", text.Name)
		}
		if usedClasses[text.Name] {
			return Class{}, lerrors.InvalidInput.Errorf("Class %q has cyclic or multiple ownership", text.Name)
		}
		usedClasses[text.Name] = true
		if entry.Parent != parent {
			return Class{}, lerrors.InvalidInput.Errorf("Class %q Parent %q conflicts with enclosing Class %q", text.Name, entry.Parent, parent)
		}
		if err := checkChildren(entry, text.SubTexts); err != nil {
			return Class{}, err
		}
		result := Class{Name: text.Name, Type: entry.Type, Meta: sourceMeta(entry.Meta, sourceContext(text.Context)), Outline: book.Outline, members: append([]sourceMember(nil), text.Members...), structureVersion: "2"}
		var err error
		result.Objects, err = makeObjects(entry, text.Sentences)
		if err != nil {
			return Class{}, err
		}
		for _, child := range text.SubTexts {
			nested, err := buildClass(child, text.Name)
			if err != nil {
				return Class{}, err
			}
			result.SubClasses = append(result.SubClasses, nested)
		}
		return result, nil
	}
	for _, chapter := range book.Content {
		entry, ok := services[chapter.Name]
		if !ok {
			return App{}, lerrors.InvalidInput.Errorf("missing Service %q", chapter.Name)
		}
		if usedServices[chapter.Name] {
			return App{}, lerrors.InvalidInput.Errorf("duplicate Service CONTENT %q", chapter.Name)
		}
		usedServices[chapter.Name] = true
		if entry.Parent != "" {
			return App{}, lerrors.InvalidInput.Errorf("Service %q Parent is not supported in StructureVersion 2", entry.Name)
		}
		if err := checkChildren(entry, chapter.Texts); err != nil {
			return App{}, err
		}
		service := Service{Name: chapter.Name, Type: entry.Type, Meta: sourceMeta(entry.Meta, sourceContext(chapter.Context)), Settings: chapter.Context.GetDict("Settings"), Outline: book.Outline, members: append([]sourceMember(nil), chapter.Members...), structureVersion: "2"}
		var err error
		service.Objects, err = makeObjects(entry, chapter.Sentences)
		if err != nil {
			return App{}, err
		}
		for _, text := range chapter.Texts {
			class, err := buildClass(text, "")
			if err != nil {
				return App{}, err
			}
			service.Classes = append(service.Classes, class)
		}
		app.Services = append(app.Services, service)
	}
	for _, entry := range book.Outline.Services {
		if !usedServices[entry.Name] {
			return App{}, lerrors.InvalidInput.Errorf("Service %q has no CONTENT", entry.Name)
		}
	}
	for _, entry := range book.Outline.Classes {
		if !usedClasses[entry.Name] {
			return App{}, lerrors.InvalidInput.Errorf("Class %q has no unique CONTENT owner (missing or cyclic membership)", entry.Name)
		}
	}
	for _, entry := range book.Outline.Objects {
		if !usedObjects[entry.Name] {
			return App{}, lerrors.InvalidInput.Errorf("Object %q has no CONTENT owner", entry.Name)
		}
	}
	return app, nil
}

func (a *App) eachObject(visit func(*Object) error) error {
	var classes func([]Class) error
	classes = func(items []Class) error {
		for i := range items {
			for j := range items[i].Objects {
				if err := visit(&items[i].Objects[j]); err != nil {
					return err
				}
			}
			if err := classes(items[i].SubClasses); err != nil {
				return err
			}
		}
		return nil
	}
	for i := range a.Services {
		for j := range a.Services[i].Objects {
			if err := visit(&a.Services[i].Objects[j]); err != nil {
				return err
			}
		}
		if err := classes(a.Services[i].Classes); err != nil {
			return err
		}
	}
	return nil
}

func objectSourcePath(object *Object) string {
	if object.sourcePath != "" {
		return object.sourcePath
	}
	if object.Location == "" {
		return "/" + object.Name + ".rs"
	}
	return object.Location + object.Name + ".rs"
}

func sourceModuleLine(parent *Object, child *Object, module *ModuleDescriptor) (string, error) {
	if module == nil || module.Name == "" {
		return "", lerrors.InvalidInput.Errorf("Object %q requires an explicit Module descriptor", child.Name)
	}
	if err := validateSourceModuleName(module.Name); err != nil {
		return "", err
	}
	visibility := ""
	switch module.Visibility {
	case "", "private":
	case "public":
		visibility = "pub "
	case "crate":
		visibility = "pub(crate) "
	case "super":
		visibility = "pub(super) "
	default:
		return "", lerrors.InvalidInput.Errorf("module %q has unsupported Visibility %q", module.Name, module.Visibility)
	}
	parentPath, childPath := objectSourcePath(parent), objectSourcePath(child)
	conventional := path.Dir(parentPath)
	base := path.Base(parentPath)
	if !parent.header.root && base != "mod.rs" {
		conventional = path.Join(conventional, strings.TrimSuffix(base, ".rs"))
	}
	expected := path.Join(conventional, module.Name+".rs")
	expectedMod := path.Join(conventional, module.Name, "mod.rs")
	attribute := ""
	if childPath != expected && childPath != expectedMod {
		// path attributes for an external module are relative to the declaring file.
		relative, err := sourceRelativePath(path.Dir(parentPath), childPath)
		if err != nil {
			return "", err
		}
		quoted, err := quoteNativeString(relative)
		if err != nil {
			return "", err
		}
		attribute = "#[path = " + quoted + "]\n"
	}
	return attribute + visibility + "mod " + module.Name + ";\n", nil
}
func sourceRelativePath(base, target string) (string, error) {
	if err := validateBundlePath(target); err != nil {
		return "", err
	}
	left := strings.Split(strings.Trim(base, "/"), "/")
	right := strings.Split(strings.Trim(target, "/"), "/")
	if base == "/" {
		left = nil
	}
	i := 0
	for i < len(left) && i < len(right) && left[i] == right[i] {
		i++
	}
	parts := make([]string, len(left)-i)
	for j := range parts {
		parts[j] = ".."
	}
	parts = append(parts, right[i:]...)
	return strings.Join(parts, "/"), nil
}
func validateSourceModuleName(name string) error {
	for i, r := range name {
		if !(r == '_' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || i > 0 && r >= '0' && r <= '9') {
			return lerrors.InvalidInput.Errorf("invalid module name %q", name)
		}
	}
	if name == "" {
		return lerrors.InvalidInput.Errorf("empty module name")
	}
	switch name {
	case "_", "as", "break", "const", "continue", "crate", "else", "enum", "extern", "false", "fn", "for", "if", "impl", "in", "let", "loop", "match", "mod", "move", "mut", "pub", "ref", "return", "self", "Self", "static", "struct", "super", "trait", "true", "type", "unsafe", "use", "where", "while", "async", "await", "dyn", "abstract", "become", "box", "do", "final", "gen", "macro", "override", "priv", "typeof", "unsized", "virtual", "yield", "try":
		return lerrors.InvalidInput.Errorf("reserved module name %q", name)
	}
	return nil
}
func sourceGraphError(app string, err error) error {
	return fmt.Errorf("application %q source graph: %w", app, err)
}

func (entry *OutlineEntry) UnmarshalJSON(data []byte) error {
	var raw struct {
		Name           string
		Type           string
		EntryType      string
		Location       string
		Parent         string
		GenericObjects []string
		DynamicObjects []string
		Dimensions     []string
		Root           string
		Module         *ModuleDescriptor
		Target         *TargetDescriptor
		Meta           json.RawMessage
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*entry = OutlineEntry{Name: raw.Name, Type: raw.Type, EntryType: raw.EntryType, Location: raw.Location, Parent: raw.Parent, GenericObjects: raw.GenericObjects, DynamicObjects: raw.DynamicObjects, Dimensions: raw.Dimensions, Root: raw.Root, Module: raw.Module, Target: raw.Target}
	if len(raw.Meta) > 0 {
		meta, err := ParseMetaJSON("", string(raw.Meta))
		if err != nil {
			return err
		}
		entry.Meta = meta
	}
	return nil
}

func (object *Object) UnmarshalJSON(data []byte) error {
	var record struct {
		Name     string
		Type     string
		Package  string
		Location string
		Private  bool
		Module   *ModuleDescriptor
		Meta     json.RawMessage
		Inputs   []string
		Sentence Sentence
		Bytes    []byte
	}
	if err := json.Unmarshal(data, &record); err != nil {
		return err
	}
	*object = Object{Name: record.Name, Type: record.Type, Package: record.Package, Location: record.Location, Private: record.Private, Module: record.Module, Inputs: record.Inputs, Sentence: record.Sentence, Bytes: record.Bytes}
	if len(record.Meta) > 0 {
		// Retain the existing serialized MetaDict representation for direct Go
		// callers; version-2 AML authors may use the same ordinary JSON metadata
		// object as Classes. Neither route treats imported source as shared context.
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(record.Meta, &fields); err != nil {
			return err
		}
		if _, ok := fields["metas"]; ok {
			return json.Unmarshal(record.Meta, &object.Meta)
		}
		meta, err := ParseMetaJSON("", string(record.Meta))
		if err != nil {
			return err
		}
		object.Meta = meta
	}
	return nil
}
