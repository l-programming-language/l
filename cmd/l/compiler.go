package main

import (
	"fmt"
	"os"

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

// process generates the single standalone App through its selected definitions.
func (c *compiler) process() *generatedProgram {
	result := c.compile()
	if result == nil {
		return nil
	}
	generated, err := generatePackages(c.log, result, nil)
	c.err = err
	return generated
}

// generatePackages renders the standalone App. With an inventory, source bytes
// and implementation fingerprints share the same prepared backend graph.
func generatePackages(backend *Logiceme.LogicemeService, result *Lyger.CompilationResult, inventory *contracts.LoweringInventory) (*generatedProgram, error) {
	if backend == nil {
		return nil, lerrors.InvalidInput.Errorf("source generator is not initialized")
	}
	if err := result.ValidateStandaloneApp(); err != nil {
		return nil, err
	}
	plan, err := result.WorkspacePlan()
	if err != nil {
		return nil, fmt.Errorf("prepare applications: %w", err)
	}
	packages := plan.Packages()
	if len(packages) != 1 || packages[0].Library {
		return nil, lerrors.Unsupported.Errorf("standalone compilation does not support additional Service libraries")
	}
	units, err := plan.Units(map[string]string{packages[0].ID: "apps/app-000"})
	if err != nil {
		return nil, fmt.Errorf("prepare workspace: %w", err)
	}
	if len(units) != 1 || units[0].Library {
		return nil, lerrors.Internal.Errorf("standalone plan must produce one App unit")
	}
	unit := units[0]
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
	return &generatedProgram{Compilation: result, Sources: generated, ApplicationIDs: []string{unit.ID}}, nil
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

type nativeWorkspace struct {
	entry *sourceBundle
}

// workspace publishes the existing composed App as the native entry. Runtime
// supervision and its additional package are outside the standalone boundary.
func (p *generatedProgram) workspace(_ *Logiceme.LogicemeService, _, _ string, _ *contracts.LoweringInventory) (w nativeWorkspace, err error) {
	if p == nil || p.Compilation == nil {
		return w, lerrors.InvalidInput.Errorf("standalone workspace requires a completed compilation")
	}
	if err := p.Compilation.ValidateStandaloneApp(); err != nil {
		return w, err
	}
	if len(p.ApplicationIDs) != 1 || len(p.LibraryIDs) != 0 || len(p.LibrarySources.ByteArrays) != 0 {
		return w, lerrors.Unsupported.Errorf("standalone workspace requires exactly one App and no additional libraries")
	}
	w.entry, err = workspaceBundle(p.Sources)
	return w, err
}

func workspaceBundle(part Logiceme.Executable) (*sourceBundle, error) {
	if len(part.ByteArrays) != 1 || len(part.MapArrays) != 1 || len(part.EnvDepArrays) != 1 || len(part.EnvDepArrays[0]) != 0 {
		return nil, lerrors.Internal.Errorf("invalid standalone workspace source dimensions")
	}
	return planSourceBundle(part.ByteArrays[0], part.MapArrays[0])
}
