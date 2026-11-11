package Lyger

import (
	"errors"
	"strings"
	"testing"

	"github.com/l-programming-language/l/lerrors"
)

// Grammar-only tests have no published records; use the same direct site bodies.
func directiveClauses(statement *lStatement) []*clauseGram {
	sites, err := statement.sGram.registrationSites()
	if err != nil {
		panic(err)
	}
	var body []*clauseGram
	for _, site := range sites {
		body = append(body, site.actionClauses()...)
	}
	return body
}

func TestWordLibraryComposedReturnDelegation(t *testing.T) {
	for name, delegate := range map[string]string{"direct": "Function.Returning", "source helper": "Finishing"} {
		t.Run(name, func(t *testing.T) {
			source := "Define \"Alternate\" as `<DEF>FW{|Func{Key: <@Name|Text@>, Inputs: <@Inputs|Props@>, Outputs: <@Outputs|Texts@>}Func|}^(<@Process@>)<DEF>`.\n" +
				"Define \"Witness\" as `<DEF>KY{|Key: ~crate::selected~, Inputs: ~&[<@Value@>]&~|}<DEF>`.\n" +
				"Create \"Finish|-es|-ed|-ing|-er\" for Function.Returning any Number.\n" +
				"Create Alternate's \"Return|-s|-ed|-ing|-er\" for " + delegate + " a New Witness of any Number.\n" +
				"Use Function.Invoke as Alternate's \"Invoke\".\n\n" +
				"    Publicly create \"Choose\" as a Generic `choose` Alternate with No Inputs and Outputs of `f64`.\nReturn 7.\n"
			c := compileDefinitionSource(t, source, "composed-return-delegation.l")
			if _, err := c.ComposeDefinition("Choose"); err != nil {
				t.Fatalf("selected Return could not delegate to the core Return: %v", err)
			}
			testNativeDefinitionLibrary(t, c, "Choose", "return_delegation", `use std::sync::atomic::{AtomicUsize,Ordering}; static RETURNS:AtomicUsize=AtomicUsize::new(0); pub fn selected(value:f64)->f64{RETURNS.fetch_add(1,Ordering::SeqCst);value} #[path="src/return_delegation.rs"] mod generated; fn main(){assert_eq!(generated::choose(),7.0);assert_eq!(RETURNS.load(Ordering::SeqCst),1);}`)
		})
	}
}

func TestWordLibraryRecursiveReturnRejectsUnresolvedMember(t *testing.T) {
	source := "Define \"Alternate\" as `<DEF>FW{|Func{Key: <@Name|Text@>, Inputs: <@Inputs|Props@>, Outputs: <@Outputs|Texts@>}Func|}^(<@Process@>)<DEF>`.\n" +
		"Create \"Repeat|-s|-ed|-ing|-er\" for Repeating any Number.\nCreate Alternate's \"Return|-s|-ed|-ing|-er\" for Repeating any Number.\nUse Function.Invoke as Alternate's \"Invoke\".\n\n" +
		"    Publicly create \"Choose\" as a Generic `choose` Alternate with No Inputs and Outputs of `f64`.\nReturn 7.\n"
	c := NewLygerService(canonicalTestOptions()).NewCompilation([]byte(source), "composed-return-cycle.l")
	c.Make()
	if !errors.Is(c.Error, lerrors.UnresolvedReference) || !strings.Contains(c.Error.Error(), "Function constructor has no Return definition") {
		t.Fatalf("recursive Return member must remain unresolved: %v", c.Error)
	}
}

func TestWordLibraryFunctionActorKeepsCallerLive(t *testing.T) {
	source := "Define \"Mark|-s|-ed|-ing|-er\" to `<DEF>MOB{|Key: ~_~|}^KY{|Key: ~crate::mark~|}<DEF>`.\n\n" +
		"    Publicly create \"Finish|-es|-ed|-ing|-er\" as a Generic `finish` Function with Inputs of \"value\" as `f64` and Outputs of `f64`.\nReturn Value.\n\n" +
		"Create \"Forward|-s|-ed|-ing|-er\" for Processing a New Process for Creating \"Finished\" as a New Finisher of any Number.\n\n" +
		"    Publicly create \"Choose\" as a Generic `choose` Function with No Inputs and Outputs of `f64`.\nForward 7.\nMark.\nReturn 42.\n"
	c := compileDefinitionSource(t, source, "function-actor-boundary.l")
	rec, err := c.exportedRecord("Choose")
	if err != nil {
		t.Fatal(err)
	}
	planner := &orchestrationPlanner{semantics: c.Semantics}
	planned, _, err := planner.instructions(rec.function.Body, newOrchestrationFlow())
	if err != nil || rec.function.Body[0].Kind != instructionCall || len(planned) != 3 || !planned[0].ReachesNext {
		t.Fatalf("Function actor terminated its caller: %+v, %v", planned, err)
	}
	testNativeDefinitionLibrary(t, c, "Choose", "function_actor", `use std::sync::atomic::{AtomicUsize,Ordering}; static MARKS:AtomicUsize=AtomicUsize::new(0); pub fn mark(){MARKS.fetch_add(1,Ordering::SeqCst);} #[path="src/function_actor.rs"] mod generated; fn main(){assert_eq!(generated::choose(),42.0);assert_eq!(MARKS.load(Ordering::SeqCst),1);}`)
}

