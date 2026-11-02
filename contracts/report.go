package contracts

import (
	"fmt"
	"time"
)

type ProfileResult struct {
	Profile     Profile      `json:"profile"`
	Status      Status       `json:"status"`
	Obligations []Obligation `json:"obligations"`
}

type ReportScope struct {
	Phase       Phase    `json:"phase"`
	Boundary    string   `json:"boundary"`
	Target      string   `json:"target"`
	Platform    string   `json:"platform,omitempty"`
	Compiler    string   `json:"compiler,omitempty"`
	PanicPolicy string   `json:"panic_policy,omitempty"`
	Features    []string `json:"features,omitempty"`
	Exclusions  []string `json:"exclusions,omitempty"`
}

type ExecutionRecord struct {
	ID                string   `json:"id"`
	Kind              string   `json:"kind"`
	Command           []string `json:"command"`
	WorkingDirectory  string   `json:"working_directory"`
	Status            string   `json:"status"`
	Started           string   `json:"started,omitempty"`
	Finished          string   `json:"finished,omitempty"`
	ExitCode          *int     `json:"exit_code,omitempty"`
	StdoutFingerprint string   `json:"stdout_fingerprint,omitempty"`
	StderrFingerprint string   `json:"stderr_fingerprint,omitempty"`
	RecordFingerprint string   `json:"record_fingerprint,omitempty"`
}

// SealExecutionRecord binds a caller-supplied actual execution record. It does
// not execute a tool, authenticate arbitrary external reports, or upgrade tests
// into proofs. Only explicit recorded operations belong in Report.Executions.
func SealExecutionRecord(record ExecutionRecord) ExecutionRecord {
	record.RecordFingerprint = ""
	record.RecordFingerprint = StableID("execution", record)
	return record
}

func (record ExecutionRecord) Validate() error {
	if record.ID == "" || !containsString([]string{"build", "test", "benchmark", "proof", "sanitizer", "integrity", "schema"}, record.Kind) || len(record.Command) == 0 || record.Command[0] == "" || record.WorkingDirectory == "" {
		return fmt.Errorf("execution record requires identity, operation kind, exact command and working directory")
	}
	if record.Status == "not_run" || record.Status == "blocked" {
		if record.Started != "" || record.Finished != "" || record.ExitCode != nil {
			return fmt.Errorf("unrun operation %s contains execution results", record.ID)
		}
		return nil
	}
	if record.Status != "passed" && record.Status != "failed" {
		return fmt.Errorf("unsupported execution status %q", record.Status)
	}
	started, firstErr := time.Parse(time.RFC3339Nano, record.Started)
	finished, secondErr := time.Parse(time.RFC3339Nano, record.Finished)
	if firstErr != nil || secondErr != nil || finished.Before(started) || record.ExitCode == nil || record.StdoutFingerprint == "" || record.StderrFingerprint == "" {
		return fmt.Errorf("executed operation %s lacks a complete actual execution record", record.ID)
	}
	if (record.Status == "passed") != (*record.ExitCode == 0) {
		return fmt.Errorf("execution status and exit code disagree for %s", record.ID)
	}
	if record.RecordFingerprint == "" || SealExecutionRecord(record).RecordFingerprint != record.RecordFingerprint {
		return fmt.Errorf("execution record %s has missing or mismatched provenance fingerprint", record.ID)
	}
	return nil
}

