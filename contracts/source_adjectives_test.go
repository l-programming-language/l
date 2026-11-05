package contracts

import (
	"bytes"
	"testing"
)

func sourceAdjectiveFixture() (SourceAdjectiveProjection, SourceTransformerDeclaration) {
	meaning := MeaningKey("source:compact-rule")
	target := SourceSymbolicTarget(meaning, "App")
	target.Nominal = "native:app"
	child := target
	child.Path, child.Type, child.Nominal = []string{"service[0]"}, "Service", "native:service"
	before := SourceShape{Target: target, Nominal: target.Nominal, ChildrenClosed: true}
	after := before
	after.Children = []SourceShape{{Target: child, Nominal: child.Nominal, Name: "main", NameKnown: true, ChildrenClosed: true}}
	write := FrameFootprint{Target: child, Kind: "write", Definite: true, Origin: Origin{Source: "source-adjective.l", Start: 12, End: 48}}
	empty := Predicate{Kind: "source_children_empty", Target: target}
	projection := SourceAdjectiveProjection{Meaning: meaning, Version: SourceProjectionVersion, Target: target, TargetNominal: target.Nominal, Origin: write.Origin, Environment: "definition-site-env", Dependencies: map[string]string{"source-adjective.l": "source-v1"}, Before: before, After: after, Footprints: []FrameFootprint{write}, RequiredAssumptions: []Predicate{empty}, Plan: Plan{Entry: Node{Kind: Sequence, Phase: Elaboration, Children: []Node{{Kind: Event, Phase: Elaboration, Effects: []Effect{{Kind: "write", Target: child, Phase: Elaboration, Definite: true, Origin: write.Origin}}}}}}, LoweringObligations: []Obligation{{ID: "native-child", Status: Unresolved, Reason: "selected child wrapper requires native conformance"}}}
	declaration := SourceTransformerDeclaration{Meaning: meaning, Version: SchemaVersion, TargetNominal: target.Nominal, Requires: []Predicate{empty}, Provides: []Predicate{{Kind: "present", Subject: "child", Target: child}, {Kind: "meaning", Subject: "nominal", Target: child, Value: Atom{Kind: "meaning", Text: string(child.Nominal)}}}, Frame: &FrameContract{Target: target, Allowed: []TargetRef{child}, Scope: "elaboration", Phase: Elaboration}}
	return projection, declaration
}

func TestSourceAdjectiveStructuralCertificate(t *testing.T) {
	projection, declaration := sourceAdjectiveFixture()
	certificate, err := VerifySourceAdjective(projection, declaration, AnalysisOptions{})
	if err != nil || certificate.LogicalStatus != Established || certificate.Scope != "source_elaboration" || len(certificate.Assumptions) != 1 || len(certificate.LoweringObligations) != 1 || certificate.LoweringObligations[0].Status != Unresolved {
		t.Fatalf("structural logical theorem conflated with native assurance: %+v %v", certificate, err)
	}
	for _, test := range []struct {
		name   string
		mutate func(*SourceAdjectiveProjection, *SourceTransformerDeclaration)
		want   Status
		mode   string
	}{
		{"missing-declared-shape", func(_ *SourceAdjectiveProjection, d *SourceTransformerDeclaration) { d.Requires = nil }, Unresolved, ""},
		{"frame-rejection", func(p *SourceAdjectiveProjection, d *SourceTransformerDeclaration) {
			d.Frame.Preserves = []TargetRef{p.Target}
		}, Contradicted, ""},
		{"runtime-profile", func(_ *SourceAdjectiveProjection, d *SourceTransformerDeclaration) {
			d.Profiles = []Profile{{Name: "NoHeapAllocation"}, {Name: "NoBlocking"}}
		}, Unresolved, ""},
		{"unsupported-projection-mode", func(p *SourceAdjectiveProjection, _ *SourceTransformerDeclaration) {
			p.ProjectionObligations = []Obligation{{ID: "reflection", Status: Unresolved, Reason: "formal-dependent selection"}}
		}, Unresolved, "instantiation_only"},
		{"unrelated-frame-target", func(_ *SourceAdjectiveProjection, d *SourceTransformerDeclaration) {
			d.Frame.Target.Root = "another-symbolic-target"
		}, Unresolved, ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			projection, declaration := sourceAdjectiveFixture()
			test.mutate(&projection, &declaration)
			got, err := VerifySourceAdjective(projection, declaration, AnalysisOptions{})
			if err != nil || got.LogicalStatus != test.want || test.mode != "" && got.Mode != test.mode {
				t.Fatalf("status=%s want=%s mode=%s: %+v %v", got.LogicalStatus, test.want, test.mode, got, err)
			}
		})
	}
}

