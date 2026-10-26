package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadmeBuildSourcesExist(t *testing.T) {
	root := filepath.Join("..", "..")
	readme, err := os.ReadFile(filepath.Join(root, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(readme), "\n") {
		if !strings.HasPrefix(line, "./bin/l build ") && !strings.HasPrefix(line, "go run ./cmd/l build ") {
			continue
		}
		// These examples use unquoted repository-relative paths. Inspect their
		// source operand, never execute documentation as shell code.
		fields := strings.Fields(line)
		source := fields[len(fields)-1]
		if !strings.HasSuffix(source, ".l") {
			t.Errorf("build example has no L source operand: %s", line)
			continue
		}
		// A project needs its source and declared local imports, not an adjacent
		// copy of the language. Native distribution tests compile these examples.
		info, err := os.Stat(filepath.Join(root, source))
		if err != nil {
			t.Errorf("build example %q references missing input %s: %v", line, source, err)
		} else if !info.Mode().IsRegular() {
			t.Errorf("build example input is not a regular file: %s", source)
		}
	}
}
