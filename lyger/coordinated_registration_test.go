package Lyger

import (
	"bytes"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestCoordinatedRegistrationAliasAndScalar(t *testing.T) {
	for _, tc := range []struct {
		name, sentence     string
		local, aliasPublic bool
		scalarPublic       bool
	}{
		{name: "root", sentence: `Use Worker as "Primary" and define "count" as 1.`},
		{name: "public alias", sentence: `Publicly use Worker as "Primary" and define "count" as 1.`, local: true, aliasPublic: true},
		{name: "public scalar", sentence: `Use Worker as "Primary" and publicly define "count" as 1.`, local: true, scalarPublic: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			prefix := "\"Worker\" is an App.\n"
			if tc.local {
				prefix += "\n    "
			}
			compilation := paragraphPublicationSource(t, prefix+tc.sentence+"\n")
			if compilation.Error != nil {
				t.Fatal(compilation.Error)
			}
			alias := paragraphBindingDeclarations(t, compilation, "Primary", 1)[0]
			count := paragraphBindingDeclarations(t, compilation, "count", 1)[0]
			worker := compilation.Memory["Worker"]
			canonical, err := canonicalRecord(alias.record, compilation.Semantics.Aliases)
			if err != nil || canonical != worker || alias.record.Value.Applicative == nil || alias.record.Value.Applicative.App != worker.Value.Applicative.App {
				t.Fatalf("alias lost its actual source App: %v", err)
			}
			value, err := count.record.Value.value()
			if err != nil || value.Type != valueNumber || value.Data != float64(1) || compilation.Semantics.Aliases[count.record] != nil {
				t.Fatalf("scalar sibling became an alias or lost its Number value: %+v, %v", value, err)
			}
			if alias.statement != count.statement || len(alias.statement.records) != 2 || alias.statement.records[0] != alias.record || alias.statement.records[1] != count.record {
				t.Fatal("coordinated declarations lost their original statement or source order")
			}
			for _, declaration := range []struct {
				name   string
				record *record
				public bool
			}{{"Primary", alias.record, tc.aliasPublic}, {"count", count.record, tc.scalarPublic}} {
				wantRoot := !tc.local || declaration.public
				if (compilation.Memory[declaration.name] == declaration.record) != wantRoot {
					t.Fatalf("%s publication ignored its own modifier", declaration.name)
				}
				if compilation.Semantics.Declarations[declaration.record] != alias.statement || !bytes.Contains(alias.statement.source.Raw, []byte(tc.sentence[:len(tc.sentence)-1])) {
					t.Fatal("publication replaced original declaration provenance")
				}
			}
		})
	}
}

func TestCoordinatedRegistrationAliasesKeepTheirOwnSources(t *testing.T) {
	for _, second := range []string{"Worker", "Other"} {
		t.Run(second, func(t *testing.T) {
			compilation := paragraphPublicationSource(t, "\"Worker\" is an App.\n\"Other\" is an App.\n"+
				"Use Worker as \"Primary\" and use "+second+" as \"Secondary\".\n")
			if compilation.Error != nil {
				t.Fatal(compilation.Error)
			}
			for _, pair := range [][2]string{{"Primary", "Worker"}, {"Secondary", second}} {
				alias := compilation.Memory[pair[0]]
				canonical, err := canonicalRecord(alias, compilation.Semantics.Aliases)
				if alias == nil || err != nil || canonical != compilation.Memory[pair[1]] {
					t.Errorf("%s must alias its own operand %s: record=%v error=%v", pair[0], pair[1], alias, err)
				}
			}
			if len(compilation.Semantics.Aliases) != 2 || len(compilation.Semantics.Apps) != 2 {
				t.Fatal("alias registration omitted a binding or duplicated an application")
			}
		})
	}
}

