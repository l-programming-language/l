// Package sourcepaths defines the distribution's source-resource boundaries.
// It performs no filesystem access or resource resolution.
package sourcepaths

import (
	"io/fs"
	"strings"
)

// Canonical reports whether name can identify a distributed L source file.
// A registry language may deliberately share another language's source files.
func Canonical(name string) bool {
	if !fs.ValidPath(name) || strings.Contains(name, "\\") || !strings.HasSuffix(name, ".l") {
		return false
	}
	parts := strings.Split(name, "/")
	if len(parts) < 3 {
		return false
	}
	switch parts[0] {
	case "builtins":
		switch parts[1] {
		case "runtime", "segments", "templates", "operations":
			return false
		}
		return len(parts) == 3 && (parts[2] == "core.l" || parts[2] == "prelude.l") || len(parts) >= 4 && parts[2] == "storage"
	case "runtime":
		return parts[1] != "rust" && parts[1] != "generated"
	case "stdlib":
		return true
	}
	return false
}
