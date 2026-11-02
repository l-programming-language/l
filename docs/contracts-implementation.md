# L Contracts v1.1 implementation record

This is an implementation ledger, not a certification of the compiler or native
runtime. The specification supplied in `L_Contracts_v1_1_Handoff` is the design
authority. The immutable [acceptance catalog](../contracts/acceptance-catalog.json)
contains 110 requirements; [local mappings](../contracts/acceptance-map.json)
record actual implementation and execution states separately.

## Baseline

- Starting revision: `23bd59d350c1b33ab5bce182000544c14a826197`.
- Starting worktree: clean; no user changes were reset or removed.
- Host: macOS ARM64. Go 1.27.3 and Python 3.11.9 available. Default Cargo is
  1.94.1; required Rust 1.98.1 is installed. Cached lint binaries exist but this
  observation does not establish their integrity or successful execution.
- The handoff validator ran successfully:
  `python3 /Users/connorsanders/Downloads/L_Contracts_v1_1_Handoff/validate_handoff.py --strict-schema`.
  This establishes package/schema/matrix consistency only, not compiler behavior.
- Eleven existing frontend tests, including their subtests, passed with:

```sh
GOCACHE="$PWD/.harness/cache/go-build" GOENV=off GOFLAGS= GOWORK=off GOTOOLCHAIN=local GOPROXY=off go test -mod=readonly -count=1 -timeout=60s ./lyger -run '^(TestBindingPreservesStorageAndInstanceIdentity|TestBindingPreservesMorphologyAndDefinition|TestBindingQualifiedMembersAndAmbiguity|TestDescriptorAliasesRemainReferences|TestDictionaryImportsCannotCaptureCallerAddedMembers|TestConstructionDescriptorAliasesPreserveExpansion|TestConstructionRejectsDescriptorForDifferentTarget|TestParagraphDescriptorTargetDeclarationIdentity|TestPlanOwnershipAtValueBoundaries|TestPlanDefiniteAssignmentUsesContinuingPaths|TestPlanLifecycleEffects)$' -v
```

The command exited 0 (`lyger`, 0.705 seconds test time). It did not compile Rust
or execute Apps. No pre-existing failure was observed in these focused checks.
Full acceptance, native compilation and runtime validation are not implied.

## Current architecture mapping

- Canonical alias traversal and semantic declaration/subject maps supply meaning
  identity; declaration-site lexical scopes remain separate from publication.
- Construction applies every matching descriptor rule in authored order. Track A
  records that policy. The new unique-applicable selector is typed-fixture-only.
- Source statements retain original spans and import provenance. Individual
  token extents require an additive origin map; generated origins must remain
  distinguishable from literal source ranges.
- Orchestration plans retain typed values, captures, branches and lifecycle
  instructions. Existing operational effect flags are not closed effect proofs.
- Storage, capture and Function helpers are selected during composition. Final
  checking must include these decisions and backend support/scaffolding choices;
  an inspection of explicit instructions alone cannot certify a whole program.
- The existing runtime dependency index records resource freshness and selection,
  not native trust. Production contract trust starts empty.
- The CLI's existing compiler and CLI-data layouts and source-only compilation
  views remain compatibility boundaries.

## Pilot observations

| Pilot | Observed source behavior |
| --- | --- |
| Storage.Require | Constructs Failure before `ok_or`; error construction is eager. |
| Pointer.Copy | Copies element signature metadata and retains the shared target. |
| Pointer read/update | Checked borrow; update checks type, reachability/cycle, mutable borrow, then replacement. Old-value destruction is separate. |
| Traverse | Ordered callback invocation; callback result does not introduce early exit. |
| IO | Rendering owns bytes; writes lock stdout, write payload/newline separately, and panic on write errors. Partial output is possible. |
| Start/Execute | Existing source composition Build+Run and Start+Await. |
| Main/Basic | Ordered target-specific construction with Process propagation. |

These observations motivate declarations and tests; source inspection alone is
not approved native implementation evidence.

## Progress