func TestCoordinatedRegistrationDefinitionLiteralOwnership(t *testing.T) {
	const first = "`<DEF>E{|Statement: ~first()~|}<DEF>`"
	const second = "`<DEF>E{|Statement: ~second()~|}<DEF>`"
	compilation := paragraphPublicationSource(t, "Define \"Alpha\" as "+first+" and define \"Beta\" as "+second+".\n"+
		"Use Beta as \"Alias\".\n")
	if compilation.Error != nil {
		t.Fatal(compilation.Error)
	}
	alpha := paragraphBindingDeclarations(t, compilation, "Alpha", 1)[0]
	beta := paragraphBindingDeclarations(t, compilation, "Beta", 1)[0]
	if alpha.record == beta.record || alpha.statement != beta.statement {
		t.Fatal("separate definitions must share source provenance without sharing identity")
	}
	canonical, err := canonicalRecord(compilation.Memory["Alias"], compilation.Semantics.Aliases)
	if err != nil || canonical != beta.record {
		t.Fatalf("definition alias lost Beta identity: %v", err)
	}
	for _, tc := range []struct{ name, literal string }{{"Alpha", first}, {"Beta", second}, {"Alias", second}} {
		t.Run(tc.name, func(t *testing.T) {
			template, err := compilation.recordTemplate(compilation.Memory[tc.name])
			if err != nil || template == nil {
				t.Fatalf("definition template = %v, %v", template, err)
			}
			if string(template.Original) != tc.literal || template.Span != beta.statement.source.Span {
				t.Fatalf("%s chose a sibling's literal or lost source span: %s", tc.name, template.Original)
			}
		})
	}
}

func TestCoordinatedRegistrationInvalidDescriptorIsAtomic(t *testing.T) {
	const source = `"Shape" is a New Adjective describing any App or Service`
	statement, err := paragraphGrammarStatement(t, source, nil)
	if err != nil {
		t.Fatal(err)
	}
	statement.source = &SourceStatement{Raw: []byte(source), Span: SourceSpan{Path: "descriptor-batch.l", Paragraph: 1, Start: SourcePosition{Line: 1, Column: 1}}}
	compilation := &Compilation{Memory: make(map[string]*record), Semantics: newSemanticProgram()}
	compilation.Scope, err = planSourceScopes([]*lStatement{statement}, compilation.Memory, nil)
	if err != nil {
		t.Fatal(err)
	}
	var target, declaration *token
	for _, tok := range statement.tokens {
		if string(tok.rawBytes) == "Service" {
			target = tok
		}
		if tok.isDeclare() {
			declaration = tok
		}
	}
	if target == nil || declaration == nil {
		t.Fatal("test requires the original second target and declaration tokens")
	}
	// Keep the second target's checked storage valid, but remove its nominal
	// identity after grammar parsing. Failure must occur in target preparation,
	// after the first target was accepted, rather than in literal scanning.
	original := target.rawBytes
	target.rawBytes = []byte("UnresolvedTarget")
	if err := compilation.registerStatement(statement); err == nil || !strings.Contains(err.Error(), "unresolved nominal target") {
		t.Fatalf("descriptor target validation error = %v", err)
	}
	if len(compilation.Memory) != 0 || len(compilation.Semantics.Descriptors) != 0 || len(compilation.Semantics.Subjects) != 0 || len(compilation.Semantics.Declarations) != 0 || len(statement.records) != 0 || compilation.Scope.declarations[declaration].Record != nil {
		t.Fatal("failed descriptor batch published a partial record, targets, or source relationships")
	}
	// Retrying the same source tokens must produce one complete publication.
	target.rawBytes = original
	if err := compilation.registerStatement(statement); err != nil {
		t.Fatal(err)
	}
	record := compilation.Memory["Shape"]
	descriptor := compilation.Semantics.Descriptors[record]
	if record == nil || descriptor == nil || len(descriptor.Targets) != 2 || len(statement.records) != 1 || compilation.Semantics.Declarations[record] != statement || compilation.Semantics.Subjects[declaration] != record {
		t.Fatal("retry lost descriptor targets, canonical identity, or original provenance")
	}
}

