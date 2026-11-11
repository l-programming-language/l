package Lyger

import (
	"strings"
	"testing"
)

const processInputWords = "Define \"Choose|-s|-d|--ing|-r\" to `<DEF>BOG^GI{|Key: ~<@Value@>~|}^(<@First@>*<@Second@>)<DEF>`.\n" +
	"Define \"Present\" as `<DEF>E{|Statement: ~Some(<@Payload|Label@>)~|}*(<@Body@>)<DEF>`.\n" +
	"Define \"Absent\" as `<DEF>E{|Statement: ~None~|}*(<@Body@>)<DEF>`.\n" +
	"Define \"Wrap|-s|-ped|-ping|-per\" to `<DEF>W^<@Body@><DEF>`.\n" +
	"Define \"Walk|-s|-ed|-ing|-er\" to `<DEF>BOP^X{|Iterable: ~<@Values@>~, Key: ~<@Second|Label@>~, Index: ~<@First|Label@>~, RangeArity: ~2~|}^(<@Body@>)<DEF>`.\n"

func TestStructuredProcessBindingArity(t *testing.T) {
	for _, tc := range []struct{ name, expression, diagnostic string }{
		{"missing", "Choose Value, a New Present using a New Process for Returning 1, and a New Absent using a New Process for Returning 0.", "missing argument for Body"},
		{"text-label", "Choose Value, a New Present of `payload` using a New Process for Returning 1, and a New Absent using a New Process for Returning 0.", "missing argument for Body"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := NewLygerService(canonicalTestOptions()).NewCompilation([]byte(processInputWords+"\n    Publicly create \"Exercise\" as a Generic Function with Inputs of \"value\" as `Option<f64>` and Outputs of `f64`.\n"+tc.expression+"\n"), "process-arity.l")
			c.Make()
			err := c.Error
			if err == nil {
				_, err = c.ComposeLibrary("Exercise", "arity")
			}
			if err == nil || !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tc.diagnostic)) {
				t.Fatalf("invalid Process binding accepted or wrong diagnostic: %v", err)
			}
		})
	}
}
