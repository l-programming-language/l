package Lyger

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/l-programming-language/l/lerrors"
)

func formTestStatement(t *testing.T, source string, callable bool) (*lStatement, *record) {
	t.Helper()
	word := indexedTestWord(t, "Read", `"Read|-s|-|-ing|-er"`, callable)
	return boundFormStatement(t, source, map[string]*record{"Read": word}), word
}

func boundFormStatement(t *testing.T, source string, records map[string]*record) *lStatement {
	t.Helper()
	gram, err := loadGrammar(canonicalTestOptions())
	if err != nil {
		t.Fatal(err)
	}
	statement, err := newLStatement([]byte(source), gram)
	if err != nil {
		t.Fatal(err)
	}
	if err := statement.tokenize(); err != nil {
		t.Fatal(err)
	}
	records["Target"] = &record{ID: "Target", Value: initEntry("Number", storagePrimitive)}
	records["Value"] = &record{ID: "Value", Value: initEntry("Number", storagePrimitive)}
	scope := &lexicalScope{records: records}
	for _, tok := range statement.tokens {
		tok.scope = scope
	}
	if _, err := (semanticResolver{semantics: newSemanticProgram()}).bindStatementMembers(statement); err != nil {
		t.Fatal(err)
	}
	statement.determineType()
	return statement
}

func TestSameSpellingFormsContext(t *testing.T) {
	for _, callable := range []bool{false, true} {
		for _, tc := range []struct{ source, form string }{
			{"Read Value", "root"},
			{"Target is Read Value", "passive"},
		} {
			t.Run(fmt.Sprintf("%s/callable=%t", tc.source, callable), func(t *testing.T) {
				statement, word := formTestStatement(t, tc.source, callable)
				if err := statement.loadGram(); err != nil {
					t.Fatal(err)
				}
				for _, tok := range statement.tokens {
					if tok.binding != nil && tok.binding.Record == word && tok.binding.Form != tc.form {
						t.Fatalf("%q selected %q, want %q", tok.rawBytes, tok.binding.Form, tc.form)
					}
				}
			})
		}
	}
}

func TestSameSpellingFormsPreserveIdentity(t *testing.T) {
	for _, family := range []struct{ root, suffixes string }{
		{"Read", "|-s|-|-ing|-er"}, {"Set", "|-s|-|-ting|-ter"},
		{"Split", "|-s|-|-ting|-ter"}, {"Cast", "|-s|-|-ing|-er"},
	} {
		for _, callable := range []bool{false, true} {
			for _, qualification := range []string{"", "Catalog.", "Catalog's ", "not Catalog's "} {
				t.Run(fmt.Sprintf("%s/%s/callable=%t", family.root, qualification, callable), func(t *testing.T) {
					id := family.root
					if qualification != "" {
						id = "Catalog." + family.root
					}
					word := indexedTestWord(t, id, `"`+family.root+family.suffixes+`"`, callable)
					owner := &record{ID: "Catalog", Value: initEntry("Dictionary", storageIterative), members: map[string]*record{id: word}}
					records := map[string]*record{id: word, "Catalog": owner}
					if qualification != "" {
						// The namespace-selected member keeps its definition even
						// when the caller has an identically spelled global verb.
						records[family.root] = indexedTestWord(t, family.root, `"`+family.root+family.suffixes+`"`, callable)
					}
					statement := boundFormStatement(t, "Target is "+qualification+strings.ToUpper(family.root)+" Value", records)
					if err := statement.loadGram(); err != nil {
						t.Fatal(err)
					}
					matched := 0
					for _, tok := range statement.tokens {
						if tok.binding != nil && tok.binding.Record == word {
							matched++
							if tok.binding.Form != "passive" || tok.lexeme().POS != "verb" || !bytes.Contains(tok.rawBytes, []byte(strings.ToUpper(family.root))) {
								t.Fatalf("context changed identity, source bytes or form: %+v", tok.binding)
							}
						}
					}
					if matched != 1 {
						t.Fatalf("selected %d word occurrences, want 1", matched)
					}
				})
			}
		}
	}
}

