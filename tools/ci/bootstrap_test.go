package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/l-programming-language/l/tools/internal/testgate"
)

func TestLockedDependencyCoverageIncludesDistinctSelectionVersions(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	seeds, err := prepareSeeds(root, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(seeds) != 4 {
		t.Fatalf("seed count %d", len(seeds))
	}
	covered := map[lockedDependency]bool{}
	for _, seed := range seeds {
		data, err := os.ReadFile(filepath.Join(seed, "Cargo.lock"))
		if err != nil {
			t.Fatal(err)
		}
		deps, err := lockDependencies(data)
		if err != nil {
			t.Fatal(err)
		}
		for dep := range deps {
			covered[dep] = true
		}
	}
	versions := map[string]bool{}
	for dep := range covered {
		if dep.Name == "syn" {
			versions[dep.Version] = true
		}
	}
	if !versions["3.0.3"] || !versions["3.0.6"] {
		t.Fatalf("lost independently locked versions: %v", versions)
	}
	locks, err := lockedInputs(root)
	if err != nil {
		t.Fatal(err)
	}
	for name := range locks {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		deps, err := lockDependencies(data)
		if err != nil {
			t.Fatal(err)
		}
		for dep := range deps {
			if !covered[dep] {
				t.Errorf("%s lacks coverage for %+v", name, dep)
			}
		}
	}
}

func TestLockParserRejectsUnsupportedOrAmbiguousPackages(t *testing.T) {
	valid := "version = 4\n[[package]]\nname = \"dependency\"\nversion = \"1.0.0\"\nsource = \"registry+https://example.test/index\"\nchecksum = \"" + strings.Repeat("a", 64) + "\"\n"
	if deps, err := lockDependencies([]byte(valid)); err != nil || len(deps) != 1 {
		t.Fatalf("valid lock: %v %v", deps, err)
	}
	for name, data := range map[string]string{"empty": "", "missing-name": strings.Replace(valid, "name = \"dependency\"\n", "", 1), "duplicate": valid + "name = \"other\"\n", "git-source": strings.Replace(valid, "registry+https", "git+https", 1), "checksum": strings.Replace(valid, strings.Repeat("a", 64), "bad", 1), "unquoted": strings.Replace(valid, "\"1.0.0\"", "1.0.0", 1)} {
		t.Run(name, func(t *testing.T) {
			if _, err := lockDependencies([]byte(data)); err == nil {
				t.Fatal("invalid lock accepted")
			}
		})
	}
}

func TestSeedCopyRejectsSymlinksAndPreservesBytes(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source")
	dest := filepath.Join(dir, "nested", "copy")
	data := []byte("bytes\r\n\x00")
	testWrite(t, source, data)
	if err := copyFile(source, dest); err != nil {
		t.Fatal(err)
	}
	a, err := testgate.FileHash(source)
	if err != nil {
		t.Fatal(err)
	}
	b, err := testgate.FileHash(dest)
	if err != nil || a != b {
		t.Fatal("seed copy changed bytes", err)
	}
	link := filepath.Join(dir, "link")
	if err := os.Symlink(source, link); err != nil {
		t.Fatal(err)
	}
	if err := copyFile(link, dest); err == nil {
		t.Fatal("symlink seed accepted")
	}
}

func TestCoverageRejectsEveryUncoveredExternalTuple(t *testing.T) {
	base := lockedDependency{Name: "dependency", Version: "1.0.0", Source: "registry+https://example.test/index", Checksum: strings.Repeat("a", 64)}
	covered := map[lockedDependency]bool{base: true}
	if err := requireCoverage(covered, covered, "fixture.lock"); err != nil {
		t.Fatal(err)
	}
	for _, missing := range []lockedDependency{
		{Name: "uncovered", Version: base.Version, Source: base.Source, Checksum: base.Checksum},
		{Name: base.Name, Version: "2.0.0", Source: base.Source, Checksum: base.Checksum},
		{Name: base.Name, Version: base.Version, Source: "registry+https://other.test/index", Checksum: base.Checksum},
		{Name: base.Name, Version: base.Version, Source: base.Source, Checksum: strings.Repeat("b", 64)},
	} {
		if err := requireCoverage(covered, map[lockedDependency]bool{missing: true}, "new.lock"); err == nil {
			t.Fatalf("uncovered tuple accepted: %+v", missing)
		}
	}
}

func TestFailedFetchStillRejectsChangedCopiedLock(t *testing.T) {
	root := t.TempDir()
	seed := filepath.Join(root, "seed")
	lock := filepath.Join(seed, "Cargo.lock")
	testWrite(t, lock, []byte("retained lock bytes"))
	bin := filepath.Join(root, "bin")
	cargo := filepath.Join(bin, "cargo")
	testWrite(t, cargo, []byte("#!/bin/sh\nprintf changed > \"$TEST_SEED_LOCK\"\nexit 9\n"))
	if err := os.Chmod(cargo, 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
	t.Setenv("TEST_SEED_LOCK", lock)
	r := runner{root: root, directory: t.TempDir(), pins: toolPins{Rust: "fixture"}}
	err := r.fetchSeed(seed)
	if err == nil || !strings.Contains(err.Error(), "exit 9") || !strings.Contains(err.Error(), "changed retained input") {
		t.Fatalf("fetch failure hid lock mutation: %v", err)
	}
	if len(r.report.Commands) != 1 {
		t.Fatal("offline fetch must not proceed after locked fetch failure")
	}
}

func TestGoRacePreparationCompilesLibrariesWithoutRunningTests(t *testing.T) {
	for _, failAt := range []int{0, 1, 2} {
		t.Run([]string{"success", "standard-library-failure", "project-library-failure"}[failAt], func(t *testing.T) {
			calls := 0
			childError := errors.New("race library compiler failed")
			err := prepareGoRaceDependencies(func(name string, timeout time.Duration, extra map[string]string, executable string, args ...string) ([]byte, error) {
				calls++
				targets := []string{"std", "./..."}
				names := []string{"go-race-standard-library", "go-race-project-libraries"}
				if calls > len(targets) || executable != "go" || strings.Join(args, " ") != "build -mod=readonly -race "+targets[calls-1] || name != names[calls-1] {
					t.Fatalf("setup must compile race libraries without building or running tests: %s %q", executable, args)
				}
				if len(extra) != 1 || extra["GOPROXY"] != "off" || timeout <= 0 {
					t.Fatal("setup must compile downloaded modules offline in the pinned Go environment")
				}
				if calls == failAt {
					return nil, childError
				}
				return nil, nil
			})
			if failAt == 0 {
				if calls != 2 || err != nil {
					t.Fatalf("incomplete library preparation: calls=%d error=%v", calls, err)
				}
			} else if calls != failAt || !errors.Is(err, childError) {
				t.Fatalf("preparation lost child failure or continued after it: calls=%d error=%v", calls, err)
			}
		})
	}
}
