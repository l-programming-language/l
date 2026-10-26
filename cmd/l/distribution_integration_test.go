//go:build darwin || linux

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	distributiontest "github.com/l-programming-language/l/internal/testsupport/distribution"
	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	"github.com/l-programming-language/l/lerrors"
	Logiceme "github.com/l-programming-language/l/logiceme"
	Lyger "github.com/l-programming-language/l/lyger"
)

// Explicit disk-resource controls use today's real initialization API. They do
// not implement embedded defaults or supply a replacement builtin/library.
func TestDistributionBundleParity(t *testing.T) {
	parallelIndependentBuildRoot(t)
	t.Run("FS11", func(t *testing.T) {
		root := distributiontest.RepoRoot(t)
		source, err := os.ReadFile(filepath.Join(root, "demo/basic/main.l"))
		if err != nil {
			t.Fatal(err)
		}
		main := filepath.Join(t.TempDir(), "main.l")
		if err := os.WriteFile(main, source, 0600); err != nil {
			t.Fatal(err)
		}
		generate := func(options compilerInitOptions) (*generatedProgram, Logiceme.Executable) {
			c := newCompiler(&cliData{jobType: "build", mainPath: main, osType: "ubuntu"})
			c.initialize(options)
			g := c.process()
			if g != nil {
				nativetest.RecordBundle(t, "compiler-process", g.Sources)
			}
			if c.err != nil || g == nil {
				t.Fatalf("selected-resource generation prerequisite; parity/runtime not exercised: %v", c.err)
			}
			assertFrontendResources(t, g.Compilation.Compilation)
			buildReturnedBundle(t, g.Sources)
			var apps []Lyger.AppArtifact
			for i, id := range g.ApplicationIDs {
				apps = append(apps, Lyger.AppArtifact{ID: id, RelativeExecutable: fmt.Sprintf("program.apps/app-%d", i)})
			}
			unit, err := g.Compilation.OrchestratorUnit("Program", apps)
			if err != nil {
				t.Fatal(err)
			}
			runtime, err := c.log.Evaluate(unit.BackendAML, Logiceme.GenerationOptions{IncludeRuntime: true})
			nativetest.RecordBundle(t, "explicit-runtime", runtime)
			if err != nil {
				t.Fatalf("selected runtime generation prerequisite: %v", err)
			}
			return g, runtime
		}
		embedded, embeddedRuntime := generate(compilerInitOptions{})
		assets := distributiontest.CompleteOverride(t)
		override, overrideRuntime := generate(compilerInitOptions{AssetRoot: assets})
		if !bytes.Equal(embedded.Compilation.AML, override.Compilation.AML) || !bytes.Equal(embedded.Compilation.BackendAML, override.Compilation.BackendAML) || !reflect.DeepEqual(embedded.ApplicationIDs, override.ApplicationIDs) || !reflect.DeepEqual(embedded.Sources, override.Sources) || !reflect.DeepEqual(embeddedRuntime, overrideRuntime) {
			t.Fatal("embedded and complete override resources produced different AML/source bundles")
		}
	})
}

