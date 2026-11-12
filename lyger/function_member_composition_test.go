package Lyger

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/l-programming-language/l/lerrors"
	Logiceme "github.com/l-programming-language/l/logiceme"
)

func TestComposedFunctionMembers(t *testing.T) {
	for _, variant := range []string{"selected", "changed", "missing return component", "missing invoke component", "missing invocation arguments", "unused invocation argument", "missing core invoke", "program invoke cannot replace missing core"} {
		t.Run(variant, func(t *testing.T) {
			files := dictionaryIndexFilesystem(t)
			const core = "builtins/eng/core.l"
			source := string(files[core].Data)
			source = strings.Replace(source, "Define any Function's \"Return|-s|-ed|-ing|-er\" to `<DEF>BOD^E{|Statement: ~return <@Statement@>;~|}<DEF>`.", "Define \"Finish|-es|-ed|-ing|-er\" to `<DEF>BOD^E{|Statement: ~return /* composed_return */ <@Statement@>;~|}<DEF>`.\nCreate Function's \"Return|-s|-ed|-ing|-er\" for Processing a New Finisher.", 1)
			const invocation = "Define \"Invoke|-s|-d|--ing|--er\" to `<DEF>KY{|Key: ~(<@Target|Callable@>)~, Inputs: ~&[<@Arguments|Atoms@>]&~|}<DEF>`."
			if strings.Count(source, invocation) != 1 {
				t.Fatal("composition fixture requires the selected core Invoke definition")
			}
			source = strings.Replace(source, invocation, "Define \"Application|-s|---ve|--------ying\" as `<DEF>KY{|Key: ~(/* composed_invoke */ <@Target|Callable@>)~, Inputs: ~&[<@Arguments|Atoms@>]&~|}<DEF>`.\nDefine \"Forward|-s|-ed|-ing|-er\" to `<DEF><@Expression@><DEF>`.\nCreate \"Invoke|-s|-d|--ing|--er\" for Forwarding a New Application.", 1)
			if variant == "changed" {
				source = strings.ReplaceAll(source, "composed_", "changed_")
			}
			for name, prefix := range map[string]string{"missing return component": `Define "Finish`, "missing invoke component": `Define "Forward`} {
				if variant == name {
					lines := strings.Split(source, "\n")
					for i, line := range lines {
						if strings.HasPrefix(line, prefix) {
							lines[i] = ""
						}
					}
					source = strings.Join(lines, "\n")
				}
			}
			if variant == "missing invocation arguments" || variant == "unused invocation argument" {
				source = strings.Replace(source, ", Inputs: ~&[<@Arguments|Atoms@>]&~", "", 1)
			}
			if variant == "unused invocation argument" {
				// An ordinary metadata target remains a valid template, so the
				// compiler-supplied Arguments binding must be rejected as unused.
				source = strings.Replace(source, "<@Target|Callable@>", "<@Target@>", 1)
			}
			missingCore := variant == "missing core invoke" || variant == "program invoke cannot replace missing core"
			if missingCore {
				source = strings.Replace(source, "Create \"Invoke|-s|-d|--ing|--er\" for Forwarding a New Application.", "", 1)
			}
			files[core] = &fstest.MapFile{Data: []byte(source)}
			program := "Create \"Words\" as a Dictionary.\n\n    Publicly create Words's \"Echo|-es|-ed|-ing|-er\" as a Generic `echo` Function with Inputs of \"value\" as `f64` and Outputs of `f64`.\nReturn Value.\n\n    Publicly create Words's \"Supply|--ies|--ied|-ing|--ier\" as a Generic `supply` Function with No Inputs and Outputs of `f64`.\nReturn a New Words's Echo of 3.\n"
			if variant == "program invoke cannot replace missing core" {
				program = invocation + "\n\n" + program
			}
			c := NewLygerService(Options{Resources: files}).NewCompilation([]byte(program), "members.l")
			c.Make()
			var aml []byte
			err := c.Error
			if err == nil {
				aml, err = c.ComposeLibrary("Words", "members")
			}
			if missingCore {
				if !errors.Is(err, lerrors.InvalidConfiguration) || !strings.Contains(err.Error(), "builtin definition Invoke is absent") {
					t.Fatalf("missing selected core Invoke must reject without a fallback: %v", err)
				}
				return
			}
			if variant == "missing invocation arguments" {
				if !errors.Is(err, lerrors.Syntax) || !strings.Contains(err.Error(), "Callable") || !strings.Contains(err.Error(), "Atoms") {
					t.Fatalf("Callable without terminal Arguments must reject its typed shape: %v", err)
				}
				return
			}
			if variant == "missing return component" || variant == "missing invoke component" {
				if err == nil {
					t.Fatal("missing composition component compiled with a fallback")
				}
				return
			}
			if variant == "unused invocation argument" {
				if !errors.Is(err, lerrors.InvalidInput) || !strings.Contains(err.Error(), "unused argument Arguments") {
					t.Fatalf("unused invocation argument: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			bundle, err := Logiceme.NewLogicemeService(files).Evaluate(aml)
			if err != nil {
				t.Fatal(err)
			}
			for _, marker := range []string{"composed_return", "composed_invoke"} {
				if variant == "changed" {
					marker = strings.Replace(marker, "composed_", "changed_", 1)
				}
				if !bytes.Contains(aml, []byte(marker)) || !strings.Contains(generatedRustSource(bundle), marker) {
					t.Fatalf("selected composition %s missing", marker)
				}
			}
			if !bytes.Equal(files[core].Data, []byte(source)) {
				t.Fatal("composition mutated selected definitions")
			}
		})
	}
}
