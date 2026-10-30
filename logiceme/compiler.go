// Author: Connor Sanders

package Logiceme

import (
	"fmt"

	"github.com/l-programming-language/l/lerrors"
)

type OutlineEntry struct {
	Root           string            `json:"Root,omitempty"`
	Module         *ModuleDescriptor `json:"Module,omitempty"`
	Target         *TargetDescriptor `json:"Target,omitempty"`
	Meta           MetaDict          `json:"Meta,omitempty"`
	Name           string            `json:"Name,omitempty"`
	Type           string            `json:"Type,omitempty"`
	EntryType      string            `json:"EntryType,omitempty"`
	Location       string            `json:"Location,omitempty"`
	Parent         string            `json:"Parent,omitempty"`
	GenericObjects []string          `json:"GenericObjects,omitempty"`
	DynamicObjects []string          `json:"DynamicObjects,omitempty"`
	Dimensions     []string          `json:"Dimensions,omitempty"`
}

type Outline struct {
	Name         string         `json:"Name,omitempty"`
	Type         string         `json:"Type,omitempty"`
	Meta         MetaDict       `json:"AppMeta,omitempty"`
	Integrations []OutlineEntry `json:"Integrations,omitempty"`
	Resources    []OutlineEntry `json:"Resources,omitempty"`
	Datasources  []OutlineEntry `json:"Datasources,omitempty"`
	Services     []OutlineEntry `json:"Services,omitempty"`
	Classes      []OutlineEntry `json:"Classes,omitempty"`
	Objects      []Object       `json:"Objects,omitempty"`
}

func (out *Outline) BuildSubOutline(name string, outlineType string) Outline {
	entry := out.GetEntry(name, outlineType)
	result := Outline{Name: name, Type: outlineType, Meta: out.Meta, Classes: append([]OutlineEntry(nil), out.Classes...), Objects: append([]Object(nil), out.Objects...)}
	if outlineType == "Service" {
		result.Services = []OutlineEntry{entry}
	}
	for _, object := range out.GetEntryObjects(entry, outlineType) {
		for i := range result.Objects {
			if result.Objects[i].Name == object.Name {
				result.Objects[i] = object
			}
		}
	}
	return result
}

func (out *Outline) entries(entryType string) []OutlineEntry {
	switch entryType {
	case "Service":
		return out.Services
	case "Class":
		return out.Classes
	}
	return nil
}

func (out *Outline) getChildren(entryName string, entryType string) []OutlineEntry {
	children := []OutlineEntry{}
	for _, entry := range out.entries(entryType) {
		if entry.Parent == entryName {
			children = append(children, entry)
		}
	}
	return children
}

func (out *Outline) GetEntryObjects(entry OutlineEntry, entryType string) []Object {
	var reObjects []Object
	var obj Object
	for _, gObj := range entry.GenericObjects {
		obj = out.GetObject(gObj)
		obj.Location = entry.Location
		reObjects = append(reObjects, obj)
	}
	for _, dObj := range entry.DynamicObjects {
		obj = out.GetObject(dObj)
		reObjects = append(reObjects, obj)
	}
	for _, child := range out.getChildren(entry.Name, entryType) {
		reObjects = append(reObjects, out.GetEntryObjects(child, entryType)...)
	}
	return reObjects
}

// GetObject from Outline
func (out *Outline) GetObject(name string) Object {
	objects := out.Objects
	for _, object := range objects {
		if object.Name == name {
			return object
		}
	}
	return Object{}
}

// GetEntry from Outline
func (out *Outline) GetEntry(name string, entryType string) OutlineEntry {
	for _, entry := range out.entries(entryType) {
		if entry.Name == name {
			return entry
		}
	}
	return OutlineEntry{}
}

type DataSource struct {
	Name      string
	Type      string
	Settings  MetaDict
	Caches    []MetaDict
	CachesMap []string
	Meta      MetaDict
	Inputs    [][]string
}

type Integration struct {
	Name      string
	Type      string
	Settings  MetaDict
	Meta      MetaDict
	Inputs    [][]string
	Resources [][]byte
}

type Class struct {
	members          []sourceMember
	structureVersion string
	Name             string
	Type             string
	Private          bool
	Meta             MetaDict
	Inputs           []string
	Integrations     []Integration
	Outline          Outline
	Objects          []Object
	SubClasses       []Class
}

