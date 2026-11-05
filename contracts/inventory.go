package contracts

import (
	"fmt"
	"maps"
)

// LoweringStep records a selected typed composition, never Rust source parsed
// back into purported source semantics. A selection is not execution evidence.
type LoweringStep struct {
	ID             string            `json:"id"`
	Meaning        MeaningKey        `json:"meaning"`
	Name           string            `json:"name"`
	Kind           string            `json:"kind"`
	Origin         Origin            `json:"origin"`
	Target         TargetRef         `json:"target"`
	Arguments      map[string]string `json:"arguments,omitempty"`
	Dependencies   []MeaningKey      `json:"dependencies,omitempty"`
	Implementation string            `json:"implementation"`
}

type LoweringInventory struct {
	Steps              []LoweringStep    `json:"steps"`
	Native             []NativeBoundary  `json:"native"`
	Fingerprints       map[string]string `json:"fingerprints"`
	BoundaryCoverage   Coverage          `json:"boundary_coverage"`
	TransitiveCoverage Coverage          `json:"transitive_coverage"`
	Unknowns           []UnknownReason   `json:"unknowns,omitempty"`
}

func (i LoweringInventory) Validate() error {
	seen := make(map[string]bool)
	for _, step := range i.Steps {
		if step.ID == "" || step.Meaning == "" || step.Implementation == "" {
			return fmt.Errorf("lowering step lacks identity, meaning or implementation fingerprint")
		}
		if seen[step.ID] {
			return fmt.Errorf("duplicate lowering occurrence %s", step.ID)
		}
		seen[step.ID] = true
	}
	return nil
}

// VerifyInventory seals the selected model. Emitters must not add dependencies
// to a prepared inventory and retain assurance computed for its former state.
func VerifyInventory(prepared, current LoweringInventory) error {
	if err := prepared.Validate(); err != nil {
		return err
	}
	if err := current.Validate(); err != nil {
		return err
	}
	before, err := Fingerprint(prepared)
	if err != nil {
		return err
	}
	after, err := Fingerprint(current)
	if err != nil {
		return err
	}
	if before != after {
		return fmt.Errorf("lowering inventory changed after contract checking")
	}
	return nil
}

// CloneInventory detaches all mutable collections before exposing a prepared
// inventory. Caller edits cannot alter the model that rendering will consume.
func CloneInventory(value LoweringInventory) LoweringInventory {
	value.Steps = append([]LoweringStep(nil), value.Steps...)
	for i := range value.Steps {
		step := &value.Steps[i]
		step.Target.Path = append([]string(nil), step.Target.Path...)
		step.Origin.Expansion = append([]int(nil), step.Origin.Expansion...)
		step.Dependencies = append([]MeaningKey(nil), step.Dependencies...)
		step.Arguments = maps.Clone(step.Arguments)
	}
	value.Native = append([]NativeBoundary(nil), value.Native...)
	for i := range value.Native {
		value.Native[i].Target.Path = append([]string(nil), value.Native[i].Target.Path...)
		value.Native[i].Origin.Expansion = append([]int(nil), value.Native[i].Origin.Expansion...)
	}
	value.Unknowns = append([]UnknownReason(nil), value.Unknowns...)
	for i := range value.Unknowns {
		value.Unknowns[i].ExcludedEffects = append([]EffectKind(nil), value.Unknowns[i].ExcludedEffects...)
		value.Unknowns[i].Components = append([]string(nil), value.Unknowns[i].Components...)
		value.Unknowns[i].Origin.Expansion = append([]int(nil), value.Unknowns[i].Origin.Expansion...)
	}
	value.Fingerprints = maps.Clone(value.Fingerprints)
	return value
}
