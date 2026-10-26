package Lyger

import (
	"bytes"
	"encoding/json"
	"fmt"
	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	Logiceme "github.com/l-programming-language/l/logiceme"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func phase1Options(t *testing.T) Options {
	t.Helper()
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	return Options{Language: DefaultLanguage, AssetRoot: root, Resources: os.DirFS(root)}
}
func phase1Evidence(t *testing.T, name string, data []byte) {
	t.Helper()
	path := filepath.Join("..", ".harness", "phase1", "frontend", filepath.FromSlash(t.Name()), name)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
}
func phase1Compile(t *testing.T, source string, imports map[string]string) (*CompilationResult, error) {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "project")
	files := map[string]string{"main.l": source}
	for name, content := range imports {
		files[name] = content
	}
	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for name, content := range files {
		phase1Evidence(t, filepath.Join("inputs", name), []byte(content))
	}
	result, err := NewLygerService(phase1Options(t)).Compile([]byte(source), filepath.Join(dir, "main.l"))
	diagnostic := "frontend passed; runtime not exercised\n"
	if err != nil {
		diagnostic = "frontend prerequisite failed; runtime not exercised: " + err.Error() + "\n"
	}
	phase1Evidence(t, "diagnostic.txt", []byte(diagnostic))
	if result != nil {
		assertFrontendResources(t, result.Compilation)
		phase1Evidence(t, "expanded.l", result.ExpandedSource)
		phase1Evidence(t, "application.aml", result.BackendAML)
	}
	return result, err
}
func phase1Require(t *testing.T, source string, imports map[string]string) *CompilationResult {
	t.Helper()
	result, err := phase1Compile(t, source, imports)
	if err != nil {
		t.Fatalf("frontend prerequisite failed; runtime not exercised: %v", err)
	}
	if result == nil || result.Compilation == nil || result.Compilation.Semantics == nil {
		t.Fatal("frontend returned no semantic program; runtime not exercised")
	}
	return result
}
func phase1Flat(in []*instruction) []*instruction {
	var out []*instruction
	for _, i := range in {
		if i.Kind.String() == "call" {
			out = append(out, phase1Flat(i.Actions)...)
		} else if i.Block != nil {
			out = append(out, phase1Flat(i.Block.Body)...)
		} else {
			out = append(out, i)
			out = append(out, phase1Flat(i.Then)...)
			out = append(out, phase1Flat(i.Else)...)
		}
	}
	return out
}
func phase1Kinds(in []*instruction) []string {
	var out []string
	for _, i := range phase1Flat(in) {
		out = append(out, i.Kind.String())
	}
	return out
}
func phase1Lifecycle(t *testing.T, r *CompilationResult, want []string) []*instruction {
	t.Helper()
	var got []string
	var ops []*instruction
	for _, i := range phase1Flat(r.Compilation.Semantics.Instructions) {
		switch i.Kind.String() {
		case "build", "run", "await", "stop":
			got = append(got, i.Kind.String())
			ops = append(ops, i)
			if i.Target == nil || i.Target.Type.String() != "App" || i.Target.Record == nil {
				t.Fatalf("%s lost typed App operand: %+v", i.Kind, i.Target)
			}
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("lifecycle IR=%v want %v", got, want)
	}
	return ops
}
func phase1Canonical(t *testing.T, r *CompilationResult, i *instruction) *record {
	t.Helper()
	rec, err := canonicalRecord(i.Target.Record, r.Compilation.Semantics.Aliases)
	if err != nil {
		t.Fatal(err)
	}
	return rec
}

const phase1Worker = "Use IO's Print from `io` as \"Print\".\n\"message\" is `child`.\nCreate \"Worker\" as a New Basic App.\nThe App Prints message.\n"

// The executable observations for these language contracts are registered with
// the CLI/runtime suites too. This layer checks binding, private IR, planning,
// and real AML -> logeme output; a frontend failure is not runtime evidence.
func phase1Read(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "testdata", "phase1", path))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func phase1Example(t *testing.T, id int) string {
	t.Helper()
	return phase1Read(t, fmt.Sprintf("language/lang-%02d.l", id))
}

func phase1Diagnostic(t *testing.T, source string, imports map[string]string, all []string, any []string) {
	t.Helper()
	phase1Require(t, phase1Worker+"Run Worker.\nAwait Worker.\n", nil)
	_, err := phase1Compile(t, source, imports)
	if err == nil {
		t.Fatal("accepted invalid Phase 1 source")
	}
	text := strings.ToLower(err.Error())

	for _, bad := range []string{"unimplemented native", "unsupported native role", "unsupported outer instruction action", "unknown instruction", "unknown verb", "unresolved reference \"await\""} {
		if strings.Contains(text, bad) {
			t.Fatalf("wrong failure prerequisite; runtime not exercised: %v", err)
		}
	}
	for _, needle := range all {
		if !strings.Contains(text, strings.ToLower(needle)) {
			t.Fatalf("diagnostic lacks %q; runtime not exercised: %v", needle, err)
		}
	}
	if len(any) > 0 {
		found := false
		for _, needle := range any {
			found = found || strings.Contains(text, strings.ToLower(needle))
		}
		if !found {
			t.Fatalf("wrong diagnostic, need one of %v; runtime not exercised: %v", any, err)
		}
	}
}

