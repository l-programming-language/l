//go:build darwin || linux

package cli

import (
	"path/filepath"
	"testing"
)

func TestCoordinatedShowExactStreams(t *testing.T) {
	nativeCases.Parallel(t)
	cli := standalone(t)
	dir := project(t, "Define \"First\" as `first|`.\nDefine \"Second\" as `second|`.\nDefine \"Third\" as `third|`.\nShow First and Second and Third.\nShow `last`.\n", nil)
	binary := buildProgram(t, cli, t.TempDir(), filepath.Join(dir, "main.l"), "")
	outcome(t, binary, t.TempDir(), nil, 0, "first|second|third|last", "")
}
