package Lyger

import (
	"fmt"
	"math"
	"sort"
	"strconv"

	"github.com/l-programming-language/l/contracts"
)

// runtimeUpdateSelection is shared by source analysis and emission. It selects
// the existing provider helper; it neither implements that helper nor changes
// the copy-before-update behavior of List and Map.
type runtimeUpdateSelection struct {
	Word     string
	Receiver string
	Snapshot bool
}

func selectedRuntimeUpdate(current *orchestrationInstruction) runtimeUpdateSelection {
	if current == nil || current.Target == nil {
		return runtimeUpdateSelection{}
	}
	switch current.Target.Type {
	case valuePointer:
		return runtimeUpdateSelection{Word: "Storage.Pointer.Update", Receiver: "Pointer"}
	case valueList:
		word := "Storage.List.Append"
		if current.Key != nil {
			word = "Storage.List.Update"
		}
		return runtimeUpdateSelection{Word: word, Receiver: "List", Snapshot: true}
	case valueMap:
		return runtimeUpdateSelection{Word: "Storage.Map.Update", Receiver: "Map", Snapshot: true}
	default:
		return runtimeUpdateSelection{}
	}
}

type contractNativeAnalysis struct {
	inputs map[string]string
	uses   []contracts.NativeUse
	scopes []contracts.AnalysisScope
}

type contractNativeEmission struct {
	record      *record
	name        string
	descriptors []*record
	arguments   map[string]string
	result      string
	checks      string
}

func (e *orchestrationEmitter) recordContractNativeEmission(name string, descriptors []string, arguments map[string]DefinitionArgument, result string) {
	if e.contractValueDepth != 0 || e.contractInstruction == nil || e.compilation.contractRecorder == nil {
		return
	}
	rec, err := e.compilation.builtinRecord(name)
	if err != nil {
		return // The actual instantiation reports a missing selected definition.
	}
	emission := contractNativeEmission{record: rec, name: name, arguments: make(map[string]string), result: result}
	for _, name := range descriptors {
		if descriptor, err := e.compilation.builtinRecord(name); err == nil {
			emission.descriptors = append(emission.descriptors, descriptor)
		}
	}
	for name, argument := range arguments {
		emission.arguments[name] = argument.Kind
	}
	r := e.compilation.contractRecorder
	r.nativeEmissions[e.contractInstruction] = append(r.nativeEmissions[e.contractInstruction], emission)
}

func (e *orchestrationEmitter) recordContractCheckedResult(result string) {
	if e.compilation.contractRecorder == nil || e.contractInstruction == nil {
		return
	}
	emissions := e.compilation.contractRecorder.nativeEmissions[e.contractInstruction]
	if len(emissions) != 0 {
		emissions[len(emissions)-1].checks = result
	}
}

func (r *contractRecorder) contractRuntimeType(typ runtimeTypeRef) contracts.NativeType {
	result := contracts.NativeType{Kind: typ.Kind.String()}
	if typ.Definition != nil {
		result.Nominal = r.definition(typ.Definition)
	}
	if typ.Element != nil {
		result.Arguments = append(result.Arguments, r.contractRuntimeType(*typ.Element))
	}
	if typ.Kind == valueFunction {
		inputs := contracts.NativeType{Kind: "inputs"}
		outputs := contracts.NativeType{Kind: "outputs"}
		for _, input := range typ.Inputs {
			inputs.Arguments = append(inputs.Arguments, r.contractRuntimeType(input))
		}
		for _, output := range typ.Outputs {
			outputs.Arguments = append(outputs.Arguments, r.contractRuntimeType(output))
		}
		result.Arguments = append(result.Arguments, inputs, outputs)
	}
	return result
}

