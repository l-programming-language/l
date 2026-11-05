package Lyger

import (
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/l-programming-language/l/contracts"
)

const sourceCachePath = "/project/source-cache.l"

func sourceCacheProjection(t *testing.T, analysis *ContractAnalysis) contracts.SourceAdjectiveProjection {
	t.Helper()
	for _, projection := range analysis.SourceProjections {
		if projection.Origin.Source == "source:source-cache.l" {
			return projection
		}
	}
	t.Fatal("source adjective has no compiler-derived projection")
	return contracts.SourceAdjectiveProjection{}
}

func sourceCacheDeclaration(t *testing.T, service *LygerService) contracts.SourceTransformerDeclaration {
	t.Helper()
	analysis := analyzeContractSource(t, service, []byte(sourceCompact), sourceCachePath, ContractAnalysisOptions{})
	projection := sourceCacheProjection(t, analysis)
	if len(projection.After.Children) != 1 {
		t.Fatalf("expected one constructed service, got %+v", projection.After)
	}
	child := projection.After.Children[0]
	return contracts.SourceTransformerDeclaration{
		Meaning: projection.Meaning, Version: "test-source/1", TargetNominal: projection.TargetNominal,
		Requires: projection.RequiredAssumptions,
		Provides: []contracts.Predicate{
			{Kind: "present", Subject: "child", Target: child.Target},
			{Kind: "meaning", Subject: "nominal", Target: child.Target, Value: contracts.Atom{Kind: "meaning", Text: string(child.Nominal)}},
		},
		Origin: projection.Origin,
	}
}

func sourceCacheCompare(t *testing.T, service *LygerService, source string, options ContractAnalysisOptions) *ContractAnalysis {
	t.Helper()
	warm, warmErr := service.AnalyzeContracts([]byte(source), sourceCachePath, options)
	options.Session = nil
	cold, coldErr := service.AnalyzeContracts([]byte(source), sourceCachePath, options)
	if (warmErr == nil) != (coldErr == nil) {
		t.Fatalf("warm/cold error differs: warm %v; cold %v", warmErr, coldErr)
	}
	if warmErr != nil {
		if warmErr.Error() != coldErr.Error() {
			t.Fatalf("warm/cold diagnostics differ: warm %v; cold %v", warmErr, coldErr)
		}
		return nil
	}
	// JSON is the deterministic interchange boundary; compiler-owned pointers
	// and unexported graph state are deliberately excluded from this comparison.
	view := func(analysis *ContractAnalysis) []byte {
		t.Helper()
		data := contractJSON(t, struct {
			Bindings       ContractBindings
			Certificates   []contracts.SourceAdjectiveCertificate
			Instantiations []contracts.SourceAdjectiveInstantiation
			Summary        contracts.DerivedSummary
		}{analysis.Bindings, analysis.SourceCertificates, analysis.SourceInstantiations, analysis.Summary})
		return data
	}
	if string(view(warm)) != string(view(cold)) {
		t.Fatal("warm source-adjective analysis differs from current cold analysis")
	}
	return warm
}

func TestSourceAdjectiveCacheSourceAndRequirements(t *testing.T) {
	service := NewLygerService(Options{})
	declaration := sourceCacheDeclaration(t, service)
	session := contracts.NewSession()
	options := ContractAnalysisOptions{Session: session, SourceTransformers: []contracts.SourceTransformerDeclaration{declaration}}
	first := sourceCacheCompare(t, service, sourceCompact, options)
	certificate := sourceCertificateFor(t, first, declaration.Meaning)
	if certificate.LogicalStatus != contracts.Established {
		t.Fatalf("nontrivial source theorem failed: %+v", certificate.LogicalObligations)
	}
	sourceCacheCompare(t, service, sourceCompact, options)
	stats := session.CacheStats()
	if stats.Binding.Hits == 0 || stats.Logical.Hits == 0 || stats.Implementation.Hits == 0 {
		t.Fatalf("warm run did not exercise all existing cache boundaries: %+v", stats)
	}

	t.Run("changed promise", func(t *testing.T) {
		changed := declaration
		projection := sourceCacheProjection(t, first)
		changed.Provides = []contracts.Predicate{{Kind: "eq", Subject: "children.count", Target: projection.Target, Value: contracts.Atom{Kind: "number", Number: 0}}}
		options.SourceTransformers = []contracts.SourceTransformerDeclaration{changed}
		result := sourceCacheCompare(t, service, sourceCompact, options)
		updated := sourceCertificateFor(t, result, changed.Meaning)
		if updated.LogicalStatus == contracts.Established || updated.DeclarationFingerprint == certificate.DeclarationFingerprint {
			t.Fatal("changed false promise reused successful source theorem")
		}
	})

	t.Run("changed source", func(t *testing.T) {
		options.SourceTransformers = []contracts.SourceTransformerDeclaration{declaration}
		// Equal-width replacement preserves the authored rule's source extent.
		// A missing declaration match is also conservative; a stale pass is not.
		changed := strings.Replace(sourceCompact, "Main Service", "Main Object ", 1)
		result := sourceCacheCompare(t, service, changed, options)
		if result != nil {
			updated := sourceCertificateFor(t, result, declaration.Meaning)
			if updated.LogicalStatus == contracts.Established || updated.ProjectionFingerprint == certificate.ProjectionFingerprint {
				t.Fatal("changed source body reused successful old structural theorem")
			}
		}
	})

	t.Run("new lexical candidate", func(t *testing.T) {
		options.SourceTransformers = []contracts.SourceTransformerDeclaration{declaration}
		before := session.CacheStats()
		result := sourceCacheCompare(t, service, sourceCompact+"Define \"AdditionalHelper\" as `new candidate`.\n", options)
		if result == nil || sourceCertificateFor(t, result, declaration.Meaning).LogicalStatus != contracts.Established {
			t.Fatal("unrelated candidate changed the selected structural behavior")
		}
		if reflect.DeepEqual(first.Bindings.Inputs, result.Bindings.Inputs) || session.CacheStats().Binding.Misses <= before.Binding.Misses {
			t.Fatal("candidate-environment change did not invalidate binding independently")
		}
	})
}

