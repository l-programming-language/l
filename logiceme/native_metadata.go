package Logiceme

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// readGlyphInputs reads existing tilde frames, never punctuation inside a value.
// The closing frame is recognized only before the next metadata key or EOF.
func readGlyphInputs(s string) (map[string]string, error) {
	out := map[string]string{}
	for strings.TrimSpace(s) != "" {
		colon := strings.Index(s, ": ~")
		if colon < 1 {
			return nil, fmt.Errorf("invalid metadata frame %q", s)
		}
		key := strings.TrimSpace(s[:colon])
		if strings.ContainsAny(key, " ,{}\n") {
			return nil, fmt.Errorf("invalid metadata key %q", key)
		}
		s = s[colon+3:]
		end := -1
		for i := 0; i < len(s); i++ {
			if s[i] != '~' {
				continue
			}
			tail := s[i+1:]
			if tail == "" {
				end = i
				break
			}
			if strings.HasPrefix(tail, ", ") {
				next := tail[2:]
				p := strings.Index(next, ": ~")
				if p > 0 && !strings.ContainsAny(next[:p], " ,{}\n~") {
					end = i
					break
				}
			}
		}
		if end < 0 {
			return nil, fmt.Errorf("missing closing metadata frame for %s", key)
		}
		if _, ok := out[key]; ok {
			return nil, fmt.Errorf("duplicate metadata field %s", key)
		}
		out[key] = s[:end]
		s = s[end+1:]
		s = strings.TrimPrefix(s, ", ")
	}
	return out, nil
}

// splitNativeFields splits only top-level commas. It is a lexical metadata reader, not
// a Rust type checker: quotes, raw strings, lifetimes, and balanced delimiters
// remain opaque. Angle nesting is enabled for type/generic syntax.
func splitNativeFields(s string) []string {
	var out []string
	start := 0
	var stack []byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '"' || (c == '\'' && i+2 < len(s) && (s[i+2] == '\'' || s[i+1] == '\\')) {
			quote := c
			i++
			for i < len(s) {
				if s[i] == '\\' {
					i += 2
					continue
				}
				if s[i] == quote {
					break
				}
				i++
			}
			continue
		}
		if c == 'r' {
			j := i + 1
			for j < len(s) && s[j] == '#' {
				j++
			}
			if j < len(s) && s[j] == '"' {
				end := `"` + s[i+1:j]
				if k := strings.Index(s[j+1:], end); k >= 0 {
					i = j + 1 + k + len(end) - 1
					continue
				}
			}
		}
		// A return arrow is one native token, not a generic closing delimiter.
		if strings.HasPrefix(s[i:], "->") {
			i++
			continue
		}
		switch c {
		case '(', '[', '{':
			stack = append(stack, c)
		case '<':
			if i > 0 && (unicode.IsLetter(rune(s[i-1])) || s[i-1] == ':' || s[i-1] == '>') {
				stack = append(stack, c)
			}
		case ')', ']', '}':
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		case '>':
			if len(stack) > 0 && stack[len(stack)-1] == '<' {
				stack = stack[:len(stack)-1]
			}
		case ',':
			if len(stack) == 0 {
				out = append(out, strings.TrimSpace(s[start:i]))
				start = i + 1
			}
		}
	}
	if strings.TrimSpace(s[start:]) != "" {
		out = append(out, strings.TrimSpace(s[start:]))
	}
	return out
}
func unframeNativeList(s string) string {
	if len(s) >= 4 && ((strings.HasPrefix(s, "&[") && strings.HasSuffix(s, "]&")) || (strings.HasPrefix(s, "#[") && strings.HasSuffix(s, "]#"))) {
		return s[2 : len(s)-2]
	}
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		return s[1 : len(s)-1]
	}
	return s
}
func readNativeList(s string) []string { return splitNativeFields(unframeNativeList(s)) }
func readNativeRecord(s, kind string) (map[string]string, error) {
	s = unframeNativeList(strings.TrimSpace(s))
	prefix, suffix := kind+"{", "}"+kind
	if !strings.HasPrefix(s, prefix) || !strings.HasSuffix(s, suffix) {
		return nil, fmt.Errorf("expected %s record", kind)
	}
	out := map[string]string{}
	for _, f := range splitNativeFields(s[len(prefix) : len(s)-len(suffix)]) {
		k, v, ok := strings.Cut(f, ": ")
		if !ok {
			return nil, fmt.Errorf("invalid %s field %q", kind, f)
		}
		if _, yes := out[k]; yes {
			return nil, fmt.Errorf("duplicate %s field %s", kind, k)
		}
		out[k] = v
	}
	return out, nil
}
func readNativeProps(s string) ([]Prop, error) {
	var out []Prop
	for _, r := range readNativeList(s) {
		m, e := readNativeRecord(r, "Prop")
		if e != nil {
			return nil, e
		}
		out = append(out, Prop{Key: m["Key"], Type: m["Type"], Value: m["Value"], StructTags: m["StructTags"]})
	}
	return out, nil
}