func TestSameSpellingFormsDoNotResolveHomonyms(t *testing.T) {
	first := indexedTestWord(t, "Read", `"Read|-s|-|-ing|-er"`, false)
	second := indexedTestWord(t, "Other", `"Read|-s|-|-ing|-er"`, false)
	records := map[string]*record{"first": first, "second": second}
	want := ""
	for _, indexed := range []bool{false, true} {
		scope := &lexicalScope{records: records}
		if indexed {
			scope.bindings = newRecordBindingIndex(records)
		}
		for repeat := 0; repeat < 10; repeat++ {
			_, _, err := scope.lookupLocal([]byte("Read"), posVerb)
			if !errors.Is(err, lerrors.AmbiguousReference) {
				t.Fatalf("unrelated records lost ambiguity: %v", err)
			}
			if want == "" {
				want = err.Error()
			} else if err.Error() != want {
				t.Fatalf("unstable ambiguity: %v; want %s", err, want)
			}
		}
	}
}

func TestSameSpellingFormsKeepDescriptorHomonymsSeparate(t *testing.T) {
	word := indexedTestWord(t, "Read", `"Read|-s|-|-ing|-er"`, false)
	descriptor := newLexeme("adjective")
	if err := descriptor.reLabel(`"Read"`); err != nil {
		t.Fatal(err)
	}
	modifier := &record{ID: "Read", Value: &entry{Type: storageLexeme, Lexeme: descriptor}}
	statement := boundFormStatement(t, "Target is not Read Value", map[string]*record{"verb": word, "descriptor": modifier})
	if err := statement.loadGram(); err != nil {
		t.Fatal(err)
	}
	if got := statement.tokens[3].binding; got.Record != word || got.Form != "passive" {
		t.Fatalf("passive context selected an unrelated descriptor: %+v", got)
	}
}

func TestSameSpellingFormsRequireDeclaredCandidate(t *testing.T) {
	for _, family := range []struct {
		label    string
		rejected []string
	}{
		{`"Read|-s||-ing|-er"`, []string{"Read", "Reads", "Reading"}},
		{`"Read|-s|-|-ing|-er"`, []string{"Reads", "Reading"}},
	} {
		for _, callable := range []bool{false, true} {
			for _, spelling := range family.rejected {
				t.Run(fmt.Sprintf("%s/%s/callable=%t", family.label, spelling, callable), func(t *testing.T) {
					word := indexedTestWord(t, "Read", family.label, callable)
					statement := boundFormStatement(t, "Target is "+spelling+" Value", map[string]*record{"Read": word})
					verb := statement.tokens[2]
					before, category := verb.binding, verb.tType
					if err := statement.loadGram(); err == nil {
						t.Fatal("undeclared or incorrect passive accepted")
					}
					if verb.binding != before || verb.tType != category || verb.binding.Record != word {
						t.Fatal("failed passive changed the selected identity or form")
					}
				})
			}
		}
	}
}

func TestSameSpellingSourceActionRequiresDeclaredPassive(t *testing.T) {
	for _, spelling := range []string{"Read", "Reads", "Reading"} {
		source := distributionCorePrefix + "Create \"Read|-s||-ing|-er\" for Printing any Text.\nCreate \"Worker\" as a New Basic App.\nWorker is " + spelling + " `payload`.\n"
		if _, err := distributionFrontend(t, source, nil); !errors.Is(err, lerrors.Syntax) {
			t.Errorf("%s without a passive declaration: %v", spelling, err)
		}
	}
}

func TestSameSpellingFormsConfiguredAndFollowingClause(t *testing.T) {
	statement, word := formTestStatement(t, "Target is Read Value and then Read Value", false)
	if err := statement.loadGram(); err != nil {
		t.Fatal(err)
	}
	var forms []string
	for _, tok := range statement.tokens {
		if tok.binding != nil && tok.binding.Record == word {
			forms = append(forms, tok.binding.Form)
		}
	}
	if strings.Join(forms, ",") != "passive,root" {
		t.Fatalf("following imperative inherited passive context: %v", forms)
	}
	configured, _ := formTestStatement(t, "Target is Put in Value", false)
	if err := configured.loadGram(); err != nil {
		t.Fatal(err)
	}
	put := configured.tokens[2]
	if put.binding != nil || put.tType != "verbs.expansive.passive" || !put.hasRole("verb", "action.expansive") {
		t.Fatalf("configured Put lost its native category: %+v", put)
	}
}

