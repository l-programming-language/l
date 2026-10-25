//go:build darwin || linux

package shutdown

import (
	"github.com/l-programming-language/l/internal/testsupport/phase1"
	"testing"
)

func TestPhase1Publication(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"BuildDoesNotExecute", "RelocatedInstalledAssets", "FailedPublicationPreservesExisting", "ImmutableCompanionGeneration"} {
		name := name
		t.Run(name, func(t *testing.T) {
			nativeCases.Parallel(t)
			phase1.RunPublication(t, name)
		})
	}
}
