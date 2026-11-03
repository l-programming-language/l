package contracts

import (
	"encoding/json"
	"fmt"
	"sync"
)

type BindingCacheKey struct {
	Schema           string                `json:"schema"`
	SourceStructure  string                `json:"source_structure"`
	Environment      BindingEnvironmentKey `json:"environment"`
	Candidates       string                `json:"candidates"`
	Applicability    string                `json:"applicability"`
	Target           TargetRef             `json:"target"`
	LexicalRole      string                `json:"lexical_role"`
	PhrasePolicy     string                `json:"phrase_policy"`
	ResolutionPolicy string                `json:"resolution_policy"`
}

type logicalCacheKey struct {
	Schema         string            `json:"schema"`
	Meaning        MeaningKey        `json:"meaning"`
	Specification  string            `json:"specification"`
	Substitutions  string            `json:"substitutions"`
	FacetSchemas   string            `json:"facet_schemas"`
	RoleInterfaces string            `json:"role_interfaces"`
	Rules          string            `json:"rules"`
	TargetModel    string            `json:"target_model"`
	Dependencies   map[string]string `json:"dependencies,omitempty"`
}

type implementationCacheKey struct {
	Schema         string            `json:"schema"`
	Logical        string            `json:"logical"`
	Implementation string            `json:"implementation"`
	Target         string            `json:"target"`
	TrustPolicy    string            `json:"trust_policy"`
	Witnesses      string            `json:"witnesses"`
	Dependencies   map[string]string `json:"dependencies,omitempty"`
}

type CacheCounters struct {
	Hits   uint64 `json:"hits"`
	Misses uint64 `json:"misses"`
}

type SessionCacheStats struct {
	Binding        CacheCounters `json:"binding"`
	Logical        CacheCounters `json:"logical"`
	Implementation CacheCounters `json:"implementation"`
}

// Session owns three independent in-memory caches. Stored values are immutable
// JSON snapshots; callers never receive aliases to an earlier analysis result.
// Logical and implementation keys and writes are private to the checkers that
// compute their results. Binding keys describe the freshly resolved environment,
// including relevant candidate absence; no cache is imported or persisted.
type Session struct {
	mu             sync.Mutex
	binding        map[string]json.RawMessage
	logical        map[string]json.RawMessage
	implementation map[string]json.RawMessage
	stats          SessionCacheStats
}

func NewSession() *Session { return &Session{} }

func (session *Session) CacheStats() SessionCacheStats {
	session.mu.Lock()
	defer session.mu.Unlock()
	return session.stats
}

// RevalidateBinding records a freshly resolved binding without supplying any
// resolver choices. Repeated observations under the same complete key must agree.
func (session *Session) RevalidateBinding(key BindingCacheKey, current ResolvedBinding) error {
	var cached ResolvedBinding
	found, err := session.loadBinding(key, &cached)
	if err != nil {
		return err
	}
	if found {
		if StableID("binding", cached) != StableID("binding", current) {
			return fmt.Errorf("binding cache disagrees with current Track A resolution at %s", current.Occurrence)
		}
		return nil
	}
	return session.storeBinding(key, current)
}

// Infer caches the conditional typed-plan analysis separately from assurance
// about the selected native implementations, target inputs and trust witnesses.
func (session *Session) Infer(plan Plan, options AnalysisOptions, inputs map[string]string, lowering LoweringInventory) (DerivedSummary, error) {
	logicalOptions := options
	logicalOptions.NativeSummaries = nil
	logicalOptions.NativeApplications = nil
	logicalOptions.NativeWitnesses = nil
	logicalOptions.ApprovedAssumptions = nil
	logical := logicalCacheKey{
		Schema: SchemaVersion, Meaning: MeaningKey(StableID("analysis-entry", plan.Entry)),
		Specification: StableID("typed-plan", plan), Substitutions: StableID("analysis-facts", logicalOptions),
		FacetSchemas: "source-facets:not-applied", RoleInterfaces: "source-roles:not-applied", Rules: AnalyzerVersion,
		TargetModel: "l-typed-plan/1", Dependencies: map[string]string{"native-specifications": StableID("specifications", options.NativeSummaries)},
	}
	// Native declarations and approvals cannot establish this reusable baseline.
	var baseline DerivedSummary
	found, err := session.loadLogical(logical, &baseline)
	if err != nil {
		return DerivedSummary{}, err
	}
	if !found {
		baseline, err = Infer(plan, logicalOptions)
		if err != nil {
			return DerivedSummary{}, err
		}
		if err := session.storeLogical(logical, baseline); err != nil {
			return DerivedSummary{}, err
		}
	}
	dependencies := make(map[string]string, len(inputs)+len(lowering.Fingerprints)+1)
	for name, fingerprint := range inputs {
		dependencies["input:"+name] = fingerprint
	}
	for name, fingerprint := range lowering.Fingerprints {
		dependencies["implementation:"+name] = fingerprint
	}
	dependencies["native-applications"] = StableID("native-applications", options.NativeApplications)
	implementation := implementationCacheKey{
		Schema: SchemaVersion, Logical: StableID("logical-key", logical),
		Implementation: StableID("lowering", lowering), Target: StableID("selected-target-inputs", inputs),
		TrustPolicy: StableID("approved-assumptions", options.ApprovedAssumptions), Witnesses: StableID("witnesses", options.NativeWitnesses), Dependencies: dependencies,
	}
	var summary DerivedSummary
	if found, err := session.loadImplementation(implementation, &summary); err != nil || found {
		return summary, err
	}
	if options.NativeApplications == nil && options.NativeSummaries == nil && options.NativeWitnesses == nil && options.ApprovedAssumptions == nil {
		summary = baseline
	} else {
		summary, err = Infer(plan, options)
		if err != nil {
			return DerivedSummary{}, err
		}
	}
	if err := session.storeImplementation(implementation, summary); err != nil {
		return DerivedSummary{}, err
	}
	return summary, nil
}

