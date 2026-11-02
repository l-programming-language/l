package contracts

import (
	"fmt"
	"strconv"
	"strings"
)

const SourceProjectionVersion = "l-source-structural/1"

// SourceTransformerDeclaration is bodyless. Its promises are checked against a
// projection obtained from the compiler, never a body supplied by metadata.
type SourceTransformerDeclaration struct {
	Meaning       MeaningKey     `json:"meaning"`
	Version       string         `json:"version"`
	TargetNominal DefinitionKey  `json:"target_nominal"`
	Requires      []Predicate    `json:"requires,omitempty"`
	Provides      []Predicate    `json:"provides,omitempty"`
	Frame         *FrameContract `json:"frame,omitempty"`
	Profiles      []Profile      `json:"profiles,omitempty"`
	Origin        Origin         `json:"origin"`
}

type SourceSignatureInput struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type SourceParameter struct {
	Meaning MeaningKey `json:"meaning"`
	Origin  Origin     `json:"origin"`
	Formal  bool       `json:"formal"`
	Ordinal int        `json:"ordinal"`
}

type SourceShape struct {
	Target          TargetRef              `json:"target"`
	Nominal         DefinitionKey          `json:"nominal"`
	Constructor     DefinitionKey          `json:"constructor,omitempty"`
	Name            string                 `json:"name,omitempty"`
	NameKnown       bool                   `json:"name_known"`
	Children        []SourceShape          `json:"children,omitempty"`
	ChildrenClosed  bool                   `json:"children_closed"`
	InputsDefined   bool                   `json:"inputs_defined"`
	OutputsDefined  bool                   `json:"outputs_defined"`
	SignatureKnown  bool                   `json:"signature_known"`
	ParametersKnown bool                   `json:"parameters_known"`
	WrappersKnown   bool                   `json:"wrappers_known"`
	Inputs          []SourceSignatureInput `json:"inputs,omitempty"`
	Outputs         []string               `json:"outputs,omitempty"`
	Wrappers        []MeaningKey           `json:"wrappers,omitempty"`
	Parameters      []SourceParameter      `json:"parameters,omitempty"`
}

// SourceAdjectiveProjection has no document decoder or metadata loader. The
// frontend produces it from resolved constructor operations. Its hashes record
// provenance; a caller-authored hash or label is not authentication or evidence.
type SourceAdjectiveProjection struct {
	Meaning               MeaningKey            `json:"meaning"`
	Version               string                `json:"version"`
	Target                TargetRef             `json:"target"`
	TargetNominal         DefinitionKey         `json:"target_nominal"`
	Origin                Origin                `json:"origin"`
	Environment           BindingEnvironmentKey `json:"environment"`
	Dependencies          map[string]string     `json:"dependencies"`
	Plan                  Plan                  `json:"plan"`
	Before                SourceShape           `json:"before"`
	After                 SourceShape           `json:"after"`
	Footprints            []FrameFootprint      `json:"footprints,omitempty"`
	RequiredAssumptions   []Predicate           `json:"required_assumptions,omitempty"`
	ProjectionObligations []Obligation          `json:"projection_obligations,omitempty"`
	LoweringObligations   []Obligation          `json:"lowering_obligations,omitempty"`
}

type SourceAdjectiveCertificate struct {
	Meaning                MeaningKey            `json:"meaning"`
	Target                 TargetRef             `json:"target"`
	TargetNominal          DefinitionKey         `json:"target_nominal"`
	Version                string                `json:"version"`
	Scope                  string                `json:"scope"`
	Mode                   string                `json:"mode"`
	ProjectionFingerprint  string                `json:"projection_fingerprint"`
	DeclarationFingerprint string                `json:"declaration_fingerprint"`
	Environment            BindingEnvironmentKey `json:"environment"`
	Dependencies           map[string]string     `json:"dependencies"`
	LogicalStatus          Status                `json:"logical_status"`
	LogicalObligations     []Obligation          `json:"logical_obligations"`
	LoweringObligations    []Obligation          `json:"lowering_obligations"`
	Summary                DerivedSummary        `json:"summary"`
	Before                 SourceShape           `json:"before"`
	After                  SourceShape           `json:"after"`
	Assumptions            []string              `json:"assumptions,omitempty"`
}

