package Logiceme

import (
	"os"
	"testing"
)

func TestStructuralScanInterleavedWhitespace(t *testing.T) {
	input, err := os.ReadFile("testdata/classes/interleaved.aml.txt")
	if err != nil {
		t.Fatal(err)
	}
	scanner, err := loadScannerBytes(input)
	if err != nil {
		t.Fatal(err)
	}
	if err := scanner.Scan(); err != nil {
		t.Fatalf("scan interleaved classes: %v; masked document: %s", err, scanner.ProgramBytes)
	}
}
