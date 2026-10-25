package main

import (
	"bytes"
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	distributiontest "github.com/l-programming-language/l/internal/testsupport/distribution"
)

func TestBuildRejectsCanonicalAssetDestinations(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"output symlink", "asset symlink", "missing output descendant", "stdlib", "stdlib symlink"} {
		t.Run(mode, func(t *testing.T) {
			installation := distributiontest.CompleteOverride(t)
			project := filepath.Join(t.TempDir(), "project")
			copyBuildProject(t, root, project, "main.l", "")
			protected := filepath.Join(installation, "lexicon")
			if strings.HasPrefix(mode, "stdlib") {
				protected = filepath.Join(installation, "stdlib")
			}
			output := filepath.Join(t.TempDir(), "output-alias")
			if mode == "asset symlink" || mode == "stdlib symlink" {
				physical := filepath.Join(t.TempDir(), "shared-assets")
				if err := os.Rename(protected, physical); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(physical, protected); err != nil {
					t.Fatal(err)
				}
				protected, output = physical, physical
			} else {
				if err := os.Symlink(protected, output); err != nil {
					t.Fatal(err)
				}
				if mode == "missing output descendant" {
					output = filepath.Join(output, "not-created", "nested")
				}
			}
			before := assetTree(t, protected)
			compiler, options := pathTestCompiler(t, installation, filepath.Join(project, "main.l"), output)
			path, err := buildExecutable(context.Background(), compiler, options)
			if err == nil || !strings.Contains(err.Error(), "compiler assets") {
				t.Fatalf("compiler asset destination was not rejected: path=%s error=%v", path, err)
			}
			if after := assetTree(t, protected); !reflect.DeepEqual(before, after) {
				t.Fatal("rejected build changed compiler assets")
			}
		})
	}
}

func assetTree(t testing.TB, root string) map[string]string {
	t.Helper()
	contents := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			relative, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			contents[relative+"/"] = "directory"
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		contents[relative] = string(data)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return contents
}

func TestBuildRejectsCanonicalSourceDestination(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	installation := distributiontest.CompleteOverride(t)
	project := filepath.Join(t.TempDir(), "project")
	copyBuildProject(t, root, project, "library.l.l", "")
	main := filepath.Join(project, "library.l.l")
	content, err := os.ReadFile(main)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(main, append([]byte("Use `library.l`.\n"), content...), 0600); err != nil {
		t.Fatal(err)
	}
	input := filepath.Join(project, "library.l")
	inputBytes := []byte("\"extra\" is 123.\n")
	if err := os.WriteFile(input, inputBytes, 0600); err != nil {
		t.Fatal(err)
	}
	// The source protection must still apply when old l-owned metadata permits
	// replacing a destination, and the output directory uses another spelling.
	companion := input + ".apps"
	if err := os.MkdirAll(filepath.Join(companion, "generations"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(companion, ".l-owner"), []byte(companionOwner+"library.l\n"), 0600); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(t.TempDir(), "project-alias")
	if err := os.Symlink(project, alias); err != nil {
		t.Fatal(err)
	}
	compiler, options := pathTestCompiler(t, installation, main, alias)
	path, err := buildExecutable(context.Background(), compiler, options)
	if err == nil || !strings.Contains(err.Error(), "source input") {
		t.Fatalf("source destination was not rejected: path=%s error=%v", path, err)
	}
	got, err := os.ReadFile(input)
	if err != nil || string(got) != string(inputBytes) {
		t.Fatalf("source input changed: %q, %v", got, err)
	}
}

func pathTestCompiler(t *testing.T, installation, source, output string) (*compiler, buildOptions) {
	t.Helper()
	command, err := parseCommand([]string{"build", "--assets", installation, "--output-dir", output, source})
	if err != nil {
		t.Fatal(err)
	}
	input, initialization, options, err := resolveBuildCommand(command)
	if err != nil {
		t.Fatal(err)
	}
	c := newCompiler(input)
	c.initialize(initialization)
	if c.err != nil {
		t.Fatal(c.err)
	}
	return c, options
}

func TestBuildRejectsNonEmittingSourceDestinations(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	installation := distributiontest.CompleteOverride(t)
	for _, tc := range []struct {
		name, content         string
		alias, companionInput bool
	}{
		{name: "empty"},
		{name: "comments only", content: "*** Library reserved for later. ***\n"},
		{name: "imports only", content: "Use `leaf.l`.\n"},
		{name: "empty through directory alias", alias: true},
		{name: "empty inside companion", companionInput: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			project := filepath.Join(t.TempDir(), "source project")
			copyBuildProject(t, root, project, "library.l.l", "")
			destination := filepath.Join(project, "library.l")
			companion := destination + ".apps"
			generation := filepath.Join(companion, "generations", "previous")
			if err := os.MkdirAll(generation, 0700); err != nil {
				t.Fatal(err)
			}
			input := destination
			if tc.companionInput {
				input = filepath.Join(generation, "library.l")
			}
			relative, err := filepath.Rel(project, input)
			if err != nil {
				t.Fatal(err)
			}
			main := filepath.Join(project, "library.l.l")
			content, err := os.ReadFile(main)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(main, append([]byte("Use `./"+filepath.ToSlash(relative)+"`.\n"), content...), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(project, "leaf.l"), []byte("\"extra\" is 123.\n"), 0600); err != nil {
				t.Fatal(err)
			}
			// Existing ownership must not allow publication over a consumed source.
			preserved := map[string][]byte{
				input:                                []byte(tc.content),
				filepath.Join(companion, ".l-owner"): []byte(companionOwner + "library.l\n"),
				filepath.Join(generation, "app-000"): []byte("previous application bytes"),
			}
			for path, content := range preserved {
				if err := os.WriteFile(path, content, 0600); err != nil {
					t.Fatal(err)
				}
			}
			output := project
			if tc.alias {
				output = filepath.Join(t.TempDir(), "project-alias")
				if err := os.Symlink(project, output); err != nil {
					t.Fatal(err)
				}
			}
			compiler, options := pathTestCompiler(t, installation, main, output)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			path, err := buildExecutable(ctx, compiler, options)
			if err == nil || !strings.Contains(err.Error(), "source input") {
				t.Errorf("non-emitting source was not protected: output=%s error=%v", path, err)
			}
			if tc.companionInput {
				if _, err := os.Lstat(destination); !os.IsNotExist(err) {
					t.Errorf("rejected build published an executable: %v", err)
				}
			}
			for path, want := range preserved {
				got, err := os.ReadFile(path)
				if err != nil || !bytes.Equal(got, want) {
					t.Errorf("input/prior output changed at %s: length=%d want=%d error=%v", path, len(got), len(want), err)
				}
			}
			entries, err := os.ReadDir(filepath.Dir(generation))
			if err != nil || len(entries) != 1 || entries[0].Name() != "previous" {
				t.Errorf("rejected build published a generation: %v, %v", entries, err)
			}
			stages, err := filepath.Glob(filepath.Join(project, ".library.l-build-*"))
			if err != nil || len(stages) != 0 {
				t.Errorf("build left staging: %v, %v", stages, err)
			}
		})
	}
}
