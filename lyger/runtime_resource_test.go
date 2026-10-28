package Lyger

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
	"testing"
	"testing/fstest"
)

// Resource-mutation tests author a complete selected distribution. Derive its
// support through the public frontend before testing the unchanged backend;
// production loading still rejects stale or incomplete provider resources.
func deriveTestRuntime(t *testing.T, resources fstest.MapFS) {
	t.Helper()
	c := NewLygerService(Options{Resources: resources}).NewCompilation([]byte("Use Runtime from `runtime`.\n"), "runtime-resource-test.l")
	c.Make()
	if c.Error != nil {
		t.Fatal(c.Error)
	}
	aml, err := c.ComposeLibrary("Runtime", "runtime")
	if err != nil {
		t.Fatal(err)
	}
	metadata := struct {
		Version int               `json:"version"`
		SHA256  string            `json:"sha256"`
		Sources map[string]string `json:"sources"`
	}{Version: 1, SHA256: fmt.Sprintf("%x", sha256.Sum256(aml)), Sources: make(map[string]string)}
	for _, origin := range c.SourceOrigins {
		if origin.Name == c.MainPath {
			continue
		}
		current, err := fs.ReadFile(resources, origin.Name)
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprintf("%x", sha256.Sum256(current)) != origin.SHA256 {
			t.Fatalf("selected source %s changed during generation", origin.Name)
		}
		metadata.Sources[origin.Name] = origin.SHA256
	}
	manifest, err := json.Marshal(metadata)
	if err != nil {
		t.Fatal(err)
	}
	resources["builtins/runtime/support.aml"] = &fstest.MapFile{Data: aml}
	resources["builtins/runtime/support.json"] = &fstest.MapFile{Data: manifest}
}
