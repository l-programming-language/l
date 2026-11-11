# Building and running L programs

> **Core-only reset:** The current core-only build supports one standalone App whose body uses core Print and static Text. Build and run the emitted executable directly. Source-level Build, Run, Stop, Await, Execute, Error/Output inspection and other runtime-dependent top-level effects are unavailable and reject. The former runtime contract below is reconstruction context, not currently implemented support.

This guide describes the Rust-only generation and native CLI runtime path. L, Lyger and Logiceme are implemented in Go. Building L programs requires the pinned Rust 1.98.1 toolchain and retained dependencies available locally. Published binaries need no language compiler or L installation.

On native macOS or Linux:

```sh
l build main.l
./main
```

The compiler prints the published executable's absolute path on success. Building generates and compiles the program. It never runs its Apps or generated deployment scripts. Running the executable executes source instructions in order while Apps run in separate OS processes until joined with Await.

## Invocation and assets

[CLI options](cli/build.md) precede the source filename (default: `main.l` in the invocation directory). Use `--` for filenames starting with a dash. Extra arguments and unknown options reject. [`-os` is unimplemented](cli/build.md#target-os).

Embedded English resources work with relocated binaries, symlinks and `go run`, independently of CWD. Relative output paths use invocation CWD. [Developer overrides](development-harness.md#compiler-defaults) support compiler resource development.

Builtins are implicit through the source prelude, which includes core and loads
Storage privately. Bare dictionary paths such as `io` resolve only through
the selected bundle's `dictionaries.json`. Neighboring `io` or `io.l` cannot
shadow them. Explicit `.l` imports resolve relative to their importing source.
Imported dictionary helpers resolve relative to their resource provider and do
not read unrelated host files. Install instructions are in [README](../README.md).

Help, version and builds return 0 on success. Invalid invocations or unsupported hosts return 2, and build or output-write failures return 1. Diagnostics retain phase/cause on stderr. Build stdout contains the executable path and newline. Standalone [`l --version` or `l -v`](cli/build.md#version) prints the version without compiler initialization.

## Optional native build evidence

```sh
l build --native-evidence-output build-environment.json main.l
```

This explicit option records pinned tool identity/version observations, selected
source and backend fingerprints, relevant Cargo configuration, the actual Cargo
invocation and output digests, and compiled artifact hashes. It does not run the
published executable or any App. A failed Cargo invocation is retained in the
evidence document once initial environment capture has succeeded. On a successful compile, the coordinator replaces the requested evidence file
before publishing the executable. An evidence-write failure preserves the
previous executable and evidence. If publication subsequently fails, the
record still identifies the compiled artifacts.

Choose a destination outside source inputs, the executable, companion generations
and custom assets. The ordinary build path performs no additional observation or
analysis when the option is absent. The document is conditional provenance for
[contract checking](contracts.md#capture-native-environment-evidence), not approval
of native implementation behavior, allocator/panic policy or the current
runtime deployment.

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

Builds retain version-1 manifests and the `apps` inventory. App-constructed libraries add no companion entry. Standalone Service declarations add no build package.

Every build uses a unique staging directory on the output filesystem. All sources and binaries finish before publication. An explicitly marked companion directory holds immutable generations. The generation is renamed into place before the adjacent executable is atomically replaced. That last rename is the publication commit point. Failed builds preserve the previous executable and its generation. An already-published generation remains if the final rename fails. Old generations are not pruned because older running orchestrators may still use them. Concurrent builds cannot mix App generations. The last successful executable replacement wins.

The runtime retains its canonical executable directory before source execution.
Later Build and Run operations resolve their embedded generation paths against
that directory and revalidate each companion. This preserves a running
generation when Linux reports its replaced executable as deleted. An initial
location error remains a Build/Run error rather than failing runtime setup.

Unrelated existing output destinations and symlinked companion/generation directories are rejected. The compiler removes only its own invocation's staging on failure. A partial compilation never overwrites a prior executable.

Private staging cleanup is always attempted. If the build already failed, a
cleanup failure is added as an IO cause without replacing the original error.
After successful publication, temporary-directory cleanup is best-effort: it
cannot turn a committed executable replacement into a reported build failure.
Native artifact inspection descriptors close before publication. Close failures
reject the build and preserve any accompanying validation or copy error.

## Source generation and native builds

Lyger preserves canonical application books and projects each App and its additional constructed Services into workspace units. Logeme evaluates their unchanged AML. The separate synthetic orchestrator requests actual runtime support. Expanded source and the observable/backend AML views retain their distinct purposes.

The CLI accepts Rust source bundles only. Whole bundles are checked for matching dimensions, safe logical paths, duplicate/colliding entries and valid Cargo target/dependency relationships. The workspace contains an orchestrator binary, ordered App binaries and App-constructed Service libraries. Recursive Classes supply modules and Objects remain files. A leading slash is a logical bundle root, never a host absolute destination. Compiler-owned visibility and module assembly preserve declared ownership. No generated source rewriting is performed.

The lifecycle runtime is a library at `support/runtime`, linked into the
orchestrator under an allocated package name that avoids workspace collisions.
Its default policy comes from the [Runtime standard library](stdlib/eng/runtime.md).
One complete asset provider selects its implementation through the `runtime`
registry key. Replacements retain the current native contract and lifecycle
guarantees; source imports do not select another runtime for an App.
Resource generation produces [support.aml](../stdlib/runtime/generated/support.aml) and
[support.json](../stdlib/runtime/generated/support.json), containing fingerprints of the
root registry and actual source inputs, plus a dependency-selection index. The
manifest's `sha256` binds that entire inventory and index to the AML bytes.
Rehashing or omitting a source entry alone cannot make stale support current.
This is bundle integrity, not authentication of a forged manifest. Explicit
native builds also recompose Runtime in memory for external providers and reject
AML or metadata that differs. This adds frontend work, without repairing assets.
The immutable embedded distribution retains its validated selection path.
Library callers can request the same check with `Lyger.CheckRuntimeResources`.
The raw AML backend does not parse L. Offline inspection validates recorded
integrity without recomposition or repair. Changed assets require explicit
regeneration through the [runtime authoring workflow](stdlib/eng/runtime.md#select-a-replacement).
The derived index adds no AML fields or logeme interfaces.

With unmodified embedded definitions, the CLI selects runtime declarations from
that index using the orchestrator's imports, retaining shared type and host
requirements. When selection is unavailable, for example with explicit resource
providers, user-authored native definitions or unrecognized runtime imports, it
uses the selected provider's full composition. Missing or stale resources reject without
falling back to embedded files. Logiceme renders complete framed AML in either
case. The native entry and host/process inputs live under `stdlib/runtime/rust/`. Their
relocation does not change generated native filenames or module topology.

The runtime builds with the workspace and needs no separately installed launcher.
Cargo can reuse it when only the entry changes and the library's actual inputs
remain unchanged within the locked build slot.

Cargo builds every workspace package with Rust 1.98.1, `--frozen`, offline resolution and native target selection. The coordinator isolates caller target/compiler flags and uses retained pinned dependency resolution. It does not fetch arbitrary new dependencies. Builds supervise child process groups and retain their existing deadlines. A private versioned cache retains compiler products, with concurrent builds serialized through artifact snapshotting so publication cannot pick up another source's executable.

Local packages may share names with retained registry packages. Lockfile references
use the actual package name, version and, when needed, registry source to distinguish
them. Dependency aliases do not change package identity. The coordinator preserves
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
package, Book, App metadata and primary target receive the same internal name.
Canonical L App IDs, library identities, source topology and publication paths
remain unchanged. Native Cargo-name introspection sees this internal alias.
Direct Logiceme inputs retain their supplied names. Published companions still
use ordinal filenames and canonical IDs, so the entrypoint and its adjacent
companion directory can be relocated together.

The coordinator independently rejects duplicate binary targets under
case-insensitive comparison before creating the build workspace. Library-only
members occupy no binary targets. Case-equivalent package names are allowed when
their binary targets are distinct.

Parser checks call the same local-FlagSet parser as the CLI. Native build checks
use the production workspace coordinator, including single-orchestrator cases.
Cache cancellation checks exercise its bounded cache admission directly. No
separate test-only compilation or lock-wait adapter remains.

Structured Cargo messages must identify the requested manifest and binary target. Verified native executables are copied into the invocation's private staging before publication. The compiler does not guess output paths, launch Go as an intermediary, or ship its build toolchain. Generated deployment scripts are artifacts and are never run by `l build`.

Native build caching uses six bounded slots. Complete input paths and bytes
select a preferred slot; another available slot can admit the build when it is
busy. Admission waits cancellably only while all slots are occupied. Each slot
owns a cross-process lock, exact source copies and Cargo dependency products.
Every build invokes Cargo, verifies manifest/target identities and snapshots
executables before releasing its lock. The cache substitutes for neither tests
nor publication. Concurrent workspaces cannot expose each other's mutable
executable paths.

CLI slots and native-test shards reconcile an explicit input inventory at stable
private `StageCurrent` paths. They validate the supplied tree and recorded paths
before changing files, preserve unchanged bytes/timestamps, and remove only
recorded obsolete inputs and emptied parent directories. Unexpected files,
symlinks, incomplete inventories and malformed state reject. Stable package paths
retain Cargo's normal source/dependency invalidation without runtime-specific
freshness exemptions.

At most 4096 absent-path hints retain only the last contents' hash and timestamp.
Reintroducing identical bytes can restore that timestamp; staging different
contents discards the former hint. Hints are consumed before writes and cannot
restore aliases, future timestamps or stale versions. Missing history causes a
rebuild. This preserves unchanged libraries across an intervening library-only
workspace without reusing older versions of changed code. Cargo still runs and
snapshots artifacts every time. The test helper's default slots share this
staging; explicit target directories retain separate graph-addressed staging.

## Service composition

Services organize source within App construction. The builtin Basic descriptor
still composes a Main Service, Object and Function. Additional Services
constructed by an App remain separate library packages with their existing
public interfaces. The primary Service retains the App's binary topology.

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

Each App runs in a separate OS process. The following invocation, result, cancellation and cleanup rules apply to Apps. `Run` returns after successful process creation or recording terminal startup failure. It does not await readiness, process exit, capture completion, or a user continuation. Internal Tokio supervision/capture tasks are permitted. Launch acknowledgment remains part of the ordered runtime operation.

`Start` is ordinary source composition Build then Run. `Execute` is Start then Await. Both retain their `any App` source signatures. Run revalidates after Build, including when Build failed. Native coordinated operations process operands in source order. A source-defined action expands its entire body for each operand before the next operand. Aliases share one canonical target and invocation slot. Start and Execute are absent from primitive-root vocabulary.

Imperative Await, outer passive Await, and applicative Await in source actions execute. Native Await in an App body reports unsupported ownership. Standalone Service body effects reject. A standalone applicative statement and the stray CONDITIONAL suffix are invalid. Existing Stop forms remain recognized.

### Invocation and results

One invocation remains outstanding until its first Await continuation finishes. Completion alone does not release the slot. Await joins process reaping, capture finalization, and owned-group cleanup, then returns the immutable invocation result. Continuations execute inline exactly once, including failed launch, nonzero exit, and per-target Stop. Public Await on global shutdown aborts source execution. It never returns to the continuation.

ClaimContinuation marks an invocation active before entering its body. Cached/self-Await returns the same result and cannot reclaim it. Nested Await of another App is allowed. Lexical paragraph identity cannot suppress another target's claim. FinishContinuation releases the slot on normal completion. Run of the same target during its own continuation is rejected, as are Start/Execute for that App. Continuation Exit exits the outer orchestrator, including from expanded source actions. Fresh accepted invocations get fresh continuation locals.

Final Output is raw stdout, including NUL, whitespace, and Unicode. Stderr is captured separately and never automatically forwarded. Stderr alone does not imply failure. Reading Output before final completion is a fatal lifecycle misuse, including unused assignments. There is no implicit join. Finalized Output may be read before Await without activating a continuation. Retained invocation snapshots and scalar copies survive reruns unchanged.

Build alone creates no invocation. Run startup failure creates an awaitable failed invocation and performs no OS Wait on an unstarted command. Await/Stop before Run, duplicate Run/Start, Build while outstanding, and Output-unavailable are fatal source misuse with source attribution. Ordinary child failures are stored results, not implicit orchestrator Exit.

Build selects its operation error. Stop signal-delivery failure selects an operation error visible through the target's Error member, remains recoverable, and preserves the invocation and ownership. A later Await, including cached Await, reselects that invocation's immutable result. A completed Stop is an idempotent no-op. Stop captures its invocation before signaling and cannot signal a replacement invocation after a race.

### Shutdown, streams and limits

Stop requests SIGKILL of the owned process group and returns after request submission/known completion. It is not a join. Natural exit and Stop may race, but only one coherent terminal result is published. Leader exit while capture/group cleanup is pending is not finalized completion.

On macOS, final group signaling can return `EPERM` during group exit. After the first successful Stop or shutdown termination request, finalization can observe group disappearance using only the time remaining in the existing two-second bound, measured from that attempt's start. Later signals do not renew the window. Only `ESRCH` from signal zero confirms disappearance. A live group, another error or an expired window preserves the cleanup failure. This sends no additional termination signal and performs no additional Wait.

Normal fallthrough and bare Exit clean all owned processes, perform no unawaited continuations, and exit 0 if cleanup succeeds. Explicit Error/fatal misuse cleans up then exits 1. SIGINT/SIGTERM abort source flow, settle ownership, and exit 130/143 respectively, including between launches or while Await is pending. Preserve an explicit primary error and append cleanup failures once in launch order. Cleanup failure on otherwise successful exit yields 1.

Generated main invokes runtime finalization, which performs cleanup before closing normal terminal admission. A fatal or signal request accepted during that cleanup must finish exiting before main can return. `Shutdown` remains an idempotent, nonterminal cleanup operation for internal callers. Successful normal finalization returns normally without converting a native panic into success.

Broken stdout/stderr pipes must trigger cleanup and exit 1, including descriptor-1/2 SIGPIPE behavior. Error reporting must not recursively write to a failed diagnostic writer. Tests use subprocess exit behavior, never a fake exit callback. Cleanup is explicit on terminal paths. Unwinding alone cannot cover process exit.

When fatal and signal requests compete, the first accepted terminal request selects the exit status and primary diagnostic. Later requests do not replace that decision during cleanup. Failure to write the diagnostic still selects status 1 and never retries that writer.

Keep the existing two-second post-exit/cancellation pipe-drain bound. There is no normal process-duration timeout, no process-count cap, and no capture-size cap. Fixed stress counts are test workload only. Every launched direct child has exactly one Wait owner. Same-group descendants must terminate even if they close inherited pipes. Escaped process groups and uncatchable orchestrator termination (SIGKILL, host/power loss) are documented limits, not skipped required feature tests.

Children inherit environment and caller CWD, receive EOF on stdin and receive no synthesized arguments. Runtime Build checks the prebuilt executable. It invokes no build tool. Show writes Text exactly and formats supported runtime values without adding a newline or replaying unrelated captured output. Direct App and Error display rejects.

The lowerer checks scalar/operator types and definite assignment across continuing branches. Exit removes a branch from that intersection. Output availability remains a runtime check. Diagnostics retain source attribution. Maintained programs import the `IO.Print` member from the `io` dictionary and use implicit Basic, Start and Execute definitions. Direct Run requires an explicit Await wherever subsequent source needs completion or continuation effects.

Phase 2 in-code concurrency, runtime argument transport, nested layout syntax, containers, cross-compilation, remote deployment and unrestricted installation of third-party dependencies remain outside this build command. The [architecture](architecture.md) describes compiler/runtime ownership and [development guide](development-harness.md) describes validation.

## Build implementation ownership

`source_bundle.go` validates and materializes returned files. `cargo_manifest.go`
decodes the bounded manifest contract. `cargo_workspace.go` coordinates members
and selects artifacts, while `cargo.go`, `cargo_artifact.go`, `cargo_lock.go` and
`cargo_cache.go` own invocation, provenance, lock identities and cache admission.
Platform-specific cache lock files retain their operating-system constraints.
`build.go` owns immutable publication. `cargo_dependencies.lock` pins dependency resolution. The Cargo package plan
carries the source entrypoint.
