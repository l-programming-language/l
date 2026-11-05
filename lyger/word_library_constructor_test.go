package Lyger

import (
	"path/filepath"

	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	"github.com/l-programming-language/l/lerrors"
	"strings"
	"testing"

	Logiceme "github.com/l-programming-language/l/logiceme"
)

const wordLibraryParametricConstructor = "Define \"Parameterised\" as `<DEF>FW{|Func{Key: <@Name|Text@>, Inputs: <@Inputs|Props@>, Outputs: <@Outputs|Texts@>}Func, GenericTypes: ~<@Types|Type@>~|}^(<@Process@>)<DEF>`.\n" +
	"Use Function's Return as Parameterised's \"Return\".\nUse Function's Invoke as Parameterised's \"Invoke\".\n"

func TestWordLibraryLeadingCompositionGroup(t *testing.T) {
	for source, want := range map[string]string{
		"W^((BOD^E*MOB^KY)*BOD^E)": "W^(BOD^E*MOB^KY*BOD^E)",
		"(MOB^KY*BOD^E)":           "MOB^KY*BOD^E",
	} {
		got, err := normalizeDefinitionGroups(source)
		if err != nil || got != want {
			t.Fatalf("normalize %q: %q, %v; want %q", source, got, err, want)
		}
	}
}

func TestWordLibraryConstructorLabelArity(t *testing.T) {
	for _, tc := range []struct {
		name, definition    string
		constructor, action int
	}{
		{"explicit operand", `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~assign~, Inputs: ~&[<@Destination|Label@>, <@Value@>]&~|}<DEF>`, 2, 3},
		{"scoped pattern", `<DEF>E{|Statement: ~Some(<@Payload|Label@>)~|}*<@Body@><DEF>`, 1, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			record := indexedTestWord(t, "Attach", `"Attach|-s|-ed|-ing|-er"`, false)
			record.Value.Lexeme.Value = tc.definition
			token := &token{binding: &tokenBinding{Record: record}}
			if got, err := definitionGrammarSignature(token, true, false); err != nil || !got.Fixed || got.Count != tc.constructor {
				t.Fatalf("constructor consumes %+v authored operands, want %d", got, tc.constructor)
			}
			if got, err := definitionGrammarSignature(token, false, false); err != nil || !got.Fixed || got.Count != tc.action {
				t.Fatalf("action consumes %+v authored operands, want %d", got, tc.action)
			}
		})
	}
}

func TestWordLibraryExplicitConstructorParameters(t *testing.T) {
	source := wordLibraryParametricConstructor + "\n" +
		"    Publicly create \"Carry|-ies|-ied|-ying|-ier\" as a Generic Parameterised of `T` with Inputs of \"value\" as `T` and Outputs of `T`.\nReturn Value.\n"
	c := compileDefinitionSource(t, source, "explicit-parameters.l")
	definition, err := c.ComposeDefinition("Carry")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"GenericTypes: ~T~", "Type: T", "Statement: ~return value;~"} {
		if !strings.Contains(definition.Expression, want) {
			t.Fatalf("explicit constructor parameters lost %q: %s", want, definition.Expression)
		}
	}
	generated := wordLibraryGeneratedSource(t, c, "Carry")
	if !strings.Contains(generated, "fn carry<T>(") || !strings.Contains(generated, "return value;") {
		t.Fatalf("generic declaration was not composed: %s", generated)
	}
}

