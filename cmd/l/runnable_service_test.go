//go:build darwin || linux

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStandaloneServiceSourceStructure(t *testing.T) {
	project := t.TempDir()
	filename := filepath.Join(project, "main.l")
	if err := os.WriteFile(filename, []byte("Create \"Worker\" as a New Main Service.\n"), 0600); err != nil {
		t.Fatal(err)
	}
	c := newCompiler(&cliData{jobType: "build", mainPath: filename, osType: "ubuntu"})
	c.initialize()
	compiled := c.compile()
	if c.err != nil || compiled == nil {
		t.Fatalf("structural standalone Service: %v", c.err)
	}
	record := compiled.Compilation.Memory["Worker"]
	if record == nil || record.Value == nil || record.Value.Applicative == nil || record.Value.Applicative.Service == nil {
		t.Fatal("standalone Service lost its semantic record")
	}
	units, err := compiled.ApplicationUnits()
	if err != nil || len(units) != 0 {
		t.Fatalf("structural Service acquired application units: %v %v", units, err)
	}
	for _, name := range []string{"main", "main.apps", "Cargo.toml", "target"} {
		if _, err := os.Stat(filepath.Join(project, name)); !os.IsNotExist(err) {
			t.Fatalf("source compilation created native output %s: %v", name, err)
		}
	}
}
