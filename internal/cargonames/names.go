// Package cargonames contains Cargo naming constraints shared by generated
// workspace projections and the native build coordinator.
package cargonames

import "strings"

// IsReservedBinary reports Cargo's reserved build-directory names. Case folding
// preserves the same generated layout on case-insensitive native filesystems.
func IsReservedBinary(name string) bool {
	switch strings.ToLower(name) {
	case "build", "deps", "examples", "incremental":
		return true
	default:
		return false
	}
}
