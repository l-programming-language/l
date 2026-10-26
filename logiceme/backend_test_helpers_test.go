package Logiceme

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func runtimeGenerationResources(t *testing.T) fs.FS {
	t.Helper()
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	return os.DirFS(root)
}

func executionRegistryPanicAsFailure(t *testing.T) {
	t.Helper()
	if failure := recover(); failure != nil {
		t.Fatalf("execution registry panicked instead of returning a result: %v", failure)
	}
}

func marshalTestJSON(t testing.TB, value interface{}) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal test snapshot: %v", err)
	}
	return data
}
