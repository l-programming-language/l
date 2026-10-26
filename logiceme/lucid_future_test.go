package Logiceme

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"reflect"
	"strings"
	"testing"

	distribution "github.com/l-programming-language/l"
	"github.com/l-programming-language/l/lerrors"
)

// These documents exercise the public framed-AML boundary. Complete expected
// bundles and native ownership observations are separate fixture witnesses.
func lucidFutureDocument(items ...string) []byte {
	var content strings.Builder
	for _, item := range append([]string{"A"}, items...) {
		content.WriteString("<|S|>`" + item + "`</S|>\n")
	}
	return []byte(`<PROGRAM><STRUCTURE><BOOK "future_probe"><OUTLINE>
<JSON>{"AppMeta":{"AppGenDir":"/src"}}</JSON>
<JSON>{"Services":[{"Name":"main","Location":"/src/","GenericObjects":["main"]}]}</JSON>
<JSON>{"Objects":[{"Name":"main","Type":"root"}]}</JSON>
</OUTLINE><SETTINGS></SETTINGS><CONTENT><CHAPTER "main"><SENTENCE "main"><|W|>` + content.String() +
		`</W|></SENTENCE></CHAPTER></CONTENT></BOOK></STRUCTURE></PROGRAM>`)
}

func lucidFutureMain(body string) string { return "FW{|Key: ~main~|}^(W^(" + body + "))" }
func lucidFutureBody(body string) string { return "CV^(W^(" + body + "))" }
func lucidFutureAwait(body string) string {
	return lucidFutureMain("BOC^(" + lucidFutureBody("BOC^("+body+")") + ")")
}

func TestLucidFutureBoundaries(t *testing.T) {
	for _, c := range []struct{ name, source, want string }{
		{"empty body", lucidFutureMain("COB{|Key: ~pending~|}^(CV^W)"), "async {"},
		{"explicit inference", lucidFutureMain("COB{|Key: ~pending~|}^(CV{|Capture: ~Infer~|}^W)"), "async {"},
		{"explicit move", lucidFutureMain("COB{|Key: ~pending~|}^(CV{|Capture: ~Move~|}^W)"), "async move {"},
		{"raw binding", lucidFutureMain("COB{|Key: ~r#type~|}^(CV^W)"), "let r#type ="},
		{"discard binding", lucidFutureMain("COB{|Keys: ~&[_]&~|}^(CV^W)"), "let _ ="},
		{"single plural key", lucidFutureMain("COB{|Keys: ~&[pending]&~|}^(CV^W)"), "let pending ="},
		{"assignment", lucidFutureMain("CO{|Key: ~pending~|}^(CV^W)"), "pending ="},
		{"native future", lucidFutureAwait("DV{|Key: ~std::future::ready(7)~|}"), "(std::future::ready(7)).await"},
		{"literal Void", lucidFutureAwait("DV{|Key: ~Void~|}"), "(Void).await"},
		{"calculation operand", lucidFutureAwait("DV^E{|Statement: ~std::future::ready(7)~|}"), "std::future::ready(7)"},
		{"invocation operand", lucidFutureAwait("DV^KY{|Key: ~std::future::ready~, Inputs: ~&[7]&~|}"), "std::future::ready(7)"},
		{"future operand", lucidFutureAwait("DV^(CV^W^BOD^E{|Statement: ~7~|})"), "7"},
		{"awaited future operand", lucidFutureAwait("DV^DV{|Key: ~std::future::ready(std::future::ready(7))~|}"), "std::future::ready(std::future::ready(7))"},
		{"await binding", lucidFutureMain("BOC^(" + lucidFutureBody("COB{|Key: ~answer~|}^DV{|Key: ~std::future::ready(7)~|}") + ")"), "let answer ="},
		{"await assignment", lucidFutureMain("BOC^(" + lucidFutureBody("CO{|Key: ~answer~|}^DV{|Key: ~std::future::ready(7)~|}") + ")"), "answer ="},
	} {
		t.Run(c.name, func(t *testing.T) {
			input := lucidFutureDocument(c.source)
			before := bytes.Clone(input)
			bundle, err := NewLogicemeService(runtimeGenerationResources(t)).Evaluate(input)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(before, input) {
				t.Fatal("generation mutated AML")
			}
			if len(bundle.ByteArrays) != 1 || len(bundle.ByteArrays[0]) != 5 {
				t.Fatal("incomplete generated bundle")
			}
			source := string(bundle.ByteArrays[0][0])
			if !strings.Contains(source, c.want) {
				t.Fatalf("missing %q in %s", c.want, source)
			}
			if strings.Contains(source, "tokio::spawn") || strings.Contains(source, "async fn") {
				t.Fatalf("future construction changed scheduling or Function definition: %s", source)
			}
			if strings.Contains(c.source, "std::future::ready") && strings.Count(source, "std::future::ready(7)") != 1 {
				t.Fatal("future operand was omitted or duplicated")
			}
			if c.name == "awaited future operand" && strings.Count(source, ".await") != 2 {
				t.Fatal("chained await must evaluate both future layers")
			}
		})
	}
}

