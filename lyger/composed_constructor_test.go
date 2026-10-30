package Lyger

import (
	"strings"
	"testing"

	Logiceme "github.com/l-programming-language/l/logiceme"
)

func TestSourceCompositionBindsEachConstructor(t *testing.T) {
	source := "Define \"Emit|-s|-ted|-ting|-ter\" to `<DEF>BOD^E{|Statement: ~println!(\"{}\", <@Value|Text@>);~|}<DEF>`.\n" +
		"Create \"Report|-s|-ed|-ing|-er\" for Processing a New Emitter of `first` and then Processing a New Emitter of `second`.\n" +
		"Define \"Worker\" as a New Basic App.\nThe App Reports.\n"
	c := compileDefinitionSource(t, source, "composition.l")
	body, err := c.lowerAppBody(c.Semantics.Apps[0])
	if err != nil {
		t.Fatal(err)
	}
	if len(body) != 1 || body[0].Expression != `W^BOD^E{|Statement: ~println!("{}", "first");~|}*W^BOD^E{|Statement: ~println!("{}", "second");~|}` {
		t.Fatalf("constructor operands lost: %+v", body)
	}
	var operands []DefinitionBinding
	for _, binding := range body[0].Bindings {
		if binding.Slot.Name == "Value" {
			operands = append(operands, binding)
		}
	}
	if len(operands) != 2 {
		t.Fatalf("constructor binding ledger = %+v", body[0].Bindings)
	}
	for i, expected := range []string{`"first"`, `"second"`} {
		binding := operands[i]
		if binding.Slot.Name != "Value" || binding.Emitted != expected || len(binding.Sources) != 1 || binding.Sources[0].Path != "composition.l" || binding.Sources[0].Start.Line != 2 {
			t.Fatalf("constructor %d provenance = %+v", i, binding)
		}
	}

	supplied := NewLygerService(canonicalTestOptions()).NewCompilation([]byte(strings.Replace(source, "The App Reports.", "The App Reports `extra`.", 1)), "composition.l")
	supplied.Make()
	if supplied.Error == nil || !strings.Contains(supplied.Error.Error(), "unused arguments") {
		t.Fatalf("unused argument accepted: %v", supplied.Error)
	}
}

func TestSourceCompositionConstructsNativeFunction(t *testing.T) {
	source := "Define \"Echo|-es|-ed|-ing|-er\" to `<DEF>BOD^E{|Statement: ~return <@Value@>;~|}<DEF>`.\n" +
		"Create \"Produce|-s|-d|--ing|--er\" for Processing a Generic `identity` Function of a New `value` Echoer with Inputs of \"value\" as `f64` and Outputs of `f64`.\n"
	c := compileDefinitionSource(t, source, "signature.l")
	instance, err := c.ComposeDefinition("Produce")
	if err != nil {
		t.Fatal(err)
	}
	want := `W^FW{|Key: ~identity~, Inputs: ~&[Prop{Key: value, Type: f64}Prop]&~, Outputs: ~&[(f64)]&~|}^BOD^E{|Statement: ~return value;~|}`
	if instance.Expression != want {
		t.Fatalf("typed source construction = %s", instance.Expression)
	}
}

func TestStaticFunctionOwnsInvocation(t *testing.T) {
	source := "    Publicly create \"Echo|-es|-ed|-ing|-er\" as a Generic Function with Inputs of \"value\" as `f64` and Outputs of `f64`.\nReturn Value.\n\n" +
		"Create \"Supply|-s|-ed|-ing|-er\" for Processing a New Echoer of 2.\n"
	c := compileDefinitionSource(t, source, "static.l")
	declaration, err := c.ComposeDefinition("Echo")
	if err != nil {
		t.Fatal(err)
	}
	if declaration.definition == nil || declaration.Symbol != "Echo" || declaration.Span.Path != "static.l" {
		t.Fatalf("declaration provenance: %+v", declaration)
	}
	call, err := c.ComposeDefinition("Supply")
	if err != nil {
		t.Fatal(err)
	}
	closure, err := definitionClosure(call)
	if err != nil {
		t.Fatal(err)
	}
	if len(closure) != 2 || closure[0].definition != declaration.definition || closure[0].Expression != declaration.Expression {
		t.Fatalf("invocation dependencies = %+v", closure)
	}
}

