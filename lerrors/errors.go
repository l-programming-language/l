// Package lerrors classifies compiler and build failures while preserving
// their diagnostic text and underlying causes.
package lerrors

import "fmt"

// Kind identifies why a compiler operation failed. Its numeric values are not
// a wire format. Unknown is a category, never a successful result.
type Kind uint8

const (
	Unknown Kind = iota
	InvalidInput
	InvalidConfiguration
	Syntax
	UnresolvedReference
	AmbiguousReference
	DuplicateDeclaration
	TypeMismatch
	Cycle
	Unsupported
	IO
	Tool
	Internal

	kindCount
)

var kindNames = [kindCount]string{
	Unknown:              "unknown",
	InvalidInput:         "invalid-input",
	InvalidConfiguration: "invalid-configuration",
	Syntax:               "syntax",
	UnresolvedReference:  "unresolved-reference",
	AmbiguousReference:   "ambiguous-reference",
	DuplicateDeclaration: "duplicate-declaration",
	TypeMismatch:         "type-mismatch",
	Cycle:                "cycle",
	Unsupported:          "unsupported",
	IO:                   "io",
	Tool:                 "tool",
	Internal:             "internal",
}

// Valid reports whether k is a defined category, including Unknown.
func (k Kind) Valid() bool { return k < kindCount }

func (k Kind) String() string {
	if !k.Valid() {
		return fmt.Sprintf("Kind(%d)", uint8(k))
	}
	return kindNames[k]
}

// Error lets a Kind be an errors.Is target.
func (k Kind) Error() string { return k.String() }

// Error retains classification separately from its displayed diagnostic.
type Error struct {
	kind    Kind
	message string
	cause   error
}

func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	return e.message
}

func (e *Error) Kind() Kind {
	if e == nil {
		return Unknown
	}
	return e.kind
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

// Is matches this error's category; errors.Is handles traversal of its causes.
func (e *Error) Is(target error) bool {
	kind, ok := target.(Kind)
	return e != nil && ok && e.kind == kind
}

// Errorf uses fmt.Errorf semantics, including preservation of every %w cause.
func (k Kind) Errorf(format string, args ...any) error {
	return k.Wrap(fmt.Errorf(format, args...), "")
}

// Wrap classifies cause. A nonempty message is the complete diagnostic; an
// empty message preserves cause.Error(). A nil cause returns a nil interface.
func (k Kind) Wrap(cause error, message string) error {
	if cause == nil {
		return nil
	}
	if message == "" {
		message = cause.Error()
	}
	return &Error{kind: k, message: message, cause: cause}
}

var (
	_ error = Unknown
	_ error = (*Error)(nil) //nolint:errcheck // Compile-time interface conformance assertion, not a returned operational error.
)
