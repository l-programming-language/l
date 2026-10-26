# Write native-backed definitions

Native-backed definitions bind typed L arguments to logeme compositions containing explicit Rust. They require the target's contracts; L does not infer ownership, lifetimes or concurrency from native text.

## Start from the existing foundation

The retained ProbeIO fixture defines a native operation and exposes it through a dictionary:

<!-- example:native-dictionary-definition -->

`Define` introduces Log and its inflections; Logger is its constructed form. `<DEF>…<DEF>` frames the body inside backticks: both markers have the same spelling. `E` emits the Rust statement; `BOD^E` wraps that Calculation as a declaration item in the composition. `{|Statement: …|}` supplies the Statement metadata; `~…~` encloses its literal native text. These delimiters are logeme syntax, not HTML.

Its caller imports the dictionary member and supplies App-body work:

<!-- example:native-dictionary-use -->

This excerpt requires sibling `probe.l`; the [complete caller](../../cmd/l/test/runnable_service/basic.l) ends with Build, Run, Await and Show. The Text deliberately includes a newline: this test-owned probe writes exact bytes, whereas public IO.Print adds a newline. The [implicit core](../../builtins/eng/core.l) is already loaded; do not redeclare it.

## Typed slots cross an encoding boundary

`<@Output|Thing@>` declares the argument label and binding kind. Here the caller supplies Text: this probe uses Rust’s `.as_bytes()`, so it does not implement every scalar kind that Thing can bind. `<@Process@>` in the core instead refers to an already instantiated definition; omitting the type does not mean Text. [Typed slots](../spec/definitions.md#typed-slots) lists the supported kinds and dotted-name rules.

Text receives protected Rust quoting; Number uses finite binary64 encoding. Native fragments retain target meaning. Literal source, diagnostic views and AML are distinct: do not normalize escapes or format AML as XML.

## Read a complete backend witness

The [declaration/evaluation witness](../../logiceme/testdata/lucid/declaration-evaluation.aml.txt) is AML, not L. [Alphabet and composition](../reference/logemes/alphabet.md) explains its glyphs, `^` integration, `*` siblings and parentheses. The returned Executable is a source bundle; materialization, Cargo compilation and execution are separate steps. Lucid's native-compilation checks do not establish general runtime behavior.

## Stay inside the approved surface

The eighteen Rust segments are frozen. Lifetime/generic inputs are glyph-local native interiors, not L generics, Func/Prop/Op fields or inherited defaults. Rust rejects Deferred metadata; scheduling supplies no automatic deferred cleanup.

Before extending a template, read [definitions](../spec/definitions.md), [builtin contracts](../../builtins/CONTRACTS.md), [Rust forms](../../builtins/README.md) and [AML](../../logiceme/AML.md).