func TestStaticFunctionBindsLocalConstructor(t *testing.T) {
	source := "Define \"Subtract|-s|-ed|-ing|-or\" to `<DEF>DOB{|Key: ~<@Name@>~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: -, Operands: &[<@Left@>, <@Right@>]&}Op]&~|}<DEF>`.\n\n" +
		"    Publicly create \"Reduce|-s|-d|--ing|--er\" as a Generic Function with Inputs of \"left\" as `f64` and \"right\" as `f64` and Outputs of `f64`.\nCreate \"Final\" as a New Subtractor of Left and Right.\nReturn Final.\n"
	c := compileDefinitionSource(t, source, "locals.l")
	definition, err := c.ComposeDefinition("Reduce")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(definition.Expression, "Operands: &[left, right]&") || !strings.Contains(definition.Expression, "return __l_name_66696e616c;") {
		t.Fatalf("local constructor lost references: %s", definition.Expression)
	}
}

func TestScopedStaticFunctionBindsLocalConstructor(t *testing.T) {
	source := "Define \"Storage\" as a New Dictionary.\nDefine \"Subtract|-s|-ed|-ing|-or\" to `<DEF>DOB{|Key: ~<@Name@>~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: -, Operands: &[<@Left@>, <@Right@>]&}Op]&~|}<DEF>`.\n\n" +
		"    Publicly create Storage's \"Reduce|-s|-d|--ing|--er\" as a Generic `native_reduce` Function with Inputs of \"left\" as `f64` and \"right\" as `f64` and Outputs of `f64`.\nCreate \"Result\" as a New Subtractor of Left and Right.\nReturn Result.\n"
	c := compileDefinitionSource(t, source, "locals.l")
	definition, err := c.ComposeDefinition("Storage.Reduce")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(definition.Expression, "Key: ~native_reduce~") || !strings.Contains(definition.Expression, "Operands: &[left, right]&") || !strings.Contains(definition.Expression, "return result;") {
		t.Fatalf("local constructor lost references: %s", definition.Expression)
	}
}

func TestStaticFunctionBranchComposition(t *testing.T) {
	source := "    Publicly create \"Choose|-s|--ing|--r\" as a Generic Function with Inputs of \"flag\" as `bool` and \"left\" as `f64` and \"right\" as `f64` and Outputs of `f64`.\nIf Flag, Return Left, else Return Right.\n"
	c := compileDefinitionSource(t, source, "branch.l")
	testNativeDefinitionLibrary(t, c, "Choose", "branch", `#[path="src/branch.rs"] mod generated; fn main(){assert_eq!(generated::choose(true,1.0,2.0),1.0);assert_eq!(generated::choose(false,1.0,2.0),2.0);}`)
}

func TestStaticSignatureOwnsSelectedTypeDeclaration(t *testing.T) {
	source := "Define \"Carrier\" as `<DEF>HY{|Key: ~NativeCarrier~, Props: ~&[]&~|}<DEF>`.\n\n" +
		"    Publicly create \"Preserve|-s|-d|--ing|--er\" as a Generic Function with Inputs of \"value\" as Carrier and Outputs of Carrier.\nReturn Value.\n"
	c := compileDefinitionSource(t, source, "types.l")
	definition, err := c.ComposeDefinition("Preserve")
	if err != nil {
		t.Fatal(err)
	}
	closure, err := definitionClosure(definition)
	if err != nil {
		t.Fatal(err)
	}
	if len(closure) != 2 || closure[0].Symbol != "Carrier" || !strings.Contains(definition.Expression, "Type: NativeCarrier") || !strings.Contains(definition.Expression, "Outputs: ~&[(NativeCarrier)]&~") {
		t.Fatalf("signature declaration dependency lost: %+v", closure)
	}
}

func TestStaticFunctionReferenceOwnsDeclaration(t *testing.T) {
	source := "Define \"Hold|-s|---eld|-ing|-er\" to `<DEF>OB{|Key: ~<@Name@>~, Value: ~<@Value@>~|}<DEF>`.\n\n" +
		"    Publicly create \"Echo|-es|-ed|-ing|-er\" as a Generic Function with Inputs of \"value\" as `f64` and Outputs of `f64`.\nReturn Value.\n\n" +
		"    Publicly create \"Provide|-s|-d|--ing|--er\" as a Generic Function with No Inputs and Outputs of `fn(f64) -> f64`.\nCreate \"Callback\" as a New Holder of Echoer.\nReturn Callback.\n"
	c := compileDefinitionSource(t, source, "references.l")
	definition, err := c.ComposeDefinition("Provide")
	if err != nil {
		t.Fatal(err)
	}
	closure, err := definitionClosure(definition)
	if err != nil {
		t.Fatal(err)
	}
	if len(closure) != 2 || closure[0].Symbol != "Echo" || !strings.Contains(definition.Expression, "Value: ~echo~") {
		t.Fatalf("callback declaration dependency lost: %+v", closure)
	}
}