func TestWordLibraryCompatibleOwnedReturn(t *testing.T) {
	source := "    Publicly create \"Owner\" as a Generic `owner` Function with Inputs of \"value\" as `f64` and Outputs of `f64`.\nPublicly create \"Finish|-es|-ed|-ing|-er\" for Returning any Number.\nReturn Value.\n\n" +
		"    Publicly create \"Other\" as a Generic `other` Function with No Inputs and Outputs of `f64`.\nFinish 7.\n"
	c := compileDefinitionSource(t, source, "compatible-owned-return.l")
	testNativeDefinitionLibrary(t, c, "Other", "compatible_return", `#[path="src/compatible_return.rs"] mod generated; fn main(){assert_eq!(generated::other(),7.0);}`)
}

func TestWordLibraryActorRejectsForeignReturn(t *testing.T) {
	source := "Define \"Alternate\" as `<DEF>FW{|Func{Key: <@Name|Text@>, Inputs: <@Inputs|Props@>, Outputs: <@Outputs|Texts@>}Func|}^(<@Process@>)<DEF>`.\n" +
		"Use Function.Return as Alternate's \"Return\".\nUse Function.Invoke as Alternate's \"Invoke\".\n\n" +
		"    Publicly create \"Owner\" as a Generic Alternate with Inputs of \"value\" as `f64` and Outputs of `f64`.\nPublicly create \"Finish|-es|-ed|-ing|-er\" for Returning any Number.\nReturn Value.\n\n" +
		"Create \"Forward|-s|-ed|-ing|-er\" for Processing a New Finisher of any Number.\n\n" +
		"    Publicly create \"Other\" as a Generic Function with Inputs of \"value\" as `f64` and Outputs of `f64`.\nForward Value.\n"
	c := NewLygerService(canonicalTestOptions()).NewCompilation([]byte(source), "foreign-actor-return.l")
	c.Make()
	if !errors.Is(c.Error, lerrors.TypeMismatch) || !strings.Contains(c.Error.Error(), "Return belongs to a different Function constructor contract") {
		t.Fatalf("actor bypassed its definition-site Return contract: %v", c.Error)
	}
}

func TestWordLibraryExactSourceFormals(t *testing.T) {
	for name, test := range map[string]struct{ source, call string }{
		"two declarations":  {"Create \"First\" for Function.Returning any Text and then Create \"Second\" for Function.Returning any Number.\n\n    Publicly create \"Choose\" as a Generic `choose` Function with No Inputs and Outputs of `f64`.\nSecond 7.\n", "generated::choose()"},
		"header and action": {"    Publicly create \"Choose\" as a Generic `choose` Function with Inputs of \"value\" as `f64` and Outputs of `f64` and then Create \"Finish|-es|-ed|-ing|-er\" for Function.Returning any Number.\nFinish Value.\n", "generated::choose(7.0)"},
	} {
		t.Run(name, func(t *testing.T) {
			c := compileDefinitionSource(t, test.source, "exact-source-formals.l")
			if _, err := c.ComposeDefinition("Choose"); err != nil {
				t.Fatal(err)
			}
			testNativeDefinitionLibrary(t, c, "Choose", "source_formals", `#[path="src/source_formals.rs"] mod generated; fn main(){assert_eq!(`+test.call+`,7.0);}`)
		})
	}
}

