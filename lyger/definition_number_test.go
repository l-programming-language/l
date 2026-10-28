package Lyger

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/l-programming-language/l/lerrors"
	Logiceme "github.com/l-programming-language/l/logiceme"
)

// These native spellings and IEEE-754 bit patterns are independent of the
// production formatter. Both public bindings and native execution use them.
var definitionNumberCases = []struct {
	name, input, literal string
	bits                 uint64
}{
	{"zero", "0", "0_f64", 0x0000000000000000},
	{"negative zero", "-0", "-0_f64", 0x8000000000000000},
	{"integer", "3", "3_f64", 0x4008000000000000},
	{"fraction", "-12.5", "-12.5_f64", 0xc029000000000000},
	{"positive sign", "+3", "3_f64", 0x4008000000000000},
	{"integral decimal", "3.0", "3_f64", 0x4008000000000000},
	{"exponent", "1e6", "1e+06_f64", 0x412e848000000000},
	{"hexadecimal", "0x1.8p+1", "3_f64", 0x4008000000000000},
	{"underscores", "1_000", "1000_f64", 0x408f400000000000},
	{"rounded integer", "9007199254740993", "9.007199254740992e+15_f64", 0x4340000000000000},
	{"smallest subnormal", "5e-324", "5e-324_f64", 0x0000000000000001},
	{"largest finite", "1.7976931348623157e308", "1.7976931348623157e+308_f64", 0x7fefffffffffffff},
	{"underflow", "1e-4000", "0_f64", 0x0000000000000000},
	{"negative underflow", "-1e-4000", "-0_f64", 0x8000000000000000},
}

func TestDefinitionNumberEncoding(t *testing.T) {
	for _, binding := range []struct{ slot, kind string }{
		{"Number", "Number"}, {"Thing", "Number"},
		{"nUmBeR", "NuMbEr"}, {"tHiNg", "nUmBeR"},
	} {
		for _, tc := range definitionNumberCases {
			t.Run(binding.slot+"/"+tc.name, func(t *testing.T) {
				raw := []byte(`<DEF>BOD^E{|Statement: ~let value: f64 = <@Value|` + binding.slot + `@>;~|}<DEF>`)
				before := bytes.Clone(raw)
				span := SourceSpan{Path: "numbers.l", Start: SourcePosition{Line: 7, Column: 3}}
				template, err := ParseDefinitionTemplate(raw, span)
				if err != nil {
					t.Fatal(err)
				}
				argument := DefinitionArgument{Kind: binding.kind, Text: tc.input}
				bindings := map[string]DefinitionArgument{"Value": argument}
				first, err := template.Instantiate(bindings)
				if err != nil {
					t.Fatal(err)
				}
				want := `BOD^E{|Statement: ~let value: f64 = ` + tc.literal + `;~|}`
				ledger := []DefinitionBinding{{Slot: DefinitionSlot{Name: "Value", Type: binding.slot}, Kind: binding.kind, Emitted: tc.literal}}
				if first.Expression != want || first.OriginalExpression != want || !reflect.DeepEqual(first.Bindings, ledger) {
					t.Errorf("Number encoding = %#v; want expression %q and ledger %#v", first, want, ledger)
				}
				if !bytes.Equal(raw, before) || !bytes.Equal(template.Original, before) || !bytes.Equal(first.Original, before) || template.Span != span || first.Span != span || !reflect.DeepEqual(bindings["Value"], argument) {
					t.Fatal("Number encoding changed input, prototype, argument, or source attribution")
				}
				second, err := template.Instantiate(bindings)
				if err != nil || !reflect.DeepEqual(first, second) {
					t.Fatalf("repeated Number instantiation changed: %#v/%v", second, err)
				}
				priorEmission := second.Bindings[0].Emitted
				first.Original[0] = '!'
				first.Bindings[0].Emitted = "changed"
				if !bytes.Equal(second.Original, before) || !bytes.Equal(template.Original, before) || second.Bindings[0].Emitted != priorEmission {
					t.Fatal("Number instances share mutable output or changed the prototype")
				}
			})
		}
	}
}

