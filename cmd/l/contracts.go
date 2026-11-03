package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	distribution "github.com/l-programming-language/l"
	"github.com/l-programming-language/l/contracts"
	Logiceme "github.com/l-programming-language/l/logiceme"
	Lyger "github.com/l-programming-language/l/lyger"
)

const contractsUsage = `Usage: l contracts inspect [--format text|json] [--output PATH] INPUT
       l contracts check --profile NAME [--profile NAME ...] [--report PATH] INPUT
       l contracts check --request-file PATH [--native-environment PATH] [--report PATH] INPUT
       l contracts evidence capture --output PATH [--target HOST_TRIPLE] INPUT
       l contracts explain --binding OCCURRENCE_OR_SOURCE:BYTE_OFFSET INPUT
       l contracts verify-dictionary [--report PATH] DICTIONARY
       l contracts verify-dictionary --assurance symbolic --adjective MEANING_OR_SOURCE:BYTE_OFFSET [--adjective ...] DICTIONARY
Options precede input. Common options: --assets ROOT, --language eng,
--manifest PATH (repeatable), --trust-policy PATH, --native-environment PATH,
--format text|json. Inspection and checking never execute external tools.
Explicit evidence capture queries installed tools; it never builds or runs Apps.
Track A preserves current resolution. Track B has no source-facing mode.
Assurance defaults to implementation. Symbolic assurance checks only selected
source adjective definitions; native/runtime assurance remains separate.
`

type contractsCommand struct {
	Action, Input, Format, Output, Report, Binding, Assets, Language, Trust, Assurance string
	Profiles, Manifests, Adjectives                                                    []string
	NativeEnvironment, RequestFile, Target                                             string
}

func parseContractsCommand(args []string) (contractsCommand, error) {
	command := contractsCommand{Format: "text", Language: Lyger.DefaultLanguage, Assurance: "implementation"}
	if len(args) == 0 {
		return command, fmt.Errorf("contracts requires a subcommand")
	}
	if args[0] == "--help" || args[0] == "-h" {
		return command, flag.ErrHelp
	}
	command.Action = args[0]
	if command.Action == "evidence" {
		if len(args) < 2 || args[1] != "capture" {
			return command, fmt.Errorf("evidence requires capture")
		}
		command.Action = "capture"
		args = append([]string{"capture"}, args[2:]...)
	}
	switch command.Action {
	case "inspect", "check", "explain", "verify-dictionary", "capture":
	default:
		return command, fmt.Errorf("unknown contracts command %q", command.Action)
	}
	flags := flag.NewFlagSet("l contracts "+command.Action, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&command.Format, "format", "text", "report format")
	flags.StringVar(&command.Output, "output", "", "inspection output")
	flags.StringVar(&command.Report, "report", "", "JSON report path")
	flags.StringVar(&command.Binding, "binding", "", "binding identity or source selector")
	flags.StringVar(&command.Assets, "assets", "", "complete resource filesystem")
	flags.StringVar(&command.Language, "language", Lyger.DefaultLanguage, "source language")
	flags.StringVar(&command.Trust, "trust-policy", "", "explicit trust policy document")
	flags.StringVar(&command.NativeEnvironment, "native-environment", "", "explicit reusable native environment evidence")
	flags.StringVar(&command.RequestFile, "request-file", "", "versioned scoped contract check requests")
	flags.StringVar(&command.Target, "target", "", "capture host target compatibility assertion")
	flags.StringVar(&command.Assurance, "assurance", "implementation", "implementation or selected symbolic assurance")
	flags.Func("adjective", "source adjective meaning or declaration source selector", func(value string) error {
		if value == "" {
			return fmt.Errorf("empty adjective selector")
		}
		command.Adjectives = append(command.Adjectives, value)
		return nil
	})
	flags.Func("manifest", "additional metadata", func(value string) error {
		if value == "" {
			return fmt.Errorf("empty manifest path")
		}
		command.Manifests = append(command.Manifests, value)
		return nil
	})
	flags.Func("profile", "required profile", func(value string) error {
		if !contracts.SupportedProfile(value) && value != "OwnedLifecycle" && value != "PreservesOnFailure" {
			return fmt.Errorf("unsupported profile %q", value)
		}
		command.Profiles = append(command.Profiles, value)
		return nil
	})
	if err := flags.Parse(args[1:]); err != nil {
		return command, err
	}
	if len(flags.Args()) != 1 || flags.Args()[0] == "" {
		return command, fmt.Errorf("contracts requires one input; options must precede it")
	}
	command.Input = flags.Args()[0]
	if command.Format != "text" && command.Format != "json" {
		return command, fmt.Errorf("unsupported report format %q", command.Format)
	}
	if command.Action == "check" && len(command.Profiles) == 0 && command.RequestFile == "" {
		return command, fmt.Errorf("check requires --profile or --request-file")
	}
	if command.RequestFile != "" && command.Action != "check" {
		return command, fmt.Errorf("--request-file requires check")
	}
	if command.Target != "" && command.Action != "capture" {
		return command, fmt.Errorf("--target is an explicit capture compatibility assertion")
	}
	if command.Target != "" && !nativeTargetMatchesHost(command.Target) {
		return command, fmt.Errorf("--target must identify the supported native host; cross-compilation is unsupported")
	}
	if command.Action == "capture" && (command.Output == "" || command.NativeEnvironment != "" || command.Trust != "" || len(command.Profiles) != 0) {
		return command, fmt.Errorf("capture requires --output and cannot consume environment evidence, approve trust or enforce profiles")
	}
	for _, profile := range command.Profiles {
		if profile == "PreservesOnFailure" {
			return command, fmt.Errorf("PreservesOnFailure requires --request-file with explicit footprint, observation and outcomes")
		}
	}
	if command.Action == "explain" && command.Binding == "" {
		return command, fmt.Errorf("explain requires --binding")
	}
	if command.Language == "" {
		return command, fmt.Errorf("--language requires a nonempty value")
	}
	if command.Assurance != "implementation" && command.Assurance != "symbolic" {
		return command, fmt.Errorf("unsupported assurance %q", command.Assurance)
	}
	if len(command.Adjectives) > 0 && command.Action != "verify-dictionary" {
		return command, fmt.Errorf("--adjective is supported only by verify-dictionary")
	}
	if command.Assurance == "symbolic" {
		if command.Action != "verify-dictionary" || len(command.Adjectives) == 0 {
			return command, fmt.Errorf("symbolic assurance requires verify-dictionary and an explicit --adjective selection")
		}
		if len(command.Profiles) > 0 {
			return command, fmt.Errorf("symbolic dictionary verification cannot enforce whole-entry --profile requirements; use contracts check")
		}
	}
	return command, nil
}

