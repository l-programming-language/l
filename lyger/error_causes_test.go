package Lyger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/l-programming-language/l/lerrors"
)

func TestDiagnosticContextPreservesCause(t *testing.T) {
	cause := errors.New("invalid operand 50%")
	t.Run("legacy statement", func(t *testing.T) {
		c := &Compilation{}
		if !c.checkLError(cause, []byte("source.l:2")) {
			t.Fatal("failure was not recorded")
		}
		if c.Error.Error() != "source.l:2 || invalid operand 50%" || !errors.Is(c.Error, cause) {
			t.Fatalf("lost message or cause: %v", c.Error)
		}
		first := c.Error
		c.checkLError(errors.New("later"), []byte("later"))
		if c.Error != first {
			t.Fatal("replaced first failure")
		}
	})
	t.Run("source scope", func(t *testing.T) {
		span := SourceSpan{Path: "imported.l", Start: SourcePosition{Line: 3, Column: 5}}
		err := span.wrapError(cause)
		if err.Error() != "imported.l:3:5: invalid operand 50%" || !errors.Is(err, cause) {
			t.Fatalf("lost message or cause: %v", err)
		}
	})
	t.Run("definition JSON", func(t *testing.T) {
		span := SourceSpan{Path: "library.l", Start: SourcePosition{Line: 3, Column: 2}}
		_, err := ParseDefinitionTemplate([]byte("[invalid]"), span)
		want := "library.l:3:2: definition: definition grouping: invalid character 'i' looking for beginning of value"
		var syntax *json.SyntaxError
		if err == nil || err.Error() != want || !errors.As(err, &syntax) || syntax.Offset != 2 {
			t.Fatalf("lost definition diagnostic or JSON cause: %v", err)
		}
	})
}

func TestGrammarAlternativeCauses(t *testing.T) {
	first, second := errors.New("first cause"), errors.New("second cause")
	state := 42
	err := selectGrammarCase(&state, map[string]string{"b": "second", "a": "first"},
		func(v int) int { return v },
		func(v *int, rule string) error {
			*v = 99
			if rule == "first" {
				return first
			}
			return second
		}, func(*int) []*token { return nil }, func(*int, string) {}, nil)
	want := "no grammar case matches (a: first cause; at end of statement; b: second cause; at end of statement)"
	if err == nil || err.Error() != want || !errors.Is(err, first) || !errors.Is(err, second) || state != 42 {
		t.Fatalf("lost grammar message, causes, or transactional state: state=%d err=%v", state, err)
	}
}

func TestEmptyGrammarAlternativesRemainAnError(t *testing.T) {
	state := 42
	err := selectGrammarCase(&state, nil, func(v int) int { return v },
		func(*int, string) error { t.Fatal("unexpected alternative"); return nil },
		func(*int) []*token { return nil }, func(*int, string) {}, nil)
	if err == nil || err.Error() != "no grammar case matches ()" || state != 42 {
		t.Fatalf("empty alternatives: state=%d err=%v", state, err)
	}
}

func TestNumericRangeErrorsReturn(t *testing.T) {
	const variable = "L_TEST_NUMERIC_RANGE"
	const completed = "numeric range error returned"
	number := strings.Repeat("9", 400)
	if mode := os.Getenv(variable); mode != "" {
		var err error
		switch mode {
		case "primitive":
			_, err = loadPrimitive([]byte(number))
		case "source":
			dir := t.TempDir()
			service := NewLygerService(canonicalTestOptions())
			_, err = service.Compile([]byte("\"quantity\" is "+number+"."), filepath.Join(dir, "overflow.l"))
			if err == nil || !strings.Contains(err.Error(), "overflow.l:1:1: tokenize:") {
				t.Fatalf("missing source attribution: %v", err)
			}
		default:
			t.Fatalf("unknown numeric test mode %q", mode)
		}
		var numeric *strconv.NumError
		if !errors.Is(err, lerrors.InvalidInput) || !errors.Is(err, strconv.ErrRange) || !errors.As(err, &numeric) || numeric.Num != number {
			t.Fatalf("numeric range cause lost: %v", err)
		}
		fmt.Println(completed)
		return
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"primitive", "source"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, executable, "-test.run=^TestNumericRangeErrorsReturn$")
			command.Env = append(os.Environ(), variable+"="+mode)
			var stdout, stderr bytes.Buffer
			command.Stdout, command.Stderr = &stdout, &stderr
			if err := command.Run(); err != nil || ctx.Err() != nil || !strings.HasPrefix(stdout.String(), completed+"\n") || stderr.Len() != 0 {
				t.Fatalf("numeric error did not return cleanly: %v\nstdout: %s\nstderr: %s", err, stdout.Bytes(), stderr.Bytes())
			}
		})
	}
}

func TestNumericProbeCompatibility(t *testing.T) {
	for _, raw := range []string{"42", "-0.125", "1e3"} {
		p, err := loadPrimitive([]byte(raw))
		want, parseErr := strconv.ParseFloat(raw, 64)
		if err != nil || parseErr != nil || p.Type != valueNumber || p.Number != want {
			t.Fatalf("Number %q changed: %v %v", raw, p, err)
		}
	}
	if _, err := loadPrimitive([]byte("pendingName")); err == nil || err.Error() != "newPrimitive unable to process input of pendingName" {
		t.Fatalf("ordinary probe changed: %v", err)
	}
}
