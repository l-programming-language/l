package Lyger

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"testing/fstest"

	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	"github.com/l-programming-language/l/lerrors"
	Logiceme "github.com/l-programming-language/l/logiceme"
)

func TestComposedFunctionMembers(t *testing.T) {
	for _, variant := range []string{"selected", "changed", "missing return component", "missing invoke component", "recursive value binding", "unused invocation argument"} {
		t.Run(variant, func(t *testing.T) {
			files := dictionaryIndexFilesystem(t)
			const core = "builtins/eng/core.l"
			source := string(files[core].Data)
			source = strings.Replace(source, "Define any Function's \"Return|-s|-ed|-ing|-er\" to `<DEF>BOD^E{|Statement: ~return <@Statement@>;~|}<DEF>`.", "Define \"Finish|-es|-ed|-ing|-er\" to `<DEF>BOD^E{|Statement: ~return /* composed_return */ <@Statement@>;~|}<DEF>`.\nCreate Function's \"Return|-s|-ed|-ing|-er\" for Processing a New Finisher.", 1)
			source = strings.Replace(source, "Define any Function's \"Invoke|-s|-d|--ing|--er\" to `<DEF>KY{|Key: ~(<@Target@>)~, Inputs: ~&[<@Arguments|Atoms@>]&~|}<DEF>`.", "Define \"Application\" as `<DEF>KY{|Key: ~(/* composed_invoke */ <@Target@>)~, Inputs: ~&[<@Arguments|Atoms@>]&~|}<DEF>`.\nDefine \"Forward|-s|-ed|-ing|-er\" to `<DEF><@Expression@><DEF>`.\nCreate Function's \"Invoke|-s|-d|--ing|--er\" for Forwarding a New Application.", 1)
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
			if variant == "unused invocation argument" {
				source = strings.Replace(source, ", Inputs: ~&[<@Arguments|Atoms@>]&~", "", 1)
			}
			if variant == "recursive value binding" {
				const storage = "stdlib/eng/storage/dictionary.l"
				data := strings.Replace(string(files[storage].Data), "MOB{|Key: ~<@Name|Label@>~|}^(<@Value@>)", "OB{|Key: ~<@Name|Label@>~, Value: ~<@Value@>~|}", 1)
				files[storage] = &fstest.MapFile{Data: []byte(data)}
			}
			files[core] = &fstest.MapFile{Data: []byte(source)}
			c := NewLygerService(Options{Resources: files}).NewCompilation([]byte("Create \"Words\" as a Dictionary.\n\n    Publicly create Words's \"Echo\" as a Generic `echo` Function with Inputs of \"value\" as `f64` and Outputs of `f64`.\nReturn Value.\n\n    Publicly create Words's \"Supply\" as a Generic `supply` Function with No Inputs and Outputs of `f64`.\nReturn a New Words's Echo of 3.\n"), "members.l")
			c.Make()
			if strings.HasPrefix(variant, "missing") {
				if c.Error == nil {
					t.Fatal("missing composition component compiled with a fallback")
				}
				return
			}
			if c.Error != nil {
				t.Fatal(c.Error)
			}
			aml, err := c.ComposeLibrary("Words", "members")
			if variant == "recursive value binding" {
				if !errors.Is(err, lerrors.Cycle) {
					t.Fatalf("non-consuming value binder: %v", err)
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
				if !bytes.Contains(aml, []byte(marker)) || !strings.Contains(builtinProcessRust(bundle), marker) {
					t.Fatalf("selected member %s missing", marker)
				}
			}
			if !bytes.Equal(files[core].Data, []byte(source)) {
				t.Fatal("composition mutated selected definitions")
			}
			if variant == "selected" {
				deriveTestRuntime(t, files)
				ordinary := "Define \"Prefix\" as `capture`.\n\n" + functionSource("Read", "", "Text", "Return Prefix.") + "Read as \"Result\".\nShow Result.\n"
				result, unit, bundle, err := builtinProcessGenerate(files, ordinary)
				if err != nil {
					t.Fatal(err)
				}
				captured := false
				for _, function := range result.Compilation.Semantics.Functions {
					captured = captured || len(function.Captures) > 0
				}
				if !captured || !bytes.Contains(unit.BackendAML, []byte("composed_return")) || !bytes.Contains(unit.BackendAML, []byte("composed_invoke")) {
					t.Fatal("ordinary captured Function lost its composed contracts")
				}
				binary := buildActualBundle(t, bundle, "entry", true, "")
				stdout, stderr, err := nativetest.RuntimeExecution.Run(t, binary, "", nil)
				if err != nil || len(stderr) != 0 || string(stdout) != "capture" {
					t.Fatalf("composed captures stdout=%q stderr=%q: %v", stdout, stderr, err)
				}
			}
		})
	}
}
