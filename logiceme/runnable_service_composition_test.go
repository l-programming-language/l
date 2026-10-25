package Logiceme

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	"github.com/l-programming-language/l/lerrors"
)

const runnableServiceDependency = `service = { package = "Worker", path = "../../services/service-000" }`

func runnableServiceRead(t testing.TB, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "runnable_service", name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// These are independently authored byte oracles, never generated from AML.
// Existing immutable support expectations are reused for the adapter; the
// Service itself has only its public library source and manifest.
func runnableServiceExpected(t *testing.T, kind string) Executable {
	t.Helper()
	switch kind {
	case "library":
		return Executable{
			MapArrays: [][]string{{"/src/main.rs", "/Cargo.toml"}},
			ByteArrays: [][][]byte{{
				runnableServiceRead(t, "library.main.rs.txt"),
				runnableServiceRead(t, "library.Cargo.toml.txt"),
			}},
			DependencyArrays: [][]string{nil}, EnvDepArrays: [][]string{nil},
		}
	case "adapter":
		return Executable{
			MapArrays: [][]string{{"/src/main.rs", "/Cargo.toml", "/src/config.rs", "/start.sh", "/stop.sh"}},
			ByteArrays: [][][]byte{{
				runnableServiceRead(t, "adapter.main.rs.txt"),
				runnableServiceRead(t, "adapter.Cargo.toml.txt"),
				lucidRead(t, "common/config.rs.txt"),
				runnableServiceRead(t, "adapter.start.sh.txt"),
				lucidRead(t, "common/stop.sh.txt"),
			}},
			DependencyArrays: [][]string{{runnableServiceDependency}}, EnvDepArrays: [][]string{nil},
		}
	case "private-entry":
		bundle := runnableServiceExpected(t, "library")
		bundle.ByteArrays[0][0] = runnableServiceRead(t, "private-entry.main.rs.txt")
		return bundle
	case "private-module":
		return Executable{
			MapArrays: [][]string{{"/src/lib.rs", "/src/worker.rs", "/Cargo.toml"}},
			ByteArrays: [][][]byte{{
				runnableServiceRead(t, "private-module.lib.rs.txt"),
				runnableServiceRead(t, "private-module.worker.rs.txt"),
				runnableServiceRead(t, "private-module.Cargo.toml.txt"),
			}},
			DependencyArrays: [][]string{nil}, EnvDepArrays: [][]string{nil},
		}
	case "private-module-adapter":
		bundle := runnableServiceExpected(t, "adapter")
		bundle.ByteArrays[0][0] = runnableServiceRead(t, "private-module-adapter.main.rs.txt")
		return bundle
	default:
		t.Fatalf("unknown runnable Service fixture %q", kind)
		return Executable{}
	}
}

func runnableServiceCompare(t *testing.T, got, want Executable) {
	t.Helper()
	if got.Name != want.Name || !reflect.DeepEqual(got.MapArrays, want.MapArrays) ||
		!reflect.DeepEqual(got.DependencyArrays, want.DependencyArrays) ||
		!reflect.DeepEqual(got.EnvDepArrays, want.EnvDepArrays) {
		t.Fatalf("complete Service bundle metadata: paths=%#v dependencies=%#v environment=%#v name=%q; want paths=%#v dependencies=%#v environment=%#v name=%q",
			got.MapArrays, got.DependencyArrays, got.EnvDepArrays, got.Name,
			want.MapArrays, want.DependencyArrays, want.EnvDepArrays, want.Name)
	}
	if len(got.ByteArrays) != 1 || len(got.ByteArrays[0]) != len(want.ByteArrays[0]) {
		t.Fatalf("Service bundle dimensions = %d; want one complete package", len(got.ByteArrays))
	}
	for i, expected := range want.ByteArrays[0] {
		if !bytes.Equal(got.ByteArrays[0][i], expected) {
			t.Errorf("%s byte mismatch:\nactual: %q\nexpected: %q", want.MapArrays[0][i], got.ByteArrays[0][i], expected)
		}
	}
}

func runnableServiceGenerate(t *testing.T, service *LogicemeService, kind string) Executable {
	t.Helper()
	input := runnableServiceRead(t, kind+".aml.txt")
	before := bytes.Clone(input)
	bundle, err := service.Evaluate(input, GenerationOptions{})
	if !bytes.Equal(input, before) {
		t.Fatal("runnable Service generation mutated framed AML")
	}
	if err != nil {
		t.Fatalf("runnable Service %s generation: %v; source/native assertions not exercised", kind, err)
	}
	nativetest.RecordBundle(t, kind, bundle)
	return bundle
}

// These controls intentionally pass before Service frontend support: existing
// frozen A/FW/W/BOD/E compositions already express both required crate roles.
// They are not evidence that L construction, adapter selection, or Run works.
func TestRunnableServiceComposition(t *testing.T) {
	for _, provider := range []string{"filesystem", "embedded"} {
		t.Run(provider, func(t *testing.T) {
			config := runtimeGenerationResources(t)
			if provider == "embedded" {
				config = nil
			}
			service := NewLogicemeService(config)
			library := runnableServiceGenerate(t, service, "library")
			runnableServiceCompare(t, library, runnableServiceExpected(t, "library"))
			adapter := runnableServiceGenerate(t, service, "adapter")
			runnableServiceCompare(t, adapter, runnableServiceExpected(t, "adapter"))
			repeated := runnableServiceGenerate(t, service, "library")
			runnableServiceCompare(t, repeated, runnableServiceExpected(t, "library"))
			runnableServiceCompare(t, library, runnableServiceExpected(t, "library"))
			runnableServiceCompare(t, adapter, runnableServiceExpected(t, "adapter"))
		})
	}
}

func TestRunnableServiceLibraryBoundaries(t *testing.T) {
	base := string(runnableServiceRead(t, "library.aml.txt"))
	for _, test := range []struct {
		name, source, diagnostic string
		runtime                  bool
	}{
		{"runtime", base, "library-only package cannot include runtime", true},
		{"settings", strings.Replace(base, "<SETTINGS></SETTINGS>", `<SETTINGS><JSON>{"Key":"TOKEN","Type":"String","Value":"service"}</JSON></SETTINGS>`, 1), "library-only package cannot supply executable configuration", false},
		{"primary", strings.Replace(base, `"Primary":false`, `"Primary":true`, 1), "library target cannot be Primary", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			input := []byte(test.source)
			before := bytes.Clone(input)
			bundle, err := NewLogicemeService(runtimeGenerationResources(t)).Evaluate(input, GenerationOptions{IncludeRuntime: test.runtime})
			if !errors.Is(err, lerrors.InvalidInput) || !strings.Contains(err.Error(), test.diagnostic) {
				t.Fatalf("library boundary = %v; want InvalidInput containing %q", err, test.diagnostic)
			}
			if !reflect.DeepEqual(bundle, Executable{}) || !bytes.Equal(input, before) {
				t.Fatal("invalid library leaked a partial source bundle or mutated AML")
			}
		})
	}
}

