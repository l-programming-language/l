package Lyger

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/l-programming-language/l/lerrors"
)

func TestMorphologyRepeatedSuffixPreservesRoles(t *testing.T) {
	word := newLexeme("verb")
	if err := word.reLabel(`"Read|-s|-|-|-"`); err != nil {
		t.Fatal(err)
	}
	for _, role := range []string{"root", "passive", "applicative", "objective"} {
		if !word.matchesForm("read", role) {
			t.Errorf("same-spelled %s disappeared from the declared family", role)
		}
	}
	if matched, role := word.match([]byte("read")); !matched || role != "root" {
		t.Fatalf("context-free lookup = %t/%s, want root precedence", matched, role)
	}
}

func TestMorphologyCompleteFamily(t *testing.T) {
	for _, tc := range []struct{ pos, label string }{
		{"noun", `"Process|-es|-ive|-ing"`},
		{"verb", `"Execute|-s|-d|--ing|--or"`},
		{"verb", `"Read|-s|-|-|-"`},
		{"noun", `"Thing|-s|-|-"`},
	} {
		t.Run(tc.label, func(t *testing.T) {
			word := newLexeme(tc.pos)
			if err := word.declareFamily(tc.label); err != nil {
				t.Fatal(err)
			}
			if err := word.validateFamily(); err != nil {
				t.Fatal(err)
			}
		})
	}
	for _, tc := range []struct{ pos, label string }{
		{"noun", `"Process"`}, {"verb", `"Start"`},
		{"noun", `"Process|-es|-ive"`}, {"verb", `"Start|-s|-ed|-ing"`},
		{"noun", `"Process|-es||-ing"`}, {"verb", `"Start|-s|-ed|-ing|"`},
		{"noun", `"Process|-es|progressive|-ing"`},
		{"noun", `"Process|-es|-ive|-ing|-or"`},
		{"verb", `"Start|-s|-ed|-ing|-er|-extra"`},
		{"noun", `"Thing|-|-|-"`}, {"verb", `"Read|-|--d|---ad|----ead"`},
		{"verb", `"Read|-|--D|---AD|----EAD"`},
		{"verb", `"Read|-s|------past|-ing|-er"`},
		{"verb", `"Read|-s|-----|-ing|-er"`},
		{"verb", `"R\u002eead|-s|-|-ing|-er"`},
		{"verb", `"Read|-s|-\u002e|-ing|-er"`},
		{"verb", "\"Read|-s|-`|-ing|-er\""},
		{"verb", `"Read|-s|-\u0001|-ing|-er"`},
		{"verb", `"é|--s|-ed|-ing|-er"`},
		{"adjective", `"Basic|-s|-ed|-ing|-er"`},
	} {
		t.Run("reject/"+tc.label, func(t *testing.T) {
			word := newLexeme(tc.pos)
			word.Root = "unchanged"
			before := *word
			if err := word.declareFamily(tc.label); !errors.Is(err, lerrors.Syntax) {
				t.Fatalf("invalid family error = %v, want Syntax", err)
			}
			if !reflect.DeepEqual(*word, before) {
				t.Fatal("invalid declaration mutated its prior word")
			}
		})
	}
}

func TestMorphologyLabelParsing(t *testing.T) {
	label, err := parseDeclarationLabel(`"R\u0065ad|-s|-|-ing|-er"`)
	if err != nil || label.Root != "Read" || len(label.Slots) != 5 {
		t.Fatalf("decoded label = %+v, %v", label, err)
	}
	for _, raw := range []string{`Read`, `"Read`, `Read"`, `""`, `"|-s|-ed|-ing|-er"`, `"Two Words"`, `"Read\q"`} {
		if _, err := parseDeclarationLabel(raw); !errors.Is(err, lerrors.Syntax) {
			t.Errorf("invalid label %s: %v", raw, err)
		}
	}
	word := newLexeme("noun")
	if err := word.reLabel(`"Message"`); err != nil || word.Root != "message" || len(word.Forms) != 0 {
		t.Fatalf("ordinary name acquired required family cases: %+v, %v", word, err)
	}
}

