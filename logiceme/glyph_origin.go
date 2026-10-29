package Logiceme

import (
	"fmt"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// Offsets are half-open UTF-8 byte ranges within the original syllable body.
// They describe source occurrences, not placeholder-expanded names or rendered
// text. The source string is immutable and shared by all origins in this tree.
type glyphOrigin struct {
	source     string
	start, end int
	form       string
	children   []*glyphOrigin
}

func attachGlyphOrigins(g *Glyph) error {
	source := g.Name
	position := 0
	origin, err := readGlyphOrigin(source, []byte(source), &position)
	if err != nil {
		return err
	}
	if position != len(source) {
		return lerrors.Syntax.Errorf("Rust composition has trailing source at byte %d", position)
	}
	var attach func(*Glyph, *glyphOrigin) error
	attach = func(node *Glyph, source *glyphOrigin) error {
		if glyphForm(node) != source.form || len(node.Conjoinments) != len(source.children) {
			return lerrors.Syntax.Errorf("Rust composition/source occurrence mismatch at bytes [%d:%d]", source.start, source.end)
		}
		node.origin = source
		// Restore real occurrence text for diagnostics and identifier hygiene;
		// expansion placeholders are never source identities.
		node.Name = source.source[source.start:source.end]
		for i := range node.Conjoinments {
			if err := attach(&node.Conjoinments[i], source.children[i]); err != nil {
				return err
			}
		}
		return nil
	}
	return attach(g, origin)
}

func readGlyphOrigin(source string, sourceBytes []byte, position *int) (*glyphOrigin, error) {
	start := *position
	for *position < len(source) && source[*position] >= 'A' && source[*position] <= 'Z' {
		*position++
	}
	if start == *position {
		return nil, lerrors.Syntax.Errorf("Rust composition requires a glyph at byte %d", start)
	}
	origin := &glyphOrigin{source: source, start: start, form: source[start:*position]}
	if strings.HasPrefix(source[*position:], "{|") {
		*position += 2
		for {
			if *position >= len(source) {
				return nil, lerrors.Syntax.Errorf("Rust composition metadata is unclosed at byte %d", start)
			}
			if strings.HasPrefix(source[*position:], "|}") {
				*position += 2
				break
			}
			if strings.HasPrefix(source[*position:], ": ~") {
				end, err := nativeValueEnd(sourceBytes, *position+3)
				if err != nil {
					return nil, err
				}
				*position = end + 1
			} else {
				*position++
			}
		}
	}
	if *position < len(source) && source[*position] == '^' {
		*position++
		group := *position < len(source) && source[*position] == '('
		if group {
			*position++
		}
		for {
			child, err := readGlyphOrigin(source, sourceBytes, position)
			if err != nil {
				return nil, err
			}
			origin.children = append(origin.children, child)
			if !group {
				break
			}
			if *position < len(source) && source[*position] == '*' {
				*position++
				continue
			}
			if *position >= len(source) || source[*position] != ')' {
				return nil, lerrors.Syntax.Errorf("Rust composition group is unclosed at byte %d", start)
			}
			*position++
			break
		}
	}
	origin.end = *position
	return origin, nil
}

type glyphOriginError struct {
	origin *glyphOrigin
	cause  error
}

func (e *glyphOriginError) Error() string {
	var location strings.Builder
	// strings.Builder writes never return errors.
	_, _ = fmt.Fprintf(&location, "glyph bytes [%d:%d]", e.origin.start, e.origin.end)
	for _, child := range e.origin.children {
		_, _ = fmt.Fprintf(&location, "; child bytes [%d:%d]", child.start, child.end)
	}
	return location.String() + ": " + e.cause.Error()
}

func (e *glyphOriginError) Unwrap() error { return e.cause }

func locateGlyphError(g *Glyph, err error) error {
	if child, ok := err.(*glyphChildError); ok && child.index > 0 && child.index <= len(g.Conjoinments) {
		child.cause = locateGlyphError(&g.Conjoinments[child.index-1], child.cause)
		return child
	}
	if g.origin != nil {
		return &glyphOriginError{origin: g.origin, cause: err}
	}
	return err
}
