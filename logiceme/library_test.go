package Logiceme

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	"github.com/l-programming-language/l/lerrors"
)

func libraryFixture(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile("testdata/library/library.aml.txt")
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestLibraryOnlyBundle(t *testing.T) {
	input := libraryFixture(t)
	before := bytes.Clone(input)
	service := NewLogicemeService(runtimeGenerationResources(t))
	bundle, err := service.Evaluate(input, GenerationOptions{})
	nativetest.RecordBundle(t, "library", bundle)
	if !bytes.Equal(input, before) {
		t.Fatal("library evaluation mutated AML")
	}
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(bundle.MapArrays, [][]string{{"/src/lib.rs", "/src/detail.rs", "/src/private.rs", "/Cargo.toml"}}) {
		t.Fatalf("library artifact inventory: %#v", bundle.MapArrays)
	}
	sources := map[string][]byte{}
	for i, p := range bundle.MapArrays[0] {
		sources[p] = bundle.ByteArrays[0][i]
	}
	for _, marker := range []string{"pub fn identity", "pub fn make", "pub mod detail;"} {
		if !bytes.Contains(sources["/src/lib.rs"], []byte(marker)) {
			t.Errorf("library lacks %s:\n%s", marker, sources["/src/lib.rs"])
		}
	}
	// HY/MY retain their template-owned initial newline after the separately
	// assembled visibility token; rustc validates these as one public item.
	for _, declaration := range []string{`pub\s+struct Shared`, `pub\s+trait Marker`} {
		if !regexp.MustCompile(declaration).Match(sources["/src/lib.rs"]) {
			t.Errorf("missing exported declaration %s", declaration)
		}
	}
	if !bytes.Contains(sources["/src/detail.rs"], []byte("pub fn preserve")) {
		t.Fatal("public child module did not export its composed function")
	}
	if bytes.Contains(sources["/src/private.rs"], []byte("pub ")) {
		t.Fatal("private Object exported its function")
	}
	for _, source := range []string{"/src/lib.rs", "/src/detail.rs", "/src/private.rs"} {
		for _, fragment := range []string{"mod config;", "mod runtime;", "fn main(", "initialize_configuration_settings"} {
			if bytes.Contains(sources[source], []byte(fragment)) {
				t.Errorf("library acquired executable source %q", fragment)
			}
		}
	}
	cargo := sources["/Cargo.toml"]
	if !bytes.Contains(cargo, []byte("[lib]\nname = \"library_fixture\"\npath = \"src/lib.rs\"")) || bytes.Contains(cargo, []byte("[[bin]]")) {
		t.Fatalf("library target manifest:\n%s", cargo)
	}
	repeated, err := service.Evaluate(input, GenerationOptions{})
	if err != nil || !reflect.DeepEqual(repeated, bundle) {
		t.Fatalf("repeat library evaluation: %v", err)
	}
	// The independent consumer uses one actual library crate, preserving nominal
	// type identity across the root and its public child module.
	files := map[string][]byte{}
	for path, data := range sources {
		files["library/"+strings.TrimPrefix(path, "/")] = bytes.Clone(data)
	}
	for destination, fixture := range map[string]string{"Cargo.toml": "workspace.toml", "Cargo.lock": "Cargo.lock", "consumer/Cargo.toml": "consumer.toml", "consumer/src/main.rs": "consumer.rs"} {
		b, err := os.ReadFile(filepath.Join("testdata/library", fixture))
		if err != nil {
			t.Fatal(err)
		}
		files[destination] = b
	}
	project := nativetest.Materialize(t, files)
	result := nativetest.Build(t, project, nativetest.BuildOptions{TargetDir: filepath.Join(t.TempDir(), "target")})
	if result.Err != nil {
		t.Fatalf("actual library and independent path consumer fail: %v\n%s", result.Err, result.Stderr)
	}
}

func TestLibraryOnlyRejectsExecutableOptions(t *testing.T) {
	base := string(libraryFixture(t))
	for _, tc := range []struct {
		name, input, diagnostic string
		runtime                 bool
	}{
		{"primary library", strings.Replace(base, `"Primary":false`, `"Primary":true`, 1), "library target cannot be Primary", false},
		{"runtime", base, "library-only package cannot include runtime", true},
		{"configuration", strings.Replace(base, "<SETTINGS></SETTINGS>", `<SETTINGS><JSON>{"Key":"TOKEN","Type":"String","Value":"value"}</JSON></SETTINGS>`, 1), "library-only package cannot supply executable configuration", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := []byte(tc.input)
			before := bytes.Clone(input)
			bundle, err := NewLogicemeService(runtimeGenerationResources(t)).Evaluate(input, GenerationOptions{IncludeRuntime: tc.runtime})
			if err == nil || !errors.Is(err, lerrors.InvalidInput) || !strings.Contains(err.Error(), tc.diagnostic) {
				t.Fatalf("library failure = %v; want %s", err, tc.diagnostic)
			}
			if !reflect.DeepEqual(bundle, Executable{}) || !bytes.Equal(input, before) {
				t.Fatal("invalid library mutated input or exposed partial bundle")
			}
		})
	}
}

func TestLibraryTargetSelectionRetainsBinaryRules(t *testing.T) {
	for _, tc := range []struct {
		name    string
		targets []TargetDescriptor
		cause   string
	}{
		{"sole library", []TargetDescriptor{{Kind: "lib", Name: "library", Root: "lib"}}, ""},
		{"mixed primary", []TargetDescriptor{{Kind: "lib", Name: "library", Root: "lib"}, {Kind: "bin", Name: "app", Root: "main", Primary: true}}, ""},
		{"mixed missing primary", []TargetDescriptor{{Kind: "lib", Name: "library", Root: "lib"}, {Kind: "bin", Name: "app", Root: "main"}}, "exactly one Primary binary"},
		{"binary missing primary", []TargetDescriptor{{Kind: "bin", Name: "app", Root: "main"}}, "exactly one Primary binary"},
		{"multiple primary", []TargetDescriptor{{Kind: "bin", Name: "app", Root: "main", Primary: true}, {Kind: "bin", Name: "other", Root: "other", Primary: true}}, "exactly one Primary binary"},
		{"multiple libraries", []TargetDescriptor{{Kind: "lib", Name: "library", Root: "lib"}, {Kind: "lib", Name: "other", Root: "other"}}, "at most one library target"},
		{"no targets", nil, "exactly one Primary binary"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var outline Outline
			objects := map[string]*Object{}
			for i := range tc.targets {
				target := tc.targets[i]
				outline.Services = append(outline.Services, OutlineEntry{Name: target.Name, GenericObjects: []string{target.Root}, Target: &target})
				objects[target.Root] = &Object{Name: target.Root}
			}
			got, err := prepareSourceTargets(outline, objects)
			if tc.cause != "" {
				if err == nil || !errors.Is(err, lerrors.InvalidInput) || !strings.Contains(err.Error(), tc.cause) {
					t.Fatalf("target validation=%v; want %s", err, tc.cause)
				}
				return
			}
			if err != nil || !reflect.DeepEqual(got, tc.targets) {
				t.Fatalf("targets changed: %#v, %v", got, err)
			}
		})
	}
}