func TestLucidFutureValidation(t *testing.T) {
	tests := []struct{ name, source, cause string }{
		{"CV file scope", "CV^W", "Rust CV: execution is invalid at file scope"},
		{"CV absent body", lucidFutureMain("BOC^CV"), "Rust CV: future construction requires exactly one W body"},
		{"CV two bodies", lucidFutureMain("BOC^CV^(W*W)"), "Rust CV: future construction requires exactly one W body"},
		{"CV wrong child", lucidFutureMain("BOC^CV^E{|Statement: ~7~|}"), "Rust CV: future construction requires exactly one W body"},
		{"CV function child", lucidFutureMain("BOC^CV^FW{|Key: ~nested~|}^W"), "Rust CV: future construction requires exactly one W body"},
		{"CV extra child", lucidFutureMain("BOC^CV^(W*E{|Statement: ~7~|})"), "Rust CV: future construction requires exactly one W body"},
		{"CV blank capture", lucidFutureMain("BOC^CV{|Capture: ~ ~|}^W"), "Capture must be Infer or Move"},
		{"CV unknown capture", lucidFutureMain("BOC^CV{|Capture: ~Clone~|}^W"), "Capture must be Infer or Move"},
		{"CV native body", lucidFutureMain("BOC^CV{|Statement: ~7~|}^W"), "unsupported metadata Statement"},
		{"CV lifetime input", lucidFutureMain("BOC^CV{|Lifetimes: ~'a~|}^W"), "unsupported metadata Lifetimes"},
		{"CV generic input", lucidFutureMain("BOC^CV{|GenericTypes: ~T~|}^W"), "unsupported metadata GenericTypes"},
		{"CV compiler projection input", lucidFutureMain("BOC^CV{|CaptureKeyword: ~move~|}^W"), "unsupported metadata CaptureKeyword"},
		{"DV file scope", "DV{|Key: ~pending~|}", "Rust DV: execution is invalid at file scope"},
		{"DV synchronous function", lucidFutureMain("BOC^DV{|Key: ~pending~|}"), "Rust DV: await requires async context"},
		{"DV absent operand", lucidFutureAwait("DV"), "Rust DV: await requires exactly one future operand"},
		{"DV blank key", lucidFutureAwait("DV{|Key: ~ ~|}"), "Rust DV: future Key must be nonempty"},
		{"DV key and operand", lucidFutureAwait("DV{|Key: ~pending~|}^CV^W"), "Rust DV: native and composed future operands are exclusive"},
		{"DV blank key and operand", lucidFutureAwait("DV{|Key: ~ ~|}^CV^W"), "Rust DV: future Key must be nonempty"},
		{"DV two operands", lucidFutureAwait("DV^(CV^W*CV^W)"), "Rust DV: await requires exactly one future operand"},
		{"DV statement operand", lucidFutureAwait("DV^V{|Key: ~worker()~|}"), "Rust DV: await requires a future expression"},
		{"DV composed scheduling", lucidFutureAwait("DV^MV^KY{|Key: ~worker~|}"), "Rust DV: await requires a future expression"},
		{"DV item operand", lucidFutureAwait("DV^HY{|Key: ~Item~|}"), "Rust DV: await requires a future expression"},
		{"DV wrong category", lucidFutureAwait("DV^W"), "Rust DV: await requires a future expression"},
		{"DV compiler projection input", lucidFutureAwait("DV{|FutureKey: ~pending~|}"), "unsupported metadata FutureKey"},
		{"DV sibling context", lucidFutureMain("COB{|Key: ~pending~|}^(CV^W)*BOC^DV{|Key: ~pending~|}"), "Rust DV: await requires async context"},
		{"DV nested function", lucidFutureMain("BOC^(" + lucidFutureBody("BOK^FW{|Key: ~nested~|}^W^BOC^DV{|Key: ~pending~|}") + ")"), "Rust DV: await requires async context"},
		{"DV nested closure", lucidFutureMain("BOC^(" + lucidFutureBody("KOB{|Key: ~nested~|}^BW^W^BOC^DV{|Key: ~pending~|}") + ")"), "Rust DV: await requires async context"},
	}
	for _, form := range []string{"V", "MV"} {
		for _, child := range []string{"CV^W", "DV{|Key: ~pending~|}"} {
			tests = append(tests, struct{ name, source, cause string }{
				form + " direct " + child, lucidFutureAwait(form + "^(" + child + ")"),
				"direct CV/DV values are unsupported",
			})
		}
	}
	for _, form := range []string{"CO", "COB"} {
		for _, c := range []struct{ name, source, cause string }{
			{"file scope", form + "{|Key: ~value~|}^(CV^W)", "execution binding requires callable context"},
			{"missing key", lucidFutureMain(form + "^(CV^W)"), "execution binding requires exactly one simple name"},
			{"blank key", lucidFutureMain(form + "{|Key: ~ ~|}^(CV^W)"), "execution binding requires exactly one simple name"},
			{"two keys", lucidFutureMain(form + "{|Keys: ~&[first, second]&~|}^(CV^W)"), "execution binding requires exactly one simple name"},
			{"key and keys", lucidFutureMain(form + "{|Key: ~first~, Keys: ~&[first]&~|}^(CV^W)"), "Key and Keys metadata are exclusive"},
			{"pattern key", lucidFutureMain(form + "{|Key: ~(first, second)~|}^(CV^W)"), "execution binding requires exactly one simple name"},
			{"absent value", lucidFutureMain(form + "{|Key: ~value~|}"), "execution binding requires exactly one CV or DV value"},
			{"two values", lucidFutureMain(form + "{|Key: ~value~|}^(CV^W*CV^W)"), "execution binding requires exactly one CV or DV value"},
			{"scheduling value", lucidFutureMain(form + "{|Key: ~value~|}^V{|Key: ~worker()~|}"), "execution binding requires exactly one CV or DV value"},
			{"composed scheduling value", lucidFutureMain(form + "{|Key: ~value~|}^MV^KY{|Key: ~worker~|}"), "execution binding requires exactly one CV or DV value"},
			{"native value", lucidFutureMain(form + "{|Key: ~value~, Value: ~worker()~|}^(CV^W)"), "unsupported metadata Value"},
			{"type input", lucidFutureMain(form + "{|Key: ~value~, Type: ~u8~|}^(CV^W)"), "unsupported metadata Type"},
			{"mutation input", lucidFutureMain(form + "{|Key: ~value~, Mutable: ~True~|}^(CV^W)"), "Mutable only applies"},
		} {
			tests = append(tests, struct{ name, source, cause string }{form + " " + c.name, c.source, c.cause})
		}
	}
	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			input := lucidFutureDocument(c.source)
			before := bytes.Clone(input)
			bundle, err := NewLogicemeService(runtimeGenerationResources(t)).Evaluate(input)
			if !bytes.Equal(before, input) {
				t.Fatal("rejection mutated AML")
			}
			if !reflect.DeepEqual(bundle, Executable{}) {
				t.Fatal("rejection returned a partial bundle")
			}
			if !errors.Is(err, lerrors.InvalidInput) || !strings.Contains(err.Error(), c.cause) {
				t.Fatalf("wanted InvalidInput containing %q; got %v", c.cause, err)
			}
		})
	}
	for _, form := range []string{"CV", "DV", "CO", "COB"} {
		t.Run(form+" rejects Deferred", func(t *testing.T) {
			_, err := NewLogicemeService(runtimeGenerationResources(t)).Evaluate(lucidFutureDocument(form + "{|Deferred: ~False~|}"))
			if !errors.Is(err, lerrors.Unsupported) || !strings.Contains(err.Error(), "Deferred is unsupported") {
				t.Fatalf("Deferred accepted or misclassified: %v", err)
			}
		})
	}
}

