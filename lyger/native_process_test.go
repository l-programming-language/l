package Lyger

import (
	"path/filepath"
	"strings"
	"testing"

	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	Logiceme "github.com/l-programming-language/l/logiceme"
)

const processPatternSource = "Define \"Dispatch|-es|-ed|-ing|-er\" to `<DEF>BOG^GI{|Key: ~<@Value@>~|}^(<@Payloadcase@>*<@Vacancy@>)<DEF>`.\n" +
	"Define \"Payloadcase|-s|-like|-ing\" as `<DEF>E{|Statement: ~Some(<@Payload|Label@>)~|}*(<@Body@>)<DEF>`.\n" +
	"Define \"Vacancy|--ies|---t|----ting\" as `<DEF>E{|Statement: ~None~|}*(<@Body@>)<DEF>`.\n"

func typedProcessSource(nativeType, signature, body string) string {
	if signature != "" {
		signature = " with " + signature
	}
	return processPatternSource +
		"Define \"Fallback|-s|-like|-ing\" as `<DEF>KY{|Key: ~crate::fallback~|}<DEF>`.\n\n" +
		"    Publicly create \"Exercise|-s|-d|--ing|-r\" as a Generic `exercise` Function with Inputs of \"value\" as `Option<" + nativeType + ">` and Outputs of `" + nativeType + "`.\n" +
		"Dispatch Value, a New Payloadcase using a New Process" + signature + " for " + body + ", and a New Vacancy using a New Process for Returning a New Fallback.\n"
}

func TestTypedProcessInputComposition(t *testing.T) {
	source := typedProcessSource("String", "Inputs of \"payload\" as `String`", "Returning Payload")
	c := compileDefinitionSource(t, source, "typed-process.l")
	generated := generatedDefinitionSource(t, c, "Exercise")
	if !strings.Contains(generated, ": String;") {
		t.Fatalf("Process input annotation disappeared: %s", generated)
	}
	testNativeDefinitionLibrary(t, c, "Exercise", "typed_process", `pub fn fallback()->String{String::new()}#[path="src/typed_process.rs"]mod generated;fn main(){assert_eq!(generated::exercise(Some(String::from("selected"))),"selected");assert_eq!(generated::exercise(None),"");}`)
}

func TestTypedProcessInputMovesOnlySelectedPayload(t *testing.T) {
	source := typedProcessSource("crate::Probe", "Inputs of \"payload\" as `crate::Probe`", "Returning Payload")
	c := compileDefinitionSource(t, source, "process-ownership.l")
	testNativeDefinitionLibrary(t, c, "Exercise", "process_ownership", `
use std::sync::atomic::{AtomicUsize,Ordering};
static DROPS:AtomicUsize=AtomicUsize::new(0);static FALLBACKS:AtomicUsize=AtomicUsize::new(0);
pub struct Probe(&'static str);impl Drop for Probe{fn drop(&mut self){DROPS.fetch_add(1,Ordering::SeqCst);}}
pub fn fallback()->Probe{FALLBACKS.fetch_add(1,Ordering::SeqCst);Probe("fallback")}
#[path="src/process_ownership.rs"]mod generated;
fn main(){let selected=generated::exercise(Some(Probe("payload")));assert_eq!(selected.0,"payload");assert_eq!(FALLBACKS.load(Ordering::SeqCst),0);assert_eq!(DROPS.load(Ordering::SeqCst),0);drop(selected);assert_eq!(DROPS.load(Ordering::SeqCst),1);let vacant=generated::exercise(None);assert_eq!(vacant.0,"fallback");assert_eq!(FALLBACKS.load(Ordering::SeqCst),1);drop(vacant);assert_eq!(DROPS.load(Ordering::SeqCst),2);}`)
}

func TestTypedProcessInputRejectsMismatchedNativeType(t *testing.T) {
	source := typedProcessSource("String", "Inputs of \"payload\" as `f64`", "Returning Payload")
	requireProcessNativeRejection(t, source, "expected `f64`, found `String`")
}

func TestTypedProcessInputRejectsDoubleMove(t *testing.T) {
	source := "Define \"Consume|-s|-d|--ing|--er\" to `<DEF>MOB{|Key: ~_~|}^KY{|Key: ~crate::consume~, Inputs: ~&[<@Value@>]&~|}<DEF>`.\n" + typedProcessSource("String", "Inputs of \"payload\" as `String`", "Consuming Payload and then Returning Payload")
	requireProcessNativeRejection(t, source, "use of moved value")
}

func requireProcessNativeRejection(t *testing.T, source, diagnosticFragment string) {
	t.Helper()
	c := compileDefinitionSource(t, source, "process-input.l")
	aml, err := c.ComposeLibrary("Exercise", "process_input")
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := Logiceme.NewLogicemeService(c.Gram.resources).Evaluate(aml)
	if err != nil {
		t.Fatal(err)
	}
	files := make(map[string][]byte)
	for index, name := range bundle.MapArrays[0] {
		files[strings.TrimPrefix(name, "/")] = bundle.ByteArrays[0][index]
	}
	files["observe.rs"] = []byte(`pub fn fallback()->String{String::new()}pub fn consume(_:String){}#[path="src/process_input.rs"]mod generated;fn main(){let _=generated::exercise(None);}`)
	directory := nativetest.Materialize(t, files)
	tools := nativetest.Toolchain(t)
	_, diagnostic, err := nativetest.BuildPrerequisite.Run(t, tools.Rustc, directory, nativetest.CleanEnvironment(tools.Rustc), "--edition=2024", "observe.rs", "-o", filepath.Join(directory, "observe"))
	if err == nil || !strings.Contains(string(diagnostic), diagnosticFragment) {
		t.Fatalf("typed Process lost native checking: %v: %s", err, diagnostic)
	}
}

func TestStructuredProcessBindingArity(t *testing.T) {
	for _, tc := range []struct{ name, signature, body, diagnostic string }{
		{"missing", "", "Returning a New Fallback", "missing argument for Body"},
		{"extra", "Inputs of \"payload\" as `String` and \"extra\" as `String`", "Returning Payload", "structured body provides 2 inputs for 1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := NewLygerService(canonicalTestOptions()).NewCompilation([]byte(typedProcessSource("String", tc.signature, tc.body)), "process-arity.l")
			c.Make()
			err := c.Error
			if err == nil {
				_, err = c.ComposeLibrary("Exercise", "process_arity")
			}
			if err == nil || !strings.Contains(err.Error(), tc.diagnostic) {
				t.Fatalf("invalid Process binding accepted or wrong diagnostic: %v", err)
			}
		})
	}
}
