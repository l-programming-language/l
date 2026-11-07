// Command composition regenerates canonical support AML from source words.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	Lyger "github.com/l-programming-language/l/lyger"
)

func main() {
	check := flag.Bool("check", false, "verify the committed composition without writing")
	flag.Parse()
	if err := run(".", *check); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(root string, check bool) error {
	aml, manifest, err := Lyger.ComposeRuntimeResources(os.DirFS(root))
	if err != nil {
		return err
	}
	metadata, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	metadata = append(metadata, '\n')
	for _, output := range []struct {
		name string
		data []byte
	}{{"runtime/generated/support.aml", aml}, {"runtime/generated/support.json", metadata}} {
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