func distributionSource(t testing.TB, assets, source string, files map[string]string) (*Lyger.CompilationResult, error) {
	t.Helper()
	dir := t.TempDir()
	for name, data := range files {
		if !filepath.IsLocal(name) {
			t.Fatalf("invalid test source path %q", name)
		}
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	main := filepath.Join(dir, "main.l")
	if err := os.WriteFile(main, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	options := canonicalCompilerOptions(t)
	if assets != "" {
		options = compilerInitOptions{AssetRoot: assets}
	}
	c := newCompiler(&cliData{jobType: "build", mainPath: main, osType: "ubuntu"})
	c.initialize(options)
	if c.err != nil {
		return nil, c.err
	}
	r := c.compile()
	if c.err != nil {
		return nil, c.err
	}
	units, err := r.ApplicationUnits()
	if err != nil {
		return nil, err
	}
	var companions []Lyger.AppArtifact
	for i, unit := range units {
		companions = append(companions, Lyger.AppArtifact{ID: unit.ID, RelativeExecutable: fmt.Sprintf("app-%d", i)})
	}
	if _, err := r.OrchestratorUnit("Distribution", companions); err != nil {
		return nil, err
	}
	return r, nil
}

func requireDistributionSource(t testing.TB, assets, source string, files map[string]string) *Lyger.CompilationResult {
	t.Helper()
	r, err := distributionSource(t, assets, source, files)
	if err != nil {
		t.Fatalf("dictionary/builtin compilation prerequisite failed; downstream assertion and runtime not exercised: %v", err)
	}
	assertFrontendResources(t, r.Compilation)
	return r
}

type distributionEntry struct {
	Entry      string `json:"entry"`
	Dictionary string `json:"dictionary"`
}

type distributionIndex struct {
	Format    int                                     `json:"format"`
	Languages map[string]map[string]distributionEntry `json:"languages"`
}

func addDistributionDictionary(t testing.TB, assets, key, dictionary, source string) string {
	t.Helper()
	path := filepath.Join(assets, "stdlib/index.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var index distributionIndex
	if err := json.Unmarshal(data, &index); err != nil {
		t.Fatal(err)
	}
	if index.Format != 1 || index.Languages["eng"] == nil {
		t.Fatal("production dictionary index has no supported English registry")
	}
	entry := "stdlib/eng/" + key + "/dictionary.l"
	index.Languages["eng"][key] = distributionEntry{Entry: entry, Dictionary: dictionary}
	encoded, err := json.Marshal(index)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, encoded, 0600); err != nil {
		t.Fatal(err)
	}
	destination := filepath.Join(assets, filepath.FromSlash(entry))
	if err := os.MkdirAll(filepath.Dir(destination), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(destination, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	return destination
}

const distributionTools = "Create \"Tools\" as a Dictionary.\nDefine Tools's \"Greeting\" as `hello`.\n\"secret\" is `private`.\nCreate Tools's \"Display|-s|-ed|-ing|-er\" for Showing any Text.\n"

const distributionGreeting = "Use Tools's Greeting from `tools` as \"Greeting\".\nShow Greeting.\n"

func TestDistributionDictionarySemantics(t *testing.T) {
	parallelIndependentBuildRoot(t)
	t.Run("DI02", func(t *testing.T) {
		for _, tc := range []struct{ name, target, member string }{
			{"unknown", "absent-dictionary", "Greeting"}, {"traversing", "tools/../tools", "Greeting"},
			{"missing export", "tools", "Absent"}, {"private export", "tools", "secret"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				assets := distributiontest.CompleteOverride(t)
				addDistributionDictionary(t, assets, "tools", "Tools", distributionTools)
				requireDistributionSource(t, assets, distributionGreeting, nil)
				source := fmt.Sprintf("Use Tools's %s from `%s` as \"Selected\".\nShow `unused`.\n", tc.member, tc.target)
				_, err := distributionSource(t, assets, source, nil)
				if err == nil || (!strings.Contains(err.Error(), tc.target) && !strings.Contains(err.Error(), tc.member)) || !strings.Contains(err.Error(), "main.l:1:") {
					t.Fatalf("selected import must reject its own target at the import site: %v", err)
				}
			})
		}
		for _, mode := range []string{"missing entry", "wrong dictionary"} {
			t.Run(mode, func(t *testing.T) {
				assets := distributiontest.CompleteOverride(t)
				entry := addDistributionDictionary(t, assets, "tools", "Tools", distributionTools)
				requireDistributionSource(t, assets, distributionGreeting, nil)
				if mode == "missing entry" {
					if err := os.Remove(entry); err != nil {
						t.Fatal(err)
					}
				} else {
					if err := os.WriteFile(entry, []byte("Create \"OtherTools\" as a Dictionary.\n"), 0600); err != nil {
						t.Fatal(err)
					}
				}
				if _, err := distributionSource(t, assets, distributionGreeting, nil); err == nil || !strings.Contains(err.Error(), "tools") {
					t.Fatalf("invalid dictionary entry accepted or wrong prerequisite failed: %v", err)
				}
			})
		}
	})
	t.Run("DI03", func(t *testing.T) {
		assets := distributiontest.CompleteOverride(t)
		addDistributionDictionary(t, assets, "tools", "Tools", distributionTools)
		selected := requireDistributionSource(t, assets, distributionGreeting, nil)
		if selected.Compilation.Memory["Tools"] != nil || selected.Compilation.Memory["secret"] != nil {
			t.Fatal("selected member import leaked namespace/private records into the program")
		}
		namespace := requireDistributionSource(t, assets, "Use Tools from `tools`.\nShow Tools's Greeting.\n", nil)
		if namespace.Compilation.Memory["Greeting"] != nil {
			t.Fatal("namespace import leaked unqualified Greeting")
		}
		for _, source := range []string{
			"Use Tools from `tools`.\nShow Greeting.\n",
			"Use Tools's Greeting from `tools` as \"Greeting\".\nShow Tools's Greeting.\n",
		} {
			if _, err := distributionSource(t, assets, source, nil); err == nil {
				t.Fatalf("unimported name was visible: %s", source)
			}
		}
	})
	t.Run("DI04", func(t *testing.T) {
		for _, alias := range []string{"", " as \"Unused\""} {
			t.Run(alias, func(t *testing.T) {
				assets := distributiontest.CompleteOverride(t)
				addDistributionDictionary(t, assets, "tools", "Tools", distributionTools)
				requireDistributionSource(t, assets, "Use Tools's Greeting from `tools`"+alias+".\nShow `control`.\n", nil)
				_, err := distributionSource(t, assets, "Use Tools's Absent from `tools`"+alias+".\nShow `control`.\n", nil)
				if err == nil || !strings.Contains(strings.ToLower(err.Error()), "absent") || !strings.Contains(err.Error(), "main.l:1:") {
					t.Fatalf("unused missing export must be validated at the import: %v", err)
				}
			})
		}
	})
	t.Run("DI05", func(t *testing.T) {
		assets := distributiontest.CompleteOverride(t)
		addDistributionDictionary(t, assets, "shared", "SharedTools", "Create \"SharedTools\" as a Dictionary.\nDefine SharedTools's \"Greeting\" as `shared`.\n")
		for _, name := range []string{"LeftTools", "RightTools"} {
			addDistributionDictionary(t, assets, strings.ToLower(name), name, "Use SharedTools's Greeting from `shared` as \"Value\".\nCreate \""+name+"\" as a Dictionary.\nUse Value as "+name+"'s \"Greeting\".\n")
		}
		source := "Use LeftTools's Greeting from `lefttools` as \"First\".\nUse RightTools's Greeting from `righttools` as \"Second\".\nUse SharedTools's Greeting from `shared` as \"Third\".\nShow First and Second and Third.\n"
		r := requireDistributionSource(t, assets, source, nil)
		canonical := func(name string) any {
			rec := r.Compilation.Memory[name]
			if rec == nil {
				t.Fatalf("missing imported alias %s", name)
			}
			for steps := 0; r.Compilation.Semantics.Aliases[rec] != nil; steps++ {
				if steps > len(r.Compilation.Semantics.Aliases) {
					t.Fatal("alias cycle")
				}
				rec = r.Compilation.Semantics.Aliases[rec]
			}
			return rec
		}
		if canonical("First") != canonical("Second") || canonical("First") != canonical("Third") {
			t.Fatal("diamond imports copied the canonical member")
		}
		declarations, importSites := 0, 0
		for _, s := range r.Compilation.SourceStatements {
			if bytes.Contains(s.Raw, []byte(`Create "SharedTools"`)) {
				declarations++
			}
			if strings.HasSuffix(s.Span.Path, "main.l") && s.Span.Start.Line <= 3 {
				importSites++
			}
		}
		if declarations != 1 || importSites != 3 {
			t.Fatalf("canonical declarations=%d/import edges=%d; want1/3", declarations, importSites)
		}
	})
	t.Run("DI06", func(t *testing.T) {
		assets := distributiontest.CompleteOverride(t)
		addDistributionDictionary(t, assets, "tools", "Tools", distributionTools)
		requireDistributionSource(t, assets, "Use Tools's Display from `tools` as \"Paint\".\nUse Tools's Display from `tools` as \"Paint\".\nPaint `same`.\n", nil)
		for _, conflicting := range []string{"Paint", "Paints"} {
			t.Run(conflicting, func(t *testing.T) {
				source := "Use Tools's Display from `tools` as \"Paint\".\nUse Tools's Greeting from `tools` as \"" + conflicting + "\".\nShow `unused`.\n"
				_, err := distributionSource(t, assets, source, nil)
				if err == nil || !strings.Contains(strings.ToLower(err.Error()), "paint") || !strings.Contains(err.Error(), ":1:") || !strings.Contains(err.Error(), ":2:") {
					t.Fatalf("conflicting inflection must report both sites: %v", err)
				}
			})
		}
	})
	t.Run("DI07", func(t *testing.T) {
		assets := distributiontest.CompleteOverride(t)
		addDistributionDictionary(t, assets, "tools", "Tools", distributionTools+"Publicly define \"published\" as `private unit value`.\n")
		r := requireDistributionSource(t, assets, "\"secret\" is `caller`.\n"+distributionGreeting, nil)
		if r.Compilation.Memory["published"] != nil {
			t.Fatal("module Public escaped into the program root")
		}
		for _, name := range []string{"published", "secret"} {
			_, err := distributionSource(t, assets, distributionGreeting+"Show "+name+".\n", nil)
			if err == nil || !strings.Contains(strings.ToLower(err.Error()), name) {
				t.Fatalf("private %s escaped: %v", name, err)
			}
		}
		// A definition referencing an absent module local must not capture a caller local.
		addDistributionDictionary(t, assets, "capture", "CaptureTools", "Create \"CaptureTools\" as a Dictionary.\nCreate CaptureTools's \"Reveal|-s|-ed|-ing|-er\" for Showing stolen.\n")
		_, err := distributionSource(t, assets, "\"stolen\" is `caller`.\nUse CaptureTools's Reveal from `capture` as \"Reveal\".\nReveal.\n", nil)
		if err == nil || !strings.Contains(strings.ToLower(err.Error()), "stolen") {
			t.Fatalf("dictionary definition captured caller-local value: %v", err)
		}
	})
	t.Run("DI09", func(t *testing.T) {
		assets := distributiontest.CompleteOverride(t)
		addDistributionDictionary(t, assets, "tools", "Tools", distributionTools)
		requireDistributionSource(t, assets, distributionGreeting, nil)
		addDistributionDictionary(t, assets, "cyclea", "CycleA", "Use `cycleb`.\nCreate \"CycleA\" as a Dictionary.\n")
		addDistributionDictionary(t, assets, "cycleb", "CycleB", "Use `cyclea`.\nCreate \"CycleB\" as a Dictionary.\n")
		_, err := distributionSource(t, assets, "Use `cyclea`.\nShow `unused`.\n", nil)
		if !errors.Is(err, lerrors.Cycle) || !strings.Contains(err.Error(), "cyclea") || !strings.Contains(err.Error(), "cycleb") {
			t.Fatalf("lost dictionary cycle chain: %v", err)
		}
		requireDistributionSource(t, assets, distributionGreeting, nil)
	})
	t.Run("DI10", func(t *testing.T) {
		for _, effect := range []string{"Show `side effect`.\n", "Exit.\n", "Create \"Worker\" as a New Basic App.\n", "Create \"Worker\" as a New Basic App.\nRun Worker.\n", "Use IO's Print from `io` as \"Print\".\nCreate \"Worker\" as a New Basic App.\nThe App Prints `attached`.\n"} {
			t.Run(strings.TrimSpace(effect), func(t *testing.T) {
				assets := distributiontest.CompleteOverride(t)
				entry := addDistributionDictionary(t, assets, "tools", "Tools", distributionTools)
				requireDistributionSource(t, assets, distributionGreeting, nil)
				if err := os.WriteFile(entry, []byte(distributionTools+effect), 0600); err != nil {
					t.Fatal(err)
				}
				_, err := distributionSource(t, assets, distributionGreeting, nil)
				if err == nil || !strings.Contains(strings.ToLower(err.Error()), "dictionary") {
					t.Fatalf("module initialization was accepted or rejected for unrelated reason: %v", err)
				}
			})
		}
		// Ordinary source inclusions retain their original executable meaning.
		requireDistributionSource(t, "", "Use `effects.l`.\nShow `after`.\n", map[string]string{"effects.l": "Show `before`.\n"})
	})
	t.Run("DI13", func(t *testing.T) {
		assets := distributiontest.CompleteOverride(t)
		entry := addDistributionDictionary(t, assets, "custom", "CustomTools", "Use `helpers.l`.\nCreate \"CustomTools\" as a Dictionary.\nUse helper as CustomTools's \"Greeting\".\n")
		if err := os.WriteFile(filepath.Join(filepath.Dir(entry), "helpers.l"), []byte("\"helper\" is `nested helper`.\n"), 0600); err != nil {
			t.Fatal(err)
		}
		source := "Use CustomTools's Greeting from `custom` as \"Greeting\".\nShow Greeting.\n"
		r := requireDistributionSource(t, assets, source, map[string]string{"helpers.l": "\"helper\" is `wrong CWD helper`.\n"})
		if r.Compilation.Memory["helper"] != nil {
			t.Fatal("nested resource helper escaped its dictionary")
		}
		if !bytes.Contains(r.ExpandedSource, []byte("nested helper")) || bytes.Contains(r.ExpandedSource, []byte("wrong CWD helper")) {
			t.Fatal("relative resource include used host/CWD file")
		}
		if err := os.WriteFile(entry, []byte("Use `../../../../outside.l`.\nCreate \"CustomTools\" as a Dictionary.\n"), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := distributionSource(t, assets, "Use `custom`.\n", nil); err == nil || !strings.Contains(err.Error(), "outside.l") {
			t.Fatalf("resource-root escape accepted: %v", err)
		}
	})
	t.Run("DI14", func(t *testing.T) {
		requireDistributionSource(t, "", "Show `control`.\n", nil)
		_, err := distributionSource(t, "", "Use `example.org/team/tools`.\nShow `unused`.\n", nil)
		if !errors.Is(err, lerrors.Unsupported) || !strings.Contains(err.Error(), "example.org/team/tools") {
			t.Fatalf("unconfigured provider must report unsupported dependency: %v", err)
		}
	})
}