func TestCoordinatedRegistrationAliasOutputAndShowInstructions(t *testing.T) {
	source := paragraphSource(t)
	source.main = paragraphGlobals + "Create \"Worker\" as a New Basic App.\nThe App Prints x1.\n" +
		"Run Worker.\nUse Worker as \"Primary\" and define \"snapshot\" as Worker's Output.\nShow snapshot.\n"
	compilation := compileMemberResolutionSource(t, source)
	if compilation.Error != nil {
		t.Fatal(compilation.Error)
	}
	worker, snapshot := compilation.Memory["Worker"], compilation.Memory["snapshot"]
	canonical, err := canonicalRecord(compilation.Memory["Primary"], compilation.Semantics.Aliases)
	if err != nil || canonical != worker || snapshot == nil || snapshot.Value.Runtime == nil || snapshot.Value.Runtime.Owner != worker {
		t.Fatalf("coordinated alias/Output declaration lost separate identities: %v", err)
	}
	var kinds []string
	for _, instruction := range compilation.Semantics.Instructions {
		kinds = append(kinds, instruction.Kind.String())
		if instruction.Kind == instructionAssign {
			if instruction.Target == nil || instruction.Target.Record != snapshot || instruction.Value == nil || instruction.Value.Owner == nil || instruction.Value.Owner.Record != worker || instruction.Statement != compilation.Semantics.Declarations[snapshot] {
				t.Fatal("Output assignment lost its own target, operand, or source statement")
			}
		}
		if instruction.Kind == instructionShow && (instruction.Value == nil || instruction.Value.Record != snapshot) {
			t.Fatal("Show stopped reading the stored snapshot")
		}
	}
	if !reflect.DeepEqual(kinds, []string{"run", "assign", "show"}) {
		t.Fatalf("alias registration swallowed or duplicated an executable sibling: %v", kinds)
	}
	result := &CompilationResult{Compilation: compilation}
	if _, err := result.OrchestratorUnit("main", executableTestCompanions(result)); err != nil {
		t.Fatal(err)
	}
}

// Prepare an additional real grammar statement against a published environment,
// stopping before publication so failure can be injected at that API boundary.
func coordinatedRegistrationPrepared(t *testing.T, compilation *Compilation, source string) (*lStatement, statementRegistration) {
	t.Helper()
	statement, err := paragraphGrammarStatement(t, source, compilation.Memory)
	if err != nil {
		t.Fatal(err)
	}
	previous := compilation.Statements[len(compilation.Statements)-1].source
	statement.source = &SourceStatement{Raw: []byte(source), Span: SourceSpan{Path: previous.Span.Path, Paragraph: previous.Span.Paragraph, Start: SourcePosition{Line: 7, Column: 1}}, occurrence: previous.occurrence}
	statements := append(append([]*lStatement(nil), compilation.Statements...), statement)
	compilation.Scope, err = planSourceScopes(statements, compilation.Memory, compilation.SourceOccurrences)
	if err != nil {
		t.Fatal(err)
	}
	sites, err := statement.sGram.registrationSites()
	if err != nil {
		t.Fatal(err)
	}
	prepared, err := compilation.prepareRegistration(statement, sites)
	if err != nil {
		t.Fatal(err)
	}
	return statement, prepared
}

