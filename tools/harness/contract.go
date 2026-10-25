package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/l-programming-language/l/tools/internal/testgate"
)

type fileContract struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Role   string `json:"role,omitempty"`
}

type testContract struct {
	Package string `json:"package"`
	Name    string `json:"name"`
}

type contract struct {
	SchemaVersion  int            `json:"schema_version"`
	Module         string         `json:"module"`
	ProtectedFiles []fileContract `json:"protected_files"`
	AbsentFiles    []string       `json:"absent_files"`
	Tests          []testContract `json:"tests"`
}

type integrityCheck struct {
	Kind    string `json:"kind"`
	Path    string `json:"path,omitempty"`
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
}

type integrityReport struct {
	OK     bool             `json:"ok"`
	Checks []integrityCheck `json:"checks"`
}

func readContract(root string) (contract, string, error) {
	var c contract
	path := filepath.Join(root, "harness", "contract.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return c, "", err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&c); err != nil {
		return c, "", fmt.Errorf("contract: %w", err)
	}
	if err = decoder.Decode(new(interface{})); err != io.EOF {
		return c, "", fmt.Errorf("contract must contain one JSON object")
	}
	if c.SchemaVersion != 2 || strings.TrimSpace(c.Module) == "" || len(c.Tests) == 0 || len(c.ProtectedFiles) == 0 {
		return c, "", fmt.Errorf("contract requires schema_version 2, module, protected_files, and original tests")
	}
	paths := make(map[string]bool)
	for _, file := range c.ProtectedFiles {
		if _, err := contractPath(root, file.Path); err != nil {
			return c, "", err
		}
		if paths[file.Path] {
			return c, "", fmt.Errorf("duplicate protected path: %s", file.Path)
		}
		paths[file.Path] = true
		hash, err := hex.DecodeString(file.SHA256)
		if err != nil || len(hash) != sha256.Size {
			return c, "", fmt.Errorf("invalid SHA-256 for %s", file.Path)
		}
	}
	for _, relative := range c.AbsentFiles {
		if _, err := contractPath(root, relative); err != nil {
			return c, "", err
		}
		if paths[relative] {
			return c, "", fmt.Errorf("duplicate or overlapping absent path: %s", relative)
		}
		paths[relative] = true
	}
	seen := make(map[string]bool)
	for _, test := range c.Tests {
		validPackage := test.Package == "."
		if strings.HasPrefix(test.Package, "./") {
			_, err := contractPath(root, strings.TrimPrefix(test.Package, "./"))
			validPackage = err == nil
		}
		if strings.TrimSpace(test.Name) == "" || !validPackage {
			return c, "", fmt.Errorf("invalid test contract: %+v", test)
		}
		key := test.Package + "\x00" + test.Name
		if seen[key] {
			return c, "", fmt.Errorf("duplicate original test: %s/%s", test.Package, test.Name)
		}
		seen[key] = true
	}
	sum := sha256.Sum256(data)
	return c, hex.EncodeToString(sum[:]), nil
}

// Contract paths use normalized repository-relative names; document links may use ../.
func contractPath(root, relative string) (string, error) {
	if relative == "." || !fs.ValidPath(relative) || strings.ContainsAny(relative, "\\\x00") {
		return "", fmt.Errorf("expected a normalized repository-relative path: %q", relative)
	}
	return filepath.Join(root, filepath.FromSlash(relative)), nil
}

