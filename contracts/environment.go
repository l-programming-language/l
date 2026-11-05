package contracts

import (
	"fmt"
	"strings"
)

type NativeToolIdentity struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	SHA256  string `json:"sha256"`
	Version string `json:"version"`
}

type NativeArtifact struct {
	Path      string `json:"path"`
	SHA256    string `json:"sha256"`
	Inputs    string `json:"inputs"`
	Execution string `json:"execution"`
}

// NativeEnvironment is observed provenance, never approval. Offline consumers
// validate source/configuration applicability; they do not attest a live host.
type NativeEnvironment struct {
	SchemaVersion string               `json:"schema_version"`
	Kind          string               `json:"kind"`
	Producer      string               `json:"producer"`
	Version       string               `json:"version"`
	Method        string               `json:"method"`
	Host          string               `json:"host"`
	Target        string               `json:"target"`
	Compiler      string               `json:"compiler"`
	PanicPolicy   string               `json:"panic_policy"`
	Runtime       string               `json:"runtime"`
	Allocator     string               `json:"allocator"`
	Configuration map[string]string    `json:"configuration"`
	Inputs        map[string]string    `json:"inputs"`
	Dependencies  map[string]string    `json:"dependencies"`
	Tools         []NativeToolIdentity `json:"tools"`
	Executions    []ExecutionRecord    `json:"executions"`
	Artifacts     []NativeArtifact     `json:"artifacts,omitempty"`
	Limitations   []string             `json:"limitations"`
	Fingerprint   string               `json:"fingerprint"`
}

func SealNativeEnvironment(value NativeEnvironment) NativeEnvironment {
	value.Fingerprint = ""
	value.Fingerprint = StableID("native-environment", value)
	return value
}

func (value NativeEnvironment) Validate() error {
	if err := validateDocumentHeader(value.SchemaVersion, value.Kind, "native_environment", value.Producer); err != nil {
		return err
	}
	expectedVersion := "1"
	if value.Method == "build" {
		expectedVersion = "2"
	}
	if value.Version != expectedVersion || !containsString([]string{"capture", "build", "test-only"}, value.Method) || value.Host == "" || value.Target == "" || value.Compiler == "" || value.PanicPolicy == "" || value.Runtime == "" || value.Allocator == "" || len(value.Inputs) == 0 {
		return fmt.Errorf("native environment requires a supported version, method and explicit configuration/input identities")
	}
	generation, placed := value.Configuration["companion-generation"]
	if placed != (value.Method == "build") || placed && (len(generation) == 0 || len(generation) > 20 || strings.Trim(generation, "0123456789") != "") {
		return fmt.Errorf("only build evidence carries a required decimal companion generation")
	}
	if value.Fingerprint == "" || SealNativeEnvironment(value).Fingerprint != value.Fingerprint {
		return fmt.Errorf("native environment fingerprint is missing or stale")
	}
	if value.Method != "test-only" && len(value.Tools) == 0 {
		return fmt.Errorf("native capture requires observed tool identities")
	}
	for _, tool := range value.Tools {
		if tool.Name == "" || tool.Path == "" || tool.SHA256 == "" || tool.Version == "" {
			return fmt.Errorf("native tool identity is incomplete")
		}
	}
	for _, execution := range value.Executions {
		if err := execution.Validate(); err != nil {
			return err
		}
	}
	for _, artifact := range value.Artifacts {
		found := false
		for _, execution := range value.Executions {
			found = found || execution.ID == artifact.Execution && execution.Kind == "build" && execution.Status == "passed"
		}
		if artifact.Path == "" || artifact.SHA256 == "" || artifact.Inputs == "" || artifact.Inputs != value.Configuration["workspace-inputs"] || !found {
			return fmt.Errorf("native artifact lacks a successful recorded build and input identity")
		}
	}
	if value.Method == "capture" && len(value.Artifacts) > 0 {
		return fmt.Errorf("environment-only capture cannot contain build artifacts")
	}
	return nil
}

// EnvironmentApplicability checks current compiler inputs without executing the
// commands recorded in the document or probing the historical tool installation.
func EnvironmentApplicability(value NativeEnvironment, inputs, dependencies map[string]string) (Status, string) {
	if err := value.Validate(); err != nil {
		return Unresolved, "native environment is invalid: " + err.Error()
	}
	for name, actual := range inputs {
		if value.Inputs[name] != actual {
			return Unresolved, "native environment source input changed or is absent: " + name
		}
	}
	for name, actual := range dependencies {
		if value.Dependencies[name] != actual {
			return Unresolved, "native environment implementation dependency changed or is absent: " + name
		}
	}
	if len(value.Inputs) != len(inputs) || len(value.Dependencies) != len(dependencies) {
		return Unresolved, "native environment selected input/dependency set changed"
	}
	return Established, "conditional on the supplied fingerprinted native environment; no live toolchain or deployment observation was performed"
}
