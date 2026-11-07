package Lyger

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func libraryFoundationUnit(t *testing.T, c *Compilation, name string) *SourceUnit {
	t.Helper()
	var found *SourceUnit
	for _, unit := range c.SourceUnits {
		if unit.Dictionary == name {
			if found != nil {
				t.Fatalf("duplicate %s source unit", name)
			}
			found = unit
		}
	}
	if found == nil {
		t.Fatalf("missing %s source unit", name)
	}
	return found
}

func libraryFoundationBinding(t *testing.T, c *Compilation, unit *SourceUnit, name string) *record {
	t.Helper()
	binding, _, err := unit.root.lookupLocal([]byte(name))
	if err != nil || binding == nil {
		t.Fatalf("%s private binding %s: %v", unit.Dictionary, name, err)
	}
	canonical, err := canonicalRecord(binding.Record, c.Semantics.Aliases)
	if err != nil || canonical == nil {
		t.Fatalf("%s canonical binding %s: %v", unit.Dictionary, name, err)
	}
	return canonical
}

// These are source-loading boundaries: native runtime:: references in Storage
// do not cause Runtime's supervisor to become a source dependency.
func TestLibraryFoundationSourceLayout(t *testing.T) {
	c := compileDefinitionSource(t, "Use Storage from `storage`.\nUse Formatting from `format`.\n", "library-layout.l")
	storage := libraryFoundationUnit(t, c, "Storage")
	format := libraryFoundationUnit(t, c, "Formatting")
	if storage.Origin.Name != "builtins/eng/storage/dictionary.l" || format.Origin.Name != "stdlib/eng/format/dictionary.l" {
		t.Fatalf("foundation entries remain in the old layout: %s, %s", storage.Origin.Name, format.Origin.Name)
	}
	want := []string{"bindings.l", "values.l", "control.l", "display.l", "collections.l", "program.l", "traversal.l", "execution.l"}
	var includes []string
	for _, occurrence := range c.SourceOccurrences {
		if occurrence.Unit == storage && occurrence.Origin.Name != storage.Origin.Name {
			includes = append(includes, strings.TrimPrefix(occurrence.Origin.Name, "builtins/eng/storage/"))
		}
	}
	if strings.Join(includes, ",") != strings.Join(want, ",") {
		t.Fatalf("ordered same-unit Storage includes = %v; want %v", includes, want)
	}
	for _, unit := range c.SourceUnits {
		if unit.Dictionary == "Runtime" || unit.Dictionary == "Numeric" {
			t.Fatalf("Formatting loaded unrelated source unit %s", unit.Dictionary)
		}
	}
	for _, statement := range c.SourceStatements {
		if strings.HasPrefix(statement.Origin.Name, "builtins/eng/storage/") && statement.Unit != storage {
			t.Fatalf("Storage include %s created a different lexical owner", statement.Origin.Name)
		}
		if strings.HasPrefix(statement.Origin.Name, "runtime/") {
			t.Fatalf("Formatting source-loaded process policy %s", statement.Origin.Name)
		}
	}
	for _, name := range []string{"Sum", "Agreement", "Choice", "First", "Second", "Integer", "Indexed", "Unsigned"} {
		if binding, _, err := c.Scope.Root.lookupLocal([]byte(name)); err != nil || binding != nil {
			t.Fatalf("foundation helper %s escaped into the program: %v", name, err)
		}
	}
	previous := -1
	for _, test := range []struct{ name, file string }{
		{"Storage.Binding", "bindings.l"}, {"Storage.Pointer", "values.l"},
		{"Storage.Require", "control.l"}, {"Storage.Number.Render", "display.l"},
		{"Storage.Index.Check", "collections.l"}, {"Storage.Program.Create", "program.l"},
		{"Storage.Environment.Create", "traversal.l"}, {"Storage.Function.Suspend", "execution.l"},
	} {
		rec, err := c.builtinRecord(test.name)
		if err != nil {
			t.Fatal(err)
		}
		declaration := c.Semantics.Declarations[rec]
		if declaration == nil || declaration.source.Unit != storage || declaration.source.Origin.Name != "builtins/eng/storage/"+test.file {
			t.Fatalf("%s lost its included source and canonical Storage owner", test.name)
		}
		index := -1
		for i, statement := range c.Statements {
			if statement == declaration {
				index = i
				break
			}
		}
		if index <= previous {
			t.Fatalf("Storage declaration order changed at %s: %d after %d", test.name, index, previous)
		}
		previous = index
	}
}

