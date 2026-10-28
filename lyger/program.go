package Lyger

import (
	"fmt"

	"github.com/l-programming-language/l/lerrors"
)

const (
	defaultAppLanguage  = "rust"
	defaultAppVersion   = "0.0.0"
	defaultAppGenDir    = "/src"
	rootComponentType   = "root"
	rootServiceLocation = "/src/"
	noParent            = "None"
	defaultHost         = "http://127.0.0.1"
)

func appMetadata(id string) map[string]string {
	return map[string]string{"AppName": id, "AppLang": defaultAppLanguage, "AppVersion": defaultAppVersion, "AppGenDir": defaultAppGenDir}
}

func newRootService(name string) *service {
	return &service{ID: name, Meta: map[string]string{"Type": rootComponentType, "Package": name, "Location": rootServiceLocation, "Parent": noParent}}
}

func newRootObject(name string) *object {
	return &object{ID: name, Meta: map[string]string{"Type": rootComponentType}}
}

// loweredApplication owns only the generated fields awaiting publication.
// The Program already owns the corresponding staged book.
type loweredApplication struct {
	Original     *App
	Generated    *App
	Construction *applicationConstruction
	Body         []DefinitionInstance
}

func (c *Compilation) lowerProgram() error {
	program := &Program{Cache: c.Memory}
	var lowered []loweredApplication
	generated := make(map[*App]*App, len(c.Semantics.Constructions))
	for _, construction := range c.Semantics.Constructions {
		original := construction.App
		staged := *original
		app := &staged
		// Prototype defaults are logical source-generation settings, not paths
		// that the frontend writes or a claim that binaries were deployed.
		app.Meta = appMetadata(app.buildName())
		app.Settings = map[string]string{"Host": defaultHost}
		app.Services = nil
		body, err := c.lowerAppBody(original)
		if err != nil {
			return fmt.Errorf("lower %s body: %w", app.ID, err)
		}
		if len(body) == 0 {
			instance, err := c.emptyConstructionSequence(construction.Root)
			if err != nil {
				return fmt.Errorf("construct %s empty body: %w", app.ID, err)
			}
			if instance != nil {
				body = []DefinitionInstance{*instance}
			}
		}
		if err := c.validateConstructionParameters(construction.Root, body); err != nil {
			return fmt.Errorf("bind %s: %w", app.ID, err)
		}
		appBook := &book{Name: app.buildName(), Type: "App", Cache: c.Memory}
		for _, serviceNode := range construction.Root.Children {
			if serviceNode.Nominal != "service" || serviceNode.Name == "" {
				return lerrors.TypeMismatch.Errorf("application child must be a named service, got %s", serviceNode.Nominal)
			}
			serviceValue := newRootService(serviceNode.Name)
			serviceChapter := &chapter{Name: serviceNode.Name, Type: "Service"}
			members, contents, err := c.lowerSourceMembers(serviceNode.Children, body, serviceNode.Name, "", make(map[*constructionNode]bool))
			if err != nil {
				return err
			}
			serviceValue.members, serviceChapter.members = members, contents
			for _, member := range members {
				if member.Object != nil {
					serviceValue.Objects = append(serviceValue.Objects, member.Object)
				} else {
					serviceValue.Classes = append(serviceValue.Classes, member.Class)
					app.Meta["StructureVersion"] = "2"
				}
			}
			for _, member := range contents {
				if member.Sentence != nil {
					serviceChapter.sentences = append(serviceChapter.sentences, member.Sentence)
				} else {
					serviceChapter.Texts = append(serviceChapter.Texts, member.Text)
				}
			}
			app.Services = append(app.Services, serviceValue)
			appBook.Chapters = append(appBook.Chapters, serviceChapter)
		}
		if err := c.prepareSourceTopology(app, appBook); err != nil {
			return fmt.Errorf("prepare %s source topology: %w", app.ID, err)
		}
		program.Books = append(program.Books, appBook)
		generated[original] = app
		lowered = append(lowered, loweredApplication{Original: original, Generated: app, Construction: construction, Body: body})
	}
	for _, app := range c.Semantics.Apps {
		if staged, ok := generated[app]; ok {
			program.Apps = append(program.Apps, staged)
		} else {
			program.Apps = append(program.Apps, app)
		}
	}
	if err := program.Make(); err != nil {
		return err
	}
	// Rendering has succeeded for the complete graph. Publish only fields this
	// phase owns, retaining canonical App identity and all runtime/user state.
	for _, result := range lowered {
		result.Original.Meta = result.Generated.Meta
		result.Original.Settings = result.Generated.Settings
		result.Original.Services = result.Generated.Services
		result.Construction.Body = result.Body
	}
	program.Apps = c.Semantics.Apps
	c.Program = program
	c.ASMBytes = append([]byte(nil), program.AMLBytes...)
	return nil
}