func TestCoordinatedRegistrationLaterFailureRollsBackEarlierMetadata(t *testing.T) {
	for _, first := range []string{"alias", "descriptor"} {
		for _, failure := range []string{"invalid value", "duplicate"} {
			t.Run(first+"/"+failure, func(t *testing.T) {
				compilation := paragraphPublicationSource(t, "\"Worker\" is an App.\n\"later\" is 9.\n")
				if compilation.Error != nil {
					t.Fatal(compilation.Error)
				}
				source := `Use Worker as "Earlier" and publicly define "later" as 2`
				if first == "descriptor" {
					source = `"Earlier" is a New Adjective describing any App and publicly define "later" as 2`
				}
				prior := compilation.Memory["later"]
				priorSource := compilation.Semantics.Declarations[prior]
				statement, prepared := coordinatedRegistrationPrepared(t, compilation, source)
				if len(prepared.declarations) != 2 {
					t.Fatalf("prepared declarations = %d; want original two grammar sites", len(prepared.declarations))
				}
				if first == "descriptor" && len(prepared.descriptors) != 1 {
					t.Fatal("test did not prepare the descriptor metadata")
				}
				if failure == "invalid value" {
					prepared.declarations[1].Record.Value = &entry{Type: storagePrimitive}
				}
				err := compilation.publishRegistration(statement, prepared)
				if err == nil || (failure == "duplicate" && !strings.Contains(err.Error(), "duplicate declaration later")) {
					t.Fatalf("invalid later declaration error = %v", err)
				}
				if len(compilation.Memory) != 2 || compilation.Memory["Earlier"] != nil || compilation.Memory["later"] != prior || compilation.Semantics.Declarations[prior] != priorSource {
					t.Fatal("rejected batch changed existing values or published its earlier binding")
				}
				subjects, declarations, descriptors := testProgramSemanticCounts(compilation)
				if len(compilation.Semantics.Aliases) != 0 || descriptors != 0 || len(compilation.Semantics.Apps) != 1 || subjects != 2 || declarations != 2 || len(statement.records) != 0 {
					t.Fatal("rejected batch changed alias, descriptor, App, or provenance relationships")
				}
				for _, declaration := range prepared.declarations {
					if compilation.Scope.declarations[declaration.Token].Record != nil || declaration.Record.storageScope != nil {
						t.Fatal("rejected batch resolved a reservation or mutated a prepared record")
					}
				}
			})
		}
	}
}

func TestCoordinatedRegistrationFailedRefinementPreservesPublishedDescriptor(t *testing.T) {
	compilation := paragraphPublicationSource(t, "\"Shape\" is a New Adjective describing any App.\nIt describes any App.\n")
	if compilation.Error != nil {
		t.Fatal(compilation.Error)
	}
	record := compilation.Memory["Shape"]
	original := compilation.Semantics.Descriptors[record]
	if original == nil || len(original.Rules) != 1 {
		t.Fatal("test requires an already published descriptor refinement")
	}
	targets := make(map[nominalIdentity]bool)
	for target, enabled := range original.Targets {
		targets[target] = enabled
	}
	rules := append([]*descriptorRule(nil), original.Rules...)
	declarationSource := compilation.Semantics.Declarations[record]
	statement, prepared := coordinatedRegistrationPrepared(t, compilation, `It describes any App and define "later" as 2`)
	if len(prepared.declarations) != 1 || len(prepared.descriptors) != 1 || prepared.descriptors[0] == original || len(prepared.descriptors[0].Rules) != len(rules)+1 {
		t.Fatal("refinement was not staged separately from its published descriptor")
	}
	validValue := prepared.declarations[0].Record.Value
	prepared.declarations[0].Record.Value = &entry{Type: storagePrimitive}
	if err := compilation.publishRegistration(statement, prepared); err == nil {
		t.Fatal("invalid later declaration was accepted")
	}
	if compilation.Semantics.Descriptors[record] != original || !reflect.DeepEqual(original.Targets, targets) || !reflect.DeepEqual(original.Rules, rules) || compilation.Semantics.Declarations[record] != declarationSource || compilation.Memory["later"] != nil || len(statement.records) != 0 {
		t.Fatal("failed batch mutated an existing descriptor or published a sibling")
	}
	prepared.declarations[0].Record.Value = validValue
	if err := compilation.publishRegistration(statement, prepared); err != nil {
		t.Fatal(err)
	}
	current := compilation.Semantics.Descriptors[record]
	if current == original || current.Record != record || len(current.Rules) != len(rules)+1 || current.Rules[len(rules)].Source != statement || current.Rules[0] != rules[0] || !reflect.DeepEqual(original.Rules, rules) || !reflect.DeepEqual(original.Targets, targets) {
		t.Fatal("successful retry lost canonical identity, rule order, source attribution, or snapshot isolation")
	}
	if compilation.Semantics.Declarations[record] != declarationSource || len(statement.records) != 1 || compilation.Semantics.Declarations[statement.records[0]] != statement {
		t.Fatal("refinement retry replaced declaration provenance or duplicated publication")
	}
}

