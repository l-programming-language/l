# L Language Specification

This reference describes retained source syntax and compiler contracts. The current executable boundary is one standalone App with core Print accepting Text. The standard libraries and Runtime have been removed; configured grammar or a native template does not make their former behavior available.

## Scope and conformance

Lyger resolves source into AML, Logiceme renders Rust and the CLI builds the selected App directly. Building does not run it. The [first program](../guide/first-program.md), [build contract](../cli-build.md) and [temporary hello-world gate](../development-harness.md) define the supported path.

## Reading the specification

- [Lexical structure](lexical-structure.md), [names and scope](names-and-scope.md), and [grammar](grammar.md) describe source interpretation.
- [Definitions](definitions.md) and [imports](imports.md) describe vocabulary ownership and composition.
- [Primitive](values.md), [collection](collections-and-categories.md), [lexical](lexical-types.md), and [applicational](applicational-types.md) types distinguish compiler categories from executable support.
- [Control flow](control-flow.md), [execution](execution.md), and [errors](errors.md) identify current limits.
- The [glossary](glossary.md) explains terminology.

## Role, support, and evidence

The [Native Words](../reference/words.md) index and [Word Glossary](../reference/word-glossary.md) describe retained meanings. There are no bundled standard-library dictionaries. Native composition follows the separate [logeme](../../builtins/LOGEMES.md) and [Rust](../../builtins/README.md) contracts.

Static inspection, generation, native compilation and execution are separate evidence stages. Retained tests are not fresh passing results. See [evidence labels](../contributing/evidence.md).

## Language guarantees and target limits

Preserve source ownership, lexical scope, canonical identity, typed arguments, once-only evaluation and provenance. Dot-qualified authored names are invalid. Native builds target macOS or Linux; runtime lifecycle, capture and output inspection are unavailable.

## Sources and review

See [architecture](../architecture.md), [validation](../development-harness.md) and the [compiler error model](../error-model.md). Reviewed [semantic mappings](../_meta/semantics.json) classify the retained inventory without widening the executable boundary.
