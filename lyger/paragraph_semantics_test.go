package Lyger

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// These are prospective paragraph acceptance tests. Compilation errors are
// failures, not expected outcomes: the implementation comes in the next task.
func paragraphSource(t *testing.T) memberResolutionSource {
	t.Helper()
	read := func(name string) string {
		content, err := os.ReadFile(filepath.Join("..", "demo", "paragraphs", name))
		if err != nil {
			t.Fatal(err)
		}
		return string(content)
	}
	return memberResolutionSource{main: read("main.l")}
}

// Isolate coordinated declaration grammar from inline punctuation and local
// scopes, using the demo's exact declaration sentence and bootstrap vocabulary.
func TestParagraphsCoordinatedDeclarations(t *testing.T) {
	source := paragraphSource(t)
	lines := strings.Split(source.main, "\n")
	if len(lines) < 2 {
		t.Fatal("paragraph demo has no declaration sentence")
	}
	source.main = lines[0] + "\n" + lines[1] + "\nCreate \"First\" as a New Basic App.\nThe App Prints x1.\nCreate \"Second\" as a New Basic App.\nThe App Prints x2.\n"
	compilation := compileMemberResolutionSource(t, source)
	if compilation.Error != nil {
		t.Fatal(compilation.Error)
	}
	if !bytes.Contains(compilation.ExpandedBytes, []byte(lines[1])) {
		t.Error("coordinated declaration changed in the expanded-source view")
	}
	for name, text := range map[string]string{"x1": "`Hello World!`", "x2": "`Hello Again World!`"} {
		rec := compilation.Memory[name]
		if rec == nil || rec.Value == nil || rec.Value.Primitive == nil || rec.Value.Primitive.Type != valueText || rec.Value.Primitive.Text != text {
			t.Errorf("coordinated declaration %s must retain exact Text %q; record=%+v", name, text, rec)
		}
	}
}

const paragraphGlobals = "Use IO's Print from `io` as \"Print\".\n\"x1\" is `Hello World!`.\n\"x2\" is `Hello Again World!`.\n"

func paragraphBlock(app, input, owner string) string {
	return "\n    Create \"" + app + "\" as a New Basic App.\nThe App Prints " + input + ".\nDefine \"y\" as " + owner + " Output.\nShow y.\n"
}

func TestParagraphsLocalBindings(t *testing.T) {
	for _, tc := range []struct {
		name, global, firstOwner, secondOwner string
		delayed                               bool
	}{
		{"named owners", "", "First's", "Second's", false},
		{"paragraph pronouns", "", "it's", "it's", false},
		{"local shadows global", "\"y\" is `outer`.\n", "First's", "Second's", false},
		{"first paragraph resolves late", "", "First's", "Second's", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := paragraphSource(t)
			source.main = paragraphGlobals + tc.global + paragraphBlock("First", "x1", tc.firstOwner) + paragraphBlock("Second", "x2", tc.secondOwner)
			if tc.delayed {
				source.main = strings.Replace(source.main, "New Basic App", "New Deferred App", 1)
				source.main += "\n\"Deferred\" describes a Generic App with a New Main Service using any Process.\n"
			}
			compilation := compileMemberResolutionSource(t, source)
			if compilation.Error != nil {
				t.Fatal(compilation.Error)
			}
			locals := map[int]*record{}
			var declarations []*record
			var owners []string
			for _, statement := range compilation.Statements {
				if filepath.Base(statement.source.Span.Path) != "main.l" {
					continue
				}
				for _, tok := range statement.tokens {
					if tok.declarationRoot() != "y" {
						continue
					}
					rec := compilation.Semantics.Subjects[tok]
					if rec == nil || rec.Value == nil {
						t.Fatal("declaration lost its canonical record")
					}
					if rec.Value.Runtime == nil { // The optional enclosing static y.
						continue
					}
					if rec.Value.Runtime.Source != statement || rec.Value.Runtime.Member != memberOutput || rec.Value.Primitive == nil || rec.Value.Primitive.Type != valueText || rec.Value.Primitive.Text != "" {
						t.Fatal("local declaration lost deferred Text, member identity, or original attribution")
					}
					locals[statement.source.Span.Paragraph] = rec
					declarations = append(declarations, rec)
					owners = append(owners, rec.Value.Runtime.OwnerID)
				}
			}
			if len(declarations) != 2 || declarations[0] == declarations[1] || !reflect.DeepEqual(owners, []string{"first", "second"}) {
				t.Fatalf("paragraphs must own distinct y storage and Output owners; records=%v owners=%v", declarations, owners)
			}
			shown := 0
			for _, statement := range compilation.Statements {
				if !strings.HasPrefix(strings.TrimSpace(string(statement.source.Raw)), "Show y.") {
					continue
				}
				shown++
				want := locals[statement.source.Span.Paragraph]
				for _, tok := range statement.tokens {
					if string(tok.rawBytes) == "y" && (want == nil || tok.binding == nil || tok.binding.Record != want) {
						t.Errorf("%s:%d: Show y did not bind the declaration in its own paragraph", statement.source.Span.Path, statement.source.Span.Start.Line)
					}
				}
			}
			if shown != 2 {
				t.Fatalf("expected both local Show statements, got %d", shown)
			}
			if tc.global == "" {
				if compilation.Memory["y"] != nil {
					t.Error("encapsulated y leaked into global memory")
				}
			} else {
				global := compilation.Memory["y"]
				if global == nil || global.Value.Primitive == nil || global.Value.Primitive.Text != "`outer`" || global.Value.Runtime != nil {
					t.Error("local y replaced enclosing global y")
				}
			}
		})
	}
}

