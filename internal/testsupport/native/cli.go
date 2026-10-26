package native

// This observer is test infrastructure. It records the real Cargo compiler and
// never supplies a generated file, executable, or alternative lowering path.
import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

var recorderOnce sync.Once
var recorderPath string
var recorderError error

func artifactRecorder(t testing.TB) string {
	t.Helper()
	path, err := artifactRecorderResult()
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func artifactRecorderResult() (string, error) {
	recorderOnce.Do(func() {
		dir, err := os.MkdirTemp("", "l-rust-provenance-")
		if err != nil {
			recorderError = err
			return
		}
		source := filepath.Join(dir, "main.go")
		recorderPath = filepath.Join(dir, "record")
		if err = os.WriteFile(source, []byte(artifactRecorderSource), 0600); err != nil {
			recorderError = err
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), BuildPrerequisiteTimeout)
		defer cancel()
		cmd := exec.CommandContext(ctx, "go", "build", "-o", recorderPath, source)
		cmd.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOWORK=off", "GOENV=off", "GOFLAGS=")
		output, err := cmd.CombinedOutput()
		if err != nil {
			recorderError = fmt.Errorf("compile provenance observer: %w: %s", err, output)
		}
	})
	return recorderPath, recorderError
}

type CLIRecorder struct {
	Directory string
	tools     ToolchainInfo
	before    map[string]bool
}

func CLIProbe(t testing.TB) *CLIRecorder {
	t.Helper()
	r := &CLIRecorder{Directory: t.TempDir(), tools: Toolchain(t)}
	// The package-wide native recorder executes real Cargo and snapshots its
	// provenance before returning. Each probe supplies private configuration;
	// each invocation creates its own record. Reusing the executable avoids a
	// fresh shell/mktemp process chain without caching any build outcome.
	runID, err := evidenceRunID()
	if err != nil {
		runID = os.Getenv(evidenceRunEnvironment)
	}
	config, err := json.Marshal(struct{ Directory, Cargo, RunID, Test, Host string }{r.Directory, r.tools.Cargo, runID, t.Name(), r.tools.Host})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(r.Directory, "cargo-proxy.json"), config, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(artifactRecorder(t), filepath.Join(r.Directory, "cargo")); err != nil {
		t.Fatal(err)
	}
	goTool := "#!/bin/sh\nprintf '%s\\n' 'Rust parity: generated-program build attempted the Go tool' >&2\nexit 97\n"
	if err := os.WriteFile(filepath.Join(r.Directory, "go"), []byte(goTool), 0700); err != nil {
		t.Fatal(err)
	}
	return r
}

func (r *CLIRecorder) Env(env []string) []string {
	// A new successful build must produce fresh Cargo evidence, even when this
	// recorder retains earlier immutable generations for publication checks.
	r.before = map[string]bool{}
	previous, err := filepath.Glob(filepath.Join(r.Directory, "cargo.*.jsonl"))
	if err != nil {
		panic(fmt.Errorf("invalid Cargo evidence inventory pattern: %w", err))
	}
	for _, path := range previous {
		r.before[strings.TrimSuffix(path, ".jsonl")] = true
	}
	if env == nil {
		env = os.Environ()
	}
	var result []string
	path := os.Getenv("PATH")
	for _, item := range env {
		if strings.HasPrefix(item, "PATH=") {
			path = strings.TrimPrefix(item, "PATH=")
			continue
		}
		key := strings.SplitN(item, "=", 2)[0]
		if key == cargoProxyConfigEnvironment || key == "RUSTC" || key == "RUSTC_WRAPPER" || key == "RUSTC_WORKSPACE_WRAPPER" || key == "RUSTFLAGS" || key == "RUSTDOCFLAGS" || key == "CARGO_ENCODED_RUSTFLAGS" || key == "RUSTUP_TOOLCHAIN" || key == "CARGO_BUILD_TARGET" || key == "CARGO_TARGET_DIR" || key == "CARGO_NET_OFFLINE" || strings.HasPrefix(key, "CARGO_TARGET_") || strings.HasPrefix(key, "CARGO_PROFILE_") {
			continue
		}
		result = append(result, item)
	}
	return append(result, cargoProxyConfigEnvironment+"="+filepath.Join(r.Directory, "cargo-proxy.json"), "PATH="+r.Directory+string(os.PathListSeparator)+path, "RUSTUP_TOOLCHAIN="+Version, "RUSTC="+r.tools.Rustc, "CARGO_NET_OFFLINE=true")
}

