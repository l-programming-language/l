# Building and running L programs

This guide describes the Rust-only generation and native CLI runtime path. L, Lyger and Logiceme are implemented in Go. Building L programs requires the pinned Rust 1.98.1 toolchain and retained dependencies available locally; published binaries need no language compiler or L installation.

On native macOS or Linux:

```sh
l build main.l
./main
```

The compiler prints the published executable's absolute path on success. Building generates and compiles the program; it never runs its Apps or generated deployment scripts. Running the executable executes source instructions in order while Apps run in separate OS processes until joined with Await.

## Invocation and assets

[CLI options](cli/build.md) precede the source filename (default: `main.l` in the invocation directory). Use `--` for filenames starting with a dash. Extra arguments and unknown options reject; [`-os` is unimplemented](cli/build.md#target-os).

Embedded English resources work with relocated binaries, symlinks and `go run`, independently of CWD. Relative output paths use invocation CWD. [Developer overrides](development-harness.md#compiler-defaults) support compiler resource development.

Builtins are implicit. Bare dictionary paths such as `io` resolve only through
the selected bundle's `stdlib/index.json`; neighboring `io` or `io.l` cannot
shadow them. Explicit `.l` imports resolve relative to their importing source.
Imported dictionary helpers resolve relative to their resource provider and do
not read unrelated host files. Install instructions are in [README](../README.md).

Help, version and builds return 0 on success; invalid invocations or unsupported hosts return 2; build or output-write failures return 1. Diagnostics retain phase/cause on stderr. Build stdout contains the executable path and newline; standalone [`l --version` or `l -v`](cli/build.md#version) prints the version without compiler initialization.

## Output and publication

For `main.l`, the default layout is:

```text
main.l
main
main.apps/
  .l-owner
  generations/
    <generation-id>/
      manifest.json
      app-000
      app-001
```

Another source stem produces the corresponding executable and companion basename. `--output-dir` moves both outputs together. A generation manifest records companion identities, ordinal filenames, native target and executable hashes. Each generated orchestrator embeds its own generation's relative companion paths. It resolves companions against its executable location, so it may be launched from another directory or moved with its companion directory.

Builds retain version-1 manifests and the `apps` inventory. App-constructed libraries add no companion entry; standalone Service declarations add no build package.

Every build uses a unique staging directory on the output filesystem. All sources and binaries finish before publication. An explicitly marked companion directory holds immutable generations; the generation is renamed into place before the adjacent executable is atomically replaced. That last rename is the publication commit point. Failed builds preserve the previous executable and its generation. An already-published generation remains if the final rename fails. Old generations are not pruned because older running orchestrators may still use them. Concurrent builds cannot mix App generations; the last successful executable replacement wins.

The runtime retains its canonical executable directory before source execution.
Later Build and Run operations resolve their embedded generation paths against
that directory and revalidate each companion. This preserves a running
generation when Linux reports its replaced executable as deleted; an initial
location error remains a Build/Run error rather than failing runtime setup.

Unrelated existing output destinations and symlinked companion/generation directories are rejected. The compiler removes only its own invocation's staging on failure. A partial compilation never overwrites a prior executable.

Private staging cleanup is always attempted. If the build already failed, a
cleanup failure is added as an IO cause without replacing the original error.
After successful publication, temporary-directory cleanup is best-effort: it
cannot turn a committed executable replacement into a reported build failure.
Native artifact inspection descriptors close before publication; close failures
reject the build and preserve any accompanying validation or copy error.

## Source generation and native builds

Lyger preserves canonical application books and projects each App and its additional constructed Services into workspace units. Logeme evaluates their unchanged AML; the separate synthetic orchestrator requests actual runtime support. Expanded source and the observable/backend AML views retain their distinct purposes.

The CLI accepts Rust source bundles only. Whole bundles are checked for matching dimensions, safe logical paths, duplicate/colliding entries and valid Cargo target/dependency relationships. The workspace contains an orchestrator binary, ordered App binaries and App-constructed Service libraries. Recursive Classes supply modules and Objects remain files. A leading slash is a logical bundle root, never a host absolute destination. Compiler-owned visibility and module assembly preserve declared ownership; no generated source rewriting is performed.

The lifecycle runtime is a compiler-owned library at `support/runtime`, linked
into the orchestrator. Its allocated package name avoids application, Service,
entry and dependency-alias collisions. Its runtime/process source bytes come
unchanged from the selected resources. This library is built with the workspace;
it introduces no separately installed runtime or launcher. Splitting it from the
entry allows Cargo to reuse it when only orchestrator source changes within the
locked build slot and the library's actual inputs remain unchanged.

Cargo builds every workspace package with Rust 1.98.1, `--frozen`, offline resolution and native target selection. The coordinator isolates caller target/compiler flags and uses retained pinned dependency resolution; it does not fetch arbitrary new dependencies. Builds supervise child process groups and retain their existing deadlines. A private versioned cache retains compiler products, with concurrent builds serialized through artifact snapshotting so publication cannot pick up another source's executable.

Local packages may share names with retained registry packages. Lockfile references
use the actual package name, version and, when needed, registry source to distinguish
them; dependency aliases do not change package identity. The coordinator preserves
retained pins and checksums and qualifies only ambiguous retained references.

The orchestrator's internal Cargo target avoids workspace package and binary target
names, as well as `build`, `deps`, `examples` and `incremental`, with
case-insensitive matching for portability. Conflicting source stems use the
first available internal name `l_entry`, `l_entry_1`, and so on.
Public output and companion names still follow the requested source stem:
`build.l` publishes `build` and `build.apps/`.

Companion Apps use unique binary target names under case-insensitive comparison.
The detached workspace projection retains the first safe target in source order
and aliases reserved names or later case-equivalent targets to `l_app`, `l_app_1`,
and subsequent free names. Allocation first reserves all canonical IDs, projected
packages and explicit targets, including later declarations and libraries.
Hyphens and underscores remain distinct. A renamed App's projected
package, Book, App metadata and primary target receive the same internal name;
canonical L App IDs, library identities, source topology and publication paths
remain unchanged. Native Cargo-name introspection sees this internal alias.
Direct Logiceme inputs retain their supplied names. Published companions still
use ordinal filenames and canonical IDs, so the entrypoint and its adjacent
companion directory can be relocated together.

The coordinator independently rejects duplicate binary targets under
case-insensitive comparison before creating the build workspace. Library-only
members occupy no binary targets; case-equivalent package names are allowed when
their binary targets are distinct.

Parser checks call the same local-FlagSet parser as the CLI. Native build checks
use the production workspace coordinator, including single-orchestrator cases;
cache cancellation checks exercise its bounded cache admission directly. No
separate test-only compilation or lock-wait adapter remains.

Structured Cargo messages must identify the requested manifest and binary target. Verified native executables are copied into the invocation's private staging before publication. The compiler does not guess output paths, launch Go as an intermediary, or ship its build toolchain. Generated deployment scripts are artifacts and are never run by `l build`.

## Service composition

Services organize source within App construction. The builtin Basic descriptor
still composes a Main Service, Object and Function. Additional Services
constructed by an App remain separate library packages with their existing
public interfaces; the primary Service retains the App's binary topology.

A standalone `Create "Worker" as a New Main Service` is a structural declaration.
It creates no library package, adapter, process or continuation. Standalone body
effects such as `The Service Prints ...` reject, as do native Build/Run/Stop/Await
and implicit Output/Error on a Service. Ordinary source-defined actions and
explicit members retain lexical lookup and can shadow native operation names.

Use an App to execute a body:

```text
Use IO's Print from `io` as "Print".
Create "Worker" as a New Basic App.
The App Prints `work completed`.
Run Worker.
Await Worker.
Show Worker's Output.
```

Raw AML still supports explicit library/bin targets and native consumers, with
its existing visibility and target checks. That backend facility does not make
standalone L Services executable. See the [coverage map](../testdata/service-execution/coverage.md).

## App execution semantics

### Launch and language forms

Each App runs in a separate OS process. The following invocation, result, cancellation and cleanup rules apply to Apps. `Run` returns after successful process creation or recording terminal startup failure. It does not await readiness, process exit, capture completion, or a user continuation. Internal Tokio supervision/capture tasks are permitted; launch acknowledgment remains part of the ordered runtime operation.

`Start` is ordinary source composition Build then Run. `Execute` is Start then Await. Both retain their `any App` source signatures. Run revalidates after Build, including when Build failed. Native coordinated operations process operands in source order; a source-defined action expands its entire body for each operand before the next operand. Aliases share one canonical target and invocation slot. Start and Execute are absent from primitive-root vocabulary.

Imperative Await, outer passive Await, and applicative Await in source actions execute. Native Await in an App body reports unsupported ownership. Standalone Service body effects reject. A standalone applicative statement and the stray CONDITIONAL suffix are invalid. Existing Stop forms remain recognized.

### Invocation and results

One invocation remains outstanding until its first Await continuation finishes. Completion alone does not release the slot. Await joins process reaping, capture finalization, and owned-group cleanup, then returns the immutable invocation result. Continuations execute inline exactly once, including failed launch, nonzero exit, and per-target Stop. Public Await on global shutdown aborts source execution; it never returns to the continuation.

ClaimContinuation marks an invocation active before entering its body. Cached/self-Await returns the same result and cannot reclaim it. Nested Await of another App is allowed; lexical paragraph identity cannot suppress another target's claim. FinishContinuation releases the slot on normal completion. Run of the same target during its own continuation is rejected, as are Start/Execute for that App. Continuation Exit exits the outer orchestrator, including from expanded source actions. Fresh accepted invocations get fresh continuation locals.

Final Output is raw stdout, including NUL, whitespace, and Unicode. Stderr is captured separately and never automatically forwarded. Stderr alone does not imply failure. Reading Output before final completion is a fatal lifecycle misuse, including unused assignments; there is no implicit join. Finalized Output may be read before Await without activating a continuation. Retained invocation snapshots and scalar copies survive reruns unchanged.

Build alone creates no invocation. Run startup failure creates an awaitable failed invocation and performs no OS Wait on an unstarted command. Await/Stop before Run, duplicate Run/Start, Build while outstanding, and Output-unavailable are fatal source misuse with source attribution. Ordinary child failures are stored results, not implicit orchestrator Exit.

Build selects its operation error. Stop signal-delivery failure selects an operation error visible through the target's Error member, remains recoverable, and preserves the invocation and ownership. A later Await, including cached Await, reselects that invocation's immutable result. A completed Stop is an idempotent no-op. Stop captures its invocation before signaling and cannot signal a replacement invocation after a race.

### Shutdown, streams and limits

Stop requests SIGKILL of the owned process group and returns after request submission/known completion; it is not a join. Natural exit and Stop may race, but only one coherent terminal result is published. Leader exit while capture/group cleanup is pending is not finalized completion.

On macOS, final group signaling can return `EPERM` during group exit. After the first successful Stop or shutdown termination request, finalization can observe group disappearance using only the time remaining in the existing two-second bound, measured from that attempt's start. Later signals do not renew the window. Only `ESRCH` from signal zero confirms disappearance; a live group, another error or an expired window preserves the cleanup failure. This sends no additional termination signal and performs no additional Wait.

Normal fallthrough and bare Exit clean all owned processes, perform no unawaited continuations, and exit 0 if cleanup succeeds. Explicit Error/fatal misuse cleans up then exits 1. SIGINT/SIGTERM abort source flow, settle ownership, and exit 130/143 respectively, including between launches or while Await is pending. Preserve an explicit primary error and append cleanup failures once in launch order; cleanup failure on otherwise successful exit yields 1.

Generated main invokes runtime finalization, which performs cleanup before closing normal terminal admission. A fatal or signal request accepted during that cleanup must finish exiting before main can return. `Shutdown` remains an idempotent, nonterminal cleanup operation for internal callers. Successful normal finalization returns normally without converting a native panic into success.

Broken stdout/stderr pipes must trigger cleanup and exit 1, including descriptor-1/2 SIGPIPE behavior. Error reporting must not recursively write to a failed diagnostic writer. Tests use subprocess exit behavior, never a fake exit callback. Cleanup is explicit on terminal paths; unwinding alone cannot cover process exit.

When fatal and signal requests compete, the first accepted terminal request selects the exit status and primary diagnostic. Later requests do not replace that decision during cleanup. Failure to write the diagnostic still selects status 1 and never retries that writer.

Keep the existing two-second post-exit/cancellation pipe-drain bound. There is no normal process-duration timeout, no process-count cap, and no capture-size cap. Fixed stress counts are test workload only. Every launched direct child has exactly one Wait owner. Same-group descendants must terminate even if they close inherited pipes. Escaped process groups and uncatchable orchestrator termination (SIGKILL, host/power loss) are documented limits, not skipped required feature tests.

Children inherit environment and caller CWD, receive EOF on stdin and receive no synthesized arguments. Runtime Build checks the prebuilt executable; it invokes no build tool. Show writes Text exactly and formats Number/Truth values without adding a newline or replaying unrelated captured output.

The lowerer checks scalar/operator types and definite assignment across continuing branches. Exit removes a branch from that intersection; Output availability remains a runtime check. Diagnostics retain source attribution. Maintained programs import the `IO.Print` member from the `io` dictionary and use implicit Basic, Start and Execute definitions. Direct Run requires an explicit Await wherever subsequent source needs completion or continuation effects.

Phase 2 in-code concurrency, runtime argument transport, nested layout syntax, containers, cross-compilation, remote deployment and unrestricted installation of third-party dependencies remain outside this build command. The [architecture](architecture.md) describes compiler/runtime ownership and [development guide](development-harness.md) describes validation.

Native build caching uses six bounded slots. Complete input paths and bytes
select a preferred slot for reuse; an available slot can admit the build when
that slot is busy. Admission waits cancellably only while all slots are occupied.
Each slot owns a cross-process lock, stable exact source copies and Cargo's
dependency products. Every build still invokes Cargo; the
coordinator checks manifest/target identities and snapshots actual executables
before releasing the lock. The cache stores no successful-test or publication
substitute. Different workspaces can build concurrently without exposing another
build's mutable executable paths.

CLI staging reconciles an explicit input inventory at a fixed private source
root in each slot. It validates the supplied tree and all recorded cached paths
before changing files, preserves unchanged bytes and timestamps, and removes
only recorded obsolete inputs and their emptied parent directories. Unexpected
files, symlinks or incomplete recorded inventories reject. This keeps Cargo's
package paths stable while allowing its normal source/dependency invalidation;
there is no runtime-specific freshness exemption. A bounded record of removed
inputs retains only their last contents' hash and timestamp. Reintroducing those
same bytes can retain their timestamp, while staging any different contents at
that path discards the former record. This avoids rebuilding unchanged libraries
after an intervening library-only workspace without reusing timestamps from an
older version of changed code. Cargo still runs and validates every build.
The test helper's default slots use the same current-workspace staging; explicit
target directories retain separate graph-addressed staging.

## Build implementation ownership

`source_bundle.go` validates and materializes returned files; `cargo_manifest.go`
decodes the bounded manifest contract. `cargo_workspace.go` coordinates members
and selects artifacts, while `cargo.go`, `cargo_artifact.go`, `cargo_lock.go` and
`cargo_cache.go` own invocation, provenance, lock identities and cache admission.
Platform-specific cache lock files retain their operating-system constraints.
`build.go` owns immutable publication. `cargo_dependencies.lock` pins dependency resolution. The Cargo package plan
carries the source entrypoint.
