package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func (r *runner) installLinters() error {
	for _, name := range []string{"golangci-lint", "actionlint"} {
		if _, err := r.installTool(name); err != nil {
			return err
		}
	}
	return nil
}

func (r *runner) installedTool(name string) (string, error) {
	pin := r.pins.GolangCI
	if name == "actionlint" {
		pin = r.pins.Actionlint
	}
	directory := filepath.Join(r.root, ".harness", "tools", name, pin.Version, runtime.GOOS+"-"+runtime.GOARCH)
	binary := filepath.Join(directory, name)
	if name == "golangci-lint" {
		if candidate := os.Getenv("L_CI_GOLANGCI_LINT"); candidate != "" {
			if !filepath.IsAbs(candidate) {
				return "", errors.New("L_CI_GOLANGCI_LINT must be an absolute path")
			}
			binary = candidate
		}
	}
	data, err := os.ReadFile(binary)
	if err != nil {
		return "", fmt.Errorf("%s candidate %s is unavailable; run scripts/l-ci install or correct L_CI_GOLANGCI_LINT: %w", name, binary, err)
	}
	archive, err := os.ReadFile(filepath.Join(directory, "archive.tar.gz"))
	if err != nil {
		return "", err
	}
	if fmt.Sprintf("%x", sha256.Sum256(archive)) != pin.Checksums[runtime.GOOS+"-"+runtime.GOARCH] {
		return "", fmt.Errorf("cached %s archive checksum mismatch; run scripts/l-ci install", name)
	}
	expected, err := archiveBinary(archive, name)
	if err != nil {
		return "", err
	}
	if !bytes.Equal(data, expected) {
		return "", fmt.Errorf("%s candidate %s differs from verified release archive", name, binary)
	}
	if err := verifyToolVersion(binary, name, pin.Version); err != nil {
		return "", err
	}
	return binary, nil
}

func (r *runner) installTool(name string) (string, error) {
	platform := runtime.GOOS + "-" + runtime.GOARCH
	pin := r.pins.GolangCI
	if name == "actionlint" {
		pin = r.pins.Actionlint
	}
	digest, ok := pin.Checksums[platform]
	if !ok {
		return "", fmt.Errorf("no %s checksum for %s", name, platform)
	}
	if decoded, err := hex.DecodeString(digest); err != nil || len(decoded) != sha256.Size {
		return "", fmt.Errorf("invalid pinned checksum for %s/%s", name, platform)
	}
	dir := filepath.Join(r.root, ".harness", "tools", name, pin.Version, platform)
	binary := filepath.Join(dir, name)
	if installed, err := r.installedTool(name); err == nil {
		return installed, nil
	}
	archive := fmt.Sprintf("golangci-lint-%s-%s.tar.gz", pin.Version, platform)
	url := "https://github.com/golangci/golangci-lint/releases/download/v" + pin.Version + "/" + archive
	if name == "actionlint" {
		archive = fmt.Sprintf("actionlint_%s_%s_%s.tar.gz", pin.Version, runtime.GOOS, runtime.GOARCH)
		url = "https://github.com/rhysd/actionlint/releases/download/v" + pin.Version + "/" + archive
	}
	data, err := os.ReadFile(filepath.Join(dir, "archive.tar.gz"))
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(data)) != digest {
		data, err = download(url, digest)
		if err != nil {
			return "", err
		}
	}
	contents, err := archiveBinary(data, name)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(dir, "archive.tar.gz"), data, 0644); err != nil {
		return "", err
	}
	tmp, err := os.CreateTemp(dir, ".install-")
	if err != nil {
		return "", err
	}
	temporary := tmp.Name()
	_, writeErr := tmp.Write(contents)
	err = errors.Join(writeErr, tmp.Close())
	if err == nil {
		err = os.Chmod(temporary, 0755)
	}
	if err == nil {
		err = verifyToolVersion(temporary, name, pin.Version)
	}
	if err == nil {
		err = os.Rename(temporary, binary)
	}
	if err != nil {
		cleanupErr := os.Remove(temporary)
		return "", errors.Join(err, cleanupErr)
	}
	return r.installedTool(name)
}

func download(url, digest string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	data, readErr := io.ReadAll(io.LimitReader(response.Body, 256<<20))
	if err := errors.Join(readErr, response.Body.Close()); err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download %s: HTTP %d", url, response.StatusCode)
	}
	if fmt.Sprintf("%x", sha256.Sum256(data)) != digest {
		return nil, fmt.Errorf("checksum mismatch for %s", url)
	}
	return data, nil
}

func archiveBinary(data []byte, name string) (result []byte, resultErr error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer func() { resultErr = errors.Join(resultErr, gz.Close()) }()
	tr := tar.NewReader(gz)
	for {
		header, nextErr := tr.Next()
		if errors.Is(nextErr, io.EOF) {
			break
		}
		if nextErr != nil {
			return nil, nextErr
		}
		if filepath.Base(header.Name) != name {
			continue
		}
		if header.Typeflag != tar.TypeReg || result != nil || header.Size < 1 || header.Size > 200<<20 {
			return nil, errors.New("invalid or duplicate tool archive entry")
		}
		result, err = io.ReadAll(tr)
		if err != nil {
			return nil, err
		}
	}
	if result == nil {
		return nil, fmt.Errorf("archive has no %s executable", name)
	}
	return result, nil
}

func verifyToolVersion(path, name, version string) error {
	arg := "version"
	if name == "actionlint" {
		arg = "-version"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	output, err := exec.CommandContext(ctx, path, arg).CombinedOutput()
	if err != nil {
		return fmt.Errorf("verify %s: %w: %s", name, err, output)
	}
	text := string(output)
	if name == "actionlint" {
		if strings.SplitN(text, "\n", 2)[0] != version {
			return fmt.Errorf("unexpected actionlint version: %s", text)
		}
	} else if !strings.Contains(text, "version "+version+" ") {
		return fmt.Errorf("unexpected golangci-lint version: %s", text)
	}
	return nil
}
