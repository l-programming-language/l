package Lyger

import (
	"bytes"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"
)

func TestRuntimeWordCleanBreakRetiredRoutes(t *testing.T) {
	c := compileDefinitionSource(t, "Use Runtime from `runtime`.\n", "runtime-word-routes.l")
	for _, name := range []string{
		"Runtime.Error.Seen", "Runtime.Error.Extend", "Runtime.Byte", "Runtime.Error.Value", "Runtime.Value",
		"Runtime.Future", "Runtime.Pin", "Runtime.Completion", "Runtime.App.Insert", "Runtime.Await",
		"Runtime.Task", "Runtime.Error.Context", "Runtime.Error.Parent", "Runtime.Error.Message", "Runtime.Error.Causes",
		"Runtime.Error.Buffer", "Runtime.Faults", "Runtime.Error.Cancellation", "Runtime.Error.Flag", "Runtime.Error.Mark",
		"Runtime.Error.Separate", "Runtime.App.Processes", "Runtime.Supervision.Owned", "Runtime.Capture.Borrow",
		"Runtime.Error.First", "Runtime.Error.Lossy", "Runtime.Supervision.Extend", "Runtime.Path.Empty", "Runtime.Canonical",
		"Runtime.Capture.Sleep", "Runtime.Capture.Wait", "Runtime.Capture.Timeout", "Runtime.App.Input", "Runtime.App.Stdout",
		"Runtime.App.Stderr", "Runtime.App.Executable", "Runtime.Shared",
	} {
		t.Run(name, func(t *testing.T) {
			if rec, err := c.builtinRecord(name); err == nil && rec != nil {
				t.Fatalf("retired qualified route %s still selects %s", name, rec.ID)
			}
		})
	}
	for _, name := range []string{
		"Runtime.State.Processes", "Runtime.Path.Name", "Runtime.Acquire", "Runtime.Race", "Runtime.Replace",
		"Runtime.Error.Data.Content", "Runtime.App.State.Content", "Runtime.Invocation.State.Content",
		"Runtime.Supervision.Stdout", "Runtime.Supervision.Stderr", "Runtime.Host.Open", "Runtime.Host.Run",
	} {
		t.Run("retained-"+name, func(t *testing.T) {
			if rec, err := c.builtinRecord(name); err != nil || rec == nil {
				t.Fatalf("required Runtime owner %s: %v", name, err)
			}
		})
	}
	for _, name := range []string{
		"Runtime.Pinner", "Runtime.Pinning", "Runtime.Pinned", "Runtime.App.Inserter", "Runtime.App.Inserting",
		"Runtime.Awaiter", "Runtime.Awaiting", "Runtime.Error.Extender", "Runtime.Error.Extending",
		"Runtime.Error.Marker", "Runtime.Error.Marking", "Runtime.Error.Separator", "Runtime.Error.Separating",
		"Runtime.Supervision.Extender", "Runtime.Supervision.Extending",
		"Runtime.App.Buiilt", "Runtime.App.Ruan", "Runtime.App.Sett", "Runtime.Invocation.Readd",
	} {
		t.Run("retired-form-"+name, func(t *testing.T) {
			if rec, err := c.builtinRecord(name); err == nil && rec != nil {
				t.Fatalf("retired family or malformed form %s still resolves", name)
			}
		})
	}
}

