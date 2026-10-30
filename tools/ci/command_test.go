package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestCIParallelismFitsHost(t *testing.T) {
	for _, tc := range []struct{ cpus, want int }{{1, 1}, {2, 2}, {3, 3}, {4, 4}, {8, 4}} {
		t.Run(strconv.Itoa(tc.cpus), func(t *testing.T) {
			if got := ciParallelism(tc.cpus); got != tc.want {
				t.Fatalf("%d CPUs: parallelism=%d, want %d", tc.cpus, got, tc.want)
			}
		})
	}
}

func TestCommandCapsInheritedParallelism(t *testing.T) {
	t.Setenv("GOMAXPROCS", "8")
	r := runner{root: t.TempDir(), directory: t.TempDir()}
	output, err := r.command("parallelism", time.Second, nil, "/bin/sh", "-c", `printf '%s' "$GOMAXPROCS"`)
	want := strconv.Itoa(ciParallelism(runtime.NumCPU()))
	if err != nil || string(output) != want {
		t.Fatalf("child GOMAXPROCS=%q, want %s: %v", output, want, err)
	}
}

func TestCommandPreservesFailureAndBothStreams(t *testing.T) {
	r := runner{root: t.TempDir(), directory: t.TempDir()}
	output, err := r.command("failure", time.Second, nil, "/bin/sh", "-c", "printf stdout; printf stderr >&2; exit 7")
	if err == nil || string(output) != "stdout" || len(r.report.Commands) != 1 || r.report.Commands[0].ExitCode != 7 || r.report.Commands[0].TimeoutMS != 1000 {
		t.Fatalf("lost child status/output: %q %v %+v", output, err, r.report.Commands)
	}
	stderr, err := os.ReadFile(r.report.Commands[0].StderrPath)
	if err != nil || string(stderr) != "stderr" {
		t.Fatal("lost stderr", err)
	}
}

func TestGoCommandDoesNotUseShadowedPath(t *testing.T) {
	dir := t.TempDir()
	shadow := filepath.Join(dir, "go")
	testWrite(t, shadow, []byte("#!/bin/sh\nprintf 'wrong go'\n"))
	if err := os.Chmod(shadow, 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	r := runner{root: t.TempDir(), directory: t.TempDir()}
	output, err := r.command("go-version", time.Second, nil, "go", "version")
	if err != nil || !strings.Contains(string(output), runtime.Version()) {
		t.Fatalf("unpinned Go: %s %v", output, err)
	}
	if r.report.Commands[0].Executable != filepath.Join(runtime.GOROOT(), "bin", "go") {
		t.Fatal("Go was not resolved before process launch")
	}
}

func TestCommandRecordsTimeoutWithoutSuccess(t *testing.T) {
	r := runner{root: t.TempDir(), directory: t.TempDir()}
	_, err := r.command("timeout", 30*time.Millisecond, nil, "/bin/sh", "-c", "sleep 5")
	if err == nil || !r.report.Commands[0].TimedOut || r.report.Commands[0].ExitCode == 0 || r.report.Commands[0].TimeoutMS != 30 {
		t.Fatal("timed-out command accepted")
	}
}
