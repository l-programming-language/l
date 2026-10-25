package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	distribution "github.com/l-programming-language/l"
	"github.com/l-programming-language/l/lerrors"
	Lyger "github.com/l-programming-language/l/lyger"
)

type buildOptions struct {
	OutputDir string
	GOOS      string
	GOARCH    string
}

type compilerInitOptions struct {
	Language  string
	AssetRoot string
}

type buildCommand struct {
	CLI      *cliData
	Assets   string
	Output   string
	Language string
}

var errUnsupportedBuildHost = lerrors.Unsupported.Errorf("native builds require macOS or Linux")

const buildUsage = `Usage: l build [--language LANGUAGE] [--assets ROOT] [--output-dir DIRECTORY] [main.l]
       l --version

--version prints the embedded major.minor.patch release version and exits.
Build native macOS or Linux executables beside the source, or in --output-dir.
The default source is main.l. Options must precede the source filename.
Language resources are embedded in the executable. The default language is eng.
Use --assets to select a complete development resource bundle instead.
Cross-compilation and the legacy -os flag are not supported.
`

// parseCommand supplies the command options and established CLI data projection.
// Its local FlagSet makes repeated parsing independent of process-global flags.
func parseCommand(args []string) (*buildCommand, error) {
	if len(args) == 0 {
		return nil, lerrors.InvalidInput.Errorf("a build command is required")
	}
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h" || args[0] == "-help") {
		return nil, flag.ErrHelp
	}
	if args[0] != "build" {
		return nil, lerrors.InvalidInput.Errorf("unknown command %q; expected build", args[0])
	}
	command := &buildCommand{}
	flags := flag.NewFlagSet("l build", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&command.Assets, "assets", "", "compiler asset root")
	flags.StringVar(&command.Output, "output-dir", "", "output directory")
	flags.StringVar(&command.Language, "language", Lyger.DefaultLanguage, "source language")
	flags.Func("os", "unsupported legacy target flag", func(string) error {
		return lerrors.InvalidInput.Errorf("-os is no longer supported; builds target the native host")
	})
	if err := flags.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil, err
		}
		return nil, lerrors.InvalidInput.Wrap(err, "")
	}
	var emptyOption string
	flags.Visit(func(f *flag.Flag) {
		if f.Value.String() == "" {
			emptyOption = f.Name
		}
	})
	if emptyOption != "" {
		return nil, lerrors.InvalidInput.Errorf("--%s requires a nonempty value", emptyOption)
	}
	positionals := flags.Args()
	if len(positionals) > 1 {
		return nil, lerrors.InvalidInput.Errorf("build accepts one source file; options must precede it")
	}
	source := "main.l"
	if len(positionals) == 1 {
		source = positionals[0]
		if source == "" {
			return nil, lerrors.InvalidInput.Errorf("source file must not be empty")
		}
	}
	// osType is an established parser compatibility field. Runtime targeting is
	// explicit in buildOptions and does not use this legacy distribution name.
	command.CLI = &cliData{jobType: "build", mainPath: source, osType: "ubuntu"}
	return command, nil
}

func nativeBuildOptions(output, goos, goarch string) (buildOptions, error) {
	if goos != "darwin" && goos != "linux" {
		return buildOptions{}, fmt.Errorf("%w; host is %s/%s", errUnsupportedBuildHost, goos, goarch)
	}
	return buildOptions{OutputDir: output, GOOS: goos, GOARCH: goarch}, nil
}

// resolveBuildCommand resolves caller paths without changing process cwd.
// Compiler assets and source-local imports deliberately have separate roots.
func resolveBuildCommand(command *buildCommand) (*cliData, compilerInitOptions, buildOptions, error) {
	if command == nil || command.CLI == nil {
		return nil, compilerInitOptions{}, buildOptions{}, lerrors.InvalidInput.Errorf("build command is required")
	}
	options, err := nativeBuildOptions("", runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return nil, compilerInitOptions{}, buildOptions{}, err
	}
	input := *command.CLI
	input.mainPath, err = filepath.Abs(input.mainPath)
	if err != nil {
		return nil, compilerInitOptions{}, buildOptions{}, lerrors.IO.Errorf("resolve source: %w", err)
	}
	initialization := compilerInitOptions{Language: command.Language}
	if command.Assets != "" {
		initialization.AssetRoot, err = filepath.Abs(command.Assets)
		if err != nil {
			return nil, compilerInitOptions{}, buildOptions{}, lerrors.IO.Errorf("resolve assets: %w", err)
		}
		info, err := os.Stat(initialization.AssetRoot)
		if err != nil {
			return nil, compilerInitOptions{}, buildOptions{}, lerrors.IO.Errorf("read compiler assets: %w", err)
		}
		if !info.IsDir() {
			return nil, compilerInitOptions{}, buildOptions{}, lerrors.InvalidInput.Errorf("compiler asset root is not a directory: %s", initialization.AssetRoot)
		}
	}
	options.OutputDir = command.Output
	if options.OutputDir == "" {
		options.OutputDir = filepath.Dir(input.mainPath)
	}
	options.OutputDir, err = filepath.Abs(options.OutputDir)
	if err != nil {
		return nil, compilerInitOptions{}, buildOptions{}, lerrors.IO.Errorf("resolve output directory: %w", err)
	}
	return &input, initialization, options, nil
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 1 && args[0] == "--version" {
		if _, err := fmt.Fprintln(stdout, distribution.Version()); err != nil {
			return 1
		}
		return 0
	}
	command, err := parseCommand(args)
	if errors.Is(err, flag.ErrHelp) {
		if _, err := fmt.Fprint(stdout, buildUsage); err != nil {
			return 1
		}
		return 0
	}
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n%s", err, buildUsage) //nolint:errcheck // Preserve the command failure status if its diagnostic stream is unavailable.
		return 2
	}
	input, initialization, options, err := resolveBuildCommand(command)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err) //nolint:errcheck // Preserve the command failure status if its diagnostic stream is unavailable.
		if errors.Is(err, errUnsupportedBuildHost) {
			return 2
		}
		return 1
	}
	c := newCompiler(input)
	c.initialize(initialization)
	if c.err != nil {
		fmt.Fprintf(stderr, "error: %v\n", c.err) //nolint:errcheck // Preserve the compiler failure when reporting it also fails.
		return 1
	}
	path, err := buildExecutable(ctx, c, options)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err) //nolint:errcheck // Preserve the native build failure when reporting it also fails.
		return 1
	}
	if _, err := fmt.Fprintln(stdout, path); err != nil {
		return 1
	}
	return 0
}