func TestStaticFunctionConditionalDeclaration(t *testing.T) {
	source := "Define \"Hold|-s|---eld|-ing|-er\" to `<DEF>OB{|Key: ~<@Name@>~, Value: ~<@Value@>~|}<DEF>`.\n\n" +
		"    Publicly create \"Choose|-s|--ing|--r\" as a Generic Function with Inputs of \"flag\" as `bool` and \"value\" as `f64` and Outputs of `f64`.\nIf Flag, Create \"Result\" as a New Holder of Value and then Return Result.\nReturn Value.\n"
	c := compileDefinitionSource(t, source, "conditional.l")
	definition, err := c.ComposeDefinition("Choose")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(definition.Expression, "Value: ~value~") || !strings.Contains(definition.Expression, "return result;") {
		t.Fatal(definition.Expression)
	}
}

func TestSourceSequenceComposesWithoutNesting(t *testing.T) {
	source := "Define \"Emit|-s|-ted|-ting|-ter\" to `<DEF>BOD^E{|Statement: ~println!(\"{}\", <@Value|Text@>);~|}<DEF>`.\n" +
		"Create \"Report|-s|-ed|-ing|-er\" for Processing a New Emitter of `first` and then Processing a New Emitter of `second`.\n\n" +
		"    Publicly create \"Notice\" as a Generic Function with No Inputs and Outputs of `()`.\nReport.\n"
	c := compileDefinitionSource(t, source, "sequence.l")
	aml, err := c.ComposeLibrary("Notice", "sequence")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Logiceme.NewLogicemeService(canonicalTestOptions().Resources).Evaluate(aml); err != nil {
		t.Fatal(err)
	}
	sequence := parseTestDefinition(t, `<DEF>W^<@Steps|Syllables@><DEF>`)
	modified := parseTestDefinition(t, `<DEF>BOD^W^<@Steps|Syllables@><DEF>`)
	leaf := DefinitionInstance{Expression: "BOD^E"}
	arguments := map[string]DefinitionArgument{"Steps": {Kind: "Syllables", Syllables: []string{leaf.Expression}, sequence: []DefinitionInstance{leaf}}}
	child := bindTestDefinition(t, modified, arguments)
	parent := bindTestDefinition(t, sequence, map[string]DefinitionArgument{"Steps": {Kind: "Syllables", Syllables: []string{child.Expression}, sequence: []DefinitionInstance{child}}})
	if child.sequence != nil || parent.Expression != "W^BOD^W^BOD^E" {
		t.Fatalf("non-sequence behavior was flattened: %+v", parent)
	}
}

func TestStaticFunctionDescriptorOwnsLocalTypeInputs(t *testing.T) {
	source := "\"Borrowing\" describes any Function as `<DEF>FW{|Func{Key: <@Name|Text@>, Inputs: <@Inputs|Props@>, Outputs: <@Outputs|Texts@>}Func, Lifetimes: ~'a~, GenericTypes: ~T: 'a~|}^(<@Process@>)<DEF>`.\n\n" +
		"    Publicly create \"Preserve\" as a Generic Borrowing Function with Inputs of \"value\" as `&'a T` and Outputs of `&'a T`.\nReturn Value.\n"
	c := compileDefinitionSource(t, source, "generic.l")
	aml, err := c.ComposeLibrary("Preserve", "generic")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Logiceme.NewLogicemeService(canonicalTestOptions().Resources).Evaluate(aml); err != nil {
		t.Fatal(err)
	}
	for _, local := range []string{"Lifetimes: ~'a~", "GenericTypes: ~T: 'a~"} {
		if !strings.Contains(string(aml), local) {
			t.Fatalf("selected Function descriptor lost %s: %s", local, aml)
		}
	}
	for _, local := range []string{"Lifetimes: ~~", "GenericTypes: ~ ~", "Deferred: ~true~", "Lifetimes: ~'a~, Lifetimes: ~'b~"} {
		template := parseTestDefinition(t, `<DEF>FW{|Func{Key: "invalid", Inputs: &[]&, Outputs: []}Func, `+local+`|}^W<DEF>`)
		if _, err := template.Instantiate(nil); err == nil {
			t.Fatalf("invalid local input accepted: %s", local)
		}
	}
}