// Change only the selected canonical operation in a resource provider. These
// probes distinguish source reuse from a descriptor or helper copying its body.
func TestRuntimeWordCleanBreakSelectedStorageOperations(t *testing.T) {
	for _, tc := range []struct {
		path, declaration, operation, consumer string
		count                                  int
	}{
		{"builtins/eng/storage/collections.l", `Define Storage's "Remember|`, "insert", "Runtime.Error.Remember", 1},
		{"builtins/eng/storage/display.l", `Define Storage's Bytes's "Extend|`, "extend_from_slice", "Runtime.Error.Attach", 2},
		{"builtins/eng/storage/execution.l", `Define Storage's "Pin|`, "Box::pin", "Runtime.Open", 2},
		{"builtins/eng/storage/execution.l", `Define Storage's "Pin|`, "Box::pin", "Runtime.Show", 1},
	} {
		t.Run(tc.consumer, func(t *testing.T) {
			files := dictionaryIndexFilesystem(t)
			file := files[tc.path]
			if file == nil {
				t.Fatalf("missing canonical resource %s", tc.path)
			}
			lines := bytes.Split(file.Data, []byte{'\n'})
			changed := 0
			const marker = "/* selected shared operation */"
			for i, line := range lines {
				if !bytes.HasPrefix(line, []byte(tc.declaration)) {
					continue
				}
				from := []byte("Key: ~" + tc.operation + "~")
				if bytes.Count(line, from) != 1 {
					t.Fatal("canonical leaf must contain its operation once")
				}
				lines[i] = bytes.Replace(line, from, []byte("Key: ~"+tc.operation+" "+marker+"~"), 1)
				changed++
			}
			if changed != 1 {
				t.Fatalf("canonical definitions changed = %d, want one", changed)
			}
			files[tc.path] = &fstest.MapFile{Data: bytes.Join(lines, []byte{'\n'})}
			c := NewLygerService(Options{Resources: files}).NewCompilation([]byte("Use Runtime from `runtime`.\n"), "runtime-selected-storage.l")
			c.Make()
			if c.Error != nil {
				t.Fatal(c.Error)
			}
			definition, err := c.ComposeDefinition(tc.consumer)
			if err != nil {
				t.Fatal(err)
			}
			if count := strings.Count(definition.Expression, marker); count != tc.count {
				t.Fatalf("%s selected canonical operation %d times, want %d", tc.consumer, count, tc.count)
			}
		})
	}
}

func TestRuntimeWordCleanBreakReleasesReplacedInvocationBeforeSpawn(t *testing.T) {
	c := compileDefinitionSource(t, "Use Runtime from `runtime`.\n", "runtime-insertion-release.l")
	definition, err := c.ComposeDefinition("Runtime.App.Supervise")
	if err != nil {
		t.Fatal(err)
	}
	insert := strings.Index(definition.Expression, "Key: ~insert~")
	release := strings.Index(definition.Expression, "Key: ~std::mem::drop~")
	spawn := strings.Index(definition.Expression, "Key: ~tokio::spawn~")
	if insert < 0 || release <= insert || spawn <= release {
		t.Fatalf("replaced invocation must be released after insertion and before spawn: %s", definition.Expression)
	}
	for _, delayed := range []bool{false, true} {
		t.Run(map[bool]string{false: "selected", true: "delayed-release-control"}[delayed], func(t *testing.T) {
			files := dictionaryIndexFilesystem(t)
			process := files["runtime/eng/process.l"]
			process.Data = bytes.ReplaceAll(process.Data, []byte("Key: ~tokio::spawn~"), []byte("Key: ~crate::runtime::host::observe_spawn~"))
			if delayed {
				// A provider-only negative control changes source order, never Rust.
				app := files["runtime/eng/app.l"]
				const release = "Storage.Release Previous.\n"
				const tail = "Return a New Nothing.\n"
				if bytes.Count(app.Data, []byte(release)) != 1 || !bytes.HasSuffix(app.Data, []byte(tail)) {
					t.Fatal("release control no longer selects the final Supervise body")
				}
				app.Data = bytes.Replace(app.Data, []byte(release), nil, 1)
				app.Data = append(bytes.TrimSuffix(app.Data, []byte(tail)), []byte(release+tail)...)
			}
			files["runtime/rust/host.rs"].Data = append(files["runtime/rust/host.rs"].Data, []byte(replacedInvocationObserver)...)
			deriveTestRuntime(t, files)
			c := NewLygerService(Options{Resources: files}).NewCompilation([]byte("Create \"Checks\" as a Dictionary.\n"), "runtime-insertion-release.l")
			c.Make()
			if c.Error != nil {
				t.Fatal(c.Error)
			}
			wordLibraryRuntimeObserver(t, c, "Checks", strings.ReplaceAll(`pub(crate) fn exercise() {
 std::thread::spawn(|| tokio::runtime::Builder::new_current_thread().enable_all().build().unwrap().block_on(crate::runtime::host::observe_replacement(EXPECTED))).join().unwrap();
 println!("word-library error assertions reached");
}`, "EXPECTED", strconv.FormatBool(!delayed)))
		})
	}
}

