package Lyger

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestStructuralServicesPreservePrivateAppBuildNames(t *testing.T) {
	source := "Create \"Worker_1\" as a New Main Service.\n\n    Create \"Worker\" as a New Basic App.\n\n    Create \"Worker\" as a New Basic App.\n"
	result := runnableServiceRequire(t, source, nil)
	plan, err := result.WorkspacePlan()
	if err != nil {
		t.Fatalf("structural Service must not change private App workspace packages: %v", err)
	}
	packages := plan.Packages()
	if len(packages) != 2 {
		t.Fatalf("workspace packages = %#v, want only two Apps", packages)
	}
	seen := make(map[string]bool)
	for _, item := range packages {
		key := strings.ToLower(item.ID)
		if seen[key] {
			t.Fatalf("generated App build names collide: %#v", packages)
		}
		seen[key] = true
	}
	if result.Compilation.Memory["Worker_1"].ID != "Worker_1" {
		t.Fatal("workspace collision handling changed the source Service identity")
	}
	appUnits, err := result.ApplicationUnits()
	var apps []string
	for _, unit := range appUnits {
		apps = append(apps, unit.ID)
	}
	if err != nil || !reflect.DeepEqual(apps, []string{"Worker_1", "Worker_2"}) {
		t.Fatalf("structural Service changed established private App build names: %v, %v", apps, err)
	}
}

func TestStructuralServicesPreserveCasefoldAppNames(t *testing.T) {
	result := runnableServiceRequire(t, "Create \"Worker\" as a New Basic App.\nCreate \"worker\" as a New Main Service.\n", nil)
	plan, err := result.WorkspacePlan()
	if err != nil {
		t.Fatal(err)
	}
	packages := plan.Packages()
	if len(packages) != 1 || packages[0].ID != "Worker" || packages[0].Library {
		t.Fatalf("structural Service must preserve the App package without allocating a library: %#v", packages)
	}
	if result.Compilation.Memory["worker"].ID != "worker" {
		t.Fatal("casefold collision handling changed the source Service identity")
	}
}

func TestStructuralServicesPreserveAppOwnedLibraries(t *testing.T) {
	source, err := os.ReadFile("test/workspace/services.l")
	if err != nil {
		t.Fatal(err)
	}
	result := runnableServiceRequire(t, string(source)+"\nCreate \"l_service_4669727374_1\" as a New Main Service.\n", nil)
	plan, err := result.WorkspacePlan()
	if err != nil {
		t.Fatalf("structural Service changed App-owned library projection: %v", err)
	}
	packages := plan.Packages()
	wantExisting := []WorkspacePackage{{"First", false}, {"l_service_4669727374_1", true}, {"l_service_4669727374_2", true}, {"Second", false}, {"l_service_5365636f6e64_1", true}, {"l_service_5365636f6e64_2", true}}
	if len(packages) != 6 || !reflect.DeepEqual(packages, wantExisting) {
		t.Fatalf("structural Service changed existing App-owned package identities: %#v", packages)
	}
}

func TestWorkspacePreservesCompletedAppGraphAPI(t *testing.T) {
	result := runnableServiceRequire(t, "Create \"Worker\" as a New Basic App.\n", nil)
	result.Compilation.Semantics = nil
	apps, err := result.ApplicationUnits()
	if err != nil || len(apps) != 1 {
		t.Fatalf("completed App graph prerequisite: %v", err)
	}
	plan, err := result.WorkspacePlan()
	if err != nil {
		t.Fatalf("App-only workspace projection should require only the completed App graph: %v", err)
	}
	units, err := plan.Units(map[string]string{"Worker": "apps/worker"})
	if err != nil || len(units) != 1 || units[0].Library || !reflect.DeepEqual(units[0].BuildUnit, apps[0]) {
		t.Fatalf("App-only workspace unit differs from the canonical App projection: %#v, %v", units, err)
	}
}
