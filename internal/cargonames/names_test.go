package cargonames

import (
	"strings"
	"testing"
)

func TestIsReservedBinary(t *testing.T) {
	for _, name := range []string{"build", "deps", "examples", "incremental"} {
		for _, spelling := range []string{name, strings.ToUpper(name), strings.ToUpper(name[:1]) + name[1:]} {
			if !IsReservedBinary(spelling) {
				t.Errorf("Cargo build-directory name %q must be reserved", spelling)
			}
		}
	}
	for _, name := range []string{"", "Worker", "build_app", "deps2", "l_app", "l_entry"} {
		if IsReservedBinary(name) {
			t.Errorf("ordinary binary name %q must remain available", name)
		}
	}
}
