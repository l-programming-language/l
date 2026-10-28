package Lyger

import (
	"bytes"
	"strings"
	"testing"
	"testing/fstest"

	Logiceme "github.com/l-programming-language/l/logiceme"
)

// These independently authored resource mutations prove that both ordinary
// Functions and Main construction consume the selected canonical definitions.
func TestCanonicalFunctionDefinitionsOwnEmission(t *testing.T) {
	const core = "builtins/eng/core.l"
	for _, test := range []struct{ member, from, to string }{
		{"Function", "^(<@Process@>)", "^(W^(BOD^E{|Statement: ~/* selected_Function */~|})*<@Process@>)"},
		{"Return", "return <@Statement@>;", "return /* selected_Return */ <@Statement@>;"},
		{"Suspend", "Statement: ~lPinned~", "Statement: ~/* selected_Suspend */ lPinned~"},
		{"Invoke", "Value: ~<@Target@>~", "Value: ~/* selected_Invoke */ <@Target@>~"},
		{"Yield", "Key: ~std::ops::ControlFlow::Continue~", "Key: ~/* selected_Yield */ std::ops::ControlFlow::Continue~"},
		{"Bind", "Key: ~l_environment_new~", "Key: ~/* selected_Bind */ l_environment_new~"},
		{"Exit", "Key: ~Ok~", "Key: ~/* selected_Exit */ Ok~"},
	} {
		t.Run(test.member, func(t *testing.T) {
			files := dictionaryIndexFilesystem(t)
			original := bytes.Clone(files[core].Data)
			lines := strings.Split(string(original), "\n")
			changed := false
			for i, line := range lines {
				if !strings.Contains(line, `"`+test.member) || !strings.Contains(line, test.from) {
					continue
				}
				lines[i] = strings.Replace(line, test.from, test.to, 1)
				changed = true
				break
			}
			if !changed {
				t.Fatal("test mutation did not select its builtin definition")
			}
			selected := []byte(strings.Join(lines, "\n"))
			files[core] = &fstest.MapFile{Data: selected}
			source := identityFunction + "Identity `value` as \"Result\".\nCreate \"Child\" as a New Basic App.\n"
			if test.member == "Return" {
				source = strings.Replace(source, "Return Value.", "Use Return as \"Finish\".\nFinish Value.", 1)
			}
			result, unit, bundle, err := builtinProcessGenerate(files, source)
			if err != nil {
				t.Fatal(err)
			}
			marker := "/* selected_" + test.member + " */"
			if !bytes.Contains(unit.BackendAML, []byte(marker)) || !strings.Contains(builtinProcessRust(bundle), marker) {
				t.Fatal("selected definition did not control ordinary Function AML and Rust")
			}
			if test.member == "Function" {
				units, err := result.ApplicationUnits()
				if err != nil || len(units) != 1 {
					t.Fatalf("Main construction: %v", err)
				}
				child, err := Logiceme.NewLogicemeService(files).Evaluate(units[0].BackendAML)
				if err != nil || !bytes.Contains(units[0].BackendAML, []byte(marker)) || !strings.Contains(builtinProcessRust(child), marker) {
					t.Fatalf("Main did not use the same Function definition: %v", err)
				}
			}
			if !bytes.Equal(files[core].Data, selected) {
				t.Fatal("generation mutated canonical source")
			}
		})
	}
}

func TestCanonicalFunctionMissingMemberDoesNotFallBack(t *testing.T) {
	for _, member := range []string{"Return", "Suspend", "Invoke", "Yield", "Bind", "Exit"} {
		t.Run(member, func(t *testing.T) {
			files := dictionaryIndexFilesystem(t)
			const core = "builtins/eng/core.l"
			var kept []string
			for _, line := range strings.Split(string(files[core].Data), "\n") {
				if strings.HasPrefix(line, `Define any Function's "`+member) {
					continue
				}
				kept = append(kept, line)
			}
			files[core] = &fstest.MapFile{Data: []byte(strings.Join(kept, "\n"))}
			_, _, _, err := builtinProcessGenerate(files, functionSource("Work", "", "", "Show `body`.")+"Work.\n")
			if err == nil || !strings.Contains(err.Error(), member) {
				t.Fatalf("missing bound %s composition did not reject: %v", member, err)
			}
		})
	}
}

func TestFunctionInvocationUsesEnclosingReturnComposition(t *testing.T) {
	c := executableTestResult(t, memberResolutionSource{main: "    Publicly create \"Done\" as a New Function with No Inputs and No Outputs.\nReturn.\n"}).Compilation
	var definition *sourceFunction
	for _, function := range c.Semantics.Functions {
		definition = function
	}
	if definition == nil {
		t.Fatal("missing Function")
	}
	caller, callee := *definition, *definition
	caller.ReturnTemplate = parseTestDefinition(t, strings.Replace(string(definition.ReturnTemplate.Original), "return ", "return /* caller */ ", 1))
	callee.ReturnTemplate = parseTestDefinition(t, strings.Replace(string(definition.ReturnTemplate.Original), "return ", "return /* callee */ ", 1))
	callee.Members = make(map[string]*DefinitionTemplate, len(definition.Members))
	for name, template := range definition.Members {
		callee.Members[name] = template
	}
	callee.Members["Invoke"] = parseTestDefinition(t, strings.Replace(string(definition.Members["Invoke"].Original), "Value: ~<@Target@>~", "Value: ~/* selected invocation */ <@Target@>~", 1))
	emitter := orchestrationEmitter{compilation: c,
		plan: &orchestrationPlan{Identities: map[*record]uint64{callee.Definition: 1}}, functionDepth: 1, functionWords: &caller,
		functionContracts: map[*record]*sourceFunction{callee.Definition: &callee},
	}
	statements, err := emitter.functionInstruction(&orchestrationInstruction{
		Source: &instruction{Kind: instructionInvoke},
		Target: &orchestrationValue{Type: valueFunction, RuntimeType: callee.signature(), Storage: &orchestrationStorage{}},
	})
	if err != nil {
		t.Fatal(err)
	}
	aml := strings.Join(statements, "*")
	if !strings.Contains(aml, "/* selected invocation */") || !strings.Contains(aml, "/* caller */") || strings.Contains(aml, "/* callee */") {
		t.Fatalf("invocation or enclosing return lost its selected definition: %s", aml)
	}
}
