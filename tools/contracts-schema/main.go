// Command contracts-schema emits a JSON Schema derived from the shared model.
package main

import (
	"fmt"
	"os"

	"github.com/l-programming-language/l/contracts"
)

func main() {
	if len(os.Args) != 2 {
		if _, err := fmt.Fprintln(os.Stderr, "usage: contracts-schema DOCUMENT_KIND"); err != nil {
			os.Exit(1)
		}
		os.Exit(2)
	}
	data, err := contracts.JSONSchema(os.Args[1])
	if err != nil {
		if _, writeErr := fmt.Fprintln(os.Stderr, err); writeErr != nil {
			os.Exit(1)
		}
		os.Exit(2)
	}
	if _, err := os.Stdout.Write(data); err != nil {
		os.Exit(1)
	}
}
