package Lyger

import (
	"errors"
	"strings"
	"testing"

	"github.com/l-programming-language/l/contracts"
	"github.com/l-programming-language/l/lerrors"
)

func TestWordLibraryRejectReturnsFromCaller(t *testing.T) {
	for _, action := range []string{"Storage.Reject Message", "Checks.Refuse Message", "Checks.Finish Message", "Checks.Nested Message"} {
		t.Run(action, func(t *testing.T) {
			source := "Use Storage from `storage`.\nCreate \"Checks\" as a Dictionary.\n" +
				"Define \"Mark|-s|-ed|-ing|-er\" to `<DEF>MOB{|Key: ~_~|}^KY{|Key: ~crate::word_observer::mark~|}<DEF>`.\n" +
				"Define \"Messagevalue\" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~crate::word_observer::message~, Inputs: ~&[<@Value@>]&~|}<DEF>`.\n" +
				"Create Checks's \"Refuse|-s|-d|--ing|--er\" for Storage.Rejecting any Text.\n" +
				"Create Checks's \"Nested\" for Processing a New Process for Storage.Rejecting any Text.\n" +
				"Create Checks's \"Finish|-es|-ed|-ing|-er\" for Function.Returning a New Storage's Rejection of a New Storage's Failure of any Text.\n\n" +
				"    Publicly create Checks's \"Choose\" as a Generic `choose` Function with Inputs of \"fail\" as `bool` and \"message\" as `&str` and Outputs of `runtime::Result<f64>`.\n" +
				"Create \"__l_nested_0\" as a New Storage's Binding of Message.\n" +
				"If Fail, " + strings.Replace(action, "Message", "a New Messagevalue of __l_nested_0", 1) + ".\nMark.\nReturn a New Storage's Success of 42.\n\n" +
				"    Publicly create Checks's \"Construct\" as a Generic `construct` Function with Inputs of \"message\" as `&str` and Outputs of `runtime::Result<f64>`.\nCreate \"Failure\" as a New Storage's Failure of Message.\nCreate \"Rejected\" as a New Storage's Rejection of Failure.\nMark.\nReturn Rejected.\n"
			c := compileDefinitionSource(t, source, "word-reject-caller.l")
			rec, err := c.exportedRecord("Checks.Choose")
			if err != nil || rec.function == nil {
				t.Fatalf("Choose Function: %v", err)
			}
			branch := rec.function.Body[1]
			if branch.Kind != instructionIf || len(branch.Then) != 1 || branch.Then[0].Kind != instructionReturn {
				t.Fatalf("returning action lost typed caller exit: %+v", branch)
			}
			planner := &orchestrationPlanner{semantics: c.Semantics}
			planned, _, err := planner.instructions(branch.Then, newOrchestrationFlow())
			if err != nil || len(planned) != 1 || planned[0].ReachesNext {
				t.Fatalf("Reject branch remains live: %+v, %v", planned, err)
			}
			wordLibraryRuntimeObserver(t, c, "Checks", `use std::sync::atomic::{AtomicUsize, Ordering};
static MARKS: AtomicUsize = AtomicUsize::new(0);
static MESSAGES: AtomicUsize = AtomicUsize::new(0);
pub(crate) fn mark() { MARKS.fetch_add(1, Ordering::SeqCst); }
pub(crate) fn message(value: &str) -> &str { MESSAGES.fetch_add(1, Ordering::SeqCst); value }
pub(crate) fn exercise() {
 assert_eq!(super::choose(true,"rejected").unwrap_err().to_string(),"rejected");
 assert_eq!(MARKS.load(Ordering::SeqCst),0,"Reject continued its caller");
 assert_eq!(MESSAGES.load(Ordering::SeqCst),1,"Reject evaluated its message again");
 assert_eq!(super::choose(false,"unused").unwrap(),42.0);
 assert_eq!(MARKS.load(Ordering::SeqCst),1);
 assert_eq!(MESSAGES.load(Ordering::SeqCst),1);
 assert_eq!(super::construct("constructed").unwrap_err().to_string(),"constructed");
 assert_eq!(MARKS.load(Ordering::SeqCst),2,"constructors returned from their caller");
 println!("word-library error assertions reached");
}`)
		})
	}
}

