//go:build darwin || linux

package main

import (
	"errors"
	"path/filepath"
	"runtime"
	"testing"
)

func TestBuildCommandOptionsAndNativeResolution(t *testing.T) {
	assets := t.TempDir()
	output := filepath.Join(t.TempDir(), "binaries")
	command, err := parseCommand([]string{"build", "--language", "selected", "--assets", assets, "--output-dir", output, "project/story.l"})
	if err != nil {
		t.Fatal(err)
	}
	if command.Assets != assets || command.Output != output || command.Language != "selected" {
		t.Fatalf("parsed options lost values: %#v", command)
	}
	t.Setenv("GOOS", "windows")
	t.Setenv("GOARCH", "386")
	input, initialization, options, err := resolveBuildCommand(command)
	if err != nil {
		t.Fatal(err)
	}
	mainPath, err := filepath.Abs("project/story.l")
	if err != nil {
		t.Fatal(err)
	}
	if input.mainPath != mainPath || initialization.Language != "selected" || initialization.AssetRoot != assets {
		t.Fatalf("resolved input/assets = %#v, %#v", input, initialization)
	}
	if options.OutputDir != output || options.GOOS != runtime.GOOS || options.GOARCH != runtime.GOARCH {
		t.Fatalf("native resolution used ambient target or lost output path: %#v", options)
	}
	if command.CLI.mainPath != "project/story.l" {
		t.Fatal("path resolution mutated the parsed command data")
	}
	for _, host := range []string{"windows", "freebsd", "plan9"} {
		if _, err := nativeBuildOptions(output, host, "amd64"); !errors.Is(err, errUnsupportedBuildHost) {
			t.Errorf("unsupported native host %s lost its validation classification: %v", host, err)
		}
	}
}
