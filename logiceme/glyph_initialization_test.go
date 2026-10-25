package Logiceme

import (
	"bytes"
	"errors"
	"testing"

	"github.com/l-programming-language/l/lerrors"
)

func TestInitializeGlyphPublicSuccess(t *testing.T) {
	system, err := loadSegmentalSystem(runtimeGenerationResources(t))
	if err != nil {
		t.Fatal(err)
	}
	input := []byte("E{|Statement: ~native()~|}")
	original := bytes.Clone(input)
	glyph := system.InitializeGlyph(input, nil, nil)
	if glyph.initializationErr != nil {
		t.Fatal(glyph.initializationErr)
	}
	if glyph.Name != string(original) || glyph.SegmentalSystem != &system || len(glyph.Vowels) != 1 || glyph.Vowels[0].Symbol != "E" || glyph.InputString != "Statement: ~native()~" {
		t.Fatalf("initializer lost source, system, form or native input: %#v", glyph)
	}
	if glyph.Content != nil || glyph.plan != nil || glyph.origin != nil {
		t.Fatal("initialization performed preparation or rendering")
	}
	if err := glyph.Compile(); err != nil {
		t.Fatal(err)
	}
	if got, want := string(glyph.Content), "\nnative()\n"; got != want {
		t.Errorf("compiled source = %q; want %q", got, want)
	}
	if !bytes.Equal(input, original) {
		t.Error("initialization or compilation mutated caller input")
	}
}

func TestInitializeGlyphPublicFailureRetainsDiagnostics(t *testing.T) {
	for _, test := range []struct {
		name, source, message string
		kind                  lerrors.Kind
		missingCalculation    bool
		vowels                int
	}{
		{"unknown symbol", "E@", `unknown glyph symbol "@" in E@`, lerrors.Unsupported, false, 1},
		{"unbalanced grouping", "E(", "unbalanced composition parenthesis delimiter in E(", lerrors.Syntax, false, 1},
		{"missing segment", "E", "segment E (Calculation) is not configured", lerrors.InvalidConfiguration, true, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			system, err := loadSegmentalSystem(runtimeGenerationResources(t))
			if err != nil {
				t.Fatal(err)
			}
			if test.missingCalculation {
				system.Calculation = Segment{}
			}
			input := []byte(test.source)
			original := bytes.Clone(input)
			glyph := system.InitializeGlyph(input, nil, nil)
			initializationErr := glyph.initializationErr
			if !errors.Is(initializationErr, test.kind) || initializationErr.Error() != test.message {
				t.Fatalf("initialization error = %v; want %s with context %q", initializationErr, test.kind, test.message)
			}
			if glyph.Name != test.source || glyph.SegmentalSystem != &system || len(glyph.Vowels) != test.vowels || (test.vowels == 1 && glyph.Vowels[0].Symbol != "E") {
				t.Fatalf("failure discarded partial diagnostic state: %#v", glyph)
			}
			for run := 0; run < 2; run++ {
				if err := glyph.Compile(); err != initializationErr {
					t.Errorf("run %d: compilation error = %v; want original initialization error %v", run, err, initializationErr)
				}
				if glyph.Content != nil || glyph.plan != nil || glyph.origin != nil {
					t.Errorf("run %d: malformed glyph reached preparation or rendering", run)
				}
				if !bytes.Equal(input, original) || glyph.Name != test.source || glyph.SegmentalSystem != &system || len(glyph.Vowels) != test.vowels {
					t.Errorf("run %d: failure mutated input or partial diagnostic state", run)
				}
			}
		})
	}
}
