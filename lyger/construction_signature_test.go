package Lyger

import (
	"strings"
	"testing"
)

func TestConstructionSignatureAbsenceIsLocal(t *testing.T) {
	const program = "\"Typed\" is a New Adjective describing any Service, Object, or Function.\n" +
		"It describes a Generic `typed` Function of any Process with SIGNATURE.\n" +
		"It describes a Generic `typed` Object with a New Objective Typed Function using any Process.\n" +
		"It describes a Generic `typed` Service with a New Typed Object using any Process.\n" +
		"\"WithTyped\" describes a Generic App with a New Main Service and a New Typed Service using any Process.\n" +
		"Create \"Sample\" as a New WithTyped App.\n"
	for _, tc := range []struct{ signature, rejected string }{
		{"No Inputs and No Outputs", ""},
		{"Inputs of `value` Text and No Outputs", "input"},
		{"Outputs of Text and No Inputs", "output"},
	} {
		t.Run(tc.signature, func(t *testing.T) {
			compilation := compileMemberResolutionSource(t, memberResolutionSource{main: strings.Replace(program, "SIGNATURE", tc.signature, 1)})
			if tc.rejected == "" {
				if compilation.Error != nil {
					t.Fatal(compilation.Error)
				}
				return
			}
			want := "nonempty function " + tc.rejected + " signature requires explicit property values"
			if compilation.Error == nil || !strings.Contains(compilation.Error.Error(), want) {
				t.Fatalf("a nested No article must not empty the enclosing signature: want %q, got %v", want, compilation.Error)
			}
		})
	}
}
