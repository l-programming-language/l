package Lyger

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/l-programming-language/l/lerrors"
)

func TestAuthoredFamiliesRejectIncompleteRegistration(t *testing.T) {
	for _, source := range []string{
		"Create \"Start\" for Building any App and then Running it.",
		"Define \"Process\" as `<DEF>W^<@Definitions|Syllables@><DEF>`.",
		"\"Bare\" is `<DEF>W^<@Definitions|Syllables@><DEF>`.",
		"Create \"Quiet|-|-|-|-\" for Printing any Text.",
		functionSource("Echo", "\"Value\" as Text", "Text", "Return Value."),
		"Use Print as \"Say\".",
		"Use Print as \"Say\".\nUse Say as \"Tell\".",
		"Use Printer as \"Writer\".",
	} {
		t.Run(source, func(t *testing.T) {
			c := compileMemberResolutionSource(t, memberResolutionSource{main: source})
			if c.Error == nil {
				t.Fatal("incomplete authored family was accepted")
			}
			if !strings.Contains(c.Error.Error(), "famil") && !strings.Contains(c.Error.Error(), "case") {
				t.Fatalf("failed outside family validation: %v", c.Error)
			}
		})
	}
}

func TestFamilyRequirementIncludesPrivateImportsAndCore(t *testing.T) {
	const invalid = "Define \"Facet\" as `<DEF>E{|Statement: ~7~|}<DEF>`.\n"
	for _, source := range []string{invalid, "Create \"Start\" for Printing any Text.\n"} {
		c := compileMemberResolutionSource(t, memberResolutionSource{main: "Use `local-words.l`.\n", library: "    " + source})
		if c.Error == nil || !strings.Contains(c.Error.Error(), "family") || !strings.Contains(c.Error.Error(), "local-words.l") {
			t.Fatalf("private imported declaration escaped validation: %v", c.Error)
		}
		files := dictionaryIndexFilesystem(t)
		files["builtins/eng/core.l"] = &fstest.MapFile{Data: append(append([]byte(nil), files["builtins/eng/core.l"].Data...), []byte("\n"+source)...)}
		c = NewLygerService(Options{Resources: files}).NewCompilation([]byte("Define \"Value\" as 7.\n"), "main.l")
		c.Make()
		if c.Error == nil || !strings.Contains(c.Error.Error(), "family") || !strings.Contains(c.Error.Error(), "core.l") {
			t.Fatalf("selected core declaration escaped validation: %v", c.Error)
		}
	}
}

func TestAuthoredFamilyAliasesPreserveIdentity(t *testing.T) {
	const library = "Create \"Catalog\" as a Dictionary.\n\n    Publicly create Catalog's \"Route|-s|-d|--ing|--er\" for Showing any Text.\n"
	for _, alias := range []string{"Route", "Guide|-s|-d|--ing|--er"} {
		root := strings.Split(alias, "|")[0]
		source := "Use `local-words.l`.\nUse Catalog's Route as \"" + alias + "\".\n" + root + " `hello`.\n"
		c := compileMemberResolutionSource(t, memberResolutionSource{main: source, library: library})
		if c.Error != nil {
			t.Fatal(c.Error)
		}
		found := false
		for local, target := range c.Semantics.Aliases {
			if strings.EqualFold(local.ID, root) {
				found = true
				canonical, err := canonicalRecord(target, c.Semantics.Aliases)
				if err != nil || !strings.EqualFold(canonical.ID, "Catalog.Route") {
					t.Fatalf("alias changed identity: %v, %v", canonical, err)
				}
			}
		}
		if !found {
			t.Fatal("missing local alias")
		}
	}

}

func TestQualifiedAliasRequiresOwnFamily(t *testing.T) {
	for _, label := range []string{"Return", "Return|-s|-ed|-ing|-er"} {
		source := "Create \"Catalog\" as a Dictionary.\nUse Function's Return as Catalog's \"" + label + "\".\n"
		c := compileMemberResolutionSource(t, memberResolutionSource{main: source})
		if label == "Return" {
			if c.Error == nil || !strings.Contains(c.Error.Error(), "family") {
				t.Fatalf("qualified alias inherited a family without unqualifying: %v", c.Error)
			}
		} else if c.Error != nil {
			t.Fatal(c.Error)
		}
	}
}

