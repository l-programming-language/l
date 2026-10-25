package Lyger

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

// These cases use one sentence per physical line and ordinary lexical scopes.
// The root Witness avoids making zero-App support a prerequisite for binding.
type paragraphBindingDeclaration struct {
	record    *record
	statement *lStatement
}

func paragraphBindingDeclarations(t *testing.T, compilation *Compilation, name string, count int) []paragraphBindingDeclaration {
	t.Helper()
	var declarations []paragraphBindingDeclaration
	for _, statement := range compilation.Statements {
		if filepath.Base(statement.source.Span.Path) != "main.l" {
			continue
		}
		for _, tok := range statement.tokens {
			if !strings.EqualFold(tok.declarationRoot(), name) {
				continue
			}
			rec := compilation.Semantics.Subjects[tok]
			if rec == nil {
				t.Fatalf("%s:%d: %s declaration lost its canonical record", statement.source.Span.Path, statement.source.Span.Start.Line, name)
			}
			declarations = append(declarations, paragraphBindingDeclaration{rec, statement})
		}
	}
	if len(declarations) != count {
		t.Fatalf("%s declarations = %d; want %d", name, len(declarations), count)
	}
	return declarations
}

func paragraphBindingUses(t *testing.T, compilation *Compilation, spelling string, want []*record) {
	t.Helper()
	var actual []*record
	for _, statement := range compilation.Statements {
		if filepath.Base(statement.source.Span.Path) != "main.l" {
			continue
		}
		for _, tok := range statement.tokens {
			if string(tok.rawBytes) != spelling || tok.isDeclare() {
				continue
			}
			if tok.binding == nil {
				t.Fatalf("%s:%d: %s use has no binding", statement.source.Span.Path, statement.source.Span.Start.Line, spelling)
			}
			rec, err := canonicalRecord(tok.binding.Record, compilation.Semantics.Aliases)
			if err != nil {
				t.Fatal(err)
			}
			actual = append(actual, rec)
		}
	}
	if len(actual) != len(want) {
		t.Fatalf("%s uses = %d; want %d", spelling, len(actual), len(want))
	}
	for index, rec := range actual {
		if rec != want[index] {
			t.Errorf("%s use %d bound record %p, want its lexical record %p", spelling, index+1, rec, want[index])
		}
	}
}

// A test-owned native definition gives namespace/nominal tests distinct source
// declarations without depending on IO's intentionally private Log/Logger.
const paragraphAnnounce = "Define \"Announce|-s|-d|-ing|-r\" to `<DEF>BOD^E{|Statement: ~::std::println!(\"{}\", <@Output|Thing@>);~|}<DEF>`.\n"

func TestParagraphBindingPrivateNominalsDistinguishIdentity(t *testing.T) {
	source := paragraphSource(t)
	source.main = paragraphWitness
	for _, body := range []string{
		`BOD^E{|Statement: ~::std::print!("{}", <@Output|Thing@>);~|}`,
		`BOD^E{|Statement: ~::std::println!("{}", <@Output|Thing@>);~|}`,
	} {
		source.main += "\n    Define \"Pulse|-s|-d|-ing\" as `<DEF>" + body + "<DEF>`.\nUse Pulse as \"Stroke\".\n"
	}
	compilation := compileMemberResolutionSource(t, source)
	if compilation.Error != nil {
		t.Fatal(compilation.Error)
	}
	definitions := paragraphBindingDeclarations(t, compilation, "Pulse", 2)
	aliases := paragraphBindingDeclarations(t, compilation, "Stroke", 2)
	if definitions[0].record == definitions[1].record {
		t.Fatal("same spelling collapsed two private nominal definitions")
	}
	args := []DefinitionArgument{{Kind: "Text", Text: "nominal"}}
	instances := make([]DefinitionInstance, len(definitions))
	for index, declaration := range definitions {
		canonical, err := canonicalRecord(aliases[index].record, compilation.Semantics.Aliases)
		if err != nil || canonical != declaration.record {
			t.Fatalf("scope %d valid nominal alias changed identity: %v", index+1, err)
		}
		instances[index], err = compilation.instantiateCall(aliases[index].record, args, make(map[*record]bool))
		if err != nil {
			t.Fatal(err)
		}
		if instances[index].Symbol != declaration.record.ID || instances[index].Span.Start.Line != declaration.statement.source.Span.Start.Line {
			t.Fatal("nominal alias changed definition spelling or source attribution")
		}
	}
	if instances[0].Expression == instances[1].Expression {
		t.Fatal("distinct local templates produced the same expansion")
	}
	for index, alias := range aliases {
		parameter := &objectGram{noun: &token{binding: &tokenBinding{Record: alias.record}}}
		constructor := &constructionNode{Nominal: "holder", Parameters: []*objectGram{parameter}}
		if err := compilation.validateConstructionParameters(constructor, []DefinitionInstance{instances[index]}); err != nil {
			t.Fatalf("valid local nominal alias rejected: %v", err)
		}
		if err := compilation.validateConstructionParameters(constructor, []DefinitionInstance{instances[1-index]}); err == nil {
			t.Errorf("constructor accepted sibling definition solely because both are named Pulse (scope %d)", index+1)
		}
	}
	for _, name := range []string{"Pulse", "Stroke"} {
		if compilation.Memory[name] != nil {
			t.Errorf("private nominal %s leaked into root memory", name)
		}
	}
}