func TestWordLibraryActorDescriptorBase(t *testing.T) {
	source := "Define \"Transmit|-s|-ted|-ting|-ter\" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~std::convert::identity~, Inputs: ~&[<@Value@>]&~|}<DEF>`.\n" +
		"Define \"Tracked\" as a New Adjective describing any Transmitter.\n" +
		"It describes any Transmitter as `<DEF>(BOD^E{|Statement: ~/* outer before */~|}*<@Transmitter@>*BOD^E{|Statement: ~let _ = &<@Name|Label@>; /* outer after */~|})<DEF>`.\n" +
		"Define \"Observed\" as a New Adjective describing any Transmitter.\n" +
		"It describes any Transmitter as `<DEF>(BOD^E{|Statement: ~let _ = &<@Value@>; /* inner before */~|}*<@Transmitter@>*BOD^E{|Statement: ~/* inner after */~|})<DEF>`.\n\n" +
		"    Publicly create \"Exercise\" as a Generic Function with Inputs of \"value\" as `f64` and Outputs of `f64`.\n" +
		"Create \"Receipt\" as a New Tracked Observed Transmitter of Value.\nReturn Receipt.\n"
	c := compileDefinitionSource(t, source, "actor-descriptor.l")
	generated := wordLibraryGeneratedSource(t, c, "Exercise")
	if strings.Count(generated, "std::convert::identity(value)") != 1 || !strings.Contains(generated, "let _ = &receipt;") || !strings.Contains(generated, "return receipt;") {
		t.Fatalf("descriptor did not preserve its base result and explicit name: %s", generated)
	}
	previous := -1
	for _, marker := range []string{"/* outer before */", "/* inner before */", "std::convert::identity(value)", "/* inner after */", "/* outer after */"} {
		at := strings.Index(generated, marker)
		if at <= previous {
			t.Fatalf("descriptor wrapper order lost %q: %s", marker, generated)
		}
		previous = at
	}
}

func TestWordLibraryFunctionActorDescriptorBase(t *testing.T) {
	source := wordLibraryParametricConstructor + "\n" +
		"Define \"Polymorphic\" as a New Adjective describing any Parameterised.\nIt describes a Generic Parameterised of `T`.\n\n" +
		"    Publicly create \"Transmit|-s|-ted|-ting|-ter\" as a Generic Polymorphic Parameterised with Inputs of \"value\" as `T` and Outputs of `T`.\nReturn Value.\n\n" +
		"Define \"Tracked\" as a New Adjective describing any Transmitter.\n" +
		"It describes any Transmitter as `<DEF>(<@Transmitter@>*BOD^E{|Statement: ~let _ = &<@Name|Label@>;~|})<DEF>`.\n\n" +
		"    Publicly create \"Exercise\" as a Generic Function with Inputs of \"value\" as `String` and Outputs of `String`.\n" +
		"Create \"Receipt\" as a New Tracked Transmitter of Value.\nReturn Receipt.\n"
	c := compileDefinitionSource(t, source, "function-actor-descriptor.l")
	generated := wordLibraryGeneratedSource(t, c, "Exercise")
	if strings.Count(generated, "(transmit)(value)") != 1 || !strings.Contains(generated, "fn transmit<T>(") || !strings.Contains(generated, "let _ = &receipt;") || !strings.Contains(generated, "return receipt;") {
		t.Fatalf("Function descriptor lost its base invocation, generic declaration, or result: %s", generated)
	}
}

func TestWordLibrarySourceActorDescriptorBase(t *testing.T) {
	source := wordLibraryParametricConstructor + "\n" +
		"    Publicly create \"Transmit|-s|-ted|-ting|-ter\" as a Generic Parameterised of `T: ?Sized` with Inputs of \"context\" as `&mut T` and \"value\" as `u64` and Outputs of `u64`.\nReturn Value.\n\n" +
		"    Create \"Complete|-s|-d|--ing|-r\" as a Generic Parameterised of `T: ?Sized` with Inputs of \"context\" as `&mut T` and \"receipt\" as `u64` and Outputs of `u64`.\nReturn Receipt.\n" +
		"Publicly define \"Tracked\" as a New Adjective describing any Transmitter.\n" +
		"It describes any Transmitter as a New Completer of any Thing and any Transmitter.\n\n" +
		"    Publicly create \"Exercise\" as a Generic Function with Inputs of \"context\" as `&mut u64` and \"value\" as `u64` and Outputs of `u64`.\n" +
		"Create \"Receipt\" as a New Tracked Transmitter of Context and Value.\nReturn Receipt.\n"
	c := compileDefinitionSource(t, source, "source-actor-descriptor.l")
	generated := wordLibraryGeneratedSource(t, c, "Exercise")
	if strings.Count(generated, "(transmit)(context, value)") != 1 || !strings.Contains(generated, "fn transmit<T: ?Sized>(") || !strings.Contains(generated, "fn complete<T: ?Sized>(") || !strings.Contains(generated, "(complete)(context, ") || !strings.Contains(generated, "return receipt;") {
		t.Fatalf("source descriptor lost the retained base, checked inputs, helper dependency, or result: %s", generated)
	}
	if strings.Index(generated, "(transmit)(context, value)") > strings.Index(generated, "(complete)(context, ") {
		t.Fatalf("base must finish before the wrapper borrows the original context: %s", generated)
	}
}

