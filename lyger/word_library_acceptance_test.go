package Lyger

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/l-programming-language/l/lerrors"
	Logiceme "github.com/l-programming-language/l/logiceme"
)

// These cases are the independent vocabulary obligations of the November word
// review. The catalog is intentionally data here, not a runtime dependency on
// a developer's Downloads directory or an implementation completion marker.
func TestWordLibraryCatalog(t *testing.T) {
	c := NewLygerService(Options{Language: "eng"}).NewCompilation([]byte("Use Storage from `storage`.\nUse IO from `io`.\n"), "word-catalog.l")
	c.Make()
	for _, test := range []struct {
		id, word, pos, forms string
		retired              bool
	}{
		{"D01", "Storage.Propagate", "verb", "Propagate Propagates Propagated Propagating Propagator", true},
		{"D02", "Storage.Function.Success", "noun", "Success", true},
		{"D03", "Storage.Function.Completion", "noun", "Completion", true},
		{"D04", "Storage.Rejection.Error", "noun", "Error", true},
		{"D05", "Storage.Rejection.Result", "noun", "Result", true},
		{"D06", "IO.Check", "verb", "Check Checks Checked Checking Checker", true},
		{"D07", "IO.Stream", "noun", "Stream", true},
		{"D08", "Storage.Indexed", "adjective", "Indexed", true},
		{"D09", "Storage.Fallible", "adjective", "Fallible", true},
		{"M01", "Retain", "verb", "Retain Retains Retained Retaining Retainer", false},
		{"M02", "Storage.App.Retain", "verb", "Retain Retains Retained Retaining Retainer", false},
		{"M03", "Storage.Iteration", "noun", "Iteration", false},
		{"M04", "Storage.Apply", "verb", "Apply Applies Applied Applying Applier", false},
		{"M05", "Storage.Traverse", "verb", "Traverse Traverses Traversed Traversing Traverser", false},
		{"M06", "Storage.Enumerate", "verb", "Enumerate Enumerates Enumerated Enumerating Enumerator", false},
		{"M07", "Storage.Examine", "verb", "Examine Examines Examined Examining Examiner", false},
		{"M08", "Storage.Review", "verb", "Review Reviews Reviewed Reviewing Reviewer", false},
		{"M09", "Storage.Inspect", "verb", "Inspect Inspects Inspected Inspecting Inspector", false},
		{"M10", "Storage.Audit", "verb", "Audit Audits Audited Auditing Auditor", false},
		{"M11", "Storage.Decide", "verb", "Decide Decides Decided Deciding Decider", false},
		{"M12", "Storage.Probe", "verb", "Probe Probes Probed Probing Prober", false},
		{"M13", "Storage.Search", "verb", "Search Searches Searched Searching Searcher", false},
		{"M14", "Storage.Require", "verb", "Require Requires Required Requiring Requirer", false},
		{"M15", "Storage.Accept", "verb", "Accept Accepts Accepted Accepting Accepter", false},
		{"M16", "Storage.Index.Check", "verb", "Check Checks Checked Checking Checker", false},
		{"M17", "Storage.Reject", "verb", "Reject Rejects Rejected Rejecting Rejector", false},
		{"M18", "Storage.Function.Yield", "verb", "Yield Yields Yielded Yielding Yielder", false},
		{"M19", "Storage.Function.Exit", "verb", "Exit Exits Exited Exiting Exiter", false},
		{"M20", "Storage.Text.Enclose", "verb", "Enclose Encloses Enclosed Enclosing Encloser", false},
		{"M21", "IO.Write", "verb", "Write Writes Wrote Writing Writer", false},
		{"M22", "IO.Print", "verb", "Print Prints Printed Printing Printer", false},
		{"N01", "Storage.Option", "dictionary", "Option", false},
		{"N02", "Storage.Option.Select", "verb", "Select Selects Selected Selecting Selector", false},
		{"N03", "Storage.Result", "dictionary", "Result", false},
		{"N04", "Storage.Result.Select", "verb", "Select Selects Selected Selecting Selector", false},
		{"N05", "Storage.Result.Expect", "verb", "Expect Expects Expected Expecting Expecter", false},
		{"N06", "Storage.Release", "verb", "Release Releases Released Releasing Releaser", false},
		{"N07", "Storage.Paired", "adjective", "Paired", false},
		{"N08", "Storage.Bytes.Extend", "verb", "Extend Extends Extended Extending Extender", false},
		{"N09", "IO.Output", "noun", "Output", false},
		{"N10", "IO.Diagnostic", "noun", "Diagnostic", false},
		{"N11", "IO.Send", "verb", "Send Sends Sent Sending Sender", false},
		{"N12", "IO.Lined", "adjective", "Lined", false},
		{"N13", "IO.Flush", "verb", "Flush Flushes Flushed Flushing Flusher", false},
	} {
		t.Run(test.id+"-"+test.word, func(t *testing.T) {
			if c.Error != nil {
				t.Fatalf("selected distribution did not compile: %v", c.Error)
			}
			rec, err := wordLibraryRecord(c, test.word)
			if test.retired {
				for _, form := range strings.Fields(test.forms) {
					name := test.word[:strings.LastIndex(test.word, ".")+1] + form
					if found, lookupErr := wordLibraryRecord(c, name); lookupErr == nil && found != nil {
						t.Errorf("retired spelling %s still resolves to %s", name, found.ID)
					}
					t.Run(form, func(t *testing.T) {
						wordLibraryRetiredSource(t, name)
					})
				}
				wordLibraryUserHomonyms(t, test.word, test.pos, strings.Fields(test.forms))
				return
			}
			if err != nil || rec == nil {
				t.Fatalf("required word %s: %v", test.word, err)
			}
			if c.Semantics.Declarations[rec] == nil {
				t.Fatal("word has no selected source declaration")
			}
			if test.pos == "dictionary" {
				kind, kindErr := instructionType(rec)
				if kindErr != nil || kind != valueDictionary {
					t.Fatalf("namespace is %s: %v", kind, kindErr)
				}
				return
			}
			word := rec.bindingWord()
			if word == nil || word.POS != test.pos {
				t.Fatalf("lexical role = %+v; want %s", word, test.pos)
			}
			prefix := rec.ID[:strings.LastIndex(rec.ID, ".")+1]
			for _, form := range strings.Fields(test.forms) {
				binding, bindErr := resolveTokenBinding([]byte(prefix+form), map[string]*record{rec.ID: rec})
				if bindErr != nil || binding == nil || binding.Record != rec || binding.POS != test.pos {
					t.Errorf("%s lost canonical lexical family: %+v, %v", form, binding, bindErr)
				}
			}
		})
	}
}

