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
	comp := ls.NewCompilation(source, path)
	if comp.Error != nil {
		return nil, comp.Error
	}
	comp.Make()
	if comp.Error != nil {
		return nil, comp.Error
	}
	if comp.Program == nil || len(comp.ASMBytes) == 0 {
		return nil, lerrors.Internal.Errorf("%s: compilation did not construct a program", path)
	}
	return &CompilationResult{
		Compilation:    comp,
		ExpandedSource: append([]byte(nil), comp.ExpandedBytes...),
		AML:            append([]byte(nil), comp.ASMBytes...),
		BackendAML:     append([]byte(nil), comp.Program.BackendBytes...),
	}, nil
}
