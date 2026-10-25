# Website deployment: prepared and disabled

**Website CI/CD is prepared but disabled. No deployment has been performed.**

All new workflow files are under `website/ci-templates/` and end in `.yml.disabled`. Existing L workflows and Dependabot configuration remain independent. No install hook, build, test or helper copies these templates into `.github/workflows/`. A manual-only workflow would still be active; these files are physically outside the workflow discovery directory.

`npm run ci:verify-disabled` checks template YAML, full-commit Action pins, non-persisted checkout credentials, dangerous events, active website workflow references, npm lifecycle hooks and website Dependabot entries. Initial/final hashes of existing workflows belong to this implementation's report, not a permanent hash lock in this check. Once maintainers deliberately activate deployment, replace the delivered-state guard with an explicit activated-state policy; do not silently make it accept both states.

## What is implemented

| Inactive file | Intended activated path | Responsibility |
| --- | --- | --- |
| `docs-check.yml.disabled` | `.github/workflows/docs-check.yml` | Unprivileged PR reference/content/type/unit/build/browser checks, all source paths. |
| `docs-release-pages.yml.disabled` | `.github/workflows/docs-release-pages.yml` | Verify real native release proof; build once; retain exact docs artifacts; publish to Pages. |
| `docs-recover-pages.yml.disabled` | `.github/workflows/docs-recover-pages.yml` | Retry or approved rollback from an existing immutable docs archive, without rebuilding. |

The release workflow is reusable through `workflow_call`. A future native release orchestrator on push to `main` must call the activated `.github/workflows/docs-release-pages.yml` after real native publication, passing `native_artifact_id` and mode `release`. Docs-only corrections may explicitly dispatch the activated workflow from main with the existing native artifact identity and mode `docs-correction`. No GITHUB_TOKEN-created release event is relied upon. No website-only path filter is present.

The production branch is configured as main based on the checkout's origin default-ref information. The current branch is not changed by this implementation. Live branch protection, Pages, environments and DNS have not been inspected or modified.

## Native release integration prerequisite

There is no implemented native release publisher in this checkout. `release-policy.json` intentionally contains `nativePublisher: null`. Every production identity path refuses to proceed until a maintainer implements and reviews that integration. Missing VERSION, native assets or validation proof also fails closed; there is no successful placeholder publisher.

The future native publisher must run the authoritative, unmodified preparation and validation:

```sh
scripts/l-ci bootstrap
scripts/l-check
```

These commands belong only in explicitly authorized future language-release validation. They are **not** run by website npm commands or this local implementation. Preserve both unfiltered acceptance phases, complete lint profiles, tagged Expected/Generated native compilation, tool pins, watchdogs and concurrency. The existing CI launcher validates report identity, discovery, phases and native outcomes; the website helper does not replace it with a weaker test runner.

The publisher must package actual native release outputs at the same clean full commit SHA, publish a real non-draft/non-prerelease `vVERSION` release, and produce a GitHub Actions artifact named `l-native-release-<full-language-SHA>`. Its ZIP must contain `native-release.json`, the native files and raw hashed validation evidence. It must then publish **that identical artifact ZIP** as `l-native-release-<SHA>.zip` on the release, the manifest as `l-native-release-<SHA>.json`, and the native artifacts individually. Serialize the manifest using the helper's `stableJSON` (two-space JSON plus final newline); publication compares that exact identity. There is no self-referential archive digest inside the ZIP: the Actions artifact API and release-asset API provide its independent digest and immutable artifact ID.

Register the approved publisher's numeric GitHub workflow ID and exact workflow path in `ci-templates/release-policy.json`. Names and caller-supplied booleans are not trust. Helpers use authenticated GitHub API observations to match repository, workflow ID/path, run/attempt, event, main branch, source SHA, the successful numeric native-publisher job ID, artifact name/run/digest, tag commit, and published native asset hashes. Call the docs workflow from a downstream job after the native producer job completes successfully. The overall orchestration can still be running or can have failed later in docs publication: that downstream failure does not invalidate already verified native proof or block an exact-artifact retry.

