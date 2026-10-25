package native

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/l-programming-language/l/internal/cargocache"
)

func TestToolchainIdentityMemoPreservesIsolation(t *testing.T) {
	directory := t.TempDir()
	counter := filepath.Join(directory, "queries")
	t.Setenv("L_TEST_IDENTITY_COUNTER", counter)
	write := func(directory, host string) {
		t.Helper()
		if err := os.MkdirAll(directory, 0700); err != nil {
			t.Fatal(err)
		}
		files := map[string]string{
			"rustup": "#!/bin/sh\nprintf 'rustup\\n' >> \"$L_TEST_IDENTITY_COUNTER\"\nif [ \"$1\" != which ] || [ \"$2\" != --toolchain ] || [ \"$3\" != 1.98.1 ]; then exit 7; fi\nprintf '%s/%s\\n' \"${0%/*}\" \"$4\"\n",
			"cargo":  "#!/bin/sh\nexit 0\n",
			"rustc":  fmt.Sprintf("#!/bin/sh\nprintf 'rustc\\n' >> \"$L_TEST_IDENTITY_COUNTER\"\nprintf 'release: 1.98.1\\nhost: %s\\n'\n", host),
		}
		for name, text := range files {
			if err := os.WriteFile(filepath.Join(directory, name), []byte(text), 0700); err != nil {
				t.Fatal(err)
			}
		}
	}
	check := func(wantHost string, queries int) {
		t.Helper()
		identity := Toolchain(t)
		if identity.Host != wantHost {
			t.Fatalf("stale toolchain identity: got %q want %q", identity.Host, wantHost)
		}
		data, err := os.ReadFile(counter)
		if err != nil || len(strings.Fields(string(data))) != queries {
			t.Fatalf("identity query count: got %q want %d (%v)", data, queries, err)
		}
	}
	first := filepath.Join(directory, "first")
	write(first, "first-host")
	t.Setenv("PATH", first)
	check("first-host", 3)
	check("first-host", 3)
	t.Setenv("RUSTUP_HOME", filepath.Join(directory, "another-home"))
	check("first-host", 6)
	second := filepath.Join(directory, "second")
	write(second, "second-host")
	t.Setenv("PATH", second)
	check("second-host", 9)
	write(second, "changed-binary-host")
	check("changed-binary-host", 12)
	check("changed-binary-host", 12)
}

func TestBuildRetainsRelativeTargetWorkspace(t *testing.T) {
	project := Materialize(t, map[string][]byte{
		"Cargo.toml":  []byte("[package]\nname = \"relative_target\"\nversion = \"0.0.0\"\nedition = \"2024\"\n"),
		"Cargo.lock":  []byte("version = 4\n\n[[package]]\nname = \"relative_target\"\nversion = \"0.0.0\"\n"),
		"src/main.rs": []byte("fn main() {}\n"),
	})
	manifest := func(result BuildResult) string {
		t.Helper()
		if result.Err != nil {
			t.Fatalf("native build failed: %v\n%s\n%s", result.Err, result.Stdout, result.Stderr)
		}
		if len(result.Artifacts) != 1 {
			t.Fatalf("expected actual native artifact, got %#v", result.Artifacts)
		}
		AssertNativeArtifact(t, result.Artifacts[0].Executable)
		for index, arg := range result.Command {
			if arg == "--manifest-path" && index+1 < len(result.Command) {
				return result.Command[index+1]
			}
		}
		t.Fatal("Cargo build did not identify its actual manifest")
		return ""
	}
	first := manifest(Build(t, project, BuildOptions{TargetDir: "target"}))
	if err := os.WriteFile(filepath.Join(project, "src/main.rs"), []byte("fn main() { let _value = 1; }\n"), 0600); err != nil {
		t.Fatal(err)
	}
	second := manifest(Build(t, project, BuildOptions{TargetDir: "target"}))
	if second != first {
		t.Fatalf("the owned target/cache entered its own source inventory: first=%s second=%s", first, second)
	}
}

