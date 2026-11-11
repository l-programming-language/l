package Lyger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"
	"testing/fstest"
)

// These are source-loading boundaries: native runtime:: references in Storage
// do not cause Runtime's supervisor to become a source dependency.
func TestLibraryFoundationSourceLayout(t *testing.T) {
	c := compileDefinitionSource(t, "Use Storage from `storage`.\nUse Formatting from `format`.\n", "library-layout.l")
	storage := libraryFoundationUnit(t, c, "Storage")
	format := libraryFoundationUnit(t, c, "Formatting")
	if storage.Origin.Name != "builtins/eng/storage/dictionary.l" || format.Origin.Name != "stdlib/eng/format/dictionary.l" {
		t.Fatalf("foundation entries remain in the old layout: %s, %s", storage.Origin.Name, format.Origin.Name)
	}
	want := []string{"bindings.l", "values.l", "control.l", "display.l", "collections.l", "program.l", "traversal.l", "execution.l"}
	var includes []string
	for _, occurrence := range c.SourceOccurrences {
		if occurrence.Unit == storage && occurrence.Origin.Name != storage.Origin.Name {
			includes = append(includes, strings.TrimPrefix(occurrence.Origin.Name, "builtins/eng/storage/"))
		}
	}
	if strings.Join(includes, ",") != strings.Join(want, ",") {
		t.Fatalf("ordered same-unit Storage includes = %v; want %v", includes, want)
	}
	for _, unit := range c.SourceUnits {
		if unit.Dictionary == "Runtime" || unit.Dictionary == "Numeric" {
			t.Fatalf("Formatting loaded unrelated source unit %s", unit.Dictionary)
		}
	}
	for _, statement := range c.SourceStatements {
		if strings.HasPrefix(statement.Origin.Name, "builtins/eng/storage/") && statement.Unit != storage {
			t.Fatalf("Storage include %s created a different lexical owner", statement.Origin.Name)
		}
		if strings.HasPrefix(statement.Origin.Name, "stdlib/eng/runtime/") {
			t.Fatalf("Formatting source-loaded process policy %s", statement.Origin.Name)
		}
	}
	for _, name := range []string{"Sum", "Agreement", "Choice", "First", "Second", "Integer", "Indexed", "Unsigned"} {
		if binding, _, err := c.Scope.Root.lookupLocal([]byte(name)); err != nil || binding != nil {
			t.Fatalf("foundation helper %s escaped into the program: %v", name, err)
		}
	}
	assertStorageEquivalent(t, dictionaryIndexFilesystem(t))
}

