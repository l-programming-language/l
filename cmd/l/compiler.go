package main

import (
	"fmt"
	"os"
	"path"

	distribution "github.com/l-programming-language/l"
	"github.com/l-programming-language/l/contracts"
	"github.com/l-programming-language/l/lerrors"
	Logiceme "github.com/l-programming-language/l/logiceme"
	Lyger "github.com/l-programming-language/l/lyger"
	"github.com/l-programming-language/l/utilities"
)

// cliData stores the read input data from the CLI
type cliData struct {
	jobType  string
	mainPath string
	osType   string
}

// compiler coordinates source analysis and generation. Native builds consume
// its returned source bundles without changing the tested lifecycle state.
type compiler struct {
	cli           *cliData
	options       Lyger.Options
	lyger         *Lyger.LygerService
	log           *Logiceme.LogicemeService
	compilation   *Lyger.Compilation
	mainBytes     []byte
	compiledBytes []byte
	asmBytes      []byte
	err           error
}

// newCompiler instantiates and returns a new compiler structure
func newCompiler(input *cliData) *compiler {
	var c compiler
	c.cli = input
	c.err = nil
	return &c
}

// initialize the compiler structure based on the user inputs
func (c *compiler) initialize(options ...compilerInitOptions) {
	if c.cli == nil {
		c.err = lerrors.InvalidInput.Errorf("compiler input is required")
		return
	}
	// Check main file from CLI arguments
	valid, err := utilities.CheckIfValidFile(c.cli.mainPath)
	if err != nil {
		c.err = fmt.Errorf("initialize: %w", err)
		return
	}
	if !valid {
		c.err = lerrors.InvalidInput.Errorf("main file is not l")
		return
	}
	// Select one resource provider for both compiler stages.
	var settings compilerInitOptions
	if len(options) > 1 {
		c.err = lerrors.InvalidInput.Errorf("initialize: at most one compiler options value is allowed")
		return
	}
	if len(options) == 1 {
		settings = options[0]
	}
	c.options = Lyger.Options{Language: settings.Language, Resources: distribution.Files(), AssetRoot: settings.AssetRoot}
	if c.options.Language == "" {
		c.options.Language = Lyger.DefaultLanguage
	}
	if settings.AssetRoot != "" {
		c.options.Resources = os.DirFS(settings.AssetRoot)
	}
	c.lyger = Lyger.NewLygerService(c.options)
	c.log = Logiceme.NewLogicemeService(c.options.Resources)
	// Working sessions are owned by the returned compilation result, while this
	// field retains the CLI's established initialized lifecycle state.
	c.compilation = &Lyger.Compilation{}
	// Load bytes of main file from c.cli.mainPath
	mainBytes, err := utilities.ReadFile(c.cli.mainPath)
	if err != nil {
		c.err = fmt.Errorf("read main source: %w", err)
		return
	}
	c.mainBytes = mainBytes
}

// compile - (main.l -> ASM via lyger)
func (c *compiler) compile() *Lyger.CompilationResult {
	return c.compileSource(false)
}

func (c *compiler) compileSource(record bool) *Lyger.CompilationResult {
	if c.err != nil {
		return nil
	}
	if c.lyger == nil || c.cli == nil {
		c.err = lerrors.InvalidInput.Errorf("compiler is not initialized")
		return nil
	}
	resolve := c.lyger.Compile
	if record {
		resolve = c.lyger.CompileForContracts
	}
	result, err := resolve(c.mainBytes, c.cli.mainPath)
	if err != nil {
		c.err = fmt.Errorf("compile: %w", err)
		return nil
	}
	c.compiledBytes = append([]byte(nil), result.ExpandedSource...)
	c.asmBytes = append([]byte(nil), result.AML...)
	return result
}

// process generates one source bundle per canonical application, in source order.
func (c *compiler) process() *generatedProgram {
	result := c.compile()
	if result == nil {
		return nil
	}
	generated, err := generatePackages(c.log, result, nil)
	c.err = err
	return generated
}

// generatePackages renders each App and constructed Service once. With an
// inventory, the bytes and implementation fingerprints share a prepared graph.
func generatePackages(backend *Logiceme.LogicemeService, result *Lyger.CompilationResult, inventory *contracts.LoweringInventory) (*generatedProgram, error) {
	if backend == nil {
		return nil, lerrors.InvalidInput.Errorf("source generator is not initialized")
	}
	plan, err := result.WorkspacePlan()
	if err != nil {
		return nil, fmt.Errorf("prepare applications: %w", err)
	}
	locations := make(map[string]string)
	apps, libraries := 0, 0
	for _, item := range plan.Packages() {
		if item.Library {
			locations[item.ID] = fmt.Sprintf("services/service-%03d", libraries)
			libraries++
		} else {
			locations[item.ID] = fmt.Sprintf("apps/app-%03d", apps)
			apps++
		}
	}
	units, err := plan.Units(locations)
	if err != nil {
		return nil, fmt.Errorf("prepare workspace: %w", err)
	}
	var bundle Logiceme.Executable
	var libraryBundle Logiceme.Executable
	var applicationIDs []string
	var libraryIDs []string
	for _, unit := range units {
		var generated Logiceme.Executable
		if inventory == nil {
			generated, err = backend.Evaluate(unit.BackendAML)
		} else {
			var prepared *Logiceme.ContractGeneration
			if prepared, err = backend.PrepareContractGeneration(unit.BackendAML, Logiceme.GenerationOptions{}); err == nil {
				generated, err = prepared.Render()
				mergeLowering(inventory, "unit:"+unit.ID, prepared.Inventory())
			}
		}
		if err != nil {
			return nil, fmt.Errorf("generate source for application %q: %w", unit.ID, err)
		}
		if len(generated.ByteArrays) != 1 || len(generated.MapArrays) != 1 ||
			len(generated.DependencyArrays) != 1 || len(generated.EnvDepArrays) != 1 ||
			len(generated.ByteArrays[0]) == 0 || len(generated.ByteArrays[0]) != len(generated.MapArrays[0]) {
			return nil, lerrors.Internal.Errorf("generate source for application %q: expected one complete source bundle", unit.ID)
		}
		destination := &bundle
		if unit.Library {
			libraryIDs = append(libraryIDs, unit.ID)
			destination = &libraryBundle
		} else {
			applicationIDs = append(applicationIDs, unit.ID)
		}
		if len(destination.ByteArrays) == 0 {
			*destination = generated
		} else {
			destination.ByteArrays = append(destination.ByteArrays, generated.ByteArrays...)
			destination.MapArrays = append(destination.MapArrays, generated.MapArrays...)
			destination.DependencyArrays = append(destination.DependencyArrays, generated.DependencyArrays...)
			destination.EnvDepArrays = append(destination.EnvDepArrays, generated.EnvDepArrays...)
		}
	}
	return &generatedProgram{Compilation: result, Sources: bundle, ApplicationIDs: applicationIDs, LibrarySources: libraryBundle, LibraryIDs: libraryIDs}, nil
}

