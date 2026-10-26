package cargocache

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func historyStage(t *testing.T, cache string, files map[string]string) string {
	t.Helper()
	source := t.TempDir()
	currentWrite(t, source, files)
	before := currentSnapshot(t, source)
	root, err := StageCurrent(source, cache, "")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, currentSnapshot(t, source)) || !reflect.DeepEqual(before, currentSnapshot(t, root)) {
		t.Fatal("staging changed supplied bytes or left an inexact source tree")
	}
	return root
}

func historyStamp(t *testing.T, root, name string, stamp time.Time) {
	t.Helper()
	if err := os.Chtimes(filepath.Join(root, name), stamp, stamp); err != nil {
		t.Fatal(err)
	}
}

func historyModTime(t *testing.T, root, name string) time.Time {
	t.Helper()
	info, err := os.Stat(filepath.Join(root, name))
	if err != nil {
		t.Fatal(err)
	}
	return info.ModTime()
}

func TestStageCurrentRetainsOnlyLastRemovedBytes(t *testing.T) {
	cache := t.TempDir()
	const name = "support/runtime/src/lib.rs"
	files := func(body string) map[string]string {
		return map[string]string{"Cargo.toml": "workspace", name: body}
	}
	absent := map[string]string{"Cargo.toml": "workspace without runtime"}
	stampA, stampB := time.Unix(123456789, 0), time.Unix(234567890, 0)
	root := historyStage(t, cache, files("A"))
	historyStamp(t, root, name, stampA)
	historyStage(t, cache, absent)
	root = historyStage(t, cache, files("A"))
	if got := historyModTime(t, root, name); !got.Equal(stampA) {
		t.Fatalf("identical removed file lost its last modification time: got %s want %s", got, stampA)
	}
	historyStage(t, cache, absent)
	root = historyStage(t, cache, files("B"))
	if got := historyModTime(t, root, name); got.Equal(stampA) {
		t.Fatal("different bytes reused a removed version's timestamp")
	}
	historyStamp(t, root, name, stampB)
	historyStage(t, cache, absent)
	root = historyStage(t, cache, files("B"))
	if got := historyModTime(t, root, name); !got.Equal(stampB) {
		t.Fatalf("latest removed version lost its timestamp: got %s want %s", got, stampB)
	}
	historyStage(t, cache, absent)
	root = historyStage(t, cache, files("A"))
	if got := historyModTime(t, root, name); got.Equal(stampA) || got.Equal(stampB) {
		t.Fatal("superseded A reused a historical timestamp after B occupied the same path")
	}
}