func TestSameSpellingFormsSpeculationIsPure(t *testing.T) {
	statement, word := formTestStatement(t, "Read Value", false)
	verb := statement.tokens[0]
	beforeBinding, beforeType := verb.binding, verb.tType
	beforeBytes := bytes.Clone(verb.rawBytes)
	clause, err := initClauseGram("clause.passive", statement.tokens, statement.gram)
	if err != nil {
		t.Fatal(err)
	}
	if err := clause.build(); err != nil {
		t.Fatal(err)
	}
	if verb.binding != beforeBinding || verb.binding.Record != word || verb.binding.Form != "root" || verb.tType != beforeType || !bytes.Equal(verb.rawBytes, beforeBytes) {
		t.Fatal("speculative clause changed its source token")
	}
	failed, _ := formTestStatement(t, "Target is Read Value unexpected", false)
	if err := failed.loadGram(); err == nil {
		t.Fatal("unconsumed trailing token accepted")
	}
	if failed.tokens[2].binding.Form != "root" {
		t.Fatal("failed whole statement committed a passive form")
	}
	continuation, _ := formTestStatement(t, "Target is Put in Value and then Put in Value unexpected", false)
	if err := continuation.loadGram(); err == nil {
		t.Fatal("failed continuation accepted an unconsumed token")
	}
	for _, tok := range continuation.tokens {
		if tok.hasRole("verb", "action.expansive") && tok.tType != "verbs.expansive.root" {
			t.Fatal("failed continuation committed a speculative passive form")
		}
	}
}

func TestSameSpellingFunctionAssignmentPriority(t *testing.T) {
	source := functionSource("Read|-s|-|-ing|-er", "\"Value\" as Text", "Text", "Return Value.")
	for _, declaration := range []string{"\"Saved\" is Read.", "Use Read as \"Alias\".\n\"Saved\" is Alias."} {
		requireFunctionGenerated(t, source+declaration+"\nSaved `copy` as \"Result\".\nShow Result.\n")
	}
}

func TestPutGrammaticalOperands(t *testing.T) {
	for _, tc := range []struct{ source, form string }{
		{"Put Value in Items", "root"},
		{"Target Puts Value in Items", "active"},
		{"Value is Put in Items", "passive"},
	} {
		t.Run(tc.form, func(t *testing.T) {
			items := &record{ID: "Items", Value: initEntry("List", storageIterative)}
			statement := boundFormStatement(t, tc.source, map[string]*record{"Items": items})
			if err := statement.loadGram(); err != nil {
				t.Fatal(err)
			}
			clause := statement.sGram.predicate.clauses[0]
			value := statement.sGram.subject
			var payload *token
			if tc.form == "passive" {
				payload = value.noun
			} else {
				payload = clause.obj.noun
			}
			if payload.binding.Record.ID != "Value" || clause.phrase.subject.noun.binding.Record != items || clause.verb.tType != "verbs.expansive."+tc.form {
				t.Fatal("grammatical form changed the value or destination identity")
			}
		})
	}
}

func TestConfiguredFormCorrections(t *testing.T) {
	gram, err := loadGrammar(canonicalTestOptions())
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ spelling, role, form string }{
		{"Ran", "action.executive", "passive"},
		{"Putter", "action.expansive", "objective"},
		{"Stopper", "action.interruptive", "objective"},
	} {
		word, category, lookupErr := gram.lex.lookup([]byte(tc.spelling))
		if lookupErr != nil || word.Type != tc.role || !strings.HasSuffix(category, "."+tc.form) {
			t.Errorf("%s classified as %s, %v; want %s/%s", tc.spelling, category, lookupErr, tc.role, tc.form)
		}
	}
	for _, spelling := range []string{"Ruan", "Puter", "Stoppper"} {
		if _, category, lookupErr := gram.lex.lookup([]byte(spelling)); lookupErr == nil {
			t.Errorf("retired malformed spelling %s remains %s", spelling, category)
		}
	}
}

func TestPassivePutAppValueIsOuterUpdate(t *testing.T) {
	result := requireDistributionFrontend(t, "Create \"Worker\" as a New Basic App.\nDefine \"Items\" as a New List of App.\nWorker is Put in Items.\n", nil)
	var updates []*instruction
	for _, operation := range result.Compilation.Semantics.Instructions {
		if operation.Kind == instructionUpdate {
			updates = append(updates, operation)
		}
	}
	if len(updates) != 1 || updates[0].Value.Type != valueApp || updates[0].Target.Type != valueList {
		t.Fatalf("App value was not retained in exactly one outer update: %+v", updates)
	}
}

