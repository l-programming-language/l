package Lyger

import (
	"bytes"
	nativetest "github.com/l-programming-language/l/internal/testsupport/native"

	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	Logiceme "github.com/l-programming-language/l/logiceme"
)

// Explicit encodings are independent of the production quoting helper. Empty
// encoded values mark ordinary controls with an independently specified Rust quote.
var definitionTextEncodingCases = []struct{ name, text, encoded string }{
	{"empty", "", ""},
	{"ordinary", "ordinary text", ""},
	{"quotes escapes Unicode", "雪, quote \" slash \\ newline\n\t", ""},
	{"definition marker", "literal <DEF> tag", ""},
	{"slot marker", "<@Other@>", ""},
	{"interior backticks", "first``middle`last", ""},
	{"literal escape notation", `\x7c \x7e \x20`, ""},
	{"glyph operators", "A^B*C", ""},
	{"reserved characters", "|~", `"\x7c\x7e"`},
	{"metadata close", "|}", `"\x7c}"`},
	{"metadata pair", "{|value|}", `"{\x7cvalue\x7c}"`},
	{"metadata lookup", "[|Meta.Key|]", `"[\x7cMeta.Key\x7c]"`},
	{"recursive metadata lookup", "[|Meta.Statement|]", `"[\x7cMeta.Statement\x7c]"`},
	{"syllable close", "</S|>", `"</S\x7c>"`},
	{"word close", "</W|>", `"</W\x7c>"`},
	{"metadata item separator", "a~, b", `"a\x7e, b"`},
	{"metadata value separator", "a: ~b", `"a: \x7eb"`},
	{"two spaces", "a  b", `"a\x20\x20b"`},
	{"odd and even space runs", "  a   b    c  ", `"\x20\x20a\x20\x20 b\x20\x20\x20\x20c\x20\x20"`},
	{"five spaces", "a     b", `"a\x20\x20\x20\x20 b"`},
	{"overlapping separators", "a~,  b", `"a\x7e,\x20\x20b"`},
	{"comment delimiters", `/*|demo|*\ end`, `"/*\x7cdemo\x7c*\\ end"`},
	{"complete Lucid header", "(/={tags: tags}=={for Header in Headers}=={header}=/)", `"(\x2f={tags: tags}=={for Header in Headers}=={header}=/)"`},
	{"Lucid opener fragment", "(/foo)", `"(\x2ffoo)"`},
	{"array delimiters", "|[x]| |{x}|", `"\x7c[x]\x7c \x7c{x}\x7c"`},
}

func TestDefinitionTextEncodingBindings(t *testing.T) {
	for _, slot := range []string{"Text", "Thing"} {
		for _, tc := range definitionTextEncodingCases {
			t.Run(slot+"/"+tc.name, func(t *testing.T) {
				raw := `<DEF>BOD^E{|Statement: ~consume(<@Value|` + slot + `@>)~|}<DEF>`
				template := parseTestDefinition(t, raw)
				argument := DefinitionArgument{Kind: "tExT", Text: tc.text}
				bindings := map[string]DefinitionArgument{"Value": argument}
				instance := bindTestDefinition(t, template, bindings)
				encoded := tc.encoded
				if encoded == "" {
					encoded = strconv.Quote(tc.text)
				}
				want := `BOD^E{|Statement: ~consume(` + encoded + `)~|}`
				if instance.Expression != want || instance.OriginalExpression != want {
					t.Errorf("emitted definition = %q / %q; want %q", instance.Expression, instance.OriginalExpression, want)
				}
				wantLedger := []DefinitionBinding{{Slot: DefinitionSlot{Name: "Value", Type: slot}, Kind: "tExT", Emitted: encoded}}
				if !reflect.DeepEqual(instance.Bindings, wantLedger) {
					t.Errorf("emitted ledger = %#v; want %#v", instance.Bindings, wantLedger)
				}
				if len(instance.Bindings) != 1 {
					t.Fatal("instantiation lost its binding")
				}
				decoded, err := nativetest.DecodeRustLiteral(instance.Bindings[0].Emitted)
				if err != nil || decoded != tc.text {
					t.Fatalf("encoded Text does not round-trip: %q, %v", decoded, err)
				}
				if argument.Text != tc.text || !reflect.DeepEqual(bindings["Value"], argument) || !bytes.Equal(instance.Original, []byte(raw)) || !bytes.Equal(template.Original, []byte(raw)) || instance.Span != template.Span {
					t.Fatal("encoding changed caller data, raw source, or provenance")
				}
				again := bindTestDefinition(t, template, bindings)
				if !reflect.DeepEqual(instance, again) {
					t.Fatal("repeated instantiation changed output or attribution")
				}
				before := again.Expression
				priorEmission := again.Bindings[0].Emitted
				bindings["Value"] = DefinitionArgument{Kind: "Text", Text: "another value"}
				instance.Original[0] = '!'
				instance.Bindings[0].Emitted = "changed"
				if again.Expression != before || again.Bindings[0].Emitted != priorEmission || !bytes.Equal(again.Original, []byte(raw)) || !bytes.Equal(template.Original, []byte(raw)) {
					t.Fatal("instances, argument map, or prototype share mutable output")
				}
			})
		}
	}
}

