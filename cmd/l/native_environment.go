package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/l-programming-language/l/contracts"
	"github.com/l-programming-language/l/lerrors"
	"maps"
)

type nativeEnvironmentQuery func(context.Context, string, []string, string, []string) ([]byte, []byte, int, error)

type nativeEnvironmentObserver struct {
	lookup func(string) (string, error)
	query  nativeEnvironmentQuery
}

func nativeBytesHash(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

func runNativeEnvironmentQuery(ctx context.Context, executable string, args []string, directory string, environment []string) ([]byte, []byte, int, error) {
	command := exec.CommandContext(ctx, executable, args...)
	command.Dir, command.Env = directory, environment
	configureBuildCommand(command)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	if command.Process != nil && command.Cancel != nil {
		if cleanupErr := command.Cancel(); cleanupErr != nil && !errors.Is(cleanupErr, os.ErrProcessDone) {
			err = errors.Join(err, cleanupErr)
		}
	}
	exit := -1
	if command.ProcessState != nil {
		exit = command.ProcessState.ExitCode()
	}
	return stdout.Bytes(), stderr.Bytes(), exit, err
}

// captureNativeEnvironment is an explicitly requested observation. Inspect,
// check and explain only read its document; they never invoke this function or
// replay its recorded commands.
func captureNativeEnvironment(ctx context.Context, report contracts.Report, assertedTarget string) (contracts.NativeEnvironment, error) {
	return captureNativeEnvironmentWith(ctx, report, assertedTarget, nativeEnvironmentObserver{lookup: exec.LookPath, query: runNativeEnvironmentQuery})
}

func captureNativeEnvironmentWith(ctx context.Context, report contracts.Report, assertedTarget string, observer nativeEnvironmentObserver) (contracts.NativeEnvironment, error) {
	value := contracts.NativeEnvironment{SchemaVersion: contracts.SchemaVersion, Kind: "native_environment", Producer: contracts.AnalyzerVersion, Version: "1", Method: "capture", Host: runtime.GOOS + "/" + runtime.GOARCH, PanicPolicy: "unknown-effective-native-policy", Allocator: "unknown-native-allocator", Inputs: maps.Clone(report.Inputs), Dependencies: maps.Clone(report.Lowering.Fingerprints), Configuration: make(map[string]string), Limitations: []string{"Environment capture records tool identity and versions only; no native build or App execution was performed.", "Panic strategy, allocator behavior, linker internals and transitive native implementation conformance remain unproved.", "The record describes this observation, not the toolchain or deployment present when the document is later inspected."}}
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		return value, lerrors.Unsupported.Errorf("native environment capture requires macOS or Linux")
	}
	if len(value.Inputs) == 0 || report.Entry == "" {
		return value, fmt.Errorf("native environment capture requires a prepared source report")
	}
	directory, err := filepath.Abs(filepath.Dir(report.Entry))
	if err != nil {
		return value, err
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	environment := append(cargoBuildEnvironment(), "RUSTUP_AUTO_INSTALL=0")
	if err := nativeConfiguration(value.Configuration, directory, environment); err != nil {
		return value, err
	}
	observe := func(executable string, args ...string) ([]byte, error) {
		started := time.Now().UTC()
		stdout, stderr, exit, queryErr := observer.query(ctx, executable, args, directory, environment)
		record := nativeExecution("environment", executable, args, directory, started, time.Now().UTC(), stdout, stderr, exit, queryErr)
		value.Executions = append(value.Executions, record)
		if queryErr != nil || exit != 0 {
			if queryErr == nil {
				queryErr = fmt.Errorf("tool returned exit %d", exit)
			}
			return nil, fmt.Errorf("observe native tool %s: %w (exit %d): %s", executable, errors.Join(queryErr, ctx.Err()), exit, strings.TrimSpace(string(stderr)))
		}
		return stdout, nil
	}
	rustup, err := observer.lookup("rustup")
	if err != nil {
		return value, fmt.Errorf("resolve pinned Rust toolchain manager: %w", err)
	}
	rustup, err = filepath.Abs(rustup)
	if err != nil {
		return value, err
	}
	version, err := observe(rustup, "--version")
	if err != nil {
		return value, err
	}
	if err := addNativeTool(&value, "rustup", rustup, string(version)); err != nil {
		return value, err
	}
	for _, name := range []string{"cargo", "rustc"} {
		resolved, queryErr := observe(rustup, "which", "--toolchain", nativeToolchainVersion, name)
		if queryErr != nil {
			return value, queryErr
		}
		tool := strings.TrimSpace(string(resolved))
		if !filepath.IsAbs(tool) || strings.ContainsAny(tool, "\r\n") {
			return value, fmt.Errorf("toolchain manager returned no unique absolute %s executable", name)
		}
		version, queryErr := observe(tool, "--version", "--verbose")
		if queryErr != nil {
			return value, queryErr
		}
		if !strings.HasPrefix(string(version), name+" "+nativeToolchainVersion+" ") && !strings.HasPrefix(string(version), name+" "+nativeToolchainVersion+"\n") {
			return value, fmt.Errorf("observed %s does not match pinned toolchain %s", name, nativeToolchainVersion)
		}
		if err := addNativeTool(&value, name, tool, string(version)); err != nil {
			return value, err
		}
		if name == "rustc" {
			value.Compiler = strings.TrimSpace(string(version))
			for _, line := range strings.Split(string(version), "\n") {
				if target, found := strings.CutPrefix(line, "host: "); found {
					if value.Target != "" {
						return value, fmt.Errorf("rustc returned multiple host targets")
					}
					value.Target = strings.TrimSpace(target)
				}
			}
		}
	}
	launcher, err := observer.lookup("cargo")
	if err != nil {
		return value, fmt.Errorf("resolve Cargo build launcher: %w", err)
	}
	launcher, err = filepath.Abs(launcher)
	if err != nil {
		return value, err
	}
	version, err = observe(launcher, "+"+nativeToolchainVersion, "--version", "--verbose")
	if err != nil {
		return value, err
	}
	for _, tool := range value.Tools {
		if tool.Name == "cargo" && tool.Version != strings.TrimSpace(string(version)) {
			return value, fmt.Errorf("Cargo launcher does not report the selected pinned Cargo version")
		}
	}
	if err := addNativeTool(&value, "cargo-launcher", launcher, string(version)); err != nil {
		return value, err
	}
	if !nativeTargetMatchesHost(value.Target) || assertedTarget != "" && assertedTarget != value.Target {
		return value, fmt.Errorf("capture target %q does not match observed native host %q; cross-compilation is unsupported", assertedTarget, value.Target)
	}
	value = contracts.SealNativeEnvironment(value)
	return value, value.Validate()
}

func nativeTargetMatchesHost(target string) bool {
	architecture := runtime.GOARCH
	switch architecture {
	case "arm64":
		architecture = "aarch64"
	case "amd64":
		architecture = "x86_64"
	case "386":
		architecture = "i686"
	}
	return strings.HasPrefix(target, architecture+"-") && (runtime.GOOS == "darwin" && strings.HasSuffix(target, "-apple-darwin") || runtime.GOOS == "linux" && strings.Contains(target, "-linux-"))
}

func addNativeTool(value *contracts.NativeEnvironment, name, path, version string) error {
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(resolved)
	if err != nil {
		return fmt.Errorf("fingerprint %s executable: %w", name, err)
	}
	if strings.TrimSpace(version) == "" {
		return fmt.Errorf("%s returned an empty version", name)
	}
	value.Tools = append(value.Tools, contracts.NativeToolIdentity{Name: name, Path: resolved, SHA256: nativeBytesHash(data), Version: strings.TrimSpace(version)})
	return nil
}

func nativeConfigurationEnvironmentKey(key string) bool {
	switch key {
	case "PATH", "CARGO_HOME", "RUSTUP_HOME", "RUSTUP_TOOLCHAIN", "RUSTFLAGS", "RUSTDOCFLAGS", "CC", "CXX", "AR",
		"CARGO_NET_OFFLINE", "CARGO_ENCODED_RUSTFLAGS", "CARGO_TERM_COLOR", "CARGO_PROFILE_DEV_DEBUG", "CARGO_INCREMENTAL",
		"RUSTC_WRAPPER", "RUSTC_WORKSPACE_WRAPPER", "RUSTC_BOOTSTRAP", "RUSTC_FORCE_INCREMENTAL":
		return true
	default:
		return false
	}
}

func nativeConfiguration(configuration map[string]string, directory string, environment []string) error {
	configuration["toolchain"] = nativeToolchainVersion
	configuration["target-selection"] = "host-tuple"
	configuration["profile"] = "dev"
	var cargoHome string
	for _, item := range environment {
		key, value, _ := strings.Cut(item, "=")
		if nativeConfigurationEnvironmentKey(key) {
			// Record only the documented native configuration settings. Arbitrary
			// Cargo/Rust variables, including credentials, are excluded even as hashes.
			configuration["environment:"+key] = nativeBytesHash([]byte(value))
		}
		if key == "CARGO_HOME" {
			cargoHome = value
		}
	}
	if cargoHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		cargoHome = filepath.Join(home, ".cargo")
	}
	if !filepath.IsAbs(cargoHome) {
		return fmt.Errorf("native environment evidence requires an absolute CARGO_HOME")
	}
	paths := map[string]bool{filepath.Join(cargoHome, "config"): true, filepath.Join(cargoHome, "config.toml"): true}
	for current := directory; ; current = filepath.Dir(current) {
		paths[filepath.Join(current, ".cargo", "config")] = true
		paths[filepath.Join(current, ".cargo", "config.toml")] = true
		if parent := filepath.Dir(current); parent == current {
			break
		}
	}
	ordered := make([]string, 0, len(paths))
	for path := range paths {
		ordered = append(ordered, path)
	}
	sort.Strings(ordered)
	for _, path := range ordered {
		data, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) {
			configuration["cargo-config:"+path] = "absent"
			continue
		}
		if err != nil {
			return fmt.Errorf("fingerprint Cargo configuration: %w", err)
		}
		configuration["cargo-config:"+path] = nativeBytesHash(data)
	}
	return nil
}