type Report struct {
	SchemaVersion        string                         `json:"schema_version"`
	Kind                 string                         `json:"kind"`
	Producer             string                         `json:"producer"`
	Entry                string                         `json:"entry"`
	Revision             string                         `json:"revision,omitempty"`
	ResolutionPolicy     string                         `json:"resolution_policy"`
	Inputs               map[string]string              `json:"inputs"`
	Bindings             []ResolvedBinding              `json:"bindings"`
	ReferenceSelections  []ReferenceSelection           `json:"reference_selections,omitempty"`
	Definitions          []DefinitionRecord             `json:"definitions,omitempty"`
	Targets              []TargetRef                    `json:"targets"`
	Declarations         []DeclaredContract             `json:"declarations,omitempty"`
	Summary              DerivedSummary                 `json:"summary"`
	Lifecycle            []AppProtocolState             `json:"lifecycle,omitempty"`
	Lowering             LoweringInventory              `json:"lowering"`
	Obligations          []Obligation                   `json:"obligations"`
	Profiles             []ProfileResult                `json:"profiles,omitempty"`
	Evidence             EvidenceGraph                  `json:"evidence"`
	NativeWitnesses      []NativeWitness                `json:"native_witnesses,omitempty"`
	TrustPolicy          *TrustPolicy                   `json:"trust_policy,omitempty"`
	Assumptions          []string                       `json:"assumptions,omitempty"`
	Executions           []ExecutionRecord              `json:"executions,omitempty"`
	Scope                ReportScope                    `json:"scope"`
	Configuration        *ConfigurationResult           `json:"configuration,omitempty"`
	Certificates         []AdjectiveCertificate         `json:"certificates,omitempty"`
	SourceCertificates   []SourceAdjectiveCertificate   `json:"source_certificates,omitempty"`
	SourceDeclarations   []SourceTransformerDeclaration `json:"source_declarations,omitempty"`
	SourceInstantiations []SourceAdjectiveInstantiation `json:"source_instantiations,omitempty"`
	SourceVerification   *SourceVerificationResult      `json:"source_verification,omitempty"`
}

func NewReport(entry string) Report {
	return Report{SchemaVersion: SchemaVersion, Kind: "contract_report", Producer: AnalyzerVersion, Entry: entry, ResolutionPolicy: TrackAPolicy, Inputs: make(map[string]string), Summary: DerivedSummary{Coverage: Unknown}, Lowering: LoweringInventory{BoundaryCoverage: Unknown, TransitiveCoverage: Unknown}, Scope: ReportScope{Phase: Invocation, Boundary: "whole_region", Target: "unspecified"}}
}

