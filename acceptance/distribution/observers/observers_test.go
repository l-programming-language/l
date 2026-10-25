//go:build darwin || linux

package observers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	distributiontest "github.com/l-programming-language/l/internal/testsupport/distribution"
	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	"github.com/l-programming-language/l/internal/testsupport/phase1"
)

func TestMain(m *testing.M) { phase1.Main(m) }

type observation struct {
	Source   string   `json:"source"`
	Stdout   string   `json:"stdout"`
	Stderr   string   `json:"stderr"`
	Status   int      `json:"status"`
	Apps     []string `json:"apps"`
	Runs     string   `json:"runs"`
	Numbered string   `json:"numbered,omitempty"`
}

func checkRuns(t testing.TB, path, want string) {
	t.Helper()
	b, err := os.ReadFile(path)
	if want == "" {
		if !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("unexpected child invocation marker=%q error=%v", b, err)
		}
		return
	}
	if err != nil || string(b) != want {
		t.Fatalf("child invocations=%q/%v; want %q", b, err, want)
	}
}
func execute(t testing.TB, binary, marker string, env []string, tc observation, stdout string) {
	t.Helper()
	out, diag, err := nativetest.RuntimeExecution.Run(t, binary, t.TempDir(), env)
	status := 0
	if err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) {
			t.Fatal(err)
		}
		status = exit.ExitCode()
	}
	if status != tc.Status || string(out) != stdout || string(diag) != tc.Stderr {
		t.Fatalf("execution=%d %q %q; want %d %q %q", status, out, diag, tc.Status, stdout, tc.Stderr)
	}
	checkRuns(t, marker, tc.Runs)
}
func assertPublished(t testing.TB, binary string, ids []string) {
	t.Helper()
	nativetest.AssertNativeArtifact(t, binary)
	var manifest struct {
		GOOS   string `json:"goos"`
		GOARCH string `json:"goarch"`
		Apps   []struct {
			ID         string `json:"id"`
			Executable string `json:"executable"`
			SHA256     string `json:"sha256"`
		} `json:"apps"`
	}
	manifestPath := nativetest.ReadPublishedGeneration(t, binary, &manifest)
	if manifest.GOOS != runtime.GOOS || manifest.GOARCH != runtime.GOARCH || len(manifest.Apps) != len(ids) {
		t.Fatalf("manifest=%+v want assertNative apps=%v", manifest, ids)
	}
	for i, app := range manifest.Apps {
		if app.ID != ids[i] {
			t.Fatalf("app=%q want %q", app.ID, ids[i])
		}
		path := filepath.Join(filepath.Dir(manifestPath), app.Executable)
		if app.SHA256 != nativetest.NativeArtifactHash(t, path) {
			t.Fatalf("App %s hash differs from manifest", app.ID)
		}
	}
}

// Cases have independent projects, resource copies, provenance recorders and
// child process groups. Bound admission before starting their existing watchdogs.
var nativeCases = make(nativetest.Admission, 4)

// These explicit fixture profiles retain the original marker, counter, and
// exact-output assertions. Production IO and demo sources are never rewritten.
func TestParagraphObservers(t *testing.T) {
	t.Parallel()
	root := filepath.Join(distributiontest.RepoRoot(t), "testdata/distribution/observers")
	data, err := os.ReadFile(filepath.Join(root, "cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cases []observation
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cases); err != nil {
		t.Fatal(err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		t.Fatalf("trailing manifest data: %v", err)
	}
	if len(cases) != 24 {
		t.Fatalf("observer coverage=%d; want 24 retained cases", len(cases))
	}
	seen := map[string]bool{}
	for _, tc := range cases {
		tc := tc
		if !filepath.IsLocal(tc.Source) || filepath.Base(tc.Source) != tc.Source || seen[tc.Source] {
			t.Fatalf("invalid/duplicate observer source %q", tc.Source)
		}
		seen[tc.Source] = true
		t.Run(tc.Source, func(t *testing.T) {
			nativeCases.Parallel(t)
			project, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			sources := map[string][]byte{}
			for _, name := range []string{tc.Source, "paragraph.l", "local-words.l"} {
				data, err := os.ReadFile(filepath.Join(root, name))
				if err != nil {
					t.Fatal(err)
				}
				sources[name] = data
				if err := os.WriteFile(filepath.Join(project, name), data, 0600); err != nil {
					t.Fatal(err)
				}
			}
			defer func() {
				for name, want := range sources {
					got, err := os.ReadFile(filepath.Join(project, name))
					if err != nil || !bytes.Equal(got, want) {
						t.Errorf("observer source %s changed: %v", name, err)
					}
				}
			}()
			marker := filepath.Join(project, "child-runs.log")
			env := []string{"L_PARAGRAPH_RUN_MARKERS=" + marker, "L_PARAGRAPH_NUMBER_OUTPUT=0"}
			recorder := nativetest.CLIProbe(t)
			out, diag, err := nativetest.BuildPrerequisite.Run(t, phase1.CLI(t), project, recorder.Env(env), "build", tc.Source)
			if err != nil {
				t.Fatalf("builtin/explicit-observer compilation prerequisite; runtime not exercised: %v\nstdout=%q\nstderr=%s", err, out, diag)
			}
			binary := filepath.Join(project, strings.TrimSuffix(tc.Source, ".l"))
			if string(out) != binary+"\n" || len(diag) != 0 {
				t.Fatalf("build streams=%q/%q", out, diag)
			}
			recorder.AssertPublished(t, binary)
			checkRuns(t, marker, "")
			assertPublished(t, binary, tc.Apps)
			if tc.Source == "unread-unavailable-output.l" {
				tc.Stderr = fmt.Sprintf("%s:6:5: assign %s", filepath.Join(project, tc.Source), tc.Stderr)
			}
			execute(t, binary, marker, env, tc, tc.Stdout)
			if tc.Numbered != "" {
				if err := os.Remove(marker); err != nil {
					t.Fatal(err)
				}
				env = []string{"L_PARAGRAPH_RUN_MARKERS=" + marker, "L_PARAGRAPH_NUMBER_OUTPUT=1"}
				execute(t, binary, marker, env, tc, tc.Numbered)
			}
		})
	}
}