func TestMorphologyRequiresEveryDeclaredCase(t *testing.T) {
	for _, family := range []struct{ pos, label string }{
		{"noun", "Parcel|-s|-like|-ing"},
		{"verb", "Fold|-s|-ed|-ing|-er"},
	} {
		parts := strings.Split(family.label, "|")
		for slot := 1; slot < len(parts); slot++ {
			for _, defect := range []string{"omitted", "empty", "malformed"} {
				t.Run(family.pos+"/"+strconv.Itoa(slot)+"/"+defect, func(t *testing.T) {
					invalid := append([]string(nil), parts...)
					switch defect {
					case "omitted":
						invalid = append(invalid[:slot], invalid[slot+1:]...)
					case "empty":
						invalid[slot] = ""
					case "malformed":
						invalid[slot] = "not-a-suffix"
					}
					word := newLexeme(family.pos)
					if err := word.declareFamily(strconv.Quote(strings.Join(invalid, "|"))); !errors.Is(err, lerrors.Syntax) {
						t.Fatalf("invalid case accepted: %v", err)
					}
					if word.Root != "" || len(word.Forms) != 0 {
						t.Fatal("rejected case published part of a family")
					}
				})
			}
		}
	}
}

func TestMorphologyDeclaredCaseOrder(t *testing.T) {
	word := newLexeme("verb")
	if err := word.declareFamily(`"Bake|-d|--ed|--ing|-r"`); err != nil {
		t.Fatal(err)
	}
	for repeat := 0; repeat < 20; repeat++ {
		if matched, role := word.match([]byte("baked")); !matched || role != "active" {
			t.Fatalf("same-spelled nonroot forms changed order: %t/%s", matched, role)
		}
	}
	if !word.matchesForm("baked", "passive") {
		t.Fatal("exact passive role was erased by context-free precedence")
	}
}

func TestMorphologyConfiguredJSONRoundTrip(t *testing.T) {
	const source = `{"verbs":{"reader":{"root":"read","pos":"verb","postfixes":{"-s":"active","-":"passive"},"prefixes":{"un-":"negative"}}}}`
	var vocabulary lexicon
	if err := json.Unmarshal([]byte(source), &vocabulary); err != nil {
		t.Fatal(err)
	}
	word := vocabulary.Verbs["reader"]
	if !word.matchesForm("read", "passive") || !word.matchesForm("unread", "negative") {
		t.Fatal("configured suffix/prefix forms were not decoded")
	}
	word.Root = "reap"
	vocabulary.Verbs["reader"] = word
	encoded, err := json.Marshal(vocabulary)
	if err != nil {
		t.Fatal(err)
	}
	var raw struct {
		Verbs map[string]struct {
			Postfixes map[string]string `json:"postfixes"`
			Prefixes  map[string]string `json:"prefixes"`
		} `json:"verbs"`
	}
	if err := json.Unmarshal(encoded, &raw); err != nil || raw.Verbs["reader"].Postfixes["-s"] != "active" || raw.Verbs["reader"].Prefixes["un-"] != "negative" {
		t.Fatalf("configuration map format changed: %s (%v)", encoded, err)
	}
	var restored lexicon
	if err := json.Unmarshal(encoded, &restored); err != nil {
		t.Fatal(err)
	}
	word = restored.Verbs["reader"]
	if !word.matchesForm("reaps", "active") || !word.matchesForm("unreap", "negative") {
		t.Fatal("configured root replacement did not retain its inflections")
	}
}

