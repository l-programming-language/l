package Lyger

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	"github.com/l-programming-language/l/lerrors"
	Logiceme "github.com/l-programming-language/l/logiceme"
)

func TestValueBindingForwardingPreservesCaptureBoundary(t *testing.T) {
	for _, word := range []string{"Capture", "Forwarded", "Routed Carrier", "Nested"} {
		t.Run(word, func(t *testing.T) {
			source := "If false, define \"Capture\" as `<DEF>MOB{|Key: ~<@Held|Label@>~|}^KY{|Key: ~std::convert::identity~, Inputs: ~&[<@Value@>]&~|}<DEF>`.\n" +
				"Define \"Carrier\" as `<DEF>MOB{|Key: ~<@Held|Label@>~|}^KY{|Key: ~std::convert::identity~, Inputs: ~&[<@Value@>]&~|}<DEF>`.\n" +
				"Define \"Routed\" as a New Adjective describing any Carrier.\nIt describes any Carrier as a New Capture.\n" +
				"Define \"Forwarded\" as a New Capture.\nDefine \"Nested\" as a New Carrier of a New Capture.\n" +
				"Create \"Speak|-s|-ed|-ing\" for Processing a New " + word + ".\nCreate \"Worker\" as a New Basic App.\nThe App Speaks 7.\n"
			_, err := distributionFrontend(t, source, nil)
			if !errors.Is(err, lerrors.Unsupported) || !strings.Contains(err.Error(), "runtime value Capture cannot be captured") {
				t.Fatalf("forwarded construction lost canonical capture rejection: %v", err)
			}
		})
	}
}

func TestValueBindingUsesSelectedResultLabel(t *testing.T) {
	files := dictionaryIndexFilesystem(t)
	const core = "builtins/eng/core.l"
	files[core] = &fstest.MapFile{Data: append(append([]byte(nil), files[core].Data...), []byte("\n"+
		"Define \"Carrier\" as `<DEF>OB{|Key: ~<@Held|Label@>~, Value: ~<@Value@>~|}<DEF>`.\n"+
		"Define \"Forwarded\" as a New Carrier.\nUse Forwarded as \"Relayed\".\n"+
		"Define \"Marked\" as a New Adjective describing any Carrier.\n"+
		"It describes any Carrier as `<DEF>OB{|Key: ~<@Written|Label@>~, Value: ~<@Operand@>~|}<DEF>`.\n"+
		"Define \"Parcel\" as a New Marked Carrier.\n"+
		"Define \"Routed\" as a New Adjective describing any Carrier.\nIt describes any Carrier as a New Parcel.\n"+
		"Define \"Vacant\" as `<DEF>OB{|Key: ~<@Receipt|Label@>~, Value: ~()~|}<DEF>`.\nDefine \"EmptyParcel\" as a New Vacant.\n"+
		"Define \"Stages\" as a New Process of a New Carrier and a New Vacant.\n")...)}
	const source = "Define \"Acquire\" as `<DEF>MOB{|Key: ~<@Acquired|Label@>~|}^KY{|Key: ~crate::acquire~|}<DEF>`.\n\n" +
		"    Publicly create \"Exercise\" as a Generic Function with No Inputs and Outputs of `f64`.\nReturn a New Routed Carrier of a New Acquire.\n"
	c := NewLygerService(Options{Language: "eng", Resources: files}).NewCompilation([]byte(source), "result-label.l")
	c.Make()
	if c.Error != nil {
		t.Fatal(c.Error)
	}
	for _, test := range []struct {
		name, word, operand string
		descriptors         []string
	}{
		{name: "ordinary", word: "Carrier", operand: "Value"},
		{name: "descriptor", word: "Carrier", operand: "Operand", descriptors: []string{"Marked"}},
		{name: "forwarded", word: "Forwarded", operand: "Value"},
		{name: "alias", word: "Relayed", operand: "Value"},
		{name: "selected source", word: "Parcel", operand: "Operand"},
		{name: "source descriptor", word: "Carrier", operand: "Operand", descriptors: []string{"Routed"}},
		{name: "zero operands", word: "EmptyParcel"},
	} {
		t.Run(test.name, func(t *testing.T) {
			emitter := orchestrationEmitter{compilation: c}
			for _, input := range []string{"first_operand", "second_operand"} {
				arguments, value := map[string]DefinitionArgument{}, "()"
				if test.operand != "" {
					arguments[test.operand], value = definitionReferenceArgument(input), input
				}
				result, err := emitter.describedValueBinding(test.word, test.descriptors, arguments)
				if err != nil {
					t.Fatal(err)
				}
				if len(emitter.pending) == 0 || strings.Count(emitter.pending[len(emitter.pending)-1], "Key: ~"+result+"~, Value: ~"+value+"~") != 1 {
					t.Fatalf("selected constructor lost explicit result or operand: %v", emitter.pending)
				}
				instance, err := c.instantiateDescribedBuiltin(test.word, test.descriptors, arguments)
				if err != nil || instance.Span.Path != core {
					t.Fatalf("selected constructor lost definition-site provenance: %s, %v", instance.Span.Path, err)
				}
			}
		})
	}
	emitter := orchestrationEmitter{compilation: c}
	if _, err := emitter.valueBinding("Stages", nil); err == nil || !strings.Contains(err.Error(), "requires an explicit result binding") {
		t.Fatalf("multi-stage construction acquired an implicit result: %v", err)
	}
	testNativeDefinitionLibrary(t, c, "Exercise", "forwarded_result", `static CALLS:std::sync::atomic::AtomicUsize=std::sync::atomic::AtomicUsize::new(0);pub fn acquire()->f64{CALLS.fetch_add(1,std::sync::atomic::Ordering::SeqCst);7.0}#[path="src/forwarded_result.rs"]mod generated;fn main(){assert_eq!(generated::exercise(),7.0);assert_eq!(CALLS.load(std::sync::atomic::Ordering::SeqCst),1);}`)
}

