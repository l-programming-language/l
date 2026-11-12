package Lyger

import "github.com/l-programming-language/l/lerrors"

// ValidateStandaloneApp checks the source boundary supported without process
// supervision. App body definitions have already been composed into its book;
// ordered outer instructions must never be discarded when publishing that App.
func (r *CompilationResult) ValidateStandaloneApp() error {
	program, err := r.buildProgram()
	if err != nil {
		return err
	}
	if len(program.Apps) != 1 {
		return lerrors.Unsupported.Errorf("standalone compilation requires exactly one App")
	}
	semantics := r.Compilation.Semantics
	if semantics == nil {
		return lerrors.Internal.Errorf("standalone compilation requires semantic instructions")
	}
	var check func([]*instruction) error
	check = func(instructions []*instruction) error {
		for _, current := range instructions {
			if current.Kind == instructionBlock && current.Block != nil {
				if err := check(current.Block.Body); err != nil {
					return err
				}
				continue
			}
			return current.Source.wrapError(lerrors.Unsupported.Errorf("outer-program %s is unsupported without Runtime; only a standalone App body can execute", current.Kind))
		}
		return nil
	}
	if err := check(semantics.Instructions); err != nil {
		return err
	}
	for _, continuation := range semantics.Continuations {
		if err := check(continuation.Body); err != nil {
			return err
		}
	}
	return nil
}