func TestPassivePutRejectsUnsupportedAndIncompatibleOperands(t *testing.T) {
	for _, tc := range []struct{ name, source, diagnostic string }{
		{"pointer", "Define \"Payload\" as 7.\nDefine \"Target\" as a New Pointer of 1.\nPayload is Put in Target.\n", "passive Put requires a List or Map"},
		{"list-value", "Define \"Payload\" as `text`.\nDefine \"Items\" as a New List of Number.\nPayload is Put in Items.\n", "List update has incompatible value type"},
		{"map-value", "Define \"Payload\" as 7.\nDefine \"Items\" as a New Map of Text.\nPayload is Put in Items as `key`.\n", "Map update has incompatible value type"},
		{"map-key", "Define \"Payload\" as `text`.\nDefine \"Items\" as a New Map of Text.\nPayload is Put in Items.\n", "field key requires Text"},
		{"orphan-passive", "Define \"Payload\" as 7.\nDefine \"Items\" as a New List of Number.\nPut Payload in Items and then Put in Items.\n", "passive Put requires a grammatical subject"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, err := distributionFrontend(t, tc.source+"Create \"Witness\" as a New Basic App.\n", nil)
			if err == nil {
				_, err = result.OrchestratorUnit("entry", executableTestCompanions(result))
			}
			if err == nil || !strings.Contains(err.Error(), tc.diagnostic) {
				t.Fatalf("expected %q, got %v", tc.diagnostic, err)
			}
		})
	}
}

func TestSameSpellingFunctionPassiveRejectsUnsupportedLowering(t *testing.T) {
	for _, inputs := range []string{"", "\"Value\" as Text"} {
		source := functionSource("Read|-s|-|-ing|-er", inputs, "", "Show `body`.\n")
		_, err := distributionFrontend(t, source+"Define \"Target\" as 1.\nDefine \"Payload\" as `value`.\nTarget is Read Payload.\n", nil)
		if !errors.Is(err, lerrors.Unsupported) || !strings.Contains(err.Error(), "passive Function invocation") {
			t.Fatalf("unsupported passive operands were dropped or misdiagnosed: %v", err)
		}
	}
}

func TestConstructionRequiresNominalRole(t *testing.T) {
	const leaf = "Define \"Fold|-s|-ed|-ing|-er\" to `<DEF>KY{|Key: ~std::convert::identity~, Inputs: ~&[<@Value@>]&~|}<DEF>`.\n"
	for _, declaration := range []string{leaf, "Create \"Fold|-s|-ed|-ing|-er\" for Showing any Number.\n"} {
		for _, spelling := range []string{"Fold", "Folds", "Folded", "Folding"} {
			for _, modifier := range []string{"New", "Generic"} {
				source := declaration + "Create \"Selected\" as a " + modifier + " " + spelling + " of 7.\n"
				c := paragraphPublicationSource(t, source)
				if !errors.Is(c.Error, lerrors.Syntax) {
					t.Errorf("%s %s accepted an operation as a noun: %v", modifier, spelling, c.Error)
				}
			}
		}
	}
	for name, source := range map[string]string{
		"actor":         leaf + "Create \"Selected\" as a New Folder of 7.\n",
		"alias actor":   leaf + "Use Fold as \"Alias\".\nCreate \"Selected\" as a New Aliaser of 7.\n",
		"noun alias":    "Define \"Folded\" as `<DEF>E{|Statement: ~7~|}<DEF>`.\nUse Folded as \"Alias\".\nCreate \"Selected\" as a New Alias.\n",
		"descriptor":    leaf + "Define \"Folded\" as a New Adjective describing any Folder.\nIt describes any Folder as `<DEF><@Folder@><DEF>`.\nCreate \"Selected\" as a New Folded Folder of 7.\n",
		"foreign noun":  leaf + "Create \"Other\" as a Dictionary.\nDefine Other's \"Folded\" as `<DEF>E{|Statement: ~7~|}<DEF>`.\nCreate \"Selected\" as a New Other.Folded.\n",
		"Function type": identityFunction + runtimeFunctionReadSource,
	} {
		t.Run(name, func(t *testing.T) {
			c := paragraphPublicationSource(t, source)
			if c.Error != nil {
				t.Fatal(c.Error)
			}
			if name == "Function type" {
				return
			}
			want := map[string]string{"actor": "Fold", "alias actor": "Fold", "noun alias": "Folded", "descriptor": "Fold", "foreign noun": "Other.Folded"}[name]
			last := c.Statements[len(c.Statements)-1]
			matched := false
			for _, tok := range last.tokens {
				if tok.binding == nil || tok.isDeclare() || tok.isPossessive() || tok.getPOS(true) != "noun" {
					continue
				}
				canonical, err := canonicalRecord(tok.binding.Record, c.Semantics.Aliases)
				if err != nil || canonical.ID != want {
					t.Fatalf("construction selected %v, want %s: %v", canonical, want, err)
				}
				matched = true
			}
			if !matched {
				t.Fatal("construction has no original nominal binding")
			}
		})
	}
}