func wordLibraryRetiredSource(t *testing.T, name string) {
	t.Helper()
	if strings.HasPrefix(name, "IO.") {
		// Check and Stream used to be private in IO. Probe their original scope,
		// rather than accepting an unrelated public-import visibility rejection.
		files := dictionaryIndexFilesystem(t)
		const path = "stdlib/eng/io/dictionary.l"
		files[path] = &fstest.MapFile{Data: append(bytes.Clone(files[path].Data), []byte("\nUse "+strings.TrimPrefix(name, "IO.")+" as \"Retired\".\n")...)}
		c := NewLygerService(Options{Language: "eng", Resources: files}).NewCompilation([]byte("Use IO from `io`.\n"), "word-retired.l")
		c.Make()
		if !errors.Is(c.Error, lerrors.UnresolvedReference) || !strings.Contains(c.Error.Error(), "dictionary.l:") || !strings.Contains(strings.ToLower(c.Error.Error()), strings.ToLower(strings.TrimPrefix(name, "IO."))) {
			t.Fatalf("retired private %s did not reject at its source use: %v", name, c.Error)
		}
		return
	}
	for _, source := range []string{
		"Use Storage from `storage`.\nUse " + name + " as \"Retired\".\n",
		"Use " + name + " from `storage` as \"Retired\".\n",
	} {
		c := NewLygerService(Options{Language: "eng"}).NewCompilation([]byte(source), "word-retired.l")
		c.Make()
		if !errors.Is(c.Error, lerrors.UnresolvedReference) || !strings.Contains(c.Error.Error(), "word-retired.l:") || !strings.Contains(strings.ToLower(c.Error.Error()), strings.ToLower(name[strings.LastIndex(name, ".")+1:])) {
			t.Fatalf("retired source use %q did not reject its selected word: %v", source, c.Error)
		}
	}
}

func wordLibraryUserHomonyms(t *testing.T, name, pos string, forms []string) {
	t.Helper()
	declaration := "Define Userwords's \"" + forms[0] + "\" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~7~|}<DEF>`.\n"
	if pos == "verb" {
		family := forms[0] + "|-s|-ed|-ing|-er"
		if forms[0] == "Propagate" {
			family = "Propagate|-s|-d|--ing|--or"
		}
		declaration = "Define Userwords's \"" + family + "\" to `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~<@Value@>~|}<DEF>`.\n"
	} else if pos == "adjective" {
		declaration = "Define Userwords's \"" + forms[0] + "\" as a New Adjective describing any Process.\n"
	}
	source := "Create \"Userwords\" as a Dictionary.\n" + declaration
	for index, form := range forms {
		source += fmt.Sprintf("Use Userwords.%s as \"Selected%d\".\n", form, index)
	}
	c := compileDefinitionSource(t, source, "word-homonyms.l")
	want, err := c.exportedRecord("Userwords." + forms[0])
	if err != nil || want == nil || want.bindingWord() == nil || want.bindingWord().POS != pos {
		t.Fatalf("user homonym of %s lost its lexical role: %v", name, err)
	}
	for index := range forms {
		got, err := canonicalRecord(c.Memory[fmt.Sprintf("Selected%d", index)], c.Semantics.Aliases)
		if err != nil || got != want || c.Semantics.Declarations[got] == nil || !strings.HasSuffix(c.Semantics.Declarations[got].source.Span.Path, "word-homonyms.l") {
			t.Fatalf("user homonym of %s lost its independent source identity: %v", name, err)
		}
	}
}

// Write and the old IO helpers are unit-private declarations, whereas Print and
// the new API are dictionary members. Inspect the actual selected owner for
// retirement checks without publishing private helpers into the test program.
func wordLibraryRecord(c *Compilation, name string) (*record, error) {
	if rec, err := c.builtinRecord(name); err == nil {
		return rec, nil
	}
	if owner, member, qualified := strings.Cut(name, "."); qualified {
		for _, unit := range c.SourceUnits {
			if unit.Dictionary != owner || unit.root == nil {
				continue
			}
			binding, _, err := unit.root.lookupLocal([]byte(member))
			if err != nil {
				return nil, err
			}
			if binding != nil {
				return canonicalRecord(binding.Record, c.Semantics.Aliases)
			}
		}
	}
	return nil, fmt.Errorf("selected word %s is absent", name)
}

