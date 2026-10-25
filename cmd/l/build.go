package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/l-programming-language/l/internal/timing"
	"github.com/l-programming-language/l/lerrors"
	Logiceme "github.com/l-programming-language/l/logiceme"
	Lyger "github.com/l-programming-language/l/lyger"
)

const companionOwner = "l-executable-v1\n"

type builtApp struct {
	ID         string `json:"id"`
	Executable string `json:"executable"`
	SHA256     string `json:"sha256"`
}

type generationManifest struct {
	Version    int        `json:"version"`
	Entrypoint string     `json:"entrypoint"`
	GOOS       string     `json:"goos"`
	GOARCH     string     `json:"goarch"`
	Apps       []builtApp `json:"apps"`
}

func buildExecutable(ctx context.Context, c *compiler, options buildOptions) (builtPath string, resultErr error) {
	complete := timing.Start("cli.build", c.cli.mainPath)
	defer func() { complete(resultErr) }()
	if options.GOOS != runtime.GOOS || options.GOARCH != runtime.GOARCH || (runtime.GOOS != "darwin" && runtime.GOOS != "linux") {
		return "", lerrors.Unsupported.Errorf("only native macOS and Linux executable builds are supported")
	}
	frontend := timing.Start("frontend.generate", c.cli.mainPath)
	generated := c.process()
	frontend(c.err)
	if c.err != nil {
		return "", c.err
	}
	if generated == nil || generated.Compilation == nil {
		return "", lerrors.Internal.Errorf("source generation produced no program")
	}
	setup := timing.Start("workspace.prepare", c.cli.mainPath)
	defer func() { setup(resultErr) }()
	ids := generated.ApplicationIDs
	sources := generated.Sources
	if len(ids) != len(sources.ByteArrays) || len(ids) != len(sources.MapArrays) || len(ids) != len(sources.DependencyArrays) || len(ids) != len(sources.EnvDepArrays) {
		return "", lerrors.Internal.Errorf("application source bundle dimensions do not match")
	}
	stem := strings.TrimSuffix(filepath.Base(c.cli.mainPath), ".l")
	if stem == "" || stem == "." || stem == ".." {
		return "", lerrors.InvalidInput.Errorf("source has no usable executable name")
	}
	output := filepath.Join(options.OutputDir, stem)
	companion := output + ".apps"
	if err := validateOutput(output, companion, stem); err != nil {
		return "", err
	}
	canonicalOutput, err := canonicalBuildPath(output)
	if err != nil {
		return "", fmt.Errorf("resolve output destination: %w", err)
	}
	canonicalCompanion, err := canonicalBuildPath(companion)
	if err != nil {
		return "", fmt.Errorf("resolve companion destination: %w", err)
	}
	for _, sourcePath := range generated.Compilation.Compilation.SourcePaths {
		input, err := canonicalBuildPath(sourcePath)
		if err != nil {
			return "", err
		}
		replacesInput, err := buildPathContains(input, canonicalOutput)
		if err != nil {
			return "", err
		}
		containsInput, err := buildPathContains(canonicalCompanion, input)
		if err != nil {
			return "", err
		}
		if replacesInput || containsInput {
			return "", lerrors.InvalidInput.Errorf("output would replace source input %s", input)
		}
	}
	if c.options.AssetRoot != "" {
		if err := validateAssetDestinations(c.options.AssetRoot, canonicalOutput, canonicalCompanion); err != nil {
			return "", err
		}
	}
	bundles := make([]*sourceBundle, len(ids))
	for index, id := range ids {
		var err error
		bundles[index], err = planSourceBundle(sources.ByteArrays[index], sources.MapArrays[index])
		if err != nil {
			return "", fmt.Errorf("application %s: %w", id, err)
		}
		if len(sources.EnvDepArrays[index]) != 0 {
			return "", lerrors.Unsupported.Errorf("application %s requires unsupported environment dependencies", id)
		}
	}
	librarySources := generated.LibrarySources
	if len(librarySources.ByteArrays) != len(generated.LibraryIDs) || len(librarySources.MapArrays) != len(generated.LibraryIDs) || len(librarySources.DependencyArrays) != len(generated.LibraryIDs) || len(librarySources.EnvDepArrays) != len(generated.LibraryIDs) {
		return "", lerrors.Internal.Errorf("generated library source dimensions disagree")
	}
	libraries := make([]*sourceBundle, len(generated.LibraryIDs))
	for index, id := range generated.LibraryIDs {
		var err error
		libraries[index], err = planSourceBundle(librarySources.ByteArrays[index], librarySources.MapArrays[index])
		if err != nil {
			return "", fmt.Errorf("Service library %s: %w", id, err)
		}
		if len(librarySources.EnvDepArrays[index]) != 0 {
			return "", lerrors.Unsupported.Errorf("Service library %s requires unsupported environment dependencies", id)
		}
	}
	if err := os.MkdirAll(options.OutputDir, 0755); err != nil {
		return "", lerrors.IO.Errorf("create output directory: %w", err)
	}
	stage, err := os.MkdirTemp(options.OutputDir, "."+stem+"-build-")
	if err != nil {
		return "", lerrors.IO.Errorf("create build staging: %w", err)
	}
	defer func() { resultErr = cleanupBuildDirectory(stage, resultErr, os.RemoveAll) }()
	generation := strings.TrimPrefix(filepath.Base(stage), "."+stem+"-build-")
	generationDir := filepath.Join(stage, "generation")
	if err := os.Mkdir(generationDir, 0700); err != nil {
		return "", lerrors.IO.Wrap(err, "")
	}
	manifest := generationManifest{Version: 1, Entrypoint: stem, GOOS: options.GOOS, GOARCH: options.GOARCH}
	artifacts := make([]Lyger.AppArtifact, len(ids))
	destinations := make([]string, len(ids))
	for index, id := range ids {
		name := fmt.Sprintf("app-%03d", index)
		artifacts[index] = Lyger.AppArtifact{ID: id, RelativeExecutable: path.Join(stem+".apps", "generations", generation, name)}
		destinations[index] = filepath.Join(generationDir, name)
	}
	// The published filename is user-facing; a Cargo package identifier is a
	// separate compiler namespace and cannot contain arbitrary filename bytes.
	packageNames := append(append([]*sourceBundle(nil), bundles...), libraries...)
	entryPackage := orchestratorPackageName(stem, packageNames)
	unit, err := generated.Compilation.OrchestratorUnit(entryPackage, artifacts)
	if err != nil {
		return "", fmt.Errorf("lower orchestrator: %w", err)
	}
	orchestrator, runtimeSupport, err := c.log.EvaluateWorkspaceEntry(unit.BackendAML, Logiceme.WorkspaceRuntime{
		Package: runtimePackageName(entryPackage, packageNames), RelativeDirectory: "../support/runtime",
	})
	if err != nil {
		return "", fmt.Errorf("generate orchestrator: %w", err)
	}
	if len(orchestrator.ByteArrays) != 1 || len(orchestrator.MapArrays) != 1 {
		return "", lerrors.Internal.Errorf("orchestrator source bundle dimensions do not match")
	}
	orchestratorBundle, err := planSourceBundle(orchestrator.ByteArrays[0], orchestrator.MapArrays[0])
	if err != nil {
		return "", fmt.Errorf("orchestrator: %w", err)
	}
	if len(runtimeSupport.ByteArrays) != 1 || len(runtimeSupport.MapArrays) != 1 || len(runtimeSupport.EnvDepArrays) != 1 || len(runtimeSupport.EnvDepArrays[0]) != 0 {
		return "", lerrors.Internal.Errorf("runtime library source bundle dimensions do not match")
	}
	runtimeBundle, err := planSourceBundle(runtimeSupport.ByteArrays[0], runtimeSupport.MapArrays[0])
	if err != nil {
		return "", fmt.Errorf("runtime library: %w", err)
	}
	entry := filepath.Join(stage, "orchestrator")
	setup(nil)
	if err := compileCargoWorkspace(ctx, bundles, orchestratorBundle, libraries, runtimeBundle, filepath.Join(stage, "workspace"), destinations, entry); err != nil {
		return "", fmt.Errorf("build Rust workspace: %w", err)
	}
	for index, destination := range destinations {
		content, err := os.ReadFile(destination)
		if err != nil {
			return "", lerrors.IO.Wrap(err, "")
		}
		hash := sha256.Sum256(content)
		manifest.Apps = append(manifest.Apps, builtApp{ID: ids[index], Executable: filepath.Base(destination), SHA256: hex.EncodeToString(hash[:])})
	}
	encoded, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return "", lerrors.Internal.Wrap(err, "")
	}
	if err := os.WriteFile(filepath.Join(generationDir, "manifest.json"), append(encoded, '\n'), 0600); err != nil {
		return "", lerrors.IO.Wrap(err, "")
	}
	if err := ctx.Err(); err != nil {
		return "", lerrors.Tool.Wrap(err, "")
	}
	publish := timing.Start("publication.final", output)
	defer func() { publish(resultErr) }()
	if err := publishGeneration(output, companion, stem, generation, generationDir, entry, os.Rename); err != nil {
		return "", fmt.Errorf("publish output: %w", err)
	}
	return output, nil
}

