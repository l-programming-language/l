package Lyger

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/l-programming-language/l/lerrors"
)

func runnableServiceCompile(t *testing.T, source string, imports map[string]string, cfg Options) (*CompilationResult, error) {
	t.Helper()
	root := t.TempDir()
	for name, text := range imports {
		writeSourceFile(t, root, name, text)
	}
	main := writeSourceFile(t, root, "main.l", source)
	input := []byte(source)
	result, err := NewLygerService(cfg).Compile(input, main)
	if !bytes.Equal(input, []byte(source)) {
		t.Fatal("frontend changed the supplied L source")
	}
	return result, err
}

func runnableServiceRequire(t *testing.T, source string, imports map[string]string) *CompilationResult {
	t.Helper()
	result, err := runnableServiceCompile(t, source, imports, canonicalTestOptions())
	if err != nil {
		t.Fatalf("frontend contract prerequisite failed; source generation and native execution not exercised: %v", err)
	}
	if result == nil || result.Compilation == nil || result.Compilation.Semantics == nil || result.Compilation.Program == nil {
		t.Fatal("frontend did not return a complete semantic program")
	}
	return result
}

func runnableServiceFixture(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("test", "runnable_service", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// Standalone Service declarations retain lexical identity but have no source
// package or invocation. App descriptor composition remains covered by
// TestWorkspaceConstructedServices and the recursive Class generation controls.
func TestServiceDeclarationsAreStructural(t *testing.T) {
	for _, tc := range []struct{ name, source string }{
		{"ordinary", "Create \"Worker\" as a New Service.\n"},
		{"main", "Create \"Worker\" as a New Main Service.\n"},
		{"descriptor-alias", "Use Main as \"Entry\".\nCreate \"Worker\" as a New Entry Service.\n"},
		{"descriptor-alias-chain", "Use Main as \"Entry\".\nUse Entry as \"Chosen\".\nCreate \"Worker\" as a New Chosen Service.\n"},
		{"composed-descriptor", "\"Wrapped\" describes a New Main Service.\nCreate \"Worker\" as a New Wrapped Service.\n"},
		{"target-alias", "Create \"Worker\" as a New Main Service.\nUse Worker as \"Alias\".\n"},
		{"independent", "Create \"First\" as a New Main Service.\nCreate \"Second\" as a New Main Service.\n"},
		{"private-duplicates", "    Create \"Worker\" as a New Main Service.\n\n    Create \"Worker\" as a New Main Service.\n"},
		{"forward-descriptor", "Create \"Earlier\" as a New Future Service.\nCreate \"Later\" as a New Main Service.\n\"Future\" describes a New Main Service.\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := runnableServiceRequire(t, tc.source+"\nShow `visible`.\n", nil)
			if len(result.Compilation.Program.Apps) != 0 || len(result.Compilation.Program.Books) != 0 || len(result.Compilation.Semantics.Constructions) != 0 || len(result.Compilation.Semantics.Continuations) != 0 {
				t.Fatal("structural Service acquired an App, source book or continuation")
			}
			apps, err := result.ApplicationUnits()
			if err != nil || len(apps) != 0 {
				t.Fatalf("structural Service acquired an application unit: %v, %v", apps, err)
			}
			workspace, err := result.WorkspacePlan()
			if err != nil || len(workspace.Packages()) != 0 {
				t.Fatalf("structural Service acquired a workspace package: %v, %v", workspace, err)
			}
			units, err := workspace.Units(nil)
			if err != nil || len(units) != 0 {
				t.Fatalf("empty workspace projection: %v, %v", units, err)
			}
			if got := paragraphActivationTrace(t, result); !reflect.DeepEqual(got, []string{"show:visible"}) {
				t.Fatalf("structural declaration captured ordinary effects: %v", got)
			}
			if _, err := result.OrchestratorUnit("entry", nil); err != nil {
				t.Fatalf("ordinary top-level program lost its orchestrator: %v", err)
			}
			services := 0
			for rec, declaration := range result.Compilation.Semantics.Declarations {
				value, err := rec.Value.value()
				if err != nil {
					t.Fatal(err)
				}
				if value.Type == valueService {
					services++
					if rec.ID == "" || declaration.source.Span.Path == "" || rec.Value.Applicative.Service.Target != nil || len(rec.Value.Applicative.Service.members) != 0 {
						t.Fatal("structural Service lost declaration identity or acquired generated contents")
					}
				}
			}
			if services == 0 {
				t.Fatal("standalone Service lost its semantic record")
			}
		})
	}
}

func TestServiceForwardDeclarationProvenance(t *testing.T) {
	source := "Create \"Earlier\" as a New Future Service.\nCreate \"Later\" as a New Main Service.\n\"Future\" describes a New Main Service.\n"
	for attempt := 0; attempt < 2; attempt++ {
		result := runnableServiceRequire(t, source, nil)
		for index, name := range []string{"Earlier", "Later"} {
			rec := result.Compilation.Memory[name]
			declaration := result.Compilation.Semantics.Declarations[rec]
			if declaration == nil || declaration.source.Span.Start.Line != index+1 || rec.Value.Applicative.Service == nil {
				t.Fatalf("Service %s lost its declaration identity/provenance", name)
			}
		}
	}
}

func TestStandaloneServiceBodiesAreUnsupported(t *testing.T) {
	for _, tc := range []struct {
		name, source, target, location string
		imports                        map[string]string
	}{
		{"definite", distributionIOImport + "Create \"Worker\" as a New Main Service.\nThe Service Prints `body`.\n", "Worker", "main.l:3:", nil},
		{"ordinary", distributionIOImport + "Create \"Worker\" as a New Service.\nWorker Prints `body`.\n", "Worker", "main.l:3:", nil},
		{"pronoun", distributionIOImport + "Create \"Worker\" as a New Main Service.\nIt Prints `body`.\n", "Worker", "main.l:3:", nil},
		{"service-named-App", distributionIOImport + "Create \"App\" as a New Main Service.\nApp Prints `body`.\n", "App", "main.l:3:", nil},
		{"alias-named-App", distributionIOImport + "Create \"Worker\" as a New Main Service.\nUse Worker as \"App\".\nApp Prints `body`.\n", "Worker", "main.l:4:", nil},
		{"imported", "Use `nested/worker.l`.\n", "Worker", "worker.l:3:", map[string]string{"nested/worker.l": "Use `words.l`.\nCreate \"Worker\" as a New Main Service.\nThe Service Paints `body`.\n", "nested/words.l": "Use IO's Print from `io` as \"Paint\".\n"}},
		{"native-body-fixture", runnableServiceFixture(t, "native-body.l"), "Worker", "main.l:3:", nil},
		{"alias-body-fixture", runnableServiceFixture(t, "native-body-alias.l"), "Worker", "main.l:5:", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, err := runnableServiceCompile(t, tc.source, tc.imports, canonicalTestOptions())
			if result != nil || !errors.Is(err, lerrors.Unsupported) || !strings.Contains(err.Error(), "standalone Service "+tc.target+" body") || !strings.Contains(err.Error(), tc.location) {
				t.Fatalf("standalone body lost contextual rejection: %v", err)
			}
		})
	}
	for _, subject := range []string{"Service", "Alias"} {
		source := distributionIOImport + "Create \"Service\" as a New Basic App.\n"
		if subject == "Alias" {
			source += "Use Service as \"Alias\".\n"
		}
		result := runnableServiceRequire(t, source+subject+" Prints `app-body`.\n", nil)
		units, err := result.ApplicationUnits()
		if err != nil || len(units) != 1 || !bytes.Contains(units[0].BackendAML, []byte("app-body")) {
			t.Fatalf("App named Service lost its canonical body: %v", err)
		}
	}
}

