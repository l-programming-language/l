package main

import (
	"path/filepath"
	"testing"
)

// The ordinary unfiltered gate regenerates AML using the current frontend and
// source resources. Matching hashes alone cannot establish this relationship.
func TestCanonicalRuntimeCompositionIsCurrent(t *testing.T) {
	if err := run(filepath.Join("..", ".."), true); err != nil {
		t.Fatal(err)
	}
}
