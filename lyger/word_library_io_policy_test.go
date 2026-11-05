package Lyger

import (
	"bytes"
	"testing"
	"testing/fstest"
)

func TestWordLibraryIOResourceHandles(t *testing.T) {
	source := "Use IO from `io`.\nCreate \"Checks\" as a Dictionary.\n\n" +
		"    Publicly create Checks's \"Output\" as a Generic `output` Function with No Inputs and Outputs of `std::io::Stdout`.\nReturn a New IO's Output.\n\n" +
		"    Publicly create Checks's \"Diagnostic\" as a Generic `diagnostic` Function with No Inputs and Outputs of `std::io::Stderr`.\nReturn a New IO's Diagnostic.\n"
	c := compileDefinitionSource(t, source, "word-io-handles.l")
	testNativeDefinitionLibrary(t, c, "Checks", "word_io_handles", `#[path="src/word_io_handles.rs"] mod generated;
fn main() { let _: std::io::Stdout = generated::output(); let _: std::io::Stderr = generated::diagnostic(); }`)
}

func TestWordLibraryWriteUsesOneGuard(t *testing.T) {
	files := dictionaryIndexFilesystem(t)
	const path = "stdlib/eng/io/dictionary.l"
	original := files[path].Data
	selected := bytes.Replace(original, []byte("Key: ~std::io::stdout~"), []byte("Key: ~crate::output~"), 1)
	if bytes.Equal(original, selected) {
		t.Fatal("Output resource leaf was not selected")
	}
	// Expose a consumer of the private Write in this selected dictionary only.
	// Its complete implementation and dependency graph remain production source.
	selected = append(selected, []byte("\n    Publicly create IO's \"Observe\" as a Generic `observe_write` Function with Inputs of \"value\" as `Vec<u8>` and Outputs of `()`.\nReturn a New Writer of Value.\n")...)
	files[path] = &fstest.MapFile{Data: selected}
	source := "Use IO from `io`.\nCreate \"Checks\" as a Dictionary.\nPublicly use IO.Observe as Checks's \"Observe\".\n"
	c := NewLygerService(Options{Language: "eng", Resources: files}).NewCompilation([]byte(source), "word-write-policy.l")
	c.Make()
	if c.Error != nil {
		t.Fatal(c.Error)
	}
	testNativeDefinitionLibrary(t, c, "Checks", "word_write_policy", `use std::io::{self, Write};
use std::sync::Mutex;
#[derive(Default)] struct State { bytes: Vec<u8>, outputs: usize, locks: usize, drops: usize, writes: usize, flushes: usize, fail: usize }
static STATE: Mutex<State> = Mutex::new(State { bytes: Vec::new(), outputs: 0, locks: 0, drops: 0, writes: 0, flushes: 0, fail: 0 });
pub struct Output;
pub struct Guard;
pub fn output() -> Output { STATE.lock().unwrap().outputs += 1; Output }
impl Output { pub fn lock(&self) -> Guard { STATE.lock().unwrap().locks += 1; Guard } }
impl Drop for Guard { fn drop(&mut self) { STATE.lock().unwrap().drops += 1; } }
impl Write for Guard {
 fn write(&mut self, bytes: &[u8]) -> io::Result<usize> {
  let mut state = STATE.lock().unwrap(); state.writes += 1;
  if state.fail == state.writes { if bytes.len() > 1 { state.bytes.push(bytes[0]); } return Err(io::Error::other("delivery")); }
  state.bytes.extend_from_slice(bytes); Ok(bytes.len())
 }
 fn flush(&mut self) -> io::Result<()> { STATE.lock().unwrap().flushes += 1; Ok(()) }
}
#[path="src/word_write_policy.rs"] mod generated;
fn reset(fail: usize) { *STATE.lock().unwrap() = State { fail, ..State::default() }; }
fn check(expected: &[u8], writes: usize) {
 let state = STATE.lock().unwrap(); assert_eq!(state.bytes, expected); assert_eq!(state.writes, writes);
 assert_eq!(state.outputs, 1); assert_eq!(state.locks, 1); assert_eq!(state.drops, 1); assert_eq!(state.flushes, 0);
}
fn failed(fail: usize, expected: &[u8], writes: usize) {
 reset(fail); let panic = std::panic::catch_unwind(|| generated::observe_write(vec![0,255,b'\n'])).unwrap_err();
 let message = panic.downcast_ref::<String>().expect("expect panic payload");
 assert!(message.starts_with("failed printing to stdout")); assert!(message.contains("delivery")); check(expected, writes);
}
fn main() {
 reset(0); generated::observe_write(vec![0,255,b'\n']); check(&[0,255,b'\n',b'\n'], 2);
 reset(0); generated::observe_write(Vec::new()); check(b"\n", 1);
 failed(1, &[0], 1); failed(2, &[0,255,b'\n'], 2);
}`)
}
