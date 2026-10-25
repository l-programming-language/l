package Lyger

import (
	"github.com/l-programming-language/l/lerrors"
)

type sourceModule struct {
	Name       string
	Visibility string `json:",omitempty"`
}

type sourceTarget struct {
	Kind    string
	Name    string
	Root    string
	Primary bool
}

type amlClassMetadata struct {
	Name           string
	Type           string
	Location       string
	Parent         string `json:",omitempty"`
	GenericObjects []string
	Dimensions     []string
	Root           string            `json:",omitempty"`
	Module         *sourceModule     `json:",omitempty"`
	Meta           map[string]string `json:",omitempty"`
}

type amlGraph struct {
	services       []amlServiceMetadata
	classes        []amlClassMetadata
	objects        []amlObjectMetadata
	serviceMembers map[string][]sourceMember
	classMembers   map[string][]sourceMember
	objectTypes    map[string]string
	rootObjects    []*object
}

func prepareAMLGraph(app *App) (*amlGraph, error) {
	graph := &amlGraph{services: []amlServiceMetadata{}, classes: []amlClassMetadata{}, objects: []amlObjectMetadata{}, serviceMembers: map[string][]sourceMember{}, classMembers: map[string][]sourceMember{}, objectTypes: map[string]string{}}
	owners := make(map[string]*object)
	classIDs := make(map[string]bool)
	active := make(map[*class]bool)
	addObject := func(obj *object) error {
		if obj == nil || obj.ID == "" {
			return lerrors.InvalidInput.Errorf("outline contains an unnamed or nil Object")
		}
		if err := validateAMLName(obj.ID); err != nil {
			return err
		}
		if _, exists := owners[obj.ID]; exists {
			return lerrors.InvalidInput.Errorf("Object %q has multiple owners or duplicate identity", obj.ID)
		}
		owners[obj.ID] = obj
		graph.objectTypes[obj.ID] = obj.Meta["Type"]
		graph.objects = append(graph.objects, amlObjectMetadata{Name: obj.ID, Type: obj.Meta["Type"], Module: obj.Module})
		return nil
	}
	var visitClass func(*class, string) error
	visitClass = func(group *class, parent string) error {
		if group == nil || group.ID == "" {
			return lerrors.InvalidInput.Errorf("outline contains an unnamed or nil Class")
		}
		if active[group] {
			return lerrors.Cycle.Errorf("Class %q recursively owns itself", group.ID)
		}
		if classIDs[group.ID] {
			return lerrors.InvalidInput.Errorf("Class %q has multiple owners or duplicate identity", group.ID)
		}
		if err := validateAMLName(group.ID); err != nil {
			return err
		}
		if declared := group.Meta["Parent"]; declared != "" && declared != parent {
			return lerrors.InvalidInput.Errorf("Class %q Parent %q disagrees with owner %q", group.ID, declared, parent)
		}
		classIDs[group.ID], active[group] = true, true
		members, err := orderedSourceMembers(group.Objects, group.Children, group.members)
		if err != nil {
			return err
		}
		graph.classMembers[group.ID] = members
		entry := amlClassMetadata{Name: group.ID, Type: group.Meta["Type"], Location: group.Meta["Location"], Parent: parent, GenericObjects: []string{}, Dimensions: []string{}, Root: group.Root, Module: group.Module}
		for key, value := range group.Meta {
			switch key {
			case "Type", "Location", "Parent":
				continue
			}
			if entry.Meta == nil {
				entry.Meta = make(map[string]string)
			}
			entry.Meta[key] = value
		}
		for _, member := range members {
			if member.Object != nil {
				entry.GenericObjects = append(entry.GenericObjects, member.Object.ID)
			} else {
				entry.Dimensions = append(entry.Dimensions, member.Class.ID)
			}
		}
		graph.classes = append(graph.classes, entry)
		for _, member := range members {
			if member.Object != nil {
				err = addObject(member.Object)
			} else {
				err = visitClass(member.Class, group.ID)
			}
			if err != nil {
				return err
			}
		}
		delete(active, group)
		return nil
	}
	for _, svc := range app.Services {
		if svc == nil || svc.ID == "" {
			return nil, lerrors.InvalidInput.Errorf("outline contains an unnamed or nil Service")
		}
		if _, exists := graph.serviceMembers[svc.ID]; exists {
			return nil, lerrors.InvalidInput.Errorf("duplicate Service %q", svc.ID)
		}
		if len(svc.Children) != 0 {
			return nil, lerrors.Unsupported.Errorf("Service %q contains unsupported child Services", svc.ID)
		}
		if err := validateAMLName(svc.ID); err != nil {
			return nil, err
		}
		members, err := orderedSourceMembers(svc.Objects, svc.Classes, svc.members)
		if err != nil {
			return nil, err
		}
		graph.serviceMembers[svc.ID] = members
		entry := amlServiceMetadata{Name: svc.ID, Type: svc.Meta["Type"], Package: svc.Meta["Package"], Location: svc.Meta["Location"], Parent: svc.Meta["Parent"], GenericObjects: []string{}, Dimensions: []string{}, Target: svc.Target}
		for _, member := range members {
			if member.Object != nil {
				entry.GenericObjects = append(entry.GenericObjects, member.Object.ID)
				err = addObject(member.Object)
			} else {
				entry.Dimensions = append(entry.Dimensions, member.Class.ID)
				err = visitClass(member.Class, "")
			}
			if err != nil {
				return nil, err
			}
		}
		graph.services = append(graph.services, entry)
	}
	catalog := make(map[string]bool)
	for _, obj := range app.Objects {
		if obj == nil || catalog[obj.ID] {
			return nil, lerrors.InvalidInput.Errorf("nil or duplicate application Object catalog entry")
		}
		catalog[obj.ID] = true
		if prior, exists := owners[obj.ID]; exists {
			if prior != obj {
				return nil, lerrors.InvalidInput.Errorf("application Object %q conflicts with its file owner", obj.ID)
			}
			continue
		}
		if err := addObject(obj); err != nil {
			return nil, err
		}
		graph.rootObjects = append(graph.rootObjects, obj)
	}
	return graph, nil
}
