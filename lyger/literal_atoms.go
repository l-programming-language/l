package Lyger

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"
)

// quoteNativeString protects native literals from AML/Lucid delimiters without changing
// their decoded Rust contents. Unlike Go's Quote, Rust uses braced Unicode
// escapes and does not permit arbitrary high-byte escapes in string literals.
func quoteNativeString(value string) string {
	var out strings.Builder
	out.WriteByte('"')
	for _, char := range value {
		switch char {
		case '"':
			out.WriteString(`\"`)
		case '\\':
			out.WriteString(`\\`)
		case '\n':
			out.WriteString(`\n`)
		case '\r':
			out.WriteString(`\r`)
		case '\t':
			out.WriteString(`\t`)
		default:
			if char < 32 || char == 127 {
				// strings.Builder.Write always returns a nil error.
				_, _ = fmt.Fprintf(&out, `\x%02x`, char)
			} else {
				out.WriteRune(char)
			}
		}
	}
	out.WriteByte('"')
	return logemeTextEscaper.Replace(out.String())
}

// Scalar Text may carry arbitrary bytes. Keep valid UTF-8's established native
// string spelling; otherwise expose a byte string rather than replacing invalid
// octets or inventing a UTF-8 conversion. Its native consumer must accept bytes.
func definitionTextLiteral(value string) string {
	if utf8.ValidString(value) {
		return quoteNativeString(value)
	}
	var out strings.Builder
	out.WriteString(`b"`)
	for index := 0; index < len(value); index++ {
		// strings.Builder.Write always returns a nil error.
		_, _ = fmt.Fprintf(&out, `\x%02x`, value[index])
	}
	out.WriteByte('"')
	return out.String()
}

func nativeNumberLiteral(number float64) (string, bool) {
	if math.IsNaN(number) || math.IsInf(number, 0) {
		return "", false
	}
	return strconv.FormatFloat(number, 'g', -1, 64) + "_f64", true
}