type SourceAdjectiveInstance struct {
	Meaning             MeaningKey       `json:"meaning"`
	Target              TargetRef        `json:"target"`
	TargetNominal       DefinitionKey    `json:"target_nominal"`
	Before              SourceShape      `json:"before"`
	After               SourceShape      `json:"after"`
	Footprints          []FrameFootprint `json:"footprints,omitempty"`
	Origin              Origin           `json:"origin"`
	LoweringObligations []Obligation     `json:"lowering_obligations,omitempty"`
}

type SourceAdjectiveInstantiation struct {
	Meaning                MeaningKey   `json:"meaning"`
	Target                 TargetRef    `json:"target"`
	Status                 Status       `json:"status"`
	Obligations            []Obligation `json:"obligations"`
	LoweringObligations    []Obligation `json:"lowering_obligations"`
	CertificateFingerprint string       `json:"certificate_fingerprint"`
}

type SourceVerificationResult struct {
	Assurance   string       `json:"assurance"`
	Scope       string       `json:"scope"`
	Selected    []MeaningKey `json:"selected"`
	Status      Status       `json:"status"`
	Obligations []Obligation `json:"obligations"`
}

func SourceSymbolicTarget(meaning MeaningKey, displayType string) TargetRef {
	return TargetRef{Root: StableID("symbolic-source-target", meaning), Type: displayType, Phase: Elaboration}
}

func SourceShapeFacts(shape SourceShape) []Predicate {
	if shape.Target.Unknown || shape.Target.Root == "" {
		return nil
	}
	facts := []Predicate{{Kind: "present", Subject: "target", Target: shape.Target}}
	if shape.Nominal != "" {
		facts = append(facts, Predicate{Kind: "meaning", Subject: "nominal", Target: shape.Target, Value: Atom{Kind: "meaning", Text: string(shape.Nominal)}})
	}
	if shape.Constructor != "" {
		facts = append(facts, Predicate{Kind: "meaning", Subject: "constructor", Target: shape.Target, Value: Atom{Kind: "meaning", Text: string(shape.Constructor)}})
	}
	if shape.NameKnown {
		facts = append(facts, Predicate{Kind: "eq", Subject: "name", Target: shape.Target, Value: Atom{Kind: "text", Text: shape.Name}})
	}
	valueFact := func(subject string, value Atom) {
		facts = append(facts, Predicate{Kind: "eq", Subject: subject, Target: shape.Target, Value: value})
	}
	if shape.InputsDefined {
		valueFact("inputs.defined", Atom{Kind: "bool", Bool: shape.InputsDefined})
	}
	if shape.OutputsDefined {
		valueFact("outputs.defined", Atom{Kind: "bool", Bool: shape.OutputsDefined})
	}
	if shape.SignatureKnown {
		valueFact("inputs.count", Atom{Kind: "number", Number: int64(len(shape.Inputs))})
		valueFact("outputs.count", Atom{Kind: "number", Number: int64(len(shape.Outputs))})
		for i, input := range shape.Inputs {
			valueFact(fmt.Sprintf("inputs[%d].name", i), Atom{Kind: "text", Text: input.Name})
			valueFact(fmt.Sprintf("inputs[%d].type", i), Atom{Kind: "type", Text: input.Type})
		}
		for i, output := range shape.Outputs {
			valueFact(fmt.Sprintf("outputs[%d].type", i), Atom{Kind: "type", Text: output})
		}
	}
	if shape.WrappersKnown {
		valueFact("wrappers.count", Atom{Kind: "number", Number: int64(len(shape.Wrappers))})
		for i, wrapper := range shape.Wrappers {
			valueFact(fmt.Sprintf("wrappers[%d].meaning", i), Atom{Kind: "meaning", Text: string(wrapper)})
		}
	}
	if shape.ParametersKnown {
		valueFact("parameters.count", Atom{Kind: "number", Number: int64(len(shape.Parameters))})
		for i, parameter := range shape.Parameters {
			valueFact(fmt.Sprintf("parameters[%d].formal", i), Atom{Kind: "bool", Bool: parameter.Formal})
			if parameter.Meaning != "" {
				valueFact(fmt.Sprintf("parameters[%d].meaning", i), Atom{Kind: "meaning", Text: string(parameter.Meaning)})
			}
		}
	}
	empty := Predicate{Kind: "source_children_empty", Target: shape.Target}
	if shape.ChildrenClosed {
		facts = append(facts, Predicate{Kind: "eq", Subject: "children.count", Target: shape.Target, Value: Atom{Kind: "number", Number: int64(len(shape.Children))}})
		if len(shape.Children) == 0 {
			facts = append(facts, empty)
		}
	}
	if len(shape.Children) > 0 {
		facts = append(facts, Predicate{Kind: "not", Args: []Predicate{empty}})
	}
	for _, entry := range []struct {
		kind         string
		known, empty bool
	}{
		{"source_signature_empty", shape.SignatureKnown, len(shape.Inputs) == 0 && len(shape.Outputs) == 0},
		{"source_parameters_empty", shape.ParametersKnown, len(shape.Parameters) == 0},
		{"source_wrappers_empty", shape.WrappersKnown, len(shape.Wrappers) == 0},
	} {
		predicate := Predicate{Kind: entry.kind, Target: shape.Target}
		if !entry.empty {
			facts = append(facts, Predicate{Kind: "not", Args: []Predicate{predicate}})
		} else if entry.known {
			facts = append(facts, predicate)
		}
	}
	for _, child := range shape.Children {
		if child.Target.Unknown || child.Target.Root == "" {
			continue
		}
		facts = append(facts, Predicate{Kind: "present", Subject: "child", Target: child.Target})
		facts = append(facts, SourceShapeFacts(child)...)
	}
	return facts
}