| Milestone | State |
| --- | --- |
| M0 | Baseline inspection and focused tests complete; ledger created. |
| M1 | Implemented; shared identities, strict interchange, Track A recording and deterministic inspection tested. |
| M2 | Implemented supported transfers, hidden lifetime behavior, recursive inference and early profiles; positive/negative/unresolved cases passed. |
| M3 | Implemented typed facets, frames, providers and versioned reference selector; all required typed cases passed. |
| M4 | Implemented protocols, evidence and three cache boundaries; source warm/cold and concrete native lifecycle cross-tests passed. |
| M5 | Implemented supported typed symbolic adjective fragment and concrete instantiation checks; source bodies without a symbolic adapter remain explicitly unverified. |
| M6 | Resource algebra, seven-pilot inventory, public commands and all 110 local mappings implemented; final frozen-tree release gate passed on macOS ARM64. |

### Implemented architecture

- `contracts/` is independent of both compiler packages and uses only the Go
  standard library. Its model separates definitions, meanings, occurrences,
  structural targets, runtime resources, declarations, obligations and evidence.
- Lyger's optional recorder observes actual scoped lookup, descriptor
  applications, native Function composition and late orchestration helpers.
  Normal compiler APIs retain their existing views. Original token ranges and
  import/generated ancestry are side tables, not new source grammar.
- `AnalyzeContracts` projects typed orchestration control flow, source Function
  bodies, arguments, capture construction, activation storage and actual App
  protocol transitions. Unmodeled native representations and cleanup remain
  all-outcome residual obligations, including after terminal source statements.
- Logiceme preflight and emission share the prepared compiler graph, support
  selection and workspace placement. Inventories fingerprint segment templates,
  selected support, compositions and configuration. Rendering remains source
  generation only. The CLI refreshes late frontend choices and combines backend
  closure before enforcement.
- The restricted typed engine supports finite facet configuration, exact role
  paths, alias-aware frames, provider conformance, symbolic adjectives and
  instantiation checks. Track B selection remains typed-only. Symbolic theorems
  about an authored transformer are conditional on separate source conformance.
- Three session caches have distinct binding, logical and implementation keys.
  Source resolution re-executes before comparing a cached binding snapshot;
  cold and warm results have the same canonical report representation. Current
  implementation dependencies never acquire assurance from a logical cache hit.
- The default embedded resource inventory is unchanged. `contracts/index.json`
  is an optional provider index/example; its empty default need not be embedded.
  Missing metadata is unknown, and custom providers have no embedded fallback.

### Focused execution evidence

The following commands used the repository Go cache and disabled ambient module
configuration. They are investigative results, not substitutes for the release
gate. Exact acceptance-ID mappings live in `contracts/acceptance-map.json`.

```sh
GOCACHE="$PWD/.harness/cache/go-build" GOENV=off GOFLAGS= GOWORK=off GOTOOLCHAIN=local GOPROXY=off go test -mod=readonly -count=1 -timeout=60s ./cmd/l -run '^TestContracts' -v
GOCACHE="$PWD/.harness/cache/go-build" GOENV=off GOFLAGS= GOWORK=off GOTOOLCHAIN=local GOPROXY=off go test -mod=readonly -count=1 -timeout=60s ./logiceme -run '^TestContracts(Backend|Workspace)' -v
GOCACHE="$PWD/.harness/cache/go-build" GOENV=off GOFLAGS= GOWORK=off GOTOOLCHAIN=local go test -mod=readonly -count=1 -timeout=10m ./lyger -run '^TestRuntimeValuesNative$/(snapshot-and-sharing|cycle-through-function|cycle-through-container)$' -v
GOCACHE="$PWD/.harness/cache/go-build" GOENV=off GOFLAGS= GOWORK=off GOTOOLCHAIN=local go test -mod=readonly -count=1 -timeout=10m ./logiceme -run '^TestContractsRuntimeProtocolCrossCheck$' -v
```

These passed respectively in 2.728s, 0.317s, 3.687s and 3.191s. The first two
generate/analyze source without native execution. The latter two compile native
fixtures and run the existing snapshot/cycle and lifecycle/cleanup observers;
they do not grant approval witnesses. The separate existing
`TestRuntimeLifetimesNative` also passed (1.520s). Additional exact command
records accompany the acceptance map.

### Release-check attempts

1. `scripts/l-ci bootstrap` initially failed because sandboxed Rustup could not
   create its temporary file outside the workspace. Recorded in
   `.harness/runs/ci-bootstrap-2455184080/stage.json`.
2. The same command with tool-cache access passed, recorded in
   `.harness/runs/ci-bootstrap-541648304/stage.json`. Tool pins and locks were
   preserved.
