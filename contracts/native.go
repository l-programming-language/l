package contracts

// NativeType is a compiler-projected type identity. Atom is an already resolved
// native type atom, never a native body or an inferred semantic description.
type NativeType struct {
	Kind      string        `json:"kind"`
	Nominal   DefinitionKey `json:"nominal,omitempty"`
	Atom      string        `json:"atom,omitempty"`
	Arguments []NativeType  `json:"arguments,omitempty"`
}

type NativeUse struct {
	ID                 string                 `json:"id"`
	Meaning            MeaningKey             `json:"meaning"`
	Name               string                 `json:"name"`
	Occurrence         BindingOccurrenceID    `json:"occurrence,omitempty"`
	Target             TargetRef              `json:"target"`
	Origin             Origin                 `json:"origin"`
	Phase              Phase                  `json:"phase"`
	Boundary           string                 `json:"boundary,omitempty"`
	Signature          *NativeSignature       `json:"signature,omitempty"`
	Implementation     string                 `json:"implementation"`
	Dependencies       map[string]string      `json:"dependencies"`
	Resources          map[string]ResourceRef `json:"resources,omitempty"`
	Values             map[string]Atom        `json:"values,omitempty"`
	Substitutions      Substitutions          `json:"substitutions,omitempty"`
	OutcomeDisposition string                 `json:"outcome_disposition"`
}

type AnalysisScope struct {
	ID         string              `json:"id"`
	NodeID     string              `json:"node_id"`
	Occurrence BindingOccurrenceID `json:"occurrence,omitempty"`
	Meaning    MeaningKey          `json:"meaning"`
	Target     TargetRef           `json:"target"`
	Origin     Origin              `json:"origin"`
	Boundary   string              `json:"boundary"`
	Phase      Phase               `json:"phase"`
	NativeUses []string            `json:"native_uses,omitempty"`
	Resources  []FailureLocation   `json:"resources,omitempty"`
	Includes   []string            `json:"includes"`
	Exclusions []string            `json:"exclusions,omitempty"`
	Complete   bool                `json:"complete"`
	Unknowns   []UnknownReason     `json:"unknowns,omitempty"`
}

// NativeOutcomeContract describes a proposition to approve, not evidence.
// State relations describe observations at completion, after owned cleanup.
type NativeOutcomeContract struct {
	Outcome       Outcome            `json:"outcome"`
	Effects       *DeclaredEffectRow `json:"effects,omitempty"`
	AbsentEffects []EffectKind       `json:"absent_effects,omitempty"`
	States        []StateRelation    `json:"states,omitempty"`
	Costs         *ResourceCost      `json:"costs,omitempty"`
}

type StateRelation struct {
	Location    FailureLocation `json:"location"`
	Observation string          `json:"observation"`
	Relation    string          `json:"relation"`
	Origin      Origin          `json:"origin"`
	Evidence    []string        `json:"evidence,omitempty"`
}

// NativeEvaluation separates conditional native approvals from actual entry
// applicability. Inference and reporting consume this same per-use result.
type NativeEvaluation struct {
	Use                NativeUse
	Requirements       []Predicate
	Applicability      Obligation
	Entry              Claim
	Components         []NativeComponent
	Guarantees         []NativeComponent
	Summary            DerivedSummary
	Evidence           []string
	ContextFingerprint string
}
