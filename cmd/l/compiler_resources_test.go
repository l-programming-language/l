//go:build darwin || linux

package main

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	distribution "github.com/l-programming-language/l"
	distributiontest "github.com/l-programming-language/l/internal/testsupport/distribution"
)

func TestCompilerResourceSelection(t *testing.T) {
	main := filepath.Join(t.TempDir(), "main.l")
	if err := os.WriteFile(main, []byte("Show `selected`.\n"), 0600); err != nil {
		t.Fatal(err)
	}
	initialize := func(t *testing.T, options compilerInitOptions) *compiler {
		t.Helper()
		c := newCompiler(&cliData{jobType: "build", mainPath: main, osType: "ubuntu"})
		c.initialize(options)
		return c
	}
	t.Run("default resources are embedded", func(t *testing.T) {
		c := initialize(t, compilerInitOptions{})
		if c.err != nil {
			t.Fatal(c.err)
		}
		compiled := c.compile()
		if c.err != nil || compiled == nil {
			t.Fatal(c.err)
		}
		assertFrontendResources(t, compiled.Compilation)
		if c.options.Language != "eng" || c.options.AssetRoot != "" {
			t.Fatalf("default resource selection = %#v", c.options)
		}
		want, err := fs.ReadFile(distribution.Files(), "builtins/eng/core.l")
		if err != nil {
			t.Fatal(err)
		}
		got, err := fs.ReadFile(c.options.Resources, "builtins/eng/core.l")
		if err != nil || !bytes.Equal(got, want) {
			t.Fatalf("default compiler does not use its embedded distribution: %v", err)
		}
	})
	t.Run("explicit language retains embedded resources", func(t *testing.T) {
		c := initialize(t, compilerInitOptions{Language: "eng"})
		if c.compile() == nil || c.err != nil {
			t.Fatal(c.err)
		}
		if c.options.Language != "eng" || c.options.AssetRoot != "" {
			t.Fatalf("explicit language selection = %#v", c.options)
		}
		if _, err := fs.Stat(c.options.Resources, "stdlib/index.json"); err != nil {
			t.Fatalf("explicit language lost embedded resources: %v", err)
		}
	})
	t.Run("CLI selects complete alternate language", func(t *testing.T) {
		assets := distributiontest.CompleteOverride(t)
		for _, directory := range []string{"lexicon", "builtins"} {
			if err := os.Rename(filepath.Join(assets, directory, "eng"), filepath.Join(assets, directory, "selected")); err != nil {
				t.Fatal(err)
			}
		}
		indexPath := filepath.Join(assets, "stdlib/index.json")
		data, err := os.ReadFile(indexPath)
		if err != nil {
			t.Fatal(err)
		}
		var index distributionIndex
		if err := json.Unmarshal(data, &index); err != nil {
			t.Fatal(err)
		}
		index.Languages["selected"] = index.Languages["eng"]
		delete(index.Languages, "eng")
		data, err = json.Marshal(index)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(indexPath, data, 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(main, []byte("Use IO's Print from `io` as \"Print\".\nShow `selected`.\n"), 0600); err != nil {
			t.Fatal(err)
		}
		command, err := parseCommand([]string{"build", "--language=selected", "--assets", assets, main})
		if err != nil {
			t.Fatal(err)
		}
		input, selected, _, err := resolveBuildCommand(command)
		if err != nil {
			t.Fatal(err)
		}
		c := newCompiler(input)
		c.initialize(selected)
		if c.process() == nil || c.err != nil {
			t.Fatalf("alternate language generation: %v", c.err)
		}
		if c.options.Language != "selected" || c.options.AssetRoot != assets {
			t.Fatalf("selected compiler options = %#v", c.options)
		}
	})
}