func TestCoordinatedRegistrationDescriptorDoesNotBorrowSiblingLiteral(t *testing.T) {
	const literal = "`<DEF>E{|Statement: ~unrelated()~|}<DEF>`"
	compilation := paragraphPublicationSource(t, "\"Unrelated\" is "+literal+" and \"Shape\" is a New Adjective describing any App.\n"+
		"It describes any App.\n")
	if compilation.Error != nil {
		t.Fatal(compilation.Error)
	}
	record := compilation.Memory["Shape"]
	descriptor := compilation.Semantics.Descriptors[record]
	if descriptor == nil || len(descriptor.Rules) != 1 || descriptor.Rules[0].Template != nil || record.Value.Lexeme.Value != "" {
		t.Fatal("descriptor absorbed the unrelated definition literal")
	}
	template, err := compilation.recordTemplate(record)
	if err != nil || template != nil {
		t.Fatalf("descriptor acquired a sibling's template: %v, %v", template, err)
	}
	unrelated, err := compilation.recordTemplate(compilation.Memory["Unrelated"])
	if err != nil || unrelated == nil || string(unrelated.Original) != literal {
		t.Fatalf("sibling definition lost its own literal: %v", err)
	}
}

func TestCoordinatedRegistrationDescriptorVisibilityAndNominalAlias(t *testing.T) {
	for _, public := range []bool{false, true} {
		name, modifier := "local", ""
		if public {
			name, modifier = "public", "Public "
		}
		t.Run(name, func(t *testing.T) {
			source := paragraphDescriptorWidget + "Use Widget as \"Handle\".\n\n    \"count\" is 1 and \"Shape\" is a New " + modifier + "Adjective describing any Handle.\nIt describes any Widget.\n"
			compilation := paragraphPublicationSource(t, source)
			if compilation.Error != nil {
				t.Fatal(compilation.Error)
			}
			shape := paragraphBindingDeclarations(t, compilation, "Shape", 1)[0]
			widget := compilation.Memory["Widget"]
			descriptor := compilation.Semantics.Descriptors[shape.record]
			if descriptor == nil || len(descriptor.Targets) != 1 || !descriptor.Targets[nominalIdentity{definition: widget}] || len(descriptor.Rules) != 1 || descriptor.Rules[0].Target != (nominalIdentity{definition: widget}) {
				t.Fatal("descriptor target/refinement did not retain the alias's canonical nominal identity")
			}
			if (compilation.Memory["Shape"] == shape.record) != public || compilation.Memory["count"] != nil {
				t.Fatal("descriptor visibility escaped its own declaration or published its sibling")
			}
			if compilation.Semantics.Declarations[shape.record] != shape.statement || descriptor.Rules[0].Source == shape.statement || !strings.Contains(string(descriptor.Rules[0].Source.source.Raw), "It describes any Widget") {
				t.Fatal("descriptor visibility/alias binding lost original declaration and refinement provenance")
			}
		})
	}
}

