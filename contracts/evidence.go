package contracts

import (
	"fmt"
	"maps"
	"slices"
)

// TrustPolicy is selected explicitly by the caller. A native manifest is a
// declaration and cannot install a policy or approve one of its own witnesses.
type TrustPolicy struct {
	Version           string   `json:"version"`
	ID                string   `json:"id"`
	Authorities       []string `json:"authorities,omitempty"`
	ApprovedWitnesses []string `json:"approved_witnesses,omitempty"`
	AllowTestOnly     bool     `json:"allow_test_only,omitempty"`
}

type NativeWitness struct {
	ID                        string            `json:"id"`
	ClaimFingerprint          string            `json:"claim_fingerprint"`
	ContractFingerprint       string            `json:"contract_fingerprint"`
	ImplementationFingerprint string            `json:"implementation_fingerprint"`
	TargetFingerprint         string            `json:"target_fingerprint"`
	Authority                 string            `json:"authority"`
	Rationale                 string            `json:"rationale"`
	Dependencies              map[string]string `json:"dependencies,omitempty"`
	TestOnly                  bool              `json:"test_only,omitempty"`
}

type EvidenceFingerprints struct {
	Contract       string            `json:"contract"`
	Implementation string            `json:"implementation"`
	Target         string            `json:"target"`
	Dependencies   map[string]string `json:"dependencies,omitempty"`
}

type EvidenceResult struct {
	Status       Status   `json:"status"`
	Reason       string   `json:"reason"`
	Assumptions  []string `json:"assumptions,omitempty"`
	Dependencies []string `json:"dependencies,omitempty"`
	Evidence     []string `json:"evidence,omitempty"`
}

// ClaimFingerprint excludes the report-local label, but includes the actual
// proposition, target, outcomes, phase, observation boundary and assumptions.
func ClaimFingerprint(claim Claim) string {
	claim.ID = ""
	value, err := Fingerprint(claim)
	if err != nil {
		return ""
	}
	return value
}

func ApproveNativeClaim(policy TrustPolicy, witness NativeWitness, claim Claim, current EvidenceFingerprints) EvidenceResult {
	unknown := func(reason string) EvidenceResult { return EvidenceResult{Status: Unresolved, Reason: reason} }
	if policy.Version != SchemaVersion || policy.ID == "" {
		return unknown("no explicitly selected supported trust policy")
	}
	if witness.ID == "" || !containsString(policy.ApprovedWitnesses, witness.ID) || witness.Authority == "" || !containsString(policy.Authorities, witness.Authority) {
		return unknown("native witness is not approved by the selected trust policy")
	}
	if witness.TestOnly && !policy.AllowTestOnly {
		return unknown("test-only native assumption is excluded by production policy")
	}
	if witness.Rationale == "" || witness.ClaimFingerprint == "" || witness.ClaimFingerprint != ClaimFingerprint(claim) {
		return unknown("native witness does not establish this exact proposition and scope")
	}
	wanted := EvidenceFingerprints{Contract: witness.ContractFingerprint, Implementation: witness.ImplementationFingerprint, Target: witness.TargetFingerprint, Dependencies: witness.Dependencies}
	if !validEvidenceFingerprints(wanted) || !sameEvidenceFingerprints(wanted, current) {
		return unknown("native witness implementation, contract, target or transitive dependency fingerprint is stale or incomplete")
	}
	dependencies := []string{"contract:" + current.Contract, "implementation:" + current.Implementation, "target:" + current.Target, "policy:" + policy.ID}
	for name, fingerprint := range current.Dependencies {
		dependencies = append(dependencies, name+":"+fingerprint)
	}
	assumptions := evidenceUnique(append(append([]string(nil), claim.Assumptions...), witness.ID))
	reason := fmt.Sprintf("established conditional on explicitly approved native assumption and declared premises %v", assumptions)
	if witness.TestOnly {
		reason = fmt.Sprintf("established only under explicitly enabled test-only assumptions %v; no production approval", assumptions)
	}
	return EvidenceResult{Status: Established, Reason: reason, Assumptions: assumptions, Dependencies: evidenceUnique(dependencies), Evidence: []string{witness.ID}}
}

func validEvidenceFingerprints(value EvidenceFingerprints) bool {
	return value.Contract != "" && value.Implementation != "" && value.Target != "" && validFingerprintDependencies(value.Dependencies)
}

func sameEvidenceFingerprints(left, right EvidenceFingerprints) bool {
	return left.Contract == right.Contract && left.Implementation == right.Implementation && left.Target == right.Target && maps.Equal(left.Dependencies, right.Dependencies)
}