func TestParagraphBindingPrivateAppsKeepRunTargetsAndBodies(t *testing.T) {
	source := paragraphSource(t)
	source.main = paragraphWitness
	for _, text := range []string{"first private body", "second private body"} {
		source.main += "\n    \"message\" is `" + text + "`.\n" +
			"Create \"Worker\" as a New Basic App.\nThe App Prints message.\nRun Worker.\n"
	}
	compilation := compileMemberResolutionSource(t, source)
	if compilation.Error != nil {
		t.Fatal(compilation.Error)
	}
	workers := paragraphBindingDeclarations(t, compilation, "Worker", 2)
	if workers[0].record == workers[1].record || workers[0].record.Value.Applicative.App == workers[1].record.Value.Applicative.App {
		t.Fatal("private Worker declarations share an App")
	}
	paragraphBindingUses(t, compilation, "Worker", []*record{workers[0].record, workers[1].record})
	var runs []*instruction
	var collectRuns func([]*instruction)
	collectRuns = func(instructions []*instruction) {
		for _, current := range instructions {
			if current.Kind == instructionRun {
				runs = append(runs, current)
			}
			if current.Block != nil {
				collectRuns(current.Block.Body)
			}
			collectRuns(current.Actions)
			collectRuns(current.Then)
			collectRuns(current.Else)
		}
	}
	collectRuns(compilation.Semantics.Instructions)
	if len(runs) != len(workers) {
		t.Fatalf("local Run instructions = %d; want %d", len(runs), len(workers))
	}
	for index, run := range runs {
		if run.Target == nil || run.Target.Record != workers[index].record || !run.Synchronous {
			t.Errorf("local Run %d lost its canonical App target or synchronous semantics", index+1)
		}
	}
	if compilation.Memory["Worker"] != nil || compilation.Memory["message"] != nil {
		t.Error("private App or body argument leaked into root memory")
	}
	result := &CompilationResult{Compilation: compilation}
	units, err := result.ApplicationUnits()
	if err != nil {
		t.Fatal(err)
	}
	if len(units) != 3 {
		t.Fatalf("application units = %d; want Witness and two private Workers", len(units))
	}
	seen := make(map[string]bool)
	for index, unit := range units {
		if unit.ID == "" || seen[unit.ID] {
			t.Fatalf("source-distinct Apps have missing or colliding build identity %q", unit.ID)
		}
		seen[unit.ID] = true
		if index == 0 {
			continue
		}
		own, other := "first private body", "second private body"
		if index == 2 {
			own, other = other, own
		}
		if !bytes.Contains(unit.BackendAML, []byte(own)) || bytes.Contains(unit.BackendAML, []byte(other)) {
			t.Errorf("private Worker %d has another scope's body", index)
		}
	}
}

func TestParagraphBindingNativeTypeMembersUseLexicalScope(t *testing.T) {
	for _, tc := range []struct {
		name, declaration, use, member string
		dictionary                     bool
	}{
		{"scalar", `Define App's "Label" as 1.`, `Show App's Label.`, "Label", false},
		{"dictionary", `Create App's "Tools" as a Dictionary.`, `Use App's Tools as "Alias".`, "Tools", true},
	} {
		for _, local := range []bool{false, true} {
			scopeName, indent := "root", ""
			if local {
				scopeName, indent = "local", "    "
			}
			t.Run(scopeName+"_"+tc.name, func(t *testing.T) {
				compilation := paragraphPublicationSource(t, indent+tc.declaration+"\n"+tc.use+"\n")
				if compilation.Error != nil {
					t.Fatal(compilation.Error)
				}
				member := paragraphBindingDeclarations(t, compilation, tc.member, 1)[0].record
				if member.ID != "App."+tc.member || member.storageScope == nil || member.storageScope.records[member.ID] != member {
					t.Fatal("native type member lost its qualified identity or owning lexical scope")
				}
				paragraphBindingUses(t, compilation, tc.member, []*record{member})
				if local {
					if member.storageScope == compilation.Scope.Root || compilation.Memory[member.ID] != nil || compilation.Memory["Alias"] != nil {
						t.Fatal("private native type member or alias escaped its lexical scope")
					}
				} else if member.storageScope != compilation.Scope.Root || compilation.Memory[member.ID] != member {
					t.Fatal("root native type member lost its canonical root binding")
				}
				if compilation.Memory["App"] != nil || compilation.Memory[tc.member] != nil {
					t.Fatal("native type namespace became an instance or unqualified member")
				}
				if tc.dictionary {
					alias := paragraphBindingDeclarations(t, compilation, "Alias", 1)[0].record
					canonical, err := canonicalRecord(alias, compilation.Semantics.Aliases)
					if err != nil || canonical != member || alias.Value.Iterative.Dictionary != member.Value.Iterative.Dictionary || member.Value.Iterative.Dictionary.ID != member.ID {
						t.Fatalf("native type member alias cloned, renamed, or detached its canonical Dictionary: %v", err)
					}
				}
			})
		}
	}
}

