package Lyger

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// sourceOccurrence retains expansion ancestry even when a file or its opening
// import emits no statements. Repeated paths receive independent identities.
type sourceOccurrence struct {
	ID         sourceOccurrenceID
	Parent     sourceOccurrenceID
	Path       string
	ImportSite SourceSpan
	Paragraphs []paragraphLayout
	Unit       *SourceUnit
	Origin     SourceOrigin
}

// builder expands source-defined bootstrap and imports before semantic work.
// SourceStatements follows BuiltBytes order and retains each originating span.
// SourcePaths includes every expanded file occurrence, even without statements.
type builder struct {
	main             *PreparedSource
	BuiltBytes       []byte
	TextCache        map[string][]byte
	SourceStatements []SourceStatement
	SourcePaths      []string
	Gram             *grammar
	occurrences      []sourceOccurrence
	nextOccurrence   sourceOccurrenceID
	units            []*SourceUnit
	dictionaries     map[string]*SourceUnit
	index            *dictionaryIndex
	origins          []SourceOrigin
}

func newPreparedBuilder(prepared *PreparedSource, g *grammar) *builder {
	return &builder{main: prepared, Gram: g, TextCache: make(map[string][]byte), dictionaries: make(map[string]*SourceUnit)}
}

func sourceIdentity(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", lerrors.IO.Wrap(err, "")
	}
	if realLink, err := filepath.EvalSymlinks(abs); err == nil {
		return realLink, nil
	}
	return filepath.Clean(abs), nil
}

func readPreparedSource(path string) (*PreparedSource, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, lerrors.IO.Wrap(err, "")
	}
	return PrepareSource(path, raw)
}

func (b *builder) appendStatement(statement SourceStatement) {
	b.SourceStatements = append(b.SourceStatements, statement)
	b.BuiltBytes = append(b.BuiltBytes, statement.Prepared...)
}

func (b *builder) expand(source *PreparedSource, unit *SourceUnit, parent sourceOccurrenceID, importSite SourceSpan, active map[string]bool, chain []string) error {
	var err error
	identity := "resource:" + source.Path
	if unit.Kind == SourceProgram {
		identity, err = sourceIdentity(source.Path)
		if err != nil {
			return err
		}
	}
	if active[identity] {
		return lerrors.Cycle.Errorf("import cycle: %s", strings.Join(append(chain, source.Path), " -> "))
	}
	active[identity] = true
	defer delete(active, identity)
	chain = append(chain, source.Path)
	origin := SourceOrigin{Identity: identity, Name: source.Path, SHA256: fmt.Sprintf("%x", sha256.Sum256(source.Original))}
	if unit.Kind == SourceProgram {
		origin.PhysicalPath = source.Path
	} else if b.Gram.assetRoot != "" {
		origin.PhysicalPath, err = filepath.Abs(filepath.Join(b.Gram.assetRoot, filepath.FromSlash(source.Path)))
		if err != nil {
			return err
		}
	}
	b.origins = append(b.origins, origin)
	b.nextOccurrence++
	occurrence := b.nextOccurrence
	b.occurrences = append(b.occurrences, sourceOccurrence{
		ID: occurrence, Parent: parent, Path: source.Path, ImportSite: importSite, Unit: unit, Origin: origin,
		Paragraphs: append([]paragraphLayout(nil), source.paragraphs...),
	})
	if unit.Kind == SourceProgram {
		b.SourcePaths = append(b.SourcePaths, source.Path)
	} else if b.Gram.assetRoot != "" {
		b.SourcePaths = append(b.SourcePaths, origin.PhysicalPath)
	}
	for key, value := range source.TextCache {
		b.TextCache[key] = bytes.Clone(value)
	}
	for _, statement := range source.Statements {
		statement.occurrence = occurrence
		statement.Unit = unit
		statement.Origin = origin
		isImport, imported, words, err := b.Gram.scanImportStatement(statement.Prepared)
		if err != nil {
			return statement.Span.wrapError(err)
		}
		if !isImport {
			b.appendStatement(statement)
			continue
		}
		name := imported.filePath
		if value, ok := source.TextCache["`"+name+"`"]; ok {
			name = unquoteImportWord(value)
		}
		if !strings.HasSuffix(name, ".l") {
			dictionary, err := b.loadDictionary(name, active, chain)
			if err != nil {
				return statement.Span.wrapError(err)
			}
			module, member := imported.module, imported.keyWord
			if module == "" {
				module, member = imported.keyWord, ""
			}
			if module == "" {
				module = dictionary.Dictionary
			}
			alias := imported.alias
			if alias == "" {
				alias = member
				if alias == "" {
					alias = module
				}
			}
			// Keep a readable binding projection, but resolve its left side only
			// in the imported unit. This does not expose that unit to the caller.
			use := string(words[0])
			as := b.Gram.lex.Prepositions["definement.declarative"].Root
			if imported.alias != "" {
				as = string(words[len(words)-2])
			}
			target := module
			if member != "" {
				target += "'s " + member
			}
			statement.Prepared = append(newLineStamp(statement.Span.Start.Line, statement.Span.Path), []byte(use+" "+target+" "+as+" "+strconv.Quote(alias)+".\n")...)
			statement.imported = &dictionaryImport{unit: dictionary, module: module, member: member}
			b.appendStatement(statement)
			continue
		}
		var child *PreparedSource
		if unit.Kind == SourceProgram {
			if !filepath.IsAbs(name) {
				name = filepath.Join(filepath.Dir(source.Path), name)
			}
			child, err = readPreparedSource(name)
		} else {
			if path.IsAbs(name) || strings.Contains(name, "\\") {
				return statement.Span.wrapError(lerrors.InvalidInput.Errorf("invalid resource include %s", name))
			}
			resolved := path.Join(path.Dir(source.Path), name)
			if !fs.ValidPath(resolved) || (resolved != unit.resourceRoot && !strings.HasPrefix(resolved, unit.resourceRoot+"/")) {
				return statement.Span.wrapError(lerrors.InvalidInput.Errorf("resource include %s escapes dictionary root", name))
			}
			child, _, err = b.resourceSource(resolved)
		}
		if err != nil {
			return statement.Span.wrapError(err)
		}
		if err = b.expand(child, unit, occurrence, statement.Span, active, chain); err != nil {
			return statement.Span.wrapError(err)
		}
		if imported.alias != "" {
			alias, err := imported.aliasStatement(statement, words)
			if err != nil {
				return statement.Span.wrapError(err)
			}
			b.appendStatement(alias)
		}
	}
	return nil
}

