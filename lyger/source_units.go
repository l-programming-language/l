package Lyger

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// SourceKind distinguishes lexical compilation units from source occurrences.
// File inclusion adds occurrences to the importing unit, never a new namespace.
type SourceKind string

const (
	SourceBuiltin    SourceKind = "builtin"
	SourceProgram    SourceKind = "program"
	SourceDictionary SourceKind = "dictionary"
)

type SourceOrigin struct {
	Identity     string
	Name         string
	SHA256       string
	PhysicalPath string
}
type SourceUnit struct {
	Kind         SourceKind
	Origin       SourceOrigin
	Dictionary   string
	root         *lexicalScope
	resourceRoot string
}
type dictionaryImport struct {
	unit           *SourceUnit
	module, member string
}
type dictionaryEntry struct {
	Entry      string `json:"entry"`
	Dictionary string `json:"dictionary"`
}
type dictionaryIndex struct {
	Format    int                                   `json:"format"`
	Languages map[string]map[string]dictionaryEntry `json:"languages"`
}

func (b *builder) resourceSource(name string) (*PreparedSource, SourceOrigin, error) {
	raw, err := fs.ReadFile(b.Gram.resources, name)
	if err != nil {
		return nil, SourceOrigin{}, lerrors.IO.Errorf("read resource %s: %w", name, err)
	}
	origin := SourceOrigin{Identity: "resource:" + name, Name: name, SHA256: fmt.Sprintf("%x", sha256.Sum256(raw))}
	if b.Gram.assetRoot != "" {
		origin.PhysicalPath, err = filepath.Abs(filepath.Join(b.Gram.assetRoot, filepath.FromSlash(name)))
		if err != nil {
			return nil, SourceOrigin{}, lerrors.IO.Wrap(err, "")
		}
	}
	source, err := PrepareSource(name, raw)
	return source, origin, err
}
func (b *builder) loadDictionary(name string, active map[string]bool, chain []string) (*SourceUnit, error) {
	if !fs.ValidPath(name) || name == "." || strings.Contains(name, "\\") {
		return nil, lerrors.InvalidInput.Errorf("invalid dictionary path %q", name)
	}
	if b.index == nil {
		index, err := readDictionaryIndex(b.Gram.resources)
		if err != nil {
			return nil, err
		}
		b.index = index
	}

	entry, ok := b.index.Languages[b.Gram.language][name]
	if !ok {
		if strings.Contains(strings.Split(name, "/")[0], ".") {
			return nil, lerrors.Unsupported.Errorf("dictionary provider is not configured for %s", name)
		}
		return nil, lerrors.UnresolvedReference.Errorf("unknown dictionary %s", name)
	}
	if active["resource:"+entry.Entry] {
		return nil, lerrors.Cycle.Errorf("dictionary import cycle: %s", strings.Join(append(chain, name), " -> "))
	}
	if unit := b.dictionaries[entry.Entry]; unit != nil {
		return unit, nil
	}
	source, origin, err := b.resourceSource(entry.Entry)
	if err != nil {
		return nil, err
	}
	unit := &SourceUnit{Kind: SourceDictionary, Origin: origin, Dictionary: entry.Dictionary, resourceRoot: path.Dir(entry.Entry)}
	b.dictionaries[entry.Entry] = unit
	if err := b.expand(source, unit, 0, SourceSpan{}, active, append(chain, name)); err != nil {
		return nil, err
	}
	b.units = append(b.units, unit)
	return unit, nil
}

func (c *Compilation) validateDictionaryExports() error {
	for _, unit := range c.SourceUnits {
		if unit.Kind != SourceDictionary {
			continue
		}
		binding, _, err := unit.root.lookupLocal([]byte(unit.Dictionary))
		if err != nil {
			return err
		}
		if binding == nil {
			return lerrors.UnresolvedReference.Errorf("dictionary %s does not export registered Dictionary %s", unit.Origin.Name, unit.Dictionary)
		}
		rec, err := canonicalRecord(binding.Record, c.Semantics.Aliases)
		if err != nil {
			return err
		}
		kind, err := instructionType(rec)
		if err != nil {
			return err
		}
		if kind != valueDictionary {
			return lerrors.TypeMismatch.Errorf("dictionary %s export %s is %s, want Dictionary", unit.Origin.Name, unit.Dictionary, kind)
		}
	}
	return nil
}

func (c *Compilation) dictionaryImportReady(statement *lStatement) (bool, error) {
	imported := statement.source.imported
	if imported == nil {
		return true, nil
	}
	if !strings.EqualFold(imported.module, imported.unit.Dictionary) {
		return false, lerrors.UnresolvedReference.Errorf("dictionary %s does not export %s", imported.unit.Origin.Name, imported.module)
	}
	binding, _, err := imported.unit.root.lookupLocal([]byte(imported.module))
	if err != nil || binding == nil {
		return false, err
	}
	kind, err := instructionType(binding.Record)
	if err != nil {
		return false, err
	}
	if kind != valueDictionary {
		return false, lerrors.TypeMismatch.Errorf("dictionary %s export %s is not a Dictionary", imported.unit.Origin.Name, imported.module)
	}
	if imported.member != "" {
		binding, _, err = c.resolver().lookupMember(binding.Record, []byte(imported.member))
		if err != nil || binding == nil {
			return false, err
		}
	}
	return true, nil
}