// Inspect actual generated Rust only. The CLI integration adds one combined-payload
// native build rather than compiling a subprocess for each matrix entry.
func definitionTextMainCalls(t *testing.T, bundle Logiceme.Executable, function string) []string {
	t.Helper()
	file := observeBundleMain(t, bundle)
	method := observedMethodName(function)
	fn := file.function(t, "main")
	if method == "l_print" {
		return ioTextArguments(t, fn.Body)
	}
	var values []string
	for _, call := range nativetest.RustCalls(fn.Body) {
		if call.Name != method {
			continue
		}
		args := call.Args
		if method == "println" {
			if len(args) != 2 || nativetest.RustTokensText(args[0]) != `"{}"` {
				t.Fatal("println must retain fixed data format and one supplied value")
			}
			args = args[1:]
		}
		if len(args) != 1 {
			t.Fatalf("%s argument count=%d", method, len(args))
		}
		tokens := args[0]
		for len(tokens) > 0 && tokens[0].Text == "&" {
			tokens = tokens[1:]
		}
		if len(tokens) == 0 {
			t.Fatalf("%s lost native Text value", method)
		}
		value, err := nativetest.DecodeRustLiteral(tokens[0].Text)
		if err != nil {
			t.Fatal(err)
		}
		suffix := nativetest.RustTokensText(tokens[1:])
		switch suffix {
		case "", ".as_bytes()", ".as_bytes().to_vec()", ".to_owned()", ".to_string()", ".to_string().into_bytes()":
		default:
			t.Fatalf("unexpected Text conversion around native literal: %s", suffix)
		}

		values = append(values, value)
	}
	return values
}

// Canonical IO dispatches one borrowed native value. Check the actual selected
// receiver and evaluation count; inspecting its fallback println body alone
// would not establish which user payload is printed.
func ioTextArguments(t *testing.T, body []nativetest.RustToken) []string {
	t.Helper()
	var values []string
	dispatches := 0
	fallbacks := 0
	for _, call := range nativetest.RustCalls(body) {
		if call.At > 0 && body[call.At-1].Text == "fn" {
			continue // A local trait/impl declaration does not invoke its method.
		}
		if call.Name == "println" {
			fallbacks++
			if len(call.Args) != 2 || nativetest.RustTokensText(call.Args[0]) != `"{}"` || nativetest.RustTokensText(call.Args[1]) != "self.0" {
				t.Fatal("IO Print added a print outside its typed fallback")
			}
		}
		if call.Name == "l_print" {
			dispatches++
			if len(call.Args) != 0 {
				t.Fatal("IO Print dispatch must not add arguments")
			}
		}
		if call.Name != "LPrintValue" {
			continue
		}
		if len(call.Args) != 1 || len(call.Args[0]) != 4 {
			t.Fatal("IO Print must borrow exactly one parenthesized native Text literal")
		}
		arg := call.Args[0]
		if arg[0].Text != "&" || arg[1].Text != "(" || arg[3].Text != ")" {
			t.Fatal("IO Print changed the borrowed Text operand")
		}
		value, err := nativetest.DecodeRustLiteral(arg[2].Text)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(nativetest.RustTokensText(body), "(&&LPrintValue(&("+arg[2].Text+"))).l_print()") {
			t.Fatal("IO Print literal is not the selected dispatch receiver")
		}
		values = append(values, value)
	}
	if dispatches != len(values) || fallbacks != dispatches {
		t.Fatal("IO Print changed dispatch or payload evaluation count")
	}
	return values
}

func TestDefinitionTextEncodingGeneratedLiterals(t *testing.T) {
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	backend := Logiceme.NewLogicemeService(os.DirFS(root))
	for _, tc := range definitionTextEncodingCases {
		t.Run("definition/"+tc.name, func(t *testing.T) {
			defer func() {
				if recovered := recover(); recovered != nil {
					t.Errorf("Text generation panicked: %v", recovered)
				}
			}()
			template := parseTestDefinition(t, `<DEF>BOD^E{|Statement: ~::std::println!("{}", <@Value|Thing@>);~|}<DEF>`)
			instance := bindTestDefinition(t, template, map[string]DefinitionArgument{"Value": {Kind: "Text", Text: tc.text}})
			document := []byte(`<PROGRAM><STRUCTURE><BOOK "TextEncoding"><OUTLINE><JSON>{"AppMeta":{"AppGenDir":"/src"}}</JSON><JSON>{"Services":[{"Name":"main","Location":"/src/","GenericObjects":["main"]}]}</JSON><JSON>{"Objects":[{"Name":"main","Type":"root"}]}</JSON></OUTLINE><SETTINGS></SETTINGS><CONTENT><CHAPTER "main"><SENTENCE "main"><|W|><|S|>` + "`A`</S|><|S|>`FW{|Key: ~main~|}^W^" + instance.Expression + "`</S|></W|></SENTENCE></CHAPTER></CONTENT></BOOK></STRUCTURE></PROGRAM>")
			bundle, err := backend.Evaluate(document, Logiceme.GenerationOptions{IncludeRuntime: false})
			recordBundle(t, bundle)
			if err != nil {
				t.Fatal(err)
			}
			if actual := definitionTextMainCalls(t, bundle, "println"); !reflect.DeepEqual(actual, []string{tc.text}) {
				t.Fatalf("generated println Text = %q; want %q", actual, tc.text)
			}
		})
	}
}
