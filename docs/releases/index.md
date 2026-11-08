# Releases and source identity

Releases are published manually from main after native validation, installer checks and documentation tests. Binaries and their matching docs archive are published together in an immutable release. Pages deployment follows. Merging a change does not publish a release.

## Download L

[L v0.0.1](https://github.com/l-programming-language/l/releases/tag/v0.0.1) is available for Apple Silicon macOS and Linux. Follow [Install L](../cli/installation.md) to get started, or browse [GitHub Releases](https://github.com/l-programming-language/l/releases) for other published versions.

## Versioning and compatibility

Releases are **English-only** until after `v1.0.0`. Release tags start at `v0.0.1` and use `vMAJOR.MINOR.PATCH`. `l --version` reports the bare version.

Backward compatibility is **not guaranteed before `v1.0.0`**. Read the release notes before upgrading, including minor and patch updates.

## Current-source word changes

The current source tree removes these former interfaces without aliases. This
note describes source migration, not an additional published release. Use the
[Storage reference](../internals/storage.md) and [IO reference](../stdlib/eng/io.md)
for current signatures and policies.

| Removed interface | Replacement |
| --- | --- |
| Storage.Propagate | Verify an explicit callback result |
| Storage.Function.Success / Completion | Parameterized Storage.Success |
| Storage.Rejection.Error / Result | Failure / Rejection with explicit arguments |
| IO.Check / Stream | Storage.Result.Expect with a message / IO.Output |
| Storage.Indexed on Iteration or Applier | Storage.Paired; unrelated Indexed meanings remain valid |
| Storage.Fallible on Routine | Explicit generic constructor arguments |

Require now constructs its error only on absence. Its message still evaluates
normally. Accept releases the old error before replacement. Core Retain takes an
explicit borrowed Arc. Removed implementation helpers Storage.Demand, IO.Slice,
Runtime.Dispose and Runtime.Retain use Require, the shared byte view,
Storage.Release and core Retain respectively. No compatibility route remains.

## Inspect this site's provenance

`/version.json` identifies the language version/commit, documentation revision, channel, and source fingerprint. Local previews are labeled **Local**. Source links identify their committed base. A fingerprint identifies content, not a passing test result.

Docs-only corrections use separate immutable `docs-vVERSION-<SHA>` snapshots. These GitHub prereleases contain documentation, leave the language release untouched, and never become the latest installer release. Maintainers can find publication and recovery procedures in the [deployment guide](../../website/DEPLOYMENT.md).
