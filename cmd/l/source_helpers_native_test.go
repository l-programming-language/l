//go:build darwin || linux

package main

import (
	"crypto/sha256"
	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"

	Logiceme "github.com/l-programming-language/l/logiceme"
	Lyger "github.com/l-programming-language/l/lyger"
)

// Source roots own separate compilers and projects, but the retained stdout
// observer temporarily changes os.Stdout. Protect that process-global boundary
// while allowing native builds and source assertions to proceed independently.
var sourceStdoutMu sync.Mutex

var sourceEvidenceMu sync.Mutex

func sourceStdout(t *testing.T, action func()) string {
	t.Helper()
	sourceStdoutMu.Lock()
	defer sourceStdoutMu.Unlock()
	return lifecycleStdout(t, action)
}

func compilerTestCLI() *cliData { return &cliData{"build", "./test/main.l", "ubuntu"} }

func expectedCompilerAML(t testing.TB) []byte {
	t.Helper()
	b, e := os.ReadFile("test/main.aml.txt")
	if e != nil {
		t.Fatal(e)
	}
	return b
}

func assertFrontendResources(t testing.TB, c *Lyger.Compilation) {
	t.Helper()
	if c == nil {
		t.Fatal("missing compilation")
	}
	for _, u := range c.SourceUnits {
		if u.Kind == Lyger.SourceBuiltin && !strings.HasSuffix(u.Origin.Name, "builtins/eng/core.l") {
			t.Errorf("Rust builtin selection: %q", u.Origin.Name)
		}
		if u.Kind == Lyger.SourceDictionary && u.Dictionary == "IO" && !strings.HasSuffix(u.Origin.Name, "stdlib/eng/io/dictionary.l") {
			t.Errorf("Rust IO selection: %q", u.Origin.Name)
		}
	}
	if c.Program != nil {
		for _, a := range c.Program.Apps {
			if a.Meta["AppLang"] != "rust" {
				t.Errorf("Rust application %s target=%q", a.ID, a.Meta["AppLang"])
			}
		}
	}
}

func assertMemberSourceBuild(t *testing.T, bundle Logiceme.Executable, text string) {
	t.Helper()
	want := [][]string{{"/src/main.rs", "/Cargo.toml", "/src/config.rs", "/start.sh", "/stop.sh"}}
	if !reflect.DeepEqual(bundle.MapArrays, want) || len(bundle.ByteArrays) != 1 || len(bundle.ByteArrays[0]) != len(want[0]) || len(bundle.DependencyArrays) != 1 || len(bundle.EnvDepArrays) != 1 {
		t.Fatalf("complete Rust source/path/dependency contract: %v", bundle.MapArrays)
	}
	source := observeBundleSources(t, bundle)
	main := source["/src/main.rs"].Function(t, "main")
	if len(main.Parameters) != 0 || (len(main.Result) != 0 && !observedUnitOutput(main.Result)) {
		t.Fatal("main must be a parameterless unit function")
	}
	config := source["/src/config.rs"]
	if config == nil {
		t.Fatal("configuration source missing")
	}
	config.Function(t, "initialize_configuration_settings")
	calls := observeCalls(main.Body)
	var initializers, printers []nativetest.RustCall
	for _, call := range calls {
		if call.Name == "initialize_configuration_settings" {
			initializers = append(initializers, call)
		}
		if call.Name == "l_print" {
			printers = append(printers, call)
		}
	}
	// Rust propagates configuration failure through an explicit guarded error
	// path. Its Err pattern and diagnostic calls are not extra initializations
	// or user prints; retain the semantic once-and-before obligation.
	if len(initializers) != 1 || len(initializers[0].Args) != 0 || len(printers) != 1 || initializers[0].At >= printers[0].At {
		t.Fatalf("main must initialize configuration then print exactly once: %#v", calls)
	}
	if got := observedPrinted(t, main.Body); !reflect.DeepEqual(got, []string{text}) {
		t.Fatalf("source text=%q want%q", got, text)
	}
	for i, p := range bundle.MapArrays[0] {
		t.Logf("%s %d bytes SHA256 %x", p, len(bundle.ByteArrays[0][i]), sha256.Sum256(bundle.ByteArrays[0][i]))
	}
	buildReturnedBundle(t, bundle)
}

func assertParagraphSourceSyntax(t *testing.T, contents [][]byte, paths []string, printed *string) {
	t.Helper()
	bundle := Logiceme.Executable{ByteArrays: [][][]byte{contents}, MapArrays: [][]string{paths}}
	main := observeBundleMain(t, bundle).Function(t, "main")
	unit := len(main.Result) == 0 || observedUnitOutput(main.Result)
	// The runtime-owned native entry returns process status; ordinary App
	// entry functions remain unit. Both are checked by rustc below.
	status := printed == nil && strings.ReplaceAll(nativetest.RustTokensText(main.Result), " ", "") == "->ExitCode"
	if len(main.Parameters) != 0 || (!unit && !status) {
		t.Fatal("paragraph main signature")
	}
	if printed != nil {
		if got := observedPrinted(t, main.Body); !reflect.DeepEqual(got, []string{*printed}) {
			t.Fatalf("paragraph print=%q want%q", got, *printed)
		}
	}
	buildReturnedBundle(t, bundle)
}
