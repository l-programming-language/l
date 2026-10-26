package main

import (
	"fmt"
	"os"

	distribution "github.com/l-programming-language/l"
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
	if c.err != nil {
		return nil
	}
	if c.lyger == nil || c.cli == nil {
		c.err = lerrors.InvalidInput.Errorf("compiler is not initialized")
		return nil
	}
	result, err := c.lyger.Compile(c.mainBytes, c.cli.mainPath)
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
	if c.log == nil {
		c.err = lerrors.InvalidInput.Errorf("source generator is not initialized")
		return nil
	}
	plan, err := result.WorkspacePlan()
	if err != nil {
		c.err = fmt.Errorf("prepare applications: %w", err)
		return nil
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
		c.err = fmt.Errorf("prepare workspace: %w", err)
		return nil
	}
	var bundle Logiceme.Executable
	var libraryBundle Logiceme.Executable
	var applicationIDs []string
	var libraryIDs []string
	for _, unit := range units {
		generated, err := c.log.Evaluate(unit.BackendAML, Logiceme.GenerationOptions{IncludeRuntime: false})
		if err != nil {
			c.err = fmt.Errorf("generate source for application %q: %w", unit.ID, err)
			return nil
		}
		if len(generated.ByteArrays) != 1 || len(generated.MapArrays) != 1 ||
			len(generated.DependencyArrays) != 1 || len(generated.EnvDepArrays) != 1 ||
			len(generated.ByteArrays[0]) == 0 || len(generated.ByteArrays[0]) != len(generated.MapArrays[0]) {
			c.err = lerrors.Internal.Errorf("generate source for application %q: expected one complete source bundle", unit.ID)
			return nil
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
	return &generatedProgram{Compilation: result, Sources: bundle, ApplicationIDs: applicationIDs, LibrarySources: libraryBundle, LibraryIDs: libraryIDs}
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
