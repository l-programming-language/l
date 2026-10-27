# Contribute to L documentation

Documentation belongs beside the language it explains. Canonical prose lives under `docs/`; the independent `website/` package projects reviewed files into a static Starlight site. Source examples, grammar, vocabulary, builtins and standard dictionaries remain owned by their existing repository paths.

Start with the repository [contributors guide](../../CONTRIBUTING.md) for issue linkage, branch setup, test coverage, protected contracts, and the pull-request checklist.

## Choose the change

For a correction to wording or an example explanation, edit the canonical Markdown source named by the page's source attribution. For a new language feature or API change, inspect the implementation and its existing witnesses, then update the relevant specification prose and coverage mappings together.

The [authoring guide](documentation.md) explains routes, links, examples and snapshots. [Evidence and review](evidence.md) explains how to label support and verification. [Local development](local-development.md) provides the website-only command sequence.

## Preserve the contract

The repository's [AGENTS instructions](../../AGENTS.md) and [development policy](../development-harness.md) apply. The current local validation pause prohibits automatically running expensive compiler bootstrap, full Go lint, full ordinary/native suites and the acceptance gate. A website build must not invoke those indirectly through npm hooks.

Compiler semantics, protected fixtures, exact expected bytes, frozen Rust templates, dependency/tool pins and test assertions are outside an ordinary documentation change. A mismatch between prose and implementation should become a precise audit item with source and witness links. It does not authorize modifying either an oracle or the compiler to make the documentation easier to write.

## Review the result readers receive

Inspect an actual built page, its source attribution, neighboring navigation, heading links and search results. Review the same substantial pages in light and dark modes and at narrow widths. The homepage may use branded artwork; long references should keep headings, code and tables legible.

Generated projections, `node_modules`, HTML output and temporary reports are disposable local state. Small reviewed mechanical snapshots under `docs/_generated/` are maintained reference inputs and must be reviewed when refreshed.

## Deployment is separate

Documentation CI runs reference, content, type, unit, build, and browser checks on pull requests. Release and deployment remain disabled.

Future release and deployment activation is documented in [deployment instructions](../../website/DEPLOYMENT.md). No automatic local setup step activates templates or changes existing workflows. Native release publication and successful validation at the matching source revision remain explicit prerequisites for production promotion.
