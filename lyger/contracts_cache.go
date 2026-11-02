package Lyger

import (
	"fmt"

	"github.com/l-programming-language/l/contracts"
)

func inferContractAnalysis(analysis *ContractAnalysis, options ContractAnalysisOptions) error {
	if options.Session == nil {
		return inferContractAnalysisCold(analysis, options.Inference)
	}
	// Track A resolution always runs against current source/providers. A cache
	// snapshot never supplies record pointers or changes the resolver's choice.
	// Replayed binding records must agree with that newly resolved snapshot.
	if err := cacheContractBindings(analysis.Bindings, options.Session); err != nil {
		return err
	}
	logicalOptions := options.Inference
	logicalOptions.NativeSummaries = nil
	logicalOptions.NativeWitnesses = nil
	logicalOptions.ApprovedAssumptions = nil
	logical := contracts.LogicalCacheKey{
		Schema: contracts.SchemaVersion, Meaning: contracts.MeaningKey(contracts.StableID("analysis-entry", analysis.Plan.Entry)),
		Specification: contracts.StableID("typed-plan", analysis.Plan), Substitutions: contracts.StableID("analysis-facts", logicalOptions),
		FacetSchemas: "source-facets:not-applied", RoleInterfaces: "source-roles:not-applied", Rules: contracts.AnalyzerVersion,
		TargetModel: "l-typed-plan/1", Dependencies: map[string]string{"native-specifications": contracts.StableID("specifications", options.Inference.NativeSummaries)},
	}
	// The reusable logical result has no native trust. It is the conditional
	// source analysis with explicit opaque obligations, not concrete assurance.
	var baseline contracts.DerivedSummary
	found, err := options.Session.LoadLogical(logical, &baseline)
	if err != nil {
		return err
	}
	if !found {
		baseline, err = contracts.Infer(analysis.Plan, logicalOptions)
		if err != nil {
			return err
		}
		if err := options.Session.StoreLogical(logical, baseline); err != nil {
			return err
		}
	}
	dependencies := make(map[string]string)
	for name, fingerprint := range analysis.Bindings.Inputs {
		dependencies["input:"+name] = fingerprint
	}
	for name, fingerprint := range analysis.Lowering.Fingerprints {
		dependencies["implementation:"+name] = fingerprint
	}
	implementation := contracts.ImplementationCacheKey{
		Schema: contracts.SchemaVersion, Logical: contracts.StableID("logical-key", logical),
		Implementation: contracts.StableID("lowering", analysis.Lowering), Target: contracts.StableID("selected-target-inputs", analysis.Bindings.Inputs),
		TrustPolicy: contracts.StableID("approved-assumptions", options.Inference.ApprovedAssumptions), Witnesses: contracts.StableID("witnesses", options.Inference.NativeWitnesses), Dependencies: dependencies,
	}
	var summary contracts.DerivedSummary
	if found, err := options.Session.LoadImplementation(implementation, &summary); err != nil {
		return err
	} else if found {
		analysis.Summary = summary
		return nil
	}
	if options.Inference.NativeSummaries == nil && options.Inference.NativeWitnesses == nil && options.Inference.ApprovedAssumptions == nil {
		summary = baseline
	} else {
		summary, err = contracts.Infer(analysis.Plan, options.Inference)
		if err != nil {
			return err
		}
	}
	if err := options.Session.StoreImplementation(implementation, summary); err != nil {
		return err
	}
	analysis.Summary = summary
	return nil
}

func inferContractAnalysisCold(analysis *ContractAnalysis, options contracts.AnalysisOptions) error {
	summary, err := contracts.Infer(analysis.Plan, options)
	if err != nil {
		return err
	}
	analysis.Summary = summary
	return nil
}

func cacheContractBindings(snapshot ContractBindings, session *contracts.Session) error {
	for _, binding := range snapshot.Bindings {
		if binding.Target.Unknown || binding.Target.Root == "" {
			// Unknown attachment is not a cacheable precise target.
			continue
		}
		key := contracts.BindingCacheKey{
			Schema: contracts.SchemaVersion, SourceStructure: contracts.StableID("source-inputs", snapshot.Inputs),
			Environment: contracts.BindingEnvironmentKey(contracts.StableID("binding-environments", []contracts.BindingEnvironmentKey{binding.DefinitionEnvironment, binding.UseEnvironment})),
			Candidates: contracts.StableID("current-candidates", struct {
				Definitions []ContractDefinition
				Candidate   string
			}{snapshot.Definitions, binding.CandidateFingerprint}),
			Applicability: contracts.StableID("resolved-applicability", binding), Target: binding.Target,
			LexicalRole: binding.LexicalRole, PhrasePolicy: "existing-source-parser/1", ResolutionPolicy: snapshot.Policy,
		}
		var cached contracts.ResolvedBinding
		found, err := session.LoadBinding(key, &cached)
		if err != nil {
			return err
		}
		if found {
			if contracts.StableID("binding", cached) != contracts.StableID("binding", binding) {
				return fmt.Errorf("binding cache disagrees with current Track A resolution at %s", binding.Occurrence)
			}
			continue
		}
		if err := session.StoreBinding(key, binding); err != nil {
			return err
		}
	}
	return nil
}

// ReinferContracts revalidates the three cache boundaries after the coordinator
// has refreshed the typed plan and selected backend inventory.
func (analysis *ContractAnalysis) ReinferContracts(options ContractAnalysisOptions) error {
	if analysis == nil {
		return fmt.Errorf("contract analysis is nil")
	}
	if err := inferContractAnalysis(analysis, options); err != nil {
		return err
	}
	if options.SourceTransformers != nil {
		return analysis.verifySourceContracts(options)
	}
	return nil
}
