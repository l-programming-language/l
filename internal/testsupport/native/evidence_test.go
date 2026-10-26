package native

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestBundleEvidenceHasImmutableRunIdentity(t *testing.T) {
	root := t.TempDir()
	t.Setenv("L_HARNESS_EVIDENCE_DIR", root)
	t.Setenv("L_HARNESS_RUN_ID", "independent-run")
	RecordBundle(t, "generation", map[string]any{"empty": []string{}, "nil": []string(nil), "native": "a\x00b"})
	RecordBundle(t, "generation", map[string]any{"later": true})
	var records []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.HasSuffix(path, ".record.json") {
			records = append(records, path)
		}
		return nil
	})
	if err != nil || len(records) != 2 {
		t.Fatalf("write-once records = %v: %v", records, err)
	}
	for _, path := range records {
		var record struct {
			RunID, Path, SHA256 string
			Bytes               int
		}
		data, err := os.ReadFile(path)
		if err != nil || json.Unmarshal(data, &record) != nil {
			t.Fatalf("record %s: %s / %v", path, data, err)
		}
		if record.RunID != "independent-run" || !strings.Contains(record.Path, string(filepath.Separator)+"independent-run"+string(filepath.Separator)) {
			t.Fatalf("wrong run identity: %+v", record)
		}
		data, err = os.ReadFile(record.Path)
		if err != nil || record.SHA256 != hexDigest(data) || record.Bytes != len(data) {
			t.Fatalf("unpinned bytes: %+v / %v", record, err)
		}
	}
}

func hexDigest(data []byte) string {
	const hex = "0123456789abcdef"
	sum := sha256.Sum256(data)
	out := make([]byte, 64)
	for i, b := range sum {
		out[i*2], out[i*2+1] = hex[b>>4], hex[b&15]
	}
	return string(out)
}

func TestCLIProxyRetainsPreCargoInputsAndRawResults(t *testing.T) {
	t.Setenv("L_HARNESS_TIMING_DIR", "")
	root := t.TempDir()
	t.Setenv("L_HARNESS_CARGO_EVIDENCE_DIR", root)
	t.Setenv("L_HARNESS_RUN_ID", "cargo-evidence-run")
	t.Setenv("UNRELATED_SECRET_FOR_EVIDENCE", "never record this value or name")
	t.Setenv("CARGO_REGISTRIES_PRIVATE_TOKEN", "never record this token")
	f := newCLIProxyFixture(t)
	if err := os.WriteFile(filepath.Join(f.workspace, "included.data"), []byte("not just Rust files\x00"), 0600); err != nil {
		t.Fatal(err)
	}
	out, diag, status := f.invoke([]string{"+" + Version, "build", "--manifest-path", filepath.Join(f.workspace, "Cargo.toml")}, "native input\x00\n", "L_PROXY_STATUS=7")
	if status != 7 || string(out) != f.output+"native input\x00\n" || string(diag) != "native stderr\x00tail\n" {
		t.Fatalf("native result altered: %d %q %q", status, out, diag)
	}
	paths, callErr := filepath.Glob(filepath.Join(root, "cargo-evidence-run", "invocations", "*", "complete.json"))
	if callErr != nil {
		t.Fatal(callErr)
	}
	if len(paths) != 1 {
		t.Fatalf("complete native evidence = %v", paths)
	}
	dir := filepath.Dir(paths[0])
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, secret := range []string{"UNRELATED_SECRET_FOR_EVIDENCE", "never record this value or name", "CARGO_REGISTRIES_PRIVATE_TOKEN", "never record this token", hexDigest([]byte("never record this token"))} {
			if strings.Contains(string(data), secret) {
				t.Errorf("unexpected environment evidence in %s: %s", path, secret)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	var complete struct {
		Complete bool
		Status   int
		Errors   []string
	}
	data, callErr := os.ReadFile(paths[0])
	if callErr != nil {
		t.Fatal(callErr)
	}
	if err := json.Unmarshal(data, &complete); err != nil || !complete.Complete || complete.Status != 7 {
		t.Fatalf("completion = %s / %v", data, err)
	}
	for path, want := range map[string][]byte{"inputs/included.data": []byte("not just Rust files\x00"), "cargo.stdout": out, "cargo.stderr": diag} {
		data, err := os.ReadFile(filepath.Join(dir, path))
		if err != nil || string(data) != string(want) {
			t.Fatalf("%s = %q / %v", path, data, err)
		}
	}
	if err := os.WriteFile(filepath.Join(f.workspace, "included.data"), []byte("later mutation"), 0600); err != nil {
		t.Fatal(err)
	}
	data, err = os.ReadFile(filepath.Join(dir, "inputs", "included.data"))
	if err != nil || string(data) != "not just Rust files\x00" {
		t.Fatalf("evidence changed: %q / %v", data, err)
	}
}

func TestBundleEvidenceSurvivesIndependentProcesses(t *testing.T) {
	if os.Getenv("L_BUNDLE_EVIDENCE_CHILD") == "yes" {
		RecordBundle(t, "same-phase", map[string]string{"value": "exact bytes"})
		return
	}
	root := t.TempDir()
	t.Setenv("L_HARNESS_EVIDENCE_DIR", root)
	t.Setenv("L_HARNESS_RUN_ID", "shared-run")
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cmd := exec.Command(executable, "-test.run=^TestBundleEvidenceSurvivesIndependentProcesses$")
			cmd.Env = append(os.Environ(), "L_BUNDLE_EVIDENCE_CHILD=yes")
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Errorf("child: %v / %s", err, out)
			}
		}()
	}
	wg.Wait()
	paths, callErr := filepath.Glob(filepath.Join(root, "shared-run", "bundles", "*", "*.record.json"))
	if callErr != nil {
		t.Fatal(callErr)
	}
	if len(paths) != 2 {
		t.Fatalf("independent process records: %v", paths)
	}
}

