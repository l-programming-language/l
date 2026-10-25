//go:build darwin || linux

package Logiceme

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// These native boundaries complement ordinary decimal values and UTF-8 Error
// messages. Go's formatter is the oracle; the actual selected Rust support is
// compiled unchanged.
func TestRuntimeScalarParity(t *testing.T) {
	t.Parallel()
	var source strings.Builder
	source.WriteString("mod runtime;\n#[path = \"process_unix.rs\"] mod process;\nfn main() -> std::process::ExitCode {\n")
	for _, number := range []float64{0, math.Copysign(0, -1), 3.5, -12.5, 999999, 1000000, -1e20, 0.0001, 0.00001, 1.2345678901234567, math.SmallestNonzeroFloat64, math.MaxFloat64,
		// Exact decimal midpoints: Rust shortest formatting selects the odd
		// upper digit; the established Go contract selects the even neighbor.
		math.Float64frombits(4835519765965977561), math.Float64frombits(14057829775473685545),
	} {
		want := strconv.FormatFloat(number, 'g', -1, 64)
		fmt.Fprintf(&source, "assert_eq!(runtime::format_number(f64::from_bits(%d)), b%q);\n", math.Float64bits(number), want)
	}
	source.WriteString(`let executor = tokio::runtime::Builder::new_current_thread().enable_all().build().unwrap();
let status = executor.block_on(async {
    let runtime = runtime::Runtime::new().unwrap();
    runtime.finish(Err(runtime::Error::from_bytes(vec![b'a', 0xff, b'\n']))).await
});
executor.shutdown_background();
std::process::ExitCode::from(status)
}`)
	path := filepath.Join(t.TempDir(), "scalars.rs")
	if err := os.WriteFile(path, []byte(source.String()), 0600); err != nil {
		t.Fatal(err)
	}
	_, build := runtimeFixture(t, false, path)
	driver := runtimeDriver(t, build)
	runtimeExecute(t, driver, nil, nil, 1, "", "a\xff\n\n")
}
