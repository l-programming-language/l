package Lyger

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/l-programming-language/l/internal/cargonames"
	"github.com/l-programming-language/l/lerrors"
)

// WorkspacePackage identifies one existing application or constructed Service.
// A Service descriptor reused in two Apps creates two package identities: its
// bound Process body belongs to that construction, not to the descriptor name.
type WorkspacePackage struct {
	ID      string
	Library bool
}

type WorkspaceUnit struct {
	BuildUnit
	Library bool
}

type workspaceDependency struct{ alias, id string }
type workspacePackage struct {
	WorkspacePackage
	app          *App
	book         *book
	dependencies []workspaceDependency
}

// WorkspacePlan is a detached source projection. Locations belong to the build
// coordinator; this plan serializes ordinary AppMeta.Deps only after it receives
// those locations. It never edits compiled Rust or the observable source graph.
type WorkspacePlan struct{ packages []workspacePackage }

func workspaceServiceLibraryID(appID string, index int, used map[string]bool) string {
	id := fmt.Sprintf("l_service_%x_%d", []byte(appID), index)
	for used[id] {
		id += "_"
	}
	used[id] = true
	return id
}

func (r *CompilationResult) WorkspacePlan() (*WorkspacePlan, error) {
	program, err := r.buildProgram()
	if err != nil {
		return nil, err
	}
	apps, err := validateBuildGraph(program)
	if err != nil {
		return nil, err
	}
	plan := &WorkspacePlan{}
	used := map[string]bool{}
	for _, app := range program.Apps {
		used[app.buildName()] = true
	}
	for _, current := range program.Books {
		app := apps[current.Name]
		if len(app.Services) != len(current.Chapters) {
			return nil, lerrors.Internal.Errorf("workspace Service and Chapter dimensions disagree")
		}
		if len(app.Services) == 1 {
			clonedApp := *app
			clonedApp.Meta = cloneWorkspaceMetadata(app.Meta)
			plan.packages = append(plan.packages, workspacePackage{WorkspacePackage: WorkspacePackage{ID: app.buildName()}, app: &clonedApp, book: current})
			continue
		}
		primary := -1
		for index, service := range app.Services {
			if service.Target != nil && service.Target.Kind == "bin" && service.Target.Primary {
				if primary != -1 {
					return nil, lerrors.InvalidInput.Errorf("workspace application %s has multiple primary Services", app.buildName())
				}
				primary = index
			}
		}
		if primary == -1 {
			return nil, lerrors.InvalidInput.Errorf("workspace application %s has no primary Service", app.buildName())
		}
		application := workspaceServicePackage(app, current, primary, app.buildName(), false)
		var libraries []workspacePackage
		aliases := map[string]bool{}
		for index, service := range app.Services {
			if index == primary {
				continue
			}
			if service.Target == nil || service.Target.Kind != "lib" || service.Target.Primary {
				return nil, lerrors.InvalidInput.Errorf("non-main Service %s requires a library target", service.ID)
			}
			id := workspaceServiceLibraryID(app.buildName(), index, used)
			alias := sourceModuleName(service.ID)
			if aliases[alias] {
				return nil, lerrors.InvalidInput.Errorf("workspace application %s has duplicate Service alias %s", app.buildName(), alias)
			}
			aliases[alias] = true
			application.dependencies = append(application.dependencies, workspaceDependency{alias, id})
			libraries = append(libraries, workspaceServicePackage(app, current, index, id, true))
		}
		plan.packages = append(plan.packages, application)
		plan.packages = append(plan.packages, libraries...)
	}
	plan.allocateCargoBinaryNames()
	return plan, nil
}

