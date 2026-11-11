package Logiceme

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/l-programming-language/l/contracts"
)

func TestGenerationCompleteBundle(t *testing.T) {
	input := lucidRead(t, "grouping.aml.txt")
	original := bytes.Clone(input)
	service := NewLogicemeService(backendTestResources(t))
	want := lucidExpectedBundle(t, "grouping")
	var evaluate func([]byte) (Executable, error) = service.Evaluate
	for _, name := range []string{"first", "reused"} {
		t.Run(name, func(t *testing.T) {
			got, err := evaluate(input)
			if err != nil {
				t.Fatalf("Rust source generation: %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("default generation differs from the complete Rust source oracle: paths=%v; want=%v", got.MapArrays, want.MapArrays)
			}
			if !bytes.Equal(input, original) {
				t.Fatal("generation mutated its AML input")
			}
		})
	}
}

func TestGenerationSettingsKeepStartupAndInventory(t *testing.T) {
	plain := lucidRead(t, "grouping.aml.txt")
	input := bytes.Replace(plain, []byte("<SETTINGS></SETTINGS>"), []byte(`<SETTINGS><JSON>{"Key":"GREETING","Type":"String","Value":"hello"}</JSON></SETTINGS>`), 1)
	original := bytes.Clone(input)
	service := NewLogicemeService(nil)
	prepared, err := service.PrepareContractGeneration(input)
	if err != nil {
		t.Fatal(err)
	}
	startupSteps := 0
	for _, step := range prepared.Inventory().Steps {
		if step.Kind == "implicit-configuration" {
			startupSteps++
			if !strings.HasSuffix(step.Name, "/configuration-startup") || step.Implementation == "" {
				t.Fatalf("configuration startup lacks its implementation identity: %+v", step)
			}
		}
	}
	if startupSteps != 1 || prepared.Inventory().TransitiveCoverage != contracts.Unknown {
		t.Fatal("configured entry must retain one startup boundary and unresolved native conformance")
	}
	bundle, err := prepared.Render()
	if err != nil {
		t.Fatal(err)
	}
	var main, config []byte
	for index, name := range bundle.MapArrays[0] {
		switch name {
		case "/src/main.rs":
			main = bundle.ByteArrays[0][index]
		case "/src/config.rs":
			config = bundle.ByteArrays[0][index]
		}
	}
	startup := []byte("config::initialize_configuration_settings()")
	start, body := bytes.Index(main, startup), bytes.Index(main, []byte("let result"))
	if start < 0 || body < start || bytes.Count(main, startup) != 1 || !bytes.Contains(main, []byte("mod config;")) {
		t.Fatalf("configuration startup must precede the composed body once:\n%s", main)
	}
	if !bytes.Contains(config, []byte(`settings.insert("GREETING", "hello");`)) || !bytes.Equal(input, original) {
		t.Fatal("selected configuration or original AML changed")
	}
	priorMain := bytes.Clone(main)
	again, err := service.Evaluate(plain)
	if err != nil || !reflect.DeepEqual(again, lucidExpectedBundle(t, "grouping")) || !bytes.Equal(main, priorMain) {
		t.Fatalf("configured generation contaminated the reused service or prior bundle: %v", err)
	}
}
