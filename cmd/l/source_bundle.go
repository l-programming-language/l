package main

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

type sourceFile struct {
	name    string
	content []byte
}

type sourceBundle struct {
	files       []sourceFile
	packagePlan *cargoPackage
}

// Validate the complete logical namespace before creating any source files.
func planSourceBundle(contents [][]byte, names []string) (*sourceBundle, error) {
	if len(contents) != len(names) {
		return nil, lerrors.InvalidInput.Errorf("source bundle dimensions do not match")
	}
	if len(contents) == 0 {
		return nil, lerrors.InvalidInput.Errorf("source bundle is empty")
	}
	bundle := &sourceBundle{}
	seen := make(map[string]bool)
	var manifest []byte
	for index, logical := range names {
		name := strings.TrimPrefix(logical, "/")
		if !fs.ValidPath(name) || name == "." || strings.ContainsAny(name, "\\:\x00") {
			return nil, lerrors.InvalidInput.Errorf("invalid generated path %q", logical)
		}
		folded := strings.ToLower(name)
		if seen[folded] {
			return nil, lerrors.InvalidInput.Errorf("duplicate generated path %q", logical)
		}
		seen[folded] = true
		if strings.HasPrefix(folded, ".cargo/") || path.Base(folded) == "rust-toolchain" || path.Base(folded) == "rust-toolchain.toml" || path.Base(folded) == "build.rs" {
			return nil, lerrors.InvalidInput.Errorf("reserved build metadata path %q", logical)
		}
		if strings.HasPrefix(folded, "tests/") || strings.HasPrefix(folded, "benches/") || strings.HasPrefix(folded, "examples/") {
			return nil, lerrors.InvalidInput.Errorf("generated test or example source is not a build input: %s", name)
		}
		if strings.HasSuffix(folded, ".go") || path.Base(folded) == "go.mod" || path.Base(folded) == "go.work" {
			return nil, lerrors.InvalidInput.Errorf("mixed native source in Rust bundle: %s", name)
		}
		if folded == "cargo.toml" {
			if name != "Cargo.toml" {
				return nil, lerrors.InvalidInput.Errorf("Rust manifest must be named Cargo.toml")
			}
			manifest = contents[index]
		} else if path.Base(folded) == "cargo.toml" {
			return nil, lerrors.Unsupported.Errorf("nested package manifest requires an explicit workspace package: %s", name)
		}
		bundle.files = append(bundle.files, sourceFile{name, append([]byte(nil), contents[index]...)})
	}
	for name := range seen {
		for parent := path.Dir(name); parent != "."; parent = path.Dir(parent) {
			if seen[parent] {
				return nil, lerrors.InvalidInput.Errorf("generated file/directory collision at %s", parent)
			}
		}
	}
	if manifest == nil {
		return nil, lerrors.Internal.Errorf("generated Rust main package requires Cargo.toml")
	}
	plan, err := readCargoPackage(manifest)
	if err != nil {
		return nil, err
	}
	if plan.library != "" && !seen[strings.ToLower(plan.library)] {
		return nil, lerrors.Internal.Errorf("Rust library target %s has no generated source", plan.library)
	}
	if plan.binary != "" && !seen[strings.ToLower(plan.entry)] && plan.library != "" && !plan.explicitBinary {
		plan.binary = ""
	} else if plan.binary != "" && !seen[strings.ToLower(plan.entry)] {
		return nil, lerrors.Internal.Errorf("Rust main target %s has no generated source", plan.entry)
	} else if plan.binary == "" && plan.library == "" {
		return nil, lerrors.Internal.Errorf("Rust package has no declared source target")
	}
	bundle.packagePlan = plan
	sort.Slice(bundle.files, func(i, j int) bool { return bundle.files[i].name < bundle.files[j].name })
	return bundle, nil
}

func (bundle *sourceBundle) write(directory string) error {
	for _, file := range bundle.files {
		name := filepath.Join(directory, filepath.FromSlash(file.name))
		if err := os.MkdirAll(filepath.Dir(name), 0700); err != nil {
			return lerrors.IO.Wrap(err, "")
		}
		if err := createSourceFile(name, file.content); err != nil {
			return err
		}
	}
	return nil
}

// Exclusive creation also catches filesystem-equivalent Unicode names that a
// logical path comparison cannot reliably normalize on every supported host.
func createSourceFile(name string, content []byte) error {
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return lerrors.IO.Wrap(err, "")
	}
	_, err = file.Write(content)
	return lerrors.IO.Wrap(errors.Join(err, file.Close()), "")
}
