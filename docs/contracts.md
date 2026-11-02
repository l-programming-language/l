# Contracts v1.1

Contracts describe selected meanings and their typed behavior. Inspection does
not change resolution, build a Cargo workspace, publish an executable, or run an
App. The default native approval policy is empty: a declaration is not proof of
its implementation. A successful inspection can contain unknowns; enforced
checking fails when a required fact is contradicted or unresolved.

## Commands

Options precede the input, as with the existing CLI:

```sh
l contracts inspect --format json --output report.json program.l
l contracts check --profile NoHeapAllocation --profile NoBlocking --report check.json program.l
l contracts explain --binding OCCURRENCE_ID program.l
l contracts explain --binding source:program.l:BYTE_OFFSET program.l
l contracts verify-dictionary --report dictionary.json io
```

Use the exact occurrence ID or `origin.source:origin.start` selector from an
inspection report. Byte offsets refer to original source where available.
Generated uses have parent origins and expansion ordinals. Reports retain
definition identity, selected meaning, occurrence, role path, definition-site
environment, use-site environment, and the existing resolver's selection reason.
Aliases share meanings; separate applications and same-type roles remain distinct.

Common options are `--language`, `--assets`, repeatable `--manifest`, and
`--trust-policy`. `--assets` selects one complete resource provider; missing
optional contract metadata does not trigger embedded metadata fallback. An
optional `contracts/index.json` contains version `1.1` and a `manifests` array
of resource-relative paths. Additional manifests contribute records; conflicting
canonical declarations fail instead of overriding by load order.

Exit status is 0 for successful inspection, 1 for failed analysis or unresolved/
contradicted enforcement, and 2 for invalid invocation or metadata. Dictionary
verification reports discovered definitions separately from symbolic
certificates. Merely loading a source dictionary never certifies its bodies.

## Typed analysis boundary

The independent `contracts` Go package holds identities, restricted predicates,
typed plans, outcome-sensitive summaries, facets, frames, role conformance,
protocols, evidence, resources and the three session caches. Lyger records actual
resolution and composition decisions. Logiceme prepares an inventory from the
same selected support and program objects it renders. The CLI includes late
orchestration/runtime selections before final enforcement. No guarantee is
inferred by parsing generated Rust.

`LygerService.AnalyzeContracts` is additive to ordinary compilation and accepts
separate analysis options. `ResolveContractBindings` exposes Track A records.
`PrepareContractGeneration` and `PrepareWorkspaceContractGeneration` expose
backend preparation, an immutable inventory view, and one-shot source rendering.
`VerifyInventory` detects a changed prepared inventory. Frontend-only analysis
keeps backend obligations unresolved.

The supported typed fragment includes sequence/branch/call/callback transfer,
outcome-sensitive demand, guards, alias-aware effects, ownership transfers and
cleanup, representation-dependent mutation, finite facets and cross-facet rules,
exact-path frames, provider substitution, and symbolic adjective checking.
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

The five interchange schemas in `contracts/schemas` are generated from the shared
Go model by `go run ./tools/contracts-schema KIND`; output is written to stdout.
Strict decoding rejects unknown fields, duplicate keys, unsupported versions and
ambiguous declaration identities. Native manifests cannot contain established
evidence. Trust documents approve named, proposition-specific witnesses with
authority, rationale, exact contract/body/target fingerprints and dependencies.
Approval of one proposition does not close unrelated native behavior.
Requirements and target restrictions are separate evidence dependencies: a
current native witness cannot bypass an unmet applicability condition. The
source-only CLI does not identify the concrete Rust toolchain or panic
configuration by running external tools. Its placeholder environment therefore
cannot grant concrete native implementation assurance, even with a newly signed
witness. Typed clients can supply a concrete configuration and independently
checked premises; test-only assumptions remain explicitly labeled.

Profiles cover the whole selected entry and all modeled outcomes by default.
API callers may explicitly narrow the scope, which remains visible in each
claim. Allocation, blocking, external IO and panic are distinct constraints;
neither nonblocking nor unknown cost establishes termination. Lifecycle ownership
does not remove external completion or runtime synchronization. Preservation on
failure requires an explicit observable footprint and selected failure outcomes.

The numerical policy widens growing components after eight growth rounds, caps
updates at 64 per strongly connected component, and retains at most 32 path
alternatives. Precision loss produces dependent unknowns, never a claim of
unbounded execution. Resource accounting separates cumulative allocation, peak
live resources and retained resources; cost formulas can contain callback cost
variables and dependent bounds.

Native reports distinguish enumerated compiler boundaries, transitive dependency
closure and implementation conformance. Opaque native internals, destructors and
external runtime behavior remain explicit unresolved boundaries without approved
evidence. Runtime failure channels, eager `Require`, snapshot/sharing semantics,
and App lifecycle checks are unchanged. Inspection records no build/test evidence;
explicit validation commands and actual results belong in the implementation
ledger and acceptance map.

See [implementation progress](contracts-implementation.md) and the
[acceptance mapping](../contracts/acceptance-map.json) for executed coverage and
remaining limitations. macOS validation establishes no Linux result.