// Each emitted main is parsed after the real two-stage compiler path. References
// in testdata are explanatory, never substituted for backend-generated bytes.
func phase1Generated(t *testing.T, r *CompilationResult) (*observedSourceFile, map[string]*observedSourceFile) {
	t.Helper()
	unit, err := r.OrchestratorUnit("entry", executableTestCompanions(r))
	if err != nil {
		t.Fatalf("orchestrator generation: %v; native phase blocked", err)
	}
	service := Logiceme.NewLogicemeService(phase1Options(t).Resources)
	generate := func(id string, aml []byte, runtime bool) *observedSourceFile {
		original := bytes.Clone(aml)
		phase1Evidence(t, filepath.Join("generated", id, "source.aml"), aml)
		bundle, err := service.Evaluate(aml, Logiceme.GenerationOptions{IncludeRuntime: runtime})
		recordBundle(t, bundle)
		if !bytes.Equal(aml, original) {
			t.Error("Rust handoff mutated frontend AML")
		}
		if err != nil {
			t.Fatalf("Rust logeme %s; native phase blocked: %v", id, err)
		}
		for i, path := range bundle.MapArrays[0] {
			phase1Evidence(t, filepath.Join("generated", id, strings.TrimPrefix(path, "/")), bundle.ByteArrays[0][i])
		}
		main := observeBundleMain(t, bundle)
		fn := main.function(t, "main")
		if len(fn.Parameters) != 0 {
			t.Fatal("generated Rust main has parameters")
		}
		if len(main.Source) > 512*1024 {
			t.Fatalf("unbounded continuation expansion: %d bytes", len(main.Source))
		}
		if runtime {
			main.function(t, "l_main")
		}
		if strings.Contains(t.Name(), "/GO-01/") {
			buildActualBundle(t, bundle, id, runtime, "")
		}
		return main
	}
	main := generate(unit.ID, unit.BackendAML, true)
	units, err := r.ApplicationUnits()
	if err != nil {
		t.Fatal(err)
	}
	children := map[string]*observedSourceFile{}
	for _, child := range units {
		if children[child.ID] != nil {
			t.Fatalf("duplicate App source %s", child.ID)
		}
		children[child.ID] = generate(child.ID, child.BackendAML, false)
	}
	if len(children) != len(r.Compilation.Semantics.Apps) {
		t.Fatal("canonical App count changed at Rust handoff")
	}
	return main, children
}

func phase1ReceiverShape(t *testing.T, f *observedSourceFile, wantAwait bool) {
	t.Helper()
	body := f.function(t, "l_main").Body
	for _, call := range nativetest.RustCalls(body) {
		if call.Name == "spawn" || call.Name == "spawn_local" {
			t.Error("paragraph continuations must stay inline in the source task")
		}
		switch call.Name {
		case "lRun", "lAwait", "lStop", "lBuild", "lShow", "lFail", "lOutput", "lError", "lShutdown":
			t.Errorf("unexpected free helper %s remains", call.Name)
		}
	}
	if wantAwait && (phase1Calls(f, "Await") == 0 || phase1Calls(f, "ClaimContinuation") == 0 || phase1Calls(f, "FinishContinuation") == 0) {
		t.Fatal("missing Rust wait/claim_continuation/finish_continuation flow")
	}
	main := f.function(t, "main")
	runs := 0
	for _, call := range nativetest.RustCalls(main.Body) {
		if call.Name == "run" && call.Receiver == "runtime" {
			runs++
		}
	}
	if runs != 1 {
		t.Fatalf("entry point has %d supervised runtime.run calls, want one", runs)
	}
}

func phase1Continuation(t *testing.T, r *CompilationResult, name string) *scopedInstructions {
	t.Helper()
	for app, body := range r.Compilation.Semantics.Continuations {
		if app.ID == name {
			return body
		}
	}
	t.Fatalf("missing %s continuation", name)
	return nil
}

func phase1Marker(t *testing.T, body []*instruction, marker string) {
	t.Helper()
	for _, i := range phase1Flat(body) {
		if i.Value != nil && i.Value.Literal != nil && strings.Contains(i.Value.Literal.Text, marker) {
			return
		}
	}
	t.Fatalf("continuation lost marker %s", marker)
}

func TestPhase1Examples(t *testing.T) {
	for id := 1; id <= 15; id++ {
		t.Run(fmt.Sprintf("LANG-%02d", id), func(t *testing.T) {
			source := phase1Example(t, id)
			if id == 3 {
				phase1Diagnostic(t, source, nil, []string{"main.l", "await", "app"}, []string{"unsupported", "ownership", "body"})
				return
			}
			var imports map[string]string
			if id == 14 {
				imports = map[string]string{"continuation.l": phase1Read(t, "language/continuation.l")}
			}
			r := phase1Require(t, source, imports)
			f, children := phase1Generated(t, r)
			phase1ReceiverShape(t, f, true)
			if phase1Calls(f, "Await") == 0 {
				t.Fatal("source Await disappeared")
			}
			for name, child := range children {
				for _, op := range []string{"Run", "Await", "Stop", "ClaimContinuation", "Shutdown", "Finish"} {
					if phase1Calls(child, op) != 0 {
						t.Errorf("%s child contains orchestrator operation %s", name, op)
					}
				}
			}
			if id == 7 {
				ops := phase1Lifecycle(t, r, []string{"run", "run", "await", "await"})
				want := []string{"First", "Second", "Second", "First"}
				for n, op := range ops {
					if phase1Canonical(t, r, op).ID != want[n] {
						t.Errorf("operand %d = %s want %s", n, op.Target.ID, want[n])
					}
				}
			}
			if id == 9 && phase1Calls(f, "Await") < 2 {
				t.Fatal("self-Await omitted by generation recursion suppression")
			}
			if id == 11 {
				phase1AssertStoredOutput(t, r)
			}
		})
	}
}