func TestServiceLifecycleAndImplicitMembersReject(t *testing.T) {
	for _, descriptor := range []string{"", "Main "} {
		for _, action := range []string{
			"Build Worker.", "Run Worker.", "Stop Worker.", "Await Worker.",
			"Worker is Built.", "Worker is Run.", "Worker is Stopped.", "Worker is Awaited.",
			"Use Worker as \"Alias\".\nRun Alias.",
			"Create \"Launch|-es|-ed|-ing\" for Running any Service.\nLaunch Worker.",
			"Create \"Host\" as a New Basic App.\nRun Host and Worker.",
			"Start Worker.", "Execute Worker.",
		} {
			t.Run(descriptor+action, func(t *testing.T) {
				source := "Create \"Worker\" as a New " + descriptor + "Service.\n" + action + "\n"
				result, err := runnableServiceCompile(t, source, nil, canonicalTestOptions())
				if result != nil || !errors.Is(err, lerrors.TypeMismatch) || !strings.Contains(strings.ToLower(err.Error()), "app") || !strings.Contains(err.Error(), "main.l:") {
					t.Fatalf("Service lifecycle must reject as an App operand: %v", err)
				}
			})
		}
		for _, expression := range []string{"Show Worker's Output.", "Exit with Worker's Error.", "Use Worker as \"Alias\".\nShow Alias's Output.", "Show it's Output.", "If Worker has an Error, exit."} {
			t.Run(descriptor+expression, func(t *testing.T) {
				result, err := runnableServiceCompile(t, "Create \"Worker\" as a New "+descriptor+"Service.\n"+expression+"\n", nil, canonicalTestOptions())
				if result != nil || err == nil || errors.Is(err, lerrors.Internal) || !strings.Contains(err.Error(), "main.l:") {
					t.Fatalf("implicit Service runtime member must reject contextually: %v", err)
				}
			})
		}
	}
}

