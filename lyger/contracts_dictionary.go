package Lyger

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
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
	counts := make(map[contracts.MeaningKey]int)
	for _, def := range result.Definitions {
		if def.Descriptor {
			counts[def.Meaning]++
		}
	}
	var selected []contracts.MeaningKey
	for _, meaning := range options.Meanings {
		if slices.Contains(selected, meaning) {
			continue
		}
		if counts[meaning] != 1 {
			return nil, fmt.Errorf("%w: selected meaning %s is absent, ambiguous or not a descriptor in this dictionary", ErrSourceContractMetadata, meaning)
		}
		selected = append(selected, meaning)
	}
	report := contracts.Report{SourceCertificates: result.SourceCertificates}
	scope := selected
	if assurance == "implementation" {
		scope = nil // This API covers the whole frontend inventory, including native residuals.
		report.Obligations = append(report.Obligations, result.Obligations...)
		for _, boundary := range result.Analysis.Summary.NativeBoundaries {
			report.Obligations = append(report.Obligations, contracts.Obligation{ID: contracts.StableID("dictionary-native", boundary), Claim: contracts.Claim{Target: boundary.Target, Scope: "implementation", Predicate: contracts.Predicate{Kind: "native_conformance", Subject: boundary.ID}}, Status: boundary.Status, Reason: boundary.Reason, Origin: boundary.Origin})
		}
	}
	verification, missing, err := report.DictionaryVerification(assurance, scope)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSourceContractMetadata, err)
	}
	verification.Selected = selected
	for _, meaning := range missing {
		verification.Obligations = append(verification.Obligations, contracts.Obligation{ID: contracts.StableID("source-projection-missing", meaning), Status: contracts.Unresolved, Reason: "selected descriptor has no supported source projection"})
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
	var sourcePath, selectedSource string
	if strings.HasSuffix(input, ".l") {
		absolute, err := filepath.Abs(input)
		if err != nil {
			return nil, lerrors.IO.Wrap(err, "dictionary path")
		}
		data, err := os.ReadFile(absolute)
		if err != nil {
			return nil, lerrors.IO.Wrap(err, "read dictionary source")
		}
		source, sourcePath = data, absolute
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
		selectedSource = "resource:" + entry.Entry
		// Use the language's existing dictionary import, not a new import mode.
		source = []byte(fmt.Sprintf("Use %s from `%s`.\n", entry.Dictionary, strings.ReplaceAll(input, "`", "\\`")))
		sourcePath = filepath.Join(string(filepath.Separator), "contracts-inspection", "dictionary.l")
	}
	analysis, err := ls.AnalyzeContracts(source, sourcePath, options)
	if err != nil {
		return nil, err
	}
	result := &DictionaryContractAnalysis{Analysis: analysis, Provider: input, Source: selectedSource}
	selected := make(map[contracts.MeaningKey]bool)
	for _, definition := range analysis.Bindings.Definitions {
		if definition.Origin.Source == selectedSource {
			selected[definition.Meaning] = true
			result.Definitions = append(result.Definitions, definition)
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
	for _, definition := range result.Definitions {
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
