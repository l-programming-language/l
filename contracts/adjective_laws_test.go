package contracts

import (
	"fmt"
	"reflect"
)

type constraintLaw struct {
	Left        MeaningKey  `json:"left"`
	Right       MeaningKey  `json:"right"`
	Target      TargetRef   `json:"target"`
	Kind        string      `json:"kind"`
	Premises    []Predicate `json:"premises"`
	Fingerprint string      `json:"fingerprint"`
}

// proveConstraintLaw concerns pure conjunction only. It never licenses changing
// evaluation, construction, wrappers, releases, or any other ordered plan node.
func proveConstraintLaw(left, right MeaningKey, target TargetRef, kind string, predicates []Predicate) (constraintLaw, error) {
	if kind != "constraint-commutation" && kind != "constraint-idempotence" {
		return constraintLaw{}, fmt.Errorf("unsupported composition law")
	}
	for _, predicate := range predicates {
		if status, _ := Entails([]Predicate{predicate}, predicate); status != Established {
			return constraintLaw{}, fmt.Errorf("unsupported pure constraint predicate")
		}
	}
	law := constraintLaw{Left: left, Right: right, Target: target, Kind: kind, Premises: append([]Predicate(nil), predicates...)}
	law.Fingerprint = StableID("constraint-law", law)
	return law, nil
}

func (law constraintLaw) applies(left, right MeaningKey, target TargetRef, facts []Predicate) bool {
	copy := law
	copy.Fingerprint = ""
	if law.Fingerprint != StableID("constraint-law", copy) || law.Left != left || law.Right != right || TargetRelation(law.Target, target) != Established {
		return false
	}
	status, _ := Entails(facts, Predicate{Kind: "and", Args: law.Premises})
	return status == Established
}

// equivalentElaboration requires the same ordered typed plan after callers
// explicitly rebase target/resource variables. Equal facet labels are insufficient.
func equivalentElaboration(left, right Plan) bool {
	return reflect.DeepEqual(elaborationWithoutProvenance(left), elaborationWithoutProvenance(right))
}

func elaborationWithoutProvenance(plan Plan) Plan {
	var strip func(Node) Node
	strip = func(node Node) Node {
		node.ID, node.Origin = "", Origin{}
		if node.Body != nil {
			body := strip(*node.Body)
			node.Body = &body
		}
		for _, children := range []*[]Node{&node.Children, &node.Then, &node.Else, &node.Cleanup} {
			*children = append([]Node(nil), (*children)...)
			for i := range *children {
				(*children)[i] = strip((*children)[i])
			}
		}
		node.Effects = append([]Effect(nil), node.Effects...)
		for i := range node.Effects {
			node.Effects[i].Origin = Origin{}
		}
		return node
	}
	plan.Entry = strip(plan.Entry)
	definitions := make(map[MeaningKey]Node, len(plan.Definitions))
	for key, node := range plan.Definitions {
		definitions[key] = strip(node)
	}
	plan.Definitions = definitions
	callbacks := make(map[string]Node, len(plan.Callbacks))
	for key, node := range plan.Callbacks {
		callbacks[key] = strip(node)
	}
	plan.Callbacks = callbacks
	return plan
}
