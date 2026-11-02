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
	SchemaVersion       string                 `json:"schema_version"`
	Kind                string                 `json:"kind"`
	Producer            string                 `json:"producer"`
	Entry               string                 `json:"entry"`
	Revision            string                 `json:"revision,omitempty"`
	ResolutionPolicy    string                 `json:"resolution_policy"`
	Inputs              map[string]string      `json:"inputs"`
	Bindings            []ResolvedBinding      `json:"bindings"`
	ReferenceSelections []ReferenceSelection   `json:"reference_selections,omitempty"`
	Definitions         []DefinitionRecord     `json:"definitions,omitempty"`
	Targets             []TargetRef            `json:"targets"`
	Declarations        []DeclaredContract     `json:"declarations,omitempty"`
	Summary             DerivedSummary         `json:"summary"`
	Lifecycle           []AppProtocolState     `json:"lifecycle,omitempty"`
	Lowering            LoweringInventory      `json:"lowering"`
	Obligations         []Obligation           `json:"obligations"`
	Profiles            []ProfileResult        `json:"profiles,omitempty"`
	Evidence            EvidenceGraph          `json:"evidence"`
	NativeWitnesses     []NativeWitness        `json:"native_witnesses,omitempty"`
	TrustPolicy         *TrustPolicy           `json:"trust_policy,omitempty"`
	Assumptions         []string               `json:"assumptions,omitempty"`
	Executions          []ExecutionRecord      `json:"executions,omitempty"`
	Scope               ReportScope            `json:"scope"`
	Configuration       *ConfigurationResult   `json:"configuration,omitempty"`
	Certificates        []AdjectiveCertificate `json:"certificates,omitempty"`
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