func TestValueBindingTransparentSelectedStorage(t *testing.T) {
	files := dictionaryIndexFilesystem(t)
	for _, test := range []struct{ path, declaration, leaf string }{
		{"builtins/eng/storage/execution.l", "Define Storage's \"Literal\" as ", "LiteralAtom"},
		{"builtins/eng/storage/execution.l", "Define Storage's \"Completion\" as ", "CompletionAtom"},
		{"builtins/eng/storage/values.l", "It describes any Storage's Tuple as ", "EmptyTuple"},
	} {
		start := bytes.Index(files[test.path].Data, []byte(test.declaration))
		if start < 0 {
			t.Fatalf("missing selected %s", test.declaration)
		}
		end := start + bytes.IndexByte(files[test.path].Data[start:], '\n')
		leaf := bytes.Replace(files[test.path].Data[start:end], []byte(test.declaration), []byte("Define \""+test.leaf+"\" as "), 1)
		replacement := append([]byte(test.declaration+"a New "+test.leaf+".\n"), leaf...)
		files[test.path] = &fstest.MapFile{Data: bytes.Replace(files[test.path].Data, files[test.path].Data[start:end], replacement, 1)}
	}
	c := NewLygerService(Options{Language: "eng", Resources: files}).NewCompilation([]byte("Define \"Value\" as 7. Show Value.\n"), "transparent-storage.l")
	c.Make()
	if c.Error != nil {
		t.Fatal(c.Error)
	}
	emitter := orchestrationEmitter{compilation: c}
	if _, err := emitter.valueBinding("Storage.Completion", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := emitter.results(nil); err != nil {
		t.Fatal(err)
	}
	if _, err := (&CompilationResult{Compilation: c}).OrchestratorUnit("entry", nil); err != nil {
		t.Fatal(err)
	}
}

func TestValueBindingNativePreservesSelectedBindingError(t *testing.T) {
	c := compileDefinitionSource(t, "Use Storage from `storage`.\n", "binding-error.l")
	produced, err := c.builtinRecord("Storage.Produced")
	if err != nil {
		t.Fatal(err)
	}
	delete(c.Semantics.Descriptors, produced)
	emitter := orchestrationEmitter{compilation: c}
	_, err = emitter.valueBinding("Storage.Value.Copy", map[string]DefinitionArgument{"Value": definitionReferenceArgument("operand")})
	if !errors.Is(err, lerrors.TypeMismatch) || !strings.Contains(err.Error(), "Storage.Produced is not a descriptor") {
		t.Fatalf("native result binding lost its selected descriptor failure: %v", err)
	}
	if len(emitter.pending) != 0 {
		t.Fatalf("failed binding published partial output: %v", emitter.pending)
	}
}

func TestStorageStagesUseExplicitBindings(t *testing.T) {
	c := compileDefinitionSource(t, "Use Storage from `storage`.\n", "explicit-stages.l")
	ref := definitionReferenceArgument
	body := DefinitionInstance{Expression: "W^(BOD^E{|Statement: ~payload~|})"}
	environment := DefinitionInstance{Expression: "KY{|Key: ~make_environment~|}"}
	invocation := DefinitionInstance{Expression: "KY{|Key: ~(closure.1)~, Inputs: ~&[context, environment]&~|}"}
	returned := DefinitionInstance{Expression: "W^(BOD^E{|Statement: ~return status;~|})"}
	for _, test := range []struct {
		name string
		args map[string]DefinitionArgument
		want []string
	}{
		{"Storage.Invoke", map[string]DefinitionArgument{"Name": ref("closure"), "Value": ref("original"), "Captured": ref("environment"), "Copy": ref("copy_environment"), "Context": ref("context"), "SourceContext": ref("caller_context"), "Future": ref("work"), "Invocation": {Kind: "Definition", Definition: &invocation}, "Outcome": ref("attempt"), "Checked": ref("flow"), "Attempt": ref("attempt"), "Selection": ref("result"), "Subject": ref("flow"), "Payload": ref("continued"), "Exit": {Kind: "Definition", Definition: &returned}}, []string{"Key: ~closure~", "Key: ~environment~", "Key: ~context~", "Key: ~work~", "Key: ~attempt~", "Key: ~flow~", "Key: ~result~", "Continue(continued)", "Statement: ~continued~", "return status;"}},
		{"Storage.Function.Suspend", map[string]DefinitionArgument{"Future": ref("work"), "Pinned": ref("boxed"), "Value": ref("work"), "Body": {Kind: "Definition", Definition: &body}}, []string{"Key: ~work~", "Key: ~boxed~", "Statement: ~boxed~"}},
		{"Storage.Function.Bind", map[string]DefinitionArgument{"Capture": ref("attempt"), "Environment": {Kind: "Definition", Definition: &environment}, "Checked": ref("owned"), "Attempt": ref("attempt"), "Converted": ref("pointer"), "Value": ref("function"), "Type": definitionTypeArgument("fn()"), "Name": ref("pair"), "Values": definitionAtomListArgument(ref("owned"), ref("pointer"))}, []string{"Key: ~attempt~", "Key: ~owned~", "Value: ~attempt?~", "Key: ~pointer~", "Key: ~pair~"}},
		{"Storage.Function.Retain", map[string]DefinitionArgument{"Captured": ref("retained"), "Environment": ref("original"), "Copy": ref("copy_environment"), "Converted": ref("pointer"), "Value": ref("function"), "Type": definitionTypeArgument("fn()"), "Name": ref("pair"), "Values": definitionAtomListArgument(ref("retained"), ref("pointer"))}, []string{"Key: ~retained~", "&original", "Key: ~pointer~", "Key: ~pair~"}},
		{"Storage.Context.Create", map[string]DefinitionArgument{"Cloned": ref("retained_runtime"), "Runtime": ref("original_runtime"), "Name": ref("context"), "Values": definitionAtomListArgument(ref("retained_runtime"), ref("program")), "Binding": ref("program_reference")}, []string{"Key: ~retained_runtime~", "Key: ~context~", "Key: ~program_reference~", "Value: ~&context.1~"}},
		{"Storage.Context.Read", map[string]DefinitionArgument{"Context": ref("context"), "RuntimeBinding": ref("host"), "ProgramBinding": ref("program")}, []string{"Key: ~host~", "Value: ~&context.0~", "Key: ~program~", "Value: ~&context.1~"}},
		{"Storage.App.Create", map[string]DefinitionArgument{"Path": {Kind: "Text", Text: "/app"}, "PathBinding": ref("path"), "Runtime": ref("host"), "AppResult": ref("application")}, []string{"Key: ~path~", "Key: ~application~", "host, path"}},
		{"Storage.App.Attribute", map[string]DefinitionArgument{"Source": {Kind: "Text", Text: "authored source"}, "SourceBinding": ref("source"), "App": ref("application")}, []string{"Key: ~source~", "&application, source"}},
		{"Storage.Entry.Observe", map[string]DefinitionArgument{"Main": ref("entry"), "Runtime": ref("host"), "MainFuture": ref("main_work"), "Future": ref("observed_work"), "Outcome": ref("status"), "Return": {Kind: "Definition", Definition: &returned}}, []string{"Key: ~main_work~", "Key: ~entry~", "Key: ~observed_work~", "Key: ~status~", "return status;"}},
		{"Storage.Program.Fail", map[string]DefinitionArgument{"Error": ref("fault"), "Runtime": ref("host"), "FailureInput": ref("input"), "Name": ref("failure"), "Value": ref("fault")}, []string{"Key: ~fault~", "host, input", "Key: ~failure~", "&[fault]&"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			instance, err := c.instantiateBuiltin(test.name, test.args)
			if err != nil {
				t.Fatal(err)
			}
			remaining := instance.Expression
			for _, want := range test.want {
				position := strings.Index(remaining, want)
				if position < 0 {
					t.Fatalf("missing or reordered explicit binding %q: %s", want, instance.Expression)
				}
				remaining = remaining[position+len(want):]
			}
			for _, retired := range []string{"lPending", "lPinned", "lObservedFuture", "lObservedStatus", "lAppSource", "pairEnvironment", "contextRuntime"} {
				if strings.Contains(instance.Expression, retired) {
					t.Errorf("retired implicit binding %s: %s", retired, instance.Expression)
				}
			}
		})
	}
}

func TestStorageStagesDoNotRetainPublicAliases(t *testing.T) {
	c := compileDefinitionSource(t, "Use Storage from `storage`.\n", "retired-stages.l")
	for _, name := range []string{
		"Storage.Closure", "Storage.Closure.Environment", "Storage.Closure.Context", "Storage.Binding.Environment", "Storage.Binding.Context", "Storage.Outcome",
		"Storage.Function.Pending", "Storage.Function.Pinned", "Storage.Function.Ready", "Storage.Function.Capture", "Storage.Function.Pair", "Storage.Function.Snapshot", "Storage.Function.Identity",
		"Storage.Context.Runtime", "Storage.Context.Pair", "Storage.Context.Binding", "Storage.Context.First", "Storage.Context.Second",
		"Storage.App.Path", "Storage.App.Handle", "Storage.App.Source", "Storage.App.Attribution",
		"Storage.Program.Failure", "Storage.Program.Rejection", "Storage.Entry.Future", "Storage.Entry.Outcome",
	} {
		if _, err := c.builtinRecord(name); err == nil {
			t.Errorf("retired public stage remains selectable: %s", name)
		}
	}
}

func TestStorageSuspendAwaitOwnershipNative(t *testing.T) {
	files := dictionaryIndexFilesystem(t)
	for _, probe := range []struct{ path, from, to string }{
		{"builtins/eng/core.l", "return <@Statement@>;", "return crate::observed_return(<@Statement@>);"},
		{"builtins/eng/storage/execution.l", "Key: ~Box::pin~", "Key: ~crate::observed_pin~"},
	} {
		original := files[probe.path].Data
		if bytes.Count(original, []byte(probe.from)) != 1 {
			t.Fatalf("probe does not select exactly one canonical operation in %s", probe.path)
		}
		files[probe.path] = &fstest.MapFile{Data: bytes.Replace(original, []byte(probe.from), []byte(probe.to), 1)}
	}
	c := NewLygerService(Options{Language: "eng", Resources: files}).NewCompilation([]byte("Use Storage from `storage`.\n"), "async-observer.l")
	c.Make()
	if c.Error != nil {
		t.Fatal(c.Error)
	}
	instantiate := func(word string, args map[string]DefinitionArgument) DefinitionInstance {
		t.Helper()
		instance, err := c.instantiateBuiltin(word, args)
		if err != nil {
			t.Fatal(err)
		}
		return instance
	}
	process := func(parts ...DefinitionInstance) DefinitionInstance {
		t.Helper()
		expressions := make([]string, len(parts))
		for index, part := range parts {
			expressions[index] = part.Expression
		}
		return instantiate("Process", map[string]DefinitionArgument{"Definitions": {Kind: "Syllables", Syllables: expressions, sequence: parts}})
	}
	var declarations []DefinitionInstance
	for _, signature := range []struct{ name, input, output string }{
		{"borrowed", "&crate::State", "std::pin::Pin<Box<dyn std::future::Future<Output = u8> + '_>>"},
		{"owned", "crate::Owned", "std::pin::Pin<Box<dyn std::future::Future<Output = String>>>"},
	} {
		awaited := instantiate("Storage.Await", map[string]DefinitionArgument{"Outcome": definitionReferenceArgument("completed"), "Future": definitionReferenceArgument("input")})
		returned := instantiate("Function.Return", map[string]DefinitionArgument{"Statement": definitionReferenceArgument("completed")})
		inner := process(awaited, returned)
		suspended := instantiate("Storage.Function.Suspend", map[string]DefinitionArgument{
			"Future": definitionReferenceArgument("work"), "Pinned": definitionReferenceArgument("pinned"), "Value": definitionReferenceArgument("work"), "Body": {Kind: "Definition", Definition: &inner},
		})
		outer := process(suspended)
		declarations = append(declarations, instantiate("Function", map[string]DefinitionArgument{
			"Name": {Kind: "Text", Text: signature.name}, "Inputs": {Kind: "Props", Props: []DefinitionProperty{{Key: "input", Type: signature.input}}},
			"Outputs": {Kind: "Texts", Texts: []string{signature.output}}, "Process": {Kind: "Definition", Definition: &outer},
		}))
	}
	aml, err := renderLibrary("async_observer", "async_observer", declarations)
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := Logiceme.NewLogicemeService(files).Evaluate(aml)
	if err != nil {
		t.Fatal(err)
	}
	output := make(map[string][]byte)
	for index, path := range bundle.MapArrays[0] {
		output[strings.TrimPrefix(path, "/")] = bundle.ByteArrays[0][index]
	}
	output["observe.rs"] = []byte(`use std::cell::Cell;
use std::future::Future;
use std::pin::Pin;
use std::rc::Rc;
use std::sync::{Arc, Mutex};
use std::task::{Context, Poll, Wake, Waker};
static EVENTS: Mutex<Vec<&'static str>> = Mutex::new(Vec::new());
pub fn observed_pin<T>(value: T) -> Pin<Box<T>> { EVENTS.lock().unwrap().push("pin"); Box::pin(value) }
pub fn observed_return<T>(value: T) -> T { EVENTS.lock().unwrap().push("return"); value }
pub struct State { value: Cell<u8> }
impl Future for &State {
 type Output=u8;
 fn poll(self: Pin<&mut Self>, _: &mut Context<'_>) -> Poll<u8> {
  EVENTS.lock().unwrap().push("borrowed poll"); Poll::Ready(self.value.get())
 }
}
pub struct Owned { value: Option<String>, drops: Rc<Cell<usize>>, ready: bool }
impl Future for Owned {
 type Output=String;
 fn poll(mut self: Pin<&mut Self>, _: &mut Context<'_>) -> Poll<String> {
  EVENTS.lock().unwrap().push("owned poll");
  if !self.ready { self.ready=true; return Poll::Pending; }
  Poll::Ready(self.value.take().unwrap())
 }
}
impl Drop for Owned { fn drop(&mut self) { self.drops.set(self.drops.get()+1); EVENTS.lock().unwrap().push("drop"); } }
struct Noop;
impl Wake for Noop { fn wake(self: Arc<Self>) {} }
fn events() -> Vec<&'static str> { std::mem::take(&mut *EVENTS.lock().unwrap()) }
#[path="src/async_observer.rs"] mod generated;
fn main() {
 let state=State{value:Cell::new(4)};
 let future=generated::borrowed(&state);
 assert_eq!(events(),vec!["pin"],"construction must pin once without executing inner Return");
 drop(future); assert_eq!(events(),Vec::<&str>::new(),"dropping an unpolled borrow must not execute Return");
 let drops=Rc::new(Cell::new(0));
 let future=generated::owned(Owned{value:Some("unpolled".to_owned()),drops:drops.clone(),ready:false});
 assert_eq!(events(),vec!["pin"]); assert_eq!(drops.get(),0);
 drop(future); assert_eq!(events(),vec!["drop"]); assert_eq!(drops.get(),1);
 let waker=Waker::from(Arc::new(Noop)); let mut context=Context::from_waker(&waker);
 let mut future=generated::borrowed(&state);
 assert_eq!(events(),vec!["pin"]); state.value.set(9);
 assert_eq!(future.as_mut().poll(&mut context),Poll::Ready(9));
 assert_eq!(events(),vec!["borrowed poll","return"]); drop(future);
 assert_eq!(events(),Vec::<&str>::new());
 let mut future=generated::owned(Owned{value:Some("owned output".to_owned()),drops:drops.clone(),ready:false});
 assert_eq!(events(),vec!["pin"]);
 assert_eq!(future.as_mut().poll(&mut context),Poll::Pending);
 assert_eq!(events(),vec!["owned poll"]); assert_eq!(drops.get(),1);
 assert_eq!(future.as_mut().poll(&mut context),Poll::Ready("owned output".to_owned()));
 assert_eq!(events(),vec!["owned poll","drop","return"]); assert_eq!(drops.get(),2);
 drop(future); assert_eq!(events(),Vec::<&str>::new()); assert_eq!(drops.get(),2);
}`)
	directory := nativetest.Materialize(t, output)
	tools := nativetest.Toolchain(t)
	binary := filepath.Join(directory, "observe")
	if _, diagnostic, err := nativetest.BuildPrerequisite.Run(t, tools.Rustc, directory, nativetest.CleanEnvironment(tools.Rustc), "--edition=2024", "observe.rs", "-o", binary); err != nil {
		t.Fatalf("native async observer: %v: %s", err, diagnostic)
	}
	if _, diagnostic, err := nativetest.RuntimeExecution.Run(t, binary, directory, nil); err != nil {
		t.Fatalf("async ownership and return boundary: %v: %s", err, diagnostic)
	}
}
