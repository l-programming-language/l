// Package native contains test-owned native observers. It never lowers L,
// rewrites generated source, or implements application lifecycle behavior.
package native

import (
	"bytes"
	"context"
	"crypto/sha256"
	"debug/buildinfo"
	"debug/elf"
	"debug/macho"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/l-programming-language/l/internal/cargocache"
	"github.com/l-programming-language/l/internal/timing"
)

const Version = "1.98.1"

type ToolchainInfo struct{ Cargo, Rustc, Host string }
type BuildOptions struct {
	Binary            string
	TestConfiguration bool
	Tests             bool
	TargetDir         string
	Timeout           time.Duration
}
type Artifact struct {
	// Executable is an immutable test-owned copy of Cargo's artifact. The
	// unmodified JSON stream retains Cargo's original path and provenance.
	PackageID, TargetName, Executable string
	TargetKinds                       []string
}

func defaultTargetDirectory(t testing.TB, host string) string {
	t.Helper()
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate Rust parity dependency cache")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(source), "..", "..", ".."))
	// The first calling test file identifies the Go package even when a test
	// temporarily changes its working directory or uses shared helper layers.
	var callers [64]uintptr
	frames := runtime.CallersFrames(callers[:runtime.Callers(2, callers[:])])
	var packageDir string
	for {
		frame, more := frames.Next()
		if strings.HasSuffix(frame.File, "_test.go") {
			packageDir = filepath.Dir(frame.File)
			break
		}
		if !more {
			break
		}
	}
	if packageDir == "" {
		var err error
		packageDir, err = os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
	}
	name, err := filepath.Rel(root, packageDir)
	if err != nil || name == ".." || strings.HasPrefix(name, ".."+string(filepath.Separator)) {
		name = fmt.Sprintf("external-%x", sha256.Sum256([]byte(packageDir)))
	}
	// Preserve the support package's persistent cache across its source relocation.
	if name == filepath.Join("internal", "testsupport", "native") {
		name = filepath.Join("internal", "testsupport", "rustparity")
	}
	return filepath.Join(root, ".harness", "cache", "rust-parity", name, Version, host)
}

type BuildResult struct {
	Command        []string
	Stdout, Stderr []byte
	Err            error
	Artifacts      []Artifact
	Host           string
}

