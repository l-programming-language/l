// Author: Connor Sanders

package Logiceme

import (
	"fmt"
	"io/fs"

	distribution "github.com/l-programming-language/l"
	"github.com/l-programming-language/l/lerrors"
)

// LogicemeService packages the functionality of the logeme system for use in the main package
type LogicemeService struct {
	resources fs.FS
}

// NewLogicemeService selects a complete resource provider, or embedded defaults.
func NewLogicemeService(resources fs.FS) *LogicemeService {
	if resources == nil {
		resources = distribution.Files()
	}
	return &LogicemeService{resources: resources}
}

// GenerationOptions selects optional support for source generation. Its zero
// value generates source without the process runtime.
type GenerationOptions struct {
	IncludeRuntime bool
}

// Evaluate assembles source code based on the contents of an ASM JIT file.
func (ls *LogicemeService) Evaluate(jit []byte, options ...GenerationOptions) (Executable, error) {
	if ls == nil || ls.resources == nil {
		return Executable{}, lerrors.InvalidConfiguration.Errorf("source generator configuration is required")
	}
	if len(options) > 1 {
		return Executable{}, lerrors.InvalidInput.Errorf("source generation accepts at most one options value")
	}
	var selected GenerationOptions
	if len(options) == 1 {
		selected = options[0]
	}
	segSys, err := loadSegmentalSystem(ls.resources)
	if err != nil {
		return Executable{}, fmt.Errorf("load source generator: %w", err)
	}
	return evaluateSource(jit, segSys, ls.resources, selected.IncludeRuntime)
}

// Support assembly is deliberately separate from ordinary source linking. It renders
// only the existing support-template inputs; it does not translate source,
// prepare signatures, infer dependencies, or resolve dependency versions.
func evaluateSource(input []byte, system SegmentalSystem, resources fs.FS, includeRuntime bool) (Executable, error) {
	templates, err := loadSupportTemplates(resources, includeRuntime)
	if err != nil {
		return Executable{}, err
	}
	return evaluateSourceWithSupport(input, system, templates, includeRuntime, nil)
}

func evaluateSourceWithSupport(input []byte, system SegmentalSystem, templates map[string][]byte, includeRuntime bool, workspace *WorkspaceRuntime) (Executable, error) {
	scanner, err := loadScannerBytes(input)
	if err != nil {
		return Executable{}, err
	}
	if err := scanSourceProgram(&scanner); err != nil {
		return Executable{}, err
	}
	if workspace != nil && len(scanner.Program.Books) != 1 {
		return Executable{}, lerrors.InvalidInput.Errorf("Rust workspace entry requires exactly one application book")
	}
	if workspace != nil && scanner.Program.Books[0].Name == workspace.Package {
		return Executable{}, lerrors.InvalidInput.Errorf("Rust workspace runtime package collides with entry package %q", workspace.Package)
	}
	program := Program{SegmentalSystem: system, Scanner: scanner}
	if err := program.Compile(); err != nil {
		return Executable{}, err
	}
	for i := range program.Compiler.Apps {
		if err := prepareSourceHeaders(&program.Compiler.Apps[i], includeRuntime && i == 0, workspace); err != nil {
			return Executable{}, sourceGraphError(program.Compiler.Apps[i].Name, err)
		}
		settings, err := bookSettings(program.Scanner.Program.Tags, program.Scanner.Program.Books[i])
		if err != nil {
			return Executable{}, sourceGraphError(program.Compiler.Apps[i].Name, err)
		}
		if libraryOnlyTargets(program.Compiler.Apps[i].sourceTargets) && len(settings) != 0 {
			return Executable{}, sourceGraphError(program.Compiler.Apps[i].Name, lerrors.InvalidInput.Errorf("library-only package cannot supply executable configuration"))
		}
		// Explicit Settings require startup before application effects. The
		// runtime entry already owns this operation; plain composed entries
		// acquire it through their prepared FW body, never a source rewrite.
		if len(settings) != 0 && !(includeRuntime && i == 0) {
			if err := program.Compiler.Apps[i].eachObject(func(object *Object) error {
				if object.header.primary && !object.header.literal {
					object.header.structural = true
					object.header.initializeConfig = true
				}
				return nil
			}); err != nil {
				return Executable{}, err
			}
		}
	}
	if err := program.Assemble(); err != nil {
		return Executable{}, err
	}
	program.Link()
	if err := program.linkBundleSupport(templates, includeRuntime, workspace); err != nil {
		return Executable{}, err
	}
	return program.Executable, nil
}
