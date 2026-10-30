package Lyger

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"strings"

	distribution "github.com/l-programming-language/l"
)

// RuntimeLibrary returns complete library AML for the selected runtime roots.
// A nil result requests the ordinary full support library. Only unchanged
// canonical native leaves have audited imports; all other providers and opaque
// authored leaves retain full support without guessing their dependencies.
func (r *CompilationResult) RuntimeLibrary(name string, imports []string) ([]byte, error) {
	if _, err := r.buildProgram(); err != nil {
		return nil, err
	}
	if err := validateAMLName(name); err != nil {
		return nil, err
	}
	c := r.Compilation
	if !c.canonicalRuntimeSelection() {
		return nil, nil
	}
	options := Options{Resources: c.Gram.resources, AssetRoot: c.Gram.assetRoot, Language: c.Gram.language}
	library := NewLygerService(options).NewCompilation([]byte("Use Runtime from `runtime`.\n"), "runtime-selection.l")
	library.Make()
	if library.Error != nil {
		return nil, library.Error
	}
	complete, err := library.composeExports("Runtime")
	if err != nil {
		return nil, err
	}
	required, err := library.composeExports("Runtime.Host")
	if err != nil {
		return nil, err
	}
	functions := make(map[string]DefinitionInstance)
	for _, definition := range complete {
		if function := definition.definition.function; function != nil {
			functions[function.Constructor.Name] = definition
		}
	}
	for _, name := range imports {
		if !strings.HasPrefix(name, "runtime::") {
			continue
		}
		root, exists := functions[strings.TrimSuffix(strings.TrimPrefix(name, "runtime::"), " as _")]
		if !exists {
			return nil, nil
		}
		required = append(required, root)
	}
	complete, err = definitionClosure(complete...)
	if err != nil {
		return nil, err
	}
	var selected []DefinitionInstance
	// Native type atoms are intentionally opaque. Retain every declaration that
	// is not a Function, including private types reached by the full library.
	for _, definition := range complete {
		if definition.definition.function == nil {
			selected = append(selected, definition)
		}
	}
	selected, err = definitionClosure(append(selected, required...)...)
	if err != nil {
		return nil, err
	}
	return renderLibrary(name, "runtime", selected)
}

func (c *Compilation) canonicalRuntimeSelection() bool {
	canonical := distribution.Files()
	for _, source := range c.SourceStatements {
		if !bytes.Contains(source.Raw, []byte("<DEF>")) {
			continue
		}
		if source.Origin.Identity != "resource:"+source.Origin.Name {
			return false
		}
		raw, err := fs.ReadFile(canonical, source.Origin.Name)
		if err != nil || fmt.Sprintf("%x", sha256.Sum256(raw)) != source.Origin.SHA256 {
			return false
		}
	}
	// This is a conservative optimization of this exact distribution, not a
	// dependency contract imposed on authored native resources. It also covers
	// changed native adapters, process definitions, templates and lexicons.
	return fs.WalkDir(canonical, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		expected, err := fs.ReadFile(canonical, name)
		if err != nil {
			return err
		}
		actual, err := fs.ReadFile(c.Gram.resources, name)
		if err != nil {
			return err
		}
		if !bytes.Equal(actual, expected) {
			return fmt.Errorf("noncanonical resource %s", name)
		}
		return nil
	}) == nil
}
