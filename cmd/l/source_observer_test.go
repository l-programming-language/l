//go:build darwin || linux

package main

import (
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	Logiceme "github.com/l-programming-language/l/logiceme"
)

func observeBundleSources(t *testing.T, b Logiceme.Executable) map[string]*nativetest.RustSource {
	t.Helper()
	return nativetest.ObserveRustBundle(t, b.ByteArrays, b.MapArrays)
}

func observeBundleMain(t *testing.T, b Logiceme.Executable) *nativetest.RustSource {
	t.Helper()
	return nativetest.ObserveRustMain(t, b.ByteArrays, b.MapArrays)
}

func observeCalls(tokens []nativetest.RustToken) []nativetest.RustCall {
	var out []nativetest.RustCall
	for _, call := range nativetest.RustCalls(tokens) {
		name := call.Name
		// Macro punctuation and control-flow grouping are not call names. In
		// particular println!(...) is one invocation, never a second `!` call.
		first, _ := utf8.DecodeRuneInString(name)
		if !(first == '_' || unicode.IsLetter(first)) || call.At > 0 && tokens[call.At-1].Text == "fn" {
			continue
		}
		switch name {
		case "if", "while", "match", "for", "loop", "return":
			continue
		}
		out = append(out, call)
	}
	return out
}

// Observe the exact source value independently of the selected Print definition's
// generated helper names or native implementation. Runtime observers cover output.
func observedText(t *testing.T, body []nativetest.RustToken, expected string) int {
	t.Helper()
	position := -1
	for i, token := range body {
		value, err := nativetest.DecodeRustLiteral(token.Text)
		if err != nil || value != expected {
			continue
		}
		if position >= 0 {
			t.Fatalf("source Text %q occurs more than once in the entry body", expected)
		}
		position = i
	}
	if position < 0 {
		t.Fatalf("entry body lost source Text %q", expected)
	}
	return position
}

func observedUnitOutput(tokens []nativetest.RustToken) bool {
	text := strings.ReplaceAll(nativetest.RustTokensText(tokens), " ", "")
	text = strings.TrimPrefix(text, "->")
	return text == "()" || text == "(())"
}