type EvidenceNode struct {
	ID           string               `json:"id"`
	Claim        string               `json:"claim"`
	Kind         string               `json:"kind"`
	Rule         string               `json:"rule,omitempty"`
	Premises     []string             `json:"premises,omitempty"`
	Witness      string               `json:"witness,omitempty"`
	Guard        string               `json:"guard,omitempty"`
	Fingerprints EvidenceFingerprints `json:"fingerprints"`
	Origin       Origin               `json:"origin"`
}

// CheckedEvidence is produced by a supported checker, not decoded from native
// declarations. It binds a checker result to one proposition and current inputs.
type CheckedEvidence struct {
	ClaimFingerprint string               `json:"claim_fingerprint"`
	Checker          string               `json:"checker"`
	Fingerprints     EvidenceFingerprints `json:"fingerprints"`
}

type GuardWitness struct {
	ID               string            `json:"id"`
	ClaimFingerprint string            `json:"claim_fingerprint"`
	Dominates        bool              `json:"dominates"`
	ResourceVersions map[string]string `json:"resource_versions,omitempty"`
	Outcomes         []Outcome         `json:"outcomes"`
	Origin           Origin            `json:"origin"`
}

// EvidenceContext contains checker outputs and current runtime-guard facts.
// Loading a manifest never populates Checked, Refutations or Observations.
type EvidenceContext struct {
	Policy            TrustPolicy
	NativeWitnesses   map[string]NativeWitness
	Current           map[string]EvidenceFingerprints
	Checked           map[string]CheckedEvidence
	Refutations       map[string]CheckedEvidence
	Observations      map[string]CheckedEvidence
	Guards            map[string]GuardWitness
	ResourceVersions  map[string]string
	InvalidatedGuards map[string]bool
}

type EvidenceGraph struct {
	Claims []Claim        `json:"claims"`
	Nodes  []EvidenceNode `json:"nodes"`
}

func (graph EvidenceGraph) Evaluate(nodeID string, context EvidenceContext) EvidenceResult {
	claims := make(map[string]Claim, len(graph.Claims))
	nodes := make(map[string]EvidenceNode, len(graph.Nodes))
	for _, claim := range graph.Claims {
		if claim.ID == "" {
			return EvidenceResult{Status: Unresolved, Reason: "evidence graph contains an empty claim identity"}
		}
		if _, exists := claims[claim.ID]; exists {
			return EvidenceResult{Status: Unresolved, Reason: "evidence graph contains duplicate claim identity " + claim.ID}
		}
		claims[claim.ID] = claim
	}
	for _, node := range graph.Nodes {
		if node.ID == "" {
			return EvidenceResult{Status: Unresolved, Reason: "evidence graph contains an empty node identity"}
		}
		if _, exists := nodes[node.ID]; exists {
			return EvidenceResult{Status: Unresolved, Reason: "evidence graph contains duplicate node identity " + node.ID}
		}
		nodes[node.ID] = node
	}
	active := make(map[string]bool)
	memo := make(map[string]EvidenceResult)
	var visit func(string) EvidenceResult
	visit = func(id string) EvidenceResult {
		if result, exists := memo[id]; exists {
			return result
		}
		if active[id] {
			return EvidenceResult{Status: Unresolved, Reason: "circular evidence requires a supported induction rule"}
		}
		node, exists := nodes[id]
		if !exists {
			return EvidenceResult{Status: Unresolved, Reason: "missing evidence node " + id}
		}
		claim, exists := claims[node.Claim]
		if !exists {
			return EvidenceResult{Status: Unresolved, Reason: "missing proposition for evidence " + id}
		}
		active[id] = true
		premises := make([]EvidenceResult, 0, len(node.Premises))
		for _, premise := range node.Premises {
			premises = append(premises, visit(premise))
		}
		result := evaluateEvidenceNode(node, claim, nodes, claims, premises, context)
		delete(active, id)
		if result.Status == Established {
			result.Assumptions = append(result.Assumptions, claim.Assumptions...)
			for _, premise := range premises {
				result.Assumptions = append(result.Assumptions, premise.Assumptions...)
				result.Dependencies = append(result.Dependencies, premise.Dependencies...)
				result.Evidence = append(result.Evidence, premise.Evidence...)
			}
		}
		result.Assumptions = evidenceUnique(result.Assumptions)
		result.Dependencies = evidenceUnique(result.Dependencies)
		result.Evidence = evidenceUnique(append(result.Evidence, id))
		memo[id] = result
		return result
	}
	return visit(nodeID)
}

