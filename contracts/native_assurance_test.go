package contracts

import "testing"

func nativeAssuranceFixture() (NativeUse, DeclaredContract, NativeEnvironment) {
	unknown := true
	signature := &NativeSignature{Inputs: []string{"value"}, Outputs: []string{"Number"}, InputTypes: []NativeType{{Kind: "number"}}, OutputTypes: []NativeType{{Kind: "number"}}}
	use := NativeUse{ID: "test-only:leaf-use", Meaning: "test-only:leaf", Name: "Leaf", Target: TargetRef{Root: "test-only:target", Phase: Invocation}, Phase: Invocation, Boundary: "native_body", Signature: signature, Implementation: "test-only:body-v1", Dependencies: map[string]string{"helper": "helper-v1"}, OutcomeDisposition: "propagate_returned_failure"}
	declaration := DeclaredContract{ID: "test-only:leaf-contract", Meaning: use.Meaning, Name: "Leaf", Signature: signature, Outcomes: []Outcome{Normal, ReturnedFailure}, OutcomeContracts: []NativeOutcomeContract{
		{Outcome: Normal, Effects: &DeclaredEffectRow{UnknownTail: &unknown}, AbsentEffects: []EffectKind{"block"}},
		{Outcome: ReturnedFailure, Effects: &DeclaredEffectRow{UnknownTail: &unknown}, AbsentEffects: []EffectKind{"block"}},
	}}
	environment := SealNativeEnvironment(NativeEnvironment{SchemaVersion: SchemaVersion, Kind: "native_environment", Producer: AnalyzerVersion, Version: "1", Method: "test-only", Host: "test-only", Target: "test-only-target", Compiler: "test-only-compiler", PanicPolicy: "test-only-unwind", Allocator: "test-only-allocator", Inputs: map[string]string{"source": "source-v1"}, Dependencies: map[string]string{"runtime": "runtime-v1"}, Limitations: []string{"Synthetic evidence; no production approval or native execution"}})
	return use, declaration, environment
}

func nativeAssuranceWitnesses(use NativeUse, declaration DeclaredContract, environment NativeEnvironment, selected func(NativeComponent) bool) (TrustPolicy, []NativeWitness) {
	policy := TrustPolicy{Version: SchemaVersion, ID: "test-only:assurance-policy", Authorities: []string{"test-only:authority"}, AllowTestOnly: true}
	var witnesses []NativeWitness
	for _, component := range evaluateNativeUse(use, declaration, environment, TrustPolicy{}, nil, Unresolved).Components {
		if selected != nil && !selected(component) {
			continue
		}
		witness := NativeWitness{ID: "test-only:" + component.Claim.ID, ClaimFingerprint: ClaimFingerprint(component.Claim), ContractFingerprint: component.Current.Contract, ImplementationFingerprint: component.Current.Implementation, TargetFingerprint: component.Current.Target, Dependencies: component.Current.Dependencies, Authority: policy.Authorities[0], Rationale: "Synthetic component assumption for this test only; grants no production approval", TestOnly: true}
		witnesses = append(witnesses, witness)
		policy.ApprovedWitnesses = append(policy.ApprovedWitnesses, witness.ID)
	}
	return policy, witnesses
}

func nativeAssuranceNode(use NativeUse) Node {
	return Node{Kind: Native, ID: use.ID, Native: use.ID, Meaning: use.Meaning, Target: use.Target, Phase: use.Phase}
}

func TestNativeAssuranceExactComponentApproval(t *testing.T) {
	use, declaration, environment := nativeAssuranceFixture()
	for _, selection := range []string{"none", "outcomes", "absence", "both"} {
		policy, witnesses := nativeAssuranceWitnesses(use, declaration, environment, func(component NativeComponent) bool {
			return selection == "both" && (component.Component == "outcomes" || component.Component == "absent:block") || selection == "outcomes" && component.Component == "outcomes" || selection == "absence" && component.Component == "absent:block"
		})
		application := evaluateNativeUse(use, declaration, environment, policy, witnesses, Established)
		summary := reviewInfer(t, Plan{Entry: nativeAssuranceNode(use)}, AnalysisOptions{NativeApplications: map[string]NativeEvaluation{use.ID: application}})
		if AllEstablished(CheckProfile(summary, Profile{Name: "NoBlocking"})) != (selection == "both") {
			t.Fatalf("selection=%s misapplied component approval: %+v", selection, summary)
		}
		for _, profile := range []string{"NoHeapAllocation", "NoExternalIO", "NoPanic"} {
			if AllEstablished(CheckProfile(summary, Profile{Name: profile})) {
				t.Fatalf("blocking approval certified %s", profile)
			}
		}
		if len(summary.Unknowns) == 0 || summary.Resources.Work.Kind != "unknown" {
			t.Fatal("component approval erased unrelated native uncertainty")
		}
	}
}

