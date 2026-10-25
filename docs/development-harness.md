# Development and acceptance

Use the repository harness for current-input integrity and the ordinary build/test gate. The [architecture](architecture.md), [build/runtime contract](cli-build.md) and [error model](error-model.md) describe supported behavior.

## Temporary validation pause

Expensive local validation requires an explicit author request. Repository Actions
automatically validate Linux when a pull request opens, receives commits, or reopens.
Manual dispatch also supports macOS. The required `L CI` check on `main` remains
removed; pull-request, deletion and force-push protections remain.

Do not automatically run local bootstrap, full-module lint, full ordinary/native
suites, or `scripts/l-check`. Run them only when explicitly requested by the user.
Hosted runs are authorized by the PR events above or an explicit dispatch request.
This policy overrides completion commands elsewhere in
repository documentation and skills, including documentation-only changes.
Use lightweight inspection appropriate to the change, and state which validation
was not run. Historical green reports do not establish validation of a new revision.
Tests, assertions, protected inputs, tool pins and command behavior are retained.

Restoring the required branch check remains a separate explicit decision.

## Full validation when requested

When explicitly requested, run `scripts/l-ci bootstrap` if setup is needed, then
`GOMAXPROCS=4 scripts/l-check`.
The completion command runs workflow linting, both Go lint profiles, harness
`doctor` and unfiltered `verify`, and the complete tagged native compilation
stage. Every stage must succeed with fresh, complete evidence. Focused tests
remain useful during development but do not establish completion.

Tool versions and verified official lint archive checksums live in
[.ci-tools.json](../.ci-tools.json); [.go-version](../.go-version) selects Go
1.27.1 and [.ci/go-checksums.txt](../.ci/go-checksums.txt) pins its fallback
archives, without changing the module's Go 1.20 language declaration. Rust remains
1.98.1, golangci-lint is 2.14.0, and actionlint is 1.7.12. Lint tools, fallback Go,
Go build caches, seed projects and run evidence stay in ignored repository
storage. Rust toolchains and Cargo/Go dependency downloads use their standard
host locations; the CLI also retains its external native build cache. Tool
updates are explicit reviewed changes.

[scripts/l-lint](../scripts/l-lint) checks the workflow and the entire Go module,
including tests, in separate ordinary and `nativecompile` passes. The
[lint configuration](../.golangci.yml) enables strict `errcheck`, `govet` (excluding
`inline`), Staticcheck `SA*`/`S*`, `ineffassign`, `unused`, `nolintlint`, and standard
`gofmt` without simplification. Resolve every warning and error. There is no
changed-lines baseline, blanket test exclusion, output cap or automatic CI fix.
Specific, explained suppressions are reserved for proven invariants and
intentional best-effort behavior; they must not hide meaningful errors.

The shared stage launcher supports `scripts/l-ci lint`, `scripts/l-ci acceptance`
and `scripts/l-ci native` for investigation and independent CI jobs. The default
`scripts/l-check` always runs all of them. It preserves harness interpretation
and rejects missing, malformed, stale or incomplete reports and native outcomes.
See the fresh ignored run directories under `.harness/runs` for logs and
structured results; cached reports or previous successful test outcomes never
stand in for a new run.

Bootstrap separates dependency preparation from validation. It verifies the native C compiler,
installs pinned tools, fetches Go dependencies without changing module files,
and constructs four disposable Cargo seeds from retained runtime, Lucid plain,
Tokio and selection inputs. Each seed passes `fetch --locked` and then
`fetch --frozen`; every external dependency tuple in retained lockfiles must be
covered and copied/original locks must remain unchanged. Setup fails with an
explanation for uncovered dependencies.

