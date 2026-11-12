# Contributing to L

Start with an [existing issue](https://github.com/l-programming-language/l/issues)
or [open an issue](https://github.com/l-programming-language/l/issues/new/choose)
describing the problem and intended outcome. Bug reports should include a small
reproduction, exact commands, expected and actual output, and the revision and
host used. Change requests should define acceptance criteria and affected
documentation. Every pull request must link to a GitHub Issue, including
documentation, maintenance, and dependency changes.

## Get oriented

Read [AGENTS.md](AGENTS.md), the [architecture](docs/architecture.md),
[build contract](docs/cli-build.md), and
[development and acceptance guide](docs/development-harness.md).
Read the applicable scoped instructions before changing
[the CLI](cmd/l/AGENTS.md), [Lyger](lyger/AGENTS.md),
[Logiceme](logiceme/AGENTS.md), or [builtin assets](builtins/AGENTS.md).

L, Lyger, and Logiceme are Go implementations. Lyger resolves source semantics
and emits AML; Logiceme composes Rust source bundles from AML and canonical
logeme resources. The current executable is one standalone App with core Print. Preserve these
boundaries when choosing where a fix belongs.

## Work on a branch

Use a branch based on current `main` and keep each PR focused on its linked issue.
If using a fork, update from the upstream repository before creating your branch.
For a checkout whose `origin` is this repository:

```sh
git switch main
git pull --ff-only origin main
git switch -c your-change-name
```

Follow [local setup](README.md#local-setup) for the compiler and the
[website guide](website/README.md) for documentation development. Tool pins are
recorded in [.ci-tools.json](.ci-tools.json), [.go-version](.go-version), and
[website/.node-version](website/.node-version); use the retained dependency locks.
Do not update pins or regenerate locks as an incidental part of another fix.

## Code, tests, and protected contracts

Reuse an existing exposing regression when available. Add tests when changed
behavior is not already covered, including relevant failure cases. Keep fixtures
small and assertions meaningful; do not weaken tests to accommodate an
implementation defect. Explain when a documentation or maintenance change needs
no new behavioral test.

During the current core-only reset, only logeme templates are protected by
[harness/contract.json](harness/contract.json). Other fixtures and tests may change
within authorized work, including removing assertions that require retired words
or invalid L. Preserve the exact active hello-world output and the frozen
logeme bytes/interfaces. Do not alter hashes to hide a failure.

Logeme changes need particular care. Read the
[builtin contracts](builtins/CONTRACTS.md), [logeme semantics](builtins/LOGEMES.md),
and [Lucid suite](logiceme/testdata/lucid/README.md) before proposing changes.
Keep `builtins/segments/`, `builtins/templates/`, their interfaces and frozen hash
expectations unchanged. Compose behavior through existing syllables and L
definitions. Native leaves perform single operations; branches, loops, error
policy and resource sequencing remain L algorithms. Preserve typed binding
provenance and canonical descriptor targets. A requested clean break removes
retired words after migrating callers, without compatibility aliases or adapters
unless explicitly required. Keep exact source generation, independent
expected/generated native compilation, runtime verification, and temporary-gate
evidence separate.

## Update documentation with behavior

Update canonical prose under `docs/` with changes to supported syntax, CLI
behavior, runtime semantics, or public contracts. Follow the
[documentation authoring guide](docs/contributing/documentation.md) for routes,
links, canonical examples, and reviewed reference snapshots. Site presentation
belongs in `website/`; generated projections under
`website/src/content/docs/docs/` are not authored source.

Review affected entries in `website/content-map.json` and `docs/_meta/` when
adding pages or changing documented source inputs. Regenerate mechanical
references only after reviewing affected semantics and mappings. A changed hash
alone does not establish that documentation is correct.

## Validate and report evidence

Follow the [validation policy](docs/development-harness.md#validation-authorization).
Explicit task authorization covers its bounded baseline, regression, native and
completion checks without repeated permission requests. Report unrun stages.

During the reset, run `scripts/l-check` for the temporary hello-world gate. It
executes source/compiler checks, real App compilation and execution, unsupported
runtime-effect rejection, selected Print behavior and frozen logeme integrity.
Use pinned tools and existing supervision. Report this reduced scope accurately;
the prior full-module/native/runtime/website suites are not current completion
requirements or passing claims. See the [current gate](docs/development-harness.md#temporary-core-only-gate).

Publication is disabled. Do not dispatch workflows, publish, or deploy without
separate authorization. Automatic PR CI runs the temporary gate only.

## Open a pull request

Complete the PR template and link the existing issue in the body. Use
`Closes #123` when merging the PR should close that issue; use `Refs #123`
for work that contributes without completing it. Describe the problem and
resulting behavior, relevant compatibility effects, documentation changes, and
validation results. Explain any checklist item that does not apply.

For logeme or protected-input changes, link the explicit scope approval and
identify the affected contract and evidence. Report unrun stages honestly;
static inspection is not a new generation, compilation, runtime, or acceptance
result. Review the final diff for unrelated changes before requesting review.
