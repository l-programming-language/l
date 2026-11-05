package Logiceme

import (
	"fmt"
	"sort"
	"strings"

	"github.com/l-programming-language/l/contracts"
	"github.com/l-programming-language/l/lerrors"
)

// ContractGeneration owns one prepared backend graph. Inventory and rendering
// share the actual target/header/support selection, without examining emitted
// Rust to infer source semantics.
type ContractGeneration struct {
	program        *Program
	templates      map[string][]byte
	includeRuntime bool
	inventory      contracts.LoweringInventory
	rendered       bool
	workspace      *WorkspaceRuntime
}

func (ls *LogicemeService) PrepareContractGeneration(input []byte, options GenerationOptions) (*ContractGeneration, error) {
	return ls.prepareGeneration(input, options, nil, true)
}

// PrepareWorkspaceContractGeneration prepares the real workspace entry and its
// selected runtime library, including the inventory used for contract checking.
func (ls *LogicemeService) PrepareWorkspaceContractGeneration(input []byte, runtime WorkspaceRuntime, libraries ...[]byte) (*ContractGeneration, error) {
	return ls.prepareGeneration(input, GenerationOptions{IncludeRuntime: true}, &runtime, true, libraries...)
}

func (ls *LogicemeService) prepareGeneration(input []byte, options GenerationOptions, workspace *WorkspaceRuntime, record bool, libraries ...[]byte) (*ContractGeneration, error) {
	if ls == nil || ls.resources == nil {
		return nil, lerrors.InvalidConfiguration.Errorf("source generator configuration is required")
	}
	if workspace != nil {
		if err := validateWorkspaceRuntime(*workspace); err != nil {
			return nil, err
		}
		if len(libraries) > 1 {
			return nil, lerrors.InvalidInput.Errorf("workspace entry accepts one runtime library")
		}
	}
	system, err := loadSegmentalSystem(ls.resources)
	if err != nil {
		return nil, fmt.Errorf("load source generator: %w", err)
	}
	var templates map[string][]byte
	if workspace == nil {
		templates, err = loadSupportTemplates(ls.resources, options.IncludeRuntime)
	} else if templates, err = loadSupportAssets(ls.resources, true); err == nil {
		if len(libraries) == 0 {
			templates["runtime/runtime.rs"], err = loadRuntimeComposition(ls.resources, templates)
		} else {
			templates["runtime/runtime.rs"], err = generateRuntimeComposition(libraries[0], system, templates, workspace.Package)
		}
	}
	if err != nil {
		return nil, err
	}
	return prepareContractGeneration(input, system, templates, options, workspace, record)
}