func evaluateEvidenceNode(node EvidenceNode, claim Claim, nodes map[string]EvidenceNode, claims map[string]Claim, premises []EvidenceResult, context EvidenceContext) EvidenceResult {
	unresolved := func(reason string) EvidenceResult { return EvidenceResult{Status: Unresolved, Reason: reason} }
	for _, premise := range premises {
		if premise.Status != Established {
			return unresolved("required premise is not established: " + premise.Reason)
		}
	}
	switch node.Kind {
	case "native_assumption":
		witness, exists := context.NativeWitnesses[node.Witness]
		if !exists {
			return unresolved("missing native approval witness")
		}
		return ApproveNativeClaim(context.Policy, witness, claim, context.Current[node.ID])
	case "runtime_guard":
		guard, exists := context.Guards[node.Guard]
		if !exists || guard.ID != node.Guard || !guard.Dominates || context.InvalidatedGuards[node.Guard] || guard.ClaimFingerprint != ClaimFingerprint(claim) {
			return unresolved("runtime guard is missing, stale, does not dominate, or checks another proposition")
		}
		if len(claim.Outcomes) == 0 {
			return unresolved("runtime guard requires an explicit covered normal outcome")
		}
		for _, outcome := range claim.Outcomes {
			if outcome != Normal || !containsOutcome(guard.Outcomes, outcome) {
				return unresolved("runtime guard covers only its normal continuation")
			}
		}
		for resource, version := range guard.ResourceVersions {
			if version == "" || context.ResourceVersions[resource] != version {
				return unresolved("guarded state changed or is not protected")
			}
		}
		if !guardResourcesCovered(claim.Predicate, guard.ResourceVersions) {
			return unresolved("guard lacks a stable version dependency for the proposition's resource")
		}
		return EvidenceResult{Status: Established, Reason: "valid dominating guard on its stable normal continuation", Evidence: []string{guard.ID}}
	case "checked_derivation", "external_proof", "counterexample", "empirical":
		checked := context.Checked
		if node.Kind == "counterexample" {
			checked = context.Refutations
		} else if node.Kind == "empirical" {
			if claim.Predicate.Kind != "observed" {
				return unresolved("empirical executions do not establish a universal logical claim")
			}
			checked = context.Observations
		}
		fact, exists := checked[node.Witness]
		if !exists || fact.Checker == "" || fact.ClaimFingerprint != ClaimFingerprint(claim) || !validEvidenceFingerprints(fact.Fingerprints) || !sameEvidenceFingerprints(fact.Fingerprints, context.Current[node.ID]) {
			return unresolved("missing or stale proposition-specific checker output")
		}
		status := Established
		if node.Kind == "counterexample" {
			status = Contradicted
		}
		return EvidenceResult{Status: status, Reason: "checked by " + fact.Checker, Evidence: []string{node.Witness}}
	case "derived":
		switch node.Rule {
		case "identity":
			if len(node.Premises) != 1 || ClaimFingerprint(claims[nodes[node.Premises[0]].Claim]) != ClaimFingerprint(claim) {
				return unresolved("identity rule requires the same proposition and scope")
			}
		case "conjunction":
			if claim.Predicate.Kind != "and" || len(node.Premises) != len(claim.Predicate.Args) || len(node.Premises) == 0 {
				return unresolved("conjunction rule requires evidence for every conjunct")
			}
			for i, premise := range node.Premises {
				wanted := claim
				wanted.Predicate = claim.Predicate.Args[i]
				if ClaimFingerprint(wanted) != ClaimFingerprint(claims[nodes[premise].Claim]) {
					return unresolved(fmt.Sprintf("conjunction premise %d establishes another proposition or scope", i))
				}
			}
		default:
			return unresolved("unsupported derivation rule " + node.Rule)
		}
		return EvidenceResult{Status: Established, Reason: "derived by " + node.Rule + " from explicit proposition-specific premises"}
	default:
		return unresolved("unsupported or unknown evidence mechanism " + node.Kind)
	}
}

func guardResourcesCovered(predicate Predicate, versions map[string]string) bool {
	if predicate.Resource.Unknown || predicate.Resource.ID != "" && versions[predicate.Resource.ID] == "" {
		return false
	}
	for _, argument := range predicate.Args {
		if !guardResourcesCovered(argument, versions) {
			return false
		}
	}
	return true
}

func evidenceUnique(values []string) []string {
	sorted := append([]string(nil), values...)
	slices.Sort(sorted)
	return slices.Compact(sorted)
}
