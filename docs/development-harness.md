# Development and acceptance

## Temporary core-only gate

Run from the repository root:

```sh
scripts/l-check
```

This invokes the existing supervised runner's `hello` stage. The **temporary hello-world gate** covers restored hello-world source/compiler tests, actual standalone App compilation and execution, runtime-effect rejection, selected core Print behavior and frozen logeme checks. Reports are written under `.harness/runs/ci-hello-*`.

Require complete named outcomes, exact expected output and successful subprocess exits. Empty selections, skipped tests, stale artifacts, unavailable tools and malformed reports are not passes. Report generation, native compilation and execution separately where their outcomes differ.

Broader compiler, tagged-native and website suites are outside this gate. Retained tests may cover compiler representations whose executable support has not been rebuilt. `scripts/l-ci acceptance`, `native`, `check` and the old harness `verify` are not completion requirements for this reset. Do not report unrun suites as passing.

## Validation authorization

Task authorization covers its necessary bounded validation. Continue authorized checks without repeated permission requests. A documentation check establishes documentation consistency, not native behavior. Historical results apply only to their recorded revision and inputs.

Do not push, deploy, publish or dispatch workflows without authorization. Publication is disabled during the reset. Existing workflow files and historical release procedures do not authorize activation.

## Current integrity and amendments

Only `builtins/segments/` and `builtins/templates/` are protected during this reset: preserve every byte, interface and frozen hash expectation. The inventory contains eighteen Rust segments and four templates. `scripts/l-harness frozen` checks these protected assets.

Other tests, fixtures, words and documentation may change within authorized reset work. Tests enforcing retired words or invalid syntax are not language requirements. Preserve useful behavioral observations with valid replacement callers when capabilities are rebuilt. Never manufacture passing outcomes or restore compatibility implementations for removed libraries.

## Validation budgets

Use the pinned tools and existing subprocess deadlines, watchdogs and concurrency. Do not relax a timeout or suppress an assertion to obtain a pass. Tool versions are recorded in [.ci-tools.json](../.ci-tools.json) and [.go-version](../.go-version): Go 1.27.3 and Rust 1.98.1.

Missing prerequisites are setup failures. Do not silently install tools, restore old Runtime assets or replay historical evidence. Preserve the original failure and report downstream stages that could not run.

## Tool preparation

Run `scripts/l-ci bootstrap` explicitly when the pinned tools or dependencies are
missing. It installs pinned tools and fetches dependencies from retained locks.
The remaining Cargo seeds cover retained compiler fixtures and independent Lucid
compositions. The optional compiler dependency-lock fragment remains unchanged,
but has no active manifest to fetch during this reset. Setup does not build or
restore the removed Runtime. Preparation is not test evidence.

## Compiler defaults

Embedded English resources are the default. Core loads directly; [dictionaries.json](../dictionaries.json) has no bundled dictionaries. There is no prelude, Storage, IO, Formatting or Runtime distribution to regenerate.

| Override | Default | Meaning |
| --- | --- | --- |
| `--assets ROOT` | Embedded resources | Select a complete provider containing the registry, lexicon and retained builtin/logeme assets. |
| `--language LANGUAGE` | `eng` | Select source-language resources within that provider. |

Options precede the source filename. Relative paths use invocation CWD without changing it. An explicit provider has no embedded fallback, host-file search or mirrored resource copy. Missing resources fail initialization or resolution at their owning boundary.

```sh
l build --assets . --language eng demo/basic/main.l
```

Use a compiler built from this checkout. An installed historical compiler can embed removed vocabulary and is not authority for current support. Offline contract inspection must not regenerate resources, invoke native tools or fetch dependencies to repair missing evidence.

## Investigating failures

Follow the failing boundary from source preparation and imports through tokens, grammar, bindings, AML and native source. Keep raw source, expanded source and output artifacts distinct. Preserve diagnostic source spans and native causes.

Use the repository's [frontend tracing](../.agents/skills/l-frontend-trace/SKILL.md), [logeme handoff](../.agents/skills/l-logeme-handoff/SKILL.md) and [acceptance](../.agents/skills/l-acceptance/SKILL.md) workflows as appropriate. Record the dirty state, host, exact command, report location and actual failures.

## Rust verification boundaries

Frozen syntax templates remain available to the backend. Their retained tests and [Lucid documentation](../logiceme/testdata/lucid/README.md) distinguish generation, native compilation and bounded execution. This does not restore source libraries or make the historical suites current acceptance requirements. Generated installer/start/stop scripts remain artifacts and are never run by the temporary gate.
