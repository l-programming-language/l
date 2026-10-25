package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"io"
	"os"
	"path/filepath"
	"reflect"

	"strings"
	"testing"
)

// Exercise the production parser's established CLI data projection.
// These tests remain serial because they also observe process-global flags.
func TestBuildCommandParsing(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want *cliData
		err  bool
		help bool
	}{
		{"default source", []string{"build"}, &cliData{"build", "main.l", "ubuntu"}, false, false},
		{"explicit source", []string{"build", "project/main.l"}, &cliData{"build", "project/main.l", "ubuntu"}, false, false},
		{"options before source", []string{"build", "--assets", "resources", "--output-dir=out", "project/main.l"}, &cliData{"build", "project/main.l", "ubuntu"}, false, false},
		{"assets before command", []string{"--assets", "resources", "build", "project/main.l"}, nil, true, false},
		{"options after source", []string{"build", "project/main.l", "--assets=resources"}, nil, true, false},
		{"source after separator", []string{"build", "--", "-source.l"}, &cliData{"build", "-source.l", "ubuntu"}, false, false},
		{"version spelling after separator", []string{"build", "--", "--version"}, &cliData{"build", "--version", "ubuntu"}, false, false},
		{"missing command", nil, nil, true, false},
		{"unknown command", []string{"run", "main.l"}, nil, true, false},
		{"extra source", []string{"build", "main.l", "other.l"}, nil, true, false},
		{"OS before command", []string{"-os", "ubuntu", "build"}, nil, true, false},
		{"OS after command", []string{"build", "--os=centos"}, nil, true, false},
		{"OS after source", []string{"build", "main.l", "-os", "ubuntu"}, nil, true, false},
		{"unknown flag", []string{"build", "main.l", "--unknown"}, nil, true, false},
		{"missing assets argument", []string{"build", "--assets"}, nil, true, false},
		{"missing language argument", []string{"build", "--language"}, nil, true, false},
		{"empty language argument", []string{"build", "--language="}, nil, true, false},
		{"empty assets argument", []string{"build", "--assets="}, nil, true, false},
		{"unavailable config flag", []string{"build", "--config", "settings.json"}, nil, true, false},
		{"help", []string{"--help"}, nil, true, true},
		{"build help", []string{"build", "--help"}, nil, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			previousArgs, previousFlags := os.Args, flag.CommandLine
			t.Cleanup(func() { os.Args, flag.CommandLine = previousArgs, previousFlags })
			os.Args = append([]string{"l"}, tc.args...)
			flag.CommandLine = flag.NewFlagSet("l", flag.ContinueOnError)
			flag.CommandLine.SetOutput(io.Discard)
			command, err := parseCommand(os.Args[1:])
			var got *cliData
			if command != nil {
				got = command.CLI
			}
			if (err != nil) != tc.err || (tc.help && !errors.Is(err, flag.ErrHelp)) {
				t.Fatalf("parseCommand(%q) CLI = %#v, %v; want error=%v help=%v", tc.args, got, err, tc.err, tc.help)
			}
			if !tc.err && !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("parseCommand(%q) CLI = %#v; want %#v", tc.args, got, tc.want)
			}
		})
	}
}