type contractMetadata struct {
	Native             []contracts.DeclaredContract
	Facets             []contracts.FacetRoleManifest
	SourceTransformers []contracts.SourceTransformerDeclaration
	Trust              contracts.TrustPolicyDocument
	Environment        *contracts.NativeEnvironment
	Requests           *contracts.CheckRequests
}

type contractMetadataError struct{ error }

func contractManifestIndex(data []byte) ([]string, error) {
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	start, err := decoder.Token()
	if err != nil || start != json.Delim('{') {
		return nil, fmt.Errorf("contract index must be an object")
	}
	seen := make(map[string]bool)
	var version string
	var manifests []string
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		name, ok := token.(string)
		if !ok || seen[name] {
			return nil, fmt.Errorf("duplicate or invalid contract index field %q", token)
		}
		seen[name] = true
		switch name {
		case "schema_version":
			err = decoder.Decode(&version)
		case "manifests":
			err = decoder.Decode(&manifests)
		default:
			return nil, fmt.Errorf("unknown contract index field %q", name)
		}
		if err != nil {
			return nil, err
		}
	}
	if _, err := decoder.Token(); err != nil {
		return nil, err
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("trailing data in contract index")
	}
	if version != contracts.SchemaVersion || !seen["manifests"] || manifests == nil {
		return nil, fmt.Errorf("contract index requires version %s and a manifests array", contracts.SchemaVersion)
	}
	return manifests, nil
}

