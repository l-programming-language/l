package Lyger

import (
	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	Logiceme "github.com/l-programming-language/l/logiceme"
	"regexp"
	"strings"
	"testing"
)

func TestEmptyConstructionRetainsRequiredArguments(t *testing.T) {
	for _, tc := range []struct {
		name, definition string
		empty            bool
	}{
		{"Sequence", "W^<@Items|Syllables@>", true},
		{"Scalar", "W^<@Item|Syllable@>", false},
		{"RequiredText", "BOD^E{|Statement: ~<@Value|Text@>~|}", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := requireDistributionFrontend(t, "Define \""+tc.name+"\" as `<DEF>"+tc.definition+"<DEF>`.\n", nil)
			c := result.Compilation
			rec := c.Memory[tc.name]
			if rec == nil {
				t.Fatal("source-defined constructor argument was not registered")
			}
			node := &constructionNode{Nominal: "holder", Parameters: []*objectGram{{noun: &token{binding: &tokenBinding{Record: rec}}}}}
			instance, err := c.emptyConstructionSequence(node)
			if err != nil || (instance != nil) != tc.empty {
				t.Fatalf("empty argument = %+v, %v; want available %v", instance, err, tc.empty)
			}
			if err := c.validateConstructionParameters(node, nil); err == nil {
				t.Fatal("nil body silently acquired a required argument")
			}
			if instance != nil {
				if instance.definition != rec || instance.Expression != "W" || len(instance.Bindings) != 1 || instance.Bindings[0].Slot.Name != "Items" {
					t.Fatalf("empty source sequence lost its canonical constructor: %+v", instance)
				}
				if err := c.validateConstructionParameters(node, []DefinitionInstance{*instance}); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestEmptyApplicationUsesSourceSequence(t *testing.T) {
	for _, source := range []string{"Create \"Empty\" as a New Basic App.\nExecute Empty.\n", "Use Basic as \"Plain\".\nCreate \"Empty\" as a New Plain App.\nExecute Empty.\n"} {
		t.Run(source, func(t *testing.T) {
			result := requireDistributionFrontend(t, source, nil)
			construction := result.Compilation.Semantics.Constructions[0]
			if len(construction.Body) != 1 || construction.Body[0].definition == nil || len(construction.Body[0].Bindings) != 1 || construction.Body[0].Bindings[0].Emitted != "" {
				t.Fatalf("empty App lost explicit empty sequence: %+v", construction.Body)
			}
			units, err := result.ApplicationUnits()
			if err != nil || len(units) != 1 {
				t.Fatalf("empty App units: %v", err)
			}
			bundle, err := Logiceme.NewLogicemeService(phase1Options(t).Resources).Evaluate(units[0].BackendAML, Logiceme.GenerationOptions{IncludeRuntime: false})
			recordBundle(t, bundle)
			if err != nil {
				t.Fatal(err)
			}
			main := observeBundleMain(t, bundle).function(t, "main")
			calls := nativetest.RustCalls(main.Body)
			// Rust may handle the configuration Result explicitly before any source body.
			if len(main.Body) > 0 && main.Body[0].Text == "if" {
				if len(calls) != 4 || calls[0].Name != "Err" || calls[1].Name != "initialize_configuration_settings" || calls[2].Name != "eprintln" || calls[3].Name != "exit" {
					t.Fatalf("empty App acquired statements beyond configuration failure handling: %s", nativetest.RustTokensText(main.Body))
				}
				pattern, config, diagnostic, exit := calls[0], calls[1], calls[2], calls[3]
				if len(pattern.Args) != 1 || len(pattern.Args[0]) != 1 || len(config.Args) != 0 || len(diagnostic.Args) != 1 || len(diagnostic.Args[0]) != 1 || len(exit.Args) != 1 || nativetest.RustTokensText(exit.Args[0]) != "1" {
					t.Fatal("configuration failure prologue changed its arguments")
				}
				name := pattern.Args[0][0].Text
				message, err := nativetest.DecodeRustLiteral(diagnostic.Args[0][0].Text)
				if err != nil || !strings.Contains(message, "{"+name+"}") {
					t.Fatal("configuration failure diagnostic no longer reports its native error")
				}
				want := "ifletErr(" + name + ")=config::initialize_configuration_settings(){::std::eprintln!(" + diagnostic.Args[0][0].Text + ");::std::process::exit(1);}"
				if nativetest.RustTokensText(main.Body) != want {
					t.Fatalf("empty App acquired source-body statements around configuration handling: %s", nativetest.RustTokensText(main.Body))
				}
				return
			}
			initializations := 0
			var prologue nativetest.RustCall
			for _, call := range calls {
				switch call.Name {
				case "initialize_configuration_settings":
					initializations++
					prologue = call
					if len(call.Args) != 0 {
						t.Fatal("configuration prologue acquired arguments")
					}
				case "unwrap", "expect": // Explicit handling of the prologue's native Result.
				default:
					t.Fatalf("empty App acquired invented source call %s", call.Name)
				}
			}
			if initializations != 1 {
				t.Fatalf("empty App has %d configuration prologues, want exactly one", initializations)
			}
			// The sole statement must be the configuration call, optionally
			// qualified and with explicit handling of its native Result.
			prefix := nativetest.RustTokensText(main.Body[:prologue.At])
			if !regexp.MustCompile(`^(::)?([A-Za-z_][A-Za-z_0-9]*::)*$`).MatchString(prefix) {
				t.Fatalf("empty App acquired statements before its prologue: %s", prefix)
			}
			suffix := nativetest.RustTokensText(main.Body[prologue.End:])
			switch suffix {
			case ";", "?;", ".unwrap();":
			default:
				if len(calls) != 2 || calls[1].Name != "expect" || len(calls[1].Args) != 1 || len(calls[1].Args[0]) != 1 || calls[1].At != prologue.End+1 || calls[1].End != len(main.Body)-1 || main.Body[prologue.End].Text != "." || main.Body[len(main.Body)-1].Text != ";" {
					t.Fatalf("empty App acquired source-body statements: %s", suffix)
				}
				if _, err := nativetest.DecodeRustLiteral(calls[1].Args[0][0].Text); err != nil {
					t.Fatalf("configuration expect diagnostic is not a literal: %v", err)
				}
			}
		})
	}
}

func TestEmptyConstructionUsesFirstParameter(t *testing.T) {
	result := requireDistributionFrontend(t, "Define \"Sequence\" as `<DEF>W^<@Items|Syllables@><DEF>`.\nDefine \"Scalar\" as `<DEF>W^<@Item|Syllable@><DEF>`.\n", nil)
	c := result.Compilation
	parameter := func(name string) *objectGram {
		rec := c.Memory[name]
		if rec == nil {
			t.Fatalf("source-defined %s argument was not registered", name)
		}
		return &objectGram{noun: &token{binding: &tokenBinding{Record: rec}}}
	}
	for _, tc := range []struct {
		name       string
		parameters []*objectGram
		empty      bool
	}{
		{"scalar before sequence", []*objectGram{parameter("Scalar"), parameter("Sequence")}, false},
		{"sequence before scalar", []*objectGram{parameter("Sequence"), parameter("Scalar")}, true},
		{"missing binding before sequence", []*objectGram{{noun: &token{}}, parameter("Sequence")}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			node := &constructionNode{Nominal: "holder", Parameters: tc.parameters}
			instance, err := c.emptyConstructionSequence(node)
			if err != nil || (instance != nil) != tc.empty {
				t.Fatalf("empty argument = %+v, %v; want first-parameter availability %v", instance, err, tc.empty)
			}
			if instance != nil && instance.definition != c.Memory["Sequence"] {
				t.Fatal("empty sequence lost its first parameter's canonical constructor")
			}
		})
	}
}
