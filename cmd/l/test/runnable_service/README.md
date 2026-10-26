# Retained process fixtures

The fourteen positive programs construct Apps directly. Standalone Services
remain structural records: declarations emit no independent source target,
process adapter, or continuation, and standalone Service bodies are unsupported.
`library-only.l` retains its original body as a negative diagnostic witness.

| Source | Contract |
| --- | --- |
| `basic.l` | Explicit Build, Run, Await, raw Output |
| `alias.l` | Basic descriptor alias, target alias, repeated Await |
| `mixed.l` | Independent App process handles and captures |
| `multiple.l` | Ordered multiple targets and separate App companions |
| `canonical-io.l` | Canonical IO Print / Basic App example |
| `unicode.l` | Quotes, Unicode, newlines and trailing output spaces |
| `snapshot.l` | Owned finalized-output snapshot across a second invocation |
| `stop.l` | Stop, completed Stop and cached Await |
| `build-only.l` | Build validates without launching |
| `empty.l` | Explicit empty entrypoint |
| `passive.l` | Passive lifecycle forms preserve primitive semantics |
| `imported.l` | Explicit imported App declaration through `worker.l` |
| `private.l` | Private paragraph and scoped descriptor alias |
| `conditional-exit.l` | Await continuation's conditional Exit exits the orchestrator |

The positive source changes reuse the independently authored App sources in
`controls/`. Those controls remain unchanged. `worker.l` uses the existing
`worker-app-control.l` bytes while `imported.l` retains its original import path.
The lifecycle tests also contain App source directly; they no longer translate a
Service program into a separate App control at runtime.

Each `expected-cases/<source>/` contains the independently authored App and
orchestrator Rust sources and manifests. The App body and orchestrator oracles
materialize the exact expected-only transformations previously checked by the
App shape controls; no generated output was captured. App manifests use the
retained ordinary App manifest contract. The original mixed Host App expectation
is unchanged. Standalone library and adapter expectation trees are
excluded by the finite amendment.

Shared `expected/config.rs` retains empty orchestrator settings.
`expected/app-config.rs` is a byte-identical copy of the existing mixed Host App
configuration, including the ordinary App's default Host setting. The comparison
preserves explicit per-case configuration and supplies the matching shared file
only when a configuration is absent. Runtime root/manifest expectations and the
selected canonical runtime/process sources are unchanged.

Actual Rust is never normalized. Only expected `@@SOURCE@@` and
`@@GENERATION@@` slots are instantiated from known source and validated manifest
locations. Separate controls compile a single independent expected App and all
fourteen complete expected workspaces. They retain the pinned external lock
fragment and assemble only temporary local package records. No retained lock is
regenerated and no expected source replaces actual generated output.

Installed-CLI tests require source generation, a real frozen/offline Cargo build,
exact complete workspace bytes, version-1 App publication, and supervised
execution. Metadata and compiler-artifact JSONL bind each published binary to its
real Cargo package, target, source path and immutable SHA256 snapshot. Additional
observers retain relocation, immutable and concurrent generations, failed rebuild
preservation, protected destinations, symlink rejection and package-name
allocation. Concurrent jobs change both App outputs and require an exact pair
from the same generation. Readiness/lifetime-lock, Stop pipe-transfer, signal,
raw-output, invocation and continuation protocols remain unchanged.

Independent native library consumption now uses the unchanged App-owned library
fixture `lyger/test/workspace/services.l`. Its actual first library is linked and
executed by an independent Rust caller. `TestConstructedServicesCompileAsLibraries`
separately retains production workspace coordination, two Apps and four libraries.
The backend's complete AML, library visibility, exact composition and separate
native compilation witnesses remain under `logiceme/testdata/runnable_service`.

Compilation prerequisites remain bounded at 40 seconds. Ordinary native runs
retain a 20-second watchdog; readiness-gated observers and the Stop probe retain
their 5-second watchdogs. Cumulative runtime allowances across rebuilding,
process-group cleanup and all test-owned protocol bytes are unchanged. The
repository's validation pause remains in force.