// The control is the same authored source loaded as one file. No historical
// checkout or duplicated fixture can silently become the expected implementation.
func assertStorageEquivalent(t *testing.T, files fstest.MapFS) {
	t.Helper()
	const root = "builtins/eng/storage/"
	const entry = root + "dictionary.l"
	joined := []byte("Create \"Storage\" as a Dictionary.\n")
	offsets := map[string]int{}
	for _, name := range []string{"bindings.l", "values.l", "control.l", "display.l", "collections.l", "program.l", "traversal.l", "execution.l"} {
		name = root + name
		offsets[name] = len(joined)
		joined = append(joined, files[name].Data...)
	}
	index, err := indexSource(joined)
	if err != nil {
		t.Fatal(err)
	}
	boundaries := map[int]bool{}
	for name, offset := range offsets {
		prepared, err := PrepareSource(name, files[name].Data)
		if err != nil {
			t.Fatal(err)
		}
		boundaries[offset+prepared.Statements[0].Span.Start.Offset] = true
	}
	compile := func() *Compilation {
		result, compileErr := NewLygerService(Options{Resources: files}).Compile([]byte("Use Storage from `storage`.\n"), "layout-equivalence.l")
		if compileErr != nil {
			t.Fatal(compileErr)
		}
		return result.Compilation
	}
	split := compile()
	files[entry] = &fstest.MapFile{Data: joined}
	for name := range offsets {
		delete(files, name)
	}
	intact := compile()
	// Compose every public Storage declaration and its selected dependency
	// closure, so lazily instantiated Process bodies also cross the same gate.
	aml := func(c *Compilation) []byte {
		result, err := c.ComposeLibrary("Storage", "storage")
		if err != nil {
			t.Fatal(err)
		}
		return result
	}
	if !bytes.Equal(aml(split), aml(intact)) {
		t.Fatal("split Storage changed selected composed bodies")
	}
	normalize := func(span SourceSpan) SourceSpan {
		if offset, ok := offsets[span.Path]; ok {
			span.Path = entry
			span.Start, span.Paragraph = index.location(span.Start.Offset + offset)
			if span.End != (SourcePosition{}) {
				span.End, _ = index.location(span.End.Offset + offset)
			}
		}
		// Order is local to the original file; the comparison retains global
		// statement order independently. All byte offsets/line/column remain.
		if span.Path == entry {
			span.Order = 0
		}
		return span
	}
	want := storageSemanticView(t, intact, normalize, boundaries)
	if got := storageSemanticView(t, split, normalize, boundaries); !bytes.Equal(got, want) {
		at := 0
		for at < len(got) && at < len(want) && got[at] == want[at] {
			at++
		}
		t.Fatalf("split Storage differs from intact source at byte %d:\nsplit: %.160s\nintact: %.160s", at, got[at:], want[at:])
	}
	t.Logf("compared %d ordered statements, %d declarations, %d descriptors, %d functions, %d pronoun bindings", len(split.Statements), len(split.Semantics.Declarations), len(split.Semantics.Descriptors), len(split.Semantics.Functions), len(split.Scope.Pronouns))
}

