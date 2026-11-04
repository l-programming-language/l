package Lyger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/l-programming-language/l/contracts"
	"github.com/l-programming-language/l/lerrors"
)

type DictionaryContractAnalysis struct {
	Analysis             *ContractAnalysis
	Provider             string
	Source               string
	Definitions          []contracts.DefinitionRecord
	Certificates         []contracts.AdjectiveCertificate
	Obligations          []contracts.Obligation
	SourceProjections    []contracts.SourceAdjectiveProjection
	SourceCertificates   []contracts.SourceAdjectiveCertificate
	SourceInstantiations []contracts.SourceAdjectiveInstantiation
	Verification         *contracts.SourceVerificationResult
}

type DictionaryVerificationOptions struct {
	Analysis  ContractAnalysisOptions
	Meanings  []contracts.MeaningKey
	Assurance string
}

// VerifyDictionaryContractsWithOptions adds explicitly selected structural
// assurance. The original API continues to return full discovery and unresolved
// implementation conformance; symbolic selection never suppresses those records.
func (ls *LygerService) VerifyDictionaryContractsWithOptions(input string, options DictionaryVerificationOptions) (*DictionaryContractAnalysis, error) {
	assurance := options.Assurance
	if assurance == "" {
		assurance = "implementation"
	}
	if assurance != "implementation" && assurance != "symbolic" {
		return nil, fmt.Errorf("%w: unsupported assurance %q", ErrSourceContractMetadata, assurance)
	}
	if assurance == "symbolic" && len(options.Meanings) == 0 {
		return nil, fmt.Errorf("%w: symbolic assurance requires selected canonical descriptor meanings", ErrSourceContractMetadata)
	}
	result, err := ls.VerifyDictionaryContracts(input, options.Analysis)
	if err != nil {
		return nil, err
	}
	verification := &contracts.SourceVerificationResult{Assurance: assurance, Scope: "whole_dictionary", Status: contracts.Established}
	selected := make(map[contracts.MeaningKey]bool)
	for _, meaning := range options.Meanings {
		if selected[meaning] {
			continue
		}
		count := 0
		for _, definition := range result.Definitions {
			if definition.Meaning == meaning && definition.Descriptor {
				count++
			}
		}
		if count != 1 {
			return nil, fmt.Errorf("%w: selected meaning %s is absent, ambiguous or not a descriptor in this dictionary", ErrSourceContractMetadata, meaning)
		}
		selected[meaning] = true
		verification.Selected = append(verification.Selected, meaning)
	}
	if assurance == "symbolic" {
		verification.Scope = "source_elaboration"
		for _, meaning := range verification.Selected {
			found := false
			for _, certificate := range result.SourceCertificates {
				if certificate.Meaning == meaning {
					verification.Obligations = append(verification.Obligations, certificate.LogicalObligations...)
					found = true
				}
			}
			if !found {
				verification.Obligations = append(verification.Obligations, contracts.Obligation{ID: contracts.StableID("source-projection-missing", meaning), Status: contracts.Unresolved, Reason: "selected descriptor has no supported source projection"})
			}
		}
	} else {
		verification.Obligations = append(verification.Obligations, result.Obligations...)
		for _, certificate := range result.SourceCertificates {
			verification.Obligations = append(verification.Obligations, certificate.LogicalObligations...)
			verification.Obligations = append(verification.Obligations, certificate.LoweringObligations...)
		}
		for _, boundary := range result.Analysis.Summary.NativeBoundaries {
			verification.Obligations = append(verification.Obligations, contracts.Obligation{ID: contracts.StableID("dictionary-native", boundary), Claim: contracts.Claim{Target: boundary.Target, Scope: "implementation", Predicate: contracts.Predicate{Kind: "native_conformance", Subject: boundary.ID}}, Status: boundary.Status, Reason: boundary.Reason, Origin: boundary.Origin})
		}
	}
	for _, obligation := range verification.Obligations {
		if obligation.Status == contracts.Contradicted {
			verification.Status = contracts.Contradicted
			break
		}
		if obligation.Status != contracts.Established {
			verification.Status = contracts.Unresolved
		}
	}
	result.Verification = verification
	return result, nil
}

