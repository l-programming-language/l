package Lyger

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/l-programming-language/l/lerrors"
	Logiceme "github.com/l-programming-language/l/logiceme"
)

func TestDefinitionLifetimeFraming(t *testing.T) {
	for name, expression := range map[string]string{
		"flat lifetime":       `FW{|Key: ~borrow~, Lifetimes: ~'a~, Inputs: ~&[Prop{Key: value, Type: &'a str}Prop]&~, Outputs: ~&[&'a str]&~|}^W^BOD^E{|Statement: ~return value;~|}`,
		"lifetimes and chars": `E{|Statement: ~('a', '雪', '\\', '\'', '\u{7c}', value, &'a str)~|}`,
		"raw string":          `E{|Statement: ~r##"inner " and |} and 'a and ,"##~|}`,
		"raw byte string":     `E{|Statement: ~br#"|}, 'a"#~|}`,
	} {
		t.Run(name, func(t *testing.T) {
			raw := []byte("<DEF>" + expression + "<DEF>")
			template, err := ParseDefinitionTemplate(raw, SourceSpan{})
			if err != nil {
				t.Fatal(err)
			}
			for count := 0; count < 2; count++ {
				instance, err := template.Instantiate(nil)
				if err != nil {
					t.Fatal(err)
				}
				// Source framing decodes only its documented quote/backslash escapes.
				want := decodeDefinitionEscapes(expression)
				if instance.Expression != want {
					t.Fatalf("native bytes changed: got %q, want %q", instance.Expression, want)
				}
				if !bytes.Equal(template.Original, raw) {
					t.Fatal("template source mutated")
				}
			}
		})
	}
	for _, literal := range []string{`r##"inner " and |} and 'a"##`, `br#"|}, 'a"#`, `'|'`, `'\''`, `'\u{7c}'`, `"|}"`} {
		source := "E{|Statement: ~" + literal + "~|}"
		end, err := definitionMetadataEnd(source, 1)
		if err != nil || end != len(source)-2 {
			t.Errorf("native literal %q ended metadata early: %d, %v", literal, end, err)
		}
	}
	fields := `Key: borrow, Inputs: &[Prop{Key: value, Type: &'a (u8, u16)}Prop]&, Outputs: ["&'a (u8, u16)"]`
	got, err := splitDefinitionFields(fields)
	want := []string{`Key: borrow`, ` Inputs: &[Prop{Key: value, Type: &'a (u8, u16)}Prop]&`, ` Outputs: ["&'a (u8, u16)"]`}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("lifetime field framing: %#v, %v", got, err)
	}
	for _, malformed := range []string{`E{|Statement: ~r##"unterminated"#~|}`, `E{|Statement: ~"unterminated~|}`} {
		if _, err := normalizeDefinitionFunctions(malformed); err == nil || !errors.Is(err, lerrors.Syntax) {
			t.Errorf("missing classified unterminated literal: %q: %v", malformed, err)
		}
	}
}

func TestDefinitionNativeIdentifierBoundary(t *testing.T) {
	function := parseTestDefinition(t, `<DEF>FW{|Func{Key: <@Name|Text@>, Inputs: <@Inputs|Props@>, Outputs: []}Func|}^W<DEF>`)
	for _, name := range []string{"map", "range", "r#type", "r#fn", "r#match", "δ", "a_1"} {
		instance, err := function.Instantiate(map[string]DefinitionArgument{"Name": {Kind: "Text", Text: name}, "Inputs": {Kind: "Props", Props: []DefinitionProperty{{Key: name, Type: "u8"}}}})
		if err != nil {
			t.Errorf("valid Rust native name %q rejected: %v", name, err)
			continue
		}
		if !strings.Contains(instance.Expression, "Key: ~"+name+"~") || !strings.Contains(instance.Expression, "Key: "+name+",") {
			t.Errorf("native name changed: %s", instance.Expression)
		}
	}
	for _, name := range []string{"fn", "match", "async", "await", "gen", "r#self", "r#Self", "r#super", "r#crate", "r#_", "_", "1a", "a-b"} {
		if _, err := function.Instantiate(map[string]DefinitionArgument{"Name": {Kind: "Text", Text: name}, "Inputs": {Kind: "Props"}}); err == nil {
			t.Errorf("invalid native function name accepted: %q", name)
		}
	}
	if _, err := function.Instantiate(map[string]DefinitionArgument{"Name": {Kind: "Text", Text: "discard"}, "Inputs": {Kind: "Props", Props: []DefinitionProperty{{Key: "_", Type: "u8"}}}}); err != nil {
		t.Errorf("explicit wildcard parameter rejected: %v", err)
	}
	if _, err := ParseDefinitionTemplate([]byte(`<DEF>E{|Statement: ~<@r#type|Text@>~|}<DEF>`), SourceSpan{}); err == nil {
		t.Fatal("native identifier rules leaked into L slot names")
	}
}

func TestDefinitionNullIsTypedAtom(t *testing.T) {
	const want = `()`
	for _, kind := range []string{"Thing", "Null"} {
		template := parseTestDefinition(t, `<DEF>E{|Statement: ~<@Value|`+kind+`@>~|}<DEF>`)
		instance, err := template.Instantiate(map[string]DefinitionArgument{"Value": {Kind: "Null"}})
		if err != nil {
			t.Fatal(err)
		}
		if instance.Expression != `E{|Statement: ~`+want+`~|}` {
			t.Fatalf("Null atom: %s", instance.Expression)
		}
	}
	for _, kind := range []string{"Text", "Number", "Truth", "Props", "Texts", "Syllables"} {
		template := parseTestDefinition(t, `<DEF>E{|Statement: ~<@Value|`+kind+`@>~|}<DEF>`)
		if _, err := template.Instantiate(map[string]DefinitionArgument{"Value": {Kind: "Null"}}); err == nil || !errors.Is(err, lerrors.TypeMismatch) {
			t.Errorf("Null bypassed %s slot: %v", kind, err)
		}
	}
}

