# Author and maintain documentation

Author language content once under `docs/`. Site-specific components, styling and configuration belong in `website/`. Canonical implementation references outside `docs/`, including AML and the Rust contracts, can be published through an explicit allowlist without copying their maintained prose into another source file.

## Add or change a page

Create ordinary Markdown with one page title, a concise introduction and meaningful section headings. Prefer stable conceptual headings over dates or generic labels. Use MDX only when a small presentation component materially improves the explanation.

Add the source to `website/content-map.json` with `source`, `route`, `title`, `description`, `section`, `order`, `role` and `generated`. The source path is relative to the repository. The route is the public URL, rooted at `/`, with a trailing slash. Give each page a unique route and useful description. Use `role` to distinguish normative contract, guide, implementation or future-design material; that role does not establish implementation support or test results.

Use `sidebar: false` for pages reached through a section index or footer. They retain their routes, search entries and source attribution.

Mark mechanical references as generated and attribute their reviewed snapshots and inputs. Edit links must target canonical sources, never the ignored Starlight projection. Preserve existing public routes and useful anchors.

Run the content check after renaming or deleting a page. The preparation pipeline updates only files it owns; authored landing-page content is not a cleanup target. Private manifests, reports and audit maintenance data are not public routes unless deliberately represented by an authored reader-facing explanation.

## Link to canonical sources

Use repository-relative Markdown links in prose. Link to the authored Markdown file and a meaningful heading fragment; the preparation pipeline maps that file to its public route. Reference-style links and image references follow the same ownership rules. A fragment must exist in the target document.

For a source file that is not published as a page, link to its repository-relative path. The site resolves a verified remote and revision when available. In a source archive without Git, it displays an honest source path instead of fabricating a URL. In a dirty checkout, a last-committed source link must be identified as such.

Keep code examples in fenced blocks and AML escaped as code. Never turn native fragments or literal source into raw HTML to improve appearance. Do not rewrite links that merely appear inside a code sample.

## Maintain word entries

[Words A–Z](../reference/words.md), builtin navigation and library lists are derived views of the same canonical entries. Keep the explanation at its existing thematic home or library page. A small library needs one overview, its real import form, a task-to-word table where useful, an alphabetical list and precise entries. Internal dictionary registration does not make its helpers application API; mark deliberately documented foundational meanings as advanced authoring context.

Review current definitions, exports, configured vocabulary and witnesses before selecting coverage. Each eligible builtin, public library meaning, noun/type or grammatical word must have an entry or a reviewed association to the same meaning. Internal value tags and non-word configuration records need honest classifications. Coverage proves completeness against this reviewed inventory; it is not a compiler-enumerated export proof.

Keep identity, qualified name, support, visibility, dictionary ownership, sources, witnesses and the exact `doc` file/anchor in `docs/_meta/semantics.json`. Add only the missing mechanical facts in `word`: `partOfSpeech`, optional `audience` (`application` by default or `authoring`), exceptional `formRoles`, and modifier `targets` referencing `{ "entry": "reviewed-id", "form": "declared-form" }`. Forms come from the existing `inflections` list; roles annotate them rather than declaring new spellings. A record ID is a documentation review identity, not automatically a compiler-resolved identity.

For genuinely overlapping records, use `wordRef` to a canonical record containing `word`. The two fields are mutually exclusive; references cannot chain or cycle. Configured vocabulary mappings use the same reference in `docs/_meta/vocabulary-review.json`. Inspect Function, Error and other overlaps by meaning; do not merge distinct senses or same-spelled words to reduce counts. Modifier target links establish that a meaning/form exists, while source review and relevant tests establish its actual applicability.

Place one marker immediately after the owning heading and any supported explicit anchor. Its ID must match the record's exact `doc` target. The first substantive block must be the definition paragraph, before generated facts or example expansion:

```markdown
### Print

<!-- word:stdlib-export:eng:io:IO.Print -->

Render a supported value and write its bytes to standard output followed by one LF newline.

**Usage:** Explain the actual import, operand roles, result and required App context.

**Example:** Explain what this canonical program demonstrates.

<!-- example:basic -->

**Behavior:** State the relevant demand, ownership, failure or partial-effect policy.

**Compose with:** Link to related canonical entries and explain their different purposes.
```

The template's instructions are authoring scaffolding, not publishable entry prose. Tailor the actual entry: a grammatical word describes its relationship and supported context rather than a fabricated callable signature; a descriptor identifies its exact target and order-sensitive effects; a noun explains what it denotes or constructs. Keep recognized-only or partial forms findable with their limits visible. Link shared lifecycle, formatting and contract explanations rather than repeating them in every entry.

Each supported meaning needs a meaningful usage example, potentially a clearly linked shared example that actually demonstrates it. Show the caller and its enclosing setup; a source-definition excerpt alone is not a use. Keep complete, invalid and fragment classifications accurate. Avoid a copy of the basic program under unrelated words. Examples of unsupported syntax must say what is unsupported and must not appear as runnable instructions.

A section owns content until the next heading at the same or higher level. Nested senses need their own markers and definition paragraphs; ordinary subsections must not substitute for a missing definition. The index gloss comes from that paragraph. Do not duplicate it in JSON or maintain a second list of definitions. Existing useful routes and anchors remain valid; never add an explicit anchor equal to the heading's automatic ID.

