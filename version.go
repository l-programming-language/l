package distribution

import (
	_ "embed"
	"strings"
)

//go:embed VERSION
var version string

// Version returns the release version embedded when the compiler was built.
func Version() string { return strings.TrimSpace(version) }