func TestSourcePredicateFragmentIsRecursiveAndTargeted(t *testing.T) {
	for _, form := range []string{"nested-runtime", "unscoped", "invented-field", "runtime-target"} {
		t.Run(form, func(t *testing.T) {
			projection, declaration := sourceAdjectiveFixture()
			predicate := Predicate{Kind: "eq", Subject: "name", Target: projection.Target, Value: Atom{Kind: "text", Text: "unchanged"}}
			switch form {
			case "nested-runtime":
				predicate = Predicate{Kind: "implies", Args: []Predicate{{Kind: "true"}, {Kind: "effects_closed", Subject: "NoBlocking"}}}
			case "unscoped":
				predicate.Target = TargetRef{}
			case "invented-field":
				predicate.Subject = "effects.NoBlocking"
			case "runtime-target":
				predicate.Target.Phase = Invocation
			}
			declaration.Requires = append(declaration.Requires, predicate)
			declaration.Provides = append(declaration.Provides, predicate)
			got, err := VerifySourceAdjective(projection, declaration, AnalysisOptions{})
			if err != nil || got.LogicalStatus == Established {
				t.Fatalf("unsupported source predicate accepted: %+v %v", got, err)
			}
		})
	}
}

func sourceInstanceFixture(projection SourceAdjectiveProjection) SourceAdjectiveInstance {
	target := projection.Target
	target.Root = "concrete-app"
	var rebase func(SourceShape) SourceShape
	rebase = func(shape SourceShape) SourceShape {
		shape.Target = rebaseAdjectiveTarget(shape.Target, projection.Target, target)
		shape.Children = append([]SourceShape(nil), shape.Children...)
		for i := range shape.Children {
			shape.Children[i] = rebase(shape.Children[i])
		}
		return shape
	}
	writes := append([]FrameFootprint(nil), projection.Footprints...)
	for i := range writes {
		writes[i].Target = rebaseAdjectiveTarget(writes[i].Target, projection.Target, target)
	}
	return SourceAdjectiveInstance{Meaning: projection.Meaning, Target: target, TargetNominal: projection.TargetNominal, Before: rebase(projection.Before), After: rebase(projection.After), Footprints: writes}
}

