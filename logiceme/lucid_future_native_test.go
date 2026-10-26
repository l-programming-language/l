//go:build nativecompile && (darwin || linux)

package Logiceme

import (
	"path/filepath"
	"testing"

	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
)

// Execute the actual generated bundles. Their assertions observe laziness,
// evaluation order, capture ownership, and the values obtained from awaiting.
func TestLucidFutureExecution(t *testing.T) {
	for _, name := range []string{"future-composed", "future-move", "future-infer"} {
		t.Run(name, func(t *testing.T) {
			r, dir := lucidRecord(t, name, "execute-generated")
			if _, ok := lucidGenerate(t, name, r, dir); !ok {
				return
			}
			project := filepath.Join(dir, "actual", "app-000")
			lucidWrite(t, filepath.Join(project, "Cargo.lock"), lucidRead(t, lucidCaseFor(name).lock))
			lucidCompileProject(t, r, project)
			if t.Failed() {
				return
			}
			binary, err := filepath.Abs(filepath.Join("..", ".harness", "cache", "lucid-rust-"+lucidToolchain, "debug", "lucid_nested_rust"))
			if err != nil {
				t.Fatal(err)
			}
			stdout, stderr, err := nativetest.RuntimeExecution.Run(t, binary, project, nil)
			if err != nil || len(stdout) != 0 || len(stderr) != 0 {
				lucidIssue(t, r, "execution stdout=%q stderr=%q error=%v", stdout, stderr, err)
			}
		})
	}
}
