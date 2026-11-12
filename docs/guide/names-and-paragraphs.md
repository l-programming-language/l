# Names and paragraphs

Names belong to lexical scopes. Four spaces before a paragraph's first code line open a local scope; a physical blank line outside comments or literals ends it. The following lines need not remain indented. Paragraphs do not create concurrent tasks.

Public/Publicly publish the selected declaration into its source unit's root. An imported dictionary retains its own root and cannot inherit caller-local declarations. Visibility does not change a declaration's canonical identity.

## Pronouns and members

Pronouns refer to the last subject defined in their original paragraph. In the [basic program](first-program.md), `The App` refers to Hello. A new paragraph has no inherited subject.

Use grammatical possessives for members, such as `Function's Return`, or import a member under an explicit local alias. Dot-qualified names are not valid authored L. Member lookup stays with the selected owner; a missing member does not fall back to another owner's namespace.

The prohibition includes declaration labels and import aliases. Text, comments, numbers, paths and native/template contents retain their own syntax. Lexical aliases follow the [canonical-root family rule](../spec/names-and-scope.md#aliases-and-canonical-identity); an alias is not a way to omit a renamed word's required forms.

## Source ownership

Keep physical source files, lexical scopes, source-order instructions and native file ownership distinct. An import retains the original source spans and its expansion occurrence. A definition retains its definition-site bindings, while each use has its own arguments and diagnostics.

See [names and scope](../spec/names-and-scope.md) and [source preparation](../spec/lexical-structure.md). Runtime continuations and capture are unavailable in the current executable boundary.
