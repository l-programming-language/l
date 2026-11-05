package contracts

import (
	"bytes"
	"strings"
	"testing"
)

func TestExplanationLEX039(t *testing.T) {
	target := lexicalTarget("primary", "palette")
	selection := SelectReference([]Candidate{lexicalCandidate("dictionary:a", lexicalType("Theme")), lexicalCandidate("dictionary:b", lexicalType("Theme"))}, SelectionContext{Target: target, Environment: "two-imports"})
	dark, light := paletteContribution(target, "dark"), paletteContribution(target, "light")
	configuration := requireConfiguration(t, []FacetSchema{paletteSchema("optional-single")}, nil, []ConfigurationContribution{dark, light}, nil)
	writeTarget := lexicalTarget("secondary", "backend")
	write := FrameFootprint{Target: writeTarget, Kind: "write", Definite: true, Origin: Origin{Source: "wrapper-body.l", Start: 25, End: 39}}
	frame := FrameContract{Target: lexicalTarget(), Scope: "elaboration", Allowed: []TargetRef{target}, Preserves: []TargetRef{writeTarget}, Origin: Origin{Source: "wrapper-interface.l"}}
	report := NewReport("typed-fixture")
	report.ReferenceSelections = []ReferenceSelection{selection}
	report.Configuration = &configuration
	report.Obligations = CheckFrame(frame, []FrameFootprint{write}, FrameContext{})
	report.Summary = DerivedSummary{Coverage: Known, Effects: []Effect{{Target: write.Target, Kind: "write", Definite: true, Origin: write.Origin}}, Resources: ZeroResourceCost()}
	explanations := ExplainReport(report)
	text := FormatExplanations(explanations)
	for _, required := range []string{"dictionary:a", "dictionary:b", "ambiguous", target.String(), writeTarget.String(), "wrapper-body.l:25-39", "wrapper-interface.l", "dark", "light", "applicability=", "requirement="} {
		if !strings.Contains(text, required) {
			t.Errorf("explanation missing %q: %s", required, text)
		}
	}
	if strings.Contains(text, "caller fault") || strings.Contains(text, "use qualifier") {
		t.Fatal("explanation invented blame or unsupported source syntax")
	}
	first, err := EncodeDocument(report)
	if err != nil {
		t.Fatal(err)
	}
	second, err := EncodeDocument(report)
	if err != nil || !bytes.Equal(first, second) {
		t.Fatal("diagnostic interchange is not deterministic", err)
	}
	document, err := DecodeDocument(first)
	if err != nil {
		t.Fatal(err)
	}
	roundtrip, ok := document.(Report)
	if !ok {
		t.Fatalf("report decoded as %T", document)
	}
	if got := FormatExplanations(ExplainReport(roundtrip)); got != text {
		t.Fatal("report roundtrip lost target/candidate/frame explanation")
	}
}
