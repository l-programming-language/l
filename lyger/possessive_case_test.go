package Lyger

import (
	"bytes"
	"strings"
	"testing"
)

func TestPossessiveCaseRetainsOriginalOwner(t *testing.T) {
	for _, tc := range []struct {
		word, owner string
		possessive  bool
	}{
		{"Tools's", "Tools", true}, {"TOOLS'S", "TOOLS", true},
		{"Tools’s", "Tools", true}, {"TOOLS’S", "TOOLS", true},
		{"Tools'", "Tools", true}, {"TOOLS'", "TOOLS", true},
		{"Tools’", "Tools", true}, {"TOOLS’", "TOOLS", true},
		{"雪TOOLS’S", "雪TOOLS", true}, {"IT'S", "IT", true},
		{"Tools", "Tools", false}, {"snow’sfall", "snow’sfall", false}, {"", "", false},
	} {
		t.Run(tc.word, func(t *testing.T) {
			owner, possessive := possessiveOwner(tc.word)
			if owner != tc.owner || possessive != tc.possessive {
				t.Fatalf("possessiveOwner(%q) = %q, %t; want %q, %t", tc.word, owner, possessive, tc.owner, tc.possessive)
			}
		})
	}
}

func TestPossessiveCaseDeclarationsAndMembers(t *testing.T) {
	for _, spelling := range []string{"Tools's", "TOOLS's", "TOOLS'S", "TOOLS’S", "TOOLS'", "TOOLS’"} {
		t.Run(spelling, func(t *testing.T) {
			source := "Create \"Tools\" as a Dictionary.\nCreate " + spelling + " \"Value\" as a Dictionary.\nUse " + spelling + " Value as \"Alias\".\n"
			c := paragraphPublicationSource(t, source)
			if c.Error != nil {
				t.Fatal(c.Error)
			}
			owner, member, alias := c.Memory["Tools"], c.Memory["Tools.Value"], c.Memory["Alias"]
			canonical, err := canonicalRecord(alias, c.Semantics.Aliases)
			if err != nil || member == nil || canonical != member || owner.members[member.ID] != member || alias.Value.Iterative.Dictionary != member.Value.Iterative.Dictionary {
				t.Fatalf("possessive case changed canonical member ownership: %v", err)
			}
			declaration := c.Semantics.Declarations[member]
			if declaration.source.Span.Start.Line != 2 || !bytes.Contains(declaration.source.Raw, []byte(spelling)) || !bytes.Equal(c.MainBytes, []byte(source)) {
				t.Fatal("possessive normalization changed original source or provenance")
			}
		})
	}
}

func TestPossessiveCaseImportOwners(t *testing.T) {
	for _, spelling := range []string{"Tools's", "TOOLS'S", "TOOLS’S", "TOOLS'", "TOOLS’"} {
		t.Run(spelling, func(t *testing.T) {
			source := memberResolutionSource{
				main:    "Use " + spelling + " Value from `local-words.l` as \"Alias\".\n",
				library: "Create \"Tools\" as a Dictionary.\nCreate Tools's \"Value\" as a Dictionary.\n",
			}
			c := compileMemberResolutionSource(t, source)
			if c.Error != nil {
				t.Fatal(c.Error)
			}
			canonical, err := canonicalRecord(c.Memory["Alias"], c.Semantics.Aliases)
			if err != nil || canonical == nil || canonical != c.Memory["Tools.Value"] {
				t.Fatalf("import owner case lost its original member: %v", err)
			}
			aliasSource := c.Semantics.Declarations[c.Memory["Alias"]].source
			if string(aliasSource.Raw) != source.main || aliasSource.Span.Start.Line != 1 {
				t.Fatal("import owner case changed alias source provenance")
			}
		})
	}
}

func TestPossessiveCaseNativeOwners(t *testing.T) {
	fixture := memberResolutionFixture(t)
	baseline := compileMemberResolutionSource(t, fixture)
	requireMemberResolutionProgram(t, baseline)
	for _, owner := range []string{"HelloWorldApp's", "HELLOWORLDAPP'S", "HELLOWORLDAPP’S", "IT'S", "IT’S"} {
		t.Run(owner, func(t *testing.T) {
			source := fixture
			if strings.Count(source.main, "it's Output") != 1 || strings.Count(source.main, "the Error") != 1 {
				t.Fatal("fixture no longer has its single Output and Error operand")
			}
			source.main = strings.Replace(source.main, "it's Output", owner+" Output", 1)
			source.main = strings.Replace(source.main, "the Error", owner+" Error", 1)
			c := compileMemberResolutionSource(t, source)
			requireMemberResolutionProgram(t, c)
			app := c.Memory["HelloWorldApp"]
			output := c.Memory["y"].Value.Runtime
			if output == nil || output.Owner != app || output.Member != memberOutput || !bytes.Equal(c.ASMBytes, baseline.ASMBytes) {
				t.Fatal("possessive spelling changed runtime owner or application body")
			}
			var uses int
			for _, statement := range c.Statements {
				for index, tok := range statement.tokens {
					if index == 0 || string(statement.tokens[index-1].rawBytes) != owner || (string(tok.rawBytes) != "Output" && string(tok.rawBytes) != "Error") {
						continue
					}
					if tok.mem.Runtime == nil || tok.mem.Runtime.Owner != app || !bytes.Contains(statement.source.Raw, []byte(owner+" "+string(tok.rawBytes))) {
						t.Fatal("native member lost its canonical owner or exact source spelling")
					}
					uses++
				}
			}
			if uses != 2 {
				t.Fatalf("native member source uses = %d; want Output and Error", uses)
			}
		})
	}
}
