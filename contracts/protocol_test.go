package contracts

import "testing"

func protocolIdle() AppProtocolState {
	return AppProtocolState{App: TargetRef{Root: "App-A"}, Stage: AppIdle, Claim: ClaimUnclaimed, Owner: "orchestrator", Authority: TransitionOwner}
}

func protocolRunning(t *testing.T) AppProtocolState {
	t.Helper()
	result := ApplyAppProtocol([]AppProtocolState{protocolIdle()}, ProtocolOperation{Kind: "Run", Invocation: "invocation-1"})
	if result.Status != Established || len(result.States) != 1 {
		t.Fatalf("Run: %+v", result)
	}
	return result.States[0]
}

func TestAppProtocolAlternativesAndAuthority(t *testing.T) {
	t.Run("BASE-038", func(t *testing.T) {
		running := protocolRunning(t)
		observer := ObserveAppProtocol([]AppProtocolState{running})
		for _, state := range observer {
			if state.App.Root != running.App.Root || state.Invocation != running.Invocation || state.Owner != running.Owner || state.Authority != Observer {
				t.Fatalf("observer lost stable facts: %+v", state)
			}
		}
		if got := ApplyAppProtocol(observer, ProtocolOperation{Kind: "Stop"}); got.Status != Contradicted {
			t.Fatalf("observer gained authority: %+v", got)
		}
	})
	t.Run("BASE-039", func(t *testing.T) {
		running := protocolRunning(t)
		closed := CloseAppEnvironment([]AppProtocolState{running})
		if len(closed) != 3 {
			t.Fatalf("external completion missing: %+v", closed)
		}
		for _, state := range closed {
			if !state.Outstanding || state.Claim != ClaimUnclaimed || state.Authority != TransitionOwner {
				t.Fatalf("external completion released slot or authority: %+v", state)
			}
		}
		if got := ApplyAppProtocol([]AppProtocolState{running}, ProtocolOperation{Kind: "ReadOutput"}); got.Status != Unresolved || !got.RuntimeObservationRequired {
			t.Fatalf("running assumed frozen/finalized: %+v", got)
		}
	})
	t.Run("BASE-040", func(t *testing.T) {
		running := protocolRunning(t)
		joined := JoinAppProtocol([]AppProtocolState{protocolIdle()}, []AppProtocolState{running})
		if len(joined) != 2 {
			t.Fatalf("branch states collapsed: %+v", joined)
		}
		if got := ApplyAppProtocol(joined, ProtocolOperation{Kind: "Await"}); got.Status != Unresolved {
			t.Fatalf("Await precondition proved for idle branch: %+v", got)
		}
	})
}

func TestAppProtocolInvocationContinuationSemantics(t *testing.T) {
	t.Run("BASE-041-model", func(t *testing.T) {
		running := protocolRunning(t)
		stopped := ApplyAppProtocol([]AppProtocolState{running}, ProtocolOperation{Kind: "Stop", Invocation: running.Invocation})
		if stopped.Status != Established {
			t.Fatal(stopped)
		}
		for _, state := range stopped.States {
			if state.Joined || !state.Outstanding {
				t.Fatalf("Stop joined/released: %+v", state)
			}
		}
		waited := ApplyAppProtocol(stopped.States, ProtocolOperation{Kind: "Await"})
		if waited.Status != Established || !waited.CancellationPossible {
			t.Fatal(waited)
		}
		claimed := ApplyAppProtocol(waited.States, ProtocolOperation{Kind: "ClaimContinuation", Invocation: running.Invocation})
		if claimed.Status != Established || !claimed.ContinuationActivated {
			t.Fatal(claimed)
		}
		repeated := ApplyAppProtocol(claimed.States, ProtocolOperation{Kind: "Await"})
		reclaimed := ApplyAppProtocol(repeated.States, ProtocolOperation{Kind: "ClaimContinuation", Invocation: running.Invocation})
		if reclaimed.Status != Established || reclaimed.ContinuationActivated {
			t.Fatalf("cached/self Await reclaimed continuation: %+v", reclaimed)
		}
		if got := ApplyAppProtocol(reclaimed.States, ProtocolOperation{Kind: "Run", Invocation: "invocation-2"}); got.Status != Contradicted {
			t.Fatalf("Run during own continuation accepted: %+v", got)
		}
		finished := ApplyAppProtocol(reclaimed.States, ProtocolOperation{Kind: "FinishContinuation", Invocation: running.Invocation})
		if finished.Status != Established {
			t.Fatal(finished)
		}
		for _, state := range finished.States {
			if state.Outstanding || state.Claim != ClaimFinished {
				t.Fatal(state)
			}
		}
		next := ApplyAppProtocol(finished.States, ProtocolOperation{Kind: "Run", Invocation: "invocation-2"})
		if next.Status != Established || next.States[0].Invocation == running.Invocation {
			t.Fatalf("new invocation identity lost: %+v", next)
		}
		if got := ApplyAppProtocol(next.States, ProtocolOperation{Kind: "Stop", Invocation: running.Invocation}); got.Status != Unresolved {
			t.Fatalf("old Stop targeted replacement: %+v", got)
		}
	})
	t.Run("startup-failure-is-awaitable", func(t *testing.T) {
		run := ApplyAppProtocol([]AppProtocolState{protocolIdle()}, ProtocolOperation{Kind: "Run", Invocation: "failed-start", StartupFailure: true})
		if run.Status != Established || run.States[0].Stage != AppFinalized || !run.States[0].Outstanding {
			t.Fatal(run)
		}
		wait := ApplyAppProtocol(run.States, ProtocolOperation{Kind: "Await"})
		claim := ApplyAppProtocol(wait.States, ProtocolOperation{Kind: "ClaimContinuation", Invocation: "failed-start"})
		if claim.Status != Established || !claim.ContinuationActivated {
			t.Fatal(claim)
		}
	})
}

func TestActivityOwnership(t *testing.T) {
	running := protocolRunning(t)
	if got := CheckOwnedLifecycle([]AppProtocolState{running}); got.Status != Established {
		t.Fatal(got)
	}
	transferred, obligation := TransferActivity(running, "explicit-supervisor")
	if obligation.Status != Established || transferred.Owner != "explicit-supervisor" || transferred.Stage != AppRunning || !transferred.Outstanding {
		t.Fatalf("transfer altered lifecycle: %+v, %+v", transferred, obligation)
	}
	running.Owner = ""
	if got := CheckOwnedLifecycle([]AppProtocolState{running}); got.Status != Unresolved {
		t.Fatalf("unowned work accepted: %+v", got)
	}
	if got := CheckOwnedLifecycle(nil); got.Status != Unresolved {
		t.Fatalf("missing inventory accepted: %+v", got)
	}
}