func TestWordLibraryNativeActionRejectsForeignReturn(t *testing.T) {
	for name, returned := range map[string]string{
		"distinct Return":     "Define any Alternate's \"Return|-s|-ed|-ing|-er\" to `<DEF>BOD^E{|Statement: ~return <@Statement@>;~|}<DEF>`.\n",
		"shared Return alias": "Use Function.Return as Alternate's \"Return\".\n",
	} {
		t.Run(name, func(t *testing.T) {
			source := "Define \"Alternate\" as `<DEF>FW{|Func{Key: <@Name|Text@>, Inputs: <@Inputs|Props@>, Outputs: <@Outputs|Texts@>}Func|}^(<@Process@>)<DEF>`.\n" + returned +
				"Use Function's Invoke as Alternate's \"Invoke\".\n\n" +
				"    Publicly create \"Owner\" as a Generic Alternate with Inputs of \"value\" as `f64` and Outputs of `f64`.\nPublicly create \"Finish|-es|-ed|-ing|-er\" for Returning any Thing.\nReturn Value.\n\n" +
				"    Publicly create \"Other\" as a Generic Function with Inputs of \"value\" as `f64` and Outputs of `f64`.\nFinish Value.\n"
			c := NewLygerService(canonicalTestOptions()).NewCompilation([]byte(source), "foreign-native-return.l")
			c.Make()
			if !errors.Is(c.Error, lerrors.TypeMismatch) || !strings.Contains(c.Error.Error(), "Return belongs to a different Function constructor contract") {
				t.Fatalf("foreign authored Return was rebound: %v", c.Error)
			}
		})
	}
}

func TestWordLibraryConditionalConstructorsKeepContinuingPath(t *testing.T) {
	for _, test := range []struct{ name, definition, action string }{
		{"guarded object", "Define \"Guarded\" as `<DEF>BOG^GI{|Key: ~false~|}^(E{|Statement: ~true~|}*W^(<@Body@>)*E{|Statement: ~false~|}*W^BOD^E{|Statement: ~()~|})<DEF>`.\n", "Processing a New Guarded for Function.Returning any Number"},
		{"adjectived Process", "Define \"Guarded\" as a New Adjective describing any Process.\nIt describes any Process as `<DEF>BOG^GI{|Key: ~false~|}^(E{|Statement: ~true~|}*<@Process@>*E{|Statement: ~false~|}*W^BOD^E{|Statement: ~()~|})<DEF>`.\n", "Processing a New Guarded Process for Function.Returning any Number"},
	} {
		t.Run(test.name, func(t *testing.T) {
			source := test.definition + "Create \"Maybe\" for " + test.action + ".\n\n" +
				"    Publicly create \"Choose\" as a Generic `choose` Function with No Inputs and Outputs of `f64`.\nMaybe 7.\nReturn 42.\n"
			c := compileDefinitionSource(t, source, "conditional-constructor-return.l")
			rec, err := c.exportedRecord("Choose")
			if err != nil || rec.function.Body[0].Kind != instructionCall {
				t.Fatalf("conditional constructor became an unconditional exit: %v", err)
			}
			generated := wordLibraryGeneratedSource(t, c, "Choose")
			if !strings.Contains(generated, "return 42") {
				t.Fatalf("conditional constructor pruned continuing Return: %s", generated)
			}
			testNativeDefinitionLibrary(t, c, "Choose", "conditional_constructor", `#[path="src/conditional_constructor.rs"] mod generated; fn main() { assert_eq!(generated::choose(),42.0); }`)
		})
	}
}

func TestWordLibraryConditionalWordTemplatePreservesReturn(t *testing.T) {
	// This opaque W wrapper is a frontend classification control. Its nested W
	// category is not a supported native Function-body sequence.
	source := "Define \"Guard|-s|-ed|-ing|-er\" to `<DEF>W^(BOG^GI{|Key: ~false~|}^(E{|Statement: ~true~|}*<@Definitions|Syllables@>*E{|Statement: ~false~|}*W^BOD^E{|Statement: ~()~|}))<DEF>`.\n" +
		"Create \"Maybe\" for Guarding a New Process for Function.Returning any Number.\n\n" +
		"    Publicly create \"Choose\" as a Generic `choose` Function with No Inputs and Outputs of `f64`.\nMaybe 7.\nReturn 42.\n"
	c := compileDefinitionSource(t, source, "conditional-word-return.l")
	rec, err := c.exportedRecord("Choose")
	if err != nil || rec.function.Body[0].Kind != instructionCall {
		t.Fatalf("conditional W prefix became an unconditional exit: %v", err)
	}
	composed, err := c.ComposeDefinition("Choose")
	if err != nil || !strings.Contains(composed.Expression, "return 42_f64;") {
		t.Fatalf("conditional W prefix pruned continuing Return: %s, %v", composed.Expression, err)
	}
}
