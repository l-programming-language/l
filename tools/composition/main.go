// Command composition regenerates canonical support AML from source words.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	Lyger "github.com/l-programming-language/l/lyger"
)

const runtimeSource = "builtins/runtime/eng/dictionary.l"

type manifest struct {
	Version int               `json:"version"`
	SHA256  string            `json:"sha256"`
	Sources map[string]string `json:"sources"`
}

func main() {
	check := flag.Bool("check", false, "verify the committed composition without writing")
	flag.Parse()
	if err := run(".", *check); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(root string, check bool) error {
	aml, metadata, err := generate(root)
	if err != nil {
		return err
	}
	for _, output := range []struct {
		name string
		data []byte
	}{{"builtins/runtime/support.aml", aml}, {"builtins/runtime/support.json", metadata}} {
		path := filepath.Join(root, filepath.FromSlash(output.name))
		if check {
			committed, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if !bytes.Equal(committed, output.data) {
				return fmt.Errorf("%s is stale; regenerate with scripts/l-go run ./tools/composition", output.name)
			}
		} else if err := os.WriteFile(path, output.data, 0644); err != nil {
			return err
		}
	}
	return nil
}

func generate(root string) ([]byte, []byte, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, nil, err
	}
	resources := os.DirFS(root)
	source, err := fs.ReadFile(resources, runtimeSource)
	if err != nil {
		return nil, nil, err
	}
	c := Lyger.NewLygerService(Lyger.Options{Resources: resources, AssetRoot: root}).NewCompilation(source, filepath.Join(root, filepath.FromSlash(runtimeSource)))
	c.Make()
	if c.Error != nil {
		return nil, nil, c.Error
	}
	aml, err := c.ComposeLibrary("Runtime", "runtime")
	if err != nil {
		return nil, nil, err
	}
	metadata := manifest{Version: 1, SHA256: fingerprint(aml), Sources: make(map[string]string)}
	for _, origin := range c.SourceOrigins {
		name := origin.Name
		if filepath.IsAbs(name) {
			name, err = filepath.Rel(root, name)
			if err != nil {
				return nil, nil, err
			}
		}
		name = filepath.ToSlash(name)
		if !fs.ValidPath(name) || !strings.HasSuffix(name, ".l") || !(strings.HasPrefix(name, "builtins/") || strings.HasPrefix(name, "stdlib/")) {
			return nil, nil, fmt.Errorf("composition source %q is outside canonical source resources", origin.Name)
		}
		current, err := fs.ReadFile(resources, name)
		if err != nil {
			return nil, nil, err
		}
		if fingerprint(current) != origin.SHA256 {
			return nil, nil, fmt.Errorf("composition source %s changed during generation", name)
		}
		metadata.Sources[name] = origin.SHA256
	}
	raw, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return nil, nil, err
	}
	return aml, append(raw, '\n'), nil
}

func fingerprint(raw []byte) string {
	digest := sha256.Sum256(raw)
	return hex.EncodeToString(digest[:])
}