func TestNativeAssuranceChangedDependenciesReject(t *testing.T) {
	for _, mutation := range []string{"signature", "boundary", "phase", "environment", "unsealed-environment", "implementation", "helper", "missing-witness", "stale-witness", "test-policy", "applicability"} {
		t.Run(mutation, func(t *testing.T) {
			use, declaration, environment := nativeAssuranceFixture()
			policy, witnesses := nativeAssuranceWitnesses(use, declaration, environment, nil)
			applicability := Established
			switch mutation {
			case "boundary":
				use.Boundary = ""
			case "phase":
				use.Phase = Destruction
			case "signature":
				changed := *use.Signature
				changed.Outputs = []string{"Text"}
				use.Signature = &changed
			case "environment":
				environment.PanicPolicy = "changed"
				environment = SealNativeEnvironment(environment)
			case "unsealed-environment":
				environment.PanicPolicy = "changed"
			case "implementation":
				use.Implementation = "body-v2"
			case "helper":
				use.Dependencies["helper"] = "helper-v2"
			case "missing-witness":
				witnesses = nil
			case "stale-witness":
				for i := range witnesses {
					witnesses[i].ImplementationFingerprint = "stale"
				}
			case "test-policy":
				policy.AllowTestOnly = false
			case "applicability":
				applicability = Unresolved
			}
			application := evaluateNativeUse(use, declaration, environment, policy, witnesses, applicability)
			components := application.Components
			for _, component := range components {
				if component.Status == Established {
					t.Fatalf("%s retained component approval: %+v", mutation, component)
				}
			}
			if AllEstablished(CheckProfile(application.Summary, Profile{Name: "NoBlocking"})) {
				t.Fatalf("%s retained guarantee", mutation)
			}
		})
	}
}

func TestNativeAssuranceResultDisposition(t *testing.T) {
	for _, disposition := range []string{"value", "propagate_returned_failure"} {
		use, declaration, environment := nativeAssuranceFixture()
		use.OutcomeDisposition = disposition
		nativeAssuranceOutcomes(&declaration, ReturnedFailure)
		policy, witnesses := nativeAssuranceWitnesses(use, declaration, environment, nil)
		application := evaluateNativeUse(use, declaration, environment, policy, witnesses, Established)
		plan := Plan{Entry: Node{Kind: Sequence, Children: []Node{nativeAssuranceNode(use), {Kind: Event, Effects: []Effect{{Kind: "allocate", Definite: true}}}}}}
		summary := reviewInfer(t, plan, AnalysisOptions{NativeApplications: map[string]NativeEvaluation{use.ID: application}})
		want := ReturnedFailure
		if disposition == "value" {
			want = Normal
		}
		if len(summary.Paths) != 1 || summary.Paths[0].Outcome != want || summary.Paths[0].ObservedOutcome != ReturnedFailure {
			t.Fatalf("%s changed result channel: %+v", disposition, summary.Paths)
		}
		allocated := false
		for _, effect := range summary.Effects {
			allocated = allocated || effect.Kind == "allocate"
		}
		if allocated != (disposition == "value") {
			t.Fatalf("%s changed continuation reachability", disposition)
		}
	}
}

