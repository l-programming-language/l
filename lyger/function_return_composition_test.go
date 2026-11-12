package Lyger

import (
	"errors"
	"strings"
	"testing"
	"testing/fstest"

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

func TestRecursiveReturnRejectsUnresolvedMember(t *testing.T) {
	source := "Define \"Alternate|-s|-ive|--ing\" as `<DEF>FW{|Func{Key: <@Name|Text@>, Inputs: <@Inputs|Props@>, Outputs: <@Outputs|Texts@>}Func|}^(<@Process@>)<DEF>`.\n" +
		"Create \"Repeat|-s|-ed|-ing|-er\" for Repeating any Number.\nCreate Alternate's \"Return|-s|-ed|-ing|-er\" for Repeating any Number.\n\n" +
		"    Publicly create \"Choose|-s|----se|--ing|-r\" as a Generic `choose` Alternate with No Inputs and Outputs of `f64`.\nReturn 7.\n"
	c := NewLygerService(canonicalTestOptions()).NewCompilation([]byte(source), "composed-return-cycle.l")
	c.Make()
	if !errors.Is(c.Error, lerrors.UnresolvedReference) || !strings.Contains(c.Error.Error(), "Function constructor has no Return definition") {
		t.Fatalf("recursive Return member must remain unresolved: %v", c.Error)
	}
}

func TestFunctionActorKeepsCallerLive(t *testing.T) {
	source := "Define \"Mark|-s|-ed|-ing|-er\" to `<DEF>MOB{|Key: ~_~|}^KY{|Key: ~crate::mark~|}<DEF>`.\n\n" +
		"    Publicly create \"Finish|-es|-ed|-ing|-er\" as a Generic `finish` Function with Inputs of \"value\" as `f64` and Outputs of `f64`.\nReturn Value.\n\n" +
		"Create \"Forward|-s|-ed|-ing|-er\" for Processing a New Process for Creating \"Finished\" as a New Finisher of any Number.\n\n" +
		"    Publicly create \"Choose|-s|----se|--ing|-r\" as a Generic `choose` Function with No Inputs and Outputs of `f64`.\nForward 7.\nMark.\nReturn 42.\n"
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

func TestCompatibleOwnedReturn(t *testing.T) {
	source := "    Publicly create \"Own|-s|-ed|-ing|-er\" as a Generic `owner` Function with Inputs of \"value\" as `f64` and Outputs of `f64`.\nPublicly create \"Finish|-es|-ed|-ing|-er\" for Returning any Number.\nReturn Value.\n\n" +
		"    Publicly create \"Respond|-s|-ed|-ing|-er\" as a Generic `other` Function with No Inputs and Outputs of `f64`.\nFinish 7.\n"
	c := compileDefinitionSource(t, source, "compatible-owned-return.l")
	testNativeDefinitionLibrary(t, c, "Respond", "compatible_return", `#[path="src/compatible_return.rs"] mod generated; fn main(){assert_eq!(generated::other(),7.0);}`)
}

func TestActorRejectsForeignReturn(t *testing.T) {
	source := "Define \"Alternate|-s|-ive|--ing\" as `<DEF>FW{|Func{Key: <@Name|Text@>, Inputs: <@Inputs|Props@>, Outputs: <@Outputs|Texts@>}Func|}^(<@Process@>)<DEF>`.\n" +
		"Use Function's Return as Alternate's \"Return|-s|-ed|-ing|-er\".\n\n" +
		"    Publicly create \"Own|-s|-ed|-ing|-er\" as a Generic Alternate with Inputs of \"value\" as `f64` and Outputs of `f64`.\nPublicly create \"Finish|-es|-ed|-ing|-er\" for Returning any Number.\nReturn Value.\n\n" +
		"Create \"Forward|-s|-ed|-ing|-er\" for Processing a New Finisher of any Number.\n\n" +
		"    Publicly create \"Respond|-s|-ed|-ing|-er\" as a Generic Function with Inputs of \"value\" as `f64` and Outputs of `f64`.\nForward Value.\n"
	c := NewLygerService(canonicalTestOptions()).NewCompilation([]byte(source), "foreign-actor-return.l")
	c.Make()
	if !errors.Is(c.Error, lerrors.TypeMismatch) || !strings.Contains(c.Error.Error(), "Return belongs to a different Function constructor contract") {
		t.Fatalf("actor bypassed its definition-site Return contract: %v", c.Error)
	}
}

func TestExactSourceFormals(t *testing.T) {
	for name, test := range map[string]struct{ source, call string }{
		"two declarations":  {"Create \"ReturnText|-s|-ed|-ing|-er\" for Function's Returning any Text and then Create \"ReturnNumber|-s|-ed|-ing|-er\" for Function's Returning any Number.\n\n    Publicly create \"Choose|-s|----se|--ing|-r\" as a Generic `choose` Function with No Inputs and Outputs of `f64`.\nReturnNumber 7.\n", "generated::choose()"},
		"header and action": {"    Publicly create \"Choose|-s|----se|--ing|-r\" as a Generic `choose` Function with Inputs of \"value\" as `f64` and Outputs of `f64` and then Create \"Finish|-es|-ed|-ing|-er\" for Function's Returning any Number.\nFinish Value.\n", "generated::choose(7.0)"},
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

func TestNativeActionRejectsForeignReturn(t *testing.T) {
	for name, returned := range map[string]string{
		"distinct Return":     "Define any Alternate's \"Return|-s|-ed|-ing|-er\" to `<DEF>BOD^E{|Statement: ~return <@Statement@>;~|}<DEF>`.\n",
		"shared Return alias": "Use Function's Return as Alternate's \"Return|-s|-ed|-ing|-er\".\n",
	} {
		t.Run(name, func(t *testing.T) {
			source := "Define \"Alternate|-s|-ive|--ing\" as `<DEF>FW{|Func{Key: <@Name|Text@>, Inputs: <@Inputs|Props@>, Outputs: <@Outputs|Texts@>}Func|}^(<@Process@>)<DEF>`.\n" + returned + "\n" +
				"    Publicly create \"Own|-s|-ed|-ing|-er\" as a Generic Alternate with Inputs of \"value\" as `f64` and Outputs of `f64`.\nPublicly create \"Finish|-es|-ed|-ing|-er\" for Returning any Thing.\nReturn Value.\n\n" +
				"    Publicly create \"Respond|-s|-ed|-ing|-er\" as a Generic Function with Inputs of \"value\" as `f64` and Outputs of `f64`.\nFinish Value.\n"
			c := NewLygerService(canonicalTestOptions()).NewCompilation([]byte(source), "foreign-native-return.l")
			c.Make()
			if !errors.Is(c.Error, lerrors.TypeMismatch) || !strings.Contains(c.Error.Error(), "Return belongs to a different Function constructor contract") {
				t.Fatalf("foreign authored Return was rebound: %v", c.Error)
			}
		})
	}
}

func TestComposedReturnDelegation(t *testing.T) {
	for name, delegate := range map[string]string{"direct": "Function's Returning", "source helper": "Finishing"} {
		t.Run(name, func(t *testing.T) {
			source := "Define \"Alternate|-s|-ive|--ing\" as `<DEF>FW{|Func{Key: <@Name|Text@>, Inputs: <@Inputs|Props@>, Outputs: <@Outputs|Texts@>}Func|}^(<@Process@>)<DEF>`.\n" +
				"Define \"Witness|-es|-like|-ing\" as `<DEF>KY{|Key: ~crate::selected~, Inputs: ~&[<@Value@>]&~|}<DEF>`.\n" +
				"Create \"Finish|-es|-ed|-ing|-er\" for Function's Returning any Number.\n" +
				"Create Alternate's \"Return|-s|-ed|-ing|-er\" for " + delegate + " a New Witness of any Number.\n\n" +
				"    Publicly create \"Choose|-s|----se|--ing|-r\" as a Generic `choose` Alternate with No Inputs and Outputs of `f64`.\nReturn 7.\n"
			c := compileDefinitionSource(t, source, "composed-return-delegation.l")
			if _, err := c.ComposeDefinition("Choose"); err != nil {
				t.Fatalf("selected Return could not delegate to the core Return: %v", err)
			}
			testNativeDefinitionLibrary(t, c, "Choose", "return_delegation", `use std::sync::atomic::{AtomicUsize,Ordering}; static RETURNS:AtomicUsize=AtomicUsize::new(0); pub fn selected(value:f64)->f64{RETURNS.fetch_add(1,Ordering::SeqCst);value} #[path="src/return_delegation.rs"] mod generated; fn main(){assert_eq!(generated::choose(),7.0);assert_eq!(RETURNS.load(Ordering::SeqCst),1);}`)
		})
	}
}

func TestFunctionInvocationUsesEnclosingReturnComposition(t *testing.T) {
	files := dictionaryIndexFilesystem(t)
	const core = "builtins/eng/core.l"
	const invocation = "Key: ~(<@Target|Callable@>)~"
	if strings.Count(string(files[core].Data), invocation) != 1 {
		t.Fatal("invocation fixture requires the selected core Invoke definition")
	}
	files[core] = &fstest.MapFile{Data: []byte(strings.Replace(string(files[core].Data), invocation, "Key: ~(/* selected invocation */ <@Target|Callable@>)~", 1))}
	source := "Define \"Alternate|-s|-ive|--ing\" as `<DEF>FW{|Func{Key: <@Name|Text@>, Inputs: <@Inputs|Props@>, Outputs: <@Outputs|Texts@>}Func|}^(<@Process@>)<DEF>`.\n" +
		"Define any Alternate's \"Return|-s|-ed|-ing|-er\" to `<DEF>BOD^E{|Statement: ~return /* callee */ <@Statement@>;~|}<DEF>`.\n\n" +
		"    Publicly create \"Finish|-es|-ed|-ing|-er\" as a Generic `supply` Alternate with Inputs of \"value\" as `f64` and Outputs of `f64`.\nReturn Value.\n\n" +
		"    Publicly create \"Call|-s|-ed|-ing|-er\" as a Generic `caller` Function with No Inputs and Outputs of `f64`.\nReturn a New Finisher of 7.\n"
	c := NewLygerService(Options{Resources: files}).NewCompilation([]byte(source), "enclosing-return.l")
	c.Make()
	if c.Error != nil {
		t.Fatal(c.Error)
	}
	caller, err := c.ComposeDefinition("Call")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(caller.Expression, "/* selected invocation */") || !strings.Contains(caller.Expression, "return ") || strings.Contains(caller.Expression, "/* callee */") {
		t.Fatalf("callee Return replaced its caller's Return: %s", caller.Expression)
	}
	testNativeDefinitionLibrary(t, c, "Call", "enclosing_return", `#[path="src/enclosing_return.rs"]mod generated;fn main(){assert_eq!(generated::caller(),7.0);}`)
}
