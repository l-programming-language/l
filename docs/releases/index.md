# Releases and source identity

L releases are published through the repository's GitHub CI/CD. Get the [latest release](https://github.com/l-programming-language/l/releases/latest), or browse [all releases](https://github.com/l-programming-language/l/releases) for a specific version.

## Download L

Choose an asset for your operating system and CPU architecture from the release you intend to use. Read that version's release notes for changes, compatibility and installation details, then follow the [installation guide](../cli/installation.md) to put `l` on your `PATH` and build your first program.

Keep the compiler, release notes and documentation aligned to the same language version. The release page also provides the corresponding tagged source.

## Versioning and compatibility

L release tags use **major.minor.patch**, written `vMAJOR.MINOR.PATCH` (for example, `v0.0.0`).

Backward compatibility is **not guaranteed before `v1.0.0`**. Read the release notes before upgrading between any pre-1.0 versions, including minor and patch updates.

## Inspect this site's provenance

The site's `/version.json` records its language version/commit, documentation revision, channel and source fingerprint when available. A documentation correction can have a newer revision while describing the same language release.

Local previews without release metadata display **Unreleased / Local**. Unknown values remain unknown. For uncommitted edits, source links identify the last committed base while the fingerprint identifies the current input bytes. Source archives without Git metadata remain usable locally.

A fingerprint identifies content; it does not establish that tests passed. Maintainers can find release verification, publication and recovery procedures in the [deployment guide](../../website/DEPLOYMENT.md).