func TestWordLibraryRetiredHelpers(t *testing.T) {
	c := NewLygerService(canonicalTestOptions()).NewCompilation([]byte("Use IO from `io`.\nUse Runtime from `runtime`.\n"), "word-retired-helpers.l")
	c.Make()
	for _, retired := range []struct{ owner, forms string }{
		{"Storage", "Demand Demands Demanded Demanding Demander"},
		{"IO", "Slice"},
		{"Runtime", "Dispose Disposes Disposed Disposing Disposer"},
		{"Runtime", "Retain Retains Retained Retaining Retainer"},
	} {
		for _, form := range strings.Fields(retired.forms) {
			name := retired.owner + "." + form
			t.Run(name, func(t *testing.T) {
				if c.Error != nil {
					t.Fatal(c.Error)
				}
				if rec, err := wordLibraryRecord(c, name); err == nil && rec != nil {
					t.Fatalf("superseded helper or alias %s remains at %s", name, rec.ID)
				}
			})
		}
	}
	for _, name := range []string{"Storage.Optional", "Storage.Pair", "Formatting.Indexed"} {
		t.Run("retained-"+name, func(t *testing.T) {
			if c.Error != nil {
				t.Fatal(c.Error)
			}
			if rec, err := wordLibraryRecord(c, name); err != nil || rec == nil {
				t.Fatalf("unrelated live definition %s was removed: %v", name, err)
			}
		})
	}
}

func TestWordLibraryIOPublicAPI(t *testing.T) {
	for _, name := range []string{"Output", "Diagnostic", "Send", "Sender", "Lined", "Flush", "Flusher"} {
		t.Run(name, func(t *testing.T) {
			c := compileDefinitionSource(t, "Use IO's "+name+" from `io` as \"Selected\".\n", "io-api.l")
			selected, err := c.exportedRecord("Selected")
			if err != nil || selected == nil || !strings.HasPrefix(selected.ID, "IO.") {
				t.Fatalf("public IO import lost its selected dictionary owner: %v, %v", selected, err)
			}
		})
	}
}

func TestWordLibraryCarrierNamespaces(t *testing.T) {
	for _, name := range []string{"Option", "Result"} {
		t.Run(name, func(t *testing.T) {
			source := "Use Storage's " + name + " from `storage` as \"Selected\".\n"
			c := compileDefinitionSource(t, source, "word-namespace.l")
			selected, err := c.exportedRecord("Selected")
			if err != nil || selected == nil || selected.ID != "Storage."+name {
				t.Fatalf("namespace import lost its owner: %v, %v", selected, err)
			}
			kind, err := instructionType(selected)
			if err != nil || kind != valueDictionary {
				t.Fatalf("namespace became a runtime carrier: %v, %v", kind, err)
			}
			invalid := NewLygerService(canonicalTestOptions()).NewCompilation([]byte(source+"\n    Publicly create \"Rejectnamespace\" as a Generic Function with Inputs of \"value\" as Selected and Outputs of `f64`.\nReturn 0.\n"), "namespace-type.l")
			invalid.Make()
			if invalid.Error == nil || !strings.Contains(strings.ToLower(invalid.Error.Error()), "type") {
				t.Fatalf("namespace accepted as a native value type or wrong diagnostic: %v", invalid.Error)
			}
		})
	}
}

func TestWordLibraryExplicitBindings(t *testing.T) {
	c := compileDefinitionSource(t, "Use Storage from `storage`.\n", "word-bindings.l")
	t.Run("M01-borrowed-Arc", func(t *testing.T) {
		instance, err := c.instantiateBuiltin("Retain", map[string]DefinitionArgument{"Name": definitionReferenceArgument("retained"), "Target": definitionReferenceArgument("borrowed")})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(instance.Expression, "Inputs: ~&[borrowed]&~") {
			t.Fatalf("core retention must consume the supplied borrow without adding another: %s", instance.Expression)
		}
	})
	t.Run("M03-iteration-label", func(t *testing.T) {
		rec, err := c.builtinRecord("Storage.Iteration")
		if err != nil {
			t.Fatal(err)
		}
		template, err := c.recordTemplate(rec)
		if err != nil || template == nil {
			t.Fatalf("iteration structural template: %v", err)
		}
		found := false
		for _, slot := range template.Slots {
			found = found || slot.Name == "Item" && slot.Type == "Label"
		}
		if !found {
			t.Fatalf("iteration needs an explicitly bound hygienic Item label: %+v", template.Slots)
		}
	})
	t.Run("M04-apply-operands", func(t *testing.T) {
		instance, err := c.instantiateBuiltin("Storage.Apply", map[string]DefinitionArgument{"Name": definitionReferenceArgument("observed"), "Action": definitionReferenceArgument("action"), "Item": definitionReferenceArgument("selected"), "Context": definitionReferenceArgument("context")})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(instance.Expression, "Key: ~observed~") || !strings.Contains(instance.Expression, "Inputs: ~&[selected, context]&~") || strings.Contains(instance.Expression, "outcome") {
			t.Fatalf("Apply lost the explicit result/item: %s", instance.Expression)
		}
	})
}

