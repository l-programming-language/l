package Lyger

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestWorkspaceConstructedServices(t *testing.T) {
	source, err := os.ReadFile("test/workspace/services.l")
	if err != nil {
		t.Fatal(err)
	}
	before := bytes.Clone(source)
	cfg := canonicalTestOptions()
	result, err := NewLygerService(cfg).Compile(source, "test/workspace/services.l")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(source, before) {
		t.Fatal("workspace preparation changed source")
	}
	apps := result.Compilation.Program.Apps
	if len(apps) != 2 || len(apps[0].Services) != 3 || len(apps[1].Services) != 3 {
		t.Fatal("constructed Service graph lost applications or Services")
	}
	for index := range apps[0].Services {
		if apps[0].Services[index] == apps[1].Services[index] {
			t.Fatal("descriptor reuse collapsed distinct constructed Service identities")
		}
	}
	originalAML, originalBackend := bytes.Clone(result.AML), bytes.Clone(result.BackendAML)
	plan, err := result.WorkspacePlan()
	if err != nil {
		t.Fatal(err)
	}
	want := []WorkspacePackage{{"First", false}, {"l_service_4669727374_1", true}, {"l_service_4669727374_2", true}, {"Second", false}, {"l_service_5365636f6e64_1", true}, {"l_service_5365636f6e64_2", true}}
	if !reflect.DeepEqual(plan.Packages(), want) {
		t.Fatalf("source-ordered workspace identities = %#v", plan.Packages())
	}
	locations := map[string]string{}
	appIndex, libraryIndex := 0, 0
	for _, item := range plan.Packages() {
		if item.Library {
			locations[item.ID] = fmt.Sprintf("services/service-%03d", libraryIndex)
			libraryIndex++
		} else {
			locations[item.ID] = fmt.Sprintf("apps/app-%03d", appIndex)
			appIndex++
		}
	}
	units, err := plan.Units(locations)
	if err != nil || len(units) != len(want) {
		t.Fatalf("workspace units: %v", err)
	}
	for index, unit := range units {
		if unit.ID != want[index].ID || unit.Library != want[index].Library || bytes.Count(unit.BackendAML, []byte("<BOOK ")) != 1 || bytes.Count(unit.BackendAML, []byte("<CHAPTER ")) != 1 {
			t.Fatalf("workspace unit %d lost its exact package role/Service", index)
		}
		own, other := "first-service-body", "second-service-body"
		if index >= 3 {
			own, other = other, own
		}
		if !bytes.Contains(unit.BackendAML, []byte(own)) || bytes.Contains(unit.BackendAML, []byte(other)) {
			t.Fatal("same descriptor confused App-specific Process bindings")
		}
		if unit.Library && (!bytes.Contains(unit.BackendAML, []byte(`"Kind":"lib"`)) || !bytes.Contains(unit.BackendAML, []byte("<SETTINGS>\n</SETTINGS>"))) {
			t.Fatal("library target acquired executable settings")
		}
	}
	units[0].BackendAML[0] = '!'
	again, err := plan.Units(locations)
	if err != nil || again[0].BackendAML[0] != '<' || !bytes.Equal(result.AML, originalAML) || !bytes.Equal(result.BackendAML, originalBackend) || !bytes.Equal(result.Compilation.Program.BackendBytes, originalBackend) {
		t.Fatal("workspace rendering mutated the source graph or an earlier result")
	}
	for _, app := range apps {
		if strings.Contains(app.Meta["Deps"], "path") || len(app.Services) != 3 {
			t.Fatal("workspace projection replaced observable application ownership")
		}
	}
	for _, bad := range []map[string]string{nil, {"First": "../escape"}} {
		if _, err := plan.Units(bad); err == nil {
			t.Fatal("incomplete workspace locations accepted")
		}
	}
}
