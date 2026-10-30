//go:build darwin || linux

package main

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestCompilerInitialize(t *testing.T) {
	// Defining our test slice. Each unit test should have the following properties:
	tCLI := compilerTestCLI()
	tConfig := getTestOptions()
	if tConfig == nil {
		t.Fatal("canonical configuration prerequisite failed")
	}
	tLog := getTestLogService(tConfig)
	tLyger := getTestLygerService(tConfig)
	tComp := getTestCompiler()
	tests := []struct {
		name    string    // The name of the test
		want    *compiler // What inputFile instance we want our function to return.
		wantErr bool      // whether we want an error.
	}{
		// Here we're declaring each unit test input and output data as defined before
		{
			"Default",
			&compiler{
				tCLI,
				*tConfig,
				tLyger,
				tLog,
				tComp,
				[]byte(INITTEST),
				nil,
				nil,
				nil,
			},
			false,
		},
		//{"Main file parameter", &compiler{tCLI, tConfig, tLog}, false, []string{"cmd", "build", "./test/main.l"}},
	}
	// Iterating over the previous test slice
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Saving the original os.Args reference
			got := newCompiler(tCLI)
			got.initialize(compilerInitOptions{AssetRoot: tConfig.AssetRoot}) // Running the function we want to test
			if !reflect.DeepEqual(got, tt.want) {                             // Asserting whether we get the correct wanted value
				t.Errorf("compiler.initialize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompilerCompile(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		config := getTestOptions()
		if config == nil {
			t.Fatal("canonical configuration prerequisite failed")
		}
		got := newCompiler(compilerTestCLI())
		got.initialize(compilerInitOptions{AssetRoot: config.AssetRoot})
		initialized := got.compilation
		result := got.compile()
		if got.err != nil || result == nil {
			t.Fatalf("builtin/dictionary compilation prerequisite; runtime not exercised: %v", got.err)
		}
		if !bytes.Equal(got.mainBytes, []byte(INITTEST)) {
			t.Fatalf("raw source changed: %s", got.mainBytes)
		}
		if len(got.asmBytes) == 0 || !bytes.Equal(got.asmBytes, result.AML) {
			t.Fatal("compile lost its returned observable AML")
		}
		if got.compilation != initialized || !reflect.DeepEqual(got.options, *config) || !bytes.Equal(got.compiledBytes, result.ExpandedSource) {
			t.Fatal("compile changed initialized state or lost expanded-source view")
		}
		var source []string
		for _, statement := range result.Compilation.SourceStatements {
			if statement.Span.Path == got.cli.mainPath {
				source = append(source, strings.TrimSpace(string(statement.Raw)))
			}
		}
		if actual := strings.Join(source, "\n"); actual != COMPTEST {
			t.Fatalf("originating user source changed:\ngot: %s\nwant: %s", actual, COMPTEST)
		}
	})
}
