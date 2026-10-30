# Specification source audit

This audit records documentation decisions against the inspected checkout. The public prose is an explanatory consolidation of maintained contracts and observed implementation. It has not received independent maintainer or standards-body ratification. Evidence classifications here identify reviewed source and retained witnesses, not an acceptance result. Validation of a proposed compiler revision is recorded separately.

## Evidence policy

Grammar recognition, semantic binding, native source generation, native compilation, process runtime, and full acceptance are distinct stages. A source hash establishes which bytes were reviewed; it does not establish correctness or execution. Historical integration outcomes remain tied to their recorded revision and host.

This source review follows the selected builtin compositions and frontend binding/type modules. Retired native-body factories and their source-string test witnesses are excluded. The refresh leaves grammar/logeme contracts and execution-evidence classifications unchanged.

The public prose must not turn implementation gaps into intended guarantees. The semantic manifest records role, support, source paths, witness paths, and SHA-256 dependencies. Existing maintained architecture, CLI/runtime, development policy, and error-model paths are preserved.

## Reviewed discrepancies

| ID | Finding | Documentation disposition | Primary evidence |
| --- | --- | --- | --- |
| SPEC-001 | Byte/Bytes names byte-oriented storage; runtime Bytes copies Text and supports checked indexed Number reads. | No invented 8-bit scalar or mutation API. | `lyger/runtime_values_source.go`, `builtins/eng/core.l`, `lyger/runtime_values_test.go` |
| SPEC-002 | Pointer is a typed managed handle; clones share a target while construction copies the initial value. | Document checked access and cycle rejection, not raw addresses or arithmetic. | `stdlib/eng/storage/dictionary.l`, `lyger/runtime_values_source.go`, `lyger/runtime_values_test.go` |
| SPEC-003 | Ordinary Null initializes an explicit value and displays as null; unassigned storage remains separate. | Preserve the separate native-definition Null atom. | `lyger/runtime_values_plan.go`, `builtins/eng/core.l`, `lyger/definition_framing_test.go` |
| SPEC-004 | List and Map have typed runtime construction, updates and reads. Dictionary remains a vocabulary namespace. | Document bounds, type checks and missing keys; do not imply general iteration. | `lyger/runtime_values_source.go`, `lyger/runtime_values_plan.go`, `lyger/runtime_values_test.go` |
| SPEC-005 | While and repeated CONDITION grammar exist; outer instructions reject While and multiple condition joins. | Mark recognized forms rejected at outer lowering; no short-circuit or source-loop contract. | `lexicon/eng/grammar.json`, `lyger/instructions.go`, `lyger/orchestration_plan.go` |
| SPEC-006 | Grammar ` | ` is sequencing, `*` optionality, and square brackets multiple association; familiar EBNF readings would invert the meaning. | Explain the exact decoder and keep machine grammar as the exact reference. | `lyger/utilities.go`, `lyger/grammar_terms.go`, `lyger/grammar_match.go` |
| SPEC-007 | Function declarations bind invocation verbs, typed signatures, private bodies and captures through canonical builtin/logeme composition. | Keep source semantics in Lyger and native rendering in Logiceme; opaque native contracts remain explicit. | `builtins/eng/core.l`, `lyger/source_functions.go`, `lyger/function_composition.go`, `lyger/function_invocation_words_test.go` |
| SPEC-008 | Vocabulary Import/Export words do not establish a replacement for actual Use-based import scanning; structured index is the provider source of truth. | Document supported Use syntax and classify the other configured action roots as partial/unresolved. | `lyger/builder.go`, `lyger/ir_kinds.go`, `stdlib/index.json` |
| SPEC-009 | Numeric parsing can recognize nonfinite values while native scalar/definition boundaries reject them with distinct categories. | Document stage-specific rejection, binary64 rounding, signed zero, range errors, and underflow. | `lyger/types.go`, `lyger/literal_atoms.go`, `lyger/definition_number_test.go` |
| SPEC-010 | General Text retains raw bytes; quote framing and definition escape decoding are separate. | No C/JSON escape model for ordinary Text; preserve invalid UTF-8 and literal delimiters in inclusions. | `lyger/source.go`, `lyger/definition_bindings.go`, `lyger/raw_text_test.go` |
| SPEC-011 | Native typed-IR comparison tests establish native operator behavior but do not alone verify every configured source comparison phrase. | Label grammar/source/operator evidence independently; retain exact unusual inclusive-order grammar spelling. | `lexicon/eng/grammar.json`, `lyger/orchestrator_test.go`, `lyger/orchestration_plan.go` |
| SPEC-012 | Standalone Service declarations remain structural and produce no packages or processes. | Describe App-only lifecycle and retained App-constructed libraries. | `docs/cli-build.md`, `lyger/workspace_units.go`, `builtins/eng/core.l`, `testdata/service-execution/coverage.md` |
| SPEC-013 | Native private entry rejection exists, while the source language does not expose a corresponding Private entry declaration form. | Preserve the native/frontend coverage distinction. | `docs/cli-build.md`, `testdata/service-execution/coverage.md` |
| SPEC-014 | Logical alphabet documentation includes target-independent design possibilities; the concrete Rust surface is frozen and bounded. | Separate model from operational forms; no four-letter inference, arbitrary vowel concatenation, or Deferred semantics. | `builtins/LOGEMES.md`, `builtins/README.md`, `builtins/CONTRACTS.md` |
| SPEC-015 | IO.Print uses private native generic dispatch, but the public source-call scalar path is narrower than arbitrary Rust Display. | Describe exact byte/number behavior and private helpers; no universal source Print API. | `stdlib/eng/io/dictionary.l`, `lyger/raw_text_test.go`, `lyger/io_numeric_test.go` |
| SPEC-016 | An imported dictionary owns a private lexical root and canonical export. Enumerating words in `.l` source cannot establish public visibility. | Use index-driven inventory plus reviewed export mappings/hashes and real frontend witness paths. | `lyger/source_units.go`, `lyger/lexical_scope.go`, `lyger/dictionary_ownership_test.go` |
| SPEC-017 | Attribute, Property and Operation remain structural/internal categories. | Keep source organization separate from runtime value operations. | `lyger/source_hierarchy.go` |
| SPEC-018 | Standalone Service bodies, native lifecycle operations and implicit runtime members reject; only Apps own continuations. | Preserve Main/Basic, primary Service topology, additional App-owned libraries and raw AML library/bin controls. `service-executable-correction.md` and `service-executable-amendment.json` retain the approved scope and assertion mapping. | Author approval; `lyger/instructions.go`, `lyger/paragraph_instructions.go`, `lyger/workspace_units.go`, `cmd/l/build.go` |

