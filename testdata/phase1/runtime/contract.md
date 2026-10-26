# Private Rust generated-runtime acceptance contract

The private hooks described here are implemented in the selected Rust support as of October 6, 2026. The runtime driver is materialized with actual `runtime.lu` and `process_unix.lu` bytes without rewriting them. Its `runtime-interface` check compiles those bytes before the dependent cases run; no substitute lifecycle is compiled and no case is skipped. The [implementation ledger](../../integration/IMPLEMENTATION.md) records verification separately from this maintained contract. Dated outcomes in that ledger do not change this contract.

The driver uses the existing Runtime/App/Invocation/Snapshot API. The following module is required only for an isolated `cfg(test)` crate. It is private to that crate and introduces no public library or language interface:

```rust
pub(crate) mod test_support {
    pub(crate) type Wait<'a> = Pin<Box<dyn Future<Output = io::Result<ExitStatus>> + Send + 'a>>;
    pub(crate) type Output = Pin<Box<dyn tokio::io::AsyncWrite + Send>>;
    #[derive(Clone, Copy, Debug, PartialEq, Eq)]
    pub(crate) enum Stream { Stdout, Stderr }
    #[derive(Clone, Copy, Debug, PartialEq, Eq)]
    pub(crate) enum Checkpoint { PreStart, SuccessfulStart, CapturedStopTarget, PrePublication }
    #[derive(Clone, Copy, Debug, PartialEq, Eq)]
    pub(crate) enum SignalTarget { Group(u32), Leader(u32) }
    #[derive(Clone, Copy, Debug, PartialEq, Eq)]
    pub(crate) enum TerminalRequest { Exit, Failure, Signal(u8) }
    pub(crate) trait Dependencies: Send + Sync {
        fn start(&self, command: &mut tokio::process::Command) -> io::Result<tokio::process::Child>;
        fn wait<'a>(&'a self, child: &'a mut tokio::process::Child) -> Wait<'a>;
        fn signal(&self, target: SignalTarget, signal: nix::sys::signal::Signal) -> io::Result<()>;
        fn capture<'a>(&self, stream: Stream, destination: Box<dyn io::Write + Send + 'a>) -> Box<dyn io::Write + Send + 'a>;
        fn checkpoint(&self, event: Checkpoint, invocation: Arc<Invocation>);
        fn terminal_entered(&self, request: TerminalRequest);
    }
    pub(crate) fn new_observed(dependencies: Arc<dyn Dependencies>, stdout: Output, stderr: Output) -> runtime::Result<Runtime>;
}
```

Start, Wait and Signal wrappers delegate to the real OS primitives except the explicitly selected fault. One actual wait operation is created and pinned per successful start; repeated polling is allowed, duplicate waiter creation/reaping is not. Group and leader signals remain distinguishable. The test driver does not own production process admission, cancellation, capture publication or shutdown policy.

Capture wraps the actual destination. A fault writer accepts three bytes, then returns its injected error on the next write; this preserves Rust Write's result contract and the original `abc` partial-capture assertion. Output writers independently exercise stdout/stderr faults and diagnostic attempt counts. Errors retain injected causal identity through the standard Error::source chain and io::Error::get_ref; converting the cause to its display text cannot satisfy the cause assertion. ExitStatus assertions use Snapshot.status and real reaping observations.

Checkpoint timing is unchanged: PreStart follows invocation allocation but precedes validation/admission (including refused attempts); SuccessfulStart follows spawn success and ownership registration; CapturedStopTarget follows exact invocation selection; PrePublication follows Wait/capture/group settlement and precedes immutable publication. Callbacks execute outside runtime locks held across callbacks. A blocked boundary must permit independent shutdown/settlement. Snapshot remains nonblocking and synchronized.

TerminalEntered is a nonblocking observation on entry to the actual terminal-request path, before request acceptance or rejection and outside the terminal-state lock. It also observes a competing losing request. It cannot select the winning request. The first real cleanup-signal barrier still establishes the already-accepted first decision.

The native Rust companion registers its lifetime before output, blocking or descendants. `ownership.rs.txt` implements exactly the existing admission/lifetime-lock protocol consumed by the external Go watchdog. The watchdog remains test infrastructure, not generated orchestration. It records live resources before fallback killing, and fallback cannot turn a failure into a pass.

All 98 registry scenario/assertion strings remain unchanged. The six concurrency scenarios retain their actual concurrency/barrier assertions as an additional normal test pass. Race/sanitizer/model instrumentation is not performed. Fatal-wait and supervision faults use native Rust panics/unwinding and panic markers, with the Rust test harness's nonzero status.

The [Lyger shutdown fixtures](../../../lyger/test/README.md) use the implemented read-only `terminal_status(&Runtime) -> Option<u8>` observer and test-only generated `l_entry_with_runtime` entrypoint. This runtime driver does not synthesize either implementation. Synchronous test barriers that wait for an independent Tokio task use `block_in_place` so the observer does not starve the runtime task whose entry it must observe; ordering and watchdog assertions remain unchanged.

## Error bytes and native causes

An explicitly supplied empty Error is present: failing with it selects status 1 and writes one diagnostic newline. Absence of an Error writes no diagnostic. Error construction and terminal output preserve raw bytes; the Rust `Display` implementation is only a UTF-8 diagnostic view.

`Error::causes() -> &[Error]` exposes the existing immutable cause tree without copying or flattening native failures. `std::error::Error::source()` retains Rust's primary-chain meaning. To inspect every cause of an aggregate, traverse `causes()` and use ordinary `source()`/`io::Error::get_ref` on each node. Clones retain the original cause identities; no formatted-message comparison is needed. The focused native regression injects two distinct causes with identical display text into real capture and cleanup operations and observes both after aggregation and cloning.
