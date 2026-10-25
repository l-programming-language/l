//go:build darwin || linux

package composition

import (
	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	"github.com/l-programming-language/l/internal/testsupport/phase1"
	"testing"
)

// Cases own their projects, controllers and process groups. Admission starts
// before the existing build/runtime watchdogs; each case keeps its causal order.
var nativeCases = make(nativetest.Admission, 2)

func TestPhase1CLI(t *testing.T) {
	t.Parallel()
	for _, id := range []string{"F01", "F03", "F13", "F14", "F15", "F16", "F24", "N01", "N02", "N03", "N04", "N09", "N10", "N11", "N12"} {
		id := id
		t.Run(id, func(t *testing.T) {
			nativeCases.Parallel(t)
			phase1.RunFixture(t, id)
		})
	}
}

func TestMain(m *testing.M) { phase1.Main(m) }