const replacedInvocationObserver = `
static REPLACED: Mutex<Option<(std::sync::Weak<Invocation>, usize, bool)>> = Mutex::new(None);
pub(crate) fn observe_spawn<F>(future: F) -> tokio::task::JoinHandle<F::Output>
where F: Future + Send + 'static, F::Output: Send + 'static {
 if let Some((old, count, released)) = REPLACED.lock().unwrap().as_mut() {
  *count += 1; *released = old.upgrade().is_none();
 }
 tokio::spawn(future)
}
pub(crate) async fn observe_replacement(expected: bool) {
 use tokio::io::AsyncWriteExt;
 let owner = super::runtime_assemble(Arc::new(NativeDependencies), Box::pin(tokio::io::sink()), Box::pin(tokio::io::sink()));
 let allocate = || super::invocation_allocate(std::sync::Weak::new(), Arc::downgrade(&owner.inner), 41, Arc::new(NativeDependencies), tokio::time::Duration::from_secs(2), None);
 let (old, _) = allocate(); let old_weak = Arc::downgrade(&old);
 let (new, receiver) = allocate();
 let mut command = Command::new("/bin/sh"); command.arg("-c").arg("read value").stdin(std::process::Stdio::piped()).kill_on_drop(true);
 process::configure(&mut command).unwrap();
 let mut child = command.spawn().unwrap(); let pid = child.id().unwrap(); let mut input = child.stdin.take().unwrap();
 {
  let mut admission = lock(&owner.inner.state);
  assert!(admission.owned.insert(41, old).is_none());
  assert_eq!(old_weak.strong_count(), 1, "map must be the old invocation's sole owner");
  *REPLACED.lock().unwrap() = Some((old_weak.clone(), 0, false));
  super::app_supervise(&new, child, receiver, &mut admission, &owner);
  assert_eq!(admission.owned.len(), 1); assert!(Arc::ptr_eq(admission.owned.get(&41).unwrap(), &new));
 }
 let (_, count, released) = REPLACED.lock().unwrap().take().unwrap();
 input.write_all(b"done\n").await.unwrap(); drop(input);
 tokio::time::timeout(tokio::time::Duration::from_secs(5), super::invocation_completed(&new)).await.unwrap();
 assert!(lock(&owner.inner.state).owned.is_empty());
 assert_eq!(lock(&new.state).snapshot.as_ref().unwrap().status, Some(0));
 assert!(new.group.lock().await.cleaned);
 assert_eq!(nix::sys::signal::kill(nix::unistd::Pid::from_raw(pid as i32), None), Err(nix::errno::Errno::ESRCH));
 assert!(old_weak.upgrade().is_none()); assert_eq!(count, 1);
 assert_eq!(released, expected, "old invocation must be released before delegating Spawn");
}
`

func TestRuntimeWordCleanBreakErrorIdentityAndBytesNative(t *testing.T) {
	source := "Use Runtime from `runtime`.\nUse Storage from `storage`.\n\n" +
		"    Publicly create \"Exercise\" as a Generic `exercise` Function with Inputs of \"first\" as `Vec<u8>` and \"second\" as `Vec<u8>` and \"stderr\" as `&[u8]` and Outputs of `(Vec<u8>, Vec<u8>)`.\n" +
		"Create \"Initial\" as a New Runtime's Error's Former of First.\n" +
		"Create \"Repeated\" as a New Runtime's Error's Copier of a New Storage's Borrower of Initial.\n" +
		"Create \"Attached\" as a New Runtime's Error's Attacher of a New Runtime's Error's Copier of a New Storage's Borrower of Initial and Stderr.\n" +
		"Create \"Later\" as a New Runtime's Error's Former of Second.\n" +
		"Create \"Causes\" as a New Storage's Vector of Initial, Repeated, and Later.\n" +
		"Create \"Joined\" as a New Runtime's Error's Joiner of Causes.\n" +
		"Create \"Required\" as a New Storage's Expecter of Joined and `three joined causes`.\n" +
		"Create \"Rendered\" as a New Runtime's Error's Renderer of a New Storage's Borrower of Required.\n" +
		"Create \"Annotated\" as a New Runtime's Error's Renderer of a New Storage's Borrower of Attached.\n" +
		"Return a New Storage's Tuple of Rendered and Annotated.\n"
	c := compileDefinitionSource(t, source, "runtime-error-identity.l")
	testNativeDefinitionLibrary(t, c, "Exercise", "runtime_error", `#[path="src/runtime_error.rs"] mod generated;
fn main() {
 let (joined,attached)=generated::exercise(vec![255],vec![],&[254,0]);
 assert_eq!(joined,vec![255,10],"duplicate Arc cause must be omitted while an empty leaf contributes a line");
 assert_eq!(attached,vec![239,191,189,58,32,254,0],"diagnostic prefix is lossy UTF-8 while stderr stays raw");
 let (joined,attached)=generated::exercise(b"first".to_vec(),b"second".to_vec(),&[]);
 assert_eq!(joined,b"first\nsecond");
 assert_eq!(attached,b"first","empty stderr preserves the original error");
}`)
}

