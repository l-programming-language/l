package Lyger

import (
	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	Logiceme "github.com/l-programming-language/l/logiceme"
	"path/filepath"
	"strings"
	"testing"
)

func TestNestedDefinitionOperands(t *testing.T) {
	source := "Define \"Hold|-s|---eld|-ing|-er\" to `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~<@Value@>~|}<DEF>`.\n" +
		"Define \"Count|-s|-ed|-ing|-er\" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~fetch_add~, Receiver: ~<@Value@>~, Inputs: ~&[1, std::sync::atomic::Ordering::SeqCst]&~|}<DEF>`.\n\n" +
		"    Publicly create \"Echo|-es|-ed|-ing|-er\" as a Generic Function with Inputs of \"value\" as `&std::sync::atomic::AtomicUsize` and Outputs of `usize`.\nReturn a New Counter of Value.\n\n" +
		"Use Echo as \"Relay\".\n\n" +
		"    Publicly create \"Result\" as a Generic Function with Inputs of \"__l_nested_0\" as `&std::sync::atomic::AtomicUsize` and Outputs of `usize`.\nCreate \"Previous\" as a New Holder of a New Relay of __l_nested_0.\nReturn a New Holder of a New Relay of __l_nested_0.\n"
	c := compileDefinitionSource(t, source, "nested.l")
	definition, err := c.ComposeDefinition("Result")
	if err != nil {
		t.Fatal(err)
	}
	line := strings.Count(source[:strings.Index(source, "Return a New Holder")], "\n") + 1
	provenance := false
	for _, binding := range definition.Bindings {
		for _, span := range binding.Sources {
			provenance = provenance || span.Path == "nested.l" && span.Start.Line == line
		}
	}
	if !provenance {
		t.Fatal("nested result lost its use-site source span")
	}
	testNativeDefinitionLibrary(t, c, "Result", "nested", `#[path="src/nested.rs"] mod generated; fn main(){let value=std::sync::atomic::AtomicUsize::new(7);assert_eq!(generated::result(&value),8);assert_eq!(value.load(std::sync::atomic::Ordering::SeqCst),9);}`)
}

func TestDefinitionChildAndSiblingGrouping(t *testing.T) {
	for source, want := range map[string]string{
		"W^(MOB^KY*(OB*BOD^E))":                 "W^(MOB^KY*OB*BOD^E)",
		"GI^(E*W^(MOB^KY*OB)*E*W^BOD^E)":        "GI^(E*W^(MOB^KY*OB)*E*W^BOD^E)",
		"W^((MOB^KY*OB))":                       "W^(MOB^KY*OB)",
		"W^(MOB{|Key: ~value~|}^KY*(OB*BOD^E))": "W^(MOB{|Key: ~value~|}^KY*OB*BOD^E)",
	} {
		got, err := normalizeDefinitionGroups(source)
		if err != nil || got != want {
			t.Fatalf("grouping %q: got %q, want %q: %v", source, got, want, err)
		}
	}
}

func testNativeDefinitionLibrary(t *testing.T, c *Compilation, name, filename, observer string) {
	t.Helper()
	aml, err := c.ComposeLibrary(name, filename)
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := Logiceme.NewLogicemeService(c.Gram.resources).Evaluate(aml)
	if err != nil {
		t.Fatal(err)
	}
	files := make(map[string][]byte)
	for index, name := range bundle.MapArrays[0] {
		files[strings.TrimPrefix(name, "/")] = bundle.ByteArrays[0][index]
	}
	files["observe.rs"] = []byte(observer)
	directory := nativetest.Materialize(t, files)
	tools := nativetest.Toolchain(t)
	binary := filepath.Join(directory, "observe")
	_, diagnostic, err := nativetest.BuildPrerequisite.Run(t, tools.Rustc, directory, nativetest.CleanEnvironment(tools.Rustc), "--edition=2024", "observe.rs", "-o", binary)
	if err != nil {
		t.Fatalf("native nested composition: %v: %s", err, diagnostic)
	}
	if _, diagnostic, err = nativetest.RuntimeExecution.Run(t, binary, directory, nil); err != nil {
		t.Fatalf("nested result: %v: %s", err, diagnostic)
	}
}