// A development bundle may link individual resources or subdirectories to a
// shared location. Protect their physical targets as well as the asset roots.
// Each physical directory is visited once, even through cyclic directory links.
func validateAssetDestinations(assetRoot string, destinations ...string) error {
	var directories []os.FileInfo
	var visit func(string) error
	visit = func(name string) error {
		physical, err := canonicalBuildPath(name)
		if err != nil {
			return fmt.Errorf("resolve compiler assets: %w", err)
		}
		for _, destination := range destinations {
			contained, err := buildPathContains(physical, destination)
			if err != nil {
				return err
			}
			if contained {
				return lerrors.InvalidInput.Errorf("output would overwrite compiler assets in %s", physical)
			}
		}
		info, err := os.Stat(physical)
		if errors.Is(err, os.ErrNotExist) {
			return nil // Loading the selected resource diagnoses missing inputs.
		}
		if err != nil {
			return lerrors.IO.Errorf("read compiler assets %s: %w", physical, err)
		}
		if !info.IsDir() {
			return nil
		}
		for _, seen := range directories {
			if os.SameFile(info, seen) {
				return nil
			}
		}
		directories = append(directories, info)
		entries, err := os.ReadDir(physical)
		if err != nil {
			return lerrors.IO.Errorf("read compiler assets %s: %w", physical, err)
		}
		for _, entry := range entries {
			if err := visit(filepath.Join(physical, entry.Name())); err != nil {
				return err
			}
		}
		return nil
	}
	for _, asset := range []string{"builtins", "lexicon", "stdlib"} {
		if err := visit(filepath.Join(assetRoot, asset)); err != nil {
			return err
		}
	}
	return nil
}