func sourceObligation(target TargetRef, origin Origin, status Status, reason string, predicate Predicate) Obligation {
	result := facetObligation(target, origin, status, reason, predicate)
	result.Claim.Scope, result.Claim.Phase = "source_elaboration", Elaboration
	result.Claim.Outcomes = []Outcome{Normal}
	return result
}

func VerifySourceAdjective(projection SourceAdjectiveProjection, declaration SourceTransformerDeclaration, options AnalysisOptions) (SourceAdjectiveCertificate, error) {
	certificate := SourceAdjectiveCertificate{Meaning: projection.Meaning, Target: projection.Target, TargetNominal: projection.TargetNominal, Version: SourceProjectionVersion, Scope: "source_elaboration", Mode: "symbolic", ProjectionFingerprint: StableID("source-projection", projection), DeclarationFingerprint: StableID("source-declaration", declaration), Environment: projection.Environment, Dependencies: projection.Dependencies, LogicalStatus: Unresolved, Before: projection.Before, After: projection.After, LoweringObligations: append([]Obligation(nil), projection.LoweringObligations...)}
	add := func(status Status, reason string, predicate Predicate) {
		certificate.LogicalObligations = append(certificate.LogicalObligations, sourceObligation(projection.Target, projection.Origin, status, reason, predicate))
	}
	if projection.Meaning == "" || projection.Target.Root == "" || projection.Target.Unknown || projection.Version != SourceProjectionVersion || projection.Environment == "" {
		add(Unresolved, "source projection lacks canonical rule, exact symbolic attachment, supported adapter or definition-site environment", Predicate{Kind: "source_projection"})
	}
	if projection.Target.Phase != Elaboration || TargetRelation(projection.Before.Target, projection.Target) != Established || TargetRelation(projection.After.Target, projection.Target) != Established {
		add(Unresolved, "source projection snapshots must refer to the exact elaboration target", Predicate{Kind: "attachment"})
	}
	if projection.Before.Nominal != projection.TargetNominal || projection.After.Nominal != projection.TargetNominal || projection.Target.Nominal != "" && projection.Target.Nominal != projection.TargetNominal {
		add(Unresolved, "source projection snapshot nominal identity is missing or inconsistent", Predicate{Kind: "meaning", Subject: "nominal"})
	}
	if declaration.Meaning == "" || declaration.Version == "" {
		add(Unresolved, "source rule has no bodyless contract declaration; discovered shape is not a certificate", Predicate{Kind: "source_declaration"})
	} else if declaration.Meaning != projection.Meaning {
		return certificate, fmt.Errorf("source declaration names a different canonical descriptor rule")
	}
	if projection.TargetNominal == "" || declaration.TargetNominal == "" {
		add(Unresolved, "source attachment requires a canonical nominal identity", Predicate{Kind: "meaning", Subject: "nominal"})
	} else if projection.TargetNominal != declaration.TargetNominal {
		add(Contradicted, "source contract nominal domain differs from the resolved descriptor target", Predicate{Kind: "meaning", Subject: "nominal", Value: Atom{Kind: "meaning", Text: string(declaration.TargetNominal)}})
	}
	certificate.LogicalObligations = append(certificate.LogicalObligations, projection.ProjectionObligations...)
	if !AllEstablished(projection.ProjectionObligations) {
		certificate.Mode = "instantiation_only"
	}
	for _, assumption := range projection.RequiredAssumptions {
		status, reason := Entails(declaration.Requires, assumption)
		add(status, "required symbolic shape premise: "+reason, assumption)
	}
	for _, requirement := range declaration.Requires {
		if !sourceStructuralPredicate(requirement, projection.Target, 0) {
			add(Unresolved, "source precondition is outside the exact structural shape fragment", requirement)
		} else if status, reason := Entails([]Predicate{requirement}, requirement); status != Established {
			add(Unresolved, "source precondition is outside the supported fragment: "+reason, requirement)
		}
		certificate.Assumptions = append(certificate.Assumptions, StableID("source-premise", requirement))
	}
	// Analyze only explicit elaboration behavior. A projection carrying runtime
	// operations is incomplete, never silently filtered into a pure theorem.
	if !sourcePlanElaboration(projection.Plan) {
		certificate.Mode = "instantiation_only"
		add(Unresolved, "source structural projection includes unsupported runtime or opaque behavior", Predicate{Kind: "source_fragment"})
	}
	summary, err := Infer(projection.Plan, sourceLogicalOptions(options))
	if err != nil {
		return certificate, err
	}
	certificate.Summary = summary
	certificate.LogicalObligations = append(certificate.LogicalObligations, summary.Obligations...)
	if len(summary.Unknowns) != 0 {
		add(Unresolved, "source elaboration plan contains unresolved behavior", Predicate{Kind: "source_projection"})
	}
	writes := append(append([]FrameFootprint(nil), projection.Footprints...), FootprintsFromEffects(summary.Effects)...)
	post := invalidateSourceFacts(declaration.Requires, writes)
	post = postFacts(post, SourceShapeFacts(projection.After))
	for _, promise := range declaration.Provides {
		if !sourceStructuralPredicate(promise, projection.Target, 0) {
			add(Unresolved, "source postcondition is outside the exact structural shape fragment", promise)
			continue
		}
		status, reason := Entails(post, promise)
		add(status, "source structural postcondition: "+reason, promise)
	}
	if declaration.Frame != nil {
		if !sourceStructuralFrame(*declaration.Frame, projection.Target) {
			add(Unresolved, "source structural certificate requires an exact target elaboration frame", Predicate{Kind: "frame"})
		} else {
			certificate.LogicalObligations = append(certificate.LogicalObligations, CheckFrame(*declaration.Frame, writes, FrameContext{})...)
		}
	}
	for _, profile := range declaration.Profiles {
		add(Unresolved, "runtime profile "+profile.Name+" is not established by a source topology theorem", Predicate{Kind: "profile", Subject: profile.Name})
	}
	if len(declaration.Provides) == 0 && declaration.Frame == nil {
		add(Unresolved, "source declaration requests no structural postcondition or frame", Predicate{Kind: "source_guarantee"})
	}
	certificate.Assumptions = evidenceUnique(append(certificate.Assumptions, summary.Assumptions...))
	certificate.LogicalStatus = obligationStatus(certificate.LogicalObligations)
	return certificate, nil
}

