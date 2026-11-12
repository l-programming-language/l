package Lyger

import (
	"fmt"

	"github.com/l-programming-language/l/contracts"
)

func (r *contractRecorder) recordLowering(rec *record, instance DefinitionInstance, args map[string]DefinitionArgument) {
	r.recordLoweringDependency(rec, instance, args, "selected-source-composition", make(map[*record]bool))
}

func (r *contractRecorder) recordLoweringDependency(rec *record, instance DefinitionInstance, args map[string]DefinitionArgument, kind string, active map[*record]bool) {
	if rec == nil {
		return
	}
	canonical, err := canonicalRecord(rec, r.compilation.Semantics.Aliases)
	if err != nil || canonical == nil {
		return
	}
	if active[canonical] {
		return
	}
	active[canonical] = true
	defer delete(active, canonical)
	meaning := r.meaning(canonical, nil)
	origin := r.origin(r.declarationTokens[canonical])
	step := contracts.LoweringStep{Meaning: meaning, Name: canonical.ID, Kind: kind, Origin: origin, Target: contracts.TargetRef{Unknown: true, Phase: contracts.Invocation}, Arguments: make(map[string]string)}
	for name, arg := range args {
		step.Arguments[name] = arg.Kind
	}
	for _, dependency := range instance.Dependencies {
		if dependency.definition != nil {
			step.Dependencies = append(step.Dependencies, r.meaning(dependency.definition, nil))
		}
	}
	statement := r.compilation.Semantics.Declarations[canonical]
	if statement != nil && statement.source != nil {
		step.Implementation = contracts.StableID("body", statement.source.Raw)
	} else {
		step.Implementation = contracts.StableID("body", instance.Original)
	}
	step.ID = contracts.StableID("lowering-use", struct {
		Meaning contracts.MeaningKey
		Ordinal int
	}{meaning, len(r.lowering)})
	r.lowering = append(r.lowering, step)
	for _, dependency := range instance.Dependencies {
		r.recordLoweringDependency(dependency.definition, dependency, nil, "selected-declaration-dependency", active)
	}
}

func (r *contractRecorder) inventory() contracts.LoweringInventory {
	result := contracts.LoweringInventory{Steps: append([]contracts.LoweringStep(nil), r.lowering...), Fingerprints: make(map[string]string), BoundaryCoverage: contracts.Partial, TransitiveCoverage: contracts.Unknown}
	seen := make(map[contracts.MeaningKey]bool)
	for _, step := range result.Steps {
		result.Fingerprints[string(step.Meaning)] = step.Implementation
		if seen[step.Meaning] {
			continue
		}
		seen[step.Meaning] = true
		result.Native = append(result.Native, contracts.NativeBoundary{ID: string(step.Meaning), Meaning: step.Meaning, Target: step.Target, Origin: step.Origin, Status: contracts.Unresolved, Reason: "selected composition has no approved native/representation conformance witness; opaque transitive behavior remains unknown"})
	}
	result.Unknowns = []contracts.UnknownReason{{Code: "native_transitive_closure", Message: "selected source compositions are inventoried; opaque authored leaves, conditional final-owner destruction and their transitive native calls require explicit contracts and conformance evidence", Components: []string{"effects", "ownership", "failure", "resources", "cleanup"}}}
	return result
}

// RecordedInventory snapshots the inputs and lowering choices from this
// compilation only, after its selected source definitions have been lowered.
func (c *CompilationResult) RecordedInventory() (ContractBindings, contracts.LoweringInventory, error) {
	if c == nil || c.Compilation == nil || c.Compilation.contractRecorder == nil {
		return ContractBindings{}, contracts.LoweringInventory{}, fmt.Errorf("compilation has no binding recorder")
	}
	r := c.Compilation.contractRecorder
	bindings, inventory := r.snapshot(), r.inventory()
	for _, source := range r.compilation.SourceOrigins {
		if bindings.Inputs[r.logicalSource(source.Name)] != source.SHA256 {
			return ContractBindings{}, contracts.LoweringInventory{}, fmt.Errorf("contract input %s changed between reads in one compilation", source.Name)
		}
	}
	return bindings, inventory, inventory.Validate()
}
