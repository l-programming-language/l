package cargocache

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// StageCurrent reconciles the current finite workspace at a fixed path inside
// a caller-locked CLI cache slot. Stable Cargo package paths permit ordinary
// dependency freshness checks across workspace changes. Source bytes remain
// authoritative; unchanged files retain their timestamps. Reintroduced files
// can retain the timestamp of their last retired bytes, never an older version.
//
// The complete source and prior owned tree are checked before mutation. Only
// recorded files and their recorded, emptied parent directories can be removed.
// The final inventory is committed last. A partial filesystem failure that
// leaves an incomplete inventory fails closed on a later attempt; it never
// authorizes removal of unknown files. Changed bytes within an intact inventory
// can be reconciled by the next attempt.
// Callers must retain the slot lock through Cargo and executable snapshots.
func StageCurrent(source, cache, excluded string) (string, error) {
	files, names, err := readInputs(source, excluded)
	if err != nil {
		return "", err
	}
	directory := filepath.Join(cache, "current")
	root := filepath.Join(directory, "source")
	if err := currentSeparateSource(source, cache); err != nil {
		return "", err
	}
	prior, err := currentInventory(cache, directory, root)
	if err != nil {
		return "", err
	}
	previous, err := currentOwnedFiles(root, prior)
	if err != nil {
		return "", err
	}
	history, err := currentReadHistory(directory, previous)
	if err != nil {
		return "", err
	}
	nextHistory, err := currentNextHistory(root, previous, files, history)
	if err != nil {
		return "", err
	}
	if err := privateDirectory(directory); err != nil {
		return "", err
	}
	// Publish consumed history before mutating sources. If staging fails,
	// obsolete hints cannot survive an intervening change at the same path.
	if err := currentWriteHistory(directory, nextHistory); err != nil {
		return "", err
	}
	if err := privateDirectory(root); err != nil {
		return "", err
	}
	for _, name := range prior {
		if _, keep := files[name]; !keep {
			if err := os.Remove(filepath.Join(root, name)); err != nil {
				return "", err
			}
		}
	}
	oldDirs, newDirs := currentParentDirectories(prior), currentParentDirectories(names)
	var obsoleteDirs []string
	for name := range oldDirs {
		if !newDirs[name] {
			obsoleteDirs = append(obsoleteDirs, name)
		}
	}
	// Children precede parents; os.Remove refuses any nonempty directory.
	sort.Slice(obsoleteDirs, func(i, j int) bool { return len(obsoleteDirs[i]) > len(obsoleteDirs[j]) })
	for _, name := range obsoleteDirs {
		if err := os.Remove(filepath.Join(root, name)); err != nil {
			return "", err
		}
	}
	for _, name := range names {
		if data, present := previous[name]; present && bytes.Equal(data, files[name]) {
			continue
		}
		path := filepath.Join(root, name)
		if _, present := previous[name]; !present {
			if _, err := os.Lstat(path); err == nil {
				return "", fmt.Errorf("current Cargo cache input aliases another path: %s", name)
			} else if !os.IsNotExist(err) {
				return "", err
			}
		}
		if err := privateDirectory(filepath.Dir(path)); err != nil {
			return "", err
		}
		if err := replaceFile(path, files[name]); err != nil {
			return "", err
		}
		if _, present := previous[name]; !present {
			if err := currentRestoreTime(path, files[name], history[name]); err != nil {
				return "", err
			}
		}
	}
	// Consuming the literal incoming names above is insufficient on filesystems
	// with case or Unicode aliases. Forget every retired name that now resolves
	// to supplied source before publishing the new inventory.
	if changed, err := currentPrunePresentHistory(root, nextHistory); err != nil {
		return "", err
	} else if changed {
		if err := currentWriteHistory(directory, nextHistory); err != nil {
			return "", err
		}
	}
	inventory, err := json.Marshal(names)
	if err != nil {
		return "", err
	}
	if err := replaceFile(filepath.Join(directory, "inputs.json"), inventory); err != nil {
		return "", err
	}
	return root, nil
}