// Resolve existing filesystem ancestors before appending an output suffix that
// has not been created yet. A dangling symlink is an error, not a missing folder.
func canonicalBuildPath(name string) (string, error) {
	current, err := filepath.Abs(name)
	if err != nil {
		return "", lerrors.IO.Wrap(err, "")
	}
	var missing []string
	for {
		resolved, err := filepath.EvalSymlinks(current)
		if err == nil {
			for index := len(missing) - 1; index >= 0; index-- {
				resolved = filepath.Join(resolved, missing[index])
			}
			return resolved, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return "", lerrors.IO.Wrap(err, "")
		}
		if _, existingErr := os.Lstat(current); !errors.Is(existingErr, os.ErrNotExist) {
			return "", lerrors.IO.Wrap(err, "")
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", lerrors.IO.Wrap(err, "")
		}
		missing = append(missing, filepath.Base(current))
		current = parent
	}
}

// Canonical spelling handles symlinks; filesystem identity also catches case
// and Unicode-equivalent directory names on filesystems that equate them.
func buildPathContains(parent, candidate string) (bool, error) {
	relative, err := filepath.Rel(parent, candidate)
	if err != nil {
		return false, lerrors.IO.Wrap(err, "")
	}
	if relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return true, nil
	}
	parentInfo, err := os.Stat(parent)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, lerrors.IO.Wrap(err, "")
	}
	for current := candidate; ; current = filepath.Dir(current) {
		info, err := os.Stat(current)
		if err == nil && os.SameFile(parentInfo, info) {
			return true, nil
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return false, lerrors.IO.Wrap(err, "")
		}
		if filepath.Dir(current) == current {
			return false, nil
		}
	}
}

