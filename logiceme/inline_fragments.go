package Logiceme

import (
	"bytes"

	"github.com/l-programming-language/l/lerrors"
)

func inlineConjoinmentNewLines(content []byte) []byte {
	return bytes.ReplaceAll(content, []byte("\n"), nil)
}

// Only operands embedded inside another source construct need exterior layout
// removed. Block and declaration collections retain their existing formatting.
func (g *Glyph) inlineConjoinment(child *Glyph) bool {
	switch glyphForm(g) {
	case "MV":
		return child.Type == "Instantiation"
	case "KY":
		return child.Type == "Operation"
	case "BI":
		return child.Type == "Calculation" || child.Type == "Definition"
	case "DO", "DOB":
		return glyphForm(child) == "CE"
	}
	return false
}

// Child content owns its metadata and import headers. Keep the headers separate
// until the caller has interpreted its enclosing template, so they cannot appear
// between a callable and its argument list or between a select guard and colon.
func prepareInlineConjoinment(child *Glyph) (source, headers []byte, err error) {
	source = scopedMetaInjection(child.Content, child.Meta)
	const prefix = `{|/|-{tags: ["Initialization"]}--{Headers}--{`
	const suffix = `}-|/|}`
	for {
		source = bytes.TrimSpace(source)
		if !bytes.HasPrefix(source, []byte(prefix)) {
			break
		}
		end := bytes.Index(source[len(prefix):], []byte(suffix))
		if end < 0 {
			return nil, nil, lerrors.InvalidConfiguration.Errorf("incomplete generated header in glyph %s", child.Name)
		}
		end += len(prefix) + len(suffix)
		headers = append(headers, source[:end]...)
		headers = append(headers, '\n')
		source = source[end:]
	}

	return source, headers, nil
}
