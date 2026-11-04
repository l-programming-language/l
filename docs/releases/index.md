# Releases and source identity

Releases are published manually from main after native validation, installer checks and documentation tests. Binaries and their matching docs archive are published together in an immutable release; Pages deployment follows. Merging a change does not publish a release.

## Download L

[L v0.0.1](https://github.com/l-programming-language/l/releases/tag/v0.0.1) is available for Apple Silicon macOS and Linux. Follow [Install L](../cli/installation.md) to get started, or browse [GitHub Releases](https://github.com/l-programming-language/l/releases) for other published versions.

## Versioning and compatibility

Releases are **English-only** until after `v1.0.0`. Release tags start at `v0.0.1` and use `vMAJOR.MINOR.PATCH`; `l --version` reports the bare version.

Backward compatibility is **not guaranteed before `v1.0.0`**. Read the release notes before upgrading, including minor and patch updates.

## Inspect this site's provenance

`/version.json` identifies the language version/commit, documentation revision, channel, and source fingerprint. Local previews are labeled **Local**; source links identify their committed base. A fingerprint identifies content, not a passing test result.

Docs-only corrections use separate immutable `docs-vVERSION-<SHA>` snapshots. These GitHub prereleases contain documentation, leave the language release untouched, and never become the latest installer release. Maintainers can find publication and recovery procedures in the [deployment guide](../../website/DEPLOYMENT.md).