func TestAliasChainsRequireCanonicalRootAndRole(t *testing.T) {
	const dictionary = "Create \"Catalog\" as a Dictionary.\n"
	for _, tc := range []struct {
		declaration, root, alias string
		valid                    bool
	}{
		{"Use Function's Return as Catalog's \"Return|-s|-ed|-ing|-er\".", "Return", "Return", true},
		{"Use Function's Return as Catalog's \"Finish|-es|-ed|-ing|-er\".", "Finish", "Finish", false},
		{"Use Function's Return as Catalog's \"Finish|-es|-ed|-ing|-er\".", "Finish", "Return", false},
		{"Use Function's Returner as Catalog's \"Return|-s|-like|-ing\".", "Return", "Return", false},
	} {
		t.Run(tc.declaration, func(t *testing.T) {
			c := compileMemberResolutionSource(t, memberResolutionSource{main: dictionary + tc.declaration + "\nUse Catalog's " + tc.root + " as \"" + tc.alias + "\".\n"})
			if tc.valid {
				if c.Error != nil {
					t.Fatal(c.Error)
				}
			} else if c.Error == nil || !strings.Contains(c.Error.Error(), "family") {
				t.Fatalf("renamed or actor alias chain bypassed declaration: %v", c.Error)
			}
		})
	}
}

func TestFunctionBindingsRequireOwnInvocationFamily(t *testing.T) {
	base := functionSource("Read|-s|-|-ing|-er", "\"Value\" as Text", "Text", "Return Value.")
	for _, declaration := range []string{"Define \"Saved\" as Read.", "\"Saved\" is Read."} {
		c := compileMemberResolutionSource(t, memberResolutionSource{main: base + declaration})
		if c.Error != nil {
			t.Fatalf("Function data assignment rejected: %v", c.Error)
		}
		c = compileMemberResolutionSource(t, memberResolutionSource{main: base + declaration + "\nSaved `value` as \"Result\"."})
		if c.Error == nil {
			t.Fatal("single-name Function copy acquired an invocation family")
		}
	}
	for _, source := range []string{
		base + "Define \"Replay|-s|-ed|-ing|-er\" as Read.\nReplay `value` as \"Result\".",
		base + functionSource("Apply|--ies|--ied|-ing|--ier", "\"Replay|-s|-ed|-ing|-er\" as Read", "Text", "Replay `value` as \"Result\".\nReturn Result.") + "Apply Read as \"Result\".",
		base + functionSource("Select|-s|-ed|-ing|-or", "", "Read", "Return Read.") + "Select as \"Replay|-s|-ed|-ing|-er\".\nReplay `value` as \"Result\".",
	} {
		c := compileMemberResolutionSource(t, memberResolutionSource{main: source})
		if c.Error != nil {
			t.Fatalf("explicit callable family failed: %v", c.Error)
		}
		if _, err := planOrchestration(c.Semantics); err != nil {
			t.Fatal(err)
		}
	}
	for _, source := range []string{
		base + functionSource("Apply|--ies|--ied|-ing|--ier", "\"Callback\" as Read", "Text", "Callback `value` as \"Result\".\nReturn Result."),
		base + "Use Reader as \"Task|-s|-like|-ing\".\nCreate \"Apply|--ies|--ied|-ing|--ier\" for Tasking any Text.\nApply `value` as \"Result\".",
	} {
		if c := compileMemberResolutionSource(t, memberResolutionSource{main: source}); c.Error == nil {
			t.Fatal("parameter or nominal alias acquired an implicit invocation family")
		}
	}
}

func TestMemberValueAliasesRemainData(t *testing.T) {
	source := memberResolutionFixture(t)
	source.main += "\nUse y as \"Snapshot\".\nShow Snapshot.\n"
	c := compileMemberResolutionSource(t, source)
	if c.Error != nil {
		t.Fatal(c.Error)
	}
	if c.Memory["Snapshot"] == nil || c.Memory["Snapshot"].Value.Runtime == nil {
		t.Fatal("member-value alias lost its source projection")
	}
}

func TestNativeDataAliasesRemainData(t *testing.T) {
	const header = "    Publicly create \"Exercise|-s|-d|--ing|--er\" as a Generic `exercise` Function with Inputs of \"value\" as `f64` and Outputs of `f64`.\n"
	for _, tc := range []struct{ name, source, observer string }{
		{"parameter", header + "Use Value as \"Snapshot\".\nReturn Snapshot.\n", `assert_eq!(generated::exercise(7.0),7.0);`},
		{"local", "Define \"Cell|-s|-ular|-ing\" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~<@Value@>~|}<DEF>`.\n\n" + header + "Create \"Local\" as a New Cell of Value.\nUse Local as \"Snapshot\".\nReturn Snapshot.\n", `assert_eq!(generated::exercise(7.0),7.0);`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := compileDefinitionSource(t, tc.source, "data-alias.l")
			testNativeDefinitionLibrary(t, c, "Exercise", "data_alias", `pub fn fallback()->String{String::new()}#[path="src/data_alias.rs"]mod generated;fn main(){`+tc.observer+`}`)
		})
	}
}

const nativeFamilyFunction = "    Publicly create \"Read|-s|-|-ing|-er\" as a Generic `read` Function with No Inputs and Outputs of `f64`.\nReturn 7.\n\n"

