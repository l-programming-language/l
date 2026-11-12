package Lyger

import (
	"testing"
)

func TestComposedProcessAliasPreservesBody(t *testing.T) {
	source := "Use Process as \"Sequence|-s|-like|--ing\".\nCreate \"Perform|-s|-ed|-ing|-er\" for Processing any Sequence.\n\n" +
		"    Publicly create \"Exercise|-s|-d|--ing|-r\" as a Generic `exercise` Function with No Inputs and Outputs of `f64`.\nPerform a New Process for Returning 7.\n"
	c := compileDefinitionSource(t, source, "process-alias.l")
	testNativeDefinitionLibrary(t, c, "Exercise", "process_alias", `#[path="src/process_alias.rs"] mod generated; fn main() { assert_eq!(generated::exercise(), 7.0); }`)
}
