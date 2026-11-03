package Lyger

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/l-programming-language/l/contracts"
)

const contractsSource = "Use IO's Print from `io` as \"Print\".\nUse Basic as \"Compact\".\nCreate \"First\" as a New Basic App.\nThe App Prints `one`.\nCreate \"Second\" as a New Compact App.\nThe App Prints `two`.\nRun First.\nAwait First.\nShow First's Output.\n"

func TestContractsTrackA(t *testing.T) {
	service := NewLygerService(Options{})
	ordinary, err := service.Compile([]byte(contractsSource), "/project/main.l")
	if err != nil {
		t.Fatal(err)
	}
	comp, observed, err := service.ResolveContractBindings([]byte(contractsSource), "/project/main.l")
	if err != nil {
		t.Fatal(err)
	}
	t.Run("BASE-009", func(t *testing.T) {
		if !bytes.Equal(ordinary.AML, comp.ASMBytes) || !bytes.Equal(ordinary.BackendAML, comp.Program.BackendBytes) || !bytes.Equal(ordinary.ExpandedSource, comp.ExpandedBytes) {
			t.Fatal("observing bindings changed source compilation")
		}
	})
	t.Run("BASE-002", func(t *testing.T) {
		var basic, alias *contracts.ResolvedBinding
		for i := range observed.Bindings {
			binding := &observed.Bindings[i]
			if binding.CanonicalName == "Basic" && binding.Target.Type == "app" {
				if binding.Spelling == "Basic" {
					basic = binding
				}
				if binding.Spelling == "Compact" {
					alias = binding
				}
			}
		}
		if basic == nil || alias == nil {
			t.Fatal("missing descriptor occurrences")
		}
		if basic.Meaning != alias.Meaning || basic.Definition != alias.Definition {
			t.Fatal("aliases did not share canonical meaning")
		}
		if basic.Occurrence == alias.Occurrence || basic.Target.Root == alias.Target.Root {
			t.Fatal("alias applications collapsed occurrences or target roots")
		}
	})
	t.Run("LEX-007", func(t *testing.T) {
		found := false
		for _, binding := range observed.Bindings {
			if binding.CanonicalName == "IO.Print" && strings.Contains(binding.Origin.Source, "main.l") && binding.Spelling == "Prints" {
				found = true
				if binding.DefinitionEnvironment == binding.UseEnvironment {
					t.Fatal("dictionary definition and caller environments collapsed")
				}
				if !strings.HasPrefix(binding.DefinitionOrigin.Source, "resource:") {
					t.Fatal("dictionary lost provider origin")
				}
			}
		}
		if !found {
			t.Fatal("missing imported IO.Print binding")
		}
	})
	t.Run("BASE-008", func(t *testing.T) {
		_, again, err := service.ResolveContractBindings([]byte(contractsSource), "/project/main.l")
		if err != nil {
			t.Fatal(err)
		}
		first, err := json.Marshal(observed)
		if err != nil {
			t.Fatal(err)
		}
		second, err := json.Marshal(again)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(first, second) {
			t.Fatal("binding report changes between compilations")
		}
	})
}

func TestContractsConfiguredNativeBindings(t *testing.T) {
	source := "Create \"Worker\" as a New Basic App.\nStart Worker.\nAwait Worker.\nStart Worker.\n"
	_, snapshot, err := NewLygerService(Options{}).ResolveContractBindings([]byte(source), "/project/native.l")
	if err != nil {
		t.Fatal(err)
	}
	var runs []contracts.ResolvedBinding
	for _, binding := range snapshot.Bindings {
		if binding.CanonicalName == "native:action.executive" && binding.Origin.Label == "expanded typed instruction" {
			runs = append(runs, binding)
		}
	}
	if len(runs) != 2 {
		t.Fatalf("expanded native Run uses=%d, want 2", len(runs))
	}
	if runs[0].Definition != runs[1].Definition || runs[0].Meaning != runs[1].Meaning || runs[0].Occurrence == runs[1].Occurrence || runs[0].Origin.Parent == runs[1].Origin.Parent {
		t.Fatal("native canonical/occurrence/callsite identities conflated")
	}
	for _, run := range runs {
		if run.Target.Unknown || run.Target.Root == "" || run.Target.Phase != contracts.Invocation {
			t.Fatalf("native typed receiver missing: %+v", run)
		}
		parent := false
		for _, binding := range snapshot.Bindings {
			parent = parent || binding.Occurrence == run.Origin.Parent
		}
		if !parent {
			t.Fatal("expanded native callsite provenance is dangling")
		}
		if run.DefinitionEnvironment == run.UseEnvironment {
			t.Fatal("native definition and caller environments conflated")
		}
	}
}

func TestContractsRepeatedSourceOrigins(t *testing.T) {
	directory := t.TempDir()
	included := filepath.Join(directory, "included.l")
	source := "\n    Use `included.l`.\n\n    Use `included.l`.\n"
	if err := os.WriteFile(included, []byte("\"value\" is `snow 雪`.\nShow value.\n"), 0600); err != nil {
		t.Fatal(err)
	}
	service := NewLygerService(Options{})
	_, first, err := service.ResolveContractBindings([]byte(source), filepath.Join(directory, "main.l"))
	if err != nil {
		t.Fatal(err)
	}
	parents := make(map[contracts.BindingOccurrenceID]bool)
	uses := make(map[contracts.BindingOccurrenceID]bool)
	for _, binding := range first.Bindings {
		if binding.Origin.Source == "source:included.l" && binding.Spelling == "value" {
			parents[binding.Origin.Parent] = true
			uses[binding.Occurrence] = true
		}
	}
	if len(parents) != 2 || len(uses) != 2 {
		t.Fatalf("repeated inclusion lost occurrence ancestry: parents=%d uses=%d", len(parents), len(uses))
	}
	_, second, err := service.ResolveContractBindings([]byte(source), filepath.Join(directory, "main.l"))
	if err != nil {
		t.Fatal(err)
	}
	left, err := json.Marshal(first)
	if err != nil {
		t.Fatal(err)
	}
	right, err := json.Marshal(second)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(left, right) {
		t.Fatal("repeated include identities nondeterministic")
	}
}