func TestLibrarySharedFoundationsHaveCanonicalOwners(t *testing.T) {
	c := compileDefinitionSource(t, "Use Storage from `storage`.\nUse Formatting from `format`.\nUse Runtime from `runtime`.\n", "shared-foundations.l")
	format := libraryFoundationUnit(t, c, "Formatting")
	runtime := libraryFoundationUnit(t, c, "Runtime")
	for _, test := range []struct{ local, owner string }{
		{"Sum", "Storage.Sum"}, {"Agreement", "Storage.Agreement"}, {"Choice", "Storage.Choice"},
		{"Alternative", "Storage.Result.Alternative"},
		{"First", "Storage.Tuple.First"}, {"Second", "Storage.Tuple.Second"},
		{"Presence", "Storage.Presence"}, {"Equality", "Storage.Equality"},
		{"Description", "Storage.Description"}, {"Encoding", "Storage.Encoding"},
		{"Length", "Storage.Length"}, {"Binding", "Storage.Binding"},
		{"Mutable", "Storage.Mutable"}, {"Present", "Storage.Present"}, {"Absent", "Storage.Absent"},
	} {
		t.Run(test.local, func(t *testing.T) {
			owner, err := c.builtinRecord(test.owner)
			if err != nil {
				t.Fatal(err)
			}
			for _, unit := range []*SourceUnit{format, runtime} {
				if selected := libraryFoundationBinding(t, c, unit, test.local); selected != owner {
					t.Fatalf("%s.%s selects %s, want canonical %s", unit.Dictionary, test.local, selected.ID, owner.ID)
				}
			}
		})
	}
}

func TestLibraryFormatterSelectionReachesRuntimeAndIO(t *testing.T) {
	files := dictionaryIndexFilesystem(t)
	const name = "stdlib/eng/format/dictionary.l"
	file := files[name]
	if file == nil {
		t.Fatalf("missing relocated formatter %s", name)
	}
	from, to := []byte("`format_number`"), []byte("`selected_format_number`")
	if bytes.Count(file.Data, from) != 1 {
		t.Fatal("formatter must own one selected native declaration")
	}
	files[name] = &fstest.MapFile{Data: bytes.Replace(file.Data, from, to, 1)}
	result, err := NewLygerService(Options{Resources: files}).Compile([]byte("Use IO from `io`.\nUse IO's Print from `io` as \"Print\".\nUse Runtime from `runtime`.\nCreate \"Worker\" as a New Basic App.\nThe App Prints 7.\n"), filepath.Join(t.TempDir(), "selected-format.l"))
	if err != nil {
		t.Fatal(err)
	}
	c := result.Compilation
	var selected *record
	for _, name := range []string{"IO", "Runtime"} {
		unit := libraryFoundationUnit(t, c, name)
		function := libraryFoundationBinding(t, c, unit, "Format")
		if function.function == nil || function.function.Constructor.Name != "selected_format_number" || function.function.Header.source.Unit.Dictionary != "Formatting" {
			t.Fatalf("%s bypassed the selected Formatting definition", name)
		}
		if selected != nil && selected != function {
			t.Fatal("Runtime and IO selected distinct formatting implementations")
		}
		selected = function
		var aml []byte
		if name == "IO" {
			units, err := result.ApplicationUnits()
			if err != nil || len(units) != 1 {
				t.Fatalf("selected IO application: %v", err)
			}
			aml = units[0].BackendAML
		} else {
			aml, err = c.ComposeLibrary(name, "selected_runtime")
			if err != nil {
				t.Fatal(err)
			}
		}
		if !bytes.Contains(aml, []byte("Key: ~selected_format_number~")) {
			t.Fatalf("%s composition omitted the selected formatter", name)
		}
	}
}

