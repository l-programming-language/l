package Lyger

import (
	"bytes"
	"errors"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/l-programming-language/l/lerrors"
)

func TestAuthoredDotNotationTokensReject(t *testing.T) {
	gram := scopeTestGrammar(t)
	for _, source := range []string{
		"Show Tools.Value",
		"Tools.Write `text`",
		"Show Tools.Inner.Value, and Other",
		"Show Tools.Inner's Value",
		"Show Tools's Inner.Value",
		`Create "Tools.Value" as a Dictionary`,
		`Use Tools as "Other.Value"`,
		`Define "Word|-s.other" as Text`,
		`Create "Tools\u002eValue" as a Dictionary`,
		`Create "Tools\x2eValue" as a Dictionary`,
	} {
		t.Run(source, func(t *testing.T) {
			statement, err := newLStatement([]byte(source), gram)
			if err == nil {
				err = statement.tokenize()
			}
			if !errors.Is(err, lerrors.Syntax) || !strings.Contains(err.Error(), "dot-qualified") || !strings.Contains(err.Error(), "possessive") {
				t.Fatalf("authored dotted name must fail during tokenization with grammatical guidance: %v", err)
			}
		})
	}
}

func TestAuthoredDotNotationKeepsSourceAttribution(t *testing.T) {
	for _, source := range []string{
		"Show Tools.Value.",
		`Create "Tools.Value" as a Dictionary.`,
		`Create "Tools\u002eValue" as a Dictionary.`,
	} {
		t.Run(source, func(t *testing.T) {
			original := []byte("*** heading ***\n\n  " + source + "\n")
			compilation := NewLygerService(canonicalTestOptions()).NewCompilation(original, "dot-source.l")
			compilation.Make()
			if !errors.Is(compilation.Error, lerrors.Syntax) || !strings.Contains(compilation.Error.Error(), "dot-source.l:3:3") || !strings.Contains(compilation.Error.Error(), "dot-qualified") {
				t.Fatalf("dot rejection lost its original source attribution: %v", compilation.Error)
			}
			if compilation.Program != nil || len(compilation.ASMBytes) != 0 || !bytes.Equal(compilation.MainBytes, original) {
				t.Fatal("dot rejection published output or changed original source")
			}
		})
	}
}

func TestAuthoredDotNotationImportsRejectBeforeExpansion(t *testing.T) {
	for _, options := range []Options{{}, canonicalTestOptions()} {
		provider := "embedded"
		if options.AssetRoot != "" {
			provider = "selected provider"
		}
		t.Run(provider, func(t *testing.T) {
			for _, source := range []string{
				"Use Tools.Value from `missing.l`.",
				"Use Tools.Value from `missing.l` as \"Value\".",
				"Use Tools.Inner's Value from `missing.l`.",
				"Use Tools's Inner.Value from `missing.l`.",
				"Use Value from `missing.l` as \"Tools.Value\".",
				"Use Value from `missing.l` as \"Tools\\u002eValue\".",
				"Use Tools.Value from `missing-dictionary`.",
			} {
				t.Run(source, func(t *testing.T) {
					result, err := compileTestSource(t, "\n  "+source+"\n", nil, options)
					if result != nil || !errors.Is(err, lerrors.Syntax) || !strings.Contains(err.Error(), "main.l:2:3") || !strings.Contains(err.Error(), "dot-qualified") {
						t.Fatalf("import must reject dotted source names before opening its dependency: %v", err)
					}
				})
			}
		})
	}
}

func TestAuthoredDotNotationImportedSourceRejects(t *testing.T) {
	result, err := compileTestSource(t, "Use `nested/words.l`.\n", map[string]string{
		"nested/words.l": "*** heading ***\n\n  Show Tools.Value.\n",
	}, canonicalTestOptions())
	if result != nil || !errors.Is(err, lerrors.Syntax) || !strings.Contains(err.Error(), filepath.Join("nested", "words.l")+":3:3") || !strings.Contains(err.Error(), "dot-qualified") {
		t.Fatalf("imported dot rejection lost source ownership: %v", err)
	}
}

func TestAuthoredDotNotationPreservesNumbersAndLiterals(t *testing.T) {
	gram := scopeTestGrammar(t)
	for _, source := range []string{
		"Show 3.14, and -.5",
		"Show +0.5e-2 and 0x1.8p2",
		"Show 1.0e999",
		"Show `Tools.Value and native.value()`",
		"Define \"Wrapper|-s|-like|---ing\" as `<DEF>E{|Statement: ~<@Owner.Member@> + <@Value|Owner.Type@>~|}<DEF>`",
		"Use `folder.name/file.l`",
		"Use Tools's Value from `folder.name/file.l` as \"Selected\"",
	} {
		t.Run(source, func(t *testing.T) {
			statement, err := newLStatement([]byte(source), gram)
			if err == nil {
				err = statement.tokenize()
			}
			if err != nil {
				t.Fatalf("non-name dots must retain their existing meaning: %v", err)
			}
		})
	}
	result, err := compileTestSource(t, "\"quantity\" is 1.0e999.\n", nil, canonicalTestOptions())
	if result != nil || !errors.Is(err, strconv.ErrRange) || strings.Contains(err.Error(), "dot-qualified") {
		t.Fatalf("numeric range failure became a name error: %v", err)
	}
	result, err = compileTestSource(t, "Use `folder.name/words.l`.\n", map[string]string{
		"folder.name/words.l": "\"message\" is `Owner.Member native.value()`.\n",
	}, canonicalTestOptions())
	if err != nil || result == nil || result.Compilation.Memory["message"].Value.Primitive.Text != "`Owner.Member native.value()`" {
		t.Fatalf("literal dots or dotted include paths changed: %v", err)
	}
}
