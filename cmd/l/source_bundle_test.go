package main

import (
	"errors"
	"testing"

	"github.com/l-programming-language/l/lerrors"
)

func TestLibraryTargetOwnsItsKind(t *testing.T) {
	manifest := []byte("[package]\nname = \"service\"\nversion = \"0.0.0\"\nedition = \"2024\"\nautobins = false\n\n[lib]\npath = \"src/main.rs\"\n")
	bundle, err := planSourceBundle([][]byte{manifest, []byte("pub fn value() -> u8 { 1 }\n")}, []string{"Cargo.toml", "src/main.rs"})
	if err != nil || bundle.packagePlan.binary != "" || bundle.packagePlan.library != "src/main.rs" {
		t.Fatalf("source filename overrode library target kind: %#v %v", bundle, err)
	}
}

func TestSourceBundleRejectsGoOnlyInput(t *testing.T) {
	bundle, err := planSourceBundle([][]byte{[]byte("package main\nfunc main() {}\n")}, []string{"/main.go"})
	if bundle != nil || !errors.Is(err, lerrors.InvalidInput) {
		t.Fatalf("Go-only source bundle accepted: bundle=%+v error=%v", bundle, err)
	}
}
