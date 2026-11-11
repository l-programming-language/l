package Lyger

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
)

type glossaryDefinition struct {
	Source string `json:"source"`
	Lines  [2]int `json:"lines"`
}

type glossaryForm struct {
	PartOfSpeech string `json:"partOfSpeech"`
	Role         string `json:"role"`
}

type glossaryWord struct {
	PartOfSpeech string                  `json:"partOfSpeech"`
	FormRoles    map[string]glossaryForm `json:"formRoles"`
	Definitions  []glossaryDefinition    `json:"definitions"`
	Aliases      []string                `json:"aliases"`
	Targets      []struct {
		Entry string `json:"entry"`
		Form  string `json:"form"`
	} `json:"targets"`
}

type glossaryEntry struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Sources     []string      `json:"sources"`
	Inflections []string      `json:"inflections"`
	Word        *glossaryWord `json:"word"`
}

// This tests lexical coverage and source provenance, not native behavior. Publicly
// selects native library roots; external L member lookup also reaches declarations
// without that modifier. Walking the actual canonical member graph keeps that
// distinction and catches new nested members, aliases, and source-unit changes.
func TestWordGlossarySourceInventory(t *testing.T) {
	options := canonicalTestOptions()
	c := NewLygerService(options).NewCompilation([]byte("Create \"Glossary\" as a New Basic App.\n"), "word-glossary-inventory.l")
	c.Make()
	if c.Error != nil {
		t.Fatal(c.Error)
	}
	meanings := make(map[string]*record)
	aliases := make(map[string][]string)
	paths := make(map[string]*record)
	var visit func(string, *record, map[*record]bool)
	visit = func(name string, rec *record, ancestors map[*record]bool) {
		canonical, canonicalErr := canonicalRecord(rec, c.Semantics.Aliases)
		if canonicalErr != nil {
			t.Fatal(canonicalErr)
		}
		if previous := paths[name]; previous != nil {
			if previous != canonical {
				t.Fatalf("external path %s selects two canonical records", name)
			}
			return
		}
		paths[name] = canonical
		meanings[canonical.ID] = canonical
		if name != canonical.ID {
			aliases[canonical.ID] = append(aliases[canonical.ID], name)
		}
		if ancestors[canonical] {
			t.Fatalf("cyclic member namespace at %s", name)
		}
		ancestors[canonical] = true
		defer delete(ancestors, canonical)
		members := make([]string, 0, len(canonical.members))
		for member := range canonical.members {
			members = append(members, member)
		}
		sort.Strings(members)
		for _, member := range members {
			rec := canonical.members[member]
			if rec == canonical {
				continue
			}
			tail := strings.TrimPrefix(member, canonical.ID+".")
			resolver := c.resolver()
			// A passive spelling can also have its own adjective declaration.
			// Check the declaration's grammatical role, not an untyped lookup
			// that would incorrectly discard one of those two meanings.
			if word := rec.bindingWord(); word != nil {
				resolver.expectedPOS, _ = parsePartOfSpeech(word.POS)
			}
			binding, _, lookupErr := resolver.lookupMember(canonical, []byte(tail))
			if lookupErr != nil || binding == nil {
				t.Fatalf("external member %s.%s: binding=%v error=%v", name, tail, binding, lookupErr)
			}
			selected, selectedErr := canonicalRecord(binding.Record, c.Semantics.Aliases)
			expected, expectedErr := canonicalRecord(rec, c.Semantics.Aliases)
			if selectedErr != nil || expectedErr != nil || selected != expected {
				t.Fatalf("external member %s.%s lost canonical identity", name, tail)
			}
			visit(name+"."+tail, rec, ancestors)
		}
	}
	// Discover core declarations from the selected builtin unit rather than a
	// list of names. Dictionary-private roots and Function/Process locals never
	// enter this external graph.
	for rec, statement := range c.Semantics.Declarations {
		if statement.source.Unit.Kind != SourceBuiltin {
			continue
		}
		canonical, canonicalErr := canonicalRecord(rec, c.Semantics.Aliases)
		if canonicalErr != nil {
			t.Fatal(canonicalErr)
		}
		if canonical != rec {
			continue
		}
		selected, lookupErr := c.builtinRecord(rec.ID)
		if lookupErr != nil || selected != rec {
			continue
		}
		visit(rec.ID, rec, make(map[*record]bool))
	}
	data, err := os.ReadFile("../docs/_meta/semantics.json")
	if err != nil {
		t.Fatal(err)
	}
	var metadata struct {
		Entries []glossaryEntry `json:"entries"`
	}
	if err = json.Unmarshal(data, &metadata); err != nil {
		t.Fatal(err)
	}
	entries := make(map[string]glossaryEntry)
	entryIDs := make(map[string]glossaryEntry)
	for _, entry := range metadata.Entries {
		if entry.Word == nil {
			continue
		}
		if prior, ok := entries[entry.Name]; ok && meanings[entry.Name] != nil {
			t.Errorf("duplicate canonical meaning %s in %s and %s", entry.Name, prior.ID, entry.ID)
		}
		entries[entry.Name] = entry
		entryIDs[entry.ID] = entry
		for _, definition := range entry.Word.Definitions {
			if strings.HasSuffix(definition.Source, ".l") && meanings[entry.Name] == nil {
				t.Errorf("documented source meaning %s is not externally reachable", entry.Name)
			}
		}
	}
	names := make([]string, 0, len(meanings))
	for name := range meanings {
		names = append(names, name)
	}
	sort.Strings(names)
	var missing []string
	for _, name := range names {
		rec := meanings[name]
		entry, ok := entries[name]
		if !ok {
			missing = append(missing, name)
			continue
		}
		t.Run(name, func(t *testing.T) {
			wantAliases := aliases[name]
			gotAliases := append([]string(nil), entry.Word.Aliases...)
			sort.Strings(wantAliases)
			sort.Strings(gotAliases)
			if strings.Join(gotAliases, "\n") != strings.Join(wantAliases, "\n") {
				t.Errorf("aliases = %v, reachable aliases = %v", gotAliases, wantAliases)
			}
			for _, alias := range wantAliases {
				if _, duplicate := entries[alias]; duplicate {
					t.Errorf("alias %s also has a canonical entry", alias)
				}
			}
			word := rec.bindingWord()
			pos := "noun"
			if word != nil {
				pos = word.POS
			}
			if entry.Word.PartOfSpeech != pos {
				t.Errorf("part of speech = %s, source = %s", entry.Word.PartOfSpeech, pos)
			}
			root := name[strings.LastIndex(name, ".")+1:]
			wantForms := map[string]string{root: pos}
			wantCases := map[string][]string{root: {"root"}}
			if word != nil {
				for _, declared := range word.Forms {
					role := declared.Role
					formPOS := pos
					formCase := role
					if pos == "noun" && role == "active" {
						formPOS = "verb"
						formCase = "applicative"
					} else if pos == "noun" && role == "descriptive" {
						formPOS = "adjective"
					} else if pos == "verb" && role == "objective" {
						formPOS = "noun"
						formCase = "actor"
					}
					form := declared.spelling(root)
					wantForms[form] = formPOS
					wantCases[form] = append(wantCases[form], formCase)
				}
			}
			gotForms := make(map[string]string)
			for _, form := range entry.Inflections {
				role, found := entry.Word.FormRoles[form]
				if !found || role.Role == "" {
					t.Errorf("form %s has no grammatical role", form)
				}
				for _, wantCase := range wantCases[form] {
					if !strings.Contains(role.Role, wantCase) {
						t.Errorf("form %s omits declared %s case", form, wantCase)
					}
				}
				gotForms[form] = role.PartOfSpeech
			}
			if !reflect.DeepEqual(gotForms, wantForms) || len(entry.Word.FormRoles) != len(wantForms) {
				t.Errorf("form parts of speech = %v, source = %v", gotForms, wantForms)
			}
			wantDefinitions := glossarySourceDefinitions(t, c, rec)
			var gotDefinitions []glossaryDefinition
			for _, got := range entry.Word.Definitions {
				if strings.HasSuffix(got.Source, ".l") {
					gotDefinitions = append(gotDefinitions, got)
				}
			}
			sort.Slice(gotDefinitions, func(i, j int) bool {
				if gotDefinitions[i].Source != gotDefinitions[j].Source {
					return gotDefinitions[i].Source < gotDefinitions[j].Source
				}
				return gotDefinitions[i].Lines[0] < gotDefinitions[j].Lines[0]
			})
			if !reflect.DeepEqual(gotDefinitions, wantDefinitions) {
				t.Errorf("source definitions = %+v, actual declarations = %+v", gotDefinitions, wantDefinitions)
			}
			for _, want := range wantDefinitions {
				listed := false
				for _, path := range entry.Sources {
					listed = listed || path == want.Source
				}
				if !listed {
					t.Errorf("definition source %s is absent from entry sources", want.Source)
				}
			}
			if descriptor := c.Semantics.Descriptors[rec]; descriptor != nil {
				wantTargets := make(map[string]bool)
				for target := range descriptor.Targets {
					wantTargets[fmt.Sprint(target)] = true
				}
				gotTargets := make(map[string]bool)
				for _, target := range entry.Word.Targets {
					meaning, found := entryIDs[target.Entry]
					if !found {
						t.Errorf("descriptor target entry %s is absent", target.Entry)
						continue
					}
					if target.Form == "" {
						target.Form = meaning.Name[strings.LastIndex(meaning.Name, ".")+1:]
					}
					if _, found := meaning.Word.FormRoles[target.Form]; !found {
						t.Errorf("descriptor target %s has no form %s", meaning.Name, target.Form)
					}
					gotTargets[strings.ToLower(meaning.Name)] = true
				}
				if !reflect.DeepEqual(gotTargets, wantTargets) {
					t.Errorf("descriptor targets = %v, source = %v", gotTargets, wantTargets)
				}
			}
		})
	}
	if len(missing) != 0 {
		t.Errorf("%d externally reachable canonical meanings lack glossary entries:\n%s", len(missing), strings.Join(missing, "\n"))
	}
	t.Logf("checked %d external paths coalesced to %d source-owned meanings", len(paths), len(meanings))
}

