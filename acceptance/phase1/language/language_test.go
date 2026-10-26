//go:build darwin || linux

package language_test

import (
	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	"os"
	"path/filepath"
	"strings"
	"testing"

	phase1 "github.com/l-programming-language/l/internal/testsupport/phase1"
)

// Cases have independent projects, resource copies, provenance recorders and
// child process groups. Bound admission before starting their existing watchdogs.
var nativeCases = make(nativetest.Admission, 4)

// Native evidence complements the frontend/IR checks for these examples. LANG-03
// is a compile-negative ownership case in Lyger; it must never reach native execution.
func TestPhase1LanguageNative(t *testing.T) {
	t.Parallel()
	examples := []struct{ id, output string }{
		{"LANG-01", "launched|continued|child\njoined|again|"},
		{"LANG-02", "passive\n"},
		{"LANG-04", "joined through an alias\n"},
		{"LANG-05", "after-start|continuation|after-await|continuation|after-execute|"},
		{"LANG-06", "canonical\n"},
		{"LANG-07", "both-launched|second-continuation|first-continuation|done|"},
		{"LANG-08", "outer-before-join|worker-before-join|other-continuation|worker-after-join|outer-after-join|"},
		{"LANG-09", "continued-once|"},
		{"LANG-11", "saved\nsaved\n"},
		{"LANG-12", "local\nlocal\nouter|"},
		{"LANG-13", "condition\n"},
		{"LANG-14", "before-await|imported\nfrom-import|imported\nafter-await|"},
		{"LANG-15", "launched|first-continuation|"},
	}
	for _, example := range examples {
		example := example
		t.Run(example.id, func(t *testing.T) {
			nativeCases.Parallel(t)
			data, err := os.ReadFile(filepath.Join(phase1.RepoRoot(t), "testdata/phase1/language", strings.ToLower(example.id)+".l"))
			if err != nil {
				t.Fatal(err)
			}
			if example.id == "LANG-14" {
				phase1.RunSource(t, string(data), example.output, filepath.Join(phase1.RepoRoot(t), "testdata/phase1/language/continuation.l"))
			} else {
				phase1.RunSource(t, string(data), example.output)
			}
		})
	}
	t.Run("LANG-10", func(t *testing.T) {
		nativeCases.Parallel(t)
		data, err := os.ReadFile(filepath.Join(phase1.RepoRoot(t), "testdata/phase1/language/lang-10.l"))
		if err != nil {
			t.Fatal(err)
		}
		phase1.RunSourceOutcomes(t, string(data), []string{
			"stop-requested|continuation|error|joined|",
			"stop-requested|continuation|success|joined|",
		})
	})
}

