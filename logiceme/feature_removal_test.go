package Logiceme

import "testing"

func TestLucidRemovedFeatureExpression(t *testing.T) {
	for _, control := range []string{"-{Feature: legacy}-", "={Feature: legacy}="} {
		t.Run(control, func(t *testing.T) {
			// The removed source grouping must not survive as a template
			// expression kind; ordinary literal text is not examined here.
			expressions := []string{"", control, "ordinary contents"}
			if got := determineExpressionType(expressions); got != "Unknown" {
				t.Fatalf("removed Feature control classified as %q", got)
			}
		})
	}
}