func glossarySourceDefinitions(t *testing.T, c *Compilation, rec *record) []glossaryDefinition {
	t.Helper()
	statements := make(map[*lStatement]bool)
	if statement := c.Semantics.Declarations[rec]; statement != nil {
		statements[statement] = true
	}
	if descriptor := c.Semantics.Descriptors[rec]; descriptor != nil {
		for _, rule := range descriptor.Rules {
			statements[rule.Source] = true
		}
	}
	if function := rec.function; function != nil {
		for _, ordered := range c.Scope.Ordered {
			for scope := ordered.Scope; scope != nil; scope = scope.parent {
				if scope == function.Scope {
					statements[ordered.Statement] = true
					break
				}
			}
		}
	}
	spans := make([]SourceSpan, 0, len(statements))
	for statement := range statements {
		if statement != nil && statement.source != nil {
			spans = append(spans, statement.source.Span)
		}
	}
	sort.Slice(spans, func(i, j int) bool {
		if spans[i].Path != spans[j].Path {
			return spans[i].Path < spans[j].Path
		}
		return spans[i].Start.Offset < spans[j].Start.Offset
	})
	var definitions []glossaryDefinition
	for i := 0; i < len(spans); {
		start, end := spans[i], spans[i]
		j := i + 1
		for j < len(spans) && spans[j].Path == start.Path && spans[j].Start.Offset <= end.End.Offset {
			if spans[j].End.Offset > end.End.Offset {
				end = spans[j]
			}
			j++
		}
		if _, err := fs.Stat(c.Gram.resources, start.Path); err != nil {
			t.Fatal(err)
		}
		last := end.End.Line
		if end.End.Column == 1 {
			last--
		}
		definitions = append(definitions, glossaryDefinition{Source: start.Path, Lines: [2]int{start.Start.Line, last}})
		i = j
	}
	if len(definitions) == 0 {
		t.Fatalf("source meaning %s has no declaration provenance", rec.ID)
	}
	return definitions
}
