package Lyger

import (
	nativetest "github.com/l-programming-language/l/internal/testsupport/native"

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

const phase1Worker = "\n\"message\" is `child`.\nCreate \"Worker\" as a New Basic App.\nThe App Prints message.\n"

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
		if call.Name == "runtime_run" {
			runs++
		}
	}
	if runs != 1 {
		t.Fatalf("entry point has %d supervised runtime_run calls, want one", runs)
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

func containsPhase1(items []string, s string) bool {
	for _, v := range items {
		if v == s {
			return true
		}
	}
	return false
}