func TestNativeAssuranceRequirementsUseEntryFacts(t *testing.T) {
	for _, before := range []bool{false, true} {
		use, declaration, environment := nativeAssuranceFixture()
		nativeAssuranceOutcomes(&declaration, Normal)
		fact := Predicate{Kind: "eq", Subject: "ready", Value: Atom{Kind: "bool", Bool: true}}
		declaration.Requirements = []Predicate{fact}
		policy, witnesses := nativeAssuranceWitnesses(use, declaration, environment, nil)
		application := evaluateNativeUse(use, declaration, environment, policy, witnesses, Established)
		nodes := []Node{nativeAssuranceNode(use), {Kind: Event, Facts: []Predicate{fact}}}
		if before {
			nodes[0], nodes[1] = nodes[1], nodes[0]
		}
		summary := reviewInfer(t, Plan{Entry: Node{Kind: Sequence, Children: nodes}}, AnalysisOptions{NativeApplications: map[string]NativeEvaluation{use.ID: application}})
		if AllEstablished(CheckProfile(summary, Profile{Name: "NoBlocking"})) != before {
			t.Fatalf("before=%t entry requirement mischecked: %+v", before, summary.Obligations)
		}
	}
}

func TestNativeAssuranceUnknownHelperDoesNotInheritExclusion(t *testing.T) {
	use, declaration, environment := nativeAssuranceFixture()
	nativeAssuranceOutcomes(&declaration, Normal)
	policy, witnesses := nativeAssuranceWitnesses(use, declaration, environment, nil)
	application := evaluateNativeUse(use, declaration, environment, policy, witnesses, Established)
	summary := reviewInfer(t, Plan{Entry: Node{Kind: Sequence, Children: []Node{nativeAssuranceNode(use), {Kind: Native, Native: "independent-opaque-helper"}}}}, AnalysisOptions{NativeApplications: map[string]NativeEvaluation{use.ID: application}})
	if AllEstablished(CheckProfile(summary, Profile{Name: "NoBlocking"})) {
		t.Fatal("leaf approval closed an independent helper")
	}
}

func TestNativeCheckedResultRunsWrapperBeforePropagatingFailure(t *testing.T) {
	use, declaration, environment := nativeAssuranceFixture()
	use.OutcomeDisposition = "value"
	nativeAssuranceOutcomes(&declaration, ReturnedFailure)
	resource := ResourceRef{ID: "state"}
	declaration.OutcomeContracts[0].States = []StateRelation{{Location: FailureLocation{Resource: resource}, Observation: "state_at_completion", Relation: "unchanged"}}
	policy, witnesses := nativeAssuranceWitnesses(use, declaration, environment, nil)
	application := evaluateNativeUse(use, declaration, environment, policy, witnesses, Established)
	for _, wrapper := range []Node{
		{Kind: Event, Effects: []Effect{{Kind: "block", Definite: true}}},
		{Kind: Event, Effects: []Effect{{Kind: "io", Resource: resource, Definite: true}}},
		{Kind: Event, Outcome: Panic},
		{Kind: Event, Outcome: Panic, Effects: []Effect{{Kind: "io", Resource: resource, Definite: true}}},
		{Kind: Native, Native: "opaque-checked-wrapper"},
	} {
		plan := Plan{Entry: Node{Kind: Sequence, Children: []Node{{Kind: CheckedResult, Children: []Node{nativeAssuranceNode(use)}, Cleanup: []Node{wrapper}}, {Kind: Event, Effects: []Effect{{Kind: "allocate", Definite: true}}}}}}
		summary := reviewInfer(t, plan, AnalysisOptions{NativeApplications: map[string]NativeEvaluation{use.ID: application}})
		want := ReturnedFailure
		if wrapper.Outcome == Panic {
			want = Panic
		}
		if summary.Paths[0].Outcome != want {
			t.Fatalf("wrapper result lost: %+v", summary.Paths)
		}
		for _, effect := range summary.Effects {
			if effect.Kind == "allocate" {
				t.Fatal("checked failure executed continuation")
			}
		}
		if len(wrapper.Effects) > 0 && wrapper.Effects[0].Kind == "block" {
			statusTest(t, summary, "NoBlocking", Contradicted)
		}
		if wrapper.Kind == Native || len(wrapper.Effects) > 0 && wrapper.Effects[0].Kind == "io" {
			footprint := FailureFootprint{Locations: []FailureLocation{{Resource: resource}}, Outcomes: []Outcome{want}, Observation: "state_at_completion"}
			if AllEstablished(CheckFailureSummaryProfile(summary, Profile{Name: "PreservesOnFailure"}, footprint, nil)) {
				t.Fatal("checked wrapper retained stale helper final-state evidence")
			}
		}
	}
}