func nativeExecution(kind, executable string, args []string, directory string, started, finished time.Time, stdout, stderr []byte, exit int, err error) contracts.ExecutionRecord {
	status := "passed"
	if err != nil || exit != 0 {
		status = "failed"
		if exit == 0 {
			exit = -1
		}
	}
	record := contracts.ExecutionRecord{Kind: kind, Command: append([]string{executable}, args...), WorkingDirectory: directory, Status: status, Started: started.Format(time.RFC3339Nano), Finished: finished.Format(time.RFC3339Nano), ExitCode: &exit, StdoutFingerprint: nativeBytesHash(stdout), StderrFingerprint: nativeBytesHash(stderr)}
	record.ID = contracts.StableID("native-execution", record)
	return contracts.SealExecutionRecord(record)
}

type nativeBuildRecorder struct {
	environment contracts.NativeEnvironment
}
type nativeBuildEvidenceKey struct{}

func nativeBuildEvidence(ctx context.Context) *nativeBuildRecorder {
	if recorder, ok := ctx.Value(nativeBuildEvidenceKey{}).(*nativeBuildRecorder); ok {
		return recorder
	}
	return nil
}

func (recorder *nativeBuildRecorder) record(command *exec.Cmd, started time.Time, stdout, stderr []byte, err error) string {
	exit := -1
	if command.ProcessState != nil {
		exit = command.ProcessState.ExitCode()
	}
	record := nativeExecution("build", command.Path, command.Args[1:], command.Dir, started, time.Now().UTC(), stdout, stderr, exit, err)
	recorder.environment.Executions = append(recorder.environment.Executions, record)
	return record.ID
}

