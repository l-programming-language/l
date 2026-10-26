package Lyger

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type memberResolutionSource struct {
	main, library string
}

func memberResolutionFixture(t *testing.T) memberResolutionSource {
	t.Helper()
	read := func(name string) string {
		content, err := os.ReadFile(filepath.Join("test", name))
		if err != nil {
			t.Fatal(err)
		}
		return string(content)
	}
	return memberResolutionSource{main: read("member-resolution.l"), library: read("local-words.l")}
}

func compileMemberResolutionSource(t *testing.T, source memberResolutionSource) *Compilation {
	t.Helper()
	directory := t.TempDir()
	files := map[string]string{"main.l": source.main}
	if source.library != "" {
		files["local-words.l"] = source.library
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	config := canonicalTestOptions()
	compilation := NewLygerService(config).NewCompilation([]byte(source.main), filepath.Join(directory, "main.l"))
	compilation.Make()
	if compilation.Error == nil {
		assertFrontendResources(t, compilation)
	}
	if !bytes.Equal(compilation.MainBytes, []byte(source.main)) {
		t.Fatal("member resolution changed original source bytes")
	}
	return compilation
}

func requireMemberResolutionProgram(t *testing.T, compilation *Compilation) {
	t.Helper()
	if compilation.Error != nil {
		t.Fatalf("compile member reference: %v", compilation.Error)
	}
	if compilation.Program == nil || len(compilation.Program.Books) != 1 || len(compilation.ProcStatements) != 0 {
		t.Fatal("member resolution did not retain one complete application")
	}
}

func memberResolutionInstructionValue(t *testing.T, value *instructionValue, owner *record, name, kind string) {
	t.Helper()
	if value == nil || value.Kind.String() != "member" || value.Member.String() != name || value.Type.String() != kind || !value.Runtime || value.Owner == nil || value.Owner.Record != owner {
		t.Fatalf("member %s lost its canonical owner or runtime type: %+v", name, value)
	}
}

func TestNativeMemberOwnersResolveFromSource(t *testing.T) {
	fixture := memberResolutionFixture(t)
	baseline := compileMemberResolutionSource(t, fixture)
	requireMemberResolutionProgram(t, baseline)
	for _, tc := range []struct {
		name, aliases, outputOwner, errorOwner string
	}{
		{name: "named output", outputOwner: "HelloWorldApp's"},
		{name: "named error", errorOwner: "HelloWorldApp's"},
		{name: "curly named members", outputOwner: "HelloWorldApp’s", errorOwner: "HelloWorldApp’s"},
		{name: "app alias", aliases: "Use HelloWorldApp as \"Worker\".\n", outputOwner: "Worker's", errorOwner: "Worker's"},
		{name: "app alias chain", aliases: "Use HelloWorldApp as \"Worker\".\nUse Worker as \"WorkerAlias\".\n", outputOwner: "WorkerAlias's", errorOwner: "WorkerAlias's"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := fixture
			source.main = strings.Replace(source.main, "Execute it.", tc.aliases+"Execute it.", 1)
			if tc.outputOwner != "" {
				source.main = strings.Replace(source.main, "it's Output", tc.outputOwner+" Output", 1)
			}
			if tc.errorOwner != "" {
				source.main = strings.Replace(source.main, "the Error", tc.errorOwner+" Error", 1)
			}
			compilation := compileMemberResolutionSource(t, source)
			requireMemberResolutionProgram(t, compilation)
			if !bytes.Equal(compilation.ASMBytes, baseline.ASMBytes) {
				t.Fatal("equivalent outer member references changed the application body")
			}
			app := compilation.Memory["HelloWorldApp"]
			output := compilation.Memory["y"].Value
			if output.Runtime == nil || output.Runtime.OwnerID != "helloworldapp" || output.Runtime.Member.String() != "output" || output.Type.String() != "Primitive" || output.Primitive.Type.String() != "Text" {
				t.Fatal("Output assignment lost its canonical deferred Text source")
			}
			var checked bool
			for _, instruction := range compilation.Semantics.Instructions {
				if instruction.Kind.String() == "if" && instruction.Condition.Operator == "has" {
					memberResolutionInstructionValue(t, instruction.Then[0].Value, app, "error", "Error")
					memberResolutionInstructionValue(t, instruction.Else[0].Value, app, "output", "Text")
					checked = true
				}
			}
			if !checked {
				t.Fatal("runtime error/output branch disappeared")
			}
			for _, statement := range compilation.Statements {
				for index, tok := range statement.tokens {
					if index == 0 || (string(tok.rawBytes) != "Output" && string(tok.rawBytes) != "Error") {
						continue
					}
					previous := string(statement.tokens[index-1].rawBytes)
					if previous != tc.outputOwner && previous != tc.errorOwner {
						continue
					}
					if tok.mem.Type.String() != "Lexeme" || tok.mem.Lexeme == nil || tok.mem.Lexeme.POS != "noun" || !strings.HasPrefix(tok.mem.Lexeme.Type, "functional.") {
						t.Fatal("runtime member resolution replaced its lexical identity")
					}
					if !strings.Contains(string(statement.source.Raw), previous+" "+string(tok.rawBytes)) || filepath.Base(statement.source.Span.Path) != "main.l" {
						t.Fatal("member reference lost its original spelling or source location")
					}
				}
			}
		})
	}
}

