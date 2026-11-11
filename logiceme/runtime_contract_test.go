//go:build darwin || linux

package Logiceme

import (
	"bytes"
	"io/fs"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	distribution "github.com/l-programming-language/l"
	distributiontest "github.com/l-programming-language/l/internal/testsupport/distribution"
	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	Lyger "github.com/l-programming-language/l/lyger"
)

func composeContractProvider(t *testing.T, resources fstest.MapFS) {
	t.Helper()
	aml, manifest, err := Lyger.ComposeRuntimeResources(resources)
	if err != nil {
		t.Fatal(err)
	}
	resources["stdlib/runtime/generated/support.aml"] = &fstest.MapFile{Data: aml}
	resources["stdlib/runtime/generated/support.json"] = &fstest.MapFile{Data: marshalTestJSON(t, manifest)}
}

func TestAlternateRuntimePolicyComposition(t *testing.T) {
	resources := distributiontest.AlternateRuntime(t)
	composeContractProvider(t, resources)
	first, metadata := bytes.Clone(resources["stdlib/runtime/generated/support.aml"].Data), bytes.Clone(resources["stdlib/runtime/generated/support.json"].Data)
	composeContractProvider(t, resources)
	if !bytes.Equal(first, resources["stdlib/runtime/generated/support.aml"].Data) || !bytes.Equal(metadata, resources["stdlib/runtime/generated/support.json"].Data) {
		t.Fatal("unchanged alternate provider regenerated different support")
	}
	if err := Lyger.CheckRuntimeResources(resources); err != nil {
		t.Fatal(err)
	}
	manifest, _, err := distribution.LoadRuntimeComposition(resources)
	if err != nil || manifest.Sources["stdlib/eng/runtime/alternate.l"] == "" || manifest.Sources["stdlib/eng/runtime/dictionary.l"] != "" {
		t.Fatalf("selected source provenance: %v, %#v", err, manifest.Sources)
	}
}

// The consumer uses only production exports, never the bundled implementation's
// private observer adapter, field layouts, injection hooks or extension methods.
func TestReplaceableRuntimeNativeContract(t *testing.T) {
	for _, alternate := range []bool{false, true} {
		name := "default"
		resources := distributiontest.Files(t)
		if alternate {
			name = "alternate"
			resources = distributiontest.AlternateRuntime(t)
			composeContractProvider(t, resources)
		}
		t.Run(name, func(t *testing.T) {
			for _, mode := range []string{"inline", "workspace"} {
				t.Run(mode, func(t *testing.T) {
					files := runtimeContractFiles(t, resources, mode)
					binary := runtimeContractBinary(t, files, "runtime-parity-public")
					want := "fixture.l:7:1: Await requires an invocation\n"
					if alternate {
						want = "Await requires an invocation (source: fixture.l:7:1)\n"
					}
					runtimeExecute(t, binary, nil, []string{"attributed"}, 1, "", want)
					runtimeExecute(t, binary, nil, nil, 1, "", "Await requires an invocation\n")
				})
			}
		})
	}
}

func runtimeContractBinary(t *testing.T, files map[string][]byte, target string) string {
	t.Helper()
	built := nativetest.Build(t, nativetest.Materialize(t, files), nativetest.BuildOptions{Binary: target})
	if built.Err != nil {
		t.Fatalf("runtime contract compilation: %v\n%s\n%s", built.Err, built.Stdout, built.Stderr)
	}
	if len(built.Artifacts) != 1 {
		t.Fatalf("contract build executable inventory: %#v", built.Artifacts)
	}
	binary := built.Artifacts[0].Executable
	nativetest.AssertNativeArtifact(t, binary)
	return binary
}

func TestReplaceableRuntimeRejectsWrongNativeContract(t *testing.T) {
	for _, tc := range []struct {
		name, before, after, diagnostic string
	}{
		{"missing export", "`app_set_source`", "`different_set_source`", "E0425"},
		{"wrong signature", "\"source\" as `String`", "\"source\" as `Vec<u8>`", "E0308"},
		{"narrowed App reference", "\"value\" as `&App` and \"source\" as `String`", "\"value\" as `&std::sync::Arc<App>` and \"source\" as `String`", "E0308"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resources := distributiontest.AlternateRuntime(t)
			const source = "stdlib/eng/runtime/process.l"
			if bytes.Count(resources[source].Data, []byte(tc.before)) != 1 {
				t.Fatal("native contract mutation did not select one declaration")
			}
			resources[source].Data = bytes.Replace(resources[source].Data, []byte(tc.before), []byte(tc.after), 1)
			composeContractProvider(t, resources)
			project := nativetest.Materialize(t, runtimeContractFiles(t, resources, "workspace"))
			built := nativetest.Build(t, project, nativetest.BuildOptions{Binary: "runtime-parity-public"})
			if built.Err == nil || !bytes.Contains(built.Stdout, []byte(tc.diagnostic)) || !bytes.Contains(built.Stdout, []byte("app_set_source")) {
				t.Fatalf("wrong runtime contract did not fail at its native interface: %v\n%s\n%s", built.Err, built.Stdout, built.Stderr)
			}
		})
	}
}

