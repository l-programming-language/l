package Logiceme

import (
	"encoding/json"
	"path"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

const runtimeHeaderBoundary = "// The source assembler supplies a same-module generated item with this contract:"

var runtimeModuleHeaders = []string{
	"pub(crate) mod runtime;\n",
	"#[cfg(any(target_os = \"linux\", target_os = \"macos\"))]\n#[path = \"process_unix.rs\"]\npub(crate) mod process;\n",
	"#[cfg(not(any(target_os = \"linux\", target_os = \"macos\")))]\n#[path = \"process_other.rs\"]\npub(crate) mod process;\n",
}
var runtimeImportHeaders = []string{"use std::io::Write;\n", "use std::panic::{catch_unwind, resume_unwind, AssertUnwindSafe};\n", "use std::process::ExitCode;\n"}

func sourceHeaderMeta(header sourceHeader) MetaDict {
	var meta MetaDict
	// The copied A template owns `use` punctuation. These lines were framed
	// by this assembler; expose their payloads through its existing Imports
	// interface, retaining the file's collected order and deduplication.
	imports := make([]string, 0, len(header.Imports))
	for _, line := range header.Imports {
		imports = append(imports, strings.TrimSuffix(strings.TrimPrefix(line, "use "), ";\n"))
	}
	encoded, _ := json.Marshal(imports) //nolint:errcheck // The import list contains only strings, which JSON can always encode.
	meta.Add("Imports", string(encoded), "PreparedStrings")
	for _, field := range []struct {
		name   string
		values []string
	}{{"HeaderAttributes", header.Attributes}, {"HeaderModules", header.Modules}, {"HeaderImports", header.Imports}} {
		values := field.values
		if values == nil {
			values = []string{}
		}
		encoded, _ := json.Marshal(values) //nolint:errcheck // The collected metadata values are strings, which JSON can always encode.
		meta.Add(field.name, string(encoded), "PreparedStrings")
	}
	return meta
}

// prepareSourceHeaders establishes explicit module relationships before
// any A is interpreted. The graph describes source ownership, never runtime lifetimes.
func prepareSourceHeaders(app *App, includeRuntime bool, workspace *WorkspaceRuntime) error {
	objects := map[string]*Object{}
	if err := app.eachObject(func(object *Object) error {
		if _, ok := objects[object.Name]; ok {
			return lerrors.InvalidInput.Errorf("duplicate Object identity %q at source path %s", object.Name, objectSourcePath(object))
		}
		objects[object.Name] = object
		_, suffix, literal, err := prepareSourceFile(object.Sentence, &app.SegmentalSystem)
		if err != nil {
			return sourceObjectError(object.Name, err)
		}
		object.sourcePath = object.Location + object.Name + "." + suffix
		if object.Location == "" {
			object.sourcePath = "/" + object.sourcePath
		}
		if err := validateBundlePath(object.sourcePath); err != nil {
			return err
		}
		object.header.literal = literal
		return nil
	}); err != nil {
		return err
	}
	v2 := app.Outline.Meta.GetVal("StructureVersion") == "2"
	for _, object := range objects {
		// Explicit topology and runtime support own these source items. The
		// legacy nonruntime bundle keeps its frozen A-only header contract.
		object.header.structural = v2 || includeRuntime
	}
	var primary *Object
	if v2 {
		ordered := app.Outline
		ordered.Services = nil
		for _, service := range app.Services {
			ordered.Services = append(ordered.Services, app.Outline.GetEntry(service.Name, "Service"))
		}
		targets, err := prepareSourceTargets(ordered, objects)
		if err != nil {
			return err
		}
		for _, target := range targets {
			root := objects[target.Root]
			root.header.root = true
			root.header.library = target.Kind == "lib"
			app.sourceTargets = append(app.sourceTargets, preparedSourceTarget{TargetDescriptor: target, Path: objectSourcePath(root), Literal: root.header.literal})
			if target.Primary {
				root.header.primary = true
				primary = root
			}
		}
		if libraryOnlyTargets(app.sourceTargets) {
			if includeRuntime {
				return lerrors.InvalidInput.Errorf("library-only package cannot include runtime")
			}
			// This is the package's source root, not an executable Primary.
			primary = objects[targets[0].Root]
		}
	} else {
		directory := app.Outline.Meta.GetVal("AppGenDir")
		if directory == "NotFound" || directory == "" {
			directory = "/src"
		}
		if directory != "/" {
			if err := validateBundlePath(directory); err != nil {
				return err
			}
		}
		for _, object := range objects {
			if objectSourcePath(object) == path.Join(directory, "main.rs") {
				primary = object
				object.header.root = true
				object.header.primary = true
			}
		}
	}
	if primary == nil {
		return lerrors.InvalidInput.Errorf("Rust application %q has no primary entry Object", app.Name)
	}
	if workspace != nil && primary.header.literal {
		return lerrors.InvalidInput.Errorf("Rust workspace entry requires a composed primary Object, not an opaque EA file")
	}
	for _, object := range objects {
		if object.header.root && !object.header.literal {
			object.header.Attributes = []string{"#![forbid(unsafe_code)]\n"}
		}
	}
	if !primary.header.literal && !primary.header.library {
		primary.header.Modules = append(primary.header.Modules, "mod config;\n")
		if includeRuntime {
			if workspace == nil {
				primary.header.Modules = append(primary.header.Modules, runtimeModuleHeaders...)
			} else {
				// Root the dependency in the extern prelude so an identically
				// named user module cannot shadow the coordinator-owned crate.
				primary.header.Imports = append(primary.header.Imports, "use ::"+workspace.Package+"::runtime;\n")
			}
			primary.header.Imports = append(primary.header.Imports, runtimeImportHeaders...)
		}
	}
	if v2 {
		classEntries := map[string]OutlineEntry{}
		for _, entry := range app.Outline.Classes {
			classEntries[entry.Name] = entry
		}
		var connect func(*Object, []Object, []Class, []sourceMember) error
		connect = func(root *Object, direct []Object, classes []Class, members []sourceMember) error {
			seen := map[string]bool{}
			if root.header.primary {
				seen["config"] = true
				if includeRuntime {
					seen["runtime"] = true
					if workspace == nil {
						seen["process"] = true
					}
				}
			}
			for _, member := range members {
				var child *Object
				var descriptor *ModuleDescriptor
				if member.Kind == "Object" {
					child = objects[direct[member.Index].Name]
					child.header.library = root.header.library
					if child.Name == root.Name {
						continue
					}
					descriptor = child.Module
				} else {
					group := classes[member.Index]
					entry := classEntries[group.Name]
					child = objects[entry.Root]
					if child == nil {
						return lerrors.InvalidInput.Errorf("Class %q requires an owned Root Object", group.Name)
					}
					owned := false
					for _, obj := range group.Objects {
						if obj.Name == entry.Root {
							owned = true
						}
					}
					if !owned {
						return lerrors.InvalidInput.Errorf("Class %q Root %q is not its direct Object", group.Name, entry.Root)
					}
					descriptor = entry.Module
					child.header.library = root.header.library
					if err := connect(child, group.Objects, group.SubClasses, group.members); err != nil {
						return err
					}
				}
				if descriptor == nil {
					return lerrors.InvalidInput.Errorf("Object/Class member %q requires Module", child.Name)
				}
				if seen[descriptor.Name] {
					return lerrors.InvalidInput.Errorf("module %q collides in Object %q", descriptor.Name, root.Name)
				}
				seen[descriptor.Name] = true
				line, err := sourceModuleLine(root, child, descriptor)
				if err != nil {
					return err
				}
				root.header.Modules = append(root.header.Modules, line)
			}
			return nil
		}
		for i := range app.Services {
			service := &app.Services[i]
			entry := app.Outline.GetEntry(service.Name, "Service")
			root := objects[entry.Target.Root]
			if err := connect(root, service.Objects, service.Classes, service.members); err != nil {
				return err
			}
		}
	}
	return app.eachObject(func(object *Object) error {
		if object.header.literal {
			return nil
		}
		imports, err := objectImports(object.Sentence, &app.SegmentalSystem)
		if err != nil {
			return err
		}
		seen := map[string]bool{}
		for _, line := range object.header.Imports {
			seen[line] = true
		}
		for _, line := range imports {
			if !seen[line] {
				object.header.Imports = append(object.header.Imports, line)
				seen[line] = true
			}
		}
		return nil
	})
}

func prepareSourceTargets(outline Outline, objects map[string]*Object) ([]TargetDescriptor, error) {
	var targets []TargetDescriptor
	names := map[string]bool{}
	roots := map[string]bool{}
	libraries, binaries, primary := 0, 0, 0
	for _, service := range outline.Services {
		if service.Target == nil {
			return nil, lerrors.InvalidInput.Errorf("Service %q requires explicit Target in StructureVersion 2", service.Name)
		}
		target := *service.Target
		if err := validateBundleAppName(target.Name); err != nil {
			return nil, err
		}
		if names[target.Kind+":"+target.Name] {
			return nil, lerrors.InvalidInput.Errorf("duplicate crate target name %q", target.Name)
		}
		names[target.Kind+":"+target.Name] = true
		if roots[target.Root] {
			return nil, lerrors.InvalidInput.Errorf("crate Root %q has multiple targets", target.Root)
		}
		roots[target.Root] = true
		root := objects[target.Root]
		if root == nil {
			return nil, lerrors.InvalidInput.Errorf("Service %q target Root %q is missing", service.Name, target.Root)
		}
		owned := false
		for _, id := range append(append([]string(nil), service.GenericObjects...), service.DynamicObjects...) {
			if id == target.Root {
				owned = true
			}
		}
		if !owned {
			return nil, lerrors.InvalidInput.Errorf("Service %q target Root %q must be a direct Object", service.Name, target.Root)
		}
		switch target.Kind {
		case "lib":
			if err := validateSourceModuleName(target.Name); err != nil {
				return nil, err
			}
			libraries++
			if target.Primary {
				return nil, lerrors.InvalidInput.Errorf("library target cannot be Primary")
			}
		case "bin":
			binaries++
			if target.Primary {
				primary++
			}
		default:
			return nil, lerrors.InvalidInput.Errorf("Service %q target Kind must be bin or lib", service.Name)
		}
		targets = append(targets, target)
	}
	if libraries > 1 {
		return nil, lerrors.InvalidInput.Errorf("one App package supports at most one library target")
	}
	if binaries == 0 && libraries != 1 || binaries > 0 && primary != 1 {
		return nil, lerrors.InvalidInput.Errorf("one App package requires binaries and exactly one Primary binary")
	}
	return targets, nil
}

func libraryOnlyTargets(targets []preparedSourceTarget) bool {
	return len(targets) == 1 && targets[0].Kind == "lib"
}

func objectImports(sentence Sentence, system *SegmentalSystem) ([]string, error) {
	var imports []string
	var walk func(*Glyph, *strings.Replacer) error
	walk = func(glyph *Glyph, restore *strings.Replacer) error {
		fields, err := readLiteralFileInputs(glyph.InputString, restore)
		if err != nil {
			return err
		}
		if raw, ok := fields["Imports"]; ok {
			parts, err := importAtoms(raw)
			if err != nil {
				return err
			}
			for _, part := range parts {
				imports = append(imports, "use "+part+";\n")
			}
		}
		for i := range glyph.Conjoinments {
			if err := walk(&glyph.Conjoinments[i], restore); err != nil {
				return err
			}
		}
		return nil
	}
	for _, word := range sentence.Words {
		for _, syllable := range word.Syllables {
			masked, restore, err := maskNativeValues([]byte(syllable.Name))
			if err != nil {
				return nil, err
			}
			glyph, err := system.initializeGlyphChecked(masked, nil, nil)
			if err != nil {
				return nil, err
			}
			if err := walk(&glyph, restore); err != nil {
				return nil, err
			}
		}
	}
	return imports, nil
}
func importAtoms(raw string) ([]string, error) {
	if strings.HasPrefix(raw, "&[") && strings.HasSuffix(raw, "]&") {
		raw = raw[2 : len(raw)-2]
	} else if strings.HasPrefix(raw, "[") && strings.HasSuffix(raw, "]") {
		raw = raw[1 : len(raw)-1]
	} else {
		return nil, lerrors.InvalidInput.Errorf("Imports requires an array")
	}
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	var parts []string
	start, depth := 0, 0
	quote := byte(0)
	for i := 0; i < len(raw); i++ {
		c := raw[i]
		if quote != 0 {
			if c == '\\' {
				i++
			} else if c == quote {
				quote = 0
			}
			continue
		}
		switch c {
		case '"', '\'':
			quote = c
		case '{', '[', '(':
			depth++
		case '}', ']', ')':
			depth--
		case ',':
			if depth == 0 {
				parts = append(parts, strings.TrimSpace(raw[start:i]))
				start = i + 1
			}
		}
	}
	parts = append(parts, strings.TrimSpace(raw[start:]))
	for i, part := range parts {
		if strings.HasPrefix(part, "\"") {
			var value string
			if err := json.Unmarshal([]byte(part), &value); err != nil {
				return nil, lerrors.InvalidInput.Errorf("import string: %w", err)
			}
			part = value
		}
		if part == "" || strings.ContainsAny(part, "\r\n;") {
			return nil, lerrors.InvalidInput.Errorf("invalid import path %q", part)
		}
		parts[i] = part
	}
	return parts, nil
}

// The A occurrence was compiled in its own Scalar tensor. Its byte length is
// structural provenance, so support insertion never searches completed source
// for attributes, functions, comments or marker-like native payloads.
func (app *CompiledApp) compiledHeaderSize(name string) (int, bool) {
	var class func(CompiledClass) (int, bool)
	class = func(item CompiledClass) (int, bool) {
		for _, object := range item.Objects {
			if object.Name == name {
				return object.headerSize, true
			}
		}
		for _, child := range item.SubClasses {
			if length, ok := class(child); ok {
				return length, true
			}
		}
		return 0, false
	}
	for _, service := range app.Services {
		for _, object := range service.Objects {
			if object.Name == name {
				return object.headerSize, true
			}
		}
		for _, item := range service.Classes {
			if length, ok := class(item); ok {
				return length, true
			}
		}
	}
	return 0, false
}
