---
name: l-acceptance
description: Run and interpret the repository acceptance gate and verify current protected inputs, expected bytes and original test outcomes. Use for baseline checks, progress verification and completion claims.
---

# Check acceptance

The [temporary validation pause](../../../docs/development-harness.md#temporary-validation-pause) governs this workflow. Expensive local validation requires explicit user authorization; configured pull-request events already authorize hosted CI. Follow that canonical policy for requested runs and report unrun stages accurately.

Read [development and acceptance](../../../docs/development-harness.md). Use `scripts/l-check` for completion and `scripts/l-harness` for its acceptance stages; do not duplicate the harness runner or manifest logic. Bootstrap pinned tools/dependencies with `scripts/l-ci bootstrap` when needed.

1. Run `doctor` for a new environment or suspected setup/resource failure.
2. Run `scripts/l-check` for completion: workflow lint, both complete Go lint profiles, doctor/verify, and tagged native compilation. Resolve all warnings/errors. Verification builds the module, discovers every package once and runs both unfiltered test phases with integrity checks before and after.
3. Inspect setup/build status, protected inputs, absent paths, package assignments and test events. Every discovered package and named outcome must complete successfully; missing, skipped or aborted tests fail acceptance.
4. Diagnose artifact mismatches with raw-byte `diff EXPECTED ACTUAL`, focused existing tests or `diagnose --frontend`. Preserve nil/empty distinctions; expected helpers outside `_test.go` remain protected.
5. Obtain a fresh full `scripts/l-check` after repair. Probes, explicit vet, focused tests and generation-only evidence do not replace completion. Reject missing, malformed, stale, skipped or incomplete validation evidence.
6. The completion command always includes the separately tagged native stage in the [maintained Rust suite](../../../logiceme/testdata/lucid/README.md). The ordinary gate covers backend generation/freeze plus frontend/runtime native tests; it does not run that separate complete-AML Cargo matrix. Keep the suite toolchain, frozen locks and subprocess deadlines; compile actual returned files unchanged and execute Rust orchestrators and Apps only within the authorized supervised tests; never run deployment scripts.

Report revision/dirty state when available, host, run location, command status, failed package/test names or success, integrity and limitations. Never regenerate expectations/hashes to turn failures green. Protected behavioral inputs and canonical resources require explicit approval for a finite amendment. Historical records belong in Git history, not acceptance prerequisites. Cross-host claims require separate native evidence.

Establish a failing regression before a behavioral repair; reuse an existing exposing test. The [Rust integration limits](../../../builtins/README.md#current-integration-limits) distinguish verified contracts from retained support. The suite guide owns coverage; historical matrices and leftover compiler tables add no obligations. For documentation-only work, check links and preserve code/assets/tests/index; authorized full validation still uses `scripts/l-check`. Do not restrict lint to changed lines, add blanket exclusions or omit native validation. Preserve the [finite CI budgets](../../../docs/development-harness.md#validation-budgets), runtime watchdogs and concurrency; further amendments need explicit scope. Suppressions must name the linter and preserved invariant or intentional best-effort behavior.
