package Lyger

import (
	"strings"
	"testing"
)

func TestTypedFormalMemberSelection(t *testing.T) {
	library := "Define any Thing's \"Render|-s|-ed|-ing|-er\" as a New Verb.\n" +
		"Define any Number's \"Render|-s|-ed|-ing|-er\" to `<DEF>BOD^E{|Statement: ~println!(\"owned {}\", <@Value|Number@>);~|}<DEF>`.\n" +
		"Create \"IO\" as a Dictionary.\nCreate IO's \"Print|-s|-ed|-ing|-er\" for Processing any Thing's Renderer.\n"
	source := "Use IO's Print from `io`.\n" +
		"Define any Number's \"Render|-s|-ed|-ing|-er\" to `<DEF>BOD^E{|Statement: ~println!(\"caller {}\", <@Value|Number@>);~|}<DEF>`.\n" +
		"Create \"Worker\" as a New Basic App.\nThe App Prints 2.\nThe App Prints 3.\n"
	for _, variant := range []string{"owned", "changed", "missing"} {
		t.Run(variant, func(t *testing.T) {
			files := dictionaryIndexFilesystem(t)
			selected := strings.Replace(library, "owned", variant, 1)
			if variant == "missing" {
				lines := strings.Split(selected, "\n")
				selected = strings.Join(append(lines[:1], lines[2:]...), "\n")
			}
			files["stdlib/eng/io/dictionary.l"].Data = []byte(selected)
			c := NewLygerService(Options{Resources: files}).NewCompilation([]byte(source), "member.l")
			c.Make()
			if variant == "missing" {
				if c.Error == nil || !strings.Contains(c.Error.Error(), "Number has no selected member Renderer") {
					t.Fatalf("missing specialization used a substitute: %v", c.Error)
				}
				return
			}
			if c.Error != nil {
				t.Fatal(c.Error)
			}
			body, err := c.lowerAppBody(c.Semantics.Apps[0])
			if err != nil || len(body) != 2 {
				t.Fatalf("source expansion: %+v, %v", body, err)
			}
			for _, step := range body {
				if strings.Count(step.Expression, variant+" {}") != 1 || strings.Contains(step.Expression, "caller") || strings.Contains(step.Expression, "default") || len(step.sequence) != 1 {
					t.Fatalf("selected definition was substituted or evaluated repeatedly: %s", step.Expression)
				}
				declaration := c.Semantics.Declarations[step.sequence[0].definition]
				if declaration == nil || declaration.source.Unit.Dictionary != "IO" {
					t.Fatalf("member lost its source ownership: %+v", declaration)
				}
			}
		})
	}
}

func TestTypedFormalFunctionDependenciesReachAppObject(t *testing.T) {
	files := dictionaryIndexFilesystem(t)
	files["stdlib/eng/io/dictionary.l"].Data = []byte("Create \"IO\" as a Dictionary.\nDefine any Thing's \"Render|-s|-ed|-ing|-er\" as a New Verb.\n" +
		"Define \"Discard|-s|-ed|-ing|-er\" to `<DEF>MOB{|Key: ~_~|}^(<@Value@>)<DEF>`.\n\n" +
		"    Publicly create any Number's \"Render|-s|-ed|-ing|-er\" as a Generic `owned` Function with Inputs of \"value\" as `f64` and Outputs of `f64`.\nReturn Value.\n\n" +
		"Create IO's \"Print|-s|-ed|-ing|-er\" for Processing a New Discarder of any Thing's Renderer.\n")
	c := NewLygerService(Options{Resources: files}).NewCompilation([]byte("Use IO's Print from `io`.\nCreate \"Worker\" as a New Basic App.\nThe App Prints 2.\nThe App Prints 3.\n"), "member.l")
	c.Make()
	if c.Error != nil {
		t.Fatal(c.Error)
	}
	aml := string(c.Program.BackendBytes)
	if strings.Count(aml, "FW{|Key: ~owned~") != 1 || strings.Count(aml, "KY{|Key: ~(owned)~") != 2 {
		t.Fatalf("App lost dependency identity or invocation cardinality: %s", aml)
	}
}
