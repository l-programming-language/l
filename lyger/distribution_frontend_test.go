package Lyger

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"testing/fstest"

	distributiontest "github.com/l-programming-language/l/internal/testsupport/distribution"
	"github.com/l-programming-language/l/lerrors"
	Logiceme "github.com/l-programming-language/l/logiceme"
)

const distributionIOImport = "Use IO's Print from `io` as \"Print\".\n"
const distributionApp = "\"message\" is `Hello World!`.\nCreate \"Hello\" as a New Basic App.\nThe App Prints message.\nExecute Hello.\nShow Hello's Output.\n"

// These cases deliberately request normal resources. No adjacent bootstrap or
// replacement definition is supplied when the default distribution is absent.
func distributionFrontend(t *testing.T, source string, files map[string]string) (*CompilationResult, error) {
	t.Helper()
	dir := t.TempDir()
	for name, text := range files {
		writeSourceFile(t, dir, name, text)
	}
	main := writeSourceFile(t, dir, "main.l", source)
	result, err := NewLygerService(Options{}).Compile([]byte(source), main)
	if err == nil {
		assertFrontendResources(t, result.Compilation)
	}
	return result, err
}

func requireDistributionFrontend(t *testing.T, source string, files map[string]string) *CompilationResult {
	t.Helper()
	result, err := distributionFrontend(t, source, files)
	if err != nil {
		t.Fatalf("default distribution/frontend prerequisite failed; binding and generated-code assertions not exercised; runtime not exercised: %v", err)
	}
	if result == nil || result.Compilation == nil || result.Compilation.Semantics == nil || result.Compilation.Scope == nil {
		t.Fatal("frontend produced no complete semantic program; runtime not exercised")
	}
	return result
}

func distributionCanonical(t *testing.T, c *Compilation, rec *record) *record {
	t.Helper()
	if rec == nil {
		t.Fatal("expected a bound record")
	}
	canonical, err := canonicalRecord(rec, c.Semantics.Aliases)
	if err != nil || canonical == nil {
		t.Fatalf("canonical binding = %v, %v", canonical, err)
	}
	return canonical
}

func distributionNegative(t *testing.T, source, detail string, kind lerrors.Kind) {
	t.Helper()
	result, err := distributionFrontend(t, source, nil)
	if result != nil || !errors.Is(err, kind) || !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(detail)) {
		t.Fatalf("wrong rejection: result=%v error=%v; want %s containing %q", result, err, kind, detail)
	}
}

