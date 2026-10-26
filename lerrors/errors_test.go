package lerrors

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"testing"
)

func TestKindNamesAndCompleteness(t *testing.T) {
	cases := []struct {
		kind Kind
		name string
	}{
		{Unknown, "unknown"},
		{InvalidInput, "invalid-input"},
		{InvalidConfiguration, "invalid-configuration"},
		{Syntax, "syntax"},
		{UnresolvedReference, "unresolved-reference"},
		{AmbiguousReference, "ambiguous-reference"},
		{DuplicateDeclaration, "duplicate-declaration"},
		{TypeMismatch, "type-mismatch"},
		{Cycle, "cycle"},
		{Unsupported, "unsupported"},
		{IO, "io"},
		{Tool, "tool"},
		{Internal, "internal"},
	}
	if Unknown != 0 || len(cases) != int(kindCount) {
		t.Fatalf("incomplete kind table: zero=%d, cases=%d, count=%d, names=%d", Unknown, len(cases), kindCount, len(kindNames))
	}
	seen := make(map[Kind]bool, len(cases))
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if seen[tc.kind] {
				t.Fatalf("duplicate category value %d", tc.kind)
			}
			seen[tc.kind] = true
			if !tc.kind.Valid() || tc.kind.String() != tc.name || tc.kind.Error() != tc.name || kindNames[tc.kind] != tc.name {
				t.Fatalf("kind %d: valid=%t, String=%q, Error=%q, indexed=%q; want %q", tc.kind, tc.kind.Valid(), tc.kind.String(), tc.kind.Error(), kindNames[tc.kind], tc.name)
			}
			var sentinel error = tc.kind
			if errors.Is(sentinel, nil) || !errors.Is(sentinel, tc.kind) {
				t.Fatal("a kind must be a non-nil, comparable error sentinel")
			}
		})
	}
	for kind := Kind(0); kind < kindCount; kind++ {
		if !seen[kind] {
			t.Errorf("category %d has no independently named test case", kind)
		}
	}
}

func TestKindInvalidValuesRemainDistinct(t *testing.T) {
	for _, kind := range []Kind{kindCount, 255} {
		t.Run(strconv.Itoa(int(kind)), func(t *testing.T) {
			want := "Kind(" + strconv.Itoa(int(kind)) + ")"
			if kind.Valid() || kind.String() != want || kind.Error() != want {
				t.Fatalf("invalid kind: valid=%t, String=%q, Error=%q; want %q", kind.Valid(), kind.String(), kind.Error(), want)
			}
			err := kind.Errorf("retained diagnostic")
			var diagnostic *Error
			if !errors.As(err, &diagnostic) || diagnostic.Kind() != kind || err.Error() != "retained diagnostic" || !errors.Is(err, kind) {
				t.Fatalf("constructor changed invalid kind or message: %#v", err)
			}
			for valid := Kind(0); valid < kindCount; valid++ {
				if errors.Is(err, valid) {
					t.Errorf("invalid kind %d matched category %s", kind, valid)
				}
			}
		})
	}
}

func TestErrorNilReceiver(t *testing.T) {
	var diagnostic *Error
	if diagnostic.Error() != "<nil>" || diagnostic.Kind() != Unknown || diagnostic.Unwrap() != nil {
		t.Fatalf("nil receiver: Error=%q, Kind=%s, Unwrap=%v", diagnostic.Error(), diagnostic.Kind(), diagnostic.Unwrap())
	}
	for _, target := range []error{nil, Unknown, Syntax, errors.New("unrelated"), diagnostic} {
		if diagnostic.Is(target) {
			t.Errorf("nil receiver matched target %#v", target)
		}
	}
}

func TestWrapNilReturnsNilInterface(t *testing.T) {
	for _, message := range []string{"", "discarded context"} {
		if err := IO.Wrap(nil, message); err != nil {
			t.Fatalf("Wrap(nil, %q) returned non-nil error interface %#v", message, err)
		}
	}
}

