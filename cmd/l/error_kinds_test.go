package main

import (
	"errors"
	"flag"

	"strings"
	"testing"

	"github.com/l-programming-language/l/lerrors"
)

func TestCommandErrorKindsPreserveControlSentinels(t *testing.T) {
	for _, args := range [][]string{nil, {"unknown"}, {"build", "--unknown"}, {"build", "-os", "linux"}, {"build", "one.l", "two.l"}} {
		t.Run("invalid "+strings.Join(args, " "), func(t *testing.T) {
			_, err := parseCommand(args)
			if !errors.Is(err, lerrors.InvalidInput) {
				t.Fatalf("invocation error = %v", err)
			}
		})
	}
	for _, args := range [][]string{{"--help"}, {"build", "--help"}} {
		_, err := parseCommand(args)
		if err != flag.ErrHelp {
			t.Fatalf("help sentinel changed: %v", err)
		}
	}
	_, err := nativeBuildOptions("", "unsupported-host", "unknown")
	if !errors.Is(err, errUnsupportedBuildHost) || !errors.Is(err, lerrors.Unsupported) {
		t.Fatalf("host error lost its identity/category: %v", err)
	}
}
