package Lyger

import (
	"testing"
)

func TestSourceDefinedFunctionConstructor(t *testing.T) {
	source := "Define \"Callable|-s|-like|-----ing\" as a New Function.\n" +
		"Use Function's Return as Callable's \"Return|-s|-ed|-ing|-er\".\nUse Function's Invoke as Callable's \"Invoke|-s|-d|--ing|--er\".\n\n" +
		"    Publicly create \"Echo|-es|-ed|-ing|-er\" as a Generic `identity` Callable with Inputs of \"value\" as `f64` and Outputs of `f64`.\nReturn Value.\n"
	c := compileDefinitionSource(t, source, "constructor.l")
	testNativeDefinitionLibrary(t, c, "Echo", "identity", `#[path="src/identity.rs"] mod generated; fn main() { assert_eq!(generated::identity(7.0), 7.0); }`)
}
