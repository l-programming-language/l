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
	word := indexedTestWord(t, "Read", `"Read|-s||-ing|-er"`, true)
	statement := boundFormStatement(t, "Target is Read Value", map[string]*record{"Read": word})
	if err := statement.loadGram(); err == nil {
		t.Fatal("Function without a declared passive gained a passive fallback")
	}
	if statement.tokens[2].binding.Form != "root" {
		t.Fatal("failed undeclared passive changed its root token")
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

func TestSameSpellingSourcePassiveAliasDefinitionSite(t *testing.T) {
	files := dictionaryIndexFilesystem(t)
	library := "\nDefine \"Secret\" as `library`.\nPublicly create IO's \"Read|-s|-|-ing|-er\" for IO.Printing Secret and then IO.Printing any Text.\n"
	files["stdlib/eng/io/dictionary.l"].Data = append(files["stdlib/eng/io/dictionary.l"].Data, []byte(library)...)
	source := "Use IO's Read from `io` as \"Alias\".\nDefine \"Secret\" as `caller shadow`.\nCreate \"Worker\" as a New Basic App.\nWorker is Alias `payload`.\nRun Worker.\nAwait Worker.\nShow Worker's Output.\n"
	result, err := NewLygerService(Options{Resources: files}).Compile([]byte(source), "/project/forms.l")
	if err != nil {
		t.Fatal(err)
	}
	captureNativeResult(t, result.Compilation, "library\npayload\n")
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

func TestPassivePutNative(t *testing.T) {
	for _, tc := range []struct{ name, source, want string }{
		{"list", "Define \"Payload\" as 7.\nDefine \"Items\" as a New List of Number.\nPayload is Put in Items.\nShow Items.\n", "[7]"},
		{"map", "Define \"Payload\" as `value`.\nDefine \"Items\" as a New Map of Text.\nPayload is Put in Items as `key`.\nShow a New Text from Items at `key`.\n", "value"},
		{"root-list", "Define \"Payload\" as 7.\nDefine \"Items\" as a New List of Number.\nPut Payload in Items.\nShow Items.\n", "[7]"},
		{"root-map", "Define \"Payload\" as `value`.\nDefine \"Items\" as a New Map of Text.\nPut Payload in Items as `key`.\nShow a New Text from Items at `key`.\n", "value"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runtimeBundleResult(t, runtimeValuesBundle(t, tc.source), tc.want, "")
		})
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

func TestPassivePutContinuationNative(t *testing.T) {
	source := "Define \"Payload\" as 7.\nDefine \"Items\" as a New List of Number.\nDefine \"Others\" as a New List of Number.\nPayload is Put in Items and then Put in Others and then Put 8 in Items.\nShow Items and Others.\n"
	runtimeBundleResult(t, runtimeValuesBundle(t, source), "[7, 8][7]", "")
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
