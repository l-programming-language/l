package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/l-programming-language/l/tools/internal/testgate"
)

// Exercise the public CLI against a disposable module, never the protected suite.
func integrationFixture(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "checkout with spaces")
	for _, dir := range []string{"harness", ".harness/cache"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0755); err != nil {
			t.Fatal(err)
		}
	}
	input := []byte("package fixture\nimport \"testing\"\nfunc TestOriginal(t *testing.T) {}\n")
	for name, data := range map[string][]byte{
		"go.mod":           []byte("module example.com/harness-fixture\n\ngo 1.20\n"),
		"fixture.go":       []byte("package fixture\n"),
		"original_test.go": input,
		"guide.md":         []byte("[Original test](original_test.go)\n"),
	} {
		if err := os.WriteFile(filepath.Join(root, name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	c := contract{
		SchemaVersion: 2, Module: "example.com/harness-fixture",
		ProtectedFiles: []fileContract{{Path: "original_test.go", SHA256: fmt.Sprintf("%x", sha256.Sum256(input))}},
		Tests:          []testContract{{Package: ".", Name: "TestOriginal"}},
	}
	if err := testgate.WriteJSON(filepath.Join(root, "harness/contract.json"), c); err != nil {
		t.Fatal(err)
	}
	// Reuse only build cache entries; the module, reports, and acceptance data are isolated.
	cmd := exec.Command("go", "env", "GOCACHE")
	cmd.Env = append(os.Environ(), "GOENV=off", "GOFLAGS=", "GOTOOLCHAIN=local")
	cache, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(strings.TrimSpace(string(cache)), filepath.Join(root, ".harness/cache/go-build")); err != nil {
		t.Fatal(err)
	}
	return root
}

func invokeIntegrationCLI(t *testing.T, args ...string) (int, runReport) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	code := runCLI(ctx, args, &stdout, &stderr)
	var reportPath string
	for _, line := range strings.Split(stdout.String(), "\n") {
		if strings.HasPrefix(line, "report: ") {
			reportPath = strings.TrimPrefix(line, "report: ")
		}
	}
	if reportPath == "" {
		t.Fatalf("no report for %v (exit %d): stdout=%s stderr=%s", args, code, &stdout, &stderr)
	}
	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatal(err)
	}
	var report runReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if report.ExitCode != code {
		t.Fatalf("CLI exit %d differs from recorded exit %d", code, report.ExitCode)
	}
	if report.SchemaVersion != 2 || bytes.Contains(data, []byte(`"baseline_revision"`)) || bytes.Contains(data, []byte(`"baseline_annotations"`)) {
		t.Fatalf("report did not use current-only schema 2: %s", data)
	}
	return code, report
}

