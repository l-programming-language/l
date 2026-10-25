# Write native-backed definitions

Native-backed definitions are an advanced extension boundary. They connect typed L arguments to logeme compositions and explicit Rust fragments. They require knowledge of the maintained backend contracts; the language does not infer arbitrary ownership, lifetime or concurrency behavior from embedded text.

## Start from the existing foundation

<!-- example:builtin-core -->

This is the canonical builtin unit displayed as a reference fragment, not a program to paste into a project. The compiler already loads it. Process supplies a syllable sequence; Function packages callable metadata; Return contributes native statement text; descriptors compose Main and Basic.

The outer quoted `<DEF>` markers identify a definition body. The framing helper requires one complete outer pair, optionally enclosed in complete backticks. Explicit internal template parsing has its own markerless boundary. Ordinary Text containing incomplete or interior markers remains ordinary data; the documentation must not reinterpret it as HTML.

## Typed slots cross an encoding boundary

An encoded slot such as `<@Output|Thing@>` carries a label and type. Binding substitutes according to that declared argument kind. Native fragments are not interchangeable with source Text: Text-to-Rust expressions use protected quoting, while native bodies retain their explicit target meaning.

Number arguments use the shared finite binary64 encoder. Literal preservation, expanded diagnostic views, and completed AML are different representations. Do not “clean up” escaping in an example or parse AML with an XML formatter; doing so can change literal content or backend delimiters.

Slot labels accept the implementation's dotted identifier contract. A valid L label is not automatically a valid Rust identifier, and a Rust keyword is not automatically forbidden as an L label. Each boundary validates its own syntax.

## Read a complete backend witness

<!-- example:aml-declaration-evaluation -->

This is a complete maintained AML input for a backend generation/native-compilation witness, presented here as a fragment of the overall toolchain. The language is AML, not L. The `W`, `FW`, `FOB`, `GA`, `I` and related pieces participate in the frozen logeme system; their composition roles belong in the [concrete reference](../reference/logemes/index.md).

The backend's returned Executable is a source bundle. A caller must validate and materialize files, invoke Cargo, select actual compiled artifacts, and execute a program separately. Lucid's compile-only evidence does not establish runtime behavior for every composition.

## Stay inside the approved surface

The eighteen Rust segments are frozen by the [Lucid suite](../../logiceme/testdata/lucid/README.md). Native lifetime/generic interiors are only accepted on their approved glyph-local forms. They are not L-language generics or ownership inference, and they do not belong to arbitrary Func/Prop/Op fields or inherited Class/Object defaults.

Rust templates exclude Deferred and reject supplied Deferred metadata. There is no automatic deferred-call cleanup contract to infer from scheduling. Unsupported composition, closure or asynchronous forms must stay classified as unsupported or design material instead of becoming tutorial promises.

Before extending a template, review [builtins contracts](../../builtins/CONTRACTS.md), [the Rust reference](../../builtins/README.md), and [AML](../../logiceme/AML.md). Documentation discrepancies authorize a documented audit item, not a compiler or frozen-fixture repair.