func TestPhase1Language(t *testing.T) {
	t.Parallel()
	ids := []string{"LEX-01", "LEX-02", "LEX-03", "LEX-04", "LEX-05", "LEX-06", "LEX-07", "LEX-08", "LEX-09", "IR-01", "IR-02", "IR-03", "IR-04", "IR-05", "IR-06", "IR-07", "IR-08", "PAR-01", "PAR-02", "PAR-03", "PAR-04", "PAR-05", "PAR-06", "PAR-07", "PAR-08", "PAR-09", "PAR-10", "PAR-11", "PAR-12", "PAR-13", "PAR-14", "PAR-15", "PAR-16", "PAR-17", "PAR-18", "PAR-19", "PAR-20", "FLOW-01", "FLOW-02", "FLOW-03", "FLOW-04", "FLOW-05", "FLOW-06", "FLOW-07", "FLOW-08", "DIAG-01", "DIAG-02", "DIAG-03", "DIAG-04", "DIAG-05", "GO-01", "GO-02", "GO-03", "GO-04", "GO-05"}
	for _, id := range ids {
		t.Run(id, func(t *testing.T) { phase1LanguageCase(t, id) })
	}
}

func phase1LanguageCase(t *testing.T, id string) {
	switch id {
	case "LEX-01":
		phase1Morphology(t, "await", "suspensive", map[string]string{"await": "root", "awaits": "active", "awaited": "passive", "awaiting": "applicative"})
		prepared, err := PrepareSource("morphology.l", []byte("AwAiT Worker.\n"))
		if err != nil {
			t.Fatal(err)
		}
		if len(prepared.Statements) != 1 || prepared.Statements[0].Span.Start.Line != 1 || prepared.Statements[0].Span.Start.Column != 1 {
			t.Fatal("Await source span lost")
		}
	case "LEX-02":
		phase1Morphology(t, "stop", "interruptive", map[string]string{"stop": "root", "stops": "active", "stopped": "passive", "stopping": "applicative"})
	case "LEX-03":
		r, err := phase1Compile(t, phase1Worker+"Use Await as \"Join\".\nRun Worker.\nJoin Worker.\n", nil)
		if err != nil {
			s := strings.ToLower(err.Error())
			if !strings.Contains(s, "alias") || (!strings.Contains(s, "native") && !strings.Contains(s, "root")) {
				t.Fatalf("root alias rejection lacks capability reason: %v", err)
			}
			return
		}
		phase1Lifecycle(t, r, []string{"run", "await"})
	case "LEX-04":
		phase1CanonicalAssets(t)
	case "LEX-05":
		for _, s := range []string{"Await", "Await Worker extra"} {
			if _, err := phase1ParseGrammar(t, s, "verbs.suspensive.root"); err == nil {
				t.Fatalf("accepted malformed %s", s)
			}
		}
	case "LEX-06":
		for _, operand := range []string{"`text`", "3", "true"} {
			t.Run(operand, func(t *testing.T) {
				phase1Diagnostic(t, "Await "+operand+".\n", nil, []string{"main.l", "app"}, []string{"type", "requires", "expected"})
			})
		}
	case "LEX-07":
		passive, err := phase1ParseGrammar(t, "Worker is awaited", "verbs.passive.root")
		if err != nil {
			t.Fatal(err)
		}
		if passive.subject == nil || string(passive.subject.noun.rawBytes) != "Worker" {
			t.Fatal("passive lost subject target")
		}
		active, err := phase1ParseGrammar(t, "Worker awaits Other", "verbs.suspensive.active")
		if err != nil {
			t.Fatal(err)
		}
		if active.subject == nil || string(active.subject.noun.rawBytes) != "Worker" || string(active.predicate.clauses[0].obj.noun.rawBytes) != "Other" {
			t.Fatal("active swapped subject and object")
		}
		r := phase1Require(t, phase1Example(t, 2), nil)
		ops := phase1Lifecycle(t, r, []string{"run", "await"})
		if phase1Canonical(t, r, ops[0]) != phase1Canonical(t, r, ops[1]) {
			t.Fatal("passive Await changed identity")
		}
	case "LEX-08":
		if _, err := phase1ParseGrammar(t, "Awaiting Worker", "verbs.suspensive.applicative"); err == nil {
			t.Fatal("standalone applicative accepted")
		}
	case "LEX-09":
		phase1Diagnostic(t, phase1Worker+"Publicly Await Worker.\n", nil, []string{"main.l"}, []string{"modifier", "adverb", "public"})
	case "IR-01":
		r := phase1Require(t, phase1Worker+"Run Worker.\nStop Worker.\nAwait Worker.\n", nil)
		phase1Lifecycle(t, r, []string{"run", "stop", "await"})
	case "IR-02":
		r := phase1Require(t, phase1Worker+"Start Worker.\nShow `after-start|`.\nAwait Worker.\n", nil)
		phase1Lifecycle(t, r, []string{"build", "run", "await"})
		calls := 0
		for _, i := range r.Compilation.Semantics.Instructions {
			if i.Kind.String() == "call" {
				calls++
				if !i.Synchronous {
					t.Fatal("source action expansion lost ordered execution")
				}
			}
		}
		if calls == 0 {
			t.Fatal("Start bypassed source action expansion")
		}
		f, _ := phase1Generated(t, r)
		phase1ReceiverShape(t, f, true)
	case "IR-03":
		s := phase1Worker + "Create \"Collect|-s|-ed|-ing|-or\" for Starting any App and then Awaiting it.\nUse Collect as \"Join\".\nUse Join as \"Gather\".\nGather Worker.\n"
		r := phase1Require(t, s, nil)
		ops := phase1Lifecycle(t, r, []string{"build", "run", "await"})
		for index, i := range ops {
			definitionFile, definitionLine := "core.l", 1
			if index == 2 {
				definitionFile, definitionLine = "main.l", 5
			}
			if i.CallSite == nil || i.CallSite.source == nil || filepath.Base(i.CallSite.source.Span.Path) != "main.l" || i.CallSite.source.Span.Start.Line != 8 || filepath.Base(i.Source.Path) != definitionFile || i.Source.Start.Line != definitionLine {
				t.Fatalf("nested alias expansion lost original definition and ultimate call attribution: %+v", i)
			}
		}
		calls := r.Compilation.Semantics.Instructions
		if len(calls) == 0 || calls[len(calls)-1].Kind.String() != "call" || calls[len(calls)-1].Definition == nil {
			t.Fatal("renamed action was spelling-special-cased")
		}
	case "IR-04":
		s := "Create \"Collect|-s|-ed|-ing|-or\" for Awaiting any App and then Showing any Text.\n" + phase1Worker + "Run Worker.\nCollect Worker and message.\n"
		r := phase1Require(t, s, nil)
		phase1Lifecycle(t, r, []string{"run", "await"})
		if got := phase1Kinds(r.Compilation.Semantics.Instructions); !reflect.DeepEqual(got, []string{"run", "await", "show"}) {
			t.Fatalf("typed two-parameter expansion=%v", got)
		}
		phase1Diagnostic(t, strings.Replace(s, "Collect Worker and message.", "Collect Worker.", 1), nil, []string{"main.l"}, []string{"argument", "arity", "requires", "parameter"})
	case "IR-05":
		r := phase1Require(t, phase1Example(t, 6), nil)
		ops := phase1Lifecycle(t, r, []string{"build", "run", "await", "await"})
		first := phase1Canonical(t, r, ops[0])
		for _, op := range ops {
			if phase1Canonical(t, r, op) != first {
				t.Fatal("alias/pronoun created second App identity")
			}
		}
		if len(r.Compilation.Semantics.Apps) != 1 {
			t.Fatal("alias duplicated App artifact")
		}
	case "IR-06":
		s := "Use IO's Print from `io` as \"Print\".\n\"message\" is `private`.\n\n    Create \"Worker\" as a New Basic App.\nThe App Prints message.\nShow `first|`.\nRun Worker.\nAwait Worker.\n\n    Create \"Worker\" as a New Basic App.\nThe App Prints message.\nShow `second|`.\nRun Worker.\nAwait Worker.\n"
		r := phase1Require(t, s, nil)
		if len(r.Compilation.Semantics.Apps) != 2 || r.Compilation.Semantics.Apps[0] == r.Compilation.Semantics.Apps[1] {
			t.Fatal("private same-spelling Apps conflated")
		}
		_, children := phase1Generated(t, r)
		if len(children) != 2 {
			t.Fatal("private Apps share generated process slot")
		}
	case "IR-07":
		s := strings.Replace(phase1Example(t, 6), "New Basic App", "New Deferred App", 1) + "\n\"Deferred\" describes a Generic App with a New Main Service using any Process.\n"
		r := phase1Require(t, s, nil)
		ops := phase1Lifecycle(t, r, []string{"build", "run", "await", "await"})
		for _, op := range ops {
			if phase1Canonical(t, r, op).ID != "Worker" {
				t.Fatal("resolution retry changed pronoun/canonical order")
			}
		}
	case "IR-08":
		r := phase1Require(t, phase1Worker+"Run Worker.\nUse `middle.l`.\nShow `tail|`.\n", map[string]string{"middle.l": "Await Worker.\n"})
		ops := phase1Lifecycle(t, r, []string{"run", "await"})
		if filepath.Base(ops[1].Source.Path) != "middle.l" || ops[1].Source.Start.Line != 1 {
			t.Fatal("imported Await lost original span")
		}
		if got := phase1Kinds(r.Compilation.Semantics.Instructions); !reflect.DeepEqual(got, []string{"run", "await", "show"}) {
			t.Fatalf("import effect order=%v", got)
		}
	case "PAR-01", "PAR-02", "PAR-03", "PAR-04", "PAR-05", "PAR-06", "PAR-07", "PAR-08", "PAR-09":
		phase1ActivationCase(t, id)
	case "PAR-10":
		source := strings.Replace(phase1Example(t, 15), "Exit.", "Create \"Leave|-s|-d|--ing\" for Exiting with an error of any Text.\nLeave `joined`.\n", 1)
		r := phase1Require(t, source, nil)
		body := phase1Continuation(t, r, "First")
		if !containsPhase1(phase1Kinds(body.Body), "exit") {
			t.Fatal("nested source Exit lost continuation ownership")
		}
		f, _ := phase1Generated(t, r)
		phase1ReceiverShape(t, f, true)
	case "PAR-11", "PAR-12", "PAR-13":
		phase1OwnershipCase(t, id)
	case "PAR-14":
		source := strings.Replace(phase1Example(t, 8), "Await Other.", "Stop Other.\nAwait Other.", 1)
		r := phase1Require(t, source, nil)
		worker := phase1Continuation(t, r, "Worker")
		other := phase1Continuation(t, r, "Other")
		phase1Marker(t, worker.Body, "worker-after-join|")
		phase1Marker(t, other.Body, "other-continuation|")
		if !containsPhase1(phase1Kinds(worker.Body), "stop") || !containsPhase1(phase1Kinds(worker.Body), "await") {
			t.Fatal("callee stole caller lifecycle operations")
		}
	case "PAR-15":
		s := phase1Worker + "\n    \"marker\" is `ordinary|`.\nRun Worker.\nAwait Worker.\nShow marker.\n"
		r := phase1Require(t, s, nil)
		phase1Lifecycle(t, r, []string{"run", "await"})
		if len(r.Compilation.Semantics.Continuations) != 0 {
			t.Fatal("ordinary paragraph became an App continuation")
		}
	case "PAR-16":
		t.Run("live-private-frame", func(t *testing.T) {
			source := strings.Replace(phase1Example(t, 12), "Show result.\n\nExecute Worker.", "Create \"Reveal|-s|-ed|-ing|-er\" for Showing result and then Showing any Text.\nReveal `tail|`.\n\nExecute Worker.", 1)
			r := phase1Require(t, source, nil)
			body := phase1Continuation(t, r, "Worker")
			var local *record
			for _, i := range phase1Flat(body.Body) {
				if i.Kind.String() == "assign" {
					local = i.Target.Record
				}
			}
			captured := false
			for _, i := range phase1Flat(body.Body) {
				if i.Kind.String() == "show" && i.Value != nil && i.Value.Record == local {
					captured = true
				}
			}
			if local == nil || !captured {
				t.Fatal("private action lost definition-site runtime capture")
			}
			f, _ := phase1Generated(t, r)
			phase1ReceiverShape(t, f, true)
		})
		t.Run("public-runtime-capture-escape", func(t *testing.T) {
			source := "Use IO's Print from `io` as \"Print\".\n\"message\" is `private`.\n\n    Create \"Worker\" as a New Public Basic App.\nThe App Prints message.\nDefine \"privateText\" as Worker's Output.\nPublicly create \"Report|-s|-ed|-ing\" for Showing privateText and then Showing any Text.\n\nRun Worker.\nAwait Worker.\nReport `caller`.\n"
			r := phase1Require(t, source, nil)
			plan, err := planOrchestration(r.Compilation.Semantics)
			if err != nil {
				t.Fatal(err)
			}
			var binding, call *orchestrationInstruction
			visitCapturePlan(plan.Root.Body, func(current *orchestrationInstruction) {
				if current.Source.Callable != r.Compilation.Memory["Report"] {
					return
				}
				switch current.Source.Kind {
				case instructionBindAction:
					binding = current
				case instructionCall:
					call = current
				}
			})
			if binding == nil || len(binding.CaptureValues) != 1 || binding.CaptureValues[0].Access != orchestrationAccessSnapshot || binding.Source.Source.Start.Line != 7 {
				t.Fatal("published capture lost its snapshot or definition source")
			}
			if call == nil || call.Source.Source.Start.Line != 11 || call.CaptureSource == nil || call.CaptureSource.Slot != binding.Callable || len(call.Actions) != 2 {
				t.Fatal("escaped call lost its bound environment, source or ordered actions")
			}
			value := call.Actions[0].Value
			if value.Storage == nil || value.Storage.Capture == nil || value.Storage.Record != binding.CaptureValues[0].Record || value.Storage.Record.ID != "privateText" || r.Compilation.Memory["privateText"] != nil {
				t.Fatal("retained capture rebound or published the private value")
			}
			if _, err := r.OrchestratorUnit("entry", executableTestCompanions(r)); err != nil {
				t.Fatalf("retained public capture generation: %v", err)
			}
		})
	case "PAR-17":
		for _, prefix := range []string{"", "    "} {
			t.Run(fmt.Sprintf("app-parent-indent-%d", len(prefix)), func(t *testing.T) {
				r := phase1Require(t, phase1Example(t, 14), map[string]string{"continuation.l": prefix + phase1Read(t, "language/continuation.l")})
				body := phase1Continuation(t, r, "Worker")
				found := false
				for _, i := range phase1Flat(body.Body) {
					if filepath.Base(i.Source.Path) == "continuation.l" {
						found = true
					}
				}
				if !found {
					t.Fatal("imported continuation lost activation context")
				}
			})
		}
		for _, prefix := range []string{"", "    "} {
			t.Run(fmt.Sprintf("ordinary-parent-indent-%d", len(prefix)), func(t *testing.T) {
				source := phase1Worker + "Run Worker.\nAwait Worker.\nDefine \"snapshot\" as Worker's Output.\nUse `continuation.l`.\nShow snapshot.\n"
				r := phase1Require(t, source, map[string]string{"continuation.l": prefix + phase1Read(t, "language/continuation.l")})
				if len(r.Compilation.Semantics.Continuations) != 0 {
					t.Fatal("ordinary parent import became App continuation")
				}
				snapshot := r.Compilation.Memory["snapshot"]
				found := false
				for _, i := range phase1Flat(r.Compilation.Semantics.Instructions) {
					if filepath.Base(i.Source.Path) == "continuation.l" && i.Kind.String() == "show" && i.Value.Record == snapshot {
						found = true
					}
				}
				if snapshot == nil || !found {
					t.Fatal("scoped import lost its live parent snapshot binding")
				}
				if prefix != "" && r.Compilation.Memory["marker"] != nil {
					t.Fatal("scoped import leaked its private marker")
				}
			})
		}
	case "PAR-18":
		source := strings.Replace(phase1Example(t, 6), "Start it.", "*** comment-only line ***\nStart it.\nShow `literal\n\nblank|`.", 1)
		r := phase1Require(t, source, nil)
		ops := phase1Lifecycle(t, r, []string{"build", "run", "await", "await"})
		if phase1Canonical(t, r, ops[1]).ID != "Worker" {
			t.Fatal("comment/literal rebound paragraph pronoun")
		}
		if !bytes.Contains(r.ExpandedSource, []byte("literal\n\nblank|")) {
			t.Fatal("literal blank bytes changed")
		}
	case "PAR-19":
		phase1Diagnostic(t, phase1Worker+" \t \nAwait it.\n", nil, []string{"main.l"}, []string{"pronoun", "subject", "unresolved"})
	case "PAR-20":
		r := phase1Require(t, phase1Example(t, 9), nil)
		f, _ := phase1Generated(t, r)
		phase1ReceiverShape(t, f, true)
		if phase1Calls(f, "Await") < 2 {
			t.Fatal("self-Await suppressed rather than cached invocation observation")
		}
		mutual := "Use IO's Print from `io` as \"Print\".\n\"message\" is `mutual`.\n\n    Create \"First\" as a New Public Basic App.\nThe App Prints message.\nAwait Second.\nShow `first|`.\n\n    Create \"Second\" as a New Public Basic App.\nThe App Prints message.\nAwait First.\nShow `second|`.\n\nRun First.\nRun Second.\nAwait First.\n"
		rr := phase1Require(t, mutual, nil)
		ff, _ := phase1Generated(t, rr)
		if phase1Calls(ff, "Await") < 3 {
			t.Fatal("mutual Await lifecycle edge omitted")
		}
	case "FLOW-01":
		phase1AssertStoredOutput(t, phase1Require(t, phase1Example(t, 11), nil))
	case "FLOW-02", "FLOW-03", "FLOW-04":
		phase1FlowCase(t, id)
	case "FLOW-05":
		r := phase1Require(t, phase1Worker+"Run Worker.\nDefine \"unused\" as Worker's Output.\nAwait Worker.\n", nil)
		ops := phase1Flat(r.Compilation.Semantics.Instructions)
		if len(ops) != 3 || ops[1].Kind.String() != "assign" || ops[1].Value == nil || ops[1].Value.Member.String() != "output" {
			t.Fatal("unused output RHS lost ordered readiness effect")
		}
		f, _ := phase1Generated(t, r)
		if phase1Calls(f, "Output") != 1 {
			t.Fatal("unused output RHS was eliminated")
		}
	case "FLOW-06":
		for _, name := range []string{"f06-nonzero-exit-preserves-streams", "f21-raw-output-and-stderr-success"} {
			t.Run(name, func(t *testing.T) { phase1FixtureFrontend(t, name) })
		}
	case "FLOW-07":
		source := strings.Replace(phase1Example(t, 7), "Show `first-continuation|`.", "Define \"result\" as First's Output.\nShow result.", 1)
		source = strings.Replace(source, "Show `second-continuation|`.", "Define \"result\" as Second's Output.\nShow result.", 1)
		r := phase1Require(t, source, nil)
		var records []*record
		for _, name := range []string{"First", "Second"} {
			for _, i := range phase1Flat(phase1Continuation(t, r, name).Body) {
				if i.Kind.String() == "assign" {
					records = append(records, i.Target.Record)
				}
			}
		}
		if len(records) != 2 || records[0] == records[1] {
			t.Fatal("independent continuations share private result storage")
		}
		f, _ := phase1Generated(t, r)
		phase1ReceiverShape(t, f, true)
	case "FLOW-08":
		source := strings.Replace(phase1Example(t, 8), "Run Other.", "Define \"snapshot\" as Worker's Output.\nRun Other.", 1)
		source = strings.Replace(source, "Show `worker-after-join|`.", "Show snapshot.\nShow `worker-after-join|`.", 1)
		r := phase1Require(t, source, nil)
		body := phase1Continuation(t, r, "Worker")
		var record *record
		for _, i := range phase1Flat(body.Body) {
			if i.Kind.String() == "assign" {
				record = i.Target.Record
			}
			if i.Kind.String() == "show" && i.Value != nil && i.Value.Record != nil && i.Value.Record.ID == "snapshot" && i.Value.Record != record {
				t.Fatal("nested Await changed caller snapshot binding")
			}
		}
		if record == nil {
			t.Fatal("missing caller activation snapshot")
		}
		f, _ := phase1Generated(t, r)
		phase1ReceiverShape(t, f, true)
	case "DIAG-01":
		phase1Diagnostic(t, phase1Worker+"Run Worker.\nUse `nested/failure.l`.\n", map[string]string{"nested/failure.l": "\nAwait Missing.\n"}, []string{"failure.l", "2", "missing"}, []string{"unresolved", "unknown", "reference"})
	case "DIAG-02":
		phase1Diagnostic(t, "Create \"Collect|-s|-ed|-ing\" for Awaiting any App.\nUse Collect as \"Join\".\nJoin `wrong`.\n", nil, []string{"main.l:3:", "app", "collect"}, []string{"type", "argument", "requires"})
	case "DIAG-03":
		r := phase1Require(t, phase1Worker+"Await Worker.\n", nil)
		ops := phase1Lifecycle(t, r, []string{"await"})
		if ops[0].Target.Record.ID != "Worker" || filepath.Base(ops[0].Source.Path) != "main.l" || ops[0].Source.Start.Line != 5 {
			t.Fatal("runtime invocation lost canonical identity/source span")
		}
	case "DIAG-04":
		phase1Diagnostic(t, phase1Example(t, 3), nil, []string{"main.l", "await", "app"}, []string{"unsupported", "ownership", "body"})
	case "DIAG-05":
		r := phase1Require(t, "Show `no-app|`.\n", nil)
		if len(r.Compilation.Semantics.Apps) != 0 {
			t.Fatal("zero-App source invented an App")
		}
		f, children := phase1Generated(t, r)
		if len(children) != 0 {
			t.Fatal("zero-App program got a child artifact")
		}
		phase1ReceiverShape(t, f, false)
	case "GO-01":
		for n := 1; n <= 15; n++ {
			if n == 3 {
				continue
			}
			t.Run(fmt.Sprintf("LANG-%02d", n), func(t *testing.T) {
				var dependencies map[string]string
				if n == 14 {
					dependencies = map[string]string{"continuation.l": phase1Read(t, "language/continuation.l")}
				}
				phase1Generated(t, phase1Require(t, phase1Example(t, n), dependencies))
			})
		}
	case "GO-02":
		r := phase1Require(t, phase1Example(t, 8), nil)
		f, _ := phase1Generated(t, r)
		phase1ReceiverShape(t, f, true)
		if phase1Calls(f, "ClaimContinuation") < 2 {
			t.Fatal("nested invocation claim omitted")
		}
	case "GO-03":
		r := phase1Require(t, phase1Example(t, 1), nil)
		_, children := phase1Generated(t, r)
		for name, f := range children {
			if phase1Calls(f, "Println") != 1 {
				t.Fatalf("%s child body lost Println", name)
			}
			for _, token := range f.Tokens {
				if text, err := nativetest.DecodeRustLiteral(token.Text); err == nil && strings.Contains(text, "continued|") {
					t.Error("orchestrator continuation leaked into child")
				}
			}
		}
	case "GO-04":
		r := phase1Require(t, phase1Example(t, 4), nil)
		phase1Lifecycle(t, r, []string{"run", "await"})
		f, _ := phase1Generated(t, r)
		phase1ReceiverShape(t, f, true)
	case "GO-05":
		r := phase1Require(t, phase1Example(t, 6), nil)
		units, err := r.ApplicationUnits()
		if err != nil {
			t.Fatal(err)
		}
		if len(units) != 1 {
			t.Fatal("Await/alias duplicated canonical child source bundle")
		}
		before := bytes.Clone(r.BackendAML)
		a, err := r.OrchestratorUnit("entry", executableTestCompanions(r))
		if err != nil {
			t.Fatal(err)
		}
		b, err := r.OrchestratorUnit("entry", executableTestCompanions(r))
		if err != nil || !bytes.Equal(a.BackendAML, b.BackendAML) || !bytes.Equal(before, r.BackendAML) {
			t.Fatal("orchestration rendering mutated or destabilized canonical source mapping")
		}
		phase1Generated(t, r)
	default:
		t.Fatalf("unregistered language case %s", id)
	}
}

