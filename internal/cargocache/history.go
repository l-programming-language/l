package cargocache

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"
)

const (
	currentHistoryFile  = "history.json"
	currentHistoryLimit = 4096
)

// History describes only the last bytes retired from a currently absent path.
// It must never retain multiple versions of one path: Cargo may have replaced
// that path's native artifacts after different bytes were staged there.
// No source or build product is retained here, only a bounded freshness hint.
type currentHistoryRecord struct {
	Path       string `json:"path"`
	SHA256     string `json:"sha256"`
	ModifiedNS *int64 `json:"modified_ns"`
}

type currentHistory struct {
	Version int                    `json:"version"`
	Files   []currentHistoryRecord `json:"files"`
}

func currentReadHistory(directory string, previous map[string][]byte) (map[string]currentHistoryRecord, error) {
	path := filepath.Join(directory, currentHistoryFile)
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return map[string]currentHistoryRecord{}, nil
	}
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("current Cargo cache history must be regular")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var history currentHistory
	if err := decoder.Decode(&history); err != nil {
		return nil, fmt.Errorf("invalid current Cargo cache history: %w", err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return nil, fmt.Errorf("invalid trailing current Cargo cache history data")
	}
	if history.Version != 1 || history.Files == nil || len(history.Files) > currentHistoryLimit {
		return nil, fmt.Errorf("invalid current Cargo cache history version or record count")
	}
	records := make(map[string]currentHistoryRecord, len(history.Files))
	for index, record := range history.Files {
		name := record.Path
		if name == "." || !fs.ValidPath(filepath.ToSlash(name)) || filepath.Clean(name) != name || strings.Contains(name, "\\") || strings.ContainsRune(name, 0) || (index > 0 && name <= history.Files[index-1].Path) {
			return nil, fmt.Errorf("invalid current Cargo cache history path %q", name)
		}
		if _, present := previous[name]; present {
			return nil, fmt.Errorf("current Cargo cache history path is still present: %s", name)
		}
		if record.ModifiedNS == nil || len(record.SHA256) != sha256.Size*2 || strings.IndexFunc(record.SHA256, func(r rune) bool {
			return !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f'))
		}) != -1 {
			return nil, fmt.Errorf("invalid current Cargo cache history digest or timestamp for %s", name)
		}
		records[name] = record
	}
	if _, err := currentPrunePresentHistory(filepath.Join(directory, "source"), records); err != nil {
		return nil, err
	}
	return records, nil
}

// Filesystem aliases can resolve even when their spelling differs from the
// inventory (including case and Unicode normalization). Such paths cannot keep
// an independent retired version. A regular ancestor also consumes the hint.
func currentPrunePresentHistory(root string, records map[string]currentHistoryRecord) (bool, error) {
	changed := false
	for name := range records {
		_, err := os.Lstat(filepath.Join(root, name))
		if err == nil || errors.Is(err, syscall.ENOTDIR) {
			delete(records, name)
			changed = true
		} else if !os.IsNotExist(err) {
			return false, err
		}
	}
	return changed, nil
}

func currentNextHistory(root string, previous, files map[string][]byte, history map[string]currentHistoryRecord) (map[string]currentHistoryRecord, error) {
	now := time.Now()
	next := make(map[string]currentHistoryRecord, len(history))
	for name, record := range history {
		// Consume every incoming path before any source write, including paths
		// whose bytes differ. A failed stage may lose reuse, never retain an
		// older version that a later build could incorrectly consider fresh.
		if _, incoming := files[name]; !incoming && !time.Unix(0, *record.ModifiedNS).After(now) {
			next[name] = record
		}
	}
	for name, data := range previous {
		if _, keep := files[name]; keep {
			continue
		}
		info, err := os.Lstat(filepath.Join(root, name))
		if err != nil {
			return nil, err
		}
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("retired current Cargo cache input must be regular: %s", name)
		}
		// Caches can contain future-dated inputs. They must undergo
		// normal Cargo freshness checks rather than preserving that timestamp.
		if info.ModTime().After(now) {
			continue
		}
		stamp := info.ModTime().UnixNano()
		next[name] = currentHistoryRecord{Path: name, SHA256: fmt.Sprintf("%x", sha256.Sum256(data)), ModifiedNS: &stamp}
	}
	return next, nil
}

func currentWriteHistory(directory string, records map[string]currentHistoryRecord) error {
	retained := make([]currentHistoryRecord, 0, len(records))
	for _, record := range records {
		retained = append(retained, record)
	}
	if len(retained) > currentHistoryLimit {
		// Forgetting old hints only costs a rebuild. Prefer the most recently
		// modified files, then use path order to make ties deterministic.
		sort.Slice(retained, func(i, j int) bool {
			if *retained[i].ModifiedNS != *retained[j].ModifiedNS {
				return *retained[i].ModifiedNS > *retained[j].ModifiedNS
			}
			return retained[i].Path < retained[j].Path
		})
		retained = retained[:currentHistoryLimit]
	}
	sort.Slice(retained, func(i, j int) bool { return retained[i].Path < retained[j].Path })
	data, err := json.Marshal(currentHistory{Version: 1, Files: retained})
	if err != nil {
		return err
	}
	return replaceFile(filepath.Join(directory, currentHistoryFile), data)
}

func currentRestoreTime(path string, data []byte, record currentHistoryRecord) error {
	if record.ModifiedNS == nil || record.SHA256 != fmt.Sprintf("%x", sha256.Sum256(data)) {
		return nil
	}
	stamp := time.Unix(0, *record.ModifiedNS)
	if stamp.After(time.Now()) {
		return nil
	}
	return os.Chtimes(path, stamp, stamp)
}
