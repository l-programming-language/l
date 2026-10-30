//go:build darwin || linux

package Logiceme

import (
	"os"
	"path/filepath"
	"testing"

	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
)

func buildErrorRuntime(t *testing.T, source string, observed bool) string {
	t.Helper()
	read := func(path string) []byte {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	files := map[string][]byte{
		"Cargo.toml":       read("testdata/runtime/Cargo.toml"),
		"Cargo.lock":       read("testdata/runtime/Cargo.lock"),
		"src/helper.rs":    read("testdata/runtime/helper.rs.txt"),
		"src/ownership.rs": read("../testdata/phase1/runtime/ownership.rs.txt"),
		"src/main.rs":      []byte(source),
	}
	for name, data := range runtimeSourceFiles(t) {
		files[name] = data
	}
	cache, err := filepath.Abs("../.harness/cache/rust-empty-error-runtime")
	if err != nil {
		t.Fatal(err)
	}
	project := nativetest.Materialize(t, files)
	options := nativetest.BuildOptions{TargetDir: cache}
	if observed {
		options.Binary = "runtime-parity-public"
		options.TestConfiguration = true
	}
	result := nativetest.Build(t, project, options)
	if observed && result.Err == nil {
		helper := nativetest.Build(t, project, nativetest.BuildOptions{TargetDir: cache, Binary: "helper"})
		result.Artifacts = append(result.Artifacts, helper.Artifacts...)
		if helper.Err != nil {
			result.Err = helper.Err
			result.Stderr = append(result.Stderr, helper.Stderr...)
		}
	}
	return runtimeDriver(t, result)
}

func TestRuntimeEmptyErrorDiagnostic(t *testing.T) {
	t.Parallel()
	driver := buildErrorRuntime(t, `mod runtime;
#[path = "process_unix.rs"] mod process;
mod ownership;
fn main() -> std::process::ExitCode {
    let executor = tokio::runtime::Builder::new_current_thread().enable_all().build().unwrap();
    let status = executor.block_on(async {
        let runtime = runtime::Runtime::new().unwrap();
        let app = runtime.app("companion/helper");
        let invocation = app.run().unwrap();
        let pid = std::env::var("L_RUNTIME_PID").unwrap();
        let until = tokio::time::Instant::now() + std::time::Duration::from_secs(3);
        while !std::path::Path::new(&pid).exists() {
            assert!(tokio::time::Instant::now() < until, "real child readiness");
            tokio::time::sleep(std::time::Duration::from_millis(1)).await;
        }
        let status = runtime.finish(Err(runtime::Error::from_bytes(Vec::new()))).await;
        assert!(invocation.snapshot().is_some(), "empty Error skipped capture/Wait settlement");
        assert!(!ownership::held(std::path::Path::new(&format!("{pid}.lock"))).unwrap(), "empty Error left child alive");
        status
    });
    std::process::ExitCode::from(status)
}
`, false)
	pid := filepath.Join(t.TempDir(), "child-pid")
	runtimeExecute(t, driver, []string{"L_RUNTIME_CASE=hold", "L_RUNTIME_PID=" + pid}, nil, 1, "", "\n")
}
