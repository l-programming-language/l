# Contracts v1.1

Contracts describe selected meanings and their typed behavior. Inspection does
not change resolution, build a Cargo workspace, publish an executable, run a
tool query, or run an App. The default native approval policy is empty: a declaration is not proof of
its implementation. A successful inspection can contain unknowns; enforced
checking fails when a required fact is contradicted or unresolved.

## Commands

Use `l contracts --help` or `l contracts -h` for contracts usage. Each analysis
subcommand requires exactly one input. Options precede that input; use `--` to
end option parsing before a filename beginning with a dash. No contracts command
builds a Cargo workspace, publishes an executable, or launches an App. Explicit
`evidence capture` is the only contracts operation that queries external tools.

### Inspect

```sh
l contracts inspect --format json --output report.json program.l
```

Inspection reports selected meanings, bindings, typed targets, source
certificates, and unresolved obligations. Successful inspection does not require
all reported properties to be established.

### Check

```sh
l contracts check --profile NoHeapAllocation --profile NoBlocking --report check.json program.l
```

Checking requires `--profile` or `--request-file`. Repeat `--profile` to enforce
several profiles over the whole selected entry and all modeled outcomes. Unknown or
contradicted required properties fail checking. A profile does not select a
different word or provider to obtain a passing result.

| Profile | Required assurance |
| --- | --- |
| `NoHeapAllocation` | No allocation in the selected modeled scope. |
| `NoBlocking` | No blocking in the selected modeled scope. |
| `NoExternalIO` | No external IO in the selected modeled scope. |
| `NoPanic` | No panic in the selected modeled scope. |
| `OwnedLifecycle` | Lifecycle ownership checks over `whole_entry`. |
| `PreservesOnFailure` | An explicit request with resource footprint, observation and failure outcomes. |

The profile definitions also govern request validation. Unknown names and
unsupported profile/scope combinations reject.

### Explain

```sh
l contracts explain --binding OCCURRENCE_ID program.l
l contracts explain --binding source:program.l:BYTE_OFFSET program.l
```

Explanation requires `--binding`. Use an exact occurrence ID or an
`origin.source:origin.start` selector from an inspection report. Byte offsets
refer to original source where available. Generated uses have parent origins and
expansion ordinals. Reports retain definition identity, selected meaning,
occurrence, role path, definition-site environment, use-site environment, and
the existing resolver's selection reason. Aliases share meanings; separate
applications and same-type roles remain distinct. A selector with no matching
binding is an invocation error.

### Verify a dictionary

```sh
l contracts verify-dictionary --report dictionary.json io
```

Dictionary input can name an indexed dictionary or a source dictionary file.
Verification defaults to `--assurance implementation` for the complete dictionary.
It reports discovered definitions separately from symbolic certificates. Merely
loading a source dictionary never certifies its bodies. Whole-dictionary
verification fails while required implementation or native obligations remain
unresolved.

### Focused symbolic verification

```sh
l contracts verify-dictionary --adjective MEANING_OR_SOURCE:BYTE_OFFSET --assurance symbolic --manifest source-contracts.json --report adjective.json dictionary.l
```

Explicit `--assurance symbolic` requires one or more repeatable `--adjective`
selectors. Each selector identifies one canonical rule meaning or its exact
source-origin byte offset; ambiguous and missing selections reject. This mode
checks only the selected definition-time structural obligations. Its report
retains concrete instantiation and native obligations separately, so success is
not a runtime guarantee. Options still precede the dictionary input. Whole-entry
`--profile` requirements cannot be combined with this narrower symbolic mode.

`--adjective` is supported only by `verify-dictionary`. With the default
implementation assurance, selecting adjectives still retains whole-dictionary
enforcement; use explicit symbolic assurance for the narrower check.

### Capture native environment evidence

```sh
l contracts evidence capture --output environment.json program.l
l contracts evidence capture --output environment.json --target aarch64-apple-darwin program.l
l build --native-evidence-output build-environment.json program.l
```

Capture always writes a JSON environment document to its required `--output`.
It is an explicit tool observation: it resolves the pinned local toolchain
and records executable hashes, version-query commands and results, the native
host target, selected source/backend fingerprints, and relevant Cargo
configuration. It never fetches a toolchain, builds a workspace, or runs an App.
`--target` is an assertion that must match the observed host; it does not enable
cross-compilation. Omit it for the current native host.
If `CARGO_HOME` is set, explicit evidence collection requires an absolute path so
configuration identity does not depend on an unrelated working directory.

