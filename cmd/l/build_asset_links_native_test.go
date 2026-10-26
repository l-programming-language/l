//go:build darwin || linux

package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	distributiontest "github.com/l-programming-language/l/internal/testsupport/distribution"
	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
)

func TestBuildAcceptsLinkedResourceBundle(t *testing.T) {
	assets := distributiontest.CompleteOverride(t)
	shared := filepath.Join(t.TempDir(), "shared library")
	if err := os.Rename(filepath.Join(assets, "stdlib"), shared); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(shared, filepath.Join(assets, "stdlib")); err != nil {
		t.Fatal(err)
	}
	// Both spellings eventually point back to previously visited directories.
	if err := os.Symlink(assets, filepath.Join(shared, "bundle")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(shared, filepath.Join(shared, "self")); err != nil {
		t.Fatal(err)
	}
	project := t.TempDir()
	main := filepath.Join(project, "main.l")
	source := "Use IO's Print from `io` as \"Print\".\nCreate \"Worker\" as a New Basic App.\nThe App Prints `linked resources`.\nExecute Worker.\nShow Worker's Output.\n"
	if err := os.WriteFile(main, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	c, options := pathTestCompiler(t, assets, main, project)
	recorder := nativetest.CLIProbe(t)
	for _, v := range recorder.Env(nil) {
		k, val, ok := strings.Cut(v, "=")
		if ok {
			t.Setenv(k, val)
		}
	}
	binary, err := buildExecutable(context.Background(), c, options)
	if err != nil {
		t.Fatal(err)
	}
	recorder.AssertPublished(t, binary)
	stdout, stderr, err := runRuntimeCommand(t, binary, project, nil)
	if err != nil || string(stdout) != "linked resources\n" || len(stderr) != 0 {
		t.Fatalf("linked complete bundle execution = %q / %q, %v", stdout, stderr, err)
	}
}