type CompiledClass struct {
	members          []sourceMember
	structureVersion string
	Name             string
	Type             string
	SegmentalSystem  SegmentalSystem
	Private          bool
	Meta             MetaDict
	Inputs           []string
	Integrations     []Integration
	Outline          Outline
	Objects          []CompiledObject
	SubClasses       []CompiledClass
	Tensor           Tensor
	Dependencies     []string
	EnvDependencies  []string
	ContentMap       []string
	Content          [][]byte
}

func (cc *CompiledClass) Link() {
	cc.Content, cc.ContentMap, cc.Dependencies, cc.EnvDependencies = linkSourceMembers(cc.structureVersion, cc.members, cc.Objects, cc.SubClasses, cc.Content, cc.ContentMap, cc.Dependencies, cc.EnvDependencies)
}

func linkSourceMembers(version string, members []sourceMember, objects []CompiledObject, classes []CompiledClass, content [][]byte, paths, dependencies, environment []string) ([][]byte, []string, []string, []string) {
	appendObject := func(co CompiledObject) {
		content = append(content, co.Content)
		paths = append(paths, co.Name)
		dependencies = append(dependencies, co.Dependencies...)
		environment = append(environment, co.EnvDependencies...)
	}
	appendClass := func(child CompiledClass) {
		child.Link()
		content = append(content, child.Content...)
		paths = append(paths, child.ContentMap...)
		dependencies = append(dependencies, child.Dependencies...)
		environment = append(environment, child.EnvDependencies...)
	}
	if version == "2" {
		for _, member := range members {
			if member.Kind == "Object" {
				appendObject(objects[member.Index])
			} else {
				appendClass(classes[member.Index])
			}
		}
		return content, paths, dependencies, environment
	}
	for _, object := range objects {
		appendObject(object)
	}
	for _, child := range classes {
		appendClass(child)
	}
	return content, paths, dependencies, environment
}

// Assemble a Class
func (c *Class) Assemble(meta MetaDict, segSys SegmentalSystem) (CompiledClass, error) {
	var combinedMeta MetaDict
	if c.structureVersion == "2" {
		combinedMeta = sourceMeta(meta, c.Meta)
	} else {
		c.Meta.SetAdd("Package", c.Name, "String")
		combinedMeta = c.Meta.Combine(meta)
	}
	cClass := CompiledClass{
		members: append([]sourceMember(nil), c.members...), structureVersion: c.structureVersion,
		Name:            c.Name,
		Type:            c.Type,
		SegmentalSystem: segSys,
		Private:         c.Private,
		Meta:            combinedMeta,
		Inputs:          c.Inputs,
		Integrations:    c.Integrations,
		Outline:         c.Outline,
		Objects:         []CompiledObject{},
		SubClasses:      []CompiledClass{},
	}
	for _, child := range c.SubClasses {
		compiled, err := child.Assemble(combinedMeta, segSys)
		if err != nil {
			return CompiledClass{}, err
		}
		cClass.SubClasses = append(cClass.SubClasses, compiled)
	}
	// Compile Objects
	for _, object := range c.Objects {
		cObject, oErr := object.Assemble(combinedMeta, segSys)
		if oErr != nil {
			return cClass, oErr
		}
		cClass.Objects = append(cClass.Objects, cObject)
	}
	return cClass, nil
}

type Service struct {
	members          []sourceMember
	structureVersion string
	Name             string
	Type             string
	Private          bool
	Settings         MetaDict
	Meta             MetaDict
	Inputs           []string
	Integrations     []Integration
	Outline          Outline
	Classes          []Class
	Objects          []Object
	SubServices      []Service
}

type CompiledService struct {
	members          []sourceMember
	structureVersion string
	Name             string
	Type             string
	SegmentalSystem  SegmentalSystem
	Private          bool
	Settings         MetaDict
	Meta             MetaDict
	Inputs           []string
	Integrations     []Integration
	Outline          Outline
	Classes          []CompiledClass
	Objects          []CompiledObject
	SubServices      []CompiledService
	Tensor           Tensor
	Dependencies     []string
	EnvDependencies  []string
	ContentMap       []string
	Content          [][]byte
}

func (cs *CompiledService) Link() {
	cs.Content, cs.ContentMap, cs.Dependencies, cs.EnvDependencies = linkSourceMembers(cs.structureVersion, cs.members, cs.Objects, cs.Classes, cs.Content, cs.ContentMap, cs.Dependencies, cs.EnvDependencies)
}

