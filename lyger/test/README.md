# Lyger compiler fixtures

Retained tests exercise source preparation, lexical binding, imports, definition
construction, typed AML and source hierarchy through current compiler APIs. They
must use actual returned AML rather than substitute expected Rust or implement L
algorithms in helpers.

Useful independent controls include configured imports and provider isolation,
literal/framing preservation, paragraph scope and captures, descriptor identity,
and recursive Class/Object lowering. These compiler observations do not establish
that removed library words can execute.

`builds/` retains manifest/lock pairs for native compiler fixtures. A helper may
attach a lock only after the actual manifest matches its independently maintained
expectation. `native/` retains the locked test dependency seed used by native
build helpers. These resources are not bundled Runtime implementations.

The current executable boundary is a standalone App using core Print. Run
`scripts/l-check` for the [temporary hello-world gate](../../docs/development-harness.md#temporary-core-only-gate).
It is not a full frontend or runtime suite claim. Use focused generic compiler
checks when changing their corresponding behavior, preserving pins, deadlines,
exact expected bytes and meaningful negative cases. Report source analysis,
generation, native compilation and execution separately.
