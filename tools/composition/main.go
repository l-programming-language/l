// Command composition regenerates canonical support AML from source words.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	Lyger "github.com/l-programming-language/l/lyger"
)

func main() {
	os.Exit(command(os.Args[1:], os.Stderr))
}

func command(args []string, stderr io.Writer) int {
	flags := flag.NewFlagSet("composition", flag.ContinueOnError)
	flags.SetOutput(stderr)
	assets := flags.String("assets", ".", "complete asset provider to regenerate or check")
	check := flags.Bool("check", false, "verify the selected composition without writing")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(stderr, "composition accepts options only") //nolint:errcheck // Diagnostic output is best effort; the command returns failure.
		return 2
	}
	if err := run(*assets, *check); err != nil {
		fmt.Fprintln(stderr, err) //nolint:errcheck // Diagnostic output is best effort; the command returns failure.
		return 1
	}
	return 0
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
	}{{"stdlib/runtime/generated/support.aml", aml}, {"stdlib/runtime/generated/support.json", metadata}} {
		path := filepath.Join(root, filepath.FromSlash(output.name))
		if check {
			committed, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if !bytes.Equal(committed, output.data) {
				return fmt.Errorf("%s is stale; regenerate the selected provider with scripts/l-go run ./tools/composition -assets <directory>", output.name)
			}
		} else if err := os.WriteFile(path, output.data, 0644); err != nil {
			return err
		}
	}
	return nil
}