func TestDistributionBuiltins(t *testing.T) {
	t.Run("BI01-implicit-foundation", func(t *testing.T) {
		const source = "Use Process as \"Steps\".\nUse Function as \"Routine\".\nUse Function's Return as \"ReturnValue\".\nUse Objective as \"Targeted\".\nUse Main as \"Entry\".\nUse Basic as \"Plain\".\nUse Start as \"Kick\".\nUse Execute as \"Operate\".\nCreate \"Worker\" as a New Plain App.\n"
		r := requireDistributionFrontend(t, source, nil)
		c := r.Compilation
		for alias, id := range map[string]string{"Steps": "Process", "Routine": "Function", "ReturnValue": "Function.Return", "Targeted": "Objective", "Entry": "Main", "Plain": "Basic", "Kick": "Start", "Operate": "Execute"} {
			canonical := distributionCanonical(t, c, c.Memory[alias])
			if canonical.ID != id || c.Semantics.Declarations[canonical] == nil {
				t.Errorf("%s did not bind source-defined %s: %+v", alias, id, canonical)
			}
			if c.Memory[id] != nil {
				t.Errorf("implicit %s was flattened into program Memory", id)
			}
		}
		if len(c.Semantics.Apps) != 1 || len(c.Semantics.Constructions) != 1 {
			t.Fatal("implicit Basic did not construct the single declared App")
		}
		if !bytes.Equal(c.MainBytes, []byte(source)) {
			t.Fatal("implicit source modified MainBytes")
		}
	})
	t.Run("BI02-lifecycle-compositions", func(t *testing.T) {
		for _, tc := range []struct {
			verb string
			want []string
		}{{"Start", []string{"build", "run"}}, {"Execute", []string{"build", "run", "await"}}} {
			t.Run(tc.verb, func(t *testing.T) {
				r := requireDistributionFrontend(t, "Create \"Worker\" as a New Basic App.\n"+tc.verb+" Worker.\n", nil)
				phase1Lifecycle(t, r, tc.want)
				calls := r.Compilation.Semantics.Instructions
				if len(calls) != 1 || calls[0].Kind != instructionCall || calls[0].Definition == nil || !calls[0].Synchronous {
					t.Fatalf("%s lost its source action definition/ordered expansion: %+v", tc.verb, calls)
				}
			})
		}
	})
	t.Run("BI03-no-primitive-spelling-dispatch", func(t *testing.T) {
		g := phase1Grammar(t)
		for _, word := range []string{"start", "execute", "print"} {
			if lex, _, err := g.lex.lookup([]byte(word)); !errors.Is(err, lerrors.UnresolvedReference) || (lex != nil && lex.Root != "") {
				t.Errorf("source word %s is in primitive vocabulary: %+v, %v", word, lex, err)
			}
		}
		r := requireDistributionFrontend(t, "Create \"Launch|-es|-ed|-ing\" for Building any App and then Running it.\nCreate \"Collect|-s|-ed|-ing\" for Launching any App and then Awaiting it.\nUse Collect as \"Gather\".\nCreate \"Worker\" as a New Basic App.\nGather Worker.\n", nil)
		phase1Lifecycle(t, r, []string{"build", "run", "await"})
	})
	t.Run("BI04-alias-morphology-and-construction", func(t *testing.T) {
		r := requireDistributionFrontend(t, "Use Start as \"Kick\".\nUse Execute as \"Operate\".\nUse Basic as \"Plain\".\nUse Function's Return as \"Yield\".\nCreate \"Worker\" as a New Plain App.\n", nil)
		for name, forms := range map[string][]string{"Kick": {"Kick", "Kicks", "Kicked", "Kicking", "Kicker"}, "Operate": {"Operate", "Operates", "Operated", "Operating", "Operator"}, "Yield": {"Yield", "Yields", "Yielded", "Yielding", "Yielder"}} {
			alias := r.Compilation.Memory[name]
			want := distributionCanonical(t, r.Compilation, alias)
			for _, form := range forms {
				binding, err := resolveTokenBinding([]byte(form), map[string]*record{name: alias})
				if err != nil || binding == nil || distributionCanonical(t, r.Compilation, binding.Record) != want {
					t.Errorf("%s lost alias morphology: %v", form, err)
				}
			}
		}
		construction := r.Compilation.Semantics.Constructions[0]
		var functions, objectiveWrappers int
		var walk func(*constructionNode)
		walk = func(node *constructionNode) {
			if node.Definition != nil && node.Definition.ID == "Function" {
				functions++
				if node.Name != "main" || !node.EmptyInputs || !node.EmptyOutputs {
					t.Errorf("Main function relationships changed: %+v", node)
				}
			}
			for _, wrapper := range node.Wrappers {
				if wrapper.Source != nil && strings.Contains(string(wrapper.Source.source.Raw), "\"Objective\" describes") {
					objectiveWrappers++
				}
			}
			for _, child := range node.Children {
				walk(child)
			}
		}
		walk(construction.Root)
		if functions != 1 || objectiveWrappers != 1 {
			t.Fatalf("Function/Objective construction relationships = %d/%d", functions, objectiveWrappers)
		}
	})
	t.Run("BI05-no-effects-or-pronoun-subject", func(t *testing.T) {
		r := requireDistributionFrontend(t, "Show `only-user-effect|`.\n", nil)
		if got := phase1Kinds(r.Compilation.Semantics.Instructions); !reflect.DeepEqual(got, []string{"show"}) || len(r.Compilation.Semantics.Continuations) != 0 {
			t.Fatalf("implicit initialization added runtime effects/continuations: %v", got)
		}
		distributionNegative(t, "Show it.\n", "no preceding subject", lerrors.UnresolvedReference)
	})
	t.Run("BI06-local-shadowing", func(t *testing.T) {
		const source = "Use Execute as \"OriginalExecute\".\n\n    Create \"Execute|-s|-d|--ing|--or\" for Showing any Text.\nExecute `local-only|`.\n\nCreate \"Worker\" as a New Basic App.\nOriginalExecute Worker.\n"
		r := requireDistributionFrontend(t, source, nil)
		original := distributionCanonical(t, r.Compilation, r.Compilation.Memory["OriginalExecute"])
		var local *record
		for rec, declaration := range r.Compilation.Semantics.Declarations {
			if rec.ID == "Execute" && declaration.source.Span.Path == r.Compilation.MainPath {
				local = rec
			}
		}
		if local == nil || local == original || r.Compilation.Memory["Execute"] != nil {
			t.Fatal("local shadow altered or published the builtin")
		}
		phase1Lifecycle(t, r, []string{"build", "run", "await"})
		if !reflect.DeepEqual(phase1Kinds(r.Compilation.Semantics.Instructions), []string{"show", "build", "run", "await"}) {
			t.Fatal("local source word did not own its call")
		}
		again := requireDistributionFrontend(t, "Create \"Worker\" as a New Basic App.\nExecute Worker.\n", nil)
		phase1Lifecycle(t, again, []string{"build", "run", "await"})
	})
}

