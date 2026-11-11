# Working on Lyger

Read root [instructions](../AGENTS.md), [architecture](../docs/architecture.md),
[definition semantics](../docs/spec/definitions.md) and
[development guidance](../docs/development-harness.md). Use frontend tracing for
compiler diagnosis and word authoring/review for vocabulary design.

Core loads directly from `builtins/<language>/core.l`. The bundled dictionary
registry is empty. Preserve selected-provider isolation and source ownership;
do not restore Storage, prelude or other removed libraries to satisfy old tests.
The current CLI build boundary is one standalone App using core Print.

## Frontend boundaries

Lyger owns grammar, bindings, types, paragraph activation, definite assignment,
captures and source composition. It emits typed AML through selected definitions;
Logiceme owns native rendering. Do not generate Function bodies, closures,
invocations or returns in Go strings, Statement or EA. Native type/name/lifetime
atoms contain no executable body. Preserve approved logeme interfaces.

Keep lexical POS, inflection role, storage category and value type distinct.
Preserve original spans, source order, literal bytes, typed origins and paragraph
identity through imports and expansion. Pronouns follow source order within their
paragraph, independently of resolution retry order. Aliases and descriptors
retain canonical identity and definition-site scope. Explicit typed Process
inputs and intermediates use lexical allocation, never rendered-name recovery or
global temporary counters.

Literal pools, semantic memory and ordered instructions have separate ownership.
Publish complete AML views atomically, retaining the full semantic graph. Only
diagnostic views may normalize UUIDs or map order. Explicit file includes remain
confined to the selected provider and dictionary root.

Preserve generic compiler coverage for lexical resolution, imports, assignment,
typed construction, captures and source hierarchy. Recognition does not prove a
construct can execute without removed support. Tests of deleted vocabulary are
not requirements for rebuilding it. Do not retain compatibility aliases, source
rewrites or hardcoded native substitutes to make them pass.

## Verification

Run the root temporary hello-world gate with pinned tools and existing budgets.
Keep the exact core output, unsupported-effect rejection and frozen logeme bytes.
Native observers must exercise actual generated source without implementing the
operation themselves. Report source analysis, generation, compilation and runtime
execution separately, including unrun suites.