func validateOutput(output, companion, stem string) error {
	// Publication installs the companion before the executable. Observe in the
	// reverse order so a concurrent first publication cannot make a stale
	// companion-absent observation reject its newly installed executable.
	outputInfo, outputErr := os.Lstat(output)
	if outputErr != nil && !errors.Is(outputErr, os.ErrNotExist) {
		return lerrors.IO.Wrap(outputErr, "")
	}
	info, err := os.Lstat(companion)
	owned := err == nil
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return lerrors.IO.Wrap(err, "")
	}
	if owned {
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return lerrors.InvalidInput.Errorf("companion path is not an owned directory: %s", companion)
		}
		marker, err := os.ReadFile(filepath.Join(companion, ".l-owner"))
		if err != nil {
			return lerrors.IO.Wrap(err, fmt.Sprintf("refusing unrelated companion directory %s", companion))
		}
		if string(marker) != companionOwner+stem+"\n" {
			return lerrors.InvalidInput.Errorf("refusing unrelated companion directory %s", companion)
		}
		info, err := os.Lstat(filepath.Join(companion, "generations"))
		if err != nil {
			return lerrors.IO.Wrap(err, fmt.Sprintf("invalid generation directory in %s", companion))
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return lerrors.InvalidInput.Errorf("invalid generation directory in %s", companion)
		}
	}
	if outputErr == nil {
		if !owned || !outputInfo.Mode().IsRegular() {
			return lerrors.InvalidInput.Errorf("refusing unrelated executable destination %s", output)
		}
	}
	return nil
}

// Commit the immutable generation before its referring executable. Old running
// orchestrators retain their old generations; completed generations are not pruned.
func publishGeneration(output, companion, stem, generation, generationDir, entry string, rename func(string, string) error) (resultErr error) {
	if err := validateOutput(output, companion, stem); err != nil {
		return err
	}
	if _, err := os.Lstat(companion); errors.Is(err, os.ErrNotExist) {
		// Publish ownership metadata as a complete directory, so interruption
		// cannot leave an unmarked companion that prevents future builds.
		prepared := filepath.Join(filepath.Dir(generationDir), "companion")
		if err := os.Mkdir(prepared, 0755); err != nil {
			return lerrors.IO.Wrap(err, "")
		}
		defer func() { resultErr = cleanupBuildDirectory(prepared, resultErr, os.RemoveAll) }()
		if err := os.WriteFile(filepath.Join(prepared, ".l-owner"), []byte(companionOwner+stem+"\n"), 0600); err != nil {
			return lerrors.IO.Wrap(err, "")
		}
		if err := os.Mkdir(filepath.Join(prepared, "generations"), 0755); err != nil {
			return lerrors.IO.Wrap(err, "")
		}
		if err := rename(prepared, companion); err != nil {
			// A concurrent publisher may have installed an owned directory.
			if _, exists := os.Lstat(companion); exists != nil {
				return lerrors.IO.Wrap(err, "")
			}
			if validation := validateOutput(output, companion, stem); validation != nil {
				return errors.Join(lerrors.IO.Wrap(err, ""), validation)
			}
		}
	}
	if err := validateOutput(output, companion, stem); err != nil {
		return err
	}
	if err := rename(generationDir, filepath.Join(companion, "generations", generation)); err != nil {
		return lerrors.IO.Wrap(err, "")
	}
	return lerrors.IO.Wrap(rename(entry, output), "")
}

func cleanupBuildDirectory(directory string, primary error, remove func(string) error) error {
	cleanupErr := remove(directory)
	// Cleanup cannot turn a committed publication into a reported failure.
	if cleanupErr == nil || primary == nil {
		return primary
	}
	return errors.Join(primary, lerrors.IO.Errorf("remove build staging %s: %w", directory, cleanupErr))
}
