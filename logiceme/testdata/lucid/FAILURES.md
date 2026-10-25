# Rust Lucid verification evidence

This ledger separates the maintained compile-only corpus from dated validation
results. The [suite guide](README.md#current-inventory) owns current counts and
commands; the [Rust reference](../../../builtins/README.md#current-integration-limits)
owns supported contracts and remaining limits. Historical results do not validate
a later checkout or authorize changes to frozen inputs.

## Latest recorded verification — October 15, 2026

The [macOS ARM64 validation](../../../testdata/service-execution/validation-2026-10-15.md)
passed bootstrap, lint, canonical acceptance and tagged native compilation.
Tested source identity:
`38096646d1a9cc5a68ea92ba2ea34d4997b526332e1cf0b1085a53911d8b44f5`.

| Phase | Recorded outcome |
| --- | --- |
| Bootstrap and lint | Pinned setup, workflow lint and both complete Go lint profiles passed. |
| Unfiltered acceptance | All 4,383 named outcomes and sixteen original obligations passed. Before/after integrity each passed 508 checks, including 386 protected hashes and eleven required-absent paths. |
| Separate native stage | `TestLucidNativeCompile` and `TestNativeMetadataArrowsCompile` passed. The Lucid portion retains 108 Expected/Generated contracts, including its intentional Rust errors. |

The unchanged Lucid inventory remains 54 complete applications with 270 exact
artifacts and 192 malformed inputs. Each native phase requires 42 successful
builds and 12 specific intentional compiler errors. These are compile-only
contracts: returned source is compiled unchanged, and these cases do not execute
applications or deployment scripts. Supervised frontend/runtime tests are a
separate part of ordinary acceptance; see the
[implementation ledger](../../../testdata/integration/IMPLEMENTATION.md) for
run artifacts and their execution scope.

This documentation update did not rerun language validation. Follow the
[validation policy](../../../docs/development-harness.md#temporary-validation-pause)
for the local pause and authorized hosted runs.

## Preserved scope and remaining boundaries

All eighteen segment assets remain frozen by
[templates.sha256.json](templates.sha256.json). Exact source/support bytes,
ordered paths, dependencies, independent expectations, classified malformed-input
causes, provider isolation and prior-bundle stability remain assertions.
Expected and Generated native phases retain their separate obligations; an
unrelated Cargo failure cannot satisfy an intentional diagnostic case.

The current API generates Rust and offers only optional runtime support through
`GenerationOptions.IncludeRuntime`; the retired output-target selector is not a
current interface. App-constructed libraries, structural Service rejection, raw
AML library/bin targets, workspace runtime support and App execution have
separate regressions. The Lucid corpus does not establish general ownership
inference, method definitions, async definitions, composed implementations or
runtime fairness/cancellation.
Consult the [current limits](../../../builtins/README.md#current-integration-limits)
before treating an old preparation branch as supported.
