package Lyger

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/l-programming-language/l/contracts"
)

const sourceAdjectivePath = "/project/source-adjectives.l"
const sourceCompact = "\"Compact\" describes a Generic App with a New Main Service using any Process.\n"

func sourceProjectionNamed(t *testing.T, analysis *ContractAnalysis, name string) contracts.SourceAdjectiveProjection {
	t.Helper()
	for _, definition := range analysis.Bindings.Definitions {
		if definition.Name == name {
			for _, projection := range analysis.SourceProjections {
				if projection.Meaning == definition.Meaning {
					return projection
				}
			}
		}
	}
	t.Fatalf("no source projection for %s", name)
	return contracts.SourceAdjectiveProjection{}
}

func sourceDeclarationFor(projection contracts.SourceAdjectiveProjection) contracts.SourceTransformerDeclaration {
	declaration := contracts.SourceTransformerDeclaration{Meaning: projection.Meaning, Version: "source-test/1", TargetNominal: projection.TargetNominal, Requires: projection.RequiredAssumptions, Origin: projection.Origin}
	if len(projection.After.Children) > 0 {
		declaration.Provides = []contracts.Predicate{{Kind: "present", Subject: "child", Target: projection.After.Children[0].Target}}
	} else {
		declaration.Provides = []contracts.Predicate{{Kind: "meaning", Subject: "nominal", Target: projection.Target, Value: contracts.Atom{Kind: "meaning", Text: string(projection.TargetNominal)}}}
	}
	return declaration
}

func sourceCertificateFor(t *testing.T, analysis *ContractAnalysis, meaning contracts.MeaningKey) contracts.SourceAdjectiveCertificate {
	t.Helper()
	for _, certificate := range analysis.SourceCertificates {
		if certificate.Meaning == meaning {
			return certificate
		}
	}
	t.Fatal("missing source certificate")
	return contracts.SourceAdjectiveCertificate{}
}