func TestWordLibrarySourceActorDescriptorNative(t *testing.T) {
	source := wordLibraryParametricConstructor + "\n" +
		"Define \"Mark|-s|-ed|-ing|-er\" to `<DEF>BOD^E{|Statement: ~std::io::Write::write_all(<@Target@>, b\"base\").unwrap();~|}<DEF>`.\n" +
		"Define \"Append|-s|-ed|-ing|-er\" to `<DEF>BOD^E{|Statement: ~std::io::Write::write_all(<@Target@>, b\"wrapped\").unwrap();~|}<DEF>`.\n\n" +
		"    Publicly create \"Transmit|-s|-ted|-ting|-ter\" as a Generic Parameterised of `T: std::io::Write + ?Sized` with Inputs of \"context\" as `&mut T` and \"value\" as `u64` and Outputs of `u64`.\nMark Context.\nReturn Value.\n\n" +
		"    Create \"Complete|-s|-d|--ing|-r\" as a Generic Parameterised of `T: std::io::Write + ?Sized` with Inputs of \"context\" as `&mut T` and \"receipt\" as `u64` and Outputs of `u64`.\nAppend Context.\nReturn Receipt.\n" +
		"Publicly define \"Tracked\" as a New Adjective describing any Transmitter.\n" +
		"It describes any Transmitter as a New Completer of any Thing and any Transmitter.\n\n" +
		"    Publicly create \"Exercise\" as a Generic Function with Inputs of \"context\" as `&mut dyn std::io::Write` and \"value\" as `u64` and Outputs of `u64`.\n" +
		"Create \"Receipt\" as a New Tracked Transmitter of Context and Value.\nReturn Receipt.\n"
	c := compileDefinitionSource(t, source, "source-actor-descriptor-native.l")
	testNativeDefinitionLibrary(t, c, "Exercise", "exercise", `#[path="src/exercise.rs"] mod generated; fn main() { let mut sink = Vec::<u8>::new(); assert_eq!(generated::exercise(&mut sink, 17), 17); assert_eq!(sink, b"basewrapped"); }`)
}

func TestWordLibraryNamedSourceWrapperEvaluatesForwardedInputOnce(t *testing.T) {
	source := wordLibraryParametricConstructor + "\n" +
		"Define \"Acquire\" as `<DEF>KY{|Key: ~acquire_context~|}<DEF>`.\n\n" +
		"    Publicly create \"Transmit|-s|-ted|-ting|-ter\" as a Generic Parameterised of `T: Copy` with Inputs of \"context\" as `T` and \"value\" as `u64` and Outputs of `u64`.\nReturn Value.\n\n" +
		"    Create \"Complete|-s|-d|--ing|-r\" as a Generic Parameterised of `T: Copy` with Inputs of \"context\" as `T` and \"receipt\" as `u64` and Outputs of `u64`.\nReturn Receipt.\n" +
		"Publicly define \"Tracked\" as a New Adjective describing any Transmitter.\n" +
		"It describes any Transmitter as a New Completer of any Thing and any Transmitter.\n"
	c := compileDefinitionSource(t, source, "named-source-wrapper.l")
	acquire, err := c.ComposeDefinition("Acquire")
	if err != nil {
		t.Fatal(err)
	}
	descriptor := c.Semantics.Descriptors[c.Memory["Tracked"]]
	if descriptor == nil || len(descriptor.Rules) != 1 {
		t.Fatal("source descriptor was not registered")
	}
	defer c.definitionScope()()
	node := &constructionNode{Definition: c.Memory["Transmit"], Wrappers: descriptor.Rules}
	instance, err := c.instantiateNamedConstruction(node, map[string]DefinitionArgument{
		"context": {Kind: definitionArgDefinition.String(), Definition: &acquire},
		"value":   definitionReferenceArgument("value"),
	}, make(map[string]bool), make(map[*record]bool))
	if err != nil {
		t.Fatal(err)
	}
	var expression strings.Builder
	for _, step := range definitionSteps(instance) {
		expression.WriteString(step.Expression)
	}
	if strings.Count(expression.String(), "acquire_context") != 1 {
		t.Fatalf("forwarded original input was evaluated again: %s", expression.String())
	}
}