func TestDefinitionNumberErrors(t *testing.T) {
	for _, slot := range []string{"Number", "Thing"} {
		for _, tc := range []struct {
			input string
			cause error
		}{
			{"", strconv.ErrSyntax}, {"sideEffect()", strconv.ErrSyntax}, {"3_f64", strconv.ErrSyntax},
			{"1e999", strconv.ErrRange}, {"-1e999", strconv.ErrRange}, {"0x1p9999", strconv.ErrRange},
			{"NaN", nil}, {"nan", nil}, {"Inf", nil}, {"+Inf", nil}, {"-Inf", nil},
			{"Infinity", nil}, {"+Infinity", nil}, {"-Infinity", nil}, {"infinity", nil},
		} {
			t.Run(slot+"/"+tc.input, func(t *testing.T) {
				raw := []byte(`<DEF>BOD^E{|Statement: ~let prior: f64 = <@Prior|Number@>; let value: f64 = <@Value|` + slot + `@>;~|}<DEF>`)
				before := bytes.Clone(raw)
				span := SourceSpan{Path: "numbers.l", Start: SourcePosition{Line: 7, Column: 3}}
				template, err := ParseDefinitionTemplate(raw, span)
				if err != nil {
					t.Fatal(err)
				}
				argument := DefinitionArgument{Kind: "Number", Text: tc.input}
				bindings := map[string]DefinitionArgument{"Prior": {Kind: "Number", Text: "2.5"}, "Value": argument}
				instance, err := template.Instantiate(bindings)
				if err == nil || !errors.Is(err, lerrors.Syntax) || errors.Is(err, lerrors.Unsupported) || errors.Is(err, lerrors.InvalidInput) {
					t.Fatalf("invalid Number lost its Syntax category: %v", err)
				}
				if !strings.Contains(err.Error(), `numbers.l:7:3: definition: parameter "Value": invalid number `+strconv.Quote(tc.input)) {
					t.Fatalf("invalid Number lost source/slot/input context: %v", err)
				}
				var numeric *strconv.NumError
				if tc.cause != nil {
					if !errors.Is(err, tc.cause) || !errors.As(err, &numeric) || numeric.Num != tc.input || numeric.Func != "ParseFloat" {
						t.Fatalf("invalid Number lost native parse cause: %v", err)
					}
				} else if errors.As(err, &numeric) {
					t.Fatalf("nonfinite Number acquired an unrelated parse error: %v", err)
				}
				if !reflect.DeepEqual(instance, DefinitionInstance{}) || !bytes.Equal(raw, before) || !bytes.Equal(template.Original, before) || !reflect.DeepEqual(bindings["Value"], argument) || template.Span != span {
					t.Fatal("failed Number binding published a partial instance or changed input")
				}
			})
		}
	}
}

func TestDefinitionNumberPrimitiveParity(t *testing.T) {
	for _, tc := range definitionNumberCases {
		t.Run(tc.name, func(t *testing.T) {
			value := &primitive{Type: valueNumber, Number: math.Float64frombits(tc.bits)}
			literal, finite := nativeNumberLiteral(value.Number)
			if !finite || literal != tc.literal || math.Float64bits(value.Number) != tc.bits {
				t.Fatalf("primitive Number changed: %q/%v; want %q", literal, finite, tc.literal)
			}
			argument, err := argumentFromToken(&token{mem: &entry{Type: storagePrimitive, Primitive: value}})
			if err != nil || argument.Kind != "Number" || argument.Text != strings.TrimSuffix(tc.literal, "_f64") {
				t.Fatalf("semantic Number argument became a native expression: %#v/%v", argument, err)
			}
		})
	}
	for _, number := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		e := &orchestrationEmitter{}
		bound, err := e.literalValue(&primitive{Type: valueNumber, Number: number})
		if bound != (emittedExpression{}) || len(e.pending) != 0 || !errors.Is(err, lerrors.Unsupported) || errors.Is(err, lerrors.Syntax) || err.Error() != "non-finite Number is not supported" {
			t.Fatalf("nonfinite primitive Number changed its failure: %#v/%v/%v", bound, e.pending, err)
		}
	}
}

