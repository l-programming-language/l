package contracts

import "fmt"

type ProtocolAuthority string
type AppStage string
type ContinuationClaim string

const (
	TransitionOwner  ProtocolAuthority = "transition_owner"
	Observer         ProtocolAuthority = "observer"
	SharedTransition ProtocolAuthority = "shared_transition"
	UnknownAuthority ProtocolAuthority = "unknown"
	AppIdle          AppStage          = "idle"
	AppRunning       AppStage          = "running"
	AppDraining      AppStage          = "draining"
	AppFinalized     AppStage          = "finalized"
	ClaimUnclaimed   ContinuationClaim = "unclaimed"
	ClaimActive      ContinuationClaim = "active"
	ClaimFinished    ContinuationClaim = "finished"
)

// AppProtocolState separates canonical App identity, the current invocation,
// the invocation slot, and the continuation claim. Finalization is not release.
type AppProtocolState struct {
	App            TargetRef         `json:"app"`
	Invocation     string            `json:"invocation,omitempty"`
	Stage          AppStage          `json:"stage"`
	Outstanding    bool              `json:"outstanding"`
	Claim          ContinuationClaim `json:"claim"`
	Owner          string            `json:"owner,omitempty"`
	Authority      ProtocolAuthority `json:"authority"`
	StopRequested  bool              `json:"stop_requested,omitempty"`
	StartupFailure bool              `json:"startup_failure,omitempty"`
	Joined         bool              `json:"joined,omitempty"`
}

type ProtocolOperation struct {
	Kind           string `json:"kind"`
	Invocation     string `json:"invocation,omitempty"`
	Owner          string `json:"owner,omitempty"`
	StartupFailure bool   `json:"startup_failure,omitempty"`
}

type ProtocolResult struct {
	States                     []AppProtocolState `json:"states"`
	Status                     Status             `json:"status"`
	Reason                     string             `json:"reason"`
	RuntimeObservationRequired bool               `json:"runtime_observation_required"`
	CancellationPossible       bool               `json:"cancellation_possible,omitempty"`
	ContinuationActivated      bool               `json:"continuation_activated,omitempty"`
}

// CloseAppEnvironment adds possible external completion without changing
// stable identity, ownership or continuation authority. A child can exit while
// its transition owner performs no operation. Draining is not final output.
func CloseAppEnvironment(states []AppProtocolState) []AppProtocolState {
	var result []AppProtocolState
	for _, state := range states {
		result = appendProtocolState(result, state)
		if state.Stage == AppRunning {
			draining := state
			draining.Stage = AppDraining
			result = appendProtocolState(result, draining)
		}
		if state.Stage == AppRunning || state.Stage == AppDraining {
			finalized := state
			finalized.Stage = AppFinalized
			result = appendProtocolState(result, finalized)
		}
	}
	return result
}

func JoinAppProtocol(branches ...[]AppProtocolState) []AppProtocolState {
	var result []AppProtocolState
	for _, branch := range branches {
		for _, state := range branch {
			result = appendProtocolState(result, state)
		}
	}
	return result
}

func ObserveAppProtocol(states []AppProtocolState) []AppProtocolState {
	result := make([]AppProtocolState, len(states))
	for i, state := range states {
		state.Authority = Observer
		result[i] = state
	}
	return CloseAppEnvironment(result)
}

func ApplyAppProtocol(states []AppProtocolState, operation ProtocolOperation) ProtocolResult {
	result := ProtocolResult{Status: Established, RuntimeObservationRequired: true}
	if len(states) == 0 {
		return ProtocolResult{Status: Unresolved, Reason: "protocol input state is unknown", RuntimeObservationRequired: true}
	}
	accepted, rejected, unknown := 0, 0, 0
	for _, state := range CloseAppEnvironment(states) {
		next, status, reason, activated := applyAppProtocolState(state, operation)
		result.States = appendProtocolState(result.States, next)
		result.ContinuationActivated = result.ContinuationActivated || activated
		switch status {
		case Established:
			accepted++
		case Contradicted:
			rejected++
			result.Reason = reason
		default:
			unknown++
			result.Reason = reason
		}
	}
	if unknown != 0 || rejected != 0 && accepted != 0 {
		result.Status = Unresolved
		result.Reason = "operation precondition does not hold over every possible state; preserve the existing runtime guard: " + result.Reason
	} else if rejected != 0 {
		result.Status = Contradicted
	} else {
		result.Reason = "operation is admitted by the modeled protocol; runtime synchronization remains required"
	}
	result.CancellationPossible = operation.Kind == "Await" || operation.Kind == "Stop" || operation.Kind == "Run"
	return result
}