func TestSourceAdjectiveConcreteBeforeSnapshot(t *testing.T) {
	projection, declaration := sourceAdjectiveFixture()
	certificate, err := VerifySourceAdjective(projection, declaration, AnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, scenario := range []string{"fresh", "existing-child", "different-canonical-same-spelling", "missing-nominal"} {
		t.Run(scenario, func(t *testing.T) {
			instance := sourceInstanceFixture(projection)
			want := Established
			switch scenario {
			case "existing-child":
				instance.Before.Children = instance.After.Children
				want = Contradicted
			case "different-canonical-same-spelling":
				instance.TargetNominal = "dictionary:another-App"
				instance.Target.Nominal = instance.TargetNominal
				instance.Before.Nominal, instance.After.Nominal = instance.TargetNominal, instance.TargetNominal
				want = Contradicted
			case "missing-nominal":
				instance.TargetNominal = ""
				want = Unresolved
			}
			got, err := VerifySourceAdjectiveInstance(projection, certificate, declaration, instance, AnalysisOptions{})
			if err != nil || got.Status != want {
				t.Fatalf("%s: status=%s want=%s error=%v obligations=%+v", scenario, got.Status, want, err, got.Obligations)
			}
		})
	}
}

func TestSourceNestedTargetNominalSurvivesSubstitution(t *testing.T) {
	projection, declaration := sourceAdjectiveFixture()
	instance := sourceInstanceFixture(projection)
	child := projection.After.Children[0].Target
	mapped := rebaseTarget(child, Substitutions{Targets: map[string]TargetRef{projection.Target.Root: instance.Target}})
	if mapped.Type != child.Type || mapped.Nominal != child.Nominal || mapped.Root != instance.Target.Root {
		t.Fatalf("nested role became enclosing target type: %+v", mapped)
	}
	certificate, err := VerifySourceAdjective(projection, declaration, AnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	result, err := VerifySourceAdjectiveInstance(projection, certificate, declaration, instance, AnalysisOptions{})
	if err != nil || result.Status != Established {
		t.Fatalf("nested source instance failed: %+v %v", result, err)
	}
	found := false
	for _, obligation := range result.Obligations {
		predicate := obligation.Claim.Predicate
		if predicate.Target.Root == instance.Target.Root && len(predicate.Target.Path) != 0 {
			found = true
			if predicate.Target.Type != child.Type || predicate.Target.Nominal != child.Nominal {
				t.Fatalf("concrete child claim lost nominal identity: %+v", predicate.Target)
			}
		}
	}
	if !found {
		t.Fatal("concrete child claim was not reported")
	}
}

func TestSourceStructuralFactInvalidation(t *testing.T) {
	root := TargetRef{Root: "app", Phase: Elaboration, Nominal: "native:app"}
	child := root
	child.Path = []string{"service[0]"}
	name := Predicate{Kind: "eq", Subject: "name", Target: root, Value: Atom{Kind: "text", Text: "original"}}
	member := Predicate{Kind: "present", Subject: "member", Target: child}
	constructor := Predicate{Kind: "meaning", Subject: "constructor", Target: root, Value: Atom{Kind: "meaning", Text: "old-constructor"}}
	for _, test := range []struct {
		name  string
		write TargetRef
		kind  string
		fact  Predicate
		want  Status
	}{
		{"ancestor-replacement", root, "replace", member, Unresolved},
		{"child-write-preserves-name", child, "write", name, Established},
		{"constructor-write", func() TargetRef { v := root; v.Path = []string{"constructor"}; return v }(), "write", constructor, Unresolved},
		{"unknown-write", TargetRef{Unknown: true}, "replace", member, Unresolved},
	} {
		t.Run(test.name, func(t *testing.T) {
			plan := Plan{Entry: Node{Kind: Sequence, Children: []Node{{Kind: Event, Facts: []Predicate{test.fact}}, {Kind: Event, Phase: Elaboration, Effects: []Effect{{Kind: EffectKind(test.kind), Target: test.write, Phase: Elaboration, Definite: true}}}, {Kind: Require, Requires: []Predicate{test.fact}}}}}
			summary := reviewInfer(t, plan, AnalysisOptions{})
			if len(summary.Obligations) != 1 || summary.Obligations[0].Status != test.want {
				t.Fatalf("stale structural fact handling: %+v", summary.Obligations)
			}
		})
	}
}

func TestSourceShapeKnownListsPreserveOrder(t *testing.T) {
	projection, _ := sourceAdjectiveFixture()
	shape := projection.Before
	shape.SignatureKnown, shape.ParametersKnown, shape.WrappersKnown = true, true, true
	shape.InputsDefined, shape.OutputsDefined = true, true
	shape.Inputs = []SourceSignatureInput{{Name: "a", Type: "&str"}, {Name: "b", Type: "u8"}}
	shape.Outputs = []string{"bool"}
	shape.Wrappers = []MeaningKey{"wrapper-a", "wrapper-a", "wrapper-b"}
	shape.Parameters = []SourceParameter{{Meaning: "parameter-a", Ordinal: 0}, {Meaning: "parameter-b", Ordinal: 1, Formal: true}}
	facts := SourceShapeFacts(shape)
	for _, want := range []Predicate{
		{Kind: "eq", Subject: "inputs.defined", Target: shape.Target, Value: Atom{Kind: "bool", Bool: true}},
		{Kind: "eq", Subject: "inputs[1].type", Target: shape.Target, Value: Atom{Kind: "type", Text: "u8"}},
		{Kind: "eq", Subject: "wrappers.count", Target: shape.Target, Value: Atom{Kind: "number", Number: 3}},
		{Kind: "eq", Subject: "wrappers[1].meaning", Target: shape.Target, Value: Atom{Kind: "meaning", Text: "wrapper-a"}},
		{Kind: "eq", Subject: "parameters[1].formal", Target: shape.Target, Value: Atom{Kind: "bool", Bool: true}},
	} {
		if got, _ := Entails(facts, want); got != Established {
			t.Fatalf("known ordered shape fact missing: %+v", want)
		}
	}
	shape.WrappersKnown = false
	if got, _ := Entails(SourceShapeFacts(shape), Predicate{Kind: "eq", Subject: "wrappers.count", Target: shape.Target, Value: Atom{Kind: "number", Number: 3}}); got != Unresolved {
		t.Fatal("open wrapper list gained exact count")
	}
}

func TestSourceManifestRejectsAuthoredBody(t *testing.T) {
	_, declaration := sourceAdjectiveFixture()
	manifest := FacetRoleManifest{SchemaVersion: SchemaVersion, Kind: "facet_and_role_manifest", Producer: "test-only", Facets: []FacetSchema{}, Roles: []RoleInterface{}, SourceTransformers: []SourceTransformerDeclaration{declaration}}
	data, err := EncodeDocument(manifest)
	if err != nil {
		t.Fatal(err)
	}
	bad := bytes.Replace(data, []byte(`"target_nominal": "native:app",`), []byte(`"target_nominal": "native:app", "body": {"entry":{"kind":"event","facts":[{"kind":"true"}] }},`), 1)
	if bytes.Equal(bad, data) {
		t.Fatal("test injection did not reach declaration")
	}
	if _, err := DecodeDocument(bad); err == nil {
		t.Fatal("metadata-authored body accepted as source projection")
	}
}

func TestSourceNarrowReportKeepsNativeResiduals(t *testing.T) {
	projection, declaration := sourceAdjectiveFixture()
	certificate, err := VerifySourceAdjective(projection, declaration, AnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	report := NewReport("source-adjective.l")
	report.SourceCertificates = []SourceAdjectiveCertificate{certificate}
	report.Obligations = certificate.LoweringObligations
	report.SourceVerification = &SourceVerificationResult{Assurance: "symbolic", Scope: "source_elaboration", Selected: []MeaningKey{projection.Meaning}, Status: Established, Obligations: certificate.LogicalObligations}
	if _, err := EncodeDocument(report); err != nil {
		t.Fatal(err)
	}
	report.SourceCertificates[0].LogicalObligations = append(report.SourceCertificates[0].LogicalObligations, Obligation{ID: "missing", Status: Unresolved})
	if _, err := EncodeDocument(report); err == nil {
		t.Fatal("source logical status hid unresolved premise")
	}
}
