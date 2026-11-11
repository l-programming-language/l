# Working on `l`

L, Lyger and Logiceme are Go implementations. Lyger resolves source semantics and emits AML; Logiceme renders Rust from the selected source definitions and frozen logeme assets. Do not introduce an interpreter or generate replacement L algorithms in Go or Rust.

<a id="current-contract"></a>

## Current reset boundary

The authored standard libraries and all builtin `.l` files except `builtins/eng/core.l` have been removed. Core loads directly. The root `dictionaries.json` has no bundled dictionaries. Do not restore Storage, IO, Formatting, Runtime, prelude, generated runtime AML, or native runtime implementations from history or caches. Runtime will be rebuilt in L in subsequent work.

The temporary supported executable is one standalone App with a core Print body. Print accepts Text and writes it followed by one newline. The CLI compiles the actual App source bundle directly. Runtime-dependent orchestration, process lifecycle, capture and output inspection are unavailable. Reject unsupported effects explicitly; never discard them to make a source appear successful.

Keep core small. Native leaves perform one bounded call, operator, projection, constructor, pattern or return. Sequencing, branches, iteration, error policy and resource selection belong in L. Never hide an algorithm or runtime in Statement, EA, an opaque closure or a native helper. Print's single native output call is not a runtime implementation.

Only logeme templates are protected during this reset: preserve all bytes and interfaces in `builtins/segments/` and `builtins/templates/`, including the frozen hash expectations. Other fixtures, words and tests may be replaced or removed for the authorized reset. Tests that enforce rejected syntax or word designs are not language requirements. Preserve useful behavioral observations by rewriting their source callers when those capabilities are rebuilt. Do not manufacture passing outcomes or retain compatibility implementations for retired words.

## Temporary acceptance

Run `scripts/l-check`. It invokes the existing supervised runner's `hello` stage: the restored hello-world compiler/source tests, actual standalone App build/execution, rejection of runtime effects, selected core Print behavior, and frozen logeme checks. Use the pinned tools and existing subprocess deadlines. Reports must say **temporary hello-world gate**, not full acceptance.

The old unfiltered, tagged-native, website, and runtime suites are outside this gate. They remain historical/regression material and may fail against the intentionally removed libraries. Do not skip tests inside the selected gate or claim unrun suites passed. `scripts/l-ci acceptance`, `native` and `check` are not completion requirements for this reset. Publication stays disabled; do not push, deploy, publish or dispatch workflows without authorization.

## Source and compiler design

Read scoped instructions in [lyger](lyger/AGENTS.md), [logiceme](logiceme/AGENTS.md), [cmd/l](cmd/l/AGENTS.md), and [builtins](builtins/AGENTS.md) before changing them. The [architecture](docs/architecture.md) and [build guide](docs/cli-build.md) identify the reset and retained historical capabilities.

Preserve source-unit ownership, lexical scope, typed parameters, once-only evaluation and provenance. Keep explicit imports confined to their selected provider/root. No CWD fallback, mirrored resources or second resolver. The compiler uses embedded assets or one complete `--assets` provider without changing CWD.

Dot-qualified names are not valid authored L. Use grammatical possessives or explicitly scoped imports; internal compiler identity strings are separate. Prefer one possessive for a declaration. Do not replace dotted qualification with unnecessarily deep ownership chains. Classify each meaning and declare its appropriate forms before choosing construction syntax. Existing source or tests do not prove a word design is correct.

Use [l-author-word](.agents/skills/l-author-word/SKILL.md) for source vocabulary, [l-review-words](.agents/skills/l-review-words/SKILL.md) for independent review, [l-acceptance](.agents/skills/l-acceptance/SKILL.md) for the current gate, and [l-frontend-trace](.agents/skills/l-frontend-trace/SKILL.md)/[l-logeme-handoff](.agents/skills/l-logeme-handoff/SKILL.md) for compiler diagnosis. Historical skill references to deleted libraries do not authorize restoring them.
