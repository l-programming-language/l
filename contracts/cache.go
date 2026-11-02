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

type LogicalCacheKey struct {
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

type ImplementationCacheKey struct {
	Schema         string            `json:"schema"`
	Logical        string            `json:"logical"`
	Implementation string            `json:"implementation"`
	Target         string            `json:"target"`
	TrustPolicy    string            `json:"trust_policy"`
	Witnesses      string            `json:"witnesses"`
	Dependencies   map[string]string `json:"dependencies,omitempty"`
}

type cacheEntry struct {
	Schema             string          `json:"schema"`
	Key                string          `json:"key"`
	PayloadFingerprint string          `json:"payload_fingerprint"`
	Payload            json.RawMessage `json:"payload"`
}

type CacheCounters struct {
	Hits     uint64 `json:"hits"`
	Misses   uint64 `json:"misses"`
	Rejected uint64 `json:"rejected"`
}

type SessionCacheStats struct {
	Binding        CacheCounters `json:"binding"`
	Logical        CacheCounters `json:"logical"`
	Implementation CacheCounters `json:"implementation"`
}

// Session owns three independent in-memory caches. Stored values are immutable
// JSON snapshots; callers never receive aliases to an earlier analysis result.
// Keys describe all current inputs, including relevant candidate absence.
type Session struct {
	mu             sync.Mutex
	binding        map[string]cacheEntry
	logical        map[string]cacheEntry
	implementation map[string]cacheEntry
	stats          SessionCacheStats
}

func NewSession() *Session { return &Session{} }

func (session *Session) CacheStats() SessionCacheStats {
	session.mu.Lock()
	defer session.mu.Unlock()
	return session.stats
}

func (session *Session) StoreBinding(key BindingCacheKey, value any) error {
	if err := validateBindingCacheKey(key); err != nil {
		return err
	}
	return session.store("binding", key, value)
}

func (session *Session) LoadBinding(key BindingCacheKey, result any) (bool, error) {
	if err := validateBindingCacheKey(key); err != nil {
		return false, err
	}
	return session.load("binding", key, result)
}

func (session *Session) StoreLogical(key LogicalCacheKey, value any) error {
	if err := validateLogicalCacheKey(key); err != nil {
		return err
	}
	return session.store("logical", key, value)
}

func (session *Session) LoadLogical(key LogicalCacheKey, result any) (bool, error) {
	if err := validateLogicalCacheKey(key); err != nil {
		return false, err
	}
	return session.load("logical", key, result)
}

func (session *Session) StoreImplementation(key ImplementationCacheKey, value any) error {
	if err := validateImplementationCacheKey(key); err != nil {
		return err
	}
	return session.store("implementation", key, value)
}

func (session *Session) LoadImplementation(key ImplementationCacheKey, result any) (bool, error) {
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

func validateLogicalCacheKey(key LogicalCacheKey) error {
	if key.Schema != SchemaVersion || key.Meaning == "" || key.Specification == "" || key.Substitutions == "" || key.FacetSchemas == "" || key.RoleInterfaces == "" || key.Rules == "" || key.TargetModel == "" || !validCacheDependencies(key.Dependencies) {
		return fmt.Errorf("logical cache requires complete specification, substitutions, schemas, role interfaces, rules and semantic target dependencies")
	}
	return nil
}

func validateImplementationCacheKey(key ImplementationCacheKey) error {
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

func (session *Session) cache(kind string) (*map[string]cacheEntry, *CacheCounters) {
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
	digest, err := Fingerprint(json.RawMessage(payload))
	if err != nil {
		return err
	}
	session.mu.Lock()
	defer session.mu.Unlock()
	cache, _ := session.cache(kind)
	if *cache == nil {
		*cache = make(map[string]cacheEntry)
	}
	(*cache)[identity] = cacheEntry{Schema: SchemaVersion, Key: identity, PayloadFingerprint: digest, Payload: payload}
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
	digest, err := Fingerprint(entry.Payload)
	if err != nil || entry.Schema != SchemaVersion || entry.Key != identity || digest != entry.PayloadFingerprint {
		delete(*cache, identity)
		counters.Rejected++
		return false, nil
	}
	if err := json.Unmarshal(entry.Payload, result); err != nil {
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
	if found, err := session.LoadBinding(key, &result); err != nil || found {
		return result, err
	}
	result = SelectReference(candidates, context)
	if err := session.StoreBinding(key, result); err != nil {
		return ReferenceSelection{}, err
	}
	return result, nil
}

type LegacyContractRecord struct {
	Version        string                `json:"version"`
	Definition     DefinitionKey         `json:"definition"`
	Meaning        MeaningKey            `json:"meaning,omitempty"`
	Rule           string                `json:"rule,omitempty"`
	Target         TargetRef             `json:"target"`
	Environment    BindingEnvironmentKey `json:"environment,omitempty"`
	Specification  string                `json:"specification,omitempty"`
	Implementation string                `json:"implementation,omitempty"`
}

type LegacyMigration struct {
	Record                             LegacyContractRecord `json:"record"`
	BindingRevalidationRequired        bool                 `json:"binding_revalidation_required"`
	ImplementationRevalidationRequired bool                 `json:"implementation_revalidation_required"`
	Missing                            []string             `json:"missing,omitempty"`
	Status                             Status               `json:"status"`
	Reason                             string               `json:"reason"`
}

// MigrateLegacyRecord preserves available data but grants no selection or
// conformance evidence. Old records must pass current resolvers/checkers before
// being admitted to a v1.1 cache; absent inputs are never fabricated.
func MigrateLegacyRecord(record LegacyContractRecord) (LegacyMigration, error) {
	if record.Version != "1.0" {
		return LegacyMigration{}, fmt.Errorf("unsupported legacy contracts version %q", record.Version)
	}
	result := LegacyMigration{Record: record, BindingRevalidationRequired: true, ImplementationRevalidationRequired: true, Status: Unresolved, Reason: "legacy record retained as declaration data; current binding and concrete assurance must be recomputed"}
	for name, absent := range map[string]bool{"definition": record.Definition == "", "meaning": record.Meaning == "", "rule": record.Rule == "", "target": record.Target.Root == "" || record.Target.Unknown, "environment": record.Environment == "", "specification": record.Specification == "", "implementation": record.Implementation == ""} {
		if absent {
			result.Missing = append(result.Missing, name)
		}
	}
	result.Missing = evidenceUnique(result.Missing)
	return result, nil
}
