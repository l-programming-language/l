package Lyger

import (
	"errors"
	"fmt"
	"sort"

	"github.com/l-programming-language/l/contracts"
)

var ErrSourceContractMetadata = errors.New("invalid source contract metadata")

type contractConstructionPath struct {
	Child   bool
	Nominal string
	Index   int
	Field   string
}

type contractConstructionOperation struct {
	Path  []contractConstructionPath
	Token *token
}

type contractSourceConstruction struct {
	Token         *token
	Descriptor    *record
	Rule          *descriptorRule
	Target        *constructionNode
	Before, After *constructionNode
	Operations    []contractConstructionOperation
	Detached      bool
}

func cloneContractConstruction(node *constructionNode) *constructionNode {
	if node == nil {
		return nil
	}
	result := *node
	result.Children = make([]*constructionNode, len(node.Children))
	for i, child := range node.Children {
		result.Children[i] = cloneContractConstruction(child)
	}
	result.Wrappers = append([]*descriptorRule(nil), node.Wrappers...)
	result.Parameters = append([]*objectGram(nil), node.Parameters...)
	result.Inputs = append([]DefinitionProperty(nil), node.Inputs...)
	result.Outputs = append([]string(nil), node.Outputs...)
	result.contractOperations = append([]contractConstructionOperation(nil), node.contractOperations...)
	result.contractRules = append([]*descriptorRule(nil), node.contractRules...)
	return &result
}

func (r *contractRecorder) recordSourceConstruction(tok *token, descriptor *record, rule *descriptorRule, node, before *constructionNode) {
	if before == nil {
		return
	}
	if node.contractIdentity == "" {
		sourceToken := tok
		if node.Source != nil {
			sourceToken = node.Source.noun
		}
		node.contractIdentity = contracts.StableID("source-delta", struct {
			Origin  contracts.Origin
			Ordinal int
		}{r.origin(sourceToken), len(r.sourceConstructions)})
	}
	r.sourceConstructions = append(r.sourceConstructions, contractSourceConstruction{Token: tok, Descriptor: descriptor, Rule: rule, Target: node, Before: before, After: cloneContractConstruction(node), Operations: append([]contractConstructionOperation(nil), node.contractOperations[len(before.contractOperations):]...)})
}

func (node *constructionNode) contractWrite(field string, tok *token, enabled bool) {
	if enabled {
		node.contractOperations = append(node.contractOperations, contractConstructionOperation{Path: []contractConstructionPath{{Field: field}}, Token: tok})
	}
}

func shiftedContractOperations(operations []contractConstructionOperation, offset int) []contractConstructionOperation {
	result := make([]contractConstructionOperation, len(operations))
	for i, operation := range operations {
		result[i] = operation
		result[i].Path = append([]contractConstructionPath(nil), operation.Path...)
		if len(result[i].Path) > 0 && result[i].Path[0].Child {
			result[i].Path[0].Index += offset
		}
	}
	return result
}

func childContractOperations(child *constructionNode, index int, tok *token) []contractConstructionOperation {
	prefix := contractConstructionPath{Child: true, Nominal: child.Nominal, Index: index}
	result := []contractConstructionOperation{{Path: []contractConstructionPath{prefix}, Token: tok}}
	for _, operation := range child.contractOperations {
		operation.Path = append([]contractConstructionPath{prefix}, operation.Path...)
		result = append(result, operation)
	}
	return result
}

func (r *contractRecorder) sourceNominal(identity nominalIdentity) contracts.DefinitionKey {
	if identity.definition != nil {
		return r.definition(identity.definition)
	}
	return contracts.DefinitionKey(contracts.StableID("native-nominal", struct {
		Native  valueType
		Lexical partOfSpeech
	}{identity.native, identity.lexical}))
}

