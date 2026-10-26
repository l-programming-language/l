# Rust runtime parity fixtures

These additive observers materialize the actual selected Rust runtime and process assets unchanged. They are separate from the maintained complete-AML logeme corpus. The Go test process is supervision infrastructure; executable runtime drivers and companions come from pinned Cargo artifacts.

`public.rs.txt` covers directly exposed operations. `observed.rs.txt` consumes the private `runtime::test_support` contract documented in [the Phase 1 runtime guide](../../../testdata/phase1/runtime/contract.md). The selected runtime implements these test-only observers to preserve the original private assertions:

```rust
pub(crate) fn latest_invocation(app: &Arc<App>) -> Option<Arc<Invocation>>;
pub(crate) fn observe_invocation(
    child: tokio::process::Child,
    dependencies: Arc<dyn Dependencies>,
    wait_delay: std::time::Duration,
) -> ObservedInvocation;
impl ObservedInvocation {
    pub(crate) async fn terminate(&self, final_signal: bool) -> io::Result<()>;
    pub(crate) async fn wait(&self) -> io::Result<std::process::ExitStatus>;
    pub(crate) fn deadline(&self) -> Option<tokio::time::Instant>;
    pub(crate) fn pid(&self) -> u32;
}
```

`ObservedInvocation` must wrap the production Invocation. Its methods delegate the actual termination, original deadline and single Wait owner. The observer must not implement replacement termination, a second timer, reaping, or error policy. Tokio's existing clock is retained. The manually owned group member remains a separate native companion whose Wait is intentionally withheld. The five original Darwin cases assert real process-group results, initial-deadline accounting, no renewal, exact signal counts and exactly-once Wait. Linux preserves the original non-Darwin no-inspection check. Missing observer APIs block these assertions with a named compilation failure; direct process-helper checks alone do not establish them.

`latest_invocation` is read-only. It permits inspecting an already-canceled attempt's completed empty snapshot without requiring public Await to succeed after shutdown. The no-launch marker, zero waits, selected error and public cancellation remain checked.

`fail_absent.rs.txt` supplies `None` to `Runtime::fail`, whose `Into<Option<Error>>` input preserves the absent-Error case. `TestRuntimeTemplateBehavior/TestRuntimeVisibleStreams/fail nil` requires exit status 1 and the exact diagnostic `cannot exit with an absent Error\n`. The runtime supplies that diagnostic; the fixture must never manufacture it.

The Cargo manifests and initial locks are test-only inputs. `generation.Cargo.lock` supports compilation of actual runtime-enabled returned bundles. No actual returned source or production support bytes are rewritten before compilation. The native helpers register with the shared ownership watchdog before observable work.