func TestWrapPreservesClassificationDisplayAndCause(t *testing.T) {
	cause := errors.New("native failure: 100% 雪\nsecond line")
	for _, tc := range []struct {
		name    string
		message string
		want    string
	}{
		{"inherited message", "", "native failure: 100% 雪\nsecond line"},
		{"complete replacement", "source.l:4:7: read failed", "source.l:4:7: read failed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := IO.Wrap(cause, tc.message)
			diagnostic, ok := err.(*Error)
			if !ok {
				t.Fatalf("Wrap returned %T; want *Error", err)
			}
			if err == cause || diagnostic.Kind() != IO || err.Error() != tc.want || diagnostic.Unwrap() != cause {
				t.Fatalf("Wrap lost classification, exact display, or cause: kind=%s, message=%q, cause=%#v", diagnostic.Kind(), err.Error(), diagnostic.Unwrap())
			}
			if !errors.Is(err, IO) || !errors.Is(err, cause) || errors.Is(err, Unknown) || errors.Is(err, Syntax) {
				t.Fatal("Wrap category or cause matching changed")
			}
		})
	}
}

func TestErrorfFormattingAndPercentText(t *testing.T) {
	message := "literal %w and %s: 100% 雪"
	cases := []struct {
		name string
		err  error
		want string
	}{
		{"formatted operands", InvalidInput.Errorf("operand %q at %d: %.1f%%", "雪", 3, 12.5), "operand \"雪\" at 3: 12.5%"},
		{"literal percent operand", Syntax.Errorf("%s", message), "literal %w and %s: 100% 雪"},
		{"empty diagnostic", Internal.Errorf(""), ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err == nil || tc.err.Error() != tc.want {
				t.Fatalf("formatted diagnostic = %v; want %q", tc.err, tc.want)
			}
			var diagnostic *Error
			if !errors.As(tc.err, &diagnostic) || diagnostic.Unwrap() == nil {
				t.Fatal("Errorf lost its formatted error cause")
			}
		})
	}
}

func TestErrorfPreservesSingleAndMultipleWrappedCauses(t *testing.T) {
	first := errors.New("first")
	second := errors.New("second")
	single := Syntax.Errorf("parse: %w", first)
	if single.Error() != "parse: first" || !errors.Is(single, first) || !errors.Is(single, Syntax) || errors.Is(single, second) {
		t.Fatalf("single wrapped cause lost: %v", single)
	}
	multiple := Tool.Errorf("run: %w; cleanup: %w", first, second)
	if multiple.Error() != "run: first; cleanup: second" || !errors.Is(multiple, first) || !errors.Is(multiple, second) || !errors.Is(multiple, Tool) {
		t.Fatalf("multiple wrapped causes lost: %v", multiple)
	}
	contextual := fmt.Errorf("compile: %w", multiple)
	if contextual.Error() != "compile: run: first; cleanup: second" || !errors.Is(contextual, Tool) || !errors.Is(contextual, first) || !errors.Is(contextual, second) {
		t.Fatalf("ordinary context wrapper changed typed error traversal: %v", contextual)
	}
	var original, throughContext *Error
	if !errors.As(multiple, &original) || !errors.As(contextual, &throughContext) || original != throughContext {
		t.Fatal("ordinary wrapper lost the original contextual error identity")
	}
}

func TestErrorIsIsShallowAndUnknownIsNotWildcard(t *testing.T) {
	leaf := errors.New("leaf")
	inner := Syntax.Wrap(leaf, "inner")
	outer, ok := IO.Wrap(inner, "outer").(*Error)
	if !ok {
		t.Fatal("wrapping a cause must return a contextual Error")
	}
	for _, target := range []error{Syntax, Unknown, leaf, inner, outer, nil, errors.New("outer")} {
		if outer.Is(target) {
			t.Errorf("shallow Is matched non-own-kind target %#v", target)
		}
	}
	if !outer.Is(IO) || !errors.Is(outer, IO) || !errors.Is(outer, Syntax) || !errors.Is(outer, leaf) || !errors.Is(outer, inner) || !errors.Is(outer, outer) {
		t.Fatal("shallow matching must leave cause traversal and pointer identity to errors.Is")
	}
	unknown := Unknown.Errorf("unclassified failure")
	if unknown == nil || !errors.Is(unknown, Unknown) || errors.Is(unknown, IO) || errors.Is(outer, Unknown) {
		t.Fatal("Unknown became success or a wildcard")
	}
}