// generatedProgram owns both the frontend explanation and generated source.
// Source persistence and execution are separate from the source generator.
type generatedProgram struct {
	Compilation    *Lyger.CompilationResult
	Sources        Logiceme.Executable
	ApplicationIDs []string
	LibrarySources Logiceme.Executable
	LibraryIDs     []string
}

func mergeLowering(out *contracts.LoweringInventory, namespace string, in contracts.LoweringInventory) {
	if out.Fingerprints == nil {
		out.Fingerprints = make(map[string]string)
	}
	for _, step := range in.Steps {
		step.ID = contracts.StableID("inventory-use", namespace+":"+step.ID)
		out.Steps = append(out.Steps, step)
	}
	out.Native = append(out.Native, in.Native...)
	out.Unknowns = append(out.Unknowns, in.Unknowns...)
	for key, value := range in.Fingerprints {
		out.Fingerprints[namespace+":"+key] = value
	}
}

func plannedBundles(sources Logiceme.Executable) ([]*sourceBundle, error) {
	var result []*sourceBundle
	for i, files := range sources.ByteArrays {
		if len(sources.EnvDepArrays[i]) != 0 {
			return nil, lerrors.Unsupported.Errorf("package %d requires unsupported environment dependencies", i)
		}
		bundle, err := planSourceBundle(files, sources.MapArrays[i])
		if err != nil {
			return nil, err
		}
		result = append(result, bundle)
	}
	return result, nil
}

type nativeWorkspace struct {
	apps, libraries []*sourceBundle
	entry, support  *sourceBundle
}

// workspace completes the real Orchestrator and runtime with the selected
// generation's companion paths. Inspection calls the same in-memory lowering.
func (p *generatedProgram) workspace(backend *Logiceme.LogicemeService, stem, generation string, inventory *contracts.LoweringInventory) (w nativeWorkspace, err error) {
	if w.apps, err = plannedBundles(p.Sources); err != nil {
		return w, err
	}
	if w.libraries, err = plannedBundles(p.LibrarySources); err != nil {
		return w, err
	}
	names := append(append([]*sourceBundle(nil), w.apps...), w.libraries...)
	entryName := orchestratorPackageName(stem, names)
	var companions []Lyger.AppArtifact
	for i, id := range p.ApplicationIDs {
		companions = append(companions, Lyger.AppArtifact{ID: id, RelativeExecutable: path.Join(stem+".apps", "generations", generation, fmt.Sprintf("app-%03d", i))})
	}
	unit, err := p.Compilation.OrchestratorUnit(entryName, companions)
	if err != nil {
		return w, fmt.Errorf("lower orchestrator: %w", err)
	}
	runtime := Logiceme.WorkspaceRuntime{Package: runtimePackageName(entryName, names), RelativeDirectory: "../support/runtime"}
	imports, err := backend.ReadImports(unit.BackendAML)
	if err != nil {
		return w, fmt.Errorf("read orchestrator imports: %w", err)
	}
	library, err := p.Compilation.RuntimeLibrary(runtime.Package, imports)
	if err != nil {
		return w, fmt.Errorf("select runtime library: %w", err)
	}
	var libraries [][]byte
	if library != nil {
		libraries = append(libraries, library)
	}
	var entry, support Logiceme.Executable
	if inventory == nil {
		entry, support, err = backend.EvaluateWorkspaceEntry(unit.BackendAML, runtime, libraries...)
	} else {
		var prepared *Logiceme.ContractGeneration
		if prepared, err = backend.PrepareWorkspaceContractGeneration(unit.BackendAML, runtime, libraries...); err == nil {
			entry, support, err = prepared.RenderWorkspace()
			mergeLowering(inventory, "orchestrator", prepared.Inventory())
		}
	}
	if err != nil {
		return w, fmt.Errorf("generate orchestrator: %w", err)
	}
	if w.entry, err = workspaceBundle(entry); err != nil {
		return w, err
	}
	w.support, err = workspaceBundle(support)
	return w, err
}

func workspaceBundle(part Logiceme.Executable) (*sourceBundle, error) {
	if len(part.ByteArrays) != 1 || len(part.MapArrays) != 1 || len(part.EnvDepArrays) != 1 || len(part.EnvDepArrays[0]) != 0 {
		return nil, lerrors.Internal.Errorf("invalid workspace entry/runtime source dimensions")
	}
	return planSourceBundle(part.ByteArrays[0], part.MapArrays[0])
}