func applyAppProtocolState(state AppProtocolState, operation ProtocolOperation) (AppProtocolState, Status, string, bool) {
	if state.App.Unknown || state.App.Root == "" || !validAppProtocolState(state) {
		return state, Unresolved, "unknown or inconsistent App protocol state", false
	}
	if operation.Kind != "ReadOutput" && operation.Kind != "Observe" {
		if state.Authority == Observer {
			return state, Contradicted, "observer has no transition authority", false
		}
		if state.Authority != TransitionOwner && state.Authority != SharedTransition {
			return state, Unresolved, "transition authority is unknown", false
		}
	}
	switch operation.Kind {
	case "Observe":
		return state, Established, "stable identity and authority retained", false
	case "Build":
		if state.Outstanding {
			return state, Contradicted, "Build conflicts with an outstanding invocation", false
		}
	case "Run":
		if state.Outstanding {
			return state, Contradicted, "Run conflicts with an outstanding invocation", false
		}
		if operation.Invocation == "" || operation.Invocation == state.Invocation {
			return state, Unresolved, "Run requires a fresh invocation identity", false
		}
		state.Invocation = operation.Invocation
		state.Stage = AppRunning
		state.Outstanding = true
		state.Claim = ClaimUnclaimed
		state.StopRequested, state.Joined = false, false
		state.StartupFailure = operation.StartupFailure
		if operation.Owner != "" {
			state.Owner = operation.Owner
		}
		if operation.StartupFailure {
			state.Stage = AppFinalized
		}
	case "Stop":
		if state.Stage == AppIdle {
			return state, Contradicted, "Stop requires an invocation", false
		}
		if operation.Invocation != "" && operation.Invocation != state.Invocation {
			return state, Unresolved, "Stop must retain its captured invocation across races", false
		}
		if state.Stage != AppFinalized {
			state.StopRequested = true
		}
	case "Await":
		if state.Stage == AppIdle {
			return state, Contradicted, "Await requires an invocation", false
		}
		// This is the covered normal continuation. Global cancellation aborts
		// source flow instead, and is reported separately on ProtocolResult.
		state.Stage = AppFinalized
		state.Joined = true
	case "ClaimContinuation":
		if state.Stage != AppFinalized {
			return state, Contradicted, "continuation claim requires finalized output", false
		}
		if operation.Invocation != "" && operation.Invocation != state.Invocation {
			return state, Unresolved, "continuation claim requires its invocation identity", false
		}
		if state.Claim == ClaimUnclaimed {
			state.Claim = ClaimActive
			return state, Established, "first claim activates this invocation's continuation", true
		}
	case "FinishContinuation":
		if operation.Invocation == "" || operation.Invocation != state.Invocation {
			return state, Unresolved, "continuation finish requires the current invocation identity", false
		}
		if state.Claim == ClaimActive {
			state.Claim = ClaimFinished
			state.Outstanding = false
		}
	case "ReadOutput":
		if state.Stage != AppFinalized {
			return state, Contradicted, "Output is unavailable before finalization; reading does not join", false
		}
	default:
		return state, Unresolved, "unsupported protocol operation " + operation.Kind, false
	}
	return state, Established, "", false
}

func validAppProtocolState(state AppProtocolState) bool {
	switch state.Stage {
	case AppIdle:
		return state.Invocation == "" && !state.Outstanding && !state.Joined
	case AppRunning, AppDraining:
		return state.Invocation != "" && state.Outstanding && state.Claim == ClaimUnclaimed && !state.Joined && !state.StartupFailure
	case AppFinalized:
		return state.Invocation != "" && (state.Claim == ClaimUnclaimed || state.Claim == ClaimActive || state.Claim == ClaimFinished) && (state.Outstanding == (state.Claim != ClaimFinished))
	default:
		return false
	}
}

func appendProtocolState(values []AppProtocolState, state AppProtocolState) []AppProtocolState {
	identity := StableID("protocol", state)
	for _, existing := range values {
		if StableID("protocol", existing) == identity {
			return values
		}
	}
	return append(values, state)
}

func CheckOwnedLifecycle(states []AppProtocolState) Obligation {
	result := Obligation{ID: "owned-lifecycle", Status: Established, Reason: "every modeled activity retains its supervisor or explicit owner; no eventual termination is claimed"}
	if len(states) == 0 {
		result.Status, result.Reason = Unresolved, "no lifecycle ownership facts"
	}
	for _, state := range states {
		if !validAppProtocolState(state) || state.App.Unknown || state.App.Root == "" {
			result.Status, result.Reason = Unresolved, "incomplete lifecycle state"
			return result
		}
		if state.Stage != AppIdle && state.Owner == "" {
			result.Status, result.Reason = Unresolved, fmt.Sprintf("invocation %s has no established owner or transfer", state.Invocation)
			return result
		}
	}
	return result
}

func TransferActivity(state AppProtocolState, owner string) (AppProtocolState, Obligation) {
	if owner == "" || state.Authority != TransitionOwner || state.Stage == AppIdle || !validAppProtocolState(state) {
		return state, Obligation{ID: "activity-transfer", Status: Unresolved, Reason: "transfer requires an existing activity, known transition owner and explicit recipient"}
	}
	state.Owner = owner
	return state, Obligation{ID: "activity-transfer", Status: Established, Reason: "activity ownership transferred explicitly; no implicit termination added"}
}

func CheckLifecycleProfile(states []AppProtocolState, profile Profile) []Obligation {
	result := CheckOwnedLifecycle(states)
	result.Claim = Claim{ID: result.ID, Predicate: Predicate{Kind: "profile", Subject: "OwnedLifecycle"}, Target: profile.Target, Scope: profile.Scope, Outcomes: profile.Outcomes}
	if profile.Name != "OwnedLifecycle" {
		result.Status, result.Reason = Unresolved, "lifecycle projection requires the OwnedLifecycle profile"
	}
	return []Obligation{result}
}