3. Early `scripts/l-check` attempts stopped on formatting or six new-code lint
   issues (checked errors/type assertions and struct conversion). These were
   fixed without suppressions. One intermediate attempt encountered a new
   cache helper while its implementation was still being written; it is not
   recorded as a repository baseline failure.
4. Lint passed in `.harness/runs/ci-lint-3349073117/stage.json`. Its following
   acceptance run `.harness/runs/ci-acceptance-1103294531/stage.json` passed
   doctor (501 checks), build and package discovery, but both test phases
   failed. New-code issues were stale generated schemas, an unnecessary extra
   embedded resource, and intermediate tests subsequently fixed. Existing
   native tests also encountered sandbox-denied Unix sockets and cache locks.
   Protected integrity and all original compiler-view tests passed in that run.
5. Subsequent full runs passed with access needed for the existing native tests.
   Final review fixes were followed by another complete run. No completion
   claim depends on the failed or intermediate attempts.

No observed focused baseline failure predates this implementation. The initial
full gate was not run before edits, so later environmental failures are not
misrepresented as a clean full baseline.

### Public command demonstration

The temporary input `.harness/contracts-v1.1/pilot.l` imports IO.Print, constructs
a Basic Worker App, then Run/Await/Show. The commands below were explicitly run;
none executes the generated App. They use the locally built
`.harness/bin/l-contracts` binary.

```sh
GOCACHE="$PWD/.harness/cache/go-build" GOENV=off GOFLAGS= GOWORK=off GOTOOLCHAIN=local GOPROXY=off go build -mod=readonly -o .harness/bin/l-contracts ./cmd/l
.harness/bin/l-contracts contracts inspect --format json --output .harness/contracts-v1.1/inspect.json .harness/contracts-v1.1/pilot.l
.harness/bin/l-contracts contracts check --profile NoHeapAllocation --profile NoBlocking --profile NoExternalIO --profile NoPanic --profile OwnedLifecycle --report .harness/contracts-v1.1/check.json --output .harness/contracts-v1.1/check.txt .harness/contracts-v1.1/pilot.l
.harness/bin/l-contracts contracts explain --binding binding-use:9f4cd9c21393a63520064435c67d19e063147208fd36982dfa92c8e1ac4d7155 --output .harness/contracts-v1.1/explain.txt .harness/contracts-v1.1/pilot.l
.harness/bin/l-contracts contracts verify-dictionary --report .harness/contracts-v1.1/dictionary.json --output .harness/contracts-v1.1/dictionary.txt io
```

Exit codes were respectively **0, 0, 1, 0, 1**. Inspection enumerated 92 native
boundaries with unknown transitive coverage and no executed-check records. All
five requested profiles were unresolved. Dictionary verification discovered
canonical declarations but did not certify source/native bodies. These are
expected failures of enforcement, not successful guarantees.

Representative exact identities from that report:

- Basic selected meaning:
  `meaning:77d3d77850fbe7de9203a5eec5426512b21c9da33090189618eb020d1da0e46e`.
- Its exact construction target:
  `construction:81b30988634ef4a0dad9f0723b4e862d05e8f1ee4b61249a41259b29a63c5ce4`,
  type `app`, phase `construction`, source origin `source:pilot.l:62-67`.
- Run selected meaning:
  `native-meaning:3d1e97cd1ad74305973eaf8fc7b98d3ddf89603b24f4317f31e4984c09862a7c`,
  on that same App root in phase `invocation`, origin `source:pilot.l:97-100`.
- Basic's reason is existing ordered descriptor composition: every matching
  canonical target rule applies. Run's reason is the configured native semantic
  role retained by the typed instruction planner. Neither is selected to make a
  profile pass.

### Supported scope and assurance limits

All `BASE-001`–`BASE-060`, `LEX-001`–`LEX-040`, and `FALS-001`–`FALS-010` have
local mappings with executed coverage at their specified supported layer. A
mapping is not a production native witness or proof of arbitrary source bodies.
The typed symbolic fragment includes nontrivial positive certificates, rejected
definitions, incompatible instances, callback costs, concrete lowering failures,
ordered wrappers and repeated construction. Source dictionary discovery remains
distinct from symbolic-body verification; unsupported source-body projection is
reported as `instantiation_only`/unresolved.

