# Write native-backed definitions

Build words from existing words, with logemes supplying small target-specific foundations. Native leaves must obey the target's type, ownership and lifetime contracts.

## Start from the existing foundation

The [core](../../builtins/eng/core.l) defines Function through a logeme composition. Objective builds on Function; Main and Basic compose the entry structure. Function bodies use Process and the selected Return member. These words are already available without import.

[IO.Print](../../stdlib/eng/io/dictionary.l) follows the same pattern: it composes typed Render functions with Write. Extend vocabulary through reusable nouns, verbs and descriptors before introducing a native leaf.

## Typed slots cross an encoding boundary

The retained ProbeIO fixture illustrates a small native leaf exposed through a dictionary:

<!-- example:native-dictionary-definition -->

Log declares its inflections; Logger is its constructed form. `<DEF>…<DEF>` frames the logeme body. `BOD^E` supplies a bounded native statement, not an alternative place to implement composed language behavior.

Its caller imports the dictionary member and supplies App-body work:

<!-- example:native-dictionary-use -->

This excerpt requires sibling `probe.l`; the [complete caller](../../cmd/l/test/runnable_service/basic.l) ends with Build, Run, Await and Show. The probe writes exact Text bytes; public IO.Print adds a newline.

`<@Output|Thing@>` declares a typed argument; this leaf's `.as_bytes()` call requires Text. `<@Process@>` instead binds an instantiated definition. See [typed slots](../spec/definitions.md#typed-slots) for the supported kinds.

Text receives protected Rust quoting; Number uses finite binary64 encoding. Preserve literal bytes and use definition composition for executable behavior.

## Read a complete backend witness

The [declaration/evaluation witness](../../logiceme/testdata/lucid/declaration-evaluation.aml.txt) shows the AML backend interface. [Alphabet and composition](../reference/logemes/alphabet.md) explains its glyphs and grouping. Generation returns a source bundle; native compilation and execution are separate steps.

## Stay inside the approved surface

The eighteen Rust segments are frozen. Lifetime/generic inputs are glyph-local native interiors, not L generics, Func/Prop/Op fields or inherited defaults. Rust rejects Deferred metadata; scheduling supplies no automatic deferred cleanup.

Compose within the existing [definition](../spec/definitions.md), [builtin](../../builtins/CONTRACTS.md), [Rust](../../builtins/README.md) and [AML](../../logiceme/AML.md) contracts.