func containsPhase1(items []string, s string) bool {
	for _, v := range items {
		if v == s {
			return true
		}
	}
	return false
}

func phase1ActivationCase(t *testing.T, id string) {
	t.Helper()
	source := phase1Example(t, 1)
	tail := map[string]string{"PAR-01": "Show `root|`.\n", "PAR-02": "Build Worker.\n", "PAR-03": "Run Worker.\nShow `root|`.\n", "PAR-04": "Run Worker.\nAwait Worker.\n", "PAR-05": "Use Worker as \"Job\".\nRun Worker.\nAwait Worker.\nAwait Job.\n", "PAR-06": "Run Worker.\nAwait Worker.\nRun Worker.\nAwait Worker.\n", "PAR-07": "Run Worker.\nStop Worker.\nAwait Worker.\n", "PAR-08": "Run Worker.\nAwait Worker.\nAwait Worker.\n", "PAR-09": "Await Worker.\n"}[id]
	source = source[:strings.Index(source, "\nRun Worker.\n")] + "\n" + tail
	r := phase1Require(t, source, nil)
	body := phase1Continuation(t, r, "Worker")
	phase1Marker(t, body.Body, "continued|")
	f, _ := phase1Generated(t, r)
	if id == "PAR-01" || id == "PAR-02" || id == "PAR-03" {
		if phase1Calls(f, "ClaimContinuation") != 0 {
			t.Fatal("continuation activated without Await")
		}
		for _, token := range f.Tokens {
			if text, err := nativetest.DecodeRustLiteral(token.Text); err == nil && strings.Contains(text, "continued|") {
				t.Error("user continuation was emitted before any Await")
			}
		}
		phase1ReceiverShape(t, f, false)
		return
	}
	phase1ReceiverShape(t, f, true)
	if id == "PAR-05" || id == "PAR-06" || id == "PAR-08" {
		if phase1Calls(f, "Await") != 2 || phase1Calls(f, "ClaimContinuation") != 2 {
			t.Fatal("each lexical Await needs invocation claim, including cached/failure joins")
		}
	}
}

