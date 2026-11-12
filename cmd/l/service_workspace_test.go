//go:build darwin || linux

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	distributiontest "github.com/l-programming-language/l/internal/testsupport/distribution"
	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
)

func TestConstructedServicesCompileAsLibraries(t *testing.T) {
	parallelIndependentBuildRoot(t)
	root := distributiontest.RepoRoot(t)
	compiler := newCompiler(&cliData{jobType: "build", mainPath: filepath.Join(root, "lyger/test/workspace/services.l"), osType: "ubuntu"})
	compiler.initialize(compilerInitOptions{AssetRoot: root})
	result := compiler.compile()
	if compiler.err != nil || result == nil {
		t.Fatalf("compile Service source: %v", compiler.err)
	}
	plan, err := result.WorkspacePlan()
	if err != nil {
		t.Fatal(err)
	}
	locations := make(map[string]string)
	appCount, libraryCount := 0, 0
	for _, item := range plan.Packages() {
		if item.Library {
			locations[item.ID] = fmt.Sprintf("services/service-%03d", libraryCount)
			libraryCount++
		} else {
			locations[item.ID] = fmt.Sprintf("apps/app-%03d", appCount)
			appCount++
		}
	}
	if appCount != 2 || libraryCount != 4 {
		t.Fatalf("workspace dimensions: %d applications, %d libraries", appCount, libraryCount)
	}
	units, err := plan.Units(locations)
	if err != nil {
		t.Fatal(err)
	}
	var apps, libraries []*sourceBundle
	var libraryIDs []string
	for _, unit := range units {
		generated, err := compiler.log.Evaluate(unit.BackendAML)
		if err != nil {
			t.Fatal(err)
		}
		nativetest.RecordBundle(t, "workspace-"+unit.ID, generated)
		bundle, err := workspaceBundle(generated)
		if err != nil {
			t.Fatal(err)
		}
		if !unit.Library {
			apps = append(apps, bundle)
			continue
		}
		if bundle.packagePlan.name != unit.ID || bundle.packagePlan.binary != "" || bundle.packagePlan.library == "" {
			t.Fatalf("Service %s is not a library-only package", unit.ID)
		}
		for _, file := range bundle.files {
			if file.name != "Cargo.toml" && !strings.HasSuffix(file.name, ".rs") {
				t.Fatalf("library contains executable support %s", file.name)
			}
		}
		libraries = append(libraries, bundle)
		libraryIDs = append(libraryIDs, unit.ID)
	}
	// An independent native caller checks the actual exported library surface;
	// generated library modules are not repaired or wrapped by the test.
	manifest := fmt.Sprintf("[package]\nname = \"service_consumer\"\nversion = \"0.0.0\"\nedition = \"2024\"\n\n[dependencies]\nleft = { package = %q, path = \"../services/service-000\" }\n", libraryIDs[0])
	entry, err := planSourceBundle([][]byte{[]byte(manifest), []byte("fn main() { left::left::left(); }\n")}, []string{"Cargo.toml", "src/main.rs"})
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	outputs := []string{filepath.Join(directory, "first"), filepath.Join(directory, "second")}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	if err := compileCargoWorkspace(ctx, apps, entry, libraries, filepath.Join(directory, "workspace"), outputs, filepath.Join(directory, "consumer")); err != nil {
		t.Fatal(err)
	}
	for index, name := range []string{"first", "second", "consumer"} {
		output, err := exec.CommandContext(ctx, filepath.Join(directory, name)).CombinedOutput()
		want := "first-service-body\n"
		if index == 1 {
			want = "second-service-body\n"
		}
		if err != nil || string(output) != want {
			t.Fatalf("native %s output = %q, %v; want %q", name, output, err, want)
		}
	}
	for index, library := range libraries {
		for _, file := range library.files {
			actual, err := os.ReadFile(filepath.Join(directory, "workspace", fmt.Sprintf("services/service-%03d", index), filepath.FromSlash(file.name)))
			if err != nil || string(actual) != string(file.content) {
				t.Fatalf("workspace changed generated library bytes: %v", err)
			}
		}
	}
}