func TestNativeFailureCoverageAndRelationStability(t *testing.T) {
	resource := ResourceRef{ID: "state"}
	footprint := FailureFootprint{Locations: []FailureLocation{{Resource: resource}}, Outcomes: []Outcome{ReturnedFailure}, Observation: "state_at_completion"}
	failure := initialPath()
	failure.Outcome = ReturnedFailure
	failure.StateRelations = []StateRelation{{Location: footprint.Locations[0], Observation: footprint.Observation, Relation: "unchanged", Evidence: []string{"test-only:state-witness"}}}
	unknown := initialPath()
	unknown.Unknowns = []UnknownReason{{Code: "test-only:uncovered", Components: []string{"outcomes"}}}
	if AllEstablished(CheckFailureSummaryProfile(summarize([]PathSummary{failure, unknown}), Profile{Name: "PreservesOnFailure"}, footprint, nil)) {
		t.Fatal("known unchanged failure hid unknown failure outcomes on another path")
	}
	analysis := inference{options: AnalysisOptions{}.normalized()}
	analysis.observeEffect(&failure, Effect{Kind: "io", Resource: resource, Definite: true})
	if AllEstablished(CheckFailureSummaryProfile(summarize([]PathSummary{failure}), Profile{Name: "PreservesOnFailure"}, footprint, nil)) {
		t.Fatal("post-relation IO retained stale completion evidence")
	}
	clean := initialPath()
	clean.Outcome = ReturnedFailure
	footprint.Locations[0].Resource = ResourceRef{Unknown: true}
	if AllEstablished(CheckFailureSummaryProfile(summarize([]PathSummary{clean}), Profile{Name: "PreservesOnFailure"}, footprint, nil)) {
		t.Fatal("unknown protected resource obtained preservation")
	}
}

func nativeAssuranceOutcomes(declaration *DeclaredContract, outcomes ...Outcome) {
	declaration.Outcomes = outcomes
	var contracts []NativeOutcomeContract
	for _, outcome := range declaration.OutcomeContracts {
		if containsOutcome(outcomes, outcome.Outcome) {
			contracts = append(contracts, outcome)
		}
	}
	declaration.OutcomeContracts = contracts
}

// evaluateNativeUse keeps single-use fixture setup outside the production API.
func evaluateNativeUse(use NativeUse, declaration DeclaredContract, environment NativeEnvironment, policy TrustPolicy, witnesses []NativeWitness, applicability Status) NativeEvaluation {
	return EvaluateNativeUses([]NativeUse{use}, []DeclaredContract{declaration}, environment, policy, witnesses, applicability, nil)[0]
}

func TestNativeBatchRequiresSelectedDependencies(t *testing.T) {
	use, declaration, environment := nativeAssuranceFixture()
	declaration.Dependencies = []string{"required-helper"}
	policy, witnesses := nativeAssuranceWitnesses(use, declaration, environment, nil)
	known := LoweringStep{Meaning: "required-helper", Implementation: "helper-v1"}
	unknown := LoweringStep{Meaning: "required-helper"}
	for _, test := range []struct {
		name  string
		steps []LoweringStep
		want  Status
	}{{"missing", nil, Unresolved}, {"unknown", []LoweringStep{unknown}, Unresolved}, {"known-then-unknown", []LoweringStep{known, unknown}, Established}, {"unknown-then-known", []LoweringStep{unknown, known}, Established}} {
		t.Run(test.name, func(t *testing.T) {
			result := EvaluateNativeUses([]NativeUse{use}, []DeclaredContract{declaration}, environment, policy, witnesses, Established, test.steps)[0]
			if result.Applicability.Status != test.want || result.Components[0].Status != test.want {
				t.Fatalf("selected dependency applicability: %+v", result.Applicability)
			}
		})
	}
}

func TestNativeEvaluationChecksActualTargetType(t *testing.T) {
	use, declaration, environment := nativeAssuranceFixture()
	use.Target.Type = "selected-type"
	policy, witnesses := nativeAssuranceWitnesses(use, declaration, environment, nil)
	evaluation := evaluateNativeUse(use, declaration, environment, policy, witnesses, Established)
	node := nativeAssuranceNode(use)
	node.Target.Type = "different-typed-target"
	summary := reviewInfer(t, Plan{Entry: node}, AnalysisOptions{NativeApplications: map[string]NativeEvaluation{use.ID: evaluation}})
	if AllEstablished(CheckProfile(summary, Profile{Name: "NoBlocking"})) {
		t.Fatal("typed attachment mismatch reused native approval")
	}
}

