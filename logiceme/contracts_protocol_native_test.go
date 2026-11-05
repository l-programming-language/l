//go:build darwin || linux

package Logiceme

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/l-programming-language/l/contracts"
)

// These are cross-tests, not native trust witnesses. The retained drivers use
// production synchronization and completion publication; this model never
// replaces the driver's runtime checks or changes selected runtime assets.
func TestContractsRuntimeProtocolCrossCheck(t *testing.T) {
	t.Run("BASE-041", func(t *testing.T) {
		state := contracts.AppProtocolState{App: contracts.TargetRef{Root: "fixture:readiness", Type: "App", Phase: contracts.Invocation}, Stage: contracts.AppIdle, Claim: contracts.ClaimUnclaimed, Owner: "orchestrator", Authority: contracts.TransitionOwner}
		run := contracts.ApplyAppProtocol([]contracts.AppProtocolState{state}, contracts.ProtocolOperation{Kind: "Run", Invocation: "first"})
		await := contracts.ApplyAppProtocol(run.States, contracts.ProtocolOperation{Kind: "Await"})
		claim := contracts.ApplyAppProtocol(await.States, contracts.ProtocolOperation{Kind: "ClaimContinuation", Invocation: "first"})
		duplicate := contracts.ApplyAppProtocol(claim.States, contracts.ProtocolOperation{Kind: "ClaimContinuation", Invocation: "first"})
		finish := contracts.ApplyAppProtocol(claim.States, contracts.ProtocolOperation{Kind: "FinishContinuation", Invocation: "first"})
		second := contracts.ApplyAppProtocol(finish.States, contracts.ProtocolOperation{Kind: "Run", Invocation: "second"})
		if run.Status != contracts.Established || await.Status != contracts.Established || !claim.ContinuationActivated || duplicate.ContinuationActivated || second.Status != contracts.Established || !second.RuntimeObservationRequired {
			t.Fatalf("modeled readiness/claim/release sequence: run=%+v await=%+v claim=%+v duplicate=%+v second=%+v", run, await, claim, duplicate, second)
		}
		_, native := runtimeFixture(t, false)
		t.Log("generation: selected production runtime support materialized unchanged")
		driver := runtimeDriver(t, native)
		t.Log("native compilation: pinned public driver and companion artifacts verified")
		mode := filepath.Join(t.TempDir(), "mode")
		if err := os.WriteFile(mode, nil, 0600); err != nil {
			t.Fatal(err)
		}
		runtimeExecute(t, driver, []string{"L_RUNTIME_CASE_FILE=" + mode}, []string{"readiness"}, 0, "", "")
		t.Log("runtime: readiness, fresh invocation, finalized output and exactly-once continuation passed")
	})
	t.Run("observed", func(t *testing.T) {
		_, native := runtimeFixture(t, true)
		t.Log("generation: selected production runtime support with retained read-only observers materialized unchanged")
		driver := runtimeDriver(t, native)
		t.Log("native compilation: pinned observed driver and companion artifacts verified")
		t.Run("BASE-036", func(t *testing.T) {
			const diagnostic = "terminate application process group: injected cleanup failure: notice\n"
			cleanup := contracts.CleanupFailure{ID: "invocation-cleanup", Message: diagnostic, LaunchOrder: 1}
			report := contracts.RuntimeCleanupReportWithCauses(diagnostic, []string{cleanup.ID}, 1, []contracts.CleanupFailure{cleanup})
			if report.ExitStatus != 1 || report.Primary != diagnostic || len(report.Secondary) != 0 || report.Coverage != contracts.Known {
				t.Fatalf("cleanup cause identity was not retained: %+v", report)
			}
			pid := filepath.Join(t.TempDir(), "pid")
			runtimeExecute(t, driver, []string{"L_RUNTIME_CASE=", "L_RUNTIME_PID=" + pid}, []string{"cleanup-once"}, 1, "", diagnostic+"\n")
			t.Log("runtime: accepted cleanup error retains its primary cause and is emitted once")
		})
		t.Run("BASE-039-external-completion", func(t *testing.T) {
			state := contracts.AppProtocolState{App: contracts.TargetRef{Root: "fixture:observed", Type: "App", Phase: contracts.Invocation}, Invocation: "active", Stage: contracts.AppRunning, Outstanding: true, Claim: contracts.ClaimUnclaimed, Owner: "orchestrator", Authority: contracts.TransitionOwner}
			observer := contracts.ObserveAppProtocol([]contracts.AppProtocolState{state})
			finalized := false
			for _, candidate := range observer {
				if candidate.Stage == contracts.AppFinalized {
					finalized = true
				}
				if candidate.Authority != contracts.Observer || !candidate.Outstanding || candidate.Joined {
					t.Fatalf("external completion changed stable authority/slot or invented a join: %+v", candidate)
				}
			}
			if !finalized || contracts.ApplyAppProtocol(observer, contracts.ProtocolOperation{Kind: "ReadOutput"}).Status != contracts.Unresolved {
				t.Fatal("observer must retain possible completion without proving a stale output guard")
			}
			pid := filepath.Join(t.TempDir(), "pid")
			runtimeExecute(t, driver, []string{"L_RUNTIME_CASE=hold", "L_RUNTIME_PID=" + pid}, []string{"stop-error-retained"}, 0, "", "")
			t.Log("runtime: external final publication observed before Await, retained Stop error and exactly-once Wait passed")
		})
	})
}