func (r *contractRecorder) sourceShape(node *constructionNode, target contracts.TargetRef) contracts.SourceShape {
	shape := contracts.SourceShape{Target: target, ChildrenClosed: true, NameKnown: true, SignatureKnown: true, ParametersKnown: true, WrappersKnown: true}
	if node == nil {
		return shape
	}
	if node.Source != nil {
		if identity, err := r.compilation.resolver().nominalIdentity(node.Source.noun); err == nil {
			shape.Nominal = r.sourceNominal(identity)
		}
	}
	shape.Constructor = r.definition(node.Definition)
	shape.Target.Nominal = shape.Nominal
	shape.Name = node.Name
	shape.InputsDefined, shape.OutputsDefined = node.InputsDefined, node.OutputsDefined
	for _, input := range node.Inputs {
		shape.Inputs = append(shape.Inputs, contracts.SourceSignatureInput{Name: input.Key, Type: input.Type})
	}
	shape.Outputs = append([]string(nil), node.Outputs...)
	for _, wrapper := range node.Wrappers {
		shape.Wrappers = append(shape.Wrappers, r.sourceRuleMeaning(wrapper))
	}
	for index, parameter := range node.Parameters {
		meaning := contracts.MeaningKey("")
		if parameter.noun != nil && parameter.noun.binding != nil {
			meaning = r.meaning(parameter.noun.binding.Record, nil)
		}
		shape.Parameters = append(shape.Parameters, contracts.SourceParameter{Meaning: meaning, Origin: r.origin(parameter.noun), Formal: parameter.article != nil && parameter.article.hasRole("article", "variable"), Ordinal: index})
	}
	for index, child := range node.Children {
		childTarget := target
		childTarget.Path = append(append([]string(nil), target.Path...), fmt.Sprintf("%s[%d]", child.Nominal, index))
		childTarget.Type = child.Nominal
		shape.Children = append(shape.Children, r.sourceShape(child, childTarget))
	}
	return shape
}

func (r *contractRecorder) sourceRuleMeaning(rule *descriptorRule) contracts.MeaningKey {
	for rec, descriptor := range r.compilation.Semantics.Descriptors {
		for _, candidate := range descriptor.Rules {
			if rule == candidate {
				return r.meaning(rec, rule)
			}
		}
	}
	return ""
}

func (r *contractRecorder) sourceFootprints(operations []contractConstructionOperation, target contracts.TargetRef) []contracts.FrameFootprint {
	var footprints []contracts.FrameFootprint
	for _, operation := range operations {
		actual := target
		actual.Path = append([]string(nil), target.Path...)
		for _, step := range operation.Path {
			if step.Child {
				actual.Path = append(actual.Path, fmt.Sprintf("%s[%d]", step.Nominal, step.Index))
			} else {
				actual.Path = append(actual.Path, step.Field)
			}
		}
		footprints = append(footprints, contracts.FrameFootprint{Target: actual, Kind: "write", Definite: true, Origin: r.origin(operation.Token)})
	}
	return footprints
}

func sourceProjectionPlan(footprints []contracts.FrameFootprint) contracts.Plan {
	plan := contracts.Plan{Entry: contracts.Node{Kind: contracts.Sequence, Phase: contracts.Elaboration}}
	for index, footprint := range footprints {
		plan.Entry.Children = append(plan.Entry.Children, contracts.Node{Kind: contracts.Event, ID: contracts.StableID("construction-operation", struct {
			Footprint contracts.FrameFootprint
			Ordinal   int
		}{footprint, index}), Origin: footprint.Origin, Phase: contracts.Elaboration, Effects: []contracts.Effect{{Kind: "write", Target: footprint.Target, Phase: contracts.Elaboration, Origin: footprint.Origin, Definite: true}}})
	}
	return plan
}