// Formatter casts retain their private unary signatures and replacement
// descriptors. Runtime casts select the generic Storage operation: changing
// that provider leaf changes Runtime, without rebinding the formatter's targets.
func TestLibraryCastOwnershipAndDescriptorTargets(t *testing.T) {
	const formatPath = "stdlib/eng/format/dictionary.l"
	const marker = "Operands: &[<@Value@>, <@Type|Type@>]&"
	const replacement = "Operands: &[(<@Value@>), <@Type|Type@>]&"
	files := dictionaryIndexFilesystem(t)
	formatFile := files[formatPath]
	if formatFile == nil {
		t.Fatalf("missing relocated formatter %s", formatPath)
	}
	probe := "\n    Publicly create Formatting's \"Casts\" as a Generic `format_casts` Function with Inputs of \"value\" as `f64` and Outputs of `(i32, u128, usize, u64)`.\n" +
		"Create \"Plain\" as a New Integer of Value.\nCreate \"Large\" as a New Wide Integer of Value.\n" +
		"Create \"Index\" as a New Indexed Integer of Value.\nCreate \"Unsignedvalue\" as a New Unsigned Integer of Value.\n" +
		"Return a New Storage's Tuple of Plain, Large, Index, and Unsignedvalue.\n"
	files[formatPath] = &fstest.MapFile{Data: append(bytes.Clone(formatFile.Data), []byte(probe)...)}
	for _, changed := range []bool{false, true} {
		if changed {
			found := 0
			for name, file := range files {
				if !strings.HasPrefix(name, "builtins/eng/storage/") || !bytes.Contains(file.Data, []byte(marker)) {
					continue
				}
				found += bytes.Count(file.Data, []byte(marker))
				files[name] = &fstest.MapFile{Data: bytes.ReplaceAll(file.Data, []byte(marker), []byte(replacement))}
			}
			if found != 1 {
				t.Fatalf("selected Storage.Integral definitions = %d; want one", found)
			}
		}
		result, err := NewLygerService(Options{Resources: files}).Compile([]byte("Use Formatting from `format`.\nUse Runtime from `runtime`.\n"), filepath.Join(t.TempDir(), "casts.l"))
		if err != nil {
			t.Fatal(err)
		}
		c := result.Compilation
		unit := libraryFoundationUnit(t, c, "Formatting")
		integer := libraryFoundationBinding(t, c, unit, "Integer")
		for _, name := range []string{"Wide", "Indexed", "Unsigned"} {
			descriptor := c.Semantics.Descriptors[libraryFoundationBinding(t, c, unit, name)]
			if descriptor == nil || !descriptor.Targets[nominalIdentity{definition: integer}] {
				t.Fatalf("%s lost the formatter's canonical Integer target", name)
			}
		}
		aml, err := c.ComposeLibrary("Formatting.Casts", "casts")
		if err != nil {
			t.Fatal(err)
		}
		value := "value"
		for _, target := range []string{"i32", "u128", "usize", "u64"} {
			if !bytes.Contains(aml, []byte("Operator: as, Operands: &["+value+", "+target+"]&")) {
				t.Fatalf("formatter cast to %s changed its replacement semantics (Storage changed=%t): %s", target, changed, aml)
			}
		}
		runtimeAML, err := c.ComposeLibrary("Runtime.Error.Identify", "runtime_cast")
		if err != nil {
			t.Fatal(err)
		}
		if observed := bytes.Contains(runtimeAML, []byte("Operator: as, Operands: &[(")); observed != changed {
			t.Fatalf("Runtime cast bypassed selected Storage.Integral: changed=%t, observed=%t", changed, observed)
		}
		if !changed {
			testNativeDefinitionLibrary(t, c, "Formatting.Casts", "casts", `#[path="src/casts.rs"] mod generated; fn main() {
 assert_eq!(generated::format_casts(7.0), (7,7,7,7));
 assert_eq!(generated::format_casts(4294967296.0), (i32::MAX,4294967296,4294967296,4294967296));
 assert_eq!(generated::format_casts(-1.0), (-1,0,0,0));
}`)
		}
	}
}

