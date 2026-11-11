package Lyger

import (
	"bytes"
	"testing"
)

func TestZeroApplicationProgramRetainsEmptyViews(t *testing.T) {
	source := memberResolutionFixture(t)
	source.main = "\"message\" is `scalar only`.\nShow message.\n"
	result := executableTestResult(t, source)
	program := result.Compilation.Program
	if len(program.Apps) != 0 || len(program.Books) != 0 {
		t.Fatal("scalar program invented a companion App")
	}
	units, err := result.ApplicationUnits()
	if err != nil || len(units) != 0 {
		t.Fatalf("zero companion units: %v, %v", units, err)
	}
	want := []byte("<PROGRAM>\n<STRUCTURE>\n</STRUCTURE>\n</PROGRAM>")
	if !bytes.Equal(result.AML, want) || !bytes.Equal(result.BackendAML, want) {
		t.Fatalf("empty application views = %q / %q", result.AML, result.BackendAML)
	}

}

func TestZeroApplicationGraphStillRejectsInconsistency(t *testing.T) {
	if apps, err := validateBuildGraph(&Program{}); err != nil || len(apps) != 0 {
		t.Fatalf("valid empty graph: %v, %v", apps, err)
	}
	for _, tc := range []struct {
		name    string
		program *Program
	}{
		{"nil", nil},
		{"missing book", &Program{Apps: []*App{{ID: "Worker"}}}},
		{"missing app", &Program{Books: []*book{{Name: "Worker"}}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := validateBuildGraph(tc.program); err == nil {
				t.Fatal("incomplete graph accepted")
			}
		})
	}
}
