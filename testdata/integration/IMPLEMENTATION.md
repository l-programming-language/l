# Rust frontend and runtime implementation evidence

This document records implementation boundaries and revision-specific validation. Current
contracts live in the [suite guide](README.md), [architecture](../../docs/architecture.md),
[CLI/runtime guide](../../docs/cli-build.md), [private runtime contract](../phase1/runtime/contract.md)
and [typed-builtin/Service coverage map](../service-execution/coverage.md).
A recorded pass establishes that revision and host, not a later working tree.

## App processes and structural Services

The [approved finite amendment](../../docs/_meta/service-executable-amendment.json)
retains structural Service declarations and rejects standalone body effects,
native lifecycle operations and implicit runtime members. Only Apps own process
identity and continuations. Main/Basic, App-owned Service libraries and raw AML
library/bin controls remain. Service-named runtime observers exercise Apps with
their existing behavioral obligations.

The [October 15 macOS ARM64 validation](../service-execution/validation-2026-10-15.md)
passed canonical acceptance: all 4,383 named outcomes, including both publication
regressions' runtime assertions, and all sixteen original required outcomes.
All 508 integrity checks passed before and after acceptance, including 386
protected hashes and eleven required-absent paths. Workflow lint, both complete
Go lint profiles, source generation and tagged native compilation passed.
Tested source identity: `38096646d1a9cc5a68ea92ba2ea34d4997b526332e1cf0b1085a53911d8b44f5`.
This documentation update did not rerun language validation.

## Evidence boundaries

- **Generation:** frontend tests send actual, unmodified AML to the backend;
  complete source and artifact expectations remain independent oracles.
- **Independent expected-Rust compilation:** authored Expected workspaces establish
  oracle feasibility. They never substitute for returned compiler output.
- **Actual generated-Rust compilation:** Generated phases compile returned files
  unchanged, with frozen locks and Cargo package/target/artifact provenance.
- **Runtime and publication:** supervised native processes establish lifecycle,
  raw output, invocation snapshots, cleanup, relocation and immutable-generation
  behavior. Generated deployment scripts are not executed.
- **Acceptance:** unfiltered verification checks current inputs before and after
  execution and requires every original named outcome. Focused commands, doctor
  and source-generation success are separate evidence.

The retained runtime contract includes all 98 registry scenarios and six
additional concurrency scenarios. Ordinary native tests do not establish
sanitizer/model-checker equivalence. A recorded host result supplies no claim
for another platform or source identity.
In-process execution, automatic Class lifecycle and cross-compilation remain
outside the supported first-release contract. Private entry visibility has
[AML/native and separate frontend boundaries](../service-execution/coverage.md#private-entry-coverage).

## Current APIs and cache layout

Public callers use `GenerationOptions{IncludeRuntime: ...}`, `WorkspaceRuntime`
and `EvaluateWorkspaceEntry` for Rust source generation and workspace entry handling.

`internal/testsupport/native` intentionally retains the physical cache suffix
`internal/testsupport/rustparity`. This preserves the existing cache location;
it is not a source, resource or API alias. Current cache behavior is documented
in the [build guide](../../docs/cli-build.md).

## Verification policy

Follow the [current integrity and amendment policy](../../docs/development-harness.md#current-integrity-and-amendments)
for protected inputs, assertions, expected bytes, resources and pins. Preserve
the [validation budgets](../../docs/development-harness.md#validation-budgets),
runtime watchdogs and concurrency.

The [local validation pause](../../docs/development-harness.md#temporary-validation-pause)
requires an explicit request for local bootstrap, full lint, full ordinary/native
suites or `scripts/l-check`. Automatic Linux validation on PR creation, new
commits and reopening is authorized; manual dispatch requires explicit
authorization. For requested full local validation, follow the
[development guide](../../docs/development-harness.md#full-validation-when-requested).
Historical evidence never substitutes for the requested fresh gate.
