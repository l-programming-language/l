package Logiceme

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/l-programming-language/l/lerrors"
	"io/fs"
	"path"
	"strings"
	"unicode/utf8"
)

func loadSupportTemplates(resources fs.FS, includeRuntime bool) (map[string][]byte, error) {
	names := []string{"config/config.lu", "install/cargo.lu", "operate/start.lu", "operate/stop.lu"}
	if includeRuntime {
		names = append(names, "runtime/main.lu", "runtime/host.lu", "runtime/process_unix.lu", "runtime/process_other.lu")
	}
	result := make(map[string][]byte, len(names))
	for _, name := range names {
		asset := path.Join("builtins/templates", name)
		content, err := readAsset(resources, asset)
		if err != nil {
			return nil, fmt.Errorf("load Rust support template %s: %w", asset, err)
		}
		requiresTags := name == "config/config.lu" || name == "install/cargo.lu"
		if err := validateLucidTemplate(content, requiresTags); err != nil {
			return nil, fmt.Errorf("load Rust support template %s: %w", asset, err)
		}
		result[name] = content
	}
	if includeRuntime {
		runtime, err := loadRuntimeComposition(resources, result)
		if err != nil {
			return nil, err
		}
		result["runtime/runtime.rs"] = runtime
	}
	return result, nil
}

// Existing support templates have one loop each. Expand only that known loop;
// native payloads are inserted once and are never reinterpreted as LU syntax.
func supportTemplateLoop(content []byte, expression string) (*LucidTag, error) {
	tags := CompileMainTags(content)
	if len(tags) != 1 || len(tags[0].Expressions) != 3 || strings.TrimSpace(tags[0].Expressions[1].ExpressionString) != expression {
		return nil, lerrors.InvalidConfiguration.Errorf("Rust support template requires exactly one %q loop", expression)
	}
	return tags[0], nil
}

func renderCargoManifest(content []byte, name, mainLocation string, dependencies []string) ([]byte, error) {
	loop, err := supportTemplateLoop(content, "for Dep in App.Deps")
	if err != nil {
		return nil, err
	}
	var rows strings.Builder
	for _, dependency := range dependencies {
		rows.WriteString(strings.NewReplacer("[|Dep|]", dependency).Replace(loop.Expressions[2].ExpressionString))
	}
	// JSON basic string escapes are also TOML basic string escapes. Go's JSON
	// encoder emits non-ASCII scalars directly rather than surrogate escapes.
	quotedName, _ := json.Marshal(name)             //nolint:errcheck // JSON marshaling a string cannot fail.
	quotedLocation, _ := json.Marshal(mainLocation) //nolint:errcheck // JSON marshaling a string cannot fail.
	return []byte(strings.NewReplacer(
		loop.ReplaceString, rows.String(),
		"[|App.Name|]", string(quotedName),
		"[|App.MainLocation|]", string(quotedLocation),
	).Replace(string(content))), nil
}

func renderConfig(content []byte, settings []MetaObj) ([]byte, error) {
	loop, err := supportTemplateLoop(content, "for Conf in Meta.Config")
	if err != nil {
		return nil, err
	}
	var rows strings.Builder
	for _, setting := range settings {
		key, err := quoteNativeString(setting.Key)
		if err != nil {
			return nil, err
		}
		value, err := quoteNativeString(setting.Value)
		if err != nil {
			return nil, err
		}
		rows.WriteString(strings.NewReplacer("[|Conf.Key|]", key, "[|Conf.Val|]", value).Replace(loop.Expressions[2].ExpressionString))
	}
	return []byte(strings.NewReplacer(loop.ReplaceString, rows.String()).Replace(string(content))), nil
}

func quoteNativeString(value string) (string, error) {
	if !utf8.ValidString(value) {
		return "", lerrors.InvalidInput.Errorf("Rust configuration strings must contain valid UTF-8")
	}
	var output strings.Builder
	output.WriteByte('"')
	for _, character := range value {
		switch character {
		case '"':
			output.WriteString(`\"`)
		case '\\':
			output.WriteString(`\\`)
		case '\n':
			output.WriteString(`\n`)
		case '\r':
			output.WriteString(`\r`)
		case '\t':
			output.WriteString(`\t`)
		default:
			if character < 0x20 || character == 0x7f {
				// strings.Builder writes never return errors.
				_, _ = fmt.Fprintf(&output, `\u{%x}`, character)
			} else {
				output.WriteRune(character)
			}
		}
	}
	output.WriteByte('"')
	return output.String(), nil
}

// The dedicated manifest template owns package/dependency layout. Explicit
// target topology replaces only its existing binary stanza, never source text.
func renderTargetCargoManifest(content []byte, name string, targets []preparedSourceTarget, dependencies []string) ([]byte, error) {
	const stanza = "[[bin]]\nname = [|App.Name|]\npath = [|App.MainLocation|]\n"
	if !bytes.Contains(content, []byte(stanza)) {
		return nil, lerrors.InvalidConfiguration.Errorf("Cargo template lacks the binary target stanza")
	}
	var entries strings.Builder
	entries.WriteString("autobins = false\nautolib = false\nautoexamples = false\nautotests = false\nautobenches = false\n\n")
	for _, target := range targets {
		if target.Kind == "lib" {
			entries.WriteString("[lib]\n")
		} else {
			entries.WriteString("[[bin]]\n")
		}
		targetName, _ := json.Marshal(target.Name)                        //nolint:errcheck // JSON marshaling a string cannot fail.
		location, _ := json.Marshal(strings.TrimPrefix(target.Path, "/")) //nolint:errcheck // JSON marshaling a string cannot fail.
		// strings.Builder writes never return errors.
		_, _ = fmt.Fprintf(&entries, "name = %s\npath = %s\n\n", targetName, location)
	}
	prepared := bytes.Replace(content, []byte(stanza), []byte(strings.TrimSuffix(entries.String(), "\n")), 1)
	return renderCargoManifest(prepared, name, "", dependencies)
}
