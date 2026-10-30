# Contribute to L documentation

Canonical prose lives under `docs/`; `website/` projects it into the documentation site. Start with [CONTRIBUTING.md](../../CONTRIBUTING.md) for issue linkage and the pull-request checklist.

## Choose the change

Edit the canonical Markdown linked by a page's source attribution. Language or API changes also require reviewing implementation evidence and updating the relevant specification and coverage mappings.

- [Authoring](documentation.md): routes, links, examples, and snapshots.
- [Evidence and review](evidence.md): support and verification labels.
- [Local development](local-development.md): setup, preview, and website checks.

## Preserve the contract

Follow [AGENTS.md](../../AGENTS.md) and the [development policy](../development-harness.md), including the local validation pause. Documentation changes do not authorize compiler changes, fixture or expected-byte updates, frozen-template changes, or weakened tests. Investigate prose/implementation mismatches at their source.

## Review the result readers receive

Check built pages, links, navigation, and search in light and dark themes and at narrow widths. Keep generated projections, dependencies, build output, and reports out of Git. Reviewed snapshots under `docs/_generated/` are maintained inputs.

## Deployment is separate

Pull requests run documentation checks without publishing. [Deployment instructions](../../website/DEPLOYMENT.md) define manual releases from main, matching native-release verification, and recovery.
