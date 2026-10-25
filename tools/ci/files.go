package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

func repositoryFiles(root string) ([]string, error) {
	cmd := exec.Command("git", "ls-files", "--cached", "--others", "--exclude-standard", "-z")
	cmd.Dir = root
	data, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var files []string
	seen := map[string]bool{}
	for _, path := range strings.Split(string(data), "\x00") {
		if path != "" && !seen[path] {
			if !filepath.IsLocal(path) {
				return nil, fmt.Errorf("unsafe repository path %q", path)
			}
			if _, err := os.Lstat(filepath.Join(root, path)); os.IsNotExist(err) {
				continue
			} else if err != nil {
				return nil, err
			}
			files = append(files, path)
			seen[path] = true
		}
	}
	sort.Strings(files)
	return files, nil
}

func cacheDigests(root string) (string, string, error) {
	files, err := repositoryFiles(root)
	if err != nil {
		return "", "", err
	}
	dependencies, source := sha256.New(), sha256.New()
	for _, name := range files {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			return "", "", err
		}
		if _, err := source.Write(append(append([]byte(name+"\x00"), data...), 0)); err != nil {
			return "", "", err
		}
		if name == "go.mod" || name == "go.sum" || name == ".ci-tools.json" || strings.HasSuffix(name, ".lock") || strings.Contains(filepath.Base(name), "Cargo") {
			if _, err := dependencies.Write(append(append([]byte(name+"\x00"), data...), 0)); err != nil {
				return "", "", err
			}
		}
	}
	return fmt.Sprintf("%x", dependencies.Sum(nil)), fmt.Sprintf("%x", source.Sum(nil)), nil
}

func copyFile(source, destination string) error {
	info, err := os.Lstat(source)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("seed input is not a regular file: %s", source)
	}
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
		return err
	}
	return os.WriteFile(destination, data, 0644)
}

func readJSON(path string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}