func phase1OwnershipCase(t *testing.T, id string) {
	t.Helper()
	s := "Use IO's Print from `io` as \"Print\".\n\"message\" is `shared`.\n\n    Create \"First\" as a New Public Basic App.\nThe App Prints message.\nShow `first|`.\nCreate \"Second\" as a New Public Basic App.\nThe App Prints message.\nShow `second|`.\n"
	if id == "PAR-12" {
		s += "Show First's Output.\n"
	}
	if id == "PAR-13" {
		s += "If Second has an Error, exit with the Error, else show First's Output.\n"
	}
	s += "\nRun First.\nRun Second.\nAwait First.\nAwait Second.\n"
	r := phase1Require(t, s, nil)
	first := phase1Continuation(t, r, "First")
	second := phase1Continuation(t, r, "Second")
	phase1Marker(t, first.Body, "first|")
	phase1Marker(t, second.Body, "second|")
	if first == second || first.Scope != second.Scope {
		t.Fatal("shared lexical scope lost distinct App continuations")
	}
	if id == "PAR-12" {
		found := false
		for _, i := range phase1Flat(first.Body) {
			found = found || (i.Value != nil && i.Value.Owner != nil && i.Value.Owner.Record.ID == "First")
		}
		if !found {
			t.Fatal("explicit RHS owner reassigned to latest App")
		}
	}
	if id == "PAR-13" {
		if !containsPhase1(phase1Kinds(second.Body), "if") {
			t.Fatal("condition owner changed to RHS App")
		}
	}
	f, _ := phase1Generated(t, r)
	phase1ReceiverShape(t, f, true)
	if phase1Calls(f, "ClaimContinuation") < 2 {
		t.Fatal("shared paragraph conflated invocation claims")
	}
}

