package Lyger

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"reflect"
	"strings"
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

func sourceLogTemplate(t *testing.T) *DefinitionTemplate {
	t.Helper()
	source, err := os.ReadFile("../stdlib/eng/io/dictionary.l")
	if err != nil {
		t.Fatal(err)
	}
	first, last := bytes.IndexByte(source, '`'), bytes.LastIndexByte(source, '`')
	if first == -1 || last <= first {
		t.Fatal("source-defined Log is missing its native definition")
	}
	template, err := ParseDefinitionTemplate(source[first:last+1], SourceSpan{Path: "stdlib/eng/io/dictionary.l"})
	if err != nil {
		t.Fatal(err)
	}
	return template
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
	before := bytes.Clone(source)
	result, err := NewLygerService(phase1Options(t)).Compile(source, "raw-text.l")
	if err != nil {
		t.Fatal(err)
	}
	units, err := result.ApplicationUnits()
	if err != nil || len(units) != 1 {
		t.Fatalf("raw Text application units: %v", err)
	}
	bundle, err := Logiceme.NewLogicemeService(phase1Options(t).Resources).Evaluate(units[0].BackendAML, Logiceme.GenerationOptions{})
	recordBundle(t, bundle)
	if err != nil {
		t.Fatal(err)
	}
	want := append(bytes.Clone(raw), []byte("\n1e+06\n")...)
	if got := runRawTextBundle(t, bundle); !bytes.Equal(got, want) {
		t.Fatalf("source-defined IO changed scalar bytes: got %x; want %x", got, want)
	}
	if !bytes.Equal(source, before) {
		t.Fatal("source-defined IO changed raw input bytes")
	}
}

func TestSourceLogTypedDispatch(t *testing.T) {
	logger := sourceLogTemplate(t)
	var body []string
	for _, argument := range []DefinitionArgument{
		{Kind: "Text", Text: string([]byte{'a', 0xff, 'b'})},
		{Kind: "Text", Text: "雪|~"},
		{Kind: "Number", Text: "1e6"},
		{Kind: "Truth", Text: "false"},
		{Kind: "Null"},
	} {
		instance, err := logger.Instantiate(map[string]DefinitionArgument{"Output": argument})
		if err != nil {
			t.Fatal(err)
		}
		body = append(body, instance.Expression)
	}
	// Instantiate the real source-defined native block around a native expression
	// to observe exactly-once evaluation and borrowing independently of L's scalar
	// argument encoder. No generated source is modified after Evaluate.
	prototype, err := os.ReadFile("../stdlib/eng/io/dictionary.l")
	if err != nil {
		t.Fatal(err)
	}
	first, last := bytes.IndexByte(prototype, '`'), bytes.LastIndexByte(prototype, '`')
	for _, expression := range []string{`owner`, `{ count += 1; String::from("once") }`} {
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
	body = append([]string{`BOD^E{|Statement: ~let owner = String::from("retained"); let mut count = 0;~|}`}, body...)
	body = append(body, `BOD^E{|Statement: ~assert_eq!(owner, "retained"); assert_eq!(count, 1);~|}`)
	document := `<PROGRAM><STRUCTURE><BOOK "Worker"><OUTLINE><JSON>{"AppMeta":{"AppGenDir":"/src"}}</JSON><JSON>{"Services":[{"Name":"main","Location":"/src/","GenericObjects":["main"]}]}</JSON><JSON>{"Objects":[{"Name":"main","Type":"root"}]}</JSON></OUTLINE><SETTINGS></SETTINGS><CONTENT><CHAPTER "main"><SENTENCE "main"><|W|><|S|>` + "`A`" + `</S|><|S|>` + "`FW{|Key: ~main~, Outputs: ~&[()]&~|}^W^(" + strings.Join(body, "*") + ")`" + `</S|></W|></SENTENCE></CHAPTER></CONTENT></BOOK></STRUCTURE></PROGRAM>`
	bundle, err := Logiceme.NewLogicemeService(phase1Options(t).Resources).Evaluate([]byte(document), Logiceme.GenerationOptions{})
	recordBundle(t, bundle)
	if err != nil {
		t.Fatal(err)
	}
	want := append([]byte{'a', 0xff, 'b'}, []byte("\n雪|~\n1e+06\nfalse\n<nil>\nretained\nonce\n")...)
	if got := runRawTextBundle(t, bundle); !bytes.Equal(got, want) {
		t.Fatalf("typed native IO changed scalar bytes: got %x; want %x", got, want)
	}
}
