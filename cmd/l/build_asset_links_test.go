package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	distributiontest "github.com/l-programming-language/l/internal/testsupport/distribution"
)

func TestBuildProtectsLinkedResourceFiles(t *testing.T) {
	for _, tc := range []struct {
		name, resource string
		directoryLink  bool
		hardLink       bool
	}{
		{"dictionary index", "stdlib/index.json", false, false},
		{"grammar", "lexicon/eng/grammar.json", false, false},
		{"runtime host adapter", "builtins/templates/runtime/host.lu", false, false},
		{"nested directory", "builtins/templates/runtime/host.lu", true, false},
		{"hard link", "stdlib/index.json", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assets := distributiontest.CompleteOverride(t)
			resource := filepath.Join(assets, filepath.FromSlash(tc.resource))
			original, err := os.ReadFile(resource)
			if err != nil {
				t.Fatal(err)
			}
			outside := t.TempDir()
			output := filepath.Join(outside, filepath.Base(resource))
			if tc.directoryLink {
				if err := os.Remove(outside); err != nil {
					t.Fatal(err)
				}
				if err := os.Rename(filepath.Dir(resource), outside); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(outside, filepath.Dir(resource)); err != nil {
					t.Fatal(err)
				}
			} else if tc.hardLink {
				if err := os.Link(resource, output); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := os.Rename(resource, output); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(output, resource); err != nil {
					t.Fatal(err)
				}
			}
			// Existing owned publication metadata must not grant permission to
			// replace a selected compiler resource through another spelling.
			companion := output + ".apps"
			if err := os.MkdirAll(filepath.Join(companion, "generations"), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(companion, ".l-owner"), []byte(companionOwner+filepath.Base(output)+"\n"), 0600); err != nil {
				t.Fatal(err)
			}
			main := filepath.Join(t.TempDir(), filepath.Base(output)+".l")
			source := "Use IO's Print from `io` as \"Print\".\nCreate \"Worker\" as a New Basic App.\nThe App Prints `linked resource`.\nExecute Worker.\n"
			if err := os.WriteFile(main, []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			c, options := pathTestCompiler(t, assets, main, outside)
			_, err = buildExecutable(context.Background(), c, options)
			if err == nil || !strings.Contains(err.Error(), "compiler assets") {
				t.Fatalf("linked compiler resource destination accepted or wrong prerequisite failed: %v", err)
			}
			after, err := os.ReadFile(output)
			if err != nil || !bytes.Equal(after, original) {
				t.Fatalf("linked resource changed: %v", err)
			}
		})
	}
}
