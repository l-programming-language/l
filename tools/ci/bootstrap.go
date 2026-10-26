package main

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/l-programming-language/l/tools/internal/testgate"
)

type lockedDependency struct{ Name, Version, Source, Checksum string }

// Only registry packages enter the download coverage comparison. The retained
// lock format has explicit quoted scalar identities; unsupported inputs fail.
func lockDependencies(data []byte) (map[lockedDependency]bool, error) {
	result := map[lockedDependency]bool{}
	var current map[string]string
	packages := 0
	finish := func() error {
		if current == nil {
			return nil
		}
		packages++
		if current["name"] == "" || current["version"] == "" {
			return errors.New("lock package lacks name/version")
		}
		if current["source"] == "" {
			return nil
		}
		checksum, err := hex.DecodeString(current["checksum"])
		if !strings.HasPrefix(current["source"], "registry+") || err != nil || len(checksum) != 32 {
			return fmt.Errorf("unsupported locked source/checksum for %s", current["name"])
		}
		result[lockedDependency{current["name"], current["version"], current["source"], current["checksum"]}] = true
		return nil
	}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "[[package]]" {
			if err := finish(); err != nil {
				return nil, err
			}
			current = map[string]string{}
			continue
		}
		if current == nil || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		if !ok || (key != "name" && key != "version" && key != "source" && key != "checksum") {
			continue
		}
		if _, exists := current[key]; exists {
			return nil, fmt.Errorf("duplicate lock field %s", key)
		}
		decoded, err := strconv.Unquote(strings.TrimSpace(value))
		if err != nil {
			return nil, fmt.Errorf("lock field %s: %w", key, err)
		}
		current[key] = decoded
	}
	if err := errors.Join(scanner.Err(), finish()); err != nil {
		return nil, err
	}
	if packages == 0 {
		return nil, errors.New("lock has no packages")
	}
	return result, nil
}

func readLockDependencies(path string) (map[lockedDependency]bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return lockDependencies(data)
}

func (r *runner) bootstrap() (resultErr error) {
	locks, err := lockedInputs(r.root)
	if err != nil {
		return err
	}
	protected := map[string]string{}
	for name, digest := range locks {
		protected[filepath.Join(r.root, name)] = digest
	}
	for _, name := range []string{"go.mod", "go.sum"} {
		digest, err := testgate.FileHash(filepath.Join(r.root, name))
		if err != nil {
			return err
		}
		protected[filepath.Join(r.root, name)] = digest
	}
	defer func() { resultErr = errors.Join(resultErr, unchangedDependencyInputs(protected)) }()
	if err := r.installLinters(); err != nil {
		return err
	}
	cEnvironment, err := r.command("go-cgo-environment", time.Minute, nil, "go", "env", "-json", "CC", "CGO_ENABLED")
	if err != nil {
		return err
	}
	var cgo struct {
		CC         string
		CGOEnabled string `json:"CGO_ENABLED"`
	}
	if err := json.Unmarshal(cEnvironment, &cgo); err != nil {
		return err
	}
	compiler := strings.Fields(cgo.CC)
	if cgo.CGOEnabled != "1" || len(compiler) == 0 {
		return errors.New("native acceptance requires CGO_ENABLED=1 and a C compiler for race-instrumented API tests")
	}
	if _, err := r.command("c-compiler-version", 15*time.Second, nil, compiler[0], append(compiler[1:], "--version")...); err != nil {
		return err
	}
	if _, err := r.command("go-download", 3*time.Minute, nil, "go", "mod", "download"); err != nil {
		return err
	}
	if _, err := r.command("rust-install", 5*time.Minute, nil, "rustup", "toolchain", "install", r.pins.Rust, "--profile", "minimal", "--no-self-update"); err != nil {
		return err
	}
	if _, err := r.command("rust-version", 15*time.Second, nil, "rustup", "run", r.pins.Rust, "rustc", "-vV"); err != nil {
		return err
	}
	seeds, err := prepareSeeds(r.root, filepath.Join(r.directory, "seeds"))
	if err != nil {
		return err
	}
	coverage := map[lockedDependency]bool{}
	for _, seed := range seeds {
		lock := filepath.Join(seed, "Cargo.lock")
		dependencies, err := readLockDependencies(lock)
		if err != nil {
			return err
		}
		for dep := range dependencies {
			coverage[dep] = true
		}
	}
	for name := range locks {
		dependencies, err := readLockDependencies(filepath.Join(r.root, name))
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		if err := requireCoverage(coverage, dependencies, name); err != nil {
			return err
		}
	}
	for _, seed := range seeds {
		if err := r.fetchSeed(seed); err != nil {
			return err
		}
	}
	if err := prepareGoRaceDependencies(r.command); err != nil {
		return err
	}
	if err := r.prepareRustDependencies(seeds); err != nil {
		return err
	}
	after, err := lockedInputs(r.root)
	if err != nil {
		return err
	}
	if len(locks) != len(after) {
		return errors.New("bootstrap changed lock inventory")
	}
	for name, digest := range locks {
		if after[name] != digest {
			return fmt.Errorf("bootstrap changed tracked lock %s", name)
		}
	}
	return nil
}