func TestWordLibraryRootActionUsesSelectedReturn(t *testing.T) {
	source := "Use Storage from `storage`.\nCreate \"Checks\" as a Dictionary.\n" +
		"Define \"Alternate\" as `<DEF>FW{|Func{Key: <@Name|Text@>, Inputs: <@Inputs|Props@>, Outputs: <@Outputs|Texts@>}Func|}^(<@Process@>)<DEF>`.\n" +
		"Define \"Witness\" as `<DEF>HY{|Key: ~ReturnWitness~, Props: ~&[]&~|}<DEF>`.\n" +
		"Define Witness's \"Finish|-es|-ed|-ing|-er\" to `<DEF>BOD^E{|Statement: ~return crate::word_observer::selected(<@Statement@>, ReturnWitness {});~|}<DEF>`.\n" +
		"Use Witness.Finish as Alternate's \"Return\".\n" +
		"Use Function's Invoke as Alternate's \"Invoke\".\n" +
		"Create Checks's \"Refuse|-s|-d|--ing|--er\" for Storage.Rejecting any Text.\n\n" +
		"    Publicly create Checks's \"Choose\" as a Generic `choose` Alternate with Inputs of \"message\" as `&str` and Outputs of `runtime::Result<f64>`.\nChecks.Refuse Message.\nReturn a New Storage's Success of 42.\n"
	c := compileDefinitionSource(t, source, "word-reject-selected-return.l")
	choose, err := c.exportedRecord("Checks.Choose")
	if err != nil {
		t.Fatal(err)
	}
	emitter := orchestrationEmitter{compilation: c}
	selection, err := emitter.composeAppSelection("Selected", &orchestrationValue{Type: valueApp}, "target", nil, choose.function.Return)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(selection, "*"), "crate::word_observer::selected") {
		t.Fatal("compiler-generated App selection bypassed the selected Return")
	}
	witness := false
	for _, dependency := range emitter.dependencies {
		witness = witness || strings.Contains(dependency.Expression, "HY{|Key: ~ReturnWitness~")
	}
	if !witness {
		t.Fatal("compiler-generated App selection lost the selected Return's declaration dependency")
	}
	wordLibraryRuntimeObserver(t, c, "Checks", `use std::sync::atomic::{AtomicUsize, Ordering};
static RETURNS: AtomicUsize = AtomicUsize::new(0);
pub(crate) fn selected<T>(value: T, _: super::ReturnWitness) -> T { RETURNS.fetch_add(1,Ordering::SeqCst); value }
pub(crate) fn exercise() {
 assert_eq!(super::choose("selected").unwrap_err().to_string(),"selected");
 assert_eq!(RETURNS.load(Ordering::SeqCst),1,"root action bypassed the selected Return");
 println!("word-library error assertions reached");
}`)
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