func TestLibrarySharedFoundationConsumerNative(t *testing.T) {
	source := "Use Storage from `storage`.\nCreate \"Consumer\" as a Dictionary.\n\n" +
		"    Publicly create Consumer's \"Exercise\" as a Generic `exercise` Function with Inputs of \"enabled\" as `bool` and \"left\" as `u64` and \"right\" as `u64` and Outputs of `(u64, bool, u64)`.\n" +
		"Create \"Added\" as a New Storage's Sum of Left and Right.\n" +
		"Create \"Agreed\" as a New Storage's Agreement of Enabled and true.\n" +
		"Create \"Chosen\" as a New Storage's Choice of Enabled, Left, and Right.\n" +
		"Create \"Pair\" as a New Storage's Tuple of Added and Chosen.\n" +
		"Create \"First\" as a New Storage's Tuple's First of Pair.\n" +
		"Create \"Second\" as a New Storage's Tuple's Second of Pair.\n" +
		"Return a New Storage's Tuple of First, Agreed, and Second.\n\n" +
		"    Publicly create Consumer's \"Fallback\" as a Generic `fallback` Function with Inputs of \"value\" as `Result<u64, u8>` and \"other\" as `u64` and Outputs of `u64`.\n" +
		"Return a New Storage's Result's Alternative of Value and Other.\n\n" +
		"    Publicly create Consumer's \"Required\" as a Generic `required` Function with Inputs of \"value\" as `Option<Box<u8>>` and Outputs of `Box<u8>`.\n" +
		"Return a New Storage's Result's Expecter of Value and `present payload`.\n"
	c := compileDefinitionSource(t, source, "independent-foundation-consumer.l")
	testNativeDefinitionLibrary(t, c, "Consumer", "foundation_consumer", `#[path="src/foundation_consumer.rs"] mod generated;
fn main() {
 assert_eq!(generated::exercise(true, 2, 3), (5, true, 2));
 assert_eq!(generated::exercise(false, 2, 3), (5, false, 3));
 assert_eq!(generated::fallback(Ok(7), 9), 7);
 assert_eq!(generated::fallback(Err(1), 9), 9);
 let payload = Box::new(7); let pointer = (&*payload) as *const u8;
 let result = generated::required(Some(payload));
 assert_eq!((&*result) as *const u8, pointer);
 std::panic::set_hook(Box::new(|_| {}));
 let failure = std::panic::catch_unwind(|| generated::required(None)).unwrap_err();
 let text = failure.downcast_ref::<String>().map(String::as_str).or_else(|| failure.downcast_ref::<&str>().copied());
 assert_eq!(text, Some("present payload"));
}`)
}

func TestLibraryFunctionTargetSharesTupleProjection(t *testing.T) {
	files := dictionaryIndexFilesystem(t)
	compile := func() *Compilation {
		t.Helper()
		result, err := NewLygerService(Options{Resources: files}).Compile([]byte("Use Storage from `storage`.\n"), filepath.Join(t.TempDir(), "function-target.l"))
		if err != nil {
			t.Fatal(err)
		}
		return result.Compilation
	}
	c := compile()
	target, err := c.builtinRecord("Storage.Function.Target")
	if err != nil {
		t.Fatal(err)
	}
	second, err := c.builtinRecord("Storage.Tuple.Second")
	if err != nil {
		t.Fatal(err)
	}
	if target != second {
		t.Fatal("Function.Target duplicates the canonical Tuple.Second projection")
	}
	const path = "builtins/eng/storage/values.l"
	const from = "Value: ~<@Value@>.1~"
	const to = "Value: ~(<@Value@>.1)~"
	if bytes.Count(files[path].Data, []byte(from)) != 1 {
		t.Fatal("Tuple.Second must own one selected projection")
	}
	files[path] = &fstest.MapFile{Data: bytes.Replace(files[path].Data, []byte(from), []byte(to), 1)}
	c = compile()
	aml, err := c.ComposeLibrary("Storage.Function.Copy", "function_target")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(aml, []byte("Value: ~(value.1)~")) {
		t.Fatal("Function.Copy bypassed the selected canonical tuple projection")
	}
}