func phase1AssertStoredOutput(t *testing.T, r *CompilationResult) {
	t.Helper()
	items := phase1Flat(r.Compilation.Semantics.Instructions)
	var assigned *record
	var value *instructionValue
	for _, i := range items {
		if i.Kind.String() == "assign" && i.Target.Record.ID == "snapshot" {
			assigned = i.Target.Record
			value = i.Value
		}
		if i.Kind.String() == "show" && i.Value != nil && i.Value.Record != nil && i.Value.Record.ID == "snapshot" {
			if i.Value.Record != assigned || i.Value.Kind.String() != "symbol" {
				t.Fatal("saved output became a new member read")
			}
		}
	}
	if assigned == nil || value == nil || value.Member.String() != "output" {
		t.Fatal("missing stable runtime snapshot assignment")
	}
	f, _ := phase1Generated(t, r)
	if phase1Calls(f, "Output") != 2 {
		t.Fatal("stored Output was reevaluated or direct latest Output disappeared")
	}
}

func phase1FlowCase(t *testing.T, id string) {
	t.Helper()
	prefix := "Use IO's Print from `io` as \"Print\".\n\"message\" is `published`.\n\"enabled\" is true.\n\n    Create \"Worker\" as a New Public Basic App.\nThe App Prints message.\nPublicly define \"result\" as Worker's Output.\n\nRun Worker.\n"
	tails := []string{}
	switch id {
	case "FLOW-02":
		tails = []string{"If enabled equals true, await Worker, else await Worker.\nShow result.\n", "Await Worker.\nIf enabled equals true, await Worker, else show `cached|`.\nShow result.\n"}
	case "FLOW-03":
		tails = []string{"If enabled equals true, await Worker.\nShow result.\n"}
	case "FLOW-04":
		tails = []string{"If enabled equals true, exit, else await Worker.\nShow result.\n"}
	}
	for index, tail := range tails {
		t.Run(fmt.Sprintf("join-path-%d", index+1), func(t *testing.T) {
			r := phase1Require(t, prefix+tail, nil)
			body := phase1Continuation(t, r, "Worker")
			var assignment *instruction
			for _, i := range phase1Flat(body.Body) {
				if i.Kind.String() == "assign" && i.Target.Record.ID == "result" {
					assignment = i
				}
			}
			if assignment == nil || assignment.Target.Record != r.Compilation.Memory["result"] {
				t.Fatal("published runtime assignment lost continuation ownership or canonical root record")
			}
			_, err := r.OrchestratorUnit("entry", executableTestCompanions(r))
			if id == "FLOW-03" {
				if err == nil || !strings.Contains(strings.ToLower(err.Error()), "assign") {
					t.Fatalf("expected continuation definite-assignment error: %v", err)
				}
			} else if err != nil {
				t.Fatalf("continuing join paths rejected: %v", err)
			}
		})
	}
	if id == "FLOW-03" {
		t.Run("partial-assignment-inside-claimed-continuation", func(t *testing.T) {
			r := phase1Require(t, prefix+"Await Worker.\nShow result.\n", nil)
			body := phase1Continuation(t, r, "Worker")

			condition := &instructionCondition{Operator: "truth", Left: &instructionValue{Kind: operandLiteral, Type: valueTruth, Literal: &primitive{Type: valueTruth, Truth: true}}}
			body.Body = []*instruction{{Kind: instructionIf, Condition: condition, Then: body.Body, Source: body.Body[0].Source, Statement: body.Body[0].Statement}}
			_, err := r.OrchestratorUnit("entry", executableTestCompanions(r))
			if err == nil || !strings.Contains(strings.ToLower(err.Error()), "assign") {
				t.Fatalf("claimed-but-uninitialized continuation accepted: %v", err)
			}
		})
	}
}

func phase1FixtureFrontend(t *testing.T, name string) {
	t.Helper()
	dir := filepath.Join("..", "testdata", "phase1")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	actual := ""
	for _, entry := range entries {
		if strings.EqualFold(entry.Name(), name) {
			actual = entry.Name()
		}
	}
	if actual == "" {
		t.Fatalf("missing fixture %s", name)
	}
	main := phase1Read(t, actual+"/main.l")
	var manifest struct {
		Files []string `json:"files"`
	}
	if err := json.Unmarshal([]byte(phase1Read(t, actual+"/case.json")), &manifest); err != nil {
		t.Fatal(err)
	}
	imports := make(map[string]string)
	for _, name := range manifest.Files {
		if name != "main.l" && strings.HasSuffix(name, ".l") {
			imports[name] = phase1Read(t, filepath.Join(actual, name))
		}
	}
	r := phase1Require(t, main, imports)
	f, _ := phase1Generated(t, r)
	if phase1Calls(f, "Output") == 0 {
		t.Fatal("result bytes never read")
	}
}