func (r *contractRecorder) contractNativeSignature(rec *record) *contracts.NativeSignature {
	if rec == nil {
		return nil
	}
	if rec.function == nil {
		template, err := r.compilation.recordTemplate(rec)
		if err != nil || template == nil {
			return nil
		}
		result := &contracts.NativeSignature{Inputs: []string{}, Outputs: []string{"native_fragment"}, OutputTypes: []contracts.NativeType{{Kind: "native_fragment"}}}
		for _, slot := range template.Slots {
			identity := slot.Name + ":" + slot.Type
			result.Inputs = append(result.Inputs, identity)
			result.InputTypes = append(result.InputTypes, contracts.NativeType{Kind: "definition_slot", Atom: identity})
		}
		return result
	}
	function := rec.function
	result := &contracts.NativeSignature{Inputs: []string{}, Outputs: []string{}}
	if function.Native {
		if function.Constructor == nil || !function.Constructor.InputsDefined || !function.Constructor.OutputsDefined {
			return nil
		}
		for _, input := range function.Constructor.Inputs {
			result.Inputs = append(result.Inputs, input.Type)
			result.InputTypes = append(result.InputTypes, contracts.NativeType{Kind: "native", Atom: input.Type})
		}
		for _, output := range function.Constructor.Outputs {
			result.Outputs = append(result.Outputs, output)
			result.OutputTypes = append(result.OutputTypes, contracts.NativeType{Kind: "native", Atom: output})
		}
		return result
	}
	for _, input := range function.Inputs {
		result.Inputs = append(result.Inputs, input.Type.Kind.String())
		result.InputTypes = append(result.InputTypes, r.contractRuntimeType(input.Type))
	}
	for _, output := range function.Outputs {
		result.Outputs = append(result.Outputs, output.Kind.String())
		result.OutputTypes = append(result.OutputTypes, r.contractRuntimeType(output))
	}
	return result
}

func (r *contractRecorder) contractImplementation(rec *record) string {
	var statements [][]byte
	if declaration := r.compilation.Semantics.Declarations[rec]; declaration != nil && declaration.source != nil {
		statements = append(statements, declaration.source.Raw)
	}
	if rec.function != nil {
		for _, instruction := range rec.function.Body {
			if instruction.Statement != nil && instruction.Statement.source != nil {
				statements = append(statements, instruction.Statement.source.Raw)
			}
		}
	}
	if len(statements) == 0 {
		return ""
	}
	return contracts.StableID("selected-native-body", statements)
}

// Native uses are keyed by canonical selection and actual occurrence. A display
// name is never a lookup key for approvals. Inputs conservatively include the
// current definition-site source closure until a narrower dependency is proven.
func (r *contractRecorder) contractNativeUse(rec *record, name string, target contracts.TargetRef, origin contracts.Origin, resources map[string]contracts.ResourceRef, values map[string]contracts.Atom, disposition string) contracts.Node {
	canonical, err := canonicalRecord(rec, r.compilation.Semantics.Aliases)
	if err != nil || canonical == nil || r.nativeAnalysis == nil {
		return contracts.Node{Kind: contracts.Native, Native: "unresolved-native-selection:" + name, Target: target, Origin: origin, Phase: contracts.Invocation}
	}
	meaning := r.meaning(canonical, nil)
	if target.Unknown || target.Root == "" {
		target = r.recordTarget(canonical, contracts.Invocation)
	}
	use := contracts.NativeUse{
		Meaning: meaning, Name: name, Target: target, Origin: origin, Phase: contracts.Invocation,
		Boundary:  "native_body",
		Signature: r.contractNativeSignature(canonical), Implementation: r.contractImplementation(canonical),
		Resources: resources, Values: values, OutcomeDisposition: disposition, Dependencies: make(map[string]string),
		Substitutions: contracts.Substitutions{Resources: resources, Values: values, Targets: map[string]contracts.TargetRef{"target": target}},
	}
	return r.appendContractNativeUse(use)
}

