// l-ci supplies the same checked entrypoints for local development and CI.
package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/l-programming-language/l/tools/internal/testgate"
)

type toolPin struct {
	Version   string            `json:"version"`
	Checksums map[string]string `json:"checksums"`
}

type toolPins struct {
	Go         string  `json:"go"`
	Rust       string  `json:"rust"`
	GolangCI   toolPin `json:"golangci-lint"`
	Actionlint toolPin `json:"actionlint"`
}

type stageReport struct {
	Stage          string          `json:"stage"`
	StartedAt      time.Time       `json:"started_at"`
	FinishedAt     time.Time       `json:"finished_at"`
	Revision       string          `json:"revision"`
	SourceIdentity string          `json:"source_identity"`
	Dirty          bool            `json:"dirty"`
	GOOS           string          `json:"goos"`
	GOARCH         string          `json:"goarch"`
	Passed         bool            `json:"passed"`
	Error          string          `json:"error,omitempty"`
	Commands       []commandResult `json:"commands"`
}

type runner struct {
	root      string
	directory string
	pins      toolPins
	report    stageReport
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		_, writeErr := fmt.Fprintln(os.Stderr, err)
		if writeErr != nil {
			os.Exit(2)
		}
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) != 1 {
		return errors.New("usage: scripts/l-ci metadata|install|bootstrap|lint|acceptance|native|check")
	}
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	pins, err := readPins(root)
	if err != nil {
		return err
	}
	if args[0] == "metadata" {
		return metadata(root, pins)
	}
	stages := []string{args[0]}
	if args[0] == "check" {
		stages = []string{"lint", "acceptance", "native"}
	}
	for _, stage := range stages {
		if err := runStage(root, pins, stage); err != nil {
			return err
		}
	}
	return nil
}

func readPins(root string) (toolPins, error) {
	var pins toolPins
	if err := readJSON(filepath.Join(root, ".ci-tools.json"), &pins); err != nil {
		return pins, err
	}
	version, err := os.ReadFile(filepath.Join(root, ".go-version"))
	if err != nil {
		return pins, err
	}
	if pins.Go == "" || strings.TrimSpace(string(version)) != pins.Go || runtime.Version() != "go"+pins.Go {
		return pins, errors.New(".ci-tools.json, .go-version and the running Go toolchain must agree")
	}
	if pins.Rust == "" || pins.GolangCI.Version == "" || pins.Actionlint.Version == "" {
		return pins, errors.New("incomplete tool pins")
	}
	return pins, nil
}

func runStage(root string, pins toolPins, stage string) error {
	if stage != "install" && stage != "bootstrap" && stage != "lint" && stage != "acceptance" && stage != "native" {
		return fmt.Errorf("unknown CI stage %q", stage)
	}
	runs := filepath.Join(root, ".harness", "runs")
	if err := os.MkdirAll(runs, 0755); err != nil {
		return err
	}
	directory, err := os.MkdirTemp(runs, "ci-"+stage+"-")
	if err != nil {
		return err
	}
	revision, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err != nil {
		return fmt.Errorf("read revision: %w", err)
	}
	_, sourceIdentity, err := cacheDigests(root)
	if err != nil {
		return err
	}
	status, err := exec.Command("git", "status", "--porcelain").Output()
	if err != nil {
		return err
	}
	r := runner{root: root, directory: directory, pins: pins, report: stageReport{Stage: stage, StartedAt: time.Now().UTC(), Revision: strings.TrimSpace(string(revision)), SourceIdentity: sourceIdentity, Dirty: len(status) != 0, GOOS: runtime.GOOS, GOARCH: runtime.GOARCH, Commands: []commandResult{}}}
	_, printErr := fmt.Printf("%s: %s\n", stage, directory)
	if printErr != nil {
		return printErr
	}
	switch stage {
	case "install":
		err = r.installLinters()
	case "bootstrap":
		err = r.bootstrap()
	case "lint":
		err = r.lint()
	case "acceptance":
		err = r.acceptance()
	case "native":
		err = r.native()
	}
	r.report.FinishedAt, r.report.Passed = time.Now().UTC(), err == nil
	if err != nil {
		r.report.Error = err.Error()
	}
	err = errors.Join(err, testgate.WriteJSON(filepath.Join(directory, "stage.json"), r.report))
	_, printErr = fmt.Printf("%s passed: %t; report: %s\n", stage, err == nil, filepath.Join(directory, "stage.json"))
	return errors.Join(err, printErr)
}

func metadata(root string, pins toolPins) error {
	dependency, source, err := cacheDigests(root)
	if err != nil {
		return err
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		return err
	}
	command := exec.Command(filepath.Join(runtime.GOROOT(), "bin", "go"), "env", "GOMODCACHE")
	command.Dir, command.Env = root, environment(root, nil)
	moduleCache, err := command.Output()
	if err != nil {
		return fmt.Errorf("read pinned Go module cache: %w", err)
	}
	values := []string{"go=" + pins.Go, "rust=" + pins.Rust, "golangci-lint=" + pins.GolangCI.Version, "actionlint=" + pins.Actionlint.Version, "dependency-key=" + dependency, "source-key=" + source, "gomodcache=" + strings.TrimSpace(string(moduleCache)), "native-cache=" + filepath.Join(cache, "l", "rust", pins.Rust, runtime.GOOS+"-"+runtime.GOARCH)}
	text := strings.Join(values, "\n") + "\n"
	if output := os.Getenv("GITHUB_OUTPUT"); output != "" {
		file, err := os.OpenFile(output, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		_, writeErr := file.WriteString(text)
		if err := errors.Join(writeErr, file.Close()); err != nil {
			return err
		}
	}
	_, err = fmt.Print(text)
	return err
}