// VerifyDictionaryContracts inspects the existing dictionary provider and its
// declaration closure. Source declarations without a supported symbolic body
// adapter remain explicitly unverified; loading them is never certification.
// Indexed dictionaries retain their private definition-site environment.
func (ls *LygerService) VerifyDictionaryContracts(input string, options ContractAnalysisOptions) (*DictionaryContractAnalysis, error) {
	var source []byte
	var sourcePath, selectedSource, provider string
	if strings.HasSuffix(input, ".l") {
		absolute, err := filepath.Abs(input)
		if err != nil {
			return nil, lerrors.IO.Wrap(err, "dictionary path")
		}
		data, err := os.ReadFile(absolute)
		if err != nil {
			return nil, lerrors.IO.Wrap(err, "read dictionary source")
		}
		source, sourcePath, provider = data, absolute, input
		selectedSource = "source:" + filepath.Base(absolute)
	} else {
		index, err := readDictionaryIndex(ls.options.Resources)
		if err != nil {
			return nil, err
		}
		entry, ok := index.Languages[ls.options.Language][input]
		if !ok {
			return nil, lerrors.UnresolvedReference.Errorf("dictionary %q is not indexed for %s", input, ls.options.Language)
		}
		provider = input
		selectedSource = "resource:" + entry.Entry
		// Use the language's existing dictionary import, not a new import mode.
		source = []byte(fmt.Sprintf("Use %s from `%s`.\n", entry.Dictionary, strings.ReplaceAll(input, "`", "\\`")))
		sourcePath = filepath.Join(string(filepath.Separator), "contracts-inspection", "dictionary.l")
	}
	analysis, err := ls.AnalyzeContracts(source, sourcePath, options)
	if err != nil {
		return nil, err
	}
	result := &DictionaryContractAnalysis{Analysis: analysis, Provider: provider, Source: selectedSource}
	selected := make(map[contracts.MeaningKey]bool)
	for _, definition := range analysis.Bindings.Definitions {
		if definition.Origin.Source == selectedSource {
			selected[definition.Meaning] = true
		}
	}
	for _, projection := range analysis.SourceProjections {
		if selected[projection.Meaning] {
			result.SourceProjections = append(result.SourceProjections, projection)
		}
	}
	for _, certificate := range analysis.SourceCertificates {
		if selected[certificate.Meaning] {
			result.SourceCertificates = append(result.SourceCertificates, certificate)
		}
	}
	for _, instance := range analysis.SourceInstantiations {
		if selected[instance.Meaning] {
			result.SourceInstantiations = append(result.SourceInstantiations, instance)
		}
	}
	for _, definition := range analysis.Bindings.Definitions {
		if definition.Origin.Source != selectedSource {
			continue
		}
		result.Definitions = append(result.Definitions, definition)
		claim := contracts.Claim{ID: contracts.StableID("dictionary-verification", definition.Meaning), Predicate: contracts.Predicate{Kind: "definition_conformance", Subject: string(definition.Meaning)}, Target: contracts.TargetRef{Root: string(definition.Definition), Type: definition.Applicability, Phase: contracts.Elaboration}, Scope: "definition", Phase: contracts.Elaboration}
		obligation := contracts.Obligation{ID: claim.ID, Claim: claim, Status: contracts.Unresolved, Reason: "canonical source declaration discovered; no supported source-body symbolic certificate or approved native conformance witness is available", Origin: definition.Origin, Dependencies: []string{string(definition.Environment), string(definition.Meaning)}}
		for _, certificate := range result.SourceCertificates {
			if certificate.Meaning == definition.Meaning {
				obligation.Reason = "source elaboration certificate reported separately; executable/native conformance remains unresolved"
				obligation.Claim.Scope = "implementation"
				break
			}
		}
		result.Obligations = append(result.Obligations, obligation)
		if definition.Descriptor {
			result.Certificates = append(result.Certificates, contracts.AdjectiveCertificate{Meaning: definition.Meaning, Status: contracts.Unresolved, Mode: "instantiation_only", TransformerFingerprint: contracts.StableID("source-descriptor", definition), Summary: contracts.DerivedSummary{Coverage: contracts.Unknown, Resources: contracts.UnknownResourceCost("source descriptor requires instantiation analysis")}, Obligations: []contracts.Obligation{obligation}})
		}
	}
	if len(result.Definitions) == 0 {
		return nil, lerrors.UnresolvedReference.Errorf("dictionary %q produced no canonical declarations at %s", input, selectedSource)
	}
	return result, nil
}