func TestNativeIncomingFactsCannotOverrideCompilerAttachment(t *testing.T) {
	for _, field := range []string{"type", "meaning", "role"} {
		for _, shape := range []string{"direct", "nested-conjunction"} {
			for _, attachment := range []string{"unscoped", "exact-target"} {
				t.Run(field+"/"+shape+"/"+attachment, func(t *testing.T) {
					use, declaration, environment := nativeAssuranceFixture()
					use.Target.Type = "CompilerType"
					use.Target.Path = []string{"compiler-role"}
					var want Predicate
					switch field {
					case "type":
						want = Predicate{Kind: "type", Subject: "target", Value: Atom{Kind: "type", Text: "WrongType"}}
					case "meaning":
						want = Predicate{Kind: "meaning", Subject: "target", Value: Atom{Kind: "meaning", Text: "wrong-meaning"}}
					case "role":
						want = Predicate{Kind: "eq", Subject: "role", Value: Atom{Kind: "role", Text: "wrong-role"}}
					}
					if attachment == "exact-target" {
						want.Target = use.Target
					}
					declaration.Requirements = []Predicate{want}
					policy, witnesses := nativeAssuranceWitnesses(use, declaration, environment, nil)
					evaluation := evaluateNativeUse(use, declaration, environment, policy, witnesses, Established)
					incoming := want
					if shape == "nested-conjunction" {
						incoming = Predicate{Kind: "and", Args: []Predicate{{Kind: "true"}, {Kind: "and", Args: []Predicate{want}}}}
					}
					plan := Plan{Entry: Node{Kind: Sequence, Children: []Node{{Kind: Event, Facts: []Predicate{incoming}}, nativeAssuranceNode(use)}}}
					summary := reviewInfer(t, plan, AnalysisOptions{NativeApplications: map[string]NativeEvaluation{use.ID: evaluation}})
					if AllEstablished(CheckProfile(summary, Profile{Name: "NoBlocking"})) {
						t.Fatalf("incoming %s fact overrode compiler attachment and enabled approved summary: entry=%+v", field, summary.Obligations)
					}
				})
			}
		}
	}
}

func TestNativeIncomingTargetTypeAuthority(t *testing.T) {
	for _, scenario := range []string{"matching", "missing-compiler-type", "negated-compiler-type"} {
		t.Run(scenario, func(t *testing.T) {
			use, declaration, environment := nativeAssuranceFixture()
			use.Target.Type = "CompilerType"
			requirement := Predicate{Kind: "type", Subject: "target", Value: Atom{Kind: "type", Text: use.Target.Type}}
			declaration.Requirements = []Predicate{requirement}
			incoming := requirement
			want := Established
			switch scenario {
			case "missing-compiler-type":
				use.Target.Type = ""
				want = Unresolved
			case "negated-compiler-type":
				incoming = Predicate{Kind: "and", Args: []Predicate{{Kind: "not", Args: []Predicate{requirement}}}}
				want = Contradicted
			}
			policy, witnesses := nativeAssuranceWitnesses(use, declaration, environment, nil)
			evaluation := evaluateNativeUse(use, declaration, environment, policy, witnesses, Established)
			plan := Plan{Entry: Node{Kind: Sequence, Children: []Node{{Kind: Event, Facts: []Predicate{incoming}}, nativeAssuranceNode(use)}}}
			summary := reviewInfer(t, plan, AnalysisOptions{NativeApplications: map[string]NativeEvaluation{use.ID: evaluation}})
			found := false
			for _, obligation := range summary.Obligations {
				if obligation.ID == evaluation.Entry.ID {
					found = true
					if obligation.Status != want {
						t.Fatalf("entry=%+v want=%s", obligation, want)
					}
				}
			}
			if !found || AllEstablished(CheckProfile(summary, Profile{Name: "NoBlocking"})) != (want == Established) {
				t.Fatalf("type authority disagrees with inference: entry found=%v want=%s", found, want)
			}
		})
	}
}
