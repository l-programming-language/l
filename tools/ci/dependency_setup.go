package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/l-programming-language/l/tools/internal/testgate"
)

type dependencyCacheTarget struct {
	Target string `json:"target"`
	Lock   string `json:"lock"`
}

type dependencyCachePreparation struct {
	dependencyCacheTarget
	Status string `json:"status"`
	Bytes  int64  `json:"bytes"`
}

// prepareRustDependencies prepares only registry libraries from unchanged,
// retained runtime manifests. Selecting those packages excludes project targets.
// This is setup work, never a compiler/application test or acceptance evidence.
func (r *runner) prepareRustDependencies(seeds []string) (resultErr error) {
	if err := rejectDependencyBuildOverrides(os.Environ()); err != nil {
		return err
	}
	var runtimeSeed string
	for _, seed := range seeds {
		if filepath.Base(seed) == "runtime" {
			runtimeSeed = seed
		}
	}
	if runtimeSeed == "" {
		return errors.New("Rust dependency preparation requires the retained runtime seed")
	}
	inputs, err := dependencyInputHashes(runtimeSeed, filepath.Join(r.root, "lyger", "test", "native"))
	if err != nil {
		return err
	}
	defer func() { resultErr = errors.Join(resultErr, unchangedDependencyInputs(inputs)) }()
	cargo, err := r.dependencyToolPath("cargo")
	if err != nil {
		return err
	}
	rustc, err := r.dependencyToolPath("rustc")
	if err != nil {
		return err
	}
	nofs := filepath.Join(r.directory, "dependency-runtime-nofs")
	for _, name := range []string{"Cargo.toml", "Cargo.lock"} {
		source := filepath.Join(r.root, "logiceme", "testdata", "runtime", name)
		before, err := testgate.FileHash(source)
		if err != nil {
			return err
		}
		defer func() { resultErr = errors.Join(resultErr, requireUnchanged(source, before)) }()
		if err := copyFile(source, filepath.Join(nofs, name)); err != nil {
			return err
		}
	}
	cache := filepath.Join(r.root, ".harness", "cache")
	if err := dependencyDirectory(cache); err != nil {
		return err
	}
	target, err := os.MkdirTemp(cache, ".registry-dependencies-")
	if err != nil {
		return err
	}
	defer func() { resultErr = errors.Join(resultErr, os.RemoveAll(target)) }()
	for _, seed := range []string{runtimeSeed, nofs} {
		if err := r.buildRegistryDependencies(seed, target, cargo, rustc); err != nil {
			return err
		}
	}
	identity, err := r.command("dependency-rust-host", time.Minute, nil, rustc, "-vV")
	if err != nil {
		return err
	}
	host, err := dependencyRustHost(identity)
	if err != nil {
		return err
	}
	userCache, err := os.UserCacheDir()
	if err != nil {
		return err
	}
	var prepared []dependencyCachePreparation
	defer func() {
		resultErr = errors.Join(resultErr, testgate.WriteJSON(filepath.Join(r.directory, "dependency-caches.json"), prepared))
	}()
	for _, destination := range dependencyCacheTargets(r.root, userCache, r.pins.Rust, runtime.GOOS, runtime.GOARCH, host) {
		status, size, err := seedDependencyCache(target, destination)
		prepared = append(prepared, dependencyCachePreparation{dependencyCacheTarget: destination, Status: status, Bytes: size})
		if err != nil {
			return fmt.Errorf("prepare dependency cache %s: %w", destination.Target, err)
		}
	}
	return nil
}

func (r *runner) buildRegistryDependencies(seed, target, cargo, rustc string) (resultErr error) {
	inputs, err := dependencyInputHashes(seed)
	if err != nil {
		return err
	}
	defer func() { resultErr = errors.Join(resultErr, unchangedDependencyInputs(inputs)) }()
	dependencies, err := readLockDependencies(filepath.Join(seed, "Cargo.lock"))
	if err != nil {
		return err
	}
	allowed := map[string]bool{}
	for dependency := range dependencies {
		allowed[dependency.Source+"#"+dependency.Name+"@"+dependency.Version] = true
	}
	args := registryDependencyBuildArgs(seed, target)
	output, err := r.command("dependencies-"+filepath.Base(seed), 5*time.Minute, map[string]string{
		"CARGO_NET_OFFLINE": "true", "RUSTUP_AUTO_INSTALL": "0", "RUSTC": rustc,
		"CARGO_PROFILE_DEV_DEBUG": "0", "CARGO_PROFILE_TEST_DEBUG": "0",
	}, cargo, args...)
	if err != nil {
		return err
	}
	return validateRegistryDependencyBuild(output, allowed)
}

func registryDependencyBuildArgs(seed, target string) []string {
	// These exact versions are shared by the retained runtime manifests/locks.
	// New dependency versions require a reviewed setup change, not resolution.
	return []string{"build", "--frozen", "--lib", "-p", "tokio@1.53.1", "-p", "nix@0.31.3", "--target", "host-tuple", "--manifest-path", filepath.Join(seed, "Cargo.toml"), "--target-dir", target, "--message-format=json"}
}

