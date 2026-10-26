package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io/fs"
	"path"
	"strconv"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

type cargoDependency struct {
	name, version, local, packageName string
}

type cargoPackage struct {
	name, version, binary, entry string
	library                      string
	dependencies                 []cargoDependency
	explicitBinary               bool
	automaticBinaries            bool
}

// Read the bounded manifest vocabulary emitted by the backend. It is not a
// general TOML loader: unsupported Cargo hooks and dependency sources reject
// before a native tool can execute them. Cargo remains the manifest authority.
func readCargoPackage(source []byte) (*cargoPackage, error) {
	p := &cargoPackage{version: "0.0.0", entry: "src/main.rs", automaticBinaries: true}
	section := ""
	seen := map[string]bool{}
	binCount := 0
	scanner := bufio.NewScanner(bytes.NewReader(source))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") {
			switch line {
			case "[package]", "[dependencies]":
				section = line
			case "[[bin]]":
				binCount++
				p.explicitBinary = true
				if binCount > 1 {
					return nil, lerrors.InvalidInput.Errorf("Rust application requires one main binary target")
				}
				section = line
			case "[lib]":
				section = line
				p.library = "src/lib.rs"
			default:
				return nil, lerrors.Unsupported.Errorf("unsupported Rust manifest section %s", line)
			}
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		if !ok || key == "" {
			return nil, lerrors.InvalidInput.Errorf("invalid Rust manifest field %q", line)
		}
		identity := section + key
		if seen[identity] {
			return nil, lerrors.InvalidInput.Errorf("duplicate Rust manifest field %s", key)
		}
		seen[identity] = true
		switch section {
		case "[package]":
			switch key {
			case "name", "version", "edition":
				text, err := strconv.Unquote(value)
				if err != nil || text == "" {
					return nil, lerrors.InvalidInput.Errorf("invalid Rust package %s", key)
				}
				if key == "name" {
					p.name = text
				} else if key == "version" {
					p.version = text
				} else if text != "2024" {
					return nil, lerrors.Unsupported.Errorf("Rust generation requires edition 2024")
				}
			case "publish", "autobins", "autolib", "autotests", "autobenches", "autoexamples":
				if value != "false" {
					return nil, lerrors.Unsupported.Errorf("unsupported Rust package %s", key)
				}
				if key == "autobins" {
					p.automaticBinaries = false
				}
			default:
				return nil, lerrors.Unsupported.Errorf("unsupported Rust package field %s", key)
			}
		case "[[bin]]", "[lib]":
			if key != "name" && key != "path" {
				return nil, lerrors.Unsupported.Errorf("unsupported Rust target field %s", key)
			}
			text, err := strconv.Unquote(value)
			if err != nil || text == "" {
				return nil, lerrors.InvalidInput.Errorf("invalid Rust target %s", key)
			}
			if key == "path" && (!fs.ValidPath(text) || strings.ContainsAny(text, "\\:\x00")) {
				return nil, lerrors.InvalidInput.Errorf("invalid Rust target path %q", text)
			}
			if section == "[[bin]]" {
				if key == "name" {
					p.binary = text
				} else {
					p.entry = text
				}
			} else if key == "path" {
				p.library = text
			}
		case "[dependencies]":
			dependency, err := readCargoDependency(key, value)
			if err != nil {
				return nil, err
			}
			p.dependencies = append(p.dependencies, dependency)
		default:
			return nil, lerrors.InvalidInput.Errorf("Rust manifest field outside a supported section: %s", key)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, lerrors.InvalidInput.Wrap(err, "read Rust manifest")
	}
	if p.name == "" {
		return nil, lerrors.InvalidInput.Errorf("Rust manifest has no package name")
	}
	if p.binary == "" && (p.explicitBinary || p.automaticBinaries) {
		p.binary = p.name
	}
	for _, ch := range p.name {
		if !(ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' || ch == '_' || ch == '-') {
			return nil, lerrors.InvalidInput.Errorf("invalid Rust package name %q", p.name)
		}
	}
	return p, nil
}

func readCargoDependency(name, value string) (cargoDependency, error) {
	d := cargoDependency{name: name, packageName: name}
	if strings.HasPrefix(value, "{") && strings.HasSuffix(value, "}") {
		interior := strings.TrimSpace(value[1 : len(value)-1])
		first, _, _ := strings.Cut(interior, "=")
		if first = strings.TrimSpace(first); first == "path" || first == "package" {
			seen := map[string]bool{}
			for _, field := range strings.Split(interior, ",") {
				key, raw, pair := strings.Cut(strings.TrimSpace(field), "=")
				key = strings.TrimSpace(key)
				text, err := strconv.Unquote(strings.TrimSpace(raw))
				if !pair || err != nil || text == "" || seen[key] || (key != "path" && key != "package") {
					return d, lerrors.InvalidInput.Errorf("invalid local Rust dependency for %s", name)
				}
				seen[key] = true
				if key == "path" {
					d.local = text
				} else {
					d.packageName = text
				}
			}
			if d.local == "" || path.IsAbs(d.local) || path.Clean(d.local) != d.local || strings.ContainsAny(d.local, "\\:\x00") {
				return d, lerrors.InvalidInput.Errorf("invalid local Rust dependency path for %s", name)
			}
			return d, nil
		}
	}
	allowed := map[string]string{"tokio": "=1.53.1", "nix": "=0.31.3"}
	version := allowed[name]
	if version == "" {
		return d, lerrors.Unsupported.Errorf("unsupported external dependency %q", name)
	}
	if !strings.HasPrefix(value, "{") || !strings.HasSuffix(value, "}") {
		return d, lerrors.Unsupported.Errorf("dependency %s requires its pinned native manifest contract", name)
	}
	// Dependency interiors are the backend's version/features pair. No git,
	// registry, build, target, patch, or path override can be smuggled through.
	interior := strings.TrimSpace(value[1 : len(value)-1])
	first, rest, ok := strings.Cut(interior, ",")
	key, raw, pair := strings.Cut(first, "=")
	got, err := strconv.Unquote(strings.TrimSpace(raw))
	if !pair || strings.TrimSpace(key) != "version" || err != nil || got != version {
		return d, lerrors.Unsupported.Errorf("dependency %s must be pinned to %s", name, version)
	}
	if ok {
		key, raw, pair = strings.Cut(strings.TrimSpace(rest), "=")
		var features []string
		if !pair || strings.TrimSpace(key) != "features" || json.Unmarshal([]byte(strings.TrimSpace(raw)), &features) != nil {
			return d, lerrors.Unsupported.Errorf("unsupported dependency options for %s", name)
		}
		allowedFeatures := map[string]bool{}
		for _, feature := range strings.Fields("rt rt-multi-thread macros process io-util io-std signal sync time") {
			allowedFeatures[feature] = true
		}
		if name == "nix" {
			allowedFeatures = map[string]bool{"signal": true, "process": true}
		}
		for _, feature := range features {
			if !allowedFeatures[feature] {
				return d, lerrors.Unsupported.Errorf("unsupported dependency feature %s/%s", name, feature)
			}
		}
	}
	d.version = strings.TrimPrefix(version, "=")
	return d, nil
}