func prepareContractGeneration(input []byte, system SegmentalSystem, templates map[string][]byte, options GenerationOptions, workspace *WorkspaceRuntime, record bool) (*ContractGeneration, error) {
	scanner, err := loadScannerBytes(input)
	if err != nil {
		return nil, err
	}
	if err := scanSourceProgram(&scanner); err != nil {
		return nil, err
	}
	program, err := prepareScannedSource(scanner, system, templates, options.IncludeRuntime, workspace)
	if err != nil {
		return nil, err
	}
	prepared := &ContractGeneration{program: program, templates: templates, includeRuntime: options.IncludeRuntime, workspace: workspace}
	if !record {
		return prepared, nil
	}
	prepared.inventory = contracts.LoweringInventory{Fingerprints: make(map[string]string), BoundaryCoverage: contracts.Known, TransitiveCoverage: contracts.Unknown}
	prepared.inventory.Fingerprints["runtime-placement"] = contracts.StableID("runtime-placement", workspace)
	// Segment templates are executable composition dependencies, even when a
	// native leaf and its AML have not changed. Hash the prepared system rather
	// than treating the support-file list as a complete implementation closure.
	prepared.inventory.Fingerprints["segmental-system"] = contracts.StableID("segmental-system", system)
	add := func(name, kind string, target contracts.TargetRef, data any, phase contracts.Phase) {
		meaning := contracts.MeaningKey(contracts.StableID("backend-meaning", name))
		digest := contracts.StableID("backend-input", data)
		step := contracts.LoweringStep{ID: contracts.StableID("backend-use", struct {
			Name   string
			Target contracts.TargetRef
		}{name, target}), Meaning: meaning, Name: name, Kind: kind, Target: target, Origin: contracts.Origin{Source: "backend:" + name, Label: "selected typed backend support"}, Implementation: digest}
		prepared.inventory.Steps = append(prepared.inventory.Steps, step)
		prepared.inventory.Fingerprints[string(meaning)] = digest
		if kind != "artifact-only" {
			prepared.inventory.Native = append(prepared.inventory.Native, contracts.NativeBoundary{ID: string(meaning), Meaning: meaning, Target: target, Origin: step.Origin, Status: contracts.Unresolved, Reason: "backend native implementation has no approved conformance witness", Implicit: phase == contracts.Cleanup || phase == contracts.Construction})
		}
	}
	for i := range program.Compiler.Apps {
		app := &program.Compiler.Apps[i]
		target := contracts.TargetRef{Root: app.Name, Type: "App", Phase: contracts.Invocation}
		if err := app.eachObject(func(object *Object) error {
			objectTarget := target
			objectTarget.Path = []string{object.Location + object.Name}
			add(app.Name+"/"+object.Location+object.Name, "authored-native-composition", objectTarget, input, contracts.Invocation)
			if object.header.initializeConfig {
				add(app.Name+"/configuration-startup", "implicit-configuration", objectTarget, templates["config/config.lu"], contracts.Construction)
			}
			return nil
		}); err != nil {
			return nil, err
		}
		if options.IncludeRuntime && i == 0 {
			add(app.Name+"/runtime-finalization", "implicit-runtime-lifecycle", target, templates["runtime/main.lu"], contracts.Cleanup)
		}
	}
	var names []string
	for name := range templates {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		kind := "artifact-only"
		phase := contracts.Elaboration
		if strings.HasPrefix(name, "runtime/") {
			kind = "runtime-support"
			phase = contracts.Invocation
		}
		add(name, kind, contracts.TargetRef{Root: "backend-support", Phase: phase}, templates[name], phase)
	}
	prepared.inventory.Unknowns = []contracts.UnknownReason{{Code: "opaque_backend_native", Message: "selected source/native and runtime support boundaries are enumerated; internal foreign calls and representation destruction require explicit contracts", Components: []string{"effects", "resources", "cleanup", "failure"}}}
	if err := prepared.inventory.Validate(); err != nil {
		return nil, err
	}
	return prepared, nil
}

func (g *ContractGeneration) Inventory() contracts.LoweringInventory {
	return contracts.CloneInventory(g.inventory)
}

func (g *ContractGeneration) Render() (Executable, error) {
	if g == nil || g.program == nil || g.rendered {
		return Executable{}, fmt.Errorf("contract generation requires an unconsumed prepared graph")
	}
	before := g.Inventory()
	g.rendered = true
	if err := g.program.Assemble(); err != nil {
		return Executable{}, err
	}
	g.program.Link()
	if err := g.program.linkBundleSupport(g.templates, g.includeRuntime, g.workspace); err != nil {
		return Executable{}, err
	}
	if before.Fingerprints != nil {
		if err := contracts.VerifyInventory(before, g.inventory); err != nil {
			return Executable{}, err
		}
	}
	return g.program.Executable, nil
}

func (g *ContractGeneration) RenderWorkspace() (Executable, Executable, error) {
	if g == nil || g.workspace == nil {
		return Executable{}, Executable{}, lerrors.InvalidInput.Errorf("expected a prepared workspace entry")
	}
	entry, err := g.Render()
	if err != nil {
		return Executable{}, Executable{}, err
	}
	support, err := workspaceRuntimeLibrary(g.workspace.Package, g.templates)
	if err != nil {
		return Executable{}, Executable{}, err
	}
	return entry, support, nil
}
