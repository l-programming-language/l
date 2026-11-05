package Lyger

import (
	"bytes"
	"strings"
	"testing"
	"testing/fstest"
)

// Selecting the one Core leaf must affect App retention as well. The observer
// counts native calls and ownership; it does not implement the L composition.
func TestWordLibraryAppRetentionUsesCore(t *testing.T) {
	files := dictionaryIndexFilesystem(t)
	const core = "builtins/eng/core.l"
	original := files[core].Data
	selected := bytes.Replace(original, []byte("Key: ~std::sync::Arc::clone~"), []byte("Key: ~crate::retain_core~"), 1)
	if bytes.Equal(original, selected) {
		t.Fatal("Core retention leaf was not selected")
	}
	files[core] = &fstest.MapFile{Data: selected}
	source := "Use Storage from `storage`.\nCreate \"Checks\" as a Dictionary.\n\n" +
		"    Publicly create Checks's \"App\" as a Generic `retain_app` Function with Inputs of \"value\" as `&std::sync::Arc<runtime::App>` and Outputs of `std::sync::Arc<runtime::App>`.\nReturn a New Storage's App's Retainer of Value.\n"
	c := NewLygerService(Options{Language: "eng", Resources: files}).NewCompilation([]byte(source), "word-app-retention.l")
	c.Make()
	if c.Error != nil {
		t.Fatal(c.Error)
	}
	testNativeDefinitionLibrary(t, c, "Checks", "word_app_retention", `extern crate self as runtime;
use std::sync::{Arc, atomic::{AtomicUsize, Ordering}};
static CALLS: AtomicUsize = AtomicUsize::new(0);
static DROPS: AtomicUsize = AtomicUsize::new(0);
pub struct App { payload: Box<[u8]> }
impl Drop for App { fn drop(&mut self) { DROPS.fetch_add(1, Ordering::SeqCst); } }
pub fn retain_core<T: ?Sized>(value: &Arc<T>) -> Arc<T> { CALLS.fetch_add(1, Ordering::SeqCst); Arc::clone(value) }
#[path="src/word_app_retention.rs"] mod generated;
fn main() {
 let original = Arc::new(App { payload: Box::new([0,255,10]) });
 let address = original.payload.as_ptr();
 let retained = generated::retain_app(&original);
 assert_eq!(CALLS.load(Ordering::SeqCst), 1, "App retention bypassed the selected Core leaf");
 assert!(Arc::ptr_eq(&original, &retained)); assert_eq!(Arc::strong_count(&original), 2);
 assert_eq!(retained.payload.as_ptr(), address); assert_eq!(&*retained.payload, &[0,255,10]);
 drop(original); assert_eq!(Arc::strong_count(&retained), 1); assert_eq!(DROPS.load(Ordering::SeqCst), 0);
 drop(retained); assert_eq!(DROPS.load(Ordering::SeqCst), 1);
}`)
}

func TestWordLibraryContinueUsesExplicitLabel(t *testing.T) {
	c := compileDefinitionSource(t, "Use Storage from `storage`.\n", "word-continue.l")
	rec, err := c.builtinRecord("Storage.Function.Continue")
	if err != nil {
		t.Fatal(err)
	}
	for _, label := range []string{"first_result", "second_result"} {
		instance, err := c.instantiateDefinition(rec, map[string]DefinitionArgument{
			"Name": definitionReferenceArgument(label), "Values": definitionReferenceArgument("owned_payload"),
		})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(instance.Expression, "Key: ~"+label+"~") || strings.Contains(instance.Expression, "lReturned") {
			t.Errorf("Continue did not bind its explicit result %s: %s", label, instance.Expression)
		}
	}
}
