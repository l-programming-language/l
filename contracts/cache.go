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
	mu                               sync.Mutex
	binding, logical, implementation sessionCache
}

type sessionCache struct {
	entries map[string]json.RawMessage
	counts  CacheCounters
}

type sessionKey interface {
	selectCache(*Session) (*sessionCache, error)
}

func NewSession() *Session { return &Session{} }

func (session *Session) CacheStats() SessionCacheStats {
	session.mu.Lock()
	defer session.mu.Unlock()
	return SessionCacheStats{Binding: session.binding.counts, Logical: session.logical.counts, Implementation: session.implementation.counts}
}

// RevalidateBinding records a freshly resolved binding without supplying any
// resolver choices. Repeated observations under the same complete key must agree.
func (session *Session) RevalidateBinding(key BindingCacheKey, current ResolvedBinding) error {
	var cached ResolvedBinding
	found, err := session.load(key, &cached)
	if err != nil {
		return err
	}
	if found {
		if StableID("binding", cached) != StableID("binding", current) {
			return fmt.Errorf("binding cache disagrees with current Track A resolution at %s", current.Occurrence)
		}
		return nil
	}
	return session.store(key, current)
}

func (options AnalysisOptions) logical() AnalysisOptions {
	options.NativeApplications = nil
	options.NativeSummaries = nil
	options.NativeWitnesses = nil
	options.ApprovedAssumptions = nil
	return options
}

// Infer caches the conditional typed-plan analysis separately from assurance
// about the selected native implementations, target inputs and trust witnesses.
func (session *Session) Infer(plan Plan, options AnalysisOptions, inputs map[string]string, lowering LoweringInventory) (DerivedSummary, error) {
	logicalOptions := options.logical()
	logical := logicalCacheKey{
		Schema: SchemaVersion, Meaning: MeaningKey(StableID("analysis-entry", plan.Entry)),
		Specification: StableID("typed-plan", plan), Substitutions: StableID("analysis-facts", logicalOptions),
		FacetSchemas: "source-facets:not-applied", RoleInterfaces: "source-roles:not-applied", Rules: inferencePolicy,
		TargetModel: "l-typed-plan/1", Dependencies: map[string]string{"native-specifications": StableID("specifications", options.NativeSummaries)},
	}
	// Native declarations and approvals cannot establish this reusable baseline.
	baseline, err := cachedResult(session, logical, func() (DerivedSummary, error) {
		return Infer(plan, logicalOptions)
	})
	if err != nil {
		return DerivedSummary{}, err
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
	return cachedResult(session, implementation, func() (DerivedSummary, error) {
		if options.NativeApplications == nil && options.NativeSummaries == nil && options.NativeWitnesses == nil && options.ApprovedAssumptions == nil {
			return baseline, nil
		}
		return Infer(plan, options)
	})
}

func (key BindingCacheKey) selectCache(session *Session) (*sessionCache, error) {
	if key.Schema != SchemaVersion || key.SourceStructure == "" || key.Environment == "" || key.Candidates == "" || key.Applicability == "" || key.Target.Root == "" || key.Target.Unknown || key.LexicalRole == "" || key.PhrasePolicy == "" || key.ResolutionPolicy == "" {
		return nil, fmt.Errorf("binding cache requires complete current source, candidate/environment, applicability, target, phrase and policy dependencies")
	}
	return &session.binding, nil
}

func (key logicalCacheKey) selectCache(session *Session) (*sessionCache, error) {
	if key.Schema != SchemaVersion || key.Meaning == "" || key.Specification == "" || key.Substitutions == "" || key.FacetSchemas == "" || key.RoleInterfaces == "" || key.Rules == "" || key.TargetModel == "" || !validFingerprintDependencies(key.Dependencies) {
		return nil, fmt.Errorf("logical cache requires complete specification, substitutions, schemas, role interfaces, rules and semantic target dependencies")
	}
	return &session.logical, nil
}

func (key implementationCacheKey) selectCache(session *Session) (*sessionCache, error) {
	if key.Schema != SchemaVersion || key.Logical == "" || key.Implementation == "" || key.Target == "" || key.TrustPolicy == "" || key.Witnesses == "" || !validFingerprintDependencies(key.Dependencies) {
		return nil, fmt.Errorf("implementation cache requires current logical, body, target, trust, witness and transitive dependency fingerprints")
	}
	return &session.implementation, nil
}

func validFingerprintDependencies(dependencies map[string]string) bool {
	for key, value := range dependencies {
		if key == "" || value == "" {
			return false
		}
	}
	return true
}

// A cache hit returns a detached snapshot. Compute misses outside the lock and
// publish only successful results, using the domain chosen by the typed key.
func cachedResult[T any](session *Session, key sessionKey, compute func() (T, error)) (T, error) {
	var value T
	if found, err := session.load(key, &value); err != nil || found {
		return value, err
	}
	value, err := compute()
	if err == nil {
		err = session.store(key, value)
	}
	return value, err
}

func (session *Session) store(key sessionKey, value any) error {
	if session == nil {
		return fmt.Errorf("contract cache session is nil")
	}
	cache, err := key.selectCache(session)
	if err != nil {
		return err
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
	if cache.entries == nil {
		cache.entries = make(map[string]json.RawMessage)
	}
	cache.entries[identity] = payload
	return nil
}

func (session *Session) load(key sessionKey, result any) (bool, error) {
	if session == nil {
		return false, fmt.Errorf("contract cache session is nil")
	}
	cache, err := key.selectCache(session)
	if err != nil {
		return false, err
	}
	identity, err := Fingerprint(key)
	if err != nil {
		return false, err
	}
	session.mu.Lock()
	defer session.mu.Unlock()
	entry, exists := cache.entries[identity]
	if !exists {
		cache.counts.Misses++
		return false, nil
	}
	if err := json.Unmarshal(entry, result); err != nil {
		return false, fmt.Errorf("decode contract cache: %w", err)
	}
	cache.counts.Hits++
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
	return cachedResult(session, key, func() (ReferenceSelection, error) {
		return SelectReference(candidates, context), nil
	})
}