func TestLucidFutureLocalInputs(t *testing.T) {
	for _, local := range []struct{ source, want string }{
		{"CV^W", "async {"},
		{"CV{|Capture: ~Infer~|}^W", "async {"},
		{"CV{|Capture: ~Move~|}^W", "async move {"},
		{"DV^E{|Statement: ~local_future~|}", "local_future"},
		{"DV{|Key: ~Void~|}", "(Void).await"},
	} {
		t.Run(local.source, func(t *testing.T) {
			ss, err := loadSegmentalSystem(runtimeGenerationResources(t))
			if err != nil {
				t.Fatal(err)
			}
			ss.fileMeta.SetAdd("Capture", "Move", "String")
			ss.fileMeta.SetAdd("CaptureKeyword", "forged_capture", "String")
			ss.fileMeta.SetAdd("Key", "inherited_future", "String")
			ss.fileMeta.SetAdd("FutureKey", "forged_future", "String")
			before := lucidKeyedSnapshot(t, ss.fileMeta)
			g, err := ss.initializeGlyphChecked([]byte(local.source), nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			if err := prepareGlyph(&g, glyphContext{callable: true, async: true, role: "expression"}, &generatedNames{used: local.source, counts: map[string]int{}}); err != nil {
				t.Fatal(err)
			}
			if err := g.compilePreparedGlyph(); err != nil {
				t.Fatal(err)
			}
			output := string(g.Content)
			if !strings.Contains(output, local.want) || strings.Contains(output, "inherited_future") || strings.Contains(output, "forged_") {
				t.Fatalf("local input lost provenance: %s", output)
			}
			if !strings.Contains(local.source, "Move") && strings.Contains(output, "async move") {
				t.Fatal("inherited capture policy changed the future")
			}
			if !bytes.Equal(before, lucidKeyedSnapshot(t, ss.fileMeta)) {
				t.Fatal("preparation mutated file metadata")
			}
		})
	}
}

func TestLucidFutureProviderIsolation(t *testing.T) {
	assets := []string{"builtins/segments/consonants/execute.lu", "builtins/segments/vowels/execution.lu", "builtins/segments/vowels/definition.lu"}
	documents := [][]byte{
		lucidFutureDocument(lucidFutureMain("COB{|Key: ~pending~|}^(CV{|Capture: ~Move~|}^W)")),
		lucidFutureDocument(lucidFutureMain("COB{|Key: ~pending~|}^(CV^W)")),
	}
	var expected []Executable
	for _, provider := range []struct {
		name      string
		resources fs.FS
	}{{"filesystem", runtimeGenerationResources(t)}, {"embedded", distribution.Files()}} {
		t.Run(provider.name, func(t *testing.T) {
			resources := &lucidRecordingFS{FS: provider.resources, reads: map[string]bool{}}
			service := NewLogicemeService(resources)
			var first Executable
			var firstBytes []byte
			for i, index := range []int{0, 1, 0} {
				input := bytes.Clone(documents[index])
				bundle, err := service.Evaluate(input)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(input, documents[index]) {
					t.Fatal("generation mutated AML")
				}
				if provider.name == "filesystem" {
					expected = append(expected, bundle)
				} else if len(expected) != 3 || !reflect.DeepEqual(bundle, expected[i]) {
					t.Fatal("resource providers generated different bundles")
				}
				if i == 0 {
					first = bundle
					firstBytes, err = json.Marshal(first)
					if err != nil {
						t.Fatal(err)
					}
				} else {
					current, err := json.Marshal(first)
					if err != nil || !bytes.Equal(firstBytes, current) {
						t.Fatal("service reuse mutated an earlier bundle")
					}
					if i == 2 && !reflect.DeepEqual(first, bundle) {
						t.Fatal("A/B/A generation changed output")
					}
				}
			}
			for _, asset := range assets {
				if !resources.reads[asset] {
					t.Errorf("selected resource %s was not read", asset)
				}
			}
		})
	}
	for _, asset := range assets {
		t.Run("missing "+asset, func(t *testing.T) {
			resources := lucidFutureMissingFS{FS: runtimeGenerationResources(t), missing: asset}
			input := bytes.Clone(documents[0])
			bundle, err := NewLogicemeService(resources).Evaluate(input)
			if !errors.Is(err, fs.ErrNotExist) || !strings.Contains(err.Error(), asset) {
				t.Fatalf("missing selected resource did not fail with its cause: %v", err)
			}
			if !reflect.DeepEqual(bundle, Executable{}) || !bytes.Equal(input, documents[0]) {
				t.Fatal("missing resource returned a bundle or mutated AML")
			}
		})
	}
}

type lucidFutureMissingFS struct {
	fs.FS
	missing string
}

func (r lucidFutureMissingFS) Open(name string) (fs.File, error) {
	if name == r.missing {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	return r.FS.Open(name)
}