func TestServiceUnsupportedTargetsAndOwners(t *testing.T) {
	for _, tc := range []struct {
		name, source, context string
		kind                  lerrors.Kind
	}{
		{"Class", "Create \"Worker\" as a New Class.\nRun Worker.\n", "Class", lerrors.TypeMismatch},
		{"Object", "Create \"Worker\" as a New Object.\nRun Worker.\n", "Object", lerrors.TypeMismatch},
		{"Text", "Define \"Worker\" as `text`.\nRun Worker.\n", "Text", lerrors.TypeMismatch},
		{"App-body", "Create \"Worker\" as a New Basic App.\nThe App Runs Worker.\n", "outer orchestrator", lerrors.Unsupported},
		{"Service-body", "Create \"Worker\" as a New Main Service.\nThe Service Runs Worker.\n", "standalone Service", lerrors.Unsupported},
		{"Start-App-constraint", "Create \"Worker\" as a New Main Service.\nStart Worker.\n", "App", lerrors.TypeMismatch},
		{"Execute-App-constraint", "Create \"Worker\" as a New Main Service.\nExecute Worker.\n", "App", lerrors.TypeMismatch},
		{"private-Service-does-not-escape", "    Create \"Worker\" as a New Main Service.\n\nRun Worker.\n", "Worker", lerrors.UnresolvedReference},
		{"private-entry-alias-does-not-escape", "    Use Main as \"Entry\".\n\nCreate \"Worker\" as a New Entry Service.\n", "Entry", lerrors.UnresolvedReference},
		{"native-Run-alias-remains-unsupported", "Use Run as \"Launch\".\n", "native root", lerrors.Unsupported},
		{"native-Await-alias-remains-unsupported", "Use Await as \"Join\".\n", "native root", lerrors.Unsupported},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, err := runnableServiceCompile(t, tc.source, nil, canonicalTestOptions())
			if result != nil || !errors.Is(err, tc.kind) || !strings.Contains(err.Error(), tc.context) || !strings.Contains(err.Error(), "main.l:") {
				t.Fatalf("unsupported target/owner diagnostic = %v; want source location and %q", err, tc.context)
			}
		})
	}
}

func TestServiceSourceActionAliasAndParameter(t *testing.T) {
	const definition = "Create \"Run|-s|-ned|-ning\" for Showing any Text.\n"
	result := runnableServiceRequire(t, definition+"Use Run as \"LocalRun\".\nLocalRun `source-action-only`.\n", nil)
	flat := phase1Flat(result.Compilation.Semantics.Instructions)
	if len(flat) != 1 || flat[0].Kind != instructionShow || flat[0].Value == nil || flat[0].Value.Literal == nil || flat[0].Value.Literal.Text != "`source-action-only`" {
		t.Fatalf("source-defined Run was redispatched as a native lifecycle operation: %v", phase1Kinds(result.Compilation.Semantics.Instructions))
	}
	result, err := runnableServiceCompile(t, definition+"Create \"Worker\" as a New Main Service.\nRun Worker.\n", nil, canonicalTestOptions())
	if result != nil || !errors.Is(err, lerrors.TypeMismatch) || !strings.Contains(strings.ToLower(err.Error()), "text") {
		t.Fatalf("source-defined Run lost its Text parameter constraint: %v", err)
	}
}

