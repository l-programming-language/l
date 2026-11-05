# Write native-backed definitions

Compose behavior in L and use native leaves for individual target operations. A leaf can perform one call, operator, constructor, borrow/projection or bounded pattern/binding. Keep branching, iteration, sequencing and error policy in the words that compose those leaves.

## Start from the existing foundation

The implicit [core](../../builtins/eng/core.l) supplies Process, Function, Objective, Main and Basic. [IO](../stdlib/eng/io.md) adds reusable byte operations through an explicit import. This excerpt shows how its Send Function separates byte projection from writing:

<!-- example:source-send -->

Borrowed names the byte view produced from Payload. Deliverer performs one `write_all` call with Writer and Borrowed; Return passes its IO Result to the caller. Send adds no newline, lock or panic policy. Its native signature borrows the writer exclusively and the payload through a shared reference. Rust checks those ownership requirements.

This dictionary fragment requires its helpers and a native caller. `Storage.Routine` accepts explicit native generic parameters: `W` and `P` are supplied type atoms, not inferred by Generic. See the [definition contract](../spec/definitions.md#process-function-and-return).

## Compose a wrapper

<!-- example:source-lined -->

Lined describes IO's Sender through the private Terminate Function. In its rule, `any IO's Sender` denotes the retained base invocation, and `any Thing` forwards its writer. The base result becomes Receipt. Result.Select runs the newline send only after payload success; failure returns through the other Process. The wrapper keeps the same writer rather than acquiring another handle.

Callers import the public Send/Lined words; Terminate remains an implementation dependency. `New IO.Lined IO's Sender` evaluates the original operands once and preserves their ownership. A wrapper may use those references; it cannot silently clone a consumed payload. The [IO contract](../stdlib/eng/io.md#explicit-byte-output) owns exact newline, partial-write and failure behavior.

Keep each policy at the level that owns it. IO.Print composes rendering, writer/guard acquisition, Lined sending and an explicit fail-fast result check. A reusable Send leaves those choices to its caller.

## Typed slots cross an encoding boundary

Native definitions use one complete backtick-delimited `<DEF>…<DEF>` frame. [Typed slots](../spec/definitions.md#typed-slots) bind checked values, names, types or compositions into approved AML fields. A Label names a result, a Type supplies a native type atom, and `<@Process@>` carries an instantiated body. These roles are not interchangeable text substitutions.

Text receives protected native quoting and Number uses finite binary64 encoding. Preserve source bytes and explicit move/borrow behavior. The retained [ProbeIO fixture](../../cmd/l/test/runnable_service/probe.l) and its [caller](../../cmd/l/test/runnable_service/basic.l) witness framing and dictionary binding; its combined stdout/write/unwrap expression is not the decomposition pattern used above.

## Read a complete backend witness

The [declaration/evaluation witness](../../logiceme/testdata/lucid/declaration-evaluation.aml.txt) shows the backend interface; [alphabet and composition](../reference/logemes/alphabet.md) explains its glyphs. Generation, native compilation and execution are separate checks. A compiled library fragment still needs its caller.

## Stay inside the approved surface

Use existing syllables and interfaces; the eighteen Rust segments are frozen. Lifetime/generic inputs are glyph-local native interiors, not Func/Prop/Op fields or inherited defaults. Do not hide algorithms, executable Function bodies or closures in Statement, EA or opaque native helpers. Functions construct, invoke and return through selected definitions. Rust rejects Deferred metadata; scheduling supplies no automatic deferred cleanup.

The [builtin](../../builtins/CONTRACTS.md), [Rust](../../builtins/README.md) and [AML](../../logiceme/AML.md) contracts define the available native surface.