func TestHarnessCLIIntegration(t *testing.T) {
	root := integrationFixture(t)
	caller := filepath.Join(t.TempDir(), "caller with spaces")
	if err := os.MkdirAll(caller, 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("L_HARNESS_ROOT", root)
	t.Setenv("L_HARNESS_CALLER_DIR", caller)
	t.Setenv("L_HARNESS_ORIGINAL_GOFLAGS", "")
	t.Setenv("GOFLAGS", "")
	t.Setenv("GOENV", "off")
	if code, report := invokeIntegrationCLI(t, "doctor"); code != 0 || !report.IntegrityAfter.OK || report.AcceptancePassed != nil {
		t.Fatalf("doctor must pass integrity without claiming acceptance: code=%d report=%+v", code, report)
	}
	for _, name := range []string{"expected raw", "actual raw"} {
		if err := os.WriteFile(filepath.Join(caller, name), []byte("hello\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if code, report := invokeIntegrationCLI(t, "diff", "expected raw", "actual raw"); code != 0 || report.Difference == nil || !report.Difference.Equal {
		t.Fatalf("equal relative files must resolve from caller directory: code=%d report=%+v", code, report)
	}
	if err := os.WriteFile(filepath.Join(caller, "actual raw"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	if code, report := invokeIntegrationCLI(t, "diff", "expected raw", "actual raw"); code != 1 || report.Difference == nil || report.Difference.FirstDifference == nil || *report.Difference.FirstDifference != 5 {
		t.Fatalf("trailing newline difference must fail at byte 5: code=%d report=%+v", code, report)
	}
	if code, _ := invokeIntegrationCLI(t, "diff", "expected raw", "missing"); code != 2 {
		t.Fatalf("missing input should be a setup error, got %d", code)
	}
	for _, tc := range []struct{ effective, parallelism int }{{1, 1}, {2, 2}, {3, 3}, {4, 3}, {8, 3}} {
		t.Run(fmt.Sprintf("verify with effective GOMAXPROCS %d", tc.effective), func(t *testing.T) {
			// The current runtime value, not the caller's environment spelling, determines the budget.
			t.Setenv("GOMAXPROCS", "7")
			previous := runtime.GOMAXPROCS(tc.effective)
			t.Cleanup(func() { runtime.GOMAXPROCS(previous) })
			code, report := invokeIntegrationCLI(t, "verify")
			if code != 0 || report.AcceptancePassed == nil || !*report.AcceptancePassed || len(report.Commands) != 3 || len(report.OriginalTests) != 1 || report.OriginalTests[0].Status != "pass" {
				t.Fatalf("source export without Git or archives must pass actual build/tests: code=%d report=%+v", code, report)
			}
			if report.Environment["effective_gomaxprocs"] != strconv.Itoa(tc.effective) || report.Environment["test_package_parallelism"] != strconv.Itoa(tc.parallelism) || report.Environment["test_case_parallelism"] != strconv.Itoa(tc.effective) || report.Environment["original_gomaxprocs"] != "7" || report.Environment["num_cpu"] != strconv.Itoa(runtime.NumCPU()) {
				t.Fatalf("scheduling environment was not recorded accurately: %+v", report.Environment)
			}
			for key, want := range map[string]string{"build_command_timeout_ms": "180000", "test_package_timeout_ms": "120000", "test_command_timeout_ms": "360000", "package_discovery_timeout_ms": "60000", "native_test_package_timeout_ms": "600000", "native_test_command_timeout_ms": "900000"} {
				if report.Environment[key] != want {
					t.Fatalf("%s=%q, want %q", key, report.Environment[key], want)
				}
			}
			wantArgs := [][]string{
				{"build", "-mod=readonly", "./..."},
				{"list", "-mod=readonly", "-json", "./..."},
				{"test", "-p=" + strconv.Itoa(tc.parallelism), "-parallel=" + strconv.Itoa(tc.effective), "-mod=readonly", "-count=1", "-timeout=120s", "-json", "example.com/harness-fixture"},
			}
			wantTimeouts := []int64{180000, 60000, 360000}
			for i, command := range report.Commands {
				if command.TimeoutMS != wantTimeouts[i] {
					t.Fatalf("command %s timeout=%d, want %d", command.Name, command.TimeoutMS, wantTimeouts[i])
				}
				if !reflect.DeepEqual(command.Args, wantArgs[i]) {
					t.Fatalf("command %s changed the ordinary gate: got %v, want %v", command.Name, command.Args, wantArgs[i])
				}
				for _, arg := range command.Args {
					if strings.HasPrefix(arg, "-overlay") || strings.HasPrefix(arg, "-run") || strings.HasPrefix(arg, "-vet") {
						t.Fatalf("acceptance used a diagnostic/filter flag: %v", command.Args)
					}
				}
			}
		})
	}
	if err := os.WriteFile(filepath.Join(root, "original_test.go"), []byte("changed acceptance\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if code, report := invokeIntegrationCLI(t, "verify"); code != 2 || len(report.Commands) != 0 || report.AcceptancePassed == nil || *report.AcceptancePassed {
		t.Fatalf("contract drift must stop acceptance before any commands: code=%d report=%+v", code, report)
	}
}

func TestHarnessCLIRecordsSchedulingEnvironment(t *testing.T) {
	root := integrationFixture(t)
	t.Setenv("L_HARNESS_ROOT", root)
	t.Setenv("L_HARNESS_ORIGINAL_GOFLAGS", "")
	t.Setenv("GOFLAGS", "")
	t.Setenv("GOENV", "off")
	previous := runtime.GOMAXPROCS(3)
	t.Cleanup(func() { runtime.GOMAXPROCS(previous) })
	for _, tc := range []struct {
		name, value string
		present     bool
	}{
		{name: "unset"},
		{name: "empty", present: true},
		{name: "explicit", value: "2", present: true},
		{name: "unusual spelling", value: " 04 ", present: true},
		{name: "non-numeric", value: "caller-value", present: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("GOMAXPROCS", tc.value)
			if !tc.present {
				if err := os.Unsetenv("GOMAXPROCS"); err != nil {
					t.Fatal(err)
				}
			}
			code, report := invokeIntegrationCLI(t, "doctor")
			if code != 0 || len(report.Commands) != 0 || report.AcceptancePassed != nil {
				t.Fatalf("doctor changed execution semantics: code=%d report=%+v", code, report)
			}
			value, present := report.Environment["original_gomaxprocs"]
			if value != tc.value || present != tc.present {
				t.Errorf("original GOMAXPROCS=(%q, %t), want (%q, %t)", value, present, tc.value, tc.present)
			}
			if report.Environment["effective_gomaxprocs"] != "3" || report.Environment["num_cpu"] != strconv.Itoa(runtime.NumCPU()) {
				t.Errorf("effective runtime values missing: %+v", report.Environment)
			}
			if _, present := report.Environment["test_package_parallelism"]; present {
				t.Error("doctor reported a verify-only package budget")
			}
		})
	}
	for name, source := range map[string]string{
		"cmd/l/fixture_test.go": "package main\nimport \"testing\"\nfunc TestCompilerInitialize(t *testing.T) {}\nfunc TestCompilerCompile(t *testing.T) {}\n",
		"lyger/fixture_test.go": "package lyger\nimport \"testing\"\nfunc Test_importScan(t *testing.T) {}\n",
	} {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
	}
	code, report := invokeIntegrationCLI(t, "diagnose")
	if code != 0 || report.AcceptancePassed != nil || len(report.Commands) != 4 {
		t.Fatalf("diagnose failed or claimed acceptance: code=%d report=%+v", code, report)
	}
	if _, present := report.Environment["test_package_parallelism"]; present {
		t.Error("diagnose reported a verify-only package budget")
	}
	wantArgs := [][]string{
		{"test", "-mod=readonly", "-count=1", "-timeout=60s", "-json", "-run", "^TestCompilerInitialize$", "./cmd/l"},
		{"test", "-mod=readonly", "-count=1", "-timeout=60s", "-json", "-run", "^TestCompilerCompile$", "./cmd/l"},
		{"test", "-mod=readonly", "-count=1", "-timeout=60s", "-json", "-run", "^Test_importScan$", "./lyger"},
		{"vet", "-mod=readonly", "./..."},
	}
	for i, command := range report.Commands {
		if !reflect.DeepEqual(command.Args, wantArgs[i]) {
			t.Errorf("diagnostic command %s changed: got %v, want %v", command.Name, command.Args, wantArgs[i])
		}
	}
}

func TestHarnessCLIPartitionsAllPackages(t *testing.T) {
	root := integrationFixture(t)
	t.Setenv("L_HARNESS_ROOT", root)
	t.Setenv("L_HARNESS_ORIGINAL_GOFLAGS", "")
	t.Setenv("GOFLAGS", "")
	t.Setenv("GOENV", "off")
	for name, source := range map[string]string{
		"cmd/l/native_test.go":   "package main\nimport \"testing\"\nfunc TestNative(t *testing.T) {}\n",
		"future/helpers_test.go": "package future\nfunc helper() {}\n",
		"none/none.go":           "package none\n",
	} {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
	}
	c, _, err := readContract(root)
	if err != nil {
		t.Fatal(err)
	}
	c.Tests = append(c.Tests, testContract{Package: "./cmd/l", Name: "TestNative"})
	if err := testgate.WriteJSON(filepath.Join(root, "harness/contract.json"), c); err != nil {
		t.Fatal(err)
	}
	code, report := invokeIntegrationCLI(t, "verify")
	hasTests, noTests := true, false
	want := []testgate.Package{{Package: c.Module, Phase: "tests", HasTests: &hasTests}, {Package: c.Module + "/cmd/l", Phase: "tests-native", HasTests: &hasTests}, {Package: c.Module + "/future", Phase: "tests", HasTests: &hasTests}, {Package: c.Module + "/none", Phase: "tests", HasTests: &noTests}}
	if code != 0 || !reflect.DeepEqual(report.TestPackages, want) || len(report.Commands) != 4 || len(report.OriginalTests) != 2 {
		t.Fatalf("partition failed: code=%d report=%+v", code, report)
	}
	for _, original := range report.OriginalTests {
		if original.Status != "pass" {
			t.Fatalf("original across phases did not pass: %+v", original)
		}
	}
	ordinary, native := report.Commands[2], report.Commands[3]
	if ordinary.Name != "tests" || !reflect.DeepEqual(ordinary.Args[7:], []string{c.Module, c.Module + "/future", c.Module + "/none"}) || native.Name != "tests-native" || native.TimeoutMS != 900000 || native.Args[5] != "-timeout=600s" || !reflect.DeepEqual(native.Args[7:], []string{c.Module + "/cmd/l"}) {
		t.Fatalf("wrong phase commands: %+v %+v", ordinary, native)
	}
	if err := os.WriteFile(filepath.Join(root, "cmd/l/native_test.go"), []byte("package main\nimport \"testing\"\nfunc TestNative(t *testing.T) { t.Fatal(\"native fixture failure\") }\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if code, report := invokeIntegrationCLI(t, "verify"); code != 1 || report.AcceptancePassed == nil || *report.AcceptancePassed || len(report.Commands) != 4 || report.OriginalTests[1].Status != "fail" {
		t.Fatalf("one failed phase became acceptance: code=%d report=%+v", code, report)
	}
}

func TestHarnessLauncherIgnoresPersistedOverlay(t *testing.T) {
	root := integrationFixture(t)
	sourceDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "scripts"), 0755); err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{"harness", "internal/testgate"} {
		target := filepath.Join(root, "tools", dir)
		if err := os.MkdirAll(target, 0755); err != nil {
			t.Fatal(err)
		}
		source := filepath.Join(sourceDir, "..", dir)
		files, err := os.ReadDir(source)
		if err != nil {
			t.Fatal(err)
		}
		for _, file := range files {
			if !strings.HasSuffix(file.Name(), ".go") || strings.HasSuffix(file.Name(), "_test.go") {
				continue
			}
			data, err := os.ReadFile(filepath.Join(source, file.Name()))
			if err != nil {
				t.Fatal(err)
			}
			data = bytes.ReplaceAll(data, []byte("github.com/l-programming-language/l/tools/internal/testgate"), []byte("example.com/harness-fixture/tools/internal/testgate"))
			if err := os.WriteFile(filepath.Join(target, file.Name()), data, 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
	launcher, err := os.ReadFile(filepath.Join(sourceDir, "../../scripts/l-harness"))
	if err != nil {
		t.Fatal(err)
	}
	launcherPath := filepath.Join(root, "scripts/l-harness")
	if err := os.WriteFile(launcherPath, launcher, 0755); err != nil {
		t.Fatal(err)
	}
	mainPath := filepath.Join(root, "tools/harness/main.go")
	mainSource, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(root, "overlay-was-executed")
	injected := strings.Replace(string(mainSource), "func main() {", fmt.Sprintf("func main() {\n os.WriteFile(%q, []byte(\"unexpected\"), 0600); os.Exit(0)", marker), 1)
	if injected == string(mainSource) {
		t.Fatal("could not prepare bootstrap overlay")
	}
	alternate := filepath.Join(root, "alternate-main.go.txt")
	if err := os.WriteFile(alternate, []byte(injected), 0644); err != nil {
		t.Fatal(err)
	}
	overlay := filepath.Join(root, "overlay.json")
	if err := testgate.WriteJSON(overlay, map[string]interface{}{"Replace": map[string]string{mainPath: alternate}}); err != nil {
		t.Fatal(err)
	}
	goEnv := filepath.Join(root, "go-env")
	// Quote the path inside GOFLAGS because this disposable checkout has spaces.
	if err := os.WriteFile(goEnv, []byte(fmt.Sprintf("GOFLAGS=%q\n", "-overlay="+overlay)), 0644); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, launcherPath, "doctor")
	cmd.Dir = t.TempDir()
	cmd.Env = append(os.Environ(), "GOENV="+goEnv, "GOFLAGS=", "L_HARNESS_ORIGINAL_GOFLAGS=")
	output, runErr := cmd.CombinedOutput()
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("persisted overlay changed launcher bootstrap (status %v): %s", runErr, output)
	}
	// Both explicit rejection and a clean isolated run are safe; neither may fake success.
	if runErr != nil {
		if exit, ok := runErr.(*exec.ExitError); !ok || exit.ExitCode() != 2 {
			t.Fatalf("launcher should reject flags or run isolated, got %v: %s", runErr, output)
		}
	}
	if !strings.Contains(string(output), "report:") {
		t.Fatalf("real harness did not produce a report: %s", output)
	}
}
