package Logiceme

import (
	"strings"
	"testing"
)

func lucidFramingTemplate(body string) []byte {
	return []byte(`{|/|-{tags: Meta.Tags}--{if Type == "Default"}--{` + body + `}-|/|}`)
}

func TestLucidFramingPreservesTextAndTagOrder(t *testing.T) {
	first := `(/={tags: Meta.Tags}=={if Meta.Key == Void}=={first}=/)`
	second := `(/={tags: Meta.Tags}=={if Meta.Key != Void}=={second}=/)`
	emptyBody := `(/={tags: Meta.Tags}=={if Meta.Key == Void}=={}=/)`
	leaf := `(/={tags: Meta.Tags}=={if Meta.Key == Void}=={fmt.Print("(/foo)", "/)", "(/foo/)", "(/=foo)", "a=/)", "雪")}=/)`
	nested := `(/={tags: Meta.Tags}=={if Meta.Key == Void}=={` + leaf + `}=/)`
	spaced := "(/ \n={tags: Meta.Tags}= \n={if Meta.Key == Void}= \n={body}= \n/)"
	for _, tc := range []struct {
		name, source string
		want         []string
	}{
		{"empty", "", nil},
		{"opening parenthesis at EOF", "(", nil},
		{"closing parenthesis at start", ")", nil},
		{"literal opener", `fmt.Print("(/foo)")`, nil},
		{"literal closer", `fmt.Print("/)")`, nil},
		{"literal pair", `fmt.Print("(/foo/)")`, nil},
		{"partial expression opener", `fmt.Print("(/=foo)")`, nil},
		{"partial expression closer", `fmt.Print("a=/)")`, nil},
		{"unicode literal", `fmt.Print("雪の文字")`, nil},
		{"empty framed body", emptyBody, []string{emptyBody}},
		{"framed literal body", leaf, []string{leaf}},
		{"nested literal body", nested, []string{nested}},
		{"siblings with intervening text", "雪" + first + "\n(/foo/)\n" + second + "光", []string{first, second}},
		{"spaced expression framing", spaced, []string{spaced}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// A boundary panic must fail this case without aborting the other
			// independent literal and framing regressions.
			defer func() {
				if recovered := recover(); recovered != nil {
					t.Errorf("framing panicked for %q: %v", tc.source, recovered)
				}
			}()
			if err := validateLucidTemplate(lucidFramingTemplate(tc.source), true); err != nil {
				t.Errorf("valid template rejected: %v", err)
			}
			got := lucidParser(tc.source, "SubTag")
			if len(got) != len(tc.want) {
				t.Fatalf("tag slices = %q; want %q", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("tag %d = %q; want unchanged source slice %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func malformedLucidFramingCases() []struct{ name, source string } {
	child := `(/={tags: Meta.Tags}=={if Meta.Key == Void}=={child}=/)`
	return []struct{ name, source string }{
		{"incomplete first expression", "(/={tags: Meta.Tags"},
		{"missing third expression", `(/={tags: Meta.Tags}=={if Meta.Key == Void}=/)`},
		{"extra expression", `(/={tags: Meta.Tags}=={if Meta.Key == Void}=={body}=={extra}=/)`},
		{"missing expression close", `(/={tags: Meta.Tags}=={if Meta.Key == Void}=={body=/)`},
		{"missing tag close", `(/={tags: Meta.Tags}=={if Meta.Key == Void}=={body}=`},
		{"nested malformed child", `(/={tags: Meta.Tags}=={if Meta.Key == Void}=={(/={tags: Meta.Tags}=={if Meta.Key == Void}=/)}=/)`},
		{"child cannot replace parent expression", `(/={tags: Meta.Tags}=={` + child + `}=/)`},
	}
}

func TestLucidFramingRejectsMalformedConstructs(t *testing.T) {
	for _, tc := range malformedLucidFramingCases() {
		t.Run(tc.name, func(t *testing.T) {
			err := validateLucidTemplate(lucidFramingTemplate(tc.source), true)
			if err == nil || !strings.Contains(err.Error(), "template tag 1") || !strings.Contains(err.Error(), "subtag") {
				t.Fatalf("malformed construct requires a contextual framing error; got %v", err)
			}
		})
	}
}
