package Lyger

import (
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestDictionaryCannotCaptureCallerNamespaceMembers(t *testing.T) {
	root := t.TempDir()
	config := configuredImportAssets(t, root, [3]string{"Use", "from", "as"})
	resources, ok := config.Resources.(fstest.MapFS)
	if !ok {
		t.Fatalf("configured import resources have type %T, want fstest.MapFS", config.Resources)
	}
	dictionary := "Create \"IO\" as a Dictionary.\nCreate IO's \"Reveal|-s|-ed|-ing|-er\" for Showing IO's Secret and then Showing any Text.\n"
	resources["stdlib/eng/io/dictionary.l"].Data = []byte(dictionary + "Define IO's \"Secret\" as `owned value`.\n")
	control := "Use IO's Reveal from `io` as \"Reveal\".\nReveal `argument`.\n"
	if _, err := NewLygerService(config).Compile([]byte(control), filepath.Join(root, "control.l")); err != nil {
		t.Fatalf("owned dictionary control failed: %v", err)
	}
	resources["stdlib/eng/io/dictionary.l"].Data = []byte(dictionary)
	source := "Use IO from `io`.\nDefine IO's \"Secret\" as `caller value`.\nUse IO's Reveal as \"Reveal\".\nReveal `ignored`.\n"
	_, err := NewLygerService(config).Compile([]byte(source), filepath.Join(root, "main.l"))
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "cannot modify") {
		t.Fatalf("caller declaration must not satisfy an unresolved dictionary-owned member: %v", err)
	}
}