func Build(t testing.TB, project string, o BuildOptions) BuildResult {
	t.Helper()
	finishIdentity := timing.Start("helper.toolchain", project)
	tc := Toolchain(t)
	finishIdentity(nil)
	defaultCache := o.TargetDir == ""
	if o.TargetDir == "" {
		o.TargetDir = defaultTargetDirectory(t, tc.Host)
	} else if !filepath.IsAbs(o.TargetDir) {
		o.TargetDir = filepath.Join(project, o.TargetDir)
	}
	var err error
	o.TargetDir, err = filepath.Abs(o.TargetDir)
	if err != nil {
		t.Fatal(err)
	}
	// Let Cargo create the target itself, including its CACHEDIR.TAG marker.
	if err = os.MkdirAll(filepath.Dir(o.TargetDir), 0700); err != nil {
		t.Fatal(err)
	}
	targetParent, err := filepath.EvalSymlinks(filepath.Dir(o.TargetDir))
	if err != nil {
		t.Fatal(err)
	}
	o.TargetDir = filepath.Join(targetParent, filepath.Base(o.TargetDir))
	excludedTarget := o.TargetDir
	if o.Timeout == 0 {
		o.Timeout = 3 * time.Minute
	}
	args := []string{"build", "--frozen", "--message-format=json", "--target", tc.Host}
	if o.Tests {
		args = []string{"test", "--no-run", "--frozen", "--message-format=json", "--target", tc.Host}
	}
	if o.Binary != "" {
		args = []string{"rustc", "--bin", o.Binary, "--frozen", "--message-format=json", "--target", tc.Host}
	}
	// A member can inherit dependencies/profile settings and the lock from its
	// workspace root. Retain that entire graph and still select the original
	// member, rather than turning the member into a standalone package.
	lookupCtx, lookupCancel := context.WithTimeout(context.Background(), 15*time.Second)
	finishLocate := timing.Start("cargo.locate-project", project)
	discovery := locateCargoWorkspace(lookupCtx, tc.Cargo, project, CleanEnvironment(tc.Rustc))
	finishLocate(discovery.Err)
	lookupCancel()
	if discovery.Err != nil {
		discovery.Err = fmt.Errorf("locate unchanged Cargo workspace: %w", discovery.Err)
		discovery.Host = tc.Host
		return discovery
	}
	workspace := filepath.Dir(strings.TrimSpace(string(discovery.Stdout)))
	workspace, workspaceErr := filepath.EvalSymlinks(workspace)
	selectedManifest, selectedErr := filepath.EvalSymlinks(filepath.Join(project, "Cargo.toml"))
	if workspaceErr != nil || selectedErr != nil {
		t.Fatalf("resolve Cargo workspace/member: %v / %v", workspaceErr, selectedErr)
	}
	relativeManifest, relativeErr := filepath.Rel(workspace, selectedManifest)
	if relativeErr != nil || !filepath.IsLocal(relativeManifest) {
		t.Fatalf("selected Cargo member is outside its workspace: %v", relativeErr)
	}
	if defaultCache {
		shard, err := cargocache.Shard(workspace, excludedTarget)
		if err != nil {
			t.Fatalf("identify actual Cargo input graph: %v", err)
		}
		o.TargetDir = filepath.Join(o.TargetDir, fmt.Sprintf("shard-%d", shard))
		if err := os.MkdirAll(filepath.Dir(o.TargetDir), 0700); err != nil {
			t.Fatal(err)
		}
	}
	// The build budget includes contention: waiting for another process must
	// not grant a fresh timeout or mutate the inputs of its running Cargo.
	ctx, cancel := context.WithTimeout(context.Background(), o.Timeout)
	defer cancel()
	finishLock := timing.Start("helper.cache_wait", o.TargetDir)
	release, lockErr := acquireBuildCache(ctx, o.TargetDir)
	finishLock(lockErr)
	if lockErr != nil {
		return BuildResult{Err: fmt.Errorf("acquire Cargo input cache: %w", lockErr), Host: tc.Host}
	}
	defer func() {
		if err := release(); err != nil {
			t.Errorf("release Cargo input cache: %v", err)
		}
	}()
	inputCache := o.TargetDir + "-inputs"
	if relativeTarget, err := filepath.Rel(workspace, o.TargetDir); err == nil && filepath.IsLocal(relativeTarget) {
		// A relative target can live within the original workspace. Keep all
		// owned cache state in that excluded subtree so later builds cannot
		// recursively copy the prior staged inputs into the next graph.
		inputCache = filepath.Join(o.TargetDir, "l-inputs")
	}
	stage := cargocache.Stage
	if defaultCache {
		// Each package's bounded default slot keeps a stable native package
		// root across member/manifest changes. Reconcile the complete input
		// inventory under the same lock held through Cargo and snapshots.
		stage = cargocache.StageCurrent
	}
	finishInputs := timing.Start("helper.inputs", o.TargetDir)
	inputs, stageErr := stage(workspace, inputCache, excludedTarget)
	if stageErr != nil {
		finishInputs(stageErr)
		t.Fatalf("retain unchanged Cargo input bytes: %v", stageErr)
	}
	if err := cargocache.Activate(inputs, o.TargetDir); err != nil {
		finishInputs(err)
		t.Fatalf("select Cargo cache input graph: %v", err)
	}
	finishInputs(nil)
	args = append(args, "--manifest-path", filepath.Join(inputs, relativeManifest), "--target-dir", o.TargetDir)
	if timing.Enabled() {
		args = append(args, "--timings")
	}
	if o.TestConfiguration {
		args = append(args, "--", "--cfg", "test")
	}
	r := BuildResult{Command: append([]string{tc.Cargo}, args...), Host: tc.Host}
	cmd := exec.CommandContext(ctx, tc.Cargo, args...)
	superviseCompiler(cmd)
	cmd.Dir = project
	cmd.Env = CleanEnvironment(tc.Rustc)
	var out, diag bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &diag
	capture := beginCargoEvidence(t, cargoEvidenceInput{Route: "direct-helper", Test: t.Name(), Command: cmd.Args, Directory: cmd.Dir, Environment: cmd.Env, OriginalWorkspace: workspace, Workspace: inputs, Manifest: filepath.Join(inputs, relativeManifest), Target: o.TargetDir, PreferredSlot: o.TargetDir, SelectedSlot: o.TargetDir, Toolchain: Version, Host: tc.Host})
	compileStarted := time.Now()
	finishCargo := timing.Start("cargo.actual", project)
	r.Err = cmd.Run()
	cargoDeadline := ctx.Err()
	finishCargo(r.Err)
	r.Stdout = bytes.Clone(out.Bytes())
	r.Stderr = bytes.Clone(diag.Bytes())
	capture.Finish(r.Stdout, r.Stderr, r.Err)
	recordCargoTiming(project, o.TargetDir, compileStarted, r.Stdout)
	if cargoDeadline != nil {
		r.Err = fmt.Errorf("Cargo subprocess deadline: %w (%v)", cargoDeadline, r.Err)
	}
	// Cargo's JSON protocol is line-delimited and may be interspersed with
	// unrelated native stdout. One such line must not discard later artifacts.
	for _, line := range bytes.Split(r.Stdout, []byte{'\n'}) {
		var event struct {
			Reason     string `json:"reason"`
			PackageID  string `json:"package_id"`
			Executable string `json:"executable"`
			Target     struct {
				Name string
				Kind []string
			}
		}
		if e := json.Unmarshal(line, &event); e != nil {
			continue
		}
		if event.Reason == "compiler-artifact" && event.Executable != "" {
			r.Artifacts = append(r.Artifacts, Artifact{event.PackageID, event.Target.Name, event.Executable, event.Target.Kind})
		}
	}
	// Cargo can reuse and replace target-directory executables on the next
	// build. Snapshot every observed executable before releasing the lock;
	// callers may retain earlier generations across later builds.
	finishSnapshots := timing.Start("helper.snapshots", project)
	if len(r.Artifacts) != 0 {
		directory := t.TempDir()
		for i := range r.Artifacts {
			artifact := &r.Artifacts[i]
			source, err := os.Open(artifact.Executable)
			if err != nil {
				t.Fatalf("retain Cargo artifact %s: %v", artifact.Executable, err)
			}
			info, err := source.Stat()
			if err != nil || !info.Mode().IsRegular() {
				source.Close() //nolint:errcheck // Preserve the failed artifact validation while releasing its read-only descriptor.
				t.Fatalf("Cargo executable is not a regular file: %s (%v)", artifact.Executable, err)
			}
			path := filepath.Join(directory, fmt.Sprintf("%d-%s", i, filepath.Base(artifact.Executable)))
			destination, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, info.Mode().Perm()&0555)
			if err != nil {
				source.Close() //nolint:errcheck // Preserve the destination creation error while releasing the read-only source.
				t.Fatal(err)
			}
			_, copyErr := io.Copy(destination, source)
			closeErr := destination.Close()
			sourceCloseErr := source.Close()
			if copyErr != nil || closeErr != nil || sourceCloseErr != nil {
				t.Fatalf("retain Cargo executable: copy=%v close=%v source close=%v", copyErr, closeErr, sourceCloseErr)
			}
			artifact.Executable = path
		}
	}
	finishSnapshots(nil)
	return r
}

