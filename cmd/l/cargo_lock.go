package main

import (
	_ "embed"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// The external records are generated once from the pinned runtime manifest.
// Local workspace records are assembled from the validated package plans.
//
//go:embed cargo_dependencies.lock
var cargoDependencyLock string

func cargoWorkspaceLock(packages []*cargoPackage) ([]byte, error) {
	dependencies, retained, err := cargoLockDependencies(packages)
	if err != nil {
		return nil, err
	}
	var output strings.Builder
	output.WriteString("# Generated local package records; external packages are retained by L.\nversion = 4\n")
	for index, p := range packages {
		// strings.Builder writes always return a nil error.
		_, _ = fmt.Fprintf(&output, "\n[[package]]\nname = %s\nversion = %s\n", strconv.Quote(p.name), strconv.Quote(p.version))
		if len(p.dependencies) != 0 {
			output.WriteString("dependencies = [\n")
			names := dependencies[index]
			sort.Strings(names)
			for _, name := range names {
				_, _ = fmt.Fprintf(&output, " %s,\n", strconv.Quote(name))
			}
			output.WriteString("]\n")
		}
	}
	if retained != "" {
		output.WriteString("\n" + retained)
	}
	return []byte(output.String()), nil
}

// Lock references identify packages, not dependency aliases. Path packages
// have no source in Cargo.lock; registry identities retain their original URL.
type cargoLockIdentity struct {
	name, version, source string
}

type cargoLockReference struct {
	start, end int
	line       int
	original   cargoLockIdentity
	identity   cargoLockIdentity
}

type cargoRetainedLock struct {
	source     string
	identities []cargoLockIdentity
	references []cargoLockReference
}

// The embedded fragment has a deliberately bounded grammar. Retain its exact
// bytes and replace only dependency string tokens that need qualification.
// Unexpected retained data is an internal asset error, never fresh resolution.
func readCargoRetainedLock(source string) (*cargoRetainedLock, error) {
	lock := &cargoRetainedLock{source: source}
	var current *cargoLockIdentity
	var fields map[string]bool
	identities := map[cargoLockIdentity]bool{}
	inDependencies := false
	dependencyNeedsEnd := false
	recordLine := 0
	fail := func(line int, message string) error {
		return lerrors.Internal.Errorf("retained Rust lock line %d: %s", line, message)
	}
	finish := func() error {
		if current == nil {
			return nil
		}
		if inDependencies {
			return fail(recordLine, "unterminated dependencies")
		}
		for _, field := range []string{"name", "version", "source", "checksum"} {
			if !fields[field] {
				return fail(recordLine, "missing package "+field)
			}
		}
		if identities[*current] {
			return fail(recordLine, fmt.Sprintf("duplicate package identity %s %s (%s)", current.name, current.version, current.source))
		}
		identities[*current] = true
		lock.identities = append(lock.identities, *current)
		return nil
	}
	offset := 0
	for index, raw := range strings.SplitAfter(source, "\n") {
		lineNumber, lineOffset := index+1, offset
		offset += len(raw)
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if line == "[[package]]" {
			if err := finish(); err != nil {
				return nil, err
			}
			current, fields, recordLine = &cargoLockIdentity{}, map[string]bool{}, lineNumber
			continue
		}
		if current == nil {
			return nil, fail(lineNumber, "field outside package record")
		}
		if inDependencies {
			if line == "]" {
				inDependencies = false
				continue
			}
			if dependencyNeedsEnd {
				return nil, fail(lineNumber, "missing comma between dependencies")
			}
			token := strings.TrimSpace(strings.TrimSuffix(line, ","))
			value, err := strconv.Unquote(token)
			if err != nil {
				return nil, fail(lineNumber, "malformed dependency string")
			}
			parts := strings.Fields(value)
			if len(parts) < 1 || len(parts) > 3 {
				return nil, fail(lineNumber, "malformed dependency identity")
			}
			id := cargoLockIdentity{name: parts[0]}
			if len(parts) > 1 {
				id.version = parts[1]
			}
			if len(parts) == 3 {
				if !strings.HasPrefix(parts[2], "(") || !strings.HasSuffix(parts[2], ")") || len(parts[2]) <= 2 {
					return nil, fail(lineNumber, "malformed dependency source")
				}
				id.source = parts[2][1 : len(parts[2])-1]
			}
			start := lineOffset + strings.Index(raw, token)
			lock.references = append(lock.references, cargoLockReference{start: start, end: start + len(token), line: lineNumber, original: id, identity: id})
			dependencyNeedsEnd = !strings.HasSuffix(line, ",")
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		if !ok || fields[key] {
			return nil, fail(lineNumber, "invalid or duplicate package field "+key)
		}
		fields[key] = true
		if key == "dependencies" {
			if value != "[" {
				return nil, fail(lineNumber, "malformed dependencies array")
			}
			inDependencies = true
			dependencyNeedsEnd = false
			continue
		}
		switch key {
		case "name", "version", "source", "checksum":
			text, err := strconv.Unquote(value)
			if err != nil || text == "" {
				return nil, fail(lineNumber, "invalid package "+key)
			}
			switch key {
			case "name":
				current.name = text
			case "version":
				current.version = text
			case "source":
				current.source = text
			}
		default:
			return nil, fail(lineNumber, "unsupported package field "+key)
		}
	}
	if err := finish(); err != nil {
		return nil, err
	}
	if len(lock.identities) == 0 {
		return nil, fail(1, "no package records")
	}
	// Resolve against the retained graph alone, before local names can change
	// the interpretation of a short reference (including windows-sys versions).
	for index := range lock.references {
		reference := &lock.references[index]
		id, ok := findCargoLockIdentity(lock.identities, reference.identity)
		if !ok {
			return nil, fail(reference.line, "unresolved or ambiguous dependency "+reference.identity.name)
		}
		reference.identity = id
	}
	return lock, nil
}

func findCargoLockIdentity(identities []cargoLockIdentity, reference cargoLockIdentity) (cargoLockIdentity, bool) {
	var result cargoLockIdentity
	found := false
	for _, id := range identities {
		if id.name != reference.name || reference.version != "" && id.version != reference.version || reference.source != "" && id.source != reference.source {
			continue
		}
		if found {
			return cargoLockIdentity{}, false
		}
		result, found = id, true
	}
	return result, found
}

// Match Cargo's v4 encoder: a unique name is short; repeated names require a
// version, and repeated name/version pairs require a non-path source as well.
func cargoLockIdentityText(id cargoLockIdentity, identities []cargoLockIdentity) string {
	nameCount, versionCount := 0, 0
	for _, other := range identities {
		if other.name == id.name {
			nameCount++
			if other.version == id.version {
				versionCount++
			}
		}
	}
	if nameCount == 1 {
		return id.name
	}
	text := id.name + " " + id.version
	if versionCount > 1 && id.source != "" {
		text += " (" + id.source + ")"
	}
	return text
}

func cargoLockDependencies(packages []*cargoPackage) ([][]string, string, error) {
	locals := make(map[string]cargoLockIdentity, len(packages))
	var identities []cargoLockIdentity
	withExternal := false
	for _, p := range packages {
		if p == nil || p.name == "" || p.version == "" {
			return nil, "", lerrors.InvalidInput.Errorf("Rust lock requires named and versioned local packages")
		}
		if _, found := locals[p.name]; found {
			return nil, "", lerrors.InvalidInput.Errorf("duplicate Rust workspace package %q", p.name)
		}
		id := cargoLockIdentity{name: p.name, version: p.version}
		locals[p.name] = id
		identities = append(identities, id)
		for _, dependency := range p.dependencies {
			withExternal = withExternal || dependency.local == ""
		}
	}
	var retained cargoRetainedLock
	if withExternal {
		parsed, err := readCargoRetainedLock(cargoDependencyLock)
		if err != nil {
			return nil, "", err
		}
		retained = *parsed
		identities = append(identities, retained.identities...)
	}
	dependencies := make([][]string, len(packages))
	for index, p := range packages {
		for _, dependency := range p.dependencies {
			var id cargoLockIdentity
			var found bool
			if dependency.local != "" {
				id, found = locals[dependency.packageName]
				if !found {
					return nil, "", lerrors.InvalidInput.Errorf("package %s local dependency %s references missing workspace package %s", p.name, dependency.name, dependency.packageName)
				}
			} else {
				id, found = findCargoLockIdentity(retained.identities, cargoLockIdentity{name: dependency.packageName, version: dependency.version})
				if !found {
					return nil, "", lerrors.Internal.Errorf("retained Rust lock lacks unique pinned dependency %s %s for %s", dependency.packageName, dependency.version, p.name)
				}
			}
			dependencies[index] = append(dependencies[index], cargoLockIdentityText(id, identities))
		}
	}
	var source strings.Builder
	offset := 0
	for _, reference := range retained.references {
		source.WriteString(retained.source[offset:reference.start])
		// Preserve even redundant original qualifiers when the reference still
		// uniquely identifies its retained package in the combined graph.
		resolved, unique := findCargoLockIdentity(identities, reference.original)
		if unique && resolved == reference.identity {
			source.WriteString(retained.source[reference.start:reference.end])
		} else {
			source.WriteString(strconv.Quote(cargoLockIdentityText(reference.identity, identities)))
		}
		offset = reference.end
	}
	source.WriteString(retained.source[offset:])
	return dependencies, source.String(), nil
}