func TestParagraphBindingNativeTypeForwardMemberShadowsOuter(t *testing.T) {
	const source = "Define App's \"Label\" as 1.\n\n" +
		"    \"scopeMarker\" is 0.\nShow App's Label.\nDefine App's \"Label\" as 2.\nShow App's Label.\n\n" +
		"Show App's Label.\n"
	compilation := paragraphPublicationSource(t, source)
	if compilation.Error != nil {
		t.Fatal(compilation.Error)
	}
	declarations := paragraphBindingDeclarations(t, compilation, "Label", 2)
	outer, local := declarations[0].record, declarations[1].record
	if outer == local || outer.ID != "App.Label" || local.ID != outer.ID || outer.storageScope == local.storageScope {
		t.Fatal("same native type member spelling collapsed source-distinct declarations")
	}
	paragraphBindingUses(t, compilation, "Label", []*record{local, local, outer})
	if declarations[1].statement.source.Span.Start.Line != 5 || compilation.Memory["App.Label"] != outer || compilation.Memory["scopeMarker"] != nil {
		t.Fatal("forward binding changed declaration provenance or root visibility")
	}
}

func TestParagraphBindingNativeTypeMembersDoNotLeakToSiblings(t *testing.T) {
	for _, tc := range []struct{ name, declaration, use, member string }{
		{"scalar", `Define App's "PrivateValue" as 1.`, `Show App's PrivateValue.`, "PrivateValue"},
		{"dictionary", `Create App's "PrivateTools" as a Dictionary.`, `Use App's PrivateTools as "Alias".`, "PrivateTools"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// The successful read on line 2 makes this a sibling-escape check,
			// not a rejection caused by an unusable declaration in its own scope.
			source := "    " + tc.declaration + "\n" + tc.use + "\n\n    " + tc.use + "\n"
			compilation := paragraphPublicationSource(t, source)
			if compilation.Error == nil || !strings.Contains(compilation.Error.Error(), "main.l:4:5") || !strings.Contains(compilation.Error.Error(), "unresolved") {
				t.Fatalf("sibling native type member reference error = %v; want unresolved reference at line 4, column 5", compilation.Error)
			}
			member := paragraphBindingDeclarations(t, compilation, tc.member, 1)[0].record
			for _, statement := range compilation.Statements {
				if statement.source.Span.Start.Line != 2 {
					continue
				}
				for _, tok := range statement.tokens {
					if string(tok.rawBytes) == tc.member && (tok.binding == nil || tok.binding.Record != member) {
						t.Fatal("native type member's valid local read did not retain its canonical record")
					}
				}
			}
			if compilation.Memory[member.ID] != nil || compilation.Memory["Alias"] != nil {
				t.Fatal("private native type member or alias leaked to root memory")
			}
		})
	}
}

// This is an explicit metadata invariant for scoped resolution, separate from
// the source-level regression: missing scope must not silently select the root.
func TestParagraphBindingRequiresTokenScope(t *testing.T) {
	for _, operation := range []string{"lookup", "bind"} {
		t.Run(operation, func(t *testing.T) {
			root := &lexicalScope{records: make(map[string]*record)}
			root.records["UnscopedName"] = &record{ID: "UnscopedName", Type: "primitive.number", Value: &entry{Type: storagePrimitive, Primitive: newPrimitive(valueNumber)}, storageScope: root}
			compilation := &Compilation{Memory: root.records, Scope: &scopePlan{Root: root}, Semantics: newSemanticProgram()}
			originalValue := &entry{Type: storagePrimitive, Primitive: newPrimitive(valueUnknown)}
			tok := &token{rawBytes: []byte("UnscopedName"), tType: "primitive.unknown", mem: originalValue}
			var err error
			var state lookupState
			if operation == "lookup" {
				var binding *tokenBinding
				binding, state, err = compilation.resolver().lookup(tok)
				if binding != nil {
					t.Error("lookup without source scope returned a root binding")
				}
			} else {
				state, err = compilation.resolver().bindToken(tok)
			}
			if err == nil || !strings.Contains(err.Error(), "scope") || !strings.Contains(err.Error(), "UnscopedName") || state != lookupAbsent {
				t.Errorf("%s without scope = %s, %v; want contextual missing-scope error", operation, state, err)
			}
			if tok.mem != originalValue || tok.binding != nil || tok.defined || tok.tType != "primitive.unknown" || string(tok.rawBytes) != "UnscopedName" {
				t.Error("missing-scope failure changed token storage, identity, or spelling")
			}
			if operation == "bind" && tok.bindingError != err {
				t.Error("binding did not retain the missing-scope diagnostic")
			}
		})
	}
}