// Source theorems describe constructor topology only. The general predicate
// language also covers runtime protocols, effects and costs; accepting those as
// premises here would let an unchanged caller assertion masquerade as source
// evidence. Every atomic source fact therefore names an exact structural field.
func sourceStructuralPredicate(predicate Predicate, target TargetRef, depth int) bool {
	if depth > 64 || predicate.Resource != (ResourceRef{}) {
		return false
	}
	switch predicate.Kind {
	case "true", "false":
		return len(predicate.Args) == 0
	case "and", "or", "not", "implies":
		if len(predicate.Args) == 0 || len(predicate.Args) > 64 || predicate.Kind == "not" && len(predicate.Args) != 1 || predicate.Kind == "implies" && len(predicate.Args) != 2 {
			return false
		}
		for _, child := range predicate.Args {
			if !sourceStructuralPredicate(child, target, depth+1) {
				return false
			}
		}
		return true
	}
	if len(predicate.Args) != 0 || predicate.Target.Unknown || predicate.Target.Root == "" || predicate.Target.Phase != Elaboration || !pathContains(target, predicate.Target) {
		return false
	}
	switch predicate.Kind {
	case "source_children_empty", "source_signature_empty", "source_parameters_empty", "source_wrappers_empty":
		return predicate.Subject == ""
	case "present":
		return predicate.Subject == "target" || predicate.Subject == "child"
	case "meaning":
		return (predicate.Subject == "nominal" || predicate.Subject == "constructor") && predicate.Value.Kind == "meaning" && predicate.Value.Text != ""
	case "eq":
		kind := sourceShapeFieldKind(predicate.Subject)
		return kind != "" && predicate.Value.Kind == kind
	case "interval", "finite":
		return strings.HasSuffix(predicate.Subject, ".count") && sourceShapeFieldKind(predicate.Subject) == "number"
	default:
		return false
	}
}

