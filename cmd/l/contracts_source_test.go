package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/l-programming-language/l/contracts"
)

func TestContractsSourceVerificationArguments(t *testing.T) {
	for _, test := range []struct {
		name string
		args []string
	}{
		{"unknown-assurance", []string{"verify-dictionary", "--assurance", "optimistic", "io"}},
		{"symbolic-inspect", []string{"inspect", "--assurance", "symbolic", "main.l"}},
		{"symbolic-check", []string{"check", "--assurance", "symbolic", "--profile", "NoBlocking", "main.l"}},
		{"symbolic-without-selection", []string{"verify-dictionary", "--assurance", "symbolic", "io"}},
		{"selection-outside-dictionary", []string{"inspect", "--adjective", "meaning", "main.l"}},
		{"empty-selection", []string{"verify-dictionary", "--assurance", "symbolic", "--adjective", "", "io"}},
		{"symbolic-whole-entry-profile", []string{"verify-dictionary", "--assurance", "symbolic", "--adjective", "meaning", "--profile", "NoBlocking", "io"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := parseContractsCommand(test.args); err == nil {
				t.Fatalf("accepted invalid source verification request: %v", test.args)
			}
		})
	}
	command, err := parseContractsCommand([]string{"verify-dictionary", "--assurance", "symbolic", "--adjective", "meaning:one", "--adjective", "source:library.l:42", "library.l"})
	if err != nil || command.Assurance != "symbolic" || len(command.Adjectives) != 2 {
		t.Fatalf("valid selected symbolic request: %+v, %v", command, err)
	}
	command, err = parseContractsCommand([]string{"verify-dictionary", "io"})
	if err != nil || command.Assurance != "implementation" || len(command.Adjectives) != 0 {
		t.Fatalf("whole dictionary default changed: %+v, %v", command, err)
	}
	if !strings.Contains(contractsUsage, "--assurance symbolic") || !strings.Contains(contractsUsage, "--adjective") {
		t.Fatal("contracts help omits selected assurance scope")
	}
}

func TestContractsAdjectiveSelection(t *testing.T) {
	definitions := []contracts.DefinitionRecord{
		{Meaning: "first", Descriptor: true, Origin: contracts.Origin{Source: "source:dictionary.l", Start: 10}},
		{Meaning: "second", Descriptor: true, Origin: contracts.Origin{Source: "source:dictionary.l", Start: 10}},
		{Meaning: "alias-first", Descriptor: false, Origin: contracts.Origin{Source: "source:dictionary.l", Start: 20}},
	}
	for _, selector := range []string{"missing", "source:dictionary.l:10", "alias-first"} {
		if _, err := selectContractAdjectives(definitions, []string{selector}); err == nil {
			t.Fatalf("invalid or ambiguous selector accepted: %q", selector)
		}
	}
	selected, err := selectContractAdjectives(definitions, []string{"first", "second", "first"})
	if err != nil || len(selected) != 2 || selected[0] != "first" || selected[1] != "second" {
		t.Fatalf("canonical selection/deduplication changed: %v, %v", selected, err)
	}
}

func TestContractsSourceExplanationKeepsSelectedRule(t *testing.T) {
	report := contracts.Report{Definitions: []contracts.DefinitionRecord{{Definition: "descriptor", Meaning: "service-rule", Descriptor: true}, {Definition: "descriptor", Meaning: "function-rule", Descriptor: true}}, SourceCertificates: []contracts.SourceAdjectiveCertificate{{Meaning: "service-rule"}, {Meaning: "function-rule"}}}
	certificates, _ := relevantSourceExplanations(report, []contracts.ResolvedBinding{{Definition: "descriptor", Meaning: "service-rule"}})
	if len(certificates) != 1 || certificates[0].Meaning != "service-rule" {
		t.Fatal("explaining a selected rule included different attachment domains")
	}
	certificates, _ = relevantSourceExplanations(report, []contracts.ResolvedBinding{{Definition: "descriptor", Meaning: "descriptor-base"}})
	if len(certificates) != 2 {
		t.Fatal("descriptor declaration explanation hid its separate rules")
	}
}

func sourceCLIReport(t *testing.T, args ...string) (contracts.Report, []byte, int) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	status := run(context.Background(), append([]string{"contracts"}, args...), &stdout, &stderr)
	var report contracts.Report
	if status != 2 {
		if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
			t.Fatalf("source report exit=%d error=%s output=%s: %v", status, stderr.String(), stdout.String(), err)
		}
	}
	return report, stdout.Bytes(), status
}

