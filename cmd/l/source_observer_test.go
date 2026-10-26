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

func observedPrinted(t *testing.T, body []nativetest.RustToken) []string {
	t.Helper()
	var values []string
	dispatches := 0
	fallbacks := 0
	for _, call := range observeCalls(body) {
		if call.Name == "println" {
			fallbacks++
			if len(call.Args) != 2 || nativetest.RustTokensText(call.Args[0]) != `"{}"` || nativetest.RustTokensText(call.Args[1]) != "self.0" {
				t.Fatal("IO Print added a print outside its typed fallback")
			}
		}
		if call.Name == "l_print" {
			dispatches++
			if len(call.Args) != 0 {
				t.Fatal("IO Print dispatch must not add arguments")
			}
		}
		if call.Name != "LPrintValue" {
			continue
		}
		if len(call.Args) != 1 || len(call.Args[0]) != 4 {
			t.Fatal("IO Print must borrow exactly one parenthesized native Text literal")
		}
		arg := call.Args[0]
		if arg[0].Text != "&" || arg[1].Text != "(" || arg[3].Text != ")" {
			t.Fatal("IO Print changed the borrowed Text operand")
		}
		text, e := nativetest.DecodeRustLiteral(arg[2].Text)
		if e != nil {
			t.Fatal(e)
		}
		want := "(&&LPrintValue(&(" + arg[2].Text + "))).l_print()"
		if !strings.Contains(nativetest.RustTokensText(body), want) {
			t.Fatal("IO Print literal is not the selected dispatch receiver")
		}
		values = append(values, text)
	}
	if dispatches != len(values) || fallbacks != dispatches {
		t.Fatal("IO Print changed dispatch or payload evaluation count")
	}
	return values
}

func observedUnitOutput(tokens []nativetest.RustToken) bool {
	text := strings.ReplaceAll(nativetest.RustTokensText(tokens), " ", "")
	text = strings.TrimPrefix(text, "->")
	return text == "()" || text == "(())"
}
