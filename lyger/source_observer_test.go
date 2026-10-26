package Lyger

import (
	"testing"

	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	Logiceme "github.com/l-programming-language/l/logiceme"
)

type observedSourceFile struct{ *nativetest.RustSource }

func (f *observedSourceFile) function(t *testing.T, name string) nativetest.RustFunction {
	t.Helper()
	return f.Function(t, name)
}

func observeBundleSources(t *testing.T, b Logiceme.Executable) map[string]*nativetest.RustSource {
	t.Helper()
	return nativetest.ObserveRustBundle(t, b.ByteArrays, b.MapArrays)
}

func observeBundleMain(t *testing.T, b Logiceme.Executable) *observedSourceFile {
	t.Helper()
	return &observedSourceFile{nativetest.ObserveRustMain(t, b.ByteArrays, b.MapArrays)}
}

func observedMethodName(name string) string {
	aliases := map[string]string{"Await": "wait", "Build": "build", "Run": "run", "Stop": "stop", "Output": "output", "Error": "error", "Show": "show", "Fail": "fail", "ClaimContinuation": "claim_continuation", "FinishContinuation": "finish_continuation", "Finish": "finish", "Shutdown": "shutdown", "SetSource": "set_source", "Println": "println"}
	if value, ok := aliases[name]; ok {
		return value
	}
	return name
}
func phase1Calls(f *observedSourceFile, name string) int {
	count := 0
	for _, fn := range f.Functions() {
		for _, call := range nativetest.RustCalls(fn.Body) {
			if call.Name == observedMethodName(name) {
				count++
			}
		}
	}
	return count
}