func validateRegistryDependencyBuild(data []byte, allowed map[string]bool) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	finished, artifacts := false, 0
	roots := map[string]bool{"tokio": false, "nix": false}
	for {
		var event struct {
			Reason    string `json:"reason"`
			PackageID string `json:"package_id"`
			Success   bool   `json:"success"`
			Target    struct {
				Name string   `json:"name"`
				Kind []string `json:"kind"`
			} `json:"target"`
		}
		err := decoder.Decode(&event)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("invalid registry dependency build stream: %w", err)
		}
		if finished {
			return errors.New("registry dependency build has events after completion")
		}
		if event.PackageID != "" && (!strings.HasPrefix(event.PackageID, "registry+") || !allowed[event.PackageID]) {
			return fmt.Errorf("dependency setup selected a local or unlocked package: %s", event.PackageID)
		}
		switch event.Reason {
		case "compiler-artifact":
			if event.PackageID == "" || len(event.Target.Kind) != 1 {
				return errors.New("dependency artifact lacks a locked package or singular target kind")
			}
			switch event.Target.Kind[0] {
			case "lib", "proc-macro", "custom-build":
			default:
				return fmt.Errorf("dependency setup compiled forbidden target kind %q", event.Target.Kind[0])
			}
			artifacts++
			if _, required := roots[event.Target.Name]; required && event.Target.Kind[0] == "lib" {
				roots[event.Target.Name] = true
			}
		case "build-finished":
			if !event.Success {
				return errors.New("registry dependency build did not succeed")
			}
			finished = true
		case "build-script-executed", "compiler-message":
			if event.PackageID == "" {
				return errors.New("dependency build event lacks its locked registry package")
			}
		default:
			return fmt.Errorf("unexpected dependency build event %q", event.Reason)
		}
	}
	if !finished || artifacts == 0 || !roots["tokio"] || !roots["nix"] {
		return errors.New("dependency build did not establish complete Tokio/nix library results")
	}
	return nil
}

func dependencyRustHost(output []byte) (string, error) {
	for _, line := range strings.Split(string(output), "\n") {
		if host, ok := strings.CutPrefix(line, "host: "); ok && host != "" && filepath.Base(host) == host && host != "." && host != ".." && !strings.ContainsAny(host, "\\\r\n ") {
			return host, nil
		}
	}
	return "", errors.New("pinned rustc did not identify a usable native host triple")
}

func dependencyCacheTargets(root, userCache, version, goos, goarch, host string) []dependencyCacheTarget {
	var targets []dependencyCacheTarget
	// This finite setup scope covers the cold hosted failures. It does not
	// prebuild arbitrary packages, isolated test targets, or the Lucid matrix.
	for slot := 0; slot < 6; slot++ {
		directory := filepath.Join(userCache, "l", "rust", version, goos+"-"+goarch, fmt.Sprintf("shard-%d", slot))
		targets = append(targets, dependencyCacheTarget{Target: filepath.Join(directory, "target"), Lock: filepath.Join(directory, "publication.lock")})
	}
	for _, packagePath := range []string{"lyger", "logiceme", "cmd/l", "acceptance/phase1/runtime"} {
		for shard := 0; shard < 4; shard++ {
			target := filepath.Join(root, ".harness", "cache", "rust-parity", filepath.FromSlash(packagePath), version, host, fmt.Sprintf("shard-%d", shard))
			targets = append(targets, dependencyCacheTarget{Target: target, Lock: filepath.Join(target, ".l-parity-publication.lock")})
		}
	}
	for _, name := range []string{"rust-runtime-parity", "rust-empty-error-runtime", "rust-empty-error-frontend"} {
		target := filepath.Join(root, ".harness", "cache", name)
		targets = append(targets, dependencyCacheTarget{Target: target, Lock: filepath.Join(target, ".l-parity-publication.lock")})
	}
	return targets
}

func seedDependencyCache(source string, destination dependencyCacheTarget) (status string, size int64, resultErr error) {
	if err := dependencyDirectory(filepath.Dir(destination.Lock)); err != nil {
		return "failed", 0, err
	}
	release, err := lockDependencyCache(destination.Lock)
	if err != nil {
		return "failed", 0, err
	}
	defer func() { resultErr = errors.Join(resultErr, release()) }()
	if err := dependencyDirectory(destination.Target); err != nil {
		return "failed", 0, err
	}
	entries, err := os.ReadDir(destination.Target)
	if err != nil {
		return "failed", 0, err
	}
	for _, entry := range entries {
		if filepath.Join(destination.Target, entry.Name()) != destination.Lock {
			return "existing", 0, nil
		}
	}
	staged, err := os.MkdirTemp(filepath.Dir(destination.Target), ".l-dependency-seed-")
	if err != nil {
		return "failed", 0, err
	}
	defer func() { resultErr = errors.Join(resultErr, os.RemoveAll(staged)) }()
	size, err = copyDependencyCache(source, staged)
	if err != nil {
		return "failed", 0, err
	}
	entries, err = os.ReadDir(staged)
	if err != nil {
		return "failed", 0, err
	}
	var installed []string
	for _, entry := range entries {
		path := filepath.Join(destination.Target, entry.Name())
		if _, err = os.Lstat(path); !errors.Is(err, fs.ErrNotExist) {
			err = fmt.Errorf("dependency cache destination appeared during locked preparation: %s", path)
			break
		}
		if err = os.Rename(filepath.Join(staged, entry.Name()), path); err != nil {
			break
		}
		installed = append(installed, path)
	}
	if err != nil {
		for _, path := range installed {
			err = errors.Join(err, os.RemoveAll(path))
		}
		return "failed", 0, err
	}
	return "seeded", size, nil
}

