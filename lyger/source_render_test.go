package Lyger

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestExpandedSourcePreservesDefinitionMeaningAndOrder(t *testing.T) {
	// Each expected line is specified independently from the display renderer.
	// Vary vocabulary, type annotations and declaration order; expansion does
	// not infer a semantic rewrite or sort declarations by their names.
	for _, tc := range []struct {
		name   string
		source []string
		want   []string
	}{
		{
			name: "typed definitions followed by descriptor",
			source: []string{
				"Define \"Process\" as `<DEF>W^<@Definitions|Syllables@><DEF>`.",
				"Define \"Function\" as `<DEF>FW{|Func{Key: <@Name|Text@>}Func|}^(<@Process@>)<DEF>`.",
				"\"Objective\" describes any Function as `<DEF>[[\"A\", \"<@Function@>\"]]<DEF>`.",
			},
			want: []string{
				"Define \"Process\" as `W^<@Definitions|Syllables@>`.",
				"Define \"Function\" as `FW{|Func{Key: <@Name|Text@>}Func|}^(<@Process@>)`.",
				"\"Objective\" describes any Function as `[[\"A\", \"<@Function@>\"]]`.",
			},
		},
		{
			name: "renamed descriptor before its target",
			source: []string{
				"\"Exported\" describes any Handler as `<DEF>[[\"A\", \"<@Handler@>\"]]<DEF>`.",
				"Define \"Handler\" as `<DEF>FW{|Func{Key: <@Label|Text@>}Func|}^(<@Sequence@>)<DEF>`.",
				"Define \"Sequence\" as `<DEF>W^<@Operations|Syllables@><DEF>`.",
			},
			want: []string{
				"\"Exported\" describes any Handler as `[[\"A\", \"<@Handler@>\"]]`.",
				"Define \"Handler\" as `FW{|Func{Key: <@Label|Text@>}Func|}^(<@Sequence@>)`.",
				"Define \"Sequence\" as `W^<@Operations|Syllables@>`.",
			},
		},
		{
			name: "retain a different target and parameter kinds",
			source: []string{
				"\"Wrapped\" describes any Definition as `<DEF>[[\"B\", \"<@Definition@>\"]]<DEF>`.",
				"Define \"Message\" as `<DEF>E{|Value: <@Content|Thing@>, Names: <@Labels|Texts@>|}<DEF>`.",
			},
			want: []string{
				"\"Wrapped\" describes any Definition as `[[\"B\", \"<@Definition@>\"]]`.",
				"Define \"Message\" as `E{|Value: <@Content|Thing@>, Names: <@Labels|Texts@>|}`.",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := []byte("*** a draft note ***\n" + strings.Join(tc.source, "\n") + "\n")
			prepared, err := PrepareSource("definitions with spaces.l", source)
			if err != nil {
				t.Fatal(err)
			}
			original := bytes.Clone(prepared.PreparedBytes)
			cache := make(map[string][]byte, len(prepared.TextCache))
			for key, value := range prepared.TextCache {
				cache[key] = bytes.Clone(value)
			}
			want := []byte(strings.Join(tc.want, "\n"))
			for attempt := 0; attempt < 3; attempt++ {
				got := renderExpandedSource(prepared.Statements, prepared.TextCache)
				if !bytes.Equal(got, want) {
					t.Fatalf("expanded source changed meaning, order or EOF:\ngot  %q\nwant %q", got, want)
				}
			}
			if !bytes.Equal(prepared.PreparedBytes, original) || !reflect.DeepEqual(prepared.TextCache, cache) {
				t.Fatal("display rendering mutated semantic input or literal storage")
			}
		})
	}
}

func TestExpandedSourcePreservesImportPlacementAndLiteralEscaping(t *testing.T) {
	root := t.TempDir()
	writeSourceFile(t, root, "definitions.l", "Define \"Sequence\" as `<DEF>W^<@Parts|Syllables@><DEF>`.\n")
	writeSourceFile(t, root, "library.l", "\"Exported\" describes any Handler as `<DEF>[[\"A\", \"<@Handler@>\"]]<DEF>`.\n")
	path := writeSourceFile(t, root, "story.l", "Use `definitions.l`.\n\"first\" is `unchanged \\\"雪\\\"`.\nUse Tools's Render from `library.l` as \"Paint\".\nDefine \"quoted\" to `<DEF>E{|Value: ~\\\"snow 雪\\\"~|}<DEF>`.\n")
	built, err := buildTemporarySource(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := "Define \"Sequence\" as `W^<@Parts|Syllables@>`.\n" +
		"\"first\" is `unchanged \\\"雪\\\"`.\n" +
		"\"Exported\" describes any Handler as `[[\"A\", \"<@Handler@>\"]]`.\n" +
		"Use Tools's Render as \"Paint\".\n" +
		"Define \"quoted\" to `E{|Value: ~\"snow 雪\"~|}`."
	if got := renderExpandedSource(testPhysicalStatements(root, built.SourceStatements), built.TextCache); string(got) != want {
		t.Fatalf("expanded imports or literal boundaries changed:\ngot  %q\nwant %q", got, want)
	}
}
