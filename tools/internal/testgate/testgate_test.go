package testgate

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestCommandsOmitOnlyEmptyPhases(t *testing.T) {
	hasTests := true
	pkg := Package{Package: "example.test/l/cmd/l", Phase: "tests-native", HasTests: &hasTests}
	got := Commands([]Package{pkg}, 4)
	want := []Command{
		{Name: "build", Args: []string{"build", "-mod=readonly", "./..."}, Timeout: 180 * time.Second},
		{Name: "packages", Args: []string{"list", "-mod=readonly", "-json", "./..."}, Timeout: 60 * time.Second},
		{Name: "tests-native", Args: []string{"test", "-p=3", "-parallel=4", "-mod=readonly", "-count=1", "-timeout=600s", "-json", "example.test/l/cmd/l"}, Timeout: 900 * time.Second},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("native-only commands: %+v", got)
	}
	if got := Commands(nil, 4); !reflect.DeepEqual(got, want[:2]) {
		t.Fatalf("pre-discovery commands: %+v", got)
	}
}

func TestEvidencePreservesMissingFields(t *testing.T) {
	var pkg Package
	if err := json.Unmarshal([]byte(`{"package":"p","phase":"tests"}`), &pkg); err != nil {
		t.Fatal(err)
	}
	if err := CheckPhase([]Package{pkg}, "tests", &Events{Tests: []Outcome{}}); err == nil {
		t.Fatal("missing has_tests accepted")
	}
	hasTests := true
	pkg.HasTests = &hasTests
	for _, data := range []string{`{"packages":[{"package":"p","status":"pass"}]}`, `{"tests":null,"packages":[{"package":"p","status":"pass"}]}`, `{"tests":[],"packages":[{"package":"p","status":"pass"}]}`} {
		var events Events
		if err := json.Unmarshal([]byte(data), &events); err != nil {
			t.Fatal(err)
		}
		if err := CheckPhase([]Package{pkg}, "tests", &events); (err == nil) != (events.Tests != nil) {
			t.Fatalf("named evidence presence lost: %s: %v", data, err)
		}
	}
	data, err := json.Marshal(Outcome{Package: "p", Status: "pass"})
	if err != nil || string(data) != `{"package":"p","status":"pass"}` {
		t.Fatalf("package outcome JSON changed: %s, %v", data, err)
	}
}