func TestServiceOrdinaryMembersAndActionsRemainLexical(t *testing.T) {
	for _, member := range []string{"Output", "Error"} {
		source := "Create \"Worker\" as a New Service.\nDefine Worker's \"" + member + "\" as `ordinary member`.\nUse Worker as \"Alias\".\nShow Alias's " + member + ".\n"
		result := runnableServiceRequire(t, source, nil)
		instructions := phase1Flat(result.Compilation.Semantics.Instructions)
		if len(instructions) != 1 || instructions[0].Kind != instructionShow || instructions[0].Value.Runtime || instructions[0].Value.Kind == operandMember {
			t.Fatalf("ordinary %s was treated as an implicit process member: %#v", member, instructions)
		}
		if _, err := result.OrchestratorUnit("entry", nil); err != nil {
			t.Fatal(err)
		}
	}
	for _, action := range []string{"Build|-s|-t|-ing", "Run|-s|-ned|-ning", "Stop|-s|-ped|-ping", "Await|-s|-ed|-ing"} {
		name := strings.Split(action, "|")[0]
		source := "Create \"" + action + "\" for Showing any Text.\n" + name + " `ordinary action`.\n"
		result := runnableServiceRequire(t, source, nil)
		if got := phase1Kinds(result.Compilation.Semantics.Instructions); !reflect.DeepEqual(got, []string{"show"}) {
			t.Fatalf("source-defined %s was replaced with a lifecycle primitive: %v", name, got)
		}
	}
}

func TestRunnableAppContinuations(t *testing.T) {
	for _, alias := range []bool{false, true} {
		t.Run(fmt.Sprintf("alias-%v", alias), func(t *testing.T) {
			source := "    Create \"Worker\" as a New Public Basic App.\nDefine \"saved\" as Worker's Output.\nShow saved.\n\nRun Worker.\nAwait Worker.\nAwait Worker.\n"
			if alias {
				source = strings.Replace(source, "\nRun Worker.", "\nUse Worker as \"Alias\".\nRun Alias.", 1)
				source = strings.ReplaceAll(source, "Await Worker.", "Await Alias.")
			}
			result := runnableServiceRequire(t, source, nil)
			if got := phase1Kinds(result.Compilation.Semantics.Instructions); !reflect.DeepEqual(got, []string{"run", "await", "await"}) {
				t.Fatalf("App continuation effects executed eagerly: %v", got)
			}
			plan, err := planOrchestration(result.Compilation.Semantics)
			if err != nil {
				t.Fatalf("App continuation planning: %v", err)
			}
			if len(plan.Root.Body) != 3 {
				t.Fatalf("root plan has %d instructions, want run and two joins", len(plan.Root.Body))
			}
			var activations []*orchestrationBlock
			for _, current := range plan.Root.Body[1:] {
				if current.Source.Kind != instructionAwait || current.Continuation == nil || len(current.Continuation.Body) != 2 {
					t.Fatal("Await lost its App-owned continuation")
				}
				body := current.Continuation.Body
				if body[0].Source.Kind != instructionAssign || body[1].Source.Kind != instructionShow || body[0].Target.Storage == nil || body[1].Value.Storage != body[0].Target.Storage {
					t.Fatal("continuation lost ordered snapshot storage and read")
				}
				activations = append(activations, current.Continuation)
			}
			if len(activations[0].Storage) != 1 || len(activations[1].Storage) != 1 {
				t.Fatal("each continuation must own exactly one stored snapshot")
			}
			if activations[0] == activations[1] || activations[0].Storage[0] == activations[1].Storage[0] {
				t.Fatal("separate joins share activation storage")
			}
		})
	}
}

// Flatten lexical blocks/actions/branches without following Await activations.
// The tests below inspect each invocation's distinct continuation explicitly.
func runnableServicePlanned(body []*orchestrationInstruction) []*orchestrationInstruction {
	var result []*orchestrationInstruction
	for _, current := range body {
		if current.Block != nil {
			result = append(result, runnableServicePlanned(current.Block.Body)...)
			continue
		}
		result = append(result, current)
		result = append(result, runnableServicePlanned(current.Actions)...)
		result = append(result, runnableServicePlanned(current.Then)...)
		result = append(result, runnableServicePlanned(current.Else)...)
	}
	return result
}

