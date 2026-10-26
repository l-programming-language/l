//go:build nativecompile

package Logiceme

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const lucidToolchain = "1.98.1"

// Separate from ordinary Go acceptance. Both phases compile only; no generated
// application or deployment script is executed, and actual output is not fixed.
func TestLucidNativeCompile(t *testing.T) {
	for _, phase := range []string{"Expected", "Generated"} {
		t.Run(phase, func(t *testing.T) {
			for _, c := range lucidCases {
				name := c.name
				t.Run(name, func(t *testing.T) {
					stage := "compile-" + strings.ToLower(phase)
					r, dir := lucidRecord(t, name, stage)
					if phase == "Expected" {
						lucidWrite(t, filepath.Join(dir, "input.aml.txt"), lucidRead(t, name+".aml.txt"))
						if !lucidSaveBundle(t, r, dir, lucidExpectedBundle(t, name)) {
							r.Status = "blocked"
							return
						}
					} else if _, ok := lucidGenerate(t, name, r, dir); !ok {
						r.Status = "blocked"
						cause := "bundle materialization failure"
						if r.GenerationError != "" {
							cause = "generation failure"
						}
						r.Issues = append(r.Issues, "native compilation blocked by "+cause)
						return
					}
					project := filepath.Join(dir, "actual", "app-000")
					if _, err := os.Stat(filepath.Join(project, "Cargo.toml")); err != nil {
						r.Status = "blocked"
						lucidIssue(t, r, "native compilation blocked: generated manifest missing: %v", err)
						return
					}
					lock := lucidCaseFor(name).lock
					lucidWrite(t, filepath.Join(project, "Cargo.lock"), lucidRead(t, lock))
					lucidCompileProject(t, r, project)
				})
			}
		})
	}
}

func lucidCompileProject(t *testing.T, r *lucidResult, project string) {
	t.Helper()
	cache, err := filepath.Abs(filepath.Join("..", ".harness", "cache", "lucid-rust-"+lucidToolchain))
	if err != nil {
		t.Fatal(err)
	}
	args := []string{"+" + lucidToolchain, "build", "--frozen", "--message-format=json", "--color=never", "--manifest-path", filepath.Join(project, "Cargo.toml"), "--target-dir", cache}
	r.Command = append([]string{"cargo"}, args...)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "cargo", args...)
	cmd.Dir = project
	// Isolate compiler flags/overrides without changing the user's global toolchain.
	for _, e := range os.Environ() {
		key := strings.SplitN(e, "=", 2)[0]
		switch key {
		case "RUSTFLAGS", "CARGO_ENCODED_RUSTFLAGS", "RUSTC", "RUSTC_WRAPPER", "RUSTC_WORKSPACE_WRAPPER", "CARGO_BUILD_TARGET", "CARGO_TARGET_DIR", "CARGO_NET_OFFLINE", "RUSTUP_AUTO_INSTALL":
			continue
		}
		cmd.Env = append(cmd.Env, e)
	}
	cmd.Env = append(cmd.Env, "CARGO_NET_OFFLINE=true", "RUSTUP_AUTO_INSTALL=0")
	lucidBoundCompiler(cmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	start := time.Now()
	runErr := cmd.Run()
	r.Duration = time.Since(start).String()
	code := -1
	if cmd.ProcessState != nil {
		code = cmd.ProcessState.ExitCode()
	}
	r.ExitCode = &code
	lucidWrite(t, filepath.Join(project, "cargo.jsonl"), stdout.Bytes())
	lucidWrite(t, filepath.Join(project, "cargo.stderr.txt"), stderr.Bytes())
	if ctx.Err() != nil {
		lucidIssue(t, r, "infrastructure: Cargo exceeded three-minute deadline")
		return
	}
	wantCode := lucidCaseFor(r.Scenario).diagnostic
	if runErr == nil {
		if wantCode != "" {
			lucidIssue(t, r, "Rust compilation: expected %s rejection, but project built", wantCode)
		}
		return
	}
	classification := "infrastructure"
	var diagnostics []string
	matched := false
	unexpected := false
	for _, line := range bytes.Split(stdout.Bytes(), []byte{'\n'}) {
		var message struct {
			Reason  string `json:"reason"`
			Message struct {
				Level   string                `json:"level"`
				Message string                `json:"message"`
				Spans   []lucidDiagnosticSpan `json:"spans"`
				Code    *struct {
					Code string `json:"code"`
				} `json:"code"`
			} `json:"message"`
		}
		if json.Unmarshal(line, &message) != nil || message.Reason != "compiler-message" || message.Message.Level != "error" {
			continue
		}
		classification = "Rust compilation"
		detail := message.Message.Message
		if message.Message.Code != nil {
			detail = message.Message.Code.Code + ": " + detail
		}
		diagnostics = append(diagnostics, detail)
		if wantCode != "" {
			if message.Message.Code == nil || message.Message.Code.Code != wantCode {
				unexpected = true
			} else {
				for _, span := range message.Message.Spans {
					if span.Primary && span.fromMainSource() {
						matched = true
					}
				}
			}
		}
	}
	if wantCode != "" && matched && !unexpected && code == 101 {
		return
	}
	if wantCode != "" {
		diagnostics = append(diagnostics, fmt.Sprintf("expected %s with primary main.rs source span: matched=%t unexpected=%t", wantCode, matched, unexpected))
	}
	if r.Stage == "compile-expected" && classification == "Rust compilation" {
		classification = "fixture/harness"
	}
	lucidIssue(t, r, "%s: cargo exit %d: %v; diagnostics=%s; stderr=%s", classification, code, runErr, strings.Join(diagnostics, "; "), fmt.Sprintf("%q", stderr.String()))
}

// Macro diagnostics may point at a dependency source but retain the invocation
// in their expansion chain. Only a concrete span in this generated main source
// establishes that the intended application was rejected.
type lucidDiagnosticSpan struct {
	File      string `json:"file_name"`
	Start     int    `json:"line_start"`
	End       int    `json:"line_end"`
	Primary   bool   `json:"is_primary"`
	Expansion *struct {
		Span *lucidDiagnosticSpan `json:"span"`
	} `json:"expansion"`
}

func (s lucidDiagnosticSpan) fromMainSource() bool {
	file := filepath.ToSlash(s.File)
	if (file == "src/main.rs" || strings.HasSuffix(file, "/src/main.rs")) && s.Start > 0 && s.End >= s.Start {
		return true
	}
	return s.Expansion != nil && s.Expansion.Span != nil && s.Expansion.Span.fromMainSource()
}
