package Lyger

import (
	"strings"
	"testing"
)

func TestFunctionParameterCannotEscapeIntoAppBody(t *testing.T) {
	for _, parameterType := range []string{"Text", "Identity"} {
		t.Run(parameterType, func(t *testing.T) {
			source := distributionIOImport + "\n" + identityFunction + "    Publicly create \"Owner\" as a New Function with Inputs of \"Argument\" as " + parameterType + " and No Outputs.\nCreate \"Child\" as a New Basic App.\nThe App Prints Argument.\nReturn.\n"
			c := compileMemberResolutionSource(t, memberResolutionSource{main: source})
			if c.Error == nil || !strings.Contains(c.Error.Error(), "runtime value Argument cannot be captured by an App body") {
				t.Fatalf("dynamic Function parameter escaped child process: %v", c.Error)
			}
		})
	}
}

func TestFunctionSignatureMalformedSourceRejects(t *testing.T) {
	for _, source := range []string{
		"    Publicly create \"Broken\" as a New Function with Inputs of \"Value\" as Text and \"Value\" as Text and No Outputs.\nReturn.\n",
		"    Publicly create \"Broken\" as a New Function with Inputs of \"Value\" as UnknownType and No Outputs.\nReturn.\n",
		"    Publicly create \"Broken\" as a New Function with Inputs of \"Value\" and Outputs of Text.\nReturn Value.\n",
		"Call `not a function` as \"Result\".\n",
		identityFunction + "Identity `value` as \"First\" and \"Second\".\n",
		identityFunction + "Show Value.\n",
	} {
		t.Run(source, func(t *testing.T) {
			c := compileMemberResolutionSource(t, memberResolutionSource{main: source})
			if c.Error == nil {
				t.Fatal("malformed Function source accepted")
			}
		})
	}
}
