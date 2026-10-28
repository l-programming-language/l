package Lyger

import (
	"bytes"
	"fmt"
	"io/fs"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	distribution "github.com/l-programming-language/l"
	Logiceme "github.com/l-programming-language/l/logiceme"
)

const builtinProcessDirectory = "builtins/operations/rust/"

// These definitions are independently authored test INPUTS, not generated-output
// replacements. Default-resource tests never use this overlay. The unannotated
// slots belong to the private compiler binder, not the public argument API.
func builtinProcessDefinitions() map[string]string {
	return map[string]string{
		"build":               `<DEF>MOB{|Key: ~lProcessResult~|}^KY{|Key: ~runtime::app_build~, Inputs: ~&[&<@Receiver@>]&~|}*OB{|Key: ~_~, Value: ~lProcessResult?~|}<DEF>`,
		"run":                 `<DEF>MOB{|Key: ~lProcessResult~|}^KY{|Key: ~runtime::app_run~, Inputs: ~&[&<@Receiver@>]&~|}*OB{|Key: ~_~, Value: ~lProcessResult?~|}<DEF>`,
		"stop":                `<DEF>MOB{|Key: ~lProcessFuture~|}^KY{|Key: ~runtime::app_stop~, Inputs: ~&[&<@Receiver@>]&~|}*COB{|Key: ~lProcessResult~|}^DV{|Key: ~lProcessFuture~|}*OB{|Key: ~_~, Value: ~lProcessResult?~|}<DEF>`,
		"await":               `<DEF>[["MOB{|Key: ~lProcessFuture~|}^KY{|Key: ~runtime::app_wait~, Inputs: ~&[&<@Receiver@>]&~|}*COB{|Key: ~lProcessResult~|}^DV{|Key: ~lProcessFuture~|}*OB{|Key: ~<@Invocation@>~, Value: ~lProcessResult?~|}", "MOB{|Key: ~lProcessClaim~|}^KY{|Key: ~runtime::invocation_claim~, Inputs: ~&[&<@Invocation@>]&~|}*OB{|Key: ~lProcessClaimed~, Value: ~lProcessClaim?~|}*BOG^GI^(E{|Statement: ~lProcessClaimed~|}*<@Body@>)"]]<DEF>`,
		"finish-continuation": `<DEF>MOB{|Key: ~_~|}^KY{|Key: ~runtime::invocation_finish~, Inputs: ~&[&<@Invocation@>]&~|}<DEF>`,
	}
}

func builtinProcessPath(operation string) string {
	return builtinProcessDirectory + "process-" + operation + ".def"
}

func builtinProcessFilesystem(t *testing.T) fstest.MapFS {
	t.Helper()
	files := dictionaryIndexFilesystem(t)
	for operation, definition := range builtinProcessDefinitions() {
		files[builtinProcessPath(operation)] = &fstest.MapFile{Data: []byte(definition)}
	}
	return files
}

// Exercise the public frontend and real backend. An error may originate in
// compilation or source projection, but invalid assets must never be ignored.
func builtinProcessGenerate(resources fs.FS, source string) (*CompilationResult, BuildUnit, Logiceme.Executable, error) {
	cfg := Options{Resources: resources}
	result, err := NewLygerService(cfg).Compile([]byte(source), "builtin-process-contract.l")
	if err != nil {
		return result, BuildUnit{}, Logiceme.Executable{}, err
	}
	unit, err := result.OrchestratorUnit("entry", executableTestCompanions(result))
	if err != nil {
		return result, unit, Logiceme.Executable{}, err
	}
	bundle, err := Logiceme.NewLogicemeService(cfg.Resources).Evaluate(unit.BackendAML, Logiceme.GenerationOptions{IncludeRuntime: true})
	return result, unit, bundle, err
}

func builtinProcessRequire(t *testing.T, resources fs.FS, source string) (*CompilationResult, BuildUnit, Logiceme.Executable) {
	t.Helper()
	result, unit, bundle, err := builtinProcessGenerate(resources, source)
	if err != nil {
		t.Fatalf("real App generation prerequisite failed: %v", err)
	}
	return result, unit, bundle
}

