//go:build nativecompile

package Logiceme

import (
	"bytes"
	"strings"
	"testing"

	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
)

func TestNativeMetadataArrowsCompile(t *testing.T) {
	for _, phase := range []string{"Expected", "Generated"} {
		t.Run(phase, func(t *testing.T) {
			bundle := nativeMetadataArrowsExpected(t)
			if phase == "Generated" {
				bundle = nativeMetadataArrowsGenerate(t, NewLogicemeService(runtimeGenerationResources(t)), nativeMetadataArrowsAML)
				nativeMetadataArrowsCompare(t, bundle, nativeMetadataArrowsExpected(t))
			}
			nativetest.RecordBundle(t, phase, bundle)
			project := nativetest.Materialize(t, nativeMetadataArrowsFiles(t, bundle))
			result := nativetest.Build(t, project, nativetest.BuildOptions{})
			t.Logf("%s native compile command: %q", phase, result.Command)
			if result.Err != nil {
				t.Fatalf("%s native compile: %v\nstdout:\n%s\nstderr:\n%s", phase, result.Err, result.Stdout, result.Stderr)
			}
		})
	}
}

func nativeMetadataArrowsFiles(t *testing.T, bundle Executable) map[string][]byte {
	t.Helper()
	if len(bundle.MapArrays) != 1 || len(bundle.ByteArrays) != 1 || len(bundle.MapArrays[0]) != len(bundle.ByteArrays[0]) {
		t.Fatal("native arrow compilation needs one complete source bundle")
	}
	files := make(map[string][]byte, len(bundle.MapArrays[0])+1)
	for i, name := range bundle.MapArrays[0] {
		if !strings.HasPrefix(name, "/") {
			t.Fatalf("native arrow bundle path lacks logical root: %q", name)
		}
		name = strings.TrimPrefix(name, "/")
		if _, exists := files[name]; exists {
			t.Fatalf("native arrow bundle repeats path %q", name)
		}
		files[name] = bytes.Clone(bundle.ByteArrays[0][i])
	}
	files["Cargo.lock"] = lucidRead(t, "common/Cargo.lock")
	return files
}