func dependencyDirectory(path string) error {
	if info, err := os.Lstat(path); err == nil {
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("dependency cache directory must be real: %s", path)
		}
		return nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if err := dependencyDirectory(filepath.Dir(path)); err != nil {
		return err
	}
	return os.Mkdir(path, 0700)
}

func copyDependencyCache(source, destination string) (int64, error) {
	rootInfo, err := os.Lstat(source)
	if err != nil {
		return 0, err
	}
	if !rootInfo.IsDir() || rootInfo.Mode()&os.ModeSymlink != 0 {
		return 0, fmt.Errorf("dependency cache source must be a real directory: %s", source)
	}
	var paths []string
	var total int64
	err = filepath.WalkDir(source, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == source {
			return nil
		}
		if !entry.IsDir() && !entry.Type().IsRegular() {
			return fmt.Errorf("dependency cache contains a nonregular entry: %s", path)
		}
		name := entry.Name()
		if strings.HasPrefix(name, ".l-") || name == "inputs.json" {
			return fmt.Errorf("dependency cache contains project ownership state: %s", path)
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		return 0, err
	}
	sort.Strings(paths)
	for _, path := range paths {
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return 0, err
		}
		info, err := os.Lstat(path)
		if err != nil {
			return 0, err
		}
		out := filepath.Join(destination, relative)
		if info.IsDir() {
			if err := os.Mkdir(out, 0700); err != nil {
				return 0, err
			}
			continue
		}
		if !info.Mode().IsRegular() {
			return 0, fmt.Errorf("dependency cache source changed type: %s", path)
		}
		if err := copyDependencyProduct(path, out, info); err != nil {
			return 0, err
		}
		total += info.Size()
	}
	return total, nil
}

func copyDependencyProduct(source, destination string, info fs.FileInfo) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	output, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, info.Mode().Perm())
	if err != nil {
		return errors.Join(err, input.Close())
	}
	_, copyErr := io.Copy(output, input)
	if err := errors.Join(copyErr, output.Close(), input.Close()); err != nil {
		return err
	}
	return os.Chtimes(destination, info.ModTime(), info.ModTime())
}

func (r *runner) dependencyToolPath(name string) (string, error) {
	output, err := r.command("dependency-"+name+"-path", time.Minute, nil, "rustup", "which", "--toolchain", r.pins.Rust, name)
	if err != nil {
		return "", err
	}
	path := strings.TrimSpace(string(output))
	if !filepath.IsAbs(path) || strings.ContainsAny(path, "\r\n") {
		return "", fmt.Errorf("pinned rustup returned an invalid %s path", name)
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 {
		return "", fmt.Errorf("pinned %s must be an executable regular file", name)
	}
	return path, nil
}

func rejectDependencyBuildOverrides(environment []string) error {
	var conflicts []string
	for _, entry := range environment {
		key, value, _ := strings.Cut(entry, "=")
		// Empty flag/wrapper values conventionally disable these overrides.
		optional := key == "RUSTFLAGS" || key == "RUSTDOCFLAGS" || key == "CARGO_ENCODED_RUSTFLAGS" || key == "CARGO_ENCODED_RUSTDOCFLAGS" || key == "RUSTC_WRAPPER" || key == "RUSTC_WORKSPACE_WRAPPER"
		// Cargo numeric/path settings are not safely unset by an empty value.
		strict := key == "RUSTC" || key == "RUSTDOC" || key == "CARGO_TARGET_DIR" || strings.HasPrefix(key, "CARGO_TARGET_") || strings.HasPrefix(key, "CARGO_PROFILE_") || strings.HasPrefix(key, "CARGO_BUILD_")
		if strict || (optional && value != "") {
			conflicts = append(conflicts, key)
		}
	}
	if len(conflicts) != 0 {
		sort.Strings(conflicts)
		return fmt.Errorf("unset Rust/Cargo build overrides before dependency setup to match isolated compiler profiles: %s", strings.Join(conflicts, ", "))
	}
	return nil
}

func dependencyInputHashes(roots ...string) (map[string]string, error) {
	inputs := map[string]string{}
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				return nil
			}
			if !entry.Type().IsRegular() {
				return fmt.Errorf("dependency input is not regular: %s", path)
			}
			digest, err := testgate.FileHash(path)
			if err != nil {
				return err
			}
			inputs[path] = digest
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return inputs, nil
}

func unchangedDependencyInputs(inputs map[string]string) error {
	var failures []error
	for path, digest := range inputs {
		failures = append(failures, requireUnchanged(path, digest))
	}
	return errors.Join(failures...)
}
