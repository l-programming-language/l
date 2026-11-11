package Lyger

import (
	"strings"
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

func generatedRustSource(bundle Logiceme.Executable) string {
	var result strings.Builder
	for i, paths := range bundle.MapArrays {
		for j, path := range paths {
			if strings.HasSuffix(path, ".rs") {
				result.Write(bundle.ByteArrays[i][j])
			}
		}
	}
	return result.String()
}
