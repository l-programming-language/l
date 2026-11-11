package Lyger

import (
	"bytes"
	"fmt"
	"testing"
)

// Observe the selected Copy/read/update definitions. The failure provider adds
// an L rejection after the real update, so publishing its mutated snapshot would
// be visible. Native observers only record effects and inspect dropped values.
func TestPassivePutOperandEffectsNative(t *testing.T) {
	for _, kind := range []string{"List", "Map"} {
		for _, fail := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/failure=%t", kind, fail), func(t *testing.T) {
				files := dictionaryIndexFilesystem(t)
				values := files["builtins/eng/storage/values.l"]
				values.Data = bytes.ReplaceAll(values.Data, []byte("Key: ~std::clone::Clone::clone~"), []byte("Key: ~crate::runtime::host::observe_copy~"))
				values.Data = bytes.Replace(values.Data, []byte("KY{|Key: ~as_bytes~, Receiver: ~<@Value@>~|}"), []byte("KY{|Key: ~crate::runtime::host::observe_text~, Inputs: ~&[&<@Value@>]&~|}"), 1)
				const leaf = "\nDefine \"Observe\" to `<DEF>MOB{|Key: ~_~|}^KY{|Key: ~crate::runtime::host::observe_put~, Inputs: ~&[<@Event@>]&~|}<DEF>`.\n" +
					"Define \"ObserveAccess\" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~crate::runtime::host::observe_access~, Inputs: ~&[<@Message@>]&~|}<DEF>`.\n" +
					"Define \"ObserveDestination\" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~crate::runtime::host::observe_destination~, Inputs: ~&[<@Value@>]&~|}<DEF>`.\n" +
					"Define \"ObserveUpdate\" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~crate::runtime::host::observe_update~, Inputs: ~&[<@Value@>]&~|}<DEF>`.\n"
				values.Data = append([]byte(leaf), values.Data...)
				mark := func(path, declaration, before, observation string) {
					t.Helper()
					file := files[path]
					start := bytes.Index(file.Data, []byte(declaration))
					if start < 0 {
						t.Fatalf("missing selected declaration %s", declaration)
					}
					body := file.Data[start:]
					end := bytes.IndexByte(body, '\n') + 1
					if before != "" {
						end = bytes.Index(body, []byte(before))
						if end < 0 {
							t.Fatalf("missing selected statement %s", before)
						}
					}
					file.Data = append(append(bytes.Clone(file.Data[:start+end]), []byte(observation)...), file.Data[start+end:]...)
				}
				mark("builtins/eng/storage/execution.l", `Publicly create Storage's "Access|`, "", "Create \"Observed\" as a New ObserveAccess of Message.\n")
				mark("builtins/eng/storage/values.l", `Publicly create Storage's `+kind+`'s "Copy|`, "", "Create \"Observed\" as a New ObserveDestination of Value.\n")
				helper, after := "Append", "Return a New Storage's Success of Added."
				initial, key, wanted := "a New List of `before`", "", "[before, payload]"
				expected := `"destination-read", "copy", "destination", "update", "changed"`
				drop := `assert_eq!(self.values.iter().map(|v| v.data.downcast_ref::<Vec<u8>>().unwrap().as_slice()).collect::<Vec<_>>(), if changed { vec![b"before".as_slice(), b"payload".as_slice()] } else { vec![b"before".as_slice()] });`
				if kind == "Map" {
					helper, after = "Update", "Create \"Unit\" as a New Storage's Nothing."
					initial, key, wanted = "a New Map of Text", " as `key`", "{key: payload}"
					expected = `"destination-read", "copy", "destination", "read", "copy", "update", "changed"`
					drop = `if changed { assert_eq!(self.values.len(),1); assert_eq!(self.values.get(b"key".as_slice()).unwrap().data.downcast_ref::<Vec<u8>>().unwrap(), b"payload"); } else { assert!(self.values.is_empty()); }`
				}
				declaration := `Publicly create Storage's ` + kind + `'s "` + helper + `|`
				mark("builtins/eng/storage/collections.l", declaration, "", "Create \"Observed\" as a New ObserveUpdate of "+kind+".\n")
				finish := "Observe `changed`.\n"
				if fail {
					finish += "Storage.Reject `selected Put failure`.\n"
				}
				mark("builtins/eng/storage/collections.l", declaration, after, finish)
				originalSize := 1
				if kind == "Map" {
					originalSize = 0
				}
				observer := fmt.Sprintf(passivePutObserver, fail, expected, kind, kind, kind, originalSize, drop)
				files["stdlib/runtime/rust/host.rs"].Data = append(files["stdlib/runtime/rust/host.rs"].Data, []byte(observer)...)
				deriveTestRuntime(t, files)
				source := "Define \"Payload\" as `payload`.\nDefine \"Items\" as " + initial + ".\nPayload is Put in Items" + key + " and then Show `continued`.\nShow Items.\nCreate \"Witness\" as a New Basic App.\n"
				result, err := NewLygerService(Options{Resources: files}).Compile([]byte(source), "passive-effects.l")
				if err != nil {
					t.Fatal(err)
				}
				_, bundle, err := orchestratorTestBundle(files, result)
				if err != nil {
					t.Fatal(err)
				}
				diagnostic := ""
				if fail {
					wanted, diagnostic = "preserved\n", "selected Put failure"
				} else {
					wanted = "continued" + wanted
				}
				runtimeBundleResult(t, bundle, wanted, diagnostic)
			})
		}
	}
}

