// Package testgate shares acceptance commands, completion evidence and artifact helpers.
package testgate

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

func FileHash(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(data)), nil
}

func WriteJSON(path string, data any) error {
	encoded, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(encoded, '\n'), 0644)
}

type Package struct {
	Package  string `json:"package"`
	Phase    string `json:"phase"`
	HasTests *bool  `json:"has_tests"`
}

type Outcome struct {
	Package string `json:"package"`
	Name    string `json:"name,omitempty"`
	Status  string `json:"status"`
}

type Events struct {
	Tests      []Outcome `json:"tests"`
	Packages   []Outcome `json:"packages"`
	ParseError string    `json:"parse_error,omitempty"`
}

const (
	BuildTimeout         = 180 * time.Second
	DiscoveryTimeout     = 60 * time.Second
	PackageTimeout       = 120 * time.Second
	PhaseTimeout         = 360 * time.Second
	NativePackageTimeout = 600 * time.Second
	NativePhaseTimeout   = 900 * time.Second
)

type Command struct {
	Name    string
	Args    []string
	Timeout time.Duration
}

// ReadPackages decodes ordinary go list output without inferring a module or phase.
func ReadPackages(input io.Reader) (map[string]bool, error) {
	packages := make(map[string]bool)
	decoder := json.NewDecoder(input)
	for {
		var pkg struct {
			ImportPath                string
			TestGoFiles, XTestGoFiles []string
		}
		if err := decoder.Decode(&pkg); err != nil {
			if errors.Is(err, io.EOF) && len(packages) != 0 {
				return packages, nil
			}
			return nil, fmt.Errorf("decode discovered packages: %w", err)
		}
		if _, exists := packages[pkg.ImportPath]; exists || pkg.ImportPath == "" {
			return nil, fmt.Errorf("empty or duplicate discovered package %q", pkg.ImportPath)
		}
		packages[pkg.ImportPath] = len(pkg.TestGoFiles)+len(pkg.XTestGoFiles) != 0
	}
}

// Commands constructs the fixed gate, omitting only phases with no assigned packages.
// CheckPhase validates the inventory and its resulting evidence.
func Commands(packages []Package, parallel int) []Command {
	phases := map[string][]string{"tests": nil, "tests-native": nil}
	for _, pkg := range packages {
		phases[pkg.Phase] = append(phases[pkg.Phase], pkg.Package)
	}
	packageParallel := parallel
	if packageParallel > 3 {
		packageParallel = 3
	}
	commands := []Command{{"build", []string{"build", "-mod=readonly", "./..."}, BuildTimeout}, {"packages", []string{"list", "-mod=readonly", "-json", "./..."}, DiscoveryTimeout}}
	for _, phase := range []string{"tests", "tests-native"} {
		if len(phases[phase]) == 0 {
			continue
		}
		packageTimeout, timeout := PackageTimeout, PhaseTimeout
		if phase == "tests-native" {
			packageTimeout, timeout = NativePackageTimeout, NativePhaseTimeout
		}
		args := []string{"test", fmt.Sprintf("-p=%d", packageParallel), fmt.Sprintf("-parallel=%d", parallel), "-mod=readonly", "-count=1", fmt.Sprintf("-timeout=%ds", int(packageTimeout/time.Second)), "-json"}
		commands = append(commands, Command{phase, append(args, phases[phase]...), timeout})
	}
	return commands
}

// CheckPhase requires complete successful outcomes, including explicit empty named evidence.
func CheckPhase(packages []Package, phase string, events *Events) error {
	if events == nil || events.Tests == nil || events.ParseError != "" {
		return fmt.Errorf("%s has missing or malformed test events", phase)
	}
	expected := make(map[string]bool)
	seen := make(map[string]bool)
	for _, pkg := range packages {
		if pkg.Phase != "tests" && pkg.Phase != "tests-native" || pkg.Package == "" || pkg.HasTests == nil || seen[pkg.Package] {
			return errors.New("invalid or duplicate package inventory")
		}
		seen[pkg.Package] = true
		if pkg.Phase == phase {
			expected[pkg.Package] = *pkg.HasTests
		}
	}
	named := make(map[string]bool)
	for _, test := range events.Tests {
		key := test.Package + "\x00" + test.Name
		if !expected[test.Package] || test.Name == "" || test.Status != "pass" || named[key] {
			return fmt.Errorf("%s has duplicate, unexpected or unsuccessful test %+v", phase, test)
		}
		named[key] = true
	}
	for _, pkg := range events.Packages {
		hasTests, exists := expected[pkg.Package]
		if !exists || pkg.Status != "pass" && (pkg.Status != "skip" || hasTests) {
			return fmt.Errorf("%s has duplicate, unexpected or unsuccessful package %+v", phase, pkg)
		}
		delete(expected, pkg.Package)
	}
	if len(expected) != 0 {
		return fmt.Errorf("%s is missing %d package outcomes", phase, len(expected))
	}
	return nil
}