func TestMorphologyNativeBindingLabels(t *testing.T) {
	for _, label := range []string{"Payload", "Payload|-s", "Payload|-s|-like|-ing"} {
		for _, kind := range []string{"Process input", "native local result"} {
			t.Run(kind+"/"+label, func(t *testing.T) {
				source := typedProcessSource("String", "Inputs of \""+label+"\" as `String`", "Returning a New Fallback")
				if kind == "native local result" {
					source = "Define \"Cell|-s|-ular|-ing\" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~<@Value@>~|}<DEF>`.\n\n" +
						"    Publicly create \"Exercise|-s|-d|--ing|-r\" as a Generic `exercise` Function with No Inputs and Outputs of `f64`.\n" +
						"Define \"" + label + "\" as a New Cell of 7.\nReturn Payload.\n"
				}
				compilation := NewLygerService(canonicalTestOptions()).NewCompilation([]byte(source), "binding-forms.l")
				compilation.Make()
				if label == "Payload|-s" {
					if !errors.Is(compilation.Error, lerrors.Syntax) || !strings.Contains(compilation.Error.Error(), "family") {
						t.Fatalf("partial %s family escaped registration: %v", kind, compilation.Error)
					}
				} else if compilation.Error != nil {
					t.Fatalf("ordinary name or complete family failed: %v", compilation.Error)
				}
			})
		}
	}
}

func TestMorphologySharedPassiveAndActorSpelling(t *testing.T) {
	const native = "Define \"Fold|-s|-ed|-ing|%s\" to `<DEF>KY{|Key: ~std::convert::identity~, Inputs: ~&[<@Value@>]&~|}<DEF>`.\n"
	for _, actor := range []struct{ fix, spelling string }{{"-er", "Folder"}, {"-ed", "Folded"}} {
		t.Run("constructor/"+actor.spelling, func(t *testing.T) {
			source := strings.Replace(native, "%s", actor.fix, 1) + "Define \"Parcel|-s|-like|-ing\" as a New " + actor.spelling + " of 7.\n"
			compilation := paragraphPublicationSource(t, source)
			if compilation.Error != nil {
				t.Fatal(compilation.Error)
			}
			found := false
			for _, statement := range testProgramStatements(compilation.Statements) {
				for _, tok := range statement.tokens {
					if string(tok.rawBytes) == actor.spelling {
						found = true
						if tok.binding == nil || tok.binding.Record != compilation.Memory["Fold"] || tok.binding.Form != "objective" {
							t.Fatalf("constructor lost canonical objective form: %+v", tok.binding)
						}
					}
				}
			}
			if !found {
				t.Fatal("actor use not observed")
			}
		})
	}
	for _, construction := range []string{"Generic", "New Marked"} {
		t.Run(construction+"/Folded", func(t *testing.T) {
			source := strings.Replace(native, "%s", "-ed", 1) +
				"Define \"Marked\" as a New Adjective describing any Folded.\n" +
				"It describes any Folded as `<DEF><@Folded@><DEF>`.\n" +
				"Define \"Parcel|-s|-like|-ing\" as a " + construction + " Folded of 7.\n"
			compilation := paragraphPublicationSource(t, source)
			if compilation.Error != nil {
				t.Fatal(compilation.Error)
			}
			for _, statement := range testProgramStatements(compilation.Statements) {
				for _, tok := range statement.tokens {
					if string(tok.rawBytes) == "Folded" && (tok.binding == nil || tok.binding.Record != compilation.Memory["Fold"] || tok.binding.Form != "objective") {
						t.Fatalf("nominal context lost canonical actor form: %+v", tok.binding)
					}
				}
			}
		})
	}
	t.Run("passive", func(t *testing.T) {
		compilation := paragraphPublicationSource(t, "Create \"Fold|-s|-ed|-ing|-ed\" for Printing any Text.\nCreate \"Worker\" as a New Basic App.\nWorker is Folded `payload`.\n")
		if compilation.Error != nil {
			t.Fatal(compilation.Error)
		}
		for _, statement := range testProgramStatements(compilation.Statements) {
			for _, tok := range statement.tokens {
				if string(tok.rawBytes) == "Folded" && (tok.binding == nil || tok.binding.Record != compilation.Memory["Fold"] || tok.binding.Form != "passive") {
					t.Fatalf("passive use lost its canonical form: %+v", tok.binding)
				}
			}
		}
	})
}