func TestBuildDefaultCacheRetainsCurrentWorkspace(t *testing.T) {
	// The support package rename must not relocate its retained dependency cache.
	wantSuffix := filepath.Join(".harness", "cache", "rust-parity", "internal", "testsupport", "rustparity", Version, "preservation-host")
	if got := defaultTargetDirectory(t, "preservation-host"); !strings.HasSuffix(got, wantSuffix) {
		t.Fatalf("support package relocation changed persistent cache: %s; want suffix %s", got, wantSuffix)
	}
	// Exercise changes in one real default cache slot. The comment only selects
	// the bounded shard; it remains part of the exact input passed to Cargo.
	type variant struct{ member, marker, value string }
	steps := []variant{{"entry_a", "first", "one"}, {"entry_a", "changed-entry", "one"}, {"entry_b", "second", "one"}, {"entry_a", "first", "one"}, {"entry_b", "changed-library", "two"}, {"entry_a", "first", "one"}}
	prepared := make(map[variant]map[string][]byte)
	var retained []struct {
		executable, output string
		bytes              []byte
	}
	firstInputs := ""
	for index, step := range steps {
		files := prepared[step]
		if files == nil {
			files = map[string][]byte{
				"Cargo.toml":                 []byte(fmt.Sprintf("[workspace]\nresolver = \"3\"\nmembers = [%q, \"shared\"]\n", step.member)),
				"Cargo.lock":                 []byte(fmt.Sprintf("version = 4\n\n[[package]]\nname = %q\nversion = \"0.0.0\"\ndependencies = [\"shared\"]\n\n[[package]]\nname = \"shared\"\nversion = \"0.0.0\"\n", step.member)),
				step.member + "/Cargo.toml":  []byte(fmt.Sprintf("[package]\nname = %q\nversion = \"0.0.0\"\nedition = \"2024\"\n[dependencies]\nshared = { path = \"../shared\" }\n", step.member)),
				step.member + "/src/main.rs": []byte(fmt.Sprintf("fn main() { println!(\"%s|{}|{}|{}\", shared::value(), env!(\"CARGO_MANIFEST_DIR\"), shared::origin()); }\n", step.marker)),
				"shared/Cargo.toml":          []byte("[package]\nname = \"shared\"\nversion = \"0.0.0\"\nedition = \"2024\"\n"),
				"shared/src/lib.rs":          []byte(fmt.Sprintf("pub fn value() -> &'static str { %q }\npub fn origin() -> &'static str { env!(\"CARGO_MANIFEST_DIR\") }\n", step.value)),
			}
			candidate := Materialize(t, files)
			sourceName := step.member + "/src/main.rs"
			original := bytes.Clone(files[sourceName])
			for nonce := 0; ; nonce++ {
				if nonce == 256 {
					t.Fatal("could not select the bounded native test cache shard")
				}
				files[sourceName] = append(bytes.Clone(original), []byte(fmt.Sprintf("// native cache witness %d\n", nonce))...)
				if err := os.WriteFile(filepath.Join(candidate, sourceName), files[sourceName], 0600); err != nil {
					t.Fatal(err)
				}
				shard, err := cargocache.Shard(candidate, "")
				if err != nil {
					t.Fatal(err)
				}
				if shard == 0 {
					break
				}
			}
			prepared[step] = files
		}
		workspace := Materialize(t, files)
		result := Build(t, filepath.Join(workspace, step.member), BuildOptions{})
		manifest := actualBuildManifest(t, result)
		inputs := filepath.Dir(filepath.Dir(manifest))
		if firstInputs == "" {
			firstInputs = inputs
		} else if inputs != firstInputs {
			t.Errorf("build %d changed the default cache package root: first=%s current=%s", index, firstInputs, inputs)
		}
		if filepath.Base(filepath.Dir(manifest)) != step.member {
			t.Fatalf("build %d selected a different workspace member: %s", index, manifest)
		}
		for name, directory := range map[string]string{"caller": workspace, "Cargo": inputs} {
			if actual := nativeInputFiles(t, directory); !reflect.DeepEqual(actual, files) {
				t.Errorf("build %d changed %s workspace bytes or retained obsolete members: got %v want %v", index, name, actual, files)
			}
		}
		if len(result.Artifacts) != 1 || result.Artifacts[0].TargetName != step.member {
			t.Fatalf("build %d omitted selected member artifact: %#v", index, result.Artifacts)
		}
		executable := result.Artifacts[0].Executable
		AssertNativeArtifact(t, executable)
		want := step.marker + "|" + step.value + "|" + filepath.Dir(manifest) + "|" + filepath.Join(inputs, "shared") + "\n"
		if output := nativeCacheOutput(t, executable); output != want {
			t.Fatalf("build %d used stale source, library, or native path: got %q want %q", index, output, want)
		}
		data, err := os.ReadFile(executable)
		if err != nil {
			t.Fatal(err)
		}
		retained = append(retained, struct {
			executable, output string
			bytes              []byte
		}{executable, want, data})
	}
	for index, prior := range retained {
		data, err := os.ReadFile(prior.executable)
		if err != nil || !bytes.Equal(data, prior.bytes) {
			t.Fatalf("later build changed executable snapshot %d: %v", index, err)
		}
		if output := nativeCacheOutput(t, prior.executable); output != prior.output {
			t.Fatalf("later build changed native snapshot %d: got %q want %q", index, output, prior.output)
		}
	}
}