func inspectIntegrity(root string, c contract, originalContractHash string) integrityReport {
	r := integrityReport{OK: true, Checks: []integrityCheck{}}
	add := func(kind, path string, err error) {
		check := integrityCheck{Kind: kind, Path: path, OK: err == nil}
		if err != nil {
			r.OK = false
			check.Message = err.Error()
		}
		r.Checks = append(r.Checks, check)
	}
	actualContractHash, err := testgate.FileHash(filepath.Join(root, "harness", "contract.json"))
	if err == nil && actualContractHash != originalContractHash {
		err = fmt.Errorf("contract changed during this invocation")
	}
	add("contract", "harness/contract.json", err)
	add("environment", "GOFLAGS", validateGoFlags(originalGoFlags()))
	moduleBytes, moduleErr := os.ReadFile(filepath.Join(root, "go.mod"))
	if moduleErr == nil {
		actualModule := ""
		for _, line := range strings.Split(string(moduleBytes), "\n") {
			fields := strings.Fields(line)
			if len(fields) >= 2 && fields[0] == "module" {
				actualModule = strings.Trim(fields[1], `"`)
				break
			}
		}
		if actualModule != c.Module {
			moduleErr = fmt.Errorf("module %q differs from contract %q", actualModule, c.Module)
		}
	}
	add("module", "go.mod", moduleErr)
	for _, file := range c.ProtectedFiles {
		path, err := contractPath(root, file.Path)
		var actual string
		if err == nil {
			actual, err = testgate.FileHash(path)
		}
		if err == nil && actual != strings.ToLower(file.SHA256) {
			err = fmt.Errorf("SHA-256 mismatch: expected %s, actual %s", file.SHA256, actual)
		}
		add("protected_file", file.Path, err)
	}
	for _, relative := range c.AbsentFiles {
		path, err := contractPath(root, relative)
		if err == nil {
			_, err = os.Lstat(path)
			if os.IsNotExist(err) {
				err = nil
			} else if err == nil {
				err = fmt.Errorf("retired input must remain absent")
			}
		}
		add("absent_file", relative, err)
	}
	add("source_inventory", "demo, cmd/l/test, lyger/test, testdata", inspectSourceMembership(root, c.ProtectedFiles))
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if entry.IsDir() {
			switch relative {
			case ".git", ".harness", ".idea", "bin", "app", "pkg", "logs",
				"website/.tools", "website/node_modules",
				"website/src/content/docs/docs", "website/src/content/docs/releases":
				return filepath.SkipDir
			}
		}
		if !entry.Type().IsRegular() || !strings.HasSuffix(entry.Name(), ".md") {
			return nil
		}
		data, err := os.ReadFile(path)
		add("document", relative, err)
		if err == nil {
			for _, issue := range checkMarkdownLinks(path, string(data)) {
				add("document_link", relative, issue)
			}
		}
		return nil
	})
	if err != nil {
		add("document_discovery", ".", err)
	}
	return r
}

// Derive the current source inventory from the same locks used for byte integrity.
func inspectSourceMembership(root string, files []fileContract) error {
	roots := []string{"demo", "cmd/l/test", "lyger/test", "testdata"}
	expected := make(map[string]bool)
	for _, file := range files {
		if strings.HasSuffix(file.Path, ".l") {
			for _, directory := range roots {
				if strings.HasPrefix(file.Path, directory+"/") {
					expected[file.Path] = true
					break
				}
			}
		}
	}
	for _, directory := range roots {
		path := filepath.Join(root, filepath.FromSlash(directory))
		err := filepath.WalkDir(path, func(path string, entry fs.DirEntry, err error) error {
			if os.IsNotExist(err) {
				return nil // Missing locked sources still fail their byte checks.
			}
			if err != nil {
				return err
			}
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".l") {
				return nil
			}
			relative, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			relative = filepath.ToSlash(relative)
			if !expected[relative] {
				return fmt.Errorf("unprotected L source: %s", relative)
			}
			delete(expected, relative)
			return nil
		})
		if err != nil {
			return err
		}
	}
	if len(expected) != 0 {
		return fmt.Errorf("protected L source missing from source inventory")
	}
	return nil
}

var markdownLinkPattern = regexp.MustCompile(`!?\[[^\]\n]*\]\(([^)\n]+)\)`)
var inlineCodePattern = regexp.MustCompile("`+[^`]*`+")
var sourceLinePattern = regexp.MustCompile(`:[0-9]+$`)

func checkMarkdownLinks(document, contents string) []error {
	var issues []error
	var fence string
	for lineIndex, line := range strings.Split(contents, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "```") || strings.HasPrefix(trim, "~~~") {
			marker := trim[:3]
			if fence == "" {
				fence = marker
			} else if marker == fence {
				fence = ""
			}
			continue
		}
		if fence != "" {
			continue
		}
		line = inlineCodePattern.ReplaceAllString(line, "")
		for _, match := range markdownLinkPattern.FindAllStringSubmatch(line, -1) {
			target := strings.TrimSpace(match[1])
			if strings.HasPrefix(target, "<") {
				if end := strings.Index(target, ">"); end >= 0 {
					target = target[1:end]
				}
			} else if i := strings.Index(target, ` "`); i >= 0 {
				target = target[:i]
			}
			if target == "" || strings.HasPrefix(target, "#") {
				continue
			}
			target = strings.SplitN(target, "#", 2)[0]
			target = sourceLinePattern.ReplaceAllString(target, "")
			if parsed, err := url.Parse(target); err == nil && parsed.Scheme != "" {
				continue // Remote/app links are intentionally outside doctor.
			}
			if decoded, err := url.PathUnescape(target); err == nil {
				target = decoded
			}
			path := filepath.FromSlash(target)
			if !filepath.IsAbs(path) {
				path = filepath.Join(filepath.Dir(document), path)
			}
			if _, err := os.Stat(path); err != nil {
				issues = append(issues, fmt.Errorf("line %d: local link %q: %w", lineIndex+1, match[1], err))
			}
		}
	}
	return issues
}
