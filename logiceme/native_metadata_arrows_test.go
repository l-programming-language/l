package Logiceme

import (
	"bytes"
	"reflect"
	"testing"

	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
)

func TestNativeMetadataArrowsRecords(t *testing.T) {
	for _, nativeType := range []string{
		"Result<fn() -> u8, ()>",
		"Vec<Result<fn()->u8, ()>>",
		"Result<Box<dyn Fn(u8) -> u8>, ()>",
		"fn(u8, u16) -> Result<(u8, u16), ()>",
		"Result<for<'a> fn(&'a u8) -> &'a u8, ()>",
	} {
		t.Run(nativeType, func(t *testing.T) {
			inputs := "&[Prop{Key: callback, Type: " + nativeType + "}Prop, Prop{Key: following, Type: bool}Prop]&"
			wantInputs := []Prop{{Key: "callback", Type: nativeType}, {Key: "following", Type: "bool"}}
			got, err := readNativeProps(inputs)
			if err != nil || !reflect.DeepEqual(got, wantInputs) {
				t.Fatalf("Props = %#v, %v; want %#v", got, err, wantInputs)
			}
			record := "Func{Key: apply, Inputs: " + inputs + ", Outputs: &[" + nativeType + ", bool]&}Func"
			want := Func{Key: "apply", Inputs: wantInputs, Outputs: []string{nativeType, "bool"}}
			function, err := readNativeFunc(map[string]string{"Func": record})
			if err != nil || !reflect.DeepEqual(function, want) {
				t.Fatalf("Func = %#v, %v; want %#v", function, err, want)
			}
			functions, err := readNativeFuncs("&[" + record + ", Func{Key: following}Func]&")
			if err != nil || !reflect.DeepEqual(functions, []Func{want, {Key: "following"}}) {
				t.Fatalf("Funcs = %#v, %v", functions, err)
			}
			lucidFuncRepeatedInterpretation(t, "Rust", record, "",
				"[|Meta.Func.Key|]([|Meta.Func.Inputs~|]) -> [|(Meta.Func.Outputs~)|]",
				"apply(callback: "+nativeType+", following: bool) -> ("+nativeType+", bool)")
		})
	}
}

func TestNativeMetadataArrowsOperands(t *testing.T) {
	for _, operand := range []string{
		"make::<Result<fn() -> u8, ()>, u8>()",
		"std::mem::size_of::<Vec<Result<fn()->u8, ()>>>()",
		"\"text -> , }Op\"",
		"r#\"text -> , }Op\"#",
		"'>'",
	} {
		t.Run(operand, func(t *testing.T) {
			raw := "Op{Type: Calculate, Operands: &[" + operand + ", following]&, Prefix: (, Postfix: ), Operator: +}Op"
			got, err := readNativeOps(raw)
			want := []Op{{Type: "Calculate", Operands: []string{operand, "following"}, Prefix: "(", Postfix: ")", Operator: "+"}}
			if err != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("Ops = %#v, %v; want %#v", got, err, want)
			}
		})
	}
}

func TestNativeMetadataArrowsFieldBoundary(t *testing.T) {
	// Test this scanner independently of readNativeList. The inner associated
	// type bound has field-like punctuation but is still inside the native type.
	native := "impl Callback<fn() -> u8, Output: Copy>"
	input := native + ", Operator: +}Op"
	if got := nativeOpFieldEnd(input, true); got != len(native) {
		t.Fatalf("native field in %q ended at %d; want %d", input, got, len(native))
	}
}

// This complete composition and its source expectation are authored here, not
// generated from the implementation or added to the frozen Lucid corpus. The
// existing plain package identity lets both compile phases use its retained lock.
const nativeMetadataArrowsAML = `<PROGRAM><STRUCTURE><BOOK "lucid_nested_rust"><OUTLINE>
<JSON>{"AppMeta":{"AppGenDir":"/src"}}</JSON>
<JSON>{"Services":[{"Name":"main","Location":"/src/","GenericObjects":["main"]}]}</JSON>
<JSON>{"Objects":[{"Name":"main","Type":"root"}]}</JSON>
</OUTLINE><SETTINGS></SETTINGS><CONTENT><CHAPTER "main"><SENTENCE "main"><|W|>` + "\n" +
	"<|S|>`A`</S|>\n" +
	"<|S|>`FW{|Key: ~flat~, Inputs: ~&[Prop{Key: callback, Type: Result<fn() -> u8, ()>}Prop, Prop{Key: following, Type: bool}Prop]&~, Outputs: ~&[Result<fn() -> u8, ()>]&~|}^W^BOD^E{|Statement: ~let _ = following; return callback;~|}`</S|>\n" +
	"<|S|>`FW{|Func: ~Func{Key: record, Inputs: &[Prop{Key: callback, Type: Vec<Result<fn()->u8, ()>>}Prop]&, Outputs: &[Vec<Result<fn()->u8, ()>>]&}Func~|}^W^BOD^E{|Statement: ~return callback;~|}`</S|>\n" +
	"<|S|>`MY{|Key: ~Callback~, Funcs: ~&[Func{Key: apply, Inputs: &[Prop{Key: callback, Type: Result<Box<dyn Fn(u8) -> u8>, ()>}Prop]&, Outputs: &[Result<fn(u8, u16) -> Result<(u8, u16), ()>, ()>]&}Func, Func{Key: borrowed, Inputs: &[Prop{Key: callback, Type: Result<for<'a> fn(&'a u8) -> &'a u8, ()>}Prop]&}Func]&~|}`</S|>\n" +
	"<|S|>`FW{|Key: ~main~, Outputs: ~&[()]&~|}^W^(BOD^DE{|Op: ~Op{Type: Calculate, Prefix: let _ =, Operator: >=, Operands: &[std::mem::size_of::<Result<fn() -> u8, ()>>(), 0]&, Postfix: ;}Op~|}*BOD^E{|Statement: ~let _: Result<fn() -> u8, ()> = flat(Err(()), true); let _: Vec<Result<fn()->u8, ()>> = record(Vec::new());~|})`</S|>\n" +
	"</W|></SENTENCE></CHAPTER></CONTENT></BOOK></STRUCTURE></PROGRAM>\n"