func TestRuntimeWordCleanBreakActionFamilies(t *testing.T) {
	c := compileDefinitionSource(t, "Use Runtime from `runtime`.\n", "runtime-word-families.l")
	unit := libraryFoundationUnit(t, c, "Runtime")
	for _, forms := range []string{
		"Advance Advances Advanced Advancing Advancer", "Decode Decodes Decoded Decoding Decoder",
		"Extend Extends Extended Extending Extender", "Canonicalize Canonicalizes Canonicalized Canonicalizing Canonicalizer",
		"Sleep Sleeps Slept Sleeping Sleeper", "Suspend Suspends Suspended Suspending Suspender",
		"Delay Delays Delayed Delaying Delayer", "Feed Feeds Fed Feeding Feeder",
		"Redirect Redirects Redirected Redirecting Redirecter", "Divert Diverts Diverted Diverting Diverter",
		"Validate Validates Validated Validating Validator", "Share Shares Shared Sharing Sharer",
	} {
		words := strings.Fields(forms)
		t.Run(words[0], func(t *testing.T) {
			rec := libraryFoundationBinding(t, c, unit, words[0])
			if rec.bindingWord().POS != "verb" {
				t.Fatalf("private %s is not an action family", words[0])
			}
			for _, form := range words {
				binding, err := resolveTokenBinding([]byte(form), map[string]*record{rec.ID: rec})
				if err != nil || binding == nil || binding.Record != rec {
					t.Errorf("private family lost %s: %v", form, err)
				}
				if exposed, err := c.builtinRecord("Runtime." + form); err == nil && exposed == rec {
					t.Errorf("private form %s became a Runtime dictionary member", form)
				}
			}
		})
	}
	for _, tc := range []struct{ name, forms string }{
		{"Runtime.Supervision.Wait", "Wait Waits Waited Waiting Waiter"},
		{"Runtime.App.Build", "Build Builds Built Building Builder"},
		{"Runtime.App.Run", "Run Runs Ran Running Runner"},
		{"Runtime.App.Set", "Set Sets Set Setting Setter"},
		{"Runtime.Invocation.Read", "Read Reads Read Reading Reader"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec, err := c.builtinRecord(tc.name)
			if err != nil || rec == nil || rec.bindingWord().POS != "verb" {
				t.Fatalf("action family %s: %v", tc.name, err)
			}
			prefix := tc.name[:strings.LastIndex(tc.name, ".")+1]
			for _, form := range strings.Fields(tc.forms) {
				binding, err := resolveTokenBinding([]byte(prefix+form), map[string]*record{rec.ID: rec})
				if err != nil || binding == nil || binding.Record != rec {
					t.Errorf("%s lost form %s: %v", tc.name, form, err)
				}
			}
		})
	}
}

func TestRuntimeWordCleanBreakSharedOutputPin(t *testing.T) {
	c := compileDefinitionSource(t, "Use Runtime from `runtime`.\n", "runtime-output-pin.l")
	pin, err := c.builtinRecord("Storage.Pin")
	if err != nil || pin == nil {
		t.Fatalf("shared Pin unavailable: %v", err)
	}
	output, err := c.builtinRecord("Runtime.Output")
	if err != nil {
		t.Fatal(err)
	}
	descriptor := c.Semantics.Descriptors[output]
	if descriptor == nil || !descriptor.Targets[nominalIdentity{definition: pin}] {
		t.Fatal("Runtime.Output must target the canonical shared Pin family")
	}
	open, err := c.ComposeDefinition("Runtime.Open")
	if err != nil {
		t.Fatal(err)
	}
	if count := strings.Count(open.Expression, "Key: ~Box::pin~"); count != 2 {
		t.Fatalf("stdout and stderr require exactly one pin each, got %d: %s", count, open.Expression)
	}
	if count := strings.Count(open.Expression, "crate::runtime::host::Output"); count != 2 {
		t.Fatalf("both stream constructors need their host trait coercion, got %d: %s", count, open.Expression)
	}
}