func TestDistributionIO(t *testing.T) {
	t.Run("IO02-explicit-import-and-private-helpers", func(t *testing.T) {
		r := requireDistributionFrontend(t, distributionIOImport+distributionApp, nil)
		for _, name := range []string{"IO", "IO.Print", "Log", "Logger"} {
			if r.Compilation.Memory[name] != nil {
				t.Errorf("selected import leaked %s", name)
			}
		}
		t.Run("Print-needs-import", func(t *testing.T) {
			distributionNegative(t, distributionApp, "Prints", lerrors.UnresolvedReference)
		})
		for _, word := range []string{"Log", "Logger"} {
			t.Run(word+"-is-private", func(t *testing.T) {
				distributionNegative(t, distributionIOImport+distributionApp+"Show "+word+".\n", word, lerrors.UnresolvedReference)
			})
		}
	})
	t.Run("IO04-alias-forms-and-canonical-body", func(t *testing.T) {
		base := requireDistributionFrontend(t, distributionIOImport+distributionApp, nil)
		for _, name := range []string{"Display", "Render"} {
			t.Run(name, func(t *testing.T) {
				source := "Use IO's Print from `io` as \"" + name + "\".\n" + strings.Replace(distributionApp, "Prints", name+"s", 1)
				r := requireDistributionFrontend(t, source, nil)
				alias := r.Compilation.Memory[name]
				canonical := distributionCanonical(t, r.Compilation, alias)
				if canonical.ID != "IO.Print" {
					t.Fatalf("canonical target = %s", canonical.ID)
				}
				for _, suffix := range []string{"", "s", "ed", "ing", "er"} {
					binding, err := resolveTokenBinding([]byte(name+suffix), map[string]*record{name: alias})
					if err != nil || binding == nil || distributionCanonical(t, r.Compilation, binding.Record) != canonical {
						t.Errorf("alias form %s%s: %v", name, suffix, err)
					}
				}
				if !bytes.Equal(base.BackendAML, r.BackendAML) {
					t.Fatal("renaming selected Print changed the child program")
				}
			})
		}
	})
	t.Run("IO05-caller-cannot-replace-private-Log", func(t *testing.T) {
		const caller = "Define \"Log|-s|-ged|-ging|-ger\" to `<DEF>BOD^E{|Statement: ~::std::print!(\"{}\", \"caller-wrong|\");~|}<DEF>`.\n"
		r := requireDistributionFrontend(t, caller+distributionIOImport+distributionApp, nil)
		if !bytes.Contains(r.BackendAML, []byte(`(&&LPrintValue(&("Hello World!"))).l_print();`)) || bytes.Contains(r.BackendAML, []byte("caller-wrong|")) {
			t.Fatalf("caller substituted dictionary-private helper:\n%s", r.BackendAML)
		}
		callerLog := r.Compilation.Memory["Log"]
		var privateLog *record
		for rec, declaration := range r.Compilation.Semantics.Declarations {
			if rec.ID == "Log" && declaration.source.Span.Path != r.Compilation.MainPath {
				privateLog = rec
			}
		}
		if callerLog == nil || privateLog == nil || privateLog == callerLog || privateLog.storageScope == callerLog.storageScope {
			t.Fatal("IO helper and caller share mutable bindings/scope")
		}
	})
	t.Run("IO06-ordinary-source-and-backend", func(t *testing.T) {
		r := requireDistributionFrontend(t, distributionIOImport+distributionApp, nil)
		print := distributionCanonical(t, r.Compilation, r.Compilation.Memory["Print"])
		declaration := r.Compilation.Semantics.Declarations[print]
		if declaration == nil || !strings.Contains(string(declaration.source.Raw), "Create IO's") || print.Value.Lexeme == nil || len(print.Value.Lexeme.Actions) == 0 {
			t.Fatal("Print lost ordinary source definition/action representation")
		}
		units, err := r.ApplicationUnits()
		if err != nil || len(units) != 1 {
			t.Fatalf("child units = %d, %v", len(units), err)
		}
		bundle, err := Logiceme.NewLogicemeService(nil).Evaluate(units[0].BackendAML, Logiceme.GenerationOptions{IncludeRuntime: false})
		recordBundle(t, bundle)
		if err != nil {
			t.Fatalf("real backend prerequisite failed; runtime not exercised: %v", err)
		}
		if calls := definitionTextMainCalls(t, bundle, "l_print"); !reflect.DeepEqual(calls, []string{"Hello World!"}) {
			t.Fatalf("actual generated print calls = %q", calls)
		}
		wrapper := "Use IO's Print from `io` as \"ImportedPrint\".\nCreate \"LocalWords\" as a Dictionary.\nUse ImportedPrint as LocalWords's \"Paint\".\n"
		local := requireDistributionFrontend(t, "Use LocalWords's Paint from `words.l` as \"Print\".\n"+distributionApp, map[string]string{"words.l": wrapper})
		if !bytes.Equal(local.BackendAML, r.BackendAML) {
			t.Fatal("ordinary local reexport changed source-defined IO behavior")
		}
	})
}

