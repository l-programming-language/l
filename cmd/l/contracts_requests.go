package main

import (
	"fmt"
	"sort"

	"github.com/l-programming-language/l/contracts"
	Lyger "github.com/l-programming-language/l/lyger"
)

func prepareNativeApplications(report *contracts.Report, metadata contractMetadata) (contracts.AnalysisOptions, error) {
	options := contracts.AnalysisOptions{NativeApplications: make(map[string]contracts.ApprovedNativeApplication)}
	environment := contracts.NativeEnvironment{}
	applicability, reason := contracts.Unresolved, "native environment evidence is missing; no external tools were executed"
	if metadata.Environment != nil {
		environment = *metadata.Environment
		report.NativeEnvironment = metadata.Environment
		applicability, reason = contracts.EnvironmentApplicability(environment, report.Inputs, report.Lowering.Fingerprints)
		configuration, message, err := validateNativeEnvironmentConfiguration(environment)
		if err != nil {
			return options, err
		}
		if configuration != contracts.Established {
			applicability, reason = configuration, message
		}
	}
	report.Scope.Exclusions = append(report.Scope.Exclusions, reason)
	for _, use := range report.NativeUses {
		for _, declaration := range metadata.Native {
			if use.Meaning != declaration.Meaning {
				continue
			}
			application, components := contracts.ApproveNativeApplication(use, declaration, environment, metadata.Trust.Policy, metadata.Trust.Witnesses, applicability)
			options.NativeApplications[use.ID] = application
			report.NativeComponents = append(report.NativeComponents, components...)
			for _, component := range components {
				report.Evidence.Claims = append(report.Evidence.Claims, component.Claim)
				report.Evidence.Nodes = append(report.Evidence.Nodes, contracts.EvidenceNode{ID: component.Claim.ID, Claim: component.Claim.ID, Kind: "native_assumption", Witness: component.Witness, Fingerprints: component.Current, Origin: use.Origin})
				report.Obligations = append(report.Obligations, contracts.Obligation{ID: component.Claim.ID, Claim: component.Claim, Status: component.Status, Reason: component.Reason, Evidence: []string{component.Claim.ID}, Dependencies: []string{contracts.StableID("applicability", component.Current)}, Origin: use.Origin})
			}
		}
	}
	return options, nil
}

func resultStatus(obligations []contracts.Obligation) contracts.Status {
	status := contracts.Established
	for _, obligation := range obligations {
		if obligation.Status == contracts.Contradicted {
			return contracts.Contradicted
		}
		if obligation.Status != contracts.Established {
			status = contracts.Unresolved
		}
	}
	return status
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
			result.Obligations = append(result.Obligations, contracts.CheckFailureSummaryProfile(result.Summary, profile, *request.Footprint, nil)...)
		} else if profile.Name == "OwnedLifecycle" {
			if result.Scope != nil {
				return fmt.Errorf("OwnedLifecycle requires whole_entry; invocation lifecycle projection is unsupported")
			}
			result.Obligations = append(result.Obligations, analysis.CheckLifecycleProfile(profile)...)
		} else {
			result.Obligations = append(result.Obligations, contracts.CheckProfile(result.Summary, profile)...)
		}
		result.Status = resultStatus(result.Obligations)
		for _, obligation := range result.Obligations {
			result.Required = append(result.Required, obligation.ID)
		}
		report.Requests = append(report.Requests, result)
		report.Assumptions = append(report.Assumptions, result.Summary.Assumptions...)
	}
	sort.Strings(report.Assumptions)
	return nil
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
