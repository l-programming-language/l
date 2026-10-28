package Logiceme

import (
	"bytes"
	"fmt"
	"path"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// WorkspaceRuntime describes a coordinator-owned Cargo library dependency.
// It is not AML metadata. RelativeDirectory is relative to the entry package.
type WorkspaceRuntime struct {
	Package           string
	RelativeDirectory string
}

// EvaluateWorkspaceEntry assembles one composed Rust entry and its separate
// runtime library. Evaluate and GenerationOptions retain their inline-runtime
// source-bundle contract. The coordinator owns the enclosing workspace/layout.
func (ls *LogicemeService) EvaluateWorkspaceEntry(jit []byte, runtime WorkspaceRuntime) (entry, support Executable, err error) {
	if ls == nil || ls.resources == nil {
		return Executable{}, Executable{}, lerrors.InvalidConfiguration.Errorf("source generator configuration is required")
	}
	if err := validateWorkspaceRuntime(runtime); err != nil {
		return Executable{}, Executable{}, err
	}
	system, err := loadSegmentalSystem(ls.resources)
	if err != nil {
		return Executable{}, Executable{}, fmt.Errorf("load source generator: %w", err)
	}
	templates, err := loadSupportTemplates(ls.resources, true)
	if err != nil {
		return Executable{}, Executable{}, err
	}
	entry, err = evaluateSourceWithSupport(jit, system, templates, true, &runtime)
	if err != nil {
		return Executable{}, Executable{}, err
	}
	support, err = workspaceRuntimeLibrary(runtime.Package, templates)
	if err != nil {
		return Executable{}, Executable{}, err
	}
	return entry, support, nil
}

func validateWorkspaceRuntime(runtime WorkspaceRuntime) error {
	if err := validateSourceModuleName(runtime.Package); err != nil {
		return fmt.Errorf("Rust workspace runtime package: %w", err)
	}
	directory := runtime.RelativeDirectory
	if directory == "" || directory == "." || directory == ".." || path.IsAbs(directory) || path.Clean(directory) != directory || strings.ContainsAny(directory, "\\\x00\r\n") {
		return lerrors.InvalidInput.Errorf("Rust workspace runtime requires a canonical relative library directory")
	}
	return nil
}

func workspaceRuntimeLibrary(name string, templates map[string][]byte) (Executable, error) {
	const root = "#![forbid(unsafe_code)]\n\npub mod runtime;\n#[cfg(any(target_os = \"linux\", target_os = \"macos\"))]\n#[path = \"process_unix.rs\"]\nmod process;\n#[cfg(not(any(target_os = \"linux\", target_os = \"macos\")))]\n#[path = \"process_other.rs\"]\nmod process;\n"
	dependencies, err := runtimeDependencies(nil)
	if err != nil {
		return Executable{}, err
	}
	targets := []preparedSourceTarget{{TargetDescriptor: TargetDescriptor{Kind: "lib", Name: name, Root: "lib"}, Path: "/src/lib.rs"}}
	cargo, err := renderTargetCargoManifest(templates["install/cargo.lu"], name, targets, dependencies)
	if err != nil {
		return Executable{}, err
	}
	return Executable{
		Name:             name,
		MapArrays:        [][]string{{"/src/lib.rs", "/src/runtime.rs", "/src/runtime/host.rs", "/src/process_unix.rs", "/src/process_other.rs", "/Cargo.toml"}},
		ByteArrays:       [][][]byte{{[]byte(root), bytes.Clone(templates["runtime/runtime.rs"]), bytes.Clone(templates["runtime/host.lu"]), bytes.Clone(templates["runtime/process_unix.lu"]), bytes.Clone(templates["runtime/process_other.lu"]), cargo}},
		DependencyArrays: [][]string{dependencies},
		EnvDepArrays:     [][]string{nil},
	}, nil
}
