package contracts

import "testing"

func TestSourceStructuralCertificateCannotSmuggleRuntimePromise(t *testing.T) {
	for _, predicate := range []Predicate{
		{Kind: "effects_closed", Subject: "NoBlocking"},
		{Kind: "ownership", Subject: "exclusive", Value: Atom{Kind: "bool", Bool: true}},
		{Kind: "protocol", Subject: "process", Value: Atom{Kind: "state", Text: "finalized"}},
		{Kind: "cost_leq", Subject: "work", Upper: func() *int64 { value := int64(0); return &value }()},
	} {
		t.Run(predicate.Kind, func(t *testing.T) {
			projection, declaration := sourceAdjectiveFixture()
			declaration.Requires = append(declaration.Requires, predicate)
			declaration.Provides = append(declaration.Provides, predicate)
			certificate, err := VerifySourceAdjective(projection, declaration, AnalysisOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if certificate.LogicalStatus == Established {
				t.Fatalf("structural source checker certified a runtime promise via caller precondition: %+v", predicate)
			}
		})
	}
}

func TestSourceImplementationReportCannotHideNativeResiduals(t *testing.T) {
	projection, declaration := sourceAdjectiveFixture()
	certificate, err := VerifySourceAdjective(projection, declaration, AnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	report := NewReport("source-fixture")
	report.SourceCertificates = []SourceAdjectiveCertificate{certificate}
	report.Obligations = certificate.LoweringObligations
	report.SourceVerification = &SourceVerificationResult{Assurance: "implementation", Scope: "whole_dictionary", Selected: []MeaningKey{certificate.Meaning}, Status: Established, Obligations: certificate.LogicalObligations}
	if _, err := EncodeDocument(report); err == nil {
		t.Fatal("implementation success hid unresolved native obligations behind logical-only evidence")
	}
	report.SourceVerification.Assurance = "symbolic"
	if _, err := EncodeDocument(report); err == nil {
		t.Fatal("symbolic success was labeled whole dictionary assurance")
	}
	report.SourceVerification.Scope = "source_elaboration"
	if _, err := EncodeDocument(report); err != nil {
		t.Fatalf("explicit narrow logical success was rejected because native assurance remains unresolved: %v", err)
	}
}