func TestContractsDefinitionSiteHelperShadow(t *testing.T) {
	files := dictionaryIndexFilesystem(t)
	files["stdlib/eng/io/dictionary.l"] = &fstest.MapFile{Data: []byte("Create \"IO\" as a Dictionary.\nDefine \"Secret\" as `dictionary owned`.\nCreate IO's \"Reveal|-s|-ed|-ing|-er\" for Showing Secret and then Showing any Text.\n")}
	source := "Use IO's Reveal from `io` as \"Reveal\".\nDefine \"Secret\" as `caller shadow`.\nReveal `argument`.\n"
	_, snapshot, err := NewLygerService(Options{Resources: files}).ResolveContractBindings([]byte(source), "/project/hygiene.l")
	if err != nil {
		t.Fatal(err)
	}
	var caller contracts.DefinitionKey
	var generated []contracts.ResolvedBinding
	for _, binding := range snapshot.Bindings {
		if binding.CanonicalName != "Secret" {
			continue
		}
		if binding.DefinitionOrigin.Source == "source:hygiene.l" {
			caller = binding.Definition
		}
		if binding.Origin.Label == "expanded typed operand" {
			generated = append(generated, binding)
		}
	}
	// Native Show carries the argument binding in its definition-site token.
	owned := false
	for _, binding := range snapshot.Bindings {
		if binding.CanonicalName == "Secret" && binding.Origin.Source == "resource:stdlib/eng/io/dictionary.l" {
			owned = true
			if binding.Definition == caller || !strings.HasPrefix(binding.DefinitionOrigin.Source, "resource:") {
				t.Fatal("caller captured dictionary helper")
			}
		}
	}
	if caller == "" || !owned || len(generated) == 0 {
		t.Fatal("fixture did not expose both private and caller helpers")
	}
	for _, binding := range generated {
		if binding.Definition == caller || binding.DefinitionEnvironment == binding.UseEnvironment {
			t.Fatal("expanded body captured caller helper")
		}
	}
}

func TestContractsDictionaryDiscoveryDoesNotCertify(t *testing.T) {
	service := NewLygerService(Options{})
	result, err := service.VerifyDictionaryContracts("io", ContractAnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Definitions) == 0 || len(result.Obligations) == 0 || contracts.AllEstablished(result.Obligations) {
		t.Fatal("loading a dictionary became conformance proof")
	}
	for _, definition := range result.Definitions {
		if definition.Origin.Source != "resource:stdlib/eng/io/dictionary.l" {
			t.Fatal("dictionary verification selected an unrelated declaration")
		}
	}
	directory := t.TempDir()
	path := filepath.Join(directory, "adjective.l")
	if err := os.WriteFile(path, []byte("\"Compact\" describes a Generic App with a New Main Service using any Process.\n"), 0600); err != nil {
		t.Fatal(err)
	}
	declared, err := service.VerifyDictionaryContracts(path, ContractAnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(declared.Certificates) != 1 || declared.Certificates[0].Mode != "instantiation_only" || declared.Certificates[0].Status != contracts.Unresolved {
		t.Fatalf("uninstantiated source adjective disappeared or self-certified: %+v", declared.Certificates)
	}
}

func TestContractsOriginalTokenSpans(t *testing.T) {
	source := "***prefix*** Use Basic as \"Compact\".\nCreate \"Worker\" as a New Compact App.\n"
	_, bindings, err := NewLygerService(Options{}).ResolveContractBindings([]byte(source), "/project/spans.l")
	if err != nil {
		t.Fatal(err)
	}
	for _, binding := range bindings.Bindings {
		if binding.Origin.Source != "source:spans.l" || binding.Origin.Label != "" {
			continue
		}
		start, end := binding.Origin.Start, binding.Origin.End
		if start < 0 || end > len(source) || start > end {
			t.Fatalf("invalid original span %+v", binding.Origin)
		}
		if string(source[start:end]) != binding.Spelling {
			t.Fatalf("span %q differs from token %q", source[start:end], binding.Spelling)
		}
	}
}

func TestContractsAnalysisIncludesLateHelpers(t *testing.T) {
	analysis, err := NewLygerService(Options{}).AnalyzeContracts([]byte(contractsSource), "/project/main.l", ContractAnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(analysis.Lowering.Steps) == 0 || len(analysis.Lowering.Native) == 0 {
		t.Fatal("helper composition missing from inventory")
	}
	if len(analysis.Orchestrator.BackendAML) == 0 {
		t.Fatal("source-only orchestration preparation missing")
	}
	for _, name := range []string{"NoHeapAllocation", "NoBlocking"} {
		obligations := contracts.CheckProfile(analysis.Summary, contracts.Profile{Name: name, Scope: "whole_region"})
		if contracts.AllEstablished(obligations) {
			t.Fatal("unapproved native closure established " + name)
		}
	}
}
