//go:build darwin || linux

package lifecycle

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
	for _, id := range []string{"F02", "F04", "F05", "F06", "F07", "F08", "F09", "F21", "F22", "F23", "N05", "N06", "N07"} {
		id := id
		t.Run(id, func(t *testing.T) {
			nativeCases.Parallel(t)
			phase1.RunFixture(t, id)
		})
	}
}

func TestMain(m *testing.M) { phase1.Main(m) }
