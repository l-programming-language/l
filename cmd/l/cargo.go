package main

// Cargo is a native build boundary, not a second frontend. This file consumes
// the source bundle as supplied and never changes generated Rust source text.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/l-programming-language/l/internal/cargocache"
	"github.com/l-programming-language/l/internal/timing"
	"github.com/l-programming-language/l/lerrors"
)

const nativeToolchainVersion = "1.98.1"

type cargoArtifact struct {
	Reason, PackageID, ManifestPath, Executable string
	InputManifestPath                           string
	Target                                      struct {
		Name string
		Kind []string
	}
	Profile struct{ Test bool }
	Message struct{ Level, Rendered string }
}

func cargoMessages(output []byte) (artifacts []cargoArtifact, sourceError bool) {
	for _, line := range bytes.Split(output, []byte{'\n'}) {
		var message cargoArtifact
		// Cargo's two underscored fields need explicit decoding; the remainder
		// matches case-insensitively with Go's JSON names.
		var wire struct {
			cargoArtifact
			PackageID    string `json:"package_id"`
			ManifestPath string `json:"manifest_path"`
		}
		if json.Unmarshal(line, &wire) != nil {
			continue
		}
		message = wire.cargoArtifact
		message.PackageID, message.ManifestPath = wire.PackageID, wire.ManifestPath
		if message.Reason == "compiler-message" && message.Message.Level == "error" {
			sourceError = true
		}
		if message.Reason == "compiler-artifact" && message.Executable != "" && !message.Profile.Test && len(message.Target.Kind) == 1 && message.Target.Kind[0] == "bin" {
			artifacts = append(artifacts, message)
		}
	}
	return artifacts, sourceError
}

func cargoBuildEnvironment() []string {
	overrides := map[string]string{"RUSTUP_TOOLCHAIN": nativeToolchainVersion, "CARGO_NET_OFFLINE": "true", "RUSTFLAGS": "", "RUSTDOCFLAGS": "", "CARGO_ENCODED_RUSTFLAGS": "", "RUSTC_WRAPPER": "", "RUSTC_WORKSPACE_WRAPPER": "", "CARGO_TERM_COLOR": "never", "CARGO_PROFILE_DEV_DEBUG": "0"}
	var env []string
	for _, item := range os.Environ() {
		key, _, _ := strings.Cut(item, "=")
		if _, replace := overrides[key]; replace || key == "RUSTC" || key == "RUSTDOC" || strings.HasPrefix(key, "CARGO_TARGET_") || strings.HasPrefix(key, "CARGO_PROFILE_") || strings.HasPrefix(key, "CARGO_BUILD_") {
			continue
		}
		env = append(env, item)
	}
	keys := make([]string, 0, len(overrides))
	for key := range overrides {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		env = append(env, key+"="+overrides[key])
	}
	return env
}

func cargoBuild(ctx context.Context, directory string) (result []cargoArtifact, resultErr error) {
	complete := timing.Start("build.native", directory)
	defer func() { complete(resultErr) }()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	target, unlock, err := cargoBuildCache(ctx, directory)
	if err != nil {
		return nil, err
	}
	defer unlock()
	stage := timing.Start("cache.stage", directory)
	defer func() { stage(resultErr) }()
	inputs, err := cargocache.StageCurrent(directory, filepath.Join(filepath.Dir(target), "inputs"), "")
	if err != nil {
		return nil, lerrors.IO.Wrap(err, "retain Rust workspace input paths")
	}
	if err := cargocache.Activate(inputs, target); err != nil {
		return nil, lerrors.IO.Wrap(err, "select Rust workspace cache inputs")
	}
	stage(nil)
	// Default Cargo target selection builds all workspace binaries and libraries.
	command := exec.CommandContext(ctx, "cargo", "+"+nativeToolchainVersion, "build", "--workspace", "--frozen", "--message-format=json", "--manifest-path", filepath.Join(inputs, "Cargo.toml"), "--target", "host-tuple", "--target-dir", target)
	command.Dir, command.Env = directory, cargoBuildEnvironment()
	configureBuildCommand(command)
	// Cargo's stdout is a framed machine stream. Human diagnostics may arrive
	// on stderr between parts of a JSON record and must never be merged into it.
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	invoke := timing.Start("cargo.wrapper", directory)
	err = command.Run()
	invoke(err)
	cleanup := timing.Start("build.cleanup", directory)
	if command.Process != nil && command.Cancel != nil {
		if cancelErr := command.Cancel(); cancelErr != nil && !errors.Is(cancelErr, os.ErrProcessDone) {
			cleanup(cancelErr)
			err = errors.Join(err, cancelErr)
		}
	}
	cleanup(nil)
	artifacts, sourceError := cargoMessages(stdout.Bytes())
	diagnostics := append(bytes.Clone(stdout.Bytes()), stderr.Bytes()...)
	if ctx.Err() != nil {
		return nil, lerrors.Tool.Wrap(errors.Join(ctx.Err(), err), fmt.Sprintf("Rust compilation interrupted: %v\n%s", ctx.Err(), diagnostics))
	}
	if err != nil {
		if sourceError {
			return nil, lerrors.Internal.Errorf("invalid generated Rust: %w\n%s", err, diagnostics)
		}
		return nil, lerrors.Tool.Errorf("Rust compilation: %w\n%s", err, diagnostics)
	}
	// Cargo's target-name paths may be reused by another workspace. Snapshot
	// every actual executable while the cross-process lock is still held.
	snapshot := timing.Start("publication.snapshot", directory)
	defer func() { snapshot(resultErr) }()
	artifactDirectory := filepath.Join(directory, ".l-artifacts")
	if err := os.Mkdir(artifactDirectory, 0700); err != nil {
		return nil, lerrors.IO.Wrap(err, "create private Rust artifact staging")
	}
	for index := range artifacts {
		manifest, err := cargoArtifactInputManifest(artifacts[index].ManifestPath, inputs, directory)
		if err != nil {
			return nil, err
		}
		artifacts[index].InputManifestPath = manifest
		snapshot := filepath.Join(artifactDirectory, strconv.Itoa(index))
		if err := copyCargoArtifact(artifacts[index].Executable, snapshot); err != nil {
			return nil, err
		}
		artifacts[index].Executable = snapshot
	}
	return artifacts, nil
}