// Literal spacing comes from the frozen A/FW/W/MY/DE templates, independently
// of the generator under test. Native type interiors retain their exact bytes.
const nativeMetadataArrowsSource = "\n\n// package main\n\n\n\n\n" +
	"fn flat( callback: Result<fn() -> u8, ()>, following: bool) -> (Result<fn() -> u8, ()>) {\n    \n        \n\nlet _ = following; return callback;\n\n\n\n    \n}\n" +
	"fn record( callback: Vec<Result<fn()->u8, ()>>) -> (Vec<Result<fn()->u8, ()>>) {\n    \n        \n\nreturn callback;\n\n\n\n    \n}\n" +
	"\n\n\n\ntrait Callback {\n\n\nfn apply(callback: Result<Box<dyn Fn(u8) -> u8>, ()>) -> (Result<fn(u8, u16) -> Result<(u8, u16), ()>, ()>);\n\n\n\n\n\nfn borrowed(callback: Result<for<'a> fn(&'a u8) -> &'a u8, ()>);\n\n\n}\n\n\n\n" +
	"fn main( ) -> (()) {\n    \n        \n\n\n let _ = (std::mem::size_of::<Result<fn() -> u8, ()>>() >= 0) ; \n\n\n\n\n\nlet _: Result<fn() -> u8, ()> = flat(Err(()), true); let _: Vec<Result<fn()->u8, ()>> = record(Vec::new());\n\n\n\n    \n}\n"

func nativeMetadataArrowsExpected(t *testing.T) Executable {
	t.Helper()
	return Executable{
		MapArrays: [][]string{{"/src/main.rs", "/Cargo.toml", "/src/config.rs", "/start.sh", "/stop.sh"}},
		ByteArrays: [][][]byte{{
			[]byte(nativeMetadataArrowsSource), lucidRead(t, "common/Cargo.toml.txt"),
			lucidRead(t, "common/config.rs.txt"), lucidRead(t, "common/start.sh.txt"), lucidRead(t, "common/stop.sh.txt"),
		}},
		DependencyArrays: [][]string{nil}, EnvDepArrays: [][]string{nil},
	}
}

func nativeMetadataArrowsGenerate(t *testing.T, service *LogicemeService, source string) Executable {
	t.Helper()
	input := []byte(source)
	before := bytes.Clone(input)
	bundle, err := service.Evaluate(input)
	nativetest.RecordBundle(t, "generation", bundle)
	if !bytes.Equal(input, before) {
		t.Fatal("native arrow generation mutated its AML")
	}
	if err != nil {
		t.Fatalf("native arrow generation: %v", err)
	}
	return bundle
}

func nativeMetadataArrowsCompare(t *testing.T, got, want Executable) {
	t.Helper()
	if !reflect.DeepEqual(got.MapArrays, want.MapArrays) ||
		!reflect.DeepEqual(got.DependencyArrays, want.DependencyArrays) ||
		!reflect.DeepEqual(got.EnvDepArrays, want.EnvDepArrays) || got.Name != want.Name {
		t.Fatalf("bundle metadata differs: got=%#v want=%#v", got, want)
	}
	if len(got.ByteArrays) != 1 || len(got.ByteArrays[0]) != len(want.ByteArrays[0]) {
		t.Fatalf("bundle content dimensions differ: %#v", got.ByteArrays)
	}
	for i, expected := range want.ByteArrays[0] {
		if !bytes.Equal(got.ByteArrays[0][i], expected) {
			t.Errorf("%s differs:\ngot  %q\nwant %q", want.MapArrays[0][i], got.ByteArrays[0][i], expected)
		}
	}
}

func TestNativeMetadataArrowsGeneration(t *testing.T) {
	for _, embedded := range []bool{false, true} {
		name := "filesystem"
		if embedded {
			name = "embedded"
		}
		t.Run(name, func(t *testing.T) {
			config := runtimeGenerationResources(t)
			if embedded {
				config = nil
			}
			service := NewLogicemeService(config)
			want := nativeMetadataArrowsExpected(t)
			first := nativeMetadataArrowsGenerate(t, service, nativeMetadataArrowsAML)
			nativeMetadataArrowsCompare(t, first, want)
			control := nativeMetadataArrowsGenerate(t, service, string(lucidRead(t, "statements.aml.txt")))
			nativeMetadataArrowsCompare(t, control, lucidExpectedBundle(t, "statements"))
			last := nativeMetadataArrowsGenerate(t, service, nativeMetadataArrowsAML)
			nativeMetadataArrowsCompare(t, last, want)
			nativeMetadataArrowsCompare(t, first, want)
		})
	}
}