func TestWordLibraryConditionalReturnKeepsContinuingPath(t *testing.T) {
	source := "Use Storage from `storage`.\n" +
		"Define \"Mark|-s|-ed|-ing|-er\" to `<DEF>MOB{|Key: ~_~|}^KY{|Key: ~crate::mark~|}<DEF>`.\n" +
		"Create \"Maybe\" for Processing a New Storage's Selector of any Truth, a New Process for Function.Returning any Number, and a New Storage.Empty Process.\n\n" +
		"    Publicly create \"Choose\" as a Generic `choose` Function with Inputs of \"fail\" as `bool` and Outputs of `f64`.\nMaybe Fail and 7.\nMark.\nReturn 42.\n"
	c := compileDefinitionSource(t, source, "conditional-return-action.l")
	rec, err := c.exportedRecord("Choose")
	if err != nil || rec.function == nil || rec.function.Body[0].Kind != instructionCall {
		t.Fatalf("conditional returning action became an unconditional exit: %v", err)
	}
	planner := &orchestrationPlanner{semantics: c.Semantics}
	planned, _, err := planner.instructions(rec.function.Body, newOrchestrationFlow())
	if err != nil || len(planned) != 3 || !planned[0].ReachesNext {
		t.Fatalf("conditional continuation was removed: %+v, %v", planned, err)
	}
	testNativeDefinitionLibrary(t, c, "Choose", "conditional_return", `use std::sync::atomic::{AtomicUsize,Ordering};
static MARKS: AtomicUsize=AtomicUsize::new(0);
pub fn mark(){MARKS.fetch_add(1,Ordering::SeqCst);}
#[path="src/conditional_return.rs"] mod generated;
fn main(){ assert_eq!(generated::choose(true),7.0);assert_eq!(MARKS.load(Ordering::SeqCst),0);
assert_eq!(generated::choose(false),42.0);assert_eq!(MARKS.load(Ordering::SeqCst),1); }`)
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

func TestWordLibraryRootReturnActionRuntimeFunction(t *testing.T) {
	source := "Create \"Finish|-es|-ed|-ing|-er\" for Function.Returning any Text.\n\n" +
		functionSource("Echo", "\"Value\" as Text", "Text", "Finish Value.\nShow `unreachable`.") +
		"Echo `kept` as \"Result\".\nShow Result.\n"
	c := compileMemberResolutionSource(t, memberResolutionSource{main: source})
	if c.Error != nil {
		t.Fatal(c.Error)
	}
	captureNativeResult(t, c, "kept")
}

func TestWordLibraryRejectContractsRetainSelectedApplication(t *testing.T) {
	source := "Use Storage from `storage`.\n\n    Publicly create \"Rejecting\" as a Generic Function with Inputs of \"message\" as `&str` and Outputs of `runtime::Result<f64>`.\nStorage.Reject Message.\n"
	analysis := analyzeContractSource(t, NewLygerService(canonicalTestOptions()), []byte(source), "reject-contracts.l", ContractAnalysisOptions{})
	c := analysis.Compilation.Compilation
	rec, err := c.exportedRecord("Rejecting")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.composeExport(rec); err != nil {
		t.Fatal(err)
	}
	planner := &orchestrationPlanner{semantics: c.Semantics}
	planned, _, err := planner.instructions(rec.function.Body, newOrchestrationFlow())
	if err != nil {
		t.Fatal(err)
	}
	recorder := c.contractRecorder
	firstUse, firstScope := len(recorder.nativeAnalysis.uses), len(recorder.nativeAnalysis.scopes)
	projected := contracts.Node{Kind: contracts.Sequence, Children: recorder.contractInstructions(planned)}
	var order []contracts.NodeKind
	contractWalk(projected, func(node contracts.Node) {
		if node.Kind == contracts.Native || node.Kind == contracts.Return {
			order = append(order, node.Kind)
		}
	})
	if len(order) != 2 || order[0] != contracts.Native || order[1] != contracts.Return {
		t.Fatalf("return erased or reordered its selected native application: %v", order)
	}
	uses := recorder.nativeAnalysis.uses[firstUse:]
	if len(uses) != 1 || uses[0].Name != "Storage.Reject" || uses[0].Implementation == "" || len(uses[0].Dependencies) == 0 {
		t.Fatalf("return lost selected implementation provenance: %+v", uses)
	}
	foundScope := false
	for _, scope := range recorder.nativeAnalysis.scopes[firstScope:] {
		foundScope = foundScope || scope.Boundary == "source_operation" && len(scope.NativeUses) == 1 && scope.NativeUses[0] == uses[0].ID
	}
	if !foundScope {
		t.Fatal("return lost its source-operation scope")
	}
	summary, err := contracts.Infer(contracts.Plan{Entry: projected}, contracts.AnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, profile := range []string{"NoHeapAllocation", "NoPanic", "NoBlocking", "NoExternalIO"} {
		if contracts.AllEstablished(contracts.CheckProfile(summary, contracts.Profile{Name: profile})) {
			t.Fatalf("return silently established %s despite unresolved native behavior", profile)
		}
	}
}