func TestSourceAdjectiveStandaloneCertificateAndCompatibility(t *testing.T) {
	service := NewLygerService(Options{})
	source := []byte(sourceCompact)
	initial := analyzeContractSource(t, service, source, sourceAdjectivePath, ContractAnalysisOptions{})
	projection := sourceProjectionNamed(t, initial, "Compact")
	if len(projection.After.Children) != 1 || len(projection.RequiredAssumptions) == 0 {
		t.Fatal("standalone source body not projected")
	}
	declaration := sourceDeclarationFor(projection)
	name := projection.Target
	name.Path = []string{"name"}
	constructor := projection.Target
	constructor.Path = []string{"constructor"}
	declaration.Frame = &contracts.FrameContract{Target: projection.Target, Allowed: []contracts.TargetRef{constructor, projection.After.Children[0].Target}, Preserves: []contracts.TargetRef{name}, Scope: "elaboration", Phase: contracts.Elaboration}
	checked := analyzeContractSource(t, service, source, sourceAdjectivePath, ContractAnalysisOptions{SourceTransformers: []contracts.SourceTransformerDeclaration{declaration}})
	certificate := sourceCertificateFor(t, checked, declaration.Meaning)
	if certificate.LogicalStatus != contracts.Established || len(certificate.LoweringObligations) == 0 {
		t.Fatalf("structural theorem or separate native obligations missing: %+v", certificate)
	}
	if len(checked.Compilation.Compilation.Semantics.Constructions) != 0 || len(checked.SourceInstantiations) != 0 {
		t.Fatal("standalone proof manufactured a concrete App")
	}
	ordinary, err := service.Compile(source, sourceAdjectivePath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(ordinary.AML, checked.Compilation.AML) || !bytes.Equal(ordinary.BackendAML, checked.Compilation.BackendAML) || !bytes.Equal(ordinary.ExpandedSource, checked.Compilation.ExpandedSource) {
		t.Fatal("source certificate changed normal compilation")
	}
	declaration.Requires = nil
	missing := analyzeContractSource(t, service, source, sourceAdjectivePath, ContractAnalysisOptions{SourceTransformers: []contracts.SourceTransformerDeclaration{declaration}})
	if sourceCertificateFor(t, missing, declaration.Meaning).LogicalStatus == contracts.Established {
		t.Fatal("empty structural premise was invented")
	}
}

func TestSourceAdjectiveDefinitionFrameRejection(t *testing.T) {
	service := NewLygerService(Options{})
	initial := analyzeContractSource(t, service, []byte(sourceCompact), sourceAdjectivePath, ContractAnalysisOptions{})
	projection := sourceProjectionNamed(t, initial, "Compact")
	declaration := sourceDeclarationFor(projection)
	declaration.Frame = &contracts.FrameContract{Target: projection.Target, Allowed: []contracts.TargetRef{projection.Target}, Preserves: []contracts.TargetRef{projection.Target}, Scope: "elaboration", Phase: contracts.Elaboration}
	checked := analyzeContractSource(t, service, []byte(sourceCompact), sourceAdjectivePath, ContractAnalysisOptions{SourceTransformers: []contracts.SourceTransformerDeclaration{declaration}})
	if sourceCertificateFor(t, checked, declaration.Meaning).LogicalStatus != contracts.Contradicted {
		t.Fatal("actual constructor writes accepted a whole-target preservation promise")
	}
}

func TestSourceAdjectiveActualInstanceUsesBeforeShape(t *testing.T) {
	service := NewLygerService(Options{})
	source := []byte(sourceCompact + "Create \"Worker\" as a New Compact App.\n")
	initial := analyzeContractSource(t, service, source, sourceAdjectivePath, ContractAnalysisOptions{})
	projection := sourceProjectionNamed(t, initial, "Compact")
	declaration := sourceDeclarationFor(projection)
	checked := analyzeContractSource(t, service, source, sourceAdjectivePath, ContractAnalysisOptions{SourceTransformers: []contracts.SourceTransformerDeclaration{declaration}})
	if len(checked.SourceInstantiations) != 1 || checked.SourceInstantiations[0].Status != contracts.Established {
		t.Fatalf("actual source attachment not verified: %+v", checked.SourceInstantiations)
	}
	// Reusing the source parser's own descriptor object twice observes the
	// genuine before-rule merge state without fabricating a proof application.
	c := checked.Compilation.Compilation
	construction := c.Semantics.Constructions[0]
	object := *construction.Root.Source
	object.adjectives = append(append([]*token(nil), object.adjectives...), object.adjectives...)
	before := len(c.contractRecorder.sourceConstructions)
	if _, err := c.constructObject(&object, make(map[*descriptorRule]bool)); err != nil {
		t.Fatal(err)
	}
	seenNonempty := false
	for _, use := range c.contractRecorder.sourceConstructions[before:] {
		if use.Descriptor.ID == "Compact" && len(use.Before.Children) > 0 {
			seenNonempty = true
			actual := c.contractRecorder.sourceShape(use.Before, projection.Target)
			status, _ := contracts.Entails(contracts.SourceShapeFacts(actual), projection.RequiredAssumptions[0])
			if status == contracts.Established {
				t.Fatal("existing children satisfied empty ordinal premise")
			}
		}
	}
	if !seenNonempty {
		t.Fatal("repeated actual constructor application was collapsed")
	}
}

func TestSourceAdjectiveCanonicalNominalAndMetadata(t *testing.T) {
	source := paragraphDescriptorWidget + "\"Outer\" describes a Generic `main` Widget of any Process with No Inputs and No Outputs.\n\n    " + paragraphDescriptorWidget + "\"Inner\" describes a Generic `main` Widget of any Process with No Inputs and No Outputs.\n"
	service := NewLygerService(Options{})
	analysis := analyzeContractSource(t, service, []byte(source), sourceAdjectivePath, ContractAnalysisOptions{})
	outer, inner := sourceProjectionNamed(t, analysis, "Outer"), sourceProjectionNamed(t, analysis, "Inner")
	if outer.Target.Type != inner.Target.Type || outer.TargetNominal == inner.TargetNominal {
		t.Fatal("same-spelled scoped nominal definitions were conflated")
	}
	declaration := sourceDeclarationFor(outer)
	declaration.TargetNominal = inner.TargetNominal
	if _, err := service.AnalyzeContracts([]byte(source), sourceAdjectivePath, ContractAnalysisOptions{SourceTransformers: []contracts.SourceTransformerDeclaration{declaration}}); !errors.Is(err, ErrSourceContractMetadata) {
		t.Fatalf("wrong canonical domain was not rejected as metadata: %v", err)
	}
}

func TestSourceAdjectiveOrderedWritesAndDefinitionSiteHelper(t *testing.T) {
	source := "\"Named\" describes a Generic `first` Service.\nIt describes a Generic `second` Service.\n\"Composite\" describes a Generic Named Service.\n\"Pair\" describes a Generic App with a New Main Service and a New Main Service.\n\n    \"Named\" describes a Generic `captured` Service.\n"
	analysis := analyzeContractSource(t, NewLygerService(Options{}), []byte(source), sourceAdjectivePath, ContractAnalysisOptions{})
	composite := sourceProjectionNamed(t, analysis, "Composite")
	if composite.After.Name != "second" {
		t.Fatalf("caller-local helper captured definition-site helper: %s", composite.After.Name)
	}
	nameWrites := 0
	for _, write := range composite.Footprints {
		if len(write.Target.Path) == 1 && write.Target.Path[0] == "name" {
			nameWrites++
		}
	}
	if nameWrites != 2 {
		t.Fatalf("ordered overwritten writes were erased: %d", nameWrites)
	}
	declaration := sourceDeclarationFor(composite)
	name := composite.Target
	name.Path = []string{"name"}
	declaration.Requires = append(declaration.Requires, contracts.Predicate{Kind: "eq", Subject: "name", Target: composite.Target, Value: contracts.Atom{Kind: "text", Text: "second"}})
	declaration.Frame = &contracts.FrameContract{Target: composite.Target, Allowed: []contracts.TargetRef{composite.Target}, Preserves: []contracts.TargetRef{name}, Scope: "elaboration", Phase: contracts.Elaboration}
	certificate, err := contracts.VerifySourceAdjective(composite, declaration, contracts.AnalysisOptions{})
	if err != nil || certificate.LogicalStatus != contracts.Contradicted {
		t.Fatal("restoring the original name erased actual typed name writes")
	}
	pair := sourceProjectionNamed(t, analysis, "Pair")
	if len(pair.After.Children) != 2 || pair.After.Children[0].Target.Path[0] == pair.After.Children[1].Target.Path[0] {
		t.Fatal("repeated canonical child construction was deduplicated")
	}
}

func TestSourceAdjectiveNativeWrapperOnlyAppends(t *testing.T) {
	source := "\"Marked\" describes any Function as `<DEF>[[\"A\", \"<@Function@>\"]]<DEF>`.\n\"Twice\" describes a Generic Marked Function.\nIt describes a Generic Marked Function.\n\"Combined\" describes a Generic Twice Function.\n"
	analysis := analyzeContractSource(t, NewLygerService(Options{}), []byte(source), sourceAdjectivePath, ContractAnalysisOptions{})
	projection := sourceProjectionNamed(t, analysis, "Marked")
	if projection.After.Constructor != "" || projection.After.NameKnown || len(projection.Footprints) != 1 || projection.Footprints[0].Target.Path[0] != "wrappers" {
		t.Fatal("raw wrapper invented constructor/name writes")
	}
	declaration := sourceDeclarationFor(projection)
	declaration.Provides = []contracts.Predicate{{Kind: "eq", Subject: "wrappers.count", Target: projection.Target, Value: contracts.Atom{Kind: "number", Number: 1}}}
	certificate, err := contracts.VerifySourceAdjective(projection, declaration, contracts.AnalysisOptions{})
	if err != nil || certificate.LogicalStatus != contracts.Established || len(certificate.LoweringObligations) == 0 {
		t.Fatalf("wrapper structure/native separation failed: %v %+v", err, certificate.LogicalObligations)
	}
	combined := sourceProjectionNamed(t, analysis, "Combined")
	if len(combined.After.Wrappers) != 2 || combined.After.Wrappers[0] != combined.After.Wrappers[1] {
		t.Fatal("repeated source wrapper applications were collapsed")
	}
	writes := 0
	for _, footprint := range combined.Footprints {
		if len(footprint.Target.Path) == 1 && footprint.Target.Path[0] == "wrappers" {
			writes++
		}
	}
	if writes != 2 {
		t.Fatalf("wrapper append order/repetition lost: %d", writes)
	}
}

func TestSourceAdjectiveSignatureAndParameterForwarding(t *testing.T) {
	source := "\"Entrypoint\" describes a Generic `entry` Function of any Process with No Inputs and No Outputs.\n"
	service := NewLygerService(Options{})
	analysis := analyzeContractSource(t, service, []byte(source), sourceAdjectivePath, ContractAnalysisOptions{})
	projection := sourceProjectionNamed(t, analysis, "Entrypoint")
	if !projection.After.InputsDefined || !projection.After.OutputsDefined || len(projection.After.Parameters) != 1 || !projection.After.Parameters[0].Formal {
		t.Fatalf("signature or Process formal lost: %+v", projection.After)
	}
	declaration := sourceDeclarationFor(projection)
	declaration.Provides = []contracts.Predicate{{Kind: "eq", Subject: "name", Target: projection.Target, Value: contracts.Atom{Kind: "text", Text: "entry"}}, {Kind: "eq", Subject: "inputs.count", Target: projection.Target, Value: contracts.Atom{Kind: "number", Number: 0}}, {Kind: "eq", Subject: "parameters.count", Target: projection.Target, Value: contracts.Atom{Kind: "number", Number: 1}}}
	checked := analyzeContractSource(t, service, []byte(source), sourceAdjectivePath, ContractAnalysisOptions{SourceTransformers: []contracts.SourceTransformerDeclaration{declaration}})
	certificate := sourceCertificateFor(t, checked, declaration.Meaning)
	if certificate.LogicalStatus != contracts.Established {
		t.Fatalf("explicit signature premises did not establish source facts: %+v", certificate.LogicalObligations)
	}
	for _, path := range certificate.Summary.Paths {
		if len(path.Demand) != 0 {
			t.Fatal("forwarding a Process was reported as invoking it")
		}
	}
}

func TestSourceAdjectiveExistingSourceChildRejectsOrdinalPremise(t *testing.T) {
	source := sourceCompact + "\"Extra\" describes a Generic App with a New `extra` Service.\nCreate \"Worker\" as a New Extra Compact App.\n"
	service := NewLygerService(Options{})
	initial := analyzeContractSource(t, service, []byte(source), sourceAdjectivePath, ContractAnalysisOptions{})
	projection := sourceProjectionNamed(t, initial, "Compact")
	declaration := sourceDeclarationFor(projection)
	checked := analyzeContractSource(t, service, []byte(source), sourceAdjectivePath, ContractAnalysisOptions{SourceTransformers: []contracts.SourceTransformerDeclaration{declaration}})
	if len(checked.SourceInstantiations) != 1 || checked.SourceInstantiations[0].Status != contracts.Contradicted {
		t.Fatalf("existing source child reused ordinal-zero certificate: %+v", checked.SourceInstantiations)
	}
}

func TestSourceAdjectiveDictionarySelectedAssuranceAPI(t *testing.T) {
	path := filepath.Join(t.TempDir(), "words.l")
	if err := os.WriteFile(path, []byte(sourceCompact), 0600); err != nil {
		t.Fatal(err)
	}
	service := NewLygerService(Options{})
	base, err := service.VerifyDictionaryContracts(path, ContractAnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(base.SourceProjections) != 1 {
		t.Fatal("dictionary projection included unrelated provider definitions")
	}
	projection := base.SourceProjections[0]
	declaration := sourceDeclarationFor(projection)
	options := DictionaryVerificationOptions{Analysis: ContractAnalysisOptions{SourceTransformers: []contracts.SourceTransformerDeclaration{declaration}}, Meanings: []contracts.MeaningKey{projection.Meaning}, Assurance: "symbolic"}
	checked, err := service.VerifyDictionaryContractsWithOptions(path, options)
	if err != nil || checked.Verification.Status != contracts.Established || len(checked.SourceCertificates[0].LoweringObligations) == 0 {
		t.Fatalf("selected symbolic API failed: %v", err)
	}
	options.Assurance = "implementation"
	implementation, err := service.VerifyDictionaryContractsWithOptions(path, options)
	if err != nil || implementation.Verification.Status == contracts.Established {
		t.Fatal("source topology certified executable dictionary implementation")
	}
	options.Assurance = "symbolic"
	options.Analysis.SourceTransformers[0].Provides = []contracts.Predicate{{Kind: "eq", Subject: "children.count", Target: projection.Target, Value: contracts.Atom{Kind: "number", Number: 0}}}
	rejected, err := service.VerifyDictionaryContractsWithOptions(path, options)
	if err != nil || rejected.Verification.Status != contracts.Contradicted {
		t.Fatalf("false structural promise did not fail selected API: %v", err)
	}
	options.Meanings = nil
	if _, err := service.VerifyDictionaryContractsWithOptions(path, options); !errors.Is(err, ErrSourceContractMetadata) {
		t.Fatal("missing symbolic selection accepted")
	}
	options.Meanings = []contracts.MeaningKey{"missing"}
	if _, err := service.VerifyDictionaryContractsWithOptions(path, options); !errors.Is(err, ErrSourceContractMetadata) {
		t.Fatal("absent symbolic selection accepted")
	}
}

func TestSourceAdjectiveNestedHelperKeepsIntermediateTarget(t *testing.T) {
	source := "\"Inner\" describes a Generic App with a New Main Service using any Process.\n\"Outer\" describes a Generic Inner App.\n\"Extra\" describes a Generic App with a New `extra` Service.\nCreate \"Worker\" as a New Extra Outer App.\n"
	service := NewLygerService(Options{})
	initial := analyzeContractSource(t, service, []byte(source), sourceAdjectivePath, ContractAnalysisOptions{})
	inner, outer := sourceProjectionNamed(t, initial, "Inner"), sourceProjectionNamed(t, initial, "Outer")
	options := ContractAnalysisOptions{SourceTransformers: []contracts.SourceTransformerDeclaration{sourceDeclarationFor(inner), sourceDeclarationFor(outer)}}
	checked := analyzeContractSource(t, service, []byte(source), sourceAdjectivePath, options)
	var temporary, receiver *contracts.SourceAdjectiveInstantiation
	for index := range checked.SourceInstantiations {
		instance := &checked.SourceInstantiations[index]
		if instance.Meaning == inner.Meaning {
			temporary = instance
		}
		if instance.Meaning == outer.Meaning {
			receiver = instance
		}
	}
	if temporary == nil || receiver == nil || temporary.Target.Root == receiver.Target.Root || temporary.Target.Generation != "source-construction-delta" {
		t.Fatalf("helper's fresh delta was rebound to final receiver: %+v", checked.SourceInstantiations)
	}
	if temporary.Status != contracts.Established || receiver.Status != contracts.Contradicted {
		t.Fatalf("fresh delta/final receiver premises were conflated: %+v", checked.SourceInstantiations)
	}
	if sourceCertificateFor(t, checked, outer.Meaning).LogicalStatus != contracts.Established {
		t.Fatal("concrete receiver incompatibility changed conditional outer theorem")
	}
}

func TestSourceAdjectiveReinferenceReplacesDeclarations(t *testing.T) {
	service := NewLygerService(Options{})
	analysis := analyzeContractSource(t, service, []byte(sourceCompact), sourceAdjectivePath, ContractAnalysisOptions{})
	projection := sourceProjectionNamed(t, analysis, "Compact")
	declaration := sourceDeclarationFor(projection)
	if err := analysis.ReinferContracts(ContractAnalysisOptions{SourceTransformers: []contracts.SourceTransformerDeclaration{declaration}}); err != nil {
		t.Fatal(err)
	}
	if sourceCertificateFor(t, analysis, declaration.Meaning).LogicalStatus != contracts.Established {
		t.Fatal("new declarations were ignored by reinference")
	}
	if err := analysis.ReinferContracts(ContractAnalysisOptions{}); err != nil {
		t.Fatal(err)
	}
	if sourceCertificateFor(t, analysis, declaration.Meaning).LogicalStatus != contracts.Established {
		t.Fatal("backend-only reinference removed source declaration")
	}
	if err := analysis.ReinferContracts(ContractAnalysisOptions{SourceTransformers: []contracts.SourceTransformerDeclaration{}}); err != nil {
		t.Fatal(err)
	}
	if sourceCertificateFor(t, analysis, declaration.Meaning).LogicalStatus == contracts.Established {
		t.Fatal("explicitly removed source declaration stayed certified")
	}
}