func (r *contractRecorder) sourceProjection(rec *record, rule *descriptorRule) contracts.SourceAdjectiveProjection {
	meaning := r.meaning(rec, rule)
	target := contracts.SourceSymbolicTarget(meaning, rule.Target.String())
	target.Nominal = r.sourceNominal(rule.Target)
	origin := r.origin(rule.Object.noun)
	if rule.Source != nil && rule.Source.source != nil {
		source := rule.Source.source
		origin = contracts.Origin{Source: r.logicalSource(source.Span.Path), Start: source.Span.Start.Offset, End: source.Span.End.Offset, Parent: r.occurrence(source.occurrence)}
	}
	projection := contracts.SourceAdjectiveProjection{Meaning: meaning, Version: contracts.SourceProjectionVersion, Target: target, TargetNominal: target.Nominal, Origin: origin, Environment: r.environment(rule.Object.noun.scope), Dependencies: make(map[string]string), Plan: contracts.Plan{Entry: contracts.Node{Kind: contracts.Sequence, Phase: contracts.Elaboration}}}
	projection.Dependencies["adapter"] = contracts.SourceProjectionVersion
	projection.Dependencies["configuration"] = r.configurationIdentity()
	projection.Dependencies["definition-environment"] = string(projection.Environment)
	projection.Dependencies["body"] = contracts.StableID("source-rule", rule.Source.source.Raw)
	before := contracts.SourceShape{Target: target, Nominal: target.Nominal}
	projection.Before, projection.After = before, before
	projection.LoweringObligations = []contracts.Obligation{{ID: contracts.StableID("source-lowering", meaning), Claim: contracts.Claim{Predicate: contracts.Predicate{Kind: "native_lowering_conformance"}, Target: target, Scope: "implementation", Phase: contracts.Invocation}, Status: contracts.Unresolved, Reason: "source topology and parameter forwarding do not establish executable template behavior, callback demand or native resource/lifetime conformance", Origin: origin}}
	var delta *constructionNode
	sourceWrapper, err := r.compilation.sourceDescriptorObject(rule)
	if err == nil && (rule.Template != nil || sourceWrapper != nil) {
		// The real template-rule branch only appends a wrapper. Its described
		// object's other spelling is not a constructor mutation.
		delta = &constructionNode{Nominal: rule.Target.String(), Source: rule.Object, Wrappers: []*descriptorRule{rule}}
		delta.contractWrite("wrappers", rule.Object.noun, true)
	} else if err == nil {
		delta, err = r.compilation.constructObjectMode(rule.Object, make(map[*descriptorRule]bool), constructionSymbolic)
	}
	if err != nil {
		projection.ProjectionObligations = append(projection.ProjectionObligations, contracts.Obligation{ID: contracts.StableID("unsupported-source", meaning), Claim: contracts.Claim{Predicate: contracts.Predicate{Kind: "source_projection", Subject: string(meaning), Target: target}, Target: target, Scope: "source_elaboration", Phase: contracts.Elaboration}, Status: contracts.Unresolved, Origin: origin, Reason: "source structural projection is unsupported: " + err.Error()})
		return projection
	}
	projection.After = r.sourceShape(delta, target)
	if reason := r.unsupportedSourceStructure(delta); reason != "" {
		projection.ProjectionObligations = append(projection.ProjectionObligations, contracts.Obligation{ID: contracts.StableID("source-reflection", meaning), Claim: contracts.Claim{Predicate: contracts.Predicate{Kind: "source_projection", Subject: string(meaning), Target: target}, Target: target, Scope: "source_elaboration", Phase: contracts.Elaboration}, Status: contracts.Unresolved, Origin: origin, Reason: reason})
	}
	projection.After.NameKnown = delta.Name != ""
	projection.After.ChildrenClosed = false
	projection.After.SignatureKnown = false
	projection.After.ParametersKnown = false
	projection.After.WrappersKnown = false
	assume := func(kind string) {
		projection.RequiredAssumptions = append(projection.RequiredAssumptions, contracts.Predicate{Kind: kind, Target: target})
	}
	if len(delta.Children) > 0 {
		assume("source_children_empty")
		projection.Before.ChildrenClosed = true
		projection.After.ChildrenClosed = true
	}
	if delta.InputsDefined || delta.OutputsDefined || len(delta.Inputs) > 0 || len(delta.Outputs) > 0 {
		assume("source_signature_empty")
		projection.Before.SignatureKnown = true
		projection.After.SignatureKnown = true
	}
	if len(delta.Parameters) > 0 {
		assume("source_parameters_empty")
		projection.Before.ParametersKnown = true
		projection.After.ParametersKnown = true
	}
	if len(delta.Wrappers) > 0 {
		assume("source_wrappers_empty")
		projection.Before.WrappersKnown = true
		projection.After.WrappersKnown = true
	}
	projection.Footprints = r.sourceFootprints(delta.contractOperations, target)
	projection.Plan = sourceProjectionPlan(projection.Footprints)
	r.sourceDependencies(delta, projection.Dependencies)
	return projection
}

func (r *contractRecorder) unsupportedSourceStructure(node *constructionNode) string {
	formal := func(tok *token) bool {
		if tok == nil {
			return false
		}
		if binding, found := r.compilation.Scope.Pronouns[tok]; found && binding.Kind == scopeActionArgument {
			return true
		}
		for _, statement := range r.compilation.Scope.Ordered {
			for _, argument := range statement.Arguments {
				if argument.Token == tok {
					return true
				}
			}
		}
		return false
	}
	objects := append([]*objectGram{node.Source}, node.Parameters...)
	for _, object := range objects {
		if object == nil {
			continue
		}
		for _, tok := range object.tokens {
			if _, possessive := possessiveOwner(string(tok.rawBytes)); possessive && (formal(tok) || tok.isPronoun()) {
				return "formal-dependent member selection remains instantiation-only"
			}
		}
	}
	for _, child := range node.Children {
		if reason := r.unsupportedSourceStructure(child); reason != "" {
			return reason
		}
	}
	return ""
}