func (r *contractRecorder) appendContractNativeUse(use contracts.NativeUse) contracts.Node {
	use.ID = contracts.StableID("native-application", struct {
		Meaning contracts.MeaningKey
		Target  contracts.TargetRef
		Origin  contracts.Origin
		Ordinal int
	}{use.Meaning, use.Target, use.Origin, len(r.nativeAnalysis.uses)})
	use.Occurrence = contracts.BindingOccurrenceID(contracts.StableID("native-application-occurrence", use.ID))
	for input, fingerprint := range r.nativeAnalysis.inputs {
		use.Dependencies["source:"+input] = fingerprint
	}
	for _, step := range r.lowering {
		use.Dependencies["selected:"+string(step.Meaning)] = step.Implementation
	}
	r.nativeAnalysis.uses = append(r.nativeAnalysis.uses, use)
	node := contracts.Node{Kind: contracts.Native, ID: use.ID, Native: use.ID, Meaning: use.Meaning, Target: use.Target, Origin: use.Origin, Phase: use.Phase, Substitutions: use.Substitutions}
	scope := contracts.AnalysisScope{ID: contracts.StableID("native-body-scope", use.ID), NodeID: node.ID, Occurrence: use.Occurrence, Meaning: use.Meaning, Target: use.Target, Origin: use.Origin, Boundary: "native_body", Phase: use.Phase, NativeUses: []string{use.ID}, Complete: true, Includes: []string{"selected native body and its outcome-conditioned callee cleanup"}, Exclusions: []string{"caller argument evaluation, capture construction and result handling", "enclosing activation and workspace startup or finalization"}}
	for _, resource := range orderedContractResources(use.Resources) {
		if resource.ID != "" && !resource.Unknown {
			scope.Resources = append(scope.Resources, contracts.FailureLocation{Resource: resource})
		}
	}
	r.nativeAnalysis.scopes = append(r.nativeAnalysis.scopes, scope)
	return node
}

// A generated destruction boundary is keyed by the selected typed
// representation, not the spelling of a source word. It is never trusted merely
// because the compiler emitted it: each occurrence still needs component
// evidence for this implementation closure and exact storage resource.
func (r *contractRecorder) contractRepresentationDestruction(slot *orchestrationStorage, origin contracts.Origin) contracts.Node {
	target := r.contractStorageTarget(slot)
	if target.Unknown || r.nativeAnalysis == nil {
		return contracts.Node{Kind: contracts.Native, Native: "unresolved-representation-destruction", Target: target, Phase: contracts.Destruction, Origin: origin}
	}
	typ := r.contractRuntimeType(slot.RuntimeType)
	meaning := contracts.MeaningKey(contracts.StableID("representation-destruction", typ))
	resources := map[string]contracts.ResourceRef{"Value": contractStorageResource(target)}
	use := contracts.NativeUse{Meaning: meaning, Name: "representation-destruction:" + slot.Type.String(), Target: target, Origin: origin, Phase: contracts.Destruction, Boundary: "native_body",
		Signature: &contracts.NativeSignature{Inputs: []string{slot.Type.String()}, Outputs: []string{}, InputTypes: []contracts.NativeType{typ}},
		Implementation: contracts.StableID("selected-representation-implementation", struct {
			Type   contracts.NativeType
			Inputs map[string]string
		}{typ, r.nativeAnalysis.inputs}),
		Dependencies: make(map[string]string), Resources: resources, OutcomeDisposition: "value", Substitutions: contracts.Substitutions{Resources: resources, Targets: map[string]contracts.TargetRef{"target": target}}}
	return r.appendContractNativeUse(use)
}

func (r *contractRecorder) contractSelectedNative(name string, target contracts.TargetRef, origin contracts.Origin, resources map[string]contracts.ResourceRef, values map[string]contracts.Atom, disposition string) contracts.Node {
	rec, err := r.compilation.builtinRecord(name)
	if err != nil {
		return contracts.Node{Kind: contracts.Native, Native: "unresolved-native-selection:" + name, Target: target, Origin: origin, Phase: contracts.Invocation}
	}
	return r.contractNativeUse(rec, name, target, origin, resources, values, disposition)
}

func contractLiteral(value *orchestrationValue) (contracts.Atom, bool) {
	if value == nil || value.Literal == nil {
		return contracts.Atom{}, false
	}
	literal := value.Literal
	switch literal.Type {
	case valueNumber:
		if literal.Number >= math.MinInt64 && literal.Number < math.MaxInt64 && math.Trunc(literal.Number) == literal.Number {
			return contracts.Atom{Kind: "number", Number: int64(literal.Number)}, true
		}
		return contracts.Atom{Kind: "binary64", Text: strconv.FormatFloat(literal.Number, 'g', -1, 64)}, true
	case valueTruth:
		return contracts.Atom{Kind: "bool", Bool: literal.Truth}, true
	case valueText:
		return contracts.Atom{Kind: "text", Text: literal.Text}, true
	case valueNull:
		return contracts.Atom{Kind: "null"}, true
	default:
		return contracts.Atom{}, false
	}
}

