package Lyger

import (
	"fmt"
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// prepareSourceTopology connects the already resolved source graph. Classes are
// modules, Objects remain files, and neither creates a runtime activation.
// Compiler-owned roots contain only source structure or the entry delegation;
// the original Object bodies are never flattened, moved, or reparsed.
func (c *Compilation) prepareSourceTopology(app *App, book *book) error {
	structural := len(app.Services) != 1
	for _, service := range app.Services {
		structural = structural || len(service.Classes) != 0 || len(service.Objects) != 1
	}
	if !structural {
		return nil // Preserve the established ordinary one-file bundle interface.
	}
	if len(app.Services) != len(book.Chapters) {
		return lerrors.Internal.Errorf("Rust source topology has inconsistent Services and Chapters")
	}
	used := map[string]bool{}
	var mainObject *object
	var mainService *service
	var mainPath []string
	var visit func([]sourceMember, []string, *service) error
	visit = func(members []sourceMember, prefix []string, service *service) error {
		for _, member := range members {
			if member.Object != nil {
				object := member.Object
				used[object.ID] = true
				object.Module = &sourceModule{Name: sourceModuleName(object.ID), Visibility: "crate"}
				for _, attribute := range object.Attributes {
					if attribute.ID == "main" {
						if mainObject != nil {
							return lerrors.InvalidInput.Errorf("Rust application %s has multiple main functions", app.buildName())
						}
						mainObject, mainService = object, service
						mainPath = append(append([]string(nil), prefix...), object.Module.Name, "main")
					}
				}
			} else if member.Class != nil {
				group := member.Class
				used[group.ID] = true
				group.Module = &sourceModule{Name: sourceModuleName(group.ID), Visibility: "crate"}
				if err := visit(group.members, append(append([]string(nil), prefix...), group.Module.Name), service); err != nil {
					return err
				}
			}
		}
		return nil
	}
	for _, service := range app.Services {
		if err := visit(service.members, nil, service); err != nil {
			return err
		}
	}
	if mainObject == nil {
		return lerrors.InvalidInput.Errorf("Rust application %s requires one main function", app.buildName())
	}
	allocate := func(base string) string {
		name := base
		for suffix := 1; used[name]; suffix++ {
			name = fmt.Sprintf("%s_%d", base, suffix)
		}
		used[name] = true
		return name
	}
	var rootClasses func([]sourceMember, []amlSourceMember) error
	rootClasses = func(members []sourceMember, contents []amlSourceMember) error {
		if len(members) != len(contents) {
			return lerrors.Internal.Errorf("Rust Class topology lost source order")
		}
		for index, member := range members {
			if member.Class == nil {
				continue
			}
			group, text := member.Class, contents[index].Text
			if text == nil || text.Name != group.ID {
				return lerrors.Internal.Errorf("Rust Class topology and content disagree")
			}
			if err := rootClasses(group.members, text.members); err != nil {
				return err
			}
			name := allocate("__l_class_" + sourceModuleName(group.ID))
			root, content, err := c.sourceTopologyRoot(name, "")
			if err != nil {
				return err
			}
			group.Root = name
			group.Objects = append(group.Objects, root)
			group.members = append(group.members, sourceMember{Object: root})
			text.sentences = append(text.sentences, content)
			text.members = append(text.members, amlSourceMember{Sentence: content})
		}
		return nil
	}
	for index, service := range app.Services {
		chapter := book.Chapters[index]
		if chapter.Name != service.ID {
			return lerrors.Internal.Errorf("Rust Service topology and content disagree")
		}
		if err := rootClasses(service.members, chapter.members); err != nil {
			return err
		}
		var root *object
		kind, name := "lib", sourceModuleName(service.ID)
		if service == mainService {
			kind, name = "bin", app.buildName()
			for _, direct := range service.Objects {
				if direct == mainObject {
					root = direct
				}
			}
		}
		if root == nil {
			id := allocate("__l_service_" + sourceModuleName(service.ID))
			delegation := ""
			if service == mainService {
				delegation = strings.Join(mainPath, "::")
			}
			var content *sentence
			var err error
			root, content, err = c.sourceTopologyRoot(id, delegation)
			if err != nil {
				return err
			}
			service.Objects = append(service.Objects, root)
			service.members = append(service.members, sourceMember{Object: root})
			chapter.sentences = append(chapter.sentences, content)
			chapter.members = append(chapter.members, amlSourceMember{Sentence: content})
		}
		service.Target = &sourceTarget{Kind: kind, Name: name, Root: root.ID, Primary: service == mainService}
	}
	app.Meta["StructureVersion"] = "2"
	return nil
}

func (c *Compilation) sourceTopologyRoot(name, delegation string) (*object, *sentence, error) {
	root := newRootObject(name)
	content := &sentence{Name: name, Type: "Object", Words: []*w{{Type: "Word", Syllables: []*s{{Type: "Syllable", CompiledBytes: []byte("A")}}}}}
	if delegation != "" {
		call, err := c.instantiateBuiltin("Delegate", map[string]DefinitionArgument{"Target": definitionReferenceArgument(delegation)})
		if err != nil {
			return nil, nil, err
		}
		body, err := c.instantiateBuiltin("Process", map[string]DefinitionArgument{"Definitions": {Kind: "Syllables", Syllables: []string{call.Expression}}})
		if err != nil {
			return nil, nil, err
		}
		entry, err := c.instantiateBuiltin("Function", map[string]DefinitionArgument{
			"Name": {Kind: "Text", Text: "main"}, "Inputs": {Kind: "Props"}, "Outputs": {Kind: "Texts"},
			"Process": {Kind: "Definition", Definition: &body},
		})
		if err != nil {
			return nil, nil, err
		}
		expression := entry.Expression
		content.Words[0].Syllables = append(content.Words[0].Syllables, &s{Type: "Syllable", CompiledBytes: []byte(expression)})
		root.Attributes = []*attribute{{ID: "main", Definition: &def{ID: "main", Type: "Function", Text: expression, Compilation: []byte(expression)}}}
	}
	return root, content, nil
}

// Rust module names are a compiler namespace. Preserve ordinary source names;
// reserve the encoding prefix too, so a native-looking source name cannot
// collide with the encoding of a keyword or another non-identifier spelling.
func sourceModuleName(name string) string {
	valid := name != "" && name != "_" && !strings.HasPrefix(name, "__l_name_")
	for index, char := range name {
		valid = valid && (char == '_' || char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || index > 0 && char >= '0' && char <= '9')
	}
	for _, keyword := range strings.Fields("as break const continue crate else enum extern false fn for if impl in let loop match mod move mut pub ref return self Self static struct super trait true type unsafe use where while async await dyn abstract become box do final gen macro override priv typeof unsized virtual yield try") {
		if name == keyword {
			valid = false
		}
	}
	if valid {
		return name
	}
	return fmt.Sprintf("__l_name_%x", []byte(name))
}
