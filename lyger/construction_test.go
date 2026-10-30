package Lyger

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

func constructionTestCompilation(t *testing.T) *Compilation {
	t.Helper()
	config := canonicalTestOptions()
	source, err := os.ReadFile("test/main.l")
	if err != nil {
		t.Fatal(err)
	}
	compilation := NewLygerService(config).NewCompilation(source, "test/main.l")
	compilation.Make()
	if compilation.Error != nil {
		t.Fatal(compilation.Error)
	}
	return compilation
}

type constructionShape struct {
	Nominal, Name, Definition     string
	InputsDefined, OutputsDefined bool
	Parameters, Wrappers          []string
	Children                      []constructionShape
}

func constructionTestShape(t *testing.T, node *constructionNode) constructionShape {
	t.Helper()
	shape := constructionShape{Nominal: node.Nominal, Name: node.Name, InputsDefined: node.InputsDefined, OutputsDefined: node.OutputsDefined}
	if node.Definition != nil {
		shape.Definition = node.Definition.ID
	}
	for _, parameter := range node.Parameters {
		nominal, err := nominalType(parameter.noun)
		if err != nil {
			t.Fatal(err)
		}
		shape.Parameters = append(shape.Parameters, nominal)
	}
	for _, wrapper := range node.Wrappers {
		shape.Wrappers = append(shape.Wrappers, wrapper.Target.String()+":"+string(wrapper.Template.rawBytes))
	}
	for _, child := range node.Children {
		shape.Children = append(shape.Children, constructionTestShape(t, child))
	}
	return shape
}

func TestConstructionDescriptorAliasesPreserveExpansion(t *testing.T) {
	compilation := constructionTestCompilation(t)
	sites, err := compilation.Semantics.Constructions[0].Source.sGram.registrationSites()
	if err != nil || len(sites) != 1 {
		t.Fatalf("construction sites = %v, %v", sites, err)
	}
	original := sites[0].valueObject()
	baseline, err := compilation.constructObject(original, make(map[*descriptorRule]bool))
	if err != nil {
		t.Fatal(err)
	}
	var descriptorIndex int = -1
	for index, adjective := range original.adjectives {
		descriptor, err := compilation.descriptorFor(adjective)
		if err != nil {
			t.Fatal(err)
		}
		if descriptor != nil {
			descriptorIndex = index
			break
		}
	}
	if descriptorIndex < 0 {
		t.Fatal("application has no source-defined descriptor")
	}
	definition := original.adjectives[descriptorIndex].binding.Record
	prior := definition
	for _, name := range []string{"Plain", "Simple"} {
		// Alias registration and grammar are tested separately. Here each fresh
		// binding must dispatch the original descriptor without changing its AST.
		value := &entry{}
		value.bindFrom(prior.Value)
		value.Lexeme.Root = strings.ToLower(name)
		alias := &record{ID: name, Type: prior.Type, Value: value}
		compilation.Semantics.Aliases[alias] = prior
		compilation.Memory[name] = alias
		variant := *original
		variant.adjectives = append([]*token(nil), original.adjectives...)
		aliasToken := *original.adjectives[descriptorIndex]
		aliasToken.rawBytes = []byte(name)
		originalScope := aliasToken.scope
		aliasToken.scope = compilation.Scope.Root
		state, bindErr := compilation.resolver().bindToken(&aliasToken)
		aliasToken.scope = originalScope
		if bindErr != nil || state != lookupFound {
			t.Fatalf("descriptor alias %s did not bind", name)
		}
		variant.adjectives[descriptorIndex] = &aliasToken
		actual, err := compilation.constructObject(&variant, make(map[*descriptorRule]bool))
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(constructionTestShape(t, actual), constructionTestShape(t, baseline)) {
			t.Fatalf("descriptor alias %s changed application expansion", name)
		}
		if len(actual.Children) == 0 || actual.Children[0] == baseline.Children[0] {
			t.Fatal("alias application reused mutable construction nodes")
		}
		actual.Children[0].Name = "changed"
		if baseline.Children[0].Name == "changed" || definition.Value.Lexeme.Root != "basic" {
			t.Fatal("alias application mutated the prototype or another instance")
		}
		prior = alias
	}
}

func TestConstructionRejectsDescriptorForDifferentTarget(t *testing.T) {
	compilation := constructionTestCompilation(t)
	sites, err := compilation.Semantics.Constructions[0].Source.sGram.registrationSites()
	if err != nil || len(sites) != 1 {
		t.Fatalf("construction sites = %v, %v", sites, err)
	}
	application := sites[0].valueObject()
	var descriptorToken *token
	for _, adjective := range application.adjectives {
		descriptor, err := compilation.descriptorFor(adjective)
		if err != nil {
			t.Fatal(err)
		}
		if descriptor != nil {
			descriptorToken = adjective
			break
		}
	}
	if descriptorToken == nil {
		t.Fatal("application has no source-defined descriptor")
	}
	definition, err := compilation.descriptorFor(descriptorToken)
	if err != nil {
		t.Fatal(err)
	}
	serviceObject := definition.Rules[0].Object.phrase.orderedObjects()[0]
	variant := *serviceObject
	variant.adjectives = append(append([]*token(nil), serviceObject.adjectives...), descriptorToken)
	if _, err := compilation.constructObject(&variant, make(map[*descriptorRule]bool)); err == nil || !strings.Contains(err.Error(), "has no constructor for service") {
		t.Fatalf("app-only descriptor silently applied to a service: %v", err)
	}
}

func TestConstructionParametersRequireProvidedNominalBody(t *testing.T) {
	compilation := constructionTestCompilation(t)
	construction := compilation.Semantics.Constructions[0]
	if len(construction.Body) == 0 {
		t.Fatal("source application has no lowered body")
	}
	if err := compilation.validateConstructionParameters(construction.Root, construction.Body); err != nil {
		t.Fatalf("actual source body does not satisfy its constructor: %v", err)
	}
	if err := compilation.validateConstructionParameters(construction.Root, nil); err == nil || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("required body was silently omitted: %v", err)
	}
	wrong := append([]DefinitionInstance(nil), construction.Body...)
	wrong[0].Symbol = "UnrelatedDefinition"
	if err := compilation.validateConstructionParameters(construction.Root, wrong); err == nil || !strings.Contains(err.Error(), "requires") {
		t.Fatalf("incompatible body nominal was silently accepted: %v", err)
	}
	if construction.Body[0].Symbol == wrong[0].Symbol {
		t.Fatal("parameter validation mutated stored body identity")
	}
}

func TestConstructionFunctionSlotsRequireDeclaredParameters(t *testing.T) {
	compilation := constructionTestCompilation(t)
	construction := compilation.Semantics.Constructions[0]
	var findDefinition func(*constructionNode) *constructionNode
	findDefinition = func(node *constructionNode) *constructionNode {
		if node.Definition != nil && node.InputsDefined && node.OutputsDefined {
			return node
		}
		for _, child := range node.Children {
			if found := findDefinition(child); found != nil {
				return found
			}
		}
		return nil
	}
	function := findDefinition(construction.Root)
	if function == nil || len(function.Parameters) == 0 {
		t.Fatal("source application contains no parameterized function")
	}
	if _, err := compilation.lowerFunction(function, construction.Body); err != nil {
		t.Fatalf("actual function declaration did not bind: %v", err)
	}
	missing := *function
	missing.Parameters = nil
	if _, err := compilation.lowerFunction(&missing, construction.Body); err == nil || !strings.Contains(err.Error(), "no declared constructor parameter") {
		t.Fatalf("unbound function slot received an invented value: %v", err)
	}
}
