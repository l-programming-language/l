# Names, scope, and evaluation state

L resolves names before emitting Rust. Source ownership, lexical scope and canonical identity remain distinct from executable support.

## Declarations and definitions

Quoted labels declare names. Grammar and resolved values determine whether a declaration introduces a value, action, native definition or descriptor. A complete declaration batch publishes atomically; rejection leaves existing bindings intact. Forward local declarations reserve names before lookup can bind an outer meaning.

## Paragraph scope

Four spaces before a paragraph's first code line open a private scope. A physical blank line outside comments or literals ends it. Parent bindings remain visible unless shadowed; private sibling bindings do not.

Public/Publicly publish the selected declaration into its source unit's root without exposing other private members or changing canonical identity.

### Public

[Public in the Word Glossary](../reference/word-glossary.md#word-builtin-public).

### Publicly

[Publicly in the Word Glossary](../reference/word-glossary.md#word-builtin-publicly).

## Pronouns

It, this and that bind the last subject defined in their original paragraph. Mentioning an operand does not change the subject. A new paragraph has no inherited antecedent. In the [basic program](../guide/first-program.md), `The App` selects the App just created.

A source action can bind a pronoun to its typed argument. Imported occurrences retain their own paragraph provenance rather than replacing the caller's antecedent.

### It

[It in the Word Glossary](../reference/word-glossary.md#word-grammar-word-it).

### This

[This in the Word Glossary](../reference/word-glossary.md#word-grammar-word-this).

### That

[That in the Word Glossary](../reference/word-glossary.md#word-grammar-word-that).

## Possessives and member lookup

Use grammatical possessives or explicitly scoped aliases. Dot-qualified authored names are invalid in references, declarations and import aliases, including quoted names and forms. This applies to all authored source, including core, private declarations and imports. Dots in Text, comments, numeric literals, import paths, native code, AML, template slots such as `<@Owner.Member@>` and internal compiler identities retain their separate meanings and bytes. Lookup first checks a complete visible binding, then commits to its canonical owner and member tables. A missing member cannot fall through to an unrelated owner. Nested access does not justify deep declaration ownership; prefer one possessive when declaring a word.

## Aliases and canonical identity

An alias retains its target's identity, members, signature and descriptor applicability. Same-spelled private declarations are different meanings. Qualified members require an explicit alias for unqualified use; private dictionary dependencies do not become caller exports.

A lexical alias may inherit forms only when it removes qualification and retains the unchanged canonical root of an already complete, validated family. For example, selecting a qualified Print root under the unqualified name Print can retain that family. A renamed alias must declare the complete family for its role. An alias selected from an actor form needs an explicit noun family, even when the alias keeps the actor's spelling. Alias chains are checked against the canonical family and selected role, so an intermediate alias cannot turn a rename or actor selection into qualification removal.

These rules apply to import aliases and local lexical aliases. They do not require inflections on ordinary value bindings. A Function value can be stored or forwarded under one value label; creating a new invocation name requires the [explicit verb family](definitions.md#declaration-roles-and-identity).

## Evaluation and definite assignment

Preserve once-only argument evaluation and definition-site bindings through source expansion. A checked value must be assigned on every continuing path to its read; the compiler must not manufacture a default or clone ownership away.

These compiler invariants do not provide a runtime capture implementation. Runtime-dependent effects and captures remain outside the standalone boundary.

## Diagnostics

Unresolved references, duplicate publication, cycles and incompatible assignments retain original spans and expanded call-site context. Internal unknown state is not a fallback value.

## Sources and witnesses

See [scope planning](../../lyger/scope.go), [lexical scope](../../lyger/lexical_scope.go), [semantic resolution](../../lyger/semantic_resolution.go) and [publication](../../lyger/publication.go).
