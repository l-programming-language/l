//go:build darwin || linux

package Logiceme

import (
	"io/fs"
	"os"
	"testing"

	distribution "github.com/l-programming-language/l"
	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
)

// Compile the unchanged unsupported adapter against its actual Tokio interface
// on the current host. This is not execution on an unsupported target platform.
func TestUnsupportedPlatformAdapterInterface(t *testing.T) {
	files := map[string][]byte{
		"src/main.rs": []byte(`mod process;
use std::io::{self, ErrorKind};

fn unsupported<T>(result: io::Result<T>) {
    assert!(matches!(result, Err(error) if error.kind() == ErrorKind::Unsupported));
}

fn main() {
    unsupported(process::Signals::new());
    unsupported(process::configure(&mut tokio::process::Command::new("never-executed")));
    unsupported(process::validate_executable(std::path::Path::new("never-executed")));
    unsupported(process::terminate_group(0));
    let executor = tokio::runtime::Builder::new_current_thread().enable_all().build().unwrap();
    executor.block_on(async {
        let mut signals = process::Signals;
        unsupported(signals.recv().await);
        unsupported(process::final_group_result(0, Ok(()), None).await);
    });
    use std::os::unix::process::ExitStatusExt;
    assert_eq!(process::status_code(std::process::ExitStatus::from_raw(0)), Some(0));
}
`),
	}
	for _, name := range []string{"Cargo.toml", "Cargo.lock"} {
		content, err := os.ReadFile("testdata/runtime/" + name)
		if err != nil {
			t.Fatal(err)
		}
		files[name] = content
	}
	adapter, err := fs.ReadFile(distribution.Files(), "runtime/rust/process_other.rs")
	if err != nil {
		t.Fatal(err)
	}
	files["src/process.rs"] = adapter
	project := nativetest.Materialize(t, files)
	result := nativetest.Build(t, project, nativetest.BuildOptions{Binary: "runtime-parity-public"})
	if result.Err != nil {
		t.Fatalf("unsupported adapter interface compilation: %v\n%s\n%s", result.Err, result.Stdout, result.Stderr)
	}
	var executable string
	for _, artifact := range result.Artifacts {
		if artifact.TargetName == "runtime-parity-public" {
			executable = artifact.Executable
		}
	}
	if executable == "" {
		t.Fatal("unsupported adapter interface build returned no native artifact")
	}
	nativetest.AssertNativeArtifact(t, executable)
	runtimeExecute(t, executable, nil, nil, 0, "", "")
}
