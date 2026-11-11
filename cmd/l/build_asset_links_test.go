package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	distribution "github.com/l-programming-language/l"
	distributiontest "github.com/l-programming-language/l/internal/testsupport/distribution"
)

func TestBuildRejectsResealedRuntime(t *testing.T) {
	for _, warm := range []bool{false, true} {
		for _, missing := range []bool{false, true} {
			t.Run(fmt.Sprintf("warm=%t/missing=%t", warm, missing), func(t *testing.T) {
				assets, project := distributiontest.CompleteOverride(t), t.TempDir()
				main := writeContractFile(t, filepath.Join(project, "main.l"), []byte("Show `freshness`.\n"))
				// A valid bundle reaches this guard without invoking native tools.
				writeContractFile(t, filepath.Join(project, "main"), []byte("unrelated destination"))
				c, options := pathTestCompiler(t, assets, main, project)
				if _, err := buildExecutable(context.Background(), c, options); err == nil || !strings.Contains(err.Error(), "unrelated executable") {
					t.Fatalf("valid provider failed before destination control: %v", err)
				}
				if warm {
					generated := c.process()
					if c.err != nil {
						t.Fatal(c.err)
					}
					if _, err := generated.workspace(c.log, "main", "control", nil); err != nil {
						t.Fatal(err)
					}
				}
				manifest, aml, err := distribution.LoadRuntimeComposition(os.DirFS(assets))
				if err != nil {
					t.Fatal(err)
				}
				const source = "stdlib/eng/runtime/errors.l"
				path := filepath.Join(assets, filepath.FromSlash(source))
				if missing {
					if err := os.Remove(path); err != nil {
						t.Fatal(err)
					}
					delete(manifest.Sources, source)
				} else {
					raw := bytes.ReplaceAll(readContractFile(t, path), []byte("source execution stopped"), []byte("changed source cancellation"))
					writeContractFile(t, path, raw)
					manifest.Sources[source] = fmt.Sprintf("%x", sha256.Sum256(raw))
				}
				manifest.SHA256, err = manifest.Fingerprint(aml)
				if err != nil {
					t.Fatal(err)
				}
				metadata, err := json.Marshal(manifest)
				if err != nil {
					t.Fatal(err)
				}
				writeContractFile(t, filepath.Join(assets, "stdlib/runtime/generated/support.json"), metadata)
				if _, _, err := distribution.LoadRuntimeComposition(os.DirFS(assets)); err != nil {
					t.Fatalf("resealed envelope must be a valid integrity control: %v", err)
				}
				if _, err := buildExecutable(context.Background(), c, options); err == nil || (!strings.Contains(err.Error(), "stale") && !strings.Contains(err.Error(), source)) {
					t.Fatalf("build did not reject stale source composition before native setup: %v", err)
				}
				if !missing {
					composeScopedTestProvider(t, assets)
					if _, err := buildExecutable(context.Background(), c, options); err == nil || !strings.Contains(err.Error(), "unrelated executable") {
						t.Fatalf("explicitly regenerated provider failed before destination control: %v", err)
					}
				}
			})
		}
	}
}

func TestBuildProtectsLinkedResourceFiles(t *testing.T) {
	for _, tc := range []struct {
		name, resource string
		directoryLink  bool
		hardLink       bool
	}{
		{"dictionary index", "dictionaries.json", false, false},
		{"grammar", "lexicon/eng/grammar.json", false, false},
		{"runtime host adapter", "stdlib/runtime/rust/host.rs", false, false},
		{"nested directory", "stdlib/runtime/rust/host.rs", true, false},
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