func TestRuntimeAssignmentUsesItsActualRHSOwner(t *testing.T) {
	source := memberResolutionFixture(t)
	source.main = strings.Replace(source.main, "If it has an Error", "\"unrelated\" is `a different paragraph subject`.\nIf HelloWorldApp has an Error", 1)
	source.main = strings.Replace(source.main, "it's Output", "HelloWorldApp's Output", 1)
	compilation := compileMemberResolutionSource(t, source)
	requireMemberResolutionProgram(t, compilation)
	output := compilation.Memory["y"].Value.Runtime
	if output == nil || output.OwnerID != "helloworldapp" {
		t.Fatal("assignment used the incoming paragraph subject instead of its explicit RHS owner")
	}
	var found bool
	for _, scoped := range compilation.Scope.Ordered {
		if scoped.Statement == output.Source {
			if scoped.IncomingSubjectID != "unrelated" {
				t.Fatal("test did not retain the distinct incoming paragraph subject")
			}
			found = true
		}
	}
	if !found {
		t.Fatal("runtime assignment lost its originating statement")
	}
}

func TestDictionaryMemberAliasesPreserveNamespace(t *testing.T) {
	fixture := memberResolutionFixture(t)
	baseline := compileMemberResolutionSource(t, fixture)
	requireMemberResolutionProgram(t, baseline)
	for _, chain := range []bool{false, true} {
		name, aliases, owner := "single alias", "Use CoreLib as \"CopyLib\".\n", "CopyLib"
		if chain {
			name, aliases, owner = "alias chain", aliases+"Use CopyLib as \"Mirror\".\n", "Mirror"
		}
		t.Run(name, func(t *testing.T) {
			source := fixture
			source.main = strings.Replace(source.main, "\"x\" is", aliases+"Use "+owner+"'s Print as \"Display\".\n\"x\" is", 1)
			source.main = strings.Replace(source.main, "The App Prints x.", "The App Displays x.", 1)
			compilation := compileMemberResolutionSource(t, source)
			requireMemberResolutionProgram(t, compilation)
			if !bytes.Equal(compilation.ASMBytes, baseline.ASMBytes) {
				t.Fatal("aliasing an existing member changed its instantiated application body")
			}
			library := compilation.Memory["CoreLib"].Value.Iterative.Dictionary
			if library.ID != "CoreLib" || compilation.Memory[owner].Value.Iterative.Dictionary != library || library.Values["Print"] != compilation.Memory["CoreLib.Print"].Value.Lexeme {
				t.Fatal("dictionary alias renamed, copied, or detached the canonical member namespace")
			}
			if compilation.Memory[owner+".Print"] != nil || distributionCanonical(t, compilation, compilation.Memory["Display"]) != distributionCanonical(t, compilation, compilation.Memory["CoreLib.Print"]) {
				t.Fatal("member alias did not retain the canonical definition")
			}
		})
	}
}