func TestSourceActionConstructorRequiresNominalRole(t *testing.T) {
	for _, form := range []string{"Fold", "Folded"} {
		source := "Define \"Fold|-s|-ed|-ing|-er\" to `<DEF>KY{|Key: ~std::convert::identity~, Inputs: ~&[<@Value@>]&~|}<DEF>`.\n" +
			"Create \"Wrap|-s|-ped|-ping|-per\" for Processing a New " + form + " of any Number.\nCreate \"Worker\" as a New Basic App.\nWorker Wraps 7.\n"
		_, err := distributionFrontend(t, source, nil)
		if !errors.Is(err, lerrors.Syntax) || !strings.Contains(err.Error(), "construction requires a noun or actor") {
			t.Fatalf("deferred %s constructor accepted: %v", form, err)
		}
	}
}

func TestConstructedActorValueRegistersNoun(t *testing.T) {
	c := paragraphPublicationSource(t, "Define \"Fold|-s|-ed|-ing|-er\" to `<DEF>KY{|Key: ~std::convert::identity~, Inputs: ~&[<@Value@>]&~|}<DEF>`.\nDefine \"Parcel\" as a New Folder of any Thing.\n")
	if c.Error != nil {
		t.Fatal(c.Error)
	}
	record := c.Memory["Parcel"]
	if record == nil || record.bindingWord().POS != "noun" {
		t.Fatal("constructed actor value inherited its operation's verb role")
	}
	if matched, _ := record.binding(newBindingQuery([]byte("Parcel"))); matched == nil || matched.lexicalType() != "nouns.generic.root" {
		t.Fatal("noun specialization lost its lexical binding")
	}
}

func TestSameSpellingFormsBuildMemoryRetry(t *testing.T) {
	const source = "Create \"Worker\" as a New Basic App.\nCreate \"Read|-s|-|-ing|-er\" for Showing any Thing.\n" +
		"Define \"Item\" as `<DEF>KY{|Key: ~std::convert::identity~, Inputs: ~&[<@Value@>]&~|}<DEF>`.\n" +
		"Define \"Wrapped\" as a New Adjective describing any Item.\nWorker is Read a New Wrapped Item of 7.\n" +
		"It describes any Item as `<DEF><@Item@><DEF>`.\n"
	c := newCompilation([]byte(source), "retry-forms.l")
	c.init(canonicalTestOptions())
	for _, phase := range []func(){c.preProcess, c.build, c.scan} {
		phase()
		if c.Error != nil {
			t.Fatal(c.Error)
		}
	}
	var statement *lStatement
	for _, candidate := range testProgramStatements(c.Statements) {
		if candidate.source.Span.Start.Line == 5 {
			statement = candidate
		}
	}
	if statement == nil {
		t.Fatal("missing original caller")
	}
	verb, span := statement.tokens[2], statement.source.Span
	var selected *record
	failed := 0
	c.Gram.constructorSignature = func(noun *token, modifiers []*token, named bool) (grammarArity, error) {
		arity, err := c.constructorGrammarSignature(noun, modifiers, named)
		if noun.scope == verb.scope && errors.Is(err, lerrors.UnresolvedReference) {
			failed++
			if verb.binding == nil || verb.binding.Form != "root" {
				t.Fatal("unresolved grammar committed passive form")
			}
			selected = verb.binding.Record
		}
		return arity, err
	}
	c.analyze()
	if c.Error != nil {
		t.Fatal(c.Error)
	}
	if failed == 0 || selected == nil || statement.tokens[2] != verb || verb.binding.Record != selected || verb.binding.Form != "passive" || statement.source.Span != span || string(verb.rawBytes) != "Read" {
		t.Fatalf("real unresolved-reference retry lost identity, form, or provenance: attempts=%d binding=%+v", failed, verb.binding)
	}
}

