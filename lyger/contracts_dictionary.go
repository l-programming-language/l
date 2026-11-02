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
	Analysis     *ContractAnalysis
	Provider     string
	Source       string
	Definitions  []contracts.DefinitionRecord
	Certificates []contracts.AdjectiveCertificate
	Obligations  []contracts.Obligation
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
	for _, definition := range analysis.Bindings.Definitions {
		if definition.Origin.Source != selectedSource {
			continue
		}
		result.Definitions = append(result.Definitions, definition)
		claim := contracts.Claim{ID: contracts.StableID("dictionary-verification", definition.Meaning), Predicate: contracts.Predicate{Kind: "definition_conformance", Subject: string(definition.Meaning)}, Target: contracts.TargetRef{Root: string(definition.Definition), Type: definition.Applicability, Phase: contracts.Elaboration}, Scope: "definition", Phase: contracts.Elaboration}
		obligation := contracts.Obligation{ID: claim.ID, Claim: claim, Status: contracts.Unresolved, Reason: "canonical source declaration discovered; no supported source-body symbolic certificate or approved native conformance witness is available", Origin: definition.Origin, Dependencies: []string{string(definition.Environment), string(definition.Meaning)}}
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