func sourceCLIFixture(t *testing.T, additionalSource ...string) (string, contracts.SourceAdjectiveCertificate) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "adjective.l")
	if err := os.WriteFile(path, []byte("\"Named\" describes a Generic `selected` App.\n"+strings.Join(additionalSource, "")), 0600); err != nil {
		t.Fatal(err)
	}
	report, _, status := sourceCLIReport(t, "inspect", "--format", "json", path)
	if status != 0 {
		t.Fatalf("source definition inspection exit %d", status)
	}
	var meaning contracts.MeaningKey
	for _, definition := range report.Definitions {
		if definition.Name == "Named" && definition.Descriptor {
			meaning = definition.Meaning
		}
	}
	for _, certificate := range report.SourceCertificates {
		if certificate.Meaning == meaning {
			if certificate.After.Name != "selected" || !certificate.After.NameKnown || certificate.TargetNominal == "" {
				t.Fatalf("source fixture did not project its actual name/nominal change: %+v", certificate)
			}
			return path, certificate
		}
	}
	t.Fatal("source rule certificate missing from inspection")
	return "", contracts.SourceAdjectiveCertificate{}
}

func sourceCLIManifest(t *testing.T, path string, certificate contracts.SourceAdjectiveCertificate, promise string) string {
	t.Helper()
	declaration := contracts.SourceTransformerDeclaration{Meaning: certificate.Meaning, Version: "1", TargetNominal: certificate.TargetNominal, Provides: []contracts.Predicate{{Kind: "eq", Subject: "name", Target: certificate.Target, Value: contracts.Atom{Kind: "text", Text: promise}}}, Origin: contracts.Origin{Source: "test-source-contract.json"}}
	manifest := contracts.FacetRoleManifest{SchemaVersion: contracts.SchemaVersion, Kind: "facet_and_role_manifest", Producer: "source-cli-test", SourceTransformers: []contracts.SourceTransformerDeclaration{declaration}}
	data, err := contracts.EncodeDocument(manifest)
	if err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(filepath.Dir(path), "source-contract.json")
	if err := os.WriteFile(output, data, 0600); err != nil {
		t.Fatal(err)
	}
	return output
}

func TestContractsSourceSymbolicVerification(t *testing.T) {
	path, certificate := sourceCLIFixture(t)
	manifest := sourceCLIManifest(t, path, certificate, "selected")
	args := []string{"verify-dictionary", "--format", "json", "--manifest", manifest, "--assurance", "symbolic", "--adjective", string(certificate.Meaning), path}
	report, first, status := sourceCLIReport(t, args...)
	if status != 0 || report.SourceVerification == nil || report.SourceVerification.Status != contracts.Established || report.SourceVerification.Scope != "source_elaboration" || report.SourceVerification.Assurance != "symbolic" {
		t.Fatalf("valid source theorem did not verify in its explicit scope: exit=%d verification=%+v", status, report.SourceVerification)
	}
	if len(report.Executions) != 0 || contracts.AllEstablished(report.Obligations) || report.Lowering.TransitiveCoverage != contracts.Unknown {
		t.Fatal("symbolic theorem erased native unknowns or fabricated execution")
	}
	_, second, repeatedStatus := sourceCLIReport(t, args...)
	if repeatedStatus != 0 || !bytes.Equal(first, second) {
		t.Fatal("selected symbolic report is not deterministic")
	}
	_, _, status = sourceCLIReport(t, "verify-dictionary", "--format", "json", "--manifest", manifest, "--assurance", "symbolic", "--adjective", "source:adjective.l:0", path)
	if status != 0 {
		t.Fatalf("unique declaration source selector exit %d", status)
	}
	whole, _, status := sourceCLIReport(t, "verify-dictionary", "--format", "json", "--manifest", manifest, "--adjective", string(certificate.Meaning), path)
	if status != 1 || whole.SourceVerification == nil || whole.SourceVerification.Scope != "whole_dictionary" || whole.SourceVerification.Status == contracts.Established {
		t.Fatal("default implementation verification silently narrowed its obligations")
	}
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil || len(entries) != 2 {
		t.Fatalf("verification created build/publication artifacts: %v, %v", entries, err)
	}
	var output bytes.Buffer
	if err := writeContractText(&output, report); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"Source verification: symbolic assurance", "source_elaboration, symbolic", string(certificate.Meaning), "Native:"} {
		if !strings.Contains(output.String(), expected) {
			t.Fatalf("text report omitted narrow scope or residual native assurance %q", expected)
		}
	}
}

func TestContractsSourceVerificationFailures(t *testing.T) {
	path, certificate := sourceCLIFixture(t)
	manifest := sourceCLIManifest(t, path, certificate, "different")
	report, _, status := sourceCLIReport(t, "verify-dictionary", "--format", "json", "--manifest", manifest, "--assurance", "symbolic", "--adjective", string(certificate.Meaning), path)
	if status != 1 || report.SourceVerification == nil || report.SourceVerification.Status != contracts.Contradicted {
		t.Fatalf("false source promise passed: exit=%d %+v", status, report.SourceVerification)
	}
	report, _, status = sourceCLIReport(t, "verify-dictionary", "--format", "json", "--assurance", "symbolic", "--adjective", string(certificate.Meaning), path)
	if status != 1 || report.SourceVerification == nil || report.SourceVerification.Status != contracts.Unresolved {
		t.Fatal("discovery without a source contract became proof")
	}
	_, _, status = sourceCLIReport(t, "verify-dictionary", "--format", "json", "--assurance", "symbolic", "--adjective", "missing-canonical-meaning", path)
	if status != 2 {
		t.Fatalf("missing selection exit %d, want invocation error", status)
	}
	injected := map[string]any{"schema_version": contracts.SchemaVersion, "kind": "facet_and_role_manifest", "producer": "body-injection-test", "source_transformers": []any{map[string]any{"meaning": certificate.Meaning, "version": "1", "target_nominal": certificate.TargetNominal, "body": map[string]any{"entry": map[string]any{"kind": "event"}}}}}
	data, err := json.Marshal(injected)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifest, data, 0600); err != nil {
		t.Fatal(err)
	}
	_, _, status = sourceCLIReport(t, "verify-dictionary", "--format", "json", "--manifest", manifest, "--assurance", "symbolic", "--adjective", string(certificate.Meaning), path)
	if status != 2 {
		t.Fatal("source requirements accepted a caller-authored replacement body")
	}
}