func TestDistributionScopes(t *testing.T) {
	t.Run("DI08-unit-roots-are-not-activation-frames", func(t *testing.T) {
		r := requireDistributionFrontend(t, distributionIOImport+"\"global\" is `root|`.\n\n    Create \"Worker\" as a New Public Basic App.\nThe App Prints global.\nDefine \"privateResult\" as Worker's Output.\nShow privateResult.\n\nRun Worker.\nAwait Worker.\nShow global.\n", nil)
		c := r.Compilation
		if c.Memory["global"] == nil || c.Memory["global"].storageScope != c.Scope.Root || c.Memory["Worker"] == nil || c.Memory["privateResult"] != nil {
			t.Fatal("builtin parent changed program-global or paragraph-private ownership")
		}
		if len(c.Semantics.Continuations) != 1 {
			t.Fatalf("continuations = %d; want only Worker", len(c.Semantics.Continuations))
		}
		if _, err := r.OrchestratorUnit("entry", executableTestCompanions(r)); err != nil {
			t.Fatalf("unit-root/definite-assignment lowering failed: %v", err)
		}
	})
	t.Run("DI11-independent-parallel-compilations", func(t *testing.T) {

		requireDistributionFrontend(t, distributionIOImport+distributionApp, nil)
		const count = 8
		var wg sync.WaitGroup
		results := make([]*CompilationResult, count)
		errs := make([]error, count)
		dirs := make([]string, count)
		for i := range dirs {
			dirs[i] = t.TempDir()
		}
		for i := range results {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				results[i], errs[i] = NewLygerService(Options{}).Compile([]byte(distributionIOImport+distributionApp), filepath.Join(dirs[i], "main.l"))
			}(i)
		}
		wg.Wait()
		var canonical []*record
		for i, r := range results {
			if errs[i] != nil || r == nil {
				t.Fatalf("parallel compilation %d: %v", i, errs[i])
			}
			current := distributionCanonical(t, r.Compilation, r.Compilation.Memory["Print"])
			for _, previous := range canonical {
				if current == previous || current.Value.Lexeme == previous.Value.Lexeme {
					t.Fatal("parallel compilations shared dictionary definitions")
				}
			}
			canonical = append(canonical, current)
			if !bytes.Equal(r.BackendAML, results[0].BackendAML) {
				t.Fatal("parallel compilation changed deterministic AML")
			}
		}
		original := canonical[1].Value.Lexeme.Root
		canonical[0].Value.Lexeme.Root = "changed-only-here"
		if canonical[1].Value.Lexeme.Root != original {
			t.Fatal("mutable source prototype escaped its compilation")
		}
	})
	t.Run("DI12-source-action-argument-possessive", func(t *testing.T) {
		const prefix = distributionIOImport + "Create \"Tools\" as a Dictionary.\nCreate Tools's \"Collect|-s|-ed|-ing|-or\" for Awaiting any App and then Showing it's Output.\nUse Tools's Collect as \"Collect\".\n\"first\" is `worker`.\n\"second\" is `other`.\nCreate \"Worker\" as a New Basic App.\nThe App Prints first.\nCreate \"Other\" as a New Basic App.\nThe App Prints second.\nRun Worker and Other.\nAwait Other.\n"
		for _, alias := range []bool{false, true} {
			t.Run(fmt.Sprint(alias), func(t *testing.T) {
				call := "Collect Worker.\n"
				if alias {
					call = "Use Collect as \"Gather\".\nGather Worker.\n"
				}
				r := requireDistributionFrontend(t, prefix+call, nil)
				flat := phase1Flat(r.Compilation.Semantics.Instructions)
				if len(flat) < 2 {
					t.Fatal("source action lost ordered effects")
				}
				join, show := flat[len(flat)-2], flat[len(flat)-1]
				worker := r.Compilation.Memory["Worker"]
				if join.Kind != instructionAwait || join.Target == nil || join.Target.Record != worker || show.Kind != instructionShow || show.Value == nil || show.Value.Owner == nil || show.Value.Owner.Record != worker || show.Value.Member.String() != "output" {
					t.Fatalf("generic action failed to retain Worker through Await and possessive: join=%+v show=%+v", join, show)
				}
			})
		}
	})
}