func TestMemberDeclarationsUseScopedCanonicalOwners(t *testing.T) {
	fixture := memberResolutionFixture(t)
	for _, tc := range []struct{ name, aliases, owner string }{
		{"explicit owner control", "", "CoreLib's"},
		{"paragraph pronoun", "", "it's"},
		{"alias chain owner", "Use CoreLib as \"CopyLib\".\nUse CopyLib as \"Mirror\".\n", "Mirror's"},
		{"alias chain pronoun", "Use CoreLib as \"CopyLib\".\nUse CopyLib as \"Mirror\".\n", "it's"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := fixture
			source.library = strings.Replace(source.library, "Use ImportedPrint as CoreLib's", tc.aliases+"Show CoreLib's Print.\nUse ImportedPrint as "+tc.owner, 1)
			source.library += "Show it.\n"
			compilation := compileMemberResolutionSource(t, source)
			requireMemberResolutionProgram(t, compilation)
			member := compilation.Memory["CoreLib.Print"]
			if member == nil || compilation.Memory["it.Print"] != nil || compilation.Memory["Mirror.Print"] != nil {
				t.Fatal("declaration was registered under its spelling instead of the canonical owner")
			}
			library := compilation.Memory["CoreLib"].Value.Iterative.Dictionary
			value, err := member.Value.value()
			if err != nil {
				t.Fatal(err)
			}
			if library.ID != "CoreLib" || library.Values["Print"] != member.Value.Lexeme || library.Types["Print"] != value.Type.String() {
				t.Fatal("member declaration did not populate its original dictionary")
			}
			declaration := compilation.Semantics.Declarations[member]
			if declaration == nil || filepath.Base(declaration.source.Span.Path) != "local-words.l" || !strings.Contains(string(declaration.source.Raw), "Use ImportedPrint as "+tc.owner) {
				t.Fatal("canonical registration lost imported declaration provenance")
			}
			var shows []*instruction
			for _, instruction := range compilation.Semantics.Instructions {
				if instruction.Kind.String() == "show" && filepath.Base(instruction.Source.Path) == "local-words.l" {
					shows = append(shows, instruction)
				}
			}
			if len(shows) != 2 || shows[0].Source.Start.Line >= declaration.source.Span.Start.Line || shows[1].Source.Start.Line <= declaration.source.Span.Start.Line || shows[0].Order >= shows[1].Order {
				t.Fatal("dependency retry changed source order or dropped the forward member read")
			}
			if shows[1].Value == nil || distributionCanonical(t, compilation, shows[1].Value.Record) != distributionCanonical(t, compilation, member) {
				t.Fatal("following pronoun did not resolve the actual alias-owned declaration")
			}
			if shows[0].Value == nil || shows[0].Value.Kind.String() != "symbol" || distributionCanonical(t, compilation, shows[0].Value.Record) != distributionCanonical(t, compilation, member) || shows[0].Value.Runtime {
				t.Fatal("static qualified member read became a runtime member access")
			}
		})
	}
}

func TestConditionalMembersRetainDistinctApplicationOwners(t *testing.T) {
	fixture := memberResolutionFixture(t)
	for _, tc := range []struct{ name, condition, errorOwner, outputOwner string }{
		{"named condition and pronoun RHS", "If HelloWorldApp has an Error, exit with the Error, else define \"y\" as it's Output.", "HelloWorldApp", "SecondApp"},
		{"possessive pronoun Error and named RHS", "If it has an Error, exit with it's Error, else define \"y\" as HelloWorldApp's Output.", "SecondApp", "HelloWorldApp"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := fixture
			source.main = strings.Replace(source.main, "Execute it.", "Create \"SecondApp\" as a New Basic App.\nThe App Prints x.\nExecute it.", 1)
			source.main = strings.Replace(source.main, "If it has an Error, exit with the Error, else define \"y\" as it's Output.", tc.condition, 1)
			compilation := compileMemberResolutionSource(t, source)
			if compilation.Error != nil {
				t.Fatal(compilation.Error)
			}
			if compilation.Program == nil || len(compilation.Program.Books) != 2 || len(compilation.ProcStatements) != 0 {
				t.Fatal("distinct owner check did not construct both declared applications")
			}
			output := compilation.Memory["y"].Value.Runtime
			if output == nil || output.OwnerID != strings.ToLower(tc.outputOwner) || output.Member.String() != "output" {
				t.Fatal("conditional assignment took its owner from another clause instead of the exact RHS")
			}
			for _, instruction := range compilation.Semantics.Instructions {
				if instruction.Kind.String() == "if" && instruction.Condition.Operator == "has" {
					if instruction.Condition.Left.Record != compilation.Memory[tc.errorOwner] || len(instruction.Then) != 1 || len(instruction.Else) != 1 {
						t.Fatal("conditional lost its owner or branch structure")
					}
					memberResolutionInstructionValue(t, instruction.Then[0].Value, compilation.Memory[tc.errorOwner], "error", "Error")
					memberResolutionInstructionValue(t, instruction.Else[0].Value, compilation.Memory[tc.outputOwner], "output", "Text")
					return
				}
			}
			t.Fatal("distinct-owner conditional disappeared")
		})
	}
}