func runtimeContractFiles(t *testing.T, resources fs.FS, mode string) map[string][]byte {
	t.Helper()
	result, err := Lyger.NewLygerService(Lyger.Options{Resources: resources}).Compile([]byte("Show `contract`.\n"), "contract.l")
	if err != nil {
		t.Fatal(err)
	}
	entry, err := result.OrchestratorUnit("rust_runtime_generation", nil)
	if err != nil {
		t.Fatal(err)
	}
	service := NewLogicemeService(resources)
	var bundle Executable
	if mode == "workspace" {
		_, bundle, err = service.EvaluateWorkspaceEntry(entry.BackendAML, WorkspaceRuntime{Package: "contract_runtime", RelativeDirectory: "../runtime"})
	} else {
		bundle, err = service.Evaluate(entry.BackendAML, GenerationOptions{IncludeRuntime: true})
	}
	if err != nil {
		t.Fatal(err)
	}
	generated := workspaceFiles(t, bundle)
	if mode == "inline" {
		// Execute the unchanged returned entry as well as the public-interface
		// observer. Installed-CLI tests exercise the actual workspace entry.
		inline := make(map[string][]byte, len(generated)+1)
		for name, data := range generated {
			inline[strings.TrimPrefix(name, "/")] = bytes.Clone(data)
		}
		lock, readErr := os.ReadFile("testdata/runtime/generation.Cargo.lock")
		if readErr != nil {
			t.Fatal(readErr)
		}
		inline["Cargo.lock"] = lock
		runtimeExecute(t, runtimeContractBinary(t, inline, ""), nil, nil, 0, "contract", "")
	}
	files := map[string][]byte{"src/main.rs": []byte(runtimeContractConsumer)}
	for _, name := range []string{"runtime.rs", "runtime/host.rs", "process_unix.rs", "process_other.rs"} {
		data := generated["/src/"+name]
		if len(data) == 0 {
			t.Fatalf("%s assembly omitted %s", mode, name)
		}
		files["src/"+name] = bytes.Clone(data)
	}
	for _, name := range []string{"Cargo.toml", "Cargo.lock"} {
		data, readErr := os.ReadFile("testdata/runtime/" + name)
		if readErr != nil {
			t.Fatal(readErr)
		}
		files[name] = data
	}
	return files
}

const runtimeContractConsumer = `mod runtime;
#[path = "process_unix.rs"] mod process;
use std::sync::Arc;

// Type-check the complete compiler contract, including exports not needed by
// the observation below. Futures may borrow their inputs. No private fields or
// policy are reconstructed here.
#[allow(dead_code)]
async fn surface(value: &runtime::Runtime, app: &Arc<runtime::App>, invocation: &Arc<runtime::Invocation>, snapshot: &runtime::Snapshot, terminal: &runtime::Terminal, formatter: &mut std::fmt::Formatter<'_>) -> runtime::Result<()> {
    let app_ref: &runtime::App = app.as_ref();
    let _: runtime::Runtime = runtime::runtime_copy(value);
    let error: runtime::Error = runtime::error_new(vec![b'x']);
    let _: runtime::Error = runtime::error_new_text("x");
    let _: runtime::Error = runtime::error_copy(&error);
    let _: std::fmt::Result = runtime::error_display(&error, formatter);
    let _: Option<&(dyn std::error::Error + 'static)> = runtime::error_native_cause(&error);
    let _: runtime::Error = runtime::error_from_io(std::io::Error::other("x"));
    let _: runtime::Snapshot = runtime::snapshot_copy(snapshot);
    let _: runtime::Terminal = runtime::terminal_copy(terminal);
    let _: Arc<runtime::App> = runtime::runtime_app(value, "unused".into());
    let _: () = runtime::app_set_source(app_ref, "source".into());
    let _: Arc<[u8]> = runtime::app_output(app_ref)?;
    let _: Option<runtime::Error> = runtime::app_error(app_ref)?;
    fn send<F: std::future::Future<Output = runtime::Result<()>> + Send>(future: F) -> F { future }
    let _: () = send(runtime::runtime_show(value, b"bytes")).await?;
    let _: runtime::Error = runtime::runtime_fail(value, Some(error));
    let _: Vec<u8> = runtime::format_number(3.5);
    let _: () = runtime::app_build(app_ref)?;
    let _: Arc<runtime::Invocation> = runtime::app_run(app)?;
    let stopping: std::pin::Pin<Box<dyn std::future::Future<Output = runtime::Result<()>> + Send + '_>> = runtime::app_stop(app_ref);
    let _: () = stopping.await?;
    let waiting: std::pin::Pin<Box<dyn std::future::Future<Output = runtime::Result<Arc<runtime::Invocation>>> + Send + '_>> = runtime::app_wait(app_ref);
    let _: Arc<runtime::Invocation> = waiting.await?;
    let _: bool = runtime::invocation_claim(invocation.as_ref())?;
    let _: () = runtime::invocation_finish(invocation);
    Ok(())
}

fn main() -> std::process::ExitCode {
    let executor = tokio::runtime::Builder::new_current_thread().enable_all().build().unwrap();
    let status: u8 = executor.block_on(async {
        let value: runtime::Runtime = runtime::runtime_new().unwrap();
        let app = runtime::runtime_app(&value, "unused-companion".into());
        if std::env::args().len() > 1 {
            runtime::app_set_source(&app, "fixture.l:7:1".into());
        }
        runtime::runtime_run(&value, async { runtime::app_wait(&app).await?; Ok(()) }).await
    });
    executor.shutdown_background();
    std::process::ExitCode::from(status)
}
`
