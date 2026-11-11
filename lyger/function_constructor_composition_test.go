package Lyger

import (
	"testing"
)

func TestSourceDefinedFunctionConstructor(t *testing.T) {
	source := "Define \"Callable\" as a New Function.\n" +
		"Use Function's Return as Callable's \"Return\".\nUse Function's Invoke as Callable's \"Invoke\".\n\n" +
		"    Publicly create \"Identity\" as a Generic Callable with Inputs of \"value\" as `f64` and Outputs of `f64`.\nReturn Value.\n"
	c := compileDefinitionSource(t, source, "constructor.l")
	testNativeDefinitionLibrary(t, c, "Identity", "identity", `#[path="src/identity.rs"] mod generated; fn main() { assert_eq!(generated::identity(7.0), 7.0); }`)
}
