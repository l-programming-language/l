package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func toolArchive(t *testing.T, headers ...tar.Header) []byte {
	t.Helper()
	return toolArchiveData(t, nil, headers...)
}

func toolArchiveData(t *testing.T, contents map[string][]byte, headers ...tar.Header) []byte {
	t.Helper()
	var buffer bytes.Buffer
	gz := gzip.NewWriter(&buffer)
	writer := tar.NewWriter(gz)
	for _, header := range headers {
		if err := writer.WriteHeader(&header); err != nil {
			t.Fatal(err)
		}
		if header.Typeflag == tar.TypeReg {
			data, supplied := contents[header.Name]
			if !supplied {
				data = bytes.Repeat([]byte("x"), int(header.Size))
			}
			if _, err := writer.Write(data); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func TestArchiveRejectsMissingDuplicateAndLinkedExecutables(t *testing.T) {
	regular := tar.Header{Name: "release/actionlint", Typeflag: tar.TypeReg, Size: 4, Mode: 0755}
	data, err := archiveBinary(toolArchive(t, regular), "actionlint")
	if err != nil || string(data) != "xxxx" {
		t.Fatalf("regular executable: %q %v", data, err)
	}
	linked := tar.Header{Name: "release/actionlint", Typeflag: tar.TypeSymlink, Linkname: "/bin/sh"}
	for name, data := range map[string][]byte{
		"missing": toolArchive(t), "duplicate": toolArchive(t, regular, regular), "symlink": toolArchive(t, linked), "malformed": []byte("not a gzip archive"),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := archiveBinary(data, "actionlint"); err == nil {
				t.Fatal("invalid tool archive accepted")
			}
		})
	}
}

type testTransport func(*http.Request) (*http.Response, error)

func (f testTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestDownloadRejectsChangedBytesAndHTTPFailures(t *testing.T) {
	original := http.DefaultClient
	t.Cleanup(func() { http.DefaultClient = original })
	status := http.StatusOK
	body := "verified release bytes"
	http.DefaultClient = &http.Client{Transport: testTransport(func(request *http.Request) (*http.Response, error) {
		if request.Method != http.MethodGet || request.Context().Err() != nil {
			t.Fatal("download did not use a live bounded GET request")
		}
		return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}
	digest := fmt.Sprintf("%x", sha256.Sum256([]byte(body)))
	if data, err := download("https://example.invalid/tool", digest); err != nil || string(data) != body {
		t.Fatalf("verified download: %q %v", data, err)
	}
	body = "substituted bytes"
	if _, err := download("https://example.invalid/tool", digest); err == nil {
		t.Fatal("checksum mismatch accepted")
	}
	status = http.StatusNotFound
	if _, err := download("https://example.invalid/tool", fmt.Sprintf("%x", sha256.Sum256([]byte(body)))); err == nil {
		t.Fatal("HTTP failure accepted despite matching body checksum")
	}
}

func TestInstalledToolRequiresPinnedArchiveAndMatchingBinary(t *testing.T) {
	root := t.TempDir()
	platform := runtime.GOOS + "-" + runtime.GOARCH
	archive := toolArchive(t, tar.Header{Name: "actionlint", Typeflag: tar.TypeReg, Size: 4, Mode: 0755})
	pin := toolPin{Version: "test", Checksums: map[string]string{platform: fmt.Sprintf("%x", sha256.Sum256(archive))}}
	r := runner{root: root, pins: toolPins{Actionlint: pin}}
	directory := filepath.Join(root, ".harness", "tools", "actionlint", "test", platform)
	testWrite(t, filepath.Join(directory, "archive.tar.gz"), archive)
	testWrite(t, filepath.Join(directory, "actionlint"), []byte("changed binary"))
	if _, err := r.installedTool("actionlint"); err == nil || !strings.Contains(err.Error(), "differs from verified") {
		t.Fatalf("modified cached binary was not rejected before execution: %v", err)
	}
	testWrite(t, filepath.Join(directory, "archive.tar.gz"), []byte("changed archive"))
	if _, err := r.installedTool("actionlint"); err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("modified cached archive accepted: %v", err)
	}
	if err := os.Remove(filepath.Join(directory, "archive.tar.gz")); err != nil {
		t.Fatal(err)
	}
	if _, err := r.installedTool("actionlint"); err == nil {
		t.Fatal("binary without verified archive accepted")
	}
}

func TestExplicitGolangciCandidateMustMatchPinnedArchive(t *testing.T) {
	root := t.TempDir()
	platform := runtime.GOOS + "-" + runtime.GOARCH
	body := []byte("#!/bin/sh\nprintf 'golangci-lint has version test built with fixture\\n'\n")
	archive := toolArchiveData(t, map[string][]byte{"release/golangci-lint": body}, tar.Header{Name: "release/golangci-lint", Typeflag: tar.TypeReg, Size: int64(len(body)), Mode: 0755})
	pin := toolPin{Version: "test", Checksums: map[string]string{platform: fmt.Sprintf("%x", sha256.Sum256(archive))}}
	r := runner{root: root, pins: toolPins{GolangCI: pin}}
	directory := filepath.Join(root, ".harness", "tools", "golangci-lint", "test", platform)
	local := filepath.Join(directory, "golangci-lint")
	candidate := filepath.Join(root, "action-installed", "golangci-lint")
	testWrite(t, filepath.Join(directory, "archive.tar.gz"), archive)
	for _, path := range []string{local, candidate} {
		testWrite(t, path, body)
		if err := os.Chmod(path, 0755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("L_CI_GOLANGCI_LINT", "")
	if selected, err := r.installedTool("golangci-lint"); err != nil || selected != local {
		t.Fatalf("local default not selected: %q %v", selected, err)
	}
	t.Setenv("L_CI_GOLANGCI_LINT", candidate)
	if selected, err := r.installedTool("golangci-lint"); err != nil || selected != candidate {
		t.Fatalf("verified action candidate not selected: %q %v", selected, err)
	}
	for name, path := range map[string]string{"missing": filepath.Join(root, "missing"), "relative": "golangci-lint", "mismatch": candidate} {
		t.Run(name, func(t *testing.T) {
			if name == "mismatch" {
				testWrite(t, candidate, []byte("different executable"))
			}
			t.Setenv("L_CI_GOLANGCI_LINT", path)
			if selected, err := r.installedTool("golangci-lint"); err == nil || selected != "" {
				t.Fatalf("explicit invalid candidate fell back: %q %v", selected, err)
			}
			if selected, err := r.installTool("golangci-lint"); err == nil || selected != "" {
				t.Fatalf("installation silently accepted a different binary: %q %v", selected, err)
			}
		})
	}
}
