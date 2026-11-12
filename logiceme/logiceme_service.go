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

// Evaluate assembles source code based on the contents of an ASM JIT file.
func (ls *LogicemeService) Evaluate(jit []byte) (Executable, error) {
	if ls == nil || ls.resources == nil {
		return Executable{}, lerrors.InvalidConfiguration.Errorf("source generator configuration is required")
	}
	prepared, err := ls.prepareGeneration(jit, false)
	if err != nil {
		return Executable{}, err
	}
	return prepared.Render()
}

func prepareScannedSource(scanner Scanner, system SegmentalSystem) (*Program, error) {
	program := Program{SegmentalSystem: system, Scanner: scanner}
	if err := program.Compile(); err != nil {
		return nil, err
	}
	for i := range program.Compiler.Apps {
		if err := prepareSourceHeaders(&program.Compiler.Apps[i]); err != nil {
			return nil, sourceGraphError(program.Compiler.Apps[i].Name, err)
		}
		settings, err := bookSettings(program.Scanner.Program.Tags, program.Scanner.Program.Books[i])
		if err != nil {
			return nil, sourceGraphError(program.Compiler.Apps[i].Name, err)
		}
		if libraryOnlyTargets(program.Compiler.Apps[i].sourceTargets) && len(settings) != 0 {
			return nil, sourceGraphError(program.Compiler.Apps[i].Name, lerrors.InvalidInput.Errorf("library-only package cannot supply executable configuration"))
		}
		// Explicit Settings require startup before application effects, composed
		// through the prepared FW body without rewriting completed source.
		if len(settings) != 0 {
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