func TestDistributionSourceProvenance(t *testing.T) {
	t.Run("MG05-exact-main-and-physical-projection", func(t *testing.T) {
		const library = "\"value\" is `snow 雪, \\\"quote\\\"`.\r\nShow value.\r\n"
		const source = "Use IO's Print from `io` as \"Print\".\r\n\r\n    Use `nested/words.l`.\r\n\r\n    Use `nested/words.l`.\r\n"
		r := requireDistributionFrontend(t, source, map[string]string{"nested/words.l": library})
		c := r.Compilation
		if !bytes.Equal(c.MainBytes, []byte(source)) {
			t.Fatal("MainBytes changed CRLF or blank lines")
		}
		root := filepath.Dir(c.MainPath)
		physical := testPhysicalStatements(root, c.SourceStatements)
		if len(physical) != 5 {
			t.Fatalf("user/include statements = %d; want import alias plus two include occurrences", len(physical))
		}
		if physical[1].occurrence == physical[3].occurrence || physical[1].Span != physical[3].Span || !bytes.Equal(physical[1].Raw, physical[3].Raw) {
			t.Fatal("repeated include lost occurrence identity or physical origin")
		}
		if string(physical[1].Raw) != strings.SplitAfter(library, "\r\n")[0] || physical[1].Span.Start.Line != 1 || physical[2].Span.Start.Line != 2 {
			t.Fatal("local literal bytes/line spans changed")
		}
		want := "Use IO's Print as \"Print\".\n\"value\" is `snow 雪, \\\"quote\\\"`.\nShow value.\n\"value\" is `snow 雪, \\\"quote\\\"`.\nShow value."
		if got := string(renderExpandedSource(physical, c.TextCache)); got != want {
			t.Fatalf("physical expanded projection = %q; want %q", got, want)
		}
		var implicit, dictionary bool
		for _, statement := range c.SourceStatements {
			if testPathWithin(root, statement.Span.Path) {
				continue
			}
			if filepath.IsAbs(statement.Span.Path) {
				t.Errorf("embedded source became a host path: %s", statement.Span.Path)
			}
			implicit = implicit || strings.Contains(string(statement.Raw), "Create \"Start|")
			dictionary = dictionary || strings.Contains(string(statement.Raw), "Create IO's")
		}
		if !implicit || !dictionary {
			t.Fatal("complete diagnostic source lost builtin or dictionary unit")
		}
		wantPaths := []string{c.MainPath, filepath.Join(root, "nested/words.l"), filepath.Join(root, "nested/words.l")}
		if !reflect.DeepEqual(c.SourcePaths, wantPaths) {
			t.Fatalf("physical input inventory = %v; want %v", c.SourcePaths, wantPaths)
		}
		for _, path := range c.SourcePaths {
			if _, err := os.Stat(path); err != nil {
				t.Errorf("nonphysical input %s: %v", path, err)
			}
		}
	})
}