func sourceShapeFieldKind(subject string) string {
	switch subject {
	case "name":
		return "text"
	case "inputs.defined", "outputs.defined":
		return "bool"
	case "children.count", "inputs.count", "outputs.count", "wrappers.count", "parameters.count":
		return "number"
	}
	start, end := strings.IndexByte(subject, '['), strings.IndexByte(subject, ']')
	if start < 1 || end <= start+1 {
		return ""
	}
	index, err := strconv.Atoi(subject[start+1 : end])
	if err != nil || index < 0 || strconv.Itoa(index) != subject[start+1:end] {
		return ""
	}
	switch subject[:start] + subject[end+1:] {
	case "inputs.name":
		return "text"
	case "inputs.type", "outputs.type":
		return "type"
	case "wrappers.meaning", "parameters.meaning":
		return "meaning"
	case "parameters.formal":
		return "bool"
	default:
		return ""
	}
}

func sourceStructuralFrame(frame FrameContract, target TargetRef) bool {
	if frame.Phase != Elaboration || frame.Scope != "elaboration" || frame.Resource != (ResourceRef{}) || TargetRelation(frame.Target, target) != Established {
		return false
	}
	for _, path := range append(append([]TargetRef(nil), frame.Allowed...), frame.Preserves...) {
		if !pathContains(target, path) {
			return false
		}
	}
	return true
}

func sourceLogicalOptions(options AnalysisOptions) AnalysisOptions {
	options.NativeSummaries = nil
	options.NativeWitnesses = nil
	options.ApprovedAssumptions = nil
	return options
}

func sourcePlanElaboration(plan Plan) bool {
	if len(plan.NativeClosure) != 0 || len(plan.Residuals) != 0 || len(plan.Callbacks) != 0 {
		return false
	}
	var supported func(Node) bool
	supported = func(node Node) bool {
		if node.Kind != Sequence && node.Kind != Event && node.Kind != Require || node.Phase != "" && node.Phase != Elaboration {
			return false
		}
		if node.Outcome != "" && node.Outcome != Normal {
			return false
		}
		for _, effect := range node.Effects {
			if effect.Phase != Elaboration || effect.Kind != "write" && effect.Kind != "replace" || effect.Target.Unknown || effect.Target.Root == "" {
				return false
			}
		}
		for _, child := range node.Children {
			if !supported(child) {
				return false
			}
		}
		return len(node.Then) == 0 && len(node.Else) == 0 && len(node.Cleanup) == 0
	}
	if !supported(plan.Entry) {
		return false
	}
	for _, body := range plan.Definitions {
		if !supported(body) {
			return false
		}
	}
	return true
}

func invalidateSourceFacts(facts []Predicate, writes []FrameFootprint) []Predicate {
	var retained []Predicate
	for _, fact := range facts {
		valid := true
		for _, write := range writes {
			if structuralFactAffected(fact, write.Target, write.Kind) {
				valid = false
				break
			}
		}
		if valid {
			retained = append(retained, fact)
		}
	}
	return retained
}

func structuralFactAffected(fact Predicate, target TargetRef, kind string) bool {
	for _, child := range fact.Args {
		if structuralFactAffected(child, target, kind) {
			return true
		}
	}
	if fact.Target.Root == "" && !fact.Target.Unknown {
		return false
	}
	if target.Unknown || target.Root == "" || fact.Target.Unknown || kind == "unknown" {
		return true
	}
	if kind == "write" && (fact.Kind == "type" && fact.Subject == "target" || fact.Kind == "meaning" && fact.Subject == "nominal" && fact.Target.Nominal != "" && string(fact.Target.Nominal) == fact.Value.Text) {
		return false
	}
	for _, dependency := range structuralFactTargets(fact) {
		if pathContains(dependency, target) || pathContains(target, dependency) {
			return true
		}
	}
	return false
}

