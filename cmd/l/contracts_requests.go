package main

import (
	"fmt"
	"sort"

	"github.com/l-programming-language/l/contracts"
	Lyger "github.com/l-programming-language/l/lyger"
)

func prepareNativeApplications(report *contracts.Report, metadata contractMetadata) (contracts.AnalysisOptions, []contracts.NativeEvaluation, error) {
	options := contracts.AnalysisOptions{NativeApplications: make(map[string]contracts.NativeEvaluation)}
	if metadata.Trust.Kind != "" {
		if err := metadata.Trust.Validate(); err != nil {
			return options, nil, err
		}
		report.TrustPolicy = &metadata.Trust.Policy
		report.NativeWitnesses = append([]contracts.NativeWitness(nil), metadata.Trust.Witnesses...)
	}
	environment := contracts.NativeEnvironment{}
	applicability, reason := contracts.Unresolved, "native environment evidence is missing; no external tools were executed"
	if metadata.Environment != nil {
		environment = *metadata.Environment
		report.NativeEnvironment = metadata.Environment
		applicability, reason = contracts.EnvironmentApplicability(environment, report.Inputs, report.Lowering.Fingerprints)
		configuration, message, err := validateNativeEnvironmentConfiguration(environment)
		if err != nil {
			return options, nil, err
		}
		if configuration != contracts.Established {
			applicability, reason = configuration, message
		}
	}
	report.Scope.Exclusions = append(report.Scope.Exclusions, reason)
	evaluations := contracts.EvaluateNativeUses(report.NativeUses, metadata.Native, environment, metadata.Trust.Policy, metadata.Trust.Witnesses, applicability, report.Lowering.Steps)
	for _, evaluation := range evaluations {
		if evaluation.Use.ID != "" {
			options.NativeApplications[evaluation.Use.ID] = evaluation
		}
	}
	return options, evaluations, nil
}

func applyContractRequests(report *contracts.Report, analysis *Lyger.ContractAnalysis, metadata contractMetadata, options contracts.AnalysisOptions) error {
	if metadata.Requests == nil {
		return nil
	}
	if err := metadata.Requests.Validate(); err != nil {
		return err
	}
	for _, request := range metadata.Requests.Checks {
		result := contracts.CheckResult{Request: request}
		profile := request.Profile
		profile.Version = contracts.SchemaVersion
		if request.Scope == "whole_entry" {
			if profile.Scope != "" && profile.Scope != "whole_region" {
				return fmt.Errorf("whole_entry requires the whole_region boundary")
			}
			profile.Scope = "whole_region"
			result.Summary = report.Summary
			result.Obligations = append(result.Obligations, report.Summary.Obligations...)
		} else {
			scope, err := analysis.FindContractScope(request.Scope)
			if err != nil {
				return err
			}
			if scope.Phase != contracts.Invocation || scope.NodeID == "" {
				return fmt.Errorf("unsupported execution boundary %s", request.Scope)
			}
			if profile.Scope != "" && profile.Scope != scope.Boundary {
				return fmt.Errorf("requested scope label differs from selected execution boundary")
			}
			if profile.Target.Root != "" && contracts.TargetRelation(profile.Target, scope.Target) != contracts.Established {
				return fmt.Errorf("requested target differs from selected execution boundary")
			}
			profile.Scope, profile.Target = scope.Boundary, scope.Target
			result.Scope = &scope
			result.Summary, err = contracts.InferScope(analysis.Plan, scope.NodeID, options)
			if err != nil {
				return err
			}
			result.Obligations = append(result.Obligations, result.Summary.Obligations...)
			if !scope.Complete || len(scope.Unknowns) > 0 {
				result.Obligations = append(result.Obligations, contracts.Obligation{ID: contracts.StableID("scope-coverage", scope), Status: contracts.Unresolved, Reason: "selected execution boundary has incomplete coverage"})
			}
		}
		if request.Footprint != nil {
			if err := validateRequestFootprint(*request.Footprint, result.Scope, report.Scopes); err != nil {
				return err
			}
		}
		result.Obligations = append(result.Obligations, checkContractProfile(analysis, result.Summary, profile, request.Footprint)...)
		result.Status = contracts.ObligationStatus(result.Obligations)
		for _, obligation := range result.Obligations {
			result.Required = append(result.Required, obligation.ID)
		}
		report.Requests = append(report.Requests, result)
		report.Assumptions = append(report.Assumptions, result.Summary.Assumptions...)
	}
	sort.Strings(report.Assumptions)
	return nil
}

func checkContractProfile(analysis *Lyger.ContractAnalysis, summary contracts.DerivedSummary, profile contracts.Profile, footprint *contracts.FailureFootprint) []contracts.Obligation {
	descriptor, _ := contracts.LookupProfile(profile.Name)
	if descriptor.WholeEntryOnly {
		return analysis.CheckLifecycleProfile(profile)
	}
	if descriptor.RequiresFootprint {
		if footprint == nil {
			footprint = &contracts.FailureFootprint{}
		}
		return contracts.CheckFailureSummaryProfile(summary, profile, *footprint, nil)
	}
	return contracts.CheckProfile(summary, profile)
}

func validateRequestFootprint(footprint contracts.FailureFootprint, selected *contracts.AnalysisScope, all []contracts.AnalysisScope) error {
	scopes := all
	if selected != nil {
		scopes = []contracts.AnalysisScope{*selected}
	}
	for _, location := range footprint.Locations {
		found := false
		for _, scope := range scopes {
			for _, available := range scope.Resources {
				// Only published projections have a supported observation model.
				found = found || contracts.StableID("location", location) == contracts.StableID("location", available)
			}
		}
		if !found {
			return fmt.Errorf("failure footprint is not a published modeled resource/projection in the selected scope: %s", location.Resource.ID)
		}
	}
	return nil
}