func TestJoinedCausesKeepOuterCategoryAndExactDisplay(t *testing.T) {
	first := Syntax.Errorf("bad source")
	second := IO.Wrap(os.ErrNotExist, "missing input")
	joined := errors.Join(first, second)
	outer := Internal.Wrap(joined, "two alternatives failed")
	if outer.Error() != "two alternatives failed" || errors.Unwrap(outer) != joined {
		t.Fatalf("joined errors changed the selected display or cause: %v", outer)
	}
	for _, target := range []error{Internal, Syntax, IO, first, second, os.ErrNotExist} {
		if !errors.Is(outer, target) {
			t.Errorf("joined error lost target %#v", target)
		}
	}
	if errors.Is(outer, TypeMismatch) || errors.Is(outer, Unknown) {
		t.Fatal("joined error matched an unrelated category")
	}
	var diagnostic *Error
	if !errors.As(outer, &diagnostic) || diagnostic != outer || diagnostic.Kind() != Internal {
		t.Fatal("errors.As must select the outer contextual classification")
	}
	if !errors.As(joined, &diagnostic) || diagnostic != first || diagnostic.Kind() != Syntax {
		t.Fatal("joined causes must retain standard left-to-right traversal")
	}
}

func TestNativeStructuredCausesRemainInspectable(t *testing.T) {
	t.Run("filesystem", func(t *testing.T) {
		cause := &os.PathError{Op: "open", Path: "missing.l", Err: os.ErrNotExist}
		err := fmt.Errorf("initialize: %w", IO.Wrap(cause, "file missing.l does not exist"))
		var actual *os.PathError
		if !errors.As(err, &actual) || actual != cause || !errors.Is(err, os.ErrNotExist) || !errors.Is(err, IO) || err.Error() != "initialize: file missing.l does not exist" {
			t.Fatalf("filesystem cause or display changed: %v", err)
		}
	})
	t.Run("JSON", func(t *testing.T) {
		var target any
		cause := json.Unmarshal([]byte(`{"broken": }`), &target)
		var original *json.SyntaxError
		if !errors.As(cause, &original) {
			t.Fatalf("test did not produce a JSON syntax error: %v", cause)
		}
		err := InvalidConfiguration.Errorf("configuration: %w", cause)
		var actual *json.SyntaxError
		if !errors.As(err, &actual) || actual != original || !errors.Is(err, cause) || !errors.Is(err, InvalidConfiguration) {
			t.Fatalf("JSON cause lost: %v", err)
		}
	})
	t.Run("numeric", func(t *testing.T) {
		_, cause := strconv.ParseFloat("invalid-number", 64)
		var original *strconv.NumError
		if !errors.As(cause, &original) {
			t.Fatalf("test did not produce a numeric error: %v", cause)
		}
		err := InvalidInput.Errorf("Number: %w", cause)
		var actual *strconv.NumError
		if !errors.As(err, &actual) || actual != original || !errors.Is(err, strconv.ErrSyntax) || !errors.Is(err, InvalidInput) {
			t.Fatalf("numeric cause lost: %v", err)
		}
	})
	t.Run("subprocess", func(t *testing.T) {
		// Construct the native error directly: classification does not need to
		// launch a process or require an OS-specific ProcessState.
		cause := &exec.ExitError{Stderr: []byte("tool diagnostics")}
		err := Tool.Wrap(cause, "Go compilation failed")
		var actual *exec.ExitError
		if !errors.As(err, &actual) || actual != cause || !errors.Is(err, cause) || !errors.Is(err, Tool) || string(actual.Stderr) != "tool diagnostics" {
			t.Fatalf("subprocess cause lost: %v", err)
		}
	})
}

func TestInspectingAndWrappingDoesNotMutateError(t *testing.T) {
	cause := &os.PathError{Op: "read", Path: "source.l", Err: os.ErrPermission}
	original, ok := IO.Wrap(cause, "cannot read source.l").(*Error)
	if !ok {
		t.Fatal("wrapping a cause must return a contextual Error")
	}
	before := *original
	beforeCause := *cause
	for i := 0; i < 3; i++ {
		wrapped := fmt.Errorf("compile: %w", original)
		joined := errors.Join(wrapped, Syntax.Errorf("another failure"))
		var inspected *Error
		if !errors.As(joined, &inspected) || inspected != original || !errors.Is(joined, IO) || !errors.Is(joined, os.ErrPermission) {
			t.Fatal("inspection did not preserve original identity")
		}
		_ = original.Error()
		_ = original.Kind()                                 //nolint:errcheck // Exercise the category getter for its immutability, not an operational error.
		_ = original.Unwrap()                               //nolint:errcheck // Exercise cause inspection; the original error is deliberately retained below.
		_ = Internal.Wrap(original, "new outer diagnostic") //nolint:errcheck // Exercise wrapping to prove it does not mutate the existing error.
	}
	if *original != before || *cause != beforeCause {
		t.Fatal("inspection or wrapping mutated the original error or native cause")
	}
}