func (session *Session) storeBinding(key BindingCacheKey, value any) error {
	if err := validateBindingCacheKey(key); err != nil {
		return err
	}
	return session.store("binding", key, value)
}

func (session *Session) loadBinding(key BindingCacheKey, result any) (bool, error) {
	if err := validateBindingCacheKey(key); err != nil {
		return false, err
	}
	return session.load("binding", key, result)
}

func (session *Session) storeLogical(key logicalCacheKey, value any) error {
	if err := validateLogicalCacheKey(key); err != nil {
		return err
	}
	return session.store("logical", key, value)
}

func (session *Session) loadLogical(key logicalCacheKey, result any) (bool, error) {
	if err := validateLogicalCacheKey(key); err != nil {
		return false, err
	}
	return session.load("logical", key, result)
}

func (session *Session) storeImplementation(key implementationCacheKey, value any) error {
	if err := validateImplementationCacheKey(key); err != nil {
		return err
	}
	return session.store("implementation", key, value)
}

func (session *Session) loadImplementation(key implementationCacheKey, result any) (bool, error) {
	if err := validateImplementationCacheKey(key); err != nil {
		return false, err
	}
	return session.load("implementation", key, result)
}

func validateBindingCacheKey(key BindingCacheKey) error {
	if key.Schema != SchemaVersion || key.SourceStructure == "" || key.Environment == "" || key.Candidates == "" || key.Applicability == "" || key.Target.Root == "" || key.Target.Unknown || key.LexicalRole == "" || key.PhrasePolicy == "" || key.ResolutionPolicy == "" {
		return fmt.Errorf("binding cache requires complete current source, candidate/environment, applicability, target, phrase and policy dependencies")
	}
	return nil
}

func validateLogicalCacheKey(key logicalCacheKey) error {
	if key.Schema != SchemaVersion || key.Meaning == "" || key.Specification == "" || key.Substitutions == "" || key.FacetSchemas == "" || key.RoleInterfaces == "" || key.Rules == "" || key.TargetModel == "" || !validCacheDependencies(key.Dependencies) {
		return fmt.Errorf("logical cache requires complete specification, substitutions, schemas, role interfaces, rules and semantic target dependencies")
	}
	return nil
}

func validateImplementationCacheKey(key implementationCacheKey) error {
	if key.Schema != SchemaVersion || key.Logical == "" || key.Implementation == "" || key.Target == "" || key.TrustPolicy == "" || key.Witnesses == "" || !validCacheDependencies(key.Dependencies) {
		return fmt.Errorf("implementation cache requires current logical, body, target, trust, witness and transitive dependency fingerprints")
	}
	return nil
}

func validCacheDependencies(dependencies map[string]string) bool {
	for key, value := range dependencies {
		if key == "" || value == "" {
			return false
		}
	}
	return true
}

func (session *Session) cache(kind string) (*map[string]json.RawMessage, *CacheCounters) {
	switch kind {
	case "binding":
		return &session.binding, &session.stats.Binding
	case "logical":
		return &session.logical, &session.stats.Logical
	default:
		return &session.implementation, &session.stats.Implementation
	}
}

func (session *Session) store(kind string, key, value any) error {
	if session == nil {
		return fmt.Errorf("contract cache session is nil")
	}
	identity, err := Fingerprint(key)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("serialize contract cache: %w", err)
	}
	session.mu.Lock()
	defer session.mu.Unlock()
	cache, _ := session.cache(kind)
	if *cache == nil {
		*cache = make(map[string]json.RawMessage)
	}
	(*cache)[identity] = payload
	return nil
}

func (session *Session) load(kind string, key, result any) (bool, error) {
	if session == nil {
		return false, fmt.Errorf("contract cache session is nil")
	}
	identity, err := Fingerprint(key)
	if err != nil {
		return false, err
	}
	session.mu.Lock()
	defer session.mu.Unlock()
	cache, counters := session.cache(kind)
	entry, exists := (*cache)[identity]
	if !exists {
		counters.Misses++
		return false, nil
	}
	if err := json.Unmarshal(entry, result); err != nil {
		return false, fmt.Errorf("decode contract cache: %w", err)
	}
	counters.Hits++
	return true, nil
}

// SelectReference fingerprints the entire supplied candidate environment, not
// only the selected meaning. Newly matching imports and changed absence facts
// therefore force the same selection as a cold call to SelectReference.
func (session *Session) SelectReference(candidates []Candidate, context SelectionContext) (ReferenceSelection, error) {
	if context.Policy == "" {
		context.Policy = ReferencePolicy
	}
	candidateFingerprint, err := Fingerprint(candidates)
	if err != nil {
		return ReferenceSelection{}, err
	}
	contextFingerprint, err := Fingerprint(context)
	if err != nil {
		return ReferenceSelection{}, err
	}
	key := BindingCacheKey{Schema: SchemaVersion, SourceStructure: contextFingerprint, Environment: context.Environment, Candidates: candidateFingerprint, Applicability: contextFingerprint, Target: context.Target, LexicalRole: context.LexicalRole, PhrasePolicy: "typed-fixture/1", ResolutionPolicy: context.Policy}
	var result ReferenceSelection
	if found, err := session.loadBinding(key, &result); err != nil || found {
		return result, err
	}
	result = SelectReference(candidates, context)
	if err := session.storeBinding(key, result); err != nil {
		return ReferenceSelection{}, err
	}
	return result, nil
}