func (r *contractRecorder) contractUpdate(current *orchestrationInstruction, target contracts.TargetRef, origin contracts.Origin) []contracts.Node {
	selection := selectedRuntimeUpdate(current)
	if selection.Word == "" {
		return []contracts.Node{{Kind: contracts.Native, Native: "unresolved-runtime-update", Target: target, Origin: origin, Phase: contracts.Invocation}}
	}
	receiver := contractStorageResource(target)
	resources := map[string]contracts.ResourceRef{"destination": receiver}
	var nodes []contracts.Node
	if selection.Snapshot {
		// The emitter copies the container before invoking the mutable helper.
		// This does not assert disjointness of nested Pointer payload resources.
		target.Generation = contracts.StableID("update-snapshot", struct {
			Target contracts.TargetRef
			Origin contracts.Origin
		}{target, origin})
		receiver = contracts.ResourceRef{ID: contracts.StableID("logical-snapshot", target), Kind: "container-snapshot"}
		nodes = append(nodes, contracts.Node{Kind: contracts.Native, Native: "runtime-update-snapshot-construction", Target: target, Origin: origin, Phase: contracts.Construction})
	}
	resources[selection.Receiver] = receiver
	values := make(map[string]contracts.Atom)
	if value, ok := contractLiteral(current.Value); ok {
		values["Value"] = value
	}
	if key, ok := contractLiteral(current.Key); ok {
		name := "Index"
		if current.Target.Type == valueMap {
			name = "Key"
		}
		values[name] = key
	}
	nodes = append(nodes, r.contractSelectedNative(selection.Word, target, origin, resources, values, "propagate_returned_failure"))
	return nodes
}

// contractEmittedUpdate translates the shared emitter's typed selection trace.
// DefinitionArgument kinds and canonical records are retained; generated native
// values are never parsed or used as proof of a helper's behavior.
func (r *contractRecorder) contractEmittedUpdate(current *orchestrationInstruction, target contracts.TargetRef, origin contracts.Origin) ([]contracts.Node, bool) {
	emissions, found := r.nativeEmissions[current.Source]
	if !found || len(emissions) == 0 {
		return nil, false
	}
	return r.contractEmittedSequence(emissions, current, target, origin), true
}

