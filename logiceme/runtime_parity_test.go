//go:build darwin || linux

package Logiceme

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	distribution "github.com/l-programming-language/l"
	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	"github.com/l-programming-language/l/internal/testsupport/phase1"
)

func runtimeFixture(t testing.TB, observed bool, sourceOverride ...string) (string, nativetest.BuildResult) {
	t.Helper()
	read := func(path string) []byte {
		b, e := os.ReadFile(path)
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	files := map[string][]byte{
		"Cargo.toml": read("testdata/runtime/Cargo.toml"), "Cargo.lock": read("testdata/runtime/Cargo.lock"),
		"src/main.rs": read("testdata/runtime/public.rs.txt"), "src/helper.rs": read("testdata/runtime/helper.rs.txt"),
		"src/ownership.rs": read("../testdata/phase1/runtime/ownership.rs.txt"),
	}
	if observed {
		files["src/main.rs"] = read("testdata/runtime/observed.rs.txt")
	}
	if len(sourceOverride) != 0 {
		files["src/main.rs"] = read(sourceOverride[0])
	}
	for _, name := range []string{"runtime", "process_unix"} {
		b, e := fs.ReadFile(distribution.Files(), "builtins/templates/runtime/"+name+".lu")
		if e != nil {
			t.Fatal(e)
		}
		files["src/"+name+".rs"] = b
	}
	project := nativetest.Materialize(t, files)
	cache, e := filepath.Abs("../.harness/cache/rust-runtime-parity")
	if e != nil {
		t.Fatal(e)
	}
	options := nativetest.BuildOptions{TargetDir: cache}
	if observed {
		options.Binary = "runtime-parity-public"
		options.TestConfiguration = true
	}
	result := nativetest.Build(t, project, options)
	if observed && result.Err == nil {
		helper := nativetest.Build(t, project, nativetest.BuildOptions{Binary: "helper", TargetDir: cache})
		result.Artifacts = append(result.Artifacts, helper.Artifacts...)
		if helper.Err != nil {
			result.Err = helper.Err
			result.Stderr = append(result.Stderr, helper.Stderr...)
		}
	}
	evidence := filepath.Join("..", ".harness", "runs", "native-implementation", "distribution", strings.ReplaceAll(t.Name(), "/", "-"))
	if e := os.MkdirAll(evidence, 0700); e != nil {
		t.Fatal(e)
	}
	for name, data := range map[string][]byte{"cargo.jsonl": result.Stdout, "cargo.stderr.txt": result.Stderr, "Cargo.toml": files["Cargo.toml"], "Cargo.lock": files["Cargo.lock"]} {
		if e := os.WriteFile(filepath.Join(evidence, name), data, 0600); e != nil {
			t.Fatal(e)
		}
	}
	for _, name := range []string{"runtime", "process_unix"} {
		got := read(filepath.Join(project, "src", name+".rs"))
		if !bytes.Equal(got, files["src/"+name+".rs"]) {
			t.Fatal("runtime asset changed during compilation")
		}
	}
	return project, result
}
func runtimeDriver(t testing.TB, result nativetest.BuildResult) string {
	t.Helper()
	if result.Err != nil {
		t.Fatalf("actual Rust support compilation blocked runtime assertions: %v\n%s\n%s", result.Err, result.Stdout, result.Stderr)
	}
	var driver, helper string
	for _, a := range result.Artifacts {
		switch a.TargetName {
		case "runtime-parity-public":
			driver = a.Executable
		case "helper":
			helper = a.Executable
		}
	}
	if driver == "" || helper == "" {
		t.Fatalf("missing Cargo executable artifacts: %+v", result.Artifacts)
	}
	nativetest.AssertNativeArtifact(t, driver)
	nativetest.AssertNativeArtifact(t, helper)
	dir := t.TempDir()
	out := filepath.Join(dir, "runtime-check")
	phase1.CopyFile(t, driver, out, 0700)
	if e := os.Mkdir(filepath.Join(dir, "companion"), 0700); e != nil {
		t.Fatal(e)
	}
	phase1.CopyFile(t, helper, filepath.Join(dir, "companion", "helper"), 0700)
	return out
}
func runtimeExecute(t testing.TB, driver string, env, args []string, status int, out, diag string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.Command(driver, args...)
	cmd.Dir = t.TempDir()
	cmd.Env = append(os.Environ(), env...)
	r := phase1.RunRuntimeDriver(ctx, cmd)
	code := 0
	if r.Err != nil {
		var exit *exec.ExitError
		if !errors.As(r.Err, &exit) {
			t.Fatalf("runtime supervision: %v", r.Err)
		}
		code = exit.ExitCode()
	}
	if r.CleanupErr != nil || len(r.Leaks) != 0 {
		t.Fatalf("runtime ownership violation before fallback: leaks=%+v cleanup=%v", r.Leaks, r.CleanupErr)
	}
	if code != status || string(r.Stdout) != out || string(r.Stderr) != diag {
		t.Fatalf("Rust runtime: status=%d stdout=%q stderr=%q; want %d %q %q", code, r.Stdout, r.Stderr, status, out, diag)
	}
}
func TestRuntimeTemplateBehavior(t *testing.T) {
	t.Parallel()
	_, result := runtimeFixture(t, false)
	t.Run("TestReadinessAndCapturedOutput", func(t *testing.T) {
		driver := runtimeDriver(t, result)
		mode := filepath.Join(t.TempDir(), "mode")
		if e := os.WriteFile(mode, nil, 0600); e != nil {
			t.Fatal(e)
		}
		runtimeExecute(t, driver, []string{"L_RUNTIME_CASE_FILE=" + mode}, []string{"readiness"}, 0, "", "")
	})
	t.Run("TestRuntimeVisibleStreams", func(t *testing.T) {
		t.Run("fail nil", func(t *testing.T) {
			_, native := runtimeFixture(t, false, "testdata/runtime/fail_absent.rs.txt")
			driver := runtimeDriver(t, native)
			runtimeExecute(t, driver, nil, nil, 1, "", "cannot exit with an absent Error\n")
		})
		for _, tc := range []struct {
			name, out, err string
			code           int
		}{
			{"show", "line\n\n雪\t", "", 0}, {"fail", "", "failure text\n", 1}, {"fail newline", "", "failure text\n\n", 1},
			{"fail multiline", "", "first\nsecond\n\n", 1},
			{"output", "", "Output is unavailable before completed invocation capture\n", 1},
		} {
			t.Run(tc.name, func(t *testing.T) {
				driver := runtimeDriver(t, result)
				runtimeExecute(t, driver, nil, []string{"api", tc.name}, tc.code, tc.out, tc.err)
			})
		}
	})
	t.Run("TestDescendantsDoNotHoldRunOpen", func(t *testing.T) {
		for _, mode := range []string{"descendant", "failed descendant", "closed descendant", "failed closed descendant"} {
			t.Run(mode, func(t *testing.T) {
				driver := runtimeDriver(t, result)
				pid := filepath.Join(t.TempDir(), "pid")
				runtimeExecute(t, driver, []string{"L_RUNTIME_CASE=" + mode, "L_RUNTIME_PID=" + pid}, []string{"descendant"}, 0, "", "")
			})
		}
	})
	t.Run("TestAlreadyCanceledRunDoesNotStartApplication", func(t *testing.T) {
		driver := runtimeDriver(t, result)
		pid := filepath.Join(t.TempDir(), "pid")
		runtimeExecute(t, driver, []string{"L_RUNTIME_CASE=hold", "L_RUNTIME_PID=" + pid}, []string{"cancelled"}, 0, "", "")
	})
	t.Run("TestSignalCancelsRunningApplication", func(t *testing.T) {
		driver := runtimeDriver(t, result)
		dir := t.TempDir()
		childPID := filepath.Join(dir, "pid")
		driverPID := filepath.Join(dir, "driver")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cmd := exec.Command(driver, "signal")
		cmd.Dir = t.TempDir()
		cmd.Env = append(os.Environ(), "L_RUNTIME_CASE=hold", "L_RUNTIME_PID="+childPID, "L_RUNTIME_DRIVER_PID="+driverPID)
		signalErr := make(chan error, 1)
		go func() {
			deadline := time.Now().Add(5 * time.Second)
			for time.Now().Before(deadline) {
				if _, e := os.Stat(childPID); e == nil {
					data, e := os.ReadFile(driverPID)
					if e != nil {
						signalErr <- e
						return
					}
					pid, e := strconv.Atoi(string(data))
					if e == nil {
						e = syscall.Kill(pid, syscall.SIGTERM)
					}
					signalErr <- e
					return
				}
				time.Sleep(10 * time.Millisecond)
			}
			signalErr <- fmt.Errorf("signal target did not start")
		}()
		r := phase1.RunRuntimeDriver(ctx, cmd)
		if e := <-signalErr; e != nil {
			t.Fatal(e)
		}
		var exit *exec.ExitError
		if !errors.As(r.Err, &exit) || exit.ExitCode() != 143 || len(r.Stdout) != 0 || len(r.Stderr) != 0 {
			t.Fatalf("signal exit=%v stdout=%q stderr=%q", r.Err, r.Stdout, r.Stderr)
		}
		if r.CleanupErr != nil || len(r.Leaks) != 0 {
			t.Fatalf("signal leaked App before fallback: %+v %v", r.Leaks, r.CleanupErr)
		}
	})
}

func TestRuntimeObservedParity(t *testing.T) {
	t.Parallel()
	_, result := runtimeFixture(t, true)
	// These profiles require the explicitly documented private observer contract.
	// Missing observer hooks are a production prerequisite failure, never a skip.
	for _, profile := range []string{"already-cancelled", "stop-error-retained", "cleanup-once", "zombie-reaped-within-original-window", "persistent-zombie-retains-permission-error", "signal-time-consumes-window-and-later-signals-do-not-renew-it", "no-prior-success-keeps-final-error", "live-permission-error-and-other-causes-remain"} {
		t.Run(profile, func(t *testing.T) {
			driver := runtimeDriver(t, result)
			pid := filepath.Join(t.TempDir(), "pid")
			code, out, diag := 0, "", ""
			if profile == "cleanup-once" {
				code = 1
				diag = "terminate application process group: injected cleanup failure: notice\n\n"
			}
			runtimeExecute(t, driver, []string{"L_RUNTIME_CASE=" + map[bool]string{true: "", false: "hold"}[profile == "cleanup-once"], "L_RUNTIME_PID=" + pid}, []string{profile}, code, out, diag)
		})
	}
}

func TestRuntimeGenerationParity(t *testing.T) {
	service := NewLogicemeService(runtimeGenerationResources(t))
	for _, location := range []string{"/src/", "/source/"} {
		t.Run(location, func(t *testing.T) {
			input := []byte(fmt.Sprintf(`<PROGRAM><STRUCTURE><BOOK "rust_runtime_generation"><OUTLINE>
<JSON>{"AppMeta":{"AppGenDir":%q,"Deps":"[\"tokio = { version = \\\"=1.53.1\\\", features = [\\\"rt-multi-thread\\\", \\\"macros\\\", \\\"process\\\", \\\"io-util\\\", \\\"io-std\\\", \\\"signal\\\", \\\"sync\\\", \\\"time\\\"] }\",\"nix = { version = \\\"=0.31.3\\\", features = [\\\"signal\\\", \\\"process\\\"] }\"]"}}</JSON>
<JSON>{"Services":[{"Name":"main","Location":%q,"GenericObjects":["main"]}]}</JSON><JSON>{"Objects":[{"Name":"main","Type":"root"}]}</JSON>
</OUTLINE><SETTINGS><JSON>{"key":"Theme","value":"ocean","type":"String"}</JSON></SETTINGS><CONTENT><CHAPTER "main"><SENTENCE "main"><|W|><|S|>`+"`A`"+`</S|><|S|>`+"`FW{|Key: ~l_main~, Inputs: ~&[Prop{Key: runtime, Type: &runtime::Runtime}Prop]&~, Outputs: ~&[impl std::future::Future<Output = runtime::Result<()>> + '_]&~|}^W^BOD^E{|Statement: ~async { Ok(()) }~|}`"+`</S|></W|></SENTENCE></CHAPTER></CONTENT></BOOK></STRUCTURE></PROGRAM>`, strings.TrimSuffix(location, "/"), location))
			original := bytes.Clone(input)
			plain, e := service.Evaluate(input, GenerationOptions{})
			if e != nil {
				t.Fatal(e)
			}
			explicit, e := service.Evaluate(input, GenerationOptions{IncludeRuntime: false})
			if e != nil || !reflect.DeepEqual(plain, explicit) {
				t.Fatalf("explicit false options changed bundle: %v", e)
			}
			nativetest.RecordBundle(t, "runtime-plain", plain)
			before := fmt.Sprintf("%#v", plain)
			withRuntime, e := service.Evaluate(input, GenerationOptions{IncludeRuntime: true})
			if e != nil {
				t.Fatal(e)
			}
			nativetest.RecordBundle(t, "runtime-enabled", withRuntime)
			if len(withRuntime.MapArrays) != 1 || len(withRuntime.MapArrays[0]) != len(plain.MapArrays[0])+3 {
				t.Fatalf("runtime support count: %v", withRuntime.MapArrays)
			}
			count := len(plain.MapArrays[0])
			if !reflect.DeepEqual(withRuntime.MapArrays[0][:count], plain.MapArrays[0]) || !reflect.DeepEqual(withRuntime.MapArrays[0][count:], []string{location + "runtime.rs", location + "process_unix.rs", location + "process_other.rs"}) {
				t.Errorf("runtime support order/locations changed: %v", withRuntime.MapArrays[0])
			}
			if !reflect.DeepEqual(withRuntime.DependencyArrays, plain.DependencyArrays) || !reflect.DeepEqual(withRuntime.EnvDepArrays, plain.EnvDepArrays) {
				t.Error("runtime support changed explicit dependency metadata")
			}
			files := map[string][]byte{}
			for i, path := range withRuntime.MapArrays[0] {
				if _, ok := files[path]; ok {
					t.Fatalf("duplicate path %s", path)
				}
				files[path] = withRuntime.ByteArrays[0][i]
			}
			for _, name := range []string{"runtime", "process_unix", "process_other"} {
				want, e := fs.ReadFile(distribution.Files(), "builtins/templates/runtime/"+name+".lu")
				if e != nil {
					t.Fatal(e)
				}
				if !bytes.Equal(files[location+name+".rs"], want) {
					t.Errorf("runtime provenance: %s", name)
				}
			}
			main := files[location+"main.rs"]
			for _, decl := range []string{"#![forbid(unsafe_code)]", "mod config;", "mod runtime;", "mod process;"} {
				if !bytes.Contains(main, []byte(decl)) {
					t.Errorf("runtime root omitted %q", decl)
				}
			}
			if !bytes.Contains(main, []byte("l_main(&runtime)")) {
				t.Error("runtime entry does not call composed application")
			}
			for i, path := range plain.MapArrays[0] {
				if path != location+"main.rs" && path != "/Cargo.toml" && !bytes.Equal(plain.ByteArrays[0][i], files[path]) {
					t.Errorf("runtime changed unrelated support %s", path)
				}
			}
			materialized := map[string][]byte{}
			for p, data := range files {
				materialized[strings.TrimPrefix(p, "/")] = bytes.Clone(data)
			}
			lock, e := os.ReadFile("testdata/runtime/generation.Cargo.lock")
			if e != nil {
				t.Fatal(e)
			}
			materialized["Cargo.lock"] = lock
			project := nativetest.Materialize(t, materialized)
			native := nativetest.Build(t, project, nativetest.BuildOptions{})
			if native.Err != nil {
				t.Errorf("actual runtime-enabled project Cargo build: %v\n%s", native.Err, native.Stderr)
			} else {
				for _, a := range native.Artifacts {
					nativetest.AssertNativeArtifact(t, a.Executable)
				}
			}
			after, e := service.Evaluate(input, GenerationOptions{})
			if e != nil || !reflect.DeepEqual(plain, after) || fmt.Sprintf("%#v", plain) != before || !bytes.Equal(input, original) {
				t.Fatalf("runtime generation leaked mutation: %v", e)
			}
		})
	}
}