// Prepare the race-instrumented libraries used by the separate API probe before
// its build watchdog starts. The probe and its tests are still freshly compiled
// and run by acceptance; setup never builds a test binary or executes code.
func prepareGoRaceDependencies(command func(string, time.Duration, map[string]string, string, ...string) ([]byte, error)) error {
	for _, target := range []struct{ name, packages string }{
		{"go-race-standard-library", "std"},
		{"go-race-project-libraries", "./..."},
	} {
		if _, err := command(target.name, 5*time.Minute, map[string]string{"GOPROXY": "off"}, "go", "build", "-mod=readonly", "-race", target.packages); err != nil {
			return err
		}
	}
	return nil
}

func requireUnchanged(path, before string) error {
	after, err := testgate.FileHash(path)
	if err != nil {
		return err
	}
	if before != after {
		return fmt.Errorf("bootstrap changed retained input %s", path)
	}
	return nil
}

func requireCoverage(coverage, dependencies map[lockedDependency]bool, name string) error {
	for dependency := range dependencies {
		if !coverage[dependency] {
			return fmt.Errorf("bootstrap seeds do not cover %s %s from %s", dependency.Name, dependency.Version, name)
		}
	}
	return nil
}

func (r *runner) fetchSeed(seed string) (resultErr error) {
	name := filepath.Base(seed)
	lock := filepath.Join(seed, "Cargo.lock")
	before, err := testgate.FileHash(lock)
	if err != nil {
		return err
	}
	defer func() { resultErr = errors.Join(resultErr, requireUnchanged(lock, before)) }()
	if _, err := r.command("fetch-"+name, 5*time.Minute, map[string]string{"CARGO_NET_OFFLINE": "false", "RUSTUP_AUTO_INSTALL": "0"}, "cargo", "+"+r.pins.Rust, "fetch", "--locked", "--manifest-path", filepath.Join(seed, "Cargo.toml")); err != nil {
		return err
	}
	_, err = r.command("offline-"+name, time.Minute, map[string]string{"RUSTUP_AUTO_INSTALL": "0"}, "cargo", "+"+r.pins.Rust, "fetch", "--frozen", "--manifest-path", filepath.Join(seed, "Cargo.toml"))
	return err
}

func lockedInputs(root string) (map[string]string, error) {
	files, err := repositoryFiles(root)
	if err != nil {
		return nil, err
	}
	result := map[string]string{}
	for _, name := range files {
		if strings.HasSuffix(name, ".lock") {
			digest, err := testgate.FileHash(filepath.Join(root, name))
			if err != nil {
				return nil, err
			}
			result[name] = digest
		}
	}
	if len(result) == 0 {
		return nil, errors.New("no retained locks found")
	}
	return result, nil
}

func prepareSeeds(root, directory string) ([]string, error) {
	runtimeSeed := filepath.Join(directory, "runtime")
	err := filepath.WalkDir(filepath.Join(root, "lyger", "test", "native"), func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(filepath.Join(root, "lyger", "test", "native"), path)
		if err != nil {
			return err
		}
		return copyFile(path, filepath.Join(runtimeSeed, relative))
	})
	if err != nil {
		return nil, err
	}
	seeds := []string{runtimeSeed}
	base := filepath.Join(root, "logiceme", "testdata", "lucid")
	for _, seed := range []struct{ name, manifest, lock, source string }{{"plain", "Cargo.toml.txt", "Cargo.lock", "grouping.rs.txt"}, {"tokio", "Cargo-tokio.toml.txt", "Cargo-tokio.lock", "communication.rs.txt"}, {"selection", "Cargo-select.toml.txt", "Cargo-select.lock", "select-communication.rs.txt"}} {
		path := filepath.Join(directory, seed.name)
		for source, dest := range map[string]string{filepath.Join("common", seed.manifest): "Cargo.toml", filepath.Join("common", seed.lock): "Cargo.lock", seed.source: "src/main.rs", "common/config.rs.txt": "src/config.rs"} {
			if err := copyFile(filepath.Join(base, source), filepath.Join(path, dest)); err != nil {
				return nil, err
			}
		}
		seeds = append(seeds, path)
	}
	sort.Strings(seeds)
	return seeds, nil
}