func TestBuildExplicitTargetRetainsGraphSpecificInputs(t *testing.T) {
	target := t.TempDir()
	files := map[string][]byte{
		"Cargo.toml":  []byte("[package]\nname = \"explicit_target\"\nversion = \"0.0.0\"\nedition = \"2024\"\n"),
		"Cargo.lock":  []byte("version = 4\n\n[[package]]\nname = \"explicit_target\"\nversion = \"0.0.0\"\n"),
		"src/main.rs": []byte("fn main() {}\n"),
	}
	project := Materialize(t, files)
	first := actualBuildManifest(t, Build(t, project, BuildOptions{TargetDir: target}))
	files["Cargo.toml"] = append(bytes.Clone(files["Cargo.toml"]), []byte("description = \"changed manifest\"\n")...)
	if err := os.WriteFile(filepath.Join(project, "Cargo.toml"), files["Cargo.toml"], 0600); err != nil {
		t.Fatal(err)
	}
	second := actualBuildManifest(t, Build(t, project, BuildOptions{TargetDir: target}))
	if first == second {
		t.Fatal("explicit target lost its graph-specific Stage path after a manifest change")
	}
	if actual := nativeInputFiles(t, filepath.Dir(second)); !reflect.DeepEqual(actual, files) {
		t.Fatal("explicit target changed supplied workspace bytes")
	}
}

func actualBuildManifest(t *testing.T, result BuildResult) string {
	t.Helper()
	if result.Err != nil {
		t.Fatalf("actual Cargo build failed: %v\n%s\n%s", result.Err, result.Stdout, result.Stderr)
	}
	for index, arg := range result.Command {
		if arg == "--manifest-path" && index+1 < len(result.Command) {
			return result.Command[index+1]
		}
	}
	t.Fatal("actual Cargo command omitted its selected manifest")
	return ""
}

func nativeInputFiles(t *testing.T, directory string) map[string][]byte {
	t.Helper()
	files := map[string][]byte{}
	err := filepath.WalkDir(directory, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("nonregular native input %s", path)
		}
		name, err := filepath.Rel(directory, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		files[filepath.ToSlash(name)] = data
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func nativeCacheOutput(t *testing.T, executable string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	output, err := exec.CommandContext(ctx, executable).CombinedOutput()
	if err != nil {
		t.Fatalf("native cache witness failed: %v (%v)\n%s", err, ctx.Err(), output)
	}
	return string(output)
}
