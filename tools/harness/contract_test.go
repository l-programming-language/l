package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/l-programming-language/l/tools/internal/testgate"
)

func writeIntegrityFile(t *testing.T, root, name, data string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}
}

func integrityFixture(t *testing.T) (string, contract, string) {
	t.Helper()
	root := t.TempDir()
	writeIntegrityFile(t, root, "go.mod", "module example.test/l\n\ngo 1.20\n")
	writeIntegrityFile(t, root, "testdata/main.l", "fixed\n")
	hash, err := testgate.FileHash(filepath.Join(root, "testdata/main.l"))
	if err != nil {
		t.Fatal(err)
	}
	c := contract{SchemaVersion: 2, Module: "example.test/l",
		ProtectedFiles: []fileContract{{Path: "testdata/main.l", SHA256: hash}},
		AbsentFiles:    []string{"retired.json"}, Tests: []testContract{{Package: ".", Name: "TestOriginal"}}}
	if err := os.MkdirAll(filepath.Join(root, "harness"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := testgate.WriteJSON(filepath.Join(root, "harness/contract.json"), c); err != nil {
		t.Fatal(err)
	}
	_, hash, err = readContract(root)
	if err != nil {
		t.Fatal(err)
	}
	return root, c, hash
}

func TestCurrentIntegrity(t *testing.T) {
	t.Setenv("L_HARNESS_ORIGINAL_GOFLAGS", "")
	for _, name := range []string{"unchanged", "changed bytes", "missing source", "absent file", "absent symlink", "contract changed"} {
		t.Run(name, func(t *testing.T) {
			root, c, hash := integrityFixture(t)
			switch name {
			case "changed bytes":
				writeIntegrityFile(t, root, "testdata/main.l", "fixed\r\n")
			case "missing source":
				if err := os.Remove(filepath.Join(root, "testdata/main.l")); err != nil {
					t.Fatal(err)
				}
			case "absent file":
				writeIntegrityFile(t, root, "retired.json", "{}")
			case "absent symlink":
				if err := os.Symlink("missing", filepath.Join(root, "retired.json")); err != nil {
					t.Fatal(err)
				}
			case "contract changed":
				writeIntegrityFile(t, root, "harness/contract.json", "{}")
			}
			if report := inspectIntegrity(root, c, hash); report.OK != (name == "unchanged") {
				t.Fatalf("unexpected integrity result: %+v", report)
			}
		})
	}
	for _, directory := range []string{"demo", "cmd/l/test", "lyger/test", "testdata"} {
		t.Run("extra source/"+directory, func(t *testing.T) {
			root, c, hash := integrityFixture(t)
			writeIntegrityFile(t, root, directory+"/extra.l", "unreviewed source")
			if report := inspectIntegrity(root, c, hash); report.OK {
				t.Fatal("unprotected source passed integrity")
			}
		})
	}
}

func TestReadContractRejectsMalformedCurrentData(t *testing.T) {
	for _, name := range []string{"old schema", "missing module", "missing files", "missing tests", "invalid hash", "short hash", "duplicate file", "duplicate absent", "overlap", "file path", "absent path", "test package", "empty test", "duplicate test", "unknown field", "trailing JSON"} {
		t.Run(name, func(t *testing.T) {
			root, c, _ := integrityFixture(t)
			switch name {
			case "old schema":
				c.SchemaVersion = 1
			case "missing module":
				c.Module = ""
			case "missing files":
				c.ProtectedFiles = nil
			case "missing tests":
				c.Tests = nil
			case "invalid hash":
				c.ProtectedFiles[0].SHA256 = strings.Repeat("z", 64)
			case "short hash":
				c.ProtectedFiles[0].SHA256 = "00"
			case "duplicate file":
				c.ProtectedFiles = append(c.ProtectedFiles, c.ProtectedFiles[0])
			case "duplicate absent":
				c.AbsentFiles = append(c.AbsentFiles, c.AbsentFiles[0])
			case "overlap":
				c.AbsentFiles = append(c.AbsentFiles, c.ProtectedFiles[0].Path)
			case "file path":
				c.ProtectedFiles[0].Path = "../outside"
			case "absent path":
				c.AbsentFiles[0] = "a/../outside"
			case "test package":
				c.Tests[0].Package = "./../outside"
			case "empty test":
				c.Tests[0].Name = ""
			case "duplicate test":
				c.Tests = append(c.Tests, c.Tests[0])
			}
			data, err := json.Marshal(c)
			if err != nil {
				t.Fatal(err)
			}
			if name == "unknown field" {
				data = append([]byte(`{"baseline_file":"old.json",`), data[1:]...)
			}
			if name == "trailing JSON" {
				data = append(data, []byte(" {}")...)
			}
			writeIntegrityFile(t, root, "harness/contract.json", string(data))
			if _, _, err := readContract(root); err == nil {
				t.Fatal("invalid contract accepted")
			}
		})
	}
}

func TestContractPathsRejectEscapes(t *testing.T) {
	root := t.TempDir()
	for _, path := range []string{"", ".", "../outside", "/tmp/outside", "./file", "a/../file", "a//file", "a/./file", "a\\file", "a\x00file"} {
		if _, err := contractPath(root, path); err == nil {
			t.Errorf("invalid contract path accepted: %q", path)
		}
	}
	if _, err := contractPath(root, "docs/my file.txt"); err != nil {
		t.Fatal(err)
	}
}

func TestIntegrityDiscoversCurrentMarkdownWithoutGit(t *testing.T) {
	t.Setenv("L_HARNESS_ORIGINAL_GOFLAGS", "")
	root, c, hash := integrityFixture(t)
	for _, directory := range []string{
		".git", ".harness", ".idea", "bin", "app", "pkg", "logs",
		"website/.tools", "website/node_modules",
		"website/src/content/docs/docs", "website/src/content/docs/releases",
	} {
		writeIntegrityFile(t, root, directory+"/ignored.md", "[ignored](missing)")
	}
	outside := t.TempDir()
	writeIntegrityFile(t, outside, "outside.md", "[outside](missing)")
	if err := os.Symlink(outside, filepath.Join(root, "linked-documents")); err != nil {
		t.Fatal(err)
	}
	if report := inspectIntegrity(root, c, hash); !report.OK {
		t.Fatalf("ignored documents affected current integrity: %+v", report)
	}
	// Remove the fake Git directory: this is a source export, with no archive data.
	if err := os.RemoveAll(filepath.Join(root, ".git")); err != nil {
		t.Fatal(err)
	}
	if report := inspectIntegrity(root, c, hash); !report.OK {
		t.Fatalf("source export failed: %+v", report)
	}
	for _, name := range []string{
		"README.md", "docs/nested/guide.md", ".agents/skills/example/SKILL.md",
		"lexicon/README.md", "logiceme/testdata/README.md",
		"testdata/runtime/contract.md", "docs/bin/guide.md",
		"website/README.md", "website/src/content/docs/404.md", "website/node_modules-not-generated/guide.md",
	} {
		t.Run(name, func(t *testing.T) {
			writeIntegrityFile(t, root, name, "[broken](missing)")
			found := false
			for _, check := range inspectIntegrity(root, c, hash).Checks {
				if check.Kind == "document_link" && check.Path == name && !check.OK {
					found = true
				}
			}
			if !found {
				t.Fatal("undiscovered broken current-document link")
			}
			writeIntegrityFile(t, root, name, "[external](https://example.com/archive)\n")
		})
	}
	if report := inspectIntegrity(root, c, hash); !report.OK {
		t.Fatalf("repaired links failed: %+v", report)
	}
}

func TestMarkdownLinksHandleRelativePathsSpacesAndSourceAnchors(t *testing.T) {
	root := t.TempDir()
	for _, path := range []string{"source.go", "a document.txt"} {
		if err := os.WriteFile(filepath.Join(root, path), nil, 0644); err != nil {
			t.Fatal(err)
		}
	}
	contents := "[source](source.go#L10) [source](source.go:10) [space](<a document.txt>) [encoded](a%20document.txt)\n" +
		"[web](https://example.com/missing) [anchor](#section)\n```text\n[example](missing-in-code)\n```\n" +
		"`[example](missing-inline)`\n[broken](missing.go:10)"
	issues := checkMarkdownLinks(filepath.Join(root, "README.md"), contents)
	if len(issues) != 1 || !strings.Contains(issues[0].Error(), "missing.go") {
		t.Fatalf("incorrect local link findings: %v", issues)
	}
}