func (r *contractRecorder) sourceDependencies(node *constructionNode, dependencies map[string]string) {
	if node.Definition != nil {
		key := r.definition(node.Definition)
		if statement := r.compilation.Semantics.Declarations[node.Definition]; statement != nil && statement.source != nil {
			dependencies["definition:"+string(key)] = contracts.StableID("source-body", statement.source.Raw)
		} else {
			dependencies["definition:"+string(key)] = r.configurationIdentity()
		}
	}
	for _, wrapper := range node.Wrappers {
		dependencies["rule:"+string(r.sourceRuleMeaning(wrapper))] = contracts.StableID("source-body", wrapper.Source.source.Raw)
	}
	for _, rule := range node.contractRules {
		dependencies["rule:"+string(r.sourceRuleMeaning(rule))] = contracts.StableID("source-body", rule.Source.source.Raw)
	}
	for _, child := range node.Children {
		r.sourceDependencies(child, dependencies)
	}
}

func (a *ContractAnalysis) verifySourceContracts(options ContractAnalysisOptions) error {
	a.SourceProjections, a.SourceCertificates, a.SourceInstantiations = nil, nil, nil
	r := a.Compilation.Compilation.contractRecorder
	declarations := make(map[contracts.MeaningKey]contracts.SourceTransformerDeclaration)
	for _, declaration := range options.SourceTransformers {
		if declaration.Meaning == "" || declaration.Version == "" {
			return fmt.Errorf("%w: source declaration requires meaning and version", ErrSourceContractMetadata)
		}
		if _, duplicate := declarations[declaration.Meaning]; duplicate {
			return fmt.Errorf("%w: duplicate declaration for %s", ErrSourceContractMetadata, declaration.Meaning)
		}
		declarations[declaration.Meaning] = declaration
	}
	type sourceRule struct {
		rec     *record
		rule    *descriptorRule
		meaning contracts.MeaningKey
	}
	var rules []sourceRule
	for rec, descriptor := range r.compilation.Semantics.Descriptors {
		for _, rule := range descriptor.Rules {
			rules = append(rules, sourceRule{rec, rule, r.meaning(rec, rule)})
		}
	}
	sort.Slice(rules, func(i, j int) bool { return rules[i].meaning < rules[j].meaning })
	projections := make(map[contracts.MeaningKey]contracts.SourceAdjectiveProjection)
	certificates := make(map[contracts.MeaningKey]contracts.SourceAdjectiveCertificate)
	for _, entry := range rules {
		projection := r.sourceProjection(entry.rec, entry.rule)
		declaration := declarations[entry.meaning]
		if declaration.Meaning != "" && declaration.TargetNominal != projection.TargetNominal {
			return fmt.Errorf("%w: declared nominal differs from selected rule %s", ErrSourceContractMetadata, entry.meaning)
		}
		verify := contracts.VerifySourceAdjective
		if options.Session != nil && declaration.Meaning != "" {
			verify = options.Session.VerifySourceAdjective
		}
		certificate, err := verify(projection, declaration, options.Inference)
		if err != nil {
			return err
		}
		a.SourceProjections = append(a.SourceProjections, projection)
		a.SourceCertificates = append(a.SourceCertificates, certificate)
		projections[entry.meaning], certificates[entry.meaning] = projection, certificate
	}
	for meaning := range declarations {
		if _, found := projections[meaning]; !found {
			return fmt.Errorf("%w: canonical descriptor rule %s is absent", ErrSourceContractMetadata, meaning)
		}
	}
	for _, use := range r.sourceConstructions {
		meaning := r.meaning(use.Descriptor, use.Rule)
		declaration, declared := declarations[meaning]
		if !declared {
			continue
		}
		target, known := r.targets[use.Target]
		if use.Detached {
			target = contracts.TargetRef{Root: use.Target.contractIdentity, Generation: "source-construction-delta", Type: use.Target.Nominal}
		} else if !known {
			target = contracts.TargetRef{Unknown: true, Type: use.Target.Nominal}
		}
		target.Phase = contracts.Elaboration
		target.Nominal = r.sourceNominal(use.Rule.Target)
		instance := contracts.SourceAdjectiveInstance{Meaning: meaning, Target: target, TargetNominal: target.Nominal, Before: r.sourceShape(use.Before, target), After: r.sourceShape(use.After, target), Footprints: r.sourceFootprints(use.Operations, target), Origin: r.origin(use.Token), LoweringObligations: projections[meaning].LoweringObligations}
		verify := contracts.VerifySourceAdjectiveInstance
		if options.Session != nil {
			verify = options.Session.VerifySourceAdjectiveInstance
		}
		result, err := verify(projections[meaning], certificates[meaning], declaration, instance, options.Inference)
		if err != nil {
			return err
		}
		a.SourceInstantiations = append(a.SourceInstantiations, result)
	}
	return nil
}