func TestParagraphsRejectEscapedLocal(t *testing.T) {
	for _, tc := range []struct{ name, tail string }{
		{"outer program", "\nShow y.\n"},
		{"sibling paragraph", "\n    Create \"Second\" as a New Basic App.\nThe App Prints x2.\nShow y.\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := paragraphSource(t)
			source.main = paragraphGlobals + paragraphBlock("First", "x1", "First's")
			if control := compileMemberResolutionSource(t, source); control.Error != nil {
				t.Fatalf("valid local declaration control did not compile: %v", control.Error)
			}
			source.main += tc.tail
			compilation := compileMemberResolutionSource(t, source)
			if compilation.Error == nil {
				t.Fatal("Show outside the declaring paragraph resolved a local y")
			}
			line := strings.Count(strings.TrimSuffix(source.main, "\n"), "\n") + 1
			message := compilation.Error.Error()
			if !strings.Contains(message, fmt.Sprintf("main.l:%d:", line)) || !strings.Contains(message, "y") {
				t.Fatalf("local visibility error must attribute the unresolved y to its source: %v", compilation.Error)
			}
		})
	}
}

// Ordinary paragraphs are general scopes, not anonymous App callbacks. Runtime
// source ordering and whole-program Exit are exercised by the real CLI suite.
func TestParagraphsOrdinaryScopesRetainDistinctDeclarations(t *testing.T) {
	source := paragraphSource(t)
	source.main = "    \"value\" is `first`.\nShow value.\n" +
		"\n    \"value\" is `second`.\nShow value.\n"
	compilation := compileMemberResolutionSource(t, source)
	if compilation.Error != nil {
		t.Fatal(compilation.Error)
	}
	var declarations []*record
	for _, statement := range compilation.Statements {
		if filepath.Base(statement.source.Span.Path) != "main.l" {
			continue
		}
		for _, tok := range statement.tokens {
			if tok.declarationRoot() == "value" {
				declarations = append(declarations, compilation.Semantics.Subjects[tok])
			}
		}
	}
	if len(declarations) != 2 || declarations[0] == nil || declarations[1] == nil || declarations[0] == declarations[1] {
		t.Fatalf("ordinary paragraph declarations lost independent identity: %v", declarations)
	}
	for index, text := range []string{"`first`", "`second`"} {
		if value := declarations[index].Value; value == nil || value.Primitive == nil || value.Primitive.Text != text {
			t.Errorf("ordinary paragraph %d lost its own value %q", index, text)
		}
	}
	if compilation.Memory["value"] != nil {
		t.Fatal("ordinary paragraph published a private declaration")
	}
}

