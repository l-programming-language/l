package Lyger

import (
	"errors"
	"strings"
	"testing"

	"github.com/l-programming-language/l/lerrors"
)

func TestNamedOperandComposition(t *testing.T) {
	for _, word := range []string{"Finisher", "Completion", "Alias"} {
		t.Run(word, func(t *testing.T) {
			source := "Define \"Finish|-es|-ed|-ing|-er\" to `<DEF>BOD^E{|Statement: ~/* <@Label|Text@> */ return <@Statement@>;~|}<DEF>`.\n" +
				"Define \"Completion\" as a New Finisher.\nUse Completion as \"Alias\".\n" +
				"Create \"Respond|-s|-ed|-ing|-er\" for Processing a New " + word + " of `marker`.\n"
			c := compileDefinitionSource(t, source, "named-operands.l")
			rec, err := c.exportedRecord("Respond")
			if err != nil {
				t.Fatal(err)
			}
			bound, err := c.instantiateDefinition(rec, map[string]DefinitionArgument{"Statement": definitionReferenceArgument("result")})
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(bound.Expression, `/* "marker" */ return result;`) {
				t.Fatal(bound.Expression)
			}
			located := false
			for _, binding := range bound.Bindings {
				for _, span := range binding.Sources {
					located = located || span.Path == "named-operands.l" && span.Start.Line == 4
				}
			}
			if !located {
				t.Fatal("explicit argument lost its authored source location")
			}
			if _, err := c.instantiateDefinition(rec, map[string]DefinitionArgument{"Statement": definitionReferenceArgument("result"), "Extra": definitionReferenceArgument("ignored")}); !errors.Is(err, lerrors.InvalidInput) {
				t.Fatalf("unused argument: %v", err)
			}
			if _, err := c.instantiateNamedDefinition(rec, nil, make(map[string]bool), map[*record]bool{rec: true}); !errors.Is(err, lerrors.Cycle) {
				t.Fatalf("recursive definition: %v", err)
			}
		})
	}
}

func TestNamedActionLiteralOperands(t *testing.T) {
	c := compileDefinitionSource(t, "Use Function's Return as \"Return\".\nCreate \"Answer|-s|-ed|-ing|-er\" for Returning 7.\n", "named-values.l")
	rec, err := c.exportedRecord("Answer")
	if err != nil {
		t.Fatal(err)
	}
	bound, err := c.instantiateDefinition(rec, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(bound.Expression, "return 7_f64;") {
		t.Fatal(bound.Expression)
	}
}

func TestNamedConstructorArgumentEvaluatesOnce(t *testing.T) {
	c := compileDefinitionSource(t, "Define \"Next\" as `<DEF>KY{|Key: ~next~|}<DEF>`.\n"+
		"Define \"Pair\" as `<DEF>KY{|Key: ~pair~, Inputs: ~&[<@Value@>, <@Value@>]&~|}<DEF>`.\n"+
		"Define \"Copy\" as a New Pair.\nCreate \"Apply|-s|-ed|-ing|-er\" for Processing a New Copy of a New Next.\n", "named-once.l")
	rec, err := c.exportedRecord("Apply")
	if err != nil {
		t.Fatal(err)
	}
	bound, err := c.instantiateDefinition(rec, nil)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(bound.Expression, "Key: ~next~") != 1 {
		t.Fatal(bound.Expression)
	}
}
