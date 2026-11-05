// Author: Connor Sanders

package Logiceme

import (
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
	prepared, err := ls.prepareGeneration(jit, selected, nil, false)
	if err != nil {
		return Executable{}, err
	}
	return prepared.Render()
}

func prepareScannedSource(scanner Scanner, system SegmentalSystem, templates map[string][]byte, includeRuntime bool, workspace *WorkspaceRuntime) (*Program, error) {
	if workspace != nil && len(scanner.Program.Books) != 1 {
		return nil, lerrors.InvalidInput.Errorf("Rust workspace entry requires exactly one application book")
	}
	if workspace != nil && scanner.Program.Books[0].Name == workspace.Package {
		return nil, lerrors.InvalidInput.Errorf("Rust workspace runtime package collides with entry package %q", workspace.Package)
	}
	program := Program{SegmentalSystem: system, Scanner: scanner}
	if err := program.Compile(); err != nil {
		return nil, err
	}
	for i := range program.Compiler.Apps {
		if err := prepareSourceHeaders(&program.Compiler.Apps[i], includeRuntime && i == 0, workspace); err != nil {
			return nil, sourceGraphError(program.Compiler.Apps[i].Name, err)
		}
		settings, err := bookSettings(program.Scanner.Program.Tags, program.Scanner.Program.Books[i])
		if err != nil {
			return nil, sourceGraphError(program.Compiler.Apps[i].Name, err)
		}
		if libraryOnlyTargets(program.Compiler.Apps[i].sourceTargets) && len(settings) != 0 {
			return nil, sourceGraphError(program.Compiler.Apps[i].Name, lerrors.InvalidInput.Errorf("library-only package cannot supply executable configuration"))
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
				return nil, err
			}
		}
	}
	return &program, nil
}

func evaluateScannedSource(scanner Scanner, system SegmentalSystem, templates map[string][]byte, includeRuntime bool, workspace *WorkspaceRuntime) (Executable, error) {
	program, err := prepareScannedSource(scanner, system, templates, includeRuntime, workspace)
	if err != nil {
		return Executable{}, err
	}
	prepared := ContractGeneration{program: program, templates: templates, includeRuntime: includeRuntime, workspace: workspace}
	return prepared.Render()
}