Setup compiles the pinned Go race standard library and project libraries with
`go build -mod=readonly -race std` and `go build -mod=readonly -race ./...`, using
the downloaded modules offline. Acceptance still creates and compiles its own
race-instrumented API probe and executes every assertion. Setup also compiles
selected locked Rust registry dependencies. Rust dependency build scripts run
as part of that compilation; Rust workspace targets, generated Apps and test
binaries are not built or executed by this preparation. Verified Rust dependency
products populate 25 existing caches: six CLI slots, four shards each for Lyger,
Logiceme, `cmd/l` and `acceptance/phase1/runtime`, and the three fixed runtime/empty-error
caches listed below. This moves dependency compilation outside acceptance timers
without running warmup tests or recording test success. The actual package build,
unfiltered tests and native compilation still run afterward with their original
assertions, fixture bytes, offline resolution and the documented budgets. Bootstrap never runs
deployment scripts, filters acceptance tests or changes retained locks.

## GitHub Actions

[CI](../.github/workflows/ci.yml) is enabled and automatically validates Linux
(Ubuntu 24.04 x64) on PR creation, new commits, and reopening. There is no separate
push trigger or run for title/description edits. Manual dispatch defaults to
`linux`; `macos` selects macOS 15 ARM64, and `both` selects both platforms. Each selected
platform prepares once and runs `scripts/l-check`: formatting, workflow lint,
both Go lint profiles, doctor, unfiltered acceptance, and Expected/Generated native
compilation. Hosted checks set `CARGO_INCREMENTAL=0`; local runs retain their
configured Cargo setting. This changes compilation reuse without changing test
selection, compiler concurrency or deadlines. A stage failure stops later stages on that platform; the other
selected platform continues. The stable aggregate `L CI` requires every selected
platform to succeed and rejects failures, cancellations and skipped jobs.
Linux-only success supplies no Darwin lint or runtime evidence. Superseded runs
are cancelled. The combined job limit is 120 minutes; the aggregate retains
5 minutes. The [finite budget policy](#validation-budgets)
preserves runtime watchdogs.

Actions retain reviewed full commit pins, read-only repository access and
checkout without persisted credentials. Bootstrap installs both pinned linters;
the shared launcher verifies archive checksums, executable bytes and versions
before execution. Weekly Dependabot updates review action pins; tool versions
remain separate explicit updates. This
workflow validates code and does not publish releases or deploy applications.

Caches separate pinned tools/dependency downloads from Go/native compiler outputs
shared by all stages on a platform, including the CLI's external native cache.
Keys identify platform, architecture, toolchains and dependency inputs; build
keys also identify the source tree. Build caches use namespace `v4` without a
stage component; dependency caches retain `v2`. The shared build cache includes
Go build, golangci-lint, Lucid and package parity outputs, plus the fixed
`rust-runtime-parity`, `rust-empty-error-runtime` and `rust-empty-error-frontend`
cache directories and their staged `-inputs` siblings. Each cache is saved once
per successful platform job under
the existing trusted-write restrictions. Reports and test outcomes are never cached.

Current evidence is uploaded on success or failure as one artifact per selected
platform, with unique names and 7-day retention. Each retains `.harness/runs/`
and the build/runtime stdout, stderr, JSON and `.l` source snapshots under
`.harness/phase1/cli/`, excluding executables and compiler caches. Its downloaded
artifact therefore contains `runs/` and `phase1/cli/` directories. Hosted
workspaces are fresh, and none of these evidence directories is restored from a
cache. Cold-cache dispatch bypasses restores and saves while retaining bootstrap
and the same checks. Existing remote caches and artifacts expire normally.
A cold run must pass without warmup tests, relaxed deadlines or retries that hide
its original failure.

The [implementation ledger](../testdata/integration/IMPLEMENTATION.md) records
validation evidence and its limits. The [native runtime contract](#native-runtime-validation)
requires observable process and resource cleanup.
Setup, lint and tagged compilation alone do not establish runtime acceptance.

`main` still requires a pull request, but its required `L CI` status check has
not been reinstated. Existing deletion and
force-push protections remain in force; merging and public release are separate
steps.

## Commands

Run from the repository root with Go and the pinned Rust toolchain/dependencies available:

```sh
scripts/l-harness doctor
scripts/l-harness verify
scripts/l-harness diagnose
scripts/l-harness diagnose --frontend
scripts/l-harness diff EXPECTED ACTUAL
```

| Command | Purpose |
| --- | --- |
| doctor | Check environment, current protected inputs, required absent paths and maintained-document links |
| verify | Check integrity before/after the build and both unfiltered test phases |
| diagnose | Run selected compiler cases individually and collect explicit vet diagnostics |
| diagnose --frontend | Add temporary compiler/import probes to diagnostics |
| diff | Compare supplied files byte-for-byte, reporting lengths, hashes and first differing offset |

The launcher builds the standard-library-only harness before loading compiler packages; [gate helpers](../tools/internal/testgate/testgate.go) are shared with CI. It may be invoked by absolute path from another directory; relative diff arguments resolve against the caller directory. Missing Go, bootstrap-build failures and argument/contract-loading errors can occur before a report is allocated.

Verification builds the module, then discovers packages with `go list -mod=readonly -json ./...`. Each package runs exactly once in two sequential, unfiltered phases: `tests-native` contains all `acceptance/` packages, `cmd/l`, `lyger`, `logiceme`, `internal/cargocache` and `internal/testsupport/native`; `tests` contains the rest. A phase is omitted only when its inventory is empty. Ordinary Go test discovery and the separate tagged native stage remain unchanged.

```sh
go build -mod=readonly ./...
go list -mod=readonly -json ./...
go test -p=N -parallel=P -mod=readonly -count=1 -timeout=120s -json ORDINARY_PACKAGES...
go test -p=N -parallel=P -mod=readonly -count=1 -timeout=10m -json NATIVE_HEAVY_PACKAGES...
```

`N` is `min(3, effective GOMAXPROCS)` and `P` is effective `GOMAXPROCS` from the harness runtime snapshot. These preserve the existing package cap and Go case default. `-p` bounds each outer test command's builds and test binaries; inherited `GOMAXPROCS`, nested builds and application settings are unchanged. The [budget policy](#validation-budgets) distinguishes phase deadlines from runtime watchdogs.

Native acceptance cases may run concurrently when projects, resource overrides,
provenance records and process groups are independent. Per-package admission is
bounded; each case retains its internal signal, publication, resource-deletion
and repeated-output sequence. Compilation and runtime use explicit separate deadlines.
Test scheduling does not change L execution semantics.

Verification permits no overlays, skips, test selectors or disabled vet. Reports require every discovered package exactly once, matching phase arguments and successful terminal outcomes. Packages without test files may report a package-level skip; named skips, failures, aborted tests, malformed/incomplete streams and missing outcomes fail. All 16 original obligations must pass across the combined phases. Focused diagnostics and explicit vet cannot replace this gate.

Exit status is 0 for success, 1 for build/test/diagnostic findings or byte mismatch, and 2 for integrity, configuration, setup or usage failure. Doctor checks prerequisites and does not establish behavioral acceptance. Before the Go harness starts, shell failures retain their native status; a missing `go` command can exit 127 without a JSON report.

## Acceptance timing diagnostics

Set `L_HARNESS_TIMING_DIR` to an absolute ignored run directory only for diagnosis.
The existing build and native-test boundaries record per-process JSONL spans for
admission, frontend/workspace setup, cache identification/wait/hold, Cargo,
provenance, publication, runtime observations and cleanup. Start records retain
interrupted spans; completed durations use a monotonic clock. Diagnostic I/O is
best effort and never enters compiler or application streams. Correlate PID/PPID,
subjects and timestamps; concurrent work-time sums are not elapsed command time.

The shared compiled test Cargo observer invokes the pinned real Cargo for every
build, with private configuration and fresh evidence per probe/invocation. It
preserves arguments, streams, status, process-group ownership and artifact
snapshots; it removes the per-build shell/helper launch chain, not Cargo work.
Instrumented runs separate actual compilation from metadata and recording,
count fresh/rebuilt artifacts, and retain Cargo's supported HTML timing reports.
Diagnostic flags and trace I/O affect measurements, so run completion gates with
the timing variable unset. No cache purge, program pre-execution or cached test
outcomes are part of this workflow.

Tests reuse invariant compilation setup in shutdown, F20 signal, retained-output
and installed-CLI cases. Each runtime profile still receives fresh private execution
files/state with verified artifact bytes. Source/resource/publication/cache
invalidation scenarios still rebuild independently. The
[implementation ledger](../testdata/integration/IMPLEMENTATION.md) records measured
results; measurements do not redefine the current budgets or authorize concurrency changes.
The retained reference profile uses `GOMAXPROCS=4`, `-p=3`, `-parallel=4`, four
native-case permits for distribution/language/CLI and two for other Phase 1
packages. The CLI's six Cargo-cache slots are unchanged. Recorded timings apply
to their host and installed toolchain, not a cold-cache guarantee.

Use [Cargo's HTML timings](https://doc.rust-lang.org/cargo/reference/timings.html)
for compiler units and external spans for process/observer startup. Retain the
[structured artifact and versioned metadata evidence](https://doc.rust-lang.org/cargo/reference/external-tools.html)
on every build, including fresh artifacts. Subprocess diagnostics must respect
[Go's cancellation and pipe-draining contracts](https://pkg.go.dev/os/exec);
neither parent completion nor closed pipes establish descendant termination.
The Darwin zombie case is supported by the owned-process regression and
[Apple's group-signaling implementation](https://github.com/apple-oss-distributions/xnu/blob/main/bsd/kern/kern_sig.c).
It does not authorize treating general permission failures as absent groups.

## Immutable build evidence and helper ownership

The harness assigns a fresh `L_HARNESS_RUN_ID` to its command environment and
records it in the report. An inherited run ID cannot cause a new harness run to
reuse evidence. Standalone test processes allocate their own namespace. Ordinary
bundle records preserve all returned artifacts, log their hashes and never
replace prior records; `L_HARNESS_EVIDENCE_DIR` still selects their root.

Set a global `L_HARNESS_CARGO_EVIDENCE_DIR` only for controlled rebuild diagnosis.
Tests that require artifact provenance also enable their own private captures.
Detailed captures retain exact staged inputs before Cargo, input hashes and path/cache
identities, an allowlist of compiler settings, raw Cargo messages and diagnostics,
artifact hashes and completion state. Copies are independent evidence files;
they never replace source inputs or executable artifacts. Missing or incomplete
evidence invalidates the diagnostic comparison without replacing Cargo's status
or altering its streams. Leave global diagnostic capture and timing variables
unset for completion gates; required test-local provenance capture still runs.
Enable timing alongside detailed capture to retain the CLI parent's actual
preferred/selected cache-slot event.
A missing event is reported as unavailable; the observer does not guess it.

The native helper holds a cross-process file lock from source reconciliation
through Cargo and immutable executable snapshots. Waiting consumes the existing
build budget. The lock remains a regular file inside excluded cache state and is
never unlinked on release. Cargo's own build lock does not protect the earlier
staging operation. The helper preserves discovery stdout, stderr and native
errors separately, and reads Cargo's line-oriented records without discarding
later artifacts after unrelated output.

The CLI slots and default native-test shards reconcile each complete source tree
at a stable `StageCurrent` path. Unchanged bytes retain their timestamps.
Reintroduced bytes matching a path's latest retired hash may retain its timestamp;
metadata for at most 4096 absent paths stays outside the source tree. History
stores no source bytes; the supplied tree remains authoritative. Incoming files
consume old hints before writes;
filesystem aliases and future timestamps cannot restore stale versions. Missing
history costs a normal rebuild; malformed state fails closed. Every build still
invokes Cargo and snapshots its actual artifacts. See the [cache contract](cli-build.md#source-generation-and-native-builds)
for identity and ownership requirements.

Use `CARGO_LOG=cargo::compiler::fingerprint=info` during a controlled rebuild to
record its reason; Cargo cannot reconstruct that explanation afterward. Compare
exact A/A and A/B/A inputs using the actual stable-staging path and private slots.
Retain seeding work separately and never normalize embedded source paths, purge
user caches, or count summed concurrent work as elapsed time. These observations
do not themselves authorize a different cache policy, admission limit or profile.

## Rust source-generation verification

The current [Rust suite](../logiceme/testdata/lucid/README.md) covers
selection, iteration, declaration, invocation, callable, trait, empty-composition,
keyed-pattern and generic/lifetime scenarios. The suite guide owns the current case inventory and counts. Normal
verification runs exact-bundle comparisons, malformed-input and focused checks,
and the 18-template freeze check. Native Rust
compilation is separately tagged and is **not** part of the untagged acceptance
command. When explicitly requested, `scripts/l-check` runs both the gate and this stage; ordinary frontend/runtime parity tests also compile and execute their controlled Rust programs:

```sh
GOMAXPROCS=4 GOCACHE="$PWD/.harness/cache/go-build" GOENV=off GOFLAGS= GOWORK=off GOTOOLCHAIN=local \
  go test -mod=readonly -count=1 -tags=nativecompile \
  -run '^(TestLucidNativeCompile|TestNativeMetadataArrowsCompile)$' -timeout=30m -json ./logiceme
```

Use the suite's retained locks and pinned Rust 1.98.1 toolchain; communication
uses Tokio =1.53.1 with rt/sync features; selection adds macros in a separate
manifest and retained lockfile. Keep frozen dependency resolution
and three-minute subprocess limits. The stage compiles independently authored
expected projects and actual returned projects, without source repair, formatting,
oracle substitution, lock regeneration or dependency fallback. It does not run
Rust applications or generated scripts. See the suite guide for artifact retention
and the [dated results](../logiceme/testdata/lucid/FAILURES.md) for evidence.

The [Service boundary coverage ledger](../testdata/service-execution/coverage.md) maps typed builtin selection, App-only lifecycle, structural standalone Service rejection, retained App-constructed libraries and raw AML library/native-consumer controls. Exact generation, independent expected-Rust compilation, generated-Rust compilation and supervised execution remain separate evidence.

The maintained cases establish only their documented compositions. Local
`Lifetimes`/`GenericTypes` inputs provide native declaration/argument text, not
ownership inference or general generic-implementation support. Retained
runtime/module support and unrelated advanced validation paths are not verified
by this corpus; consult the
[integration limits](../builtins/README.md#current-integration-limits).
Do not restore the removed Rust matrix or change frozen definitions merely to
obtain a passing run. Future interface or coverage changes need an explicitly
scoped task; protected bytes still require a finite approved amendment.

## Current integrity and amendments

[harness/contract.json](../harness/contract.json) protects current test sources, expected-value helpers, inputs, case oracles and canonical resources with explicit approved hashes. It records required absent legacy paths and the original named-test obligations. Current source membership is derived from the protected source paths, not an old migration ledger. Expected helpers outside `_test.go` are equally protected. Adding a current fixture must update the current contract through review, not revive a historical migration manifest.

Never regenerate hashes or expected values to make an implementation failure pass. Preserve raw expected streams/statuses, actual behavioral assertions, canonical grammar/vocabulary and source-defined builtins. A fixture/oracle amendment requires explicit authorization identifying its exact scope and reason. Historical exceptions grant no standing permission to change expectations. Source expansion preserves declaration targets, typed slots and order; compiler output must follow source semantics rather than fixture-specific substitutions.

The October 25, 2026 organization migration changes only the Go module/import
prefix from `github.com/jecsand838/l` to `github.com/l-programming-language/l`.
The ten protected files containing that prefix receive matching import-only
changes and updated hashes. All behavioral assertions, expected output bytes,
canonical resources, test obligations and the other 376 protected hashes remain
unchanged; this amendment does not authorize behavioral oracle changes.

The harness checks current files directly. Ordinary build, tests, doctor and verify do not require archived documents, attachments, historical baseline reports, a Git object lookup, or a network fetch for provenance. Initial Go dependency setup can still need network access. Markdown link checking discovers maintained documentation; it does not preserve a closed list of obsolete plans. Repository skills remain under `.agents/skills`.

## Reproducible execution and evidence

The launcher and child commands use `GOENV=off`, empty `GOFLAGS`, `GOWORK=off`, `GOTOOLCHAIN=local` and `.harness/cache/go-build`. Persisted Go settings cannot inject overlays, filters or alternate tooling. Reports record current revision/dirty state when available, resolved Go executable and relevant environment; missing Git history does not invalidate a source-export run. No user-level Go settings are changed.

The report's environment fields distinguish inherited configuration from the harness's effective runtime snapshot:

| Field | Meaning |
| --- | --- |
| `original_gomaxprocs` | Exact inherited `GOMAXPROCS` string; omitted when unset, present with an empty string when explicitly empty |
| `effective_gomaxprocs` | Harness process's `runtime.GOMAXPROCS(0)` snapshot, used to select the verification budget |
| `num_cpu` | Harness process's `runtime.NumCPU()` value |
| `test_package_parallelism` | Selected outer `-p` budget, present only in verification reports |
| `test_case_parallelism` | Explicit `-parallel` budget matching the effective runtime snapshot, present only in verification reports |
| `build_command_timeout_ms` | Canonical verification build deadline: 180000 milliseconds |
| `package_discovery_timeout_ms` | Package inventory deadline: 60000 milliseconds |
| `test_package_timeout_ms` | Ordinary verification test-binary deadline: 120000 milliseconds |
| `test_command_timeout_ms` | Complete ordinary verification test-command deadline: 360000 milliseconds |
| `native_test_package_timeout_ms` | Native-heavy test-binary deadline: 600000 milliseconds |
| `native_test_command_timeout_ms` | Complete native-heavy phase deadline: 900000 milliseconds |

These values do not measure every child process's effective concurrency. The executed command arguments record the selected `-p` and `-parallel` values directly.

Harness and CI commands record effective `timeout_ms`, including defaults. On macOS/Linux, cancellation and completion clean the owned process group, including descendants that closed inherited pipes. Command status and captured streams remain visible; unexpected cleanup errors join the error. Reports distinguish deadline, cancellation and post-wait signals, retaining PID/PGID, raw errors and monotonic durations. On Darwin, post-wait `EPERM` after successful deadline `SIGKILL` may indicate an unreaped zombie group. Only then may signal-zero probes establish disappearance within the remaining WaitDelay budget. Reports retain raw `EPERM` and record verified `ESRCH` separately. Persistent errors fail, signals are not resent, and the original timeout remains a failure. Probe success requires its named outcome and report, not an empty test-selector match.

An explicitly supplied setting such as `GOMAXPROCS=4 scripts/l-harness verify` has a broader effect than the outer `-p` limit: inherited runtime settings also affect test binaries, nested Go builds and application processes. Report that setting with the result; a passing run under one configuration does not erase failures from another.

Reports live in ignored `.harness/runs/<UTC-time-random>/`. Schema-2 verification reports require `test_packages` inventory/phase assignments and package terminal outcomes alongside named test events. `report.json` names command logs and raw events; JSON stdout and `test_events_path` may share a file. Readable test output and stderr remain available. Frontend reports live under `frontend/compiler/` and `frontend/imports/`, with inert `.go.txt` sources overlaid only for diagnostics. Follow report paths rather than assuming filenames.

The CI acceptance launcher retains a canonical report even when its child fails,
after validating its schema, invocation identity, timestamps and repository-local
path. It preserves the child failure and any report-validation or copy errors.
Malformed, stale or escaped report paths are rejected, never substituted with
earlier successful evidence.

Retain the latest acceptance report and raw evidence cited by the current change. Uncited superseded runs and disposable binaries/copies may be removed after preserving unique work; never prune running work or rewrite retained reports. Go owns its cache. An explicit cache purge uses `go clean -cache` with the repository GOCACHE and makes later builds slower.

Report the command status, failed package/test names or complete success, input-integrity result, run location, host and limitations. Diagnostic views may normalize incidental UUIDs/map ordering, but raw acceptance bytes and source instruction order remain unchanged.

## Native runtime validation

Native tests require permission to build/launch executables, bind local Unix sockets, observe processes, signal/reap owned groups and run the race detector. A constrained sandbox can pass doctor while blocking those operations; report that as an environment limitation, never weaken assertions or treat it as a passing gate.

A cross-host claim requires separate native macOS and Linux gate records. Go compiler/API/helper race checks and actual Rust lifecycle/concurrency observers establish distinct properties; no Rust sanitizer equivalence is claimed. Use the complete [current suite](../testdata/integration/README.md) and generated shutdown coverage. Generation-only or cross-compilation evidence does not prove native process behavior; no old host observation substitutes for a current run.

Apps have no normal-duration timeout, process-count cap or capture-size cap. Runtime pipe draining after exit/cancellation retains its two-second bound. Same-group descendants are owned; deliberately escaped groups and uncatchable orchestrator termination are outside the guarantee. Test watchdogs diagnose missing events and never redefine these production semantics.

Native tests use the real compiler/runtime and current case manifests, source fixtures and controlled companions. Compare exact stdout/stderr/status and observe process lifetime/Wait ownership where required. The controller records leaks before fallback cleanup; cleanup never turns a failed runtime case into passing evidence. Generated deployment scripts are never executed.

The socket observer accepts drain EOF or `ConnectionReset`. A missing release
socket or release `BrokenPipe`/`ConnectionReset` still requires `stopped()` to
confirm disconnection and lifetime-lock release under the existing deadline;
other I/O failures remain failures.

## Investigating failures

Trace the earliest failed boundary: selected resources and source origins; imports and literal identity; token roles/storage/value types; grammar case and consumed span; scoped records/bindings; then ordered IR, AML and source bundles. Package tests retain package-local working directories. Temporary probes may expose private state without altering protected tests.

For a backend failure, preserve the exact input and distinguish observable AML from canonical framed backend AML. Establish a failing regression before repair; reuse an existing exposing test when available. Use `TestDemoGeneratesBuildableSource` and `TestMemberSourceGeneratesBuildableSource` for frontend handoffs, the [Service coverage ledger](../testdata/service-execution/coverage.md) for App-constructed libraries, raw AML native consumers and App process boundaries, and the stored Lucid complete AML cases for the Rust API boundary. Parser acceptance, source emission, native compilation and execution are distinct claims. The CLI/compiler diagnostics must stay separate from application stdout.

Harness self-tests use disposable synthetic files/modules:

```sh
GOCACHE="$PWD/.harness/cache/go-build" GOENV=off GOFLAGS= GOWORK=off GOTOOLCHAIN=local \
  go test -mod=readonly -count=1 -timeout=60s ./tools/harness
```

## Rust verification boundaries

Frontend library selection, native atom encoding, typed orchestration and CLI
builds use Rust. L, Lyger and Logiceme remain Go programs. The
[implementation ledger](../testdata/integration/IMPLEMENTATION.md) records current
outcomes; the [current integrity contract](#current-integrity-and-amendments)
protects their inputs and original required outcomes.

The backend Lucid corpus remains compile-only. Supervised frontend/runtime tests
exercise actual generated Rust binaries, Cargo artifact provenance, native cleanup
and standalone relocation. They never run deployment scripts. Report exact
generation, native compilation, runtime and full acceptance separately, with the
first real failure and downstream blocked phases retained.

## Validation budgets

Both unfiltered ordinary and native-heavy test phases are mandatory. The relocation
fixture reuses the package's compiled CLI; package discovery, test identities and
assertions remain unchanged.

| Boundary | Approved budget |
| --- | --- |
| Package discovery / canonical build | 60 / 180 seconds |
| Ordinary package / phase | 120 / 360 seconds |
| Native-heavy package / phase | 10 / 15 minutes |
| Verification launcher / Actions check job | 26 / 120 minutes |

Verification reserves discovery + build + both phases + one minute for reporting.
Suite caps guard cumulative work, not application performance; they do not promise
every child can exhaust its individual allowance. Doctor retains seven minutes;
direct Cargo, diagnostic and tagged-native deadlines, runtime watchdogs, tool pins
and all concurrency remain unchanged. These budgets also govern cache-disabled
and local runs. Further deadline amendments require explicit scope.

The 40-second prerequisites cover installed CLI and Go probe/helper compilation
and supervised `l build` commands. Mixed helpers explicitly select compilation
or execution. Application/API-probe execution, readiness, signal, cleanup and
isolated assertion watchdogs retain their limits. Publication execution follows
preparatory rebuilding. An old process surviving a rebuild pauses only its
remaining runtime allowance during bounded compilation; check liveness before
and after, then resume that allowance. Never grant a fresh execution budget.

## Compiler defaults

Released binaries embed English resources; users need no overrides. These options support compiler development:

| Override | Default | Developer use |
| --- | --- | --- |
| `--assets ROOT` | Embedded resources | Replace the entire resource filesystem with a complete `lexicon`, `builtins` and `stdlib` bundle, including backend templates. |
| `--language LANGUAGE` | `eng` | Select matching source resources within that bundle. Other source languages remain out of scope until after `v1.0.0`. |

From the repository root, exercise its English resource files with an installed compiler:

```sh
l build --assets . --language eng demo/basic/main.l
```

Both options precede the source filename. Relative asset paths use invocation CWD without changing it. Empty values reject; invalid or unavailable languages fail initialization. Overrides have no embedded fallback or settings file. Keep the bundle's templates and dependencies compatible with the compiler. Bare imports use its `stdlib/index.json`; dictionary helpers resolve inside that provider, never unrelated host files.

Lyger accepts typed language, resource and provenance options; Logiceme accepts the same filesystem and constructs 18 glyphs from private defaults. Generated Rust configuration contracts, absent `configuration/` and `configurations/` directories, all eleven distribution probes, literal preservation, resource diagnostics, isolation and CWD-independent compilation remain required. See [complete CLI help](_generated/cli-help.md).

## October 14, 2026 standalone Service retirement amendment

The author approved the selective retirement recorded in the
[exact protected amendment](./_meta/service-executable-amendment.json).
Standalone Service declarations remain structural; they create no package or
process. Standalone body effects, native lifecycle operations, implicit runtime
members and Service continuation ownership are retired. Main/Basic construction,
primary Service topology, additional App-constructed libraries and raw AML
library/bin controls remain unchanged.

The finite contract updates 53 reviewed hashes, removes 67 retired entries and
adds 31 App-oracle entries, leaving 386 protections. All 302 unrelated protected
hashes, sixteen original required outcomes and required-absent entries remain.
App runtime observers preserve their output, cleanup, provenance, publication
and watchdog obligations. This scope adds no linking feature, library-root
migration or in-process task runtime. Validation remains subject to the pause;
static inspection does not establish a fresh native or acceptance result.