func TestNamedNativeMembersWaitForTheirOwner(t *testing.T) {
	source := memberResolutionFixture(t)
	source.main = "Show HelloWorldApp's Output.\n" + source.main
	compilation := compileMemberResolutionSource(t, source)
	requireMemberResolutionProgram(t, compilation)
	if len(compilation.Semantics.Instructions) == 0 {
		t.Fatal("forward native member read disappeared")
	}
	first := compilation.Semantics.Instructions[0]
	if first.Kind.String() != "show" || first.Source.Start.Line != 1 || filepath.Base(first.Source.Path) != "main.l" {
		t.Fatal("forward native member read changed source order or provenance")
	}
	memberResolutionInstructionValue(t, first.Value, compilation.Memory["HelloWorldApp"], "output", "Text")
}

func TestNativeTypeMemberDeclarationPreservesNamespace(t *testing.T) {
	source := memberResolutionFixture(t)
	baseline := compileMemberResolutionSource(t, source)
	requireMemberResolutionProgram(t, baseline)
	// This test owns one unused member definition, not a copy of Function.Return.
	declaration := "Define any App's \"Inspect|-s|-ed|-ing|-er\" to `<DEF>BOD^E{|Statement: ~inspect(<@Statement|Text@>)~|}<DEF>`.\n"
	source.main = declaration + source.main
	compilation := compileMemberResolutionSource(t, source)
	requireMemberResolutionProgram(t, compilation)
	member := compilation.Memory["App.Inspect"]
	if member == nil || compilation.Memory["App"] != nil || compilation.Memory["Inspect"] != nil {
		t.Fatal("native type member became an instance or an unqualified global")
	}
	origin := compilation.Semantics.Declarations[member]
	if origin == nil || filepath.Base(origin.source.Span.Path) != "main.l" || !strings.Contains(string(origin.source.Raw), "App's \"Inspect") {
		t.Fatal("native type member lost declaration provenance")
	}
	if !bytes.Equal(compilation.ASMBytes, baseline.ASMBytes) {
		t.Fatal("registering an unused native type member changed the application body")
	}
}

func TestNativeMemberBindingDiscardsStaleGlobalValue(t *testing.T) {
	compilation := compileMemberResolutionSource(t, memberResolutionFixture(t))
	requireMemberResolutionProgram(t, compilation)
	statement, err := newLStatement([]byte("Show HelloWorldApp's Output"), compilation.Gram)
	if err != nil {
		t.Fatal(err)
	}
	if err := statement.tokenize(); err != nil {
		t.Fatal(err)
	}
	for _, tok := range statement.tokens {
		tok.scope = compilation.Scope.Root
	}
	global := &record{ID: "Output", Value: &entry{Type: storagePrimitive, Primitive: &primitive{Type: valueText, Text: "`unrelated global`"}}}
	compilation.Memory[global.ID] = global
	member := statement.tokens[len(statement.tokens)-1]
	state, bindErr := compilation.resolver().bindToken(member)
	if bindErr != nil || state != lookupFound || member.binding.Record != global {
		t.Fatal("could not arrange a stale unqualified global binding")
	}
	ready, err := compilation.resolver().bindStatementMembers(statement)
	if err != nil || !ready {
		t.Fatalf("native owner/member did not resolve: %v", err)
	}
	if member.binding != nil || member.bindingError != nil || member.mem.Runtime == nil || member.mem.Runtime.OwnerID != "helloworldapp" || member.mem.Runtime.Member.String() != "output" || member.mem.Type.String() != "Lexeme" || !member.hasRole("noun", "functional.deflective") {
		t.Fatal("native member retained the stale global value or lost its lexical role")
	}
	if global.Value.Primitive.Text != "`unrelated global`" || global.Value.Runtime != nil {
		t.Fatal("native resolution mutated the unrelated global record")
	}
}

