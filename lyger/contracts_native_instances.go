package Lyger

import (
	"encoding/json"

	"github.com/l-programming-language/l/contracts"
)

// Instantiation copies analysis records, never compiler source or emitted code.
// Callee-owned storage and native occurrences belong to one actual invocation;
// canonical definition meanings remain unchanged across those instances.
func (a *ContractAnalysis) instantiateNativeCalls() error {
	uses := make(map[string]contracts.NativeUse)
	for _, use := range a.NativeUses {
		uses[use.ID] = use
	}
	scopes := append([]contracts.AnalysisScope(nil), a.Scopes...)
	var instantiate func(*contracts.Node, map[contracts.MeaningKey]bool) error
	instantiate = func(node *contracts.Node, active map[contracts.MeaningKey]bool) error {
		for _, children := range [][]contracts.Node{node.Children, node.Then, node.Else, node.Cleanup} {
			for i := range children {
				if err := instantiate(&children[i], active); err != nil {
					return err
				}
			}
		}
		definition, found := a.Plan.Definitions[node.Meaning]
		if node.Kind != contracts.Call || !found || active[node.Meaning] {
			return nil
		}
		body, err := cloneContractRecord(definition)
		if err != nil {
			return err
		}
		instance := contractCallInstance{id: contracts.StableID("source-call-instance", node.ID), resources: make(map[string]contracts.ResourceRef), targets: make(map[string]contracts.TargetRef), nodes: make(map[string]string), uses: make(map[string]contracts.NativeUse)}
		instance.collect(body, uses)
		if err := instance.rewrite(&body, uses); err != nil {
			return err
		}
		for _, id := range instance.order {
			a.NativeUses = append(a.NativeUses, instance.uses[id])
		}
		for _, scope := range scopes {
			if id, ok := instance.nodes[scope.NodeID]; ok {
				scope.ID = contracts.StableID("instantiated-scope", []string{instance.id, scope.ID})
				scope.NodeID = id
				scope.Occurrence = contracts.BindingOccurrenceID(contracts.StableID("instantiated-occurrence", []string{instance.id, string(scope.Occurrence)}))
				scope.Target = instance.target(scope.Target)
				a.Scopes = append(a.Scopes, scope)
			}
		}
		node.Body = &body
		active[node.Meaning] = true
		err = instantiate(node.Body, active)
		delete(active, node.Meaning)
		return err
	}
	if err := instantiate(&a.Plan.Entry, make(map[contracts.MeaningKey]bool)); err != nil {
		return err
	}
	// Definition templates are logical inputs, not actual invocation scopes.
	// Public scope identities enumerate only concrete reachable typed nodes.
	reachable := make(map[string]bool)
	nativeUses := make(map[string]bool)
	var visit func(contracts.Node)
	visit = func(node contracts.Node) {
		reachable[node.ID] = true
		if node.Native != "" {
			nativeUses[node.Native] = true
		}
		for _, children := range [][]contracts.Node{node.Children, node.Then, node.Else, node.Cleanup} {
			for _, child := range children {
				visit(child)
			}
		}
		if node.Body != nil {
			visit(*node.Body)
		}
	}
	visit(a.Plan.Entry)
	var actual []contracts.AnalysisScope
	for _, scope := range a.Scopes {
		if reachable[scope.NodeID] {
			actual = append(actual, scope)
		}
	}
	a.Scopes = actual
	var actualUses []contracts.NativeUse
	for _, use := range a.NativeUses {
		if nativeUses[use.ID] {
			actualUses = append(actualUses, use)
		}
	}
	a.NativeUses = actualUses
	return nil
}

type contractCallInstance struct {
	id        string
	resources map[string]contracts.ResourceRef
	targets   map[string]contracts.TargetRef
	nodes     map[string]string
	uses      map[string]contracts.NativeUse
	order     []string
}