Use these literal directives for navigation:

```markdown
<!-- word-index -->
<!-- word-index:builtins -->
<!-- word-index:eng:io -->
```

The first creates the global alphabetical view, the second builtin navigation, and the third selects an existing registry lexicon/key. Other registered libraries use the same form. Only nonempty letter groups get navigation. Generated lists are excluded from full-text search so the actual entries provide ownership, forms, targets and explanations. Fenced and inline directive examples stay literal.

When changing a word, update its canonical entry, reviewed associations and meaningful examples together. Check source dependencies before refreshing hashes or mechanical references. The full production inventory stays mandatory during migration: test a pilot with a coherent fixture inventory, not an allow-missing switch or temporary public exclusion.

## Include an existing example

Register the canonical source in `docs/_meta/examples.json`, then put an inclusion comment in the authored page:

```markdown
<!-- example:basic -->
```

The manifest uses `schemaVersion: 1` and an `examples` array. Each entry has an `id`, repository-relative `source`, classification, description, SHA-256, validation state and witness paths. Complete examples use `runnable`, `compile-only` or `invalid`; partial definitions and backend illustrations use `fragment`. Set `lang` when the source is AML, Rust or another language rather than L.

Expected streams are Base64-encoded exact bytes, accompanied by expected exit status where known. A newline is a byte in the expected result, not a formatting convention. Intentionally invalid inputs require the expected diagnostic boundary. Distinguish compiler-negative fixtures from valid programs that intentionally terminate with an application error `Exit`. Programs requiring a test controller are not standalone runnable examples; the optional verifier must explicitly support that setup before running them.

Never copy canonical source into a second manually edited code fence. The inclusion mechanism reads the canonical bytes during preparation, chooses a safe fence, and retains source attribution. Invalid display bytes must be identified explicitly; silently decoding them with replacement characters is not faithful inclusion.

For a focused excerpt from a larger source, add `"lines": { "start": 12, "end": 20 }` to a manifest entry classified as `fragment`. These are one-based, inclusive source line numbers. The SHA-256 still covers the **entire original file**, so changes outside the excerpt require review too. Preparation preserves each selected line's bytes, indentation and line ending, shows the range in the code title and source attribution, and links to those lines when a verified source URL exists. A final newline does not create another source line. Invalid ranges and non-UTF-8 excerpt sources fail preparation; use whole-file escaped-byte display for sources containing invalid UTF-8.

Excerpts are explanatory fragments, not complete programs. The optional example verifier excludes every `fragment` entry. Show a complete canonical program separately when readers need something to build and run, and explain any enclosing declarations or typed placeholders omitted from an excerpt.

Use new documentation-only `.l` fragments only when the maintained example is embedded in an implementation test or a small illustration clarifies a lexical rule. Label them as static inspection and identify their source or witness without claiming execution. Do not invent syntax for a category represented only by a type tag, grammar entry or backend-native interface.

## Update coverage and snapshots

The coverage manifest connects inventory entries to meaningful documentation anchors and source/test evidence. Mechanical extraction is suitable for JSON grammar rules, vocabulary records, root dictionary-registry entries and known source-derived CLI help. Public visibility, overload behavior, type semantics and operational support require reviewed mappings to frontend behavior and witnesses. Include foundational and runtime entries in coverage while preserving their internal classification; registry membership does not add them to the public standard-library catalog.

After relevant source changes, run `npm run references:check` from `website/`. A mismatch must lead to review of the affected prose and mappings. Use `npm run references:generate` to refresh mechanical snapshots only after that review; changing a fingerprint alone cannot establish semantic correctness.

Checks must reject new unclassified inventory, stale dependencies, nonexistent evidence and missing documentation anchors. Do not “fix” coverage by linking every entry to a generic overview. Closely related forms can share a substantive section, but their support boundaries still need explicit treatment.

## Keep verification layers separate

The default website workflow is Node-only. It checks authored content, projections, references, static output and browser behavior; it does not execute L snippets. `npm run examples:verify -- --compiler /absolute/path/to/l` is separately opt-in and requires an already available compiler plus the native toolchain needed by its builds. It is subject to the repository's validation policy and is not a substitute for full acceptance.

A manual example run must report the exact source fingerprint, host, compiler identity, command, expected/actual bytes and checked stage. A failed or unrun stage stays failed or unrun. Existing witness paths use the status `existing-witness-not-run` until an actual run supplies new evidence.

## Review identity and branding

Root `VERSION` supplies the embedded compiler version and local documentation label. Update it for each release. Local and archive builds cannot establish production identity; authenticated release proof remains required. Docs-only corrections retain the released language identity and record their own revision.

Use the curated vector logo geometry and theme variants. Diagrams need nearby explanatory prose, especially because their visible labels are outlined paths. Use system fonts, Starlight controls and the small L theme layer; do not activate the asset pack's alternate theme, navigation or copy scripts.

Semantic review manifests use schema 3: each manifest records a source hash once, and its entries identify dependencies through `sources`. Refreshing a hash requires reviewing every dependent classification. The generated coverage snapshot also retains the complete mechanical inventory, with its input fingerprints separate from reviewed semantic hashes. Its `notes`, `evidenceSets` and `classifications` catalogs share repeated fields; `referenceEntries` expands them without losing attribution.