func TestWordLibrarySourceWrapperProjectionPreservesTarget(t *testing.T) {
	source := "Define \"Relay|-s|-ed|-ing|-er\" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~std::convert::identity~, Inputs: ~&[<@Value@>]&~|}<DEF>`.\n" +
		"Define \"Finish|-es|-ed|-ing|-er\" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~finish~, Inputs: ~&[<@Context@>, <@Receipt@>]&~|}<DEF>`.\n" +
		"Define \"Observed\" as a New Adjective describing any Relayer.\n" +
		"It describes any Relayer as a New Finisher of any Thing and any Relayer.\n"
	analysis := analyzeContractSource(t, NewLygerService(canonicalTestOptions()), []byte(source), sourceAdjectivePath, ContractAnalysisOptions{})
	projection := sourceProjectionNamed(t, analysis, "Observed")
	if projection.After.Constructor != "" || len(projection.After.Wrappers) != 1 || len(projection.Footprints) != 1 || projection.Footprints[0].Target.Path[0] != "wrappers" {
		t.Fatalf("source wrapper changed its target in symbolic projection: %+v", projection)
	}
}

func TestWordLibraryAliasedFunctionBoundary(t *testing.T) {
	source := "Define \"Choose|-s|-d|--ing|-r\" to `<DEF>BOG^GI{|Key: ~<@Condition@>~|}^(E{|Statement: ~true~|}*<@Left@>*E{|Statement: ~false~|}*<@Right@>)<DEF>`.\n\n" +
		"    Publicly create \"Forward\" as a Generic Function with Inputs of \"first\" as `u64` and \"second\" as `u64` and Outputs of `u64`.\nReturn Second.\n\n" +
		"Use Forward as \"Routed\".\n\n" +
		"    Publicly create \"Exercise\" as a Generic Function with Inputs of \"value\" as `u64` and Outputs of `u64`.\n" +
		"Choose true, a New Process for Returning a New Routed of Value and Value, and a New Process for Returning Value.\n"
	c := compileDefinitionSource(t, source, "aliased-function-boundary.l")
	generated := wordLibraryGeneratedSource(t, c, "Exercise")
	if strings.Count(generated, "(forward)(value, value)") != 1 {
		t.Fatalf("alias lost its selected Function or consumed a sibling body: %s", generated)
	}
}

func TestWordLibrarySourceSequenceRetainsPrerequisite(t *testing.T) {
	source := "Define \"Next\" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~crate::next_value~|}<DEF>`.\n" +
		"Define \"Mark|-s|-ed|-ing|-er\" to `<DEF>BOD^E{|Statement: ~crate::mark(<@Value@>);~|}<DEF>`.\n" +
		"Create \"Forward|-s|-ed|-ing|-er\" for Processing a New Process for Marking any Thing.\n" +
		"Create \"Prepare|-s|-d|--ing|--er\" for Forwarding a New Next.\n" +
		"Create \"Exercise\" for Processing a Generic `exercise` Function of a New Process of a New Preparer with No Inputs and No Outputs.\n"
	c := compileDefinitionSource(t, source, "source-sequence-prerequisite.l")
	instance, err := c.ComposeDefinition("Exercise")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(instance.Expression, "Key: ~crate::next_value~") != 1 || strings.Count(instance.Expression, "crate::mark(") != 1 || strings.Index(instance.Expression, "Key: ~crate::next_value~") > strings.Index(instance.Expression, "crate::mark(") {
		t.Fatalf("flattened source Process lost, repeated, or reordered its prerequisite: %s", instance.Expression)
	}
}

