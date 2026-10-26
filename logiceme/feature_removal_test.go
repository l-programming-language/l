package Logiceme

import "testing"

func TestLucidRemovedFeatureExpression(t *testing.T) {
	for _, control := range []string{"-{Feature: sample}-", "={Feature: sample}="} {
		t.Run(control, func(t *testing.T) {
			// The unsupported source grouping must not be a template
			// expression kind; ordinary literal text is not examined here.
			expressions := []string{"", control, "ordinary contents"}
			if got := determineExpressionType(expressions); got != "Unknown" {
				t.Fatalf("unsupported Feature control classified as %q", got)
			}
		})
	}
}