func TestRetainedAppImportedContinuation(t *testing.T) {
	for _, indent := range []string{"", "    "} {
		t.Run(fmt.Sprintf("import-indent-%d", len(indent)), func(t *testing.T) {
			source := "    Create \"Worker\" as a New Public Basic App.\nUse `nested/continuation.l`.\n\nRun Worker.\nAwait Worker.\n"
			imports := map[string]string{"nested/continuation.l": indent + "Define \"saved\" as Worker's Output.\nShow saved.\n"}
			result := runnableServiceRequire(t, source, imports)
			plan, err := planOrchestration(result.Compilation.Semantics)
			if err != nil {
				t.Fatalf("imported App continuation plan: %v", err)
			}
			root := runnableServicePlanned(plan.Root.Body)
			if len(root) != 2 || root[0].Source.Kind != instructionRun || root[1].Source.Kind != instructionAwait || root[1].Continuation == nil {
				t.Fatal("imported App continuation escaped into immediate root effects")
			}
			body := runnableServicePlanned(root[1].Continuation.Body)
			if len(body) != 2 || body[0].Source.Kind != instructionAssign || body[1].Source.Kind != instructionShow || body[0].Target.Storage == nil || body[1].Value.Storage != body[0].Target.Storage {
				t.Fatal("imported continuation lost its local snapshot/read binding")
			}
			if body[0].Value.Owner == nil || body[0].Value.Owner.Record != result.Compilation.Memory["Worker"] {
				t.Fatal("imported Output rebound away from the enclosing App")
			}
			for index, current := range body {
				if filepath.Base(current.Source.Source.Path) != "continuation.l" || current.Source.Source.Start.Line != index+1 {
					t.Fatal("imported continuation lost physical source location")
				}
			}
		})
	}
}

func TestRetainedAppNestedAwaitWithStructuralService(t *testing.T) {
	for _, afterSecond := range []bool{false, true} {
		t.Run(fmt.Sprintf("structural-after-second-%v", afterSecond), func(t *testing.T) {
			structural := "Create \"Library\" as a New Main Service.\nUse Library as \"Alias\".\n"
			first := "    Create \"First\" as a New Public Basic App.\n" +
				"Define \"firstResult\" as First's Output.\nAwait Second.\nShow firstResult.\n"
			second := "Create \"Second\" as a New Public Basic App.\n"
			if afterSecond {
				second += structural
			} else {
				first += structural
			}
			source := first + second + "Define \"secondResult\" as Second's Output.\nShow secondResult.\n\n" +
				"Run First and Second.\nAwait First.\n"
			result := runnableServiceRequire(t, source, nil)
			plan, err := planOrchestration(result.Compilation.Semantics)
			if err != nil {
				t.Fatalf("mixed nested Await plan: %v", err)
			}
			root := runnableServicePlanned(plan.Root.Body)
			if len(root) != 3 || root[2].Source.Kind != instructionAwait || root[2].Continuation == nil {
				t.Fatal("mixed paragraph effects lost their outer Await activation")
			}
			outer := root[2].Continuation
			body := runnableServicePlanned(outer.Body)
			if len(body) != 3 || body[0].Source.Kind != instructionAssign || body[1].Source.Kind != instructionAwait || body[2].Source.Kind != instructionShow || body[1].Continuation == nil {
				t.Fatal("caller snapshot, nested join and caller read lost their order")
			}
			inner := body[1].Continuation
			callee := runnableServicePlanned(inner.Body)
			if len(callee) != 2 || callee[0].Source.Kind != instructionAssign || callee[1].Source.Kind != instructionShow {
				t.Fatal("nested App/App join lost the callee-owned continuation")
			}
			if outer.Scope != inner.Scope || outer.ID == inner.ID || len(outer.Storage) != 1 || len(inner.Storage) != 1 || outer.Storage[0] == inner.Storage[0] {
				t.Fatal("Apps sharing one paragraph do not own independent invocation storage")
			}
			if body[0].Value.Owner.Record != result.Compilation.Memory["First"] || callee[0].Value.Owner.Record != result.Compilation.Memory["Second"] || body[2].Value.Storage != outer.Storage[0] || callee[1].Value.Storage != inner.Storage[0] {
				t.Fatal("nested join rebound an Output owner or resumed the callee's storage")
			}
			if body[0].Value.Access != orchestrationAccessOwned || body[2].Value.Access != orchestrationAccessBorrow || outer.Storage[0].Initial != nil || outer.Storage[0].Mutable || inner.Storage[0].Mutable {
				t.Fatal("nested typed plan added hidden initialization, blanket clones or unnecessary mutability")
			}
		})
	}
}

