package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/l-programming-language/l/internal/cargonames"
	"github.com/l-programming-language/l/lerrors"
)

func orchestratorPackageName(stem string, applications []*sourceBundle) string {
	used := map[string]bool{}
	for _, application := range applications {
		if application.packagePlan != nil {
			used[strings.ToLower(application.packagePlan.name)] = true
			if application.packagePlan.binary != "" {
				used[strings.ToLower(application.packagePlan.binary)] = true
			}
		}
	}
	valid := stem != ""
	for index, char := range stem {
		valid = valid && (char == '_' || char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || index > 0 && (char >= '0' && char <= '9' || char == '-'))
	}
	if valid && !cargonames.IsReservedBinary(stem) && !used[strings.ToLower(stem)] {
		return stem
	}
	name := "l_entry"
	for index := 1; used[strings.ToLower(name)]; index++ {
		name = fmt.Sprintf("l_entry_%d", index)
	}
	return name
}

func runtimePackageName(entry string, packages []*sourceBundle) string {
	native := func(name string) string { return strings.ToLower(strings.ReplaceAll(name, "-", "_")) }
	used := map[string]bool{native(entry): true}
	for _, bundle := range packages {
		if bundle == nil || bundle.packagePlan == nil {
			continue
		}
		used[native(bundle.packagePlan.name)] = true
		used[native(bundle.packagePlan.library)] = true
		for _, dependency := range bundle.packagePlan.dependencies {
			used[native(dependency.name)] = true
		}
	}
	name := "l_runtime_support"
	for suffix := 1; used[native(name)]; suffix++ {
		name = fmt.Sprintf("l_runtime_support_%d", suffix)
	}
	return name
}

// Each package retains its complete source namespace. A package can expose a
// library alongside its binary; shared Services can be separate library-only
// members. Publication destinations never depend on Cargo target-directory
// naming: they are populated from successful compiler-artifact messages.
func compileCargoWorkspace(ctx context.Context, applications []*sourceBundle, orchestrator *sourceBundle, libraries []*sourceBundle, support *sourceBundle, directory string, destinations []string, entry string) error {
	if len(applications) != len(destinations) {
		return lerrors.Internal.Errorf("Rust workspace application dimensions do not match")
	}
	if orchestrator == nil || orchestrator.packagePlan == nil || orchestrator.packagePlan.binary == "" {
		return lerrors.Internal.Errorf("Rust workspace has no orchestrator binary")
	}
	type member struct {
		path   string
		bundle *sourceBundle
		output string
	}
	members := []member{{"orchestrator", orchestrator, entry}}
	for index, bundle := range applications {
		if bundle == nil || bundle.packagePlan == nil || bundle.packagePlan.binary == "" {
			return lerrors.Internal.Errorf("Rust App %d has no binary target", index)
		}
		members = append(members, member{fmt.Sprintf("apps/app-%03d", index), bundle, destinations[index]})
	}
	for index, bundle := range libraries {
		if bundle == nil || bundle.packagePlan == nil || bundle.packagePlan.library == "" || bundle.packagePlan.binary != "" {
			return lerrors.Internal.Errorf("Rust Service %d requires a library-only package", index)
		}
		members = append(members, member{fmt.Sprintf("services/service-%03d", index), bundle, ""})
	}
	if support != nil {
		if support.packagePlan == nil || support.packagePlan.library == "" || support.packagePlan.binary != "" {
			return lerrors.Internal.Errorf("Rust runtime requires a library-only package")
		}
		members = append(members, member{"support/runtime", support, ""})
	}
	var manifest strings.Builder
	manifest.WriteString("[workspace]\nresolver = \"3\"\nmembers = [\n")
	seen := map[string]bool{}
	binaries := map[string]*cargoPackage{}
	var packages []*cargoPackage
	byPath := map[string]*cargoPackage{}
	for _, member := range members {
		p := member.bundle.packagePlan
		if seen[p.name] {
			return lerrors.InvalidInput.Errorf("duplicate Rust workspace package %q", p.name)
		}
		seen[p.name] = true
		// Workspace binaries share Cargo's output directory. Distinct package
		// identities do not protect case-equivalent executable filenames.
		if p.binary != "" {
			key := strings.ToLower(p.binary)
			if prior := binaries[key]; prior != nil {
				return lerrors.InvalidInput.Errorf("Rust workspace binary targets %q (package %s) and %q (package %s) collide on case-insensitive filesystems", prior.binary, prior.name, p.binary, p.name)
			}
			binaries[key] = p
		}
		packages = append(packages, p)
		byPath[member.path] = p
		// strings.Builder writes always return a nil error.
		_, _ = fmt.Fprintf(&manifest, " %s,\n", strconv.Quote(member.path))
	}
	for _, member := range members {
		for _, dependency := range member.bundle.packagePlan.dependencies {
			if dependency.local == "" {
				continue
			}
			target := byPath[path.Join(member.path, dependency.local)]
			if target == nil || target.name != dependency.packageName || target.library == "" {
				return lerrors.InvalidInput.Errorf("package %s local dependency %s does not resolve to a workspace library", member.bundle.packagePlan.name, dependency.name)
			}
		}
	}
	manifest.WriteString("]\n")
	lock, err := cargoWorkspaceLock(packages)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(directory, 0700); err != nil {
		return lerrors.IO.Wrap(err, "create Rust workspace")
	}
	if err := createSourceFile(filepath.Join(directory, "Cargo.toml"), []byte(manifest.String())); err != nil {
		return err
	}
	if err := createSourceFile(filepath.Join(directory, "Cargo.lock"), lock); err != nil {
		return err
	}
	for _, member := range members {
		if err := member.bundle.write(filepath.Join(directory, filepath.FromSlash(member.path))); err != nil {
			return err
		}
	}
	artifacts, err := cargoBuild(ctx, directory)
	if err != nil {
		return err
	}
	for _, member := range members {
		if member.output == "" {
			continue
		}
		manifestPath, err := filepath.EvalSymlinks(filepath.Join(directory, filepath.FromSlash(member.path), "Cargo.toml"))
		if err != nil {
			return lerrors.IO.Wrap(err, "resolve Rust package manifest")
		}
		var selected string
		for _, artifact := range artifacts {
			if artifact.Target.Name != member.bundle.packagePlan.binary {
				continue
			}
			actual, err := filepath.EvalSymlinks(artifact.InputManifestPath)
			if err != nil || actual != manifestPath {
				continue
			}
			if selected != "" {
				return lerrors.Tool.Errorf("Cargo returned duplicate artifacts for package %s", member.bundle.packagePlan.name)
			}
			selected = artifact.Executable
		}
		if selected == "" {
			return lerrors.Tool.Errorf("Cargo returned no main artifact for package %s", member.bundle.packagePlan.name)
		}
		if err := copyCargoArtifact(selected, member.output); err != nil {
			return err
		}
	}
	return nil
}