func TestSourceTypeSlotRetainsSelectedDeclaration(t *testing.T) {
	source := "Define \"Carrier\" as `<DEF>HY{|Key: ~NativeCarrier~, Props: ~&[]&~|}<DEF>`.\n" +
		"Define \"Buffer\" as `<DEF>MOB{|Key: ~<@Name@>~, Type: ~Vec<<@Element|Type@>>~|}^KY{|Key: ~Vec::new~|}<DEF>`.\n\n" +
		"    Publicly create \"Collect\" as a Generic Function with No Inputs and Outputs of `Vec<NativeCarrier>`.\nCreate \"Result\" as a New Buffer of Carrier.\nReturn Result.\n"
	c := compileDefinitionSource(t, source, "type-slot.l")
	aml, err := c.ComposeLibrary("Collect", "typed")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(aml), "HY{|Key: ~NativeCarrier~") != 1 || !strings.Contains(string(aml), "Type: ~Vec<NativeCarrier>~") {
		t.Fatalf("selected type was quoted or lost its declaration: %s", aml)
	}
	if _, err := Logiceme.NewLogicemeService(canonicalTestOptions().Resources).Evaluate(aml); err != nil {
		t.Fatal(err)
	}
	template := parseTestDefinition(t, `<DEF>OB{|Key: ~value~, Type: ~<@Shape|Type@>~|}<DEF>`)
	value := bindTestDefinition(t, template, map[string]DefinitionArgument{"Shape": {Kind: "Text", Text: "Vec<u8>"}})
	if value.Expression != "OB{|Key: ~value~, Type: ~Vec<u8>~|}" {
		t.Fatal(value.Expression)
	}
	for _, invalid := range []DefinitionArgument{definitionReferenceArgument("runtime_value"), {Kind: "Text", Text: ""}, {Kind: "Definition", Definition: &DefinitionInstance{Expression: "BOD^E"}}} {
		if _, err := template.Instantiate(map[string]DefinitionArgument{"Shape": invalid}); err == nil {
			t.Fatalf("invalid type data accepted: %+v", invalid)
		}
	}
}

func TestSourceInvocationAtomOrderAndCardinality(t *testing.T) {
	template := parseTestDefinition(t, `<DEF>KY{|Key: ~<@Callee@>~, Inputs: ~&[<@Values|Atoms@>]&~|}<DEF>`)
	for _, arguments := range [][]DefinitionArgument{nil, {{Kind: "Number", Text: "2"}, {Kind: "Text", Text: "a\nb"}, definitionReferenceArgument("last")}} {
		bindings, err := definitionArguments(template, append([]DefinitionArgument{definitionReferenceArgument("selected")}, arguments...))
		if err != nil {
			t.Fatal(err)
		}
		instance, err := template.Instantiate(bindings)
		if err != nil {
			t.Fatal(err)
		}
		expected := "KY{|Key: ~selected~, Inputs: ~&[]&~|}"
		if len(arguments) > 0 {
			expected = `KY{|Key: ~selected~, Inputs: ~&[2_f64, "a\nb", last]&~|}`
		}
		if instance.Expression != expected {
			t.Fatalf("source invocation changed argument bytes, order or count: %s", instance.Expression)
		}
	}
	bindings, err := definitionArguments(template, []DefinitionArgument{definitionReferenceArgument("selected"), {Kind: "Definition", Definition: &DefinitionInstance{Expression: "BOD^E"}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := template.Instantiate(bindings); err == nil {
		t.Fatal("executable composition accepted as an invocation metadata atom")
	}
}

func TestSourceNativeDeclarationOwnsConstituentTypes(t *testing.T) {
	source := "Define \"Value\" as `<DEF>HY{|Key: ~Value~, Props: ~&[]&~|}<DEF>`.\n" +
		"Define \"Layout\" as `<DEF>HY{|Key: ~Container~, Props: ~&[Prop{Key: value, Type: Option<<@Type|Type@>>}Prop]&~|}<DEF>`.\n" +
		"Define \"Container\" as a New Layout of Value.\n\n" +
		"    Publicly create \"Keep|-s|--ept|-ing|-er\" as a Generic Function with Inputs of \"value\" as Container and Outputs of Container.\nReturn Value.\n"
	c := compileDefinitionSource(t, source, "types.l")
	declaration, err := c.ComposeDefinition("Keep")
	if err != nil {
		t.Fatal(err)
	}
	closure, err := definitionClosure(declaration)
	if err != nil {
		t.Fatal(err)
	}
	if len(closure) != 3 || closure[0].Symbol != "Value" || closure[1].Symbol != "Container" || closure[1].Span.Path != "types.l" {
		t.Fatalf("source constituent closure = %+v", closure)
	}
	if !strings.Contains(closure[1].Expression, "Type: Option<Value>") {
		t.Fatal(closure[1].Expression)
	}
	aml, err := c.ComposeLibrary("Keep", "types")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Logiceme.NewLogicemeService(canonicalTestOptions().Resources).Evaluate(aml); err != nil {
		t.Fatal(err)
	}
}