func (recorder *nativeBuildRecorder) verifyTools() error {
	for _, tool := range recorder.environment.Tools {
		data, err := os.ReadFile(tool.Path)
		if err != nil {
			return err
		}
		if nativeBytesHash(data) != tool.SHA256 {
			return fmt.Errorf("native tool changed after environment capture: %s", tool.Name)
		}
	}
	return nil
}

func (recorder *nativeBuildRecorder) verifyLauncher(path string) error {
	actual, err := filepath.EvalSymlinks(path)
	if err != nil {
		return err
	}
	for _, tool := range recorder.environment.Tools {
		if tool.Name == "cargo-launcher" && tool.Path == actual {
			return nil
		}
	}
	return fmt.Errorf("Cargo launcher changed after native environment capture")
}

func (recorder *nativeBuildRecorder) artifact(path, inputs, execution string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	recorder.environment.Artifacts = append(recorder.environment.Artifacts, contracts.NativeArtifact{Path: path, SHA256: nativeBytesHash(data), Inputs: inputs, Execution: execution})
	return nil
}

func (recorder *nativeBuildRecorder) write(path string) error {
	if recorder == nil {
		return nil
	}
	data, err := contracts.EncodeDocument(contracts.SealNativeEnvironment(recorder.environment))
	if err != nil {
		return err
	}
	return replaceFile(path, append(data, '\n'))
}

