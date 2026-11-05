package Lyger

import (
	"github.com/l-programming-language/l/lerrors"
)

// CompilationResult owns the working semantic state and its distinct output
// views. A CLI snapshot need not own or erase the working compilation.
type CompilationResult struct {
	Compilation    *Compilation
	ExpandedSource []byte
	AML            []byte
	BackendAML     []byte
}

// Compile prepares, analyzes, and lowers source through the existing frontend.
func (ls *LygerService) Compile(source []byte, path string) (*CompilationResult, error) {
	return ls.compile(source, path, false)
}

// CompileForContracts enables definition-site and lowering recording on the
// same compilation whose source bundles the coordinator will generate.
func (ls *LygerService) CompileForContracts(source []byte, path string) (*CompilationResult, error) {
	return ls.compile(source, path, true)
}

func (ls *LygerService) compile(source []byte, path string, record bool) (*CompilationResult, error) {
	comp := ls.NewCompilation(source, path)
	if comp.Error != nil {
		return nil, comp.Error
	}
	if record {
		comp.contractRecorder = newContractRecorder(comp)
	}
	comp.Make()
	if comp.Error != nil {
		return nil, comp.Error
	}
	if comp.Program == nil || len(comp.ASMBytes) == 0 {
		return nil, lerrors.Internal.Errorf("%s: compilation did not construct a program", path)
	}
	if record {
		comp.contractRecorder.indexResolvedTargets()
	}
	return &CompilationResult{
		Compilation:    comp,
		ExpandedSource: append([]byte(nil), comp.ExpandedBytes...),
		AML:            append([]byte(nil), comp.ASMBytes...),
		BackendAML:     append([]byte(nil), comp.Program.BackendBytes...),
	}, nil
}
