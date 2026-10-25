package Lyger

import (
	"bytes"
	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	Logiceme "github.com/l-programming-language/l/logiceme"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func parseTestDefinition(t *testing.T, source string) *DefinitionTemplate {
	t.Helper()
	template, err := ParseDefinitionTemplate([]byte(source), SourceSpan{Path: "library.l", Start: SourcePosition{Line: 4, Column: 2}})
	if err != nil {
		t.Fatal(err)
	}
	return template
}

func bindTestDefinition(t *testing.T, template *DefinitionTemplate, bindings map[string]DefinitionArgument) DefinitionInstance {
	t.Helper()
	instance, err := template.Instantiate(bindings)
	if err != nil {
		t.Fatal(err)
	}
	return instance
}

func TestDefinitionGroupingAndIndependentUses(t *testing.T) {
	sequence := parseTestDefinition(t, `<DEF>W^<@Steps|Syllables@><DEF>`)
	steps := []string{`BOD^E{|Statement: ~first()~|}`, `BOD^E{|Statement: ~second()~|}`}
	bindings := map[string]DefinitionArgument{"Steps": {Kind: "Syllables", Syllables: steps}}
	instance := bindTestDefinition(t, sequence, bindings)
	want := "W^(" + steps[0] + "*" + steps[1] + ")"
	if instance.Expression != want || len(instance.Bindings) != 1 {
		t.Fatalf("ordered sequence = %#v", instance)
	}
	steps[0] = "A"
	if instance.Expression != want || instance.Bindings[0].Emitted != strings.TrimPrefix(want, "W^") {
		t.Fatal("argument mutation changed an existing instance")
	}
	group := parseTestDefinition(t, `<DEF>[["A", "<@Callable@>"]]<DEF>`)
	first := bindTestDefinition(t, group, map[string]DefinitionArgument{"Callable": {Kind: "Definition", Definition: &instance}})
	second := bindTestDefinition(t, group, map[string]DefinitionArgument{"Callable": {Kind: "Syllable", Text: "W"}})
	if !reflect.DeepEqual(first.Words, [][]string{{"A", want}}) || !reflect.DeepEqual(second.Words, [][]string{{"A", "W"}}) || first.Expression != "" {
		t.Fatalf("grouping was flattened: %#v / %#v", first, second)
	}
	first.Words[0][0] = "E"
	if second.Words[0][0] != "A" {
		t.Fatal("repeated grouping instances share mutable words")
	}
}

func TestDefinitionMalformedTemplates(t *testing.T) {
	for _, source := range []string{
		"", `<DEF>A`, `<DEF>A<DEF><DEF>`, `<DEF><@Value|Text<DEF>`,
		`<DEF><@Value|@><DEF>`, `<DEF><@Value|Text@>^<@Value|Number@><DEF>`, `<DEF>[["A"], 3]<DEF>`,
	} {
		if _, err := ParseDefinitionTemplate([]byte(source), SourceSpan{}); err == nil {
			t.Errorf("malformed template accepted: %q", source)
		}
	}
	// A currently unused member may be registered without silently rewriting
	// its body. Its malformed expression must fail when actually instantiated.
	unused := parseTestDefinition(t, `<DEF>BOD^E{|Statement: ~return <@Value|Text@>~|})<DEF>`)
	if _, err := unused.Instantiate(map[string]DefinitionArgument{"Value": {Kind: "Text", Text: "result"}}); err == nil {
		t.Fatal("unbalanced instantiated expression was accepted")
	}
}

func TestDefinitionTypedLiteralBinding(t *testing.T) {
	raw := "`" + `<DEF>BOD^E{|Statement: ~::std::println!("{}", <@Value|Thing@>);~|}<DEF>` + "`"
	template := parseTestDefinition(t, raw)
	if !reflect.DeepEqual(template.Slots, []DefinitionSlot{{Name: "Value", Type: "Thing"}}) {
		t.Fatalf("typed slots = %#v", template.Slots)
	}
	for _, text := range []string{"first", "quote \" and slash \\ and newline\n and <@Other@>"} {
		instance := bindTestDefinition(t, template, map[string]DefinitionArgument{"Value": {Kind: "Text", Text: text}})
		want := `BOD^E{|Statement: ~::std::println!("{}", ` + strconv.Quote(text) + `);~|}`
		if instance.Expression != want {
			t.Fatalf("expression = %q; want %q", instance.Expression, want)
		}
		instance.Original[0] = '!'
		if !bytes.Equal(template.Original, []byte(raw)) {
			t.Fatal("instance changed prototype source")
		}
	}
	for _, argument := range []DefinitionArgument{{Kind: "Number", Text: "-12.5"}, {Kind: "Truth", Text: "false"}, {Kind: "Null"}} {
		if _, err := template.Instantiate(map[string]DefinitionArgument{"Value": argument}); err != nil {
			t.Errorf("valid primitive %#v: %v", argument, err)
		}
	}
	for _, bindings := range []map[string]DefinitionArgument{
		nil,
		{"Value": {Kind: "Syllables", Syllables: []string{"A"}}},
		{"Value": {Kind: "Number", Text: "sideEffect()"}},
		{"Value": {Kind: "Truth", Text: "yes"}},
	} {
		if _, err := template.Instantiate(bindings); err == nil || !strings.Contains(err.Error(), "library.l:4:2") || !strings.Contains(err.Error(), "Value") {
			t.Errorf("invalid argument lost slot/source error: %v", err)
		}
	}
}

func TestDefinitionFunctionMetadata(t *testing.T) {
	template := parseTestDefinition(t, `<DEF>FW{|Func{Key: <@ID|Text@>, Inputs: <@Parameters|Props@>, Outputs: <@Results|Texts@>}Func|}^(<@Body@>)<DEF>`)
	if !template.IsFunction() || parseTestDefinition(t, `<DEF>W^<@Body@><DEF>`).IsFunction() {
		t.Fatal("function kind does not follow the root metadata structure")
	}
	bindings := map[string]DefinitionArgument{
		"ID":         {Kind: "Text", Text: "entry"},
		"Parameters": {Kind: "Props"},
		"Results":    {Kind: "Texts"},
		"Body":       {Kind: "Syllable", Text: "W"},
	}
	instance := bindTestDefinition(t, template, bindings)
	if instance.Expression != `FW{|Key: ~entry~|}^W` || !strings.Contains(instance.OriginalExpression, `Func{Key: "entry", Inputs: &[]&, Outputs: []}Func`) {
		t.Fatalf("function signature transformation = %#v", instance)
	}
	bindings["Parameters"] = DefinitionArgument{Kind: "Props", Props: []DefinitionProperty{{Key: "items", Type: "Vec<String>"}}}
	bindings["Results"] = DefinitionArgument{Kind: "Texts", Texts: []string{"Vec<String>", "runtime::Error"}}
	instance = bindTestDefinition(t, template, bindings)
	if !strings.Contains(instance.Expression, `Inputs: ~&[Prop{Key: items, Type: Vec<String>}Prop]&~`) || !strings.Contains(instance.Expression, `Outputs: ~&[(Vec<String>), (runtime::Error)]&~`) {
		t.Fatalf("nonempty signature lost types/cardinality: %s", instance.Expression)
	}
	for _, argument := range []DefinitionArgument{{Kind: "Text", Text: ""}, {Kind: "Text", Text: "not an identifier"}, {Kind: "Number", Text: "1"}} {
		bindings["ID"] = argument
		if _, err := template.Instantiate(bindings); err == nil {
			t.Errorf("invalid function name accepted: %#v", argument)
		}
	}
}

func TestDefinitionBackendComposition(t *testing.T) {
	writer := parseTestDefinition(t, `<DEF>BOD^E{|Statement: ~::std::println!("{}", <@Message|Thing@>);~|}<DEF>`)
	sequence := parseTestDefinition(t, `<DEF>W^<@Steps|Syllables@><DEF>`)
	callable := parseTestDefinition(t, `<DEF>FW{|Func{Key: <@ID|Text@>, Inputs: <@Parameters|Props@>, Outputs: <@Results|Texts@>}Func|}^(<@Body@>)<DEF>`)
	group := parseTestDefinition(t, `<DEF>[["A", "<@Callable@>"]]<DEF>`)
	messages := []string{"first line", "second \"quoted\" line"}
	var steps []string
	for _, message := range messages {
		steps = append(steps, bindTestDefinition(t, writer, map[string]DefinitionArgument{"Message": {Kind: "Text", Text: message}}).Expression)
	}
	body := bindTestDefinition(t, sequence, map[string]DefinitionArgument{"Steps": {Kind: "Syllables", Syllables: steps}})
	function := bindTestDefinition(t, callable, map[string]DefinitionArgument{"ID": {Kind: "Text", Text: "main"}, "Parameters": {Kind: "Props"}, "Results": {Kind: "Texts"}, "Body": {Kind: "Definition", Definition: &body}})
	objective := bindTestDefinition(t, group, map[string]DefinitionArgument{"Callable": {Kind: "Definition", Definition: &function}})
	returning := bindTestDefinition(t, callable, map[string]DefinitionArgument{"ID": {Kind: "Text", Text: "identity"}, "Parameters": {Kind: "Props", Props: []DefinitionProperty{{Key: "items", Type: "Vec<String>"}}}, "Results": {Kind: "Texts", Texts: []string{"Vec<String>", "Option<String>"}}, "Body": {Kind: "Syllable", Text: `W^BOD^E{|Statement: ~return (items, None);~|}`}})
	objective.Words[0] = append(objective.Words[0], returning.Expression)
	var contents strings.Builder
	for _, word := range objective.Words {
		contents.WriteString("<|W|>\n")
		for _, syllable := range word {
			contents.WriteString("<|S|>`" + syllable + "`</S|>\n")
		}
		contents.WriteString("</W|>\n")
	}
	document := `<PROGRAM><STRUCTURE><BOOK "Example"><OUTLINE><JSON>{"AppMeta":{"AppGenDir":"/src"}}</JSON><JSON>{"Services":[{"Name":"main","Location":"/src/","GenericObjects":["main"]}]}</JSON><JSON>{"Objects":[{"Name":"main","Type":"root"}]}</JSON></OUTLINE><SETTINGS></SETTINGS><CONTENT><CHAPTER "main"><SENTENCE "main">` + contents.String() + `</SENTENCE></CHAPTER></CONTENT></BOOK></STRUCTURE></PROGRAM>`
	bundle, err := Logiceme.NewLogicemeService(phase1Options(t).Resources).Evaluate([]byte(document), Logiceme.GenerationOptions{IncludeRuntime: false})
	recordBundle(t, bundle)
	if err != nil {
		t.Fatal(err)
	}
	file := observeBundleMain(t, bundle)
	main := file.function(t, "main")
	identity := file.function(t, "identity")
	if len(main.Parameters) != 0 {
		t.Fatal("main acquired inputs")
	}
	if nativetest.RustTokensText(identity.Parameters) != "items:Vec<String>" || nativetest.RustTokensText(identity.Result) != "->((Vec<String>),(Option<String>))" {
		t.Fatalf("identity lost typed signature: %s / %s", nativetest.RustTokensText(identity.Parameters), nativetest.RustTokensText(identity.Result))
	}
	if !reflect.DeepEqual(definitionTextMainCalls(t, bundle, "println"), messages) {
		t.Fatal("ordered native print payloads changed")
	}
	if !strings.Contains(nativetest.RustTokensText(identity.Body), "return(items,None);") {
		t.Fatal("identity return body changed")
	}
}
