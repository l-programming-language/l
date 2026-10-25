# Author and maintain documentation

Author language content once under `docs/`. Site-specific components, styling and configuration belong in `website/`. Canonical implementation references outside `docs/`, including AML and the Rust contracts, can be published through an explicit allowlist without copying their maintained prose into another source file.

## Add or change a page

Create ordinary Markdown with one page title, a concise introduction and meaningful section headings. Prefer stable conceptual headings over dates or generic labels. Use MDX only when a small presentation component materially improves the explanation.

Add the source to `website/content-map.json` with `source`, `route`, `title`, `description`, `section`, `order`, `role` and `generated`. The source path is relative to the repository. The route is the public URL, rooted at `/`, with a trailing slash. Give each page a unique route and useful description. Use `role` to distinguish normative contract, guide, implementation or future-design material; that role does not establish implementation support or test results.

Mark a mechanical reference as generated and point its source attribution at the reviewed snapshot and inputs. An edit link must never invite readers to edit the ignored Starlight projection. Preserve the four existing maintained guide paths and useful anchors.

Run the content check after renaming or deleting a page. The preparation pipeline updates only files it owns; authored landing-page content is not a cleanup target. Private manifests, reports and audit maintenance data are not public routes unless deliberately represented by an authored reader-facing explanation.

## Link to canonical sources

Use repository-relative Markdown links in prose. Link to the authored Markdown file and a meaningful heading fragment; the preparation pipeline maps that file to its public route. Reference-style links and image references follow the same ownership rules. A fragment must exist in the target document.

For a source file that is not published as a page, link to its repository-relative path. The site resolves a verified remote and revision when available. In a source archive without Git, it displays an honest source path instead of fabricating a URL. In a dirty checkout, a last-committed source link must be identified as such.

Keep code examples in fenced blocks and AML escaped as code. Never turn native fragments or literal source into raw HTML to improve appearance. Do not rewrite links that merely appear inside a code sample.

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

The coverage manifest connects inventory entries to meaningful documentation anchors and source/test evidence. Mechanical extraction is suitable for JSON grammar rules, vocabulary records, stdlib index entries and known source-derived CLI help. Public visibility, overload behavior, type semantics and operational support require reviewed mappings to frontend behavior and witnesses.

After relevant source changes, run `npm run references:check` from `website/`. A mismatch must lead to review of the affected prose and mappings. Use `npm run references:generate` to refresh mechanical snapshots only after that review; changing a fingerprint alone cannot establish semantic correctness.

Checks must reject new unclassified inventory, stale dependencies, nonexistent evidence and missing documentation anchors. Do not “fix” coverage by linking every entry to a generic overview. Closely related forms can share a substantive section, but their support boundaries still need explicit treatment.

## Keep verification layers separate

The default website workflow is Node-only. It checks authored content, projections, references, static output and browser behavior; it does not execute L snippets. `npm run examples:verify -- --compiler /absolute/path/to/l` is separately opt-in and requires an already available compiler plus the native toolchain needed by its builds. It is subject to the repository's validation policy and is not a substitute for full acceptance.

A manual example run must report the exact source fingerprint, host, compiler identity, command, expected/actual bytes and checked stage. A failed or unrun stage stays failed or unrun. Existing witness paths use the status `existing-witness-not-run` until an actual run supplies new evidence.

## Review identity and branding

Local and archive builds must not invent release versions, commit IDs or timestamps. Future root `VERSION` support is a release contract, not permission to initialize a fictitious official number. Production builds fail closed without valid release identity; docs-only corrections retain the released language identity and record a separate documentation revision.

Use the curated vector logo geometry and theme variants. Diagrams need nearby explanatory prose, especially because their visible labels are outlined paths. Use system fonts, Starlight controls and the small L theme layer; do not activate the asset pack's alternate theme, navigation or copy scripts.
