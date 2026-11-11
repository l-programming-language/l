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
		"")...)}
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
			defer c.definitionScope()()
			for _, input := range []string{"first_operand", "second_operand"} {
				value := "()"
				if test.operand != "" {
					value = input
				}
				rec, err := c.builtinRecord(test.word)
				if err != nil {
					t.Fatal(err)
				}
				node, err := c.definitionConstruction(rec, test.descriptors...)
				if err != nil {
					t.Fatal(err)
				}
				var values []DefinitionArgument
				if test.operand != "" {
					values = []DefinitionArgument{definitionReferenceArgument(input)}
				}
				bound, _, err := c.instantiateStaticConstruction(node, values, make(map[*record]bool), nil)
				if err != nil {
					t.Fatal(err)
				}
				if bound.result == nil || strings.Count(bound.Expression, "Key: ~"+bound.result.reference+"~, Value: ~"+value+"~") != 1 {
					t.Fatalf("selected constructor lost explicit result or operand: %+v", bound)
				}
				if bound.Span.Path != core {
					t.Fatalf("selected constructor lost definition-site provenance: %s", bound.Span.Path)
				}
			}
		})
	}
	defer c.definitionScope()()
	var steps []DefinitionInstance
	for index := 0; index < 2; index++ {
		rec, err := c.builtinRecord("EmptyParcel")
		if err != nil {
			t.Fatal(err)
		}
		node, err := c.definitionConstruction(rec)
		if err != nil {
			t.Fatal(err)
		}
		instance, _, err := c.instantiateStaticConstruction(node, nil, make(map[*record]bool), nil)
		if err != nil {
			t.Fatal(err)
		}
		steps = append(steps, instance)
	}
	stages := composeDefinitionSequence(steps)
	if _, err := c.atomArgument(DefinitionArgument{Kind: "Definition", Definition: &stages}); err == nil || !strings.Contains(err.Error(), "requires an explicit result binding") {
		t.Fatalf("multi-stage construction acquired an implicit result: %v", err)
	}
	testNativeDefinitionLibrary(t, c, "Exercise", "forwarded_result", `static CALLS:std::sync::atomic::AtomicUsize=std::sync::atomic::AtomicUsize::new(0);pub fn acquire()->f64{CALLS.fetch_add(1,std::sync::atomic::Ordering::SeqCst);7.0}#[path="src/forwarded_result.rs"]mod generated;fn main(){assert_eq!(generated::exercise(),7.0);assert_eq!(CALLS.load(std::sync::atomic::Ordering::SeqCst),1);}`)
}
