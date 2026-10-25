package main

import (
	"bytes"
	"errors"

	"os"
	"os/exec"
	"path/filepath"

	"testing"
)

func paragraphDemoSources(t *testing.T, root string) map[string][]byte {
	t.Helper()
	sources := make(map[string][]byte)
	for _, name := range []string{"main.l", "local-words.l"} {
		content, err := os.ReadFile(filepath.Join(root, "demo/paragraphs", name))
		if err != nil {
			t.Fatal(err)
		}
		sources[name] = content
	}
	return sources
}

func writeParagraphProject(t *testing.T, sources map[string][]byte, main []byte) string {
	t.Helper()
	directory, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for name, content := range sources {
		if name == "main.l" {
			content = main
		}
		if err := os.WriteFile(filepath.Join(directory, name), content, 0600); err != nil {
			t.Fatal(err)
		}
	}
	return directory
}

// The exact paragraph demo remains source input. Its newline-sensitive Error
// branches are generated here but are not executed by this source-only check.

type paragraphExample struct {
	filename, stdout, stderr string
	status                   int
	apps, printed            []string
}

// These are executable language examples, not rewrites of the author's main.l.
// Every positive case must compile and execute; current compiler failures are RED.
var paragraphExamples = []paragraphExample{
	{filename: "fallthrough.l", stdout: "First\nSecond\ndone", apps: []string{"First", "Second"}, printed: []string{"First", "Second"}},
	{filename: "exit-program.l", stdout: "First\n", apps: []string{"First", "Second"}, printed: []string{"First", "Second"}},
	{filename: "exit-error.l", stdout: "First\n", stderr: "paragraph stopped\n", status: 1, apps: []string{"First", "Second"}, printed: []string{"First", "Second"}},
	{filename: "multiple-subjects.l", stdout: "Second\nsecond-end|First\n", apps: []string{"First", "Second"}, printed: []string{"First", "Second"}},
	{filename: "repeated-activation.l", stdout: "Repeat\nRepeat\nRepeat\nRepeat\nRepeat\nRepeat\n", apps: []string{"Repeat"}, printed: []string{"Repeat"}},
	{filename: "run-within-paragraph.l", stdout: "Other\nInner\nfinisheddone", apps: []string{"Worker", "Other"}, printed: []string{"Inner", "Other"}},
	{filename: "coordinate-existing.l", stdout: "Second\nFirst\ndone", apps: []string{"First", "Second"}, printed: []string{"First", "Second"}},
	{filename: "public-values.l", stdout: "public7true"},
	{filename: "local-shadowing.l", stdout: "inner7truesecondouter"},
	{filename: "local-definitions.l", stdout: "First local\nSecond local\ndone", apps: []string{"First", "Second"}, printed: []string{"First local", "Second local"}},
	{filename: "local-imports.l", stdout: "Imported\ndone", apps: []string{"Worker"}, printed: []string{"Imported"}},
}

func paragraphExampleSource(t *testing.T, root, filename string) []byte {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(root, "demo", "paragraphs", filename))
	if err != nil {
		t.Fatal(err)
	}
	return content
}

func paragraphExampleProject(t *testing.T, sources map[string][]byte, filename string, main []byte) string {
	t.Helper()
	project := writeParagraphProject(t, sources, main)
	if filename != "main.l" {
		if err := os.Rename(filepath.Join(project, "main.l"), filepath.Join(project, filename)); err != nil {
			t.Fatal(err)
		}
	}
	return project
}

func paragraphProjectUnchanged(t *testing.T, project, filename string, main []byte, sources map[string][]byte) {
	t.Helper()
	for name, original := range sources {
		if name == "main.l" {
			name, original = filename, main
		}
		current, err := os.ReadFile(filepath.Join(project, name))
		if err != nil || !bytes.Equal(current, original) {
			t.Errorf("paragraph compiler/runtime changed source %s: %v", name, err)
		}
	}
}

// Execute actual checked-in demos with production IO. Marker/counter cases now
// live in acceptance/distribution/observers and use explicit test-owned sources.

func paragraphExecution(t *testing.T, stdout, stderr []byte, err error, status int, wantOut, wantError string) {
	t.Helper()
	code := 0
	if err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) {
			t.Fatalf("paragraph executable did not complete: %v", err)
		}
		code = exit.ExitCode()
	}
	if code != status || string(stdout) != wantOut || string(stderr) != wantError {
		t.Fatalf("paragraph execution status=%d stdout=%q stderr=%q; want %d %q %q", code, stdout, stderr, status, wantOut, wantError)
	}
}
