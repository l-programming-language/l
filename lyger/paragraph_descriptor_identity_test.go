package Lyger

import (
	"bytes"
	"strings"
	"testing"
)

const paragraphDescriptorWidget = "Define \"Widget|-s|-al|-ing\" as `<DEF>FW{|Func{Key: <@Name|Text@>, Inputs: <@Inputs|Props@>, Outputs: <@Outputs|Texts@>}Func|}^(<@Process@>)<DEF>`.\n"

func TestParagraphDescriptorTargetDeclarationIdentity(t *testing.T) {
	for _, target := range []string{"Widget", "Handle"} {
		t.Run(target, func(t *testing.T) {
			source := paragraphSource(t)
			source.main = paragraphDescriptorWidget +
				"Use Widget as \"Handle\".\n" +
				"\"Restricted\" is a New Adjective describing any Widget.\n" +
				"It describes a Generic `main` " + target + " of any Process with No Inputs and No Outputs.\n"
			compilation := executableTestResult(t, source).Compilation
			restricted := compilation.Memory["Restricted"]
			descriptor := compilation.Semantics.Descriptors[restricted]
			if restricted == nil || descriptor == nil || len(descriptor.Rules) != 1 {
				t.Fatal("valid canonical target did not register its refinement")
			}
		})
	}
}

func TestParagraphDescriptorNominalConstructionIdentity(t *testing.T) {
	for _, tc := range []struct {
		name, nominal, local string
		valid                bool
	}{
		{"canonical", "Widget", "", true},
		{"canonical alias", "Handle", "", true},
		{"distinct same spelling", "Widget", "\n    " + paragraphDescriptorWidget, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := paragraphSource(t)
			source.main = paragraphGlobals + paragraphDescriptorWidget +
				"Use Widget as \"Handle\".\n" +
				"\"Wrapped\" describes any Widget as `<DEF>[[\"A\", \"<@Widget@>\"]]<DEF>`.\n" +
				"\"Entry\" describes a Generic `main` Widget of any Process with No Inputs and No Outputs.\n" + tc.local +
				"\"Shell\" describes a Generic `main` Object with a New Wrapped Entry " + tc.nominal + " using any Process.\n" +
				"\"Tower\" describes a Generic `main` Service with a New Shell Object using any Process.\n" +
				"\"Custom\" describes a Generic App with a New Tower Service using any Process.\n" +
				"Create \"Worker\" as a New Custom App.\nThe App Prints x1.\n"
			compilation := compileMemberResolutionSource(t, source)
			if !tc.valid {
				if compilation.Error == nil {
					definitions := paragraphBindingDeclarations(t, compilation, "Widget", 2)
					if definitions[0].record == definitions[1].record {
						t.Fatal("prerequisite failed: distinct scoped nominal declarations share a record")
					}
					t.Fatal("descriptor accepted a distinct scoped nominal solely because both definitions are named Widget")
				}
				message := strings.ToLower(compilation.Error.Error())
				if !strings.Contains(message, "wrapped") || !strings.Contains(message, "widget") || !strings.Contains(message, "constructor") {
					t.Fatalf("wrong-target rejection lost descriptor/nominal cause: %v", compilation.Error)
				}
				return
			}
			if compilation.Error != nil {
				t.Fatal(compilation.Error)
			}
			widget := paragraphBindingDeclarations(t, compilation, "Widget", 1)[0].record
			alias := paragraphBindingDeclarations(t, compilation, "Handle", 1)[0].record
			canonical, err := canonicalRecord(alias, compilation.Semantics.Aliases)
			if err != nil || canonical != widget {
				t.Fatalf("nominal alias lost its source definition: %v", err)
			}
			if len(compilation.Semantics.Constructions) != 1 {
				t.Fatal("source-defined nominal did not construct the application")
			}
			root := compilation.Semantics.Constructions[0].Root
			if len(root.Children) != 1 || len(root.Children[0].Children) != 1 || len(root.Children[0].Children[0].Children) != 1 {
				t.Fatal("descriptor expansion lost the service/object/function structure")
			}
			function := root.Children[0].Children[0].Children[0]
			if function.Definition != widget || len(function.Wrappers) != 1 {
				t.Fatal("descriptor alias changed its canonical function or wrapper")
			}
			units, err := (&CompilationResult{Compilation: compilation}).ApplicationUnits()
			if err != nil || len(units) != 1 || !bytes.Contains(units[0].BackendAML, []byte(`Inputs: ~&["Hello World!"]&~`)) {
				t.Fatalf("canonical nominal failed to produce the source-defined App body: %v", err)
			}
		})
	}
}

func TestParagraphDescriptorNativeTargetControl(t *testing.T) {
	source := paragraphSource(t)
	source.main = paragraphWitness
	compilation := executableTestResult(t, source).Compilation
	if len(compilation.Semantics.Constructions) != 1 || compilation.Semantics.Constructions[0].Root.Nominal != "app" {
		t.Fatal("native App target did not retain its constructor")
	}
	units, err := (&CompilationResult{Compilation: compilation}).ApplicationUnits()
	if err != nil || len(units) != 1 || !bytes.Contains(units[0].BackendAML, []byte(`Inputs: ~&["Hello World!"]&~`)) {
		t.Fatalf("native target did not produce its original body: %v", err)
	}
}
