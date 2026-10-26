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
[CLI/runtime contract](docs/cli-build.md), and
[development and acceptance guide](docs/development-harness.md).
Read the applicable scoped instructions before changing
[the CLI](cmd/l/AGENTS.md), [Lyger](lyger/AGENTS.md),
[Logiceme](logiceme/AGENTS.md), or [builtin assets](builtins/AGENTS.md).

L, Lyger, and Logiceme are Go implementations. Lyger resolves source semantics
and emits AML; Logiceme composes Rust source bundles from AML and canonical
logeme resources. Apps run in separate OS processes. Preserve those boundaries
when choosing where a fix belongs.

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

[harness/contract.json](harness/contract.json) protects current inputs, canonical
resources, assertions, expected-value helpers, and literal expected bytes.
Changing a protected item requires explicit author approval identifying the
finite scope and reason. Earlier amendments are not standing permission.
Follow the [integrity policy](docs/development-harness.md#current-integrity-and-amendments);
never regenerate expectations or hashes just to make a failure pass.

Logeme changes need particular care. Read the
[builtin contracts](builtins/CONTRACTS.md), [logeme semantics](builtins/LOGEMES.md),
and [Lucid suite](logiceme/testdata/lucid/README.md) before proposing changes.
The 18 Rust segments and their interfaces are frozen; any amendment must be
explicitly scoped. Prefer composition through existing syllables and definitions.
Identify changed segments, metadata interfaces, and support assets in the issue
and PR. Keep exact source generation, independent expected/generated native
compilation, runtime verification, and full acceptance evidence separate.

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

**Expensive local validation is temporarily paused unless explicitly requested.**
Do not automatically run bootstrap, full-module lint, full ordinary/native
suites, or `scripts/l-check`. Use lightweight inspection appropriate to the
change and state what was not run. See the
[current validation policy](docs/development-harness.md#temporary-validation-pause).

When full local validation is explicitly requested, run from the repository root:

```sh
scripts/l-ci bootstrap
GOMAXPROCS=4 scripts/l-check
```

The full command requires workflow lint, both complete Go lint profiles,
harness doctor and unfiltered acceptance, and the tagged native-compilation
stage. Resolve every lint warning/error. Focused checks do not replace full
acceptance, and an earlier passing revision does not validate new changes.
Preserve test discovery, runtime watchdogs, concurrency, and the
[finite validation budgets](docs/development-harness.md#validation-budgets).

Pull-request events run hosted Linux validation automatically. The existing
`L CI` gate, documentation checks, dedicated lint/security checks, and issue-link
check provide separate results. Website checks validate content and the static
site; they do not establish compiler or native-runtime correctness. Linux
results do not establish macOS validation. Consult the workflow logs and retained
reports for the current revision, and distinguish failures, skipped checks,
and checks still running. The development guide owns branch-protection policy.

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