func TestDefinitionNullCompleteApplication(t *testing.T) {
	c := requireDistributionFrontend(t, "Create \"Child\" as a New Basic App.\n", nil).Compilation

	formatted, err := c.instantiateBuiltin("Null.Format", map[string]DefinitionArgument{"Name": definitionReferenceArgument("nullDisplay")})
	if err != nil {
		t.Fatal(err)
	}
	writer := parseTestDefinition(t, `<DEF>BOD^E{|Statement: ~::std::println!("{}", <@Value|Thing@>);~|}<DEF>`)
	value := definitionReferenceArgument("nullDisplay")
	value.Kind = "Null"
	instance, err := writer.Instantiate(map[string]DefinitionArgument{"Value": value})
	if err != nil {
		t.Fatal(err)
	}
	body, err := c.instantiateBuiltin("Process", map[string]DefinitionArgument{"Definitions": {Kind: "Syllables", Syllables: []string{formatted.Expression, instance.Expression}}})
	if err != nil {
		t.Fatal(err)
	}
	function, err := c.instantiateBuiltin("Function", map[string]DefinitionArgument{
		"Name": {Kind: "Text", Text: "main"}, "Inputs": {Kind: "Props"}, "Outputs": {Kind: "Texts"},
		"Process": {Kind: "Definition", Definition: &body},
	})
	if err != nil {
		t.Fatal(err)
	}
	document := `<PROGRAM><STRUCTURE><BOOK "Worker"><OUTLINE><JSON>{"AppMeta":{"AppGenDir":"/src"}}</JSON><JSON>{"Services":[{"Name":"main","Location":"/src/","GenericObjects":["main"]}]}</JSON><JSON>{"Objects":[{"Name":"main","Type":"root"}]}</JSON></OUTLINE><SETTINGS></SETTINGS><CONTENT><CHAPTER "main"><SENTENCE "main"><|W|><|S|>` + "`A`" + `</S|><|S|>` + "`" + function.Expression + "`" + `</S|></W|></SENTENCE></CHAPTER></CONTENT></BOOK></STRUCTURE></PROGRAM>`
	bundle, err := Logiceme.NewLogicemeService(phase1Options(t).Resources).Evaluate([]byte(document), Logiceme.GenerationOptions{})
	recordBundle(t, bundle)
	if err != nil {
		t.Fatal(err)
	}
	binary := buildActualBundle(t, bundle, "Worker", false, "")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("Null application: %v: %s", err, stderr.Bytes())
	}
	if stdout.String() != "<nil>\n" || stderr.Len() != 0 {
		t.Fatalf("Null output: stdout=%q stderr=%q", stdout.Bytes(), stderr.Bytes())
	}
}

func TestDefinitionLifetimeCompleteApplication(t *testing.T) {
	flat := parseTestDefinition(t, `<DEF>FW{|Key: ~borrow~, Lifetimes: ~'a~, Inputs: ~&[Prop{Key: value, Type: &'a str}Prop]&~, Outputs: ~&[&'a str]&~|}^W^BOD^E{|Statement: ~return value;~|}<DEF>`)
	borrow := bindTestDefinition(t, flat, nil)
	record := parseTestDefinition(t, `<DEF>FW{|Func{Key: <@Name|Text@>, Inputs: <@Inputs|Props@>, Outputs: <@Outputs|Texts@>}Func|}^W^BOD^E{|Statement: ~return r#match;~|}<DEF>`)
	raw := bindTestDefinition(t, record, map[string]DefinitionArgument{
		"Name":    {Kind: "Text", Text: "r#type"},
		"Inputs":  {Kind: "Props", Props: []DefinitionProperty{{Key: "r#match", Type: "&'static str"}}},
		"Outputs": {Kind: "Texts", Texts: []string{"&'static str"}},
	})
	main := `FW{|Key: ~main~, Outputs: ~&[()]&~|}^W^BOD^E{|Statement: ~let value = String::from("borrowed"); assert_eq!(borrow(&value), "borrowed"); assert_eq!(r#type("raw"), "raw");~|}`
	var words strings.Builder
	for _, expression := range []string{"A", borrow.Expression, raw.Expression, main} {
		words.WriteString("<|W|><|S|>`" + expression + "`</S|></W|>")
	}
	document := `<PROGRAM><STRUCTURE><BOOK "Worker"><OUTLINE><JSON>{"AppMeta":{"AppGenDir":"/src"}}</JSON><JSON>{"Services":[{"Name":"main","Location":"/src/","GenericObjects":["main"]}]}</JSON><JSON>{"Objects":[{"Name":"main","Type":"root"}]}</JSON></OUTLINE><SETTINGS></SETTINGS><CONTENT><CHAPTER "main"><SENTENCE "main">` + words.String() + `</SENTENCE></CHAPTER></CONTENT></BOOK></STRUCTURE></PROGRAM>`
	bundle, err := Logiceme.NewLogicemeService(phase1Options(t).Resources).Evaluate([]byte(document), Logiceme.GenerationOptions{})
	recordBundle(t, bundle)
	if err != nil {
		t.Fatal(err)
	}
	binary := buildActualBundle(t, bundle, "Worker", false, "")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary)
	if output, err := command.CombinedOutput(); err != nil || len(output) != 0 {
		t.Fatalf("native lifetime/raw identifier program: %v: %s", err, output)
	}
}