Native manifest schema 1 contains:

- `repository`, `languageVersion`, `tag`, `languageCommit`, `dirty: false`.
- `sourceFingerprint`: the shared website source inventory hash with its `sha256:` prefix.
- `releaseFingerprint`: SHA-256 of sorted Git tree entries outside `docs/` and `website/`, including modes, paths and blob identities. This conservative inventory treats an unclassified non-documentation change as release-relevant.
- `producer: {runId, runAttempt, jobId}` identifying the actual native producer, with its numeric job ID checked for successful completion at the matching run/attempt/source SHA.
- Nonempty `artifacts`, each with `path`, `releaseAsset` (matching basename), `sha256`, `target` and `compilerVersion`.
- Nonempty `evidence`, each with repository-independent relative `path` and SHA-256.
- `validation: {command: "scripts/l-check", stages: [...]}`, with exactly lint/acceptance/native entries pointing to hashed canonical stage reports. Reports must share the language SHA, clean status and canonical source identity; commands must have completed successfully and include every required stage command.

Manifest checks do not execute language tools. Fixture tests exercise rejection and integrity logic only; they do not count as native-release validation. Keep producer run records and artifact metadata available; if GitHub no longer provides the required provenance, a new trusted proof/attestation integration is required and promotion remains blocked. Durable release archives do not silently substitute for missing producer authority.

## Release and correction identity

An initial language release requires a root VERSION, matching tag, and identical language/docs SHA. This task creates no VERSION and invents no released L version.

A docs-only correction retains the published language version/commit and native manifest. Both conservative Git release fingerprints and shared reference-source fingerprints must still match. Changes outside `docs/` and `website/`, including code, tests, resources, workflows, lockfiles, install scripts or root README, are conservatively release-relevant. A substantive language change requires a new release decision. A changed historical/internal Markdown file included in the source inventory also requires an explicit reviewed inventory decision; no automatic exception is inferred.

The public version identity records `languageVersion`, `languageCommit`, `docsCommit`, `docsRevision`, `channel`, `dirty`, `sourceFingerprint`, and the release fingerprint. Production docsRevision is the full docs commit; local docsRevision is the deterministic documentation-content digest. These different forms are identified by channel. Source archives without Git remain usable for local builds and cannot impersonate production releases.

Production identity is obtained only through `authenticatedReleaseIdentity()` in `scripts/release.mjs`, using the real native archive and authenticated API evidence. Setting `DOCS_CHANNEL=production` alone grants nothing. The authenticated preparation receives:

```text
DOCS_CHANNEL=production
DOCS_RELEASE_MODE=release|docs-correction
DOCS_NATIVE_ARTIFACT_ID=<immutable GitHub artifact ID>
DOCS_NATIVE_ARCHIVE=<downloaded original ZIP>
DOCS_RELEASE_ARTIFACTS=<safely extracted native ZIP directory>
DOCS_RELEASE_MANIFEST=<that directory>/native-release.json
GITHUB_TOKEN=<read-only validation token>
```

The archive reader rejects traversal, symlinks, duplicate files, malformed directories, unsupported compression/encryption and ZIP64. Its total expansion bound is 1 GiB. It verifies extracted bytes and the manifest against the original ZIP. Future larger native packages need a separately reviewed bounded streaming implementation, not a removed safety check.

## Exact artifact publication and promotion

Website checks run once before sealing the tested `dist/`. A deterministic tar archive, complete file inventory, version-file hash and native-manifest hash form `promotion.json`. The publisher retains these immutable release assets:

```text
docs-<full-docs-SHA>.tar.gz
docs-<full-docs-SHA>.promotion.json
docs-<full-docs-SHA>.native.json
```

An existing asset of the same name must match exactly; it is never overwritten. Pages expands the validated archive, verifies every file and rejects unlisted files, then uploads those exact site bytes. Packaging into the Pages transport artifact does not rebuild the site or pull from a moving branch.