func locateCargoWorkspace(ctx context.Context, cargo, project string, environment []string) BuildResult {
	cmd := exec.CommandContext(ctx, cargo, "locate-project", "--workspace", "--frozen", "--message-format", "plain", "--manifest-path", filepath.Join(project, "Cargo.toml"))
	cmd.Dir, cmd.Env = project, environment
	var out, diag bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &diag
	err := cmd.Run()
	return BuildResult{Command: cmd.Args, Stdout: bytes.Clone(out.Bytes()), Stderr: bytes.Clone(diag.Bytes()), Err: err}
}

// CleanEnvironment isolates compilation from caller target/flag/tool overrides.
// Dependency caches and registry configuration remain the caller's installed setup.
func CleanEnvironment(rustc string) []string {
	var env []string
	for _, item := range os.Environ() {
		key := strings.SplitN(item, "=", 2)[0]
		if key == "RUSTC" || key == "RUSTC_WRAPPER" || key == "RUSTC_WORKSPACE_WRAPPER" || key == "RUSTFLAGS" || key == "RUSTDOCFLAGS" || key == "CARGO_ENCODED_RUSTFLAGS" || key == "CARGO_ENCODED_RUSTDOCFLAGS" || key == "RUSTUP_TOOLCHAIN" || key == "CARGO_BUILD_RUSTFLAGS" || key == "CARGO_BUILD_RUSTDOCFLAGS" || key == "CARGO_BUILD_BUILD_DIR" || key == "CARGO_BUILD_TARGET_DIR" || key == "CARGO_BUILD_TARGET" || key == "CARGO_TARGET_DIR" || strings.HasPrefix(key, "CARGO_TARGET_") || strings.HasPrefix(key, "CARGO_PROFILE_") {
			continue
		}
		env = append(env, item)
	}
	return append(env, "RUSTC="+rustc, "CARGO_TERM_COLOR=never", "CARGO_PROFILE_DEV_DEBUG=0", "CARGO_PROFILE_TEST_DEBUG=0")
}