// Assemble a Service
func (s *Service) Assemble(meta MetaDict, segSys SegmentalSystem) (CompiledService, error) {
	var combinedMeta MetaDict
	if s.structureVersion == "2" {
		combinedMeta = sourceMeta(meta, s.Meta)
		if combinedMeta.GetVal("Package") == "NotFound" {
			combinedMeta.SetAdd("Package", s.Name, "String")
		}
	} else {
		s.Meta.SetAdd("Package", s.Name, "String")
		combinedMeta = s.Meta.Combine(meta)
	}
	cService := CompiledService{
		members: append([]sourceMember(nil), s.members...), structureVersion: s.structureVersion,
		Name:            s.Name,
		Type:            s.Type,
		SegmentalSystem: segSys,
		Private:         s.Private,
		Settings:        s.Settings,
		Meta:            combinedMeta,
		Inputs:          s.Inputs,
		Integrations:    s.Integrations,
		Outline:         s.Outline,
		Classes:         []CompiledClass{},
		Objects:         []CompiledObject{},
		SubServices:     []CompiledService{},
	}
	// Assemble Classes
	for _, class := range s.Classes {
		cClass, cErr := class.Assemble(combinedMeta, segSys)
		if cErr != nil {
			return cService, cErr
		}
		cService.Classes = append(cService.Classes, cClass)
	}
	// Assemble Objects
	for _, object := range s.Objects {
		cObject, oErr := object.Assemble(combinedMeta, segSys)
		if oErr != nil {
			return cService, oErr
		}
		cService.Objects = append(cService.Objects, cObject)
	}
	return cService, nil
}

func outlinedObjects(sentences []Sentence, outline Outline) []Object {
	objects := BuildObjects(sentences)
	for index := range objects {
		objects[index].Location = outline.GetObject(objects[index].Name).Location
	}
	return objects
}

func (s *Service) BuildClass(text Text) {
	re := Class{
		Name:    text.Name,
		Meta:    text.Context.GetDict("Meta"),
		Outline: s.Outline.BuildSubOutline(text.Name, "Class"),
	}
	re.Objects = outlinedObjects(text.Sentences, re.Outline)
	s.Classes = append(s.Classes, re)
}

type App struct {
	sourceTargets   []preparedSourceTarget
	Name            string
	Language        string
	SegmentalSystem SegmentalSystem
	Settings        MetaDict
	Meta            MetaDict
	Integrations    []Integration
	DataSources     []DataSource
	Outline         Outline
	Services        []Service
}

type CompiledApp struct {
	sourceTargets   []preparedSourceTarget
	Name            string
	Language        string
	SegmentalSystem SegmentalSystem
	Settings        MetaDict
	Meta            MetaDict
	Integrations    []Integration
	DataSources     []DataSource
	Outline         Outline
	Services        []CompiledService
	Tensor          Tensor
	Dependencies    []string
	EnvDependencies []string
	ContentMap      []string
	Content         [][]byte
}

// Link links the generator source code together, generates file headers, & executes any post-processes
func (ca *CompiledApp) Link() {
	for _, cs := range ca.Services {
		cs.Link()
		ca.Content = append(ca.Content, cs.Content...)
		ca.ContentMap = append(ca.ContentMap, cs.ContentMap...)
		ca.Dependencies = append(ca.Dependencies, cs.Dependencies...)
		ca.EnvDependencies = append(ca.EnvDependencies, cs.EnvDependencies...)
	}
	removeDuplicateHeaders(&ca.Dependencies)
	removeDuplicateHeaders(&ca.EnvDependencies)
}

func (a *App) BuildService(chap Chapter) {
	newService := Service{
		Name:     chap.Name,
		Private:  true,
		Settings: chap.Context.GetDict("Settings"),
		Meta:     chap.Context.GetDict("Meta"),
		Outline:  a.Outline.BuildSubOutline(chap.Name, "Service"),
	}
	newService.Objects = outlinedObjects(chap.Sentences, newService.Outline)
	for _, text := range chap.Texts {
		newService.BuildClass(text)
	}
	a.Services = append(a.Services, newService)
}

func (a *App) BuildServices(chaps []Chapter) error {
	for _, chap := range chaps {
		a.BuildService(chap)
	}
	return nil
}

