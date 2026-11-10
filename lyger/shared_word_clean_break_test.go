package Lyger

import (
	"strings"
	"testing"
)

func TestSharedWordFamilies(t *testing.T) {
	c := compileDefinitionSource(t, "Use Storage from `storage`.\n", "shared-word-families.l")
	for _, test := range []struct{ name, pos, forms string }{
		{"Storage.Cast", "verb", "Cast Casts Casting Caster"},
		{"Storage.Dereference", "verb", "Dereference Dereferences Dereferenced Dereferencing Dereferencer"},
		{"Storage.Extract", "verb", "Extract Extracts Extracted Extracting Extractor"},
		{"Storage.Share", "verb", "Share Shares Shared Sharing Sharer"},
		{"Storage.Expect", "verb", "Expect Expects Expected Expecting Expecter"},
		{"Storage.Unwrap", "verb", "Unwrap Unwraps Unwrapped Unwrapping Unwrapper"},
		{"Storage.Number.Format", "verb", "Format Formats Formatted Formatting Formatter"},
		{"Storage.Copyable", "adjective", "Copyable"},
	} {
		t.Run(test.name, func(t *testing.T) {
			rec, err := wordLibraryRecord(c, test.name)
			if err != nil || rec == nil {
				t.Fatalf("missing selected family %s: %v", test.name, err)
			}
			if word := rec.bindingWord(); word == nil || word.POS != test.pos {
				t.Fatalf("%s has wrong grammatical role: %+v", test.name, word)
			}
			prefix := test.name[:strings.LastIndex(test.name, ".")+1]
			for _, form := range strings.Fields(test.forms) {
				got, err := wordLibraryRecord(c, prefix+form)
				if err != nil || got != rec {
					t.Errorf("%s lost canonical identity: %v", prefix+form, err)
				}
			}
		})
	}
	for _, name := range []string{"Storage.Integral", "Storage.Optional", "Storage.Result.Expect", "Storage.Result.Alternative", "Storage.Function.Copyable", "Storage.Number.Finite"} {
		t.Run("retired-"+name, func(t *testing.T) {
			if rec, err := c.builtinRecord(name); err == nil && rec != nil {
				t.Fatalf("retired canonical meaning %s remains at %s", name, rec.ID)
			}
		})
	}
}

func TestSharedWordOwnership(t *testing.T) {
	source := "Use Storage from `storage`.\nCreate \"Trial\" as a Dictionary.\n\n" +
		"    Publicly create Trial's \"Extract\" as a Generic `trial_extract` Storage.Routine of `T, E` with Inputs of \"value\" as `Result<T, E>` and Outputs of `Option<T>`.\nReturn a New Storage's Extractor of Value.\n\n" +
		"    Publicly create Trial's \"Share\" as a Generic `trial_share` Storage.Routine of `T` with Inputs of \"value\" as `T` and Outputs of `std::rc::Rc<T>`.\nReturn a New Storage's Sharer of Value.\n\n" +
		"    Publicly create Trial's \"Fallback\" as a Generic `trial_fallback` Storage.Routine of `T, E` with Inputs of \"value\" as `Result<T, E>` and \"fallback\" as `T` and Outputs of `T`.\nReturn a New Storage's Unwrapper of Value and Fallback.\n"
	c := compileDefinitionSource(t, source, "shared-word-ownership.l")
	testNativeDefinitionLibrary(t, c, "Trial", "shared_words", `#[path="src/shared_words.rs"] mod generated;
use std::cell::RefCell;
use std::rc::Rc;
struct Item(u8, Rc<RefCell<Vec<u8>>>);
impl Drop for Item { fn drop(&mut self) { self.1.borrow_mut().push(self.0); } }
fn main() {
 let events = Rc::new(RefCell::new(Vec::new()));
 let item = |id| Item(id, Rc::clone(&events));
 let extracted = generated::trial_extract::<_, Item>(Ok(item(1))).unwrap();
 assert!(events.borrow().is_empty());
 drop(extracted); assert_eq!(*events.borrow(), [1]);
 assert!(generated::trial_extract::<Item, _>(Err(item(2))).is_none());
 assert_eq!(*events.borrow(), [1,2]);
 let shared = generated::trial_share(item(3));
 let same = Rc::clone(&shared); assert!(Rc::ptr_eq(&shared, &same));
 drop(shared); assert_eq!(*events.borrow(), [1,2]);
 drop(same); assert_eq!(*events.borrow(), [1,2,3]);
 let successful = generated::trial_fallback::<_, Item>(Ok(item(4)), item(5));
 assert_eq!(*events.borrow(), [1,2,3,5]); drop(successful);
 let fallback = generated::trial_fallback(Err(item(6)), item(7));
 assert_eq!(*events.borrow(), [1,2,3,5,4,6]); drop(fallback);
 assert_eq!(*events.borrow(), [1,2,3,5,4,6,7]);
}`)
}

func TestSelectorPatternsArePrivate(t *testing.T) {
	c := compileDefinitionSource(t, "Use Storage from `storage`.\n", "selector-ownership.l")
	for _, name := range []string{"Storage.Option.Present", "Storage.Option.Absent", "Storage.Option.Choice", "Storage.Result.Accepted", "Storage.Result.Rejected", "Storage.Result.Choice"} {
		t.Run(name, func(t *testing.T) {
			if rec, err := c.builtinRecord(name); err == nil && rec != nil {
				t.Fatalf("public selector implementation %s remains at %s", name, rec.ID)
			}
		})
	}
}
