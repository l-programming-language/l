// Package cargocache keeps Cargo input paths stable without changing source
// bytes. Callers must hold their native build-cache lock through compilation and
// artifact retention; this package neither invokes Cargo nor caches test results.
package cargocache

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"
)

func readInputs(source, excluded string) (map[string][]byte, []string, error) {
	files := map[string][]byte{}
	var names []string
	err := filepath.WalkDir(source, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == excluded {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("Cargo input must be regular: %s", path)
		}
		name, err := filepath.Rel(source, path)
		if err != nil || !fs.ValidPath(filepath.ToSlash(name)) {
			return fmt.Errorf("invalid Cargo input path %q: %v", path, err)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		names = append(names, name)
		files[name] = data
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	sort.Strings(names)
	return files, names, nil
}

// Shards bounds independent compilation caches and their dependency copies.
const Shards = 4

// Shard bounds concurrent native compilation by actual supplied input identity.
// Every shard still executes Cargo; it never substitutes cached build success.
func Shard(source, excluded string) (int, error) {
	files, names, err := readInputs(source, excluded)
	if err != nil {
		return 0, err
	}
	digest := sha256.New()
	for _, name := range names {
		for _, data := range [][]byte{[]byte(filepath.ToSlash(name)), files[name]} {
			_ = binary.Write(digest, binary.LittleEndian, uint64(len(data))) //nolint:errcheck // A uint64 is encodable and hash.Hash.Write cannot fail.
			_, _ = digest.Write(data)
		}
	}
	return int(digest.Sum(nil)[0]) % Shards, nil
}

// Stage copies a finite set of regular input files. Manifests, locks and the
// complete path inventory identify the graph; native source contents can change
// within that graph so Cargo can retain its ordinary incremental products.
func Stage(source, cache, excluded string) (string, error) {
	files, names, err := readInputs(source, excluded)
	if err != nil {
		return "", err
	}
	digest := sha256.New()
	write := func(data []byte) {
		_ = binary.Write(digest, binary.LittleEndian, uint64(len(data))) //nolint:errcheck // A uint64 is encodable and hash.Hash.Write cannot fail.
		_, _ = digest.Write(data)
	}
	for _, name := range names {
		write([]byte(filepath.ToSlash(name)))
		if base := filepath.Base(name); base == "Cargo.toml" || base == "Cargo.lock" {
			write(files[name])
		}
	}
	directory := filepath.Join(cache, fmt.Sprintf("%x", digest.Sum(nil)))
	root := filepath.Join(directory, "source")
	if err := privateDirectory(root); err != nil {
		return "", err
	}
	inventory, err := json.Marshal(names)
	if err != nil {
		return "", err
	}
	if prior, err := os.ReadFile(filepath.Join(directory, "inputs.json")); err == nil && !bytes.Equal(prior, inventory) {
		return "", fmt.Errorf("Cargo cache inventory disagrees with its identity")
	} else if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	for _, name := range names {
		path := filepath.Join(root, name)
		if err := privateDirectory(filepath.Dir(path)); err != nil {
			return "", err
		}
		if info, err := os.Lstat(path); err == nil && !info.Mode().IsRegular() {
			return "", fmt.Errorf("Cargo cache input must be regular: %s", path)
		} else if err != nil && !os.IsNotExist(err) {
			return "", err
		}
		if prior, err := os.ReadFile(path); err == nil && bytes.Equal(prior, files[name]) {
			continue
		} else if err != nil && !os.IsNotExist(err) {
			return "", err
		}
		if err := replaceFile(path, files[name]); err != nil {
			return "", err
		}
	}
	if err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		name, err := filepath.Rel(root, path)
		if err != nil || !entry.Type().IsRegular() {
			return fmt.Errorf("invalid Cargo cache input %s", path)
		}
		if _, supplied := files[name]; !supplied {
			return fmt.Errorf("unexpected Cargo cache input %s", name)
		}
		return nil
	}); err != nil {
		return "", err
	}
	if err := replaceFile(filepath.Join(directory, "inputs.json"), inventory); err != nil {
		return "", err
	}
	return root, nil
}

// Activate invalidates Cargo's local freshness state when a different input
// graph last used this target directory. Cargo may otherwise treat an older,
// same-named package in another directory as fresh. Record before compilation:
// even a failed build may replace some local artifacts.
func Activate(root, target string) error {
	if err := privateDirectory(target); err != nil {
		return err
	}
	marker := filepath.Join(target, ".l-active-inputs")
	if data, err := os.ReadFile(marker); err == nil && string(data) == root {
		return nil
	} else if err != nil && !os.IsNotExist(err) {
		return err
	}
	// Cargo compares source mtimes with the start of the last compilation.
	// Future-dating these inputs would make the upcoming build immediately
	// stale and force another compilation of unchanged local libraries.
	stamp := time.Now()
	if err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("Cargo input must be regular: %s", path)
		}
		return os.Chtimes(path, stamp, stamp)
	}); err != nil {
		return err
	}
	return replaceFile(marker, []byte(root))
}

func privateDirectory(path string) error {
	if info, err := os.Lstat(path); err == nil {
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("Cargo cache directory is not a directory: %s", path)
		}
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := privateDirectory(filepath.Dir(path)); err != nil {
		return err
	}
	return os.Mkdir(path, 0700)
}

func replaceFile(path string, data []byte) (resultErr error) {
	file, err := os.CreateTemp(filepath.Dir(path), ".input-")
	if err != nil {
		return err
	}
	temporaryPath := file.Name()
	defer func() { resultErr = cleanupInputTemporary(temporaryPath, resultErr, os.Remove) }()
	if _, err := file.Write(data); err != nil {
		_ = file.Close() //nolint:errcheck // The write already failed; retain that cause while closing the temporary file.
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryPath, path)
}

func cleanupInputTemporary(path string, primary error, remove func(string) error) error {
	cleanupErr := remove(path)
	// Rename consumes the temporary path. Cleanup cannot undo a committed replacement.
	if primary == nil || cleanupErr == nil || errors.Is(cleanupErr, fs.ErrNotExist) {
		return primary
	}
	return errors.Join(primary, fmt.Errorf("remove Cargo cache temporary input %q: %w", path, cleanupErr))
}