func TestContractsAuthoredTransformerCannotCertifySource(t *testing.T) {
	path, certificate := sourceCLIFixture(t)
	manifest := contracts.FacetRoleManifest{SchemaVersion: contracts.SchemaVersion, Kind: "facet_and_role_manifest", Producer: "authored-model-test", Transformers: []contracts.ContractTransformer{{Meaning: certificate.Meaning, Version: "1", Target: certificate.Target, Body: contracts.Plan{Entry: contracts.Node{Kind: contracts.Event}}, Profiles: []contracts.Profile{{Name: "NoBlocking"}}}}}
	data, err := contracts.EncodeDocument(manifest)
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(filepath.Dir(path), "authored-model.json")
	if err := os.WriteFile(file, data, 0600); err != nil {
		t.Fatal(err)
	}
	report, _, status := sourceCLIReport(t, "verify-dictionary", "--format", "json", "--manifest", file, "--assurance", "symbolic", "--adjective", string(certificate.Meaning), path)
	if status != 1 || report.SourceVerification == nil || report.SourceVerification.Status != contracts.Unresolved {
		t.Fatal("authored typed model was substituted for source conformance")
	}
}

func TestContractsSourceInstanceExplainAndWholeEntryCheck(t *testing.T) {
	path, certificate := sourceCLIFixture(t, "Create \"Worker\" as a New Named Basic App.\n")
	manifest := sourceCLIManifest(t, path, certificate, "selected")
	report, _, status := sourceCLIReport(t, "inspect", "--format", "json", "--manifest", manifest, path)
	if status != 0 {
		t.Fatalf("source instance inspection exit %d", status)
	}
	var selected contracts.ResolvedBinding
	for _, binding := range report.Bindings {
		if binding.Meaning == certificate.Meaning && !binding.Target.Unknown && binding.Target.Root != "" {
			selected = binding
		}
	}
	if selected.Occurrence == "" {
		t.Fatal("fixture has no typed descriptor application binding")
	}
	explained, _, status := sourceCLIReport(t, "explain", "--format", "json", "--manifest", manifest, "--binding", string(selected.Occurrence), path)
	if status != 0 || len(explained.SourceCertificates) != 1 || len(explained.SourceInstantiations) != 1 || explained.SourceCertificates[0].Meaning != selected.Meaning || explained.SourceInstantiations[0].Target.Root != selected.Target.Root {
		t.Fatalf("explain lost selected source theorem or exact application: exit=%d certificates=%+v instances=%+v", status, explained.SourceCertificates, explained.SourceInstantiations)
	}
	if len(explained.SourceInstantiations[0].LoweringObligations) == 0 {
		t.Fatal("source application explanation hid native lowering residuals")
	}
	checked, _, status := sourceCLIReport(t, "check", "--format", "json", "--manifest", manifest, "--profile", "NoBlocking", path)
	if status != 1 || len(checked.Profiles) != 1 || checked.Profiles[0].Status == contracts.Established {
		t.Fatal("source topology theorem certified whole-entry runtime behavior")
	}
	var stdout, stderr bytes.Buffer
	if status := run(context.Background(), []string{"contracts", "explain", "--manifest", manifest, "--binding", string(selected.Occurrence), path}, &stdout, &stderr); status != 0 {
		t.Fatalf("text source explanation: %d %s", status, stderr.String())
	}
	for _, text := range []string{"Source adjective", "Source adjective instance", string(certificate.Meaning), "source_elaboration", "Native:"} {
		if !strings.Contains(stdout.String(), text) {
			t.Fatalf("text explanation missing %q", text)
		}
	}
}

func TestContractsSourceMetadataMustNameResolvedRule(t *testing.T) {
	path, certificate := sourceCLIFixture(t)
	certificate.Meaning = "absent-source-rule"
	manifest := sourceCLIManifest(t, path, certificate, "selected")
	_, _, status := sourceCLIReport(t, "inspect", "--format", "json", "--manifest", manifest, path)
	if status != 2 {
		t.Fatalf("source metadata with absent canonical identity exit %d", status)
	}
}