func TestMemberOwnerBindingRejectsInvalidAliasesAndAmbiguity(t *testing.T) {
	for _, kind := range []string{"missing target", "alias cycle", "ambiguous reference"} {
		t.Run(kind, func(t *testing.T) {
			compilation := compileMemberResolutionSource(t, memberResolutionFixture(t))
			requireMemberResolutionProgram(t, compilation)
			app := compilation.Memory["HelloWorldApp"]
			alias := &record{ID: "Worker", Value: app.Value}
			compilation.Memory[alias.ID] = alias
			compilation.Semantics.Aliases[alias] = app
			want := kind
			switch kind {
			case "missing target":
				compilation.Semantics.Aliases[alias] = nil
				want = "no target"
			case "alias cycle":
				compilation.Semantics.Aliases[app] = alias
			case "ambiguous reference":
				compilation.Memory["worker"] = &record{ID: "worker", Value: app.Value}
			}
			statement, err := newLStatement([]byte("Show Worker's Output"), compilation.Gram)
			if err != nil {
				t.Fatal(err)
			}
			if err := statement.tokenize(); err != nil {
				t.Fatal(err)
			}
			for _, tok := range statement.tokens {
				tok.scope = compilation.Scope.Root
			}
			if _, err = compilation.resolver().bindStatementMembers(statement); err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("invalid owner resolution returned %v", err)
			}
		})
	}
}

func TestNativeMembersWaitForTheirOwnerDeclaration(t *testing.T) {
	source := memberResolutionFixture(t)
	source.main = strings.Replace(source.main, "New Basic App.", "New Later App.\nShow it's Output.", 1)
	source.main += "\n\"Later\" describes a Generic App with a New Main Service using any Process.\n"
	compilation := compileMemberResolutionSource(t, source)
	requireMemberResolutionProgram(t, compilation)
	var found bool
	for _, instruction := range compilation.Semantics.Instructions {
		if instruction.Kind.String() == "show" && instruction.Value != nil && instruction.Value.Kind.String() == "member" {
			memberResolutionInstructionValue(t, instruction.Value, compilation.Memory["HelloWorldApp"], "output", "Text")
			found = true
		}
	}
	if !found {
		t.Fatal("native member read disappeared while its owner waited on a descriptor")
	}
}

func TestMemberResolutionRejectsUnrelatedNamespacesAndScopes(t *testing.T) {
	fixture := memberResolutionFixture(t)
	requireMemberResolutionProgram(t, compileMemberResolutionSource(t, fixture))
	for _, tc := range []struct{ name, main, library, want string }{
		{"global member is not a dictionary member", strings.Replace(fixture.main, "\"x\" is", "Create \"Empty\" as a Dictionary.\nUse Empty's Print as \"Invalid\".\n\"x\" is", 1), fixture.library, "Empty's Print"},
		{"text is not a runtime application", fixture.main + "\nShow x's Output.\n", fixture.library, "Output"},
		{"bare App type is not a runtime application", fixture.main + "\nShow App's Output.\n", fixture.library, "App"},
		{"named owner must exist", fixture.main + "\nShow MissingApp's Output.\n", fixture.library, "MissingApp"},
		{"named member must exist", fixture.main + "\nShow HelloWorldApp's MissingField.\n", fixture.library, "MissingField"},
		{"pronoun cannot cross a paragraph", fixture.main, strings.Replace(fixture.library, "Use ImportedPrint as CoreLib's", "\nUse ImportedPrint as it's", 1), "source paragraph"},
		{"pronoun cannot cross an imported file", "\"outside\" is `caller subject`.\n" + fixture.main, "Show it.\n" + fixture.library, "local-words.l:1:1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := fixture
			source.main, source.library = tc.main, tc.library
			compilation := compileMemberResolutionSource(t, source)
			if compilation.Error == nil || !strings.Contains(compilation.Error.Error(), tc.want) {
				t.Fatalf("invalid owner/member was accepted or lost its diagnostic: %v", compilation.Error)
			}
			if compilation.Program != nil || len(compilation.ASMBytes) != 0 {
				t.Fatal("invalid member resolution published application output")
			}
		})
	}
}