func TestCoordinatedRegistrationImportedRetriesKeepOccurrenceIdentity(t *testing.T) {
	const declaration = `Use origin as "Copy" and define "count" as 1`
	source := memberResolutionSource{
		main:    "\"origin\" is `outer`.\n\n    Use `local-words.l`.\nShow Copy.\nShow count.\n\n    Use `local-words.l`.\nShow Copy.\nShow count.\n",
		library: declaration + ".\n\"origin\" is `left`.\n",
	}
	compilation := compileMemberResolutionSource(t, source)
	if compilation.Error != nil {
		t.Fatal(compilation.Error)
	}
	var copies, counts []paragraphBindingDeclaration
	for _, statement := range compilation.Statements {
		if filepath.Base(statement.source.Span.Path) != "local-words.l" {
			continue
		}
		for _, tok := range statement.tokens {
			switch tok.declarationRoot() {
			case "copy":
				copies = append(copies, paragraphBindingDeclaration{compilation.Semantics.Subjects[tok], statement})
			case "count":
				counts = append(counts, paragraphBindingDeclaration{compilation.Semantics.Subjects[tok], statement})
			}
		}
	}
	if len(copies) != 2 || len(counts) != 2 {
		t.Fatalf("imported declarations = %d aliases, %d scalars", len(copies), len(counts))
	}
	for i, copy := range copies {
		canonical, err := canonicalRecord(copy.record, compilation.Semantics.Aliases)
		if err != nil || canonical == nil || canonical == compilation.Memory["origin"] {
			t.Fatalf("import retry selected an outer source: %v", err)
		}
		value, err := canonical.Value.value()
		if err != nil || value.Type != valueText || value.Data != "`left`" {
			t.Fatalf("imported alias value = %+v, %v", value, err)
		}
		count := counts[i]
		if copy.statement != count.statement || len(copy.statement.records) != 2 || copy.statement.records[0] != copy.record || copy.statement.records[1] != count.record {
			t.Fatal("retried batch lost its original declaration order")
		}
		original := copy.statement.source
		if filepath.Base(original.Span.Path) != "local-words.l" || original.Span.Start != (SourcePosition{Offset: 0, Line: 1, Column: 1}) || string(original.Raw) != declaration+".\n" {
			t.Fatalf("imported declaration lost exact source: %+v", original)
		}
		for _, rec := range []*record{copy.record, count.record} {
			found := false
			for _, token := range copy.statement.tokens {
				found = found || compilation.Semantics.Subjects[token] == rec
			}
			if !found || compilation.Semantics.Declarations[rec] != copy.statement {
				t.Fatal("retry replaced original token/statement identity")
			}
		}
	}
	if copies[0].record == copies[1].record || copies[0].statement.source.occurrence == copies[1].statement.source.occurrence || copies[0].statement.source.Span != copies[1].statement.source.Span {
		t.Fatal("repeated imports conflated canonical identity, occurrence, or original span")
	}
}

func TestCoordinatedRegistrationTrailingDescriptorConstructs(t *testing.T) {
	source := paragraphSource(t)
	source.main = paragraphGlobals +
		"\"count\" is 1 and \"Style\" is a New Adjective describing any App.\n" +
		"It describes a Generic App with a New Main Service using any Process.\n" +
		"Create \"Worker\" as a New Style App.\nThe App Prints x1.\n"
	compilation := compileMemberResolutionSource(t, source)
	if compilation.Error != nil {
		t.Fatal(compilation.Error)
	}
	style := paragraphBindingDeclarations(t, compilation, "Style", 1)[0]
	count := paragraphBindingDeclarations(t, compilation, "count", 1)[0]
	descriptor := compilation.Semantics.Descriptors[style.record]
	if descriptor == nil || descriptor.Record != style.record || len(descriptor.Targets) != 1 || len(descriptor.Rules) != 1 {
		t.Fatal("trailing descriptor lost its canonical record, target, or refinement")
	}
	if count.statement != style.statement || descriptor.Rules[0].Source == style.statement || !strings.Contains(string(descriptor.Rules[0].Source.source.Raw), "It describes") {
		t.Fatal("descriptor declaration/refinement provenance was merged or synthesized")
	}
	units, err := (&CompilationResult{Compilation: compilation}).ApplicationUnits()
	if err != nil || len(units) != 1 || !bytes.Contains(units[0].BackendAML, []byte(`(&&LPrintValue(&("Hello World!"))).l_print();`)) {
		t.Fatalf("source descriptor did not construct its App body: %v", err)
	}
}
