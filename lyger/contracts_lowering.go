package Lyger

import (
	"fmt"
	"sort"

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

// ContractAnalysis retains both the real compiler graph and the typed analysis
// input. Backend support obligations remain pending until coordinator closure.
type ContractAnalysis struct {
	Compilation          *CompilationResult
	Bindings             ContractBindings
	Plan                 contracts.Plan
	Summary              contracts.DerivedSummary
	Lowering             contracts.LoweringInventory
	Orchestrator         BuildUnit
	Lifecycle            []contracts.AppProtocolState
	LifecycleObligations []contracts.Obligation
	SourceProjections    []contracts.SourceAdjectiveProjection
	SourceCertificates   []contracts.SourceAdjectiveCertificate
	SourceInstantiations []contracts.SourceAdjectiveInstantiation
	NativeUses           []contracts.NativeUse
	Scopes               []contracts.AnalysisScope
}

type ContractAnalysisOptions struct {
	Inference          contracts.AnalysisOptions
	Session            *contracts.Session
	SourceTransformers []contracts.SourceTransformerDeclaration
}

// RefreshLoweringInventory includes helper choices made by subsequent real
// compilation operations, such as the final coordinator entry and runtime
// library. Call this before merging backend coverage and reinferring Plan.
// Pending backend closure remains an independent obligation.
func (a *ContractAnalysis) RefreshLoweringInventory() error {
	if a == nil || a.Compilation == nil || a.Compilation.Compilation == nil || a.Compilation.Compilation.contractRecorder == nil {
		return fmt.Errorf("contract analysis has no binding recorder")
	}
	r := a.Compilation.Compilation.contractRecorder
	inventory := r.inventory()
	if err := inventory.Validate(); err != nil {
		return err
	}
	oldNative := make(map[string]bool)
	for _, boundary := range a.Lowering.Native {
		oldNative[contracts.StableID("source-boundary", boundary)] = true
	}
	var native []contracts.NativeBoundary
	for _, boundary := range a.Plan.NativeClosure {
		if !oldNative[contracts.StableID("source-boundary", boundary)] {
			native = append(native, boundary)
		}
	}
	a.Plan.NativeClosure = append(native, inventory.Native...)
	oldResiduals := make(map[string]bool)
	for _, unknown := range a.Lowering.Unknowns {
		oldResiduals[contracts.StableID("source-residual", unknown)] = true
	}
	var residuals []contracts.UnknownReason
	for _, unknown := range a.Plan.Residuals {
		if !oldResiduals[contracts.StableID("source-residual", unknown)] {
			residuals = append(residuals, unknown)
		}
	}
	a.Plan.Residuals = append(residuals, inventory.Unknowns...)
	a.Lowering = inventory
	a.Bindings = r.snapshot()
	return a.refreshNativeScopes()
}

func (ls *LygerService) AnalyzeContracts(source []byte, path string, options ContractAnalysisOptions) (*ContractAnalysis, error) {
	c, _, err := ls.ResolveContractBindings(source, path)
	if err != nil {
		return nil, err
	}
	result := &CompilationResult{Compilation: c, ExpandedSource: append([]byte(nil), c.ExpandedBytes...), AML: append([]byte(nil), c.ASMBytes...), BackendAML: append([]byte(nil), c.Program.BackendBytes...)}
	plan, err := planOrchestration(c.Semantics)
	if err != nil {
		return nil, err
	}
	r := c.contractRecorder
	typed := contracts.Plan{}
	// Use deterministic logical companion references; inspection never allocates
	// a generation, materializes a workspace, invokes Cargo or runs the program.
	var companions []AppArtifact
	for index, app := range c.Semantics.Apps {
		companions = append(companions, AppArtifact{ID: app.buildName(), RelativeExecutable: fmt.Sprintf("apps/app-%03d", index)})
	}
	entry, err := result.OrchestratorUnit("contracts_entry", companions)
	if err != nil {
		return nil, err
	}
	inventory := r.inventory()
	if err := inventory.Validate(); err != nil {
		return nil, err
	}
	// The typed frontend cannot certify opaque helper implementations. These
	// residual obligations close the summary even when explicit statements look
	// clean, and include helper selection performed during composition.
	typed.NativeClosure = append(typed.NativeClosure, inventory.Native...)
	typed.Residuals = append(typed.Residuals, inventory.Unknowns...)
	typed.NativeClosure = append(typed.NativeClosure, contracts.NativeBoundary{ID: "backend-support-pending", Status: contracts.Unresolved, Reason: "backend support and finalization require target closure", Origin: contracts.Origin{Label: "coordinator closure pending"}, Implicit: true})
	lifecycle, lifecycleObligations := r.contractLifecycle(plan)
	analysis := &ContractAnalysis{Compilation: result, Bindings: r.snapshot(), Plan: typed, Lowering: inventory, Orchestrator: entry, Lifecycle: lifecycle, LifecycleObligations: lifecycleObligations}
	if err := analysis.refreshNativeScopes(); err != nil {
		return nil, err
	}
	if err := inferContractAnalysis(analysis, options); err != nil {
		return nil, err
	}
	if err := analysis.verifySourceContracts(options); err != nil {
		return nil, err
	}
	return analysis, nil
}

func (r *contractRecorder) contractBlock(block *orchestrationBlock) []contracts.Node {
	if block == nil {
		return nil
	}
	var body []contracts.Node
	for _, slot := range block.Storage {
		if slot.Initial != nil || slot.Parameter {
			body = append(body, r.contractAcquire(slot, contracts.Origin{Label: "typed activation entry"}))
		}
	}
	body = append(body, r.contractInstructions(block.Body)...)
	// This is a storage activation, never an implicit App termination boundary.
	// Exceptional cleanup remains unknown unless the actual runtime models it.
	return []contracts.Node{{Kind: contracts.Scope, ID: fmt.Sprintf("activation:%d", block.ID), Children: body, CleanupOutcomes: []contracts.Outcome{contracts.Normal}, Origin: contracts.Origin{Label: "typed activation storage release"}}}
}

func (r *contractRecorder) contractAppTarget(app *App) contracts.TargetRef {
	if app == nil {
		return contracts.TargetRef{Unknown: true, Type: "App", Phase: contracts.Invocation}
	}
	for _, construction := range r.compilation.Semantics.Constructions {
		if construction.App == app {
			target, ok := r.targets[construction.Root]
			if ok {
				target.Phase = contracts.Invocation
				return target
			}
		}
	}
	return contracts.TargetRef{Unknown: true, Type: "App", Phase: contracts.Invocation}
}

func (r *contractRecorder) contractStorageTarget(slot *orchestrationStorage) contracts.TargetRef {
	if slot == nil {
		return contracts.TargetRef{Unknown: true, Phase: contracts.Invocation}
	}
	if slot.Capture != nil {
		return contracts.TargetRef{Unknown: true, Type: slot.Type.String(), Phase: contracts.Invocation, Generation: "capture environment requires instance substitution"}
	}
	root := string(r.definition(slot.Record))
	if root == "" {
		return contracts.TargetRef{Unknown: true, Type: slot.Type.String(), Phase: contracts.Invocation}
	}
	owner := -1
	scope := "generated activation"
	if slot.Owner != nil {
		owner = slot.Owner.ID
		scope = r.scopeID(slot.Owner.Scope)
	}
	generation := contracts.StableID("activation-storage", struct {
		Owner, Slot int
		Scope       string
	}{owner, slot.ID, scope})
	return contracts.TargetRef{Root: root, Type: slot.Type.String(), Phase: contracts.Invocation, Generation: generation}
}

func contractStorageResource(target contracts.TargetRef) contracts.ResourceRef {
	if target.Unknown {
		return contracts.ResourceRef{Unknown: true, Kind: "storage"}
	}
	return contracts.ResourceRef{ID: contracts.StableID("storage-instance", struct{ Root, Generation string }{target.Root, target.Generation}), Kind: "storage"}
}

func (r *contractRecorder) contractAcquire(slot *orchestrationStorage, origin contracts.Origin) contracts.Node {
	target := r.contractStorageTarget(slot)
	resource := contractStorageResource(target)
	if resource.Unknown {
		return contracts.Node{Kind: contracts.Native, Native: "unresolved-storage-construction", Origin: origin, Target: target, Phase: contracts.Construction}
	}
	destructor := contracts.Node{Kind: contracts.Event, Origin: origin, Target: target, Phase: contracts.Destruction}
	// The selected scalar storage representation has no destructor.
	if slot.Type != valueNumber && slot.Type != valueTruth && slot.Type != valueNull {
		destructor = r.contractRepresentationDestruction(slot, origin)
	}
	return contracts.Node{Kind: contracts.Acquire, Resource: resource, Target: target, Ownership: "owned", SuspensionSafe: true, Phase: contracts.Construction, Origin: origin, Cleanup: []contracts.Node{destructor}}
}

func (r *contractRecorder) contractFunctions(body []*orchestrationInstruction, definitions map[contracts.MeaningKey]contracts.Node) {
	for _, current := range body {
		if current == nil {
			continue
		}
		if current.Function != nil && current.Function.Body != nil {
			function := current.Function
			origin := r.origin(r.declarationTokens[function.Source.Record])
			// Callee lowering (prologue, value transfer and return composition)
			// is an independent native obligation, not an empty effect inferred
			// from the explicit source body or charged to the invocation wrapper.
			children := []contracts.Node{r.contractFunctionBodyOverhead(function, origin)}
			children = append(children, r.contractBlock(function.Body)...)
			definitions[r.meaning(function.Source.Record, nil)] = contracts.Node{Kind: contracts.Sequence, Children: children, Origin: origin}
			r.contractFunctions(function.Body.Body, definitions)
		}
		if current.Block != nil {
			r.contractFunctions(current.Block.Body, definitions)
		}
		if current.Continuation != nil {
			r.contractFunctions(current.Continuation.Body, definitions)
		}
		r.contractFunctions(current.Then, definitions)
		r.contractFunctions(current.Else, definitions)
		r.contractFunctions(current.Actions, definitions)
		r.contractFunctions(current.AppCases, definitions)
	}
}

func (r *contractRecorder) contractInstructions(instructions []*orchestrationInstruction) []contracts.Node {
	var result []contracts.Node
	for _, current := range instructions {
		if current == nil || current.Source == nil {
			continue
		}
		source := current.Source
		firstUse := 0
		if r.nativeAnalysis != nil {
			firstUse = len(r.nativeAnalysis.uses)
		}
		var scopeMeaning contracts.MeaningKey
		origin := r.origin(source.SourceToken)
		if origin.Source == "" {
			origin = contracts.Origin{Source: r.logicalSource(source.Source.Path), Start: source.Source.Start.Offset, End: source.Source.End.Offset}
		}
		node := contracts.Node{Kind: contracts.Sequence, ID: contracts.StableID("instruction", struct {
			Origin contracts.Origin
			Order  int
			Kind   string
		}{origin, source.Order, source.Kind.String()}), Origin: origin, Phase: contracts.Invocation}
		for _, argument := range current.Arguments {
			node.Children = append(node.Children, r.contractValue(argument, origin)...)
		}
		node.Children = append(node.Children, r.contractValue(current.Value, origin)...)
		target := contracts.TargetRef{Unknown: true, Phase: contracts.Invocation}
		resource := contracts.ResourceRef{Unknown: true}
		if current.App != nil {
			target = r.contractAppTarget(current.App)
			resource = contracts.ResourceRef{ID: target.Root, Kind: "process", Unknown: target.Unknown}
		}
		if current.Target != nil && current.Target.Storage != nil && current.App == nil {
			target = r.contractStorageTarget(current.Target.Storage)
			resource = contractStorageResource(target)
		}
		event := func(kind contracts.EffectKind) contracts.Node {
			return contracts.Node{Kind: contracts.Event, Origin: origin, Target: target, Phase: contracts.Invocation, Effects: []contracts.Effect{{Kind: kind, Resource: resource, Target: target, Phase: contracts.Invocation, Origin: origin}}}
		}
		switch source.Kind {
		case instructionIf:
			if current.Condition != nil {
				node.Children = append(node.Children, r.contractValue(current.Condition.Left, origin)...)
				node.Children = append(node.Children, r.contractValue(current.Condition.Right, origin)...)
			}
			node.Children = append(node.Children, contracts.Node{Kind: contracts.Branch, Origin: origin, Then: r.contractInstructions(current.Then), Else: r.contractInstructions(current.Else)})
		case instructionBlock:
			node.Children = append(node.Children, r.contractBlock(current.Block)...)
		case instructionCall:
			node.Children = append(node.Children, r.contractInstructions(current.Actions)...)
			if source.Application != nil && source.Application.verb != nil && source.Application.verb.binding != nil {
				selected := source.Application.verb.binding.Record
				scopeMeaning = r.meaning(selected, nil)
				node.Children = append(node.Children, r.contractNativeUse(selected, selected.ID, r.recordTarget(selected, contracts.Invocation), origin, nil, nil, "value"))
			}
		case instructionAssign:
			if current.Target != nil && current.Target.Storage != nil && !resource.Unknown {
				node.Children = append(node.Children, contracts.Node{Kind: contracts.Release, Resource: resource, Target: target, Origin: origin, Phase: contracts.Destruction}, r.contractAcquire(current.Target.Storage, origin))
			}
			node.Children = append(node.Children, event("write"))
		case instructionUpdate:
			if emitted, ok := r.contractEmittedUpdate(current, target, origin); ok {
				// The trace includes actual argument evaluation, boxing, snapshot,
				// checked invocation and final storage publication, in that order.
				node.Children = emitted
			} else {
				node.Children = append(node.Children, r.contractValue(current.Target, origin)...)
				node.Children = append(node.Children, r.contractValue(current.Key, origin)...)
				node.Children = append(node.Children, r.contractUpdate(current, target, origin)...)
			}
			// Publication follows the checked update. A returned failure must not
			// acquire a write which the selected emitter never performs.
			node.Children = append(node.Children, event("write"))
		case instructionBuild:
			node.Children = append(node.Children, event("io"))
		case instructionRun:
			node.Children = append(node.Children, event("spawn"))
		case instructionStop:
			node.Children = append(node.Children, event("signal"))
		case instructionAwait:
			node.Children = append(node.Children, event("wait"), event("suspend"))
			node.Children = append(node.Children, r.contractBlock(current.Continuation)...)
		case instructionShow:
			node.Children = append(node.Children, event("io"))
		case instructionExit:
			node.Children = append(node.Children, contracts.Node{Kind: contracts.Native, Native: "runtime-terminal-cleanup", Phase: contracts.Cleanup, Origin: origin}, contracts.Node{Kind: contracts.Event, Outcome: contracts.Abort, Origin: origin})
		case instructionBindAction, instructionAliasAction, instructionFunction:
			for _, value := range current.CaptureValues {
				node.Children = append(node.Children, r.contractValue(value, origin)...)
			}
			if current.Function != nil {
				for _, value := range current.Function.Captures {
					node.Children = append(node.Children, r.contractValue(value, origin)...)
				}
			}
			handles := current.CaptureHandles
			if current.Function != nil {
				handles = append(append([]*orchestrationCallableReference(nil), handles...), current.Function.CaptureHandles...)
			}
			for range handles {
				node.Children = append(node.Children, contracts.Node{Kind: contracts.Native, Native: "capture-handle-retain", Origin: origin, Phase: contracts.Construction})
			}
			node.Children = append(node.Children, contracts.Node{Kind: contracts.Native, Native: "callable-capture-construction", Phase: contracts.Construction, Origin: origin})
			if current.Target != nil && current.Target.Storage != nil && !resource.Unknown {
				node.Children = append(node.Children, contracts.Node{Kind: contracts.Release, Resource: resource, Target: target, Origin: origin, Phase: contracts.Destruction}, r.contractAcquire(current.Target.Storage, origin))
			}
			if current.Callable != nil {
				callable := current.Callable
				owner := -1
				if callable.Owner != nil {
					owner = callable.Owner.ID
				}
				capture := contracts.ResourceRef{ID: contracts.StableID("capture-instance", struct {
					Definition  contracts.DefinitionKey
					Owner, Slot int
				}{r.definition(callable.Record), owner, callable.ID}), Kind: "capture-environment"}
				node.Children = append(node.Children, contracts.Node{Kind: contracts.Release, Resource: capture, Origin: origin, Phase: contracts.Destruction}, contracts.Node{Kind: contracts.Acquire, Resource: capture, Ownership: "owned", SuspensionSafe: true, Origin: origin, Phase: contracts.Construction, Cleanup: []contracts.Node{{Kind: contracts.Native, Native: "capture-environment-destruction", Origin: origin, Phase: contracts.Destruction}}})
			}
		case instructionInvoke:
			node.Children = append(node.Children, r.contractValue(current.Target, origin)...)
			var function *sourceFunction
			if current.Target != nil && current.Target.Record != nil {
				selected, err := canonicalRecord(current.Target.Record, r.compilation.Semantics.Aliases)
				if err == nil {
					function = r.compilation.Semantics.Functions[selected]
				}
			}
			if function != nil && !function.Native {
				scopeMeaning = r.meaning(function.Record, nil)
				node.Children = append(node.Children, r.contractFunctionInvocation(current, function, origin, node.ID))
			} else if function != nil {
				scopeMeaning = r.meaning(function.Record, nil)
				node.Children = append(node.Children, r.contractNativeUse(function.Record, function.Record.ID, r.recordTarget(function.Record, contracts.Invocation), origin, nil, nil, "value"))
			} else {
				node.Children = append(node.Children, contracts.Node{Kind: contracts.Native, Native: "typed-function-invocation", Origin: origin, Phase: contracts.Invocation})
			}
			for _, value := range current.Results {
				if value.Storage != nil {
					destination := r.contractStorageTarget(value.Storage)
					owner := contractStorageResource(destination)
					if !owner.Unknown {
						node.Children = append(node.Children, contracts.Node{Kind: contracts.Release, Resource: owner, Target: destination, Phase: contracts.Destruction, Origin: origin}, r.contractAcquire(value.Storage, origin))
					}
				}
			}
		case instructionReturn:
			for _, value := range current.Results {
				node.Children = append(node.Children, r.contractValue(value, origin)...)
			}
			node.Children = append(node.Children, contracts.Node{Kind: contracts.Return, Origin: origin, Phase: contracts.Invocation})
		default:
			node.Children = append(node.Children, contracts.Node{Kind: contracts.Native, Native: "unmodeled-instruction:" + source.Kind.String(), Origin: origin})
		}
		if len(current.AppCases) > 0 {
			// Dynamic App dispatch preserves alternatives, not sequential effects.
			var alternatives []contracts.Node
			for _, branch := range current.AppCases {
				alternatives = append(alternatives, contracts.Node{Kind: contracts.Sequence, Children: r.contractInstructions([]*orchestrationInstruction{branch})})
			}
			if len(alternatives) > 0 {
				joined := alternatives[len(alternatives)-1]
				for i := len(alternatives) - 2; i >= 0; i-- {
					joined = contracts.Node{Kind: contracts.Branch, Then: []contracts.Node{alternatives[i]}, Else: []contracts.Node{joined}}
				}
				node.Children = append(r.contractValue(current.Target, origin), joined)
			}
		}
		result = append(result, node)
		switch source.Kind {
		case instructionUpdate, instructionInvoke, instructionCall:
			r.recordContractScope(node, target, origin, firstUse, scopeMeaning)
		}
	}
	return result
}

func (r *contractRecorder) contractValue(value *orchestrationValue, origin contracts.Origin) []contracts.Node {
	if value == nil {
		return nil
	}
	var nodes []contracts.Node
	for _, element := range value.Elements {
		nodes = append(nodes, r.contractValue(element, origin)...)
	}
	nodes = append(nodes, r.contractValue(value.Owner, origin)...)
	nodes = append(nodes, r.contractValue(value.Payload, origin)...)
	nodes = append(nodes, r.contractValue(value.Key, origin)...)
	read := value.Source != nil && value.Source.Kind == operandRead
	if (value.Type == valueNumber || value.Type == valueTruth || value.Type == valueNull) && !read {
		return nodes
	}
	name := "representation:" + value.Type.String()
	if read {
		name = "checked-representation-read:" + value.Type.String()
	}
	target := r.contractStorageTarget(value.Storage)
	if value.App != nil {
		target = r.contractAppTarget(value.App)
	}
	access := "copy"
	switch value.Access {
	case orchestrationAccessBorrow:
		access = "borrow"
	case orchestrationAccessOwned:
		access = "owned"
	case orchestrationAccessSnapshot:
		access = "logical_snapshot"
	}
	if value.Type == valuePointer && value.Access == orchestrationAccessSnapshot {
		access = "shared_pointer_target"
	}
	nodes = append(nodes, contracts.Node{Kind: contracts.Event, Origin: origin, Target: target, Phase: contracts.Invocation, Facts: []contracts.Predicate{{Kind: "ownership", Subject: contracts.StableID("value-access", struct {
		Target contracts.TargetRef
		Origin contracts.Origin
	}{target, origin}), Value: contracts.Atom{Kind: "ownership", Text: access}, Target: target}}})
	if value.Source != nil && value.Source.Member.Valid() {
		name = "runtime-member:" + value.Source.Member.String()
	}
	nodes = append(nodes, contracts.Node{Kind: contracts.Native, Native: name, Origin: origin, Target: target, Phase: contracts.Invocation})
	return nodes
}

// NativeNames reports canonical selected compositions for inventory/debugging.
// It does not infer native behavior from names.
func (a *ContractAnalysis) NativeNames() []string {
	set := make(map[string]bool)
	for _, step := range a.Lowering.Steps {
		set[step.Name] = true
	}
	var names []string
	for name := range set {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