// Visibility comes from real version-2 source ownership. The existence of an
// adapter does not permit automatically exporting a private item or ancestor.
func TestRunnableServicePrivateComposition(t *testing.T) {
	for _, kind := range []string{"private-entry", "private-module", "private-module-adapter"} {
		t.Run(kind, func(t *testing.T) {
			service := NewLogicemeService(runtimeGenerationResources(t))
			bundle := runnableServiceGenerate(t, service, kind)
			runnableServiceCompare(t, bundle, runnableServiceExpected(t, kind))
		})
	}
}

func runnableServicePrivateDiagnostic(t *testing.T, result nativetest.BuildResult) {
	t.Helper()
	var exit *exec.ExitError
	if !errors.As(result.Err, &exit) || exit.ExitCode() != 101 {
		t.Fatalf("private entry must fail in rustc, not setup: %v\n%s", result.Err, result.Stderr)
	}
	matched := 0
	for _, line := range bytes.Split(result.Stdout, []byte{'\n'}) {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var event struct {
			Reason  string `json:"reason"`
			Message struct {
				Level string `json:"level"`
				Code  *struct {
					Code string `json:"code"`
				} `json:"code"`
				Spans []struct {
					File    string `json:"file_name"`
					Line    int    `json:"line_start"`
					Primary bool   `json:"is_primary"`
				} `json:"spans"`
			} `json:"message"`
		}
		if err := json.Unmarshal(line, &event); err != nil {
			t.Fatalf("invalid Cargo diagnostic event: %v", err)
		}
		if event.Reason != "compiler-message" || event.Message.Level != "error" {
			continue
		}
		if event.Message.Code == nil || event.Message.Code.Code != "E0603" {
			t.Fatalf("private entry failed for an unrelated Rust error: %s", line)
		}
		for _, span := range event.Message.Spans {
			if span.Primary && span.Line > 0 && strings.HasSuffix(filepath.ToSlash(span.File), "adapters/service-000/src/main.rs") {
				matched++
			}
		}
	}
	if matched != 1 {
		t.Fatalf("E0603 primary adapter source spans = %d; want one\n%s", matched, result.Stdout)
	}
}