func TestCargoEvidenceDisabledAndIncompleteCapture(t *testing.T) {
	t.Run("disabled", func(t *testing.T) {
		t.Setenv(cargoEvidenceEnvironment, "")
		capture := beginCargoEvidence(t, cargoEvidenceInput{Workspace: "/does/not/exist"})
		capture.Finish(nil, nil, nil)
		if capture.directory != "" || capture.recorder != "" {
			t.Fatal("disabled capture initialized a recorder")
		}
	})
	for _, kind := range []string{"changed-source", "symlink", "oversize", "evidence-root-file", "invalid-run"} {
		t.Run(kind, func(t *testing.T) {
			root := t.TempDir()
			t.Setenv(cargoEvidenceEnvironment, root)
			t.Setenv(evidenceRunEnvironment, "direct-run")
			f := newCLIProxyFixture(t)
			if kind == "symlink" {
				if err := os.Symlink("src/main.rs", filepath.Join(f.workspace, "source-link")); err != nil {
					t.Fatal(err)
				}
			}
			if kind == "oversize" {
				f, err := os.Create(filepath.Join(f.workspace, "oversized.data"))
				if err != nil {
					t.Fatal(err)
				}
				err = f.Truncate((64 << 20) + 1)
				closeErr := f.Close()
				if err != nil || closeErr != nil {
					t.Fatalf("bounded input setup: %v / %v", err, closeErr)
				}
			}
			if kind == "evidence-root-file" {
				path := filepath.Join(root, "file")
				if err := os.WriteFile(path, []byte("occupied"), 0600); err != nil {
					t.Fatal(err)
				}
				t.Setenv(cargoEvidenceEnvironment, path)
			}
			if kind == "invalid-run" {
				t.Setenv(evidenceRunEnvironment, "../escape")
			}
			capture := beginCargoEvidence(t, cargoEvidenceInput{Route: "controlled-direct", Command: []string{controlledCargo(t), "build"}, Directory: f.workspace, Workspace: f.workspace, Manifest: filepath.Join(f.workspace, "Cargo.toml"), Environment: f.env})
			if kind == "changed-source" {
				if err := os.WriteFile(filepath.Join(f.workspace, "src/main.rs"), []byte("changed after snapshot"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			capture.Finish([]byte(f.output), []byte("native diagnostic"), nil)
			capture.Finish([]byte("must not overwrite"), nil, nil)
			if kind == "evidence-root-file" || kind == "invalid-run" {
				if capture.directory != "" {
					t.Fatalf("invalid setup created capture directory: %s", capture.directory)
				}
				return
			}
			data, err := os.ReadFile(filepath.Join(capture.directory, "complete.json"))
			if err != nil {
				t.Fatal(err)
			}
			var complete struct {
				Complete bool
				Errors   []string
			}
			if err = json.Unmarshal(data, &complete); err != nil || complete.Complete || len(complete.Errors) == 0 {
				t.Fatalf("incomplete result: %s / %v", data, err)
			}
			data, err = os.ReadFile(filepath.Join(capture.directory, "cargo.stdout"))
			if err != nil || string(data) != f.output {
				t.Fatalf("raw output overwritten: %q / %v", data, err)
			}
		})
	}
}

func TestCLIProxyEvidenceConcurrentInputIdentity(t *testing.T) {
	root := t.TempDir()
	t.Setenv(cargoEvidenceEnvironment, root)
	t.Setenv(evidenceRunEnvironment, "concurrent-run")
	a, b := newCLIProxyFixture(t), newCLIProxyFixture(t)
	if err := os.WriteFile(filepath.Join(b.workspace, "src/main.rs"), []byte("different B source"), 0600); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for _, fixture := range []cliProxyFixture{a, b, a} {
		wg.Add(1)
		go func(f cliProxyFixture) {
			defer wg.Done()
			out, diag, status := f.invoke([]string{"build", "--manifest-path", filepath.Join(f.workspace, "Cargo.toml")}, "")
			if status != 0 || string(out) != f.output || string(diag) != "native stderr\x00tail\n" {
				t.Errorf("concurrent capture changed native result: %d / %q / %q", status, out, diag)
			}
		}(fixture)
	}
	wg.Wait()
	paths, callErr := filepath.Glob(filepath.Join(root, "concurrent-run", "invocations", "*", "complete.json"))
	if callErr != nil {
		t.Fatal(callErr)
	}
	if len(paths) != 3 {
		t.Fatalf("completion records: %v", paths)
	}
	counts := map[string]int{}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var value struct {
			Complete    bool
			InputDigest string
		}
		if err = json.Unmarshal(data, &value); err != nil || !value.Complete {
			t.Fatalf("incomplete %s: %s / %v", path, data, err)
		}
		counts[value.InputDigest]++
	}
	if len(counts) != 2 {
		t.Fatalf("A/B/A content identities: %v", counts)
	}
}

func TestCLIProxyEvidenceFailurePreservesNativeResult(t *testing.T) {
	root := t.TempDir()
	occupied := filepath.Join(root, "occupied")
	if err := os.WriteFile(occupied, []byte("file"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv(cargoEvidenceEnvironment, occupied)
	f := newCLIProxyFixture(t)
	out, diag, status := f.invoke([]string{"build", "--manifest-path", filepath.Join(f.workspace, "Cargo.toml")}, "input\x00", "L_PROXY_STATUS=7")
	if status != 7 || string(out) != f.output+"input\x00" || string(diag) != "native stderr\x00tail\n" {
		t.Fatalf("capture failure changed native result: %d / %q / %q", status, out, diag)
	}
}

func TestCLIProxyEvidenceUsesCompleteWorkspaceAndMetadataStatus(t *testing.T) {
	for _, metadataFailure := range []bool{false, true} {
		t.Run(map[bool]string{false: "member", true: "metadata_nonzero"}[metadataFailure], func(t *testing.T) {
			root := t.TempDir()
			t.Setenv(cargoEvidenceEnvironment, root)
			t.Setenv(evidenceRunEnvironment, "workspace-run")
			f := newCLIProxyFixture(t)
			member := filepath.Join(f.workspace, "member", "Cargo.toml")
			if err := os.Mkdir(filepath.Dir(member), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(member, []byte("selected member manifest"), 0600); err != nil {
				t.Fatal(err)
			}
			extra := []string{}
			if metadataFailure {
				extra = append(extra, "L_PROXY_METADATA_ERROR=valid-json")
			}
			out, diag, status := f.invoke([]string{"build", "--manifest-path", member}, "", extra...)
			if status != 0 || string(out) != f.output || string(diag) != "native stderr\x00tail\n" {
				t.Fatalf("native result changed: %d / %q / %q", status, out, diag)
			}
			paths, callErr := filepath.Glob(filepath.Join(root, "workspace-run", "invocations", "*", "complete.json"))
			if callErr != nil {
				t.Fatal(callErr)
			}
			if len(paths) != 1 {
				t.Fatalf("completion inventory %v", paths)
			}
			data, err := os.ReadFile(paths[0])
			if err != nil {
				t.Fatal(err)
			}
			var record struct {
				Complete bool
				Errors   []string
			}
			if err = json.Unmarshal(data, &record); err != nil || record.Complete == metadataFailure {
				t.Fatalf("metadata completion: %s / %v", data, err)
			}
			data, err = os.ReadFile(filepath.Join(filepath.Dir(paths[0]), "inputs", "src", "main.rs"))
			if err != nil || string(data) != "independent native source" {
				t.Fatalf("workspace capture lost root input: %q / %v", data, err)
			}
		})
	}
}

func TestCargoEvidenceCompletionIsWriteOnce(t *testing.T) {
	root := t.TempDir()
	trace := filepath.Join(t.TempDir(), "trace")
	t.Setenv(cargoEvidenceEnvironment, root)
	t.Setenv(evidenceRunEnvironment, "write-once-run")
	t.Setenv("L_HARNESS_TIMING_DIR", trace)
	f := newCLIProxyFixture(t)
	capture := beginCargoEvidence(t, cargoEvidenceInput{Route: "controlled-direct", Command: []string{controlledCargo(t), "build"}, Directory: f.workspace, Workspace: f.workspace, Manifest: filepath.Join(f.workspace, "Cargo.toml"), Environment: f.env})
	path := filepath.Join(capture.directory, "complete.json")
	if err := os.WriteFile(path, []byte("reserved existing completion"), 0600); err != nil {
		t.Fatal(err)
	}
	capture.Finish([]byte(f.output), nil, nil)
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "reserved existing completion" {
		t.Fatalf("completion overwritten: %q / %v", data, err)
	}
	paths, callErr := filepath.Glob(filepath.Join(trace, "*.jsonl"))
	if callErr != nil {
		t.Fatal(callErr)
	}
	found := false
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range strings.Split(string(data), "\n") {
			var event struct {
				Phase  string
				Fields map[string]any
			}
			if json.Unmarshal([]byte(line), &event) != nil || event.Phase != "cargo.evidence" {
				continue
			}
			if complete, ok := event.Fields["Complete"]; ok {
				found = true
				if complete != false {
					t.Errorf("failed completion write advertised success: %s", line)
				}
			}
		}
	}
	if !found {
		t.Fatal("missing incomplete completion trace")
	}
}

func TestCLIProxyEvidenceRetainsActualCacheAdmission(t *testing.T) {
	root := t.TempDir()
	trace := filepath.Join(t.TempDir(), "trace")
	t.Setenv(cargoEvidenceEnvironment, root)
	t.Setenv(evidenceRunEnvironment, "slot-run")
	t.Setenv("L_HARNESS_TIMING_DIR", trace)
	f := newCLIProxyFixture(t)
	target := filepath.Join(t.TempDir(), "selected", "target")
	if err := os.MkdirAll(trace, 0700); err != nil {
		t.Fatal(err)
	}
	// This is a controlled observer input, not a claim about a real cache.
	event := map[string]any{"phase": "cache.slot", "subject": f.workspace, "started_at": time.Now().UTC(), "fields": map[string]any{"target": target, "preferred_slot": 1, "selected_slot": 5}}
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(trace, fmt.Sprint(os.Getpid())+".jsonl"), append(data, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	_, _, status := f.invoke([]string{"build", "--manifest-path", filepath.Join(f.workspace, "Cargo.toml"), "--target-dir", target}, "")
	if status != 0 {
		t.Fatalf("controlled Cargo status: %d", status)
	}
	paths, callErr := filepath.Glob(filepath.Join(root, "slot-run", "invocations", "*", "start.json"))
	if callErr != nil {
		t.Fatal(callErr)
	}
	if len(paths) != 1 {
		t.Fatalf("start records: %v", paths)
	}
	data, err = os.ReadFile(paths[0])
	if err != nil {
		t.Fatal(err)
	}
	var record struct {
		Input                struct{ PreferredSlot, SelectedSlot string }
		CacheSlotAttribution struct{ Available bool }
	}
	if err = json.Unmarshal(data, &record); err != nil || !record.CacheSlotAttribution.Available || record.Input.PreferredSlot != "1" || record.Input.SelectedSlot != "5" {
		t.Fatalf("actual slot attribution: %s / %v", data, err)
	}
}
