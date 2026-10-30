package Lyger

import (
	"bytes"
	"strings"
	"testing"
)

func TestSourceDefinedFunctionConstructor(t *testing.T) {
	source := "Define \"Callable\" as a New Function.\n" +
		"Use Function's Return as Callable's \"Return\".\nUse Function's Invoke as Callable's \"Invoke\".\n\n" +
		"    Publicly create \"Identity\" as a Generic Callable with Inputs of \"value\" as `f64` and Outputs of `f64`.\nReturn Value.\n"
	c := NewLygerService(canonicalTestOptions()).NewCompilation([]byte(source), "constructor.l")
	c.Make()
	if c.Error != nil {
		t.Fatal(c.Error)
	}
	testNativeDefinitionLibrary(t, c, "Identity", "identity", `#[path="src/identity.rs"] mod generated; fn main() { assert_eq!(generated::identity(7.0), 7.0); }`)
}

func TestRuntimeFunctionPreservesConstructorDescriptor(t *testing.T) {
	source := "Define \"Tagged\" as a New Adjective describing any Function.\n" +
		"It describes any Function as `<DEF>FW{|Func{Key: <@Name|Text@>, Inputs: <@Inputs|Props@>, Outputs: <@Outputs|Texts@>}Func|}^(W^(BOD^E{|Statement: ~/* selected descriptor */~|})*<@Process@>)<DEF>`.\n\n" +
		strings.Replace(identityFunction, "New Function", "New Tagged Function", 1) + "Identity `value` as \"Result\".\n"
	_, unit, bundle, err := builtinProcessGenerate(dictionaryIndexFilesystem(t), source)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(unit.BackendAML, []byte("/* selected descriptor */")) || !strings.Contains(builtinProcessRust(bundle), "/* selected descriptor */") {
		t.Fatal("ordinary Function discarded its selected descriptor")
	}
}