func (report Report) Validate() error {
	if err := validateDocumentHeader(report.SchemaVersion, report.Kind, "contract_report", report.Producer); err != nil {
		return err
	}
	if report.Entry == "" || report.ResolutionPolicy == "" || report.Scope.Boundary == "" || report.Scope.Target == "" {
		return fmt.Errorf("contract report requires entry, resolution policy and explicit observation/target scope")
	}
	if err := validateModelEnums(report); err != nil {
		return err
	}
	if err := report.Lowering.Validate(); err != nil {
		return err
	}
	seen := make(map[BindingOccurrenceID]bool)
	for _, binding := range report.Bindings {
		if binding.Occurrence == "" || seen[binding.Occurrence] || binding.Meaning == "" || binding.Definition == "" {
			return fmt.Errorf("report contains empty, duplicate or ambiguous canonical binding identity")
		}
		seen[binding.Occurrence] = true
	}
	executions := make(map[string]ExecutionRecord)
	for _, execution := range report.Executions {
		if err := execution.Validate(); err != nil {
			return err
		}
		if _, exists := executions[execution.ID]; exists {
			return fmt.Errorf("duplicate execution identity %s", execution.ID)
		}
		executions[execution.ID] = execution
	}
	claims := make(map[string]bool)
	for _, claim := range report.Evidence.Claims {
		if claim.ID == "" || claims[claim.ID] {
			return fmt.Errorf("duplicate or empty evidence claim identity")
		}
		claims[claim.ID] = true
	}
	nodes := make(map[string]bool)
	for _, node := range report.Evidence.Nodes {
		if node.ID == "" || nodes[node.ID] || !claims[node.Claim] {
			return fmt.Errorf("evidence node has duplicate/empty identity or missing proposition")
		}
		nodes[node.ID] = true
		if node.Kind == "empirical" || node.Kind == "external_proof" {
			execution, exists := executions[node.Witness]
			if !exists || execution.Status != "passed" && execution.Status != "failed" {
				return fmt.Errorf("evidence %s claims executed verification without an execution record", node.ID)
			}
			if execution.Kind == "schema" || execution.Kind == "integrity" || node.Kind == "external_proof" && execution.Kind != "proof" {
				return fmt.Errorf("evidence %s misrepresents the recorded operation's scope", node.ID)
			}
		}
	}
	for _, node := range report.Evidence.Nodes {
		for _, premise := range node.Premises {
			if !nodes[premise] {
				return fmt.Errorf("evidence %s references missing premise %s", node.ID, premise)
			}
		}
	}
	obligations := append([]Obligation(nil), report.Obligations...)
	obligations = append(obligations, report.Summary.Obligations...)
	sourceDeclarations := make(map[MeaningKey]SourceTransformerDeclaration)
	for _, declaration := range report.SourceDeclarations {
		if declaration.Meaning == "" || declaration.Version == "" || declaration.TargetNominal == "" {
			return fmt.Errorf("source declaration lacks canonical identity")
		}
		if _, exists := sourceDeclarations[declaration.Meaning]; exists {
			return fmt.Errorf("duplicate source declaration %s", declaration.Meaning)
		}
		sourceDeclarations[declaration.Meaning] = declaration
	}
	for _, certificate := range report.SourceCertificates {
		if declaration, exists := sourceDeclarations[certificate.Meaning]; exists && StableID("source-declaration", declaration) != certificate.DeclarationFingerprint {
			return fmt.Errorf("source certificate declaration fingerprint is stale")
		}
		if certificate.Meaning == "" || certificate.Scope != "source_elaboration" || certificate.Version != SourceProjectionVersion || certificate.ProjectionFingerprint == "" || certificate.DeclarationFingerprint == "" {
			return fmt.Errorf("source certificate lacks canonical provenance or explicit elaboration scope")
		}
		if certificate.Mode != "symbolic" && certificate.Mode != "instantiation_only" || certificate.LogicalStatus == Established && certificate.Mode != "symbolic" {
			return fmt.Errorf("source certificate has unsupported mode or claims an instantiation-only symbolic success")
		}
		if certificate.LogicalStatus == Established && (len(certificate.LogicalObligations) == 0 || !AllEstablished(certificate.LogicalObligations)) {
			return fmt.Errorf("source certificate hides unresolved logical obligations")
		}
		if certificate.LogicalStatus == Established && (certificate.Target.Unknown || certificate.Target.Root == "" || certificate.Target.Phase != Elaboration || certificate.TargetNominal == "" || certificate.Environment == "") {
			return fmt.Errorf("source certificate success lacks an exact nominal elaboration target and binding environment")
		}
		obligations = append(obligations, certificate.LogicalObligations...)
		obligations = append(obligations, certificate.LoweringObligations...)
	}
	for _, instance := range report.SourceInstantiations {
		if instance.Status == Established && (len(instance.Obligations) == 0 || !AllEstablished(instance.Obligations)) {
			return fmt.Errorf("source instantiation hides unresolved obligations")
		}
		obligations = append(obligations, instance.Obligations...)
		obligations = append(obligations, instance.LoweringObligations...)
	}
	if verification := report.SourceVerification; verification != nil {
		if verification.Assurance != "symbolic" && verification.Assurance != "implementation" || verification.Assurance == "symbolic" && verification.Scope != "source_elaboration" || verification.Assurance == "implementation" && verification.Scope != "whole_dictionary" {
			return fmt.Errorf("unsupported source verification assurance or scope")
		}
		if verification.Status == Established && (len(verification.Selected) == 0 || len(verification.Obligations) == 0 || !AllEstablished(verification.Obligations)) {
			return fmt.Errorf("source verification lacks selected established obligations")
		}
		if verification.Status == Established && verification.Assurance == "symbolic" {
			for _, meaning := range verification.Selected {
				found := false
				for _, certificate := range report.SourceCertificates {
					found = found || certificate.Meaning == meaning && certificate.LogicalStatus == Established
				}
				if !found {
					return fmt.Errorf("symbolic verification names a missing or unestablished source certificate %s", meaning)
				}
			}
		}
		if verification.Status == Established && verification.Assurance == "implementation" {
			if err := report.validateSourceImplementationAssurance(obligations); err != nil {
				return err
			}
		}
		obligations = append(obligations, verification.Obligations...)
	}
	for _, result := range report.Profiles {
		if len(result.Obligations) == 0 {
			return fmt.Errorf("profile result %s has no checked obligations", result.Profile.Name)
		}
		if result.Status == Established && !AllEstablished(result.Obligations) {
			return fmt.Errorf("profile result %s hides unresolved or contradicted obligations", result.Profile.Name)
		}
		obligations = append(obligations, result.Obligations...)
	}
	for _, obligation := range obligations {
		for _, reference := range obligation.Evidence {
			if !nodes[reference] {
				return fmt.Errorf("obligation %s references missing evidence node %s", obligation.ID, reference)
			}
		}
	}
	return nil
}