func TestWordLibrarySourceWrapperOperandsOnce(t *testing.T) {
	const leaf = "Define \"Transmit|-s|-ted|-ting|-ter\" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~crate::transmit~, Inputs: ~&[<@Context@>, <@Value@>]&~|}<DEF>`.\n"
	const composed = "Define \"Send|-s|-ed|-ing|-er\" to `<DEF>KY{|Key: ~crate::transmit~, Inputs: ~&[<@Context@>, <@Value@>]&~|}<DEF>`.\nCreate \"Transmit|-s|-ted|-ting|-ter\" for Sending any Thing and any Text.\n"
	for _, test := range []struct{ name, base, modifiers, wrapperInput, events string }{
		{"two-wrappers", leaf, "Tracked Noted", "any Thing", "1,2,3,4,5"},
		{"repeated-wrapper", leaf, "Tracked Tracked", "any Thing", "1,2,3,5,5"},
		{"source-target", composed, "Tracked Noted", "any Thing", "1,2,3,4,5"},
		{"wrapper-owned-operand", leaf, "Tracked", "a New Witness", "1,2,6,3,5"},
	} {
		t.Run(test.name, func(t *testing.T) {
			source := "Define \"Acquire\" as `<DEF>KY{|Key: ~crate::acquire~|}<DEF>`.\n" +
				"Define \"Witness\" as `<DEF>KY{|Key: ~crate::witness~|}<DEF>`.\n" +
				"Define \"Payload\" as `<DEF>KY{|Key: ~crate::payload~|}<DEF>`.\n" + test.base +
				"Define \"Complete|-s|-d|--ing|-r\" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~crate::complete~, Inputs: ~&[<@Context@>, <@Receipt@>]&~|}<DEF>`.\n" +
				"Define \"Finish|-es|-ed|-ing|-er\" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~crate::finish~, Inputs: ~&[<@Context@>, <@Receipt@>]&~|}<DEF>`.\n" +
				"Define \"Tracked\" as a New Adjective describing any Transmitter.\nIt describes any Transmitter as a New Completer of " + test.wrapperInput + " and any Transmitter.\n" +
				"Define \"Noted\" as a New Adjective describing any Transmitter.\nIt describes any Transmitter as a New Finisher of any Thing and any Transmitter.\n\n" +
				"    Publicly create \"Exercise\" as a Generic Function with No Inputs and Outputs of `String`.\n" +
				"Create \"Receipt\" as a New " + test.modifiers + " Transmitter of a New Acquire and a New Payload.\nReturn Receipt.\n"
			c := compileDefinitionSource(t, source, "wrapper-operands-once.l")
			testNativeDefinitionLibrary(t, c, "Exercise", "wrapper_operands", `
static EVENTS: std::sync::Mutex<Vec<u8>> = std::sync::Mutex::new(Vec::new());
pub fn acquire() -> u64 { EVENTS.lock().unwrap().push(1); 7 }
pub fn witness() -> u64 { EVENTS.lock().unwrap().push(6); 7 }
pub fn payload() -> String { EVENTS.lock().unwrap().push(2); String::from("payload") }
pub fn transmit(context: u64, value: String) -> String { assert_eq!(context, 7); EVENTS.lock().unwrap().push(3); value }
pub fn finish(context: u64, value: String) -> String { assert_eq!(context, 7); EVENTS.lock().unwrap().push(4); value }
pub fn complete(context: u64, value: String) -> String { assert_eq!(context, 7); EVENTS.lock().unwrap().push(5); value }
#[path="src/wrapper_operands.rs"] mod generated;
fn main() { for _ in 0..2 { EVENTS.lock().unwrap().clear(); assert_eq!(generated::exercise(), "payload"); assert_eq!(&*EVENTS.lock().unwrap(), &[`+test.events+`]); } }
`)
		})
	}
}