// aliasStatement removes the file clause from a validated import, retaining
// its configured operators and the caller's spelling, location and scope.
func (i *importStatement) aliasStatement(original SourceStatement, words [][]byte) (SourceStatement, error) {
	expected := 6
	if i.module != "" {
		expected++
	}
	if i.alias == "" || len(words) != expected {
		return SourceStatement{}, lerrors.Internal.Errorf("invalid imported alias structure")
	}
	from := len(words) - 4
	remaining := append([][]byte(nil), words[:from]...)
	remaining = append(remaining, words[from+2:]...)
	body := append(bytes.Join(remaining, []byte(" ")), '.', '\n')
	alias := original
	alias.Prepared = append(newLineStamp(original.Span.Start.Line, original.Span.Path), body...)
	return alias, nil
}

func (b *builder) process() error {
	if b.main == nil {
		return lerrors.Internal.Errorf("cannot build a nil prepared source")
	}
	if b.Gram == nil {
		return lerrors.Internal.Errorf("cannot build source without an import grammar")
	}
	bootstrap, origin, err := b.resourceSource("builtins/" + b.Gram.language + "/core.l")
	if err != nil {
		return err
	}
	builtin := &SourceUnit{Kind: SourceBuiltin, Origin: origin, resourceRoot: path.Dir(origin.Name)}
	if err := b.expand(bootstrap, builtin, 0, SourceSpan{}, make(map[string]bool), nil); err != nil {
		return err
	}
	b.units = append(b.units, builtin)
	identity, err := sourceIdentity(b.main.Path)
	if err != nil {
		return err
	}
	program := &SourceUnit{Kind: SourceProgram, Origin: SourceOrigin{Identity: identity, Name: b.main.Path, PhysicalPath: identity, SHA256: fmt.Sprintf("%x", sha256.Sum256(b.main.Original))}}
	if err := b.expand(b.main, program, 0, SourceSpan{}, make(map[string]bool), nil); err != nil {
		return err
	}
	b.units = append(b.units, program)
	// Dependency-first diagnostic and semantic ordering, without changing local
	// inclusion order or physical occurrence identities inside any unit.
	statements := b.SourceStatements
	b.SourceStatements, b.BuiltBytes = nil, nil
	for _, unit := range b.units {
		for _, statement := range statements {
			if statement.Unit == unit {
				b.appendStatement(statement)
			}
		}
	}
	return nil
}
