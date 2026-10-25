package Logiceme

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"

	"github.com/l-programming-language/l/lerrors"
)

// Object ...
type Object struct {
	Module           *ModuleDescriptor `json:"Module,omitempty"`
	header           sourceHeader
	structureVersion string
	sourcePath       string
	Name             string   `json:"Name,omitempty"`
	Type             string   `json:"Type,omitempty"`
	Package          string   `json:"Package,omitempty"`
	Location         string   `json:"Location,omitempty"`
	Private          bool     `json:"Private,omitempty"`
	Meta             MetaDict `json:"Meta,omitempty"`
	Inputs           []string `json:"Inputs,omitempty"`
	Sentence         Sentence `json:"Sentence,omitempty"`
	Bytes            []byte   `json:"Bytes,omitempty"`
}

// CompiledObject ...
type CompiledObject struct {
	headerSize      int
	Name            string
	Type            string
	Package         string
	Location        string
	SegmentalSystem SegmentalSystem
	Private         bool
	Meta            MetaDict
	Inputs          []string
	Dependencies    []string
	EnvDependencies []string
	Tensor          Tensor
	Content         []byte
}

// Assemble an Object
func (obj *Object) Assemble(meta MetaDict, segSys SegmentalSystem) (CompiledObject, error) {
	if obj.structureVersion == "2" {
		meta = sourceMeta(meta, obj.Sentence.Meta)
		obj.Sentence.Meta = MetaDict{}
	}
	if meta.GetVal("Package") == "NotFound" {
		meta.SetAdd("Package", obj.Name, "String")
	}
	if segSys.Language == "Rust" {
		segSys.fileMeta = sourceMeta(meta, sourceHeaderMeta(obj.header))
		segSys.itemVisibility = ""
		segSys.initializeConfig = obj.header.initializeConfig
		if obj.structureVersion == "2" && !obj.Private {
			segSys.itemVisibility = "pub(crate) "
			if obj.header.library {
				segSys.itemVisibility = "pub "
			}
		}
	}
	obj.Sentence.Tensor.SegmentalSystem = segSys
	extension := ""
	var literal bool
	if segSys.Language == "Rust" {
		contents, suffix, isLiteral, err := prepareSourceFile(obj.Sentence, &segSys)
		if err != nil {
			return CompiledObject{}, sourceObjectError(obj.Location+obj.Name, err)
		}
		extension, literal = "."+suffix, isLiteral
		if literal {
			obj.Sentence.Content = contents
		}
	}
	if !literal {
		if err := obj.Sentence.Assemble(meta); err != nil {
			return CompiledObject{}, err
		}
		obj.Sentence.Link()
	}
	var cName string
	if obj.Location != "" {
		cName = obj.Location + obj.Name
	} else {
		cName = "/" + obj.Name
	}
	cName += extension
	if segSys.Language == "Rust" {
		if err := validateBundlePath(cName); err != nil {
			return CompiledObject{}, sourceObjectError(cName, err)
		}
	}
	headerSize := 0
	if segSys.Language == "Rust" && !literal {
		for _, word := range obj.Sentence.Tensor.Dimensions {
			if len(word.Dimensions) > 0 {
				headerSize = len(word.Dimensions[0].Content)
				break
			}
		}
		if obj.header.structural {
			// These are typed source-graph items, separate from the unchanged
			// A glyph's package comment and import payloads. Retain the boundary
			// as provenance for support assembly; never recover it from source.
			var prefix bytes.Buffer
			for _, attribute := range obj.header.Attributes {
				prefix.WriteString(attribute)
			}
			for _, module := range obj.header.Modules {
				prefix.WriteString(module)
			}
			headerSize += prefix.Len()
			prefix.Write(obj.Sentence.Content)
			obj.Sentence.Content = prefix.Bytes()
		}
	}
	cObject := CompiledObject{
		headerSize:      headerSize,
		Name:            cName,
		Type:            obj.Type,
		Package:         obj.Package,
		Location:        obj.Location,
		SegmentalSystem: segSys,
		Private:         obj.Private,
		Meta:            meta,
		Inputs:          obj.Inputs,
		Dependencies:    obj.Sentence.Dependencies,
		EnvDependencies: obj.Sentence.EnvDependencies,
		Tensor:          obj.Sentence.Tensor,
		Content:         obj.Sentence.Content,
	}
	return cObject, nil
}

// Compile Object
func (obj *Object) Compile() {
	obj.Sentence.Compile()
}

// BuildObjects
func BuildObjects(sentences []Sentence) []Object {
	reObjs := []Object{}
	for _, sentence := range sentences {
		obj := Object{Name: sentence.Name, Sentence: sentence}
		obj.Compile()
		reObjs = append(reObjs, obj)
	}
	return reObjs
}