// Project the real typed graph, preserving pointer identity through ordered
// token/declaration references. Only physical include coordinates and parser
// configuration pointers are normalized; bodies and binding edges are retained.
func storageSemanticView(t *testing.T, c *Compilation, normalize func(SourceSpan) SourceSpan, boundaries map[int]bool) []byte {
	t.Helper()
	refs := map[uintptr]string{}
	bind := func(value any, name string) {
		pointer := reflect.ValueOf(value).Pointer()
		if pointer != 0 {
			refs[pointer] = name
		}
	}
	var scopes []*lexicalScope
	var scope func(*lexicalScope)
	scope = func(s *lexicalScope) {
		if s == nil || refs[reflect.ValueOf(s).Pointer()] != "" {
			return
		}
		scope(s.parent)
		bind(s, fmt.Sprintf("scope:%d", len(scopes)))
		scopes = append(scopes, s)
	}
	for _, unit := range c.SourceUnits {
		bind(unit, "unit:"+string(unit.Kind)+":"+unit.Dictionary)
		scope(unit.root)
	}
	for i, statement := range c.Statements {
		bind(statement, fmt.Sprintf("statement:%d", i))
		for j, tok := range statement.tokens {
			bind(tok, fmt.Sprintf("token:%d:%d", i, j))
			scope(tok.scope)
		}
	}
	var records []*record
	for rec := range c.Semantics.Declarations {
		records = append(records, rec)
	}
	sort.Slice(records, func(i, j int) bool {
		a, b := records[i], records[j]
		left := refs[reflect.ValueOf(c.Semantics.Declarations[a]).Pointer()] + ":" + a.ID
		right := refs[reflect.ValueOf(c.Semantics.Declarations[b]).Pointer()] + ":" + b.ID
		return left < right
	})
	for i, rec := range records {
		bind(rec, fmt.Sprintf("record:%d", i))
		if rec.function != nil {
			bind(rec.function, fmt.Sprintf("function:%d", i))
		}
	}
	var nodes []any
	var alive []reflect.Value
	var view func(reflect.Value) any
	view = func(value reflect.Value) any {
		if !value.IsValid() {
			return nil
		}
		if value.Type() == reflect.TypeOf(SourceSpan{}) {
			position := func(name string) SourcePosition {
				v := value.FieldByName(name)
				return SourcePosition{Offset: int(v.FieldByName("Offset").Int()), Line: int(v.FieldByName("Line").Int()), Column: int(v.FieldByName("Column").Int())}
			}
			return normalize(SourceSpan{Path: value.FieldByName("Path").String(), Start: position("Start"), End: position("End"), Paragraph: int(value.FieldByName("Paragraph").Int()), Order: int(value.FieldByName("Order").Int())})
		}
		switch value.Kind() {
		case reflect.Pointer:
			if value.IsNil() {
				return nil
			}
			if name := refs[value.Pointer()]; name != "" {
				return name
			}
			if value.Type() == reflect.TypeOf((*grammar)(nil)) || value.Type() == reflect.TypeOf((*grammeme)(nil)) {
				return value.Type().String()
			}
			index := len(nodes)
			alive = append(alive, value)
			name := fmt.Sprintf("node:%d", index)
			refs[value.Pointer()] = name
			nodes = append(nodes, nil)
			nodes[index] = view(value.Elem())
			return name
		case reflect.Interface:
			return view(value.Elem())
		case reflect.Struct:
			fields := map[string]any{}
			for i := 0; i < value.NumField(); i++ {
				fields[value.Type().Field(i).Name] = view(value.Field(i))
			}
			return fields
		case reflect.Slice, reflect.Array:
			if value.Kind() == reflect.Slice && value.IsNil() {
				return nil
			}
			items := make([]any, value.Len())
			for i := range items {
				items[i] = view(value.Index(i))
			}
			return items
		case reflect.Map:
			if value.IsNil() {
				return nil
			}
			items := map[string]any{}
			keys := map[string]reflect.Value{}
			for _, key := range value.MapKeys() {
				encoded, err := json.Marshal(view(key))
				if err != nil {
					t.Fatal(err)
				}
				keys[string(encoded)] = key
			}
			ordered := make([]string, 0, len(keys))
			for key := range keys {
				ordered = append(ordered, key)
			}
			sort.Strings(ordered)
			for _, key := range ordered {
				items[key] = view(value.MapIndex(keys[key]))
			}
			return items
		case reflect.String:
			return value.String()
		case reflect.Bool:
			return value.Bool()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return value.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return value.Uint()
		case reflect.Float32, reflect.Float64:
			return value.Float()
		default:
			t.Fatalf("unobserved semantic field %s", value.Type())
			return nil
		}
	}
	var rows []any
	add := func(value any) { rows = append(rows, view(reflect.ValueOf(value))) }
	for _, statement := range c.Statements {
		add(statement.sType)
		add(statement.source.Span)
		add(statement.source.Unit)
		add(statement.sGram)
		for _, tok := range statement.tokens {
			add(*tok)
			add(c.Scope.Pronouns[tok])
			add(c.Semantics.Subjects[tok])
		}
	}
	for _, ordered := range c.Scope.Ordered {
		if span := normalize(ordered.Source); span.Path == "builtins/eng/storage/dictionary.l" && boundaries[span.Start.Offset] {
			// A new file resets the otherwise unused incoming subject. The
			// actual paragraph pronoun targets and typed bodies above still match.
			for _, tok := range ordered.Statement.tokens {
				if binding, ok := c.Scope.Pronouns[tok]; ok && binding.Kind == scopeParagraph {
					t.Fatal("Storage include boundary consumes a preceding paragraph subject")
				}
			}
			ordered.IncomingSubjectID, ordered.IncomingSubjectToken = "", nil
		}
		add(ordered)
	}
	for _, s := range scopes {
		add(s.kind)
		add(s.unit)
		add(s.source)
		add(s.parent)
		add(s.function)
		add(s.callableHeader)
		add(s.records)
	}
	for _, rec := range records {
		selected, err := c.Semantics.selectedDefinition(rec)
		if err != nil {
			t.Fatal(err)
		}
		add(*rec)
		add(selected)
		add(c.Semantics.Aliases[rec])
		add(c.Semantics.Descriptors[rec])
		if rec.function != nil {
			add(*rec.function)
		}
	}
	raw, err := json.Marshal([]any{rows, nodes})
	runtime.KeepAlive(alive)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
