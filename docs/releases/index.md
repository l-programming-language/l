# Releases and source identity

L releases are published through the repository's GitHub CI/CD. Get the [latest release](https://github.com/l-programming-language/l/releases/latest), or browse [all releases](https://github.com/l-programming-language/l/releases) for a specific version.

## Download L

Choose the binary for your operating system and CPU architecture. English grammar, vocabulary, builtins, standard library, and backend resources are embedded. Follow [installation](../cli/installation.md) for the external build-tool prerequisites and `PATH` setup.

## Versioning and compatibility

Releases are **English-only**. Other source languages are out of scope until after `v1.0.0`. Tags use **major.minor.patch**, starting at `v0.0.1`; `l --version` reports the bare version, `0.0.1`.

Backward compatibility is **not guaranteed before `v1.0.0`**. Read the release notes before upgrading between any pre-1.0 versions, including minor and patch updates.

## Inspect this site's provenance

The site's `/version.json` records its language version/commit, documentation revision, channel and source fingerprint when available. A documentation correction can have a newer revision while describing the same language release.

Local previews read `VERSION` and display **0.0.1 / Local**. Uncommitted source links identify the last committed base; the fingerprint identifies current bytes. Source archives remain usable without Git metadata.

A fingerprint identifies content; it does not establish that tests passed. Maintainers can find release verification, publication and recovery procedures in the [deployment guide](../../website/DEPLOYMENT.md).
