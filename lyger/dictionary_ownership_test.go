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
	installTestDictionary(t, resources, "catalog", "Catalog", "Create \"Catalog\" as a Dictionary.\n")
	dictionary := "Create \"Catalog\" as a Dictionary.\nCreate Catalog's \"Reveal|-s|-ed|-ing|-er\" for Showing Catalog's Secret and then Showing any Text.\n"
	installTestDictionary(t, resources, "catalog", "Catalog", dictionary+"Define Catalog's \"Secret\" as `owned value`.\n")
	control := "Use Catalog's Reveal from `catalog` as \"Reveal\".\nReveal `argument`.\n"
	if _, err := NewLygerService(config).Compile([]byte(control), filepath.Join(root, "control.l")); err != nil {
		t.Fatalf("owned dictionary control failed: %v", err)
	}
	resources["stdlib/eng/catalog/dictionary.l"].Data = []byte(dictionary)
	source := "Use Catalog from `catalog`.\nDefine Catalog's \"Secret\" as `caller value`.\nUse Catalog's Reveal as \"Reveal\".\nReveal `ignored`.\n"
	_, err := NewLygerService(config).Compile([]byte(source), filepath.Join(root, "main.l"))
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "cannot modify") {
		t.Fatalf("caller declaration must not satisfy an unresolved dictionary-owned member: %v", err)
	}
}
