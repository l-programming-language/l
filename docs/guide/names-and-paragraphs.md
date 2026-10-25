# Names and paragraphs

L resolves names through lexical source scopes, declaration records and member tables. A paragraph is both a source organization unit and, in the supported execution model, a place where local storage and continuation activation can be tracked. It is not an implicit concurrently scheduled task.

## Read the physical boundaries

A first code line with a four-space opening prefix begins an encapsulated paragraph. Its following lines need not remain indented. A physical blank line closes it. Keep the blank lines when copying an example: flattening a program into one paragraph can change its meaning.

<!-- example:local-shadowing -->

Here the first paragraph defines its own `message`, `count` and `enabled`. The second defines another `message`. The last Show occurs in the enclosing scope and refers to the original `outer` value. The retained expected stdout is `inner7truesecondouter` with no trailing newline.

Unmarked declarations stay local to their lexical scope. `Publicly define` and supported Public construction forms publish into the containing program. Public is a visibility rule; it is not a lifetime, mutability, network endpoint or process-selection modifier.

## Pronouns refer to subjects

Pronouns bind the last subject defined in their source paragraph. They do not mean the most recently printed value or whichever App most recently completed. In the basic program, `The App` refers to the App just constructed. In a source-defined action such as `Building any App and then Running it`, `it` follows the action's bound subject.

Prefer explicit names when an example introduces multiple subjects. The maintained [multiple-subjects example](../../demo/paragraphs/multiple-subjects.l) uses `First's Output` and `Second's Output` explicitly while joining in reverse order. Possessives perform member access; they are not a second declaration of the owner.

Native Output/Error lookup requires an actual App. A type namespace may hold static definitions, but merely naming `App` does not create a running invocation.

## A continuation has an activation

<!-- example:paragraph-fallthrough -->

Each App-led paragraph declares an App body and a continuation. The first Await for that invocation joins its result and activates the associated continuation. `Execute First and Second` expands a complete Execute for First before the complete Execute for Second. The expected stream is `First\nSecond\ndone`.

Repeated Await returns the same finalized result without repeating the continuation. A fresh accepted invocation gets fresh local continuation storage. An ordinary paragraph executes when reached instead of waiting for an App invocation.

## Assignment must be established

The typed plan checks definite assignment on continuing control-flow paths. A declaration in one conditional branch does not establish a value in another continuing branch. An Exit branch stops continuing and is treated accordingly. Output readiness remains a runtime check even when name and type resolution succeed.

See [source meaning and scope ownership](../architecture.md#source-meaning-and-output-views), [paragraph examples](../../demo/paragraphs/README.md), and the [language specification](../spec/index.md) for the complete binding rules. Next, [define new vocabulary](definitions.md).