func TestRuntimeWordCleanBreakOutputNativeCoercion(t *testing.T) {
	source := "Use Runtime from `runtime`.\nUse Storage from `storage`.\nCreate \"Checks\" as a Dictionary.\n\n" +
		"    Publicly create Checks's \"Outputs\" as a Generic `outputs` Function with No Inputs and Outputs of `(crate::runtime::host::Output, crate::runtime::host::Output)`.\n" +
		"Create \"Output\" as a New Runtime.Output Storage's Pinner of a New Runtime's Stdout.\n" +
		"Create \"Error\" as a New Runtime.Output Storage's Pinner of a New Runtime's Stderr.\n" +
		"Return a New Storage's Tuple of Output and Error.\n"
	c := compileDefinitionSource(t, source, "runtime-output-coercion.l")
	wordLibraryRuntimeObserver(t, c, "Checks", `pub(crate) fn exercise() {
 let (output,error)=super::outputs();
 let _: crate::runtime::host::Output=output;
 let _: crate::runtime::host::Output=error;
 println!("word-library error assertions reached");
}`)
}

func TestRuntimeWordCleanBreakJoinSourceAndSerialOrder(t *testing.T) {
	source := "Use Runtime from `runtime`.\nUse Storage from `storage`.\n\n" +
		"    Publicly create \"Exercise\" as a Generic `exercise` Function with Inputs of \"values\" as `&[crate::Item]` and \"settle\" as `for<'a> fn(&'a crate::Item) -> crate::Step<'a>` and \"complete\" as `for<'a> fn(&'a crate::Item) -> crate::Step<'a>` and Outputs of `impl std::future::Future<Output = ()> + '_`.\n" +
		"Return a New Storage's Future of a New Process for Runtime.Joining Values and Settle and then Runtime.Joining Values and Complete.\n"
	c := compileDefinitionSource(t, "Use Runtime from `runtime`.\n", "runtime-join-source.l")
	join, err := c.builtinRecord("Runtime.Join")
	if err != nil {
		t.Fatal(err)
	}
	body, err := c.Semantics.definitionActions(join)
	if err != nil || len(body) == 0 {
		t.Fatalf("Join must expose its iteration and await as L source actions: %v", err)
	}
	c = compileDefinitionSource(t, source, "runtime-join-order.l")
	testNativeDefinitionLibrary(t, c, "Exercise", "runtime_join", `use std::future::Future;
use std::pin::Pin;
use std::sync::{Arc, Mutex};
use std::task::{Context, Poll, Wake, Waker};
static EVENTS: Mutex<Vec<(u8,u8,u8)>> = Mutex::new(Vec::new());
pub struct Item(u8);
pub struct Step<'a> { item: &'a Item, phase: u8, polled: bool }
impl Future for Step<'_> {
 type Output=();
 fn poll(mut self: Pin<&mut Self>, _: &mut Context<'_>) -> Poll<()> {
  EVENTS.lock().unwrap().push((self.phase,self.item.0,if self.polled {2} else {1}));
  if self.polled { Poll::Ready(()) } else { self.polled=true; Poll::Pending }
 }
}
fn step(item: &Item, phase: u8) -> Step<'_> { EVENTS.lock().unwrap().push((phase,item.0,0)); Step{item,phase,polled:false} }
fn settle(item: &Item) -> Step<'_> { step(item,0) }
fn complete(item: &Item) -> Step<'_> { step(item,1) }
struct Noop;
impl Wake for Noop { fn wake(self: Arc<Self>) {} }
#[path="src/runtime_join.rs"] mod generated;
fn main() {
 let items=[Item(3),Item(7)];
 let waker=Waker::from(Arc::new(Noop)); let mut context=Context::from_waker(&waker);
 let mut future=Box::pin(generated::exercise(&items,settle,complete));
 assert!(EVENTS.lock().unwrap().is_empty());
 let mut finished=false;
 for _ in 0..5 { if future.as_mut().poll(&mut context).is_ready() { finished=true; break; } }
 assert!(finished,"serial Join did not finish within its exact poll allowance");
 assert_eq!(*EVENTS.lock().unwrap(),vec![(0,3,0),(0,3,1),(0,3,2),(0,7,0),(0,7,1),(0,7,2),(1,3,0),(1,3,1),(1,3,2),(1,7,0),(1,7,1),(1,7,2)]);
}`)
}
