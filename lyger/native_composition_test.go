package Lyger

import "testing"

func TestNativeConditionComparisonsAndNegation(t *testing.T) {
	const source = "    Publicly create \"Compare|-s|-d|--ing|-r\" as a Generic `compare` Function with Inputs of \"left\" as `f64` and \"right\" as `f64` and Outputs of `f64`.\nIf Left is Right, Return 1, else Return 0.\n"
	for _, tc := range []struct {
		name, operator            string
		negated                   bool
		less, equal, greater, nan string
	}{
		{"equal", "is", false, "0.0", "1.0", "0.0", "0.0"},
		{"different", "isnt", false, "1.0", "0.0", "1.0", "1.0"},
		{"less", "less", false, "1.0", "0.0", "0.0", "0.0"},
		{"greater", "greater", false, "0.0", "0.0", "1.0", "0.0"},
		{"less-or-equal", "less_equal", false, "1.0", "1.0", "0.0", "0.0"},
		{"greater-or-equal", "greater_equal", false, "0.0", "1.0", "1.0", "0.0"},
		{"not-less", "less", true, "0.0", "1.0", "1.0", "1.0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := compileDefinitionSource(t, source, "native-condition.l")
			rec, err := c.exportedRecord("Compare")
			if err != nil {
				t.Fatal(err)
			}
			// Exercise source-checked operands with each resolved comparison.
			condition := rec.function.Body[0].Condition
			condition.Operator, condition.Negated = tc.operator, tc.negated
			testNativeDefinitionLibrary(t, c, "Compare", "native_condition", `#[path="src/native_condition.rs"]mod generated;fn main(){assert_eq!(generated::compare(-1.0,1.0),`+tc.less+`);assert_eq!(generated::compare(1.0,1.0),`+tc.equal+`);assert_eq!(generated::compare(1.0,-1.0),`+tc.greater+`);assert_eq!(generated::compare(f64::NAN,1.0),`+tc.nan+`);}`)
		})
	}
	t.Run("not-truth", func(t *testing.T) {
		c := compileDefinitionSource(t, "    Publicly create \"Choose|-s|----se|--ing|-r\" as a Generic `choose` Function with Inputs of \"flag\" as `bool` and Outputs of `f64`.\nIf Flag, Return 1, else Return 0.\n", "negated-truth.l")
		rec, err := c.exportedRecord("Choose")
		if err != nil {
			t.Fatal(err)
		}
		rec.function.Body[0].Condition.Negated = true
		testNativeDefinitionLibrary(t, c, "Choose", "negated_truth", `#[path="src/negated_truth.rs"]mod generated;fn main(){assert_eq!(generated::choose(true),0.0);assert_eq!(generated::choose(false),1.0);}`)
	})

}
