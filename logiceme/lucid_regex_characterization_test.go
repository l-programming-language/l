package Logiceme

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"regexp"
	"regexp/syntax"
	"strconv"
	"testing"
)

func TestLucidParserRegexCharacterization(t *testing.T) {
	// Keep the original spellings as the capture oracle. Read the expressions
	// from lucidParser itself so this also checks the production patterns.
	pairs := []struct{ selector, original, simplified string }{
		{"MainTag", `\{\|\/\|((?:.*?\r?\n?)*)\|\/\|\}`, `\{\|/\|((?:.*?\r?\n?)*)\|/\|}`},
		{"Expression", `\-\{((?:.*?\r?\n?)*)\}\-`, `-\{((?:.*?\r?\n?)*)}-`},
		{"SubTag", `\(\/((?:.*?\r?\n?)*)\/\)`, `\(/((?:.*?\r?\n?)*)/\)`},
		{"SubExpression", `\=\{((?:.*?\r?\n?)*)\}\=`, `=\{((?:.*?\r?\n?)*)}=`},
		{"ExpressionVar", `\[\|((?:.*?\r?\n?)*)\|\]`, `\[\|((?:.*?\r?\n?)*)\|]`},
		{"VariableKey", `\[((?:.*?\r?\n?)*)\]`, `\[((?:.*?\r?\n?)*)]`},
		{"Imports", `import \(((?:.*?\r?\n?)*)\n\)`, `import \(((?:.*?\r?\n?)*)\n\)`},
		{"Dependencies", `\@\@\"((?:.*?\r?\n?)*)\"\@\@`, `@@"((?:.*?\r?\n?)*)"@@`},
		{"Links", `\!\!\"((?:.*?\r?\n?)*)\"\!\!`, `!!"((?:.*?\r?\n?)*)"!!`},
		{"Alias", `\<A\{((?:.*?\r?\n?)*)\}A\>`, `<A\{((?:.*?\r?\n?)*)}A>`},
		{"EnvDependencies", `\<E\{((?:.*?\r?\n?)*)\}E\>`, `<E\{((?:.*?\r?\n?)*)}E>`},
	}
	inputs := []string{
		"", "incomplete [|", "[a]", "[[a]]", "[|a]|b|]", "x [|a|] [|b|] y",
		"{|/|x|/|}", "{|/|one\r\ntwo|/|}", "-{a}- -{b}-", "(/a/)(/b/)", "={a}=",
		"@@\"α\nβ\"@@", "!!\"path\"!!", "<A{a}A><E{b}E>", "import (a\n)",
		"literal ] } @ / ! < >", "{|/|before -{inside}- after|/|}",
	}
	production := lucidParserRegexSource(t)
	if len(production) != len(pairs) {
		t.Fatalf("production regex count = %d; want %d", len(production), len(pairs))
	}
	for index, pair := range pairs {
		t.Run(pair.selector, func(t *testing.T) {
			originalTree, err := syntax.Parse(pair.original, syntax.Perl)
			if err != nil {
				t.Fatal(err)
			}
			for name, pattern := range map[string]string{"simplified": pair.simplified, "production": production[index]} {
				tree, err := syntax.Parse(pattern, syntax.Perl)
				if err != nil {
					t.Fatalf("%s pattern: %v", name, err)
				}
				if !originalTree.Equal(tree) {
					t.Errorf("%s pattern changes the original regex syntax tree", name)
				}
			}
			original := regexp.MustCompile(pair.original)
			simplified := regexp.MustCompile(pair.simplified)
			current := regexp.MustCompile(production[index])
			for inputIndex, input := range inputs {
				t.Run(fmt.Sprintf("input_%02d", inputIndex), func(t *testing.T) {
					wantIndices := original.FindAllStringSubmatchIndex(input, -1)
					wantCaptures := original.FindAllStringSubmatch(input, -1)
					for name, expression := range map[string]*regexp.Regexp{"simplified": simplified, "production": current} {
						if got := expression.FindAllStringSubmatchIndex(input, -1); !reflect.DeepEqual(got, wantIndices) {
							t.Errorf("%s capture indices for %q = %v; want %v", name, input, got, wantIndices)
						}
						if got := expression.FindAllStringSubmatch(input, -1); !reflect.DeepEqual(got, wantCaptures) {
							t.Errorf("%s captured source for %q = %q; want %q", name, input, got, wantCaptures)
						}
					}
					want := original.FindAllString(input, -1)
					if pair.selector == "SubTag" || pair.selector == "SubExpression" {
						// None of this corpus contains a complete three-expression
						// frame. These selectors intentionally use structural parsing.
						want = nil
					}
					if got := lucidParser(input, pair.selector); !reflect.DeepEqual(got, want) {
						t.Errorf("production parser for %q = %q; want %q", input, got, want)
					}
				})
			}
		})
	}
}

// The ordered literals are read only from lucidParser, not other regex users.
func lucidParserRegexSource(t *testing.T) []string {
	t.Helper()
	source, err := parser.ParseFile(token.NewFileSet(), "lucid.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var patterns []string
	for _, declaration := range source.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Name.Name != "lucidParser" {
			continue
		}
		ast.Inspect(function.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || selector.Sel.Name != "MustCompile" {
				return true
			}
			qualifier, ok := selector.X.(*ast.Ident)
			if !ok || qualifier.Name != "regexp" || len(call.Args) != 1 {
				return true
			}
			literal, ok := call.Args[0].(*ast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				t.Fatal("lucidParser regex is no longer a string literal")
			}
			pattern, err := strconv.Unquote(literal.Value)
			if err != nil {
				t.Fatal(err)
			}
			patterns = append(patterns, pattern)
			return false
		})
	}
	return patterns
}

func TestLucidParserStructuralDispatchCharacterization(t *testing.T) {
	child := "(/={child meta}=={child condition}=={α\r\nβ}=/)"
	body := "before ] } " + child + " after"
	interior := "={meta}=={condition}=={" + body + "}="
	parent := "(/" + interior + "/)"
	for _, test := range []struct {
		name, selector, input string
		want                  []string
	}{
		{"nested tags and repeated siblings", "SubTag", "prefix " + parent + "\r\n" + child + " suffix", []string{parent, child}},
		{"nested expression body", "SubExpression", interior, []string{"={meta}=", "={condition}=", "={" + body + "}="}},
		{"literal tag delimiters", "SubTag", "(/literal/)", nil},
		{"incomplete tag", "SubTag", "(/={meta}=={condition}=/)", nil},
		{"incomplete expressions", "SubExpression", "={meta}=={condition}=", nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := lucidParser(test.input, test.selector); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("production parser = %q; want exact source slices %q", got, test.want)
			}
		})
	}
}