func TestWordLibraryComposedAlgorithms(t *testing.T) {
	c := compileDefinitionSource(t, "Use Storage from `storage`.\n", "word-algorithms.l")
	for _, test := range []struct {
		name, word string
		patterns   []string
		forbidden  []string
	}{
		{"M14-lazy-Require", "Storage.Require", []string{"GI{|Key:", "Some(", "None"}, []string{"ok_or", "unwrap"}},
		{"M15-explicit-Accept", "Storage.Accept", []string{"GI{|Key:", "Ok(", "Err(", "std::mem::drop"}, []string{"Key: ~ok~", "unwrap"}},
		{"M16-lazy-Index", "Storage.Index.Check", []string{"is_finite", "fract", "usize::MAX"}, nil},
		{"M20-append-Enclose", "Storage.Text.Enclose", []string{"extend_from_slice"}, []string{"Vec::from", "join~"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			rec, err := wordLibraryRecord(c, test.word)
			if err != nil {
				t.Fatal(err)
			}
			instance, err := c.composeExport(rec)
			if err != nil {
				t.Fatal(err)
			}
			for _, pattern := range test.patterns {
				if !strings.Contains(instance.Expression, pattern) {
					t.Errorf("source composition lacks %q: %s", pattern, instance.Expression)
				}
			}
			for _, pattern := range test.forbidden {
				if strings.Contains(instance.Expression, pattern) {
					t.Errorf("source composition retains %q", pattern)
				}
			}
			if test.word == "Storage.Index.Check" {
				reject, lookupErr := c.builtinRecord("Storage.Reject")
				if lookupErr != nil || reject.function != nil {
					t.Fatalf("Index requires the shared returning source action: %v", lookupErr)
				}
				failure := strings.Index(instance.Expression, "Key: ~runtime::error_new_text~")
				branch := strings.Index(instance.Expression, "GI")
				if failure < 0 || branch < 0 || failure < branch {
					t.Error("Index must invoke shared Reject only inside a selected failure branch")
				}
			}
			if test.word == "Storage.Accept" {
				release := strings.Index(instance.Expression, "std::mem::drop")
				failure := strings.Index(instance.Expression, "runtime::error_new_text")
				if release < 0 || failure < release {
					t.Error("Accept must release the replaced error before constructing the new error")
				}
			}
			if test.word == "Storage.Text.Enclose" && strings.Count(instance.Expression, "extend_from_slice") != 2 {
				t.Error("Enclose must append value and closing through the shared byte primitive")
			}
		})
	}
}

func TestWordLibraryRejectComposition(t *testing.T) {
	c := compileDefinitionSource(t, "Use Storage from `storage`.\n", "word-reject.l")
	rec, err := c.builtinRecord("Storage.Reject")
	if err != nil || rec.function != nil {
		t.Fatalf("Reject must be the selected returning source action: %v", err)
	}
	instance, err := c.instantiateCall(rec, []DefinitionArgument{{Kind: "Text", Text: "rejected"}}, make(map[*record]bool))
	if err != nil {
		t.Fatal(err)
	}
	previous := -1
	for _, step := range []string{"Key: ~runtime::error_new_text~", "Key: ~Err~", "Statement: ~return "} {
		at := strings.Index(instance.Expression, step)
		if at <= previous {
			t.Fatalf("Reject lost ordered constructor/return step %q: %s", step, instance.Expression)
		}
		previous = at
	}
}

func TestWordLibraryLazyErrorBehavior(t *testing.T) {
	source := "Use Storage from `storage`.\nCreate \"Checks\" as a Dictionary.\n" +
		"Publicly use Storage.Require as Checks's \"Require\".\n" +
		"Publicly use Storage.Accept as Checks's \"Accept\".\n" +
		"Publicly use Storage.Index.Check as Checks's \"Index\".\n"
	// Change one selected KY call target to an observing delegate. The delegate
	// invokes the real runtime error constructor; neither the error model nor any
	// tested selection, conversion, release or return is replaced by a test stub.
	resources := dictionaryIndexFilesystem(t)
	const storage = "stdlib/eng/storage/dictionary.l"
	from := []byte("Key: ~runtime::error_new_text~")
	if bytes.Count(resources[storage].Data, from) != 1 {
		t.Fatal("Failure must own the sole runtime error-construction leaf")
	}
	resources[storage] = &fstest.MapFile{Data: bytes.Replace(resources[storage].Data, from, []byte("Key: ~crate::word_observer::error_new_text~"), 1)}
	deriveTestRuntime(t, resources)
	c := NewLygerService(Options{Resources: resources}).NewCompilation([]byte(source), "word-errors.l")
	c.Make()
	if c.Error != nil {
		t.Fatal(c.Error)
	}
	observer := `use std::sync::Mutex;
static EVENTS: Mutex<Vec<&'static str>> = Mutex::new(Vec::new());
pub(crate) fn error_new_text(message: &str) -> crate::runtime::Error { EVENTS.lock().unwrap().push("new"); crate::runtime::error_new_text(message) }
struct OldError { panic: bool }
impl Drop for OldError { fn drop(&mut self) { EVENTS.lock().unwrap().push("drop"); if self.panic { panic!("old error destructor"); } } }
fn message() -> &'static str { EVENTS.lock().unwrap().push("message"); "missing" }
pub(crate) fn exercise() {
 std::panic::set_hook(Box::new(|_| {}));
 let value = Box::new(7_u8); let address = &*value as *const u8;
 let kept = super::l_required(Some(value), message()).unwrap();
 assert_eq!(&*kept as *const u8, address); assert_eq!(&*EVENTS.lock().unwrap(), &["message"], "Require allocated a failure on success or changed message evaluation"); EVENTS.lock().unwrap().clear();
 assert_eq!(super::l_required::<u8>(None, "missing").unwrap_err().to_string(), "missing");
 assert_eq!(&*EVENTS.lock().unwrap(), &["new"]); EVENTS.lock().unwrap().clear();
 let accepted = Box::new(8_u8); let accepted_address = &*accepted as *const u8;
 assert_eq!(&*super::l_borrowed::<Box<u8>, OldError>(Ok(accepted), message()).unwrap() as *const u8, accepted_address);
 assert_eq!(&*EVENTS.lock().unwrap(), &["message"]); EVENTS.lock().unwrap().clear();
 assert_eq!(super::l_borrowed::<u8, OldError>(Err(OldError { panic: false }), "replacement").unwrap_err().to_string(), "replacement");
 assert_eq!(&*EVENTS.lock().unwrap(), &["drop", "new"]); EVENTS.lock().unwrap().clear();
 assert!(std::panic::catch_unwind(|| super::l_borrowed::<u8, OldError>(Err(OldError { panic: true }), "replacement")).is_err());
 assert_eq!(&*EVENTS.lock().unwrap(), &["drop"]); EVENTS.lock().unwrap().clear();
 for value in [0.0, -0.0, 1.0, 1024.0] { assert_eq!(super::l_index(value).unwrap(), value as usize); }
 assert!(EVENTS.lock().unwrap().is_empty(), "Index allocated a failure on success");
 let limit = usize::MAX as f64;
 for value in [f64::NAN, f64::INFINITY, f64::NEG_INFINITY, -1.0, 0.5, limit, f64::from_bits(limit.to_bits()+1)] {
  assert_eq!(super::l_index(value).unwrap_err().to_string(), "index must be a nonnegative integer");
 }
 assert_eq!(EVENTS.lock().unwrap().len(), 7);
 let below = f64::from_bits(limit.to_bits()-1);
 if below.fract() == 0.0 { assert_eq!(super::l_index(below).unwrap(), below as usize); }
 println!("word-library error assertions reached");
}`
	wordLibraryRuntimeObserver(t, c, "Checks", observer)
}