func TestSameSpellingNativeFunctionRequiresDeclaredPassive(t *testing.T) {
	for _, word := range []string{"Read", "Reads", "Reading"} {
		source := "    Publicly create \"Read|-s||-ing|-er\" as a Generic Function with Inputs of \"value\" as `f64` and Outputs of `f64`.\nReturn Value.\n\nCreate \"Worker\" as a New Basic App.\nWorker is " + word + " 7.\n"
		_, err := distributionFrontend(t, source, nil)
		if !errors.Is(err, lerrors.Syntax) || !strings.Contains(err.Error(), "declared passive form") {
			t.Fatalf("%s acquired an undeclared native Function passive: %v", word, err)
		}
	}
}

func TestConstructorRoleActorAliasAndDescriptor(t *testing.T) {
	const words = "Define \"Fold|-s|-ed|-ing|-er\" to `<DEF>KY{|Key: ~std::convert::identity~, Inputs: ~&[<@Value@>]&~|}<DEF>`.\nCreate \"Catalog\" as a Dictionary.\nUse Folder as Catalog's \"Parcel\".\nUse Catalog.Parcel as \"Copy\".\nUse Fold as \"Relay\".\nUse Folded as \"Spared\".\n" +
		"Define \"Tracked\" as a New Adjective describing any Folder.\nIt describes any Folder as `<DEF><@Folder@><DEF>`.\n"
	c := paragraphPublicationSource(t, words+"Create \"Selected\" as a New Copy of 7.\nCreate \"Wrapped\" as a New Tracked Relay of 7.\nCreate \"Passivealias\" as a New Tracked Spared of 7.\n")
	if c.Error != nil {
		t.Fatal(c.Error)
	}
	aliasedTarget := strings.ReplaceAll(strings.ReplaceAll(words, "any Folder", "any Copy"), "<@Folder@>", "<@Copy@>")
	if throughAlias := paragraphPublicationSource(t, aliasedTarget+"Create \"Wrapped\" as a New Tracked Relay of 7.\n"); throughAlias.Error != nil {
		t.Fatal(throughAlias.Error)
	}
	for _, name := range []string{"Catalog.Parcel", "Copy"} {
		alias := c.Memory[name]
		if alias == nil || alias.bindingWord().POS != "noun" || len(alias.bindingWord().Postfixes) != 0 || len(alias.bindingWord().Prefixes) != 0 {
			t.Fatalf("actor alias %s inherited operation morphology", name)
		}
		canonical, err := canonicalRecord(alias, c.Semantics.Aliases)
		if err != nil || canonical != c.Memory["Fold"] {
			t.Fatalf("actor alias %s lost canonical identity: %v", name, err)
		}
	}
	for spelling, form := range map[string]string{"Relay": "root", "Relayed": "passive", "Relayer": "objective"} {
		binding, ok := c.Memory["Relay"].binding(newBindingQuery([]byte(spelling)))
		if !ok || binding.Form != form || binding.POS != "verb" {
			t.Fatalf("root alias lost %s: %+v", form, binding)
		}
	}
	passiveAlias := c.Memory["Spared"]
	if passiveAlias == nil || passiveAlias.bindingWord().POS != "verb" || c.Semantics.Aliases[passiveAlias] != c.Memory["Fold"] {
		t.Fatal("passive source alias lost its renamed verb family")
	}
	for _, statement := range testProgramStatements(c.Statements) {
		for _, tok := range statement.tokens {
			if string(tok.rawBytes) == "Folded" && (tok.binding == nil || tok.binding.Record != c.Memory["Fold"] || tok.binding.Form != "passive") {
				t.Fatal("aliasing changed the original passive source role")
			}
		}
	}
	for _, construction := range []string{"Relay", "Relayed", "Public Relay", "Tracked Relayed", "Spared", "Public Spared", "Tracked Folded", "Tracked Spareded"} {
		bad := paragraphPublicationSource(t, words+"Create \"Invalid\" as a New "+construction+" of 7.\n")
		if !errors.Is(bad.Error, lerrors.Syntax) {
			t.Fatalf("unrelated operation role %s accepted: %v", construction, bad.Error)
		}
	}
}
