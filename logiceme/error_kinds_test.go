package Logiceme

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/l-programming-language/l/lerrors"
)

func TestBackendErrorKindsAtPublicBoundary(t *testing.T) {
	config := runtimeGenerationResources(t)
	missing := filepath.Join(t.TempDir(), "absent")
	for _, tc := range []struct {
		name, message string
		kind          lerrors.Kind
		run           func() error
	}{
		{"nil service", "source generator configuration is required", lerrors.InvalidConfiguration, func() error { _, err := (*LogicemeService)(nil).Evaluate(nil); return err }},
		{"options", "source generation accepts at most one options value", lerrors.InvalidInput, func() error {
			_, err := NewLogicemeService(config).Evaluate(nil, GenerationOptions{}, GenerationOptions{})
			return err
		}},
		{"asset read", missing, lerrors.IO, func() error { _, err := ReadSysFile(missing); return err }},
		{"templates", "builtins/templates/config/config.lu", lerrors.IO, func() error { _, err := loadSupportTemplates(os.DirFS(missing), false); return err }},
		{"evaluate assets", "builtins/segments/consonants/execute.lu", lerrors.IO, func() error { _, err := NewLogicemeService(os.DirFS(missing)).Evaluate(nil); return err }},
		{"evaluate AML", "Rust document requires exactly one complete STRUCTURE", lerrors.Syntax, func() error { _, err := NewLogicemeService(config).Evaluate([]byte("<PROGRAM></PROGRAM>")); return err }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.run()
			if err == nil {
				t.Fatal("invalid input returned no error")
			}
			if !errors.Is(err, tc.kind) || !strings.Contains(err.Error(), tc.message) {
				t.Errorf("error = %v; want kind %s and unchanged context %q", err, tc.kind, tc.message)
			}
			if tc.kind == lerrors.IO {
				var pathError *os.PathError
				if !errors.Is(err, os.ErrNotExist) || !errors.As(err, &pathError) || pathError.Path != tc.message {
					t.Errorf("backend lost original filesystem cause: %v", err)
				}
			}
		})
	}
}

func TestBackendTemplateErrorKinds(t *testing.T) {
	for _, tc := range malformedLucidFramingCases() {
		t.Run(tc.name, func(t *testing.T) {
			err := validateLucidTemplate(lucidFramingTemplate(tc.source), true)
			if !errors.Is(err, lerrors.InvalidConfiguration) || err == nil || !strings.Contains(err.Error(), "template tag 1") {
				t.Errorf("malformed template = %v; want invalid configuration with template context", err)
			}
		})
	}
	t.Run("loaded segment", func(t *testing.T) {
		root := t.TempDir()
		path := filepath.Join(root, "builtins/segments/consonants/probe.lu")
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("damaged template"), 0600); err != nil {
			t.Fatal(err)
		}
		segment := Segment{Name: "probe", SubType: "consonant"}
		err := segment.loadTemplate(os.DirFS(root))
		if !errors.Is(err, lerrors.InvalidConfiguration) || err == nil || !strings.Contains(err.Error(), "builtins/segments/consonants/probe.lu") || !strings.Contains(err.Error(), "no Lucid tags") {
			t.Errorf("segment loader lost template classification or path: %v", err)
		}
	})
}

func TestBackendAMLParserErrorKindsAndCauses(t *testing.T) {
	for _, tc := range []struct {
		name, cause string
		run         func() error
	}{
		{"metadata syntax", "syntax", func() error { _, err := ParseMetaJSON("AppMeta", `{"AppMeta":}`); return err }},
		{"metadata value", "", func() error { _, err := ParseMetaJSON("AppMeta", `{"AppMeta":{"Label":42}}`); return err }},
		{"JSON tag", "", func() error { _, err := ScanJSON(nil, []byte("<JSON>{}")); return err }},
		{"outline field", "type", func() error { _, err := ScanOutline(nil, []byte(`<JSON>{"Objects":{}}</JSON>`)); return err }},
		{"syllable framing", "", func() error { _, err := ParseSyllable([]byte("<|S|>A</S|>")); return err }},
		{"syllable count", "range", func() error {
			_, err := ParseSyllable([]byte("<|S|>" + strings.Repeat("9", 100) + "(`A`)1</S|>"))
			return err
		}},
		{"syllable exponent", "range", func() error {
			_, err := ParseSyllable([]byte("<|S|>1(`A`)" + strings.Repeat("9", 100) + "</S|>"))
			return err
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.run()
			if !errors.Is(err, lerrors.Syntax) {
				t.Errorf("AML error = %v; want syntax kind", err)
			}
			switch tc.cause {
			case "syntax":
				var native *json.SyntaxError
				if !errors.As(err, &native) {
					t.Errorf("JSON syntax cause lost: %v", err)
				}
			case "type":
				var native *json.UnmarshalTypeError
				if !errors.As(err, &native) || native.Field != "Objects" {
					t.Errorf("JSON type cause lost: %v", err)
				}
			case "range":
				var native *strconv.NumError
				if !errors.Is(err, strconv.ErrRange) || !errors.As(err, &native) || native.Num != strings.Repeat("9", 100) {
					t.Errorf("integer overflow cause lost: %v", err)
				}
			}
		})
	}
}