func (r *CLIRecorder) artifacts(t testing.TB) []string {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(r.Directory, "cargo.*.artifacts", "*"))
	if err != nil {
		t.Fatal(err)
	}
	return paths
}

func (r *CLIRecorder) freshArtifacts(t testing.TB) []string {
	t.Helper()
	var found []string
	for _, path := range r.artifacts(t) {
		base := strings.TrimSuffix(filepath.Dir(path), ".artifacts")
		if !r.before[base] {
			found = append(found, path)
		}
	}
	if len(found) == 0 {
		t.Fatal("Rust build provenance: this build produced no fresh Cargo binary artifacts; runtime not exercised")
	}
	return found
}

func (r *CLIRecorder) AssertPublished(t testing.TB, entry string) {
	t.Helper()
	r.Workspace(t)
	fresh := r.freshArtifacts(t)
	paths := r.artifacts(t)
	if len(paths) == 0 {
		t.Fatal("Rust build provenance: no Cargo executable artifacts; runtime not exercised")
	}
	// The immutable generation contract is shared with the Go baseline. Do not
	// prescribe Cargo member names or use an arbitrary target-directory guess.
	targets := []string{entry}
	err := filepath.WalkDir(entry+".apps", func(path string, e os.DirEntry, err error) error {
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if e.IsDir() || e.Name() != "manifest.json" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var manifest struct {
			Apps []struct{ Executable, SHA256 string }
		}
		if err = json.Unmarshal(data, &manifest); err != nil {
			return err
		}
		for _, app := range manifest.Apps {
			p := filepath.Join(filepath.Dir(path), app.Executable)
			b, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			if fmt.Sprintf("%x", sha256.Sum256(b)) != app.SHA256 {
				return fmt.Errorf("published App hash mismatch: %s", p)
			}
			targets = append(targets, p)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, target := range targets {
		AssertNativeArtifact(t, target)
		data, err := os.ReadFile(target)
		if err != nil {
			t.Fatal(err)
		}
		want := sha256.Sum256(data)
		matched := false
		candidates := paths
		if target == entry {
			candidates = fresh
		}
		for _, artifact := range candidates {
			got, err := os.ReadFile(artifact)
			if err == nil && sha256.Sum256(got) == want {
				matched = true
				break
			}
		}
		if !matched {
			t.Fatalf("published executable %s does not match any Cargo executable artifact", target)
		}
	}
}

// Runs synchronously before Cargo returns to the CLI, which may delete its
// temporary build directory. Copies are evidence, never build inputs.
const artifactRecorderSource = `package main
import("bufio";"bytes";"context";"crypto/rand";"crypto/sha256";"encoding/json";"fmt";"io";"os";"os/exec";"path/filepath";"strings";"syscall";"time")
func main(){if len(os.Args)>1&&strings.HasPrefix(os.Args[1],"__evidence_"){os.Exit(evidenceCommand(os.Args[1:]))};if filepath.Base(os.Args[0])=="cargo"{os.Exit(proxyCargo(os.Args[1:]))};if len(os.Args)>1&&os.Args[1]=="__timed_cargo"{os.Exit(timedCargo(os.Args[2:]))};if err:=record(os.Args[1]);err!=nil{fmt.Fprintln(os.Stderr,err);os.Exit(96)}}
func record(base string) error{
 f,err:=os.Open(base+".jsonl");if err!=nil{return err};defer f.Close()
 scan:=bufio.NewScanner(f);scan.Buffer(make([]byte,4096),8*1024*1024)
 for scan.Scan(){ var m struct{Reason,Executable string;Target struct{Kind []string}};if json.Unmarshal(scan.Bytes(),&m)!=nil||m.Reason!="compiler-artifact"||m.Executable==""{continue}; bin:=false;for _,kind:=range m.Target.Kind{if kind=="bin"{bin=true}};if !bin{continue}
 b,err:=os.ReadFile(m.Executable);if err!=nil{return err};dir:=base+".artifacts";if err=os.MkdirAll(dir,0700);err!=nil{return err}
 if err=os.WriteFile(filepath.Join(dir,fmt.Sprintf("%x",sha256.Sum256(b))),b,0700);err!=nil{return err}
 }
 if err=scan.Err();err!=nil{return err}
 metadata,err:=os.ReadFile(base+".metadata.json");if err!=nil{return err}
 var info map[string]json.RawMessage;var cwd string;if json.Unmarshal(metadata,&info)!=nil||json.Unmarshal(info["workspace_root"],&cwd)!=nil||cwd==""{return nil}
 return filepath.WalkDir(cwd,func(path string,e os.DirEntry,err error)error{
 if err!=nil{return err};rel,err:=filepath.Rel(cwd,path);if err!=nil{return err}
 if e.IsDir(){if e.Name()=="target"||e.Name()==".git"{return filepath.SkipDir};return nil}
 if strings.HasSuffix(rel,".rs")||e.Name()=="Cargo.toml"||e.Name()=="Cargo.lock"{
 b,err:=os.ReadFile(path);if err!=nil{return err};dest:=filepath.Join(base+".workspace",rel);if err=os.MkdirAll(filepath.Dir(dest),0700);err!=nil{return err};return os.WriteFile(dest,b,0600)
 };return nil
 })
}
` + artifactTimingSource + artifactProxySource + artifactEvidenceSource

// Workspace returns native Cargo metadata, not a test-authored directory model.
func (r *CLIRecorder) Workspace(t testing.TB) {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(r.Directory, "cargo.*.metadata.json"))
	if err != nil {
		t.Fatal(err)
	}
	valid := false
	for _, path := range files {
		if r.before[strings.TrimSuffix(path, ".metadata.json")] {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var m struct {
			WorkspaceRoot string   `json:"workspace_root"`
			Members       []string `json:"workspace_members"`
			Packages      []struct {
				ID      string
				Targets []struct {
					Name string
					Kind []string
				}
				Dependencies []struct{ Name string }
			}
		}
		if err = json.Unmarshal(data, &m); err != nil {
			continue
		}
		if m.WorkspaceRoot == "" || len(m.Members) == 0 {
			continue
		}
		ids := map[string]bool{}
		bins := 0
		for _, p := range m.Packages {
			if p.ID == "" || ids[p.ID] {
				t.Fatalf("Cargo returned missing/duplicate package identity: %q", p.ID)
			}
			ids[p.ID] = true
			seen := map[string]bool{}
			for _, target := range p.Targets {
				key := target.Name + strings.Join(target.Kind, ",")
				if target.Name == "" || seen[key] {
					t.Fatalf("Cargo returned invalid target identity %q", key)
				}
				seen[key] = true
				for _, kind := range target.Kind {
					if kind == "bin" {
						bins++
					}
				}
			}
		}
		for _, id := range m.Members {
			if !ids[id] {
				t.Fatalf("workspace member %s has no Cargo package", id)
			}
		}
		if bins == 0 {
			t.Fatal("Cargo workspace has no binary target")
		}
		valid = true
	}
	if !valid {
		t.Fatal("Rust workspace prerequisite: no valid Cargo metadata from actual build")
	}
}
