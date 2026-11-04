package contracts

import "testing"

func nativeAssuranceFixture() (NativeUse, DeclaredContract, NativeEnvironment) {
	unknown := true
	signature := &NativeSignature{Inputs: []string{"value"}, Outputs: []string{"Number"}, InputTypes: []NativeType{{Kind: "number"}}, OutputTypes: []NativeType{{Kind: "number"}}}
	use := NativeUse{ID: "test-only:leaf-use", Meaning: "test-only:leaf", Name: "Leaf", Target: TargetRef{Root: "test-only:target", Phase: Invocation}, Phase: Invocation, Signature: signature, Implementation: "test-only:body-v1", Dependencies: map[string]string{"helper": "helper-v1"}, OutcomeDisposition: "propagate_returned_failure"}
	declaration := DeclaredContract{ID: "test-only:leaf-contract", Meaning: use.Meaning, Name: "Leaf", Signature: signature, Outcomes: []Outcome{Normal, ReturnedFailure}, OutcomeContracts: []NativeOutcomeContract{
		{Outcome: Normal, Effects: &DeclaredEffectRow{UnknownTail: &unknown}, AbsentEffects: []EffectKind{"block"}},
		{Outcome: ReturnedFailure, Effects: &DeclaredEffectRow{UnknownTail: &unknown}, AbsentEffects: []EffectKind{"block"}},
	}}
	environment := SealNativeEnvironment(NativeEnvironment{SchemaVersion: SchemaVersion, Kind: "native_environment", Producer: AnalyzerVersion, Version: "1", Method: "test-only", Host: "test-only", Target: "test-only-target", Compiler: "test-only-compiler", PanicPolicy: "test-only-unwind", Runtime: "test-only-runtime", Allocator: "test-only-allocator", Inputs: map[string]string{"source": "source-v1"}, Dependencies: map[string]string{"runtime": "runtime-v1"}, Limitations: []string{"Synthetic evidence; no production approval or native execution"}})
	return use, declaration, environment
}

func nativeAssuranceWitnesses(use NativeUse, declaration DeclaredContract, environment NativeEnvironment, selected func(NativeComponent) bool) (TrustPolicy, []NativeWitness) {
	policy := TrustPolicy{Version: SchemaVersion, ID: "test-only:assurance-policy", Authorities: []string{"test-only:authority"}, AllowTestOnly: true}
	var witnesses []NativeWitness
	for _, component := range NativeComponentClaims(use, declaration, environment) {
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
		application, _ := ApproveNativeApplication(use, declaration, environment, policy, witnesses, Established)
		summary := reviewInfer(t, Plan{Entry: nativeAssuranceNode(use)}, AnalysisOptions{NativeApplications: map[string]ApprovedNativeApplication{use.ID: application}})
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
	for _, mutation := range []string{"signature", "environment", "unsealed-environment", "implementation", "helper", "missing-witness", "stale-witness", "test-policy", "applicability"} {
		t.Run(mutation, func(t *testing.T) {
			use, declaration, environment := nativeAssuranceFixture()
			policy, witnesses := nativeAssuranceWitnesses(use, declaration, environment, nil)
			applicability := Established
			switch mutation {
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
			application, components := ApproveNativeApplication(use, declaration, environment, policy, witnesses, applicability)
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
		application, _ := ApproveNativeApplication(use, declaration, environment, policy, witnesses, Established)
		plan := Plan{Entry: Node{Kind: Sequence, Children: []Node{nativeAssuranceNode(use), {Kind: Event, Effects: []Effect{{Kind: "allocate", Definite: true}}}}}}
		summary := reviewInfer(t, plan, AnalysisOptions{NativeApplications: map[string]ApprovedNativeApplication{use.ID: application}})
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
		application, _ := ApproveNativeApplication(use, declaration, environment, policy, witnesses, Established)
		nodes := []Node{nativeAssuranceNode(use), {Kind: Event, Facts: []Predicate{fact}}}
		if before {
			nodes[0], nodes[1] = nodes[1], nodes[0]
		}
		summary := reviewInfer(t, Plan{Entry: Node{Kind: Sequence, Children: nodes}}, AnalysisOptions{NativeApplications: map[string]ApprovedNativeApplication{use.ID: application}})
		if AllEstablished(CheckProfile(summary, Profile{Name: "NoBlocking"})) != before {
			t.Fatalf("before=%t entry requirement mischecked: %+v", before, summary.Obligations)
		}
	}
}

func TestNativeAssuranceUnknownHelperDoesNotInheritExclusion(t *testing.T) {
	use, declaration, environment := nativeAssuranceFixture()
	nativeAssuranceOutcomes(&declaration, Normal)
	policy, witnesses := nativeAssuranceWitnesses(use, declaration, environment, nil)
	application, _ := ApproveNativeApplication(use, declaration, environment, policy, witnesses, Established)
	summary := reviewInfer(t, Plan{Entry: Node{Kind: Sequence, Children: []Node{nativeAssuranceNode(use), {Kind: Native, Native: "independent-opaque-helper"}}}}, AnalysisOptions{NativeApplications: map[string]ApprovedNativeApplication{use.ID: application}})
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
	application, _ := ApproveNativeApplication(use, declaration, environment, policy, witnesses, Established)
	for _, wrapper := range []Node{
		{Kind: Event, Effects: []Effect{{Kind: "block", Definite: true}}},
		{Kind: Event, Effects: []Effect{{Kind: "io", Resource: resource, Definite: true}}},
		{Kind: Event, Outcome: Panic},
		{Kind: Event, Outcome: Panic, Effects: []Effect{{Kind: "io", Resource: resource, Definite: true}}},
		{Kind: Native, Native: "opaque-checked-wrapper"},
	} {
		plan := Plan{Entry: Node{Kind: Sequence, Children: []Node{{Kind: CheckedResult, Children: []Node{nativeAssuranceNode(use)}, Cleanup: []Node{wrapper}}, {Kind: Event, Effects: []Effect{{Kind: "allocate", Definite: true}}}}}}
		summary := reviewInfer(t, plan, AnalysisOptions{NativeApplications: map[string]ApprovedNativeApplication{use.ID: application}})
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