Environment fingerprints use this exact allowlist: `PATH`, `CARGO_HOME`,
`RUSTUP_HOME`, `RUSTUP_TOOLCHAIN`, `RUSTFLAGS`, `RUSTDOCFLAGS`, `CC`, `CXX`, `AR`,
`CARGO_NET_OFFLINE`, `CARGO_ENCODED_RUSTFLAGS`, `CARGO_TERM_COLOR`,
`CARGO_PROFILE_DEV_DEBUG`, `CARGO_INCREMENTAL`, `RUSTC_WRAPPER`,
`RUSTC_WORKSPACE_WRAPPER`, `RUSTC_BOOTSTRAP`, and `RUSTC_FORCE_INCREMENTAL`.
Only the values selected by the existing sanitized build environment are hashed.
Credential variables and other unlisted variables are omitted entirely, including
their hashes; non-test evidence containing unlisted environment keys rejects.
Conventional Cargo configuration files and their absence are tracked separately.

Build evidence uses record version 2 and includes the actual Cargo invocation,
locked workspace input digest, artifact hashes and companion generation.
Offline checks reconstruct that placement in memory. Environment-only capture
remains at record version 1.
A build record without the required generation cannot establish current build
assurance. Failed Cargo invocations remain visible. Artifact evidence records
compilation even if subsequent publication fails.

Neither document approves native behavior. Panic strategy, allocator behavior,
linker internals and transitive native conformance remain unproved by these
observations. `inspect`, `check`, `explain` and dictionary verification consume
these documents offline: they compare current source, selected dependencies and
relevant Cargo configuration, but never replay command strings, run version
queries or re-attest historical tool binaries.

### Scoped checks and failure footprints

First inspect `scopes`, `native_uses`, and their published resource locations.
Copy an exact scope ID into a version-1 `contract_check_requests` document; use
`whole_entry` to retain the whole selected program. Do not invent resource IDs
from spelling or select a same-type role as a substitute.

A request file has this shape, with placeholders replaced by values from the
inspection report:

```json
{
  "schema_version": "1.1",
  "kind": "contract_check_requests",
  "producer": "project-contracts",
  "version": "1",
  "checks": [{
    "id": "update-preservation",
    "scope": "SCOPE_ID_FROM_INSPECT",
    "profile": {
      "name": "PreservesOnFailure",
      "version": "1.1",
      "scope": "",
      "target": {"root": ""}
    },
    "footprint": {
      "locations": [{"resource": {"id": "RESOURCE_ID_FROM_SCOPE"}, "path": []}],
      "observation": "state_at_completion",
      "outcomes": ["returned_failure"]
    }
  }]
}
```

The request's selected scope supplies the profile boundary and target. Copy each
complete location from that scope's `resources`, including its resource kind and
path when present. An absent, ambiguous or unsupported scope, or a location that
the scope does not expose, rejects. `PreservesOnFailure` requires a nonempty
footprint, explicit failure outcomes and `state_at_completion`; a bare
`--profile PreservesOnFailure` is invalid. `OwnedLifecycle` currently requires
`whole_entry`.

For List/Map update failure examples, distinguish the native helper's input
snapshot from the caller's destination. A `native_body` scope can expose both
resources: an unchanged caller destination does not prove that the helper left
its input snapshot unchanged after a partial failure. Select the exact published
resource and observation being claimed.

```sh
l contracts check --request-file checks.json --native-environment environment.json --manifest native-contracts.json --trust-policy reviewed-policy.json --report check.json program.l
```

A native manifest's `outcome_contracts` describes outcome-specific effect rows,
`absent_effects`, costs and completion-state relations. These are declarations.
Inspect with the manifest and environment selected to obtain `native_components`:
each component identifies its exact proposition, native use, target and current
dependency fingerprints. Reviewed trust witnesses must match those propositions
individually. An approved absence of allocation cannot certify absent blocking,
a closed outcome set, cleanup, or unchanged state.

Declared `guarantees` are separate report propositions; they do not replace
outcome-specific summary components. Both use the same exact native selection,
signature, environment and dependency checks, with requirements checked at the
actual invocation entry. Facts established later cannot satisfy those entry
requirements.

Native uses distinguish the selected `native_body` from
`invocation_composition_overhead` and `source_body_lowering_overhead`. Invocation
overhead covers the selected wrapper and its context handling. Source-body
overhead covers prologue, value transfers and return composition. These claims
exclude separately analyzed source operations and owned-storage destruction;
approving one cannot discharge the others. Published scopes retain these
inclusions and exclusions explicitly.

Checking applies only current, applicable component evidence, then reinfers the
selected consumers and profiles. Source guards and entry requirements still
apply. An unchanged completion-state relation may justify preservation despite
an earlier write; a partial failure with an observable changed state contradicts
it. Missing or stale evidence leaves the dependent obligation unresolved.
Independent opaque helpers remain unresolved, including outside a successfully
checked narrower scope. Reports retain every controlling request obligation and
its assumptions. Test-only trust policies remain labeled and grant no production
approval.

## Options and reports

Analysis subcommands use the following resource and report options. Explicit
capture requires `--output` and rejects input environment evidence, trust policy
and profile enforcement:

