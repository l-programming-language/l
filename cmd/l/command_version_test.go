package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	distribution "github.com/l-programming-language/l"
)

func TestVersionCommand(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	for _, flag := range []string{"--version", "-v"} {
		t.Run(flag, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			status := run(context.Background(), []string{flag}, &stdout, &stderr)
			if want := distribution.Version() + "\n"; status != 0 || stdout.String() != want || stderr.Len() != 0 {
				t.Fatalf("version: status=%d stdout=%q stderr=%q; want 0, %q, empty stderr", status, stdout.String(), stderr.String(), want)
			}
		})
	}
}

func TestVersionCommandInvalidArguments(t *testing.T) {
	for _, args := range [][]string{
		{"--version", "extra"}, {"--version", "build"}, {"--version", "--version"},
		{"--version", "--help"}, {"--help", "--version"}, {"build", "--version"},
		{"-v", "extra"}, {"-v", "build"}, {"-v", "-v"},
		{"-v", "--help"}, {"--help", "-v"}, {"build", "-v"},
		{"--version", "-v"}, {"-v", "--version"},
		{"--version=true"}, {"-v=true"}, {"-V"}, {"-version"}, {"version"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			status := run(context.Background(), args, &stdout, &stderr)
			if status != 2 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "error:") {
				t.Fatalf("invalid version arguments: status=%d stdout=%q stderr=%q", status, stdout.String(), stderr.String())
			}
		})
	}
}
