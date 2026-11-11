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
		{"dictionary index", "dictionaries.json", false, false},
		{"grammar", "lexicon/eng/grammar.json", false, false},
		{"core", "builtins/eng/core.l", false, false},
		{"nested directory", "builtins/eng/core.l", true, false},
		{"hard link", "dictionaries.json", false, true},
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
			source := "Create \"Worker\" as a New Basic App.\nThe App Prints `linked resource`.\n"
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

func TestAssetDestinationsIgnoreRetiredRuntimeRoot(t *testing.T) {
	assets := distributiontest.CompleteOverride(t)
	output := filepath.Join(t.TempDir(), "program")
	if err := validateAssetDestinations(assets, output); err != nil {
		t.Fatalf("complete provider control: %v", err)
	}
	if err := os.Symlink(filepath.Join(t.TempDir(), "missing"), filepath.Join(assets, "runtime")); err != nil {
		t.Fatal(err)
	}
	if err := validateAssetDestinations(assets, output); err != nil {
		t.Fatalf("retired root affected the selected provider: %v", err)
	}
}
