package Lyger

import (
	"errors"
	"strings"
	"testing"

	"github.com/l-programming-language/l/lerrors"
)

func TestWordLibraryProcessReservations(t *testing.T) {
	const prefix = "Use Storage from `storage`.\nDefine \"Saved\" as 8.\nDefine \"Perform|-s|-ed|-ing|-er\" to `<DEF><@Body@><DEF>`.\n\n"
	for _, test := range []struct {
		name, inputs, body string
		kind               error
	}{
		{"duplicate-input", "\"values\" as `&[(u64,u64)]`", "Perform a New Storage.Paired Storage's Iteration of Values and a New Process with Inputs of \"Item\" as `&u64` and \"Item\" as `&u64` for Returning 0.", lerrors.DuplicateDeclaration},
		{"duplicate-local", "\"values\" as `&[u64]`", "Perform a New Storage's Iteration of Values and a New Process with Inputs of \"Item\" as `&u64` for Creating \"Local\" as a New Storage's Binding of Item and then Creating \"Local\" as a New Storage's Binding of Item.", lerrors.DuplicateDeclaration},
		{"sibling-local", "\"value\" as `Option<u64>`", "Storage.Option.Select Value, a New Process with Inputs of \"Item\" as `u64` for Creating \"Private\" as a New Storage's Binding of Item, and a New Process for Returning Private.", lerrors.UnresolvedReference},
		{"pending-local", "\"value\" as `u64`", "Create \"Before\" as a New Storage's Binding of Saved.\nDefine \"Saved\" as Missing.", lerrors.UnresolvedReference},
	} {
		t.Run(test.name, func(t *testing.T) {
			source := prefix + "    Publicly create \"Exercise\" as a Generic Function with Inputs of " + test.inputs + " and Outputs of `f64`.\n" + test.body + "\nReturn 0.\n"
			c := NewLygerService(canonicalTestOptions()).NewCompilation([]byte(source), "process-reservations.l")
			c.Make()
			if !errors.Is(c.Error, test.kind) {
				t.Fatalf("invalid Process scope: %v; want %v", c.Error, test.kind)
			}
			if test.name == "pending-local" {
				found := false
				for _, statement := range c.Scope.Ordered {
					if strings.HasPrefix(string(statement.Statement.source.Raw), "Create \"Before\"") {
						for _, tok := range statement.Statement.tokens {
							if string(tok.rawBytes) == "Saved" {
								found = true
								if tok.binding != nil {
									t.Fatal("pending ordinary local fell back to inherited Saved")
								}
							}
						}
					}
				}
				if !found {
					t.Fatal("missing pending-local reference observer")
				}
			}
		})
	}
}