const passivePutObserver = `
struct PutObservation { active: bool, finished: bool, destination: usize, snapshot: usize, events: Vec<&'static str> }
static PUT: Mutex<PutObservation> = Mutex::new(PutObservation { active:false, finished:false, destination:0, snapshot:0, events:Vec::new() });
const PUT_FAILURE: bool = %t;
const PUT_EXPECTED: &[&str] = &[%s];
pub(crate) fn observe_put(event: &'static str) { let mut state=PUT.lock().unwrap(); if state.active { state.events.push(event); } }
pub(crate) fn observe_access(message: &str) {
 let mut state=PUT.lock().unwrap();
 if message == "symbol Items is not initialized" && !state.finished { state.active=true; state.events.push("destination-read"); }
}
pub(crate) fn observe_text(value: &str) -> &[u8] {
 if value == "continued" { finish_put(); }
 else if value == "key" { observe_put("read"); }
 value.as_bytes()
}
pub(crate) fn observe_destination(value: &crate::L%s) {
 let mut state=PUT.lock().unwrap();
 if state.active { state.destination=value as *const _ as usize; state.events.push("destination"); }
}
pub(crate) fn observe_update(value: &crate::L%s) {
 let mut state=PUT.lock().unwrap();
 if state.active { state.snapshot=value as *const _ as usize; state.events.push("update"); }
}
pub(crate) fn observe_copy<T: Clone>(value: &T) -> T {
 if std::any::type_name::<T>() == "alloc::vec::Vec<u8>" { observe_put("copy"); }
 value.clone()
}
fn finish_put() {
 let mut state=PUT.lock().unwrap(); state.events.push("continuation"); state.active=false; state.finished=true;
 assert_eq!(state.events, [PUT_EXPECTED, &["drop-original","continuation"]].concat()); assert!(!PUT_FAILURE);
}
impl Drop for crate::L%s {
 fn drop(&mut self) {
  let mut state=PUT.lock().unwrap(); if !state.active { return; }
  let address=self as *const _ as usize;
  let original=address == state.destination; let changed=address == state.snapshot;
  if !original && !changed { return; }
  assert_ne!(original,changed); assert_eq!(self.values.len() != %d, changed);
  %s
  state.events.push(if changed { "drop-updated" } else { "drop-original" });
  if PUT_FAILURE && original {
   state.active=false; state.finished=true;
   assert_eq!(state.events, [PUT_EXPECTED, &["drop-updated","drop-original"]].concat()); println!("preserved");
  }
 }
}
`