func TestWordLibrarySourceWrapperOwnership(t *testing.T) {
	for _, test := range []struct {
		name, typ, functions, observer string
		reject                         bool
	}{
		{"shared", "&String", `pub fn transmit(context: &String, value: u64) -> u64 { assert_eq!(context,"context"); value } pub fn complete(context: &String, value: u64) -> u64 { assert_eq!(context,"context"); value }`, `let context=String::from("context"); assert_eq!(generated::exercise(&context,7),7); assert_eq!(context,"context");`, false},
		{"exclusive", "&mut String", `pub fn transmit(context: &mut String, value: u64) -> u64 { context.push('b'); value } pub fn complete(context: &mut String, value: u64) -> u64 { context.push('w'); value }`, `let mut context=String::new(); assert_eq!(generated::exercise(&mut context,7),7); assert_eq!(context,"bw");`, false},
		{"consumed", "String", `pub fn transmit(context: String, value: u64) -> u64 { drop(context); value } pub fn complete(context: String, value: u64) -> u64 { drop(context); value }`, ``, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			source := "Define \"Transmit|-s|-ted|-ting|-ter\" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~crate::transmit~, Inputs: ~&[<@Context@>, <@Value@>]&~|}<DEF>`.\n" +
				"Define \"Complete|-s|-d|--ing|-r\" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~crate::complete~, Inputs: ~&[<@Context@>, <@Receipt@>]&~|}<DEF>`.\n" +
				"Define \"Tracked\" as a New Adjective describing any Transmitter.\nIt describes any Transmitter as a New Completer of any Thing and any Transmitter.\n\n" +
				"    Publicly create \"Exercise\" as a Generic Function with Inputs of \"context\" as `" + test.typ + "` and \"value\" as `u64` and Outputs of `u64`.\n" +
				"Create \"Receipt\" as a New Tracked Transmitter of Context and Value.\nReturn Receipt.\n"
			c := compileDefinitionSource(t, source, "wrapper-ownership.l")
			observer := test.functions + ` #[path="src/wrapper_ownership.rs"] mod generated; fn main() { ` + test.observer + ` }`
			if !test.reject {
				testNativeDefinitionLibrary(t, c, "Exercise", "wrapper_ownership", observer)
				return
			}
			aml, err := c.ComposeLibrary("Exercise", "wrapper_ownership")
			if err != nil {
				t.Fatal(err)
			}
			bundle, err := Logiceme.NewLogicemeService(c.Gram.resources).Evaluate(aml)
			if err != nil {
				t.Fatal(err)
			}
			files := make(map[string][]byte)
			for index, name := range bundle.MapArrays[0] {
				files[strings.TrimPrefix(name, "/")] = bundle.ByteArrays[0][index]
			}
			files["observe.rs"] = []byte(observer)
			directory := nativetest.Materialize(t, files)
			tools := nativetest.Toolchain(t)
			_, diagnostic, err := nativetest.BuildPrerequisite.Run(t, tools.Rustc, directory, nativetest.CleanEnvironment(tools.Rustc), "--edition=2024", "observe.rs", "-o", filepath.Join(directory, "observe"))
			if err == nil || !strings.Contains(string(diagnostic), "use of moved value") {
				t.Fatalf("wrapper must not clone a consumed operand: %v: %s", err, diagnostic)
			}
		})
	}
}

func TestWordLibraryDescribedConstructorBoundary(t *testing.T) {
	source := "Define \"Choose|-s|-d|--ing|-r\" to `<DEF>BOG^GI{|Key: ~<@Condition@>~|}^(E{|Statement: ~true~|}*<@Left@>*E{|Statement: ~false~|}*<@Right@>)<DEF>`.\n" +
		"Define \"Transmit|-s|-ted|-ting|-ter\" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~std::convert::identity~, Inputs: ~&[<@Value@>]&~|}<DEF>`.\n" +
		"Define \"Tracked\" as a New Adjective describing any Transmitter.\n" +
		"It describes any Transmitter as `<DEF>(<@Transmitter@>*BOD^E{|Statement: ~let _ = &<@Name|Label@>;~|})<DEF>`.\n\n" +
		"    Publicly create \"Exercise\" as a Generic Function with Inputs of \"condition\" as `bool` and \"value\" as `f64` and Outputs of `f64`.\n" +
		"Choose Condition, a New Process for Returning a New Tracked Transmitter of Value, and a New Process for Returning 7.\n"
	c := compileDefinitionSource(t, source, "described-constructor-boundary.l")
	testNativeDefinitionLibrary(t, c, "Exercise", "described_boundary", `#[path="src/described_boundary.rs"] mod generated; fn main() { assert_eq!(generated::exercise(true, 3.0), 3.0); assert_eq!(generated::exercise(false, 3.0), 7.0); }`)
}

