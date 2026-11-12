package Logiceme

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/l-programming-language/l/lerrors"
	"path"
	"strings"
	"unicode/utf8"
)

func (p *Program) linkBundleSupport(templates map[string][]byte) error {
	apps := p.Compiler.CompiledApps
	if len(apps) == 0 || len(apps) != len(p.Scanner.Program.Books) || len(apps) != len(p.Executable.MapArrays) || len(apps) != len(p.Executable.ByteArrays) {
		return lerrors.Internal.Errorf("Rust assembly requires one complete bundle per application book")
	}
	for i := range apps {
		app := &apps[i]
		if err := validateBundleAppName(app.Name); err != nil {
			return err
		}
		if len(p.Executable.MapArrays[i]) != len(p.Executable.ByteArrays[i]) {
			return lerrors.Internal.Errorf("Rust application %q has unequal source/path lengths", app.Name)
		}
		var seen []string
		for _, name := range p.Executable.MapArrays[i] {
			if err := validateBundlePath(name); err != nil {
				return err
			}
			for _, existing := range seen {
				if bundlePathsCollide(name, existing) {
					return lerrors.InvalidInput.Errorf("Rust application %q source path %s collides with %s", app.Name, name, existing)
				}
			}
			seen = append(seen, name)
		}
		mainDirectory := app.Outline.Meta.GetVal("AppGenDir")
		if mainDirectory == "NotFound" || mainDirectory == "" {
			mainDirectory = "/src"
		}
		if mainDirectory != "/" {
			if err := validateBundlePath(mainDirectory); err != nil {
				return fmt.Errorf("Rust application %q AppGenDir: %w", app.Name, err)
			}
		}
		libraryOnly := libraryOnlyTargets(app.sourceTargets)
		mainPath := path.Join(mainDirectory, "main.rs")
		binaryName := app.Name
		for _, target := range app.sourceTargets {
			if target.Primary || libraryOnly {
				mainPath = target.Path
				mainDirectory = path.Dir(mainPath)
				binaryName = target.Name
			}
		}
		mainIndex := -1
		for j, name := range p.Executable.MapArrays[i] {
			if name == mainPath {
				mainIndex = j
				break
			}
		}
		if mainIndex < 0 {
			return lerrors.InvalidInput.Errorf("Rust application %q requires generated entry source %s", app.Name, mainPath)
		}
		dependencies, err := appDependencies(app.Outline.Meta)
		if err != nil {
			return fmt.Errorf("Rust application %q dependencies: %w", app.Name, err)
		}
		settings, err := bookSettings(p.Scanner.Program.Tags, p.Scanner.Program.Books[i])
		if err != nil {
			return fmt.Errorf("Rust application %q configuration: %w", app.Name, err)
		}
		cargo, err := renderCargoManifest(templates["install/cargo.lu"], app.Name, strings.TrimPrefix(mainPath, "/"), dependencies)
		if err != nil {
			return err
		}
		if len(app.sourceTargets) > 0 {
			cargo, err = renderTargetCargoManifest(templates["install/cargo.lu"], app.Name, app.sourceTargets, dependencies)
			if err != nil {
				return err
			}
		}
		files := []struct {
			name    string
			content []byte
		}{{"/Cargo.toml", cargo}}
		if !libraryOnly {
			config, err := renderConfig(templates["config/config.lu"], settings)
			if err != nil {
				return err
			}
			start := strings.NewReplacer("[|App.Name|]", shellQuoted(binaryName)).Replace(string(templates["operate/start.lu"]))
			files = append(files, []struct {
				name    string
				content []byte
			}{
				{path.Join(mainDirectory, "config.rs"), config},
				{"/start.sh", []byte(start)},
				{"/stop.sh", bytes.Clone(templates["operate/stop.lu"])},
			}...)
		}
		for _, file := range files {
			for _, existing := range seen {
				if bundlePathsCollide(file.name, existing) {
					return lerrors.InvalidInput.Errorf("Rust support path %s collides with %s in application %q", file.name, existing, app.Name)
				}
			}
			seen = append(seen, file.name)
		}
		for _, file := range files {
			p.Executable.MapArrays[i] = append(p.Executable.MapArrays[i], file.name)
			p.Executable.ByteArrays[i] = append(p.Executable.ByteArrays[i], file.content)
		}
		p.Executable.DependencyArrays[i] = append([]string(nil), dependencies...)
	}
	return nil
}

func bundlePathsCollide(left, right string) bool {
	return left == right || strings.HasPrefix(left, right+"/") || strings.HasPrefix(right, left+"/")
}

func validateBundleAppName(name string) error {
	if name == "" || !utf8.ValidString(name) {
		return lerrors.InvalidInput.Errorf("Rust application name must be a nonempty UTF-8 Cargo name")
	}
	for _, character := range name {
		if !(character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9' || character == '_' || character == '-') {
			return lerrors.InvalidInput.Errorf("Rust application name %q must contain only ASCII letters, digits, underscores, or hyphens", name)
		}
	}
	return nil
}

func validateBundlePath(name string) error {
	if !strings.HasPrefix(name, "/") || name == "/" || path.Clean(name) != name || strings.ContainsAny(name, "\\\x00\r\n") {
		return lerrors.InvalidInput.Errorf("Rust bundle path %q must be a canonical project-relative logical path beginning with /", name)
	}
	return nil
}

// Deps is one AppMeta string containing a JSON array of complete normalized
// TOML dependency rows. Generic metadata array splitting is not suitable for
// rows that themselves contain feature lists and commas.
func appDependencies(meta MetaDict) ([]string, error) {
	raw := meta.GetVal("Deps")
	if raw == "NotFound" {
		return nil, nil
	}
	var rows []string
	if err := json.Unmarshal([]byte(raw), &rows); err != nil {
		return nil, lerrors.InvalidInput.Errorf("AppMeta.Deps requires a JSON-encoded array of TOML dependency rows: %w", err)
	}
	if rows == nil {
		return nil, lerrors.InvalidInput.Errorf("AppMeta.Deps must be an array, not null")
	}
	seen := make(map[string]bool)
	for _, row := range rows {
		key, value, present := strings.Cut(row, "=")
		key = strings.TrimSpace(key)
		if !present || key == "" || strings.TrimSpace(value) == "" || strings.TrimSpace(row) != row || strings.ContainsAny(row, "\r\n\x00") {
			return nil, lerrors.InvalidInput.Errorf("AppMeta.Deps entry must be one normalized TOML dependency assignment: %q", row)
		}
		for _, character := range key {
			if !(character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9' || character == '_' || character == '-') {
				return nil, lerrors.InvalidInput.Errorf("AppMeta.Deps entry requires a bare dependency key: %q", key)
			}
		}
		if seen[key] {
			return nil, lerrors.InvalidInput.Errorf("AppMeta.Deps contains duplicate dependency %q", key)
		}
		seen[key] = true
	}
	return rows, nil
}

func shellQuoted(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}
