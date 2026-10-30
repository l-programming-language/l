package Lyger

import (
	"fmt"
	"strings"
	"testing"
)

func TestComposedSourceFormalBindings(t *testing.T) {
	c := compileDefinitionSource(t, "Use Storage from `storage`.\nCreate \"Repeat|-s|-ed|-ing|-er\" for Processing a New Storage's Discarder of any Text and then Processing a New Storage's Discarder of it.\n", "formal.l")
	rec, err := c.exportedRecord("Repeat")
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		arguments []DefinitionArgument
		valid     bool
	}{{[]DefinitionArgument{{Kind: "Text", Text: "value"}}, true}, {nil, false}, {[]DefinitionArgument{{Kind: "Number", Text: "1"}}, false}} {
		_, err := c.instantiateCall(rec, test.arguments, make(map[*record]bool))
		if (err == nil) != test.valid {
			t.Fatalf("formal arguments %+v: %v", test.arguments, err)
		}
	}
}

func TestComposedSourceFormalEvaluatesValueOnce(t *testing.T) {
	source := "Use Storage from `storage`.\n" +
		"Define \"Current\" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~get~, Receiver: ~<@Value@>~|}<DEF>`.\n" +
		"Define \"Next\" as `<DEF>DOB{|Key: ~<@Name|Label@>~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: +, Operands: &[<@Value@>, 1.0_f64]&}Op]&~|}<DEF>`.\n" +
		"Define \"Update|-s|-d|--ing|--er\" to `<DEF>MOB{|Key: ~_~|}^KY{|Key: ~set~, Receiver: ~<@Target@>~, Inputs: ~&[<@Value@>]&~|}<DEF>`.\n" +
		"Create \"Repeat|-s|-ed|-ing|-er\" for Processing a New Storage's Discarder of any Thing and then Processing a New Storage's Discarder of it.\n\n" +
		"    Publicly create \"Tick|-s|-ed|-ing|-er\" as a Generic Function with Inputs of \"value\" as `&std::cell::Cell<f64>` and Outputs of `f64`.\nCreate \"Before\" as a New Current of Value.\nCreate \"After\" as a New Next of Before.\nUpdate Value and After.\nReturn After.\n\n" +
		"    Publicly create \"Exercise\" as a Generic Function with Inputs of \"value\" as `&std::cell::Cell<f64>` and Outputs of `f64`.\nRepeat a New Ticker of Value.\nReturn a New Current of Value.\n"
	c := compileDefinitionSource(t, source, "formal-once.l")
	testNativeDefinitionLibrary(t, c, "Exercise", "formal", `#[path="src/formal.rs"] mod generated; fn main() { let value = std::cell::Cell::new(0.0); assert_eq!(generated::exercise(&value), 1.0); }`)
}

func TestComposedTraversalBehavior(t *testing.T) {
	var source strings.Builder
	source.WriteString("Use Storage from `storage`.\nCreate \"Checks\" as a Dictionary.\n")
	for _, variant := range []struct{ name, word, values, callback, result, ending string }{
		{"Visit", "Traverse", "&[u64]", "fn(&u64, &mut Vec<u64>)", "()", "Return a New Storage's Nothing."},
		{"Pairs", "Enumerate", "&std::collections::BTreeMap<u64, u64>", "fn(&u64, &u64, &mut Vec<u64>)", "()", "Return a New Storage's Nothing."},
		{"Check", "Inspect", "&[u64]", "fn(&u64, &mut Vec<u64>) -> Result<(), u8>", "Result<(), u8>", "Return a New Storage's Success of a New Storage's Nothing."},
		{"Audit", "Audit", "&std::collections::BTreeMap<u64, u64>", "fn(&u64, &u64, &mut Vec<u64>) -> Result<(), u8>", "Result<(), u8>", "Return a New Storage's Success of a New Storage's Nothing."},
		{"Seek", "Search", "&[u64]", "fn(&u64, &mut Vec<u64>) -> Result<bool, u8>", "Result<bool, u8>", "Return a New Storage's Success of false."},
	} {
		fmt.Fprintf(&source, "\n    Publicly create Checks's %q as a Generic `%s` Function with Inputs of \"values\" as `%s` and \"action\" as `%s` and \"context\" as `&mut Vec<u64>` and Outputs of `%s`.\n", variant.name, strings.ToLower(variant.name), variant.values, variant.callback, variant.result)
		if variant.word == "Search" {
			source.WriteString("Create \"Found\" as a New Storage's Success of true.\nStorage.Search Values, Action, Context, and a New Function's Returner of Found.\n")
		} else {
			fmt.Fprintf(&source, "Storage.%s Values, Action, and Context.\n", variant.word)
		}
		source.WriteString(variant.ending + "\n")
	}
	c := compileDefinitionSource(t, source.String(), "traversal.l")
	testNativeDefinitionLibrary(t, c, "Checks", "traversal", `#[path="src/traversal.rs"] mod generated;
fn visit(value: &u64, seen: &mut Vec<u64>) { seen.push(*value); }
fn pair(key: &u64, value: &u64, seen: &mut Vec<u64>) { seen.extend([*key, *value]); }
fn check(value: &u64, seen: &mut Vec<u64>) -> Result<(), u8> { visit(value, seen); if *value == 2 { Err(9) } else { Ok(()) } }
fn audit(key: &u64, value: &u64, seen: &mut Vec<u64>) -> Result<(), u8> { pair(key, value, seen); if *key == 2 { Err(9) } else { Ok(()) } }
fn find(value: &u64, seen: &mut Vec<u64>) -> Result<bool, u8> { visit(value, seen); if *value == 3 { Err(7) } else { Ok(*value == 2) } }
fn main() {
 let mut seen = vec![]; generated::visit(&[1, 2, 3], visit, &mut seen); assert_eq!(seen, vec![1, 2, 3]);
 let pairs = std::collections::BTreeMap::from([(3, 30), (1, 10), (2, 20)]);
 seen.clear(); generated::pairs(&pairs, pair, &mut seen); assert_eq!(seen, vec![1, 10, 2, 20, 3, 30]);
 seen.clear(); assert_eq!(generated::check(&[1, 2, 3], check, &mut seen), Err(9)); assert_eq!(seen, vec![1, 2]);
 seen.clear(); assert_eq!(generated::audit(&pairs, audit, &mut seen), Err(9)); assert_eq!(seen, vec![1, 10, 2, 20]);
 seen.clear(); assert_eq!(generated::seek(&[1, 2, 3], find, &mut seen), Ok(true)); assert_eq!(seen, vec![1, 2]);
 seen.clear(); assert_eq!(generated::seek(&[3, 2], find, &mut seen), Err(7)); assert_eq!(seen, vec![3]);
 seen.clear(); assert_eq!(generated::seek(&[1], find, &mut seen), Ok(false)); assert_eq!(seen, vec![1]);
 seen.clear(); assert_eq!(generated::seek(&[], find, &mut seen), Ok(false)); assert!(seen.is_empty());
}`)
}

func TestComposedProcessAliasPreservesBody(t *testing.T) {
	source := "Use Process as \"Sequence\".\nCreate \"Perform|-s|-ed|-ing|-er\" for Processing any Sequence.\n\n" +
		"    Publicly create \"Exercise\" as a Generic `exercise` Function with No Inputs and Outputs of `f64`.\nPerform a New Process for Returning 7.\n"
	c := compileDefinitionSource(t, source, "process-alias.l")
	testNativeDefinitionLibrary(t, c, "Exercise", "process_alias", `#[path="src/process_alias.rs"] mod generated; fn main() { assert_eq!(generated::exercise(), 7.0); }`)
}