func nativeFunctionFamilyCopies(family, target string, passive bool) []struct{ name, source string } {
	intermediate := strings.Replace(family, "Task", "Saved", 1)
	var cases []struct{ name, source string }
	declarations := []struct{ name, format string }{
		{"alias", "Use %s as \"%s\".\n"},
		{"define", "Define \"%[2]s\" as %[1]s.\n"},
	}
	if passive {
		declarations = append(declarations, struct{ name, format string }{"passive", "\"%[2]s\" is %[1]s.\n"})
	}
	for _, declaration := range declarations {
		cases = append(cases, struct{ name, source string }{declaration.name, fmt.Sprintf(declaration.format, target, family)})
		for _, prior := range []struct{ name, source, target string }{
			{"copy", fmt.Sprintf("Define \"%s\" as %s.\n", intermediate, target), "Saved"},
			{"qualified", fmt.Sprintf("Create \"Catalog\" as a Dictionary.\nDefine Catalog's \"%s\" as %s.\n", intermediate, target), "Catalog's Saved"},
		} {
			cases = append(cases, struct{ name, source string }{prior.name + "/" + declaration.name, prior.source + fmt.Sprintf(declaration.format, prior.target, family)})
		}
	}
	return cases
}

func TestNativeFunctionBindingsRequireOwnInvocationFamily(t *testing.T) {
	for _, alias := range []struct{ name, family, active string }{
		{"distinct", "Task|-s|-like|-ing", "Tasking"},
		{"root overlap", "Task|-s|-like|-", "Task"},
		{"plural overlap", "Task|-s|-like|-s", "Tasks"},
		{"descriptive overlap", "Task|-s|-like|-like", "Tasklike"},
	} {
		for _, copy := range nativeFunctionFamilyCopies(alias.family, "Reader", true) {
			for _, route := range []string{"positional action", "named action", "native body"} {
				t.Run(alias.name+"/"+copy.name+"/"+route, func(t *testing.T) {
					source := nativeFamilyFunction + copy.source + "Create \"Perform|-s|-ed|-ing|-er\" for " + alias.active + ".\n"
					if route == "native body" {
						source += "\n    Publicly create \"Exercise|-s|-d|--ing|--er\" as a Generic `exercise` Function with No Inputs and Outputs of `f64`.\nReturn a New Performer.\n"
					}
					c := compileDefinitionSource(t, source, "native-invocation-family.l")
					var err error
					switch route {
					case "positional action":
						_, err = c.ComposeDefinition("Perform")
					case "named action":
						_, err = c.instantiateNamedDefinition(c.Memory["Perform"], map[string]DefinitionArgument{}, map[string]bool{}, make(map[*record]bool))
					case "native body":
						_, err = c.ComposeDefinition("Exercise")
					}
					if !errors.Is(err, lerrors.Syntax) || !strings.Contains(err.Error(), "this binding's complete verb family") {
						t.Fatalf("noun alias acquired Function invocation permission through %s: %v", route, err)
					}
				})
			}
		}
	}
}

func TestNativeFunctionBindingFamiliesPreserveConstructionAndCalls(t *testing.T) {
	for _, tc := range []struct{ name, family, target, action, result string }{
		{"noun constructor", "Task|-s|-like|-ing", "Reader", "", "Task"},
		{"overlapping noun constructor", "Task|-s|-like|-s", "Reader", "", "Tasks"},
		{"explicit invocation family", "Task|-s|-ed|-ing|-er", "Read", "Create \"Perform|-s|-ed|-ing|-er\" for Tasking.\n", "Performer"},
	} {
		for _, copy := range nativeFunctionFamilyCopies(tc.family, tc.target, tc.action == "") {
			t.Run(tc.name+"/"+copy.name, func(t *testing.T) {
				source := nativeFamilyFunction + copy.source + tc.action + "\n    Publicly create \"Exercise|-s|-d|--ing|--er\" as a Generic `exercise` Function with No Inputs and Outputs of `f64`.\nReturn a New " + tc.result + ".\n"
				c := compileDefinitionSource(t, source, "native-invocation-control.l")
				if tc.action != "" {
					instance, err := c.instantiateNamedDefinition(c.Memory["Perform"], map[string]DefinitionArgument{}, map[string]bool{}, make(map[*record]bool))
					if err != nil || !strings.Contains(instance.Expression, "Key: ~(read)~") {
						t.Fatalf("named source action lost its explicit invocation family: %s, %v", instance.Expression, err)
					}
				}
				testNativeDefinitionLibrary(t, c, "Exercise", "invocation_family", `#[path="src/invocation_family.rs"]mod generated;fn main(){assert_eq!(generated::exercise(),7.0);}`)
			})
		}
	}
}
