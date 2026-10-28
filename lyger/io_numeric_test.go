package Lyger

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"
)

func TestSourceLogNumberSourceParity(t *testing.T) {
	var source strings.Builder
	source.WriteString("Use IO's Print from `io` as \"Print\".\nCreate \"Worker\" as a New Basic App.\n")
	var want bytes.Buffer
	literals := []string{"0", "3.5", "-12.5", "999999", "1000000", "100000000", "-100000000", "0.0001", "0.00001", "100000000000000000000", "1.2345678901234567"}
	// Exercise the retained binary64 boundaries through actual L source, rather
	// than inserting Rust expressions into the retired Log template.
	for _, value := range []float64{math.Copysign(0, -1), math.SmallestNonzeroFloat64, math.MaxFloat64, math.Float64frombits(0x0010000000000000), math.Float64frombits(4835519765965977561), math.Float64frombits(14057829775473685545)} {
		literals = append(literals, strconv.FormatFloat(value, 'g', -1, 64))
	}
	for _, literal := range literals {
		fmt.Fprintf(&source, "The App Prints %s.\n", literal)
		value, err := strconv.ParseFloat(literal, 64)
		if err != nil {
			t.Fatal(err)
		}
		want.WriteString(strconv.FormatFloat(value, 'g', -1, 64) + "\n")
	}
	if got := runSourceIO(t, []byte(source.String()), "numeric-io.l"); !bytes.Equal(got, want.Bytes()) {
		t.Fatalf("source-defined numeric IO changed source values: got %q; want %q", got, want.Bytes())
	}
}
