# Logemes and the Rust target

A logeme is a small logical unit represented by a symbol and a native template. Syllables compose those units; sequences connect syllables. Lyger lowers resolved L source to AML, and the Go implementation called Logiceme interprets that framed input into Rust source bundles. Source generation is not compilation or execution.

## Three contracts to keep separate

The [alphabet](alphabet.md) is a target-independent design model. The [Rust asset reference](../../../builtins/README.md) defines the bounded operational surface. The [Lucid corpus](../../../logiceme/testdata/lucid/README.md) supplies exact-generation, native compilation and focused future-execution witnesses. None implies that every possible syllable or every parser-recognized L phrase is executable.

| Reference | Read it for |
| --- | --- |
| [Alphabet and composition](alphabet.md) | Logical ownership, syllable positions, integration and grouping. |
| [Declarations and definitions](declarations.md) | A/B, O/F, adapters and complete-file escape. |
| [Calculations and descriptions](calculations.md) | E/D, U/H, native values, operations and fields. |
| [Selection](selection.md) | I/G, keyed GI and bounded CI/BI readiness selection. |
| [Callables](callables.md) | W/K, FW, BW and named/anonymous KY. |
| [Iteration](iteration.md) | X/P, condition bindings, traversal and labels. |
| [Instantiation and scheduling](instantiation.md) | Y/M, records, traits, V/C and Deferred rejection. |
| [Metadata](metadata.md) | Func/Prop/Op, absence, provenance and local generic/lifetime inputs. |
| [AML and source bundles](aml.md) | Framing, hierarchy, source ownership and Evaluate. |
| [Evidence and support](support.md) | Maintained corpus, malformed families and integration limits. |

## Evidence status

The descriptions here are source-inspected implementation reference, informed by maintained contracts and existing tests. No backend generation, native compilation or runtime execution is performed by a normal website build. Historical green results in the [ledger](../../../logiceme/testdata/lucid/FAILURES.md) apply to their recorded revision and host.

Every selected segment and supported case family is classified in the documentation coverage manifest. This is an implementation inventory and audit aid, not independent maintainer ratification of a new language standard.