func TestMorphologyNominalSpeculationAndRetry(t *testing.T) {
	word := indexedTestWord(t, "Fold", `"Fold|-s|-ed|-ing|-ed"`, false)
	statement := boundFormStatement(t, `Create "Parcel|-s|-like|-ing" as a New Folded`, map[string]*record{"Fold": word})
	actor := statement.tokens[len(statement.tokens)-1]
	before, category, raw := actor.binding, actor.tType, bytes.Clone(actor.rawBytes)
	object, err := initObjectGram("object", statement.tokens[3:], statement.gram)
	if err == nil {
		err = object.build()
	}
	if err != nil {
		t.Fatal(err)
	}
	if actor.binding != before || actor.binding.Form != "passive" || actor.tType != category || !bytes.Equal(actor.rawBytes, raw) {
		t.Fatal("speculative constructor changed the source token")
	}
	failed := boundFormStatement(t, `Create "Parcel|-s|-like|-ing" as a New Folded Value`, map[string]*record{"Fold": word})
	actor = failed.tokens[len(failed.tokens)-2]
	before, category, raw = actor.binding, actor.tType, bytes.Clone(actor.rawBytes)
	if err := failed.loadGram(); err == nil {
		t.Fatal("trailing garbage was accepted")
	}
	if actor.binding != before || actor.binding.Form != "passive" || actor.tType != category || !bytes.Equal(actor.rawBytes, raw) {
		t.Fatal("failed statement committed an actor view")
	}
	failed.tokens = failed.tokens[:len(failed.tokens)-1]
	failed.determineType()
	if err := failed.loadGram(); err != nil {
		t.Fatal(err)
	}
	if actor.binding.Record != word || actor.binding.Form != "objective" || !bytes.Equal(actor.rawBytes, raw) {
		t.Fatalf("successful retry lost canonical actor role: %+v", actor.binding)
	}
}

func TestMorphologyDescriptiveCollisionContext(t *testing.T) {
	word := newLexeme("noun")
	if err := word.declareFamily(`"Parcel|-s|-s|-ing"`); err != nil {
		t.Fatal(err)
	}
	parcel := &record{ID: "Parcel", Value: &entry{Type: storageLexeme, Lexeme: word}}
	for _, trailing := range []string{"", " Target"} {
		statement := boundFormStatement(t, `Create "Saved" as a New Parcels Value`+trailing, map[string]*record{"Parcel": parcel})
		var modifier *token
		for _, tok := range statement.tokens {
			if string(tok.rawBytes) == "Parcels" {
				modifier = tok
			}
		}
		before, category := modifier.binding, modifier.tType
		err := statement.loadGram()
		if trailing != "" {
			if err == nil || modifier.binding != before || modifier.tType != category || modifier.binding.Form != "plural" {
				t.Fatalf("failed adjective candidate changed the source token: %v", err)
			}
		} else if err != nil || modifier.binding.Record != parcel || modifier.binding.Form != "descriptive" || modifier.getPOS(true) != "adjective" {
			t.Fatalf("descriptive collision was not selected: %+v, %v", modifier.binding, err)
		}
	}
	statement := boundFormStatement(t, `Show Parcels`, map[string]*record{"Parcel": parcel})
	if err := statement.loadGram(); err != nil || statement.tokens[1].binding.Form != "plural" {
		t.Fatalf("bare nominal lost context-free precedence: %v", err)
	}
	// A noun's descriptive spelling can precede its own root without resolving
	// another word or consuming that final root as an additional adjective.
	word = newLexeme("noun")
	if err := word.declareFamily(`"Parcel|-s|-|-ing"`); err != nil {
		t.Fatal(err)
	}
	parcel.Value.Lexeme = word
	statement = boundFormStatement(t, `Create "Saved" as a New Parcel Parcel`, map[string]*record{"Parcel": parcel})
	if err := statement.loadGram(); err != nil {
		t.Fatal(err)
	}
	var roles []string
	for _, tok := range statement.tokens {
		if string(tok.rawBytes) == "Parcel" {
			roles = append(roles, tok.binding.Form)
		}
	}
	if !reflect.DeepEqual(roles, []string{"descriptive", "root"}) {
		t.Fatalf("same-root modifier and noun roles = %v", roles)
	}
}

