package Lyger

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"testing"
	"testing/fstest"
)

// Resource-mutation tests author a complete selected distribution. Derive its
// support through the public frontend before testing the unchanged backend;
// production loading still rejects stale or incomplete provider resources.
func deriveTestRuntime(t *testing.T, resources fstest.MapFS) {
	t.Helper()
	aml, metadata, err := ComposeRuntimeResources(resources)
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := json.Marshal(metadata)
	if err != nil {
		t.Fatal(err)
	}
	resources["runtime/generated/support.aml"] = &fstest.MapFile{Data: aml}
	resources["runtime/generated/support.json"] = &fstest.MapFile{Data: manifest}
}

type changingCompositionRegistry struct {
	fs.FS
	reads int
}

func (r *changingCompositionRegistry) Open(name string) (fs.File, error) {
	if name != "dictionaries.json" {
		return r.FS.Open(name)
	}
	raw, err := fs.ReadFile(r.FS, name)
	if err != nil {
		return nil, err
	}
	raw = append(raw, bytes.Repeat([]byte("\n"), r.reads)...)
	r.reads++
	return fstest.MapFS{name: &fstest.MapFile{Data: raw}}.Open(name)
}

func TestRuntimeResourceCompositionUsesCoherentRegistry(t *testing.T) {
	t.Run("changed", func(t *testing.T) {
		if _, _, err := ComposeRuntimeResources(&changingCompositionRegistry{FS: dictionaryIndexFilesystem(t)}); err == nil {
			t.Fatal("composition accepted a registry changed during generation")
		}
	})
	t.Run("nil", func(t *testing.T) {
		defer func() {
			if value := recover(); value != nil {
				t.Fatalf("nil provider panicked: %v", value)
			}
		}()
		if _, _, err := ComposeRuntimeResources(nil); err == nil {
			t.Fatal("composition accepted an absent selected provider")
		}
	})
	t.Run("retargeted", func(t *testing.T) {
		files := dictionaryIndexFilesystem(t)
		const old = "runtime/eng/dictionary.l"
		const selected = "runtime/eng/selected.l"
		files[selected] = &fstest.MapFile{Data: bytes.Clone(files[old].Data)}
		files["dictionaries.json"].Data = bytes.ReplaceAll(files["dictionaries.json"].Data, []byte(old), []byte(selected))
		_, metadata, err := ComposeRuntimeResources(files)
		if err != nil || metadata.Sources[selected] == "" || metadata.Sources[old] != "" {
			t.Fatalf("composition ignored selected Runtime entry: %v, sources=%v", err, metadata.Sources)
		}
	})
}