func TestParagraphBindingDelayedInflectionShadowsOuter(t *testing.T) {
	source := paragraphSource(t)
	source.main = paragraphWitness + paragraphAnnounce +
		"Create \"Launch|-es|-ed|-ing\" for Processing a New Announcer.\n" +
		"\n    \"scopeMarker\" is `private`.\n" +
		"Create \"Launch|-es|-ed|-ing\" for Processing a New Whisperer.\n" +
		"Create \"Reader\" as a New Basic App.\nThe App Launches x2.\n" +
		"Define \"Whisper|-s|-ed|-ing|-er\" to `<DEF>BOD^E{|Statement: ~::std::print!(\"{}\", <@Output|Thing@>);~|}<DEF>`.\n" +
		"\nCreate \"OuterReader\" as a New Basic App.\nThe App Launches x1.\n"
	compilation := compileMemberResolutionSource(t, source)
	if compilation.Error != nil {
		t.Fatal(compilation.Error)
	}
	definitions := paragraphBindingDeclarations(t, compilation, "Launch", 2)
	outer, local := definitions[0].record, definitions[1].record
	if outer == local {
		t.Fatal("delayed local definition reused the enclosing record")
	}
	paragraphBindingUses(t, compilation, "Launches", []*record{local, outer})
	if compilation.Memory["Launch"] != outer || compilation.Memory["Whisper"] != nil || compilation.Memory["scopeMarker"] != nil {
		t.Error("delayed local publication changed the root environment")
	}
	for _, output := range []string{`::std::print!("{}", "Hello Again World!");`, `::std::println!("{}", "Hello World!");`} {
		if !bytes.Contains(compilation.Program.BackendBytes, []byte(output)) {
			t.Errorf("lexically selected definition did not produce %s", output)
		}
	}
}

func TestParagraphBindingSiblingDictionaryMembersKeepOwners(t *testing.T) {
	source := paragraphSource(t)
	source.main = paragraphWitness + paragraphAnnounce
	for index, text := range []string{"x1", "x2"} {
		source.main += fmt.Sprintf("\n    Create \"Tools\" as a Dictionary.\n"+
			"Create Tools's \"Speak|-s|-ed|-ing|-er\" for Processing a New Announcer.\n"+
			"Use Tools's Speak as \"Display\".\n"+
			"Create \"Reader%d\" as a New Basic App.\nThe App Displays %s.\n", index+1, text)
	}
	compilation := compileMemberResolutionSource(t, source)
	if compilation.Error != nil {
		t.Fatal(compilation.Error)
	}
	owners := paragraphBindingDeclarations(t, compilation, "Tools", 2)
	members := paragraphBindingDeclarations(t, compilation, "Speak", 2)
	aliases := paragraphBindingDeclarations(t, compilation, "Display", 2)
	if owners[0].record == owners[1].record || members[0].record == members[1].record {
		t.Fatal("sibling dictionary or member declarations share a canonical record")
	}
	for index := range owners {
		owner, member := owners[index].record, members[index].record
		if owner.Value == nil || owner.Value.Iterative == nil || owner.Value.Iterative.Dictionary == nil || member.Value == nil || member.Value.Lexeme == nil {
			t.Fatal("dictionary declaration lost its typed owner or member")
		}
		dictionary := owner.Value.Iterative.Dictionary
		if dictionary.Values["Speak"] != member.Value.Lexeme {
			t.Errorf("scope %d dictionary namespace points to another owner's member", index+1)
		}
		canonical, err := canonicalRecord(aliases[index].record, compilation.Semantics.Aliases)
		if err != nil || canonical != member {
			t.Errorf("scope %d member alias lost its canonical owner: %v", index+1, err)
		}
	}
	paragraphBindingUses(t, compilation, "Displays", []*record{members[0].record, members[1].record})
	for _, name := range []string{"Tools", "Tools.Speak", "Display"} {
		if compilation.Memory[name] != nil {
			t.Errorf("private %s leaked into root memory", name)
		}
	}
}