| Option | Behavior |
| --- | --- |
| `--format text\|json` | Select the displayed report format; default `text`. |
| `--output PATH` | Write the displayed report to a file instead of stdout, using the selected format. |
| `--report PATH` | Also write the structured JSON report, regardless of the displayed format. |
| `--language LANGUAGE` | Select source resources; default `eng`. An explicit empty language rejects. |
| `--assets ROOT` | Use one complete resource filesystem in place of embedded resources. |
| `--manifest PATH` | Add a metadata manifest; repeat to contribute multiple manifests. |
| `--trust-policy PATH` | Load an explicit trust-policy document; production native approval otherwise starts empty. |
| `--native-environment PATH` | Read a previously captured native environment without executing recorded commands or probing tool versions. |
| `--request-file PATH` | For `check`, load explicit scoped checks and failure-preservation footprints. |

Both destinations are checked together against physical imports, selected
metadata and compiler assets, including file aliases. When both flags are used,
the destinations must be distinct. Relative paths resolve from the working
directory. Successful file replacement uses private permissions (0600). Keep the exit status alongside the
report when recording an enforced check.

`--assets` selects one complete resource provider; missing optional contract
metadata does not trigger embedded metadata fallback. An optional
`contracts/index.json` contains `schema_version` equal to `1.1` and a `manifests`
array of resource-relative paths. Additional manifests contribute records;
conflicting canonical declarations fail instead of overriding by load order.

## Exit statuses

| Status | Meaning |
| --- | --- |
| `0` | Successful inspection or explanation; successful enforcement at its reported scope; or displayed help. Inspection can still contain unknowns. |
| `1` | Failed analysis or report IO, or required enforcement obligations that are contradicted or unresolved. |
| `2` | Invalid invocation, selected metadata, or selector. |

Symbolic verification returning `0` establishes only the selected structural
obligations. Unresolved native assurance remains visible in its report.

## Source adjective bodies

The source adapter observes the same resolved construction operations used by
ordinary compilation. It supports canonical constructors, fixed names and
literal signature atoms, ordered child roles, parameter forwarding, and
structural writes and frames. Dictionary helpers retain their definition-site
bindings. Every matching
descriptor rule still applies in order; repeated constructions remain distinct.

For example, the existing source grammar can express:

```text
"Compact" describes a Generic App with a New Main Service using any Process.
```

A symbolic check can establish its constructed roles and exact structural
footprint without creating an App instance. Fixed child ordinals require an
explicit premise about the complete existing child list. Concrete applications
check that premise against the target immediately before the rule applies.
Canonical nominal identity, rather than a displayed type name, determines
target compatibility.

Facet/role manifests can supply bodyless `source_transformers` declarations keyed
by the selected rule's canonical meaning. Requirements and promised properties
come from those declarations; the implementation plan, actual writes and source
provenance come from the compiler. An authored `transformers` body remains a
separate typed model and cannot stand in for the source implementation.

Start with `inspect --format json` to obtain the rule meaning, canonical nominal
domain, symbolic target, and unmet shape premises. Copy those identities into the
declaration; do not invent IDs from adjective spelling. Structural promises use
the reported shape fields: canonical nominal/constructor identity, names, exact
child presence/counts, signatures, and ordered parameter/wrapper selections.
Runtime effects, ownership, protocol and cost promises remain unsupported by this
structural adapter even when repeated as preconditions.

Source certificates separate the structural theorem from instantiation and
lowering assurance. Native templates retain their ordered wrapper identity but
their executable behavior remains opaque. Forwarding `Process` supplies a
parameter, not an invocation-count, purity or cost proof. Unsupported target
inspection remains instantiation-only. Structural source verification never
establishes runtime `NoBlocking` or `NoHeapAllocation` by omitting native work.

## Typed analysis boundary

The independent `contracts` Go package holds identities, restricted predicates,
typed plans, outcome-sensitive summaries, facets, frames, role conformance,
protocols, evidence, resources and the three session caches. Lyger records actual
resolution and composition decisions. Logiceme prepares an inventory from the
same selected support and program objects it renders. The CLI includes late
orchestration/runtime selections before final enforcement. No guarantee is
inferred by parsing generated Rust.

`EvaluateNativeUses` evaluates declarations against selected typed native uses.
Its `NativeEvaluation` results supply both inference components and report
propositions; invocation-entry requirements remain typed-analysis obligations.

