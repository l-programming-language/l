package Lyger

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"

	Logiceme "github.com/l-programming-language/l/logiceme"
)

// The binary64 formatting policy is an independent oracle.
// Instantiate the selected source-defined Log; no generated source is patched.
func TestSourceLogNumberFormatting(t *testing.T) {
	prototype, err := os.ReadFile("../stdlib/eng/io/dictionary.l")
	if err != nil {
		t.Fatal(err)
	}
	first, last := bytes.IndexByte(prototype, '`'), bytes.LastIndexByte(prototype, '`')
	if first < 0 || last <= first {
		t.Fatal("source-defined Log has no native definition")
	}
	var body []string
	var want bytes.Buffer
	for _, value := range []float64{
		0, math.Copysign(0, -1), 3.5, -12.5,
		999999, 1000000, -100000000, 0.0001, 0.00001,
		1e20, 1.2345678901234567, math.SmallestNonzeroFloat64, math.MaxFloat64,
		math.Float64frombits(0x0010000000000000),
		math.Float64frombits(4835519765965977561), math.Float64frombits(14057829775473685545),
	} {
		expression := fmt.Sprintf("f64::from_bits(%d)", math.Float64bits(value))
		template, err := ParseDefinitionTemplate([]byte(strings.Replace(string(prototype[first:last+1]), "<@Output|Thing@>", expression, 1)), SourceSpan{})
		if err != nil {
			t.Fatal(err)
		}
		instance, err := template.Instantiate(nil)
		if err != nil {
			t.Fatal(err)
		}
		body = append(body, instance.Expression)
		want.WriteString(strconv.FormatFloat(value, 'g', -1, 64) + "\n")
	}
	// Native integer Thing operands keep their own Display contract rather than
	// being coerced into L's binary64 representation by a vocabulary special case.
	for _, expression := range []string{"100000000_i64", "-100000000_i64"} {
		template, err := ParseDefinitionTemplate([]byte(strings.Replace(string(prototype[first:last+1]), "<@Output|Thing@>", expression, 1)), SourceSpan{})
		if err != nil {
			t.Fatal(err)
		}
		instance, err := template.Instantiate(nil)
		if err != nil {
			t.Fatal(err)
		}
		body = append(body, instance.Expression)
	}
	want.WriteString("100000000\n-100000000\n")
	document := `<PROGRAM><STRUCTURE><BOOK "Worker"><OUTLINE><JSON>{"AppMeta":{"AppGenDir":"/src"}}</JSON><JSON>{"Services":[{"Name":"main","Location":"/src/","GenericObjects":["main"]}]}</JSON><JSON>{"Objects":[{"Name":"main","Type":"root"}]}</JSON></OUTLINE><SETTINGS></SETTINGS><CONTENT><CHAPTER "main"><SENTENCE "main"><|W|><|S|>` + "`A`" + `</S|><|S|>` + "`FW{|Key: ~main~, Outputs: ~&[()]&~|}^W^(" + strings.Join(body, "*") + ")`" + `</S|></W|></SENTENCE></CHAPTER></CONTENT></BOOK></STRUCTURE></PROGRAM>`
	bundle, err := Logiceme.NewLogicemeService(phase1Options(t).Resources).Evaluate([]byte(document), Logiceme.GenerationOptions{})
	recordBundle(t, bundle)
	if err != nil {
		t.Fatal(err)
	}
	if got := runRawTextBundle(t, bundle); !bytes.Equal(got, want.Bytes()) {
		t.Fatalf("source-defined Number formatting differs from binary64 contract:\ngot %q\nwant %q", got, want.Bytes())
	}
}

func TestSourceLogNumberSourceParity(t *testing.T) {
	var source strings.Builder
	source.WriteString("Use IO's Print from `io` as \"Print\".\nCreate \"Worker\" as a New Basic App.\n")
	var want bytes.Buffer
	for _, literal := range []string{"0", "3.5", "-12.5", "999999", "1000000", "100000000", "-100000000", "0.0001", "0.00001", "100000000000000000000", "1.2345678901234567"} {
		fmt.Fprintf(&source, "The App Prints %s.\n", literal)
		value, err := strconv.ParseFloat(literal, 64)
		if err != nil {
			t.Fatal(err)
		}
		want.WriteString(strconv.FormatFloat(value, 'g', -1, 64) + "\n")
	}
	input := []byte(source.String())
	before := bytes.Clone(input)
	result, err := NewLygerService(phase1Options(t)).Compile(input, "numeric-io.l")
	if err != nil {
		t.Fatal(err)
	}
	units, err := result.ApplicationUnits()
	if err != nil || len(units) != 1 {
		t.Fatalf("numeric IO application units: %v", err)
	}
	bundle, err := Logiceme.NewLogicemeService(phase1Options(t).Resources).Evaluate(units[0].BackendAML, Logiceme.GenerationOptions{})
	recordBundle(t, bundle)
	if err != nil {
		t.Fatal(err)
	}
	if got := runRawTextBundle(t, bundle); !bytes.Equal(got, want.Bytes()) {
		t.Fatalf("source-defined numeric IO changed source values: got %q; want %q", got, want.Bytes())
	}
	if !bytes.Equal(input, before) {
		t.Fatal("numeric IO compilation changed source input")
	}
}
