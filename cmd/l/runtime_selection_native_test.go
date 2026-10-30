//go:build darwin || linux

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestSelectedRuntimeLibraryNative(t *testing.T) {
	for _, tc := range []struct {
		name, source, output string
		apps                 []string
	}{
		{"value", "Show `selected`.\n", "selected", nil},
		{"App", "Use IO's Print from `io` as \"Print\".\nCreate \"Worker\" as a New Basic App.\nThe App Prints `selected`.\nExecute Worker.\nShow Worker's Output.\n", "selected\n", []string{"Worker"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			project, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(project, "main.l"), []byte(tc.source), 0600); err != nil {
				t.Fatal(err)
			}
			built := runnableBuildProject(t, project, tc.apps)
			status, stdout, stderr := runnableRun(t, built.binary, t.TempDir(), nil)
			if status != 0 || string(stdout) != tc.output || len(stderr) != 0 {
				t.Fatalf("selected runtime: status=%d stdout=%q stderr=%q", status, stdout, stderr)
			}
			paths, err := filepath.Glob(filepath.Join(built.recorder.Directory, "cargo.*.workspace", "support/runtime/src/runtime.rs"))
			if err != nil || len(paths) != 1 {
				t.Fatalf("runtime source evidence: %v, %v", paths, err)
			}
			code, err := os.ReadFile(paths[0])
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Contains(code, []byte("fn app_build(")) != (len(tc.apps) > 0) || !bytes.Contains(code, []byte("fn runtime_show(")) {
				t.Fatal("installed compiler did not compile the selected runtime closure")
			}
		})
	}
}