func wordLibraryRuntimeObserver(t *testing.T, c *Compilation, name, observer string) {
	t.Helper()
	declarations, err := c.composeExports(name)
	if err != nil {
		t.Fatal(err)
	}
	declarations, err = definitionClosure(declarations...)
	if err != nil {
		t.Fatal(err)
	}
	entry, err := c.futureEntry("l_main", "runtime::Result<()>", []string{
		`MOB{|Key: ~observed~|}^KY{|Key: ~crate::word_observer::exercise~|}`,
		`MOB{|Key: ~result~|}^KY{|Key: ~Ok~, Inputs: ~&[observed]&~|}`,
		`BOD^E{|Statement: ~result~|}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	declarations = append(declarations, entry)
	main := newRootObject("main")
	primary := newRootService("main")
	primary.Objects = []*object{main}
	app := &App{ID: "entry", Meta: appMetadata("entry"), Services: []*service{primary}}
	word := &w{Syllables: []*s{{CompiledBytes: []byte("A")}, {CompiledBytes: []byte(`E{|Statement: ~mod word_observer;~|}`)}}}
	for _, declaration := range declarations {
		word.Syllables = append(word.Syllables, &s{CompiledBytes: []byte(declaration.Expression)})
	}
	program := &Program{Books: []*book{{Name: "entry", Chapters: []*chapter{{Name: "main", sentences: []*sentence{{Name: "main", Words: []*w{word}}}}}}}}
	_, aml, err := renderProgram(program, []*App{app})
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := Logiceme.NewLogicemeService(c.Gram.resources).Evaluate(aml, Logiceme.GenerationOptions{IncludeRuntime: true})
	if err != nil {
		t.Fatal(err)
	}
	bundle.MapArrays[0] = append(bundle.MapArrays[0], "/src/word_observer.rs")
	bundle.ByteArrays[0] = append(bundle.ByteArrays[0], []byte(observer))
	runtimeBundleResult(t, bundle, "word-library error assertions reached\n", "")
}

func TestWordLibraryIOBehavior(t *testing.T) {
	source := "Use IO from `io`.\nUse Storage from `storage`.\nCreate \"Checks\" as a Dictionary.\n\n" +
		"    Publicly create Checks's \"Output\" as a Generic `output` Function with No Inputs and Outputs of `std::io::Stdout`.\nReturn a New IO's Output.\n\n" +
		"    Publicly create Checks's \"Diagnostic\" as a Generic `diagnostic` Function with No Inputs and Outputs of `std::io::Stderr`.\nReturn a New IO's Diagnostic.\n\n" +
		"    Publicly create Checks's \"Send\" as a Generic `send` Function with Inputs of \"writer\" as `&mut crate::Recorder` and \"payload\" as `&[u8]` and Outputs of `std::io::Result<()>`.\nReturn a New IO's Sender of Writer and Payload.\n\n" +
		"    Publicly create Checks's \"Line\" as a Generic `line` Function with Inputs of \"writer\" as `&mut crate::Recorder` and \"payload\" as `&[u8]` and Outputs of `std::io::Result<()>`.\nReturn a New IO.Lined IO's Sender of Writer and Payload.\n\n" +
		"    Publicly create Checks's \"Flush\" as a Generic `flush` Function with Inputs of \"writer\" as `&mut crate::Recorder` and Outputs of `std::io::Result<()>`.\nReturn a New IO's Flusher of Writer.\n\n" +
		"    Publicly create Checks's \"Expect\" as a Generic `expect` Function with Inputs of \"value\" as `Result<Box<u8>, &'static str>` and \"message\" as `&str` and Outputs of `Box<u8>`.\nReturn a New Storage's Result's Expecter of Value and Message.\n"
	c := compileDefinitionSource(t, source, "word-io.l")
	observer := `use std::io::{self, Write};
use std::collections::VecDeque;
enum Step { Short(usize), Interrupted, Zero, Fail }
#[derive(Default)] pub struct Recorder { bytes: Vec<u8>, calls: usize, steps: VecDeque<Step>, flushes: usize, flush_fails: bool }
impl Write for Recorder {
 fn write(&mut self, bytes: &[u8]) -> io::Result<usize> {
  self.calls += 1; assert!(self.calls <= 16, "writer failed to make bounded progress");
  let count = match self.steps.pop_front() {
   Some(Step::Short(n)) => bytes.len().min(n),
   Some(Step::Interrupted) => return Err(io::ErrorKind::Interrupted.into()),
   Some(Step::Zero) => return Ok(0),
   Some(Step::Fail) => { self.bytes.extend_from_slice(&bytes[..bytes.len().min(2).saturating_sub(1)]); return Err(io::Error::other("delivery")); },
   None => bytes.len(),
  };
  self.bytes.extend_from_slice(&bytes[..count]); Ok(count)
 }
 fn flush(&mut self) -> io::Result<()> { self.flushes += 1; if self.flush_fails { Err(io::Error::other("flush")) } else { Ok(()) } }
}
#[path="src/word_io.rs"] mod generated;
fn main() {
 let _: std::io::Stdout = generated::output(); let _: std::io::Stderr = generated::diagnostic();
 let mut raw = Recorder::default(); generated::send(&mut raw, &[0,255,b'\n']).unwrap();
 assert_eq!(raw.bytes, [0,255,b'\n']); assert_eq!(raw.calls, 1); assert_eq!(raw.flushes, 0);
 let mut line = Recorder::default(); generated::line(&mut line, b"abc").unwrap();
 assert_eq!(line.bytes, b"abc\n"); assert_eq!(line.calls, 2); assert_eq!(line.flushes, 0);
 let mut payload_failure = Recorder { steps: [Step::Fail].into(), ..Recorder::default() };
 assert_eq!(generated::line(&mut payload_failure, b"abc").unwrap_err().to_string(), "delivery");
 assert_eq!(payload_failure.bytes, b"a"); assert_eq!(payload_failure.calls, 1); assert_eq!(payload_failure.flushes, 0);
 let mut line_failure = Recorder { steps: [Step::Short(3), Step::Fail].into(), ..Recorder::default() };
 assert_eq!(generated::line(&mut line_failure, b"abc").unwrap_err().to_string(), "delivery");
 assert_eq!(line_failure.bytes, b"abc"); assert_eq!(line_failure.calls, 2); assert_eq!(line_failure.flushes, 0);
 for (name, lined, payload, steps, expected, failure, calls) in [
  ("short send", false, &b"abc"[..], vec![Step::Short(1), Step::Short(1)], &b"abc"[..], None, 3),
  ("short lined", true, &b"abc"[..], vec![Step::Short(1), Step::Short(1)], &b"abc\n"[..], None, 4),
  ("interrupted payload", true, &b"abc"[..], vec![Step::Interrupted, Step::Short(1), Step::Interrupted], &b"abc\n"[..], None, 5),
  ("interrupted newline", true, &b"abc"[..], vec![Step::Short(3), Step::Interrupted], &b"abc\n"[..], None, 3),
  ("zero send", false, &b"abc"[..], vec![Step::Zero], &b""[..], Some(io::ErrorKind::WriteZero), 1),
  ("zero payload", true, &b"abc"[..], vec![Step::Short(1), Step::Zero], &b"a"[..], Some(io::ErrorKind::WriteZero), 2),
  ("zero newline", true, &b"abc"[..], vec![Step::Short(3), Step::Zero], &b"abc"[..], Some(io::ErrorKind::WriteZero), 2),
  ("empty send", false, &b""[..], vec![], &b""[..], None, 0),
  ("empty lined", true, &b""[..], vec![Step::Interrupted], &b"\n"[..], None, 2),
  ("existing newline", true, &b"abc\n"[..], vec![], &b"abc\n\n"[..], None, 2),
 ] {
  let mut writer = Recorder { steps: steps.into(), ..Recorder::default() };
  let result = if lined { generated::line(&mut writer, payload) } else { generated::send(&mut writer, payload) };
  assert_eq!(result.err().map(|e| e.kind()), failure, "{name}");
  assert_eq!(writer.bytes, expected, "{name}"); assert_eq!(writer.calls, calls, "{name}");
  assert_eq!(writer.flushes, 0, "{name}"); assert!(writer.steps.is_empty(), "{name}");
 }
 generated::flush(&mut raw).unwrap(); assert_eq!(raw.flushes, 1); assert_eq!(raw.bytes, [0,255,b'\n']);
 raw.flush_fails = true; assert_eq!(generated::flush(&mut raw).unwrap_err().to_string(), "flush"); assert_eq!(raw.flushes, 2);
 let payload = Box::new(9); let address = &*payload as *const u8;
 assert_eq!(&*generated::expect(Ok(payload), "failure") as *const u8, address);
 let panic = std::panic::catch_unwind(|| generated::expect(Err("native"), "failed printing to stdout")).unwrap_err();
 let message = panic.downcast_ref::<String>().expect("expect panic payload");
 assert!(message.starts_with("failed printing to stdout")); assert!(message.contains("native"));
}`
	testNativeDefinitionLibrary(t, c, "Checks", "word_io", observer)
}

func TestWordLibraryBytesBehavior(t *testing.T) {
	source := "Use Storage from `storage`.\nCreate \"Checks\" as a Dictionary.\n\n" +
		"    Publicly create Checks's \"Extend\" as a Generic `extend` Function with Inputs of \"buffer\" as `&mut Vec<u8>` and \"bytes\" as `&[u8]` and Outputs of `()`.\nReturn a New Storage's Bytes's Extender of Buffer and Bytes.\n\n" +
		"    Publicly create Checks's \"Enclose\" as a Generic `enclose` Function with Inputs of \"value\" as `Vec<u8>` and \"opening\" as `&str` and \"closing\" as `&str` and Outputs of `Vec<u8>`.\nReturn a New Storage's Text's Encloser of Value, Opening, and Closing.\n"
	c := compileDefinitionSource(t, source, "word-bytes.l")
	observer := `#[path="src/word_bytes.rs"] mod generated;
fn main() {
 let source = [0, 255, b'\n']; let mut target = vec![42]; generated::extend(&mut target, &source);
 assert_eq!(target, [42,0,255,b'\n']); assert_eq!(source, [0,255,b'\n']);
 generated::extend(&mut target, &[]); assert_eq!(target, [42,0,255,b'\n']);
 assert_eq!(generated::enclose(source.to_vec(), "[", "]"), [b'[',0,255,b'\n',b']']);
 assert_eq!(generated::enclose(Vec::new(), "", ""), Vec::<u8>::new());
 assert_eq!(generated::enclose(vec![0], "雪", "é"), ["雪".as_bytes(), &[0], "é".as_bytes()].concat());
 let large = vec![253; 65537]; let enclosed = generated::enclose(large, "(", ")");
 assert_eq!(enclosed.len(), 65539); assert_eq!(enclosed[0], b'('); assert_eq!(*enclosed.last().unwrap(), b')'); assert!(enclosed[1..65538].iter().all(|b| *b == 253));
}`
	testNativeDefinitionLibrary(t, c, "Checks", "word_bytes", observer)
}

func TestWordLibraryRetentionAndReleaseBehavior(t *testing.T) {
	source := "Use Storage from `storage`.\nCreate \"Checks\" as a Dictionary.\n\n" +
		"    Publicly create Checks's \"Arc\" as a Generic `retain_arc` Function with Inputs of \"value\" as `&std::sync::Arc<[u8]>` and Outputs of `std::sync::Arc<[u8]>`.\nReturn a New Retainer of Value.\n\n" +
		"    Publicly create Checks's \"Rc\" as a Generic `retain_rc` Function with Inputs of \"value\" as `&std::rc::Rc<crate::Dropped>` and Outputs of `std::rc::Rc<crate::Dropped>`.\nReturn a New Storage's Retainer of Value.\n\n" +
		"    Publicly create Checks's \"Release\" as a Generic `release` Function with Inputs of \"value\" as `std::rc::Rc<crate::Dropped>` and Outputs of `()`.\nStorage.Release Value.\nReturn a New Storage's Nothing.\n"
	c := compileDefinitionSource(t, source, "word-retention.l")
	observer := `use std::sync::atomic::{AtomicUsize, Ordering};
static DROPS: AtomicUsize = AtomicUsize::new(0);
pub struct Dropped;
impl Drop for Dropped { fn drop(&mut self) { DROPS.fetch_add(1, Ordering::SeqCst); } }
#[path="src/word_retention.rs"] mod generated;
fn main() {
 let original: std::sync::Arc<[u8]> = std::sync::Arc::from([0,255]);
 let retained = generated::retain_arc(&original); assert!(std::sync::Arc::ptr_eq(&original, &retained)); assert_eq!(std::sync::Arc::strong_count(&original), 2);
 drop(retained); assert_eq!(std::sync::Arc::strong_count(&original), 1);
 let original = std::rc::Rc::new(Dropped); let retained = generated::retain_rc(&original);
 assert!(std::rc::Rc::ptr_eq(&original, &retained)); assert_eq!(std::rc::Rc::strong_count(&original), 2);
 generated::release(retained); assert_eq!(DROPS.load(Ordering::SeqCst), 0); assert_eq!(std::rc::Rc::strong_count(&original), 1);
 generated::release(original); assert_eq!(DROPS.load(Ordering::SeqCst), 1);
}`
	testNativeDefinitionLibrary(t, c, "Checks", "word_retention", observer)
}

func TestWordLibraryFunctionResultBindings(t *testing.T) {
	c := compileDefinitionSource(t, "Use Storage from `storage`.\n", "word-function-results.l")
	returned, err := c.builtinRecord("Function.Return")
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct{ member, operand string }{{"Yield", "Values"}, {"Exit", "Value"}} {
		t.Run(test.member, func(t *testing.T) {
			emitter := orchestrationEmitter{compilation: c}
			words, err := emitter.functionOperation(&sourceFunction{Return: returned}, test.member, map[string]DefinitionArgument{test.operand: definitionReferenceArgument("selected_value")})
			if err != nil {
				t.Fatal(err)
			}
			body := strings.Join(words, "*")
			if !strings.Contains(body, "selected_value") || !strings.Contains(body, "Key: ~Ok~") || !strings.Contains(body, "return ") {
				t.Fatalf("lost selected result or Function.Return composition: %s", body)
			}
			for _, hidden := range []string{"lReturned", "lSuccess", "lExit"} {
				if strings.Contains(body, hidden) {
					t.Errorf("result composition retains fixed binding %s", hidden)
				}
			}
		})
	}
}

func TestWordLibrarySelectorScopes(t *testing.T) {
	source := "Use Storage from `storage`.\nCreate \"Checks\" as a Dictionary.\n" +
		"Define \"Sum\" as `<DEF>DOB{|Key: ~<@Name|Label@>~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: +, Operands: &[<@Left@>, <@Right@>]&}Op]&~|}<DEF>`.\n\n" +
		"    Publicly create Checks's \"Choose\" as a Generic `choose` Function with Inputs of \"left\" as `Option<u64>` and \"right\" as `Result<u64, u8>` and Outputs of `Result<u64, u8>`.\n" +
		"Storage.Option.Select Left, a New Process with Inputs of \"Outer\" as `u64` for Storage.Result.Selecting Right, a New Process with Inputs of \"Inner\" as `u64` for Returning a New Storage's Success of a New Sum of Outer and Inner, and a New Process with Inputs of \"Failure\" as `u8` for Returning a New Storage's Rejection of Failure, and a New Process for Returning a New Storage's Success of a New Storage's Integral of 42 and `u64`.\n\n" +
		"Define \"Offer\" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~crate::take_option~|}<DEF>`.\n" +
		"Define \"Response\" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~crate::take_result~|}<DEF>`.\n" +
		"Define \"Acceptedpayload\" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~crate::accepted_payload~, Inputs: ~&[<@Value@>]&~|}<DEF>`.\n" +
		"Define \"Rejectedpayload\" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~crate::rejected_payload~, Inputs: ~&[<@Value@>]&~|}<DEF>`.\n" +
		"Define \"Missingpayload\" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~crate::missing_payload~|}<DEF>`.\n\n" +
		"    Publicly create Checks's \"Options\" as a Generic `options` Function with No Inputs and Outputs of `u8`.\n" +
		"Storage.Option.Select a New Offer, a New Process with Inputs of \"Payload\" as `crate::Payload` for Returning a New Acceptedpayload of Payload, and a New Process for Returning a New Missingpayload.\n\n" +
		"    Publicly create Checks's \"Results\" as a Generic `results` Function with No Inputs and Outputs of `u8`.\n" +
		"Storage.Result.Select a New Response, a New Process with Inputs of \"Payload\" as `crate::Payload` for Returning a New Acceptedpayload of Payload, and a New Process with Inputs of \"Failure\" as `crate::Payload` for Returning a New Rejectedpayload of Failure.\n"
	source += "\n\nDefine \"Localvalue\" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~crate::local~, Inputs: ~&[<@Tag@>]&~|}<DEF>`.\n"
	for _, test := range []struct{ name, carrier, selector, rejected string }{
		{"Optiondrops", "Option<crate::Payload>", "Option", "a New Process for Creating \"Local\" as a New Localvalue of 2 and then Storage.Releasing Caller and then Returning a New Missingpayload"},
		{"Resultdrops", "Result<crate::Payload,crate::Payload>", "Result", "a New Process with Inputs of \"Failure\" as `crate::Payload` for Creating \"Local\" as a New Localvalue of 2 and then Storage.Releasing Caller and then Returning a New Rejectedpayload of Failure"},
	} {
		source += fmt.Sprintf("\n    Publicly create Checks's %q as a Generic `%s` Function with Inputs of \"value\" as `%s` and \"caller\" as `crate::Caller` and Outputs of `u8`.\nStorage.%s.Select Value, a New Process with Inputs of \"Payload\" as `crate::Payload` for Creating \"Local\" as a New Localvalue of 1 and then Returning a New Acceptedpayload of Payload, and %s.\n", test.name, strings.ToLower(test.name), test.carrier, test.selector, test.rejected)
	}
	c := compileDefinitionSource(t, source, "selector-scopes.l")
	testNativeDefinitionLibrary(t, c, "Checks", "selectors", `use std::sync::{Mutex, atomic::{AtomicBool, Ordering}};
static PRESENT: AtomicBool = AtomicBool::new(true);
static EVENTS: Mutex<Vec<(&'static str,u8)>> = Mutex::new(Vec::new());
pub struct Payload(u8);
impl Drop for Payload { fn drop(&mut self) { EVENTS.lock().unwrap().push(("drop",self.0)); } }
pub struct Caller;
impl Drop for Caller { fn drop(&mut self) { EVENTS.lock().unwrap().push(("caller-drop",0)); } }
pub struct Local(u8);
impl Drop for Local { fn drop(&mut self) { EVENTS.lock().unwrap().push(("local-drop",self.0)); } }
pub fn local(tag: f64) -> Local { EVENTS.lock().unwrap().push(("local-new",tag as u8)); Local(tag as u8) }
fn caller() -> Caller { EVENTS.lock().unwrap().push(("caller-new",0)); Caller }
pub fn take_option() -> Option<Payload> { EVENTS.lock().unwrap().push(("option",0)); if PRESENT.load(Ordering::SeqCst) { Some(Payload(7)) } else { None } }
pub fn take_result() -> Result<Payload,Payload> { EVENTS.lock().unwrap().push(("result",0)); if PRESENT.load(Ordering::SeqCst) { Ok(Payload(8)) } else { Err(Payload(9)) } }
pub fn accepted_payload(value: Payload) -> u8 { EVENTS.lock().unwrap().push(("accepted",value.0)); value.0 }
pub fn rejected_payload(value: Payload) -> u8 { EVENTS.lock().unwrap().push(("rejected",value.0)); value.0 }
pub fn missing_payload() -> u8 { EVENTS.lock().unwrap().push(("absent",0)); 42 }
fn events(expected: &[(&'static str,u8)]) { assert_eq!(&*EVENTS.lock().unwrap(),expected); EVENTS.lock().unwrap().clear(); }
#[path="src/selectors.rs"] mod generated; fn main() {
 assert_eq!(generated::choose(Some(2), Ok(9)), Ok(11));
 assert_eq!(generated::choose(Some(7), Ok(3)), Ok(10));
 assert_eq!(generated::choose(Some(2), Err(6)), Err(6));
 assert_eq!(generated::choose(None, Ok(9)), Ok(42));
 assert_eq!(generated::choose(None, Err(6)), Ok(42));
 assert_eq!(generated::options(),7); events(&[("option",0),("accepted",7),("drop",7)]);
 assert_eq!(generated::results(),8); events(&[("result",0),("accepted",8),("drop",8)]);
 PRESENT.store(false,Ordering::SeqCst);
 assert_eq!(generated::options(),42); events(&[("option",0),("absent",0)]);
 assert_eq!(generated::results(),9); events(&[("result",0),("rejected",9),("drop",9)]);
 // Caller is constructed before selection and referenced only in the other arm.
 // Its normal drop remains observable even when that arm's Local never exists.
 for result in [false,true] {
  let owned=caller();
  let value=if result { generated::resultdrops(Ok(Payload(7)),owned) } else { generated::optiondrops(Some(Payload(7)),owned) };
  assert_eq!(value,7); events(&[("caller-new",0),("local-new",1),("accepted",7),("drop",7),("local-drop",1),("caller-drop",0)]);
 }
 assert_eq!(generated::optiondrops(None,caller()),42);
 events(&[("caller-new",0),("local-new",2),("caller-drop",0),("absent",0),("local-drop",2)]);
 assert_eq!(generated::resultdrops(Err(Payload(9)),caller()),9);
 events(&[("caller-new",0),("local-new",2),("caller-drop",0),("rejected",9),("drop",9),("local-drop",2)]);
}`)
}
