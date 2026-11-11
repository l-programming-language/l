package Lyger

import (
	"errors"

	"strings"
	"testing"
	"testing/fstest"

	"github.com/l-programming-language/l/lerrors"
)

func TestValueBindingForwardingPreservesCaptureBoundary(t *testing.T) {
	for _, word := range []string{"Capture", "Forwarded", "Routed Carrier", "Nested"} {
		t.Run(word, func(t *testing.T) {
			source := "If false, define \"Capture\" as `<DEF>MOB{|Key: ~<@Held|Label@>~|}^KY{|Key: ~std::convert::identity~, Inputs: ~&[<@Value@>]&~|}<DEF>`.\n" +
				"Define \"Carrier\" as `<DEF>MOB{|Key: ~<@Held|Label@>~|}^KY{|Key: ~std::convert::identity~, Inputs: ~&[<@Value@>]&~|}<DEF>`.\n" +
				"Define \"Routed\" as a New Adjective describing any Carrier.\nIt describes any Carrier as a New Capture.\n" +
				"Define \"Forwarded\" as a New Capture.\nDefine \"Nested\" as a New Carrier of a New Capture.\n" +
				"Create \"Speak|-s|-ed|-ing\" for Processing a New " + word + ".\nCreate \"Worker\" as a New Basic App.\nThe App Speaks 7.\n"
			_, err := distributionFrontend(t, source, nil)
			if !errors.Is(err, lerrors.Unsupported) || !strings.Contains(err.Error(), "runtime value Capture cannot be captured") {
				t.Fatalf("forwarded construction lost canonical capture rejection: %v", err)
			}
		})
	}
}

func TestValueBindingUsesSelectedResultLabel(t *testing.T) {
	files := dictionaryIndexFilesystem(t)
	const core = "builtins/eng/core.l"
	files[core] = &fstest.MapFile{Data: append(append([]byte(nil), files[core].Data...), []byte("\n"+
		"Define \"Carrier\" as `<DEF>OB{|Key: ~<@Held|Label@>~, Value: ~<@Value@>~|}<DEF>`.\n"+
		"Define \"Forwarded\" as a New Carrier.\nUse Forwarded as \"Relayed\".\n"+
		"Define \"Marked\" as a New Adjective describing any Carrier.\n"+
		"It describes any Carrier as `<DEF>OB{|Key: ~<@Written|Label@>~, Value: ~<@Operand@>~|}<DEF>`.\n"+
		"Define \"Parcel\" as a New Marked Carrier.\n"+
		"Define \"Routed\" as a New Adjective describing any Carrier.\nIt describes any Carrier as a New Parcel.\n"+
		"Define \"Vacant\" as `<DEF>OB{|Key: ~<@Receipt|Label@>~, Value: ~()~|}<DEF>`.\nDefine \"EmptyParcel\" as a New Vacant.\n"+
		"Define \"Stages\" as a New Process of a New Carrier and a New Vacant.\n")...)}
	const source = "Define \"Acquire\" as `<DEF>MOB{|Key: ~<@Acquired|Label@>~|}^KY{|Key: ~crate::acquire~|}<DEF>`.\n\n" +
		"    Publicly create \"Exercise\" as a Generic Function with No Inputs and Outputs of `f64`.\nReturn a New Routed Carrier of a New Acquire.\n"
	c := NewLygerService(Options{Language: "eng", Resources: files}).NewCompilation([]byte(source), "result-label.l")
	c.Make()
	if c.Error != nil {
		t.Fatal(c.Error)
	}
	for _, test := range []struct {
		name, word, operand string
		descriptors         []string
	}{
		{name: "ordinary", word: "Carrier", operand: "Value"},
		{name: "descriptor", word: "Carrier", operand: "Operand", descriptors: []string{"Marked"}},
		{name: "forwarded", word: "Forwarded", operand: "Value"},
		{name: "alias", word: "Relayed", operand: "Value"},
		{name: "selected source", word: "Parcel", operand: "Operand"},
		{name: "source descriptor", word: "Carrier", operand: "Operand", descriptors: []string{"Routed"}},
		{name: "zero operands", word: "EmptyParcel"},
	} {
		t.Run(test.name, func(t *testing.T) {
			emitter := orchestrationEmitter{compilation: c}
			for _, input := range []string{"first_operand", "second_operand"} {
				arguments, value := map[string]DefinitionArgument{}, "()"
				if test.operand != "" {
					arguments[test.operand], value = definitionReferenceArgument(input), input
				}
				result, err := emitter.describedValueBinding(test.word, test.descriptors, arguments)
				if err != nil {
					t.Fatal(err)
				}
				if len(emitter.pending) == 0 || strings.Count(emitter.pending[len(emitter.pending)-1], "Key: ~"+result+"~, Value: ~"+value+"~") != 1 {
					t.Fatalf("selected constructor lost explicit result or operand: %v", emitter.pending)
				}
				instance, err := c.instantiateDescribedBuiltin(test.word, test.descriptors, arguments)
				if err != nil || instance.Span.Path != core {
					t.Fatalf("selected constructor lost definition-site provenance: %s, %v", instance.Span.Path, err)
				}
			}
		})
	}
	emitter := orchestrationEmitter{compilation: c}
	if _, err := emitter.valueBinding("Stages", nil); err == nil || !strings.Contains(err.Error(), "requires an explicit result binding") {
		t.Fatalf("multi-stage construction acquired an implicit result: %v", err)
	}
	testNativeDefinitionLibrary(t, c, "Exercise", "forwarded_result", `static CALLS:std::sync::atomic::AtomicUsize=std::sync::atomic::AtomicUsize::new(0);pub fn acquire()->f64{CALLS.fetch_add(1,std::sync::atomic::Ordering::SeqCst);7.0}#[path="src/forwarded_result.rs"]mod generated;fn main(){assert_eq!(generated::exercise(),7.0);assert_eq!(CALLS.load(std::sync::atomic::Ordering::SeqCst),1);}`)
}