func TestMorphologyNounActiveOverlap(t *testing.T) {
	for _, tc := range []struct{ label, spelling string }{
		{"Parcel|-s|-like|-s", "Parcels"},
		{"Parcel|-s|-like|-like", "Parcellike"},
		{"Parcel|-s|-like|-", "Parcel"},
		{"Parcel|-s|-like|-ing", "Parceling"},
	} {
		t.Run(tc.spelling, func(t *testing.T) {
			source := "Define \"" + tc.label + "\" as `<DEF>E{|Statement: ~<@Value|Text@>~|}<DEF>`.\n" +
				"Create \"Fold|-s|-ed|-ing|-er\" for " + tc.spelling + " any Text.\n\n" +
				"    Publicly create \"Exercise|-s|-d|--ing|--er\" as a Generic `exercise` Function with No Inputs and Outputs of `&'static str`.\n" +
				"Return a New Folder of `payload`.\n"
			compilation := compileMemberResolutionSource(t, memberResolutionSource{main: source})
			if compilation.Error != nil {
				t.Fatal(compilation.Error)
			}
			var calls int
			for _, statement := range testProgramStatements(compilation.Statements) {
				for _, tok := range statement.tokens {
					if string(tok.rawBytes) == tc.spelling {
						calls++
						if tok.binding == nil || tok.binding.Record != compilation.Memory["Parcel"] || tok.binding.Form != "active" || tok.getPOS(true) != "verb" {
							t.Fatalf("noun application lost its selected active role: %+v", tok.binding)
						}
					}
				}
			}
			if calls != 1 {
				t.Fatalf("observed %d noun applications, want one", calls)
			}
			testNativeDefinitionLibrary(t, compilation, "Exercise", "noun_active", `#[path="src/noun_active.rs"]mod generated;fn main(){assert_eq!(generated::exercise(),"payload");}`)
		})
	}
}

func morphologyNounRecord(t *testing.T, id, label string) *record {
	t.Helper()
	word := newLexeme("noun")
	if err := word.declareFamily(label); err != nil {
		t.Fatal(err)
	}
	if err := word.define([]byte("<DEF>E{|Statement: ~<@Value|Text@>~|}<DEF>")); err != nil {
		t.Fatal(err)
	}
	return &record{ID: id, Value: &entry{Type: storageLexeme, Lexeme: word}}
}

func TestMorphologyNounActiveGrammarViews(t *testing.T) {
	for _, tc := range []struct{ label, spelling, initial string }{
		{`"Parcel|-s|-like|-s"`, "Parcels", "plural"},
		{`"Parcel|-s|-like|-like"`, "Parcellike", "descriptive"},
		{`"Parcel|-s|-like|-"`, "Parcel", "root"},
		{`"Parcel|-s|-like|-ing"`, "Parceling", "active"},
	} {
		for _, grammar := range []string{"clause.applicative", "clause.generic.applicative"} {
			for _, prefix := range []string{"", "not Catalog's "} {
				t.Run(tc.spelling+"/"+grammar+"/"+prefix, func(t *testing.T) {
					id := "Parcel"
					if prefix != "" {
						id = "Catalog.Parcel"
					}
					selected := morphologyNounRecord(t, id, tc.label)
					records := map[string]*record{id: selected}
					if prefix != "" {
						// The owner also has an active/root overlap. Its possessive
						// must remain a prefix, and the foreign homonym must stay foreign.
						owner := morphologyNounRecord(t, "Catalog", `"Catalog|-s|-like|-"`)
						owner.members = map[string]*record{id: selected}
						records["Catalog"] = owner
						records["Parcel"] = morphologyNounRecord(t, "Parcel", tc.label)
					}
					statement := boundFormStatement(t, prefix+tc.spelling+" Value", records)
					head := statement.tokens[len(statement.tokens)-2]
					before, category, raw := head.binding, head.tType, bytes.Clone(head.rawBytes)
					if before == nil || before.Record != selected || before.Form != tc.initial {
						t.Fatalf("context-free lookup changed before grammar: %+v", before)
					}
					clause, err := initClauseGram(grammar, statement.tokens, statement.gram)
					if err == nil {
						err = clause.build()
					}
					if err != nil {
						t.Fatal(err)
					}
					if clause.verb != head || len(clause.tokens) != len(statement.tokens) {
						t.Fatal("applicative grammar did not retain the complete original clause")
					}
					if head.binding != before || head.tType != category || !bytes.Equal(head.rawBytes, raw) {
						t.Fatal("speculative applicative grammar changed its source token")
					}
					accepted := &statementGram{clauses: []*clauseGram{clause}}
					if err := accepted.commitForms(); err != nil {
						t.Fatal(err)
					}
					if head.binding.Record != selected || head.binding.Form != "active" || head.getPOS(true) != "verb" || !bytes.Equal(head.rawBytes, raw) {
						t.Fatalf("accepted clause changed identity or lost its active form: %+v", head.binding)
					}
					if prefix != "" && (statement.tokens[1].binding.Form != "root" || !statement.tokens[1].isPossessive()) {
						t.Fatal("applicative context consumed its possessive owner")
					}
				})
			}
		}
	}
}

