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
	aliases := map[string]string{"Await": "app_wait", "Build": "app_build", "Run": "app_run", "Stop": "app_stop", "Output": "app_output", "Error": "app_error", "Show": "runtime_show", "Fail": "runtime_fail", "ClaimContinuation": "invocation_claim", "FinishContinuation": "invocation_finish", "Finish": "runtime_finish", "Shutdown": "runtime_shutdown", "SetSource": "app_set_source"}
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
