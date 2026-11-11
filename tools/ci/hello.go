package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/l-programming-language/l/tools/internal/testgate"
)

// hello is the temporary source-library reset gate. It intentionally makes no
// full-module acceptance or retired runtime-suite claim.
func (r *runner) hello() (resultErr error) {
	contractPath := filepath.Join(r.root, "harness", "contract.json")
	contractHash, err := testgate.FileHash(contractPath)
	if err != nil {
		return err
	}
	if err := r.frozen("frozen-before"); err != nil {
		return err
	}
	defer func() {
		resultErr = errors.Join(resultErr, r.frozen("frozen-after"), requireUnchanged(contractPath, contractHash))
	}()
	parallel := ciParallelism(runtime.NumCPU())
	packages := parallel
	if packages > 3 {
		packages = 3
	}
	output, commandErr := r.command("hello-tests", testgate.NativePhaseTimeout, nil, "go", "test",
		fmt.Sprintf("-p=%d", packages), fmt.Sprintf("-parallel=%d", parallel), "-mod=readonly", "-count=1",
		fmt.Sprintf("-timeout=%ds", int(testgate.NativePackageTimeout/time.Second)), "-json",
		"-run=^(TestCoreHelloWorld|TestCoreHelloWorldRejectsRuntimeEffects|TestCoreHelloWorldUsesSelectedPrint|TestCompilerInitialize|TestCompilerCompile|TestDemoCompilationRetainsSemanticProgram|TestLucidTemplatesFrozen)$", "./cmd/l", "./lyger", "./logiceme")
	return errors.Join(commandErr, validateHelloOutcomes(output))
}

func (r *runner) frozen(name string) error {
	started := time.Now().UTC()
	output, commandErr := r.command(name, 7*time.Minute, nil, filepath.Join(r.root, "scripts", "l-harness"), "frozen")
	if err := errors.Join(commandErr, r.retainHarnessReport(output, "frozen", started)); err != nil {
		return err
	}
	return os.Rename(filepath.Join(r.directory, "frozen.report.json"), filepath.Join(r.directory, name+".report.json"))
}

// Require the actual named roots and every observed child to pass. Go exits zero
// when a selector matches nothing, so exit status alone cannot establish this gate.
func validateHelloOutcomes(data []byte) error {
	expected := map[string][]string{
		"github.com/l-programming-language/l/cmd/l":    {"TestCoreHelloWorld", "TestCoreHelloWorldRejectsRuntimeEffects", "TestCoreHelloWorldUsesSelectedPrint", "TestCompilerInitialize", "TestCompilerCompile"},
		"github.com/l-programming-language/l/lyger":    {"TestDemoCompilationRetainsSemanticProgram"},
		"github.com/l-programming-language/l/logiceme": {"TestLucidTemplatesFrozen"},
	}
	started, completed := map[string]bool{}, map[string]bool{}
	running, passed := map[string]bool{}, map[string]bool{}
	decoder := json.NewDecoder(bytes.NewReader(data))
	for {
		var event nativeEvent
		if err := decoder.Decode(&event); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("invalid hello test events: %w", err)
		}
		roots, known := expected[event.Package]
		if !known || completed[event.Package] {
			return fmt.Errorf("unexpected hello event package or completed package: %s", event.Package)
		}
		key := event.Package + "/" + event.Test
		if event.Test != "" {
			root, _, _ := strings.Cut(event.Test, "/")
			allowed := false
			for _, wanted := range roots {
				allowed = allowed || wanted == root
			}
			if !allowed {
				return fmt.Errorf("unexpected test in hello gate: %s", key)
			}
		}
		if event.Action == "start" {
			if started[event.Package] || event.Test != "" {
				return fmt.Errorf("invalid hello package start: %s", event.Package)
			}
			started[event.Package] = true
			continue
		}
		if !started[event.Package] {
			return fmt.Errorf("hello event precedes package start: %s", event.Package)
		}
		switch event.Action {
		case "output", "pause", "cont":
		case "run":
			if event.Test == "" || running[key] || passed[key] {
				return fmt.Errorf("invalid hello test start: %s", key)
			}
			if split := strings.LastIndexByte(event.Test, '/'); split >= 0 && !running[event.Package+"/"+event.Test[:split]] {
				return fmt.Errorf("hello child ran without parent: %s", key)
			}
			running[key] = true
		case "pass":
			for active := range running {
				if strings.HasPrefix(active, key+"/") || event.Test == "" && strings.HasPrefix(active, key) {
					return fmt.Errorf("hello parent passed before child: %s", active)
				}
			}
			if event.Test == "" {
				for _, root := range roots {
					if !passed[event.Package+"/"+root] {
						return fmt.Errorf("hello package passed without required test: %s", root)
					}
				}
				completed[event.Package] = true
			} else {
				if !running[key] {
					return fmt.Errorf("hello test passed without running: %s", key)
				}
				delete(running, key)
				passed[key] = true
			}
		default:
			return fmt.Errorf("unsuccessful or invalid hello test event %s: %s", event.Action, key)
		}
	}
	if len(completed) != len(expected) || len(running) != 0 {
		return errors.New("hello gate is missing completed named test outcomes")
	}
	return nil
}