func TestWordLibraryTraversalBehavior(t *testing.T) {
	const imports = "Use Storage from `storage`.\nCreate \"Checks\" as a Dictionary.\n"
	const observers = `use std::sync::atomic::{AtomicUsize, Ordering};
static READS: AtomicUsize = AtomicUsize::new(0);
static FACTORIES: std::sync::Mutex<Vec<u64>> = std::sync::Mutex::new(Vec::new());
pub fn gather() -> &'static [u64] { READS.fetch_add(1, Ordering::SeqCst); FACTORIES.lock().unwrap().push(10); &[1,2] }
pub fn callback() -> fn(&u64, &mut Vec<u64>) { FACTORIES.lock().unwrap().push(20); observed_visit }
fn observed_visit(value: &u64, seen: &mut Vec<u64>) { FACTORIES.lock().unwrap().push(100+value); visit(value,seen); }
fn visit(value: &u64, seen: &mut Vec<u64>) { seen.push(*value); }
fn ignored(value: &u64, seen: &mut Vec<u64>) -> Result<(), u8> { visit(value, seen); Err(9) }
fn panics(value: &u64, seen: &mut Vec<u64>) -> Result<(), u8> { visit(value, seen); if *value == 2 { panic!("callback panic"); } Err(9) }
fn pair(first: &u64, second: &u64, seen: &mut Vec<u64>) { seen.extend([*first,*second]); }
fn ignored_pair(first: &u64, second: &u64, seen: &mut Vec<u64>) -> Result<(), u8> { pair(first, second, seen); Err(9) }
fn find(value: &u64, seen: &mut Vec<u64>) -> Result<bool, u8> { visit(value, seen); if *value == 3 { Err(7) } else { Ok(value % 2 == 0) } }
pub fn found(seen: &mut Vec<u64>) { seen.push(99); }
pub fn missing(key: &u64, seen: &mut Vec<u64>) { seen.extend([*key,0]); }
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
			`let mut seen=vec![]; generated::visit(&[1,2,3], ignored, &mut seen); assert_eq!(seen,[1,2,3]); seen.clear(); generated::pairs(&[(1,10),(2,20)],ignored_pair,&mut seen); assert_eq!(seen,[1,10,2,20]); seen.clear(); std::panic::set_hook(Box::new(|_|{})); assert!(std::panic::catch_unwind(std::panic::AssertUnwindSafe(||generated::visit(&[1,2,3],panics,&mut seen))).is_err()); assert_eq!(seen,[1,2]);
for values in [&[][..], &[1][..]] { seen.clear(); generated::visit(values,ignored,&mut seen); assert_eq!(seen,values); }
for (values,expected) in [(&[][..],&[][..]),(&[(1,10)][..],&[1,10][..])] { seen.clear(); generated::pairs(values,ignored_pair,&mut seen); assert_eq!(seen,expected); }`},
		{"iterable-once", "Define \"Gathered\" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~crate::gather~|}<DEF>`.\n\n" +
			"Define \"Callback\" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~crate::callback~|}<DEF>`.\n\n" +
			"    Publicly create Checks's \"Exercise\" as a Generic `exercise` Function with Inputs of \"action\" as `fn(&u64, &mut Vec<u64>)` and \"context\" as `&mut Vec<u64>` and Outputs of `()`.\nStorage.Traverse a New Gathered, Action, and Context.\nReturn a New Storage's Nothing.\n\n" +
			"    Publicly create Checks's \"Factories\" as a Generic `factories` Function with Inputs of \"context\" as `&mut Vec<u64>` and Outputs of `()`.\nStorage.Traverse a New Gathered, a New Callback, and Context.\nReturn a New Storage's Nothing.\n",
			`let mut seen=vec![]; generated::exercise(visit,&mut seen); assert_eq!(seen,[1,2]); assert_eq!(READS.load(Ordering::SeqCst),1);
FACTORIES.lock().unwrap().clear(); seen.clear(); generated::factories(&mut seen); assert_eq!(seen,[1,2]); assert_eq!(READS.load(Ordering::SeqCst),2); assert_eq!(&*FACTORIES.lock().unwrap(),&[10,20,101,102]);`},
		{"discarded-success", "\n    Publicly create Checks's \"Check\" as a Generic `check` Function with Inputs of \"values\" as `&[u64]` and \"action\" as `fn(&u64, &mut Vec<u64>) -> Result<crate::Dropped,u8>` and \"context\" as `&mut Vec<u64>` and Outputs of `Result<(),u8>`.\nStorage.Inspect Values, Action, and Context.\nReturn a New Storage's Success of a New Storage's Nothing.\n\n" +
			"    Publicly create Checks's \"Audit\" as a Generic `audit` Function with Inputs of \"values\" as `&[(u64,u64)]` and \"action\" as `fn(&u64, &u64, &mut Vec<u64>) -> Result<crate::Dropped,u8>` and \"context\" as `&mut Vec<u64>` and Outputs of `Result<(),u8>`.\nStorage.Audit Values, Action, and Context.\nReturn a New Storage's Success of a New Storage's Nothing.\n",
			`let mut seen=vec![]; assert_eq!(generated::check(&[1,2,3],checked_drop,&mut seen),Err(9)); assert_eq!(&*DROP_EVENTS.lock().unwrap(),&[1,101,2]); DROP_EVENTS.lock().unwrap().clear(); assert_eq!(generated::audit(&[(1,10),(2,20),(3,30)],paired_drop,&mut seen),Err(9)); assert_eq!(&*DROP_EVENTS.lock().unwrap(),&[1,110,2]);
for (values,expected) in [(&[][..],&[][..]),(&[1][..],&[1,101][..])] { DROP_EVENTS.lock().unwrap().clear(); assert_eq!(generated::check(values,checked_drop,&mut seen),Ok(())); assert_eq!(&*DROP_EVENTS.lock().unwrap(),expected); }
for (values,expected) in [(&[][..],&[][..]),(&[(1,10)][..],&[1,110][..])] { DROP_EVENTS.lock().unwrap().clear(); assert_eq!(generated::audit(values,paired_drop,&mut seen),Ok(())); assert_eq!(&*DROP_EVENTS.lock().unwrap(),expected); }`},
		{"repeated-labels", "\n    Publicly create Checks's \"Exercise\" as a Generic `exercise` Function with Inputs of \"item\" as `&[u64]` and \"action\" as `fn(&u64, &mut Vec<u64>)` and \"outcome\" as `&mut Vec<u64>` and Outputs of `()`.\nStorage.Traverse Item, Action, and Outcome.\nStorage.Traverse Item, Action, and Outcome.\nReturn a New Storage's Nothing.\n",
			`let mut seen=vec![]; generated::exercise(&[1,2],visit,&mut seen); assert_eq!(seen,[1,2,1,2]);`},
		{"nested-labels", "Define \"Perform|-s|-ed|-ing|-er\" to `<DEF><@Body@><DEF>`.\n\n" +
			"Define \"Missing\" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~crate::missing~, Inputs: ~&[<@Key@>, <@Context@>]&~|}<DEF>`.\n" +
			"Define \"Note|-s|-d|--ing|-r\" to `<DEF>MOB{|Key: ~_~|}^KY{|Key: ~crate::found~, Inputs: ~&[<@Context@>]&~|}<DEF>`.\n\n" +
			"    Publicly create Checks's \"Exercise\" as a Generic `exercise` Function with Inputs of \"values\" as `&[u64]` and \"action\" as `fn(&u64, &u64, &mut Vec<u64>)` and \"context\" as `&mut Vec<u64>` and Outputs of `()`.\nPerform a New Storage's Iteration of Values and a New Process with Inputs of \"Item\" as `_` for Creating \"Saved\" as a New Storage's Binding of Item and then Performing a New Storage's Iteration of Values and a New Process with Inputs of \"Item\" as `_` for Creating \"Outcome\" as a New Storage.Paired Storage's Applier of Action, Saved, Item, and Context.\nReturn a New Storage's Nothing.\n\n" +
			"    Publicly create Checks's \"Pairselect\" as a Generic `pairselect` Function with Inputs of \"item\" as `&[(u64,Option<u64>)]` and \"action\" as `fn(&u64, &u64, &mut Vec<u64>)` and \"outcome\" as `&mut Vec<u64>` and Outputs of `()`.\nCreate \"Context\" as a New Storage's Binding of Outcome.\nPerform a New Storage.Paired Storage's Iteration of Item and a New Process with Inputs of \"Key\" as `&u64` and \"Item\" as `&Option<u64>` for Creating \"Savedkey\" as a New Storage's Binding of Key and then Storage.Option.Selecting Item, a New Process with Inputs of \"Outcome\" as `&u64` for Creating \"Item\" as a New Storage.Paired Storage's Applier of Action, Savedkey, Outcome, and Context, and a New Process for Creating \"Outcome\" as a New Missing of Savedkey and Context.\nReturn a New Storage's Nothing.\n\n" +
			"    Publicly create Checks's \"Inside\" as a Generic `inside` Function with Inputs of \"item\" as `Option<&[u64]>` and \"key\" as `fn(&u64, &mut Vec<u64>)` and \"outcome\" as `&mut Vec<u64>` and Outputs of `()`.\nStorage.Option.Select Item, a New Process with Inputs of \"Item\" as `&[u64]` for Storage.Traversing Item, Key, and Outcome, and a New Process for Noting Outcome.\nReturn a New Storage's Nothing.\n",
			`let mut seen=vec![]; generated::exercise(&[1,2],pair,&mut seen); assert_eq!(seen,[1,1,1,2,2,1,2,2]);
for _ in 0..2 { seen.clear(); generated::pairselect(&[(1,Some(10)),(2,None),(3,Some(30))],pair,&mut seen); assert_eq!(seen,[1,10,2,0,3,30]); }
for (values,expected) in [(Some(&[4,5][..]),&[4,5][..]),(None,&[99][..])] { seen.clear(); generated::inside(values,visit,&mut seen); assert_eq!(seen,expected); }`},
		{"continuation-process", "Define \"Note|-s|-d|--ing|-r\" to `<DEF>MOB{|Key: ~_~|}^KY{|Key: ~crate::found~, Inputs: ~&[<@Context@>]&~|}<DEF>`.\n\n" +
			"    Publicly create Checks's \"Exercise\" as a Generic `exercise` Function with Inputs of \"values\" as `&[u64]` and \"action\" as `fn(&u64, &mut Vec<u64>) -> Result<bool,u8>` and \"context\" as `&mut Vec<u64>` and Outputs of `Result<(),u8>`.\nStorage.Search Values, Action, Context, and a New Process for Noting Context.\nReturn a New Storage's Success of a New Storage's Nothing.\n",
			`let mut seen=vec![]; assert_eq!(generated::exercise(&[1,2,4],find,&mut seen),Ok(())); assert_eq!(seen,[1,2,99,4,99]); seen.clear(); assert_eq!(generated::exercise(&[2,3,4],find,&mut seen),Err(7)); assert_eq!(seen,[2,99,3]);`},
	} {
		t.Run(test.name, func(t *testing.T) {
			c := compileDefinitionSource(t, imports+test.source, "word-traversal.l")
			if test.name == "nested-labels" {
				definitions, err := c.composeExports("Checks")
				if err != nil {
					t.Fatal(err)
				}
				inputs := make(map[*record]DefinitionArgument)
				for _, rec := range c.Semantics.Subjects {
					if argument := rec.definitionArgument; argument != nil && argument.processType != nil && strings.HasSuffix(argument.source.Path, "word-traversal.l") {
						inputs[rec] = *argument
					}
				}
				labels := make(map[*sourceFunction]map[string]bool)
				for rec, input := range inputs {
					owner := rec.storageScope
					for owner != nil && owner.function == nil {
						owner = owner.parent
					}
					if owner == nil {
						t.Fatal("nested input lost its enclosing Function")
					}
					if labels[owner.function] == nil {
						labels[owner.function] = make(map[string]bool)
					}
					found := false
					for _, definition := range definitions {
						for _, binding := range definition.Bindings {
							for _, span := range binding.Sources {
								found = found || binding.Slot.Type == "Label" && binding.Emitted == input.reference && span == input.source
							}
						}
					}
					if !found || labels[owner.function][input.reference] {
						t.Fatalf("nested Process input lost its unique label/source ledger: %+v", input)
					}
					labels[owner.function][input.reference] = true
				}
				if len(inputs) != 6 {
					t.Fatalf("nested Process input identities = %d, want 6", len(inputs))
				}
			}
			testNativeDefinitionLibrary(t, c, "Checks", "traversal_extra", observers+"fn main() {"+test.observe+"}")
		})
	}
}
