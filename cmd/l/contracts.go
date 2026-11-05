package main

import (
	"context"
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
		if _, supported := contracts.LookupProfile(value); !supported {
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
		descriptor, _ := contracts.LookupProfile(profile)
		if descriptor.RequiresFootprint {
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
	inputPaths         []string // Physical inputs reserved against report writes; never serialized.
	Native             []contracts.DeclaredContract
	Facets             []contracts.FacetRoleManifest
	SourceTransformers []contracts.SourceTransformerDeclaration
	Trust              contracts.TrustPolicyDocument
	Environment        *contracts.NativeEnvironment
	Requests           *contracts.CheckRequests
}

type contractMetadataError struct{ error }

func loadContractMetadata(resources fs.FS, command contractsCommand) (contractMetadata, error) {
	var result contractMetadata
	const index = "contracts/index.json"
	// Indexed paths belong to the selected provider. Embedded resources have
	// no filesystem input path; explicit command-line paths always do.
	type input struct {
		path, option string
		resource     bool
	}
	var inputs []input
	if command.Assets != "" {
		result.inputPaths = append(result.inputPaths, filepath.Join(command.Assets, index))
	}
	if data, err := fs.ReadFile(resources, index); err == nil {
		names, err := contracts.ReadManifestIndex(data)
		if err != nil {
			return result, err
		}
		for _, name := range names {
			if !fs.ValidPath(name) {
				return result, fmt.Errorf("invalid resource contract manifest %q", name)
			}
			inputs = append(inputs, input{name, "--manifest", true})
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return result, err
	}
	for _, name := range command.Manifests {
		inputs = append(inputs, input{name, "--manifest", false})
	}
	inputs = append(inputs, input{command.Trust, "--trust-policy", false}, input{command.NativeEnvironment, "--native-environment", false}, input{command.RequestFile, "--request-file", false})
	records := make(map[string]string)
	register := func(value any, keys ...string) error {
		fingerprint, err := contracts.Fingerprint(value)
		if err != nil {
			return err
		}
		for _, key := range keys {
			if previous, exists := records[key]; exists && previous != fingerprint {
				return fmt.Errorf("conflicting declarations for %s", key)
			}
			records[key] = fingerprint
		}
		return nil
	}
	for _, selected := range inputs {
		if selected.path == "" {
			continue
		}
		var data []byte
		var err error
		if selected.resource {
			data, err = fs.ReadFile(resources, selected.path)
			if command.Assets != "" {
				result.inputPaths = append(result.inputPaths, filepath.Join(command.Assets, filepath.FromSlash(selected.path)))
			}
		} else {
			data, err = os.ReadFile(selected.path)
			result.inputPaths = append(result.inputPaths, selected.path)
		}
		if err != nil {
			return result, err
		}
		document, err := contracts.DecodeDocument(data)
		if err != nil {
			return result, err
		}
		switch selected.option {
		case "--trust-policy":
			policy, ok := document.(contracts.TrustPolicyDocument)
			if !ok {
				return result, fmt.Errorf("--trust-policy requires a trust_policy document")
			}
			result.Trust = policy
			continue
		case "--native-environment":
			environment, ok := document.(contracts.NativeEnvironment)
			if !ok {
				return result, fmt.Errorf("--native-environment requires a native_environment document")
			}
			result.Environment = &environment
			continue
		case "--request-file":
			requests, ok := document.(contracts.CheckRequests)
			if !ok {
				return result, fmt.Errorf("--request-file requires a contract_check_requests document")
			}
			result.Requests = &requests
			continue
		}
		switch typed := document.(type) {
		case contracts.NativeContractManifest:
			for _, declaration := range typed.Declarations {
				key := "canonical meaning:" + string(declaration.Meaning)
				_, duplicate := records[key]
				if err := register(declaration, key, "native contract:"+declaration.ID); err != nil {
					return result, err
				}
				if !duplicate {
					result.Native = append(result.Native, declaration)
				}
			}
		case contracts.FacetRoleManifest:
			for _, facet := range typed.Facets {
				if err := register(facet, "facet:"+facet.ID); err != nil {
					return result, err
				}
			}
			for _, role := range typed.Roles {
				if err := register(role, "role:"+role.ID); err != nil {
					return result, err
				}
			}
			for _, transformer := range typed.Transformers {
				if err := register(transformer, "transformer:"+string(transformer.Meaning)); err != nil {
					return result, err
				}
			}
			for _, transformer := range typed.SourceTransformers {
				key := "source transformer:" + string(transformer.Meaning)
				_, duplicate := records[key]
				if err := register(transformer, key); err != nil {
					return result, err
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
	if err := command.validateDestinations(analysis.Compilation.Compilation, metadata.inputPaths); err != nil {
		return contractsError(stderr, 2, err)
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
		environment, err := captureNativeEnvironment(ctx, report, command.Target)
		if err != nil {
			return contractsError(stderr, 1, err)
		}
		data, err := contracts.EncodeDocument(environment)
		if err != nil {
			return contractsError(stderr, 1, err)
		}
		if err := replaceFile(command.Output, append(data, '\n')); err != nil {
			return contractsError(stderr, 1, err)
		}
		return 0
	}
	var verification *contracts.SourceVerificationResult
	if dictionary != nil {
		report.Definitions = dictionary.Definitions
		report.Obligations = append(report.Obligations, dictionary.Obligations...)
		report.Certificates = append(report.Certificates, dictionary.Certificates...)
		report.SourceCertificates = dictionary.SourceCertificates
		report.SourceInstantiations = dictionary.SourceInstantiations
		var missing []contracts.MeaningKey
		verification, missing, err = report.DictionaryVerification(command.Assurance, selectedAdjectives)
		if err != nil {
			return contractsError(stderr, 2, err)
		}
		for _, meaning := range missing {
			id := contracts.StableID("missing-source-certificate", meaning)
			verification.Obligations = append(verification.Obligations, contracts.Obligation{ID: id, Claim: contracts.Claim{ID: id, Predicate: contracts.Predicate{Kind: "source_projection", Subject: string(meaning)}, Scope: "source_elaboration", Phase: contracts.Elaboration}, Status: contracts.Unresolved, Reason: "selected declaration has no compiler-derived source certificate"})
		}
		if len(selectedAdjectives) > 0 {
			report.SourceVerification = verification
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
	data = append(data, '\n')
	if command.Report != "" {
		if err := replaceFile(command.Report, data); err != nil {
			return contractsError(stderr, 1, err)
		}
	}
	if command.Format == "text" {
		data = contractText(report)
	}
	if command.Output == "" {
		_, err = stdout.Write(data)
	} else {
		err = replaceFile(command.Output, data)
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
	if verification != nil && verification.Status != contracts.Established {
		return 1
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

// Both paths have already been canonicalized. SameFile also catches hard links.
func sameContractPath(left, right string) bool {
	if left == right {
		return true
	}
	a, errA := os.Stat(left)
	b, errB := os.Stat(right)
	return errA == nil && errB == nil && os.SameFile(a, b)
}

func (command *contractsCommand) validateDestinations(compilation *Lyger.Compilation, metadata []string) error {
	if command.Output == "" && command.Report == "" {
		return nil
	}
	inputs := append([]string(nil), metadata...)
	for _, path := range compilation.SourcePaths {
		if command.Action == "verify-dictionary" && !strings.HasSuffix(command.Input, ".l") && path == compilation.MainPath {
			continue // The named-dictionary entry is a synthetic import.
		}
		inputs = append(inputs, path)
	}
	var destinations []string
	for _, destination := range []*string{&command.Output, &command.Report} {
		if *destination == "" {
			continue
		}
		path, err := canonicalBuildPath(*destination)
		if err != nil {
			return err
		}
		*destination = path
		destinations = append(destinations, path)
	}
	if len(destinations) == 2 && sameContractPath(command.Output, command.Report) {
		return fmt.Errorf("--output and --report must not replace the same file")
	}
	for _, input := range inputs {
		physical, err := canonicalBuildPath(input)
		if err != nil {
			return err
		}
		for _, destination := range destinations {
			if sameContractPath(physical, destination) {
				return fmt.Errorf("contract output must not replace selected source or metadata input %s", input)
			}
		}
	}
	for _, destination := range destinations {
		info, err := os.Stat(destination)
		switch {
		case err == nil && !info.Mode().IsRegular():
			return fmt.Errorf("contract output requires a regular file: %s", destination)
		case errors.Is(err, os.ErrNotExist):
			if parent, err := os.Stat(filepath.Dir(destination)); err != nil || !parent.IsDir() {
				return fmt.Errorf("contract output requires an existing parent directory: %s", destination)
			}
		case err != nil:
			return err
		}
	}
	if command.Assets != "" {
		return validateAssetDestinations(command.Assets, destinations...)
	}
	return nil
}

func prepareContractReport(command contractsCommand, analysis *Lyger.ContractAnalysis, resources fs.FS, metadata contractMetadata) (contracts.Report, error) {
	report := contracts.NewReport(command.Input)
	report.Lifecycle = analysis.Lifecycle
	report.Declarations = metadata.Native
	report.SourceDeclarations = metadata.SourceTransformers
	report.Scope = contracts.ReportScope{Phase: contracts.Invocation, Boundary: "whole_region", Target: "native-workspace", Platform: runtime.GOOS + "/" + runtime.GOARCH, Compiler: distribution.Version(), PanicPolicy: "selected-runtime", Exclusions: []string{"external forced termination and host loss", "scheduler fairness", "no build or runtime tests executed by inspection", "opaque native dependency closure and implementation conformance remain unresolved"}}
	backend := Logiceme.NewLogicemeService(resources)
	var backendInventory contracts.LoweringInventory
	generated, err := generatePackages(backend, analysis.Compilation, &backendInventory)
	if err != nil {
		return report, err
	}
	generation := "contract-inspection"
	if env := metadata.Environment; env != nil && env.Method == "build" {
		// DecodeDocument validates the generation before it is used in typed
		// AML placement. This path is never opened or executed by inspection.
		generation = env.Configuration["companion-generation"]
	}
	stem := strings.TrimSuffix(filepath.Base(command.Input), filepath.Ext(command.Input))
	if _, err := generated.workspace(backend, stem, generation, &backendInventory); err != nil {
		return report, err
	}
	if err := analysis.RefreshLoweringInventory(); err != nil {
		return report, err
	}
	report.Lowering = contracts.CloneInventory(analysis.Lowering)
	mergeLowering(&report.Lowering, "backend", backendInventory)
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
	inferenceOptions, nativeEvaluations, err := prepareNativeApplications(&report, metadata)
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
		profile := contracts.Profile{Name: name, Version: contracts.SchemaVersion, Scope: "whole_region"}
		obligations := checkContractProfile(analysis, report.Summary, profile, nil)
		report.Profiles = append(report.Profiles, contracts.ProfileResult{Profile: profile, Status: contracts.ObligationStatus(obligations), Obligations: obligations})
		report.Obligations = append(report.Obligations, obligations...)
	}
	appendNativeEvaluations(&report, nativeEvaluations, metadata.Trust)
	if err := applyContractRequests(&report, analysis, metadata, inferenceOptions); err != nil {
		return report, contractMetadataError{err}
	}
	return report, nil
}

func contractText(report contracts.Report) []byte {
	var output []byte
	output = fmt.Appendf(output, "L Contracts %s — %s\nResolution: %s\nScope: %s (%s), %s\nCoverage: %s; native transitive closure: %s\n", report.SchemaVersion, report.Entry, report.ResolutionPolicy, report.Scope.Boundary, report.Scope.Platform, report.Scope.Target, report.Summary.Coverage, report.Lowering.TransitiveCoverage)
	for _, scope := range report.Scopes {
		output = fmt.Appendf(output, "Scope %s: %s at %s:%d meaning=%s target=%s\n", scope.ID, scope.Boundary, scope.Origin.Source, scope.Origin.Start, scope.Meaning, scope.Target.String())
		for _, location := range scope.Resources {
			output = fmt.Appendf(output, "  resource: %s projection=%v\n", location.Resource.ID, location.Path)
		}
	}
	for _, request := range report.Requests {
		output = fmt.Appendf(output, "Request %s: %s (%s, %s)\n", request.Request.ID, request.Status, request.Request.Profile.Name, request.Request.Scope)
	}
	if verification := report.SourceVerification; verification != nil {
		output = fmt.Appendf(output, "Source verification: %s assurance, %s scope, %s\n", verification.Assurance, verification.Scope, verification.Status)
		for _, meaning := range verification.Selected {
			output = fmt.Appendf(output, "  selected adjective: %s\n", meaning)
		}
		if verification.Assurance == "symbolic" {
			output = append(output, "Only selected source elaboration obligations determine this result. Native and runtime implementation assurance is reported separately below.\n"...)
		}
	}
	for _, certificate := range report.SourceCertificates {
		output = fmt.Appendf(output, "Source adjective %s: %s (%s, %s), target=%s nominal=%s\n", certificate.Meaning, certificate.LogicalStatus, certificate.Scope, certificate.Mode, certificate.Target.String(), certificate.TargetNominal)
	}
	for _, instance := range report.SourceInstantiations {
		output = fmt.Appendf(output, "Source adjective instance %s: %s target=%s\n", instance.Meaning, instance.Status, instance.Target.String())
	}
	for _, profile := range report.Profiles {
		output = fmt.Appendf(output, "%s: %s\n", profile.Profile.Name, profile.Status)
	}
	for _, assumption := range report.Assumptions {
		output = fmt.Appendf(output, "Assumption: %s\n", assumption)
	}
	for _, certificate := range report.Certificates {
		output = fmt.Appendf(output, "Certificate: %s (%s): %s\n", certificate.Meaning, certificate.Mode, certificate.Status)
		for _, assumption := range certificate.Assumptions {
			output = fmt.Appendf(output, "  Assumption: %s\n", assumption)
		}
	}
	for _, binding := range report.Bindings {
		output = fmt.Appendf(output, "\n%s %q -> %s\n  meaning: %s\n  target: %s\n  origin: %s:%d\n  selection: %s\n", binding.Occurrence, binding.Spelling, binding.CanonicalName, binding.Meaning, binding.Target.String(), binding.Origin.Source, binding.Origin.Start, binding.SelectionReason)
	}
	var explanations []contracts.Explanation
	for _, explanation := range contracts.ExplainReport(report) {
		if explanation.Kind != "binding" {
			explanations = append(explanations, explanation)
		}
	}
	output = append(output, contracts.FormatExplanations(explanations)...)
	var boundaries []string
	for _, boundary := range report.Lowering.Native {
		boundaries = append(boundaries, boundary.ID+": "+boundary.Reason)
	}
	sort.Strings(boundaries)
	for _, boundary := range boundaries {
		output = fmt.Appendf(output, "Native: %s\n", boundary)
	}
	return append(output, "Executed checks: none (source inspection only).\n"...)
}