func TestWordLibraryDescriptorAritySelection(t *testing.T) {
	const choice = "Define \"Choose|-s|-d|--ing|-r\" to `<DEF>BOG^GI{|Key: ~<@Condition@>~|}^(E{|Statement: ~true~|}*<@Left@>*E{|Statement: ~false~|}*<@Right@>)<DEF>`.\n"
	const leaf = "Define \"Transmit|-s|-ted|-ting|-ter\" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~std::convert::identity~, Inputs: ~&[<@Value@>]&~|}<DEF>`.\n"
	const descriptor = "Define \"Tracked\" as a New Adjective describing any Transmitter.\n"
	const wrapper = "It describes any Transmitter as `<DEF>(<@Transmitter@>*BOD^E{|Statement: ~let _ = &<@Name|Label@>;~|})<DEF>`.\n"
	for _, test := range []struct{ name, words, local, expression, after string }{
		{"function-alias", "\n    Publicly create \"Transmit|-s|-ted|-ting|-ter\" as a Generic Function with Inputs of \"value\" as `f64` and Outputs of `f64`.\nReturn Value.\n\n" + descriptor + wrapper + "Use Transmit as \"Routed\".\n", "", "a New Tracked Routed of Value", ""},
		{"source-wrapper", leaf + "Define \"Finish|-es|-ed|-ing|-er\" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~crate::finish~, Inputs: ~&[<@Original@>, <@Receipt@>]&~|}<DEF>`.\n" + descriptor + "It describes any Transmitter as a New Finisher of any Thing and any Transmitter.\n", "", "a New Tracked Transmitter of Value", ""},
		{"replacement", leaf + descriptor + "It describes any Transmitter as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~crate::finish~, Inputs: ~&[<@First@>, <@Second@>]&~|}<DEF>`.\n", "", "a New Tracked Transmitter of 1 and Value", ""},
		{"forward-refinement", leaf, descriptor, "a New Tracked Transmitter of Value", wrapper},
		{"later-replacement", leaf, descriptor + wrapper, "a New Tracked Transmitter of 1 and Value", "It describes any Transmitter as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~crate::finish~, Inputs: ~&[<@First@>, <@Second@>]&~|}<DEF>`.\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			source := choice + test.words + "\n    Publicly create \"Exercise\" as a Generic Function with Inputs of \"condition\" as `bool` and \"value\" as `f64` and Outputs of `f64`.\n" +
				test.local + "Choose Condition, a New Process for Returning " + test.expression + ", and a New Process for Returning 7.\n" + test.after
			c := compileDefinitionSource(t, source, "descriptor-arity-selection.l")
			testNativeDefinitionLibrary(t, c, "Exercise", "descriptor_arity", `pub fn finish(_: f64, value: f64) -> f64 { value } #[path="src/descriptor_arity.rs"] mod generated; fn main() { assert_eq!(generated::exercise(true, 3.0), 3.0); assert_eq!(generated::exercise(false, 3.0), 7.0); }`)
		})
	}
}

func wordLibraryGeneratedSource(t *testing.T, compilation *Compilation, name string) string {
	t.Helper()
	aml, err := compilation.ComposeLibrary(name, "word_library")
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := Logiceme.NewLogicemeService(compilation.Gram.resources).Evaluate(aml)
	if err != nil {
		t.Fatalf("%v\nAML: %s", err, aml)
	}
	var source strings.Builder
	for index, path := range bundle.MapArrays[0] {
		if strings.HasSuffix(path, ".rs") {
			source.Write(bundle.ByteArrays[0][index])
		}
	}
	return source.String()
}

func TestWordLibrarySourceWrapperIdentity(t *testing.T) {
	const leaf = "Define \"Transmit|-s|-ted|-ting|-ter\" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~std::convert::identity~, Inputs: ~&[<@Value@>]&~|}<DEF>`.\n"
	for _, test := range []struct {
		name, local, target string
		reject              bool
	}{
		{"alias", "", "Routed", false},
		{"same-spelling-foreign-target", leaf, "Transmitter", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			source := leaf + "Use Transmit as \"Routed\".\n" +
				"Define \"Finish|-es|-ed|-ing|-er\" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~crate::finish~, Inputs: ~&[<@Original@>, <@Receipt@>]&~|}<DEF>`.\n" +
				"Define \"Tracked\" as a New Adjective describing any Transmitter.\nIt describes any Transmitter as a New Finisher of any Thing and any Routed.\n\n" +
				"    Publicly create \"Exercise\" as a Generic Function with Inputs of \"value\" as `f64` and Outputs of `f64`.\n" + test.local +
				"Return a New Tracked " + test.target + " of Value.\n"
			c := NewLygerService(canonicalTestOptions()).NewCompilation([]byte(source), "source-wrapper-identity.l")
			c.Make()
			if test.reject {
				if c.Error == nil || !strings.Contains(c.Error.Error(), "has no constructor") {
					t.Fatalf("source descriptor accepted a foreign same-spelled target: %v", c.Error)
				}
				return
			}
			if c.Error != nil {
				t.Fatal(c.Error)
			}
			testNativeDefinitionLibrary(t, c, "Exercise", "wrapper_identity", `pub fn finish(original:f64,value:f64)->f64 { assert_eq!(original,value); value } #[path="src/wrapper_identity.rs"] mod generated; fn main(){assert_eq!(generated::exercise(7.0),7.0);}`)
		})
	}
}

