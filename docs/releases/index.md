# Releases and source identity

Publication of new releases is currently disabled pending [reviewed activation](../../website/DEPLOYMENT.md#deliberate-activation-after-review). When enabled, manual releases from main require native validation, installer checks and documentation tests. Binaries and matching docs archives form one immutable release; Pages deployment follows. Merging a change does not publish a release.

## Current checkout

The core-only reset has removed authored libraries and Runtime. Build the compiler from this checkout to use its standalone App/Print boundary. Previously published binaries may embed older vocabulary and do not validate current behavior. See [installation](../cli/installation.md).

## Versioning and compatibility

Releases are **English-only** until after `v1.0.0`. Release tags start at `v0.0.1` and use `vMAJOR.MINOR.PATCH`. `l --version` reports the bare version.

Backward compatibility is **not guaranteed before `v1.0.0`**. Read the release notes before upgrading, including minor and patch updates.

## Current-source word changes

Only `builtins/eng/core.l` remains as authored builtin source, and the root dictionary registry is empty. Storage, IO, Formatting, Runtime and prelude are removed without compatibility aliases. Core Print accepts Text and adds one newline. Runtime-dependent lifecycle and capture are unavailable pending reconstruction in L.

## Inspect this site's provenance

`/version.json` identifies the language version/commit, documentation revision, channel, and source fingerprint. Local previews are labeled **Local**. Source links identify their committed base. A fingerprint identifies content, not a passing test result.

Docs-only corrections use separate immutable `docs-vVERSION-<SHA>` snapshots. These GitHub prereleases contain documentation, leave the language release untouched, and never become the latest installer release. Maintainers can find publication and recovery procedures in the [deployment guide](../../website/DEPLOYMENT.md).