func TestMorphologyNounActiveRetry(t *testing.T) {
	for _, tc := range []struct{ label, spelling string }{
		{`"Parcel|-s|-like|-s"`, "Parcels"},
		{`"Parcel|-s|-like|-like"`, "Parcellike"},
		{`"Parcel|-s|-like|-"`, "Parcel"},
	} {
		t.Run(tc.spelling, func(t *testing.T) {
			selected := morphologyNounRecord(t, "Parcel", tc.label)
			statement := boundFormStatement(t, `Create "Fold|-s|-ed|-ing|-er" for `+tc.spelling+" Value Target", map[string]*record{"Parcel": selected})
			head := statement.tokens[len(statement.tokens)-3]
			before, category, raw := head.binding, head.tType, bytes.Clone(head.rawBytes)
			if err := statement.loadGram(); err == nil {
				t.Fatal("trailing garbage was accepted")
			}
			if head.binding != before || head.tType != category || !bytes.Equal(head.rawBytes, raw) {
				t.Fatal("failed applicative statement committed a form")
			}
			statement.tokens = statement.tokens[:len(statement.tokens)-1]
			statement.determineType()
			if err := statement.loadGram(); err != nil {
				t.Fatal(err)
			}
			if head.binding.Record != selected || head.binding.Form != "active" || !bytes.Equal(head.rawBytes, raw) {
				t.Fatalf("successful retry lost its canonical noun application: %+v", head.binding)
			}
		})
	}
}

func TestMorphologyNounActiveRequiresContextAndDeclaredForm(t *testing.T) {
	for _, tc := range []struct{ label, spelling, role, grammar string }{
		{`"Parcel|-s|-like|-ing"`, "Parcels", "plural", "clause.applicative"},
		{`"Parcel|-s|-like|-ing"`, "Parcellike", "descriptive", "clause.generic.applicative"},
		{`"Parcel|-s|-like|-ing"`, "Parcel", "root", "clause.applicative"},
		{`"Parcel|-s|-like|-s"`, "Parcels", "plural", "clause.imperative"},
		{`"Parcel|-s|-like|-like"`, "Parcellike", "descriptive", "clause.passive"},
		{`"Parcel|-s|-like|-"`, "Parcel", "root", "clause"},
	} {
		t.Run(tc.spelling+"/"+tc.grammar, func(t *testing.T) {
			selected := morphologyNounRecord(t, "Parcel", tc.label)
			statement := boundFormStatement(t, tc.spelling+" Value", map[string]*record{"Parcel": selected})
			head := statement.tokens[0]
			before, category := head.binding, head.tType
			clause, err := initClauseGram(tc.grammar, statement.tokens, statement.gram)
			if err == nil {
				err = clause.build()
			}
			if err == nil {
				t.Fatal("noun was promoted without a declared applicable form and context")
			}
			if head.binding != before || head.binding.Record != selected || head.binding.Form != tc.role || head.tType != category {
				t.Fatal("rejected clause changed the context-free noun role")
			}
		})
	}
}
