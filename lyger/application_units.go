package Lyger

import (
	"fmt"

	"github.com/l-programming-language/l/lerrors"
)

// AppArtifact identifies a prebuilt companion relative to the orchestrator.
type AppArtifact struct {
	ID                 string
	RelativeExecutable string
}

// BuildUnit is one complete, canonical, single-application logeme document.
type BuildUnit struct {
	ID         string
	BackendAML []byte
}

func (r *CompilationResult) buildProgram() (*Program, error) {
	if r == nil || r.Compilation == nil || r.Compilation.Program == nil {
		return nil, lerrors.InvalidInput.Errorf("build units require a completed compilation")
	}
	if r.Compilation.Error != nil {
		return nil, fmt.Errorf("build units: %w", r.Compilation.Error)
	}
	return r.Compilation.Program, nil
}

// ApplicationUnits preserves the source graph and existing observable snapshots.
// A separate document for each book fits logeme's single-application API.
func (r *CompilationResult) ApplicationUnits() ([]BuildUnit, error) {
	program, err := r.buildProgram()
	if err != nil {
		return nil, err
	}
	apps, err := validateBuildGraph(program)
	if err != nil {
		return nil, err
	}
	units := make([]BuildUnit, 0, len(program.Books))
	for _, currentBook := range program.Books {
		app := apps[currentBook.Name]
		_, backend, err := renderProgram(&Program{Books: []*book{currentBook}}, []*App{app})
		if err != nil {
			return nil, err
		}
		units = append(units, BuildUnit{ID: app.buildName(), BackendAML: backend})
	}
	return units, nil
}

func validateBuildGraph(program *Program) (map[string]*App, error) {
	if program == nil {
		return nil, lerrors.Internal.Errorf("build units require an application book")
	}
	apps := make(map[string]*App, len(program.Apps))
	for _, app := range program.Apps {
		if app == nil || app.buildName() == "" || apps[app.buildName()] != nil {
			return nil, lerrors.Internal.Errorf("build graph has a nil, unnamed, or duplicate application")
		}
		apps[app.buildName()] = app
	}
	if len(program.Books) != len(apps) {
		return nil, lerrors.Internal.Errorf("application and book counts differ")
	}
	seen := make(map[string]bool)
	for _, currentBook := range program.Books {
		if currentBook == nil || apps[currentBook.Name] == nil || seen[currentBook.Name] {
			return nil, lerrors.Internal.Errorf("application book has no unique canonical application")
		}
		seen[currentBook.Name] = true
	}
	return apps, nil
}