// prepareSourceFile recognizes the literal-file primitive at its owning Object,
// before source rendering. Native values are opaque while glyph identities and
// child relationships are parsed by the ordinary segmental system.
func prepareSourceFile(sentence Sentence, system *SegmentalSystem) (contents []byte, extension string, literal bool, err error) {
	if err := validateGenericInputInheritance(system); err != nil {
		return nil, "", false, err
	}
	extension = "rs"
	var forms []string
	var fileGlyph *Glyph
	var fileRestore *strings.Replacer
	for _, word := range sentence.Words {
		for _, syllable := range word.Syllables {
			masked, restore, maskErr := maskNativeValues([]byte(syllable.Name))
			if maskErr != nil {
				return nil, "", false, maskErr
			}
			glyph, parseErr := system.initializeGlyphChecked(masked, nil, nil)
			if parseErr != nil {
				return nil, "", false, &sourceFileDiagnostic{cause: parseErr, message: restore.Replace(parseErr.Error())}
			}
			form := glyphForm(&glyph)
			forms = append(forms, form)
			if nestedErr := validateSourceFilePlacement(&glyph, ""); nestedErr != nil {
				return nil, "", false, nestedErr
			}
			if form != "EA" {
				if syllable.hasExplicitModifiers || syllable.Count != 1 || syllable.Exponent != 1 || syllable.Polarity {
					return nil, "", false, lerrors.Unsupported.Errorf("Rust syllable modifiers have no supported repetition, exponent or polarity semantics")
				}
				continue
			}
			if syllable.hasExplicitModifiers || syllable.Count != 1 || syllable.Exponent != 1 || syllable.Polarity {
				return nil, "", false, lerrors.InvalidInput.Errorf("EA excludes syllable modifiers")
			}
			fileGlyph, fileRestore = &glyph, restore
		}
	}
	if fileGlyph == nil {
		if len(forms) == 0 || forms[0] != "A" {
			return nil, "", false, lerrors.InvalidInput.Errorf("composed Object requires one leading A source header")
		}
		for _, form := range forms[1:] {
			if form == "A" {
				return nil, "", false, lerrors.InvalidInput.Errorf("composed Object requires exactly one leading A source header")
			}
		}
		return nil, extension, false, nil
	}
	if len(forms) != 1 {
		return nil, "", false, lerrors.InvalidInput.Errorf("EA requires one standalone glyph; found %s", strings.Join(forms, ", "))
	}
	values, parseErr := readLiteralFileInputs(fileGlyph.InputString, fileRestore)
	if parseErr != nil {
		return nil, "", false, parseErr
	}
	body, present := values["Contents"]
	if !present {
		return nil, "", false, lerrors.InvalidInput.Errorf("EA requires Contents (an empty value is allowed)")
	}
	if supplied := values["Type"]; supplied != "" {
		if strings.ContainsAny(supplied, "/\\") || strings.IndexFunc(supplied, unicode.IsControl) >= 0 || supplied == "." || supplied == ".." {
			return nil, "", false, lerrors.InvalidInput.Errorf("EA Type must be a file extension: %q", supplied)
		}
		extension = supplied
	}
	if err := validateGlyphMetadata(fileGlyph, glyphContext{}); err != nil {
		return nil, "", false, err
	}
	return bytes.Clone([]byte(body)), extension, true, nil
}

func validateSourceFilePlacement(glyph *Glyph, parent string) error {
	form := glyphForm(glyph)
	if form == "A" && (parent != "" || len(glyph.Conjoinments) != 0) {
		return lerrors.InvalidInput.Errorf("A is a standalone source header; it cannot be nested or consume children")
	}
	if form == "EA" {
		if parent != "" {
			return lerrors.InvalidInput.Errorf("EA requires a standalone file; nested under %s", parent)
		}
		if len(glyph.Conjoinments) != 0 {
			return lerrors.InvalidInput.Errorf("EA excludes child glyphs")
		}
	}
	for i := range glyph.Conjoinments {
		if err := validateSourceFilePlacement(&glyph.Conjoinments[i], form); err != nil {
			return err
		}
	}
	return nil
}

// The glyph parser sees only opaque value tokens. Reading these framed fields
// cannot mistake a colon or a Type-like string inside Contents for metadata.
// Restore each value once; never run literal contents through the LU linker.
func readLiteralFileInputs(input string, restore *strings.Replacer) (map[string]string, error) {
	values := make(map[string]string)
	if input == "" {
		return values, nil
	}
	for _, field := range strings.Split(input, "~, ") {
		key, value, ok := strings.Cut(field, ": ~")
		if !ok || key == "" {
			return nil, lerrors.Syntax.Errorf("EA metadata requires Key: ~value~ framing")
		}
		value = strings.TrimSuffix(value, "~")
		if _, exists := values[key]; exists && (key == "Contents" || key == "Type") {
			return nil, lerrors.InvalidInput.Errorf("EA duplicate %s", key)
		}
		values[key] = restore.Replace(value)
	}
	return values, nil
}

func sourceObjectError(name string, err error) error {
	return fmt.Errorf("Rust source object %q: %w", name, err)
}

// Keep opaque framing tokens out of diagnostics without discarding the parser's
// classified error or its underlying cause.
type sourceFileDiagnostic struct {
	cause   error
	message string
}

func (e *sourceFileDiagnostic) Error() string { return e.message }
func (e *sourceFileDiagnostic) Unwrap() error { return e.cause }