// Cargo binary names belong to this detached workspace projection. Canonical
// package IDs still connect source Apps to orchestration and publication.
func (p *WorkspacePlan) allocateCargoBinaryNames() {
	occupied := make(map[string]bool)
	for _, item := range p.packages {
		occupied[strings.ToLower(item.ID)] = true
		occupied[strings.ToLower(item.book.Name)] = true
		for _, service := range item.app.Services {
			if service.Target != nil {
				occupied[strings.ToLower(service.Target.Name)] = true
			}
		}
	}
	selectedBinary := make(map[string]bool)
	for index := range p.packages {
		item := &p.packages[index]
		if item.Library {
			continue
		}
		binary := item.book.Name
		for _, service := range item.app.Services {
			if service.Target != nil && service.Target.Kind == "bin" && service.Target.Primary {
				binary = service.Target.Name
			}
		}
		binaryKey := strings.ToLower(binary)
		if !cargonames.IsReservedBinary(binary) && !selectedBinary[binaryKey] {
			selectedBinary[binaryKey] = true
			continue
		}
		name := "l_app"
		for suffix := 1; occupied[strings.ToLower(name)]; suffix++ {
			name = fmt.Sprintf("l_app_%d", suffix)
		}
		occupied[strings.ToLower(name)] = true
		selectedBinary[strings.ToLower(name)] = true
		app, content := *item.app, *item.book
		app.Meta = cloneWorkspaceMetadata(app.Meta)
		app.buildID, app.Meta["AppName"], content.Name = name, name, name
		app.Services = make([]*service, len(item.app.Services))
		for serviceIndex, original := range item.app.Services {
			clonedService := cloneWorkspaceService(original, false)
			if clonedService.Target != nil && clonedService.Target.Kind == "bin" && clonedService.Target.Primary {
				clonedService.Target.Name = name
			}
			app.Services[serviceIndex] = clonedService
		}
		item.app, item.book = &app, &content
	}
}

func (p *WorkspacePlan) Packages() []WorkspacePackage {
	if p == nil {
		return nil
	}
	result := make([]WorkspacePackage, len(p.packages))
	for index, item := range p.packages {
		result[index] = item.WorkspacePackage
	}
	return result
}

func (p *WorkspacePlan) Units(locations map[string]string) ([]WorkspaceUnit, error) {
	if p == nil || len(locations) != len(p.packages) {
		return nil, lerrors.InvalidInput.Errorf("workspace locations must cover every package exactly once")
	}
	seen := map[string]bool{}
	for _, item := range p.packages {
		location := locations[item.ID]
		if !fs.ValidPath(location) || location == "." || strings.ContainsAny(location, "\\:\x00") || seen[location] {
			return nil, lerrors.InvalidInput.Errorf("invalid or duplicate workspace location for %s", item.ID)
		}
		seen[location] = true
	}
	for location := range seen {
		for parent := path.Dir(location); parent != "."; parent = path.Dir(parent) {
			if seen[parent] {
				return nil, lerrors.InvalidInput.Errorf("workspace package locations overlap")
			}
		}
	}
	units := make([]WorkspaceUnit, 0, len(p.packages))
	for _, item := range p.packages {
		app := *item.app
		app.Meta = cloneWorkspaceMetadata(item.app.Meta)
		rows, names, err := workspaceDependencyRows(app.Meta)
		if err != nil {
			return nil, fmt.Errorf("workspace package %s: %w", item.ID, err)
		}
		if len(item.dependencies) != 0 {
			for _, dependency := range item.dependencies {
				if names[dependency.alias] {
					return nil, lerrors.InvalidInput.Errorf("workspace package %s dependency %q collides with constructed Service", item.ID, dependency.alias)
				}
				names[dependency.alias] = true
				relative, err := filepath.Rel(filepath.FromSlash(locations[item.ID]), filepath.FromSlash(locations[dependency.id]))
				if err != nil {
					return nil, lerrors.InvalidInput.Wrap(err, "workspace dependency location")
				}
				rows = append(rows, dependency.alias+" = { package = "+strconv.Quote(dependency.id)+", path = "+strconv.Quote(filepath.ToSlash(relative))+" }")
			}
			encoded, err := json.Marshal(rows)
			if err != nil {
				return nil, err
			}
			app.Meta["Deps"] = string(encoded)
		}
		_, backend, err := renderProgram(&Program{Books: []*book{item.book}}, []*App{&app})
		if err != nil {
			return nil, err
		}
		units = append(units, WorkspaceUnit{BuildUnit: BuildUnit{ID: item.ID, BackendAML: backend}, Library: item.Library})
	}
	return units, nil
}

