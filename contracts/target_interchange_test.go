package contracts

import (
	"encoding/json"
	"testing"
)

func TestTargetIdentitySurvivesEmptyPathInterchange(t *testing.T) {
	compiler := TargetRef{Root: "instance:primary", Type: "Theme", Phase: Construction}
	var authored TargetRef
	if err := json.Unmarshal([]byte(`{"root":"instance:primary","path":[],"type":"Theme","phase":"construction"}`), &authored); err != nil {
		t.Fatal(err)
	}
	if TargetRelation(compiler, authored) != Established {
		t.Fatal("an explicit empty path changed the exact attachment identity")
	}
	authored.Path = []string{"Secondary"}
	if TargetRelation(compiler, authored) != Contradicted {
		t.Fatal("normalization collapsed a distinct role")
	}
}

func TestAdjectiveOmittedCallbackCostsAndOutcomesStayUnknown(t *testing.T) {
	transformer := supportedTransformer()
	certificate := requireCertificate(t, transformer)
	if _, known := certificate.Summary.Resources.Work.Constant(); known {
		t.Fatal("a closed effect row silently supplied a constant callback work bound")
	}
	if certificate.Status != Established {
		t.Fatal("unknown cost invalidated an independent nonblocking theorem")
	}
	callback := transformer.Callbacks["action"]
	callback.Outcomes = nil
	transformer.Callbacks["action"] = callback
	transformer.Profiles = append(transformer.Profiles, Profile{Name: "NoPanic"})
	certificate = requireCertificate(t, transformer)
	if certificate.Status != Unresolved || len(certificate.Summary.Unknowns) == 0 {
		t.Fatal("omitted callback outcome coverage became success-only execution")
	}
}