// Assemble App
func (a *App) Assemble(meta MetaDict) (CompiledApp, error) {
	// Assemble CompiledApp
	combinedMeta := a.Meta.Combine(meta)
	cApp := CompiledApp{
		sourceTargets:   append([]preparedSourceTarget(nil), a.sourceTargets...),
		Name:            a.Name,
		Language:        a.Language,
		SegmentalSystem: a.SegmentalSystem,
		Settings:        a.Settings,
		Meta:            combinedMeta,
		Integrations:    a.Integrations,
		DataSources:     a.DataSources,
		Outline:         a.Outline,
		Services:        []CompiledService{},
	}
	// Assemble Services
	for _, service := range a.Services {
		cService, cErr := service.Assemble(combinedMeta, a.SegmentalSystem)
		if cErr != nil {
			if a.Language == "Rust" {
				return cApp, fmt.Errorf("Rust application %q: %w", a.Name, cErr)
			}
			return cApp, cErr
		}
		cApp.Services = append(cApp.Services, cService)
	}
	return cApp, nil
}

func BuildApp(book Book) (App, error) {
	reApp := App{
		Name:         book.Name,
		Language:     "Rust",
		Settings:     book.Settings,
		Meta:         book.Meta,
		Outline:      book.Outline,
		DataSources:  book.DataSources,
		Integrations: book.Integrations,
	}
	if book.Outline.Meta.GetVal("StructureVersion") == "2" {
		return buildSourceGraphApp(reApp, book)
	}
	if version := book.Outline.Meta.GetVal("StructureVersion"); version != "NotFound" && version != "" {
		return App{}, lerrors.InvalidInput.Errorf("unsupported StructureVersion %q", version)
	}
	for _, chapter := range book.Content {
		if len(chapter.Texts) > 0 {
			return App{}, lerrors.InvalidInput.Errorf("populated Classes require AppMeta.StructureVersion 2")
		}
	}
	cErr := reApp.BuildServices(book.Content)
	if cErr != nil {
		return App{}, cErr
	}
	return reApp, nil
}

// Executable - represents all source code generated for an App ready to be saved to a file system & compiled.
type Executable struct {
	Name             string
	ByteArrays       [][][]byte
	MapArrays        [][]string
	DependencyArrays [][]string
	EnvDepArrays     [][]string
}

// Compiler - structures & processes all program information related to compiling an Executable
type Compiler struct {
	Name            string
	Meta            MetaDict
	SegmentalSystem SegmentalSystem
	Books           []Book
	Apps            []App
	CompiledApps    []CompiledApp
}

// Compile - Builds the dimensionality of the app & compiles logiceme contents
func (c *Compiler) Compile() error {
	apps := []App{}
	for _, book := range c.Books {
		build, buildErr := BuildApp(book)
		if buildErr != nil {
			return buildErr
		}
		build.SegmentalSystem = c.SegmentalSystem
		apps = append(apps, build)
	}
	c.Apps = apps
	return nil
}

// Assemble - Assembles the source code by injecting meta data & eliminating content from un-used logiceme cases.
func (c *Compiler) Assemble() error {
	compiledApps := []CompiledApp{}
	for _, app := range c.Apps {
		cApp, cErr := app.Assemble(c.Meta)
		if cErr != nil {
			return cErr
		}
		compiledApps = append(compiledApps, cApp)
	}
	c.CompiledApps = compiledApps
	return nil
}

// Program represents an entire CGU Cloud App
type Program struct {
	Name            string
	Type            string
	Description     string
	SegmentalSystem SegmentalSystem
	Scanner         Scanner
	Bytes           []byte
	Settings        MetaDict
	Inputs          []MetaDict
	Compiler        Compiler
	Executable      Executable
}

// Compile Program
// Build Recursively iterates through a Scanner and translates the Books into Apps using the provided segmental system.
func (p *Program) Compile() error {
	compiler := Compiler{
		Name:            p.Name,
		Meta:            MetaDict{},
		Books:           p.Scanner.Program.Books,
		SegmentalSystem: p.SegmentalSystem,
	}
	err := compiler.Compile()
	if err != nil {
		return err
	}
	p.Compiler = compiler
	return nil
}

// Assemble Program
// Assemble recursively injects structural context into the Built Apps of the Program.
func (p *Program) Assemble() error {
	err := p.Compiler.Assemble()
	if err != nil {
		return err
	}
	return nil
}

// Link collects the ordered source bundles and their dependencies.
func (p *Program) Link() {
	var executable Executable
	executable.Name = p.Name
	for _, ca := range p.Compiler.CompiledApps {
		ca.Link()
		executable.ByteArrays = append(executable.ByteArrays, ca.Content)
		executable.MapArrays = append(executable.MapArrays, ca.ContentMap)
		executable.DependencyArrays = append(executable.DependencyArrays, ca.Dependencies)
		executable.EnvDepArrays = append(executable.EnvDepArrays, ca.EnvDependencies)
	}
	p.Executable = executable
}