// AppMeta.Deps already contains a JSON array of complete normalized TOML
// assignments. Preserve those native rows, including nested feature-list
// commas; only validate the existing framing/key boundary before adding paths.
// Cargo and the backend continue to own dependency semantics and resolution.
func workspaceDependencyRows(meta map[string]string) ([]string, map[string]bool, error) {
	names := make(map[string]bool)
	raw, present := meta["Deps"]
	if !present {
		return nil, names, nil
	}
	var rows []string
	if err := json.Unmarshal([]byte(raw), &rows); err != nil {
		return nil, nil, lerrors.InvalidInput.Errorf("AppMeta.Deps requires a JSON-encoded array of TOML dependency rows: %w", err)
	}
	if rows == nil {
		return nil, nil, lerrors.InvalidInput.Errorf("AppMeta.Deps must be an array, not null")
	}
	for _, row := range rows {
		key, value, assignment := strings.Cut(row, "=")
		key = strings.TrimSpace(key)
		if !assignment || key == "" || strings.TrimSpace(value) == "" || strings.TrimSpace(row) != row || strings.ContainsAny(row, "\r\n\x00") {
			return nil, nil, lerrors.InvalidInput.Errorf("AppMeta.Deps entry must be one normalized TOML dependency assignment: %q", row)
		}
		for _, character := range key {
			if !(character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9' || character == '_' || character == '-') {
				return nil, nil, lerrors.InvalidInput.Errorf("AppMeta.Deps entry requires a bare dependency key: %q", key)
			}
		}
		if names[key] {
			return nil, nil, lerrors.InvalidInput.Errorf("AppMeta.Deps contains duplicate dependency %q", key)
		}
		names[key] = true
	}
	return rows, names, nil
}

func workspaceServicePackage(original *App, current *book, index int, id string, library bool) workspacePackage {
	app, content := *original, *current
	app.Meta = cloneWorkspaceMetadata(original.Meta)
	app.ID, app.buildID, app.Meta["AppName"] = id, "", id
	app.Services = []*service{cloneWorkspaceService(original.Services[index], library)}
	if library {
		app.Settings = nil
		app.Services[0].Target.Name = id
	}
	content.Name = id
	content.Chapters = []*chapter{current.Chapters[index]}
	return workspacePackage{WorkspacePackage: WorkspacePackage{ID: id, Library: library}, app: &app, book: &content}
}

func cloneWorkspaceMetadata(meta map[string]string) map[string]string {
	clonedMetadata := make(map[string]string, len(meta))
	for name, value := range meta {
		clonedMetadata[name] = value
	}
	return clonedMetadata
}

func cloneWorkspaceService(original *service, public bool) *service {
	clonedService := *original
	clonedService.Meta = cloneWorkspaceMetadata(original.Meta)
	if original.Target != nil {
		target := *original.Target
		clonedService.Target = &target
	}
	var cloneMembers func([]sourceMember) ([]sourceMember, []*object, []*class)
	cloneMembers = func(members []sourceMember) ([]sourceMember, []*object, []*class) {
		var order []sourceMember
		var objects []*object
		var classes []*class
		for _, member := range members {
			if member.Object != nil {
				value := *member.Object
				value.Meta = cloneWorkspaceMetadata(value.Meta)
				if value.Module != nil {
					module := *value.Module
					if public && module.Visibility != "private" {
						module.Visibility = "public"
					}
					value.Module = &module
				}
				objects = append(objects, &value)
				order = append(order, sourceMember{Object: &value})
			} else {
				value := *member.Class
				value.Meta = cloneWorkspaceMetadata(value.Meta)
				if value.Module != nil {
					module := *value.Module
					if public && module.Visibility != "private" {
						module.Visibility = "public"
					}
					value.Module = &module
				}
				value.members, value.Objects, value.Children = cloneMembers(value.members)
				classes = append(classes, &value)
				order = append(order, sourceMember{Class: &value})
			}
		}
		return order, objects, classes
	}
	clonedService.members, clonedService.Objects, clonedService.Classes = cloneMembers(original.members)
	return &clonedService
}
