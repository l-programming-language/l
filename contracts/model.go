// Package contracts defines the compiler-independent L Contracts v1.1 model.
// Declarations are inputs to checking, never certificates of their own validity.
package contracts

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
)

const SchemaVersion = "1.1"
const AnalyzerVersion = "l-contracts-1.1/3"
const TrackAPolicy = "l-existing-resolution/1"
const ReferencePolicy = "unique-applicable/1"

type DefinitionKey string
type MeaningKey string
type BindingOccurrenceID string
type InstanceKey string
type BindingEnvironmentKey string
type Coverage string
type Status string
type Outcome string
type Phase string
type EffectKind string

const (
	Known           Coverage = "known"
	Partial         Coverage = "partial"
	Unknown         Coverage = "unknown"
	NotApplicable   Coverage = "not_applicable"
	Established     Status   = "established"
	Contradicted    Status   = "contradicted"
	Unresolved      Status   = "unresolved"
	Normal          Outcome  = "normal"
	ReturnedFailure Outcome  = "returned_failure"
	GuardRejected   Outcome  = "guard_rejected"
	Panic           Outcome  = "panic"
	Abort           Outcome  = "abort"
	Cancelled       Outcome  = "cancelled"
	ProcessFailure  Outcome  = "process_failure"
	Elaboration     Phase    = "elaboration"
	Construction    Phase    = "construction"
	Invocation      Phase    = "invocation"
	Destruction     Phase    = "destruction"
	Cleanup         Phase    = "cleanup"
)

type Origin struct {
	Source    string              `json:"source"`
	Start     int                 `json:"start"`
	End       int                 `json:"end"`
	Parent    BindingOccurrenceID `json:"parent,omitempty"`
	Expansion []int               `json:"expansion,omitempty"`
	Label     string              `json:"label,omitempty"`
}

type DefinitionIdentity struct {
	Source      string   `json:"source"`
	Scope       string   `json:"scope"`
	Owner       []string `json:"owner,omitempty"`
	Declaration Origin   `json:"declaration"`
	Role        string   `json:"role"`
}

type MeaningIdentity struct {
	Definition    DefinitionKey `json:"definition"`
	Rule          Origin        `json:"rule"`
	Applicability string        `json:"applicability"`
	Version       string        `json:"version"`
}

type TargetRef struct {
	Root       string        `json:"root"`
	Path       []string      `json:"path,omitempty"`
	Type       string        `json:"type,omitempty"`
	Nominal    DefinitionKey `json:"nominal,omitempty"`
	Phase      Phase         `json:"phase,omitempty"`
	Generation string        `json:"generation,omitempty"`
	Unknown    bool          `json:"unknown,omitempty"`
}

func (t TargetRef) String() string {
	if t.Unknown {
		return "unresolved attachment"
	}
	if len(t.Path) == 0 {
		return t.Root
	}
	return t.Root + "/" + strings.Join(t.Path, "/")
}

type ResourceRef struct {
	ID      string `json:"id"`
	Kind    string `json:"kind,omitempty"`
	Unknown bool   `json:"unknown,omitempty"`
}

type Atom struct {
	Kind   string `json:"kind"`
	Text   string `json:"text,omitempty"`
	Number int64  `json:"number,omitempty"`
	Bool   bool   `json:"bool,omitempty"`
}

type Predicate struct {
	Kind     string      `json:"kind"`
	Subject  string      `json:"subject,omitempty"`
	Value    Atom        `json:"value,omitempty"`
	Args     []Predicate `json:"args,omitempty"`
	Target   TargetRef   `json:"target,omitempty"`
	Resource ResourceRef `json:"resource,omitempty"`
	Lower    *int64      `json:"lower,omitempty"`
	Upper    *int64      `json:"upper,omitempty"`
}

type Effect struct {
	Kind      EffectKind  `json:"kind"`
	Resource  ResourceRef `json:"resource"`
	Target    TargetRef   `json:"target"`
	Phase     Phase       `json:"phase"`
	Condition string      `json:"condition,omitempty"`
	Definite  bool        `json:"definite"`
	Origin    Origin      `json:"origin"`
}

type Claim struct {
	ID          string    `json:"id"`
	Predicate   Predicate `json:"predicate"`
	Target      TargetRef `json:"target"`
	Scope       string    `json:"scope"`
	Phase       Phase     `json:"phase"`
	Outcomes    []Outcome `json:"outcomes"`
	Assumptions []string  `json:"assumptions,omitempty"`
}

type Obligation struct {
	ID           string   `json:"id"`
	Claim        Claim    `json:"claim"`
	Status       Status   `json:"status"`
	Reason       string   `json:"reason"`
	Origin       Origin   `json:"origin"`
	Evidence     []string `json:"evidence,omitempty"`
	Dependencies []string `json:"dependencies,omitempty"`
}

type Profile struct {
	Name     string    `json:"name"`
	Version  string    `json:"version"`
	Scope    string    `json:"scope"`
	Target   TargetRef `json:"target"`
	Outcomes []Outcome `json:"outcomes,omitempty"`
}

type ResolvedBinding struct {
	Occurrence            BindingOccurrenceID   `json:"occurrence"`
	Definition            DefinitionKey         `json:"definition"`
	Meaning               MeaningKey            `json:"meaning"`
	Spelling              string                `json:"spelling"`
	CanonicalName         string                `json:"canonical_name"`
	LexicalRole           string                `json:"lexical_role"`
	Target                TargetRef             `json:"target"`
	Origin                Origin                `json:"origin"`
	DefinitionOrigin      Origin                `json:"definition_origin"`
	DefinitionEnvironment BindingEnvironmentKey `json:"definition_environment"`
	UseEnvironment        BindingEnvironmentKey `json:"use_environment"`
	Policy                string                `json:"policy"`
	SelectionReason       string                `json:"selection_reason"`
	CandidateFingerprint  string                `json:"candidate_fingerprint"`
	Substitutions         map[string]string     `json:"substitutions,omitempty"`
	Dependencies          []string              `json:"dependencies,omitempty"`
}

// DefinitionRecord exposes canonical declarations independently of occurrence
// discovery. It carries no assertion that a declaration has been verified.
type DefinitionRecord struct {
	Definition    DefinitionKey         `json:"definition"`
	Meaning       MeaningKey            `json:"meaning"`
	Name          string                `json:"name"`
	Role          string                `json:"role"`
	Origin        Origin                `json:"origin"`
	Environment   BindingEnvironmentKey `json:"environment"`
	Applicability string                `json:"applicability,omitempty"`
	Descriptor    bool                  `json:"descriptor"`
	Native        bool                  `json:"native"`
}

// Fingerprint uses JSON's sorted map keys. Callers preserve semantic sequence
// order and canonicalize only collections whose model explicitly denotes sets.
func Fingerprint(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("contract fingerprint: %w", err)
	}
	return fmt.Sprintf("%x", sha256.Sum256(data)), nil
}

// StableID accepts only serializable model values. An invalid value returns an
// empty identity; validators reject empty identities rather than reusing them.
func StableID(prefix string, value any) string {
	hash, err := Fingerprint(value)
	if err != nil {
		return ""
	}
	return prefix + ":" + hash
}

func AllEstablished(obligations []Obligation) bool {
	for _, obligation := range obligations {
		if obligation.Status != Established {
			return false
		}
	}
	return true
}
