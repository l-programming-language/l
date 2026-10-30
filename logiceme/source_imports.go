package Logiceme

import (
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// ReadImports returns authored file imports from complete AML. It reads the
// existing Imports metadata; opaque native leaves are never parsed for calls.
func (ls *LogicemeService) ReadImports(jit []byte) ([]string, error) {
	if ls == nil || ls.resources == nil {
		return nil, lerrors.InvalidConfiguration.Errorf("source generator configuration is required")
	}
	system, err := loadSegmentalSystem(ls.resources)
	if err != nil {
		return nil, err
	}
	scanner, err := loadScannerBytes(jit)
	if err != nil {
		return nil, err
	}
	if err := scanSourceProgram(&scanner); err != nil {
		return nil, err
	}
	program := Program{SegmentalSystem: system, Scanner: scanner}
	if err := program.Compile(); err != nil {
		return nil, err
	}
	var imports []string
	seen := make(map[string]bool)
	for index := range program.Compiler.Apps {
		app := &program.Compiler.Apps[index]
		if err := app.eachObject(func(object *Object) error {
			var analysis sourceFileAnalysis
			if _, _, _, err := prepareSourceFile(object.Sentence, &app.SegmentalSystem, &analysis); err != nil {
				return err
			}
			for _, line := range analysis.imports {
				name := strings.TrimSuffix(strings.TrimPrefix(line, "use "), ";\n")
				if !seen[name] {
					imports = append(imports, name)
					seen[name] = true
				}
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}
	return imports, nil
}