func structuralFactTargets(fact Predicate) []TargetRef {
	var fields []string
	switch {
	case fact.Kind == "source_signature_empty":
		fields = []string{"inputs", "outputs"}
	case fact.Kind == "source_parameters_empty":
		fields = []string{"parameters"}
	case fact.Kind == "source_wrappers_empty":
		fields = []string{"wrappers"}
	case fact.Subject == "name":
		fields = []string{"name"}
	case fact.Subject == "constructor":
		fields = []string{"constructor"}
	case strings.HasPrefix(fact.Subject, "inputs.") || strings.HasPrefix(fact.Subject, "inputs["):
		fields = []string{"inputs"}
	case strings.HasPrefix(fact.Subject, "outputs.") || strings.HasPrefix(fact.Subject, "outputs["):
		fields = []string{"outputs"}
	case strings.HasPrefix(fact.Subject, "parameters.") || strings.HasPrefix(fact.Subject, "parameters["):
		fields = []string{"parameters"}
	case strings.HasPrefix(fact.Subject, "wrappers.") || strings.HasPrefix(fact.Subject, "wrappers["):
		fields = []string{"wrappers"}
	default:
		return []TargetRef{fact.Target}
	}
	var targets []TargetRef
	for _, field := range fields {
		target := fact.Target
		target.Path = append(append([]string(nil), target.Path...), field)
		targets = append(targets, target)
	}
	return targets
}

func VerifySourceAdjectiveInstance(projection SourceAdjectiveProjection, certificate SourceAdjectiveCertificate, declaration SourceTransformerDeclaration, instance SourceAdjectiveInstance, options AnalysisOptions) (SourceAdjectiveInstantiation, error) {
	verified, err := VerifySourceAdjective(projection, declaration, options)
	result := SourceAdjectiveInstantiation{Meaning: projection.Meaning, Target: instance.Target, Status: Unresolved, CertificateFingerprint: StableID("source-certificate", verified), LoweringObligations: append(append([]Obligation(nil), verified.LoweringObligations...), instance.LoweringObligations...)}
	if err != nil {
		return result, err
	}
	result.Obligations = append(result.Obligations, verified.LogicalObligations...)
	add := func(status Status, reason string, predicate Predicate) {
		result.Obligations = append(result.Obligations, sourceObligation(instance.Target, instance.Origin, status, reason, predicate))
	}
	if certificate.ProjectionFingerprint != verified.ProjectionFingerprint || certificate.DeclarationFingerprint != verified.DeclarationFingerprint {
		add(Unresolved, "source certificate dependencies changed", Predicate{Kind: "certificate"})
	}
	if instance.Meaning != projection.Meaning || instance.Target.Unknown || instance.Target.Root == "" || instance.Target.Phase != Elaboration || instance.TargetNominal == "" {
		add(Unresolved, "source instance lacks exact selected rule, attachment or nominal identity", Predicate{Kind: "attachment"})
	} else if instance.TargetNominal != projection.TargetNominal {
		add(Contradicted, "source instance canonical nominal identity differs from the certified domain", Predicate{Kind: "meaning", Subject: "nominal"})
	}
	if TargetRelation(instance.Before.Target, instance.Target) != Established || TargetRelation(instance.After.Target, instance.Target) != Established || instance.Before.Nominal != instance.TargetNominal || instance.After.Nominal != instance.TargetNominal || instance.Target.Nominal != "" && instance.Target.Nominal != instance.TargetNominal {
		add(Unresolved, "concrete before/after snapshots do not identify the selected nominal target", Predicate{Kind: "attachment"})
	}
	requirements := rebasePredicates(declaration.Requires, projection.Target, instance.Target)
	for _, requirement := range requirements {
		status, reason := Entails(SourceShapeFacts(instance.Before), requirement)
		add(status, "concrete before-rule requirement: "+reason, requirement)
	}
	for _, required := range rebasePredicates(projection.RequiredAssumptions, projection.Target, instance.Target) {
		status, reason := Entails(SourceShapeFacts(instance.Before), required)
		add(status, "concrete ordinal/shape premise: "+reason, required)
	}
	post := invalidateSourceFacts(SourceShapeFacts(instance.Before), instance.Footprints)
	post = postFacts(post, SourceShapeFacts(instance.After))
	for _, promise := range rebasePredicates(declaration.Provides, projection.Target, instance.Target) {
		status, reason := Entails(post, promise)
		add(status, "concrete structural postcondition: "+reason, promise)
	}
	if declaration.Frame != nil {
		transformer := rebaseTransformer(ContractTransformer{Target: projection.Target, Frame: declaration.Frame}, instance.Target)
		result.Obligations = append(result.Obligations, CheckFrame(*transformer.Frame, instance.Footprints, FrameContext{})...)
	}
	result.Status = obligationStatus(result.Obligations)
	return result, nil
}