func TestBuildCommandInstalledAssetsAndDefaultOutput(t *testing.T) {
	installation := t.TempDir()
	bin := filepath.Join(installation, "bin")
	if err := os.Mkdir(bin, 0700); err != nil {
		t.Fatal(err)
	}
	executable := filepath.Join(bin, "l")
	if err := os.WriteFile(executable, []byte("test executable location"), 0600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(t.TempDir(), "l")
	if err := os.Symlink(executable, link); err != nil {
		t.Fatal(err)
	}
	command, err := parseCommand([]string{"build", "project/main.l"})
	if err != nil {
		t.Fatal(err)
	}
	input, initialization, options, err := resolveBuildCommand(command)
	if err != nil {
		t.Fatal(err)
	}
	if initialization.AssetRoot != "" || initialization.Language != "eng" {
		t.Fatalf("FS02: default resources must be embedded, independent of symlink location: %#v", initialization)
	}
	if options.OutputDir != filepath.Dir(input.mainPath) || !filepath.IsAbs(options.OutputDir) {
		t.Fatalf("output is not beside source: %#v, %#v", input, options)
	}
	misplaced := filepath.Join(t.TempDir(), "l")
	if err := os.WriteFile(misplaced, nil, 0600); err != nil {
		t.Fatal(err)
	}
	if _, initialization, _, err := resolveBuildCommand(command); err != nil || initialization.AssetRoot != "" || initialization.Language != "eng" {
		t.Fatalf("FS01: relocated compiler must use embedded resources without --assets: %#v, %v", initialization, err)
	}
}

func TestBuildCommandRunDiagnostics(t *testing.T) {
	assets := t.TempDir()
	missingAssets := filepath.Join(assets, "missing")
	assetFile := filepath.Join(assets, "plain-file")
	if err := os.WriteFile(assetFile, []byte("not an asset directory"), 0600); err != nil {
		t.Fatal(err)
	}
	brokenAssets := filepath.Join(assets, "broken-link")
	if err := os.Symlink(missingAssets, brokenAssets); err != nil {
		t.Fatal(err)
	}
	mainPath := filepath.Join(t.TempDir(), "main.l")
	if err := os.WriteFile(mainPath, []byte("Show `resource check`."), 0600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name    string
		args    []string
		status  int
		message string
	}{
		{"global help", []string{"--help"}, 0, "Usage:"},
		{"command help", []string{"build", "--help"}, 0, "Usage:"},
		{"invalid invocation", []string{"run"}, 2, "unknown command"},
		{"missing asset directory", []string{"build", "--assets", missingAssets, mainPath}, 1, missingAssets},
		{"asset root is a file", []string{"build", "--assets", assetFile, mainPath}, 1, "not a directory"},
		{"broken asset symlink", []string{"build", "--assets", brokenAssets, mainPath}, 1, brokenAssets},
		{"missing vocabulary", []string{"build", "--assets", assets, mainPath}, 1, "lexicon/eng/vocabulary.json"},
		{"unavailable language", []string{"build", "--language=absent", mainPath}, 1, "lexicon/absent/vocabulary.json"},
		{"invalid language", []string{"build", "--language=../eng", mainPath}, 1, "valid language is required"},
		{"blank language", []string{"build", "--language= ", mainPath}, 1, "valid language is required"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			status := run(context.Background(), tc.args, &stdout, &stderr)
			if status != tc.status {
				t.Fatalf("status=%d; want %d; stdout=%q stderr=%q", status, tc.status, stdout.String(), stderr.String())
			}
			if tc.status == 0 {
				if stderr.Len() != 0 || !strings.Contains(stdout.String(), tc.message) {
					t.Fatalf("help streams: stdout=%q stderr=%q", stdout.String(), stderr.String())
				}
			} else if stdout.Len() != 0 || !strings.Contains(stderr.String(), tc.message) || strings.Contains(stderr.String(), "<nil>") {
				t.Fatalf("error streams lost provenance: stdout=%q stderr=%q", stdout.String(), stderr.String())
			}
		})
	}
}

func TestBuildCommandParsingDoesNotRegisterGlobalFlags(t *testing.T) {
	previousArgs, previousFlags := os.Args, flag.CommandLine
	t.Cleanup(func() { os.Args, flag.CommandLine = previousArgs, previousFlags })
	os.Args = []string{"l", "build", "main.l"}
	flag.CommandLine = flag.NewFlagSet("unrelated", flag.ContinueOnError)
	flag.CommandLine.SetOutput(io.Discard)
	before := flag.CommandLine
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Errorf("repeated CLI parsing panicked: %v", recovered)
		}
	}()
	for attempt := 0; attempt < 2; attempt++ {
		if _, err := parseCommand(os.Args[1:]); err != nil {
			t.Fatal(err)
		}
	}
	if flag.CommandLine != before || before.Lookup("os") != nil {
		t.Fatal("CLI parsing changed the process-global flag registry")
	}
}

func TestBuildCommandOutputFailure(t *testing.T) {
	closed, err := os.CreateTemp(t.TempDir(), "closed-output-")
	if err != nil {
		t.Fatal(err)
	}
	if err := closed.Close(); err != nil {
		t.Fatal(err)
	}
	if status := run(context.Background(), []string{"--help"}, closed, io.Discard); status != 1 {
		t.Fatalf("help with unavailable output returned status %d; want 1", status)
	}
	var stderr bytes.Buffer
	if status := run(context.Background(), []string{"--version"}, closed, &stderr); status != 1 || stderr.Len() != 0 {
		t.Fatalf("version with unavailable output: status=%d stderr=%q; want 1, empty stderr", status, stderr.String())
	}
	if status := run(context.Background(), []string{"invalid-command"}, io.Discard, closed); status != 2 {
		t.Fatalf("unavailable diagnostic changed invocation status to %d; want 2", status)
	}
}

func TestBuildCommandEmptyProcessArguments(t *testing.T) {
	if _, err := parseCommand(nil); err == nil {
		t.Fatal("an empty process argument list should return a command error")
	}
}
