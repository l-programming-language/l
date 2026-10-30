# Manual releases and documentation deployment

**Publication is implemented but disabled.** The committed [release policy](ci-templates/release-policy.json) has
`publicationEnabled: false` and `nativePublisher: null`. Production helpers and
every publication job reject that state before writing to GitHub or Pages.
Manual dry runs can validate the implementation while the repository is private.

## Workflows and source identity

[Release L](../.github/workflows/release.yml) is the only native release entry
point. It uses `workflow_dispatch`, rejects refs other than `refs/heads/main`,
and checks both the original actor and the actor starting a rerun for Maintain
or Admin access. Each job checks out the dispatch's captured full SHA. The
requested version must equal root `VERSION`; the source tag is `vVERSION` at
that SHA. Merges and tag pushes do not publish anything.

The native publisher runs preparation and canonical `scripts/l-check` on
`darwin-arm64`, `darwin-amd64`, `linux-arm64`, and `linux-amd64`. It retains fresh
lint, acceptance, and tagged-native reports for every target. Full gates keep
the existing pins, both unfiltered test phases, watchdogs and concurrency. A
manual dry run prepares validation and packages without enabling publication.
Do not dispatch it merely to inspect this implementation: the
[validation pause](../docs/development-harness.md#temporary-validation-pause)
still governs expensive local runs and hosted dispatches.

[Documentation release](../.github/workflows/docs-release-pages.yml) is callable
from a downstream job after the native publisher completes. It is also manually
dispatchable for a docs-only correction. [Documentation recovery](../.github/workflows/docs-recover-pages.yml)
is manual-only. All three paths reauthorize immediately before publication or
deployment, including reruns that skip an earlier preflight. A later failure in
docs does not invalidate a completed successful native producer job.

Both docs workflows share `l-documentation-production` concurrency with
`cancel-in-progress: false`. The native orchestrator uses a distinct group so a
parent waiting on its reusable docs child cannot deadlock. Promotion also checks
the previously observed production digest; concurrency alone does not promise
queue order or protect against stale requests.

## Releases in this repository

`l-programming-language/l` hosts each release's installer, final binary archives,
checksums, provenance, full native validation proof and docs recovery assets.
The repository must be public before production publication, so source and
retained raw validation evidence are publicly accessible along with downloads.
There is no separate distribution repository or metadata commit: the single
`vVERSION` tag resolves directly to the captured compiler commit.

Publication uses each job's automatic `GITHUB_TOKEN`; checksums identify the
final distributed bytes. See [activation](#deliberate-activation-after-review)
for permissions and credentials.

The first planned version is `0.0.1`. Production docs require a published,
non-prerelease source release. Pre-1.0 is allowed; GitHub's prerelease flag is not.
Version changes are ordinary reviewed source changes, not automatic workflow
commits. The installed compiler embeds `VERSION` directly.

Production macOS packages below `v1.0.0` are unsigned and do not require Apple
credentials. Starting with `v1.0.0`, production macOS packages must be Developer
ID signed and notarized before packaging and installer verification. Dry runs
always produce unsigned packages and skip signing credentials. The release
manifest and package README record the signing status of the actual binary.
Unsigned downloads may require user approval or be blocked by macOS security
settings; see the [installation guide](../docs/cli/installation.md).

## Native release proof

The publisher uploads `l-native-release-<full-language-SHA>` as a GitHub Actions
artifact. Its ZIP contains `native-release.json`, the native assets and hashed
canonical evidence. The **identical ZIP** is retained on the public release as
`l-native-release-<SHA>.zip`, along with the stable-JSON manifest
`l-native-release-<SHA>.json` and the individual native assets. The manifest does
not contain its own archive digest: GitHub's independent artifact and release
asset digests bind the archive.

Schema 2 keeps source repository, version/tag, clean SHA, source and release
fingerprints, and numeric producer run/attempt/job identities. It adds
`validation.targets`, containing one distinct record per configured native target.
Each record has exactly the original lint, acceptance, and native stage reports.
The verifier checks each report's `goos`/`goarch`, clean revision, source identity,
command inventory, successful completion, and hash. No platform can reuse another
platform's report. Each target also requires a separately hashed installer smoke
record matching its exact archive and version, successful isolated Cargo/Rustup
setup, and offline native compilation. Schema 1 remains readable for its existing tests and historical
protocol, but the new production policy requires schema 2 and manual dispatch.

Authenticated GitHub observations verify the approved numeric workflow ID/path,
main ref, run/attempt, successful producer job, source tag, archive digest and
published asset hashes. Workflow names and caller-supplied pass booleans do not
establish authority. Keep producer run records and artifact metadata available:
durable archives do not silently replace missing producer authority.

## Documentation build and deployment

The site remains static Astro/Starlight with origin `https://ldocs.org` and base
`/`. The release workflow installs locked Node dependencies and Chromium, checks
reference freshness, types/content and unit tests, builds once, verifies static
output, and runs the browser suite against that output. These website checks do
not execute L, Go or Rust.

Production identity is authenticated through `scripts/release.mjs`, not merely
by setting `DOCS_CHANNEL=production`. Initial docs and language releases share
one clean SHA. A docs-only correction retains the language version and native
manifest and must preserve both language-source fingerprints. Changes outside
`docs/` and `website/`, including workflows, release/install scripts, lockfiles and
root README, remain release-relevant.

The tested `dist/` is sealed into an archive with the full file inventory,
version-file hash and native-manifest hash. The same repository release retains:

```text
docs-<full-docs-SHA>.tar.gz
docs-<full-docs-SHA>.promotion.json
docs-<full-docs-SHA>.native.json
```

Existing asset names must match their original hashes; they are never overwritten.
Pages deploys those exact bytes using the existing `github-pages` environment
and custom domain. Checkout never persists credentials; activation preserves
DNS and the site's base path.

The current append-only docs archive model requires GitHub's immutable-release
setting to be off for this repository's releases. Activation must disable that
setting or revise the docs storage design before publication. Enabling it later
requires separate docs snapshot releases/storage or reviewed staging changes.
It must not be enabled on an already-published native release and then used to
append docs assets.

Before publication and again before deployment, the workflow reconciles a
successful Pages deployment record, live `/version.json`, and the retained
promotion manifest. First deployment requires no successful prior deployment and
an HTTP 404 at `/version.json`; ambiguous or unavailable states reject. After
deployment it verifies identity, homepage, specification, CLI and Pagefind routes.

## Retry, rollback and corrections

A failed Pages deployment leaves the native release intact. Dispatch recovery
from main with `mode: retry`, the existing language tag, full docs SHA, and current
production digest. Recovery downloads and validates the retained archive; it
never rebuilds from a moving branch. An already-live identical digest is
idempotent. A changed current digest rejects a stale request.

For a retry before the first successful deployment, enter the literal text `null`
for `expected_current_digest`. The workflow independently verifies that no
successful Pages deployment exists and `/version.json` returns HTTP 404. `null`
is not a wildcard: a different live site rejects that request. If the exact
requested site is already live, recovery verifies it without deploying again.

For a failed native publication, rerun **failed jobs only** in the same workflow
run. That reuses the successful sealing job's original archive and producer
attempt; publication and deployment still reauthorize the new rerun actor. A
full rerun collides with the deliberately immutable artifact names and is rejected
instead of replacing evidence. Start a new manual dry-run dispatch when fresh
validation is needed. A rebuilt package cannot overwrite existing release bytes;
use a newly reviewed version for a changed release.

`mode: rollback` permits older docs ancestry while retaining its original language
and docs identities. It never moves a tag or replaces a native binary. Docs-only
corrections use the documentation release workflow with an existing native
artifact ID and `mode: docs-correction`; fingerprints must still match.

## Deliberate activation after review

1. Merge the reviewed implementation with publication still disabled. GitHub
   must register `release.yml` before its numeric workflow ID can be recorded.
2. Make `l-programming-language/l` public before production publication and resolve
   the immutable-release setting described above. Configure credentials below
   according to the planned release version.
3. Configure `release-build`, `release-publication`, `documentation-publication`
   and `github-pages` environments to permit main only. Keep the existing ldocs.org
   association and enable HTTPS.
   Additional reviewers are optional; Maintain/Admin dispatch is the release
   authorization, subject to available GitHub plan protections.
4. In a separately reviewed commit, set the registered native workflow ID/path
   and `publicationEnabled: true`. Preserve all four
   configured target records and manual-only trigger checks.
5. Explicitly dispatch a dry run and inspect fresh per-target evidence and
   clean-home installer smoke results. Apply the [signing policy](#releases-in-this-repository)
   to the subsequent production run; historical green evidence does not establish
   the new revision's status.
6. Explicitly dispatch the versioned publication and inspect its binary assets,
   public provenance, deployed identity and HTTPS routes. Native publication and
   Pages deployment are separate operations; use retained artifacts for recovery.

Pre-1.0 releases need no custom secrets or variables, GitHub App or personal
access token. Prepare Apple secrets only for `v1.0.0` or later production releases:

| Environment | Values | Purpose |
| --- | --- | --- |
| `release-build` | Secrets `MACOS_CERTIFICATE_P12` (base64), `MACOS_CERTIFICATE_PASSWORD`, `SIGNING_IDENTITY`, `APPLE_ID`, `APPLE_TEAM_ID`, `APPLE_APP_PASSWORD` | Production `v1.0.0` onward: temporary Developer ID keychain and Apple notarization. |
| `release-publication`, `documentation-publication`, `github-pages` | No custom secrets or variables | Use each job's narrowly scoped automatic `GITHUB_TOKEN`; retain main-only environment rules. |

Keep the repository's default workflow permissions read-only. The native release
and docs archive publishers request `contents: write` only in their publication
jobs; the Pages job requests `pages: write` and `id-token: write`. Repository tag
rules must permit the workflow token to create the version tag at the captured
commit. Native publication assembles all release assets in a draft before
publishing; docs recovery assets are then appended to that same release.

The legacy-named `npm run ci:verify-disabled` command now validates the committed
manual release policy as well as unprivileged PR checks. It permits only the three
reviewed release workflow paths, rejects automatic release triggers, mutable
Action pins, missing authorization and publication work before authorization.
Current policy remains disabled. The `.yml.disabled` files in `ci-templates/` are
retained reviewable templates; no runtime helper copies or activates them.
