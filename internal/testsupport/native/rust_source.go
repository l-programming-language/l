package native

// These bounded observers inspect actual returned Rust. They deliberately do
// not translate Rust into Go AST, render from IR, or substitute expected source.
// rustc remains the native grammar/type checker in build witnesses.
import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"
)

type RustToken struct {
	Text       string
	start, end int
}
type RustSource struct {
	Source string
	Tokens []RustToken
	pairs  map[int]int
}
type RustFunction struct {
	Name                     string
	Parameters, Result, Body []RustToken
}
type RustCall struct {
	Receiver, Name string
	Args           [][]RustToken
	At, End        int
}

func ObserveRustSource(t *testing.T, source []byte) *RustSource {
	t.Helper()
	f := &RustSource{Source: string(source), pairs: map[int]int{}}
	s := f.Source
	for i := 0; i < len(s); {
		c := s[i]
		if c == ' ' || c == '\t' || c == '\r' || c == '\n' {
			i++
			continue
		}
		if strings.HasPrefix(s[i:], "//") {
			j := strings.IndexByte(s[i:], '\n')
			if j < 0 {
				break
			}
			i += j + 1
			continue
		}
		if strings.HasPrefix(s[i:], "/*") {
			depth := 1
			i += 2
			for i < len(s) && depth > 0 {
				if strings.HasPrefix(s[i:], "/*") {
					depth++
					i += 2
				} else if strings.HasPrefix(s[i:], "*/") {
					depth--
					i += 2
				} else {
					i++
				}
			}
			if depth != 0 {
				t.Fatal("unterminated Rust block comment")
			}
			continue
		}
		start := i
		// Quoted and raw strings remain one token, including byte-string prefixes.
		prefix := i
		if c == 'b' && i+1 < len(s) && (s[i+1] == '"' || s[i+1] == 'r') {
			prefix++
		}
		if s[prefix] == 'r' {
			q := prefix + 1
			for q < len(s) && s[q] == '#' {
				q++
			}
			if q < len(s) && s[q] == '"' {
				end := "\"" + s[prefix+1:q]
				tail := strings.Index(s[q+1:], end)
				if tail < 0 {
					t.Fatal("unterminated Rust raw string")
				}
				i = q + 1 + tail + len(end)
				f.Tokens = append(f.Tokens, RustToken{s[start:i], start, i})
				continue
			}
		}
		if s[prefix] == '"' {
			i = prefix + 1
			for i < len(s) {
				if s[i] == '\\' {
					i += 2
					continue
				}
				if s[i] == '"' {
					i++
					break
				}
				i++
			}
			if i > len(s) || s[i-1] != '"' {
				t.Fatal("unterminated Rust string")
			}
			f.Tokens = append(f.Tokens, RustToken{s[start:i], start, i})
			continue
		}
		r, n := utf8.DecodeRuneInString(s[i:])
		if unicode.IsLetter(r) || c == '_' {
			i += n
			for i < len(s) {
				r, n = utf8.DecodeRuneInString(s[i:])
				if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
					break
				}
				i += n
			}
		} else if c >= '0' && c <= '9' {
			i++
			for i < len(s) && ((s[i] >= '0' && s[i] <= '9') || s[i] == '.' || s[i] == '_') {
				i++
			}
		} else {
			i++
			if i < len(s) {
				op := s[start : i+1]
				if op == "::" || op == "->" || op == "=>" || op == "==" || op == "!=" || op == "&&" || op == "||" {
					i++
				}
			}
		}
		f.Tokens = append(f.Tokens, RustToken{s[start:i], start, i})
	}
	var stack []int
	for i, tok := range f.Tokens {
		switch tok.Text {
		case "(", "[", "{":
			stack = append(stack, i)
		case ")", "]", "}":
			if len(stack) == 0 {
				t.Fatalf("unpaired Rust delimiter %s", tok.Text)
			}
			j := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if map[string]string{"(": ")", "[": "]", "{": "}"}[f.Tokens[j].Text] != tok.Text {
				t.Fatal("mismatched Rust delimiters")
			}
			f.pairs[i] = j
			f.pairs[j] = i
		}
	}
	if len(stack) != 0 {
		t.Fatal("unclosed Rust delimiter")
	}
	return f
}
func RustTokensText(tokens []RustToken) string {
	var b strings.Builder
	for _, t := range tokens {
		b.WriteString(t.Text)
	}
	return b.String()
}
func (f *RustSource) Functions() []RustFunction {
	var out []RustFunction
	for i := 0; i+2 < len(f.Tokens); i++ {
		if f.Tokens[i].Text != "fn" {
			continue
		}
		name := f.Tokens[i+1].Text
		j := i + 2
		for j < len(f.Tokens) && f.Tokens[j].Text != "(" && f.Tokens[j].Text != "{" {
			j++
		}
		close, ok := f.pairs[j]
		if !ok || f.Tokens[j].Text != "(" {
			continue
		}
		k := close + 1
		for k < len(f.Tokens) && f.Tokens[k].Text != "{" && f.Tokens[k].Text != ";" {
			k++
		}
		if k == len(f.Tokens) || f.Tokens[k].Text != "{" {
			continue
		}
		end := f.pairs[k]
		out = append(out, RustFunction{name, f.Tokens[j+1 : close], f.Tokens[close+1 : k], f.Tokens[k+1 : end]})
		i = end
	}
	return out
}
func (f *RustSource) Function(t *testing.T, name string) RustFunction {
	t.Helper()
	var found []RustFunction
	for _, fn := range f.Functions() {
		if fn.Name == name {
			found = append(found, fn)
		}
	}
	if len(found) != 1 {
		t.Fatalf("actual Rust function %s count=%d; want one", name, len(found))
	}
	return found[0]
}
func RustTokenArgs(tokens []RustToken) [][]RustToken {
	var out [][]RustToken
	depth, start := 0, 0
	for i, t := range tokens {
		switch t.Text {
		case "(", "[", "{":
			depth++
		case ")", "]", "}":
			depth--
		case ",":
			if depth == 0 {
				out = append(out, tokens[start:i])
				start = i + 1
			}
		}
	}
	if start < len(tokens) {
		out = append(out, tokens[start:])
	}
	return out
}
func RustCalls(tokens []RustToken) []RustCall {
	var out []RustCall
	for i := 0; i+1 < len(tokens); i++ {
		name := tokens[i].Text
		if name == "!" { // Macro punctuation is part of its preceding named call.
			continue
		}
		j := i + 1
		if tokens[j].Text == "!" {
			j++
		}
		if j >= len(tokens) || tokens[j].Text != "(" {
			continue
		}
		depth, end := 1, j+1
		for end < len(tokens) && depth > 0 {
			if tokens[end].Text == "(" {
				depth++
			}
			if tokens[end].Text == ")" {
				depth--
			}
			end++
		}
		if depth != 0 {
			continue
		}
		receiver := ""
		if i >= 2 && tokens[i-1].Text == "." {
			receiver = tokens[i-2].Text
		}
		out = append(out, RustCall{receiver, name, RustTokenArgs(tokens[j+1 : end-1]), i, end})
	}
	return out
}
func DecodeRustLiteral(token string) (string, error) {
	s := token
	if strings.HasPrefix(s, "b\"") {
		s = s[1:]
	}
	if strings.HasPrefix(s, "br") {
		s = s[1:]
	}
	if strings.HasPrefix(s, "r") {
		q := strings.IndexByte(s, '"')
		if q < 0 {
			return "", fmt.Errorf("invalid raw Rust string %s", token)
		}
		hash := s[1:q]
		end := "\"" + hash
		if !strings.HasSuffix(s, end) {
			return "", fmt.Errorf("invalid raw string")
		}
		return s[q+1 : len(s)-len(end)], nil
	}
	var out strings.Builder
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return "", fmt.Errorf("not a Rust string: %s", token)
	}
	for i := 1; i < len(s)-1; i++ {
		if s[i] != '\\' {
			out.WriteByte(s[i])
			continue
		}
		i++
		switch s[i] {
		case 'n':
			out.WriteByte('\n')
		case 'r':
			out.WriteByte('\r')
		case 't':
			out.WriteByte('\t')
		case '0':
			out.WriteByte(0)
		case '\\', '"', '\'':
			out.WriteByte(s[i])
		case 'x':
			if i+2 >= len(s) {
				return "", fmt.Errorf("short hex escape")
			}
			v, e := strconv.ParseUint(s[i+1:i+3], 16, 8)
			if e != nil {
				return "", e
			}
			out.WriteByte(byte(v))
			i += 2
		case 'u':
			if i+1 >= len(s) || s[i+1] != '{' {
				return "", fmt.Errorf("Rust unicode escape needs braces")
			}
			j := strings.IndexByte(s[i+2:], '}')
			if j < 0 {
				return "", fmt.Errorf("unterminated unicode escape")
			}
			v, e := strconv.ParseUint(strings.ReplaceAll(s[i+2:i+2+j], "_", ""), 16, 32)
			if e != nil {
				return "", e
			}
			out.WriteRune(rune(v))
			i += j + 2
		default:
			return "", fmt.Errorf("unsupported Rust escape \\%c", s[i])
		}
	}
	return out.String(), nil
}
func ObserveRustBundle(t *testing.T, contents [][][]byte, paths [][]string) map[string]*RustSource {
	t.Helper()
	if len(contents) != 1 || len(paths) != 1 || len(contents[0]) != len(paths[0]) {
		t.Fatal("unpaired Rust bundle")
	}
	out := map[string]*RustSource{}
	seen := map[string]bool{}
	for i, path := range paths[0] {
		if seen[path] {
			t.Fatalf("duplicate generated path %s", path)
		}
		seen[path] = true
		if strings.HasSuffix(path, ".go") {
			t.Errorf("Go artifact remains in Rust bundle: %s", path)
		}
		if strings.HasSuffix(path, ".rs") {
			out[path] = ObserveRustSource(t, contents[0][i])
		}
	}
	if len(out) == 0 {
		t.Fatal("actual generated bundle contains no Rust source")
	}
	return out
}
func ObserveRustMain(t *testing.T, contents [][][]byte, paths [][]string) *RustSource {
	t.Helper()
	var found *RustSource
	for p, f := range ObserveRustBundle(t, contents, paths) {
		if filepath.Base(p) == "main.rs" {
			if found != nil {
				t.Fatal("duplicate main.rs")
			}
			found = f
		}
	}
	if found == nil {
		t.Fatal("actual Rust bundle has no main.rs")
	}
	return found
}