func TestDefinitionNumberNativeBits(t *testing.T) {
	var body []string
	var want strings.Builder
	for _, slot := range []string{"Number", "Thing"} {
		template := parseTestDefinition(t, `<DEF>BOD^E{|Statement: ~{ let value: f64 = <@Value|`+slot+`@>; ::std::println!("{:016x}", value.to_bits()); }~|}<DEF>`)
		for _, tc := range definitionNumberCases {
			instance := bindTestDefinition(t, template, map[string]DefinitionArgument{"Value": {Kind: "Number", Text: tc.input}})
			body = append(body, instance.Expression)
			fmt.Fprintf(&want, "%016x\n", tc.bits)
		}
	}
	// Opaque native integer expressions keep their own type and spelling.
	body = append(body, `BOD^E{|Statement: ~let native: i64 = 3_i64; assert_eq!(native, 3_i64);~|}`)
	document := []byte(`<PROGRAM><STRUCTURE><BOOK "Worker"><OUTLINE><JSON>{"AppMeta":{"AppGenDir":"/src"}}</JSON><JSON>{"Services":[{"Name":"main","Location":"/src/","GenericObjects":["main"]}]}</JSON><JSON>{"Objects":[{"Name":"main","Type":"root"}]}</JSON></OUTLINE><SETTINGS></SETTINGS><CONTENT><CHAPTER "main"><SENTENCE "main"><|W|><|S|>` + "`A`" + `</S|><|S|>` + "`FW{|Key: ~main~, Outputs: ~&[()]&~|}^W^(" + strings.Join(body, "*") + ")`" + `</S|></W|></SENTENCE></CHAPTER></CONTENT></BOOK></STRUCTURE></PROGRAM>`)
	before := bytes.Clone(document)
	bundle, err := Logiceme.NewLogicemeService(phase1Options(t).Resources).Evaluate(document, Logiceme.GenerationOptions{})
	recordBundle(t, bundle)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(document, before) {
		t.Fatal("Number generation changed backend input")
	}
	if got := runRawTextBundle(t, bundle); string(got) != want.String() {
		t.Fatalf("native Number bits = %q; want %q", got, want.String())
	}
}

func TestDefinitionNumberSourceAction(t *testing.T) {
	source := []byte("Define \"Check|-s|-ed|-ing|-er\" to `<DEF>BOD^E{|Statement: ~{ let value: f64 = <@Value|Number@>; ::std::println!(\"{:016x}\", value.to_bits()); }~|}<DEF>`.\n" +
		"Create \"Inspect|-s|-ed|-ing\" for Processing a New Checker.\n" +
		"Create \"Worker\" as a New Basic App.\n" +
		"The App Inspects 3.\nThe App Inspects -0.\nThe App Inspects 9007199254740993.\n")
	before := bytes.Clone(source)
	result, err := NewLygerService(phase1Options(t)).Compile(source, "number-action.l")
	if err != nil {
		t.Fatal(err)
	}
	units, err := result.ApplicationUnits()
	if err != nil || len(units) != 1 {
		t.Fatalf("Number action application units: %v", err)
	}
	bundle, err := Logiceme.NewLogicemeService(phase1Options(t).Resources).Evaluate(units[0].BackendAML, Logiceme.GenerationOptions{})
	recordBundle(t, bundle)
	if err != nil {
		t.Fatal(err)
	}
	if got := runRawTextBundle(t, bundle); string(got) != "4008000000000000\n8000000000000000\n4340000000000000\n" {
		t.Fatalf("L Number action lost binary64 values: %q", got)
	}
	if !bytes.Equal(source, before) {
		t.Fatal("Number action compilation changed source input")
	}
}