## Standard-library expansion contract

New libraries are expected. Documentation must enumerate every actual `(language, key)` in `stdlib/index.json`, not hardcode IO as the architecture. Each entry requires a meaningful dictionary page, reviewed public export set, exact entry/include dependencies, visibility witnesses where available, and support classifications.

A recommended stable identity is `stdlib:<language>:<key>` for a dictionary and `stdlib-export:<language>:<key>:<canonical-name>` for an export. Renaming/removing a key must invalidate the old mapping. Changed source bytes must invalidate review hashes. An unresolved public surface is a failed documentation classification check, not an invitation to classify every regex-found declaration as public.

The current dictionary exports IO and IO.Print; private rendering and writing support remains classified as implementation detail. Future dictionary metadata may be obtained through a separately invoked narrow exporter using real frontend resolution if reviewed mappings are no longer sufficient. It must not silently run compiler/native suites during the Node-only website build.

## Existing witnesses and validation status

The semantic manifest lists inspected test files, not current pass results. The integration ledger's dated Linux/macOS records are historical evidence for those exact revisions. Website schema/link/hash checks validate documentation integrity only. Compiler bootstrap, full-module lint, ordinary/native suites, `scripts/l-check`, and `scripts/l-harness verify` were deliberately not invoked by this work.

## Maintenance decisions

Shared review hashes apply to every dependent mapping. Refreshing a hash requires reviewing each affected classification; a passing website check does not validate compiler behavior.