// readNativeOps reads record boundaries independently of scalar fragments. Prefix and
// Postfix may each contain unmatched punctuation; only Operands is a structured
// list. Balancing an entire record would silently absorb subsequent fields.
func readNativeOps(s string) ([]Op, error) {
	s = strings.TrimSpace(unframeNativeList(strings.TrimSpace(s)))
	var out []Op
	for s != "" {
		if !strings.HasPrefix(s, "Op{") {
			return nil, fmt.Errorf("expected Op record")
		}
		s = s[3:]
		fields := map[string]string{}
		for {
			key, rest, ok := strings.Cut(s, ": ")
			key = strings.TrimSpace(key)
			if !ok || !isNativeMetadataKey(key) {
				return nil, fmt.Errorf("invalid Op field %q", s)
			}
			if _, exists := fields[key]; exists {
				return nil, fmt.Errorf("duplicate Op field %s", key)
			}
			end := nativeOpFieldEnd(rest, key == "Operands")
			if end < 0 {
				return nil, fmt.Errorf("missing closing Op record")
			}
			fields[key] = strings.TrimSpace(rest[:end])
			s = rest[end:]
			if strings.HasPrefix(s, "}Op") {
				s = strings.TrimSpace(s[3:])
				break
			}
			s = strings.TrimSpace(s[1:])
		}
		out = append(out, Op{Type: fields["Type"], Prefix: unframeNativeList(fields["Prefix"]), Postfix: unframeNativeList(fields["Postfix"]), Operator: fields["Operator"], Operands: readNativeList(fields["Operands"])})
		if s != "" {
			if s[0] != ',' {
				return nil, fmt.Errorf("expected comma after Op record")
			}
			s = strings.TrimSpace(s[1:])
		}
	}
	return out, nil
}

func isNativeMetadataKey(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return false
		}
	}
	return true
}

func nativeOpFieldEnd(s string, structured bool) int {
	var stack []byte
	for i := 0; i < len(s); i++ {
		if end := nativeMetadataLiteralEnd(s, i); end > i {
			i = end - 1
			continue
		}
		// Keep this boundary scan consistent with the native list reader.
		if strings.HasPrefix(s[i:], "->") {
			i++
			continue
		}
		if len(stack) == 0 {
			if strings.HasPrefix(s[i:], "}Op") {
				return i
			}
			if s[i] == ',' {
				next := strings.TrimSpace(s[i+1:])
				if key, _, ok := strings.Cut(next, ": "); ok && isNativeMetadataKey(key) {
					return i
				}
			}
		}
		if !structured {
			continue
		}
		switch s[i] {
		case '(', '[', '{':
			stack = append(stack, s[i])
		case '<':
			if i > 0 && (unicode.IsLetter(rune(s[i-1])) || s[i-1] == ':' || s[i-1] == '>') {
				stack = append(stack, '<')
			}
		case ')', ']', '}':
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		case '>':
			if len(stack) > 0 && stack[len(stack)-1] == '<' {
				stack = stack[:len(stack)-1]
			}
		}
	}
	return -1
}

// Return the end of a quoted native atom so marker-like contents remain data.
// A lifetime apostrophe is not a character literal.
func nativeMetadataLiteralEnd(s string, start int) int {
	if s[start] == 'r' {
		j := start + 1
		for j < len(s) && s[j] == '#' {
			j++
		}
		if j < len(s) && s[j] == '"' {
			closingDelimiter := `"` + s[start+1:j]
			if end := strings.Index(s[j+1:], closingDelimiter); end >= 0 {
				return j + 1 + end + len(closingDelimiter)
			}
			return len(s)
		}
	}
	quote := s[start]
	if quote != '"' && quote != '\'' {
		return start
	}
	if quote == '\'' {
		if start+1 >= len(s) {
			return start
		}
		_, width := utf8.DecodeRuneInString(s[start+1:])
		if s[start+1] != '\\' && (start+1+width >= len(s) || s[start+1+width] != '\'') {
			return start
		}
	}
	for j := start + 1; j < len(s); j++ {
		if s[j] == '\\' {
			j++
			continue
		}
		if s[j] == quote {
			return j + 1
		}
	}
	return len(s)
}

// readNativeFuncs keeps each existing Func record and its nested native types intact.
// A malformed record invalidates the collection; never publish a partial list.
func readNativeFuncs(raw string) ([]Func, error) {
	var functions []Func
	for _, record := range readNativeList(raw) {
		function, err := readNativeFunc(map[string]string{"Func": record})
		if err != nil {
			return nil, err
		}
		functions = append(functions, function)
	}
	return functions, nil
}

func readNativeFunc(m map[string]string) (Func, error) {
	if raw := m["Func"]; raw != "" {
		var e error
		m, e = readNativeRecord(raw, "Func")
		if e != nil {
			return Func{}, e
		}
	}
	f := Func{Key: m["Key"], Outputs: readNativeList(m["Outputs"])}
	var e error
	f.Inputs, e = readNativeProps(m["Inputs"])
	if e != nil {
		return f, e
	}
	if m["Receiver"] != "" {
		p, err := readNativeProps(m["Receiver"])
		if err != nil {
			return f, err
		}
		if len(p) != 1 {
			return f, fmt.Errorf("Receiver requires one Prop")
		}
		f.Receiver = p[0]
	}
	return f, nil
}
