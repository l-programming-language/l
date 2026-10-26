# Paragraph examples

These maintained sources express Phase 1 directly: `Run` launches, `Await` joins
and activates the invocation's continuation, and `Execute` performs both through
the language's source-defined builtin composition. Tests compile each example's
actual bytes. They do not insert Await, rewrite Execute, or install a project
bootstrap.

The examples now target the [embedded compiler and runtime](../../docs/cli-build.md).
The compiler ships those definitions and dictionary resources in its executable.

Print is explicitly imported from the standard `io` dictionary. Basic, Main,
Process, Start and Execute are supplied by the language. The local-import example
also uses [local-words.l](local-words.l), a small project-owned member reexport;
it contains no printing implementation or copied builtin definitions.

The intended public build invocation is:

```sh
l build fallthrough.l
./fallthrough
```

Each executable takes its entry filename's stem. Its companion Apps live in the
adjacent companion directory. Building does not execute the program.

## Behavior illustrated

A first line indented by four spaces opens a source paragraph. Its continuation
lines need not stay indented; a physical blank line closes it. Public declarations
publish into the containing program. Unmarked declarations remain private to their
lexical scope.

An App-led paragraph runs its associated continuation on the first Await of an
invocation. Repeated Await observes the same result without repeating effects.
An ordinary paragraph executes when reached. Exit terminates the whole L program,
including later source statements and App continuations.

The table uses escaped output: `\n` is an actual newline. Successful cases have
empty stderr. Print appends a newline; Show preserves the supplied text exactly.

| Entry | Purpose | Expected stdout | Status / stderr |
| --- | --- | --- | --- |
| [main.l](main.l) | Preserve the original two-App negative example and raw-output mismatch | Empty | 1; `text values of both x1 & y should be Hello World\n` |
| [fallthrough.l](fallthrough.l) | Sequential Execute with local snapshots | `First\nSecond\ndone` | 0 |
| [exit-program.l](exit-program.l) | Bare Exit suppresses later continuations | `First\n` | 0 |
| [exit-error.l](exit-error.l) | Error Exit terminates the program | `First\n` | 1; `paragraph stopped\n` |
| [multiple-subjects.l](multiple-subjects.l) | Reverse joins select each App's effects | `Second\nsecond-end\|First\n` | 0 |
| [repeated-activation.l](repeated-activation.l) | Two completed invocations retain separate local and outer snapshots | `Repeat\nRepeat\nRepeat\nRepeat\nRepeat\nRepeat\n` | 0 |
| [run-within-paragraph.l](run-within-paragraph.l) | A continuation starts and awaits another App | `Other\nInner\nfinisheddone` | 0 |
| [coordinate-existing.l](coordinate-existing.l) | A paragraph coordinates previously declared Apps | `Second\nFirst\ndone` | 0 |
| [public-values.l](public-values.l) | Public scalars and a project-owned Dictionary member | `public7true` | 0 |
| [local-shadowing.l](local-shadowing.l) | Sibling and outer values retain their identities | `inner7truesecondouter` | 0 |
| [local-definitions.l](local-definitions.l) | Independent paragraph-local Launch actions and Compact descriptors | `First local\nSecond local\ndone` | 0 |
| [local-imports.l](local-imports.l) | Explicit local file/member aliases reexport IO.Print | `Imported\ndone` | 0 |

The negative main example intentionally compares `Hello World!` with raw child
output `Hello World!\n`; output must not be trimmed to make it pass. The separate
[basic demo](../basic/main.l) is the successful Hello World example.

The [CLI suite](../../cmd/l/paragraph_integration_test.go) preserves exact output,
error and publication assertions. Test-owned observer companions separately record
launches and number output when an assertion needs more evidence than captured
stdout. They do not replace the installed IO dictionary or rewrite these demos.
