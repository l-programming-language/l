package distribution

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	data, err := os.ReadFile("VERSION")
	if err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\n$`).Match(data) {
		t.Fatalf("VERSION must contain bare major.minor.patch and one newline; got %q", data)
	}
	if got, want := Version(), strings.TrimSuffix(string(data), "\n"); got != want {
		t.Fatalf("embedded version = %q; want %q", got, want)
	}
}