func TestWordLibraryConstructorGrammarCompilationIsolation(t *testing.T) {
	service := NewLygerService(canonicalTestOptions())
	for _, test := range []struct{ name, arguments, inputs string }{
		{"one", "Value", "<@Value@>"},
		{"two", "Value and Value", "<@First@>, <@Second@>"},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			source := "Define \"Choose|-s|-d|--ing|-r\" to `<DEF>BOG^GI{|Key: ~<@Condition@>~|}^(E{|Statement: ~true~|}*<@Left@>*E{|Statement: ~false~|}*<@Right@>)<DEF>`.\n" +
				"Define \"Transmit|-s|-ted|-ting|-ter\" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~base~, Inputs: ~&[<@Value@>]&~|}<DEF>`.\n" +
				"Define \"Tracked\" as a New Adjective describing any Transmitter.\nIt describes any Transmitter as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~selected~, Inputs: ~&[" + test.inputs + "]&~|}<DEF>`.\n\n" +
				"    Publicly create \"Exercise\" as a Generic Function with Inputs of \"condition\" as `bool` and \"value\" as `f64` and Outputs of `f64`.\n" +
				"Choose Condition, a New Process for Returning a New Tracked Transmitter of " + test.arguments + ", and a New Process for Returning 7.\n"
			c := service.NewCompilation([]byte(source), "isolated-"+test.name+".l")
			c.Make()
			if c.Error != nil {
				t.Fatal(c.Error)
			}
			first, err := c.ComposeDefinition("Exercise")
			if err != nil {
				t.Fatal(err)
			}
			second, err := c.ComposeDefinition("Exercise")
			if err != nil {
				t.Fatal(err)
			}
			if first.Expression != second.Expression || !strings.Contains(first.Expression, "Key: ~selected~") {
				t.Fatalf("constructor selection mutated on repeated composition: %v", err)
			}
		})
	}
}

func TestWordLibraryConstructorGrammarDiscardsFailedAlternative(t *testing.T) {
	gram, err := loadGrammar(canonicalTestOptions())
	if err != nil {
		t.Fatal(err)
	}
	gram.grammarMap["statement.objective.imperative"] = grammeme{Structure: "CASE", Cases: map[string]string{"declarative": "SUBJECT"}}
	gram.grammarMap["subject"] = grammeme{Structure: "CASE", Cases: map[string]string{
		"a_rejected": "ID | PHRASE.GENERIC",
		"z_selected": "ID | PHRASE.GENERIC",
	}}
	tokens := grammarTestTokens(t, "Marker of 3", gram)
	tokens[0].tType, tokens[0].defined = "nouns.generic.root", true
	attempts := 0
	gram.constructorSignature = func(*token, []*token, bool) (grammarArity, error) {
		attempts++
		if attempts == 1 {
			return grammarArity{}, lerrors.UnresolvedReference.Errorf("discarded constructor alternative")
		}
		return grammarArity{Count: 1, Fixed: true}, nil
	}
	statement := &lStatement{gram: gram, sType: "verbs.declarative.root", tokens: tokens}
	if err := statement.loadGram(); err != nil {
		t.Fatal(err)
	}
	if attempts < 2 || statement.sGram.subject.cType != "z_selected" {
		t.Fatalf("valid constructor alternative was not selected: %d attempts", attempts)
	}
}
