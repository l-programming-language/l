//go:build darwin || linux

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/l-programming-language/l/internal/testsupport/phase1"
)

func TestNativeEvidenceRecordsActualBuildWithoutRunningApp(t *testing.T) {
	source := contractCLIInput(t)
	project := filepath.Dir(source)
	output := filepath.Join(project, "native-evidence.json")
	stdout, stderr, err := runBuildCommand(t, phase1.CLI(t), project, nil, "build", "--native-evidence-output", output, source)
	if err != nil {
		t.Fatalf("evidence-producing native build failed: %v\n%s\n%s", err, stdout, stderr)
	}
	environment := readBuildEnvironment(t, output)
	binary := strings.TrimSpace(string(stdout))
	generation := environment.Configuration["companion-generation"]
	counts := map[string]int{}
	for _, artifact := range environment.Artifacts {
		counts[artifact.SHA256]++
	}
	for _, file := range []string{binary, filepath.Join(binary+".apps", "generations", generation, "app-000")} {
		data := readContractFile(t, file)
		counts[nativeBytesHash(data)]--
	}
	for _, count := range counts {
		if count != 0 {
			t.Fatalf("artifact inventory differs from published binaries: %v", counts)
		}
	}
	// Consume the actual record under the same kernel process/network restrictions
	// used for missing, hostile and test-only evidence.
	runOfflineHelper(t, "L_CONTRACTS_BUILD_RECORD="+output, "L_CONTRACTS_BUILD_SOURCE="+source)
	builds := 0
	for _, execution := range environment.Executions {
		if execution.Kind == "build" {
			builds++
			if execution.Status != "passed" || !slices.Contains(execution.Command, "--frozen") || execution.Validate() != nil {
				t.Fatalf("build evidence is not the actual frozen invocation: %+v", execution)
			}
		} else if execution.Kind != "environment" {
			t.Fatalf("evidence collection executed an unexpected operation: %+v", execution)
		}
	}
	if builds != 1 {
		t.Fatalf("evidence caused %d Cargo builds, want one", builds)
	}
	// Both destinations pass preflight. A missing evidence parent fails after
	// compilation, before the candidate generation becomes visible.

	old := nativeBytesHash(readContractFile(t, binary))
	missing := filepath.Join(project, "absent", "record.json")
	_, stderr, err = runBuildCommand(t, phase1.CLI(t), project, nil, "build", "--native-evidence-output", missing, source)
	entries, readErr := os.ReadDir(filepath.Join(binary+".apps", "generations"))
	if err == nil || !strings.Contains(string(stderr), "write requested native evidence") || readErr != nil || len(entries) != 1 || entries[0].Name() != generation || nativeBytesHash(readContractFile(t, binary)) != old || readBuildEnvironment(t, output).Fingerprint != environment.Fingerprint {
		t.Fatalf("late evidence failure disturbed an installed generation: %v %s", err, stderr)
	}
}

func TestNativeEvidenceRetainsFailedCargoInvocation(t *testing.T) {
	cargo, err := exec.LookPath("cargo")
	if err != nil {
		t.Fatal(err)
	}
	cargo, err = filepath.Abs(cargo)
	if err != nil {
		t.Fatal(err)
	}
	for _, failCargo := range []bool{true, false} {
		t.Run(map[bool]string{true: "cargo", false: "publish"}[failCargo], func(t *testing.T) {
			project := t.TempDir()
			output := filepath.Join(project, "entry")
			generation, entry := stagedPublication(t, project, "old")
			if err := publishGeneration(output, output+".apps", "entry", "old", generation, entry, os.Rename); err != nil {
				t.Fatal(err)
			}
			tools := t.TempDir()
			// Forward every version probe. The publication case uses real Cargo and
			// adds a colliding nonempty generation only after successful compilation.
			action := `echo 'test-only controlled build failure' >&2; exit 7`
			if !failCargo {
				action = `
"$REAL_CARGO" "$@" || exit $?
stage=$(pwd -P)
stage=${stage%/workspace}
dir="$TEST_GENERATIONS/${stage##*-build-}"
mkdir "$dir" && touch "$dir/occupied" || exit 1
exit 0`
			}
			script := "#!/bin/sh\nfor arg do\n if [ \"$arg\" = build ]; then\n" + action + "\n fi\ndone\nexec \"$REAL_CARGO\" \"$@\"\n"
			launcher := writeContractFile(t, filepath.Join(tools, "cargo"), []byte(script))
			if err := os.Chmod(launcher, 0700); err != nil {
				t.Fatal(err)
			}
			source := writeContractFile(t, filepath.Join(project, "entry.l"), []byte("Exit.\n"))
			record := filepath.Join(project, "native-evidence.json")
			env := []string{"PATH=" + tools + string(os.PathListSeparator) + os.Getenv("PATH"), "REAL_CARGO=" + cargo, "TEST_GENERATIONS=" + filepath.Join(output+".apps", "generations")}
			stdout, stderr, buildErr := runBuildCommand(t, phase1.CLI(t), project, env, "build", "--native-evidence-output", record, source)
			old := readContractFile(t, output)
			if buildErr == nil || len(stdout) != 0 || string(old) != "old" {
				t.Fatalf("failed build replaced previous publication: %v %s", buildErr, stderr)
			}
			environment := readBuildEnvironment(t, record)
			last := environment.Executions[len(environment.Executions)-1]
			if failCargo {
				if !strings.Contains(string(stderr), "test-only controlled build failure") || len(environment.Artifacts) != 0 || last.Kind != "build" || last.Status != "failed" || last.ExitCode == nil || *last.ExitCode != 7 {
					t.Fatalf("wrong evidence for failed Cargo invocation: %v, %s", last, stderr)
				}
			} else if !strings.Contains(string(stderr), "publish output") || last.Kind != "build" || last.Status != "passed" || len(environment.Artifacts) == 0 {
				t.Fatalf("compilation evidence disappeared after failed publication: %v, %s", last, stderr)
			}
		})
	}
}