func TestRetainedAppDefiniteAssignment(t *testing.T) {
	const source = "Create \"Worker\" as a New Basic App.\nRun Worker.\nAwait Worker.\n\n    If Worker has an Error, exit with Worker's Error, else define \"saved\" as Worker's Output.\nShow saved.\n"
	for _, scenario := range []string{"exit-or-assignment", "both-branches-assign", "continuing-unassigned-branch"} {
		t.Run(scenario, func(t *testing.T) {
			result := runnableServiceRequire(t, source, nil)
			var branch *instruction
			for _, current := range phase1Flat(result.Compilation.Semantics.Instructions) {
				if current.Kind == instructionIf {
					branch = current
				}
			}
			if branch == nil || len(branch.Then) != 1 || branch.Then[0].Kind != instructionExit || len(branch.Else) != 1 || branch.Else[0].Kind != instructionAssign {
				t.Fatal("source did not produce the checked Error/Output branch prerequisite")
			}
			// These variations test the typed flow join using an actual resolved
			// App member and declaration. They do not invent new L assignment
			// syntax or initialize the target before the conditional.
			switch scenario {
			case "both-branches-assign":
				assignment := *branch.Else[0]
				branch.Then = []*instruction{&assignment}
			case "continuing-unassigned-branch":
				branch.Then = nil
			}
			plan, err := planOrchestration(result.Compilation.Semantics)
			if scenario == "continuing-unassigned-branch" {
				if plan != nil || !errors.Is(err, lerrors.InvalidInput) || !strings.Contains(err.Error(), "saved is not definitely assigned") || !strings.Contains(err.Error(), "main.l:6:") {
					t.Fatalf("unassigned App snapshot accepted or lost diagnostic: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("continuing paths do initialize the App snapshot: %v", err)
			}
			var plannedBranch, read *orchestrationInstruction
			for _, current := range runnableServicePlanned(plan.Root.Body) {
				if current.Source == branch {
					plannedBranch = current
				}
				if current.Source.Kind == instructionShow {
					read = current
				}
			}
			if plannedBranch == nil || read == nil || !plannedBranch.ReachesNext || read.Value.Storage == nil || read.Value.Storage != plannedBranch.Else[0].Target.Storage || read.Value.Storage.Initial != nil || read.Value.Storage.Mutable || read.Value.Access != orchestrationAccessBorrow {
				t.Fatal("continuing-path assignment lost its single uninitialized immutable Text slot")
			}
			if scenario == "exit-or-assignment" && plannedBranch.Then[0].ReachesNext {
				t.Fatal("Exit branch incorrectly contributes an unassigned continuing path")
			}
		})
	}
}

func TestRetainedAppSnapshotMutability(t *testing.T) {
	source := "    Create \"Worker\" as a New Public Basic App.\nDefine \"saved\" as Worker's Output.\n\"copy\" is saved.\nPublicly define \"published\" as copy.\nShow saved.\n\nRun Worker.\nAwait Worker.\nAwait Worker.\nShow published.\n"
	result := runnableServiceRequire(t, source, nil)
	plan, err := planOrchestration(result.Compilation.Semantics)
	if err != nil {
		t.Fatalf("App snapshot mutability plan: %v", err)
	}
	root := runnableServicePlanned(plan.Root.Body)
	if len(root) != 4 || root[1].Continuation == nil || root[2].Continuation == nil || root[3].Source.Kind != instructionShow {
		t.Fatal("snapshot publication lost repeated Await and final read order")
	}
	var previous *orchestrationStorage
	for _, join := range root[1:3] {
		body := runnableServicePlanned(join.Continuation.Body)
		if len(body) != 4 || body[0].Source.Kind != instructionAssign || body[1].Source.Kind != instructionAssign || body[2].Source.Kind != instructionAssign || body[3].Source.Kind != instructionShow {
			t.Fatal("snapshot/copy/publication/read order changed")
		}
		saved, copied, published := body[0].Target.Storage, body[1].Target.Storage, body[2].Target.Storage
		if saved == nil || copied == nil || published == nil || saved == copied || saved == previous || saved.Initial != nil || copied.Initial != nil || saved.Mutable || copied.Mutable || !published.Mutable || !published.Published {
			t.Fatal("private one-write storage and repeatedly published storage have incorrect ownership/mutability")
		}
		if body[0].Value.Access != orchestrationAccessOwned || body[1].Value.Access != orchestrationAccessSnapshot || body[1].Value.Storage != saved || body[2].Value.Access != orchestrationAccessSnapshot || body[2].Value.Storage != copied || body[3].Value.Access != orchestrationAccessBorrow || body[3].Value.Storage != saved || root[3].Value.Storage != published {
			t.Fatal("typed snapshot ownership was replaced by fresh member reads or blanket cloning")
		}
		previous = saved
	}
}

func TestRetainedAppTypedCapture(t *testing.T) {
	for _, tc := range []struct {
		name, value string
		escape      bool
	}{
		{"active-runtime-capture", "Worker's Output", false},
		{"escaped-runtime-capture", "Worker's Output", true},
		{"escaped-static-capture", "`definition text`", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := "    Create \"Worker\" as a New Public Basic App.\nDefine \"privateText\" as " + tc.value + ".\nPublicly create \"Report|-s|-ed|-ing\" for Showing privateText and then Showing any Text.\n"
			if !tc.escape {
				source += "Report `inside`.\n"
			}
			source += "\nRun Worker.\nAwait Worker.\n"
			if tc.escape {
				source += "Report `outside`.\n"
			}
			result := runnableServiceRequire(t, source, nil)
			plan, err := planOrchestration(result.Compilation.Semantics)
			if err != nil {
				t.Fatalf("valid App capture rejected: %v", err)
			}
			var body []*orchestrationInstruction
			root := runnableServicePlanned(plan.Root.Body)
			if tc.escape {
				body = root
			} else {
				if len(root) != 2 || root[1].Continuation == nil {
					t.Fatal("active capture has no App continuation")
				}
				body = runnableServicePlanned(root[1].Continuation.Body)
			}
			var call *orchestrationInstruction
			for _, current := range body {
				if current.Source.Kind == instructionCall {
					call = current
				}
			}
			if call == nil || len(call.Actions) != 2 || call.Actions[0].Value == nil || call.Actions[0].Value.Type != valueText || call.Source.Definition == nil {
				t.Fatal("source-defined capture lost its ordered typed action expansion")
			}
			value := call.Actions[0].Value
			if tc.name == "escaped-static-capture" {
				if value.Storage != nil || value.Literal == nil || value.Literal.Text != "`definition text`" || value.Access != orchestrationAccessOwned {
					t.Fatal("escaped static capture did not retain its definition-site literal")
				}
			} else {
				var binding *orchestrationInstruction
				visitCapturePlan(root[1].Continuation.Body, func(current *orchestrationInstruction) {
					if current.Source.Kind == instructionBindAction && current.Source.Callable == result.Compilation.Memory["Report"] {
						binding = current
					}
				})
				if binding == nil || len(binding.CaptureValues) != 1 || binding.Callable.Owner != plan.Root || call.CaptureSource == nil || call.CaptureSource.Slot != binding.Callable {
					t.Fatal("published action did not retain its continuation's captured environment")
				}
				snapshot := binding.CaptureValues[0]
				if snapshot.Storage == nil || snapshot.Storage.Record.ID != "privateText" || snapshot.Storage.Initial != nil || snapshot.Storage.Mutable || snapshot.Access != orchestrationAccessSnapshot || snapshot.Storage.Owner != root[1].Continuation {
					t.Fatal("capture binding did not snapshot the initialized private App activation storage")
				}
				if value.Storage == nil || value.Storage == snapshot.Storage || value.Storage.Record != snapshot.Record || value.Storage.Capture == nil || value.Storage.Capture.Name != call.CaptureName || value.Storage.Capture.Index != 0 || value.Access != orchestrationAccessBorrow {
					t.Fatal("active or escaped call did not read the selected environment's typed Text slot")
				}
			}
			for _, action := range call.Actions {
				if action.Source.CallSite != call.Source.Statement || action.Source.Statement != call.Source.Definition {
					t.Fatal("capture expansion lost definition-site and invocation provenance")
				}
			}
		})
	}
}
