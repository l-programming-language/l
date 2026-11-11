//go:build darwin || linux

package main

import (
	"bytes"
	"context"

	"fmt"
	"os"
	"path/filepath"

	"testing"
	"time"
)

func TestCargoWorkspaceBuildsLibraryAndBinaries(t *testing.T) {
	makeBundle := func(name, kind, source, dependencies string) *sourceBundle {
		t.Helper()
		manifest := fmt.Sprintf("[package]\nname = %q\nversion = \"0.0.0\"\nedition = \"2024\"\n", name)
		file := "src/main.rs"
		if kind == "lib" {
			manifest += "\n[lib]\npath = \"src/lib.rs\"\n"
			file = "src/lib.rs"
		}
		manifest += dependencies
		bundle, err := planSourceBundle([][]byte{[]byte(manifest), []byte(source)}, []string{"Cargo.toml", file})
		if err != nil {
			t.Fatal(err)
		}
		return bundle
	}
	app := makeBundle("application", "bin", "fn main() { assert_eq!(service::value(), 7); }\n", "\n[dependencies]\nservice = { path = \"../../services/service-000\" }\n")
	service := makeBundle("service", "lib", "pub fn value() -> u8 { 7 }\n", "")
	entry := makeBundle("entry", "bin", "fn main() {}\n", "")
	directory := t.TempDir()
	workspace := filepath.Join(directory, "workspace")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	if err := compileCargoWorkspace(ctx, []*sourceBundle{app}, entry, []*sourceBundle{service}, workspace, []string{filepath.Join(directory, "app")}, filepath.Join(directory, "entry")); err != nil {
		t.Fatal(err)
	}
	for _, member := range []struct {
		prefix string
		bundle *sourceBundle
	}{{"apps/app-000", app}, {"entry", entry}, {"services/service-000", service}} {
		for _, file := range member.bundle.files {
			got, err := os.ReadFile(filepath.Join(workspace, member.prefix, file.name))
			if err != nil || !bytes.Equal(got, file.content) {
				t.Fatalf("workspace changed %s/%s: %v", member.prefix, file.name, err)
			}
		}
	}
	for _, name := range []string{"app", "entry"} {
		info, err := os.Stat(filepath.Join(directory, name))
		if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 {
			t.Fatalf("missing Cargo artifact %s: %v", name, err)
		}
	}
}
