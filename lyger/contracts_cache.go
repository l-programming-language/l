package Lyger

import (
	"fmt"

	"github.com/l-programming-language/l/contracts"
)

func inferContractAnalysis(analysis *ContractAnalysis, options ContractAnalysisOptions) error {
	var summary contracts.DerivedSummary
	var err error
	if options.Session == nil {
		summary, err = contracts.Infer(analysis.Plan, options.Inference)
	} else {
		// Track A always resolves first. A cache entry only checks this resolution.
		if err := cacheContractBindings(analysis.Bindings, options.Session); err != nil {
			return err
		}
		summary, err = options.Session.Infer(analysis.Plan, options.Inference, analysis.Bindings.Inputs, analysis.Lowering)
	}
	if err == nil {
		analysis.Summary = summary
	}
	return err
}

func cacheContractBindings(snapshot ContractBindings, session *contracts.Session) error {
	source := contracts.StableID("source-inputs", snapshot.Inputs)
	definitions := contracts.StableID("source-definitions", snapshot.Definitions)
	for _, binding := range snapshot.Bindings {
		if binding.Target.Unknown || binding.Target.Root == "" {
			// Unknown attachment is not a cacheable precise target.
			continue
		}
		key := contracts.BindingCacheKey{
			Schema: contracts.SchemaVersion, SourceStructure: source,
			Environment: contracts.BindingEnvironmentKey(contracts.StableID("binding-environments", []contracts.BindingEnvironmentKey{binding.DefinitionEnvironment, binding.UseEnvironment})),
			Candidates: contracts.StableID("current-candidates", struct {
				Definitions string
				Candidate   string
			}{definitions, binding.CandidateFingerprint}),
			Applicability: contracts.StableID("resolved-applicability", binding), Target: binding.Target,
			LexicalRole: binding.LexicalRole, PhrasePolicy: "existing-source-parser/1", ResolutionPolicy: snapshot.Policy,
		}
		if err := session.RevalidateBinding(key, binding); err != nil {
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
