//go:build darwin || linux

package cli

import "testing"

func TestEmptyApplicationExecutes(t *testing.T) {
	nativeCases.Parallel(t)
	cli := standalone(t)
	dir := project(t, "Create \"Empty\" as a New Basic App.\nExecute Empty.\nIf Empty has an Error, exit with Empty's Error.\nShow Empty's Output.\n", nil)
	binary := buildProgram(t, cli, dir, "main.l", "")
	outcome(t, binary, dir, nil, 0, "", "")
}