// Default services require no adjacent configuration or resource files.
func TestDistributionServices(t *testing.T) {
	t.Run("FS05", func(t *testing.T) {
		before, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		project := t.TempDir()
		input := []byte("Use IO's Print from `io` as \"Print\".\nCreate \"Hello\" as a New Basic App.\nThe App Prints `Hello World!`.\nExecute Hello.\nShow Hello's Output.\n")
		main := filepath.Join(project, "main.l")
		if err := os.WriteFile(main, input, 0600); err != nil {
			t.Fatal(err)
		}
		result, err := NewLygerService(Options{}).Compile(input, main)
		if err != nil {
			t.Fatalf("default-resource/frontend prerequisite failed; runtime not exercised: %v", err)
		}
		units, err := result.ApplicationUnits()
		if err != nil || len(units) != 1 {
			t.Fatalf("application units = %v, %v", units, err)
		}
		bundle, err := Logiceme.NewLogicemeService(nil).Evaluate(units[0].BackendAML)
		if err != nil || len(bundle.ByteArrays) != 1 || len(bundle.ByteArrays[0]) == 0 {
			t.Fatalf("same-resource backend prerequisite failed; runtime not exercised: %v", err)
		}
		after, err := os.Getwd()
		if err != nil || after != before {
			t.Fatalf("compilation changed CWD: before=%q after=%q error=%v", before, after, err)
		}
	})
}

func TestFrontendLoadsRelocatedAssetsAndSourceLocalImports(t *testing.T) {
	for _, missing := range []string{"", "grammar.json", "vocabulary.json"} {
		name := missing
		if name == "" {
			name = "complete assets"
		}
		t.Run(name, func(t *testing.T) {
			assets := distributiontest.CompleteOverride(t)
			project := t.TempDir()
			cfg := Options{Language: DefaultLanguage, AssetRoot: assets, Resources: os.DirFS(assets)}
			input := []byte("Use Local's Greeting from `words.l` as \"Greeting\".\nShow Greeting.\n")
			main := filepath.Join(project, "story.l")
			helper := []byte("Create \"Local\" as a Dictionary.\nDefine Local's \"Greeting\" as `local value|`.\n")
			if err := os.WriteFile(filepath.Join(project, "words.l"), helper, 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(main, input, 0600); err != nil {
				t.Fatal(err)
			}
			control, err := NewLygerService(cfg).Compile(input, main)
			if err != nil {
				t.Fatalf("valid relocated-source prerequisite failed; intended assertion not exercised; runtime not exercised: %v", err)
			}
			if missing == "" {
				if control == nil || !strings.Contains(string(control.ExpandedSource), "local value|") {
					t.Fatal("local source inclusion lost its literal")
				}
				return
			}
			path := filepath.Join(assets, "lexicon", cfg.Language, missing)
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
			result, err := NewLygerService(cfg).Compile(input, main)
			if result != nil || !errors.Is(err, os.ErrNotExist) || !strings.Contains(err.Error(), missing) {
				t.Fatalf("missing selected asset lost origin/cause: result=%v error=%v", result, err)
			}
		})
	}
}

// A selected provider is complete; physical provenance cannot select or augment it.
func TestFrontendResourceSelection(t *testing.T) {
	selected := fstest.MapFS{"selected": {Data: []byte("selected\r\n雪")}}
	options := Options{Resources: selected, AssetRoot: t.TempDir()}
	service := NewLygerService(options)
	options.Language = "private mutation"
	options.Resources = fstest.MapFS{}
	content, err := fs.ReadFile(service.options.Resources, "selected")
	if err != nil || !bytes.Equal(content, selected["selected"].Data) {
		t.Fatalf("physical provenance changed resource selection: %q %v", content, err)
	}
	if _, err := fs.ReadFile(service.options.Resources, "lexicon/eng/grammar.json"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("incomplete selected filesystem fell back: %v", err)
	}
	if service.options.Language != DefaultLanguage || service.options.Language == options.Language || reflect.DeepEqual(service.options.Resources, options.Resources) {
		t.Fatal("caller mutation changed the service options")
	}
	fresh := NewLygerService(Options{})
	if fresh.options.Language != DefaultLanguage || fresh.options.AssetRoot != "" {
		t.Fatal("default service shares mutable selection state")
	}
	if _, err := fs.ReadFile(fresh.options.Resources, "lexicon/eng/grammar.json"); err != nil {
		t.Fatalf("default service lost embedded resources: %v", err)
	}
}