func TestRunnableServicePrivateCompositionNative(t *testing.T) {
	for _, kind := range []string{"private-entry", "private-module"} {
		t.Run(kind, func(t *testing.T) {
			adapterKind := "adapter"
			if kind == "private-module" {
				adapterKind = "private-module-adapter"
			}
			for _, phase := range []string{"Expected", "Generated"} {
				t.Run(phase, func(t *testing.T) {
					library, adapter := runnableServiceExpected(t, kind), runnableServiceExpected(t, adapterKind)
					if phase == "Generated" {
						service := NewLogicemeService(runtimeGenerationResources(t))
						library = runnableServiceGenerate(t, service, kind)
						adapter = runnableServiceGenerate(t, service, adapterKind)
						runnableServiceCompare(t, library, runnableServiceExpected(t, kind))
						runnableServiceCompare(t, adapter, runnableServiceExpected(t, adapterKind))
					}
					files := runnableServiceWorkspace(t, library, adapter)
					project := nativetest.Materialize(t, files)
					result := nativetest.Build(t, project, nativetest.BuildOptions{Timeout: nativetest.BuildPrerequisiteTimeout})
					runnableServicePrivateDiagnostic(t, result)
					for name, expected := range files {
						actual, err := os.ReadFile(filepath.Join(project, filepath.FromSlash(name)))
						if err != nil || !bytes.Equal(actual, expected) {
							t.Errorf("native rejection changed input %s: %v", name, err)
						}
					}
				})
			}
		})
	}
}

func runnableServiceWorkspace(t *testing.T, library, adapter Executable) map[string][]byte {
	t.Helper()
	files := map[string][]byte{
		"Cargo.toml": runnableServiceRead(t, "workspace.Cargo.toml.txt"),
		"Cargo.lock": runnableServiceRead(t, "Cargo.lock"),
	}
	for _, member := range []struct {
		directory string
		bundle    Executable
	}{{"services/service-000", library}, {"adapters/service-000", adapter}} {
		if len(member.bundle.MapArrays) != 1 || len(member.bundle.ByteArrays) != 1 || len(member.bundle.MapArrays[0]) != len(member.bundle.ByteArrays[0]) {
			t.Fatalf("incomplete %s source bundle", member.directory)
		}
		for i, path := range member.bundle.MapArrays[0] {
			if !strings.HasPrefix(path, "/") || strings.Contains(path, "..") {
				t.Fatalf("invalid logical source path %q", path)
			}
			name := member.directory + path
			if _, found := files[name]; found {
				t.Fatalf("duplicate source file %s", name)
			}
			files[name] = bytes.Clone(member.bundle.ByteArrays[0][i])
		}
	}
	return files
}

// Expected and Generated compile separately against one retained local-only
// workspace lock. No generated source/manifest is repaired or replaced. This
// test executes Cargo, never the adapter or generated deployment scripts.
func TestRunnableServiceCompositionNative(t *testing.T) {
	for _, phase := range []string{"Expected", "Generated"} {
		t.Run(phase, func(t *testing.T) {
			library, adapter := runnableServiceExpected(t, "library"), runnableServiceExpected(t, "adapter")
			if phase == "Generated" {
				service := NewLogicemeService(runtimeGenerationResources(t))
				library = runnableServiceGenerate(t, service, "library")
				adapter = runnableServiceGenerate(t, service, "adapter")
				runnableServiceCompare(t, library, runnableServiceExpected(t, "library"))
				runnableServiceCompare(t, adapter, runnableServiceExpected(t, "adapter"))
			}
			files := runnableServiceWorkspace(t, library, adapter)
			project := nativetest.Materialize(t, files)
			result := nativetest.Build(t, project, nativetest.BuildOptions{Timeout: nativetest.BuildPrerequisiteTimeout})
			t.Logf("%s native compilation command: %q; runtime not exercised", phase, result.Command)
			if result.Err != nil {
				t.Fatalf("%s Service adapter/library compilation: %v\nstdout:\n%s\nstderr:\n%s", phase, result.Err, result.Stdout, result.Stderr)
			}
			var adapters int
			for _, artifact := range result.Artifacts {
				if artifact.Executable != "" {
					if artifact.TargetName != "l_service_worker" || !reflect.DeepEqual(artifact.TargetKinds, []string{"bin"}) {
						t.Errorf("unexpected native executable: %#v", artifact)
					}
					adapters++
				}
			}
			if adapters != 1 {
				t.Errorf("adapter executable count = %d; want 1", adapters)
			}
			for name, expected := range files {
				actual, err := os.ReadFile(filepath.Join(project, filepath.FromSlash(name)))
				if err != nil || !bytes.Equal(actual, expected) {
					t.Errorf("native compilation changed retained input %s: %v", name, err)
				}
			}
		})
	}
}