func TestSourceAdjectiveCacheNativeAndTrust(t *testing.T) {
	files := dictionaryIndexFilesystem(t)
	service := NewLygerService(Options{Resources: files})
	declaration := sourceCacheDeclaration(t, service)
	session := contracts.NewSession()
	options := ContractAnalysisOptions{Session: session, SourceTransformers: []contracts.SourceTransformerDeclaration{declaration}}
	first := sourceCacheCompare(t, service, sourceCompact, options)
	certificate := sourceCertificateFor(t, first, declaration.Meaning)
	if certificate.LogicalStatus != contracts.Established || len(certificate.LoweringObligations) == 0 {
		t.Fatal("expected structural theorem with independent unresolved native lowering")
	}
	core := files["builtins/eng/core.l"]
	original := string(core.Data)
	changed := strings.Replace(original, `[["A", "<@Function@>"]]`, `[["B", "<@Function@>"]]`, 1)
	if changed == original {
		t.Fatal("native-wrapper fault injection did not select Objective's actual template")
	}
	core.Data = []byte(changed)
	before := session.CacheStats()
	updated := sourceCacheCompare(t, service, sourceCompact, options)
	actual := sourceCertificateFor(t, updated, declaration.Meaning)
	if actual.LogicalStatus != contracts.Established || len(actual.LoweringObligations) == 0 {
		t.Fatal("native bytes must not turn a structural theorem into native approval")
	}
	if session.CacheStats().Implementation.Misses <= before.Implementation.Misses {
		t.Fatal("changed native implementation reused old concrete assurance")
	}
	for _, obligation := range actual.LoweringObligations {
		if obligation.Status == contracts.Established {
			t.Fatal("opaque native implementation was approved by structural checking")
		}
	}

	before = session.CacheStats()
	options.Inference = contracts.AnalysisOptions{
		NativeWitnesses:     map[string]string{"test-only-unapproved-boundary": "changed-witness"},
		ApprovedAssumptions: map[string]bool{"test-only-unused-assumption": true},
	}
	trusted := sourceCacheCompare(t, service, sourceCompact, options)
	if sourceCertificateFor(t, trusted, declaration.Meaning).LogicalStatus != contracts.Established {
		t.Fatal("irrelevant trust change altered structural proof")
	}
	if session.CacheStats().Implementation.Misses <= before.Implementation.Misses {
		t.Fatal("trust/witness change did not invalidate implementation assurance")
	}
}

func TestSourceAdjectiveDictionaryHelperIsolation(t *testing.T) {
	files := dictionaryIndexFilesystem(t)
	files["stdlib/eng/io/dictionary.l"] = &fstest.MapFile{Data: []byte("Create \"IO\" as a Dictionary.\n\"Trimmed\" describes a Generic `dictionary_owned` Service.\n\"Compact\" describes a Generic Trimmed Service.\n")}
	source := "Use IO from `io`.\n\"Trimmed\" describes a Generic `caller_owned` Service.\n\"Wrapped\" describes a Generic Trimmed Service.\n"
	analysis := analyzeContractSource(t, NewLygerService(Options{Resources: files}), []byte(source), sourceCachePath, ContractAnalysisOptions{})
	projection := sourceProjectionNamed(t, analysis, "Compact")
	if projection.After.Name != "dictionary_owned" || projection.Origin.Source != "resource:stdlib/eng/io/dictionary.l" {
		t.Fatalf("dictionary adjective helper was captured by caller: %+v", projection)
	}
	caller := sourceProjectionNamed(t, analysis, "Wrapped")
	if caller.After.Name != "caller_owned" || caller.Environment == projection.Environment || len(projection.Dependencies) == 0 {
		t.Fatal("dictionary helper provenance lost its definition-site environment")
	}
}