Native trust remains empty unless the caller explicitly selects an approved
policy. Test-only assumptions do not grant production approval. Real pilot
reports keep opaque transitive calls, conditional destruction, failure and
representation costs unresolved. Direct native observations now separately
cover signature metadata copying/shared Pointer targets, held-borrow rejection,
ownership-cycle rejection, snapshots, lifecycle claim/release, external
completion, and cleanup cause identity. These observations cover the exercised
cases, not every native path or target configuration.

The source projection conservatively leaves capture-instance substitutions and
normal Function-return ownership transfers unresolved where the typed adapter
does not supply enough detail. Environmental exclusions include host loss,
external forced termination and scheduler fairness. No contract-driven runtime
optimization, failure-channel change, lazy Require, implicit termination,
production example vocabulary or frozen interface change was introduced.

Track A byte-equivalence, definition-site hygiene, distinct same-type roles and
unchanged default resolution are tested. Track B ends at the typed reference
selector; there is no source-facing activation mode. Future precision work can
extend source symbolic-body/capture/return adapters without changing this
compatibility boundary or weakening unresolved obligations.

The final soundness review added concrete-target type consistency to adjective,
provider and reference-selector checking, so caller facts cannot override the
actual attachment. Native declaration requirements/restrictions now form explicit
evidence-graph premises, and conditional claim assumptions propagate through
derivations. Reports cannot certify an unidentified Rust toolchain/panic
configuration by signing the source-only placeholder scope. Target identity also
normalizes omitted versus empty root paths across JSON interchange. Focused
regressions for these fixes passed before the final full gate. Omitted symbolic
callback costs and outcome coverage also remain unknown; a closed effect row
does not supply a zero work bound or success-only execution. Independent
nonblocking theorems remain available when their premises are established.

## Final release validation

`scripts/l-ci bootstrap` passed as recorded above. The final `scripts/l-check`
exited **0**. All three stages recorded the same unchanged source fingerprint:
`91445022a916ddbc5a60417eace9ff170c6b9b0cfea80455d9a665ad7eec608f`.

| Stage | Outcome | Local report |
| --- | --- | --- |
| Workflow lint, formatting, ordinary and nativecompile Go lint profiles | Passed, no warnings/errors | `.harness/runs/ci-lint-1812429660/stage.json` |
| Harness doctor/build/discovery and both unfiltered test phases | Passed | `.harness/runs/ci-acceptance-318679881/stage.json` |
| Complete tagged native-compilation stage | Passed | `.harness/runs/ci-native-3166192262/stage.json` |

The authoritative verification report is
`.harness/runs/ci-acceptance-318679881/verify.report.json`. Doctor passed 501
checks. Protected integrity passed before and after. All 16 original
compiler-view test records passed. The ordinary phase recorded 736 test/subtest
results and the native phase 4,457, with no failures. These counts include
subtests and are not a count of independent proofs. Build and package discovery
passed separately. The tagged stage ran exactly:

```sh
go test -mod=readonly -count=1 -tags=nativecompile -run '^(TestLucidNativeCompile|TestNativeMetadataArrowsCompile)$' -timeout=30m -json ./logiceme
```

The public CLI binary was then rebuilt and the four demonstration commands
above rerun, retaining exit codes `0,1,0,1`. Their final reports still enumerate
92 native boundaries, unknown transitive closure, unresolved enforced profiles,
and no hidden execution records. Only this documentation and the acceptance
ledger were updated after the release run to record its outcome; implementation
code and tests were unchanged.

All supported milestone gates are complete. Remaining work is precision and
assurance expansion: a source symbolic-body adapter, more complete
capture/return ownership projection, explicit concrete native-environment
identification, and independently reviewed native conformance witnesses. The
next concrete extension is a source-body-to-symbolic-plan adapter that preserves
definition-site bindings and checks one supported adjective against its declared
target requirements. These limits remain visible obligations, not passing
annotations. No Linux result, production native certification, commit or push
is claimed.

## Validation policy

The user explicitly authorized the release validation in the implementation
plan: run `scripts/l-ci bootstrap`, then `scripts/l-check`. Preserve both
unfiltered acceptance phases, tagged native compilation, original outcomes,
protected bytes, tool pins, deadlines and concurrency. Record blocked operations
without substituting an earlier green result. No push, publication, CI dispatch
or commit is authorized by this task.