func builtinProcessRust(bundle Logiceme.Executable) string {
	var result strings.Builder
	for i, paths := range bundle.MapArrays {
		for j, path := range paths {
			if strings.HasSuffix(path, ".rs") {
				result.Write(bundle.ByteArrays[i][j])
			}
		}
	}
	return result.String()
}

const builtinProcessSource = "Create \"Worker\" as a New Basic App.\nBuild Worker.\nRun Worker.\nStop Worker.\nAwait Worker.\n"

func TestBuiltinProcessLogemeCatalog(t *testing.T) {
	t.Run("authored definition syntax control", func(t *testing.T) {
		for operation, raw := range builtinProcessDefinitions() {
			template, err := ParseDefinitionTemplate([]byte(raw), SourceSpan{Path: builtinProcessPath(operation)})
			if err != nil {
				t.Fatal(err)
			}
			// This only checks fixture composition syntax. Production selection
			// must still perform its own private checked binding of typed values.
			bindings := map[string]DefinitionArgument{
				"Receiver":   {Kind: "Syllable", Text: "lApp0"},
				"Invocation": {Kind: "Syllable", Text: "lInvocation0"},
				"Body":       {Kind: "Syllable", Text: "W^BOD^E{|Statement: ~{}~|}"},
			}
			instance, err := template.Instantiate(bindings)
			if err != nil {
				t.Fatalf("invalid independently authored %s fixture: %v", operation, err)
			}
			if operation == "await" && (len(instance.Words) != 1 || len(instance.Words[0]) != 2) {
				t.Fatalf("Await must retain binding then guarded continuation: %+v", instance)
			}
		}
	})
	t.Run("default definitions and closed signatures", func(t *testing.T) {
		for operation := range builtinProcessDefinitions() {
			t.Run(operation, func(t *testing.T) {
				path := builtinProcessPath(operation)
				raw, err := fs.ReadFile(distribution.Files(), path)
				if err != nil {
					t.Fatalf("default typed process implementation %s is unavailable: %v", path, err)
				}
				template, err := ParseDefinitionTemplate(raw, SourceSpan{Path: path})
				if err != nil {
					t.Fatal(err)
				}
				want := map[string]bool{"Receiver": true}
				if operation == "await" {
					want = map[string]bool{"Receiver": true, "Invocation": true, "Body": true}
				} else if operation == "finish-continuation" {
					want = map[string]bool{"Invocation": true}
				}
				for _, slot := range template.Slots {
					if !want[slot.Name] || slot.Type != "" {
						t.Errorf("%s exposes unexpected/private-slot annotation: %+v", path, slot)
					}
					delete(want, slot.Name)
				}
				if len(want) != 0 {
					t.Errorf("%s missing typed slots: %v", path, want)
				}
			})
		}
	})
	t.Run("every implementation reads selected resources", func(t *testing.T) {
		for operation, definition := range builtinProcessDefinitions() {
			t.Run(operation, func(t *testing.T) {
				files := builtinProcessFilesystem(t)
				marker := "/* selected_process_" + strings.ReplaceAll(operation, "-", "_") + " */ "
				field := "Key: ~runtime::"
				if operation == "await" {
					field = "Value: ~lProcessResult?"
				}
				if strings.Count(definition, field) != 1 {
					t.Fatalf("%s fixture requires one selected marker location", operation)
				}
				definition = strings.Replace(definition, field, strings.Replace(field, "~", "~"+marker, 1), 1)
				files[builtinProcessPath(operation)] = &fstest.MapFile{Data: []byte(definition)}
				before := bytes.Clone(files[builtinProcessPath(operation)].Data)
				_, unit, bundle := builtinProcessRequire(t, files, builtinProcessSource)
				if !bytes.Contains(unit.BackendAML, []byte(marker)) || !strings.Contains(builtinProcessRust(bundle), marker) {
					t.Errorf("%s resource was ignored: selected marker must survive actual AML and Rust", operation)
				}
				if !bytes.Equal(before, files[builtinProcessPath(operation)].Data) {
					t.Fatal("typed instantiation mutated the resource prototype")
				}
			})
		}
	})
	t.Run("missing implementations cannot fall back", func(t *testing.T) {
		for operation := range builtinProcessDefinitions() {
			t.Run(operation, func(t *testing.T) {
				files := builtinProcessFilesystem(t)
				path := builtinProcessPath(operation)
				delete(files, path)
				_, _, _, err := builtinProcessGenerate(files, builtinProcessSource)
				if err == nil || !strings.Contains(err.Error(), path) {
					t.Fatalf("missing selected definition must fail with its resource path; got %v", err)
				}
			})
		}
	})
	t.Run("malformed implementation diagnostics", func(t *testing.T) {
		for _, tc := range []struct{ name, operation, definition, context string }{
			{"unpaired marker", "run", `<DEF>BOD^E{|Statement: ~<@Receiver@>.run()?;~|}`, ""},
			{"nested marker", "run", `<DEF><DEF>BOD^E{|Statement: ~<@Receiver@>.run()?;~|}<DEF>`, ""},
			{"empty composition", "run", `<DEF><DEF>`, ""},
			{"unbalanced composition", "run", `<DEF>(BOD^E{|Statement: ~<@Receiver@>.run()?;~|}<DEF>`, ""},
			{"unknown slot", "run", `<DEF>BOD^E{|Statement: ~<@Other@>.run()?;~|}<DEF>`, "Other"},
			{"missing receiver", "run", `<DEF>BOD^E{|Statement: ~{}~|}<DEF>`, "Receiver"},
			{"text receiver", "run", `<DEF>BOD^E{|Statement: ~<@Receiver|Text@>.run()?;~|}<DEF>`, "Receiver"},
			{"native public annotation", "run", `<DEF>BOD^E{|Statement: ~<@Receiver|Native@>.run()?;~|}<DEF>`, "Receiver"},
			{"await loses typed body", "await", `<DEF>OB{|Key: ~<@Invocation@>~, Value: ~<@Receiver@>.wait().await?~|}<DEF>`, "Body"},
			{"finish loses invocation", "finish-continuation", `<DEF>BOD^E{|Statement: ~{}~|}<DEF>`, "Invocation"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				files := builtinProcessFilesystem(t)
				path := builtinProcessPath(tc.operation)
				files[path] = &fstest.MapFile{Data: []byte(tc.definition)}
				_, _, _, err := builtinProcessGenerate(files, builtinProcessSource)
				if err == nil || !strings.Contains(err.Error(), path) || !strings.Contains(err.Error(), tc.context) {
					t.Fatalf("invalid %s must fail with resource/slot provenance; got %v", tc.name, err)
				}
			})
		}
	})
	t.Run("selected renamed descriptor preserves App composition", func(t *testing.T) {
		files := builtinProcessFilesystem(t)
		const core = "builtins/eng/core.l"
		if bytes.Count(files[core].Data, []byte(`"Main" is a New Adjective`)) != 1 {
			t.Fatal("renamed descriptor input requires the canonical builtin declaration")
		}
		files[core] = &fstest.MapFile{Data: bytes.ReplaceAll(files[core].Data, []byte("Main"), []byte("Entry"))}
		deriveTestRuntime(t, files)
		builtinProcessRequire(t, files, builtinProcessSource)
		cfg := Options{Resources: files}
		result, err := runnableServiceCompile(t, "Create \"Worker\" as a New Entry Service.\n", nil, cfg)
		if err != nil {
			t.Fatal(err)
		}
		plan, err := result.WorkspacePlan()
		if err != nil || len(plan.Packages()) != 0 || len(result.Compilation.Program.Apps) != 0 {
			t.Fatalf("renamed descriptor granted a standalone Service an artifact: %v", err)
		}
		_, err = runnableServiceCompile(t, "Create \"Worker\" as a New Entry Service.\nRun Worker.\n", nil, cfg)
		if err == nil || !strings.Contains(err.Error(), "App operand") {
			t.Fatalf("renamed descriptor granted Service process eligibility: %v", err)
		}
	})
	t.Run("selection isolation and fresh receivers", func(t *testing.T) {
		source := "Create \"First\" as a New Basic App.\nCreate \"Second\" as a New Basic App.\nRun First.\nRun Second.\nAwait First.\nAwait Second.\n"
		var previous Logiceme.Executable
		for _, label := range []string{"alpha", "beta", ""} {
			files := builtinProcessFilesystem(t)
			if label != "" {
				definition := strings.Replace(builtinProcessDefinitions()["run"], "Key: ~runtime::", "Key: ~/* "+label+" */ runtime::", 1)
				files[builtinProcessPath("run")] = &fstest.MapFile{Data: []byte(definition)}
			}
			var provider fs.FS = files
			if label == "" {
				provider = distribution.Files()
			}
			_, unit, bundle := builtinProcessRequire(t, provider, source)
			for i := 0; i < 2; i++ {
				want := fmt.Sprintf("Inputs: ~&[&lApp%d]&~", i)
				if !bytes.Contains(unit.BackendAML, []byte(want)) {
					t.Errorf("%s omitted independently bound receiver %s", label, want)
				}
			}
			rust := builtinProcessRust(bundle)
			for _, candidate := range []string{"alpha", "beta"} {
				wantCount := 0
				if candidate == label {
					wantCount = 2
				}
				if got := strings.Count(rust, "/* "+candidate+" */"); got != wantCount {
					t.Errorf("selection %q has %d %s markers, want %d", label, got, candidate, wantCount)
				}
			}
			if label == "beta" && !strings.Contains(builtinProcessRust(previous), "/* alpha */") {
				t.Error("earlier bundle lost its independent resource selection")
			}
			previous = bundle
		}
	})
	t.Run("public Native argument remains unsupported", func(t *testing.T) {
		template := parseTestDefinition(t, `<DEF>BOD^E{|Statement: ~<@Receiver|Native@>.run()?;~|}<DEF>`)
		if _, err := template.Instantiate(map[string]DefinitionArgument{"Receiver": {Kind: "Native", Text: "process"}}); err == nil {
			t.Fatal("private process binding leaked into public DefinitionArgument")
		}
	})
}

