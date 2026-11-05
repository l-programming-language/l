package Lyger

import "testing"

func TestWordLibraryTraversalBehavior(t *testing.T) {
	const imports = "Use Storage from `storage`.\nCreate \"Checks\" as a Dictionary.\n"
	const observers = `use std::sync::atomic::{AtomicUsize, Ordering};
static READS: AtomicUsize = AtomicUsize::new(0);
pub fn gather() -> &'static [u64] { READS.fetch_add(1, Ordering::SeqCst); &[1,2] }
fn visit(value: &u64, seen: &mut Vec<u64>) { seen.push(*value); }
fn ignored(value: &u64, seen: &mut Vec<u64>) -> Result<(), u8> { visit(value, seen); Err(9) }
fn panics(value: &u64, seen: &mut Vec<u64>) -> Result<(), u8> { visit(value, seen); if *value == 2 { panic!("callback panic"); } Err(9) }
fn pair(first: &u64, second: &u64, seen: &mut Vec<u64>) { seen.extend([*first,*second]); }
fn ignored_pair(first: &u64, second: &u64, seen: &mut Vec<u64>) -> Result<(), u8> { pair(first, second, seen); Err(9) }
fn find(value: &u64, seen: &mut Vec<u64>) -> Result<bool, u8> { visit(value, seen); if *value == 3 { Err(7) } else { Ok(value % 2 == 0) } }
pub fn found(seen: &mut Vec<u64>) { seen.push(99); }
static DROP_EVENTS: std::sync::Mutex<Vec<u64>> = std::sync::Mutex::new(Vec::new());
pub struct Dropped(u64);
impl Drop for Dropped { fn drop(&mut self) { DROP_EVENTS.lock().unwrap().push(100+self.0); } }
fn checked_drop(value: &u64, _: &mut Vec<u64>) -> Result<Dropped,u8> { DROP_EVENTS.lock().unwrap().push(*value); if *value == 2 { Err(9) } else { Ok(Dropped(*value)) } }
fn paired_drop(first: &u64, second: &u64, _: &mut Vec<u64>) -> Result<Dropped,u8> { DROP_EVENTS.lock().unwrap().push(*first); if *first == 2 { Err(9) } else { Ok(Dropped(*second)) } }
#[path="src/traversal_extra.rs"] mod generated;
`
	for _, test := range []struct{ name, source, observe string }{
		{"ignored-errors", "\n    Publicly create Checks's \"Visit\" as a Generic `visit` Function with Inputs of \"values\" as `&[u64]` and \"action\" as `fn(&u64, &mut Vec<u64>) -> Result<(), u8>` and \"context\" as `&mut Vec<u64>` and Outputs of `()`.\nStorage.Traverse Values, Action, and Context.\nReturn a New Storage's Nothing.\n\n" +
			"    Publicly create Checks's \"Pairs\" as a Generic `pairs` Function with Inputs of \"values\" as `&[(u64,u64)]` and \"action\" as `fn(&u64, &u64, &mut Vec<u64>) -> Result<(), u8>` and \"context\" as `&mut Vec<u64>` and Outputs of `()`.\nStorage.Enumerate Values, Action, and Context.\nReturn a New Storage's Nothing.\n",
			`let mut seen=vec![]; generated::visit(&[1,2,3], ignored, &mut seen); assert_eq!(seen,[1,2,3]); seen.clear(); generated::pairs(&[(1,10),(2,20)],ignored_pair,&mut seen); assert_eq!(seen,[1,10,2,20]); seen.clear(); std::panic::set_hook(Box::new(|_|{})); assert!(std::panic::catch_unwind(std::panic::AssertUnwindSafe(||generated::visit(&[1,2,3],panics,&mut seen))).is_err()); assert_eq!(seen,[1,2]);`},
		{"iterable-once", "Define \"Gathered\" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~crate::gather~|}<DEF>`.\n\n" +
			"    Publicly create Checks's \"Exercise\" as a Generic `exercise` Function with Inputs of \"action\" as `fn(&u64, &mut Vec<u64>)` and \"context\" as `&mut Vec<u64>` and Outputs of `()`.\nStorage.Traverse a New Gathered, Action, and Context.\nReturn a New Storage's Nothing.\n",
			`let mut seen=vec![]; generated::exercise(visit,&mut seen); assert_eq!(seen,[1,2]); assert_eq!(READS.load(Ordering::SeqCst),1);`},
		{"discarded-success", "\n    Publicly create Checks's \"Check\" as a Generic `check` Function with Inputs of \"values\" as `&[u64]` and \"action\" as `fn(&u64, &mut Vec<u64>) -> Result<crate::Dropped,u8>` and \"context\" as `&mut Vec<u64>` and Outputs of `Result<(),u8>`.\nStorage.Inspect Values, Action, and Context.\nReturn a New Storage's Success of a New Storage's Nothing.\n\n" +
			"    Publicly create Checks's \"Audit\" as a Generic `audit` Function with Inputs of \"values\" as `&[(u64,u64)]` and \"action\" as `fn(&u64, &u64, &mut Vec<u64>) -> Result<crate::Dropped,u8>` and \"context\" as `&mut Vec<u64>` and Outputs of `Result<(),u8>`.\nStorage.Audit Values, Action, and Context.\nReturn a New Storage's Success of a New Storage's Nothing.\n",
			`let mut seen=vec![]; assert_eq!(generated::check(&[1,2,3],checked_drop,&mut seen),Err(9)); assert_eq!(&*DROP_EVENTS.lock().unwrap(),&[1,101,2]); DROP_EVENTS.lock().unwrap().clear(); assert_eq!(generated::audit(&[(1,10),(2,20),(3,30)],paired_drop,&mut seen),Err(9)); assert_eq!(&*DROP_EVENTS.lock().unwrap(),&[1,110,2]);`},
		{"repeated-labels", "\n    Publicly create Checks's \"Exercise\" as a Generic `exercise` Function with Inputs of \"item\" as `&[u64]` and \"action\" as `fn(&u64, &mut Vec<u64>)` and \"outcome\" as `&mut Vec<u64>` and Outputs of `()`.\nStorage.Traverse Item, Action, and Outcome.\nStorage.Traverse Item, Action, and Outcome.\nReturn a New Storage's Nothing.\n",
			`let mut seen=vec![]; generated::exercise(&[1,2],visit,&mut seen); assert_eq!(seen,[1,2,1,2]);`},
		{"nested-labels", "Define \"Perform|-s|-ed|-ing|-er\" to `<DEF><@Body@><DEF>`.\n\n" +
			"    Publicly create Checks's \"Exercise\" as a Generic `exercise` Function with Inputs of \"values\" as `&[u64]` and \"action\" as `fn(&u64, &u64, &mut Vec<u64>)` and \"context\" as `&mut Vec<u64>` and Outputs of `()`.\nPerform a New Storage's Iteration of Values and a New Process with Inputs of \"Item\" as `_` for Creating \"Saved\" as a New Storage's Binding of Item and then Performing a New Storage's Iteration of Values and a New Process with Inputs of \"Item\" as `_` for Creating \"Outcome\" as a New Storage.Paired Storage's Applier of Action, Saved, Item, and Context.\nReturn a New Storage's Nothing.\n",
			`let mut seen=vec![]; generated::exercise(&[1,2],pair,&mut seen); assert_eq!(seen,[1,1,1,2,2,1,2,2]);`},
		{"continuation-process", "Define \"Note|-s|-d|--ing|-r\" to `<DEF>MOB{|Key: ~_~|}^KY{|Key: ~crate::found~, Inputs: ~&[<@Context@>]&~|}<DEF>`.\n\n" +
			"    Publicly create Checks's \"Exercise\" as a Generic `exercise` Function with Inputs of \"values\" as `&[u64]` and \"action\" as `fn(&u64, &mut Vec<u64>) -> Result<bool,u8>` and \"context\" as `&mut Vec<u64>` and Outputs of `Result<(),u8>`.\nStorage.Search Values, Action, Context, and a New Process for Noting Context.\nReturn a New Storage's Success of a New Storage's Nothing.\n",
			`let mut seen=vec![]; assert_eq!(generated::exercise(&[1,2,4],find,&mut seen),Ok(())); assert_eq!(seen,[1,2,99,4,99]); seen.clear(); assert_eq!(generated::exercise(&[2,3,4],find,&mut seen),Err(7)); assert_eq!(seen,[2,99,3]);`},
	} {
		t.Run(test.name, func(t *testing.T) {
			c := compileDefinitionSource(t, imports+test.source, "word-traversal.l")
			testNativeDefinitionLibrary(t, c, "Checks", "traversal_extra", observers+"fn main() {"+test.observe+"}")
		})
	}
}
