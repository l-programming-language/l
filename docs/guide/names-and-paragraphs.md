# Names and paragraphs

A name belongs to a lexical scope. Paragraphs can introduce local names and organize behavior; they are not automatically concurrent tasks.

## Read the physical boundaries

Four spaces before a paragraph's first code line open a local scope. Following lines need not stay indented; a physical blank line outside comments or literals closes it. Preserve these boundaries when copying examples.

<!-- example:local-shadowing -->

The initial message is `outer`. Each indented paragraph owns its own message; the first also owns count and enabled. The final Show refers to the outer message. Expected stdout is `inner7truesecondouter`, without a newline.

Unmarked declarations remain in their scope. Public/Publicly publish only the selected declaration into the enclosing program scope; inside an indexed dictionary, publication stays within that dictionary's own source unit. Public changes visibility only and preserves the declaration's identity.

## Pronouns refer to subjects

Pronouns bind the last subject defined in their source paragraph, not the last printed value or completed App. In the basic program, The App refers to the App just constructed. In `Building any App and then Running it`, it refers to the action's App argument.

Use explicit names when several subjects appear: the [multiple-subjects example](../../demo/paragraphs/multiple-subjects.l) joins in reverse order using `First's Output` and `Second's Output`. These possessives select members without redeclaring their owners. Native Output/Error needs an actual App; a static type namespace alone has no invocation.

## A continuation has an activation

<!-- example:paragraph-fallthrough -->

Here `The App Prints firstText` runs inside First. The following Define and Show form its **continuation**: orchestrator actions deferred until the invocation's first Await joins it. The Second paragraph works the same way.

`Execute First and Second` starts and awaits First, runs First's continuation, then does the same for Second. Finally Show writes `done`; expected stdout is `First\nSecond\ndone`. Repeated Await returns the finalized result without repeating the continuation. A new invocation gets fresh continuation storage; an ordinary paragraph executes when reached.

## Assignment must be established

A value must be assigned on every continuing path before use. Assignment in one conditional branch does not establish it on another; an Exit branch no longer continues. Output readiness is checked at runtime even when names and types resolve.

See [scope ownership](../architecture.md#source-meaning-and-output-views), [paragraph examples](../../demo/paragraphs/README.md), and the [specification](../spec/index.md). Next, [define new vocabulary](definitions.md).