func (r *contractRecorder) contractEmittedSequence(emissions []contractNativeEmission, current *orchestrationInstruction, target contracts.TargetRef, origin contracts.Origin) []contracts.Node {
	selection := selectedRuntimeUpdate(current)
	original := contractStorageResource(target)
	receiver := original
	mutationTarget := target
	if selection.Snapshot {
		mutationTarget.Generation = contracts.StableID("update-snapshot", struct {
			Target contracts.TargetRef
			Origin contracts.Origin
		}{target, origin})
		receiver = contracts.ResourceRef{ID: contracts.StableID("logical-snapshot", mutationTarget), Kind: "container-snapshot"}
	}
	values := make(map[string]contracts.Atom)
	if value, ok := contractLiteral(current.Value); ok {
		values["Value"] = value
	}
	if key, ok := contractLiteral(current.Key); ok {
		name := "Index"
		if current.Target.Type == valueMap {
			name = "Key"
		}
		values[name] = key
	}
	var nodes []contracts.Node
	resultIndices := make(map[string]int)
	for _, emission := range emissions {
		actualTarget := target
		resources := map[string]contracts.ResourceRef{"destination": original}
		disposition := "value"
		if emission.name == selection.Word {
			actualTarget = mutationTarget
			resources[selection.Receiver] = receiver
		}
		if emission.name == "Storage.List.Copy" || emission.name == "Storage.Map.Copy" {
			resources["Value"], resources["result"] = original, receiver
		}
		node := r.contractNativeUse(emission.record, emission.name, actualTarget, origin, resources, values, disposition)
		if r.nativeAnalysis != nil && len(r.nativeAnalysis.uses) != 0 {
			use := &r.nativeAnalysis.uses[len(r.nativeAnalysis.uses)-1]
			if emission.name == "Storage.Invoke" {
				use.Boundary = "invocation_composition_overhead"
				r.setLastNativeScopeBoundary(use.Boundary, []string{"selected invocation composition, owned context/environment copies and wrapper finalization"}, []string{"separately modeled callee body effects, costs and owned storage destruction", "caller argument preparation, result publication and activation-scoped temporaries"})
			}
			if use.Signature == nil {
				// Authored source compositions expose named binding slots rather
				// than a native Function declaration. Record that interface as
				// such, without guessing a Rust signature from rendered text.
				use.Signature = &contracts.NativeSignature{Inputs: []string{}, Outputs: []string{"native_fragment"}, OutputTypes: []contracts.NativeType{{Kind: "native_fragment"}}}
				var names []string
				for name := range emission.arguments {
					names = append(names, name)
				}
				sort.Strings(names)
				for _, name := range names {
					use.Signature.Inputs = append(use.Signature.Inputs, name)
					use.Signature.InputTypes = append(use.Signature.InputTypes, contracts.NativeType{Kind: "composition_argument", Atom: name})
				}
			}
			use.Dependencies["argument-interface"] = contracts.StableID("typed-native-arguments", emission.arguments)
			var modifiers []contracts.MeaningKey
			for _, descriptor := range emission.descriptors {
				modifiers = append(modifiers, r.meaning(descriptor, nil))
			}
			use.Dependencies["ordered-modifiers"] = contracts.StableID("modifiers", modifiers)
		}
		if emission.checks != "" {
			if index, ok := resultIndices[emission.checks]; ok && index == len(nodes)-1 {
				nodes[index] = contracts.Node{Kind: contracts.CheckedResult, ID: contracts.StableID("checked-result", node.ID), Children: []contracts.Node{nodes[index]}, Cleanup: []contracts.Node{node}, Origin: origin, Target: actualTarget, Phase: contracts.Invocation}
				continue
			}
			// A future emitter which separates a result from its checked use
			// requires a typed dataflow extension, never guessed propagation.
			nodes = append(nodes, contracts.Node{Kind: contracts.Native, Native: "unresolved-checked-result-dataflow", Origin: origin, Target: actualTarget, Phase: contracts.Invocation})
		}
		nodes = append(nodes, node)
		if emission.result != "" {
			resultIndices[emission.result] = len(nodes) - 1
		}
	}
	return nodes
}

func (r *contractRecorder) contractFunctionBodyOverhead(function *orchestrationFunction, origin contracts.Origin) contracts.Node {
	target := r.contractStorageTarget(function.Self)
	resources := map[string]contracts.ResourceRef{"self": contractStorageResource(target)}
	node := r.contractNativeUse(function.Source.Record, function.Source.Record.ID+" body lowering", target, origin, resources, nil, "value")
	if r.nativeAnalysis != nil && len(r.nativeAnalysis.uses) != 0 {
		r.nativeAnalysis.uses[len(r.nativeAnalysis.uses)-1].Boundary = "source_body_lowering_overhead"
		r.setLastNativeScopeBoundary("source_body_lowering_overhead", []string{"implicit source-body prologue, value transfers, return composition and remaining lowering contributions"}, []string{"separately modeled source operations and owned storage destruction", "caller invocation composition and caller activation-scoped temporaries"})
	}
	return node
}

func (r *contractRecorder) setLastNativeScopeBoundary(boundary string, includes, excludes []string) {
	if len(r.nativeAnalysis.scopes) == 0 {
		return
	}
	scope := &r.nativeAnalysis.scopes[len(r.nativeAnalysis.scopes)-1]
	scope.Boundary, scope.Includes, scope.Exclusions = boundary, includes, excludes
}