func TestStageCurrentConsumesHistoryBeforePresentChanges(t *testing.T) {
	cache := t.TempDir()
	const name = "src/lib.rs"
	stampA, stampB := time.Unix(123456789, 0), time.Unix(234567890, 0)
	root := historyStage(t, cache, map[string]string{name: "A"})
	historyStamp(t, root, name, stampA)
	historyStage(t, cache, map[string]string{})
	root = historyStage(t, cache, map[string]string{name: "A"})
	if got := historyModTime(t, root, name); !got.Equal(stampA) {
		t.Fatal("unchanged reintroduction did not consume the retired timestamp")
	}
	data, err := os.ReadFile(filepath.Join(cache, "current", "history.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), name) {
		t.Fatal("present source retained a retired timestamp record")
	}
	root = historyStage(t, cache, map[string]string{name: "B"})
	historyStamp(t, root, name, stampB)
	historyStage(t, cache, map[string]string{})
	root = historyStage(t, cache, map[string]string{name: "A"})
	if got := historyModTime(t, root, name); got.Equal(stampA) || got.Equal(stampB) {
		t.Fatal("present source update left a reusable older timestamp")
	}
}

func historyRecord(name, body string, stamp int64) string {
	return fmt.Sprintf(`{"path":%q,"sha256":"%x","modified_ns":%d}`, name, sha256.Sum256([]byte(body)), stamp)
}

func TestStageCurrentRejectsInvalidHistoryBeforeMutation(t *testing.T) {
	valid := historyRecord("retired.rs", "retired", 123456789)
	cases := map[string]string{
		"unknown-field":        `{"version":1,"files":[` + valid + `],"unexpected":true}`,
		"unknown-record-field": `{"version":1,"files":[{"path":"retired.rs","sha256":"` + strings.Repeat("0", 64) + `","modified_ns":1,"unexpected":true}]}`,
		"unknown-version":      `{"version":2,"files":[]}`,
		"missing-version":      `{"files":[]}`,
		"missing-files":        `{"version":1}`,
		"null-files":           `{"version":1,"files":null}`,
		"duplicate":            `{"version":1,"files":[` + valid + `,` + valid + `]}`,
		"unsorted":             `{"version":1,"files":[` + valid + `,` + historyRecord("earlier.rs", "earlier", 1) + `]}`,
		"escape":               `{"version":1,"files":[` + historyRecord("../escape", "bad", 1) + `]}`,
		"nul":                  `{"version":1,"files":[{"path":"a\u0000b","sha256":"` + strings.Repeat("0", 64) + `","modified_ns":1}]}`,
		"dot":                  `{"version":1,"files":[` + historyRecord(".", "bad", 1) + `]}`,
		"backslash":            `{"version":1,"files":[` + historyRecord(`a\b`, "bad", 1) + `]}`,
		"bad-digest":           `{"version":1,"files":[{"path":"retired.rs","sha256":"no","modified_ns":1}]}`,
		"missing-time":         `{"version":1,"files":[{"path":"retired.rs","sha256":"` + strings.Repeat("0", 64) + `"}]}`,
		"present-owned-path":   `{"version":1,"files":[` + historyRecord("src/main.rs", "before", 1) + `]}`,
		"trailing-data":        `{"version":1,"files":[]} {}`,
		"symlink":              "",
		"unknown-cache-entry":  "",
	}
	var tooMany []string
	for index := 0; index < 4097; index++ {
		tooMany = append(tooMany, historyRecord(fmt.Sprintf("retired/%04d.rs", index), "retired", 1))
	}
	cases["too-many-records"] = `{"version":1,"files":[` + strings.Join(tooMany, ",") + `]}`
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			cache := t.TempDir()
			root := historyStage(t, cache, map[string]string{"Cargo.toml": "manifest", "src/main.rs": "before"})
			path := filepath.Join(cache, "current", "history.json")
			switch name {
			case "symlink":
				if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
					t.Fatal(err)
				}
				if err := os.Symlink(filepath.Join(root, "Cargo.toml"), path); err != nil {
					t.Fatal(err)
				}
			case "unknown-cache-entry":
				currentWrite(t, filepath.Dir(root), map[string]string{"unknown.json": "unexpected"})
			default:
				if err := os.WriteFile(path, []byte(data), 0600); err != nil {
					t.Fatal(err)
				}
			}
			source := t.TempDir()
			currentWrite(t, source, map[string]string{"Cargo.toml": "changed", "src/main.rs": "after"})
			beforeSource, beforeCache := currentSnapshot(t, source), currentSnapshot(t, cache)
			if _, err := StageCurrent(source, cache, ""); err == nil {
				t.Fatal("malformed history was accepted")
			}
			if !reflect.DeepEqual(beforeSource, currentSnapshot(t, source)) || !reflect.DeepEqual(beforeCache, currentSnapshot(t, cache)) {
				t.Fatal("malformed history caused source or cache mutation")
			}
		})
	}
}

