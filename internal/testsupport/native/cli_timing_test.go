package native

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestTimedCargoRecorderPreservesStreamsStatusAndEvidence(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		for _, status := range []int{0, 7} {
			t.Run(fmt.Sprintf("enabled_%t_status_%d", enabled, status), func(t *testing.T) {
				directory := t.TempDir()
				trace := filepath.Join(directory, "trace")
				if enabled {
					t.Setenv("L_HARNESS_TIMING_DIR", trace)
				} else {
					t.Setenv("L_HARNESS_TIMING_DIR", "")
				}
				t.Setenv("L_TIMING_TEST_STATUS", fmt.Sprint(status))
				t.Setenv("L_TIMING_TEST_ROOT", directory)
				cargo := controlledCargo(t)
				t.Setenv("L_CONTROLLED_CARGO_MODE", "timing")
				if err := os.Mkdir(filepath.Join(directory, "workspace"), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(directory, "workspace", "Cargo.toml"), []byte("authored manifest"), 0600); err != nil {
					t.Fatal(err)
				}
				record := filepath.Join(directory, "cargo-record")
				cmd := exec.Command(artifactRecorder(t), "__timed_cargo", record, cargo, "build", "--manifest-path", filepath.Join(directory, "workspace", "Cargo.toml"), "--target-dir", filepath.Join(directory, "target"))
				var out, diag bytes.Buffer
				cmd.Stdout = &out
				cmd.Stderr = &diag
				err := cmd.Run()
				gotStatus := 0
				if err != nil {
					if exit, ok := err.(*exec.ExitError); ok {
						gotStatus = exit.ExitCode()
					} else {
						t.Fatal(err)
					}
				}
				if gotStatus != status {
					t.Fatalf("status=%d want%d: %v stderr=%s", gotStatus, status, err, diag.Bytes())
				}
				wantOut := "{\"reason\":\"compiler-artifact\",\"fresh\":true}\n{\"reason\":\"compiler-artifact\",\"fresh\":false}\n"
				if out.String() != wantOut || diag.String() != "native stderr\x00tail\n" {
					t.Fatalf("streams=%q/%q", out.Bytes(), diag.Bytes())
				}
				data, err := os.ReadFile(record + ".jsonl")
				if err != nil || string(data) != wantOut {
					t.Fatalf("compiler evidence=%q %v", data, err)
				}
				data, err = os.ReadFile(filepath.Join(record+".workspace", "Cargo.toml"))
				if err != nil || string(data) != "authored manifest" {
					t.Fatalf("workspace evidence=%q %v", data, err)
				}
				args, err := os.ReadFile(filepath.Join(directory, "args"))
				if err != nil {
					t.Fatal(err)
				}
				if strings.Contains(string(args), "--timings") != enabled {
					t.Fatalf("diagnostic argument=%q", args)
				}
				if !enabled {
					if _, err := os.Stat(trace); !os.IsNotExist(err) {
						t.Fatalf("disabled trace created: %v", err)
					}
					return
				}
				paths, err := filepath.Glob(filepath.Join(trace, "*.jsonl"))
				if err != nil || len(paths) != 1 {
					t.Fatalf("trace files=%v %v", paths, err)
				}
				data, err = os.ReadFile(paths[0])
				if err != nil {
					t.Fatal(err)
				}
				phases := map[string]bool{}
				for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
					var event struct {
						Event, Phase, Outcome string
						Fields                map[string]any
					}
					if err := json.Unmarshal([]byte(line), &event); err != nil {
						t.Fatal(err)
					}
					if event.Event == "finish" {
						phases[event.Phase] = true
						if event.Phase == "cargo.actual" && (event.Outcome == "error") != (status != 0) {
							t.Fatalf("wrong Cargo result %+v", event)
						}
					}
					if event.Phase == "cargo.artifacts" && (event.Fields["fresh"] != float64(1) || event.Fields["rebuilt"] != float64(1)) {
						t.Fatalf("wrong stats %+v", event)
					}
				}
				for _, phase := range []string{"cargo.actual", "cargo.metadata", "cargo.recorder"} {
					if !phases[phase] {
						t.Fatalf("missing phase %s: %s", phase, data)
					}
				}
				reports, err := filepath.Glob(filepath.Join(trace, "cargo-timings", "*.html"))
				if err != nil || len(reports) != 1 {
					t.Fatalf("timing reports=%v %v", reports, err)
				}
			})
		}
	}
}

func TestTimedCargoRecorderPreservesSignalAndInput(t *testing.T) {
	for _, tc := range []struct {
		name, input, stdout, stderr string
		status                      int
	}{
		{name: "signal", stderr: "native signal stderr\x00tail\n", status: 143},
		{name: "stdin", input: "supplied stdin\x00tail\n", stdout: "supplied stdin\x00tail\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			directory := t.TempDir()
			t.Setenv("L_HARNESS_TIMING_DIR", filepath.Join(directory, "trace"))
			cargo := controlledCargo(t)
			t.Setenv("L_CONTROLLED_CARGO_MODE", tc.name)
			record := filepath.Join(directory, "record")
			cmd := exec.Command(artifactRecorder(t), "__timed_cargo", record, cargo, "build")
			cmd.Stdin = strings.NewReader(tc.input)
			var out, diag bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &diag
			err := cmd.Run()
			status := 0
			if err != nil {
				if exit, ok := err.(*exec.ExitError); ok {
					status = exit.ExitCode()
				} else {
					t.Fatal(err)
				}
			}
			if status != tc.status || out.String() != tc.stdout || diag.String() != tc.stderr {
				t.Fatalf("signal/input contract: status=%d stdout=%q stderr=%q; want%d %q %q", status, out.Bytes(), diag.Bytes(), tc.status, tc.stdout, tc.stderr)
			}
			data, err := os.ReadFile(record + ".jsonl")
			if err != nil || string(data) != tc.stdout {
				t.Fatalf("recorded native output=%q %v", data, err)
			}
		})
	}
}