// The selected invocation composition owns its environment copy and call
// finalization. Its approved contract is deliberately independent of the
// source-body summary: the latter cannot certify opaque wrapper behavior.
func (r *contractRecorder) contractFunctionInvocation(current *orchestrationInstruction, function *sourceFunction, origin contracts.Origin, instructionID string) contracts.Node {
	body := contracts.Node{Kind: contracts.Call, ID: contracts.StableID("source-call", instructionID), Meaning: r.meaning(function.Record, nil), Origin: origin, Phase: contracts.Invocation}
	target := r.recordTarget(function.Record, contracts.Invocation)
	if current.Target != nil && current.Target.Storage != nil {
		target = r.contractStorageTarget(current.Target.Storage)
	}
	invocation := contracts.Node{Kind: contracts.Sequence, ID: contracts.StableID("source-invocation", instructionID), Target: target, Origin: origin, Phase: contracts.Invocation}
	selected := -1
	for i := range r.nativeEmissions[current.Source] {
		emission := &r.nativeEmissions[current.Source][i]
		if emission.name == "Storage.Invoke" {
			selected = i
			break
		}
	}
	if selected < 0 {
		invocation.Children = append(invocation.Children, contracts.Node{Kind: contracts.Native, Native: "unresolved-source-invocation-composition", Target: target, Origin: origin, Phase: contracts.Invocation})
	} else {
		// This prefix is produced by actual ordered argument evaluation,
		// transfer, callable capture/environment copying and invocation.
		// Subsequent caller result publication is a separate boundary.
		invocation.Children = r.contractEmittedSequence(r.nativeEmissions[current.Source][:selected+1], current, target, origin)
	}
	invocation.Children = append(invocation.Children, body)
	if r.nativeAnalysis != nil {
		scope := contracts.AnalysisScope{ID: contracts.StableID("source-invocation-scope", invocation.ID), NodeID: invocation.ID,
			Occurrence: contracts.BindingOccurrenceID(contracts.StableID("invocation-occurrence", invocation.ID)), Meaning: body.Meaning, Target: target, Origin: origin, Phase: contracts.Invocation,
			Boundary: "source_invocation", Complete: true,
			Includes:   []string{"ordered argument evaluation, transfer and callable capture/environment copying", "selected callable invocation composition and its owned environment copy/finalization", "substituted source body, parameters, locals and retained-self final-owner cleanup"},
			Exclusions: []string{"caller result publication and preexisting caller-owned values", "caller activation-scoped emitter temporaries", "workspace startup and finalization"}}
		r.nativeAnalysis.scopes = append(r.nativeAnalysis.scopes, scope)
	}
	return invocation
}

func (a *ContractAnalysis) refreshNativeScopes() error {
	c := a.Compilation.Compilation
	plan, err := planOrchestration(c.Semantics)
	if err != nil {
		return err
	}
	r := c.contractRecorder
	r.nativeAnalysis = &contractNativeAnalysis{inputs: r.snapshot().Inputs}
	a.Plan.Entry = contracts.Node{Kind: contracts.Sequence, ID: "entry", Children: r.contractBlock(plan.Root)}
	a.Plan.Definitions = make(map[contracts.MeaningKey]contracts.Node)
	r.contractFunctions(plan.Root.Body, a.Plan.Definitions)
	a.NativeUses, a.Scopes = r.nativeAnalysis.uses, r.nativeAnalysis.scopes
	if err := a.instantiateNativeCalls(); err != nil {
		return err
	}
	a.refreshScopeDependencies()
	return nil
}