func TestStageCurrentBoundsRetiredHistory(t *testing.T) {
	cache := t.TempDir()
	root := historyStage(t, cache, map[string]string{"current.rs": "current"})
	var records []string
	for index := 0; index < 4096; index++ {
		records = append(records, historyRecord(fmt.Sprintf("retired/%04d.rs", index), "retired", int64(index+1)))
	}
	path := filepath.Join(cache, "current", "history.json")
	if err := os.WriteFile(path, []byte(`{"version":1,"files":[`+strings.Join(records, ",")+`]}`), 0600); err != nil {
		t.Fatal(err)
	}
	historyStage(t, cache, map[string]string{})
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var retained struct {
		Version int `json:"version"`
		Files   []struct {
			Path string `json:"path"`
		} `json:"files"`
	}
	if err := json.Unmarshal(data, &retained); err != nil {
		t.Fatal(err)
	}
	if retained.Version != 1 || len(retained.Files) != 4096 {
		t.Fatalf("unbounded or discarded history: version=%d records=%d", retained.Version, len(retained.Files))
	}
	for index, record := range retained.Files {
		if index > 0 && retained.Files[index-1].Path >= record.Path {
			t.Fatal("retained history is not sorted and unique")
		}
	}
	if _, err := os.Stat(filepath.Join(root, "current.rs")); !os.IsNotExist(err) {
		t.Fatal("retired source remained in the current input tree")
	}
}

func TestStageCurrentMissingHistoryRebuildsNormally(t *testing.T) {
	cache := t.TempDir()
	stamp := time.Unix(123456789, 0)
	root := historyStage(t, cache, map[string]string{"src/lib.rs": "A"})
	historyStamp(t, root, "src/lib.rs", stamp)
	historyStage(t, cache, map[string]string{})
	if err := os.Remove(filepath.Join(cache, "current", "history.json")); err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	root = historyStage(t, cache, map[string]string{"src/lib.rs": "A"})
	if got := historyModTime(t, root, "src/lib.rs"); got.Equal(stamp) {
		t.Fatal("missing metadata restored an unavailable timestamp")
	}
}

func TestStageCurrentDoesNotRestoreFilesystemAliases(t *testing.T) {
	for _, names := range [][2]string{{"A.rs", "a.rs"}, {"caf\u00e9.rs", "cafe\u0301.rs"}} {
		t.Run(names[0], func(t *testing.T) {
			cache := t.TempDir()
			stampA, stampB := time.Unix(123456789, 0), time.Unix(234567890, 0)
			root := historyStage(t, cache, map[string]string{names[0]: "A"})
			historyStamp(t, root, names[0], stampA)
			historyStage(t, cache, map[string]string{})
			root = historyStage(t, cache, map[string]string{names[1]: "B"})
			_, aliasErr := os.Lstat(filepath.Join(root, names[0]))
			if aliasErr != nil && !os.IsNotExist(aliasErr) {
				t.Fatal(aliasErr)
			}
			aliases := aliasErr == nil
			historyStamp(t, root, names[1], stampB)
			historyStage(t, cache, map[string]string{})
			root = historyStage(t, cache, map[string]string{names[0]: "A"})
			got := historyModTime(t, root, names[0])
			if aliases && (got.Equal(stampA) || got.Equal(stampB)) {
				t.Fatal("filesystem-equivalent path reused an older version's timestamp")
			}
			if !aliases && !got.Equal(stampA) {
				t.Fatal("distinct filesystem paths lost their independent retired timestamps")
			}
		})
	}
}

func TestStageCurrentDoesNotRetainFutureTimestamps(t *testing.T) {
	cache := t.TempDir()
	future := time.Now().Add(time.Hour)
	root := historyStage(t, cache, map[string]string{"src/lib.rs": "A"})
	historyStamp(t, root, "src/lib.rs", future)
	historyStage(t, cache, map[string]string{})
	root = historyStage(t, cache, map[string]string{"src/lib.rs": "A"})
	if got := historyModTime(t, root, "src/lib.rs"); got.After(time.Now()) {
		t.Fatal("future timestamp was retained for reintroduced source")
	}
	historyStage(t, cache, map[string]string{})
	data := `{"version":1,"files":[` + historyRecord("src/lib.rs", "A", future.UnixNano()) + `]}`
	if err := os.WriteFile(filepath.Join(cache, "current", "history.json"), []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	root = historyStage(t, cache, map[string]string{"src/lib.rs": "A"})
	if got := historyModTime(t, root, "src/lib.rs"); got.After(time.Now()) {
		t.Fatal("future metadata timestamp was restored")
	}
}