func TestBuiltinProcessLogemeAppParity(t *testing.T) {
	t.Run("compiler owned effects and canonical receiver", func(t *testing.T) {
		result, _, _ := builtinProcessRequire(t, builtinProcessFilesystem(t), builtinProcessSource)
		plan, err := planOrchestration(result.Compilation.Semantics)
		if err != nil {
			t.Fatal(err)
		}
		want := []orchestrationEffects{orchestrationEffectBuild, orchestrationEffectRun, orchestrationEffectStop, orchestrationEffectAwait}
		if len(plan.Root.Body) != len(want) || len(plan.Apps) != 1 {
			t.Fatalf("unexpected typed lifecycle plan: %+v", plan)
		}
		for i, current := range plan.Root.Body {
			if current.Effects != want[i] || !current.ReachesNext || current.App != plan.Apps[0] || current.Target == nil || current.Target.Type != valueApp || current.Target.App != plan.Apps[0] {
				t.Errorf("operation %d lost compiler-owned effects/type/identity: %+v", i, current)
			}
			if current.Source == nil || current.Source.Source.Path != "builtin-process-contract.l" || current.Source.Source.Start.Line != i+2 {
				t.Errorf("operation %d lost source provenance", i)
			}
		}
	})
	for _, tc := range []struct{ name, source string }{
		{"lifecycle", builtinProcessSource},
		{"two receivers", "Create \"First\" as a New Basic App.\nCreate \"Second\" as a New Basic App.\nRun First.\nRun Second.\nAwait Second.\nAwait First.\n"},
		{"source action order", "Create \"First\" as a New Basic App.\nCreate \"Second\" as a New Basic App.\nExecute First and Second.\n"},
		{"source action alias", "Create \"Worker\" as a New Basic App.\nUse Start as \"Launch\".\nLaunch Worker.\nAwait Worker.\n"},
		{"runtime snapshot", "Create \"Worker\" as a New Basic App.\nRun Worker.\nAwait Worker.\nDefine \"snapshot\" as Worker's Output.\nRun Worker.\nAwait Worker.\nShow snapshot.\n"},
		{"continuation early exit", "    Create \"Worker\" as a New Public Basic App.\nIf true, Exit.\nShow `unreached`.\n\nRun Worker.\nAwait Worker.\nAwait Worker.\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			baseline, baselineUnit, baselineBundle := builtinProcessRequire(t, distribution.Files(), tc.source)
			result, unit, bundle := builtinProcessRequire(t, builtinProcessFilesystem(t), tc.source)
			if !bytes.Equal(baseline.AML, result.AML) || !bytes.Equal(baseline.BackendAML, result.BackendAML) || !bytes.Equal(baselineUnit.BackendAML, unit.BackendAML) || !reflect.DeepEqual(baselineBundle, bundle) {
				t.Fatal("extracting process definitions changed raw App AML or complete Rust bundle")
			}
			before := bytes.Clone(result.BackendAML)
			again, err := result.OrchestratorUnit("entry", executableTestCompanions(result))
			if err != nil || !bytes.Equal(unit.BackendAML, again.BackendAML) || !bytes.Equal(before, result.BackendAML) {
				t.Fatalf("rendering lost determinism or changed application AML: %v", err)
			}
			for _, forbidden := range []string{"Deferred:", "tokio::spawn("} {
				if bytes.Contains(unit.BackendAML, []byte(forbidden)) {
					t.Errorf("lifecycle extraction introduced %s", forbidden)
				}
			}
		})
	}
	t.Run("literal lifecycle glyph sequence", func(t *testing.T) {
		_, unit, _ := builtinProcessRequire(t, distribution.Files(), builtinProcessSource)
		ordered := []string{
			`KY{|Key: ~runtime::app_build~, Inputs: ~&[&lApp0]&~|}`,
			`KY{|Key: ~runtime::app_run~, Inputs: ~&[&lApp0]&~|}`,
			`KY{|Key: ~runtime::app_stop~, Inputs: ~&[&lApp0]&~|}`,
			`KY{|Key: ~runtime::app_wait~, Inputs: ~&[&lApp0]&~|}`,
			`KY{|Key: ~runtime::invocation_claim~, Inputs: ~&[&lInvocation0]&~|}`,
			`BOG^GI^(E{|Statement: ~lProcessClaimed~|}*W^MOB{|Key: ~_~|}^KY{|Key: ~runtime::invocation_finish~, Inputs: ~&[&lInvocation0]&~|})`,
		}
		rest := string(unit.BackendAML)
		for _, expected := range ordered {
			index := strings.Index(rest, expected)
			if index < 0 {
				t.Fatalf("missing independently authored lifecycle composition %s", expected)
			}
			rest = rest[index+len(expected):]
		}
	})
	t.Run("Start and Execute per operand order", func(t *testing.T) {
		for _, action := range []string{"Start", "Execute"} {
			source := "Create \"First\" as a New Basic App.\nCreate \"Second\" as a New Basic App.\n" + action + " First and Second.\n"
			result, _, _ := builtinProcessRequire(t, distribution.Files(), source)
			want := []string{"build", "run", "build", "run"}
			if action == "Execute" {
				want = []string{"build", "run", "await", "build", "run", "await"}
			}
			ops := phase1Lifecycle(t, result, want)
			for i, op := range ops {
				wantTarget := "First"
				if i >= len(ops)/2 {
					wantTarget = "Second"
				}
				if phase1Canonical(t, result, op) != result.Compilation.Memory[wantTarget] {
					t.Fatalf("%s operand order changed at %d", action, i)
				}
				if op.CallSite == nil {
					t.Fatalf("%s expansion lost call-site attribution", action)
				}
			}
		}
	})
	t.Run("source Run is not native redispatch", func(t *testing.T) {
		source := "Create \"Run|-s|-ning\" for Showing any Text.\nRun `ordinary action`.\n"
		result, unit, _ := builtinProcessRequire(t, distribution.Files(), source)
		if got := phase1Kinds(result.Compilation.Semantics.Instructions); !reflect.DeepEqual(got, []string{"show"}) || bytes.Contains(unit.BackendAML, []byte("Key: ~runtime::app_run~")) {
			t.Fatalf("source-defined Run was redispatched as a primitive: %v", got)
		}
	})
	t.Run("native root aliases retain existing rejection", func(t *testing.T) {
		for _, operation := range []string{"Build", "Run", "Stop", "Await"} {
			source := "Create \"Worker\" as a New Basic App.\nUse " + operation + " as \"Action\".\n"
			_, _, _, err := builtinProcessGenerate(distribution.Files(), source)
			if err == nil || !strings.Contains(err.Error(), "alias of native root") || !strings.Contains(err.Error(), operation) {
				t.Errorf("native %s alias changed its current lexical contract: %v", operation, err)
			}
		}
	})
}
