# Releases and source identity

This documentation build represents the current source checkout. No authoritative released L version is inferred from the website package version, a directory name, a generated artifact-schema version, or a historical test report. When no verified language release is supplied, the visible channel is **Unreleased / Local**.

## Build from the available source

The [installation and CLI guide](../cli/index.md) describes building the compiler and a first native program from source. This site does not invent prebuilt downloads, “latest release” links or historical release notes. Repository changes and a successful merge are not themselves a public release.

Native builds target the current supported macOS or Linux host. Cross-compilation and Phase 2 in-code concurrency remain outside the supported build command. A Linux test record does not establish macOS behavior.

## Inspect this site's provenance

The generated `/version.json` records language version/commit, documentation commit/revision, channel, dirty state and a source fingerprint when available. Unknown release or Git values stay null or explicitly unknown. A dirty working tree is not fully identified by HEAD; its fingerprint records the relevant current input bytes.

Source links may point to the last committed base when local edits are present. The page attribution distinguishes that base from uncommitted content. A source archive without `.git` remains usable locally but cannot manufacture a commit identity or verified remote.

Build metadata avoids an implicit wall-clock timestamp so identical documented inputs can produce reproducible identity. A source fingerprint detects changes; it does not prove semantic correctness, successful compilation or runtime behavior.

## Future release contract

The site supports a future root `VERSION` authority without creating a fictitious initial release. Production promotion requires a valid immutable language release identity, successful matching native release validation/publication evidence, and the exact validated documentation artifact.

A docs-only correction can preserve the released language identity while recording its separate documentation revision, provided release-relevant language inputs have not changed. Changes to those inputs require a new release decision. Do not silently publish documentation for unshipped semantics under an old language version.

Initially there is one current documentation site. A future release can archive its exact built documentation plus identity manifest as a release asset. There are no fake historical versions and no multi-version plugin. A root-built site cannot simply be moved under a version prefix without rebuilding its routes, assets, canonical URLs and search scope.

## Deployment remains disabled

Website CI/CD is prepared but disabled. No deployment has been performed.

Future checks and release/Pages templates reside outside active workflow locations. Native release integration, production approval, domain/HTTPS setup and same-artifact retry/rollback are documented in [the deployment guide](../../website/DEPLOYMENT.md). None of those external settings is changed by installing, building or previewing this site locally.