func TestParagraphsCoordinatedDispatch(t *testing.T) {
	for _, tc := range []struct {
		name, dispatch string
		want           []string
	}{
		{"separate calls control", "Execute First.\nExecute Second.\n", []string{"First", "Second"}},
		{"coordinated calls", "Execute First and Second.\n", []string{"First", "Second"}},
		{"reverse coordination", "Execute Second and First.\n", []string{"Second", "First"}},
		{"canonical alias", "Use Second as \"Other\".\nExecute Other and First.\n", []string{"Second", "First"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := paragraphSource(t)
			// Public Apps escape their defining paragraphs; coordination retains
			// the source-defined action and canonical targets, including aliases.
			source.main = paragraphGlobals + "\n    Create \"First\" as a New Public Basic App.\nThe App Prints x1.\n\n    Create \"Second\" as a New Public Basic App.\nThe App Prints x2.\n\n" + tc.dispatch
			compilation := compileMemberResolutionSource(t, source)
			if compilation.Error != nil {
				t.Fatal(compilation.Error)
			}
			var operations []string
			var collect func([]*instruction)
			collect = func(instructions []*instruction) {
				for _, current := range instructions {
					if current.Kind == instructionRun || current.Kind == instructionBuild {
						if !current.Synchronous || current.Target == nil || current.Target.Record == nil {
							t.Fatal("expanded action lost its synchronous canonical target")
						}
						operations = append(operations, current.Kind.String()+" "+current.Target.Record.ID)
					}
					collect(current.Actions)
				}
			}
			collect(compilation.Semantics.Instructions)
			var want []string
			for _, app := range tc.want {
				want = append(want, "build "+app, "run "+app)
			}
			if !reflect.DeepEqual(operations, want) {
				t.Fatalf("source-defined Execute expansion = %v; want %v", operations, want)
			}
		})
	}
}

func TestParagraphsMultipleSubjectsKeepExactMemberOwners(t *testing.T) {
	source := paragraphSource(t)
	source.main = paragraphGlobals + "\n    Create \"First\" as a New Basic App.\nThe App Prints x1.\n" +
		"Create \"Second\" as a New Basic App.\nThe App Prints x2.\n" +
		"If it has an Error, exit with it's Error, else define \"firstResult\" as First's Output.\n" +
		"Define \"secondResult\" as Second's Output.\nShow firstResult.\nShow secondResult.\n"
	compilation := compileMemberResolutionSource(t, source)
	if compilation.Error != nil {
		t.Fatal(compilation.Error)
	}
	first, _ := paragraphDeclaration(t, compilation, "First")
	second, _ := paragraphDeclaration(t, compilation, "Second")
	for _, tc := range []struct {
		name  string
		owner *record
	}{
		{"firstResult", first}, {"secondResult", second},
	} {
		rec, statement := paragraphDeclaration(t, compilation, tc.name)
		if rec.Value.Runtime == nil || rec.Value.Runtime.Source != statement || rec.Value.Runtime.Member != memberOutput || rec.Value.Runtime.OwnerID != strings.ToLower(tc.owner.ID) {
			t.Errorf("%s did not retain the canonical owner of its own RHS", tc.name)
		}
	}
	pronouns := 0
	for _, statement := range compilation.Statements {
		if !strings.HasPrefix(strings.TrimSpace(string(statement.source.Raw)), "If it has an Error") {
			continue
		}
		for _, tok := range statement.tokens {
			binding, exists := compilation.Scope.Pronouns[tok]
			if !exists {
				continue
			}
			pronouns++
			if binding.Kind != scopeParagraph || compilation.Semantics.Subjects[binding.TargetToken] != second {
				t.Error("conditional pronoun lost the last declared subject while its RHS names another App")
			}
		}
	}
	if pronouns != 2 {
		t.Fatalf("expected condition and possessive Error pronouns, got %d", pronouns)
	}
	for _, name := range []string{"First", "Second", "firstResult", "secondResult"} {
		if compilation.Memory[name] != nil {
			t.Errorf("private subject or dependent value %s leaked globally", name)
		}
	}
}

func TestParagraphsDemoCompiles(t *testing.T) {
	source := paragraphSource(t)
	compilation := compileMemberResolutionSource(t, source)
	if compilation.Error != nil {
		t.Fatal(compilation.Error)
	}
	if compilation.Semantics == nil || compilation.Program == nil || len(compilation.ProcStatements) != 0 {
		t.Fatal("paragraph demo did not complete semantic construction")
	}
	var apps []string
	for _, app := range compilation.Semantics.Apps {
		apps = append(apps, app.ID)
	}
	if !reflect.DeepEqual(apps, []string{"HelloWorldApp1", "HelloWorldApp2"}) || len(compilation.Program.Books) != 2 {
		t.Fatalf("paragraph App construction order = %v; want both demo Apps", apps)
	}
	for _, text := range []string{`(&&LPrintValue(&("Hello World!"))).l_print();`, `(&&LPrintValue(&("Hello Again World!"))).l_print();`} {
		if !bytes.Contains(compilation.ASMBytes, []byte(text)) || !bytes.Contains(compilation.Program.BackendBytes, []byte(text)) {
			t.Errorf("application AML lost paragraph body %s", text)
		}
	}
}