func (report Report) validateSourceImplementationAssurance(obligations []Obligation) error {
	if !AllEstablished(obligations) || len(report.Summary.Unknowns) != 0 || len(report.Lowering.Unknowns) != 0 || !sourceCompleteCoverage(report.Summary.Coverage) || !sourceCompleteCoverage(report.Lowering.BoundaryCoverage) || !sourceCompleteCoverage(report.Lowering.TransitiveCoverage) {
		return fmt.Errorf("whole-dictionary implementation assurance hides unresolved analysis or native closure")
	}
	for _, boundary := range append(append([]NativeBoundary(nil), report.Summary.NativeBoundaries...), report.Lowering.Native...) {
		if boundary.Status != Established {
			return fmt.Errorf("whole-dictionary implementation assurance hides native boundary %s", boundary.ID)
		}
	}
	for _, path := range report.Summary.Paths {
		if len(path.Unknowns) != 0 || !AllEstablished(path.Obligations) {
			return fmt.Errorf("whole-dictionary implementation assurance hides an unresolved execution path")
		}
		for _, boundary := range path.NativeBoundaries {
			if boundary.Status != Established {
				return fmt.Errorf("whole-dictionary implementation assurance hides path native boundary %s", boundary.ID)
			}
		}
	}
	for _, certificate := range report.SourceCertificates {
		if certificate.LogicalStatus != Established {
			return fmt.Errorf("whole-dictionary implementation assurance hides unresolved source certificate %s", certificate.Meaning)
		}
	}
	for _, instance := range report.SourceInstantiations {
		if instance.Status != Established {
			return fmt.Errorf("whole-dictionary implementation assurance hides unresolved source instantiation %s", instance.Meaning)
		}
	}
	for _, certificate := range report.Certificates {
		if certificate.Status != Established || !AllEstablished(certificate.Obligations) {
			return fmt.Errorf("whole-dictionary implementation assurance hides unresolved adjective certificate %s", certificate.Meaning)
		}
	}
	for _, profile := range report.Profiles {
		if profile.Status != Established || len(profile.Obligations) == 0 || !AllEstablished(profile.Obligations) {
			return fmt.Errorf("whole-dictionary implementation assurance hides unresolved profile %s", profile.Profile.Name)
		}
	}
	if report.Configuration != nil && !AllEstablished(report.Configuration.Obligations) {
		return fmt.Errorf("whole-dictionary implementation assurance hides unresolved configuration")
	}
	return nil
}

func sourceCompleteCoverage(coverage Coverage) bool {
	return coverage == Known || coverage == NotApplicable
}