`LygerService.AnalyzeContracts` is additive to ordinary compilation and accepts
separate analysis options. `ResolveContractBindings` exposes Track A records.
`ContractAnalysisOptions.SourceTransformers` supplies bodyless source declarations;
the result exposes compiler-derived projections, source certificates and concrete
instantiations. `VerifyDictionaryContractsWithOptions` adds explicit canonical
meaning selection and assurance scope through `DictionaryVerificationOptions`,
while the existing dictionary API retains its default behavior. Session reuse
revalidates bindings before logical reuse and checks current native dependencies
again at concrete assurance boundaries.
`Session` owns private in-memory caches and returns independent snapshots. It
reuses the current verified source certificate while checking each concrete
instance and its lowering. It has no persisted-cache import or legacy migration
API; incomplete dependencies require fresh analysis.
`PrepareContractGeneration` and `PrepareWorkspaceContractGeneration` expose
backend preparation, an immutable inventory view, and one-shot source rendering.
`VerifyInventory` detects a changed prepared inventory. Frontend-only analysis
keeps backend obligations unresolved.

The supported typed fragment includes sequence/branch/call/callback transfer,
outcome-sensitive demand, guards, alias-aware effects, ownership transfers and
cleanup, representation-dependent mutation, finite facets and cross-facet rules,
exact-path frames, provider substitution, and symbolic adjective checking.
Direct selections, rule-derived selections and supplied facet premises share
schema, target, domain and cardinality validation. Rule closure precedes closed
configuration checks. Unknown contribution/rule conditions remain obligations;
rules never perform an implicit replacement. Authorized replacements preserve
release obligations and invalidate facts about the replaced slot.
`facet:ID:SCOPE` configuration predicates are accepted only when declared schema identities
make the split unambiguous, including IDs or scopes containing colons.
Adjective certificates are rechecked at instantiation; concrete callback effects,
captures and lowering can invalidate a valid symbolic certificate. Repetition and
wrapper order remain observable. Reflection outside this fragment is explicitly
instantiation-only. Typed fixtures exercise vocabulary that the source language
does not yet expose; these fixtures do not create production Theme/Backend APIs.

The unique-applicable reference selector is versioned and available only over
explicit typed candidates. Unknown applicability blocks a unique selection.
Ordinary source resolution still applies every currently matching descriptor in
order. This release adds no Track B source mode, phrase recognition, attachment
rule, import precedence, or contract-driven fallback.

## Evidence and limits

The interchange schemas in `contracts/schemas` are generated from the shared
Go model by `go run ./tools/contracts-schema KIND`; output is written to stdout.
The seven readable snapshots are retained for offline authoring without a Go
toolchain and marked as generated for diff review. Their tests compare exact
generated bytes with the shared model.
Strict decoding rejects unknown fields, duplicate keys, unsupported versions and
ambiguous declaration identities. Native manifests cannot contain established
evidence. Trust documents approve named, proposition-specific witnesses with
authority, rationale, exact contract/body/target fingerprints and dependencies.
Approval of one proposition does not close unrelated native behavior.
Requirements and target restrictions are separate evidence dependencies: a
current native witness cannot bypass an unmet applicability condition. Explicit
capture or build evidence supplies a reusable native-environment identity;
source-only inspection without it retains an unresolved environment dependency.
Evidence remains conditional on the recorded environment and selected native
contracts. It does not attest the current deployment or grant approval to a
native implementation merely because compilation succeeded. Typed clients can
supply independently checked premises; test-only assumptions remain explicitly
labeled.

Profiles cover the whole selected entry and all modeled outcomes by default.
Typed API callers and explicit request-file checks may narrow the scope, which
remains visible in each claim. Allocation, blocking, external IO and panic are distinct constraints;
neither nonblocking nor unknown cost establishes termination. Lifecycle ownership
does not remove external completion or runtime synchronization. Preservation on
failure requires an explicit observable footprint and selected failure outcomes.

The numerical policy widens growing components after eight growth rounds, caps
updates at 64 per strongly connected component, and retains at most 32 path
alternatives. Each inference has a cumulative budget of 8,192 analysis steps
across repeats, calls and path alternatives. Exhaustion reports partial coverage
and leaves unfinished obligations unresolved. Resource accounting separates cumulative allocation, peak
live resources and retained resources; cost formulas can contain callback cost
variables and dependent bounds.

Native reports distinguish enumerated compiler boundaries, transitive dependency
closure and implementation conformance. Opaque native internals, destructors and
external runtime behavior remain explicit unresolved boundaries without approved
evidence. `Require` constructs a failure only on its selected failure path;
`Accept` releases a replaced error before constructing its replacement. Those
source compositions do not establish general allocation or destructor-cost
bounds. Runtime failure channels, snapshot/sharing semantics and App lifecycle
checks retain their contracts. Inspection records no build/test evidence;
explicit validation commands and their results remain separate from requirement
coverage and native approval.

See the [acceptance mapping](../contracts/acceptance-map.json) for test witnesses
and current limitations, and the [evidence index](contracts-evidence.md) for
revision-specific historical records and their availability. Platform coverage
belongs to each recorded revision; historical results do not validate a changed
checkout.
