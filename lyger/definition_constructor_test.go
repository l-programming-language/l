package Lyger

import (
	"errors"
	"path/filepath"

	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	"github.com/l-programming-language/l/lerrors"
	"strings"
	"testing"

	Logiceme "github.com/l-programming-language/l/logiceme"
)

const parametricConstructorSource = "Define \"Parameterised|-s|-like|---ing\" as `<DEF>FW{|Func{Key: <@Name|Text@>, Inputs: <@Inputs|Props@>, Outputs: <@Outputs|Texts@>}Func, GenericTypes: ~<@Types|Type@>~|}^(<@Process@>)<DEF>`.\n" +
	"Use Function's Return as Parameterised's \"Return|-s|-ed|-ing|-er\".\n"

func TestLeadingCompositionGroup(t *testing.T) {
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

func TestConstructorLabelArity(t *testing.T) {
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

func TestExplicitConstructorParameters(t *testing.T) {
	source := parametricConstructorSource + "\n" +
		"    Publicly create \"Carry|--ies|--ied|-ing|--ier\" as a Generic Parameterised of `T` with Inputs of \"value\" as `T` and Outputs of `T`.\nReturn Value.\n"
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
	generated := generatedDefinitionSource(t, c, "Carry")
	if !strings.Contains(generated, "fn carry<T>(") || !strings.Contains(generated, "return value;") {
		t.Fatalf("generic declaration was not composed: %s", generated)
	}
}

func TestActorDescriptorBase(t *testing.T) {
	source := "Define \"Transmit|-s|-ted|-ting|-ter\" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~std::convert::identity~, Inputs: ~&[<@Value@>]&~|}<DEF>`.\n" +
		"Define \"Tracked\" as a New Adjective describing any Transmitter.\n" +
		"It describes any Transmitter as `<DEF>(BOD^E{|Statement: ~/* outer before */~|}*<@Transmitter@>*BOD^E{|Statement: ~let _ = &<@Name|Label@>; /* outer after */~|})<DEF>`.\n" +
		"Define \"Observed\" as a New Adjective describing any Transmitter.\n" +
		"It describes any Transmitter as `<DEF>(BOD^E{|Statement: ~let _ = &<@Value@>; /* inner before */~|}*<@Transmitter@>*BOD^E{|Statement: ~/* inner after */~|})<DEF>`.\n\n" +
		"    Publicly create \"Exercise|-s|-d|--ing|-r\" as a Generic Function with Inputs of \"value\" as `f64` and Outputs of `f64`.\n" +
		"Create \"Receipt\" as a New Tracked Observed Transmitter of Value.\nReturn Receipt.\n"
	c := compileDefinitionSource(t, source, "actor-descriptor.l")
	generated := generatedDefinitionSource(t, c, "Exercise")
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

func TestSourceWrapperOwnership(t *testing.T) {
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
				"    Publicly create \"Exercise|-s|-d|--ing|-r\" as a Generic Function with Inputs of \"context\" as `" + test.typ + "` and \"value\" as `u64` and Outputs of `u64`.\n" +
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

func generatedDefinitionSource(t *testing.T, compilation *Compilation, name string) string {
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

func TestSourceWrapperIdentity(t *testing.T) {
	const leaf = "Define \"Transmit|-s|-ted|-ting|-ter\" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~std::convert::identity~, Inputs: ~&[<@Value@>]&~|}<DEF>`.\n"
	for _, test := range []struct {
		name, local, target string
		reject              bool
	}{
		{"alias", "", "Routed", false},
		{"same-spelling-foreign-target", leaf, "Transmitter", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			source := leaf + "Use Transmit as \"Routed|-s|-ed|-ing|-er\".\n" +
				"Define \"Finish|-es|-ed|-ing|-er\" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~crate::finish~, Inputs: ~&[<@Original@>, <@Receipt@>]&~|}<DEF>`.\n" +
				"Define \"Tracked\" as a New Adjective describing any Transmitter.\nIt describes any Transmitter as a New Finisher of any Thing and any Routed.\n\n" +
				"    Publicly create \"Exercise|-s|-d|--ing|-r\" as a Generic Function with Inputs of \"value\" as `f64` and Outputs of `f64`.\n" + test.local +
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

func TestConstructorGrammarDiscardsFailedAlternative(t *testing.T) {
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

func TestSourceFunctionReplacementRejects(t *testing.T) {
	source := "Define \"Changed\" as a New Adjective describing any Function.\nIt describes any Function as a New Function of any Process.\nDefine \"Wrapped|-s|-like|----ing\" as a New Changed Function.\n\n" +
		"    Publicly create \"Exercise|-s|-d|--ing|-r\" as a Generic Changed Function with No Inputs and Outputs of `f64`.\nReturn 7.\n"
	c := compileDefinitionSource(t, source, "source-function.l")
	if template, err := c.callableTemplate(c.Memory["Wrapped"], make(map[*record]bool)); err != nil || template != nil {
		t.Fatalf("source replacement cannot supply a native Function signature: %v", err)
	}
	if _, err := c.ComposeLibrary("Exercise", "source_function"); !errors.Is(err, lerrors.TypeMismatch) {
		t.Fatalf("source replacement of a Function signature must reject: %v", err)
	}
}

func TestReplacementResultSignature(t *testing.T) {
	const named = "<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~std::convert::identity~, Inputs: ~&[<@Value@>]&~|}<DEF>"
	const expression = "<DEF>KY{|Key: ~std::convert::identity~, Inputs: ~&[<@Value@>]&~|}<DEF>"
	for _, tc := range []struct{ name, base, replacement string }{{"removes-name", named, expression}, {"adds-name", expression, named}} {
		t.Run(tc.name, func(t *testing.T) {
			for _, mode := range []string{"direct-control", "replaced"} {
				t.Run(mode, func(t *testing.T) {
					definition, modifiers := tc.replacement, ""
					if mode == "replaced" {
						definition, modifiers = tc.base, "Changed "
					}
					source := "Define \"Transmit|-s|-ted|-ting|-ter\" to `" + definition + "`.\n" +
						"Define \"Changed\" as a New Adjective describing any Transmitter.\nIt describes any Transmitter as `" + tc.replacement + "`.\n\n" +
						"    Publicly create \"Exercise|-s|-d|--ing|-r\" as a Generic Function with Inputs of \"value\" as `f64` and Outputs of `f64`.\nReturn a New " + modifiers + "Transmitter of Value.\n"
					c := compileDefinitionSource(t, source, "word-library-result-signature.l")
					testNativeDefinitionLibrary(t, c, "Exercise", "result_signature", `#[path="src/result_signature.rs"]mod generated;fn main(){assert_eq!(generated::exercise(7.0),7.0);}`)
				})
			}
		})
	}
}

func TestReplacementNamedResult(t *testing.T) {
	const named = "<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~std::convert::identity~, Inputs: ~&[<@Value@>]&~|}<DEF>"
	const expression = "<DEF>KY{|Key: ~std::convert::identity~, Inputs: ~&[<@Value@>]&~|}<DEF>"
	for _, pair := range [][2]string{{named, expression}, {expression, named}} {
		source := "Define \"Transmit|-s|-ted|-ting|-ter\" to `" + pair[0] + "`.\nDefine \"Changed\" as a New Adjective describing any Transmitter.\nIt describes any Transmitter as `" + pair[1] + "`.\n\n    Publicly create \"Exercise|-s|-d|--ing|-r\" as a Generic Function with Inputs of \"value\" as `f64` and Outputs of `f64`.\nCreate \"Receipt\" as a New Changed Transmitter of Value.\nReturn Receipt.\n"
		c := compileDefinitionSource(t, source, "replacement-named-result.l")
		testNativeDefinitionLibrary(t, c, "Exercise", "named_result", `#[path="src/named_result.rs"]mod generated;fn main(){assert_eq!(generated::exercise(7.0),7.0);}`)
	}
}

func TestInheritedSignatures(t *testing.T) {
	for _, test := range []struct {
		name, leaf, declaration, expression, expected string
		count                                         int
		fixed                                         bool
	}{
		{"zero", "<DEF>KY{|Key: ~crate::constant~|}<DEF>", "Create \"Transmit|-s|-ted|-ting|-ter\" for Forwarding a New Pass.", "a New Transmitter", "7.0", 0, true},
		{"variadic", "<DEF>KY{|Key: ~crate::sum~, Inputs: ~&[<@Values|Atoms@>]&~|}<DEF>", "Create \"Transmit|-s|-ted|-ting|-ter\" for Forwarding a New Pass.", "a New Transmitter of 2 and 3", "5.0", 1, false},
		{"source-value-alias", "<DEF>KY{|Key: ~std::convert::identity~, Inputs: ~&[<@Value@>]&~|}<DEF>", "Define \"Transmitter|-s|-like|---ing\" as a New Pass.\nUse Transmitter as \"Routed|-s|-like|--ing\".", "a New Routed of 3", "3.0", 1, true},
		{"callable-member", "<DEF>KY{|Key: ~std::convert::identity~, Inputs: ~&[<@Value@>]&~|}<DEF>", "Create Function's \"Transmit|-s|-ted|-ting|-ter\" for Forwarding a New Pass.", "a New Function's Transmitter of 3", "3.0", 1, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			source := "Define \"Forward|-s|-ed|-ing|-er\" to `<DEF><@Expression@><DEF>`.\nDefine \"Pass|-es|-like|-ing\" as `" + test.leaf + "`.\n" + test.declaration + "\n\n    Publicly create \"Exercise|-s|-d|--ing|-r\" as a Generic Function with No Inputs and Outputs of `f64`.\nReturn " + test.expression + ".\n"
			c := compileDefinitionSource(t, source, "inherited-signature.l")
			name := "Transmit"
			if test.name == "callable-member" {
				name = "Function.Transmit"
			}
			if test.name == "source-value-alias" {
				name = "Transmitter"
			}
			var selected *record
			for rec := range c.Semantics.Declarations {
				if strings.EqualFold(rec.ID, name) {
					selected = rec
				}
			}
			if selected == nil {
				t.Fatalf("no declaration %s", name)
			}
			header, err := c.sourceHeader(selected, false, make(map[*record]bool))
			if err != nil || header.arity() != (grammarArity{Count: test.count, Fixed: test.fixed}) {
				t.Fatalf("selected signature: %+v, %v", header.arity(), err)
			}
			testNativeDefinitionLibrary(t, c, "Exercise", "inherited_signature", `pub fn constant()->f64{7.0}pub fn sum(x:f64,y:f64)->f64{x+y}#[path="src/inherited_signature.rs"]mod generated;fn main(){assert_eq!(generated::exercise(),`+test.expected+`);}`)
		})
	}
}

func TestReferenceOnlyActionRejects(t *testing.T) {
	source := "Define \"Forward|-s|-ed|-ing|-er\" to `<DEF><@Value@><DEF>`.\n\n    Publicly create \"Exercise|-s|-d|--ing|-r\" as a Generic Function with Inputs of \"value\" as `f64` and Outputs of `f64`.\nForward Value.\nReturn 7.\n"
	c := NewLygerService(canonicalTestOptions()).NewCompilation([]byte(source), "reference-only-action.l")
	c.Make()
	err := c.Error
	if err == nil {
		_, err = c.ComposeDefinition("Exercise")
	}
	if err == nil || !strings.Contains(err.Error(), "value reference is not an executable") {
		t.Fatalf("reference-only action: %v", err)
	}
}

func TestIdentityRetainsPrerequisites(t *testing.T) {
	template, err := ParseDefinitionTemplate([]byte("<DEF><@Body@><DEF>"), SourceSpan{Path: "identity.l"})
	if err != nil {
		t.Fatal(err)
	}
	prior := DefinitionInstance{Expression: "MOB{|Key: ~value~|}^KY{|Key: ~crate::acquire~|}", Span: SourceSpan{Path: "operand.l"}}
	body := DefinitionInstance{Expression: "BOD^E{|Statement: ~crate::consume(value);~|}", before: []DefinitionInstance{prior}, Span: SourceSpan{Path: "body.l"}}
	instance, err := template.Instantiate(map[string]DefinitionArgument{"Body": {Kind: "Definition", Definition: &body}})
	if err != nil {
		t.Fatal(err)
	}
	steps := definitionSteps(instance)
	if len(steps) != 2 || steps[0].Expression != prior.Expression || steps[0].Span != prior.Span || steps[1].Expression != body.Expression {
		t.Fatalf("identity lost or reordered an operand prerequisite: %+v", steps)
	}
}

// Retain existing characterization entry points while production uses the shared binder.
func (c *Compilation) instantiateNamedDefinition(rec *record, arguments map[string]DefinitionArgument, used map[string]bool, active map[*record]bool, values ...DefinitionArgument) (DefinitionInstance, error) {
	return c.instantiateSourceCall(rec, values, arguments, used, active, nil)
}

func definitionArguments(template *DefinitionTemplate, args []DefinitionArgument) (map[string]DefinitionArgument, error) {
	return bindDefinitionArguments(template, args, nil, nil)
}

func TestFunctionActorDescriptorBase(t *testing.T) {
	source := parametricConstructorSource + "\n" +
		"Define \"Polymorphic\" as a New Adjective describing any Parameterised.\nIt describes a Generic Parameterised of `T`.\n\n" +
		"    Publicly create \"Transmit|-s|-ted|-ting|-ter\" as a Generic Polymorphic Parameterised with Inputs of \"value\" as `T` and Outputs of `T`.\nReturn Value.\n\n" +
		"Define \"Tracked\" as a New Adjective describing any Transmitter.\n" +
		"It describes any Transmitter as `<DEF>(<@Transmitter@>*BOD^E{|Statement: ~let _ = &<@Name|Label@>;~|})<DEF>`.\n\n" +
		"    Publicly create \"Exercise|-s|-d|--ing|-r\" as a Generic Function with Inputs of \"value\" as `String` and Outputs of `String`.\n" +
		"Create \"Receipt\" as a New Tracked Transmitter of Value.\nReturn Receipt.\n"
	c := compileDefinitionSource(t, source, "function-actor-descriptor.l")
	generated := generatedDefinitionSource(t, c, "Exercise")
	if strings.Count(generated, "(transmit)(value)") != 1 || !strings.Contains(generated, "fn transmit<T>(") || !strings.Contains(generated, "let _ = &receipt;") || !strings.Contains(generated, "return receipt;") {
		t.Fatalf("Function descriptor lost its base invocation, generic declaration, or result: %s", generated)
	}
}

func TestSourceActorDescriptorBase(t *testing.T) {
	source := parametricConstructorSource + "\n" +
		"    Publicly create \"Transmit|-s|-ted|-ting|-ter\" as a Generic Parameterised of `T: ?Sized` with Inputs of \"context\" as `&mut T` and \"value\" as `u64` and Outputs of `u64`.\nReturn Value.\n\n" +
		"    Create \"Complete|-s|-d|--ing|-r\" as a Generic Parameterised of `T: ?Sized` with Inputs of \"context\" as `&mut T` and \"receipt\" as `u64` and Outputs of `u64`.\nReturn Receipt.\n" +
		"Publicly define \"Tracked\" as a New Adjective describing any Transmitter.\n" +
		"It describes any Transmitter as a New Completer of any Thing and any Transmitter.\n\n" +
		"    Publicly create \"Exercise|-s|-d|--ing|-r\" as a Generic Function with Inputs of \"context\" as `&mut u64` and \"value\" as `u64` and Outputs of `u64`.\n" +
		"Create \"Receipt\" as a New Tracked Transmitter of Context and Value.\nReturn Receipt.\n"
	c := compileDefinitionSource(t, source, "source-actor-descriptor.l")
	generated := generatedDefinitionSource(t, c, "Exercise")
	if strings.Count(generated, "(transmit)(context, value)") != 1 || !strings.Contains(generated, "fn transmit<T: ?Sized>(") || !strings.Contains(generated, "fn complete<T: ?Sized>(") || !strings.Contains(generated, "(complete)(context, ") || !strings.Contains(generated, "return receipt;") {
		t.Fatalf("source descriptor lost the retained base, checked inputs, helper dependency, or result: %s", generated)
	}
	if strings.Index(generated, "(transmit)(context, value)") > strings.Index(generated, "(complete)(context, ") {
		t.Fatalf("base must finish before the wrapper borrows the original context: %s", generated)
	}
}

func TestSourceActorDescriptorNative(t *testing.T) {
	source := parametricConstructorSource + "\n" +
		"Define \"Mark|-s|-ed|-ing|-er\" to `<DEF>BOD^E{|Statement: ~std::io::Write::write_all(<@Target@>, b\"base\").unwrap();~|}<DEF>`.\n" +
		"Define \"Append|-s|-ed|-ing|-er\" to `<DEF>BOD^E{|Statement: ~std::io::Write::write_all(<@Target@>, b\"wrapped\").unwrap();~|}<DEF>`.\n\n" +
		"    Publicly create \"Transmit|-s|-ted|-ting|-ter\" as a Generic Parameterised of `T: std::io::Write + ?Sized` with Inputs of \"context\" as `&mut T` and \"value\" as `u64` and Outputs of `u64`.\nMark Context.\nReturn Value.\n\n" +
		"    Create \"Complete|-s|-d|--ing|-r\" as a Generic Parameterised of `T: std::io::Write + ?Sized` with Inputs of \"context\" as `&mut T` and \"receipt\" as `u64` and Outputs of `u64`.\nAppend Context.\nReturn Receipt.\n" +
		"Publicly define \"Tracked\" as a New Adjective describing any Transmitter.\n" +
		"It describes any Transmitter as a New Completer of any Thing and any Transmitter.\n\n" +
		"    Publicly create \"Exercise|-s|-d|--ing|-r\" as a Generic Function with Inputs of \"context\" as `&mut dyn std::io::Write` and \"value\" as `u64` and Outputs of `u64`.\n" +
		"Create \"Receipt\" as a New Tracked Transmitter of Context and Value.\nReturn Receipt.\n"
	c := compileDefinitionSource(t, source, "source-actor-descriptor-native.l")
	testNativeDefinitionLibrary(t, c, "Exercise", "exercise", `#[path="src/exercise.rs"] mod generated; fn main() { let mut sink = Vec::<u8>::new(); assert_eq!(generated::exercise(&mut sink, 17), 17); assert_eq!(sink, b"basewrapped"); }`)
}

func TestNamedSourceWrapperEvaluatesForwardedInputOnce(t *testing.T) {
	source := parametricConstructorSource + "\n" +
		"Define \"Acquire|-s|---sitive|--ing\" as `<DEF>KY{|Key: ~acquire_context~|}<DEF>`.\n\n" +
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
	}, make(map[string]bool), make(map[*record]bool), nil)
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

func TestSourceSequenceRetainsPrerequisite(t *testing.T) {
	source := "Define \"Next|-s|-like|-ing\" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~crate::next_value~|}<DEF>`.\n" +
		"Define \"Mark|-s|-ed|-ing|-er\" to `<DEF>BOD^E{|Statement: ~crate::mark(<@Value@>);~|}<DEF>`.\n" +
		"Create \"Forward|-s|-ed|-ing|-er\" for Processing a New Process for Marking any Thing.\n" +
		"Create \"Prepare|-s|-d|--ing|--er\" for Forwarding a New Next.\n" +
		"Create \"Exercise|-s|-d|--ing|-r\" for Processing a Generic `exercise` Function of a New Process of a New Preparer with No Inputs and No Outputs.\n"
	c := compileDefinitionSource(t, source, "source-sequence-prerequisite.l")
	instance, err := c.ComposeDefinition("Exercise")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(instance.Expression, "Key: ~crate::next_value~") != 1 || strings.Count(instance.Expression, "crate::mark(") != 1 || strings.Index(instance.Expression, "Key: ~crate::next_value~") > strings.Index(instance.Expression, "crate::mark(") {
		t.Fatalf("flattened source Process lost, repeated, or reordered its prerequisite: %s", instance.Expression)
	}
	// The native observer uses the ordinary declaration role for its Function;
	// the nested Prepare -> Forward -> Process prerequisite is unchanged.
	nativeSource := strings.Replace(source, "Create \"Exercise|-s|-d|--ing|-r\" for Processing a Generic `exercise` Function of a New Process of a New Preparer with No Inputs and No Outputs.\n", "\n    Publicly create \"Exercise|-s|-d|--ing|-r\" as a Generic `exercise` Function with No Inputs and No Outputs.\nPrepare.\n", 1)
	native := compileDefinitionSource(t, nativeSource, "source-sequence-prerequisite-native.l")
	testNativeDefinitionLibrary(t, native, "Exercise", "source_prerequisite", `static EVENTS:std::sync::Mutex<Vec<u8>>=std::sync::Mutex::new(Vec::new());pub fn next_value()->u64{EVENTS.lock().unwrap().push(1);7}pub fn mark(value:u64){assert_eq!(value,7);EVENTS.lock().unwrap().push(2);}#[path="src/source_prerequisite.rs"]mod generated;fn main(){generated::exercise();assert_eq!(&*EVENTS.lock().unwrap(),&[1,2]);}`)
}

func TestSourceWrapperOperandsOnce(t *testing.T) {
	const leaf = "Define \"Transmit|-s|-ted|-ting|-ter\" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~crate::transmit~, Inputs: ~&[<@Context@>, <@Value@>]&~|}<DEF>`.\n"
	const composed = "Define \"Send|-s|-ed|-ing|-er\" to `<DEF>KY{|Key: ~crate::transmit~, Inputs: ~&[<@Context@>, <@Value@>]&~|}<DEF>`.\nCreate \"Transmit|-s|-ted|-ting|-ter\" for Sending any Thing and any Text.\n"
	for _, test := range []struct{ name, base, modifiers, wrapperInput, events string }{
		{"two-wrappers", leaf, "Tracked Noted", "any Thing", "1,2,3,4,5"},
		{"repeated-wrapper", leaf, "Tracked Tracked", "any Thing", "1,2,3,5,5"},
		{"source-target", composed, "Tracked Noted", "any Thing", "1,2,3,4,5"},
		{"wrapper-owned-operand", leaf, "Tracked", "a New Witness", "1,2,6,3,5"},
	} {
		t.Run(test.name, func(t *testing.T) {
			source := "Define \"Acquire|-s|---sitive|--ing\" as `<DEF>KY{|Key: ~crate::acquire~|}<DEF>`.\n" +
				"Define \"Witness|-es|-like|-ing\" as `<DEF>KY{|Key: ~crate::witness~|}<DEF>`.\n" +
				"Define \"Payload|-s|-like|-ing\" as `<DEF>KY{|Key: ~crate::payload~|}<DEF>`.\n" + test.base +
				"Define \"Complete|-s|-d|--ing|-r\" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~crate::complete~, Inputs: ~&[<@Context@>, <@Receipt@>]&~|}<DEF>`.\n" +
				"Define \"Finish|-es|-ed|-ing|-er\" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~crate::finish~, Inputs: ~&[<@Context@>, <@Receipt@>]&~|}<DEF>`.\n" +
				"Define \"Tracked\" as a New Adjective describing any Transmitter.\nIt describes any Transmitter as a New Completer of " + test.wrapperInput + " and any Transmitter.\n" +
				"Define \"Noted\" as a New Adjective describing any Transmitter.\nIt describes any Transmitter as a New Finisher of any Thing and any Transmitter.\n\n" +
				"    Publicly create \"Exercise|-s|-d|--ing|-r\" as a Generic Function with No Inputs and Outputs of `String`.\n" +
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

func TestTransparentValueWrapper(t *testing.T) {
	const words = "Define \"Acquire|-s|---sitive|--ing\" as `<DEF>KY{|Key: ~crate::acquire~|}<DEF>`.\n" +
		"Define \"Transmit|-s|-ted|-ting|-ter\" to `<DEF><@Value@><DEF>`.\n" +
		"Define \"Finish|-es|-ed|-ing|-er\" to `<DEF>KY{|Key: ~crate::finish~, Inputs: ~&[<@Original@>, <@Receipt@>]&~|}<DEF>`.\n" +
		"Define \"Tracked\" as a New Adjective describing any Transmitter.\nIt describes any Transmitter as a New Finisher of any Thing and any Transmitter.\n\n" +
		"    Publicly create \"Exercise|-s|-d|--ing|-r\" as a Generic Function with No Inputs and Outputs of `f64`.\n"
	for _, test := range []struct{ name, declaration, expression string }{
		{"control", "", "a New Transmitter"},
		{"wrapped", "", "a New Tracked Transmitter"},
		{"source-value", "Define \"Routed|-s|-like|--ing\" as a New Transmitter.\nDefine \"Observed\" as a New Adjective describing any Routed.\nIt describes any Routed as a New Finisher of any Thing and any Routed.\n", "a New Observed Routed"},
		{"source-action", "Create \"Route|-s|-d|--ing|-r\" for Transmitting a New Transmitter.\nDefine \"Observed\" as a New Adjective describing any Router.\nIt describes any Router as a New Finisher of any Thing and any Router.\n", "a New Observed Router"},
	} {
		t.Run(test.name, func(t *testing.T) {
			declarations, body, _ := strings.Cut(words, "    Publicly create")
			c := compileDefinitionSource(t, declarations+test.declaration+"\n    Publicly create"+body+"Return "+test.expression+" of a New Acquire.\n", "word-library-transparent-value.l")
			testNativeDefinitionLibrary(t, c, "Exercise", "transparent_value", `static CALLS:std::sync::atomic::AtomicUsize=std::sync::atomic::AtomicUsize::new(0);pub fn acquire()->f64{CALLS.fetch_add(1,std::sync::atomic::Ordering::SeqCst);7.0}pub fn finish(original:f64,receipt:f64)->f64{assert_eq!(original,receipt);receipt}#[path="src/transparent_value.rs"]mod generated;fn main(){assert_eq!(generated::exercise(),7.0);assert_eq!(CALLS.load(std::sync::atomic::Ordering::SeqCst),1);}`)
		})
	}
}

func TestNamedCallableKeyIsOperand(t *testing.T) {
	source := "Define \"Relay|-s|-ed|-ing|-er\" to `<DEF>KY{|Key: ~<@Target@>~, Inputs: ~&[<@Value@>]&~|}<DEF>`.\n\n    Publicly create \"Exercise|-s|-d|--ing|-r\" as a Generic Function with Inputs of \"target\" as `fn(f64)->f64` and \"value\" as `f64` and Outputs of `f64`.\nCreate \"Receipt\" as a New Relayer of Target and Value.\nReturn Receipt.\n"
	c := compileDefinitionSource(t, source, "callable-key.l")
	testNativeDefinitionLibrary(t, c, "Exercise", "callable_key", `#[path="src/callable_key.rs"]mod generated;fn main(){assert_eq!(generated::exercise(std::convert::identity,7.0),7.0);}`)
}

func TestTransparentOwnedValueIsNotCloned(t *testing.T) {
	source := "Define \"Acquire|-s|---sitive|--ing\" as `<DEF>KY{|Key: ~crate::acquire~|}<DEF>`.\nDefine \"Transmit|-s|-ted|-ting|-ter\" to `<DEF><@Value@><DEF>`.\nDefine \"Finish|-es|-ed|-ing|-er\" to `<DEF>KY{|Key: ~crate::finish~, Inputs: ~&[<@Original@>, <@Receipt@>]&~|}<DEF>`.\nDefine \"Tracked\" as a New Adjective describing any Transmitter.\nIt describes any Transmitter as a New Finisher of any Thing and any Transmitter.\n\n    Publicly create \"Exercise|-s|-d|--ing|-r\" as a Generic Function with No Inputs and Outputs of `String`.\nReturn a New Tracked Transmitter of a New Acquire.\n"
	c := compileDefinitionSource(t, source, "transparent-owned.l")
	aml, err := c.ComposeLibrary("Exercise", "transparent_owned")
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
	files["observe.rs"] = []byte(`pub fn acquire()->String{String::from("value")}pub fn finish(_:String,value:String)->String{value}#[path="src/transparent_owned.rs"]mod generated;fn main(){let _ = generated::exercise();}`)
	directory := nativetest.Materialize(t, files)
	tools := nativetest.Toolchain(t)
	_, diagnostic, err := nativetest.BuildPrerequisite.Run(t, tools.Rustc, directory, nativetest.CleanEnvironment(tools.Rustc), "--edition=2024", "observe.rs", "-o", filepath.Join(directory, "observe"))
	if err == nil || !strings.Contains(string(diagnostic), "use of moved value") {
		t.Fatalf("transparent operand was cloned or evaluated again: %v: %s", err, diagnostic)
	}
}

// Retain existing characterization entry points while production uses the shared binder.

func TestInheritedConstructorDirectOperand(t *testing.T) {
	const words = "Define \"Forward|-s|-ed|-ing|-er\" to `<DEF><@Expression@><DEF>`.\n" +
		"Define \"Pass|-es|-like|-ing\" as `<DEF>KY{|Key: ~std::convert::identity~, Inputs: ~&[<@Value@>]&~|}<DEF>`.\n" +
		"Create \"Transmit|-s|-ted|-ting|-ter\" for Forwarding a New Pass.\n\n" +
		"    Publicly create \"Exercise|-s|-d|--ing|-r\" as a Generic Function with Inputs of \"condition\" as `bool` and \"value\" as `f64` and Outputs of `f64`.\n"
	for _, test := range []struct{ name, body, observer string }{
		{"direct", "Return a New Transmitter of Value.\n", `assert_eq!(generated::exercise(true,3.0),3.0);`},
		{"branch", "If Condition, Return a New Transmitter of Value, else Return 7.\n", `assert_eq!(generated::exercise(true,3.0),3.0);assert_eq!(generated::exercise(false,3.0),7.0);`},
	} {
		t.Run(test.name, func(t *testing.T) {
			c := compileDefinitionSource(t, words+test.body, "inherited-operand.l")
			testNativeDefinitionLibrary(t, c, "Exercise", "inherited_operand", `#[path="src/inherited_operand.rs"]mod generated;fn main(){`+test.observer+`}`)
		})
	}
}
