# Standalone Service retirement validation — October 15, 2026

The explicitly authorized macOS ARM64 run passed bootstrap and the canonical
`scripts/l-check` gate. Both publication regressions reached their final
preservation and runtime assertions. These results describe the tested source;
this documentation update did not run language validation.

## Tested source and results

- Revision: `de4378d831d3ba36eb7342ff394e9d49e5ffb059`, with the approved changes
  in the working tree.
- Source identity: `38096646d1a9cc5a68ea92ba2ea34d4997b526332e1cf0b1085a53911d8b44f5`.
- Host/toolchain: macOS ARM64, Go 1.27.1 and Rust 1.98.1.

| Stage | Recorded result |
| --- | --- |
| Bootstrap | Pinned tools, dependencies and prerequisites passed |
| Lint | Formatting, workflow lint and both complete Go lint profiles passed |
| Unfiltered acceptance | All 4,383 named outcomes across both phases passed, including all sixteen original required outcomes |
| Integrity | All 508 checks passed before and after acceptance, including 386 protected hashes and eleven required-absent paths |
| Source generation | Retained frontend/backend and exact-source observers passed |
| Runtime and publication | App lifecycle, captured output, continuations, Stop, relocation, concurrent publication and both generation-preservation regressions passed |
| Tagged native compilation | The complete separate stage passed, including `TestLucidNativeCompile` and `TestNativeMetadataArrowsCompile` |

Generation, independent expected-source compilation, generated-source compilation
and supervised runtime are separate evidence. The [coverage map](coverage.md)
identifies App process, structural Service, App-constructed library and raw AML
observers. Generated deployment scripts remain unexecuted. This macOS result
does not establish Linux validation for the tested source.

## Recorded reports

The existing reports are under `.harness/runs/`:

| Stage | Report |
| --- | --- |
| Bootstrap | `ci-bootstrap-4236559208/stage.json` |
| Lint | `ci-lint-3784724979/stage.json` |
| Acceptance | `ci-acceptance-1667428328/verify.report.json` |
| Tagged native compilation | `ci-native-477355706/stage.json` |

The [finite amendment](../../docs/_meta/service-executable-amendment.json) retains
the exact approved scope, hashes and assertion mapping. Tool pins, runtime
watchdogs, suite budgets and concurrency remain unchanged.
