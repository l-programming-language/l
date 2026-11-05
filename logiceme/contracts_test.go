package Logiceme

import (
	"reflect"
	"testing"

	"github.com/l-programming-language/l/contracts"
)

func TestContractsBackendPreflight(t *testing.T) {
	service := NewLogicemeService(nil)
	for _, runtime := range []bool{false, true} {
		input := libraryFixture(t)
		if runtime {
			input = workspaceRuntimeInput()
		}
		options := GenerationOptions{IncludeRuntime: runtime}
		ordinary, err := service.Evaluate(input, options)
		if err != nil {
			t.Fatal(err)
		}
		prepared, err := service.PrepareContractGeneration(input, options)
		if err != nil {
			t.Fatal(err)
		}
		// A caller cannot change a prepared proof by editing an exported effect row.
		prepared.inventory.Unknowns[0].ExcludedEffects = []contracts.EffectKind{"block"}
		inventory := prepared.Inventory()
		inventory.Unknowns[0].ExcludedEffects[0] = "allocate"
		if prepared.inventory.Unknowns[0].ExcludedEffects[0] != "block" {
			t.Fatal("snapshot shares effect exclusions with prepared inventory")
		}
		inventory = prepared.Inventory()
		if len(inventory.Native) == 0 || inventory.TransitiveCoverage != contracts.Unknown {
			t.Fatal("opaque native closure was erased")
		}
		generated, err := prepared.Render()
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(ordinary, generated) {
			t.Fatal("preflight changed emitted bundle")
		}
		if err := contracts.VerifyInventory(inventory, prepared.Inventory()); err != nil {
			t.Fatal(err)
		}
		inventory.Steps[0].Name = "tampered"
		if err := contracts.VerifyInventory(inventory, prepared.Inventory()); err == nil {
			t.Fatal("inventory tampering retained old assurance")
		}
	}
}

func TestContractsWorkspacePreflight(t *testing.T) {
	service := NewLogicemeService(nil)
	placement := WorkspaceRuntime{Package: "l_runtime_support", RelativeDirectory: "../support/runtime"}
	ordinary, support, err := service.EvaluateWorkspaceEntry(workspaceRuntimeInput(), placement)
	if err != nil {
		t.Fatal(err)
	}
	prepared, err := service.PrepareWorkspaceContractGeneration(workspaceRuntimeInput(), placement)
	if err != nil {
		t.Fatal(err)
	}
	generated, library, err := prepared.RenderWorkspace()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(ordinary, generated) || !reflect.DeepEqual(support, library) {
		t.Fatal("workspace preflight changed entry sources")
	}
}

func TestContractsBackendDependenciesInvalidateAssurance(t *testing.T) {
	service := NewLogicemeService(nil)
	system, err := loadSegmentalSystem(service.resources)
	if err != nil {
		t.Fatal(err)
	}
	templates, err := loadSupportTemplates(service.resources, false)
	if err != nil {
		t.Fatal(err)
	}
	input := libraryFixture(t)
	first, err := prepareContractGeneration(input, system, templates, GenerationOptions{}, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	// A temporary in-memory body change leaves the frozen resource untouched.
	system.Declaration.Content = append(append([]byte(nil), system.Declaration.Content...), '\n')
	second, err := prepareContractGeneration(input, system, templates, GenerationOptions{}, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := contracts.VerifyInventory(first.Inventory(), second.Inventory()); err == nil {
		t.Fatal("segment body change retained old implementation assurance")
	}
	for _, prepared := range []*ContractGeneration{first, second} {
		if prepared.Inventory().TransitiveCoverage != contracts.Unknown {
			t.Fatal("native fingerprints were mistaken for transitive conformance")
		}
	}
}
