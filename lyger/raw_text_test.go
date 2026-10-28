package Lyger

import (
	"bytes"
	"context"
	"os/exec"
	"reflect"
	"testing"
	"time"

	Logiceme "github.com/l-programming-language/l/logiceme"
)

func TestDefinitionRawTextLiteral(t *testing.T) {
	for _, tc := range []struct{ name, value, literal string }{
		{"invalid octets", string([]byte{'A', 0xff, 'B', 0x80, 'C'}), `b"\x41\xff\x42\x80\x43"`},
		{"mixed Unicode and markup", "雪" + string([]byte{0xff, '|', '~', 0, '"', '\\'}), `b"\xe9\x9b\xaa\xff\x7c\x7e\x00\x22\x5c"`},
		{"incomplete sequence", string([]byte{0xf0, 0x9f}), `b"\xf0\x9f"`},
		{"valid Unicode unchanged", "雪|~", `"雪\x7c\x7e"`},
	} {
		for _, slot := range []string{"Text", "Thing"} {
			t.Run(tc.name+"/"+slot, func(t *testing.T) {
				template := parseTestDefinition(t, `<DEF>E{|Statement: ~consume(<@Value|`+slot+`@>)~|}<DEF>`)
				bindings := map[string]DefinitionArgument{"Value": {Kind: "Text", Text: tc.value}}
				first := bindTestDefinition(t, template, bindings)
				want := `E{|Statement: ~consume(` + tc.literal + `)~|}`
				if first.Expression != want || first.OriginalExpression != want || len(first.Bindings) != 1 || first.Bindings[0].Emitted != tc.literal {
					t.Fatalf("native Text bytes changed: %#v; expected %s", first, want)
				}
				second := bindTestDefinition(t, template, bindings)
				if !reflect.DeepEqual(first, second) || bindings["Value"].Text != tc.value {
					t.Fatal("native Text rendering changed caller data or repeated output")
				}
			})
		}
	}
}

func runRawTextBundle(t *testing.T, bundle Logiceme.Executable) []byte {
	t.Helper()
	binary := buildActualBundle(t, bundle, "Worker", false, "")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil || stderr.Len() != 0 {
		t.Fatalf("raw Text application: %v; stderr=%q", err, stderr.Bytes())
	}
	return stdout.Bytes()
}

func TestSourceLogPreservesRawText(t *testing.T) {
	raw := []byte{'A', 0xff, 'B', 0x80, 'C'}
	source := append([]byte("Use IO's Print from `io` as \"Print\".\nCreate \"Worker\" as a New Basic App.\nThe App Prints `"), raw...)
	source = append(source, []byte("`.\nThe App Prints 1000000.\n")...)
	want := append(bytes.Clone(raw), []byte("\n1e+06\n")...)
	for _, payload := range [][]byte{{'a', 0xff, 'b'}, []byte("雪|~"), []byte("retained"), []byte("once")} {
		source = append(source, []byte("The App Prints `")...)
		source = append(source, payload...)
		source = append(source, []byte("`.\n")...)
		want = append(want, payload...)
		want = append(want, '\n')
	}
	source = append(source, []byte("The App Prints false.\n")...)
	want = append(want, []byte("false\n")...)
	if got := runSourceIO(t, source, "raw-text.l"); !bytes.Equal(got, want) {
		t.Fatalf("source-defined IO changed scalar bytes: got %x; want %x", got, want)
	}
}

func runSourceIO(t *testing.T, source []byte, path string) []byte {
	t.Helper()
	before := bytes.Clone(source)
	result, err := NewLygerService(phase1Options(t)).Compile(source, path)
	if err != nil {
		t.Fatal(err)
	}
	units, err := result.ApplicationUnits()
	if err != nil || len(units) != 1 {
		t.Fatalf("IO application units: %v", err)
	}
	bundle, err := Logiceme.NewLogicemeService(phase1Options(t).Resources).Evaluate(units[0].BackendAML, Logiceme.GenerationOptions{})
	recordBundle(t, bundle)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(source, before) {
		t.Fatal("IO compilation changed source input")
	}
	return runRawTextBundle(t, bundle)
}
