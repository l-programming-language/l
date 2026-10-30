package Lyger

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	distribution "github.com/l-programming-language/l"
	"github.com/l-programming-language/l/lerrors"
)

// RuntimeLibrary projects complete AML from the immutable distribution's derived
// dependency index. Custom providers and opaque authored leaves retain the full
// source-owned support library; no native payload is inspected for dependencies.
func (r *CompilationResult) RuntimeLibrary(name string, imports []string) ([]byte, error) {
	if _, err := r.buildProgram(); err != nil {
		return nil, err
	}
	if err := validateAMLName(name); err != nil {
		return nil, err
	}
	if !r.Compilation.canonicalRuntimeSelection() {
		return nil, nil
	}
	manifest, aml, err := distribution.LoadRuntimeComposition(distribution.Files())
	if err != nil {
		return nil, err
	}
	spans, exists := manifest.Selection[""]
	if !exists {
		return nil, nil
	}
	spans = append([][2]int(nil), spans...)
	for _, name := range imports {
		if !strings.HasPrefix(name, "runtime::") {
			continue
		}
		selected, exists := manifest.Selection[strings.TrimSuffix(strings.TrimPrefix(name, "runtime::"), " as _")]
		if !exists {
			return nil, nil
		}
		spans = append(spans, selected...)
	}
	head, tail, err := runtimeLibraryFrame(name)
	if err != nil {
		return nil, err
	}
	for _, span := range spans {
		if span[0] < 0 || span[0] >= span[1] || span[1] > len(aml) {
			return nil, lerrors.InvalidConfiguration.Errorf("composed runtime selection range is outside its AML")
		}
	}
	for _, span := range mergeRuntimeSpans(spans) {
		head = append(head, aml[span[0]:span[1]]...)
	}
	return append(head, tail...), nil
}

func (c *Compilation) canonicalRuntimeSelection() bool {
	if !distribution.IsEmbedded(c.Gram.resources) {
		return false
	}
	checked := make(map[SourceOrigin]bool)
	for _, source := range c.SourceStatements {
		if checked[source.Origin] || !bytes.Contains(source.Raw, []byte("<DEF>")) {
			continue
		}
		if source.Origin.Identity != "resource:"+source.Origin.Name {
			return false
		}
		raw, err := fs.ReadFile(c.Gram.resources, source.Origin.Name)
		if err != nil || fmt.Sprintf("%x", sha256.Sum256(raw)) != source.Origin.SHA256 {
			return false
		}
		checked[source.Origin] = true
	}
	return true
}

// ComposeRuntimeLibrary derives selection ranges from declaration dependencies
// during resource generation, never while compiling an end user's program.
func (c *Compilation) ComposeRuntimeLibrary() ([]byte, map[string][][2]int, error) {
	roots, err := c.composeExports("Runtime")
	if err != nil {
		return nil, nil, err
	}
	host, err := c.composeExports("Runtime.Host")
	if err != nil {
		return nil, nil, err
	}
	complete, err := definitionClosure(append(roots, host...)...)
	if err != nil {
		return nil, nil, err
	}
	aml, err := renderLibrary("runtime", "runtime", complete)
	if err != nil {
		return nil, nil, err
	}
	head, tail, err := runtimeLibraryFrame("runtime")
	if err != nil {
		return nil, nil, err
	}
	positions := make(map[*record][2]int, len(complete))
	position := len(head)
	for _, definition := range complete {
		encoded, err := renderLibrary("runtime", "runtime", []DefinitionInstance{definition})
		if err != nil {
			return nil, nil, err
		}
		body := encoded[len(head) : len(encoded)-len(tail)]
		end := position + len(body)
		if end > len(aml)-len(tail) || !bytes.Equal(aml[position:end], body) {
			return nil, nil, lerrors.Internal.Errorf("runtime declaration framing changed during composition")
		}
		positions[definition.definition] = [2]int{position, end}
		position = end
	}
	// Native type atoms are opaque: retain the same non-Function/type floor as
	// full source selection, together with roots required by the host adapters.
	for _, definition := range complete {
		if definition.definition.function == nil {
			host = append(host, definition)
		}
	}
	base, err := definitionClosure(host...)
	if err != nil {
		return nil, nil, err
	}
	selection := make(map[string][][2]int)
	retained := make(map[*record]bool)
	for _, definition := range base {
		retained[definition.definition] = true
		selection[""] = append(selection[""], positions[definition.definition])
	}
	for _, root := range roots {
		if root.definition.function == nil {
			continue
		}
		name := root.definition.function.Constructor.Name
		selected, err := definitionClosure(root)
		if err != nil {
			return nil, nil, err
		}
		selection[name] = nil
		for _, definition := range selected {
			if !retained[definition.definition] {
				selection[name] = append(selection[name], positions[definition.definition])
			}
		}
	}
	for name, spans := range selection {
		selection[name] = mergeRuntimeSpans(spans)
	}
	return aml, selection, nil
}

func runtimeLibraryFrame(name string) ([]byte, []byte, error) {
	frame, err := renderLibrary(name, "runtime", nil)
	if err != nil {
		return nil, nil, err
	}
	end := bytes.Index(frame, []byte("</SENTENCE>"))
	return bytes.Clone(frame[:end]), frame[end:], nil
}

func mergeRuntimeSpans(spans [][2]int) [][2]int {
	sort.Slice(spans, func(i, j int) bool { return spans[i][0] < spans[j][0] })
	var merged [][2]int
	for _, span := range spans {
		if n := len(merged); n != 0 && span[0] <= merged[n-1][1] {
			if span[1] > merged[n-1][1] {
				merged[n-1][1] = span[1]
			}
		} else {
			merged = append(merged, span)
		}
	}
	return merged
}