func (c *contractCallInstance) collect(node contracts.Node, uses map[string]contracts.NativeUse) {
	if node.Kind == contracts.Acquire && !node.Resource.Unknown && node.Resource.ID != "" {
		resource := node.Resource
		resource.ID = contracts.StableID("invocation-resource", []string{c.id, resource.ID})
		c.resources[node.Resource.ID] = resource
		if !node.Target.Unknown && node.Target.Root != "" {
			target := node.Target
			target.Generation = contracts.StableID("invocation-target", []string{c.id, target.Generation})
			c.targets[contracts.StableID("target", node.Target)] = target
		}
	}
	if use, ok := uses[node.Native]; ok {
		for _, resource := range use.Resources {
			if !resource.Unknown && resource.ID != "" && resource.Kind == "container-snapshot" {
				original := resource
				resource.ID = contracts.StableID("invocation-resource", []string{c.id, resource.ID})
				c.resources[original.ID] = resource
				if !use.Target.Unknown && use.Target.Root != "" {
					target := use.Target
					target.Generation = contracts.StableID("invocation-target", []string{c.id, target.Generation})
					c.targets[contracts.StableID("target", use.Target)] = target
				}
			}
		}
	}
	for _, children := range [][]contracts.Node{node.Children, node.Then, node.Else, node.Cleanup} {
		for _, child := range children {
			c.collect(child, uses)
		}
	}
}

func (c *contractCallInstance) target(target contracts.TargetRef) contracts.TargetRef {
	if replacement, ok := c.targets[contracts.StableID("target", target)]; ok {
		return replacement
	}
	return target
}

func (c *contractCallInstance) resource(resource contracts.ResourceRef) contracts.ResourceRef {
	if replacement, ok := c.resources[resource.ID]; ok {
		return replacement
	}
	return resource
}

func (c *contractCallInstance) substitutions(sub *contracts.Substitutions) {
	for name, resource := range sub.Resources {
		sub.Resources[name] = c.resource(resource)
	}
	for name, target := range sub.Targets {
		sub.Targets[name] = c.target(target)
	}
}

func (c *contractCallInstance) predicate(predicate *contracts.Predicate) {
	predicate.Target, predicate.Resource = c.target(predicate.Target), c.resource(predicate.Resource)
	for i := range predicate.Args {
		c.predicate(&predicate.Args[i])
	}
}

func (c *contractCallInstance) rewrite(node *contracts.Node, uses map[string]contracts.NativeUse) error {
	if node.ID != "" {
		old := node.ID
		node.ID = contracts.StableID("instantiated-node", []string{c.id, old})
		c.nodes[old] = node.ID
	}
	node.Target, node.Resource = c.target(node.Target), c.resource(node.Resource)
	c.substitutions(&node.Substitutions)
	for i := range node.Effects {
		node.Effects[i].Target = c.target(node.Effects[i].Target)
		node.Effects[i].Resource = c.resource(node.Effects[i].Resource)
	}
	for _, predicates := range [][]contracts.Predicate{node.Facts, node.Requires} {
		for i := range predicates {
			c.predicate(&predicates[i])
		}
	}
	if template, ok := uses[node.Native]; ok {
		use, exists := c.uses[node.Native]
		if !exists {
			var err error
			use, err = cloneContractRecord(template)
			if err != nil {
				return err
			}
			use.ID = contracts.StableID("instantiated-native-use", []string{c.id, template.ID})
			use.Occurrence = contracts.BindingOccurrenceID(contracts.StableID("instantiated-occurrence", use.ID))
			use.Target = c.target(use.Target)
			for name, resource := range use.Resources {
				use.Resources[name] = c.resource(resource)
			}
			c.substitutions(&use.Substitutions)
			use.Dependencies["source-invocation"] = c.id
			c.uses[node.Native] = use
			c.order = append(c.order, node.Native)
		}
		node.Native = use.ID
		node.Target = use.Target
		node.Substitutions = use.Substitutions
	}
	for _, children := range [][]contracts.Node{node.Children, node.Then, node.Else, node.Cleanup} {
		for i := range children {
			if err := c.rewrite(&children[i], uses); err != nil {
				return err
			}
		}
	}
	return nil
}

func cloneContractRecord[T any](value T) (T, error) {
	var clone T
	encoded, err := json.Marshal(value)
	if err != nil {
		return clone, err
	}
	err = json.Unmarshal(encoded, &clone)
	return clone, err
}