func (a *ContractAnalysis) refreshScopeDependencies() {
	uses := make(map[string]contracts.NativeUse)
	for _, use := range a.NativeUses {
		uses[use.ID] = use
	}
	nodes := make(map[string]contracts.Node)
	var inventory func(contracts.Node)
	inventory = func(node contracts.Node) {
		if node.ID != "" {
			nodes[node.ID] = node
		}
		if node.Body != nil {
			inventory(*node.Body)
		}
		for _, children := range [][]contracts.Node{node.Children, node.Then, node.Else, node.Cleanup} {
			for _, child := range children {
				inventory(child)
			}
		}
	}
	inventory(a.Plan.Entry)
	for _, definition := range a.Plan.Definitions {
		inventory(definition)
	}
	for i := range a.Scopes {
		scope := &a.Scopes[i]
		scope.NativeUses, scope.Resources = nil, nil
		seenUses, seenCalls := make(map[string]bool), make(map[contracts.MeaningKey]bool)
		var visit func(contracts.Node)
		visit = func(node contracts.Node) {
			if use, ok := uses[node.Native]; ok && !seenUses[use.ID] {
				seenUses[use.ID] = true
				scope.NativeUses = append(scope.NativeUses, use.ID)
				for _, resource := range orderedContractResources(use.Resources) {
					if !resource.Unknown && resource.ID != "" {
						scope.Resources = append(scope.Resources, contracts.FailureLocation{Resource: resource})
					}
				}
			}
			if node.Body != nil {
				visit(*node.Body)
			} else if node.Kind == contracts.Call && !seenCalls[node.Meaning] {
				seenCalls[node.Meaning] = true
				if definition, ok := a.Plan.Definitions[node.Meaning]; ok {
					visit(definition)
				}
			}
			for _, children := range [][]contracts.Node{node.Children, node.Then, node.Else, node.Cleanup} {
				for _, child := range children {
					visit(child)
				}
			}
		}
		visit(nodes[scope.NodeID])
	}
}

func (r *contractRecorder) recordContractScope(node contracts.Node, target contracts.TargetRef, origin contracts.Origin, firstUse int, meaning contracts.MeaningKey) {
	if r.nativeAnalysis == nil {
		return
	}
	scope := contracts.AnalysisScope{ID: contracts.StableID("source-operation-scope", node.ID), NodeID: node.ID, Meaning: meaning, Target: target, Origin: origin, Phase: contracts.Invocation, Boundary: "source_operation", Complete: true, Includes: []string{"ordered argument evaluation and selected operation", "modeled operation-owned construction, replacement and callee cleanup"}, Exclusions: []string{"enclosing activation acquisition and finalization, including activation-scoped emitter temporaries", "workspace startup and finalization"}}
	for _, use := range r.nativeAnalysis.uses[firstUse:] {
		scope.NativeUses = append(scope.NativeUses, use.ID)
		for _, location := range orderedContractResources(use.Resources) {
			if location.ID != "" && !location.Unknown {
				scope.Resources = append(scope.Resources, contracts.FailureLocation{Resource: location})
			}
		}
	}
	if scope.Meaning == "" && len(r.nativeAnalysis.uses[firstUse:]) == 1 {
		scope.Meaning = r.nativeAnalysis.uses[firstUse].Meaning
	}
	if scope.Meaning == "" {
		scope.Meaning = contracts.MeaningKey(contracts.StableID("typed-operation", node.ID))
	}
	if scope.Target.Unknown || scope.Target.Root == "" {
		scope.Target = contracts.TargetRef{Root: node.ID, Type: "typed-operation", Phase: contracts.Invocation}
	}
	scope.Occurrence = contracts.BindingOccurrenceID(contracts.StableID("operation-occurrence", node.ID))
	r.nativeAnalysis.scopes = append(r.nativeAnalysis.scopes, scope)
}

func orderedContractResources(resources map[string]contracts.ResourceRef) []contracts.ResourceRef {
	keys := make([]string, 0, len(resources))
	for key := range resources {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var result []contracts.ResourceRef
	seen := make(map[string]bool)
	for _, key := range keys {
		resource := resources[key]
		identity := contracts.StableID("resource", resource)
		if !seen[identity] {
			result = append(result, resource)
			seen[identity] = true
		}
	}
	return result
}

func (a *ContractAnalysis) FindContractScope(id string) (contracts.AnalysisScope, error) {
	var found []contracts.AnalysisScope
	for _, scope := range a.Scopes {
		selector := fmt.Sprintf("%s:%d", scope.Origin.Source, scope.Origin.Start)
		if scope.ID == id || string(scope.Occurrence) == id || selector == id {
			found = append(found, scope)
		}
	}
	if len(found) != 1 {
		return contracts.AnalysisScope{}, fmt.Errorf("contract scope %q selects %d typed boundaries; use an exact scope identity", id, len(found))
	}
	return found[0], nil
}
