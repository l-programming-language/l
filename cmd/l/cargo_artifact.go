package main

import (
	"bytes"
	"debug/buildinfo"
	"debug/elf"
	"debug/macho"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/l-programming-language/l/lerrors"
)

func cargoArtifactInputManifest(actual, cached, original string) (string, error) {
	actual, err := filepath.EvalSymlinks(actual)
	if err != nil {
		return "", lerrors.Tool.Wrap(err, "resolve Cargo artifact manifest")
	}
	cached, err = filepath.EvalSymlinks(cached)
	if err != nil {
		return "", lerrors.IO.Wrap(err, "resolve cached Cargo input")
	}
	original, err = filepath.EvalSymlinks(original)
	if err != nil {
		return "", lerrors.IO.Wrap(err, "resolve original Cargo input")
	}
	relative, err := filepath.Rel(cached, actual)
	if err != nil || !filepath.IsLocal(relative) || filepath.Base(relative) != "Cargo.toml" {
		return "", lerrors.Tool.Errorf("Cargo artifact manifest is outside the supplied workspace")
	}
	input := filepath.Join(original, relative)
	want, err := os.ReadFile(input)
	if err != nil {
		return "", lerrors.Tool.Wrap(err, "read supplied Cargo artifact manifest")
	}
	got, err := os.ReadFile(filepath.Join(cached, relative))
	if err != nil || !bytes.Equal(got, want) {
		return "", lerrors.Tool.Errorf("Cargo artifact manifest differs from the supplied package: %v", err)
	}
	return input, nil
}

func copyCargoArtifact(source, destination string) (resultErr error) {
	info, err := os.Lstat(source)
	if err != nil {
		return lerrors.Tool.Errorf("Cargo executable is missing: %w", err)
	}
	if !info.Mode().IsRegular() || info.Size() == 0 || info.Mode().Perm()&0111 == 0 {
		return lerrors.Tool.Errorf("Cargo output is not a nonempty regular executable")
	}
	if _, err := buildinfo.ReadFile(source); err == nil {
		return lerrors.Tool.Errorf("Cargo returned a Go executable")
	}
	if runtime.GOOS == "darwin" {
		file, err := macho.Open(source)
		if err != nil {
			return lerrors.Tool.Errorf("Cargo did not produce a native Mach-O executable: %w", err)
		}
		defer func() { resultErr = closeCargoArtifact(file, resultErr) }()
		if file.Type != macho.TypeExec || runtime.GOARCH == "arm64" && file.Cpu != macho.CpuArm64 || runtime.GOARCH == "amd64" && file.Cpu != macho.CpuAmd64 {
			return lerrors.Tool.Errorf("Cargo executable has a different native architecture")
		}
	} else {
		file, err := elf.Open(source)
		if err != nil {
			return lerrors.Tool.Errorf("Cargo did not produce a native ELF executable: %w", err)
		}
		defer func() { resultErr = closeCargoArtifact(file, resultErr) }()
		if file.Type != elf.ET_EXEC && file.Type != elf.ET_DYN || runtime.GOARCH == "arm64" && file.Machine != elf.EM_AARCH64 || runtime.GOARCH == "amd64" && file.Machine != elf.EM_X86_64 {
			return lerrors.Tool.Errorf("Cargo executable has a different native architecture")
		}
	}
	if filepath.Clean(source) == filepath.Clean(destination) {
		return nil
	}
	content, err := os.ReadFile(source)
	if err != nil {
		return lerrors.IO.Wrap(err, "read Cargo artifact")
	}
	file, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0700)
	if err != nil {
		return lerrors.IO.Wrap(err, "publish Cargo artifact")
	}
	_, err = file.Write(content)
	return lerrors.IO.Wrap(errors.Join(err, file.Close()), "copy Cargo artifact")
}

func closeCargoArtifact(closer io.Closer, primary error) error {
	closeErr := closer.Close()
	if closeErr == nil {
		return primary
	}
	closeErr = lerrors.IO.Errorf("close Cargo artifact: %w", closeErr)
	if primary == nil {
		return closeErr
	}
	return errors.Join(primary, closeErr)
}