// Replace only complete output. A failed write must leave any previous file intact.
func replaceFile(path string, data []byte) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".l-output-")
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	if err = errors.Join(err, file.Close()); err == nil {
		err = os.Rename(file.Name(), path)
	}
	if err != nil {
		return errors.Join(err, os.Remove(file.Name()))
	}
	return nil
}

func nativeWorkspaceFingerprint(directory string) (string, error) {
	inputs := make(map[string]string)
	err := filepath.WalkDir(directory, func(path string, entry os.DirEntry, visitErr error) error {
		if visitErr != nil {
			return visitErr
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("native evidence workspace contains a symlink: %s", path)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(directory, path)
		if err != nil {
			return err
		}
		inputs[filepath.ToSlash(relative)] = nativeBytesHash(data)
		return nil
	})
	if err != nil {
		return "", err
	}
	return contracts.StableID("native-workspace", inputs), nil
}

func resolveNativeEvidenceOutput(path string, inputs []string, output, companion, assets string) (string, error) {
	if path == "" {
		return path, nil
	}
	var err error
	path, err = canonicalBuildPath(path)
	if err != nil {
		return "", err
	}
	if info, err := os.Stat(path); err == nil && !info.Mode().IsRegular() {
		return "", lerrors.InvalidInput.Errorf("native evidence requires a regular file destination: %s", path)
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	output, err = canonicalBuildPath(output)
	if err != nil {
		return "", err
	}
	companion, err = canonicalBuildPath(companion)
	if err != nil {
		return "", err
	}
	for _, input := range inputs {
		canonical, err := canonicalBuildPath(input)
		if err != nil {
			return "", err
		}
		if sameNativeEvidencePath(path, canonical) {
			return "", lerrors.InvalidInput.Errorf("native evidence output must not replace source input %s", input)
		}
	}
	if sameNativeEvidencePath(path, output) {
		return "", lerrors.InvalidInput.Errorf("native evidence output must not replace the executable")
	}
	within, err := buildPathContains(companion, path)
	if err != nil {
		return "", err
	}
	if within {
		return "", lerrors.InvalidInput.Errorf("native evidence output must be outside companion generations")
	}
	if assets != "" {
		if err := validateAssetDestinations(assets, path); err != nil {
			return "", err
		}
	}
	return path, nil
}

// Both paths have already been canonicalized. SameFile also catches hard links.
func sameNativeEvidencePath(left, right string) bool {
	if left == right {
		return true
	}
	a, errA := os.Stat(left)
	b, errB := os.Stat(right)
	return errA == nil && errB == nil && os.SameFile(a, b)
}