func TestPhase1ContinuationVariants(t *testing.T) {
	t.Parallel()
	const ioImport = "Use IO's Print from `io` as \"Print\".\n"
	const globals = ioImport + "\"firstMessage\" is `first`.\n\"secondMessage\" is `second`.\n"
	cases := []struct{ name, source, output string }{
		{"passive-await-trailing-effects",
			ioImport + "\"message\" is `passive`.\nCreate \"Worker\" as a New Basic App.\nThe App Prints message.\nRun Worker.\nWorker is awaited and then show Worker's Output.\nWorker is run.\nWorker is awaited and then Execute Worker.\nShow Worker's Output.\n", "passive\npassive\n"},
		{"shared-paragraph-distinct-claims",
			globals + "\n    Create \"First\" as a New Public Basic App.\nThe App Prints firstMessage.\nShow `first-tail|`.\nCreate \"Second\" as a New Public Basic App.\nThe App Prints secondMessage.\nShow `second-tail|`.\n\nRun First and Second.\nAwait Second.\nAwait First.\nAwait Second.\n",
			"second-tail|first-tail|"},
		{"shared-paragraph-explicit-result-owner",
			globals + "\n    Create \"First\" as a New Public Basic App.\nThe App Prints firstMessage.\nShow `first-tail|`.\nCreate \"Second\" as a New Public Basic App.\nThe App Prints secondMessage.\nShow `second-tail|`.\nShow First's Output.\n\nRun First and Second.\nAwait First.\nAwait Second.\n",
			"first-tail|first\nsecond-tail|"},
		{"nested-await-preserves-caller-local",
			globals + "\n    Create \"First\" as a New Public Basic App.\nThe App Prints firstMessage.\nDefine \"result\" as First's Output.\nAwait Second.\nShow result.\n\n    Create \"Second\" as a New Public Basic App.\nThe App Prints secondMessage.\nDefine \"result\" as Second's Output.\nShow result.\n\nRun First and Second.\nAwait First.\nRun First and Second.\nAwait First.\n",
			"second\nfirst\nsecond\nfirst\n"},
		{"shared-paragraph-nested-reruns",
			globals + "\n    Create \"First\" as a New Public Basic App.\nThe App Prints firstMessage.\nDefine \"firstResult\" as First's Output.\nAwait Second.\nShow firstResult.\nCreate \"Second\" as a New Public Basic App.\nThe App Prints secondMessage.\nDefine \"secondResult\" as Second's Output.\nPublicly define \"published\" as Second's Output.\nShow secondResult.\n\nRun First and Second.\nAwait First.\nShow published.\nAwait First.\nRun First and Second.\nAwait First.\nShow published.\n",
			"second\nfirst\nsecond\nsecond\nfirst\nsecond\n"},
		{"shared-paragraph-cached-join",
			globals + "\n    Create \"First\" as a New Public Basic App.\nThe App Prints firstMessage.\nDefine \"firstResult\" as First's Output.\nShow `before|`.\nAwait Second.\nShow firstResult.\nCreate \"Second\" as a New Public Basic App.\nThe App Prints secondMessage.\nDefine \"secondResult\" as Second's Output.\nShow secondResult.\n\nRun First and Second.\nAwait Second.\nAwait First.\n",
			"second\nbefore|first\n"},
		{"shared-paragraph-three-level-join",
			globals + "\"thirdMessage\" is `third`.\n\n    Create \"First\" as a New Public Basic App.\nThe App Prints firstMessage.\nDefine \"firstResult\" as First's Output.\nAwait Second.\nShow firstResult.\nCreate \"Second\" as a New Public Basic App.\nThe App Prints secondMessage.\nDefine \"secondResult\" as Second's Output.\nAwait Third.\nShow secondResult.\nCreate \"Third\" as a New Public Basic App.\nThe App Prints thirdMessage.\nDefine \"thirdResult\" as Third's Output.\nShow thirdResult.\n\nRun First and Second and Third.\nAwait First.\n",
			"third\nsecond\nfirst\n"},
		{"shared-paragraph-mutual-join",
			globals + "\n    Create \"First\" as a New Public Basic App.\nThe App Prints firstMessage.\nDefine \"firstResult\" as First's Output.\nAwait Second.\nShow firstResult.\nCreate \"Second\" as a New Public Basic App.\nThe App Prints secondMessage.\nDefine \"secondResult\" as Second's Output.\nAwait First.\nShow secondResult.\n\nRun First and Second.\nAwait First.\nAwait Second.\n",
			"second\nfirst\n"},
		{"shared-paragraph-nested-exit",
			globals + "\n    Create \"First\" as a New Public Basic App.\nThe App Prints firstMessage.\nDefine \"firstResult\" as First's Output.\nAwait Second.\nShow firstResult.\nCreate \"Second\" as a New Public Basic App.\nThe App Prints secondMessage.\nShow `second-exit|`.\nExit.\n\nRun First and Second.\nAwait First.\nShow `unreachable|`.\n",
			"second-exit|"},
		{"mutual-await",
			globals + "\n    Create \"First\" as a New Public Basic App.\nThe App Prints firstMessage.\nAwait Second.\nShow `first-tail|`.\n\n    Create \"Second\" as a New Public Basic App.\nThe App Prints secondMessage.\nAwait First.\nShow `second-tail|`.\n\nRun First and Second.\nAwait First.\nAwait Second.\n",
			"second-tail|first-tail|"},
		{"cached-self-await-through-alias",
			ioImport + "\"message\" is `self`.\n\n    Create \"Worker\" as a New Public Basic App.\nThe App Prints message.\nUse Worker as \"Job\".\nAwait Job.\nShow `once|`.\n\nRun Worker.\nAwait Worker.\nAwait Worker.\n", "once|"},
		{"declared-never-launched",
			ioImport + "\"message\" is `child-must-not-run`.\n\n    Create \"Worker\" as a New Public Basic App.\nThe App Prints message.\nShow `continuation-must-not-run`.\n\nShow `root|`.\n", "root|"},
		{"build-only-no-continuation",
			ioImport + "\"message\" is `child-must-not-run`.\n\n    Create \"Worker\" as a New Public Basic App.\nThe App Prints message.\nShow `continuation-must-not-run`.\n\nBuild Worker.\nShow `root|`.\n", "root|"},
		{"execute-exit-prevents-next-operand",
			globals + "\n    Create \"First\" as a New Public Basic App.\nThe App Prints firstMessage.\nShow `first|`.\nExit.\n\n    Create \"Second\" as a New Public Basic App.\nThe App Prints secondMessage.\nShow `second-must-not-run`.\n\nExecute First and Second.\nShow `tail-must-not-run`.\n", "first|"},
		{"ordinary-scoped-await",
			ioImport + "\"message\" is `ordinary`.\nCreate \"Worker\" as a New Basic App.\nThe App Prints message.\n\n    Run Worker.\nShow `before|`.\nAwait Worker.\nShow Worker's Output.\n\nShow `after|`.\n", "before|ordinary\nafter|"},
	}
	for _, item := range cases {
		item := item
		t.Run(item.name, func(t *testing.T) {
			nativeCases.Parallel(t)
			phase1.RunSource(t, item.source, item.output)
		})
	}
}

func TestMain(m *testing.M) { phase1.Main(m) }
