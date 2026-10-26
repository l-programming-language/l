package Lyger

import (
	"fmt"

	"github.com/l-programming-language/l/lerrors"
)

// sourceMember preserves the ordering of file wrappers and nested Classes.
// It has no relationship to lexical scope, runtime activation or value ownership.
type sourceMember struct {
	Object *object
	Class  *class
}

// amlSourceMember is the matching ordered content projection. The typed slices
// on chapter/text remain catalog projections; rendering traverses these members.
type amlSourceMember struct {
	Sentence *sentence
	Text     *text
}

func (c *Compilation) lowerSourceMembers(nodes []*constructionNode, body []DefinitionInstance, packageName, parent string, active map[*constructionNode]bool) ([]sourceMember, []amlSourceMember, error) {
	var members []sourceMember
	var contents []amlSourceMember
	for _, node := range nodes {
		if node == nil || node.Name == "" {
			return nil, nil, lerrors.InvalidInput.Errorf("source hierarchy requires a named Class or Object")
		}
		if active[node] {
			return nil, nil, lerrors.Cycle.Errorf("recursive source construction %q", node.Name)
		}
		active[node] = true
		switch node.Nominal {
		case "object":
			obj := newRootObject(node.Name)
			file := &sentence{Name: node.Name, Type: "Object"}
			for _, function := range node.Children {
				if function == nil {
					return nil, nil, lerrors.InvalidInput.Errorf("object %q has a nil function", node.Name)
				}
				instance, err := c.lowerFunction(function, body)
				if err != nil {
					return nil, nil, fmt.Errorf("lower object %q: %w", node.Name, err)
				}
				words := instance.Words
				if len(words) == 0 {
					words = [][]string{{instance.Expression}}
				}
				for _, syllables := range words {
					word := &w{Type: "Word"}
					for _, expression := range syllables {
						word.Syllables = append(word.Syllables, &s{Type: "Syllable", CompiledBytes: []byte(expression)})
					}
					file.Words = append(file.Words, word)
				}
				obj.Attributes = append(obj.Attributes, &attribute{ID: function.Name, Definition: &def{ID: function.Name, Text: instance.Expression, Type: "Function", Compilation: []byte(instance.Expression)}})
			}
			members = append(members, sourceMember{Object: obj})
			contents = append(contents, amlSourceMember{Sentence: file})
		case "class":
			group := &class{ID: node.Name, Meta: map[string]string{"Type": rootComponentType, "Package": packageName, "Location": rootServiceLocation, "Parent": parent}}
			content := &text{Name: node.Name, Type: "Class"}
			children, childContents, err := c.lowerSourceMembers(node.Children, body, packageName, node.Name, active)
			if err != nil {
				return nil, nil, fmt.Errorf("lower class %q: %w", node.Name, err)
			}
			group.members, content.members = children, childContents
			for _, child := range children {
				if child.Object != nil {
					group.Objects = append(group.Objects, child.Object)
				} else {
					group.Children = append(group.Children, child.Class)
				}
			}
			for _, child := range childContents {
				if child.Sentence != nil {
					content.sentences = append(content.sentences, child.Sentence)
				} else {
					content.children = append(content.children, child.Text)
				}
			}
			members = append(members, sourceMember{Class: group})
			contents = append(contents, amlSourceMember{Text: content})
		default:
			return nil, nil, lerrors.TypeMismatch.Errorf("source hierarchy child must be a Class or Object, got %s", node.Nominal)
		}
		delete(active, node)
	}
	return members, contents, nil
}

// orderedSourceMembers keeps older flat internal callers valid while making an
// explicitly ordered graph authoritative and checking its typed projections.
func orderedSourceMembers(objects []*object, classes []*class, members []sourceMember) ([]sourceMember, error) {
	if members == nil {
		for _, obj := range objects {
			members = append(members, sourceMember{Object: obj})
		}
		for _, group := range classes {
			members = append(members, sourceMember{Class: group})
		}
	}
	objectSet, classSet := make(map[*object]bool), make(map[*class]bool)
	for _, obj := range objects {
		if obj == nil || objectSet[obj] {
			return nil, lerrors.InvalidInput.Errorf("nil or duplicate Object membership")
		}
		objectSet[obj] = true
	}
	for _, group := range classes {
		if group == nil || classSet[group] {
			return nil, lerrors.InvalidInput.Errorf("nil or duplicate Class membership")
		}
		classSet[group] = true
	}
	for _, member := range members {
		if (member.Object == nil) == (member.Class == nil) {
			return nil, lerrors.InvalidInput.Errorf("source member requires exactly one Object or Class")
		}
		if member.Object != nil {
			if !objectSet[member.Object] {
				return nil, lerrors.InvalidInput.Errorf("Object order disagrees with membership")
			}
			delete(objectSet, member.Object)
		} else {
			if !classSet[member.Class] {
				return nil, lerrors.InvalidInput.Errorf("Class order disagrees with membership")
			}
			delete(classSet, member.Class)
		}
	}
	if len(objectSet) != 0 || len(classSet) != 0 {
		return nil, lerrors.InvalidInput.Errorf("source order omits members")
	}
	return members, nil
}

func orderedAMLSourceMembers(sentences []*sentence, texts []*text, members []amlSourceMember) ([]amlSourceMember, error) {
	if members == nil {
		for _, file := range sentences {
			members = append(members, amlSourceMember{Sentence: file})
		}
		for _, group := range texts {
			members = append(members, amlSourceMember{Text: group})
		}
	}
	fileSet, textSet := make(map[*sentence]bool), make(map[*text]bool)
	for _, file := range sentences {
		if file == nil || fileSet[file] {
			return nil, lerrors.InvalidInput.Errorf("nil or duplicate Sentence membership")
		}
		fileSet[file] = true
	}
	for _, group := range texts {
		if group == nil || textSet[group] {
			return nil, lerrors.InvalidInput.Errorf("nil or duplicate Text membership")
		}
		textSet[group] = true
	}
	for _, member := range members {
		if (member.Sentence == nil) == (member.Text == nil) {
			return nil, lerrors.InvalidInput.Errorf("content member requires exactly one Sentence or Text")
		}
		if member.Sentence != nil {
			if !fileSet[member.Sentence] {
				return nil, lerrors.InvalidInput.Errorf("Sentence order disagrees with membership")
			}
			delete(fileSet, member.Sentence)
		} else {
			if !textSet[member.Text] {
				return nil, lerrors.InvalidInput.Errorf("Text order disagrees with membership")
			}
			delete(textSet, member.Text)
		}
	}
	if len(fileSet) != 0 || len(textSet) != 0 {
		return nil, lerrors.InvalidInput.Errorf("content order omits members")
	}
	return members, nil
}