The publication and Pages jobs have separate privileges and protected environments. Read-only checks never use `pull_request_target`. The publisher alone gets contents write; Pages alone gets pages write and id-token write. Checkout does not persist credentials. Both release and recovery use the same concurrency group with cancel-in-progress false. Queue order is not assumed: each promotion compares the observed production archive digest to its captured baseline and checks commit ancestry. Stale promotions and ordinary downgrades reject.

Before publication and again before deployment, `fetch-current` reconciles a successful github-pages deployment record, live `/version.json` and its retained release promotion manifest. A first deployment is recognized only when there is no successful deployment and the configured origin returns HTTP 404; DNS/unavailable/ambiguous states fail. Explicit environment approval remains required. Recovery may run trusted workflow code at a newer SHA while deploying older documented content, so the Pages workflow commit is not misrepresented as the archived docs commit.

Postdeployment checks compare deployed identity to the sealed archive and request homepage, specification, CLI and Pagefind assets. They do not invent a performance score or substitute for the full browser suite run before publication.

## Future activation checklist

Activation is a future maintainer operation. None of these steps has been performed:

1. Implement and validate the native publisher contract, real VERSION/tag policy, supported native targets, evidence retention and producer workflow registration. Configure its all-path push-to-main orchestration to explicitly call the reusable docs workflow after publication.
2. Review the dormant templates and current official [GitHub Pages custom-workflow guide](https://docs.github.com/en/pages/getting-started-with-github-pages/using-custom-workflows-with-github-pages). Replace no existing workflow or required check; retain existing L CI policy. Reverify Action pins before activation.
3. Create protected `documentation-publication` and `github-pages` environments restricted to main with production approval. Permit the minimal token permissions shown in the templates. No personal production credential is needed.
4. Configure repository Pages source as GitHub Actions. Configure `ldocs.org` as the custom domain in Pages, follow GitHub's currently documented DNS/ownership-verification values, wait for provisioning and enable HTTPS. Do not guess DNS records from the repository name or change the site's `/` base to `/l/`.
5. Manually copy the three reviewed templates to the exact activated paths above. Update the explicit delivered-state automation guard policy at that time. No helper in this package performs activation. Keep `/website` npm Dependabot inactive unless separately approved.
6. Run unprivileged PR validation and a reviewed native release against immutable SHAs, approve publication/deployment, then examine the actual production identity, HTTPS, canonical URLs, navigation and built search. Record live evidence separately from the local fixture/actionlint results.

Node is pinned in `.node-version` and all setup-node steps read it. Package installation uses npm ci and the lockfile. Action refs were resolved read-only for these templates: setup-node v4 `49933ea5288caeca8642d1e84afbd3f7d6820020`, configure-pages v5 `983d7736d9b0ae728b81ab479565c72886d7745b`, upload-pages-artifact v3 `56afc609e74202658d3ffba0e8f6dda462b719fa`, deploy-pages v4 `d6db90164ac5ed86f2b6aed7e0febac5b3c0c03e`, download-artifact v4 `d3f86a106a0bac45b974a628896c90dbdf5c8093`. Checkout/upload-artifact retain existing repository pins. These are pinned template dependencies, not evidence of a live workflow run.

## Retry and rollback

If native publication succeeds but docs publication/deployment fails, retain the failure and retry the same sealed artifact. A failed Pages deployment does not justify rebuilding against a newer branch. A retry whose digest is already live is idempotent. A different current production digest rejects a stale retry.

After activation, dispatch `docs-recover-pages.yml` from main with mode retry or rollback, the existing language tag, full docs SHA and expected current archive digest. It downloads the retained release assets, validates release/native producer authority, checks every file and version identity, and uses the original archive. Rollback alone permits backwards ancestry and requires the production environment approval; it never moves a tag, changes the language version or overwrites the original release assets. Postdeployment identity checks still run.

To keep deployment disabled, leave all templates here, leave nativePublisher null, and retain the disabled guard in the website aggregate. Do not copy workflows, configure credentials, dispatch jobs or change remote settings during ordinary local documentation work.