func Materialize(t testing.TB, files map[string][]byte) string {
	t.Helper()
	dir := t.TempDir()
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if name == "" || filepath.IsAbs(name) || filepath.ToSlash(filepath.Clean(name)) != name || name == ".." || strings.HasPrefix(name, "../") || strings.Contains(name, "\\") {
			t.Fatalf("unsafe test bundle path %q", name)
		}
		path := filepath.Join(dir, filepath.FromSlash(name))
		if e := os.MkdirAll(filepath.Dir(path), 0700); e != nil {
			t.Fatal(e)
		}
		if e := os.WriteFile(path, files[name], 0600); e != nil {
			t.Fatal(e)
		}
	}
	return dir
}

// AssertNativeArtifact checks the published native artifact. Cargo provenance is a
// separate required observation: absence of Go build information alone is not proof.
func AssertNativeArtifact(t testing.TB, path string) {
	t.Helper()
	info, e := os.Stat(path)
	if e != nil || !info.Mode().IsRegular() || info.Size() == 0 || info.Mode().Perm()&0111 == 0 {
		t.Fatalf("native executable prerequisite %s: %v", path, e)
	}
	if _, e := buildinfo.ReadFile(path); e == nil {
		t.Fatalf("published artifact is a Go executable: %s", path)
	}
	switch runtime.GOOS {
	case "darwin":
		f, e := macho.Open(path)
		if e != nil {
			t.Fatalf("native Mach-O: %v", e)
		}
		defer f.Close() //nolint:errcheck // The binary is opened read-only; validation, not descriptor cleanup, determines this observation.
		want := macho.CpuAmd64
		if runtime.GOARCH == "arm64" {
			want = macho.CpuArm64
		}
		if f.Cpu != want {
			t.Fatalf("native CPU=%v want %v", f.Cpu, want)
		}
	case "linux":
		f, e := elf.Open(path)
		if e != nil {
			t.Fatalf("native ELF: %v", e)
		}
		defer f.Close() //nolint:errcheck // The binary is opened read-only; validation, not descriptor cleanup, determines this observation.
		want := elf.EM_X86_64
		if runtime.GOARCH == "arm64" {
			want = elf.EM_AARCH64
		}
		if f.Machine != want {
			t.Fatalf("native CPU=%v want %v", f.Machine, want)
		}
	default:
		t.Fatalf("native parity supports macOS/Linux, got %s", runtime.GOOS)
	}
}