func loadContractMetadata(resources fs.FS, command contractsCommand) (contractMetadata, error) {
	var result contractMetadata
	var blobs [][]byte
	if data, err := fs.ReadFile(resources, "contracts/index.json"); err == nil {
		manifests, err := contractManifestIndex(data)
		if err != nil {
			return result, err
		}
		for _, name := range manifests {
			if !fs.ValidPath(name) {
				return result, fmt.Errorf("invalid resource contract manifest %q", name)
			}
			body, err := fs.ReadFile(resources, name)
			if err != nil {
				return result, err
			}
			blobs = append(blobs, body)
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return result, err
	}
	for _, name := range command.Manifests {
		data, err := os.ReadFile(name)
		if err != nil {
			return result, err
		}
		blobs = append(blobs, data)
	}
	seen := make(map[contracts.MeaningKey]string)
	records := make(map[string]string)
	register := func(kind, identity string, value any) error {
		fingerprint, err := contracts.Fingerprint(value)
		if err != nil {
			return err
		}
		key := kind + ":" + identity
		if previous, exists := records[key]; exists && previous != fingerprint {
			return fmt.Errorf("conflicting %s declarations for %s", kind, identity)
		}
		records[key] = fingerprint
		return nil
	}
	for _, data := range blobs {
		document, err := contracts.DecodeDocument(data)
		if err != nil {
			return result, err
		}
		switch typed := document.(type) {
		case contracts.NativeContractManifest:
			for _, declaration := range typed.Declarations {
				if err := register("native contract", declaration.ID, declaration); err != nil {
					return result, err
				}
				fingerprint, err := contracts.Fingerprint(declaration)
				if err != nil {
					return result, err
				}
				if old, ok := seen[declaration.Meaning]; ok {
					if old != fingerprint {
						return result, fmt.Errorf("conflicting declarations for canonical meaning %s", declaration.Meaning)
					}
					continue
				}
				seen[declaration.Meaning] = fingerprint
				result.Native = append(result.Native, declaration)
			}
		case contracts.FacetRoleManifest:
			for _, facet := range typed.Facets {
				if err := register("facet", facet.ID, facet); err != nil {
					return result, err
				}
			}
			for _, role := range typed.Roles {
				if err := register("role", role.ID, role); err != nil {
					return result, err
				}
			}
			for _, transformer := range typed.Transformers {
				if err := register("transformer", string(transformer.Meaning), transformer); err != nil {
					return result, err
				}
			}
			for _, transformer := range typed.SourceTransformers {
				if err := register("source transformer", string(transformer.Meaning), transformer); err != nil {
					return result, err
				}
				duplicate := false
				for _, existing := range result.SourceTransformers {
					duplicate = duplicate || existing.Meaning == transformer.Meaning
				}
				if !duplicate {
					result.SourceTransformers = append(result.SourceTransformers, transformer)
				}
			}
			result.Facets = append(result.Facets, typed)
		default:
			return result, fmt.Errorf("--manifest accepts declarations, never reports, certificates or trust policies")
		}
	}
	if command.Trust != "" {
		data, err := os.ReadFile(command.Trust)
		if err != nil {
			return result, err
		}
		document, err := contracts.DecodeDocument(data)
		if err != nil {
			return result, err
		}
		policy, ok := document.(contracts.TrustPolicyDocument)
		if !ok {
			return result, fmt.Errorf("--trust-policy requires a trust_policy document")
		}
		result.Trust = policy
	}
	for kind, path := range map[string]string{"native_environment": command.NativeEnvironment, "contract_check_requests": command.RequestFile} {
		if path == "" {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return result, err
		}
		document, err := contracts.DecodeDocument(data)
		if err != nil {
			return result, err
		}
		switch value := document.(type) {
		case contracts.NativeEnvironment:
			if kind != "native_environment" {
				return result, fmt.Errorf("request file requires contract_check_requests")
			}
			result.Environment = &value
		case contracts.CheckRequests:
			if kind != "contract_check_requests" {
				return result, fmt.Errorf("native environment requires native_environment")
			}
			result.Requests = &value
		default:
			return result, fmt.Errorf("expected %s document", kind)
		}
	}
	return result, nil
}

func runContracts(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	command, err := parseContractsCommand(args)
	if errors.Is(err, flag.ErrHelp) {
		if _, err := io.WriteString(stdout, contractsUsage); err != nil {
			return 1
		}
		return 0
	}
	if err != nil {
		return contractsError(stderr, 2, err)
	}
	if err := ctx.Err(); err != nil {
		return contractsError(stderr, 1, err)
	}
	resources := distribution.Files()
	if command.Assets != "" {
		resources = os.DirFS(command.Assets)
	}
	metadata, err := loadContractMetadata(resources, command)
	if err != nil {
		return contractsError(stderr, 2, err)
	}
	service := Lyger.NewLygerService(Lyger.Options{Language: command.Language, Resources: resources, AssetRoot: command.Assets})
	inputPath, err := filepath.Abs(command.Input)
	if err != nil {
		return contractsError(stderr, 1, err)
	}
	var analysis *Lyger.ContractAnalysis
	var dictionary *Lyger.DictionaryContractAnalysis
	analysisOptions := Lyger.ContractAnalysisOptions{SourceTransformers: metadata.SourceTransformers}
	if command.Action == "verify-dictionary" {
		dictionary, err = service.VerifyDictionaryContracts(command.Input, analysisOptions)
		if err == nil {
			analysis = dictionary.Analysis
		}
	} else {
		var source []byte
		source, err = os.ReadFile(inputPath)
		if err != nil {
			return contractsError(stderr, 1, err)
		}
		analysis, err = service.AnalyzeContracts(source, inputPath, analysisOptions)
	}
	if err != nil {
		if errors.Is(err, Lyger.ErrSourceContractMetadata) {
			return contractsError(stderr, 2, err)
		}
		return contractsError(stderr, 1, err)
	}
	var selectedAdjectives []contracts.MeaningKey
	if dictionary != nil && len(command.Adjectives) > 0 {
		selectedAdjectives, err = selectContractAdjectives(dictionary.Definitions, command.Adjectives)
		if err != nil {
			return contractsError(stderr, 2, err)
		}
	}
	report, err := prepareContractReport(command, analysis, resources, metadata)
	if err != nil {
		var metadataError contractMetadataError
		if errors.As(err, &metadataError) {
			return contractsError(stderr, 2, err)
		}
		return contractsError(stderr, 1, err)
	}
	if command.Action == "capture" {
		if err := validateContractReportDestinations(command, report); err != nil {
			return contractsError(stderr, 2, err)
		}
		environment, err := captureNativeEnvironment(ctx, report, command.Target)
		if err != nil {
			return contractsError(stderr, 1, err)
		}
		data, err := contracts.EncodeDocument(environment)
		if err != nil {
			return contractsError(stderr, 1, err)
		}
		if err := writeContractsReport(command.Output, inputPath, data); err != nil {
			return contractsError(stderr, 1, err)
		}
		return 0
	}
	if dictionary != nil {
		report.Definitions = dictionary.Definitions
		report.Obligations = append(report.Obligations, dictionary.Obligations...)
		report.Certificates = append(report.Certificates, dictionary.Certificates...)
		report.SourceCertificates = dictionary.SourceCertificates
		report.SourceInstantiations = dictionary.SourceInstantiations
		if len(command.Adjectives) > 0 {
			report.SourceVerification, err = selectedSourceVerification(report, command.Assurance, selectedAdjectives)
			if err != nil {
				return contractsError(stderr, 2, err)
			}
		}
	}
	if command.Action == "explain" {
		var selected []contracts.ResolvedBinding
		for _, binding := range report.Bindings {
			selector := binding.Origin.Source + ":" + strconv.Itoa(binding.Origin.Start)
			if string(binding.Occurrence) == command.Binding || selector == command.Binding {
				selected = append(selected, binding)
			}
		}
		if len(selected) == 0 {
			return contractsError(stderr, 2, fmt.Errorf("no binding matches %q; use an occurrence ID or source:byte-offset from inspect", command.Binding))
		}
		report.Bindings = selected
		report.SourceCertificates, report.SourceInstantiations = relevantSourceExplanations(report, selected)
	}
	data, err := contracts.EncodeDocument(report)
	if err != nil {
		return contractsError(stderr, 1, err)
	}
	if err := validateContractReportDestinations(command, report); err != nil {
		return contractsError(stderr, 2, err)
	}
	if command.Report != "" {
		if err := writeContractsReport(command.Report, inputPath, data); err != nil {
			return contractsError(stderr, 1, err)
		}
	}
	output := stdout
	var file *os.File
	if command.Output != "" {
		if sameContractPath(command.Output, inputPath) {
			return contractsError(stderr, 2, fmt.Errorf("report output must not replace source input"))
		}
		file, err = os.OpenFile(command.Output, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return contractsError(stderr, 1, err)
		}
		output = file
	}
	if command.Format == "json" {
		_, err = output.Write(append(data, '\n'))
	} else {
		err = writeContractText(output, report)
	}
	if file != nil {
		err = errors.Join(err, file.Close())
	}
	if err != nil {
		return contractsError(stderr, 1, err)
	}
	if command.Action == "check" {
		if len(command.Profiles) > 0 && !contracts.AllEstablished(report.Obligations) {
			return 1
		}
		for _, request := range report.Requests {
			if request.Status != contracts.Established {
				return 1
			}
		}
	}
	if command.Action == "verify-dictionary" {
		if command.Assurance == "symbolic" {
			if report.SourceVerification == nil || report.SourceVerification.Status != contracts.Established {
				return 1
			}
			return 0
		}
		if !contracts.AllEstablished(report.Obligations) {
			return 1
		}
		if report.SourceVerification != nil && report.SourceVerification.Status != contracts.Established {
			return 1
		}
		for _, certificate := range report.Certificates {
			if certificate.Status != contracts.Established {
				return 1
			}
		}
		for _, certificate := range report.SourceCertificates {
			if certificate.LogicalStatus != contracts.Established || !contracts.AllEstablished(certificate.LoweringObligations) {
				return 1
			}
		}
	}
	return 0
}

func contractsError(output io.Writer, status int, err error) int {
	if _, writeErr := fmt.Fprintf(output, "error: %v\n", err); writeErr != nil {
		// The original failure status remains meaningful when stderr is closed.
		return status
	}
	return status
}

func selectContractAdjectives(definitions []contracts.DefinitionRecord, selectors []string) ([]contracts.MeaningKey, error) {
	var selected []contracts.MeaningKey
	seen := make(map[contracts.MeaningKey]bool)
	for _, selector := range selectors {
		matches := make(map[contracts.MeaningKey]bool)
		for _, definition := range definitions {
			if !definition.Descriptor {
				continue
			}
			origin := definition.Origin.Source + ":" + strconv.Itoa(definition.Origin.Start)
			if string(definition.Meaning) == selector || origin == selector {
				matches[definition.Meaning] = true
			}
		}
		if len(matches) != 1 {
			return nil, fmt.Errorf("adjective selector %q matches %d canonical descriptor rules; use a unique meaning ID or declaration source:byte-offset from inspect", selector, len(matches))
		}
		for meaning := range matches {
			if !seen[meaning] {
				selected = append(selected, meaning)
				seen[meaning] = true
			}
		}
	}
	return selected, nil
}

func selectedSourceVerification(report contracts.Report, assurance string, selected []contracts.MeaningKey) (*contracts.SourceVerificationResult, error) {
	result := &contracts.SourceVerificationResult{Assurance: assurance, Scope: "source_elaboration", Selected: append([]contracts.MeaningKey(nil), selected...), Status: contracts.Established}
	if len(selected) == 0 {
		return nil, fmt.Errorf("source verification requires an explicit nonempty adjective selection")
	}
	setStatus := func(status contracts.Status) {
		if status == contracts.Contradicted || result.Status == contracts.Established && status != contracts.Established {
			result.Status = status
		}
	}
	for _, meaning := range selected {
		found := false
		for _, certificate := range report.SourceCertificates {
			if certificate.Meaning != meaning {
				continue
			}
			if found {
				return nil, fmt.Errorf("ambiguous compiler-derived source certificates for %s", meaning)
			}
			found = true
			result.Obligations = append(result.Obligations, certificate.LogicalObligations...)
			setStatus(certificate.LogicalStatus)
		}
		if !found {
			setStatus(contracts.Unresolved)
			id := contracts.StableID("missing-source-certificate", meaning)
			result.Obligations = append(result.Obligations, contracts.Obligation{ID: id, Claim: contracts.Claim{ID: id, Predicate: contracts.Predicate{Kind: "source_projection", Subject: string(meaning)}, Scope: "source_elaboration", Phase: contracts.Elaboration}, Status: contracts.Unresolved, Reason: "selected declaration has no compiler-derived source certificate"})
		}
	}
	if assurance != "symbolic" {
		result.Scope = "whole_dictionary"
		result.Obligations = append(result.Obligations, report.Obligations...)
		for _, certificate := range report.Certificates {
			setStatus(certificate.Status)
			result.Obligations = append(result.Obligations, certificate.Obligations...)
		}
		for _, certificate := range report.SourceCertificates {
			setStatus(certificate.LogicalStatus)
			result.Obligations = append(result.Obligations, certificate.LogicalObligations...)
			result.Obligations = append(result.Obligations, certificate.LoweringObligations...)
		}
	}
	var obligations []contracts.Obligation
	seen := make(map[string]bool)
	for _, obligation := range result.Obligations {
		setStatus(obligation.Status)
		fingerprint := contracts.StableID("verification-obligation", obligation)
		if !seen[fingerprint] {
			obligations = append(obligations, obligation)
			seen[fingerprint] = true
		}
	}
	result.Obligations = obligations
	return result, nil
}

func relevantSourceExplanations(report contracts.Report, bindings []contracts.ResolvedBinding) ([]contracts.SourceAdjectiveCertificate, []contracts.SourceAdjectiveInstantiation) {
	meanings := make(map[contracts.MeaningKey]bool)
	for _, binding := range bindings {
		meanings[binding.Meaning] = true
		exactRule := false
		for _, certificate := range report.SourceCertificates {
			exactRule = exactRule || certificate.Meaning == binding.Meaning
		}
		if exactRule {
			continue
		}
		for _, definition := range report.Definitions {
			if definition.Descriptor && definition.Definition == binding.Definition {
				meanings[definition.Meaning] = true
			}
		}
	}
	var certificates []contracts.SourceAdjectiveCertificate
	for _, certificate := range report.SourceCertificates {
		if meanings[certificate.Meaning] {
			certificates = append(certificates, certificate)
		}
	}
	var instances []contracts.SourceAdjectiveInstantiation
	for _, instance := range report.SourceInstantiations {
		if !meanings[instance.Meaning] {
			continue
		}
		for _, binding := range bindings {
			actual := instance.Target
			actual.Phase = binding.Target.Phase
			if binding.Target.Unknown || binding.Target.Root == "" || contracts.TargetRelation(actual, binding.Target) == contracts.Established {
				instances = append(instances, instance)
				break
			}
		}
	}
	return certificates, instances
}

func sameContractPath(left, right string) bool {
	leftInfo, leftErr := os.Stat(left)
	rightInfo, rightErr := os.Stat(right)
	if leftErr == nil && rightErr == nil && os.SameFile(leftInfo, rightInfo) {
		return true
	}
	a, err := filepath.Abs(left)
	if err != nil {
		return false
	}
	b, err := filepath.Abs(right)
	return err == nil && a == b
}
func writeContractsReport(path, input string, data []byte) error {
	if sameContractPath(path, input) {
		return fmt.Errorf("report must not replace source input")
	}
	return os.WriteFile(path, append(data, '\n'), 0600)
}

func validateContractReportDestinations(command contractsCommand, report contracts.Report) error {
	inputs := append([]string{command.Input, command.Trust, command.NativeEnvironment, command.RequestFile}, command.Manifests...)
	for path := range report.Inputs {
		if filepath.IsAbs(path) {
			inputs = append(inputs, path)
		}
	}
	for _, output := range []string{command.Output, command.Report} {
		if output == "" {
			continue
		}
		for _, input := range inputs {
			if input != "" && sameContractPath(input, output) {
				return fmt.Errorf("contract output must not replace selected source or metadata input %s", input)
			}
		}
		if command.Assets != "" {
			if err := validateAssetDestinations(command.Assets, output); err != nil {
				return err
			}
		}
	}
	return nil
}

func prepareContractReport(command contractsCommand, analysis *Lyger.ContractAnalysis, resources fs.FS, metadata contractMetadata) (contracts.Report, error) {
	report := contracts.NewReport(command.Input)
	report.Bindings, report.Targets, report.Inputs = analysis.Bindings.Bindings, analysis.Bindings.Targets, analysis.Bindings.Inputs
	report.Definitions = analysis.Bindings.Definitions
	report.Lifecycle = analysis.Lifecycle
	report.Declarations = metadata.Native
	report.SourceDeclarations = metadata.SourceTransformers
	// Collect backend inventories separately until all frontend composition is
	// complete; OrchestratorUnit and RuntimeLibrary can select late helpers.
	report.Lowering = contracts.LoweringInventory{Fingerprints: make(map[string]string)}
	report.Scope = contracts.ReportScope{Phase: contracts.Invocation, Boundary: "whole_region", Target: "native-workspace", Platform: runtime.GOOS + "/" + runtime.GOARCH, Compiler: distribution.Version(), PanicPolicy: "selected-runtime", Exclusions: []string{"external forced termination and host loss", "scheduler fairness", "no build or runtime tests executed by inspection", "opaque native dependency closure and implementation conformance remain unresolved"}}
	backend := Logiceme.NewLogicemeService(resources)
	workspace, err := analysis.Compilation.WorkspacePlan()
	if err != nil {
		return report, err
	}
	locations := make(map[string]string)
	apps, libraries := 0, 0
	for _, item := range workspace.Packages() {
		if item.Library {
			locations[item.ID] = fmt.Sprintf("services/service-%03d", libraries)
			libraries++
		} else {
			locations[item.ID] = fmt.Sprintf("apps/app-%03d", apps)
			apps++
		}
	}
	units, err := workspace.Units(locations)
	if err != nil {
		return report, err
	}
	var bundles []*sourceBundle
	merge := func(namespace string, inventory contracts.LoweringInventory) {
		for _, step := range inventory.Steps {
			step.ID = contracts.StableID("inventory-use", namespace+":"+step.ID)
			report.Lowering.Steps = append(report.Lowering.Steps, step)
		}
		report.Lowering.Native = append(report.Lowering.Native, inventory.Native...)
		report.Lowering.Unknowns = append(report.Lowering.Unknowns, inventory.Unknowns...)
		for key, value := range inventory.Fingerprints {
			report.Lowering.Fingerprints[namespace+":"+key] = value
		}
	}
	for _, unit := range units {
		prepared, err := backend.PrepareContractGeneration(unit.BackendAML, Logiceme.GenerationOptions{})
		if err != nil {
			return report, err
		}
		inventory := prepared.Inventory()
		bundle, err := prepared.Render()
		if err != nil {
			return report, err
		}
		if err := contracts.VerifyInventory(inventory, prepared.Inventory()); err != nil {
			return report, err
		}
		merge(unit.ID, inventory)
		for i := range bundle.ByteArrays {
			planned, err := planSourceBundle(bundle.ByteArrays[i], bundle.MapArrays[i])
			if err != nil {
				return report, err
			}
			bundles = append(bundles, planned)
		}
	}
	stem := strings.TrimSuffix(filepath.Base(command.Input), filepath.Ext(command.Input))
	entryName := orchestratorPackageName(stem, bundles)
	var companions []Lyger.AppArtifact
	for _, item := range workspace.Packages() {
		if !item.Library {
			companions = append(companions, Lyger.AppArtifact{ID: item.ID, RelativeExecutable: fmt.Sprintf("%s.apps/generations/contract-inspection/app-%03d", stem, len(companions))})
		}
	}
	entry, err := analysis.Compilation.OrchestratorUnit(entryName, companions)
	if err != nil {
		return report, err
	}
	runtimeName := runtimePackageName(entryName, bundles)
	imports, err := backend.ReadImports(entry.BackendAML)
	if err != nil {
		return report, err
	}
	library, err := analysis.Compilation.RuntimeLibrary(runtimeName, imports)
	if err != nil {
		return report, err
	}
	var selected [][]byte
	if library != nil {
		selected = append(selected, library)
	}
	prepared, err := backend.PrepareWorkspaceContractGeneration(entry.BackendAML, Logiceme.WorkspaceRuntime{Package: runtimeName, RelativeDirectory: "../support/runtime"}, selected...)
	if err != nil {
		return report, err
	}
	inventory := prepared.Inventory()
	if _, err := prepared.Render(); err != nil {
		return report, err
	}
	if err := contracts.VerifyInventory(inventory, prepared.Inventory()); err != nil {
		return report, err
	}
	merge("orchestrator", inventory)
	if err := analysis.RefreshLoweringInventory(); err != nil {
		return report, err
	}
	backendInventory := report.Lowering
	report.Lowering = contracts.CloneInventory(analysis.Lowering)
	merge("backend", backendInventory)
	report.Bindings, report.Targets, report.Inputs = analysis.Bindings.Bindings, analysis.Bindings.Targets, analysis.Bindings.Inputs
	report.Definitions = analysis.Bindings.Definitions
	// Backend obligations are outside source control flow: even a source Exit
	// cannot hide runtime startup, terminal cleanup or opaque native dependencies.
	plan := analysis.Plan
	plan.NativeClosure = nil
	for _, boundary := range analysis.Plan.NativeClosure {
		if boundary.ID != "backend-support-pending" {
			plan.NativeClosure = append(plan.NativeClosure, boundary)
		}
	}
	plan.NativeClosure = append(plan.NativeClosure, backendInventory.Native...)
	plan.Residuals = append(append([]contracts.UnknownReason(nil), plan.Residuals...), backendInventory.Unknowns...)
	analysis.Plan, analysis.Lowering = plan, report.Lowering
	report.NativeUses, report.Scopes = analysis.NativeUses, analysis.Scopes
	inferenceOptions, err := prepareNativeApplications(&report, metadata)
	if err != nil {
		return report, contractMetadataError{err}
	}
	if err := analysis.ReinferContracts(Lyger.ContractAnalysisOptions{Inference: inferenceOptions}); err != nil {
		return report, err
	}
	report.Summary = analysis.Summary
	report.SourceCertificates = analysis.SourceCertificates
	report.SourceInstantiations = analysis.SourceInstantiations
	for _, instance := range report.SourceInstantiations {
		report.Obligations = append(report.Obligations, instance.Obligations...)
		report.Obligations = append(report.Obligations, instance.LoweringObligations...)
	}
	for _, certificate := range report.SourceCertificates {
		for _, declaration := range metadata.SourceTransformers {
			if declaration.Meaning == certificate.Meaning {
				report.Obligations = append(report.Obligations, certificate.LogicalObligations...)
				break
			}
		}
	}
	known := make(map[contracts.MeaningKey]bool)
	for _, binding := range report.Bindings {
		known[binding.Meaning] = true
	}
	for _, definition := range report.Definitions {
		known[definition.Meaning] = true
	}
	for _, step := range report.Lowering.Steps {
		known[step.Meaning] = true
	}
	for _, use := range report.NativeUses {
		known[use.Meaning] = true
	}
	for _, declaration := range metadata.Native {
		if !known[declaration.Meaning] {
			return report, contractMetadataError{fmt.Errorf("native manifest references absent or ambiguous meaning %s (%s)", declaration.Meaning, declaration.Name)}
		}
	}
	for _, manifest := range metadata.Facets {
		for _, transformer := range manifest.Transformers {
			if !known[transformer.Meaning] {
				return report, contractMetadataError{fmt.Errorf("transformer manifest references absent canonical meaning %s", transformer.Meaning)}
			}
			certificate, err := contracts.VerifyAdjective(transformer, contracts.AnalysisOptions{})
			if err != nil {
				return report, contractMetadataError{err}
			}
			// This proves a theorem about the authored typed transformer, not
			// conformance of a source body to that model. That dependency remains
			// explicit until a source-body adapter or approved witness discharges it.
			assumption := "source-transformer-conformance:" + string(transformer.Meaning)
			certificate.Assumptions = append(certificate.Assumptions, assumption)
			report.Obligations = append(report.Obligations, contracts.Obligation{ID: assumption, Claim: contracts.Claim{ID: assumption, Predicate: contracts.Predicate{Kind: "transformer_conformance", Subject: string(transformer.Meaning)}, Target: transformer.Target, Scope: "definition", Phase: contracts.Elaboration}, Status: contracts.Unresolved, Reason: "symbolic theorem checks the authored typed model; source-body conformance to that model is not established", Origin: transformer.Origin})
			report.Certificates = append(report.Certificates, certificate)
		}
	}
	report.Obligations = append(report.Obligations, report.Summary.Obligations...)
	for _, name := range command.Profiles {
		profile := contracts.Profile{Name: name, Version: contracts.SchemaVersion, Scope: "whole-region"}
		var obligations []contracts.Obligation
		switch name {
		case "OwnedLifecycle":
			obligations = analysis.CheckLifecycleProfile(profile)
		case "PreservesOnFailure":
			// No source syntax or CLI convention supplies an observable footprint.
			// Use the same checker as typed clients so that the missing premise is
			// diagnosed precisely rather than treating all errors as transactional.
			obligations = contracts.CheckFailureSummaryProfile(report.Summary, profile, contracts.FailureFootprint{}, nil)
		default:
			obligations = contracts.CheckProfile(report.Summary, profile)
		}
		status := contracts.Established
		for _, obligation := range obligations {
			if obligation.Status == contracts.Contradicted {
				status = contracts.Contradicted
				break
			}
			if obligation.Status != contracts.Established {
				status = contracts.Unresolved
			}
		}
		report.Profiles = append(report.Profiles, contracts.ProfileResult{Profile: profile, Status: status, Obligations: obligations})
		report.Obligations = append(report.Obligations, obligations...)
	}
	if err := applyContractTrust(&report, metadata); err != nil {
		return report, err
	}
	if err := applyContractRequests(&report, analysis, metadata, inferenceOptions); err != nil {
		return report, contractMetadataError{err}
	}
	return report, nil
}

func writeContractText(output io.Writer, report contracts.Report) error {
	if _, err := fmt.Fprintf(output, "L Contracts %s — %s\nResolution: %s\nScope: %s (%s), %s\nCoverage: %s; native transitive closure: %s\n", report.SchemaVersion, report.Entry, report.ResolutionPolicy, report.Scope.Boundary, report.Scope.Platform, report.Scope.Target, report.Summary.Coverage, report.Lowering.TransitiveCoverage); err != nil {
		return err
	}
	for _, scope := range report.Scopes {
		if _, err := fmt.Fprintf(output, "Scope %s: %s at %s:%d meaning=%s target=%s\n", scope.ID, scope.Boundary, scope.Origin.Source, scope.Origin.Start, scope.Meaning, scope.Target.String()); err != nil {
			return err
		}
		for _, location := range scope.Resources {
			if _, err := fmt.Fprintf(output, "  resource: %s projection=%v\n", location.Resource.ID, location.Path); err != nil {
				return err
			}
		}
	}
	for _, request := range report.Requests {
		if _, err := fmt.Fprintf(output, "Request %s: %s (%s, %s)\n", request.Request.ID, request.Status, request.Request.Profile.Name, request.Request.Scope); err != nil {
			return err
		}
	}
	if verification := report.SourceVerification; verification != nil {
		if _, err := fmt.Fprintf(output, "Source verification: %s assurance, %s scope, %s\n", verification.Assurance, verification.Scope, verification.Status); err != nil {
			return err
		}
		for _, meaning := range verification.Selected {
			if _, err := fmt.Fprintf(output, "  selected adjective: %s\n", meaning); err != nil {
				return err
			}
		}
		if verification.Assurance == "symbolic" {
			if _, err := fmt.Fprintln(output, "Only selected source elaboration obligations determine this result. Native and runtime implementation assurance is reported separately below."); err != nil {
				return err
			}
		}
	}
	for _, certificate := range report.SourceCertificates {
		if _, err := fmt.Fprintf(output, "Source adjective %s: %s (%s, %s), target=%s nominal=%s\n", certificate.Meaning, certificate.LogicalStatus, certificate.Scope, certificate.Mode, certificate.Target.String(), certificate.TargetNominal); err != nil {
			return err
		}
	}
	for _, instance := range report.SourceInstantiations {
		if _, err := fmt.Fprintf(output, "Source adjective instance %s: %s target=%s\n", instance.Meaning, instance.Status, instance.Target.String()); err != nil {
			return err
		}
	}
	for _, profile := range report.Profiles {
		if _, err := fmt.Fprintf(output, "%s: %s\n", profile.Profile.Name, profile.Status); err != nil {
			return err
		}
	}
	for _, assumption := range report.Assumptions {
		if _, err := fmt.Fprintf(output, "Assumption: %s\n", assumption); err != nil {
			return err
		}
	}
	for _, certificate := range report.Certificates {
		if _, err := fmt.Fprintf(output, "Certificate: %s (%s): %s\n", certificate.Meaning, certificate.Mode, certificate.Status); err != nil {
			return err
		}
		for _, assumption := range certificate.Assumptions {
			if _, err := fmt.Fprintf(output, "  Assumption: %s\n", assumption); err != nil {
				return err
			}
		}
	}
	for _, binding := range report.Bindings {
		if _, err := fmt.Fprintf(output, "\n%s %q -> %s\n  meaning: %s\n  target: %s\n  origin: %s:%d\n  selection: %s\n", binding.Occurrence, binding.Spelling, binding.CanonicalName, binding.Meaning, binding.Target.String(), binding.Origin.Source, binding.Origin.Start, binding.SelectionReason); err != nil {
			return err
		}
	}
	var explanations []contracts.Explanation
	for _, explanation := range contracts.ExplainReport(report) {
		if explanation.Kind != "binding" {
			explanations = append(explanations, explanation)
		}
	}
	if _, err := io.WriteString(output, contracts.FormatExplanations(explanations)); err != nil {
		return err
	}
	var boundaries []string
	for _, boundary := range report.Lowering.Native {
		boundaries = append(boundaries, boundary.ID+": "+boundary.Reason)
	}
	sort.Strings(boundaries)
	for _, boundary := range boundaries {
		if _, err := fmt.Fprintln(output, "Native: "+boundary); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(output, "Executed checks: none (source inspection only).")
	return err
}