func currentSeparateSource(source, cache string) error {
	source, err := currentPhysicalPath(source)
	if err != nil {
		return err
	}
	cache, err = currentPhysicalPath(cache)
	if err != nil {
		return err
	}
	for _, pair := range [][2]string{{source, cache}, {cache, source}} {
		relative, err := filepath.Rel(pair[0], pair[1])
		if err != nil {
			return err
		}
		if relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))) {
			return fmt.Errorf("Cargo source and current cache must not contain each other")
		}
	}
	return nil
}

// Resolve the existing prefix for containment checks even when a cache suffix
// has not been created yet. This does not change the supplied path spelling
// used by Cargo (and therefore native CARGO_MANIFEST_DIR expressions).
func currentPhysicalPath(path string) (string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	var suffix []string
	for {
		resolved, err := filepath.EvalSymlinks(path)
		if err == nil {
			for index := len(suffix) - 1; index >= 0; index-- {
				resolved = filepath.Join(resolved, suffix[index])
			}
			return resolved, nil
		}
		if !os.IsNotExist(err) || filepath.Dir(path) == path {
			return "", err
		}
		suffix = append(suffix, filepath.Base(path))
		path = filepath.Dir(path)
	}
}

func currentInventory(cache, directory, root string) ([]string, error) {
	for _, path := range []string{cache, directory, root} {
		if info, err := os.Lstat(path); err == nil {
			if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
				return nil, fmt.Errorf("Cargo cache directory must be a real directory: %s", path)
			}
		} else if !os.IsNotExist(err) {
			return nil, err
		}
	}
	entries, err := os.ReadDir(directory)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	for _, entry := range entries {
		if entry.Name() != "source" && entry.Name() != "inputs.json" && entry.Name() != currentHistoryFile {
			return nil, fmt.Errorf("unexpected current Cargo cache entry %s", entry.Name())
		}
	}
	path := filepath.Join(directory, "inputs.json")
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("Cargo cache inventory must be regular")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var names []string
	if err := json.Unmarshal(data, &names); err != nil {
		return nil, fmt.Errorf("invalid current Cargo cache inventory: %w", err)
	}
	for index, name := range names {
		if name == "." || !fs.ValidPath(filepath.ToSlash(name)) || filepath.Clean(name) != name || strings.Contains(name, "\\") || (index > 0 && name <= names[index-1]) {
			return nil, fmt.Errorf("invalid current Cargo cache inventory path %q", name)
		}
	}
	return names, nil
}

func currentParentDirectories(names []string) map[string]bool {
	directories := map[string]bool{}
	for _, name := range names {
		for parent := filepath.Dir(name); parent != "."; parent = filepath.Dir(parent) {
			directories[parent] = true
		}
	}
	return directories
}

func currentOwnedFiles(root string, names []string) (map[string][]byte, error) {
	owned, directories := map[string]bool{}, currentParentDirectories(names)
	for _, name := range names {
		owned[name] = true
	}
	files := map[string][]byte{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if os.IsNotExist(err) && path == root && len(names) == 0 {
			return nil
		}
		if err != nil {
			return err
		}
		name, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if name == "." {
			return nil
		}
		if entry.IsDir() {
			if !directories[name] {
				return fmt.Errorf("unexpected current Cargo cache directory %s", name)
			}
			return nil
		}
		if !entry.Type().IsRegular() || !owned[name] {
			return fmt.Errorf("unexpected or nonregular current Cargo cache input %s", name)
		}
		data, err := os.ReadFile(path)
		files[name] = data
		return err
	})
	if err != nil {
		return nil, err
	}
	if len(files) != len(names) {
		return nil, fmt.Errorf("current Cargo cache is missing recorded input files")
	}
	return files, nil
}
